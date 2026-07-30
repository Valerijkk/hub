# MUTATION W13 slot 1 refill — TASK-170 (be1) · 2026-07-30

> Checker: **be1** · Env: localhost Postgres/Mongo · Go 1.25 · `-tags hubtest`  
> Criteria: `.planning/backlog/archive/2026-07-29/TASK-170-2026-07-29-a3-8-4-schemas.md` `[MUTATION CRITERIA]`  
> Коммит не создавался (инструкция слота).

## Verdict

**2/2 PROVEN** (+001b) · TASK-170 → **Закрыта** · leftover MUTATED: **нет**  
Rework: удалён `task170ApplyActiveUnique` (DDL-in-test); unique index применяется из
`migrations/20260729_task170_sprint_active_unique_setnull.sql` (DROP Schema.Create
копии + Exec CREATE из файла). Cascade — Schema.Create из codegen (`migrate/schema.go`).

| ID | Verdict | RED test |
|----|---------|----------|
| A3W8-4-001 | **PROVEN** | `TestConstraint_TASK170_DuplicateActiveSprintRejected` (+ `TestSchema_TASK170_MigrateSQLExists`) |
| A3W8-4-001b | **PROVEN** | `TestConstraint_TASK170_SoftDeletedActiveAllowsNewActive` (+ `TestSchema_TASK170_MigrateSQLExists`) |
| A3W8-4-002 | **PROVEN** | `TestCascade_TASK170_SprintHardDeleteSetsIssueSprintNull` (+ MigrateSQLExists + SprintActiveUniqueAndSetNull) |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)` · `docs-only IDs: (none)`

## Per-ID evidence

### A3W8-4-001 — partial unique active sprint (migrate SQL)

```text
BASELINE TASK170 → PASS (~0.56s)

BREAK: migrations/20260729_task170_sprint_active_unique_setnull.sql
  — удалить блок CREATE UNIQUE INDEX … sprint_project_sprints …
→ FAIL TestSchema_TASK170_MigrateSQLExists
       executable CREATE UNIQUE INDEX extractable (empty)
→ FAIL TestConstraint_TASK170_DuplicateActiveSprintRejected
       migrate SQL must contain CREATE UNIQUE INDEX sprint_project_sprints
RESTORE → PASS (~0.64s)
```

Guard: apply-from-file `task170ApplyActiveUniqueFromMigrate` — DROP + Exec канона.

### A3W8-4-001b — WHERE must include `deleted_at IS NULL`

```text
BREAK: тот же SQL — WHERE status = 'active' (без AND deleted_at IS NULL)
→ FAIL TestSchema_TASK170_MigrateSQLExists
       does not contain "status = 'active' AND deleted_at IS NULL"
→ FAIL TestConstraint_TASK170_SoftDeletedActiveAllowsNewActive
       UNIQUE constraint failed: sprints.project_sprints
       (soft-deleted active blocks new active)
RESTORE → PASS (~0.68s)
```

### A3W8-4-002 — Sprint→issues OnDelete SetNull

```text
BREAK (valid syntax, no generate):
  migrate SQL: ON DELETE SET NULL → ON DELETE RESTRICT
  ent/migrate/schema.go: issues_sprints_issues OnDelete SetNull → NoAction
  ent/schema/sprint.go: entsql.OnDelete(entsql.SetNull) → NoAction
→ FAIL TestSchema_TASK170_SprintActiveUniqueAndSetNull
       does not contain "entsql.OnDelete(entsql.SetNull)"
→ FAIL TestSchema_TASK170_MigrateSQLExists
       does not contain "ON DELETE SET NULL"
→ FAIL TestCascade_TASK170_SprintHardDeleteSetsIssueSprintNull
       FOREIGN KEY constraint failed
RESTORE → PASS (~0.58s)
```

Runtime cascade = Schema.Create из codegen (допустимо критериями «migrate file / Schema.Create»).
Migrate SQL доказан companion source-assert.

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -count=1 -timeout=180s -vet=off ./testframework/ent/ -run TASK170 -v
# FINAL GREEN PASS ~0.58–0.76s

go test -tags hubtest -count=1 -timeout=300s -vet=off ./testframework/sprint/ ./testframework/ent/
# regression PASS ~1.7s / ~0.9s
```

## Gaps

- Shared/prod apply migrate SQL — по-прежнему `needs-owner-apply` + pre-migrate cleanup duplicate actives.
- IndexWhere→`go generate` как путь мутации — **запрещён** (исторический MUTATION_FAILED_TO_BUILD); не использовался.
- SQLite ALTER FK из migrate SQL не применяется (нет ADD CONSTRAINT); cascade IT = Schema.Create.

## Handoff

- Evidence: этот файл · TASK-170 **Закрыта** · README sync  
- be-ver → code-reviewer  
- Production restored after BREAK; leftover MUTATED: **нет**  
- Коммит: не создавался
