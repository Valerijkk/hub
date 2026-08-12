# W1 агент 1 — tracker handlers

## Прочитанные файлы
| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| backend/internal/tracker/handler.go | 115 | 1 | замечания |
| backend/internal/tracker/handler_automation.go | 126 | 0 | чисто |
| backend/internal/tracker/handler_boards.go | 101 | 0 | чисто |
| backend/internal/tracker/handler_comments.go | 246 | 0 | чисто |
| backend/internal/tracker/handler_custom_fields.go | 94 | 1 | замечания |
| backend/internal/tracker/handler_issue_message_link.go | 118 | 0 | чисто |
| backend/internal/tracker/handler_issues.go | 358 | 1 | замечания |
| backend/internal/tracker/handler_labels.go | 156 | 0 | чисто |
| backend/internal/tracker/handler_links.go | 137 | 0 | чисто |
| backend/internal/tracker/handler_time.go | 153 | 0 | чисто |
| backend/internal/tracker/handler_workflow.go | 206 | 4 | замечания |
| backend/internal/tracker/dto.go | 339 | 1 | замечания |
| backend/internal/tracker/events.go | 84 | 0 | чисто |
| backend/internal/tracker/export_hooks.go | 141 | 0 | чисто |
| backend/internal/tracker/provisioner.go | 29 | 0 | чисто |
| backend/internal/tracker/workflow.go | 29 | 0 | чисто |
| backend/testframework/tracker/handler_automation_flow_test.go | 124 | 0 | чисто |
| backend/testframework/tracker/handler_flow_test.go | 320 | 1 | замечания |
| backend/testframework/tracker/handler_jql_flow_test.go | 52 | 0 | чисто |
| backend/testframework/tracker/handler_rbac_test.go | 132 | 0 | чисто |
| backend/testframework/tracker/handler_routes_test.go | 156 | 0 | чисто |
| backend/testframework/tracker/handler_wave103_test.go | 121 | 0 | чисто |
| backend/testframework/tracker/handler_wave106_test.go | 138 | 1 | замечания |
| backend/testframework/tracker/handler_wave109_test.go | 96 | 0 | чисто |
| backend/testframework/tracker/handler_wave112_test.go | 120 | 0 | чисто |
| backend/testframework/tracker/handler_wave114_test.go | 65 | 0 | чисто |
| backend/testframework/tracker/handler_wave115_test.go | 45 | 0 | чисто |
| backend/testframework/tracker/handler_wave117_test.go | 50 | 0 | чисто |
| backend/testframework/tracker/handler_wave118_test.go | 110 | 0 | чисто |
| backend/testframework/tracker/handler_wave119_test.go | 178 | 1 | замечания |
| backend/testframework/tracker/handler_wave120_test.go | 220 | 0 | чисто |
| backend/testframework/tracker/handler_wave121_test.go | 96 | 0 | чисто |
| backend/testframework/tracker/handler_wave124_test.go | 25 | 0 | чисто |
| backend/testframework/tracker/handler_wave95_test.go | 141 | 1 | замечания |
| backend/testframework/tracker/handler_workflow_labels_test.go | 231 | 0 | чисто |
| backend/testframework/tracker/provisioner_test.go | 28 | 0 | чисто |

Файлов в списке: 36. Прочитано: 36. Пропущено: 0.

## Перепроверка прошлого аудита (A3 / grok-2026-07)

| Старый ID | Статус в текущем коде | Комментарий |
|-----------|----------------------|-------------|
| A3-002 | **исправлено** — не дублировать | Handlers передают `projectID`; `FindIssueByKey` scoped (`HasProjectWith`) |
| A3-011 | **исправлено** — не дублировать | `UpdateComment`/`DeleteComment` передают `projectID`, `issueKey`, `userID` |
| A3-022 | **живо** | JQL всё ещё `h.service.repo.GetDB()` → W1-A1-008 |
| A3-005 / A3-006 / A3-013 / A3-018 / A3-021 | вне scope файлов A1 | repo/service/automation — другие агенты |

