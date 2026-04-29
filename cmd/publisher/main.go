// publisher the client agent. Approves USDC, calls
// AegisContract.postJob with the canonical 3-line-item escrow.
// Prints the resulting jobId.
//
// Required env:
//
//	AEGIS_RPC          chain RPC (HTTP is fine; no subscriptions needed)
//	AEGIS_CONTRACT     AegisContract address
//	USDC_CONTRACT       MockUSDC (or real USDC) address
//	CLIENT_PRIVATE_KEY  hex private key
//	EXECUTOR_ADDRESS    EOA the client picks for this job
//
// Optional flags (USDC scaled to 6 decimals):
//
//	--reimbursement   default 100000000 (100 USDC)
//	--fee             default 10000000 (10 USDC)
//	--bounty          default 30000000 (30 USDC)
//	--spec-hash       hex32 (default keccak256("local-spec"))
package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/hpsing/aegis/internal/aegis"
	clientpkg "github.com/hpsing/aegis/internal/client"
)

func main() {
	reimb := flag.Int64("reimbursement", 100_000_000, "executor reimbursement in USDC (6 decimals)")
	fee := flag.Int64("fee", 10_000_000, "executor fee")
	bounty := flag.Int64("bounty", 30_000_000, "verifier bounty")
	specHashHex := flag.String("spec-hash", "", "32-byte hex spec hash; default keccak256(\"local-spec\")")
	flag.Parse()

	rpcURL := mustEnv("AEGIS_RPC")
	aegisAddr := common.HexToAddress(mustEnv("AEGIS_CONTRACT"))
	usdcAddr := common.HexToAddress(mustEnv("USDC_CONTRACT"))
	executorAddr := common.HexToAddress(mustEnv("EXECUTOR_ADDRESS"))
	keyHex := strings.TrimPrefix(mustEnv("CLIENT_PRIVATE_KEY"), "0x")

	key, err := ethcrypto.HexToECDSA(keyHex)
	if err != nil {
		log.Fatalf("[publisher] bad CLIENT_PRIVATE_KEY: %v", err)
	}

	var specHash common.Hash
	if *specHashHex == "" {
		specHash = ethcrypto.Keccak256Hash([]byte("local-spec"))
	} else {
		raw, err := hex.DecodeString(strings.TrimPrefix(*specHashHex, "0x"))
		if err != nil || len(raw) != 32 {
			log.Fatalf("[publisher] --spec-hash must be 32-byte hex")
		}
		copy(specHash[:], raw)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rpc, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		log.Fatalf("[publisher] dial: %v", err)
	}
	defer rpc.Close()
	chainID, err := rpc.ChainID(ctx)
	if err != nil {
		log.Fatalf("[publisher] chain id: %v", err)
	}

	usdc, err := aegis.NewMockUSDC(usdcAddr, rpc)
	if err != nil {
		log.Fatalf("[publisher] bind USDC: %v", err)
	}
	q, err := aegis.NewAegisContract(aegisAddr, rpc)
	if err != nil {
		log.Fatalf("[publisher] bind Aegis: %v", err)
	}

	tx, err := clientpkg.PostJob(
		ctx, usdc, q, key, chainID, aegisAddr, executorAddr, specHash,
		big.NewInt(*reimb), big.NewInt(*fee), big.NewInt(*bounty),
	)
	if err != nil {
		log.Fatalf("[publisher] postJob: %v", err)
	}

	log.Printf("[publisher] postJob tx=%s — wait for it to mine, then read the JobPosted event for jobId", tx.Hash().Hex())
	fmt.Println(tx.Hash().Hex())
}

func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("[publisher] missing env %s", k)
	}
	return v
}
