package uiserver

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/hpsing/aegis/internal/aegis"
	"github.com/hpsing/aegis/internal/axl"
	clientpkg "github.com/hpsing/aegis/internal/client"
	"github.com/hpsing/aegis/internal/executor"
)

// runPostJob mirrors the cmd/publisher + cmd/executor handshake in a
// single in-process call. Used by POST /api/post-job so the demo can
// re-run a fresh job without spawning subprocesses.
//
// Steps (same order as scripts/e2e-demo.sh):
//  1. Approve USDC + postJob (signed by Treasury).
//  2. Wait for the receipt; extract jobId from the JobPosted log.
//  3. Push canonical spec to each verifier peer over AXL (pub daemon).
//  4. Build synthetic swap (executor mode=honest), submitClaim signed
//     by Executor.
//
// Returns the jobId + the postJob and submitClaim tx hashes. The
// commit/reveal/settle flow lands via the long-running verifier
// processes — those are spawned by demo-up.sh and observed via the
// chain watcher + log tailer, not invoked here.
func (s *LiveSource) runPostJob(ctx context.Context) (PostJobResult, error) {
	pjLog := func(stage, msg string, kv ...any) {
		// Structured echo to stdout AND the SSE hub, so the timeline
		// pane shows the off-chain steps that aren't on-chain events
		// (mint, approve, axl publish, executor build).
		log.Printf("[post-job] %s: %s%s", stage, msg, fmtKV(kv))
		evt := EventEnvelope{
			"kind":  "postjob.step",
			"stage": stage,
			"msg":   msg,
		}
		for i := 0; i+1 < len(kv); i += 2 {
			if k, ok := kv[i].(string); ok {
				evt[k] = kv[i+1]
			}
		}
		s.hub.Broadcast(evt)
	}

	pjLog("start", "POST /api/post-job invoked")
	if s.cfg.TreasuryPKHex == "" || s.cfg.ExecutorPKHex == "" {
		return PostJobResult{}, errors.New("post-job: TREASURY_PK + EXECUTOR_PK required (set UI_TREASURY_PK / UI_EXECUTOR_PK)")
	}
	if (s.cfg.USDCAddr == common.Address{}) {
		return PostJobResult{}, errors.New("post-job: USDC_CONTRACT not configured")
	}

	pubURL := s.cfg.AXLDaemons["pub"]
	if pubURL == "" {
		return PostJobResult{}, errors.New("post-job: pub AXL daemon URL not configured")
	}
	verifierPeers := s.collectVerifierPeerIDs()
	if len(verifierPeers) == 0 {
		return PostJobResult{}, errors.New("post-job: no verifier AXL peer ids known yet (start verifiers first)")
	}
	pjLog("config", "verifier peers ready", "count", len(verifierPeers))

	treasuryKey, err := ethcrypto.HexToECDSA(strings.TrimPrefix(s.cfg.TreasuryPKHex, "0x"))
	if err != nil {
		return PostJobResult{}, fmt.Errorf("treasury key: %w", err)
	}
	executorKey, err := ethcrypto.HexToECDSA(strings.TrimPrefix(s.cfg.ExecutorPKHex, "0x"))
	if err != nil {
		return PostJobResult{}, fmt.Errorf("executor key: %w", err)
	}
	executorAddr := s.cfg.ExecutorAddress
	if (executorAddr == common.Address{}) {
		executorAddr = ethcrypto.PubkeyToAddress(executorKey.PublicKey)
	}

	chainID, err := s.client.ChainID(ctx)
	if err != nil {
		return PostJobResult{}, fmt.Errorf("chainID: %w", err)
	}

	usdc, err := aegis.NewMockUSDC(s.cfg.USDCAddr, s.client)
	if err != nil {
		return PostJobResult{}, fmt.Errorf("bind USDC: %w", err)
	}

	// Preflight (idempotent): make sure treasury has enough USDC to
	// cover one job (160 mUSDC = reimb 100 + fee 10 + bounty 30 +
	// slack), and that the executor has gas.
	pjLog("preflight", "checking treasury USDC + executor gas")
	if err := s.preflightPostJob(ctx, usdc, treasuryKey, chainID, executorAddr); err != nil {
		pjLog("preflight", "FAILED", "err", err.Error())
		return PostJobResult{}, fmt.Errorf("preflight: %w", err)
	}
	pjLog("preflight", "ok")

	specJSON := canonicalSpecJSON(executorAddr, s.cfg.USDCAddr, big.NewInt(transferAmountMUSDC))
	specHash := ethcrypto.Keccak256Hash(specJSON)
	pjLog("spec", "canonical spec built", "specHash", specHash.Hex())

	// 1. postJob (incl. USDC approve)
	pjctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	pjLog("postJob", "submitting tx (treasury signs)")
	postTx, err := clientpkg.PostJob(
		pjctx, usdc, s.aegis, treasuryKey, chainID, s.cfg.AegisAddr, executorAddr,
		specHash,
		big.NewInt(100_000_000), // executor reimbursement = 100 mUSDC
		big.NewInt(10_000_000),  // executor fee = 10 mUSDC
		big.NewInt(30_000_000),  // verifier bounty = 30 mUSDC
	)
	if err != nil {
		pjLog("postJob", "FAILED", "err", err.Error())
		return PostJobResult{}, fmt.Errorf("postJob: %w", err)
	}
	postJobTx := postTx.Hash().Hex()
	pjLog("postJob", "tx broadcast", "tx", postJobTx)

	// 2. wait for the JobPosted event so we have jobId
	jobID, postBlock, err := s.waitForJobPosted(ctx, postTx.Hash())
	if err != nil {
		pjLog("postJob", "JobPosted not found in receipt", "err", err.Error())
		return PostJobResult{TxHash: postJobTx}, fmt.Errorf("await JobPosted: %w", err)
	}
	pjLog("postJob", "JobPosted event landed", "jobId", jobID.String(), "block", postBlock)

	// 3. AXL spec publish
	pjLog("axl", "publishing spec to verifier peers", "count", len(verifierPeers))
	if err := publishSpecToAXL(ctx, pubURL, verifierPeers, specHash, specJSON); err != nil {
		// Non-fatal for the chain side, but the verifiers will abstain
		// without the spec — surface it loudly.
		pjLog("axl", "spec publish FAILED — verifiers will abstain", "err", err.Error())
		return PostJobResult{TxHash: postJobTx, JobID: jobID.String()},
			fmt.Errorf("axl spec publish: %w", err)
	}
	pjLog("axl", "spec published")

	// 4. solver action: actually transfer mUSDC executor→client on chain.
	transferAmount := big.NewInt(transferAmountMUSDC)
	pjLog("executor", "preflight: minting mUSDC to executor wallet if low")
	if err := s.preflightExecutorBalance(ctx, usdc, executorKey, chainID, transferAmount); err != nil {
		pjLog("executor", "preflight mint FAILED", "err", err.Error())
		return PostJobResult{TxHash: postJobTx, JobID: jobID.String()},
			fmt.Errorf("executor balance preflight: %w", err)
	}
	pjLog("executor", "transferring mUSDC client", "amount", transferAmount.String(), "client", clientAddrFor(treasuryKey).Hex())
	transferTx, err := signedERC20Transfer(ctx, usdc, executorKey, chainID, clientAddrFor(treasuryKey), transferAmount)
	if err != nil {
		pjLog("executor", "transfer FAILED", "err", err.Error())
		return PostJobResult{TxHash: postJobTx, JobID: jobID.String()},
			fmt.Errorf("executor transfer: %w", err)
	}
	if _, err := awaitReceipt(ctx, s.client, transferTx); err != nil {
		pjLog("executor", "transfer receipt timeout", "tx", transferTx.Hex())
		return PostJobResult{TxHash: postJobTx, JobID: jobID.String()},
			fmt.Errorf("transfer receipt: %w", err)
	}
	pjLog("executor", "mUSDC transfer landed", "tx", transferTx.Hex())

	// reportedOutcomeHash binds the (txHash, amount) tuple — the
	// verifier compares this against keccak(actualTx, actualAmount).
	reportedHash := ethcrypto.Keccak256Hash([]byte(transferTx.Hex() + "|" + transferAmount.String()))
	pjLog("executor", "submitting claim (executor signs)", "reported", transferAmount.String())
	scctx, cancel2 := context.WithTimeout(ctx, 60*time.Second)
	defer cancel2()
	subTx, err := executor.SubmitClaim(scctx, s.aegis, executorKey, chainID, jobID, transferTx, reportedHash)
	if err != nil {
		pjLog("executor", "submitClaim FAILED", "err", err.Error())
		return PostJobResult{TxHash: postJobTx, JobID: jobID.String()},
			fmt.Errorf("submitClaim: %w", err)
	}
	pjLog("executor", "submitClaim landed", "tx", subTx.Hash().Hex())
	pjLog("done", "verifiers will now commit/reveal/settle via the chain watcher")
	return PostJobResult{
		JobID:  jobID.String(),
		TxHash: subTx.Hash().Hex(),
	}, nil
}

