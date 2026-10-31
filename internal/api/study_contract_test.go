package api_test

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/ibra2626/workey/internal/apidoc"
	"github.com/ibra2626/workey/internal/clock"
	"github.com/ibra2626/workey/internal/plan"
)

func TestStudyRESTContractAndLegacyWrites(t *testing.T) {
	h, st := newAPI(t)
	ctx := context.Background()
	do(t, h, "PUT", "/api/v1/outline", []byte(apidoc.ExampleOutline), nil)
	do(t, h, "PUT", "/api/v1/plan", []byte(apidoc.ExamplePlan), nil)
	var spec struct {
		Components struct {
			Schemas map[string]*jsonschema.Schema `json:"schemas"`
		} `json:"components"`
	}
	decode(t, do(t, h, "GET", "/api/v1/openapi.json", nil, nil), &spec)
	check := func(name string, data []byte) {
		t.Helper()
		rs, err := spec.Components.Schemas[name].Resolve(nil)
		if err != nil {
			t.Fatal(err)
		}
		var v any
		if err := json.Unmarshal(data, &v); err != nil {
			t.Fatal(err)
		}
		if err := rs.Validate(v); err != nil {
			t.Fatalf("%s: %v\n%s", name, err, data)
		}
	}
	svc := plan.New(st, clock.Fake{Current: time.Date(2026, 10, 26, 20, 30, 0, 0, clock.Location)})
	if _, err := svc.StartStudy(ctx, plan.SourceWeb, 1, plan.EntryTenMinutes, 0, clock.Date{}); err != nil {
		t.Fatal(err)
	}
	check("Day", do(t, h, "GET", "/api/v1/today", nil, nil).Body.Bytes())
	check("Week", do(t, h, "GET", "/api/v1/week", nil, nil).Body.Bytes())
	svc = plan.New(st, clock.Fake{Current: time.Date(2026, 10, 26, 20, 40, 0, 0, clock.Location)})
	if _, err := svc.PauseStudy(ctx, plan.SourceWeb, 1, &plan.Bookmark{ItemOrdinal: 0, Stage: plan.StageSoru, ResumePage: 5}, 1); err != nil {
		t.Fatal(err)
	}
	reread := func() plan.SessionView {
		t.Helper()
		w := do(t, h, "GET", "/api/v1/plan?from=2026-10-26&to=2026-10-26", nil, nil)
		check("Plan", w.Body.Bytes())
		var out struct {
			Sessions []plan.SessionView `json:"sessions"`
		}
		decode(t, w, &out)
		return out.Sessions[0]
	}
	want := apidoc.StudyResponseExample()
	if got := reread(); !reflect.DeepEqual(got, want) {
		t.Fatalf("study example drift: %+v", got)
	}
	changed := strings.Replace(apidoc.ExamplePlan, `"duration_min": 90`, `"duration_min": 100`, 1)
	w := do(t, h, "PUT", "/api/v1/plan", []byte(changed), nil)
	check("PlanResult", w.Body.Bytes())
	var result plan.SavePlanResult
	decode(t, w, &result)
	if !reflect.DeepEqual(result, apidoc.StartedResultExample()) {
		t.Fatalf("skip example drift: %+v", result)
	}
	if got := reread(); !reflect.DeepEqual(got, want) {
		t.Fatal("started work changed")
	}
	// Existing clients send no study version or new fields.
	for _, tc := range []struct {
		method, status string
		version        int64
	}{{"POST", plan.StudyClosed, 3}, {"POST", plan.StudyClosed, 3}, {"DELETE", plan.StudyPaused, 4}, {"DELETE", plan.StudyPaused, 4}} {
		w := do(t, h, tc.method, "/api/v1/sessions/1/complete", nil, nil)
		if w.Code != 200 {
			t.Fatalf("legacy %s: %s", tc.method, w.Body)
		}
		check("Session", w.Body.Bytes())
		var v plan.SessionView
		decode(t, w, &v)
		if v.Study.Status != tc.status || v.Study.Version != tc.version || v.Study.Bookmark.ResumePage != 5 {
			t.Fatalf("completion: %+v", v)
		}
	}
	check("Day", do(t, h, "GET", "/api/v1/today", nil, nil).Body.Bytes())
	check("Week", do(t, h, "GET", "/api/v1/week", nil, nil).Body.Bytes())
	var progress plan.ProgressView
	decode(t, do(t, h, "GET", "/api/v1/progress", nil, nil), &progress)
	if progress.DoneMin != 0 || progress.SessionsDone != 0 {
		t.Fatalf("partial progress: %+v", progress)
	}
}
