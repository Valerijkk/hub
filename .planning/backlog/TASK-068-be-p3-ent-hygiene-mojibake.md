# TASK-068 — Ent MaxLen/hygiene + mojibake comments

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P3** · Статус: Закрыта
> Источник: TASK-008 W12 · Classes: P3-BE-17,P3-BE-18 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P3-BACKEND.md`
> Mutation C: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-qa1-task068.md` · **11/11 PROVEN** · qa-auto1 · 2026-07-30

## [OUTCOME]

Дефекты P3-BE-17/18 закрыты.

## [ROLE]

- Role: be2
- Responsibility: фикс + тесты на слое guard
- Handoff to: стадия C (мутация независимым агентом)

## [SCOPE AND CONSTRAINTS]

- Files allowed: backend/ent/schema/**; `.planning/backlog/README.md`; этот файл
- Do not touch: несвязанные модули; secrets; DEV_MODE в CI
- Constraints: Docker + `-tags hubtest` для BE; Vitest для FE; один логический коммит на подкласс при необходимости
- Approval record: миграции Ent — только с явным согласованием владельца

## [REFERENCES]

- Реестр W12 (P3-BE-17,P3-BE-18)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P3-BE-17,P3-BE-18 с мутацией «да» закрыты по реестру
- **Мутация (C):** MaxLen violation; UTF-8 comments без mojibake
- Regression: соответствующий `testframework` / Vitest domain package
- Evidence: вывод тестов, SHA коммитов
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

- Коммит(ы) Conventional (тип/scope EN, описание RU) + путь задачи в теле
- Статус README

## [OPERATING RULES]

- Guard тестировать на реальном слое, не только mock
- Co-authored-by запрещён

## [MUTATION CRITERIA] — verifier-ready

> W13 slot 1 refill / be1 · criteria only (no mutate) · 2026-07-30 · HEAD guards verified · Classes **P3-BE-17**, **P3-BE-18**

| ID | Production file:line (guard) | слой | Exact test RED | Kill recipe |
|----|------------------------------|------|----------------|-------------|
| W8-A1-009 | `backend/ent/schema/apikey.go` / `call.go` / `channelinvite.go` — один Unique source (нет duplicate `index.Fields(…).Unique()`) | schema | `TestSchema_P3BE17_HygieneMarkers` · `backend/testframework/ent/schema_p3be17_18_task068_test.go` | вернуть `index.Fields("key_hash"|"room_name"|"token").Unique()` → NotContains RED |
| W8-A2-019 | `backend/ent/schema/comment.go:30` — `field.JSON("mentions", []uuid.UUID{})` | schema + Ent runtime | `TestComment_MentionsUUID_P3BE17` (+ markers) · `backend/testframework/ent/schema_p3be17_18_task068_test.go` | `[]string{}` / string mentions → type/Contains UUID marker RED |
| W8-A3-017 | `backend/ent/schema/invite.go:25` — `code` `.MaxLen(64)` | schema + Ent validation | `TestInvite_CodeMaxLen_P3BE17` · `backend/testframework/ent/schema_p3be17_18_task068_test.go` | убрать MaxLen → 65-char Create ok → `ent.IsValidationError` RED |
| W8-A3-018 | `backend/ent/schema/newspost.go:26` — `body` `.MaxLen(50000)` | schema + Ent validation | `TestNewsPost_BodyMaxLen_P3BE17` · `backend/testframework/ent/schema_p3be17_18_task068_test.go` | убрать MaxLen → 50001 Create ok → ValidationError RED |
| W8-A1-010 | `backend/ent/schema/apikey.go:52` (+ backuprecord/board godoc) — «Рёбра (edges)» | schema comments | `TestSchema_P3BE18_NoMojibake` · `backend/testframework/ent/schema_p3be17_18_task068_test.go` | вернуть «ибра» / битый mojibake → hits non-empty RED |
| W8-A2-020 | `backend/ent/schema/githubintegration.go` (+ githublink/gitlablink comments) | schema comments | same `TestSchema_P3BE18_NoMojibake` | inject mojibake in comments → hits RED |
| W8-A3-016 | `backend/ent/schema/issue.go` edges godoc | schema comments | same | «ибра» → hits RED |
| W8-A4-011 | `backend/ent/schema/project.go` / `projecttemplate.go` edges godoc | schema comments | same | «ибра» → hits RED |
| W8-A5-003 | `backend/ent/schema/testcase.go` Comment UTF-8 | schema comments | same | mojibake in Comment → hits RED |
| W8-A5-007 | `backend/ent/schema/testrun.go` allure_run_id Comment | schema comments | same | mojibake in Comment → hits RED |
| W8-A5-011 | `testcase.go` / `testcaseversion.go` / `testrun.go` / `userblock.go` / `webhookdelivery.go` / `workflowstatus.go` edges godoc | schema comments | same | «ибра» → hits RED |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Flag |
|----|-----------|------|
| W8-A1-009 | source-read schema markers | **OK** (schema, не mock) |
| W8-A2-019 | Ent Create Comment + source marker | **OK** |
| W8-A3-017 / W8-A3-018 | Ent validation (sqlite enttest) | **OK** |
| W8-A1-010…W8-A5-011 | filesystem grep schema/*.go | **OK** — schema comments hygiene, не runtime Ent mock |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)` · `docs-only IDs: (none)` — schema comments mutatable; не product docs/

### Gaps

- Docs-only (wiki/vision): **N/A**.
- Class source IDs = **11/11** с RED (P3-BE-17: 4 runtime/source; P3-BE-18: 7 via shared mojibake grep).
