package ratelimit

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisRateLimiter struct {
	rdb    *redis.Client
	max    int
	window time.Duration
}

func NewRedisRateLimiter(rdb *redis.Client, max int, window time.Duration) RateLimiter {
	return &redisRateLimiter{
		rdb:    rdb,
		max:    max,
		window: window,
	}
}

func (r *redisRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	pipe := r.rdb.TxPipeline()
	count := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, r.window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	return int(count.Val()) <= r.max, nil
}
