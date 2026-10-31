package plan

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/ibra2626/workey/internal/clock"
)

func TestStudyReadSnapshotsAndProgress(t *testing.T) {
	ctx := context.Background()
	svc, st, first, second := studyFixture(t)
	date := svc.TodayDate()
	before, err := svc.Progress(ctx, studyExam)
	if err != nil {
		t.Fatal(err)
	}
	legacy, err := svc.Range(ctx, date, date)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(legacy)
	if strings.Contains(string(b), `"study"`) {
		t.Fatal("legacy response gained study")
	}
	if _, err := svc.StartStudy(ctx, SourceWeb, first, EntryTenMinutes, 0, studyExam); err != nil {
		t.Fatal(err)
	}
	day, err := svc.Today(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if day.ActiveStudy == nil || day.ActiveStudy.ID != first || day.LatestPausedStudy != nil || day.Sessions[0].Study.Bookmark != nil {
		t.Fatalf("active day: %+v", day)
	}
	if _, err := svc.PauseStudy(ctx, SourceWeb, first, &Bookmark{1, StageSoru, 105}, 1); err != nil {
		t.Fatal(err)
	}
	count := auditCount(t, st)
	got, err := svc.Range(ctx, date, date)
	if err != nil {
		t.Fatal(err)
	}
	v := got[0]
	if v.State != StateToday || v.Completed || v.StudyVersion != 2 || v.Study == nil || v.Study.Status != StudyPaused {
		t.Fatalf("paused: %+v", v)
	}
	bm := v.Study.Bookmark
	if bm == nil || bm.ItemOrdinal != 1 || bm.ResumePage != 105 || bm.Item.Book != "G1" || bm.Notation != "G1·1, s. 100–110" {
		t.Fatalf("resolved bookmark: %+v", bm)
	}
	week, err := svc.Week(ctx, date)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(week.Days[0].Sessions[0], v) || week.DoneMin != 0 || week.SessionsDone != 0 {
		t.Fatalf("week differs: %+v", week)
	}
	after, err := svc.Progress(ctx, studyExam)
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatalf("partial progress changed: %v", err)
	}
	if auditCount(t, st) != count {
		t.Fatal("reads wrote audit")
	}

	// Midnight leaves yesterday's partial work visible beside today's plan.
	svc.clk = clock.Fake{Current: istanbul(t, "2026-10-27 00:01")}
	if _, err := svc.StartStudy(ctx, SourceWeb, second, EntryNormal, 0, studyExam); err != nil {
		t.Fatal(err)
	}
	day, err = svc.Today(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if day.ActiveStudy == nil || day.ActiveStudy.ID != second || day.LatestPausedStudy == nil || day.LatestPausedStudy.ID != first || day.LatestPausedStudy.State != StateMissed || len(day.Sessions) != 1 || day.Sessions[0].ID != second {
		t.Fatalf("midnight: %+v", day)
	}
	// A makeup link remains separate from the original study/bookmark.
	req := SavePlanRequest{From: "2026-10-28", To: "2026-10-28", Sessions: []SessionInput{{Date: "2026-10-28", StartTime: "19:00", Kind: KindAksam, DurationMin: 90, Items: []Item{v.Items[0]}, MakeupOf: &SlotRef{Date: "2026-10-26", StartTime: "20:30"}}}}
	if _, err := svc.SavePlan(ctx, SourceAPI, req); err != nil {
		t.Fatal(err)
	}
	day, err = svc.Today(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if day.LatestPausedStudy.State != StateMoved || day.LatestPausedStudy.MovedTo == nil || day.LatestPausedStudy.Study.Bookmark.ResumePage != 105 {
		t.Fatalf("makeup context: %+v", day.LatestPausedStudy)
	}
	completed, err := svc.Complete(ctx, SourceAPI, first)
	if err != nil || completed.Study.Status != StudyClosed || completed.Study.Bookmark.ResumePage != 105 {
		t.Fatalf("complete: %+v %v", completed, err)
	}
	if _, err := svc.Undo(ctx, SourceMCP, first); err != nil {
		t.Fatal(err)
	}
	// Rest and post-exam reads do not hide or auto-complete open work.
	for _, at := range []string{"2026-11-01 09:00", "2027-07-19 09:00"} {
		svc.clk = clock.Fake{Current: istanbul(t, at)}
		day, err = svc.Today(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if !day.IsRest || day.ActiveStudy.ID != second || day.LatestPausedStudy.ID != first || day.LatestPausedStudy.Study.Version != 4 {
			t.Fatalf("rest/post-exam: %+v", day)
		}
	}
}
