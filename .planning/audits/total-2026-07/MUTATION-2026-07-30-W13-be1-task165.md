# MUTATION W13 slot 1 refill — TASK-165 (be1) · 2026-07-30

> Checker: **be1** · Env: localhost Postgres/Mongo · Go 1.25 · `-tags hubtest`  
> Criteria: `.planning/backlog/archive/2026-07-29/TASK-165-2026-07-29-a3-7-4-storage-dash-backup.md` `[MUTATION CRITERIA]`  
> Коммит не создавался (инструкция слота).

## Verdict

**5/5 PROVEN** · TASK-165 → **Закрыта** · leftover MUTATED: **нет**

| ID | Verdict | RED test |
|----|---------|----------|
| A3W7-4-001 | **PROVEN** | `TestGetProjectStats_AggregateFailClosed_A3W74_001`; `TestGetAdminStats_CountsFailClosed_A3W74_001` (+ `BeforeTop`) |
| A3W7-4-002 | **PROVEN** | `TestDeleteBackup_S3FailClosedWarn_A3W74_002`; `TestCleanupOldBackups_S3FailWarn_A3W74_002` |
| A3W7-4-003 | **PROVEN** | `TestGetLayout_StrangerProjectForbidden_A3W74_003` |
| A3W7-4-004 | **PROVEN** | `TestHandler_DeleteFile_SoftDeleteNotFound404_A3W74_004` |
| A3W7-4-005 | **PROVEN** | `TestAggregateProjectCompletedByDay_Bounded_A3W74_005` |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: A3W7-4-004` (GAP_NOTED — handler map via stub SoftDelete NotFound; kill на prod `handler.go`)

## Per-ID evidence

### A3W7-4-001 — Aggregate/Admin Count* fail-closed

```text
BASELINE AggregateFailClosed + CountsFailClosed + BeforeTop -> PASS (~1.2s)

BREAK: dashboard/service.go —
  Aggregate err → slog + continue (empty widgets, return stats,nil path)
  CountUsers err → slog soft-skip (не return nil, err)
-> FAIL AggregateFailClosed: Error expected but got nil
-> FAIL CountsFailClosed: Contains "count users" (soft → next Count* message)
RESTORE -> PASS (~1.2s)
```

### A3W7-4-002 — S3 DeleteObject Warn + fail-closed

```text
BASELINE DeleteBackup_S3FailClosedWarn + CleanupOldBackups_S3FailWarn -> PASS (~8.9s)

BREAK: backup/service.go —
  DeleteBackup: _, _ = DeleteObject (no Warn, no return err → DB delete)
  cleanupOldBackups: _, _ = DeleteObject (no Warn, no continue → DB delete)
-> FAIL DeleteBackup: Error expected but got nil
-> FAIL Cleanup: Warn missing + "backup_record not found" (DB row deleted)
RESTORE -> PASS (~8.9s)
```

### A3W7-4-003 — GetLayout membership

```text
BASELINE TestGetLayout_StrangerProjectForbidden_A3W74_003 -> PASS (~1.0s)

BREAK: dashboard/service.go GetLayout — убрать IsProjectMember / ErrForbidden
-> FAIL Error expected but got nil (stranger OK)
RESTORE -> PASS (~1.0s)
```

### A3W7-4-004 — SoftDelete NotFound → 404

```text
BASELINE TestHandler_DeleteFile_SoftDeleteNotFound404_A3W74_004 -> PASS (~0.8s)

BREAK: storage/handler.go Delete — убрать ent.IsNotFound → 404 map
-> FAIL expected 404 actual 500
RESTORE -> PASS (~0.8s)
```

Note (mock≠Ent / GAP_NOTED): тест бьёт Fiber Delete + stub SoftDelete→`&ent.NotFoundError{}`,
не race SoftDelete на живом Ent. Guard kill на реальном prod handler → RED именованного теста.

### A3W7-4-005 — burndown AggregateProjectCompletedByDay bounded

```text
BASELINE TestAggregateProjectCompletedByDay_Bounded_A3W74_005 -> PASS (~1.1s)

BREAK: dashboard/repository.go AggregateProjectCompletedByDay —
  Count/day → Issue.Query().All() + per-issue map keys
-> FAIL LessOrEqual len(counts)<=4 got 5
RESTORE -> PASS (~1.1s)
```

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -count=1 -timeout=180s ./testframework/dashboard/ -run `
  "TestGetProjectStats_AggregateFailClosed_A3W74_001|TestGetAdminStats_CountsFailClosed_A3W74_001|TestGetAdminStats_CountsFailClosed_BeforeTop_A3W74_001|TestGetLayout_StrangerProjectForbidden_A3W74_003|TestAggregateProjectCompletedByDay_Bounded_A3W74_005"
# FINAL GREEN PASS ~1.1s

go test -tags hubtest -count=1 -timeout=120s ./testframework/backup/ -run `
  "TestDeleteBackup_S3FailClosedWarn_A3W74_002|TestCleanupOldBackups_S3FailWarn_A3W74_002"
# FINAL GREEN PASS ~7.2s

go test -tags hubtest -count=1 -timeout=120s ./testframework/storage/ -run `
  "TestHandler_DeleteFile_SoftDeleteNotFound404_A3W74_004"
# FINAL GREEN PASS ~1.1s
```

## Gaps

- A3W7-4-004: **mock≠Ent** documented — PROVEN на handler map; нет отдельного Ent SoftDelete race IT.
- Soft optional charts (burndown/testRuns/activity) остаются soft-fail — kill 001 только critical Aggregate/Admin Count*.
- Docs-only: **N/A**

## Handoff

- Evidence: этот файл · TASK-165 **Закрыта** · README sync  
- be-ver → code-reviewer  
- Production restored after each BREAK; `git diff` на dashboard/backup/storage prod files — empty; no `MUTATED` leftover  
- Prod code unchanged this slot (fix already on HEAD; mutation prove only)
