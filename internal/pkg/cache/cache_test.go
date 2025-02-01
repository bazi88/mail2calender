package cache

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type MemoryCache struct {
	data sync.Map
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{}
}

func (c *MemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	c.data.Store(key, value)
	return nil
}

func (c *MemoryCache) Get(ctx context.Context, key string) (interface{}, error) {
	if value, ok := c.data.Load(key); ok {
		return value, nil
	}
	return nil, ErrKeyNotFound
}

func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.data.Delete(key)
	return nil
}

var ErrKeyNotFound = errors.New("key not found")

func TestCache(t *testing.T) {
	cache := NewMemoryCache()
	ctx := context.Background()

	t.Run("Set and Get", func(t *testing.T) {
		err := cache.Set(ctx, "key1", "value1", time.Minute)
		assert.NoError(t, err)

		val, err := cache.Get(ctx, "key1")
		assert.NoError(t, err)
		assert.Equal(t, "value1", val)
	})

	t.Run("Delete", func(t *testing.T) {
		err := cache.Set(ctx, "key2", "value2", time.Minute)
		assert.NoError(t, err)

		err = cache.Delete(ctx, "key2")
		assert.NoError(t, err)

		_, err = cache.Get(ctx, "key2")
		assert.Error(t, err)
	})
}
