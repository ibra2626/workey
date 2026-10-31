package plan

import (
	"context"
	"database/sql"
	"math"
	"time"

	"github.com/ibra2626/workey/internal/clock"
	"github.com/ibra2626/workey/internal/store"
)

// WeeksTotal is the fixed length of the v1 plan: 38 weeks from the first phase's
// Monday to the week before the exam (PRODUCT.md "Operating Context").
const WeeksTotal = 38

// SessionState is a session's derived status for a given "today" (Istanbul).
type SessionState string

const (
	StatePlanned SessionState = "planned" // dated after today, not done
	StateToday   SessionState = "today"   // dated today, not done
	StateDone    SessionState = "done"    // completed
	StateMissed  SessionState = "missed"  // past, not done, no makeup
	StateMoved   SessionState = "moved"   // past, not done, referenced by a makeup session
)

// MovedInfo names the makeup session a missed session was moved to.
type MovedInfo struct {
	Date    string `json:"date"`    // 'YYYY-MM-DD' of the makeup session
	Weekday string `json:"weekday"` // capitalised Turkish weekday, e.g. "Cumartesi"
}

// SessionView is a session prepared for a screen: its stored fields, its derived
// state, the resolved topic-group name and the canonical item notation.
type SessionView struct {
	Study          *StudyView   `json:"study,omitempty" jsonschema:"Optional read-only study context. Absence is not proof of no work. Partial work does not count as completion."`
	StudyVersion   int64        `json:"-"` // internal web form concurrency token
	ID             int64        `json:"id"`
	Date           string       `json:"date"`
	StartTime      string       `json:"start_time"`
	DurationMin    int          `json:"duration_min"`
	Kind           string       `json:"kind"`
	Completed      bool         `json:"completed"`
	State          SessionState `json:"state"`
	TopicGroupKey  string       `json:"topic_group_key,omitempty"`
	TopicGroupName string       `json:"topic_group_name,omitempty"`
	Items          []Item       `json:"items"`
	Notation       string       `json:"notation"`
	MovedTo        *MovedInfo   `json:"moved_to,omitempty"`
}

// DayView is the Bugün model for one date: the day's sessions, tomorrow's first
// session and the week's generated note lines.
type DayView struct {
	ActiveStudy       *SessionView  `json:"active_study,omitempty" jsonschema:"Currently active work, even outside this date; takes priority without changing its schedule."`
	LatestPausedStudy *SessionView  `json:"latest_paused_study,omitempty" jsonschema:"Most recently paused work, even outside this date. Resuming is an explicit choice; today remains visible. May have moved_to."`
	Date              string        `json:"date"`
	Weekday           string        `json:"weekday"` // "Çarşamba"
	Heading           string        `json:"heading"` // "Çarşamba, 28 Ekim"
	IsRest            bool          `json:"is_rest"` // no sessions this day
	Sessions          []SessionView `json:"sessions"`
	Tomorrow          *SessionView  `json:"tomorrow,omitempty"`
	Notes             []string      `json:"notes"`
}

// DayRow is one day inside a WeekView, rest days included.
type DayRow struct {
	Date         string        `json:"date"`
	Weekday      string        `json:"weekday"`       // "Pazartesi"
	WeekdayShort string        `json:"weekday_short"` // "Pzt"
	DayNum       int           `json:"day_num"`       // 26
	IsRest       bool          `json:"is_rest"`
	Sessions     []SessionView `json:"sessions"`
}

// WeekView is the Hafta model: seven rows plus the week's totals, plan position
// and note lines.
type WeekView struct {
	Start           string   `json:"start"`       // Monday 'YYYY-MM-DD'
	End             string   `json:"end"`         // Sunday
	RangeLabel      string   `json:"range_label"` // "26 Eki–1 Kas"
	WeekNumber      int      `json:"week_number"` // plan week 1..38, 0 if outside the plan
	PhaseName       string   `json:"phase_name,omitempty"`
	Days            []DayRow `json:"days"`
	PlannedMin      int      `json:"planned_min"`
	DoneMin         int      `json:"done_min"`
	SessionsPlanned int      `json:"sessions_planned"`
	SessionsDone    int      `json:"sessions_done"`
	MovedCount      int      `json:"moved_count"`
	Notes           []string `json:"notes"`
}

