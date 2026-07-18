# Edge expansions wave5 — Auth / RBAC / Platform

> **Тип:** только **Edge** (границы: JWT kid/rotation, TG-bind, email casefold/plus, header/proxy, RBAC bit-cap, platform DoS/headers).  
> **Волна:** wave5 · **не копии** wave1 `*100…`, wave2 `*200…`, wave3 `*300…`, wave4 `*400…`.  
> **ID:** `ECC-AUTH-500+`, `ECC-RBAC-500+`, `ECC-PLT-500+`.  
> **Gap:** только `TODO` | `Future` | `Manual`.  
> **Владелец:** `qa-manual` · 2026-07-10 · **не коммитить** без ревью.  
> **Каталог:** `.planning/july-2026/EDGE-CORNER-CATALOG.md` — **не изменялся**.

### Отличие от wave1–4 (фокус wave5)

| Тема | Wave1–2 (не повторять) | Wave5 (`*500+`) — новое |
|------|------------------------|-------------------------|
| Auth | TTL/rate/JWT 15m/refresh 30d, unicode IDN, device cap | **JWT kid/rotation**, **TG bind/unbind**, **email casefold/plus/Türkİ**, **Auth header parse**, **XFF/IPv6 bucket**, **last_active throttle** |
| RBAC | cache TTL 5m, invite batch, founder=1 | **custom-role caps**, **uint64 bit edge**, **archive freeze**, **system×project role**, **cache key namespace**, **pending reject→re-reg** |
| Platform | pagination OBO, soft-delete 29/30/31, BodyLimit 25MB | **header/query size**, **gzip/chunk bomb**, **CORS/OPTIONS**, **path null/unicode**, **If-Match/ETag**, **429 Retry-After**, **SSRF outbound**, **shutdown drain** |

---

## 1. Auth — `ECC-AUTH-500+`

### 1.1 JWT kid / signing key rotation

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-AUTH-500 | Edge | Access JWT signed with **current** kid; verify with keyring containing current+prev | 200; claims OK | API | P0 | NFR-09 key rotation | TODO |
| ECC-AUTH-501 | Edge | Access JWT signed with **previous** kid still in grace window | 200 until grace end | API | P0 | rotation grace | TODO |
| ECC-AUTH-502 | Edge | Access JWT with kid **removed** from keyring | 401; no fallback to «any key» | API | P0 | fail-closed | TODO |
| ECC-AUTH-503 | Edge | Access JWT header `alg=none` / empty alg | 401; never accept unsigned | API | P0 | JWT security | TODO |
| ECC-AUTH-504 | Edge | Access JWT `alg=HS256` when server expects only configured algs | 401 if mismatch | API | P0 | alg allowlist | TODO |
| ECC-AUTH-505 | Edge | Access JWT missing `kid` while multi-key mode on | 401 **или** try primary only (documented); no key-scan oracle | API | P0 | kid required | TODO |
| ECC-AUTH-506 | Edge | Access JWT `kid` = unknown string / overlong kid (>256) | 401; no panic; constant-ish reject | API | P1 | parse | TODO |
| ECC-AUTH-507 | Edge | Mid-lifetime rotation: token issued pre-rotate, used post-rotate within grace | still 200; after grace 401 | API | P0 | grace boundary | TODO |
| ECC-AUTH-508 | Edge | Refresh issues **new** access always with **current** kid (not stale) | new access kid=current | API | P1 | refresh×rotation | TODO |

