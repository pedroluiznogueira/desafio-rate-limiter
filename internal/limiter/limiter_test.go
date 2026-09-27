package limiter_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"rate-limiter/internal/limiter"
)

func newTestLimiter(t *testing.T, blockDuration time.Duration) *limiter.Limiter {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{Addr: addr})
	if err := client.Ping(context.Background()).Err(); err != nil {
		t.Skipf("redis not reachable at %s, skipping: %v", addr, err)
	}

	return limiter.New(limiter.NewRedisStrategy(client), blockDuration)
}

func uniqueKey(base string) string {
	return base + "-" + time.Now().Format("150405.000000000")
}

func TestAllow_UnderLimit(t *testing.T) {
	lim := newTestLimiter(t, 2*time.Second)
	key := uniqueKey("under-limit")

	for i := 0; i < 3; i++ {
		allowed, err := lim.Allow(context.Background(), key, 5)
		if err != nil {
			t.Fatalf("allow: %v", err)
		}
		if !allowed {
			t.Fatalf("request %d should be allowed (limit 5)", i+1)
		}
	}
}

func TestAllow_BlocksAfterLimitExceeded(t *testing.T) {
	lim := newTestLimiter(t, 2*time.Second)
	key := uniqueKey("over-limit")
	limit := 3

	for i := 0; i < limit; i++ {
		allowed, err := lim.Allow(context.Background(), key, limit)
		if err != nil {
			t.Fatalf("allow: %v", err)
		}
		if !allowed {
			t.Fatalf("request %d should be allowed (limit %d)", i+1, limit)
		}
	}

	if allowed, err := lim.Allow(context.Background(), key, limit); err != nil || allowed {
		t.Fatalf("request over the limit should be blocked (allowed=%v err=%v)", allowed, err)
	}

	// the 1s counting window has passed, but the key should still be inside its block duration
	time.Sleep(1100 * time.Millisecond)
	if allowed, err := lim.Allow(context.Background(), key, limit); err != nil || allowed {
		t.Fatalf("key should still be blocked after the counting window resets (allowed=%v err=%v)", allowed, err)
	}
}

func TestPrecedence_TokenLimitOverridesIP(t *testing.T) {
	lim := newTestLimiter(t, 2*time.Second)

	ipKey := uniqueKey("ip-precedence")
	tokenKey := uniqueKey("token-precedence")
	ipLimit := 2
	tokenLimit := 10

	for i := 0; i < ipLimit; i++ {
		if allowed, err := lim.Allow(context.Background(), ipKey, ipLimit); err != nil || !allowed {
			t.Fatalf("ip request %d should be allowed: %v", i+1, err)
		}
	}
	if allowed, _ := lim.Allow(context.Background(), ipKey, ipLimit); allowed {
		t.Fatal("ip should now be blocked")
	}

	// a token with its own higher limit keeps working well past the ip's own limit
	for i := 0; i < ipLimit+3; i++ {
		allowed, err := lim.Allow(context.Background(), tokenKey, tokenLimit)
		if err != nil {
			t.Fatalf("token allow: %v", err)
		}
		if !allowed {
			t.Fatalf("token request %d should be allowed under its own higher limit (%d)", i+1, tokenLimit)
		}
	}
}
