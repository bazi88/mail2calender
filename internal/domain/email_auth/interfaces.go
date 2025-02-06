package email_auth

import (
	"context"
	"time"
)

// EmailProvider represents supported email providers
type EmailProvider string

const (
	Gmail   EmailProvider = "gmail"
	Outlook EmailProvider = "outlook"
)

// OAuthConfig contains OAuth2 configuration
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

// EmailToken represents OAuth2 token information
type EmailToken struct {
	AccessToken  string
	TokenType    string
	RefreshToken string
	Expiry       time.Time
	Provider     EmailProvider
}

// EmailAuthService defines the interface for email authentication
type EmailAuthService interface {
	// GetAuthURL generates the OAuth2 authorization URL for the specified provider
	GetAuthURL(ctx context.Context, provider EmailProvider) (string, error)

	// ExchangeCode exchanges the authorization code for access token
	ExchangeCode(ctx context.Context, provider EmailProvider, code string) (*EmailToken, error)

	// RefreshToken refreshes the access token using refresh token
	RefreshToken(ctx context.Context, token *EmailToken) (*EmailToken, error)

	// RevokeToken revokes the access token
	RevokeToken(ctx context.Context, token *EmailToken) error
}
