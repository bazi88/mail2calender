// Package http cung cấp các handlers và types cho HTTP delivery layer
package http

import "context"

type EmailAuthService interface {
	ExchangeCodeForToken(ctx context.Context, code string) error
}
