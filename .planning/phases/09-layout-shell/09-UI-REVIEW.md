# Phase 9 -- UI Review

**Audited:** 2026-03-28
**Baseline:** HTML mockups in `редизайн/02-dashboard (2).html` (CSS specs for sidebar, topbar, user profile)
**Screenshots:** Not captured (no dev server detected on ports 3000, 5173, 8080)

---

## Pillar Scores

| Pillar | Score | Key Finding |
|--------|-------|-------------|
| 1. Copywriting | 3/4 | All labels are localized via i18n; one hardcoded Russian string in sidebar-user aria-label |
| 2. Visuals | 4/4 | Gradient edge, accent glow, glassmorphic topbar all match mockup spec |
| 3. Color | 3/4 | Design tokens used consistently; one hardcoded hex array in sidebar-projects |
| 4. Typography | 4/4 | Font sizes and weights match mockup CSS exactly |
| 5. Spacing | 4/4 | All spacing values match mockup px-for-px; arbitrary values are intentional spec matches |
| 6. Experience Design | 3/4 | Collapse/expand, mobile overlay, keyboard shortcuts present; no loading state for projects query |

**Overall: 21/24**

---

## Top 3 Priority Fixes

1. **Hardcoded aria-label "Razvyernut" in sidebar-user.tsx:52** -- Screen reader users get untranslated Russian label when locale is not Russian -- Replace with `t("sidebar.expand")` using the i18n function (pass `t` as prop to `SidebarUserProfile`).

2. **No loading/error state for projects query in sidebar-projects.tsx** -- When projects API is slow or fails, the projects section renders nothing with no feedback -- Add a skeleton or spinner when `isLoading` is true, and a subtle error indicator when `isError` is true from the `useQuery` result.

3. **Hardcoded hex colors in PROJECT_DOT_COLORS array (sidebar-projects.tsx:15)** -- Colors bypass the design token system, preventing theme consistency -- Replace `["#3B82F6", "#8B5CF6", "#10B981", "#F59E0B"]` with CSS variable references like `var(--blue)`, `var(--purple)`, `var(--green)`, `var(--amber)` via inline style, or use Tailwind color classes (`bg-blue-500`, etc.) to stay in the token pipeline.

---

## Detailed Findings

### Pillar 1: Copywriting (3/4)

Strong localization coverage throughout the layout shell. All navigation labels go through `t()` with sensible `defaultValue` fallbacks:

- `sidebar.tsx:76-78` -- "Создать" tooltip with `t("sidebar.createProject")`
- `sidebar.tsx:88` -- collapse button uses `t("sidebar.collapse")`
- `sidebar.tsx:119` -- expand button uses `t("sidebar.expand")`
- `sidebar-nav.tsx` -- all 6 nav items use `t("nav.*")` keys
- `mobile-nav.tsx` -- all items + "More" button use `t("nav.*")` keys

**Issue:** `sidebar-user.tsx:52` has a hardcoded `aria-label="Развернуть"` instead of using the `t()` function. The component does not receive `t` as a prop, only `collapsed` and `toggle`.

Empty state for projects (`sidebar-projects.tsx:256-259`) uses `t("sidebar.noProjects")` -- good.

User display name fallback chain (`sidebar-user.tsx:23`) is well-structured: `display_name -> name -> email -> "Пользователь"`. The hardcoded fallback "Пользователь" should ideally also use `t()`.

### Pillar 2: Visuals (4/4)

The implementation matches the mockup CSS spec with high fidelity:

**Gradient edge line** (`sidebar.tsx:35`):
- Mockup: `linear-gradient(180deg,transparent 5%,rgba(251,191,36,.05) 30%,rgba(91,156,245,.04) 60%,transparent 95%)`
- Implementation: Identical gradient values. Uses absolute div instead of `::after` pseudo -- valid Tailwind approach, same visual result.

**Active nav glow** (`sidebar-link.tsx:65-77`):
- Mockup: `::before` with bg amber, `box-shadow: 0 0 12px var(--amber)`, `::after` with gradient overlay
- Implementation: Two conditional `<span>` elements with `bg-primary` and `shadow-[0_0_12px_var(--primary)]`. Uses `--primary` instead of `--amber` to support user accent color -- intentional improvement over mockup.

**Glassmorphic topbar** (`topbar.tsx:15`):
- Mockup: `backdrop-filter: blur(24px) saturate(180%)`, `background: rgba(13,14,18,.75)`
- Implementation: `backdrop-blur-[24px] backdrop-saturate-[180%] bg-[rgba(13,14,18,.75)]` -- exact match.

**Topbar gradient edge** (`topbar.tsx:47-50`):
- Mockup: `linear-gradient(90deg,transparent,rgba(251,191,36,.06),transparent)` at `left:10% right:10%`
- Implementation: Identical. Uses absolute div -- matches approach used in sidebar.

All icon-only buttons have aria-labels or tooltips. Decorative elements have `aria-hidden="true"`. Logo has hover scale animation matching mockup's `transform:scale(1.08)` (implementation uses `scale-105` -- minor 3% difference, visually negligible).

### Pillar 3: Color (3/4)

Design token adoption is thorough. No Shadcn semantic color classes remain in the audited files:

- Sidebar: `bg-sidebar`, `text-t1`, `text-t2`, `text-t3`, `text-t4`, `border-b1`
- Topbar: `border-b1`, `text-t2`
- Mobile nav: `text-t3` (inactive), `text-primary` (active), `bg-base`, `border-b1`
- Active states: `bg-[var(--active)]`, `hover:bg-[var(--hover)]`
- Shadows: `var(--gB)`, `var(--gG)` -- token-based glow colors

