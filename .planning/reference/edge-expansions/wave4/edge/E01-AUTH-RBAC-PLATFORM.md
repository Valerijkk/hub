# Edge expansions wave4 — Auth / RBAC / Platform

> **Тип:** только **Edge** (границы: JWT claims/alg, magic-code charset, Authorization parse, email case-fold, TG bind, refresh encoding, custom-role bounds, membership/role matrix, archived RBAC, rate headers, UUID variants, Content-Type/query, version int extremes, cursor≠offset, health/CORS/header size).  
> **Волна:** wave4 · **не пересекает** wave1 `*100…`, wave2 `*200…`, wave3 `*300…`.  
> **ID:** `ECC-AUTH-400+`, `ECC-RBAC-400+`, `ECC-PLT-400+`.  
> **Gap:** только `TODO`.  
> **Владелец:** `qa-manual` · 2026-07-10 · **не коммитить** без ревью.  
> **Каталог:** `.planning/archive/july-2026/EDGE-CORNER-CATALOG.md` — **не изменялся**.

---

## 1. Auth — `ECC-AUTH-400+`

### 1.1 JWT claims / alg / header

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-AUTH-400 | Edge | Access JWT `alg=none` (unsigned) | 401; не принимать unsigned | API | P0 | NFR-09, JWT | TODO |
| ECC-AUTH-401 | Edge | Access JWT `alg=HS256` vs configured `RS256`/`HS512` mismatch | 401; нет fallback на weaker alg | API | P0 | alg allowlist | TODO |
| ECC-AUTH-402 | Edge | Access JWT header `kid` empty / unknown / very long (>1KB) | 401; нет panic; нет key confusion | API | P0 | kid lookup | TODO |
| ECC-AUTH-403 | Edge | Access JWT `sub` = nil UUID `00000000-…` | 401; не маппить на «system» | API | P0 | NFR-25 | TODO |
| ECC-AUTH-404 | Edge | Access JWT `sub` valid UUID but user soft-deleted | 401/403; нет silent revive | API | P0 | auth×trash | TODO |
| ECC-AUTH-405 | Edge | Access JWT missing `exp` / missing `iat` | 401; reject incomplete claims | API | P0 | NFR-09 | TODO |
| ECC-AUTH-406 | Edge | Access JWT `exp − iat` = 15m exact vs 15m+1s issued | exact OK if within policy; oversize TTL → 401 | API | P0 | access TTL=15m | TODO |
| ECC-AUTH-407 | Edge | Access JWT `aud`/`iss` wrong or empty (if claims used) | 401; documented required claims | API | P1 | JWT claims | TODO |
| ECC-AUTH-408 | Edge | Access JWT payload size just under / over parser limit | under OK; over 401; нет OOM | API | P1 | DoS JWT | TODO |
| ECC-AUTH-409 | Edge | Access JWT with extra unknown claims (large custom) | ignore extras **или** 401; no privilege from unknown | API | P1 | claim ignore | TODO |

### 1.2 Magic-code charset / format

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-AUTH-410 | Edge | Magic code: leading zeros `000123` vs stored `123` | exact string match; leading zeros significant **или** normalize documented | API | P0 | VerifyCode | TODO |
| ECC-AUTH-411 | Edge | Magic code: all digits min length vs max length exact | min OK; max OK; max+1 → 400 | API | P0 | rune length | TODO |
| ECC-AUTH-412 | Edge | Magic code: mixed case letters if alphabet allows | case policy: reject or fold; one canonical Redis compare | API | P1 | charset | TODO |
| ECC-AUTH-413 | Edge | Magic code: only whitespace / tabs / newlines | 400; не trim-to-empty success | API | P0 | ECC-AUTH-113 deepen | TODO |
| ECC-AUTH-414 | Edge | Magic code: hex-looking vs decimal-looking same length | treated as opaque string; no int parse | API | P1 | opaque code | TODO |
| ECC-AUTH-415 | Edge | Magic code: null byte `%00` mid-string | 400; Redis key not truncated | API | P0 | null inject | TODO |

