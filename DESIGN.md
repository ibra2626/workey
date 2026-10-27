---
name: Workey
description: A calm training log for a 38-week DGS sayısal plan; tonight's session first, one tap to mark it done.
colors:
  ground: "#eef1ee"
  card: "#ffffff"
  ink: "#16302a"
  ink-hover: "#0d211c"
  ink-2: "#4a5d57"
  ink-3: "#5f706a"
  line: "#d5ddd8"
  line-strong: "#b8c4be"
  z1: "#66a69d"
  z2: "#8aa7c4"
  z3: "#d0963a"
  z1-soft: "#c3dcd8"
  z3-soft: "#f7ecd9"
typography:
  headline-display:
    fontFamily: "Barlow, system-ui, sans-serif"
    fontSize: "44px"
    fontWeight: 700
    lineHeight: 1.25
  headline:
    fontFamily: "Barlow, system-ui, sans-serif"
    fontSize: "28px"
    fontWeight: 700
    lineHeight: 1.25
  title:
    fontFamily: "Barlow, system-ui, sans-serif"
    fontSize: "15px"
    fontWeight: 600
    lineHeight: 1.5
  body:
    fontFamily: "Barlow, system-ui, sans-serif"
    fontSize: "15px"
    fontWeight: 400
    lineHeight: 1.5
  meta:
    fontFamily: "Barlow, system-ui, sans-serif"
    fontSize: "13px"
    fontWeight: 400
    lineHeight: 1.4
    fontFeature: "tnum"
  label:
    fontFamily: "Barlow, system-ui, sans-serif"
    fontSize: "12px"
    fontWeight: 600
    lineHeight: 1.33
  numeral-display:
    fontFamily: "Barlow Condensed, Barlow, sans-serif"
    fontSize: "44px"
    fontWeight: 600
    lineHeight: 1
    fontFeature: "tnum"
  numeral-stat:
    fontFamily: "Barlow Condensed, Barlow, sans-serif"
    fontSize: "30px"
    fontWeight: 600
    lineHeight: 1
    fontFeature: "tnum"
  numeral-inline:
    fontFamily: "Barlow Condensed, Barlow, sans-serif"
    fontSize: "17px"
    fontWeight: 600
    lineHeight: 1.3
    fontFeature: "tnum"
  numeral-hero:
    fontFamily: "Barlow Condensed, Barlow, sans-serif"
    fontSize: "64px"
    fontWeight: 600
    lineHeight: 1
    fontFeature: "tnum"
  numeral-rail:
    fontFamily: "Barlow Condensed, Barlow, sans-serif"
    fontSize: "40px"
    fontWeight: 600
    lineHeight: 1
    fontFeature: "tnum"
  numeral-page:
    fontFamily: "Barlow Condensed, Barlow, sans-serif"
    fontSize: "36px"
    fontWeight: 600
    lineHeight: 1
    fontFeature: "tnum"
  wordmark:
    fontFamily: "Barlow Condensed, Barlow, sans-serif"
    fontSize: "24px"
    fontWeight: 600
    lineHeight: 1.33
  numeral-profile:
    fontFamily: "Barlow Condensed, Barlow, sans-serif"
    fontSize: "22px"
    fontWeight: 600
    lineHeight: 1
    fontFeature: "tnum"
  numeral-inline-lg:
    fontFamily: "Barlow Condensed, Barlow, sans-serif"
    fontSize: "20px"
    fontWeight: 600
    lineHeight: 1.2
    fontFeature: "tnum"
  title-lg:
    fontFamily: "Barlow, system-ui, sans-serif"
    fontSize: "16px"
    fontWeight: 600
    lineHeight: 1.5
  body-small:
    fontFamily: "Barlow, system-ui, sans-serif"
    fontSize: "14px"
    fontWeight: 400
    lineHeight: 1.45
  guide-section:
    fontFamily: "Barlow, system-ui, sans-serif"
    fontSize: "20px"
    fontWeight: 600
    lineHeight: 1.4
  code:
    fontFamily: "monospace"
    fontSize: "13px"
    fontWeight: 400
    lineHeight: 1.6
  numeral-unit:
    fontFamily: "Barlow Condensed, Barlow, sans-serif"
    fontSize: "18px"
    fontWeight: 500
    lineHeight: 1
    fontFeature: "tnum"
