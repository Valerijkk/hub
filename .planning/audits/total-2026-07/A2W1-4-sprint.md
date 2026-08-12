# A2 W1 агент 4 — sprint (re-audit)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> HEAD: `0d6a659fe9dd56a473cea0ace40ccae280970cbc` (`0d6a659f`)  
> Baseline: [A1-4-sprint.md](A1-4-sprint.md)  
> Периметр: 19 файлов (6 prod + 13 tests) · Прочитано: 19 · Пропущено: 0

## Сверка с A1 (W1-A4-*)

| A1 ID | Tag | Статус | Комментарий |
|-------|-----|--------|-------------|
| W1-A4-001 | FIXED-OK | закрыт | `CompleteSprintAtomic` одна Ent-tx |
| W1-A4-002 | FIXED-OK | закрыт | Reveal Lua CAS |
| W1-A4-003 | FIXED-OK | закрыт | StartPoker pipeline votes+TTL |
| W1-A4-004 | FIXED-OK | закрыт | CastVote Lua fail-closed HGET |
| W1-A4-005 | FIXED-OK | закрыт | `authorizePokerSession` + `project_id` в Redis |
| W1-A4-006 | FIXED-OK | закрыт | `EnsureIssueInProject` на StartPoker |
| W1-A4-007 | FIXED-OK | закрыт | UpdateSprint end≥start |
| W1-A4-008 | RESIDUAL | живо | schema Sprint **без** `version`; UpdateOneID last-write-wins. TASK-038: нужна Ent-миграция / approval |
| W1-A4-009 | FIXED-OK | закрыт | `StartSprintAtomic` + conflict |
| W1-A4-010 | FIXED-OK | закрыт | Velocity Count fail-closed |
| W1-A4-011 | FIXED-OK | закрыт | Reveal script err → Internal |
| W1-A4-012 | FIXED-OK | закрыт | ListSprints counts fail-closed |
| W1-A4-013 | FIXED-OK | закрыт | `ClearIssuesSprint(projectID, …)` |
| W1-A4-014 | FIXED-OK | закрыт | poker stranger 403 (`p2_be11_poker_start_test`) |
| W1-A4-015 | FIXED-OK | закрыт | mock/Ent missing-issue (`add_issues_missing_task059`) |
| W1-A4-016 | FIXED-OK | закрыт | StartSprint_NotFound точный 404 |

**REGRESS:** 0

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| backend/internal/sprint/dto.go | 73 | 0 | чисто |
| backend/internal/sprint/export_hooks.go | 14 | 0 | чисто |
| backend/internal/sprint/handler.go | 495 | 0 | чисто |
| backend/internal/sprint/poker.go | 214 | 0 | чисто |
| backend/internal/sprint/repository.go | 506 | 1 | замечания |
| backend/internal/sprint/service.go | 324 | 0 | чисто |
| backend/testframework/sprint/fail_closed_a3015_test.go | 38 | 0 | чисто |
| backend/testframework/sprint/fail_closed_a3016_test.go | 53 | 0 | чисто |
| backend/testframework/sprint/handler_flow_test.go | 185 | 0 | чисто |
| backend/testframework/sprint/handler_rbac_test.go | 125 | 0 | чисто |
| backend/testframework/sprint/handler_routes_test.go | 177 | 0 | чисто |
| backend/testframework/sprint/handler_wave113_test.go | 97 | 0 | чисто |
| backend/testframework/sprint/list_batch_counts_a13006_test.go | 137 | 0 | чисто |
| backend/testframework/sprint/repository_ent_test.go | 194 | 0 | чисто |
| backend/testframework/sprint/repository_get_in_project_a3001_test.go | 88 | 0 | чисто |
| backend/testframework/sprint/repository_list_no_issues_a13006_test.go | 53 | 0 | чисто |
| backend/testframework/sprint/service_extended_test.go | 172 | 0 | чисто |
| backend/testframework/sprint/service_test.go | 621 | 0 | чисто |
| backend/testframework/sprint/service_wave114_test.go | 16 | 0 | чисто |

Файлов в списке: 19. Прочитано: 19. Пропущено: 0.

## Находки (активные A2)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|----------|-------------------|------------|
| A2W1-A4-001 | RESIDUAL | P2 | `backend/ent/schema/sprint.go:17-43`; `repository.go:144-175` | Sprint schema **без** `version`; `UpdateSprint`/`UpdateSprintStatus`/`UpdateRetroNotes` — `UpdateOneID` без VersionEQ | TZ: optimistic lock на каждой изменяемой Ent-схеме. Concurrent PUT/retro/start → silent last-write-wins (A3-019 CAS poker закрыт отдельно) | Добавить `field.Int("version").Default(1)` + миграция (approval); `Where(VersionEQ).AddVersion(1)` → 409 |

## Чисто

- A3-001 GetSprintInProject / IDOR suite — без регрессии.
- CompleteSprintAtomic + StartSprintAtomic + poker Lua/RBAC — подтверждены чтением + `p2_be11_poker_start_test`.
- WIP/Start fail-closed (A3-015/016) живы.
- Fiber v3 RBAC на sprint routes; poker vote/reveal через `authorizePokerSession`.

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 1 |
| P3 | 0 |
| **Total** | **1** |

A1 baseline: 16 → FIXED-OK 15 · RESIDUAL 1 · NEW 0 · REGRESS 0.