### 1.3 Authorization header parse

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-AUTH-416 | Edge | `Authorization: Bearer` (no token) | 401 | API | P0 | middleware | TODO |
| ECC-AUTH-417 | Edge | `Authorization: Bearer` + only spaces | 401 | API | P0 | trim | TODO |
| ECC-AUTH-418 | Edge | `Authorization: bearer <token>` (lowercase scheme) | 200 if policy case-insensitive **или** 401 documented | API | P1 | RFC7235 | TODO |
| ECC-AUTH-419 | Edge | `Authorization: Bearer <token> trailing-garbage` | 401; не silent truncate | API | P0 | parse | TODO |
| ECC-AUTH-420 | Edge | Two `Authorization` headers (duplicate) | 401 or first-wins documented; no confuse dual JWT | API | P0 | header dup | TODO |
| ECC-AUTH-421 | Edge | `Authorization: Basic …` on Magic-Link-only API | 401; no password path | API | P0 | TZ Magic Link only | TODO |

### 1.4 Email case-fold / normalize

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-AUTH-422 | Edge | Email `User@Example.COM` vs stored `user@example.com` | one canonical account; rate/code same key | API | P0 | case-fold | TODO |
| ECC-AUTH-423 | Edge | Email local-part `+tag` / dots Gmail-style if accepted | policy: distinct vs fold; no dual accounts unintended | API | P1 | plus-address | TODO |
| ECC-AUTH-424 | Edge | Email with trailing `.` before `@` / domain trailing `.` | 400 | API | P1 | RFC edge | TODO |
| ECC-AUTH-425 | Edge | Email: consecutive dots `a..b@ex.com` | 400 | API | P1 | RFC5321 | TODO |

### 1.5 Telegram bind (not login) + session idle

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-AUTH-426 | Edge | TG bind: valid chat_id min/max int64 boundary | min/max OK; overflow 400; bind ≠ login JWT | API | P0 | ADR-AUTH-01 | TODO |
| ECC-AUTH-427 | Edge | TG bind: empty telegram_id / whitespace | 400 | API | P1 | TG bind | TODO |
| ECC-AUTH-428 | Edge | TG unbind when already unbound (idempotent) | 200/404 idempotent; no 500 | API | P1 | unbind | TODO |
| ECC-AUTH-429 | Edge | Legacy TG login route still absent (404 boundary) | 404; no JWT mint via TG | API | P0 | ECC-AUTH-015 deepen | TODO |
| ECC-AUTH-430 | Edge | Session `last_active`: refresh updates; access-only may not | documented; list order by last_active stable | API | P1 | ListSessions | TODO |
| ECC-AUTH-431 | Edge | Idle session: no refresh for `30d−1s` vs `30d+1s` | −1s refresh OK; +1s 401 (Redis TTL) | API | P0 | NFR-10 deepen | TODO |
| ECC-AUTH-432 | Edge | Refresh token string length: min opaque / max Redis value / max+1 | max OK; overflow reject; no truncate match | API | P0 | refresh format | TODO |
| ECC-AUTH-433 | Edge | Refresh token: URL-safe base64 padding `=` boundary | accept canonical encoding; reject corrupt padding | API | P1 | encoding | TODO |
| ECC-AUTH-434 | Edge | Profile `avatar_url`: empty / max URL length / `javascript:` / private IP | empty OK or clear; max OK; SSRF schemes 400 | API | P0 | NF-07 sibling | TODO |
| ECC-AUTH-435 | Edge | Profile `display_name`: max−1 / max / max+1 grapheme | max OK; +1 → 400; grapheme≠byte | API | P1 | ECC-AUTH-116 deepen | TODO |

---

## 2. RBAC / Org — `ECC-RBAC-400+`

