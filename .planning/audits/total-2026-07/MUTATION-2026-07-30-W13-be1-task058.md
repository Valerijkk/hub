# MUTATION W13 slot 1 refill — TASK-058 (be1) · 2026-07-30

> Checker: **be1** · Env: localhost Postgres/Mongo · Go 1.25 · `-tags hubtest`  
> Criteria: `.planning/backlog/archive/legacy/TASK-058-be-p3-assert-free-tests.md` `[MUTATION CRITERIA]`  
> Driver: `.planning/audits/total-2026-07/_run_mutations_w13_task058.py`  
> Raw: `_mut-batch-w13-task058.json`  
> Коммит не создавался (инструкция слота).

## Verdict

**17/17 PROVEN** (из них **3 PROVEN + GAP_NOTED** mock≠Ent) · criteria residual **W6-A3-015 GAP_NOTED** · leftover MUTATED: **нет**  
TASK-058 → **Закрыта**

| ID | Verdict | RED test |
|----|---------|----------|
| W1-A4-016 | **PROVEN** | `TestHandler_StartSprint_NotFound` |
| W1-A3-011 | **PROVEN** | `TestAutomationConsumer_processMessage_Events` |
| W7-A1-009 | **PROVEN** | `TestIsQuietHours_CrossMidnight` |
| W7-A4-020 | **PROVEN** | `TestOrphanCleanup_AttemptsS3DeleteWhenClientConfigured` |
| W4-A2-008 | **PROVEN** | `TestWorkerConstants` |
| W4-A4-011 | **PROVEN** | `TestLoad_clearsEnvBetweenTests` |
| W6-A2-004 | **PROVEN** | `TestOrganizationService_PendingUsersCRUD` |
| W6-A2-005 | **PROVEN** | `TestOrganizationService_PurgeExpiredPending` |
| W5-A5-005 | **PROVEN** | `TestSearchScopedEntities_EmptyProjectIDsReturnNothing` |
| W5-A5-006 | **PROVEN** | `TestDeleteSavedFilter_FilterMissing` |
| W3-A3-011 | **PROVEN** + **GAP_NOTED** | `TestMockRepository_AddReactionContract` |
| W6-A4-014 | **PROVEN** | `TestProcessEvent_InvalidJSON_NaksWithoutCreate` |
| W6-A5-007 | **PROVEN** | `TestAuditService_ListTrash_WithItems` |
| W6-A5-008 | **PROVEN** + **GAP_NOTED** | `TestAuditService_ReassignTasks` |
| W6-A5-009 | **PROVEN** + **GAP_NOTED** | `TestRecordHTTPRequest_SkipsOnCreateError` |
| W2-A5-010 | **PROVEN** | `TestMapServiceError_AllBranches` |
| W3-A1-012 | **PROVEN** | `TestHandler_Wave123_MuteChannel_Success` |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: W3-A3-011, W6-A5-008, W6-A5-009` · `docs-only IDs: (none)` · residual class gap: **W6-A3-015 GAP_NOTED**

## Per-ID evidence

### W1-A4-016 — StartSprint missing → HTTP 404

```text
BASELINE TestHandler_StartSprint_NotFound -> PASS (~0.9s)
BREAK: sprint/service.go StartSprint — NotFound → BadRequest
-> FAIL Equal 404 got 400
RESTORE -> PASS
```

Guard: `backend/internal/sprint/service.go` StartSprint NotFound branch.

### W1-A3-011 — automation Ack/Nak

```text
BASELINE TestAutomationConsumer_processMessage_Events -> PASS
BREAK: automation_consumer.go processErr → Ack instead of Nak
-> FAIL unknown/bad JSON naks==1 (got 0), acks==0 (got 1)
RESTORE -> PASS
```

### W7-A1-009 — quiet-hours cross-midnight

```text
BASELINE TestIsQuietHours_CrossMidnight -> PASS
BREAK: invert cross-midnight return
-> FAIL assert.True / assert.False
RESTORE -> PASS
```

### W7-A4-020 — S3 delete fail keeps meta

```text
BASELINE TestOrphanCleanup_AttemptsS3DeleteWhenClientConfigured -> PASS
BREAK: orphan_cleanup.go — remove continue on S3 err (meta still deleted)
-> FAIL meta Get NoError (row gone)
RESTORE -> PASS
```

### W4-A2-008 — outbox BatchLimit export

```text
BASELINE TestWorkerConstants -> PASS
BREAK: batchLimit 50→99
-> FAIL ExportBatchLimit Equal 50
RESTORE -> PASS
```

### W4-A4-011 — config Load DEV_MODE

```text
BASELINE TestLoad_clearsEnvBetweenTests -> PASS
BREAK: DevMode always false (ignore env)
-> FAIL expected DevMode=true
RESTORE -> PASS
```

### W6-A2-004 — PendingUsers Approve state

```text
BASELINE TestOrganizationService_PendingUsersCRUD -> PASS
BREAK: ApprovePendingUser no-op
-> FAIL StatusActive assert (stays Pending)
RESTORE -> PASS
```

### W6-A2-005 — PurgeExpiredPending cutoff

```text
BASELINE TestOrganizationService_PurgeExpiredPending -> PASS
BREAK: cutoff −1y instead of −7d
-> FAIL purge count Equal 1 (got 0)
RESTORE -> PASS
```

### W5-A5-005 — empty projectIDs → 0 hits

```text
BASELINE TestSearchScopedEntities_EmptyProjectIDsReturnNothing -> PASS
BREAK: searchIssues empty → fake hit
-> FAIL assert.Nil issues
RESTORE -> PASS
```

### W5-A5-006 — ErrFilterNotFound sentinel

```text
BASELINE TestDeleteSavedFilter_FilterMissing -> PASS
BREAK: return fmt.Errorf instead of ErrFilterNotFound
-> FAIL ErrorIs ErrFilterNotFound
RESTORE -> PASS
```

### W3-A3-011 — mock AddReaction NotFound

```text
BASELINE TestMockRepository_AddReactionContract -> PASS
BREAK: mock AddReaction missing → nil (no NotFound)
-> FAIL require.Error / IsNotFound
RESTORE -> PASS
```

**GAP_NOTED** mock≠Ent: kill на fixture mock (намеренно criteria).

### W6-A4-014 — audit invalid JSON Nak

```text
BASELINE TestProcessEvent_InvalidJSON_NaksWithoutCreate -> PASS
BREAK: Ack on bad JSON
-> FAIL naks==1 (got 0)
RESTORE -> PASS
```

### W6-A5-007 — ListTrash cutoff ≈ now−30d

```text
BASELINE TestAuditService_ListTrash_WithItems -> PASS
BREAK: ListTrash cutoff −7d (TrashRetentionDays stays 30)
-> FAIL cutoff window assert
RESTORE -> PASS
```

### W6-A5-008 — ReassignTasks counters

```text
BASELINE TestAuditService_ReassignTasks -> PASS
BREAK: ReassignTasks return 0 without repo call
-> FAIL count Equal 3 / call-count
RESTORE -> PASS
```

**GAP_NOTED** mock≠Ent.

### W6-A5-009 — RecordHTTPRequest createErr skip

```text
BASELINE TestRecordHTTPRequest_SkipsOnCreateError -> PASS
BREAK: never call CreateAuditLog
-> FAIL createCalls Len 1 (got 0)
RESTORE -> PASS
```

**GAP_NOTED** mock≠Ent.

### W2-A5-010 — ErrForbidden → 403

```text
BASELINE TestMapServiceError_AllBranches -> PASS
BREAK: drop Forbidden case (falls to 500)
-> FAIL Equal 403 got 500
RESTORE -> PASS
```

### W3-A1-012 — unmute = POST /unmute

```text
BASELINE TestHandler_Wave123_MuteChannel_Success -> PASS
BREAK: remove POST /unmute route
-> FAIL unmute StatusOK (got 404/405)
RESTORE -> PASS
```

## Commands

```powershell
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"
$env:PYTHONIOENCODING="utf-8"
cd c:\Users\Valerii\Desktop\hub
python .planning/audits/total-2026-07/_run_mutations_w13_task058.py
# exit 0 · 17/17 PROVEN · leftover_mutated=false · ~280s
```

## Gaps

- Docs-only: **N/A**
- Class residual без отдельного assert в criteria: **W6-A3-015 GAP_NOTED**
- mock≠Ent (intentional): W3-A3-011, W6-A5-008, W6-A5-009 — kill PROVEN на service/fixture path

## Leftover check

`Select-String MUTATED` on all touched prod/fixture files → empty after final `git checkout --`.
