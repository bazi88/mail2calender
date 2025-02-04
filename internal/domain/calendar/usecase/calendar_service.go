package usecase

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// CalendarService coordinates email processing and calendar integration
type CalendarService struct {
	emailProcessor EmailProcessor
	calendar       GoogleCalendarService
	tracer         trace.Tracer
}

// NewCalendarService creates a new instance of CalendarService
func NewCalendarService(
	emailProcessor EmailProcessor,
	calendar GoogleCalendarService,
) *CalendarService {
	return &CalendarService{
		emailProcessor: emailProcessor,
		calendar:       calendar,
		tracer:         otel.Tracer("calendar-service"),
	}
}

// ProcessEmailToCalendar handles the complete flow from email to calendar event
func (s *CalendarService) ProcessEmailToCalendar(ctx context.Context, emailContent string, userID string) error {
	ctx, span := s.tracer.Start(ctx, "ProcessEmailToCalendar")
	defer span.End()

	span.SetAttributes(attribute.String("user_id", userID))

	// 1. Validate email
	if err := s.emailProcessor.ValidateEmail(ctx, emailContent); err != nil {
		span.RecordError(err)
		return fmt.Errorf("email validation failed: %v", err)
	}

	// 2. Process email to extract event
	event, err := s.emailProcessor.ProcessEmail(ctx, emailContent)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("email processing failed: %v", err)
	}

	// 3. Create calendar event
	calendarEvent, err := s.calendar.CreateCalendarEvent(ctx, event, userID)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("calendar event creation failed: %v", err)
	}

	span.SetAttributes(
		attribute.String("event.id", calendarEvent.Id),
		attribute.String("event.summary", calendarEvent.Summary),
	)

	return nil
}

// UpdateCalendarEvent updates an existing calendar event
func (s *CalendarService) UpdateCalendarEvent(ctx context.Context, emailContent string, eventID string, userID string) error {
	ctx, span := s.tracer.Start(ctx, "UpdateCalendarEvent")
	defer span.End()

	span.SetAttributes(
		attribute.String("user_id", userID),
		attribute.String("event_id", eventID),
	)

	// 1. Validate email
	if err := s.emailProcessor.ValidateEmail(ctx, emailContent); err != nil {
		span.RecordError(err)
		return fmt.Errorf("email validation failed: %v", err)
	}

	// 2. Process email to extract event
	event, err := s.emailProcessor.ProcessEmail(ctx, emailContent)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("email processing failed: %v", err)
	}

	// 3. Update calendar event
	_, err = s.calendar.UpdateCalendarEvent(ctx, eventID, event, userID)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("calendar event update failed: %v", err)
	}

	return nil
}

// DeleteCalendarEvent removes an event from the calendar
func (s *CalendarService) DeleteCalendarEvent(ctx context.Context, eventID string, userID string) error {
	ctx, span := s.tracer.Start(ctx, "DeleteCalendarEvent")
	defer span.End()

	span.SetAttributes(
		attribute.String("user_id", userID),
		attribute.String("event_id", eventID),
	)

	if err := s.calendar.DeleteCalendarEvent(ctx, eventID, userID); err != nil {
		span.RecordError(err)
		return fmt.Errorf("calendar event deletion failed: %v", err)
	}

	return nil
}
