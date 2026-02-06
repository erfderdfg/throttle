package limiter

import (
	_ "embed"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:embed sliding_window.lua
var slidingWindowScript string

// SlidingWindowLimiter is a distributed rate limiter using a strict sliding window
// backed by a Redis sorted set.
type SlidingWindowLimiter struct {
	client    *redis.Client
	scriptSHA string
	limiterConfig
}
