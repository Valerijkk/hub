# Phase 48 — MOD-07 1C Connector

**Status:** COMPLETE  
**Branch:** `feat/v2.5-trust`  
**Requirement:** MOD-07 — webhook export approved leave to 1C/ZUP

## Goal

When admin approves a leave request, Hub POSTs a JSON webhook to a configurable 1C endpoint. Admin configures URL, bearer token, and employee ID mapping field.

## Delivered

### Backend

- `backend/internal/integrations/onec/` — settings store, service, admin handler, user lookup
- `hr.LeaveExporter` interface — called from `ApproveLeave` (non-blocking; export errors logged, approval not rolled back)
- Routes: `GET/PUT /api/admin/integrations/onec/settings`, `POST .../test`
- Wiring in `main.go`: onec service passed to `hr.NewService(..., onecSvc)`

### Frontend

- Admin dashboard tab «1C» — `onec-settings.tsx`, `admin-onec.ts`

### Security

- `safehttp.ValidateOutboundURL` + `NewSafeHTTPClient` for webhook POST
- Bearer token stored in SystemSetting (not returned on GET)

## Payload (leave.approved)

```json
{
  "event": "leave.approved",
  "leave_id": "uuid",
  "employee": {
    "hub_user_id": "uuid",
    "email": "...",
    "phone": "...",
    "display_name": "...",
    "external_id": "mapped field"
  },
  "leave_type": "vacation",
  "start_date": "2026-07-01",
  "end_date": "2026-07-14",
  "reviewer_id": "uuid"
}
```

## Verification

```bash
cd backend && go build ./... && go test ./internal/integrations/onec/... ./internal/hr/...
cd frontend && npx tsc --noEmit
```

## Next

Phase 49+ — V4-04 Lark Base clone (large) · commit v4 block on `feat/v2.5-trust`
