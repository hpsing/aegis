package uiserver

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hpsing/aegis/internal/aegis"
)

// backfillEvents scans events in [head-BackfillBlocks, head] and seeds
// the in-memory mirror. Called once on startup.
func (s *LiveSource) backfillEvents(ctx context.Context) error {
	head, err := s.client.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("head: %w", err)
	}
	from := uint64(0)
	if s.cfg.BackfillBlocks > 0 && head > s.cfg.BackfillBlocks {
		from = head - s.cfg.BackfillBlocks
	}
	s.mu.Lock()
	s.lastBlock = head
	s.mu.Unlock()
	return s.scanRange(ctx, from, head, false /* don't broadcast SSE for backfill */)
}

// watchChain polls for new events forever. Single goroutine; cheap on
// 0G Galileo (~2s blocks, <300 logs/period). Subscriptions via WS would
// be cheaper but the public RPC is HTTP-only.
func (s *LiveSource) watchChain(ctx context.Context) {
	t := time.NewTicker(s.cfg.PollInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
		head, err := s.client.BlockNumber(ctx)
		if err != nil {
			continue
		}
		s.mu.Lock()
		from := s.lastBlock + 1
		s.mu.Unlock()
		if from > head {
			continue
		}
		if err := s.scanRange(ctx, from, head, true); err != nil {
			// transient — log and keep going next tick
			continue
		}
		s.mu.Lock()
		s.lastBlock = head
		s.mu.Unlock()
	}
}

// scanRange filters all five Aegis events in [from, to] and applies
// each. If broadcast is true, also pushes a chain.* event to the SSE
// hub for each.
func (s *LiveSource) scanRange(ctx context.Context, from, to uint64, broadcast bool) error {
	opts := &bind.FilterOpts{Start: from, End: &to, Context: ctx}

	if it, err := s.aegis.FilterJobPosted(opts, nil, nil, nil); err == nil {
		for it.Next() {
			s.applyJobPosted(ctx, it.Event, broadcast)
		}
		_ = it.Close()
	}
	if it, err := s.aegis.FilterClaimSubmitted(&bind.FilterOpts{Start: from, End: &to, Context: ctx}, nil); err == nil {
		for it.Next() {
			s.applyClaimSubmitted(ctx, it.Event, broadcast)
		}
		_ = it.Close()
	}
	if it, err := s.aegis.FilterVoteCommitted(&bind.FilterOpts{Start: from, End: &to, Context: ctx}, nil, nil); err == nil {
		for it.Next() {
			s.applyVoteCommitted(ctx, it.Event, broadcast)
		}
		_ = it.Close()
	}
	if it, err := s.aegis.FilterVoteRevealed(&bind.FilterOpts{Start: from, End: &to, Context: ctx}, nil, nil); err == nil {
		for it.Next() {
			s.applyVoteRevealed(ctx, it.Event, broadcast)
		}
		_ = it.Close()
	}
	if it, err := s.aegis.FilterJobSettled(&bind.FilterOpts{Start: from, End: &to, Context: ctx}, nil); err == nil {
		for it.Next() {
			s.applyJobSettled(ctx, it.Event, broadcast)
		}
		_ = it.Close()
	}
	if broadcast {
		s.recomputeTotals()
	}
	return nil
}

func (s *LiveSource) applyJobPosted(ctx context.Context, ev *aegis.AegisContractJobPosted, broadcast bool) {
	id := ev.JobId.Uint64()
	ts := s.blockTimestamp(ctx, ev.Raw.BlockNumber)
	s.mu.Lock()
	j, ok := s.jobs[id]
	if !ok {
		j = &jobState{}
		s.jobs[id] = j
	}
	j.summary.ID = ev.JobId.String()
	j.summary.Status = "Posted"
	j.summary.Client = ev.Client.Hex()
	j.summary.Executor = ev.Executor.Hex()
	j.summary.PostedBlock = ev.Raw.BlockNumber
	j.summary.PostedAt = ts
	j.detail.JobSummary = j.summary
	j.detail.SpecHash = hexHash(ev.SpecHash)
	j.detail.PostJobTx = ev.Raw.TxHash.Hex()
	s.mu.Unlock()
	if broadcast {
		s.hub.Broadcast(EventEnvelope{
			"kind":     "chain.JobPosted",
			"jobId":    j.summary.ID,
			"client":   j.summary.Client,
			"executor": j.summary.Executor,
			"tx":       j.detail.PostJobTx,
			"block":    ev.Raw.BlockNumber,
			"blockTs":  ts,
		})
	}
}