### 1.2 Authorization header / token transport

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-AUTH-509 | Edge | `Authorization: Bearer` + token with leading/trailing spaces | trim OK **или** 401; documented | API | P1 | header parse | TODO |
| ECC-AUTH-510 | Edge | `Authorization: bearer` (lowercase) / `BEARER` | accept case-insensitive scheme **или** 401 strict | API | P1 | RFC7235 | TODO |
| ECC-AUTH-511 | Edge | `Authorization: Bearer` without token / only whitespace | 401; `{error,request_id}` | API | P0 | empty bearer | TODO |
| ECC-AUTH-512 | Edge | Two `Authorization` headers (duplicate) | reject 400/401; не брать «последний выигрывает» silently insecure | API | P0 | header smuggle | TODO |
| ECC-AUTH-513 | Edge | `Authorization: Bearer <access>` + cookie with different refresh/access | API uses documented source only; no mix-up identity | API | P0 | transport | TODO |
| ECC-AUTH-514 | Edge | Token in query `?access_token=` (if unsupported) | 401/ignore; **не** принимать query token in prod | API | P0 | TZ security | TODO |

### 1.3 Email casefold / plus / locale edges

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-AUTH-515 | Edge | Email `User@Example.COM` vs `user@example.com` | один канонический ключ Redis/DB; нет dual account | API | P0 | casefold | TODO |
| ECC-AUTH-516 | Edge | Turkish `İ`/`i` in local-part (`ı`/`I` locale trap) | NFC+casefold policy; no dual; 400 if reject-unicode | API | P0 | locale casefold | TODO |
| ECC-AUTH-517 | Edge | German `ß` vs `ss` in local-part | documented normalize **или** treat distinct; no silent merge wrong | API | P1 | casefold | TODO |
| ECC-AUTH-518 | Edge | Plus-tag `user+tag@example.com` vs `user@example.com` | distinct accounts **или** strip-plus policy documented; rate keys match policy | API | P1 | plus addressing | TODO |
| ECC-AUTH-519 | Edge | Gmail-style dots `u.s.er@gmail.com` vs `user@gmail.com` (if provider-aware) | **не** provider-guess в Hub; treat literal distinct unless ADR | API | P2 | no Gmail magic | TODO |
| ECC-AUTH-520 | Edge | Email with trailing `.` in domain / double-dot | 400; no DNS probe side-effect | API | P1 | RFC | TODO |

### 1.4 Telegram bind (not login) boundaries

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-AUTH-521 | Edge | TG bind: valid initData at `auth_date` age = max−1s | bind 200 | API | P0 | ADR-AUTH-01 | TODO |
| ECC-AUTH-522 | Edge | TG bind: `auth_date` age = max+1s | 400/401 expired; no bind | API | P0 | initData TTL | TODO |
| ECC-AUTH-523 | Edge | TG bind when telegram_id already bound to **other** user | 409/400; no steal | API | P0 | unique tg | TODO |
| ECC-AUTH-524 | Edge | TG unbind then re-bind same tg_id within 1s | OK; audit 2 events; no orphan | API | P1 | rebind | TODO |
| ECC-AUTH-525 | Edge | TG bind rate: N-th vs N+1 in window | N OK; N+1 429 | API | P1 | rate | TODO |
| ECC-AUTH-526 | Edge | Legacy `/auth/telegram/login` (if route still registered) | 404; **не** выдаёт JWT | API | P0 | ADR-AUTH-01 | TODO |

### 1.5 Session last_active / IP / deactivate edges

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-AUTH-527 | Edge | Refresh updates `last_active`: two refreshes Δt < throttle | at most one DB write **или** every write — documented; no write storm | API | P1 | last_active | TODO |
| ECC-AUTH-528 | Edge | Session created from IP-A; refresh from IP-B (NAT change) | refresh OK (no IP pin) **или** step-up policy; documented | API | P1 | IP change | TODO |
| ECC-AUTH-529 | Edge | User deactivated while access JWT still unexpired | next API 401/403; refresh 401; sessions revoked | API | P0 | TC-ADMIN-02 deepen | TODO |
| ECC-AUTH-530 | Edge | Magic-code length: exactly 6 digits incl. leading zeros `000123` | accept as string digits; not int-parse drop zeros | API | P0 | code format | TODO |
| ECC-AUTH-531 | Edge | Magic-code: 5 digits / 7 digits / hex letters | 400; fail counter rules documented | API | P1 | length boundary | TODO |
| ECC-AUTH-532 | Edge | IPv4-mapped IPv6 client (`::ffff:1.2.3.4`) vs IPv4 `1.2.3.4` rate bucket | **один** bucket; нельзя обойти auth rate сменой формы IP | API | P0 | NFR-05 | TODO |
| ECC-AUTH-533 | Edge | `X-Forwarded-For` spoof chain when trusted proxy misconfigured | fail-closed to socket IP **или** first trusted hop only; no client-picked bucket | API | P0 | proxy trust | TODO |

