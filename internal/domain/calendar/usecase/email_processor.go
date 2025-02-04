package usecase

import (
	"context"
)

// EmailProcessor defines interface for processing emails into calendar events
type EmailProcessor interface {
	ProcessEmail(ctx context.Context, emailContent string) (*EmailEvent, error)
	ValidateEmail(ctx context.Context, emailContent string) error
}

// EmailValidator defines interface for email validation
type EmailValidator interface {
	ValidateDKIM(email string) error
	ValidateSPF(email string) error
	ValidateSender(email string) error
}
