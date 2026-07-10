# Edge expansions wave2 — Auth / RBAC / Platform

> **Тип:** только **Edge** (границы: unicode, device limits, batch invite, cache TTL, pagination off-by-one, soft-delete day 29/30/31).  
> **Волна:** wave2 · **не пересекает** wave1 `ECC-*-100…` (см. `edge-expansions/01-AUTH-RBAC-PLATFORM.md`).  
> **ID:** `ECC-AUTH-200+`, `ECC-RBAC-200+`, `ECC-PLT-200+`.  
> **Gap:** только `TODO` | `Future` | `Manual`.  
> **Владелец:** `qa-manual` · 2026-07-10 · **не коммитить** без ревью.  
> **Каталог:** `.planning/EDGE-CORNER-CATALOG.md` — **не изменялся**.

---

## 1. Auth — `ECC-AUTH-200+`

### 1.1 Unicode / IDN emails

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-AUTH-200 | Edge | Email local-part: Cyrillic `пользователь@example.com` | 400 reject **или** documented NFC+punycode path; нет panic; нет silent truncate | API | P0 | RFC6531/TZ Magic Link, ECC-AUTH-112 sibling | TODO |
| ECC-AUTH-201 | Edge | Email domain IDN: `user@мск.рф` (Cyrillic TLD) | reject **или** punycode `xn--…` normalize before Redis key; один канонический ключ rate/code | API | P0 | IDNA2008 | TODO |
| ECC-AUTH-202 | Edge | Email: NFC vs NFD одной буквы (`é` vs `e`+combining) | одна нормализованная форма; два send не обходят rate как «разные» emails | API | P0 | unicode normalize | TODO |
| ECC-AUTH-203 | Edge | Email: fullwidth `@` / `.` (U+FF20, U+FF0E) | 400; не трактовать как ASCII separators | API | P1 | homoglyph | TODO |
| ECC-AUTH-204 | Edge | Email: zero-width joiner / BOM / soft-hyphen в local | strip→reject или 400; Redis key без невидимых rune | API | P0 | ZWJ/BOM | TODO |
| ECC-AUTH-205 | Edge | Email: emoji local `😀@example.com` | 400 (preferred) или reject path; нет crash в mailer | API | P1 | ECC-AUTH-112 deepen | TODO |
| ECC-AUTH-206 | Edge | Email: RTL mark / bidi override в local-part | 400; нет spoof в UI/logs | API | P0 | bidi spoof | TODO |
| ECC-AUTH-207 | Edge | Email: mixed script local (Latin+Cyrillic lookalikes `а`/`a`) | 400 или policy reject; не создавать dual accounts | API | P0 | homoglyph account | TODO |
| ECC-AUTH-208 | Edge | Email length: unicode local 64 grapheme vs 64 bytes vs 254 total | границы по RFC/байт-лимиту; overflow 400; grapheme≠byte | API | P1 | ECC-AUTH-111 deepen | TODO |
| ECC-AUTH-209 | Edge | Magic-link send: unicode email → Resend/SMTP envelope | mailer получает ASCII/punycode; 5xx не оставляет orphan Redis code без audit | API | P1 | hub-mail / Resend | TODO |
| ECC-AUTH-210 | Edge | Verify: код валиден, email в body с другой unicode-формой того же адреса | match после normalize **или** 400; нет login чужого ключа | API | P0 | VerifyCode key | TODO |

### 1.2 Session / device limits

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-AUTH-211 | Edge | Device limit: ровно `MaxSessions−1` активных → новый login | login OK; sessions = Max | API | P0 | ListSessions / CreateSession | TODO |
| ECC-AUTH-212 | Edge | Device limit: ровно `MaxSessions` → новый login | reject 403/429 **или** LRU revoke oldest + create; политика documented | API | P0 | session cap | TODO |
| ECC-AUTH-213 | Edge | Device limit: `MaxSessions+1` concurrent verify (parallel devices) | ≤Max живых refresh; лишние fail-closed | API | P0 | concurrent create | TODO |
| ECC-AUTH-214 | Edge | Device limit = 1: второй device login | первый refresh 401; второй жив **или** второй reject | API | P0 | single-device policy | TODO |
| ECC-AUTH-215 | Edge | Device limit = 0 / negative config (misconfig boundary) | fail-closed (deny new) или safe default ≥1; нет unlimited | API | P0 | config edge | TODO |
| ECC-AUTH-216 | Edge | Same User-Agent / fingerprint, разные refresh: count as N devices | каждый session row = 1 device; limit по rows не по UA | API | P1 | session model | TODO |
| ECC-AUTH-217 | Edge | Revoke oldest then immediately login at cap boundary | после revoke slot free; login 200 | API | P1 | revoke+create | TODO |
| ECC-AUTH-218 | Edge | ListSessions at exact MaxSessions | все Max видны; order stable; total=Max | API | P1 | TC-AUTH-08 deepen | TODO |
| ECC-AUTH-219 | Edge | Refresh on device beyond soft-cap (stale row after policy change Max↓) | sessions > new Max: refresh OK until revoke **или** force trim on refresh | API | P1 | config change | TODO |
| ECC-AUTH-220 | Edge | Logout-all at MaxSessions then login | 0 sessions → 1; Redis refresh keys cleared | API | P0 | revoke-all | TODO |

