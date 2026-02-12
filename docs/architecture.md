# Architecture

## Overview

This library provides two rate limiter implementations sharing the same `RateLimiter` interface:

- **MemoryLimiter** — in-process token bucket, no external dependencies
- **RedisLimiter** — distributed token bucket, backed by Redis
- **SlidingWindowLimiter** — strict sliding window, backed by Redis sorted sets

## Components

*Full documentation in progress...*