// transferAmountMUSDC is the amount the executor (acting as a solver)
// transfers to the client to satisfy the job. Matches Spec.AmountIn
// in the canonical spec, which is what the verifier checks against.
const transferAmountMUSDC = int64(50_000_000) // 50 mUSDC, 6 decimals

// clientAddrFor returns the EOA derived from the treasury key. It's
// the address the executor's mUSDC.Transfer must target — equal to
// the address that signs postJob.
func clientAddrFor(k *ecdsa.PrivateKey) common.Address {
	return ethcrypto.PubkeyToAddress(k.PublicKey)
}

// signedERC20Transfer broadcasts a token.transfer(to, value) signed
// by the given key. Returns the tx hash. Caller awaits the receipt.
func signedERC20Transfer(
	ctx context.Context,
	usdc *aegis.MockUSDC,
	signerKey *ecdsa.PrivateKey,
	chainID *big.Int,
	to common.Address,
	value *big.Int,
) (common.Hash, error) {
	opts, err := bind.NewKeyedTransactorWithChainID(signerKey, chainID)
	if err != nil {
		return common.Hash{}, fmt.Errorf("transactor: %w", err)
	}
	opts.Context = ctx
	tx, err := usdc.Transfer(opts, to, value)
	if err != nil {
		return common.Hash{}, fmt.Errorf("transfer: %w", err)
	}
	return tx.Hash(), nil
}

