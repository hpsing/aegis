package uniswapv3

import (
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

// USDC and WETH on Base mainnet (chain id 8453). These addresses are
// canonical and used as the test vector for ComputePoolAddress.
const (
	baseUSDC = "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"
	baseWETH = "0x4200000000000000000000000000000000000006"
)

func TestComputePoolAddress_OrderInvariant(t *testing.T) {
	usdc := common.HexToAddress(baseUSDC)
	weth := common.HexToAddress(baseWETH)

	a := ComputePoolAddress(BaseDeployment, usdc, weth, 500)
	b := ComputePoolAddress(BaseDeployment, weth, usdc, 500)
	require.Equal(t, a, b, "pool address must be order-invariant")
	require.NotEqual(t, common.Address{}, a, "non-zero")
}

func TestComputePoolAddress_FeeTierAffectsResult(t *testing.T) {
	usdc := common.HexToAddress(baseUSDC)
	weth := common.HexToAddress(baseWETH)
	low := ComputePoolAddress(BaseDeployment, usdc, weth, 500)
	mid := ComputePoolAddress(BaseDeployment, usdc, weth, 3000)
	require.NotEqual(t, low, mid, "different fee tier → different pool")
}

// TestKnownPool_USDC_WETH_500 anchors the address derivation to the real
// Base USDC/WETH 500-bps pool. If this test ever fails, the init code
// hash or factory address has drifted — re-anchor against
// https://basescan.org/address/0xd0b53d9277642d899df5c87a3966a349a798f224
// (the canonical pool) BEFORE shipping.
func TestKnownPool_USDC_WETH_500(t *testing.T) {
	usdc := common.HexToAddress(baseUSDC)
	weth := common.HexToAddress(baseWETH)
	got := ComputePoolAddress(BaseDeployment, usdc, weth, 500)
	want := common.HexToAddress("0xd0b53D9277642d899DF5C87A3966A349A798F224")
	require.True(t,
		strings.EqualFold(got.Hex(), want.Hex()),
		"USDC/WETH 500-bps pool must match basescan: got=%s want=%s", got.Hex(), want.Hex())
}

func TestDeploymentForChain(t *testing.T) {
	d, ok := DeploymentForChain(8453)
	require.True(t, ok)
	require.Equal(t, BaseDeployment, d)

	_, ok = DeploymentForChain(1)
	require.False(t, ok, "Ethereum mainnet not configured for hackathon")
}
