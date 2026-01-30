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

type limiterConfig struct {
	prefix   string
	timeout  time.Duration
	recorder MetricsRecorder
}

type Option func(*limiterConfig)

func WithPrefix(prefix string) Option  { return func(c *limiterConfig) { c.prefix = prefix } }
func WithTimeout(d time.Duration) Option { return func(c *limiterConfig) { c.timeout = d } }
func WithRecorder(r MetricsRecorder) Option { return func(c *limiterConfig) { c.recorder = r } }

// RedisLimiter is a distributed rate limiter backed by Redis.
type RedisLimiter struct {
	client    *redis.Client
	scriptSHA string
	limiterConfig
}

func NewRedisLimiter(client *redis.Client, opts ...Option) (*RedisLimiter, error) {
	l := &RedisLimiter{client: client, limiterConfig: limiterConfig{
		prefix: "limiter:", timeout: 5 * time.Second, recorder: &NoOpMetricsRecorder{},
	}}
	for _, opt := range opts { opt(&l.limiterConfig) }
	ctx, cancel := context.WithTimeout(context.Background(), l.timeout)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil { return nil, err }
	sha, err := client.ScriptLoad(ctx, tokenBucketScript).Result()
	if err != nil { return nil, err }
	l.scriptSHA = sha
	return l, nil
}

func (r *RedisLimiter) Allow(ctx context.Context, id Identity, limit Limit) (Decision, error) {
	start := time.Now()

	defer func() {
		r.recorder.Observe("ratelimit.latency", time.Since(start).Seconds(), map[string]string{
			"namespace": string(id.Namespace),
			"status":    "ok",
		})
	}()

	key := r.prefix + string(id.Namespace) + ":" + id.Key
	now := float64(time.Now().UnixMicro()) / 1e6
	cost := 1.0
	ratePerSecond := float64(limit.Rate) / limit.Period.Seconds()

	result, err := r.client.EvalSha(ctx, r.scriptSHA, []string{key},
		ratePerSecond, limit.Burst, now, cost,
	).Result()
	if err != nil { return Decision{}, err }

	values, ok := result.([]interface{})
	if !ok || len(values) != 4 {
		return Decision{}, errors.New("invalid lua response format")
	}

	allowedVal   := int64(convertToFloat(values[0]))
	remainingVal := int64(convertToFloat(values[1]))
	retryAfterF  := convertToFloat(values[2])
	resetTimeF   := convertToFloat(values[3])

	return Decision{
		Allow:      allowedVal == 1,
		Remaining:  remainingVal,
		RetryAfter: time.Duration(retryAfterF * float64(time.Second)),
		ResetTime:  time.UnixMicro(int64(resetTimeF * 1e6)),
	}, nil
}

func convertToFloat(val interface{}) float64 {
	switch v := val.(type) {
	case int64:   return float64(v)
	case float64: return v
	case string:  f, _ := strconv.ParseFloat(v, 64); return f
	default:      return 0
	}
}
