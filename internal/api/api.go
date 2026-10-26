// Package api is workey's REST interface at /api/v1. Handlers are thin: each
// parses the request, calls internal/plan with source "api" and renders the
// service's typed result or error. No business rule lives here — validation,
// freezing, idempotency, notation and state all come from the plan service.
package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/ibra2626/workey/internal/clock"
	"github.com/ibra2626/workey/internal/httpx"
	"github.com/ibra2626/workey/internal/plan"
)

// maxBodyBytes caps a request body. Plan writes are small (a week of sessions);
// 1 MiB is generous and bounds memory.
const maxBodyBytes = 1 << 20

// Handler serves /api/v1 over a plan service. examDate feeds the progress
// countdown (from WORKEY_EXAM_DATE).
type Handler struct {
	svc      *plan.Service
	examDate clock.Date
}

// New builds the fully wrapped /api/v1 handler: the routes behind security
// headers, a body-size limit and a Bearer-token guard (apiToken). Mount it at
// "/api/v1/" on the server mux.
func New(svc *plan.Service, examDate clock.Date, apiToken string) http.Handler {
	h := &Handler{svc: svc, examDate: examDate}
	return httpx.SecurityHeaders(httpx.LimitBody(maxBodyBytes, httpx.BearerAuth(apiToken, h.routes())))
}

// routes registers the method+pattern routes (Go 1.22+ routing).
func (h *Handler) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/plan", h.getPlan)
	mux.HandleFunc("PUT /api/v1/plan", h.putPlan)
	mux.HandleFunc("GET /api/v1/outline", h.getOutline)
	mux.HandleFunc("PUT /api/v1/outline", h.putOutline)
	mux.HandleFunc("GET /api/v1/progress", h.getProgress)
	mux.HandleFunc("GET /api/v1/today", h.getToday)
	mux.HandleFunc("GET /api/v1/week", h.getWeek)
	mux.HandleFunc("POST /api/v1/sessions/{id}/complete", h.completeSession)
	mux.HandleFunc("DELETE /api/v1/sessions/{id}/complete", h.undoSession)
	return mux
}

// GET /api/v1/plan?from=YYYY-MM-DD&to=YYYY-MM-DD
func (h *Handler) getPlan(w http.ResponseWriter, r *http.Request) {
	from, okFrom := plan.ParseDate(r.URL.Query().Get("from"))
	to, okTo := plan.ParseDate(r.URL.Query().Get("to"))
	var fields []httpx.FieldDetail
	if !okFrom {
		fields = append(fields, httpx.FieldDetail{Path: "from", Code: plan.CodeInvalid, Message: "from must be a date in YYYY-MM-DD form"})
	}
	if !okTo {
		fields = append(fields, httpx.FieldDetail{Path: "to", Code: plan.CodeInvalid, Message: "to must be a date in YYYY-MM-DD form"})
	}
	if len(fields) > 0 {
		httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.Error{Code: "validation_failed", Message: "invalid query parameters", Fields: fields})
		return
	}
	sessions, err := h.svc.Range(r.Context(), from, to)
	if err != nil {
		writePlanError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"sessions": sessions})
}

// PUT /api/v1/plan  (body: plan.SavePlanRequest; optional Idempotency-Key header)
func (h *Handler) putPlan(w http.ResponseWriter, r *http.Request) {
	var req plan.SavePlanRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if key := r.Header.Get("Idempotency-Key"); key != "" {
		req.IdempotencyKey = key
	}
	res, err := h.svc.SavePlan(r.Context(), plan.SourceAPI, req)
	if err != nil {
		writePlanError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, res)
}

