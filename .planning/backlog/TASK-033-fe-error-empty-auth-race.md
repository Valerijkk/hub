# TASK-033 — FE error≠empty, ErrorBoundary reset, auth race

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P1** · Статус: Открыта  
> Источник: TASK-008 W12 · Classes: FE-P1-ERROR-AS-EMPTY · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P1-FRONTEND.md`

## [OUTCOME]

Дефекты класса `FE-P1-ERROR-AS-EMPTY` закрыты: denial/regression на слое guard зелёные, критерий мутации воспроизводим.

## [ROLE]

- Role: fe1
- Responsibility: фикс + тесты на слое guard
- Handoff to: стадия C (мутация независимым агентом)

## [SCOPE AND CONSTRAINTS]

- Files allowed: frontend/src/hooks/use-auth.ts; components/layout/**; dashboard/**; `.planning/backlog/README.md`; этот файл
- Do not touch: несвязанные модули; secrets; DEV_MODE в CI
- Constraints: Docker + `-tags hubtest` для BE; Vitest для FE; один логический коммит на подкласс при необходимости
- Approval record: миграции Ent — только с явным согласованием владельца

## [REFERENCES]

- Реестр W12 P1 FRONTEND (ClassIDs FE-P1-ERROR-AS-EMPTY)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из FE-P1-ERROR-AS-EMPTY с мутацией «да» закрыты по реестру
- **Мутация (C):** glotнуть error → «Нет активности»; late setUser after logout
- Regression: соответствующий `testframework` / Vitest domain package
- Evidence: вывод тестов, SHA коммитов
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

- Коммит(ы) Conventional (тип/scope EN, описание RU) + путь задачи в теле
- Статус README

## [OPERATING RULES]

- Guard тестировать на реальном слое, не только mock
- Co-authored-by запрещён
