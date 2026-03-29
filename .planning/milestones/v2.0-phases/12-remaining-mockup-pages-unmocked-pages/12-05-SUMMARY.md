---
phase: 12-remaining-mockup-pages-unmocked-pages
plan: 05
subsystem: ui
tags: [tailwind, design-system, css-tokens, projects, sprints, jql, settings, dashboard]

requires:
  - phase: 08-design-system-foundation
    provides: Design system CSS variables and tokens (bg-base/card/raised, text-t1/t2/t3, border-b1, --r1..r4, --sh1..sh2)
  - phase: 10-ui-primitives
    provides: Restyled Shadcn primitives with design system tokens
  - phase: 11-page-redesigns
    provides: Dashboard widget grid patterns, Kanban card patterns, auth form patterns
provides:
  - All 6 unmocked page groups styled with design system tokens
  - Projects list with gradient card grid
  - Project overview with dashboard-style sections
  - Sprint management pages with consistent card/surface styling
  - JQL search with design system colors and chip filter pattern
  - Project settings with sidebar+form pattern
  - Project dashboard with template switcher
affects: []

tech-stack:
  added: []
  patterns:
    - "Unmocked pages extrapolate from established design system patterns"
    - "Gradient card pattern: bg-gradient-to-b from-card to-raised border border-b1 rounded-[var(--r4)]"
    - "Section title pattern: text-[13px] font-bold uppercase tracking-[0.06em] text-t3"

key-files:
  created: []
  modified:
    - frontend/src/routes/_authenticated/projects/index.tsx
    - frontend/src/routes/_authenticated/projects/$projectId/index.tsx
    - frontend/src/routes/_authenticated/projects/$projectId/dashboard.lazy.tsx
    - frontend/src/routes/_authenticated/projects/$projectId/sprints/index.tsx
    - frontend/src/routes/_authenticated/projects/$projectId/sprints/$sprintId.tsx
    - frontend/src/modules/tracker/sprint-board.tsx
    - frontend/src/modules/tracker/sprint-backlog.tsx
    - frontend/src/modules/tracker/sprint-retro.tsx
    - frontend/src/modules/tracker/burndown-chart.tsx
    - frontend/src/modules/tracker/velocity-chart.tsx
    - frontend/src/modules/tracker/planning-poker.tsx
    - frontend/src/components/search/jql-search-page.tsx
    - frontend/src/components/search/jql-editor.tsx
    - frontend/src/routes/_authenticated/projects/$projectId/settings/workflow.tsx
    - frontend/src/routes/_authenticated/projects/$projectId/settings/labels.tsx
    - frontend/src/routes/_authenticated/projects/$projectId/settings/integrations.tsx

key-decisions:
  - "Removed Card/CardHeader/CardContent Shadcn components in favor of plain divs with design system gradient classes for closer pattern match"
  - "Velocity chart restyled from Shadcn Card wrapper to inline gradient container matching burndown chart pattern"
  - "Planning poker cards use border-blue glow shadow for selected state instead of bg-primary"

patterns-established:
  - "Unmocked page animation: animate-[eu_450ms_var(--spring)_both] on page wrappers"
  - "Page title: text-[18px] font-extrabold tracking-[-0.03em] text-t1"
  - "Settings page header: title + description in mb-9 wrapper"

requirements-completed: [XTRA-01, XTRA-02, XTRA-03, XTRA-04, XTRA-05, XTRA-06, XTRA-07]

duration: 14min
completed: 2026-03-28
---

# Phase 12 Plan 05: Unmocked Pages Design System Tokens Summary

**All 6 unmocked page groups (Projects list, Project overview, Sprints, JQL search, Project settings, Project dashboard) styled with design system tokens matching mocked page patterns**

## Performance

- **Duration:** 14 min
- **Started:** 2026-03-28T13:17:05Z
- **Completed:** 2026-03-28T13:31:16Z
- **Tasks:** 2
- **Files modified:** 16

