package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	limiter "github.com/erfderdfg/throttle"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{Addr: redisAddr})

	// Shared Prometheus registry — exposed on /metrics
	reg := prometheus.NewRegistry()
	rec := limiter.NewPrometheusRecorder(reg)

	// Token bucket: burst-tolerant, for read endpoints
	tbLimiter, err := limiter.NewRedisLimiter(client,
		limiter.WithPrefix("demo:tb:"),
		limiter.WithTimeout(100*time.Millisecond),
		limiter.WithRecorder(rec),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Sliding window: strict, for security-sensitive routes
	swLimiter, err := limiter.NewSlidingWindowLimiter(client,
		limiter.WithPrefix("demo:sw:"),
		limiter.WithTimeout(100*time.Millisecond),
		limiter.WithRecorder(rec),
	)
	if err != nil {
		log.Fatal(err)
	}

	// GET /api/data — burst-tolerant token bucket (100 req/s, burst 200)
	http.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
		id := limiter.Identity{Namespace: "ip", Key: r.RemoteAddr}
		lim := limiter.Limit{Rate: 100, Period: time.Second, Burst: 200}

		dec, err := tbLimiter.Allow(r.Context(), id, lim)
		if err != nil {
			log.Printf("Limiter error: %v", err)
		} else if !dec.Allow {
			w.Header().Set("Retry-After", fmt.Sprintf("%.3f", dec.RetryAfter.Seconds()))
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprintln(w, "Rate limit exceeded")
			return
		}

		fmt.Fprintln(w, "OK")
	})

	// POST /api/login — strict sliding window (5 req/s, no burst tolerance)
	http.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		id := limiter.Identity{Namespace: "ip", Key: r.RemoteAddr}
		lim := limiter.Limit{Rate: 5, Period: time.Second, Burst: 5}

		dec, err := swLimiter.Allow(r.Context(), id, lim)
		if err != nil {
			log.Printf("Limiter error: %v", err)
		} else if !dec.Allow {
			w.Header().Set("Retry-After", fmt.Sprintf("%.3f", dec.RetryAfter.Seconds()))
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprintln(w, "Rate limit exceeded")
			return
		}

		fmt.Fprintln(w, "Authenticated")
	})

	// GET /metrics — Prometheus exposition
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	log.Printf("Server listening on :8080 (Redis: %s)", redisAddr)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
