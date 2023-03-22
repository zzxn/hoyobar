package mycache

import (
	"context"
	"log"
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
	err = errors.Wrapf(err, "fail to get %v from redis", key)
	if err != nil {
		log.Println(err)
	}
	return res, err
}

// Set implements Cache
func (r *RedisCache) Set(ctx context.Context, key string, value string, d time.Duration) error {
	err := r.rdb.Set(ctx, key, value, d).Err()
	err = errors.Wrapf(err, "fail to set value with key %v", key)
	if err != nil {
		log.Println(err)
	}
	return err
}

// GetInt64 implements Cache
func (r *RedisCache) GetInt64(ctx context.Context, key string) (int64, error) {
	cmd := r.rdb.Get(ctx, key)
	err := cmd.Err()
	if err == redis.Nil {
		return 0, ErrNotFound
	}
	if err != nil {
		err = errors.Wrapf(err, "fail to get %v from redis", key)
		log.Println(err)
		return 0, err
	}
	value, err := cmd.Int64()
	if err != nil {
		err = errors.Wrapf(err, "fail to parse int64 with key %v", key)
		log.Println(err)
		return 0, err
	}
	return value, nil
}

// SetInt64 implements Cache
func (r *RedisCache) SetInt64(ctx context.Context, key string, value int64, d time.Duration) error {
	err := r.rdb.Set(ctx, key, value, d).Err()
	err = errors.Wrapf(err, "fail to set int64 with key %v", key)
	if err != nil {
		log.Println(err)
	}
	return err
}
