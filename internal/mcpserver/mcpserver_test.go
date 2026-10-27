package mcpserver_test

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ibra2626/workey/internal/clock"
	"github.com/ibra2626/workey/internal/mcpserver"
	"github.com/ibra2626/workey/internal/plan"
	"github.com/ibra2626/workey/internal/store"
	web "github.com/ibra2626/workey/internal/web"
)

const testToken = "test-token-123"

// newServer builds the /mcp handler over a migrated temp store whose clock is
// fixed at 2026-10-26 09:00 Istanbul (so the sample week's first day is "today"
// and the rest are future, all writable), and serves it from an httptest
// server. It returns the base URL.
func newServer(t *testing.T) string {
	_, srv := newWorkflow(t, "")
	return srv.URL + "/mcp"
}

// newWorkflow mounts both transports over one store. A fixed address is only
// used by the opt-in visual review; normal tests use an ephemeral port.
func newWorkflow(t *testing.T, addr string) (*store.Store, *httptest.Server) {
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
	a, err := web.NewAuth(svc, clock.Date{Year: 2027, Month: 7, Day: 18}, "", "", false, true)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.Handle("/mcp", h)
	mux.Handle("/", a.Handler())
	srv := httptest.NewUnstartedServer(mux)
	if addr != "" {
		_ = srv.Listener.Close()
		srv.Listener, err = net.Listen("tcp", addr)
		if err != nil {
			t.Fatal(err)
		}
	}
	srv.Start()
	t.Cleanup(srv.Close)
	return st, srv
}

