package verifier

import "math/big"

// Verdict is the binary result of CheckClaim. The empty string sentinel
// "" is the abstain case — used when we hit a transient network error
// and don't want to vote (better to lose the bounty than be slashed).
type Verdict string

const (
	Pass    Verdict = "PASS"
	Fail    Verdict = "FAIL"
	Abstain Verdict = ""
)

const (
	ReasonOK                         = "ok"
	ReasonTxNotFound                 = "tx_not_found"
	ReasonTxReverted                 = "tx_reverted"
	ReasonWrongChain                 = "wrong_chain"
	ReasonNoSwapEvent                = "no_swap_event"
	ReasonWrongTokenPair             = "wrong_token_pair"
	ReasonRecipientMismatch          = "recipient_mismatch"
	ReasonSlippageExceeded           = "slippage_exceeded"
	ReasonClaimAmountMismatchesChain = "claim_amount_out_mismatches_chain"
	ReasonUnsupportedRoute           = "unsupported_route"
	ReasonUnsupportedAction          = "unsupported_action"
	ReasonInsufficientConfirmations  = "insufficient_confirmations"
	ReasonUnknownDeployment          = "unknown_deployment"
	ReasonMissingClientAddress       = "missing_client_address"
	ReasonNoTransferEvent            = "no_transfer_event"
	ReasonAmountInsufficient         = "amount_insufficient"
)

// Result is what CheckClaim hands back. Verdict is the headline; Reason
// is a machine-parseable string for logging; Details exposes the math.
type Result struct {
	Verdict Verdict
	Reason  string
	Details Details
}

type Details struct {
	ActualAmountOut   *big.Int
	ActualSlippageBps uint64
	BlockNumber       uint64
	TxStatus          string // "success" | "reverted" | ""
}

// pass / fail / abstain are tiny constructors so call sites read cleanly.
func pass(d Details) Result { return Result{Verdict: Pass, Reason: ReasonOK, Details: d} }
func fail(reason string, d Details) Result {
	return Result{Verdict: Fail, Reason: reason, Details: d}
}
