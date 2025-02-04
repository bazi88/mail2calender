package usecase

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	calendar "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

const defaultTimezone = "Asia/Ho_Chi_Minh"

type googleCalendarServiceImpl struct {
	oauthConfig *OAuthConfig
	tracer      trace.Tracer
	userID      string
}

// NewGoogleCalendarService creates a new instance of GoogleCalendarService
func NewGoogleCalendarService(oauth *OAuthConfig, tracer trace.Tracer, userID string) GoogleCalendarService {
	return &googleCalendarServiceImpl{
		oauthConfig: oauth,
		tracer:      tracer,
		userID:      userID,
	}
}

func (g *googleCalendarServiceImpl) ListEvents(ctx context.Context, startTime, endTime time.Time, attendees []string) ([]*GoogleCalendarEvent, error) {
	ctx, span := g.tracer.Start(ctx, "GoogleCalendar.ListEvents")
	defer span.End()

	span.SetAttributes(
		attribute.String("start_time", startTime.Format(time.RFC3339)),
		attribute.String("end_time", endTime.Format(time.RFC3339)),
		attribute.Int("attendees_count", len(attendees)),
	)

	client, err := g.getCalendarService(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get calendar service: %v", err)
	}

	// Query for events
	events, err := client.Events.List("primary").
		TimeMin(startTime.Format(time.RFC3339)).
		TimeMax(endTime.Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime").
		Do()

	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to list events: %v", err)
	}

	// Convert to domain events
	result := make([]*GoogleCalendarEvent, 0, len(events.Items))
	for _, event := range events.Items {
		// Extract attendees
		attendeesList := make([]string, 0, len(event.Attendees))
		for _, attendee := range event.Attendees {
			attendeesList = append(attendeesList, attendee.Email)
		}

		// Convert start time
		var startTime time.Time
		if event.Start.DateTime != "" {
			startTime, _ = time.Parse(time.RFC3339, event.Start.DateTime)
		} else {
			startTime, _ = time.Parse("2006-01-02", event.Start.Date)
		}

		// Convert end time
		var endTime time.Time
		if event.End.DateTime != "" {
			endTime, _ = time.Parse(time.RFC3339, event.End.DateTime)
		} else {
			endTime, _ = time.Parse("2006-01-02", event.End.Date)
		}

		result = append(result, &GoogleCalendarEvent{
			ID:             event.Id,
			Summary:        event.Summary,
			Start:          startTime,
			End:            endTime,
			Location:       event.Location,
			Attendees:      attendeesList,
			IsAllDay:       event.Start.DateTime == "",
			IsRecurring:    event.RecurringEventId != "",
			RecurrenceRule: firstOrEmpty(event.Recurrence),
		})
	}

	return result, nil
}

func (g *googleCalendarServiceImpl) CreateEvent(ctx context.Context, event *GoogleCalendarEvent) error {
	ctx, span := g.tracer.Start(ctx, "GoogleCalendar.CreateEvent")
	defer span.End()

	span.SetAttributes(
		attribute.String("event_id", event.ID),
		attribute.String("summary", event.Summary),
	)

	client, err := g.getCalendarService(ctx)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get calendar service: %v", err)
	}

	calendarEvent := &calendar.Event{
		Summary:     event.Summary,
		Location:    event.Location,
		Description: "",
		Start:       g.convertToEventDateTime(event.Start, event.IsAllDay),
		End:         g.convertToEventDateTime(event.End, event.IsAllDay),
	}

	// Add attendees
	for _, email := range event.Attendees {
		calendarEvent.Attendees = append(calendarEvent.Attendees, &calendar.EventAttendee{
			Email: email,
		})
	}

	// Add recurrence if specified
	if event.IsRecurring && event.RecurrenceRule != "" {
		calendarEvent.Recurrence = []string{event.RecurrenceRule}
	}

	_, err = client.Events.Insert("primary", calendarEvent).Do()
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create event: %v", err)
	}

	return nil
}

func (g *googleCalendarServiceImpl) UpdateEvent(ctx context.Context, event *GoogleCalendarEvent) error {
	ctx, span := g.tracer.Start(ctx, "GoogleCalendar.UpdateEvent")
	defer span.End()

	span.SetAttributes(
		attribute.String("event_id", event.ID),
		attribute.String("summary", event.Summary),
	)

	client, err := g.getCalendarService(ctx)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get calendar service: %v", err)
	}

	calendarEvent := &calendar.Event{
		Summary:     event.Summary,
		Location:    event.Location,
		Description: "",
		Start:       g.convertToEventDateTime(event.Start, event.IsAllDay),
		End:         g.convertToEventDateTime(event.End, event.IsAllDay),
	}

	// Add attendees
	for _, email := range event.Attendees {
		calendarEvent.Attendees = append(calendarEvent.Attendees, &calendar.EventAttendee{
			Email: email,
		})
	}

	// Add recurrence if specified
	if event.IsRecurring && event.RecurrenceRule != "" {
		calendarEvent.Recurrence = []string{event.RecurrenceRule}
	}

	_, err = client.Events.Update("primary", event.ID, calendarEvent).Do()
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to update event: %v", err)
	}

	return nil
}

