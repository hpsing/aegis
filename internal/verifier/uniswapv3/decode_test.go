package uniswapv3

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func TestDecodeSwap_RoundTrip(t *testing.T) {
	pool := common.HexToAddress("0xd0b53D9277642d899DF5C87A3966A349A798F224")
	sender := common.HexToAddress("0xdead000000000000000000000000000000000001")
	recipient := common.HexToAddress("0xc11e000000000000000000000000000000000002")
	amount0 := big.NewInt(10_000_000_000) // 10K USDC paid in
	amount1 := big.NewInt(-2_846_000_000_000_000_000) // ~2.846 ETH received (negative = pool sent)
	sqrt := new(big.Int).SetUint64(1_400_000_000_000_000_000)
	liq := new(big.Int).SetUint64(123_456_789)
	tick := big.NewInt(-200_000)

	log, err := BuildSwapLog(pool, sender, recipient, amount0, amount1, sqrt, liq, tick)
	require.NoError(t, err)
	got, err := DecodeSwap(log)
	require.NoError(t, err)
	require.Equal(t, pool, got.PoolAddress)
	require.Equal(t, sender, got.Sender)
	require.Equal(t, recipient, got.Recipient)
	require.Equal(t, 0, got.Amount0.Cmp(amount0))
	require.Equal(t, 0, got.Amount1.Cmp(amount1))
	require.Equal(t, 0, got.Tick.Cmp(tick))
}

func TestDecodeSwap_NotASwapLog(t *testing.T) {
	log := &types.Log{
		Address: common.Address{},
		Topics:  []common.Hash{common.HexToHash("0xdeadbeef")},
	}
	_, err := DecodeSwap(log)
	require.Error(t, err)
}

func TestFindSwapInReceipt_PicksMatchingPool(t *testing.T) {
	pool := common.HexToAddress("0xd0b53D9277642d899DF5C87A3966A349A798F224")
	otherPool := common.HexToAddress("0xfeedface00000000000000000000000000000000")
	recipient := common.HexToAddress("0xc11e000000000000000000000000000000000002")
	amount0 := big.NewInt(1)
	amount1 := big.NewInt(-1)
	sqrt := big.NewInt(0)
	liq := big.NewInt(0)
	tick := big.NewInt(0)

	// Swap log on otherPool first; a Swap log on the matched pool second.
	r := &types.Receipt{
		Logs: []*types.Log{
			MustBuildSwapLog(otherPool, recipient, recipient, amount0, amount1, sqrt, liq, tick),
			MustBuildSwapLog(pool, recipient, recipient, amount0, amount1, sqrt, liq, tick),
		},
	}
	got, err := FindSwapInReceipt(r, pool)
	require.NoError(t, err)
	require.Equal(t, pool, got.PoolAddress)
}

func TestFindSwapInReceipt_NoMatch(t *testing.T) {
	pool := common.HexToAddress("0xd0b53D9277642d899DF5C87A3966A349A798F224")
	r := &types.Receipt{Logs: []*types.Log{}}
	_, err := FindSwapInReceipt(r, pool)
	require.Error(t, err)
}
