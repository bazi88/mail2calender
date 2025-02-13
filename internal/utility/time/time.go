package time

import (
	"log"
	"time"
)

// Parse parses a date string into a time.Time value
func Parse(date string, format ...string) time.Time {
	switch {
	case len(format) == 0:
		return parse3339(date)
	case len(format) == 1:
		return parseWithFormat(date, format[0])
	default:
		return parse3339(date)
	}
}

func parseISO8601(iso8601 string) time.Time {
	timeWant, err := time.Parse("2006-01-02T15:04:05", iso8601)
	if err != nil {
		log.Panic(err)
	}
	return timeWant
}

func parse3339(rfc3339 string) time.Time {
	timeWant, err := time.Parse(time.RFC3339, rfc3339)
	if err != nil {
		log.Panic(err)
	}
	return timeWant
}

func parseWithFormat(date string, format string) time.Time {
	if format == time.RFC3339 {
		return parse3339(date)
	}
	panic("time format not supported")
}