// GroupProgress is one topic group's completion and DGS question weight.
type GroupProgress struct {
	Key          string  `json:"key"`
	Name         string  `json:"name"`
	DGSMin       *int    `json:"dgs_min,omitempty"`
	DGSMax       *int    `json:"dgs_max,omitempty"`
	PlannedMin   int     `json:"planned_min"`
	DoneMin      int     `json:"done_min"`
	DoneFraction float64 `json:"done_fraction"` // 0..1
	Percent      int     `json:"percent"`       // rounded done_fraction*100
}

// WeekCell is one column of the 38-week progress strip.
type WeekCell struct {
	WeekNumber   int     `json:"week_number"` // 1..38
	PhaseKey     string  `json:"phase_key,omitempty"`
	PhaseName    string  `json:"phase_name,omitempty"`
	PlannedMin   int     `json:"planned_min"`
	DoneMin      int     `json:"done_min"`
	DoneFraction float64 `json:"done_fraction"` // 0..1
}

// ProgressView is the İlerleme model: plan position, totals, per-group
// completion, the exam countdown and the 38-week strip.
type ProgressView struct {
	HasPlan         bool            `json:"has_plan"`
	WeekNumber      int             `json:"week_number"` // elapsed weeks (current week), 0 if outside
	WeeksTotal      int             `json:"weeks_total"` // 38
	SessionsDone    int             `json:"sessions_done"`
	SessionsPlanned int             `json:"sessions_planned"`
	DoneMin         int             `json:"done_min"`
	PlannedMin      int             `json:"planned_min"`
	DoneHours       string          `json:"done_hours"`    // "1,5"
	PlannedHours    string          `json:"planned_hours"` // "342"
	CurrentPhase    string          `json:"current_phase,omitempty"`
	DaysToExam      int             `json:"days_to_exam"` // 0 once the exam has passed
	AfterExam       bool            `json:"after_exam"`
	ExamDate        string          `json:"exam_date,omitempty"` // "18 Tem"
	Groups          []GroupProgress `json:"groups"`
	Strip           []WeekCell      `json:"strip"`
}

// Day returns the Bugün model for date. State is always derived against the
// clock's "today", so Day can be read for any date.
func (s *Service) Day(ctx context.Context, date clock.Date) (*DayView, error) {
	var view *DayView
	err := s.st.WithTx(ctx, func(tx *sql.Tx) error {
		var err error
		view, err = s.dayTx(ctx, date, tx)
		return err
	})
	if err != nil {
		return nil, err
	}
	return view, nil
}

// Read each form's session state and study version from the same snapshot.
func (s *Service) dayTx(ctx context.Context, date clock.Date, db *sql.Tx) (*DayView, error) {
	today := clock.Today(s.clk).String()

	tgName, tgKey, err := s.topicGroupNamesTx(ctx, db)
	if err != nil {
		return nil, err
	}
	movedBy, err := s.movedByMissedTx(ctx, db)
	if err != nil {
		return nil, err
	}

	daySessions, err := s.st.SessionsByDateRange(ctx, db, date.String(), date.String())
	if err != nil {
		return nil, err
	}
	view := &DayView{
		Date:    date.String(),
		Weekday: WeekdayTR(date.String()),
		Heading: FormatDayHeading(date.String()),
		IsRest:  len(daySessions) == 0,
	}
	studies, err := s.st.StudiesByDateRange(ctx, db, date.String(), addDays(date, 1).String())
	if err != nil {
		return nil, err
	}
	for _, ss := range daySessions {
		sv := buildSessionView(ss, today, movedBy, tgName, tgKey)
		attachStudy(&sv, studies[ss.ID], ss.Items)
		view.Sessions = append(view.Sessions, sv)
	}

	tomorrow := addDays(date, 1)
	tomSessions, err := s.st.SessionsByDateRange(ctx, db, tomorrow.String(), tomorrow.String())
	if err != nil {
		return nil, err
	}
	if len(tomSessions) > 0 {
		sv := buildSessionView(tomSessions[0], today, movedBy, tgName, tgKey)
		attachStudy(&sv, studies[tomSessions[0].ID], tomSessions[0].Items)
		view.Tomorrow = &sv
	}

	weekStart := clock.WeekStart(date)
	weekSessions, err := s.st.SessionsByDateRange(ctx, db, weekStart.String(), addDays(weekStart, 6).String())
	if err != nil {
		return nil, err
	}
	view.Notes = noteLines(weekSessions, movedBy)
	studyContext, err := s.readStudyContextTx(ctx, db)
	if err != nil {
		return nil, err
	}
	resolve := func(entry *StudyContextEntry) *SessionView {
		if entry == nil {
			return nil
		}
		sv := buildSessionView(entry.Session, today, movedBy, tgName, tgKey)
		attachStudy(&sv, &entry.Study, entry.Session.Items)
		return &sv
	}
	view.ActiveStudy = resolve(studyContext.Active)
	view.LatestPausedStudy = resolve(studyContext.LatestPaused)
	return view, nil
}

