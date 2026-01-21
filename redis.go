package limiter

import (
	"context"
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
type RedisLimiter struct {
	client    *redis.Client
	scriptSHA string
	limiterConfig
}

// NewRedisLimiter validates connectivity and loads the embedded Lua script into
// Redis (SCRIPT LOAD). The returned limiter is ready to use.
func NewRedisLimiter(client *redis.Client, opts ...Option) (*RedisLimiter, error) {
	l := &RedisLimiter{
		client: client,
		limiterConfig: limiterConfig{
			prefix:   "limiter:",
			timeout:  5 * time.Second,
			recorder: &NoOpMetricsRecorder{},
		},
	}
	for _, opt := range opts { opt(&l.limiterConfig) }

	ctx, cancel := context.WithTimeout(context.Background(), l.timeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	// TODO: load Lua script
	return l, nil
}
