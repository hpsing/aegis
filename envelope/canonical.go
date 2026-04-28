// Package envelope owns the wire format every Quorum agent speaks: canonical
// JSON, secp256k1 signing (matching Ethereum's signing scheme), the
// envelope wrapper, and the typed payload structs.
//
// The wire format is FROZEN at Version=1. Any change to the JSON shape
// requires bumping Version. Internal signing-scheme changes bump
// SigVersion (a Go-internal const) without touching the wire shape.
package envelope

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

// Canonicalize returns deterministic JSON: sorted object keys, no extraneous
// whitespace. The on-chain side computes keccak256 of these same bytes,
// so the Go-and-Solidity digests must agree byte-for-byte (see
// contracts/test/CanonicalDigest.t.sol once step 3 is in).
func Canonicalize(v any) ([]byte, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("canonicalize marshal: %w", err)
	}
	var generic any
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&generic); err != nil {
		return nil, fmt.Errorf("canonicalize unmarshal: %w", err)
	}
	var buf bytes.Buffer
	if err := writeCanonical(&buf, generic); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func writeCanonical(buf *bytes.Buffer, v any) error {
	switch t := v.(type) {
	case nil:
		buf.WriteString("null")
		return nil
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		buf.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				buf.WriteByte(',')
			}
			kb, err := json.Marshal(k)
			if err != nil {
				return err
			}
			buf.Write(kb)
			buf.WriteByte(':')
			if err := writeCanonical(buf, t[k]); err != nil {
				return err
			}
		}
		buf.WriteByte('}')
		return nil
	case []any:
		buf.WriteByte('[')
		for i, item := range t {
			if i > 0 {
				buf.WriteByte(',')
			}
			if err := writeCanonical(buf, item); err != nil {
				return err
			}
		}
		buf.WriteByte(']')
		return nil
	default:
		raw, err := json.Marshal(v)
		if err != nil {
			return err
		}
		buf.Write(raw)
		return nil
	}
}
