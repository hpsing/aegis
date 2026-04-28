package envelope

import (
	"encoding/json"
	"time"
)

// Wire-format constants. Bump Version if any field is added/renamed/removed
// in the JSON shape. Never bump Version for signing-scheme changes —
// that's what SigVersion is for.
const (
	Version            = 1
	FreshnessWindow    = 60 * time.Second
	TypeExecutionClaim = "execution_claim"
	TypeVerifierVote   = "verifier_vote"
)

// Envelope is the FROZEN outer wrapper for every signed message.
type Envelope struct {
	V         int             `json:"v"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	Ts        int64           `json:"ts"`
	Nonce     string          `json:"nonce"`
	SignerPub string          `json:"signer_pub"`
	Sig       string          `json:"sig,omitempty"`
}

// VerifyResult is the outcome of Verify(). On failure, Reason is a stable
// machine-parseable string used by verifier loops for log filtering.
type VerifyResult struct {
	OK     bool
	Reason string
}

// envelopeToMap re-shapes an Envelope into the same JSON object the
// receiver will see, minus the sig field, so canonicalization is
// unambiguous on both sides.
func envelopeToMap(e Envelope) map[string]any {
	var payload any
	// Payload is set by Sign() to valid JSON, so this never fails in
	// practice. If it ever does, signing produces a self-detecting
	// invalid envelope that fails Verify with bad_signature on the receiver.
	_ = json.Unmarshal(e.Payload, &payload)
	return map[string]any{
		"v":          e.V,
		"type":       e.Type,
		"payload":    payload,
		"ts":         e.Ts,
		"nonce":      e.Nonce,
		"signer_pub": e.SignerPub,
	}
}
