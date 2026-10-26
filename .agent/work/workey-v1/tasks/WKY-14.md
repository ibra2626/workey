# WKY-14 — Visual review of all screens and v1 gate

## Goal
Every screen (Giriş, Bugün, Hafta, İlerleme) critiqued at mobile and desktop against DESIGN.md and fixed; the v1 exit criteria met and reported.

## Scope
- Screenshots at 390×844 and 1280×800 for all four screens plus key states (Bugün done/undo, rest day, empty plan; Hafta with moved and missed rows).
- Critique with `/impeccable` (critique/audit) against DESIGN.md: Three Zones, Soft Token, Step-Up, Ink Label, Two Widths, Exact Notation, Sentence Case, First Panel, Flat Ledger rules; 12px text floor; 4.5:1 contrast; focus-visible on every control; keyboard path; `prefers-reduced-motion`.
- Fix findings in templates/CSS; re-screenshot.
- Align DESIGN.md wording with the v1 decisions (Seri and Denemeler removed, "Geri al" link, single-block profile for uzun_blok/deneme) via the impeccable documenter flow, with the user's OK.
- Gate (local Windows target; deploy is out of this gate while WKY-12 is blocked): `go test ./...`, `go vet ./...`, `gofmt -l .`, govulncheck report (findings listed, not hidden).
- Non-goals: commit/push/PR (only on the user's explicit request).

## Acceptance criteria
- [ ] Before/after screenshot list with each finding and its fix in Progress.
- [ ] `go test ./...` passes; `go vet ./...` clean; govulncheck output summarized.
- [ ] CURRENT.md states v1 complete or lists what is left.

## Checks
- `go test ./...`
- `go vet ./...`
- `gofmt -l .`
- `go run golang.org/x/vuln/cmd/govulncheck@latest ./...`

## Progress
Not started.
