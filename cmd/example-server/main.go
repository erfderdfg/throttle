package main

import (
	"log"
	"os"

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
	_ = rec

	log.Printf("Server starting (Redis: %s)", redisAddr)
}
