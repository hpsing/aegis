package envelope

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCanonicalize_SortsObjectKeys(t *testing.T) {
	a, err := Canonicalize(map[string]any{"b": 1, "a": 2})
	require.NoError(t, err)
	b, err := Canonicalize(map[string]any{"a": 2, "b": 1})
	require.NoError(t, err)
	require.Equal(t, string(a), string(b))
	require.Equal(t, `{"a":2,"b":1}`, string(a))
}

func TestCanonicalize_NestedSorts(t *testing.T) {
	got, err := Canonicalize(map[string]any{
		"x": map[string]any{"b": 1, "a": 2},
		"y": []any{3, 1, 2},
	})
	require.NoError(t, err)
	// Outer keys sorted: x then y. Inner object's keys also sorted.
	// Array order preserved.
	require.Equal(t, `{"x":{"a":2,"b":1},"y":[3,1,2]}`, string(got))
}

func TestCanonicalize_PrimitivesRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want string
	}{
		{"null", nil, "null"},
		{"true", true, "true"},
		{"false", false, "false"},
		{"int", 42, "42"},
		{"float", 1.5, "1.5"},
		{"empty string", "", `""`},
		{"unicode string", "héllo 🌍", `"héllo 🌍"`},
		{"empty array", []any{}, "[]"},
		{"empty object", map[string]any{}, "{}"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := Canonicalize(c.in)
			require.NoError(t, err)
			require.Equal(t, c.want, string(got))
		})
	}
}

func TestCanonicalize_StructWithJSONTags(t *testing.T) {
	type Inner struct {
		Z int `json:"z"`
		A int `json:"a"`
	}
	type Outer struct {
		Inner Inner  `json:"inner"`
		Name  string `json:"name"`
	}
	got, err := Canonicalize(Outer{Inner: Inner{Z: 1, A: 2}, Name: "x"})
	require.NoError(t, err)
	// Field order in Go source doesn't matter — keys come out sorted.
	require.Equal(t, `{"inner":{"a":2,"z":1},"name":"x"}`, string(got))
}

func TestCanonicalize_SpecMatchesAcrossFieldOrder(t *testing.T) {
	// Two structurally-equal ClaimSpecs marshalled in different orders
	// must produce identical canonical bytes — this is the cross-language
	// invariant that the on-chain side relies on.
	s1 := ClaimSpec{Action: "uniswap_v3_swap", ChainID: 8453, FeeTier: 500}
	s2 := ClaimSpec{FeeTier: 500, Action: "uniswap_v3_swap", ChainID: 8453}
	a, err := Canonicalize(s1)
	require.NoError(t, err)
	b, err := Canonicalize(s2)
	require.NoError(t, err)
	require.Equal(t, string(a), string(b))
}
