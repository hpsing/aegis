package keeperhub

import (
	"fmt"
	"sync"
)

// idempotencyCache stores prior Receipts keyed by IdempotencyKey so
// repeated SendTransaction calls with the same key return the same
// outcome. Per-process only — production KeeperHub server-side
// deduplication is the canonical guard.
type idempotencyCache struct {
	mu      sync.Mutex
	results map[string]Receipt
}

func newIdempotencyCache() *idempotencyCache {
	return &idempotencyCache{results: map[string]Receipt{}}
}

// get returns (receipt, true) if a prior result is cached for key.
func (c *idempotencyCache) get(key string) (Receipt, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	r, ok := c.results[key]
	return r, ok
}

// put records the receipt under key. Subsequent get(key) calls return
// the same Receipt.
func (c *idempotencyCache) put(key string, r Receipt) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.results[key] = r
}

// formatKey is the canonical idempotency key for a (jobID, purpose).
func formatKey(jobID uint64, p Purpose) string {
	return fmt.Sprintf("%d-%s", jobID, p)
}
