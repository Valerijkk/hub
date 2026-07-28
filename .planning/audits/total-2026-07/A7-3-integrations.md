# W7 агент A3 — github+gitlab+imex+integrations/onec

> TASK-008 · READ-ONLY · 2026-07-28  
> Фокус: webhook signatures · OAuth tokens in logs · SSRF · ConstantTimeCompare gitlab.

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/github/converters.go` | 33 | 0 | чисто |
| `backend/internal/github/dto.go` | 105 | 0 | чисто |
| `backend/internal/github/export_hooks.go` | 95 | 0 | чисто |
| `backend/internal/github/handler.go` | 343 | 2 | замечания |
| `backend/internal/github/pr_review.go` | 84 | 0 | чисто |
| `backend/internal/github/preview.go` | 151 | 0 | чисто |
| `backend/internal/github/repository.go` | 493 | 2 | замечания |
| `backend/internal/github/service.go` | 560 | 2 | замечания |
| `backend/internal/github/webhook_verify.go` | 74 | 0 | чисто |
| `backend/testframework/github/converters_test.go` | 66 | 0 | чисто |
| `backend/testframework/github/ent_repository_test.go` | 289 | 0 | чисто |
| `backend/testframework/github/handler_extended_test.go` | 141 | 0 | чисто |
| `backend/testframework/github/handler_validation_test.go` | 153 | 0 | чисто |
| `backend/testframework/github/handler_wave101_test.go` | 58 | 0 | чисто |
| `backend/testframework/github/handler_wave103_test.go` | 126 | 0 | чисто |
| `backend/testframework/github/pr_review_test.go` | 107 | 0 | чисто |
| `backend/testframework/github/preview_test.go` | 128 | 0 | чисто |
| `backend/testframework/github/preview_wave103_test.go` | 57 | 0 | чисто |
| `backend/testframework/github/repository_helpers_test.go` | 54 | 0 | чисто |
| `backend/testframework/github/repository_wave101_test.go` | 59 | 0 | чисто |
| `backend/testframework/github/repository_wave103_test.go` | 134 | 0 | чисто |
| `backend/testframework/github/service_http_test.go` | 19 | 0 | чисто |
| `backend/testframework/github/service_mock_test.go` | 549 | 0 | чисто |
| `backend/testframework/github/service_test.go` | 154 | 0 | чисто |
| `backend/testframework/github/service_wave104_test.go` | 69 | 0 | чисто |
| `backend/testframework/github/service_wave115_test.go` | 38 | 0 | чисто |
| `backend/testframework/github/webhook_secret_binding_a5008_test.go` | 89 | 0 | чисто |
| `backend/internal/gitlab/dto.go` | 133 | 0 | чисто |
| `backend/internal/gitlab/export_hooks.go` | 39 | 0 | чисто |
| `backend/internal/gitlab/handler.go` | 298 | 0 | чисто |
| `backend/internal/gitlab/mr_review.go` | 156 | 0 | чисто |
| `backend/internal/gitlab/repository.go` | 285 | 1 | замечания |
| `backend/internal/gitlab/service.go` | 604 | 2 | замечания |
| `backend/internal/gitlab/webhook_verify.go` | 59 | 1 | замечания |
| `backend/testframework/gitlab/converters_test.go` | 28 | 0 | чисто |
| `backend/testframework/gitlab/ent_repository_test.go` | 229 | 0 | чисто |
| `backend/testframework/gitlab/handler_extended_test.go` | 152 | 0 | чисто |
| `backend/testframework/gitlab/handler_test.go` | 158 | 0 | чисто |
| `backend/testframework/gitlab/handler_wave115_test.go` | 37 | 0 | чисто |
| `backend/testframework/gitlab/mr_review_test.go` | 78 | 0 | чисто |
| `backend/testframework/gitlab/service_http_test.go` | 301 | 0 | чисто |
| `backend/testframework/gitlab/service_mock_test.go` | 276 | 0 | чисто |
| `backend/testframework/gitlab/service_test.go` | 74 | 0 | чисто |
| `backend/testframework/gitlab/webhook_token_binding_a5009_test.go` | 84 | 0 | чисто |
| `backend/internal/imex/dto.go` | 47 | 0 | чисто |
| `backend/internal/imex/export_hooks.go` | 47 | 0 | чисто |
| `backend/internal/imex/handler.go` | 247 | 1 | замечания |
| `backend/internal/imex/jira_import.go` | 357 | 0 | чисто |
| `backend/internal/imex/pdf.go` | 216 | 1 | замечания |
| `backend/internal/imex/repository.go` | 183 | 2 | замечания |
| `backend/internal/imex/service.go` | 321 | 1 | замечания |
| `backend/testframework/imex/handler_flow_test.go` | 225 | 0 | чисто |
| `backend/testframework/imex/jira_import_flow_test.go` | 97 | 0 | чисто |
| `backend/testframework/imex/jira_import_test.go` | 102 | 0 | чисто |
| `backend/testframework/imex/pdf_export_test.go` | 182 | 0 | чисто |
| `backend/testframework/imex/pdf_scope_a5010_ent_test.go` | 121 | 0 | чисто |
| `backend/testframework/imex/rbac_a5002_test.go` | 139 | 0 | чисто |
| `backend/testframework/imex/repository_pdf_test.go` | 78 | 0 | чисто |
| `backend/testframework/imex/repository_test.go` | 141 | 0 | чисто |
| `backend/testframework/imex/service_extended_test.go` | 172 | 0 | чисто |
| `backend/testframework/imex/service_test.go` | 102 | 0 | чисто |
| `backend/internal/integrations/onec/dto.go` | 130 | 1 | замечания |
| `backend/internal/integrations/onec/export_hooks.go` | 51 | 0 | чисто |
| `backend/internal/integrations/onec/handler.go` | 77 | 0 | чисто |
| `backend/internal/integrations/onec/service.go` | 217 | 0 | чисто |
| `backend/internal/integrations/onec/settings_store.go` | 59 | 1 | замечания |
| `backend/internal/integrations/onec/user_lookup.go` | 31 | 0 | чисто |
| `backend/testframework/onec/fixtures_test.go` | 26 | 0 | чисто |
| `backend/testframework/onec/handler_test.go` | 82 | 0 | чисто |
| `backend/testframework/onec/service_extended_test.go` | 237 | 0 | чисто |
| `backend/testframework/onec/settings_store_wave101_test.go` | 72 | 0 | чисто |

Файлов в списке: 71. Прочитано: 71. Пропущено: 0.

## Counts

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 2 |
| P2 | 8 |
| P3 | 3 |
| **Σ** | **13** |

### Топ-3

1. **W7-A3-001 (P1)** — `TriggerPipeline` публикует `req.Variables` в NATS → CI-секреты в шине/логах подписчиков.
2. **W7-A3-002 (P1)** — `GET /issues/:issueKey/github-links` только `authMw`, без project RBAC → IDOR по ключу задачи.
3. **W7-A3-003 (P2)** — GitLab `ConstantTimeCompare` на сырых строках: при `len(token)!=len(secret)` ранний выход → утечка длины секрета по таймингу.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|---------|-------------------|------------|
| W7-A3-001 | P1 | `gitlab/service.go:424-431` | `PublishJSON("gitlab.pipeline.triggered", … "variables": req.Variables)` | CI variables часто содержат токены/пароли; NATS/outbox/логи consumers получают plaintext | Публиковать только keys или redact values (`***`); тест: trigger с secret var → payload без plaintext |
| W7-A3-002 | P1 | `github/handler.go:47`; `github/service.go:468-484` | `ListGitHubLinks` по `issueKey` под общим `authMw` без `PermProjectView` / membership | Остальные github-роуты scoped к project+RBAC; любой auth-юзер, знающий ключ, читает PR/commit links чужого проекта | Route под project group + RBAC, либо resolve issue→project + membership check; denial-тест |
| W7-A3-003 | P2 | `gitlab/webhook_verify.go:41-45` | `subtle.ConstantTimeCompare([]byte(token), []byte(secret))` без выравнивания длин | Go CTC при разной длине сразу `0` (не constant-time по длине) → oracle длины webhook secret | Сравнивать `sha256(token)` vs `sha256(secret)` или HMAC обоих; тест unequal-length reject без early `==` |
| W7-A3-004 | P2 | `github/repository.go:381-424`; `gitlab/repository.go:247-284` | `ListAllWebhookSecrets` + in-memory cache всех plaintext секретов; verify path (A5-008/009) больше не вызывает | Мёртвый attack-surface: decrypt всех секретов инстанса в RAM; риск при ошибочном reuse any-of loop | Удалить API/cache или оставить unused с deprecate; verify только через FindByRepo/Path |
| W7-A3-005 | P2 | `github/handler.go:106-122` | CreateIntegration не проверяет пустые `repo_owner`/`repo_name`/`access_token`/`webhook_secret` (в отличие от gitlab handler) | `validate` tags не enforced Bind; пустой PAT/secret → мёртвая интеграция / webhook 403 «secret не настроен» | Require nonempty как gitlab; 400-тесты |
| W7-A3-006 | P2 | `gitlab/service.go:306`, `410`, `498`, `554` | Ошибки GitLab API: `fmt.Sprintf/… + string(respBody)` в Internal/error | Тело upstream (сообщения, иногда фрагменты контекста) утекает клиенту Hub | Логировать body server-side; клиенту только статус/generic message |
| W7-A3-007 | P2 | `onec/dto.go:26-31`, `101-107`; `onec/settings_store.go:40-58` | `Config.BearerToken` пишется plaintext JSON в `system_setting.value` | Компрометация БД = bearer 1С; github/gitlab токены уже AES-GCM (HI-05) | Шифровать token через `platform/crypto` до Upsert; GET по-прежнему только `bearer_token_set` |
| W7-A3-008 | P2 | `imex/repository.go:138-142` vs `128-135` | `ListProjectIssuesPlain` (ZIP) без `DeletedAtIsNil`; CSV export фильтрует soft-delete | Полный ZIP отдаёт задачи из корзины | Добавить `issue.DeletedAtIsNil()`; тест trash не в ZIP |
| W7-A3-009 | P2 | `imex/service.go:246-248`; `imex/handler.go:118-121` | `ExportIssuesExcel` = CSV, но Content-Type `…spreadsheetml.sheet` / `issues.xlsx` | Клиент получает CSV под видом XLSX — сломанный Excel-экспорт | Реализовать xlsx или убрать format=excel до реализации; тест magic bytes |
| W7-A3-010 | **FIXED** (TASK-097) | P2 | `github/repository.go`; `gitlab/repository.go` | DeleteIntegration → SetDeletedAt; list/find DeletedAtIsNil | github/gitlab soft_delete_task097_test.go |
| W7-A3-011 | P3 | `github/service.go:282-283`, `288-292`, `547-549` | `_ = LogActivity`; `PublishJSON` без проверки ошибки | Потеря аудита автосмены статуса / silent drop NATS | Проверять/логировать err; не глотать `_` |
| W7-A3-012 | P3 | `imex/pdf.go:205-208` | Обрезка ячейки PDF по `len(string)` байтам | Кириллица/emoji режется посередине руны | `[]rune` truncate как `github.truncate` |
| W7-A3-013 | P3 | `imex/repository.go:68-71` | Default workflow status = `QueryWorkflowStatuses().First()` без Order | Нестабильный стартовый статус импорта при нескольких статусах | Order by position/category todo; тест |

## Чисто (фокус + прочее)

| Зона | Что проверено |
|------|----------------|
| GitHub webhook HMAC | `webhook_verify.go`: prefix `sha256=`, hex decode, bind по owner/repo (A5-008), `hmac.Equal`/`ConstantTimeCompare` на MAC, пустой/битый signature → 403; skip path нет |
| GitLab webhook token | `webhook_verify.go` + handler: пустой `X-Gitlab-Token` → 403; bind по `path_with_namespace` (A5-009); CTC используется (см. W7-A3-003) |
| Binding regression | `webhook_secret_binding_a5008_test.go`, `webhook_token_binding_a5009_test.go` — foreign secret/token → 403; mutation notes на any-of loop |
| OAuth/PAT в логах | `slog.*` в github/gitlab/onec — **нет** AccessToken/PRIVATE-TOKEN/Bearer в полях логов; DTO ответов (`IntegrationResponse`, `SettingsResponse`) токены не отдают |
| SSRF outbound | GitHub: `NewSafeHTTPClient` в service; preview → `FetchOpenGraphMetadata`; GitLab: `ValidateOutboundURL(gitlabURL)` на create + SafeHTTPClient; 1С: `ValidateOutboundURL` в Update/post; тесты loopback block (`service_http_test`, onec handler) |
| Body DoS | Webhook/handlers: `LimitReader` + `maxWebhookBodyBytes` (10MB); GitLab API reads ≤10MB; imex import ≤25MB |
| Token at-rest (GH/GL) | `encryptField` AES-GCM при cipher≠nil; response converters без token/secret |
| 1С masking | `toSettingsResponse` → `bearer_token_set` only; admin routes + `requireAdmin` |
| imex RBAC/PDF IDOR | A5-002 permissions; A5-010 `GetTestRunInProject`/`GetSprintInProject` + pdf_scope tests |
| Layering | Handler→Service→Repository; `*ent.Client` только в Ent* adapters (repo/settings/lookup), не в Handler/Service |
| Тесты | Asserts присутствуют; A5-008/009/010/002 покрыты; service_http SSRF smoke OK |

## Фокус — краткий вердикт

| Тема | Статус |
|------|--------|
| Webhook signatures | **OK** (bind + no-skip); мёртвый `ListAllWebhookSecrets` — tech debt (004) |
| OAuth tokens in logs | **OK** (не логируются); риск — NATS variables (001) и plaintext 1С token in DB (007) |
| SSRF | **OK** SafeHTTP + ValidateOutboundURL на user URLs |
| ConstantTimeCompare gitlab | **Используется**, но сырые строки → length oracle (003); GitHub HMAC digests — равной длины, ок |

## Handoff

- Стадия B: завести задачи на **001–002** первыми; 003–007 security/data; 008–010 product/TZ; 011–013 backlog P3.
- Mutation hints уже в A5-008/009 тестах для webhook binding; для 001 — assert NATS payload без variable values.
