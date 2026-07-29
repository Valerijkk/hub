# TASK-044 — AI stream/privacy + search JQL/SSRF

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · На проверке
> Источник: TASK-008 W12 · Classes: P2-BE-17,P2-BE-18 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P2-BACKEND.md`

## [OUTCOME]

Дефекты P2-BE-17/18 закрыты.

## [ROLE]

- Role: be1
- Responsibility: фикс + тесты на слое guard
- Handoff to: стадия C (мутация независимым агентом)

## [SCOPE AND CONSTRAINTS]

- Files allowed: backend/internal/{ai,search}/**; `.planning/backlog/README.md`; этот файл
- Do not touch: несвязанные модули; secrets; DEV_MODE в CI
- Constraints: Docker + `-tags hubtest` для BE; Vitest для FE; один логический коммит на подкласс при необходимости
- Approval record: миграции Ent — только с явным согласованием владельца

## [REFERENCES]

- Реестр W12 (P2-BE-17,P2-BE-18)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P2-BE-17,P2-BE-18 с мутацией «да» закрыты по реестру
- **Мутация (C):** cancel→stream stop; AllowExternalAI default false; OR hits; Meili SSRF blocked
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

Meili SSRF hubtest + stream cancel mid-read IT. Status → Исправлена.

## [W18-C3 MUTATION — 2026-07-28 be2]

**FAIL.** Evidence: `.planning/audits/total-2026-07/W18-C3-MUTATION.md`

- AllowExternalAI default false→true → `TestGetUserPreference_DefaultAllowExternalAIFalse` RED ✓
- Drop `issue.Or` wrap → `TestToEntPredicates_ORBuildsUnion` RED ✓
- **Fail:** no hubtest for Meili SSRF (`MEILI_HOST=http://169.254.169.254` → nil indexer)
- **Fail:** no hubtest for stream cancel mid-read (`adapter_stream` ctx.Err)

## [W1-be2 MUTATION — 2026-07-29]

**FAIL.** Evidence: .planning/audits/total-2026-07/MUTATION-2026-07-29-W1-be2.md

- Meili SSRF: validateMeiliHost allow-all → TestValidateMeiliHost_BlocksMetadata_P2BE18 / TestNewMeiliIndexerFromEnv_RejectsMetadataHost_P2BE18 RED
- Stream cancel: drop ctx.Err mid-loop → TestChatCompletionStream_CancelMidRead_P2BE17 stays GREEN (HTTP request ctx). Gap → Открыта.
