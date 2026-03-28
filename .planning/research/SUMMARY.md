# Project Research Summary

**Project:** Hub v2.0 — Pixel-Perfect UI Redesign
**Domain:** Visual redesign of React 19 + Tailwind CSS v4 + Shadcn/ui application (11 HTML mockups)
**Researched:** 2026-03-27
**Confidence:** HIGH

## Executive Summary

Hub v2.0 is a visual-only redesign — no backend changes, no new features, no new npm packages. The scope is implementing 11 HTML mockups as a pixel-perfect dark theme, deriving a light theme from those tokens, extrapolating 6 unmocked pages from established patterns, and delivering premium micro-interactions. The project already has everything required: React 19, Tailwind CSS v4, Shadcn/ui, `motion` v12, `next-themes`, and a CSS variable theming system. The entire work surface is `app.css`, `index.html`, and the component/page layer in `src/`. Zero new dependencies.

The recommended approach is design-tokens-first: replace all CSS variables in a single atomic commit, let the cascade do 60–70% of the visual work automatically, then refine components and pages toward pixel-perfect mockup match. The mockups define a richer token system than the current Shadcn defaults — 5-level surface elevation (`--void` through `--input`), 4-level text hierarchy (`--t1` through `--t4`), 3-level borders (`--b1` through `--b3`), named easing curves (`--ease`, `--spring`), and shadow presets (`--sh1` through `--sh3`). These map cleanly onto the existing `@theme inline` / `:root` / `.dark` architecture without breaking any Shadcn component logic.

The primary risk is the existing `app.css` carrying 56 `!important` declarations that will create specificity conflicts the moment new component styles are applied. This must be addressed in Phase 1 before any restyling begins. Secondary risks are: (1) light theme accessibility — no light mockups exist, so the derived palette must be validated for WCAG AA contrast; (2) JSX structure changes silently breaking interactive behavior in Kanban, Messenger, and Calls; (3) the global CSS variable swap triggering radius/shadow/color ripples across all 235 source files simultaneously. All three risks have clear prevention strategies documented in PITFALLS.md.

---

## Key Findings

### Recommended Stack

The stack requires zero new npm packages. See STACK.md for full details.

**Font swap:** Replace Geist with Inter (weights 400–900) + JetBrains Mono (400–600) via Google Fonts CDN URL swap in `index.html`. Update `body { font-family: var(--sans) }` in `app.css`. Do not add `@fontsource/*` packages.

**Core technologies (all already installed):**
- `Tailwind CSS v4` — CSS-first `@theme inline` config drives all color/radius/shadow utilities from CSS variables; no tailwind.config.js changes needed
- `motion` v12 (`^12.38.0`) — currently used in 1 file; expand to page entrance animations, staggered lists, layout shifts; CSS transitions handle hover/press/focus (95% of interactions)
- `next-themes` (`^0.4.6`) — class-based `.dark` toggle; no changes to ThemeProvider; only CSS variable values change
- `class-variance-authority` (`^0.7.1`) — extend Button/Badge variants for new design states
- Native CSS — all glassmorphism, gradients, noise textures, and easing achievable without libraries

**New CSS utilities needed in `app.css`:** noise overlay (`body::after`), gradient top border on cards, dot glow, grid-lines background (login page). All are pure CSS, no libraries.

### Expected Features

Full details in FEATURES.md.

**Must have (table stakes):**
- CSS design token system matching mockup palette — `--void`/`--base`/`--raised`/`--card`/`--input` surfaces; `--t1`–`--t4` text; `--b1`–`--b3` borders; `--r1`–`--r4` radius; `--sh1`–`--sh3` shadows; `--ease`/`--spring` easing
- Dark theme pixel-perfect match to all 11 mockups
- Light theme derived from dark tokens with WCAG AA contrast verification
- System theme preference detection (already working via `next-themes`)
- All 11 mockup pages restyled: Login, Dashboard, Messenger, Kanban, Issue Detail, Calls, TMS, Settings, Admin, Onboarding, Trash
- All 15 Shadcn/ui base components updated to new design language
- Spring/bounce entrance animations (400ms, staggered, `--spring` easing)
- Hover micro-interactions: translateY(-2px to -6px), border elevation, shadow bloom

