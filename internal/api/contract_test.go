package api_test

import (
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/ibra2626/workey/internal/apidoc"
	"github.com/ibra2626/workey/internal/plan"
)

func TestOpenAPIContract(t *testing.T) {
	h, _ := newAPI(t)
	for _, path := range []string{"/api/v1/agent-guide", "/api/v1/openapi.json"} {
		r := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != 401 {
			t.Fatalf("unprotected %s: %d", path, w.Code)
		}
	}
	rec := do(t, h, "GET", "/api/v1/openapi.json", nil, nil)
	var spec struct {
		OpenAPI    string                                `json:"openapi"`
		Paths      map[string]map[string]json.RawMessage `json:"paths"`
		Components struct {
			Schemas map[string]json.RawMessage `json:"schemas"`
		} `json:"components"`
	}
	decode(t, rec, &spec)
	if spec.OpenAPI != "3.1.0" {
		t.Fatal(spec.OpenAPI)
	}
	validate := func(name string, data []byte) {
		t.Helper()
		var s jsonschema.Schema
		if err := json.Unmarshal(spec.Components.Schemas[name], &s); err != nil {
			t.Fatal(err)
		}
		rs, err := s.Resolve(nil)
		if err != nil {
			t.Fatalf("%s resolve: %v", name, err)
		}
		var v any
		if err := json.Unmarshal(data, &v); err != nil {
			t.Fatal(err)
		}
		if err := rs.Validate(v); err != nil {
			t.Fatalf("%s contract: %v\n%s", name, err, data)
		}
	}
	compareExample := func(name string, data []byte, example any) {
		t.Helper()
		b, err := json.Marshal(example)
		if err != nil {
			t.Fatal(err)
		}
		var actual, expected any
		if err := json.Unmarshal(data, &actual); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(b, &expected); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("%s example drift\nactual %s\nexample %s", name, data, b)
		}
	}
	// Empty read responses include null slices: the contract must describe those too.
	for _, op := range apidoc.Operations() {
		if spec.Paths[op.Path][strings.ToLower(op.Method)] == nil {
			t.Fatalf("undocumented %s %s", op.Method, op.Path)
		}
		if op.Method != "GET" || op.Output == "" {
			continue
		}
		path := "/api/v1" + op.Path
		if op.Path == "/plan" {
			path += "?from=2026-10-26&to=2026-11-01"
		}
		w := do(t, h, "GET", path, nil, nil)
		if w.Code != 200 {
			t.Fatalf("%s: %d %s", path, w.Code, w.Body)
		}
		validate(op.Output, w.Body.Bytes())
	}
	// Seed the catalog first: the plan example references source M1. The POST
	// returns the created source; its example must match the real response.
	validate("SaveSource", []byte(apidoc.ExampleSource))
	ws := do(t, h, "POST", "/api/v1/sources", []byte(apidoc.ExampleSource), nil)
	if ws.Code != 200 {
		t.Fatalf("example sources: %d %s", ws.Code, ws.Body)
	}
	validate("SaveSourceResult", ws.Body.Bytes())
	compareExample("SaveSourceResult", ws.Body.Bytes(), apidoc.ResponseExample("SaveSourceResult"))
	for _, tc := range []struct{ path, name, body, out string }{
		{"outline", "SaveOutline", apidoc.ExampleOutline, "OutlineResult"},
		{"plan", "SavePlan", apidoc.ExamplePlan, "PlanResult"},
	} {
		validate(tc.name, []byte(tc.body))
		w := do(t, h, "PUT", "/api/v1/"+tc.path, []byte(tc.body), nil)
		if w.Code != 200 {
			t.Fatalf("example %s: %d %s", tc.path, w.Code, w.Body)
		}
		validate(tc.out, w.Body.Bytes())
		compareExample(tc.out, w.Body.Bytes(), apidoc.ResponseExample(tc.out))
	}
	for _, op := range apidoc.Operations() {
		if op.Method != "GET" || op.Output == "" {
			continue
		}
		path := "/api/v1" + op.Path
		if op.Path == "/plan" {
			path += "?from=2026-10-26&to=2026-10-26"
		}
		w := do(t, h, "GET", path, nil, nil)
		validate(op.Output, w.Body.Bytes())
		compareExample(op.ID, w.Body.Bytes(), apidoc.OperationResponseExample(op))
	}
	for _, method := range []string{"POST", "DELETE"} {
		w := do(t, h, method, "/api/v1/sessions/1/complete", nil, nil)
		if w.Code != 200 {
			t.Fatalf("%s: %s", method, w.Body)
		}
		validate("Session", w.Body.Bytes())
		compareExample(method, w.Body.Bytes(), apidoc.OperationResponseExample(apidoc.Operation{Method: method, Output: "Session"}))
	}
	w := do(t, h, "GET", "/api/v1/agent-guide", nil, nil)
	if !strings.Contains(w.Body.String(), `"today":"2026-10-26"`) {
		t.Fatal(w.Body.String())
	}
	validate("Error", do(t, h, "GET", "/api/v1/plan", nil, nil).Body.Bytes())
	// A non-empty header takes precedence over a conflicting body key.
	body := strings.Replace(apidoc.ExamplePlan, `"from":`, `"idempotency_key":"body-key", "from":`, 1)
	first := do(t, h, "PUT", "/api/v1/plan", []byte(body), map[string]string{"Idempotency-Key": "header-key"})
	if first.Code != 200 {
		t.Fatal(first.Body.String())
	}
	changed := strings.Replace(body, `"duration_min": 90`, `"duration_min": 100`, 1)
	if w := do(t, h, "PUT", "/api/v1/plan", []byte(changed), map[string]string{"Idempotency-Key": "header-key"}); w.Code != 409 {
		t.Fatalf("expected header conflict: %d %s", w.Code, w.Body)
	}
	if w := do(t, h, "PUT", "/api/v1/plan", []byte(changed), nil); w.Code != 200 {
		t.Fatalf("body key should be unused: %d %s", w.Code, w.Body)
	}
}

