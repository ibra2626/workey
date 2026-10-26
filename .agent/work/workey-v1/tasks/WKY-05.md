# WKY-05 — Plan service: read models and completion

## Goal
Everything the screens, REST and MCP read, plus complete/undo, computed once in `internal/plan`.

## Scope
- `Day(date)`: sessions with items and derived state `planned|today|done|missed|moved` (moved = missed and referenced by a makeup session; carries the makeup date and weekday), tomorrow's session, generated note lines ("Salı yapılamadı; M1·2 cumartesi bloğuna eklendi.").
- `Week(date)`: Monday–Sunday rows incl. rest days, planned vs done minutes, sessions done/planned, moved count, plan week number (1..38) and current phase.
- `Progress()`: weeks elapsed/total, sessions and hours done/planned (from the saved plan), per topic group completed share and DGS question range, days to exam (`WORKEY_EXAM_DATE`), 38-week strip data (per-week done fraction, phase membership).
- `Complete(ctx, source, sessionID)` / `Undo(...)`: allowed for today and past, rejected for future; idempotent; audited.
- Turkish weekday/month names and date formatting ("Çarşamba, 28 Ekim", "26 Eki–1 Kas") live here, not in templates.
- Empty states as data: no plan yet, rest day, before plan start, after exam.
- Non-goals: transport, templates.

## Acceptance criteria
- [ ] Fake-clock tests on `testdata/sample-week.json` reproduce the mockup figures (90 / 540 dk, 1 / 5 seans, 1 taşındı) once Monday is completed and Tuesday is missed with a Saturday makeup.
- [ ] Completing a future session → typed error; completing twice → one audit row; undo restores.
- [ ] Istanbul midnight boundary test (23:59 vs 00:00 local).

## Checks
- `go test ./internal/plan/...`
- `go vet ./...`

## Progress
Not started.
