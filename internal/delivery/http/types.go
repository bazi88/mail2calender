package http

import "context"

type EmailAuthService interface {
	ExchangeCodeForToken(ctx context.Context, code string) error
}