rounded:
  xs: "2px"
  bar: "4px"
  block: "6px"
  nav: "8px"
  panel: "14px"
  pill: "9999px"
spacing:
  tight: "8px"
  gutter: "12px"
  panel-mobile: "20px"
  panel: "24px"
  panel-lg: "28px"
  page: "32px"
components:
  button-primary:
    backgroundColor: "{colors.ink}"
    textColor: "{colors.card}"
    typography: "{typography.title-lg}"
    rounded: "{rounded.pill}"
    height: "54px"
    padding: "0 40px"
  button-primary-hover:
    backgroundColor: "{colors.ink-hover}"
    textColor: "{colors.card}"
  panel:
    backgroundColor: "{colors.card}"
    textColor: "{colors.ink}"
    rounded: "{rounded.panel}"
    padding: "{spacing.panel}"
  tab-item:
    backgroundColor: "{colors.card}"
    textColor: "{colors.ink-3}"
    typography: "{typography.label}"
    height: "68px"
  tab-item-active:
    backgroundColor: "{colors.card}"
    textColor: "{colors.ink}"
  rail-item:
    textColor: "{colors.ink-3}"
    typography: "{typography.body}"
    rounded: "{rounded.nav}"
    padding: "10px 12px"
  rail-item-active:
    backgroundColor: "{colors.ground}"
    textColor: "{colors.ink}"
  row-today:
    backgroundColor: "{colors.z3-soft}"
    textColor: "{colors.ink}"
  bar-done:
    backgroundColor: "{colors.ink}"
    rounded: "{rounded.bar}"
  bar-planned:
    backgroundColor: "{colors.line}"
    rounded: "{rounded.bar}"
  bar-today:
    backgroundColor: "{colors.z3-soft}"
    rounded: "{rounded.bar}"
  bar-rest:
    backgroundColor: "{colors.line}"
    rounded: "{rounded.pill}"
    height: "3px"
  progress-track:
    backgroundColor: "{colors.line}"
    rounded: "{rounded.bar}"
    height: "6px"
  input-password:
    backgroundColor: "{colors.card}"
    textColor: "{colors.ink}"
    typography: "{typography.body}"
    rounded: "{rounded.nav}"
    padding: "10px 12px"
  usage-control:
    backgroundColor: "{colors.card}"
    textColor: "{colors.ink}"
    typography: "{typography.body}"
    rounded: "{rounded.nav}"
    padding: "8px 12px"
  usage-control-selected:
    backgroundColor: "{colors.ground}"
    textColor: "{colors.ink}"
    rounded: "{rounded.nav}"
  guide-selected:
    backgroundColor: "{colors.ink}"
    textColor: "{colors.card}"
    rounded: "{rounded.nav}"
  usage-code:
    backgroundColor: "{colors.ground}"
    textColor: "{colors.ink}"
    typography: "{typography.code}"
    padding: "16px"
---

# Design System: Workey

## Overview

**Creative North Star: "The Training Log"**

Workey reads like the log of a calm, well-coached runner. The 38-week DGS plan is periodised like marathon preparation: phases (Faz), weekly volume in minutes, a fixed number of sessions (seans). Each evening session is drawn the way a training app draws an interval workout, as three blocks of intensity on a time axis. Numbers do the talking, set large in a narrow condensed face like a sports-watch display; words explain them in a plain humanist sans.

The world is light and quiet: a pale grey-green ground, white panels, dark forest-green ink, and exactly three zone colours that belong to the three parts of a session (konu, soru, yanlış defteri). Nothing else is coloured. A missed evening is a measurement, not a verdict: it is hatched, labelled "cumartesiye taşındı", and the weekly total is shown unchanged. Volume and completed sessions are figures in the log, never rewards. v1 has no streak, target score, or deneme-results panels.

