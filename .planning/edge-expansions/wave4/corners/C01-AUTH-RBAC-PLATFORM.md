# Corner expansions wave4 — Auth / RBAC / Platform (C600+)

> **Тип:** только **Corner** (multi-pod/K3s split, PWA/SW offline replay, Telegram bind≠login, XFF/proxy rate-key, unicode/homoglyph email, WS upgrade∩revoke, audit/outbox mid-privileged, pending-7d expanded, secret-rotate multi-replica).  
> **Не дублирует** parent (`ECC-*-C200…`), wave1 (`ECC-*-C300…`), wave2 (`ECC-*-C400…`) и Edge `*100+`.  
> **ID:** `ECC-AUTH-C600+`, `ECC-RBAC-C600+`, `ECC-PLT-C600+`.  
> **Gap:** `TODO` | `Manual` | `Future`.  
> **Код-якоря:** `auth/service_session.go`, `auth/magiclink.go`, `auth/middleware.go`, `rbac/cache.go`, `rbac/middleware.go`, `platform/middleware/ratelimit.go`, WS upgrade auth, Telegram bind (≠login), pending purge job, Caddy `X-Forwarded-For`.  
> **Владелец:** `qa-manual` · 2026-07-10 · wave4 · каталог `EDGE-CORNER-CATALOG.md` **не трогать**.

Формат: `| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |`

---

## 1. Auth / Sessions — multi-pod · PWA/SW · Telegram bind · email unicode — `ECC-AUTH-C600+`

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-AUTH-C600 | Corner | K3s rolling: pod A stores refresh Redis; pod B (new image) Refresh mid deploy | refresh OK via shared Redis; no pod-local session assumption; JWT secret same | API | P0 | multi-pod refresh | Manual |
| ECC-AUTH-C601 | Corner | Two pods different JWT secret briefly (FA-06 rotate lag) access on A, refresh on B | fail-closed; no mint with mismatched secret; documented rotate window | API | P0 | FA-06 multi-replica | Manual |
| ECC-AUTH-C602 | Corner | Redis Sentinel failover mid RefreshToken (master→replica promote) | retry/5xx then OK; no dual live refresh chains; hash consistent | API | P0 | Redis failover | Manual |
| ECC-AUTH-C603 | Corner | Sticky session broken: logout hits pod A, next refresh hits pod B | refresh 401; Redis key gone globally; no zombie on B | API | P0 | sticky≠source of truth | TODO |
| ECC-AUTH-C604 | Corner | PWA Service Worker caches `/auth/refresh` 200; after revoke replays cached body | client must not treat stale 200 as live; SW cache bypass for auth; API still 401 on real call | E2E | P0 | PWA/SW auth | Manual |
| ECC-AUTH-C605 | Corner | Offline queue: 3 mutating API calls buffered; online after revoke-all | all fail 401; no replay privilege; queue cleared/reauth | E2E | P0 | offline replay | Manual |
| ECC-AUTH-C606 | Corner | Tab sleep 8h: access expired, SW wakes, parallel refresh from 2 tabs | one rotation wins; sibling pre-sleep refresh rejected; no dual chain | API | P0 | sleep∩rotation | TODO |
| ECC-AUTH-C607 | Corner | Browser restore session: IndexedDB holds old refresh after GDPR hard-delete | refresh 401; no PII in error; local storage wipe on 401 policy | E2E | P0 | GDPR∩PWA store | Manual |
| ECC-AUTH-C608 | Corner | Telegram bind start → user deactivate mid callback | bind reject; no telegram_id on deactivated; notify path dead | API | P0 | TG bind≠login | TODO |
| ECC-AUTH-C609 | Corner | Telegram unbind while notification worker delivers to same chat_id | unbind wins; no delivery after unbind commit; no re-bind via notify | API | P0 | TG unbind∩delivery | TODO |
| ECC-AUTH-C610 | Corner | Two devices start Telegram bind same user parallel (two deep-links) | one bind wins; other 409/410; single telegram_id | API | P0 | TG double-bind | TODO |
| ECC-AUTH-C611 | Corner | Telegram bind token reuse after success (replay link) | second bind 400/410; no hijack to other Hub user | API | P0 | TG token single-use | TODO |
| ECC-AUTH-C612 | Corner | Magic Link send to `User+tag@Domain.com` while account is `user@domain.com` | case/plus policy documented; one identity; no second user row | API | P0 | email normalize | TODO |
| ECC-AUTH-C613 | Corner | Homoglyph email `usеr@domain.com` (Cyrillic е) vs `user@domain.com` invite | reject lookalike or distinct; no silent merge; no founder confuse | API | P0 | homoglyph | TODO |
| ECC-AUTH-C614 | Corner | Profile email change requested → Magic Link still valid for old email | old code fail-closed XOR documented; new email needs new send; no dual inbox login | API | P0 | email change∩code | TODO |
| ECC-AUTH-C615 | Corner | Resend: new code supersedes; verify with old code after email lag 4m | old 400; only latest Redis code; rate 3/10m intact | API | P0 | code supersede | TODO |
| ECC-AUTH-C616 | Corner | Pending user day6.9: approve API + 7d purge job + Magic Link verify parallel | exactly one terminal: member XOR purged; no JWT for purged | API | P0 | pending-7d expanded | TODO |
| ECC-AUTH-C617 | Corner | Pending reject then re-invite same email before purge cron | new pending row or revive policy; old reject not block forever wrongly | API | P1 | reject∩re-invite | TODO |
| ECC-AUTH-C618 | Corner | Pending purge deletes user row while ListSessions admin open | list empty/404; no PII leak of purged pending | API | P0 | purge∩admin list | TODO |
| ECC-AUTH-C619 | Corner | WS upgrade with valid access; revoke-all mid Upgrade handshake | upgrade fail or immediate close; no subscribed rooms after revoke | WS | P0 | WS∩revoke | TODO |
| ECC-AUTH-C620 | Corner | WS connected; access TTL ends; refresh OK on HTTP; WS still old claims | WS re-auth/close policy; no privileged WS frames on expired access alone | WS | P0 | WS JWT TTL | TODO |
| ECC-AUTH-C621 | Corner | Device metadata spoof: ListSessions shows UA/IP; attacker PATCHes session row via IDOR | no client-writable session meta; IDOR 403/404 | API | P0 | session IDOR | TODO |
| ECC-AUTH-C622 | Corner | Concurrent `/me` profile PUT + admin GDPR delete | GDPR wins; PUT 401/404; no partial PII persist | API | P0 | /me∩GDPR | TODO |
| ECC-AUTH-C623 | Corner | Invite deep-link open while registration_mode flips to closed mid accept | accept reject; no half-member; deep-link error UX safe | API/E2E | P0 | deep-link∩mode | TODO |
| ECC-AUTH-C624 | Corner | Onboarding incomplete + Magic Link re-verify new device | new session OK; onboarding gate still blocks projects until complete | API | P1 | onboarding∩device | TODO |
| ECC-AUTH-C625 | Corner | Audit/outbox insert fail after successful login session create | session usable; audit retry; login not rolled back; no silent privilege hide | API | P1 | audit mid-login | TODO |

