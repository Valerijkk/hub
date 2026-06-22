# Phase 121 — Notification Consumer + Tracker Push to 82%

**Branch:** `feat/v2.5-trust`  
**Goal:** notification ≥82%, tracker ≥82%, COV ≥82%  
**Verify:** Docker race + coverprofile

## Success Criteria

| Criterion | Target |
|-----------|--------|
| `notification` package | ≥82% |
| `tracker` package | ≥82% |
| COV overall | ≥82% |
| Docker race | PASS |
| TF-M01/M02 pilot | export_test.go + testframework/notification tests |

## Tasks

### 121.1 Notification consumer gaps
- `handleStatusChanged` — success path (assignee ≠ actor)
- `handleMessageReceived` — DM recipient + valid mention
- `handleCallStarted` — notifies channel member
- `processEvent` — TypeCallStarted + TypeMessageReceived routing

### 121.2 Test framework pilot
- `internal/notification/export_test.go` — export hooks
- `backend/testframework/notification/fixtures_test.go` — DRY env
- `backend/testframework/notification/consumer_wave121_test.go` — new tests

### 121.3 Tracker push
- Handler routes: GetBoardData, labels, workflow (if below 82%)

## Commands

```bash
go test -race -count=1 ./internal/notification/... ./testframework/notification/...
go test -coverprofile=cov ./internal/notification/...
```
