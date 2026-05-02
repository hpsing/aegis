package uiserver

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// walletAuthMW gates state-changing endpoints behind an EIP-191 signature
// from an address on the AEGIS_WHITELIST allowlist. Unset or empty
// AEGIS_WHITELIST disables the gate (local dev / first-boot convenience).
//
// Headers required when the gate is active:
//   X-Aegis-Wallet: 0x-prefixed checksum address claiming to be the caller
//   X-Aegis-Sig:    EIP-191 signature over the message below, hex-encoded
//   X-Aegis-Ts:     unix seconds (string), within ±300s of server time
//
// Message signed (matches the SPA in web/src/wallet.ts):
//   "Aegis API access\nwallet: <wallet>\nts: <ts>"
//
// We only gate POSTs to /api/post-job — every other endpoint reads
// public chain state, and SSE EventSource can't send custom headers.
//
// Why EIP-191 (not SIWE/EIP-712): hackathon scope. EIP-191 needs zero
// extra deps in the browser (window.ethereum.personal_sign) and the
// recovered-signer check below is enough to prove wallet possession.
// If we widen the threat model (e.g. cross-origin replay), upgrade to
// SIWE with a server-issued nonce.
func walletAuthMW(next http.Handler) http.Handler {
	whitelist := loadWhitelist()
	gated := map[string]bool{
		"/api/post-job": true,
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !gated[r.URL.Path] || r.Method != http.MethodPost || len(whitelist) == 0 {
			next.ServeHTTP(w, r)
			return
		}
		if err := verifyWalletAuth(r, whitelist); err != nil {
			http.Error(w, "unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func verifyWalletAuth(r *http.Request, whitelist map[common.Address]bool) error {
	walletHdr := strings.TrimSpace(r.Header.Get("X-Aegis-Wallet"))
	sigHdr := strings.TrimSpace(r.Header.Get("X-Aegis-Sig"))
	tsHdr := strings.TrimSpace(r.Header.Get("X-Aegis-Ts"))
	if walletHdr == "" || sigHdr == "" || tsHdr == "" {
		return fmt.Errorf("missing X-Aegis-Wallet / X-Aegis-Sig / X-Aegis-Ts headers")
	}
	if !common.IsHexAddress(walletHdr) {
		return fmt.Errorf("X-Aegis-Wallet is not a valid address")
	}
	wallet := common.HexToAddress(walletHdr)
	if !whitelist[wallet] {
		return fmt.Errorf("wallet %s is not on the allowlist", wallet.Hex())
	}

	// Anti-replay: timestamp must be within ±5 minutes of server clock.
	var ts int64
	if _, err := fmt.Sscanf(tsHdr, "%d", &ts); err != nil {
		return fmt.Errorf("X-Aegis-Ts is not an integer")
	}
	now := time.Now().Unix()
	if ts < now-300 || ts > now+300 {
		return fmt.Errorf("X-Aegis-Ts %d is outside ±300s window (server now=%d)", ts, now)
	}

	sig, err := hex.DecodeString(strings.TrimPrefix(sigHdr, "0x"))
	if err != nil {
		return fmt.Errorf("X-Aegis-Sig hex decode: %w", err)
	}
	if len(sig) != 65 {
		return fmt.Errorf("X-Aegis-Sig must be 65 bytes (got %d)", len(sig))
	}
	// Ethereum signs with v in {27,28}; go-ethereum's Ecrecover wants {0,1}.
	if sig[64] >= 27 {
		sig[64] -= 27
	}

	msg := fmt.Sprintf("Aegis API access\nwallet: %s\nts: %d", strings.ToLower(wallet.Hex()), ts)
	prefixed := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(msg), msg)
	digest := ethcrypto.Keccak256([]byte(prefixed))

	pubkey, err := ethcrypto.SigToPub(digest, sig)
	if err != nil {
		return fmt.Errorf("recover signer: %w", err)
	}
	recovered := ethcrypto.PubkeyToAddress(*pubkey)
	if recovered != wallet {
		return fmt.Errorf("recovered signer %s != claimed wallet %s", recovered.Hex(), wallet.Hex())
	}
	return nil
}

// loadWhitelist reads AEGIS_WHITELIST once at construction. Empty /
// unset returns an empty map, which the middleware treats as "gate
// disabled" (so local dev with no env keeps working).
var (
	whitelistOnce sync.Once
	whitelistVal  map[common.Address]bool
)

func loadWhitelist() map[common.Address]bool {
	whitelistOnce.Do(func() {
		raw := strings.TrimSpace(os.Getenv("AEGIS_WHITELIST"))
		whitelistVal = map[common.Address]bool{}
		if raw == "" {
			return
		}
		for _, s := range strings.Split(raw, ",") {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			if !common.IsHexAddress(s) {
				fmt.Fprintf(os.Stderr, "AEGIS_WHITELIST: skipping invalid address %q\n", s)
				continue
			}
			whitelistVal[common.HexToAddress(s)] = true
		}
		if len(whitelistVal) == 0 {
			fmt.Fprintln(os.Stderr, "AEGIS_WHITELIST set but parsed 0 valid addresses — gate disabled")
		} else {
			fmt.Fprintf(os.Stderr, "AEGIS_WHITELIST: gating /api/post-job to %d wallet(s)\n", len(whitelistVal))
		}
	})
	return whitelistVal
}
