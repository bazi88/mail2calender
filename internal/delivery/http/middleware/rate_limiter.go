package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisRateLimiter xử lý giới hạn request sử dụng Redis
type RedisRateLimiter struct {
	client *redis.Client
	limit  int
	window time.Duration
}

// NewRedisRateLimiter tạo một rate limiter mới
func NewRedisRateLimiter(client *redis.Client, limit int, window time.Duration) *RedisRateLimiter {
	return &RedisRateLimiter{
		client: client,
		limit:  limit,
		window: window,
	}
}

// Limit là middleware để giới hạn số lượng request
func (r *RedisRateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Lấy IP của client làm key
		key := fmt.Sprintf("rate_limit:%s", req.RemoteAddr)

		// Kiểm tra số lượng request trong window
		val, err := r.client.Get(req.Context(), key).Int()
		if err != nil && err != redis.Nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Nếu chưa có key, tạo mới với TTL là window
		if err == redis.Nil {
			err = r.client.Set(req.Context(), key, 1, r.window).Err()
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			next.ServeHTTP(w, req)
			return
		}

		// Nếu đã vượt quá limit
		if val >= r.limit {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		// Tăng counter
		err = r.client.Incr(req.Context(), key).Err()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		next.ServeHTTP(w, req)
	})
}