## Accomplishments
- Projects list page shows gradient cards in responsive grid with hover effects and design system typography
- Project overview uses dashboard-style card sections with uppercase tracking section titles and stat values
- Sprint management pages (board, backlog, detail, retro) use consistent gradient card pattern with design system tokens
- Burndown and velocity charts wrapped in gradient containers with section title pattern
- Planning poker cards use gradient back styling with blue glow for selected state
- JQL search uses design system color badges, chip filter pattern, and styled results table
- Project settings (workflow, labels, integrations) use sidebar+form pattern with gradient status items
- All old Shadcn defaults (bg-muted, text-muted-foreground, bg-background, text-foreground, text-primary, bg-primary, text-destructive, bg-popover) replaced with design system tokens

## Task Commits

Each task was committed atomically:

1. **Task 1: Restyle Projects list, Project overview, and Project dashboard** - `baacd3a` (feat)
2. **Task 2: Restyle Sprint pages, JQL, Project settings** - `320953b` (feat)

## Files Created/Modified
- `frontend/src/routes/_authenticated/projects/index.tsx` - Projects list page with gradient card grid
- `frontend/src/routes/_authenticated/projects/$projectId/index.tsx` - Project overview with dashboard-style sections
- `frontend/src/routes/_authenticated/projects/$projectId/dashboard.lazy.tsx` - Project dashboard with styled template switcher
- `frontend/src/routes/_authenticated/projects/$projectId/sprints/index.tsx` - Sprint list page wrapper with animation
- `frontend/src/routes/_authenticated/projects/$projectId/sprints/$sprintId.tsx` - Sprint detail with gradient header card
- `frontend/src/modules/tracker/sprint-board.tsx` - Sprint board with gradient stat cards and sprint cards
- `frontend/src/modules/tracker/sprint-backlog.tsx` - Backlog with design system issue rows
- `frontend/src/modules/tracker/sprint-retro.tsx` - Retro with gradient column containers
- `frontend/src/modules/tracker/burndown-chart.tsx` - Burndown chart in gradient container
- `frontend/src/modules/tracker/velocity-chart.tsx` - Velocity chart restyled from Card to gradient container
- `frontend/src/modules/tracker/planning-poker.tsx` - Poker cards with gradient back and blue glow
- `frontend/src/components/search/jql-search-page.tsx` - JQL search with design system badges and table
- `frontend/src/components/search/jql-editor.tsx` - JQL editor autocomplete with design system dropdown
- `frontend/src/routes/_authenticated/projects/$projectId/settings/workflow.tsx` - Workflow settings with gradient status items
- `frontend/src/routes/_authenticated/projects/$projectId/settings/labels.tsx` - Labels settings with color dot hover glow
- `frontend/src/routes/_authenticated/projects/$projectId/settings/integrations.tsx` - Integrations with gradient repo cards

## Decisions Made
- Removed Card/CardHeader/CardContent Shadcn components in favor of plain divs with design system gradient classes for closer pattern match with mocked pages
- Velocity chart restyled from Shadcn Card wrapper to inline gradient container matching burndown chart pattern
- Planning poker cards use border-blue glow shadow for selected state instead of bg-primary

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Removed unused Card component imports**
- **Found during:** Task 1 (after replacing Card components with plain divs)
- **Issue:** TypeScript error TS6192 for unused imports in projects/index.tsx and $projectId/index.tsx
- **Fix:** Removed unused Card, CardContent, CardDescription, CardHeader, CardTitle imports
- **Files modified:** projects/index.tsx, $projectId/index.tsx
- **Verification:** TypeScript compiles without errors in modified files
- **Committed in:** 320953b (Task 2 commit, with other Task 2 files)

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Necessary cleanup after replacing Shadcn Card with plain divs. No scope creep.

## Issues Encountered
- Pre-existing TypeScript errors in settings/profile.tsx, onboarding.tsx, and admin/audit-log.tsx are not caused by this plan's changes and are out of scope

## Known Stubs
None - all pages have real data sources wired through existing API calls and stores.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- All 6 unmocked page groups now use design system tokens consistently
- Phase 12 (remaining mockup pages + unmocked pages) is fully complete
- Ready for Phase 13 or UI review

## Self-Check: PASSED

All key files verified present. Both commit hashes (baacd3a, 320953b) confirmed in git log.

---
*Phase: 12-remaining-mockup-pages-unmocked-pages*
*Completed: 2026-03-28*
