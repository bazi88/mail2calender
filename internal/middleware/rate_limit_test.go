package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

func TestRedisRateLimiter(t *testing.T) {
	// Setup mock Redis server
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	// Create Redis client connected to mock server
	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
		DB:   0,
	})
	defer redisClient.Close()

	limiter := NewRedisRateLimiter(redisClient, 2, time.Second)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("Allow requests within limit", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		limiter.Limit(handler).ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)

		rr = httptest.NewRecorder()
		limiter.Limit(handler).ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Block requests over limit", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		limiter.Limit(handler).ServeHTTP(rr, req)
		assert.Equal(t, http.StatusTooManyRequests, rr.Code)
		assert.NotEmpty(t, rr.Header().Get("Retry-After"))
	})

	t.Run("Reset limit after window", func(t *testing.T) {
		mr.FastForward(time.Second)

		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		limiter.Limit(handler).ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestRedisRateLimiterErrors(t *testing.T) {
	// Setup Redis client with wrong address to simulate errors
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6380", // Wrong port
		DB:   0,
	})
	defer redisClient.Close()

	limiter := NewRedisRateLimiter(redisClient, 2, time.Second)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("Redis connection error", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		limiter.Limit(handler).ServeHTTP(rr, req)
		assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
	})
}

func TestSetSecurityHeaders(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler := SetSecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(w, req)

	res := w.Result()
	if res.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("Expected nosniff, got %s", res.Header.Get("X-Content-Type-Options"))
	}
	// Thêm các kiểm tra khác cho các tiêu đề bảo mật
}
