package usecase

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/mail"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type EmailContent struct {
	PlainText   string
	HTML        string
	RichText    string
	Attachments []EmailAttachment
	Metadata    EmailMetadata
	Links       []string
}

// emailProcessorImpl implements EmailProcessor interface with monitoring
type emailProcessorImpl struct {
	tracer     trace.Tracer
	validator  EmailValidator
	nerService NERService
}

// NewEmailProcessorImpl creates a new instance of EmailProcessor with monitoring
func NewEmailProcessorImpl(validator EmailValidator, nerService NERService) EmailProcessor {
	return &emailProcessorImpl{
		tracer:     otel.Tracer("email-processor"),
		validator:  validator,
		nerService: nerService,
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

	// Extract full email content with attachments
	content, err := ep.extractEmailContent(ctx, msg)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to extract email content: %v", err)
	}

	// Extract event information using NLP
	event, err := ep.extractEventInfo(ctx, msg, content)
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

func (ep *emailProcessorImpl) extractEmailContent(ctx context.Context, msg *mail.Message) (*EmailContent, error) {
	ctx, span := ep.tracer.Start(ctx, "extractEmailContent")
	defer span.End()

	content := &EmailContent{}

	// Extract metadata
	content.Metadata = ep.extractMetadata(msg)

	// Parse content based on MIME type
	mediaType, params, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
	if err != nil {
		mediaType = "text/plain" // Default to plain text
	}

	if strings.HasPrefix(mediaType, "multipart/") {
		mr := multipart.NewReader(msg.Body, params["boundary"])
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				continue // Skip problematic parts
			}

			partContent, err := ioutil.ReadAll(part)
			if err != nil {
				continue
			}

			contentType := part.Header.Get("Content-Type")
			switch {
			case strings.HasPrefix(contentType, "text/plain"):
				content.PlainText = string(partContent)
			case strings.HasPrefix(contentType, "text/html"):
				content.HTML = string(partContent)
			case strings.HasPrefix(contentType, "text/richtext"):
				content.RichText = string(partContent)
			default:
				// Handle attachments
				if filename := part.FileName(); filename != "" {
					content.Attachments = append(content.Attachments, EmailAttachment{
						Filename:    filename,
						ContentType: contentType,
						Data:        partContent,
					})
				}
			}
		}
	} else {
		// Handle single part messages
		body, err := ioutil.ReadAll(msg.Body)
		if err == nil {
			if strings.HasPrefix(mediaType, "text/html") {
				content.HTML = string(body)
			} else {
				content.PlainText = string(body)
			}
		}
	}

	// Extract links from HTML content
	content.Links = ep.extractLinks(content.HTML)

	return content, nil
}

func (ep *emailProcessorImpl) extractLinks(htmlContent string) []string {
	// Simple link extraction - can be improved with proper HTML parsing
	links := []string{}
	startIdx := 0
	for {
		hrefIdx := strings.Index(htmlContent[startIdx:], "href=\"")
		if hrefIdx == -1 {
			break
		}
		hrefIdx += startIdx + 6 // len("href=\"")
		endIdx := strings.Index(htmlContent[hrefIdx:], "\"")
		if endIdx == -1 {
			break
		}
		endIdx += hrefIdx
		link := htmlContent[hrefIdx:endIdx]
		if strings.HasPrefix(link, "http") {
			links = append(links, link)
		}
		startIdx = endIdx
	}
	return links
}

