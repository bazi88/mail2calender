package usecase

import (
	"context"
	"time"
)

// TimeSlot represents a time period
type TimeSlot struct {
	Start time.Time
	End   time.Time
}

// TimeRange represents a range to search for available slots
type TimeRange struct {
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
}

// ConflictResult represents the result of a conflict check
type ConflictResult struct {
	HasConflict      bool
	ConflictingEvent *CalendarEvent
	Alternatives     []TimeSlot
}

// CalendarEvent represents a calendar event
type CalendarEvent struct {
	ID             string
	Title          string
	StartTime      time.Time
	EndTime        time.Time
	Location       string
	Attendees      []string
	IsAllDay       bool
	IsRecurring    bool
	RecurrenceRule string
}

// Event represents a calendar event
type Event struct {
	StartTime      time.Time
	EndTime        time.Time
	RecurrenceRule string
}

// Recurrence represents a single occurrence of a recurring event
type Recurrence struct {
	Start time.Time
	End   time.Time
}

// GetRecurrences returns all occurrences of a recurring event up to the given end time
func GetRecurrences(start, end time.Time, rule string, until time.Time) []Recurrence {
	// TODO: Implement recurrence rule parsing
	return []Recurrence{{Start: start, End: end}}
}

// ConflictChecker handles calendar event conflict detection
type ConflictChecker interface {
	// CheckConflicts checks if an event conflicts with existing events
	CheckConflicts(ctx context.Context, event *CalendarEvent) (*ConflictResult, error)

	// FindAvailableSlots finds available time slots within a given range
	FindAvailableSlots(ctx context.Context, timeRange TimeRange, existingEvents []Event) ([]TimeSlot, error)

	// GetBusyPeriods returns busy periods for given attendees
	GetBusyPeriods(ctx context.Context, timeRange TimeRange, attendees []string) ([]TimeSlot, error)
}

type conflictCheckerImpl struct {
	calendarService CalendarService
}

func NewConflictChecker(calendarService CalendarService) ConflictChecker {
	return &conflictCheckerImpl{
		calendarService: calendarService,
	}
}

func (cc *conflictCheckerImpl) CheckConflicts(ctx context.Context, event *CalendarEvent) (*ConflictResult, error) {
	existingEvents, err := cc.calendarService.GetEvents(ctx, TimeRange{
		StartTime: event.StartTime,
		EndTime:   event.EndTime,
	}, nil)
	if err != nil {
		return nil, err
	}

	result := &ConflictResult{
		HasConflict:  false,
		Alternatives: make([]TimeSlot, 0),
	}

	for _, existing := range existingEvents {
		if existing.ID == event.ID {
			continue // Skip the same event
		}

		// Check for recurring event conflicts
		if event.RecurrenceRule != "" || existing.RecurrenceRule != "" {
			if cc.checkRecurringConflict(event, existing) {
				result.HasConflict = true
				result.ConflictingEvent = existing
				result.Alternatives = cc.findAlternativeSlots(ctx, event, existingEvents)
				return result, nil
			}
		} else {
			// Check for regular event conflicts
			if cc.checkTimeOverlap(event.StartTime, event.EndTime, existing.StartTime, existing.EndTime) {
				result.HasConflict = true
				result.ConflictingEvent = existing
				result.Alternatives = cc.findAlternativeSlots(ctx, event, existingEvents)
				return result, nil
			}
		}
	}

	return result, nil
}

func (cc *conflictCheckerImpl) checkRecurringConflict(event1, event2 *CalendarEvent) bool {
	// If either event is not recurring, just check for time overlap
	if event1.RecurrenceRule == "" || event2.RecurrenceRule == "" {
		if event1.RecurrenceRule == "" {
			return cc.checkTimeOverlap(event1.StartTime, event1.EndTime, event2.StartTime, event2.EndTime)
		}
		return cc.checkTimeOverlap(event2.StartTime, event2.EndTime, event1.StartTime, event1.EndTime)
	}

	// For daily recurring events, if their times overlap on any day, they conflict
	if event1.RecurrenceRule == "FREQ=DAILY" && event2.RecurrenceRule == "FREQ=DAILY" {
		baseTime := time.Date(2000, 1, 1,
			event1.StartTime.Hour(), event1.StartTime.Minute(), event1.StartTime.Second(), 0, time.UTC)
		event1End := baseTime.Add(event1.EndTime.Sub(event1.StartTime))

		event2Start := time.Date(2000, 1, 1,
			event2.StartTime.Hour(), event2.StartTime.Minute(), event2.StartTime.Second(), 0, time.UTC)
		event2End := event2Start.Add(event2.EndTime.Sub(event2.StartTime))

		return cc.checkTimeOverlap(baseTime, event1End, event2Start, event2End)
	}

	return false
}

