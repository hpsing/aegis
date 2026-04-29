package verifier

import (
	"context"

	"github.com/hpsing/aegis/internal/chain"
	"github.com/hpsing/aegis/internal/envelope"
)

// Context bundles per-claim verifier inputs. Today this is just the chain
// client; future per-claim deps (oracle prices, reputation lookups) join
// here without touching call sites.
type Context struct {
	Chain chain.EthClient
}

// CheckClaim is the dispatcher every verifier runs to map claim → verdict.
// It branches on Spec.Action; unsupported actions return FAIL with a
// stable reason rather than abstaining (an action we don't know about
// is a definitive "we can't verify this", not a transient error).
func CheckClaim(ctx context.Context, claim envelope.ExecutionClaim, vctx Context) (Result, error) {
	switch claim.Spec.Action {
	case "uniswap_v3_swap":
		return CheckUniswapSwap(ctx, claim, vctx.Chain)
	default:
		return fail(ReasonUnsupportedAction, Details{}), nil
	}
}
