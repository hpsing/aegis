package keeperhub

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// LiveClient is the production Client backed by KeeperHub's MCP at
// https://app.keeperhub.com/mcp.
//
// Integration model: `call_workflow`. KeeperHub's role is to BUILD the
// calldata + record the workflow invocation in its audit trail. We sign
// and broadcast locally with the verifier's own key. This:
//
//   - Keeps msg.sender == verifier without uploading our key to KH.
//   - Lets us control gas (KH's chain defaults are wrong for 0G Galileo:
//     1.5 gwei tip, but the chain mempool requires 2 gwei).
//   - Still gives KH the prize-relevant audit trail per workflow call.
//
// Pre-requisite: each Purpose's workflow must be LISTED in KH (have a
// listedSlug). See scripts/list-keeperhub-workflows.sh.
type LiveClient struct {
	mcp *MCPClient

	workflowSlugs map[Purpose]string

	verifierKey *ecdsa.PrivateKey
	ethClient   *ethclient.Client
	chainID     *big.Int

	maxAttempts    int
	retryBaseDelay time.Duration
	receiptTimeout time.Duration
}

// ErrLiveWriteNotConfigured is returned by SendTransaction when no
// workflow slug is configured for the requested Purpose.
var ErrLiveWriteNotConfigured = errors.New("keeperhub live: no workflow slug configured for this Purpose; pass via LiveClientConfig.WorkflowSlugs")

// LiveClientConfig is what cmd/verifier passes in.
type LiveClientConfig struct {
	Endpoint string // empty = https://app.keeperhub.com/mcp
	APIKey   string // required

	// WorkflowSlugs maps Purpose → listedSlug (per-org). Slugs come from
	// scripts/list-keeperhub-workflows.sh; persisted in
	// configs/keeperhub.toml.
	WorkflowSlugs map[Purpose]string

	// VerifierKey signs the on-chain tx locally. Should match the org's
	// uploaded wallet so the audit-trail wallet identity matches the
	// on-chain signer (visual-consistency only — KH never sees the key).
	VerifierKey *ecdsa.PrivateKey

	// EthClient handles broadcast + receipt polling.
	EthClient *ethclient.Client
	ChainID   *big.Int

	// MaxAttempts retries the call_workflow → sign → broadcast cycle on
	// transient errors. Default 3.
	MaxAttempts int
	// RetryBaseDelay is the first backoff delay; each attempt doubles it.
	// 0 = 2s. Capped at 30s.
	RetryBaseDelay time.Duration
	// ReceiptTimeout caps how long we wait for on-chain inclusion. 0 = 90s.
	ReceiptTimeout time.Duration
}

// NewLiveClient builds + opens a LiveClient. Caller MUST defer Close.
func NewLiveClient(ctx context.Context, cfg LiveClientConfig) (*LiveClient, error) {
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, errors.New("keeperhub live: APIKey required")
	}
	if cfg.VerifierKey == nil {
		return nil, errors.New("keeperhub live: VerifierKey required (we sign locally)")
	}
	if cfg.EthClient == nil {
		return nil, errors.New("keeperhub live: EthClient required (we broadcast locally)")
	}
	if cfg.ChainID == nil {
		return nil, errors.New("keeperhub live: ChainID required for EIP-155 signing")
	}

	mcp := NewMCPClient(cfg.Endpoint, cfg.APIKey)
	if err := mcp.Open(ctx); err != nil {
		return nil, fmt.Errorf("keeperhub live: handshake: %w", err)
	}

	slugs := cfg.WorkflowSlugs
	if slugs == nil {
		slugs = map[Purpose]string{}
	}
	ma := cfg.MaxAttempts
	if ma <= 0 {
		ma = 3
	}
	rbd := cfg.RetryBaseDelay
	if rbd <= 0 {
		rbd = 2 * time.Second
	}
	rt := cfg.ReceiptTimeout
	if rt <= 0 {
		rt = 90 * time.Second
	}
	return &LiveClient{
		mcp:            mcp,
		workflowSlugs:  slugs,
		verifierKey:    cfg.VerifierKey,
		ethClient:      cfg.EthClient,
		chainID:        cfg.ChainID,
		maxAttempts:    ma,
		retryBaseDelay: rbd,
		receiptTimeout: rt,
	}, nil
}

