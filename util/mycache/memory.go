package mycache

import (
	"fmt"
	"time"

	gocache "github.com/patrickmn/go-cache"
)

type MemoryCache struct {
	cache *gocache.Cache
}

var _ Cache = (*MemoryCache)(nil)

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		cache: gocache.New(30*time.Minute, 30*time.Minute),
	}
}

func (c *MemoryCache) Set(key string, value interface{}, d time.Duration) error {
	c.cache.Set(key, value, d)
	return nil
}

func (c *MemoryCache) Get(key string) (interface{}, error) {
	item, _ := c.cache.Get(key)
	return item, nil
}

func (c *MemoryCache) GetString(key string) (string, error) {
	item, _ := c.cache.Get(key)
	if item == nil {
		return "", nil
	}
	if s, ok := item.(string); ok {
		return s, nil
	}
	return "", fmt.Errorf("cache value type is wrong")
}
