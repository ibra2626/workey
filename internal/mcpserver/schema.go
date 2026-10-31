package mcpserver

import (
	"github.com/google/jsonschema-go/jsonschema"

	"github.com/ibra2626/workey/internal/plan"
)

// outputSchema keeps additive study fields discoverable from tools/list. The
// SDK infers the Go shape; enrich enum/bounds and preserve nil-slice wire values.
func outputSchema[T any]() (*jsonschema.Schema, error) {
	s, err := jsonschema.For[T](nil)
	if err != nil {
		return nil, err
	}
	var visit func(*jsonschema.Schema)
	visit = func(s *jsonschema.Schema) {
		if s == nil {
			return
		}
		if s.Type == "array" {
			s.Type = ""
			s.Types = []string{"array", "null"}
		}
		for name, p := range s.Properties {
			switch name {
			case "status":
				p.Enum = []any{plan.StudyActive, plan.StudyPaused, plan.StudyClosed}
			case "entry_mode":
				p.Enum = []any{plan.EntryNormal, plan.EntryTenMinutes}
			case "stage":
				p.Enum = []any{plan.StageKonu, plan.StageSoru, plan.StageYanlisDefteri, plan.StageCalisma}
			case "reason":
				p.Enum = []any{plan.SkipPastDate, plan.SkipCompleted, plan.SkipStarted}
			case "item_ordinal":
				p.Minimum = jsonschema.Ptr(0.0)
			case "version", "resume_page":
				p.Minimum = jsonschema.Ptr(1.0)
			case "study":
				p.ReadOnly = true
			case "first_started_at", "last_started_at", "updated_at":
				p.Format = "date-time"
			}
			visit(p)
		}
		visit(s.Items)
		for _, d := range s.Defs {
			visit(d)
		}
		for _, d := range s.AnyOf {
			visit(d)
		}
	}
	visit(s)
	return s, nil
}

// Shared schema fragments. The patterns and enums mirror the plan service's
// validation so an AI client can build a valid call from tools/list alone; the
// service re-checks everything, so these are a convenience, not the guarantee.
const (
	datePattern = `^\d{4}-\d{2}-\d{2}$`       // 'YYYY-MM-DD'
	timePattern = `^([01]\d|2[0-3]):[0-5]\d$` // 'HH:MM', 24-hour
)

// bookEnum and kindEnum list the accepted item books and session kinds.
var (
	bookEnum = []any{"M1", "M2", "M3", "G1", "G2"}
	kindEnum = []any{plan.KindAksam, plan.KindUzunBlok, plan.KindDeneme}
)

// setProp sets a child property's attributes when it is present, so schema
// enrichment is resilient to the inferred shape. It returns the child so calls
// can be chained onto nested objects.
func setProp(parent *jsonschema.Schema, name, description string) *jsonschema.Schema {
	if parent == nil || parent.Properties == nil {
		return nil
	}
	child := parent.Properties[name]
	if child == nil {
		return nil
	}
	if description != "" {
		child.Description = description
	}
	return child
}

// itemsOf returns an array schema's element schema, or nil.
func itemsOf(s *jsonschema.Schema) *jsonschema.Schema {
	if s == nil {
		return nil
	}
	return s.Items
}

// savePlanSchema infers the schema for plan.SavePlanRequest and enriches it with
// patterns, enums, bounds and AI-facing descriptions. Returns nil on an
// inference error; the caller treats that as fatal at startup.
func savePlanSchema() (*jsonschema.Schema, error) {
	s, err := jsonschema.For[plan.SavePlanRequest](nil)
	if err != nil {
		return nil, err
	}

	from := setProp(s, "from", "Inclusive start date of the range to replace, 'YYYY-MM-DD' (Europe/Istanbul). Example: 2026-10-26.")
	if from != nil {
		from.Pattern = datePattern
	}
	to := setProp(s, "to", "Inclusive end date of the range to replace, 'YYYY-MM-DD' (Europe/Istanbul). Example: 2026-11-01.")
	if to != nil {
		to.Pattern = datePattern
	}
	setProp(s, "idempotency_key", "Optional caller-chosen key. Repeating a request with the same key replays the first result; reusing a key with a different body is rejected.")

	sessions := setProp(s, "sessions", "The complete set of open sessions for every date in [from, to]. Each date in range that is today or later is made to match exactly this set (by date + start_time): omitted unstarted open future sessions are deleted. Any study record protects the session: identical=unchanged, changed=skipped reason=started, omitted=retained. Past/completed reasons take priority. Example: 2026-10-26.")
	enrichSessionSchema(itemsOf(sessions))

	return s, nil
}

