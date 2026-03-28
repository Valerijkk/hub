# Feature Landscape: Hub v2.0 UI Redesign

**Domain:** Premium dark UI redesign of existing SaaS platform (11 HTML mockups)
**Researched:** 2026-03-27
**Confidence:** HIGH

---

## Context

Hub v2.0 is a visual-only redesign. All 11 pages are fully functional. The goal is pixel-perfect implementation of HTML mockups (dark theme), plus light theme derivation, consistent extrapolation of unmocked pages, and premium interaction quality. No new features, no backend changes.

The mockups define a specific design language: 5-layer dark palette (#08090B to #1E2028), Inter + JetBrains Mono typography, spring/bounce easing, noise texture overlays, gradient mesh backgrounds, glassmorphic sidebar, and subtle glow effects on active elements.

---

## Table Stakes

Features users expect from a premium UI redesign. Missing any = redesign feels amateur or half-done.

### Design System Foundation

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| CSS variable-based design tokens | Every modern design system (Linear, Vercel, shadcn) uses semantic tokens. Hardcoded colors = unmaintainable | MEDIUM | Replace current HSL tokens with mockup palette. Semantic naming: --void, --base, --raised, --card, --input for surfaces; --t1 through --t4 for text; --b1 through --b3 for borders |
| Consistent border radius system | Mockups use 6/8/12/16px + pill. Current code uses `--radius: 0.75rem` with calc derivatives | LOW | Map to --r1 (6px), --r2 (8px), --r3 (12px), --r4 (16px), --pill (100px). Apply consistently across all shadcn components |
| Shadow elevation system | Mockups define 3 shadow levels (sh1/sh2/sh3). Current dark theme has no shadow hierarchy | LOW | --sh1 subtle (1px 3px), --sh2 medium (4px 16px), --sh3 heavy (12px 40px). All use rgba(0,0,0) not rgba(white) for dark mode |
| Typography scale with Inter + JetBrains Mono | Mockups use specific weight/size/tracking combos. Current code loads Inter but inconsistently | LOW | Font weights: 400/500/600/700/800/900. Code elements use JetBrains Mono. Letter-spacing: tight (-0.04em) for headings, normal for body |
| Noise texture overlay | Every mockup page includes a fractal noise SVG overlay at 3% opacity. This creates the "premium matte" feel | LOW | Single `body::after` with fixed SVG noise. Pointer-events: none. Already in mockups as inline SVG data URI |
| Ambient gradient backgrounds | Mockups use subtle radial gradients on body::before (colored glows at edges). Different per page | LOW | Per-page ambient gradients. Dashboard: amber/blue. Messenger: blue/purple. Login: blue/purple/green mesh |

### Theme Switching (Light + Dark)

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Dark theme matching mockups exactly | The mockups ARE the dark theme. Pixel-perfect implementation is the entire point | HIGH | This is the bulk of the work. Every component must match mockup colors, spacing, borders, shadows |
| Light theme derived from dark | Users expect light/dark toggle in 2026. Every premium SaaS (Linear, Vercel, Notion, Figma) offers it | MEDIUM | No light mockups exist. Must derive systematically: invert surface hierarchy (white to light grays), keep accent colors, adjust shadows to rgba(0,0,0,0.05-0.15), adjust border opacity |
| System theme preference detection | OS-level dark/light preference respected by default. `prefers-color-scheme` is standard | LOW | Already implemented via next-themes with `enableSystem`. Keep this behavior |
| No flash on page load (FOUC prevention) | Theme flash on load looks broken. Users notice immediately | LOW | next-themes handles this with a script tag injected before render. Current `disableTransitionOnChange` is correct for initial load but needs review for theme toggle |
| Smooth theme transition on toggle | When user clicks dark/light toggle, colors should transition smoothly (not instant swap) | LOW | Remove `disableTransitionOnChange` from ThemeProvider. Add `transition: background-color 200ms, color 200ms, border-color 200ms` to key elements. Exclude initial load |
| Accent color preservation | Current system supports 5 accent colors (red, blue, green, purple, orange). Must work in both themes | MEDIUM | Accent colors already exist via data-accent attribute. Ensure all 5 work with new dark palette AND derived light palette. May need separate accent opacity values per theme |

### Per-Page Pixel-Perfect Implementation (11 Mockup Pages)

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Login page redesign | Split layout (55% animated gradient + 45% form). Floating feature cards. Gradient mesh background | MEDIUM | Major visual change from current login. Code verification input (OTP style). Telegram button. Spring entrance animation |
| Dashboard redesign | 3-column widget grid. Gradient greeting header. Quick action buttons with hover lift. Task items with colored dots and monospace keys | HIGH | Most complex layout. Multiple widget types: tasks, activity, charts, deadlines, time tracking. Staggered entrance animations |
| Messenger redesign | Channel list sidebar. Message bubbles. Thread panel. Typing indicators. Glassmorphic elements | HIGH | Must preserve all existing messenger functionality (WebSocket, threads, reactions, file sharing) while reskinning |
| Kanban board redesign | Column headers with counts. Card design with priority dots, labels, avatars. Drag ghost styling | HIGH | Must work with pragmatic-drag-and-drop. Column backgrounds, card hover states, WIP limit indicators |
| Issue Detail redesign | Two-panel layout (content + sidebar). Rich text area. Activity timeline. Relation chips. Status/priority selectors | HIGH | Complex page with many sub-components. Must coordinate with Plate editor styling |
| Calls page redesign | Participant grid. Control bar. Screen share panel. Invite block (when alone) | MEDIUM | Must work with LiveKit components. Audio visualizations. Glassmorphic controls |
| TMS page redesign | Test case tree. Run execution panel. Status badges. Folder hierarchy | HIGH | Tree view with expandable nodes. Step editor. Result recording UI |
| Settings page redesign | Form sections. Toggle switches. Profile fields. Theme/accent preview | MEDIUM | Standard form layout. Must include theme toggle and accent color picker preview |
| Admin page redesign | Stats cards. User management table. System health indicators | MEDIUM | Data tables with sorting. KPI cards with trend indicators |
| Onboarding page redesign | Step wizard. Profile setup. Tour highlights | MEDIUM | Multi-step flow. Progress indicators. Welcome animations |
| Trash page redesign | Deleted items list. Restore actions. Retention countdown | LOW | Simple list with filters. Restore/permanent-delete actions |

### Animation Quality

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Entrance animations on page load | Mockups specify `enterUp` keyframes (opacity 0 + translateY 16px to visible). Staggered per element | LOW | Use motion (ex framer-motion) `AnimatePresence` + staggered delays. 400ms duration with spring easing |
| Hover state transitions | Every interactive element needs smooth hover. Mockups show: translateY(-2px to -6px), border-color change, box-shadow elevation, background shift | LOW | 200-280ms transitions. Use `var(--ease): cubic-bezier(.25,.1,0,1)` from mockups. Cards lift, buttons scale |
| Active/pressed state feedback | Buttons: `scale(.97)`. Cards: quick press response. Links: color transition | LOW | 150ms max for press feedback. Matches mockup `btn:active{transform:scale(.97)}` |
| Spring easing curves | Mockups define `--spring: cubic-bezier(.175,.885,.32,1.1)` and `--ease: cubic-bezier(.25,.1,0,1)`. These create the "snappy but organic" feel | LOW | Register as CSS custom properties. Use in all transitions. spring has overshoot (1.1), ease is smooth deceleration |
| Loading skeletons | Users expect skeleton screens during data fetches, not spinners. Linear and Vercel set the standard | MEDIUM | Create skeleton variants for each widget type. Match dark theme surface colors (--raised pulsing to --card). Wave animation preferred over pulse |
| Sidebar active item glow | Mockups show active sidebar item with amber left border + glow shadow + gradient background leak | LOW | CSS-only: `box-shadow: 0 0 12px var(--amber)` on `::before` pseudo-element. Gradient overlay via `::after` |

### Component Library Updates

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Updated shadcn/ui base components | All 15 existing UI components (button, card, dialog, etc.) must use new design tokens | MEDIUM | Restyle each component. Focus on: button gradients (not flat), input focus glow rings, card glass effects, dialog backdrop blur |
| Custom scrollbar styling | Mockups define thin 5px scrollbars with transparent track and subtle thumb | LOW | Already partially in mockups CSS. Apply globally via `::-webkit-scrollbar` rules |
| Backdrop blur on overlays | Topbar uses `backdrop-filter: blur(24px) saturate(180%)`. Popover/dialog should follow | LOW | Apply to topbar, dropdown menus, command palette, dialogs. Note: Firefox performance consideration (now supported since FF 103) |
| Gradient text for headings | Mockups use `background-clip: text` gradients on key headings (greeting, page titles, logo) | LOW | Create utility class or component. Per-page gradient directions: amber-orange for dashboard, blue-purple for general |

---

## Differentiators

Features that elevate the redesign from "competent" to "premium." Not expected, but immediately noticed.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| **Mesh gradient animated backgrounds** | Login page has a 20s animated mesh with hue-rotate. Dashboard has ambient colored glows. This creates depth that flat backgrounds lack. Linear, Raycast, and Arc browser use this technique | LOW | CSS-only: radial-gradient layers + keyframe hue-rotate/brightness. Different mesh per page section. Performance: GPU-composited, no reflows |
| **Glassmorphic sidebar** | Sidebar has `::after` gradient line (amber to blue), hover states with `var(--hover)`, and glowing active indicators. Not just a dark panel but a layered glass effect | LOW | Already defined in mockups. Key elements: 1px gradient border, glow on active nav item, logo with box-shadow glow, online status dot glow |
| **Floating card animations** | Login page feature cards float with staggered 7-9s infinite animations. Dashboard widgets lift on hover with shadow elevation change | LOW | CSS @keyframes for float. motion for hover lift. Stagger with nth-child delays |
| **Focus ring glow effect** | Input focus shows blue glow ring: `box-shadow: 0 0 0 3px rgba(91,156,245,.08), 0 0 20px rgba(91,156,245,.04)` plus slight scale(1.005). More premium than a flat outline | LOW | Apply to all focusable elements. Coordinate with accent color system so focus ring matches user's chosen accent |
| **Gradient CTA buttons** | Login/primary buttons use `linear-gradient(135deg, var(--blueS), #2563EB)` with inner light overlay via `::before`. Hover adds deeper box-shadow and slight lift | LOW | Reserved for primary CTAs only. Secondary buttons stay flat. Creates clear visual hierarchy |
| **Colored dot indicators with glow-on-hover** | Task status dots, project indicators, and online status all have `box-shadow: 0 0 8px currentColor` on hover. Subtle but polished | LOW | CSS transition on box-shadow. Use `currentColor` for automatic color matching. Applied to: task priority dots, sidebar project dots, user online status |
| **Monospace task keys** | Issue keys (HUB-123) use JetBrains Mono with reduced font size and muted color. Gives a technical, developer-tool aesthetic | LOW | Already in mockups. Apply to all entity IDs across all modules |
| **Top border light line on cards** | Every card/widget has a 1px top highlight: `linear-gradient(90deg, transparent, rgba(255,255,255,.04), transparent)` via `::before`. Simulates light source from above | LOW | Single CSS rule on all card components. Creates 3D depth perception without actual 3D transforms |
| **Page-specific ambient gradients** | Each page has unique subtle colored radial gradients on body::before. Dashboard: amber glow top-right. Messenger: blue glow. This creates visual identity per section | LOW | Set via CSS class on body/main container per route. Very subtle (2-3% opacity) so it's felt not seen |
| **Sidebar gradient separator line** | Right edge of sidebar has a gradient line: `linear-gradient(180deg, transparent 5%, rgba(amber,.05) 30%, rgba(blue,.04) 60%, transparent 95%)` instead of a flat border | LOW | CSS `::after` pseudo-element. More alive than `border-right: 1px solid` |

---

## Anti-Features

Features to explicitly NOT build during the redesign.

| Anti-Feature | Why Tempting | Why Avoid | What to Do Instead |
|--------------|-------------|-----------|-------------------|
| **Full custom component library** | "Let's build everything from scratch for maximum control" | shadcn/ui components are already integrated and well-tested. Rewriting introduces bugs and increases scope 5x | Restyle existing shadcn/ui components via CSS variable overrides. Only create new components for mockup-specific patterns not in shadcn |
| **Complex page transition animations** | "Linear has smooth page transitions" | Route-level animations with AnimatePresence add complexity to routing, cause layout shift bugs, and interact poorly with TanStack Router | Entrance animations within pages (staggered fade-up). No cross-page transitions. Pages load fast enough that transitions are unnecessary |
| **Parallax/scroll-triggered animations** | "Modern websites have scroll animations" | Hub is a productivity tool, not a marketing site. Scroll animations slow down workflows and add motion sickness risk | Static layout with hover micro-interactions only. Respect `prefers-reduced-motion` |
| **3D transforms / perspective effects** | "Cards could tilt on hover like Stripe" | GPU-intensive, causes rendering issues on lower-end machines, distracting in a daily-use work tool | Subtle 2D transforms only: translateY, scale. Maximum lift: -6px (per mockups) |
| **Custom icon set** | "Let's design bespoke icons for Hub" | Weeks of design work for diminishing returns. Lucide icons (used by shadcn) are comprehensive and consistent | Keep Lucide icons. Mockups use inline SVGs that map to standard Lucide equivalents |
| **Animated gradient buttons everywhere** | "The login CTA gradient looks great, use it on all buttons" | Gradient overuse cheapens the effect. Creates visual noise in data-dense pages | Gradient buttons for primary CTAs only (max 1-2 per page). All other buttons: flat with hover background change |
| **Dark-mode-only ship** | "Mockups are dark, just ship dark" | ~30-40% of users prefer light mode. Shipping without light mode causes immediate complaints and accessibility issues (some users need high-contrast light) | Derive light theme systematically from dark tokens. Ship both from day one |
| **Per-component animations with motion** | "Wrap every element in motion.div" | Bundle size bloat, render overhead, maintenance burden. Most components need simple CSS transitions, not JS-driven animations | Use CSS transitions for hover/focus/active states (95% of interactions). Reserve motion for: entrance animations, layout shifts (AnimatePresence), drag-and-drop ghost |
| **Theme transition animation on every element** | "Everything should smoothly cross-fade on theme toggle" | Applying transition to all elements causes jank on theme switch (hundreds of elements transitioning simultaneously) | Transition only key surfaces: body background, sidebar background, card backgrounds, text color. Let borders and small elements snap instantly |

---

## Feature Dependencies

```
Design System Foundation (tokens, variables, fonts)
    |
    +---> Shadcn/ui Component Restyling (button, card, input, dialog...)
    |         |
    |         +---> Page-level Redesigns (use restyled components)
    |
    +---> Light Theme Derivation (need dark tokens finalized first)
    |         |
    |         +---> Theme Toggle Polish (needs both themes working)
    |
    +---> Animation System (easing curves, entrance keyframes)
              |
              +---> Page-level Animations (per-page entrance, hover states)

Pages without mockups (depend on design system + at least 3 mockup pages done):
    Design System + Sidebar + Topbar + Widget patterns from Dashboard
        |
        +---> Projects List (reuse card grid from Dashboard)
        +---> Project Overview (reuse widget grid from Dashboard)
        +---> Sprint Management (reuse Kanban column patterns + Dashboard widgets)
        +---> JQL Search (reuse input + filter patterns from Kanban/TMS)
        +---> Project Settings (reuse form patterns from Settings)
        +---> Project Dashboard (reuse widget grid from Dashboard)
```

### Critical Path

1. **Design tokens** must be finalized before ANY component work
2. **Sidebar + topbar** must be done early because they appear on EVERY page
3. **Dashboard** should be first full page (most widget patterns reused elsewhere)
4. **Settings** should be done before unmocked pages (form patterns reused in Project Settings)
5. **Light theme** should be derived AFTER dark theme is stable (changing dark tokens forces light rework)

---

## Pages Without Mockups: Extrapolation Strategy

Six pages need consistent styling without explicit mockups. The approach: compose from patterns already defined in mockup pages.

### Pattern Mapping

| Unmocked Page | Primary Pattern Source | Secondary Source | Key Components to Reuse |
|---------------|----------------------|------------------|------------------------|
| **Projects List** | Dashboard (widget grid, cards) | Sidebar (project dots, colors) | Card grid, hover lift, colored dots, search/filter bar |
| **Project Overview** | Dashboard (widget grid, stats) | TMS (tree/hierarchy display) | Widget grid, stat cards, activity feed, progress bars |
| **Sprint Backlog** | Kanban (column layout, cards) | Dashboard (task items) | Issue cards, drag-and-drop columns, filter bar, priority dots |
| **Sprint Detail/Burndown/Velocity** | Dashboard (chart widgets) | Admin (stat cards) | Tremor chart styling with dark tokens, stat KPI cards, date range selectors |
| **Planning Poker** | Calls (participant grid) | Dashboard (action buttons) | Participant cards, reveal animation, voting buttons, Fibonacci scale display |
| **JQL Search** | TMS (filter panel, results list) | Messenger (search bar) | Code-style input (JetBrains Mono), results table, filter chips, saved queries list |
| **Project Settings** | Settings (form layout) | Admin (table layout) | Form sections, toggle switches, role permission table, workflow editor |
| **Project Dashboard** | Dashboard (widget grid) | Admin (stats) | Same widget system as personal dashboard, project-scoped data |

### Design Rules for Unmocked Pages

1. **Use --base as page background** (not --void, which is for body/root only)
2. **56px topbar** with glassmorphic blur, gradient text page title, same breadcrumb pattern
3. **Widget grid**: `grid-template-columns: repeat(3, 1fr)` with 16px gap, same card styling (gradient background, 1px top highlight, hover lift)
4. **Data tables**: Follow Admin page table patterns (row hover, alternating subtle backgrounds, monospace IDs)
5. **Form sections**: Follow Settings page patterns (labeled sections, --input background on fields, blue focus glow ring)
6. **Entrance animations**: Same staggered fade-up pattern: `animation: eu 400ms var(--spring) [delay]ms both`
7. **Empty states**: Centered muted text with icon, following the same muted color (--t3) pattern

---

## MVP Recommendation for Redesign Phases

### Phase 1: Foundation (do first)
1. Design token system (CSS variables matching mockup palette)
2. Shadcn/ui component restyling (all 15 components)
3. Sidebar + topbar (shared layout, appears everywhere)
4. Animation utilities (easing curves, entrance keyframes, hover patterns)

### Phase 2: Core Pages (highest visual impact)
1. Login (standalone page, great for validation)
2. Dashboard (most patterns, reused by unmocked pages)
3. Messenger (most used daily, complex)
4. Kanban (visual centerpiece of tracker)

### Phase 3: Remaining Mockup Pages
1. Issue Detail (complex, many sub-components)
2. TMS (tree views, execution panels)
3. Calls (LiveKit component restyling)
4. Settings + Admin + Onboarding + Trash

### Phase 4: Unmocked Pages + Light Theme
1. Derive light theme from finalized dark tokens
2. Apply design system to all unmocked pages
3. Theme toggle polish (smooth transitions, no FOUC)
4. Skeleton loading states for all pages

**Defer:** Complex per-page ambient gradients can be added as polish after all pages are functionally restyled.

---

## Sources

- [How we redesigned the Linear UI (part II)](https://linear.app/now/how-we-redesigned-the-linear-ui) - Linear's design system approach with LCH color space and minimal token generation
- [A calmer interface for a product in motion - Linear](https://linear.app/now/behind-the-latest-design-refresh) - Linear's latest design refresh philosophy
- [Dark Mode Web Design for Better SaaS UX](https://webspidersolutions.com/dark-mode-web-design-saas/) - Dark mode not a palette swap, requires intentional secondary interface design
- [Dark Mode UI: Best Practices 2025](https://www.graphiceagle.com/dark-mode-ui/) - Avoid pure black, use dark grays, proper contrast ratios
- [Theming - shadcn/ui](https://ui.shadcn.com/docs/theming) - Official shadcn CSS variable theming system
- [Dark Mode - shadcn/ui](https://ui.shadcn.com/docs/dark-mode) - Official dark mode implementation with next-themes
- [Skeleton loading screen design - LogRocket](https://blog.logrocket.com/ux-design/skeleton-loading-screen-design/) - Skeleton layouts should match final content structure
- [Micro-interaction ideas with Framer Motion](https://siadesign.ee/en/blog/micro-interactions-framer-motion/) - 200-500ms timing for micro-animations, whileInView for entrance, whileHover for feedback
- [Advanced animation patterns with Framer Motion](https://blog.maximeheckel.com/posts/advanced-animation-patterns-with-framer-motion/) - Layout animations and spring physics
- [Color tokens: guide to light and dark modes](https://medium.com/design-bootcamp/color-tokens-guide-to-light-and-dark-modes-in-design-systems-146ab33023ac) - Semantic token mapping between light and dark themes
- [Designing a Scalable and Accessible Dark Theme](https://www.fourzerothree.in/p/scalable-accessible-dark-mode) - WCAG compliance in dark mode, readable + elegant + scalable
- [SaaS Design Trends 2026](https://jetbase.io/blog/saas-design-trends-best-practices) - Subtle animations and micro-interactions as essential tools for premium UX
- Hub HTML mockups (`redesign/01-login.html` through `redesign/11-trash.html`) - Primary design source, analyzed directly

---
*Feature research for: Hub v2.0 UI Redesign*
*Researched: 2026-03-27*
