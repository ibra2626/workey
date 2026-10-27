package views

import (
	"math"
	"strconv"
	"strings"

	"github.com/a-h/templ"

	"github.com/ibra2626/workey/internal/plan"
)

// enDash is U+2013, the range separator used throughout the notation (see
// DESIGN.md Exact Notation Rule). It is defined here so the view layer can build
// page spans without reaching into the plan package's unexported formatters.
const enDash = "–"

// --- page containers -----------------------------------------------------
//
// The handlers assemble these from the plan read models and hand them to the
// templ screens. No business rule lives here: the service has already derived
// every figure, state and notation string.

// TodayPage is the Bugün screen model: the day itself plus the surrounding
// figures the header and stat tiles need (the service computes them on the Week
// and Progress read models).
type TodayPage struct {
	Day            *plan.DayView
	Week           *plan.WeekView // this week, for the desktop "Bu hafta" mini chart
	HasPlan        bool
	WeekNumber     int    // plan week 1..38, 0 outside the plan
	PhaseName      string // full phase name, e.g. "Faz 1 · İşlem temeli"
	DaysToExam     int
	AfterExam      bool
	HasExam        bool
	ExamDate       string // "18 Temmuz 2027" for the desktop rail
	WeekDoneMin    int
	WeekPlannedMin int
	SessionsDone   int
	SessionsPlan   int
	PhaseLabel     string       // "Faz 1 · İşlem temeli · Antrenmanlarla Matematik 1" style line
	PhaseBlocks    []PhaseBlock // current phase week blocks (desktop)
	CSRF           string
}

// PhaseBlock is one week block of the current phase panel on desktop Bugün.
type PhaseBlock struct {
	Label     string // "H1 · bu hafta" / "H2"
	Sub       string // "90 / 540 dk" for the current week, else ""
	FillPct   int    // ink fill width (done fraction) for the current week
	IsCurrent bool
}

// WeekPage is the Hafta screen model.
type WeekPage struct {
	Week       *plan.WeekView
	DaysToExam int
	HasExam    bool
	PrevDate   string // ?tarih= for the previous week
	NextDate   string // ?tarih= for the next week
	CSRF       string
}

// ProgressPage is the İlerleme screen model.
type ProgressPage struct {
	Progress   *plan.ProgressView
	DaysToExam int
	HasExam    bool
	ExamFull   string // "18 Temmuz 2027" for the rail
	CSRF       string
}

// --- session interval profile -------------------------------------------

// zoneBlock is one block of the session profile (DESIGN.md "Session profile").
type zoneBlock struct {
	Label     string // "Konu", "Soru", "Yanlış defteri", or the single-block label
	Pages     string // "s. 41–44" / "s. 45–58, Antr. 1–3"; "" when unknown
	Duration  string // "20 dk"
	Fill      string // CSS colour token, e.g. "var(--color-z1)"
	Flex      int    // flex weight (minutes), for the proportional width
	HeightPct int    // bar height as a percentage of the chart
	Single    bool   // true for the uzun_blok / deneme single block
}

// sessionProfile returns the interval blocks for a session. An akşam session
// splits into konu / soru / yanlış defteri at 20/50/20 of its minutes (the
// mockup's 90 → 20/50/20) with the zone heights from DESIGN.md. uzun_blok and
// deneme have no konu/soru breakdown, so they get one calm Phase-Mist block
// spanning the width (the single-block profile; see DECISIONS 2026-10-26
// WKY-09).
func sessionProfile(sv plan.SessionView) []zoneBlock {
	if sv.Kind == plan.KindAksam {
		konu := sv.DurationMin * 2 / 9
		soru := sv.DurationMin * 5 / 9
		yanlis := sv.DurationMin - konu - soru
		return []zoneBlock{
			{Label: "Konu", Pages: konuPages(sv), Duration: dk(konu), Fill: "var(--color-z1)", Flex: konu, HeightPct: 43},
			{Label: "Soru", Pages: soruPages(sv), Duration: dk(soru), Fill: "var(--color-z3)", Flex: soru, HeightPct: 92},
			{Label: "Yanlış defteri", Duration: dk(yanlis), Fill: "var(--color-z2)", Flex: yanlis, HeightPct: 64},
		}
	}
	label := "Uzun blok"
	if sv.Kind == plan.KindDeneme {
		label = "Deneme"
	}
	return []zoneBlock{
		{Label: label, Duration: dk(sv.DurationMin), Fill: "var(--color-z1-soft)", Flex: 1, HeightPct: 80, Single: true},
	}
}

