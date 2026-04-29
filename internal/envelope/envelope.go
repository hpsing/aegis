package envelope

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// Wire-format constants.
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

// Sign canonicalizes the inner envelope (without sig) and secp256k1-signs it.
func Sign(typ string, payload any, kp Keypair) (Envelope, error) {
	pb, err := json.Marshal(payload)
	if err != nil {
		return Envelope{}, fmt.Errorf("marshal payload: %w", err)
	}
	nb := make([]byte, 16)
	if _, err := rand.Read(nb); err != nil {
		return Envelope{}, fmt.Errorf("nonce: %w", err)
	}
	e := Envelope{
		V:         Version,
		Type:      typ,
		Payload:   pb,
		Ts:        time.Now().UnixMilli(),
		Nonce:     hex.EncodeToString(nb),
		SignerPub: kp.PublicHex(),
	}
	sig, err := signCanonical(envelopeToMap(e), kp.Priv)
	if err != nil {
		return Envelope{}, err
	}
	e.Sig = sig
	return e, nil
}

// VerifyResult is the outcome of Verify(). On failure, Reason is a stable
// machine-parseable string used by verifier loops for log filtering.
type VerifyResult struct {
	OK     bool
	Reason string
}

// Verify checks version, freshness, and signature against signer_pub.
func Verify(e Envelope) VerifyResult {
	if e.V != Version {
		return VerifyResult{false, "unsupported_version"}
	}
	age := time.Since(time.UnixMilli(e.Ts))
	if age > FreshnessWindow || age < -FreshnessWindow {
		return VerifyResult{false, "stale_or_future"}
	}
	if e.Sig == "" || e.SignerPub == "" {
		return VerifyResult{false, "missing_sig_or_signer"}
	}
	ok, err := verifyCanonical(envelopeToMap(e), e.Sig, e.SignerPub)
	if err != nil {
		return VerifyResult{false, "bad_sig_encoding"}
	}
	if !ok {
		return VerifyResult{false, "bad_signature"}
	}
	return VerifyResult{OK: true}
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
