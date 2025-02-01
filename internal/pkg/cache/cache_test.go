package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {
	cache := &Cache{
		items: make(map[string]cacheItem),
	}
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
		assert.ErrorIs(t, err, ErrKeyNotFound)
	})

	t.Run("Expiration", func(t *testing.T) {
		err := cache.Set(ctx, "key3", "value3", time.Millisecond)
		assert.NoError(t, err)

		time.Sleep(time.Millisecond * 2)
		_, err = cache.Get(ctx, "key3")
		assert.ErrorIs(t, err, ErrKeyNotFound)
	})
}

func TestNewWithCleanupInterval(t *testing.T) {
	cache := NewWithCleanupInterval(time.Millisecond)
	err := cache.Set(context.Background(), "key", "value", time.Millisecond)
	assert.NoError(t, err)

	time.Sleep(time.Millisecond * 3) // Wait for cleanup
	_, err = cache.Get(context.Background(), "key")
	assert.ErrorIs(t, err, ErrKeyNotFound)
}
