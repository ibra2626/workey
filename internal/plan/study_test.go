package plan

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	"github.com/ibra2626/workey/internal/clock"
	"github.com/ibra2626/workey/internal/store"
)

var studyExam = clock.Date{Year: 2027, Month: 7, Day: 18}

// All sources in these tests are disposable sample fixtures, not a real plan.
func studyFixture(t *testing.T) (*Service, *store.Store, int64, int64) {
	t.Helper()
	svc, st := newService(t, istanbul(t, "2026-10-26 21:00"))
	req := loadWeek(t, svc)
	req.Sessions[0].Items = append(req.Sessions[0].Items, Item{Book: "G1", ChapterNo: 1, ChapterTitle: "Temsili kaynak", PageFrom: 100, PageTo: 110})
	if _, err := svc.SavePlan(context.Background(), SourceAPI, req); err != nil {
		t.Fatal(err)
	}
	first := sessionAt(t, st, "2026-10-26", "20:30").ID
	second := sessionAt(t, st, "2026-10-27", "20:30").ID
	return svc, st, first, second
}

func storedStudy(t *testing.T, st *store.Store, id int64) *store.Study {
	t.Helper()
	v, err := st.StudyBySession(context.Background(), st.DB(), id)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func requireConflict(t *testing.T, err error) {
	t.Helper()
	if !errors.Is(err, ErrStudyConflict) {
		t.Fatalf("want conflict, got %v", err)
	}
}

func TestStudyLifecycleVersionsAndBookmarks(t *testing.T) {
	ctx := context.Background()
	svc, st, id, _ := studyFixture(t)
	beforeSession, _ := st.SessionByID(ctx, st.DB(), id)
	v, err := svc.StartStudy(ctx, SourceWeb, id, EntryTenMinutes, 0, studyExam)
	if err != nil {
		t.Fatal(err)
	}
	if v.Version != 1 || v.Status != StudyActive || v.ItemOrdinal != nil {
		t.Fatalf("start: %+v", v)
	}
	count := auditCount(t, st)
	if _, err := svc.StartStudy(ctx, SourceWeb, id, EntryTenMinutes, 0, studyExam); err != nil {
		t.Fatal(err)
	}
	if auditCount(t, st) != count {
		t.Fatal("start retry wrote audit")
	}
	_, err = svc.StartStudy(ctx, SourceWeb, id, EntryNormal, 0, studyExam)
	requireConflict(t, err)
	bookmark := &Bookmark{ItemOrdinal: 1, Stage: StageSoru, ResumePage: 105} // fallback to main range
	v, err = svc.PauseStudy(ctx, SourceWeb, id, bookmark, 1)
	if err != nil {
		t.Fatal(err)
	}
	if v.Version != 2 || *v.ResumePage != 105 {
		t.Fatalf("pause: %+v", v)
	}
	count = auditCount(t, st)
	if _, err := svc.PauseStudy(ctx, SourceWeb, id, bookmark, 1); err != nil {
		t.Fatal(err)
	}
	if auditCount(t, st) != count {
		t.Fatal("pause retry wrote audit")
	}
	_, err = svc.PauseStudy(ctx, SourceWeb, id, &Bookmark{1, StageSoru, 106}, 1)
	requireConflict(t, err)
	// An explicit correction on paused work is allowed; a nil bookmark never clears it.
	v, err = svc.PauseStudy(ctx, SourceWeb, id, &Bookmark{1, StageSoru, 106}, 2)
	if err != nil || v.Version != 3 {
		t.Fatalf("correction: %+v %v", v, err)
	}
	firstStarted := v.FirstStartedAt
	svc.clk = clock.Fake{Current: istanbul(t, "2026-10-27 00:01")}
	v, err = svc.StartStudy(ctx, SourceWeb, id, EntryNormal, 3, studyExam)
	if err != nil {
		t.Fatal(err)
	}
	if v.FirstStartedAt != firstStarted || v.LastStartedAt != svc.nowUTC() || *v.ResumePage != 106 {
		t.Fatalf("resume lost data: %+v", v)
	}
	_, err = svc.StartStudy(ctx, SourceWeb, id, EntryNormal, 0, studyExam)
	requireConflict(t, err) // stale ABA request
	v, err = svc.PauseStudy(ctx, SourceWeb, id, nil, 4)
	if err != nil || v.Version != 5 || *v.ResumePage != 106 {
		t.Fatalf("close: %+v %v", v, err)
	}
	if _, err := svc.Complete(ctx, SourceWeb, id, 5); err != nil {
		t.Fatal(err)
	}
	v = storedStudy(t, st, id)
	if v.Status != StudyClosed || v.Version != 6 || *v.ResumePage != 106 {
		t.Fatalf("complete: %+v", v)
	}
	count = auditCount(t, st)
	if _, err := svc.Complete(ctx, SourceWeb, id, 5); err != nil {
		t.Fatal(err)
	}
	if auditCount(t, st) != count {
		t.Fatal("completion retry wrote audit")
	}
	_, err = svc.PauseStudy(ctx, SourceWeb, id, bookmark, 5)
	requireConflict(t, err)
	_, err = svc.Undo(ctx, SourceWeb, id, 4)
	requireConflict(t, err)
	// Legacy API/MCP calls still work without a version.
	if _, err := svc.Undo(ctx, SourceMCP, id); err != nil {
		t.Fatal(err)
	}
	v = storedStudy(t, st, id)
	if v.Status != StudyPaused || v.Version != 7 || *v.ResumePage != 106 {
		t.Fatalf("undo: %+v", v)
	}
	_, err = svc.Complete(ctx, SourceWeb, id, 5)
	requireConflict(t, err)
	afterSession, _ := st.SessionByID(ctx, st.DB(), id)
	if !reflect.DeepEqual(beforeSession.Items, afterSession.Items) || beforeSession.Date != afterSession.Date || beforeSession.DurationMin != afterSession.DurationMin {
		t.Fatal("study changed plan")
	}
}

func TestStudyBookmarkValidation(t *testing.T) {
	for _, kind := range []string{KindAksam, KindUzunBlok, KindDeneme} {
		for _, stage := range []string{StageKonu, StageSoru, StageYanlisDefteri, StageCalisma, ""} {
			for _, page := range []int{0, 9, 10, 12, 13, 14, 20, 21} {
				ss := store.Session{Kind: kind, Items: []store.SessionItem{{Ordinal: 0, PageFrom: 10, PageTo: 20, KonuFrom: ptr(10), KonuTo: ptr(12), SoruFrom: ptr(13), SoruTo: ptr(20)}}}
				valid := kind != KindAksam && stage == StageCalisma && page >= 10 && page <= 20
				if kind == KindAksam {
					valid = (stage == StageKonu && page >= 10 && page <= 12) || (stage == StageSoru && page >= 13 && page <= 20) || (stage == StageYanlisDefteri && page >= 10 && page <= 20)
				}
				err := validateBookmark(ss, Bookmark{0, stage, page})
				if (err == nil) != valid {
					t.Errorf("kind=%s stage=%s page=%d err=%v", kind, stage, page, err)
				}
			}
		}
	}
	ctx := context.Background()
	svc, st, id, _ := studyFixture(t)
	if _, err := svc.StartStudy(ctx, SourceWeb, id, EntryNormal, 0, studyExam); err != nil {
		t.Fatal(err)
	}
	before := storedStudy(t, st, id)
	count := auditCount(t, st)
	for _, b := range []Bookmark{{-1, StageKonu, 1}, {10, StageSoru, 1}, {0, StageCalisma, 1}, {1, StageSoru, 111}, {1, "", 105}, {1, StageSoru, 0}} {
		if _, err := svc.PauseStudy(ctx, SourceWeb, id, &b, 1); err == nil {
			t.Fatalf("accepted %+v", b)
		}
		if !reflect.DeepEqual(before, storedStudy(t, st, id)) || auditCount(t, st) != count {
			t.Fatal("invalid bookmark changed data")
		}
	}
}

func TestStudySingleActiveAndConcurrentStart(t *testing.T) {
	ctx := context.Background()
	svc, st, first, second := studyFixture(t)
	svc.clk = clock.Fake{Current: istanbul(t, "2026-10-27 21:00")}
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for _, id := range []int64{first, second} {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			_, err := svc.StartStudy(ctx, SourceWeb, id, EntryNormal, 0, studyExam)
			errs <- err
		}(id)
	}
	wg.Wait()
	close(errs)
	success, conflicts := 0, 0
	for err := range errs {
		if err == nil {
			success++
		} else if errors.Is(err, ErrStudyConflict) {
			conflicts++
		} else {
			t.Fatal(err)
		}
	}
	if success != 1 || conflicts != 1 {
		t.Fatalf("success=%d conflict=%d", success, conflicts)
	}
	c, err := svc.ReadStudyContext(ctx)
	if err != nil || c.Active == nil {
		t.Fatalf("context: %+v %v", c, err)
	}
	if _, err := svc.PauseStudy(ctx, SourceWeb, c.Active.Session.ID, nil, 1); err != nil {
		t.Fatal(err)
	}
	other := first
	if other == c.Active.Session.ID {
		other = second
	}
	if _, err := svc.StartStudy(ctx, SourceWeb, other, EntryNormal, 0, studyExam); err != nil {
		t.Fatal(err)
	}
	c, err = svc.ReadStudyContext(ctx)
	if err != nil || c.Active.Session.ID != other || c.LatestPaused == nil || c.LatestPaused.Study.ResumePage != nil {
		t.Fatalf("context: %+v %v", c, err)
	}
	if v := storedStudy(t, st, c.LatestPaused.Session.ID); v.Version != 2 {
		t.Fatal("reading mutated study")
	}
}

