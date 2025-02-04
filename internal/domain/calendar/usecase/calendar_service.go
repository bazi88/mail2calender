package usecase

import (
	"context"
	"time"
)

// CalendarService defines the interface for calendar operations
type CalendarService interface {
	// GetEvents returns calendar events for the given time range and attendees
	GetEvents(ctx context.Context, timeRange TimeRange, attendees []string) ([]*CalendarEvent, error)

	// CreateEvent creates a new calendar event
	CreateEvent(ctx context.Context, event *CalendarEvent) error

	// UpdateEvent updates an existing calendar event
	UpdateEvent(ctx context.Context, event *CalendarEvent) error

	// DeleteEvent deletes an existing calendar event
	DeleteEvent(ctx context.Context, eventID string) error

	// GetWorkingHours returns working hours for given attendees
	GetWorkingHours(ctx context.Context, attendees []string) (map[string]*WorkingHours, error)
}

// WorkingHours represents a user's working hours
type WorkingHours struct {
	TimeZone string
	Schedule []WeeklySchedule
}

// WeeklySchedule represents working hours for each day of the week
type WeeklySchedule struct {
	DayOfWeek time.Weekday
	StartTime time.Time
	EndTime   time.Time
}

// calendarServiceImpl implements CalendarService interface
type calendarServiceImpl struct {
	googleCalendar GoogleCalendarService
}

// NewCalendarService creates a new calendar service instance
func NewCalendarService(googleCalendar GoogleCalendarService) CalendarService {
	return &calendarServiceImpl{
		googleCalendar: googleCalendar,
	}
}

func (cs *calendarServiceImpl) GetEvents(ctx context.Context, timeRange TimeRange, attendees []string) ([]*CalendarEvent, error) {
	// Get events from Google Calendar
	events, err := cs.googleCalendar.ListEvents(ctx, timeRange.StartTime, timeRange.EndTime, attendees)
	if err != nil {
		return nil, err
	}

	// Convert Google Calendar events to our domain model
	result := make([]*CalendarEvent, len(events))
	for i, event := range events {
		result[i] = &CalendarEvent{
			ID:             event.ID,
			Title:          event.Summary,
			StartTime:      event.Start,
			EndTime:        event.End,
			Location:       event.Location,
			Attendees:      event.Attendees,
			IsAllDay:       event.IsAllDay,
			IsRecurring:    event.IsRecurring,
			RecurrenceRule: event.RecurrenceRule,
		}
	}

	return result, nil
}

func (cs *calendarServiceImpl) CreateEvent(ctx context.Context, event *CalendarEvent) error {
	// Convert to Google Calendar event
	gEvent := &GoogleCalendarEvent{
		Summary:        event.Title,
		Start:          event.StartTime,
		End:            event.EndTime,
		Location:       event.Location,
		Attendees:      event.Attendees,
		IsAllDay:       event.IsAllDay,
		IsRecurring:    event.IsRecurring,
		RecurrenceRule: event.RecurrenceRule,
	}

	return cs.googleCalendar.CreateEvent(ctx, gEvent)
}

func (cs *calendarServiceImpl) UpdateEvent(ctx context.Context, event *CalendarEvent) error {
	// Convert to Google Calendar event
	gEvent := &GoogleCalendarEvent{
		ID:             event.ID,
		Summary:        event.Title,
		Start:          event.StartTime,
		End:            event.EndTime,
		Location:       event.Location,
		Attendees:      event.Attendees,
		IsAllDay:       event.IsAllDay,
		IsRecurring:    event.IsRecurring,
		RecurrenceRule: event.RecurrenceRule,
	}

	return cs.googleCalendar.UpdateEvent(ctx, gEvent)
}

func (cs *calendarServiceImpl) DeleteEvent(ctx context.Context, eventID string) error {
	return cs.googleCalendar.DeleteEvent(ctx, eventID)
}

func (cs *calendarServiceImpl) GetWorkingHours(ctx context.Context, attendees []string) (map[string]*WorkingHours, error) {
	// Get working hours from Google Calendar
	workingHours, err := cs.googleCalendar.GetWorkingHours(ctx, attendees)
	if err != nil {
		return nil, err
	}

	// Convert Google Calendar working hours to our domain model
	result := make(map[string]*WorkingHours)
	for email, hours := range workingHours {
		schedules := make([]WeeklySchedule, len(hours.Schedule))
		for i, schedule := range hours.Schedule {
			schedules[i] = WeeklySchedule{
				DayOfWeek: schedule.DayOfWeek,
				StartTime: schedule.StartTime,
				EndTime:   schedule.EndTime,
			}
		}

		result[email] = &WorkingHours{
			TimeZone: hours.TimeZone,
			Schedule: schedules,
		}
	}

	return result, nil
}

// GoogleCalendarEvent represents a Google Calendar event
type GoogleCalendarEvent struct {
	ID             string
	Summary        string
	Start          time.Time
	End            time.Time
	Location       string
	Attendees      []string
	IsAllDay       bool
	IsRecurring    bool
	RecurrenceRule string
}

// GoogleWorkingHours represents working hours from Google Calendar
type GoogleWorkingHours struct {
	TimeZone string
	Schedule []GoogleWeeklySchedule
}

// GoogleWeeklySchedule represents Google Calendar weekly schedule
type GoogleWeeklySchedule struct {
	DayOfWeek time.Weekday
	StartTime time.Time
	EndTime   time.Time
}

// GoogleCalendarService defines operations for Google Calendar
type GoogleCalendarService interface {
	// ListEvents lists events from Google Calendar
	ListEvents(ctx context.Context, startTime, endTime time.Time, attendees []string) ([]*GoogleCalendarEvent, error)

	// CreateEvent creates a new event in Google Calendar
	CreateEvent(ctx context.Context, event *GoogleCalendarEvent) error

	// UpdateEvent updates an existing event in Google Calendar
	UpdateEvent(ctx context.Context, event *GoogleCalendarEvent) error

	// DeleteEvent deletes an event from Google Calendar
	DeleteEvent(ctx context.Context, eventID string) error

	// GetWorkingHours gets working hours for attendees from Google Calendar
	GetWorkingHours(ctx context.Context, attendees []string) (map[string]*GoogleWorkingHours, error)
}
