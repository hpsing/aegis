package keeperhub

import (
	"context"
	"errors"
)

// Client is the surface verifier code talks to. Both MockClient and
// LiveClient satisfy it.
type Client interface {
	// SendTransaction submits a tx for reliable execution. Resolves when
	// the tx has landed on chain (or definitively failed). Idempotent on
	// Metadata.IdempotencyKey — duplicate calls return the cached receipt.
	SendTransaction(ctx context.Context, in SendTxInput) (Receipt, error)

	// GetAuditTrail returns every keeper action recorded for the job.
	// Used by the proof-bundle builder and (in production) by
	// auditors querying KeeperHub's dashboard.
	GetAuditTrail(ctx context.Context, jobID uint64) ([]AuditEntry, error)
}

// ErrUnimplemented is returned by stub Client methods until the live
// MCP wiring lands (Phase D).
var ErrUnimplemented = errors.New("keeperhub: not yet implemented")

// DefaultIdempotencyKey returns the canonical key for a (jobID, purpose)
// pair. Caller can override Metadata.IdempotencyKey if they need finer
// granularity (e.g. retry attempts on the same logical step).
func DefaultIdempotencyKey(jobID uint64, p Purpose) string {
	return formatKey(jobID, p)
}
