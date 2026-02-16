# go-rate-limiter

A lightweight, distributed rate limiting library for Go, built on the **token bucket**
and **sliding window** algorithms.

## Installation

```bash
go get github.com/erfderdfg/go-rate-limiter
```

## Backends

| Backend | State | Use case |
|---------|-------|----------|
| `MemoryLimiter` | In-process | Tests, single-instance deployments |
| `RedisLimiter` | Distributed | Multi-replica token bucket |
| `SlidingWindowLimiter` | Distributed | Strict count over a rolling window |

## Usage

```go
// In-memory (no Redis needed)
ml := limiter.NewMemoryLimiter()
dec, _ := ml.Allow(ctx, limiter.Identity{Namespace: "ip", Key: ip},
    limiter.Limit{Rate: 10, Period: time.Second, Burst: 20})

// Distributed token bucket
rl, _ := limiter.NewRedisLimiter(redisClient, limiter.WithPrefix("myapp:"))
dec, _ = rl.Allow(ctx, id, limit)

// Strict sliding window
sl, _ := limiter.NewSlidingWindowLimiter(redisClient, limiter.WithPrefix("myapp:sw:"))
dec, _ = sl.Allow(ctx, id, limit)

if !dec.Allow {
    w.Header().Set("Retry-After", fmt.Sprintf("%.3f", dec.RetryAfter.Seconds()))
    w.WriteHeader(http.StatusTooManyRequests)
    return
}
```
