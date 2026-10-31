# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Stack

Go + templ + htmx, Tailwind CSS (standalone CLI), SQLite. No JavaScript framework. Server-rendered HTML and ordinary form fallbacks; htmx updates and small local scripts provide feedback, focus, copy and motion enhancements.

## Users

One user: İbrahim, a full-time web developer preparing for the DGS sayısal (quantitative) section in the evenings and on weekends. He uses the app on both desktop and phone, but the primary moment is 20:30 on the phone, right before the evening block: one glance answers "Bugün ne çalışacağım?", and when the work is done, one tap marks it "tamamlandı".

## Product Purpose

Workey is a personal study-plan tracker for DGS sayısal preparation. It shows what to study on a given day — which book, which chapter, which page range — and records whether it was done. It does not teach, explain, or generate study content.

Success: the user always knows tonight's assignment at a glance, can mark it done in seconds, and can see whether the 38-week plan is on track for the exam date.

## Positioning

The plan is authored outside the app: an AI assistant builds it and writes it into Workey through an MCP/API interface. Workey is the single source of truth that the AI writes to and the user follows. It is a precise ledger of a book-and-page plan, not a generic to-do list or a content platform.

## Operating Context

- Exam: 18 July 2027. The plan spans 38 weeks, starting Monday 26 October 2026.
- Rhythm: 4 weekday evenings × 90 minutes (20:30–22:00), plus 1 weekend block of 3 hours.
- Evening structure: 20 min topic → 50 min questions → 20 min mistake notebook (yanlış defteri).
- Sources: Antrenmanlarla Matematik 1–3 and Antrenmanlarla Geometri 1–2.
- Plan items reference a book, a chapter and a page range, written like `M2·17, s. 210–225` (Matematik 2, chapter 17, pages 210–225).
- Studying happens on paper with the physical books. Since 2026-10-30 the app may also stay open during the session as a study room (stage, elapsed time, ambient sound, a quiet companion); it never requires it.

## Capabilities and Constraints

- Shows the daily/weekly assignment (day, book, chapter, page range) and tracks completion.
- Bugün offers an exact first step, normal or 10-minute entry, and one active study session. The 10-minute target changes neither planned duration nor completion. Active work opens a study room that shows the current stage and elapsed minutes computed in the browser from the start time; this is display only, never stored, sent or counted toward progress.
- Active work and the optional source/stage/resume-page bookmark live in SQLite and survive refresh, another device and process restart. Resuming old work is an explicit choice; Istanbul midnight changes today's plan without moving or closing the old session.
- “Burada bıraktım” records the page to resume, without incrementing it. “Çalışmayı kapat” needs no page and preserves an existing bookmark. “Bitirdim” works without starting; completion closes study atomically, and undo preserves the bookmark as paused work.
- Started sessions cannot be replaced or removed by save_plan. Web forms detect stale study versions; legacy REST/MCP completion remains compatible without a version and cannot detect stale intent. Study context is read-only in REST/MCP.
- Progress counts completed sessions' planned minutes. Partial work is shown separately as “Kısmen çalışıldı”; no unique-content or actual-time claim is made. Original and makeup sessions count separately if both are explicitly completed.
- API/MCP guides share a source-material checklist and optional first-week/replanning prompts. Materials are supplied to the connected assistant; Workey provides no PDF upload, storage or OCR.
- Exposes an MCP/API interface so an AI assistant can create and update the plan. The app itself creates no plan content.
- Single user; no multi-user, social, or sharing features.
- UI language is Turkish. Turkish characters (ş, ğ, ı, İ, ç, ö, ü) must render correctly, and case conversion must respect Turkish rules (i ↔ İ, ı ↔ I).
- Must work well on both desktop and phone screens.
- v1 decisions (2026-10-26, see `.agent/state/DECISIONS.md`): no question counts, deneme results or streak; missed sessions without a study record show "yapılamadı", partial work is identified separately, and the AI may link a makeup session; no target score.

## Brand Commitments

Name: Workey. The supplied logo is embedded from `web/static/workey.png`; the locked forest-ink palette, Barlow typography and visual rules are in `DESIGN.md`.

Personality, in the user's words: sakin, net, kararlı (calm, clear, decisive). It should feel calm, focused, and reassuring: "a long road, but I am moving forward." It must never feel like a tracker that makes the user feel guilty. Missed or late items are stated plainly, not as reproach.

Explicitly rejected by the user (binding):
- purple-blue gradients
- glassmorphism
- emoji used as icons
- gamification such as badges, streak rewards, or confetti
- cards nested inside cards
- lorem ipsum or other filler text

## Evidence on Hand

The repository contains the Go application, embedded web assets, SQLite migrations, REST/MCP interfaces and synthetic test fixtures. The real plan is authored through the AI/MCP interface; these fixtures do not establish the user's actual plan or progress. Do not fabricate plan items, progress figures, or scores in product copy; sample data must be clearly labeled as such. Study-flow acceptance and remaining browser checks are recorded in `docs/study-flow-acceptance.md`.

## Product Principles

1. Tonight's assignment comes first: what to open, which pages, nothing to decode.
2. The app follows the plan, it does not author it. Content and scheduling come through the API.
3. Logging takes seconds, so it never competes with the 90-minute study window: one tap to mark done.
4. Honest progress without guilt: show where the plan stands against the exam date, including slippage, framed as a long road being travelled, not as failure.
5. Book references stay exact and in the user's own notation (`M2·17, s. 210–225`).