Density is moderate and glanceable. On the phone at 20:30 the first panel answers the whole question: book reference, chapter title, page range, the session profile, one dark pill button.

**Key Characteristics:**
- Flat white panels on a pale grey-green ground; no shadows, no nesting.
- Three zone colours, each bound to one session part; everything else is ink, greys and line.
- Condensed tabular numerals for every figure; humanist sans for every sentence.
- State is carried by fill and pattern (ink, hatch, ochre outline, line grey, 3px rest stub), never by badges.
- One primary action per surface: the dark ink pill "Tamamlandı".

## Colors

A cool, low-chroma forest palette with three muted "heart-rate zone" accents reserved for the parts of a session.

### Primary
- **Forest Ink** (ink): the text colour, the primary button, the "done" fill in every chart, the current-week marker in the 38-week strip, and the filled portion of progress tracks. Doing the work turns things ink.
- **Pressed Forest** (ink-hover): hover state of the primary button only.

### Secondary (session zones)
- **Easy-Pace Teal** (z1): Konu, the 20-minute topic block. Lowest block in the session profile, legend swatch.
- **Main-Set Ochre** (z3): Soru, the 50-minute question block. Tallest block in the profile; also the 2px outline that marks today in weekly bar charts.
- **Recovery Blue-Grey** (z2): Yanlış defteri, the 20-minute mistake-notebook block; mid height.

### Tertiary (soft fills)
- **Phase Mist** (z1-soft): planned weeks of the current phase in the 38-week strip and the phase-week blocks; also the text-selection highlight (Forest Ink text on Phase Mist).
- **Today Sand** (z3-soft): the background of today's row in lists and tables, and the fill of today's bar.

### Neutral
- **Trail Ground** (ground): app background; also the active rail item's background on desktop.
- **Card White** (card): panels, the bottom tab bar, the desktop rail.
- **Moss Ink** (ink-2): secondary text that must hold on tinted fills (book line, chapter reference, sub-lines inside a today row).
- **Lichen Ink** (ink-3): tertiary text: meta lines, units, axis ticks, inactive tabs, rest days.
- **Hairline** (line): dividers, panel-internal rules, chart gridlines, planned and future bars, empty progress tracks.
- **Strong Hairline** (line-strong): the 1px inset ring on hatched "moved" bars and dashed future markers in charts.

### Named Rules
**The Three Zones Rule.** z1, z2 and z3 mean konu, yanlış defteri and soru, in every chart, legend and swatch. They are never used for decoration, status, buttons or links. The only extra job allowed is z3's ochre outline marking today.

**The Soft Token Rule.** Tints are explicit tokens (z1-soft, z3-soft), never opacity modifiers on a variable colour. The production stack (templ + Tailwind standalone CLI) cannot compile `bg-[color:var(--z1)]/35`; add a named soft token instead.

**The Step-Up Rule.** On Today Sand, secondary text uses Moss Ink, never Lichen Ink (Lichen on Sand is 4.47:1, under the floor). The build already does this in today's list row.

