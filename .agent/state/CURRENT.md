# CURRENT — WKY-05 done, plan read models + completion live

Updated 2026-10-26.

Active task: .agent/work/workey-v1/tasks/WKY-05.md (done; not committed)

- Feature: `workey-v1` — Go + templ + htmx + Tailwind single binary; Giriş / Bugün / Hafta / İlerleme; REST `/api/v1` and MCP `/mcp` over one `internal/plan` service; SQLite (modernc). Local Windows only; WKY-12 (deploy) blocked by the user.
- Plan: `.agent/work/workey-v1/TASKS.md`; shared invariants + layout in `.agent/work/workey-v1/README.md` (read with each task). Module path `github.com/ibra2626/workey`, go1.27.3.
- Done: WKY-01 (skeleton/config/clock/server), WKY-02 (frontend toolchain + base layout), WKY-03 (store + schema), WKY-04 (plan write path), WKY-05 (read models + completion). WKY-01..04 are committed (HEAD `0ae785d`); WKY-05 is **not committed** yet. Index statuses 02–05 set to done.

## WKY-05 result (done, not committed)
The read side and completion, all computed in `internal/plan` so web/REST/MCP only render.

- **New files** — `read.go` (view models + `Day`/`Week`/`Progress`), `complete.go` (`Complete`/`Undo`), `dates.go` (Turkish calendar + formatters). Plus tests `read_test.go`, `complete_test.go`, `dates_test.go`.
- **`Day(ctx, date)` → `DayView`**: the day's `SessionView`s, `Tomorrow` (next day's first session), week-scoped `Notes`, `IsRest`, Turkish `Heading` ("Çarşamba, 28 Ekim").
- **`Week(ctx, date)` → `WeekView`**: 7 `DayRow`s (rest days included), `PlannedMin`/`DoneMin`, `SessionsPlanned`/`SessionsDone`, `MovedCount`, `WeekNumber` (1..38), `PhaseName`, `RangeLabel` ("26 Eki–1 Kas"), `Notes`. Planned totals count every session (incl. missed/moved); done counts only completed.
- **`Progress(ctx, examDate)` → `ProgressView`**: `HasPlan`, `WeekNumber`/`WeeksTotal`(=38), sessions & hours done/planned (`DoneHours`/`PlannedHours`, decimal comma), `CurrentPhase`, `DaysToExam`/`AfterExam`/`ExamDate`, 5× `GroupProgress` (DGS range + completed share = done/planned minutes, `Percent`), 38× `WeekCell` strip (per-week done fraction + phase membership). `examDate` is a **parameter** (from `WORKEY_EXAM_DATE`), not a `New` arg — `New(st, clk)` unchanged.
- **Derived `SessionState`** `planned|today|done|missed|moved`: done > (past → moved if referenced by a makeup link, else missed) > today > planned. Moved row carries `MovedTo{Date, Weekday}`; "moved" is derived from the link (`store.ListMakeupLinks`), never stored. Note line: "Salı yapılamadı; M1·2 cumartesi bloğuna eklendi." (missed weekday capitalised, makeup weekday lowercased, `Ref` of the missed first item).
- **`Complete`/`Undo(ctx, source, sessionID)`**: today-or-past only (future → `ValidationError` code `future`; missing → `not_found`); idempotent (already in target state → no write, no audit row); audited as action `complete`/`undo` with before/after JSON and **NULL** request hash; return the updated `SessionView`.
- **`dates.go`**: weekday/month name tables (full/lower/short, lowercase written out — no Turkish case conversion), `FormatDayHeading`, `FormatShortDate`, `FormatWeekRange` (en dash U+2013), `FormatHoursTR` (90→"1,5", 540→"9", 75→"1,25"). Civil-date math (`toDate`/`addDays`/`civilDays`) on UTC midnights — zone-free, DST-proof. `WeeksTotal = 38`; plan start = first phase's Monday (fallback earliest session's Monday).
- **Store additions (WKY-03 layer, persistence only):** `queries.go` + `SessionByID`, `AllSessions`, `SetSessionCompletion`, `ListMakeupLinks` (+ type `MakeupLink`), private `sessionItems`; `errors.go` + code `CodeFuture = "future"` and `fieldError` helper.
- **Checks all green:** `go test ./internal/config/... ./internal/clock/... ./internal/store/... ./internal/plan/... ./cmd/...`, `go vet ./...`, `gofmt -l .` empty, `CGO_ENABLED=0 go build ./...`. No templ/Tailwind touched (no UI files). DECISIONS entry added (WKY-05).
- **Uncommitted now:** new `internal/plan/{read,complete,dates}.go` + their `_test.go`; modified `internal/store/queries.go`, `internal/plan/errors.go`; `.agent/` docs (TASKS index, WKY-05 Progress, DECISIONS, this file).

## Next step
WKY-06 (REST API `/api/v1` with Bearer auth) — depends on WKY-05. Build `internal/api` + `internal/httpx` (bearer middleware, JSON errors, security headers) over `plan.Service`: write endpoints (save plan/outline, complete/undo) and read endpoints (Day/Week/Progress) that parse → call → render the service's typed results; `ValidationError` → structured JSON, `ErrIdempotencyConflict` → 409. New session → read TASKS.md index then `tasks/WKY-06.md`. Nothing is committed for WKY-05; a release session stages `internal/plan/{read,complete,dates}.go` (+tests), `internal/store/queries.go`, `internal/plan/errors.go`, and the `.agent/` docs.
