# W4 агент A1 — platform http+middleware

> TASK-008 · READ-ONLY · 2026-07-28  
> Фокус: SafeDialer/SSRF, CORS, rate limit, request ID, security headers, opengraph `ValidateOutboundURL`.  
> Сверка: TZ (`docs/tz_constraints.md` ошибки/rate limits), prior `grok-2026-07` A9-014 (RL до auth — wiring, вне списка), FIX-W17 AppError PASS (контракт `message`/`code` ≠ TZ `error`).

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/platform/http/export_hooks.go` | 17 | 0 | чисто |
| `backend/internal/platform/http/helpers.go` | 51 | 0 | чисто |
| `backend/internal/platform/http/opengraph.go` | 120 | 1 | замечания |
| `backend/internal/platform/http/safedialer.go` | 146 | 2 | замечания |
| `backend/internal/platform/middleware/cors.go` | 85 | 1 | замечания |
| `backend/internal/platform/middleware/export_hooks.go` | 46 | 0 | чисто |
| `backend/internal/platform/middleware/logging.go` | 45 | 0 | чисто |
| `backend/internal/platform/middleware/middleware.go` | 42 | 1 | замечания |
| `backend/internal/platform/middleware/ratelimit.go` | 350 | 2 | замечания |
| `backend/internal/platform/middleware/requestid.go` | 23 | 1 | замечания |
| `backend/internal/platform/middleware/security_headers.go` | 16 | 1 | замечания |
| `backend/testframework/platform/http/helpers_test.go` | 203 | 1 | замечания |
| `backend/testframework/platform/http/opengraph_test.go` | 47 | 1 | замечания |
| `backend/testframework/platform/http/safedialer_test.go` | 62 | 1 | замечания |
| `backend/testframework/platform/middleware/middleware_test.go` | 791 | 0 | чисто |

Файлов в списке: 15. Прочитано: 15. Пропущено: 0.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|---------|-------------------|------------|
| W4-A1-001 | P1 | `safedialer.go:65-75`; `safedialer.go:121-128` | **`isBlockedIP` не нормализует IPv4-mapped IPv6.** `net.IPNet.Contains` требует одинаковую длину IP; сеть `127.0.0.0/8` (4 байта) **не** матчит `::ffff:127.0.0.1` / `::ffff:169.254.169.254` (16 байт). `SafeDialContext` резолвит AAAA и dial по IP | SSRF: attacker-controlled DNS (или литерал host) с AAAA `::ffff:169.254.169.254` / loopback проходит список `blockedCIDRs` и уходит на metadata/loopback. `ValidateOutboundURL` IP не проверяет (`:80-95`) — весь периметр на dial | В `isBlockedIP`: `if v4 := ip.To4(); v4 != nil { ip = v4 }`; тест на `::ffff:127.0.0.1`, `::ffff:169.254.169.254`, `::ffff:10.0.0.1` → `ErrBlockedAddress` |
| W4-A1-002 | P2 | `requestid.go:12-18` | Клиентский `X-Request-ID` принимается **как есть**: без `utils.CopyString`, без лимита длины/алфавита. Комментарий про Pitfall 1 закрывает только ветку `uuid.NewString()` | (1) fasthttp buffer reuse → порча `Locals`/`логов` при reuse ctx; (2) огромный/мусорный ID в ответах и access-логах; (3) подмена correlation ID | Всегда `CopyString`; если header задан — валидировать UUID (или `[A-Za-z0-9._-]{1,128}`), иначе генерировать свой |
| W4-A1-003 | P2 | `ratelimit.go:338-341`; `ratelimit.go:113-133` | **cleanup race:** при пустом window — `Unlock` → `entries.Delete(key)`. Параллельный `allow` мог `LoadOrStore` тот же `*slidingWindowEntry`, дописать timestamps, после чего `Delete` выкидывает ключ из map | Orphaned entry + следующий запрос создаёт **новый** пустой bucket → сброс счётчика / кратковременный bypass in-memory RL (API 100/s, files/messages) на тике cleanup (5 мин) | `CompareAndDelete` (Go 1.20+) под тем же условием «пусто», или delete только если `Load` всё ещё указывает на тот же pointer; не Delete после unlock без CAS |
| W4-A1-004 | P2 | `ratelimit.go:207-239` | Redis sliding window: `ZAdd` **до** проверки `ZCard > Max`. Отклонённый (429) запрос уже в ZSET | Окно раздувается отказами → ложные 429 для легитимных клиентов после burst; asymmetric vs in-memory (`allow` не append при reject, `:128-130`) | Pipeline: ZCARD → если `< Max` тогда ZADD; или ZADD + при `card > Max` ZREM member и 429; Lua-скрипт атомарно |
| W4-A1-005 | P2 | `middleware.go:24-40` | `ErrorHandler` отдаёт `AppError` с JSON `code`/`message`/`request_id` | TZ (`tz_constraints.md`): ответ ошибок `{"error":"message","request_id":"uuid"}`. Контракт API ≠ инвариант ТЗ (prior FIX-W17 принял AppError shape — **расхождение зафиксировать**, не замазывать) | Либо сменить сериализацию на TZ (`error` + `request_id`, code в HTTP status), либо ADR/правка TZ под фактический `AppError`; синхронизировать wiki/API |
| W4-A1-006 | P2 | `helpers_test.go` (нет кейса); guard `helpers.go:46-48` | Нет теста на отклонение `uuid.Nil` (`00000000-…`) — ME-12 | Мутация «убрать check `id == uuid.Nil`» **не краснеет** в `helpers_test` | `TestParseUUIDParam_RejectsNilUUID` + assert 400 |
| W4-A1-007 | P3 | `security_headers.go:7-14` | Нет `Strict-Transport-Security` (и опционально `Permissions-Policy`) | При прямом expose API без Caddy HSTS — downgrade/MITM на cookie-less Bearer всё ещё полезен для mixed content / bookmarks; неполный базовый набор headers | Добавить HSTS в prod (или явно делегировать Caddy в комментарии + тест/доке); не дублировать слабый max-age |
| W4-A1-008 | P3 | `opengraph.go:103-106`; `opengraph.go:26-28` | `og:image` кладётся в `ImageURL` **без** `ValidateOutboundURL` | Fetch страницы защищён; потребители meta могут дернуть `ImageURL` (private/`file:`) вторым fetch без guard | Прогонять `ValidateOutboundURL` (и схему) на `og:image` перед возвратом; иначе очищать поле |
| W4-A1-009 | P3 | `safedialer.go:80-95` | `ValidateOutboundURL` не режет IP-литералы из `blockedCIDRs` на этапе parse | Ранний fail только на dial; расхождение DX/тестов; при ошибочном клиенте без SafeDial — дыра (контракт «вызвать Validate до Do» недостаточен один) | После parse host: `net.ParseIP` → `isBlockedIP` → `ErrBlockedAddress` |
| W4-A1-010 | P3 | `opengraph_test.go:43-46`; `safedialer_test.go:41-48` | `BlocksLoopback` / `IsBlockedIP` не покрывают IPv4-mapped; loopback OG не делает `ErrorIs(..., ErrBlockedAddress)` | Регресс W4-A1-001 не ловится текущими тестами | Добавить table-cases `::ffff:…`; в OG — `assert.ErrorIs(ErrBlockedAddress)` |
| W4-A1-011 | P3 | `cors.go:55-71` | `parseCORSOrigins` trim/`*`/empty, но **не** валидирует абсолютный origin (`https://host`) | В `CORS_ORIGINS` можно прописать мусор/`null`/path — Fiber cors поведение неочевидно; prod fail-start только на **пустом** списке после фильтра | Reject без `http://`/`https://`, без path/query; unit-cases в `ExportParseCORSOrigins` |

