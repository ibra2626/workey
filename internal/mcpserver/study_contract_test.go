package mcpserver_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/ibra2626/workey/internal/clock"
	"github.com/ibra2626/workey/internal/plan"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestStudyMCPOutputContractAndLegacyWrites(t *testing.T) {
	ctx := context.Background()
	st, srv := newWorkflow(t, "")
	cs := connect(t, srv.URL+"/mcp", testToken)
	seedPlan(t, cs)
	tools, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	schemas := map[string]*jsonschema.Resolved{}
	for _, tool := range tools.Tools {
		if tool.Name != "get_plan" && tool.Name != "complete_session" && tool.Name != "save_plan" {
			continue
		}
		b, err := json.Marshal(tool.OutputSchema)
		if err != nil {
			t.Fatal(err)
		}
		var s jsonschema.Schema
		if err := json.Unmarshal(b, &s); err != nil {
			t.Fatal(err)
		}
		r, err := s.Resolve(nil)
		if err != nil {
			t.Fatal(err)
		}
		schemas[tool.Name] = r
		if tool.Name == "get_plan" {
			for _, field := range []string{"study", "bookmark", "resume_page", "item_ordinal", "notation", "ten_minutes", "yanlis_defteri"} {
				if !strings.Contains(string(b), field) {
					t.Fatalf("output schema missing %s", field)
				}
			}
		}
	}
	call := func(name string, args any, out any) {
		t.Helper()
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
		if err != nil {
			t.Fatal(err)
		}
		if res.IsError {
			t.Fatal(errText(res))
		}
		b, err := json.Marshal(res.StructuredContent)
		if err != nil {
			t.Fatal(err)
		}
		var raw any
		if err := json.Unmarshal(b, &raw); err != nil {
			t.Fatal(err)
		}
		if err := schemas[name].Validate(raw); err != nil {
			t.Fatalf("%s schema: %v\n%s", name, err, b)
		}
		if out != nil {
			if err := json.Unmarshal(b, out); err != nil {
				t.Fatal(err)
			}
		}
	}
	args := map[string]any{"from": "2026-10-26", "to": "2026-11-01"}
	var out struct {
		Sessions []plan.SessionView `json:"sessions"`
	}
	call("get_plan", args, &out)
	if out.Sessions[0].Study != nil {
		t.Fatal("legacy study present")
	}
	id := out.Sessions[0].ID
	svc := plan.New(st, clock.Fake{Current: time.Date(2026, 10, 26, 20, 30, 0, 0, clock.Location)})
	if _, err := svc.StartStudy(ctx, plan.SourceWeb, id, plan.EntryTenMinutes, 0, clock.Date{}); err != nil {
		t.Fatal(err)
	}
	call("get_plan", args, &out)
	if out.Sessions[0].Study.Status != plan.StudyActive {
		t.Fatal("active missing")
	}
	if _, err := svc.PauseStudy(ctx, plan.SourceWeb, id, &plan.Bookmark{ItemOrdinal: 0, Stage: plan.StageSoru, ResumePage: 5}, 1); err != nil {
		t.Fatal(err)
	}
	call("get_plan", args, &out)
	if out.Sessions[0].Study.Bookmark.ResumePage != 5 {
		t.Fatal("bookmark missing")
	}
	var week plan.SavePlanRequest
	b, _ := json.Marshal(loadFixture(t, "sample-week.json"))
	if err := json.Unmarshal(b, &week); err != nil {
		t.Fatal(err)
	}
	week.Sessions[0].DurationMin = 100
	var result plan.SavePlanResult
	call("save_plan", week, &result)
	if len(result.Skipped) != 1 || result.Skipped[0].Reason != plan.SkipStarted {
		t.Fatalf("started: %+v", result)
	}
	call("get_plan", args, &out)
	if out.Sessions[0].DurationMin != 90 || out.Sessions[0].Study.Version != 2 {
		t.Fatal("started changed")
	}
	for _, tc := range []struct {
		done    bool
		status  string
		version int64
	}{{true, plan.StudyClosed, 3}, {true, plan.StudyClosed, 3}, {false, plan.StudyPaused, 4}} {
		var v plan.SessionView
		call("complete_session", map[string]any{"session_id": id, "done": tc.done}, &v)
		if v.Study.Status != tc.status || v.Study.Version != tc.version || v.Study.Bookmark.ResumePage != 5 {
			t.Fatalf("legacy completion: %+v", v)
		}
	}
	var progress plan.ProgressView
	callOK(t, cs, "get_progress", map[string]any{}, &progress)
	if progress.DoneMin != 0 || progress.SessionsDone != 0 {
		t.Fatal("partial progress counted")
	}
}
