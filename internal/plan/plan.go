// Package plan is workey's single business-logic layer. It owns domain types,
// input validation, the book-reference notation formatter and the service that
// writes the plan into internal/store. Web, REST and MCP only parse input, call
// this service and render its output; no rule is implemented twice.
//
// Writes are validated, idempotent, frozen-past-aware and audited: sessions
// dated before today (Europe/Istanbul) and completed sessions are never changed
// or deleted, the same request applied twice leaves the same state and adds no
// audit rows, and every change records its source with before/after JSON.
package plan

// Source identifies who is writing: an MCP tool, the REST API or the web UI.
// It is recorded on every audit row.
type Source string

const (
	SourceMCP Source = "mcp"
	SourceAPI Source = "api"
	SourceWeb Source = "web"
)

// valid reports whether s is one of the known sources.
func (s Source) valid() bool {
	switch s {
	case SourceMCP, SourceAPI, SourceWeb:
		return true
	default:
		return false
	}
}

// Session kinds. These are display hints only: since WKY-22 a session's kind is
// a free label (the store imposes no CHECK), so a blank or unknown kind is
// accepted. The constants name the kinds the v1 screens still special-case.
const (
	KindAksam    = "aksam"
	KindUzunBlok = "uzun_blok"
	KindDeneme   = "deneme"
)

// Item is one catalog reference inside a session: a source (named by its stable
// key), an optional section (by ordinal within the source), and a unit range
// and/or a location range. unit_from/unit_to are 1-based unit ordinals within
// the section; loc_from/loc_to are raw numbers in the source's location unit
// (page / minute / lesson / other). The service resolves key → id and section
// ordinal → id at write time. An item stores references only — never topic text
// or question content. See internal/store.SessionItem for the persisted shape.
type Item struct {
	SourceKey      string `json:"source_key"`
	SectionOrdinal *int   `json:"section_ordinal,omitempty"`
	UnitFrom       *int   `json:"unit_from,omitempty"`
	UnitTo         *int   `json:"unit_to,omitempty"`
	LocFrom        *int   `json:"loc_from,omitempty"`
	LocTo          *int   `json:"loc_to,omitempty"`
}

// SlotRef addresses a session by its (date, start_time) key, the stable handle
// an external caller uses — for example to link a makeup session to the missed
// one.
type SlotRef struct {
	Date      string `json:"date"`       // 'YYYY-MM-DD'
	StartTime string `json:"start_time"` // 'HH:MM'
}

// SessionInput is one planned session in a SavePlan request.
type SessionInput struct {
	Date        string   `json:"date"`                  // 'YYYY-MM-DD'
	StartTime   string   `json:"start_time"`            // 'HH:MM'
	DurationMin int      `json:"duration_min"`          //
	Kind        string   `json:"kind"`                  // aksam | uzun_blok | deneme
	TopicGroup  string   `json:"topic_group,omitempty"` // topic-group key, resolved to an id
	MakeupOf    *SlotRef `json:"makeup_of,omitempty"`   // the missed session this makes up
	Items       []Item   `json:"items"`
}

// slot returns the session's (date, start_time) key.
func (si SessionInput) slot() SlotRef { return SlotRef{Date: si.Date, StartTime: si.StartTime} }

// SavePlanRequest declares the open sessions for the dates in [From, To]. See
// Service.SavePlan for the replace semantics.
type SavePlanRequest struct {
	From           string         `json:"from"` // 'YYYY-MM-DD', inclusive
	To             string         `json:"to"`   // 'YYYY-MM-DD', inclusive
	Sessions       []SessionInput `json:"sessions"`
	IdempotencyKey string         `json:"idempotency_key,omitempty"`
}

// SessionRef identifies an affected session in a SavePlan result.
type SessionRef struct {
	ID        int64  `json:"id"`
	Date      string `json:"date"`
	StartTime string `json:"start_time"`
}

