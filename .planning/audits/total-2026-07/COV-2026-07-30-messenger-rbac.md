# Coverage lift — messenger + rbac (2026-07-30)

## Verdict

| Package | Before (remeasure) | After | Gate ≥82% | Delta |
|---------|-------------------:|------:|:---------:|------:|
| `internal/messenger` | **69.6%** | **82.8%** | ✅ PASS | **+13.2 pp** |
| `internal/rbac` | **77.5%** | **86.8%** | ✅ PASS | **+9.3 pp** |

Campaign: Stage C W-last / TASK-008 package floor.  
Measured: **2026-07-30** (local host, `-tags hubtest`).

Prior cite (2026-07-29 full-suite profile): messenger **69.7%**, rbac **77.9%** — aligned with today’s package-scoped remeasure.

## Method

```powershell
$env:POSTGRES_TEST_URL = 'postgresql://hub:hub@localhost:5432/hub?sslmode=disable'
$env:MONGODB_TEST_URL  = 'mongodb://hub:hub@localhost:27017/?authSource=admin'
$env:CGO_ENABLED = '1'

# PowerShell: quote -coverpkg / -coverprofile
go test -tags=hubtest -count=1 -timeout=600s -covermode=atomic `
  "-coverpkg=github.com/valerii/hub/backend/internal/messenger" `
  ./testframework/messenger/... `
  "-coverprofile=$env:TEMP\cov_messenger.out"

go test -tags=hubtest -count=1 -timeout=600s -covermode=atomic `
  "-coverpkg=github.com/valerii/hub/backend/internal/rbac" `
  ./testframework/rbac/... `
  "-coverprofile=$env:TEMP\cov_rbac.out"

go tool cover "-func=..."
```

Scope: package-targeted `coverpkg` + matching `testframework/<pkg>/...` (not full `./testframework/...` overall).

## Gap analysis (before)

### messenger

Largest hole: `mock_repository.go` (~331 stmts, **~1.5%** covered).  
Testframework keeps a **duplicate** mock in `mock_repository_test.go`, so the in-package mock (used only via `HookTestRepo`) was almost never exercised — dragging package % below historical 82.3%.

Secondary gaps (unchanged this wave): `ent_repository`, handlers, promote/bot edges.

### rbac

~4–5 pp under floor. Zero-cover: `InvalidateUserCache`, `CanCreateTaskInProject`, `DeleteRole`, `IsNotFound`.  
Under-covered: handler 4xx branches, `RequireSystemRole` edges, custom-role mismatch/system/legacy bitmask paths.

## Files added

| File | Role |
|------|------|
| `backend/testframework/messenger/hook_mock_coverage_task008_test.go` | Full `Repository` surface on `HookTestRepo` → covers in-package `mock_repository` |
| `backend/testframework/rbac/coverage_task008_test.go` | Service/repo/handler/middleware branches for rbac gaps |

No production code changes. **No commits** (campaign rule).

## After notes

- messenger `mock_repository.go`: **~94.9%** (was ~1.5%).
- Remaining messenger room: real Ent/handler/service paths (not required for ≥82% gate).
- rbac above historical 84.7% security-wave row on this package-scoped measure.

## Docs sync

- This report: `.planning/audits/total-2026-07/COV-2026-07-30-messenger-rbac.md`
- Brief row update: `.planning/COVERAGE.md` (package lift note)
