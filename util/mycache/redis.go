package mycache

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	rdb *redis.Client
}

var _ Cache = (*RedisCache)(nil)

func NewRedisCache(rdb *redis.Client) *RedisCache {
	return &RedisCache{
		rdb: rdb,
	}
}

// Get implements Cache
func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	res, err := r.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", ErrNotFound
	}
	return res, errors.Wrapf(err, "fail to get %v from redis", key)
}

// Set implements Cache
func (r *RedisCache) Set(ctx context.Context, key string, value string, d time.Duration) error {
	err := r.rdb.Set(ctx, key, value, d).Err()
	return errors.Wrapf(err, "fail to get %v from redis", key)
}
