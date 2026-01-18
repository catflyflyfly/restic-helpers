package cron

import (
	"fmt"

	"github.com/robfig/cron/v3"
)

// CalendarInterval represents a launchd StartCalendarInterval entry.
// See: https://manpagez.com/man/5/launchd.plist/
type CalendarInterval struct {
	Minute  *int `plist:"Minute,omitempty"`
	Hour    *int `plist:"Hour,omitempty"`
	Day     *int `plist:"Day,omitempty"`
	Weekday *int `plist:"Weekday,omitempty"`
	Month   *int `plist:"Month,omitempty"`
}

func setMinute(ci *CalendarInterval, v int)  { ci.Minute = &v }
func setHour(ci *CalendarInterval, v int)    { ci.Hour = &v }
func setDay(ci *CalendarInterval, v int)     { ci.Day = &v }
func setWeekday(ci *CalendarInterval, v int) { ci.Weekday = &v }
func setMonth(ci *CalendarInterval, v int)   { ci.Month = &v }

const MaxScheduleEntries = 50

// ParseCron parses a cron expression and returns calendar intervals for launchd.
func ParseCron(expr string) ([]CalendarInterval, error) {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(expr)
	if err != nil {
		return nil, fmt.Errorf("invalid cron expression: %w", err)
	}

	// Type assert to access the parsed bit fields directly.
	// SpecSchedule stores each cron field as a uint64 bit set.
	spec, ok := schedule.(*cron.SpecSchedule)
	if !ok {
		return nil, fmt.Errorf("unexpected schedule type")
	}

	return expandSpec(spec)
}

// expandSpec expands a SpecSchedule into calendar intervals.
// It creates the cartesian product of all field combinations.
func expandSpec(spec *cron.SpecSchedule) ([]CalendarInterval, error) {
	// Extract which values are set in each bit field.
	// nil means wildcard (*), so we don't set that field.
	minutes := bitsToSlice(spec.Minute, 0, 59)
	hours := bitsToSlice(spec.Hour, 0, 23)
	days := bitsToSlice(spec.Dom, 1, 31)
	months := bitsToSlice(spec.Month, 1, 12)
	weekdays := bitsToSlice(spec.Dow, 0, 6)

	if minutes == nil && hours == nil && days == nil && months == nil && weekdays == nil {
		return nil, fmt.Errorf("cron expression is all wildcards")
	}

	// Calculate total combinations for the cartesian product.
	count := max(1, len(minutes)) * max(1, len(hours)) * max(1, len(days)) * max(1, len(months)) * max(1, len(weekdays))
	if count > MaxScheduleEntries {
		return nil, fmt.Errorf("cron expression expands to %d entries (max %d)", count, MaxScheduleEntries)
	}

	return intervalCombinations(minutes, hours, days, weekdays, months), nil
}

// bitsToSlice extracts set bit positions of uint64 bit field in cron.SpecSchedule.
// Returns nil if all bits in range are set (wildcard).
// e.g., bits=0b101, min=0, max=2 → [0, 2]
// e.g., bits=0b1111111, min=0, max=6 → nil (wildcard)
func bitsToSlice(bits uint64, min, max int) []int {
	var result []int
	for i := min; i <= max; i++ {
		if bits&(1<<uint(i)) != 0 {
			result = append(result, i)
		}
	}
	// Full range = wildcard
	if len(result) == max-min+1 {
		return nil
	}
	return result
}

// intervalCombinations builds the cartesian product of all field values.
// nil slices are treated as wildcards (field not set).
func intervalCombinations(minutes, hours, days, weekdays, months []int) []CalendarInterval {
	intervals := []CalendarInterval{{}}
	intervals = expandCombinations(intervals, minutes, setMinute)
	intervals = expandCombinations(intervals, hours, setHour)
	intervals = expandCombinations(intervals, days, setDay)
	intervals = expandCombinations(intervals, weekdays, setWeekday)
	intervals = expandCombinations(intervals, months, setMonth)
	return intervals
}

// expand multiplies intervals by values. If vals is nil (wildcard), returns as-is.
func expandCombinations(intervals []CalendarInterval, vals []int, setter func(*CalendarInterval, int)) []CalendarInterval {
	if vals == nil {
		return intervals
	}
	result := make([]CalendarInterval, 0, len(intervals)*len(vals))
	for _, interval := range intervals {
		for _, v := range vals {
			ci := interval
			setter(&ci, v)
			result = append(result, ci)
		}
	}
	return result
}
