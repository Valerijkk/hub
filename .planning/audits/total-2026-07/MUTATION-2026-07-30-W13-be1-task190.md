# MUTATION W13 slot 1 refill — TASK-190 (be1) · 2026-07-30

> Checker: **be1** · Env: localhost Postgres/Mongo · Go 1.25 · `-tags hubtest`  
> Criteria: `.planning/backlog/TASK-190-2026-07-29-audit-be-tracker-messenger-softdelete.md` `[MUTATION CRITERIA]`  
> Коммит не создавался (инструкция слота).

## Verdict

**4/4 PROVEN** (+004 GAP_NOTED mock≠Ent) · TASK-190 → **Закрыта** · leftover MUTATED: **нет**

| ID | Verdict | RED test |
|----|---------|----------|
| AUDIT-TM-001 | **PROVEN** | `TestCreateIssue_RejectsSoftDeletedDefaultStatus_AUDIT_TM_001` |
| AUDIT-TM-002 | **PROVEN** | `TestGetBoardData_SkipsSoftDeletedStatuses_AUDIT_TM_002` |
| AUDIT-TM-003 | **PROVEN** | `TestFindMessageByID_SkipsSoftDeleted_AUDIT_TM_003` |
| AUDIT-TM-004 | **PROVEN** + **GAP_NOTED** | `TestCreateChannel_RejectsForeignProject_AUDIT_TM_004` |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: AUDIT-TM-004` · `docs-only IDs: (none)`

## Per-ID evidence

### AUDIT-TM-001 — findDefaultStatusTx `DeletedAtIsNil`

```text
BASELINE TestCreateIssue_RejectsSoftDeletedDefaultStatus_AUDIT_TM_001 -> PASS (~0.9s)

BREAK: tracker/repo_issues.go findDefaultStatusTx —
  убрать workflowstatus.DeletedAtIsNil()
-> FAIL audit_tm_softdelete_task190_test.go:33
   An error is expected but got nil.
RESTORE -> PASS (~1.0s, paired with TM-002 RED run)
```

Guard: `backend/internal/tracker/repo_issues.go:141-148` — EntRepository.

### AUDIT-TM-002 — GetBoardData `WithWorkflowStatuses` `DeletedAtIsNil`

```text
BASELINE TestGetBoardData_SkipsSoftDeletedStatuses_AUDIT_TM_002 -> PASS (~0.9s)

BREAK: tracker/repo_boards.go WithWorkflowStatuses —
  убрать sq.Where(workflowstatus.DeletedAtIsNil())
-> FAIL audit_tm_softdelete_task190_test.go:54-55
   soft-deleted status must not appear on board; DeletedAt Expected nil
RESTORE -> PASS (~1.1s)
```

Guard: `backend/internal/tracker/repo_boards.go:72-74` — EntRepository.

### AUDIT-TM-003 — FindMessageByID / UpdateMessage `deleted_at: nil`

```text
BASELINE TestFindMessageByID_SkipsSoftDeleted_AUDIT_TM_003 -> PASS (~0.9s)

BREAK: messenger/mongo_repository.go FindMessageByID + UpdateMessage —
  убрать "deleted_at": nil из обоих filter
-> FAIL audit_tm_softdelete_task190_test.go:38
   Expected mongo.ErrNoDocuments but got nil (Find soft-deleted succeeded)
RESTORE -> PASS (~1.0s, paired with TM-004 RED run)
```

Guard: `backend/internal/messenger/mongo_repository.go:85-90` + `:131-136` — MongoRepository (live Mongo).

### AUDIT-TM-004 — CreateChannel `IsProjectMember` → Forbidden

```text
BASELINE TestCreateChannel_RejectsForeignProject_AUDIT_TM_004 -> PASS (~0.9s)

BREAK: messenger/service_channels.go CreateChannel —
  убрать IsProjectMember check / всегда пропускать project_id
-> FAIL service_test.go:499
   expected AppError 403, got <nil>
RESTORE -> PASS (~0.9s)
```

Guard: `backend/internal/messenger/service_channels.go:30-37` — service.

**GAP_NOTED** mock≠Ent: kill на prod service path PROVEN через `mockRepository.IsProjectMember`; Ent `IsProjectMember` / projectmember IT не kill-target.

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -count=1 -timeout=180s ./testframework/tracker/ -run `
  "TestCreateIssue_RejectsSoftDeletedDefaultStatus_AUDIT_TM_001|TestGetBoardData_SkipsSoftDeletedStatuses_AUDIT_TM_002"
# FINAL GREEN PASS ~1.0s

go test -tags hubtest -count=1 -timeout=180s ./testframework/messenger/ -run `
  "TestFindMessageByID_SkipsSoftDeleted_AUDIT_TM_003|TestCreateChannel_RejectsForeignProject_AUDIT_TM_004"
# FINAL GREEN PASS ~0.9s
```

## Gaps

- Docs-only: **N/A**
- mock≠Ent: **AUDIT-TM-004** — service guard proven; Ent membership IT вне scope
- Do not touch: откат A3 TASK-134/144/145 guards
- Optional P3 TM-005…008 — вне критериев

## Handoff

- Evidence: этот файл · TASK-190 **Закрыта** · README sync  
- be-ver → code-reviewer  
- Production restored after BREAK; `git diff` на 4 guard-файлах — empty; no `MUTATED` leftover  
- Prod code unchanged this slot (fix already on HEAD; mutation prove only)
