package axl

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Envelope is the canonical wrapper for application messages flowing
// over AXL. Every payload sent by a Aegis agent is one of these so
// the receiver can route by Type before parsing the body.
//
// We hand-spec the envelope (rather than reusing internal/envelope's
// signing wrapper) because:
//   - AXL already provides authenticated transport peer-to-peer
//   - Aegis's on-chain commit-reveal is the authoritative consensus;
//     these envelopes are the off-chain data layer (job spec + replica
//     of vote data for auditability)
type Envelope struct {
	Type      string          `json:"type"`
	Version   int             `json:"version"`
	Timestamp int64           `json:"ts"` // unix ms
	Payload   json.RawMessage `json:"payload"`
}

const envelopeVersion = 1

// Envelope types. Keep these stable — the verifier loop dispatches on
// them. Any new types should be additive.
const (
	TypeSpecPublish = "spec_publish" // publisher → verifiers, before claim
	TypeVoteCommit  = "vote_commit"  // verifier → swarm, after on-chain commit
	TypeVoteReveal  = "vote_reveal"  // verifier → swarm, after on-chain reveal
	TypeAgentCard   = "agent_card"   // verifier → publisher, on startup (peer registration)
	TypeAck         = "ack"          // pairwise ack on receipt (optional)
)

// SpecPayload is the body of a TypeSpecPublish envelope. The verifier
// validates keccak256(canonical_json(Spec)) == on-chain specHash before
// using it; we bind both fields for robustness against indexing bugs.
type SpecPayload struct {
	JobID    uint64          `json:"jobId"`
	SpecHash string          `json:"specHash"` // 0x-prefixed 32-byte hex
	Spec     json.RawMessage `json:"spec"`     // canonical JSON
}

// VotePayload is the body of TypeVoteCommit and TypeVoteReveal. For
// commit only `CommitHash` is set; for reveal `Verdict` + `Nonce` are
// set. Listeners (other verifiers, the demo UI, auditors) index votes
// by JobID.
type VotePayload struct {
	JobID      uint64 `json:"jobId"`
	Verifier   string `json:"verifier"` // 0x-prefixed verifier EOA
	Phase      string `json:"phase"`    // "commit" | "reveal"
	CommitHash string `json:"commitHash,omitempty"`
	Verdict    *bool  `json:"verdict,omitempty"`
	Nonce      string `json:"nonce,omitempty"` // 0x-prefixed 32-byte hex
	TxHash     string `json:"txHash,omitempty"`
}

// AgentCardPayload is the body of TypeAgentCard. Each verifier announces
// itself to the publisher on startup so the publisher can address spec
// publishes by peer id.
type AgentCardPayload struct {
	Address  string `json:"address"`  // 0x-prefixed verifier EOA
	PeerID   string `json:"peerId"`   // AXL public key (this node's identity)
	Role     string `json:"role"`     // "verifier" | "publisher" | "executor"
	Greeting string `json:"greeting"` // free-form
}

// MarshalEnvelope wraps a payload in an Envelope and serializes it.
func MarshalEnvelope(envType string, payload any) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}
	env := Envelope{
		Type:      envType,
		Version:   envelopeVersion,
		Timestamp: time.Now().UnixMilli(),
		Payload:   body,
	}
	return json.Marshal(env)
}

// UnmarshalEnvelope parses a wire byte slice back to an Envelope.
func UnmarshalEnvelope(raw []byte) (Envelope, error) {
	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return Envelope{}, fmt.Errorf("envelope decode: %w", err)
	}
	if env.Type == "" {
		return Envelope{}, fmt.Errorf("envelope: empty type")
	}
	return env, nil
}

// SpecHashOf computes keccak256-equivalent hash for matching against
// the on-chain specHash. The canonical hash function in our system is
// keccak256, but here we use sha256 for envelope integrity since the
// chain hash is already verified separately by the verifier loop.
//
// The on-chain specHash check uses keccak256 (see internal/envelope).
// This helper is for debug logging only.
func SpecHashOf(canonical []byte) string {
	h := sha256.Sum256(canonical)
	return "0x" + hex.EncodeToString(h[:])
}

// NormalizePeerID strips any 0x prefix and lowercases. The daemon's
// X-Destination-Peer-Id is case-insensitive; we normalize to keep
// configs and logs comparable.
func NormalizePeerID(id string) string {
	return strings.ToLower(strings.TrimPrefix(id, "0x"))
}
