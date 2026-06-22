# Phase 109 — Triple Boundary Push (Search + Messenger + Tracker)

**Branch:** `feat/v2.5-trust`  
**Goal:** Push lowest packages toward 76%+ and maintain COV ≥79%.

## Targets

| Package | Before | Target |
|---------|--------|--------|
| `search` | 75.1% | ≥75.5% |
| `messenger` | 75.4% | ≥75.5% |
| `tracker` | 75.2% | ≥75.5% |
| COV overall | 79.1% | ≥79.2% |

## Tasks

1. **Search** — empty-scope shard paths + filter not-found + handler validation (`service_wave109_test.go`, `handler_wave109_test.go`)
2. **Messenger** — typing/heartbeat, list messages invalid before, send/edit/reaction JSON errors (`handler_wave109_test.go`)
3. **Tracker** — GetBoardData swimlanes (assignee/priority), board handler swimlane query, time/comment error paths (`service_wave109_test.go`, `handler_wave109_test.go`)

## Verification

```powershell
cd backend
go test -count=1 ./internal/search/... ./internal/messenger/... ./internal/tracker/...
go test -coverprofile=cov109_all ./internal/...
go tool cover -func=cov109_all | Select-String total
```
