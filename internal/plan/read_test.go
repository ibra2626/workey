package plan

import (
	"context"
	"testing"

	"github.com/ibra2626/workey/internal/clock"
	"github.com/ibra2626/workey/internal/store"
)

// setupMovedWeek saves the sample week (all five sessions created on Monday
// morning), completes Monday, then — from Wednesday evening — links the
// Saturday long block as the makeup for the missed Tuesday. It returns a
// service whose clock is the Wednesday and the shared store, matching the
// mockup state: 1 done, 1 moved to Saturday.
func setupMovedWeek(t *testing.T) (*Service, *store.Store) {
	t.Helper()
	ctx := context.Background()
	svc, st := newService(t, istanbul(t, "2026-10-26 09:00"))
	week := loadWeek(t, svc)
	if _, err := svc.SavePlan(ctx, SourceAPI, week); err != nil {
		t.Fatalf("SavePlan: %v", err)
	}
	mon := sessionAt(t, st, "2026-10-26", "20:30")
	if _, err := svc.Complete(ctx, SourceWeb, mon.ID); err != nil {
		t.Fatalf("Complete Monday: %v", err)
	}

	// From Wednesday, add the makeup link onto Saturday's block.
	svc2 := New(st, clock.Fake{Current: istanbul(t, "2026-10-28 21:00")})
	moved := week
	moved.Sessions = append([]SessionInput(nil), week.Sessions...)
	sat := moved.Sessions[len(moved.Sessions)-1]
	sat.MakeupOf = &SlotRef{Date: "2026-10-27", StartTime: "20:30"}
	moved.Sessions[len(moved.Sessions)-1] = sat
	if _, err := svc2.SavePlan(ctx, SourceAPI, moved); err != nil {
		t.Fatalf("SavePlan makeup: %v", err)
	}
	return svc2, st
}

func TestWeekReproducesMockupFigures(t *testing.T) {
	ctx := context.Background()
	svc, _ := setupMovedWeek(t)

	wk, err := svc.Week(ctx, clock.Date{Year: 2026, Month: 10, Day: 28})
	if err != nil {
		t.Fatalf("Week: %v", err)
	}
	if wk.PlannedMin != 540 || wk.DoneMin != 90 {
		t.Errorf("minutes = %d / %d, want 90 / 540", wk.DoneMin, wk.PlannedMin)
	}
	if wk.SessionsDone != 1 || wk.SessionsPlanned != 5 {
		t.Errorf("sessions = %d / %d, want 1 / 5", wk.SessionsDone, wk.SessionsPlanned)
	}
	if wk.MovedCount != 1 {
		t.Errorf("moved = %d, want 1", wk.MovedCount)
	}
	if wk.WeekNumber != 1 {
		t.Errorf("week number = %d, want 1", wk.WeekNumber)
	}
	if wk.RangeLabel != "26 Eki–1 Kas" {
		t.Errorf("range = %q, want 26 Eki–1 Kas", wk.RangeLabel)
	}
	if wk.PhaseName != "Faz 1 · İşlem temeli" {
		t.Errorf("phase = %q", wk.PhaseName)
	}
	if len(wk.Days) != 7 {
		t.Fatalf("days = %d, want 7", len(wk.Days))
	}

	// Monday done, Tuesday moved to Saturday, Wednesday today, Friday rest.
	if got := wk.Days[0].Sessions[0].State; got != StateDone {
		t.Errorf("Monday state = %q, want done", got)
	}
	tue := wk.Days[1].Sessions[0]
	if tue.State != StateMoved {
		t.Errorf("Tuesday state = %q, want moved", tue.State)
	}
	if tue.MovedTo == nil || tue.MovedTo.Weekday != "Cumartesi" {
		t.Errorf("Tuesday MovedTo = %+v, want weekday Cumartesi", tue.MovedTo)
	}
	if got := wk.Days[2].Sessions[0].State; got != StateToday {
		t.Errorf("Wednesday state = %q, want today", got)
	}
	if !wk.Days[4].IsRest {
		t.Errorf("Friday should be a rest day")
	}

	if len(wk.Notes) != 1 {
		t.Fatalf("notes = %v, want one line", wk.Notes)
	}
	const wantNote = "Salı yapılamadı; M1·2 cumartesi bloğuna eklendi."
	if wk.Notes[0] != wantNote {
		t.Errorf("note = %q, want %q", wk.Notes[0], wantNote)
	}
}

