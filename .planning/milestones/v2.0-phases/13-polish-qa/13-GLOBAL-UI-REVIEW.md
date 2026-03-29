# Phase 13 -- Global Cross-Cutting UI Review

**Audited:** 2026-03-28
**Baseline:** Hub Design System v2.0 (app.css tokens) + HTML mockup references
**Screenshots:** Not captured (no dev server detected on ports 3000, 5173, 8080)
**Scope:** Full frontend audit -- 145+ TSX files across routes, modules, components, UI primitives

---

## Pillar Scores

| Pillar | Score | Key Finding |
|--------|-------|-------------|
| 1. Copywriting | 3/4 | Russian copy is contextual and specific; no generic "Click Here" labels found; "Submit" usage is all internal handler names, not user-facing |
| 2. Visuals | 3/4 | Strong visual hierarchy via gradient cards, glow buttons, and entrance animations; icon-only buttons mostly have aria-labels (29 occurrences across 14 files) |
| 3. Color | 2/4 | 110+ hardcoded hex colors in TSX files; 403 legacy Shadcn token usages remain across 83 files; design system token adoption is partial |
| 4. Typography | 3/4 | Consistent use of design system font stack; font sizes use both Tailwind scale and arbitrary values (968 arbitrary bracket values); weight distribution is reasonable |
| 5. Spacing | 2/4 | 968 arbitrary bracket values for spacing/sizing across 115 files; mockup-matched pages use precise px values but unmocked pages mix Tailwind scale and arbitrary |
| 6. Experience Design | 3/4 | Loading states in 66 files; empty states in 76 files; error boundaries in 2 files; disabled states in 71 files; reduced-motion handled globally in app.css; 190 transition-all usages in 70 files |

**Overall: 16/24**

---

## Top 3 Priority Fixes

1. **Legacy Shadcn tokens persist in 83 files (403 occurrences)** -- Users will see visual inconsistencies between redesigned and unredesigned components, especially in light theme where `bg-muted` and `text-muted-foreground` resolve to different values than `bg-raised` and `text-t2`. -- Search-and-replace `text-muted-foreground` with `text-t2`, `bg-muted` with `bg-raised`, `bg-background` with `bg-base`, `text-foreground` with `text-t1` across all TSX files. Priority targets: TMS module (153 occurrences in 14 files), messenger components (89 occurrences in 21 files), dashboard widgets (31 occurrences in 12 files).

2. **190 `transition-all` usages across 70 files hurt performance** -- `transition-all` forces the browser to watch every CSS property for changes, causing unnecessary composite layer creation. This is especially impactful on kanban boards, calendars, and list views with many animated elements. -- Replace `transition-all` with specific property transitions: `transition-colors` for color-only changes, `transition-[transform,box-shadow,border-color]` for interactive cards, `transition-[transform,opacity]` for entrance effects. The button.tsx primitive already uses `transition-all` and should be the first fix as it propagates everywhere.

3. **110+ hardcoded hex colors in TSX files bypass the design system** -- Hardcoded colors will not respond to theme changes or accent color selection. Worst offenders: burndown-chart.tsx (11 hex values for chart colors), velocity-chart.tsx (7 hex values), login.tsx (4 hex values for feature card icons), onboarding.tsx (6 hex values). -- For chart libraries (Recharts), create CSS custom properties for chart colors (e.g., `--chart-line`, `--chart-grid`) and reference them via `getComputedStyle`. For inline styles, replace hex values with `var(--blue)`, `var(--t3)`, etc.

---

## Detailed Findings

### Pillar 1: Copywriting (3/4)

**Strengths:**
- All user-facing copy is in Russian, contextual, and domain-specific
- No generic labels like "Click Here", "OK", or "Cancel" found in user-facing text
- "Submit" usage is exclusively internal handler/variable names (`handleSubmit`, `isSubmitting`, `onSubmit`), not displayed to users
- Button labels are action-specific: "Создать" (Create), "Сохранить" (Save), "Удалить" (Delete)
- Empty states are present with contextual messaging in 76 files

