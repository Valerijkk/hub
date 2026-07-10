# Corner expansions wave2 — Auth / RBAC / Platform (C400+)

> **Тип:** только **Corner** (multi-device revoke races, invite+SSO future mix, founder transfer mid-flight, rate-limit+clock skew, soft-delete restore race day30).  
> **Не дублирует** parent (`ECC-*-C200…`), wave1 corners (`ECC-*-C300…`) и Edge `01-AUTH-RBAC-PLATFORM.md` (`*100+`).  
> **ID:** `ECC-AUTH-C400+`, `ECC-RBAC-C400+`, `ECC-PLT-C400+`.  
> **Gap:** `TODO` | `Manual` | `Future`.  
> **Код-якоря:** `auth/service_session.go`, `auth/magiclink.go`, `auth/middleware.go`, `rbac/cache.go`, `rbac/middleware.go`, `platform/middleware/ratelimit.go`, trash/purge jobs (NFR-11).  
> **Владелец:** `qa-manual` · 2026-07-10 · wave2 · каталог `EDGE-CORNER-CATALOG.md` **не трогать**.

Формат: `| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |`

---

## 1. Auth / Sessions — multi-device revoke · invite+SSO · clock skew — `ECC-AUTH-C400+`

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-AUTH-C400 | Corner | Device A+B+C live; revoke-all from A while B Refresh + C access in-flight | B refresh 401; C finishes at most once then 401; Redis refresh keys empty for user | API | P0 | multi-device revoke | TODO |
| ECC-AUTH-C401 | Corner | Revoke device B from A while B simultaneously RevokeSession(A) | both sessions dead; no survivor refresh; audit two revoke events | API | P0 | cross-revoke race | TODO |
| ECC-AUTH-C402 | Corner | Revoke-all vs new MagicLink verify on 4th device same second | new session XOR all-revoked; if verify wins after revoke-all → only new session live | API | P0 | login∩revoke-all | TODO |
| ECC-AUTH-C403 | Corner | ListSessions mid revoke-all (pagination of devices) | list never shows revoked as active; no IDOR of other users | API | P1 | sessions list | TODO |
| ECC-AUTH-C404 | Corner | Device B stolen refresh reused after A revoked-only-B (not all) | B 401 reuse/dead; A still refreshes; sibling devices intact | API | P0 | NFR-10 + selective revoke | TODO |
| ECC-AUTH-C405 | Corner | Revoke B → Redis Del OK → DB session row update fail | refresh dead by Redis; DB reconcile on next path; no zombie JWT mint | API | P0 | Redis/DB mid-revoke | TODO |
| ECC-AUTH-C406 | Corner | Revoke B → DB mark revoked OK → Redis Del fail | hash/status reject refresh; opportunistic Redis cleanup; no new access | API | P0 | DB-first revoke | TODO |
| ECC-AUTH-C407 | Corner | Three tabs same device_id: revoke-one vs refresh rotation storm | at most one live refresh chain; revoked chain never resurrects | API | P0 | device_id≠session | TODO |
| ECC-AUTH-C408 | Corner | Logout device A while admin RevokeSession(A) parallel | idempotent dead; no 500; refresh gone once | API | P1 | logout∩admin revoke | TODO |
| ECC-AUTH-C409 | Corner | Access JWT device B valid until exp after selective revoke B (no denylist) | policy: either short TTL window documented OR immediate 401 via session check; no silent privilege after revoke | API | P0 | JWT TTL vs revoke | Manual |
| ECC-AUTH-C410 | Corner | Invite token valid + org policy flips to `sso_only` mid accept (ADR-AUTH-SSO) | accept fail-closed; no member without SSO path; Magic Link reject for that email | API | P0 | invite+SSO mix | Future |
| ECC-AUTH-C411 | Corner | Pending invite email == JIT SSO first-login same second | one user row; invite consumed XOR JIT create; no dual identity | API | P0 | invite∩JIT | Future |
| ECC-AUTH-C412 | Corner | Magic Link verify while SSO session already live (policy `both`) | second method binds same user; no second founder/org; sessions both listed | API | P0 | dual-auth methods | Future |
| ECC-AUTH-C413 | Corner | Break-glass Magic Link founder while SSO IdP outage + revoke-all mid verify | founder gets session; revoke-all vs verify: documented winner; audit break-glass | API | P0 | founder break-glass | Future |
| ECC-AUTH-C414 | Corner | Invite accept step1 validate OK → SSO redirect → user cancels IdP → retry Magic Link | invite still valid or single-use spent policy; no half-member | API | P1 | invite multi-step SSO | Future |
| ECC-AUTH-C415 | Corner | LDAP sync demote mapped role while Magic Link refresh in-flight | refresh OK; next API 403 on elevated; session not privilege source | API | P0 | LDAP-sync∩session | Future |
| ECC-AUTH-C416 | Corner | Clock skew +2m: client sends refresh near server exp boundary | reject if past server exp; no extend via skewed iat | API | P0 | clock skew JWT | TODO |
| ECC-AUTH-C417 | Corner | NTP step −15m mid auth rate window after 3 MagicLink sends | 4th still 429 or window rules documented; no silent unlock | API | P0 | rate+clock skew | Manual |
| ECC-AUTH-C418 | Corner | NTP step +15m: auth fails counter TTL appears expired early | lockout must not vanish early without policy; fail-closed preferred | API | P0 | fails TTL skew | Manual |
| ECC-AUTH-C419 | Corner | Device A revoke-all; clock skew makes Redis PTTL on refresh keys look live | server wall/Redis TTL truth; revoke wins; no refresh after revoke | API | P0 | revoke∩PTTL skew | Manual |
| ECC-AUTH-C420 | Corner | Soft-deleted user day29 restore while devices still hold pre-delete refresh | restore ≠ resurrect old refresh; must re-verify; old Redis keys dead | API | P0 | soft-delete+session | TODO |
| ECC-AUTH-C421 | Corner | Soft-delete user day30 purge job vs concurrent Magic Link send | purge XOR restore-path; no JWT for purged; send 404/410 | API | P0 | day30∩auth | TODO |
| ECC-AUTH-C422 | Corner | Multi-device: revoke B and C in one bulk while A rotates refresh | A rotation completes or fail-closed; B/C dead; no cross-device token mix | API | P0 | bulk revoke | TODO |
| ECC-AUTH-C423 | Corner | Invite max-uses=2: SSO accept + Magic Link accept parallel same invite | exactly ≤2 members; no 3rd; uses atomic | API | P0 | invite+SSO race | Future |
| ECC-AUTH-C424 | Corner | Founder transfer initiated: old founder sessions stay until transfer commit | mid-flight old founder APIs OK until commit; after commit old founder demoted sessions policy | API | P0 | founder transfer∩session | TODO |
| ECC-AUTH-C425 | Corner | After founder transfer commit: revoke-all on old founder from new founder | old sessions dead; new founder sessions intact | API | P0 | post-transfer revoke | TODO |

