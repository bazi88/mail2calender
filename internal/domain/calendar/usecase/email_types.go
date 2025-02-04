package usecase

import (
	"net/mail"
	"time"
)

// EmailMetadata represents metadata from email headers
type EmailMetadata struct {
	MessageID         string
	From              *mail.Address
	To                []*mail.Address
	Cc                []*mail.Address
	ReplyTo           []*mail.Address
	Date              time.Time
	References        []string
	InReplyTo         string
	ContentType       string
	ContentTransfer   string
	ContentDispostion string
}

// EmailAttachment represents an email attachment
type EmailAttachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

// EmailEvent represents a calendar event extracted from an email
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
