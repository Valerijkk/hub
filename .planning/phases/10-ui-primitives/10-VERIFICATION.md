---
phase: 10-ui-primitives
verified: 2026-03-28T12:55:00Z
status: gaps_found
score: 9/10 must-haves verified
gaps:
  - truth: "User sees Avatar with online indicator glow capability"
    status: partial
    reason: "avatar.tsx only adds bg-raised and text-t2 to AvatarFallback. No online indicator glow ring, dot, or shadow is present anywhere in the component or its known consumers. PRIM-09 requirement explicitly states 'online indicator glow'."
    artifacts:
      - path: "frontend/src/components/ui/avatar.tsx"
        issue: "AvatarFallback has bg-raised text-t2 (correct) but no online status indicator (ring, dot, glow shadow, or ::after pseudo) present in the component or any consumer page"
    missing:
      - "Online status glow: an absolute-positioned dot element or ring on AvatarRoot for consumers to opt into, e.g. data-online attribute with ring shadow, or a separate OnlineAvatar wrapper component"
human_verification:
  - test: "Visually confirm Button gradient renders correctly"
    expected: "Primary button shows blue gradient (blueS to blue) with glow shadow on hover and scale on active"
    why_human: "CSS gradient and box-shadow rendering cannot be verified by static grep"
  - test: "Visually confirm Card top-edge highlight is visible"
    expected: "Card shows 1px gradient line at top edge transitioning from transparent through b3 to transparent"
    why_human: "CSS ::before pseudo-element rendering cannot be verified programmatically"
  - test: "Visually confirm Dialog backdrop blur on open"
    expected: "Opening a dialog shows blurred background (8px blur) at 65% opacity"
    why_human: "backdrop-filter rendering and animation cannot be verified statically"
  - test: "Visually confirm Tabs active underline indicator"
    expected: "Active tab shows bottom border in primary blue with glow shadow beneath it"
    why_human: "Tailwind border-b-2 rendering and visual glow require browser rendering"
---

# Phase 10: UI Primitives Verification Report

**Phase Goal:** Все 10 базовых Shadcn-компонентов рестайлированы под новый дизайн-язык, и любая страница, использующая эти компоненты, автоматически выглядит обновлённой
**Verified:** 2026-03-28T12:55:00Z
**Status:** gaps_found
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths (from ROADMAP Success Criteria)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | User sees redesigned Button (gradient primary, glow hover) and Card (gradient bg, top-edge highlight, hover glow) | ✓ VERIFIED | button.tsx: `bg-gradient-to-br from-[var(--blueS)] to-[var(--blue)]`, `shadow-[0_0_20px_rgba(91,156,245,0.25)]`, `active:scale-[.96]`; card.tsx: `bg-gradient-to-b from-card to-raised`, `before:h-px before:bg-gradient-to-r before:via-b3` |
| 2 | User sees redesigned Input (blue focus glow ring), Badge (status-specific glow), and Avatar (online indicator glow) | ✗ PARTIAL | Input: verified (`focus-visible:shadow-[0_0_0_3px_rgba(91,156,245,0.06)]`); Badge: verified (`var(--gB)`, `var(--gR)`, `var(--gG)`, `var(--gP)`); Avatar: **partial** — design tokens added but no online glow indicator present |
| 3 | User sees redesigned Dialog/Popover (entrance animation, backdrop blur), DropdownMenu, Tabs (active indicator), Tooltip, ScrollArea (slim 5px scrollbar) | ✓ VERIFIED | dialog.tsx: `backdrop-blur-[8px]`, `bg-base`, `rounded-[var(--r4)]`, `ease-[var(--spring)]`; tabs.tsx: `border-b-2 border-transparent`, `data-[state=active]:border-primary`; dropdown-menu.tsx: `bg-raised`, `border-b2`, `var(--hover)`; tooltip.tsx: `rounded-[var(--r1)] border border-b1 bg-raised`; scroll-area.tsx: `w-[5px]`, `bg-b2 hover:bg-b3` |
| 4 | All existing page functionality continues to work without regression | ? UNCERTAIN | All behavioral props (forwardRef, asChild, Slot, Radix primitives, Portal, data-[state], data-[side]) verified preserved in all 10 files. Full regression requires human testing. |

