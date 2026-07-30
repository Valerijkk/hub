# MUTATION W13 slot 4 refill — TASK-161 (qa-auto1) · 2026-07-30

> Checker: **qa-auto1** · Env: localhost `POSTGRES_TEST_URL` / `MONGODB_TEST_URL` · Go 1.25  
> Criteria: `.planning/backlog/TASK-161-2026-07-29-a3-6-5-audit-tests.md` `[MUTATION CRITERIA]`  
> Коммит: **не создавался** (инструкция слота).

## Verdict

**4/4 killable PROVEN** (001–002 **PROVEN + GAP_NOTED** mock≠Ent) · **3 N/A** hollow · **0 NOT_PROVEN** · TASK-161 → **Закрыта**

| ID | Verdict | RED test | Mutation (BREAK) | Layer |
|----|---------|----------|------------------|-------|
| A3W6-5-001 | **PROVEN** + **GAP_NOTED** | `TestAuditService_PurgeExpired_CapturesCutoffs_A2W6A5001` | убрать `purge("crm_contacts"/"crm_deals")` | service + mock callCounts |
| A3W6-5-002 | **PROVEN** + **GAP_NOTED** | `TestAuditService_SoftDelete_RecordsAllTypes_A2W6A5004`; `TestAuditService_RestoreItem_RecordsAllTypes_A2W6A5004` | убрать cases `"crm_contact"`/`"crm_deal"` SoftDelete+Restore | service + mock maps |
| A3W6-5-003 | **PROVEN** | `TestRepository_CRMContact_RetentionHidesOlderThan30Days`; `TestRepository_CRMDeal_RetentionHidesOlderThan30Days` | снять `DeletedAtGT(cutoff)` в ListDeletedCRM* | **EntRepository** |
| A3W6-5-004 | **PROVEN** | `TestAuditService_DeactivateUser_InvalidatesCache` | `if false { InvalidateUserStatusCache… }` | service Redis |
| A3W6-5-005 | **N/A** | N/A | hollow `ListTrash_WithProjectFilter` удалён; цемент CapturesProjectID | N/A |
| A3W6-5-006 | **N/A** | N/A | hollow Succeeds/AllTypes удалены; kill = 001/002 | N/A |
| A3W6-5-007 | **N/A** | N/A | hollow `DeleteUserGDPR_Success` удалён; цемент SuccessCascadeCalls | N/A |

Production restored (`git checkout --`) после каждого BREAK. Leftover MUTATED: **нет**.  
`layer_mismatch IDs: (none)` · `mock≠Ent IDs: A3W6-5-001, A3W6-5-002` (GAP_NOTED, не NOT_PROVEN).

## Per-ID evidence

### A3W6-5-001 — PurgeExpired CRM keys

```text
BASELINE → PASS (~5.5s)

BREAK: service.go PurgeExpired — убрать purge crm_contacts/crm_deals
→ FAIL: purge crm_contacts must be called once (expected 1, got 0)
→ FAIL: cutoff for crm_contacts Should be true
RESTORE → PASS (~5.1s)
```

**GAP_NOTED:** assert через `purgeMockRepo.callCounts` (mock≠Ent); prod wiring = service purge keys. Ent PurgeExpiredCRM* не мутировался.

### A3W6-5-002 — SoftDelete/Restore CRM cases

```text
BASELINE → PASS (~5.3s)

BREAK: service.go SoftDelete + RestoreItem — убрать cases crm_contact/crm_deal
→ FAIL SoftDelete: неизвестный тип сущности: crm_contact
→ FAIL RestoreItem: неизвестный тип сущности: crm_contact
RESTORE → PASS (~5.9s)
```

**GAP_NOTED:** RecordsAllTypes бьёт mock softDeleteCalls/restoreCalls; prod = service switch.

### A3W6-5-003 — ListDeletedCRM* retention (Ent)

```text
BASELINE → PASS (~5.0s)

BREAK: repository.go — DeletedAtNotNil() без DeletedAtGT(cutoff) для CRM contact+deal
→ FAIL CRMContact: total expected 0, got 1 (+ items not empty)
→ FAIL CRMDeal: total expected 0, got 1 (+ items not empty)
RESTORE → PASS (~5.3s)
```

Ent OK — predicate на реальной sqlite Ent-схеме.

### A3W6-5-004 — Deactivate cache invalidate

```text
BASELINE → PASS (~5.5s)

BREAK: deactivateUser — if false { auth.InvalidateUserStatusCache… } (import сохранён)
→ FAIL: InvalidateUserStatusCache must Del user status key (Should be false, key still Exists)
RESTORE → PASS (~8.0s)
```

(Первый naive kill «удалить вызов» дал unused-import build fail — перепрогон с `if false` для assert-level RED.)

### A3W6-5-005 / 006 / 007 — hollow N/A

```text
ListTrash_WithProjectFilter — нет func Test* в testframework/audit/
PurgeExpired_Succeeds / SoftDelete_AllTypes / RestoreItem_AllTypes — нет
DeleteUserGDPR_Success (без Cascade) — нет; остаётся SuccessCascadeCalls_A2W6A5009
```

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -count=1 -timeout=300s ./testframework/audit/ `
  -run "TestAuditService_PurgeExpired_CapturesCutoffs_A2W6A5001|TestAuditService_SoftDelete_RecordsAllTypes_A2W6A5004|TestAuditService_RestoreItem_RecordsAllTypes_A2W6A5004|TestRepository_CRMContact_RetentionHidesOlderThan30Days|TestRepository_CRMDeal_RetentionHidesOlderThan30Days|TestAuditService_DeactivateUser_InvalidatesCache|TestHandler_ListTrash_CapturesProjectID_A2W6A5006|TestAuditService_DeleteUserGDPR_SuccessCascadeCalls_A2W6A5009"
```

Runner: `.planning/audits/total-2026-07/_run_mutations_w13_task161.py` · JSON: `_mut-batch-w13-task161.json`.  
004 assert-level re-prove: inline `if false` kill (не в script first pass).

## Gaps

- 001/002: **GAP_NOTED** mock≠Ent — service-layer wiring доказан; Ent CRM purge/soft/restore predicates не kill target этого ID.
- 005–007: N/A hollow deletion (не docs-only).
- Не ослабляли PropagatesError / non-CRM RetentionHides.

## Handoff

- Evidence: этот файл · TASK-161 **Закрыта** · README sync  
- be-ver → code-reviewer (fix/tests already FIXED; mutation PROVEN / N/A)  
- Коммит: не создавался
