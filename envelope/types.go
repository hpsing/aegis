package envelope

// ClaimSpec is the workload the executor claims to have performed.
type ClaimSpec struct {
	Action                  string `json:"action"` // "uniswap_v3_swap" only in step 1-4
	ChainID                 int64  `json:"chain_id"`
	TokenIn                 string `json:"token_in"`  // EIP-55 address
	TokenOut                string `json:"token_out"` // EIP-55 address
	AmountIn                string `json:"amount_in"` // decimal string
	MaxSlippageBps          int    `json:"max_slippage_bps"`
	ReferenceQuoteBlock     int64  `json:"reference_quote_block"`
	ReferenceQuoteAmountOut string `json:"reference_quote_amount_out"` // decimal string
	FeeTier                 int    `json:"fee_tier"`
}

// ReportedResult is the executor's self-reported outcome. Steps 1-3
// trust it; step 4 swaps in real RPC verification.
type ReportedResult struct {
	TxHash      string `json:"tx_hash"`
	ChainID     int64  `json:"chain_id"`
	BlockNumber int64  `json:"block_number"`
	AmountOut   string `json:"amount_out"` // decimal string
}

// ExecutionClaim is the payload of an execution_claim envelope.
//
// ClientAddress was added in step 4: per docs/architecture.md §6.4, the
// verifier checks the swap's `recipient` matches the client's address
// (the rug-pull guard). Empty in step-1/2 envelopes — verifier
// abstains in that case. In step 5, the verifier cross-checks this
// field against the on-chain `Job.client` from QuorumContract.
type ExecutionClaim struct {
	Spec           ClaimSpec      `json:"spec"`
	ReportedResult ReportedResult `json:"reported_result"`
	ExecutorPub    string         `json:"executor_pub"`
	ClientAddress  string         `json:"client_address,omitempty"`
	ClaimID        string         `json:"claim_id"`
}

// VerifierVote is the payload of a verifier_vote envelope.
type VerifierVote struct {
	ClaimID     string `json:"claim_id"`
	Verdict     string `json:"verdict"` // "PASS" | "FAIL"
	Reason      string `json:"reason"`  // machine-parseable code
	VerifierPub string `json:"verifier_pub"`
}
