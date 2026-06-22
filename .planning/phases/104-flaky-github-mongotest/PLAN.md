# Phase 104 — Flaky GitHub Fix + Mongo Testutil

**Branch:** `feat/v2.5-trust`  
**Depends on:** Phase 103  
**Goal:** Eliminate network-dependent github handler tests; unify MongoDB test setup for CI

## Scope

| Item | Status |
|------|--------|
| Flaky `IntegrationsCRUD` (real api.github.com) | Fixed via httptest mock in `newGitHubRoutesEnv` |
| `ListWorkflowRuns` service coverage | `service_wave104_test.go` mock API |
| Shared `internal/testutil/mongotest` | New package, refactored 4 mongo test files |
| CI `MONGODB_TEST_URL` | Already present in `.github/workflows/ci.yml` |

## Tasks

### Wave 104a — github flaky fix

- [x] `newGitHubRoutesEnv` uses `NewGitHubServiceWithHTTPClient` + empty workflow runs mock
- [x] `TestListWorkflowRuns_MockGitHubAPI` success path
- [x] `TestListWorkflowRuns_ContinuesOnAPIError` 403 → empty list

### Wave 104b — mongotest consolidation

- [x] `internal/testutil/mongotest/setup.go` — `Connect(t, prefix)`
- [x] Refactor: messenger, tms, calls, search mongo tests

## Deferred (Phase 105+)

- Local dev: document `docker compose -f deployments/docker-compose.test.yml up mongo`
- Commit + PR (on user request)

## Success criteria

- [x] `TestHandler_RegisterRoutes_IntegrationsCRUD` no network / no timeout
- [x] github ≥80%
- [x] race PASS github package
