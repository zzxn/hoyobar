package mycache

import (
	"context"
	"errors"
	"time"
)

type Cache interface {
	Set(ctx context.Context, key string, value string, d time.Duration) error
	Get(ctx context.Context, key string) (string, error)
}

var (
	ErrNotFound = errors.New("cache key not found")
)
