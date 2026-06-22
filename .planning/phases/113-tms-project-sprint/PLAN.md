# Phase 113 — TMS + Project + Sprint + Messenger + Tracker Push

**Branch:** `feat/v2.5-trust`  
**Goal:** Cross-module push to COV **≥80%** (from 79.7%).

## Targets

| Package | Before | After |
|---------|--------|-------|
| `tms` | 76.9% | **78.1%** |
| `project` | 78.3% | **80.1%** |
| `sprint` | 79.8% | **81.7%** |
| `messenger` | 76.2% | **76.3%** |
| `tracker` | 77.6% | **77.8%** |
| COV overall | 79.7% | **80.0%** |

## Tasks

1. **TMS** — CreateTestCase/UpdateTestCase validation, DeleteSuite/Component not-found, report handler GET paths (`service_wave113_test.go`, `handler_wave113_test.go`)
2. **Project** — team/member error paths, incoming cross-link delete, handler JSON errors (`service_wave113_test.go`, `handler_wave113_test.go`)
3. **Sprint** — handler invalid JSON + StartSprint not-found (`handler_wave113_test.go`)
4. **Messenger** — LeaveChannel/DeleteChannel invalid ID, ForwardMessage handler (`handler_wave113_test.go`)
5. **Tracker** — DeleteComment/CustomField not-found service paths (`service_wave113_test.go`)

## Verification

```powershell
cd backend
go test -count=1 ./internal/tms/... ./internal/project/... ./internal/sprint/... ./internal/messenger/... ./internal/tracker/...
go test -coverprofile=cov113_all ./internal/...
go tool cover -func=cov113_all | Select-String total
go test -race ./internal/tms/... ./internal/project/... ./internal/sprint/... ./internal/messenger/... ./internal/tracker/...
```
