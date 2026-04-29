package uniswapv3

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// BuildSwapLog constructs a properly-formatted *types.Log for a V3 Swap
// event. Used by tests AND the local-devnet executor (which generates
// synthetic receipts to feed verifiers without doing a real Uniswap call).
//
// The function is in a non-test file so external packages can import it.
// Returns an error rather than fataling — callers in test code should
// `require.NoError`, callers in cmd code can panic.
func BuildSwapLog(
	pool, sender, recipient common.Address,
	amount0, amount1, sqrtPriceX96, liquidity *big.Int,
	tick *big.Int,
) (*types.Log, error) {
	dataArgs := SwapEventABI.Events["Swap"].Inputs[2:]
	data, err := dataArgs.Pack(amount0, amount1, sqrtPriceX96, liquidity, tick)
	if err != nil {
		return nil, fmt.Errorf("pack swap data: %w", err)
	}
	return &types.Log{
		Address: pool,
		Topics: []common.Hash{
			SwapEventID,
			common.BytesToHash(common.LeftPadBytes(sender.Bytes(), 32)),
			common.BytesToHash(common.LeftPadBytes(recipient.Bytes(), 32)),
		},
		Data: data,
	}, nil
}

// MustBuildSwapLog panics on error — convenient at test/fixture call
// sites where failure is impossible in practice.
func MustBuildSwapLog(
	pool, sender, recipient common.Address,
	amount0, amount1, sqrtPriceX96, liquidity *big.Int,
	tick *big.Int,
) *types.Log {
	log, err := BuildSwapLog(pool, sender, recipient, amount0, amount1, sqrtPriceX96, liquidity, tick)
	if err != nil {
		panic(err)
	}
	return log
}
