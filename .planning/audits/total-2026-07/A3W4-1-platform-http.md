# A3 W4 агент 1 — platform http+middleware (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Baseline A2: [A2W4-1-platform-http.md](A2W4-1-platform-http.md)  
> Stage A: [A4-1-platform-http.md](A4-1-platform-http.md)  
> Периметр: 15 файлов (W4-ASSIGNMENTS A1) · Прочитано: 15 · Пропущено: 0

## Сверка с A2 (A2W4-1-*)

| A2 ID | Tag (A2) | Статус в HEAD | Доказательство |
|-------|----------|---------------|----------------|
| A2W4-1-001 | FIXED → FIXED-OK | закрыт | `security_headers.go:8-11` — HSTS намеренно не в Fiber; `deployments/Caddyfile.prod:5-6` edge HSTS; `middleware_test.go:316-318` assert Empty HSTS |
| A2W4-1-002 | FIXED → FIXED-OK | закрыт | `helpers_test.go:149-166` `TestParseUUIDParam_RejectsNilUUID` + guard `helpers.go:46-48` |
| A2W4-1-003 | FIXED → FIXED-OK | закрыт | `ratelimit.go:213-266` Lua `ZCARD`→условный `ZADD`; mutation-тест вне слота `ratelimit_failclosed_test.go` (не в A1 list) |
| W4-A1-001…011 (Stage A via A2) | FIXED-OK | без регресса | mapped IPv6/`To4`, RequestID CopyString+алфавит, CompareAndDelete cleanup, ErrorHandler TZ body, OG `ValidateOutboundURL`, IP-literal Validate, CORS absolute origin |

**REGRESS:** 0 · **RESIDUAL (A2 actives):** 0 · **NEW (этот слот):** 1

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/platform/http/export_hooks.go` | 25 | 0 | чисто |
| `backend/internal/platform/http/helpers.go` | 50 | 0 | чисто (ME-12 nil UUID) |
| `backend/internal/platform/http/opengraph.go` | 122 | 0 | чисто |
| `backend/internal/platform/http/safedialer.go` | 264 | 0 | чисто (+ offline AI consent path) |
| `backend/internal/platform/middleware/cors.go` | 112 | 0 | чисто |
| `backend/internal/platform/middleware/export_hooks.go` | 78 | 0 | чисто |
| `backend/internal/platform/middleware/logging.go` | 44 | 0 | чисто (ME-03 stripQuery) |
| `backend/internal/platform/middleware/middleware.go` | 51 | 0 | чисто (TZ `error`+`request_id`) |
| `backend/internal/platform/middleware/ratelimit.go` | 375 | 0 | чисто (Lua atomic; Redis→mem fail-closed) |
| `backend/internal/platform/middleware/requestid.go` | 47 | 0 | prod чисто; gap в тестах слота → A3W4-1-001 |
| `backend/internal/platform/middleware/security_headers.go` | 20 | 0 | чисто (HSTS на Caddy) |
| `backend/testframework/platform/http/helpers_test.go` | 221 | 0 | чисто |
| `backend/testframework/platform/http/opengraph_test.go` | 82 | 0 | чисто |
| `backend/testframework/platform/http/safedialer_test.go` | 202 | 0 | чисто |
| `backend/testframework/platform/middleware/middleware_test.go` | 822 | 1 | замечания (RequestID reject) |

Файлов в списке: 15. Прочитано: 15. Пропущено: 0.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|---------|-------------------|------------|
| A3W4-1-001 | NEW | P3 | `middleware_test.go:52-69` (+ отсутствие кейса); prod `requestid.go:35-47` | Нет mutation-ready теста на reject невалидного/oversized `X-Request-ID` | `TestRequestID_PreservesExistingHeader` принимает только safe `my-custom-request-id-123`. Мутация «убрать `validRequestID` / алфавит / `maxRequestIDLen`» не краснеет в слоте A1 — регресс W4-A1-002 (CopyString+sanitize) не ловится | Добавить `TestRequestID_RejectsInvalid` / `RejectsOversized`: header с пробелами/`<>`/len>128 → сгенерированный UUID ≠ client value; assert Locals/response |

## FIXED-OK (сверка A2 + Stage A фокус)

| Зона | Проверено в HEAD |
|------|------------------|
| SSRF SafeDial | `normalizeIPForBlocklist`+`To4`; dial-time DNS; reject if any IP blocked; dial resolved IP; IP-literal в `ValidateOutboundURL`; OG Validate+LimitReader+og:image Validate |
| Offline AI | `ValidateOfflineAIURL` / `SafeDialContextAllowLoopback` не ослабляют глобальный SafeDial (тесты P1-BE-U01) |
| Rate limit TZ | Auth 3/10м, API 100/с, files 10/мин, messages 60/мин, AI 20/мин; Redis Lua atomic; nil/error → in-memory (не open) |
| RequestID prod | CopyString + UUID/алфавит ≤128 |
| CORS | `*`+credentials reject; absolute origin; prod empty → Exit |
| Security headers | nosniff/DENY/Referrer/CSP; HSTS только Caddy |
| Logging | path без query (токены) |
| ErrorHandler | только `{"error","request_id"}` |

## Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 0 |
| P3 | 1 |
| **Total** | **1** |

| Метрика | Значение |
|---------|----------|
| Файлов | **15 / 15 / 0** |
| Активных | **1** (NEW 1 · REGRESS 0 · RESIDUAL 0) |
| FIXED-OK (A2 actives) | **3/3** |
| FIXED-OK (Stage A via A2) | **9+** без регресса |

### Топ-3

1. **A3W4-1-001 (P3, NEW)** — нет mutation-теста reject невалидного `X-Request-ID` в `middleware_test.go`.
2. _(нет P2+)_ — A2 residuals HSTS / nil-UUID / Lua — **FIXED-OK**.
3. _(нет P1/P0)_ — SSRF/CORS/RL/ErrorHandler без регресса.

---

Файлов в списке: 15. Прочитано: 15. Пропущено: 0.  
Counts: P0=0 · P1=0 · P2=0 · P3=1 · Total=1.  
Топ: A3W4-1-001.
