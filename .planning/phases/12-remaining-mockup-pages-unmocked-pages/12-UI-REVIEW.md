# Phase 12 -- UI Review

**Audited:** 2026-03-28
**Baseline:** Abstract 6-pillar standards (no UI-SPEC.md exists)
**Screenshots:** Not captured (no dev server detected on ports 3000, 5173, 8080)

---

## Pillar Scores

| Pillar | Score | Key Finding |
|--------|-------|-------------|
| 1. Copywriting | 3/4 | Context-specific Russian copy throughout; a few legacy `text-muted-foreground` fallback strings and "OK" button labels remain |
| 2. Visuals | 4/4 | Clear visual hierarchy via gradient surfaces, extrabold titles, color-coded badges, and consistent icon treatment |
| 3. Color | 3/4 | Design system tokens used pervasively; residual hardcoded hex values in workflow/label color palettes and onboarding gradients |
| 4. Typography | 3/4 | Consistent use of pixel-exact `text-[Npx]` sizes from mockups; legacy Tailwind scale (`text-sm`, `text-xs`, `text-lg`) persists on unredesigned sections |
| 5. Spacing | 3/4 | Mockup pages use pixel-precise arbitrary spacing throughout; high consistency within Phase 12 pages but heavy reliance on arbitrary values |
| 6. Experience Design | 3/4 | Loading skeletons, empty states, and error toasts present; `window.confirm()` used for destructive actions in 3 locations instead of styled dialogs; no `aria-label` attributes on icon-only buttons |

**Overall: 19/24**

---

## Top 3 Priority Fixes

1. **Legacy Shadcn token leaks in calls.tsx "info" tab and messenger pages** -- Users see inconsistent surface colors (bg-muted, text-muted-foreground, bg-background) that clash with the redesigned pages -- Replace with design system equivalents: `bg-card` for `bg-muted/50`, `text-t3` for `text-muted-foreground`, `bg-base` for `bg-background` in calls.tsx lines 81-162, messenger/index.tsx, messenger/$channelId.tsx, __root.tsx error boundaries

2. **`window.confirm()` used for destructive actions on issue delete and comment delete** -- Users get a jarring native browser dialog instead of the styled confirm dialog already built for trash page -- Extract the confirm dialog pattern from trash.tsx and reuse in $issueKey.tsx:136, comment-list.tsx:102, and $projectId/index.tsx:241

3. **Zero `aria-label` attributes on icon-only buttons across all Phase 12 pages** -- Screen readers cannot identify button purposes for action buttons (dots menu in issue header, calendar nav arrows, call controls) -- Add `aria-label` props to all icon-only `<button>` and `<svg>` interactive elements

---

## Detailed Findings

### Pillar 1: Copywriting (3/4)

Phase 12 pages use contextual, domain-specific Russian copy that avoids generic labels.

**Positives:**
- CTA buttons are action-specific: "Восстановить", "Удалить навсегда", "Создать проект", "Новый звонок", "Начать работу"
- Empty states have descriptive messages with context ("Корзина пуста -- Здесь будут отображаться удалённые элементы")
- Error toasts use specific messages ("Не удалось обновить задачу", "Не удалось восстановить проект")
- Loading states communicate what's loading ("Загрузка комментариев...")
- Onboarding step titles are motivating ("Знакомство с Hub", "Выберите оформление")

**Issues:**
- `settings/profile.tsx:165` and `notification-settings.tsx:63`: Button label is just "OK" -- could be "Сохранить"
- `comment-list.tsx:113`: "Нет комментариев" uses `text-muted-foreground` (legacy token) instead of `text-t3`
- `comment-list.tsx:90`: "Загрузка комментариев..." also uses `text-muted-foreground`
- `calls.tsx:113,157`: The "info" tab section still uses `text-muted-foreground` and `text-xs` (legacy scale)
- `$issueKey.tsx:124`: "Задача не найдена" empty state uses `text-muted-foreground` rather than `text-t3`

### Pillar 2: Visuals (4/4)

Visual hierarchy is well-established across all Phase 12 pages.

