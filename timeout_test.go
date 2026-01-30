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

	limit := Limit{
		Rate:   100,
		Burst:  100,
		Period: time.Second,
	}
	id := Identity{Namespace: "test", Key: "user_cancel"}

	_, err = limiter.Allow(ctx, id, limit)

	if err == nil {
		t.Fatal("Expected an error due to cancelled context, but got nil")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected error to be context.Canceled, but got: %v", err)
	}
}

func TestRedisLimiter_Deadline(t *testing.T) {
	opt, _ := redis.ParseURL("redis://localhost:6379")
	client := redis.NewClient(opt)
	defer client.Close()

	limiter, err := NewRedisLimiter(client)
	if err != nil {
		t.Skipf("Skipping test: Redis not available (%v)", err)
	}

	// Use a deadline already in the past so it is guaranteed to be expired
	// before Allow is called, regardless of local Redis speed.
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	limit := Limit{
		Rate:   100,
		Burst:  100,
		Period: time.Second,
	}
	id := Identity{Namespace: "test", Key: "user_deadline"}

	_, err = limiter.Allow(ctx, id, limit)

	if err == nil {
		t.Fatal("Expected timeout error, but got nil")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected error to be context.DeadlineExceeded, but got: %v", err)
	}
}
