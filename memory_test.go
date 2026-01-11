package limiter

import (
	"context"
	"testing"
	"time"
)

func TestMemoryLimiter_Allow_Basics(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	limiter := NewMemoryLimiter()

	limit := Limit{Rate: 10, Period: time.Second, Burst: 10}
	id := Identity{Namespace: "test", Key: "user_1"}

	decision, _ := limiter.Allow(ctx, id, limit)

	if !decision.Allow {
		t.Error("Expected request to be allowed, but got denied!.")
	}
	if decision.Remaining != 9 {
		t.Logf("Expected 9 remaining tokens got %d instead!", decision.Remaining)
	}
}
