# Edge expansions — Auth / RBAC / Platform

> **Тип:** только **Edge** (границы: лимиты, TTL, empty/full, max size, concurrent, unicode, pagination, clock).  
> **Источник:** NFR/TZ (`docs/tz_constraints.md`, `docs/system_specification.md`) + код (`auth/magiclink.go`, `rbac/cache.go`, `wiring_fiber.go`).  
> **Не дублирует** `ECC-*-001…033` из `.planning/archive/july-2026/EDGE-CORNER-CATALOG.md`.  
> **Gap:** только `TODO` | `Future`.  
> **Владелец:** `qa-manual` · 2026-07-10 · не коммитить в каталог без ревью.

---

## 1. Auth — `ECC-AUTH-100+`

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-AUTH-100 | Edge | Magic-link код: TTL ровно `5m−1s` vs `5m+1s` (Redis key expiry) | до expiry verify OK; после — reject | API | P0 | magicLinkTTL=5m, TC-AUTH-02 | TODO |
| ECC-AUTH-101 | Edge | Magic-link rate window: 3-й запрос в окне 10m vs 4-й на границе окна | 3-й 200; 4-й 429; после `magicLinkRateTTL` снова 200 | API | P0 | NFR-05, TC-AUTH-10 | TODO |
| ECC-AUTH-102 | Edge | Verify fail counter: ровно `magicLinkVerifyMax=5` fails → 6-я | 5 fails принимаются как wrong-code; 6-я lockout | API | P0 | HI-04, ECC-AUTH-017 sibling | TODO |
| ECC-AUTH-103 | Edge | Fail-counter TTL ≈ magicLinkTTL: попытка сразу после expiry счётчика | счётчик сброшен; verify снова разрешён | API | P1 | magiclink.go recordFailure | TODO |
| ECC-AUTH-104 | Edge | Access JWT: `exp = now+15m` claim boundary (iat/exp second precision) | `now < exp` → 200; `now >= exp` → 401 | API | P0 | NFR-09 | TODO |
| ECC-AUTH-105 | Edge | Access JWT: `nbf`/`iat` в будущем на +1s vs +60s | +1s в clock-skew window или 401; +60s → 401 | API | P1 | NFR-09, JWT lib | TODO |
| ECC-AUTH-106 | Edge | Refresh Redis TTL: ключ на `30d−1s` vs `30d+1s` | до TTL refresh OK; после — 401 | API | P0 | NFR-10 | TODO |
| ECC-AUTH-107 | Edge | Concurrent refresh × N (N=2,5,10) одного refresh-токена | ровно один 200 + rotation; остальные 401; нет двух живых refresh | API | P0 | ECC-AUTH-022 sibling | TODO |
| ECC-AUTH-108 | Edge | Empty sessions list: новый user, 0 сессий после revoke-all | `[]`, HTTP 200, не null | API | P1 | TC-AUTH-08 | TODO |
| ECC-AUTH-109 | Edge | Sessions list: max practical (много устройств) — pagination/order | стабильный order (created_at desc); без OOM/timeout | API | P2 | ListSessions | TODO |
| ECC-AUTH-110 | Edge | Revoke session: текущая vs чужая vs уже отозванная (idempotent) | текущая logout-like; чужая 200; повтор revoke → 404/idempotent | API | P1 | TC-AUTH-09 | TODO |
| ECC-AUTH-111 | Edge | Email length boundary: 1 char local / 254 chars / 255 chars | ≤254 accept path; 255 → 400 | API | P1 | RFC5321, ECC-AUTH-012 sibling | TODO |
| ECC-AUTH-112 | Edge | Email unicode / IDN / emoji local-part | reject или punycode-нормализация; нет panic | API | P1 | unicode/utf8 handler | TODO |
| ECC-AUTH-113 | Edge | Magic code: empty string / whitespace-only / max rune length+1 | 400; multibyte считается по rune | API | P1 | ECC-AUTH-021 sibling | TODO |
| ECC-AUTH-114 | Edge | Invite token на границе expiry (org/project invite TTL −1s / +1s) | до expiry register/join OK; после — 400/410 | API | P0 | US-M09, TC-PROJ-08 | TODO |
| ECC-AUTH-115 | Edge | Pending user: age `7d−1s` vs `7d+1s` до PurgeExpiredPending | −1s ещё в queue; +1s purged | API | P0 | pendingUserTTL, ORG-06 | TODO |
| ECC-AUTH-116 | Edge | Profile display_name: empty / 1 char / max length / unicode+ZWJ emoji | empty→400 или default; max OK; overflow 400; emoji stored intact | API | P1 | TC-AUTH-07 | TODO |
| ECC-AUTH-117 | Edge | Timezone field: empty / `UTC` / `America/Sao_Paulo` / invalid IANA | empty/invalid → 400; valid stored; API timestamps остаются UTC | API | P1 | NFR-12 | TODO |
| ECC-AUTH-118 | Edge | Auth rate limit key: same IP many emails vs same email many IPs | лимит по IP (TZ); email-flood с разных IP не обходит IP-bucket | API | P0 | NFR-05 | TODO |
| ECC-AUTH-119 | Edge | Logout при уже expired access, но живом refresh | logout по session/refresh → refresh revoked; повтор refresh 401 | API | P1 | TC-AUTH-12 | TODO |
| ECC-AUTH-120 | Edge | Refresh сразу после create session (TTL full 30d remaining) | 200; новый access; session last_active обновлён | API | P2 | NFR-10 | TODO |
| ECC-AUTH-121 | Edge | DEV_MODE code-in-body: только non-release build; release → нет code | release build не отдаёт code; non-prod OK | API/CI | P0 | TZ DEV_MODE | TODO |
| ECC-AUTH-122 | Edge | Concurrent magic-link send same email (parallel SetNX) | один 200, остальные rate/conflict; один код в Redis | API | P1 | SetNX rateKey | TODO |
| ECC-AUTH-123 | Edge | Clock: server NTP jump ±30s mid-JWT lifetime | токены валидны в пределах remaining TTL; нет массового logout | Manual | P1 | NFR-09 | TODO |
| ECC-AUTH-124 | Edge | Onboarding flag: false→true boundary; повтор complete | первый 200 onboarding=true; повтор idempotent 200 | API | P2 | TC-AUTH-11 | TODO |

