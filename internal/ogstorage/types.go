package ogstorage

//   - AgentCard       — public verifier metadata, pointed to by iNFT.agentCardURI
//   - VoteRecord      — append-only entry on a per-verifier Log
//   - ProofBundle     — per-job long-form receipt (built by the verifier
//     that calls settle; written once per job)

const AgentCardVersion = "1.0" // for internal use.

// AgentCard is the public document describing a verifier. Hosted on
// 0G Storage; pointer lives on the verifier's ERC-7857 iNFT.
type AgentCard struct {
	Version                 string            `json:"version"`
	Name                    string            `json:"name"`
	Description             string            `json:"description"`
	Type                    string            `json:"type"` // "execution_verifier" today
	SupportedActions        []string          `json:"supported_actions"`
	Endpoints               map[string]string `json:"endpoints"`
	WalletAddress           string            `json:"wallet_address"`
	RegisteredAtBlock       uint64            `json:"registered_at_block"`
	StakeToken              string            `json:"stake_token"`
	MinStake                string            `json:"min_stake"`
	VerifierProtocolVersion int               `json:"verifier_protocol_version"`
}

// VoteRecord is one row of a verifier's append-only vote-history Log on
// 0G Storage. An external observer can reconstruct accuracy as
// `correct/total` by replaying these.
type VoteRecord struct {
	JobID        uint64 `json:"job_id"`
	VerdictVoted string `json:"verdict_voted"` // "PASS" | "FAIL"
	VerdictFinal string `json:"verdict_final"` // "PASS" | "FAIL"
	WasCorrect   bool   `json:"was_correct"`
	Timestamp    int64  `json:"timestamp"` // Unix ms
	SettleTxHash string `json:"settle_tx_hash"`
}

// ProofBundle is the long-form receipt for a settled job. Written ONCE
// per job by whichever verifier ends up calling settle (or as a
// post-settlement fanout). Anchors all the evidence judges might inspect.
type ProofBundle struct {
	Version           string               `json:"version"`
	JobID             uint64               `json:"job_id"`
	SpecHash          string               `json:"spec_hash"`
	SpecCanonicalJSON string               `json:"spec_canonical_json"`
	ClientAddress     string               `json:"client_address"`
	ExecutorAddress   string               `json:"executor_address"`
	Claim             BundleClaim          `json:"claim"`
	VerifierVotes     []BundleVerifierVote `json:"verifier_votes"`
	Settlement        BundleSettlement     `json:"settlement"`
	KeeperAudit       []BundleKeeperEntry  `json:"keeper_audit,omitempty"`
	ProducedBy        string               `json:"produced_by"`
	ProducedAt        int64                `json:"produced_at"`
	Signature         string               `json:"signature,omitempty"`
}

// BundleClaim captures the core claim details that define a job's "what happened".
type BundleClaim struct {
	TxHash              string `json:"tx_hash"`
	BlockNumber         uint64 `json:"block_number"`
	ReportedOutcomeHash string `json:"reported_outcome_hash"`
}

// BundleVerifierVote captures one vote by one verifier on a claim,
// as well as the final settled verdict and correctness.
// This is the core "who said what, and were they right?"
// data point for external observers to audit/verifier performance.
type BundleVerifierVote struct {
	VerifierAddress         string `json:"verifier_address"`
	VerifierINftID          uint64 `json:"verifier_inft_id"`
	Verdict                 string `json:"verdict"`
	RevealTxHash            string `json:"reveal_tx_hash"`
	WasCorrect              bool   `json:"was_correct"`
	ComputedSlippageBps     uint64 `json:"computed_slippage_bps,omitempty"`
	VerifierProtocolVersion int    `json:"verifier_protocol_version"`
}

// BundleSettlement captures the settlement details of a job:
// when it settled, what the final verdict was, and the tx that settled it.
type BundleSettlement struct {
	SettleTxHash string `json:"settle_tx_hash"`
	FinalVerdict string `json:"final_verdict"`
	BlockNumber  uint64 `json:"block_number"`
}

// BundleKeeperEntry captures one keeper action relevant to a job.
type BundleKeeperEntry struct {
	Purpose   string `json:"purpose"`
	TxHash    string `json:"tx_hash"`
	Attempts  int    `json:"attempts"`
	LatencyMs int64  `json:"latency_ms"`
	Note      string `json:"note,omitempty"`
}