**Should have (differentiators):**
- Animated mesh gradient backgrounds (login: 20s hue-rotate loop; dashboard: ambient glows)
- Glassmorphic sidebar with gradient edge line (`::after` pseudo-element)
- Floating card animations on login page (staggered 7–9s infinite)
- Focus ring glow matching user's accent color
- Gradient CTA buttons on primary actions only (max 1–2 per page)
- Colored dot indicators with `box-shadow` glow on hover
- Top-edge 1px highlight on every card via `::before`
- Monospace task keys (JetBrains Mono, muted color) across all entity IDs

**Defer to polish pass (not blocking):**
- Per-page ambient body gradients (Dashboard: amber glow, Messenger: blue glow)
- Skeleton loading states for all widgets
- Smooth theme toggle transition (remove `disableTransitionOnChange`)
- Dead CSS class cleanup in `app.css`

**Anti-features — explicitly do not build:**
- Route-level page transition animations (AnimatePresence on routes causes layout shift bugs)
- Parallax or scroll-triggered animations (Hub is a productivity tool, not a marketing site)
- 3D transforms / perspective tilt on cards
- Gradient buttons beyond primary CTAs
- Dark-mode-only ship (30–40% of users need light mode)
- `motion.div` wrappers on every element (CSS transitions handle 95% of interactions)

### Architecture Approach

The architecture does not change — only token values and visual styling change. The existing data flow (`next-themes` sets `.dark` on `<html>` → CSS cascade resolves `:root`/`.dark` variables → `@theme inline` maps to Tailwind utilities → all `bg-*`/`text-*`/`border-*` classes update automatically) continues unchanged. See ARCHITECTURE.md for the complete migration sequence.

**Migration sequence:**
1. Design tokens — replace `:root` and `.dark` variable blocks, remap `@theme inline`, add new utility entries (void/base/raised/t1–t4/b1–b3)
2. Typography + fonts — `index.html` font swap, `body` font-family update
3. Accent color remap — keep `[data-accent="X"]` mechanism; update values to match new palette
4. Layout shell — AppShell (topbar always visible on desktop), Sidebar (gradient edge, active glow), Topbar (backdrop blur)
5. UI primitives — Button (gradient variants), Card (gradient bg + top highlight), Input (blue focus glow), Badge (status glow), Dialog (entrance animation)
6. Per-page restyling — 11 mockup pages in user-flow order
7. Unmocked pages — compose from established patterns (Dashboard grid → Projects; Settings forms → Project Settings; Kanban columns → Sprint Backlog)
8. Polish — light mode QA, animation audit, accessibility, dead CSS removal

**Key insight:** Zero new files needed for the design system foundation. All changes modify existing files.

**Patterns to enforce:**
- Surface elevation: always use `--void`/`--base`/`--raised`/`--card`/`--input` — never hardcoded hex or arbitrary `bg-[#...]` classes
- Text hierarchy: `text-t1`/`text-t2`/`text-t3`/`text-t4` for all content
- Border levels: `border-b1`/`border-b2`/`border-b3` based on prominence
- Gradient cards: `bg-gradient-to-b from-card to-raised` + `::before` top highlight + `hover:border-b2`
- Spring animations: `transition: all 250ms var(--spring)` — no inline style overrides

**Major components:**
1. `app.css` — single source of truth for design tokens, typography, animation utilities, and base styles
2. AppShell + Sidebar + Topbar — layout chrome; must be consistent before any page content work
3. 15 Shadcn/ui UI primitives — shared across all pages; restyled once, used everywhere
4. Per-page modules (`src/modules/**/*.tsx`) — pixel-perfect implementation of each mockup

### Critical Pitfalls

Full details with prevention strategies in PITFALLS.md.

1. **56 `!important` declarations in `app.css`** — New styles applied to inputs, selects, and Radix popovers will silently lose to global overrides. Audit and remove all in Phase 1 before any component restyling. Move Radix-specific overrides into individual Shadcn component files using `cn()`.

2. **CSS variable swap must be atomic** — Half-migrated state creates jarring visual mismatches app-wide. Swap all `:root`/`.dark` variable blocks in a single commit. Accept the global ripple; fix individual components page-by-page in subsequent phases.

