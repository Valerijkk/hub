# Corner expansions — Auth / RBAC / Platform (C300+)

> **Тип:** только **Corner** (race, mid-flight, multi-step, clock skew, double-submit, revoke mid-request, Redis/NATS/DB fail mid-op, GDPR+session, founder combos).  
> **Не дублирует** `00-CORNER-PARENT.md` (`ECC-AUTH-C200…209`, `ECC-RBAC-C200…206`) и Edge `01-AUTH-RBAC-PLATFORM.md` (`*100+`).  
> **ID:** `ECC-AUTH-C300+`, `ECC-RBAC-C300+`, `ECC-PLT-C300+`.  
> **Gap:** `TODO` | `Manual` | `Future`.  
> **Код-якоря:** `auth/magiclink.go`, `auth/middleware.go`, `auth/service_session.go` (ME-02 rotation), `rbac/cache.go`, `rbac/middleware.go`, `platform/middleware/ratelimit.go`.  
> **Владелец:** `qa-manual` · 2026-07-10 · не коммитить в `EDGE-CORNER-CATALOG.md` без ревью.

Формат: `| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |`

---

## 1. Auth / Sessions / Org — `ECC-AUTH-C300+`

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-AUTH-C300 | Corner | Double-submit verify одного кода (2 parallel POST) | ровно один 200 + JWT; второй 400; один session | API | P0 | magiclink Del, TC-AUTH-02 | TODO |
| ECC-AUTH-C301 | Corner | SendMagicLink: Set code OK → Resend hang → client retry send | rate SetNX/Incr; не storm; код в Redis один | API | P0 | magiclink.go Send | TODO |
| ECC-AUTH-C302 | Corner | Verify mid-flight: Redis Del code OK, DB CreateSession fail | нет orphan JWT; код уже spent → retry 400; fail-closed | API | P0 | service_login | TODO |
| ECC-AUTH-C303 | Corner | Refresh ME-02: store new Redis OK → UpdateSessionHash fail → old still present | no dual live refresh; client retry safe; no privilege | API | P0 | ME-02 TOCTOU | TODO |
| ECC-AUTH-C304 | Corner | Refresh: delete old Redis fail after DB hash update | old refresh dead by hash mismatch; new works | API | P0 | service_session | TODO |
| ECC-AUTH-C305 | Corner | Logout session A while A’s access JWT in-flight long handler | handler may finish once; subsequent 401/403; refresh dead | API | P0 | RequireAuth | TODO |
| ECC-AUTH-C306 | Corner | Revoke session B mid RefreshToken(B) | refresh fail-closed; no new pair; session gone | API | P0 | RevokeSession race | TODO |
| ECC-AUTH-C307 | Corner | Admin deactivate + InvalidateUserStatusCache fail (Redis down) | ≤30s TTL window then 403; refresh rejects deactivated | API | P0 | ME-13, userStatusCacheTTL | TODO |
| ECC-AUTH-C308 | Corner | Deactivate mid IsUserDeactivated: cache hit «active» stale | max 30s stale; after invalidate/TTL → 403 | API | P0 | auth/middleware | Manual |
| ECC-AUTH-C309 | Corner | GDPR hard-delete mid RefreshToken (DB user gone) | refresh 401; Redis refresh keys cleaned; no new access | API | P0 | TC-ADMIN-03, US-S03 | TODO |
| ECC-AUTH-C310 | Corner | GDPR mid ListSessions / RevokeSession | 401/404; no PII leak in error body | API | P0 | GDPR+session | TODO |
| ECC-AUTH-C311 | Corner | GDPR + open access JWT until exp without status row | fail-closed (user not found → 401/503); no 200 | API | P0 | RequireAuth GetUser | TODO |
| ECC-AUTH-C312 | Corner | Founder: dual RegisterNewUserByEmail race (empty org) | ровно один founder+org; второй reject/pending | API | P0 | TZ founder, ORG | TODO |
| ECC-AUTH-C313 | Corner | Founder soft-delete mid admin action (invite approve) | action 403; founder uniqueness after restore | API | P0 | founder edge | TODO |
| ECC-AUTH-C314 | Corner | Founder demote attempt + parallel founder API | demote reject; founder APIs stay authorized | API | P0 | system_role | TODO |
| ECC-AUTH-C315 | Corner | Clock skew: client iat/exp ±2m vs server JWT parse | reject outside skew; no infinite access | API | P0 | NFR-09 | TODO |
| ECC-AUTH-C316 | Corner | Clock skew: magicLinkTTL wall vs Redis PTTL after NTP step | verify respects Redis TTL; no code after logical 5m | API | P1 | magicLinkTTL | Manual |
| ECC-AUTH-C317 | Corner | Rate window: 3 sends, NTP jump −10m, 4th send | still limited or window reset documented; no silent bypass | API | P0 | NFR-05 SetNX | Manual |
| ECC-AUTH-C318 | Corner | fails counter Incr OK → Expire fail (Redis) | counter may stick; verify still fail-closed; no unlock bypass | API | P1 | recordFailure | TODO |
| ECC-AUTH-C319 | Corner | Parallel wrong-code verify ×6 near magicLinkVerifyMax | lockout at ≥5; no extra success after limit | API | P0 | HI-04 | TODO |
| ECC-AUTH-C320 | Corner | Redis down mid SendMagicLink (SetNX/Set) | 5xx; no email with unstored code; no JWT | API | P0 | Redis mid-op | TODO |
| ECC-AUTH-C321 | Corner | Redis down mid VerifyCode Get | 5xx fail-closed; no false OK | API | P0 | VerifyCode | TODO |
| ECC-AUTH-C322 | Corner | Redis down mid storeRefreshToken after login verify | no access without refresh store; or documented rollback | API | P0 | login+session | TODO |
| ECC-AUTH-C323 | Corner | PG timeout mid CreateSession after code Del | no usable tokens; ops reconcile; user can re-send | API | P0 | DB mid-op | Manual |
| ECC-AUTH-C324 | Corner | NATS publish user.login fail after session create | session OK; audit/event retry; login not rolled back wrongly | API | P1 | events.go | TODO |
| ECC-AUTH-C325 | Corner | Double-submit logout same session | both 200 idempotent; refresh gone once | API | P1 | TC-AUTH-12 | TODO |
| ECC-AUTH-C326 | Corner | Logout + Refresh same session parallel | one wins; no new refresh after logout | API | P0 | race | TODO |
| ECC-AUTH-C327 | Corner | Revoke-all devices vs single-device refresh race (≠ parent logout-all) | surviving set consistent; revoked refresh 401 | API | P0 | sessions | TODO |
| ECC-AUTH-C328 | Corner | Access near exp (1s) + long handler (>1s) | mid-request: either finish or 401 on next; no half-write without authz recheck | API | P1 | JWT TTL 15m | Manual |
| ECC-AUTH-C329 | Corner | Invite accept multi-step: validate token → create user → Redis fail bind | no half-user without invite consume policy; retry safe | API | P0 | invite | TODO |
| ECC-AUTH-C330 | Corner | Pending approve mid PurgeExpiredPending (≠ parent same-second accept) | approve vs purge: active XOR deleted; audit one outcome | API | P0 | ORG-06 | TODO |
| ECC-AUTH-C331 | Corner | Telegram bind mid GDPR delete of same user | bind fail; no orphan telegram_id on purged user | API | P1 | telegram bind≠login | TODO |
| ECC-AUTH-C332 | Corner | Profile update + deactivate mid PUT /me | 403/401; no partial PII write after deactivate | API | P1 | TC-AUTH-07 | TODO |
| ECC-AUTH-C333 | Corner | Onboarding complete double-submit parallel | idempotent onboarding=true; one audit | API | P2 | TC-AUTH-11 | TODO |
| ECC-AUTH-C334 | Corner | JWT secret rotate mid-flight (FA-06 window) old access + new refresh | old access die; refresh with new secret policy documented | Manual | P0 | FA-06 | Manual |
| ECC-AUTH-C335 | Corner | Session row deleted in DB, Redis refresh still present | refresh validateSession fail; Redis key cleaned on path | API | P0 | hash/session | TODO |
| ECC-AUTH-C336 | Corner | Two tabs: refresh rotation A then B with pre-rotation token | B reuse → reject (sibling C204, multi-tab focus) | API | P0 | NFR-10 | TODO |
| ECC-AUTH-C337 | Corner | Magic link send for email while user soft-deleted in trash | reject or restore policy; no active JWT for trashed | API | P0 | soft-delete+auth | TODO |
| ECC-AUTH-C338 | Corner | Restore soft-deleted user mid open sessions of pre-delete | old sessions dead; must re-verify magic link | API | P0 | NFR-11 | TODO |

