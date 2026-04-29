// executor is a simple CLI that simulates an off-chain executor. For a given job, it
// a synthetic Uniswap V3 swap receipt (no real swap happens), then
// calls AegisContract.submitClaim with the resulting tx hash.
// Production replaces BuildSyntheticSwap with a real Uniswap call
// routed through KeeperHub (step 7).
//
// Required env:
//
//	AEGIS_RPC             chain RPC for AegisContract
//	AEGIS_CONTRACT        AegisContract address
//	EXECUTOR_PRIVATE_KEY   hex private key (must match the executor address in the job)
//	JOB_ID                 the job id to claim against
//	CLIENT_ADDRESS         the swap recipient for the synthetic receipt
//
// Required flag:
//
//	--mode                 honest | high-slippage | tx-not-found
package main

import (
	"context"
	"flag"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/hpsing/aegis/internal/aegis"
	"github.com/hpsing/aegis/internal/executor"
)

func main() {
	modeFlag := flag.String("mode", "honest", "honest | high-slippage | tx-not-found")
	flag.Parse()

	mode, err := executor.ParseMode(*modeFlag)
	if err != nil {
		log.Fatalf("[executor] %v", err)
	}

	rpcURL := mustEnv("AEGIS_RPC")
	aegisAddr := common.HexToAddress(mustEnv("AEGIS_CONTRACT"))
	keyHex := strings.TrimPrefix(mustEnv("EXECUTOR_PRIVATE_KEY"), "0x")
	jobIDStr := mustEnv("JOB_ID")
	clientAddr := common.HexToAddress(mustEnv("CLIENT_ADDRESS"))

	key, err := ethcrypto.HexToECDSA(keyHex)
	if err != nil {
		log.Fatalf("[executor] bad EXECUTOR_PRIVATE_KEY: %v", err)
	}
	jobID, ok := new(big.Int).SetString(jobIDStr, 10)
	if !ok {
		log.Fatalf("[executor] bad JOB_ID: %s", jobIDStr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rpc, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		log.Fatalf("[executor] dial: %v", err)
	}
	defer rpc.Close()
	chainID, err := rpc.ChainID(ctx)
	if err != nil {
		log.Fatalf("[executor] chain id: %v", err)
	}

	q, err := aegis.NewAegisContract(aegisAddr, rpc)
	if err != nil {
		log.Fatalf("[executor] bind: %v", err)
	}

	// Hardcoded spec for local mode — TOOD:: match with publisher posts.
	usdcToken := common.HexToAddress("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913")
	wethToken := common.HexToAddress("0x4200000000000000000000000000000000000006")
	refOut, _ := new(big.Int).SetString("2848000000000000000", 10)
	swap := executor.BuildSyntheticSwap(mode, 8453, usdcToken, wethToken, 500, clientAddr, refOut)
	log.Printf("[executor] mode=%s tx=%s amount_out=%s", mode, swap.TxHash.Hex(), swap.ReportedAmountOut)

	// reportedOutcomeHash is just an opaque commitment; in production it'd
	// be the keccak of the canonical-JSON ReportedResult.
	reportedHash := ethcrypto.Keccak256Hash([]byte(swap.TxHash.Hex() + "|" + swap.ReportedAmountOut.String()))

	tx, err := executor.SubmitClaim(ctx, q, key, chainID, jobID, swap.TxHash, reportedHash)
	if err != nil {
		log.Fatalf("[executor] submitClaim: %v", err)
	}
	log.Printf("[executor] submitClaim tx=%s", tx.Hash().Hex())
}

func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("[executor] missing env %s", k)
	}
	return v
}
