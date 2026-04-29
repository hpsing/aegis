// verifier — the verifier role. Subscribes to AegisContract.
// events, runs CheckClaim against a swap-chain RPC, commits + reveals
// votes. The actual loop lives in internal/verifier — main() just wires
// env vars to LoopConfig.
//
// Required env:
//
//	AEGIS_RPC               WS endpoint of the chain hosting AegisContract
//	AEGIS_CONTRACT          AegisContract address
//	REGISTRY_CONTRACT        VerifierRegistry address
//	VERIFIER_PRIVATE_KEY     hex private key (no 0x prefix); EOA must be registered
//	SWAP_RPC                 RPC for the chain where the swap happened
//
// Optional flags:
//
//	--corrupt                inverts verdict before commit (adversarial demo)
//	--label                  log prefix (defaults to last 6 chars of address)
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/hpsing/aegis/internal/chain"
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

	onchain, err := verifier.NewOnChain(ctx, aegisRPC, aegisAddr, registryAddr, key)
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

	loop := verifier.NewLoop(verifier.LoopConfig{
		OnChain:  onchain,
		SwapRPC:  swapClient,
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
