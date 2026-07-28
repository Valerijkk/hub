# TASK-036 — DTO StructValidator + safe error mapping

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: Исправлена  
> W18-C2b: PASS (be1 mutation) · артефакт `.planning/audits/total-2026-07/W18-C2b-MUTATION.md`  
> Источник: TASK-008 W12 · Classes: P2-BE-01,P2-BE-07 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P2-BACKEND.md`

## [OUTCOME]

Дефекты P2-BE-01/07 закрыты: StructValidator + TZ error contract.

## [ROLE]

- Role: be1
- Responsibility: фикс + тесты на слое guard
- Handoff to: стадия C (мутация независимым агентом)

## [SCOPE AND CONSTRAINTS]

- Files allowed: backend/internal/**; backend/cmd/server/**; `.planning/backlog/README.md`; этот файл
- Do not touch: несвязанные модули; secrets; DEV_MODE в CI
- Constraints: Docker + `-tags hubtest` для BE; Vitest для FE; один логический коммит на подкласс при необходимости
- Approval record: миграции Ent — только с явным согласованием владельца

## [REFERENCES]

- Реестр W12 (P2-BE-01,P2-BE-07)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P2-BE-01,P2-BE-07 с мутацией «да» закрыты по реестру
- **Мутация (C):** пустой title→400; Forbidden без internals; снять validator→201
- Regression: соответствующий `testframework` / Vitest domain package
- Evidence: вывод тестов, SHA коммитов
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

- Коммит(ы) Conventional (тип/scope EN, описание RU) + путь задачи в теле
- Статус README

## [OPERATING RULES]

- Guard тестировать на реальном слое, не только mock
- Co-authored-by запрещён
