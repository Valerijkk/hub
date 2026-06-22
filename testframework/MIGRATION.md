# Test Migration Script — internal/*_test.go → backend/testframework/

> Run from repo root. **Docker verify after each module batch.**

## Usage

```powershell
.\scripts\migrate-go-tests.ps1 -Module notification -DryRun
.\scripts\migrate-go-tests.ps1 -Module notification
```

## Prerequisites per module

1. Create `internal/<module>/export_hooks.go` with `//go:build hubtest` and Export* hooks
2. Create `backend/testframework/<module>/fixtures_test.go` with shared env (DRY)
3. Run migration script
4. `docker run ... go test -tags hubtest ./testframework/<module>/... ./internal/<module>/...`

## Transform rules

| From (internal) | To (testframework) |
|-----------------|-------------------|
| `package notification` | `package notification_test` |
| `c.handleX(` | `notification.ExportHandleX(c, ` |
| `repo.db` | `notification.ExportRepositoryDB(repo)` |
| `buildPushURL` | `notification.ExportBuildPushURL` |
| `newNotificationTestEnv` | use `fixtures_test.go` helpers |

## Module order (TF-M02–M05)

1. notification (66) — IN PROGRESS
2. auth, ws, rbac, audit
3. tracker, tms, project, messenger
4. remaining modules

## Frontend (TF-M06)

Move `frontend/testframework/` → `testframework/frontend/` and update `vite.config.ts` root.
