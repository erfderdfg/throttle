package main

import (
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	client := redis.NewClient(&redis.Options{Addr: redisAddr})
	_ = client
	log.Printf("Redis: %s", redisAddr)
}