**Issues:**
- Some error/loading states use terse text: `components/tracker/activity-feed.tsx:49` shows "Загрузка активности..." and `line:54` shows "Нет активности" -- these could be more descriptive
- `components/files/dropzone.tsx:44` uses bare "Загрузка..." without context of what is loading

### Pillar 2: Visuals (3/4)

**Strengths:**
- Clear visual hierarchy through gradient card backgrounds (`bg-gradient-to-b from-card to-raised`), glow button effects, and the `::before` highlight line pattern used consistently across card components
- 29 `aria-label` attributes across 14 files -- mostly on icon-only call controls and sidebar toggles
- Entrance animations (`dash-enter`, `kanban-enter`, `msg-enter-x`, `animate-fade-in`) provide consistent motion language
- The design system in app.css defines a complete z-index scale preventing stacking conflicts

**Issues:**
- Several icon-only buttons in messenger components lack aria-labels (message-actions.tsx, reaction-picker.tsx)
- The `__root.tsx` error/404 pages (lines 28-90) still use legacy Shadcn visual patterns (`bg-background`, `text-foreground`, `text-muted-foreground`) and lack the gradient card treatment used elsewhere
- Some interactive elements in TMS module lack visual feedback (hover states missing on some table rows)

### Pillar 3: Color (2/4)

**Critical: Legacy Token Migration Incomplete**

The design system defines a clear token architecture in app.css with compatibility aliases:
- `--color-muted-foreground` maps to `var(--t2)`
- `--color-background` maps to `var(--base)`
- `--color-muted` maps to `var(--raised)`

While the aliases provide backward compatibility in dark theme, **direct usage of new tokens is inconsistent**:

| Token Category | New Token Usage | Legacy Token Usage |
|---|---|---|
| Design system tokens (text-t1, text-t2, bg-base, etc.) | 637 occurrences in 65 files | -- |
| Legacy Shadcn tokens (text-muted-foreground, bg-muted, etc.) | -- | 403 occurrences in 83 files |
| Semantic color tokens (text-blue, bg-red, etc.) | 406 occurrences in 86 files | -- |

**Worst legacy token concentration by module:**
1. `modules/tms/` -- 153 legacy occurrences across 14 files (autotest-setup.tsx alone has 42)
2. `components/messenger/` -- 89 legacy occurrences across 21 files
3. `components/dashboard/` -- 31 legacy occurrences across 12 files
4. `routes/_authenticated/projects/$projectId/tms/cases_.$caseKey.tsx` -- 27 legacy occurrences in a single file
5. `routes/_authenticated/projects/$projectId/issues/index.tsx` -- 15 legacy occurrences

**Hardcoded colors (110+ in 38 files):**
- Chart components: burndown-chart.tsx uses `#94a3b8`, `#3b82f6`, `#71717a`, `rgba(24,24,27,0.95)` etc. for Recharts styling
- velocity-chart.tsx: `#71717a`, `rgba(255,255,255,0.04)` for axes and grid
- login.tsx: `#5B9CF5`, `#34D399`, `#A78BFA`, `#FBBF24` for feature card icons (these match design tokens but are hardcoded)
- ParticipantList.tsx: `#2a2f3e` hardcoded background, `ring-2 ring-[#2a2f3e]`
- onboarding.tsx: `#2563EB`, `#047857` in gradient buttons, `#ddd` borders for light theme preview

**Light theme concerns:**
- Scrollbar styles in app.css use `rgba(255,255,255,...)` which is invisible on white backgrounds (lines 373-382)
- `page-header-blur` class uses `rgba(13,14,18,0.8)` which is dark-only (line 659)

### Pillar 4: Typography (3/4)

**Font stack:** Correctly uses `var(--sans)` (Inter) globally via `body { font-family: var(--sans) }` in app.css.

