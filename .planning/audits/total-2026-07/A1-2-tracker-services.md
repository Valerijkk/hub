# W1 агент 2 — tracker services

**Дата:** 2026-07-28  
**Режим:** READ-ONLY (кроме этого отчёта)  
**Задача:** TASK-008 / волна W1 / агент A2  
**Перепроверка:** `.planning/audits/grok-2026-07/03-tracking.md` (A3-002/005/011 закрыты в сервисе; A3-013/018 всё ещё открыты)

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/tracker/service.go` | 97 | 0 | чисто |
| `backend/internal/tracker/service_activity.go` | 35 | 0 | чисто (FindIssueByKey + projectID) |
| `backend/internal/tracker/service_boards.go` | 127 | 0 | чисто (GetBoardData scoped) |
| `backend/internal/tracker/service_comments.go` | 131 | 0 | чисто (A3-011 закрыт: issue+author) |
| `backend/internal/tracker/service_converters.go` | 319 | 0 | чисто |
| `backend/internal/tracker/service_custom_fields.go` | 56 | 1 | IDOR Delete без projectID |
| `backend/internal/tracker/service_issue_message_link.go` | 219 | 2 | channel без scope; silent mirror errors |
| `backend/internal/tracker/service_issues.go` | 344 | 4 | unscoped refs; Move без version; HasMore; raw err |
| `backend/internal/tracker/service_labels.go` | 84 | 0 | чисто (DeleteLabel + projectID) |
| `backend/internal/tracker/service_links.go` | 125 | 0 | чисто (FindIssueInProject / Delete scoped) |
| `backend/internal/tracker/service_time.go` | 99 | 0 | чисто (A3-005 закрыт на repo; service pass-through) |
| `backend/internal/tracker/service_workflow.go` | 114 | 3 | IDOR status/transition CRUD без project |
| `backend/internal/tracker/automation_service.go` | 167 | 0 | чисто (Find/Delete scoped; Create validates) |
| `backend/testframework/tracker/service_converters_extended_test.go` | 175 | 0 | чисто |
| `backend/testframework/tracker/service_converters_test.go` | 204 | 0 | чисто |
| `backend/testframework/tracker/service_helpers_test.go` | 95 | 0 | чисто |
| `backend/testframework/tracker/service_issue_message_link_test.go` | 145 | 0 | чисто (project mismatch) |
| `backend/testframework/tracker/service_test.go` | 1307 | 1 | mock StartTimer/DeleteCustomField без guard; HasMore не assert |
| `backend/testframework/tracker/service_wave100_test.go` | 111 | 0 | чисто |
| `backend/testframework/tracker/service_wave109_test.go` | 79 | 0 | чисто |
| `backend/testframework/tracker/service_wave112_test.go` | 201 | 0 | чисто |
| `backend/testframework/tracker/service_wave113_test.go` | 74 | 1 | TestDeleteComment_NotFound мимо DeleteComment |
| `backend/testframework/tracker/service_wave114_test.go` | 126 | 0 | чисто (error paths UUID) |
| `backend/testframework/tracker/service_wave115_test.go` | 80 | 0 | чисто |
| `backend/testframework/tracker/service_wave117_test.go` | 62 | 0 | чисто |
| `backend/testframework/tracker/service_wave118_test.go` | 87 | 0 | чисто (sqlite Ent) |
| `backend/testframework/tracker/service_wave119_test.go` | 69 | 0 | чисто |
| `backend/testframework/tracker/service_wave121_test.go` | 49 | 0 | чисто |
| `backend/testframework/tracker/service_workflow_extra_test.go` | 207 | 0 | чисто (happy-path CRUD) |
| `backend/testframework/tracker/comment_scope_a3011_test.go` | 58 | 0 | чисто (Ent guard + mutation hint) |
| `backend/testframework/tracker/comment_soft_delete_a8006_test.go` | 96 | 0 | чисто (soft-delete + source assert) |
| `backend/testframework/tracker/log_time_a3005_test.go` | 90 | 0 | чисто (A3-005 Ent) |
| `backend/testframework/tracker/start_timer_concurrent_a3005_test.go` | 72 | 0 | чисто (Postgres race) |

Файлов в списке: 33. Прочитано: 33. Пропущено: 0.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|----------|-------------------|------------|
| W1-A2-001 | P1 | `service_custom_fields.go:48-49` (+ handler не передаёт `projectId`) | `DeleteCustomField(ctx, fieldID)` без `projectID`; repo `DeleteOneID` | Member проекта A с `PermProjectManageWorkflow` удаляет кастомное поле проекта B по UUID (IDOR). Контраст с `DeleteLabel(ctx, projectID, labelID)` | Сигнатура `DeleteCustomField(ctx, projectID, fieldID)`; repo `Where(ID, HasProjectWith(projectID))` → 404 |
| W1-A2-002 | P1 | `service_workflow.go:41-42`, `53-54` | `UpdateWorkflowStatus` / `DeleteWorkflowStatus` только по `statusID` | RBAC смотрит membership в `:projectId` URL; статус чужого проекта обновляется/удаляется по UUID | `Update/Delete...(ctx, projectID, statusID, …)`; repo `Where(ID, HasProjectWith(projectID))` |
| W1-A2-003 | P1 | `service_workflow.go:81-91` | `CreateWorkflowTransition` без `projectID`; статусы не проверяются на принадлежность проекту | Можно связать статусы разных проектов; далее Move/UpdateIssue проходят `ValidateTransition` на чужой `to_status` | Принимать `projectID`; оба статуса `HasProjectWith(projectID)` и один проект; иначе 400/404 |
| W1-A2-004 | P1 | `service_workflow.go:106-107` | `DeleteWorkflowTransition` только по `transitionID` | Удаление перехода чужого проекта по UUID при RBAC своего `:projectId` | Scope delete: transition ∈ project (через FromStatus.Project) |
| W1-A2-005 | P1 | `service_issues.go:58-74`, `51-56` | `CreateIssue`: `StatusID` / `ParentID` / `Labels` парсятся, но не проверяются на ∈ `projectID` | На create нет `ValidateTransition`; FK не гарантирует `status.project_id == issue.project_id` → задача проекта A со статусом/родителем/метками проекта B | Перед create: `FindStatusInProject` / `FindIssueInProject` / labels `Where(HasProject…)`; mismatch → 400/404 |
| W1-A2-006 | P1 | `service_issue_message_link.go:79-93` | `CreateIssueMessageLink` валидирует issue∈project, но не `channel_id` ∈ project / membership | Связь задачи с чужим каналом; mirror комментариев утекает в чужой тред | Резолв channel + `channel.project_id == projectID` (и membership); иначе 404 |
| W1-A2-007 | P2 | `service_issues.go:330` (A3-013 **открыт**) | `MoveIssue` → `repo.MoveIssue` без `version` / `VersionEQ` | Гонка двух drag-and-drop: last-write-wins; WIP check TOCTOU до update | `MoveIssueRequest.Version` + `VersionEQ`/`AddVersion`; WIP в той же tx |
| W1-A2-008 | P2 | `service_issues.go:182` (A3-018 **открыт**) | `HasMore: len(items) == limit` | Последняя страница ровно `limit` элементов → ложный `HasMore` | Fetch `limit+1` или `len(items)==limit && offset+limit < total` / cursor |
| W1-A2-009 | P2 | `service_issues.go:251-256`, `308-320` | `UpdateIssue`/`MoveIssue`: `newStatusID` не проверяется на ∈ project после `ValidateTransition` | Вместе с W1-A2-003: переход на статус чужого проекта корректен для workflow engine | После parse: status `HasProjectWith(projectID)`; иначе 404 |
| W1-A2-010 | P2 | `service_issues.go:77-79` | `CreateIssue` возвращает сырой `err` из repo | Ent/FK/constraint тексты могут уйти клиенту вместо `{"error","request_id"}` AppError | Маппинг: NotFound/Constraint → Conflict/BadRequest; остальное → Internal |
| W1-A2-011 | P2 | `service_wave113_test.go:32-38` | `TestDeleteComment_NotFound` зовёт `DeleteComment` без issue в mock → падает на `FindIssueByKey` | Override `DeleteComment` → NotFound **никогда не вызывается**; assert `"не найден"` зелёный и на «задача не найдена». Сломанный маппинг ошибки DeleteComment тест не ловит | Создать issue+comment; stub только DeleteComment; assert `комментарий не найден` / 404 code |
| W1-A2-012 | P2 | `service_test.go:500-514`, `657-659` | `mockRepo.StartTimer` не отклоняет второй open; `DeleteCustomField` без project | Service-тесты зелёные при регрессии 409 open-timer на service path и при отсутствии project-scope delete (guard живёт только в Ent-тестах A3-005 / нет IDOR-теста поля) | В mock: StartTimer → 409 если open; DeleteCustomField(projectID, id); отдельные denial-тесты на Ent |
| W1-A2-013 | P3 | `service_issue_message_link.go:207-216` | `mirrorIssueCommentsToThreads`: `err` list/mirror → `return`/`continue` без лога | Сбой зеркалирования невидим в ops; треды расходятся с комментариями | `slog.Warn` + metric; не глотать |
| W1-A2-014 | P3 | `service_issues.go:287-295` | `DeleteIssue(..., userID)` — `userID` не используется; нет `LogActivity` | Нет аудита «кто удалил»; параметр мёртвый | `LogActivity(..., "deleted", …)` до soft-delete; или убрать unused |
| W1-A2-015 | P3 | `service_workflow.go:71-74` | `ListWorkflowTransitions`: `t.Edges.FromStatus.ID` / `ToStatus.ID` без nil-check | При отсутствии eager edges — nil deref panic | Guard nil → skip/500; repo уже `WithFromStatus/WithToStatus` — защитный слой |

## Чисто

- **A3-002 закрыт:** все issue lookups в сервисе через `FindIssueByKey(ctx, projectID, key)` (activity/comments/time/links/issues/message-link).
- **A3-011 закрыт:** `UpdateComment`/`DeleteComment`/`ToggleReaction` передают `issueID`+`userID` (author+issue scope).
- **A3-005 закрыт на repo:** Ent-тесты `log_time_a3005` / `start_timer_concurrent_a3005`; service только pass-through + comment про 409.
- **A3-011-target / link:** `FindIssueInProject`, `DeleteIssueLink(source, linkID)`.
- **Block F:** `DeleteLabel(projectID, labelID)`, `GetBoardData(projectID, boardID)`.
- **Optimistic lock UpdateIssue:** version через repo `VersionEQ` (покрыто `TestUpdateIssueOptimisticLocking`).
- **DeleteIssue:** soft-delete в repo (`SetDeletedAt`) — service делегирует.
- **AutomationService:** Create validates trigger/action; max rules; Find/Delete rule scoped by projectID.
- **Регрессионные Ent-тесты:** `comment_scope_a3011`, `comment_soft_delete_a8006` — mutation-ready, не mock-only.

### Сводка

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 6 |
| P2 | 6 |
| P3 | 3 |
| **Total** | **15** |

**Топ-3:** W1-A2-001 (DeleteCustomField IDOR) · W1-A2-002/003/004 (workflow status/transition IDOR) · W1-A2-005 (CreateIssue unscoped Status/Parent/Labels).
