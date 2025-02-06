package email_auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type EmailAuthService struct {
	oauth2Config *oauth2.Config
	tokenStore   TokenStore
}

type TokenStore interface {
	SaveToken(ctx context.Context, userID string, token *oauth2.Token) error
	GetToken(ctx context.Context, userID string) (*oauth2.Token, error)
}

func NewEmailAuthService(clientID, clientSecret string, tokenStore TokenStore) *EmailAuthService {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost:8080/auth/callback",
		Scopes: []string{
			"https://www.googleapis.com/auth/gmail.readonly",
			"https://www.googleapis.com/auth/calendar.events",
		},
		Endpoint: google.Endpoint,
	}

	return &EmailAuthService{
		oauth2Config: config,
		tokenStore:   tokenStore,
	}
}

func (s *EmailAuthService) GetAuthURL() string {
	return s.oauth2Config.AuthCodeURL("state", oauth2.AccessTypeOffline)
}

func (s *EmailAuthService) HandleCallback(ctx context.Context, code string, userID string) error {
	token, err := s.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return fmt.Errorf("failed to exchange code for token: %v", err)
	}

	return s.tokenStore.SaveToken(ctx, userID, token)
}

func (s *EmailAuthService) GetClient(ctx context.Context, userID string) (*http.Client, error) {
	token, err := s.tokenStore.GetToken(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %v", err)
	}

	if token.Expiry.Before(time.Now()) {
		ts := s.oauth2Config.TokenSource(ctx, token)
		newToken, err := ts.Token()
		if err != nil {
			return nil, fmt.Errorf("failed to refresh token: %v", err)
		}
		if err := s.tokenStore.SaveToken(ctx, userID, newToken); err != nil {
			return nil, fmt.Errorf("failed to save refreshed token: %v", err)
		}
		token = newToken
	}

	return s.oauth2Config.Client(ctx, token), nil
}
