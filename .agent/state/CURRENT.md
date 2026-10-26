# CURRENT — WKY-03 done, SQLite store + schema live

Updated 2026-10-26.

Active task: .agent/work/workey-v1/tasks/WKY-03.md (done; nothing committed)

- Feature: `workey-v1` — Go + templ + htmx + Tailwind single binary; Giriş / Bugün / Hafta / İlerleme; REST `/api/v1` and MCP `/mcp` over one `internal/plan` service; SQLite (modernc). Local Windows only; WKY-12 (deploy) blocked by the user.
- Plan: `.agent/work/workey-v1/TASKS.md`; shared invariants + layout in `.agent/work/workey-v1/README.md` (read with each task). Module path `github.com/ibra2626/workey`, go1.27.3.

## WKY-03 result (done)
- **Package `internal/store`** on `modernc.org/sqlite` v1.61.0 (pure Go, no CGO; database/sql driver name `sqlite`). Now a direct dep in go.mod (`go mod tidy` run).
- **`store.go`:** `Open(path)` builds a DSN with pragmas `busy_timeout(5000)`, `foreign_keys(1)`, `journal_mode(WAL)` and sets `SetMaxOpenConns(1)` (single-writer pattern; pings on open). `Close()`, `DB()` (read-only queries). `WithTx(ctx, func(tx dbtx) error)` commits on success, rolls back on error/panic. `Migrate(ctx)` ensures `schema_migrations`, then applies each embedded `migrations/*.sql` not yet recorded (filename order, one tx each, records filename) — idempotent. `dbtx` interface (Exec/Query/QueryRow-Context) is satisfied by `*sql.DB` and `*sql.Tx`.
- **`migrations/0001_init.sql`** (`//go:embed migrations/*.sql`): `phases`, `topic_groups`, `sessions`, `session_items`, `audit_log`, `idempotency_keys`. Dates/times/timestamps are TEXT. CHECKs: `sessions.kind IN (aksam,uzun_blok,deneme)`, `completed_source` null-or-(mcp|api|web), `session_items.book IN (M1,M2,M3,G1,G2)`, `page_from<=page_to`, optional konu/soru/antrenman/phase/dgs range ordering, `audit_log.source IN (mcp,api,web)`. FKs: `session_items.session_id → sessions(id) ON DELETE CASCADE`, `sessions.makeup_of → sessions(id) ON DELETE SET NULL`, `sessions.topic_group_id → topic_groups(id)`. `UNIQUE(date,start_time)`, `UNIQUE(session_id,ordinal)`; indexes on `sessions(date)` and `session_items(session_id)`.
- **`queries.go`:** structs `Session` (+`Items`), `SessionItem`, `AuditEntry`, `IdempotencyRecord` (string dates, `*int`/`*int64`/`*string` for nullables). Helpers, all taking `dbtx`: `InsertSession`, `InsertSessionItem` (set `.ID`), `SessionsByDateRange(from,to)` (sessions in `[from,to]` ordered by date,start_time, items loaded via a second JOIN query and attached), `InsertAudit`, `GetIdempotency` (`(nil,nil)` when absent), `PutIdempotency` (plain INSERT → PK rejects duplicate; callers read-before-write). No business rules here.
- **`store_test.go`** (temp-dir DB via `t.TempDir()`): migrate idempotent + all six tables present; `PRAGMA foreign_keys=1` + orphan-item FK violation; date-range query returns the two in-range sessions in order with items (soru range round-trips); cascade delete of items; `makeup_of` SET NULL on parent delete; CHECK failures (bad kind, page_from>page_to, bad audit source); UNIQUE(date,start_time); audit insert round trip; idempotency round trip + duplicate-key rejection.
- **Checks all green:** `go test ./internal/store/...` ok; `go vet ./...` clean; `CGO_ENABLED=0 go build ./...` ok; `gofmt -l .` empty. CLAUDE.md "Checks" updated (added `CGO_ENABLED=0` build note and `internal/store` to the test list). DECISIONS entry added (modernc rationale, single-writer + WAL, embedded migration approach, string date storage).
- Nothing committed (commit only on explicit request). New untracked: `internal/store/`; modified: `go.mod`, `go.sum`, `CLAUDE.md`, `.agent/...`.

## Next step
WKY-04 (next in TASKS.md) — likely `internal/plan` domain types, validation, notation formatter, and the service that uses `internal/store`. Reuse `store.WithTx` + the query helpers; keep all rules (validation, freezing, idempotency orchestration, `M2·17, s. 210–225` notation) in `internal/plan`, not in the store. New session → read TASKS.md index then `tasks/WKY-04.md`.
