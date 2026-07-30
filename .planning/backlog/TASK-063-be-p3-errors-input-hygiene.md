# TASK-063 — Safe errors + input hygiene

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P3** · Статус: Закрыта
> Источник: TASK-008 W12 · Classes: P3-BE-08,P3-BE-09 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P3-BACKEND.md`
> Mutation C: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-qa1-task063.md` · 13/13 PROVEN · 2026-07-30

## [OUTCOME]

Дефекты P3-BE-08/09 закрыты.

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

- Реестр W12 (P3-BE-08,P3-BE-09)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P3-BE-08,P3-BE-09 с мутацией «да» закрыты по реестру
- **Мутация (C):** raw err internals → нет; clamp/allowlist → 400
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

> W13 slot 4 refill / qa-auto1 · criteria only (no mutate) · 2026-07-30 · HEAD guards verified · Classes **P3-BE-08**, **P3-BE-09**

| ID | Production file:line (guard) | слой | Exact test RED | Kill recipe |
|----|------------------------------|------|----------------|-------------|
| W2-A1-011 | `backend/internal/tms/handler_webhook.go:115-117` Allure; `:215-217` JUnit — stable 400 msg (не `err.Error()`) | handler | `TestWebhookAllure_InvalidJSON_StableMessage_P3BE08`; `TestWebhookJUnit_InvalidXML_StableMessage_P3BE08` · `backend/testframework/tms/webhook_safe_errors_task063_test.go` | вернуть parser `err.Error()` в 400 → NotContains/Equal RED |
| W5-A1-008 | `backend/internal/ai/handler.go:198-201` — `errors.Is(ErrInvalidPrefix|ErrEmptySuggestion)`; sentinels `composer_suggest.go:19-23` | handler / service | service: `TestSuggestComposer_PrefixTooShort` · `backend/testframework/ai/composer_suggest_test.go`; **handler HTTP `errors.Is` branch: GAP** | заменить `errors.Is` на Contains / не возвращать sentinel → service ErrorIs RED; handler map **GAP** |
| W5-A4-011 | `backend/internal/search/handler.go:132-134` createFilter; `:174-176` listFilters — stable 500 body | handler | `TestHandler_CreateFilter_SafeErrorBody_P3BE08` · `backend/testframework/search/safe_errors_task063_test.go`; **listFilters: GAP** | raw `err.Error()` в 500 body → NotContains internals RED (create); list **GAP** |
| W6-A4-011 | `backend/internal/audit/handler.go:56-62` — requireAdmin: ErrUserNotFound→401, other→500 | handler | `TestHandler_requireAdmin_RepoError_500_P3BE08`; `…_UserNotFound_401_P3BE08` · `backend/testframework/audit/require_admin_errors_task063_test.go` | любой IsSystemAdmin err → 401 → StatusInternalServerError assert RED |
| W7-A2-011 | `backend/internal/hr/service.go:187-188` ErrAccessDenied; map `handler.go:151-153` | handler map | `TestMapServiceError_AccessDenied_P3BE08` · `backend/testframework/hr/cancel_access_denied_task063_test.go` | map → `"admin access required"` → Equal `"leave access denied"` RED |
| W7-A5-008 | `backend/internal/marketplace/service.go:294-295` ErrNotOwner; map `handler.go:236-237` | service / handler | `TestMapServiceError_NotOwner_P3BE08` · `backend/testframework/marketplace/errors_hygiene_task063_test.go`; HTTP: `TestHandler_DeleteListing_Forbidden_W7A5016` · `delete_forbidden_task059_test.go` | stranger → ErrForbidden / map `"admin access required"` → ErrorIs ErrNotOwner RED |
| W1-A5-011 | `backend/internal/project/service.go:31,39-51` resolveFrontendURL; `toInviteResponse` `:559-567` | service | `TestCreateInvite_FrontendBase_P3BE09`; `…_FrontendURLEnv_P3BE09` · `backend/testframework/project/invite_frontend_task063_test.go` | toInviteResponse(..., "") → Link без frontend base → Contains RED |
| W2-A4-012 | `backend/internal/tms/importer.go:313-327` — `detectImportFormat` ToLower+HasSuffix | importer (pure) | `TestDetectImportFormat_CaseInsensitive_P3BE09` · `backend/testframework/tms/import_format_case_task063_test.go` | HasSuffix без ToLower → `.CSV` → empty format RED |
| W2-A5-008 | `backend/internal/bases/service.go:217-219` + allowlist `:328-335` — ColType | service | `TestCreateColumn_InvalidColType_P3BE09` · `backend/testframework/bases/input_hygiene_task063_test.go` | убрать isValidColumnType → invalid type не ErrInvalidInput RED |
| W2-A5-011 | `backend/internal/bases/service.go:106-108` UpdateBase `version < 1`; UpdateRecord `:268-270` | service | `TestUpdateBase_OptimisticLock_VersionEQ_W2A5004` · `ol_tx_merge_hasmore_task038_test.go`; `TestUpdateRecord_VersionZero_P3BE09` · `input_hygiene_task063_test.go` | drop `version < 1` → version=0 не ErrInvalidInput RED |
| W3-A1-011 | `backend/internal/messenger/handler.go:31-40` — queryInt clamp `v < 1 || v > 100` | handler helper | `TestQueryInt` / `"negative falls back to default"` · `backend/testframework/messenger/handler_helpers_test.go` | убрать clamp → `limit=-5` проходит → default assert RED |
| W5-A1-009 | `backend/internal/ai/admin_handler.go:81-83` UpdateProvider nil UUID; `:100-102` TestProvider; `user_handler.go:71-73` | handler | `TestAdminUpdateProvider_NilUUID_P3BE09` · `backend/testframework/ai/nil_uuid_task063_test.go`; **TestProvider + user Update: GAP** | убрать `id == uuid.Nil` → не 400 `"invalid id"` RED (admin Update); Test/user **GAP** |
| W7-A5-010 | `backend/internal/marketplace/service.go:51-53` — isValidCategory перед repo | service | `TestListApproved_InvalidCategory_P3BE09` · `backend/testframework/marketplace/errors_hygiene_task063_test.go` | убрать category allowlist → invalid category не ErrInvalidInput RED |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Flag |
|----|-----------|------|
| W2-A1-011 | Fiber webhook handler | **OK** |
| W5-A1-008 | service ErrorIs; handler HTTP map | service **OK** · handler **GAP** |
| W5-A4-011 | createFilter HTTP + mock repo | **OK** create; list **GAP** |
| W6-A4-011 | handler + mock adminErrRepo | **mock≠Ent** |
| W7-A2-011, W7-A5-008 | mapServiceError (+ HTTP Ent delete) | **OK** |
| W1-A5-011 … W7-A5-010 (P3-BE-09) | service / pure / ExportQueryInt | **OK** (кроме W5-A1-009 partial GAP) |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: W6-A4-011` · `docs-only IDs: (none)`

### Gaps

- Docs-only: **N/A**.
- Runnable primary RED: **10/13** (полный kill на primary path).
- Partial/GAP: W5-A1-008 handler HTTP; W5-A4-011 listFilters; W5-A1-009 TestProvider + user UpdateProvider nil UUID.
