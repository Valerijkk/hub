# A2 W1 агент 3 — tracker repos + automation (re-audit)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> HEAD: `0d6a659fe9dd56a473cea0ace40ccae280970cbc` (`0d6a659f`)  
> Baseline: [A1-3-tracker-repos.md](A1-3-tracker-repos.md)  
> Периметр: 37 файлов (15 prod + 22 tests) · Прочитано: 37 · Пропущено: 0

## Сверка с A1 (W1-A3-*)

| A1 ID | Tag | Статус | Комментарий |
|-------|-----|--------|-------------|
| W1-A3-001 | FIXED-OK | закрыт | DeleteCustomField `HasProjectWith` |
| W1-A3-002 | FIXED-OK | закрыт | Update/Delete status scoped |
| W1-A3-003 | FIXED-OK | закрыт | Create/Delete transition scoped |
| W1-A3-004 | FIXED-OK | закрыт | MoveIssue VersionEQ + DeletedAtIsNil |
| W1-A3-005 | FIXED-OK | закрыт | UpdateIssue `DeletedAtIsNil` |
| W1-A3-006 | FIXED-OK | закрыт | LogTime/StopTimer в одной tx |
| W1-A3-007 | RESIDUAL | живо | Engine Handle* void; consumer Ack после вызова даже при DB fail list rules |
| W1-A3-008 | FIXED-OK | закрыт | `logOutboxPublishErr` вместо `_ = Publish` |
| W1-A3-009 | FIXED-OK | закрыт | UpdateAutomationRule `ProjectIDEQ` |
| W1-A3-010 | FIXED-OK | закрыт | ListActivities/Comments/TimeLogs + Limit caps |
| W1-A3-011 | FIXED-OK | закрыт | Ack/Nak asserts в тестах |

**REGRESS:** 0

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| backend/internal/tracker/repository.go | 273 | 0 | чисто |
| backend/internal/tracker/repo_activity.go | 44 | 0 | чисто |
| backend/internal/tracker/repo_automation.go | 116 | 0 | чисто |
| backend/internal/tracker/repo_boards.go | 107 | 0 | чисто |
| backend/internal/tracker/repo_comments.go | 175 | 0 | чисто |
| backend/internal/tracker/repo_custom_fields.go | 46 | 0 | чисто |
| backend/internal/tracker/repo_issue_message_link.go | 83 | 0 | чисто |
| backend/internal/tracker/repo_issues.go | 436 | 0 | чисто |
| backend/internal/tracker/repo_labels.go | 72 | 0 | чисто |
| backend/internal/tracker/repo_links.go | 122 | 0 | чисто |
| backend/internal/tracker/repo_time.go | 170 | 0 | чисто |
| backend/internal/tracker/repo_workflow.go | 224 | 0 | чисто |
| backend/internal/tracker/automation_consumer.go | 122 | 1 | замечания |
| backend/internal/tracker/automation_dto.go | 73 | 0 | чисто |
| backend/internal/tracker/automation_engine.go | 129 | 1 | замечания |
| backend/testframework/tracker/automation_ack_a3006_test.go | 60 | 0 | чисто |
| backend/testframework/tracker/automation_consumer_test.go | 117 | 0 | чисто |
| backend/testframework/tracker/automation_engine_events_test.go | 90 | 0 | чисто |
| backend/testframework/tracker/automation_engine_test.go | 86 | 0 | чисто |
| backend/testframework/tracker/automation_engine_wave100_test.go | 231 | 0 | чисто |
| backend/testframework/tracker/automation_service_extra_test.go | 70 | 0 | чисто |
| backend/testframework/tracker/board_label_idor_blockf_test.go | 42 | 0 | чисто |
| backend/testframework/tracker/link_scope_a3011_postgres_ent_test.go | 36 | 0 | чисто |
| backend/testframework/tracker/link_scope_a3011_test.go | 36 | 0 | чисто |
| backend/testframework/tracker/link_target_scope_a3011_postgres_ent_test.go | 43 | 0 | чисто |
| backend/testframework/tracker/link_target_scope_a3011_test.go | 79 | 0 | чисто |
| backend/testframework/tracker/postgres_ent_helper_test.go | 102 | 0 | чисто |
| backend/testframework/tracker/reaction_scope_a3011_postgres_ent_test.go | 31 | 0 | чисто |
| backend/testframework/tracker/reaction_scope_a3011_test.go | 31 | 0 | чисто |
| backend/testframework/tracker/repository_board_cap_a13001_test.go | 43 | 0 | чисто |
| backend/testframework/tracker/repository_extended_test.go | 97 | 0 | чисто |
| backend/testframework/tracker/repository_more_test.go | 161 | 0 | чисто |
| backend/testframework/tracker/repository_project_links_a13004_test.go | 49 | 0 | чисто |
| backend/testframework/tracker/repository_test.go | 149 | 0 | чисто |
| backend/testframework/tracker/repository_wave100_test.go | 205 | 0 | чисто |
| backend/testframework/tracker/repository_wave120_test.go | 48 | 0 | чисто |
| backend/testframework/tracker/repository_wave124_test.go | 119 | 0 | чисто |

Файлов в списке: 37. Прочитано: 37. Пропущено: 0.

## Находки (активные A2)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|----------|-------------------|------------|
| A2W1-A3-001 | RESIDUAL | P2 | `automation_engine.go:26-35`; `automation_consumer.go:84-121` | `Handle*` не возвращают `error`; при `ListEnabledAutomationRules` fail — slog + return; consumer всё равно `Ack` | Transient DB error → событие потеряно; MaxDeliver бесполезен. bad JSON → Nak (часть A3-006) есть; engine-path — нет. TASK-040 P2-BE-10 **не** включал tracker | Engine → `error`; consumer: err → Nak/NakWithDelay; успех → Ack |
| A2W1-A3-002 | NEW | P3 | `automation_consumer.go:84-121` | `parseIssueEventKind` → `""` (неизвестный subject): switch no-op, затем Ack | Сообщение с битым/чужим subject молча Ack без метрики | default: slog.Warn + Nak **или** metric `automation_unknown_subject` |

## Чисто

- Project-scope IDOR guards на CF/workflow/issues/labels/boards — без регрессии.
- Outbox publish: `logOutboxPublishErr` (не silent `_=`).
- List caps: MaxIssueActivities / Comments / TimeLogs / ProjectIssueLinks.
- UpdateAutomationRule Where `IDEQ`+`ProjectIDEQ`.
- Ent twins (sqlite+PG) для link/reaction/board/label — живы.

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 1 |
| P3 | 1 |
| **Total** | **2** |

A1 baseline: 11 → FIXED-OK 10 · RESIDUAL 1 · NEW 1 · REGRESS 0.
