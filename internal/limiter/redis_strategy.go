package limiter

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStrategy struct {
	Client *redis.Client
}

func NewRedisStrategy(client *redis.Client) *RedisStrategy {
	return &RedisStrategy{Client: client}
}

func (r *RedisStrategy) Increment(ctx context.Context, key string, window time.Duration) (int64, error) {
	countKey := "ratelimit:count:" + key

	count, err := r.Client.Incr(ctx, countKey).Result()
	if err != nil {
		return 0, err
	}

	if count == 1 {
		r.Client.Expire(ctx, countKey, window)
	}

	return count, nil
}

func (r *RedisStrategy) IsBlocked(ctx context.Context, key string) (bool, error) {
	blockKey := "ratelimit:block:" + key

	exists, err := r.Client.Exists(ctx, blockKey).Result()
	if err != nil {
		return false, err
	}

	return exists > 0, nil
}

func (r *RedisStrategy) Block(ctx context.Context, key string, duration time.Duration) error {
	blockKey := "ratelimit:block:" + key
	return r.Client.Set(ctx, blockKey, "1", duration).Err()
}
