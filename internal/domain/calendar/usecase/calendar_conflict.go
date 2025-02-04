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

// ConflictResult represents the result of conflict checking
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
	// Get existing events for all attendees during the event time
	existingEvents, err := cc.calendarService.GetEvents(ctx, TimeRange{
		StartTime: event.StartTime,
		EndTime:   event.EndTime,
	}, event.Attendees)

	if err != nil {
		return nil, err
	}

	// Check for conflicts
	var conflictingEvent *CalendarEvent
	for _, existing := range existingEvents {
		if cc.eventsOverlap(event, existing) {
			conflictingEvent = existing
			break
		}
	}

	if conflictingEvent == nil {
		return &ConflictResult{
			HasConflict: false,
		}, nil
	}

	// If there's a conflict, find alternative time slots
	alternatives, err := cc.findAlternativeSlots(ctx, event, existingEvents)
	if err != nil {
		return nil, err
	}

	return &ConflictResult{
		HasConflict:      true,
		ConflictingEvent: conflictingEvent,
		Alternatives:     alternatives,
	}, nil
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
	merged := cc.mergeBusyPeriods(busyPeriods)

	// Find gaps between busy periods that fit the duration
	var availableSlots []TimeSlot
	current := timeRange.StartTime

	for _, busy := range merged {
		if current.Add(timeRange.Duration).Before(busy.Start) {
			availableSlots = append(availableSlots, TimeSlot{
				Start: current,
				End:   current.Add(timeRange.Duration),
			})
		}
		current = busy.End
	}

	// Check if there's space after the last busy period
	if current.Add(timeRange.Duration).Before(timeRange.EndTime) {
		availableSlots = append(availableSlots, TimeSlot{
			Start: current,
			End:   current.Add(timeRange.Duration),
		})
	}

	return availableSlots, nil
}

func (cc *conflictCheckerImpl) GetBusyPeriods(ctx context.Context, timeRange TimeRange, attendees []string) ([]TimeSlot, error) {
	// Get events for all attendees
	events, err := cc.calendarService.GetEvents(ctx, timeRange, attendees)
	if err != nil {
		return nil, err
	}

	// Convert events to time slots
	busyPeriods := make([]TimeSlot, 0, len(events))
	for _, event := range events {
		// Handle all-day events
		if event.IsAllDay {
			busyPeriods = append(busyPeriods, TimeSlot{
				Start: time.Date(event.StartTime.Year(), event.StartTime.Month(), event.StartTime.Day(), 0, 0, 0, 0, event.StartTime.Location()),
				End:   time.Date(event.EndTime.Year(), event.EndTime.Month(), event.EndTime.Day(), 23, 59, 59, 0, event.EndTime.Location()),
			})
			continue
		}

		// Handle recurring events
		if event.IsRecurring && event.RecurrenceRule != "" {
			recurrenceSlots := cc.expandRecurringEvent(event, timeRange)
			busyPeriods = append(busyPeriods, recurrenceSlots...)
			continue
		}

		// Regular events
		busyPeriods = append(busyPeriods, TimeSlot{
			Start: event.StartTime,
			End:   event.EndTime,
		})
	}

	return busyPeriods, nil
}

func (cc *conflictCheckerImpl) eventsOverlap(event1, event2 *CalendarEvent) bool {
	// If either event is recurring, expand it and check all instances
	if event1.IsRecurring || event2.IsRecurring {
		timeRange := TimeRange{
			StartTime: minTime(event1.StartTime, event2.StartTime),
			EndTime:   maxTime(event1.EndTime, event2.EndTime),
		}

		if event1.IsRecurring {
			slots1 := cc.expandRecurringEvent(event1, timeRange)
			for _, slot := range slots1 {
				if cc.timeSlotOverlaps(slot, TimeSlot{Start: event2.StartTime, End: event2.EndTime}) {
					return true
				}
			}
		}

		if event2.IsRecurring {
			slots2 := cc.expandRecurringEvent(event2, timeRange)
			for _, slot := range slots2 {
				if cc.timeSlotOverlaps(slot, TimeSlot{Start: event1.StartTime, End: event1.EndTime}) {
					return true
				}
			}
		}

		return false
	}

	// Handle all-day events
	if event1.IsAllDay || event2.IsAllDay {
		return cc.datesOverlap(
			event1.StartTime, event1.EndTime,
			event2.StartTime, event2.EndTime,
		)
	}

	// Regular events: check if one event's start time falls within the other event's time range
	return !(event1.EndTime.Before(event2.StartTime) || event1.StartTime.After(event2.EndTime))
}

func (cc *conflictCheckerImpl) timeSlotOverlaps(slot1, slot2 TimeSlot) bool {
	return !(slot1.End.Before(slot2.Start) || slot1.Start.After(slot2.End))
}

func (cc *conflictCheckerImpl) datesOverlap(start1, end1, start2, end2 time.Time) bool {
	// Compare dates without time
	date1Start := time.Date(start1.Year(), start1.Month(), start1.Day(), 0, 0, 0, 0, start1.Location())
	date1End := time.Date(end1.Year(), end1.Month(), end1.Day(), 23, 59, 59, 0, end1.Location())
	date2Start := time.Date(start2.Year(), start2.Month(), start2.Day(), 0, 0, 0, 0, start2.Location())
	date2End := time.Date(end2.Year(), end2.Month(), end2.Day(), 23, 59, 59, 0, end2.Location())

	return !(date1End.Before(date2Start) || date1Start.After(date2End))
}

func (cc *conflictCheckerImpl) findAlternativeSlots(ctx context.Context, event *CalendarEvent, existingEvents []*CalendarEvent) ([]TimeSlot, error) {
	// Define time range for searching alternatives (e.g., ±7 days from original time)
	searchRange := TimeRange{
		StartTime: event.StartTime.Add(-7 * 24 * time.Hour),
		EndTime:   event.StartTime.Add(7 * 24 * time.Hour),
		Duration:  event.EndTime.Sub(event.StartTime),
	}

	// Find available slots
	return cc.FindAvailableSlots(ctx, searchRange, event.Attendees)
}

func (cc *conflictCheckerImpl) mergeBusyPeriods(periods []TimeSlot) []TimeSlot {
	if len(periods) <= 1 {
		return periods
	}

	var merged []TimeSlot
	current := periods[0]

	for i := 1; i < len(periods); i++ {
		if current.End.After(periods[i].Start) || current.End.Equal(periods[i].Start) {
			// Merge overlapping periods
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
		// If we can't parse the rule, return just the original event
		return []TimeSlot{{Start: event.StartTime, End: event.EndTime}}
	}

	duration := event.EndTime.Sub(event.StartTime)
	return rule.GetRecurrences(event.StartTime, timeRange.EndTime, duration)
}

// Helper functions for time comparisons
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
