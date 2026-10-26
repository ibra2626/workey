# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Stack

Go + templ + htmx, Tailwind CSS (standalone CLI). No JavaScript framework. Server-rendered HTML; interactivity through htmx only.

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
- Studying happens on paper with the physical books; the app is consulted before and after, not during, the session.

## Capabilities and Constraints

- Shows the daily/weekly assignment (day, book, chapter, page range) and tracks completion.
- Exposes an MCP/API interface so an AI assistant can create and update the plan. The app itself creates no plan content.
- Single user; no multi-user, social, or sharing features.
- UI language is Turkish. Turkish characters (ş, ğ, ı, İ, ç, ö, ü) must render correctly, and case conversion must respect Turkish rules (i ↔ İ, ı ↔ I).
- Must work well on both desktop and phone screens.
- v1 decisions (2026-10-26, see `.agent/state/DECISIONS.md`): no question counts, deneme results or streak; missed sessions stay visible as "yapılamadı" and the AI may link a makeup session; no target score.

## Brand Commitments

Name: Workey. No logo or palette yet.

Personality, in the user's words: sakin, net, kararlı (calm, clear, decisive). It should feel calm, focused, and reassuring: "a long road, but I am moving forward." It must never feel like a tracker that makes the user feel guilty. Missed or late items are stated plainly, not as reproach.

Explicitly rejected by the user (binding):
- purple-blue gradients
- glassmorphism
- emoji used as icons
- gamification such as badges, streak rewards, or confetti
- cards nested inside cards
- lorem ipsum or other filler text

## Evidence on Hand

No content, data, or assets exist yet; the repository is empty. The 38-week plan will arrive through the AI/MCP interface. Do not fabricate plan items, progress figures, or scores in product copy; sample data must be clearly labeled as such.

## Product Principles

1. Tonight's assignment comes first: what to open, which pages, nothing to decode.
2. The app follows the plan, it does not author it. Content and scheduling come through the API.
3. Logging takes seconds, so it never competes with the 90-minute study window: one tap to mark done.
4. Honest progress without guilt: show where the plan stands against the exam date, including slippage, framed as a long road being travelled, not as failure.
5. Book references stay exact and in the user's own notation (`M2·17, s. 210–225`).