**Score:** 9/10 truths verified (1 partial, 1 uncertain)

### Required Artifacts

| Artifact | Expected | Level 1 (Exists) | Level 2 (Substantive) | Level 3 (Wired) | Status |
|----------|----------|-------------------|----------------------|-----------------|--------|
| `frontend/src/components/ui/button.tsx` | Gradient primary with glow hover | ✓ | ✓ `bg-gradient-to-br`, `shadow-[0_0_20px...]`, `active:scale-[.96]`, `var(--r2)` | ✓ Used across app | ✓ VERIFIED |
| `frontend/src/components/ui/card.tsx` | Gradient bg, top-edge highlight, hover glow | ✓ | ✓ `bg-gradient-to-b from-card to-raised`, `before:h-px before:bg-gradient-to-r before:via-b3`, `hover:-translate-y-0.5` | ✓ Used across app | ✓ VERIFIED |
| `frontend/src/components/ui/input.tsx` | Blue focus glow ring | ✓ | ✓ `h-[42px]`, `var(--r2)`, `var(--input)`, `rgba(91,156,245,0.06)` shadow, `scale-[1.005]` | ✓ Used across app | ✓ VERIFIED |
| `frontend/src/components/ui/badge.tsx` | Status-specific glow backgrounds | ✓ | ✓ `var(--gB)`, `var(--gR)`, `var(--gG)`, `var(--gP)`, `text-[11px]`, `font-bold`, `px-[7px]` | ✓ Used across app | ✓ VERIFIED |
| `frontend/src/components/ui/scroll-area.tsx` | 5px slim scrollbar | ✓ | ✓ `w-[5px]`, `h-[5px]`, `bg-b2 hover:bg-b3` | ✓ Radix scrollbar wired | ✓ VERIFIED |
| `frontend/src/components/ui/avatar.tsx` | Design tokens + online glow capability | ✓ | ✗ PARTIAL — `bg-raised text-t2` present, but no online status indicator (ring, dot, shadow, pseudo) | ✓ Used across app | ✗ PARTIAL |
| `frontend/src/components/ui/tooltip.tsx` | Raised bg, b1 border, r1 radius | ✓ | ✓ `rounded-[var(--r1)]`, `border-b1`, `bg-raised`, `text-xs`, `text-t1` | ✓ Radix content wired | ✓ VERIFIED |
| `frontend/src/components/ui/dialog.tsx` | Backdrop blur, spring entrance, raised surface, 16px radius | ✓ | ✓ `backdrop-blur-[8px]`, `bg-black/65`, `bg-base`, `border-b2`, `rounded-[var(--r4)]`, `ease-[var(--spring)]` | ✓ DialogPrimitive preserved | ✓ VERIFIED |
| `frontend/src/components/ui/tabs.tsx` | Bottom bar active indicator, accent glow | ✓ | ✓ `border-b-2 border-transparent`, `data-[state=active]:border-primary`, `data-[state=active]:shadow-[0_2px_8px_rgba(91,156,245,0.15)]`, `text-[13px]` | ✓ TabsPrimitive preserved | ✓ VERIFIED |
| `frontend/src/components/ui/dropdown-menu.tsx` | Raised background, hover highlights | ✓ | ✓ `bg-raised`, `border-b2`, `var(--r2)`, `var(--r1)`, `var(--hover)`, `text-t1`, `bg-b1` separator | ✓ Portal preserved | ✓ VERIFIED |

### Key Link Verification

| From | To | Via | Pattern | Status |
|------|----|-----|---------|--------|
| button.tsx | app.css tokens | CSS variable references | `from-primary` / `from-[var(--blueS)]` | ✓ WIRED — `from-[var(--blueS)] to-[var(--blue)]` confirmed in file; tokens `--blue`, `--blueS` confirmed in app.css |
| card.tsx | app.css tokens | CSS variable references | `from-card` | ✓ WIRED — `from-card to-raised` confirmed; `--raised` token defined in app.css |
| badge.tsx | app.css tokens | CSS variable references | `var(--gB)` | ✓ WIRED — `var(--gB)`, `var(--gR)`, `var(--gG)`, `var(--gP)` all confirmed; tokens defined in app.css |
| scroll-area.tsx | app.css scrollbar styles | Consistent 5px width | `w-[5px]` | ✓ WIRED — `w-[5px]` and `h-[5px]` both present |
| dialog.tsx | app.css tokens | CSS variable references | `var(--r4)` | ✓ WIRED — `rounded-[var(--r4)]`, `ease-[var(--spring)]` confirmed; tokens in app.css |
| tabs.tsx | app.css tokens | CSS variable references | `text-t3` | ✓ WIRED — `text-t3` in TabsList, `ease-[var(--ease)]` in TabsTrigger; tokens in app.css |

