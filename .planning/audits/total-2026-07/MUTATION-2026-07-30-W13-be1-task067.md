# MUTATION W13 slot 1 refill — TASK-067 (be1) · 2026-07-30

> Checker: **be1** · Env: localhost Postgres/Mongo · Go 1.25 · `-tags hubtest`  
> Criteria: `.planning/backlog/archive/legacy/TASK-067-be-p3-v4-dev-storage.md` `[MUTATION CRITERIA]`  
> Driver: `.planning/audits/total-2026-07/_run_mutations_w13_task067.py`  
> Raw: `_mut-batch-w13-task067.json`  
> Коммит не создавался (инструкция слота).

## Verdict

**9/10 PROVEN** · **1 GAP_NOTED** (W7-A2-012, no dedicated RED) · **0 NOT_PROVEN** · leftover MUTATED: **нет**  
TASK-067 → **Закрыта** (runnable done; NOT_PROVEN отсутствует)

| ID | Verdict | RED |
|----|---------|-----|
| W7-A2-012 | **GAP_NOTED** | *(none — self-guard DSN only)* |
| W7-A2-014 | **PROVEN** | `TestCoffeePartner_NoEmailPII_Task067` |
| W7-A2-016 | **PROVEN** | `TestDeleteNews_WritesAudit_Task067` |
| W7-A3-013 | **PROVEN** | `TestAllocateIssueKey_PrefersDefaultStatus_Task067` |
| W7-A4-015 | **PROVEN** | `TestOrphanCleanup_DrainsBeyondBatch_Task067` |
| W7-A4-016 | **PROVEN** | `TestApplyTemplate_StrangerForbidden_Task067`; `TestSaveLayout_StrangerProjectForbidden_Task067` |
| W7-A4-018 | **PROVEN** | `TestEntRepository_SetRestoredAtAndDelete` |
| W7-A5-012 | **PROVEN** | `TestLogin_RejectsNonActive_Task067` |
| W7-A5-013 | **PROVEN** | `TestSeedWorkflowTransitions_CreateFailPropagates_A3W75002`; `TestSeedIssueLinks_CreateFailPropagates_A3W75002`; `TestSeedChannels_CreateFailPropagates_A3W75002` |
| W7-A5-014 | **PROVEN** | `TestSeedData_NoTelegramLoginCopy_Task067` |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)` · `docs-only IDs: (none)` · `gap IDs: W7-A2-012`

## Per-ID evidence

### W7-A2-012 — uuid SQLite DSN — GAP_NOTED

```text
No dedicated named RED on uuid-in-DSN.
Guard: backend/testframework/intranet/handler_test.go:23 (self-guard hygiene).
Kill recipe would share DSN across parallel tests → flaky; no assert exists to trip RED.
Verdict: GAP_NOTED
```

### W7-A2-014 — CoffeePartner no email PII

```text
BASELINE TestCoffeePartner_NoEmailPII_Task067 -> PASS (~6.5s)
BREAK: Email field in CoffeePartnerResponse + toCoffeePartner
-> FAIL NotContains "email" / secret@hub.local
RESTORE -> PASS (~5.5s)
```

### W7-A2-016 — SoftDeleteNewsPost AuditLog

```text
BASELINE TestDeleteNews_WritesAudit_Task067 -> PASS (~5.6s)
BREAK: drop AuditLog.Create in SoftDeleteNewsPost
-> FAIL Equal(1) got 0
RESTORE -> PASS (~5.4s)
```

### W7-A3-013 — AllocateIssueKey IsDefault

```text
BASELINE TestAllocateIssueKey_PrefersDefaultStatus_Task067 -> PASS (~7.0s)
BREAK: First() by SortOrder without IsDefaultEQ
-> FAIL Equal(todo.ID) got Done
RESTORE -> PASS (~6.4s)
```

### W7-A4-015 — OrphanCleanup drain loop

```text
BASELINE TestOrphanCleanup_DrainsBeyondBatch_Task067 -> PASS (~7.1s)
BREAK: single batch, no for-drain
-> FAIL Get orphan expected Error got nil (tail remains)
RESTORE -> PASS (~6.1s)
```

### W7-A4-016 — Dashboard layout ACL

```text
BASELINE ApplyTemplate+SaveLayout stranger tests -> PASS (~7.0s)
BREAK: drop IsProjectMember in ApplyTemplate + SaveLayout
-> FAIL both: expected ErrForbidden got nil
RESTORE -> PASS (~6.2s)
```

### W7-A4-018 — SetRestoredAt not ErrorMessage

```text
BASELINE TestEntRepository_SetRestoredAtAndDelete -> PASS (~7.7s)
BREAK: SetErrorMessage(at) instead of SetRestoredAt
-> FAIL RestoredAt NotNil expected
RESTORE -> PASS (~5.7s)
```

### W7-A5-012 — dev login StatusActive

```text
BASELINE TestLogin_RejectsNonActive_Task067 -> PASS (~5.1s)
BREAK: drop StatusActive check
-> FAIL expected 403 got 200
RESTORE -> PASS (~4.8s)
```

### W7-A5-013 — seed helper fail-propagate

```text
BASELINE three Seed*CreateFailPropagates tests -> PASS (~4.9s)
BREAK: continue / swallow Create err in transitions+links+channels
-> FAIL all three: Error expected got nil
RESTORE -> PASS (~4.8s)
```

Outer `seed.go:194-198` partial→500 not separately mutated (helper RED covers W7-A5-013 criteria; HTTP twin residual noted in TASK gaps).

### W7-A5-014 — demo Telegram = notification bind

```text
BASELINE TestSeedData_NoTelegramLoginCopy_Task067 -> PASS (~5.0s)
BREAK: titles «Telegram вход…» / «Telegram авторизация…»
-> FAIL NotContains telegram вход / авторизация
RESTORE -> PASS (~4.5s)
```

## Residual

- **W7-A2-012 GAP_NOTED** — нет dedicated RED на uuid DSN; не блокирует «Закрыта» (NOT_PROVEN=0).
- W7-A5-013: helper-level PROVEN; полный HTTP Seed `partial` body twin — желателен (как в criteria Gaps).
- leftover `MUTATED` в prod: **нет** (`git checkout` после каждого ID; final restore clean).
