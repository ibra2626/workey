# WKY-02 — Frontend toolchain: templ, Tailwind, fonts, htmx, base layout

## Goal
templ components render through the server with the DESIGN.md token set, self-hosted Barlow / Barlow Condensed and local htmx, all embedded in the binary.

## Scope
- templ as a Go tool (`go get -tool github.com/a-h/templ/cmd/templ`); decide whether generated `*_templ.go` is committed and record it in DECISIONS.
- Tailwind v4 standalone CLI binary in `/bin/` (gitignored). `web/styles/input.css` with `@theme` tokens copied exactly from DESIGN.md front matter (colours incl. `z1-soft`, `z3-soft`, radii, spacing, type scale) and `@source` pointing at `internal/web/views`. Output `web/static/app.css` (minified). No opacity modifiers on variable colours (Soft Token Rule).
- Fonts: Barlow 400/500/600/700 and Barlow Condensed 500/600/700, woff2 latin + latin-ext (Turkish glyphs), `font-display: swap`, OFL license file alongside. htmx 2.x minified as `web/static/htmx.min.js`. Ask the user before each download (file, source, size).
- `web/static` embedded via `embed.FS`; served under `/static/` with long cache headers and a version/hash query.
- `internal/web/views/layout.templ`: `<html lang="tr">`, viewport meta, CSS/htmx links, mobile bottom tab bar (68px) and desktop 232px rail (Bugün / Hafta / İlerleme, `aria-current`, stroked SVG icons per DESIGN.md, focus rings), slot for page content. A placeholder page proves the layout.
- `scripts/build.sh` (templ generate → tailwind → go build). `.claude/launch.json` gains a `workey-dev` configuration for the app server.
- Non-goals: real screens, data, auth.

## Acceptance criteria
- [ ] `go tool templ generate` then `go build ./...` clean.
- [ ] Placeholder page renders at 390px and 1280px in the preview; ş ğ ı İ ç ö ü render in both fonts (computed font is Barlow, not a fallback).
- [ ] No request leaves localhost (network panel); no CDN URL anywhere in the repo.
- [ ] Test: layout renders `lang="tr"` and the active tab with `aria-current="page"`.
- [ ] DECISIONS entries: templ module, Tailwind version, htmx version, font source.

## Checks
- `go tool templ generate`
- `./bin/tailwindcss -i web/styles/input.css -o web/static/app.css --minify`
- `go build ./... && go vet ./...`
- `go test ./internal/web/...`
- Preview screenshots at 390 and 1280.

## Progress
Done (2026-10-26).
- templ `v0.3.1100` added as a Go tool; generated `*_templ.go` committed (DECISIONS). Tailwind standalone CLI `v4.3.3` at `./bin/tailwindcss.exe` (gitignored). htmx `2.0.11` → `web/static/htmx.min.js`. Barlow 400/500/600/700 + Barlow Condensed 500/600/700, latin + latin-ext woff2 (14 files) + `OFL.txt` → `web/static/fonts/`. All user-approved in-session.
- `web/styles/input.css`: `@theme` tokens (13 colours incl. z1-soft/z3-soft, two font families, radii, spacing, type scale), `@font-face` per subset with `unicode-range` + `font-display: swap`, `@source` → views, `.cn`/`.tnum` utilities. Compiled minified to `web/static/app.css`.
- `web/assets.go`: `embed.FS` for `web/static`, `GET /static/` handler with immutable cache + `?v=` hash. `internal/web/views/` templ: `Layout` (`<html lang="tr">`, viewport, CSS/htmx, desktop 232px rail + mobile 68px tab bar with stroked SVG icons, `aria-current="page"`, focus rings), `Placeholder` page. `internal/web/web.go` serves `/`, `/hafta`, `/ilerleme`. Wired into `cmd/workey` mux. `scripts/build.sh` + `.claude/launch.json` `workey-dev`.
- Checks all clean: `go tool templ generate` (idempotent), `go build ./...`, `go vet ./...`, `gofmt -l .` (empty), `go test ./internal/config/... ./internal/clock/... ./internal/web/... ./cmd/...` (all ok). Browser (built-in): placeholder renders at 375 and 1280; rail/tab-bar switch at `md`; ş ğ ı İ ç ö ü covered by all loaded Barlow + Barlow Condensed faces (`document.fonts.check` true, computed font Barlow, not fallback); no console errors; htmx loaded; no CDN/network reference in source.
- Not committed (awaiting explicit request).
