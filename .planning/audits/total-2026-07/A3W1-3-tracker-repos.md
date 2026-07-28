# A3 W1 агент 3 — tracker repos + automation (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Baseline: [A2W1-3-tracker-repos.md](A2W1-3-tracker-repos.md) · A1: [A1-3-tracker-repos.md](A1-3-tracker-repos.md)  
> Периметр: 37 файлов (15 prod + 22 tests) · Прочитано: **37/37** · Пропущено: **0**

## Сверка с A2 (A2W1-A3-*)

| A2 ID | Tag | Статус | Комментарий |
|-------|-----|--------|-------------|
| A2W1-A3-001 | FIXED-OK | закрыт | TASK-088: `Handle*` → `error`; `ListEnabledAutomationRules` fail → `fmt.Errorf` → consumer `Nak`; тест `automation_nak_task088_test.go` |
| A2W1-A3-002 | FIXED-OK | закрыт | TASK-088: unknown subject → `Nak` + `automationUnknownSubjectTotal`; тест `TestAutomationConsumer_NakOnUnknownSubject_A2W1_A3_002` |

**REGRESS:** 1 (см. A3W1-A3-001 — побочный эффект A2W1-A3-002 при широком `FilterSubject`)

### Сверка с A1 (кратко)

| A1 ID | Tag на A3 |
|-------|-----------|
| W1-A3-001…006, 008…011 | FIXED-OK (без регрессии) |
| W1-A3-007 | FIXED-OK list-rules path (TASK-088); action-side effects — RESIDUAL A3W1-A3-002 |

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| backend/internal/tracker/repository.go | 305 | 0 | чисто — ISP; `logOutboxPublishErr`; MoveIssue VersionEQ в контракте |
| backend/internal/tracker/repo_activity.go | 53 | 0 | чисто — `MaxIssueActivities` Limit |
| backend/internal/tracker/repo_automation.go | 127 | 0 | чисто — Update/Delete `ProjectIDEQ` |
| backend/internal/tracker/repo_boards.go | 127 | 0 | чисто — project scope + soft-delete + column cap |
| backend/internal/tracker/repo_comments.go | 196 | 0 | чисто — soft-delete; VersionEQ; scoped react; list cap |
| backend/internal/tracker/repo_custom_fields.go | 62 | 0 | чисто — soft Delete + `HasProjectWith` |
| backend/internal/tracker/repo_issue_message_link.go | 92 | 0 | чисто — Delete scoped; channel membership helpers |
| backend/internal/tracker/repo_issues.go | 516 | 0 | чисто — VersionEQ+DeletedAtIsNil Update/Move; JQL; IsProjectMember |
| backend/internal/tracker/repo_labels.go | 97 | 0 | чисто — soft Delete scoped; CountLabelsInProject |
| backend/internal/tracker/repo_links.go | 138 | 0 | чисто — Delete scoped; batch cap; outbox log |
| backend/internal/tracker/repo_time.go | 191 | 1 | замечания — StartTimer lock без DeletedAtIsNil |
| backend/internal/tracker/repo_workflow.go | 246 | 0 | чисто — status/transition project-scoped |
| backend/internal/tracker/automation_consumer.go | 149 | 1 | замечания — FilterSubject шире handled kinds |
| backend/internal/tracker/automation_dto.go | 82 | 0 | чисто — DTO/константы |
| backend/internal/tracker/automation_engine.go | 142 | 1 | замечания — executeRule глотает action err |
| backend/testframework/tracker/automation_ack_a3006_test.go | 67 | 0 | чисто — Nak/Ack bad JSON |
| backend/testframework/tracker/automation_consumer_test.go | 133 | 0 | чисто — Ack/Nak asserts + unknown Nak |
| backend/testframework/tracker/automation_engine_events_test.go | 106 | 0 | чисто — Ent + stubs |
| backend/testframework/tracker/automation_engine_test.go | 95 | 0 | чисто — post channel + max rules |
| backend/testframework/tracker/automation_engine_wave100_test.go | 263 | 0 | чисто — assert-bearing edge cases |
| backend/testframework/tracker/automation_service_extra_test.go | 81 | 0 | чисто — validation / not found |
| backend/testframework/tracker/board_label_idor_blockf_test.go | 57 | 0 | чисто — Ent IDOR board/label |
| backend/testframework/tracker/link_scope_a3011_postgres_ent_test.go | 45 | 0 | чисто — Postgres twin DeleteIssueLink |
| backend/testframework/tracker/link_scope_a3011_test.go | 45 | 0 | чисто — sqlite DeleteIssueLink |
| backend/testframework/tracker/link_target_scope_a3011_postgres_ent_test.go | 52 | 0 | чисто — FindIssueInProject PG |
| backend/testframework/tracker/link_target_scope_a3011_test.go | 93 | 0 | чисто — FindIssueInProject + service |
| backend/testframework/tracker/postgres_ent_helper_test.go | 108 | 0 | чисто — Skip без PG |
| backend/testframework/tracker/reaction_scope_a3011_postgres_ent_test.go | 39 | 0 | чисто — ToggleReaction PG |
| backend/testframework/tracker/reaction_scope_a3011_test.go | 39 | 0 | чисто — ToggleReaction sqlite |
| backend/testframework/tracker/repository_board_cap_a13001_test.go | 51 | 0 | чисто — MaxBoardIssuesPerColumn |
| backend/testframework/tracker/repository_extended_test.go | 127 | 0 | чисто — labels/links/time |
| backend/testframework/tracker/repository_more_test.go | 252 | 0 | чисто — comments/CF/timer/board |
| backend/testframework/tracker/repository_project_links_a13004_test.go | 58 | 0 | чисто — ListProjectIssueLinks batch |
| backend/testframework/tracker/repository_test.go | 177 | 0 | чисто — A3-002 + VersionEQ CRUD |
| backend/testframework/tracker/repository_wave100_test.go | 238 | 0 | чисто — outbox/board/automation |
| backend/testframework/tracker/repository_wave120_test.go | 59 | 0 | чисто — comment outbox + delete rule |
| backend/testframework/tracker/repository_wave124_test.go | 137 | 0 | чисто — optional fields + list filters |

