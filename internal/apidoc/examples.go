package apidoc

import (
	"encoding/json"
	"time"

	"github.com/ibra2626/workey/internal/clock"
	"github.com/ibra2626/workey/internal/plan"
)

// ResponseExample describes a disposable scenario on 26 October 2026, never live data.
// The contract test runs this scenario and compares every example with its real response.
func ResponseExample(name string) any {
	date := time.Date(2026, 10, 26, 9, 0, 0, 0, clock.Location)
	item := plan.Item{Book: "M1", ChapterNo: 1, ChapterTitle: "Temel kavramlar", PageFrom: 1, PageTo: 10}
	session := plan.SessionView{ID: 1, Date: "2026-10-26", StartTime: "20:30", DurationMin: 90, Kind: plan.KindAksam, State: plan.StateToday, TopicGroupKey: "temel", TopicGroupName: "Temel kavramlar", Items: []plan.Item{item}, Notation: "M1·1, s. 1–10"}
	switch name {
	case "AgentGuide":
		return plan.New(nil, clock.Fake{Current: date}).AgentGuide(clock.Date{Year: 2027, Month: 7, Day: 18})
	case "Plan":
		return struct {
			Sessions []plan.SessionView `json:"sessions"`
		}{[]plan.SessionView{session}}
	case "Session":
		return session
	case "PlanResult":
		return plan.SavePlanResult{Created: []plan.SessionRef{{ID: 1, Date: session.Date, StartTime: session.StartTime}}}
	case "OutlineResult":
		return plan.SaveOutlineResult{Phases: plan.OutlineChange{Created: []string{"faz-1"}}, TopicGroups: plan.OutlineChange{Created: []string{"temel"}}}
	case "Outline":
		var outline plan.OutlineView
		_ = json.Unmarshal([]byte(ExampleOutline), &outline)
		return outline
	case "Day":
		return plan.DayView{Date: session.Date, Weekday: "Pazartesi", Heading: "Pazartesi, 26 Ekim", Sessions: []plan.SessionView{session}}
	case "Week":
		week := plan.WeekView{Start: session.Date, End: "2026-11-01", RangeLabel: "26 Eki–1 Kas", WeekNumber: 1, PhaseName: "İşlem temeli", PlannedMin: 90, SessionsPlanned: 1}
		for i := 0; i < 7; i++ {
			d := date.AddDate(0, 0, i)
			ds := d.Format("2006-01-02")
			row := plan.DayRow{Date: ds, Weekday: plan.WeekdayTR(ds), WeekdayShort: plan.WeekdayShortTR(ds), DayNum: d.Day(), IsRest: i != 0}
			if i == 0 {
				row.Sessions = []plan.SessionView{session}
			}
			week.Days = append(week.Days, row)
		}
		return week
	case "Progress":
		progress := plan.ProgressView{HasPlan: true, WeekNumber: 1, WeeksTotal: 38, SessionsPlanned: 1, PlannedMin: 90, DoneHours: "0", PlannedHours: "1,5", CurrentPhase: "İşlem temeli", DaysToExam: 265, ExamDate: "18 Tem", Groups: []plan.GroupProgress{{Key: "temel", Name: "Temel kavramlar", PlannedMin: 90}}}
		for i := 1; i <= 38; i++ {
			cell := plan.WeekCell{WeekNumber: i}
			if i <= 7 {
				cell.PhaseKey = "faz-1"
				cell.PhaseName = "İşlem temeli"
			}
			if i == 1 {
				cell.PlannedMin = 90
			}
			progress.Strip = append(progress.Strip, cell)
		}
		return progress
	}
	return nil
}

func OperationResponseExample(op Operation) any {
	example := ResponseExample(op.Output)
	if op.Method == "POST" && op.Output == "Session" {
		session := example.(plan.SessionView)
		session.Completed = true
		session.State = plan.StateDone
		return session
	}
	return example
}

func OperationResponseJSON(op Operation) string {
	b, _ := json.MarshalIndent(OperationResponseExample(op), "", "  ")
	return string(b)
}

// StudyResponseExample is a separate synthetic paused-work scenario. Keeping
// the legacy examples unchanged demonstrates that study is optional.
func StudyResponseExample() plan.SessionView {
	s := ResponseExample("Session").(plan.SessionView)
	s.Study = &plan.StudyView{Status: plan.StudyPaused, EntryMode: plan.EntryTenMinutes,
		FirstStartedAt: "2026-10-26T17:30:00Z", LastStartedAt: "2026-10-26T17:30:00Z",
		UpdatedAt: "2026-10-26T17:40:00Z", Version: 2,
		Bookmark: &plan.BookmarkView{ItemOrdinal: 0, Stage: plan.StageSoru,
			ResumePage: 5, Item: s.Items[0], Notation: s.Notation}}
	return s
}

func StartedResultExample() plan.SavePlanResult {
	return plan.SavePlanResult{Skipped: []plan.SkippedSession{{Date: "2026-10-26", StartTime: "20:30", Reason: plan.SkipStarted}}}
}
