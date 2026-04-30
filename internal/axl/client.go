// Gensyn AXL daemon's HTTP API for use by publisher / verifier / executor agents.(default 9002)
//
//	GET  /topology                 → node's peer/tree state
//	POST /send                     → fire-and-forget binary message
//	GET  /recv                     → poll inbound queue (204 if empty)
//
// We don't expose /mcp or /a2a here; Aegis swarm uses raw /send + /recv
// for spec distribution and vote envelope replication. JSON-RPC framing
// would add overhead we don't need.
package axl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client is the Go client for one local AXL node.
type Client struct {
	baseURL string
	http    *http.Client
}

// NewClient builds a client. baseURL is the local AXL node's HTTP API,
// e.g. http://127.0.0.1:9002.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{},
	}
}

// Topology is the response shape of GET /topology. Field names match
// the daemon's JSON output verbatim.
type Topology struct {
	OurIPv6      string          `json:"our_ipv6"`
	OurPublicKey string          `json:"our_public_key"`
	Peers        json.RawMessage `json:"peers"`
	Tree         json.RawMessage `json:"tree"`
}

// Topology fetches /topology. The daemon returns immediately even if no
// peers are connected yet (Peers/Tree may be empty arrays).
func (c *Client) Topology(ctx context.Context) (Topology, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/topology", nil)
	if err != nil {
		return Topology{}, err
	}
	res, err := c.http.Do(req)
	if err != nil {
		return Topology{}, fmt.Errorf("topology: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return Topology{}, fmt.Errorf("topology: %s (%d): %s", res.Status, res.StatusCode, body)
	}
	var t Topology
	if err := json.NewDecoder(res.Body).Decode(&t); err != nil {
		return Topology{}, fmt.Errorf("topology decode: %w", err)
	}
	return t, nil
}

// PeerID returns this node's hex-encoded ed25519 public key — the
// addressable identity for /send. Cached after the first call.
func (c *Client) PeerID(ctx context.Context) (string, error) {
	t, err := c.Topology(ctx)
	if err != nil {
		return "", err
	}
	return t.OurPublicKey, nil
}

// Send fires a binary payload at the given peer. Fire-and-forget: the
// daemon enqueues it and returns 200 OK. Delivery is best-effort over
// Yggdrasil; we don't get receipt confirmation.
func (c *Client) Send(ctx context.Context, destPeerID string, payload []byte) error {
	if destPeerID == "" {
		return fmt.Errorf("axl send: empty destPeerID")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/send", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("X-Destination-Peer-Id", destPeerID)
	req.Header.Set("Content-Type", "application/octet-stream")
	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("axl send: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("axl send: %s (%d): %s", res.Status, res.StatusCode, body)
	}
	return nil
}

// Message is one inbound message from /recv.
type Message struct {
	FromPeerID string
	Payload    []byte
}

// Recv polls for one inbound message. Returns (nil, nil) if the queue
// is empty (HTTP 204). The daemon never blocks; callers wanting a poll
// loop should sleep between calls.
func (c *Client) Recv(ctx context.Context) (*Message, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/recv", nil)
	if err != nil {
		return nil, err
	}
	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("axl recv: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNoContent {
		return nil, nil
	}
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("axl recv: %s (%d): %s", res.Status, res.StatusCode, body)
	}
	from := res.Header.Get("X-From-Peer-Id")
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("axl recv body: %w", err)
	}
	return &Message{FromPeerID: from, Payload: body}, nil
}

// BaseURL returns the configured node URL (for logging).
func (c *Client) BaseURL() string { return c.baseURL }
