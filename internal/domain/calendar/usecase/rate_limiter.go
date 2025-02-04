package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// RateLimiter handles API rate limiting using Redis
type RateLimiter interface {
	// Allow checks if the request is allowed based on rate limits
	Allow(ctx context.Context, userID string) (bool, error)
	// GetRemainingQuota returns the number of requests remaining for the user
	GetRemainingQuota(ctx context.Context, userID string) (int64, error)
}

type RateLimiterConfig struct {
	RequestsPerHour int64
	BurstSize       int64
	RedisKeyPrefix  string
}

type rateLimiterImpl struct {
	redis  *redis.Client
	config RateLimiterConfig
	tracer trace.Tracer
}

// NewRateLimiter creates a new instance of RateLimiter
func NewRateLimiter(redisClient *redis.Client, config RateLimiterConfig) RateLimiter {
	return &rateLimiterImpl{
		redis:  redisClient,
		config: config,
		tracer: otel.Tracer("rate-limiter"),
	}
}

func (r *rateLimiterImpl) Allow(ctx context.Context, userID string) (bool, error) {
	ctx, span := r.tracer.Start(ctx, "RateLimiter.Allow")
	defer span.End()

	span.SetAttributes(attribute.String("user_id", userID))

	key := fmt.Sprintf("%s:%s", r.config.RedisKeyPrefix, userID)
	hourKey := fmt.Sprintf("%s:hour", key)
	burstKey := fmt.Sprintf("%s:burst", key)

	// Start a Redis transaction
	pipe := r.redis.Pipeline()

	// Check hourly limit
	hourlyCount := pipe.Incr(ctx, hourKey)
	pipe.Expire(ctx, hourKey, time.Hour)

	// Check burst limit
	burstCount := pipe.Incr(ctx, burstKey)
	pipe.Expire(ctx, burstKey, time.Minute)

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		span.RecordError(err)
		return false, fmt.Errorf("failed to execute Redis pipeline: %v", err)
	}

	// Get results
	hourlyVal := hourlyCount.Val()
	burstVal := burstCount.Val()

	span.SetAttributes(
		attribute.Int64("hourly_count", hourlyVal),
		attribute.Int64("burst_count", burstVal),
	)

	// Check if either limit is exceeded
	if hourlyVal > r.config.RequestsPerHour {
		return false, nil
	}

	if burstVal > r.config.BurstSize {
		return false, nil
	}

	return true, nil
}

func (r *rateLimiterImpl) GetRemainingQuota(ctx context.Context, userID string) (int64, error) {
	ctx, span := r.tracer.Start(ctx, "RateLimiter.GetRemainingQuota")
	defer span.End()

	span.SetAttributes(attribute.String("user_id", userID))

	key := fmt.Sprintf("%s:%s:hour", r.config.RedisKeyPrefix, userID)

	// Get current count
	val, err := r.redis.Get(ctx, key).Int64()
	if err == redis.Nil {
		// No requests made yet
		return r.config.RequestsPerHour, nil
	}
	if err != nil {
		span.RecordError(err)
		return 0, fmt.Errorf("failed to get quota from Redis: %v", err)
	}

	remaining := r.config.RequestsPerHour - val
	if remaining < 0 {
		remaining = 0
	}

	span.SetAttributes(attribute.Int64("remaining_quota", remaining))
	return remaining, nil
}

// Helper function to create Redis key for rate limiting
func (r *rateLimiterImpl) getRedisKey(userID string) string {
	return fmt.Sprintf("%s:%s", r.config.RedisKeyPrefix, userID)
}

// Example usage in CalendarService:
/*
type CalendarService struct {
    emailProcessor EmailProcessor
    calendar      GoogleCalendarService
    rateLimiter   RateLimiter
    tracer        trace.Tracer
}

func (s *CalendarService) ProcessEmailToCalendar(ctx context.Context, emailContent string, userID string) error {
    allowed, err := s.rateLimiter.Allow(ctx, userID)
    if err != nil {
        return fmt.Errorf("rate limiter error: %v", err)
    }
    if !allowed {
        return fmt.Errorf("rate limit exceeded for user %s", userID)
    }

    // Continue with normal processing...
}
*/