func (ep *emailProcessorImpl) extractMetadata(msg *mail.Message) EmailMetadata {
	metadata := EmailMetadata{
		MessageID:         msg.Header.Get("Message-ID"),
		ContentType:       msg.Header.Get("Content-Type"),
		ContentTransfer:   msg.Header.Get("Content-Transfer-Encoding"),
		ContentDispostion: msg.Header.Get("Content-Disposition"),
	}

	// Parse date
	if date := msg.Header.Get("Date"); date != "" {
		if t, err := mail.ParseDate(date); err == nil {
			metadata.Date = t
		}
	}

	// Parse addresses
	if from := msg.Header.Get("From"); from != "" {
		if addr, err := mail.ParseAddress(from); err == nil {
			metadata.From = addr
		}
	}

	metadata.To = ep.parseAddressList(msg.Header.Get("To"))
	metadata.Cc = ep.parseAddressList(msg.Header.Get("Cc"))
	metadata.ReplyTo = ep.parseAddressList(msg.Header.Get("Reply-To"))

	// Parse references
	if refs := msg.Header.Get("References"); refs != "" {
		metadata.References = strings.Fields(refs)
	}
	metadata.InReplyTo = msg.Header.Get("In-Reply-To")

	return metadata
}

func (ep *emailProcessorImpl) parseAddressList(addresses string) []*mail.Address {
	if addresses == "" {
		return nil
	}
	if list, err := mail.ParseAddressList(addresses); err == nil {
		return list
	}
	return nil
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

func (ep *emailProcessorImpl) extractEventInfo(ctx context.Context, msg *mail.Message, content *EmailContent) (*EmailEvent, error) {
	ctx, span := ep.tracer.Start(ctx, "extractEventInfo")
	defer span.End()

	subject := msg.Header.Get("Subject")

	// Combine all text content for NER processing
	textContent := strings.Join([]string{
		content.PlainText,
		ep.stripHTML(content.HTML), // Strip HTML tags for text processing
		content.RichText,
	}, "\n")

	// Extract dates using NER service
	dates, err := ep.extractDates(ctx, subject, textContent)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	if len(dates) < 2 {
		return nil, fmt.Errorf("could not extract start and end times")
	}

	// First date is start time, second is end time
	startTime := dates[0]
	endTime := dates[1]

	// Extract location using NER
	location, err := ep.nerService.ExtractLocation(ctx, textContent)
	if err != nil {
		// Log error but don't fail - location is optional
		span.RecordError(err)
	}

	// Extract attendees from headers and content
	attendees := ep.extractAttendees(msg.Header)

	return &EmailEvent{
		Subject:     subject,
		Description: textContent,
		StartTime:   startTime,
		EndTime:     endTime,
		Location:    location,
		Attendees:   attendees,
		Metadata:    content.Metadata,
		Attachments: content.Attachments,
	}, nil
}

func (ep *emailProcessorImpl) stripHTML(html string) string {
	// Simple HTML stripping - can be improved with proper HTML parsing
	text := strings.ReplaceAll(html, "<br>", "\n")
	text = strings.ReplaceAll(text, "<br/>", "\n")
	text = strings.ReplaceAll(text, "<br />", "\n")

	// Remove all other HTML tags
	for {
		start := strings.Index(text, "<")
		if start == -1 {
			break
		}
		end := strings.Index(text[start:], ">")
		if end == -1 {
			break
		}
		text = text[:start] + " " + text[start+end+1:]
	}

	return strings.TrimSpace(text)
}

func (ep *emailProcessorImpl) extractDates(ctx context.Context, subject, body string) ([]time.Time, error) {
	// Combine subject and body for date extraction
	text := subject + "\n" + body

	// Extract dates using NER service
	dates, err := ep.nerService.ExtractDateTime(ctx, text)
	if err != nil {
		return nil, fmt.Errorf("failed to extract dates: %v", err)
	}

	// Sort dates by time
	sortDates(dates)

	// If only one date found, use it as start time and add 1 hour for end time
	if len(dates) == 1 {
		dates = append(dates, dates[0].Add(1*time.Hour))
	}

	// Ensure we have at least two dates
	if len(dates) < 2 {
		now := time.Now()
		dates = []time.Time{now, now.Add(1 * time.Hour)}
	}

	return dates, nil
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

// sortDates sorts a slice of dates in ascending order
func sortDates(dates []time.Time) {
	for i := 0; i < len(dates)-1; i++ {
		for j := i + 1; j < len(dates); j++ {
			if dates[j].Before(dates[i]) {
				dates[i], dates[j] = dates[j], dates[i]
			}
		}
	}
}