### Data-Flow Trace (Level 4)

Not applicable. All artifacts are pure presentational components — they accept props/children and render className strings. No data fetching or dynamic state involved.

### Behavioral Spot-Checks

| Behavior | Check | Result | Status |
|----------|-------|--------|--------|
| Button commit d009655 exists | `git log --oneline d009655` | `d009655 feat(10-01): restyle Button with gradient primary, glow hover, active scale` | ✓ PASS |
| Card commit ddfcd9e exists | `git log --oneline ddfcd9e` | `ddfcd9e feat(10-01): restyle Card with gradient bg, top-edge highlight, hover lift` | ✓ PASS |
| Input commit d5dcb9a exists | `git log --oneline d5dcb9a` | `d5dcb9a feat(10-01): restyle Input with blue focus glow ring, design tokens` | ✓ PASS |
| Badge/ScrollArea/Avatar/Tooltip commits exist | `git log` output | All 8 commits verified: d009655, ddfcd9e, d5dcb9a, 524d587, 983da8b, a4e16e9, 3d019b7, b1f5626 | ✓ PASS |
| CSS tokens present in app.css | grep for `--r2`, `--r4`, `--gB`, `--spring` | All tokens found: `--r1: 6px; --r2: 8px; --r4: 16px`, `--gB: rgba(59,126,224,.12)`, `--spring: cubic-bezier(.175,.885,.32,1.1)` | ✓ PASS |

### Requirements Coverage

| Requirement ID | Source Plan | Description | Status | Evidence |
|---------------|-------------|-------------|--------|----------|
| PRIM-01 | 10-01-PLAN.md | Redesigned Button with gradient primary, ghost/secondary variants, glow hover | ✓ SATISFIED | `bg-gradient-to-br from-[var(--blueS)] to-[var(--blue)]`, glow shadows, `active:scale-[.96]` in button.tsx |
| PRIM-02 | 10-01-PLAN.md | Redesigned Card with gradient bg, top-edge highlight, hover glow | ✓ SATISFIED | `bg-gradient-to-b from-card to-raised`, `before:h-px before:bg-gradient-to-r before:via-b3`, hover effects in card.tsx |
| PRIM-03 | 10-01-PLAN.md | Redesigned Input with blue focus glow ring and subtle scale on focus | ✓ SATISFIED | `focus-visible:shadow-[0_0_0_3px_rgba(91,156,245,0.06)]`, `focus-visible:scale-[1.005]` in input.tsx |
| PRIM-04 | 10-02-PLAN.md | Redesigned Badge with status-specific glow backgrounds | ✓ SATISFIED | `var(--gB)`, `var(--gR)`, `var(--gG)`, `var(--gP)` variants, `text-[11px] font-bold` in badge.tsx |
| PRIM-05 | 10-03-PLAN.md | Redesigned Dialog/Popover with entrance animation and backdrop blur | ✓ SATISFIED | `backdrop-blur-[8px]`, `bg-black/65`, `ease-[var(--spring)]`, `rounded-[var(--r4)]` in dialog.tsx |
| PRIM-06 | 10-03-PLAN.md | Redesigned Tabs with active indicator matching section color | ✓ SATISFIED | `border-b-2 border-transparent`, `data-[state=active]:border-primary`, accent glow shadow in tabs.tsx |
| PRIM-07 | 10-03-PLAN.md | Redesigned DropdownMenu matching new design language | ✓ SATISFIED | `bg-raised`, `border-b2`, `var(--r2)`, `var(--r1)`, `var(--hover)`, `bg-b1` in dropdown-menu.tsx |
| PRIM-08 | 10-02-PLAN.md | Redesigned ScrollArea with slim 5px scrollbar | ✓ SATISFIED | `w-[5px]`, `h-[5px]`, `bg-b2 hover:bg-b3 transition-colors` in scroll-area.tsx |
| PRIM-09 | 10-02-PLAN.md | Redesigned Avatar with online indicator glow | ✗ PARTIAL | AvatarFallback has `bg-raised text-t2` (design tokens updated) but no online status indicator — no ring, dot, glow shadow, or wrapper for online state present in avatar.tsx or any consumer |
| PRIM-10 | 10-02-PLAN.md | Redesigned Tooltip matching new palette and radius | ✓ SATISFIED | `rounded-[var(--r1)]`, `border border-b1`, `bg-raised`, `text-xs text-t1` in tooltip.tsx |