// enrichSessionSchema annotates one session object's schema (the element of the
// sessions array).
func enrichSessionSchema(sess *jsonschema.Schema) {
	if sess == nil {
		return
	}
	date := setProp(sess, "date", "Session date, 'YYYY-MM-DD' (Europe/Istanbul). Past days are frozen: a session dated before today is never changed and comes back in 'skipped'.")
	if date != nil {
		date.Pattern = datePattern
	}
	start := setProp(sess, "start_time", "Local start time, 'HH:MM' 24-hour (Europe/Istanbul). Together with date this is the stable handle for a session.")
	if start != nil {
		start.Pattern = timePattern
	}
	dur := setProp(sess, "duration_min", "Planned length in minutes.")
	if dur != nil {
		dur.Minimum = jsonschema.Ptr(1.0)
		dur.Maximum = jsonschema.Ptr(1440.0)
	}
	kind := setProp(sess, "kind", "Session kind: 'aksam' (evening block), 'uzun_blok' (long block) or 'deneme' (mock exam, no results are stored).")
	if kind != nil {
		kind.Enum = kindEnum
	}
	setProp(sess, "topic_group", "Optional topic-group key (from save_outline). Used only to attribute planned/done minutes in progress; it is not topic text.")

	makeup := setProp(sess, "makeup_of", "Optional link to the missed session this one makes up, addressed by its date + start_time. The target must be a past, uncompleted session; the missed day then reads as moved.")
	if makeup != nil {
		mdate := setProp(makeup, "date", "Date of the missed session, 'YYYY-MM-DD'.")
		if mdate != nil {
			mdate.Pattern = datePattern
		}
		mstart := setProp(makeup, "start_time", "Start time of the missed session, 'HH:MM'.")
		if mstart != nil {
			mstart.Pattern = timePattern
		}
	}

	items := setProp(sess, "items", "One or more book references for the session. Only book, chapter and page/range references are stored — never topic explanations or question content.")
	if items != nil {
		items.MinItems = jsonschema.Ptr(1)
	}
	enrichItemSchema(itemsOf(items))
}

// enrichItemSchema annotates one item object's schema.
func enrichItemSchema(it *jsonschema.Schema) {
	if it == nil {
		return
	}
	book := setProp(it, "book", "Book code: M1, M2, M3 (matematik) or G1, G2 (geometri).")
	if book != nil {
		book.Enum = bookEnum
	}
	ch := setProp(it, "chapter_no", "Chapter number within the book.")
	if ch != nil {
		ch.Minimum = jsonschema.Ptr(1.0)
	}
	setProp(it, "chapter_title", "Chapter title as printed in the book (a reference label, not topic content).")
	pf := setProp(it, "page_from", "First page of the range (inclusive).")
	if pf != nil {
		pf.Minimum = jsonschema.Ptr(1.0)
	}
	pt := setProp(it, "page_to", "Last page of the range (inclusive); must be >= page_from.")
	if pt != nil {
		pt.Minimum = jsonschema.Ptr(1.0)
	}
	boundOptional(it, "konu_from", "Optional first 'konu' (lesson) page of the range.")
	boundOptional(it, "konu_to", "Optional last 'konu' (lesson) page; must be >= konu_from.")
	boundOptional(it, "soru_from", "Optional first 'soru' (question) page of the range.")
	boundOptional(it, "soru_to", "Optional last 'soru' (question) page; must be >= soru_from.")
	boundOptional(it, "antrenman_from", "Optional first 'antrenman' (drill) number of the range.")
	boundOptional(it, "antrenman_to", "Optional last 'antrenman' (drill) number; must be >= antrenman_from.")
}