// Close releases the MCP session.
func (l *LiveClient) Close() {
	if l.mcp != nil {
		l.mcp.Close()
	}
}

// SendTransaction routes a write through KeeperHub.
//
// Per attempt:
//  1. Decode in.Data → typed args for the workflow's trigger.
//  2. call_workflow(slug, inputs) → KH returns {to, data, value}.
//  3. Sign and broadcast locally with verifier's key (we control gas).
//  4. Wait for the receipt.
//
// Transient errors retry with exponential backoff; permanent ones fail
// fast. Contract reverts surface as permanent.
func (l *LiveClient) SendTransaction(ctx context.Context, in SendTxInput) (Receipt, error) {
	slug, ok := l.workflowSlugs[in.Metadata.Purpose]
	if !ok || slug == "" {
		return Receipt{}, ErrLiveWriteNotConfigured
	}

	purpose, args, err := DecodeCallArgs(in.Data)
	if err != nil {
		return Receipt{}, fmt.Errorf("decode calldata: %w", err)
	}
	if purpose != in.Metadata.Purpose {
		return Receipt{}, fmt.Errorf("purpose mismatch: calldata says %q, metadata says %q",
			purpose, in.Metadata.Purpose)
	}

	inputs := make(map[string]any, len(args))
	for k, v := range args {
		inputs[k] = v
	}

	start := time.Now()
	var lastErr error
	for attempt := 1; attempt <= l.maxAttempts; attempt++ {
		receipt, err := l.callOnce(ctx, slug, inputs, in, start)
		if err == nil {
			receipt.Attempts = attempt
			return receipt, nil
		}
		lastErr = err
		if !isTransientError(err) {
			return Receipt{}, err
		}
		if attempt == l.maxAttempts {
			break
		}
		delay := l.retryBaseDelay * time.Duration(1<<(attempt-1))
		if delay > 30*time.Second {
			delay = 30 * time.Second
		}
		select {
		case <-ctx.Done():
			return Receipt{}, ctx.Err()
		case <-time.After(delay):
		}
	}
	return Receipt{}, fmt.Errorf("%w (after %d attempts)", lastErr, l.maxAttempts)
}

// callOnce runs one full call_workflow → sign → broadcast → wait cycle.
func (l *LiveClient) callOnce(
	ctx context.Context, slug string, inputs map[string]any, in SendTxInput, start time.Time,
) (Receipt, error) {
	raw, err := l.mcp.CallWorkflow(ctx, slug, inputs)
	if err != nil {
		return Receipt{}, fmt.Errorf("call_workflow: %w", err)
	}
	to, data, value, err := parseCallWorkflowResult(raw)
	if err != nil {
		return Receipt{}, fmt.Errorf("parse call_workflow: %w (raw=%s)", err, string(raw))
	}
	// Sanity check: the unsigned calldata KH built MUST point at the
	// contract the caller asked for. Mismatch = misconfigured workflow.
	if !bytesEqualHex(to, in.To.Bytes()) {
		return Receipt{}, fmt.Errorf("call_workflow returned to=%s, expected %s",
			common.Bytes2Hex(to), in.To.Hex())
	}
	if len(data) < 4 {
		return Receipt{}, fmt.Errorf("call_workflow returned %d-byte calldata (need ≥4)", len(data))
	}

	txHash, err := l.signAndBroadcast(ctx, common.BytesToAddress(to), data, value, in.GasLimit)
	if err != nil {
		return Receipt{}, fmt.Errorf("sign+broadcast: %w", err)
	}

	rcpt, err := l.waitForReceipt(ctx, txHash)
	if err != nil {
		return Receipt{}, err
	}
	rcpt.TotalLatencyMs = time.Since(start).Milliseconds()
	return rcpt, nil
}

