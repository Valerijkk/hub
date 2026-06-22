# Phase 118 — Notification Consumer + Tracker Automation Push

**Branch:** `feat/v2.5-trust`  
**Goal:** Push `notification`/`tracker` to **≥80%**; maintain COV **≥81%** (from 81.8%).

## Targets

| Package | Before | After |
|---------|--------|-------|
| `notification` | 79.2% | **80.2%** |
| `tracker` | 79.5% | **80.1%** |
| `storage` | 86.8% | 86.8% |
| COV overall | 81.8% | **81.9%** |

## Tasks

1. **Notification consumer** — status changed no assignee/actor-is-assignee; call/message short subject; self-mention skip; run completed invalid UUIDs; comment invalid JSON
2. **Notification handler** — unauthorized unread-count/preferences; getUserID uuid.UUID locals
3. **Notification service** — project-scoped preference disabled; invalid timezone quiet hours; empty preferences list
4. **Tracker handler** — automation/board/custom-field invalid JSON; delete issue not found; invalid project IDs (routes env)
5. **Tracker service** — custom field duplicate; automation list empty + empty name; swimlane unknown branch

## New test files

- `backend/internal/notification/consumer_wave118_test.go`
- `backend/internal/notification/handler_wave118_test.go`
- `backend/internal/notification/service_wave118_test.go`
- `backend/internal/tracker/handler_wave118_test.go`
- `backend/internal/tracker/service_wave118_test.go`

## Verification (Docker)

```powershell
docker compose -f deployments/docker-compose.test.yml up -d

docker run --rm --network deployments_default `
  -v "${PWD}/backend:/app" -v hub-go-mod-cache:/go/pkg/mod -w /app `
  -e CGO_ENABLED=1 `
  -e POSTGRES_TEST_URL="postgresql://hub:hub@postgres:5432/hub?sslmode=disable" `
  -e MONGODB_TEST_URL="mongodb://hub:hub@mongo:27017/?authSource=admin" `
  golang:1.25-bookworm `
  sh -c "go test -race -count=1 -timeout=600s ./internal/... && go test -coverprofile=cov118_all ./internal/... && go tool cover -func=cov118_all | tail -1"
```

## Next (Phase 119)

- Push `notification`/`tracker` toward **82%**
- Frontend Vitest Wave 3
- Messenger CreateChannel NATS fan-out
