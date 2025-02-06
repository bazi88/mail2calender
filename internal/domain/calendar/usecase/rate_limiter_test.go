package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis, func()) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return client, mr, func() {
		client.Close()
		mr.Close()
	}
}

func TestRateLimiter_Allow(t *testing.T) {
	redisClient, mr, cleanup := setupTestRedis(t)
	defer cleanup()

	config := RateLimiterConfig{
		RequestsPerHour: 10,
		BurstSize:       3,
		RedisKeyPrefix:  "test",
	}

	limiter := NewRateLimiter(redisClient, config)
	ctx := context.Background()

	tests := []struct {
		name      string
		setup     func(mr *miniredis.Miniredis)
		userID    string
		wantAllow bool
		wantErr   bool
	}{
		{
			name:      "first request should be allowed",
			setup:     func(mr *miniredis.Miniredis) {},
			userID:    "user1",
			wantAllow: true,
			wantErr:   false,
		},
		{
			name: "should respect hourly limit",
			setup: func(mr *miniredis.Miniredis) {
				// Giả lập đã có 10 request trong giờ
				if err := mr.Set("test:user2:hour", "10"); err != nil {
					t.Errorf("failed to set rate limit: %v", err)
				}
				mr.SetTTL("test:user2:hour", time.Hour)
			},
			userID:    "user2",
			wantAllow: false,
			wantErr:   false,
		},
		{
			name: "should respect burst limit",
			setup: func(mr *miniredis.Miniredis) {
				// Giả lập đã có 3 request trong phút
				if err := mr.Set("test:user3:burst", "3"); err != nil {
					t.Errorf("failed to set rate limit: %v", err)
				}
				mr.SetTTL("test:user3:burst", time.Minute)
			},
			userID:    "user3",
			wantAllow: false,
			wantErr:   false,
		},
		{
			name: "should handle redis error",
			setup: func(mr *miniredis.Miniredis) {
				mr.SetError("simulated error")
			},
			userID:    "user4",
			wantAllow: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset Redis state
			mr.FlushAll()
			tt.setup(mr)

			allowed, err := limiter.Allow(ctx, tt.userID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantAllow, allowed)
		})
	}
}

func TestRateLimiter_GetRemainingQuota(t *testing.T) {
	redisClient, mr, cleanup := setupTestRedis(t)
	defer cleanup()

	config := RateLimiterConfig{
		RequestsPerHour: 10,
		BurstSize:       3,
		RedisKeyPrefix:  "test",
	}

	limiter := NewRateLimiter(redisClient, config)
	ctx := context.Background()

	tests := []struct {
		name          string
		setup         func(mr *miniredis.Miniredis)
		userID        string
		wantRemaining int64
		wantErr       bool
	}{
		{
			name:          "new user should have full quota",
			setup:         func(mr *miniredis.Miniredis) {},
			userID:        "user1",
			wantRemaining: 10,
			wantErr:       false,
		},
		{
			name: "should return correct remaining quota",
			setup: func(mr *miniredis.Miniredis) {
				// Giả lập đã sử dụng 3 request
				if err := mr.Set("test:user2:hour", "3"); err != nil {
					t.Errorf("failed to set rate limit: %v", err)
				}
				mr.SetTTL("test:user2:hour", time.Hour)
			},
			userID:        "user2",
			wantRemaining: 7,
			wantErr:       false,
		},
		{
			name: "should return zero when quota exceeded",
			setup: func(mr *miniredis.Miniredis) {
				// Giả lập đã sử dụng hết quota
				if err := mr.Set("test:user3:hour", "10"); err != nil {
					t.Errorf("failed to set rate limit: %v", err)
				}
				mr.SetTTL("test:user3:hour", time.Hour)
			},
			userID:        "user3",
			wantRemaining: 0,
			wantErr:       false,
		},
		{
			name: "should handle redis error",
			setup: func(mr *miniredis.Miniredis) {
				mr.SetError("simulated error")
			},
			userID:        "user4",
			wantRemaining: 0,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset Redis state
			mr.FlushAll()
			tt.setup(mr)

			remaining, err := limiter.GetRemainingQuota(ctx, tt.userID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantRemaining, remaining)
		})
	}
}

func TestRateLimiter_Expiration(t *testing.T) {
	redisClient, mr, cleanup := setupTestRedis(t)
	defer cleanup()

	config := RateLimiterConfig{
		RequestsPerHour: 10,
		BurstSize:       3,
		RedisKeyPrefix:  "test",
	}

	limiter := NewRateLimiter(redisClient, config)
	ctx := context.Background()
	userID := "user1"

	// Giả lập đã sử dụng 5 request
	if err := mr.Set("test:user1:hour", "5"); err != nil {
		t.Errorf("failed to set rate limit: %v", err)
	}
	mr.SetTTL("test:user1:hour", time.Hour)

	// Kiểm tra quota
	remaining, err := limiter.GetRemainingQuota(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, int64(5), remaining)

	// Fast forward time
	mr.FastForward(time.Hour)

	// Kiểm tra quota lại - phải được reset
	remaining, err = limiter.GetRemainingQuota(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, int64(10), remaining)
}
