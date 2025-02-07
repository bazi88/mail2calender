package usecase

import (
	"context"
	"os"
	"testing"
	"time"

	"mail2calendar/internal/domain/calendar/logger"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"
)

// Mock token store for testing
type mockTokenStore struct {
	mock.Mock
}

func (m *mockTokenStore) GetToken(ctx context.Context, userID string) (*oauth2.Token, error) {
	args := m.Called(ctx, userID)
	if token := args.Get(0); token != nil {
		return token.(*oauth2.Token), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockTokenStore) SaveToken(ctx context.Context, userID string, token *oauth2.Token) error {
	args := m.Called(ctx, userID, token)
	return args.Error(0)
}

func (m *mockTokenStore) DeleteToken(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func TestNewOAuthConfig(t *testing.T) {
	// Setup environment variables for test
	envVars := map[string]string{
		"GOOGLE_OAUTH_CLIENT_ID":     "test-client-id",
		"GOOGLE_OAUTH_CLIENT_SECRET": "test-client-secret",
		"GOOGLE_OAUTH_REDIRECT_URL":  "http://localhost:8080/callback",
	}

	for k, v := range envVars {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}

	l, _ := logger.New(nil)
	config, err := NewOAuthConfig(l)

	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "test-client-id", config.config.ClientID)
	assert.Equal(t, "test-client-secret", config.config.ClientSecret)
	assert.Equal(t, "http://localhost:8080/callback", config.config.RedirectURL)
	assert.Contains(t, config.config.Scopes, "https://www.googleapis.com/auth/calendar")
}

func TestOAuthConfig_GetToken(t *testing.T) {
	l, _ := logger.New(nil)
	mockStore := new(mockTokenStore)
	config := &OAuthConfig{
		config:     &oauth2.Config{},
		tokenStore: mockStore,
		logger:     l,
		maxRetries: 2,
		retryDelay: time.Millisecond,
	}

	tests := []struct {
		name      string
		userID    string
		setupMock func()
		wantErr   bool
	}{
		{
			name:   "successful token retrieval",
			userID: "user1",
			setupMock: func() {
				validToken := &oauth2.Token{
					AccessToken: "valid-token",
					Expiry:      time.Now().Add(time.Hour),
				}
				mockStore.On("GetToken", mock.Anything, "user1").Return(validToken, nil)
			},
			wantErr: false,
		},
		{
			name:   "token not found",
			userID: "user2",
			setupMock: func() {
				mockStore.On("GetToken", mock.Anything, "user2").Return(nil, redis.Nil)
			},
			wantErr: true,
		},
		{
			name:   "expired token with successful refresh",
			userID: "user3",
			setupMock: func() {
				expiredToken := &oauth2.Token{
					AccessToken: "expired-token",
					Expiry:      time.Now().Add(-time.Hour),
				}
				mockStore.On("GetToken", mock.Anything, "user3").Return(expiredToken, nil)
				mockStore.On("SaveToken", mock.Anything, "user3", mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "retry success",
			userID: "user4",
			setupMock: func() {
				validToken := &oauth2.Token{
					AccessToken: "valid-token",
					Expiry:      time.Now().Add(time.Hour),
				}
				mockStore.On("GetToken", mock.Anything, "user4").
					Return(nil, redis.Nil).Once().
					Return(validToken, nil).Once()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			token, err := config.GetToken(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, token)
			}

			mockStore.AssertExpectations(t)
		})
	}
}

func TestRedisTokenStore(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: os.Getenv("TEST_REDIS_ADDR"),
	})
	if os.Getenv("TEST_REDIS_ADDR") == "" {
		t.Skip("Skipping Redis tests: TEST_REDIS_ADDR not set")
	}

	store := &RedisTokenStore{
		client: redisClient,
		prefix: "test_token:",
		ttl:    time.Hour,
	}

	ctx := context.Background()
	userID := "test_user"
	token := &oauth2.Token{
		AccessToken:  "test-token",
		TokenType:    "Bearer",
		RefreshToken: "test-refresh",
		Expiry:       time.Now().Add(time.Hour),
	}

	t.Run("save and get token", func(t *testing.T) {
		// Save token
		err := store.SaveToken(ctx, userID, token)
		assert.NoError(t, err)

		// Get token
		savedToken, err := store.GetToken(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, token.AccessToken, savedToken.AccessToken)
		assert.Equal(t, token.TokenType, savedToken.TokenType)
		assert.Equal(t, token.RefreshToken, savedToken.RefreshToken)
	})

	t.Run("delete token", func(t *testing.T) {
		// Save token
		err := store.SaveToken(ctx, userID, token)
		assert.NoError(t, err)

		// Delete token
		err = store.DeleteToken(ctx, userID)
		assert.NoError(t, err)

		// Try to get deleted token
		_, err = store.GetToken(ctx, userID)
		assert.Error(t, err)
		assert.Equal(t, redis.Nil, err)
	})
}

func TestOAuthConfig_GetClient(t *testing.T) {
	l, _ := logger.New(nil)
	mockStore := new(mockTokenStore)
	config := &OAuthConfig{
		config:     &oauth2.Config{},
		tokenStore: mockStore,
		logger:     l,
		maxRetries: 2,
		retryDelay: time.Millisecond,
	}

	t.Run("successful client creation", func(t *testing.T) {
		validToken := &oauth2.Token{
			AccessToken: "valid-token",
			Expiry:      time.Now().Add(time.Hour),
		}
		mockStore.On("GetToken", mock.Anything, "user1").Return(validToken, nil)

		client, err := config.GetClient(context.Background(), "user1")
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("failed client creation", func(t *testing.T) {
		mockStore.On("GetToken", mock.Anything, "user2").Return(nil, redis.Nil)

		client, err := config.GetClient(context.Background(), "user2")
		assert.Error(t, err)
		assert.Nil(t, client)
	})
}
