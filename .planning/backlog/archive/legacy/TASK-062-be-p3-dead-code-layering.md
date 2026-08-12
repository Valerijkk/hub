# TASK-062 — Dead code / layering residuals

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P3** · Статус: Закрыта
> Источник: TASK-008 W12 · Classes: P3-BE-07 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P3-BACKEND.md`

## [OUTCOME]

Дефекты P3-BE-07 закрыты.

## [ROLE]

- Role: be2
- Responsibility: фикс + тесты на слое guard
- Handoff to: стадия C (мутация независимым агентом)

## [SCOPE AND CONSTRAINTS]

- Files allowed: backend/internal/**; `.planning/backlog/README.md`; этот файл
- Do not touch: несвязанные модули; secrets; DEV_MODE в CI
- Constraints: Docker + `-tags hubtest` для BE; Vitest для FE; один логический коммит на подкласс при необходимости
- Approval record: миграции Ent — только с явным согласованием владельца

## [REFERENCES]

- Реестр W12 (P3-BE-07)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P3-BE-07 с мутацией «да» закрыты по реестру
- **Мутация (C):** удалённый dead path → compile/test fail если ещё ссылаются
- Regression: соответствующий `testframework` / Vitest domain package
- Evidence: вывод тестов, SHA коммитов
- Stop when: green + статус «Закрыта» в README
- **C mutation (2026-07-30 / be2):** `MUTATION-2026-07-30-W13-be2-task062.md` — **7 PROVEN** · **6 GAP_NOTED** · **1 N/A** · **0 NOT_PROVEN** → Закрыта

## [OUTPUT SCHEMA]

- Коммит(ы) Conventional (тип/scope EN, описание RU) + путь задачи в теле
- Статус README

## [OPERATING RULES]

- Guard тестировать на реальном слое, не только mock
- Co-authored-by запрещён

## [MUTATION CRITERIA] — verifier-ready

> W13 slot 4 refill / qa-auto1 · criteria only (no mutate) · 2026-07-30 · HEAD guards verified · Class **P3-BE-07**

| ID | Production file:line (guard) | слой | Exact test RED | Kill recipe |
|----|------------------------------|------|----------------|-------------|
| W1-A1-008 | `backend/internal/tracker/handler_issues.go:277-278` — JQL через `h.service.SearchIssuesJQL` (не `GetDB`) | handler | *(GAP — нет RED на handler→Ent bypass)* · adjacent: `TestHandler_SearchIssuesJQL_Success` · `backend/testframework/tracker/handler_jql_flow_test.go` | handler: `GetDB().Issue.Query()` вместо SearchIssuesJQL → **GAP** (тест не краснеет на layering) |
| W1-A2-014 | `backend/internal/tracker/service_issues.go:378-380` — `LogActivity(…, "deleted", …)` перед soft-delete | service | `TestDeleteIssue_LogsDeletedActivity_Task062` · `backend/testframework/tracker/dead_code_layering_task062_test.go` | убрать LogActivity из DeleteIssue → deleted activity assert RED |
| W1-A2-015 | `backend/internal/tracker/service_workflow.go:76-78` — skip transition при nil From/ToStatus | service | `TestListWorkflowTransitions_SkipsNilEdges_Task062` · `backend/testframework/tracker/dead_code_layering_task062_test.go` | убрать nil-guard → panic/nil edge в list → RED |
| W2-A2-018 | `backend/internal/tms/service_runs.go:279-310` — dead `_ = result` удалён; reload via GetTestResult | service | *(GAP — нет assert отсутствия dead assign)* | восстановить `_ = result` → **GAP** (остаётся GREEN) |
| W2-A4-010 | `backend/internal/tms/testrail_import.go:291-292` — только `tryReadCSV`; dead `csv.NewReader` удалён | parser | *(GAP)* · adjacent `TestParseTestRailCSV` | восстановить unused `csv.NewReader` → **GAP** (compile/behavior green) |
| W2-A4-011 | `backend/internal/tms/testrail_import.go:87-90` — package-level `stepNumberPrefixRegexp` | parser | `TestTrimStepNumbering` · `backend/testframework/tms/coverage_wave_test.go` (behavior only) | MustCompile внутрь trim на каждый вызов → **GAP** для perf (output green) |
| W2-A5-007 | `backend/internal/bases/repository.go:214-239` — CreateTable tx: table + default Name column | repository (Ent) | `TestCreateTable_DefaultNameColumnInTx_Task062` · `backend/testframework/bases/create_table_tx_task062_test.go` | CreateColumn вне tx / без default Name → rollback/missing column RED |
| W3-A2-018 | `backend/internal/messenger/service_messages.go:265-267` — `source.ChannelID != channelID` → NotFound | service | `TestForwardMessage_SourceChannelMismatch_Task062` · `backend/testframework/messenger/forward_channel_scope_task062_test.go` | убрать channel scope check → forward чужого source ok → NotFound assert RED |
| W3-A3-009 | `backend/internal/messenger/presence.go:49-50` — godoc synced to 90s `presenceAwayThreshold` | docs | **N/A** | docs-only godoc; runtime unchanged |
| W4-A3-007 | `backend/internal/platform/observability/metrics_allowlist.go:22-37` + `prometheus.go:77` — `authorizeMetricsIP`; dead MetricsAllowMiddleware удалён | observability | `TestMetricsEndpoint_BlocksNonMatchingCIDR` · `backend/testframework/platform/observability/metrics_allowlist_test.go`; `TestMetricsEndpoint_DeniesOutsideAllowlist_WithProxyPeer` · `metrics_route_deny_task059_test.go` | skip authorizeMetricsIP в MetricsEndpoint → outside CIDR 200 → RED |
| W4-A4-010 | `backend/internal/platform/database/postgres.go` — `OpenRawDB` **deleted** | platform/database | *(compile-guard / GAP)* · comment `postgres_pool_p1be25_test.go:31` | восстановить `OpenRawDB()` без callers → **GAP** (нет RED на dead API) |
| W4-A5-009 | `backend/internal/apigateway/apikey.go:44,79-82` — dead `stopCh` удалён; `stopOnce`+`close(touchJobs)` | service | *(GAP — нет Close/lifecycle RED)* | восстановить unused `stopCh` → **GAP** |
| W6-A1-009 | `backend/internal/auth/repository.go:262-263` — `CreateOrganizationWithFounder(ctx, email)`; dead inviteToken param удалён | repository (Ent) | compile-guard: `TestEntRepository_OrganizationPendingFlow` · `backend/testframework/auth/repository_organization_test.go` | вернуть ignored `inviteToken` param → call sites compile fail |
| W7-A4-017 | `backend/internal/dashboard/service.go:302-312` `RequireSystemAdmin`; handler calls service (не repo) | service | `TestRequireSystemAdmin_Layering_Task062` · `backend/testframework/dashboard/admin_layering_task062_test.go` | non-admin → `return nil` вместо ErrForbidden → Forbidden assert RED |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Flag |
|----|-----------|------|
| W1-A1-008 | handler JQL flow (не layering assert) | **layer_mismatch** + **GAP** |
| W1-A2-014, W1-A2-015 | service + mockRepo | **mock≠Ent** |
| W2-A2-018, W2-A4-010, W2-A4-011 | — / parser behavior | **GAP** |
| W2-A5-007 | Ent CreateTable tx | **OK** |
| W3-A2-018 | service + mockRepository | **mock≠Ent** |
| W3-A3-009 | docs | **N/A** docs-only |
| W4-A3-007 | MetricsEndpoint CIDR | **OK** |
| W4-A4-010, W4-A5-009 | — | **GAP** |
| W6-A1-009 | Ent org create (compile signature) | **OK** (compile-guard) |
| W7-A4-017 | Ent RequireSystemAdmin | **OK** |

`layer_mismatch IDs: W1-A1-008` · `mock≠Ent IDs: W1-A2-014, W1-A2-015, W3-A2-018` · `docs-only IDs: W3-A3-009`

### Gaps

- Docs-only: **W3-A3-009** → N/A.
- Strong behavioral RED: **6/14** (014, 015, 007, 018, A3-007, 017).
- Compile-guard: W6-A1-009 (и потенциально W4-A4-010).
- GAP (no mutation-proof RED): W1-A1-008, W2-A2-018, W2-A4-010, W2-A4-011, W4-A4-010, W4-A5-009.