// konuPages renders the konu page span of a session from its items' konu
// splits, e.g. "s. 41–44"; "" when no item carries a konu split.
func konuPages(sv plan.SessionView) string {
	if from, to, ok := spanOf(sv.Items, func(it plan.Item) (*int, *int) { return it.KonuFrom, it.KonuTo }); ok {
		return "s. " + numSpan(from, to)
	}
	return ""
}

// soruPages renders the soru page span and antrenman range, e.g.
// "s. 45–58, Antr. 1–3"; "" when nothing is known.
func soruPages(sv plan.SessionView) string {
	var parts []string
	if from, to, ok := spanOf(sv.Items, func(it plan.Item) (*int, *int) { return it.SoruFrom, it.SoruTo }); ok {
		parts = append(parts, "s. "+numSpan(from, to))
	}
	if from, to, ok := spanOf(sv.Items, func(it plan.Item) (*int, *int) { return it.AntrenmanFrom, it.AntrenmanTo }); ok {
		parts = append(parts, "Antr. "+numSpan(from, to))
	}
	return strings.Join(parts, ", ")
}

// spanOf returns the min-from and max-to over the items whose selected pair of
// optional bounds is present, and false when none is.
func spanOf(items []plan.Item, pick func(plan.Item) (*int, *int)) (from, to int, ok bool) {
	for _, it := range items {
		f, t := pick(it)
		if f == nil || t == nil {
			continue
		}
		if !ok {
			from, to, ok = *f, *t, true
			continue
		}
		if *f < from {
			from = *f
		}
		if *t > to {
			to = *t
		}
	}
	return from, to, ok
}

// numSpan renders an inclusive numeric span with an en dash, or a single number
// when the ends coincide.
func numSpan(from, to int) string {
	if from == to {
		return strconv.Itoa(from)
	}
	return strconv.Itoa(from) + enDash + strconv.Itoa(to)
}

// dk renders a minute count with its unit, e.g. "20 dk".
func dk(min int) string {
	return strconv.Itoa(min) + " dk"
}

// firstSession returns the day's first session (its evening session), or the
// zero value when the day has none. Callers guard on len first.
func firstSession(d *plan.DayView) plan.SessionView {
	if d == nil || len(d.Sessions) == 0 {
		return plan.SessionView{}
	}
	return d.Sessions[0]
}

// chapterRef renders a session's leading book·chapter reference, e.g. "M1·3",
// from its first item.
func chapterRef(sv plan.SessionView) string {
	if len(sv.Items) == 0 {
		return ""
	}
	return plan.Ref(sv.Items[0])
}

// chapterTitle returns the first item's chapter title.
func chapterTitle(sv plan.SessionView) string {
	if len(sv.Items) == 0 {
		return ""
	}
	return sv.Items[0].ChapterTitle
}

// chapterBook returns the full book name behind the first item's book code.
func chapterBook(sv plan.SessionView) string {
	if len(sv.Items) == 0 {
		return ""
	}
	return bookName(sv.Items[0].Book)
}

// headPages renders the session's overall page span, e.g. "s. 41–58", across
// all its items.
func headPages(sv plan.SessionView) string {
	if from, to, ok := spanOf(sv.Items, func(it plan.Item) (*int, *int) { return &it.PageFrom, &it.PageTo }); ok {
		return "s. " + numSpan(from, to)
	}
	return ""
}

// tomorrowLine renders the "Yarın" sub-line: "s. 59–76 · 20:30 · 90 dk".
func tomorrowLine(sv plan.SessionView) string {
	parts := []string{}
	if p := headPages(sv); p != "" {
		parts = append(parts, p)
	}
	parts = append(parts, sv.StartTime, dk(sv.DurationMin))
	return strings.Join(parts, " "+middleDot+" ")
}

// middleDot is U+00B7, the separator joining a session's meta fields.
const middleDot = "·"

// bookName maps a book code to its full title for the book line under a chapter
// (e.g. "M1" → "Antrenmanlarla Matematik 1"). Unknown codes render as the code.
func bookName(code string) string {
	switch code {
	case "M1":
		return "Antrenmanlarla Matematik 1"
	case "M2":
		return "Antrenmanlarla Matematik 2"
	case "M3":
		return "Antrenmanlarla Matematik 3"
	case "G1":
		return "Antrenmanlarla Geometri 1"
	case "G2":
		return "Antrenmanlarla Geometri 2"
	default:
		return code
	}
}