---

## 2. RBAC / Org — `ECC-RBAC-500+`

### 2.1 Custom roles / bitmask caps

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-RBAC-500 | Edge | Custom roles count = max−1 → create one | 201; count=max | API | P1 | custom_roles cap | Future |
| ECC-RBAC-501 | Edge | Custom roles count = max → create one | 400; no orphan row | API | P1 | custom_roles cap | Future |
| ECC-RBAC-502 | Edge | Custom role name length = max / max+1 / empty | OK / 400 / 400 | API | P1 | name boundary | TODO |
| ECC-RBAC-503 | Edge | Custom role name unicode + ZWJ / RTL | store NFC **или** 400; list no spoof | API | P1 | unicode name | TODO |
| ECC-RBAC-504 | Edge | Bitmask: set **all defined** bits | Has(each)=true; serialize len=8 | Unit | P0 | TC-RBAC | TODO |
| ECC-RBAC-505 | Edge | Bitmask: bit index = 63 (MSB uint64) if defined | set/clear/Has OK; no sign-extend bug | Unit | P0 | uint64 edge | TODO |
| ECC-RBAC-506 | Edge | Bitmask: bit index = 64 / negative | 400/no-op safe; no panic | Unit | P0 | overflow | TODO |
| ECC-RBAC-507 | Edge | Assign custom role with 0 bits to member | member Has(any write)=false; read per policy | API | P1 | empty mask | TODO |
| ECC-RBAC-508 | Edge | Clone/copy permissions from Owner template into custom | result ≤ Owner bits; cannot exceed Owner | API | P0 | escalate guard | TODO |

### 2.2 Archive / system×project / membership

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-RBAC-509 | Edge | Project archived: write API at boundary (create issue) | 403/400 read-only; GET list OK | API | P0 | US-S10 | TODO |
| ECC-RBAC-510 | Edge | Project archived: role change / invite create | 403; no new members | API | P1 | archive freeze | TODO |
| ECC-RBAC-511 | Edge | Unarchive → first write | 200; cache invalidate; fresh allow | API | P1 | unarchive | TODO |
| ECC-RBAC-512 | Edge | Org `system_role=admin` but project role=Viewer | project writes denied; org admin endpoints OK | API | P0 | system×project | TODO |
| ECC-RBAC-513 | Edge | Founder without project membership hits project API | 403/404; founder ≠ global Owner of all projects | API | P0 | TZ founder scope | TODO |
| ECC-RBAC-514 | Edge | Remove member who is sole Developer (Owners remain) | 200; list without member; cache clear | API | P1 | remove | TODO |
| ECC-RBAC-515 | Edge | Soft-deleted user still in members table row | APIs treat as absent; restore re-activates membership policy documented | API | P0 | NFR-11×RBAC | TODO |