**Positives:**
- Page titles consistently use `text-[18px] font-extrabold tracking-[-0.03em]` creating a clear focal point
- Two-column layout on issue detail (`grid gap-6 lg:grid-cols-[1fr_320px]`) with void-bg sidebar creates clear content separation
- Gradient card backgrounds (`from-card to-raised`) establish depth consistently across projects list, TMS run cards, trash type cards, and onboarding tour cards
- Color-coded badges (priority badges with red/blue split, status badges with green/gray, countdown badges with red/orange/blue) provide scannable visual cues
- Staggered entry animations (`animationDelay: ${idx * 40}ms`) add polish to table rows and card grids
- Admin KPI cards use sparkline SVGs providing data visualization at a glance
- Onboarding progress indicator uses color transitions (green=done, blue=active, gray=pending) for clear step tracking

**No issues found.**

### Pillar 3: Color (3/4)

Design system tokens are well-applied on Phase 12 pages. The token system (--blue, --green, --purple, --red, --amber, --orange, --gray) with glow variants (--gB, --gP, --gG, --gR) is used consistently.

**Positives:**
- Accent color usage is purposeful: blue for primary actions, purple for calls, green for success/done states, red for destructive actions
- Badge colors follow a clear semantic pattern (status badges, priority badges, countdown badges)
- Gradient buttons use `from-[var(--blueS)] to-[#2563EB]` consistently for primary actions
- Glow shadows (`shadow-[0_0_20px_var(--gB)]`) are reserved for interactive CTA buttons

**Issues:**
- `workflow.tsx:28-35`: 8 hardcoded hex colors for workflow status palette (`#6b7280`, `#3b82f6`, etc.) -- these are data-driven color pickers and acceptable as configuration values
- `labels.tsx:29-38`: 10 hardcoded hex colors for label palette -- same pattern, acceptable
- `onboarding.tsx:461,463`: Hardcoded `#ddd` for theme preview borders -- should use `var(--b2)` or `var(--b3)`
- `onboarding.tsx:480`: Hardcoded `#fff` in system theme preview gradient -- acceptable for visual demonstration
- `calls.tsx:107`: Dynamic Tailwind classes `bg-${s.color}-500/10 text-${s.color}-500` -- these won't work at runtime as Tailwind purges unused classes; should use explicit class mappings
- `projects/index.tsx:75`: Same dynamic Tailwind class pattern `bg-${s.color}-500/10` -- same runtime issue
- `calls.tsx:81,152`: Uses `bg-card` and `hover:bg-muted/50` (Shadcn tokens) instead of design system tokens in the "info" tab

### Pillar 4: Typography (3/4)

Phase 12 pages use a consistent pixel-exact typography system derived from mockups.

**Font sizes observed on Phase 12 pages:**
- `text-[10px]`, `text-[10.5px]`, `text-[11px]`, `text-[11.5px]`, `text-[12px]`, `text-[12.5px]`, `text-[13px]`, `text-[13.5px]`, `text-[14px]`, `text-[15px]`, `text-[16px]`, `text-[17px]`, `text-[18px]`, `text-[24px]`, `text-[26px]`, `text-[30px]`

This is 16 distinct font sizes, which is high but matches a pixel-perfect mockup approach. The sizes follow a deliberate scale.

**Font weights used consistently:**
- `font-medium` (nav items, table cells)
- `font-semibold` (labels, badges, buttons, active states)
- `font-bold` (section titles, names, card headings)
- `font-extrabold` (page titles, step numbers)
- `font-black` (onboarding headings, KPI values)

5 weights is slightly high but each serves a clear hierarchical purpose.

**Issues:**
- `calls.tsx:83`: Uses `text-lg` (Tailwind scale) in the "info" tab instead of `text-[Npx]` matching the mockup pattern
- `calls.tsx:107,112,113,156,157`: Uses `text-sm`, `text-xs` (Tailwind scale) instead of pixel-exact sizes
- `__root.tsx:30,77`: Error boundary uses `text-2xl`, `text-xl` (Tailwind scale) -- pre-Phase 12, not restyled
- `messenger/index.tsx:27,30`: Uses `text-xl`, `text-sm` -- pre-Phase 12, not restyled
- Mixed typography systems between Phase 12 pages (pixel-exact) and unredesigned pages (Tailwind scale) creates inconsistency

