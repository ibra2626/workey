package views

import (
	"strconv"
	"strings"
	"time"

	"github.com/ibra2626/workey/internal/plan"
)

// PlanPage is the Plan screen model: the whole saved plan as month calendars.
type PlanPage struct {
	Overview   *plan.PlanOverview
	WeekNumber int // plan week of today, 0 outside the plan
	DaysToExam int
	HasExam    bool
	Month      string // ?ay= "2026-11"; empty or unknown shows the current month
	CSRF       string
}

// calDay is one cell of a month grid. Blank cells pad the first and last week.
type calDay struct {
	Date     string
	Day      int
	Blank    bool
	IsToday  bool
	IsPast   bool
	InPlan   bool   // inside a phase span
	Phase    string // phase name without its "Faz N" prefix
	Col      int    // 0 = Monday .. 6 = Sunday
	LastRows bool   // in the bottom two rows, so the detail opens upward
	Sessions []plan.SessionView
}

// calMonth is one month of the Plan calendar.
type calMonth struct {
	ID        string // "2026-11"
	Title     string // "Kasım 2026"
	IsCurrent bool
	Phases    []string
	Sessions  int
	Done      int
	Days      []calDay
}

var calMonthsTR = [13]string{"", "Ocak", "Şubat", "Mart", "Nisan", "Mayıs", "Haziran", "Temmuz", "Ağustos", "Eylül", "Ekim", "Kasım", "Aralık"}

var calMonthsShortTR = [13]string{"", "Oca", "Şub", "Mar", "Nis", "May", "Haz", "Tem", "Ağu", "Eyl", "Eki", "Kas", "Ara"}

// planCalendar returns every plan month and the index of the one to show: the
// ?ay= month when it is in the plan, otherwise today's month, otherwise the
// nearest end of the plan.
func planCalendar(p PlanPage) ([]calMonth, int) {
	months := planMonths(p.Overview)
	if len(months) == 0 {
		return nil, 0
	}
	for i, m := range months {
		if m.ID == p.Month {
			return months, i
		}
	}
	for i, m := range months {
		if m.IsCurrent {
			return months, i
		}
	}
	if p.Overview.Today > months[len(months)-1].ID {
		return months, len(months) - 1
	}
	return months, 0
}

// calStripLabel is a month's short name in the month strip; the first month
// and every January carry the year.
func calStripLabel(months []calMonth, i int) string {
	m := months[i]
	label := calMonthsShortTR[monthOf(m.ID+"-01")]
	if i == 0 || strings.HasSuffix(m.ID, "-01") {
		label += " " + m.ID[:4]
	}
	return label
}

// calStripFill is a month's completed share for the strip's small track.
func calStripFill(m calMonth) string {
	if m.Sessions == 0 {
		return "width:0%"
	}
	return "width:" + strconv.Itoa(m.Done*100/m.Sessions) + "%"
}

func planMonthHref(id string) string {
	return "/plan?ay=" + id
}

// calWeekdays heads the grid columns, Monday first.
var calWeekdays = []string{"Pzt", "Sal", "Çar", "Per", "Cum", "Cmt", "Paz"}

