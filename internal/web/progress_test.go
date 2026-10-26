package web

import (
	"strings"
	"testing"
)

func TestIlerlemeRendersProgress(t *testing.T) {
	st := newStore(t)
	a := newAuthOn(t, st)
	seedMovedWeek(t, st, a) // Monday done → 90 dk = 1,5 sa
	body := getAuthed(t, a, "/ilerleme")

	for _, want := range []string{
		"/ 38 hafta",  // plan length from the saved plan, not hard-coded
		"1,5",         // done hours with a decimal comma (90 min)
		" soru",       // DGS weight, never "s."
		"Temel kavramlar",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("İlerleme body missing %q", want)
		}
	}
}

func TestIlerlemeZeroPercentIsLichenInk(t *testing.T) {
	st := newStore(t)
	a := newAuthOn(t, st)
	seedMovedWeek(t, st, a)
	body := getAuthed(t, a, "/ilerleme")
	// A topic group with no completed minutes shows its percentage in Lichen Ink.
	if !strings.Contains(body, `text-ink-3">%0`) {
		t.Error("a zero percentage is not rendered in Lichen Ink")
	}
}

func TestIlerlemeHasNoDenemeOrSeviye(t *testing.T) {
	st := newStore(t)
	a := newAuthOn(t, st)
	seedMovedWeek(t, st, a)
	body := strings.ToLower(getAuthed(t, a, "/ilerleme"))
	for _, forbidden := range []string{"deneme", "seviye"} {
		if strings.Contains(body, forbidden) {
			t.Errorf("İlerleme still contains the removed %q panel", forbidden)
		}
	}
}

func TestIlerlemePartialPlanRenders(t *testing.T) {
	st := newStore(t)
	a := newAuthOn(t, st)
	seedWeek(t, st) // only week 1 saved, nothing completed
	body := getAuthed(t, a, "/ilerleme")
	if !strings.Contains(body, "/ 38 hafta") {
		t.Error("partial plan did not render the 38-week frame")
	}
	// Denominators come from the saved plan, never the mockup's 342/190.
	if strings.Contains(body, "342") || strings.Contains(body, "190") {
		t.Error("İlerleme shows a hard-coded denominator (342/190)")
	}
}

func TestIlerlemeEmptyStateWithoutPlan(t *testing.T) {
	a := newTestAuth(t) // no seed
	body := getAuthed(t, a, "/ilerleme")
	if !strings.Contains(body, "Plan henüz yazılmadı.") {
		t.Error("empty İlerleme does not show the no-plan state")
	}
}