func TestDayBugun(t *testing.T) {
	ctx := context.Background()
	svc, _ := setupMovedWeek(t)

	day, err := svc.Day(ctx, clock.Date{Year: 2026, Month: 10, Day: 28})
	if err != nil {
		t.Fatalf("Day: %v", err)
	}
	if day.Heading != "Çarşamba, 28 Ekim" {
		t.Errorf("heading = %q", day.Heading)
	}
	if day.IsRest || len(day.Sessions) != 1 {
		t.Fatalf("today sessions = %+v", day.Sessions)
	}
	if day.Sessions[0].State != StateToday {
		t.Errorf("today state = %q, want today", day.Sessions[0].State)
	}
	if day.Tomorrow == nil || day.Tomorrow.Notation != "M1·4, s. 59"+enDash+"76" {
		t.Errorf("tomorrow = %+v, want M1·4, s. 59–76", day.Tomorrow)
	}
	if len(day.Notes) != 1 {
		t.Errorf("notes = %v, want one line", day.Notes)
	}
}

func TestDayRestDay(t *testing.T) {
	ctx := context.Background()
	svc, _ := setupMovedWeek(t)

	day, err := svc.Day(ctx, clock.Date{Year: 2026, Month: 10, Day: 30}) // Friday, no session
	if err != nil {
		t.Fatalf("Day: %v", err)
	}
	if !day.IsRest || len(day.Sessions) != 0 {
		t.Errorf("Friday should be a rest day, got %+v", day.Sessions)
	}
}

func TestProgress(t *testing.T) {
	ctx := context.Background()
	svc, st := newService(t, istanbul(t, "2026-10-26 09:00"))
	week := loadWeek(t, svc)
	if _, err := svc.SavePlan(ctx, SourceAPI, week); err != nil {
		t.Fatalf("SavePlan: %v", err)
	}
	mon := sessionAt(t, st, "2026-10-26", "20:30")
	if _, err := svc.Complete(ctx, SourceWeb, mon.ID); err != nil {
		t.Fatalf("Complete: %v", err)
	}

	exam := clock.Date{Year: 2027, Month: 7, Day: 18}
	prog, err := svc.Progress(ctx, exam)
	if err != nil {
		t.Fatalf("Progress: %v", err)
	}
	if !prog.HasPlan {
		t.Fatalf("HasPlan = false, want true")
	}
	if prog.WeekNumber != 1 || prog.WeeksTotal != 38 {
		t.Errorf("weeks = %d / %d, want 1 / 38", prog.WeekNumber, prog.WeeksTotal)
	}
	if prog.SessionsDone != 1 || prog.SessionsPlanned != 5 {
		t.Errorf("sessions = %d / %d, want 1 / 5", prog.SessionsDone, prog.SessionsPlanned)
	}
	if prog.DoneHours != "1,5" || prog.PlannedHours != "9" {
		t.Errorf("hours = %q / %q, want 1,5 / 9", prog.DoneHours, prog.PlannedHours)
	}
	if prog.CurrentPhase != "Faz 1 · İşlem temeli" {
		t.Errorf("current phase = %q", prog.CurrentPhase)
	}
	if prog.ExamDate != "18 Tem" {
		t.Errorf("exam label = %q, want 18 Tem", prog.ExamDate)
	}
	if want := civilDays(clock.Date{Year: 2026, Month: 10, Day: 26}, exam); prog.DaysToExam != want {
		t.Errorf("days to exam = %d, want %d", prog.DaysToExam, want)
	}
	if len(prog.Groups) != 5 {
		t.Fatalf("groups = %d, want 5", len(prog.Groups))
	}
	g := prog.Groups[0]
	if g.Key != "temel-kavramlar" {
		t.Fatalf("first group = %q, want temel-kavramlar", g.Key)
	}
	if g.PlannedMin != 180 || g.DoneMin != 90 || g.Percent != 50 {
		t.Errorf("temel-kavramlar = done %d / plan %d (%d%%), want 90 / 180 (50%%)", g.DoneMin, g.PlannedMin, g.Percent)
	}
	if len(prog.Strip) != 38 {
		t.Fatalf("strip = %d, want 38", len(prog.Strip))
	}
	if prog.Strip[0].WeekNumber != 1 || prog.Strip[0].PlannedMin != 540 || prog.Strip[0].DoneMin != 90 {
		t.Errorf("strip[0] = %+v, want week 1, 90/540", prog.Strip[0])
	}
	if prog.Strip[0].PhaseName != "Faz 1 · İşlem temeli" {
		t.Errorf("strip[0] phase = %q", prog.Strip[0].PhaseName)
	}
}

func TestProgressNoPlan(t *testing.T) {
	ctx := context.Background()
	svc, _ := newService(t, istanbul(t, "2026-10-26 09:00"))
	prog, err := svc.Progress(ctx, clock.Date{Year: 2027, Month: 7, Day: 18})
	if err != nil {
		t.Fatalf("Progress: %v", err)
	}
	if prog.HasPlan {
		t.Errorf("HasPlan = true on an empty plan")
	}
	if len(prog.Strip) != 0 || prog.WeekNumber != 0 {
		t.Errorf("empty plan should have no strip and no week number: %+v", prog)
	}
}