// TestCatalogRoundTrip exercises the catalog endpoints end to end: an empty read,
// a whole-source write, a read-back with stable ids, a unit-result write, and a
// re-write that preserves the unit id (and so its progress).
func TestCatalogRoundTrip(t *testing.T) {
	h, _ := newAPI(t)

	// A fresh catalog reads empty.
	var empty struct {
		Sources []plan.SourceView `json:"sources"`
	}
	decode(t, do(t, h, "GET", "/api/v1/sources", nil, nil), &empty)
	if len(empty.Sources) != 0 {
		t.Fatalf("fresh catalog not empty: %+v", empty.Sources)
	}

	// POST /sources creates M1.
	var saved plan.SaveSourceResult
	decode(t, do(t, h, "POST", "/api/v1/sources", []byte(apidoc.ExampleSource), nil), &saved)
	if !saved.Created || saved.Source.Key != "M1" {
		t.Fatalf("save source: %+v", saved)
	}

	// GET /sources returns it with stable ids.
	var cat plan.CatalogView
	decode(t, do(t, h, "GET", "/api/v1/sources", nil, nil), &cat)
	if len(cat.Sources) != 1 || len(cat.Sources[0].Sections) != 4 {
		t.Fatalf("catalog read: %+v", cat)
	}
	unitID := cat.Sources[0].Sections[2].Units[0].ID // section 3, first Antrenman

	// POST /units/{id}/result marks it done with doğru/yanlış/boş.
	var marked plan.UnitResultView
	w := do(t, h, "POST", "/api/v1/units/"+itoa(unitID)+"/result", []byte(apidoc.ExampleUnitResult), nil)
	if w.Code != 200 {
		t.Fatalf("unit result: %d %s", w.Code, w.Body)
	}
	decode(t, w, &marked)
	if !marked.Done || marked.UnitID != unitID || marked.Correct == nil || *marked.Correct != 5 {
		t.Fatalf("unit result view: %+v", marked)
	}
	if marked.Notation != "M1·3, Antrenman 1" {
		t.Fatalf("unit notation = %q", marked.Notation)
	}

	// Re-writing the source (changed name, same ordinals) keeps the unit id.
	src := strings.Replace(apidoc.ExampleSource, "Matematik 1", "Matematik 1 (düzeltildi)", 1)
	var resaved plan.SaveSourceResult
	decode(t, do(t, h, "POST", "/api/v1/sources", []byte(src), nil), &resaved)
	if resaved.Created || resaved.Unchanged {
		t.Fatalf("re-save should be an update: %+v", resaved)
	}
	decode(t, do(t, h, "GET", "/api/v1/sources", nil, nil), &cat)
	if got := cat.Sources[0].Sections[2].Units[0].ID; got != unitID {
		t.Fatalf("unit id changed across re-write: %d -> %d", unitID, got)
	}
	// An identical re-write is a no-op.
	decode(t, do(t, h, "POST", "/api/v1/sources", []byte(src), nil), &resaved)
	if !resaved.Unchanged {
		t.Fatalf("identical re-save not unchanged: %+v", resaved)
	}

	// A non-positive unit id is a 422; an unknown unit id is a typed 404.
	if w := do(t, h, "POST", "/api/v1/units/0/result", []byte(apidoc.ExampleUnitResult), nil); w.Code != 422 {
		t.Fatalf("unit id 0: %d", w.Code)
	}
	if w := do(t, h, "POST", "/api/v1/units/99999/result", []byte(apidoc.ExampleUnitResult), nil); w.Code != 404 {
		t.Fatalf("unknown unit: %d %s", w.Code, w.Body)
	}
}

// TestPlanUnknownSource checks that a plan item naming a source with no catalog
// entry is rejected with the service's typed field error, surfaced unchanged by
// the API layer (validation_failed → 422; unknown_source is not not_found).
func TestPlanUnknownSource(t *testing.T) {
	h, _ := newAPI(t)
	seedSources(t, h)
	do(t, h, "PUT", "/api/v1/outline", fixture(t, "sample-outline.json"), nil)

	body := []byte(`{"from":"2026-10-26","to":"2026-10-26","sessions":[{"date":"2026-10-26","start_time":"20:30","duration_min":90,"kind":"aksam","items":[{"source_key":"ZZZ","loc_from":1,"loc_to":10}]}]}`)
	rec := do(t, h, "PUT", "/api/v1/plan", body, nil)
	if rec.Code != 422 {
		t.Fatalf("unknown source: status %d, want 422; body %s", rec.Code, rec.Body)
	}
	var env struct {
		Error struct {
			Code   string `json:"code"`
			Fields []struct {
				Path string `json:"path"`
				Code string `json:"code"`
			} `json:"fields"`
		} `json:"error"`
	}
	decode(t, rec, &env)
	if env.Error.Code != "validation_failed" {
		t.Fatalf("code = %q, want validation_failed", env.Error.Code)
	}
	found := false
	for _, f := range env.Error.Fields {
		if f.Path == "sessions[0].items[0].source_key" && f.Code == plan.CodeUnknownSource {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing unknown_source field: %+v", env.Error.Fields)
	}
}