Файлов в списке: 37. Прочитано: 37. Пропущено: 0.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|----------|-------------------|------------|
| A3W1-A3-001 | REGRESS | P2 | `automation_consumer.go:51-57`, `125-129`; publishers: `repo_issues.go:394`, `repo_comments.go:55`, `repo_links.go:59` | `FilterSubject: "project.*.issue.>"` ловит `updated` / `comment_created` / `link_created`; после TASK-088 default → **Nak** (раньше молчаливый Ack) | Каждый edit/comment/link → до 5 redeliveries + Warn spam; тормозит real automation (moved/created/assigned). MaxDeliver сжигается на ожидаемых no-op | `FilterSubjects`: `…issue.moved`, `…created`, `…assigned` **или** known-ignore kinds → Ack/Term без metric; Nak только для мусора |
| A3W1-A3-002 | RESIDUAL | P3 | `automation_engine.go:84-134` | `executeRule` глотает `PostAutomationMessage` / `NotifyAutomation` (slog); `Handle*` всё равно `return nil` → consumer Ack | Transient messenger/notifier fail → side-effect потерян; list-rules path FIXED, action path — нет. Тест `PostMessageError` закрепляет swallow | `executeRule` → `error`; consumer: action fail → NakWithDelay **или** outbox retry queue; идемпотентность правил |
| A3W1-A3-003 | NEW | P3 | `repo_time.go:79-82` | `StartTimer` lock: `UpdateOneID` **без** `DeletedAtIsNil` (в отличие от `LogTime`/`StopTimer` AddTimeSpent) | На soft-deleted issue создаётся open timer; `StopTimer` потом падает на AddTimeSpent → orphan running timer | `Where(issue.DeletedAtIsNil())` на lock; NotFound если trash |

## Чисто (подтверждено после remediation)

- IDOR/scopes: CF/workflow/labels/boards/FindIssueByKey/FindIssueInProject/DeleteIssueLink/ToggleReaction — без регрессии; Ent twins sqlite+PG живы.
- Optimistic lock: `UpdateIssue` / `MoveIssue` — `VersionEQ` + `DeletedAtIsNil` + 409 conflict.
- Time: `LogTime`/`StopTimer` в одной tx; StartTimer conflict + row-lock.
- Outbox: `logOutboxPublishErr` (не silent `_=`).
- List caps: activities/comments/timelogs/board column/project links.
- Automation: engine error → Nak; bad JSON → Nak; unknown garbage → Nak+metric (TASK-088 evidence вне assignment: `automation_nak_task088_test.go`).
- Soft-delete lists: boards/labels/CF `DeletedAtIsNil`.

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 1 |
| P3 | 2 |
| **Total** | **3** |

A2 baseline: 2 активных → FIXED-OK 2 · RESIDUAL 1 (новый ID на action path) · NEW 1 · REGRESS 1.

**Топ ID:** A3W1-A3-001 (FilterSubject Nak storm) · A3W1-A3-002 (action err swallow) · A3W1-A3-003 (StartTimer soft-delete).
