package limiter

import (
	"context"
	_ "embed"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:embed sliding_window.lua
var slidingWindowScript string

// SlidingWindowLimiter is a distributed rate limiter using a strict sliding window
// backed by a Redis sorted set. Each request is recorded with its timestamp as the
// score; expired entries are pruned on every call.
type SlidingWindowLimiter struct {
	client    *redis.Client
	scriptSHA string
	limiterConfig
}

// NewSlidingWindowLimiter validates connectivity and loads the embedded Lua script
// into Redis (SCRIPT LOAD). The returned limiter is ready to use.
func NewSlidingWindowLimiter(client *redis.Client, opts ...Option) (*SlidingWindowLimiter, error) {
	l := &SlidingWindowLimiter{
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

	if err := client.Ping(ctx).Err(); err != nil { return nil, err }

	sha, err := client.ScriptLoad(ctx, slidingWindowScript).Result()
	if err != nil { return nil, err }
	l.scriptSHA = sha
	return l, nil
}

// Allow checks whether a request for the given identity should be allowed within
// the rolling window defined by limit.Period.
func (l *SlidingWindowLimiter) Allow(ctx context.Context, id Identity, limit Limit) (Decision, error) {
	// TODO: call EvalSha with sliding window args
	return Decision{}, nil
}