// signAndBroadcast builds, signs, and submits an EIP-1559 tx with gas
// values appropriate for 0G Galileo. We set the priority fee at 2 gwei
// minimum (chain requirement) and let baseFee*2 + tip set the fee cap,
// which is the canonical safety margin.
func (l *LiveClient) signAndBroadcast(
	ctx context.Context, to common.Address, data []byte, value *big.Int, gasLimitHint uint64,
) (common.Hash, error) {
	from := crypto.PubkeyToAddress(l.verifierKey.PublicKey)

	nonce, err := l.ethClient.PendingNonceAt(ctx, from)
	if err != nil {
		return common.Hash{}, fmt.Errorf("nonce: %w", err)
	}

	head, err := l.ethClient.HeaderByNumber(ctx, nil)
	if err != nil {
		return common.Hash{}, fmt.Errorf("header: %w", err)
	}
	tip := big.NewInt(2_000_000_000) // 2 gwei — 0G Galileo's mempool minimum
	feeCap := new(big.Int).Add(new(big.Int).Mul(head.BaseFee, big.NewInt(2)), tip)

	gasLimit := gasLimitHint
	if gasLimit == 0 {
		// Defensive default. Estimation against revert-likely calls is
		// fragile; we'd rather over-pay slightly than fail to broadcast.
		gasLimit = 500_000
	}

	if value == nil {
		value = big.NewInt(0)
	}

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   l.chainID,
		Nonce:     nonce,
		GasTipCap: tip,
		GasFeeCap: feeCap,
		Gas:       gasLimit,
		To:        &to,
		Value:     value,
		Data:      data,
	})

	signed, err := types.SignTx(tx, types.LatestSignerForChainID(l.chainID), l.verifierKey)
	if err != nil {
		return common.Hash{}, fmt.Errorf("sign: %w", err)
	}
	if err := l.ethClient.SendTransaction(ctx, signed); err != nil {
		return common.Hash{}, fmt.Errorf("send: %w", err)
	}
	return signed.Hash(), nil
}

// waitForReceipt polls TransactionReceipt until the tx lands or the
// timeout fires. Status==0 surfaces as a "execution reverted" error
// (so isTransientError treats it as permanent).
func (l *LiveClient) waitForReceipt(ctx context.Context, txHash common.Hash) (Receipt, error) {
	deadline := time.Now().Add(l.receiptTimeout)
	tick := time.NewTicker(2 * time.Second)
	defer tick.Stop()
	for {
		rcpt, err := l.ethClient.TransactionReceipt(ctx, txHash)
		if err == nil {
			if rcpt.Status == types.ReceiptStatusFailed {
				return Receipt{}, fmt.Errorf("execution reverted: tx=%s block=%d", txHash.Hex(), rcpt.BlockNumber.Uint64())
			}
			return Receipt{
				TxHash:      rcpt.TxHash,
				BlockNumber: rcpt.BlockNumber.Uint64(),
				GasUsed:     rcpt.GasUsed,
				Attempts:    1,
			}, nil
		}
		// "not found" == still pending; anything else propagates.
		if !strings.Contains(err.Error(), "not found") {
			return Receipt{}, fmt.Errorf("receipt fetch: %w", err)
		}
		if time.Now().After(deadline) {
			return Receipt{}, fmt.Errorf("receipt timeout: tx=%s did not land within %s", txHash.Hex(), l.receiptTimeout)
		}
		select {
		case <-ctx.Done():
			return Receipt{}, ctx.Err()
		case <-tick.C:
		}
	}
}

// isTransientError reports whether an error is worth retrying. We retry
// on RPC-layer timeouts/5xx (KH's dRPC free tier, transient 0G hiccups)
// but NOT on contract reverts, validation errors, or auth errors.
func isTransientError(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	switch {
	case strings.Contains(s, "execution reverted"),
		strings.Contains(s, "Cannot convert"),
		strings.Contains(s, "is not a function"),
		strings.Contains(s, "Invalid function arguments"),
		strings.Contains(s, "ErrLiveWriteNotConfigured"),
		strings.Contains(s, "purpose mismatch"),
		strings.Contains(s, "decode calldata"),
		strings.Contains(s, "401"),
		strings.Contains(s, "403"):
		return false
	}
	switch {
	case strings.Contains(s, "408"),
		strings.Contains(s, "Request timeout"),
		strings.Contains(s, "free tier"),
		strings.Contains(s, "502"),
		strings.Contains(s, "503"),
		strings.Contains(s, "504"),
		strings.Contains(s, "EOF"),
		strings.Contains(s, "i/o timeout"),
		strings.Contains(s, "connection reset"),
		strings.Contains(s, "no such host"),
		strings.Contains(s, "receipt timeout"):
		return true
	}
	return false
}

// GetAuditTrail returns an empty trail — for the LiveClient path the
// authoritative audit trail lives in KH's dashboard (one workflow
// invocation per call) plus on-chain logs we emit ourselves.
func (l *LiveClient) GetAuditTrail(ctx context.Context, jobID uint64) ([]AuditEntry, error) {
	return nil, nil
}