func (s *LiveSource) applyClaimSubmitted(ctx context.Context, ev *aegis.AegisContractClaimSubmitted, broadcast bool) {
	id := ev.JobId.Uint64()
	ts := s.blockTimestamp(ctx, ev.Raw.BlockNumber)
	s.mu.Lock()
	j, ok := s.jobs[id]
	if !ok {
		j = &jobState{summary: JobSummary{ID: ev.JobId.String()}}
		s.jobs[id] = j
	}
	j.summary.Status = "ClaimSubmitted"
	j.detail.JobSummary = j.summary
	j.detail.TxHash = hexHash(ev.TxHash)
	j.detail.ReportedOutcomeHash = hexHash(ev.ReportedOutcomeHash)
	j.detail.SubmitClaimTx = ev.Raw.TxHash.Hex()
	s.mu.Unlock()
	if broadcast {
		s.hub.Broadcast(EventEnvelope{
			"kind":           "chain.ClaimSubmitted",
			"jobId":          j.summary.ID,
			"tx":             j.detail.SubmitClaimTx,
			"reportedTxHash": j.detail.TxHash,
			"blockTs":        ts,
		})
	}
}

func (s *LiveSource) applyVoteCommitted(ctx context.Context, ev *aegis.AegisContractVoteCommitted, broadcast bool) {
	id := ev.JobId.Uint64()
	ts := s.blockTimestamp(ctx, ev.Raw.BlockNumber)
	s.mu.Lock()
	j, ok := s.jobs[id]
	if !ok {
		j = &jobState{summary: JobSummary{ID: ev.JobId.String()}}
		s.jobs[id] = j
	}
	vote := VerifierVote{
		Verifier:    ev.Verifier.Hex(),
		TxHash:      ev.Raw.TxHash.Hex(),
		BlockNumber: ev.Raw.BlockNumber,
		CommitHash:  hexHash(ev.CommitHash),
	}
	if !hasVote(j.detail.Commits, vote.Verifier) {
		j.detail.Commits = append(j.detail.Commits, vote)
	}
	s.mu.Unlock()
	s.recordCommitTime(ev.JobId.String(), vote.Verifier, ts)
	if broadcast {
		s.hub.Broadcast(EventEnvelope{
			"kind":     "chain.VoteCommitted",
			"jobId":    ev.JobId.String(),
			"verifier": vote.Verifier,
			"tx":       vote.TxHash,
			"blockTs":  ts,
		})
	}
}

func (s *LiveSource) applyVoteRevealed(ctx context.Context, ev *aegis.AegisContractVoteRevealed, broadcast bool) {
	id := ev.JobId.Uint64()
	ts := s.blockTimestamp(ctx, ev.Raw.BlockNumber)
	s.mu.Lock()
	j, ok := s.jobs[id]
	if !ok {
		j = &jobState{summary: JobSummary{ID: ev.JobId.String()}}
		s.jobs[id] = j
	}
	vote := VerifierVote{
		Verifier:    ev.Verifier.Hex(),
		TxHash:      ev.Raw.TxHash.Hex(),
		BlockNumber: ev.Raw.BlockNumber,
		Verdict:     verdictStr(ev.Verdict),
	}
	if !hasVote(j.detail.Reveals, vote.Verifier) {
		j.detail.Reveals = append(j.detail.Reveals, vote)
	}
	s.mu.Unlock()
	if broadcast {
		s.hub.Broadcast(EventEnvelope{
			"kind":     "chain.VoteRevealed",
			"jobId":    ev.JobId.String(),
			"verifier": vote.Verifier,
			"verdict":  vote.Verdict,
			"tx":       vote.TxHash,
			"blockTs":  ts,
		})
	}
}

func (s *LiveSource) applyJobSettled(ctx context.Context, ev *aegis.AegisContractJobSettled, broadcast bool) {
	id := ev.JobId.Uint64()
	ts := s.blockTimestamp(ctx, ev.Raw.BlockNumber)
	s.mu.Lock()
	j, ok := s.jobs[id]
	if !ok {
		j = &jobState{summary: JobSummary{ID: ev.JobId.String()}}
		s.jobs[id] = j
	}
	j.summary.Status = "Settled"
	j.summary.FinalVerdict = verdictStr(ev.FinalVerdict)
	j.summary.ForVotes = int(ev.ForVotes.Int64())
	j.summary.TotalReveals = int(ev.TotalReveals.Int64())
	j.summary.SettledAt = ts
	j.detail.JobSummary = j.summary
	j.detail.SettleTx = ev.Raw.TxHash.Hex()
	s.mu.Unlock()
	s.recordSettleLatency(ev.JobId.String(), ts)
	if broadcast {
		s.hub.Broadcast(EventEnvelope{
			"kind":         "chain.JobSettled",
			"jobId":        ev.JobId.String(),
			"finalVerdict": j.summary.FinalVerdict,
			"forVotes":     j.summary.ForVotes,
			"totalReveals": j.summary.TotalReveals,
			"tx":           j.detail.SettleTx,
			"blockTs":      ts,
		})
	}
}

