package plan

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ibra2626/workey/internal/clock"
	"github.com/ibra2626/workey/internal/store"
)

// istanbul builds a time in Europe/Istanbul from "YYYY-MM-DD HH:MM".
func istanbul(t *testing.T, s string) time.Time {
	t.Helper()
	tm, err := time.ParseInLocation("2006-01-02 15:04", s, clock.Location)
	if err != nil {
		t.Fatalf("parse time %q: %v", s, err)
	}
	return tm
}

// newService opens a migrated store on a temp database and wraps it in a
// Service whose clock is fixed at now.
func newService(t *testing.T, now time.Time) (*Service, *store.Store) {
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
	return New(st, clock.Fake{Current: now}), st
}

// loadFixture decodes testdata/<name> into v.
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

func auditCount(t *testing.T, st *store.Store) int {
	t.Helper()
	var n int
	if err := st.DB().QueryRow("SELECT COUNT(1) FROM audit_log").Scan(&n); err != nil {
		t.Fatalf("count audit: %v", err)
	}
	return n
}

func sessionAuditCount(t *testing.T, st *store.Store) int {
	t.Helper()
	var n int
	if err := st.DB().QueryRow("SELECT COUNT(1) FROM audit_log WHERE entity = 'session'").Scan(&n); err != nil {
		t.Fatalf("count session audit: %v", err)
	}
	return n
}

func sessionAt(t *testing.T, st *store.Store, date, startTime string) *store.Session {
	t.Helper()
	s, err := st.SessionByDateAndTime(context.Background(), st.DB(), date, startTime)
	if err != nil {
		t.Fatalf("SessionByDateAndTime: %v", err)
	}
	return s
}

// loadWeek applies the sample outline then the sample week at the service's
// clock, returning the decoded week request.
func loadWeek(t *testing.T, svc *Service) SavePlanRequest {
	t.Helper()
	ctx := context.Background()
	var outline SaveOutlineRequest
	loadFixture(t, "sample-outline.json", &outline)
	if _, err := svc.SaveOutline(ctx, SourceAPI, outline); err != nil {
		t.Fatalf("SaveOutline: %v", err)
	}
	var week SavePlanRequest
	loadFixture(t, "sample-week.json", &week)
	return week
}

func TestSavePlanSampleWeekCreateAndIdempotent(t *testing.T) {
	ctx := context.Background()
	svc, st := newService(t, istanbul(t, "2026-10-26 09:00"))
	week := loadWeek(t, svc)

	res, err := svc.SavePlan(ctx, SourceAPI, week)
	if err != nil {
		t.Fatalf("SavePlan: %v", err)
	}
	if len(res.Created) != 5 {
		t.Fatalf("created = %d, want 5", len(res.Created))
	}
	if len(res.Updated)+len(res.Deleted)+len(res.Unchanged)+len(res.Skipped) != 0 {
		t.Fatalf("unexpected non-create results: %+v", res)
	}

	// The Wednesday session round-trips its konu/soru/antrenman split.
	wed := sessionAt(t, st, "2026-10-28", "20:30")
	if wed == nil || len(wed.Items) != 1 {
		t.Fatalf("wednesday session not stored as expected: %+v", wed)
	}
	it := wed.Items[0]
	if it.AntrenmanFrom == nil || *it.AntrenmanFrom != 1 || it.AntrenmanTo == nil || *it.AntrenmanTo != 3 {
		t.Errorf("wednesday antrenman = %v..%v, want 1..3", it.AntrenmanFrom, it.AntrenmanTo)
	}

	auditAfterCreate := sessionAuditCount(t, st)
	if auditAfterCreate != 5 {
		t.Fatalf("session audit rows after create = %d, want 5", auditAfterCreate)
	}

	// Same payload again: everything unchanged, no new audit rows.
	res2, err := svc.SavePlan(ctx, SourceAPI, week)
	if err != nil {
		t.Fatalf("SavePlan repeat: %v", err)
	}
	if len(res2.Unchanged) != 5 || len(res2.Created) != 0 {
		t.Fatalf("repeat results: created=%d unchanged=%d, want 0/5", len(res2.Created), len(res2.Unchanged))
	}
	if n := sessionAuditCount(t, st); n != auditAfterCreate {
		t.Fatalf("audit rows grew on idempotent repeat: %d -> %d", auditAfterCreate, n)
	}
}

