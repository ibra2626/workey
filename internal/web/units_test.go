package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/ibra2626/workey/internal/clock"
	"github.com/ibra2626/workey/internal/plan"
)

// wedDate is the sample week's Wednesday (the web tests' "today"): its evening
// session references M1·3 with the Antrenman 1–3 unit range.
var wedDate = clock.Date{Year: 2026, Month: 10, Day: 28}

// wedSessionID returns the id of the Wednesday session.
func wedSessionID(t *testing.T, a *Auth) int64 {
	t.Helper()
	day, err := a.svc.Day(context.Background(), wedDate)
	if err != nil || len(day.Sessions) == 0 {
		t.Fatalf("Wednesday session: %v (sessions=%d)", err, len(day.Sessions))
	}
	return day.Sessions[0].ID
}

// m1Section3Units returns the units of M1 section 3 (Antrenman 1–3) with their ids.
func m1Section3Units(t *testing.T, a *Auth) []plan.UnitView {
	t.Helper()
	cat, err := a.svc.Sources(context.Background())
	if err != nil {
		t.Fatalf("Sources: %v", err)
	}
	for _, src := range cat.Sources {
		if src.Key != "M1" {
			continue
		}
		for _, sec := range src.Sections {
			if sec.Ordinal == 3 {
				return sec.Units
			}
		}
	}
	t.Fatal("M1 section 3 not found")
	return nil
}

func TestBugunRendersUnitControls(t *testing.T) {
	st := newStore(t)
	a := newAuthOn(t, st)
	seedWeek(t, st)
	sessionID := wedSessionID(t, a)
	units := m1Section3Units(t, a)

	body := getAuthed(t, a, "/")
	for _, want := range []string{
		"Birimler",          // the per-session unit list heading
		"M1·3, Antrenman 1", // exact per-unit notation (middle dot + section)
		"M1·3, Antrenman 3", // the whole 1–3 range is listed
		">Bitti<",           // the per-unit mark-done control
		"/oturum/" + strconv.FormatInt(sessionID, 10) + "/birim/" + strconv.FormatInt(units[0].ID, 10) + "/sonuc",
		`name="correct"`, `name="wrong"`, `name="blank"`, // D/Y/B inputs (book source)
	} {
		if !strings.Contains(body, want) {
			t.Errorf("Bugün body missing %q", want)
		}
	}
}

func TestUnitMarkingIndependentOfSession(t *testing.T) {
	ctx := context.Background()
	st := newStore(t)
	a := newAuthOn(t, st)
	seedWeek(t, st)
	sessionID := wedSessionID(t, a)
	units := m1Section3Units(t, a)
	cookie := login(t, a)
	sid, _ := a.decodeSession(cookie.Value)

	// Mark one unit done from Bugün (htmx). The session must NOT complete.
	form := url.Values{}
	form.Set("_csrf", a.csrfToken(sid))
	form.Set("done", "true")
	form.Set("correct", "4")
	form.Set("wrong", "1")
	form.Set("blank", "0")
	req := httptest.NewRequest(http.MethodPost,
		"/oturum/"+strconv.FormatInt(sessionID, 10)+"/birim/"+strconv.FormatInt(units[0].ID, 10)+"/sonuc",
		strings.NewReader(form.Encode()))
	req.AddCookie(cookie)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	res := do(t, a, req)
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("unit mark status = %d, want 200", res.StatusCode)
	}

	results, err := a.svc.UnitResults(ctx, []int64{units[0].ID})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := results[units[0].ID]; !ok {
		t.Error("unit not recorded done after marking")
	}
	day, _ := a.svc.Day(ctx, wedDate)
	if day.Sessions[0].State == plan.StateDone || day.Sessions[0].Completed {
		t.Error("marking a unit done wrongly completed the session")
	}

	// Conversely: completing the session must NOT mark the other units done.
	cform := url.Values{}
	cform.Set("_csrf", a.csrfToken(sid))
	creq := httptest.NewRequest(http.MethodPost,
		"/oturum/"+strconv.FormatInt(sessionID, 10)+"/tamamla", strings.NewReader(cform.Encode()))
	creq.AddCookie(cookie)
	creq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	creq.Header.Set("HX-Request", "true")
	cres := do(t, a, creq)
	cres.Body.Close()
	if cres.StatusCode != http.StatusOK {
		t.Fatalf("complete status = %d, want 200", cres.StatusCode)
	}
	all := make([]int64, len(units))
	for i, u := range units {
		all[i] = u.ID
	}
	results, _ = a.svc.UnitResults(ctx, all)
	if len(results) != 1 {
		t.Errorf("completing the session changed unit results: got %d done, want 1", len(results))
	}
}