**The Ink Label Rule.** Labels on zone-coloured blocks are Forest Ink, never white. Ink passes on all three zones: Teal 5.04:1 (z1 was retuned from #5b9c93, which gave only 4.44:1), Ochre 5.44:1, Blue-Grey 5.64:1. White on any zone fails.

## Typography

**Display Font:** Barlow Condensed (with Barlow, sans-serif), weights 500/600/700, always tabular.
**Body Font:** Barlow (with system-ui, sans-serif), weights 400/500/600/700.

**Character:** Road-sign and sports-watch legibility. Figures are narrow and large so volume, days and page ranges read at a glance; descriptions stay in normal-width Barlow so sentences stay calm.

### Hierarchy
- **Numeral Display** (600, 44px, line-height 1): the headline figure of a panel (weekly volume, plan week, days to exam on the rail at 40px). The desktop Hafta hero scales to 64px and desktop page ranges sit at 36px; same face and weight.
- **Numeral Stat** (600, 30px, line-height 1): the three stat tiles on mobile Bugün.
- **Numeral Inline** (600, 16-20px): durations ("20 dk"), percentages, day-of-month numbers in the week list, the book reference ("M1·3") above a chapter title.
- **Numeral Unit** (500, 16-22px, Lichen Ink; 500 is the condensed default weight, so unit spans need no weight class): the denominator or unit after a figure ("/ 540 dk", "/ 38 hafta", "gün").
- **Headline Display** (700, 44px, tight): tonight's chapter title on desktop Bugün only.
- **Headline** (700, 28px, tight): chapter title on mobile Bugün, page titles on desktop.
- **Title** (600, 15-17px): date headers, list item names, topic-group names, button text (17px).
- **Body** (400, 15px, normal; relaxed for notes): descriptions, table cells, notes.
- **Meta** (400, 13px, tabular, Lichen Ink): page ranges and times under a title, side context in headers.
- **Label** (600, 12px, Lichen Ink, sentence case): names a measurement, column or panel ("Bu hafta", "Sınava", "Haftalık hacim", table headers). Never uppercase, never tracked.
- **Guide section** (600, 20px, line-height 1.4): guide panel h2 headings. Nested headings and disclosure headings use 16px semibold. Guide prose stays 15px with line-height 1.6 and a maximum measure of 72ch; code uses the system monospace stack at 13px, line-height 1.6.

### Named Rules
**The Two Widths Rule.** If it is a number the user reads as a quantity (minutes, days, weeks, pages, percentages), it is Barlow Condensed with tabular figures. If it is a sentence, it is Barlow. Times and page ranges inside running text keep Barlow with `tnum`.

**The Exact Notation Rule.** Book references are written in the user's notation, `M2·17, s. 210–225`: middle dot between book and chapter, "s." only for sayfa, en dash in ranges, decimal comma ("8,5", "1,5").

**The Sentence Case Rule.** No uppercase transforms anywhere. Turkish casing stays correct by never asking CSS to do it; if uppercase is ever unavoidable the element must sit under `lang="tr"` so i maps to İ.

## Layout

Mobile (reference width 390px): a single column. A 20px-inset header line (date left, "Hafta 1 · Faz 1" meta right), then panels inset 12px from the screen edge with 12px gaps; panel padding 20px (14px in stat tiles). The bottom tab bar is fixed to the bottom, 68px tall. The first panel is always tonight's session and holds the primary action.

Desktop (reference width 1280px): a 232px left rail plus a main area with 32px padding and 20px gaps. Bugün uses a two-column grid (main panel plus a 340px side column of small panels: Bu hafta, Yarın, Not) with a full-width phase panel below. İlerleme stacks the plan strip and topic groups in one column; v1 has no score or deneme-results side panels. Hafta stacks a volume panel (260px figure column plus chart) above a full-width table. Panel padding is 24-32px (larger for the main session panel).

Rhythm: 8px between small units, 12px between mobile panels and list items, 20px between desktop panels. Figures sit with `leading-none` and 4px above them for their label.

The app shell switches from the fixed bottom bar to the desktop rail at 768px.
Usage pages retain the same shell in a maximum 1200px reading container. Their
panels use 20px padding below 768px and 28px above it, with 12/20px gaps. At
1200px viewport width, a 180px contents column sits beside the flexible reading
column, separated by 24px; below that, the contents list is a native disclosure.
Long code lines scroll locally. The mobile main area reserves 88px below its
content for the 68px fixed navigation.

**The First Panel Rule.** On Bugün, at any width, the first panel contains book reference, chapter, page range, session profile and "Tamamlandı" with no scrolling on a 390 x 844 screen.

## Elevation & Depth

Flat. Product surfaces carry no shadows; depth is tonal only: white panels on the grey-green ground, the rail and tab bar separated by a 1px hairline. The only box-shadows in the build are inset strokes that encode state (2px ochre ring for today, 1px strong-hairline ring around hatched bars), not elevation. The soft drop shadow around each artboard in the exploration file is presentation chrome and is not part of the system.

**The Flat Ledger Rule.** If a surface needs to stand out, change its fill (Today Sand) or its content, never lift it with a shadow, blur or gradient.

## Shapes

Soft but not bubbly. Panels have generously rounded corners (14px). Chart bars are small rounded rectangles whose radius scales with their size: 2-3px in the dense 38-week strip, 4-5px in weekly bars and phase blocks, 6px for the desktop session-profile blocks. Rest days are a 3px fully rounded stub on the baseline. Desktop rail items use 8px corners; the primary button is a full pill. Legend swatches are 10-12px squares with 2px corners. Icons are stroked line drawings (1.8 stroke, round caps and joins) at 22px in the tab bar, 20px in the rail and button, 16px in tables.

## Components

### Buttons
Decisive and singular: one dark pill per surface.
- **Shape:** full pill.
- **Primary ("Tamamlandı"):** Forest Ink fill, white 16px semibold text, a 20px check icon before the label, 54px tall full-width on mobile, 56px tall with 40px side padding on desktop. The login action uses the same pill at 54px and 15px semibold text.
- **Hover / Focus:** background eases to Pressed Forest over 200ms using the compiled Tailwind default ease; focus-visible is a 2px Forest Ink outline offset 3px (a teal ring would be only 2.5:1 against the ground). Reduced motion reduces transition and animation duration to 0.01ms.
- **Secondary:** guide controls are flat white rectangles with an 8px radius, 1px Hairline border and a minimum 44px height. Padding is 8px by 12px; hover changes the border to Forest Ink and the fill to Trail Ground. Selected view/client links use the ground fill, ink border and semibold text; selected guide navigation uses ink fill and white text. Focus is a 2px ink outline offset 2px. Copy controls are hidden until JavaScript attaches them; disabled controls use a wait cursor.

After completion, replace the pill with a plain Tamamlandı status and a quiet Geri al form button. Completion and undo refresh the session, mobile summary, desktop weekly volume and phase blocks together.

### Interaction feedback (27 October 2026)

The primary completion action now reads **Seansı tamamla**; **Tamamlandı** is
reserved for the completed state. This supersedes the action label in the
original specifications above and below. The 54/56px forest pill and existing
layout remain. A small outlined checkbox, pointer cursor, darker hover fill,
visible keyboard focus and 120ms press response communicate interactivity.
The check appears on hover/focus and is always visible on touch devices.

Completion, undo and weekly row actions replace their label with a spinner and
Turkish pending text while preserving button geometry. All actions in the
affected content region are disabled until the request finishes; competing
submissions are dropped. Requests time out after 20 seconds. Failed requests
show an accessible message with recovery steps and restore controls. Keyboard
focus follows the same session's action after the content is replaced.

Shared navigation and login show an indeterminate 3px ink loading line after
180ms; fast loads do not flash it. Navigation uses normal page rendering;
snapshot view transitions and whole-content fades are removed because they
competed with chart motion and made navigation feel stalled. The completed
check draws once (280ms). Guide disclosure and copy-result feedback use 150ms fades.
There are no celebrations, new colours or shadows.
Reduced motion removes spatial movement, looping and page transitions while
retaining loading text, a static loading line and the final completion state.
The Today content grid uses one column below 1280px to keep the session action
and side summaries inside intermediate viewports; the shell still switches at
768px. Wide desktop retains the 340px side column.

Chart fills grow upward from their own baseline over 650ms using composited
scaleY transforms and `cubic-bezier(.22, .61, .36, 1)`. Horizontal topic tracks
grow from the left with scaleX. Session labels occupy a separate stationary
layer so text never scales. Animation layers are released at completion; there
is no permanent will-change, clip-path animation, loop or layout movement.
Rest-day stubs stay static. Reduced-motion users see the final chart immediately.

Critical first-screen Barlow weights and Turkish subsets are preloaded before
the stylesheet using identical font URLs and anonymous CORS. Font-display is
optional: if a font misses the short initial rendering window, the existing
system fallback remains stable for that visit instead of swapping mid-animation.

### Cards / Containers (panels)
- **Corner Style:** 14px.
- **Background:** Card White on Trail Ground.
- **Shadow Strategy:** none (see Elevation & Depth).
- **Border:** none; internal lists use hairline dividers.
- **Internal Padding:** 20px mobile, 24-32px desktop; stat tiles 14px.
- Panels sit side by side or stacked, never inside one another. A list or table inside a panel uses dividers, not inner cards.

### Inputs and selectable text
The password input is Card White with Forest Ink text, a 1px Hairline border,
8px corners, 10px by 12px padding and 15px body text. Its label is 12px semibold
Lichen Ink. Focus changes the border to ink and adds a 2px ink outline offset 2px.
Readonly guide textareas use 13px monospace at line-height 1.6, 16px padding,
8px corners, a Hairline border and vertical resizing. They retain native text
selection and the same visible focus treatment.

### Lists and tables
- **Week list (mobile):** rows 12px x 16px, condensed day number (20px) in a 36px column, title + meta line. Done rows end with a check icon; moved rows say "cumartesiye taşındı" in meta; rest rows are a single Lichen Ink "Dinlenme" line.
- **Week table (desktop):** label-style header row over a hairline, 15px body, hairline dividers, tabular figures. A "Durum" column carries words: "Tamamlandı" with check, "Cumartesiye taşındı", "Bugün 20:30", "Planlı".
- **Today row:** Today Sand background, semibold, secondary text stepped up to Moss Ink.

### Navigation
- **Mobile bottom tab bar:** Card White, top hairline, four equal tabs (Bugün, Hafta, İlerleme, Kullanım) with a 22px stroked icon over a 12px label. Inactive Lichen Ink; hover turns Forest Ink over 150ms; active Forest Ink and semibold, marked with `aria-current="page"`. Every tab is a real link (`href`) and shows a 2px Forest Ink inset focus ring (offset -2px, 8px corners) on focus-visible.
- **Desktop left rail:** 232px, Card White, right hairline. "Workey" wordmark in Barlow Condensed 24px semibold, then five items: Bugün, Hafta, İlerleme, MCP Kullanımı and API Kullanımı (20px icon, 15px label, 10px x 12px padding, 8px corners). Active item gets a Trail Ground fill and Forest Ink semibold text. Hover and focus-visible follow the tab bar: Forest Ink text on hover, 2px Forest Ink outline offset 2px. The rail's foot holds the days-to-exam figure on study screens.

### Guide reading controls
Guide pages use sibling white panels and internal dividers. Native disclosures
have minimum 44px summary targets; their 16px semibold headings stay inline with
the native disclosure marker. The wide contents list is sticky 24px from the
top and scrolls within the viewport when long; section links are real fragments.
Code blocks have a language/copy toolbar above an ink-on-ground preformatted
area. The area is keyboard focusable and scrolls horizontally without widening
the page. Copy success/failure uses 13px Lichen Ink live text rather than a zone
colour. On failure, hidden text is revealed, focused and selected for manual
copying. The long plan prompt keeps its copy button and replacement warning
outside its initially closed preview. Full Markdown explicitly includes all
clients, while formatted content shows only the selected client.

### Session profile (signature)
Tonight's 90 minutes drawn as an interval workout on a 20:30-22:00 axis (ticks every 30 minutes, 12-13px tabular Lichen Ink). Three blocks share the width in proportion to their minutes (20 / 50 / 20) and stand at heights that read as intensity: konu low (about 43% of the chart height, Easy-Pace Teal), soru tall (about 92%, Main-Set Ochre), yanlış defteri mid (about 64%, Recovery Blue-Grey), with a 4px gap between them. On desktop each block carries its name (Barlow 15px semibold) and duration (Barlow Condensed 22px semibold) in Forest Ink; the profile is 150px tall. On mobile the 92px profile blocks are unlabelled and a three-row legend below pairs a 10px zone swatch, the part name with its pages, and the condensed duration.

For uzun_blok and deneme, use one Phase Mist block with the full duration and start/end ticks; do not invent a 20/50/20 split. Multi-source sessions list every book/chapter/page reference in flat rows inside the session panel. The deneme session kind remains valid input, but v1 has no deneme score tracking.

### Volume bars (signature)
Daily minutes for the week, Monday to Sunday, bottom-aligned with equal columns. State vocabulary, used identically everywhere:
- **Done:** Forest Ink fill.
- **Moved:** 135° hatch of Hairline stripes (3px on, 3px off) with a 1px Strong Hairline inset ring; the text says where it went.
- **Today:** Today Sand fill with a 2px Main-Set Ochre inset ring, in every chart including the small desktop "Bu hafta" chart; the day label below is Forest Ink semibold. Day labels under bars are 12px, the smallest text size in the system.
- **Planned:** Hairline grey fill.
- **Rest:** a 3px fully rounded Hairline stub.

### Plan strip (38 weeks)
Thirty-eight equal columns, 2-4px gaps: the current week shows its done fraction in Forest Ink, the remaining weeks of the current phase are Phase Mist, later phases are Hairline. A small legend (Yapılan, Faz 1 planı, Sonraki fazlar) and a date axis sit around it. On desktop Bugün the current phase appears as seven 44px week blocks in Phase Mist, the current one partly filled with ink.

### Progress tracks
6px (mobile) or 8px (desktop) Hairline tracks with a Forest Ink fill for completed share, paired with a condensed percentage; zero values show the percentage in Lichen Ink.

## Do's and Don'ts

### Do:
- **Do** answer "Bugün ne çalışacağım?" in the first panel: book reference, chapter title, page range, session profile, then "Tamamlandı".
- **Do** keep completion to one tap on the single Forest Ink pill.
- **Do** state a missed or moved session plainly in words ("Salı yapılamadı; M1·2 cumartesi bloğuna eklendi.") alongside the hatched bar, and show weekly totals unchanged.
- **Do** set every quantity in Barlow Condensed with tabular figures, and every sentence in Barlow.
- **Do** write Turkish with correct ş ğ ı İ ç ö ü, sentence case, decimal comma, and the `M2·17, s. 210–225` notation.
- **Do** put Forest Ink labels on zone blocks and keep all text at 4.5:1 or above; on Today Sand use Moss Ink for secondary text.
- **Do** use the explicit soft tokens (z1-soft, z3-soft) for tints, and reference tokens instead of hard-coded hex in SVG fills.
- **Do** use stroked SVG icons (1.8 stroke, round caps) and give every interactive element a visible focus state.
- **Do** label sample figures as sample data until the plan arrives through the API (the exploration file says "örnek veri" in its note and section headings).
- **Do** keep 12px as the floor for any text, chart axes included.

### Don't:
- **Don't** use purple-blue gradients, or any colour gradient on a surface (the moved-bar hatch is a stripe pattern, not a gradient).
- **Don't** use glassmorphism, backdrop blur or drop shadows on product surfaces.
- **Don't** use emoji as icons.
- **Don't** add badges, streak rewards, celebrations or confetti; v1 uses a Seans tile instead of Seri; volume remains a measurement.
- **Don't** nest panels inside panels; use dividers inside a panel.
- **Don't** ship lorem ipsum or other filler text.
- **Don't** colour anything with z1, z2 or z3 that is not the matching session part (today's ochre ring excepted).
- **Don't** put white text on a zone colour.
- **Don't** use opacity modifiers on CSS-variable colours (`/35`); they do not compile in the production Tailwind CLI.
- **Don't** mark a missed day in red or with warning icons; hatching plus a plain sentence is the whole treatment.
- **Don't** abbreviate "soru" as "s."; "s." means sayfa. Write "8–10 soru" at every width.
- **Don't** uppercase labels or rely on CSS case transforms for Turkish text.