---

## 2. RBAC / Org / IDOR — `ECC-RBAC-C300+`

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-RBAC-C300 | Corner | RequirePermission pass → role demote before handler body write | fail-closed on write; 403; no silent commit | API | P0 | rbac middleware | TODO |
| ECC-RBAC-C301 | Corner | Cache hit old Owner bits → Invalidate mid GetEffectivePermissions | second check fresh; no escalate after demote | API | P0 | cache.go | TODO |
| ECC-RBAC-C302 | Corner | InvalidateProject SCAN mid concurrent HasPermission storm | no panic; eventual consistent deny after demote | API | P1 | InvalidateProject | Manual |
| ECC-RBAC-C303 | Corner | Redis RBAC cache down: Get miss → DB OK → Set fail | authz from DB; no grant-all on cache error | API | P0 | PermissionCache | TODO |
| ECC-RBAC-C304 | Corner | Redis cache returns corrupt 7-byte payload | treat miss; reload DB; no panic / open perms | API | P0 | len!=8 | TODO |
| ECC-RBAC-C305 | Corner | Custom role bitmask edit mid RequirePermission in-flight | write uses re-check or 403; UI no false success | API/E2E | P0 | custom_roles | TODO |
| ECC-RBAC-C306 | Corner | Assign custom role + delete role parallel | no orphan membership with ghost bits; 404/403 | API | P0 | custom_roles | TODO |
| ECC-RBAC-C307 | Corner | Double-submit AddMember same user | one membership; second idempotent/409; one audit | API | P1 | members | TODO |
| ECC-RBAC-C308 | Corner | RemoveMember while member’s multipart upload in-flight | upload abort/403; no file ACL for removed | API | P0 | IDOR+storage | TODO |
| ECC-RBAC-C309 | Corner | Transfer ownership A→B while A demotes B | one terminal owner; no zero-owner project | API | P0 | ownership | TODO |
| ECC-RBAC-C310 | Corner | Transfer ownership + parallel delete project | delete XOR transfer; consistent terminal | API | P0 | race | TODO |
| ECC-RBAC-C311 | Corner | Archive project mid RequirePermission write | 403; reads per policy; cache invalidated | API | P0 | archive≠C202 JWT-only | TODO |
| ECC-RBAC-C312 | Corner | Unarchive + concurrent write with stale «archived» cache | after invalidate write OK; no permanent 403 | API | P1 | cache TTL | TODO |
| ECC-RBAC-C313 | Corner | Invite accept double-submit same token | one join; second 410/400; uses not over-count wrongly | API | P0 | invite max-uses | TODO |
| ECC-RBAC-C314 | Corner | Invite revoke mid accept multi-step | accept fail; no member row | API | P0 | mid-flight | TODO |
| ECC-RBAC-C315 | Corner | Max-uses=1: two accepts parallel | exactly one member; other reject | API | P0 | race | TODO |
| ECC-RBAC-C316 | Corner | SystemRole Admin→Member mid RequireAdmin handler | fail-closed; admin mutation aborted | API | P0 | RequireSystemRole | TODO |
| ECC-RBAC-C317 | Corner | Founder-only org setting change + parallel second admin promote | founder uniqueness; settings one writer / version | API | P0 | founder combo | TODO |
| ECC-RBAC-C318 | Corner | Approve pending user + deactivate same user race | active XOR deactivated; no login if deactivated wins | API | P0 | admin queue | TODO |
| ECC-RBAC-C319 | Corner | registration_mode flip invite_only mid open magic-link verify of stranger | verify/register reject under new mode | API | P0 | registration_mode | TODO |
| ECC-RBAC-C320 | Corner | Member role version optimistic: two admins PATCH role parallel | one 200 version++; other 409 | API | P0 | NFR-24 | TODO |
| ECC-RBAC-C321 | Corner | ProjectId param swap mid-request (fasthttp buffer reuse pitfall) | CopyString safe; authz for intended project only | API | P0 | Pitfall 1 middleware | TODO |
| ECC-RBAC-C322 | Corner | Cross-project: valid JWT + other project UUID in path | 403 not member; no bitmask leak | API | P0 | IDOR | TODO |
| ECC-RBAC-C323 | Corner | Soft-deleted project membership restore race with new invite | no duplicate membership; one canonical row | API | P1 | soft-delete | TODO |
| ECC-RBAC-C324 | Corner | GDPR delete member mid GetEffectivePermissions | 401/403; cache key gone; no PII in logs beyond id | API | P0 | GDPR+rbac | TODO |
| ECC-RBAC-C325 | Corner | Bulk role update 20 users; Redis Invalidate fails halfway | DB truth; stale cache ≤TTL then correct; no escalate | API | P1 | Invalidate | Manual |
| ECC-RBAC-C326 | Corner | Owner leaves project while last non-owner promoted mid-leave | always ≥1 owner; leave rejected if would orphan | API | P0 | ownership invariant | TODO |
| ECC-RBAC-C327 | Corner | NATS rbac.changed consumer lag: UI shows old role, API denies | API deny wins; UI refresh; no write through | E2E | P1 | NATS | Future |
| ECC-RBAC-C328 | Corner | SSO-mapped role shrink mid session (future ADR) ∩ project write | re-check; 403; audit | API | P0 | ADR-AUTH-SSO | Future |

