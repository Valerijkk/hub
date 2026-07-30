# TASK-041 — Sprint poker CAS/RBAC + StartSprint atomic

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: Закрыта
> Источник: TASK-008 W12 · Classes: P2-BE-11 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P2-BACKEND.md`

## [OUTCOME]

Дефекты P2-BE-11 закрыты.

## [ROLE]

- Role: be2
- Responsibility: фикс + тесты на слое guard
- Handoff to: стадия C (мутация независимым агентом)

## [SCOPE AND CONSTRAINTS]

- Files allowed: backend/internal/sprint/**; `.planning/backlog/README.md`; этот файл
- Do not touch: несвязанные модули; secrets; DEV_MODE в CI
- Constraints: Docker + `-tags hubtest` для BE; Vitest для FE; один логический коммит на подкласс при необходимости
- Approval record: миграции Ent — только с явным согласованием владельца

## [REFERENCES]

- Реестр W12 (P2-BE-11)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P2-BE-11 с мутацией «да» закрыты по реестру
- **Мутация (C):** vote after reveal→reject; два Start→один active
- Regression: соответствующий `testframework` / Vitest domain package
- Evidence: вывод тестов, SHA коммитов
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

- Коммит(ы) Conventional (тип/scope EN, описание RU) + путь задачи в теле
- Статус README

## [OPERATING RULES]

- Guard тестировать на реальном слое, не только mock
- Co-authored-by запрещён

## [REWORK 2026-07-28]

Ent sequential second Start → ErrActiveSprintConflict (не SQLite race).

## [W18-C3 MUTATION — 2026-07-28 be2]

**FAIL.** Evidence: `.planning/audits/total-2026-07/W18-C3-MUTATION.md`

- Vote after reveal: drop reject on Lua result 0 → `TestPoker_VoteAfterReveal_Rejected` RED ✓ (miniredis)
- Dual Start: remove `ErrActiveSprintConflict` in `StartSprintAtomic` → `TestStartSprintAtomic_Ent_Concurrent` stays **GREEN** (SQLite tx serialization masks missing conflict). Mock concurrent embeds guard in mock.
- Need Ent-runtime test that sequential/second Start after first active → conflict without relying on SQLite single-tx accident.

## [W1-be2 MUTATION — 2026-07-29]

**PASS.** Evidence: .planning/audits/total-2026-07/MUTATION-2026-07-29-W1-be2.md

- Remove ErrActiveSprintConflict in StartSprintAtomic → TestStartSprintAtomic_Ent_SecondActiveConflicts RED
- GREEN after restore. W18 SQLite concurrent gap closed by sequential Ent test.

## [MUTATION CRITERIA] — verifier-ready

> W13 slot 3 refill / be2 · criteria only (no mutate) · 2026-07-30 · HEAD guards verified · Class P2-BE-11

| ID | Production file:line (guard) | слой | Exact test RED | Kill recipe |
|----|------------------------------|------|----------------|-------------|
| W1-A4-002 | `backend/internal/sprint/poker.go:29-36` + `CastVote:141-145` — Lua `revealed=='true'→0` → `NewBadRequest("голоса уже раскрыты")` | PokerService (Redis Lua) | `TestPoker_VoteAfterReveal_Rejected` · `backend/testframework/sprint/p2_be11_poker_start_test.go` | убрать reject на Lua result `0` / не проверять revealed в script → vote after reveal OK → «раскрыты» assert RED |
| W1-A4-009 | `backend/internal/sprint/repository.go:252-264` — `StartSprintAtomic`: active other → `*ErrActiveSprintConflict` | EntRepository | `TestStartSprintAtomic_Ent_SecondActiveConflicts` · `p2_be11_poker_start_test.go` | убрать `ErrActiveSprintConflict` return (пропустить conflict-check) → второй Start active / ErrorAs fail → RED |
| W1-A4-005 | `backend/internal/sprint/handler.go:452-466` — `authorizePokerSession`: non-member → 403 | handler + RBAC | `TestPoker_RBAC_StrangerForbidden` · `p2_be11_poker_start_test.go` | убрать `authorizePokerSession` / PermProjectView check на vote → outsider 2xx → 403 assert RED |
| W1-A4-007 | `backend/internal/sprint/service.go:107-109` — `UpdateSprint`: `end.Before(start)` → BadRequest | service | `TestUpdateSprint_EndBeforeStart_BadRequest` · `p2_be11_poker_start_test.go` | убрать date-order check → end&lt;start Update OK / EndDate mutated → BadRequest assert RED |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Flag |
|----|-----------|------|
| 002 | PokerService + miniredis Lua CAS | **OK** — Redis (prod Lua); не Ent |
| 009 | Ent `StartSprintAtomic` sequential second Start | **OK** — Ent |
| 005 | Fiber handler + Ent RBAC membership | **OK** — Ent RBAC |
| 007 | SprintService + mockRepository | **mock≠Ent** (GAP_NOTED): date guard на service; не Ent UpdateSprint IT |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: W1-A4-007` · `docs-only IDs: (none)`

### Gaps

- Docs-only: **N/A**.
- `TestStartSprint_Concurrent_OnlyOneActive` / `TestStartSprintAtomic_Ent_Concurrent` — **не** canonical kill: mock embeds guard / SQLite tx race (W18 FAIL). Канон dual-Start: `TestStartSprintAtomic_Ent_SecondActiveConflicts`.
- VotesKey TTL (`TestPoker_VotesKeyTTL_InitializedOnStart`) — auxiliary; не ACCEPTANCE kill.

## [W13-be1 MUTATION — 2026-07-30]

**PASS → Закрыта.** Evidence: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-be1-task041.md`

| ID | Verdict | RED |
|----|---------|-----|
| W1-A4-002 | PROVEN | `TestPoker_VoteAfterReveal_Rejected` |
| W1-A4-009 | PROVEN | `TestStartSprintAtomic_Ent_SecondActiveConflicts` |
| W1-A4-005 | PROVEN | `TestPoker_RBAC_StrangerForbidden` |
| W1-A4-007 | PROVEN + GAP_NOTED (mock≠Ent) | `TestUpdateSprint_EndBeforeStart_BadRequest` |

leftover MUTATED: нет.
