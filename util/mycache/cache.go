package mycache

import (
	"errors"
	"time"
)

type Cache interface {
	Set(key string, value interface{}, d time.Duration) error
	Get(key string) (interface{}, error)
	GetString(key string) (string, error)
}

var (
	ErrNotFound  = errors.New("cache key not found")
	ErrWrongType = errors.New("cache value type is wrong")
)
