package ogstorage

import (
	"context"
	"errors"
)

// Storage is the surface verifier code uses to persist agent metadata
// and vote history. Two implementations satisfy it:
//   - LiveStorage   — wires the real 0G Storage SDK; deferred to step 9
//     when we provision the testnet deployer wallet
//
// All methods take a context so callers can bound retries / timeouts.
// Methods are idempotent where possible — re-uploading the same agent
// card returns the same URI.
type Storage interface {
	// UploadAgentCard writes the card to storage and returns the URI
	// the iNFT should point at. The URI scheme is implementation-defined
	// (mock uses "mock://card/<id>"; production uses "0g://...").
	UploadAgentCard(ctx context.Context, card AgentCard) (uri string, err error)

	// AppendVoteRecord adds one row to the verifier's per-tokenId Log.
	// Returns the URI of the appended entry (some backends call this
	// the "log id" or "stream id"; opaque to callers).
	AppendVoteRecord(ctx context.Context, verifierTokenID uint64, record VoteRecord) (entryURI string, err error)

	// GetAccuracyHistory replays every VoteRecord ever appended for the
	// given tokenId, in chronological order. Used by external auditors
	// to verify on-chain accuracy stats match what the Log says.
	GetAccuracyHistory(ctx context.Context, verifierTokenID uint64) ([]VoteRecord, error)

	// UploadProofBundle writes a per-job ProofBundle. Returns the URI
	// (callers usually emit this on-chain or attach it to JobSettled
	// off-chain notifications).
	UploadProofBundle(ctx context.Context, bundle ProofBundle) (uri string, err error)
}

// ErrUnimplemented is returned by stub storage implementations whose
// real backing isn't wired up yet (step 9).
var ErrUnimplemented = errors.New("ogstorage: not yet implemented (live SDK wiring is step 9)")
