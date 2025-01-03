package cache

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestRedisCache(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})
	defer client.Close()

	cache := NewRedisCache(client)
	ctx := context.Background()

	// Xóa tất cả keys trước khi test
	client.FlushDB(ctx)

	t.Run("Set and Get", func(t *testing.T) {
		key := "test:1"
		value := TestStruct{Name: "test", Value: 123}

		err := cache.Set(ctx, key, value, time.Minute)
		assert.NoError(t, err)

		var result TestStruct
		err = cache.Get(ctx, key, &result)
		assert.NoError(t, err)
		assert.Equal(t, value, result)
	})

	t.Run("Get non-existent key", func(t *testing.T) {
		var result TestStruct
		err := cache.Get(ctx, "non:existent", &result)
		assert.NoError(t, err)
		assert.Equal(t, TestStruct{}, result)
	})

	t.Run("Delete", func(t *testing.T) {
		key := "test:delete"
		value := TestStruct{Name: "delete-test", Value: 456}

		err := cache.Set(ctx, key, value, time.Minute)
		assert.NoError(t, err)

		err = cache.Delete(ctx, key)
		assert.NoError(t, err)

		var result TestStruct
		err = cache.Get(ctx, key, &result)
		assert.NoError(t, err)
		assert.Equal(t, TestStruct{}, result)
	})

	t.Run("DeletePattern", func(t *testing.T) {
		// Set nhiều keys với pattern giống nhau
		for i := 0; i < 3; i++ {
			key := "pattern:test:" + string(rune('a'+i))
			value := TestStruct{Name: "pattern-test", Value: i}
			err := cache.Set(ctx, key, value, time.Minute)
			assert.NoError(t, err)
		}

		// Xóa theo pattern
		err := cache.DeletePattern(ctx, "pattern:test:*")
		assert.NoError(t, err)

		// Kiểm tra xem các keys đã bị xóa chưa
		var result TestStruct
		err = cache.Get(ctx, "pattern:test:a", &result)
		assert.NoError(t, err)
		assert.Equal(t, TestStruct{}, result)
	})

	t.Run("Expiration", func(t *testing.T) {
		key := "test:expire"
		value := TestStruct{Name: "expire-test", Value: 789}

		err := cache.Set(ctx, key, value, 100*time.Millisecond)
		assert.NoError(t, err)

		// Đợi cho key hết hạn
		time.Sleep(200 * time.Millisecond)

		var result TestStruct
		err = cache.Get(ctx, key, &result)
		assert.NoError(t, err)
		assert.Equal(t, TestStruct{}, result)
	})
}
