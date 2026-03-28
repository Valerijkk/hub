# Phase 8 -- UI Review

**Audited:** 2026-03-28
**Baseline:** HTML mockups in `redesign/` folder (source of truth for all design tokens)
**Screenshots:** Not captured (no dev server detected on ports 3000, 5173, 8080)

---

## Pillar Scores

| Pillar | Score | Key Finding |
|--------|-------|-------------|
| 1. Visual Fidelity | 4/4 | Dark tokens match mockup hex values exactly; light theme is a coherent inversion |
| 2. Consistency | 3/4 | Token usage is uniform across utilities, but a few hardcoded dark-only rgba values leak through |
| 3. Responsiveness | 3/4 | Tokens are viewport-agnostic by design; no responsive issues in token layer itself |
| 4. Accessibility | 1/4 | Zero focus-visible styles, zero prefers-reduced-motion, noise overlay lacks aria consideration |
| 5. Performance | 3/4 | 27 keyframes are all transform/opacity (compositable) except meshShift which animates filter |
| 6. Code Quality | 3/4 | Zero !important, clean token indirection, minor redundancy in shadow-card and scrollbar dark-only styles |

**Overall: 17/24**

---

## Top 3 Priority Fixes

1. **No focus-visible styles anywhere in app.css** -- Keyboard users have no visual focus indicator on interactive elements. Add a global `*:focus-visible { outline: 2px solid var(--primary); outline-offset: 2px; }` rule in `@layer base` and a ring utility `box-shadow: 0 0 0 2px var(--primary)` for components.

2. **No prefers-reduced-motion media query** -- Users who request reduced motion still see 27 animations including infinite loops (cardFloat, pulse-online, blink, breathing). Wrap all animation declarations in `@media (prefers-reduced-motion: no-preference) { ... }` or add a blanket `@media (prefers-reduced-motion: reduce) { *, *::before, *::after { animation-duration: 0.01ms !important; transition-duration: 0.01ms !important; } }`.

3. **page-header-blur hardcoded to dark rgba** -- Line 655 uses `rgba(13, 14, 18, 0.8)` which renders as near-black in light theme instead of a light translucent background. Add a light-theme variant or use `color-mix(in srgb, var(--base) 80%, transparent)`.

---

## Detailed Findings

### Pillar 1: Visual Fidelity (4/4)

All dark theme tokens match the mockup `:root` declarations exactly:

| Token | Mockup Value | app.css `.dark` Value | Match |
|-------|-------------|----------------------|-------|
| --void | #08090B | #08090B | Yes |
| --base | #0D0E12 | #0D0E12 | Yes |
| --raised | #141519 | #141519 | Yes |
| --card | #191A20 | #191A20 | Yes |
| --input | #1E2028 | #1E2028 | Yes |
| --sidebar | #0B0C0F | #0B0C0F | Yes |
| --t1 | #F2F3F5 | #F2F3F5 | Yes |
| --t2 | #8C8F97 | #8C8F97 | Yes |
| --t3 | #4E5058 | #4E5058 | Yes |
| --t4 | #363840 | #363840 | Yes |
| --b1 | rgba(255,255,255,.045) | rgba(255,255,255,.045) | Yes |
| --b2 | rgba(255,255,255,.08) | rgba(255,255,255,.08) | Yes |
| --b3 | rgba(255,255,255,.13) | rgba(255,255,255,.13) | Yes |
| --blue | #5B9CF5 | #5B9CF5 | Yes |
| --purple | #A78BFA | #A78BFA | Yes |
| --green | #34D399 | #34D399 | Yes |
| --amber | #FBBF24 | #FBBF24 | Yes |
| --red | #F87171 | #F87171 | Yes |
| --r1 | 6px | 6px | Yes |
| --r2 | 8px | 8px | Yes |
| --r3 | 12px | 12px | Yes |
| --r4 | 16px | 16px | Yes |

Easing tokens, shadows, glow backgrounds, font stacks all match across mockups. Noise overlay SVG feTurbulence parameters (baseFrequency .9, numOctaves 4, opacity .03 inner, .3 outer) are identical to mockup `body::after`.

Light theme values are a reasonable inversion not present in dark-only mockups -- white surfaces, gray text hierarchy, reduced shadow opacity. This was a stated design decision in 08-CONTEXT.md ("full visual QA in Phase 13").

Font loading: `index.html` line 12 loads Inter (400-900) + JetBrains Mono (400, 600) matching the dashboard/messenger mockup requirements.

### Pillar 2: Consistency (3/4)

**Strengths:**
- All component utilities (kanban-card, dash-widget, dash-qa-btn, glass, etc.) use design tokens consistently: `var(--card)`, `var(--b1)`, `var(--r3)`, `var(--sh2)`, `var(--ease)`.
- Dual @theme inline mapping ensures both old Shadcn classes (`bg-card`, `text-foreground`) and new tokens (`bg-void`, `text-t1`) work.
- Badge classes properly split light/dark with `.dark` prefix.

