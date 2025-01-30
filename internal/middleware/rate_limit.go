package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
)

// RateLimiter định nghĩa interface cho rate limiting
type RateLimiter interface {
	Limit(next http.Handler) http.Handler
}

type RedisRateLimiter struct {
	redisClient *redis.Client
	maxRequests int
	duration    time.Duration
	keyPrefix   string
}

// NewRedisRateLimiter tạo một rate limiter mới
func NewRedisRateLimiter(redisClient *redis.Client, maxRequests int, duration time.Duration, keyPrefix string) *RedisRateLimiter {
	return &RedisRateLimiter{
		redisClient: redisClient,
		maxRequests: maxRequests,
		duration:    duration,
		keyPrefix:   keyPrefix,
	}
}

// Limit middleware để giới hạn số lượng request
func (rl *RedisRateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Tạo key dựa trên IP hoặc user ID (nếu đã đăng nhập)
		identifier := r.RemoteAddr
		if userID := GetUserIDFromContext(r.Context()); userID != "" {
			identifier = userID
		}
		key := fmt.Sprintf("%s:%s", rl.keyPrefix, identifier)

		// Kiểm tra và tăng số lượng request
		count, err := rl.redisClient.Incr(ctx, key).Result()
		if err != nil {
			http.Error(w, "Rate limit error", http.StatusInternalServerError)
			return
		}

		// Nếu là request đầu tiên, set thời gian hết hạn
		if count == 1 {
			rl.redisClient.Expire(ctx, key, rl.duration)
		}

		// Thêm headers để client biết giới hạn
		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.maxRequests))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", rl.maxRequests-int(count)))

		// Kiểm tra nếu vượt quá giới hạn
		if count > int64(rl.maxRequests) {
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(rl.duration).Unix()))
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetUserIDFromContext lấy user ID từ context nếu có
func GetUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}
	return ""
}

// DefaultRateLimiter tạo một rate limiter với cấu hình mặc định
func DefaultRateLimiter(redisClient *redis.Client) *RedisRateLimiter {
	return NewRedisRateLimiter(
		redisClient,
		100,              // 100 requests
		time.Minute,      // per minute
		"rate_limit:api", // prefix cho Redis key
	)
}

// RateLimitByPath tạo một rate limiter cho một path cụ thể
func RateLimitByPath(redisClient *redis.Client, path string, maxRequests int, duration time.Duration) *RedisRateLimiter {
	return NewRedisRateLimiter(
		redisClient,
		maxRequests,
		duration,
		fmt.Sprintf("rate_limit:%s", path),
	)
}