// preflightExecutorBalance mints mUSDC to the executor if its balance
// is below `needed`. MockUSDC.mint is open (the demo deployment uses
// the open-mint variant), so the executor signs its own mint — no
// treasury hop. Idempotent.
func (s *LiveSource) preflightExecutorBalance(
	ctx context.Context,
	usdc *aegis.MockUSDC,
	executorKey *ecdsa.PrivateKey,
	chainID *big.Int,
	needed *big.Int,
) error {
	executorAddr := ethcrypto.PubkeyToAddress(executorKey.PublicKey)
	bal, err := usdc.BalanceOf(&bind.CallOpts{Context: ctx}, executorAddr)
	if err != nil {
		return fmt.Errorf("balanceOf executor: %w", err)
	}
	if bal.Cmp(needed) >= 0 {
		return nil
	}
	mintAmt := new(big.Int).Sub(new(big.Int).Mul(needed, big.NewInt(2)), bal)
	opts, err := bind.NewKeyedTransactorWithChainID(executorKey, chainID)
	if err != nil {
		return fmt.Errorf("mint transactor: %w", err)
	}
	opts.Context = ctx
	tx, err := usdc.Mint(opts, executorAddr, mintAmt)
	if err != nil {
		return fmt.Errorf("mint: %w", err)
	}
	if _, err := awaitReceipt(ctx, s.client, tx.Hash()); err != nil {
		return fmt.Errorf("mint receipt: %w", err)
	}
	return nil
}

