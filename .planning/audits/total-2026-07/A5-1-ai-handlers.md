# W5 агент A1 — ai handlers

> TASK-008 · READ-ONLY · 2026-07-28  
> Фокус: admin-only routes, BYOK secrets, rate limit, PII in logs.  
> Сверка: `wiring_audit_ai.go` (AIRateLimit только на user `Handler`), TZ rate limits, DTO `has_api_key` без ключа, prior grok A4 (SSRF/BYOK owner — repo/adapters вне списка).

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/ai/admin_handler.go` | 151 | 3 | замечания |
| `backend/internal/ai/dto.go` | 171 | 0 | чисто |
| `backend/internal/ai/export_hooks.go` | 226 | 0 | чисто |
| `backend/internal/ai/handler.go` | 256 | 3 | замечания |
| `backend/internal/ai/user_handler.go` | 82 | 2 | замечания |
| `backend/testframework/ai/admin_handler_test.go` | 113 | 0 | чисто |
| `backend/testframework/ai/admin_handler_wave110_test.go` | 122 | 0 | чисто |
| `backend/testframework/ai/coverage_wave81_test.go` | 223 | 1 | замечания |
| `backend/testframework/ai/handler_extended_test.go` | 179 | 0 | чисто |
| `backend/testframework/ai/handler_wave115_test.go` | 178 | 1 | замечания |
| `backend/testframework/ai/user_handler_test.go` | 79 | 2 | замечания |

Файлов в списке: 11. Прочитано: 11. Пропущено: 0.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|---------|-------------------|------------|
| W5-A1-001 | P2 | `handler.go:127-131` | **AssistJQL** мапит только `ErrNoProvider` → 503; всё остальное → **502 + `err.Error()`** | `AssistJQL` → `ChatCompletion` → `ErrDailyQuotaExceeded` / `ErrProviderRateLimited` / `ErrExternalAIConsentRequired` / `ErrContentPolicyViolation` / `"query is required"` уходят как Bad Gateway без `Retry-After`. Клиент не отличит квоту от сбоя провайдера; текст ошибки утекает | Те же ветки, что в `SuggestComposer`/`ChatCompletions` (429+Retry-After, 403 consent, 400 policy/validation) |
| W5-A1-002 | P2 | `handler.go:186-207` | Stream-ветка `ChatCompletions`: handler ждёт `ErrProviderRateLimited` / `ErrExternalAIConsentRequired`, но **`StartChatStream` их не возвращает** (consent/RPM — внутри `runChatStream` после `202`) | Мёртвые ветки; HTTP **202 Accepted** до consent/provider; отказ только по WS `error` (`service.go:490-491`). Неверный контракт vs non-stream; нагрузка/стоимость до fail | Перед `go runChatStream`: `resolveProviders` + consent + content policy (уже в validate) + preflight; либо sync-ошибки наружу с теми же статусами, что non-stream |
| W5-A1-003 | P2 | `admin_handler.go:23-32`; `user_handler.go:20-24` | **`RegisterRoutes` admin и BYOK `/ai/my` без AI rate-limit MW**; wiring (`wiring_audit_ai.go:65-68`) вешает `AIRateLimit` (20/мин) **только** на user `Handler` | Admin `POST .../providers/:id/test` бьёт внешний LLM без AI RL (только global API 100/s). Скомпрометированный admin / утечка сессии → cost abuse. `/ai/my` CRUD без AI RL слабее, но асимметрия с `/ai/*` | Прокинуть `rateLimitMw` как у `Handler.RegisterRoutes`; минимум на `TestProvider` + chat-adjacent admin; BYOK mutate — отдельный мягкий лимит |
| W5-A1-004 | P2 | `admin_handler.go:67-69`; `admin_handler.go:84-86`; `user_handler.go:53-55`; `user_handler.go:74-79` | Create/Update провайдера отдают **`err.Error()`** клиенту | Ошибки encrypt/cipher/`invalid base URL: %w` / внутренности репо попадают в HTTP body (паттерн A7-007). Риск утечки деталей крипто/конфига рядом с BYOK | Клиенту — стабильные коды (`invalid body` / `invalid base_url` / `save failed`); детали — `slog` + `request_id` |
| W5-A1-005 | P2 | `admin_handler.go:97-101` | `TestProvider` при успехе сервиса отдаёт JSON с полем **`error: err.Error()`** от upstream adapter (сервис кладёт туда сырой текст) | Upstream/HTTP-клиент может включить URL, фрагменты auth, тело ответа провайдера — уходит admin-клиенту без санитизации | Санитизировать/`provider unreachable`; полный текст только в server log |
| W5-A1-006 | P2 | `user_handler_test.go:35-68`; `user_handler_test.go:70-79` | BYOK-тесты: happy-path CRUD + invalid UUID; **нет foreign-owner denial** | Guard `GetUserProvider` (`service.go:158`) / mock ownership (`service_test.go:145-147`) не проверяется с HTTP-слоя. Мутация «убрать ownership check» handler-тестами **не краснеет** | `TestUserHandler_UpdateProvider_ForeignOwner`: провайдер user B, PATCH от user A → 404; assert JSON без `api_key` |
| W5-A1-007 | P2 | `handler_wave115_test.go` (нет stream-кейсов); `coverage_wave81_test.go:51-74` | Stream: только `202 Accepted` happy-path; **нет** кейсов quota/consent на stream HTTP; wave115 покрывает quota/consent только non-stream / composer | Регресс W5-A1-002 и мёртвых веток `handler.go:192-198` не ловится | После фикса preflight: stream + quota → 429; stream + no consent → 403; до фикса — тест, что 202 не означает успех LLM (документировать WS) |
| W5-A1-008 | P3 | `handler.go:168-170` | SuggestComposer: статус 400 через **`strings.Contains(err.Error(), ...)`** | Хрупко к тексту ошибки; смешивает валидацию с transport | Sentinel `ErrInvalidPrefix` / `ErrEmptySuggestion` + `errors.Is` |
| W5-A1-009 | P3 | `admin_handler.go:76-78`; `admin_handler.go:93-95`; `user_handler.go:66-68` | `:id` через `uuid.Parse` **без** отсечения `uuid.Nil` (ME-12) | Nil UUID → not found/400 опосредованно; расхождение с `platform/http.ParseUUIDParam` | Reject `id == uuid.Nil` → 400 |
| W5-A1-010 | P3 | `admin_handler.go:74-88`; `admin_handler.go:91-101` | Admin Update/Test идут в `GetProvider`/`UpdateProvider` **без фильтра `scope=system`** | List — только system; по UUID admin патчит/пингует **user BYOK** (чужой ключ в Test). Для single-org admin часто ок, но это скрытый cross-scope доступ к секрету | Update/Test: требовать `scope=system` (или явный `?include_user=1`); user BYOK — только `/ai/my` |
| W5-A1-011 | P3 | `user_handler_test.go:44-58`; `admin_handler_test.go:55-69` | Create в тестах **без `api_key`**; нет assert, что в JSON **нет** поля `api_key` / есть только `has_api_key` | Регресс «ключ в ответе» не поймается (фокус BYOK secrets) | Create с `APIKey: "sk-test"` → decode `ProviderResponse`; `assert.NotContains(raw, "sk-test")`; `HasAPIKey==true` |

## Чисто

| Файл / зона | Что проверено |
|-------------|----------------|
| `admin_handler.go` admin-only | Group `/admin/ai` + `authMw` + `requireAdmin()`; `isAdminFn` error/`!isAdmin` → **403 fail-closed**; wiring → `auditSvc.IsSystemAdmin` |
| `dto.go` BYOK secrets | Request: `api_key`; Response: **`has_api_key` bool**, поля ключа нет; usage/settings без секретов |
| `export_hooks.go` | Только `//go:build hubtest` re-export; prod-логики нет; секреты не логируются |
| `handler.go` auth + RL hook | Все user routes под `authMw`; `rateLimitMw ...` применяется (`ai.Use`); `userIDFromCtx` 401; quota → 429+Retry-After на chat/composer; consent → 403; policy → 400 |
| `handler.go` PII/logs | **Нет `slog`/`log` в handlers**; IP/UA передаются в service для audit — не access-log PII на этом слое |
| `user_handler.go` BYOK owner | List/Create/Update с `userIDFromCtx`; Update → `UpdateUserProvider` (ownership в service/repo) |
| `admin_handler_test.go` | Forbidden non-admin; happy providers/settings/test/usage; asserts |
| `admin_handler_wave110_test.go` | Invalid JSON/fields/UUID/not-found/prompt-too-long/update/history; asserts |
| `handler_extended_test.go` | bot/models/usage/prefs/invalid bodies/`userIDFromCtx`/JQL+composer invalid; asserts |
| `handler_wave115_test.go` (кроме stream gap) | Quota/consent/policy на composer+chat non-stream; AssistJQL no provider; asserts |
| Residual OK | Days history clamp в service (≤90) — handler передаёт raw, guard ниже; BaseURL SSRF — `ValidateOutboundURL` в repo/adapters (не handler) |

**Admin-only:** периметр `requireAdmin` на группе выглядит корректным; тест Forbidden на одном GET достаточен при group MW.

**BYOK secrets (handlers):** ключ в JSON-ответах handlers не отдаётся (`ProviderResponse`); риск — `err.Error()` / `TestProviderResponse.Error`, не поле `api_key`.

**Rate limit:** user `/ai/*` (chat/models/…) — AI 20/мин через wiring; admin + `/ai/my` — дыра (W5-A1-003).

**PII in logs:** в файлах списка handlers логирования нет — чисто на HTTP-слое.

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | **11 / 11 / 0** |
| Находок | **11** |
| P0 | 0 |
| P1 | 0 |
| P2 | 7 |
| P3 | 4 |

### Топ-3

1. **W5-A1-001 (P2)** — AssistJQL отдаёт квоту/consent/policy как 502 + сырой `err.Error()`.  
2. **W5-A1-002 (P2)** — stream `202` до consent/RPM; мёртвые error-ветки в handler.  
3. **W5-A1-003 (P2)** — admin `TestProvider` / `/ai/my` без `AIRateLimit` (cost abuse).
