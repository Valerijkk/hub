# TASK-061 — Repo ProjectID/DeletedAt + list caps residual

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P3** · Статус: Закрыта
> Источник: TASK-008 W12 · Classes: P3-BE-05,P3-BE-06 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P3-BACKEND.md`
> Mutation C: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-be2-task061.md` · **11/11 PROVEN** · be2 · 2026-07-30

## [OUTCOME]

Дефекты P3-BE-05/06 закрыты.

## [ROLE]

- Role: be1
- Responsibility: фикс + тесты на слое guard
- Handoff to: стадия C (мутация независимым агентом)

## [SCOPE AND CONSTRAINTS]

- Files allowed: backend/internal/**; `.planning/backlog/README.md`; этот файл
- Do not touch: несвязанные модули; secrets; DEV_MODE в CI
- Constraints: Docker + `-tags hubtest` для BE; Vitest для FE; один логический коммит на подкласс при необходимости
- Approval record: миграции Ent — только с явным согласованием владельца

## [REFERENCES]

- Реестр W12 (P3-BE-05,P3-BE-06)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P3-BE-05,P3-BE-06 с мутацией «да» закрыты по реестру
- **Мутация (C):** без ProjectIDEQ → cross-project; без Limit → unbounded
- Regression: соответствующий `testframework` / Vitest domain package
- Evidence: вывод тестов, SHA коммитов
- Stop when: green + статус «Закрыта» в README (mutation 11/11 PROVEN)

## [OUTPUT SCHEMA]

- Коммит(ы) Conventional (тип/scope EN, описание RU) + путь задачи в теле
- Статус README

## [OPERATING RULES]

- Guard тестировать на реальном слое, не только mock
- Co-authored-by запрещён

## [MUTATION CRITERIA] — verifier-ready

> W13 slot 4 refill / qa-auto1 · criteria only (no mutate) · 2026-07-30 · HEAD guards verified · Classes **P3-BE-05**, **P3-BE-06**

| ID | Production file:line (guard) | слой | Exact test RED | Kill recipe |
|----|------------------------------|------|----------------|-------------|
| W1-A3-009 | `backend/internal/tracker/repo_automation.go:68-73` — `Update().Where(IDEQ, ProjectIDEQ, DeletedAtIsNil)` | repository (Ent) | `TestEntRepository_UpdateAutomationRule_RequiresProjectID_P3BE05` · `backend/testframework/tracker/repo_defense_task061_test.go` | убрать `ProjectIDEQ` → foreign projectID обновляет rule → `ent.IsNotFound` assert RED |
| W1-A4-013 | `backend/internal/sprint/repository.go:455-476` — ClearIssuesSprint + `HasProjectWith(project.IDEQ)` | repository (Ent) | `TestEntRepository_ClearIssuesSprint_RequiresProject_P3BE05` · `backend/testframework/sprint/clear_sprint_scoped_task061_test.go` | убрать `HasProjectWith` → Clear на foreign projectID очищает sprint → RED |
| W2-A3-009 | `backend/internal/tms/ent_repository.go:370-376` DeleteTestCase (`DeletedAtIsNil`+project); `:611-616` MoveSuite `HasProjectWith`; `:1307-1311` AddFilesToResult `HasRunWith` | repository (Ent) | `TestEntRepository_DeleteTestCase_IdempotentNotFound_P3BE05`; `…_MoveSuite_RequiresProject_P3BE05`; `…_AddFilesToResult_RequiresRun_P3BE05` · `backend/testframework/tms/repo_defense_task061_test.go` | снять DeletedAtIsNil / HasProjectWith / HasRunWith → 2-й delete / foreign move / wrong run ok → `ent.IsNotFound` RED |
| W2-A3-010 | `backend/internal/tms/ent_repository.go:1068-1078` — `entissue.DeletedAtIsNil()` в GetProjectIssuesWithTestCases | repository (Ent) | `TestEntRepository_GetProjectIssuesWithTestCases_SkipsSoftDeleted_P3BE05` · `backend/testframework/tms/repo_defense_task061_test.go` | снять `DeletedAtIsNil` → trash issue в coverage list → Empty assert RED |
| W3-A3-008 | `backend/internal/messenger/ent_repository.go:147-152` — `HasChannelWith(DeletedAtIsNil)` в GetChannelMemberIDs | repository (Ent) | `TestEntRepository_GetChannelMemberIDs_SkipsSoftDeleted_P3BE05` · `backend/testframework/messenger/member_ids_softdelete_task061_test.go` | снять HasChannelWith(DeletedAtIsNil) → после DeleteChannel members non-empty → RED |
| W3-A5-014 | `backend/internal/calls/repository.go:303-308` — `HasChannelWith(DeletedAtIsNil)` в GetChannelMemberRole | repository (Ent) | `TestRepository_GetChannelMemberRole_SkipsSoftDeletedChannel_P3BE05` · `backend/testframework/calls/member_role_softdelete_task061_test.go` | снять HasChannelWith(DeletedAtIsNil) → role после soft-delete channel → empty assert RED |
| W7-A4-014 | `backend/internal/storage/repository.go:186-190` — SoftDelete `DeletedAtIsNil` | repository (Ent) | `TestEntRepository_SoftDelete_IdempotentNotFound_P3BE05` · `backend/testframework/storage/softdelete_idempotent_task061_test.go` | UpdateOneID без DeletedAtIsNil → 2-й SoftDelete ok → `ent.IsNotFound` RED |
| W7-A5-011 | `backend/internal/marketplace/repository.go:217-221` — SoftDeleteListing `DeletedAtIsNil`+`VersionEQ` | repository (Ent) | `TestEntRepository_SoftDeleteListing_RequiresVersion_P3BE05` · `backend/testframework/marketplace/softdelete_version_task061_test.go` | снять VersionEQ / DeletedAtIsNil → stale/repeat delete ok → `ent.IsNotFound` RED |
| W1-A3-010 | `backend/internal/tracker/repo_activity.go:51` Limit(MaxIssueActivities); `repo_comments.go:93` MaxIssueComments; `repo_time.go:191` MaxIssueTimeLogs | repository (Ent) | `TestEntRepository_ListActivities_Caps_P3BE06`; `…_ListComments_Caps_P3BE06`; `…_ListTimeLogs_Caps_P3BE06` · `backend/testframework/tracker/list_caps_task061_test.go` | снять `.Limit(MaxIssue*)` при N>cap → `len > cap` RED |
| W2-A5-012 | `backend/internal/bases/repository.go:114` — `Limit(MaxOwnerBases)` | repository (Ent) | `TestEntRepository_ListBases_Caps_P3BE06` · `backend/testframework/bases/list_bases_cap_task061_test.go` | снять `.Limit(MaxOwnerBases)` → unbounded list RED |
| W7-A2-013 | `backend/internal/crm/repository.go:78` ListContacts; `:167` ListDeals — `Limit(MaxCRMList)` | repository (Ent) | `TestEntRepository_ListContactsAndDeals_Caps_P3BE06` · `backend/testframework/crm/list_cap_task061_test.go` | снять `.Limit(MaxCRMList)` → `len > MaxCRMList` RED |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Flag |
|----|-----------|------|
| W1-A3-009 … W7-A5-011 (P3-BE-05) | Ent repository (Postgres hubtest / sqlite enttest) | **OK** |
| W1-A3-010, W2-A5-012, W7-A2-013 (P3-BE-06) | Ent repository Limit caps | **OK** |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)` · `docs-only IDs: (none)`

### Gaps

- Docs-only: **N/A**.
- Class source IDs = **11/11** с dedicated RED (`*task061*`).
- Note: W2-A3-009 DeleteTestCase project-scope guard есть; dedicated foreign-project RED — через MoveSuite pattern.
