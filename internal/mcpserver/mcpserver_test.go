package mcpserver_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ibra2626/workey/internal/clock"
	"github.com/ibra2626/workey/internal/mcpserver"
	"github.com/ibra2626/workey/internal/plan"
	"github.com/ibra2626/workey/internal/store"
)

const testToken = "test-token-123"

// newServer builds the /mcp handler over a migrated temp store whose clock is
// fixed at 2026-10-26 09:00 Istanbul (so the sample week's first day is "today"
// and the rest are future, all writable), and serves it from an httptest
// server. It returns the base URL.
func newServer(t *testing.T) string {
	_, srv := newMCP(t)
	return srv.URL + "/mcp"
}

// newMCP mounts the /mcp transport over a migrated temp store with the fixed test
// clock, web-free. The cross-transport workflow that also mounts internal/web
// lives in webworkflow_test.go (newWorkflow); since WKY-28 it runs by default.
func newMCP(t *testing.T) (*store.Store, *httptest.Server) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "workey.db")
	st, err := store.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	if err := st.Migrate(context.Background()); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	now, err := time.ParseInLocation("2006-01-02 15:04", "2026-10-26 09:00", clock.Location)
	if err != nil {
		t.Fatalf("parse time: %v", err)
	}
	svc := plan.New(st, clock.Fake{Current: now})
	h, err := mcpserver.New(svc, clock.Date{Year: 2027, Month: 7, Day: 18}, testToken)
	if err != nil {
		t.Fatalf("mcpserver.New: %v", err)
	}
	mux := http.NewServeMux()
	mux.Handle("/mcp", h)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return st, srv
}

// authRT adds a Bearer token to every request (none when token is empty).
type authRT struct {
	token string
	base  http.RoundTripper
}

func (a authRT) RoundTrip(r *http.Request) (*http.Response, error) {
	if a.token != "" {
		r.Header.Set("Authorization", "Bearer "+a.token)
	}
	return a.base.RoundTrip(r)
}

// connect opens an MCP client session over Streamable HTTP with the given token.
func connect(t *testing.T, url, token string) *mcp.ClientSession {
	t.Helper()
	tr := &mcp.StreamableClientTransport{
		Endpoint:   url,
		HTTPClient: &http.Client{Transport: authRT{token: token, base: http.DefaultTransport}},
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "v0"}, nil)
	cs, err := client.Connect(context.Background(), tr, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { _ = cs.Close() })
	return cs
}

// callOK calls a tool, fails on a transport error or a tool error, and decodes
// the structured result into v (when v is non-nil).
func callOK(t *testing.T, cs *mcp.ClientSession, name string, args any, v any) {
	t.Helper()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("%s: transport error: %v", name, err)
	}
	if res.IsError {
		t.Fatalf("%s: unexpected tool error: %s", name, errText(res))
	}
	if v != nil {
		buf, err := json.Marshal(res.StructuredContent)
		if err != nil {
			t.Fatalf("%s: marshal structured content: %v", name, err)
		}
		if err := json.Unmarshal(buf, v); err != nil {
			t.Fatalf("%s: decode structured content: %v", name, err)
		}
	}
}

func errText(res *mcp.CallToolResult) string {
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			return tc.Text
		}
	}
	return "(no content)"
}

