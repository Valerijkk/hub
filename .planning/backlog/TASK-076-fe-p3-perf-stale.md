# TASK-076 — FE perf stale intervals

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P3** · Статус: На проверке
> Источник: TASK-008 W12 · Classes: P3-FE-09 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P3-FRONTEND.md`

## [OUTCOME]

Дефекты P3-FE-09 закрыты.

## [ROLE]

- Role: fe2
- Responsibility: фикс + тесты на слое guard
- Handoff to: стадия C (мутация независимым агентом)

## [SCOPE AND CONSTRAINTS]

- Files allowed: frontend/src/**; `.planning/backlog/README.md`; этот файл
- Do not touch: несвязанные модули; secrets; DEV_MODE в CI
- Constraints: Docker + `-tags hubtest` для BE; Vitest для FE; один логический коммит на подкласс при необходимости
- Approval record: миграции Ent — только с явным согласованием владельца

## [REFERENCES]

- Реестр W12 (P3-FE-09)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P3-FE-09 с мутацией «да» закрыты по реестру
- **Мутация (C):** interval cleanup; Set hoist
- Regression: соответствующий `testframework` / Vitest domain package
- Evidence: вывод тестов, SHA коммитов
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

- Коммит(ы) Conventional (тип/scope EN, описание RU) + путь задачи в теле
- Статус README

## [OPERATING RULES]

- Guard тестировать на реальном слое, не только mock
- Co-authored-by запрещён