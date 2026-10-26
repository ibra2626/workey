package plan

import (
	"strconv"
	"strings"
)

// Notation runes, named so the exact code points are unmistakable:
//   - middleDot U+00B7 separates book and chapter (M2·17) and joins items;
//   - enDash U+2013 forms numeric ranges (s. 210–225).
//
// Turkish decimal values use a comma, but a book reference carries no decimals,
// so the comma appears only where a formatter prints a measured number.
const (
	middleDot = "·"
	enDash    = "–"
)

// Ref formats an item's book and chapter: "M2·17".
func Ref(it Item) string {
	return it.Book + middleDot + strconv.Itoa(it.ChapterNo)
}

// pageRange formats a page span: "s. 210–225", or "s. 210" when from == to.
func pageRange(from, to int) string {
	if from == to {
		return "s. " + strconv.Itoa(from)
	}
	return "s. " + strconv.Itoa(from) + enDash + strconv.Itoa(to)
}

// numRange formats a plain numeric span: "1–3", or "1" when from == to.
func numRange(from, to int) string {
	if from == to {
		return strconv.Itoa(from)
	}
	return strconv.Itoa(from) + enDash + strconv.Itoa(to)
}

// FormatItem renders one item in the canonical notation, e.g.
// "M2·17, s. 210–225" or, with an antrenman span, "M1·3, s. 41–58, Antr. 1–3".
func FormatItem(it Item) string {
	s := Ref(it) + ", " + pageRange(it.PageFrom, it.PageTo)
	if it.AntrenmanFrom != nil && it.AntrenmanTo != nil {
		s += ", Antr. " + numRange(*it.AntrenmanFrom, *it.AntrenmanTo)
	}
	return s
}

// FormatItems joins the items of a session with " · ", e.g.
// "M1·2, s. 23–40 · M1·3, s. 41–58, Antr. 1–3". An empty slice yields "".
func FormatItems(items []Item) string {
	parts := make([]string, 0, len(items))
	for _, it := range items {
		parts = append(parts, FormatItem(it))
	}
	return strings.Join(parts, " "+middleDot+" ")
}