---

## 2. RBAC / Org — founder transfer · invite+SSO · soft-delete — `ECC-RBAC-C400+`

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-RBAC-C400 | Corner | Founder transfer A→B: A confirms mid B deactivate | transfer abort; A remains sole founder; B not founder | API | P0 | founder transfer mid-flight | TODO |
| ECC-RBAC-C401 | Corner | Founder transfer A→B: commit OK → cache still shows A founder bits | ≤TTL then B-only founder; A founder APIs 403 after invalidate | API | P0 | cache∩transfer | TODO |
| ECC-RBAC-C402 | Corner | Founder transfer A→B while A edits org settings (version N) | settings 409 or transfer serializes; one founder writer | API | P0 | optimistic∩transfer | TODO |
| ECC-RBAC-C403 | Corner | Founder transfer A→B while parallel second transfer A→C | exactly one terminal founder; other reject; never zero founder | API | P0 | dual transfer race | TODO |
| ECC-RBAC-C404 | Corner | Founder transfer mid-flight: invite approve by A then commit to B | approve attributed; after commit A cannot approve; audit actor correct | API | P0 | admin queue∩transfer | TODO |
| ECC-RBAC-C405 | Corner | Founder transfer: B not yet member / pending only | reject transfer; founder uniqueness preserved | API | P0 | transfer preconditions | TODO |
| ECC-RBAC-C406 | Corner | Founder transfer commit → NATS rbac.changed fail | DB founder=B; retry event; API uses DB not event | API | P1 | NATS mid-transfer | TODO |
| ECC-RBAC-C407 | Corner | Founder soft-delete mid transfer A→B (A trashed before commit) | transfer fail-closed; restore A still founder XOR documented recovery | API | P0 | soft-delete∩transfer | TODO |
| ECC-RBAC-C408 | Corner | Invite accept + SSO JIT role map Owner while project Owner transfer race | project always ≥1 owner; invite role not escalate past policy | API | P0 | invite+SSO∩ownership | Future |
| ECC-RBAC-C409 | Corner | Org policy `sso_only`: existing Magic Link members keep sessions; new invite Magic Link blocked | sessions until revoke; new joins SSO-only; no privilege expand | API | P0 | policy flip | Future |
| ECC-RBAC-C410 | Corner | SSO group→custom role map shrink mid RequirePermission write | write 403; no commit with old bits | API | P0 | SSO map mid-flight | Future |
| ECC-RBAC-C411 | Corner | Invite revoke + SSO first-login with same email token race | no member if revoke wins; no orphan SSO user without policy | API | P0 | revoke∩JIT | Future |
| ECC-RBAC-C412 | Corner | Break-glass founder Magic Link while SSO-mapped Admin demotes custom roles | founder APIs always; demote cannot strip founder | API | P0 | founder invariant | Future |
| ECC-RBAC-C413 | Corner | Soft-deleted membership day29 restore vs new invite same user+project | one membership row; no duplicate bits; version consistent | API | P0 | soft-delete restore race | TODO |
| ECC-RBAC-C414 | Corner | Soft-deleted membership day30 purge vs restore click | purge XOR restore; after purge restore 404; no ghost ACL | API | P0 | day30 membership | TODO |
| ECC-RBAC-C415 | Corner | Project soft-delete day30 purge mid member role PATCH | 404/410; role not updated on purged project | API | P0 | project trash | TODO |
| ECC-RBAC-C416 | Corner | Custom role soft-deleted day30 purge while assigned members active | members remapped/deny; no ghost bitmask grant | API | P0 | custom_role purge | TODO |
| ECC-RBAC-C417 | Corner | Founder transfer A→B + RemoveMember(B) parallel | transfer reject if B removed; or remove reject if transfer holds lock | API | P0 | transfer∩remove | TODO |
| ECC-RBAC-C418 | Corner | Multi-device: admin revoke all sessions of demoted Owner mid ownership transfer | transfer completes with B; demoted A sessions dead; no write as Owner | API | P0 | revoke∩transfer | TODO |
| ECC-RBAC-C419 | Corner | Invite max-uses race: SSO accept + pending approve Magic Link same seat | seat accounting exact; no over-admit | API | P0 | invite+SSO seats | Future |
| ECC-RBAC-C420 | Corner | registration_mode `invite_only` → `sso_both` mid open invite link | accept follows new mode; stranger Magic Link still blocked if invite required | API | P1 | mode flip | Future |
| ECC-RBAC-C421 | Corner | Founder transfer rollback/cancel after step1 (two-phase) while B already got temp elevate | B never durable founder; A remains; audit cancel | API | P0 | two-phase transfer | TODO |
| ECC-RBAC-C422 | Corner | Soft-delete org setting entity day30 vs founder transfer commit | transfer not blocked by unrelated trash; settings consistent | API | P2 | trash∩org | TODO |

