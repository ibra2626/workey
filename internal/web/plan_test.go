package web

import (
	"strings"
	"testing"
)

func TestPlanRendersWholePlanAsCalendar(t *testing.T) {
	st := newStore(t)
	a := newAuthOn(t, st)
	seedWeek(t, st)
	body := getAuthed(t, a, "/plan")

	for _, want := range []string{
		"İşlem temeli",                      // phase name without its "Faz N" prefix
		"26 Eki–13 Ara · H1–H7",             // phase span and plan weeks
		"Ekim 2026",                         // month calendar heading
		`aria-label="Aylar"`,                // month strip
		`aria-describedby="gun-2026-10-27"`, // a session day points at its detail
		"Salı, 27 Ekim",                     // detail heading
		"Sayı basamakları",                  // chapter title of a session
		"M1·2, s.",                          // book reference
		"Konu: Temel kavramlar",             // topic group of a session
		`href="/plan"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("Plan body missing %q", want)
		}
	}
}

func TestPlanShowsOneMonthChosenByQuery(t *testing.T) {
	st := newStore(t)
	a := newAuthOn(t, st)
	seedWeek(t, st)
	body := getAuthed(t, a, "/plan?ay=2026-12")
	if !strings.Contains(body, `id="ay-2026-12"`) || strings.Contains(body, `id="ay-2026-10"`) {
		t.Error("?ay= did not show exactly the chosen month")
	}
	if !strings.Contains(body, `href="/plan?ay=2026-11"`) || !strings.Contains(body, "Bu aya dön") {
		t.Error("month navigation links are missing")
	}
	if strings.Contains(body, "gun-2026-10-27") {
		t.Error("days of another month are rendered")
	}
}

func TestPlanEmptyStateWithoutPlan(t *testing.T) {
	a := newTestAuth(t)
	body := getAuthed(t, a, "/plan")
	if !strings.Contains(body, "Plan henüz yazılmadı.") {
		t.Error("empty Plan does not show the no-plan state")
	}
}