3. **JSX restructuring breaks interactive behavior** — Changing DOM structure to match mockups silently detaches Kanban drag-and-drop, Messenger infinite scroll, Radix focus trapping, and form chains. Rule: restyle with CSS classes first; change JSX only when required, and verify all interactions afterward.

4. **Radix primitive breakage from Shadcn restyling** — Removing `data-*`, `aria-*`, `asChild`, `forwardRef`, or `Portal` props when restyling Shadcn components causes dropdowns that don't close, dialogs without focus trapping, and broken keyboard navigation. Only modify `className` props — never behavioral props.

5. **Light mode accessibility failure** — Mechanically inverted dark palette will fail WCAG AA contrast (4.5:1 for text, 3:1 for UI components) on first attempt. Run `axe-core` audit immediately after defining light tokens. Zero violations required before merging Phase 1.

---

## Implications for Roadmap

Based on combined research, the redesign maps to 6 phases. The critical constraint is dependency ordering: tokens must exist before components, shared chrome (sidebar/topbar) must be done before page content, dark theme must be stable before light theme visual QA.

### Phase 1: Design System Foundation

**Rationale:** CSS variables are the single point of control for the entire visual layer. Changing them first lets the cascade do most of the work automatically. The `!important` cleanup gates everything — any subsequent component work hits the specificity wall without it. Light theme tokens belong here too: defining them after the dark theme is "stable" is a trap that forces rework when dark tokens change.
**Delivers:** Updated `app.css` with new `:root`/`.dark` variable blocks; entire app shifts to new palette; Inter/JetBrains Mono fonts loaded; new semantic utilities available (`bg-void`, `text-t3`, `border-b2`, etc.); zero `!important` declarations; light theme defined with verified WCAG AA contrast; z-index scale tokens; Firefox scrollbar fix; easing variable definitions (`--ease`, `--spring`).
**Addresses:** Design token system, font swap, noise overlay, CSS easing variables, radius/shadow systems (FEATURES.md table stakes)
**Avoids:** Pitfall 1 (important minefield), Pitfall 3 (CSS variable atomic swap), Pitfall 5 (light mode accessibility), Pitfall 8 (global token ripple), Pitfall 11 (accent system incompatibility), Pitfall 15 (z-index wars)

### Phase 2: Layout Shell

**Rationale:** Sidebar and topbar frame every single page. Until the chrome matches the new design language, every page looks partially broken regardless of how well individual content is styled. This phase also validates that the token system from Phase 1 works end-to-end in real component context before investing in page-level work.
**Delivers:** Sidebar with gradient edge line (`::after`), active nav item glow, new `--sidebar-bg` background, updated typography; Topbar always visible on desktop with `backdrop-filter: blur(24px)` and gradient bottom edge; AppShell content padding updated to 28px.
**Avoids:** Pitfall 2 (JSX breakage — sidebar collapse/expand must keep working at all viewport sizes), Pitfall 10 (inconsistent UI during migration)

### Phase 3: Shared UI Primitives

**Rationale:** 15 Shadcn/ui components appear across all pages. Restyling them before page-level work means every page starts from a correct visual foundation. With `!important` gone from Phase 1, component-level styling is clean and predictable.
**Delivers:** Button (gradient primary, ghost/secondary variants, spring easing, glow states), Card (gradient bg + top-edge highlight + hover glow), Input (blue focus glow ring + scale 1.005 on focus), Badge (status glow backgrounds using `--gB`/`--gR` etc.), Dialog/Popover (entrance animations + backdrop blur), Tabs, DropdownMenu, ScrollArea (5px slim scrollbar), Avatar (online indicator glow), Tooltip.
**Avoids:** Pitfall 4 (Radix primitive breakage), Pitfall 6 (animation performance — CSS transitions for hover, `motion` only for layout and mount)

### Phase 4: Core Mockup Pages

