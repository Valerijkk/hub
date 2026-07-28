# W1 агент A3 — tracker repos + automation

> TASK-008 · стадия A · READ-ONLY · 2026-07-28  
> Перепроверка фиксов из `.planning/audits/grok-2026-07/03-tracking.md` на текущем коде.

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/tracker/repository.go` | 284 | 0 | чисто — ISP-интерфейсы; A3-002/A3-011 scopes в контракте |
| `backend/internal/tracker/repo_activity.go` | 48 | 1 | ListActivities без Limit |
| `backend/internal/tracker/repo_automation.go` | 114 | 1 | Update без projectID в Where |
| `backend/internal/tracker/repo_boards.go` | 122 | 0 | чисто — project scope + SoftDelete + MaxBoardIssuesPerColumn |
| `backend/internal/tracker/repo_comments.go` | 179 | 2 | soft-delete OK; `_=` Publish; List без Limit |
| `backend/internal/tracker/repo_custom_fields.go` | 42 | 1 | DeleteCustomField без project scope (IDOR) |
| `backend/internal/tracker/repo_issue_message_link.go` | 68 | 0 | чисто — Delete scoped по issueID |
| `backend/internal/tracker/repo_issues.go` | 443 | 3 | VersionEQ OK; Move без lock; Update без DeletedAtIsNil; `_=` Publish |
| `backend/internal/tracker/repo_labels.go` | 70 | 0 | чисто — DeleteLabel project-scoped (Block F) |
| `backend/internal/tracker/repo_links.go` | 136 | 1 | Delete scoped; `_=` Publish; batch cap OK |
| `backend/internal/tracker/repo_time.go` | 161 | 1 | A3-005 fixed; LogTime/StopTimer не атомарны с AddTimeSpent |
| `backend/internal/tracker/repo_workflow.go` | 182 | 2 | Update/Delete status + Create/Delete transition без project |
| `backend/internal/tracker/automation_consumer.go` | 131 | 1 | Nak на bad JSON; Ack при ошибках engine |
| `backend/internal/tracker/automation_dto.go` | 82 | 0 | чисто — DTO/константы |
| `backend/internal/tracker/automation_engine.go` | 138 | 1 | ошибки list/execute только slog, без return |
| `backend/testframework/tracker/automation_ack_a3006_test.go` | 67 | 0 | чисто — Nak/Ack mutation-ready |
| `backend/testframework/tracker/automation_consumer_test.go` | 132 | 1 | processMessage_Events без assert на Ack/Nak |
| `backend/testframework/tracker/automation_engine_events_test.go` | 106 | 0 | чисто — Ent + stubs |
| `backend/testframework/tracker/automation_engine_test.go` | 95 | 0 | чисто — post channel + max rules |
| `backend/testframework/tracker/automation_engine_wave100_test.go` | 250 | 1 | несколько тестов без assert |
| `backend/testframework/tracker/automation_service_extra_test.go` | 81 | 0 | чисто — validation / not found |
| `backend/testframework/tracker/board_label_idor_blockf_test.go` | 57 | 0 | чисто — Ent IDOR board/label |
| `backend/testframework/tracker/link_scope_a3011_postgres_ent_test.go` | 45 | 0 | чисто — Postgres twin DeleteIssueLink |
| `backend/testframework/tracker/link_scope_a3011_test.go` | 45 | 0 | чисто — sqlite Ent DeleteIssueLink |
| `backend/testframework/tracker/link_target_scope_a3011_postgres_ent_test.go` | 52 | 0 | чисто — FindIssueInProject PG |
| `backend/testframework/tracker/link_target_scope_a3011_test.go` | 93 | 0 | чисто — FindIssueInProject + service cross-project |
| `backend/testframework/tracker/postgres_ent_helper_test.go` | 108 | 0 | чисто — helper Skip без PG |
| `backend/testframework/tracker/reaction_scope_a3011_postgres_ent_test.go` | 39 | 0 | чисто — ToggleReaction PG |
| `backend/testframework/tracker/reaction_scope_a3011_test.go` | 39 | 0 | чисто — ToggleReaction sqlite |
| `backend/testframework/tracker/repository_board_cap_a13001_test.go` | 51 | 0 | чисто — MaxBoardIssuesPerColumn |
| `backend/testframework/tracker/repository_extended_test.go` | 127 | 0 | чисто — labels/links/time/seed |
| `backend/testframework/tracker/repository_more_test.go` | 205 | 0 | чисто — comments/CF/timer/board filters |
| `backend/testframework/tracker/repository_project_links_a13004_test.go` | 58 | 0 | чисто — ListProjectIssueLinks batch |
| `backend/testframework/tracker/repository_test.go` | 177 | 0 | чисто — A3-002 + VersionEQ + CRUD |
| `backend/testframework/tracker/repository_wave100_test.go` | 238 | 0 | чисто — outbox/board filters/automation handler |
| `backend/testframework/tracker/repository_wave120_test.go` | 59 | 0 | чисто — comment outbox + delete rule |
| `backend/testframework/tracker/repository_wave124_test.go` | 137 | 0 | чисто — optional fields + list filters/cursor |

Файлов в списке: 37. Прочитано: 37. Пропущено: 0.

## Перепроверка grok-2026-07 (релевантные A3)

| ID | Было | Текущий код | Статус |
|----|------|-------------|--------|
| A3-002 | FindIssueByKey без project | `repo_issues.go:147-163` + `HasProjectWith`; тест `repository_test.go` RejectsForeignProject | **FIXED** |
| A3-005 | LogTime без EndedAt → panic StopTimer | `repo_time.go:22-29` SetStartedAt+SetEndedAt; StopTimer `StartedAtNotNil`; StartTimer conflict+tx lock | **FIXED** |
| A3-006 | всегда `_ = msg.Ack()` | bad JSON → `Nak` (`automation_consumer.go:113-117`); успех → Ack; engine-ошибки всё ещё Ack | **PARTIAL** |
| A3-011 | comment/react/link без scope | Update/Delete comment + ToggleReaction + DeleteIssueLink scoped; postgres twins | **FIXED** |
| A3-013 | MoveIssue без VersionEQ | `repo_issues.go:416-419` UpdateOneID без VersionEQ | **OPEN** |
| A3-021 | `_ = publisher.Publish` | comments/issues/links/move — по-прежнему | **OPEN** |

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|----------|-------------------|------------|
| W1-A3-001 | P1 | `repo_custom_fields.go:40-41` | `DeleteCustomField(ctx, fieldID)` без `projectID` | Handler RBAC смотрит membership в `:projectId`, но удаление идёт только по UUID поля → member проекта A с `PermProjectManageWorkflow` удаляет поле проекта B | Как `DeleteLabel`: `Delete().Where(IDEQ, HasProjectWith(projectID))`; affected==0 → NotFound; прокинуть `projectID` через service/handler |
| W1-A3-002 | P1 | `repo_workflow.go:39-50`, `54-55` | `UpdateWorkflowStatus` / `DeleteWorkflowStatus` только по `statusID` | Тот же класс IDOR: `:projectId` в URL не участвует в мутации статуса чужого проекта | Where/Update с `HasProjectWith(projectID)`; signature + service/handler |
| W1-A3-003 | P1 | `repo_workflow.go:82-86`, `90-91` | `CreateWorkflowTransition` / `DeleteWorkflowTransition` без проверки, что статусы/переход ∈ project URL | Можно связать/удалить статусы чужого проекта, зная UUID (handler даже не парсит `projectId` на create) | Перед create: оба статуса `HasProjectWith(projectID)`; delete transition через from_status ∈ project |
| W1-A3-004 | P2 | `repo_issues.go:416-419` | `MoveIssue` — `UpdateOneID` без `VersionEQ` / `AddVersion` (A3-013) | Два concurrent drag-and-drop → last-write-wins; WIP TOCTOU | Как UpdateIssue: `Where(ID, VersionEQ).AddVersion(1)`; conflict → 409 |
| W1-A3-005 | P2 | `repo_issues.go:316-320` | `UpdateIssue` фильтрует `IDEQ`+`VersionEQ`, **без** `DeletedAtIsNil` | Зная ID+version, можно править задачу в корзине (soft-deleted) | Добавить `issue.DeletedAtIsNil()` в Where update |
| W1-A3-006 | P2 | `repo_time.go:35-46`, `135-148` | `LogTime`/`StopTimer`: запись TimeLog и `AddTimeSpentMinutes` не в одной tx | При ошибке второго шага — лог есть, сумма на issue нет (или наоборот при StopTimer) | Одна Ent tx: update TimeLog + Issue; Rollback при ошибке |
| W1-A3-007 | P2 | `automation_consumer.go:94-121`; `automation_engine.go:31-34` | Handle* не возвращают error; consumer Ack после вызова даже если `ListEnabledAutomationRules` упал (остаток A3-006) | Transient DB/NATS side-effect failure → событие потеряно навсегда; MaxDeliver бесполезен | Engine → `error`; consumer: err → Nak/NakWithDelay; только успех → Ack |
| W1-A3-008 | P3 | `repo_issues.go:376`, `432`; `repo_comments.go:47`; `repo_links.go:58` | `_ = r.publisher.Publish(...)` глотает ошибки outbox (A3-021) | Клиенты/automation не получают события без сигнала | slog+metric; критичные пути — fail или PublishInTx |
| W1-A3-009 | P3 | `repo_automation.go:55-69` | `UpdateAutomationRule(ruleID)` без `ProjectIDEQ` | HTTP смягчён `FindAutomationRule(projectID, ruleID)`, но repo позволяет UpdateOneID любого rule; TOCTOU | `Update().Where(IDEQ, ProjectIDEQ)` как у Delete |
| W1-A3-010 | P3 | `repo_activity.go:42-47`; `repo_comments.go:68-79`; `repo_time.go:155-160` | ListActivities / ListComments / ListTimeLogs без Limit | Неограниченная выборка на горячих issue с длинной историей | Limit/cursor (как ListIssues) |
| W1-A3-011 | P3 | `automation_consumer_test.go:87-131`; `automation_engine_wave100_test.go:154-180`, `216-249`, `27-32` | Тесты вызывают consumer/engine без assert на исход | Assert-free: пройдут и на сломанном Ack/Nak/noop | Assert Ack/Nak counts; assert «no panic + no side effect» явно |

## Чисто

- **A3-002 / FindIssueByKey + FindIssueInProject** — project scope + soft-delete; regression в `repository_test.go` и link_target/reaction twins (sqlite + Postgres).
- **A3-005 timer** — LogTime закрыт EndedAt; StartTimer reject duplicate + row-lock tx; StopTimer `StartedAtNotNil`; покрыто `repository_more_test.go` + отдельный `log_time_a3005_test.go` (вне списка, подтверждение фикса).
- **A3-006 (часть)** — bad JSON → Nak; тест `automation_ack_a3006_test.go` mutation-ready.
- **Comments A3-011 / A8-006** — Update/Delete scoped author+issue; soft `DeletedAt`; ToggleReaction scoped; List фильтрует trash.
- **Links A3-011** — DeleteIssueLink `HasSourceWith(sourceIssueID)`; Postgres twin.
- **Boards / Labels Block F** — GetBoardData + DeleteLabel project-scoped; cap 200/колонка (A13-001).
- **ListProjectIssueLinks** — batch + Limit 2000 (A13-004).
- **UpdateIssue** — `VersionEQ` + AddVersion + 409 conflict — образец optimistic lock.
- **CreateIssue / SeedDefaultWorkflow** — tx + Rollback.
- **DeleteIssue** — soft delete `SetDeletedAt`.
- **DeleteAutomationRule / FindAutomationRule / List*** — project scoped.
- **automation_dto** — только DTO/константы, без I/O.
- **Postgres Ent helpers** — Skip без `POSTGRES_TEST_URL`, не Fatal; twins бьют в реальный guard.

### Сводка counts

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 3 |
| P2 | 4 |
| P3 | 4 |
| **Total** | **11** |

**Топ-3:** W1-A3-001 (DeleteCustomField IDOR) · W1-A3-002 (WorkflowStatus IDOR) · W1-A3-003 (WorkflowTransition cross-project).