## Чисто

| Файл / зона | Что проверено |
|-------------|----------------|
| `safedialer.go` SSRF baseline | Схемы только http/https; userinfo ban; empty host; dial-time DNS + blocklist RFC1918/CGNAT/link-local/metadata/`::1`/ULA/multicast; reject if **any** resolved IP blocked; dial by resolved IP (anti-rebinding на connect); `NewSafeHTTPClient` timeouts; **нет** `http.DefaultClient` |
| `opengraph.go` NF-07-класс | `ValidateOutboundURL` → `NewSafeHTTPClient(10s)` → `LimitReader` 1 MiB; body close; non-2xx error; UA/Accept заданы |
| `helpers.go` | `UserIDFromCtx` type/empty/parse → 401; `ParseUUIDParam` CopyString + ME-12 nil UUID → 400 |
| `cors.go` HI-11 | `*` отвергается с credentials; trim; prod без origins → `os.Exit(1)`; dev fallback localhost:5173/8080 |
| `logging.go` ME-03 | `stripQuery` перед slog; path без `?token=` |
| `ratelimit.go` TZ presets | Auth 3/10м, API 100/1с, files 10/мин, messages 60/мин, AI 20/мин; Redis failover → in-memory (не fail-open без лимита); cleanup на `ctx.Done` (ME-5); email extractor без Bind |
| `security_headers.go` | nosniff, DENY frame, Referrer-Policy, CSP `default-src 'self'` |
| `requestid.go` (генерируемая ветка) | `uuid.NewString()` + echo `X-Request-ID` |
| `export_hooks.go` (оба) | Только hubtest re-export — без prod-логики |
| `helpers_test.go` (кроме Nil) | Success/missing/invalid/wrong-type userID; ParseUUID/ProjectID asserts |
| `opengraph_test.go` (базовое) | OG parse + fallback title; `file://` → `ErrBlockedScheme` |
| `safedialer_test.go` (базовое) | scheme/userinfo/empty; dial loopback → `ErrBlockedAddress`; SafeHTTPClient Get loopback error |
| `middleware_test.go` | RequestID generate/preserve/unique; ErrorHandler App/fiber/generic; CORS allow/deny/credentials/parse; SecurityHeaders; RL allow/block/window/by-user/Redis miniredis/nil/failover/cleanup cancel; stripQuery; hashIP; ExtractEmail |

**Периметр SSRF (кроме W4-A1-001):** для «обычных» A/AAAA публичных+RFC1918 guard выглядит зрелым; opengraph вызывает Validate+Safe client. IPv4-mapped — главный реальный пробел в этом списке файлов.

**Residual вне списка:** A9-014 — `RateLimitByUser` на `/files`/`/messenger` до auth в `wiring_fiber.go` → IP fallback (`ratelimit.go:91-95` by design). Не дублировать как новую находку кода в A1.

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | **15 / 15 / 0** |
| Находок | **11** |
| P0 | 0 |
| P1 | 1 |
| P2 | 5 |
| P3 | 5 |

### Топ-3

1. **W4-A1-001 (P1)** — IPv4-mapped IPv6 обходит `isBlockedIP` / SafeDial SSRF.  
2. **W4-A1-003 (P2)** — race cleanup in-memory rate limiter → сброс bucket.  
3. **W4-A1-002 (P2)** — небезопасное доверие к клиентскому `X-Request-ID`.
