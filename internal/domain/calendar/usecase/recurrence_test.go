package usecase

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseRecurrenceRule(t *testing.T) {
	tests := []struct {
		name          string
		rule          string
		expectedError bool
		expectedRule  *RecurrenceRule
	}{
		{
			name:          "empty rule",
			rule:          "",
			expectedError: true,
		},
		{
			name:          "invalid format",
			rule:          "invalid",
			expectedError: true,
		},
		{
			name: "daily recurrence",
			rule: "RRULE:FREQ=DAILY",
			expectedRule: &RecurrenceRule{
				Frequency: FreqDaily,
				Interval:  1,
			},
		},
		{
			name: "weekly on Monday and Wednesday",
			rule: "RRULE:FREQ=WEEKLY;BYDAY=MO,WE",
			expectedRule: &RecurrenceRule{
				Frequency: FreqWeekly,
				Interval:  1,
				ByDay:     []Weekday{Monday, Wednesday},
			},
		},
		{
			name: "monthly on the 15th",
			rule: "RRULE:FREQ=MONTHLY;BYMONTHDAY=15",
			expectedRule: &RecurrenceRule{
				Frequency:  FreqMonthly,
				Interval:   1,
				ByMonthDay: []int{15},
			},
		},
		{
			name: "yearly in June and July",
			rule: "RRULE:FREQ=YEARLY;BYMONTH=6,7",
			expectedRule: &RecurrenceRule{
				Frequency: FreqYearly,
				Interval:  1,
				ByMonth:   []time.Month{time.June, time.July},
			},
		},
		{
			name: "complex rule with count",
			rule: "RRULE:FREQ=WEEKLY;COUNT=10;BYDAY=MO,WE,FR",
			expectedRule: &RecurrenceRule{
				Frequency: FreqWeekly,
				Count:     intPtr(10),
				Interval:  1,
				ByDay:     []Weekday{Monday, Wednesday, Friday},
			},
		},
		{
			name: "rule with interval",
			rule: "RRULE:FREQ=DAILY;INTERVAL=2",
			expectedRule: &RecurrenceRule{
				Frequency: FreqDaily,
				Interval:  2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule, err := ParseRecurrenceRule(tt.rule)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedRule.Frequency, rule.Frequency)
			assert.Equal(t, tt.expectedRule.Interval, rule.Interval)

			if tt.expectedRule.Count != nil {
				assert.Equal(t, *tt.expectedRule.Count, *rule.Count)
			}
			if len(tt.expectedRule.ByDay) > 0 {
				assert.Equal(t, tt.expectedRule.ByDay, rule.ByDay)
			}
			if len(tt.expectedRule.ByMonthDay) > 0 {
				assert.Equal(t, tt.expectedRule.ByMonthDay, rule.ByMonthDay)
			}
			if len(tt.expectedRule.ByMonth) > 0 {
				assert.Equal(t, tt.expectedRule.ByMonth, rule.ByMonth)
			}
		})
	}
}

func TestGetRecurrences(t *testing.T) {
	baseTime := time.Date(2024, 2, 1, 10, 0, 0, 0, time.UTC)
	oneHour := time.Hour

	tests := []struct {
		name            string
		rule            *RecurrenceRule
		start           time.Time
		end             time.Time
		duration        time.Duration
		expectedCount   int
		expectedPattern func(time.Time) bool
	}{
		{
			name: "daily for 3 days",
			rule: &RecurrenceRule{
				Frequency: FreqDaily,
				Count:     intPtr(3),
				Interval:  1,
			},
			start:         baseTime,
			end:           baseTime.AddDate(0, 0, 10), // End date doesn't matter when Count is set
			duration:      oneHour,
			expectedCount: 3,
			expectedPattern: func(t time.Time) bool {
				return t.Hour() == 10 && t.Minute() == 0
			},
		},
		{
			name: "weekly on Monday for 2 weeks",
			rule: &RecurrenceRule{
				Frequency: FreqWeekly,
				Count:     intPtr(2),
				Interval:  1,
				ByDay:     []Weekday{Monday},
			},
			start:         baseTime,
			end:           baseTime.AddDate(0, 0, 14),
			duration:      oneHour,
			expectedCount: 2,
			expectedPattern: func(t time.Time) bool {
				return t.Weekday() == time.Monday
			},
		},
		{
			name: "monthly on 15th for 2 months",
			rule: &RecurrenceRule{
				Frequency:  FreqMonthly,
				Count:      intPtr(2),
				Interval:   1,
				ByMonthDay: []int{15},
			},
			start:         baseTime,
			end:           baseTime.AddDate(0, 2, 0),
			duration:      oneHour,
			expectedCount: 2,
			expectedPattern: func(t time.Time) bool {
				return t.Day() == 15
			},
		},
		{
			name: "yearly in June and July for 2 years",
			rule: &RecurrenceRule{
				Frequency: FreqYearly,
				Count:     intPtr(4),
				Interval:  1,
				ByMonth:   []time.Month{time.June, time.July},
			},
			start:         baseTime,
			end:           baseTime.AddDate(2, 0, 0),
			duration:      oneHour,
			expectedCount: 4,
			expectedPattern: func(t time.Time) bool {
				return t.Month() == time.June || t.Month() == time.July
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slots := tt.rule.GetRecurrences(tt.start, tt.end, tt.duration)
			assert.Equal(t, tt.expectedCount, len(slots))

			if tt.expectedPattern != nil {
				for _, slot := range slots {
					assert.True(t, tt.expectedPattern(slot.Start))
					assert.Equal(t, tt.duration, slot.End.Sub(slot.Start))
				}
			}
		})
	}
}

func intPtr(i int) *int {
	return &i
}