### 2.3 Cache key / pending / invite edges (wave5)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-RBAC-516 | Edge | Cache key: `userID`+`projectID` namespace; two projects same user | independent masks; no cross-project bleed | API | P0 | cache key | TODO |
| ECC-RBAC-517 | Edge | Cache key collision probe: crafted UUID string forms | keys use binary UUID; no string-concat collision | Unit | P0 | key safety | TODO |
| ECC-RBAC-518 | Edge | InvalidateUser while ×N parallel HasPermission | all see deny/allow consistent with DB after invalidate; no grant-all | API | P0 | invalidate storm | TODO |
| ECC-RBAC-519 | Edge | Pending reject → same email re-register immediately | new pending row; old gone; no active user | API | P1 | reject→re-reg | TODO |
| ECC-RBAC-520 | Edge | Pending reject at age `7d−1s` then purge job | reject wins or purge — no active; no duplicate email unique violation storm | API | P1 | reject×purge | TODO |
| ECC-RBAC-521 | Edge | Org invite revoke then accept with old token | 400/410; membership unchanged | API | P0 | revoke | TODO |
| ECC-RBAC-522 | Edge | Project invite max **open** invites = cap / cap+1 | cap OK; +1 400 | API | P2 | open invite cap | Future |
| ECC-RBAC-523 | Edge | Member role update `version=current` vs `version=0` | current→200 version++; 0→409 | API | P0 | NFR-24 | TODO |
| ECC-RBAC-524 | Edge | RequirePermission middleware: missing project_id in route context | 400/403 fail-closed; не skip check | API | P0 | middleware | TODO |

---

## 3. Platform cross-cutting — `ECC-PLT-500+`

### 3.1 Header / query / path limits

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-PLT-500 | Edge | Single header value length = server max / max+1 | OK / 431 or 400; no hang | API | P1 | header limit | TODO |
| ECC-PLT-501 | Edge | Header count = max / max+1 (many custom X-*) | OK / reject; no O(n²) parse blowup | API | P1 | header count | TODO |
| ECC-PLT-502 | Edge | Query string length = max / max+1 | OK / 414/400 | API | P1 | query limit | TODO |
| ECC-PLT-503 | Edge | Path with `%00` null / embedded null | 400; не truncate at null | API | P0 | null byte | TODO |
| ECC-PLT-504 | Edge | Path `//double//slash` / trailing slash variants | canonical route match; no dual handlers leak | API | P1 | routing | TODO |
| ECC-PLT-505 | Edge | Path unicode NFC vs NFD same resource | one resource; no dual ACL bypass | API | P0 | path normalize | TODO |
| ECC-PLT-506 | Edge | `X-Request-ID` incoming length > 128 / non-UUID | server generates own UUID; client value not trusted as request_id body | API | P1 | NFR-13 | TODO |

### 3.2 Body encoding / compression / methods

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-PLT-507 | Edge | `Content-Encoding: gzip` small compressed → huge uncompressed (>BodyLimit) | 413/400; no memory bomb | API | P0 | zip bomb | TODO |
| ECC-PLT-508 | Edge | Chunked transfer: total decoded size = BodyLimit / +1 | = OK or documented; +1 reject | API | P0 | chunk×BodyLimit | TODO |
| ECC-PLT-509 | Edge | Both `Content-Length` and `Transfer-Encoding: chunked` | reject 400 (HTTP smuggle guard) | API | P0 | smuggle | TODO |
| ECC-PLT-510 | Edge | `Content-Type: application/json; charset=utf-16` | 400/415 **или** reject non-utf8 | API | P1 | charset | TODO |
| ECC-PLT-511 | Edge | HTTP method `TRACE` / `TRACK` | 405/404; no body echo (XST) | API | P0 | TRACE off | TODO |
| ECC-PLT-512 | Edge | `OPTIONS` CORS preflight for allowed origin | 204/200 + ACAO; disallowed origin no reflect | API | P1 | CORS | TODO |
| ECC-PLT-513 | Edge | JSON array length = max elements / max+1 (bulk endpoints) | OK / 400; documented cap | API | P1 | array cap | Future |
| ECC-PLT-514 | Edge | JSON nesting depth = max / max+1 | OK / 400; no stack overflow | API | P0 | depth (≠ BodyLimit alone) | TODO |

