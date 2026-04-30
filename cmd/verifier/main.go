// verifier — the step-5 verifier role. Subscribes to QuorumContract
// events, runs CheckClaim against a swap-chain RPC, commits + reveals
// votes. The actual loop lives in internal/verifier — main() just wires
// env vars to LoopConfig.
//
// Required env:
//
//	AEGIS_RPC                WS endpoint of the chain hosting AegisContract
//	AEGIS_CONTRACT           AegisContract address
//	REGISTRY_CONTRACT        VerifierRegistry address
//	VERIFIER_PRIVATE_KEY     hex private key (no 0x prefix); EOA must be registered
//	SWAP_RPC                 RPC for the chain where the swap happened
//
// Optional 0G Storage env (verifier writes VoteRecords on settle):
//
//	OG_PRIVATE_KEY           hex private key for the 0G upload-paying wallet.
//	                         If unset, falls back to in-memory MockStorage.
//	OG_GALILEO_RPC           default https://evmrpc-testnet.0g.ai
//	OG_INDEXER_URL           default https://indexer-storage-testnet-turbo.0g.ai
//	OG_VERIFIER_INFT_ID      tokenId of this verifier's iNFT (decimal)
//
// Optional KeeperHub env (Option C — write path through MCP):
//
//	KEEPERHUB_API_KEY        org API key. If set, all on-chain writes go
//	                         through KH workflows instead of the local
//	                         MockClient. Must match this verifier's org.
//	KEEPERHUB_VERIFIER_INDEX 1, 2, or 3 — selects [verifier_N] in the
//	                         KH config so we know which workflow ids
//	                         to call.
//	KEEPERHUB_CONFIG         path to keeperhub.toml (default: configs/keeperhub.toml)
//	KEEPERHUB_MCP_URL        override (default: https://app.keeperhub.com/mcp)
//
// Optional flags:
//
//	--corrupt                inverts verdict before commit (adversarial demo)
//	--label                  log prefix (defaults to last 6 chars of address)
package main

import (
	"context"
	"crypto/ecdsa"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/hpsing/aegis/internal/chain"
	"github.com/hpsing/aegis/internal/keeperhub"
	"github.com/hpsing/aegis/internal/ogstorage"
	"github.com/hpsing/aegis/internal/verifier"
)

func main() {
	corrupt := flag.Bool("corrupt", false, "adversarial mode: invert verdict before commit")
	label := flag.String("label", "", "log prefix (defaults to addr suffix)")
	flag.Parse()

	aegisRPC := mustEnv("AEGIS_RPC")
	aegisAddr := common.HexToAddress(mustEnv("AEGIS_CONTRACT"))
	registryAddr := common.HexToAddress(mustEnv("REGISTRY_CONTRACT"))
	keyHex := strings.TrimPrefix(mustEnv("VERIFIER_PRIVATE_KEY"), "0x")
	swapRPC := mustEnv("SWAP_RPC")

	key, err := ethcrypto.HexToECDSA(keyHex)
	if err != nil {
		log.Fatalf("[verifier] bad VERIFIER_PRIVATE_KEY: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stop
		log.Println("[verifier] shutdown signal")
		cancel()
	}()

	// KeeperHub client: if the org API key + verifier index are present,
	// use LiveClient (KH signs and broadcasts via its server-side wallet
	// integration). Otherwise fall back to MockClient (local key signs,
	// in-process audit trail).
	keeper, keeperCloser, err := buildKeeperClient(ctx, aegisRPC, key)
	if err != nil {
		log.Fatalf("[verifier] keeperhub: %v", err)
	}
	defer keeperCloser()

	onchain, err := verifier.NewOnChain(ctx, aegisRPC, aegisAddr, registryAddr, key, keeper)
	if err != nil {
		log.Fatalf("[verifier] onchain: %v", err)
	}
	defer onchain.Close()

	swapClient, err := chain.NewPublicClient(ctx, chain.PublicClientConfig{
		PrimaryURL:  swapRPC,
		FallbackURL: os.Getenv("SWAP_RPC_FALLBACK"),
	})
	if err != nil {
		log.Fatalf("[verifier] swap rpc: %v", err)
	}
	defer swapClient.Close()

	logLabel := *label
	if logLabel == "" {
		addr := onchain.Address().Hex()
		logLabel = "verifier-" + addr[len(addr)-6:]
	}

	storage, storageCloser, err := ogstorage.FromEnv()
	if err != nil {
		log.Fatalf("[verifier] ogstorage: %v", err)
	}
	defer func() { _ = storageCloser.Close() }()
	if ogstorage.IsLive() {
		log.Printf("[verifier] 0G Storage: LIVE (uploads will burn $OG)")
	} else {
		log.Printf("[verifier] 0G Storage: mock (set OG_PRIVATE_KEY to enable live)")
	}

	var iNftID uint64
	if v := os.Getenv("OG_VERIFIER_INFT_ID"); v != "" {
		n, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			log.Fatalf("[verifier] bad OG_VERIFIER_INFT_ID: %v", err)
		}
		iNftID = n
	}

	loop := verifier.NewLoop(verifier.LoopConfig{
		OnChain:  onchain,
		SwapRPC:  swapClient,
		Storage:  storage,
		INftID:   iNftID,
		Corrupt:  *corrupt,
		LogLabel: logLabel,
	})

	if err := loop.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("[verifier] loop: %v", err)
	}
}

func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("[verifier] missing env %s", k)
	}
	return v
}

// buildKeeperClient picks LiveClient when the KH env vars are present,
// otherwise MockClient. Returns the client + a closer that the caller
// must defer.
func buildKeeperClient(
	ctx context.Context,
	quorumRPC string,
	key *ecdsa.PrivateKey,
) (keeperhub.Client, func(), error) {
	apiKey := os.Getenv("KEEPERHUB_API_KEY")
	idxStr := os.Getenv("KEEPERHUB_VERIFIER_INDEX")
	if apiKey == "" || idxStr == "" {
		log.Printf("[verifier] KEEPERHUB_API_KEY or KEEPERHUB_VERIFIER_INDEX not set")
	}

	idx, err := strconv.Atoi(idxStr)
	if err != nil || idx < 1 || idx > 3 {
		return nil, nil, fmt.Errorf("KEEPERHUB_VERIFIER_INDEX must be 1, 2, or 3 (got %q)", idxStr)
	}

	cfgPath := os.Getenv("KEEPERHUB_CONFIG")
	if cfgPath == "" {
		cfgPath = "configs/keeperhub.toml"
	}
	cfg, err := keeperhub.LoadConfig(cfgPath)
	if err != nil {
		return nil, nil, fmt.Errorf("load %s: %w", cfgPath, err)
	}
	vcfg, err := cfg.VerifierByIndex(idx)
	if err != nil {
		return nil, nil, err
	}

	mcpURL := os.Getenv("KEEPERHUB_MCP_URL")
	if mcpURL == "" {
		mcpURL = cfg.MCPEndpoint
	}

	live, err := keeperhub.NewLiveClient(ctx, keeperhub.LiveClientConfig{
		Endpoint:    mcpURL,
		APIKey:      apiKey,
		WorkflowIDs: vcfg.WorkflowIDs(),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("live client: %w", err)
	}
	log.Printf("[verifier] keeperhub: LIVE (org wallet=%s, idx=%d, endpoint=%s)",
		vcfg.Address, idx, live.Endpoint())
	return live, live.Close, nil
}
