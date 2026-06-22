# Phase 43 — HR-lite (MOD-06)

**Status:** COMPLETE  
**Branch:** `feat/v2.5-trust` (uncommitted)  
**Requirement:** MOD-06 — Leave requests → admin approval → calendar block

## Goal

HR-lite: заявки на отсутствие, очередь согласования admin/founder, одобренные отпуска в unified calendar, уведомления при submit/decision.

## Deliverables

### Ent
- `LeaveRequest` — leave_type, status, start/end dates, reason, review_note, version, reviewed_at
- User edges: `leave_requests`, `leave_reviews`

### Backend (`internal/hr/`)
- CRUD flow: create, list mine, list pending (admin), approve/reject/cancel
- Overlap validation for pending/approved ranges
- `hrNotifierAdapter` → NotificationService on submit/decision

### Calendar integration
- `EventTypeLeave` in `internal/calendar`
- `ListApprovedLeaveInRange` in calendar repository

### Frontend
- `/hr` — form + my requests + admin approval queue
- Sidebar + Ctrl+K
- Unified calendar filter/color for `leave`

## Verification

```bash
cd backend && go build ./... && go test ./internal/hr/... ./internal/calendar/...
cd frontend && npx tsc --noEmit && npx vitest run src/modules/hr/
```

## Next

Phase 44 — Kudos / thanks (MOD-03)
