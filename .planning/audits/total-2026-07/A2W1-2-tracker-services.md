# A2 W1 агент 2 — tracker services (re-audit)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> HEAD: `0d6a659fe9dd56a473cea0ace40ccae280970cbc` (`0d6a659f`)  
> Baseline: [A1-2-tracker-services.md](A1-2-tracker-services.md)  
> Периметр: 33 файла (13 prod + 20 tests) · Прочитано: 33 · Пропущено: 0

## Сверка с A1 (W1-A2-*)

| A1 ID | Tag | Статус | Комментарий |
|-------|-----|--------|-------------|
| W1-A2-001 | FIXED-OK | закрыт | `DeleteCustomField(ctx, projectID, fieldID)` |
| W1-A2-002 | FIXED-OK | закрыт | Update/Delete workflow status + projectID |
| W1-A2-003 | FIXED-OK | закрыт | CreateTransition + status ∈ project |
| W1-A2-004 | FIXED-OK | закрыт | DeleteTransition scoped |
| W1-A2-005 | FIXED-OK | закрыт | CreateIssue: Status/Parent/Labels ∈ project |
| W1-A2-006 | FIXED-OK | закрыт | `FindChannelInProject` |
| W1-A2-007 | FIXED-OK | закрыт | MoveIssue VersionEQ + AddVersion |
| W1-A2-008 | FIXED-OK | закрыт | ListIssues `limit+1` → HasMore |
| W1-A2-009 | FIXED-OK | закрыт | FindWorkflowStatusInProject на Update/Move |
| W1-A2-010 | FIXED-OK | закрыт | ErrorHandler маскирует non-AppError → `"internal server error"` |
| W1-A2-011 | FIXED-OK | закрыт | DeleteComment NotFound тест |
| W1-A2-012 | FIXED-OK | закрыт | mock/Ent guards |
| W1-A2-013 | FIXED-OK | закрыт | mirror → `slog.WarnContext` |
| W1-A2-014 | FIXED-OK | закрыт | DeleteIssue LogActivity «deleted» |
| W1-A2-015 | FIXED-OK | закрыт | nil-guard edges в ListTransitions |
| W1-A1-008 (service) | RESIDUAL | живо | SearchIssuesJQL всё ещё `repo.GetDB()` |

**REGRESS:** 0

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| backend/internal/tracker/service.go | 87 | 0 | чисто |
| backend/internal/tracker/service_activity.go | 29 | 0 | чисто |
| backend/internal/tracker/service_boards.go | 115 | 0 | чисто |
| backend/internal/tracker/service_comments.go | 116 | 0 | чисто |
| backend/internal/tracker/service_converters.go | 296 | 0 | чисто |
| backend/internal/tracker/service_custom_fields.go | 49 | 0 | чисто |
| backend/internal/tracker/service_issue_message_link.go | 230 | 0 | чисто |
| backend/internal/tracker/service_issues.go | 423 | 2 | замечания |
| backend/internal/tracker/service_labels.go | 74 | 0 | чисто |
| backend/internal/tracker/service_links.go | 111 | 0 | чисто |
| backend/internal/tracker/service_time.go | 83 | 0 | чисто |
| backend/internal/tracker/service_workflow.go | 110 | 0 | чисто |
| backend/internal/tracker/automation_service.go | 158 | 0 | чисто |
| backend/testframework/tracker/service_converters_extended_test.go | 158 | 0 | чисто |
| backend/testframework/tracker/service_converters_test.go | 187 | 0 | чисто |
| backend/testframework/tracker/service_helpers_test.go | 84 | 0 | чисто |
| backend/testframework/tracker/service_issue_message_link_test.go | 130 | 0 | чисто |
| backend/testframework/tracker/service_test.go | 1201 | 0 | чисто |
| backend/testframework/tracker/service_wave100_test.go | 93 | 0 | чисто |
| backend/testframework/tracker/service_wave109_test.go | 65 | 0 | чисто |
| backend/testframework/tracker/service_wave112_test.go | 172 | 0 | чисто |
| backend/testframework/tracker/service_wave113_test.go | 60 | 0 | чисто |
| backend/testframework/tracker/service_wave114_test.go | 107 | 0 | чисто |
| backend/testframework/tracker/service_wave115_test.go | 68 | 0 | чисто |
| backend/testframework/tracker/service_wave117_test.go | 51 | 0 | чисто |
| backend/testframework/tracker/service_wave118_test.go | 71 | 0 | чисто |
| backend/testframework/tracker/service_wave119_test.go | 60 | 0 | чисто |
| backend/testframework/tracker/service_wave121_test.go | 42 | 0 | чисто |
| backend/testframework/tracker/service_workflow_extra_test.go | 170 | 0 | чисто |
| backend/testframework/tracker/comment_scope_a3011_test.go | 46 | 0 | чисто |
| backend/testframework/tracker/comment_soft_delete_a8006_test.go | 80 | 0 | чисто |
| backend/testframework/tracker/log_time_a3005_test.go | 75 | 0 | чисто |
| backend/testframework/tracker/start_timer_concurrent_a3005_test.go | 64 | 0 | чисто |

Файлов в списке: 33. Прочитано: 33. Пропущено: 0.

## Находки (активные A2)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|----------|-------------------|------------|
| A2W1-A2-001 | RESIDUAL | P3 | `service_issues.go:346-375` | `SearchIssuesJQL` ходит в Ent через `s.repo.GetDB().Issue.Query()` | Handler layering закрыт (W1-A1-008), но service всё ещё зависит от Ent client; нарушает «repo API only» и усложняет mock | `repo.SearchIssuesJQL(ctx, projectID, predicates, limit)` без экспорта `GetDB` |
| A2W1-A2-002 | NEW | P2 | `service_issues.go:48-53`, `288-293` | `AssigneeID` парсится, **не** проверяется membership в `projectID` | Можно назначить UUID пользователя вне проекта → уведомления/UI на чужого; Status/Parent/Labels уже scoped, assignee — нет | `FindMember(projectID, assigneeID)` / CountMembers; mismatch → 400/404 |

## Чисто

- IDOR workflow/CF/refs/channel — закрыты (TASK-013).
- Optimistic lock Move/Update, HasMore limit+1, soft-delete activity.
- AutomationService Create/Find/Delete project-scoped.
- A3-002/005/011 на service path — без регрессии.

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 1 |
| P3 | 1 |
| **Total** | **2** |

A1 baseline: 15 → FIXED-OK 15 · RESIDUAL carried 1 (JQL GetDB) · NEW 1 · REGRESS 0.