### 2.1 Custom role name / slug / bitmask

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-RBAC-400 | Edge | Custom role name: empty / 1 char / max / max+1 | empty/overflow 400; max OK | API | P1 | custom_roles | TODO |
| ECC-RBAC-401 | Edge | Custom role slug: reserved `owner`/`viewer`/`admin` | 400 conflict with defaults | API | P0 | reserved | TODO |
| ECC-RBAC-402 | Edge | Custom role slug: unicode / spaces / UPPER | reject or normalize; one canonical | API | P1 | slug | TODO |
| ECC-RBAC-403 | Edge | Bitmask: all known bits set (Owner-equivalent custom) | create OK iff policy allows; Has() all true | Unit | P0 | bitmask full | TODO |
| ECC-RBAC-404 | Edge | Bitmask: only highest defined bit | Has(that)=true; others false | Unit | P1 | bit edge | TODO |
| ECC-RBAC-405 | Edge | Bitmask: JSON number vs string vs array of names | accept one schema; others 400 | API | P1 | API contract | TODO |
| ECC-RBAC-406 | Edge | Custom roles count: 0 / 1 / MaxCustomRoles / Max+1 | Max OK; Max+1 → 400 | API | P1 | org limit | TODO |
| ECC-RBAC-407 | Edge | Delete custom role still assigned to N members | 400 **или** reassign-to-Viewer; no orphan mask | API | P0 | role delete | TODO |

### 2.2 Membership count / role transition matrix

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-RBAC-408 | Edge | Project members: exactly 1 (sole Owner) | list size=1; remove Owner → 400 | API | P0 | last Owner | TODO |
| ECC-RBAC-409 | Edge | Project members: at MaxMembers−1 / Max / Max+1 invite | Max OK; Max+1 reject | API | P1 | member cap | TODO |
| ECC-RBAC-410 | Edge | Role transition Viewer→Developer (min elevate) | 200; cache invalidate; bits ⊃ Viewer | API | P0 | role matrix | TODO |
| ECC-RBAC-411 | Edge | Role transition Developer→Owner when Owners=MaxOwners | reject if cap; else OK | API | P1 | owner cap | TODO |
| ECC-RBAC-412 | Edge | Role transition Owner→Viewer when Owners≥2 | 200; remaining Owners≥1 | API | P0 | demote OK | TODO |
| ECC-RBAC-413 | Edge | Role transition Owner→Viewer when sole Owner | 400; founder/owner invariant | API | P0 | ECC-RBAC-107 deepen | TODO |
| ECC-RBAC-414 | Edge | Assign custom role then flip bit off while user online | next Has deny; session JWT not source of perms | API | P0 | cache×bits | TODO |
| ECC-RBAC-415 | Edge | Org `system_role=founder` count exactly 1 after member add | founder unchanged; new member never founder | API | P0 | TZ founder | TODO |

### 2.3 Archived / soft-deleted project × RBAC

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-RBAC-416 | Edge | Archived project: read list members OK; write role change | read 200; write 403/400 | API | P0 | US-S10 | TODO |
| ECC-RBAC-417 | Edge | Archived project: create invite | 403/400 | API | P1 | invite×archive | TODO |
| ECC-RBAC-418 | Edge | Soft-deleted project: HasPermission / member APIs | 404; cache miss→deny | API | P0 | NFR-11 | TODO |
| ECC-RBAC-419 | Edge | Soft-deleted member user still in project_members row | APIs treat as inactive; 403; cleanup job optional | API | P1 | member×trash | TODO |
| ECC-RBAC-420 | Edge | Restore project from trash: memberships intact | members list same; perms reload | API | P1 | restore×RBAC | TODO |

### 2.4 Invite uses extremes / org settings enum

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-RBAC-421 | Edge | Invite `max_uses=0` | 400 (useless) **или** unlimited documented; prefer reject 0 | API | P0 | max_uses | TODO |
| ECC-RBAC-422 | Edge | Invite `max_uses=1` exact: use #1 OK, #2 reject | one join; second 410/400 | API | P0 | ECC-RBAC-103 deepen | TODO |
| ECC-RBAC-423 | Edge | Invite `max_uses=INT32_MAX` / overflow in JSON | clamp or 400; no wrap to negative | API | P1 | int overflow | TODO |
| ECC-RBAC-424 | Edge | `registration_mode` enum: `open` / `invite_only` / unknown / empty | known OK; unknown/empty 400 | API | P0 | registration_mode | TODO |
| ECC-RBAC-425 | Edge | Org invite TTL: `0` / negative / max allowed / max+1 | 0/neg 400; max OK; max+1 400 | API | P1 | invite TTL | TODO |
| ECC-RBAC-426 | Edge | Pending approve: user already active (double approve) | idempotent 200 or 409; no dual membership | API | P1 | approve idem | TODO |
| ECC-RBAC-427 | Edge | Pending reject then re-register same email | new pending **или** block; no active without approve | API | P1 | reject→retry | TODO |
| ECC-RBAC-428 | Edge | List project roles: empty custom + 4 defaults | defaults always present; custom=[] OK | API | P2 | roles list | TODO |

