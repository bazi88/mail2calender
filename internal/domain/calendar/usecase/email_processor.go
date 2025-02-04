package usecase

import (
	"context"
	"net/mail"
	"time"
)

type EmailAttachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

type EmailMetadata struct {
	MessageID         string
	Date              time.Time
	From              *mail.Address
	To                []*mail.Address
	Cc                []*mail.Address
	ReplyTo           []*mail.Address
	InReplyTo         string
	References        []string
	ContentType       string
	ContentTransfer   string
	ContentDispostion string
}

// EmailEvent represents extracted event information from an email
type EmailEvent struct {
	Subject     string
	Description string
	StartTime   time.Time
	EndTime     time.Time
	Location    string
	Attendees   []string
	Metadata    EmailMetadata
	Attachments []EmailAttachment
}

// EmailProcessor defines the interface for processing emails into calendar events
type EmailProcessor interface {
	// ProcessEmail parses an email and extracts event information
	ProcessEmail(ctx context.Context, emailContent string) (*EmailEvent, error)

	// ValidateEmail checks if the email is from a trusted source and properly signed
	ValidateEmail(ctx context.Context, emailContent string) error
}

type EmailValidator interface {
	ValidateDKIM(email string) error
	ValidateSPF(email string) error
	ValidateSender(email string) error
}
