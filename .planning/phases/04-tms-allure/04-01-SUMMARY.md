---
phase: 04-tms-allure
plan: 01
subsystem: tms
tags: [ent, schemas, crud, versioning, test-management]
dependency_graph:
  requires: []
  provides: [tms-schemas, tms-service, tms-handler, tms-repository]
  affects: [backend/ent/schema, backend/cmd/server/main.go]
tech_stack:
  added: []
  patterns: [repository-interface, service-layer, handler-pattern, snapshot-versioning, run-state-machine]
key_files:
  created:
    - backend/ent/schema/testcase.go
    - backend/ent/schema/testcaseversion.go
    - backend/ent/schema/testsuite.go
    - backend/ent/schema/testrun.go
    - backend/ent/schema/testresult.go
    - backend/ent/schema/component.go
    - backend/internal/tms/repository.go
    - backend/internal/tms/ent_repository.go
    - backend/internal/tms/dto.go
    - backend/internal/tms/events.go
    - backend/internal/tms/service.go
    - backend/internal/tms/handler.go
    - backend/internal/tms/service_test.go
  modified:
    - backend/ent/schema/project.go
    - backend/ent/schema/user.go
    - backend/cmd/server/main.go
decisions:
  - "NewEntRepository(client) without OutboxPublisher -- TMS module does not need outbox events for Phase 4 plan 01"
  - "Offset/limit pagination for TMS (not cursor-based like tracker) -- simpler for test case browsing"
  - "Suite tree built in-memory from flat list -- practical for <1000 suites per project"
  - "Snapshot versioning with JSON map for flexibility -- allows rollback without schema migration"
metrics:
  duration: 10min
  completed: "2026-03-21T01:55:40Z"
---

# Phase 04 Plan 01: Test Case CRUD Backend Summary

TMS module with 6 Ent schemas, complete handler/service/repository, versioning via JSON snapshots, folder tree, run state machine, and 7 unit tests.

## Tasks Completed

### Task 1: Ent schemas + repository interface + DTOs + events
**Commit:** bf9dc58

- 6 Ent schemas: TestCase (classic+BDD, steps as JSONB), TestCaseVersion (snapshot versioning), TestSuite (self-referencing tree), TestRun (status counts), TestResult (with Allure fields), Component
- Project schema: added `test_case_counter` field and 4 TMS edges (test_cases, test_runs, test_suites, components)
- User schema: added 4 TMS back-reference edges (created_test_cases, created_runs, executed_results, test_case_versions)
- Repository interface: 25+ methods covering all CRUD, versioning, search, batch operations
- DTOs: 15 request/response types with JSON tags and validation
- Events: 4 NATS event types (TestRunCreated, TestRunCompleted, CoverageUpdated, TestCaseLinked)
- Ent codegen passes, backend compiles

### Task 2: Service + EntRepository + Handler + wiring + tests
**Commit:** fb8e5a3

- EntRepository: full Ent implementation with atomic counter increment in transaction, eager-loading, bulk insert for Allure import
- TMSService: versioning creates snapshot BEFORE update, rollback applies snapshot as update, run state machine (created->in_progress->completed/aborted)
- Handler: 22 HTTP endpoints under /projects/:projectId/tms with RBAC permissions, Swagger annotations, Russian GoDoc
- Wired into main.go following existing module pattern
- 7 unit tests with mock repository: classic creation, BDD creation, version snapshot on update, rollback, valid transitions, invalid transition rejection, reopen

## Deviations from Plan

None -- plan executed exactly as written.

## Verification Results

- `go generate ./ent` -- passes (6 new schemas generate correctly)
- `go build ./...` -- compiles without errors
- `go test ./internal/tms/... -v -count=1` -- all 7 tests pass
- TMS routes registered in main.go
- Swagger annotations on all 22 handler methods
