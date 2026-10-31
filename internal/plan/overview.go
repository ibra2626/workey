package plan

import (
	"context"

	"github.com/ibra2626/workey/internal/clock"
)

// OverviewWeek is one calendar week (Monday start) of the whole-plan view.
type OverviewWeek struct {
	Start      string        `json:"start"`       // Monday, 'YYYY-MM-DD'
	Number     int           `json:"number"`      // plan week 1..38, 0 outside the plan
	RangeLabel string        `json:"range_label"` // "26 Eki–1 Kas"
	IsCurrent  bool          `json:"is_current"`
	Sessions   []SessionView `json:"sessions"`
	PlannedMin int           `json:"planned_min"`
	DoneMin    int           `json:"done_min"`
}

// OverviewPhase is one outline phase with its weeks. Sessions outside every
// phase land in a trailing phase with an empty Name.
type OverviewPhase struct {
	Key             string         `json:"key,omitempty"`
	Name            string         `json:"name"`
	RangeLabel      string         `json:"range_label,omitempty"` // "27 Eki–13 Ara"
	StartDate       string         `json:"start_date,omitempty"`
	EndDate         string         `json:"end_date,omitempty"`
	FirstWeek       int            `json:"first_week"`
	LastWeek        int            `json:"last_week"`
	IsCurrent       bool           `json:"is_current"`
	IsPast          bool           `json:"is_past"`
	TopicGroups     []string       `json:"topic_groups"` // distinct names in first-seen order
	Weeks           []OverviewWeek `json:"weeks"`
	SessionsPlanned int            `json:"sessions_planned"`
	SessionsDone    int            `json:"sessions_done"`
}

// PlanOverview is the whole saved plan grouped by phase and week, for the Plan
// screen. Counts follow the same rules as Progress: planned minutes of
// completed sessions, never measured time.
type PlanOverview struct {
	HasPlan         bool            `json:"has_plan"`
	Phases          []OverviewPhase `json:"phases"`
	SessionsPlanned int             `json:"sessions_planned"`
	SessionsDone    int             `json:"sessions_done"`
	PlannedMin      int             `json:"planned_min"`
	DoneMin         int             `json:"done_min"`
	CurrentWeek     string          `json:"current_week"` // Monday of today's week
	Today           string          `json:"today"`
}

// Overview returns every saved session grouped by outline phase and week.
// Phases without sessions are kept so the whole plan period stays visible.
func (s *Service) Overview(ctx context.Context) (*PlanOverview, error) {
	db := s.st.DB()
	phases, err := s.st.ListPhases(ctx, db)
	if err != nil {
		return nil, err
	}
	all, err := s.st.AllSessions(ctx, db)
	if err != nil {
		return nil, err
	}
	today := clock.Today(s.clk)
	current := clock.WeekStart(today).String()
	view := &PlanOverview{HasPlan: len(phases) > 0 || len(all) > 0, CurrentWeek: current, Today: today.String()}

	var sessions []SessionView
	if len(all) > 0 {
		from, _ := ParseDate(all[0].Date)
		to, _ := ParseDate(all[len(all)-1].Date)
		if sessions, err = s.Range(ctx, from, to); err != nil {
			return nil, err
		}
	}

	out := make([]OverviewPhase, len(phases))
	for i, p := range phases {
		out[i] = OverviewPhase{
			Key:        p.Key,
			Name:       p.Name,
			RangeLabel: FormatWeekRange(p.StartDate, p.EndDate),
			StartDate:  p.StartDate,
			EndDate:    p.EndDate,
			IsCurrent:  p.StartDate <= today.String() && today.String() <= p.EndDate,
			IsPast:     p.EndDate < today.String(),
		}
		if d, ok := ParseDate(p.StartDate); ok {
			out[i].FirstWeek = weekNumber(phases, clock.WeekStart(d))
		}
		if d, ok := ParseDate(p.EndDate); ok {
			out[i].LastWeek = weekNumber(phases, clock.WeekStart(d))
		}
	}
	var loose OverviewPhase
	seenGroup := map[*OverviewPhase]map[string]bool{}

	for _, sv := range sessions {
		target := &loose
		for i := range phases {
			if phases[i].StartDate <= sv.Date && sv.Date <= phases[i].EndDate {
				target = &out[i]
				break
			}
		}
		d, ok := ParseDate(sv.Date)
		if !ok {
			continue
		}
		ws := clock.WeekStart(d)
		if n := len(target.Weeks); n == 0 || target.Weeks[n-1].Start != ws.String() {
			target.Weeks = append(target.Weeks, OverviewWeek{
				Start:      ws.String(),
				Number:     weekNumber(phases, ws),
				RangeLabel: FormatWeekRange(ws.String(), addDays(ws, 6).String()),
				IsCurrent:  ws.String() == current,
			})
		}
		w := &target.Weeks[len(target.Weeks)-1]
		w.Sessions = append(w.Sessions, sv)
		target.SessionsPlanned++
		view.SessionsPlanned++
		w.PlannedMin += sv.DurationMin
		view.PlannedMin += sv.DurationMin
		if sv.Completed {
			w.DoneMin += sv.DurationMin
			view.DoneMin += sv.DurationMin
			target.SessionsDone++
			view.SessionsDone++
		}
		if sv.TopicGroupName != "" {
			if seenGroup[target] == nil {
				seenGroup[target] = map[string]bool{}
			}
			if !seenGroup[target][sv.TopicGroupName] {
				seenGroup[target][sv.TopicGroupName] = true
				target.TopicGroups = append(target.TopicGroups, sv.TopicGroupName)
			}
		}
	}
	view.Phases = out
	if len(loose.Weeks) > 0 {
		view.Phases = append(view.Phases, loose)
	}
	return view, nil
}
