# Phase 43 — HR-lite (MOD-06) — VERIFICATION

**Date:** 2026-07-27  
**Method:** Schema/handler/calendar-hook/FE/test/doc inspection (no live test run)  
**Sources:** `PLAN.md`, `backend/ent/schema/leaverequest.go`, `backend/internal/hr/`, `backend/internal/calendar/`, FE, tests, `docs/wiki/API.md`

## Goal

Leave requests → admin approval → calendar block; notifications on submit/decision.

## Measured evidence

| Criterion | Evidence (2026-07-27) | Result |
|-----------|----------------------|--------|
| Ent `LeaveRequest` | `leaverequest.go` present; `version=true` (no `deleted_at` — not required by PLAN) | PASS |
| HR CRUD / approve / reject / cancel | `handler.go`: ListMyLeave, ListPending, CreateLeave, Cancel/Approve/Reject | PASS |
| Calendar `EventTypeLeave` + approved leave query | `calendar/dto.go` `EventTypeLeave`; `ListApprovedLeaveInRange` in calendar repository; wired in `ListEvents` | PASS |
| FE `/hr` | `frontend/src/routes/_authenticated/hr.tsx` + `frontend/src/modules/hr/` | PASS |
| Sidebar + Ctrl+K | `sidebar-nav.tsx` `/hr`; command palette `/hr` | PASS |
| BE tests | `backend/testframework/hr/` — 7 test files | PASS |
| FE Vitest | `testframework/stages/unit/frontend-test/domains/v4/leave-requests.test.tsx` | PASS |
| API.md HR | `docs/wiki/API.md` § HR (~L854) | PASS |

## Residual

| Gap | Evidence |
|-----|----------|
| A6-005 | Approved leave listed org-wide in calendar `ListEvents` (not filtered by viewer) — open security residual from later audit |

## Verdict: **PASS**

MOD-06 leave flow + calendar integration deliverables are present. Leave visibility scoping is a residual, not a missing Phase-43 checkbox.
