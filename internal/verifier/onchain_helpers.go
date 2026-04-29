package verifier

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// signerAddrFrom is the canonical pubkey → 20-byte Ethereum address.
func signerAddrFrom(pub *ecdsa.PublicKey) common.Address {
	return ethcrypto.PubkeyToAddress(*pub)
}

// commitHash mirrors the contract-side commit hash recipe:
//
//	keccak256(abi.encode(verdict, nonce, msg.sender))
//
// Including msg.sender prevents a verifier from copying another's commit
// and front-running them.
func commitHash(verdict bool, nonce [32]byte, signer common.Address) [32]byte {
	// abi.encode adds 32-byte slots: bool → 32 bytes (0/1 right-padded zero),
	// bytes32 → as-is, address → left-padded to 32 bytes.
	buf := make([]byte, 0, 96)
	if verdict {
		buf = append(buf, common.LeftPadBytes([]byte{1}, 32)...)
	} else {
		buf = append(buf, make([]byte, 32)...)
	}
	buf = append(buf, nonce[:]...)
	buf = append(buf, common.LeftPadBytes(signer.Bytes(), 32)...)
	return ethcrypto.Keccak256Hash(buf)
}