**Orphaned requirements:** None. All 10 PRIM IDs appear in plan frontmatter. REQUIREMENTS.md marks all 10 as Complete for Phase 10.

**Note on PRIM-09 discrepancy:** REQUIREMENTS.md marks PRIM-09 as `[x] Complete` but the implementation only updates AvatarFallback colors. The "online indicator glow" visual behavior described in the requirement has no implementation in the codebase.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| frontend/src/components/ui/input.tsx | 11 | `ring-offset-background` retained alongside new focus glow | ℹ️ Info | Not a stub — this class is harmless alongside the new `focus-visible:border-[rgba...]` and `focus-visible:shadow-[...]` approach. No visual conflict since `focus-visible:outline-none` suppresses the default ring. |

No TODO/FIXME markers, no empty implementations, no placeholder returns found across all 10 files.

### Human Verification Required

#### 1. Button gradient and glow rendering

**Test:** Open any page with a primary Button (e.g., login form submit button). Inspect visually.
**Expected:** Gradient from `--blueS` to `--blue` visible; blue glow shadow on hover; button lifts slightly and then scales down on click.
**Why human:** CSS gradient, box-shadow glow, and transform animations require browser rendering.

#### 2. Card top-edge highlight

**Test:** Open any page using Card components (e.g., Dashboard widgets, Settings panels). Inspect the top edge of a Card.
**Expected:** 1px horizontal gradient line visible at very top of card, transitioning from transparent through the `--b3` border color to transparent.
**Why human:** CSS `::before` pseudo-element rendering cannot be verified statically.

#### 3. Dialog backdrop blur

**Test:** Trigger any dialog in the app. Observe the overlay behind the dialog.
**Expected:** Background content is visible but blurred (8px), with 65% dark overlay. Dialog itself appears with spring (overshoot) entrance animation.
**Why human:** `backdrop-filter: blur(8px)` rendering and spring easing animations require browser testing.

#### 4. Tabs active underline indicator

**Test:** Navigate to any tabbed page (e.g., Project settings, TMS, Admin). Click between tabs.
**Expected:** Active tab shows a solid primary-blue underline at the bottom, with a faint blue glow shadow beneath it. Inactive tabs show muted text with no underline.
**Why human:** Tailwind `border-b-2` with data attribute rendering and glow shadow require visual confirmation.

#### 5. Regression: existing functionality unchanged

**Test:** Use app normally — create a project, open messenger, drag kanban cards, open a dialog, use dropdowns.
**Expected:** All functionality works identically to before. No visual regressions in existing pages.
**Why human:** Functional regression testing requires end-to-end interaction.

### Gaps Summary

**1 gap blocking full goal achievement:**

**PRIM-09 — Avatar online indicator glow not implemented.**

The requirement `PRIM-09: User sees redesigned Avatar with online indicator glow` is only partially fulfilled. The AvatarFallback background was updated to design tokens (`bg-raised text-t2`), which satisfies the "redesigned" aspect, but the "online indicator glow" — a visible ring, dot, or glow shadow indicating user online presence — is absent from avatar.tsx and from all pages that use Avatar.

REQUIREMENTS.md has this marked as `[x] Complete`, which is inconsistent with the codebase. Either:
- The online indicator glow belongs to a higher-level wrapper component (not in scope for this phase), OR
- The glow pattern needs to be added to avatar.tsx as an optional visual affordance

The gap is **partial** severity — the component is substantially updated with design tokens, and the absence of an online indicator glow does not prevent the other 9 components from working. This is a scoping/completeness issue rather than a blocking regression.

---

_Verified: 2026-03-28T12:55:00Z_
_Verifier: Claude (gsd-verifier)_
