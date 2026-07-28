# A3 W1 агент 1 — tracker handlers (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`)  
> Baseline: [A2W1-1-tracker-handlers.md](A2W1-1-tracker-handlers.md)  
> Периметр: 36 файлов (16 prod + 20 tests) · Прочитано: 36 · Пропущено: 0

## Сверка с A2 (W1-A1-*)

| A1/A2 ID | Tag | Статус в HEAD | Комментарий |
|----------|-----|---------------|-------------|
| W1-A1-001 | FIXED-OK | закрыт | `DeleteCustomField(ctx, projectID, fieldID)` — `handler_custom_fields.go:83-97` |
| W1-A1-002 | FIXED-OK | закрыт | `UpdateWorkflowStatus(ctx, projectID, statusID, …)` — `handler_workflow.go:86-106` |
| W1-A1-003 | FIXED-OK | закрыт | `DeleteWorkflowStatus(ctx, projectID, statusID)` — `handler_workflow.go:121-135` |
| W1-A1-004 | FIXED-OK | закрыт | `CreateWorkflowTransition` парсит `projectID` — `handler_workflow.go:177-193` |
| W1-A1-005 | FIXED-OK | закрыт | `DeleteWorkflowTransition` scoped — `handler_workflow.go:208-222` |
| W1-A1-006 | FIXED-OK | закрыт | `boards.Post` → `PermTaskEdit` — `handler.go:74` |
| W1-A1-007 | FIXED-OK | закрыт на handler | DTO `validate` теги; Bind → StructValidator в wiring (вне слота); handlers без обхода |
| W1-A1-008 | FIXED-OK | закрыт | JQL → `service.SearchIssuesJQL` — `handler_issues.go:277-282` |
| W1-A1-009 | FIXED-OK | закрыт | Ent-negative / IDOR suite вне mock-only (A2); регрессии нет |
| W1-A1-010 | FIXED-OK | закрыт | foreign-project delete CF закрыт на handler+service scope |
| W1-A1-011 | FIXED-OK | закрыт | Ent-negative authz (TASK-037); mock happy-path остаётся доп. покрытием |
| W1-A1-012 | FIXED-OK | закрыт на prod | validator + struct-validator test; wave119 mock-env без StructValidator — не регресс prod |
| A2 активные | FIXED-OK | 0 открытых | A2 слот был чист; активных A2-находок не было |

**REGRESS:** 0 · **RESIDUAL (этот слот):** 0 · **NEW (этот слот):** 0

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| backend/internal/tracker/handler.go | 116 | 0 | чисто |
| backend/internal/tracker/handler_automation.go | 126 | 0 | чисто |
| backend/internal/tracker/handler_boards.go | 101 | 0 | чисто |
| backend/internal/tracker/handler_comments.go | 246 | 0 | чисто |
| backend/internal/tracker/handler_custom_fields.go | 98 | 0 | чисто |
| backend/internal/tracker/handler_issue_message_link.go | 118 | 0 | чисто |
| backend/internal/tracker/handler_issues.go | 283 | 0 | чисто |
| backend/internal/tracker/handler_labels.go | 156 | 0 | чисто |
| backend/internal/tracker/handler_links.go | 137 | 0 | чисто |
| backend/internal/tracker/handler_time.go | 153 | 0 | чисто |
| backend/internal/tracker/handler_workflow.go | 223 | 0 | чисто |
| backend/internal/tracker/dto.go | 363 | 0 | чисто |
| backend/internal/tracker/events.go | 84 | 0 | чисто |
| backend/internal/tracker/export_hooks.go | 146 | 0 | чисто |
| backend/internal/tracker/provisioner.go | 29 | 0 | чисто |
| backend/internal/tracker/workflow.go | 29 | 0 | чисто |
| backend/testframework/tracker/handler_automation_flow_test.go | 124 | 0 | чисто |
| backend/testframework/tracker/handler_flow_test.go | 320 | 0 | чисто |
| backend/testframework/tracker/handler_jql_flow_test.go | 52 | 0 | чисто |
| backend/testframework/tracker/handler_rbac_test.go | 132 | 0 | чисто |
| backend/testframework/tracker/handler_routes_test.go | 156 | 0 | чисто |
| backend/testframework/tracker/handler_wave103_test.go | 121 | 0 | чисто |
| backend/testframework/tracker/handler_wave106_test.go | 138 | 0 | чисто |
| backend/testframework/tracker/handler_wave109_test.go | 96 | 0 | чисто |
| backend/testframework/tracker/handler_wave112_test.go | 120 | 0 | чисто |
| backend/testframework/tracker/handler_wave114_test.go | 65 | 0 | чисто |
| backend/testframework/tracker/handler_wave115_test.go | 45 | 0 | чисто |
| backend/testframework/tracker/handler_wave117_test.go | 50 | 0 | чисто |
| backend/testframework/tracker/handler_wave118_test.go | 110 | 0 | чисто |
| backend/testframework/tracker/handler_wave119_test.go | 178 | 0 | чисто |
| backend/testframework/tracker/handler_wave120_test.go | 220 | 0 | чисто |
| backend/testframework/tracker/handler_wave121_test.go | 96 | 0 | чисто |
| backend/testframework/tracker/handler_wave124_test.go | 25 | 0 | чисто |
| backend/testframework/tracker/handler_wave95_test.go | 141 | 0 | чисто |
| backend/testframework/tracker/handler_workflow_labels_test.go | 231 | 0 | чисто |
| backend/testframework/tracker/provisioner_test.go | 28 | 0 | чисто |

Файлов в списке: 36. Прочитано: 36. Пропущено: 0.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|----------|-------------------|------------|
| — | — | — | — | Нет активных находок в handler-слоте | A2 closes подтверждены; регрессий IDOR/perm/JQL/layering нет | — |

## Чисто

- RBAC Fiber v3 order сохранён; CreateBoard = `PermTaskEdit`; mutations с корректными Perm*.
- Workflow/CF/labels/links/comments/time/automation/message-links — `ParseProjectID` + scoped service calls; нет `_ =` / логов секретов / Ent в handlers.
- JQL только через `SearchIssuesJQL`; soft-delete issue делегирован service; optimistic `version` в DTO UpdateIssue/UpdateComment.
- `h.automation == nil` → 501; пустой issueKey → 400 на comments/issues/time/links.
- Wave/mock тесты — happy-path + invalid JSON/UUID; IDOR закрыт вне mock (`handler_rbac_test`, Ent-negative suites из A2).

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 0 |
| P3 | 0 |
| **Total** | **0** |

A2 baseline: 0 active → FIXED-OK 0+12 (A1 closes) · REGRESS 0 · RESIDUAL 0 · NEW 0.
