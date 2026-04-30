package keeperhub

import (
	"errors"
	"testing"
)

func TestIsTransientError(t *testing.T) {
	cases := []struct {
		name      string
		err       error
		transient bool
	}{
		// Permanent: contract reverts and config bugs — must NOT retry.
		{"contract revert", errors.New("Contract call failed: execution reverted (unknown custom error)"), false},
		{"bigint coercion", errors.New(`Cannot convert {{jobId}} to a BigInt`), false},
		{"trim bug", errors.New("$.trim is not a function"), false},
		{"json parse bug", errors.New("Invalid function arguments JSON: Unexpected token"), false},
		{"purpose mismatch", errors.New(`purpose mismatch: calldata says "commit"`), false},
		{"unauthorized", errors.New("server returned 401 Unauthorized"), false},

		// Transient: KH's dRPC free tier and generic network errors.
		{"408 timeout", errors.New("server response 408 Request Timeout"), true},
		{"drpc free tier", errors.New(`responseBody: {"message":"Request timeout on the free tier"}`), true},
		{"503 unavailable", errors.New("upstream returned 503"), true},
		{"504 gateway", errors.New("server response 504 Gateway Timeout"), true},
		{"connection reset", errors.New("read tcp: connection reset by peer"), true},
		{"i/o timeout", errors.New("dial tcp: i/o timeout"), true},

		// Edge cases.
		{"nil error", nil, false},
		{"unknown error", errors.New("something went wrong"), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := isTransientError(tc.err)
			if got != tc.transient {
				t.Errorf("isTransientError(%q) = %v, want %v", errString(tc.err), got, tc.transient)
			}
		})
	}
}

func errString(err error) string {
	if err == nil {
		return "<nil>"
	}
	return err.Error()
}
