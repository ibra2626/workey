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

	"github.com/ibra2626/workey/internal/apidoc"

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
	mux.Handle("GET /kullanim", a.requireAuth(http.HandlerFunc(a.screenUsage)))
	mux.Handle("GET /kullanim/mcp", a.requireAuth(http.HandlerFunc(a.screenUsage)))
	mux.Handle("GET /kullanim/api", a.requireAuth(http.HandlerFunc(a.screenUsage)))
	mux.Handle("GET /kullanim/api/openapi.json", a.requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Disposition", `attachment; filename="workey-openapi.json"`)
		apidoc.ServeHTTP(w, r)
	})))
	mux.Handle("GET /{$}", a.requireAuth(http.HandlerFunc(a.screenToday)))
	mux.Handle("GET /hafta", a.requireAuth(http.HandlerFunc(a.screenWeek)))
	mux.Handle("GET /ilerleme", a.requireAuth(http.HandlerFunc(a.screenProgress)))
	mux.Handle("POST /oturum/{id}/tamamla", a.requireAuth(a.requireCSRF(http.HandlerFunc(a.completeSession))))
	mux.Handle("POST /oturum/{id}/geri-al", a.requireAuth(a.requireCSRF(http.HandlerFunc(a.undoSession))))
	mux.Handle("POST /hafta/oturum/{id}/tamamla", a.requireAuth(a.requireCSRF(http.HandlerFunc(a.weekComplete))))
	mux.Handle("POST /hafta/oturum/{id}/geri-al", a.requireAuth(a.requireCSRF(http.HandlerFunc(a.weekUndo))))
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

// screenProgress renders the İlerleme screen.
func (a *Auth) screenProgress(w http.ResponseWriter, r *http.Request) {
	prog, err := a.svc.Progress(r.Context(), a.examDate)
	if err != nil {
		http.Error(w, "sunucu hatası", http.StatusInternalServerError)
		return
	}
	p := views.ProgressPage{
		Progress:   prog,
		DaysToExam: prog.DaysToExam,
		HasExam:    a.examDate != (clock.Date{}),
		CSRF:       csrfTokenFromRequest(r),
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = views.Progress(p).Render(r.Context(), w)
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
// gets the session and refreshed summaries; a plain POST redirects to Bugün.
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
	_ = views.TodayMain(p).Render(r.Context(), w)
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

// screenWeek renders the Hafta screen for the week containing ?tarih= (default
// today).
func (a *Auth) screenWeek(w http.ResponseWriter, r *http.Request) {
	date := a.weekDateParam(r)
	p, err := a.buildWeekPage(r.Context(), date, csrfTokenFromRequest(r))
	if err != nil {
		http.Error(w, "sunucu hatası", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = views.Week(p).Render(r.Context(), w)
}

// weekComplete / weekUndo mark a session from the Hafta screen (source web). An
// htmx request gets the swapped main content (figures + row); a plain form POST
// redirects back to that week.
func (a *Auth) weekComplete(w http.ResponseWriter, r *http.Request) { a.weekSetDone(w, r, true) }
func (a *Auth) weekUndo(w http.ResponseWriter, r *http.Request)     { a.weekSetDone(w, r, false) }

func (a *Auth) weekSetDone(w http.ResponseWriter, r *http.Request, done bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "geçersiz oturum", http.StatusBadRequest)
		return
	}
	var sv *plan.SessionView
	if done {
		sv, err = a.svc.Complete(r.Context(), plan.SourceWeb, id)
	} else {
		sv, err = a.svc.Undo(r.Context(), plan.SourceWeb, id)
	}
	if err != nil {
		// A future-session or not-found completion is a client error.
		http.Error(w, "işlem tamamlanamadı", http.StatusUnprocessableEntity)
		return
	}
	date, ok := plan.ParseDate(sv.Date)
	if !ok {
		http.Error(w, "sunucu hatası", http.StatusInternalServerError)
		return
	}
	p, err := a.buildWeekPage(r.Context(), date, csrfTokenFromRequest(r))
	if err != nil {
		http.Error(w, "sunucu hatası", http.StatusInternalServerError)
		return
	}
	if r.Header.Get("HX-Request") == "" {
		http.Redirect(w, r, "/hafta?tarih="+p.Week.Start, http.StatusSeeOther)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = views.WeekMain(p).Render(r.Context(), w)
}

// weekDateParam reads ?tarih=YYYY-MM-DD, falling back to today when it is absent
// or unparseable.
func (a *Auth) weekDateParam(r *http.Request) clock.Date {
	if v := r.URL.Query().Get("tarih"); v != "" {
		if d, ok := plan.ParseDate(v); ok {
			return d
		}
	}
	return a.svc.TodayDate()
}

// buildWeekPage assembles the Hafta view model for the week containing date: the
// week read model, the exam countdown for the rail, and the previous/next week
// ?tarih= links.
func (a *Auth) buildWeekPage(ctx context.Context, date clock.Date, csrf string) (views.WeekPage, error) {
	wk, err := a.svc.Week(ctx, date)
	if err != nil {
		return views.WeekPage{}, err
	}
	prog, err := a.svc.Progress(ctx, a.examDate)
	if err != nil {
		return views.WeekPage{}, err
	}
	p := views.WeekPage{
		Week:       wk,
		DaysToExam: prog.DaysToExam,
		HasExam:    a.examDate != (clock.Date{}),
		CSRF:       csrf,
	}
	if start, ok := plan.ParseDate(wk.Start); ok {
		p.PrevDate = shiftDays(start, -7).String()
		p.NextDate = shiftDays(start, 7).String()
	}
	return p, nil
}

// shiftDays returns date moved by n days (zone-free civil arithmetic; Turkey has
// no DST since 2016).
func shiftDays(date clock.Date, n int) clock.Date {
	t := time.Date(date.Year, date.Month, date.Day, 0, 0, 0, 0, time.UTC).AddDate(0, 0, n)
	return clock.Date{Year: t.Year(), Month: t.Month(), Day: t.Day()}
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
