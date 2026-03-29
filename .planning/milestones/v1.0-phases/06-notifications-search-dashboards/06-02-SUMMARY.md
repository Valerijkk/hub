---
phase: 06-notifications-search-dashboards
plan: 02
subsystem: search
tags: [tsvector, gin-index, cmdk, zustand, mongodb-text, saved-filters, keyboard-shortcuts]

requires:
  - phase: 01-foundation
    provides: Ent ORM, Fiber HTTP, auth middleware, PostgreSQL + MongoDB connections
  - phase: 02-tracker
    provides: Issues, boards, workflow statuses for search and saved filters
  - phase: 03-messenger
    provides: MongoDB messages collection with text index for message search
  - phase: 04-tms
    provides: Test cases with case_key for search
provides:
  - Global search across issues, test_cases, users, projects, channels, files, messages
  - tsvector GENERATED STORED columns + GIN indexes on 5 PG tables
  - SavedFilter Ent schema with share functionality
  - Ctrl+K command palette with grouped results
  - Ctrl+N keyboard shortcut for issue creation
  - Saved filters CRUD for tracker
affects: [dashboards, tracker-filters]

tech-stack:
  added: [cmdk]
  patterns: [parallel-search-merge, tsvector-generated-stored, raw-sql-for-search]

key-files:
  created:
    - backend/ent/schema/savedfilter.go
    - backend/internal/search/indexer.go
    - backend/internal/search/dto.go
    - backend/internal/search/service.go
    - backend/internal/search/handler.go
    - backend/internal/search/service_test.go
    - frontend/src/api/search.ts
    - frontend/src/stores/search-store.ts
    - frontend/src/components/search/command-palette.tsx
    - frontend/src/components/search/search-results.tsx
    - frontend/src/components/search/saved-filters.tsx
  modified:
    - backend/ent/schema/user.go
    - backend/ent/schema/project.go
    - backend/cmd/server/main.go
    - backend/internal/platform/database/postgres.go
    - frontend/src/components/layout/topbar.tsx
    - frontend/src/components/layout/app-shell.tsx

key-decisions:
  - "Raw SQL DB connection for tsvector queries (Ent ORM does not support tsvector natively)"
  - "tsvector GENERATED ALWAYS AS ... STORED with 'simple' config for language-agnostic search"
  - "Parallel goroutines for PG + MongoDB search with mutex-protected result aggregation"
  - "Permission filtering in SQL WHERE clauses (project membership) to avoid N+1"
  - "cmdk library for command palette with shouldFilter=false (server-side filtering)"
  - "Custom event hub:create-issue for Ctrl+N decoupled from specific dialog component"

patterns-established:
  - "Raw SQL via database.OpenRawDB() for features Ent cannot express (tsvector, window functions)"
  - "Parallel search with sync.Mutex-protected append + sort-by-score merge"
  - "Keyboard shortcut registration in AppShell useEffect with cleanup"

requirements-completed: [SRCH-01, SRCH-02, UI-05]

duration: 12min
completed: 2026-03-21
---

# Phase 06 Plan 02: Global Search Summary

**Parallel PG tsvector + MongoDB $text search with Ctrl+K command palette, saved/shared filters, and Ctrl+N shortcut**

## Performance

- **Duration:** 12 min
- **Started:** 2026-03-21T03:58:33Z
- **Completed:** 2026-03-21T04:10:53Z
- **Tasks:** 2
- **Files modified:** 17

## Accomplishments
- Global search across 7 entity types (issues, test_cases, users, projects, channels, files, messages) with parallel PG + MongoDB queries
- tsvector GENERATED STORED columns + GIN indexes on 5 PostgreSQL tables for full-text search
- SavedFilter schema with owner/project edges, share toggle, and CRUD API
- Ctrl+K command palette with cmdk, grouped results by type, recent searches from localStorage
- Ctrl+N keyboard shortcut dispatching custom event for issue creation dialog

## Task Commits

1. **Task 1: Search backend** - `2ea6703` (feat)
2. **Task 2: Search frontend** - `8dbaada` (feat)

