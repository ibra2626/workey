package plan

import "testing"

func ptr(n int) *int { return &n }

func TestFormatItem(t *testing.T) {
	tests := []struct {
		name string
		item Item
		want string
	}{
		{
			name: "brief example from the spec",
			item: Item{Book: "M2", ChapterNo: 17, ChapterTitle: "Oran orantı", PageFrom: 210, PageTo: 225},
			want: "M2·17, s. 210–225",
		},
		{
			name: "turkish chapter title does not affect notation, antrenman span",
			item: Item{Book: "M1", ChapterNo: 3, ChapterTitle: "Üslü ifadeler", PageFrom: 41, PageTo: 58, AntrenmanFrom: ptr(1), AntrenmanTo: ptr(3)},
			want: "M1·3, s. 41–58, Antr. 1–3",
		},
		{
			name: "single page and single antrenman collapse the range",
			item: Item{Book: "G2", ChapterNo: 9, ChapterTitle: "Çember", PageFrom: 120, PageTo: 120, AntrenmanFrom: ptr(4), AntrenmanTo: ptr(4)},
			want: "G2·9, s. 120, Antr. 4",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatItem(tt.item); got != tt.want {
				t.Errorf("FormatItem() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestNotationExactRunes pins the middle dot (U+00B7) and en dash (U+2013) so a
// look-alike ASCII hyphen or period can never slip in.
func TestNotationExactRunes(t *testing.T) {
	got := FormatItem(Item{Book: "M2", ChapterNo: 17, ChapterTitle: "x", PageFrom: 210, PageTo: 225})
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

func TestFormatItems(t *testing.T) {
	items := []Item{
		{Book: "M1", ChapterNo: 2, ChapterTitle: "Sayı basamakları", PageFrom: 23, PageTo: 40},
		{Book: "M1", ChapterNo: 3, ChapterTitle: "Üslü ifadeler", PageFrom: 41, PageTo: 58, AntrenmanFrom: ptr(1), AntrenmanTo: ptr(3)},
	}
	want := "M1·2, s. 23–40 · M1·3, s. 41–58, Antr. 1–3"
	if got := FormatItems(items); got != want {
		t.Errorf("FormatItems() = %q, want %q", got, want)
	}
	if got := FormatItems(nil); got != "" {
		t.Errorf("FormatItems(nil) = %q, want empty", got)
	}
}
