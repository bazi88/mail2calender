package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisRateLimiter struct {
	redisClient *redis.Client
	limit       int
	window      time.Duration
}

func NewRedisRateLimiter(redisClient *redis.Client, limit int, window time.Duration) *RedisRateLimiter {
	return &RedisRateLimiter{
		redisClient: redisClient,
		limit:       limit,
		window:      window,
	}
}

func (r *RedisRateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		key := fmt.Sprintf("rate_limit:%s:%s", req.RemoteAddr, req.URL.Path)

		count, err := r.redisClient.Incr(context.Background(), key).Result()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if count == 1 {
			r.redisClient.Expire(context.Background(), key, r.window)
		}

		if count > int64(r.limit) {
			retryAfter := int(r.window.Seconds())
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, req)
	})
}
