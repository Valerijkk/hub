# Corner expansions wave3 — Auth / RBAC / Platform (C500+)

> **Тип:** только **Corner** (email-identity races, deploy/drain mid-auth, Redis failover, pending-7d storms, XFF/IP dual-stack rate, PWA/bfcache, audit/PII, Ent isolation, backup/imex∩RBAC).  
> **Не дублирует** parent (`ECC-*-C200…`), wave1 (`ECC-*-C300…`), wave2 (`ECC-*-C400…`) и Edge `01-AUTH-RBAC-PLATFORM.md` (`*100+`).  
> **ID:** `ECC-AUTH-C500+`, `ECC-RBAC-C500+`, `ECC-PLT-C500+`.  
> **Gap:** `TODO` | `Manual` | `Future`.  
> **Код-якоря:** `auth/service_session.go`, `auth/magiclink.go`, `auth/middleware.go`, `rbac/cache.go`, `rbac/middleware.go`, `platform/middleware/ratelimit.go`, pending purge (ORG-06), audit, imex/backup, trash jobs.  
> **Владелец:** `qa-manual` · 2026-07-10 · wave3 · каталог `EDGE-CORNER-CATALOG.md` **не трогать**.

Формат: `| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |`

---

## 1. Auth / Sessions — email-identity · drain · Redis failover · pending-7d — `ECC-AUTH-C500+`

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-AUTH-C500 | Corner | Profile email change commit while Magic Link code for **old** email still in Redis | verify old email reject; no session for stale identity; new email needs fresh send | API | P0 | email change∩code | TODO |
| ECC-AUTH-C501 | Corner | Magic Link send to `User@Ex.com` while pending row exists as `user@ex.com` | one canonical email; no dual pending/active; case-fold consistent | API | P0 | email normalize | TODO |
| ECC-AUTH-C502 | Corner | Homoglyph email (`аdmin@…` Cyrillic a) vs ASCII `admin@…` parallel register/invite | reject or distinct; no privilege merge; no founder collision | API | P0 | homoglyph IDN | Manual |
| ECC-AUTH-C503 | Corner | `user+tag@domain` alias send while base `user@domain` already member | policy: alias≠base or merge documented; no second user with same mailbox owner | API | P1 | plus-alias | TODO |
| ECC-AUTH-C504 | Corner | Two browsers open same Magic Link URL; Browser A verifies; B submits same code | A 200; B 400 spent; one session (or A+B only if multi-device policy after single code — fail-closed one) | API/E2E | P0 | multi-browser code | TODO |
| ECC-AUTH-C505 | Corner | PWA service worker serves stale access JWT after logout on other tab | next API 401; SW must not resurrect refresh from Cache API | E2E | P0 | PWA∩logout | Manual |
| ECC-AUTH-C506 | Corner | Browser bfcache restore of authenticated page after revoke-all | UI revalidates; no silent privileged mutations from restored form | E2E | P1 | bfcache | Manual |
| ECC-AUTH-C507 | Corner | Graceful shutdown: SIGTERM mid VerifyCode after Redis Del, before CreateSession | no orphan JWT; code spent; user re-send; process exits clean | API | P0 | drain∩verify | Manual |
| ECC-AUTH-C508 | Corner | Rolling deploy: pod A issues refresh; pod B has new JWT secret (FA-06 rotate window) | refresh fail-closed or dual-verify window documented; no alg-none accept | API | P0 | FA-06 deploy | Manual |
| ECC-AUTH-C509 | Corner | Redis primary failover mid storeRefreshToken (write OK on old primary, replica lag) | no dual live refresh; validate against authoritative store; fail-closed if ambiguous | API | P0 | Redis failover | Manual |
| ECC-AUTH-C510 | Corner | Redis Cluster slot migrate mid RefreshToken Get+Del+Set | at most one new chain; no token resurrection on both slots | API | P0 | cluster migrate | Manual |
| ECC-AUTH-C511 | Corner | Pending user day7−ε: admin Approve + PurgeExpiredPending same ms | active XOR purged; never active+deleted; audit one outcome | API | P0 | ORG-06∩approve | TODO |
| ECC-AUTH-C512 | Corner | Pending day7: purge job deletes user while Magic Link verify in-flight | verify 401/410; no JWT for purged pending | API | P0 | pending∩verify | TODO |
| ECC-AUTH-C513 | Corner | Invite accept creates pending; admin Reject + user re-send Magic Link same second | reject wins → no login; or re-queue policy; no active without approve | API | P0 | reject∩magic | TODO |
| ECC-AUTH-C514 | Corner | Telegram bind confirm while concurrent Logout all sessions | bind may complete on user id; sessions dead; telegram≠login session | API | P1 | telegram bind | TODO |
| ECC-AUTH-C515 | Corner | Telegram unbind mid notification delivery + concurrent GDPR erase | unbind/GDPR: no telegram_id on purged; delivery fail soft | API | P0 | telegram∩GDPR | TODO |
| ECC-AUTH-C516 | Corner | ListSessions cursor page1; revoke-all; page2 fetch | page2 empty/consistent; no revoked session payloads; no IDOR | API | P1 | sessions pagination | TODO |
| ECC-AUTH-C517 | Corner | Device UA/label PATCH while selective RevokeSession same device_id | revoke wins; label update 404/401; no zombie labeled session | API | P1 | device meta∩revoke | TODO |
| ECC-AUTH-C518 | Corner | Access JWT in query string (deep-link mistake) + Referer leak to third party | server rejects query JWT if unsupported; logs no full token; rotate if detected | API | P0 | token leak | Manual |
| ECC-AUTH-C519 | Corner | `Authorization: Bearer` + cookie refresh double-submit different users (shared browser profile) | identity from access wins consistently; no cross-user refresh bind | API | P0 | dual credential | TODO |
| ECC-AUTH-C520 | Corner | Onboarding incomplete: Magic Link re-verify on 2nd device mid onboarding wizard | both sessions see onboarding=false until complete; complete once; no skip ACL | API/E2E | P1 | onboarding∩multi-device | TODO |
| ECC-AUTH-C521 | Corner | Audit log write fail after successful login (NATS/DB audit down) | session still valid; audit retry/outbox; login not rolled back; no PII in fallback logs | API | P1 | audit∩login | TODO |
| ECC-AUTH-C522 | Corner | Mass revoke-all for 50 users (admin script) while each has refresh in-flight | all refreshes 401; no survivor; rate/backpressure no 500 storm | API | P0 | mass revoke | Manual |
| ECC-AUTH-C523 | Corner | Soft-deleted user in trash: admin restore + concurrent Magic Link send + old refresh present | old refresh dead; send only after restore policy; one identity | API | P0 | restore∩send | TODO |
| ECC-AUTH-C524 | Corner | Invite email rewritten (admin edits invite) after link already opened (validate OK) | accept uses final email or invalidate link; no account for stale email | API | P1 | invite edit mid | TODO |
| ECC-AUTH-C525 | Corner | PG read-replica lag: ListSessions sees old set after revoke on primary | revoke authoritative on refresh path; list may lag ≤bound then correct; no refresh success on revoked | API | P0 | replica lag | Manual |

