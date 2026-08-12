# A3 W5 агент 1 — ai handlers (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Mirror: [A2W5-1-ai-handlers.md](A2W5-1-ai-handlers.md) · Stage A: [A5-1-ai-handlers.md](A5-1-ai-handlers.md) · Assignment: [W5-ASSIGNMENTS.md](W5-ASSIGNMENTS.md) § A1  
> Чек-лист: error / IDOR / BYOK secrets / PII logs / stream cancel / RBAC  
> Периметр: 11 файлов (5 prod + 6 tests) · Прочитано: **11/11** · Пропущено: **0**

## Сверка с A2 (A2W5-A1-*)

| A2 ID | Tag | Статус в HEAD | Комментарий |
|-------|-----|---------------|-------------|
| A2W5-A1-001 | FIXED-OK | закрыт | AssistJQL: sentinel map как SuggestComposer (quota→429+Retry-After, consent→403, policy→400, no provider→503) — `handler.go:126-151`; regression `task085_regress_test.go` `TestHandler_AssistJQL_DailyQuotaExceeded` (вне слота); TASK-085 |
| A2W5-A1-002 | FIXED-OK | закрыт | `RegisterRoutes` admin/user принимают `rateLimitMw` — `admin_handler.go:22-27`, `user_handler.go:19-24`; wiring `AIRateLimit` на все три — `wiring_audit_ai.go:66-70` (вне слота) |
| A2W5-A1-003 | FIXED-OK | закрыт | Create/Update → `mapProviderMutateError` — `admin_handler.go:72-74`, `:91-93`; `user_handler.go:58-60`, `:84-86`; `client_errors.go` + TASK-113 |
| A2W5-A1-004 | FIXED-OK | закрыт | `sanitizeProviderTestError` в service TestProvider (handler отдаёт JSON as-is) — evidence `service.go:256-261`, `client_errors.go:43-52`; тест `task113_sanitize_handlers_test.go` (вне слота) |
| A2W5-A1-005 | FIXED-OK | закрыт | Response DTO `has_api_key` без plaintext — `dto.go:55-70`; Create JSON assert TASK-113 (вне слота) |

### Cascade FIXED-OK (Stage A / W5-A1-*, без регрессии)

| Stage A / prior | Evidence |
|-----------------|----------|
| W5-A1-002 stream preflight | `StartChatStream` → `resolveProviders` до 202 — `service.go:491-493`; handler мапит consent/quota/RPM/policy — `handler.go:211-228` |
| W5-A1-006 BYOK foreign-owner | `UpdateUserProvider` + Ent HTTP — `authz_ent_negative_task037_test.go` (вне слота) |
| W5-A1-008 SuggestComposer sentinel | `errors.Is(ErrInvalidPrefix\|ErrEmptySuggestion)` — `handler.go:191-194` |
| W5-A1-009 uuid.Nil | admin/user Update/Test — `admin_handler.go:81-83`, `:100-102`; `user_handler.go:71-73`; `nil_uuid_task063_test.go` (вне слота) |
| W5-A1-010 system scope | `requireSystemProvider` на Update/Test — `service.go:218-281` |
| Stream cancel (W5-A3-005) | не `Background()`; `WithoutCancel`+timeout 5m; pre-start `ctx.Err()` → cancel — `service.go:495-507` |

**REGRESS:** 0 · **RESIDUAL (этот слот):** 2 · **NEW (этот слот):** 1

## Прочитанные файлы

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `backend/internal/ai/admin_handler.go` | 158 | 0 | FIXED-OK RL MW + mapProviderMutateError + uuid.Nil; requireAdmin fail-closed |
| `backend/internal/ai/dto.go` | 171 | 0 | чисто — `has_api_key`, без plaintext key |
| `backend/internal/ai/export_hooks.go` | 266 | 0 | чисто — hubtest re-export only |
| `backend/internal/ai/handler.go` | 286 | 2 | NEW stream ErrNoProvider→400; RESIDUAL AssistJQL empty→502 |
| `backend/internal/ai/user_handler.go` | 89 | 0 | FIXED-OK RL MW + mapProviderMutateError + uuid.Nil + userID |
| `backend/testframework/ai/admin_handler_test.go` | 113 | 0 | чисто — Forbidden + happy providers/settings/test/usage |
| `backend/testframework/ai/admin_handler_wave110_test.go` | 122 | 0 | чисто — invalid JSON/fields/UUID/not-found/prompt |
| `backend/testframework/ai/coverage_wave81_test.go` | 223 | 1 | RESIDUAL — stream только 202 happy-path |
| `backend/testframework/ai/handler_extended_test.go` | 182 | 0 | чисто — bot/models/usage/prefs/invalid bodies |
| `backend/testframework/ai/handler_wave115_test.go` | 178 | 0 | чисто — quota/consent/policy non-stream + JQL no provider; **нет** stream error cases |
| `backend/testframework/ai/user_handler_test.go` | 79 | 0 | happy CRUD; foreign-owner вне файла (TASK-037) |