**Rationale:** Login is standalone and great for early validation of the token system visually. Dashboard contains the most reused widget patterns and must be done early so unmocked pages can reference it directly. Messenger and Kanban are the highest-complexity, highest-frequency pages — they are the heart of the product.
**Delivers:** Login (split layout, animated mesh, floating cards, spring entrance); Dashboard (widget grid, greeting gradient, quick actions, staggered entrance); Messenger (channel list, message bubbles, composer glassmorphism); Kanban (column headers, card hover states, priority dots, WIP indicators).
**Avoids:** Pitfall 2 (full per-page interactive behavior checklist for Kanban drag-and-drop, Messenger scroll and Plate editor, Calls LiveKit UI), Pitfall 6 (Kanban 50+ card animations must be CSS-only via `animation-delay` stagger), Pitfall 13 (Plate editor and LiveKit style leaks from global CSS)

### Phase 5: Remaining Mockup Pages + Unmocked Pages

**Rationale:** The remaining 7 mockup pages (Issue Detail, Calls, TMS, Settings, Admin, Onboarding, Trash) have established component foundations from Phases 3–4. Unmocked pages come last because they are composed from patterns already proven — no design invention required, just systematic pattern application.
**Delivers:** All 11 mockup pages pixel-perfect; 6 unmocked pages (Projects List, Sprint Backlog, Sprint Detail/Burndown, Planning Poker, JQL Search, Project Settings/Dashboard) styled consistently with the design system using established pattern mappings from FEATURES.md.
**Avoids:** Pitfall 9 (mobile regression — check 375px viewport after each page), Pitfall 13 (third-party component leaks in Calls/Messenger)

### Phase 6: Polish and QA

**Rationale:** Visual completion is not production readiness. Animation timing audit, light mode visual QA across all pages, accessibility audit, and dead CSS removal require a dedicated phase — they cannot be squeezed into page phases without cutting corners.
**Delivers:** Animation timing audit (60fps at 4x CPU throttle on Dashboard and Kanban); Lighthouse accessibility zero contrast violations in both themes; mobile QA at 375px/768px/1024px; smooth theme toggle transition (200ms background/color transition on key surfaces); dead CSS class removal from `app.css`; final WCAG AA verification pass; per-page ambient gradient backgrounds (polish cap).
**Avoids:** Pitfall 5 (light mode accessibility), Pitfall 6 (animation performance), Pitfall 9 (mobile regression), Pitfall 16 (dead CSS accumulation)

### Phase Ordering Rationale

- **Tokens before components** (Phases 1→3): 60–70% of visual change happens automatically through the cascade. Inverting this order means every component invents its own color values and must be redone when tokens are finalized.
- **Shell before pages** (Phase 2 before 4): Sidebar and topbar frame every page. Mismatched chrome makes every page look broken regardless of content quality.
- **`!important` cleanup gates everything**: Phase 2+ cannot reliably start until the specificity minefield is cleared in Phase 1.
- **Dark tokens before light visual QA** (light tokens defined in Phase 1, visual QA in Phase 6): Light theme token values are defined in Phase 1 alongside dark, but visual QA of light mode is deferred to Phase 6 when all pages are stable and not changing.
- **Mockup pages before unmocked pages** (Phase 4–5): Unmocked pages are composed from mockup patterns. Doing them first produces inconsistent extrapolations that must be redone.

### Research Flags

Phases where execution is straightforward (standard patterns, no additional research needed):
- **Phase 1:** CSS variable architecture is fully documented in ARCHITECTURE.md with exact values from mockup analysis. Implementation is mechanical — direct value substitution.
- **Phase 2:** Sidebar and topbar structures identified from mockup analysis. Scope is clearly bounded.
- **Phase 3:** All 15 Shadcn/ui components are well-understood. Restyling patterns via `cn()` and CSS variable overrides are thoroughly documented in Shadcn's own docs.
- **Phase 6:** Performance profiling (Chrome DevTools), accessibility auditing (axe-core), and mobile QA (DevTools responsive mode) are standard verification workflows.

Phases requiring execution-time vigilance (not research, but careful discipline):
- **Phase 4 — Kanban:** pragmatic-drag-and-drop is sensitive to DOM structure changes. Write smoke tests before restyling if none exist. Verify drag behavior after every JSX modification.
- **Phase 4 — Messenger:** Plate editor styling isolation requires careful scoping after `!important` removal. Investigate `PlateEditor` container class structure before work begins.
- **Phase 5 — Calls:** LiveKit components render their own DOM. Test call UI after any global button/avatar style change to catch style leaks early.

