package verifier

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/hpsing/aegis/internal/chain"
	"github.com/hpsing/aegis/internal/envelope"
)

// CheckUSDCTransfer verifies that the executor (acting as a solver)
// actually transferred at least Spec.AmountIn of Spec.TokenIn to the
// client. Designed for the demo's "real solver" path: the spec
// describes a token + amount + chain, the executor signs an ERC-20
// Transfer on chain, and the verifier reads the tx receipt to confirm
// the Transfer event matches.
//
// This is the architecturally-honest version of step-5 mode — instead
// of the synthetic Uniswap receipt, the solver brings real working
// capital and forwards it to the client. The "swap" is just a token
// transfer, but the verification surface is the same shape: chain
// truth vs. claimed result.
func CheckUSDCTransfer(
	ctx context.Context,
	claim envelope.ExecutionClaim,
	rpc chain.EthClient,
) (Result, error) {
	spec := claim.Spec
	reported := claim.ReportedResult

	if claim.ClientAddress == "" || !common.IsHexAddress(claim.ClientAddress) {
		return fail(ReasonMissingClientAddress, Details{}), nil
	}
	clientAddr := common.HexToAddress(claim.ClientAddress)

	chainID, err := rpc.ChainID(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("chain id: %w", err)
	}
	if chainID.Int64() != spec.ChainID {
		return fail(ReasonWrongChain, Details{}), nil
	}

	if !common.IsHexAddress(spec.TokenIn) {
		return fail(ReasonWrongTokenPair, Details{}), nil
	}
	tokenAddr := common.HexToAddress(spec.TokenIn)

	expectedAmount, ok := new(big.Int).SetString(spec.AmountIn, 10)
	if !ok || expectedAmount.Sign() <= 0 {
		return fail(ReasonAmountInsufficient, Details{}), nil
	}

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
		return fail(ReasonTxReverted, Details{
			BlockNumber: receipt.BlockNumber.Uint64(),
			TxStatus:    "reverted",
		}), nil
	}

	// Walk logs for an ERC-20 Transfer to the client from the right
	// token contract. Multiple Transfers are tolerated — we accept any
	// matching event (e.g., the executor could batch sends).
	transferTopic := ethcrypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))
	var actualAmount *big.Int
	for _, lg := range receipt.Logs {
		if lg.Address != tokenAddr {
			continue
		}
		if len(lg.Topics) < 3 || lg.Topics[0] != transferTopic {
			continue
		}
		// Topics[2] is the indexed `to` address (right-padded to 32 bytes).
		toAddr := common.BytesToAddress(lg.Topics[2].Bytes())
		if toAddr != clientAddr {
			continue
		}
		amount := new(big.Int).SetBytes(lg.Data)
		if actualAmount == nil || amount.Cmp(actualAmount) > 0 {
			actualAmount = amount
		}
	}

	if actualAmount == nil {
		return fail(ReasonNoTransferEvent, Details{
			BlockNumber: receipt.BlockNumber.Uint64(),
			TxStatus:    "success",
		}), nil
	}
	if actualAmount.Cmp(expectedAmount) < 0 {
		return fail(ReasonAmountInsufficient, Details{
			ActualAmountOut: actualAmount,
			BlockNumber:     receipt.BlockNumber.Uint64(),
			TxStatus:        "success",
		}), nil
	}

	return pass(Details{
		ActualAmountOut: actualAmount,
		BlockNumber:     receipt.BlockNumber.Uint64(),
		TxStatus:        "success",
	}), nil
}
