package cache

import (
	"context"
	"sync"
	"time"
)

type cacheItem struct {
	value      interface{}
	expiration time.Time
}

type Cache struct {
	sync.RWMutex
	items map[string]cacheItem
}

func NewWithCleanupInterval(interval time.Duration) *Cache {
	cache := &Cache{
		items: make(map[string]cacheItem),
	}
	go cache.startCleanup(interval)
	return cache
}

func (c *Cache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	c.Lock()
	defer c.Unlock()

	c.items[key] = cacheItem{
		value:      value,
		expiration: time.Now().Add(expiration),
	}
	return nil
}

func (c *Cache) Get(ctx context.Context, key string) (interface{}, error) {
	c.RLock()
	defer c.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, nil
	}

	if !item.expiration.IsZero() && item.expiration.Before(time.Now()) {
		return nil, nil
	}

	return item.value, nil
}

func (c *Cache) startCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		c.cleanup()
	}
}

func (c *Cache) cleanup() {
	c.Lock()
	defer c.Unlock()
	now := time.Now()
	for key, item := range c.items {
		if !item.expiration.IsZero() && item.expiration.Before(now) {
			delete(c.items, key)
		}
	}
}
