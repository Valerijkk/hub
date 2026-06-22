# Phase 107 — CRM + Dashboard + Bases Boundary Push

**Branch:** `feat/v2.5-trust`  
**Goal:** Push `crm`, `dashboard`, `bases` toward 79%+ and lift overall COV past 78.8%.

## Targets

| Package | Before | Target |
|---------|--------|--------|
| `crm` | 76.7% | ≥78% |
| `dashboard` | 77.6% | ≥80% |
| `bases` | 78.1% | ≥78.5% |
| COV overall | 78.8% | ≥78.9% |

## Tasks

1. **CRM handler errors** — version conflict, invalid JSON, delete not-found, invalid stage (`handler_wave107_test.go`)
2. **Dashboard service paths** — personal dashboard with tasks/deadlines/time; project stats burndown/velocity/activity; recent test runs (`service_wave107_test.go`)
3. **Bases handler paths** — version conflict, delete not-found, table view (`handler_wave107_test.go`)

## Patterns

- Dashboard ent env: `newDashboardRepoEnv(t)` + miniredis
- ActivityLog links via Issue, not Project
- TestRun: `SetCreatedByID`, not `SetCreatedBy`
- Bases: reuse `setupBasesTest` / `createBaseViaAPI`; avoid duplicating `TestHandler_ListBases`

## Verification

```powershell
cd backend
go test -count=1 ./internal/crm/... ./internal/dashboard/... ./internal/bases/...
go test -coverprofile=cover107_all ./internal/...
go tool cover -func=cover107_all | Select-String total
```
