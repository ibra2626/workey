// Package clock provides an injectable time source anchored to Europe/Istanbul,
// plus a civil-date type and week boundaries (Monday start). The Istanbul
// location is loaded once at package init; time/tzdata is imported so the binary
// carries its own timezone database and does not depend on the host's.
package clock

import (
	"fmt"
	"time"

	_ "time/tzdata"
)

// Location is Europe/Istanbul, loaded once at package initialisation.
var Location = mustLoadIstanbul()

func mustLoadIstanbul() *time.Location {
	loc, err := time.LoadLocation("Europe/Istanbul")
	if err != nil {
		panic("clock: cannot load Europe/Istanbul: " + err.Error())
	}
	return loc
}

// Clock is the current-time source. Production uses System; tests inject Fake.
type Clock interface {
	Now() time.Time
}

// System is a Clock backed by the wall clock.
type System struct{}

// Now returns the current wall-clock time.
func (System) Now() time.Time { return time.Now() }

// Fake is a Clock that always returns a fixed time, for tests.
type Fake struct {
	Current time.Time
}

// Now returns the fixed time.
func (f Fake) Now() time.Time { return f.Current }

// Date is a civil date (year, month, day) with no time-of-day or zone. It is
// comparable, so values can be compared with ==.
type Date struct {
	Year  int
	Month time.Month
	Day   int
}

// String formats the date as YYYY-MM-DD.
func (d Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
}

// Today returns the civil date in Istanbul for the clock's current time.
func Today(c Clock) Date {
	return dateOf(c.Now().In(Location))
}

// WeekStart returns the Monday of the week containing d (as a civil date).
func WeekStart(d Date) Date {
	t := time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, Location)
	// time.Weekday has Sunday=0 .. Saturday=6; shift so Monday is the start.
	offset := (int(t.Weekday()) + 6) % 7
	return dateOf(t.AddDate(0, 0, -offset))
}

func dateOf(t time.Time) Date {
	return Date{Year: t.Year(), Month: t.Month(), Day: t.Day()}
}
