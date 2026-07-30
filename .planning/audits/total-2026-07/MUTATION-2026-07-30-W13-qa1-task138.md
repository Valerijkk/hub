# MUTATION W13 slot 4 refill — TASK-138 (qa-auto1) · 2026-07-30

> Checker: **qa-auto1** · Env: localhost `POSTGRES_TEST_URL` / `MONGODB_TEST_URL` (`127.0.0.1` after compose port republish) · Go 1.25  
> Criteria: `.planning/backlog/TASK-138-2026-07-29-a3-2-1-tms-handlers.md` `[MUTATION CRITERIA]`  
> Коммит: **не создавался** (инструкция слота).

## Verdict

**1/1 PROVEN** · **0 NOT_PROVEN** · TASK-138 → **Закрыта**

| ID | Verdict | RED test | Mutation (BREAK) | Layer |
|----|---------|----------|------------------|-------|
| A3W2-1-001 | **PROVEN** | `TestEntRepository_UpdateSuite_OptimisticLock_A3W2_1_001` | drop `testsuite.VersionEQ(input.Version)` in `UpdateSuite` | **EntRepository** (канон OL) |
| A3W2-1-001 | **PROVEN** (DTO arm) | `TestUpdateSuite_RequiresVersion_StructValidator` | drop `validate:"required"` on `UpdateSuiteRequest.Version` | DTO / StructValidator |

Production restored (`git checkout --`) после каждого BREAK. Leftover MUTATED: **нет**.  
`layer_mismatch IDs: (none)` · `mock≠Ent IDs: A3W2-1-001` (service-only mock path **не** использован для kill; Ent RED закрывает OL — criteria GAP_NOTED).

## Per-ID evidence

### A3W2-1-001 — UpdateSuite optimistic lock (Ent + DTO)

```text
BASELINE → PASS (~2.9s)
  TestUpdateSuite_RequiresVersion_StructValidator
  TestEntRepository_UpdateSuite_OptimisticLock_A3W2_1_001

BREAK #1 (prefer Ent OL): ent_repository.go UpdateSuite — снять VersionEQ
→ FAIL: TestEntRepository_UpdateSuite_OptimisticLock_A3W2_1_001
        An error is expected but got nil  (stale last-write-wins)
RESTORE → PASS (~3.2s)

BREAK #2 (DTO required): dto.go UpdateSuiteRequest.Version — без validate:"required"
→ FAIL: TestUpdateSuite_RequiresVersion_StructValidator
        expected 400, got 500  (omit/0 проходит Bind; дальше conflict/ошибка)
RESTORE → PASS (~2.8s)

FINAL (оба теста) → PASS
```

Service mock `TestUpdateSuite_OptimisticLock_Service_A3W2_1_001` **не** мутировался (mock≠Ent); kill target = Ent `VersionEQ`.

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@127.0.0.1:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@127.0.0.1:27017/?authSource=admin"

go test -tags hubtest -race -count=1 -timeout=180s ./testframework/tms/ `
  -run "TestUpdateSuite_RequiresVersion_StructValidator|TestEntRepository_UpdateSuite_OptimisticLock_A3W2_1_001"
```

Note: host publish `:5432` потребовал `docker restart deployments-postgres-1` (TCP refused до рестарта). Concurrent leftover `MUTATED` в `bases/service.go` (чужой слот) восстановлен перед baseline, иначе `tms` не собирался.

## Gaps

- Response `version` (`toSuiteResponse`) — не отдельный kill (criteria Gaps).
- Concurrent PUT stress — out of scope.
- DTO RED даёт 500 (не «omit проходит до 200»): Bind без required → version=0 → repo conflict path; assert 400 всё равно краснеет → kill валиден.

## Handoff

- Evidence: этот файл · TASK-138 **Закрыта** · README sync  
- be-ver → code-reviewer (fix code already FIXED; mutation PROVEN)  
- Коммит: не создавался
