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

// sourceKindEnum and locationUnitEnum list the closed sets the store enforces on
// a source; sessionKinds are the known session kinds, given only as examples
// since a session's kind is a free label.
var (
	sourceKindEnum   = toAnySlice(plan.SourceKinds())
	locationUnitEnum = toAnySlice(plan.LocationUnits())
	sessionKinds     = []string{plan.KindAksam, plan.KindUzunBlok, plan.KindDeneme}
)

// toAnySlice widens a string slice to the []any an enum schema expects.
func toAnySlice(ss []string) []any {
	out := make([]any, len(ss))
	for i, s := range ss {
		out[i] = s
	}
	return out
}

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
	setProp(sess, "kind", "Session kind, a free label. Known values used by the screens: 'aksam' (evening block), 'uzun_blok' (long block), 'deneme' (mock exam, no results are stored). Any other label is accepted.")
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

	items := setProp(sess, "items", "One or more catalog references for the session. Only source/section/unit/location references are stored — never topic explanations or question content. The source catalog must already exist (save_source).")
	if items != nil {
		items.MinItems = jsonschema.Ptr(1)
	}
	enrichItemSchema(itemsOf(items))
}

// enrichItemSchema annotates one item object's schema (the new catalog reference
// shape: a source key, an optional section ordinal, and a unit range and/or a
// location range).
func enrichItemSchema(it *jsonschema.Schema) {
	if it == nil {
		return
	}
	setProp(it, "source_key", "Stable key of the source this item references (from save_source / get_sources). Required.")
	boundOptional(it, "section_ordinal", "Optional 1-based ordinal of the section within the source.")
	boundOptional(it, "unit_from", "Optional first unit ordinal of the range (1-based, within the section). Requires section_ordinal.")
	boundOptional(it, "unit_to", "Optional last unit ordinal of the range; must be >= unit_from.")
	boundOptional(it, "loc_from", "Optional first location number of the range, in the source's location unit (page/minute/lesson/other).")
	boundOptional(it, "loc_to", "Optional last location number; must be >= loc_from.")
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
		boundOptional(g, "question_min", "Optional lower bound of the group's expected question count.")
		boundOptional(g, "question_max", "Optional upper bound of the group's expected question count.")
	}

	return s, nil
}

// saveSourceSchema infers and enriches the schema for plan.SaveSourceRequest: a
// source (with its closed kind/location-unit sets) and its full ordered set of
// sections, each carrying its ordered units.
func saveSourceSchema() (*jsonschema.Schema, error) {
	s, err := jsonschema.For[plan.SaveSourceRequest](nil)
	if err != nil {
		return nil, err
	}
	setProp(s, "idempotency_key", "Optional caller-chosen key, with the same replay semantics as save_plan.")

	src := setProp(s, "source", "The source's own fields. The structure is replaced whole, keyed by 'key'.")
	if src != nil {
		setProp(src, "key", "Stable machine key for the source (e.g. 'M1', 'algoritma-101'). Re-writing the same key preserves unit progress where unit places are kept.")
		setProp(src, "name", "Display name of the source.")
		kind := setProp(src, "kind", "Source kind: book, video, course or other. A video source's units carry no doğru/yanlış/boş.")
		if kind != nil {
			kind.Enum = sourceKindEnum
		}
		lu := setProp(src, "location_unit", "How the source counts ranges: page, minute, lesson or other. Drives the notation prefix (page → 's. ', minute → 'dk ', lesson → 'ders ').")
		if lu != nil {
			lu.Enum = locationUnitEnum
		}
		ord := setProp(src, "ordinal", "Order of the source in the catalog, from 1.")
		if ord != nil {
			ord.Minimum = jsonschema.Ptr(0.0)
		}
	}

	sections := setProp(s, "sections", "The complete ordered set of sections for the source. A section is matched by its ordinal; an ordinal absent from the request is deleted (cascading its units and their progress).")
	if sec := itemsOf(sections); sec != nil {
		so := setProp(sec, "ordinal", "1-based order of the section within the source; also its stable match key across re-writes.")
		if so != nil {
			so.Minimum = jsonschema.Ptr(1.0)
		}
		setProp(sec, "name", "Section name as printed in the source (a reference label, not topic content).")
		boundOptional(sec, "loc_from", "Optional first location number of the section's span, in the source's location unit.")
		boundOptional(sec, "loc_to", "Optional last location number; must be >= loc_from.")

		units := setProp(sec, "units", "The complete ordered set of units in the section. A unit is matched by its ordinal within the section; keeping the ordinal keeps the unit id and its progress.")
		if u := itemsOf(units); u != nil {
			uo := setProp(u, "ordinal", "1-based order of the unit within the section; its stable match key across re-writes.")
			if uo != nil {
				uo.Minimum = jsonschema.Ptr(1.0)
			}
			setProp(u, "label", "The source's own word for the unit (e.g. 'Antrenman', 'Test', 'Föy', 'Ders'). Free text; never enumerated.")
			boundOptional(u, "number", "Optional number the reader sees (e.g. Antrenman 5); defaults to the ordinal when absent.")
			boundOptional(u, "loc_from", "Optional first location number of the unit, in the source's location unit.")
			boundOptional(u, "loc_to", "Optional last location number; must be >= loc_from.")
			setProp(u, "title", "Optional title the source itself gives this unit (e.g. the video's title), at most 200 characters. A reference, never topic content. save_source is a full replace: send it again on every re-write or it is cleared.")
			setProp(u, "url", "Optional http/https link that opens this unit (e.g. the video's URL). Never invent one. save_source is a full replace: send it again on every re-write or it is cleared.")
		}
	}

	return s, nil
}

// setUnitResultSchema infers and enriches the schema for plan.UnitResultRequest.
func setUnitResultSchema() (*jsonschema.Schema, error) {
	s, err := jsonschema.For[plan.UnitResultRequest](nil)
	if err != nil {
		return nil, err
	}
	id := setProp(s, "unit_id", "Id of the unit to mark, from get_sources.")
	if id != nil {
		id.Minimum = jsonschema.Ptr(1.0)
	}
	setProp(s, "done", "true to mark the unit done, false to undo. Independent of session completion.")
	boundOptional(s, "correct", "Optional doğru count (non-negative). Not allowed for a video source's unit.")
	boundOptional(s, "wrong", "Optional yanlış count (non-negative). Not allowed for a video source's unit.")
	boundOptional(s, "blank", "Optional boş count (non-negative). Not allowed for a video source's unit.")
	// D/Y/B are non-negative, not strictly positive; relax the shared min to 0.
	for _, name := range []string{"correct", "wrong", "blank"} {
		if c := s.Properties[name]; c != nil {
			c.Minimum = jsonschema.Ptr(0.0)
		}
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
