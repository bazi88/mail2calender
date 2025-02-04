package usecase

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseRecurrenceRule(t *testing.T) {
	tests := []struct {
		name     string
		rule     string
		expected *RecurrenceRule
		wantErr  bool
	}{
		{
			name: "daily recurrence",
			rule: "RRULE:FREQ=DAILY;COUNT=5",
			expected: &RecurrenceRule{
				Frequency: FreqDaily,
				Count:     intPtr(5),
				Interval:  1,
			},
		},
		{
			name: "weekly on Monday and Wednesday",
			rule: "RRULE:FREQ=WEEKLY;BYDAY=MO,WE",
			expected: &RecurrenceRule{
				Frequency: FreqWeekly,
				Interval:  1,
				ByDay:     []Weekday{Monday, Wednesday},
			},
		},
		{
			name: "monthly on the 15th",
			rule: "RRULE:FREQ=MONTHLY;BYMONTHDAY=15",
			expected: &RecurrenceRule{
				Frequency:  FreqMonthly,
				Interval:   1,
				ByMonthDay: []int{15},
			},
		},
		{
			name: "yearly in June and July",
			rule: "RRULE:FREQ=YEARLY;BYMONTH=6,7",
			expected: &RecurrenceRule{
				Frequency: FreqYearly,
				Interval:  1,
				ByMonth:   []time.Month{time.June, time.July},
			},
		},
		{
			name:    "invalid format",
			rule:    "NOT_A_RULE",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule, err := ParseRecurrenceRule(tt.rule)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected.Frequency, rule.Frequency)
			assert.Equal(t, tt.expected.Interval, rule.Interval)

			if tt.expected.Count != nil {
				assert.Equal(t, *tt.expected.Count, *rule.Count)
			}
			if len(tt.expected.ByDay) > 0 {
				assert.Equal(t, tt.expected.ByDay, rule.ByDay)
			}
			if len(tt.expected.ByMonthDay) > 0 {
				assert.Equal(t, tt.expected.ByMonthDay, rule.ByMonthDay)
			}
			if len(tt.expected.ByMonth) > 0 {
				assert.Equal(t, tt.expected.ByMonth, rule.ByMonth)
			}
		})
	}
}

func TestRecurrenceRule_GetRecurrences(t *testing.T) {
	loc, _ := time.LoadLocation("UTC")
	now := time.Date(2025, 1, 1, 10, 0, 0, 0, loc)
	oneHour := time.Hour

	tests := []struct {
		name          string
		rule          *RecurrenceRule
		start         time.Time
		end           time.Time
		duration      time.Duration
		expectedCount int
		expectedDates []time.Time
	}{
		{
			name: "daily for 5 days",
			rule: &RecurrenceRule{
				Frequency: FreqDaily,
				Count:     intPtr(5),
				Interval:  1,
			},
			start:         now,
			end:           now.AddDate(0, 0, 10),
			duration:      oneHour,
			expectedCount: 5,
			expectedDates: []time.Time{
				now,
				now.AddDate(0, 0, 1),
				now.AddDate(0, 0, 2),
				now.AddDate(0, 0, 3),
				now.AddDate(0, 0, 4),
			},
		},
		{
			name: "weekly on Monday and Wednesday",
			rule: &RecurrenceRule{
				Frequency: FreqWeekly,
				Interval:  1,
				ByDay:     []Weekday{Monday, Wednesday},
			},
			start:         now,
			end:           now.AddDate(0, 0, 14),
			duration:      oneHour,
			expectedCount: 4, // 2 weeks * 2 days per week
		},
		{
			name: "monthly on the 15th",
			rule: &RecurrenceRule{
				Frequency:  FreqMonthly,
				Interval:   1,
				ByMonthDay: []int{15},
			},
			start:         now,
			end:           now.AddDate(0, 3, 0),
			duration:      oneHour,
			expectedCount: 3, // 3 months * 1 day per month
		},
		{
			name: "yearly in June",
			rule: &RecurrenceRule{
				Frequency: FreqYearly,
				ByMonth:   []time.Month{time.June},
			},
			start:         now,
			end:           now.AddDate(2, 0, 0),
			duration:      oneHour,
			expectedCount: 2, // 2 years * 1 month per year
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slots := tt.rule.GetRecurrences(tt.start, tt.end, tt.duration)
			assert.Equal(t, tt.expectedCount, len(slots))

			if len(tt.expectedDates) > 0 {
				for i, expected := range tt.expectedDates {
					assert.Equal(t, expected.Year(), slots[i].Start.Year())
					assert.Equal(t, expected.Month(), slots[i].Start.Month())
					assert.Equal(t, expected.Day(), slots[i].Start.Day())
					assert.Equal(t, expected.Hour(), slots[i].Start.Hour())
					assert.Equal(t, expected.Minute(), slots[i].Start.Minute())
				}
			}

			// Check that all slots have the correct duration
			for _, slot := range slots {
				assert.Equal(t, tt.duration, slot.End.Sub(slot.Start))
			}
		})
	}
}

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}
