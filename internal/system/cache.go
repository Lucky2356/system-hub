package system

import (
	"sync"
	"time"
)

// listCacheTTL is how long a service/container listing stays fresh. Several
// dashboard consumers (counts, favorites, problem checks) request the same
// listing within one refresh tick; the cache collapses them into a single
// `systemctl list-units` / `docker ps` invocation.
const listCacheTTL = 2500 * time.Millisecond

type listCache[T any] struct {
	mu        sync.Mutex
	data      []T
	fetchedAt time.Time
	valid     bool
}

// get returns cached data if fresh, otherwise calls fetch and caches the
// result. Errors are never cached, so a transient failure retries next call.
func (c *listCache[T]) get(fetch func() ([]T, error)) ([]T, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.valid && time.Since(c.fetchedAt) < listCacheTTL {
		return c.data, nil
	}

	data, err := fetch()
	if err != nil {
		return nil, err
	}

	c.data = data
	c.fetchedAt = time.Now()
	c.valid = true
	return data, nil
}

// invalidate drops the cached data; the next get refetches. Called after
// mutating actions (start/stop/restart) so the UI sees fresh state.
func (c *listCache[T]) invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.valid = false
	c.data = nil
}
