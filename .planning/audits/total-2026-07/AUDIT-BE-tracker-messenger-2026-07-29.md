# AUDIT-BE — tracker + messenger (file-auditor)

> TASK-008 · File-auditor · READ-ONLY · 2026-07-29  
> HEAD: `617af22372be61b98e7ea63d825b35075f495588` (`617af223`) · model `cursor-grok-4.5`  
> Периметр: **все** production `.go` (без `*_test.go`) в  
> `backend/internal/tracker/` + `backend/internal/messenger/`  
> Фокус: **IDOR · soft-delete · caps · AppError · layering**  
> Артефакт: только этот файл · **не коммитить** · **TASK-008 оставить открытой**

## Итог

Повторный пофайловый проход по HEAD после remediation A3 (TASK-134 / TASK-143…144 / TASK-108 и др.).  
Большинство A3/A2 closes по IDOR/RBAC/soft-delete/caps **подтверждены**.  
Остались **8 активных** находок (0×P0 · 0×P1 · **4×P2** · **4×P3**).

**TASK-008 не закрывать.**

## Прочитанные файлы

### `backend/internal/tracker/` (44)

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| automation_consumer.go | 156 | 0 | чисто — FilterSubjects moved/created/assigned; known no-op Ack |
| automation_dto.go | 73 | 0 | чисто |
| automation_engine.go | 144 | 0 | чисто — Post/Notify err → return (Nak path) |
| automation_service.go | 189 | 0 | чисто — channel_id ∈ project; max rules cap |
| dto.go | 320 | 0 | чисто — validate/version на Update* |
| events.go | 75 | 0 | чисто |
| export_hooks.go | 116 | 0 | чисто — hubtest re-exports; без *ent.Client в handler |
| handler.go | 98 | 0 | чисто — Fiber v3 RBAC order; project-scoped routes |
| handler_automation.go | 121 | 0 | чисто — ParseProjectID + scoped service |
| handler_boards.go | 88 | 0 | чисто |
| handler_comments.go | 222 | 0 | чисто — project+issueKey на все мутации |
| handler_custom_fields.go | 86 | 0 | чисто — DeleteCustomField(projectID, fieldID) |
| handler_issue_message_link.go | 113 | 0 | чисто |
| handler_issues.go | 254 | 0 | чисто — JQL через service; cursor list |
| handler_labels.go | 137 | 0 | чисто |
| handler_links.go | 124 | 0 | чисто |
| handler_time.go | 138 | 0 | чисто |
| handler_workflow.go | 197 | 0 | чисто — status/transition project-scoped |
| provisioner.go | 23 | 0 | чисто |
| repository.go | 279 | 0 | чисто — ISP; *ent.Client только в EntRepository |
| repo_activity.go | 44 | 0 | чисто — MaxIssueActivities Limit |
| repo_automation.go | 116 | 0 | чисто — ProjectIDEQ Update/Delete |
| repo_boards.go | 112 | 1 | **AUDIT-TM-002** — статусы без soft-delete filter |
| repo_comments.go | 194 | 0 | чисто — soft-delete; VersionEQ; MaxIssueComments |
| repo_custom_fields.go | 54 | 0 | чисто — soft Delete + project scope |
| repo_issue_message_link.go | 87 | 0 | чисто — FindChannelInProject + membership |
| repo_issues.go | 474 | 1 | **AUDIT-TM-001** — findDefaultStatusTx без DeletedAtIsNil |
| repo_labels.go | 96 | 0 | чисто — soft Delete scoped; CountLabelsInProject |
| repo_links.go | 123 | 1 | **AUDIT-TM-005** — ListIssueLinks без Limit |
| repo_time.go | 172 | 0 | чисто — StartTimer lock + DeletedAtIsNil; MaxIssueTimeLogs |
| repo_workflow.go | 272 | 0 | чисто — soft-delete status; FindDefaultStatus с DeletedAtIsNil |
| service.go | 87 | 0 | чисто — layering: repo interface, без *ent.Client |
| service_activity.go | 29 | 0 | чисто |
| service_boards.go | 115 | 0 | чисто — GetBoardData project scope |
| service_comments.go | 127 | 0 | чисто — AppError 404/409 passthrough |
| service_converters.go | 301 | 0 | чисто |
| service_custom_fields.go | 49 | 0 | чисто |
| service_issue_message_link.go | 227 | 0 | чисто — channel∈project + IsChannelMember |
| service_issues.go | 449 | 0 | чисто — assignee/labels/status/parent scope; List limit≤100 |
| service_labels.go | 81 | 0 | чисто — FindLabelGroupInProject |
| service_links.go | 120 | 0 | чисто — FindIssueInProject target; AppError wrap |
| service_time.go | 83 | 1 | **AUDIT-TM-006** — LogTime глотает AppError NotFound |
| service_workflow.go | 115 | 0 | чисто |
| workflow.go | 24 | 0 | чисто |