func TestSavePlanChangeOneSession(t *testing.T) {
	ctx := context.Background()
	svc, st := newService(t, istanbul(t, "2026-10-26 09:00"))
	week := loadWeek(t, svc)
	if _, err := svc.SavePlan(ctx, SourceAPI, week); err != nil {
		t.Fatalf("SavePlan: %v", err)
	}
	before := auditCount(t, st)

	// Change only Thursday's page range.
	week.Sessions[3].Items[0].PageTo = 80
	res, err := svc.SavePlan(ctx, SourceAPI, week)
	if err != nil {
		t.Fatalf("SavePlan change: %v", err)
	}
	if len(res.Updated) != 1 || len(res.Unchanged) != 4 {
		t.Fatalf("change results: updated=%d unchanged=%d, want 1/4", len(res.Updated), len(res.Unchanged))
	}
	if res.Updated[0].Date != "2026-10-29" {
		t.Errorf("updated session date = %s, want 2026-10-29", res.Updated[0].Date)
	}
	thu := sessionAt(t, st, "2026-10-29", "20:30")
	if thu.Items[0].PageTo != 80 {
		t.Errorf("thursday page_to = %d, want 80", thu.Items[0].PageTo)
	}
	if n := auditCount(t, st); n != before+1 {
		t.Errorf("audit rows = %d, want %d (one update)", n, before+1)
	}
}

func TestSavePlanPastDateSkipped(t *testing.T) {
	ctx := context.Background()
	svc, _ := newService(t, istanbul(t, "2026-10-28 09:00"))
	week := loadWeek(t, svc)

	res, err := svc.SavePlan(ctx, SourceAPI, week)
	if err != nil {
		t.Fatalf("SavePlan: %v", err)
	}
	if len(res.Skipped) != 2 {
		t.Fatalf("skipped = %d, want 2 (Mon 26, Tue 27)", len(res.Skipped))
	}
	for _, sk := range res.Skipped {
		if sk.Reason != SkipPastDate {
			t.Errorf("skip reason = %s, want %s", sk.Reason, SkipPastDate)
		}
	}
	if len(res.Created) != 3 {
		t.Errorf("created = %d, want 3 (Wed, Thu, Sat)", len(res.Created))
	}
}

func TestSavePlanCompletedUntouched(t *testing.T) {
	ctx := context.Background()
	svc, st := newService(t, istanbul(t, "2026-10-26 09:00"))
	week := loadWeek(t, svc)
	if _, err := svc.SavePlan(ctx, SourceAPI, week); err != nil {
		t.Fatalf("SavePlan: %v", err)
	}

	// Mark Wednesday complete directly (completion is WKY-05's concern).
	wed := sessionAt(t, st, "2026-10-28", "20:30")
	if _, err := st.DB().ExecContext(ctx,
		`UPDATE sessions SET completed_at = ?, completed_source = ? WHERE id = ?`,
		"2026-10-28T19:00:00Z", "web", wed.ID); err != nil {
		t.Fatalf("mark completed: %v", err)
	}
	before := auditCount(t, st)

	// Try to change Wednesday's pages: it must be skipped and left intact.
	week.Sessions[2].Items[0].PageTo = 60
	res, err := svc.SavePlan(ctx, SourceAPI, week)
	if err != nil {
		t.Fatalf("SavePlan: %v", err)
	}
	if len(res.Skipped) != 1 || res.Skipped[0].Reason != SkipCompleted {
		t.Fatalf("skipped = %+v, want one completed", res.Skipped)
	}
	if len(res.Deleted) != 0 {
		t.Fatalf("a completed session must never be deleted, got %+v", res.Deleted)
	}
	again := sessionAt(t, st, "2026-10-28", "20:30")
	if again.Items[0].PageTo != 58 {
		t.Errorf("completed session changed: page_to = %d, want 58", again.Items[0].PageTo)
	}
	if n := auditCount(t, st); n != before {
		t.Errorf("audit rows changed for a frozen session: %d -> %d", before, n)
	}
}

func TestSavePlanShrinkDeletesOpenFuture(t *testing.T) {
	ctx := context.Background()
	svc, st := newService(t, istanbul(t, "2026-10-26 09:00"))
	week := loadWeek(t, svc)
	if _, err := svc.SavePlan(ctx, SourceAPI, week); err != nil {
		t.Fatalf("SavePlan: %v", err)
	}

	// Keep only Monday and Wednesday in the same window.
	shrunk := week
	shrunk.Sessions = []SessionInput{week.Sessions[0], week.Sessions[2]}
	res, err := svc.SavePlan(ctx, SourceAPI, shrunk)
	if err != nil {
		t.Fatalf("SavePlan shrink: %v", err)
	}
	if len(res.Deleted) != 3 {
		t.Fatalf("deleted = %d, want 3 (Tue, Thu, Sat)", len(res.Deleted))
	}
	if len(res.Unchanged) != 2 {
		t.Fatalf("unchanged = %d, want 2 (Mon, Wed)", len(res.Unchanged))
	}
	if sessionAt(t, st, "2026-10-27", "20:30") != nil {
		t.Errorf("Tuesday session should have been deleted")
	}
	if sessionAt(t, st, "2026-10-26", "20:30") == nil {
		t.Errorf("Monday session should remain")
	}
}

