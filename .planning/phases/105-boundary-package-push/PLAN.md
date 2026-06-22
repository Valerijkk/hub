# Phase 105 — Boundary Package Push

**Branch:** `feat/v2.5-trust`  
**Goal:** Поднять пакеты на грани 75% (**marketplace**, **project**, **tms**) и удержать overall COV ≥78.4%.

## Targets

| Package | Before | Target |
|---------|--------|--------|
| `marketplace` | 75.4% | ≥78% |
| `project` | 75.4% | ≥77% |
| `tms` | 75.3% | ≥76% |
| COV overall | 78.4% | ≥78.4% |

## Tasks

### Wave A — marketplace
- [x] `mapServiceError` — все 7 веток (unit + handler)
- [x] `WithdrawListing` / `ownerStatusChange` / `transitionReview` edge cases
- [x] Handler: withdraw 404/403/409/400, approve version conflict, invalid create

### Wave B — project
- [x] Service: Unarchive/Restore/Delete not-found + internal paths
- [x] `memberCountOrZero` fallback при ошибке MemberCount
- [x] `CreateTeam` constraint + internal error
- [x] Templates: GetTemplate not-found, DeleteTemplate custom success

### Wave C — tms
- [x] Environment handlers: nil UUID, missing name, invalid env ID, delete 404
- [x] Service: Update/Delete environment not-found
- [x] `CompareEnvironments` with runs/results + handler happy path

## Verification

```powershell
cd backend
go test -count=1 ./internal/marketplace/... ./internal/project/... ./internal/tms/...
go test -coverprofile=cover105_all ./internal/...
go tool cover -func=cover105_all | Select-String total
```

## Success criteria

- [x] marketplace ≥78%
- [x] project ≥77%
- [x] tms ≥76%
- [x] overall COV ≥78.4%
- [x] all new tests PASS