---

## 2. RBAC / Org — pending storms · custom-role imex · Ent isolation — `ECC-RBAC-C500+`

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-RBAC-C500 | Corner | registration_mode flip `open→invite_only` while stranger Magic Link verify mid CreateUser | reject under new mode; no member without invite | API | P0 | mode∩verify | TODO |
| ECC-RBAC-C501 | Corner | Admin queue: Approve user A + Change system_role A→Admin parallel with A’s first API | either Member-then-Admin or Admin; never privilege without DB commit; cache ≤TTL | API | P0 | approve∩role | TODO |
| ECC-RBAC-C502 | Corner | Pending purge day7 deletes B while project invite for B email still valid | invite accept → pending/re-create policy; no ghost project membership on purged UUID | API | P0 | purge∩invite | TODO |
| ECC-RBAC-C503 | Corner | Two admins: Reject pending vs Approve pending same user | one terminal; second 409/410; audit single actor outcome | API | P0 | approve∩reject | TODO |
| ECC-RBAC-C504 | Corner | Custom role JSON import (imex) while members assigned old bitmask | import versioned; members remapped or denied; no open-perm on parse fail | API | P0 | imex∩custom_role | TODO |
| ECC-RBAC-C505 | Corner | Custom role export mid concurrent bitmask PATCH | export consistent snapshot XOR 409; no torn bitmask in file | API | P1 | imex race | TODO |
| ECC-RBAC-C506 | Corner | Clone custom role R1→R2 while delete R1 + reassign members | no member left on deleted role; R2 independent bits; audit | API | P0 | clone∩delete | TODO |
| ECC-RBAC-C507 | Corner | Ent tx: AddMember + RequirePermission on same user before commit visible | isolation: permission after commit only; no dirty read grant | API | P0 | Ent isolation | TODO |
| ECC-RBAC-C508 | Corner | Ent tx rollback after tentative membership insert (constraint fail) | no cache Set for rolled-back member; API 403 | API | P0 | rollback∩cache | TODO |
| ECC-RBAC-C509 | Corner | Project Owner transfer A→B while B GDPR hard-delete starts | transfer reject; project keeps A owner; no ownerless | API | P0 | transfer∩GDPR | TODO |
| ECC-RBAC-C510 | Corner | Bulk RemoveMember 30 users; InvalidateProject SCAN times out mid-way | DB removals durable; stale allow ≤TTL then deny; no permanent escalate | API | P0 | bulk∩cache | Manual |
| ECC-RBAC-C511 | Corner | Archive project + soft-delete project race (two admins) | one terminal archived/trashed; membership writes 403/404 | API | P0 | archive∩trash | TODO |
| ECC-RBAC-C512 | Corner | Unarchive while trash restore of same project UUID (illegal dual path) | exactly one live path; other 409/404; ACL single | API | P0 | unarchive∩restore | TODO |
| ECC-RBAC-C513 | Corner | SystemRole promote Member→Admin while target’s refresh rotates | new access still Member until promote commits; next request Admin; no gap escalate early | API | P0 | promote∩refresh | TODO |
| ECC-RBAC-C514 | Corner | Founder-only setting PATCH version N + second founder fantasy (forged) | only real founder 200; forged 403; version intact | API | P0 | founder IDOR | TODO |
| ECC-RBAC-C515 | Corner | Invite max-uses=5: 5 parallel accepts + 1 revoke of invite mid-flight | ≤5 members; revoked token rejects; uses atomic | API | P0 | accept storm | TODO |
| ECC-RBAC-C516 | Corner | Org setting `default_project_role` change mid invite accept | accept applies role at commit time documented; no silent Owner | API | P1 | default role flip | TODO |
| ECC-RBAC-C517 | Corner | Backup restore point-in-time: RBAC tables older than sessions tables | fail-closed authz; ops runbook; no Owner from stale backup alone | Manual | P0 | backup skew | Manual |
| ECC-RBAC-C518 | Corner | Imex project members CSV: row for deactivated user + parallel reactivate | membership follows active policy; deactivated cannot act; no duplicate rows | API | P1 | imex∩deactivate | TODO |
| ECC-RBAC-C519 | Corner | RequireAdmin handler: admin demoted after middleware pass, before side-effect email | side-effect not sent OR compensated; fail-closed preferred | API | P0 | TOCTOU admin | TODO |
| ECC-RBAC-C520 | Corner | Cross-project copy issue ACL check while source membership revoked mid-copy | copy abort 403; no dest issue with leaked fields | API | P0 | IDOR copy | TODO |
| ECC-RBAC-C521 | Corner | Custom role name unicode collision (NFC vs NFD) assign two roles | distinct IDs; no bitmask merge by normalized name alone | API | P1 | unicode roles | TODO |
| ECC-RBAC-C522 | Corner | Pending user approve creates org member while concurrent project-only invite accept | one user; memberships consistent; no double billing seat if counted | API | P1 | org∩project join | TODO |
| ECC-RBAC-C523 | Corner | LDAP group shrink (Future) + custom role local elevate same second | local elevate XOR LDAP policy documented; no silent max(bits) | API | P0 | LDAP∩local | Future |
| ECC-RBAC-C524 | Corner | SSO JIT provision (Future) while admin Deactivate same email pending | deactivate wins → no login; or JIT blocked; audit | API | P0 | JIT∩deactivate | Future |

