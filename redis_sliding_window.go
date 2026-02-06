package limiter

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:embed sliding_window.lua
var slidingWindowScript string

// SlidingWindowLimiter is a distributed rate limiter using a strict sliding window
// backed by a Redis sorted set. Each request is recorded with its timestamp as the
// score; expired entries are pruned on every call. Unlike a token bucket, this
// algorithm does not accumulate burst capacity — it enforces an exact count over
// a rolling window, making it well-suited for security-sensitive routes.
type SlidingWindowLimiter struct {
	client    *redis.Client
	scriptSHA string
	limiterConfig
}

// NewSlidingWindowLimiter validates connectivity and loads the embedded Lua script
// into Redis (SCRIPT LOAD). The returned limiter is ready to use.
//
// Accepts the same Option values as NewRedisLimiter (WithPrefix, WithTimeout,
// WithRecorder).
func NewSlidingWindowLimiter(client *redis.Client, opts ...Option) (*SlidingWindowLimiter, error) {
	l := &SlidingWindowLimiter{
		client: client,
		limiterConfig: limiterConfig{
			prefix:   "limiter:",
			timeout:  5 * time.Second,
			recorder: &NoOpMetricsRecorder{},
		},
	}

	for _, opt := range opts {
		opt(&l.limiterConfig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), l.timeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	sha, err := client.ScriptLoad(ctx, slidingWindowScript).Result()
	if err != nil {
		return nil, err
	}

	l.scriptSHA = sha
	return l, nil
}

// Allow checks whether a request for the given identity should be allowed within
// the rolling window defined by limit.Period. The window holds at most limit.Rate
// requests; limit.Burst is not used. Each call has a fixed cost of 1.
func (l *SlidingWindowLimiter) Allow(ctx context.Context, id Identity, limit Limit) (Decision, error) {
	start := time.Now()
	status := "error"

	defer func() {
		l.recorder.Observe("ratelimit.latency", time.Since(start).Seconds(), map[string]string{
			"namespace": string(id.Namespace),
			"status":    status,
		})
	}()

	key := l.prefix + string(id.Namespace) + ":" + id.Key
	now := float64(time.Now().UnixMicro()) / 1e6
	cost := 1.0
	nonce := fmt.Sprintf("%d", rand.Int64())

	cmd := l.client.EvalSha(ctx, l.scriptSHA, []string{key},
		limit.Rate,            // ARGV[1]: max requests in window
		limit.Period.Seconds(), // ARGV[2]: window duration in seconds
		now,                   // ARGV[3]: current timestamp
		cost,                  // ARGV[4]: cost of this request
		nonce,                 // ARGV[5]: unique member ID
	)

	result, err := cmd.Result()
	if err != nil {
		l.recorder.Add("ratelimit.errors", 1, map[string]string{
			"namespace": string(id.Namespace),
			"type":      "redis_eval",
		})
		return Decision{}, err
	}

	values, ok := result.([]interface{})
	if !ok || len(values) != 4 {
		l.recorder.Add("ratelimit.errors", 1, map[string]string{
			"namespace": string(id.Namespace),
			"type":      "invalid_format",
		})
		return Decision{}, errors.New("invalid lua response format")
	}

	allowedVal := int64(convertToFloat(values[0]))
	remainingVal := int64(convertToFloat(values[1]))
	retryAfterFloat := convertToFloat(values[2])
	resetTimeFloat := convertToFloat(values[3])

	if allowedVal == 1 {
		status = "allowed"
	} else {
		status = "denied"
	}

	l.recorder.Add("ratelimit.call", 1, map[string]string{
		"namespace": string(id.Namespace),
		"status":    status,
	})

	return Decision{
		Allow:      allowedVal == 1,
		Remaining:  remainingVal,
		RetryAfter: time.Duration(retryAfterFloat * float64(time.Second)),
		ResetTime:  time.UnixMicro(int64(resetTimeFloat * 1e6)),
	}, nil
}