Файлов tracker: **44**. Прочитано: **44**. Пропущено: **0**.

### `backend/internal/messenger/` (33)

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| auto_channels.go | 153 | 0 | чисто — lifecycleCtx; CreateChannel+members через service |
| combined_repository.go | 13 | 0 | чисто — Ent+Mongo embedding |
| dto.go | 258 | 0 | чисто |
| ent_repository.go | 667 | 1 | **AUDIT-TM-007** — ToggleStar/Mute без soft-delete channel |
| events.go | 99 | 0 | чисто |
| export_hooks.go | 115 | 0 | чисто — hubtest |
| handler.go | 130 | 1† | †маршруты без project RBAC (см. AUDIT-TM-004) |
| handler_channels.go | 511 | 0 | чисто — parseChannelID; membership в service |
| handler_messages.go | 259 | 0 | чисто — queryInt cap 100 |
| handler_presence.go | 109 | 0 | чисто |
| handler_reactions.go | 368 | 0 | чисто — polls/pins/bookmarks/invites handlers |
| handler_search.go | 38 | 0 | чисто |
| hookfixture.go | 72 | 0 | чисто — in-memory @hub-ai |
| limits.go | 18 | 0 | чисто — list*/unread TTL caps |
| mock_repository.go | 660 | 0 | чисто — soft-delete + AddReaction dedup (test double) |
| mongo_repository.go | 290 | 1 | **AUDIT-TM-003** — FindMessageByID без deleted_at |
| presence.go | 168 | 0 | чисто — away=90s / stale=10m; godoc синхронизирован |
| repository.go | 205 | 0 | чисто — ISP |
| service.go | 255 | 0 | чисто — layering comment SC-14-3; presence cap+FilterUsersSharingChannel |
| service_bookmarks.go | 78 | 0 | чисто — IsMember + msg.ChannelID==channelID |
| service_bot.go | 172 | 0 | чисто — systemBotID Nil guards |
| service_channel_invites.go | 244 | 0 | чисто — UseCountLT; revoke scoped; token не в NATS |
| service_channels.go | 351 | 1 | **AUDIT-TM-004** — project_id без membership |
| service_dm.go | 43 | 0 | чисто — orphan cleanup |
| service_hook.go | 18 | 0 | чисто |
| service_messages.go | 282 | 1 | **AUDIT-TM-008** — Forward IsMember err→403 |
| service_pins.go | 62 | 0 | чисто — IsMember + msg∈channel |
| service_polls.go | 253 | 0 | чисто — membership до Close; GetPollResults scoped |
| service_promote.go | 281 | 0 | чисто — promoteCheck fail-closed + truncateRunes |
| service_reactions.go | 49 | 0 | чисто — path==msg.ChannelID + IsMember |
| service_search.go | 46 | 0 | чисто — empty scope → []; membership на channel_id |
| truncate.go | 13 | 0 | чисто |
| unread.go | 76 | 0 | чисто — TTL Expire; ClearChannelForUsers |

Файлов messenger: **33**. Прочитано: **33**. Пропущено: **0**.

**Итого файлов: 77. Прочитано: 77. Пропущено: 0.**

† `handler.go:60` — `channels.Post("/", h.CreateChannel)` только `authMw`; project-scope enforcement ожидается в service (сейчас отсутствует) → finding привязан к service.

## Находки (активные)

