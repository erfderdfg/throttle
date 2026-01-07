package limiter

import "time"

type state struct {
	tokens     float64
	lastRefill time.Time
}
