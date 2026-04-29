//go:build live

// Run with:
//
//	export OG_PRIVATE_KEY=<hex private key for funded wallet>
//	# optional overrides:
//	# export OG_GALILEO_RPC=https://evmrpc-testnet.0g.ai
//	# export OG_INDEXER_URL=https://indexer-storage-testnet-turbo.0g.ai
//	go test -tags=live -timeout=300s -run TestLive ./internal/ogstorage/
//
// Each run consumes a small amount of $OG (faucet at https://faucet.0g.ai).
// We deliberately do ONE upload to verify wiring; assertions are minimal.
//
// export OG_PRIVATE_KEY=xxxx
// go test -tags=live -timeout=300s -run TestLive_UploadAgentCard -v ./internal/ogstorage/
package ogstorage

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLive_UploadAgentCard(t *testing.T) {
	if os.Getenv("OG_PRIVATE_KEY") == "" {
		t.Skip("OG_PRIVATE_KEY not set; skipping live test")
	}

	storage, closer, err := FromEnv()
	require.NoError(t, err)
	defer closer.Close()
	require.True(t, IsLive(), "expected LiveStorage when OG_PRIVATE_KEY is set")

	card := AgentCard{
		Version:                 AgentCardVersion,
		Name:                    "Quorum Verifier (live smoke)",
		Description:             "Live-test upload from internal/ogstorage/live_test.go",
		Type:                    "execution_verifier",
		SupportedActions:        []string{"uniswap_v3_swap"},
		Endpoints:               map[string]string{"axl_peer_id": "smoke-test"},
		WalletAddress:           "0x7C9DcA2fB05cFc732794CEe846888f3B1D7FE06a",
		RegisteredAtBlock:       0,
		StakeToken:              "USDC",
		MinStake:                "100",
		VerifierProtocolVersion: 1,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()

	uri, err := storage.UploadAgentCard(ctx, card)
	require.NoError(t, err, "live upload to 0G Galileo failed")
	require.Contains(t, uri, "0g://", "URI should be the 0g:// scheme")
	require.Greater(t, len(uri), len("0g://0x")+10, "URI should carry a non-trivial root hash")

	t.Logf("LIVE UPLOAD OK — uri=%s", uri)
	t.Logf("inspect at: https://storagescan-galileo.0g.ai/")
}
