# TASK-057 — Silent error swallow (_=/Count→0)

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P3** · Статус: Закрыта
> Мутация C: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-be1-task057.md` · 5 PROVEN (+2 GAP_NOTED mock≠Ent) · 2 GAP · 2026-07-30
> Источник: TASK-008 W12 · Classes: P3-BE-01 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P3-BACKEND.md`

## [OUTCOME]

Дефекты P3-BE-01 закрыты.

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

- Реестр W12 (P3-BE-01)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P3-BE-01 с мутацией «да» закрыты по реестру
- **Мутация (C):** вернуть err.Error в log → тест slog; убрать slog → silent снова
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

> W13 slot 4 refill / qa-auto1 · criteria only (no mutate) · 2026-07-30 · HEAD guards verified · Class **P3-BE-01**

| ID | Production file:line (guard) | слой | Exact test RED | Kill recipe |
|----|------------------------------|------|----------------|-------------|
| W1-A3-008 | `backend/internal/tracker/repository.go:290-294` — `logOutboxPublishErr` → `slog.WarnContext(…, "outbox publish failed", …)` на post-commit Publish err | repository (Ent + outbox) | `TestUpdateIssue_PublishFail_LogsWarn_TASK057` · `backend/testframework/tracker/outbox_publish_slog_task057_test.go` | `_ = Publish` / убрать slog → лог без `"outbox publish failed"` → slog assert RED |
| W1-A4-012 | `backend/internal/sprint/service.go:72-75` — `GetSprintIssueCounts` err → `NewInternal("…счётчиков…")` (не empty map) | service | `TestListSprints_CountsError_FailClosed_TASK057` · `backend/testframework/sprint/list_fail_closed_task057_test.go` | silent `counts=make(map…)` → `err==nil` + zeros → Internal/500 assert RED |
| W1-A4-010 | `backend/internal/sprint/repository.go:568-578` — velocity `Count` err → `return nil, fmt.Errorf(…)` (не `committed=0`) | repository / Ent | *(нет dedicated TASK-057 RED)* | вернуть `_ , _ = Count` / zero on err → velocity zeros без err → **GAP** (нужен Ent Count-fail inject) |
| W3-A3-010 | `backend/internal/messenger/unread.go:79-81` — ParseInt fail → `slog.WarnContext(…, "unread counter corrupt", …)` + skip field | service (Redis) | `TestGetAll_CorruptCounter_LogsWarn_TASK057` · `backend/testframework/messenger/unread_corrupt_task057_test.go` | `n,_:=ParseInt` без slog / coerce 0 → лог пуст / `ch-bad==0` → assert RED |
| W2-A4-013 | `backend/internal/tms/exporter.go:172-188` — CSV `Write`/`Flush`/`w.Error()` → `fmt.Errorf("…CSV…")` | exporter (pure) | `TestExportRunResultsToCSV_WriteError_TASK057` · `backend/testframework/tms/exporter_write_err_task057_test.go` | `_ = w.Write` / ignore Flush err → `err==nil` на broken writer → RED |
| W6-A4-015 | `backend/internal/audit/service.go:456-459` — `CreateAuditLog` fail после ban → `return fmt.Errorf("audit ban log: %w", err)` | service | `TestBanUser` (createErr branch) · `backend/testframework/audit/service_security_test.go` | swallow CreateAuditLog → BanUser `err==nil` без trail → fail-closed assert RED |
| W7-A5-007 | `backend/internal/marketplace/service.go:287-291` — `IsSystemAdmin` err → `fmt.Errorf("check system admin: %w", err)` (не Forbidden) | service | *(нет dedicated TASK-057 RED)* | mask as `ErrNotOwner`/Forbidden → 403 вместо 500 → **GAP** |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Flag |
|----|-----------|------|
| W1-A3-008 | Ent UpdateIssue + failing outbox Publish + slog capture | **OK** — Ent repo path |
| W1-A4-012 | SprintService + mock `GetSprintIssueCounts` err | **mock≠Ent** (GAP_NOTED): fail-closed на service; не Ent GROUP BY IT |
| W1-A4-010 | — | **GAP** + intended Ent repo layer |
| W3-A3-010 | UnreadService + miniredis corrupt HSET | **OK** — Redis path (не Ent) |
| W2-A4-013 | pure exporter + errWriter | **OK** — no DB |
| W6-A4-015 | AuditService + mockRepo createErr | **mock≠Ent** (GAP_NOTED) |
| W7-A5-007 | — | **GAP** |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: W1-A4-012, W6-A4-015` · `docs-only IDs: (none)`

### Gaps

- Docs-only: **N/A**.
- Runnable now: **5/7** listed (008, 012, 010-unread, 013, 015); **GAP** W1-A4-010 velocity Count + W7-A5-007 admin-check err.
- Class residuals без dedicated TASK-057 RED (slog/fail-closed в коде есть): W1-A2-013, W1-A4-011, W1-A5-009/010, W2-A2-020, W5-A2-010, W6-A4-013, W7-A1-010, W7-A2-015, W7-A3-011, W7-A4-019, W7-A5-015 — **GAP_NOTED** (mutate приоритет — runnable 5).