---

## 2. RBAC / Org — cascade demote · bitmask · multi-pod cache · pending — `ECC-RBAC-C600+`

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-RBAC-C600 | Corner | Pod A cache Owner bits; pod B Invalidate after demote; write hits A before TTL | ≤TTL stale then 403; no permanent escalate; DB deny on recheck path | API | P0 | multi-pod RBAC cache | TODO |
| ECC-RBAC-C601 | Corner | Bulk demote 50 members; NATS rbac.changed flood; one member write mid-storm | that write 403 if demoted; others eventual; no grant-all on lag | API | P0 | bulk demote storm | TODO |
| ECC-RBAC-C602 | Corner | Custom role bitmask all-1s / overflow edit while assigned to 10 users | reject invalid mask; no wrap-around grant; members unchanged on reject | API | P0 | bitmask edge | TODO |
| ECC-RBAC-C603 | Corner | Custom role A→B reassign mid RequirePermission for permission only in A | fail-closed; 403; no commit with A bits after B assign | API | P0 | role swap mid-flight | TODO |
| ECC-RBAC-C604 | Corner | Org Admin demoted to Member while approving pending queue item | approve abort 403; pending unchanged; audit no false approve | API | P0 | admin demote∩queue | TODO |
| ECC-RBAC-C605 | Corner | Project Owner cascade: remove from org while last project Owner | org remove blocked OR ownership transfer required; never zero-owner project | API | P0 | org∩project cascade | TODO |
| ECC-RBAC-C606 | Corner | Two admins: one archives project, other adds member same ms | add 403/404 if archive wins; no member on archived | API | P0 | archive∩add | TODO |
| ECC-RBAC-C607 | Corner | Invite to project P1 while user only member of P2; accept then P1 soft-deleted | accept fail or membership purged; no ACL on deleted P1 | API | P0 | invite∩project trash | TODO |
| ECC-RBAC-C608 | Corner | Pending approve creates member; parallel invite max-uses consume same email | one identity; uses/seat exact; no duplicate membership | API | P0 | approve∩invite | TODO |
| ECC-RBAC-C609 | Corner | SystemRole promote Member→Admin on pod A; RequireAdmin on pod B cache miss/hit mix | B loads DB; promote visible ≤TTL; no Admin API before commit | API | P0 | promote multi-pod | TODO |
| ECC-RBAC-C610 | Corner | Founder-only setting PATCH + Admin tries same setting parallel (not transfer) | founder wins/409; Admin 403; version monotonic | API | P0 | founder-only settings | TODO |
| ECC-RBAC-C611 | Corner | RemoveMember mid multipart + mid WebSocket room for same project | upload/WS both lose authz; room kick; no file ACL | API/WS | P0 | remove∩WS∩upload | TODO |
| ECC-RBAC-C612 | Corner | Cross-project UUID swap in batch role update body (20 ids, 1 foreign) | foreign 403; others apply or all-or-nothing documented; no partial escalate | API | P0 | batch IDOR | TODO |
| ECC-RBAC-C613 | Corner | Soft-deleted custom role restored while members already remapped to Viewer | restore does not auto-regrant old bits; explicit reassign | API | P0 | role restore bits | TODO |
| ECC-RBAC-C614 | Corner | registration_mode `invite_only` + open Magic Link for existing Member email | Member login OK; stranger still blocked; no role escalate on login | API | P0 | mode∩existing | TODO |
| ECC-RBAC-C615 | Corner | LDAP/SSO future: mapped group removed mid custom-role inheritance resolve | deny; no stale inherited grant after map shrink | API | P0 | SSO inherit | Future |
| ECC-RBAC-C616 | Corner | Audit write fail mid AddMember after DB membership commit | member exists; audit retry; UI may show member; no rollback leaving ghost deny forever | API | P1 | audit∩AddMember | TODO |
| ECC-RBAC-C617 | Corner | Permission cache key collision risk: userA projectX vs userB truncated key | isolated bits; no cross-user grant (key format invariant) | Unit | P0 | cache key | TODO |
| ECC-RBAC-C618 | Corner | Concurrent TransferOwnership A→B and TransferOwnership A→C on two pods | one owner terminal; never zero; loser 409 | API | P0 | dual transfer multi-pod | TODO |
| ECC-RBAC-C619 | Corner | Viewer opens issue draft UI; role→Owner mid-flight; submit with stale Viewer client | server authz Owner OK; no client-only gate trust | API/E2E | P1 | stale client role | TODO |
| ECC-RBAC-C620 | Corner | Pending 7d purge of user who was also invited as project Guest token live | purge clears invite bind; accept after purge 410; no ghost Guest | API | P0 | purge∩project invite | TODO |

