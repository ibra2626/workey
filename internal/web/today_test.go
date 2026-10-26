package web

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/ibra2626/workey/internal/clock"
	"github.com/ibra2626/workey/internal/plan"
)

// seedWeek loads the sample outline and week fixtures into the service, so the
// protected screens render the mockup state (today = Wednesday 28 Ekim 2026).
func seedWeek(t *testing.T, svc *plan.Service) {
	t.Helper()
	ctx := context.Background()
	var outline plan.SaveOutlineRequest
	loadFixture(t, "sample-outline.json", &outline)
	if _, err := svc.SaveOutline(ctx, plan.SourceAPI, outline); err != nil {
		t.Fatalf("SaveOutline: %v", err)
	}
	var week plan.SavePlanRequest
	loadFixture(t, "sample-week.json", &week)
	if _, err := svc.SavePlan(ctx, plan.SourceAPI, week); err != nil {
		t.Fatalf("SavePlan: %v", err)
	}
}

func loadFixture(t *testing.T, name string, v any) {
	t.Helper()
	buf, err := os.ReadFile(filepath.Join("..", "..", "testdata", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	if err := json.Unmarshal(buf, v); err != nil {
		t.Fatalf("decode fixture %s: %v", name, err)
	}
}

// getAuthed performs an authenticated GET and returns the response body.
func getAuthed(t *testing.T, a *Auth, path string) string {
	t.Helper()
	cookie := login(t, a)
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.AddCookie(cookie)
	res := do(t, a, req)
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("GET %s status = %d, want 200", path, res.StatusCode)
	}
	b, _ := io.ReadAll(res.Body)
	return string(b)
}

func TestBugunRendersTonightsSession(t *testing.T) {
	a := newTestAuth(t)
	seedWeek(t, a.svc)
	body := getAuthed(t, a, "/")

	// Turkish date heading and the exact book notation and page range.
	for _, want := range []string{
		"Çarşamba, 28 Ekim", // Turkish day heading
		"M1·3",              // book·chapter, middle dot U+00B7
		"Üslü ifadeler",     // chapter title
		"s. 41–58",          // page range with en dash U+2013
		"s. 41–44",          // konu split
		"Antr. 1–3",         // antrenman range
		"Tamamlandı",        // the primary action
	} {
		if !strings.Contains(body, want) {
			t.Errorf("Bugün body missing %q", want)
		}
	}
}

func TestBugunHasSeansTileAndNoSeri(t *testing.T) {
	a := newTestAuth(t)
	seedWeek(t, a.svc)
	body := getAuthed(t, a, "/")

	if !strings.Contains(body, "Seans") {
		t.Error("Bugün is missing the Seans stat tile")
	}
	if strings.Contains(body, "Seri") {
		t.Error("Bugün still shows the removed Seri tile (DECISIONS: no Seri)")
	}
}

func TestBugunMarksActiveNav(t *testing.T) {
	a := newTestAuth(t)
	seedWeek(t, a.svc)
	body := getAuthed(t, a, "/")
	if !strings.Contains(body, `aria-current="page"`) {
		t.Error("Bugün does not mark the active nav item with aria-current")
	}
}

func TestBugunEmptyStateWithoutPlan(t *testing.T) {
	a := newTestAuth(t) // no seed
	body := getAuthed(t, a, "/")
	if !strings.Contains(body, "Plan henüz yazılmadı.") {
		t.Error("empty Bugün does not show the no-plan state")
	}
}

// todaySessionID returns the id of today's (Wednesday) session.
func todaySessionID(t *testing.T, a *Auth) int64 {
	t.Helper()
	day, err := a.svc.Day(context.Background(), clock.Date{Year: 2026, Month: 10, Day: 28})
	if err != nil {
		t.Fatalf("Day: %v", err)
	}
	if len(day.Sessions) == 0 {
		t.Fatal("no session on Wednesday")
	}
	return day.Sessions[0].ID
}

func TestCompleteViaHTMXSwapsPanel(t *testing.T) {
	a := newTestAuth(t)
	seedWeek(t, a.svc)
	cookie := login(t, a)
	sid, _ := a.decodeSession(cookie.Value)
	id := todaySessionID(t, a)

	req := httptest.NewRequest(http.MethodPost, "/oturum/"+strconv.FormatInt(id, 10)+"/tamamla", nil)
	req.AddCookie(cookie)
	req.Header.Set(csrfHeader, a.csrfToken(sid))
	req.Header.Set("HX-Request", "true")
	res := do(t, a, req)
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("complete via htmx status = %d, want 200", res.StatusCode)
	}
	b, _ := io.ReadAll(res.Body)
	body := string(b)
	// The swapped fragment is just the panel (not the whole page), in the done
	// state with the quiet undo link.
	if !strings.Contains(body, `id="bugun-panel"`) {
		t.Error("swap response is not the first panel")
	}
	if strings.Contains(body, "<nav") || strings.Contains(body, "<!DOCTYPE") {
		t.Error("swap response should be the panel fragment, not the full page")
	}
	if !strings.Contains(body, "Geri al") {
		t.Error("completed panel is missing the Geri al link")
	}
}

func TestCompleteWithoutCSRFIsForbidden(t *testing.T) {
	a := newTestAuth(t)
	seedWeek(t, a.svc)
	cookie := login(t, a)
	id := todaySessionID(t, a)

	req := httptest.NewRequest(http.MethodPost, "/oturum/"+strconv.FormatInt(id, 10)+"/tamamla", nil)
	req.AddCookie(cookie)
	req.Header.Set("HX-Request", "true") // no CSRF token
	res := do(t, a, req)
	defer res.Body.Close()
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("complete without CSRF status = %d, want 403", res.StatusCode)
	}
}