// planMonths lays the saved plan out as month grids, from the first phase or
// session to the last one.
func planMonths(ov *plan.PlanOverview) []calMonth {
	byDate := map[string][]plan.SessionView{}
	first, last := "", ""
	widen := func(d string) {
		if d == "" {
			return
		}
		if first == "" || d < first {
			first = d
		}
		if last == "" || d > last {
			last = d
		}
	}
	for _, ph := range ov.Phases {
		widen(ph.StartDate)
		widen(ph.EndDate)
		for _, w := range ph.Weeks {
			for _, sv := range w.Sessions {
				byDate[sv.Date] = append(byDate[sv.Date], sv)
				widen(sv.Date)
			}
		}
	}
	start, err1 := time.Parse(time.DateOnly, first)
	end, err2 := time.Parse(time.DateOnly, last)
	if err1 != nil || err2 != nil {
		return nil
	}
	phaseAt := func(d string) (string, bool) {
		for _, ph := range ov.Phases {
			if ph.StartDate != "" && ph.StartDate <= d && d <= ph.EndDate {
				return phaseTail(ph.Name), true
			}
		}
		return "", false
	}

	var months []calMonth
	for m := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC); !m.After(end); m = m.AddDate(0, 1, 0) {
		cm := calMonth{
			ID:        m.Format("2006-01"),
			Title:     calMonthsTR[m.Month()] + " " + strconv.Itoa(m.Year()),
			IsCurrent: strings.HasPrefix(ov.Today, m.Format("2006-01")),
		}
		lead := (int(m.Weekday()) + 6) % 7
		for i := 0; i < lead; i++ {
			cm.Days = append(cm.Days, calDay{Blank: true, Col: i})
		}
		seen := map[string]bool{}
		for d := m; d.Month() == m.Month(); d = d.AddDate(0, 0, 1) {
			ds := d.Format(time.DateOnly)
			phase, in := phaseAt(ds)
			if in && !seen[phase] {
				seen[phase] = true
				cm.Phases = append(cm.Phases, phase)
			}
			day := calDay{
				Date: ds, Day: d.Day(), Col: (int(d.Weekday()) + 6) % 7,
				IsToday: ds == ov.Today, IsPast: ds < ov.Today,
				InPlan: in, Phase: phase, Sessions: byDate[ds],
			}
			for _, sv := range day.Sessions {
				cm.Sessions++
				if sv.Completed {
					cm.Done++
				}
			}
			cm.Days = append(cm.Days, day)
		}
		for len(cm.Days)%7 != 0 {
			cm.Days = append(cm.Days, calDay{Blank: true, Col: len(cm.Days) % 7})
		}
		rows := len(cm.Days) / 7
		for i := range cm.Days {
			cm.Days[i].LastRows = rows > 3 && i/7 >= rows-2
		}
		months = append(months, cm)
	}
	return months
}

// calDayClass carries the cell's state and which way its detail opens.
func calDayClass(d calDay) string {
	c := "cal-day"
	if d.Blank {
		return c + " cal-blank"
	}
	if !d.InPlan {
		c += " cal-out"
	}
	if d.IsToday {
		c += " cal-today"
	}
	if d.IsPast {
		c += " cal-past"
	}
	if d.Col >= 4 {
		c += " cal-pop-left"
	}
	if d.LastRows {
		c += " cal-pop-up"
	}
	return c
}

// calChipState names a session's state for the CSS hooks.
func calChipState(sv plan.SessionView) string {
	if partialStudy(sv) {
		return "partial"
	}
	return string(sv.State)
}

// calChipTitle is the short chip text: the chapter, or the block kind.
func calChipTitle(sv plan.SessionView) string {
	switch sv.Kind {
	case plan.KindDeneme:
		return "Deneme"
	case plan.KindUzunBlok:
		return "Uzun blok"
	}
	if t := chapterTitle(sv); t != "" {
		return t
	}
	return sv.Notation
}

// calDayLabel is the accessible name of a cell, e.g. "31 Ekim, 1 seans".
func calDayLabel(d calDay) string {
	label := strconv.Itoa(d.Day) + " " + calMonthsTR[monthOf(d.Date)]
	if n := len(d.Sessions); n > 0 {
		label += ", " + strconv.Itoa(n) + " seans"
	}
	return label
}

func monthOf(date string) int {
	if len(date) < 7 {
		return 0
	}
	m, _ := strconv.Atoi(date[5:7])
	return m
}

// planKindLabel names the session kind for the detail panel.
func planKindLabel(sv plan.SessionView) string {
	switch sv.Kind {
	case plan.KindDeneme:
		return "Deneme"
	case plan.KindUzunBlok:
		return "Uzun blok"
	default:
		return "Akşam seansı"
	}
}

// planRowTitle names a session in the detail panel by its chapter.
func planRowTitle(sv plan.SessionView) string {
	if t := chapterTitle(sv); t != "" {
		return t
	}
	return planKindLabel(sv)
}

// planPhaseLine renders a phase's span, e.g. "27 Eki–13 Ara · H1–H7".
func planPhaseLine(ph plan.OverviewPhase) string {
	var parts []string
	if ph.RangeLabel != "" {
		parts = append(parts, ph.RangeLabel)
	}
	if ph.FirstWeek > 0 && ph.LastWeek > 0 {
		parts = append(parts, "H"+strconv.Itoa(ph.FirstWeek)+"–H"+strconv.Itoa(ph.LastWeek))
	}
	return strings.Join(parts, " "+middleDot+" ")
}

// planPhaseTitle is the phase name without its "Faz N" prefix; sessions
// outside every phase are grouped as "Dönem dışı".
func planPhaseTitle(ph plan.OverviewPhase) string {
	if ph.Name == "" {
		return "Dönem dışı"
	}
	return phaseTail(ph.Name)
}