---

## 3. Platform — XFF/proxy · WS · rate multi-pod · outbox — `ECC-PLT-C600+`

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-PLT-C600 | Corner | Client spoofs `X-Forwarded-For` chain to reset auth rate bucket per fake IP | Caddy trusted hops only; spoof ignored; 3/10m per real client | API | P0 | XFF spoof | Manual |
| ECC-PLT-C601 | Corner | Two pods: rate counter Redis; burst hits pod A then B same window | shared Redis count; total ≤limit; no per-pod double budget | API | P0 | multi-pod rate | TODO |
| ECC-PLT-C602 | Corner | Redis rate down on pod A (memFallback) while pod B still Redis | brief asymmetry documented; no unlimited on either; heal consistent | API | P0 | split rate backend | Manual |
| ECC-PLT-C603 | Corner | Proxy buffers full response after handler 429; client retries with stale Retry-After | server recompute; no negative; second 429 correct | API | P1 | proxy∩429 | Manual |
| ECC-PLT-C604 | Corner | Caddy timeout 30s; handler commits DB at 29s then client disconnect | no duplicate on retry if idempotent key/absent; ops reconcile | Manual | P1 | proxy timeout commit | Manual |
| ECC-PLT-C605 | Corner | WS upgrade behind Caddy: missing Upgrade headers mid auth middleware | 400/426; no partial room subscribe; rate not bypassed | WS | P0 | WS upgrade | TODO |
| ECC-PLT-C606 | Corner | WS + HTTP same user: message rate 60/min split across both paths on two pods | aggregate ≤60; no bypass via path/pod split | API/WS | P0 | NFR-07 multi-pod | TODO |
| ECC-PLT-C607 | Corner | Files 10/min: presigned PUT direct to SeaweedFS after API rate allow, then revoke membership | PUT may complete once; Hub ACL deny list/download after; no permanent ACL | API | P0 | presign∩revoke | TODO |
| ECC-PLT-C608 | Corner | BodyLimit 25MB + chunked Transfer-Encoding + Content-Length mismatch | reject; no OOM; no partial entity | API | P0 | body limit abuse | Manual |
| ECC-PLT-C609 | Corner | RequestID + async NATS publish after response: error in consumer needs correlation | consumer logs request_id/outbox id; no silent drop forever | API | P1 | outbox correlate | TODO |
| ECC-PLT-C610 | Corner | Outbox publish OK; consumer handler authz re-check user demoted | side-effect skip/deny; at-least-once idempotent; no privileged action | API | P0 | outbox∩demote | TODO |
| ECC-PLT-C611 | Corner | PG pool exhausted on pod A; pod B healthy; sticky sends all to A | 503 on A; no authz bypass; load shed; B unaffected | API | P0 | pool∩sticky | Manual |
| ECC-PLT-C612 | Corner | Soft-delete purge job on pod A; restore API on pod B same UUID day30−ε | XOR terminal; advisory lock/version; no resurrect after purge | API | P0 | purge multi-pod | TODO |
| ECC-PLT-C613 | Corner | Optimistic lock: writer on A version N, writer on B version N, both commit | one 200; other 409; Ent version truth | API | P0 | version multi-pod | TODO |
| ECC-PLT-C614 | Corner | UTC `deleted_at` + FE in TZ with DST spring-forward listing trash | item not vanish wrongly; sort by UTC; day bucket documented | E2E | P1 | DST∩trash | TODO |
| ECC-PLT-C615 | Corner | Pagination offset: admin lists users; GDPR delete between pages | no PII of deleted on page2; no 500 | API | P0 | GDPR∩pagination | TODO |
| ECC-PLT-C616 | Corner | ErrorHandler returns `{error,request_id}` while client also sends forged request_id header | body uses server id; logs server id; forge ignored | API | P1 | request_id forge | TODO |
| ECC-PLT-C617 | Corner | Rate key from email body: UTF-8 overlong / null byte in email field | 400; no panic; counter not poisoned | API | P0 | email key abuse | TODO |
| ECC-PLT-C618 | Corner | memFallback cleanup goroutine + allow at window edge on two pods (no Redis) | per-pod limits only if Redis down; documented; not silent prod mode | Unit | P0 | HI-10 dual mem | Manual |
| ECC-PLT-C619 | Corner | Backup job snapshot mid rate-limit Redis FLUSH on auth keys | auth re-login; backup not contain live refresh secrets plaintext | Manual | P0 | backup∩secrets | Manual |
| ECC-PLT-C620 | Corner | Fiber context cancel mid multi-step: authz OK → rate OK → handler canceled before write | no commit; rate may count; retry safe; no half-row | API | P0 | ctx cancel | TODO |
| ECC-PLT-C621 | Corner | Concurrent trash restore parent+child on two pods with FK order | no FK panic; order policy; both live or documented partial | API | P1 | cascade restore multi-pod | TODO |
| ECC-PLT-C622 | Corner | Health ready=true while Redis auth store read-only mid Sentinel | ready false or auth fail-closed; no 200 login without store | Manual | P0 | readiness∩Redis | Manual |

