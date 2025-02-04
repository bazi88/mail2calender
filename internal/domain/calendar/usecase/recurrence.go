package usecase

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// RecurrenceFrequency defines how often the event repeats
type RecurrenceFrequency string

const (
	FreqDaily   RecurrenceFrequency = "DAILY"
	FreqWeekly  RecurrenceFrequency = "WEEKLY"
	FreqMonthly RecurrenceFrequency = "MONTHLY"
	FreqYearly  RecurrenceFrequency = "YEARLY"
)

// Weekday represents days of the week
type Weekday int

const (
	Sunday Weekday = iota
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
)

// RecurrenceRule represents an RFC 5545 RRULE
type RecurrenceRule struct {
	Frequency  RecurrenceFrequency
	Until      *time.Time   // End date for recurrence
	Count      *int         // Number of occurrences
	Interval   int          // Frequency interval (every N days/weeks/etc)
	ByDay      []Weekday    // Days of the week
	ByMonthDay []int        // Days of the month
	ByMonth    []time.Month // Months of the year
}

// ParseRecurrenceRule parses an RFC 5545 RRULE string
func ParseRecurrenceRule(rule string) (*RecurrenceRule, error) {
	if !strings.HasPrefix(rule, "RRULE:") {
		return nil, fmt.Errorf("invalid RRULE format: missing RRULE prefix")
	}

	parts := strings.Split(strings.TrimPrefix(rule, "RRULE:"), ";")
	r := &RecurrenceRule{Interval: 1} // Default interval is 1

	for _, part := range parts {
		kv := strings.Split(part, "=")
		if len(kv) != 2 {
			continue
		}

		key, value := kv[0], kv[1]
		switch key {
		case "FREQ":
			r.Frequency = RecurrenceFrequency(value)
		case "UNTIL":
			t, err := time.Parse("20060102T150405Z", value)
			if err == nil {
				r.Until = &t
			}
		case "COUNT":
			if count, err := strconv.Atoi(value); err == nil {
				r.Count = &count
			}
		case "INTERVAL":
			if interval, err := strconv.Atoi(value); err == nil {
				r.Interval = interval
			}
		case "BYDAY":
			r.ByDay = parseByDay(value)
		case "BYMONTHDAY":
			r.ByMonthDay = parseByMonthDay(value)
		case "BYMONTH":
			r.ByMonth = parseByMonth(value)
		}
	}

	return r, nil
}

// GetRecurrences returns all recurrence dates between start and end
func (r *RecurrenceRule) GetRecurrences(start, end time.Time, eventDuration time.Duration) []TimeSlot {
	var slots []TimeSlot
	current := start
	count := 0

	for current.Before(end) {
		if r.Count != nil && count >= *r.Count {
			break
		}
		if r.Until != nil && current.After(*r.Until) {
			break
		}

		if r.isOccurrence(current) {
			slots = append(slots, TimeSlot{
				Start: current,
				End:   current.Add(eventDuration),
			})
			count++
		}

		// Advance to next potential occurrence
		switch r.Frequency {
		case FreqDaily:
			current = current.AddDate(0, 0, r.Interval)
		case FreqWeekly:
			current = current.AddDate(0, 0, 7*r.Interval)
		case FreqMonthly:
			current = current.AddDate(0, r.Interval, 0)
		case FreqYearly:
			current = current.AddDate(r.Interval, 0, 0)
		}
	}

	return slots
}

func (r *RecurrenceRule) isOccurrence(t time.Time) bool {
	// Check BYMONTH
	if len(r.ByMonth) > 0 {
		monthMatch := false
		for _, month := range r.ByMonth {
			if t.Month() == month {
				monthMatch = true
				break
			}
		}
		if !monthMatch {
			return false
		}
	}

	// Check BYMONTHDAY
	if len(r.ByMonthDay) > 0 {
		dayMatch := false
		for _, day := range r.ByMonthDay {
			if t.Day() == day {
				dayMatch = true
				break
			}
		}
		if !dayMatch {
			return false
		}
	}

	// Check BYDAY
	if len(r.ByDay) > 0 {
		dayMatch := false
		for _, day := range r.ByDay {
			if Weekday(t.Weekday()) == day {
				dayMatch = true
				break
			}
		}
		if !dayMatch {
			return false
		}
	}

	return true
}

// Helper functions to parse RRULE components
func parseByDay(value string) []Weekday {
	var days []Weekday
	dayMap := map[string]Weekday{
		"SU": Sunday,
		"MO": Monday,
		"TU": Tuesday,
		"WE": Wednesday,
		"TH": Thursday,
		"FR": Friday,
		"SA": Saturday,
	}

	for _, day := range strings.Split(value, ",") {
		if d, ok := dayMap[day]; ok {
			days = append(days, d)
		}
	}
	return days
}

func parseByMonthDay(value string) []int {
	var days []int
	for _, day := range strings.Split(value, ",") {
		if d, err := strconv.Atoi(day); err == nil && d >= 1 && d <= 31 {
			days = append(days, d)
		}
	}
	return days
}

func parseByMonth(value string) []time.Month {
	var months []time.Month
	for _, month := range strings.Split(value, ",") {
		if m, err := strconv.Atoi(month); err == nil && m >= 1 && m <= 12 {
			months = append(months, time.Month(m))
		}
	}
	return months
}
