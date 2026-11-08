package plan

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

const (
	dateLayout = "2006-01-02"
	timeLayout = "15:04"
)

// validDate reports whether s is a real calendar date in 'YYYY-MM-DD' form.
func validDate(s string) bool {
	_, err := time.Parse(dateLayout, s)
	return err == nil
}

// validTime reports whether s is a valid 'HH:MM' time of day.
func validTime(s string) bool {
	_, err := time.Parse(timeLayout, s)
	return err == nil
}

// validateSavePlan checks the request's structure without touching the store:
// dates and times, the [from, to] window, enums, per-item ranges and the
// uniqueness of (date, start_time) within the request. Store-dependent checks
// (topic-group and makeup resolution) happen in the service.
func validateSavePlan(req SavePlanRequest) error {
	var b errBuilder

	fromOK := validDate(req.From)
	toOK := validDate(req.To)
	if !fromOK {
		b.add("from", CodeInvalid, "from must be a date in YYYY-MM-DD form")
	}
	if !toOK {
		b.add("to", CodeInvalid, "to must be a date in YYYY-MM-DD form")
	}
	if fromOK && toOK && req.From > req.To {
		b.add("to", CodeRangeOrder, "to must not be before from")
	}

	seen := map[SlotRef]int{}
	for i, si := range req.Sessions {
		p := fmt.Sprintf("sessions[%d]", i)

		dateOK := validDate(si.Date)
		if !dateOK {
			b.add(p+".date", CodeInvalid, "date must be a date in YYYY-MM-DD form")
		} else if fromOK && toOK && (si.Date < req.From || si.Date > req.To) {
			b.add(p+".date", CodeOutOfRange, "date must fall within [from, to]")
		}

		if !validTime(si.StartTime) {
			b.add(p+".start_time", CodeInvalid, "start_time must be a time in HH:MM form")
		}
		if si.DurationMin <= 0 {
			b.add(p+".duration_min", CodeInvalid, "duration_min must be greater than zero")
		}
		// kind is a free label since WKY-22: no enum gate. A blank or unknown
		// kind is accepted; the screens special-case only the known constants.

		if dateOK && validTime(si.StartTime) {
			slot := si.slot()
			if j, dup := seen[slot]; dup {
				b.add(p, CodeDuplicate, fmt.Sprintf("duplicate (date, start_time) also used by sessions[%d]", j))
			} else {
				seen[slot] = i
			}
		}

		if si.MakeupOf != nil {
			if !validDate(si.MakeupOf.Date) {
				b.add(p+".makeup_of.date", CodeInvalid, "makeup_of.date must be a date in YYYY-MM-DD form")
			}
			if !validTime(si.MakeupOf.StartTime) {
				b.add(p+".makeup_of.start_time", CodeInvalid, "makeup_of.start_time must be a time in HH:MM form")
			}
		}

		if len(si.Items) == 0 {
			b.add(p+".items", CodeRequired, "a session needs at least one item")
		}
		for j, it := range si.Items {
			validateItem(&b, fmt.Sprintf("%s.items[%d]", p, j), it)
		}
	}

	return b.result()
}

// validateItem checks one catalog reference structurally, without touching the
// store: a non-empty source key, positive ordinals, ordered ranges, and a unit
// range only where a section is named. The catalog-reference checks (does the
// source/section/unit exist) run in the service against the stored catalog.
func validateItem(b *errBuilder, p string, it Item) {
	if it.SourceKey == "" {
		b.add(p+".source_key", CodeRequired, "source_key is required")
	}
	if it.SectionOrdinal != nil && *it.SectionOrdinal < 1 {
		b.add(p+".section_ordinal", CodeInvalid, "section_ordinal must be greater than zero")
	}
	if (it.UnitFrom != nil || it.UnitTo != nil) && it.SectionOrdinal == nil {
		b.add(p+".section_ordinal", CodeRequired, "section_ordinal is required when a unit range is given")
	}
	if it.UnitFrom != nil && *it.UnitFrom < 1 {
		b.add(p+".unit_from", CodeInvalid, "unit_from must be greater than zero")
	}
	if it.UnitTo != nil && *it.UnitTo < 1 {
		b.add(p+".unit_to", CodeInvalid, "unit_to must be greater than zero")
	}
	validateOptionalRange(b, p, "unit", it.UnitFrom, it.UnitTo)
	validateOptionalRange(b, p, "loc", it.LocFrom, it.LocTo)
	if it.UnitFrom == nil && it.UnitTo == nil && it.LocFrom == nil && it.LocTo == nil {
		b.add(p+".unit_from", CodeRequired, "an item needs a unit range or a location range")
	}
}

