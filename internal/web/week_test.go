package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/ibra2626/workey/internal/clock"
	"github.com/ibra2626/workey/internal/plan"
	"github.com/ibra2626/workey/internal/store"
)

// sessionID returns the id of the first session on the given date.
func sessionID(t *testing.T, a *Auth, date clock.Date) int64 {
	t.Helper()
	day, err := a.svc.Day(context.Background(), date)
	if err != nil {
		t.Fatalf("Day %s: %v", date, err)
	}
	if len(day.Sessions) == 0 {
		t.Fatalf("no session on %s", date)
	}
	return day.Sessions[0].ID
}

// seedMovedWeek seeds the sample week, completes Monday and links Saturday's
// long block as the makeup for the missed Tuesday, reproducing the mockup state
// (today = Wednesday 28): 1 done, 1 moved to Saturday.
func seedMovedWeek(t *testing.T, st *store.Store, a *Auth) {
	t.Helper()
	ctx := context.Background()
	seedWeek(t, st)
	mon := sessionID(t, a, clock.Date{Year: 2026, Month: 10, Day: 26})
	if _, err := a.svc.Complete(ctx, plan.SourceWeb, mon); err != nil {
		t.Fatalf("Complete Monday: %v", err)
	}
	var week plan.SavePlanRequest
	loadFixture(t, "sample-week.json", &week)
	last := len(week.Sessions) - 1 // Saturday long block
	week.Sessions[last].MakeupOf = &plan.SlotRef{Date: "2026-10-27", StartTime: "20:30"}
	if _, err := a.svc.SavePlan(ctx, plan.SourceAPI, week); err != nil {
		t.Fatalf("SavePlan makeup: %v", err)
	}
}

func TestHaftaRendersEveryState(t *testing.T) {
	st := newStore(t)
	a := newAuthOn(t, st) // clock = Wednesday 28 Ekim 2026
	seedWeek(t, st)       // no completion, no makeup
	body := getAuthed(t, a, "/hafta")

	// Mon/Tue are past and undone → missed; Wed is today; Thu planned; Fri/Sun rest.
	for _, want := range []string{
		"Hafta 1",      // plan week in the header
		"26 Eki–1 Kas", // week range, en dash
		"Yapılamadı",   // missed (no makeup)
		"Bugün 20:30",  // today
		"Planlı",       // future weekday
		"Dinlenme",     // rest day
		"Uzun blok",    // Saturday block
	} {
		if !strings.Contains(body, want) {
			t.Errorf("Hafta body missing %q", want)
		}
	}
}

func TestHaftaMovedRowAndNoSeri(t *testing.T) {
	st := newStore(t)
	a := newAuthOn(t, st)
	seedMovedWeek(t, st, a)
	body := getAuthed(t, a, "/hafta")

	if !strings.Contains(body, "Cumartesiye taşındı") {
		t.Error("Hafta does not show the moved-to weekday text")
	}
	if !strings.Contains(body, "Tamamlandı") {
		t.Error("Hafta does not show the completed Monday row")
	}
	if strings.Contains(body, "Seri") {
		t.Error("Hafta desktop header still shows the removed Seri figure")
	}
}

func TestHaftaOtherWeekByDate(t *testing.T) {
	st := newStore(t)
	a := newAuthOn(t, st)
	seedWeek(t, st)
	// The week before the plan: no sessions, all rest.
	body := getAuthed(t, a, "/hafta?tarih=2026-10-19")
	if !strings.Contains(body, "Dinlenme") {
		t.Error("an empty week should render rest rows")
	}
	if strings.Contains(body, "Üslü ifadeler") {
		t.Error("the previous week should not show this week's sessions")
	}
}

func TestHaftaUndoPastRowViaHTMX(t *testing.T) {
	st := newStore(t)
	a := newAuthOn(t, st)
	seedMovedWeek(t, st, a)
	cookie := login(t, a)
	sid, _ := a.decodeSession(cookie.Value)
	mon := sessionID(t, a, clock.Date{Year: 2026, Month: 10, Day: 26})

	req := httptest.NewRequest(http.MethodPost, "/hafta/oturum/"+strconv.FormatInt(mon, 10)+"/geri-al", nil)
	req.AddCookie(cookie)
	req.Header.Set(csrfHeader, a.csrfToken(sid))
	req.Header.Set("HX-Request", "true")
	res := do(t, a, req)
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("undo past row status = %d, want 200", res.StatusCode)
	}
	b := readBody(res)
	if !strings.Contains(b, `id="hafta-main"`) {
		t.Error("undo response is not the Hafta main fragment")
	}
	if strings.Contains(b, "<!DOCTYPE") {
		t.Error("undo response should be a fragment, not the whole page")
	}
	// Monday is no longer done.
	day, _ := a.svc.Day(context.Background(), clock.Date{Year: 2026, Month: 10, Day: 26})
	if day.Sessions[0].Completed {
		t.Error("undo did not clear Monday's completion")
	}
}

func TestHaftaFutureRowRejected(t *testing.T) {
	st := newStore(t)
	a := newAuthOn(t, st)
	seedWeek(t, st)
	cookie := login(t, a)
	sid, _ := a.decodeSession(cookie.Value)
	thu := sessionID(t, a, clock.Date{Year: 2026, Month: 10, Day: 29}) // future

	req := httptest.NewRequest(http.MethodPost, "/hafta/oturum/"+strconv.FormatInt(thu, 10)+"/tamamla", nil)
	req.AddCookie(cookie)
	req.Header.Set(csrfHeader, a.csrfToken(sid))
	req.Header.Set("HX-Request", "true")
	res := do(t, a, req)
	defer res.Body.Close()
	if res.StatusCode < 400 || res.StatusCode >= 500 {
		t.Fatalf("future row action status = %d, want 4xx", res.StatusCode)
	}
}
