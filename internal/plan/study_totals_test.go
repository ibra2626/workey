package plan

import (
	"context"
	"testing"

	"github.com/ibra2626/workey/internal/clock"
)

func TestPartialOriginalAndMakeupTotals(t *testing.T) {
	ctx := context.Background()
	svc, st, id, _ := studyFixture(t)
	if _, err := svc.StartStudy(ctx, SourceWeb, id, EntryTenMinutes, 0, studyExam); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.PauseStudy(ctx, SourceWeb, id, &Bookmark{1, StageSoru, 105}, 1); err != nil {
		t.Fatal(err)
	}
	svc.clk = clock.Fake{Current: istanbul(t, "2026-10-28 21:00")}
	rows, err := svc.Range(ctx, clock.Date{Year: 2026, Month: 10, Day: 26}, svc.TodayDate())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SavePlan(ctx, SourceAPI, SavePlanRequest{From: "2026-10-28", To: "2026-10-28", Sessions: []SessionInput{{Date: "2026-10-28", StartTime: "19:00", DurationMin: 90, Kind: KindAksam, Items: rows[0].Items, MakeupOf: &SlotRef{Date: "2026-10-26", StartTime: "20:30"}}}}); err != nil {
		t.Fatal(err)
	}
	makeup := sessionAt(t, st, "2026-10-28", "19:00").ID
	check := func(want int) {
		t.Helper()
		wk, err := svc.Week(ctx, svc.TodayDate())
		if err != nil || wk.DoneMin != want {
			t.Fatalf("week: %+v %v want=%d", wk, err, want)
		}
		p, err := svc.Progress(ctx, studyExam)
		if err != nil || p.DoneMin != want {
			t.Fatalf("progress: %+v %v want=%d", p, err, want)
		}
	}
	check(0)
	if _, err := svc.Complete(ctx, SourceWeb, id); err != nil {
		t.Fatal(err)
	}
	check(90)
	if _, err := svc.Complete(ctx, SourceWeb, makeup); err != nil {
		t.Fatal(err)
	}
	check(180)
	if _, err := svc.Undo(ctx, SourceWeb, id); err != nil {
		t.Fatal(err)
	}
	check(90)
	v := storedStudy(t, st, id)
	if v.Status != StudyPaused || v.ResumePage == nil || *v.ResumePage != 105 {
		t.Fatal("undo lost bookmark")
	}
}
