# MUTATION W13 slot 1 refill — TASK-142 (be1) · 2026-07-30

> Checker: **be1** · Env: localhost Postgres/Mongo · Go 1.25 · `-tags hubtest`  
> Criteria: `.planning/backlog/archive/2026-07-29/TASK-142-2026-07-29-a3-2-5-bases.md` `[MUTATION CRITERIA]`  
> Коммит не создавался (инструкция слота).

## Verdict

**3/3 PROVEN** · TASK-142 → **Закрыта** · leftover MUTATED: **нет**

| ID | Verdict | RED test |
|----|---------|----------|
| A3W2-A5-001 | **PROVEN** | `TestRepository_SoftDeleteAndRestoreHubBase_Task142`; `TestRepository_PurgeExpiredHubBases_Task142`; `TestAuditService_RestoreHubBase_ViaTrashAPI_Task142` |
| A3W2-A5-002 | **PROVEN** | `TestTask142_SoftDeleteTable_CascadesChildren`; `TestTask142_SoftDeleteTableThenBase_NoOrphans` |
| A3W2-A5-003 | **PROVEN** | `TestTask142_NameMaxLen_InvalidInput400` |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)`

## Per-ID evidence

### A3W2-A5-001 — hubbase ListTrash / Restore / Purge

```text
BASELINE TestRepository_SoftDeleteAndRestoreHubBase_Task142
         TestRepository_PurgeExpiredHubBases_Task142
         TestAuditService_RestoreHubBase_ViaTrashAPI_Task142 -> PASS (~1.6s)

BREAK: audit/service.go — убрать case "hubbase" из ListTrash/RestoreItem/SoftDelete
       + убрать purge("hubbases");
       audit/repository.go — stub RestoreHubBase (no-op) + stub PurgeExpiredHubBases (return 0,nil)
-> FAIL SoftDeleteAndRestore: list total after restore expected 0 got 1; hub_base not found (deleted_at uncleared)
-> FAIL PurgeExpiredHubBases: expected n=1 got 0; HubBase.Get still succeeds
-> FAIL RestoreHubBase_ViaTrashAPI: ListTrash total expected 1 got 0
RESTORE (git checkout service.go repository.go) -> PASS (~1.0s)
```

### A3W2-A5-002 — SoftDeleteTable cascade + SoftDeleteBase all tableIDs

```text
BASELINE TestTask142_SoftDeleteTable_CascadesChildren
         TestTask142_SoftDeleteTableThenBase_NoOrphans -> PASS

BREAK: bases/repository.go SoftDeleteTable — только table deleted_at (без records/columns);
       SoftDeleteBase — вернуть hubtable.DeletedAtIsNil() filter
-> FAIL CascadesChildren: aliveCols=2 / aliveRecs=1 (expect 0)
-> FAIL SoftDeleteTableThenBase_NoOrphans: aliveCols=2 / aliveRecs=1 (orphans)
RESTORE (git checkout repository.go) -> PASS (~0.6s)
```

### A3W2-A5-003 — validateNameLen → ErrInvalidInput / 400

```text
BASELINE TestTask142_NameMaxLen_InvalidInput400 -> PASS

BREAK: bases/service.go validateNameLen — always return nil (skip RuneCount guard)
-> FAIL ErrorIs ErrInvalidInput: got Ent MaxLen wrap
       ("create base: ent: validator failed for field \"HubBase.name\": value is greater than the required length")
RESTORE (git checkout service.go) -> PASS (included in FINAL GREEN)
```

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -count=1 -timeout=180s ./testframework/audit/ ./testframework/bases/ -run `
  "TestRepository_SoftDeleteAndRestoreHubBase_Task142|TestRepository_PurgeExpiredHubBases_Task142|TestAuditService_RestoreHubBase_ViaTrashAPI_Task142|TestTask142_SoftDeleteTable_CascadesChildren|TestTask142_SoftDeleteTableThenBase_NoOrphans|TestTask142_NameMaxLen_InvalidInput400"
# FINAL GREEN PASS ~1.2s audit / ~0.8s bases
```

## Gaps

- SoftDelete/Restore call-map ×9 в `task114_residual_test.go` — доп. switch coverage; канон kill — Task142 Ent trash tests (PROVEN).
- FE admin trash UI — out of scope · N/A

## Handoff

- Evidence: этот файл · TASK-142 **Закрыта** · README sync  
- be-ver → code-reviewer  
- Production restored after each BREAK; `git status` на `backend/internal/{audit,bases}/{service,repository}.go` — clean; no `MUTATED` leftover  
- Prod code unchanged this slot (fix already on HEAD; mutation prove only)
