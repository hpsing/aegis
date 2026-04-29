package ogstorage

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// Default 0G Galileo testnet endpoints. Override via env.
const (
	DefaultGalileoRPC      = "https://evmrpc-testnet.0g.ai"
	DefaultIndexerTurboURL = "https://indexer-storage-testnet-turbo.0g.ai"
)

// FromEnv returns a Storage backed by the real 0G network when
// OG_PRIVATE_KEY is set, else a MockStorage suitable for tests + CI.
//
// Honors:
//
//	OG_PRIVATE_KEY    hex private key for the upload-paying wallet (REQUIRED for live)
//	OG_GALILEO_RPC    EVM RPC URL (default: https://evmrpc-testnet.0g.ai)
//	OG_INDEXER_URL    storage indexer URL (default: turbo testnet indexer)
//
// The returned io.Closer must be closed to release the underlying web3
// client. For MockStorage Close is a no-op.
func FromEnv() (Storage, io.Closer, error) {
	pk := strings.TrimPrefix(os.Getenv("OG_PRIVATE_KEY"), "0x")
	if pk == "" {
		return nil, noopCloser{}, fmt.Errorf("ogstorage: OG_PRIVATE_KEY not set; cannot use live storage")
	}
	rpc := envOr("OG_GALILEO_RPC", DefaultGalileoRPC)
	idx := envOr("OG_INDEXER_URL", DefaultIndexerTurboURL)
	live, err := NewLiveStorage(rpc, idx, pk)
	if err != nil {
		return nil, nil, fmt.Errorf("ogstorage: live setup: %w", err)
	}
	return live, closerFunc(live.Close), nil
}

// IsLive reports whether FromEnv would return a Live (rather than Mock)
// storage given the current env. Useful for cmd-startup logging.
func IsLive() bool {
	return os.Getenv("OG_PRIVATE_KEY") != ""
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

type noopCloser struct{}

func (noopCloser) Close() error { return nil }

type closerFunc func()

func (f closerFunc) Close() error { f(); return nil }