func TestUnitMarkingNoJSAndCSRF(t *testing.T) {
	st := newStore(t)
	a := newAuthOn(t, st)
	seedWeek(t, st)
	sessionID := wedSessionID(t, a)
	units := m1Section3Units(t, a)
	cookie := login(t, a)
	sid, _ := a.decodeSession(cookie.Value)
	unitPath := "/oturum/" + strconv.FormatInt(sessionID, 10) + "/birim/" + strconv.FormatInt(units[0].ID, 10) + "/sonuc"

	// No CSRF token: forbidden, nothing written.
	bad := url.Values{}
	bad.Set("done", "true")
	req := httptest.NewRequest(http.MethodPost, unitPath, strings.NewReader(bad.Encode()))
	req.AddCookie(cookie)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := do(t, a, req)
	res.Body.Close()
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("no-CSRF status = %d, want 403", res.StatusCode)
	}
	if got, _ := a.svc.UnitResults(context.Background(), []int64{units[0].ID}); len(got) != 0 {
		t.Error("unit written without CSRF")
	}

	// No-JS form POST (hidden token, no HX-Request) redirects back to Bugün.
	good := url.Values{}
	good.Set("_csrf", a.csrfToken(sid))
	good.Set("done", "true")
	req2 := httptest.NewRequest(http.MethodPost, unitPath, strings.NewReader(good.Encode()))
	req2.AddCookie(cookie)
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res2 := do(t, a, req2)
	res2.Body.Close()
	if res2.StatusCode != http.StatusSeeOther || res2.Header.Get("Location") != "/" {
		t.Fatalf("no-JS unit status = %d loc %q, want 303 /", res2.StatusCode, res2.Header.Get("Location"))
	}
}

func TestHaftaUnitMarkingNoJSRedirectsToWeek(t *testing.T) {
	st := newStore(t)
	a := newAuthOn(t, st)
	seedWeek(t, st)
	sessionID := wedSessionID(t, a)
	units := m1Section3Units(t, a)
	cookie := login(t, a)
	sid, _ := a.decodeSession(cookie.Value)

	body := getAuthed(t, a, "/hafta")
	if !strings.Contains(body, "Birimler") || !strings.Contains(body, "M1·3, Antrenman 1") {
		t.Error("Hafta body missing unit controls")
	}

	form := url.Values{}
	form.Set("_csrf", a.csrfToken(sid))
	form.Set("done", "true")
	path := "/hafta/oturum/" + strconv.FormatInt(sessionID, 10) + "/birim/" +
		strconv.FormatInt(units[0].ID, 10) + "/sonuc?tarih=2026-10-26"
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.AddCookie(cookie)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := do(t, a, req)
	res.Body.Close()
	if res.StatusCode != http.StatusSeeOther || res.Header.Get("Location") != "/hafta?tarih=2026-10-26" {
		t.Fatalf("Hafta no-JS unit status = %d loc %q, want 303 /hafta?tarih=2026-10-26",
			res.StatusCode, res.Header.Get("Location"))
	}
}

func TestProgressShowsUnitAggregates(t *testing.T) {
	ctx := context.Background()
	st := newStore(t)
	a := newAuthOn(t, st)
	seedWeek(t, st)
	units := m1Section3Units(t, a)
	if _, err := a.svc.SetUnitResult(ctx, plan.SourceWeb, plan.UnitResultRequest{
		UnitID: units[0].ID, Done: true, Correct: ip(7), Wrong: ip(2), Blank: ip(1),
	}); err != nil {
		t.Fatalf("SetUnitResult: %v", err)
	}

	body := getAuthed(t, a, "/ilerleme")
	for _, want := range []string{
		"Birim ilerlemesi",
		"/ " + strconv.Itoa(len(units)) + " birim", // honest denominator = catalog units (3)
		"Doğru", "Yanlış", "Boş", "Çözülen",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("İlerleme body missing %q", want)
		}
	}
	// Done count 1 and answered 10 (7+2+1) must appear as figures.
	for _, want := range []string{">1</span>", ">10</dd>"} {
		if !strings.Contains(body, want) {
			t.Errorf("İlerleme figure missing %q", want)
		}
	}
}