**Size distribution (Tailwind scale usage: 688 occurrences across 129 files):**
- `text-xs` through `text-5xl` used across the app
- Arbitrary sizes also common: `text-[11px]`, `text-[12px]`, `text-[13px]`, `text-[14px]`, `text-[15px]`, `text-[16px]`
- This is expected for mockup-matched pages where precise sizing was specified

**Weight distribution (677 occurrences across 145 files):**
- `font-semibold` -- most common, used for headings and labels
- `font-bold` -- used for buttons and CTAs
- `font-medium` -- used for body emphasis
- `font-extrabold` -- used sparingly for onboarding CTA buttons
- This is a clean 4-weight system, appropriate for the application

**Base font size:** Set to 14px in app.css body rule, matching the mockup specification.

### Pillar 5: Spacing (2/4)

**Arbitrary values are heavily used:** 968 arbitrary bracket values (`p-[`, `m-[`, `w-[`, `h-[`, `gap-[`, `text-[`) across 115 files.

This is a deliberate tradeoff: mockup-matched pages (login, onboarding, dashboard, messenger, kanban, calls, TMS, settings, admin, trash) use pixel-precise values from HTML references. This ensures visual fidelity but sacrifices consistency with the Tailwind spacing scale.

**Consistency patterns:**
- Button padding: Varies between `px-4 py-2` (Tailwind), `px-[14px] py-[5px]`, `px-5 py-[9px]`, `px-7 py-[10px]` -- at least 4 different button padding patterns
- Card padding: Mostly `p-6` via Card component, but some use `p-[22px]`, `p-4`, `p-8`
- Border radius: Consistently uses `var(--r1)` through `var(--r4)` via `rounded-[var(--r2)]` pattern -- this is good
- Gap values: Mix of `gap-2`, `gap-3`, `gap-4` (Tailwind) and `gap-[6px]`, `gap-[7px]`, `gap-[9px]` (arbitrary)

**Positive:** Border radius and shadow usage is fully tokenized via CSS custom properties.

### Pillar 6: Experience Design (3/4)

**State coverage analysis:**

| State Type | Coverage | Files |
|---|---|---|
| Loading states (`isLoading`, `skeleton`, etc.) | Present | 66 files |
| Empty states (`isEmpty`, `.length === 0`, etc.) | Present | 76 files |
| Error boundaries (`ErrorBoundary`) | Minimal | 2 files (root + messenger) |
| Disabled states | Good | 71 files, 162 occurrences |
| Reduced motion | Global | app.css `@media (prefers-reduced-motion: reduce)` covers all animations |

**Strengths:**
- Global `prefers-reduced-motion` media query at app.css:1017 blanket-disables all animations and transitions
- Global `focus-visible` outline set in app.css:325-328 with `2px solid var(--primary)` and `outline-offset: 2px`
- Button primitive has `focus-visible:ring-2` built in
- Input primitive has custom focus glow effect
- Confirmation patterns exist for destructive actions (trash page has delete confirmation dialog)

**Issues:**
- Only 2 files implement ErrorBoundary -- most routes lack error boundary protection
- 190 `transition-all` usages across 70 files: while not a user-facing bug, this impacts rendering performance, especially on pages with many interactive elements (kanban board with 50+ cards, issue list with 100+ rows)
- `transition-all` in button.tsx (line 8) propagates to every Button instance in the app
- The `focus-visible` count in TSX files is only 5 occurrences in 4 files (button, input, textarea, tabs) -- other interactive elements rely on the global CSS rule which is less specific

---

## Cross-Page Consistency Analysis

### Pages using NEW design system tokens (well-migrated):
- `routes/login.tsx` -- fully token-based surfaces and text
- `routes/onboarding.tsx` -- fully token-based with mockup-matched styling
- `routes/_authenticated/trash.tsx` -- redesigned with new tokens
- `routes/_authenticated/settings.tsx` -- redesigned layout
- `routes/_authenticated/settings/profile.tsx` -- redesigned
- `routes/_authenticated/admin/dashboard.tsx` -- redesigned
- `routes/_authenticated/calls.tsx` -- redesigned
- `routes/_authenticated/dashboard.lazy.tsx` -- redesigned
- `components/layout/*` -- all layout shell components use new tokens

