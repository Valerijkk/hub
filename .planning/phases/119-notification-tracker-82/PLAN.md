# Phase 119 — Notification/Tracker Push to 82%

**Branch:** `feat/v2.5-trust`  
**Goal:** Push `notification`/`tracker` toward **82%**; maintain COV **≥81%** (from 81.9%).

## Targets

| Package | Before | After |
|---------|--------|-------|
| `notification` | 80.2% | **81.9%** |
| `tracker` | 80.1% | **80.7%** |
| COV overall | 81.9% | **82.0%** |

## Tasks

1. **Notification consumer** — comment added invalid UUIDs, no assignee, self-mention skip; blocker invalid UUIDs + no assignee; processEvent run_completed + link_created
2. **Notification push** — nil Start/Stop, queue overflow, sendToUser no subscriptions
3. **Notification handler** — MarkRead unauthorized; invalid is_read filter ignored
4. **Notification service** — read/unread list split; urgent flag persisted; mark/delete no-op on missing ID
5. **Tracker handler (routes env)** — comments/activity flow; message-links validation; thread links query; automation not-found; log time invalid JSON
6. **Tracker automation service** — UpdateRule success; configString non-string coercion

## New test files

- `backend/internal/notification/consumer_wave119_test.go`
- `backend/internal/notification/push_wave119_test.go`
- `backend/internal/notification/handler_wave119_test.go`
- `backend/internal/notification/service_wave119_test.go`
- `backend/internal/tracker/handler_wave119_test.go`
- `backend/internal/tracker/service_wave119_test.go`

## Verification (Docker)

```powershell
docker compose -f deployments/docker-compose.test.yml up -d

docker run --rm --network deployments_default `
  -v "${PWD}/backend:/app" -v hub-go-mod-cache:/go/pkg/mod -w /app `
  -e CGO_ENABLED=1 `
  -e POSTGRES_TEST_URL="postgresql://hub:hub@postgres:5432/hub?sslmode=disable" `
  -e MONGODB_TEST_URL="mongodb://hub:hub@mongo:27017/?authSource=admin" `
  golang:1.25-bookworm `
  sh -c "go test -race -count=1 -timeout=600s ./internal/... && go test -coverprofile=cov119_all ./internal/... && go tool cover -func=cov119_all | tail -1"
```

## Next (Phase 120)

- Push `tracker` to **≥82%** (comments/reactions/time handlers, automation engine)
- `tms`/`project` packages toward 82%
- Frontend Vitest Wave 3
