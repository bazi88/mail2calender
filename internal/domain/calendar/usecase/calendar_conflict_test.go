package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock calendar service
type mockCalendarService struct {
	mock.Mock
}

func (m *mockCalendarService) GetEvents(ctx context.Context, timeRange TimeRange, attendees []string) ([]*CalendarEvent, error) {
	args := m.Called(ctx, timeRange, attendees)
	return args.Get(0).([]*CalendarEvent), args.Error(1)
}

func (m *mockCalendarService) CreateEvent(ctx context.Context, event *CalendarEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *mockCalendarService) UpdateEvent(ctx context.Context, event *CalendarEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *mockCalendarService) DeleteEvent(ctx context.Context, eventID string) error {
	args := m.Called(ctx, eventID)
	return args.Error(0)
}

func (m *mockCalendarService) GetWorkingHours(ctx context.Context, attendees []string) (map[string]*WorkingHours, error) {
	args := m.Called(ctx, attendees)
	return args.Get(0).(map[string]*WorkingHours), args.Error(1)
}

func TestConflictChecker_CheckConflicts(t *testing.T) {
	loc, _ := time.LoadLocation("UTC")
	now := time.Date(2025, 1, 1, 10, 0, 0, 0, loc)

	tests := []struct {
		name            string
		event           *CalendarEvent
		existingEvents  []*CalendarEvent
		expectConflict  bool
		conflictingWith string
	}{
		{
			name: "no conflict with regular events",
			event: &CalendarEvent{
				ID:        "new",
				Title:     "New Meeting",
				StartTime: now,
				EndTime:   now.Add(1 * time.Hour),
			},
			existingEvents: []*CalendarEvent{
				{
					ID:        "existing",
					Title:     "Existing Meeting",
					StartTime: now.Add(2 * time.Hour),
					EndTime:   now.Add(3 * time.Hour),
				},
			},
			expectConflict: false,
		},
		{
			name: "conflict with regular event",
			event: &CalendarEvent{
				ID:        "new",
				Title:     "New Meeting",
				StartTime: now,
				EndTime:   now.Add(1 * time.Hour),
			},
			existingEvents: []*CalendarEvent{
				{
					ID:        "existing",
					Title:     "Existing Meeting",
					StartTime: now.Add(30 * time.Minute),
					EndTime:   now.Add(90 * time.Minute),
				},
			},
			expectConflict:  true,
			conflictingWith: "existing",
		},
		{
			name: "conflict with daily recurring event",
			event: &CalendarEvent{
				ID:        "new",
				Title:     "New Meeting",
				StartTime: now.AddDate(0, 0, 2),
				EndTime:   now.AddDate(0, 0, 2).Add(1 * time.Hour),
			},
			existingEvents: []*CalendarEvent{
				{
					ID:             "recurring",
					Title:          "Daily Standup",
					StartTime:      now,
					EndTime:        now.Add(30 * time.Minute),
					IsRecurring:    true,
					RecurrenceRule: "RRULE:FREQ=DAILY;COUNT=5",
				},
			},
			expectConflict:  true,
			conflictingWith: "recurring",
		},
		{
			name: "conflict with weekly recurring event",
			event: &CalendarEvent{
				ID:        "new",
				Title:     "New Meeting",
				StartTime: now.AddDate(0, 0, 7),
				EndTime:   now.AddDate(0, 0, 7).Add(1 * time.Hour),
			},
			existingEvents: []*CalendarEvent{
				{
					ID:             "recurring",
					Title:          "Weekly Team Meeting",
					StartTime:      now,
					EndTime:        now.Add(1 * time.Hour),
					IsRecurring:    true,
					RecurrenceRule: "RRULE:FREQ=WEEKLY;COUNT=4",
				},
			},
			expectConflict:  true,
			conflictingWith: "recurring",
		},
		{
			name: "conflict between two recurring events",
			event: &CalendarEvent{
				ID:             "new",
				Title:          "New Recurring",
				StartTime:      now,
				EndTime:        now.Add(1 * time.Hour),
				IsRecurring:    true,
				RecurrenceRule: "RRULE:FREQ=WEEKLY;BYDAY=MO,WE,FR",
			},
			existingEvents: []*CalendarEvent{
				{
					ID:             "existing",
					Title:          "Existing Recurring",
					StartTime:      now,
					EndTime:        now.Add(30 * time.Minute),
					IsRecurring:    true,
					RecurrenceRule: "RRULE:FREQ=WEEKLY;BYDAY=WE,FR",
				},
			},
			expectConflict:  true,
			conflictingWith: "existing",
		},
		{
			name: "conflict with all-day recurring event",
			event: &CalendarEvent{
				ID:        "new",
				Title:     "New Meeting",
				StartTime: now,
				EndTime:   now.Add(1 * time.Hour),
			},
			existingEvents: []*CalendarEvent{
				{
					ID:             "recurring",
					Title:          "Company Holiday",
					StartTime:      time.Date(2025, 1, 1, 0, 0, 0, 0, loc),
					EndTime:        time.Date(2025, 1, 1, 23, 59, 59, 0, loc),
					IsRecurring:    true,
					IsAllDay:       true,
					RecurrenceRule: "RRULE:FREQ=YEARLY;BYMONTH=1",
				},
			},
			expectConflict:  true,
			conflictingWith: "recurring",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockCalendarService)
			mockService.On("GetEvents", mock.Anything, mock.Anything, mock.Anything).
				Return(tt.existingEvents, nil)

			checker := NewConflictChecker(mockService)
			result, err := checker.CheckConflicts(context.Background(), tt.event)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectConflict, result.HasConflict)

			if tt.expectConflict {
				assert.NotNil(t, result.ConflictingEvent)
				assert.Equal(t, tt.conflictingWith, result.ConflictingEvent.ID)
				assert.NotEmpty(t, result.Alternatives)
			} else {
				assert.Nil(t, result.ConflictingEvent)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestConflictChecker_FindAvailableSlots(t *testing.T) {
	loc, _ := time.LoadLocation("UTC")
	now := time.Date(2025, 1, 1, 10, 0, 0, 0, loc)

	tests := []struct {
		name           string
		timeRange      TimeRange
		existingEvents []*CalendarEvent
		expectedSlots  int
	}{
		{
			name: "find slots with no existing events",
			timeRange: TimeRange{
				StartTime: now,
				EndTime:   now.Add(4 * time.Hour),
				Duration:  30 * time.Minute,
			},
			existingEvents: nil,
			expectedSlots:  8, // 4 hours with 30-minute slots
		},
		{
			name: "find slots with regular events",
			timeRange: TimeRange{
				StartTime: now,
				EndTime:   now.Add(4 * time.Hour),
				Duration:  30 * time.Minute,
			},
			existingEvents: []*CalendarEvent{
				{
					StartTime: now.Add(1 * time.Hour),
					EndTime:   now.Add(2 * time.Hour),
				},
			},
			expectedSlots: 6, // 8 possible slots - 2 blocked slots
		},
		{
			name: "find slots with recurring events",
			timeRange: TimeRange{
				StartTime: now,
				EndTime:   now.Add(8 * time.Hour),
				Duration:  1 * time.Hour,
			},
			existingEvents: []*CalendarEvent{
				{
					StartTime:      now,
					EndTime:        now.Add(30 * time.Minute),
					IsRecurring:    true,
					RecurrenceRule: "RRULE:FREQ=HOURLY;COUNT=4",
				},
			},
			expectedSlots: 4, // 8 hours - 4 blocked hours
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockCalendarService)
			mockService.On("GetEvents", mock.Anything, mock.Anything, mock.Anything).
				Return(tt.existingEvents, nil)

			checker := NewConflictChecker(mockService)
			slots, err := checker.FindAvailableSlots(context.Background(), tt.timeRange, []string{})

			assert.NoError(t, err)
			assert.Len(t, slots, tt.expectedSlots)

			// Verify slot durations
			for _, slot := range slots {
				assert.Equal(t, tt.timeRange.Duration, slot.End.Sub(slot.Start))
			}

			// Verify slots are within range
			for _, slot := range slots {
				assert.True(t, !slot.Start.Before(tt.timeRange.StartTime))
				assert.True(t, !slot.End.After(tt.timeRange.EndTime))
			}

			mockService.AssertExpectations(t)
		})
	}
}