---

## 3. Platform — XFF/IP · BodyLimit∩authz · shutdown · observability — `ECC-PLT-C500+`

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-PLT-C500 | Corner | Client spoofs `X-Forwarded-For` chain to reset auth 3/10m bucket | trusted proxy only; spoof ignored; real IP still limited | API | P0 | XFF∩NFR-05 | TODO |
| ECC-PLT-C501 | Corner | Same user IPv4 + IPv6 dual-stack: auth rate keys diverge | both keys limited independently or canonicalized; no unlimited via v6 | API | P0 | dual-stack rate | Manual |
| ECC-PLT-C502 | Corner | API 100/s per-IP: NAT shared 50 users burst | fairness/doc; 429 without unlocking auth bucket; no crash | API | P1 | NAT fairness | Manual |
| ECC-PLT-C503 | Corner | BodyLimit 25MB: multipart completes authz pass then body exceeds mid-stream | 413; no entity commit; no storage object orphan ACL | API | P0 | NFR-04∩authz | Manual |
| ECC-PLT-C504 | Corner | Presigned PUT (no Fiber BodyLimit) while membership revoked after presign | upload may land in Seaweed; app attach rejects; no ACL for revoked | API | P0 | NF-09∩RBAC | TODO |
| ECC-PLT-C505 | Corner | Graceful drain: in-flight POST past authz, SIGTERM before DB commit | connection finish or fail-closed; no half-row; LB removes instance | API | P0 | drain∩write | Manual |
| ECC-PLT-C506 | Corner | Rolling update: two app versions disagree on rate key schema | no permanent bypass; keys versioned/prefixed; 429 still enforced | API | P0 | deploy∩rate | Manual |
| ECC-PLT-C507 | Corner | Panic in handler after authz + partial DB write; Recover middleware | JSON error+request_id; tx rolled back; no privilege leak in stack to client | API | P0 | Recover∩tx | TODO |
| ECC-PLT-C508 | Corner | RequestID + audit row: audit insert fails, client got 200 | client truth vs audit eventually; ops alert; no PII in orphan logs | API | P1 | audit lag | TODO |
| ECC-PLT-C509 | Corner | Soft-delete purge job + imex export of same table concurrent | export omits purged or consistent snapshot; no resurrect via import of purged PK | API | P0 | purge∩imex | TODO |
| ECC-PLT-C510 | Corner | Imex import UUID collision with soft-deleted row in trash | reject or restore-path; no dual live+trash same UUID | API | P0 | imex∩trash | TODO |
| ECC-PLT-C511 | Corner | Optimistic lock version overflow / wrap fantasy on hot entity | no wrap to 0 grant; 409 or bigint safe | API | P2 | version wrap | TODO |
| ECC-PLT-C512 | Corner | UTC DB + FE DST fallback hour: duplicate local timestamps on create sort | order by UTC id/created_at; no missing row in day view | E2E | P1 | NFR-12 DST | TODO |
| ECC-PLT-C513 | Corner | Pagination offset: admin deletes item 0..n while client on page 3 | no skip-to-foreign-ACL item; 404 on direct fetch | API | P0 | NFR-14∩ACL | TODO |
| ECC-PLT-C514 | Corner | Files 10/min + parallel presign + Fiber upload paths | both count toward limit or documented; 11th 429; no bypass via presign-only | API | P0 | NFR-08 paths | TODO |
| ECC-PLT-C515 | Corner | Messages 60/min: HTTP send + WS send + retry with idempotency key | idempotent one message; rate counts attempts policy documented | API/WS | P0 | NFR-07∩idem | TODO |
| ECC-PLT-C516 | Corner | Redis rate memFallback on node A; node B still Redis-up — split brain counters | per-node undercount possible briefly; no infinite; after heal converge | API | P0 | HI-10 split | Manual |
| ECC-PLT-C517 | Corner | Health/ready probe bypasses auth but hits same DB pool as authed traffic | probes don’t starve; auth fail-closed under pool exhaust | API | P1 | probe∩pool | Manual |
| ECC-PLT-C518 | Corner | OpenTelemetry span export fail mid privileged handler | handler result unchanged; no span attribute dumps tokens/PII | API | P1 | OTEL∩PII | TODO |
| ECC-PLT-C519 | Corner | Backup job locks table mid RequirePermission+write | write waits/503; no authz skip; lock timeout fail-closed | API | P0 | backup lock | Manual |
| ECC-PLT-C520 | Corner | GDPR hard-delete user cascades while rate-limit keys for that user linger | keys expire/TTL; cannot mint new JWT; no PII in rate key material | API | P0 | GDPR∩rate keys | TODO |
| ECC-PLT-C521 | Corner | Concurrent trash restore of parent+child with FK: child first | order safe; no FK panic; ACL only after both live or documented partial | API | P1 | cascade restore | TODO |
| ECC-PLT-C522 | Corner | Caddy trusted proxy misconfig: client sets forged XFF, rate+audit IP wrong | detect/misconfig runbook; fail safe limit; audit notes proxy trust | Manual | P0 | Caddy∩XFF | Manual |
| ECC-PLT-C523 | Corner | SeaweedFS down after authz OK for attach metadata commit | metadata not committed without object XOR compensating delete; no ACL to missing blob | API | P0 | storage mid | TODO |
| ECC-PLT-C524 | Corner | Clock monotonic vs wall: purge watermark uses wall jump + rate window uses Redis PEXPIRE | each subsystem correct per policy; no cross-reset (purge≠rate unlock) | Manual | P0 | clock domains | Manual |

