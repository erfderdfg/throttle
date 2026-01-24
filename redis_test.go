package limiter

import (
	"context"
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

	_, err := NewRedisLimiter(client)
	if err != nil {
		t.Fatalf("Failed to create RedisLimiter: %v", err)
	}
}
