package chain

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// PublicClient is the production EthClient. It wraps *ethclient.Client
// with primary + fallback URLs and per-call retry. If both URLs fail
// the caller receives the last error and SHOULD abstain rather than
// vote (better to lose the bounty than be slashed for a bad read).
type PublicClient struct {
	primary    *ethclient.Client
	fallback   *ethclient.Client
	retryCount int
	timeout    time.Duration
}

// PublicClientConfig holds RPC URLs read from env. RetryCount and Timeout
// have sensible defaults if zero.
type PublicClientConfig struct {
	PrimaryURL  string
	FallbackURL string
	RetryCount  int
	Timeout     time.Duration
}

// NewPublicClient dials primary and (if non-empty) fallback. Either may
// be nil if dial fails — calls degrade gracefully.
func NewPublicClient(ctx context.Context, cfg PublicClientConfig) (*PublicClient, error) {
	if cfg.PrimaryURL == "" {
		return nil, errors.New("chain: PrimaryURL required")
	}
	if cfg.RetryCount == 0 {
		cfg.RetryCount = 3
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Second
	}
	primary, err := ethclient.DialContext(ctx, cfg.PrimaryURL)
	if err != nil {
		return nil, fmt.Errorf("dial primary %s: %w", cfg.PrimaryURL, err)
	}
	c := &PublicClient{
		primary:    primary,
		retryCount: cfg.RetryCount,
		timeout:    cfg.Timeout,
	}
	if cfg.FallbackURL != "" {
		fb, err := ethclient.DialContext(ctx, cfg.FallbackURL)
		if err == nil {
			c.fallback = fb
		}
		// Fallback dial failure is non-fatal — primary still works.
	}
	return c, nil
}

func (c *PublicClient) Close() {
	if c.primary != nil {
		c.primary.Close()
	}
	if c.fallback != nil {
		c.fallback.Close()
	}
}

func (c *PublicClient) TransactionReceipt(
	ctx context.Context, txHash common.Hash,
) (*types.Receipt, error) {
	var receipt *types.Receipt
	err := c.tryBoth(ctx, func(ec *ethclient.Client) error {
		callCtx, cancel := context.WithTimeout(ctx, c.timeout)
		defer cancel()
		r, err := ec.TransactionReceipt(callCtx, txHash)
		if err != nil {
			return err
		}
		receipt = r
		return nil
	})
	return receipt, err
}

func (c *PublicClient) ChainID(ctx context.Context) (*big.Int, error) {
	var id *big.Int
	err := c.tryBoth(ctx, func(ec *ethclient.Client) error {
		callCtx, cancel := context.WithTimeout(ctx, c.timeout)
		defer cancel()
		v, err := ec.ChainID(callCtx)
		if err != nil {
			return err
		}
		id = v
		return nil
	})
	return id, err
}

func (c *PublicClient) BlockNumber(ctx context.Context) (uint64, error) {
	var n uint64
	err := c.tryBoth(ctx, func(ec *ethclient.Client) error {
		callCtx, cancel := context.WithTimeout(ctx, c.timeout)
		defer cancel()
		v, err := ec.BlockNumber(callCtx)
		if err != nil {
			return err
		}
		n = v
		return nil
	})
	return n, err
}

// tryBoth runs op against primary, retries up to retryCount, and falls
// back to the secondary client if the primary keeps failing. ethereum.NotFound
// is wrapped as ErrTxNotFound so callers can branch cleanly.
func (c *PublicClient) tryBoth(ctx context.Context, op func(*ethclient.Client) error) error {
	var lastErr error
	for _, ec := range []*ethclient.Client{c.primary, c.fallback} {
		if ec == nil {
			continue
		}
		for attempt := 0; attempt < c.retryCount; attempt++ {
			if err := ctx.Err(); err != nil {
				return err
			}
			err := op(ec)
			if err == nil {
				return nil
			}
			if errors.Is(err, ethereum.NotFound) {
				return fmt.Errorf("%w: %v", ErrTxNotFound, err)
			}
			// Don't retry on cancellation.
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				lastErr = err
				continue
			}
			// Don't retry on obvious caller errors (bad address etc.).
			if isFatal(err) {
				return err
			}
			lastErr = err
			// Linear backoff between attempts.
			time.Sleep(time.Duration(attempt+1) * 150 * time.Millisecond)
		}
	}
	if lastErr == nil {
		lastErr = errors.New("no RPC client available")
	}
	return lastErr
}

// isFatal flags errors we shouldn't retry. Currently a heuristic — we
// could expand this if specific JSON-RPC error codes need carve-outs.
func isFatal(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "invalid argument") ||
		strings.Contains(msg, "invalid method")
}
