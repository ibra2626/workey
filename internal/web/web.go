// Package web holds the server-rendered HTTP handlers. Handlers only parse the
// request, call the plan service and render templ views; no business rule lives
// here (see .agent/work/workey-v1/README.md invariants).
//
// All UI pages require a single-user session cookie; the Giriş screen and the
// static assets are the only public routes. CSRF is enforced on every
// state-changing request, both by a per-session token (htmx hx-headers) and by
// net/http cross-origin protection across the whole web mux. The REST API and
// MCP server authenticate with a Bearer token and are not mounted here.
package web

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/ibra2626/workey/internal/clock"
	"github.com/ibra2626/workey/internal/plan"
	"github.com/ibra2626/workey/internal/web/views"
	assets "github.com/ibra2626/workey/web"
)

// Handler builds the web UI handler: the login and static routes are public,
// every other page sits behind requireAuth, and the whole tree is wrapped in
// net/http cross-origin protection so a cross-site form or fetch cannot drive a
// state change. Mount it at "/" on the parent mux; more specific parent
// patterns (/api/v1/, /mcp, /healthz) still take precedence.
func (a *Auth) Handler() http.Handler {
	mux := http.NewServeMux()

	// Public: static assets and the login screen.
	mux.Handle("GET /static/", assets.StaticHandler())
	mux.HandleFunc("GET /giris", a.loginForm)
	mux.HandleFunc("POST /giris", a.loginSubmit)

	// Protected pages.
	mux.Handle("GET /{$}", a.requireAuth(http.HandlerFunc(a.screenToday)))
	mux.Handle("GET /hafta", a.requireAuth(page("hafta", "Workey · Hafta", "Hafta")))
	mux.Handle("GET /ilerleme", a.requireAuth(page("ilerleme", "Workey · İlerleme", "İlerleme")))
	mux.Handle("POST /oturum/{id}/tamamla", a.requireAuth(a.requireCSRF(http.HandlerFunc(a.completeSession))))
	mux.Handle("POST /oturum/{id}/geri-al", a.requireAuth(a.requireCSRF(http.HandlerFunc(a.undoSession))))
	mux.Handle("POST /cikis", a.requireAuth(a.requireCSRF(http.HandlerFunc(a.logout))))

	cop := http.NewCrossOriginProtection()
	return cop.Handler(mux)
}