// TestMCPWebWorkflow proves the cross-transport contract, including the real
// form/CSRF path and audit provenance rather than completing through MCP only.
func TestMCPWebWorkflow(t *testing.T) {
	st, srv := newWorkflow(t, "")
	cs := connect(t, srv.URL+"/mcp", testToken)
	seedPlan(t, cs)
	get := func(path string) string {
		t.Helper()
		res, err := http.Get(srv.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		b, err := io.ReadAll(res.Body)
		if err != nil || res.StatusCode != http.StatusOK {
			t.Fatalf("GET %s: status=%d err=%v", path, res.StatusCode, err)
		}
		return string(b)
	}
	today := get("/")
	if !strings.Contains(today, "Temel kavramlar") || !strings.Contains(today, "s. 1–22") {
		t.Fatal("MCP plan missing from Bugün")
	}
	week := get("/hafta")
	for _, title := range []string{"Sayı basamakları", "Üslü ifadeler", "Köklü ifadeler", "Dinlenme"} {
		if !strings.Contains(week, title) {
			t.Fatalf("Hafta missing %q", title)
		}
	}
	var before plan.ProgressView
	callOK(t, cs, "get_progress", map[string]any{}, &before)
	if before.SessionsDone != 0 || before.SessionsPlanned != 5 || before.PlannedMin != 540 {
		t.Fatalf("before: %+v", before)
	}
	csrf := regexp.MustCompile(`name="_csrf" value="([^"]+)"`).FindStringSubmatch(today)
	action := regexp.MustCompile(`action="(/oturum/[0-9]+/tamamla)"`).FindStringSubmatch(today)
	if len(csrf) != 2 || len(action) != 2 {
		t.Fatal("missing completion form")
	}
	post := func(path string) {
		t.Helper()
		res, err := http.PostForm(srv.URL+path, url.Values{"_csrf": {csrf[1]}})
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("POST %s: %d", path, res.StatusCode)
		}
	}
	post(action[1])
	var after plan.ProgressView
	callOK(t, cs, "get_progress", map[string]any{}, &after)
	if after.SessionsDone != 1 || after.DoneMin != 90 || after.DoneHours != "1,5" {
		t.Fatalf("after: %+v", after)
	}
	if !strings.Contains(get("/ilerleme"), "1,5 / 9 sa") {
		t.Fatal("web progress disagrees with MCP")
	}
	if !strings.Contains(get("/"), "Geri al") {
		t.Fatal("completion not visible on Bugün")
	}
	countAudit := func() int {
		var n int
		if err := st.DB().QueryRow("SELECT COUNT(*) FROM audit_log").Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	n := countAudit()
	var repeat plan.SavePlanResult
	callOK(t, cs, "save_plan", loadFixture(t, "sample-week.json"), &repeat)
	if len(repeat.Unchanged) != 4 || len(repeat.Skipped) != 1 || len(repeat.Created)+len(repeat.Updated)+len(repeat.Deleted) != 0 {
		t.Fatalf("repeat modified completed plan: %+v", repeat)
	}
	post(action[1]) // repeated completion is also a no-op
	if countAudit() != n {
		t.Fatal("repeat wrote duplicate audit entries")
	}
	var sessions int
	if err := st.DB().QueryRow("SELECT COUNT(*) FROM sessions").Scan(&sessions); err != nil || sessions != 5 {
		t.Fatalf("sessions=%d err=%v", sessions, err)
	}
	rows, err := st.DB().Query("SELECT at_utc, source, action, COUNT(*) FROM audit_log GROUP BY at_utc, source, action ORDER BY source, action")
	if err != nil {
		t.Fatal(err)
	}
	seenMCP, seenWeb := false, false
	for rows.Next() {
		var at, source, action string
		var count int
		if err := rows.Scan(&at, &source, &action, &count); err != nil {
			t.Fatal(err)
		}
		t.Logf("audit: %s | %s | %s | %d", at, source, action, count)
		seenMCP = seenMCP || source == "mcp"
		seenWeb = seenWeb || source == "web" && action == "complete" && count == 1
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	_ = rows.Close()
	if !seenMCP || !seenWeb {
		t.Fatal("missing audit source")
	}
	post(strings.Replace(action[1], "/tamamla", "/geri-al", 1))
	callOK(t, cs, "get_progress", map[string]any{}, &after)
	if after.SessionsDone != 0 || after.DoneMin != 0 {
		t.Fatal("web undo did not reach MCP progress")
	}
}

// TestVisualReview is an opt-in, disposable local preview. No real database or
// Claude configuration is touched. Stop it with Ctrl-C after taking evidence.
func TestVisualReview(t *testing.T) {
	if os.Getenv("WORKEY_VISUAL_REVIEW") != "1" {
		t.Skip("set WORKEY_VISUAL_REVIEW=1 for disposable browser fixtures")
	}
	st, srv := newWorkflow(t, "127.0.0.1:18113")
	cs := connect(t, srv.URL+"/mcp", testToken)
	seedPlan(t, cs)
	// Monday moved to Saturday; Tuesday remains missed on the Wednesday view.
	week := loadFixture(t, "sample-week.json")
	sessions := week["sessions"].([]any)
	sessions[4].(map[string]any)["makeup_of"] = map[string]any{"date": "2026-10-26", "start_time": "20:30"}
	buf, err := json.Marshal(week)
	if err != nil {
		t.Fatal(err)
	}
	var req plan.SavePlanRequest
	if err := json.Unmarshal(buf, &req); err != nil {
		t.Fatal(err)
	}
	wednesday := plan.New(st, clock.Fake{Current: time.Date(2026, 10, 28, 9, 0, 0, 0, clock.Location)})
	if _, err := wednesday.SavePlan(context.Background(), plan.SourceMCP, req); err != nil {
		t.Fatal(err)
	}
	for i, scenario := range []struct {
		day    int
		seeded bool
		login  bool
	}{{28, true, false}, {30, true, false}, {28, false, false}, {28, false, true}, {31, true, false}} {
		scenarioStore := st
		if !scenario.seeded {
			var err error
			scenarioStore, err = store.Open(filepath.Join(t.TempDir(), "empty.db"))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = scenarioStore.Close() })
			if err := scenarioStore.Migrate(context.Background()); err != nil {
				t.Fatal(err)
			}
		}
		svc := plan.New(scenarioStore, clock.Fake{Current: time.Date(2026, 10, scenario.day, 9, 0, 0, 0, clock.Location)})
		a, err := web.NewAuth(svc, clock.Date{Year: 2027, Month: 7, Day: 18}, "", "", false, !scenario.login)
		if err != nil {
			t.Fatal(err)
		}
		s := httptest.NewUnstartedServer(a.Handler())
		_ = s.Listener.Close()
		s.Listener, err = net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(18114+i))
		if err != nil {
			t.Fatal(err)
		}
		s.Start()
		t.Cleanup(s.Close)
		t.Logf("review day=%d seeded=%v login=%v: %s", scenario.day, scenario.seeded, scenario.login, s.URL)
	}
	t.Log("SAMPLE DATA ONLY; review ready for 30 minutes")
	select {
	case <-time.After(30 * time.Minute):
	}
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

// seedPlan saves the sample outline then the sample week, so the week's
// topic-group references resolve.
func seedPlan(t *testing.T, cs *mcp.ClientSession) {
	t.Helper()
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
	if guide.Today != "2026-10-26" || guide.Timezone != "Europe/Istanbul" || guide.ExamDate != "2027-07-18" || len(guide.Books) != 5 || len(guide.Workflow) == 0 {
		t.Fatalf("invalid guide: %+v", guide)
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
	for _, want := range []string{`"M1"`, `"uzun_blok"`, `"pattern"`, `"makeup_of"`, `"enum"`} {
		if !strings.Contains(schema, want) {
			t.Errorf("save_plan input schema missing %q", want)
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