func hasVote(votes []VerifierVote, verifier string) bool {
	v := strings.ToLower(verifier)
	for _, x := range votes {
		if strings.ToLower(x.Verifier) == v {
			return true
		}
	}
	return false
}

// fetchJob reads a Job from chain on demand (cache miss path).
func (s *LiveSource) fetchJob(ctx context.Context, id *big.Int) (JobDetail, error) {
	raw, err := s.aegis.Jobs(&bind.CallOpts{Context: ctx}, id)
	if err != nil {
		return JobDetail{}, fmt.Errorf("jobs(%s): %w", id, err)
	}
	d := JobDetail{
		JobSummary: JobSummary{
			ID:       id.String(),
			Status:   statusStr(raw.Status),
			Client:   raw.Client.Hex(),
			Executor: raw.Executor.Hex(),
		},
		SpecHash:            hexHash(raw.SpecHash),
		TxHash:              hexHash(raw.TxHash),
		ReportedOutcomeHash: hexHash(raw.ReportedOutcomeHash),
		ClaimDeadline:       int64(raw.ClaimDeadline),
		CommitDeadline:      int64(raw.CommitDeadline),
		RevealDeadline:      int64(raw.RevealDeadline),
		Commits:             []VerifierVote{},
		Reveals:             []VerifierVote{},
		ProofBundles:        []ProofBundleRef{},
	}
	return d, nil
}

// refreshVerifiers reads the registry's verifierAddresses + per-address
// state. The registry doesn't expose iteration, but it does emit
// VerifierRegistered events; we scan them to discover addresses, then
// call verifiers(addr) for each.
func (s *LiveSource) refreshVerifiers(ctx context.Context) error {
	addrs, err := s.discoverVerifiers(ctx)
	if err != nil {
		return err
	}
	out := make([]VerifierInfo, 0, len(addrs))
	for _, a := range addrs {
		v, err := s.reg.Verifiers(&bind.CallOpts{Context: ctx}, a)
		if err != nil {
			continue
		}
		bps := 0
		if v.VotesTotal.Sign() > 0 {
			bps = int(new(big.Int).Div(new(big.Int).Mul(v.VotesCorrect, big.NewInt(10000)), v.VotesTotal).Int64())
		}
		// iNftId is the second `register(stake, iNftId)` arg. The current
		// scripts pass 0 (meaning "not minted yet"); skip the field rather
		// than render a misleading "iNFT #0" chip.
		iNFTId := ""
		if v.INftId.Sign() > 0 {
			iNFTId = v.INftId.String()
		}
		info := VerifierInfo{
			Address:      a.Hex(),
			INFTID:       iNFTId,
			Stake:        formatMUSDC(v.Stake),
			VotesTotal:   int(v.VotesTotal.Int64()),
			VotesCorrect: int(v.VotesCorrect.Int64()),
			AccuracyBps:  bps,
			Active:       v.Active,
			Role:         "verifier",
		}
		out = append(out, info)
	}
	s.mu.Lock()
	s.verifiers = out
	s.totals.SwarmTotal = len(out)
	s.mu.Unlock()
	return nil
}

func (s *LiveSource) discoverVerifiers(ctx context.Context) ([]common.Address, error) {
	head, err := s.client.BlockNumber(ctx)
	if err != nil {
		return nil, err
	}
	// Registry scan ignores BackfillBlocks: there are only a handful of
	// VerifierRegistered events for the lifetime of a deployment, and
	// they happen at registration time (potentially long before the UI
	// boots), so always scan the full range.
	it, err := s.reg.FilterVerifierRegistered(&bind.FilterOpts{Start: 0, End: &head, Context: ctx}, nil)
	if err != nil {
		return nil, fmt.Errorf("filter VerifierRegistered: %w", err)
	}
	defer it.Close()
	seen := map[common.Address]bool{}
	addrs := []common.Address{}
	for it.Next() {
		a := it.Event.Verifier
		if !seen[a] {
			seen[a] = true
			addrs = append(addrs, a)
		}
	}
	return addrs, nil
}

// formatMUSDC turns a uint256 in micro-USDC (6 decimals) into a clean
// integer string of mUSDC. 100_000_000 → "100".
func formatMUSDC(v *big.Int) string {
	if v == nil {
		return "0"
	}
	out := new(big.Int).Quo(v, big.NewInt(1_000_000))
	return out.String()
}