**Issues:**
- `page-header-blur` (line 655): hardcoded `rgba(13, 14, 18, 0.8)` -- works only in dark theme.
- `subtle-pulse` animation (line 479): uses `rgba(59, 130, 246, 0.4)` -- old blue (#3B82F6) instead of new blue token (#5B9CF5 dark / #3B7EE0 light). Same issue in `btn-glow` (line 629) which uses `rgba(91, 156, 245, ...)` -- correct for dark but not tokenized.
- `pulse-subtle` (line 516): uses `rgba(34, 197, 94, 0.4)` instead of `var(--green)`.
- Scrollbar thumb (lines 369, 378): hardcoded `rgba(255, 255, 255, ...)` -- invisible on light theme.
- `.shadow-card` dark variant (line 266-268) is identical to light variant -- redundant rule.

### Pillar 3: Responsiveness (3/4)

Design tokens are inherently viewport-agnostic -- CSS custom properties do not change per breakpoint, which is correct for a token foundation phase. No responsive issues in the token layer itself.

**Minor notes:**
- Login mockup has `@media(max-width:1024px){.left{display:none}.right{flex:1}}` -- app.css login utilities don't include this breakpoint, but this is likely handled in the React component with Tailwind responsive classes.
- No responsive font scaling defined in tokens (e.g., fluid typography). This is acceptable for Phase 8 scope.

### Pillar 4: Accessibility (1/4)

**Critical gaps:**

1. **No focus-visible styles** -- Zero occurrences of `focus-visible` or `:focus` in the entire 1009-line file. Interactive elements (buttons, inputs, links) have transition styles (line 345) but no focus ring. Keyboard navigation is effectively invisible.

2. **No prefers-reduced-motion** -- 27 `@keyframes` definitions, including 7 infinite animations (cardFloat, pulse-online, blink, breathing, shimmer, pulse-subtle, speaking-pulse). Users with vestibular disorders or motion sensitivity have no opt-out.

3. **Noise overlay z-index** -- `body::after` at z-index 9999 with `pointer-events: none` is non-interactive, but screen readers may still attempt to describe the pseudo-element if content is somehow exposed. Current implementation is safe (empty `content: ''`), but worth noting.

4. **Scrollbar contrast** -- Light theme scrollbar thumb uses dark-theme rgba values, making it invisible on white backgrounds.

### Pillar 5: Performance (3/4)

**Strengths:**
- All 27 animations use `transform` and `opacity` -- GPU-compositable properties that don't trigger layout or paint.
- Noise overlay is a single fixed pseudo-element with `pointer-events: none` -- rendered once, GPU-composited.
- Theme transitions use 150-200ms durations -- fast enough to feel instant, long enough for smoothness.
- No `will-change` declarations (good -- avoids unnecessary memory allocation).

**Issues:**
- `meshShift` animation (line 669) animates `filter: hue-rotate() brightness()` -- triggers repaint on every frame for 20 seconds continuously. This is a login-only animation so impact is limited.
- `body::after` noise SVG is an inline data URI with feTurbulence filter. Rendering cost depends on viewport size. At 1440x900 or 4K this could be expensive on low-end hardware. Consider adding `contain: strict` to limit repaint scope.
- 7 infinite animations could accumulate if multiple views are mounted simultaneously (e.g., multiple pulse indicators).

### Pillar 6: Code Quality (3/4)

**Strengths:**
- Zero `!important` declarations (all 56 were removed in Plan 02).
- Clean token hierarchy: primitive tokens (--void, --base) -> semantic tokens (--background, --card) -> utility classes.
- Well-organized sections with comment headers.
- Consistent naming: surfaces (void/base/raised/card/input), text (t1-t4), borders (b1-b3), shadows (sh1-sh4).
- Easing tokens are reusable and semantically named (--ease, --spring, --bounce).

**Issues:**
- `.shadow-card` (lines 262-268): `.dark .shadow-card` rule is identical to the light rule -- redundant.
- Scrollbar styles (lines 358-378): hardcoded dark-only colors, no `.dark` / `:root` split.
- `mesh-shift` keyframe (line 271) is defined but never used -- `meshShift` (line 669) is the actual animation name used.
- Some hardcoded colors leak through in animation utilities (subtle-pulse, pulse-subtle, btn-glow) instead of using CSS variable references.
- Token duplication: `--sans` and `--mono` are defined identically in both `:root` and `.dark` blocks. Since they don't change per theme, they could be defined once in `:root` only.

---

## Files Audited

- `frontend/src/app.css` (1009 lines) -- full design token system, utilities, animations
- `frontend/index.html` (19 lines) -- font loading, meta tags
- `redesign/01-login (1).html` -- token comparison source
- `redesign/02-dashboard (2).html` -- token comparison source
- `redesign/03-messenger (2).html` -- token comparison source
- `.planning/phases/08-design-system-foundation/08-CONTEXT.md` -- phase decisions
- `.planning/phases/08-design-system-foundation/08-01-SUMMARY.md` -- Plan 01 execution record
- `.planning/phases/08-design-system-foundation/08-02-SUMMARY.md` -- Plan 02 execution record