// axisTicks returns the time-axis labels under a session profile: every 30
// minutes from the start to the end for an akşam session, or just the start and
// end for a single-block (uzun_blok / deneme) session.
func axisTicks(start string, durationMin int, single bool) []string {
	base, ok := parseHHMM(start)
	if !ok {
		return nil
	}
	if single {
		return []string{start, formatHHMM(base + durationMin)}
	}
	var ticks []string
	for m := 0; m <= durationMin; m += 30 {
		ticks = append(ticks, formatHHMM(base+m))
	}
	return ticks
}

// parseHHMM parses "HH:MM" into minutes past midnight.
func parseHHMM(s string) (int, bool) {
	if len(s) != 5 || s[2] != ':' {
		return 0, false
	}
	h, err := strconv.Atoi(s[:2])
	if err != nil {
		return 0, false
	}
	m, err := strconv.Atoi(s[3:])
	if err != nil {
		return 0, false
	}
	return h*60 + m, true
}

// formatHHMM renders minutes past midnight as "HH:MM" (wrapping past 24:00).
func formatHHMM(min int) string {
	min %= 24 * 60
	if min < 0 {
		min += 24 * 60
	}
	h, m := min/60, min%60
	return pad2(h) + ":" + pad2(m)
}

func pad2(n int) string {
	if n < 10 {
		return "0" + strconv.Itoa(n)
	}
	return strconv.Itoa(n)
}

// --- volume bars (Hafta) -------------------------------------------------

// volumeBar is one day's bar in the weekly volume chart.
type volumeBar struct {
	State        plan.SessionState // today | done | moved | missed | planned
	Rest         bool
	HeightPct    int    // 0 for rest
	WeekdayShort string // "Pzt"
	DayNum       int
	Minutes      int
	IsToday      bool
}

// weekBars builds the seven daily bars, scaling heights to the busiest day of
// the week so the tallest bar reads near full height.
func weekBars(wk *plan.WeekView) []volumeBar {
	maxMin := 0
	mins := make([]int, len(wk.Days))
	for i, d := range wk.Days {
		for _, s := range d.Sessions {
			mins[i] += s.DurationMin
		}
		if mins[i] > maxMin {
			maxMin = mins[i]
		}
	}
	bars := make([]volumeBar, 0, len(wk.Days))
	for i, d := range wk.Days {
		b := volumeBar{
			WeekdayShort: d.WeekdayShort,
			DayNum:       d.DayNum,
			Minutes:      mins[i],
			Rest:         d.IsRest,
			State:        dayState(d),
		}
		b.IsToday = b.State == plan.StateToday
		if !d.IsRest && maxMin > 0 {
			b.HeightPct = int(math.Round(float64(mins[i]) / float64(maxMin) * 100))
			if b.HeightPct < 10 {
				b.HeightPct = 10
			}
		}
		bars = append(bars, b)
	}
	return bars
}

// dayState reduces a day's sessions to the one state its bar should show:
// today wins, then done (all sessions done), then moved/missed (past, undone),
// else planned.
func dayState(d plan.DayRow) plan.SessionState {
	if len(d.Sessions) == 0 {
		return plan.StatePlanned
	}
	allDone := true
	anyMoved := false
	for _, s := range d.Sessions {
		switch s.State {
		case plan.StateToday:
			return plan.StateToday
		case plan.StateMoved, plan.StateMissed:
			anyMoved = true
		}
		if s.State != plan.StateDone {
			allDone = false
		}
	}
	switch {
	case allDone:
		return plan.StateDone
	case anyMoved:
		return plan.StateMoved
	default:
		return plan.StatePlanned
	}
}

// barClass returns the fill utility classes for a volume bar of the given state
// (the ochre today ring and the hatch pattern are inline styles in the markup).
func barClass(b volumeBar) string {
	switch {
	case b.IsToday:
		return "bg-z3-soft"
	case b.State == plan.StateDone:
		return "bg-ink"
	case b.State == plan.StateMoved || b.State == plan.StateMissed:
		return "hatch"
	default:
		return "bg-line"
	}
}

// --- week list / table status words ------------------------------------

// statusWord renders a session's status for the Hafta table and list, e.g.
// "Tamamlandı", "Cumartesiye taşındı", "Bugün 20:30", "Planlı".
func statusWord(sv plan.SessionView) string {
	switch sv.State {
	case plan.StateDone:
		return "Tamamlandı"
	case plan.StateToday:
		return "Bugün " + sv.StartTime
	case plan.StateMoved:
		if sv.MovedTo != nil {
			return capitalFirst(weekdayDative(sv.MovedTo.Weekday)) + " taşındı"
		}
		return "Taşındı"
	case plan.StateMissed:
		return "Yapılamadı"
	default:
		if sv.Kind == plan.KindUzunBlok || sv.Kind == plan.KindDeneme {
			return "Planlı · " + sv.StartTime
		}
		return "Planlı"
	}
}

