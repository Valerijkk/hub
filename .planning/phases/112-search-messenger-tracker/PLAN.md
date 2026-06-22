# Phase 112 — Search + Messenger + Tracker Push

**Branch:** `feat/v2.5-trust`  
**Goal:** Triple push `search (75.8%) + messenger (75.8%) + tracker (75.9%)` → COV **≥80%**.

## Targets

| Package | Before | Target |
|---------|--------|--------|
| `tracker` | 75.9% | ≥77.5% |
| `messenger` | 75.8% | ≥76.5% |
| `search` | 75.8% | ≥76% |
| COV overall | 79.5% | ≥80% |

## Tasks

1. **Tracker** — UpdateIssue validation/transitions, DeleteComment/CustomField service, handler JSON errors (`service_wave112_test.go`, `handler_wave112_test.go`)
2. **Messenger** — MarkChannelRead, unread, bookmarks remove, delete message, promote invalid JSON (`handler_wave112_test.go`)
3. **Search** — list filters, share invalid ID, search-all-types service path (`handler_wave112_test.go`)

## Verification

```powershell
cd backend
go test -count=1 ./internal/tracker/... ./internal/messenger/... ./internal/search/...
go test -coverprofile=cov112_all ./internal/...
go tool cover -func=cov112_all | Select-String total
go test -race ./internal/tracker/... ./internal/messenger/... ./internal/search/...
```
