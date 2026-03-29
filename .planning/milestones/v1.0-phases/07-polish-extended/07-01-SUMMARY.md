---
phase: 07-polish-extended
plan: 01
subsystem: tracker-scrum-kanban
tags: [sprint, scrum, kanban, planning-poker, burndown, velocity, wip-limits, swimlanes]
dependency_graph:
  requires: [tracker, ent, redis, nats]
  provides: [sprint-management, planning-poker, burndown-chart, velocity-chart, wip-limits, swimlanes]
  affects: [tracker-service, board-schema, issue-schema, workflowstatus-schema, project-schema, main]
tech_stack:
  added: [recharts-areachart, recharts-barchart]
  patterns: [wip-checker-interface, poker-redis-nats, fibonacci-voting, swimlane-grouping]
key_files:
  created:
    - backend/ent/schema/sprint.go
    - backend/internal/sprint/dto.go
    - backend/internal/sprint/repository.go
    - backend/internal/sprint/service.go
    - backend/internal/sprint/poker.go
    - backend/internal/sprint/handler.go
    - backend/internal/sprint/service_test.go
    - frontend/src/api/sprint.ts
    - frontend/src/stores/sprint-store.ts
    - frontend/src/modules/tracker/sprint-board.tsx
    - frontend/src/modules/tracker/sprint-backlog.tsx
    - frontend/src/modules/tracker/planning-poker.tsx
    - frontend/src/modules/tracker/burndown-chart.tsx
    - frontend/src/modules/tracker/velocity-chart.tsx
    - frontend/src/modules/tracker/sprint-retro.tsx
    - frontend/src/modules/tracker/kanban-board.tsx
    - frontend/src/routes/_authenticated/projects/$projectId/sprints/index.tsx
    - frontend/src/routes/_authenticated/projects/$projectId/sprints/$sprintId.tsx
    - frontend/src/routes/_authenticated/projects/$projectId/backlog.tsx
  modified:
    - backend/ent/schema/issue.go
    - backend/ent/schema/workflowstatus.go
    - backend/ent/schema/board.go
    - backend/ent/schema/project.go
    - backend/internal/tracker/service.go
    - backend/cmd/server/main.go
    - frontend/src/routeTree.gen.ts
decisions:
  - WIPChecker interface injected into TrackerService for loose coupling between sprint and tracker modules
  - Planning Poker via Redis hashes with 5min TTL + NATS events for real-time broadcast
  - Fibonacci voting values validated server-side, votes hidden until explicit reveal
  - SprintService.CompleteSprint auto-carries incomplete issues to next planning sprint
  - Burndown uses linear interpolation for past days, current remaining for future days
  - Swimlanes implemented as frontend-only grouping (no backend changes needed)
  - Recharts used directly for burndown AreaChart and velocity BarChart (consistent with Phase 6 pattern)
metrics:
  duration: 15min
  completed: "2026-03-21T08:23:00Z"
  tasks_completed: 2
  tasks_total: 2
  files_created: 20
  files_modified: 7
  test_count: 10
---

# Phase 07 Plan 01: Sprint Management & Kanban Enhancements Summary

Sprint CRUD with planning/active/completed lifecycle, Planning Poker via Redis+NATS with Fibonacci voting, burndown/velocity Recharts charts, WIP-limits on workflow statuses blocking MoveIssue, and swimlanes by assignee/priority/type on Kanban board.

## What Was Built

### Backend (Task 1)

**Sprint Ent Schema** (`backend/ent/schema/sprint.go`):
- Fields: id, name, goal, status (planning/active/completed), start_date, end_date, retro_notes (JSONB), timestamps
- Edges: project (M2O), issues (O2M via Issue.sprint)
- Indexes: (project_id, status), (project_id, start_date)

**Issue Schema Update**: Added optional `sprint` edge (M2O to Sprint) for nullable sprint_id FK.

**WorkflowStatus Schema Update**: Added `wip_limit` int field (default 0 = no limit) for KANB-02.

**Board Schema Update**: Added `swimlane_mode` enum (none/assignee/priority/type) for KANB-03.

**Sprint Module** (`backend/internal/sprint/`):
- `repository.go`: EntRepository with CRUD, FindActiveSprint, FindOrCreateNextPlanningSprint, FindIncompleteIssues, MoveIssuesToSprint, burndown/velocity data queries
- `service.go`: SprintService with CreateSprint, StartSprint (rejects when active exists), CompleteSprint (auto-carry incomplete issues), UpdateRetroNotes, GetBurndown, GetVelocity, CheckWIPLimit
- `poker.go`: PokerService with StartPokerSession, CastVote (Fibonacci validation, hidden votes), RevealVotes, GetSession -- Redis hashes + NATS events
- `handler.go`: Fiber HTTP handlers for all sprint and poker endpoints
- `dto.go`: Request/response DTOs for sprint, poker, burndown, velocity

**WIP Limit Integration**: WIPChecker interface on TrackerService, injected from SprintService via SetWIPChecker. MoveIssue checks target status WIP limit before allowing move.

**10 Unit Tests**: CreateSprint, EndDateValidation, StartSprint success/rejection, CompleteSprint with auto-carry, WIP limit enforcement (block/allow/no-limit), GetBurndown, GetVelocity.

### Frontend (Task 2)

**API Client** (`frontend/src/api/sprint.ts`): Full sprint CRUD, poker session management, burndown, velocity.

**Zustand Store** (`frontend/src/stores/sprint-store.ts`): activeSprint, sprints, burndown, velocity, pokerSession state with fetch/mutation actions.

**Components**:
- `sprint-board.tsx`: Sprint lifecycle management with start/complete buttons and confirmation dialog
- `sprint-backlog.tsx`: Issue list with checkbox selection and batch move to sprint
- `burndown-chart.tsx`: Recharts AreaChart with ideal (dashed) vs remaining lines
- `velocity-chart.tsx`: Recharts BarChart with committed/completed bars per sprint
- `sprint-retro.tsx`: Three-section retro form (went_well/to_improve/actions) saving as JSONB
- `planning-poker.tsx`: Fibonacci card UI, hidden votes, reveal with average/median stats
- `kanban-board.tsx`: WIP limit display (count/limit with red border), swimlane mode selector

**Routes**: /projects/:projectId/sprints, /sprints/:sprintId, /backlog with TanStack Router.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Fixed Ent codegen environment Label collision**
- Found during: Task 1, go generate
- Issue: Environment schema "env_label" field produced conflicting `Label` constant in generated code
- Fix: Re-ran `go generate ./ent` which regenerated with correct field name
- Files modified: backend/ent/environment/* (generated)

## Verification Results

- `go test ./internal/sprint/... -v` -- 10/10 tests pass
- `go build ./internal/sprint/...` -- sprint module compiles
- `npx tsc --noEmit` -- no TypeScript errors in sprint/kanban files
- `npx vite build` -- frontend builds successfully

## Notes

- Full server build (`go build ./cmd/server`) has pre-existing errors in github module (not related to this plan)
- Pre-existing CallChat.tsx TypeScript errors (not related to this plan)