// waitForJobPosted polls for the postJob receipt and returns the jobId
// from the JobPosted event log + the block it landed in. 0G's RPC
// occasionally returns null for fresh receipts, so we retry up to 30s.
func (s *LiveSource) waitForJobPosted(ctx context.Context, txHash common.Hash) (*big.Int, uint64, error) {
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, 0, ctx.Err()
		default:
		}
		rcpt, err := s.client.TransactionReceipt(ctx, txHash)
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}
		// Find the JobPosted log. The contract's FilterJobPosted helps,
		// but here we already have the tx hash; just iterate.
		for _, l := range rcpt.Logs {
			if l.Address != s.cfg.AegisAddr {
				continue
			}
			ev, err := s.aegis.ParseJobPosted(*l)
			if err == nil {
				return ev.JobId, rcpt.BlockNumber.Uint64(), nil
			}
		}
		return nil, 0, errors.New("postJob receipt landed but no JobPosted log")
	}
	return nil, 0, errors.New("postJob receipt timeout")
}

// collectVerifierPeerIDs returns the AXL peer IDs of the verifier
// daemons (everything in the AXL config except "pub"). The daemons
// expose them via /topology — pollTopology already calls those, so we
// can read them off the mirror.
func (s *LiveSource) collectVerifierPeerIDs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []string{}
	for _, n := range s.topology.Nodes {
		if n.Role == "pub" || !n.Online {
			continue
		}
		if n.PeerID != "" {
			out = append(out, n.PeerID)
		}
	}
	return out
}

// canonicalSpecJSON describes the workload the executor (acting as a
// solver) is committing to perform. Demo build: a real mUSDC transfer
// on 0G Galileo. The verifier reads the on-chain Transfer event and
// confirms the recipient + amount match.
//
// Field order matches the canonical-JSON sort so the on-chain
// specHash is reproducible. Keys must match what
// internal/verifier/dispatcher.go and check_usdc_transfer.go read.
//
// The publisher binary (cmd/publisher) still emits the older
// uniswap_v3_swap shape — that path goes through CheckUniswapSwap
// and produces the synthetic-mode FAIL we documented earlier. The
// UI's runPostJob path uses this newer transfer shape and produces
// PASS jobs.
func canonicalSpecJSON(executor, usdc common.Address, transferAmount *big.Int) []byte {
	// Keys here MUST match the json tags on envelope.ClaimSpec.
	// Earlier camelCase keys (tokenIn / amountIn) silently produced
	// empty fields when the verifier unmarshaled, which manifested
	// as unanimous FAIL with no clear reason. Snake-case is correct.
	body := map[string]any{
		"action":    "mock_usdc_transfer",
		"amount_in": transferAmount.String(),
		"chain_id":  int64(16602), // 0G Galileo — must match SwapRPC
		"executor":  executor.Hex(),
		"intent":    "transfer",
		"token_in":  usdc.Hex(),
	}
	out, _ := json.Marshal(body)
	return out
}

