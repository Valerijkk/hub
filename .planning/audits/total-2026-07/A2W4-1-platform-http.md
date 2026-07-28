# A2 W4 агент 1 — platform http+middleware

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> Зеркало: `W4-ASSIGNMENTS.md` A1 · Stage A: `A4-1-platform-http.md`  
> Теги: `NEW` | `REGRESS` | `RESIDUAL` | `FIXED-OK`

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/platform/http/export_hooks.go` | 15 | 0 | чисто |
| `backend/internal/platform/http/helpers.go` | 46 | 0 | чисто (guard ME-12 на месте) |
| `backend/internal/platform/http/opengraph.go` | 110 | 0 | чисто |
| `backend/internal/platform/http/safedialer.go` | 159 | 0 | чисто |
| `backend/internal/platform/middleware/cors.go` | 102 | 0 | чисто |
| `backend/internal/platform/middleware/export_hooks.go` | 66 | 0 | чисто |
| `backend/internal/platform/middleware/logging.go` | 36 | 0 | чисто |
| `backend/internal/platform/middleware/middleware.go` | 43 | 0 | чисто |
| `backend/internal/platform/middleware/ratelimit.go` | 313 | 1 | замечания |
| `backend/internal/platform/middleware/requestid.go` | 39 | 0 | чисто |
| `backend/internal/platform/middleware/security_headers.go` | 13 | 1 | замечания |
| `backend/testframework/platform/http/helpers_test.go` | 166 | 1 | замечания |
| `backend/testframework/platform/http/opengraph_test.go` | 73 | 0 | чисто |
| `backend/testframework/platform/http/safedialer_test.go` | 103 | 0 | чисто |
| `backend/testframework/platform/middleware/middleware_test.go` | 678 | 0 | чисто |

Файлов в списке: 15. Прочитано: 15. Пропущено: 0.

## Находки (активные)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|---------|-------------------|------------|
| A2W4-1-001 | RESIDUAL | P3 | `security_headers.go:7-14` | Нет `Strict-Transport-Security` (W4-A1-007 / P3-BE-U01) | При прямом expose API без Caddy — нет HSTS; продукт делегирует Caddy, но в коде/доке нет явного контракта | HSTS в prod **или** комментарий+доки «HSTS только на Caddy» + тест/ADR |
| A2W4-1-002 | RESIDUAL | P3 | `helpers_test.go` (нет кейса); guard `helpers.go:46-48` | Нет `TestParseUUIDParam_RejectsNilUUID` (W4-A1-006) | Мутация «убрать `id == uuid.Nil`» не краснеет в `helpers_test` | Добавить кейс `00000000-…` → 400 |
| A2W4-1-003 | RESIDUAL | P3 | `ratelimit.go:207-243` | Redis RL: ZAdd → ZCard → ZRem при 429 (фикс W4-A1-004). Узкое окно: параллельный запрос видит inflated card до ZRem | Кратковременные ложные 429 под burst; не полный bypass, но residual race vs атомарный Lua | Lua: ZCARD+ZADD условно; или ZADD только если card < Max |

## FIXED-OK (сверка Stage A)

| Stage A ID | Sev | Статус | Доказательство в коде |
|------------|-----|--------|------------------------|
| W4-A1-001 | P1 | FIXED-OK | `normalizeIPForBlocklist` + `To4()`; ValidateOutboundURL IP-literal; TASK-019 |
| W4-A1-002 | P2 | FIXED-OK | `requestid.go`: CopyString + алфавит/длина ≤128 |
| W4-A1-003 | P2 | FIXED-OK | cleanup `CompareAndDelete` |
| W4-A1-004 | P2 | FIXED-OK (частично) | ZRem после reject; residual race = A2W4-1-003 |
| W4-A1-005 | P2 | FIXED-OK | `ErrorHandler` → `{"error","request_id"}` |
| W4-A1-008 | P3 | FIXED-OK | `og:image` через `ValidateOutboundURL` |
| W4-A1-009 | P3 | FIXED-OK | IP-literal block в Validate |
| W4-A1-010 | P3 | FIXED-OK | mapped IPv6 / OG ErrorIs в тестах (расширены) |
| W4-A1-011 | P3 | FIXED-OK | `isAbsoluteCORSOrigin` |

## Чисто (фокус)

| Зона | Проверено |
|------|-----------|
| SSRF | mapped IPv6 blocklist; dial-time DNS; SafeHTTPClient; OG Validate+LimitReader |
| Rate limit TZ | Auth/API/files/messages presets; Redis failover → in-memory (не open) |
| RequestID | client header sanitized |
| CORS | `*`+credentials reject; absolute origin |

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | **15 / 15 / 0** |
| Активных | **3** (NEW 0 · REGRESS 0 · RESIDUAL 3) |
| FIXED-OK | **9** |
| P0 | 0 |
| P1 | 0 |
| P2 | 0 |
| P3 | 3 |
