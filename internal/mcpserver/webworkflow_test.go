//go:build webworkflow

// The cross-transport workflow proves MCP writes and web (form/CSRF) completions
// share one store with correct audit provenance, and drives the opt-in visual
// review. It needs internal/web, which the v2 catalog migration leaves red until
// WKY-26/27 rebuild its views; until then build and run these with
// `go test -tags webworkflow ./internal/mcpserver/...` (WKY-28 re-enables them by
// default). The default MCP test build stays web-free (see newMCP).
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

	"github.com/ibra2626/workey/internal/clock"
	"github.com/ibra2626/workey/internal/mcpserver"
	"github.com/ibra2626/workey/internal/plan"
	"github.com/ibra2626/workey/internal/store"
	web "github.com/ibra2626/workey/internal/web"
)

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
