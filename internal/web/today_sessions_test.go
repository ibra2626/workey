package web

import (
	"context"
	"strings"
	"testing"

	"github.com/ibra2626/workey/internal/plan"
)

func intp(n int) *int { return &n }

// seedTwoBlocks writes today's sample day as two blocks over M1·3 Antrenman 1–2
// and 3, and gives Antrenman 3 a title and link.
func seedTwoBlocks(t *testing.T, a *Auth) {
	t.Helper()
	ctx := context.Background()
	cat, err := a.svc.Sources(ctx)
	if err != nil {
		t.Fatalf("Sources: %v", err)
	}
	var req plan.SaveSourceRequest
	for _, src := range cat.Sources {
		if src.Key != "M1" {
			continue
		}
		req.Source = plan.SourceInput{Key: src.Key, Name: src.Name, Kind: src.Kind, LocationUnit: src.LocationUnit, Ordinal: src.Ordinal}
		for _, sec := range src.Sections {
			in := plan.SectionInput{Ordinal: sec.Ordinal, Name: sec.Name, LocFrom: sec.LocFrom, LocTo: sec.LocTo}
			for _, u := range sec.Units {
				ui := plan.UnitInput{Ordinal: u.Ordinal, Label: u.Label, Number: u.Number, LocFrom: u.LocFrom, LocTo: u.LocTo}
				if sec.Ordinal == 3 && u.Ordinal == 3 {
					title, link := "Üçüncü video", "https://example.com/v3"
					ui.Title, ui.URL = &title, &link
				}
				in.Units = append(in.Units, ui)
			}
			req.Sections = append(req.Sections, in)
		}
	}
	if _, err := a.svc.SaveSource(ctx, plan.SourceAPI, req); err != nil {
		t.Fatalf("SaveSource: %v", err)
	}
	item := func(from, to int) []plan.Item {
		return []plan.Item{{SourceKey: "M1", SectionOrdinal: intp(3), UnitFrom: intp(from), UnitTo: intp(to)}}
	}
	day := wedDate.String()
	if _, err := a.svc.SavePlan(ctx, plan.SourceAPI, plan.SavePlanRequest{From: day, To: day, Sessions: []plan.SessionInput{
		{Date: day, StartTime: "14:00", DurationMin: 90, Kind: plan.KindAksam, Items: item(1, 2)},
		{Date: day, StartTime: "16:30", DurationMin: 120, Kind: plan.KindUzunBlok, Items: item(3, 3)},
	}}); err != nil {
		t.Fatalf("SavePlan: %v", err)
	}
}

func TestBugunShowsSecondBlockWithTimeAndLink(t *testing.T) {
	st := newStore(t)
	a := newAuthOn(t, st)
	seedWeek(t, st)
	seedTwoBlocks(t, a)

	body := getAuthed(t, a, "/")
	for _, want := range []string{
		"16:30–18:30 · 120 dk",
		`href="https://example.com/v3"`,
		`rel="noopener noreferrer"`,
		"Üçüncü video",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("Bugün body missing %q", want)
		}
	}
}

func TestClosedSessionLeftoverReported(t *testing.T) {
	ctx := context.Background()
	st := newStore(t)
	a := newAuthOn(t, st)
	seedWeek(t, st)
	seedTwoBlocks(t, a)

	day, err := a.svc.Day(ctx, wedDate)
	if err != nil || len(day.Sessions) != 2 {
		t.Fatalf("Day: %v (%d sessions)", err, len(day.Sessions))
	}
	if _, err := a.svc.Complete(ctx, plan.SourceWeb, day.Sessions[0].ID); err != nil {
		t.Fatalf("complete first block: %v", err)
	}

	got, err := a.svc.LeftoverUnits(ctx)
	if err != nil {
		t.Fatalf("LeftoverUnits: %v", err)
	}
	if len(got) != 2 || got[0].Notation != "M1·3, Antrenman 1" || got[0].SessionDate != wedDate.String() {
		t.Fatalf("leftovers = %+v", got)
	}

	body := getAuthed(t, a, "/")
	for _, want := range []string{"Kalan birimler", "Kalan: M1·3, Antrenman 1", "henüz planda yok", "16:30–18:30"} {
		if !strings.Contains(body, want) {
			t.Errorf("Bugün body missing %q", want)
		}
	}
}
