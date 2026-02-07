package limiter

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestSlidingWindowLimiter_Integration(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("Skipping integration test: Redis not available (%v)", err)
	}

	limiter, err := NewSlidingWindowLimiter(client)
	if err != nil {
		t.Fatalf("Failed to create SlidingWindowLimiter: %v", err)
	}

	t.Run("BasicFlow", func(t *testing.T) {
		key := fmt.Sprintf("sw_basic_%d", time.Now().UnixNano())
		id := Identity{Namespace: "sw_test", Key: key}
		limit := Limit{Rate: 2, Period: time.Second, Burst: 2}

		dec, err := limiter.Allow(ctx, id, limit)
		if err != nil { t.Fatalf("Redis error: %v", err) }
		if !dec.Allow { t.Error("Expected first request to be allowed") }
		if dec.Remaining != 1 { t.Errorf("Expected 1 remaining, got %d", dec.Remaining) }

		dec, _ = limiter.Allow(ctx, id, limit)
		if !dec.Allow { t.Error("Expected second request to be allowed") }

		dec, _ = limiter.Allow(ctx, id, limit)
		if dec.Allow { t.Error("Expected third request to be denied (limit=2)") }
	})
}
