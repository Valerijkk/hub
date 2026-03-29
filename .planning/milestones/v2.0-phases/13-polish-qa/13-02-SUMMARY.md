---
phase: 13-polish-qa
plan: 02
subsystem: ui
tags: [css, performance, dead-code, animations, compositor]

# Dependency graph
requires:
  - phase: 12-remaining-pages
    provides: "All page CSS styles in app.css"
provides:
  - "Clean app.css with no unused custom CSS classes"
  - "All transitions use compositor-safe properties (no transition:all)"
  - "Dashboard/Kanban animations verified 60fps-safe (transform+opacity only)"
affects: []

# Tech tracking
tech-stack:
  added: []
  patterns: ["Explicit transition properties instead of transition:all"]

key-files:
  created: []
  modified:
    - "frontend/src/app.css"

key-decisions:
  - "Kept all @keyframes blocks even when animation classes removed (may be used via Tailwind arbitrary values)"
  - "Replaced global transition:all on interactive elements with explicit property list for compositor safety"

patterns-established:
  - "Compositor-safe transitions: always list specific properties, never use transition:all"

requirements-completed: []

# Metrics
duration: 4min
completed: 2026-03-28
---

# Phase 13 Plan 02: Dead CSS Audit and Animation Performance Summary

**Removed 32 dead CSS classes (277 lines) from app.css and replaced all 4 transition:all declarations with explicit compositor-safe property lists**

## What Was Done

### Task 1: Dead CSS Audit and Removal

Systematic audit of every custom CSS class in app.css against all .tsx/.ts files in frontend/src/.

**32 dead classes removed:**
- Utilities: hover-lift, click-scale, focus-gradient, animate-mesh, glass-subtle, card-premium, page-title, card-accent-top, stagger-1 through stagger-4
- Module gradients: gradient-tracker, gradient-messenger, gradient-calls, gradient-dashboard, gradient-settings
- Animation classes: animate-message-in, animate-modal-in, animate-pulse-subtle, animate-slide-in-right, animate-toast-in
- Components: typing-dots (+ .dot children + typing-bounce keyframe), widget-card, step-connector, table-striped, module-header-gradient (+ 5 strip variants), msg-hover-actions, composer-glow, send-glow, login-mobile-bg, dash-enter-1

**Result:** app.css reduced from 1340 to ~1012 lines (277 lines removed). All remaining custom classes verified as referenced in at least one .tsx/.ts file.

### Task 2: Animation Performance Audit

Replaced all `transition: all` declarations with explicit compositor-safe property lists:

| Selector | Before | After |
|----------|--------|-------|
| `button, a, input, select, textarea` (base) | `all 150ms` | `color, background-color, border-color, box-shadow, opacity 150ms` |
| `.kanban-card` | `all 260ms` | `transform, box-shadow, border-color, opacity 260ms` |
| `.kanban-fchip` | `all 180ms` | `transform, border-color, color 180ms` |
| `.dash-widget` | `all 280ms` | `transform, box-shadow, border-color 280ms` |

**Keyframe verification:**
- `dashEnterUp`: opacity + transform only (compositor-safe)
- `kanbanEnterUp`: opacity + transform only (compositor-safe)
- All other keyframes: use transform, opacity, filter, or box-shadow (acceptable)

## Deviations from Plan

None - plan executed exactly as written.

## Verification

1. Vite build passes with zero errors
2. `grep -c "transition.*all" frontend/src/app.css` returns 0
3. All @keyframes for Dashboard/Kanban use only transform+opacity
4. No unused custom CSS classes remain

## Commits

| Task | Commit | Description |
|------|--------|-------------|
| 1 | b392f63 | Remove 32 dead CSS classes from app.css |
| 2 | 3f2e3df | Replace transition:all with compositor-safe properties |
