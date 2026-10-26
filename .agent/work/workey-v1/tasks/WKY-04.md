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
- [ ] Table tests: create a week; repeat the same request (no diff, no audit); change one session; past-date session skipped; completed session untouched; makeup link valid/invalid; shrinking the set deletes only open future sessions.
- [ ] Notation tests with Turkish chapter titles; exact runes U+00B7 and U+2013.
- [ ] Audit rows carry source and before/after JSON.
- [ ] Fixture `testdata/sample-week.json` (labelled sample data) for the week of 2026-10-26 matching the mockup rows, plus `testdata/sample-outline.json` (Faz 1 · İşlem temeli, five topic groups from the mockup).

## Checks
- `go test ./internal/plan/...`
- `go vet ./...`

## Progress
Not started.
