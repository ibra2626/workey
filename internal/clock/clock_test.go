package clock

import (
	"testing"
	"time"
)

func TestTodayIstanbul(t *testing.T) {
	// 2026-10-25 21:30 UTC is 2026-10-26 00:30 in Istanbul (UTC+3, no DST).
	c := Fake{Current: time.Date(2026, time.October, 25, 21, 30, 0, 0, time.UTC)}
	got := Today(c)
	want := Date{Year: 2026, Month: time.October, Day: 26}
	if got != want {
		t.Errorf("Today() = %v, want %v", got, want)
	}
}

func TestWeekStartMonday(t *testing.T) {
	// 2026-10-28 is a Wednesday; its week starts Monday 2026-10-26.
	got := WeekStart(Date{Year: 2026, Month: time.October, Day: 28})
	want := Date{Year: 2026, Month: time.October, Day: 26}
	if got != want {
		t.Errorf("WeekStart() = %v, want %v", got, want)
	}
}

func TestWeekStartOnMondayIsSameDay(t *testing.T) {
	d := Date{Year: 2026, Month: time.October, Day: 26} // Monday
	if got := WeekStart(d); got != d {
		t.Errorf("WeekStart(%v) = %v, want %v", d, got, d)
	}
}

func TestWeekStartOnSunday(t *testing.T) {
	// 2026-11-01 is a Sunday; its week starts Monday 2026-10-26.
	got := WeekStart(Date{Year: 2026, Month: time.November, Day: 1})
	want := Date{Year: 2026, Month: time.October, Day: 26}
	if got != want {
		t.Errorf("WeekStart() = %v, want %v", got, want)
	}
}

func TestDateString(t *testing.T) {
	got := Date{Year: 2027, Month: time.July, Day: 18}.String()
	if got != "2027-07-18" {
		t.Errorf("String() = %q, want 2027-07-18", got)
	}
}