// publishSpecToAXL sends a TypeSpecPublish envelope to every verifier
// peer via the publisher's AXL daemon HTTP API. Mirrors the helper in
// cmd/publisher/main.go; lifted here so the UI is not shell-coupled.
func publishSpecToAXL(ctx context.Context, nodeURL string, peerIDs []string, specHash common.Hash, specJSON []byte) error {
	c := axl.NewClient(nodeURL)
	payload := axl.SpecPayload{JobID: 0, SpecHash: specHash.Hex(), Spec: specJSON}
	body, err := axl.MarshalEnvelope(axl.TypeSpecPublish, payload)
	if err != nil {
		return fmt.Errorf("marshal envelope: %w", err)
	}
	var firstErr error
	for _, p := range peerIDs {
		if err := c.Send(ctx, p, body); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// buildPostJobPreview returns the spec + escrow amounts the server
// WOULD post if /api/post-job were called right now. The UI shows
// this in a confirmation modal before firing the real call.
//
// All fields mirror canonicalSpecJSON / runPostJob exactly so the
// preview is honest. `Ready` is true iff every prerequisite (PKs,
// USDC contract, online verifier peers) is satisfied.
func (s *LiveSource) buildPostJobPreview() (PostJobPreview, error) {
	executorAddr := s.cfg.ExecutorAddress
	if (executorAddr == common.Address{}) && s.cfg.ExecutorPKHex != "" {
		k, err := ethcrypto.HexToECDSA(strings.TrimPrefix(s.cfg.ExecutorPKHex, "0x"))
		if err == nil {
			executorAddr = ethcrypto.PubkeyToAddress(k.PublicKey)
		}
	}

	clientAddr := common.Address{}
	if s.cfg.TreasuryPKHex != "" {
		k, err := ethcrypto.HexToECDSA(strings.TrimPrefix(s.cfg.TreasuryPKHex, "0x"))
		if err == nil {
			clientAddr = ethcrypto.PubkeyToAddress(k.PublicKey)
		}
	}

	specJSON := canonicalSpecJSON(executorAddr, s.cfg.USDCAddr, big.NewInt(transferAmountMUSDC))
	specHash := ethcrypto.Keccak256Hash(specJSON)

	// Decode the canonical spec we just built so we can show its fields
	// in the preview without re-typing them. Fields here must stay in
	// sync with canonicalSpecJSON.
	var raw map[string]any
	_ = json.Unmarshal(specJSON, &raw)
	getStr := func(k string) string { v, _ := raw[k].(string); return v }
	getInt := func(k string) int {
		switch v := raw[k].(type) {
		case float64:
			return int(v)
		case int:
			return v
		}
		return 0
	}

	peers := []PreviewPeer{}
	s.mu.RLock()
	for _, n := range s.topology.Nodes {
		if n.Role == "pub" || !n.Online {
			continue
		}
		peers = append(peers, PreviewPeer{Role: n.Role, PeerID: n.PeerID})
	}
	s.mu.RUnlock()

	blockers := []string{}
	if s.cfg.TreasuryPKHex == "" {
		blockers = append(blockers, "TREASURY_PK not configured")
	}
	if s.cfg.ExecutorPKHex == "" {
		blockers = append(blockers, "EXECUTOR_PK not configured")
	}
	if (s.cfg.USDCAddr == common.Address{}) {
		blockers = append(blockers, "USDC_CONTRACT not configured")
	}
	if s.cfg.AXLDaemons["pub"] == "" {
		blockers = append(blockers, "publisher AXL daemon not configured")
	}
	if len(peers) == 0 {
		blockers = append(blockers, "no verifier AXL peers online — start verifiers first")
	}

	chainID := int64(0)
	if v, ok := raw["chain_id"].(float64); ok {
		chainID = int64(v)
	}

	return PostJobPreview{
		Spec: PreviewSpec{
			Action:         getStr("action"),
			ChainID:        chainID,
			Intent:         getStr("intent"),
			TokenIn:        getStr("token_in"),
			TokenOut:       getStr("token_out"),
			AmountIn:       getStr("amount_in"),
			MaxSlippageBps: getInt("max_slippage_bps"),
		},
		SpecHash:      specHash.Hex(),
		Client:        clientAddr.Hex(),
		Executor:      executorAddr.Hex(),
		Reimbursement: "100000000", // 100 mUSDC
		Fee:           "10000000",  // 10 mUSDC
		Bounty:        "30000000",  // 30 mUSDC
		TotalEscrow:   "140000000", // 140 mUSDC
		VerifierPeers: peers,
		Ready:         len(blockers) == 0,
		Blockers:      blockers,
	}, nil
}

// fmtKV stringifies a flat ("key", value, "key", value, …) slice for
// the stdout log line. Hub events get a separate structured envelope.
func fmtKV(kv []any) string {
	if len(kv) == 0 {
		return ""
	}
	var b strings.Builder
	for i := 0; i+1 < len(kv); i += 2 {
		fmt.Fprintf(&b, " %v=%v", kv[i], kv[i+1])
	}
	return b.String()
}

// preflightPostJob makes the POST /api/post-job button self-contained:
// mints USDC to the treasury if its balance is below `needed`, approves
// AegisContract for the upcoming pull, and verifies the executor has
// some 0G for gas. All txs are signed locally; receipts are awaited so
// the next call (postJob) can rely on state being settled.
func (s *LiveSource) preflightPostJob(
	ctx context.Context,
	usdc *aegis.MockUSDC,
	treasuryKey *ecdsa.PrivateKey,
	chainID *big.Int,
	executorAddr common.Address,
) error {
	const needed = int64(160_000_000) // 100 reimb + 10 fee + 30 bounty + 20 slack
	treasuryAddr := ethcrypto.PubkeyToAddress(treasuryKey.PublicKey)

	bal, err := usdc.BalanceOf(&bind.CallOpts{Context: ctx}, treasuryAddr)
	if err != nil {
		return fmt.Errorf("balanceOf: %w", err)
	}
	if bal.Cmp(big.NewInt(needed)) < 0 {
		mintAmt := new(big.Int).Sub(big.NewInt(2*needed), bal)
		opts, err := bind.NewKeyedTransactorWithChainID(treasuryKey, chainID)
		if err != nil {
			return fmt.Errorf("transactor: %w", err)
		}
		opts.Context = ctx
		tx, err := usdc.Mint(opts, treasuryAddr, mintAmt)
		if err != nil {
			return fmt.Errorf("mint: %w", err)
		}
		if _, err := awaitReceipt(ctx, s.client, tx.Hash()); err != nil {
			return fmt.Errorf("mint receipt: %w", err)
		}
	}

	// Approve regardless — cheap, idempotent, simplifies retries on
	// flaky 0G null-response RPC.
	apprOpts, err := bind.NewKeyedTransactorWithChainID(treasuryKey, chainID)
	if err != nil {
		return fmt.Errorf("approve transactor: %w", err)
	}
	apprOpts.Context = ctx
	apprTx, err := usdc.Approve(apprOpts, s.cfg.AegisAddr, big.NewInt(needed*2))
	if err != nil {
		return fmt.Errorf("approve: %w", err)
	}
	if _, err := awaitReceipt(ctx, s.client, apprTx.Hash()); err != nil {
		return fmt.Errorf("approve receipt: %w", err)
	}

	// Executor gas check — surface a clear error rather than letting
	// the eventual submitClaim fail with "insufficient funds".
	exBal, err := s.client.BalanceAt(ctx, executorAddr, nil)
	if err == nil && exBal.Sign() == 0 {
		return fmt.Errorf("executor %s has 0 0G — fund it first (e.g. `cast send %s --value 0.05ether ...`)",
			executorAddr.Hex(), executorAddr.Hex())
	}
	return nil
}

// awaitReceipt polls until a tx's receipt is available. 0G Galileo's
// public RPC sometimes returns null for fresh receipts; we tolerate it
// for ~30s before giving up.
func awaitReceipt(ctx context.Context, c interface {
	TransactionReceipt(ctx context.Context, h common.Hash) (*types.Receipt, error)
}, h common.Hash) (*types.Receipt, error) {
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		r, err := c.TransactionReceipt(ctx, h)
		if err == nil && r != nil {
			return r, nil
		}
		time.Sleep(2 * time.Second)
	}
	return nil, fmt.Errorf("receipt %s: timeout", h.Hex())
}

// Compile-time check: post-job uses ParseJobPosted via the abigen
// filterer. The type-assertion path here is a smoke check that the
// helpers haven't drifted.
var (
	_ = (*aegis.AegisContractFilterer)(nil)
	_ = (*aegis.AegisContractCaller)(nil)
	_ = bytes.Buffer{}
	_ = http.StatusOK
)