| ID | Severity | Фокус | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------|-------------|----------|-------------------|------------|
| AUDIT-TM-001 | P2 | soft-delete | `tracker/repo_issues.go:140-146` | `findDefaultStatusTx` фильтрует только `IsDefaultEQ(true)`, **без** `DeletedAtIsNil` | Публичный `FindDefaultStatus` (`repo_workflow.go:128-135`) уже исключает корзину; CreateIssue в tx может привязать задачу к soft-deleted default status | Добавить `workflowstatus.DeletedAtIsNil()`; Ent-тест: soft-delete default → CreateIssue NotFound/новый default |
| AUDIT-TM-002 | P2 | soft-delete | `tracker/repo_boards.go:72-73` | `GetBoardData` → `WithWorkflowStatuses` **без** `DeletedAtIsNil` (issues фильтруются) | Soft-deleted колонки снова на Kanban; ListWorkflowStatuses их скрывает → UI/API drift | `sq.Where(workflowstatus.DeletedAtIsNil())`; regression board после DeleteWorkflowStatus |
| AUDIT-TM-003 | P2 | soft-delete | `messenger/mongo_repository.go:84-90` | `FindMessageByID` — filter только `_id`, **без** `deleted_at: nil` | List/Search скрывают trash; Edit/React/Pin/Forward/Promote/Delete работают по UUID soft-deleted сообщения (контент из корзины) | Добавить `"deleted_at": nil` (или `$in: [nil, null]`); UpdateMessage `$set` только при `deleted_at: nil`; denial-тесты |
| AUDIT-TM-004 | P2 | IDOR | `messenger/service_channels.go:25-31` (+ `handler.go:60`) | `CreateChannel` принимает любой `project_id` после parse; **нет** project membership / `PermProjectView` | FK к Project пропускает чужой UUID; канал попадает в `FindChannelInProject` / automation `channel_id` scope чужого проекта | До SetProjectID: `IsProjectMember` или RBAC `PermProjectView`/`PermTaskEdit`; иначе 403; Ent-тест foreign project |
| AUDIT-TM-005 | P3 | caps | `tracker/repo_links.go:73-96` | `ListIssueLinks` — два `.All()` без `Limit` | Comments/time/activity/board уже capped; issue с тысячами links → OOM/latency на Get links | `Limit(MaxIssueLinks)` (напр. 200) на outgoing+incoming или после merge; тест cap |
| AUDIT-TM-006 | P3 | AppError | `tracker/service_time.go:26-28` | `LogTime`: любой repo err → `NewInternal`; `repo.LogTime` отдаёт `NewNotFound` при soft-deleted issue | Race delete-after-FindIssueByKey → клиент видит 500 вместо 404; StartTimer/StopTimer пробрасывают AppError | `errors.As(*AppError)` passthrough; иначе Internal |
| AUDIT-TM-007 | P3 | soft-delete | `messenger/ent_repository.go:264-338` | `ToggleStar` / `MuteChannel` / `UnmuteChannel` query ChannelMember **без** `HasChannelWith(DeletedAtIsNil)` | `IsMember`/`GetMemberRole` fail-closed на trash; star/mute всё ещё мутируют membership soft-deleted канала | Добавить тот же predicate, что у IsMember; либо service IsMember до repo |
| AUDIT-TM-008 | P3 | AppError | `messenger/service_messages.go:255-257` | `ForwardMessage`: `if err != nil \|\| !isMember` → единый `NewForbidden` | DB err на IsMember маскируется под 403 (зеркало JoinChannel/AcceptInvite, где err → 500) | Разделить: `err != nil` → Internal; `!isMember` → Forbidden |

## Сверка с A3 (кратко)

| A3 ID | Статус на HEAD `617af223` |
|-------|---------------------------|
| A3W1-2-001…006 (labels/links/comments/automation channel) | FIXED-OK |
| A3W1-A3-001 FilterSubject Nak storm | FIXED-OK (`AutomationFilterSubjects`) |
| A3W1-A3-002 executeRule swallow | FIXED-OK (err return) |
| A3W1-A3-003 StartTimer soft-delete | FIXED-OK (`DeletedAtIsNil` на lock) |
| A3W3-2-001…006 (Join/Create/Bookmark/ClosePoll/AppError) | FIXED-OK |
| A3W3-3-001 ListUserChannels/WithMembers caps | FIXED-OK |
| A3W3-3-002 ListMembers soft-delete | FIXED-OK |
| A3W3-3-003 presence godoc | FIXED-OK |

**REGRESS известных A3 closes: 0.** Находки ниже — **NEW residual** относительно закрытого A3.

## Чисто (подтверждено по фокусу)

### IDOR / authz
- Tracker: все HTTP-мутации под `/projects/:projectId` + `rbac.RequirePermission`; FindIssueByKey / FindIssueInProject / workflow/CF/label/board/link/message-link scoped.
- Messenger: GetChannel/ListMessages/Edit/Delete/Reaction/Pin/Poll/Bookmark/Invite — membership; Forward source+target; Promote fail-closed; Search empty-scope; Presence FilterUsersSharingChannel.

### Soft-delete
- Issue/Comment/Label/CF/Board/Channel/Message(list) — `DeletedAtIsNil` / `deleted_at: nil` на основных list/get путях.
- Workflow status Delete — soft; StartTimer/LogTime/StopTimer — lock/AddTimeSpent с DeletedAtIsNil.

### Caps
- Issues list/JQL ≤100; comments/time/activity/board column/project links; messenger listMembers/pins/invites/bookmarks/votes/userChannels; queryIntMax=100; presence ≤50.

### AppError / layering
- Handlers не держат `*ent.Client`; service → Repository interface.
- ErrorHandler отдаёт `{"error","request_id"}`; сырой `fmt.Errorf` → generic 500 (без утечки деталей).
- Большинство service-путей wrap → AppError; исключения зафиксированы в AUDIT-TM-006/008.

## Сводка counts

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 4 |
| P3 | 4 |
| **Total** | **8** |

**Топ:** AUDIT-TM-003 (FindMessageByID trash) · AUDIT-TM-004 (CreateChannel project_id IDOR) · AUDIT-TM-001/002 (workflow soft-delete gaps).

## Операционные правила

- Режим: **READ-ONLY** по продуктовому коду.
- Коммит **не** создавался.
- **TASK-008 не закрывать** (нужны backlog TASK + независимая mutation verification).

*Артефакт AUDIT-BE-tracker-messenger-2026-07-29 · TASK-008 · readonly · RU · 2026-07-29*
