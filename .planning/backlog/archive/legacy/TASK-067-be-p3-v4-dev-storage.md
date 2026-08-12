# TASK-067 — V4 storage/dashboard + dev login/seed

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P3** · Статус: Закрыта
> Mutation C (W13 be1): 9/10 PROVEN · 1 GAP_NOTED (W7-A2-012) · evidence `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-be1-task067.md`
> Источник: TASK-008 W12 · Classes: P3-BE-15,P3-BE-16 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P3-BACKEND.md`

## [OUTCOME]

Дефекты P3-BE-15/16 закрыты.

## [ROLE]

- Role: be1
- Responsibility: фикс + тесты на слое guard
- Handoff to: стадия C (мутация независимым агентом)

## [SCOPE AND CONSTRAINTS]

- Files allowed: backend/internal/{intranet,dashboard,storage,auth}/**; `.planning/backlog/README.md`; этот файл
- Do not touch: несвязанные модули; secrets; DEV_MODE в CI
- Constraints: Docker + `-tags hubtest` для BE; Vitest для FE; один логический коммит на подкласс при необходимости
- Approval record: миграции Ent — только с явным согласованием владельца

## [REFERENCES]

- Реестр W12 (P3-BE-15,P3-BE-16)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P3-BE-15,P3-BE-16 с мутацией «да» закрыты по реестру
- **Мутация (C):** coffee PII; layout ACL; dev login only !release
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

> W13 slot 1 refill / be1 · criteria only (no mutate) · 2026-07-30 · HEAD guards verified · Classes **P3-BE-15**, **P3-BE-16**

| ID | Production file:line (guard) | слой | Exact test RED | Kill recipe |
|----|------------------------------|------|----------------|-------------|
| W7-A2-012 | `backend/testframework/intranet/handler_test.go:23` — uuid-suffixed SQLite DSN | testframework | **GAP** — нет dedicated assert на uuid в DSN (self-guard) | `file:intranet?mode=memory&cache=shared` без uuid → parallel flaky share |
| W7-A2-014 | `backend/internal/intranet/coffee_dto.go:18-22` + `coffee_repository.go:178-193` — `CoffeePartnerResponse` без email | dto / repository | `TestCoffeePartner_NoEmailPII_Task067` · `backend/testframework/intranet/v4_defense_task067_test.go` | вернуть `Email` в DTO/`toCoffeePartner` → JSON содержит `"email"` / secret@ → NotContains RED |
| W7-A2-016 | `backend/internal/intranet/repo_polls.go:84-105` — SoftDeleteNewsPost + AuditLog в tx | repository (Ent) | `TestDeleteNews_WritesAudit_Task067` · `backend/testframework/intranet/v4_defense_task067_test.go` | убрать AuditLog.Create → Count=0 → Equal(1) RED |
| W7-A3-013 | `backend/internal/imex/repository.go:70-84` — default status `IsDefaultEQ(true)` (+ SortOrder fallback) | repository (Ent) | `TestAllocateIssueKey_PrefersDefaultStatus_Task067` · `backend/testframework/imex/default_status_task067_test.go` | вернуть `First()` без IsDefault → может вернуть Done → Equal(todo.ID) RED |
| W7-A4-015 | `backend/internal/storage/orphan_cleanup.go:32-73` — drain loop `for { List…; if <batch break }` | service | `TestOrphanCleanup_DrainsBeyondBatch_Task067` · `backend/testframework/storage/orphan_drain_task067_test.go` | один батч без loop → хвост остаётся → Get orphan Error assert RED |
| W7-A4-016 | `backend/internal/dashboard/service.go:438-443` ApplyTemplate; `:499-505` SaveLayout — `IsProjectMember` | service (Ent) | `TestApplyTemplate_StrangerForbidden_Task067`; `TestSaveLayout_StrangerProjectForbidden_Task067` · `backend/testframework/dashboard/layout_acl_task067_test.go` | убрать membership check → stranger OK → `errors.Is(ErrForbidden)` RED |
| W7-A4-018 | `backend/internal/backup/repository.go:64-67` — `SetRestoredAt` (не ErrorMessage) | repository (Ent) | `TestEntRepository_SetRestoredAtAndDelete` · `backend/testframework/backup/repository_test.go` | `SetErrorMessage(note)` вместо RestoredAt → ErrorMessage non-nil → Nil assert RED |
| W7-A5-012 | `backend/internal/dev/handler.go:110-113` — `/api/dev/login` требует `StatusActive` | handler (Ent) | `TestLogin_RejectsNonActive_Task067` · `backend/testframework/dev/login_status_task067_test.go` | убрать StatusActive check → 200+JWT → StatusForbidden RED |
| W7-A5-013 | `backend/internal/dev/seed.go:194-198` — `partial` → 500 + `partial:true`; helpers fail-propagate | handler / seed | `TestSeedWorkflowTransitions_CreateFailPropagates_A3W75002`; `TestSeedIssueLinks_CreateFailPropagates_A3W75002`; `TestSeedChannels_CreateFailPropagates_A3W75002` · `backend/testframework/dev/seed_partial_a3w75_task166_test.go` | helpers `return n, nil` на err / убрать partial→500 → Seed 200 при битых шагах → Error assert RED |
| W7-A5-014 | `backend/internal/dev/seed_data.go:107-108`, `:168-172` — demo Telegram = notification bind | seed data | `TestSeedData_NoTelegramLoginCopy_Task067` · `backend/testframework/dev/login_status_task067_test.go` | вернуть «Telegram вход/авторизация» в titles → NotContains RED |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Flag |
|----|-----------|------|
| W7-A2-012 | testframework DSN hygiene | **GAP** — no dedicated RED |
| W7-A2-014 | ExportToCoffeePartner + Ent User | **OK** |
| W7-A2-016 | Service.DeleteNews → Ent SoftDelete+Audit | **OK** |
| W7-A3-013 | Ent AllocateIssueKey | **OK** |
| W7-A4-015 | OrphanCleanupService + Ent FileMetadata | **OK** |
| W7-A4-016 | DashboardService + Ent membership | **OK** |
| W7-A4-018 | Ent backup SetRestoredAt | **OK** |
| W7-A5-012 | Fiber dev handler + Ent user | **OK** |
| W7-A5-013 | seed helpers (export) — не полный HTTP Seed | **OK** / residual A3W7-5-002 covered |
| W7-A5-014 | pure ExportDemoIssues/TestCases | **OK** |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)` · `docs-only IDs: (none)`

### Gaps

- Docs-only: **N/A**.
- Class source IDs = **10/10** listed; **9/10** с dedicated RED (**W7-A2-012** GAP — self-guard only).
- W7-A5-013: RED на helper fail-propagate (+ outer `seed.go:194-198` guard); полный HTTP Seed `partial` body — желателен twin.
