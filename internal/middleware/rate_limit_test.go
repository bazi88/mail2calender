package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

func TestRedisRateLimiter(t *testing.T) {
	// Ensure Redis is clean before starting tests
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	redisClient.FlushAll(context.Background())

	rateLimiter := NewRedisRateLimiter(redisClient, 2, time.Second)

	t.Run("Allow requests within limit", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		})

		middleware := rateLimiter.Limit(handler)

		// Make two requests - both should succeed
		for i := 0; i < 2; i++ {
			req := httptest.NewRequest(http.MethodGet, "/test-allow", nil)
			rec := httptest.NewRecorder()
			middleware.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "success", rec.Body.String())
		}
	})

	t.Run("Block requests over limit", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		})

		middleware := rateLimiter.Limit(handler)

		// Make three requests - third one should be blocked
		for i := 0; i < 3; i++ {
			req := httptest.NewRequest(http.MethodGet, "/test-block", nil)
			rec := httptest.NewRecorder()

			middleware.ServeHTTP(rec, req)

			if i < 2 {
				// First two requests should succeed
				assert.Equal(t, http.StatusOK, rec.Code, "Request %d should succeed", i+1)
				assert.Equal(t, "success", rec.Body.String())
			} else {
				// Third request should be blocked
				assert.Equal(t, http.StatusTooManyRequests, rec.Code, "Request %d should be blocked", i+1)
				assert.Contains(t, rec.Body.String(), "Too Many Requests")
				assert.NotEmpty(t, rec.Header().Get("Retry-After"))
			}
		}
	})

	t.Run("Reset limit after window", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		})

		middleware := rateLimiter.Limit(handler)

		// First request
		req := httptest.NewRequest(http.MethodGet, "/test-reset", nil)
		rec := httptest.NewRecorder()
		middleware.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		// Wait for the rate limit window to expire
		time.Sleep(time.Second)

		// Should be allowed after window expires
		req = httptest.NewRequest(http.MethodGet, "/test-reset", nil)
		rec = httptest.NewRecorder()
		middleware.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	// Clean up Redis after all tests
	redisClient.FlushAll(context.Background())
	redisClient.Close()
}
