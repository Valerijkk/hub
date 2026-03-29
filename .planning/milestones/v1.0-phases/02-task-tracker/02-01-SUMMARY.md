---
phase: 02-task-tracker
plan: 01
subsystem: database
tags: [ent, schema, tracker, issue, workflow, kanban, lexorank, repository-pattern]

# Dependency graph
requires:
  - phase: 01-foundation
    provides: Ent ORM setup, Project/User schemas, repository pattern, outbox events
provides:
  - 12 new Ent schemas for tracker module (Issue, IssueLink, WorkflowStatus, WorkflowTransition, Comment, Reaction, ActivityLog, TimeLog, Label, LabelGroup, Board, CustomFieldDefinition)
  - Repository interface with 30+ methods for full tracker CRUD
  - Request/response DTOs for all tracker API endpoints
  - 8 NATS event types for tracker pub/sub
  - Generated Ent code for all new entities
affects: [02-02, 02-03, 02-04, 02-05]

# Tech tracking
tech-stack:
  added: [cristalhq/lexorank]
  patterns: [project-scoped workflow statuses, self-referencing issue hierarchy, JSONB custom fields, cursor-based pagination generic]

key-files:
  created:
    - backend/ent/schema/issue.go
    - backend/ent/schema/issuelink.go
    - backend/ent/schema/workflowstatus.go
    - backend/ent/schema/workflowtransition.go
    - backend/ent/schema/comment.go
    - backend/ent/schema/reaction.go
    - backend/ent/schema/activitylog.go
    - backend/ent/schema/timelog.go
    - backend/ent/schema/label.go
    - backend/ent/schema/labelgroup.go
    - backend/ent/schema/board.go
    - backend/ent/schema/customfielddefinition.go
    - backend/internal/tracker/dto.go
    - backend/internal/tracker/events.go
    - backend/internal/tracker/repository.go
  modified:
    - backend/ent/schema/project.go
    - backend/ent/schema/user.go

key-decisions:
  - "FieldChange struct defined both in schema (for Ent JSON storage) and in tracker DTOs (for API responses)"
  - "User schema extended with edges for comments, reactions, activity_logs, time_logs to support Ent back-references"

patterns-established:
  - "Tracker Repository interface mirrors project Repository: context-first, uuid.UUID params, *ent.Type returns"
  - "Internal input structs (CreateIssueInput) separate from HTTP request DTOs (CreateIssueRequest) -- uuid.UUID vs string"
  - "PaginatedResponse[T] generic for cursor-based pagination across all list endpoints"

requirements-completed: [TASK-01, TASK-02, TASK-03, TASK-04, TASK-05, TASK-11]

# Metrics
duration: 5min
completed: 2026-03-20
---

# Phase 02 Plan 01: Tracker Data Model Summary

**12 Ent schemas with 9 issue types, project-scoped workflow, LexoRank positioning, and 30+ method Repository interface contract**

## Performance

- **Duration:** 5 min
- **Started:** 2026-03-20T23:41:01Z
- **Completed:** 2026-03-20T23:46:18Z
- **Tasks:** 2
- **Files modified:** 19

## Accomplishments
- Created 12 new Ent schemas covering issues, workflow, comments, time tracking, labels, boards, and custom fields
- Defined comprehensive Repository interface with 30+ methods as the contract for all tracker CRUD operations
- All DTOs with json/validate tags for 16 request types and 17 response types including generic PaginatedResponse[T]
- 8 NATS event types covering issue lifecycle, comments, and timer events

## Task Commits

Each task was committed atomically:

1. **Task 1: Create all Ent schemas for tracker module + install lexorank** - `206d5f6` (feat)
2. **Task 2: Create tracker module skeleton (Repository interface, DTOs, events)** - `d763664` (feat)

## Files Created/Modified
- `backend/ent/schema/issue.go` - Core issue entity with 9 types, custom fields JSONB, rank, version
- `backend/ent/schema/issuelink.go` - Directional issue links with 11 link types
- `backend/ent/schema/workflowstatus.go` - Project-scoped workflow statuses with category
- `backend/ent/schema/workflowtransition.go` - Allowed transitions between statuses
- `backend/ent/schema/comment.go` - Issue comments with Plate JSON, mentions, internal flag
- `backend/ent/schema/reaction.go` - Emoji reactions on comments
- `backend/ent/schema/activitylog.go` - Field-level change tracking with FieldChange struct
- `backend/ent/schema/timelog.go` - Time tracking with timer support
- `backend/ent/schema/label.go` - Project-scoped labels with colors and groups
- `backend/ent/schema/labelgroup.go` - Label grouping for organization
- `backend/ent/schema/board.go` - Kanban boards with saved filter config
- `backend/ent/schema/customfielddefinition.go` - Custom field definitions with 8 field types
- `backend/ent/schema/project.go` - Added edges to all new tracker entities
- `backend/ent/schema/user.go` - Added edges for comments, reactions, activity_logs, time_logs
- `backend/internal/tracker/dto.go` - All request/response DTOs with validation
- `backend/internal/tracker/events.go` - 8 NATS event types
- `backend/internal/tracker/repository.go` - Repository interface + EntRepository skeleton

## Decisions Made
- Extended User schema with tracker edges (comments, reactions, activity_logs, time_logs) to support Ent's required back-reference pattern for `edge.From().Ref()` -- this was not in the plan but required by Ent code generation
- Defined FieldChange struct in both `schema/activitylog.go` (for Ent JSON column) and `tracker/dto.go` (for API responses) to keep packages decoupled

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Added Ref() to edge.From() declarations and User schema edges**
- **Found during:** Task 1 (Ent code generation)
- **Issue:** Ent code generation failed because edge.From("author/user", User.Type) in Comment, Reaction, ActivityLog, TimeLog lacked .Ref() and User schema lacked corresponding edge.To() declarations
- **Fix:** Added edge.To("comments/reactions/activity_logs/time_logs") to User schema and .Ref() to all back-reference edges
- **Files modified:** backend/ent/schema/user.go, backend/ent/schema/comment.go, backend/ent/schema/reaction.go, backend/ent/schema/activitylog.go, backend/ent/schema/timelog.go
- **Verification:** go generate ./ent succeeds, go build ./... passes
- **Committed in:** 206d5f6 (Task 1 commit)

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Auto-fix necessary for Ent code generation. No scope creep.

## Issues Encountered
None beyond the deviation documented above.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- All schemas generate cleanly, providing typed Ent client for Plans 02-05
- Repository interface contract defined, enabling parallel implementation of CRUD (Plan 02), workflow (Plan 03), comments/time (Plan 04)
- DTOs ready for HTTP handler implementation
- Event types ready for NATS integration

## Self-Check: PASSED

All 15 created files verified present. Both task commits (206d5f6, d763664) verified in git log.

---
*Phase: 02-task-tracker*
*Completed: 2026-03-20*