### Pages with MIXED token usage (partially migrated):
- `routes/_authenticated/projects/$projectId/issues/index.tsx` -- 15 legacy tokens
- `routes/_authenticated/projects/$projectId/tms/cases_.$caseKey.tsx` -- 27 legacy tokens
- `routes/_authenticated/messenger/*` -- messenger components mix old and new
- `components/dashboard/widgets/*` -- dashboard widgets have 31 legacy tokens

### Pages with HEAVY legacy token usage (not migrated):
- `modules/tms/autotest-setup.tsx` -- 42 legacy tokens, appears to be pre-redesign code
- `modules/tms/components/traceability-matrix.tsx` -- 16 legacy tokens
- `modules/tms/components/environment-comparison.tsx` -- 18 legacy tokens
- `modules/tms/components/test-case-metadata.tsx` -- 14 legacy tokens
- `modules/tms/components/coverage-report.tsx` -- 13 legacy tokens
- `components/messenger/invite-link-dialog.tsx` -- 11 legacy tokens
- `components/messenger/poll-message.tsx` -- 9 legacy tokens

### Files with zero design system awareness:
- `routes/__root.tsx` -- error/404 pages use only legacy tokens (7 occurrences)
- `components/admin/trash-manager.tsx` -- 8 legacy tokens, no new tokens
- `modules/tms/trigger-ci-dialog.tsx` -- 10 legacy tokens, no new tokens

---

## Light Theme Compatibility Issues

1. **Scrollbar styling** (app.css:373-382): Uses `rgba(255,255,255,...)` for thumb color -- invisible on white/light backgrounds. Should use `rgba(0,0,0,...)` for `:root` and `rgba(255,255,255,...)` for `.dark` only.

2. **Page header blur** (app.css:659): `background: rgba(13,14,18,0.8)` is a dark-only value. Needs a light variant or should use `var(--base)` with opacity.

3. **Glass morphism** (app.css:250-259): Light and dark variants are defined -- this is correct.

4. **Gradient overlays** in card `::before` pseudo-elements reference `rgba(255,255,255,...)` which creates a white line on already-white cards in light theme -- subtle but worth checking visually.

5. **Glow effects** on buttons (e.g., `shadow-[0_0_20px_rgba(91,156,245,0.25)]`) may be too subtle on light backgrounds.

---

## Files Audited

### UI Primitives (15 files)
- `components/ui/avatar.tsx`
- `components/ui/badge.tsx`
- `components/ui/button.tsx`
- `components/ui/card.tsx`
- `components/ui/dialog.tsx`
- `components/ui/dropdown-menu.tsx`
- `components/ui/input.tsx`
- `components/ui/label.tsx`
- `components/ui/popover.tsx`
- `components/ui/scroll-area.tsx`
- `components/ui/select.tsx`
- `components/ui/separator.tsx`
- `components/ui/tabs.tsx`
- `components/ui/textarea.tsx`
- `components/ui/tooltip.tsx`

### Layout Shell (7 files)
- `components/layout/app-shell.tsx`
- `components/layout/sidebar.tsx`
- `components/layout/sidebar-link.tsx`
- `components/layout/sidebar-nav.tsx`
- `components/layout/sidebar-projects.tsx`
- `components/layout/sidebar-user.tsx`
- `components/layout/topbar.tsx`

### Routes (30+ files)
- All files under `routes/` including login, onboarding, authenticated routes

### Modules (40+ files)
- All files under `modules/` including admin, calls, messenger, settings, tms, tracker

### Components (50+ files)
- All files under `components/` including admin, auth, calls, dashboard, files, messenger, notifications, search, tracker

### Design System Foundation
- `app.css` (1025 lines -- complete token system, animations, utilities)