func TestSavePlanMakeupLink(t *testing.T) {
	ctx := context.Background()
	svc, st := newService(t, istanbul(t, "2026-10-28 09:00"))

	// A past, uncompleted session: the missed Tuesday block.
	now := svc.nowUTC()
	missed := &store.Session{Date: "2026-10-27", StartTime: "20:30", DurationMin: 90, Kind: KindAksam, CreatedAt: now, UpdatedAt: now}
	if err := st.InsertSession(ctx, st.DB(), missed); err != nil {
		t.Fatalf("insert missed: %v", err)
	}

	item := Item{Book: "M1", ChapterNo: 2, ChapterTitle: "Sayı basamakları", PageFrom: 23, PageTo: 40}

	t.Run("valid past target links", func(t *testing.T) {
		req := SavePlanRequest{From: "2026-10-28", To: "2026-10-31", Sessions: []SessionInput{
			{Date: "2026-10-31", StartTime: "10:00", DurationMin: 180, Kind: KindUzunBlok,
				MakeupOf: &SlotRef{Date: "2026-10-27", StartTime: "20:30"}, Items: []Item{item}},
		}}
		res, err := svc.SavePlan(ctx, SourceAPI, req)
		if err != nil {
			t.Fatalf("SavePlan: %v", err)
		}
		if len(res.Created) != 1 {
			t.Fatalf("created = %d, want 1", len(res.Created))
		}
		sat := sessionAt(t, st, "2026-10-31", "10:00")
		if sat.MakeupOf == nil || *sat.MakeupOf != missed.ID {
			t.Errorf("makeup_of = %v, want %d", sat.MakeupOf, missed.ID)
		}
	})

	t.Run("unknown target rejected", func(t *testing.T) {
		req := SavePlanRequest{From: "2026-10-28", To: "2026-10-31", Sessions: []SessionInput{
			{Date: "2026-10-30", StartTime: "10:00", DurationMin: 180, Kind: KindUzunBlok,
				MakeupOf: &SlotRef{Date: "2026-10-20", StartTime: "20:30"}, Items: []Item{item}},
		}}
		_, err := svc.SavePlan(ctx, SourceAPI, req)
		ve, ok := AsValidationError(err)
		if !ok {
			t.Fatalf("want ValidationError, got %v", err)
		}
		if ve.Errors[0].Code != CodeNotFound {
			t.Errorf("code = %s, want %s", ve.Errors[0].Code, CodeNotFound)
		}
	})

	t.Run("future target rejected", func(t *testing.T) {
		future := &store.Session{Date: "2026-11-10", StartTime: "20:30", DurationMin: 90, Kind: KindAksam, CreatedAt: now, UpdatedAt: now}
		if err := st.InsertSession(ctx, st.DB(), future); err != nil {
			t.Fatalf("insert future: %v", err)
		}
		req := SavePlanRequest{From: "2026-10-28", To: "2026-10-31", Sessions: []SessionInput{
			{Date: "2026-10-30", StartTime: "10:00", DurationMin: 180, Kind: KindUzunBlok,
				MakeupOf: &SlotRef{Date: "2026-11-10", StartTime: "20:30"}, Items: []Item{item}},
		}}
		_, err := svc.SavePlan(ctx, SourceAPI, req)
		ve, ok := AsValidationError(err)
		if !ok {
			t.Fatalf("want ValidationError, got %v", err)
		}
		if ve.Errors[0].Code != CodeMakeupTarget {
			t.Errorf("code = %s, want %s", ve.Errors[0].Code, CodeMakeupTarget)
		}
	})
}

