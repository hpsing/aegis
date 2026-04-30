package keeperhub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// LiveClient is the production Client backed by the real KeeperHub MCP
// at https://app.keeperhub.com/mcp.
//
// Architecture (Option C):
//
//   - Each verifier process runs against its own KH org. The org has one
//     wallet integration (the verifier's EOA). KH signs writes with that
//     wallet, so on-chain msg.sender == the right verifier.
//   - The org has three pre-published workflows (commit, reveal, settle)
//     wired to web3/write-contract against AegisContract on 0G Galileo.
//   - SendTransaction takes the verifier's pre-built calldata, decodes
//     it back to typed args, calls execute_workflow against the matching
//     workflow id, and polls get_execution_status until terminal.
//
// FEEDBACK.md captures the trust caveat: the verifier's signing key
// lives in KH's vault. For production we'd want per-actor wallet
// variants on a single shared org; for the hackathon it's the simplest
// path that preserves msg.sender semantics across all 7 settlement-
// critical txs per job.
type LiveClient struct {
	mcp         *MCPClient
	workflowIDs map[Purpose]string

	// Polling controls.
	pollInterval time.Duration
	pollDeadline time.Duration

	// Retry controls.
	maxAttempts    int
	retryBaseDelay time.Duration
}

// ErrLiveWriteNotConfigured is returned by SendTransaction when no
// workflow id is configured for the requested Purpose.
var ErrLiveWriteNotConfigured = errors.New("keeperhub live: no workflow id configured for this Purpose; pass via LiveClientConfig.WorkflowIDs")

// LiveClientConfig is what cmd/verifier passes in.
type LiveClientConfig struct {
	Endpoint     string             // empty = https://app.keeperhub.com/mcp
	APIKey       string             // required
	WorkflowIDs  map[Purpose]string // commit/reveal/settle ids for this verifier's org
	PollInterval time.Duration      // 0 = 2s
	PollDeadline time.Duration      // 0 = 90s
	// MaxAttempts retries the whole execute→poll cycle on transient
	// errors (e.g. KH's dRPC free tier 408ing). Default 3.
	MaxAttempts int
	// RetryBaseDelay is the first backoff delay; each attempt doubles it.
	// 0 = 2s. Capped at 30s.
	RetryBaseDelay time.Duration
}

// NewLiveClient builds + opens a LiveClient. Caller MUST defer Close.
func NewLiveClient(ctx context.Context, cfg LiveClientConfig) (*LiveClient, error) {
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, errors.New("keeperhub live: APIKey required")
	}
	mcp := NewMCPClient(cfg.Endpoint, cfg.APIKey)
	if err := mcp.Open(ctx); err != nil {
		return nil, fmt.Errorf("keeperhub live: handshake: %w", err)
	}
	ids := cfg.WorkflowIDs
	if ids == nil {
		ids = map[Purpose]string{}
	}
	pi := cfg.PollInterval
	if pi <= 0 {
		pi = 2 * time.Second
	}
	pd := cfg.PollDeadline
	if pd <= 0 {
		pd = 90 * time.Second
	}
	ma := cfg.MaxAttempts
	if ma <= 0 {
		ma = 3
	}
	rbd := cfg.RetryBaseDelay
	if rbd <= 0 {
		rbd = 2 * time.Second
	}
	return &LiveClient{
		mcp:            mcp,
		workflowIDs:    ids,
		pollInterval:   pi,
		pollDeadline:   pd,
		maxAttempts:    ma,
		retryBaseDelay: rbd,
	}, nil
}

// Close releases the MCP session.
func (l *LiveClient) Close() {
	if l.mcp != nil {
		l.mcp.Close()
	}
}

