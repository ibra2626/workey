package plan

import (
	"strconv"
	"strings"
	"time"
)

// Turkish calendar names. Weekday tables are indexed by time.Weekday
// (Sunday=0 .. Saturday=6); month tables are indexed by time.Month (1..12), so
// index 0 is an unused placeholder. Lowercase forms are written out rather than
// produced by case conversion, so the Turkish i/İ and ı/I rules never apply by
// accident.
var (
	weekdaysTR      = [7]string{"Pazar", "Pazartesi", "Salı", "Çarşamba", "Perşembe", "Cuma", "Cumartesi"}
	weekdaysLowerTR = [7]string{"pazar", "pazartesi", "salı", "çarşamba", "perşembe", "cuma", "cumartesi"}
	weekdaysShortTR = [7]string{"Paz", "Pzt", "Sal", "Çar", "Per", "Cum", "Cmt"}

	monthsTR      = [13]string{"", "Ocak", "Şubat", "Mart", "Nisan", "Mayıs", "Haziran", "Temmuz", "Ağustos", "Eylül", "Ekim", "Kasım", "Aralık"}
	monthsShortTR = [13]string{"", "Oca", "Şub", "Mar", "Nis", "May", "Haz", "Tem", "Ağu", "Eyl", "Eki", "Kas", "Ara"}
)

// parseCivil parses a 'YYYY-MM-DD' date string. It reads the date in UTC so the
// weekday and components depend only on the civil date, not on any zone.
func parseCivil(date string) (time.Time, bool) {
	t, err := time.Parse(dateLayout, date)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

// WeekdayTR returns the capitalised Turkish weekday for a date, e.g.
// "Çarşamba"; an unparseable date yields "".
func WeekdayTR(date string) string {
	t, ok := parseCivil(date)
	if !ok {
		return ""
	}
	return weekdaysTR[int(t.Weekday())]
}

// weekdayLowerTR returns the lowercase Turkish weekday, e.g. "cumartesi".
func weekdayLowerTR(date string) string {
	t, ok := parseCivil(date)
	if !ok {
		return ""
	}
	return weekdaysLowerTR[int(t.Weekday())]
}

// WeekdayShortTR returns the three-letter Turkish weekday, e.g. "Pzt".
func WeekdayShortTR(date string) string {
	t, ok := parseCivil(date)
	if !ok {
		return ""
	}
	return weekdaysShortTR[int(t.Weekday())]
}

// FormatDayHeading renders a full day heading, e.g. "Çarşamba, 28 Ekim".
func FormatDayHeading(date string) string {
	t, ok := parseCivil(date)
	if !ok {
		return ""
	}
	return weekdaysTR[int(t.Weekday())] + ", " + strconv.Itoa(t.Day()) + " " + monthsTR[int(t.Month())]
}

// FormatShortDate renders a day with its abbreviated month, e.g. "26 Eki".
func FormatShortDate(date string) string {
	t, ok := parseCivil(date)
	if !ok {
		return ""
	}
	return strconv.Itoa(t.Day()) + " " + monthsShortTR[int(t.Month())]
}

// FormatWeekRange renders an inclusive date span with an en dash and no spaces
// around it, e.g. "26 Eki–1 Kas".
func FormatWeekRange(start, end string) string {
	return FormatShortDate(start) + enDash + FormatShortDate(end)
}

// FormatHoursTR renders a minute count as Turkish hours with a decimal comma:
// 90 → "1,5", 540 → "9", 75 → "1,25". Whole hours carry no fractional part.
func FormatHoursTR(min int) string {
	s := strconv.FormatFloat(float64(min)/60, 'f', 2, 64)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return strings.Replace(s, ".", ",", 1)
}
