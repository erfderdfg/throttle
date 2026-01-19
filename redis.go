package limiter

import (
	_ "embed"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:embed token_bucket.lua
var tokenBucketScript string

// limiterConfig holds shared configuration for Redis-backed rate limiters.
type limiterConfig struct {
	prefix   string
	timeout  time.Duration
	recorder MetricsRecorder
}

// Option configures a Redis-backed rate limiter.
type Option func(*limiterConfig)

// WithPrefix sets the Redis key prefix. Default is "limiter:".
func WithPrefix(prefix string) Option {
	return func(c *limiterConfig) { c.prefix = prefix }
}

// WithTimeout sets the timeout for Redis operations during initialization. Default is 5s.
func WithTimeout(timeout time.Duration) Option {
	return func(c *limiterConfig) { c.timeout = timeout }
}

// WithRecorder sets the metrics recorder. Default is NoOpMetricsRecorder.
func WithRecorder(recorder MetricsRecorder) Option {
	return func(c *limiterConfig) { c.recorder = recorder }
}

// RedisLimiter is a distributed rate limiter backed by Redis.
//
// It uses a Lua script to perform the token-bucket update atomically, which
// allows multiple application instances to enforce a single shared limit.
type RedisLimiter struct {
	client    *redis.Client
	scriptSHA string
	limiterConfig
}
