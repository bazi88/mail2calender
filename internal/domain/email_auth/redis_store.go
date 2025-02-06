package email_auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/oauth2"
)

type RedisTokenStore struct {
	redisClient *redis.Client
	keyPrefix   string
	tokenTTL    time.Duration
}

func NewRedisTokenStore(redisClient *redis.Client) *RedisTokenStore {
	return &RedisTokenStore{
		redisClient: redisClient,
		keyPrefix:   "oauth2_token:",
		tokenTTL:    24 * time.Hour * 30, // 30 days
	}
}

func (s *RedisTokenStore) tokenKey(userID string) string {
	return s.keyPrefix + userID
}

func (s *RedisTokenStore) SaveToken(ctx context.Context, userID string, token *oauth2.Token) error {
	tokenBytes, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %v", err)
	}

	key := s.tokenKey(userID)
	if err := s.redisClient.Set(ctx, key, tokenBytes, s.tokenTTL).Err(); err != nil {
		return fmt.Errorf("failed to save token to redis: %v", err)
	}

	return nil
}

func (s *RedisTokenStore) GetToken(ctx context.Context, userID string) (*oauth2.Token, error) {
	key := s.tokenKey(userID)
	tokenBytes, err := s.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("token not found for user %s", userID)
		}
		return nil, fmt.Errorf("failed to get token from redis: %v", err)
	}

	var token oauth2.Token
	if err := json.Unmarshal(tokenBytes, &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %v", err)
	}

	return &token, nil
}