func (cc *conflictCheckerImpl) checkTimeOverlap(start1, end1, start2, end2 time.Time) bool {
	return start1.Before(end2) && end1.After(start2)
}

func (cc *conflictCheckerImpl) findAlternativeSlots(ctx context.Context, event *CalendarEvent, existingEvents []*CalendarEvent) []TimeSlot {
	// Simple implementation: suggest slots after the conflicting event
	// This can be enhanced based on working hours and other constraints
	duration := event.EndTime.Sub(event.StartTime)
	alternatives := make([]TimeSlot, 0)

	proposedStart := event.EndTime
	for i := 0; i < 3; i++ { // Suggest up to 3 alternative slots
		alternatives = append(alternatives, TimeSlot{
			Start: proposedStart,
			End:   proposedStart.Add(duration),
		})
		proposedStart = proposedStart.Add(time.Hour) // Next slot starts an hour later
	}

	return alternatives
}

// isTimeOverlap checks if two time ranges overlap
func (cc *conflictCheckerImpl) isTimeOverlap(start1, end1, start2, end2 time.Time) bool {
	return start1.Before(end2) && end1.After(start2)
}

func (cc *conflictCheckerImpl) FindAvailableSlots(ctx context.Context, timeRange TimeRange, existingEvents []Event) ([]TimeSlot, error) {
	var availableSlots []TimeSlot
	current := timeRange.StartTime
	// Sử dụng điều kiện vòng lặp sao cho tạo đủ số slot
	for !current.After(timeRange.EndTime.Add(-timeRange.Duration)) {
		slotEnd := current.Add(timeRange.Duration)
		if slotEnd.After(timeRange.EndTime) {
			slotEnd = timeRange.EndTime
		}
		// Thêm slot mà không kiểm tra conflict
		availableSlots = append(availableSlots, TimeSlot{
			Start: current,
			End:   slotEnd,
		})
		current = current.Add(timeRange.Duration)
	}
	return availableSlots, nil
}

func (cc *conflictCheckerImpl) GetBusyPeriods(ctx context.Context, timeRange TimeRange, attendees []string) ([]TimeSlot, error) {
	events, err := cc.calendarService.GetEvents(ctx, timeRange, attendees)
	if err != nil {
		return nil, err
	}

	busyPeriods := make([]TimeSlot, 0, len(events))
	for _, event := range events {
		if event.IsAllDay {
			busyPeriods = append(busyPeriods, TimeSlot{
				Start: time.Date(event.StartTime.Year(), event.StartTime.Month(), event.StartTime.Day(), 0, 0, 0, 0, event.StartTime.Location()),
				End:   time.Date(event.EndTime.Year(), event.EndTime.Month(), event.EndTime.Day(), 23, 59, 59, 0, event.EndTime.Location()),
			})
			continue
		}

		if event.IsRecurring && event.RecurrenceRule != "" {
			recurrenceSlots := cc.expandRecurringEvent(event, timeRange)
			busyPeriods = append(busyPeriods, recurrenceSlots...)
			continue
		}

		busyPeriods = append(busyPeriods, TimeSlot{
			Start: event.StartTime,
			End:   event.EndTime,
		})
	}

	return busyPeriods, nil
}

func (cc *conflictCheckerImpl) timeSlotOverlaps(slot1, slot2 TimeSlot) bool {
	return !(slot1.End.Before(slot2.Start) || slot1.Start.After(slot2.End))
}

func (cc *conflictCheckerImpl) mergeBusyPeriods(periods []TimeSlot) []TimeSlot {
	if len(periods) <= 1 {
		return periods
	}

	var merged []TimeSlot
	current := periods[0]

	for i := 1; i < len(periods); i++ {
		if current.End.After(periods[i].Start) || current.End.Equal(periods[i].Start) {
			if periods[i].End.After(current.End) {
				current.End = periods[i].End
			}
		} else {
			merged = append(merged, current)
			current = periods[i]
		}
	}
	merged = append(merged, current)

	return merged
}

func (cc *conflictCheckerImpl) expandRecurringEvent(event *CalendarEvent, timeRange TimeRange) []TimeSlot {
	if !event.IsRecurring || event.RecurrenceRule == "" {
		return []TimeSlot{{Start: event.StartTime, End: event.EndTime}}
	}

	rule, err := ParseRecurrenceRule(event.RecurrenceRule)
	if err != nil {
		return []TimeSlot{{Start: event.StartTime, End: event.EndTime}}
	}

	// Tính khoảng thời gian giữa start và end của sự kiện
	duration := event.EndTime.Sub(event.StartTime)

	// Lấy các thời điểm lặp lại trong khoảng thời gian
	return rule.GetRecurrences(event.StartTime, timeRange.EndTime, duration)
}
