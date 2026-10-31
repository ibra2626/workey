package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ibra2626/workey/internal/clock"
	"github.com/ibra2626/workey/internal/plan"
	"github.com/ibra2626/workey/internal/store"
)

func studyPost(t *testing.T, a *Auth, cookie *http.Cookie, id int64, action, body string, hx bool) (int, string, http.Header) {
	t.Helper()
	r := httptest.NewRequest("POST", "/oturum/"+strconv.FormatInt(id, 10)+"/"+action, strings.NewReader(body))
	if cookie != nil {
		r.AddCookie(cookie)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if hx {
		r.Header.Set("HX-Request", "true")
	}
	res := do(t, a, r)
	defer res.Body.Close()
	return res.StatusCode, readBody(res), res.Header
}

func TestStudyFormsLifecycleAndIndependentSessions(t *testing.T) {
	ctx := context.Background()
	st := newStore(t)
	a := newAuthOn(t, st)
	seedWeek(t, st)
	first, second := login(t, a), login(t, a)
	sid, _ := a.decodeSession(first.Value)
	sid2, _ := a.decodeSession(second.Value)
	csrf := "_csrf=" + url.QueryEscape(a.csrfToken(sid)) + "&"
	csrf2 := "_csrf=" + url.QueryEscape(a.csrfToken(sid2)) + "&"
	id := todaySessionID(t, a)
	post := func(cookie *http.Cookie, action, body string, hx bool, want int, texts ...string) string {
		t.Helper()
		status, html, headers := studyPost(t, a, cookie, id, action, body, hx)
		if status != want {
			t.Fatalf("%s status %d: %s", action, status, html)
		}
		if want == 303 && !strings.HasPrefix(headers.Get("Location"), "/") {
			t.Fatal("missing redirect")
		}
		for _, text := range texts {
			if !strings.Contains(html, text) {
				t.Errorf("%s missing %q", action, text)
			}
		}
		return html
	}
	post(first, "baslat", csrf+"expected_study_version=0&entry_mode=ten_minutes", true, 200, "Başlangıç hedefi: 10 dakika", "Bugünün planı", "Çalışmayı kapat", "Burada bıraktım")
	if body := getAuthed(t, a, "/"); !strings.Contains(body, "Açık çalışma") {
		t.Fatal("refresh lost active study")
	}
	post(second, "birak", csrf2+"expected_study_version=0&pause_mode=bookmark&item_ordinal=0&stage=soru&resume_page=51", true, 409, `value="51"`, `name="expected_study_version" value="0"`, "başka bir sekmede")
	post(first, "birak", csrf+"expected_study_version=1&pause_mode=bookmark&item_ordinal=0&stage=konu&resume_page=999", true, 422, `value="999"`, "aşama aralığında", "Geçerli sayfa aralıkları")
	post(first, "birak", csrf+"expected_study_version=1&pause_mode=bookmark&item_ordinal=0&stage=konu&resume_page=abc", false, 422, `value="abc"`, "<!doctype html>")
	post(first, "birak", csrf+"expected_study_version=1&pause_mode=bookmark&item_ordinal=0&stage=soru&resume_page=58", false, 303)
	v, err := st.StudyBySession(ctx, st.DB(), id)
	if err != nil || v.Status != plan.StudyPaused || *v.ResumePage != 58 || v.Version != 2 {
		t.Fatalf("pause: %+v %v", v, err)
	}
	// Reopen the real disposable SQLite file, then read through a new service.
	var seq int
	var name, dbPath string
	if err := st.DB().QueryRowContext(ctx, "PRAGMA database_list").Scan(&seq, &name, &dbPath); err != nil {
		t.Fatal(err)
	}
	if err := st.Close(); err != nil {
		t.Fatal(err)
	}
	st, err = store.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	a.svc = plan.New(st, clock.Fake{Current: testNow})
	body := getAuthed(t, a, "/")
	for _, text := range []string{"Kaldığım yerden devam et", "s. 58 sayfasından devam et", "Kısmen çalışıldı"} {
		if !strings.Contains(body, text) {
			t.Errorf("reread missing %q", text)
		}
	}
	post(second, "baslat", csrf2+"expected_study_version=2&entry_mode=normal", false, 303)
	post(first, "birak", csrf+"expected_study_version=1&pause_mode=close", true, 409)
	post(second, "tamamla", csrf2+"expected_study_version=3", true, 200, "Tamamlandı", "90 / 540 dk")
	post(first, "birak", csrf+"expected_study_version=3&pause_mode=bookmark&item_ordinal=0&stage=soru&resume_page=57", true, 409)
	post(second, "geri-al", csrf2+"expected_study_version=4", true, 200, "Kaldığım yerden devam et", "0 / 540 dk")
	post(first, "baslat", csrf+"expected_study_version=5&entry_mode=normal", true, 200)
	post(first, "birak", csrf+"expected_study_version=6&pause_mode=close", true, 200, "Çalışma kapatıldı; devam yeri kaydedilmedi")
	v, err = st.StudyBySession(ctx, st.DB(), id)
	if err != nil || v.Version != 7 || v.Status != plan.StudyPaused || *v.ResumePage != 58 {
		t.Fatalf("close must preserve bookmark: %+v %v", v, err)
	}
}

func TestStudyBookmarkKindsAndInvalidSelections(t *testing.T) {
	for _, kind := range []string{plan.KindUzunBlok, plan.KindDeneme} {
		t.Run(kind, func(t *testing.T) {
			ctx := context.Background()
			st := newStore(t)
			a := newAuthOn(t, st)
			seedWeek(t, st)
			var fixture plan.SavePlanRequest
			loadFixture(t, "sample-week.json", &fixture)
			ss := fixture.Sessions[2]
			ss.Kind = kind
			ss.Items = append(ss.Items, fixture.Sessions[3].Items[0])
			if _, err := a.svc.SavePlan(ctx, plan.SourceAPI, plan.SavePlanRequest{From: ss.Date, To: ss.Date, Sessions: []plan.SessionInput{ss}}); err != nil {
				t.Fatal(err)
			}
			id := todaySessionID(t, a)
			cookie := login(t, a)
			sid, _ := a.decodeSession(cookie.Value)
			csrf := "_csrf=" + url.QueryEscape(a.csrfToken(sid)) + "&"
			if status, _, _ := studyPost(t, a, cookie, id, "baslat", csrf+"expected_study_version=0&entry_mode=normal", true); status != 200 {
				t.Fatal(status)
			}
			for _, fields := range []string{"item_ordinal=42&stage=calisma&resume_page=70", "item_ordinal=1&stage=konu&resume_page=70"} {
				status, body, _ := studyPost(t, a, cookie, id, "birak", csrf+"expected_study_version=1&pause_mode=bookmark&"+fields, true)
				if status != 422 || !strings.Contains(body, `value="70"`) {
					t.Fatalf("invalid selection %d %s", status, body)
				}
			}
			if status, body, _ := studyPost(t, a, cookie, id, "birak", csrf+"expected_study_version=1&pause_mode=bookmark&item_ordinal=1&stage=calisma&resume_page=76", true); status != 200 || !strings.Contains(body, "s. 76 sayfasından devam et") {
				t.Fatalf("last page %d %s", status, body)
			}
			v, err := st.StudyBySession(ctx, st.DB(), id)
			if err != nil || *v.ItemOrdinal != 1 || *v.ResumePage != 76 || *v.Stage != plan.StageCalisma {
				t.Fatalf("bookmark %+v %v", v, err)
			}
		})
	}
}

func TestStudyPeriodClosedHidesResumeWithoutExam(t *testing.T) {
	ctx := context.Background()
	st := newStore(t)
	a := newAuthOn(t, st)
	seedWeek(t, st)
	id := todaySessionID(t, a)
	if _, err := a.svc.StartStudy(ctx, plan.SourceWeb, id, plan.EntryNormal, 0, testExamDate); err != nil {
		t.Fatal(err)
	}
	if _, err := a.svc.PauseStudy(ctx, plan.SourceWeb, id, nil, 1); err != nil {
		t.Fatal(err)
	}
	a.examDate = clock.Date{}
	a.svc = plan.New(st, clock.Fake{Current: time.Date(2027, 7, 18, 9, 0, 0, 0, clock.Location)})
	body := getAuthed(t, a, "/")
	if !strings.Contains(body, "Kaldığın yer") || strings.Contains(body, `name="entry_mode"`) {
		t.Fatal("closed plan must hide resume but keep paused context")
	}
}

func TestStudyFormsAuthCSRFAndValidation(t *testing.T) {
	st := newStore(t)
	a := newAuthOn(t, st)
	seedWeek(t, st)
	id := todaySessionID(t, a)
	cookie := login(t, a)
	sid, _ := a.decodeSession(cookie.Value)
	csrf := "_csrf=" + url.QueryEscape(a.csrfToken(sid)) + "&"
	for _, action := range []string{"baslat", "birak"} {
		if status, _, _ := studyPost(t, a, nil, id, action, "", true); status != 401 {
			t.Errorf("unauth %s=%d", action, status)
		}
		if status, _, _ := studyPost(t, a, cookie, id, action, "", true); status != 403 {
			t.Errorf("csrf %s=%d", action, status)
		}
		for _, body := range []string{"", "expected_study_version=-1", "expected_study_version=abc", "expected_study_version=0&expected_study_version=1"} {
			if status, _, _ := studyPost(t, a, cookie, id, action, csrf+body, true); status != 400 {
				t.Errorf("version %s %s=%d", action, body, status)
			}
		}
	}
	if status, _, _ := studyPost(t, a, cookie, id, "baslat", csrf+"expected_study_version=0&entry_mode=unknown", true); status != 422 {
		t.Errorf("mode=%d", status)
	}
	day, _ := a.svc.Day(context.Background(), clock.Date{Year: 2026, Month: 10, Day: 29})
	if status, _, _ := studyPost(t, a, cookie, day.Sessions[0].ID, "baslat", csrf+"expected_study_version=0&entry_mode=normal", true); status != 422 {
		t.Errorf("future=%d", status)
	}
}

func TestStudyActiveSurvivesRestAndExamAndCannotSwitch(t *testing.T) {
	ctx := context.Background()
	st := newStore(t)
	a := newAuthOn(t, st)
	seedWeek(t, st)
	id := todaySessionID(t, a)
	if _, err := a.svc.StartStudy(ctx, plan.SourceWeb, id, plan.EntryNormal, 0, testExamDate); err != nil {
		t.Fatal(err)
	}
	day, _ := a.svc.Day(ctx, clock.Date{Year: 2026, Month: 10, Day: 27})
	other := day.Sessions[0]
	cookie := login(t, a)
	sid, _ := a.decodeSession(cookie.Value)
	csrf := "_csrf=" + url.QueryEscape(a.csrfToken(sid)) + "&"
	if status, _, _ := studyPost(t, a, cookie, other.ID, "baslat", csrf+"expected_study_version=0&entry_mode=normal", true); status != 409 {
		t.Fatalf("switch=%d", status)
	}
	for _, now := range []time.Time{time.Date(2026, 11, 1, 9, 0, 0, 0, clock.Location), time.Date(2027, 7, 19, 9, 0, 0, 0, clock.Location)} {
		a.svc = plan.New(st, clock.Fake{Current: now})
		body := getAuthed(t, a, "/")
		if !strings.Contains(body, "Açık çalışma") || !strings.Contains(body, "Çalışmayı kapat") || !strings.Contains(body, "Bitirdim") {
			t.Fatal("empty state hid active controls")
		}
	}
	if status, _, _ := studyPost(t, a, cookie, id, "birak", csrf+"expected_study_version=1&pause_mode=close", false); status != 200 {
		t.Fatalf("after exam close=%d", status)
	}
	body := getAuthed(t, a, "/")
	if !strings.Contains(body, "Kaldığın yer") || strings.Contains(body, `value="ten_minutes"`) {
		t.Fatal("paused after exam missing or start offered")
	}
}
