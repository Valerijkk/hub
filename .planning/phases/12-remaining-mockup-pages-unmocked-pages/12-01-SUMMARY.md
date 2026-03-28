---
phase: 12-remaining-mockup-pages-unmocked-pages
plan: 01
subsystem: ui
tags: [tailwind, css, issue-detail, calls, calendar, mockup, redesign]

requires:
  - phase: 11-core-page-redesign-four-main-pages
    provides: Design system tokens, component primitives, sidebar/topbar styling
provides:
  - Issue Detail page restyled to mockup 05 (two-column layout, comments, activity, sidebar meta)
  - Calls page restyled to mockup 06 (calendar grid, event blocks, purple accent, controls)
  - calls-btn-p CSS utility for purple gradient button overlay
affects: [12-remaining-mockup-pages-unmocked-pages]

tech-stack:
  added: []
  patterns:
    - "Issue sidebar: void bg panel with uppercase field labels and hover field rows"
    - "Calendar week grid: border-b1 container with 52px sticky day headers and 60px time slots"
    - "Event blocks: r1 rounded with brightness/scale hover and border-left color accent"

key-files:
  created: []
  modified:
    - frontend/src/routes/_authenticated/projects/$projectId/issues/$issueKey.tsx
    - frontend/src/modules/tracker/issue-header.tsx
    - frontend/src/modules/tracker/issue-description.tsx
    - frontend/src/modules/tracker/issue-sidebar-meta.tsx
    - frontend/src/components/tracker/comment-list.tsx
    - frontend/src/components/tracker/activity-feed.tsx
    - frontend/src/components/tracker/checklist.tsx
    - frontend/src/components/tracker/issue-attachments.tsx
    - frontend/src/routes/_authenticated/calls.tsx
    - frontend/src/modules/calls/call-calendar.tsx
    - frontend/src/app.css

key-decisions:
  - "Activity feed uses colored dots instead of avatars for compact mockup-matching design"
  - "Comment layout changed from bordered cards to flex gap-14 rows with gradient avatar fallbacks"
  - "Sidebar meta panel uses direct styling (w-320px, bg-void, p-22px) matching mockup exactly"

patterns-established:
  - "Issue sidebar: uppercase 10.5px labels, 7px margin, hover field rows with r2 rounding"
  - "Calendar events: border-left accent color, brightness hover, ellipsis title overflow"

requirements-completed: [ISSU-01, ISSU-02, ISSU-03, ISSU-04, CALL-01, CALL-02, CALL-03, CALL-04]

duration: 7min
completed: 2026-03-28
---

# Phase 12 Plan 01: Issue Detail + Calls Page Redesign Summary

**Issue Detail restyled with two-column layout, gradient sidebar meta, styled comments/activity; Calls page with purple gradient title, calendar grid with event blocks and current-time indicator**

## Performance

- **Duration:** 7 min
- **Started:** 2026-03-28T13:18:09Z
- **Completed:** 2026-03-28T13:25:20Z
- **Tasks:** 2
- **Files modified:** 11

## Accomplishments
- Issue Detail page matches mockup 05 with mono breadcrumbs, 24px extrabold title, gradient attachment cards, green checklist progress, flex comment layout, colored activity dots, and void-bg sidebar meta panel
- Calls page matches mockup 06 with gradient purple-to-blue title, purple accent button with glow shadow, calendar controls with pill today button, view switcher with card active state, week grid with sticky headers and today indicator, event blocks with brightness/scale hover
- All event handlers, state management, mutations, and API calls untouched

## Task Commits

1. **Task 1: Restyle Issue Detail page** - `9bc6943` (feat)
2. **Task 2: Restyle Calls page** - `4387ba3` (feat)

## Files Created/Modified
- `frontend/src/routes/_authenticated/projects/$projectId/issues/$issueKey.tsx` - Entry animation, skeleton bg-card/50, overflow-y scroll on left column
- `frontend/src/modules/tracker/issue-header.tsx` - Mono breadcrumbs, blue key, t4 separators
- `frontend/src/modules/tracker/issue-description.tsx` - 24px extrabold title, 14px description text
- `frontend/src/modules/tracker/issue-sidebar-meta.tsx` - Void bg panel, uppercase labels, hover rows, pill badges
- `frontend/src/components/tracker/comment-list.tsx` - Flex gap-14 layout, gradient avatar, bold name, t4 time
- `frontend/src/components/tracker/activity-feed.tsx` - Colored dots, semibold names, t4 old/new values
- `frontend/src/components/tracker/checklist.tsx` - Green progress bar with glow, styled checkboxes
- `frontend/src/components/tracker/issue-attachments.tsx` - Gradient card backgrounds, hover translate/shadow
- `frontend/src/routes/_authenticated/calls.tsx` - Gradient title, purple button, flex layout
- `frontend/src/modules/calls/call-calendar.tsx` - Calendar nav, view switcher, week grid, event blocks, time line
- `frontend/src/app.css` - Added calls-btn-p CSS utility class

## Decisions Made
- Activity feed switched from avatar+border cards to colored-dot+borderless rows for compact mockup match
- Comment layout changed from bordered cards to flex gap-14 with gradient avatar fallbacks
- Sidebar meta uses direct void bg and pixel-exact padding rather than Shadcn Card wrapper
- Removed unused Badge/Avatar/Separator imports where replaced by direct styling

## Deviations from Plan

None - plan executed exactly as written.

## Known Stubs

None - all components render with real data sources.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Issue Detail and Calls pages fully restyled
- Ready for remaining mockup pages (TMS, Settings, Admin, Onboarding, Trash)

---
*Phase: 12-remaining-mockup-pages-unmocked-pages*
*Completed: 2026-03-28*