---

## Счётчик

| Домен | ID range | Кол-во Corner | P0 | P1 | P2 |
|-------|----------|---------------|----|----|-----|
| Auth | ECC-AUTH-C500…C525 | **26** | 18 | 8 | 0 |
| RBAC/Org | ECC-RBAC-C500…C524 | **25** | 20 | 5 | 0 (+2 Future) |
| Platform | ECC-PLT-C500…C524 | **25** | 18 | 6 | 1 |
| **Итого** | | **76** | **56** | **19** | **1** |

| Gap | Кол-во |
|-----|-------:|
| TODO | 52 |
| Manual | 22 |
| Future | 2 |

### Фокус-покрытие (wave3)

| Фокус | Пример ID | Кол-во (прибл.) |
|-------|-----------|----------------:|
| Email-identity / normalize / change mid-flight | AUTH-C500–C504, C524 | ≥6 |
| PWA / bfcache / multi-browser code | AUTH-C504–C506, C520 | ≥4 |
| Deploy / drain / FA-06 / Redis failover | AUTH-C507–C510, C525; PLT-C505–C506 | ≥8 |
| Pending 7d purge × approve/invite storms | AUTH-C511–C513; RBAC-C500–C503, C502, C515, C522 | ≥10 |
| Telegram bind≠login × GDPR | AUTH-C514–C515 | ≥2 |
| Custom-role imex / clone / unicode | RBAC-C504–C506, C521; PLT-C509–C510 | ≥6 |
| Ent isolation / TOCTOU admin | RBAC-C507–C508, C519 | ≥3 |
| XFF / dual-stack / NAT rate | PLT-C500–C502, C522 | ≥4 |
| BodyLimit / presign / storage mid | PLT-C503–C504, C514, C523 | ≥4 |
| Observability / audit / OTEL PII | AUTH-C521; PLT-C507–C508, C518 | ≥4 |
| Backup / GDPR × platform | RBAC-C517; PLT-C519–C520, C524 | ≥4 |

> Минимум 55 — **выполнено (76)**.  
> Happy / Neg / Edge — **не включены**.  
> Parent `C200+` / wave1 `C300+` / wave2 `C400+` — **не дублированы по ID и сценариям**.  
> `EDGE-CORNER-CATALOG.md` — **не изменялся**.  
> Коммит — **не создавался**.

---

*qa-manual · 2026-07-10 · wave3 · handoff → `qa-manual-ver`*