// RawToolCall is the escape hatch for read-only MCP tools that don't
// fit the SendTransaction shape — used by cmd/keeperhub-demo.
func (l *LiveClient) RawToolCall(
	ctx context.Context, name string, args any,
) (json.RawMessage, error) {
	return l.mcp.ToolCall(ctx, name, args)
}

// ListTools exposes the underlying tools/list call.
func (l *LiveClient) ListTools(ctx context.Context) (json.RawMessage, error) {
	return l.mcp.ListTools(ctx)
}

// Endpoint + HasAPIKey forwarders for startup logging.
func (l *LiveClient) Endpoint() string { return l.mcp.Endpoint() }
func (l *LiveClient) HasAPIKey() bool  { return l.mcp.HasAPIKey() }

// parseCallWorkflowResult unwraps the call_workflow MCP envelope and
// extracts the (to, data, value) tuple KH returns for write workflows.
//
// Expected payload (after MCP text-unwrap):
//
//	{
//	  "to":   "0x...",
//	  "data": "0x...",
//	  "value": "0" | "1000000000000000000" | "0x..."
//	}
//
// Any of the three fields may live one level deep under "result" or
// "calldata" depending on KH version; we probe a few shapes.
func parseCallWorkflowResult(raw json.RawMessage) (to, data []byte, value *big.Int, err error) {
	inner, err := unwrapMCPText(raw)
	if err != nil {
		return nil, nil, nil, err
	}
	candidates := []map[string]any{inner}
	for _, k := range []string{"result", "calldata", "output", "data"} {
		if sub, ok := inner[k].(map[string]any); ok {
			candidates = append(candidates, sub)
		}
	}
	for _, m := range candidates {
		toS, _ := m["to"].(string)
		dataS, _ := m["data"].(string)
		valueS, _ := m["value"].(string)
		if toS == "" || dataS == "" {
			continue
		}
		to, err = decodeHex(toS)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("decode to: %w", err)
		}
		data, err = decodeHex(dataS)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("decode data: %w", err)
		}
		value = parseValue(valueS)
		return to, data, value, nil
	}
	return nil, nil, nil, fmt.Errorf("no {to, data} found in payload (keys=%v)", mapKeys(inner))
}

// parseExecutionID is retained for any caller still using execute_workflow
// (e.g. test helpers); LiveClient.callOnce no longer uses it.
func parseExecutionID(raw json.RawMessage) (string, error) {
	inner, err := unwrapMCPText(raw)
	if err != nil {
		return "", err
	}
	for _, k := range []string{"executionId", "execution_id", "id"} {
		if v, ok := inner[k].(string); ok && v != "" {
			return v, nil
		}
	}
	return "", fmt.Errorf("no executionId field in payload (keys=%v)", mapKeys(inner))
}

// unwrapMCPText handles the standard MCP wrapping
// `{content: [{type: "text", text: "<json>"}]}` and returns the inner
// JSON object. If the raw payload is already a flat object, returns it.
func unwrapMCPText(raw json.RawMessage) (map[string]any, error) {
	var outer struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(raw, &outer); err == nil && len(outer.Content) > 0 && outer.Content[0].Text != "" {
		var inner map[string]any
		if err := json.Unmarshal([]byte(outer.Content[0].Text), &inner); err != nil {
			return nil, fmt.Errorf("parse inner text: %w", err)
		}
		return inner, nil
	}
	var flat map[string]any
	if err := json.Unmarshal(raw, &flat); err != nil {
		return nil, fmt.Errorf("parse flat: %w", err)
	}
	return flat, nil
}

func mapKeys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func decodeHex(s string) ([]byte, error) {
	return hex.DecodeString(strings.TrimPrefix(s, "0x"))
}

func bytesEqualHex(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// parseValue accepts decimal or 0x-hex; empty/missing → 0.
func parseValue(s string) *big.Int {
	s = strings.TrimSpace(s)
	if s == "" || s == "0" {
		return big.NewInt(0)
	}
	if strings.HasPrefix(s, "0x") {
		v, ok := new(big.Int).SetString(strings.TrimPrefix(s, "0x"), 16)
		if !ok {
			return big.NewInt(0)
		}
		return v
	}
	v, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return big.NewInt(0)
	}
	return v
}