Файлов в списке: **11**. Прочитано: **11**. Пропущено: **0**.

## Находки (активные A3)

| ID | Tag | Sev | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|-----|-------------|---------|---------------|------------|
| A3W5-1-001 | NEW | P2 | `handler.go:210-229` | Stream-ветка `ChatCompletions` после preflight `resolveProviders` **не** мапит `ErrNoProvider` / `ErrProviderDisabled` (и offline-only `fmt.Errorf`) — catch-all → **400 + `err.Error()`** | Non-stream те же sentinel → **503** (`handler.go:238-239`). Контракт stream≠non-stream; клиент видит Bad Request вместо Service Unavailable; сырой текст (`no provider…`, `AI_OFFLINE_ONLY`) уходит в body | Добавить те же ветки, что non-stream/AssistJQL: NoProvider/Disabled → 503; AllProvidersFailed → 502; неизвестное → стабильный `"provider error"` без `err.Error()` |
| A3W5-1-002 | RESIDUAL (W5-A1-001 catch-all) | P3 | `handler.go:151`; `service.go:552-553` | AssistJQL: пустой `query` → `fmt.Errorf("query is required")` → catch-all **502 + err.Error()** | Валидация должна быть 400; квота/consent уже FIXED, но validation residual остаётся Bad Gateway | Sentinel `ErrQueryRequired` + 400 в handler (как `ErrInvalidPrefix`) **или** проверка `TrimSpace(query)==""` в handler до service |
| A3W5-1-003 | RESIDUAL (W5-A1-007) | P3 | `coverage_wave81_test.go:51-74`; `handler_wave115_test.go` (нет stream-кейсов) | Stream HTTP: только `202 Accepted` happy-path; **нет** quota→429 / consent→403 / no-provider→503 на `stream:true` | Регресс A3W5-1-001 и мёртвых/неверных статусов не ловится в файлах слота (вне слота тоже нет stream HTTP error suite) | После фикса маппинга: stream+quota→429; stream+no consent→403; stream+no provider→503 |

## Чисто (перепроверено)

- **RBAC/admin:** `requireAdmin` fail-closed (`err \|\| !isAdmin` → 403) — `admin_handler.go:38-49`; тест Forbidden.
- **IDOR/BYOK owner:** List/Create/Update с `userIDFromCtx`; Update → `UpdateUserProvider`; foreign-owner Ent+HTTP TASK-037 (вне слота).
- **BYOK secrets:** `ProviderResponse.HasAPIKey`; Create/Update не отдают `api_key`; TestProvider error санитизирован в service.
- **PII logs:** в handlers списка **нет** `slog`/`log`; IP/UA только в service audit args.
- **AIRateLimit:** user `/ai/*` + admin `/admin/ai` + `/ai/my` принимают MW; wiring прокидывает (вне слота).
- **Stream cancel:** preflight + `WithoutCancel`/timeout; отмена ctx до старта горутины — fail; post-202 cancel через WS/deadline (by design).
- **error contract (non-stream chat):** catch-all `"provider error"` без internals — `handler.go:261`.
- **Layering:** нет `*ent.Client` / `http.DefaultClient` / `io.ReadAll` в prod handlers слота.

### Drift

В списке A1 нет `client_errors.go`, `wiring_audit_ai.go`, `service.go`, TASK-037/063/085/113 тестов — сверены read-only для FIXED-OK evidence; находки без файл:строка в периметре не заводились.

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 1 |
| P3 | 2 |
| **Total** | **3** |

### Теги vs A2

| Tag | Count |
|-----|------:|
| FIXED-OK (A2 active) | 5 (`A2W5-A1-001`…`005`) |
| REGRESS | 0 |
| RESIDUAL | 2 (`A3W5-1-002`, `A3W5-1-003`) |
| NEW | 1 (`A3W5-1-001`) |

A2 baseline open (file-table REGRESS): 2 → FIXED-OK 2 · плюс FIXED cascade 003–005 · NEW 1 · RESIDUAL 2 · REGRESS 0.

### Топ

1. **A3W5-1-001 (P2 NEW)** — stream `ErrNoProvider` → 400 вместо 503 (`handler.go:229`).
2. **A3W5-1-002 (P3 RESIDUAL)** — AssistJQL empty query → 502.
3. **A3W5-1-003 (P3 RESIDUAL)** — нет stream HTTP error-тестов в слоте.
