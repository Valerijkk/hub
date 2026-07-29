# TASK-034 — FE a11y nested interactive + aria-label

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P1** · Статус: Закрыта  
> Источник: TASK-008 W12 · Classes: FE-P1-A11Y-CONTROLS · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P1-FRONTEND.md`

## [OUTCOME]

Дефекты класса `FE-P1-A11Y-CONTROLS` закрыты: denial/regression на слое guard зелёные, критерий мутации воспроизводим.

## [ROLE]

- Role: fe2
- Responsibility: фикс + тесты на слое guard
- Handoff to: стадия C (мутация независимым агентом)

## [SCOPE AND CONSTRAINTS]

- Files allowed: frontend/src/components/layout/**; `.planning/backlog/README.md`; этот файл
- Do not touch: несвязанные модули; secrets; DEV_MODE в CI
- Constraints: Docker + `-tags hubtest` для BE; Vitest для FE; один логический коммит на подкласс при необходимости
- Approval record: миграции Ent — только с явным согласованием владельца

## [REFERENCES]

- Реестр W12 P1 FRONTEND (ClassIDs FE-P1-A11Y-CONTROLS)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из FE-P1-A11Y-CONTROLS с мутацией «да» закрыты по реестру
- **Мутация (C):** button⊃Link → axe nested-interactive FAIL
- Regression: соответствующий `testframework` / Vitest domain package
- Evidence: вывод тестов, SHA коммитов
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

- Коммит(ы) Conventional (тип/scope EN, описание RU) + путь задачи в теле
- Статус README

## [OPERATING RULES]

- Guard тестировать на реальном слое, не только mock
- Co-authored-by запрещён

## [W18-C5 EVIDENCE]

- Проверяющий: fe1 (W18-C5), автор Stage B ≠ проверяющий
- Mutation: PASS — см. .planning/audits/total-2026-07/W18-C5-MUTATION.md
- Статус: Исправлена
