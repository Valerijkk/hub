---
phase: 12-remaining-mockup-pages-unmocked-pages
plan: 04
subsystem: ui
tags: [onboarding, trash, tailwind, css-animations, mockup-matching]

requires:
  - phase: 08-design-system-tokens
    provides: CSS variables, design tokens, eu keyframe base patterns
provides:
  - Pixel-perfect onboarding 3-step flow (profile, tour, theme+accent)
  - Pixel-perfect trash page (table, badges, lifecycle, confirm dialog)
  - eu keyframe animation for enter-up transitions
  - scIn keyframe for dialog scale-in
  - onb-tour-card, onb-btn-next, onb-btn-go CSS utility classes
affects: [phase-13]

tech-stack:
  added: []
  patterns: [eu-keyframe-enter-up, tour-card-pseudo-highlight, gradient-nav-buttons]

key-files:
  created: []
  modified:
    - frontend/src/routes/onboarding.tsx
    - frontend/src/routes/_authenticated/trash.tsx
    - frontend/src/app.css

key-decisions:
  - "Removed Shadcn Button/Input/Card/Label deps from onboarding for exact mockup CSS match"
  - "Added 6 tour cards (matching mockup) instead of original 5 list items"
  - "Trash uses native table element with sticky headers per mockup pattern"
  - "Added confirm dialog with backdrop blur inline (no separate component)"

patterns-established:
  - "eu keyframe: opacity 0 + translateY(14px) to visible, reusable for page/card enter animations"
  - "scIn keyframe: scale(0.95) to scale(1) for dialog entrance"

requirements-completed: [ONBD-01, ONBD-02, ONBD-03, ONBD-04, ONBD-05, TRSH-01, TRSH-02, TRSH-03]

duration: 4min
completed: 2026-03-28
---

# Phase 12 Plan 04: Onboarding + Trash Restyle Summary

**Onboarding 3-step wizard (progress, profile form, tour cards, theme/accent) and trash table with days-remaining badges, restore/delete buttons, lifecycle about tab, and confirm dialog -- all matching mockups 10+11**

## Performance

- **Duration:** 4 min
- **Started:** 2026-03-28T13:17:35Z
- **Completed:** 2026-03-28T13:21:27Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments
- Onboarding page fully restyled: 3-step progress with active/done states, avatar upload, 2-col form grid, 3-col tour card grid with stagger animation, theme cards, accent dots, gradient nav buttons
- Trash page fully restyled: native table with sticky headers, type icons, days-remaining badges (OK/warn/danger), restore/delete buttons, about tab with lifecycle steps and type cards, confirm dialog with backdrop blur
- Added eu/scIn keyframes and onb-tour-card/onb-btn-next/onb-btn-go CSS utilities to app.css

## Task Commits

Each task was committed atomically:

1. **Task 1: Restyle Onboarding page** - `2143ded` (feat)
2. **Task 2: Restyle Trash page** - `532c590` (feat)

## Files Created/Modified
- `frontend/src/routes/onboarding.tsx` - Complete restyle: progress indicator, profile form, tour cards, theme/accent selection, gradient buttons
- `frontend/src/routes/_authenticated/trash.tsx` - Complete restyle: table layout, badges, action buttons, about tab, confirm dialog
- `frontend/src/app.css` - Added eu/scIn keyframes, onb-tour-card, onb-btn-next, onb-btn-go CSS classes

## Decisions Made
- Removed Shadcn Button/Input/Card/Label dependencies from onboarding for exact mockup CSS match (consistent with Phase 11 login approach)
- Changed from 5 list items to 6 tour cards in grid matching mockup exactly (tracker, messenger, calls, TMS, calendar, integrations)
- Trash uses native HTML table instead of card list for proper column alignment per mockup
- Confirm dialog implemented inline with state rather than as separate component

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical] Added confirm dialog to trash**
- **Found during:** Task 2 (Trash restyle)
- **Issue:** Original trash had no delete confirmation; mockup shows confirm dialog
- **Fix:** Added confirmProject state and overlay dialog with cancel/delete buttons
- **Files modified:** frontend/src/routes/_authenticated/trash.tsx
- **Verification:** TypeScript compiles, dialog matches mockup CSS
- **Committed in:** 532c590

---

**Total deviations:** 1 auto-fixed (1 missing critical)
**Impact on plan:** Essential UX safety for destructive action. No scope creep.

## Issues Encountered
- Pre-existing TypeScript errors in settings/profile.tsx (not related to this plan's changes, out of scope)

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Onboarding and Trash pages pixel-perfect per mockups
- All interactive functionality preserved (step nav, form state, restore mutation)
- Ready for remaining Phase 12 plans

---
*Phase: 12-remaining-mockup-pages-unmocked-pages*
*Completed: 2026-03-28*