## Files Created/Modified
- `backend/ent/schema/savedfilter.go` - SavedFilter Ent schema (name, filter_config JSON, shared, owner/project edges)
- `backend/internal/search/indexer.go` - EnsureSearchIndexes: tsvector columns + GIN indexes via raw SQL
- `backend/internal/search/dto.go` - SearchResult, SearchRequest, SearchResponse, SavedFilterDTO, SavedFilterCreateRequest
- `backend/internal/search/service.go` - SearchService with parallel tsvector + $text search, SavedFilter CRUD
- `backend/internal/search/handler.go` - HTTP handler for /api/search and /api/filters routes
- `backend/internal/search/service_test.go` - Unit tests for merging, limit, UUID array conversion
- `backend/internal/platform/database/postgres.go` - Added OpenRawDB() for direct SQL access
- `backend/cmd/server/main.go` - Search service wiring with rawDB + EnsureSearchIndexes
- `frontend/src/api/search.ts` - API client for search and saved filters endpoints
- `frontend/src/stores/search-store.ts` - Zustand store with debounced search and localStorage recent searches
- `frontend/src/components/search/command-palette.tsx` - Ctrl+K command palette with cmdk
- `frontend/src/components/search/search-results.tsx` - SearchResultItem with type-specific icons
- `frontend/src/components/search/saved-filters.tsx` - SavedFilters CRUD component for tracker
- `frontend/src/components/layout/app-shell.tsx` - Global keyboard shortcuts + CommandPalette mount
- `frontend/src/components/layout/topbar.tsx` - Active search button replacing disabled placeholder

## Decisions Made
- Used `database.OpenRawDB()` for direct SQL access because Ent ORM does not support tsvector columns, GENERATED ALWAYS AS, or plainto_tsquery functions natively
- Used `'simple'` text search configuration instead of language-specific ones -- works for both Russian and English without stemming
- Channel search uses ILIKE instead of tsvector (channels table has only a `name` field)
- Dispatch `hub:create-issue` custom event for Ctrl+N rather than coupling to a specific dialog store
- Added notification edges to User schema (parallel agent's Notification schema referenced them but hadn't added them yet)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Added notification edges to User schema**
- **Found during:** Task 1 (go generate ./ent)
- **Issue:** Parallel agent (06-01) created Notification/NotificationPreference schemas referencing User.notifications and User.notification_preferences edges that didn't exist yet
- **Fix:** Added `edge.To("notifications", Notification.Type)` and `edge.To("notification_preferences", NotificationPreference.Type)` to User schema, removed duplicates caused by concurrent modifications
- **Files modified:** backend/ent/schema/user.go
- **Verification:** `go generate ./ent` succeeds
- **Committed in:** 2ea6703

**2. [Rule 3 - Blocking] Raw SQL DB for tsvector queries**
- **Found during:** Task 1 (Ent Client.Driver() is unexported)
- **Issue:** Ent's Client does not expose the underlying sql.Driver or *sql.DB; the `driver` field is unexported
- **Fix:** Created `database.OpenRawDB()` helper that opens a separate *sql.DB connection for raw SQL queries
- **Files modified:** backend/internal/platform/database/postgres.go, backend/cmd/server/main.go
- **Verification:** `go build ./internal/search/...` succeeds
- **Committed in:** 2ea6703

---

**Total deviations:** 2 auto-fixed (2 blocking)
**Impact on plan:** Both fixes necessary for the build to succeed. No scope creep.

## Issues Encountered
- Fiber v3 Ctx does not have QueryInt method -- used strconv.Atoi with c.Query() instead
- Ent ProjectMember schema uses edge-based FK (no explicit project_id field) -- used raw SQL JOIN for getUserProjectIDs

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Search infrastructure ready for dashboard integration (search widgets)
- SavedFilters component ready for tracker board/list page integration
- tsvector indexes will auto-update on data changes (GENERATED STORED)

---
*Phase: 06-notifications-search-dashboards*
*Completed: 2026-03-21*
