package executor

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/hpsing/aegis/internal/aegis"
	"github.com/hpsing/aegis/internal/verifier/uniswapv3"
)

// Mode controls the synthetic swap output. Mirrors step-05's
// `--mode={honest, high-slippage, tx-not-found}` from cmd/publisher.
type Mode string

const (
	ModeHonest       Mode = "honest"
	ModeHighSlippage Mode = "high-slippage"
	ModeTxNotFound   Mode = "tx-not-found"
)

// SyntheticSwap is the receipt + metadata the executor generates locally
// to feed verifiers. The receipt is added to the shared MockClient (test)
// or, in production, would be a REAL Base receipt the executor's tx
// produced.
type SyntheticSwap struct {
	TxHash            common.Hash
	BlockNumber       uint64
	Receipt           *types.Receipt // nil when Mode == ModeTxNotFound
	ReportedAmountOut *big.Int       // amount the executor will REPORT in submitClaim
}

// BuildSyntheticSwap returns a fixture-based receipt for local devnet
// testing. The receipt is suitable for chain.MockClient.Receipts.
//
//	tokenIn, tokenOut, feeTier — match the publisher's ClaimSpec
//	chainID — must match the verifier's swap-chain id (8453 = Base for local mode)
//	clientAddr — the address that should appear as the Swap log's recipient
//	referenceAmountOut — the spec's reference quote; honest mode delivers exactly this
func BuildSyntheticSwap(
	mode Mode,
	chainID int64,
	tokenIn, tokenOut common.Address,
	feeTier uint32,
	clientAddr common.Address,
	referenceAmountOut *big.Int,
) SyntheticSwap {
	tx := randomHash()
	if mode == ModeTxNotFound {
		return SyntheticSwap{TxHash: tx, ReportedAmountOut: referenceAmountOut}
	}

	deployment, ok := uniswapv3.DeploymentForChain(chainID)
	if !ok {
		panic(fmt.Sprintf("executor: unknown deployment for chain %d", chainID))
	}
	pool := uniswapv3.ComputePoolAddress(deployment, tokenIn, tokenOut, feeTier)

	// honest = exactly reference. high-slippage = ~50 bps below reference.
	actualOut := new(big.Int).Set(referenceAmountOut)
	if mode == ModeHighSlippage {
		actualOut.Mul(actualOut, big.NewInt(9_950))
		actualOut.Div(actualOut, big.NewInt(10_000))
	}

	// Determine token0/token1 ordering to set amount0/amount1 sign correctly.
	token0 := tokenIn
	if compareTokens(tokenIn, tokenOut) > 0 {
		token0 = tokenOut
	}
	var amount0, amount1 *big.Int
	if tokenOut == token0 {
		amount0 = new(big.Int).Neg(actualOut)
		amount1 = big.NewInt(10_000_000_000)
	} else {
		amount0 = big.NewInt(10_000_000_000)
		amount1 = new(big.Int).Neg(actualOut)
	}

	log := uniswapv3.MustBuildSwapLog(
		pool, randomAddress(), clientAddr,
		amount0, amount1, big.NewInt(0), big.NewInt(0), big.NewInt(0),
	)
	receipt := &types.Receipt{
		Status:      types.ReceiptStatusSuccessful,
		BlockNumber: big.NewInt(11_999_500),
		Logs:        []*types.Log{log},
		TxHash:      tx,
	}
	return SyntheticSwap{
		TxHash:            tx,
		BlockNumber:       receipt.BlockNumber.Uint64(),
		Receipt:           receipt,
		ReportedAmountOut: actualOut,
	}
}

// SubmitClaim signs and broadcasts AegisContract.submitClaim().
func SubmitClaim(
	ctx context.Context,
	rpcClient *aegis.AegisContract,
	signerKey *ecdsa.PrivateKey,
	chainID *big.Int,
	jobID *big.Int,
	txHash common.Hash,
	reportedOutcomeHash common.Hash,
) (*types.Transaction, error) {
	opts, err := bind.NewKeyedTransactorWithChainID(signerKey, chainID)
	if err != nil {
		return nil, fmt.Errorf("transactor: %w", err)
	}
	opts.Context = ctx
	tx, err := rpcClient.SubmitClaim(opts, jobID, txHash, reportedOutcomeHash)
	if err != nil {
		return nil, fmt.Errorf("submitClaim: %w", err)
	}
	return tx, nil
}

// ParseMode normalizes a CLI flag into a Mode enum.
func ParseMode(s string) (Mode, error) {
	switch Mode(s) {
	case ModeHonest, ModeHighSlippage, ModeTxNotFound:
		return Mode(s), nil
	default:
		return "", errors.New("executor: unknown mode " + s)
	}
}

// compareTokens replicates uniswapv3's internal ordering.
func compareTokens(a, b common.Address) int {
	for i := 0; i < 20; i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	return 0
}

func randomHash() common.Hash {
	var h common.Hash
	_, _ = rand.Read(h[:])
	return h
}

func randomAddress() common.Address {
	var a common.Address
	_, _ = rand.Read(a[:])
	return a
}
