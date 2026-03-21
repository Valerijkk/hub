---
phase: 04-tms-allure
plan: 03
subsystem: tms
tags: [allure, parser, matcher, mongodb, nats, testing, ci]

# Dependency graph
requires:
  - phase: 04-tms-allure/04-02
    provides: "TMS Ent entities, EntRepository, service, handler, frontend"
provides:
  - "Defensive Allure JSON parser (AllureResult types)"
  - "Multi-strategy test case matcher (annotation > fullName > class+method)"
  - "MongoDB raw Allure results storage"
  - "CombinedRepository (Ent + MongoDB)"
  - "Allure import API endpoints"
  - "Test run history and flaky test reports"
  - "NATS event on run completion"
affects: [04-tms-allure/04-04, 04-tms-allure/04-05]

# Tech tracking
tech-stack:
  added: [mongo-driver/v2]
  patterns: [CombinedRepository, defensive JSON parsing, multi-strategy matching, nil-safe NATS publish]

key-files:
  created:
    - backend/internal/tms/allure_types.go
    - backend/internal/tms/allure_parser.go
    - backend/internal/tms/allure_matcher.go
    - backend/internal/tms/allure_parser_test.go
    - backend/internal/tms/mongo_repository.go
    - backend/internal/tms/combined_repository.go
  modified:
    - backend/internal/tms/repository.go
    - backend/internal/tms/service.go
    - backend/internal/tms/handler.go
    - backend/internal/tms/dto.go
    - backend/internal/tms/service_test.go
    - backend/cmd/server/main.go

key-decisions:
  - "CombinedRepository pattern reused from messenger module for dual PG+Mongo storage"
  - "NATS publishEvent nil-safe pattern for testability without mock NATS client"
  - "Flaky detection via status variance across last 10 runs"

patterns-established:
  - "AllureResult types: all fields optional with pointers/omitempty for defensive parsing"
  - "Multi-strategy matcher: annotation > fullName > class+method priority order"
  - "Batch import: ParseAllureResults -> MatchTestCase -> BatchCreateTestResults pipeline"

requirements-completed: [ALLU-01, ALLU-02, ALLU-03]

# Metrics
duration: 7min
completed: 2026-03-21
---

# Phase 04 Plan 03: Allure Integration Summary

**Defensive Allure JSON parser with multi-strategy test case matching, MongoDB raw storage, import API, and run history/flaky reports**

## Performance

- **Duration:** 7 min
- **Started:** 2026-03-21T02:17:43Z
- **Completed:** 2026-03-21T02:25:04Z
- **Tasks:** 2
- **Files modified:** 12

## Accomplishments
- Defensive Allure parser handles valid, minimal, empty, malformed, and null-pointer JSON without panics
- Multi-strategy matcher finds test cases by annotation (link/label), fullName, or class+method with priority order
- MongoDB raw Allure JSON storage for auditability via CombinedRepository pattern
- Import API creates automatic test runs with matched/unmatched results, computes stats, publishes NATS event
- Run history and flaky test detection endpoints for reporting

## Task Commits

Each task was committed atomically:

1. **Task 1: Allure parser + matcher + types + tests (TDD)** - `5a74847` (feat)
2. **Task 2: MongoDB repository + combined repository + import handler + wiring** - `92c0488` (feat)

**Plan metadata:** (pending)

## Files Created/Modified
- `backend/internal/tms/allure_types.go` - Go structs for Allure JSON format (AllureResult, Step, Label, Link, etc.)
- `backend/internal/tms/allure_parser.go` - Defensive parser: ParseAllureResults, extractLabelValue, allureStatusToTMS
- `backend/internal/tms/allure_matcher.go` - Multi-strategy matcher: annotation > fullName > class+method
- `backend/internal/tms/allure_parser_test.go` - 14 TDD tests covering parser and matcher scenarios
- `backend/internal/tms/mongo_repository.go` - MongoDB CRUD for raw Allure results with indexes
- `backend/internal/tms/combined_repository.go` - Dual Ent+Mongo repository behind Repository interface
- `backend/internal/tms/repository.go` - Added StoreAllureResults, GetAllureResultsByRunID to interface
- `backend/internal/tms/service.go` - ImportAllureResults, ImportAllureToExistingRun, GetTestRunHistory, GetFlakyTests, publishEvent
- `backend/internal/tms/handler.go` - POST allure/import, POST runs/:runId/allure, GET reports/history, GET reports/flaky
- `backend/internal/tms/dto.go` - ImportAllureRequest, TestRunSummary DTOs
- `backend/internal/tms/service_test.go` - Updated mocks for new Repository interface
- `backend/cmd/server/main.go` - CombinedRepository wiring with MongoDB + NATS

## Decisions Made
- Reused CombinedRepository pattern from messenger module for consistency
- NATS client passed to TMSService constructor (nil-safe for tests without NATS)
- Flaky detection: compare statuses across last 10 runs per test case
- Moved TMS initialization after MongoDB client creation in main.go for proper dependency order

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Fixed existing service_test.go mock repository**
- **Found during:** Task 2
- **Issue:** NewTMSService constructor signature changed (added NATS param); mockRepository missing new interface methods
- **Fix:** Updated all NewTMSService calls with nil NATS; added StoreAllureResults/GetAllureResultsByRunID stubs to mock
- **Files modified:** backend/internal/tms/service_test.go
- **Verification:** All 25 tests pass
- **Committed in:** 92c0488 (Task 2 commit)

**2. [Rule 3 - Blocking] Moved TMS wiring after MongoDB connection in main.go**
- **Found during:** Task 2
- **Issue:** TMS CombinedRepository needs mongoDB variable which is created later in the file
- **Fix:** Moved TMS initialization block after messenger MongoDB setup
- **Files modified:** backend/cmd/server/main.go
- **Verification:** go build ./... succeeds
- **Committed in:** 92c0488 (Task 2 commit)

---

**Total deviations:** 2 auto-fixed (2 blocking)
**Impact on plan:** Both fixes necessary for compilation. No scope creep.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Allure import pipeline ready for CI integration
- NATS event tms.run.completed published on completion (ready for Telegram notification in 04-04/04-05)
- Run history and flaky test APIs available for frontend dashboard

---
*Phase: 04-tms-allure*
*Completed: 2026-03-21*
