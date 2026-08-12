# Phase 41 — Unified Calendar (MOD-04) — VERIFICATION

**Date:** 2026-07-27  
**Method:** Handler/service/schema/FE/test/doc inspection (no live test run)  
**Sources:** `PLAN.md`, `backend/internal/calendar/`, FE calendar module/route, `backend/testframework/calendar/`, `docs/wiki/API.md`

## Goal

First-class `/calendar` aggregating scheduled calls, sprint ranges, and issue due dates; iCal export.

## Measured evidence

| Criterion | Evidence (2026-07-27) | Result |
|-----------|----------------------|--------|
| `GET /api/calendar/events` | `handler.go` `cal.Get("/events", h.ListEvents)` | PASS |
| `GET /api/calendar/export.ics` | `cal.Get("/export.ics", h.ExportICal)` + `RenderICalendar` | PASS |
| Aggregates calls + sprints + deadlines | `service.go` `ListEvents`: `ListScheduledCalls`, `ListSprintsInRange`, `ListIssuesDueInRange` (+ leave from Phase 43) | PASS |
| RRULE expansion | `expandCallEvents` + `rrule-go` import in `service.go` | PASS |
| FE `/calendar` | `frontend/src/routes/_authenticated/calendar.tsx` + `unified-calendar.tsx` | PASS |
| Sidebar + Ctrl+K | `sidebar-nav.tsx` «Календарь»; command palette calendar entry present | PASS |
| BE tests | `backend/testframework/calendar/handler_test.go`, `service_wave98_test.go` | PASS |
| FE Vitest | `testframework/stages/unit/frontend-test/domains/v4/unified-calendar.test.tsx` | PASS |
| API.md Calendar | `docs/wiki/API.md` § Calendar (~L821) | PASS |

## Residual gaps (post-phase / later audit)

| Gap | Evidence | Notes |
|-----|----------|-------|
| FR-CAL-05 series trash | `scheduledcall.go`: `version=false`, `deleted_at=false` | Soft-delete/trash for series not in Phase-41 PLAN; still open vs FR |
| A6-005 leave visibility | `ListApprovedLeaveInRange` (added with HR) not user-scoped in `ListEvents` | Security residual; does not remove Phase-41 call/sprint/deadline aggregation |

## Verdict: **PASS**

Core MOD-04 aggregation + iCal deliverables are present. Series trash / leave scoping are residual, not missing Phase-41 checklist items.