---

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Direct inspection of `package.json`, `app.css`, `index.html`, and all 11 mockup CSS blocks. Zero new packages confirmed from two independent angles. |
| Features | HIGH | All 11 mockups analyzed directly. Feature set is bounded and explicit — this is implementation, not product discovery. Anti-features are clear from scope constraints. |
| Architecture | HIGH | `app.css` (1053 lines), all 15 Shadcn components, and AppShell/Sidebar/Topbar inspected directly. Migration sequence verified against Tailwind v4 `@theme inline` behavior. Variable mapping tables are concrete, not theoretical. |
| Pitfalls | HIGH | 56 `!important` declarations counted directly with grep. 235 source files analyzed for interactive behavior risk. Community and official sources agree on Radix breakage patterns and CSS variable strategies. |

**Overall confidence:** HIGH

### Gaps to Address

- **Light theme exact values:** ARCHITECTURE.md provides a complete derived light palette (`--void: #FFFFFF`, `--base: #F8F9FA`, etc.) as first-pass values. Contrast ratios must be verified with `axe-core` during Phase 1 execution. Muted text values (`--t3`, `--t4`) on light backgrounds are most likely to need adjustment.

- **Font loading strategy — CDN vs self-hosted:** STACK.md recommends Google Fonts CDN (matches current pattern, internal tool). PITFALLS.md recommends self-hosted WOFF2 with `size-adjust` fallback for CLS prevention. Decide at Phase 1 start; CDN is acceptable for an internal 10–50 user tool but document the decision.

- **Unmocked page layouts:** 6 unmocked pages have pattern mappings in FEATURES.md but no explicit designs. The first 2–3 unmocked pages in Phase 5 will establish the reference pattern — allocate extra review time there to prevent divergence.

- **Plate editor dark theme isolation:** Plate rich-text editor may require additional CSS scoping once global `input`/`textarea` overrides are removed. Investigate `PlateEditor` container class structure before Phase 4 Messenger work begins.

---

## Sources

### Primary (HIGH confidence — direct codebase analysis)
- `frontend/src/app.css` (1053 lines) — current design token system, 56 `!important` declarations, custom utility classes
- `frontend/package.json` — confirmed `motion ^12.38.0`, `next-themes ^0.4.6`, no missing packages
- `frontend/index.html` — Google Fonts CDN (Geist, confirmed needs swap to Inter)
- `redesign/01-login.html` through `redesign/11-trash.html` — all 11 mockups analyzed for CSS tokens, fonts, animations, layout patterns
- `frontend/src/components/ui/*.tsx` — all 15 Shadcn/ui components analyzed for Radix prop surface
- `frontend/src/components/layout/app-shell.tsx`, `sidebar.tsx`, `topbar.tsx` — layout structure analysis

### Secondary (HIGH confidence — official documentation)
- [Shadcn/ui Theming](https://ui.shadcn.com/docs/theming) — CSS variable theming, `@theme inline` integration
- [Shadcn/ui Dark Mode](https://ui.shadcn.com/docs/dark-mode) — `next-themes` with `.dark` class
- [Tailwind CSS v4 Migration Guide](https://tailwindcss.com/blog/tailwindcss-v4) — `@theme inline` config, utility generation
- [Motion Performance Guide](https://motion.dev/docs/performance) — compositor-safe properties, layout animation costs
- [WCAG Dark/Light Mode](https://blog.logrocket.com/how-do-you-implement-accessible-linear-design-across-light-and-dark-modes/) — 4.5:1 AA ratio requirements

### Secondary (MEDIUM confidence — community and design sources)
- [Linear UI Redesign (Part II)](https://linear.app/now/how-we-redesigned-the-linear-ui) — design system approach reference
- [Color Tokens Guide](https://medium.com/design-bootcamp/color-tokens-guide-to-light-and-dark-modes-in-design-systems-146ab33023ac) — semantic token mapping between themes
- [SaaS Design Trends 2026](https://jetbase.io/blog/saas-design-trends-best-practices) — micro-interactions as premium UX standard
- [Skeleton Loading Design](https://blog.logrocket.com/ux-design/skeleton-loading-screen-design/) — skeleton structure matching content layout

---
*Research completed: 2026-03-27*
*Ready for roadmap: yes*
