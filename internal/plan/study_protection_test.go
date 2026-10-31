package plan

import (
	"context"
	"reflect"
	"testing"

	"github.com/ibra2626/workey/internal/clock"
	"github.com/ibra2626/workey/internal/store"
)

func TestSavePlanStartedProtection(t *testing.T) {
	for _, status := range []string{"active", "paused", "closed"} {
		for _, change := range []string{"same", "source", "page", "extra-item", "duration", "kind", "group", "omit", "empty", "move", "move-date", "multi-replace"} {
			t.Run(status+"/"+change, func(t *testing.T) {
				ctx := context.Background()
				svc, st := newService(t, istanbul(t, "2026-10-26 09:00"))
				week := loadWeek(t, svc)
				if _, err := svc.SavePlan(ctx, SourceAPI, week); err != nil {
					t.Fatal(err)
				}
				input := week.Sessions[0]
				before := sessionAt(t, st, input.Date, input.StartTime)
				stage, ordinal, page := "konu", 0, before.Items[0].PageFrom
				study := &store.Study{SessionID: before.ID, Status: status, EntryMode: "normal", FirstStartedAt: svc.nowUTC(), LastStartedAt: svc.nowUTC(), UpdatedAt: svc.nowUTC(), ItemOrdinal: &ordinal, Stage: &stage, ResumePage: &page, Version: 1}
				if err := st.InsertStudy(ctx, st.DB(), study); err != nil {
					t.Fatal(err)
				}
				count := auditCount(t, st)
				wantSkip, wantUnchanged, wantCreate, wantDelete, wantUpdate := 0, 4, 0, 0, 0
				switch change {
				case "same":
					wantUnchanged = 5
				case "source":
					week.Sessions[0].Items[0].ChapterTitle = "Değişen temsili kaynak"
					wantSkip = 1
				case "page":
					week.Sessions[0].Items[0].PageTo++
					wantSkip = 1
				case "extra-item":
					week.Sessions[0].Items = append(week.Sessions[0].Items, week.Sessions[0].Items[0])
					wantSkip = 1
				case "duration":
					week.Sessions[0].DurationMin++
					wantSkip = 1
				case "kind":
					week.Sessions[0].Kind = KindDeneme
					wantSkip = 1
				case "group":
					week.Sessions[0].TopicGroup = ""
					wantSkip = 1
				case "omit":
					week.Sessions = week.Sessions[1:]
				case "empty":
					week.Sessions = nil
					wantDelete, wantUnchanged = 4, 0
				case "move":
					week.Sessions[0].StartTime = "19:00"
					wantCreate = 1
				case "move-date":
					week.Sessions[0].Date = "2026-11-01"
					wantCreate = 1
				case "multi-replace":
					week.Sessions[0].DurationMin++
					wantSkip = 1
					week.Sessions[1].DurationMin++
					wantUpdate = 1
					week.Sessions = week.Sessions[:2]
					wantDelete = 3
					wantUnchanged = 0
				}
				week.IdempotencyKey = status + "-" + change
				result, err := svc.SavePlan(ctx, SourceAPI, week)
				if err != nil {
					t.Fatal(err)
				}
				if len(result.Skipped) != wantSkip || len(result.Unchanged) != wantUnchanged || len(result.Created) != wantCreate || len(result.Deleted) != wantDelete || len(result.Updated) != wantUpdate {
					t.Fatalf("unexpected result: %+v", result)
				}
				if wantSkip > 0 && result.Skipped[0].Reason != SkipStarted {
					t.Fatalf("skip: %+v", result.Skipped)
				}
				after := sessionAt(t, st, input.Date, input.StartTime)
				if !reflect.DeepEqual(before, after) {
					t.Fatalf("started session changed: before=%+v after=%+v", before, after)
				}
				read, err := st.StudyBySession(ctx, st.DB(), before.ID)
				if err != nil || !reflect.DeepEqual(read, study) {
					t.Fatalf("study changed: %+v, %v", read, err)
				}
				if got := auditCount(t, st); got != count+wantCreate+wantUpdate+wantDelete {
					t.Fatalf("audit = %d", got)
				}
				count = auditCount(t, st)
				replay, err := svc.SavePlan(ctx, SourceAPI, week)
				if err != nil || !reflect.DeepEqual(result, replay) || auditCount(t, st) != count {
					t.Fatalf("replay: %+v, %v", replay, err)
				}
				week.IdempotencyKey = ""
				if _, err := svc.SavePlan(ctx, SourceAPI, week); err != nil {
					t.Fatal(err)
				}
				if auditCount(t, st) != count {
					t.Fatal("repeat added audit")
				}
			})
		}
	}
}

func TestStartedSkipPrecedence(t *testing.T) {
	ctx := context.Background()
	svc, st := newService(t, istanbul(t, "2026-10-26 09:00"))
	week := loadWeek(t, svc)
	if _, err := svc.SavePlan(ctx, SourceAPI, week); err != nil {
		t.Fatal(err)
	}
	for i := range 2 {
		ss := sessionAt(t, st, week.Sessions[i].Date, week.Sessions[i].StartTime)
		v := &store.Study{SessionID: ss.ID, Status: "paused", EntryMode: "normal", FirstStartedAt: svc.nowUTC(), LastStartedAt: svc.nowUTC(), UpdatedAt: svc.nowUTC(), Version: 1}
		if err := st.InsertStudy(ctx, st.DB(), v); err != nil {
			t.Fatal(err)
		}
	}
	svc.clk = clock.Fake{Current: istanbul(t, "2026-10-27 09:00")}
	ss := sessionAt(t, st, week.Sessions[1].Date, week.Sessions[1].StartTime)
	now, source := svc.nowUTC(), "web"
	if err := st.SetSessionCompletion(ctx, st.DB(), ss.ID, &now, &source, now); err != nil {
		t.Fatal(err)
	}
	result, err := svc.SavePlan(ctx, SourceAPI, week)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Skipped) != 2 || result.Skipped[0].Reason != SkipPastDate || result.Skipped[1].Reason != SkipCompleted {
		t.Fatalf("precedence: %+v", result)
	}
}
