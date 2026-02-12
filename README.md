# go-rate-limiter

A lightweight, distributed rate limiting library for Go.

## Installation

```bash
go get github.com/erfderdfg/go-rate-limiter
```

## Quick Start

```go
import limiter "github.com/erfderdfg/go-rate-limiter"

ml := limiter.NewMemoryLimiter()
dec, _ := ml.Allow(ctx, limiter.Identity{Namespace: "ip", Key: r.RemoteAddr},
    limiter.Limit{Rate: 10, Period: time.Second, Burst: 20})
if !dec.Allow {
    w.WriteHeader(http.StatusTooManyRequests)
    return
}
```

More documentation coming soon.
