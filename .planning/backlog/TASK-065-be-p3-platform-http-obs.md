# TASK-065 — Platform HTTP + obs residuals

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P3** · Статус: Открыта  
> Источник: TASK-008 W12 · Classes: P3-BE-11,P3-BE-12 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P3-BACKEND.md`

## [OUTCOME]

Дефекты P3-BE-11/12 закрыты.

## [ROLE]

- Role: be1
- Responsibility: фикс + тесты на слое guard
- Handoff to: стадия C (мутация независимым агентом)

## [SCOPE AND CONSTRAINTS]

- Files allowed: backend/internal/platform/**; `.planning/backlog/README.md`; этот файл
- Do not touch: несвязанные модули; secrets; DEV_MODE в CI
- Constraints: Docker + `-tags hubtest` для BE; Vitest для FE; один логический коммит на подкласс при необходимости
- Approval record: миграции Ent — только с явным согласованием владельца

## [REFERENCES]

- Реестр W12 (P3-BE-11,P3-BE-12)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P3-BE-11,P3-BE-12 с мутацией «да» закрыты по реестру
- **Мутация (C):** OG/Validate/CORS + shutdown/metrics probes
- Regression: соответствующий `testframework` / Vitest domain package
- Evidence: вывод тестов, SHA коммитов
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

- Коммит(ы) Conventional (тип/scope EN, описание RU) + путь задачи в теле
- Статус README

## [OPERATING RULES]

- Guard тестировать на реальном слое, не только mock
- Co-authored-by запрещён