// SendTransaction routes a write through a KeeperHub workflow.
//
// Flow per attempt:
//  1. execute_workflow(workflowId, input).
//  2. Poll get_execution_status until terminal.
//
// On transient errors (KH's dRPC free tier 408 timeouts, network blips,
// 5xx) we retry up to maxAttempts with exponential backoff. Permanent
// errors (contract reverts, bad calldata, missing workflow) fail fast.
func (l *LiveClient) SendTransaction(ctx context.Context, in SendTxInput) (Receipt, error) {
	workflowID, ok := l.workflowIDs[in.Metadata.Purpose]
	if !ok || workflowID == "" {
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

	input := make(map[string]any, len(args))
	for k, v := range args {
		input[k] = v
	}

	start := time.Now()
	var lastErr error
	for attempt := 1; attempt <= l.maxAttempts; attempt++ {
		receipt, err := l.executeOnce(ctx, workflowID, input, start)
		if err == nil {
			receipt.Attempts = attempt
			return receipt, nil
		}
		lastErr = err
		if !isTransientError(err) {
			// Permanent error — fail fast, no retry.
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

// executeOnce is one attempt of the execute→poll cycle.
func (l *LiveClient) executeOnce(
	ctx context.Context, workflowID string, input map[string]any, start time.Time,
) (Receipt, error) {
	rawExec, err := l.mcp.ExecuteWorkflow(ctx, workflowID, input)
	if err != nil {
		return Receipt{}, fmt.Errorf("execute_workflow: %w", err)
	}
	executionID, err := parseExecutionID(rawExec)
	if err != nil {
		return Receipt{}, fmt.Errorf("parse execute_workflow response: %w (raw=%s)", err, string(rawExec))
	}
	return l.pollUntilTerminal(ctx, executionID, start)
}

// isTransientError reports whether an error is worth retrying. We retry
// on RPC-layer timeouts/5xx (KH's dRPC free tier) but NOT on contract
// reverts, validation errors, or auth errors — those are permanent.
func isTransientError(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	// Permanent: contract reverts, calldata bugs, schema bugs.
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
	// Transient: timeouts, 5xx, and dRPC's hand-rolled rate-limit message.
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
		strings.Contains(s, "no such host"):
		return true
	}
	// Default: assume permanent. Better to surface a real error than to
	// silently retry a bug-class fault and look slow.
	return false
}

// GetAuditTrail today returns an empty trail rather than erroring, so
// the ProofBundle builder can run on either client. A future revision
// can call list_workflow_executions filtered by jobID metadata.
func (l *LiveClient) GetAuditTrail(ctx context.Context, jobID uint64) ([]AuditEntry, error) {
	return nil, nil
}

// RawToolCall is the escape hatch for read-only MCP tools that don't
// fit the SendTransaction shape — used by cmd/keeperhub-demo to
// exercise tools like list_action_schemas, search_workflows, etc.
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

// pollUntilTerminal repeatedly calls get_execution_status until status
// is one of: succeeded, failed, cancelled. Returns a Receipt built from
// the final payload, or an error if the deadline elapses.
func (l *LiveClient) pollUntilTerminal(
	ctx context.Context, executionID string, start time.Time,
) (Receipt, error) {
	deadline := time.Now().Add(l.pollDeadline)
	ticker := time.NewTicker(l.pollInterval)
	defer ticker.Stop()

	for {
		raw, err := l.mcp.GetExecutionStatus(ctx, executionID)
		if err != nil {
			return Receipt{}, fmt.Errorf("get_execution_status: %w", err)
		}
		status, txHash, ok, errMsg, parseErr := parseStatus(raw)
		if parseErr != nil {
			return Receipt{}, fmt.Errorf("parse status: %w (raw=%s)", parseErr, string(raw))
		}
		switch status {
		case "succeeded", "completed", "success":
			return Receipt{
				TxHash:         common.HexToHash(txHash),
				TotalLatencyMs: time.Since(start).Milliseconds(),
				Attempts:       1,
			}, nil
		case "failed", "error", "cancelled":
			if errMsg == "" {
				errMsg = "(no error message in status payload)"
			}
			return Receipt{}, fmt.Errorf("execution %s %s: %s", executionID, status, errMsg)
		}
		_ = ok // status is non-terminal; keep polling

		if time.Now().After(deadline) {
			return Receipt{}, fmt.Errorf("execution %s did not terminate within %s (last status=%q)",
				executionID, l.pollDeadline, status)
		}
		select {
		case <-ctx.Done():
			return Receipt{}, ctx.Err()
		case <-ticker.C:
		}
	}
}

// parseExecutionID extracts the executionId from the raw execute_workflow
// response. KH wraps results as {content: [{type:"text", text:"<json>"}]}.
// We unwrap once, parse the inner JSON, and try a few likely keys.
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

// parseStatus extracts (status, transactionHash, succeeded?, errorMessage)
// from the raw get_execution_status response.
//
// KeeperHub status payload shape (observed):
//
//	{
//	  "status": "error" | "running" | "success" | ...,
//	  "nodeStatuses": [{nodeId, status, output?}],
//	  "errorContext": {failedNodeId, error: "..."},
//	  ...
//	}
//
// We scan nodeStatuses[*].output for a transactionHash on success and
// errorContext.error for the failure reason.
func parseStatus(raw json.RawMessage) (status, txHash string, ok bool, errMsg string, err error) {
	inner, err := unwrapMCPText(raw)
	if err != nil {
		return "", "", false, "", err
	}
	if s, ok := inner["status"].(string); ok {
		status = strings.ToLower(s)
	} else if s, ok := inner["state"].(string); ok {
		status = strings.ToLower(s)
	}
	// Prefer the dedicated errorContext.error.
	if ec, ok := inner["errorContext"].(map[string]any); ok {
		if v, ok := ec["error"].(string); ok && v != "" {
			errMsg = v
		}
	}
	if errMsg == "" {
		if v, ok := inner["error"].(string); ok {
			errMsg = v
		}
	}
	// Scan nodeStatuses for an action node with a transactionHash output.
	if arr, ok := inner["nodeStatuses"].([]any); ok {
		for _, n := range arr {
			node, ok := n.(map[string]any)
			if !ok {
				continue
			}
			out, _ := node["output"].(map[string]any)
			for _, k := range []string{"transactionHash", "txHash", "tx_hash"} {
				if v, ok := out[k].(string); ok && v != "" {
					txHash = v
					break
				}
			}
			if txHash != "" {
				break
			}
		}
	}
	// Fallback: top-level / result / output.
	if txHash == "" {
		for _, k := range []string{"transactionHash", "txHash", "tx_hash"} {
			if v, ok := inner[k].(string); ok && v != "" {
				txHash = v
				break
			}
		}
	}
	if txHash == "" {
		for _, k := range []string{"output", "result"} {
			if sub, ok := inner[k].(map[string]any); ok {
				for _, k2 := range []string{"transactionHash", "txHash", "tx_hash"} {
					if v, ok := sub[k2].(string); ok && v != "" {
						txHash = v
						break
					}
				}
			}
		}
	}
	ok = (status == "succeeded" || status == "completed" || status == "success")
	return status, txHash, ok, errMsg, nil
}

// unwrapMCPText handles the standard MCP response wrapping
// `{content: [{type: "text", text: "<json>"}]}` and returns the inner
// JSON object. If the raw payload is already a flat object, returns it
// directly.
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
