package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/ibra2626/workey/internal/api"
	"github.com/ibra2626/workey/internal/clock"
	"github.com/ibra2626/workey/internal/plan"
	"github.com/ibra2626/workey/internal/store"
)

const testToken = "test-token-123"

// newAPI builds an /api/v1 handler over a migrated temp store whose clock is
// fixed at 2026-10-26 09:00 Istanbul (so the sample week is all in the future
// and fully creatable).
func newAPI(t *testing.T) (http.Handler, *store.Store) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "workey.db")
	st, err := store.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	if err := st.Migrate(context.Background()); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	now, err := time.ParseInLocation("2006-01-02 15:04", "2026-10-26 09:00", clock.Location)
	if err != nil {
		t.Fatalf("parse time: %v", err)
	}
	svc := plan.New(st, clock.Fake{Current: now})
	h := api.New(svc, clock.Date{Year: 2027, Month: 7, Day: 18}, testToken)
	return h, st
}

// fixture reads a testdata file as raw bytes for use as a request body.
func fixture(t *testing.T, name string) []byte {
	t.Helper()
	buf, err := os.ReadFile(filepath.Join("..", "..", "testdata", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return buf
}

// do sends an authenticated request and returns the recorder.
func do(t *testing.T, h http.Handler, method, target string, body []byte, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	var r *http.Request
	if body != nil {
		r = httptest.NewRequest(method, target, bytes.NewReader(body))
	} else {
		r = httptest.NewRequest(method, target, nil)
	}
	r.Header.Set("Authorization", "Bearer "+testToken)
	for k, v := range headers {
		r.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	return rec
}

// seed loads the sample outline then the sample week through the API.
func seed(t *testing.T, h http.Handler) {
	t.Helper()
	if rec := do(t, h, http.MethodPut, "/api/v1/outline", fixture(t, "sample-outline.json"), nil); rec.Code != http.StatusOK {
		t.Fatalf("seed outline: status %d, body %s", rec.Code, rec.Body)
	}
	if rec := do(t, h, http.MethodPut, "/api/v1/plan", fixture(t, "sample-week.json"), nil); rec.Code != http.StatusOK {
		t.Fatalf("seed plan: status %d, body %s", rec.Code, rec.Body)
	}
}

func decode(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.Unmarshal(rec.Body.Bytes(), v); err != nil {
		t.Fatalf("decode body %q: %v", rec.Body.String(), err)
	}
}

func TestAuth(t *testing.T) {
	h, _ := newAPI(t)

	// No token.
	r := httptest.NewRequest(http.MethodGet, "/api/v1/today", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("no token: status %d, want 401", rec.Code)
	}
	if got := rec.Header().Get("WWW-Authenticate"); got != "Bearer" {
		t.Errorf("WWW-Authenticate = %q, want Bearer", got)
	}

	// Wrong token.
	r = httptest.NewRequest(http.MethodGet, "/api/v1/today", nil)
	r.Header.Set("Authorization", "Bearer wrong")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("wrong token: status %d, want 401", rec.Code)
	}

	// Correct token reaches the handler.
	if rec := do(t, h, http.MethodGet, "/api/v1/today", nil, nil); rec.Code != http.StatusOK {
		t.Errorf("valid token: status %d, want 200", rec.Code)
	}
}

func TestRoundTrip(t *testing.T) {
	h, _ := newAPI(t)

	// PUT outline then plan.
	var planRes plan.SavePlanResult
	rec := do(t, h, http.MethodPut, "/api/v1/outline", fixture(t, "sample-outline.json"), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT outline: %d %s", rec.Code, rec.Body)
	}
	rec = do(t, h, http.MethodPut, "/api/v1/plan", fixture(t, "sample-week.json"), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT plan: %d %s", rec.Code, rec.Body)
	}
	decode(t, rec, &planRes)
	if len(planRes.Created) != 5 {
		t.Errorf("created = %d, want 5", len(planRes.Created))
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("content-type = %q", ct)
	}

	// GET plan for the week.
	rec = do(t, h, http.MethodGet, "/api/v1/plan?from=2026-10-26&to=2026-11-01", nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET plan: %d %s", rec.Code, rec.Body)
	}
	var planGet struct {
		Sessions []plan.SessionView `json:"sessions"`
	}
	decode(t, rec, &planGet)
	if len(planGet.Sessions) != 5 {
		t.Fatalf("GET plan sessions = %d, want 5", len(planGet.Sessions))
	}
	if planGet.Sessions[0].Notation != "M1·1, s. 1"+"–"+"22" {
		t.Errorf("first notation = %q", planGet.Sessions[0].Notation)
	}

	// GET today (Monday), week, progress, outline.
	rec = do(t, h, http.MethodGet, "/api/v1/today", nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET today: %d %s", rec.Code, rec.Body)
	}
	var day plan.DayView
	decode(t, rec, &day)
	if day.Heading != "Pazartesi, 26 Ekim" || len(day.Sessions) != 1 {
		t.Errorf("today = %q / %d sessions", day.Heading, len(day.Sessions))
	}

	rec = do(t, h, http.MethodGet, "/api/v1/week?date=2026-10-28", nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET week: %d %s", rec.Code, rec.Body)
	}
	var week plan.WeekView
	decode(t, rec, &week)
	if week.PlannedMin != 540 || week.SessionsPlanned != 5 {
		t.Errorf("week = %d min / %d sessions, want 540 / 5", week.PlannedMin, week.SessionsPlanned)
	}

	rec = do(t, h, http.MethodGet, "/api/v1/progress", nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET progress: %d %s", rec.Code, rec.Body)
	}
	var prog plan.ProgressView
	decode(t, rec, &prog)
	if !prog.HasPlan || prog.WeeksTotal != 38 || len(prog.Groups) != 5 {
		t.Errorf("progress = %+v", prog)
	}

	rec = do(t, h, http.MethodGet, "/api/v1/outline", nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET outline: %d %s", rec.Code, rec.Body)
	}
	var outline plan.OutlineView
	decode(t, rec, &outline)
	if len(outline.Phases) != 1 || len(outline.TopicGroups) != 5 {
		t.Errorf("outline = %d phases, %d groups", len(outline.Phases), len(outline.TopicGroups))
	}
}

func TestPutPlanTwiceUnchanged(t *testing.T) {
	h, _ := newAPI(t)
	seed(t, h)

	rec := do(t, h, http.MethodPut, "/api/v1/plan", fixture(t, "sample-week.json"), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT plan repeat: %d %s", rec.Code, rec.Body)
	}
	var res plan.SavePlanResult
	decode(t, rec, &res)
	if len(res.Unchanged) != 5 || len(res.Created) != 0 {
		t.Errorf("repeat: created=%d unchanged=%d, want 0/5", len(res.Created), len(res.Unchanged))
	}
}

func TestIdempotencyKey(t *testing.T) {
	h, _ := newAPI(t)
	if rec := do(t, h, http.MethodPut, "/api/v1/outline", fixture(t, "sample-outline.json"), nil); rec.Code != http.StatusOK {
		t.Fatalf("outline: %d", rec.Code)
	}
	week := fixture(t, "sample-week.json")
	hdr := map[string]string{"Idempotency-Key": "week-1"}

	if rec := do(t, h, http.MethodPut, "/api/v1/plan", week, hdr); rec.Code != http.StatusOK {
		t.Fatalf("first: %d %s", rec.Code, rec.Body)
	}
	// Same key + same body → replay, 200.
	rec := do(t, h, http.MethodPut, "/api/v1/plan", week, hdr)
	if rec.Code != http.StatusOK {
		t.Fatalf("replay: %d %s", rec.Code, rec.Body)
	}
	var res plan.SavePlanResult
	decode(t, rec, &res)
	if len(res.Created) != 5 {
		t.Errorf("replay created = %d, want the stored 5", len(res.Created))
	}

	// Same key + different body → 409 conflict.
	var req plan.SavePlanRequest
	if err := json.Unmarshal(week, &req); err != nil {
		t.Fatalf("unmarshal week: %v", err)
	}
	req.Sessions = req.Sessions[:4]
	changed, _ := json.Marshal(req)
	rec = do(t, h, http.MethodPut, "/api/v1/plan", changed, hdr)
	if rec.Code != http.StatusConflict {
		t.Fatalf("conflict: status %d, want 409; body %s", rec.Code, rec.Body)
	}
	var env struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	decode(t, rec, &env)
	if env.Error.Code != "idempotency_conflict" {
		t.Errorf("conflict code = %q", env.Error.Code)
	}
}

func TestValidationErrorShape(t *testing.T) {
	h, _ := newAPI(t)
	// A plan with a bad date range and an empty item list.
	body := []byte(`{"from":"2026-10-26","to":"2026-11-01","sessions":[{"date":"bad","start_time":"20:30","duration_min":90,"kind":"aksam","items":[]}]}`)
	rec := do(t, h, http.MethodPut, "/api/v1/plan", body, nil)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422; body %s", rec.Code, rec.Body)
	}
	var env struct {
		Error struct {
			Code   string `json:"code"`
			Fields []struct {
				Path string `json:"path"`
				Code string `json:"code"`
			} `json:"fields"`
		} `json:"error"`
	}
	decode(t, rec, &env)
	if env.Error.Code != "validation_failed" || len(env.Error.Fields) == 0 {
		t.Fatalf("error shape = %+v", env.Error)
	}
	var sawDate bool
	for _, f := range env.Error.Fields {
		if f.Path == "sessions[0].date" {
			sawDate = true
		}
	}
	if !sawDate {
		t.Errorf("expected a field error at sessions[0].date, got %+v", env.Error.Fields)
	}
}

func TestGetPlanMissingParams(t *testing.T) {
	h, _ := newAPI(t)
	rec := do(t, h, http.MethodGet, "/api/v1/plan", nil, nil)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", rec.Code)
	}
}

