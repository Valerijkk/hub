# Phase 116 — Search PG Integration + Messenger/TMS Push

**Branch:** `feat/v2.5-trust`  
**Goal:** COV **≥81%** (from 80.7%).

## Targets

| Package | Before | After |
|---------|--------|-------|
| `search` | 76% | **92.2%** |
| `messenger` | 77% | **81.7%** |
| `tms` | 79% | **80.3%** |
| `notification` | 78.5% | **78.7%** |
| `ai` | 79% | **80.5%** |
| `bases` | 80% | **82.7%** |
| COV overall | 80.7% | **81.6%** |

## Tasks

1. **Search PG integration** — tsvector shards: issues, test_cases, files, users (admin/scoped), channels; mutex-serialized schema reset; `EnsureSearchIndexes` idempotency
2. **Search handler/JQL** — empty query, share invalid JSON, JQL parse edge cases
3. **Messenger service** — UpdateChannel/DeleteChannel forbidden paths, AddMembers RBAC, ToggleStar, UnmuteChannel
4. **Messenger handler** — JoinChannel/AddMembers/ListMembers invalid input
5. **TMS service** — UpdateRunStatus not-found, AddTestCasesToRun invalid case ID, StartTestRun not-found
6. **Notification handler** — List invalid per_page clamp
7. **Bugfix** — `searchFiles`: scan UUID `project_files` into `string` (was `*string`, rows silently dropped)

## New test files

- `backend/internal/search/service_wave116_postgres_test.go`
- `backend/internal/search/handler_wave116_test.go`
- `backend/internal/search/jql_wave116_test.go`
- `backend/internal/messenger/service_wave116_test.go`, `handler_wave116_test.go`
- `backend/internal/tms/service_wave116_test.go`
- `backend/internal/notification/handler_wave116_test.go`

## Verification (Docker — avoids Windows Firewall prompts)

```powershell
docker compose -f deployments/docker-compose.test.yml up -d

docker run --rm --network deployments_default `
  -v "${PWD}/backend:/app" -v hub-go-mod-cache:/go/pkg/mod -w /app `
  -e CGO_ENABLED=1 `
  -e POSTGRES_TEST_URL="postgresql://hub:hub@postgres:5432/hub?sslmode=disable" `
  -e MONGODB_TEST_URL="mongodb://hub:hub@mongo:27017/?authSource=admin" `
  golang:1.25-bookworm `
  sh -c "go test -race -count=1 -timeout=600s ./internal/... && go test -coverprofile=cov116_all ./internal/... && go tool cover -func=cov116_all | tail -1"
```

## Next (Phase 117)

- `storage` (79.6%) + `notification` (78.7%) + `tracker` (79.1%) push toward 82%
- Messenger `CreateChannel` NATS fan-out + poll vote repo paths
- Vitest/Playwright Wave 3 (frontend)
