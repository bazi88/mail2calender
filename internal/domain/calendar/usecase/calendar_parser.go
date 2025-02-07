package usecase

import (
	"bytes"
	"fmt"
	"time"

	ical "github.com/arran4/golang-ical"
)

// ICalendarParser defines methods for parsing calendar data
type ICalendarParser interface {
	ParseICSAttachment(data []byte) (*CalendarEvent, error)
}

type calendarParserImpl struct{}

func NewCalendarParser() ICalendarParser {
	return &calendarParserImpl{}
}

func (p *calendarParserImpl) ParseICSAttachment(data []byte) (*CalendarEvent, error) {
	calendar, err := ical.ParseCalendar(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ICS: %w", err)
	}

	// Get the first event from the calendar
	for _, event := range calendar.Events() {
		startTime, _ := time.Parse("20060102T150405Z", event.GetProperty(ical.ComponentPropertyDtStart).Value)
		endTime, _ := time.Parse("20060102T150405Z", event.GetProperty(ical.ComponentPropertyDtEnd).Value)

		attendees := make([]string, 0)
		for _, attendee := range event.Attendees() {
			attendees = append(attendees, attendee.Email())
		}

		return &CalendarEvent{
			Title:       event.GetProperty(ical.ComponentPropertySummary).Value,
			StartTime:   startTime,
			EndTime:     endTime,
			Location:    event.GetProperty(ical.ComponentPropertyLocation).Value,
			Attendees:   attendees,
			IsRecurring: event.GetProperty(ical.ComponentPropertyRrule) != nil,
		}, nil
	}

	return nil, fmt.Errorf("no events found in ICS file")
}
