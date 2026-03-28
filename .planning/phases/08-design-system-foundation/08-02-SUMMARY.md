---
phase: 08-design-system-foundation
plan: 02
subsystem: frontend-css
tags: [css, important-cleanup, noise-overlay, design-system]
dependency_graph:
  requires: [08-01]
  provides: [clean-css-specificity, noise-texture-overlay]
  affects: [all-shadcn-components, all-pages]
tech_stack:
  added: []
  patterns: [css-layer-base, svg-feturbulence-noise, z-index-scale]
key_files:
  modified:
    - frontend/src/app.css
decisions:
  - Removed all 56 !important declarations to unblock component-level restyling in Phases 9-12
  - Noise overlay uses inline SVG feTurbulence (no external asset dependency)
  - z-index 9999 chosen to stay below Radix portal z-index (10000+)
metrics:
  duration: 2min
  completed: 2026-03-28
  tasks: 2
  files: 1
---

# Phase 08 Plan 02: CSS Important Cleanup and Noise Overlay Summary

Removed all 56 !important declarations from app.css (form overrides, dark theme shadows/focus rings) and added feTurbulence SVG noise texture overlay as body::after at z-index 9999 with pointer-events: none.

## What Was Done

### Task 1: Remove all !important declarations from app.css
**Commit:** 934f984

Deleted the entire form-override section from app.css:
- Select/Radix trigger overrides (24 !important props)
- Radix content wrapper styles (10 !important props)
- Input/textarea overrides (10 !important props)
- Checkbox/radio overrides (5 !important props)
- Dark theme dropdown shadows and focus rings (7 !important props)
- Firefox-specific select rule (part of the section)

Preserved: scrollbar styles, kanban board styles, animations, badges, glass effects, card-premium, page-title, progress-fill, typing-dots, and all other non-form sections.

**Verification:** `grep -c "!important" frontend/src/app.css` returns 0.

### Task 2: Add noise texture overlay and verify build
**Commit:** 9901432

Added noise texture overlay inside `@layer base {}` after the body rule:
- `body::after` with `position: fixed; inset: 0`
- Inline SVG with feTurbulence fractalNoise (baseFrequency .9, 4 octaves)
- `opacity: .3` outer, `.03` inner rect = extremely subtle grain
- `z-index: var(--z-noise, 9999)` -- below Radix portals
- `pointer-events: none` -- fully non-interactive

Vite build completes successfully with zero errors.

## Deviations from Plan

None -- plan executed exactly as written.

## Known Stubs

None.

## Decisions Made

1. **Full section deletion over selective !important removal** -- removing entire form-override blocks is cleaner than stripping !important from each property, since Shadcn component classes now handle all form styling through normal specificity.
2. **Noise overlay z-index 9999** -- uses CSS variable `--z-noise` with 9999 fallback, safely below Radix portal z-index (typically 10000+), ensuring dropdowns/dialogs render above noise.

## Commits

| Task | Commit | Description |
|------|--------|-------------|
| 1 | 934f984 | Remove all 56 !important declarations |
| 2 | 9901432 | Add noise texture overlay as body::after |

## Self-Check: PASSED
