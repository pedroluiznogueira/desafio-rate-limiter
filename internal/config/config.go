package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	IPLimit       int
	TokenLimits   map[string]int
	BlockDuration time.Duration
	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

func Load() Config {
	return Config{
		IPLimit:       getInt("RATE_LIMIT_IP", 10),
		TokenLimits:   parseTokenLimits(os.Getenv("RATE_LIMIT_TOKENS")),
		BlockDuration: getDuration("RATE_LIMIT_BLOCK_DURATION", 5*time.Minute),
		RedisAddr:     getString("REDIS_ADDR", "localhost:6379"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB:       getInt("REDIS_DB", 0),
	}
}

// LimitFor returns the token-specific limit and whether one is configured.
// A token without an explicit entry falls back to IP-based limiting.
func (c Config) LimitFor(token string) (int, bool) {
	if token == "" {
		return 0, false
	}
	limit, ok := c.TokenLimits[token]
	return limit, ok
}

func getString(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}

func parseTokenLimits(raw string) map[string]int {
	limits := make(map[string]int)
	if raw == "" {
		return limits
	}

	for _, pair := range strings.Split(raw, ",") {
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) != 2 {
			continue
		}
		limit, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}
		limits[parts[0]] = limit
	}

	return limits
}
