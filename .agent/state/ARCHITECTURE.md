# Architecture map

Bootstrapped 2026-10-26; updated 2026-11-04 to the implemented v2 source-catalog
model (WKY-21..28). A navigation map, not a full audit.

- Entry point: `cmd/workey/main.go` — loads `internal/config`, opens and migrates
  the store, builds the `internal/plan` service with an Istanbul clock, then a mux
  (`newMux`) mounting `GET /healthz`, the REST API at `/api/v1/`, MCP at `/mcp`,
  and the web UI at `/`.
- Modules / ownership (`internal/`):
  - `config` — env + optional `.env`; requires API token / password hash /
    session key unless `WORKEY_DEV=1`. Never logs values.
  - `clock` — injectable Europe/Istanbul clock; tests never use the wall clock.
  - `store` — SQLite (`modernc.org/sqlite`, pure Go) + embedded migrations
    (`migrations/0001_init.sql` full v2 schema, `0002_study_flow.sql`). Tables:
    `sources`, `sections`, `units`, `unit_progress`, `sessions`, `session_items`
    (catalog references only), `phases`, `topic_groups`, `study`, `audit_log`,
    `idempotency_keys`. CRUD only — no business rules.
  - `plan` — **the single business-logic layer**: catalog domain types, validation,
    the source+unit notation formatter (`notation.go`), the plan/outline/source/
    unit-result service path, read models for the screens (`read.go`, `units.go`,
    `sources.go`), the agent guide (`guide.go`). Writes are idempotent,
    frozen-past-aware and audited.
  - `web` (+ `web/views`, templ) — Bugün / Hafta / İlerleme, the Kaynaklar
    correction page, per-unit marking with optional D/Y/B, usage pages. Password +
    session cookie + CSRF; thin over the service.
  - `api` — REST `/api/v1` (plan, sources, progress, …); `apidoc` — OpenAPI
    contract + examples. Bearer token.
  - `mcpserver` — MCP Streamable HTTP at `/mcp` (official Go SDK): `get_agent_guide`,
    `get_plan`/`save_plan`, `complete_session`, `get_sources`/`save_source`,
    `set_unit_result`, `get_outline`/`save_outline`, `get_progress`. Bearer token.
  - `httpx` — shared security headers and Bearer auth middleware.
- Data and persistence: a source catalog (sources → sections → units) with unit
  progress stored separately so re-writing a source (`save_source`) never wipes
  progress. Session items store only references (source key + section ordinal +
  unit/location ranges), never topic text or question content.
- Security boundaries: single user. Web auth is password + signed session cookie +
  CSRF; REST and MCP share one Bearer-token guard. Writes are audited (time UTC,
  source mcp|api|web, action, entity id, before/after JSON). Secrets only from the
  environment.
- Verification: `internal/mcpserver/webworkflow_test.go` drives the live `/mcp`
  endpoint plus the web UI over one store, including the WKY-28 catalog end-to-end
  gate (`TestCatalogEndToEnd`). See `docs/v2-gate-verification.md`.
