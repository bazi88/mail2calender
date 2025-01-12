package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimiter(t *testing.T) {
	// Khởi tạo miniredis
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	// Tạo Redis client kết nối đến miniredis
	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
		DB:   0,
	})
	defer redisClient.Close()

	// Tạo rate limiter cho test (3 requests/second)
	rateLimiter := NewRedisRateLimiter(redisClient, 3, time.Second, "test")

	// Handler đơn giản để test
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test case 1: Trong giới hạn
	t.Run("Within limit", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			rr := httptest.NewRecorder()

			rateLimiter.RateLimit(handler).ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
			assert.Equal(t, "3", rr.Header().Get("X-RateLimit-Limit"))
			remaining := fmt.Sprintf("%d", 2-i)
			assert.Equal(t, remaining, rr.Header().Get("X-RateLimit-Remaining"))
		}
	})

	// Test case 2: Vượt quá giới hạn
	t.Run("Exceeds limit", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		rateLimiter.RateLimit(handler).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusTooManyRequests, rr.Code)
		assert.Equal(t, "3", rr.Header().Get("X-RateLimit-Limit"))
		assert.Equal(t, "-1", rr.Header().Get("X-RateLimit-Remaining"))
		assert.NotEmpty(t, rr.Header().Get("X-RateLimit-Reset"))
	})

	// Test case 3: Reset sau khi hết thời gian
	t.Run("Reset after duration", func(t *testing.T) {
		// Fast-forward thời gian trong miniredis
		mr.FastForward(time.Second)

		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		rateLimiter.RateLimit(handler).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "3", rr.Header().Get("X-RateLimit-Limit"))
		assert.Equal(t, "2", rr.Header().Get("X-RateLimit-Remaining"))
	})
}