---

## 3. Platform cross-cutting — `ECC-PLT-400+`

### 3.1 Rate-limit response headers

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-PLT-400 | Edge | API rate: response includes `Retry-After` on 429 (exact 101st) | header present; value ≥1s; body `{error,request_id}` | API | P0 | NFR-06 | TODO |
| ECC-PLT-401 | Edge | API rate: `X-RateLimit-Remaining=0` at exact 100th then 101→429 | remaining semantics documented; no negative | API | P1 | rate headers | TODO |
| ECC-PLT-402 | Edge | Auth rate 3/10m: 429 body shape + Retry-After | same error contract; no stack | API | P0 | NFR-05 | TODO |
| ECC-PLT-403 | Edge | Files rate 10/min: 10th Remaining=0; 11th 429 | boundary headers consistent | API | P1 | NFR-08 | TODO |
| ECC-PLT-404 | Edge | Messages rate 60/min: window reset at exact 60s boundary | 61st after reset → 200; not sticky | API | P0 | NFR-07 | TODO |

### 3.2 UUID variants / PK parse

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-PLT-405 | Edge | Path PK: UUID v4 uppercase vs lowercase | both accept **или** normalize; same entity | API | P1 | NFR-25 | TODO |
| ECC-PLT-406 | Edge | Path PK: UUID v1 / v5 / nil | 400 (v4-only policy) | API | P0 | NFR-25 | TODO |
| ECC-PLT-407 | Edge | Path PK: UUID with braces `{…}` / urn:uuid: | 400 | API | P1 | parse | TODO |
| ECC-PLT-408 | Edge | Path PK: truncated UUID (35 chars) / 37 chars | 400 | API | P1 | length | TODO |
| ECC-PLT-409 | Edge | Query filter `ids=` empty / one / max batch / max+1 | empty=[]; max OK; max+1 400 | API | P1 | batch IDs | TODO |

### 3.3 Content-Type / Accept / query / headers

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-PLT-410 | Edge | `Content-Type: application/json; charset=utf-8` vs missing charset | both OK for JSON body | API | P1 | CT parse | TODO |
| ECC-PLT-411 | Edge | `Content-Type: text/plain` on JSON POST | 415/400 | API | P1 | CT enforce | TODO |
| ECC-PLT-412 | Edge | `Accept: application/xml` only | 406 **или** still JSON documented | API | P2 | content nego | TODO |
| ECC-PLT-413 | Edge | Query string length: max−1 / max / max+1 | max OK; +1 → 414/400 | API | P1 | URL limit | TODO |
| ECC-PLT-414 | Edge | Header count / single header value size at proxy limit | 431/400; no crash | API | P1 | header DoS | TODO |
| ECC-PLT-415 | Edge | Unknown query params ignored vs strict | documented; no 500 | API | P2 | query unknown | TODO |

### 3.4 Optimistic version int extremes

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-PLT-416 | Edge | `version=1` first update after create | 200; version=2 | API | P0 | NFR-24 | TODO |
| ECC-PLT-417 | Edge | `version=INT32_MAX` stored then update | 200 version+1 **или** 400 overflow policy | API | P1 | int max | TODO |
| ECC-PLT-418 | Edge | `version=-1` / `version=0` on existing entity | 409/400; no wrap | API | P0 | NFR-24 | TODO |
| ECC-PLT-419 | Edge | `version` as float `1.0` / string `"1"` in JSON | 400 strict int | API | P1 | JSON types | TODO |
| ECC-PLT-420 | Edge | If-Match / version header (if supported) vs body version mismatch | 409; single source of truth documented | API | P2 | If-Match | TODO |

