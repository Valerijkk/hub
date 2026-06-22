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

## Frontend (TF-M06) ✅ · Pipeline domains (TF-M10) ✅

Vitest: `testframework/stages/unit/frontend-test/domains/{domain}/`

Update `frontend/vite.config.ts`:

- `testframeworkDir`: `../testframework/stages/unit/frontend-test`
- vitest `include`: `../testframework/stages/unit/frontend-test/domains/**/*.test.{ts,tsx}`

E2E: `testframework/stages/e2e/domains/{domain}/` + `shared/` (POM, steps)

Run `node frontend/scripts/link-testframework-node-modules.mjs` before Vitest/E2E (symlink `node_modules`).

Pipeline manifests: `testframework/pipeline/jobs/*.yaml` · Map: `testframework/PIPELINE.md`

## Go (TF-M05) ✅

Stay in `backend/testframework/` — same Go module as `internal/` (required for `//go:build hubtest` export hooks).
