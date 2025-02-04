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

func setupTestRedis(t *testing.T) (*redis.Client, func()) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return client, func() {
		client.Close()
		mr.Close()
	}
}

func TestRateLimiter_Allow(t *testing.T) {
	redisClient, cleanup := setupTestRedis(t)
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
		setup     func()
		userID    string
		wantAllow bool
		wantErr   bool
	}{
		{
			name:      "first request should be allowed",
			setup:     func() {},
			userID:    "user1",
			wantAllow: true,
			wantErr:   false,
		},
		{
			name: "should respect hourly limit",
			setup: func() {
				// Simulate 10 previous requests
				for i := 0; i < 10; i++ {
					allowed, err := limiter.Allow(ctx, "user2")
					require.NoError(t, err)
					require.True(t, allowed)
				}
			},
			userID:    "user2",
			wantAllow: false,
			wantErr:   false,
		},
		{
			name: "should respect burst limit",
			setup: func() {
				// Simulate 3 quick requests
				for i := 0; i < 3; i++ {
					allowed, err := limiter.Allow(ctx, "user3")
					require.NoError(t, err)
					require.True(t, allowed)
				}
			},
			userID:    "user3",
			wantAllow: false,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

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
	redisClient, cleanup := setupTestRedis(t)
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
		setup         func()
		userID        string
		wantRemaining int64
		wantErr       bool
	}{
		{
			name:          "new user should have full quota",
			setup:         func() {},
			userID:        "user1",
			wantRemaining: 10,
			wantErr:       false,
		},
		{
			name: "should return correct remaining quota",
			setup: func() {
				// Make 3 requests
				for i := 0; i < 3; i++ {
					_, err := limiter.Allow(ctx, "user2")
					require.NoError(t, err)
				}
			},
			userID:        "user2",
			wantRemaining: 7,
			wantErr:       false,
		},
		{
			name: "should return zero when quota exceeded",
			setup: func() {
				// Make 10 requests
				for i := 0; i < 10; i++ {
					_, err := limiter.Allow(ctx, "user3")
					require.NoError(t, err)
				}
			},
			userID:        "user3",
			wantRemaining: 0,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

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
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer redisClient.Close()

	config := RateLimiterConfig{
		RequestsPerHour: 10,
		BurstSize:       3,
		RedisKeyPrefix:  "test",
	}

	limiter := NewRateLimiter(redisClient, config)
	ctx := context.Background()
	userID := "user1"

	// Make some requests
	for i := 0; i < 5; i++ {
		_, err := limiter.Allow(ctx, userID)
		require.NoError(t, err)
	}

	// Check quota
	remaining, err := limiter.GetRemainingQuota(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, int64(5), remaining)

	// Fast forward time
	mr.FastForward(time.Hour)

	// Check quota again - should be reset
	remaining, err = limiter.GetRemainingQuota(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, int64(10), remaining)
}