// Week returns the Hafta model for the Monday–Sunday week containing date.
func (s *Service) Week(ctx context.Context, date clock.Date) (*WeekView, error) {
	var view *WeekView
	err := s.st.WithTx(ctx, func(tx *sql.Tx) error {
		var err error
		view, err = s.weekTx(ctx, date, tx)
		return err
	})
	if err != nil {
		return nil, err
	}
	return view, nil
}

func (s *Service) weekTx(ctx context.Context, date clock.Date, db *sql.Tx) (*WeekView, error) {
	today := clock.Today(s.clk).String()

	tgName, tgKey, err := s.topicGroupNamesTx(ctx, db)
	if err != nil {
		return nil, err
	}
	movedBy, err := s.movedByMissedTx(ctx, db)
	if err != nil {
		return nil, err
	}

	start := clock.WeekStart(date)
	end := addDays(start, 6)
	sessions, err := s.st.SessionsByDateRange(ctx, db, start.String(), end.String())
	if err != nil {
		return nil, err
	}
	studies, err := s.st.StudiesByDateRange(ctx, db, start.String(), end.String())
	if err != nil {
		return nil, err
	}
	byDate := map[string][]store.Session{}
	for _, ss := range sessions {
		byDate[ss.Date] = append(byDate[ss.Date], ss)
	}

	view := &WeekView{
		Start:      start.String(),
		End:        end.String(),
		RangeLabel: FormatWeekRange(start.String(), end.String()),
	}
	for i := 0; i < 7; i++ {
		d := addDays(start, i)
		ds := d.String()
		row := DayRow{
			Date:         ds,
			Weekday:      WeekdayTR(ds),
			WeekdayShort: WeekdayShortTR(ds),
			DayNum:       d.Day,
			IsRest:       len(byDate[ds]) == 0,
		}
		for _, ss := range byDate[ds] {
			sv := buildSessionView(ss, today, movedBy, tgName, tgKey)
			attachStudy(&sv, studies[ss.ID], ss.Items)
			row.Sessions = append(row.Sessions, sv)
		}
		view.Days = append(view.Days, row)
	}

	for _, ss := range sessions {
		view.SessionsPlanned++
		view.PlannedMin += ss.DurationMin
		if ss.CompletedAt != nil {
			view.SessionsDone++
			view.DoneMin += ss.DurationMin
		}
		if _, moved := movedBy[ss.ID]; moved && ss.CompletedAt == nil {
			view.MovedCount++
		}
	}
	view.Notes = noteLines(sessions, movedBy)

	phases, err := s.st.ListPhases(ctx, db)
	if err != nil {
		return nil, err
	}
	view.WeekNumber = weekNumber(phases, start)
	if p := phaseOf(phases, start.String()); p != nil {
		view.PhaseName = p.Name
	}
	return view, nil
}

