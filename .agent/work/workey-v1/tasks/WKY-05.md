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
- [x] Fake-clock tests on `testdata/sample-week.json` reproduce the mockup figures (90 / 540 dk, 1 / 5 seans, 1 taşındı) once Monday is completed and Tuesday is missed with a Saturday makeup.
- [x] Completing a future session → typed error; completing twice → one audit row; undo restores.
- [x] Istanbul midnight boundary test (23:59 vs 00:00 local).

## Checks
- `go test ./internal/plan/...`
- `go vet ./...`

## Progress
Done (2026-10-26; nothing committed).

- **Read models in `internal/plan/read.go`** — all three computed in the service, screens/REST/MCP just render:
  - `Day(ctx, date)` → `DayView`: the day's `SessionView`s, `Tomorrow` (next day's first session), week-scoped `Notes`, `IsRest`, Turkish `Heading` ("Çarşamba, 28 Ekim").
  - `Week(ctx, date)` → `WeekView`: 7 `DayRow`s (rest days included), `PlannedMin`/`DoneMin`, `SessionsPlanned`/`SessionsDone`, `MovedCount`, `WeekNumber` (1..38), `PhaseName`, `RangeLabel` ("26 Eki–1 Kas"), `Notes`.
  - `Progress(ctx, examDate)` → `ProgressView`: `HasPlan`, `WeekNumber`/`WeeksTotal` (38), sessions & hours done/planned (`DoneHours`/`PlannedHours` with decimal comma), `CurrentPhase`, `DaysToExam`/`AfterExam`/`ExamDate`, per-group `GroupProgress` (DGS range + completed share by minutes), 38-cell `Strip` (per-week done fraction + phase membership). `examDate` is a parameter (from `WORKEY_EXAM_DATE`), not a `New` arg.
- **Derived state** `planned|today|done|missed|moved` (`SessionState`): done > (past: moved if referenced by a makeup, else missed) > today > planned. A moved session carries `MovedTo{Date, Weekday}`. Note line: "Salı yapılamadı; M1·2 cumartesi bloğuna eklendi." (missed weekday capitalised, makeup weekday lowercased, `Ref` of the missed item).
- **Completion in `internal/plan/complete.go`** — `Complete`/`Undo(ctx, source, sessionID)`: today-or-past only (future → `ValidationError` code `future`), missing → `not_found`; idempotent (no-op when already in the target state, no audit row); audited as `complete`/`undo` with before/after JSON and NULL request hash; both return the updated `SessionView`.
- **Turkish calendar in `internal/plan/dates.go`** — weekday/month name tables (full/lower/short), `FormatDayHeading`, `FormatShortDate`, `FormatWeekRange` (en dash U+2013), `FormatHoursTR` (90→"1,5", 540→"9"). Lowercase weekday forms are written out, not case-converted, so the Turkish i/İ rule never trips.
- **Store additions (persistence only):** `queries.go` adds `SessionByID`, `AllSessions`, `SetSessionCompletion`, `ListMakeupLinks` (+ `MakeupLink`) and a private `sessionItems` loader; `errors.go` adds code `future` + `fieldError` helper.
- **Civil-date arithmetic** (`toDate`/`addDays`/`civilDays`) uses UTC midnights — zone-free and DST-proof (Turkey has no DST since 2016); `WeeksTotal = 38`; plan start = first phase's Monday (fallback: earliest session's Monday).
- **Tests** (`read_test.go`, `complete_test.go`, `dates_test.go`): Week mockup figures (90/540, 1/5, 1 taşındı) + states + note line; Day Bugün (heading, tomorrow, notes) + rest day; Progress (week 1/38, hours 1,5/9, days-to-exam, 5 groups with 50% temel-kavramlar, 38-cell strip) + empty plan; complete future-rejected/missing/idempotent-one-audit, undo restores + idempotent, Istanbul 23:59-vs-00:00 boundary; Turkish date/hours formatting with exact runes. All green: `go test ./...`, `go vet ./...`, `gofmt -l .` empty, `CGO_ENABLED=0 go build ./...`.
