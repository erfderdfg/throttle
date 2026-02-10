package main

import (
	"log"
	"os"
	"time"

	limiter "github.com/erfderdfg/go-rate-limiter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	client := redis.NewClient(&redis.Options{Addr: redisAddr})

	reg := prometheus.NewRegistry()
	rec := limiter.NewPrometheusRecorder(reg)

	tbLimiter, err := limiter.NewRedisLimiter(client,
		limiter.WithPrefix("demo:tb:"),
		limiter.WithTimeout(100*time.Millisecond),
		limiter.WithRecorder(rec),
	)
	if err != nil {
		log.Fatal(err)
	}
	_ = tbLimiter

	log.Printf("Server starting (Redis: %s)", redisAddr)
}