## Находки
| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|----------|-------------------|------------|
| W1-A1-001 | P1 | `backend/internal/tracker/handler_custom_fields.go:83-93` | `DeleteCustomField` не парсит `:projectId` и вызывает `DeleteCustomField(ctx, fieldID)` | RBAC проверяет только membership в URL-проекте. Зная UUID поля чужого проекта, member с `PermProjectManageWorkflow` удаляет чужое определение (`DeleteOneID` без project scope в service/repo) | Парсить `projectID`; `DeleteCustomField(ctx, projectID, fieldID)` + `Where(ID, HasProjectWith(...))`; mismatch → 404 |
| W1-A1-002 | P1 | `backend/internal/tracker/handler_workflow.go:86-102` | `UpdateWorkflowStatus` игнорирует `:projectId`, мутирует по голому `statusID` | IDOR: правка статуса workflow чужого проекта при наличии manage-permission в своём | Передать `projectID`; update с `HasProjectWith(projectID)` |
| W1-A1-003 | P1 | `backend/internal/tracker/handler_workflow.go:117-127` | `DeleteWorkflowStatus` — только `statusID`, без project scope | Hard-delete статуса чужого проекта по UUID | Scope delete по проекту; 404 при mismatch |
| W1-A1-004 | P1 | `backend/internal/tracker/handler_workflow.go:170-180` | `CreateWorkflowTransition` не читает `projectID`; `CreateWorkflowTransition(ctx, req)` | Можно связать статусы другого проекта (или cross-project from/to), обойдя URL-scope | Парсить `projectID`; проверить, что `from_status_id` и `to_status_id` ∈ project; иначе 404/400 |
| W1-A1-005 | P1 | `backend/internal/tracker/handler_workflow.go:195-205` | `DeleteWorkflowTransition` — только `transitionID` | Удаление перехода чужого workflow по UUID | Delete scoped: transition принадлежит статусам проекта из URL |
| W1-A1-006 | P2 | `backend/internal/tracker/handler.go:73` | `boards.Post("/", h.CreateBoard, … PermProjectView …)` | Создание доски — мутация, но достаточно `PermProjectView` (в отличие от labels/`PermTaskEdit`, workflow/`PermProjectManageWorkflow`) | Требовать `PermProjectManageWorkflow` или отдельный manage-boards permission |
| W1-A1-007 | P2 | `backend/internal/tracker/dto.go:16-143` (+ handlers `Bind().JSON` без validator) | Теги `validate:"required,…"` на DTO не исполняются: в wiring нет StructValidator; handlers только `Bind().JSON` | Пустые/невалидные тела принимаются. Доказательство: `handler_wave119_test.go:40` шлёт `{"body":"…"}` без `body_text`/`body_json` и ждёт 201 | Включить validator на Bind (Fiber StructValidator) или явный `validate.Struct` в handlers; отсекать пустой `body_text` в service |
| W1-A1-008 | P3 | `backend/internal/tracker/handler_issues.go:285-311` | JQL: `h.service.repo.GetDB().Issue.Query()` из handler (бывш. A3-022) | Нарушение layering: handler знает Ent; обходит service API | `service.SearchJQL(ctx, projectID, userID, query)` |
| W1-A1-009 | P2 | `backend/testframework/tracker/handler_wave106_test.go:122-138` | `WorkflowStatusUpdate` кладёт статус в mock **без** project edge и ждёт 200 | Тест зелёный при живом IDOR (W1-A1-002); нет negative-кейса «чужой projectId + чужой statusId → 404» | Ent/sqlite fixture + cross-project denial; assert 404 |
| W1-A1-010 | P2 | `backend/testframework/tracker/handler_wave95_test.go:129-141` | `DeleteCustomField_Success` на mock без project ownership | Тест зелёный при W1-A1-001; guard должен жить в repo с project scope, а тест бьёт по mock | Реальный Ent + foreign-project delete → 404; не только happy-path на mock |
| W1-A1-011 | P2 | `backend/testframework/tracker/handler_flow_test.go:20-61` (+ wave103/wave95 mock setups) | Большой пласт handler-тестов на `mockRepo`/`newTestService()` | Project-scope / ownership guards в Ent-repo не проверяются; IDOR-класс (W1-A1-001…005) mock’ом не ловится | Для authz/IDOR — integration против Ent (как `handler_rbac_test` / `handler_routes_test`); mock — только parse/HTTP shape |
| W1-A1-012 | P2 | `backend/testframework/tracker/handler_wave119_test.go:40-47` | Create comment с `{"body":"Wave119 comment"}` (не контракт DTO) → 201 | Тест закрепляет сломанную валидацию (W1-A1-007): проходит на коде без required-check | Требовать `body_text`+`body_json`; ожидать 400 на старом теле |

## Чисто

Файлы без находок — что проверялось:

- **handler_automation.go** — nil-guard automation; ParseProjectID; UserID; UUID ruleId; projectID прокидывается в List/Create/Update/Delete; ошибки Bind → 400; нет `_ = err`.
- **handler_boards.go** — projectID + boardID; swimlane query; Create/List/GetBoardData scoped через service (`GetBoardData` + project).
- **handler_comments.go** — projectID+issueKey на всех CRUD/reactions/activity; userID на мутациях; пустой issueKey → 400. A3-011 закрыт на handler-слое.
- **handler_issue_message_link.go** — projectID; валидация channel_id/message_id; UUID linkId; Create с userID.
- **handler_labels.go** — DeleteLabel передаёт projectID+labelID; groups/list/create scoped.
- **handler_links.go** — Create/List/Delete с projectID+issueKey; batch `ListProjectIssueLinks` по project.
- **handler_time.go** — Log/Start/Stop/List с projectID+issueKey+userID; пустой key → 400.
- **events.go** — только event DTO, без логики/секретов.
- **export_hooks.go** — hubtest re-exports, без prod path.
- **provisioner.go** — тонкая делегация SeedDefaultWorkflow / CreateDefaultBoard.
- **workflow.go** — ValidateTransition через FindTransition; BadRequest на запрет; без panic.
- **handler_automation_flow_test.go** — CRUD automation + comment update/delete на sqlite; asserts есть.
- **handler_jql_flow_test.go** — success + empty query 400; asserts.
- **handler_rbac_test.go** — Fiber v3 RBAC order; outsider 403 (RegisterRoutes); member 201.
- **handler_routes_test.go** — create/delete issue soft path, custom fields, board, invalid JQL body.
- **handler_wave103/109/112/114/115/117/118/120/121/124** — в основном invalid JSON/UUID/projectId и happy-path; asserts присутствуют; assert-free нет.
- **handler_workflow_labels_test.go** — label/workflow/time/link flows на mock; покрытие позитива (не IDOR-negative).
- **provisioner_test.go** — Seed + Main Board; assert имени доски.

Дополнительно по «чистым» handler-путям issues (кроме JQL): Create/Get/Update/Delete/Move передают `projectID`+`issueKey` (A3-002 закрыт на этом слое); soft-delete issue через service; Version в UpdateIssueRequest присутствует в DTO.
