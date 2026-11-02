package plan

import "testing"

func ptr(n int) *int { return &n }

// units builds a RefUnit slice from a label and its numbers, the shape the
// catalog resolver hands the formatter.
func units(label string, numbers ...int) []RefUnit {
	out := make([]RefUnit, 0, len(numbers))
	for _, n := range numbers {
		out = append(out, RefUnit{Label: label, Number: n})
	}
	return out
}

func TestFormatRef(t *testing.T) {
	tests := []struct {
		name string
		ref  ItemRef
		want string
	}{
		{
			name: "source, section and a page location range",
			ref:  ItemRef{SourceKey: "M2", SectionOrdinal: ptr(17), LocFrom: ptr(210), LocTo: ptr(225), LocationUnit: "page"},
			want: "M2·17, s. 210–225",
		},
		{
			name: "two unit label groups collapse from one contiguous range",
			ref: ItemRef{SourceKey: "M1", SectionOrdinal: ptr(4),
				Units: append(units("Antrenman", 5, 6, 7), units("Test", 1, 2)...)},
			want: "M1·4, Antrenman 5–7, Test 1–2",
		},
		{
			name: "single unit number does not repeat",
			ref:  ItemRef{SourceKey: "G2", SectionOrdinal: ptr(9), Units: units("Föy", 4)},
			want: "G2·9, Föy 4",
		},
		{
			name: "minute location unit uses its own prefix",
			ref:  ItemRef{SourceKey: "VID", SectionOrdinal: ptr(2), LocFrom: ptr(10), LocTo: ptr(25), LocationUnit: "minute"},
			want: "VID·2, dk 10–25",
		},
		{
			name: "source without a section",
			ref:  ItemRef{SourceKey: "M1", LocFrom: ptr(1), LocTo: ptr(76), LocationUnit: "page"},
			want: "M1, s. 1–76",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatRef(tt.ref); got != tt.want {
				t.Errorf("FormatRef() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestNotationExactRunes pins the middle dot (U+00B7) and en dash (U+2013) so a
// look-alike ASCII hyphen or period can never slip in.
func TestNotationExactRunes(t *testing.T) {
	got := FormatRef(ItemRef{SourceKey: "M2", SectionOrdinal: ptr(17), LocFrom: ptr(210), LocTo: ptr(225), LocationUnit: "page"})
	var sawMiddleDot, sawEnDash bool
	for _, r := range got {
		switch r {
		case '·':
			sawMiddleDot = true
		case '–':
			sawEnDash = true
		case '-':
			t.Errorf("notation %q contains an ASCII hyphen; want en dash U+2013", got)
		}
	}
	if !sawMiddleDot {
		t.Errorf("notation %q is missing the middle dot U+00B7", got)
	}
	if !sawEnDash {
		t.Errorf("notation %q is missing the en dash U+2013", got)
	}
}

func TestFormatRefs(t *testing.T) {
	refs := []ItemRef{
		{SourceKey: "M1", SectionOrdinal: ptr(2), LocFrom: ptr(23), LocTo: ptr(40), LocationUnit: "page"},
		{SourceKey: "M1", SectionOrdinal: ptr(4), Units: units("Antrenman", 1, 2, 3)},
	}
	want := "M1·2, s. 23–40 · M1·4, Antrenman 1–3"
	if got := FormatRefs(refs); got != want {
		t.Errorf("FormatRefs() = %q, want %q", got, want)
	}
	if got := FormatRefs(nil); got != "" {
		t.Errorf("FormatRefs(nil) = %q, want empty", got)
	}
}