// Progress returns the İlerleme model. examDate is supplied by the caller (from
// WORKEY_EXAM_DATE); a zero date omits the countdown.
func (s *Service) Progress(ctx context.Context, examDate clock.Date) (*ProgressView, error) {
	today := clock.Today(s.clk)
	db := s.st.DB()

	phases, err := s.st.ListPhases(ctx, db)
	if err != nil {
		return nil, err
	}
	groups, err := s.st.ListTopicGroups(ctx, db)
	if err != nil {
		return nil, err
	}
	all, err := s.st.AllSessions(ctx, db)
	if err != nil {
		return nil, err
	}

	view := &ProgressView{
		WeeksTotal: WeeksTotal,
		HasPlan:    len(phases) > 0 || len(all) > 0,
	}

	// Totals over the whole saved plan.
	doneByGroup := map[int64]int{}
	planByGroup := map[int64]int{}
	for _, ss := range all {
		view.SessionsPlanned++
		view.PlannedMin += ss.DurationMin
		if ss.TopicGroupID != nil {
			planByGroup[*ss.TopicGroupID] += ss.DurationMin
		}
		if ss.CompletedAt != nil {
			view.SessionsDone++
			view.DoneMin += ss.DurationMin
			if ss.TopicGroupID != nil {
				doneByGroup[*ss.TopicGroupID] += ss.DurationMin
			}
		}
	}
	view.DoneHours = FormatHoursTR(view.DoneMin)
	view.PlannedHours = FormatHoursTR(view.PlannedMin)

	for _, g := range groups {
		gp := GroupProgress{
			Key:        g.Key,
			Name:       g.Name,
			DGSMin:     g.DGSQuestionMin,
			DGSMax:     g.DGSQuestionMax,
			PlannedMin: planByGroup[g.ID],
			DoneMin:    doneByGroup[g.ID],
		}
		gp.DoneFraction = fraction(gp.DoneMin, gp.PlannedMin)
		gp.Percent = int(math.Round(gp.DoneFraction * 100))
		view.Groups = append(view.Groups, gp)
	}

	// Exam countdown.
	if examDate != (clock.Date{}) {
		view.ExamDate = FormatShortDate(examDate.String())
		d := civilDays(today, examDate)
		if d <= 0 {
			view.AfterExam = true
		} else {
			view.DaysToExam = d
		}
	}

	// Plan position and 38-week strip need a plan start (the first phase's
	// Monday, or the earliest session's Monday when phases are absent).
	planStart, ok := planStartMonday(phases, all)
	if !ok {
		return view, nil
	}
	view.WeekNumber = weekNumber(phases, clock.WeekStart(today))
	if p := phaseOf(phases, today.String()); p != nil {
		view.CurrentPhase = p.Name
	}

	donePerWeek := make([]int, WeeksTotal)
	planPerWeek := make([]int, WeeksTotal)
	for _, ss := range all {
		d, ok := ParseDate(ss.Date)
		if !ok {
			continue
		}
		idx := civilDays(planStart, clock.WeekStart(d)) / 7
		if idx < 0 || idx >= WeeksTotal {
			continue
		}
		planPerWeek[idx] += ss.DurationMin
		if ss.CompletedAt != nil {
			donePerWeek[idx] += ss.DurationMin
		}
	}
	for i := 0; i < WeeksTotal; i++ {
		monday := addDays(planStart, i*7)
		cell := WeekCell{
			WeekNumber: i + 1,
			PlannedMin: planPerWeek[i],
			DoneMin:    donePerWeek[i],
		}
		cell.DoneFraction = fraction(cell.DoneMin, cell.PlannedMin)
		if p := phaseOf(phases, monday.String()); p != nil {
			cell.PhaseKey = p.Key
			cell.PhaseName = p.Name
		}
		view.Strip = append(view.Strip, cell)
	}
	return view, nil
}

