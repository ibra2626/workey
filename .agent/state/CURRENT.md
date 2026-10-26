# CURRENT — WKY-04 done, plan write path live

Updated 2026-10-26.

Active task: .agent/work/workey-v1/tasks/WKY-04.md (done; nothing committed)

- Feature: `workey-v1` — Go + templ + htmx + Tailwind single binary; Giriş / Bugün / Hafta / İlerleme; REST `/api/v1` and MCP `/mcp` over one `internal/plan` service; SQLite (modernc). Local Windows only; WKY-12 (deploy) blocked by the user.
- Plan: `.agent/work/workey-v1/TASKS.md`; shared invariants + layout in `.agent/work/workey-v1/README.md` (read with each task). Module path `github.com/ibra2626/workey`, go1.27.3.
- Done so far: WKY-01 (skeleton/config/clock/server), WKY-02 (frontend toolchain + base layout), WKY-03 (store + schema), WKY-04 (plan write path). Index statuses for 02/03/04 set to done.

## WKY-04 result (done)
- **`internal/plan`** is the single business-logic layer (web/REST/MCP will only parse → call → render). Files:
  - `plan.go` — domain types: `Source` (mcp|api|web), `Item`, `SlotRef`, `SessionInput`, `SavePlanRequest`/`Result`, `SaveOutlineRequest`/`Result`, enum sets (kinds, books).
  - `notation.go` — `FormatItem` → `M2·17, s. 210–225` (middle dot U+00B7, en dash U+2013); optional `, Antr. 1–3`; single values collapse the range; `FormatItems` joins items with ` · `.
  - `errors.go` — `FieldError{path,code,message}`, `ValidationError`, `AsValidationError`, `ErrIdempotencyConflict`, codes (`required|invalid|range_order|out_of_range|duplicate|not_found|makeup_target|referenced`), `errBuilder`.
  - `validate.go` — store-free structural validation of SavePlan/SaveOutline.
  - `service.go` — `Service{store, clock}`, `New`, `SavePlan`, audit/idempotency/snapshot/equality helpers, result sorting.
  - `outline.go` — `SaveOutline`.
- **`SavePlan(ctx, source, req)`**: declarative range replace over `[from,to]`. Today-or-later open (uncompleted) sessions become exactly the request's set, matched by `(date,start_time)`: match→update-in-place (id kept) or unchanged; absent open-future→delete; frozen (date<today or completed) never touched, incoming targeting them → `Skipped` (`past_date`|`completed`). Idempotent: no key → repeat all-unchanged, 0 audit (per-session content equality); with `idempotency_key` → request hash + stored response in `idempotency_keys`, repeat replays response, different payload same key → `ErrIdempotencyConflict`. `makeup_of` = `SlotRef` resolved to missed session id (must exist, be past, uncompleted). Returns created/updated/deleted/unchanged/skipped sorted by (date,start_time). One `store.WithTx`; audit row per create/update/delete (source + before/after JSON + request hash). Item update = delete+re-insert with ordinals 0..n.
- **`SaveOutline(ctx, source, req)`**: phases + topic groups replaced by key (create/update/unchanged/delete), idempotent, audited; deleting a topic group still referenced by sessions → `CodeReferenced`, whole tx rolls back.
- **Store changes (WKY-03 layer):** `store.WithTx` callback now takes `*sql.Tx` (was unexported `dbtx`) so other packages compose helpers in a tx; query helpers still accept `dbtx` (so `*sql.DB` works read-only); `store_test.go` updated (imports `database/sql`). New helpers: `catalog.go` (`Phase`, `TopicGroup` + list/insert/UpdateByKey/DeleteByKey, `TopicGroupByKey`, `CountSessionsByTopicGroup`); `queries.go` + `UpdateSession`, `DeleteSession`, `DeleteSessionItems`, `SessionByDateAndTime`.
- **Fixtures (`testdata/`, labelled sample data via `_comment`):** `sample-week.json` (week 1 26 Eki–1 Kas: Mon/Tue/Thu evenings + Wed with konu 41–44/soru 45–58/antr 1–3 + Sat uzun_blok 2 items) and `sample-outline.json` (Faz 1 · İşlem temeli + 5 topic groups).
- **Tests green:** `notation_test.go`, `validate_test.go`, `service_test.go` cover create+idempotent (fixtures), change-one, past-date-skip, completed-untouched, shrink-deletes-open-future, makeup valid/unknown/future, idempotency-key replay+conflict, audit before/after, outline replace + referenced-delete. Checks all pass: `go test ./internal/plan/... ./internal/store/... ./internal/config/... ./internal/clock/... ./cmd/...`, `go vet ./...`, `gofmt -l .` empty, `CGO_ENABLED=0 go build ./...`.
- Nothing committed. New untracked: `internal/plan/`, `internal/store/catalog.go`, `testdata/`. Modified: `internal/store/store.go`, `internal/store/queries.go`, `internal/store/store_test.go`, `.agent/...`. DECISIONS entry added (WKY-04).

## Next step
WKY-05 (read models and completion) — depends on WKY-04. Build the read side (Bugün/Hafta/İlerleme view models) and completion/undo (mark done, audit, frozen rules) on top of `internal/plan` + store. New session → read TASKS.md index then `tasks/WKY-05.md`; reuse `plan.Service`, `store` query helpers, and the notation formatter. Nothing is committed yet; a release session would stage `internal/plan/`, the store changes, `testdata/`, and the `.agent/` docs.
