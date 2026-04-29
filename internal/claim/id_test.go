package claim

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/hpsing/aegis/internal/envelope"
)

func TestID_DeterministicForSameInput(t *testing.T) {
	spec := envelope.ClaimSpec{
		Action:                  "uniswap_v3_swap",
		ChainID:                 8453,
		TokenIn:                 "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
		TokenOut:                "0x4200000000000000000000000000000000000006",
		AmountIn:                "10000000000",
		MaxSlippageBps:          30,
		ReferenceQuoteBlock:     11999000,
		ReferenceQuoteAmountOut: "2848000000000000000",
		FeeTier:                 500,
	}
	executor := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
	var nonce [16]byte
	copy(nonce[:], []byte("0123456789abcdef"))

	a, err := ID(spec, executor, nonce)
	require.NoError(t, err)
	b, err := ID(spec, executor, nonce)
	require.NoError(t, err)
	require.Equal(t, a, b, "ID must be deterministic")
	require.NotEqual(t, common.Hash{}, a, "ID must not be zero")
}

func TestID_NonceChangesResult(t *testing.T) {
	spec := envelope.ClaimSpec{Action: "uniswap_v3_swap", ChainID: 8453}
	exec := common.HexToAddress("0xdeadbeef00000000000000000000000000000000")
	var n1, n2 [16]byte
	n1[0] = 1
	n2[0] = 2
	a, err := ID(spec, exec, n1)
	require.NoError(t, err)
	b, err := ID(spec, exec, n2)
	require.NoError(t, err)
	require.NotEqual(t, a, b)
}

func TestID_ExecutorChangesResult(t *testing.T) {
	spec := envelope.ClaimSpec{Action: "uniswap_v3_swap", ChainID: 8453}
	var nonce [16]byte
	a, err := ID(spec, common.HexToAddress("0x1"), nonce)
	require.NoError(t, err)
	b, err := ID(spec, common.HexToAddress("0x2"), nonce)
	require.NoError(t, err)
	require.NotEqual(t, a, b)
}

// TestID_KnownVector locks the digest format. If this test ever fails,
// the on-chain Solidity side will produce a different hash than Go and
// step 3's CanonicalDigest.t.sol will break too. Re-anchor BOTH in
// lockstep, never just one.
func TestID_KnownVector(t *testing.T) {
	spec := envelope.ClaimSpec{
		Action:                  "uniswap_v3_swap",
		ChainID:                 8453,
		TokenIn:                 "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
		TokenOut:                "0x4200000000000000000000000000000000000006",
		AmountIn:                "10000000000",
		MaxSlippageBps:          30,
		ReferenceQuoteBlock:     11999000,
		ReferenceQuoteAmountOut: "2848000000000000000",
		FeeTier:                 500,
	}
	executor := common.HexToAddress("0x000000000000000000000000000000000000beef")
	var nonce [16]byte // all zeros

	got, err := ID(spec, executor, nonce)
	require.NoError(t, err)
	require.Equal(t,
		"0x6679ed880ab2a9070fe0f2f29187daaba4610b818ecfc71bddae2ecf9f911b9b",
		got.Hex(),
		"if this fails, re-anchor BOTH this test AND the Solidity-side digest test in step 3")
}
