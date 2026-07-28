# TASK-040 — Fail-closed defaults + rate-limit ZSET

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: Исправлена  
> Источник: TASK-008 W12 · Classes: P2-BE-10 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P2-BACKEND.md`

## [OUTCOME]

Дефекты P2-BE-10 закрыты.

## [ROLE]

- Role: be1
- Responsibility: фикс + тесты на слое guard
- Handoff to: стадия C (мутация независимым агентом)

## [SCOPE AND CONSTRAINTS]

- Files allowed: backend/internal/{ws,auth,platform,messenger,calls}/**; `.planning/backlog/README.md`; этот файл
- Do not touch: несвязанные модули; secrets; DEV_MODE в CI
- Constraints: Docker + `-tags hubtest` для BE; Vitest для FE; один логический коммит на подкласс при необходимости
- Approval record: миграции Ent — только с явным согласованием владельца

## [REFERENCES]

- Реестр W12 (P2-BE-10)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P2-BE-10 с мутацией «да» закрыты по реестру
- **Мутация (C):** Redis fail→deny; authz==nil→refuse; 429 не раздувает ZSET
- Regression: соответствующий `testframework` / Vitest domain package
- Evidence: вывод тестов, SHA коммитов
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

- Коммит(ы) Conventional (тип/scope EN, описание RU) + путь задачи в теле
- Статус README

## [OPERATING RULES]

- Guard тестировать на реальном слое, не только mock
- Co-authored-by запрещён

## [W18-C3 MUTATION — 2026-07-28 be2]

**PASS.** Evidence: `.planning/audits/total-2026-07/W18-C3-MUTATION.md`

- authz==nil→allowAll → `TestWSHandler_NilAuthorizer_RefusesUpgrade` RED
- skip ZRem on 429 → `TestRateLimitRedis_RejectedDoesNotInflateZSET` RED (card 7≠2)
- settings fail→open → `TestCheckRegistrationAllowed_SettingsFail_InviteOnly` RED  
Restore GREEN.
