package usecase

import (
	"context"
	"sort"
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

// ConflictChecker handles calendar event conflict detection
type ConflictChecker interface {
	// CheckConflicts checks if an event conflicts with existing events
	CheckConflicts(ctx context.Context, event *CalendarEvent) (*ConflictResult, error)

	// FindAvailableSlots finds available time slots within a given range
	FindAvailableSlots(ctx context.Context, timeRange TimeRange, attendees []string) ([]TimeSlot, error)

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

func (cc *conflictCheckerImpl) FindAvailableSlots(ctx context.Context, timeRange TimeRange, attendees []string) ([]TimeSlot, error) {
	// Get busy periods for all attendees
	busyPeriods, err := cc.GetBusyPeriods(ctx, timeRange, attendees)
	if err != nil {
		return nil, err
	}

	// Sort busy periods by start time
	sort.Slice(busyPeriods, func(i, j int) bool {
		return busyPeriods[i].Start.Before(busyPeriods[j].Start)
	})

	// Merge overlapping busy periods
	mergedBusy := cc.mergeBusyPeriods(busyPeriods)

	// Generate all possible slots
	var allSlots []TimeSlot
	for current := timeRange.StartTime; current.Add(timeRange.Duration).Before(timeRange.EndTime) || current.Add(timeRange.Duration).Equal(timeRange.EndTime); {
		slot := TimeSlot{
			Start: current,
			End:   current.Add(timeRange.Duration),
		}
		allSlots = append(allSlots, slot)
		current = current.Add(timeRange.Duration)
	}

	// Filter out slots that overlap with busy periods
	availableSlots := make([]TimeSlot, 0, len(allSlots))
	for _, slot := range allSlots {
		overlaps := false
		for _, busy := range mergedBusy {
			if cc.timeSlotOverlaps(slot, busy) {
				overlaps = true
				break
			}
		}
		if !overlaps {
			availableSlots = append(availableSlots, slot)
		}
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

func (cc *conflictCheckerImpl) datesOverlap(start1, end1, start2, end2 time.Time) bool {
	date1Start := time.Date(start1.Year(), start1.Month(), start1.Day(), 0, 0, 0, 0, start1.Location())
	date1End := time.Date(end1.Year(), end1.Month(), end1.Day(), 23, 59, 59, 0, end1.Location())
	date2Start := time.Date(start2.Year(), start2.Month(), start2.Day(), 0, 0, 0, 0, start2.Location())
	date2End := time.Date(end2.Year(), end2.Month(), end2.Day(), 23, 59, 59, 0, end2.Location())

	return !(date1End.Before(date2Start) || date1Start.After(date2End))
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

func minTime(t1, t2 time.Time) time.Time {
	if t1.Before(t2) {
		return t1
	}
	return t2
}

func maxTime(t1, t2 time.Time) time.Time {
	if t1.After(t2) {
		return t1
	}
	return t2
}

func (cc *conflictCheckerImpl) checkDailyRecurrenceOverlap(event1, event2 *CalendarEvent) bool {
	// Daily recurrence overlap logic
	// This is a placeholder, you need to implement the actual logic
	return false
}

func (cc *conflictCheckerImpl) expandRecurringEvent(event *CalendarEvent, timeRange TimeRange) []TimeSlot {
	// Expand recurring event logic
	// This is a placeholder, you need to implement the actual logic
	return []TimeSlot{}
}
