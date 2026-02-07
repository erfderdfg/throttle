package limiter

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestSlidingWindowLimiter_Integration(t *testing.T) {
	opts := &redis.Options{Addr: "localhost:6379"}
	client := redis.NewClient(opts)

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
		// Max 2 requests per second, strict (no burst accumulation)
		limit := Limit{Rate: 2, Period: time.Second, Burst: 2}

		dec, err := limiter.Allow(ctx, id, limit)
		if err != nil {
			t.Fatalf("Redis error: %v", err)
		}
		if !dec.Allow {
			t.Error("Expected first request to be allowed")
		}
		if dec.Remaining != 1 {
			t.Errorf("Expected 1 remaining, got %d", dec.Remaining)
		}

		dec, err = limiter.Allow(ctx, id, limit)
		if err != nil {
			t.Fatal(err)
		}
		if !dec.Allow {
			t.Error("Expected second request to be allowed")
		}

		dec, err = limiter.Allow(ctx, id, limit)
		if err != nil {
			t.Fatal(err)
		}
		if dec.Allow {
			t.Error("Expected third request to be denied (limit=2)")
		}
		if dec.RetryAfter <= 0 {
			t.Error("Expected positive RetryAfter on denial")
		}
	})

	t.Run("WindowExpiry", func(t *testing.T) {
		key := fmt.Sprintf("sw_expiry_%d", time.Now().UnixNano())
		id := Identity{Namespace: "sw_test", Key: key}
		// Max 1 request per 100ms window
		limit := Limit{Rate: 1, Period: 100 * time.Millisecond, Burst: 1}

		dec, err := limiter.Allow(ctx, id, limit)
		if err != nil {
			t.Fatal(err)
		}
		if !dec.Allow {
			t.Fatal("Expected first request to be allowed")
		}

		dec, err = limiter.Allow(ctx, id, limit)
		if err != nil {
			t.Fatal(err)
		}
		if dec.Allow {
			t.Fatal("Expected second request to be denied before window expires")
		}

		// Wait for the window to expire
		time.Sleep(120 * time.Millisecond)

		dec, err = limiter.Allow(ctx, id, limit)
		if err != nil {
			t.Fatal(err)
		}
		if !dec.Allow {
			t.Error("Expected request to be allowed after window expiry")
		}
	})

	t.Run("DistributedState", func(t *testing.T) {
		key := fmt.Sprintf("sw_dist_%d", time.Now().UnixNano())
		id := Identity{Namespace: "sw_test", Key: key}
		limit := Limit{Rate: 1, Period: time.Second, Burst: 1}

		limiterA, _ := NewSlidingWindowLimiter(client)
		limiterA.Allow(ctx, id, limit)

		limiterB, _ := NewSlidingWindowLimiter(client)
		dec, err := limiterB.Allow(ctx, id, limit)
		if err != nil {
			t.Fatal(err)
		}
		if dec.Allow {
			t.Error("Instance B should see the request recorded by Instance A")
		}
	})
}