// validateOptionalRange checks a paired optional range: both set or both unset,
// and from <= to.
func validateOptionalRange(b *errBuilder, p, name string, from, to *int) {
	switch {
	case from == nil && to == nil:
		return
	case from == nil || to == nil:
		b.add(fmt.Sprintf("%s.%s_from", p, name), CodeInvalid,
			fmt.Sprintf("%s_from and %s_to must be set together", name, name))
	case *from > *to:
		b.add(fmt.Sprintf("%s.%s_to", p, name), CodeRangeOrder,
			fmt.Sprintf("%s_to must not be before %s_from", name, name))
	}
}

// knownSet reports whether v is one of allowed.
func knownSet(v string, allowed []string) bool {
	for _, a := range allowed {
		if v == a {
			return true
		}
	}
	return false
}

// validateSaveSource checks a source write structurally, without touching the
// store: the source's own fields (key, name, kind and location_unit against their
// closed sets), and each section and unit (required fields, positive and unique
// ordinals, paired/ordered ranges). The catalog is replaced whole, so there is no
// cross-source check here.
func validateSaveSource(req SaveSourceRequest) error {
	var b errBuilder

	if req.Source.Key == "" {
		b.add("source.key", CodeRequired, "key is required")
	}
	if req.Source.Name == "" {
		b.add("source.name", CodeRequired, "name is required")
	}
	if !knownSet(req.Source.Kind, SourceKinds()) {
		b.add("source.kind", CodeInvalid, "kind must be one of book, video, course, other")
	}
	if !knownSet(req.Source.LocationUnit, LocationUnits()) {
		b.add("source.location_unit", CodeInvalid, "location_unit must be one of page, minute, lesson, other")
	}

	secOrdinals := map[int]int{}
	for i, sec := range req.Sections {
		p := fmt.Sprintf("sections[%d]", i)
		if sec.Name == "" {
			b.add(p+".name", CodeRequired, "name is required")
		}
		if sec.Ordinal < 1 {
			b.add(p+".ordinal", CodeInvalid, "ordinal must be greater than zero")
		} else if j, dup := secOrdinals[sec.Ordinal]; dup {
			b.add(p+".ordinal", CodeDuplicate, fmt.Sprintf("duplicate ordinal also used by sections[%d]", j))
		} else {
			secOrdinals[sec.Ordinal] = i
		}
		validateOptionalRange(&b, p, "loc", sec.LocFrom, sec.LocTo)

		unitOrdinals := map[int]int{}
		for k, u := range sec.Units {
			up := fmt.Sprintf("%s.units[%d]", p, k)
			if u.Label == "" {
				b.add(up+".label", CodeRequired, "label is required")
			}
			if u.Ordinal < 1 {
				b.add(up+".ordinal", CodeInvalid, "ordinal must be greater than zero")
			} else if j, dup := unitOrdinals[u.Ordinal]; dup {
				b.add(up+".ordinal", CodeDuplicate, fmt.Sprintf("duplicate ordinal also used by sections[%d].units[%d]", i, j))
			} else {
				unitOrdinals[u.Ordinal] = k
			}
			if u.Number != nil && *u.Number < 1 {
				b.add(up+".number", CodeInvalid, "number must be greater than zero")
			}
			validateOptionalRange(&b, up, "loc", u.LocFrom, u.LocTo)
			if u.Title != nil && len([]rune(*u.Title)) > maxUnitTitleRunes {
				b.add(up+".title", CodeInvalid, "başlık en fazla 200 karakter olabilir")
			}
			if u.URL != nil && !validUnitURL(*u.URL) {
				b.add(up+".url", CodeInvalid, "bağlantı http veya https ile başlayan geçerli bir adres olmalı")
			}
		}
	}

	return b.result()
}

