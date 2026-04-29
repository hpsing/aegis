package envelope

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestVerify_StaleEnvelope(t *testing.T) {
	kp, err := GenerateKeypair()
	require.NoError(t, err)
	env, err := Sign(TypeExecutionClaim, map[string]any{"x": 1}, kp)
	require.NoError(t, err)

	// Pretend the envelope was signed 2x freshness ago. We re-sign with
	// the old timestamp so the signature itself is valid; only freshness
	// fails.
	env.Ts = time.Now().Add(-2 * FreshnessWindow).UnixMilli()
	sig, err := signCanonical(envelopeToMap(env), kp.Priv)
	require.NoError(t, err)
	env.Sig = sig

	res := Verify(env)
	require.False(t, res.OK)
	require.Equal(t, "stale_or_future", res.Reason)
}

func TestVerify_FutureEnvelope(t *testing.T) {
	kp, err := GenerateKeypair()
	require.NoError(t, err)
	env, err := Sign(TypeExecutionClaim, map[string]any{"x": 1}, kp)
	require.NoError(t, err)

	env.Ts = time.Now().Add(2 * FreshnessWindow).UnixMilli()
	sig, err := signCanonical(envelopeToMap(env), kp.Priv)
	require.NoError(t, err)
	env.Sig = sig

	res := Verify(env)
	require.False(t, res.OK)
	require.Equal(t, "stale_or_future", res.Reason)
}

func TestVerify_WrongVersion(t *testing.T) {
	kp, err := GenerateKeypair()
	require.NoError(t, err)
	env, err := Sign(TypeExecutionClaim, map[string]any{"x": 1}, kp)
	require.NoError(t, err)
	env.V = 999

	res := Verify(env)
	require.False(t, res.OK)
	require.Equal(t, "unsupported_version", res.Reason)
}

func TestVerify_AcceptsBoundaryFreshness(t *testing.T) {
	kp, err := GenerateKeypair()
	require.NoError(t, err)
	env, err := Sign(TypeExecutionClaim, map[string]any{"x": 1}, kp)
	require.NoError(t, err)

	// Sign with a ts that's just under the freshness window.
	env.Ts = time.Now().Add(-FreshnessWindow + time.Second).UnixMilli()
	sig, err := signCanonical(envelopeToMap(env), kp.Priv)
	require.NoError(t, err)
	env.Sig = sig
	require.True(t, Verify(env).OK)
}
