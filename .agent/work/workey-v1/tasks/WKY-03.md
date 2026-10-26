# WKY-03 — SQLite store and schema

## Goal
A tested persistence layer on modernc.org/sqlite with embedded migrations holding everything v1 needs.

## Scope
- `internal/store`: open with WAL, `foreign_keys=ON`, `busy_timeout`; single-writer pattern; embedded ordered `.sql` migrations tracked in `schema_migrations`.
- Tables (names may be refined; record changes in DECISIONS):
  - `phases` (id, key, ordinal, name, start_date, end_date). Plan start = first phase start.
  - `topic_groups` (id, key, name, ordinal, dgs_question_min, dgs_question_max).
  - `sessions` (id, date, start_time, duration_min, kind `aksam|uzun_blok|deneme`, topic_group_id null, makeup_of null → sessions.id, completed_at null, completed_source null, created_at, updated_at); unique (date, start_time).
  - `session_items` (id, session_id, ordinal, book `M1|M2|M3|G1|G2`, chapter_no, chapter_title, page_from, page_to, konu_from/to null, soru_from/to null, antrenman_from/to null).
  - `audit_log` (id, at_utc, source `mcp|api|web`, action, entity, entity_id, before_json, after_json, request_hash null).
  - `idempotency_keys` (key, request_hash, response_json, created_at).
- Store methods are plain queries inside a transaction helper (`WithTx`); no business rules here.
- Non-goals: validation, freezing logic, HTTP.

## Acceptance criteria
- [ ] Migrations run on an empty DB and are no-ops on a migrated DB.
- [ ] CHECK constraints for enums and `page_from <= page_to`; FK cascade from sessions to items.
- [ ] Tests on a temp-dir DB: sessions with items by date range, audit insert, idempotency key round trip.
- [ ] DECISIONS entry: modernc.org/sqlite (no CGO, single binary) and the migration approach.

## Checks
- `go test ./internal/store/...`
- `go vet ./...`
- `CGO_ENABLED=0 go build ./...`

## Progress
Not started.
