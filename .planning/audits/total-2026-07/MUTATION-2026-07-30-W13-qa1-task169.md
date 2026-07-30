# MUTATION W13 slot 4 refill — TASK-169 (qa-auto1) · 2026-07-30

> Checker: **qa-auto1** · Env: localhost `POSTGRES_TEST_URL` / `MONGODB_TEST_URL` · Go 1.25  
> Criteria: `.planning/backlog/TASK-169-2026-07-29-a3-8-3-schemas.md` `[MUTATION CRITERIA]`  
> Коммит: **не создавался** (инструкция слота).

## Verdict

**2/2 PROVEN** · **0 NOT_PROVEN** · TASK-169 → **Закрыта**

| ID | Verdict | RED test | Mutation (BREAK) | Layer |
|----|---------|----------|------------------|-------|
| A3W8-3-001 | **PROVEN** | `TestCascade_TASK169_HubTableHardDeleteRemovesColumnsRecords` (+ `TestSchema_TASK169_OnDeleteCascades`) | HubTable `Cascade`→`SetNull` + migrate FK `hub_columns_*` / `hub_records_*` → `NoAction` | schema + migrate DDL |
| A3W8-3-002 | **PROVEN** | `TestCascade_TASK169_JournalEntryHardDeleteRemovesLines` (+ `TestSchema_TASK169_OnDeleteCascades`) | JournalEntry lines `Cascade`→`SetNull` + migrate FK `journal_lines_*` → `NoAction` | schema + migrate DDL |

Production restored (`git checkout --` schema; migrate backup rewrite) после каждого BREAK. Leftover MUTATED: **нет**.  
`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)`.

## Per-ID evidence

### A3W8-3-001 — HubTable→columns/records Cascade

```text
BASELINE TASK169 → PASS (~4.6s)

BREAK: hubtable.go
  cascade := entsql.OnDelete(entsql.Cascade)
  → entsql.OnDelete(entsql.SetNull) // MUTATED A3W8-3-001
  migrate schema.go:
  hub_columns_hub_tables_columns Cascade→NoAction
  hub_records_hub_tables_records Cascade→NoAction
→ FAIL TestSchema_TASK169_OnDeleteCascades
        hubtable.go missing "entsql.OnDelete(entsql.Cascade)"
→ FAIL TestCascade_TASK169_HubTableHardDeleteRemovesColumnsRecords
        ent: constraint failed: FOREIGN KEY constraint failed
RESTORE → PASS (~4.4s)
```

### A3W8-3-002 — JournalEntry→lines Cascade

```text
BREAK: journalentry.go
  Annotations(entsql.OnDelete(entsql.Cascade))
  → Annotations(entsql.OnDelete(entsql.SetNull)) // MUTATED A3W8-3-002
  migrate schema.go:
  journal_lines_journal_entries_lines Cascade→NoAction
→ FAIL TestSchema_TASK169_OnDeleteCascades
        journalentry.go missing "entsql.OnDelete(entsql.Cascade)"
→ FAIL TestCascade_TASK169_JournalEntryHardDeleteRemovesLines
        ent: constraint failed: FOREIGN KEY constraint failed
RESTORE → PASS (~4.0s)
```

```text
FINAL TASK169 → PASS (~3.6s)
leftover_mutated=False
```

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -count=1 -timeout=180s -vet=off ./testframework/accounting/ -run TASK169
```

Runner: `.planning/audits/total-2026-07/_run_mutations_w13_task169.py`  
JSON: `_mut-batch-w13-task169.json`.

## Gaps

- Soft-delete tests (`*_SoftDeleteKeeps*`) — anti-regression, не kill-target (OnDelete не вызывается на soft path) — **GAP_NOTED**, не NOT_PROVEN.
- Schema-source Contains-guard — companion; канон Exact RED — HardDelete* выше.
- Migrate SQL `20260729_task169_*` — needs-owner-apply (unchanged).
- Dup `testframework/ent/schema_ondelete_task169_test.go` — out of this run package; accounting = criteria Exact.

## Handoff

- Evidence: этот файл · TASK-169 **Закрыта** · README sync  
- be-ver → code-reviewer (fix/tests already FIXED; mutation PROVEN)  
- Коммит: не создавался
