package verifier

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hpsing/aegis/internal/chain"
	"github.com/hpsing/aegis/internal/envelope"
	"github.com/hpsing/aegis/internal/verifier/uniswapv3"
)

// MinConfirmations is how many blocks past the receipt's block we
// require before voting. Reorgs on Base are rare past 3 confirmations.
const MinConfirmations = 3

// CheckUniswapSwap is the pure verification function. Given a signed
// ExecutionClaim and a chain client, it returns a Verdict the verifier
// can ship as a vote. All math is *big.Int — no float anywhere.
//
// Returns (Result, nil) for a definitive PASS or FAIL.
// Returns ({}, error) for transient network errors — caller should
// abstain (don't vote) rather than vote against the truth.
func CheckUniswapSwap(
	ctx context.Context,
	claim envelope.ExecutionClaim,
	rpc chain.EthClient,
) (Result, error) {
	spec := claim.Spec
	reported := claim.ReportedResult

	// The architecture-doc §6.4 rug-pull guard requires comparing the
	// swap's recipient to the client. If we don't know who the client
	// is, we can't check it — abstain by FAILing.
	if claim.ClientAddress == "" || !common.IsHexAddress(claim.ClientAddress) {
		return fail(ReasonMissingClientAddress, Details{}), nil
	}
	clientAddr := common.HexToAddress(claim.ClientAddress)

	// Confirm we're talking to the right chain.
	chainID, err := rpc.ChainID(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("chain id: %w", err)
	}
	if chainID.Int64() != spec.ChainID {
		return fail(
			ReasonWrongChain,
			Details{},
		), nil
	}

	// Pool resolution must come from a deployment we know about.
	deployment, ok := uniswapv3.DeploymentForChain(spec.ChainID)
	if !ok {
		return fail(ReasonUnknownDeployment, Details{}), nil
	}
	if !common.IsHexAddress(spec.TokenIn) || !common.IsHexAddress(spec.TokenOut) {
		return fail(ReasonWrongTokenPair, Details{}), nil
	}
	tokenIn := common.HexToAddress(spec.TokenIn)
	tokenOut := common.HexToAddress(spec.TokenOut)
	pool := uniswapv3.ComputePoolAddress(deployment, tokenIn, tokenOut, uint32(spec.FeeTier))

	// Receipt fetch.
	if !strings.HasPrefix(reported.TxHash, "0x") || len(reported.TxHash) != 66 {
		return fail(ReasonTxNotFound, Details{}), nil
	}
	txHash := common.HexToHash(reported.TxHash)
	receipt, err := rpc.TransactionReceipt(ctx, txHash)
	if err != nil {
		if errors.Is(err, chain.ErrTxNotFound) {
			return fail(ReasonTxNotFound, Details{}), nil
		}
		return Result{}, fmt.Errorf("receipt: %w", err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return fail(ReasonTxReverted, Details{TxStatus: "reverted"}), nil
	}

	// Reorg safety.
	head, err := rpc.BlockNumber(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("head: %w", err)
	}
	if receipt.BlockNumber.Uint64()+MinConfirmations > head {
		return fail(ReasonInsufficientConfirmations, Details{
			BlockNumber: receipt.BlockNumber.Uint64(),
			TxStatus:    "success",
		}), nil
	}

	// Find the Swap event from the matched pool. Multi-hop swaps emit
	// multiple Swap logs across multiple pools — we explicitly match
	// only the (tokenIn, tokenOut, feeTier) pool. If the swap routed
	// through a router, the matched pool's Swap log will likely have
	// recipient == router (not client) → rug-pull check fires below.
	swap, err := uniswapv3.FindSwapInReceipt(receipt, pool)
	if err != nil {
		return fail(ReasonNoSwapEvent, Details{
			BlockNumber: receipt.BlockNumber.Uint64(),
			TxStatus:    "success",
		}), nil
	}

	// Recipient must equal the client (architecture-doc §6.4 invariant).
	// SwapRouter02-routed swaps put the router as recipient → reject as
	// unsupported_route, NOT recipient_mismatch (cleaner narrative).
	if swap.Recipient != clientAddr {
		// Heuristic: if recipient is a known router (or just not the
		// client), flag as router routing. We don't ship a router list
		// for hackathon scope — any non-client recipient is treated as
		// recipient_mismatch (the safest interpretation).
		return fail(ReasonRecipientMismatch, Details{
			BlockNumber: receipt.BlockNumber.Uint64(),
			TxStatus:    "success",
		}), nil
	}

	// Pull the actualAmountOut from the matched Swap event. In V3 the
	// pool sends the OUT token to the recipient; that side appears as
	// a NEGATIVE amount0 or amount1 (signed convention, see decode.go).
	// We pick whichever side corresponds to tokenOut.
	token0, token1 := tokenIn, tokenOut
	if bytes.Compare(token0.Bytes(), token1.Bytes()) > 0 {
		token0, token1 = token1, token0
	}
	var amountOut *big.Int
	if tokenOut == token0 {
		// pool sent token0 → amount0 is negative
		amountOut = new(big.Int).Neg(swap.Amount0)
	} else {
		amountOut = new(big.Int).Neg(swap.Amount1)
	}
	if amountOut.Sign() < 0 {
		// Defensive: shouldn't happen if the swap actually swapped IN→OUT.
		return fail(ReasonWrongTokenPair, Details{
			BlockNumber: receipt.BlockNumber.Uint64(),
			TxStatus:    "success",
		}), nil
	}

	// Anti-fraud: the executor's `reported_result.amount_out` must match
	// what's actually on chain. An empty AmountOut means we don't have
	// the executor's reported value (step-5 local mode skips AXL/0G
	// fanout) — in that case we trust the chain. Production paths
	// (step 7+) always populate AmountOut via the envelope.
	if reported.AmountOut != "" {
		reportedOut, ok := new(big.Int).SetString(reported.AmountOut, 10)
		if !ok || reportedOut.Cmp(amountOut) != 0 {
			return fail(ReasonClaimAmountMismatchesChain, Details{
				ActualAmountOut: amountOut,
				BlockNumber:     receipt.BlockNumber.Uint64(),
				TxStatus:        "success",
			}), nil
		}
	}

	// Slippage math.
	refOut, ok := new(big.Int).SetString(spec.ReferenceQuoteAmountOut, 10)
	if !ok || refOut.Sign() <= 0 {
		return fail(ReasonWrongTokenPair, Details{
			ActualAmountOut: amountOut,
			BlockNumber:     receipt.BlockNumber.Uint64(),
		}), nil
	}
	slippageBps := slippage(refOut, amountOut)
	d := Details{
		ActualAmountOut:   amountOut,
		ActualSlippageBps: slippageBps,
		BlockNumber:       receipt.BlockNumber.Uint64(),
		TxStatus:          "success",
	}
	if slippageBps > uint64(spec.MaxSlippageBps) {
		return fail(ReasonSlippageExceeded, d), nil
	}
	return pass(d), nil
}

// slippage returns ceil((refOut - actualOut) * 10000 / refOut), clamped
// to 0 when actualOut >= refOut (positive slippage / good fill).
func slippage(refOut, actualOut *big.Int) uint64 {
	if actualOut.Cmp(refOut) >= 0 {
		return 0
	}
	diff := new(big.Int).Sub(refOut, actualOut)
	diff.Mul(diff, big.NewInt(10_000))
	diff.Div(diff, refOut)
	return diff.Uint64()
}
