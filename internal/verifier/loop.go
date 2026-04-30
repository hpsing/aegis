package verifier

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/hpsing/aegis/internal/aegis"
	"github.com/hpsing/aegis/internal/chain"
	"github.com/hpsing/aegis/internal/envelope"
	"github.com/hpsing/aegis/internal/ogstorage"
)

// LoopConfig is the per-process verifier configuration. All inputs come
// from cmd/verifier's env-var parsing — Loop never reads env directly.
type LoopConfig struct {
	OnChain    *OnChain          // AegisContract + Registry bindings
	SwapRPC    chain.EthClient   // chain where the swap happened
	Storage    ogstorage.Storage // 0G Storage; nil disables post-settlement hooks
	INftID     uint64            // verifier's iNFT token id; 0 if not minted yet
	Corrupt    bool              // adversarial test: invert verdict before commit
	SkipReveal bool              // censoring test: commit but never reveal
	LogLabel   string            // prefix for log lines (e.g. "verifier-1")
}

// Loop is the chain-event-driven verifier. State is held in-memory; on
// crash we lose it (acceptable per step-05 scope). Events come from
// AegisContract subscriptions, votes go back via OnChain.
type Loop struct {
	cfg   LoopConfig
	mu    sync.Mutex
	jobs  map[string]*jobState // keyed by jobID.String()
	stats Stats
}

// Stats are exposed for tests that want to assert behavior at the end of
// a scenario.
type Stats struct {
	JobsHandled          uint64
	CommitsLanded        uint64
	RevealsLanded        uint64
	SettlesAttempt       uint64
	StorageAppendOK      uint64
	StorageAppendErrors  uint64
	ProofBundlesUploaded uint64
	ProofBundleErrors    uint64
}

type jobState struct {
	jobID    *big.Int
	verdict  bool
	nonce    [32]byte
	revealed bool
	settled  bool
}

func NewLoop(cfg LoopConfig) *Loop {
	return &Loop{cfg: cfg, jobs: map[string]*jobState{}}
}

// Run subscribes to chain events and processes them. Returns when ctx
// is cancelled. Subscriptions are torn down cleanly on exit.
func (l *Loop) Run(ctx context.Context) error {
	posted := make(chan *aegis.AegisContractJobPosted, 16)
	claimed := make(chan *aegis.AegisContractClaimSubmitted, 16)
	settled := make(chan *aegis.AegisContractJobSettled, 16)

	subPosted, err := l.cfg.OnChain.WatchJobPosted(ctx, posted)
	if err != nil {
		return fmt.Errorf("watch JobPosted: %w", err)
	}
	defer subPosted.Unsubscribe()

	subClaimed, err := l.cfg.OnChain.WatchClaimSubmitted(ctx, claimed)
	if err != nil {
		return fmt.Errorf("watch ClaimSubmitted: %w", err)
	}
	defer subClaimed.Unsubscribe()

	subSettled, err := l.cfg.OnChain.WatchJobSettled(ctx, settled)
	if err != nil {
		return fmt.Errorf("watch JobSettled: %w", err)
	}
	defer subSettled.Unsubscribe()

	l.logf("loop running, addr=%s", l.cfg.OnChain.Address().Hex())

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-subPosted.Err():
			return fmt.Errorf("posted sub: %w", err)
		case err := <-subClaimed.Err():
			return fmt.Errorf("claimed sub: %w", err)
		case err := <-subSettled.Err():
			return fmt.Errorf("settled sub: %w", err)
		case ev := <-posted:
			l.onJobPosted(ctx, ev)
		case ev := <-claimed:
			l.onClaimSubmitted(ctx, ev)
		case ev := <-settled:
			l.onJobSettled(ctx, ev)
		}
	}
}

func (l *Loop) onJobPosted(_ context.Context, ev *aegis.AegisContractJobPosted) {
	l.mu.Lock()
	defer l.mu.Unlock()
	key := ev.JobId.String()
	if _, ok := l.jobs[key]; ok {
		return // duplicate event (chain reorg / replay)
	}
	l.jobs[key] = &jobState{jobID: new(big.Int).Set(ev.JobId)}
	l.logf("JobPosted id=%s client=%s executor=%s", key, ev.Client.Hex(), ev.Executor.Hex())
}

