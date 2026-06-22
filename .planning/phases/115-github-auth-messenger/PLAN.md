# Phase 115 — GitHub/GitLab + Auth + Messenger + Bases + AI Push

**Branch:** `feat/v2.5-trust`  
**Goal:** Push COV **≥81%** (from 80.2%).

## Targets

| Package | Before | After |
|---------|--------|-------|
| `bases` | 78.6% | **~80%** |
| `ai` | 78.9% | **~79%** |
| `messenger` | 76.9% | **~77%** |
| `search` | 75.9% | **~76%** |
| `tms` | 78.7% | **~79%** |
| `notification` | 78.4% | **~78.5%** |
| `auth` | 82.0% | **82%** |
| `github` | 81.6% | **~82%** |
| `gitlab` | 80.3% | **~80%** |
| `tracker` | 78.9% | **~79%** |
| COV overall | 80.2% | **80.7%** (stretch 81% missed by 0.3pp) |

## Tasks

1. **GitHub/GitLab** — CreateBranch not-found service + TriggerPipeline/CreateBranch invalid JSON handlers
2. **Auth** — ApprovePendingUser invalid ID/not-found + ListPendingUsers member forbidden
3. **Messenger** — AcceptInvite edge cases + handler pin/delete + JoinChannel/GetChannel/AddMembers service paths
4. **TMS** — handler invalid run IDs + service Reopen/Complete/CreateTestResult errors
5. **Tracker** — duplicate workflow/label group service + handler invalid project ID
6. **Bases** — handler/service table/column/record error paths (new module push)
7. **AI** — handler quota/consent/content-policy + AssistJQL no-provider
8. **Search** — typed shard Search (user/file/test_case) admin path
9. **Notification** — handler delete invalid ID + list invalid page

## New test files

- `backend/internal/github/service_wave115_test.go`
- `backend/internal/gitlab/handler_wave115_test.go`
- `backend/internal/auth/handler_wave115_test.go`
- `backend/internal/messenger/service_wave115_test.go`, `handler_wave115_test.go`
- `backend/internal/tracker/service_wave115_test.go`, `handler_wave115_test.go`
- `backend/internal/tms/handler_wave115_test.go`, `service_wave115_test.go`
- `backend/internal/bases/handler_wave115_test.go`, `service_wave115_test.go`
- `backend/internal/ai/handler_wave115_test.go`
- `backend/internal/search/service_wave115_test.go`
- `backend/internal/notification/handler_wave115_test.go`

## Verification

```powershell
cd backend
go test -count=1 ./internal/...
go test -coverprofile=cov115_all ./internal/...
go tool cover -func=cov115_all | Select-String total
go test -race ./internal/auth/... ./internal/messenger/... ./internal/ai/... ./internal/bases/...
```

## Next (Phase 116)

- Search `searchUsers`/`searchIssues` tsvector paths (PG integration or SQL mock)
- Messenger `CreateChannel` NATS fan-out + poll vote repo paths
- Push COV **≥81%** (+0.3pp)
