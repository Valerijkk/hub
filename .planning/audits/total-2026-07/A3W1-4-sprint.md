# A3 W1 агент 4 — sprint (FINAL re-audit)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · ветка `cursor-grok-4.5`  
> Mirror: [A2W1-4-sprint.md](A2W1-4-sprint.md) · Baseline A1: [A1-4-sprint.md](A1-4-sprint.md)  
> Периметр: 19 файлов (6 prod + 13 tests) · Прочитано: **19** · Пропущено: **0**  
> Plan: [A3-PLAN.md](A3-PLAN.md) · Assignment: [W1-ASSIGNMENTS.md](W1-ASSIGNMENTS.md) § A4

## Сверка с A2 (A2W1-A4-*)

| A2 ID | Tag | Статус | Комментарий |
|-------|-----|--------|-------------|
| A2W1-A4-001 | FIXED-OK | закрыт | Sprint: `VersionMixin` + migrate TASK-095; `UpdateSprint` → `VersionEQ`+`AddVersion`; DTO `version`; тест вне периметра `version_eq_task095_test.go` |

**REGRESS:** 0

### Сверка A1 IDs (через A2 FIXED-OK) — подтверждение на HEAD

| A1 ID | Tag | Evidence (файл:строка) |
|-------|-----|------------------------|
| W1-A4-001 | FIXED-OK | `CompleteSprintAtomic` `repository.go:335-371`; service `service.go:177-178` |
| W1-A4-002 | FIXED-OK | `revealVotesScript` CAS `poker.go:45-57`, `159-171` |
| W1-A4-003 | FIXED-OK | Start pipeline votes+TTL `poker.go:73-91` |
| W1-A4-004 | FIXED-OK | `castVoteScript` fail-closed HGET `poker.go:27-43`, `137-146` |
| W1-A4-005 | FIXED-OK | `authorizePokerSession` + `project_id` Redis `handler.go:439-467`, `poker.go:79-83` |
| W1-A4-006 | FIXED-OK | `EnsureIssueInProject` `handler.go:426-429`, `service.go:141-149` |
| W1-A4-007 | FIXED-OK | UpdateSprint end≥start `service.go:100-109` |
| W1-A4-008 | FIXED-OK | schema Mixin + `UpdateSprint` VersionEQ (см. A2W1-A4-001) |
| W1-A4-009 | FIXED-OK | `StartSprintAtomic` `repository.go:220-272` |
| W1-A4-010 | FIXED-OK | Velocity Count fail-closed `repository.go:545-563` |
| W1-A4-011 | FIXED-OK | Reveal script err → Internal `poker.go:162-165` |
| W1-A4-012 | FIXED-OK | ListSprints counts fail-closed `service.go:72-76` |
| W1-A4-013 | FIXED-OK | `ClearIssuesSprint(projectID,…)` `repository.go:439-467` |
| W1-A4-014 | FIXED-OK | poker stranger 403 — prod guard + suite вне периметра `p2_be11_poker_start_test.go` |
| W1-A4-015 | FIXED-OK | mock `SetIssuesSprint` + `issueProject` `service_test.go:230-237`; extended test mapping |
| W1-A4-016 | FIXED-OK | StartSprint_NotFound exact 404 `handler_wave113_test.go:30-42` |

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/sprint/dto.go` | 88 | 0 | чисто |
| `backend/internal/sprint/export_hooks.go` | 19 | 0 | чисто |
| `backend/internal/sprint/handler.go` | 565 | 0 | чисто |
| `backend/internal/sprint/poker.go` | 245 | 0 | чисто |
| `backend/internal/sprint/repository.go` | 594 | 2 | замечания |
| `backend/internal/sprint/service.go` | 377 | 1 | замечания |
| `backend/testframework/sprint/fail_closed_a3015_test.go` | 46 | 0 | чисто |
| `backend/testframework/sprint/fail_closed_a3016_test.go` | 61 | 0 | чисто |
| `backend/testframework/sprint/handler_flow_test.go` | 216 | 0 | чисто |
| `backend/testframework/sprint/handler_rbac_test.go` | 145 | 0 | чисто |
| `backend/testframework/sprint/handler_routes_test.go` | 204 | 0 | чисто |
| `backend/testframework/sprint/handler_wave113_test.go` | 111 | 0 | чисто |
| `backend/testframework/sprint/list_batch_counts_a13006_test.go` | 154 | 0 | чисто |
| `backend/testframework/sprint/repository_ent_test.go` | 231 | 0 | чисто |
| `backend/testframework/sprint/repository_get_in_project_a3001_test.go` | 98 | 0 | чисто |
| `backend/testframework/sprint/repository_list_no_issues_a13006_test.go` | 63 | 0 | чисто |
| `backend/testframework/sprint/service_extended_test.go` | 207 | 0 | чисто |
| `backend/testframework/sprint/service_test.go` | 728 | 0 | чисто |
| `backend/testframework/sprint/service_wave114_test.go` | 18 | 0 | чисто |

Файлов в списке: 19. Прочитано: 19. Пропущено: 0.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|----------|-------------------|------------|
| A3W1-A4-001 | NEW | P2 | `repository.go:322-332` (+ move в `348-351`) | `FindIncompleteIssues` без `issue.DeletedAtIsNil()` | Soft-deleted (корзина) незавершённые задачи всё ещё с `sprint_id` попадают в Complete → `CompleteSprintAtomic` переносит их в next planning; tracker list уже фильтрует trash (`repo_issues.go`), sprint complete — нет | Добавить `issue.DeletedAtIsNil()` в Where; regression: soft-delete issue → Complete → trash issue не в next sprint |
| A3W1-A4-002 | NEW | P3 | `repository.go:469-474`, `504-525`, `545-563`, `117-119`/`132-134` | Count/burndown/velocity/`WithIssues` без фильтра soft-deleted issues | Задачи в корзине завышают `issue_count`, burndown remaining, velocity committed/completed и появляются в detail GetSprint | `DeletedAtIsNil` во всех Issue-запросах sprint metrics + `WithIssues(q.Where(…DeletedAtIsNil()))` |
| A3W1-A4-003 | RESIDUAL | P3 | `repository.go:200-206`; `service.go:193-217` | `UpdateRetroNotes` — `UpdateOneID` без `VersionEQ`/`AddVersion`; DTO без `version` | После TASK-095 OL закрыт на PUT name/dates; concurrent PUT `/retro` всё ещё last-write-wins (часть исходного W1-A4-008 / A2W1-A4-001) | `version` в `RetroNotesRequest` + `VersionEQ`+`AddVersion` → 409; `UpdateSprintStatus` (start/complete) — server FSM, OL не обязателен |

## Чисто

- **IDOR A3-001 / RT-005:** все read/mutate через `GetSprintInProject`; `SetIssuesSprint`/`ClearIssuesSprint`/`IssueInProject` с project scope.
- **A3-015 / A3-016:** WIP и Start fail-closed; тесты живы; Start идёт через `StartSprintAtomic`.
- **Poker A3-019 / W1-A4-002…006:** Lua CAS vote/reveal, TTL на Start, `project_id` + `authorizePokerSession`, issue∈project.
- **Complete/Start atomic:** `CompleteSprintAtomic` + hook rollback; `StartSprintAtomic` project row-lock.
- **Fiber v3 RBAC:** `RegisterRoutes` — handler затем `RequirePermission`.
- **N+1 list:** batch `GetSprintIssueCounts`; List без `WithIssues`.
- **Sprint soft-delete (TASK-097):** `DeletedAtIsNil` на sprint list/get/active — без регрессии.
- **dto / export_hooks / fail_closed / handler_* / list_batch / repository_* / service_* tests:** asserts есть; mock OL на UpdateSprint; IDOR suite в `service_test.go`.

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 1 |
| P3 | 2 |
| **Total** | **3** |

### Теги vs A2

| Tag | Count |
|-----|------:|
| FIXED-OK (A2 active) | 1 (`A2W1-A4-001`) |
| FIXED-OK (A1 cascade) | 16 |
| RESIDUAL | 1 (`A3W1-A4-003`) |
| NEW | 2 (`A3W1-A4-001`, `A3W1-A4-002`) |
| REGRESS | 0 |

**Топ IDs:** `A3W1-A4-001` (soft-delete incomplete move) · `A3W1-A4-002` (metrics count trash) · `A3W1-A4-003` (retro без OL).
