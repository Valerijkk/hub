# Phase 46 — Public status page (V4-06)

**Status:** COMPLETE  
**Branch:** `feat/v2.5-trust` (uncommitted)  
**Requirement:** V4-06 — Public status page

## Goal

Публичная страница `/status` без авторизации: статус Postgres, Redis, Mongo, NATS, API.

## Deliverables

### Backend (`internal/status/`)
- `GET /api/status` — aggregated probes, 503 if degraded

### Frontend
- `/status` — public route, auto-refresh 30s
- Link to login

## Verification

```bash
cd backend && go test ./internal/status/...
cd frontend && npx vitest run src/modules/status/
```

## Next

Hub One v4 block complete · commit feat/v2.5-trust · MOD-07 1C (P3) or v4.5 milestones
