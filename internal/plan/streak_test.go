package plan

import (
	"testing"
	"time"

	"github.com/ibra2626/workey/internal/clock"
	"github.com/ibra2626/workey/internal/store"
)

// date builds an İstanbul civil date for the streak tests.
func date(y int, m time.Month, d int) clock.Date {
	return clock.Date{Year: y, Month: m, Day: d}
}

// sess builds a scheduled session on date d; done marks it completed.
func sess(d string, done bool) store.Session {
	s := store.Session{Date: d, DurationMin: 90}
	if done {
		ts := d + "T19:00:00Z"
		s.CompletedAt = &ts
	}
	return s
}

func TestComputeStreak(t *testing.T) {
	today := date(2026, time.October, 28) // Wednesday

	cases := []struct {
		name         string
		all          []store.Session
		wantStreak   int
		wantStudied  bool
		wantStudyDay bool
	}{
		{
			name:       "no sessions",
			all:        nil,
			wantStreak: 0,
		},
		{
			name:         "today and yesterday done",
			all:          []store.Session{sess("2026-10-27", true), sess("2026-10-28", true)},
			wantStreak:   2,
			wantStudied:  true,
			wantStudyDay: true,
		},
		{
			name:         "today open counts the standing run, not a miss",
			all:          []store.Session{sess("2026-10-26", true), sess("2026-10-27", true), sess("2026-10-28", false)},
			wantStreak:   2,
			wantStudied:  false,
			wantStudyDay: true,
		},
		{
			name:         "rest day is transparent",
			all:          []store.Session{sess("2026-10-26", true), sess("2026-10-28", true)}, // 27 is a rest day
			wantStreak:   2,
			wantStudied:  true,
			wantStudyDay: true,
		},
		{
			name:         "a missed study day ends the run",
			all:          []store.Session{sess("2026-10-27", false), sess("2026-10-28", true)},
			wantStreak:   1,
			wantStudied:  true,
			wantStudyDay: true,
		},
		{
			name:        "run standing before today, today after the plan",
			all:         []store.Session{sess("2026-10-26", true), sess("2026-10-27", true)},
			wantStreak:  2, // today (28) is a rest day: transparent, not a reset
			wantStudied: false,
		},
		{
			name:         "single open study day with no prior run",
			all:          []store.Session{sess("2026-10-28", false)},
			wantStreak:   0,
			wantStudied:  false,
			wantStudyDay: true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			streak, studied, studyDay := computeStreak(c.all, today)
			if streak != c.wantStreak {
				t.Errorf("streak = %d, want %d", streak, c.wantStreak)
			}
			if studied != c.wantStudied {
				t.Errorf("studiedToday = %v, want %v", studied, c.wantStudied)
			}
			if studyDay != c.wantStudyDay {
				t.Errorf("studyDayToday = %v, want %v", studyDay, c.wantStudyDay)
			}
		})
	}
}
