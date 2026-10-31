package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/ibra2626/workey/internal/plan"
)

func TestWeekPartialBookmarkSurvivesCompletionAndUndo(t *testing.T) {
	ctx := context.Background()
	st := newStore(t)
	a := newAuthOn(t, st)
	seedWeek(t, st)
	cookie := login(t, a)
	sid, _ := a.decodeSession(cookie.Value)
	rows, err := a.svc.Range(ctx, seedClock, seedClock)
	if err != nil || len(rows) != 1 {
		t.Fatalf("seed row: %v %v", rows, err)
	}
	id := rows[0].ID
	if _, err := a.svc.StartStudy(ctx, plan.SourceWeb, id, plan.EntryNormal, 0, testExamDate); err != nil {
		t.Fatal(err)
	}
	if _, err := a.svc.PauseStudy(ctx, plan.SourceWeb, id, &plan.Bookmark{ItemOrdinal: 0, Stage: plan.StageKonu, ResumePage: rows[0].Items[0].PageFrom}, 1); err != nil {
		t.Fatal(err)
	}
	check := func(body string, partial bool) {
		t.Helper()
		if strings.Contains(body, "Kısmen çalışıldı") != partial || !strings.Contains(body, "Kaldığın yer:") || !strings.Contains(body, "Devam edeceğin sayfa:") {
			t.Fatalf("partial=%t missing status/bookmark: %s", partial, body)
		}
	}
	check(getAuthed(t, a, "/hafta"), true)
	for i, action := range []string{"tamamla", "geri-al"} {
		r := httptest.NewRequest("POST", "/hafta/oturum/"+strconv.FormatInt(id, 10)+"/"+action, strings.NewReader("expected_study_version="+strconv.Itoa(i+2)))
		r.AddCookie(cookie)
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.Header.Set(csrfHeader, a.csrfToken(sid))
		r.Header.Set("HX-Request", "true")
		res := do(t, a, r)
		body := readBody(res)
		res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("%s: %d %s", action, res.StatusCode, body)
		}
		check(body, action == "geri-al")
	}
}

func TestWebCompletionStudyVersion(t *testing.T) {
	for _, prefix := range []string{"/oturum/", "/hafta/oturum/"} {
		t.Run(prefix, func(t *testing.T) {
			ctx := context.Background()
			st := newStore(t)
			a := newAuthOn(t, st)
			seedWeek(t, st)
			cookie := login(t, a)
			sid, _ := a.decodeSession(cookie.Value)
			id := todaySessionID(t, a)
			if _, err := a.svc.StartStudy(ctx, plan.SourceWeb, id, plan.EntryNormal, 0, testExamDate); err != nil {
				t.Fatal(err)
			}
			post := func(action, body string, status int) {
				t.Helper()
				req := httptest.NewRequest(http.MethodPost, prefix+strconv.FormatInt(id, 10)+"/"+action, strings.NewReader(body))
				req.AddCookie(cookie)
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				req.Header.Set(csrfHeader, a.csrfToken(sid))
				req.Header.Set("HX-Request", "true")
				res := do(t, a, req)
				defer res.Body.Close()
				if res.StatusCode != status {
					t.Fatalf("%s %s status=%d want=%d body=%s", action, body, res.StatusCode, status, readBody(res))
				}
			}
			for _, path := range []string{"/", "/hafta"} {
				body := getAuthed(t, a, path)
				if !strings.Contains(body, `name="expected_study_version" value="1"`) {
					t.Fatalf("%s form missing observed version", path)
				}
			}
			post("tamamla", "", http.StatusConflict) // cached old form
			post("geri-al", "", http.StatusConflict)
			post("tamamla", "expected_study_version=0", http.StatusConflict)
			for _, body := range []string{"expected_study_version=-1", "expected_study_version=no", "expected_study_version=", "expected_study_version=1&expected_study_version=2"} {
				post("tamamla", body, http.StatusBadRequest)
			}
			post("tamamla", "expected_study_version=1", http.StatusOK)
			post("geri-al", "", http.StatusConflict)
			post("geri-al", "expected_study_version=1", http.StatusConflict)
			for _, path := range []string{"/", "/hafta"} {
				if body := getAuthed(t, a, path); !strings.Contains(body, `name="expected_study_version" value="2"`) {
					t.Fatalf("%s undo form missing version", path)
				}
			}
			post("geri-al", "expected_study_version=2", http.StatusOK)
			v, err := st.StudyBySession(ctx, st.DB(), id)
			if err != nil || v.Version != 3 || v.Status != plan.StudyPaused {
				t.Fatalf("study: %+v %v", v, err)
			}
		})
	}
}
