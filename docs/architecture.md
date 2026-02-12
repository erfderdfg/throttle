# Architecture

## Overview

This library provides two rate limiter implementations sharing the same `RateLimiter` interface:

- **MemoryLimiter** — in-process token bucket, no external dependencies
- **RedisLimiter** — distributed token bucket, backed by Redis
- **SlidingWindowLimiter** — strict sliding window, backed by Redis sorted sets

## Token Bucket Algorithm

Each identity maps to a "bucket" of tokens:

1. Tokens accumulate over time at `Rate` tokens per `Period`.
2. The bucket is capped at `Burst` tokens.
3. Each `Allow` call consumes one token.
4. If no token is available the request is denied with a `RetryAfter` hint.

## Redis Interaction

`RedisLimiter` and `SlidingWindowLimiter` use `EVALSHA` to run Lua scripts atomically:

- `token_bucket.lua` — atomic read/refill/deduct for the token bucket
- `sliding_window.lua` — atomic prune/count/insert for the sliding window

Lua atomicity means no race conditions between replicas without distributed locks.

## Metrics

Both Redis-backed limiters accept a `MetricsRecorder` via `WithRecorder`. The included
`PrometheusRecorder` exposes:

- `ratelimit_calls_total{namespace, status}` — allow/deny counters
- `ratelimit_errors_total{namespace, type}` — error counters
- `ratelimit_latency_seconds{namespace, status}` — admission latency histogram