// Range returns the sessions in [from, to] (inclusive) as SessionViews with
// derived state, resolved topic-group names and item notation, ordered by date
// then start time. It backs GET /api/v1/plan.
func (s *Service) Range(ctx context.Context, from, to clock.Date) ([]SessionView, error) {
	var views []SessionView
	err := s.st.WithTx(ctx, func(tx *sql.Tx) error {
		today := clock.Today(s.clk).String()
		tgName, tgKey, err := s.topicGroupNamesTx(ctx, tx)
		if err != nil {
			return err
		}
		movedBy, err := s.movedByMissedTx(ctx, tx)
		if err != nil {
			return err
		}
		sessions, err := s.st.SessionsByDateRange(ctx, tx, from.String(), to.String())
		if err != nil {
			return err
		}
		studies, err := s.st.StudiesByDateRange(ctx, tx, from.String(), to.String())
		if err != nil {
			return err
		}
		views = make([]SessionView, 0, len(sessions))
		for _, ss := range sessions {
			sv := buildSessionView(ss, today, movedBy, tgName, tgKey)
			attachStudy(&sv, studies[ss.ID], ss.Items)
			views = append(views, sv)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return views, nil
}

// Today returns the Bugün model for the clock's current date.
func (s *Service) Today(ctx context.Context) (*DayView, error) {
	return s.Day(ctx, clock.Today(s.clk))
}

// TodayDate returns the clock's current civil date (Istanbul). It is a read
// accessor so callers can default a missing ?date= to today.
func (s *Service) TodayDate() clock.Date {
	return clock.Today(s.clk)
}

// PhaseInfo and TopicGroupInfo mirror the stored outline for read responses.
type PhaseInfo struct {
	Key       string `json:"key"`
	Ordinal   int    `json:"ordinal"`
	Name      string `json:"name"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// TopicGroupInfo is one topic group in an OutlineView.
type TopicGroupInfo struct {
	Key            string `json:"key"`
	Name           string `json:"name"`
	Ordinal        int    `json:"ordinal"`
	DGSQuestionMin *int   `json:"dgs_question_min,omitempty"`
	DGSQuestionMax *int   `json:"dgs_question_max,omitempty"`
}

// OutlineView is the saved phases and topic groups, for GET /api/v1/outline.
type OutlineView struct {
	Phases      []PhaseInfo      `json:"phases"`
	TopicGroups []TopicGroupInfo `json:"topic_groups"`
}

// Outline returns the current phases and topic groups ordered by ordinal.
func (s *Service) Outline(ctx context.Context) (*OutlineView, error) {
	db := s.st.DB()
	phases, err := s.st.ListPhases(ctx, db)
	if err != nil {
		return nil, err
	}
	groups, err := s.st.ListTopicGroups(ctx, db)
	if err != nil {
		return nil, err
	}
	out := &OutlineView{}
	for _, p := range phases {
		out.Phases = append(out.Phases, PhaseInfo{
			Key: p.Key, Ordinal: p.Ordinal, Name: p.Name,
			StartDate: p.StartDate, EndDate: p.EndDate,
		})
	}
	for _, g := range groups {
		out.TopicGroups = append(out.TopicGroups, TopicGroupInfo{
			Key: g.Key, Name: g.Name, Ordinal: g.Ordinal,
			DGSQuestionMin: g.DGSQuestionMin, DGSQuestionMax: g.DGSQuestionMax,
		})
	}
	return out, nil
}

// --- shared read helpers ---

// buildSessionView prepares one stored session for a screen.
func buildSessionView(ss store.Session, today string, movedBy map[int64]store.MakeupLink, tgName, tgKey map[int64]string) SessionView {
	sv := SessionView{
		ID:          ss.ID,
		Date:        ss.Date,
		StartTime:   ss.StartTime,
		DurationMin: ss.DurationMin,
		Kind:        ss.Kind,
		Completed:   ss.CompletedAt != nil,
		Items:       toDomainItems(ss.Items),
	}
	sv.Notation = FormatItems(sv.Items)
	if ss.TopicGroupID != nil {
		sv.TopicGroupName = tgName[*ss.TopicGroupID]
		sv.TopicGroupKey = tgKey[*ss.TopicGroupID]
	}
	link, moved := movedBy[ss.ID]
	sv.State = sessionState(ss, today, moved)
	if moved {
		sv.MovedTo = &MovedInfo{Date: link.Date, Weekday: WeekdayTR(link.Date)}
	}
	return sv
}

// sessionState derives a session's screen state against today.
func sessionState(ss store.Session, today string, moved bool) SessionState {
	switch {
	case ss.CompletedAt != nil:
		return StateDone
	case ss.Date < today:
		if moved {
			return StateMoved
		}
		return StateMissed
	case ss.Date == today:
		return StateToday
	default:
		return StatePlanned
	}
}

// noteLines generates the plain, blameless note lines for the moved (missed but
// rescheduled) sessions in the given set, e.g.
// "Salı yapılamadı; M1·2 cumartesi bloğuna eklendi."
func noteLines(sessions []store.Session, movedBy map[int64]store.MakeupLink) []string {
	var notes []string
	for _, ss := range sessions {
		if ss.CompletedAt != nil {
			continue
		}
		link, moved := movedBy[ss.ID]
		if !moved || len(ss.Items) == 0 {
			continue
		}
		ref := Ref(toDomainItems(ss.Items)[0])
		notes = append(notes, WeekdayTR(ss.Date)+" yapılamadı; "+ref+" "+weekdayLowerTR(link.Date)+" bloğuna eklendi.")
	}
	return notes
}

// topicGroupNames returns id→name and id→key maps for every topic group.
func (s *Service) topicGroupNames(ctx context.Context, q *sql.DB) (name, key map[int64]string, err error) {
	groups, err := s.st.ListTopicGroups(ctx, q)
	if err != nil {
		return nil, nil, err
	}
	name = map[int64]string{}
	key = map[int64]string{}
	for _, g := range groups {
		name[g.ID] = g.Name
		key[g.ID] = g.Key
	}
	return name, key, nil
}

// movedByMissed maps a missed session's id to the makeup link that references
// it, so read models can mark it "moved".
func (s *Service) movedByMissed(ctx context.Context, q *sql.DB) (map[int64]store.MakeupLink, error) {
	links, err := s.st.ListMakeupLinks(ctx, q)
	if err != nil {
		return nil, err
	}
	byMissed := make(map[int64]store.MakeupLink, len(links))
	for _, l := range links {
		byMissed[l.MakeupOf] = l
	}
	return byMissed, nil
}

// toDomainItems converts stored items to domain items for notation and views.
func toDomainItems(items []store.SessionItem) []Item {
	out := make([]Item, 0, len(items))
	for _, it := range items {
		out = append(out, Item{
			Book:          it.Book,
			ChapterNo:     it.ChapterNo,
			ChapterTitle:  it.ChapterTitle,
			PageFrom:      it.PageFrom,
			PageTo:        it.PageTo,
			KonuFrom:      it.KonuFrom,
			KonuTo:        it.KonuTo,
			SoruFrom:      it.SoruFrom,
			SoruTo:        it.SoruTo,
			AntrenmanFrom: it.AntrenmanFrom,
			AntrenmanTo:   it.AntrenmanTo,
		})
	}
	return out
}

// fraction returns done/planned bounded to [0, 1], and 0 when planned is 0.
func fraction(done, planned int) float64 {
	if planned <= 0 {
		return 0
	}
	f := float64(done) / float64(planned)
	if f > 1 {
		return 1
	}
	return f
}

// phaseOf returns the phase whose [start, end] span contains date, or nil.
func phaseOf(phases []store.Phase, date string) *store.Phase {
	for i := range phases {
		if phases[i].StartDate <= date && date <= phases[i].EndDate {
			return &phases[i]
		}
	}
	return nil
}

// weekNumber returns the 1-based plan week of the Monday weekStart, or 0 when
// there is no plan or the week falls outside [1, WeeksTotal].
func weekNumber(phases []store.Phase, weekStart clock.Date) int {
	start, ok := planStartFromPhases(phases)
	if !ok {
		return 0
	}
	n := civilDays(start, weekStart)/7 + 1
	if n < 1 || n > WeeksTotal {
		return 0
	}
	return n
}

// planStartFromPhases returns the Monday of the earliest phase's start date.
func planStartFromPhases(phases []store.Phase) (clock.Date, bool) {
	earliest := ""
	for _, p := range phases {
		if earliest == "" || p.StartDate < earliest {
			earliest = p.StartDate
		}
	}
	if earliest == "" {
		return clock.Date{}, false
	}
	d, ok := ParseDate(earliest)
	if !ok {
		return clock.Date{}, false
	}
	return clock.WeekStart(d), true
}

// planStartMonday returns the plan's anchoring Monday: the first phase's Monday,
// or, when phases are absent, the earliest session's Monday.
func planStartMonday(phases []store.Phase, sessions []store.Session) (clock.Date, bool) {
	if start, ok := planStartFromPhases(phases); ok {
		return start, true
	}
	earliest := ""
	for _, ss := range sessions {
		if earliest == "" || ss.Date < earliest {
			earliest = ss.Date
		}
	}
	if earliest == "" {
		return clock.Date{}, false
	}
	d, ok := ParseDate(earliest)
	if !ok {
		return clock.Date{}, false
	}
	return clock.WeekStart(d), true
}

// --- civil-date arithmetic (zone-free; Turkey has no DST since 2016) ---

// ParseDate parses a 'YYYY-MM-DD' string into a civil Date.
func ParseDate(s string) (clock.Date, bool) {
	t, ok := parseCivil(s)
	if !ok {
		return clock.Date{}, false
	}
	return clock.Date{Year: t.Year(), Month: t.Month(), Day: t.Day()}, true
}

// addDays returns d shifted by n days.
func addDays(d clock.Date, n int) clock.Date {
	t := time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.UTC).AddDate(0, 0, n)
	return clock.Date{Year: t.Year(), Month: t.Month(), Day: t.Day()}
}

// civilDays returns the whole-day difference b-a of two civil dates.
func civilDays(a, b clock.Date) int {
	ta := time.Date(a.Year, a.Month, a.Day, 0, 0, 0, 0, time.UTC)
	tb := time.Date(b.Year, b.Month, b.Day, 0, 0, 0, 0, time.UTC)
	return int(math.Round(tb.Sub(ta).Hours() / 24))
}
