package plan

import "testing"

func TestTurkishDateFormatting(t *testing.T) {
	if got := FormatDayHeading("2026-10-28"); got != "Çarşamba, 28 Ekim" {
		t.Errorf("FormatDayHeading = %q, want %q", got, "Çarşamba, 28 Ekim")
	}
	if got := FormatShortDate("2026-10-26"); got != "26 Eki" {
		t.Errorf("FormatShortDate = %q, want %q", got, "26 Eki")
	}
	// The range separator must be the en dash U+2013, not an ASCII hyphen.
	const wantRange = "26 Eki–1 Kas"
	if got := FormatWeekRange("2026-10-26", "2026-11-01"); got != wantRange {
		t.Errorf("FormatWeekRange = %q, want %q", got, wantRange)
	}
	if got := WeekdayTR("2026-10-27"); got != "Salı" {
		t.Errorf("WeekdayTR = %q, want Salı", got)
	}
	if got := weekdayLowerTR("2026-10-31"); got != "cumartesi" {
		t.Errorf("weekdayLowerTR = %q, want cumartesi", got)
	}
	if got := WeekdayShortTR("2026-10-26"); got != "Pzt" {
		t.Errorf("WeekdayShortTR = %q, want Pzt", got)
	}
}

func TestFormatHoursTR(t *testing.T) {
	cases := map[int]string{90: "1,5", 540: "9", 75: "1,25", 0: "0", 20520: "342"}
	for min, want := range cases {
		if got := FormatHoursTR(min); got != want {
			t.Errorf("FormatHoursTR(%d) = %q, want %q", min, got, want)
		}
	}
}
