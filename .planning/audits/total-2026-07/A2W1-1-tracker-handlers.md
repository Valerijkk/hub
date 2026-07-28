# A2 W1 агент 1 — tracker handlers (re-audit)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> HEAD: `0d6a659fe9dd56a473cea0ace40ccae280970cbc` (`0d6a659f`)  
> Baseline: [A1-1-tracker-handlers.md](A1-1-tracker-handlers.md)  
> Периметр: 36 файлов (16 prod + 20 tests) · Прочитано: 36 · Пропущено: 0

## Сверка с A1 (W1-A1-*)

| A1 ID | Tag | Статус в HEAD | Комментарий |
|-------|-----|---------------|-------------|
| W1-A1-001 | FIXED-OK | закрыт | `DeleteCustomField` → `projectID` + service/repo scope; Ent `project_scope_idor_task013` |
| W1-A1-002 | FIXED-OK | закрыт | `UpdateWorkflowStatus(ctx, projectID, statusID, …)` |
| W1-A1-003 | FIXED-OK | закрыт | `DeleteWorkflowStatus` scoped |
| W1-A1-004 | FIXED-OK | закрыт | `CreateWorkflowTransition` парсит `projectID` |
| W1-A1-005 | FIXED-OK | закрыт | `DeleteWorkflowTransition` scoped |
| W1-A1-006 | FIXED-OK | закрыт | `boards.Post` → `PermTaskEdit` (не View) |
| W1-A1-007 | FIXED-OK | закрыт | Fiber `StructValidator` в `wiring_fiber.go` |
| W1-A1-008 | FIXED-OK | закрыт на handler | JQL → `service.SearchIssuesJQL` (Ent `GetDB` остаётся в service — см. A2W1-2) |
| W1-A1-009 | FIXED-OK | закрыт | Ent-negative + IDOR suite |
| W1-A1-010 | FIXED-OK | закрыт | foreign-project delete CF |
| W1-A1-011 | FIXED-OK | закрыт | Ent-negative authz (TASK-037) |
| W1-A1-012 | FIXED-OK | закрыт | validator + `handler_struct_validator_test` |

**REGRESS:** 0 · **RESIDUAL (этот слот):** 0 · **NEW (этот слот):** 0

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| backend/internal/tracker/handler.go | 98 | 0 | чисто |
| backend/internal/tracker/handler_automation.go | 121 | 0 | чисто |
| backend/internal/tracker/handler_boards.go | 88 | 0 | чисто |
| backend/internal/tracker/handler_comments.go | 222 | 0 | чисто |
| backend/internal/tracker/handler_custom_fields.go | 86 | 0 | чисто |
| backend/internal/tracker/handler_issue_message_link.go | 113 | 0 | чисто |
| backend/internal/tracker/handler_issues.go | 249 | 0 | чисто |
| backend/internal/tracker/handler_labels.go | 137 | 0 | чисто |
| backend/internal/tracker/handler_links.go | 124 | 0 | чисто |
| backend/internal/tracker/handler_time.go | 138 | 0 | чисто |
| backend/internal/tracker/handler_workflow.go | 197 | 0 | чисто |
| backend/internal/tracker/dto.go | 314 | 0 | чисто |
| backend/internal/tracker/events.go | 75 | 0 | чисто |
| backend/internal/tracker/export_hooks.go | 116 | 0 | чисто |
| backend/internal/tracker/provisioner.go | 23 | 0 | чисто |
| backend/internal/tracker/workflow.go | 24 | 0 | чисто |
| backend/testframework/tracker/handler_automation_flow_test.go | 107 | 0 | чисто |
| backend/testframework/tracker/handler_flow_test.go | 266 | 0 | чисто |
| backend/testframework/tracker/handler_jql_flow_test.go | 44 | 0 | чисто |
| backend/testframework/tracker/handler_rbac_test.go | 112 | 0 | чисто |
| backend/testframework/tracker/handler_routes_test.go | 132 | 0 | чисто |
| backend/testframework/tracker/handler_wave103_test.go | 106 | 0 | чисто |
| backend/testframework/tracker/handler_wave106_test.go | 117 | 0 | чисто |
| backend/testframework/tracker/handler_wave109_test.go | 82 | 0 | чисто |
| backend/testframework/tracker/handler_wave112_test.go | 101 | 0 | чисто |
| backend/testframework/tracker/handler_wave114_test.go | 55 | 0 | чисто |
| backend/testframework/tracker/handler_wave115_test.go | 37 | 0 | чисто |
| backend/testframework/tracker/handler_wave117_test.go | 42 | 0 | чисто |
| backend/testframework/tracker/handler_wave118_test.go | 92 | 0 | чисто |
| backend/testframework/tracker/handler_wave119_test.go | 152 | 0 | чисто |
| backend/testframework/tracker/handler_wave120_test.go | 187 | 0 | чисто |
| backend/testframework/tracker/handler_wave121_test.go | 80 | 0 | чисто |
| backend/testframework/tracker/handler_wave124_test.go | 20 | 0 | чисто |
| backend/testframework/tracker/handler_wave95_test.go | 124 | 0 | чисто |
| backend/testframework/tracker/handler_workflow_labels_test.go | 191 | 0 | чисто |
| backend/testframework/tracker/provisioner_test.go | 23 | 0 | чисто |

Файлов в списке: 36. Прочитано: 36. Пропущено: 0.

## Находки (активные A2)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|----------|-------------------|------------|
| — | — | — | — | Нет активных находок в handler-слоте | Все A1 P1/P2 handler IDOR/perm/validator закрыты; Ent-negative есть | — |

## Чисто

- RBAC Fiber v3 order сохранён; CreateBoard = `PermTaskEdit`.
- Workflow/CF/labels/links/comments/time — `ParseProjectID` + scoped service calls.
- JQL handler не трогает Ent (`SearchIssuesJQL`).
- DTO validate-теги исполняются через StructValidator.
- Wave/mock тесты остаются happy-path; IDOR закрыт `project_scope_idor_task013` / `authz_ent_negative_task037` / `create_board_perm_task042`.

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 0 |
| P3 | 0 |
| **Total** | **0** |

A1 baseline: 12 → FIXED-OK 12 · REGRESS 0.
