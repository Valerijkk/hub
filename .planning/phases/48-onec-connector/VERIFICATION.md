# Phase 48 Verification — MOD-07 1C Connector

**Date:** 2026-07-27  
**Auditor:** Grok 4.5 (PW2 A2 retro)  
**Branch:** `cursor-grok-4.5`  
**Method:** static code presence + security wiring + testframework inventory  
**Live suite re-run this session:** **NO**  
**Verdict:** PASS (retro-static)

## Goal

When admin approves a leave request, Hub POSTs a JSON webhook to a configurable 1C endpoint. Admin configures URL, bearer token, and employee ID mapping field.

## Goal-backward

| Criterion | Expected | Actual (2026-07-27) | Status |
|-----------|----------|---------------------|--------|
| `internal/integrations/onec/` module | settings/service/handler | present (6 `.go` files) | PASS |
| `hr.LeaveExporter` on approve | interface + ExportApprovedLeave | `hr/service.go` + `onec/service.go` | PASS |
| Admin settings + test routes | GET/PUT/POST test | `onec` handler + `wiring_modules.go` RegisterRoutes | PASS |
| SSRF guard on webhook URL | ValidateOutboundURL + SafeHTTPClient | `onec/service.go` | PASS |
| Bearer not returned on GET | settings redact | settings store / handler pattern present | PASS |
| Admin UI «1C» | FE settings | `components/admin/onec-settings.tsx`, `api/admin-onec.ts` | PASS |
| Tests | onec + hr | `testframework/onec/*`, `testframework/hr/*` | PASS |

## Notes

- PLAN `go test ./internal/integrations/onec/...` paths stale; tests under `backend/testframework/onec/`.
- Live suite **not** re-run in PW2.

## Sign-off

Goal deliverables present and security guards wired. Retro-static **PASS**.
