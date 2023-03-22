package mycache

import (
	"context"
	"errors"
	"math/rand"
	"time"
)

type Cache interface {
	Set(ctx context.Context, key string, value string, d time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	SetInt64(ctx context.Context, key string, value int64, d time.Duration) error
	GetInt64(ctx context.Context, key string) (int64, error)
}

var (
	ErrNotFound = errors.New("cache key not found")
)

func RandomExpire(d time.Duration, jitter float64) time.Duration {
	scale := 1.0 + 2.0*(rand.Float64()-0.5)*jitter
	ms := d.Milliseconds()
	return time.Duration(scale*float64(ms)) * time.Millisecond
}
