# CURRENT — first-study-day fixes committed on feat/kalan-birim-video

Updated 2026-11-08.
No active task. Feature workey-v2-katalog (WKY-21..28) is complete and merged (PR #2).

## Objective of this work
After the first real study day: make unfinished units visible, show the day's
second block with its time and videos, and keep the plan consistent.

## Done (committed on branch feat/kalan-birim-video; not pushed, no PR)
- Migration `0003_unit_media.sql`: optional `units.title`, `units.url`; through
  store, plan (UnitInput/UnitView/snapshot), MCP schema, guide, Kaynaklar form.
- `internal/plan/leftover.go`: unit timeline, `LeftoverUnits`,
  `get_progress.leftover_units`; UnitRow `CarriedOver`, `NextDate`, title, url.
- Bugün: `primarySession` (first open), `todayOtherSession`, `sessionClock`,
  "Kalan birimler" panel, "Kalan: … · 12 Kas seansında", unit link row, video
  first step names the video.
- Tests: `internal/plan/leftover_test.go`, `internal/web/today_sessions_test.go`;
  migration count test now expects 3.
- Decision recorded in `.agent/state/DECISIONS.md` (2026-11-08).

## Live data changes (workey.db via MCP)
- save_plan: 12 Kas G1·1 Antrenman 5–9 (s. 17–26); 15 Kas Antrenman 10–14 (s. 27–36).
- save_source TR: 20 units with title + YouTube url (playlist index = ders number;
  the Tanıtım video, index 1, is not a catalog unit).
- Backup before migration: `workey.pre-unit-media.db` (gitignored).

## Checks (2026-11-08, Go 1.27.3)
templ generate; Tailwind rebuilt; `go build ./...`; `CGO_ENABLED=0 go build ./...`;
`go vet ./...`; focused test set config/clock/store/plan/web/api/apidoc/mcpserver/cmd
all ok; gofmt clean on LF. Rendered Bugün verified on a DB copy at 127.0.0.1:8099.
Not run: govulncheck (no dependency change).

## Runtime note
The app on 127.0.0.1:8080 currently runs in the Claude Terminal tab "workey sunucu".

## Next
Push/PR on request. Held: WKY-12 deploy, WKY-20 real 200% zoom acceptance.
PRODUCT.md still describes paper books; video-first reframing is the user's call.
