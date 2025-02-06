package usecase

import (
	"bytes"
	"fmt"
	"time"
)

type CalendarEvent interface {
	GetSummary() string
	GetDescription() string
	GetStartTime() time.Time
	GetEndTime() time.Time
	GetLocation() string
}

type ICSEvent struct {
	Summary     string
	Description string
	StartTime   time.Time
	EndTime     time.Time
	Location    string
}

func (e *ICSEvent) GetSummary() string      { return e.Summary }
func (e *ICSEvent) GetDescription() string  { return e.Description }
func (e *ICSEvent) GetStartTime() time.Time { return e.StartTime }
func (e *ICSEvent) GetEndTime() time.Time   { return e.EndTime }
func (e *ICSEvent) GetLocation() string     { return e.Location }

func parseICSAttachment(data []byte) (CalendarEvent, error) {
	calendar, err := ical.ParseCalendar(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ICS: %w", err)
	}

	// Get the first event from the calendar
	for _, event := range calendar.Events() {
		icsEvent := &ICSEvent{
			Summary:     event.GetProperty(ical.ComponentPropertySummary).Value,
			Description: event.GetProperty(ical.ComponentPropertyDescription).Value,
			Location:    event.GetProperty(ical.ComponentPropertyLocation).Value,
		}

		// Parse start and end times
		if startProp := event.GetProperty(ical.ComponentPropertyDtStart); startProp != nil {
			icsEvent.StartTime, _ = time.Parse("20060102T150405Z", startProp.Value)
		}
		if endProp := event.GetProperty(ical.ComponentPropertyDtEnd); endProp != nil {
			icsEvent.EndTime, _ = time.Parse("20060102T150405Z", endProp.Value)
		}

		return icsEvent, nil
	}

	return nil, fmt.Errorf("no events found in ICS file")
}