func loadFixture(t *testing.T, name string) map[string]any {
	t.Helper()
	buf, err := os.ReadFile(filepath.Join("..", "..", "testdata", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	var m map[string]any
	if err := json.Unmarshal(buf, &m); err != nil {
		t.Fatalf("decode fixture %s: %v", name, err)
	}
	delete(m, "_comment") // sample-data annotation; not part of the request schema
	return m
}

// seedSources writes the sample source catalog so the sample week's item
// references (source M1, its sections and units) resolve.
func seedSources(t *testing.T, cs *mcp.ClientSession) {
	t.Helper()
	callOK(t, cs, "save_source", loadFixture(t, "sample-sources.json"), nil)
}

// seedPlan writes the sample source catalog and outline then the sample week, so
// every reference (sources, sections, units, topic groups) resolves.
func seedPlan(t *testing.T, cs *mcp.ClientSession) {
	t.Helper()
	seedSources(t, cs)
	callOK(t, cs, "save_outline", loadFixture(t, "sample-outline.json"), nil)
	callOK(t, cs, "save_plan", loadFixture(t, "sample-week.json"), nil)
}

func TestListTools(t *testing.T) {
	cs := connect(t, newServer(t), testToken)
	res, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	want := map[string]bool{
		"get_agent_guide": false,
		"get_plan":        false, "save_plan": false, "complete_session": false,
		"get_outline": false, "save_outline": false, "get_progress": false,
		"get_sources": false, "save_source": false, "set_unit_result": false,
	}
	for _, tool := range res.Tools {
		if _, ok := want[tool.Name]; !ok {
			t.Errorf("unexpected tool %q", tool.Name)
			continue
		}
		want[tool.Name] = true
		if tool.Description == "" {
			t.Errorf("tool %q has no description", tool.Name)
		}
		if tool.Annotations == nil || tool.Annotations.OpenWorldHint == nil || *tool.Annotations.OpenWorldHint {
			t.Errorf("tool %q needs closed-world annotations", tool.Name)
		}
		readOnly := strings.HasPrefix(tool.Name, "get_")
		if tool.Annotations != nil && (tool.Annotations.ReadOnlyHint != readOnly || !tool.Annotations.IdempotentHint || tool.Annotations.DestructiveHint == nil || *tool.Annotations.DestructiveHint == readOnly) {
			t.Errorf("tool %q incorrect operation hints", tool.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Errorf("missing tool %q", name)
		}
	}
}

func TestAgentGuide(t *testing.T) {
	cs := connect(t, newServer(t), testToken)
	if !strings.Contains(cs.InitializeResult().Instructions, "get_agent_guide") {
		t.Fatal("initialize must direct agents to the guide")
	}
	var guide plan.AgentGuide
	callOK(t, cs, "get_agent_guide", map[string]any{}, &guide)
	if guide.Today != "2026-10-26" || guide.Timezone != "Europe/Istanbul" || guide.ExamDate != "2027-07-18" || len(guide.SourceKinds) != 4 || len(guide.LocationUnits) != 4 || len(guide.Workflow) == 0 {
		t.Fatalf("invalid guide: %+v", guide)
	}
	for _, rule := range guide.Rules {
		for _, banned := range []string{"M1", "M2", "M3", "G1", "G2", "DGS"} {
			if strings.Contains(rule, banned) {
				t.Errorf("guide rule still carries fixed-book/DGS wording %q: %s", banned, rule)
			}
		}
	}
}

// TestSavePlanSchemaSelfDescribing checks that a reader of tools/list alone has
// what they need to build a save_plan call: the key semantics in the
// description and the field constraints (book/kind enums, date/time patterns)
// in the schema. Run with -v to print the schema for the task record.
func TestSavePlanSchemaSelfDescribing(t *testing.T) {
	cs := connect(t, newServer(t), testToken)
	res, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	var tool *mcp.Tool
	for _, tl := range res.Tools {
		if tl.Name == "save_plan" {
			tool = tl
		}
	}
	if tool == nil {
		t.Fatal("save_plan not listed")
	}
	for _, want := range []string{"idempotency", "frozen", "skipped", "makeup_of", "today or later"} {
		if !strings.Contains(tool.Description, want) {
			t.Errorf("save_plan description missing %q", want)
		}
	}
	buf, err := json.MarshalIndent(tool.InputSchema, "", "  ")
	if err != nil {
		t.Fatalf("marshal schema: %v", err)
	}
	schema := string(buf)
	for _, want := range []string{`uzun_blok`, `"pattern"`, `"makeup_of"`, `"source_key"`, `"section_ordinal"`} {
		if !strings.Contains(schema, want) {
			t.Errorf("save_plan input schema missing %q", want)
		}
	}
	// The fixed book-code enum (M1..G2) is gone with the catalog migration.
	for _, banned := range []string{`"M1"`, `"M2"`, `"G1"`} {
		if strings.Contains(schema, banned) {
			t.Errorf("save_plan input schema still carries book-code enum %q", banned)
		}
	}
	t.Logf("save_plan description:\n%s", tool.Description)
	t.Logf("save_plan inputSchema:\n%s", schema)
}

// TestSavePlanIdempotentThenReads exercises the full round trip the task
// requires: save the sample week, save it again (all unchanged), read it back,
// complete a session and read progress.
func TestSavePlanIdempotentThenReads(t *testing.T) {
	cs := connect(t, newServer(t), testToken)
	seedSources(t, cs)
	callOK(t, cs, "save_outline", loadFixture(t, "sample-outline.json"), nil)
	week := loadFixture(t, "sample-week.json")

	var first plan.SavePlanResult
	callOK(t, cs, "save_plan", week, &first)
	if len(first.Created) != 5 {
		t.Fatalf("first save: created = %d, want 5", len(first.Created))
	}
	if len(first.Unchanged) != 0 {
		t.Fatalf("first save: unchanged = %d, want 0", len(first.Unchanged))
	}

	var second plan.SavePlanResult
	callOK(t, cs, "save_plan", week, &second)
	if len(second.Unchanged) != 5 {
		t.Fatalf("second save: unchanged = %d, want 5", len(second.Unchanged))
	}
	if len(second.Created)+len(second.Updated)+len(second.Deleted) != 0 {
		t.Fatalf("second save: changed something: %+v", second)
	}

	var got struct {
		Sessions []plan.SessionView `json:"sessions"`
	}
	callOK(t, cs, "get_plan", map[string]any{"from": "2026-10-26", "to": "2026-11-01"}, &got)
	if len(got.Sessions) != 5 {
		t.Fatalf("get_plan: sessions = %d, want 5", len(got.Sessions))
	}

	// Complete the session dated today (2026-10-26).
	var todayID int64
	for _, s := range got.Sessions {
		if s.Date == "2026-10-26" {
			todayID = s.ID
		}
	}
	if todayID == 0 {
		t.Fatal("get_plan: no session dated 2026-10-26")
	}
	var completed plan.SessionView
	callOK(t, cs, "complete_session", map[string]any{"session_id": todayID, "done": true}, &completed)
	if !completed.Completed || completed.State != plan.StateDone {
		t.Fatalf("complete_session: got completed=%v state=%q", completed.Completed, completed.State)
	}

	var prog plan.ProgressView
	callOK(t, cs, "get_progress", map[string]any{}, &prog)
	if prog.SessionsPlanned != 5 {
		t.Errorf("get_progress: sessions_planned = %d, want 5", prog.SessionsPlanned)
	}
	if prog.SessionsDone != 1 {
		t.Errorf("get_progress: sessions_done = %d, want 1", prog.SessionsDone)
	}
}

// TestCompleteDefaultsToDone checks that omitting "done" completes the session
// (schema default true).
func TestCompleteDefaultsToDone(t *testing.T) {
	cs := connect(t, newServer(t), testToken)
	seedPlan(t, cs)

	var got struct {
		Sessions []plan.SessionView `json:"sessions"`
	}
	callOK(t, cs, "get_plan", map[string]any{"from": "2026-10-26", "to": "2026-10-26"}, &got)
	if len(got.Sessions) != 1 {
		t.Fatalf("get_plan: sessions = %d, want 1", len(got.Sessions))
	}
	var completed plan.SessionView
	callOK(t, cs, "complete_session", map[string]any{"session_id": got.Sessions[0].ID}, &completed)
	if !completed.Completed {
		t.Fatal("complete_session without done did not complete the session")
	}
}

// TestToolErrorCarriesCode checks a future-session completion returns a tool
// error whose JSON carries the same field code as REST (code "future").
func TestToolErrorCarriesCode(t *testing.T) {
	cs := connect(t, newServer(t), testToken)
	seedPlan(t, cs)

	var got struct {
		Sessions []plan.SessionView `json:"sessions"`
	}
	callOK(t, cs, "get_plan", map[string]any{"from": "2026-10-31", "to": "2026-10-31"}, &got)
	if len(got.Sessions) != 1 {
		t.Fatalf("get_plan: sessions = %d, want 1", len(got.Sessions))
	}
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "complete_session",
		Arguments: map[string]any{"session_id": got.Sessions[0].ID, "done": true},
	})
	if err != nil {
		t.Fatalf("transport error: %v", err)
	}
	if !res.IsError {
		t.Fatal("completing a future session did not return a tool error")
	}
	var env struct {
		Code   string `json:"code"`
		Fields []struct {
			Path string `json:"path"`
			Code string `json:"code"`
		} `json:"fields"`
	}
	if err := json.Unmarshal([]byte(errText(res)), &env); err != nil {
		t.Fatalf("error text is not the JSON envelope: %v (%s)", err, errText(res))
	}
	if env.Code != "validation_failed" {
		t.Errorf("error code = %q, want validation_failed", env.Code)
	}
	if len(env.Fields) != 1 || env.Fields[0].Code != plan.CodeFuture {
		t.Errorf("error fields = %+v, want one with code %q", env.Fields, plan.CodeFuture)
	}
}

// TestSaveSourceRoundTripKeepsProgress writes the sample catalog, reads it back,
// marks a unit done with D/Y/B, re-writes the source keeping the unit ordinals
// and confirms the unit id and its progress survive.
func TestSaveSourceRoundTripKeepsProgress(t *testing.T) {
	st, srv := newMCP(t)
	cs := connect(t, srv.URL+"/mcp", testToken)

	var saved plan.SaveSourceResult
	callOK(t, cs, "save_source", loadFixture(t, "sample-sources.json"), &saved)
	if !saved.Created || saved.Source.Key != "M1" {
		t.Fatalf("save_source: %+v", saved)
	}

	var cat plan.CatalogView
	callOK(t, cs, "get_sources", map[string]any{}, &cat)
	unitID := findUnit(t, cat, "M1", 3, 1)

	var marked plan.UnitResultView
	callOK(t, cs, "set_unit_result", map[string]any{"unit_id": unitID, "done": true, "correct": 5, "wrong": 1, "blank": 0}, &marked)
	if !marked.Done || marked.Correct == nil || *marked.Correct != 5 {
		t.Fatalf("set_unit_result: %+v", marked)
	}

	// Re-write the source with a changed section name but the same ordinals; the
	// unit keeps its id, so its progress must survive.
	src := loadFixture(t, "sample-sources.json")
	src["source"].(map[string]any)["name"] = "Matematik 1 (düzeltildi)"
	var resaved plan.SaveSourceResult
	callOK(t, cs, "save_source", src, &resaved)
	if resaved.Created || resaved.Unchanged {
		t.Fatalf("re-save should be an update: %+v", resaved)
	}

	var cat2 plan.CatalogView
	callOK(t, cs, "get_sources", map[string]any{}, &cat2)
	if got := findUnit(t, cat2, "M1", 3, 1); got != unitID {
		t.Fatalf("unit id changed across re-write: %d -> %d", unitID, got)
	}

	var correct int
	if err := st.DB().QueryRow("SELECT correct FROM unit_progress WHERE unit_id = ?", unitID).Scan(&correct); err != nil {
		t.Fatalf("unit progress gone after re-write: %v", err)
	}
	if correct != 5 {
		t.Fatalf("unit progress changed: correct=%d, want 5", correct)
	}

	// An identical re-write is a no-op.
	var again plan.SaveSourceResult
	callOK(t, cs, "save_source", src, &again)
	if !again.Unchanged {
		t.Fatalf("identical save_source not reported unchanged: %+v", again)
	}
}

// TestSavePlanUnknownSource checks that an item naming a source with no catalog
// entry is rejected with the typed unknown_source field error.
func TestSavePlanUnknownSource(t *testing.T) {
	cs := connect(t, newServer(t), testToken)
	callOK(t, cs, "save_outline", loadFixture(t, "sample-outline.json"), nil)
	// No save_source: the plan's items reference source "M1" which does not exist.
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "save_plan",
		Arguments: loadFixture(t, "sample-week.json"),
	})
	if err != nil {
		t.Fatalf("transport error: %v", err)
	}
	if !res.IsError {
		t.Fatal("save_plan with an unknown source did not return a tool error")
	}
	var env struct {
		Code   string `json:"code"`
		Fields []struct {
			Path string `json:"path"`
			Code string `json:"code"`
		} `json:"fields"`
	}
	if err := json.Unmarshal([]byte(errText(res)), &env); err != nil {
		t.Fatalf("error text is not the JSON envelope: %v (%s)", err, errText(res))
	}
	if env.Code != "validation_failed" {
		t.Errorf("error code = %q, want validation_failed", env.Code)
	}
	foundUnknown := false
	for _, f := range env.Fields {
		if f.Code == plan.CodeUnknownSource {
			foundUnknown = true
		}
	}
	if !foundUnknown {
		t.Errorf("error fields = %+v, want one with code %q", env.Fields, plan.CodeUnknownSource)
	}
}

