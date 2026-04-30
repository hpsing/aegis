package keeperhub

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/hpsing/aegis/internal/aegis"
)

// aegisABI parses the abigen-generated AegisContract ABI once at init.
// Required for decoding the calldata that LiveClient.SendTransaction
// receives from internal/verifier/onchain.go's noSendTransactor path.
var aegisABI abi.ABI

func init() {
	parsed, err := abi.JSON(strings.NewReader(aegis.AegisContractABI))
	if err != nil {
		panic(fmt.Sprintf("keeperhub: parse AegisContractABI: %v", err))
	}
	aegisABI = parsed
}

// DecodeCallArgs takes ABI-encoded calldata for one of the three Quorum
// methods we route through KeeperHub (commitVote, revealVote, settle)
// and returns:
//   - the matching Purpose (commit/reveal/settle)
//   - a map of typed arguments ready to send as `execute_workflow.input`
//
// Each value is encoded as a string the way KeeperHub expects:
//   - uint256 → decimal string ("123")
//   - bytes32 → 0x-prefixed hex string ("0xabcd…")
//   - bool    → "true" / "false"
//
// Returns an error if the selector doesn't match one of our three known
// methods. Other methods can't be routed through KeeperHub today.
func DecodeCallArgs(data []byte) (Purpose, map[string]string, error) {
	if len(data) < 4 {
		return "", nil, fmt.Errorf("calldata too short (%d bytes)", len(data))
	}
	selector := data[:4]
	argsRaw := data[4:]

	method, err := matchMethod(selector)
	if err != nil {
		return "", nil, err
	}
	values, err := method.Inputs.Unpack(argsRaw)
	if err != nil {
		return "", nil, fmt.Errorf("unpack %s args: %w", method.Name, err)
	}

	out := make(map[string]string, len(method.Inputs))
	for i, input := range method.Inputs {
		s, err := encodeArgForKH(values[i])
		if err != nil {
			return "", nil, fmt.Errorf("encode %s.%s: %w", method.Name, input.Name, err)
		}
		out[input.Name] = s
	}

	switch method.Name {
	case "commitVote":
		return PurposeCommit, out, nil
	case "revealVote":
		return PurposeReveal, out, nil
	case "settle":
		return PurposeSettle, out, nil
	default:
		return "", nil, fmt.Errorf("unsupported method %q", method.Name)
	}
}

// matchMethod finds the abi.Method whose 4-byte selector matches.
func matchMethod(selector []byte) (*abi.Method, error) {
	for name := range aegisABI.Methods {
		m := aegisABI.Methods[name]
		if bytes.Equal(m.ID, selector) {
			return &m, nil
		}
	}
	return nil, fmt.Errorf("unknown method selector 0x%s", hex.EncodeToString(selector))
}

// encodeArgForKH stringifies one decoded ABI value for the
// execute_workflow.input map. KeeperHub's web3/write-contract action
// re-encodes from these strings using the workflow's stored ABI.
func encodeArgForKH(v any) (string, error) {
	switch x := v.(type) {
	case *big.Int:
		return x.String(), nil
	case bool:
		if x {
			return "true", nil
		}
		return "false", nil
	case [32]byte:
		return "0x" + hex.EncodeToString(x[:]), nil
	case []byte:
		return "0x" + hex.EncodeToString(x), nil
	case string:
		return x, nil
	default:
		return "", fmt.Errorf("unsupported arg type %T", v)
	}
}