// rowName renders a week-row session name: "M1·3 Üslü ifadeler" for a chapter
// session, or "Uzun blok" / "Deneme" for the block kinds.
func rowName(sv plan.SessionView) string {
	switch sv.Kind {
	case plan.KindUzunBlok:
		return "Uzun blok"
	case plan.KindDeneme:
		return "Deneme"
	default:
		return chapterRef(sv) + " " + chapterTitle(sv)
	}
}

// rowMeta renders the mobile week-row meta line: the page span and a status
// detail, e.g. "s. 41–58 · bugün 20:30", "s. 23–40 · cumartesiye taşındı".
func rowMeta(sv plan.SessionView) string {
	pages := headPages(sv)
	detail := ""
	switch sv.State {
	case plan.StateToday:
		detail = "bugün " + sv.StartTime
	case plan.StateMoved:
		if sv.MovedTo != nil {
			detail = weekdayDative(sv.MovedTo.Weekday) + " taşındı"
		} else {
			detail = "taşındı"
		}
	case plan.StateMissed:
		detail = "yapılamadı"
	default:
		if sv.Kind == plan.KindUzunBlok || sv.Kind == plan.KindDeneme {
			detail = sv.StartTime + " · " + dk(sv.DurationMin)
		} else {
			detail = dk(sv.DurationMin)
		}
	}
	if pages == "" {
		return detail
	}
	return pages + " " + middleDot + " " + detail
}

// pagesPlain renders a session's page span without the "s." prefix, for the
// Hafta table "Sayfa" column, e.g. "41–58".
func pagesPlain(sv plan.SessionView) string {
	if from, to, ok := spanOf(sv.Items, func(it plan.Item) (*int, *int) { return &it.PageFrom, &it.PageTo }); ok {
		return numSpan(from, to)
	}
	return ""
}

// sureCell renders the Hafta table "Süre" column: the duration, or an em dash
// for a moved session (its minutes moved with it).
func sureCell(sv plan.SessionView) string {
	if sv.State == plan.StateMoved {
		return "—"
	}
	return dk(sv.DurationMin)
}

// rowActionable reports whether a row offers a mark-done / undo control: today,
// past-done and past-missed rows do; future, planned-future, moved and rest
// rows do not (DECISIONS: only today and past can be marked).
func rowActionable(sv plan.SessionView) bool {
	switch sv.State {
	case plan.StateDone, plan.StateToday, plan.StateMissed:
		return true
	default:
		return false
	}
}

// rowActive reports whether a row's name reads as active (bold): done and today
// rows, matching the mockup's week list and table.
func rowActive(sv plan.SessionView) bool {
	return sv.State == plan.StateDone || sv.State == plan.StateToday
}

// rowActionSlug returns the completion route slug for an actionable row.
func rowActionSlug(sv plan.SessionView) string {
	if sv.State == plan.StateDone {
		return "geri-al"
	}
	return "tamamla"
}

// rowActionLabel returns the button label for an actionable row.
func rowActionLabel(sv plan.SessionView) string {
	if sv.State == plan.StateDone {
		return "Geri al"
	}
	return "Bitir"
}

// weekForm returns the form attributes for a Hafta row's completion control.
// The whole main content (#hafta-main) is swapped so the summary figures and
// the row update together.
func weekForm(id int64, action string) templ.Attributes {
	url := "/hafta/oturum/" + strconv.FormatInt(id, 10) + "/" + action
	return templ.Attributes{
		"action":    url,
		"hx-post":   url,
		"hx-target": "#hafta-main",
		"hx-swap":   "outerHTML",
	}
}

// phaseTail returns the descriptive part of a phase name after "Faz N · ", e.g.
// "Faz 1 · İşlem temeli" → "İşlem temeli"; the whole string when it has no "·".
func phaseTail(name string) string {
	if i := strings.Index(name, " · "); i >= 0 {
		return name[i+len(" · "):]
	}
	return name
}

