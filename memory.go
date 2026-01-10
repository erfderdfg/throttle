package limiter

import (
	"context"
	"sync"
	"time"
)

type state struct {
	tokens     float64
	lastRefill time.Time
}

// MemoryLimiter is an in-process token-bucket rate limiter.
//
// It is safe for concurrent use by multiple goroutines, but its state is local
// to the process and is not shared across replicas. Use RedisLimiter when you
// need a single global limit across multiple instances.
type MemoryLimiter struct {
	mu      sync.Mutex
	buckets map[string]*state
}

// NewMemoryLimiter constructs a MemoryLimiter with empty state.
func NewMemoryLimiter() *MemoryLimiter {
	return &MemoryLimiter{
		buckets: make(map[string]*state),
	}
}

// Allow checks whether a request for the given identity should be allowed under
// the provided limit. Each call has a fixed cost of 1 token.
func (m *MemoryLimiter) Allow(ctx context.Context, id Identity, limit Limit) (Decision, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	key := string(id.Namespace) + ":" + id.Key
	st, exists := m.buckets[key]
	if !exists {
		m.buckets[key] = &state{
			tokens:     float64(limit.Burst) - 1,
			lastRefill: now,
		}
		return Decision{
			Allow:      true,
			Remaining:  limit.Burst - 1,
			RetryAfter: 0,
			ResetTime:  now,
		}, nil
	}

	elapsed := now.Sub(st.lastRefill)
	if elapsed < 0 {
		elapsed = 0
	}
	delta := float64(elapsed) / float64(limit.Period)
	tokensToAdd := delta * float64(limit.Rate)
	newBalance := st.tokens + tokensToAdd
	if newBalance > float64(limit.Burst) {
		newBalance = float64(limit.Burst)
	}
	st.tokens = newBalance
	st.lastRefill = now
	// TODO: allow/deny check
	return Decision{Allow: true, Remaining: int64(st.tokens), ResetTime: now}, nil
}
