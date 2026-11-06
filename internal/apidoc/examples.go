package apidoc

import (
	"encoding/json"
	"time"

	"github.com/ibra2626/workey/internal/clock"
	"github.com/ibra2626/workey/internal/plan"
)

// pint returns a pointer to i, for the optional integer fields of an example item.
func pint(i int) *int { return &i }

// strptr returns a pointer to s, for the optional string fields of an example.
func strptr(s string) *string { return &s }

// exampleM1Source is the read view of source "M1" as the ExampleSource request
// writes it into a freshly migrated store: deterministic ids (source 1; sections
// 1–4; the three Antrenman units 1–3 in section 3). It backs both the Catalog and
// SaveSourceResult examples so they stay in step with the seeded contract run.
func exampleM1Source() plan.SourceView {
	return plan.SourceView{
		ID: 1, Key: "M1", Name: "Matematik 1", Kind: plan.SourceKindBook,
		LocationUnit: plan.LocationUnitPage, Ordinal: 1,
		Sections: []plan.SectionView{
			{ID: 1, Ordinal: 1, Name: "Sayı basamakları", LocFrom: pint(1), LocTo: pint(22), Units: []plan.UnitView{}},
			{ID: 2, Ordinal: 2, Name: "Bölme ve bölünebilme", LocFrom: pint(23), LocTo: pint(40), Units: []plan.UnitView{}},
			{ID: 3, Ordinal: 3, Name: "Üslü ifadeler", LocFrom: pint(41), LocTo: pint(58), Units: []plan.UnitView{
				{ID: 1, Ordinal: 1, Label: "Antrenman", Number: pint(1), LocFrom: pint(41), LocTo: pint(46)},
				{ID: 2, Ordinal: 2, Label: "Antrenman", Number: pint(2), LocFrom: pint(47), LocTo: pint(52)},
				{ID: 3, Ordinal: 3, Label: "Antrenman", Number: pint(3), LocFrom: pint(53), LocTo: pint(58)},
			}},
			{ID: 4, Ordinal: 4, Name: "Köklü ifadeler", LocFrom: pint(59), LocTo: pint(76), Units: []plan.UnitView{}},
		},
	}
}

// ResponseExample describes a disposable scenario on 26 October 2026, never live data.
// The contract test runs this scenario and compares every example with its real response.
func ResponseExample(name string) any {
	date := time.Date(2026, 10, 26, 9, 0, 0, 0, clock.Location)
	item := plan.ItemView{SourceKey: "M1", SourceName: "Matematik 1", SectionOrdinal: pint(1), SectionName: "Sayı basamakları", LocFrom: pint(1), LocTo: pint(10)}
	session := plan.SessionView{ID: 1, Date: "2026-10-26", StartTime: "20:30", DurationMin: 90, Kind: plan.KindAksam, State: plan.StateToday, TopicGroupKey: "temel", TopicGroupName: "Temel kavramlar", Items: []plan.ItemView{item}, Notation: "M1·1, s. 1–10"}
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
	case "Catalog":
		return plan.CatalogView{Sources: []plan.SourceView{exampleM1Source()}}
	case "SaveSourceResult":
		return plan.SaveSourceResult{Source: exampleM1Source(), Created: true}
	case "UnitResultView":
		return plan.UnitResultView{
			UnitID: 1, SourceKey: "M1", SectionOrdinal: 3, Label: "Antrenman",
			Number: pint(1), Notation: "M1·3, Antrenman 1", Done: true,
			DoneAt: strptr("2026-10-26T06:00:00Z"), Correct: pint(5), Wrong: pint(1), Blank: pint(0),
		}
	case "Progress":
		progress := plan.ProgressView{HasPlan: true, WeekNumber: 1, WeeksTotal: 38, SessionsPlanned: 1, PlannedMin: 90, DoneHours: "0", PlannedHours: "1,5", CurrentPhase: "İşlem temeli", DaysToExam: 265, ExamDate: "18 Tem", StudyDayToday: true, Groups: []plan.GroupProgress{{Key: "temel", Name: "Temel kavramlar", PlannedMin: 90}}}
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