---

## 3. Platform — rate-limit+clock skew · soft-delete day30 — `ECC-PLT-C400+`

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-PLT-C400 | Corner | API 100/s window: NTP −30s after 99 hits then burst 50 | still capped; no full window reset bypass; 429 correct | API | P0 | rate+clock skew | Manual |
| ECC-PLT-C401 | Corner | Auth 3/10m + NTP +10m jump after 2 sends | 3rd may allow by wall; 4th must not exceed 3 in true 10m sliding/doc | API | P0 | NFR-05 skew | Manual |
| ECC-PLT-C402 | Corner | Files 10/min: clock skew + Redis PTTL disagree | server uses Redis TTL; no extra uploads via skewed client clock | API | P0 | NFR-08∩PTTL | Manual |
| ECC-PLT-C403 | Corner | Messages 60/min hit + NTP step mid WS frame batch | count monotonic; no negative remaining; 429/backpressure | API/WS | P0 | NFR-07 skew | Manual |
| ECC-PLT-C404 | Corner | RateLimitRedis window key expires early due to Redis clock skew vs app | memFallback or fail-closed; no unlimited | API | P0 | Redis clock | Manual |
| ECC-PLT-C405 | Corner | Dual bucket: auth 429 while API bucket reset by skew | auth stays limited; API independent; no cross-unlock | API | P0 | multi-bucket skew | TODO |
| ECC-PLT-C406 | Corner | Soft-delete entity at T+0; restore at day30 −1s vs purge job at day30 | restore XOR purge; winner audited; no resurrect after hard purge | API | P0 | day30 restore race | TODO |
| ECC-PLT-C407 | Corner | Soft-delete restore double-submit at day29.999 vs purge cron tick | one terminal; second 409/404; no FK panic | API | P0 | restore∩purge | TODO |
| ECC-PLT-C408 | Corner | Purge job batch 100; item 50 restored mid-batch | restored skipped; others purged; job continues | API | P0 | batch mid-restore | TODO |
| ECC-PLT-C409 | Corner | NTP jump makes deleted_at appear >30d early → purge runs | no early purge (monotonic/job watermark) or documented; no data loss | Manual | P0 | NFR-11 skew | Manual |
| ECC-PLT-C410 | Corner | NTP jump −2d: trash items look young; purge skips forever risk | watermark/absolute UTC; eventually purge; no eternal trash | Manual | P0 | eternal trash skew | Manual |
| ECC-PLT-C411 | Corner | Restore soft-deleted user-owned entity while user GDPR hard-deleted | restore deny; no PII rehydrate | API | P0 | GDPR∩restore | TODO |
| ECC-PLT-C412 | Corner | Soft-delete → restore → soft-delete again within day30 window | second trash TTL from new deleted_at; not original | API | P1 | TTL reset policy | TODO |
| ECC-PLT-C413 | Corner | Optimistic version update on entity restored mid concurrent PUT from stale client | 409 or 404; no write on wrong generation | API | P0 | version∩restore | TODO |
| ECC-PLT-C414 | Corner | Rate limit 429 Retry-After computed with skewed server clock | Retry-After ≥0 sane; client not told wait negative | API | P1 | Retry-After skew | TODO |
| ECC-PLT-C415 | Corner | memFallback rate limiter after Redis down: clock skew on process restart | counters reset policy documented; no permanent bypass across nodes | API | P0 | HI-10∩skew | Manual |
| ECC-PLT-C416 | Corner | Soft-delete project day30 purge while multi-device sessions still call project APIs | 404/410; sessions not privilege; RBAC cache miss deny | API | P0 | purge∩API | TODO |
| ECC-PLT-C417 | Corner | Trash restore of role-binding day30−ε while founder transfer commits | binding restore cannot recreate second founder; authz invariants hold | API | P0 | restore∩founder | TODO |
| ECC-PLT-C418 | Corner | Pagination trash list: item purged between page1 and page2 | no leak of purged payload; stable empty slot OK | API | P1 | NFR-14∩purge | TODO |
| ECC-PLT-C419 | Corner | Concurrent restore same UUID from two admins at day30 boundary | one 200; other 409/404; single live row | API | P0 | double restore | TODO |
| ECC-PLT-C420 | Corner | Rate window edge: allow last token + clock step back + immediate retry | no double-spend of same window slot beyond limit | API | P0 | token bucket skew | Manual |
| ECC-PLT-C421 | Corner | Soft-delete cascade children; restore parent only at day29 | children policy: stay purged or restore tree; no orphan ACL grant | API | P1 | cascade restore | TODO |
| ECC-PLT-C422 | Corner | Purge job + backup snapshot mid day30 wave | snapshot consistent or skip-in-trash documented; restore after backup not resurrect purged | Manual | P1 | backup∩purge | Manual |

