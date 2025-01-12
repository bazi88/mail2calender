package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestStruct struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestRedisCache(t *testing.T) {
	// Khởi tạo miniredis
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	// Tạo Redis client kết nối đến miniredis
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
		DB:   0,
	})
	defer client.Close()

	cache := NewRedisCache(client)
	ctx := context.Background()

	t.Run("Set and Get", func(t *testing.T) {
		key := "test:1"
		value := TestStruct{Name: "test", Value: 123}

		err := cache.Set(ctx, key, value, time.Minute)
		require.NoError(t, err)

		var result TestStruct
		err = cache.Get(ctx, key, &result)
		require.NoError(t, err)
		assert.Equal(t, value, result)
	})

	t.Run("Get non-existent key", func(t *testing.T) {
		var result TestStruct
		err := cache.Get(ctx, "non:existent", &result)
		require.NoError(t, err)
		assert.Equal(t, TestStruct{}, result)
	})

	t.Run("Delete", func(t *testing.T) {
		key := "test:delete"
		value := TestStruct{Name: "delete-test", Value: 456}

		err := cache.Set(ctx, key, value, time.Minute)
		require.NoError(t, err)

		err = cache.Delete(ctx, key)
		require.NoError(t, err)

		var result TestStruct
		err = cache.Get(ctx, key, &result)
		require.NoError(t, err)
		assert.Equal(t, TestStruct{}, result)
	})

	t.Run("DeletePattern", func(t *testing.T) {
		// Set nhiều keys với pattern giống nhau
		for i := 0; i < 3; i++ {
			key := "pattern:test:" + string(rune('a'+i))
			value := TestStruct{Name: "pattern-test", Value: i}
			err := cache.Set(ctx, key, value, time.Minute)
			require.NoError(t, err)
		}

		// Xóa theo pattern
		err := cache.DeletePattern(ctx, "pattern:test:*")
		require.NoError(t, err)

		// Kiểm tra xem các keys đã bị xóa chưa
		var result TestStruct
		err = cache.Get(ctx, "pattern:test:a", &result)
		require.NoError(t, err)
		assert.Equal(t, TestStruct{}, result)
	})

	t.Run("Expiration", func(t *testing.T) {
		key := "test:expire"
		value := TestStruct{Name: "expire-test", Value: 789}

		err := cache.Set(ctx, key, value, 100*time.Millisecond)
		require.NoError(t, err)

		// Fast-forward thời gian trong miniredis
		mr.FastForward(200 * time.Millisecond)

		var result TestStruct
		err = cache.Get(ctx, key, &result)
		require.NoError(t, err)
		assert.Equal(t, TestStruct{}, result)
	})
}
