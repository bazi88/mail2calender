package middleware

import (
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
		ctx := req.Context()

		// Kiểm tra kết nối Redis
		_, err := r.redisClient.Ping(ctx).Result()
		if err != nil {
			http.Error(w, "Rate limiter unavailable", http.StatusServiceUnavailable)
			return
		}

		// Tăng counter và set expire
		pipe := r.redisClient.Pipeline()
		incr := pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, r.window)

		_, err = pipe.Exec(ctx)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		count := incr.Val()
		if count > int64(r.limit) {
			retryAfter := int(r.window.Seconds())
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, req)
	})
}
