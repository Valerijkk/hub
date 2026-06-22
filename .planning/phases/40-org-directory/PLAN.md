# Phase 40 — Org Directory (MOD-02)

**Status:** COMPLETE  
**Branch:** `feat/v2.5-trust` (uncommitted)  
**Requirement:** MOD-02 — tree by department; profile cards; search in Ctrl+K

## Goal

Справочник сотрудников Hub One v4: группировка по отделам, карточки профиля, интеграция с глобальным поиском.

## Deliverables

### Backend
- `GET /api/intranet/directory` — active users grouped by `department`
- `GET /api/intranet/directory/users/:userId` — profile card (404 for pending/deactivated)
- Extend `internal/intranet/` repository/service/handler
- Search user URLs → `/intranet/directory?user={id}`

### Frontend
- `/intranet/directory` — `OrgDirectory` component
- Sidebar «Справочник»
- Ctrl+K action «Справочник сотрудников»
- Profile dialog via `?user=` search param

### Tests
- `handler_test.go` — directory grouping + pending user 404
- `org-directory.test.tsx` — render + filter

## Verification

```bash
cd backend && go test ./internal/intranet/...
cd frontend && npx tsc --noEmit && npm run lint && npx vitest run
```

## Next

Phase 41 — Unified calendar (MOD-04)
