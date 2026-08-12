# Phase 114 Verification

**Date:** 2026-06-22  
**Branch:** `feat/v2.5-trust`  
**Verdict:** PARTIAL (80.2% vs 81% stretch goal)

## Goal

COV **80.0% → ≥81%**

## Results

| Metric | Before | After | Target |
|--------|--------|-------|--------|
| COV overall | 80.0% | **80.2%** | ≥81% |
| `search` | 75.9% | **~76%** | push |
| `messenger` | 76.3% | **~76.5%** | push |
| `tms` | 78.1% | **78.5%** | push |
| `tracker` | 77.8% | **78.7%** | push |
| `sprint` | 81.7% | **81.7%** | hold |
| `intranet` | ~81% | **~81%** | hold |

## Tests Added

| File | Coverage focus |
|------|----------------|
| `backend/internal/search/service_wave114_test.go` | empty scope shards, no-mongo messages |
| `backend/internal/messenger/service_wave114_test.go` | RevokeInvite/ListInvites RBAC |
| `backend/internal/messenger/handler_wave114_test.go` | ListPins, invalid channel IDs |
| `backend/internal/tms/service_wave114_test.go` | ExecuteTestResult error paths |
| `backend/internal/tms/handler_wave114_test.go` | execute/assign JSON errors |
| `backend/internal/tracker/service_wave114_test.go` | workflow/label/issue-link not-found |
| `backend/internal/tracker/handler_wave114_test.go` | workflow/label invalid IDs |
| `backend/internal/sprint/service_wave114_test.go` | UpdateSprint not-found |
| `backend/internal/intranet/service_wave114_test.go` | DeleteNews forbidden/not-found |

## Commands Run

```
go test -count=1 ./internal/search/... ./internal/messenger/... ./internal/tms/... ./internal/tracker/... ./internal/sprint/... ./internal/intranet/...  → PASS
go test -coverprofile=cov114_all ./internal/...  → total 80.2%
go test -race ./internal/search/... ./internal/messenger/... ./internal/tms/... ./internal/tracker/... ./internal/sprint/... ./internal/intranet/...  → PASS
```

## Notes

- 81% stretch goal not reached; remaining gains require PG tsvector/Mongo integration paths or net-new ent tests.
- Phase 115 should target `searchUsers`/`searchIssues` SQL paths (needs PG testcontainer) or `messenger/service_bot.go` + `notification/telegram_bind.go`.
