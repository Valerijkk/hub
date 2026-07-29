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

- Files allowed: frontend/src/api/**; `.planning/backlog/README.md`; этот файл
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
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

- Коммит(ы) Conventional (тип/scope EN, описание RU) + путь задачи в теле
- Статус README

## [OPERATING RULES]

- Guard тестировать на реальном слое, не только mock
- Co-authored-by запрещён

## [REWORK 2026-07-28]

tsc @ts-expect-error fixture + vitest spawn tsc.

## [W18-C5 EVIDENCE]

- Проверяющий: fe1 (W18-C5)
- Mutation: **FAIL** — FAIL: expectTypeOf/version не краснеет при vitest run
- Статус: Открыта (до rework tsc fixture)

## [MUTATION 2026-07-29 W1-fe1]

- Проверяющий: fe1 (≠ author fe2)
- Evidence: `.planning/audits/total-2026-07/MUTATION-2026-07-29-W1-fe1.md`
- Mutation: `version?: number` → tsc fixture RED; expectTypeOf alone остался GREEN (no-op)
- Restore → Vitest 6/6 GREEN
- Verdict: **PASS** · Статус: **Закрыта**