### 1.3 Invite / pending (auth-side boundaries)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-AUTH-221 | Edge | Invite token: unicode payload / non-ASCII in token string | 400; нет panic on parse | API | P1 | invite token | TODO |
| ECC-AUTH-222 | Edge | Pending user email unicode + admin approve | approve OK iff email normalized; user active with canonical email | API | P0 | ORG pending | TODO |
| ECC-AUTH-223 | Edge | Pending purge: age exactly `7d` vs `7d−1s` vs `7d+1s` (auth queue) | −1s keep; ≥7d purged; нет half-user | API | P0 | pendingUserTTL, ECC-AUTH-115 deepen | TODO |
| ECC-AUTH-224 | Edge | Magic-link for pending (not yet approved) at day 6 vs day 7 | day6: code may send but login blocked; day7: user gone | API | P1 | pending+magic | TODO |

---

## 2. RBAC / Org — `ECC-RBAC-200+`

### 2.1 Invite batch

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-RBAC-200 | Edge | Batch invite: `N=0` / empty array | 400; нет side-effects | API | P1 | invite batch | TODO |
| ECC-RBAC-201 | Edge | Batch invite: `N=1` (min useful) | 1 invite created; 200; audit 1 | API | P1 | invite batch | TODO |
| ECC-RBAC-202 | Edge | Batch invite: `N=MaxBatch` exact | all created; 200; uses/TTL per item | API | P0 | batch limit | TODO |
| ECC-RBAC-203 | Edge | Batch invite: `N=MaxBatch+1` | 400 entire batch **или** documented partial; предпочтительно all-or-nothing | API | P0 | batch limit | TODO |
| ECC-RBAC-204 | Edge | Batch invite: duplicate emails in one payload | dedupe → 1 invite **или** 400; нет dual tokens | API | P0 | dedupe | TODO |
| ECC-RBAC-205 | Edge | Batch invite: mix valid + invalid email (incl. unicode bad) | all-or-nothing 400 **или** per-item errors array; нет silent skip without report | API | P0 | partial fail | TODO |
| ECC-RBAC-206 | Edge | Batch invite: all emails already members | 409/400; 0 new invites | API | P1 | membership | TODO |
| ECC-RBAC-207 | Edge | Batch invite: role per-row Viewer/Dev/Owner boundaries | Owner-in-batch reject if policy forbids; Viewer OK | API | P1 | RBAC×invite | TODO |
| ECC-RBAC-208 | Edge | Batch invite: max-uses=1 on each of MaxBatch | each token independent; accept N times across tokens | API | P1 | max-uses | TODO |
| ECC-RBAC-209 | Edge | Batch invite: TTL −1s/+1s after create (clock on batch job) | all share same expiry bucket ±1s; expired reject | API | P1 | invite TTL | TODO |
| ECC-RBAC-210 | Edge | Batch invite: concurrent two batches overlapping emails | at most one open invite per email; second conflict | API | P0 | race batch | TODO |
| ECC-RBAC-211 | Edge | Batch invite list pagination after large batch | offset/limit shows all; no missing token | API | P2 | invite list | TODO |