---

## Счётчик

| Домен | ID range | Кол-во Corner | P0 | P1 | P2 |
|-------|----------|---------------|----|----|-----|
| Auth | ECC-AUTH-C600…C625 | **26** | 22 | 4 | 0 |
| RBAC/Org | ECC-RBAC-C600…C620 | **21** | 18 | 2 | 0 (+1 Future) |
| Platform | ECC-PLT-C600…C622 | **23** | 16 | 5 | 0 |
| **Итого** | | **70** | **56** | **11** | **0** |

| Gap | Кол-во |
|-----|-------:|
| TODO | 48 |
| Manual | 20 |
| Future | 2 |

### Фокус-покрытие (wave4)

| Фокус | Пример ID | Кол-во (прибл.) |
|-------|-----------|----------------:|
| Multi-pod / K3s / Redis failover | AUTH-C600–C603; RBAC-C600, C609, C618; PLT-C601–C602, C611–C613, C618, C622 | ≥14 |
| PWA / SW / offline replay | AUTH-C604–C607 | ≥4 |
| Telegram bind ≠ login | AUTH-C608–C611 | ≥4 |
| Unicode / homoglyph / email change | AUTH-C612–C615; PLT-C617 | ≥5 |
| Pending 7d expanded | AUTH-C616–C618; RBAC-C604, C608, C620 | ≥6 |
| WS upgrade ∩ authz/revoke | AUTH-C619–C620; RBAC-C611; PLT-C605–C606 | ≥6 |
| XFF / proxy / rate multi-pod | PLT-C600–C604, C601 | ≥5 |
| Audit/outbox mid-privileged | AUTH-C625; RBAC-C616; PLT-C609–C610 | ≥4 |
| Bitmask / cascade demote | RBAC-C601–C603, C605, C613 | ≥5 |
| FA-06 / backup secrets | AUTH-C601; PLT-C619 | ≥2 |

> Минимум 55 — **выполнено (70)**.  
> Happy / Neg / Edge — **не включены**.  
> Parent `C200+` / wave1 `C300+` / wave2 `C400+` — **не дублированы по ID и сценариям**.  
> `EDGE-CORNER-CATALOG.md` — **не изменялся**.  
> Коммит — **не создавался**.

---

*qa-manual · 2026-07-10 · wave4 · handoff → `qa-manual-ver`*
