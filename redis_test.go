package limiter

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestRedisLimiter_Integration(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("Skipping integration test: Redis not available (%v)", err)
	}

	limiter, err := NewRedisLimiter(client)
	if err != nil {
		t.Fatalf("Failed to create RedisLimiter: %v", err)
	}

	t.Run("BasicFlow", func(t *testing.T) {
		key := fmt.Sprintf("it_test_%d", time.Now().UnixNano())
		id := Identity{Namespace: "integration", Key: key}
		limit := Limit{Rate: 10, Period: time.Second, Burst: 2}

		dec, err := limiter.Allow(ctx, id, limit)
		if err != nil { t.Fatalf("Redis error: %v", err) }
		if !dec.Allow { t.Error("Expected first request to be Allowed") }
		if dec.Remaining != 1 { t.Errorf("Expected 1 remaining, got %d", dec.Remaining) }

		dec, err = limiter.Allow(ctx, id, limit)
		if err != nil { t.Fatal(err) }
		if !dec.Allow { t.Error("Expected second request to be Allowed") }

		dec, err = limiter.Allow(ctx, id, limit)
		if err != nil { t.Fatal(err) }
		if dec.Allow { t.Error("Expected third request to be Denied") }
		if dec.RetryAfter <= 0 { t.Error("Expected positive RetryAfter on denial") }
	})
}
