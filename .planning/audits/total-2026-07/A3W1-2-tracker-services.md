# A3 W1 агент 2 — tracker services (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`)  
> Baseline: [A2W1-2-tracker-services.md](A2W1-2-tracker-services.md)  
> Периметр: 33 файла (13 prod + 20 tests) · Прочитано: 33 · Пропущено: 0

## Сверка с A2 (W1-A2-*)

| A2 ID | Tag | Статус в HEAD | Комментарий |
|-------|-----|---------------|-------------|
| A2W1-A2-001 | FIXED-OK | закрыт | `SearchIssuesJQL` → `repo.SearchIssuesJQL` (`service_issues.go:374-390`); `GetDB` убран (TASK-122) |
| A2W1-A2-002 | FIXED-OK | закрыт | Create/UpdateIssue: `IsProjectMember` (`service_issues.go:105-114`, `308-320`); тесты `assignee_membership_task088_test.go` (TASK-088) |
| W1-A2-001…015 (A1 closes) | FIXED-OK | без регрессии | project-scope workflow/CF/refs/channel, OL Move/Update, HasMore limit+1, soft-delete activity, nil-guard transitions |

**REGRESS:** 0

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| backend/internal/tracker/service.go | 97 | 0 | чисто |
| backend/internal/tracker/service_activity.go | 35 | 0 | чисто |
| backend/internal/tracker/service_boards.go | 127 | 0 | чисто |
| backend/internal/tracker/service_comments.go | 136 | 1 | замечание |
| backend/internal/tracker/service_converters.go | 328 | 0 | чисто |
| backend/internal/tracker/service_custom_fields.go | 56 | 0 | чисто |
| backend/internal/tracker/service_issue_message_link.go | 247 | 1 | замечание |
| backend/internal/tracker/service_issues.go | 478 | 1 | замечание |
| backend/internal/tracker/service_labels.go | 84 | 1 | замечание |
| backend/internal/tracker/service_links.go | 125 | 1 | замечание |
| backend/internal/tracker/service_time.go | 99 | 0 | чисто |
| backend/internal/tracker/service_workflow.go | 121 | 0 | чисто |
| backend/internal/tracker/automation_service.go | 170 | 1 | замечание |
| backend/testframework/tracker/service_converters_extended_test.go | 175 | 0 | чисто |
| backend/testframework/tracker/service_converters_test.go | 204 | 0 | чисто |
| backend/testframework/tracker/service_helpers_test.go | 95 | 0 | чисто |
| backend/testframework/tracker/service_issue_message_link_test.go | 145 | 0 | чисто |
| backend/testframework/tracker/service_test.go | 1415 | 0 | чисто (mock `IsProjectMember` opt-in) |
| backend/testframework/tracker/service_wave100_test.go | 111 | 0 | чисто |
| backend/testframework/tracker/service_wave109_test.go | 79 | 0 | чисто |
| backend/testframework/tracker/service_wave112_test.go | 201 | 0 | чисто (UUID parse only для labels) |
| backend/testframework/tracker/service_wave113_test.go | 74 | 0 | чисто |
| backend/testframework/tracker/service_wave114_test.go | 126 | 0 | чисто |
| backend/testframework/tracker/service_wave115_test.go | 80 | 0 | чисто |
| backend/testframework/tracker/service_wave117_test.go | 62 | 0 | чисто |
| backend/testframework/tracker/service_wave118_test.go | 87 | 0 | чисто |
| backend/testframework/tracker/service_wave119_test.go | 69 | 0 | чисто |
| backend/testframework/tracker/service_wave121_test.go | 49 | 0 | чисто |
| backend/testframework/tracker/service_workflow_extra_test.go | 206 | 0 | чисто |
| backend/testframework/tracker/comment_scope_a3011_test.go | 62 | 0 | чисто |
| backend/testframework/tracker/comment_soft_delete_a8006_test.go | 96 | 0 | чисто |
| backend/testframework/tracker/log_time_a3005_test.go | 90 | 0 | чисто |
| backend/testframework/tracker/start_timer_concurrent_a3005_test.go | 72 | 0 | чисто |

Файлов в списке: 33. Прочитано: 33. Пропущено: 0.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|----------|-------------------|------------|
| A3W1-2-001 | NEW | P1 | `service_issues.go:330-339` | `UpdateIssue` принимает `Labels` без `CountLabelsInProject` | CreateIssue (P1-BE-03, `:96-104`) отклоняет чужие метки; Update → `ClearLabels().AddLabelIDs` без scope → cross-project M2M. Тест только UUID parse (`service_wave112_test.go:56-74`), нет foreign-label denial | Перед `input.Labels = &labelIDs` — `CountLabelsInProject`; mismatch → 400; Ent-тест зеркало `TestCreateIssue_RejectsForeignLabel_P1BE03` |
| A3W1-2-002 | NEW | P2 | `service_labels.go:17-26` | `CreateLabel`: `groupID` парсится, **не** проверяется ∈ `projectID` | `repo.CreateLabel` делает `SetGroupID` без scope (`repo_labels.go:25-26`) → метка проекта A может ссылаться на LabelGroup проекта B | `FindLabelGroupInProject` / count; NotFound → 404/400 |
| A3W1-2-003 | NEW | P2 | `automation_service.go:140-143` (+ Create/Update path `:47-77`, `:81-115`) | `action_config.channel_id` только non-empty; нет `FindChannelInProject` | Engine постит в любой UUID канала (`automation_engine.go:90-105`) без project filter → правило проекта A → пост в канал проекта B | В Create/UpdateRule: `FindChannelInProject(projectID, channelID)`; mismatch → 400/404 |
| A3W1-2-004 | NEW | P3 | `service_issue_message_link.go:183-189` | `ListThreadIssueLinks`: `FindIssueByID` в цикле | N+1 при многих связях треда | Batch load / WithIssue eager в `ListIssueMessageLinksByThread` |
| A3W1-2-005 | NEW | P3 | `service_links.go:40-42` | `CreateIssueLink` возвращает `err` репо без wrap в AppError | Non-AppError может утечь наверх (в отличие от соседних методов) | Маппить constraint/internal → AppError; AppError пробрасывать |
| A3W1-2-006 | NEW | P3 | `service_comments.go:92-95` | Любой `ent.IsNotFound` от `UpdateComment` → 409 «version mismatch» | Repo `affected==0` покрывает missing/wrong author/soft-deleted **и** version conflict (`repo_comments.go:114-115`); DeleteComment корректно даёт 404. Клиент получает 409 вместо 404 | Как в `UpdateIssue`: distinguish alive+wrong version → 409 vs missing → 404 |

## Чисто

- Assignee membership (TASK-088) и JQL layering (TASK-122) — FIXED-OK, регрессии нет.
- IDOR workflow/CF/board/label delete/channel/message-link/issue-link target — service path scoped.
- Optimistic lock MoveIssue (`existing.Version`), UpdateComment `version ≥ 1`, soft-delete activity на DeleteIssue.
- Time LogTime/StartTimer/StopTimer project+issueKey; concurrent StartTimer покрыт Postgres-тестом.
- Comment soft-delete A8-006 + scope A3-011 без регрессии на service path.

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 1 |
| P2 | 2 |
| P3 | 3 |
| **Total** | **6** |

A2 baseline: 2 active → FIXED-OK 2 · REGRESS 0 · RESIDUAL 0 · NEW 6.