// findUnit returns the id of the unit at (sourceKey, sectionOrdinal, unitOrdinal)
// in a catalog view, failing the test when it is absent.
func findUnit(t *testing.T, cat plan.CatalogView, sourceKey string, sectionOrdinal, unitOrdinal int) int64 {
	t.Helper()
	for _, s := range cat.Sources {
		if s.Key != sourceKey {
			continue
		}
		for _, sec := range s.Sections {
			if sec.Ordinal != sectionOrdinal {
				continue
			}
			for _, u := range sec.Units {
				if u.Ordinal == unitOrdinal {
					return u.ID
				}
			}
		}
	}
	t.Fatalf("unit %s·%d·%d not found in catalog", sourceKey, sectionOrdinal, unitOrdinal)
	return 0
}

// TestUnauthorized checks that connecting without a Bearer token fails (401 on
// the initialize request).
func TestUnauthorized(t *testing.T) {
	url := newServer(t)
	tr := &mcp.StreamableClientTransport{
		Endpoint:   url,
		HTTPClient: &http.Client{Transport: authRT{base: http.DefaultTransport}},
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "v0"}, nil)
	if _, err := client.Connect(context.Background(), tr, nil); err == nil {
		t.Fatal("connect without a token succeeded, want 401 failure")
	}
}
