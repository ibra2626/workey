package api_test

import (
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/ibra2626/workey/internal/apidoc"
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
