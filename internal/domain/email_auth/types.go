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

// OAuthConfig contains OAuth configuration for an email provider
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

// EmailToken represents an OAuth token for email access
type EmailToken struct {
	AccessToken  string
	TokenType    string
	RefreshToken string
	Expiry       time.Time
	Provider     EmailProvider
}

// EmailAuthService defines the interface for email authentication operations
type EmailAuthService interface {
	GetAuthURL(ctx context.Context, provider EmailProvider) (string, error)
	ExchangeCode(ctx context.Context, provider EmailProvider, code string) (*EmailToken, error)
	RefreshToken(ctx context.Context, token *EmailToken) (*EmailToken, error)
	RevokeToken(ctx context.Context, token *EmailToken) error
}
