# Phase 106 — Triple Boundary Push (search / messenger / tracker)

**Branch:** `feat/v2.5-trust`  
**Goal:** Поднять три пакета на грани 75% и overall COV ≥78.7%.

## Targets

| Package | Before | Target |
|---------|--------|--------|
| `search` | 75.1% | ≥75.5% |
| `messenger` | 75.1% | ≥75.5% |
| `tracker` | 75.1% | ≥75.5% |
| COV overall | 78.7% | ≥78.7% |

## Tasks

### Wave A — search
- [x] `Search()` all entity types parallel goroutines
- [x] Admin `searchUsers` → unscoped routing
- [x] `searchChannels` ILIKE fallback path (sqlite)
- [x] Closed DB errors for channel/project ID lookups

### Wave B — messenger
- [x] Handler error paths: GetChannel, CreateChannel, Join private, DM, Mute, AddMembers
- [x] ValidateInvite missing token → `valid: false`

### Wave C — tracker
- [x] ListComments after create
- [x] Issue link CRUD + time logs + workflow status update
- [x] GetIssue not-found handler path

## Verification

```powershell
cd backend
go test -count=1 ./internal/search/... ./internal/messenger/... ./internal/tracker/...
go test -coverprofile=cover106_all ./internal/...
go tool cover -func=cover106_all | Select-String total
```

## Success criteria

- [x] messenger ≥75.4%
- [x] tracker ≥75.2%
- [x] overall COV ≥78.7%
- [x] all new tests PASS

## Notes

- `search` SQL happy paths требуют PostgreSQL tsvector — локально покрыты error/empty paths; mongo message search в CI (`MONGODB_TEST_URL`).
