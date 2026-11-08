package plan

import (
	"context"
	"testing"
)

func strp(s string) *string { return &s }

func TestValidateUnitLinks(t *testing.T) {
	base := func(u UnitInput) SaveSourceRequest {
		u.Ordinal, u.Label = 1, "Ders"
		return SaveSourceRequest{
			Source:   SourceInput{Key: "V", Name: "V", Kind: SourceKindVideo, LocationUnit: LocationUnitLesson, Ordinal: 1},
			Sections: []SectionInput{{Ordinal: 1, Name: "S", Units: []UnitInput{u}}},
		}
	}
	cases := []struct {
		name string
		url  string
		ok   bool
	}{
		{"https", "https://www.youtube.com/watch?v=abc", true},
		{"http", "http://example.com/a", true},
		{"javascript", "javascript:alert(1)", false},
		{"relative", "/watch?v=abc", false},
		{"no host", "https://", false},
	}
	for _, c := range cases {
		err := validateSaveSource(base(UnitInput{URL: strp(c.url)}))
		if (err == nil) != c.ok {
			t.Errorf("%s: err = %v, want ok=%v", c.name, err, c.ok)
		}
	}
}

func TestNormalizeUnitLinksTreatsBlankAsAbsent(t *testing.T) {
	secs := []SectionInput{{Units: []UnitInput{{Title: strp("  "), URL: strp(" https://a.example/x ")}}}}
	normalizeUnitLinks(secs)
	if secs[0].Units[0].Title != nil {
		t.Errorf("blank title kept: %q", *secs[0].Units[0].Title)
	}
	if got := secs[0].Units[0].URL; got == nil || *got != "https://a.example/x" {
		t.Errorf("url = %v", got)
	}
}

func TestUnitSlotOrdering(t *testing.T) {
	a := unitSlot{date: "2026-11-08", start: "14:00"}
	b := unitSlot{date: "2026-11-08", start: "16:30"}
	c := unitSlot{date: "2026-11-12", start: "09:00"}
	if !a.before(b) || !b.before(c) || c.before(a) || a.before(a) {
		t.Fatal("unitSlot ordering wrong")
	}
}

func TestAnnotateCarriedOverAndNext(t *testing.T) {
	tl := &unitTimeline{
		closed: map[int64][]unitSlot{5: {{date: "2026-11-08", start: "14:00"}}},
		open:   map[int64][]unitSlot{5: {{date: "2026-11-12", start: "20:00"}}},
	}
	closedRow := UnitRow{UnitID: 5}
	tl.annotate(&closedRow, SessionView{Date: "2026-11-08", StartTime: "14:00"})
	if closedRow.CarriedOver || closedRow.NextDate != "2026-11-12" {
		t.Errorf("closed session row = %+v", closedRow)
	}
	nextRow := UnitRow{UnitID: 5}
	tl.annotate(&nextRow, SessionView{Date: "2026-11-12", StartTime: "20:00"})
	if !nextRow.CarriedOver || nextRow.NextDate != "" {
		t.Errorf("next session row = %+v", nextRow)
	}
}

var _ = context.Background