// maxUnitTitleRunes and maxUnitURLBytes bound a unit's optional reference fields.
const (
	maxUnitTitleRunes = 200
	maxUnitURLBytes   = 2048
)

// validUnitURL accepts only absolute http/https links with a host, so a stored
// link is always safe to render as a clickable href.
func validUnitURL(raw string) bool {
	if len(raw) > maxUnitURLBytes {
		return false
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

// normalizeUnitLinks trims each unit's optional title and url and treats a blank
// value as absent, so a form's empty input and an omitted field store the same.
func normalizeUnitLinks(sections []SectionInput) {
	for i := range sections {
		for k := range sections[i].Units {
			u := &sections[i].Units[k]
			u.Title = trimmedOrNil(u.Title)
			u.URL = trimmedOrNil(u.URL)
		}
	}
}

func trimmedOrNil(v *string) *string {
	if v == nil {
		return nil
	}
	t := strings.TrimSpace(*v)
	if t == "" {
		return nil
	}
	return &t
}

// validateSaveOutline checks phases and topic groups without touching the store:
// required fields, date ordering, question ranges and key uniqueness within the
// request.
func validateSaveOutline(req SaveOutlineRequest) error {
	var b errBuilder

	phaseKeys := map[string]int{}
	for i, p := range req.Phases {
		path := fmt.Sprintf("phases[%d]", i)
		if p.Key == "" {
			b.add(path+".key", CodeRequired, "key is required")
		} else if j, dup := phaseKeys[p.Key]; dup {
			b.add(path+".key", CodeDuplicate, fmt.Sprintf("duplicate key also used by phases[%d]", j))
		} else {
			phaseKeys[p.Key] = i
		}
		if p.Name == "" {
			b.add(path+".name", CodeRequired, "name is required")
		}
		sOK, eOK := validDate(p.StartDate), validDate(p.EndDate)
		if !sOK {
			b.add(path+".start_date", CodeInvalid, "start_date must be a date in YYYY-MM-DD form")
		}
		if !eOK {
			b.add(path+".end_date", CodeInvalid, "end_date must be a date in YYYY-MM-DD form")
		}
		if sOK && eOK && p.StartDate > p.EndDate {
			b.add(path+".end_date", CodeRangeOrder, "end_date must not be before start_date")
		}
	}

	groupKeys := map[string]int{}
	for i, g := range req.TopicGroups {
		path := fmt.Sprintf("topic_groups[%d]", i)
		if g.Key == "" {
			b.add(path+".key", CodeRequired, "key is required")
		} else if j, dup := groupKeys[g.Key]; dup {
			b.add(path+".key", CodeDuplicate, fmt.Sprintf("duplicate key also used by topic_groups[%d]", j))
		} else {
			groupKeys[g.Key] = i
		}
		if g.Name == "" {
			b.add(path+".name", CodeRequired, "name is required")
		}
		switch {
		case g.QuestionMin == nil && g.QuestionMax == nil:
		case g.QuestionMin == nil || g.QuestionMax == nil:
			b.add(path+".question_min", CodeInvalid, "question_min and question_max must be set together")
		case *g.QuestionMin > *g.QuestionMax:
			b.add(path+".question_max", CodeRangeOrder, "question_max must not be before question_min")
		}
	}

	return b.result()
}