---

## 2. RBAC / Org — `ECC-RBAC-100+`

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-RBAC-100 | Edge | Permission cache TTL: hit at `5m−1s` vs miss at `5m+1s` | до TTL stale-or-fresh из Redis; после — reload from DB | API | P0 | defaultCacheTTL=5m | TODO |
| ECC-RBAC-101 | Edge | Role change → Invalidate: следующий request сразу после invalidate | новые permissions без ожидания TTL | API | P0 | ECC-RBAC-015 sibling | TODO |
| ECC-RBAC-102 | Edge | Empty project members after remove all non-owners | list `[{owner}]`; HTTP 200; нельзя empty owners | API | P1 | ECC-RBAC-016 sibling | TODO |
| ECC-RBAC-103 | Edge | Invite max-uses: use N−1 / N / N+1 | N−1,N OK; N+1 reject | API | P1 | ECC-RBAC-024 sibling | TODO |
| ECC-RBAC-104 | Edge | Invite code: empty / min length / max length / unicode | empty/overflow 400; unicode reject или normalized | API | P1 | TC-PROJ-07/08 | TODO |
| ECC-RBAC-105 | Edge | Custom role bitmask: bit 0 / last defined bit / unknown high bit | known bits OK; unknown → safe no-op или 400 | Unit | P1 | TC-RBAC-01..04 | TODO |
| ECC-RBAC-106 | Edge | Concurrent Add/Remove same permission bit | итоговая маска consistent; idempotent Add/Remove | Unit | P1 | TC-RBAC-04 | TODO |
| ECC-RBAC-107 | Edge | Org: ровно один founder; попытка назначить второго на границе | reject; founder count = 1 | API | P0 | TZ system_role founder | TODO |
| ECC-RBAC-108 | Edge | Pending queue: empty list vs full page of pending users | empty `[]` 200; pagination boundaries на admin list | API | P1 | admin approve | TODO |
| ECC-RBAC-109 | Edge | Approve pending на границе 7d purge race (approve vs purge job) | либо active user, либо deleted — нет half-state | API | P0 | ORG-06, pendingUserTTL | TODO |
| ECC-RBAC-110 | Edge | Registration mode flip: open↔invite_only; in-flight magic-link | запросы после flip видят новый mode; нет dual-accept | API | P1 | registration_mode | TODO |
| ECC-RBAC-111 | Edge | Default roles matrix boundaries: Viewer min / Owner max bits | Viewer ⊂ Dev ⊂ Owner; Has() на граничных битах | Unit | P0 | TC-RBAC-05 | TODO |
| ECC-RBAC-112 | Edge | Project invite list: 0 invites / max open invites | empty OK; создание сверх лимита (если есть) → 400 | API | P2 | invite.create | TODO |
| ECC-RBAC-113 | Edge | Member role update с `version` optimistic (stale version) | 409 Conflict; version не прыгает | API | P0 | NFR-24 | TODO |
| ECC-RBAC-114 | Edge | RBAC cache cold start: Redis empty, first HasPermission | load DB → Set cache TTL; subsequent hit | API | P1 | rbac/cache.go | TODO |
| ECC-RBAC-115 | Edge | InvalidateProject: все user-keys проекта сброшены | любой member видит fresh perms | API | P1 | InvalidateProject | TODO |
| ECC-RBAC-116 | Edge | Empty custom-role permissions mask (0 bits) | роль создаётся; Has(any)=false; assign не даёт escalate | API | P1 | custom_roles | TODO |
| ECC-RBAC-117 | Edge | Org invite link: TTL boundary −1s/+1s | join OK / expired | API | P0 | GetOrgInvite | TODO |
| ECC-RBAC-118 | Edge | Multi-org create 2nd org (out of scope boundary) | reject; 1 org на инстанс | API | P1 | NFR-28, TZ | TODO |

