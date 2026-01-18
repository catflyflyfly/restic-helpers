package cron

import (
	"testing"
)

func TestParseCronValid(t *testing.T) {
	tests := []struct {
		name  string
		expr  string
		count int
	}{
		{"daily at midnight", "0 0 * * *", 1},
		{"twice daily", "0 6,18 * * *", 2},
		{"weekdays at 9am", "0 9 * * 1-5", 5},
		{"specific day and time", "30 14 15 * *", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			intervals, err := ParseCron(tt.expr)
			if err != nil {
				t.Errorf("ParseCron() error = %v", err)
				return
			}
			if len(intervals) != tt.count {
				t.Errorf("ParseCron() returned %d intervals, want %d", len(intervals), tt.count)
			}
		})
	}
}

func TestParseCronInvalid(t *testing.T) {
	tests := []struct {
		name string
		expr string
	}{
		{"invalid expression", "invalid"},
		{"too many fields", "0 0 * * * *"},
		{"empty expression", ""},
		{"too many entries", "* * * * *"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseCron(tt.expr)
			if err == nil {
				t.Error("ParseCron() expected error, got nil")
			}
		})
	}
}

func TestBitsToSlice(t *testing.T) {
	tests := []struct {
		name string
		bits uint64
		min  int
		max  int
		want []int
	}{
		{"single bit 0", 0b1, 0, 7, []int{0}},
		{"single bit 3", 0b1000, 0, 7, []int{3}},
		{"bits 0 and 2", 0b101, 0, 7, []int{0, 2}},
		{"bits 1-3", 0b1110, 0, 7, []int{1, 2, 3}},
		{"full range (wildcard)", 0b1110, 1, 3, nil},
		{"no bits set (wildcard)", 0, 0, 7, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bitsToSlice(tt.bits, tt.min, tt.max)
			if len(got) != len(tt.want) {
				t.Errorf("bitsToSlice() = %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("bitsToSlice() = %v, want %v", got, tt.want)
					return
				}
			}
		})
	}
}

func TestIntervalCombinations(t *testing.T) {
	tests := []struct {
		name                                   string
		minutes, hours, days, weekdays, months []int
		want                                   []CalendarInterval
	}{
		{
			"all wildcards",
			nil, nil, nil, nil, nil,
			[]CalendarInterval{{}},
		},
		{
			"minute only",
			[]int{0, 30}, nil, nil, nil, nil,
			[]CalendarInterval{{Minute: intPtr(0)}, {Minute: intPtr(30)}},
		},
		{
			"hour only",
			nil, []int{9, 18}, nil, nil, nil,
			[]CalendarInterval{{Hour: intPtr(9)}, {Hour: intPtr(18)}},
		},
		{
			"weekday only",
			nil, nil, nil, []int{1, 5}, nil,
			[]CalendarInterval{{Weekday: intPtr(1)}, {Weekday: intPtr(5)}},
		},
		{
			"minute and hour",
			[]int{0, 30}, []int{9, 18}, nil, nil, nil,
			[]CalendarInterval{
				{Minute: intPtr(0), Hour: intPtr(9)},
				{Minute: intPtr(0), Hour: intPtr(18)},
				{Minute: intPtr(30), Hour: intPtr(9)},
				{Minute: intPtr(30), Hour: intPtr(18)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := intervalCombinations(tt.minutes, tt.hours, tt.days, tt.weekdays, tt.months)
			if len(got) != len(tt.want) {
				t.Errorf("length = %d, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if !equalInterval(got[i], tt.want[i]) {
					t.Errorf("got[%d] = %+v, want %+v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// intPtr is a test helper to create *int values for CalendarInterval comparisons.
func intPtr(i int) *int {
	return &i
}

func equalInterval(a, b CalendarInterval) bool {
	return ptrEqual(a.Minute, b.Minute) &&
		ptrEqual(a.Hour, b.Hour) &&
		ptrEqual(a.Day, b.Day) &&
		ptrEqual(a.Weekday, b.Weekday) &&
		ptrEqual(a.Month, b.Month)
}

func ptrEqual(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
