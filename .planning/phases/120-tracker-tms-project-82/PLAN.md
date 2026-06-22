# Phase 120 — Tracker/TMS/Project Push to 82%

**Branch:** `feat/v2.5-trust`  
**Goal:** Push `tracker`/`tms`/`project` toward **82%**; maintain COV **≥82%** (from 82.0%).

## Targets

| Package | Before | After |
|---------|--------|-------|
| `tracker` | 80.7% | **81.2%** |
| `tms` | 79.0% | **80.6%** |
| `project` | 80.1% | **80.4%** |
| COV overall | 82.0% | **82.1%** |

## Tasks

1. **Tracker routes env** — comment reaction/update/delete, MoveIssue, label groups, timer, automation CRUD, issue links, log time
2. **Tracker repository** — CreateComment outbox + mentions, DeleteRule success
3. **Project handler** — RemoveMember invalid ID, ListTeams, CreateInvite invalid JSON, ListMembers
4. **Project service (ent)** — ListProjects member count, archive/unarchive flow, DeleteProject not found
5. **TMS handler** — CompareEnvironments validation, reports (flaky/traceability/defects), ListSuites/Runs
6. **TMS ent/service** — UpdateEnvironment, GetLatestTestRunForEnvironment, default ListTestRuns limit, GetCoverageReport empty

## New test files

- `backend/internal/tracker/handler_wave120_test.go`
- `backend/internal/tracker/repository_wave120_test.go`
- `backend/internal/project/handler_wave120_test.go`
- `backend/internal/project/service_wave120_test.go`
- `backend/internal/tms/handler_wave120_test.go`
- `backend/internal/tms/service_wave120_test.go`

## Verification (Docker)

```powershell
docker compose -f deployments/docker-compose.test.yml up -d

docker run --rm --network deployments_default `
  -v "${PWD}/backend:/app" -v hub-go-mod-cache:/go/pkg/mod -w /app `
  -e CGO_ENABLED=1 `
  -e POSTGRES_TEST_URL="postgresql://hub:hub@postgres:5432/hub?sslmode=disable" `
  -e MONGODB_TEST_URL="mongodb://hub:hub@mongo:27017/?authSource=admin" `
  golang:1.25-bookworm `
  sh -c "go test -race -count=1 -timeout=600s ./internal/... && go test -coverprofile=cov120_all ./internal/... && go tool cover -func=cov120_all | tail -1"
```

## Next (Phase 121)

- **tracker → ≥82%**: GetBoardData swimlanes via routes, CreateLabel/DeleteLabel, workflow transition handlers
- **tms → ≥82%**: handler_runs ListTestResults/ExecuteTestResult, ent CreateTestResult edges
- **project → ≥82%**: AddTeamToProject flow, team member CRUD handlers