// GET /api/v1/outline
func (h *Handler) getOutline(w http.ResponseWriter, r *http.Request) {
	out, err := h.svc.Outline(r.Context())
	if err != nil {
		writePlanError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

// PUT /api/v1/outline  (body: plan.SaveOutlineRequest; optional Idempotency-Key)
func (h *Handler) putOutline(w http.ResponseWriter, r *http.Request) {
	var req plan.SaveOutlineRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if key := r.Header.Get("Idempotency-Key"); key != "" {
		req.IdempotencyKey = key
	}
	res, err := h.svc.SaveOutline(r.Context(), plan.SourceAPI, req)
	if err != nil {
		writePlanError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, res)
}

// GET /api/v1/progress
func (h *Handler) getProgress(w http.ResponseWriter, r *http.Request) {
	prog, err := h.svc.Progress(r.Context(), h.examDate)
	if err != nil {
		writePlanError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, prog)
}

// GET /api/v1/today
func (h *Handler) getToday(w http.ResponseWriter, r *http.Request) {
	day, err := h.svc.Today(r.Context())
	if err != nil {
		writePlanError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, day)
}

// GET /api/v1/week?date=YYYY-MM-DD  (date optional; defaults to today)
func (h *Handler) getWeek(w http.ResponseWriter, r *http.Request) {
	date := h.svc.TodayDate()
	if q := r.URL.Query().Get("date"); q != "" {
		d, ok := plan.ParseDate(q)
		if !ok {
			httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.Error{
				Code: "validation_failed", Message: "invalid query parameters",
				Fields: []httpx.FieldDetail{{Path: "date", Code: plan.CodeInvalid, Message: "date must be a date in YYYY-MM-DD form"}},
			})
			return
		}
		date = d
	}
	week, err := h.svc.Week(r.Context(), date)
	if err != nil {
		writePlanError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, week)
}

// POST /api/v1/sessions/{id}/complete
func (h *Handler) completeSession(w http.ResponseWriter, r *http.Request) {
	id, ok := sessionID(w, r)
	if !ok {
		return
	}
	view, err := h.svc.Complete(r.Context(), plan.SourceAPI, id)
	if err != nil {
		writePlanError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, view)
}

// DELETE /api/v1/sessions/{id}/complete  (undo)
func (h *Handler) undoSession(w http.ResponseWriter, r *http.Request) {
	id, ok := sessionID(w, r)
	if !ok {
		return
	}
	view, err := h.svc.Undo(r.Context(), plan.SourceAPI, id)
	if err != nil {
		writePlanError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, view)
}

// --- helpers (parsing and error mapping only) ---

// sessionID parses the {id} path value, writing a 422 and returning false when
// it is not a positive integer.
func sessionID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.Error{
			Code: "validation_failed", Message: "invalid session id",
			Fields: []httpx.FieldDetail{{Path: "id", Code: plan.CodeInvalid, Message: "id must be a positive integer"}},
		})
		return 0, false
	}
	return id, true
}

// decodeJSON reads the request body into dst, writing a 400 and returning false
// on malformed JSON or an over-limit body.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.Error{
			Code: "bad_request", Message: "request body is not valid JSON",
		})
		return false
	}
	return true
}

// writePlanError maps a plan-service error to an HTTP response:
//   - ValidationError → 422, or 404 when every field is not_found;
//   - ErrIdempotencyConflict → 409;
//   - anything else → 500 with a generic message (the real error is not leaked).
func writePlanError(w http.ResponseWriter, err error) {
	if ve, ok := plan.AsValidationError(err); ok {
		fields := make([]httpx.FieldDetail, 0, len(ve.Errors))
		allNotFound := len(ve.Errors) > 0
		for _, fe := range ve.Errors {
			fields = append(fields, httpx.FieldDetail{Path: fe.Path, Code: fe.Code, Message: fe.Message})
			if fe.Code != plan.CodeNotFound {
				allNotFound = false
			}
		}
		status, code := http.StatusUnprocessableEntity, "validation_failed"
		if allNotFound {
			status, code = http.StatusNotFound, "not_found"
		}
		httpx.WriteError(w, status, httpx.Error{Code: code, Message: "the request did not pass validation", Fields: fields})
		return
	}
	if errors.Is(err, plan.ErrIdempotencyConflict) {
		httpx.WriteError(w, http.StatusConflict, httpx.Error{
			Code: "idempotency_conflict", Message: "this idempotency key was already used with a different request",
		})
		return
	}
	httpx.WriteError(w, http.StatusInternalServerError, httpx.Error{Code: "internal", Message: "internal error"})
}
