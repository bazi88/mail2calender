package usecase

import (
	"context"
	"testing"
	"time"

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
		name           string
		event          *CalendarEvent
		existingEvents []*CalendarEvent
		expectConflict bool
		conflictingID  string
	}{
		{
			name: "conflict with daily recurring event",
			event: &CalendarEvent{
				ID:             "new-event",
				StartTime:      parseTime("2025-02-05T09:00:00Z"),
				EndTime:        parseTime("2025-02-05T10:00:00Z"),
				RecurrenceRule: "FREQ=DAILY",
			},
			existingEvents: []*CalendarEvent{
				{
					ID:             "existing-event",
					StartTime:      parseTime("2025-02-05T09:30:00Z"),
					EndTime:        parseTime("2025-02-05T10:30:00Z"),
					RecurrenceRule: "FREQ=DAILY",
				},
			},
			expectConflict: true,
			conflictingID:  "existing-event",
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
	checker := NewConflictChecker(nil)
	now := time.Now()

	tests := []struct {
		name           string
		timeRange      TimeRange
		existingEvents []Event
		want           int
		wantErr        bool
	}{
		{
			name: "find_slots_with_regular_events",
			timeRange: TimeRange{
				StartTime: now,
				EndTime:   now.Add(6 * time.Hour),
				Duration:  time.Hour,
			},
			existingEvents: []Event{
				{
					StartTime: now.Add(2 * time.Hour),
					EndTime:   now.Add(3 * time.Hour),
				},
			},
			want:    6,
			wantErr: false,
		},
		{
			name: "find_slots_with_recurring_events",
			timeRange: TimeRange{
				StartTime: now,
				EndTime:   now.Add(4 * time.Hour),
				Duration:  time.Hour,
			},
			existingEvents: []Event{
				{
					StartTime:      now.Add(time.Hour),
					EndTime:        now.Add(2 * time.Hour),
					RecurrenceRule: "FREQ=DAILY",
				},
			},
			want:    4,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checker.FindAvailableSlots(context.Background(), tt.timeRange, tt.existingEvents)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConflictChecker.FindAvailableSlots() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("ConflictChecker.FindAvailableSlots() = got %d slots, want %d slots", len(got), tt.want)
			}
		})
	}
}
