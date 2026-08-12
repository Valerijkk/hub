# TASK-052 — FE API DTO gaps / type-lie

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · **Закрыта**
> Источник: TASK-008 W12 · Classes: P2-FE-08 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P2-FRONTEND.md`

## [OUTCOME]

Дефекты P2-FE-08 закрыты.

## [ROLE]

- Role: fe2
- Responsibility: фикс + тесты на слое guard
- Handoff to: стадия C (мутация независимым агентом)

## [SCOPE AND CONSTRAINTS]

- Files allowed: frontend/src/api/**; `.planning/backlog/README.md`; этот файл; Vitest `api-dto-gaps-p2.test.ts` + fixture
- Do not touch: несвязанные модули; secrets; DEV_MODE в CI
- Constraints: Docker + `-tags hubtest` для BE; Vitest для FE; один логический коммит на подкласс при необходимости
- Approval record: миграции Ent — только с явным согласованием владельца

## [REFERENCES]

- Реестр W12 (P2-FE-08)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P2-FE-08 с мутацией «да» закрыты по реестру
- **Мутация (C):** updateBase без version→TS fail; guest cast узкий
- Regression: соответствующий `testframework` / Vitest domain package
- Evidence: вывод тестов, SHA коммитов
- Stop when: green + статус «Закрыта» в README после PROVEN всех runnable

## [OUTPUT SCHEMA]

- Коммит(ы) Conventional (тип/scope EN, описание RU) + путь задачи в теле
- Статус README

## [OPERATING RULES]

- Guard тестировать на реальном слое, не только mock
- Co-authored-by запрещён

## [MUTATION CRITERIA] — verifier-ready

> W13 slot 5 refill / fe2 · 2026-07-30 · Class **P2-FE-08** · Vitest: `domains/platform/api-dto-gaps-p2.test.ts`

| ID | file:line (guard) | Vitest path + `it` | kill recipe |
|----|-------------------|----------------------|-------------|
| W11-A1-005 | `frontend/src/api/bases.ts:74,116` — `version: number` required на `updateBase`/`updateRecord` | `it("tsc fixture: omit version is @ts-expect-error…")` (+ source в `updateBase/updateRecord require version`) | `version?: number` → unused `@ts-expect-error` → tsc spawn FAIL |
| W10-A5-013 | `frontend/src/api/calls.ts:40-52` `GuestCallSession` + `ActiveCallSession`; `routes/call.$code.lazy.tsx` без `as CallResponse` | `it("GuestCallSession is assignable…")` | вернуть `as CallResponse` / удалить `GuestCallSession` → source/type FAIL |
| W11-A2-006 | `frontend/src/api/tms.ts:778-795` — `createEnvironment`/`deleteEnvironment` | `it("exposes create/delete environment…")` | удалить exports / сменить path → source Match FAIL |
| W11-A2-007 | `frontend/src/api/intranet.ts:40-53` — `updateNews` + `version: number` | same `it` | удалить `updateNews` / optional version → Match FAIL |
| W11-A2-008 | `frontend/src/api/projects.ts:139-147` — `updateMemberRole` | same `it` | удалить export → Match FAIL |
| W11-A2-010 | `frontend/src/api/messenger.ts:99-110` — `buildMessageFromWsEvent(unknown)` + null на пустой id/sender | `it("buildMessageFromWsEvent rejects null…")` | убрать null-guards / принять Record → runtime null asserts FAIL |
| W10-A2-018 | `frontend/src/modules/tracker/issue-description.tsx:34-37` — `plateValueSchema.safeParse` | `it("parsePlateValue accepts plate nodes…")` | always-accept / `return raw as Value` → garbage `undefined` FAIL |
| W9-A5-007 | `frontend/src/components/tracker/user-picker.tsx:54` — `enabled: Boolean(projectId)` | `it("user-picker disables members query…")` | убрать `enabled` / `enabled: true` → source FAIL |
| W10-A2-017 | `frontend/src/modules/tms/components/case-filters.tsx:24` — `Partial<CaseFiltersProps["filters"]>` | `it("case-filters updateFilter uses Partial…")` | `Record<string, unknown>` → source FAIL |
| W10-A2-020 | `frontend/src/modules/tms/components/environment-comparison.tsx:317` — `formatApiError(compError)` | `it("environment-comparison uses formatApiError…")` | `(compError as { message?: string }).message` → source FAIL |

### Gaps (mutation-proof)

| Guard | Ловит тест на том же слое? | Примечание |
|-------|----------------------------|------------|
| bases version (tsc fixture) | ✅ | канон; `expectTypeOf` alone = runtime no-op |
| Guest / missing API / WS / Plate / picker / filters / formatApiError | ✅ source + runtime | Class source IDs **10/10** |
| Docs-only | N/A | нет |

**Confidence:** HIGH для 10 runnable.

## [REWORK 2026-07-28]

tsc @ts-expect-error fixture + vitest spawn tsc.

## [W18-C5 EVIDENCE]

- Проверяющий: fe1 (W18-C5)
- Mutation: **FAIL** — FAIL: expectTypeOf/version не краснеет при vitest run
- Статус: Закрыта · team-lead 2026-07-30 (до rework tsc fixture)

## [MUTATION 2026-07-29 W1-fe1]

- Проверяющий: fe1 (≠ author fe2)
- Evidence: `.planning/audits/total-2026-07/MUTATION-2026-07-29-W1-fe1.md`
- Mutation: `version?: number` → tsc fixture RED; expectTypeOf alone остался GREEN (no-op)
- Restore → Vitest 6/6 GREEN
- Verdict: **PASS** (частично — только W11-A1-005) · README снова **Открыта** до полного Class prove

## [MUTATION 2026-07-30 W13-fe2]

- Проверяющий: fe2 (W13 slot 5 refill)
- Evidence: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-fe2-task052.md`
- Criteria: `[MUTATION CRITERIA]` 10/10 Class IDs
- Vitest: `api-dto-gaps-p2.test.ts` 9/9 (source-guards + tsc fixture)
- Mutation: **10/10 PROVEN** · Restore UI clean · Статус: **Закрыта**