// loginForm renders the Giriş screen. An already-authenticated visitor (or any
// visitor in dev-open mode) is sent straight to Bugün.
func (a *Auth) loginForm(w http.ResponseWriter, r *http.Request) {
	if a.devOpen || a.authenticated(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	renderLogin(w, r, "")
}

// loginSubmit verifies the password, applies login throttling and, on success,
// issues the session cookie and redirects to Bugün.
func (a *Auth) loginSubmit(w http.ResponseWriter, r *http.Request) {
	if a.devOpen {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if wait := a.throttle.retryAfter(); wait > 0 {
		w.Header().Set("Retry-After", strconv.Itoa(int(wait.Round(time.Second)/time.Second)))
		w.WriteHeader(http.StatusTooManyRequests)
		renderLogin(w, r, "Çok fazla deneme yapıldı. Biraz bekleyip tekrar deneyin.")
		return
	}
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		renderLogin(w, r, "İstek okunamadı.")
		return
	}
	ok, err := VerifyPassword(a.passwordHash, r.PostFormValue("parola"))
	if err != nil || !ok {
		a.throttle.fail()
		w.WriteHeader(http.StatusUnauthorized)
		renderLogin(w, r, "Parola hatalı.")
		return
	}
	a.throttle.success()
	a.issueSession(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// logout clears the session cookie and returns to the Giriş screen.
func (a *Auth) logout(w http.ResponseWriter, r *http.Request) {
	a.clearSession(w)
	http.Redirect(w, r, "/giris", http.StatusSeeOther)
}

// authenticated reports whether the request carries a valid session cookie.
func (a *Auth) authenticated(r *http.Request) bool {
	c, err := r.Cookie(sessionCookie)
	if err != nil {
		return false
	}
	_, ok := a.decodeSession(c.Value)
	return ok
}

// renderLogin writes the Giriş screen with an optional error sentence. The
// status code must already be set by the caller.
func renderLogin(w http.ResponseWriter, r *http.Request, errMsg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = views.Login(errMsg).Render(r.Context(), w)
}

// page renders a placeholder view for one section. The CSRF token is read from
// the request's session and passed to the layout. Replaced by real screens in
// WKY-10 (Hafta) and WKY-11 (İlerleme).
func page(active, title, heading string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		csrf := csrfTokenFromRequest(r)
		_ = views.Placeholder(active, title, heading, csrf).Render(r.Context(), w)
	}
}

// screenToday renders the Bugün screen for the clock's current date.
func (a *Auth) screenToday(w http.ResponseWriter, r *http.Request) {
	p, err := a.buildTodayPage(r.Context(), csrfTokenFromRequest(r))
	if err != nil {
		http.Error(w, "sunucu hatası", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = views.Today(p).Render(r.Context(), w)
}

// completeSession marks tonight's session done (source web). An htmx request
// gets the swapped first panel; a plain form POST is redirected back to Bugün.
func (a *Auth) completeSession(w http.ResponseWriter, r *http.Request) {
	a.setSessionDone(w, r, true)
}

// undoSession clears a session's completion under the same swap/redirect rules.
func (a *Auth) undoSession(w http.ResponseWriter, r *http.Request) {
	a.setSessionDone(w, r, false)
}

func (a *Auth) setSessionDone(w http.ResponseWriter, r *http.Request, done bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "geçersiz oturum", http.StatusBadRequest)
		return
	}
	if done {
		_, err = a.svc.Complete(r.Context(), plan.SourceWeb, id)
	} else {
		_, err = a.svc.Undo(r.Context(), plan.SourceWeb, id)
	}
	if err != nil {
		// A future-session or not-found completion is a client error; any other
		// failure is a server error. Both render without leaking internals.
		http.Error(w, "işlem tamamlanamadı", http.StatusUnprocessableEntity)
		return
	}
	if r.Header.Get("HX-Request") == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	p, err := a.buildTodayPage(r.Context(), csrfTokenFromRequest(r))
	if err != nil {
		http.Error(w, "sunucu hatası", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = views.TodayPanel(p).Render(r.Context(), w)
}

// buildTodayPage assembles the Bugün view model from the plan read models: the
// day itself, this week's figures and the exam countdown. The service has
// already derived every state and notation string.
func (a *Auth) buildTodayPage(ctx context.Context, csrf string) (views.TodayPage, error) {
	today := a.svc.TodayDate()
	day, err := a.svc.Day(ctx, today)
	if err != nil {
		return views.TodayPage{}, err
	}
	wk, err := a.svc.Week(ctx, today)
	if err != nil {
		return views.TodayPage{}, err
	}
	prog, err := a.svc.Progress(ctx, a.examDate)
	if err != nil {
		return views.TodayPage{}, err
	}
	p := views.TodayPage{
		Day:            day,
		Week:           wk,
		HasPlan:        prog.HasPlan,
		WeekNumber:     wk.WeekNumber,
		PhaseName:      wk.PhaseName,
		DaysToExam:     prog.DaysToExam,
		AfterExam:      prog.AfterExam,
		HasExam:        a.examDate != (clock.Date{}),
		WeekDoneMin:    wk.DoneMin,
		WeekPlannedMin: wk.PlannedMin,
		SessionsDone:   wk.SessionsDone,
		SessionsPlan:   wk.SessionsPlanned,
		CSRF:           csrf,
	}
	p.PhaseLabel, p.PhaseBlocks = phaseBlocks(prog)
	return p, nil
}

// phaseBlocks builds the desktop current-phase week blocks from the progress
// strip: the weeks that fall in the current phase, each showing its done
// fraction (filled for the current week), with the current week sub-labelled by
// its minutes.
func phaseBlocks(prog *plan.ProgressView) (label string, blocks []views.PhaseBlock) {
	if prog.CurrentPhase == "" {
		return "", nil
	}
	for _, c := range prog.Strip {
		if c.PhaseName != prog.CurrentPhase {
			continue
		}
		b := views.PhaseBlock{
			Label:     "H" + strconv.Itoa(c.WeekNumber),
			IsCurrent: c.WeekNumber == prog.WeekNumber,
			FillPct:   int(c.DoneFraction*100 + 0.5),
		}
		if b.IsCurrent {
			b.Label += " · bu hafta"
			b.Sub = strconv.Itoa(c.DoneMin) + " / " + strconv.Itoa(c.PlannedMin) + " dk"
		}
		blocks = append(blocks, b)
	}
	return prog.CurrentPhase, blocks
}
