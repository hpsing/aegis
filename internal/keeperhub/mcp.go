package keeperhub

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
)

// MCP HTTP transport for KeeperHub at https://app.keeperhub.com/mcp.
//
// Lifecycle:
//
//	1. Open() sends `initialize` and captures Mcp-Session-Id from the
//	   response headers. Required before any tools/call.
//	2. ToolCall(name, args) issues a tools/call request including the
//	   session header.
//	3. Close() best-effort terminates the session.
//
// We don't use a third-party MCP client because (a) the wire is one
// JSON-RPC envelope per call and (b) we want explicit control over
// the session-id capture quirk. See FEEDBACK.md for our notes on the
// MCP transport ergonomics.

const (
	defaultMCPURL        = "https://app.keeperhub.com/mcp"
	mcpProtocolVersion   = "2025-03-26"
	mcpClientName        = "quorum-verifier"
	mcpClientVersion     = "0.1.0"
	mcpSessionHeaderName = "Mcp-Session-Id"
	mcpAcceptHeader      = "application/json, text/event-stream"
)

// MCPClient is a stateful KeeperHub MCP client. NOT goroutine-safe at
// the request level; callers should serialize ToolCall invocations.
type MCPClient struct {
	endpoint string
	apiKey   string
	http     *http.Client

	mu        sync.Mutex
	sessionID string
	nextID    atomic.Int64
}

// NewMCPClient builds a client. endpoint defaults to
// https://app.keeperhub.com/mcp if empty.
func NewMCPClient(endpoint, apiKey string) *MCPClient {
	if endpoint == "" {
		endpoint = defaultMCPURL
	}
	return &MCPClient{
		endpoint: endpoint,
		apiKey:   apiKey,
		http:     &http.Client{},
	}
}

type mcpRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int64  `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type mcpResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type initializeParams struct {
	ProtocolVersion string         `json:"protocolVersion"`
	Capabilities    map[string]any `json:"capabilities"`
	ClientInfo      struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"clientInfo"`
}

// Open performs the MCP handshake. Captures the session id from the
// response headers — the server requires it on every subsequent call.
func (c *MCPClient) Open(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.sessionID != "" {
		return nil // already open
	}

	params := initializeParams{
		ProtocolVersion: mcpProtocolVersion,
		Capabilities:    map[string]any{},
	}
	params.ClientInfo.Name = mcpClientName
	params.ClientInfo.Version = mcpClientVersion

	id := c.nextID.Add(1)
	body, err := json.Marshal(mcpRequest{
		JSONRPC: "2.0", ID: id, Method: "initialize", Params: params,
	})
	if err != nil {
		return fmt.Errorf("marshal initialize: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", mcpAcceptHeader)

	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	defer res.Body.Close()

	sid := res.Header.Get(mcpSessionHeaderName)
	if sid == "" {
		return fmt.Errorf("initialize: server did not return %s header (status %d)", mcpSessionHeaderName, res.StatusCode)
	}

	respBytes, _ := io.ReadAll(res.Body)
	respBytes = stripSSEPrefix(respBytes)
	var resp mcpResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return fmt.Errorf("initialize parse: %w (body=%q)", err, string(respBytes))
	}
	if resp.Error != nil {
		return fmt.Errorf("initialize error: %s (code=%d)", resp.Error.Message, resp.Error.Code)
	}
	c.sessionID = sid
	return nil
}

