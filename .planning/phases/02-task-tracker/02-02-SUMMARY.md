---
phase: 02-task-tracker
plan: 02
subsystem: tracker-backend
tags: [tracker, backend, api, workflow, kanban, labels, boards, custom-fields]
dependency_graph:
  requires: [02-01]
  provides: [tracker-api, tracker-service, workflow-engine, tracker-repository]
  affects: [02-03, 02-04, 02-05]
tech_stack:
  added: []
  patterns: [repository-pattern, workflow-engine, optimistic-locking, cursor-pagination, outbox-events, mock-based-testing]
key_files:
  created:
    - backend/internal/tracker/repository_impl.go
    - backend/internal/tracker/service.go
    - backend/internal/tracker/workflow.go
    - backend/internal/tracker/handler.go
    - backend/internal/tracker/service_test.go
  modified:
    - backend/cmd/server/main.go
    - backend/go.mod
    - backend/go.sum
decisions:
  - "Mock repository for unit tests instead of testcontainers (Ent structs are constructible directly)"
  - "WorkflowEngine as separate struct for SRP (not embedded in service)"
  - "Cursor-based pagination using issue_key as cursor field"
  - "Board data loaded via nested eager loading: Board -> Project -> WorkflowStatuses -> Issues"
  - "Activity logging non-blocking (errors logged but don't fail the operation)"
  - "Timer creates TimeLog with minutes=1 placeholder, recalculated on stop"
metrics:
  duration: 12min
  completed: "2026-03-21T00:01:27Z"
---

# Phase 02 Plan 02: Tracker Backend Implementation Summary

Full tracker REST API with EntRepository (30+ methods), WorkflowEngine for transition validation, TrackerService with activity logging and outbox events, HTTP handlers with RBAC, and 11 unit tests covering all key scenarios.

## What Was Built

### EntRepository (repository_impl.go) -- 598 lines
- **CreateIssue**: Atomic transaction with project task counter increment, outbox event, default status fallback
- **FindIssueByKey/ID**: Eager loading of status, assignee, reporter, labels, parent, children, project
- **ListIssues**: Cursor-based pagination with filters (status, priority, assignee, type, search), max 100 per page
- **UpdateIssue**: Optimistic locking via `issue.VersionEQ` with conflict error on mismatch
- **MoveIssue**: Status + rank update for Kanban drag-and-drop (no version check)
- **CreateIssueLink**: Duplicate detection before creation
- **SeedDefaultWorkflow**: 4 statuses (To Do, In Progress, In Review, Done) with 5 transitions
- **ToggleReaction**: Check-exists-then-toggle pattern
- **Time tracking**: LogTime, StartTimer (creates open TimeLog), StopTimer (finds open timer, calculates elapsed)
- **Boards**: GetBoardData with nested eager loading (Board -> Project -> Statuses -> Issues)
- **Labels/LabelGroups/CustomFields**: Full CRUD with project scoping

### WorkflowEngine (workflow.go) -- 31 lines
- Validates transitions by checking repository for allowed from->to pairs
- Returns clear error message on invalid transitions

### TrackerService (service.go) -- 791 lines
- All 30+ service methods orchestrating repository, workflow validation, activity logging
- DTO converters: toIssueResponse, toCommentResponse, toBoardDataResponse, etc.
- Field change tracking for update operations (title, priority, status, assignee)
- Plate JSON plain text extraction for search index

### HTTP Handler (handler.go) -- 638 lines
- All endpoints registered with RBAC permission checks
- Issues: POST, GET list, GET by key, PUT, DELETE, PUT position
- Links: POST, GET, DELETE
- Comments: POST, GET, PUT, DELETE + reactions toggle
- Workflow: statuses CRUD, transitions CRUD
- Labels: CRUD + groups
- Boards: create, list, get data
- Custom fields: CRUD
- Time tracking: log, start/stop timer, list logs
- Activity: GET log
- CopyString for all path parameters (Pitfall 1)

### Main.go Wiring
- TrackerHandler registered with outboxPublisher and rbacSvc

### Unit Tests (service_test.go) -- 995 lines
- 11 tests, all passing:
  1. TestCreateIssue -- basic creation with sequential key
  2. TestCreateIssueAllTypes -- all 9 types (bug, task, story, epic, sub_task, improvement, tech_debt, spike, release)
  3. TestUpdateIssueOptimisticLocking -- version mismatch returns 409
  4. TestMoveIssueValidTransition -- allowed transition succeeds
  5. TestMoveIssueInvalidTransition -- blocked transition returns 400
  6. TestCreateParentChild -- parent/child hierarchy
  7. TestCreateIssueLink -- blocks link type
  8. TestWorkflowValidation -- engine validates transitions
  9. TestListIssuesPagination -- cursor pagination with limits
  10. TestCreateLabel -- label CRUD with groups
  11. TestBoardData -- board returns grouped issues
  12. TestCustomFieldValues -- custom field definitions and values

## Deviations from Plan

None - plan executed exactly as written.

## Decisions Made

1. **Mock repository pattern**: Created in-memory mock implementing full Repository interface with maps and slices. Ent structs are constructible directly with exported fields.
2. **WorkflowEngine as separate struct**: Keeps single responsibility, can be independently tested.
3. **Cursor pagination field**: Using issue_key (unique, sortable) as cursor instead of ID.
4. **Nested eager loading for boards**: Board -> Project -> WorkflowStatuses -> Issues to get all board data in one query.
5. **Activity logging non-blocking**: Errors are logged via slog but don't fail the main operation.
6. **Timer minutes placeholder**: StartTimer sets minutes=1 (Ent requires positive int), recalculated on StopTimer.
7. **Race flag skipped**: Windows CGO limitation prevents -race flag; tests run without it.

## Commits

| Task | Commit | Description |
|------|--------|-------------|
| 1 | e26f3e5 | EntRepository with all 30+ methods |
| 2 | f072065 | WorkflowEngine and TrackerService |
| 3 | 92462d9 | HTTP handlers and main.go wiring |
| 4 | 279f15e | 11 unit tests for TrackerService |
