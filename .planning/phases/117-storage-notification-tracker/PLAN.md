# Phase 117 — Storage RBAC + Notification Quiet Hours + Tracker Time

**Branch:** `feat/v2.5-trust`  
**Goal:** Push low packages toward 82%; maintain COV **≥81%** (from 81.6%).

## Targets

| Package | Before | After |
|---------|--------|-------|
| `storage` | 79.6% | **86.8%** |
| `notification` | 78.7% | **79.2%** |
| `tracker` | 79.1% | **79.5%** |
| COV overall | 81.6% | **81.8%** |

## Tasks

1. **Storage handler RBAC** — viewer forbidden upload/delete; stranger forbidden download; confirm invalid JSON; invalid user UUID
2. **Storage orphan cleanup** — S3 delete branch when client configured
3. **Notification service** — quiet hours same-day + invalid format; dispatch suppressed during quiet hours; preferences with project scope
4. **Notification handler** — list is_read filter; page/per_page defaults; push unsubscribe empty body
5. **Tracker service** — stop timer no active; log/start/list time issue not found
6. **Tracker handler** — create issue / issue link invalid JSON; stop timer invalid project

## New test files

- `backend/internal/storage/handler_wave117_test.go`, `service_wave117_test.go`
- `backend/internal/notification/handler_wave117_test.go`, `service_wave117_test.go`
- `backend/internal/tracker/handler_wave117_test.go`, `service_wave117_test.go`

## Verification (Docker)

```powershell
docker compose -f deployments/docker-compose.test.yml up -d

docker run --rm --network deployments_default `
  -v "${PWD}/backend:/app" -v hub-go-mod-cache:/go/pkg/mod -w /app `
  -e CGO_ENABLED=1 `
  -e POSTGRES_TEST_URL="postgresql://hub:hub@postgres:5432/hub?sslmode=disable" `
  -e MONGODB_TEST_URL="mongodb://hub:hub@mongo:27017/?authSource=admin" `
  golang:1.25-bookworm `
  sh -c "go test -race -count=1 -timeout=600s ./internal/... && go test -coverprofile=cov117_all ./internal/... && go tool cover -func=cov117_all | tail -1"
```

## Next (Phase 118)

- `notification` (79.2%) + `tracker` (79.5%) → 82%
- Frontend Vitest Wave 3
- Messenger CreateChannel NATS fan-out
