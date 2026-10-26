# Architectural Decisions

Record accepted decisions with date, context, decision, alternatives, and consequences. Do not record proposals as accepted decisions.

## 2026-10-26 — Stack: Go + templ + htmx + Tailwind standalone CLI
- Context: greenfield single-user web app; user is a web developer.
- Decision: server-rendered Go with templ, htmx for interactivity, Tailwind via standalone CLI, no JS framework (user's choice during `/impeccable init`).
- Alternatives: static HTML/CSS, Next.js/React, Astro (offered, not chosen).
- Consequences: Tailwind opacity modifiers on CSS-variable colors do not compile; use explicit soft tokens (see DESIGN.md Soft Token Rule).

## 2026-10-26 — Visual direction locked: "Antrenman günlüğü" (Yön 04)
- Context: six directions explored (notebook, Swiss grid, night focus, training log, two hybrids).
- Decision: user chose Yön 04; recorded in `DESIGN.md` and `.impeccable/design.json`; reference artifact `design/directions/04-antrenman-gunlugu.html`. Other directions and review screenshots deleted.
- Alternatives: 01–03, 05, 06 (deleted, not recoverable — never committed).
- Consequences: teal z1 retuned #5b9c93 → #66a69d so ink labels pass 4.5:1; focus rings use Forest Ink.

## 2026-10-26 — v1 runs locally on Windows; deploy task on hold
- Context: the brief asked for systemd + Nginx deploy files; the user will run v1 only locally on Windows for now.
- Decision (user): WKY-12 (Linux build, systemd, Nginx) is blocked until the user lifts it; a Linux server may come later. The v1 gate does not include deploy.
- Alternatives: write the deploy files now without using them.
- Consequences: default bind stays `127.0.0.1:8080`; code stays OS-neutral. Phone access is not possible with this setup unless the app is later exposed on the LAN or deployed; that needs a separate decision (TLS, Secure cookies).

## 2026-10-26 — Go toolchain: go1.27.3
- Context: installed Go is 1.27.0; go.dev reports go1.27.3 as latest stable (2026-10-19). Patch releases carry security fixes that govulncheck would report against stdlib.
- Decision: target go1.27.3 (`go 1.27.0` + `toolchain go1.27.3` in go.mod); upgrade the system Go before or at WKY-01.
- Alternatives: stay on 1.27.0 and accept stdlib govulncheck findings.

## 2026-10-26 — WKY-01: module path, tzdata embed, Go upgrade done
- Context: scaffolding the Go skeleton; needed a module path, a timezone source that does not depend on the host, and the pinned toolchain in place.
- Decision: module path `github.com/ibra2626/workey` (user-confirmed, matches git user ibra2626). `internal/clock` imports `_ "time/tzdata"` so Europe/Istanbul loads from the embedded tzdata and the single binary works on a machine without a system tz database. System Go upgraded to go1.27.3 (user did it this session); `go.mod` is `go 1.27.0` + `toolchain go1.27.3`.
- Alternatives: bare module name `workey` (not future-proof for a remote repo); rely on host tzdata (breaks on minimal systems); keep system Go at 1.27.0 and let GOTOOLCHAIN=auto download (user preferred upgrading outright).
- Consequences: all internal packages import under `github.com/ibra2626/workey/...`. tzdata adds a small amount to the binary; acceptable for a self-contained local app.

## 2026-10-26 — v1 scope: no streak, no deneme tracking
- Context: the mockup shows a "Seri" tile, a "Seviye testi 8,5 / 25" panel, a "Denemeler · net" chart and an "H4 · deneme" marker; the v1 brief puts question/deneme tracking and streak-style gamification out of scope.
- Decision (user, during `/task-plan workey-v1`): remove the Seri tile, the seviye testi panel and the Denemeler chart. The third mobile Bugün stat tile becomes "Seans 1 / 5" (done / planned sessions this week). A deneme day may exist only as a planned session of kind `deneme`; no results are stored.
- Alternatives: remove everything with two stat tiles; keep Seri as a pure measurement.
- Consequences: DESIGN.md mentions of Seri and Denemeler are superseded for v1; screen tasks follow this record.

## 2026-10-26 — Missed sessions: past is frozen, makeup link
- Context: auto-rescheduling is out of scope, but the design shows "cumartesiye taşındı".
- Decision (user): a session whose date (Europe/Istanbul) has passed without completion shows as "yapılamadı". Plan writes never touch days before today or completed sessions. When the AI adds a makeup session it may set `makeup_of` to the missed session's id; the missed row then reads "<gün>e taşındı" with the hatched bar.
- Alternatives: "yapılamadı" only, no link; let the AI overwrite past uncompleted sessions.
- Consequences: save-plan responses report skipped (frozen) sessions; the moved state is derived from the link, never set directly.

## 2026-10-26 — Completion: undo and marking from Hafta
- Decision (user): after "Tamamlandı" the panel shows a quiet "Geri al" text link. Hafta rows for today and past days can be marked done or undone. Future days cannot be marked. Every change is audited with its source (web/api/mcp).
- Alternatives: today only with undo; no undo.

## 2026-10-26 — Session data model: 1..n items, no free-text note
- Decision (user): a session has date, start time, duration (minutes), kind (`aksam`, `uzun_blok`, `deneme`), optional topic group, optional `makeup_of`, and 1..n items. An item has book (M1, M2, M3, G1, G2), chapter number, chapter title, page range, and optional splits: konu pages, soru pages, antrenman range. No free-text note field; the desktop "Not" panel is generated from state (missed/moved sessions).
- Alternatives: same plus a 140-character AI note; one item per session.
- Consequences: the app stores references only (book, chapter, pages), never topic explanations or question content.