// boundOptional annotates an optional positive-integer field.
func boundOptional(parent *jsonschema.Schema, name, description string) {
	c := setProp(parent, name, description)
	if c != nil {
		c.Minimum = jsonschema.Ptr(1.0)
	}
}

// saveOutlineSchema infers and enriches the schema for plan.SaveOutlineRequest.
func saveOutlineSchema() (*jsonschema.Schema, error) {
	s, err := jsonschema.For[plan.SaveOutlineRequest](nil)
	if err != nil {
		return nil, err
	}
	setProp(s, "idempotency_key", "Optional caller-chosen key, with the same replay semantics as save_plan.")

	phases := setProp(s, "phases", "The complete set of plan phases. A phase present by key is created or updated; a key absent from the request is deleted.")
	if ph := itemsOf(phases); ph != nil {
		setProp(ph, "key", "Stable machine key for the phase (e.g. 'faz-1').")
		setProp(ph, "name", "Display name of the phase.")
		ord := setProp(ph, "ordinal", "Order of the phase in the plan, from 1.")
		if ord != nil {
			ord.Minimum = jsonschema.Ptr(0.0)
		}
		sd := setProp(ph, "start_date", "Phase start date, 'YYYY-MM-DD' (Europe/Istanbul).")
		if sd != nil {
			sd.Pattern = datePattern
		}
		ed := setProp(ph, "end_date", "Phase end date, 'YYYY-MM-DD'; must be >= start_date.")
		if ed != nil {
			ed.Pattern = datePattern
		}
	}

	groups := setProp(s, "topic_groups", "The complete set of topic groups. Matched by key (create/update/delete). A group still referenced by sessions cannot be deleted.")
	if g := itemsOf(groups); g != nil {
		setProp(g, "key", "Stable machine key for the topic group (referenced by a session's topic_group).")
		setProp(g, "name", "Display name of the topic group.")
		ord := setProp(g, "ordinal", "Order of the group, from 1.")
		if ord != nil {
			ord.Minimum = jsonschema.Ptr(0.0)
		}
		boundOptional(g, "dgs_question_min", "Optional lower bound of the group's expected DGS question count.")
		boundOptional(g, "dgs_question_max", "Optional upper bound of the group's expected DGS question count.")
	}

	return s, nil
}

// getPlanInput is the argument schema for get_plan.
type getPlanInput struct {
	From string `json:"from" jsonschema:"inclusive start date, 'YYYY-MM-DD' (Europe/Istanbul), e.g. 2026-10-26"`
	To   string `json:"to" jsonschema:"inclusive end date, 'YYYY-MM-DD' (Europe/Istanbul), e.g. 2026-11-01"`
}

// getPlanSchema infers and enriches the schema for getPlanInput.
func getPlanSchema() (*jsonschema.Schema, error) {
	s, err := jsonschema.For[getPlanInput](nil)
	if err != nil {
		return nil, err
	}
	if f := s.Properties["from"]; f != nil {
		f.Pattern = datePattern
	}
	if t := s.Properties["to"]; t != nil {
		t.Pattern = datePattern
	}
	return s, nil
}

// completeInput is the argument schema for complete_session.
type completeInput struct {
	SessionID int64 `json:"session_id" jsonschema:"id of the session to mark, from get_plan"`
	Done      bool  `json:"done" jsonschema:"true to complete the session, false to undo a completion; defaults to true"`
}

// completeSchema infers and enriches the schema for completeInput.
func completeSchema() (*jsonschema.Schema, error) {
	s, err := jsonschema.For[completeInput](nil)
	if err != nil {
		return nil, err
	}
	if id := s.Properties["session_id"]; id != nil {
		id.Minimum = jsonschema.Ptr(1.0)
	}
	// Make done optional and default it to true, so callers can complete a
	// session by sending only session_id.
	if done := s.Properties["done"]; done != nil {
		done.Default = []byte("true")
	}
	s.Required = without(s.Required, "done")
	return s, nil
}

// without returns names with target removed (order preserved).
func without(names []string, target string) []string {
	out := names[:0:0]
	for _, n := range names {
		if n != target {
			out = append(out, n)
		}
	}
	return out
}
