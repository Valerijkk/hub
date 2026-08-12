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
| A3W1-2-001 | FIXED | P1 | `service_issues.go` | `UpdateIssue` принимал `Labels` без `CountLabelsInProject` | CreateIssue (P1-BE-03) отклоняет чужие метки; Update → cross-project M2M | FIXED TASK-134: `CountLabelsInProject` + `TestUpdateIssue_RejectsForeignLabel_A3W1_2_001` |
| A3W1-2-002 | FIXED | P2 | `service_labels.go` | `CreateLabel`: `groupID` без project scope | метка A → LabelGroup B | FIXED TASK-134: `FindLabelGroupInProject` + `TestCreateLabel_RejectsForeignGroup_A3W1_2_002` |
| A3W1-2-003 | FIXED | P2 | `automation_service.go` | `channel_id` только non-empty | правило A → канал B | FIXED TASK-134: `ensureActionChannelInProject` + `TestCreateUpdateRule_RejectsForeignChannel_A3W1_2_003` |
| A3W1-2-004 | FIXED | P3 | `service_issue_message_link.go` / `repo_issue_message_link.go` | N+1 `FindIssueByID` | batch/eager | FIXED TASK-134: `WithIssue` eager + call-count assert |
| A3W1-2-005 | FIXED | P3 | `service_links.go` | raw repo err | AppError wrap | FIXED TASK-134: constraint→409 / else→500 + AppError passthrough |
| A3W1-2-006 | FIXED | P3 | `service_comments.go` / `repo_comments.go` | любой IsNotFound → 409 | alive+wrong version → 409; missing → 404 | FIXED TASK-134: Exist distinguish + `TestUpdateComment_MissingVsVersionConflict_A3W1_2_006` |

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
| P1 | 0 |
| P2 | 0 |
| P3 | 0 |
| **Total** | **0** |

A3 residual NEW 6 → FIXED 6 (TASK-134, 2026-07-29). A2 baseline: FIXED-OK 2 · REGRESS 0.
