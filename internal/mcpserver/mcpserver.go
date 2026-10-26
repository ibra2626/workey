// Package mcpserver exposes the plan service to an AI client as MCP tools over
// Streamable HTTP at /mcp, using the official Go SDK
// (github.com/modelcontextprotocol/go-sdk). Like internal/api it is a thin
// transport: each tool parses its typed arguments, calls internal/plan with
// source "mcp" and returns the service's result or maps its error to a tool
// error. No business rule lives here.
//
// All tool and field descriptions are in English and state that times are
// Europe/Istanbul and that only book/chapter/page references are stored, never
// topic text or question content, so an AI client can call the tools correctly
// from tools/list alone.
package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ibra2626/workey/internal/clock"
	"github.com/ibra2626/workey/internal/httpx"
	"github.com/ibra2626/workey/internal/plan"
)

// serverName and serverVersion identify this MCP server to clients.
const (
	serverName    = "workey"
	serverVersion = "v1"
)

// handler holds the dependencies every tool needs: the single business-logic
// service and the exam date used by get_progress.
type handler struct {
	svc      *plan.Service
	examDate clock.Date
}

// planResult is the get_plan output: the sessions in the requested range.
type planResult struct {
	Sessions []plan.SessionView `json:"sessions"`
}

// New builds the fully wrapped /mcp handler: the MCP Streamable HTTP endpoint
// behind security headers and the same Bearer-token guard as the REST API
// (mcpToken). Mount it at "/mcp" on the server mux.
func New(svc *plan.Service, examDate clock.Date, mcpToken string) (http.Handler, error) {
	srv, err := newMCPServer(&handler{svc: svc, examDate: examDate})
	if err != nil {
		return nil, err
	}
	stream := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return srv }, nil)
	return httpx.SecurityHeaders(httpx.BearerAuth(mcpToken, stream)), nil
}

// newMCPServer registers the six tools on a new MCP server. It returns an error
// (rather than panicking) if any input schema fails to build, so startup can
// fail cleanly.
func newMCPServer(h *handler) (*mcp.Server, error) {
	srv := mcp.NewServer(&mcp.Implementation{Name: serverName, Version: serverVersion}, nil)

	savePlanIn, err := savePlanSchema()
	if err != nil {
		return nil, fmt.Errorf("mcpserver: save_plan schema: %w", err)
	}
	saveOutlineIn, err := saveOutlineSchema()
	if err != nil {
		return nil, fmt.Errorf("mcpserver: save_outline schema: %w", err)
	}
	getPlanIn, err := getPlanSchema()
	if err != nil {
		return nil, fmt.Errorf("mcpserver: get_plan schema: %w", err)
	}
	completeIn, err := completeSchema()
	if err != nil {
		return nil, fmt.Errorf("mcpserver: complete_session schema: %w", err)
	}

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_plan",
		Description: "Return the planned sessions for a date range [from, to] (inclusive), Europe/Istanbul. Each session has its stored items (book/chapter/page references only), its derived state (planned, today, done, missed, moved), the resolved topic-group name and the canonical notation (e.g. 'M2·17, s. 210–225'). Use the returned 'id' with complete_session.",
		InputSchema: getPlanIn,
	}, h.getPlan)

	mcp.AddTool(srv, &mcp.Tool{
		Name: "save_plan",
		Description: "Declaratively replace the open sessions for every date in [from, to] (Europe/Istanbul). For each date in range that is today or later, the open (uncompleted) sessions are made to match exactly the given set, keyed by date + start_time: a matching session is updated in place, a new one is created, and an open future session omitted from the request is deleted. " +
			"Past days and completed sessions are frozen — they are never changed or deleted, and any incoming session that targets one is returned in 'skipped' rather than applied. " +
			"The call is safe to repeat: sending the same request twice leaves the same state and reports the sessions as 'unchanged' (pass idempotency_key to also replay the exact first result). " +
			"Link a makeup session to the missed one with 'makeup_of' (the missed session's date + start_time). Only book/chapter/page references are stored, never topic text or question content.",
		InputSchema: savePlanIn,
	}, h.savePlan)

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "complete_session",
		Description: "Mark a session done (done=true, the default) or undo its completion (done=false), addressed by the 'id' from get_plan. Allowed for today and past sessions only; a future session is rejected. Idempotent: setting the state it is already in changes nothing.",
		InputSchema: completeIn,
	}, h.completeSession)

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_outline",
		Description: "Return the plan outline: the phases (with their date spans) and the topic groups (with their keys and optional DGS question ranges), ordered. Topic-group keys here are the ones a session's topic_group refers to.",
	}, h.getOutline)

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "save_outline",
		Description: "Declaratively replace the plan's phases and topic groups, each matched by key (create / update / delete). A phase or group present by key is created or updated; a key absent from the request is deleted. A topic group still referenced by sessions cannot be deleted. Idempotent, with the same idempotency_key replay semantics as save_plan.",
		InputSchema: saveOutlineIn,
	}, h.saveOutline)

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_progress",
		Description: "Return the progress summary (Europe/Istanbul): current plan week out of the fixed 38, sessions done vs planned, done vs planned minutes and hours, current phase, the exam countdown, per-topic-group completion and the 38-week strip.",
	}, h.getProgress)

	return srv, nil
}

