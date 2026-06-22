# Phase 114 — Search + Messenger + TMS + Tracker Push

**Branch:** `feat/v2.5-trust`  
**Goal:** Push COV **≥81%** (from 80.0%).

## Targets

| Package | Before | After |
|---------|--------|-------|
| `search` | 75.9% | **76.0%** |
| `messenger` | 76.3% | **76.5%** |
| `tms` | 78.1% | **78.5%** |
| `tracker` | 77.8% | **78.7%** |
| `sprint` | 81.7% | **81.7%** |
| `intranet` | ~81% | **~81%** |
| COV overall | 80.0% | **80.2%** |

## Tasks

1. **Search** — empty-scope shard paths + no-mongo messages (`service_wave114_test.go`)
2. **Messenger** — RevokeInvite/ListInvites forbidden + handler pins/polls (`service_wave114_test.go`, `handler_wave114_test.go`)
3. **TMS** — ExecuteTestResult error paths + handler JSON (`service_wave114_test.go`, `handler_wave114_test.go`)
4. **Tracker** — workflow/label/issue-link not-found + handler invalid IDs (`service_wave114_test.go`, `handler_wave114_test.go`)
5. **Sprint** — UpdateSprint not-found (`service_wave114_test.go`)
6. **Intranet** — DeleteNews forbidden/not-found service (`service_wave114_test.go`)

## Verification

```powershell
cd backend
go test -count=1 ./internal/search/... ./internal/messenger/... ./internal/tms/... ./internal/tracker/... ./internal/sprint/... ./internal/intranet/...
go test -coverprofile=cov114_all ./internal/...
go tool cover -func=cov114_all | Select-String total
go test -race ./internal/search/... ./internal/messenger/... ./internal/tms/... ./internal/tracker/... ./internal/sprint/... ./internal/intranet/...
```
