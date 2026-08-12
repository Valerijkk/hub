# TASK-059 — Mutation-proof gaps к P1/P2 фиксам

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P3** · Статус: **Закрыта**
> Источник: TASK-008 W12 · Classes: P3-BE-03 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P3-BACKEND.md`
> Mutation C: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-be2-task059.md` · **9/10 PROVEN** · W1-A4-015 **GAP_NOTED** (mock≠Ent) · be2 W13 slot 3 · 2026-07-30

## [OUTCOME]

Дефекты P3-BE-03 закрыты.

## [ROLE]

- Role: be1
- Responsibility: фикс + тесты на слое guard
- Handoff to: стадия C (мутация независимым агентом)

## [SCOPE AND CONSTRAINTS]

- Files allowed: backend/testframework/**; `.planning/backlog/README.md`; этот файл
- Do not touch: несвязанные модули; secrets; DEV_MODE в CI
- Constraints: Docker + `-tags hubtest` для BE; Vitest для FE; один логический коммит на подкласс при необходимости
- Approval record: миграции Ent — только с явным согласованием владельца

## [REFERENCES]

- Реестр W12 (P3-BE-03)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P3-BE-03 с мутацией «да» закрыты по реестру
- **Мутация (C):** снять guard P1/P2 → новый denial-тест краснеет
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

> W13 slot 4 refill / qa-auto1 · criteria only (no mutate) · 2026-07-30 · HEAD guards verified · Class **P3-BE-03**
> Kill = снять P1/P2 guard → denial/Ent/HTTP-negative RED.

| ID | Production file:line (guard) | слой | Exact test RED | Kill recipe |
|----|------------------------------|------|----------------|-------------|
| W1-A4-015 | `backend/internal/sprint/repository.go:419-466` — `SetIssuesSprint` missing/foreign issue → `ErrIssueNotInProject`; service maps conflict | repository (+ service) | `TestAddIssuesToSprint_RejectsMissingIssueMapping_W1A4015` · `backend/testframework/sprint/add_issues_missing_task059_test.go` | убрать mapping check / always Update → add foreign UUID `err==nil` → denial RED |
| W2-A5-009 | `backend/internal/bases/repository.go` — table via `HasBaseWith(hubbase.ID(…))` | repository / Ent | `TestGetTableView_CrossBaseTable_NotFound` · `backend/testframework/bases/cross_base_table_task059_test.go` | убрать HasBaseWith → cross-base table 200 → NotFound assert RED |
| W4-A2-007 | `backend/internal/platform/outbox/publisher.go:20-24` — `PublishInTx` внутри Ent tx; Rollback → 0 rows | outbox / Ent | `TestPublisher_PublishInTx_RollbackLeavesNoEvents` · `backend/testframework/platform/outbox/publish_rollback_task059_test.go` | Publish вне tx / commit on rollback → leftover rows → count=0 RED |
| W4-A3-009 | `backend/internal/platform/observability/route_label.go:12-19` — `httpRouteLabel` → template / `unmatched` (не raw UUID path) | observability | `TestHTTPMetricsMiddleware_RouteIsTemplate_NotRawUUID` · `backend/testframework/platform/observability/metrics_route_deny_task059_test.go` | label=`c.Path()` → UUID in metrics → template assert RED |
| W4-A3-010 | metrics endpoint allowlist + proxy peer deny | observability / HTTP | `TestMetricsEndpoint_DeniesOutsideAllowlist_WithProxyPeer` · `metrics_route_deny_task059_test.go` | relax CIDR/peer check → 200 from denied peer → deny assert RED |
| W5-A4-012 | search JQL OR → Ent union predicates (не AND-only) | search / Ent | `TestToEntPredicates_ORUnionVsAND_EntCount_W5A4012` · `backend/testframework/search/jql_or_ent_task059_test.go` | OR→AND flatten → union count mismatch → Ent count RED |
| W5-A1-011 | `backend/internal/ai/service.go:630-642` — `toProviderResponse` → `HasAPIKey`, **нет** raw `api_key` | service / handler DTO | `TestUserHandler_CreateProvider_RedactsAPIKey_W5A1011` · `backend/testframework/ai/api_key_redact_audit_task059_test.go` | сериализовать `APIKey`/`sk-*` в JSON → redact assert RED |
| W5-A3-019 | audit `LogAIChat` details — metadata only (без prompt/response) | audit / AI | `TestAuditRepoLogger_LogAIChat_NoPromptInDetails_W5A3019` · `api_key_redact_audit_task059_test.go` | добавить prompt в details → PII assert RED |
| W6-A5-006 | `backend/internal/audit/repository.go:450-462` — `Restore*` → `ClearDeletedAt()` | repository / Ent | `TestRestoreIssue_ClearsDeletedAt_W6A5006` (+ TestCase/File) · `backend/testframework/audit/restore_deleted_at_task059_test.go` | Restore без ClearDeletedAt → `DeletedAt!=nil` → nil assert RED |
| W7-A5-016 | marketplace `DeleteListing` owner/admin check → HTTP 403 stranger | handler (+ service) | `TestHandler_DeleteListing_Forbidden_W7A5016` · `backend/testframework/marketplace/delete_forbidden_task059_test.go` | убрать owner/admin check → 204 stranger → 403 assert RED |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Flag |
|----|-----------|------|
| W1-A4-015 | SprintService + mock issueProject map | **mock≠Ent** (GAP_NOTED): denial на service/mock; Ent `SetIssuesSprint` IT — sibling |
| W2-A5-009 | Ent HasBaseWith cross-base | **OK** — Ent |
| W4-A2-007 | Ent tx Rollback outbox | **OK** — Ent |
| W4-A3-009/010 | Fiber metrics middleware | **OK** — HTTP |
| W5-A4-012 | Ent predicate count OR vs AND | **OK** — Ent |
| W5-A1-011 / W5-A3-019 | AI handler + audit details | **OK** — Ent fixtures |
| W6-A5-006 | Ent ClearDeletedAt | **OK** — Ent |
| W7-A5-016 | Fiber DELETE + service RBAC | **OK** — HTTP (не только service) |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: W1-A4-015` · `docs-only IDs: (none)`

### Gaps

- Docs-only: **N/A**.
- Runnable dedicated TASK-059 RED: **10** IDs выше.
- Class residuals без `*_task059_test.go` (poker RBAC / soft-delete / calls recordings / AI stream denials W5-A2-013…017 / W5-A3-016…021 / W3-A3-012 / W3-A5-015-016 / W4-A1-010 / W4-A5-010-011 / W5-A5-007 / W6-A5-010 / W1-A4-014): покрытие может жить в других TASK/файлах — **GAP_NOTED** для mutate этой TASK пока нет явного RED path здесь.
