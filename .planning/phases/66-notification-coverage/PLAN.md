# Phase 66 — Notification Coverage Uplift (42.5% → 50%)

**Branch:** `feat/v2.5-trust`  
**Status:** COMPLETE  
**Date:** 2026-06-25

## Goal

Close COV gap: `notification/` package **≥50%** + telegram handler route fix.

## Deliverables

### Backend tests
- `handler_telegram_bind_test.go` — bind/status/confirm/unlink/test HTTP
- `repository_extra_test.go` — channel/members, telegram fields, assignee helper
- `consumer_extra_test.go` — blocker, DM message, mention, call started
- `push_lifecycle_test.go` — Start/Stop, nil-safe Send, VAPIDPublicKey

### Bug fix
- `handler.go` RegisterRoutes: static paths (`/telegram`, `/read-all`) before `/:id`

## Results

| Metric | Before | After |
|--------|--------|-------|
| `notification/` coverage | 42.5% | **55.7%** |

## Verification

```bash
cd backend && go test -count=1 -coverprofile=cov_notif.txt ./internal/notification/...
go tool cover -func cov_notif.txt
```

## Success criteria

- [x] notification ≥50%
- [x] all notification tests pass
- [x] telegram DELETE /notifications/telegram not shadowed by /:id