---

## 3. Platform cross-cutting — `ECC-PLT-100+`

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-PLT-100 | Edge | Pagination offset=0 limit=1 (first page min) | 1 item или empty; total/has_more корректны | API | P1 | NFR-14 | TODO |
| ECC-PLT-101 | Edge | Pagination offset = total−1 / offset = total / offset > total | last item / empty page 200 / empty 200 | API | P1 | NFR-14 | TODO |
| ECC-PLT-102 | Edge | Pagination limit = max allowed / max+1 / limit=0 / negative | max OK; max+1 clamp или 400; 0/neg → 400 или default | API | P1 | ECC-PLT-015 sibling | TODO |
| ECC-PLT-103 | Edge | Pagination offset negative / non-integer / huge (2^31) | 400; нет wrap/scan-all | API | P1 | NFR-14 | TODO |
| ECC-PLT-104 | Edge | Empty collection: offset/limit на пустом списке | `[]` + total=0; 200 | API | P1 | NFR-14 | TODO |
| ECC-PLT-105 | Edge | Error body: 4xx/5xx всегда `{error, request_id}` UUID v4 | shape strict; request_id валидный UUID | API | P0 | NFR-13 | TODO |
| ECC-PLT-106 | Edge | request_id: отсутствует входящий vs передан клиентом | server генерит UUID; client value не доверять слепо / echo policy | API | P1 | middleware RequestID | TODO |
| ECC-PLT-107 | Edge | Soft-delete retention: day 0 / day 29 / day 30 / day 31 | 0–29 restorable; 30+ purged | API | P0 | NFR-11 | TODO |
| ECC-PLT-108 | Edge | Soft-delete clock: `deleted_at + 30d` ровно на границе job tick | purge срабатывает ≥30d; нет early delete на 29d23h | API | P1 | NFR-11 | TODO |
| ECC-PLT-109 | Edge | Trash list empty vs full (много soft-deleted) | empty `[]`; full paginated без timeout | API | P2 | US-S03 | TODO |
| ECC-PLT-110 | Edge | UTC: create/update timestamps всегда `Z` / offset 0 в JSON | нет local offset в API; FE конвертит | API | P0 | NFR-12 | TODO |
| ECC-PLT-111 | Edge | UTC boundary: event at `23:59:59.999Z` vs `00:00:00Z` next day | хранение/сортировка UTC-correct | API | P1 | NFR-12 | TODO |
| ECC-PLT-112 | Edge | API rate: 100 req/s exact vs 101-я в том же окне 1s | ≤100 → 200; 101 → 429 + Retry-After | API | P0 | NFR-06, APIRateLimit | TODO |
| ECC-PLT-113 | Edge | API rate window slide: burst 100, wait 1s, again 100 | второй burst OK; нет sticky block | API | P1 | ratelimit.go | TODO |
| ECC-PLT-114 | Edge | Files rate 10/min: 10-й vs 11-й | 10 OK; 11 → 429 | API | P0 | NFR-08 | TODO |
| ECC-PLT-115 | Edge | Messages rate 60/min: 60-й vs 61-й | 60 OK; 61 → 429 | API | P0 | NFR-07 | TODO |
| ECC-PLT-116 | Edge | BodyLimit: body = 25MB−1 / 25MB / 25MB+1 | −1 и = OK или documented; +1 → 413/400 | API | P0 | NFR-04, BodyLimit | TODO |
| ECC-PLT-117 | Edge | BodyLimit: empty body `Content-Length: 0` vs missing CL | 200/400 по методу; нет hang | API | P2 | Fiber | TODO |
| ECC-PLT-118 | Edge | Optimistic `version`: update with version=0 / current / current+1 / stale | 0/stale → 409; current → 200 version++ | API | P0 | NFR-24 | TODO |
| ECC-PLT-119 | Edge | Optimistic version: concurrent updates same entity (2 writers) | один 200 version+1; второй 409 | API | P0 | NFR-24 | TODO |
| ECC-PLT-120 | Edge | UUID PK: valid v4 / nil UUID / string slug / int id | v4 OK; остальные 400 | API | P1 | NFR-25 | TODO |
| ECC-PLT-121 | Edge | Rate-limit Redis down → in-memory fallback boundary | fallback работает; после Redis up — consistent | API | P1 | RateLimitRedis | TODO |
| ECC-PLT-122 | Edge | Large JSON just under BodyLimit but deep nesting | reject depth или parse OK; нет stack overflow | API | P1 | DoS | TODO |
| ECC-PLT-123 | Edge | Soft-delete restore at day 29 23:59 vs day 30 00:01 | 29 restore OK; 30+ 404 | API | P0 | NFR-11 | TODO |
| ECC-PLT-124 | Edge | Pagination + soft-delete: deleted items не в основной list, есть в trash | offset/limit не «дырявят» live list | API | P1 | NFR-11/14 | TODO |
| ECC-PLT-125 | Edge | Error request_id uniqueness: 100 parallel errors | 100 unique request_id; корреляция в логах | API | P2 | NFR-13 | TODO |

---

## Счётчик

| Домен | ID range | Кол-во Edge | P0 | P1 | P2 |
|-------|----------|-------------|----|----|-----|
| Auth | ECC-AUTH-100…124 | **25** | 10 | 12 | 3 |
| RBAC/Org | ECC-RBAC-100…118 | **19** | 6 | 11 | 2 |
| Platform | ECC-PLT-100…125 | **26** | 10 | 12 | 4 |
| **Итого** | | **70** | **26** | **35** | **9** |

> Минимум 40 / цель 50+ — **выполнено (70)**.  
> Happy / Neg / Corner — **не включены**.  
> Основной `EDGE-CORNER-CATALOG.md` — **не изменялся**.
