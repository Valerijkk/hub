# Phase 41 — Unified Calendar (MOD-04)

**Status:** COMPLETE  
**Branch:** `feat/v2.5-trust` (uncommitted)  
**Requirement:** MOD-04 — `/calendar` sprints, calls, deadlines; iCal export

## Goal

First-class `/calendar` route aggregating scheduled calls, sprint ranges, and issue due dates.

## Deliverables

### Backend (`internal/calendar/`)
- `GET /api/calendar/events?from=&to=` — user-scoped aggregation
- `GET /api/calendar/export.ics` — RFC 5545 export
- RRULE expansion for recurring calls via `rrule-go`

### Frontend
- `/calendar` — `UnifiedCalendar` (reuses calls month/week/day/agenda views)
- Filter chips: calls / sprints / deadlines
- iCal download button
- Sidebar + Ctrl+K shortcut

### Tests
- `handler_test.go` — events + ical
- `unified-calendar.test.tsx`

## Verification

```bash
cd backend && go test ./internal/calendar/...
cd frontend && npx tsc --noEmit && npm run lint && npx vitest run
```

## Next

Phase 42 — CRM-lite (MOD-05) or Kudos (MOD-03) per PO order
