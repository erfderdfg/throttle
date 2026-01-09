package limiter

import (
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
