package verifier

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/hpsing/aegis/internal/keeperhub"
	"github.com/hpsing/aegis/internal/ogstorage"
)

// BuildProofBundle assembles a per-job ProofBundle from on-chain state +
// the keeper's audit trail. Called from the post-settlement hook by the
// verifier that lands `settle` (per architecture-doc §10).
//
// Inputs:
//   - jobID         the job that just settled
//   - finalVerdict  PASS=true / FAIL=false from the JobSettled event
//   - settleTxHash  the on-chain settle tx (canonical anchor)
//   - settleBlock   block number of the settle
//   - producedBy    verifier address that built the bundle
//
// The bundle is uploaded to ogstorage.Storage; the returned URI is what
// the demo UI / external auditors point at. Errors are logged but
// non-fatal — the on-chain Registry is the source of truth, the bundle
// is a long-form receipt.
func BuildProofBundle(
	ctx context.Context,
	storage ogstorage.Storage,
	keeper keeperhub.Client,
	job Job,
	jobID *big.Int,
	finalVerdict bool,
	settleTxHashHex string,
	settleBlock uint64,
	producedBy string,
) (string, error) {
	auditEntries, err := keeper.GetAuditTrail(ctx, jobID.Uint64())
	if err != nil {
		// Audit trail is informational; build bundle without it rather than
		// failing — the on-chain anchors below are the canonical receipts.
		auditEntries = nil
	}

	bundle := ogstorage.ProofBundle{
		Version:         "1.0",
		JobID:           jobID.Uint64(),
		SpecHash:        "0x" + hex.EncodeToString(job.SpecHash[:]),
		ClientAddress:   job.Client.Hex(),
		ExecutorAddress: job.Executor.Hex(),
		Claim: ogstorage.BundleClaim{
			TxHash:              "0x" + hex.EncodeToString(job.TxHash[:]),
			ReportedOutcomeHash: "0x" + hex.EncodeToString(job.ReportedOutcomeHash[:]),
		},
		Settlement: ogstorage.BundleSettlement{
			SettleTxHash: settleTxHashHex,
			FinalVerdict: verdictWord(finalVerdict),
			BlockNumber:  settleBlock,
		},
		KeeperAudit: keeperEntriesToBundle(auditEntries),
		ProducedBy:  producedBy,
		ProducedAt:  time.Now().UnixMilli(),
	}

	uri, err := storage.UploadProofBundle(ctx, bundle)
	if err != nil {
		return "", fmt.Errorf("upload proof bundle: %w", err)
	}
	return uri, nil
}

func keeperEntriesToBundle(in []keeperhub.AuditEntry) []ogstorage.BundleKeeperEntry {
	out := make([]ogstorage.BundleKeeperEntry, 0, len(in))
	for _, e := range in {
		latency := e.Landed - e.Submitted
		if latency < 0 {
			latency = 0
		}
		out = append(out, ogstorage.BundleKeeperEntry{
			Purpose:   string(e.Purpose),
			TxHash:    e.TxHash.Hex(),
			Attempts:  e.Attempts,
			LatencyMs: latency,
			Note:      e.Note,
		})
	}
	return out
}

func verdictWord(v bool) string {
	if v {
		return "PASS"
	}
	return "FAIL"
}
