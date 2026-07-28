# TASK-025 — Ent: coffee unique, notif pref partial unique, journal↔invoice O2O

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P1** · Статус: Исправлена  

> Источник: TASK-008 W12 · Classes: P1-BE-40..42 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P1-BACKEND.md`

## [OUTCOME]

Дефекты класса `P1-BE-40..42` закрыты: denial/regression на слое guard зелёные, критерий мутации воспроизводим.

## [ROLE]

- Role: be1
- Responsibility: фикс + тесты на слое guard
- Handoff to: стадия C (мутация независимым агентом)

## [SCOPE AND CONSTRAINTS]

- Files allowed: backend/ent/schema/**; migrations as needed; `.planning/backlog/README.md`; этот файл
- Do not touch: несвязанные модули; secrets; DEV_MODE в CI
- Constraints: Docker + `-tags hubtest` для BE; Vitest для FE; один логический коммит на подкласс при необходимости
- Approval record: миграции Ent — только с явным согласованием владельца

## [REFERENCES]

- Реестр W12 P1 BACKEND (ClassIDs P1-BE-40..42)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P1-BE-40..42 с мутацией «да» закрыты по реестру
- **Мутация (C):** двойной enroll → unique; два NULL project pref → одна строка; вторая Invoice → constraint
- Regression: соответствующий `testframework` / Vitest domain package
- Evidence: вывод тестов, SHA коммитов
- Stop when: green + статус «Исправлена» в README

## [IMPLEMENTATION NOTE — 2026-07-28 be1]

Схемы обновлены:

- `CoffeeEnrollment`: `index.Edges("user").Unique()` + `User.To("coffee_enrollment").Unique()`
- `NotificationPreference`: два partial unique (`project_id IS NULL` / `IS NOT NULL`)
- `JournalEntry.To("invoice").Unique()` + `Invoice` `index.Edges("journal_entry").Unique()`

Миграция (additive SQL): `backend/migrations/20260728_task025_ent_schema_constraints.sql`.
**Prod apply — только с явным согласованием владельца** (pre-check дубликатов в файле).
Dev/test: Ent `Schema.Create` / `enttest` поднимает индексы на чистой БД.

Тесты мутации: `*_task025_test.go` в intranet / notification / accounting.

## [OUTPUT SCHEMA]

- Коммит(ы) Conventional (тип/scope EN, описание RU) + путь задачи в теле
- Статус README

## [OPERATING RULES]

- Guard тестировать на реальном слое, не только mock
- Co-authored-by запрещён

## [W18-C3 MUTATION — 2026-07-28 be2]

**PASS.** Evidence: `.planning/audits/total-2026-07/W18-C3-MUTATION.md`

- Mutated runtime Unique in `ent/migrate/schema.go` (generated DDL used by enttest).
- Coffee / notif partial unique / invoice O2O → respective TASK025 tests RED then GREEN. Ent-runtime layer.
