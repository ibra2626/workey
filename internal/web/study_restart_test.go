package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ibra2626/workey/internal/clock"
	"github.com/ibra2626/workey/internal/plan"
	"github.com/ibra2626/workey/internal/store"
)

// Exercise the real web handler in separate OS processes against one disposable
// SQLite file. HTTP recording avoids requiring a loopback socket in the sandbox.
// Each step starts from a fresh process; no service, connection or auth instance
// survives the boundary. Synthetic credentials are confined to this test.
type restartRequest struct {
	Path, Body, Cookie, At string
}

type restartResponse struct {
	Status int
	Body   string
	Cookie string
}

func TestStudyRestartWorker(t *testing.T) {
	if os.Getenv("WORKEY_TEST_RESTART_WORKER") != "1" {
		return
	}
	var input restartRequest
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		t.Fatal(err)
	}
	st, err := store.Open(os.Getenv("WORKEY_TEST_RESTART_DB"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	at, err := time.Parse(time.RFC3339, input.At)
	if err != nil {
		t.Fatal(err)
	}
	a := newAuthOn(t, st)
	a.svc = plan.New(st, clock.Fake{Current: at})
	method := http.MethodGet
	if input.Body != "" {
		method = http.MethodPost
	}
	r := httptest.NewRequest(method, input.Path, strings.NewReader(input.Body))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if input.Cookie != "" {
		r.Header.Set("Cookie", input.Cookie)
	}
	res := do(t, a, r)
	defer res.Body.Close()
	out := restartResponse{Status: res.StatusCode, Body: readBody(res)}
	for _, c := range res.Cookies() {
		if c.Name == sessionCookie {
			out.Cookie = c.Name + "=" + c.Value
		}
	}
	if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
		t.Fatal(err)
	}
	// Suppress the test runner's trailing PASS on the JSON channel.
	os.Exit(0)
}

func TestStudyProcessRestartAndIstanbulMidnight(t *testing.T) {
	st := newStore(t)
	seedWeek(t, st)
	a := newAuthOn(t, st)
	id := todaySessionID(t, a)
	var seq int
	var name, dbPath string
	if err := st.DB().QueryRow("PRAGMA database_list").Scan(&seq, &name, &dbPath); err != nil {
		t.Fatal(err)
	}
	if err := st.Close(); err != nil {
		t.Fatal(err)
	}
	at := "2026-10-28T20:59:00Z" // 23:59 Istanbul, despite UTC still being Oct 28
	step := func(path, body, cookie string, want int) restartResponse {
		t.Helper()
		input, err := json.Marshal(restartRequest{path, body, cookie, at})
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(os.Args[0], "-test.run=^TestStudyRestartWorker$")
		cmd.Env = append(os.Environ(), "WORKEY_TEST_RESTART_WORKER=1", "WORKEY_TEST_RESTART_DB="+filepath.Clean(dbPath))
		cmd.Stdin = strings.NewReader(string(input))
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("worker: %v: %s", err, output)
		}
		var out restartResponse
		if err := json.Unmarshal(output, &out); err != nil {
			t.Fatalf("worker output: %v: %s", err, output)
		}
		if out.Status != want {
			t.Fatalf("%s status %d, want %d: %s", path, out.Status, want, out.Body)
		}
		return out
	}
	first := step("/giris", "parola="+url.QueryEscape(testPassword), "", 303).Cookie
	second := step("/giris", "parola="+url.QueryEscape(testPassword), "", 303).Cookie
	if first == "" || second == "" || first == second {
		t.Fatal("independent browser sessions missing")
	}
	csrf := func(cookie string) string {
		sid, ok := a.decodeSession(strings.TrimPrefix(cookie, sessionCookie+"="))
		if !ok {
			t.Fatal("invalid session")
		}
		return "_csrf=" + url.QueryEscape(a.csrfToken(sid)) + "&"
	}
	path := "/oturum/" + strconv.FormatInt(id, 10) + "/"
	step(path+"baslat", csrf(first)+"expected_study_version=0&entry_mode=ten_minutes", first, 303)
	if body := step("/", "", second, 200).Body; !strings.Contains(body, "Başlangıç hedefi: 10 dakika") {
		t.Fatal("active study lost across process/device boundary")
	}
	at = "2026-10-28T21:01:00Z" // 00:01 Oct 29 in Istanbul
	body := step("/", "", second, 200).Body
	for _, text := range []string{"Perşembe, 29 Ekim", "2026-10-28", "Bugünün planı", "M1·4"} {
		if !strings.Contains(body, text) {
			t.Errorf("midnight page missing %q", text)
		}
	}
	step(path+"birak", csrf(first)+"expected_study_version=1&pause_mode=bookmark&item_ordinal=0&stage=soru&resume_page=58", first, 303)
	body = step("/", "", second, 200).Body
	if !strings.Contains(body, "s. 58 sayfasından devam et") || !strings.Contains(body, "M1·4") {
		t.Fatal("paused work or today's choice missing")
	}
	step(path+"baslat", csrf(second)+"expected_study_version=2&entry_mode=normal", second, 303)
	step(path+"birak", csrf(first)+"expected_study_version=1&pause_mode=bookmark&item_ordinal=0&stage=soru&resume_page=57", first, 409)
	step(path+"tamamla", csrf(second)+"expected_study_version=3", second, 303)
	step(path+"geri-al", csrf(second)+"expected_study_version=4", second, 303)
	st, err := store.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	v, err := st.StudyBySession(context.Background(), st.DB(), id)
	if err != nil || v == nil || v.Status != plan.StudyPaused || v.Version != 5 || v.ResumePage == nil || *v.ResumePage != 58 {
		t.Fatalf("final bookmark after restarts/finish/undo: %+v %v", v, err)
	}
}
