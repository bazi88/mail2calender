package email_auth

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type emailAuthServiceImpl struct {
	oauth2Configs map[EmailProvider]*oauth2.Config
	tokenStore    TokenStore
}

type TokenStore interface {
	SaveToken(ctx context.Context, userID string, token *oauth2.Token) error
	GetToken(ctx context.Context, userID string) (*oauth2.Token, error)
}

func NewEmailAuthService(tokenStore TokenStore, configs map[EmailProvider]OAuthConfig) EmailAuthService {
	oauth2Configs := make(map[EmailProvider]*oauth2.Config)

	// Configure Gmail
	if config, ok := configs[Gmail]; ok {
		oauth2Configs[Gmail] = &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			RedirectURL:  config.RedirectURL,
			Scopes:       config.Scopes,
			Endpoint:     google.Endpoint,
		}
	}

	// Configure Outlook (if needed)
	if config, ok := configs[Outlook]; ok {
		oauth2Configs[Outlook] = &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			RedirectURL:  config.RedirectURL,
			Scopes:       config.Scopes,
			Endpoint:     google.Endpoint, // Replace with Outlook endpoint
		}
	}

	return &emailAuthServiceImpl{
		oauth2Configs: oauth2Configs,
		tokenStore:    tokenStore,
	}
}

func (s *emailAuthServiceImpl) GetAuthURL(ctx context.Context, provider EmailProvider) (string, error) {
	config, ok := s.oauth2Configs[provider]
	if !ok {
		return "", fmt.Errorf("unsupported email provider: %s", provider)
	}
	return config.AuthCodeURL("state", oauth2.AccessTypeOffline), nil
}

func (s *emailAuthServiceImpl) ExchangeCode(ctx context.Context, provider EmailProvider, code string) (*EmailToken, error) {
	config, ok := s.oauth2Configs[provider]
	if !ok {
		return nil, fmt.Errorf("unsupported email provider: %s", provider)
	}

	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	return &EmailToken{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
		Provider:     provider,
	}, nil
}

func (s *emailAuthServiceImpl) RefreshToken(ctx context.Context, token *EmailToken) (*EmailToken, error) {
	config, ok := s.oauth2Configs[token.Provider]
	if !ok {
		return nil, fmt.Errorf("unsupported email provider: %s", token.Provider)
	}

	oauthToken := &oauth2.Token{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	}

	tokenSource := config.TokenSource(ctx, oauthToken)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	return &EmailToken{
		AccessToken:  newToken.AccessToken,
		TokenType:    newToken.TokenType,
		RefreshToken: newToken.RefreshToken,
		Expiry:       newToken.Expiry,
		Provider:     token.Provider,
	}, nil
}

func (s *emailAuthServiceImpl) RevokeToken(ctx context.Context, token *EmailToken) error {
	// Implementation depends on the provider
	// For Gmail, you would call the revoke endpoint
	// For Outlook, you would call their revoke endpoint
	return nil
}
