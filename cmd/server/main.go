package main

import (
	"log"
	"net/http"

	"github.com/redis/go-redis/v9"

	"rate-limiter/internal/config"
	"rate-limiter/internal/httpmiddleware"
	"rate-limiter/internal/limiter"
)

func main() {
	cfg := config.Load()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	strategy := limiter.NewRedisStrategy(redisClient)
	lim := limiter.New(strategy, cfg.BlockDuration)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	handler := httpmiddleware.RateLimit(cfg, lim)(mux)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
