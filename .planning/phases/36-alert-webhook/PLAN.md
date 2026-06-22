# Phase 36: Alert Webhook → Audit Log

**Goal:** Close OBS-01 production loop — Alertmanager fires/resolves alerts appear in admin Audit Log.

**Branch:** `feat/v2.5-trust` (uncommitted)

## Tasks

- [x] `POST /api/webhooks/alertmanager` — Bearer `ALERT_WEBHOOK_SECRET`
- [x] `RecordAlertmanagerWebhook` → audit entries `alert.firing` / `alert.resolved`, entity `alert`
- [x] Skip HTTP audit middleware for webhook path
- [x] Config: `ALERT_WEBHOOK_SECRET` in `ServerConfig`
- [x] Wire in `main.go` via `RegisterAlertWebhook`
- [x] Frontend: audit filters + badge for alert actions
- [x] Deploy: `.env.example`, `docker-compose.prod.yml`, `alertmanager.prod.example.yml`
- [x] Tests: `handler_alerts_test.go` (auth, record, disabled when secret empty)

## Verification

```bash
cd backend && go test -count=1 ./internal/audit/...
cd backend && go build ./cmd/server/...
cd frontend && npx tsc --noEmit && npx vitest run
```

## Notes

- Webhook disabled when `ALERT_WEBHOOK_SECRET` empty (route not registered).
- No JWT — constant-time Bearer compare only.
- Hub constraint: no email alert channel; audit log is in-app admin visibility.

**Status:** COMPLETE 2026-06-24
