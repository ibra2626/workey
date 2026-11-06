package views

import (
	"strconv"

	"github.com/a-h/templ"

	"github.com/ibra2626/workey/internal/plan"
)

// SourcesPage is the Kaynaklar screen model: the whole catalog (sources →
// sections → units) plus the per-source and per-section progress aggregates, the
// CSRF token for the edit forms, and an optional feedback banner after a save.
type SourcesPage struct {
	Catalog        *plan.CatalogView
	SourceProgress map[int64]plan.ProgressTotals // by source id
	SectionProg    map[int64]plan.ProgressTotals // by section id
	CSRF           string
	Feedback       string
	Invalid        bool
}

// sourceTotals returns a source's progress totals, or the zero value when it has
// no done units yet.
func (p SourcesPage) sourceTotals(id int64) plan.ProgressTotals {
	return p.SourceProgress[id]
}

// sectionTotals returns a section's progress totals, or the zero value.
func (p SourcesPage) sectionTotals(id int64) plan.ProgressTotals {
	return p.SectionProg[id]
}

// isVideoKind reports whether a source kind carries no D/Y/B (a video is marked
// done only; see plan.SourceKindVideo).
func isVideoKind(kind string) bool {
	return kind == plan.SourceKindVideo
}

// kindLabel and unitLabel render the catalog's closed-set codes in Turkish for
// the select controls; an unknown code renders as itself.
func kindLabel(kind string) string {
	switch kind {
	case plan.SourceKindBook:
		return "Kitap"
	case plan.SourceKindVideo:
		return "Video"
	case plan.SourceKindCourse:
		return "Kurs"
	case plan.SourceKindOther:
		return "Diğer"
	default:
		return kind
	}
}

func unitLabel(unit string) string {
	switch unit {
	case plan.LocationUnitPage:
		return "Sayfa"
	case plan.LocationUnitMinute:
		return "Dakika"
	case plan.LocationUnitLesson:
		return "Ders"
	case plan.LocationUnitOther:
		return "Diğer"
	default:
		return unit
	}
}

// sourceKindOptions and locationUnitOptions list the accepted values for the
// selects, straight from the plan package's closed sets.
func sourceKindOptions() []string   { return plan.SourceKinds() }
func locationUnitOptions() []string { return plan.LocationUnits() }

// --- catalog summary helpers --------------------------------------------
//
// The Kaynaklar page shows each source and section as a summary line first and
// opens the editable fields on demand, so the catalog reviews in two screens
// instead of hundreds of always-open inputs.

// sourceUnitCount is the total number of study units under a source.
func sourceUnitCount(src plan.SourceView) int {
	n := 0
	for i := range src.Sections {
		n += len(src.Sections[i].Units)
	}
	return n
}

// unitNoun is the word for a study unit in the user's terms: "ders" for a
// lesson-located source (video/course), "birim" otherwise.
func unitNoun(src plan.SourceView) string {
	if src.LocationUnit == plan.LocationUnitLesson {
		return "ders"
	}
	return "birim"
}

// locRangeLabel renders a section's or unit's location span in the source's own
// notation: "s. 210–225" for pages, "1–15. ders" for lessons, "20–40. dk" for
// minutes. Returns "" when no bound is set.
func locRangeLabel(unit string, from, to *int) string {
	if from == nil && to == nil {
		return ""
	}
	span := ""
	switch {
	case from != nil && to != nil && *from != *to:
		span = strconv.Itoa(*from) + "–" + strconv.Itoa(*to)
	case from != nil:
		span = strconv.Itoa(*from)
	default:
		span = strconv.Itoa(*to)
	}
	switch unit {
	case plan.LocationUnitPage:
		return "s. " + span
	case plan.LocationUnitLesson:
		return span + ". ders"
	case plan.LocationUnitMinute:
		return span + ". dk"
	default:
		return span
	}
}

// --- edit-form field helpers --------------------------------------------
//
// The whole source is re-submitted on every save; the hidden ordinal lists let
// the handler walk the structure and keep each section/unit ordinal (and so its
// id and unit progress). Field names encode the ordinals: section fields are
// "s<ord>_<suffix>", unit fields "s<sec>u<unit>_<suffix>".

func secFieldName(ordinal int, suffix string) string {
	return "s" + strconv.Itoa(ordinal) + "_" + suffix
}

func unitFieldName(secOrdinal, unitOrdinal int, suffix string) string {
	return "s" + strconv.Itoa(secOrdinal) + "u" + strconv.Itoa(unitOrdinal) + "_" + suffix
}

// sectionOrdinalsCSV / unitOrdinalsCSV render the hidden ordinal lists.
func sectionOrdinalsCSV(src plan.SourceView) string {
	return ordinalsCSV(len(src.Sections), func(i int) int { return src.Sections[i].Ordinal })
}

func unitOrdinalsCSV(sec plan.SectionView) string {
	return ordinalsCSV(len(sec.Units), func(i int) int { return sec.Units[i].Ordinal })
}

func ordinalsCSV(n int, at func(int) int) string {
	out := ""
	for i := 0; i < n; i++ {
		if i > 0 {
			out += ","
		}
		out += strconv.Itoa(at(i))
	}
	return out
}

// optIntValue renders an optional integer field as an input value: empty for
// nil, the decimal otherwise.
func optIntValue(p *int) string {
	if p == nil {
		return ""
	}
	return strconv.Itoa(*p)
}

// sourceActionURL is the save route for a source's edit form.
func sourceActionURL(key string) string {
	return "/kaynaklar/" + key
}

// sourceFormAttrs returns the htmx attributes for a source's edit form: the
// whole Kaynaklar main content (#kaynaklar-main) is swapped so the feedback
// banner and every card's figures refresh together. The plain action keeps the
// no-JS form POST working.
func sourceFormAttrs(key string) templ.Attributes {
	url := sourceActionURL(key)
	return templ.Attributes{
		"action":          url,
		"method":          "post",
		"hx-post":         url,
		"hx-target":       "#kaynaklar-main",
		"hx-swap":         "outerHTML",
		"hx-disabled-elt": "#kaynaklar-main button",
		"hx-sync":         "#kaynaklar-main:drop",
		"hx-request":      `{"timeout":20000}`,
	}
}