### 2.2 RBAC cache TTL

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-RBAC-212 | Edge | Cache TTL: HasPermission at `5m−1s` after Set | cache hit; same bitmask; no DB | API | P0 | defaultCacheTTL=5m, ECC-RBAC-100 deepen | TODO |
| ECC-RBAC-213 | Edge | Cache TTL: HasPermission at `5m+1s` after Set | miss → DB reload → Set new TTL | API | P0 | defaultCacheTTL | TODO |
| ECC-RBAC-214 | Edge | Cache TTL: exactly `5m` (Redis PTTL=0 boundary) | treat expired as miss; reload | API | P0 | PTTL=0 | TODO |
| ECC-RBAC-215 | Edge | Cache TTL: role demote at `t=0`, check at `t=5m−1s` **without** Invalidate | may see stale Owner until TTL; document stale window ≤5m | API | P0 | stale window | TODO |
| ECC-RBAC-216 | Edge | Cache TTL: demote + Invalidate then check at `t=0+ε` | fresh deny; no wait for TTL | API | P0 | ECC-RBAC-101 deepen | TODO |
| ECC-RBAC-217 | Edge | Cache TTL: Redis key exists with TTL remaining 1s → sleep 2s → Has | miss+reload; no use-after-expire | API | P1 | expiry race | TODO |
| ECC-RBAC-218 | Edge | Cache TTL: cold miss storm ×100 parallel first Has | thundering herd OK; all consistent mask; no grant-all | API | P1 | cold start | TODO |
| ECC-RBAC-219 | Edge | Cache TTL: Set with corrupt payload near expiry | miss path; DB truth; no open perms | API | P0 | len!=8 sibling | TODO |
| ECC-RBAC-220 | Edge | Cache TTL: InvalidateUser vs InvalidateProject near TTL edge | both clear keys; next Has loads DB | API | P1 | Invalidate* | TODO |
| ECC-RBAC-221 | Edge | Cache TTL config override: TTL=0 / TTL=1s / TTL=1h | 0→no-cache or reject; 1s expires fast; 1h holds; no unlimited if 0 means ∞ | API | P1 | config | TODO |
| ECC-RBAC-222 | Edge | Cache TTL + custom role bit flip at `5m−1s` without invalidate | stale until TTL; after +1s new bits | API | P1 | custom_roles | TODO |

### 2.3 Org / membership edges (wave2)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-RBAC-223 | Edge | Batch approve pending: N=Max vs Max+1 | Max OK; Max+1 400 | API | P1 | admin queue | TODO |
| ECC-RBAC-224 | Edge | Invite batch + registration_mode=invite_only | batch creates tokens; open register still blocked | API | P1 | registration_mode | TODO |
| ECC-RBAC-225 | Edge | Empty permission cache after Redis FLUSH → first API call | DB load; 200/403 correct; no 500 open | API | P0 | cache cold | TODO |

---

## 3. Platform cross-cutting — `ECC-PLT-200+`

### 3.1 Pagination off-by-one

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-PLT-200 | Edge | Collection size=10; `offset=9 limit=1` | last item; has_more=false | API | P0 | NFR-14 off-by-one | TODO |
| ECC-PLT-201 | Edge | size=10; `offset=10 limit=1` | `[]` 200; total=10; has_more=false | API | P0 | offset=total | TODO |
| ECC-PLT-202 | Edge | size=10; `offset=11 limit=1` | `[]` 200; не 404 | API | P1 | offset>total | TODO |
| ECC-PLT-203 | Edge | size=10; `offset=0 limit=10` | all 10; has_more=false | API | P1 | full page | TODO |
| ECC-PLT-204 | Edge | size=10; `offset=0 limit=9` | 9 items; has_more=true | API | P0 | has_more boundary | TODO |
| ECC-PLT-205 | Edge | size=1; `offset=0 limit=1` / `offset=1 limit=1` | first / empty | API | P1 | singleton | TODO |
| ECC-PLT-206 | Edge | size=0; `offset=0 limit=1` | `[]` total=0 has_more=false | API | P1 | empty | TODO |
| ECC-PLT-207 | Edge | `limit=1` walk: offsets 0..total | каждый id ровно 1 раз; union = full set | API | P0 | no skip/dup | TODO |
| ECC-PLT-208 | Edge | `offset=total−1 limit=2` | 1 item (last); не wrap to start | API | P0 | clamp end | TODO |
| ECC-PLT-209 | Edge | Page boundary + stable sort tie (same created_at) | deterministic secondary key (id); no flicker swap | API | P1 | sort stability | TODO |
| ECC-PLT-210 | Edge | `offset`/`limit` as strings `"09"` / `"1e1"` / `"+1"` | 400 or strict int parse; нет float/scientific | API | P1 | parse | TODO |
| ECC-PLT-211 | Edge | Chat cursor: empty cursor vs last cursor vs cursor past end | first page / last page / empty; no off-by-one message | API | P1 | cursor≠offset | TODO |

