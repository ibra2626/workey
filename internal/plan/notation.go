package plan

import (
	"strconv"
	"strings"
)

// Notation runes, named so the exact code points are unmistakable:
//   - middleDot U+00B7 separates a source key from its section ordinal (M1·4)
//     and joins the items of a session;
//   - enDash U+2013 forms numeric ranges (Antrenman 5–7, s. 210–225).
//
// Turkish decimal values use a comma, but a catalog reference carries no
// decimals, so the comma appears only where a formatter prints a measured
// number elsewhere.
const (
	middleDot = "·"
	enDash    = "–"
)

// RefUnit is one unit resolved from the catalog for notation: its source-given
// label and the number shown to the reader (the unit's own number, or its
// ordinal when it has none).
type RefUnit struct {
	Label  string
	Number int
}

// ItemRef is a session item resolved against the catalog, carrying exactly what
// the formatter needs — no raw ids. Units are the resolved units covered by the
// item's unit range, in order.
type ItemRef struct {
	SourceKey      string
	SectionOrdinal *int
	Units          []RefUnit
	LocFrom        *int
	LocTo          *int
	LocationUnit   string // page | minute | lesson | other; prefixes the loc range
}

// head formats the source-and-section lead, e.g. "M1·4" or "M1".
func (r ItemRef) head() string {
	if r.SectionOrdinal != nil {
		return r.SourceKey + middleDot + strconv.Itoa(*r.SectionOrdinal)
	}
	return r.SourceKey
}

// unitGroups renders the item's units as label groups with number ranges, e.g.
// "Antrenman 5–7, Test 1–2". Consecutive units sharing a label collapse into one
// range; a single number does not repeat. Empty when the item names no units.
func (r ItemRef) unitGroups() string {
	if len(r.Units) == 0 {
		return ""
	}
	var parts []string
	i := 0
	for i < len(r.Units) {
		label := r.Units[i].Label
		from := r.Units[i].Number
		to := from
		j := i + 1
		for j < len(r.Units) && r.Units[j].Label == label {
			to = r.Units[j].Number
			j++
		}
		parts = append(parts, strings.TrimSpace(label+" "+numRange(from, to)))
		i = j
	}
	return strings.Join(parts, ", ")
}

// locRange renders the item's location span with a unit-appropriate prefix, e.g.
// "s. 210–225" for pages or "dk 10–25" for minutes. Empty when no range is set.
func (r ItemRef) locRange() string {
	if r.LocFrom == nil || r.LocTo == nil {
		return ""
	}
	return locPrefix(r.LocationUnit) + numRange(*r.LocFrom, *r.LocTo)
}

// locPrefix maps a source's location unit to the short label that precedes a
// location range. 'page' keeps the v1 "s. " (sayfa); an unknown unit prints no
// prefix so the bare range still reads.
func locPrefix(unit string) string {
	switch unit {
	case "page":
		return "s. "
	case "minute":
		return "dk "
	case "lesson":
		return "ders "
	default:
		return ""
	}
}

// numRange formats an inclusive numeric span: "5–7", or "5" when from == to.
func numRange(from, to int) string {
	if from == to {
		return strconv.Itoa(from)
	}
	return strconv.Itoa(from) + enDash + strconv.Itoa(to)
}

// FormatRef renders one resolved item in the canonical notation, e.g.
// "M1·4, Antrenman 5–7, Test 1–2" or "M2·17, s. 210–225". The parts present
// (section, units, location) join with ", ".
func FormatRef(r ItemRef) string {
	parts := []string{r.head()}
	if g := r.unitGroups(); g != "" {
		parts = append(parts, g)
	}
	if l := r.locRange(); l != "" {
		parts = append(parts, l)
	}
	return strings.Join(parts, ", ")
}

// FormatRefs joins a session's resolved items with " · ", e.g.
// "M1·4, Antrenman 5–7 · M1·5, s. 59–76". An empty slice yields "".
func FormatRefs(refs []ItemRef) string {
	parts := make([]string, 0, len(refs))
	for _, r := range refs {
		parts = append(parts, FormatRef(r))
	}
	return strings.Join(parts, " "+middleDot+" ")
}
