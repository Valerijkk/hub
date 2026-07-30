# TASK-058 — Assert-free / слабые тесты

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P3** · Статус: Закрыта
> Источник: TASK-008 W12 · Classes: P3-BE-02 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P3-BACKEND.md`
> Mutation C: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-be1-task058.md` · 17/17 PROVEN (3× GAP_NOTED mock≠Ent) · W6-A3-015 GAP_NOTED residual

## [OUTCOME]

Дефекты P3-BE-02 закрыты.

## [ROLE]

- Role: be2
- Responsibility: фикс + тесты на слое guard
- Handoff to: стадия C (мутация независимым агентом)

## [SCOPE AND CONSTRAINTS]

- Files allowed: backend/testframework/**; `.planning/backlog/README.md`; этот файл
- Do not touch: несвязанные модули; secrets; DEV_MODE в CI
- Constraints: Docker + `-tags hubtest` для BE; Vitest для FE; один логический коммит на подкласс при необходимости
- Approval record: миграции Ent — только с явным согласованием владельца

## [REFERENCES]

- Реестр W12 (P3-BE-02)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P3-BE-02 с мутацией «да» закрыты по реестру
- **Мутация (C):** точный status/Ack assert краснеет при регрессии
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

> W13 slot 4 refill / qa-auto1 · criteria only (no mutate) · 2026-07-30 · HEAD guards verified · Class **P3-BE-02**
> Kill = ослабить assert / сломать ветку → точный assert RED (assert-free suite остаётся зелёным).

| ID | Production file:line (guard) | слой | Exact test RED | Kill recipe |
|----|------------------------------|------|----------------|-------------|
| W1-A4-016 | `backend/internal/sprint/handler.go:196-206` — `StartSprint` missing → service NotFound → HTTP 404 + `error` body | handler | `TestHandler_StartSprint_NotFound` · `backend/testframework/sprint/handler_wave113_test.go` | ослабить до `NotEqual(OK)` / убрать body assert → 500/400 проходит → exact 404/`не найден` RED |
| W1-A3-011 | `backend/internal/tracker` automation consumer `processMessage` — valid→Ack, unknown/bad JSON→Nak | consumer | `TestAutomationConsumer_processMessage_Events` · `backend/testframework/tracker/automation_consumer_test.go` | убрать Ack/Nak count asserts → сломанный Ack suite зелёный; kill = assert `acks==1`/`naks` RED при no-Ack |
| W7-A1-009 | `backend/internal/notification` quiet-hours cross-midnight (`start>end`) | service | `TestIsQuietHours_CrossMidnight` · `backend/testframework/notification/service_extended_test.go` | `_ = ExportIsQuietHours` / убрать expected bool → invert branch зелёный → `assert.True/False` RED |
| W7-A4-020 | `backend/internal/storage` S3 delete fail → meta row остаётся | service | assert в `backend/testframework/storage/service_wave117_test.go` (W7-A4-020) | `_ = err` на delete / не проверять meta → orphan assert RED |
| W4-A2-008 | `backend/internal/platform/outbox` — `BatchLimit`/`PollInterval` via `Export*` | constants | `TestWorkerConstants` · `backend/testframework/platform/outbox/outbox_test.go` | локальные копии const без export → drift зелёный; сломать export const → Equal RED |
| W4-A4-011 | `backend/internal/platform/config` Load — clear env between tests | config test | `TestLoad_clearsEnvBetweenTests` · `backend/testframework/platform/config/config_test.go` | assert-free Unsetenv → leak env зелёный → isolation assert RED |
| W6-A2-004 | `backend/internal/auth` Organization PendingUsers CRUD | service | `TestOrganizationService_PendingUsersCRUD` · `backend/testframework/auth/service_organization_test.go` | только `NoError` → no-op service зелёный → state assert RED |
| W6-A2-005 | `backend/internal/auth` PurgeExpiredPending cutoff | service | `TestOrganizationService_PurgeExpiredPending` · `backend/testframework/auth/service_organization_test.go` | только `n==0` без cutoff → broken cutoff зелёный → purge count RED |
| W5-A5-005 | `backend/internal/search` empty projectIDs → 0 hits (канон empty-scope) | service / Ent | `TestSearchScopedEntities_EmptyProjectIDsReturnNothing` · `backend/testframework/search/service_scope_test.go` | убрать empty-scope assert / дубли wave* → NF-03 regress зелёный → empty assert RED |
| W5-A5-006 | `backend/internal/search` `ErrFilterNotFound` sentinel | service | assert в `backend/testframework/search/service_wave109_test.go` | ослабить до generic err → wrong sentinel зелёный → `ErrorIs` RED |
| W3-A3-011 | messenger mock `AddReaction` NotFound contract | mock fixture | `TestMockRepository_AddReactionContract` · `backend/testframework/messenger/mock_repository_test.go` | мёртвый `*_test.go` без Test* / убрать NotFound assert → silent fixture |
| W6-A4-014 | audit consumer invalid JSON → Nak, no Create | consumer | `TestProcessEvent_InvalidJSON_NaksWithoutCreate` · `backend/testframework/audit/consumer_test.go` | убрать Nak/Create asserts → Ack-on-bad-JSON зелёный → Nak assert RED |
| W6-A5-007 | audit ListTrash cutoff ≈ now−30d | service | `TestAuditService_ListTrash_WithItems` · `backend/testframework/audit/service_purge_test.go` | без cutoff assert → wrong window зелёный → cutoff RED |
| W6-A5-008 | audit ReassignTasks mock counters | service | `TestAuditService_ReassignTasks` · `backend/testframework/audit/service_limits_test.go` | без call-count assert → no-op reassign зелёный → counter RED |
| W6-A5-009 | `RecordHTTPRequest` CreateAuditLog err → slog, no panic | service | `TestRecordHTTPRequest_SkipsOnCreateError` · `backend/testframework/audit/service_security_test.go` | убрать createErr inject assert → panic/silent swallow зелёный → skip/slog RED |
| W2-A5-010 | bases `ErrForbidden` → HTTP 403 в `MapServiceError` | handler map | `TestMapServiceError_AllBranches` · `backend/testframework/bases/service_wave98_test.go` | убрать Forbidden→403 ветку из таблицы → 403 regress зелёный → Equal RED |
| W3-A1-012 | messenger unmute = `POST /unmute` (нет `DELETE /mute`) | handler routes | assert в `backend/testframework/messenger/handler_wave123_test.go` | вернуть DELETE /mute route / ослабить method assert → wrong verb зелёный → route RED |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Flag |
|----|-----------|------|
| W1-A4-016 | Fiber handler + service NotFound | **OK** — HTTP |
| W1-A3-011 | consumer + counting JS msg | **OK** — mock engine OK for Ack |
| W7-A1-009 | NotificationService + Ent user quiet fields | **OK** — Ent |
| W3-A3-011 | mock repository contract | **mock≠Ent** (намеренно: fixture hygiene) |
| W6-A5-008/009 | AuditService + mockRepo | **mock≠Ent** |
| W4-A2-008 / W4-A4-011 | package const / config Load | **OK** — не Ent runtime |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: W3-A3-011, W6-A5-008, W6-A5-009` · `docs-only IDs: (none)`

### Gaps

- Docs-only: **N/A**.
- Class residuals без отдельного усиленного assert в комментарии P3-BE-02: W6-A3-015 — **GAP_NOTED**.
- Mutation proof = точный status/Ack/bool/cutoff; typecheck/`go test` green на assert-free suite **не** считается PASS.