func TestStudyTemporalRulesAndMakeup(t *testing.T) {
	ctx := context.Background()
	svc, st, id, future := studyFixture(t)
	_, err := svc.StartStudy(ctx, SourceWeb, future, EntryNormal, 0, studyExam)
	if ve, ok := AsValidationError(err); !ok || ve.Errors[0].Code != CodeFuture {
		t.Fatalf("future: %v", err)
	}
	if _, err := svc.StartStudy(ctx, SourceWeb, 999, EntryNormal, 0, studyExam); err == nil {
		t.Fatal("missing accepted")
	}
	if _, err := svc.StartStudy(ctx, SourceWeb, id, "bad", 0, studyExam); err == nil {
		t.Fatal("bad mode accepted")
	}
	if _, err := svc.StartStudy(ctx, SourceWeb, id, EntryNormal, -1, studyExam); err == nil {
		t.Fatal("negative version accepted")
	}
	if _, err := svc.StartStudy(ctx, Source("bad"), id, EntryNormal, 0, studyExam); err == nil {
		t.Fatal("bad source accepted")
	}
	if _, err := svc.PauseStudy(ctx, SourceWeb, id, nil, 0); err == nil {
		t.Fatal("never-started pause accepted")
	}
	if _, err := svc.StartStudy(ctx, SourceWeb, id, EntryNormal, 0, studyExam); err != nil {
		t.Fatal(err)
	}
	svc.clk = clock.Fake{Current: istanbul(t, "2026-10-27 00:00")}
	c, err := svc.ReadStudyContext(ctx)
	if err != nil || c.Active.Session.Date != "2026-10-26" {
		t.Fatalf("midnight: %+v %v", c, err)
	}
	if _, err := svc.PauseStudy(ctx, SourceWeb, id, &Bookmark{1, StageKonu, 100}, 1); err != nil {
		t.Fatal(err)
	}
	// A makeup points to the old work without inheriting its bookmark.
	req := SavePlanRequest{From: "2026-10-27", To: "2026-10-27", Sessions: []SessionInput{{Date: "2026-10-27", StartTime: "20:30", Kind: KindAksam, DurationMin: 90, Items: []Item{{Book: "M1", ChapterNo: 1, ChapterTitle: "Temsili", PageFrom: 1, PageTo: 5}}, MakeupOf: &SlotRef{Date: "2026-10-26", StartTime: "20:30"}}}}
	if _, err := svc.SavePlan(ctx, SourceAPI, req); err != nil {
		t.Fatal(err)
	}
	_, err = svc.StartStudy(ctx, SourceWeb, id, EntryNormal, 2, studyExam)
	var moved *StudyMovedError
	if !errors.As(err, &moved) || moved.SessionID != future {
		t.Fatalf("moved: %v", err)
	}
	c, err = svc.ReadStudyContext(ctx)
	if err != nil || c.LatestPaused.MovedTo.ID != future || c.LatestPaused.BookmarkItem.Book != "G1" {
		t.Fatalf("resolved context: %+v %v", c, err)
	}
	if _, err := svc.Complete(ctx, SourceWeb, id, 2); err != nil {
		t.Fatal(err)
	}
	makeup, _ := st.SessionByID(ctx, st.DB(), future)
	if makeup.CompletedAt != nil || storedStudy(t, st, future) != nil {
		t.Fatal("completion mutated makeup")
	}
	if _, err := svc.Undo(ctx, SourceWeb, id, 3); err != nil {
		t.Fatal(err)
	}
	// Rest day permits deliberate resume, while the original slot stays fixed.
	svc.clk = clock.Fake{Current: istanbul(t, "2026-11-01 09:00")}
	if _, err := svc.StartStudy(ctx, SourceWeb, future, EntryNormal, 0, studyExam); err != nil {
		t.Fatal(err)
	}
	// Beyond plan/exam, existing work can still be paused/completed/undone.
	svc.clk = clock.Fake{Current: istanbul(t, "2027-07-19 00:00")}
	if _, err := svc.PauseStudy(ctx, SourceWeb, future, nil, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.StartStudy(ctx, SourceWeb, future, EntryNormal, 2, studyExam); err == nil {
		t.Fatal("post-exam start accepted")
	}
	if _, err := svc.Complete(ctx, SourceAPI, future); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Undo(ctx, SourceMCP, future); err != nil {
		t.Fatal(err)
	}
	svc.clk = clock.Fake{Current: istanbul(t, "2026-12-14 09:00")}
	if _, err := svc.StartStudy(ctx, SourceWeb, future, EntryNormal, 4, studyExam); err == nil {
		t.Fatal("post-plan start accepted")
	}
}

func TestStudyRollbackOnAuditFailure(t *testing.T) {
	for _, operation := range []string{"start", "pause", "complete", "undo"} {
		t.Run(operation, func(t *testing.T) {
			ctx := context.Background()
			svc, st, id, _ := studyFixture(t)
			if operation != "start" {
				if _, err := svc.StartStudy(ctx, SourceWeb, id, EntryNormal, 0, studyExam); err != nil {
					t.Fatal(err)
				}
			}
			if operation == "undo" {
				if _, err := svc.Complete(ctx, SourceWeb, id, 1); err != nil {
					t.Fatal(err)
				}
			}
			before := storedStudy(t, st, id)
			ssBefore, _ := st.SessionByID(ctx, st.DB(), id)
			count := auditCount(t, st)
			if _, err := st.DB().Exec(`CREATE TRIGGER fail_study_audit BEFORE INSERT ON audit_log WHEN NEW.entity = 'session_study' BEGIN SELECT RAISE(ABORT, 'test audit failure'); END`); err != nil {
				t.Fatal(err)
			}
			var err error
			switch operation {
			case "start":
				_, err = svc.StartStudy(ctx, SourceWeb, id, EntryNormal, 0, studyExam)
			case "pause":
				_, err = svc.PauseStudy(ctx, SourceWeb, id, &Bookmark{1, StageKonu, 105}, 1)
			case "complete":
				_, err = svc.Complete(ctx, SourceWeb, id, 1)
			case "undo":
				_, err = svc.Undo(ctx, SourceWeb, id, 2)
			}
			if err == nil {
				t.Fatal("audit failure did not fail operation")
			}
			ssAfter, _ := st.SessionByID(ctx, st.DB(), id)
			if !reflect.DeepEqual(before, storedStudy(t, st, id)) || !reflect.DeepEqual(ssBefore, ssAfter) || auditCount(t, st) != count {
				t.Fatal("failed operation was not atomic")
			}
		})
	}
}

func TestStudyNoBookmarkCompletionUndo(t *testing.T) {
	ctx := context.Background()
	svc, st, id, _ := studyFixture(t)
	if _, err := svc.StartStudy(ctx, SourceWeb, id, EntryNormal, 0, studyExam); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Complete(ctx, SourceAPI, id); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Undo(ctx, SourceMCP, id); err != nil {
		t.Fatal(err)
	}
	v := storedStudy(t, st, id)
	if v.Status != StudyPaused || v.Version != 3 || v.ResumePage != nil {
		t.Fatalf("undo: %+v", v)
	}
	if _, err := svc.PauseStudy(ctx, SourceWeb, id, nil, 3); err != nil {
		t.Fatal(err)
	}
	if storedStudy(t, st, id).Version != 3 {
		t.Fatal("nil pause mutated empty bookmark")
	}
}

func TestStudyContextSurvivesReopenAndSelectsLastPaused(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "study.db")
	st, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()
	if err := st.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	svc := New(st, clock.Fake{Current: istanbul(t, "2026-10-26 21:00")})
	req := loadWeek(t, svc)
	if _, err := svc.SavePlan(ctx, SourceAPI, req); err != nil {
		t.Fatal(err)
	}
	id := sessionAt(t, st, "2026-10-26", "20:30").ID
	second := sessionAt(t, st, "2026-10-27", "20:30").ID
	svc.clk = clock.Fake{Current: istanbul(t, "2026-10-27 21:00")}
	if _, err := svc.StartStudy(ctx, SourceWeb, id, EntryTenMinutes, 0, studyExam); err != nil {
		t.Fatal(err)
	}
	if err := st.Close(); err != nil {
		t.Fatal(err)
	}
	st, err = store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	svc = New(st, clock.Fake{Current: istanbul(t, "2026-10-27 21:00")})
	c, err := svc.ReadStudyContext(ctx)
	if err != nil || c.Active == nil || c.Active.Study.EntryMode != EntryTenMinutes || c.Active.Session.ID != id {
		t.Fatalf("reopened: %+v %v", c, err)
	}
	if _, err := svc.PauseStudy(ctx, SourceWeb, id, nil, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.StartStudy(ctx, SourceWeb, second, EntryNormal, 0, studyExam); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.PauseStudy(ctx, SourceWeb, second, nil, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.StartStudy(ctx, SourceWeb, id, EntryNormal, 2, studyExam); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.PauseStudy(ctx, SourceWeb, id, nil, 3); err != nil {
		t.Fatal(err)
	}
	// All writes share one clock second. Audit order, rather than larger ID,
	// identifies the last deliberate pause.
	c, err = svc.ReadStudyContext(ctx)
	if err != nil || c.Active != nil || c.LatestPaused.Session.ID != id {
		t.Fatalf("latest pause: %+v %v", c, err)
	}
}

func TestStudyTwoTabsCannotOverwriteBookmark(t *testing.T) {
	ctx := context.Background()
	svc, st, id, _ := studyFixture(t)
	if _, err := svc.StartStudy(ctx, SourceWeb, id, EntryNormal, 0, studyExam); err != nil {
		t.Fatal(err)
	}
	count := auditCount(t, st)
	errs := make(chan error, 2)
	for _, page := range []int{104, 105} {
		go func(page int) {
			_, err := svc.PauseStudy(ctx, SourceWeb, id, &Bookmark{1, StageSoru, page}, 1)
			errs <- err
		}(page)
	}
	a, b := <-errs, <-errs
	if a != nil {
		a, b = b, a
	}
	if a != nil || !errors.Is(b, ErrStudyConflict) || auditCount(t, st) != count+1 {
		t.Fatalf("concurrent pause: %v %v", a, b)
	}
}

func TestStudyCompletionRollsBackIfSessionAuditFails(t *testing.T) {
	ctx := context.Background()
	svc, st, id, _ := studyFixture(t)
	if _, err := svc.StartStudy(ctx, SourceWeb, id, EntryNormal, 0, studyExam); err != nil {
		t.Fatal(err)
	}
	count := auditCount(t, st)
	if _, err := st.DB().Exec(`CREATE TRIGGER fail_complete_audit BEFORE INSERT ON audit_log WHEN NEW.entity = 'session' AND NEW.action = 'complete' BEGIN SELECT RAISE(ABORT, 'test failure after study audit'); END`); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Complete(ctx, SourceWeb, id, 1); err == nil {
		t.Fatal("audit failure ignored")
	}
	v := storedStudy(t, st, id)
	ss, err := st.SessionByID(ctx, st.DB(), id)
	if err != nil || ss.CompletedAt != nil || v.Status != StudyActive || v.Version != 1 || auditCount(t, st) != count {
		t.Fatalf("completion leaked on failure: %+v %+v %v", ss, v, err)
	}
}