---

## Счётчик

| Домен | ID range | Кол-во Corner | P0 | P1 | P2 |
|-------|----------|---------------|----|----|-----|
| Auth | ECC-AUTH-C400…C425 | **26** | 22 | 2 | 0 (+7 Future in set) |
| RBAC/Org | ECC-RBAC-C400…C422 | **23** | 18 | 2 | 1 (+8 Future) |
| Platform | ECC-PLT-C400…C422 | **23** | 16 | 5 | 0 |
| **Итого** | | **72** | **56** | **9** | **1** |

| Gap | Кол-во |
|-----|-------:|
| TODO | 42 |
| Manual | 15 |
| Future | 15 |

### Фокус-покрытие (wave2)

| Фокус | Пример ID | Кол-во (прибл.) |
|-------|-----------|----------------:|
| Multi-device revoke races | C400–C409, C422, C425; RBAC-C418 | ≥12 |
| Invite+SSO future mix | AUTH-C410–C415, C423; RBAC-C408–C412, C419–C420 | ≥14 |
| Founder transfer mid-flight | AUTH-C424–C425; RBAC-C400–C407, C417, C421; PLT-C417 | ≥12 |
| Rate-limit + clock skew | AUTH-C416–C419; PLT-C400–C405, C414–C415, C420 | ≥12 |
| Soft-delete restore race day30 | AUTH-C420–C421; RBAC-C413–C416; PLT-C406–C413, C416–C419, C421–C422 | ≥18 |

> Минимум 55 — **выполнено (72)**.  
> Happy / Neg / Edge — **не включены**.  
> Parent `C200+` / wave1 `C300+` — **не дублированы по ID и сценариям**.  
> `EDGE-CORNER-CATALOG.md` — **не изменялся**.  
> Коммит — **не создавался**.

---

*qa-manual · 2026-07-10 · wave2 · handoff → `qa-manual-ver`*
