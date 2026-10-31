# CURRENT — Study room on Bugün done; SF-06 next

Updated 2026-10-30.
Active task: .agent/work/study-flow/tasks/SF-05.md
Next: SF-06 integration/visual acceptance (now including the study room) and
product docs. WKY-20 actual 200% zoom acceptance pending; WKY-12 deployment held.

## Study room B · Seans çizelgesi — 2026-10-30 (latest)

User picked mock-up B. studyRoom (study.templ) now: label/title/first step, one
row per stage (roomRows in study.go: real clock ranges from last_started_at minus
bookmark offset, pages), states Bitti / şu an + minute / N dk sonra set by
room.js, plain over-time sentence, Bitirdim + Burada bıraktım; side column with
companion and sound chips. Removed profile/needle/axis, roomStage/roomTicks/
roomTenPct, header sound menu. Checks: templ, Tailwind, build (CGO off too),
vet, gofmt, go test ./internal/web/... (91) passed; IAB 1440/375 states, 179-min
over-time and no console errors. Mock-ups: scratchpad oda-alternatifleri.html.
Follow-up: header no longer shows the phase name (only "Hafta N / 38"); topic line
"Konu: <topic group> · <book>, N. bölüm" (studyTopic) under the room/paused title
and "Konu: ..." on the plan panel; desktop phase panel label is now
"Bu dönemin haftaları · H1–H7". web tests passed; IAB checked.

## Plan screen — 2026-10-31 (calendar)

GET /plan ("Plan" in rail + mobile tab bar, 5 tabs, icon i-plan). Per user, the
plan is a wide month calendar: plan.Service.Overview (internal/plan/overview.go,
phases with start/end, weeks, sessions, today) → views/planpage.go planMonths
(Mon-first grids) → planpage.templ: phase summary, month strip (?ay=YYYY-MM,
prev/next, default today's month; planCalendar), legend, ONE month panel,
session chips by state; hover/focus opens a wide detail (day, phase, kind/time,
chapter, notation, Konu, konu/soru/yanlış split, state) that grows out of the
cell via clip-path (CSS only, flat ink border, no shadow); phones: tap opens a
sheet above the tab bar. CSS .plan-*/.cal-* in input.css. Tests:
internal/web/plan_test.go; usage_test expects grid-cols-5. Checks: templ,
Tailwind, build, vet, gofmt, go test ./internal/web/... ./internal/plan/...
passed; IAB 375/900/1440 no overflow, focus opens/closes detail, no console
errors. Not checked: real mouse hover visually (pane screenshots timed out).

## Study room simplification + sounds — 2026-10-30

Room now has one header sound menu (Kapalı/Kütüphane/Kafe/Yağmur, volume, chime)
and two actions (Bitirdim pill, Burada bıraktım outlined pill; close-without-page
inside the disclosure). Fullscreen and in-room plan link removed. room.js sound
engine rewritten: convolver room, panning, formant murmur/whispers, footsteps,
eraser, pencil, paper riffles, café cups/stir/grinder/steam/chairs. Checks:
node --check, templ, Tailwind, go build, gofmt, go test ./internal/web/...
passed; IAB 1440/375 layout, menu, Escape; café/library ran ~30s each at volume
0 without console errors. Then, per user, beds/reverb/chatter cut (café 2 tables),
events spaced out, rain darker and sparser. Audible quality NOT verified.

## Study room — 2026-10-30 (first version, superseded by B)

Room state (stage + elapsed minutes) is computed in web/static/room.js from
last_started_at + bookmarked-stage offset; nothing new is stored. Ambient sound
is Web Audio (prefs in localStorage); room.js loads on Bugün only. PRODUCT.md,
DESIGN.md (Study room) and DECISIONS.md updated. Not verified: audible output on
real speakers, Safari/iOS audio unlock, screen reader, full suite, govulncheck.

## SF-05 result — 2026-10-29

Week shows partial study, resolved source/stage/resume page, no-bookmark fallback
and makeup relation; versioned finish/undo remain. Today/Week/Progress explain
planned minutes of completed sessions, partial exclusion and separate original +
makeup counting (no unique-content claim). Shared API/MCP source checklist and
optional first-week/replan prompts use closed previews; exam prompt retained.
Docs planning-inputs sections generated from shared content with parity tests.
Go 1.27.3 focused web/views/plan/apidoc tests, templ/Tailwind, normal/CGO-off
builds, vet, changed Go gofmt and diff whitespace passed. SQLite totals and real
HTTP finish/undo bookmark regressions added. Pre-existing views/views.go format
output untouched. No live browser/clipboard/zoom, full suite/race or govulncheck;
SF-06 owns integrated acceptance. Same checkout/codex/study-flow-experiment;
personal env/data and prior edits retained, no commit/push/merge/deploy.
Receipt failed normally and elevated: specified SUBU Python/Atlas absent.
Evidence: .agent/work/study-flow/tasks/SF-05.md.
## SF-04 result — 2026-10-29

Today now has authenticated/CSRF/versioned start and pause forms, normal/10-minute
entry, a deterministic first page, active-first workspace, explicit old paused
resume and today's plan access. Bookmark form prefills saved values, shows every
source/stage range and next-stage/source/finish guidance; raw invalid input and
its old version survive errors. No automatic page increment. Completion is one
action without starting; existing versioned Complete/Undo retained. Rest/exam
empty states cannot hide active close/complete. StudyPeriodOpen uses the same
service bounds as StartStudy; partial work never changes completion totals.
Htmx 422/409 feedback/focus and busy controls work; network failure preserves
fields and restores controls. No-JS start/bookmark redirect; close explanation
and errors render full pages. No timer/localStorage/interval added.
Go 1.27.3 focused web/views/plan tests, templ + Tailwind 4.3.3, normal/CGO-off
builds, vet, changed Go gofmt, JS syntax and diff whitespace passed. Actual SQLite
close/reopen, independent auth sessions and all kinds covered. IAB: long Turkish
multi-source first step/actions visible at 390x844; 320px open form and 1280px
layout without overflow; keyboard focus, two-tab stale close, real dropped POST
network failure and CSP script-src none lifecycle checked. Temporary fixture
servers/source removed. Personal env/data untouched; SF-01..03 edits preserved.
Same checkout/codex/study-flow-experiment; no commit/push/merge/deploy.
No full suite/race/vulnerability rescan/screen reader/actual reduced-motion
setting test; SF-06 integration acceptance stays separate. Evidence: SF-04.md.
Beyin receipt failed in normal/elevated runs: specified SUBU Python/Atlas absent.

## Retained SF-01..03 results — 2026-10-29

SF-01: 0002 session_study migration, store queries and atomic started-session
SavePlan protection; normal/CGO-off builds and focused plan/store checks passed.
SF-02: transactional StartStudy/PauseStudy/ReadStudyContext, single active,
versions/idempotency/bookmark validation and lifecycle; versioned web completion,
legacy unversioned API/MCP completion retained. Focused checks/builds passed.
SF-03: optional SessionView.study and Day active/latest paused context; batched
snapshot reads, atomic Complete/Undo output; OpenAPI/MCP/shared synthetic usage
contract docs and parity tests. Focused plan/store/API/apidoc/MCP/web/views,
vet/builds passed (MCP localhost elevated). Detailed evidence in tasks/SF-01.md,
SF-02.md and SF-03.md. Receipt failed there too: runtime/Atlas absent.