func TestCompleteAndUndo(t *testing.T) {
	h, st := newAPI(t)
	seed(t, h)

	ctx := context.Background()
	mon, err := st.SessionByDateAndTime(ctx, st.DB(), "2026-10-26", "20:30")
	if err != nil || mon == nil {
		t.Fatalf("find Monday: %v %v", mon, err)
	}
	sat, err := st.SessionByDateAndTime(ctx, st.DB(), "2026-10-31", "10:00")
	if err != nil || sat == nil {
		t.Fatalf("find Saturday: %v %v", sat, err)
	}

	// Completing a future session (Saturday) → 422 with code future.
	rec := do(t, h, http.MethodPost, "/api/v1/sessions/"+itoa(sat.ID)+"/complete", nil, nil)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("future complete: status %d, want 422; body %s", rec.Code, rec.Body)
	}
	var env struct {
		Error struct {
			Fields []struct {
				Code string `json:"code"`
			} `json:"fields"`
		} `json:"error"`
	}
	decode(t, rec, &env)
	if len(env.Error.Fields) == 0 || env.Error.Fields[0].Code != plan.CodeFuture {
		t.Errorf("future error = %+v", env.Error)
	}

	// Completing Monday (today) → 200, done.
	rec = do(t, h, http.MethodPost, "/api/v1/sessions/"+itoa(mon.ID)+"/complete", nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("complete Monday: %d %s", rec.Code, rec.Body)
	}
	var view plan.SessionView
	decode(t, rec, &view)
	if view.State != plan.StateDone || !view.Completed {
		t.Errorf("completed view = %+v", view)
	}

	// Undo → 200, not done.
	rec = do(t, h, http.MethodDelete, "/api/v1/sessions/"+itoa(mon.ID)+"/complete", nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("undo: %d %s", rec.Code, rec.Body)
	}
	decode(t, rec, &view)
	if view.Completed {
		t.Errorf("undo left done: %+v", view)
	}

	// Unknown id → 404.
	rec = do(t, h, http.MethodPost, "/api/v1/sessions/99999/complete", nil, nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("unknown id: status %d, want 404", rec.Code)
	}
}

// itoa is a tiny local int64→string helper to keep the test readable.
func itoa(n int64) string {
	return strconv.FormatInt(n, 10)
}
