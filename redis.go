package limiter

import "time"

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
	return func(c *limiterConfig) {
		c.prefix = prefix
	}
}