### 3.2 Soft-delete day 29 / 30 / 31

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-PLT-212 | Edge | Soft-delete age **day 29** (deleted_at + 29d) | в trash; **restore OK**; purge job skips | API | P0 | NFR-11, ECC-PLT-107 deepen | TODO |
| ECC-PLT-213 | Edge | Soft-delete age **day 30** exact (`deleted_at + 30d`) | purge eligible; restore **reject** 404/410 after purge tick | API | P0 | NFR-11 | TODO |
| ECC-PLT-214 | Edge | Soft-delete age **day 31** | already hard-purged; restore 404; нет PII in API | API | P0 | NFR-11 | TODO |
| ECC-PLT-215 | Edge | Day 29 23:59:59Z vs day 30 00:00:00Z (second boundary) | 29 restore OK; 30+ purged on job | API | P0 | ECC-PLT-123 deepen | TODO |
| ECC-PLT-216 | Edge | Day 30 −1s vs day 30 +1s relative to deleted_at | −1s restorable; +1s purge candidate | API | P0 | retention math | TODO |
| ECC-PLT-217 | Edge | Purge job tick while entity at day 29 | not deleted; still in trash next list | API | P0 | early purge guard | TODO |
| ECC-PLT-218 | Edge | Purge job tick while entity at day 30 | hard delete; gone from trash | API | P0 | purge | TODO |
| ECC-PLT-219 | Edge | Purge job tick while entity at day 31 (late job) | idempotent; already gone; no error storm | API | P1 | idempotent purge | TODO |
| ECC-PLT-220 | Edge | Restore at day 29 then soft-delete again: retention clock resets | new 30d window from new deleted_at | API | P1 | clock reset | TODO |
| ECC-PLT-221 | Edge | Trash list pagination: mix day 1 / 29 / 30-eligible | day30 may vanish mid-list after job; live+trash separation OK | API | P1 | list+purge | TODO |
| ECC-PLT-222 | Edge | Soft-delete user day 29: magic-link send | reject (trashed); no JWT | API | P0 | auth×trash | TODO |
| ECC-PLT-223 | Edge | Soft-delete user day 30+: magic-link / refresh | 401/404; Redis sessions cleaned by job/path | API | P0 | auth×purge | TODO |
| ECC-PLT-224 | Edge | Admin hard-delete (GDPR) on day 5 vs day 29 | immediate purge both; trash bypass | API | P0 | GDPR vs retention | TODO |
| ECC-PLT-225 | Edge | Leap-day / month-length: deleted_at Jan 31 + 30d | retention by duration 30*24h (UTC), not calendar «+1 month» | API | P1 | duration vs calendar | TODO |

### 3.3 Platform misc (wave2 boundaries)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-PLT-226 | Edge | Error body on unicode email 400: `{error, request_id}` only | no stack; request_id UUID | API | P1 | NFR-13 | TODO |
| ECC-PLT-227 | Edge | Optimistic version on restore-from-trash entity | version continues or resets per policy; next update uses returned version | API | P1 | NFR-24+11 | TODO |
| ECC-PLT-228 | Edge | UTC: soft-delete `deleted_at` always Z; FE shows local day 29/30 | API UTC; FE label may differ by TZ — document | API/E2E | P1 | NFR-12 | TODO |

---

## Счётчик

| Домен | ID range | Кол-во Edge | P0 | P1 | P2 |
|-------|----------|-------------|----|----|-----|
| Auth | ECC-AUTH-200…224 | **25** | 16 | 9 | 0 |
| RBAC/Org | ECC-RBAC-200…225 | **26** | 12 | 13 | 1 |
| Platform | ECC-PLT-200…228 | **29** | 16 | 13 | 0 |
| **Итого** | | **80** | **44** | **35** | **1** |

### Углубление (запрошенные темы)

| Тема | ID coverage | Count |
|------|-------------|-------|
| Unicode emails | AUTH-200…210, 221–222; PLT-226 | **14** |
| Session device limits | AUTH-211…220 | **10** |
| Invite batch | RBAC-200…211, 223–224 | **14** |
| RBAC cache TTL | RBAC-212…222, 225 | **12** |
| Pagination off-by-one | PLT-200…211 | **12** |
| Soft-delete day 29/30/31 | PLT-212…225, 227–228 | **16** |

> Минимум 50 — **выполнено (80)**.  
> Happy / Neg / Corner — **не включены**.  
> Wave1 `*100+` — **не пересечён**.  
> `EDGE-CORNER-CATALOG.md` — **не изменялся**.  
> Коммит — **не создавался**.