// ToolCall invokes a registered MCP tool. `args` is marshalled as the
// `arguments` field. Returns the raw `result` payload — caller decodes.
func (c *MCPClient) ToolCall(ctx context.Context, name string, args any) (json.RawMessage, error) {
	c.mu.Lock()
	sid := c.sessionID
	c.mu.Unlock()
	if sid == "" {
		return nil, fmt.Errorf("mcp: not initialized; call Open first")
	}

	id := c.nextID.Add(1)
	params := struct {
		Name      string `json:"name"`
		Arguments any    `json:"arguments,omitempty"`
	}{Name: name, Arguments: args}

	body, err := json.Marshal(mcpRequest{
		JSONRPC: "2.0", ID: id, Method: "tools/call", Params: params,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal tools/call: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", mcpAcceptHeader)
	req.Header.Set(mcpSessionHeaderName, sid)

	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tools/call %s: %w", name, err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	raw = stripSSEPrefix(raw)
	var resp mcpResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("tools/call %s parse: %w (body=%q)", name, err, string(raw))
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("tools/call %s: %s (code=%d)", name, resp.Error.Message, resp.Error.Code)
	}
	return resp.Result, nil
}

// ListTools is shorthand for `tools/list` — useful for verifying
// connectivity at startup. Returns the raw JSON for caller inspection.
func (c *MCPClient) ListTools(ctx context.Context) (json.RawMessage, error) {
	c.mu.Lock()
	sid := c.sessionID
	c.mu.Unlock()
	if sid == "" {
		return nil, fmt.Errorf("mcp: not initialized")
	}

	id := c.nextID.Add(1)
	body, err := json.Marshal(mcpRequest{JSONRPC: "2.0", ID: id, Method: "tools/list"})
	if err != nil {
		return nil, err
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", mcpAcceptHeader)
	req.Header.Set(mcpSessionHeaderName, sid)

	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tools/list: %w", err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	raw = stripSSEPrefix(raw)
	var resp mcpResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("tools/list: %s", resp.Error.Message)
	}
	return resp.Result, nil
}

// ExecuteWorkflow triggers a manual run of `workflowID` with the given
// trigger inputs. Returns the raw `result` payload — caller decodes the
// executionId out of it (KH wraps it as {content: [{text: "<json>"}]}).
func (c *MCPClient) ExecuteWorkflow(
	ctx context.Context, workflowID string, input map[string]any,
) (json.RawMessage, error) {
	args := map[string]any{"workflowId": workflowID}
	if input != nil {
		args["input"] = input
	}
	return c.ToolCall(ctx, "execute_workflow", args)
}

// GetExecutionStatus polls a single execution by id. Same wrapping
// (text-wrapped JSON) — caller decodes the status field.
func (c *MCPClient) GetExecutionStatus(
	ctx context.Context, executionID string,
) (json.RawMessage, error) {
	return c.ToolCall(ctx, "get_execution_status", map[string]any{"executionId": executionID})
}

// CallWorkflow invokes a LISTED workflow by its `listedSlug`. Per the KH
// docs: "For write workflows, returns unsigned calldata {to, data, value}
// for the caller to submit."
func (c *MCPClient) CallWorkflow(
	ctx context.Context, slug string, inputs map[string]any,
) (json.RawMessage, error) {
	args := map[string]any{"slug": slug}
	if inputs != nil {
		args["inputs"] = inputs
	}
	return c.ToolCall(ctx, "call_workflow", args)
}

// Close best-effort releases server-side state. Most MCP servers GC
// sessions after inactivity; explicit close is a courtesy.
func (c *MCPClient) Close() {
	c.mu.Lock()
	c.sessionID = ""
	c.mu.Unlock()
}

// stripSSEPrefix handles servers that respond with SSE framing
// (`event: message\ndata: <json>\n\n`) by extracting the JSON body.
// If the body is plain JSON, it's returned unchanged.
func stripSSEPrefix(body []byte) []byte {
	if !bytes.Contains(body, []byte("data: ")) {
		return body
	}
	// Find the data: line(s) and concatenate their payloads.
	var out bytes.Buffer
	for _, line := range bytes.Split(body, []byte("\n")) {
		if bytes.HasPrefix(line, []byte("data: ")) {
			out.Write(bytes.TrimPrefix(line, []byte("data: ")))
		}
	}
	if out.Len() == 0 {
		return body
	}
	return out.Bytes()
}

// SessionID exposes the current session id for logging / debugging.
func (c *MCPClient) SessionID() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.sessionID
}

// Endpoint returns the configured URL (sanitized — no API key).
func (c *MCPClient) Endpoint() string {
	return c.endpoint
}

// HasAPIKey reports whether an API key is configured (useful for
// startup logging without leaking the key).
func (c *MCPClient) HasAPIKey() bool {
	return strings.TrimSpace(c.apiKey) != ""
}