func (g *googleCalendarServiceImpl) DeleteEvent(ctx context.Context, eventID string) error {
	ctx, span := g.tracer.Start(ctx, "GoogleCalendar.DeleteEvent")
	defer span.End()

	span.SetAttributes(attribute.String("event_id", eventID))

	client, err := g.getCalendarService(ctx)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get calendar service: %v", err)
	}

	err = client.Events.Delete("primary", eventID).Do()
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to delete event: %v", err)
	}

	return nil
}

func (g *googleCalendarServiceImpl) GetWorkingHours(ctx context.Context, attendees []string) (map[string]*GoogleWorkingHours, error) {
	ctx, span := g.tracer.Start(ctx, "GoogleCalendar.GetWorkingHours")
	defer span.End()

	span.SetAttributes(attribute.Int("attendees_count", len(attendees)))

	client, err := g.getCalendarService(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get calendar service: %v", err)
	}

	// Get timezone for primary calendar
	primaryTz := defaultTimezone
	settings, err := client.Settings.List().Do()
	if err == nil {
		for _, setting := range settings.Items {
			if setting.Id == "timezone" {
				primaryTz = setting.Value
				break
			}
		}
	}

	timeMin := time.Now().Format(time.RFC3339)
	timeMax := time.Now().AddDate(0, 0, 7).Format(time.RFC3339)

	// Build calendar items for query
	items := make([]*calendar.FreeBusyRequestItem, len(attendees))
	for i, email := range attendees {
		items[i] = &calendar.FreeBusyRequestItem{Id: email}
	}

	// Query free/busy information
	query := &calendar.FreeBusyRequest{
		TimeMin: timeMin,
		TimeMax: timeMax,
		Items:   items,
	}

	freeBusy, err := client.Freebusy.Query(query).Do()
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to query freebusy: %v", err)
	}

	// Process results
	result := make(map[string]*GoogleWorkingHours)
	for _, email := range attendees {
		if calendar, ok := freeBusy.Calendars[email]; ok {
			workingHours := &GoogleWorkingHours{
				TimeZone: primaryTz,
				Schedule: g.extractWorkingSchedule(calendar.Busy),
			}

			// Try to get user-specific timezone
			userSettings, err := client.Settings.List().Do()
			if err == nil {
				for _, setting := range userSettings.Items {
					if setting.Id == "timezone" {
						workingHours.TimeZone = setting.Value
						break
					}
				}
			}

			result[email] = workingHours
		}
	}

	return result, nil
}

// Helper functions

func (g *googleCalendarServiceImpl) getCalendarService(ctx context.Context) (*calendar.Service, error) {
	client, err := g.oauthConfig.GetClient(ctx, g.userID)
	if err != nil {
		return nil, err
	}

	return calendar.NewService(ctx, option.WithHTTPClient(client))
}

func (g *googleCalendarServiceImpl) convertToEventDateTime(t time.Time, isAllDay bool) *calendar.EventDateTime {
	if isAllDay {
		return &calendar.EventDateTime{
			Date: t.Format("2006-01-02"),
		}
	}
	return &calendar.EventDateTime{
		DateTime: t.Format(time.RFC3339),
	}
}

func (g *googleCalendarServiceImpl) extractWorkingSchedule(busySlots []*calendar.TimePeriod) []GoogleWeeklySchedule {
	// Default working hours (9 AM - 5 PM, Mon-Fri)
	schedules := make([]GoogleWeeklySchedule, 5)
	for i := 0; i < 5; i++ {
		schedules[i] = GoogleWeeklySchedule{
			DayOfWeek: time.Weekday(i + 1), // Monday = 1
			StartTime: time.Date(0, 0, 0, 9, 0, 0, 0, time.Local),
			EndTime:   time.Date(0, 0, 0, 17, 0, 0, 0, time.Local),
		}
	}
	return schedules
}

func firstOrEmpty(slice []string) string {
	if len(slice) > 0 {
		return slice[0]
	}
	return ""
}
