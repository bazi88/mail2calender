package usecase

import (
	"fmt"
	"strings"
	"time"
)

// TimezoneUtil handles timezone conversions and standardization
type TimezoneUtil struct {
	defaultTimezone string
}

// NewTimezoneUtil creates a new TimezoneUtil with default timezone
func NewTimezoneUtil(defaultTz string) *TimezoneUtil {
	if defaultTz == "" {
		defaultTz = "UTC"
	}
	return &TimezoneUtil{
		defaultTimezone: defaultTz,
	}
}

// ConvertTime converts time between timezones
func (tu *TimezoneUtil) ConvertTime(t time.Time, fromTz, toTz string) (time.Time, error) {
	// Load source timezone
	fromLoc, err := time.LoadLocation(fromTz)
	if err != nil {
		return t, fmt.Errorf("invalid source timezone '%s': %v", fromTz, err)
	}

	// Load target timezone
	toLoc, err := time.LoadLocation(toTz)
	if err != nil {
		return t, fmt.Errorf("invalid target timezone '%s': %v", toTz, err)
	}

	// Convert time to source timezone then to target timezone
	return t.In(fromLoc).In(toLoc), nil
}

// StandardizeToUTC converts time to UTC
func (tu *TimezoneUtil) StandardizeToUTC(t time.Time) time.Time {
	return t.UTC()
}

// LocalizeTime converts UTC time to user's timezone
func (tu *TimezoneUtil) LocalizeTime(t time.Time, userTz string) (time.Time, error) {
	if userTz == "" {
		userTz = tu.defaultTimezone
	}

	loc, err := time.LoadLocation(userTz)
	if err != nil {
		return t, fmt.Errorf("invalid timezone '%s': %v", userTz, err)
	}

	return t.In(loc), nil
}

// ParseTimeInTimezone parses time string in a specific timezone
func (tu *TimezoneUtil) ParseTimeInTimezone(timeStr, layout, timezone string) (time.Time, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timezone '%s': %v", timezone, err)
	}

	return time.ParseInLocation(layout, timeStr, loc)
}

// GuessTimezone attempts to guess timezone from timezone abbreviation
func (tu *TimezoneUtil) GuessTimezone(abbr string) string {
	abbr = strings.ToUpper(abbr)
	timezoneMap := map[string]string{
		"EST":  "America/New_York",
		"EDT":  "America/New_York",
		"CST":  "America/Chicago",
		"CDT":  "America/Chicago",
		"MST":  "America/Denver",
		"MDT":  "America/Denver",
		"PST":  "America/Los_Angeles",
		"PDT":  "America/Los_Angeles",
		"GMT":  "UTC",
		"UTC":  "UTC",
		"ICT":  "Asia/Bangkok",
		"JST":  "Asia/Tokyo",
		"IST":  "Asia/Kolkata",
		"AEST": "Australia/Sydney",
	}

	if tz, ok := timezoneMap[abbr]; ok {
		return tz
	}
	return tu.defaultTimezone
}

// AdjustTimeToWorkingHours adjusts time to fall within working hours
func (tu *TimezoneUtil) AdjustTimeToWorkingHours(t time.Time, workingHours *GoogleWorkingHours) time.Time {
	if workingHours == nil {
		return t
	}

	// Convert time to working hours timezone
	loc, err := time.LoadLocation(workingHours.TimeZone)
	if err != nil {
		return t
	}
	localTime := t.In(loc)

	// Find schedule for current day
	var schedule *GoogleWeeklySchedule
	for _, s := range workingHours.Schedule {
		if s.DayOfWeek == localTime.Weekday() {
			schedule = &s
			break
		}
	}

	if schedule == nil {
		return t
	}

	// If time is before working hours, move to start of working hours
	if localTime.Hour() < schedule.StartTime.Hour() ||
		(localTime.Hour() == schedule.StartTime.Hour() && localTime.Minute() < schedule.StartTime.Minute()) {
		return time.Date(
			localTime.Year(),
			localTime.Month(),
			localTime.Day(),
			schedule.StartTime.Hour(),
			schedule.StartTime.Minute(),
			0,
			0,
			loc,
		)
	}

	// If time is after working hours, move to next day's start
	if localTime.Hour() > schedule.EndTime.Hour() ||
		(localTime.Hour() == schedule.EndTime.Hour() && localTime.Minute() > schedule.EndTime.Minute()) {
		nextDay := localTime.AddDate(0, 0, 1)
		return time.Date(
			nextDay.Year(),
			nextDay.Month(),
			nextDay.Day(),
			schedule.StartTime.Hour(),
			schedule.StartTime.Minute(),
			0,
			0,
			loc,
		)
	}

	return localTime
}
