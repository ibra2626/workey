# WKY-04 — Plan service: write path (save plan, outline, audit)

## Goal
The single business-logic entry for writing the plan: validated, idempotent, frozen-past-aware, audited.

## Scope
- `internal/plan`: domain types (Session, Item, Phase, TopicGroup, Source), validation, notation formatter (`M2·17, s. 210–225`; multi-item join; antrenman `Antr. 1–3`), service struct taking store + clock.
- `SavePlan(ctx, source, req)` with `req = {from, to, sessions[], idempotency_key?}`:
  - Declarative range replace: for dates in [from, to] that are today or later, the open (uncompleted) sessions become exactly the given set, matched by (date, start_time). Unchanged sessions keep their id; changed ones update in place; missing ones are deleted.
  - Frozen: sessions before today (Istanbul) and completed sessions are never modified or deleted; incoming sessions that target them are returned in `skipped[]` with a reason.
  - `makeup_of` must reference an existing missed session (past, uncompleted).
  - Idempotent: the same payload twice gives identical state and zero new audit rows on the second call. With `idempotency_key` the stored response is replayed; a different payload under the same key is rejected.
  - Returns `{created, updated, deleted, unchanged, skipped}` with ids.
- `SaveOutline(ctx, source, {phases[], topic_groups[]})`: replace by key, idempotent, audited; removing a topic group that sessions reference fails.
- Typed validation errors (field path + machine code + English message) that REST and MCP render without re-checking.
- Non-goals: read models, completion, transport.

## Acceptance criteria
- [x] Table tests: create a week; repeat the same request (no diff, no audit); change one session; past-date session skipped; completed session untouched; makeup link valid/invalid; shrinking the set deletes only open future sessions.
- [x] Notation tests with Turkish chapter titles; exact runes U+00B7 and U+2013.
- [x] Audit rows carry source and before/after JSON.
- [x] Fixture `testdata/sample-week.json` (labelled sample data) for the week of 2026-10-26 matching the mockup rows, plus `testdata/sample-outline.json` (Faz 1 · İşlem temeli, five topic groups from the mockup).

## Checks
- `go test ./internal/plan/...`
- `go vet ./...`

## Progress
Done (2026-10-26; nothing committed).

- **`internal/plan`** is the single business-logic layer. Files: `plan.go` (domain types: `Source`, `Item`, `SlotRef`, `SessionInput`, request/result types, enum sets), `notation.go` (formatter), `errors.go` (`FieldError`/`ValidationError` + codes + `errBuilder`), `validate.go` (structural, store-free), `service.go` (`Service`, `SavePlan`, audit/idempotency/snapshot helpers), `outline.go` (`SaveOutline`).
- **`SavePlan(ctx, source, req)`** — declarative range replace over `[from,to]`: open (today-or-later, uncompleted) sessions become exactly the given set, matched by `(date,start_time)`; matched→update-in-place (id kept) or unchanged; absent open future→delete; frozen (past date or completed) never touched, incoming targeting them → `Skipped` (`past_date`/`completed`). Idempotent with no key (repeat → all unchanged, 0 audit) and with `idempotency_key` (hash stored response replayed; different payload under same key → `ErrIdempotencyConflict`). `makeup_of` is a `SlotRef` resolved to the missed session's id (must exist, be past and uncompleted). Returns `{created,updated,deleted,unchanged,skipped}` sorted by (date,start_time). One `store.WithTx`; audit row per create/update/delete with source + before/after JSON + request hash.
- **`SaveOutline(ctx, source, req)`** — phases and topic groups replaced by key (create/update/unchanged/delete), idempotent, audited; deleting a topic group still referenced by sessions fails with `CodeReferenced` (whole tx rolls back).
- **Notation** (`notation.go`): `FormatItem` → `M2·17, s. 210–225` (middle dot U+00B7, en dash U+2013), optional `, Antr. 1–3`; single values collapse the range; `FormatItems` joins with ` · `. Runes pinned by test (ASCII hyphen rejected).
- **Store additions (WKY-03 layer):** `WithTx` callback now takes `*sql.Tx` (was the unexported `dbtx`) so other packages can compose helpers; `store_test.go` updated. New helpers: `catalog.go` (Phase/TopicGroup types + list/insert/update-by-key/delete-by-key, `TopicGroupByKey`, `CountSessionsByTopicGroup`); in `queries.go` `UpdateSession`, `DeleteSession`, `DeleteSessionItems`, `SessionByDateAndTime`.
- **Fixtures:** `testdata/sample-week.json` (week 1, 4 evenings + Saturday long block, Wednesday carries konu/soru/antrenman split) and `testdata/sample-outline.json` (Faz 1 + five topic groups); both labelled sample data via a `_comment`.
- **Tests:** `notation_test.go`, `validate_test.go`, `service_test.go` (create+idempotent via fixtures, change-one, past-date-skip, completed-untouched, shrink-deletes-open-future, makeup valid/unknown/future, idempotency-key replay+conflict, audit before/after, outline replace + referenced-delete). All green: `go test ./internal/plan/... ./internal/store/...`, `go vet ./...`, `gofmt -l .` empty, `CGO_ENABLED=0 go build ./...`.
