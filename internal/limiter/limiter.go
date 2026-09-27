package limiter

import (
	"context"
	"time"
)

type Strategy interface {
	Increment(ctx context.Context, key string, window time.Duration) (int64, error)
	IsBlocked(ctx context.Context, key string) (bool, error)
	Block(ctx context.Context, key string, duration time.Duration) error
}

type Limiter struct {
	Strategy      Strategy
	BlockDuration time.Duration
}

func New(strategy Strategy, blockDuration time.Duration) *Limiter {
	return &Limiter{Strategy: strategy, BlockDuration: blockDuration}
}

func (l *Limiter) Allow(ctx context.Context, key string, limit int) (bool, error) {
	blocked, err := l.Strategy.IsBlocked(ctx, key)
	if err != nil {
		return false, err
	}
	if blocked {
		return false, nil
	}

	count, err := l.Strategy.Increment(ctx, key, time.Second)
	if err != nil {
		return false, err
	}

	if count > int64(limit) {
		if err := l.Strategy.Block(ctx, key, l.BlockDuration); err != nil {
			return false, err
		}
		return false, nil
	}

	return true, nil
}
