package auth

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisNonceStore struct {
	rdb *redis.Client
}

func NewRedisNonceStore(rdb *redis.Client) NonceStore {
	return &redisNonceStore{rdb: rdb}
}

func (s *redisNonceStore) Save(ctx context.Context, ns, wallet, nonce string, ttl time.Duration) error {
	return s.rdb.Set(ctx, nonceKey(ns, wallet, nonce), "1", ttl).Err()
}

func (s *redisNonceStore) Consume(ctx context.Context, ns, wallet, nonce string) (bool, error) {
	res, err := s.rdb.GetDel(ctx, nonceKey(ns, wallet, nonce)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}
	return res == "1", nil
}