### 3.3 Optimistic concurrency / rate headers / shutdown

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-PLT-515 | Edge | `If-Match` / ETag = current version vs stale | current→200; stale→412/409 | API | P1 | NFR-24 headers | Future |
| ECC-PLT-516 | Edge | 429 response: `Retry-After` present, integer seconds ≥1 | client can backoff; shape stable | API | P1 | NFR-06 | TODO |
| ECC-PLT-517 | Edge | Rate limit: authenticated user key vs anonymous IP key same socket | separate buckets; user limit ≠ IP auth-bypass | API | P0 | ratelimit keys | TODO |
| ECC-PLT-518 | Edge | Graceful shutdown: in-flight request during drain window | completes 200 **или** 503 documented; no truncated JSON | API | P1 | shutdown | Manual |
| ECC-PLT-519 | Edge | `/health` / `/ready` without auth | 200; **no** secrets/PII in body | API | P1 | probes | TODO |
| ECC-PLT-520 | Edge | `/metrics` scrape: auth required **или** network-policy only | no open high-cardinality leak to world | API | P1 | metrics | TODO |

### 3.4 Outbound SSRF / UUID / soft-delete cascade (platform)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-PLT-521 | Edge | `ValidateOutboundURL`: `http://127.0.0.1/` / `::1` / link-local | reject; NF-07 | API | P0 | WebPush/webhook | TODO |
| ECC-PLT-522 | Edge | `ValidateOutboundURL`: DNS to private after public resolve (rebinding class) | reject or pin IP; no SSRF | API | P0 | NF-07 | TODO |
| ECC-PLT-523 | Edge | `ValidateOutboundURL`: `file://` / `gopher://` / empty host | reject | API | P0 | scheme allowlist | TODO |
| ECC-PLT-524 | Edge | UUID path param: v1 / v7 / nil / uppercase hex | only v4 (policy) OK; others 400 | API | P1 | NFR-25 | TODO |
| ECC-PLT-525 | Edge | Soft-delete parent with N children: cascade depth boundary | children soft-deleted or blocked; no orphan live child | API | P0 | NFR-11 cascade | TODO |
| ECC-PLT-526 | Edge | Restore parent: children restore policy (all / none / selective) | documented; version fields consistent | API | P1 | restore cascade | TODO |
| ECC-PLT-527 | Edge | Security headers on API error: no `Server` version leak; no stack | `{error,request_id}` only | API | P1 | NFR-13 | TODO |
| ECC-PLT-528 | Edge | Multipart form: boundary empty / overlong / binary junk | 400; no hang | API | P1 | multipart | TODO |

---

## Счётчик

| Домен | ID range | Кол-во Edge | P0 | P1 | P2 |
|-------|----------|-------------|----|----|-----|
| Auth | ECC-AUTH-500…533 | **34** | 21 | 12 | 1 |
| RBAC/Org | ECC-RBAC-500…524 | **25** | 14 | 10 | 1 |
| Platform | ECC-PLT-500…528 | **29** | 12 | 17 | 0 |
| **Итого** | | **88** | **47** | **39** | **2** |

### Фокус-покрытие (wave5)

| Тема | ID coverage | Count |
|------|-------------|-------|
| JWT kid / alg / rotation | AUTH-500…508 | **9** |
| Auth header / transport | AUTH-509…514 | **6** |
| Email casefold / plus / locale | AUTH-515…520 | **6** |
| Telegram bind (not login) | AUTH-521…526 | **6** |
| Session IP / deactivate / code / XFF | AUTH-527…533 | **7** |
| Custom roles / uint64 bits | RBAC-500…508 | **9** |
| Archive / system×project | RBAC-509…515 | **7** |
| Cache key / pending / invite | RBAC-516…524 | **9** |
| Header/path/query limits | PLT-500…506 | **7** |
| Encoding / CORS / methods | PLT-507…514 | **8** |
| Rate headers / shutdown / probes | PLT-515…520 | **6** |
| SSRF / UUID / cascade | PLT-521…528 | **8** |

> Минимум 50 — **выполнено (88)**.  
> Happy / Neg / Corner — **не включены**.  
> Wave1–4 `*100…*400` — **не копировались** (новые темы).  
> `EDGE-CORNER-CATALOG.md` — **не изменялся**.  
> Коммит — **не создавался**.