// onClaimSubmitted runs the verification synchronously then submits a
// commit. Re-fetches full Job state from chain rather than trusting
// event ordering — events can race.
func (l *Loop) onClaimSubmitted(ctx context.Context, ev *aegis.AegisContractClaimSubmitted) {
	jobID := new(big.Int).Set(ev.JobId)
	key := jobID.String()
	l.logf("ClaimSubmitted id=%s txHash=%x", key, ev.TxHash)

	job, err := l.cfg.OnChain.GetJob(ctx, jobID)
	if err != nil {
		l.logf("getJob %s: %v", key, err)
		return
	}

	verdict, err := l.runCheck(ctx, jobID, job, ev)
	if err != nil {
		l.logf("verify %s abstaining: %v", key, err)
		return
	}
	if l.cfg.Corrupt {
		verdict = !verdict
		l.logf("CORRUPT MODE: inverted verdict for job %s", key)
	}

	var nonce [32]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		l.logf("nonce: %v", err)
		return
	}

	l.mu.Lock()
	st := l.jobs[key]
	if st == nil {
		st = &jobState{jobID: jobID}
		l.jobs[key] = st
	}
	st.verdict = verdict
	st.nonce = nonce
	l.mu.Unlock()

	commitCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	receipt, err := l.cfg.OnChain.CommitVote(commitCtx, jobID, verdict, nonce)
	if err != nil {
		l.logf("commit %s: %v", key, err)
		return
	}
	l.bump(&l.stats.CommitsLanded)
	l.logf("commit %s verdict=%v tx=%s attempts=%d", key, verdict, receipt.TxHash.Hex(), receipt.Attempts)

	if l.cfg.SkipReveal {
		l.logf("CENSOR MODE: skipping reveal for job %s", key)
		return
	}
	// Schedule reveal after the commit deadline.
	go l.scheduleReveal(ctx, jobID, job)
}

// scheduleReveal polls the chain's block timestamp until it's past
// commitDeadline+1, then reveals. Polling (not wall-clock sleep) so
// that fast-forwarded chain time (evm_increaseTime in tests, or chain
// stalls in production) is detected.
func (l *Loop) scheduleReveal(ctx context.Context, jobID *big.Int, job Job) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(500 * time.Millisecond):
		}
		latest, err := l.cfg.OnChain.GetJob(ctx, jobID)
		if err == nil {
			job = latest
		}
		now, err := l.headTimestamp(ctx)
		if err != nil {
			continue
		}
		if now > job.CommitDeadline {
			break
		}
	}

	l.mu.Lock()
	st := l.jobs[jobID.String()]
	l.mu.Unlock()
	if st == nil || st.revealed {
		return
	}

	rctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	receipt, err := l.cfg.OnChain.RevealVote(rctx, jobID, st.verdict, st.nonce)
	if err != nil {
		l.logf("reveal %s: %v", jobID, err)
		return
	}
	l.mu.Lock()
	st.revealed = true
	l.mu.Unlock()
	l.bump(&l.stats.RevealsLanded)
	l.logf("reveal %s verdict=%v tx=%s attempts=%d", jobID, st.verdict, receipt.TxHash.Hex(), receipt.Attempts)
}

// headTimestamp returns the latest block's timestamp (Unix seconds).
func (l *Loop) headTimestamp(ctx context.Context) (uint64, error) {
	return l.cfg.OnChain.HeadTimestamp(ctx)
}

func (l *Loop) onJobSettled(ctx context.Context, ev *aegis.AegisContractJobSettled) {
	key := ev.JobId.String()
	l.logf("JobSettled id=%s verdict=%v for=%d total=%d",
		key, ev.FinalVerdict, ev.ForVotes.Int64(), ev.TotalReveals.Int64())
	l.mu.Lock()
	st := l.jobs[key]
	if st != nil {
		st.settled = true
	}
	l.mu.Unlock()
	l.bump(&l.stats.JobsHandled)

	// Post-settlement hooks: (1) append a VoteRecord to our per-iNFT log,
	// (2) build + upload a ProofBundle for the whole job. Both fire only
	// for jobs we actually voted on. Failures are logged + counted but
	// never propagated — the on-chain Registry is the source of truth.
	if l.cfg.Storage == nil || st == nil || !st.revealed {
		return
	}
	go l.appendVoteRecord(ctx, ev, st)
	go l.uploadProofBundle(ctx, ev)
}

// uploadProofBundle assembles the per-job ProofBundle from on-chain state
// + the keeper audit trail and uploads to 0G Storage. Best-effort.
func (l *Loop) uploadProofBundle(ctx context.Context, ev *aegis.AegisContractJobSettled) {
	jobID := new(big.Int).Set(ev.JobId)
	job, err := l.cfg.OnChain.GetJob(ctx, jobID)
	if err != nil {
		l.logf("proof bundle %s: getJob: %v", jobID, err)
		return
	}
	uri, err := BuildProofBundle(
		ctx,
		l.cfg.Storage,
		l.cfg.OnChain.keeper,
		job,
		jobID,
		ev.FinalVerdict,
		"0x"+strings.ToLower(hex.EncodeToString(ev.Raw.TxHash[:])),
		ev.Raw.BlockNumber,
		l.cfg.OnChain.Address().Hex(),
	)
	if err != nil {
		l.logf("proof bundle %s: %v", jobID, err)
		l.bump(&l.stats.ProofBundleErrors)
		return
	}
	l.bump(&l.stats.ProofBundlesUploaded)
	l.logf("proof bundle %s -> %s", jobID, uri)
}

