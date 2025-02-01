package cache

import (
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