// SkippedSession reports an incoming session that was not applied because it
// targets a frozen slot (past, completed, or changed after starting).
type SkippedSession struct {
	Date      string `json:"date"`
	StartTime string `json:"start_time"`
	Reason    string `json:"reason" jsonschema:"past_date or completed take priority; started means a changed started session was preserved. Omitted started sessions are retained; identical ones are unchanged."` // machine code: past_date | completed | started
}

// Skip reason codes.
const (
	SkipPastDate  = "past_date"
	SkipCompleted = "completed"
	SkipStarted   = "started"
)

// SavePlanResult reports what a SavePlan call changed, grouped by disposition.
type SavePlanResult struct {
	Created   []SessionRef     `json:"created"`
	Updated   []SessionRef     `json:"updated"`
	Deleted   []SessionRef     `json:"deleted"`
	Unchanged []SessionRef     `json:"unchanged"`
	Skipped   []SkippedSession `json:"skipped"`
}

// PhaseInput is one phase in a SaveOutline request.
type PhaseInput struct {
	Key       string `json:"key"`
	Ordinal   int    `json:"ordinal"`
	Name      string `json:"name"`
	StartDate string `json:"start_date"` // 'YYYY-MM-DD'
	EndDate   string `json:"end_date"`   // 'YYYY-MM-DD'
}

// TopicGroupInput is one topic group in a SaveOutline request.
type TopicGroupInput struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Ordinal     int    `json:"ordinal"`
	QuestionMin *int   `json:"question_min,omitempty"`
	QuestionMax *int   `json:"question_max,omitempty"`
}

// SaveOutlineRequest declares the full set of phases and topic groups; each is
// replaced by key.
type SaveOutlineRequest struct {
	Phases         []PhaseInput      `json:"phases"`
	TopicGroups    []TopicGroupInput `json:"topic_groups"`
	IdempotencyKey string            `json:"idempotency_key,omitempty"`
}

// OutlineChange lists the keys affected for one entity kind.
type OutlineChange struct {
	Created   []string `json:"created"`
	Updated   []string `json:"updated"`
	Deleted   []string `json:"deleted"`
	Unchanged []string `json:"unchanged"`
}

// SaveOutlineResult reports what a SaveOutline call changed.
type SaveOutlineResult struct {
	Phases      OutlineChange `json:"phases"`
	TopicGroups OutlineChange `json:"topic_groups"`
}

// --- catalog write-path inputs (the save_source path, WKY-24/25) ---------
//
// A source is written whole: its definition plus its full ordered set of
// sections, each carrying its ordered units. The service resolves the source by
// key and reconciles sections/units by id so a re-write keeps unit progress
// (see internal/store UpdateSection / UpdateUnit vs ReplaceSections). These
// types are defined here now so WKY-24 (MCP) and WKY-25 (REST) share one shape.

// SourceInput is a source's own fields in a SaveSource request.
type SourceInput struct {
	Key          string `json:"key"`
	Name         string `json:"name"`
	Kind         string `json:"kind"`          // book | video | course | other (free label)
	LocationUnit string `json:"location_unit"` // page | minute | lesson | other
	Ordinal      int    `json:"ordinal"`
}

// SectionInput is one section of a source, with its ordered units.
type SectionInput struct {
	Ordinal int         `json:"ordinal"`
	Name    string      `json:"name"`
	LocFrom *int        `json:"loc_from,omitempty"`
	LocTo   *int        `json:"loc_to,omitempty"`
	Units   []UnitInput `json:"units"`
}

// UnitInput is one study unit inside a section. Label is the source's own word
// ("Antrenman" / "Test" / "Föy" / "Ders"); the code never enumerates them.
type UnitInput struct {
	Ordinal int    `json:"ordinal"`
	Label   string `json:"label"`
	Number  *int   `json:"number,omitempty"`
	LocFrom *int   `json:"loc_from,omitempty"`
	LocTo   *int   `json:"loc_to,omitempty"`
}

// SaveSourceRequest declares one source and its full set of sections/units,
// replaced by the source's key.
type SaveSourceRequest struct {
	Source         SourceInput    `json:"source"`
	Sections       []SectionInput `json:"sections"`
	IdempotencyKey string         `json:"idempotency_key,omitempty"`
}
