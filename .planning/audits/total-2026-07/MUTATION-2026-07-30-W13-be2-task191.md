# MUTATION W13 slot 3 refill — TASK-191 (be2) · 2026-07-30

> Checker: **be2** · Env: localhost Postgres/Mongo · Go 1.25 · `-tags hubtest`  
> Criteria: `.planning/backlog/TASK-191-2026-07-29-audit-ent-automationrule-softdelete.md` `[MUTATION CRITERIA]`  
> Findings: **EA-001-001**, **EA-001-002**, **EA-001-003** (3 IDs)  
> Коммит не создавался. Ent runtime + schema source kills; leftover MUTATED: **нет**.  
> Migrate SQL `20260729_task191_automationrule_soft_delete.sql` — **N/A runtime** (needs-owner-apply).

## Verdict

**3/3 PROVEN** · migrate SQL **N/A** · TASK-191 → **Закрыта** · leftover MUTATED: **нет**

| ID | Verdict | RED test |
|----|---------|----------|
| EA-001-001 | **PROVEN** | `TestEntRepository_AutomationRuleSoftDelete_TASK191` |
| EA-001-002 | **PROVEN** | `TestEntRepository_AutomationRuleSoftDelete_TASK191` (+ Count: `TestEntRepository_AutomationRuleCountIgnoresTrash_TASK191`) |
| EA-001-003 | **PROVEN** | `TestSchema_TASK191_AutomationRuleDeletedAt` (dup `TestSchema_TASK191_SoftDeleteField`) |
| migrate SQL | **N/A** | `TestMigrate_TASK191_*` = file marker only (owner-apply gate) |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)` · `docs-only IDs: (none)`

## Per-ID evidence

### EA-001-001 — DeleteAutomationRule soft `SetDeletedAt` (не hard `.Delete()`)

```text
BASELINE
  go test -tags hubtest -count=1 ./testframework/tracker/ \
    -run TestEntRepository_AutomationRuleSoftDelete_TASK191$
  -> PASS (~1.0s)

BREAK (EntRepository correct layer):
  backend/internal/tracker/repo_automation.go DeleteAutomationRule
  — заменить Update+SetDeletedAt на AutomationRule.Delete().Where(...).Exec
  — MUTATED EA-001-001

  -> FAIL TestEntRepository_AutomationRuleSoftDelete_TASK191 (~1.1s)
       ent: automation_rule not found
       (EA-001: hard Delete would remove the row)

RESTORE soft Update + SetDeletedAt (+ time import)
  -> PASS (~1.2s)

git diff repo_automation.go → empty
MUTATED marker → отсутствует
```

Layer OK: Ent `AutomationRule` + `tracker.EntRepository` (sqlite/postgres hubtest env).

### EA-001-002 — `DeletedAtIsNil()` на List/Find/engine

```text
BASELINE
  go test -tags hubtest -count=1 ./testframework/tracker/ \
    -run "TestEntRepository_AutomationRuleSoftDelete_TASK191$|TestEntRepository_AutomationRuleCountIgnoresTrash_TASK191"
  -> PASS (после restore 001)

BREAK (EntRepository correct layer):
  backend/internal/tracker/repo_automation.go ListAutomationRules
  — убрать automationrule.DeletedAtIsNil() из Where
  — MUTATED EA-001-002

  -> FAIL TestEntRepository_AutomationRuleSoftDelete_TASK191 (~0.8s)
       Should not be: <rule UUID>
       (EA-001: soft-deleted rule must not appear in ListAutomationRules)

RESTORE DeletedAtIsNil on ListAutomationRules
  -> PASS (final suite)

git diff repo_automation.go → empty
MUTATED marker → отсутствует
```

Layer OK: Ent List after soft Delete; Count trash filter remains on Count path (separate assert, not broken by List-only kill).

### EA-001-003 — schema `field.Time("deleted_at")` + indexes

```text
BASELINE
  go test -tags hubtest -count=1 ./testframework/tracker/ \
    -run TestSchema_TASK191_AutomationRuleDeletedAt$
  -> PASS (в общем TASK191 ~1.0s)

BREAK (schema source correct layer):
  backend/ent/schema/automationrule.go
  — убрать field.Time("deleted_at") и indexes deleted_at / project_id+deleted_at
  — MUTATED EA-001-003
  (generated Ent DDL не перегенерировали — static Contains guard)

  -> FAIL TestSchema_TASK191_AutomationRuleDeletedAt (~1.1s)
       does not contain field.Time("deleted_at")
       does not contain index.Fields("deleted_at")
       does not contain index.Fields("project_id", "deleted_at")

RESTORE deleted_at field + indexes
  -> PASS Task191 tracker+ent (~1.2s / ~0.8s)

git diff automationrule.go → empty
MUTATED marker → отсутствует
```

Layer OK: static schema source Contains (не Ent runtime DDL). Flag per criteria: **OK — static schema**.

### Migrate SQL — N/A runtime

```text
backend/migrations/20260729_task191_automationrule_soft_delete.sql
  needs-owner-apply / BLOCKED waiting owner
  TestMigrate_TASK191_* = file marker only
  → N/A runtime mutate (не killable prod guard в процессе)
```

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -count=1 -timeout=120s ./testframework/tracker/ -run "TASK191|Task191"
go test -tags hubtest -count=1 -timeout=60s ./testframework/ent/ -run "TASK191|Task191"
# BASELINE PASS · FINAL PASS (tracker ~1.2s · ent ~0.8s)
```

## Gaps

- Migrate SQL — **N/A** (owner-apply gate; not a runtime kill).
- layer_mismatch / mock≠Ent / docs-only: **none**.

## Handoff

- Evidence: этот файл · TASK-191 **Закрыта** · README sync  
- be-ver → code-reviewer  
- Production restored after each BREAK; `git diff` на `repo_automation.go` + `schema/automationrule.go` — empty · no leftover MUTATED  
- Migrate остаётся **BLOCKED waiting owner** (отдельный ops gate, не блокер mutation-close)
