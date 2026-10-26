# Project map

Bootstrapped on 2026-10-26. Filled from verified repository evidence; unverified items are marked.

- Product: personal DGS sayısal study-plan tracker for one user; shows tonight's book/chapter/page assignment and records completion. Source: `PRODUCT.md`.
- Stack and installed versions: Go + templ + htmx + Tailwind standalone CLI (user decision recorded in `PRODUCT.md` ## Stack). **Unverified:** no `go.mod`, templ files or Tailwind config exist yet; no versions installed or pinned.
- Main directories:
  - `PRODUCT.md`, `DESIGN.md` — product truth and locked design system.
  - `design/directions/04-antrenman-gunlugu.html` — static reference mockup (Tailwind CDN), 3 screens × mobile/desktop.
  - `.impeccable/` — design sidecar (`design.json`) and detector config (`config.json`).
  - `.claude/launch.json` — preview server (`python -m http.server 8765` serving `design/directions`).
- Focused verification commands: none yet (no code).
- Constraints: Turkish UI with correct Turkish casing; single user; plan content arrives via MCP/API only; mobile-first (primary use 20:30 on the phone). Exam 2027-07-18; plan of 38 weeks from 2026-10-26.
