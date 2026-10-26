package plan

import (
	"fmt"
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
		if !validKinds[si.Kind] {
			b.add(p+".kind", CodeInvalid, "kind must be one of aksam, uzun_blok, deneme")
		}

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

// validateItem checks one book reference.
func validateItem(b *errBuilder, p string, it Item) {
	if !validBooks[it.Book] {
		b.add(p+".book", CodeInvalid, "book must be one of M1, M2, M3, G1, G2")
	}
	if it.ChapterNo <= 0 {
		b.add(p+".chapter_no", CodeInvalid, "chapter_no must be greater than zero")
	}
	if it.ChapterTitle == "" {
		b.add(p+".chapter_title", CodeRequired, "chapter_title is required")
	}
	if it.PageFrom <= 0 {
		b.add(p+".page_from", CodeInvalid, "page_from must be greater than zero")
	}
	if it.PageTo <= 0 {
		b.add(p+".page_to", CodeInvalid, "page_to must be greater than zero")
	}
	if it.PageFrom > 0 && it.PageTo > 0 && it.PageFrom > it.PageTo {
		b.add(p+".page_to", CodeRangeOrder, "page_to must not be before page_from")
	}
	validateOptionalRange(b, p, "konu", it.KonuFrom, it.KonuTo)
	validateOptionalRange(b, p, "soru", it.SoruFrom, it.SoruTo)
	validateOptionalRange(b, p, "antrenman", it.AntrenmanFrom, it.AntrenmanTo)
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

// validateSaveOutline checks phases and topic groups without touching the store:
// required fields, date ordering, DGS ranges and key uniqueness within the
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
		case g.DGSQuestionMin == nil && g.DGSQuestionMax == nil:
		case g.DGSQuestionMin == nil || g.DGSQuestionMax == nil:
			b.add(path+".dgs_question_min", CodeInvalid, "dgs_question_min and dgs_question_max must be set together")
		case *g.DGSQuestionMin > *g.DGSQuestionMax:
			b.add(path+".dgs_question_max", CodeRangeOrder, "dgs_question_max must not be before dgs_question_min")
		}
	}

	return b.result()
}