**Issue:** `sidebar-projects.tsx:15` defines `PROJECT_DOT_COLORS` with hardcoded hex values `["#3B82F6", "#8B5CF6", "#10B981", "#F59E0B"]`. These are Tailwind blue-500, violet-500, emerald-500, amber-500 equivalents but bypass the token system. The mockup uses these same colors but in a design system context they should reference tokens.

**Accent usage is well-scoped:** `bg-primary` appears only on the active accent bar (`sidebar-link.tsx:67`) and its gradient overlay (`sidebar-link.tsx:74`) -- exactly 2 elements per active item, not overused.

`rgba(255,255,255,.04)` in `sidebar-user.tsx:74` for the role badge pill matches the mockup spec for `.sb-ur` exactly.

### Pillar 4: Typography (4/4)

Font sizes and weights match the mockup CSS spec precisely:

| Element | Mockup | Implementation | Match |
|---------|--------|----------------|-------|
| Brand "Hub" | 18px / 800 / -0.04em | `text-[18px] font-[800] tracking-[-0.04em]` | Exact |
| Nav items | 13.5px / 500 | `text-[13.5px] font-medium` | Exact |
| Nav active | 13.5px / 600 | `font-semibold` | Exact |
| User name | 13px / 600 | `text-[13px] font-semibold` | Exact |
| Role badge | 10px / 600 | `text-[10px] font-semibold` | Exact |
| Section label | 10px / 700 / uppercase / 0.1em | `text-[10px] font-bold uppercase tracking-[0.1em]` | Exact |
| Topbar title | 18px / 800 / -0.03em | `text-[18px] font-[800] tracking-[-0.03em]` | Exact |
| Mobile nav | 10px / 500 | `text-[10px] font-medium` | Exact |

Distinct font sizes in layout files: `text-xs` (logo letter), `text-sm` (logo letter expanded), `text-[10px]`, `text-[11.5px]`, `text-[13px]`, `text-[13.5px]`, `text-[18px]`. All are intentional per mockup spec -- no rogue sizes.

Font weights: `font-medium` (500), `font-semibold` (600), `font-bold` (700), `font-[800]`. Four weights, all deliberate per mockup hierarchy.

### Pillar 5: Spacing (4/4)

Spacing values match mockup CSS spec:

| Element | Mockup | Implementation | Match |
|---------|--------|----------------|-------|
| Sidebar width | 260px | `w-[260px]` | Exact |
| Sidebar collapsed | 64px (implied) | `w-[64px]` | Exact |
| Header height | 56px | `h-14` (= 56px) | Exact |
| Header padding | 0 18px | `px-[18px]` | Exact |
| Nav item gap | 11px | `gap-[11px]` | Exact |
| Nav item padding | 9px 12px | `py-[9px] px-3` (= 12px) | Exact |
| User section padding | 14px 18px | `px-[18px] h-14` | Exact |
| Avatar size | 34px | `h-[34px] w-[34px]` | Exact |
| Online indicator | 10px | `h-[10px] w-[10px]` | Exact |
| Online border | 2.5px | `border-[2.5px]` | Exact |
| Logo size | 30px | `h-[30px] w-[30px]` | Exact |
| Logo radius | 9px | `rounded-[9px]` | Exact |
| Project dot | 7px | `w-[7px] h-[7px]` | Exact |
| Topbar height | 56px | `h-14` | Exact |
| Topbar padding | 0 24px | `px-6` (= 24px) | Exact |

All arbitrary spacing values (`[18px]`, `[30px]`, etc.) are justified by the mockup spec -- they are not random, they are design decisions. No spacing inconsistencies detected.

### Pillar 6: Experience Design (3/4)

**Collapse/expand:** Works via `useSidebarStore` Zustand store. Transition: `duration-200 ease-[var(--ease)]`. Collapsed mode shows tooltips on hover (sidebar-link.tsx:100-108). Collapsed user section has tooltip showing display name.

**Mobile:** Overlay sidebar with `bg-black/60 backdrop-blur-sm` backdrop. Escape key handler on overlay div. Bottom nav with 5 items. "More" menu overlay with Escape handler. `pb-20` on content area to clear bottom nav.

**Keyboard:** Global `Ctrl+N` shortcut for new issue (`app-shell.tsx:24-28`).

**Accessibility:** All icon-only buttons have aria-labels. Decorative elements have `aria-hidden="true"`. Tooltips on collapsed sidebar items.

**Missing states:**
- `sidebar-projects.tsx` uses `useQuery` for projects but does not destructure or handle `isLoading` or `isError`. When the API is slow, the section is blank with no skeleton or spinner.
- No `ErrorBoundary` wrapping the sidebar or its sections.
- The "More" overlay in mobile-nav uses `onClick` on the backdrop div but no `role="dialog"` or `aria-modal` for assistive technology.

**Topbar title prop:** Intentionally unwired (will connect per-page later) -- documented as known stub. Not a deficiency for this phase.

---

## Files Audited

- `frontend/src/components/layout/sidebar.tsx` -- Sidebar shell (152 lines)
- `frontend/src/components/layout/sidebar-link.tsx` -- Nav item with active glow (112 lines)
- `frontend/src/components/layout/sidebar-user.tsx` -- User profile section (87 lines)
- `frontend/src/components/layout/sidebar-projects.tsx` -- Projects with expandable submenus (263 lines)
- `frontend/src/components/layout/sidebar-nav.tsx` -- Navigation links block (157 lines)
- `frontend/src/components/layout/topbar.tsx` -- Glassmorphic top bar (53 lines)
- `frontend/src/components/layout/app-shell.tsx` -- App layout wrapper (77 lines)
- `frontend/src/components/mobile-nav.tsx` -- Mobile bottom navigation (222 lines)
- `редизайн/02-dashboard (2).html` -- Design reference mockup (CSS spec)