### Pillar 5: Spacing (3/4)

Phase 12 pages use pixel-exact arbitrary spacing values from mockups. The spacing is internally consistent within each page.

**Spacing patterns observed:**
- Page padding: `p-6`, `px-6`, `px-9 py-7` (settings)
- Card padding: `p-5`, `p-[22px]`, `p-6`
- Gaps: `gap-[14px]`, `gap-[10px]`, `gap-[9px]`, `gap-4`, `gap-6`, `gap-3`
- Margins: `mb-[18px]`, `mb-[14px]`, `mb-[7px]`, `mb-2`, `mb-6`, `mb-9`

**Positives:**
- Sidebar meta fields use consistent `mb-[18px]` wrapper spacing with `mb-[7px]` label spacing
- Table cells use consistent `px-[18px] py-4` across trash and TMS pages
- Card internal spacing is consistent within each page type

**Issues:**
- 300 occurrences of arbitrary pixel values (`[Npx]`) across Phase 12 route files -- this is a heavy reliance on non-standard spacing, though it matches mockups
- Inconsistent card padding between pages: projects list uses `p-5`, onboarding tour cards use `p-6`, trash type cards use `p-[22px]` -- minor but noticeable
- Settings layout uses negative margins (`-mx-4 -my-4 md:-mx-6 md:-my-4`) to break out of parent padding, which is a layout workaround

### Pillar 6: Experience Design (3/4)

State coverage is good across Phase 12 pages. Loading, empty, and error states are handled in the key flows.

**State coverage analysis:**
- Loading states: 17 files with loading/isLoading patterns in routes; issue detail has full skeleton, projects list has card skeletons, trash has shimmer blocks
- Error states: 18 files with error handling; toast notifications for mutation failures ("Не удалось обновить задачу", "Не удалось восстановить проект")
- Empty states: 13 files with empty state handling; projects list has illustrated CTA empty state, trash has icon+text empty state, comments list has text fallback
- Disabled states: Buttons use `disabled:opacity-50` (calls button), `disabled` prop on restore button during pending mutation

**Positives:**
- Trash page has a well-designed confirm dialog with backdrop blur, scale-in animation, and clear "Отмена" / "Удалить навсегда" actions
- Onboarding shows clear step progression with animated transitions between steps
- Staggered row animations provide visual feedback on data load

**Issues:**
- `$issueKey.tsx:136`: Delete issue uses `window.confirm()` -- native browser dialog, no styling, breaks design language
- `comment-list.tsx:102`: Delete comment uses `window.confirm()` -- same issue
- `$projectId/index.tsx:241`: Delete project uses `window.confirm()` -- same issue
- `labels.tsx:166,195`: Delete label uses `window.confirm()` -- same issue (5 total occurrences)
- Zero `aria-label` attributes found across all route-level components -- icon-only buttons (calendar arrows, dots menu, close buttons, call control) are inaccessible to screen readers
- Trash "Очистить все" button (trash.tsx:89-94) has no onClick handler -- it renders but does nothing when clicked
- Onboarding form inputs at step 1: "Фамилия" field (line 324) and timezone/language selects (lines 361,370) are uncontrolled -- they have no `value` or `onChange` bindings, so user input is lost on form submission

---

## Files Audited

- `frontend/src/routes/_authenticated/projects/$projectId/issues/$issueKey.tsx`
- `frontend/src/modules/tracker/issue-sidebar-meta.tsx`
- `frontend/src/components/tracker/comment-list.tsx`
- `frontend/src/routes/_authenticated/calls.tsx`
- `frontend/src/routes/_authenticated/settings.tsx`
- `frontend/src/routes/_authenticated/admin/dashboard.tsx`
- `frontend/src/routes/onboarding.tsx`
- `frontend/src/routes/_authenticated/trash.tsx`
- `frontend/src/routes/_authenticated/projects/index.tsx`
- `frontend/src/app.css` (design tokens, CSS utilities)
- `frontend/components.json` (shadcn config)
- Phase summaries: 12-01 through 12-05
- Phase context: 12-CONTEXT.md, 12-VERIFICATION.md

---

*Audited by Claude (gsd-ui-reviewer) on 2026-03-28*
