package claim

import (
	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/hpsing/aegis/internal/envelope"
)

// ID returns keccak256(canonical_json(spec) || "|" || executor.Bytes() || "|" || nonce[:]).
// Both sides of the pipe (Go agents and Solidity contract) must produce
// the same 32 bytes.
func ID(spec envelope.ClaimSpec, executor common.Address, nonce [16]byte) (common.Hash, error) {
	canonical, err := envelope.Canonicalize(spec)
	if err != nil {
		return common.Hash{}, err
	}
	h := ethcrypto.NewKeccakState()
	_, _ = h.Write(canonical)
	_, _ = h.Write([]byte("|"))
	_, _ = h.Write(executor.Bytes())
	_, _ = h.Write([]byte("|"))
	_, _ = h.Write(nonce[:])
	var out common.Hash
	_, _ = h.Read(out[:])
	return out, nil
}