func TestSavePlanIdempotencyKey(t *testing.T) {
	ctx := context.Background()
	svc, st := newService(t, istanbul(t, "2026-10-26 09:00"))
	week := loadWeek(t, svc)
	week.IdempotencyKey = "week-1-v1"

	res, err := svc.SavePlan(ctx, SourceAPI, week)
	if err != nil {
		t.Fatalf("SavePlan: %v", err)
	}
	if len(res.Created) != 5 {
		t.Fatalf("created = %d, want 5", len(res.Created))
	}
	after := auditCount(t, st)

	// Repeat with the same key and payload: the stored response is replayed and
	// no new audit rows are written.
	res2, err := svc.SavePlan(ctx, SourceAPI, week)
	if err != nil {
		t.Fatalf("SavePlan replay: %v", err)
	}
	if len(res2.Created) != 5 {
		t.Errorf("replayed created = %d, want the stored 5", len(res2.Created))
	}
	if n := auditCount(t, st); n != after {
		t.Errorf("audit rows grew on replay: %d -> %d", after, n)
	}

	// Same key, different payload: conflict.
	conflict := week
	conflict.Sessions = week.Sessions[:4]
	if _, err := svc.SavePlan(ctx, SourceAPI, conflict); err != ErrIdempotencyConflict {
		t.Errorf("want ErrIdempotencyConflict, got %v", err)
	}
}

func TestSavePlanAuditBeforeAfter(t *testing.T) {
	ctx := context.Background()
	svc, st := newService(t, istanbul(t, "2026-10-26 09:00"))
	week := loadWeek(t, svc)
	if _, err := svc.SavePlan(ctx, SourceAPI, week); err != nil {
		t.Fatalf("SavePlan: %v", err)
	}
	week.Sessions[0].Items[0].PageTo = 25
	if _, err := svc.SavePlan(ctx, SourceAPI, week); err != nil {
		t.Fatalf("SavePlan update: %v", err)
	}

	var source, action string
	var before, after *string
	err := st.DB().QueryRowContext(ctx,
		`SELECT source, action, before_json, after_json FROM audit_log
		 WHERE action = 'update' ORDER BY id DESC LIMIT 1`).
		Scan(&source, &action, &before, &after)
	if err != nil {
		t.Fatalf("read audit: %v", err)
	}
	if source != string(SourceAPI) {
		t.Errorf("audit source = %q, want api", source)
	}
	if before == nil || after == nil {
		t.Fatalf("update audit missing before/after JSON")
	}
	var b, a map[string]any
	if err := json.Unmarshal([]byte(*before), &b); err != nil {
		t.Fatalf("before json: %v", err)
	}
	if err := json.Unmarshal([]byte(*after), &a); err != nil {
		t.Fatalf("after json: %v", err)
	}
	if b["id"] != a["id"] {
		t.Errorf("before/after id mismatch: %v vs %v", b["id"], a["id"])
	}
}

func TestSaveOutlineReplaceAndReferencedDelete(t *testing.T) {
	ctx := context.Background()
	svc, st := newService(t, istanbul(t, "2026-10-26 09:00"))

	var outline SaveOutlineRequest
	loadFixture(t, "sample-outline.json", &outline)
	res, err := svc.SaveOutline(ctx, SourceAPI, outline)
	if err != nil {
		t.Fatalf("SaveOutline: %v", err)
	}
	if len(res.TopicGroups.Created) != 5 || len(res.Phases.Created) != 1 {
		t.Fatalf("outline create: tg=%d phases=%d, want 5/1", len(res.TopicGroups.Created), len(res.Phases.Created))
	}

	// Idempotent repeat.
	res2, err := svc.SaveOutline(ctx, SourceAPI, outline)
	if err != nil {
		t.Fatalf("SaveOutline repeat: %v", err)
	}
	if len(res2.TopicGroups.Unchanged) != 5 || len(res2.Phases.Unchanged) != 1 {
		t.Fatalf("outline repeat not all unchanged: %+v", res2)
	}

	// A session references the "temel-kavramlar" group.
	var week SavePlanRequest
	loadFixture(t, "sample-week.json", &week)
	if _, err := svc.SavePlan(ctx, SourceAPI, week); err != nil {
		t.Fatalf("SavePlan: %v", err)
	}

	// Removing the referenced group from the outline must fail.
	reduced := outline
	reduced.TopicGroups = outline.TopicGroups[1:] // drop temel-kavramlar
	_, err = svc.SaveOutline(ctx, SourceAPI, reduced)
	ve, ok := AsValidationError(err)
	if !ok {
		t.Fatalf("want ValidationError for referenced delete, got %v", err)
	}
	if ve.Errors[0].Code != CodeReferenced {
		t.Errorf("code = %s, want %s", ve.Errors[0].Code, CodeReferenced)
	}
	// The group must still be present (rolled back).
	g, err := st.TopicGroupByKey(ctx, st.DB(), "temel-kavramlar")
	if err != nil {
		t.Fatalf("TopicGroupByKey: %v", err)
	}
	if g == nil {
		t.Errorf("referenced group was deleted despite the error")
	}
}
