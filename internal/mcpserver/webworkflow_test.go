// The cross-transport workflow proves MCP writes and web (form/CSRF) completions
// share one store with correct audit provenance, including the catalog end-to-end
// run (book + video sources, unit results, screens), and drives the opt-in visual
// review. It needs internal/web, which the v2 catalog migration left red until
// WKY-26/27 rebuilt its views; WKY-28 re-enabled these tests by default (they ran
// behind `-tags webworkflow` in the interim). The default MCP test build in
// mcpserver_test stays otherwise web-free (see newMCP).
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

// TestCatalogEndToEnd is the WKY-28 gate: it drives the live /mcp endpoint to
// define two different source kinds (a book counted in pages with Antrenman/Test
// units, and a video counted in lessons with Ders units), plan a week referencing
// both, mark units done (with D/Y/B on the book, none on the video), then reads
// the web screens and MCP progress to prove unit completion and session
// completion stay independent — all with no code change and no fixed book/DGS
// assumption anywhere in the flow.
func TestCatalogEndToEnd(t *testing.T) {
	st, srv := newWorkflow(t, "")
	cs := connect(t, srv.URL+"/mcp", testToken)

	// 1. A book-kind source: pages, sections, Antrenman + Test units.
	book := map[string]any{
		"source": map[string]any{"key": "GR", "name": "İngilizce Gramer", "kind": "book", "location_unit": "page", "ordinal": 1},
		"sections": []any{
			map[string]any{"ordinal": 1, "name": "Tenses", "loc_from": 1, "loc_to": 30, "units": []any{
				map[string]any{"ordinal": 1, "label": "Antrenman", "number": 1, "loc_from": 1, "loc_to": 10},
				map[string]any{"ordinal": 2, "label": "Antrenman", "number": 2, "loc_from": 11, "loc_to": 20},
				map[string]any{"ordinal": 3, "label": "Test", "number": 1, "loc_from": 21, "loc_to": 30},
			}},
		},
	}
	var bookRes plan.SaveSourceResult
	callOK(t, cs, "save_source", book, &bookRes)
	if !bookRes.Created {
		t.Fatalf("book source not created: %+v", bookRes)
	}

	// 2. A video-kind source: lessons, Ders units, no D/Y/B.
	video := map[string]any{
		"source": map[string]any{"key": "VD", "name": "Matematik Video Serisi", "kind": "video", "location_unit": "lesson", "ordinal": 2},
		"sections": []any{
			map[string]any{"ordinal": 1, "name": "Limit", "loc_from": 1, "loc_to": 3, "units": []any{
				map[string]any{"ordinal": 1, "label": "Ders", "number": 1, "loc_from": 1, "loc_to": 1},
				map[string]any{"ordinal": 2, "label": "Ders", "number": 2, "loc_from": 2, "loc_to": 2},
				map[string]any{"ordinal": 3, "label": "Ders", "number": 3, "loc_from": 3, "loc_to": 3},
			}},
		},
	}
	callOK(t, cs, "save_source", video, nil)

	// 3. get_sources: confirm both sources, their kinds/location units, and labels.
	var cat plan.CatalogView
	callOK(t, cs, "get_sources", map[string]any{}, &cat)
	if len(cat.Sources) != 2 {
		t.Fatalf("get_sources returned %d sources, want 2", len(cat.Sources))
	}
	unitID := func(key string, sectionOrdinal, unitOrdinal int) int64 {
		t.Helper()
		for _, src := range cat.Sources {
			if src.Key != key {
				continue
			}
			for _, sec := range src.Sections {
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
		t.Fatalf("unit %s·%d ordinal %d not found", key, sectionOrdinal, unitOrdinal)
		return 0
	}
	for _, src := range cat.Sources {
		switch src.Key {
		case "GR":
			if src.Kind != "book" || src.LocationUnit != "page" || len(src.Sections) != 1 || len(src.Sections[0].Units) != 3 {
				t.Fatalf("GR structure wrong: %+v", src)
			}
			if src.Sections[0].Units[2].Label != "Test" {
				t.Errorf("GR unit 3 label = %q, want Test", src.Sections[0].Units[2].Label)
			}
		case "VD":
			if src.Kind != "video" || src.LocationUnit != "lesson" || len(src.Sections[0].Units) != 3 {
				t.Fatalf("VD structure wrong: %+v", src)
			}
			if src.Sections[0].Units[0].Label != "Ders" {
				t.Errorf("VD unit 1 label = %q, want Ders", src.Sections[0].Units[0].Label)
			}
		default:
			t.Errorf("unexpected source key %q", src.Key)
		}
	}

	// 4. save_plan a week referencing both sources: today's block studies the book,
	// tomorrow's block the video.
	planReq := map[string]any{
		"from": "2026-10-26", "to": "2026-11-01",
		"sessions": []any{
			map[string]any{"date": "2026-10-26", "start_time": "20:30", "duration_min": 90, "kind": "aksam", "items": []any{
				map[string]any{"source_key": "GR", "section_ordinal": 1, "unit_from": 1, "unit_to": 3, "loc_from": 1, "loc_to": 30},
			}},
			map[string]any{"date": "2026-10-27", "start_time": "20:30", "duration_min": 60, "kind": "aksam", "items": []any{
				map[string]any{"source_key": "VD", "section_ordinal": 1, "unit_from": 1, "unit_to": 3, "loc_from": 1, "loc_to": 3},
			}},
		},
	}
	var saved plan.SavePlanResult
	callOK(t, cs, "save_plan", planReq, &saved)
	if len(saved.Created) != 2 || len(saved.Updated)+len(saved.Deleted) != 0 {
		t.Fatalf("save_plan result unexpected: %+v", saved)
	}

	// get_plan echoes both sources' references in the canonical notation.
	var planRead struct {
		Sessions []plan.SessionView `json:"sessions"`
	}
	callOK(t, cs, "get_plan", map[string]any{"from": "2026-10-26", "to": "2026-11-01"}, &planRead)
	var notations strings.Builder
	for _, s := range planRead.Sessions {
		notations.WriteString(s.Notation)
		notations.WriteString(" | ")
	}
	all := notations.String()
	for _, want := range []string{"GR·1", "Antrenman", "VD·1", "Ders"} {
		if !strings.Contains(all, want) {
			t.Errorf("get_plan notation %q missing %q", all, want)
		}
	}

	// 5. set_unit_result: the book's first Antrenman with D/Y/B; the video's first
	// Ders with none.
	var bookUnit plan.UnitResultView
	callOK(t, cs, "set_unit_result", map[string]any{
		"unit_id": unitID("GR", 1, 1), "done": true, "correct": 8, "wrong": 2, "blank": 0,
	}, &bookUnit)
	if !bookUnit.Done || bookUnit.Correct == nil || *bookUnit.Correct != 8 || !strings.Contains(bookUnit.Notation, "GR·1") {
		t.Fatalf("book unit result wrong: %+v", bookUnit)
	}
	var videoUnit plan.UnitResultView
	callOK(t, cs, "set_unit_result", map[string]any{"unit_id": unitID("VD", 1, 1), "done": true}, &videoUnit)
	if !videoUnit.Done || videoUnit.Correct != nil || videoUnit.Wrong != nil || videoUnit.Blank != nil {
		t.Fatalf("video unit must store no D/Y/B: %+v", videoUnit)
	}

	// 6. Read the screens and prove unit vs session completion are independent.
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
	for _, want := range []string{"Birimler", "GR·1, Antrenman 1", "GR·1, Test 1", ">Bitti<", `name="correct"`} {
		if !strings.Contains(today, want) {
			t.Errorf("Bugün missing %q", want)
		}
	}
	week := get("/hafta")
	if !strings.Contains(week, "VD·1, Ders 1") {
		t.Error("Hafta missing the video unit reference")
	}

	unitDone := func() int {
		var n int
		if err := st.DB().QueryRow("SELECT COUNT(*) FROM unit_progress").Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	if n := unitDone(); n != 2 {
		t.Fatalf("unit_progress rows = %d, want 2 (book Antrenman 1 + video Ders 1)", n)
	}

	// Two units are done, but no session has been completed.
	var prog plan.ProgressView
	callOK(t, cs, "get_progress", map[string]any{}, &prog)
	if prog.SessionsDone != 0 {
		t.Fatalf("marking units wrongly completed a session: SessionsDone=%d", prog.SessionsDone)
	}

	// Complete today's session through the real web form (CSRF), then confirm the
	// session counts as done while the unit progress is untouched.
	csrf := regexp.MustCompile(`name="_csrf" value="([^"]+)"`).FindStringSubmatch(today)
	action := regexp.MustCompile(`action="(/oturum/[0-9]+/tamamla)"`).FindStringSubmatch(today)
	if len(csrf) != 2 || len(action) != 2 {
		t.Fatal("missing completion form on Bugün")
	}
	res, err := http.PostForm(srv.URL+action[1], url.Values{"_csrf": {csrf[1]}})
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("complete POST: %d", res.StatusCode)
	}
	callOK(t, cs, "get_progress", map[string]any{}, &prog)
	if prog.SessionsDone != 1 {
		t.Fatalf("web completion did not reach MCP progress: SessionsDone=%d", prog.SessionsDone)
	}
	if n := unitDone(); n != 2 {
		t.Fatalf("session completion changed unit progress: unit_progress rows = %d, want 2", n)
	}

	// İlerleme shows the honest unit aggregate: 2 done out of all 6 catalog units.
	progress := get("/ilerleme")
	for _, want := range []string{"Birim ilerlemesi", "6 birim", "Çözülen"} {
		if !strings.Contains(progress, want) {
			t.Errorf("İlerleme missing %q", want)
		}
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
