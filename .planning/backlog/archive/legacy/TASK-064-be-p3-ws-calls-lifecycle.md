# TASK-064 — WS/calls lifecycle residuals

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P3** · Статус: Закрыта
> Источник: TASK-008 W12 · Classes: P3-BE-10 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P3-BACKEND.md`
> Mutation C (W13 slot 4 refill / qa-auto1 · 2026-07-30): **5/5 PROVEN + 1 N/A** — `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-qa1-task064.md`

## [OUTCOME]

Дефекты P3-BE-10 закрыты.

## [ROLE]

- Role: be2
- Responsibility: фикс + тесты на слое guard
- Handoff to: стадия C (мутация независимым агентом)

## [SCOPE AND CONSTRAINTS]

- Files allowed: backend/internal/{ws,calls}/**; `.planning/backlog/README.md`; этот файл
- Do not touch: несвязанные модули; secrets; DEV_MODE в CI
- Constraints: Docker + `-tags hubtest` для BE; Vitest для FE; один логический коммит на подкласс при необходимости
- Approval record: миграции Ent — только с явным согласованием владельца

## [REFERENCES]

- Реестр W12 (P3-BE-10)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P3-BE-10 с мутацией «да» закрыты по реестру
- **Мутация (C):** rooms cleanup; nil-RBAC fail-closed
- Regression: соответствующий `testframework` / Vitest domain package
- Evidence: вывод тестов, SHA коммитов
- Stop when: green + статус «Исправлена» в README

### Residual

- **W3-A4-006** JWT `?token=` → short-lived WS ticket: отложено (реестр: «после product»). Access JWT в query остаётся до ADR/product ticket exchange.

## [OUTPUT SCHEMA]

- Коммит(ы) Conventional (тип/scope EN, описание RU) + путь задачи в теле
- Статус README

## [OPERATING RULES]

- Guard тестировать на реальном слое, не только mock
- Co-authored-by запрещён

## [MUTATION CRITERIA] — verifier-ready

> W13 slot 4 refill / qa-auto1 · criteria only (no mutate) · 2026-07-30 · HEAD guards verified · Class **P3-BE-10**

| ID | Production file:line (guard) | слой | Exact test RED | Kill recipe |
|----|------------------------------|------|----------------|-------------|
| W3-A4-006 | `backend/internal/ws/handler.go:157-160` — `?token=` → 401 (`token query parameter is not allowed`) | handler (WS upgrade) | **N/A** — deferred Residual («после product»); adjacent вне слота: `TestWSHandler_QueryToken_Denied` · `backend/testframework/ws/handler_auth_test.go` | **N/A** (product deferred JWT ticket exchange) |
| W3-A4-007 | `backend/internal/ws/hub.go:330-331` — empty room → `h.rooms.Delete(room)` | WS Hub | `TestHub_EmptyRoomDeletedFromMap_TASK064`; `TestHub_EmptyRoomDeletedOnUnregister_TASK064` · `backend/testframework/ws/lifecycle_task064_test.go` | убрать `rooms.Delete` при empty → ExportRoomExists true → RED |
| W3-A4-008 | `backend/internal/ws/bridge.go:53-55` — Start err → `b.Stop()` rollback | WS Bridge | `TestBridge_StartFail_CallsStop_TASK064` · `backend/testframework/ws/lifecycle_task064_test.go` | на err подписки не вызывать Stop → ExportBridgeSubsLen > 0 → RED |
| W3-A4-009 | `backend/internal/ws/handler.go:252-267` — deny/authz err → TypeError frame | WS readPump | `TestReadPump_DeniedSubscribe_SendsErrorFrame_TASK064`; `TestReadPump_AuthorizerError_SendsErrorFrame_TASK064` · `backend/testframework/ws/lifecycle_task064_test.go` | убрать enqueueServerMessage на deny → timeout без error frame → RED |
| W3-A4-010 | `backend/internal/ws/export_hooks.go:88-96` — ExportClientRooms копия под `c.mu` | WS Client | `TestExportClientRooms_CopyUnderMu_TASK064` · `backend/testframework/ws/lifecycle_task064_test.go` | вернуть `c.rooms` без copy/mu → mutation видна в snap → RED |
| W3-A5-013 | `backend/internal/calls/handler_recording.go:129-132` list; `:194-197` download — `rbacSvc == nil` → 500 | handler | `TestHandler_ListProjectRecordings_NilRBAC_FailClosed_TASK064`; `…_GetRecordingDownloadURL_NilRBAC_FailClosed_TASK064` · `backend/testframework/calls/nil_rbac_recordings_task064_test.go` | убрать nil-check → 500→200 → RED |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Flag |
|----|-----------|------|
| W3-A4-006 | deferred product | **N/A** |
| W3-A4-007 … W3-A4-010 | real Hub/Bridge/Client (+ stub Authorizer) | **OK** |
| W3-A5-013 | Fiber handler + Ent recording fixture, rbacSvc=nil | **OK** |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)` · `docs-only IDs: (none)` · `N/A IDs: W3-A4-006`

### Gaps

- Docs-only: **N/A**.
- Runnable: **5/6**; N/A deferred: **W3-A4-006**.
