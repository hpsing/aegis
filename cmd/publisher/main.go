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
// Required AXL env (off-chain spec distribution):
//
//	AXL_NODE_URL          local AXL daemon, e.g. http://127.0.0.1:9002
//	AXL_VERIFIER_PEERS    comma-separated peer ids of the 3 verifiers
//
// After postJob lands the publisher serializes the canonical-JSON spec
// and sends it to each verifier peer via AXL. Verifiers fetch and
// validate keccak256(spec) == specHash before use.
//
// AXL is required — without it Quorum's swarm doesn't satisfy the
// "verifiers communicate over AXL" architecture invariant.
// Optional flags (USDC scaled to 6 decimals):
//
//	--reimbursement   default 100000000 (100 USDC)
//	--fee             default 10000000 (10 USDC)
//	--bounty          default 30000000 (30 USDC)
//	--spec-hash       hex32 (default keccak256(canonical(spec))
package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
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
	"github.com/hpsing/aegis/internal/axl"
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

	// Build a canonical-JSON spec describing the verification target.
	// AXL distribution sends this body to each verifier; on-chain we
	// store only its keccak256. When --spec-hash is overridden, we
	// honor that and skip AXL publish.
	var specHash common.Hash
	var specJSON []byte
	if *specHashHex == "" {
		specJSON = canonicalSpec(executorAddr)
		specHash = ethcrypto.Keccak256Hash(specJSON)
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

	// AXL spec distribution is required for prize-relevant flow.
	if specJSON == nil {
		log.Fatalf("[publisher] AXL is required: a canonical spec must be built (got nil; check --spec-hash override)")
	}
	axlURL := mustEnv("AXL_NODE_URL")
	peers := splitPeers(mustEnv("AXL_VERIFIER_PEERS"))
	if len(peers) == 0 {
		log.Fatalf("[publisher] AXL_VERIFIER_PEERS empty after parsing")
	}
	if err := publishSpecToAXL(ctx, axlURL, peers, specHash, specJSON); err != nil {
		log.Fatalf("[publisher] AXL publish: %v", err)
	}
	log.Printf("[publisher] AXL: published spec(%s) to %d verifiers", specHash.Hex()[:10]+"…", len(peers))
}

// canonicalSpec builds a deterministic JSON body for the demo. Real
// production would derive this from --intent / --token-in / etc flags
// and the canonical encoder in internal/envelope. For demo purposes
// the contents are stable across runs except for the executor.
func canonicalSpec(executor common.Address) []byte {
	// Field order matches our envelope canonicalizer (sorted keys).
	body := map[string]any{
		"amountIn":       "10000000000",
		"deadline":       1735000000,
		"executor":       executor.Hex(),
		"intent":         "swap",
		"maxSlippageBps": 30,
		"tokenIn":        "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
		"tokenOut":       "0x4200000000000000000000000000000000000006",
	}
	out, _ := json.Marshal(body)
	return out
}

func splitPeers(csv string) []string {
	parts := strings.Split(csv, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// publishSpecToAXL sends a TypeSpecPublish envelope to every verifier
// peer. We don't gather acks — fire and forget; verifiers' recv loops
// pick the spec up.
func publishSpecToAXL(
	ctx context.Context, nodeURL string, peerIDs []string, specHash common.Hash, specJSON []byte,
) error {
	client := axl.NewClient(nodeURL)
	payload := axl.SpecPayload{
		JobID:    0, // jobId not known at publish time (set by chain); receivers index by SpecHash
		SpecHash: specHash.Hex(),
		Spec:     specJSON,
	}
	envBytes, err := axl.MarshalEnvelope(axl.TypeSpecPublish, payload)
	if err != nil {
		return fmt.Errorf("marshal envelope: %w", err)
	}
	var firstErr error
	for _, p := range peerIDs {
		if err := client.Send(ctx, p, envBytes); err != nil {
			log.Printf("[publisher] AXL send to %s…: %v", p[:10], err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("[publisher] missing env %s", k)
	}
	return v
}
