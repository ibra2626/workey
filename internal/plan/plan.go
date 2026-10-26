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

// Session kinds.
const (
	KindAksam    = "aksam"
	KindUzunBlok = "uzun_blok"
	KindDeneme   = "deneme"
)

// validKinds is the set accepted for a session's kind.
var validKinds = map[string]bool{
	KindAksam:    true,
	KindUzunBlok: true,
	KindDeneme:   true,
}

// validBooks is the set accepted for an item's book.
var validBooks = map[string]bool{
	"M1": true, "M2": true, "M3": true, "G1": true, "G2": true,
}

// Item is one book reference inside a session: a book, a chapter and a page
// range, with optional konu / soru / antrenman splits. It stores references
// only — no topic text or question content.
type Item struct {
	Book          string `json:"book"`
	ChapterNo     int    `json:"chapter_no"`
	ChapterTitle  string `json:"chapter_title"`
	PageFrom      int    `json:"page_from"`
	PageTo        int    `json:"page_to"`
	KonuFrom      *int   `json:"konu_from,omitempty"`
	KonuTo        *int   `json:"konu_to,omitempty"`
	SoruFrom      *int   `json:"soru_from,omitempty"`
	SoruTo        *int   `json:"soru_to,omitempty"`
	AntrenmanFrom *int   `json:"antrenman_from,omitempty"`
	AntrenmanTo   *int   `json:"antrenman_to,omitempty"`
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
// targets a frozen slot (a past date or a completed session).
type SkippedSession struct {
	Date      string `json:"date"`
	StartTime string `json:"start_time"`
	Reason    string `json:"reason"` // machine code: past_date | completed
}

// Skip reason codes.
const (
	SkipPastDate  = "past_date"
	SkipCompleted = "completed"
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
	Key            string `json:"key"`
	Name           string `json:"name"`
	Ordinal        int    `json:"ordinal"`
	DGSQuestionMin *int   `json:"dgs_question_min,omitempty"`
	DGSQuestionMax *int   `json:"dgs_question_max,omitempty"`
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
