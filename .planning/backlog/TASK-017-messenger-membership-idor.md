# TASK-017 — Messenger: membership/owner scope на mutate

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P1** · Статус: Закрыта
> Источник: TASK-008 W12 · Classes: P1-BE-15..20 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P1-BACKEND.md`

## [OUTCOME]

Дефекты класса `P1-BE-15..20` закрыты: denial/regression на слое guard зелёные, критерий мутации воспроизводим.

## [ROLE]

- Role: be1
- Responsibility: фикс + тесты на слое guard
- Handoff to: стадия C (мутация независимым агентом)

## [SCOPE AND CONSTRAINTS]

- Files allowed: backend/internal/messenger/**; testframework/messenger/**; `.planning/backlog/README.md`; этот файл
- Do not touch: несвязанные модули; secrets; DEV_MODE в CI
- Constraints: Docker + `-tags hubtest` для BE; Vitest для FE; один логический коммит на подкласс при необходимости
- Approval record: миграции Ent — только с явным согласованием владельца

## [REFERENCES]

- Реестр W12 P1 BACKEND (ClassIDs P1-BE-15..20)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P1-BE-15..20 с мутацией «да» закрыты по реестру
- **Мутация (C):** non-member reaction/pin/bookmark/poll; B удаляет bookmark A → 404
- Regression: соответствующий `testframework` / Vitest domain package
- Evidence: вывод тестов, SHA коммитов
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

- Коммит(ы) Conventional (тип/scope EN, описание RU) + путь задачи в теле
- Статус README

## [OPERATING RULES]

- Guard тестировать на реальном слое, не только mock
- Co-authored-by запрещён

## Stage C (W18-C1 mutation) — PASS

- Mutation: drop `UserIDEQ` in `RemoveBookmark` (Ent)
- RED: `TestRemoveBookmark_RejectsForeignOwner_P1BE15` — foreign delete succeeded
- GREEN: after restore
- Mock-only?: no (Ent layer)
