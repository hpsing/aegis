package envelope

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// SigVersion is bumped whenever we change the signing scheme without
// changing the wire shape. Step 1 used ed25519 (SigVersion=1); step 2
// switched to secp256k1 / Ethereum-compatible signing (SigVersion=2).
// SigVersion is NOT a wire field — receivers infer the scheme from
// signer_pub's length and don't need to negotiate.
const SigVersion = 2

// Keypair is a secp256k1 identity. The same private key is later used to
// sign on-chain transactions, so generation goes through go-ethereum/crypto.
type Keypair struct {
	Priv *ecdsa.PrivateKey
}

func GenerateKeypair() (Keypair, error) {
	priv, err := ethcrypto.GenerateKey()
	if err != nil {
		return Keypair{}, fmt.Errorf("secp256k1 keygen: %w", err)
	}
	return Keypair{Priv: priv}, nil
}

// PublicHex returns the 64-byte uncompressed pubkey as 128 hex chars
// (no 0x prefix, no 0x04 leading byte). Matches the step-1 convention
// of "raw pubkey hex" — only the length differs (ed25519 was 64 chars).
func (k Keypair) PublicHex() string {
	pub := ethcrypto.FromECDSAPub(&k.Priv.PublicKey) // 65 bytes: 0x04 || X || Y
	return hex.EncodeToString(pub[1:])               // drop the 0x04 prefix
}

// Address derives the 20-byte Ethereum address from the public key.
// Matches keccak256(pubkey_64bytes)[12:] exactly.
func (k Keypair) Address() common.Address {
	return ethcrypto.PubkeyToAddress(k.Priv.PublicKey)
}

// signCanonical signs the canonical-JSON encoding of payload. Returns 65
// bytes (r || s || v) hex-encoded so receivers can ecrecover if needed.
func signCanonical(payload any, priv *ecdsa.PrivateKey) (string, error) {
	msg, err := Canonicalize(payload)
	if err != nil {
		return "", err
	}
	digest := ethcrypto.Keccak256(msg)
	sig, err := ethcrypto.Sign(digest, priv)
	if err != nil {
		return "", fmt.Errorf("secp256k1 sign: %w", err)
	}
	return hex.EncodeToString(sig), nil
}

// verifyCanonical verifies sigHex against the canonical-JSON of payload,
// expecting it to be signed by the keypair whose uncompressed pubkey
// hex matches pubHex.
func verifyCanonical(payload any, sigHex, pubHex string) (bool, error) {
	msg, err := Canonicalize(payload)
	if err != nil {
		return false, err
	}
	digest := ethcrypto.Keccak256(msg)

	sig, err := hex.DecodeString(strings.TrimPrefix(sigHex, "0x"))
	if err != nil {
		return false, fmt.Errorf("decode sig: %w", err)
	}

	pubBytes, err := hex.DecodeString(strings.TrimPrefix(pubHex, "0x"))
	if err != nil {
		return false, fmt.Errorf("decode pub: %w", err)
	}
	// VerifySignature wants the uncompressed (65-byte) form with the 0x04
	// prefix, and a 64-byte signature (no recovery id).
	if len(pubBytes) == 64 {
		pubBytes = append([]byte{0x04}, pubBytes...)
	}
	if len(sig) == 65 {
		sig = sig[:64]
	}
	return ethcrypto.VerifySignature(pubBytes, digest, sig), nil
}
