# TASK-024 — Notif Nak, GitLab redact, GitHub RBAC, storage, backup, coffee

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P1** · Статус: На проверке  
> Источник: TASK-008 W12 · Classes: P1-BE-35..39,43 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P1-BACKEND.md`

## [OUTCOME]

Дефекты класса `P1-BE-35..39,43` закрыты: denial/regression на слое guard зелёные, критерий мутации воспроизводим.

## [ROLE]

- Role: be2
- Responsibility: фикс + тесты на слое guard
- Handoff to: стадия C (мутация независимым агентом)

## [SCOPE AND CONSTRAINTS]

- Files allowed: backend/internal/{notification,gitlab,github,storage,backup,intranet}/**; `.planning/backlog/README.md`; этот файл
- Do not touch: несвязанные модули; secrets; DEV_MODE в CI
- Constraints: Docker + `-tags hubtest` для BE; Vitest для FE; один логический коммит на подкласс при необходимости
- Approval record: миграции Ent — только с явным согласованием владельца

## [REFERENCES]

- Реестр W12 P1 BACKEND (ClassIDs P1-BE-35..39,43)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P1-BE-35..39,43 с мутацией «да» закрыты по реестру
- **Мутация (C):** repo err → Nak; secret var redact; stranger GitHub links 403; no S3 → MarkFailed
- Regression: соответствующий `testframework` / Vitest domain package
- Evidence: вывод тестов, SHA коммитов
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

- Коммит(ы) Conventional (тип/scope EN, описание RU) + путь задачи в теле
- Статус README

## [OPERATING RULES]

- Guard тестировать на реальном слое, не только mock
- Co-authored-by запрещён
