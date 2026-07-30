# 🔒 Audit BE: tracker soft-delete status/board + messenger FindMessage/CreateChannel

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **Закрыта**  
> Источник: [`AUDIT-BE-tracker-messenger-2026-07-29.md`](../../../audits/total-2026-07/AUDIT-BE-tracker-messenger-2026-07-29.md)  
> Исполнители: **be1** (tracker + messenger) · Finding IDs: `AUDIT-TM-001`…`004`  
> Mutation: [`MUTATION-2026-07-30-W13-be1-task190.md`](../../../audits/total-2026-07/MUTATION-2026-07-30-W13-be1-task190.md) · 4/4 PROVEN (+004 GAP_NOTED mock≠Ent)

## [OUTCOME]

Default/board workflow statuses не поднимают soft-deleted; `FindMessageByID`/`UpdateMessage` не обслуживают trash-сообщения; `CreateChannel` не принимает чужой `project_id` без membership.

## [ROLE]

- be1: AUDIT-TM-001…004
- Handoff: be-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Allowed: `backend/internal/tracker/**`, `backend/internal/messenger/**` (по ID); tests tracker/messenger
- Optional P3 TM-005…008 в том же PR если дёшево
- Do not touch: откат A3 TASK-134/144/145 guards
- Commits EN/RU; no AI co-author

## [ACCEPTANCE]

- 001/002: soft-deleted status не default / не на board; mutation RED
- 003: soft-deleted message → NotFound на Edit/React/Pin path
- 004: foreign project_id CreateChannel → 403/404
- Stop: 4/4 FIXED · mutation 4/4 PROVEN · «Закрыта»

## [EVIDENCE] be1 2026-07-29

**Verdict: 4/4 FIXED** · статус был **На проверке** (до mutation W13)

| ID | Status | Evidence |
|----|--------|----------|
| AUDIT-TM-001 | FIXED | `findDefaultStatusTx` + `DeletedAtIsNil`; `TestCreateIssue_RejectsSoftDeletedDefaultStatus_AUDIT_TM_001` |
| AUDIT-TM-002 | FIXED | `GetBoardData` `WithWorkflowStatuses` + `DeletedAtIsNil`; `TestGetBoardData_SkipsSoftDeletedStatuses_AUDIT_TM_002` |
| AUDIT-TM-003 | FIXED | Mongo `FindMessageByID`/`UpdateMessage` filter `deleted_at:nil`; `TestFindMessageByID_SkipsSoftDeleted_AUDIT_TM_003` |
| AUDIT-TM-004 | FIXED | `IsProjectMember` + CreateChannel 403; `TestCreateChannel_RejectsForeignProject_AUDIT_TM_004` |

```
go test -tags hubtest -race -count=1 -timeout=600s ./testframework/tracker/ ./testframework/messenger/
ok  tracker   ~12s
ok  messenger ~5s
```

### Mutation notes

| ID | Сломать | Ожидание |
|----|---------|----------|
| 001 | убрать DeletedAtIsNil из findDefaultStatusTx | CreateIssue soft-default тест FAIL |
| 002 | убрать DeletedAtIsNil в WithWorkflowStatuses | board soft-status тест FAIL |
| 003 | убрать deleted_at:nil из Find/Update | soft-deleted Find/Update тест FAIL |
| 004 | убрать IsProjectMember check | foreign CreateChannel 403 тест FAIL |

## [EVIDENCE] mutation be1 2026-07-30 W13 slot 1 refill

**Verdict: 4/4 PROVEN** (+004 GAP_NOTED mock≠Ent) · статус **Закрыта**

| ID | Verdict | RED |
|----|---------|-----|
| AUDIT-TM-001 | PROVEN | `TestCreateIssue_RejectsSoftDeletedDefaultStatus_AUDIT_TM_001` — err expected, got nil |
| AUDIT-TM-002 | PROVEN | `TestGetBoardData_SkipsSoftDeletedStatuses_AUDIT_TM_002` — soft status on board |
| AUDIT-TM-003 | PROVEN | `TestFindMessageByID_SkipsSoftDeleted_AUDIT_TM_003` — ErrNoDocuments expected, got nil |
| AUDIT-TM-004 | PROVEN + GAP_NOTED | `TestCreateChannel_RejectsForeignProject_AUDIT_TM_004` — 403 expected, got nil |

Report: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-be1-task190.md`

Handoff: **be-ver → code-reviewer**

## [MUTATION CRITERIA] — verifier-ready

> W13 slot 3 refill / be2 · criteria only (no mutate) · 2026-07-30 · HEAD guards verified

| ID | Production file:line (guard) | слой | Exact test RED | Kill recipe |
|----|------------------------------|------|----------------|-------------|
| AUDIT-TM-001 | `backend/internal/tracker/repo_issues.go:141-148` — `findDefaultStatusTx`: `workflowstatus.DeletedAtIsNil()` | EntRepository | `TestCreateIssue_RejectsSoftDeletedDefaultStatus_AUDIT_TM_001` · `backend/testframework/tracker/audit_tm_softdelete_task190_test.go` | убрать `DeletedAtIsNil` из findDefaultStatusTx → CreateIssue OK на trash default → `ent.IsNotFound` assert RED |
| AUDIT-TM-002 | `backend/internal/tracker/repo_boards.go:72-74` — `GetBoardData` `WithWorkflowStatuses`: `workflowstatus.DeletedAtIsNil()` | EntRepository | `TestGetBoardData_SkipsSoftDeletedStatuses_AUDIT_TM_002` · `audit_tm_softdelete_task190_test.go` | убрать `DeletedAtIsNil` в WithWorkflowStatuses → soft status на board → skip-assert RED |
| AUDIT-TM-003 | `backend/internal/messenger/mongo_repository.go:85-90` + `:131-136` — `FindMessageByID`/`UpdateMessage` filter `"deleted_at": nil` | MongoRepository | `TestFindMessageByID_SkipsSoftDeleted_AUDIT_TM_003` · `backend/testframework/messenger/audit_tm_softdelete_task190_test.go` | убрать `deleted_at: nil` из Find и/или Update filter → soft-deleted Find/Update succeeds → `ErrNoDocuments` assert RED |
| AUDIT-TM-004 | `backend/internal/messenger/service_channels.go:30-37` — `CreateChannel`: `IsProjectMember` → `NewForbidden` | service | `TestCreateChannel_RejectsForeignProject_AUDIT_TM_004` · `backend/testframework/messenger/service_test.go` | убрать `IsProjectMember` check / всегда `ok=true` → foreign project_id CreateChannel 201 → 403 assert RED |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Flag |
|----|-----------|------|
| 001 | Ent `WorkflowStatus` + `CreateIssue` tx | **OK** — Ent |
| 002 | Ent `GetBoardData` + soft WorkflowStatus | **OK** — Ent |
| 003 | live Mongo (`mongotest`) SoftDelete → Find/Update | **OK** — Mongo (не mock) |
| 004 | MessengerService + `mockRepository.IsProjectMember` | **mock≠Ent** (GAP_NOTED): service guard proven; Ent `IsProjectMember` / projectmember IT не kill-target |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: AUDIT-TM-004` · `docs-only IDs: (none)`

### Gaps

- Docs-only: **N/A** — все 4 ID runtime.
- Do not touch: откат A3 TASK-134/144/145 guards.
- Optional P3 TM-005…008 — вне критериев этой таблицы.
