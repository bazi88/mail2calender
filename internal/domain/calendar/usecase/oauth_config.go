package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"mono-golang/internal/domain/calendar/logger"
)

// OAuthConfig handles OAuth2 configuration and token management
type OAuthConfig struct {
	config     *oauth2.Config
	tokenStore TokenStore
	logger     *logger.Logger
	maxRetries int
	retryDelay time.Duration
}

// TokenStore defines interface for token storage
type TokenStore interface {
	GetToken(ctx context.Context, userID string) (*oauth2.Token, error)
	SaveToken(ctx context.Context, userID string, token *oauth2.Token) error
	DeleteToken(ctx context.Context, userID string) error
}

// RedisTokenStore implements TokenStore using Redis
type RedisTokenStore struct {
	client *redis.Client
	prefix string
	ttl    time.Duration
}

// NewOAuthConfig creates new OAuth configuration
func NewOAuthConfig(l *logger.Logger) (*OAuthConfig, error) {
	clientID := getEnvOrPanic("GOOGLE_OAUTH_CLIENT_ID")
	clientSecret := getEnvOrPanic("GOOGLE_OAUTH_CLIENT_SECRET")
	redirectURL := getEnvOrPanic("GOOGLE_OAUTH_REDIRECT_URL")

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/calendar",
			"https://www.googleapis.com/auth/calendar.events",
		},
		Endpoint: google.Endpoint,
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     getEnvOrDefault("REDIS_ADDR", "localhost:6379"),
		Password: getEnvOrDefault("REDIS_PASSWORD", ""),
		DB:       0,
	})

	tokenStore := &RedisTokenStore{
		client: redisClient,
		prefix: "oauth_token:",
		ttl:    24 * time.Hour,
	}

	return &OAuthConfig{
		config:     config,
		tokenStore: tokenStore,
		logger:     l,
		maxRetries: 3,
		retryDelay: 1 * time.Second,
	}, nil
}

// GetToken retrieves token for a user with retry logic
func (oc *OAuthConfig) GetToken(ctx context.Context, userID string) (*oauth2.Token, error) {
	var token *oauth2.Token
	var err error

	for i := 0; i < oc.maxRetries; i++ {
		token, err = oc.tokenStore.GetToken(ctx, userID)
		if err == nil {
			break
		}

		if i < oc.maxRetries-1 {
			oc.logger.Warn("Failed to get token, retrying...", logger.Fields{
				"user_id": userID,
				"attempt": i + 1,
				"error":   err.Error(),
			})
			time.Sleep(oc.retryDelay)
			continue
		}

		return nil, fmt.Errorf("failed to get token after %d attempts: %v", oc.maxRetries, err)
	}

	if token.Valid() {
		return token, nil
	}

	// Token expired, try to refresh
	newToken, err := oc.config.TokenSource(ctx, token).Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %v", err)
	}

	if err := oc.tokenStore.SaveToken(ctx, userID, newToken); err != nil {
		oc.logger.Error("Failed to save refreshed token", logger.Fields{
			"user_id": userID,
			"error":   err.Error(),
		})
	}

	return newToken, nil
}

// SaveToken saves OAuth token for a user
func (oc *OAuthConfig) SaveToken(ctx context.Context, userID string, token *oauth2.Token) error {
	return oc.tokenStore.SaveToken(ctx, userID, token)
}

// DeleteToken removes OAuth token for a user
func (oc *OAuthConfig) DeleteToken(ctx context.Context, userID string) error {
	return oc.tokenStore.DeleteToken(ctx, userID)
}

// GetClient returns an HTTP client with valid OAuth token
func (oc *OAuthConfig) GetClient(ctx context.Context, userID string) (*http.Client, error) {
	token, err := oc.GetToken(ctx, userID)
	if err != nil {
		return nil, err
	}
	return oc.config.Client(ctx, token), nil
}

// RedisTokenStore implementation

func (s *RedisTokenStore) GetToken(ctx context.Context, userID string) (*oauth2.Token, error) {
	data, err := s.client.Get(ctx, s.prefix+userID).Bytes()
	if err != nil {
		return nil, err
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}

	return &token, nil
}

func (s *RedisTokenStore) SaveToken(ctx context.Context, userID string, token *oauth2.Token) error {
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}

	return s.client.Set(ctx, s.prefix+userID, data, s.ttl).Err()
}

func (s *RedisTokenStore) DeleteToken(ctx context.Context, userID string) error {
	return s.client.Del(ctx, s.prefix+userID).Err()
}

// Helper functions

func getEnvOrPanic(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("required environment variable %s is not set", key))
	}
	return value
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
