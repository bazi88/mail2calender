package usecase

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Frequency constants
const (
	FreqDaily   = "DAILY"
	FreqWeekly  = "WEEKLY"
	FreqMonthly = "MONTHLY"
	FreqYearly  = "YEARLY"
	FreqHourly  = "HOURLY"
)

// Weekday type and constants
type Weekday string

const (
	Monday    Weekday = "MO"
	Tuesday   Weekday = "TU"
	Wednesday Weekday = "WE"
	Thursday  Weekday = "TH"
	Friday    Weekday = "FR"
	Saturday  Weekday = "SA"
	Sunday    Weekday = "SU"
)

// RecurrenceRule represents a recurring event rule
type RecurrenceRule struct {
	Frequency  string
	Count      *int
	Interval   int
	ByDay      []Weekday
	ByMonth    []time.Month
	ByMonthDay []int
}

// ParseRecurrenceRule parses an RRULE string into a RecurrenceRule struct
func ParseRecurrenceRule(ruleStr string) (*RecurrenceRule, error) {
	if !strings.HasPrefix(ruleStr, "RRULE:") {
		return nil, fmt.Errorf("invalid recurrence rule format: missing RRULE prefix")
	}

	rule := &RecurrenceRule{
		Interval: 1, // Default interval
	}

	parts := strings.Split(strings.TrimPrefix(ruleStr, "RRULE:"), ";")

	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}

		key := kv[0]
		value := kv[1]

		switch key {
		case "FREQ":
			rule.Frequency = value
		case "COUNT":
			count, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("invalid COUNT value: %v", err)
			}
			rule.Count = &count
		case "INTERVAL":
			interval, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("invalid INTERVAL value: %v", err)
			}
			rule.Interval = interval
		case "BYDAY":
			days := strings.Split(value, ",")
			rule.ByDay = make([]Weekday, len(days))
			for i, day := range days {
				rule.ByDay[i] = Weekday(day)
			}
		case "BYMONTH":
			monthStrs := strings.Split(value, ",")
			for _, monthStr := range monthStrs {
				month, err := strconv.Atoi(monthStr)
				if err != nil {
					return nil, fmt.Errorf("invalid BYMONTH value: %v", err)
				}
				rule.ByMonth = append(rule.ByMonth, time.Month(month))
			}
		case "BYMONTHDAY":
			dayStrs := strings.Split(value, ",")
			for _, dayStr := range dayStrs {
				day, err := strconv.Atoi(dayStr)
				if err != nil {
					return nil, fmt.Errorf("invalid BYMONTHDAY value: %v", err)
				}
				rule.ByMonthDay = append(rule.ByMonthDay, day)
			}
		}
	}

	return rule, nil
}

// GetRecurrences returns all recurrence times within the given range
func (r *RecurrenceRule) GetRecurrences(start time.Time, end time.Time, duration time.Duration) []TimeSlot {
	var slots []TimeSlot
	count := 0
	maxCount := -1
	if r.Count != nil {
		maxCount = *r.Count
	}

	interval := time.Duration(r.Interval)

	switch r.Frequency {
	case FreqDaily:
		for current := start; !current.After(end) && (maxCount == -1 || count < maxCount); current = current.AddDate(0, 0, int(interval)) {
			slots = append(slots, TimeSlot{
				Start: current,
				End:   current.Add(duration),
			})
			count++
		}

	case FreqWeekly:
		for current := start; !current.After(end) && (maxCount == -1 || count < maxCount); current = current.AddDate(0, 0, 7*int(interval)) {
			if len(r.ByDay) == 0 {
				slots = append(slots, TimeSlot{
					Start: current,
					End:   current.Add(duration),
				})
				count++
			} else {
				// Generate slots for each specified weekday
				for _, day := range r.ByDay {
					daySlot := current
					switch day {
					case Monday:
						daySlot = getNextWeekday(current, time.Monday)
					case Tuesday:
						daySlot = getNextWeekday(current, time.Tuesday)
					case Wednesday:
						daySlot = getNextWeekday(current, time.Wednesday)
					case Thursday:
						daySlot = getNextWeekday(current, time.Thursday)
					case Friday:
						daySlot = getNextWeekday(current, time.Friday)
					case Saturday:
						daySlot = getNextWeekday(current, time.Saturday)
					case Sunday:
						daySlot = getNextWeekday(current, time.Sunday)
					}
					if !daySlot.After(end) {
						slots = append(slots, TimeSlot{
							Start: daySlot,
							End:   daySlot.Add(duration),
						})
					}
				}
				count++
			}
		}

	case FreqMonthly:
		for current := start; !current.After(end) && (maxCount == -1 || count < maxCount); current = current.AddDate(0, int(interval), 0) {
			if len(r.ByMonthDay) > 0 {
				for _, day := range r.ByMonthDay {
					daySlot := time.Date(current.Year(), current.Month(), day, current.Hour(), current.Minute(), current.Second(), current.Nanosecond(), current.Location())
					if !daySlot.After(end) && !daySlot.Before(start) {
						slots = append(slots, TimeSlot{
							Start: daySlot,
							End:   daySlot.Add(duration),
						})
					}
				}
			} else {
				slots = append(slots, TimeSlot{
					Start: current,
					End:   current.Add(duration),
				})
			}
			count++
		}

	case FreqYearly:
		for current := start; !current.After(end) && (maxCount == -1 || count < maxCount); current = current.AddDate(int(interval), 0, 0) {
			if len(r.ByMonth) > 0 {
				year := current.Year()
				for _, month := range r.ByMonth {
					monthSlot := time.Date(year, month, current.Day(), current.Hour(), current.Minute(), current.Second(), current.Nanosecond(), current.Location())
					if !monthSlot.After(end) && !monthSlot.Before(start) {
						slots = append(slots, TimeSlot{
							Start: monthSlot,
							End:   monthSlot.Add(duration),
						})
					}
				}
			} else {
				slots = append(slots, TimeSlot{
					Start: current,
					End:   current.Add(duration),
				})
			}
			count++
		}
	}

	return slots
}

// Helper function to get the next occurrence of a weekday
func getNextWeekday(current time.Time, weekday time.Weekday) time.Time {
	daysUntil := int(weekday - current.Weekday())
	if daysUntil <= 0 {
		daysUntil += 7
	}
	return current.AddDate(0, 0, daysUntil)
}
