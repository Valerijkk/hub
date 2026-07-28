# TASK-023 — RBAC cache invalidate + GDPR erase atomic

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P1** · Статус: Открыта  
> Источник: TASK-008 W12 · Classes: P1-BE-33..34 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P1-BACKEND.md`

## [OUTCOME]

Дефекты класса `P1-BE-33..34` закрыты: denial/regression на слое guard зелёные, критерий мутации воспроизводим.

## [ROLE]

- Role: be1
- Responsibility: фикс + тесты на слое guard
- Handoff to: стадия C (мутация независимым агентом)

## [SCOPE AND CONSTRAINTS]

- Files allowed: backend/internal/rbac/**; backend/internal/audit/**; testframework/{rbac,audit}/**; `.planning/backlog/README.md`; этот файл
- Do not touch: несвязанные модули; secrets; DEV_MODE в CI
- Constraints: Docker + `-tags hubtest` для BE; Vitest для FE; один логический коммит на подкласс при необходимости
- Approval record: миграции Ent — только с явным согласованием владельца

## [REFERENCES]

- Реестр W12 P1 BACKEND (ClassIDs P1-BE-33..34)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P1-BE-33..34 с мутацией «да» закрыты по реестру
- **Мутация (C):** demote → сразу 403; erase fail cascade → HardDelete не вызван
- Regression: соответствующий `testframework` / Vitest domain package
- Evidence: вывод тестов, SHA коммитов
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

- Коммит(ы) Conventional (тип/scope EN, описание RU) + путь задачи в теле
- Статус README

## [OPERATING RULES]

- Guard тестировать на реальном слое, не только mock
- Co-authored-by запрещён
