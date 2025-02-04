package usecase

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// emailProcessorImpl implements EmailProcessor interface with monitoring
type emailProcessorImpl struct {
	tracer    trace.Tracer
	validator EmailValidator
}

// EmailValidator handles email authentication and security checks
type EmailValidator interface {
	ValidateDKIM(email string) error
	ValidateSPF(email string) error
	ValidateSender(email string) error
}

// NewEmailProcessor creates a new instance of EmailProcessor with monitoring
func NewEmailProcessorImpl(validator EmailValidator) EmailProcessor {
	return &emailProcessorImpl{
		tracer:    otel.Tracer("email-processor"),
		validator: validator,
	}
}

func (ep *emailProcessorImpl) ProcessEmail(ctx context.Context, emailContent string) (*EmailEvent, error) {
	ctx, span := ep.tracer.Start(ctx, "ProcessEmail")
	defer span.End()

	// Parse email
	msg, err := ep.parseEmail(ctx, emailContent)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to parse email: %v", err)
	}

	// Extract event information using NLP
	event, err := ep.extractEventInfo(ctx, msg)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to extract event info: %v", err)
	}

	// Validate event data
	if err := ep.validateEvent(ctx, event); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("invalid event data: %v", err)
	}

	span.SetAttributes(
		attribute.String("event.subject", event.Subject),
		attribute.String("event.start_time", event.StartTime.Format(time.RFC3339)),
		attribute.Int("event.attendees_count", len(event.Attendees)),
	)

	return event, nil
}

func (ep *emailProcessorImpl) ValidateEmail(ctx context.Context, emailContent string) error {
	ctx, span := ep.tracer.Start(ctx, "ValidateEmail")
	defer span.End()

	if err := ep.validator.ValidateDKIM(emailContent); err != nil {
		span.RecordError(err)
		return fmt.Errorf("DKIM validation failed: %v", err)
	}

	if err := ep.validator.ValidateSPF(emailContent); err != nil {
		span.RecordError(err)
		return fmt.Errorf("SPF validation failed: %v", err)
	}

	if err := ep.validator.ValidateSender(emailContent); err != nil {
		span.RecordError(err)
		return fmt.Errorf("sender validation failed: %v", err)
	}

	return nil
}

func (ep *emailProcessorImpl) parseEmail(ctx context.Context, emailContent string) (*mail.Message, error) {
	ctx, span := ep.tracer.Start(ctx, "parseEmail")
	defer span.End()

	msg, err := mail.ReadMessage(strings.NewReader(emailContent))
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return msg, nil
}

func (ep *emailProcessorImpl) extractEventInfo(ctx context.Context, msg *mail.Message) (*EmailEvent, error) {
	ctx, span := ep.tracer.Start(ctx, "extractEventInfo")
	defer span.End()

	// TODO: Implement NLP-based extraction
	// For now, using basic parsing
	subject := msg.Header.Get("Subject")

	// Parse email body
	body := "" // TODO: Implement body parsing considering MIME types

	// Extract potential dates from subject and body
	startTime, endTime, err := ep.extractDates(ctx, subject, body)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Extract attendees from To and CC fields
	attendees := ep.extractAttendees(msg.Header)

	return &EmailEvent{
		Subject:     subject,
		Description: body,
		StartTime:   startTime,
		EndTime:     endTime,
		Location:    "", // TODO: Extract location using NLP
		Attendees:   attendees,
	}, nil
}

func (ep *emailProcessorImpl) extractDates(ctx context.Context, subject, body string) (time.Time, time.Time, error) {
	// TODO: Implement proper date extraction using NLP
	// For now, returning placeholder dates
	now := time.Now()
	return now, now.Add(time.Hour), nil
}

func (ep *emailProcessorImpl) extractAttendees(header mail.Header) []string {
	attendees := make(map[string]struct{})

	// Extract from To field
	if to := header.Get("To"); to != "" {
		addresses, err := mail.ParseAddressList(to)
		if err == nil {
			for _, addr := range addresses {
				attendees[addr.Address] = struct{}{}
			}
		}
	}

	// Extract from Cc field
	if cc := header.Get("Cc"); cc != "" {
		addresses, err := mail.ParseAddressList(cc)
		if err == nil {
			for _, addr := range addresses {
				attendees[addr.Address] = struct{}{}
			}
		}
	}

	// Convert to slice
	result := make([]string, 0, len(attendees))
	for addr := range attendees {
		result = append(result, addr)
	}

	return result
}

func (ep *emailProcessorImpl) validateEvent(ctx context.Context, event *EmailEvent) error {
	if event.Subject == "" {
		return fmt.Errorf("event subject is required")
	}

	if event.StartTime.IsZero() {
		return fmt.Errorf("event start time is required")
	}

	if event.EndTime.IsZero() {
		return fmt.Errorf("event end time is required")
	}

	if event.EndTime.Before(event.StartTime) {
		return fmt.Errorf("event end time cannot be before start time")
	}

	return nil
}