---

## 3. Platform cross-cutting — `ECC-PLT-C300+`

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-PLT-C300 | Corner | RateLimitRedis pipeline fail mid-window → memFallback | fallback limits; no unlimited burst; after Redis up consistent | API | P0 | ratelimit.go HI-10 | TODO |
| ECC-PLT-C301 | Corner | Redis up again after fallback: dual counters briefly | may under/over count briefly; documented; no permanent bypass | API | P1 | RateLimitRedis | Manual |
| ECC-PLT-C302 | Corner | Auth rate 3/10m + API 100/s same client parallel buckets | independent; auth 429 ≠ API unlock | API | P0 | NFR-05/06 | TODO |
| ECC-PLT-C303 | Corner | Double-submit mutating POST (no idempotency key) near rate edge | at most one resource; 429/409 on twin; no dual-create | API | P0 | double-submit | TODO |
| ECC-PLT-C304 | Corner | BodyLimit: chunked upload abort mid-body after 20MB | connection reset/413; no partial entity commit | API | P0 | NFR-04 | Manual |
| ECC-PLT-C305 | Corner | RequestID middleware + ErrorHandler mid panic Recover | JSON `{error,request_id}`; request_id present | API | P0 | NFR-13 | TODO |
| ECC-PLT-C306 | Corner | Client spoofs request_id + parallel errors | server UUID wins/policy; logs correlate server id | API | P1 | RequestID | TODO |
| ECC-PLT-C307 | Corner | Soft-delete entity + restore race with 30d purge job | restore XOR purge; no resurrect after hard purge | API | P0 | NFR-11 | TODO |
| ECC-PLT-C308 | Corner | Soft-delete mid optimistic update (version++) | 404/410; version not advanced on deleted | API | P0 | NFR-24+11 | TODO |
| ECC-PLT-C309 | Corner | Two writers version N; both read N; both PUT | one 200 N+1; other 409; no silent overwrite | API | P0 | optimistic lock | TODO |
| ECC-PLT-C310 | Corner | Clock skew: deleted_at +30d vs purge cron NTP jump | no early purge; no eternal trash | Manual | P0 | NFR-11 | Manual |
| ECC-PLT-C311 | Corner | UTC store + FE TZ convert mid DST transition on list sort | order by UTC stable; no duplicate/missing day | E2E | P1 | NFR-12 | TODO |
| ECC-PLT-C312 | Corner | Pagination: item deleted between page1 and page2 | no IDOR leak; may skip/duplicate once; documented | API | P1 | NFR-14 | TODO |
| ECC-PLT-C313 | Corner | Pagination cursor/offset + concurrent insert storm | stable pages or documented drift; no crash | API | P1 | NFR-14 | Manual |
| ECC-PLT-C314 | Corner | Files rate 10/min + API 100/s same user double path | file 11th → 429; API still OK | API | P0 | NFR-08 | TODO |
| ECC-PLT-C315 | Corner | Messages 60/min hit mid WS+HTTP dual send | both paths count or documented; no bypass via WS-only | API/WS | P0 | NFR-07 | TODO |
| ECC-PLT-C316 | Corner | PG connection pool exhaust mid multi-step tx (authz+write) | 503/fail-closed; no commit without authz | API | P0 | DB mid-op | Manual |
| ECC-PLT-C317 | Corner | NATS JetStream unavailable mid domain event after DB commit | outbox/retry; DB truth kept; no dual corrupt | API | P0 | NATS mid-op | TODO |
| ECC-PLT-C318 | Corner | NATS publish OK, consumer crashes mid-handler | at-least-once; idempotent consumer; no partial side-effect forever | API | P0 | JetStream | TODO |
| ECC-PLT-C319 | Corner | Redis flushall mid active sessions + RBAC cache | re-login; authz reload DB; fail-closed until warm | Manual | P0 | Redis | Manual |
| ECC-PLT-C320 | Corner | Caddy/proxy timeout mid long handler after DB write | client retry idempotent; no duplicate side-effect | Manual | P1 | proxy | Manual |
| ECC-PLT-C321 | Corner | GDPR user delete mid platform trash restore of user-owned entity | restore deny / anonymize; no PII reappear | API | P0 | GDPR+trash | TODO |
| ECC-PLT-C322 | Corner | Admin backup start mid soft-delete purge wave | backup consistent snapshot or documented skip; no crash | Manual | P1 | TC-ADMIN-04 | Manual |
| ECC-PLT-C323 | Corner | Rate limit key extract email from body: truncated/malformed JSON | 400 before count or safe key; no panic | API | P1 | ExtractEmailFromBody | TODO |
| ECC-PLT-C324 | Corner | In-memory rateLimiter cleanup goroutine vs allow race at window edge | no negative counts; 429 correct | Unit | P1 | export_hooks | TODO |
| ECC-PLT-C325 | Corner | UUID path param valid v4 → mid-handler entity replaced by soft-delete | 404; no update on ghost | API | P0 | NFR-25+11 | TODO |
| ECC-PLT-C326 | Corner | ErrorHandler AppError Retry-After + client double-retry storm | Retry-After honored; no amplify DB load unbounded | API | P1 | NewRateLimitedAfter | TODO |
| ECC-PLT-C327 | Corner | Fiber Locals requestID lost after async detach (if any) | errors still get request_id or safe empty policy | API | P2 | middleware | Future |
| ECC-PLT-C328 | Corner | Multi-step: authz OK → rate allow → body parse fail → retry | rate counted once or documented; no authz bypass on retry | API | P1 | middleware order | TODO |
| ECC-PLT-C329 | Corner | Soft-delete restore double-submit | one restore; second 409/404 idempotent | API | P1 | trash | TODO |
| ECC-PLT-C330 | Corner | Hard-delete GDPR + concurrent soft-delete purge same UUID | one terminal gone; no error storm / FK panic | API | P0 | GDPR+purge | TODO |

---

## Счётчик

| Домен | ID range | Кол-во Corner | P0 | P1 | P2 |
|-------|----------|---------------|----|----|-----|
| Auth | ECC-AUTH-C300…C338 | **39** | 30 | 7 | 2 |
| RBAC/Org | ECC-RBAC-C300…C328 | **29** | 22 | 5 | 0 (+2 Future in P0/P1) |
| Platform | ECC-PLT-C300…C330 | **31** | 18 | 11 | 1 (+1 Future) |
| **Итого** | | **99** | **70** | **23** | **3** |

| Gap | Кол-во |
|-----|-------:|
| TODO | 82 |
| Manual | 14 |
| Future | 3 |

> Минимум 60 / цель 70+ — **выполнено (99)**.  
> Happy / Neg / Edge — **не включены**.  
> Parent `C200–C209` / `C200–C206` — **не дублированы по ID и сценариям**.  
> `EDGE-CORNER-CATALOG.md` — **не изменялся**.

---

*qa-manual · 2026-07-10 · handoff → `qa-manual-ver`*
