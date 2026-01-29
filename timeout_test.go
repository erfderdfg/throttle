package limiter

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestRedisLimiter_ContextCancellation(t *testing.T) {
	opt, _ := redis.ParseURL("redis://localhost:6379")
	client := redis.NewClient(opt)
	defer client.Close()

	limiter, err := NewRedisLimiter(client)
	if err != nil {
		t.Skipf("Skipping test: Redis not available (%v)", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	limit := Limit{Rate: 100, Burst: 100, Period: time.Second}
	id := Identity{Namespace: "test", Key: "user_cancel"}

	_, err = limiter.Allow(ctx, id, limit)
	if err == nil {
		t.Fatal("Expected an error due to cancelled context, but got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled, got: %v", err)
	}
}
