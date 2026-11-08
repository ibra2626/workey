package web

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/ibra2626/workey/internal/plan"
	"github.com/ibra2626/workey/internal/web/views"
)

// errBadForm marks an unparseable Kaynaklar edit form (missing or non-numeric
// fields), surfaced to the user as a 422 with a generic sentence.
var errBadForm = errors.New("geçersiz form")

// screenSources renders the Kaynaklar screen: the agent-extracted catalog with
// per-source/section progress and an in-place correction form per source.
func (a *Auth) screenSources(w http.ResponseWriter, r *http.Request) {
	p, err := a.buildSourcesPage(r.Context(), csrfTokenFromRequest(r))
	if err != nil {
		http.Error(w, "sunucu hatası", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = views.Sources(p).Render(r.Context(), w)
}

// buildSourcesPage assembles the Kaynaklar view model from the catalog read and
// the per-source/section progress aggregates.
func (a *Auth) buildSourcesPage(ctx context.Context, csrf string) (views.SourcesPage, error) {
	cat, err := a.svc.Sources(ctx)
	if err != nil {
		return views.SourcesPage{}, err
	}
	srcProg, err := a.svc.SourceProgress(ctx)
	if err != nil {
		return views.SourcesPage{}, err
	}
	secProg, err := a.svc.SectionProgress(ctx)
	if err != nil {
		return views.SourcesPage{}, err
	}
	return views.SourcesPage{
		Catalog:        cat,
		SourceProgress: srcProg,
		SectionProg:    secProg,
		CSRF:           csrf,
	}, nil
}

// saveSource applies an in-place edit of one source: it rebuilds the full
// SaveSourceRequest from the submitted structure (every section and unit, with
// their original ordinals preserved so unit progress survives the re-write) and
// calls the plan service as source "web". An htmx request gets the swapped main
// content; a plain form POST redirects back to Kaynaklar.
func (a *Auth) saveSource(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	req, perr := parseSaveSource(r, key)
	var err error
	if perr != nil {
		err = perr
	} else {
		_, err = a.svc.SaveSource(r.Context(), plan.SourceWeb, req)
	}

	if err == nil && r.Header.Get("HX-Request") == "" {
		http.Redirect(w, r, "/kaynaklar", http.StatusSeeOther)
		return
	}

	p, readErr := a.buildSourcesPage(r.Context(), csrfTokenFromRequest(r))
	if readErr != nil {
		http.Error(w, "sunucu hatası", http.StatusInternalServerError)
		return
	}
	status := http.StatusOK
	if err != nil {
		status = http.StatusUnprocessableEntity
		p.Invalid = true
		p.Feedback = sourceErrorMessage(err)
	} else {
		p.Feedback = "“" + key + "” kaydedildi."
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if r.Header.Get("HX-Request") != "" {
		_ = views.SourcesMain(p).Render(r.Context(), w)
	} else {
		_ = views.Sources(p).Render(r.Context(), w)
	}
}

// sourceErrorMessage renders a plan error as a single Turkish sentence for the
// page banner, joining typed validation field messages.
func sourceErrorMessage(err error) string {
	if ve, ok := plan.AsValidationError(err); ok {
		var msgs []string
		for _, f := range ve.Errors {
			msgs = append(msgs, f.Message)
		}
		if len(msgs) > 0 {
			return strings.Join(msgs, " ")
		}
	}
	return "Kaydedilemedi. Girilen değerleri kontrol edin."
}

// parseSaveSource rebuilds a whole-source SaveSourceRequest from the edit form.
// The hidden ordinal lists drive the walk, so an untouched section or unit keeps
// its ordinal (and therefore its id and progress) across the re-write.
func parseSaveSource(r *http.Request, key string) (plan.SaveSourceRequest, error) {
	if err := r.ParseForm(); err != nil {
		return plan.SaveSourceRequest{}, errBadForm
	}
	ordinal, err := strconv.Atoi(strings.TrimSpace(r.PostFormValue("ordinal")))
	if err != nil {
		return plan.SaveSourceRequest{}, errBadForm
	}
	req := plan.SaveSourceRequest{Source: plan.SourceInput{
		Key:          key,
		Name:         strings.TrimSpace(r.PostFormValue("name")),
		Kind:         r.PostFormValue("kind"),
		LocationUnit: r.PostFormValue("location_unit"),
		Ordinal:      ordinal,
	}}
	for _, so := range splitOrdinals(r.PostFormValue("section_ordinals")) {
		sp := "s" + strconv.Itoa(so)
		secLocFrom, err := formOptInt(r.PostFormValue(sp + "_loc_from"))
		if err != nil {
			return plan.SaveSourceRequest{}, errBadForm
		}
		secLocTo, err := formOptInt(r.PostFormValue(sp + "_loc_to"))
		if err != nil {
			return plan.SaveSourceRequest{}, errBadForm
		}
		sec := plan.SectionInput{
			Ordinal: so,
			Name:    strings.TrimSpace(r.PostFormValue(sp + "_name")),
			LocFrom: secLocFrom,
			LocTo:   secLocTo,
		}
		for _, uo := range splitOrdinals(r.PostFormValue(sp + "_unit_ordinals")) {
			up := sp + "u" + strconv.Itoa(uo)
			number, err := formOptInt(r.PostFormValue(up + "_number"))
			if err != nil {
				return plan.SaveSourceRequest{}, errBadForm
			}
			unitLocFrom, err := formOptInt(r.PostFormValue(up + "_loc_from"))
			if err != nil {
				return plan.SaveSourceRequest{}, errBadForm
			}
			unitLocTo, err := formOptInt(r.PostFormValue(up + "_loc_to"))
			if err != nil {
				return plan.SaveSourceRequest{}, errBadForm
			}
			sec.Units = append(sec.Units, plan.UnitInput{
				Ordinal: uo,
				Label:   strings.TrimSpace(r.PostFormValue(up + "_label")),
				Number:  number,
				LocFrom: unitLocFrom,
				LocTo:   unitLocTo,
				Title:   formOptStr(r.PostFormValue(up + "_title")),
				URL:     formOptStr(r.PostFormValue(up + "_url")),
			})
		}
		req.Sections = append(req.Sections, sec)
	}
	return req, nil
}

// splitOrdinals parses a comma-separated ordinal list ("1,2,3"), skipping blanks
// and unparseable entries.
func splitOrdinals(s string) []int {
	var out []int
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if n, err := strconv.Atoi(part); err == nil {
			out = append(out, n)
		}
	}
	return out
}

// formOptStr returns an optional text field; the service treats blank as absent.
func formOptStr(v string) *string {
	return &v
}

// formOptInt parses an optional integer field: empty is nil, a non-numeric value
// is an error the caller surfaces as an invalid form.
func formOptInt(v string) (*int, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return nil, errBadForm
	}
	return &n, nil
}
