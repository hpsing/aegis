package uniswapv3

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// Per-chain Uniswap V3 deployment constants. We hardcode the values for
// chains we verify against. Adding a new chain = appending one entry.
type Deployment struct {
	Factory      common.Address
	InitCodeHash common.Hash
}

// V3 deployments. Initcode hash is from the canonical Uniswap V3 deploy:
// keccak256(UniswapV3Pool init bytecode).
var (
	BaseDeployment = Deployment{
		Factory:      common.HexToAddress("0x33128a8fC17869897dcE68Ed026d694621f6FDfD"),
		InitCodeHash: hashOrPanic("e34f199b19b2b4f47f68442619d555527d244f78a3297ea89325f843f87b8b54"),
	}
)

// DeploymentForChain returns the canonical V3 deployment for the given
// EVM chain id, or false if we don't have one configured.
func DeploymentForChain(chainID int64) (Deployment, bool) {
	switch chainID {
	case 8453: // Base mainnet
		return BaseDeployment, true
	case 84_532: // Base Sepolia (same V3 deployment)
		return BaseDeployment, true
	default:
		return Deployment{}, false
	}
}

// ComputePoolAddress derives the deterministic V3 pool address for
// (tokenA, tokenB, fee) on the given deployment. The order of tokenA /
// tokenB is normalised internally — Uniswap stores tokens in ascending
// order (token0 < token1). Returns the same address regardless of input order.
func ComputePoolAddress(d Deployment, tokenA, tokenB common.Address, fee uint32) common.Address {
	token0, token1 := tokenA, tokenB
	if bytes.Compare(token0.Bytes(), token1.Bytes()) > 0 {
		token0, token1 = token1, token0
	}
	// salt = keccak256(abi.encode(token0, token1, fee))
	salt := ethcrypto.Keccak256(
		common.LeftPadBytes(token0.Bytes(), 32),
		common.LeftPadBytes(token1.Bytes(), 32),
		uint256BE(uint64(fee)),
	)
	// CREATE2 address = keccak256(0xff || factory || salt || initCodeHash)[12:]
	preimage := append([]byte{0xff}, d.Factory.Bytes()...)
	preimage = append(preimage, salt...)
	preimage = append(preimage, d.InitCodeHash.Bytes()...)
	hash := ethcrypto.Keccak256(preimage)
	return common.BytesToAddress(hash[12:])
}

// uint256BE returns the big-endian 32-byte encoding of v.
func uint256BE(v uint64) []byte {
	out := make([]byte, 32)
	for i := 0; i < 8; i++ {
		out[31-i] = byte(v >> (8 * i))
	}
	return out
}

func hashOrPanic(hexStr string) common.Hash {
	b, err := hex.DecodeString(hexStr)
	if err != nil || len(b) != 32 {
		panic(fmt.Sprintf("uniswapv3: bad init code hash %q: %v", hexStr, err))
	}
	var h common.Hash
	copy(h[:], b)
	return h
}