// --- tool handlers (parse + call + map; no business rules) ---

// getPlan backs get_plan.
func (h *handler) getPlan(ctx context.Context, _ *mcp.CallToolRequest, in getPlanInput) (*mcp.CallToolResult, planResult, error) {
	from, okFrom := plan.ParseDate(in.From)
	to, okTo := plan.ParseDate(in.To)
	var b errBuilder
	if !okFrom {
		b.add("from", plan.CodeInvalid, "from must be a date in YYYY-MM-DD form")
	}
	if !okTo {
		b.add("to", plan.CodeInvalid, "to must be a date in YYYY-MM-DD form")
	}
	if err := b.err(); err != nil {
		return nil, planResult{}, err
	}
	sessions, err := h.svc.Range(ctx, from, to)
	if err != nil {
		return nil, planResult{}, mapError(err)
	}
	return nil, planResult{Sessions: sessions}, nil
}

// savePlan backs save_plan.
func (h *handler) savePlan(ctx context.Context, _ *mcp.CallToolRequest, in plan.SavePlanRequest) (*mcp.CallToolResult, *plan.SavePlanResult, error) {
	res, err := h.svc.SavePlan(ctx, plan.SourceMCP, in)
	if err != nil {
		return nil, nil, mapError(err)
	}
	return nil, res, nil
}

// completeSession backs complete_session.
func (h *handler) completeSession(ctx context.Context, _ *mcp.CallToolRequest, in completeInput) (*mcp.CallToolResult, *plan.SessionView, error) {
	var (
		view *plan.SessionView
		err  error
	)
	if in.Done {
		view, err = h.svc.Complete(ctx, plan.SourceMCP, in.SessionID)
	} else {
		view, err = h.svc.Undo(ctx, plan.SourceMCP, in.SessionID)
	}
	if err != nil {
		return nil, nil, mapError(err)
	}
	return nil, view, nil
}

// getOutline backs get_outline.
func (h *handler) getOutline(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, *plan.OutlineView, error) {
	out, err := h.svc.Outline(ctx)
	if err != nil {
		return nil, nil, mapError(err)
	}
	return nil, out, nil
}

// saveOutline backs save_outline.
func (h *handler) saveOutline(ctx context.Context, _ *mcp.CallToolRequest, in plan.SaveOutlineRequest) (*mcp.CallToolResult, *plan.SaveOutlineResult, error) {
	res, err := h.svc.SaveOutline(ctx, plan.SourceMCP, in)
	if err != nil {
		return nil, nil, mapError(err)
	}
	return nil, res, nil
}

// getProgress backs get_progress.
func (h *handler) getProgress(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, *plan.ProgressView, error) {
	prog, err := h.svc.Progress(ctx, h.examDate)
	if err != nil {
		return nil, nil, mapError(err)
	}
	return nil, prog, nil
}

// --- error mapping (same codes and field paths as the REST API) ---

// toolError carries a structured error envelope. Its Error() string is the JSON
// {"code","message","fields"} that the SDK embeds in the tool result's text
// content with IsError=true, so an MCP client reads the same machine codes and
// field paths as a REST client.
type toolError struct {
	env httpx.Error
}

func (e *toolError) Error() string {
	buf, err := json.Marshal(e.env)
	if err != nil {
		return e.env.Message
	}
	return string(buf)
}

// mapError converts a plan-service error into a tool error, mirroring the REST
// handler: a ValidationError becomes validation_failed, or not_found when every
// field is not_found; an idempotency conflict becomes idempotency_conflict;
// anything else becomes a generic internal error (the real error is not leaked).
func mapError(err error) error {
	if ve, ok := plan.AsValidationError(err); ok {
		fields := make([]httpx.FieldDetail, 0, len(ve.Errors))
		allNotFound := len(ve.Errors) > 0
		for _, fe := range ve.Errors {
			fields = append(fields, httpx.FieldDetail{Path: fe.Path, Code: fe.Code, Message: fe.Message})
			if fe.Code != plan.CodeNotFound {
				allNotFound = false
			}
		}
		code := "validation_failed"
		if allNotFound {
			code = "not_found"
		}
		return &toolError{env: httpx.Error{Code: code, Message: "the request did not pass validation", Fields: fields}}
	}
	if errors.Is(err, plan.ErrIdempotencyConflict) {
		return &toolError{env: httpx.Error{Code: "idempotency_conflict", Message: "this idempotency key was already used with a different request"}}
	}
	return &toolError{env: httpx.Error{Code: "internal", Message: "internal error"}}
}

// errBuilder accumulates argument-parsing field errors into a validation
// tool error, matching the REST field-error shape.
type errBuilder struct {
	fields []httpx.FieldDetail
}

func (b *errBuilder) add(path, code, message string) {
	b.fields = append(b.fields, httpx.FieldDetail{Path: path, Code: code, Message: message})
}

func (b *errBuilder) err() error {
	if len(b.fields) == 0 {
		return nil
	}
	return &toolError{env: httpx.Error{Code: "validation_failed", Message: "invalid arguments", Fields: b.fields}}
}
