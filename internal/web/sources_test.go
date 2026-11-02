package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ibra2626/workey/internal/plan"
)

func ip(v int) *int { return &v }

// seedM1 writes the sample book source M1 (one section, two units) through the
// service and returns its read view.
func seedM1(t *testing.T, a *Auth) plan.SourceView {
	t.Helper()
	ctx := context.Background()
	req := plan.SaveSourceRequest{
		Source: plan.SourceInput{Key: "M1", Name: "Matematik 1", Kind: plan.SourceKindBook, LocationUnit: plan.LocationUnitPage, Ordinal: 1},
		Sections: []plan.SectionInput{{Ordinal: 1, Name: "Sayı basamakları", LocFrom: ip(1), LocTo: ip(10), Units: []plan.UnitInput{
			{Ordinal: 1, Label: "Antrenman", Number: ip(1), LocFrom: ip(1), LocTo: ip(5)},
			{Ordinal: 2, Label: "Antrenman", Number: ip(2), LocFrom: ip(6), LocTo: ip(10)},
		}}},
	}
	res, err := a.svc.SaveSource(ctx, plan.SourceAPI, req)
	if err != nil {
		t.Fatalf("SaveSource M1: %v", err)
	}
	return res.Source
}

// m1EditForm builds a no-JS edit form for M1 that keeps every ordinal (so unit
// progress survives) and applies the given source name.
func m1EditForm(csrf, name string) string {
	v := url.Values{}
	v.Set("_csrf", csrf)
	v.Set("ordinal", "1")
	v.Set("name", name)
	v.Set("kind", plan.SourceKindBook)
	v.Set("location_unit", plan.LocationUnitPage)
	v.Set("section_ordinals", "1")
	v.Set("s1_name", "Sayı basamakları")
	v.Set("s1_loc_from", "1")
	v.Set("s1_loc_to", "10")
	v.Set("s1_unit_ordinals", "1,2")
	v.Set("s1u1_label", "Antrenman")
	v.Set("s1u1_number", "1")
	v.Set("s1u1_loc_from", "1")
	v.Set("s1u1_loc_to", "5")
	v.Set("s1u2_label", "Antrenman")
	v.Set("s1u2_number", "2")
	v.Set("s1u2_loc_from", "6")
	v.Set("s1u2_loc_to", "10")
	return v.Encode()
}

func TestSourcesPageRendersCatalogAndReport(t *testing.T) {
	ctx := context.Background()
	a := newTestAuth(t)
	src := seedM1(t, a)
	unitID := src.Sections[0].Units[0].ID
	if _, err := a.svc.SetUnitResult(ctx, plan.SourceAPI, plan.UnitResultRequest{UnitID: unitID, Done: true, Correct: ip(4), Wrong: ip(1), Blank: ip(0)}); err != nil {
		t.Fatalf("SetUnitResult: %v", err)
	}
	body := getAuthed(t, a, "/kaynaklar")
	for _, want := range []string{
		"Kaynaklar", "Matematik 1", "Sayı basamakları", "Antrenman",
		"Kitap", "Sayfa", "Tamamlanan", "Doğru", "Yanlış", "Boş",
		`name="section_ordinals"`, `name="s1_unit_ordinals"`, `aria-current="page"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("Kaynaklar body missing %q", want)
		}
	}
	// Reporting figures must match the aggregates: DoneCount 1 and Correct 4.
	for _, want := range []string{">1</dd>", ">4</dd>"} {
		if !strings.Contains(body, want) {
			t.Errorf("report figure missing %q", want)
		}
	}
}

func TestSourcesEditPreservesUnitProgress(t *testing.T) {
	ctx := context.Background()
	a := newTestAuth(t)
	src := seedM1(t, a)
	unitID := src.Sections[0].Units[0].ID
	if _, err := a.svc.SetUnitResult(ctx, plan.SourceAPI, plan.UnitResultRequest{UnitID: unitID, Done: true, Correct: ip(4), Wrong: ip(1), Blank: ip(0)}); err != nil {
		t.Fatalf("SetUnitResult: %v", err)
	}
	cookie := login(t, a)
	sid, _ := a.decodeSession(cookie.Value)

	// No-JS form POST (token in the hidden field) renames the source.
	req := httptest.NewRequest(http.MethodPost, "/kaynaklar/M1",
		strings.NewReader(m1EditForm(a.csrfToken(sid), "Matematik 1 — düzeltildi")))
	req.AddCookie(cookie)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := do(t, a, req)
	res.Body.Close()
	if res.StatusCode != http.StatusSeeOther || res.Header.Get("Location") != "/kaynaklar" {
		t.Fatalf("edit status = %d loc %q, want 303 /kaynaklar", res.StatusCode, res.Header.Get("Location"))
	}

	cat, err := a.svc.Sources(ctx)
	if err != nil || len(cat.Sources) != 1 {
		t.Fatalf("read back: %v", err)
	}
	got := cat.Sources[0]
	if got.Name != "Matematik 1 — düzeltildi" {
		t.Errorf("name not saved: %q", got.Name)
	}
	if got.Sections[0].Units[0].ID != unitID {
		t.Errorf("unit id changed %d -> %d (progress would be lost)", unitID, got.Sections[0].Units[0].ID)
	}
	prog, err := a.svc.SourceProgress(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if tot := prog[got.ID]; tot.DoneCount != 1 || tot.Correct != 4 || tot.Wrong != 1 {
		t.Errorf("progress not preserved across edit: %+v", tot)
	}
}

func TestSourcesEditRequiresCSRF(t *testing.T) {
	a := newTestAuth(t)
	seedM1(t, a)
	cookie := login(t, a)

	// No token at all: forbidden, nothing written.
	req := httptest.NewRequest(http.MethodPost, "/kaynaklar/M1",
		strings.NewReader(m1EditForm("", "Değişmemeli")))
	req.AddCookie(cookie)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := do(t, a, req)
	res.Body.Close()
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("POST without CSRF status = %d, want 403", res.StatusCode)
	}
	cat, _ := a.svc.Sources(context.Background())
	if cat.Sources[0].Name != "Matematik 1" {
		t.Errorf("source changed without CSRF: %q", cat.Sources[0].Name)
	}
}

func TestSourcesVideoShowsDoneOnly(t *testing.T) {
	ctx := context.Background()
	a := newTestAuth(t)
	req := plan.SaveSourceRequest{
		Source: plan.SourceInput{Key: "V1", Name: "Ders videoları", Kind: plan.SourceKindVideo, LocationUnit: plan.LocationUnitMinute, Ordinal: 1},
		Sections: []plan.SectionInput{{Ordinal: 1, Name: "Giriş", Units: []plan.UnitInput{
			{Ordinal: 1, Label: "Ders", Number: ip(1)},
		}}},
	}
	res, err := a.svc.SaveSource(ctx, plan.SourceAPI, req)
	if err != nil {
		t.Fatalf("SaveSource V1: %v", err)
	}
	unitID := res.Source.Sections[0].Units[0].ID
	if _, err := a.svc.SetUnitResult(ctx, plan.SourceAPI, plan.UnitResultRequest{UnitID: unitID, Done: true}); err != nil {
		t.Fatalf("SetUnitResult video: %v", err)
	}
	body := getAuthed(t, a, "/kaynaklar")
	if !strings.Contains(body, "Tamamlanan") || !strings.Contains(body, "Video") {
		t.Error("video card missing done count")
	}
	if strings.Contains(body, "Doğru") || strings.Contains(body, "Yanlış") {
		t.Error("video card must not show D/Y/B")
	}
}
