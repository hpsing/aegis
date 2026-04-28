package envelope

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSign_RoundTrip(t *testing.T) {
	kp, err := GenerateKeypair()
	require.NoError(t, err)

	payload := map[string]any{"hello": "world", "n": 42}
	env, err := Sign(TypeExecutionClaim, payload, kp)
	require.NoError(t, err)
	require.Equal(t, Version, env.V)
	require.Equal(t, TypeExecutionClaim, env.Type)
	require.NotEmpty(t, env.Sig)
	require.NotEmpty(t, env.SignerPub)
	require.Equal(t, kp.PublicHex(), env.SignerPub)

	require.True(t, Verify(env).OK)
}

func TestVerify_TamperedPayload(t *testing.T) {
	kp, err := GenerateKeypair()
	require.NoError(t, err)
	env, err := Sign(TypeExecutionClaim, map[string]any{"x": 1}, kp)
	require.NoError(t, err)

	// Replace payload bytes — sig should no longer cover them.
	env.Payload = []byte(`{"x":2}`)
	res := Verify(env)
	require.False(t, res.OK)
	require.Equal(t, "bad_signature", res.Reason)
}

func TestVerify_TamperedTimestamp(t *testing.T) {
	kp, err := GenerateKeypair()
	require.NoError(t, err)
	env, err := Sign(TypeExecutionClaim, map[string]any{"x": 1}, kp)
	require.NoError(t, err)

	env.Ts += 1 // 1ms shift, still inside freshness window but invalidates sig
	res := Verify(env)
	require.False(t, res.OK)
	require.Equal(t, "bad_signature", res.Reason)
}

func TestVerify_WrongPubkey(t *testing.T) {
	kp1, err := GenerateKeypair()
	require.NoError(t, err)
	kp2, err := GenerateKeypair()
	require.NoError(t, err)

	env, err := Sign(TypeExecutionClaim, map[string]any{"x": 1}, kp1)
	require.NoError(t, err)
	env.SignerPub = kp2.PublicHex() // claim someone else signed it
	res := Verify(env)
	require.False(t, res.OK)
	require.Equal(t, "bad_signature", res.Reason)
}

func TestVerify_MissingSig(t *testing.T) {
	kp, err := GenerateKeypair()
	require.NoError(t, err)
	env, err := Sign(TypeExecutionClaim, map[string]any{"x": 1}, kp)
	require.NoError(t, err)
	env.Sig = ""
	res := Verify(env)
	require.False(t, res.OK)
	require.Equal(t, "missing_sig_or_signer", res.Reason)
}

func TestKeypair_AddressDerivedFromPubkey(t *testing.T) {
	kp, err := GenerateKeypair()
	require.NoError(t, err)
	addr := kp.Address()
	// Address is 20 bytes; PublicHex is 128 chars (64 bytes uncompressed,
	// no 0x04 prefix). These are loosely related (address = last 20 of
	// keccak(pubkey)) but the test just enforces the format invariants.
	require.Equal(t, 128, len(kp.PublicHex()))
	require.NotEqual(t, [20]byte{}, addr) // not zero
}
