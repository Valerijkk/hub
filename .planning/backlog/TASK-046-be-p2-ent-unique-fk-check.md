# TASK-046 — Ent unique/FK/CHECK (без VersionMixin ADR)

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · На проверке
> Источник: TASK-008 W12 · Classes: P2-BE-22 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P2-BACKEND.md`

## [OUTCOME]

Дефекты P2-BE-22 закрыты. U02/U03 — после ADR.

## [ROLE]

- Role: be1
- Responsibility: фикс + тесты на слое guard
- Handoff to: стадия C (мутация независимым агентом)

## [SCOPE AND CONSTRAINTS]

- Files allowed: backend/ent/schema/**; backend/ent/migrate/**; `.planning/backlog/README.md`; этот файл
- Do not touch: несвязанные модули; secrets; DEV_MODE в CI
- Constraints: Docker + `-tags hubtest` для BE; Vitest для FE; один логический коммит на подкласс при необходимости
- Approval record: миграции Ent — только с явным согласованием владельца

## [REFERENCES]

- Реестр W12 (P2-BE-22)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P2-BE-22 с мутацией «да» закрыты по реестру
- **Мутация (C):** duplicate (project,name)→unique; XOR debit/credit CHECK
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

enttest Component unique + JournalLine XOR CHECK runtime.

## [W18-C3 MUTATION — 2026-07-28 be2]

**FAIL.** Evidence: `.planning/audits/total-2026-07/W18-C3-MUTATION.md`

- `TestSchema_P2BE22_UniqueFKCheck` RED when Unique/CHECK strings removed from schema source — but test is **source grep only**, not Ent-runtime duplicate Create / XOR insert.
- Insufficient vs ACCEPTANCE «duplicate (project,name)→unique; XOR debit/credit CHECK». Need enttest like TASK-025.

## [W1-be2 MUTATION — 2026-07-29]

**PASS.** Evidence: .planning/audits/total-2026-07/MUTATION-2026-07-29-W1-be2.md

- migrate/schema.go Unique false / drop XOR Checks → TestComponent_UniqueNamePerProject_P2BE22, TestJournalLine_DebitXorCredit_P2BE22 RED
- GREEN after restore. Runtime enttest (not schema grep).
