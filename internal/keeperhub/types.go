// Package keeperhub is the verifier's reliable-execution rail for
// on-chain writes. Every commit, reveal, settle, and slash flows
// through Client.SendTransaction so we get a single audit trail per
// job — KeeperHub's prize-relevant value-prop.
//
// Two implementations satisfy Client:
//
//   - MockClient: signs + broadcasts directly via ethclient with the
//     verifier's own key, recording an in-memory audit trail. Used by
//     unit + integration tests.
//   - LiveClient (step 7 Phase D): forwards to the hosted MCP at
//     https://app.keeperhub.com/mcp via mark3labs/mcp-go.
//
// The interface is the same; cmd/verifier picks one based on env vars.
package keeperhub

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// Purpose tags every transaction so the audit trail is filterable per
// settlement step. Stable strings — they appear in ProofBundle JSON.
type Purpose string

const (
	PurposeRegister Purpose = "register"
	PurposeCommit   Purpose = "commit"
	PurposeReveal   Purpose = "reveal"
	PurposeSettle   Purpose = "settle"
	PurposeSlash    Purpose = "slash"
	PurposeOther    Purpose = "other"
)

// Metadata is propagated through KeeperHub so the audit trail can
// reconstruct WHY each tx existed. JobID is the AegisContract jobId.
type Metadata struct {
	JobID   uint64
	Purpose Purpose
	// IdempotencyKey is set by the caller (default: jobID-purpose). If
	// the same key shows up twice, the second call returns the first
	// call's receipt — protects against verifier crashes between send
	// and receipt-write.
	IdempotencyKey string
}

// SendTxInput is the high-level shape of a write the verifier wants
// landed on chain. Calldata is built off-chain (via abigen's NoSend
// transactor), then forwarded here so KeeperHub manages signing,
// nonce, gas, and retry.
type SendTxInput struct {
	ChainID  uint64
	To       common.Address
	Data     []byte
	Value    *big.Int // nil == 0
	GasLimit uint64   // 0 means "let KeeperHub estimate"
	Metadata Metadata
}

// Receipt is what every successful SendTransaction returns. Mirrors
// types.Receipt's headline fields plus KeeperHub-specific bookkeeping.
type Receipt struct {
	TxHash         common.Hash
	BlockNumber    uint64
	GasUsed        uint64
	Attempts       int   // 1 for happy path; >1 if KeeperHub retried
	TotalLatencyMs int64 // wall-clock from SendTransaction call to receipt
	// Signature from KeeperHub asserting it managed this tx (real client
	// only; MockClient leaves this empty).
	KeeperSignature string
}

// AuditEntry is one row of the audit trail returned by GetAuditTrail.
type AuditEntry struct {
	JobID     uint64
	Purpose   Purpose
	TxHash    common.Hash
	Status    string // "landed" | "failed" | "pending"
	Attempts  int
	Submitted int64 // Unix ms
	Landed    int64 // Unix ms; 0 if not yet landed
	Note      string
}