// appendVoteRecord runs post-settlement on a goroutine. Retries are
// best-effort; missing records don't slash us.
func (l *Loop) appendVoteRecord(
	ctx context.Context, ev *aegis.AegisContractJobSettled, st *jobState,
) {
	rec := ogstorage.VoteRecord{
		JobID:        ev.JobId.Uint64(),
		VerdictVoted: verdictString(st.verdict),
		VerdictFinal: verdictString(ev.FinalVerdict),
		WasCorrect:   st.verdict == ev.FinalVerdict,
		Timestamp:    time.Now().UnixMilli(),
		// SettleTxHash is on the event log we received; the event struct
		// abigen generated includes Raw which has the tx hash, but for
		// hackathon scope we leave it empty — auditors can correlate via
		// jobID. Production fills this in.
	}
	uri, err := l.cfg.Storage.AppendVoteRecord(ctx, l.cfg.INftID, rec)
	if err != nil {
		l.logf("storage append job=%d: %v", rec.JobID, err)
		l.bump(&l.stats.StorageAppendErrors)
		return
	}
	l.bump(&l.stats.StorageAppendOK)
	l.logf("storage append job=%d entry=%s", rec.JobID, uri)
}

func verdictString(v bool) string {
	if v {
		return "PASS"
	}
	return "FAIL"
}

// runCheck reconstructs the ExecutionClaim from on-chain state + the
// reported result hashes, then runs CheckClaim. The "swap chain"
// EthClient is configured by the caller — production points at Base,
// local devnet uses a mock.
func (l *Loop) runCheck(
	ctx context.Context, _ *big.Int, job Job, ev *aegis.AegisContractClaimSubmitted,
) (bool, error) {
	// Reconstruct a synthetic ExecutionClaim from on-chain state. The
	// Spec is NOT on chain (only its hash is) — for step 5 local mode,
	// we derive a default spec because we know what the publisher
	// posts. Step 7 will switch to fetching the spec via 0G Storage.
	claim := envelope.ExecutionClaim{
		Spec:           defaultLocalSpec(),
		ReportedResult: envelope.ReportedResult{TxHash: hexHash(ev.TxHash[:])},
		ClientAddress:  job.Client.Hex(),
	}
	res, err := CheckClaim(ctx, claim, Context{Chain: l.cfg.SwapRPC})
	if err != nil {
		return false, err
	}
	if res.Verdict == Abstain {
		return false, errors.New("abstain")
	}
	return res.Verdict == Pass, nil
}

// Stats returns a snapshot. Mostly useful for integration tests.
func (l *Loop) Stats() Stats {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.stats
}

func (l *Loop) bump(addr *uint64) {
	l.mu.Lock()
	*addr++
	l.mu.Unlock()
}

func (l *Loop) logf(format string, args ...any) {
	prefix := l.cfg.LogLabel
	if prefix == "" {
		prefix = "[verifier]"
	} else {
		prefix = "[" + prefix + "]"
	}
	log.Printf(prefix+" "+format, args...)
}

// defaultLocalSpec is the ClaimSpec the local-devnet publisher uses. In
// production the spec comes from 0G Storage (step 6/7); for now it's
// hardcoded so verifiers can independently reconstruct it.
func defaultLocalSpec() envelope.ClaimSpec {
	return envelope.ClaimSpec{
		Action:                  "uniswap_v3_swap",
		ChainID:                 8453,
		TokenIn:                 "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
		TokenOut:                "0x4200000000000000000000000000000000000006",
		AmountIn:                "10000000000",
		MaxSlippageBps:          30,
		ReferenceQuoteBlock:     11_999_000,
		ReferenceQuoteAmountOut: "2848000000000000000",
		FeeTier:                 500,
	}
}

// hexHash takes a 32-byte slice and returns it as 0x-prefixed lowercase hex.
func hexHash(b []byte) string {
	if len(b) == 0 {
		return "0x"
	}
	return "0x" + strings.ToLower(hex.EncodeToString(b))
}

// SettleAfter calls settle() once `wait` has elapsed. Used by tests / a
// settler agent. Suppresses the "JobAlreadySettled" revert that's
// expected when multiple settlers race (per step-05 gotcha).
func (l *Loop) SettleAfter(ctx context.Context, jobID *big.Int, wait time.Duration) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(wait):
	}
	tctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	receipt, err := l.cfg.OnChain.Settle(tctx, jobID)
	l.bump(&l.stats.SettlesAttempt)
	if err != nil {
		// Expected when another settler beat us to it.
		l.logf("settle %s: %v (likely already settled)", jobID, err)
		return
	}
	l.logf("settle %s tx=%s", jobID, receipt.TxHash.Hex())
}

// AddrPrefix is a 6-char display abbreviation for an Ethereum address.
func AddrPrefix(a common.Address) string {
	return a.Hex()[:8]
}