// weekdayDative returns the Turkish dative ("-e/-a") form of a capitalised
// weekday, e.g. "Cumartesi" → "cumartesiye". The tables use it as "… taşındı".
func weekdayDative(weekday string) string {
	switch weekday {
	case "Pazartesi":
		return "pazartesiye"
	case "Salı":
		return "salıya"
	case "Çarşamba":
		return "çarşambaya"
	case "Perşembe":
		return "perşembeye"
	case "Cuma":
		return "cumaya"
	case "Cumartesi":
		return "cumartesiye"
	case "Pazar":
		return "pazara"
	default:
		return strings.ToLower(weekday)
	}
}

// capitalFirst upper-cases the first rune only, Turkish-safe for the forms this
// package produces (none begins with i/ı).
func capitalFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	return strings.ToUpper(string(r[0])) + string(r[1:])
}

// --- progress strip ------------------------------------------------------

// stripCell is one column of the 38-week plan strip: a base fill (Phase Mist in
// the current phase, Hairline later) with a Forest-Ink overlay for the done
// fraction.
type stripCell struct {
	BaseFill string
	InkPct   int
}

// dgsLabel renders a topic group's DGS question weight, e.g. "8–10 soru" (never
// abbreviated "s."); "" when the group carries no weight.
func dgsLabel(gp plan.GroupProgress) string {
	if gp.DGSMin == nil || gp.DGSMax == nil {
		return ""
	}
	return numSpan(*gp.DGSMin, *gp.DGSMax) + " soru"
}

// currentPhaseRange returns the first and last plan week of the current phase,
// from the strip, e.g. 1 and 7 for "Faz 1 · H1–7"; (0, 0) when unknown.
func currentPhaseRange(v *plan.ProgressView) (first, last int) {
	for _, c := range v.Strip {
		if c.PhaseName == "" || c.PhaseName != v.CurrentPhase {
			continue
		}
		if first == 0 || c.WeekNumber < first {
			first = c.WeekNumber
		}
		if c.WeekNumber > last {
			last = c.WeekNumber
		}
	}
	return first, last
}

// phaseRangeLabel renders "Faz 1 · H1–7" for the current phase, or just the
// phase short name when the week range is unknown.
func phaseRangeLabel(v *plan.ProgressView) string {
	first, last := currentPhaseRange(v)
	if first == 0 {
		return phaseShort(v.CurrentPhase)
	}
	return phaseShort(v.CurrentPhase) + " " + middleDot + " H" + strconv.Itoa(first) + enDash + "H" + strconv.Itoa(last)
}

// progressStrip builds the 38 columns: the current phase's weeks sit on Phase
// Mist, later phases on Hairline; each column shows its done fraction in ink.
func progressStrip(v *plan.ProgressView) []stripCell {
	out := make([]stripCell, 0, len(v.Strip))
	for _, c := range v.Strip {
		base := "var(--color-line)"
		if c.PhaseName != "" && c.PhaseName == v.CurrentPhase {
			base = "var(--color-z1-soft)"
		}
		out = append(out, stripCell{BaseFill: base, InkPct: int(math.Round(c.DoneFraction * 100))})
	}
	return out
}

// --- small text helpers --------------------------------------------------

// phaseShort returns the leading "Faz N" of a full phase name
// ("Faz 1 · İşlem temeli" → "Faz 1"); the whole string when it has no "·".
func phaseShort(name string) string {
	if i := strings.Index(name, " · "); i >= 0 {
		return name[:i]
	}
	return name
}

// examFigure renders the days-to-exam count for the Sınava tile, or "—" when
// no exam date is configured or it has passed.
func examFigure(p TodayPage) string {
	if !p.HasExam || p.DaysToExam <= 0 {
		return "—"
	}
	return strconv.Itoa(p.DaysToExam)
}

// tomorrowHint renders the rest-day hint about tomorrow's session.
func tomorrowHint(sv *plan.SessionView) string {
	if sv == nil {
		return "de dinlenme."
	}
	return chapterRef(*sv) + " " + chapterTitle(*sv) + " var."
}

// completeForm returns the form attributes (action + htmx) for the completion
// and undo controls. action is "tamamla" or "geri-al".
func completeForm(id int64, action string) templ.Attributes {
	url := "/oturum/" + strconv.FormatInt(id, 10) + "/" + action
	return templ.Attributes{
		"action":    url,
		"hx-post":   url,
		"hx-target": "#bugun-main",
		"hx-swap":   "outerHTML",
	}
}

// percentClass dims a zero percentage to Lichen Ink (DESIGN.md progress tracks).
func percentClass(percent int) string {
	if percent == 0 {
		return "text-ink-3"
	}
	return "text-ink"
}

// pct renders a width/height percentage string like "17%".
func pct(p int) string {
	return strconv.Itoa(p) + "%"
}
