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

func parseTime(timeStr string) time.Time {
	t, _ := time.Parse(time.RFC3339, timeStr)
	return t
}

func TestConflictChecker_CheckConflicts(t *testing.T) {
	tests := []struct {
		name            string
		event          *CalendarEvent
		existingEvents []*CalendarEvent
		expectConflict bool
		conflictingID  string
	}{
		{
			name: "conflict with daily recurring event",
			event: &CalendarEvent{
				ID:        "new-event",
				StartTime: parseTime("2025-02-05T09:00:00Z"),
				EndTime:   parseTime("2025-02-05T10:00:00Z"),
				RecurrenceRule: "FREQ=DAILY",
			},
			existingEvents: []*CalendarEvent{
				{
					ID:        "existing-event",
					StartTime: parseTime("2025-02-05T09:30:00Z"),
					EndTime:   parseTime("2025-02-05T10:30:00Z"),
					RecurrenceRule: "FREQ=DAILY",
				},
			},
			expectConflict: true,
			conflictingID: "existing-event",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockCalendarService)
			mockService.On("GetEvents", mock.Anything, mock.Anything, mock.Anything).
				Return(tt.existingEvents, nil)

			checker := NewConflictChecker(mockService)
			result, err := checker.CheckConflicts(context.Background(), tt.event)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result == nil {
				t.Fatal("expected non-nil result")
			}

			if result.HasConflict != tt.expectConflict {
				t.Errorf("expected conflict=%v, got=%v", tt.expectConflict, result.HasConflict)
			}

			if tt.expectConflict {
				if result.ConflictingEvent == nil {
					t.Error("expected non-nil conflicting event")
				} else if result.ConflictingEvent.ID != tt.conflictingID {
					t.Errorf("expected conflicting event ID=%s, got=%s", 
						tt.conflictingID, result.ConflictingEvent.ID)
				}
				if len(result.Alternatives) == 0 {
					t.Error("expected non-empty alternatives")
				}
			}
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

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

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
