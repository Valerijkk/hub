# 🔒 Audit BE: tracker soft-delete status/board + messenger FindMessage/CreateChannel

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **На проверке**  
> Источник: [`AUDIT-BE-tracker-messenger-2026-07-29.md`](../audits/total-2026-07/AUDIT-BE-tracker-messenger-2026-07-29.md)  
> Исполнители: **be1** (tracker + messenger) · Finding IDs: `AUDIT-TM-001`…`004`

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
- Stop: 4/4 FIXED · «На проверке»

## [EVIDENCE] be1 2026-07-29

**Verdict: 4/4 FIXED** · статус **На проверке**

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

Handoff: **be-ver → code-reviewer**