### 3.5 Cursor vs offset / health / CORS / routing

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-PLT-421 | Edge | Issues list: both `cursor` and `offset` sent | 400 prefer one mode **или** offset wins documented | API | P1 | NFR-14/15 | TODO |
| ECC-PLT-422 | Edge | Chat cursor: opaque min length / max / corrupt base64 | corrupt 400; max OK | API | P1 | cursor | TODO |
| ECC-PLT-423 | Edge | Chat cursor: `limit=0` / negative / max+1 | 400 or clamp; no full scan | API | P0 | NFR-15 | TODO |
| ECC-PLT-424 | Edge | `GET /health` / `/ready` without auth | 200; no PII; ready fails if DB down | API | P1 | ops | TODO |
| ECC-PLT-425 | Edge | CORS preflight `OPTIONS` on auth route | 204/200; Allow-Headers include Authorization | API | P1 | CORS | TODO |
| ECC-PLT-426 | Edge | Trailing slash `/api/v1/auth/me/` vs without | 200 both **или** 301/404 one policy | API | P2 | routing | TODO |
| ECC-PLT-427 | Edge | Method override `X-HTTP-Method-Override: DELETE` on POST | ignore override **или** documented; no accidental delete | API | P0 | method override | TODO |
| ECC-PLT-428 | Edge | Error body: `request_id` always UUID v4 (not v1/nil) | validate format; unique per request | API | P0 | NFR-13 deepen | TODO |
| ECC-PLT-429 | Edge | Soft-delete restore: entity `version` after restore boundary | next update requires returned version; 409 on stale | API | P1 | NFR-11+24 | TODO |
| ECC-PLT-430 | Edge | UTC leap-second / `:60` timestamp in client input | reject or normalize; DB stores valid UTC | API | P2 | NFR-12 | TODO |
| ECC-PLT-431 | Edge | BodyLimit: multipart boundary exactly at 25MB file part | −1 OK; +1 413; presigned PUT out of scope documented | API | P0 | NFR-04 / NF-09 | TODO |
| ECC-PLT-432 | Edge | Concurrent optimistic updates: 3 writers same version | exactly one 200; two 409; final version = start+1 | API | P0 | NFR-24 deepen | TODO |

---

## Счётчик

| Домен | ID range | Кол-во Edge | P0 | P1 | P2 |
|-------|----------|-------------|----|----|-----|
| Auth | ECC-AUTH-400…435 | **36** | 22 | 14 | 0 |
| RBAC/Org | ECC-RBAC-400…428 | **29** | 14 | 14 | 1 |
| Platform | ECC-PLT-400…432 | **33** | 12 | 16 | 5 |
| **Итого** | | **98** | **48** | **44** | **6** |

### Углубление (темы wave4)

| Тема | ID coverage | Count |
|------|-------------|-------|
| JWT claims / alg / header | AUTH-400…409 | **10** |
| Magic-code charset / format | AUTH-410…415 | **6** |
| Authorization header parse | AUTH-416…421 | **6** |
| Email case-fold / normalize | AUTH-422…425 | **4** |
| TG bind + session idle + refresh encoding + profile | AUTH-426…435 | **10** |
| Custom role name/slug/bitmask | RBAC-400…407 | **8** |
| Membership / role transition matrix | RBAC-408…415 | **8** |
| Archived / soft-deleted × RBAC | RBAC-416…420 | **5** |
| Invite uses / org settings enum | RBAC-421…428 | **8** |
| Rate-limit response headers | PLT-400…404 | **5** |
| UUID variants / PK parse | PLT-405…409 | **5** |
| Content-Type / query / headers | PLT-410…415 | **6** |
| Optimistic version int extremes | PLT-416…420 | **5** |
| Cursor≠offset / health / CORS / routing / BodyLimit | PLT-421…432 | **12** |

> Минимум 50 — **выполнено (98)**.  
> Gap — **все TODO**.  
> Happy / Neg / Corner — **не включены**.  
> Wave1 `*100+` / wave2 `*200+` / wave3 `*300+` — **не пересечены**.  
> `EDGE-CORNER-CATALOG.md` — **не изменялся**.  
> Коммит — **не создавался**.
