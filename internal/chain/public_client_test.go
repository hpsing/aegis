package chain

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// jsonrpcServer returns a tiny test server that always replies with the
// given body for every JSON-RPC request. Lets us exercise dial +
// chain id round-trip without hitting a real network.
func jsonrpcServer(body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
}

func TestNewPublicClient_RequiresPrimary(t *testing.T) {
	_, err := NewPublicClient(context.Background(), PublicClientConfig{})
	require.Error(t, err)
}

func TestPublicClient_ChainID_RoundTrip(t *testing.T) {
	// 0x2105 = 8453 (Base).
	srv := jsonrpcServer(`{"jsonrpc":"2.0","id":1,"result":"0x2105"}`)
	defer srv.Close()

	c, err := NewPublicClient(context.Background(), PublicClientConfig{
		PrimaryURL: srv.URL,
		RetryCount: 1,
	})
	require.NoError(t, err)
	defer c.Close()

	id, err := c.ChainID(context.Background())
	require.NoError(t, err)
	require.Equal(t, int64(8453), id.Int64())
}

func TestPublicClient_FallbackUsedOnPrimaryFailure(t *testing.T) {
	// Primary returns a JSON-RPC error; fallback returns a real chain id.
	primary := jsonrpcServer(`{"jsonrpc":"2.0","id":1,"error":{"code":-32603,"message":"boom"}}`)
	defer primary.Close()
	fallback := jsonrpcServer(`{"jsonrpc":"2.0","id":1,"result":"0x2105"}`)
	defer fallback.Close()

	c, err := NewPublicClient(context.Background(), PublicClientConfig{
		PrimaryURL:  primary.URL,
		FallbackURL: fallback.URL,
		RetryCount:  1,
	})
	require.NoError(t, err)
	defer c.Close()

	id, err := c.ChainID(context.Background())
	require.NoError(t, err)
	require.Equal(t, int64(8453), id.Int64())
}

func TestPublicClient_Close_NilSafe(t *testing.T) {
	c := &PublicClient{} // zero-value
	require.NotPanics(t, func() { c.Close() })
}
