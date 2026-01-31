package limiter

import (
	"context"
	_ "embed"
	"errors"
	"strconv"
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
	return func(c *limiterConfig) {
		c.prefix = prefix
	}
}

// WithTimeout sets the timeout for Redis operations during initialization. Default is 5s.
func WithTimeout(timeout time.Duration) Option {
	return func(c *limiterConfig) {
		c.timeout = timeout
	}
}

// WithRecorder sets the metrics recorder. Default is NoOpMetricsRecorder.
func WithRecorder(recorder MetricsRecorder) Option {
	return func(c *limiterConfig) {
		c.recorder = recorder
	}
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

// NewRedisLimiter validates connectivity and loads the embedded Lua script into
// Redis (SCRIPT LOAD). The returned limiter is ready to use.
func NewRedisLimiter(client *redis.Client, opts ...Option) (*RedisLimiter, error) {
	limiter := &RedisLimiter{
		client: client,
		limiterConfig: limiterConfig{
			prefix:   "limiter:",
			timeout:  5 * time.Second,
			recorder: &NoOpMetricsRecorder{},
		},
	}

	for _, opt := range opts {
		opt(&limiter.limiterConfig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), limiter.timeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	sha, err := client.ScriptLoad(ctx, tokenBucketScript).Result()
	if err != nil {
		return nil, err
	}

	limiter.scriptSHA = sha
	return limiter, nil
}

// Allow checks whether a request for the given identity should be allowed under
// the provided limit. Each call has a fixed cost of 1 token.
func (r *RedisLimiter) Allow(ctx context.Context, id Identity, limit Limit) (Decision, error) {
	start := time.Now()
	status := "error"

	defer func() {
		r.recorder.Observe("ratelimit.latency", time.Since(start).Seconds(), map[string]string{
			"namespace": string(id.Namespace),
			"status":    status,
		})
	}()

	key := r.prefix + string(id.Namespace) + ":" + id.Key
	now := float64(time.Now().UnixMicro()) / 1e6
	cost := 1.0
	ratePerSecond := float64(limit.Rate) / limit.Period.Seconds()

	cmd := r.client.EvalSha(ctx, r.scriptSHA, []string{key},
		ratePerSecond,
		limit.Burst,
		now,
		cost,
	)

	result, err := cmd.Result()
	if err != nil {
		r.recorder.Add("ratelimit.errors", 1, map[string]string{
			"namespace": string(id.Namespace),
			"type":      "redis_eval",
		})
		return Decision{}, err
	}

	values, ok := result.([]interface{})
	if !ok || len(values) != 4 {
		r.recorder.Add("ratelimit.errors", 1, map[string]string{
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

	r.recorder.Add("ratelimit.call", 1, map[string]string{
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

func convertToFloat(val interface{}) float64 {
	switch v := val.(type) {
	case int64:
		return float64(v)
	case float64:
		return v
	case string:
		f, _ := strconv.ParseFloat(v, 64)
		return f
	default:
		return 0
	}
}
