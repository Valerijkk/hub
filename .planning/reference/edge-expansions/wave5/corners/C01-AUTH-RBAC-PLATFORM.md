# Corner expansions wave5 — Auth / RBAC / Platform (C700+)

> **Тип:** только **Corner** (client-abort/ctx-cancel, CSRF/Origin/CORS∩session, WS-upgrade∩revoke, maintenance/feature-flag mid-flight, avatar/PII∩GDPR, audit-retention purge, org↔project invite chains, seat/quota races, HTTP/2 multiplex identity, strangler auth-hop Future).  
> **Не дублирует** parent (`ECC-*-C200…`), wave1 (`ECC-*-C300…`), wave2 (`ECC-*-C400…`), wave3 (`ECC-*-C500…`), wave4 (`ECC-*-C600…` если есть) и Edge `01-AUTH-RBAC-PLATFORM.md` (`*100+`).  
> **ID:** `ECC-AUTH-C700+`, `ECC-RBAC-C700+`, `ECC-PLT-C700+`.  
> **Gap:** `TODO` | `Manual` | `Future`.  
> **Код-якоря:** `auth/middleware.go`, `auth/service_session.go`, `auth/magiclink.go`, `rbac/middleware.go`, `rbac/cache.go`, `platform/middleware/ratelimit.go`, WS upgrade, audit retention, trash/GDPR, apigateway (если есть), ADR-ARCH strangler.  
> **Владелец:** `qa-manual` · 2026-07-10 · wave5 · каталог `EDGE-CORNER-CATALOG.md` **не трогать**.

Формат: `| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |`

---

## 1. Auth / Sessions — abort · CSRF/Origin · WS-upgrade · maintenance · PII — `ECC-AUTH-C700+`

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-AUTH-C700 | Corner | Client TCP RST mid VerifyCode after Redis Get OK, before Del+CreateSession | no JWT issued; code still usable XOR spent+re-send policy; no half-session | API | P0 | abort∩verify | Manual |
| ECC-AUTH-C701 | Corner | Client cancel mid RefreshToken after new Redis SET, before old Del | at most one live chain; orphan new key cleaned on next path; no dual refresh | API | P0 | abort∩ME-02 | TODO |
| ECC-AUTH-C702 | Corner | Browser AbortController cancels Magic Link send after rate Incr, before email | rate counted; no code email if store failed; no unlock bypass | API | P1 | abort∩send | TODO |
| ECC-AUTH-C703 | Corner | HTTP/2 multiplex: Stream1 Refresh user A + Stream2 API with stale access user B same conn | identities never cross; responses bound to stream auth context | API | P0 | H2 multiplex | Manual |
| ECC-AUTH-C704 | Corner | Keep-alive connection: logout on tab1; tab2 reuses TCP with old Authorization | tab2 401; connection not sticky-privileged | API | P0 | conn reuse | TODO |
| ECC-AUTH-C705 | Corner | CSRF: cross-site POST Refresh with cookies, no/invalid Origin | reject if cookie-based refresh; Bearer-only path unaffected; no new tokens | API | P0 | CSRF∩refresh | TODO |
| ECC-AUTH-C706 | Corner | CORS preflight OK then credentialed Magic Link verify from disallowed Origin | verify rejected; no ACAO reflect attacker; no session Set-Cookie to evil | API | P0 | CORS∩verify | Manual |
| ECC-AUTH-C707 | Corner | `Sec-Fetch-Site: cross-site` + state-changing RevokeSession | reject or SameSite policy; no cross-site revoke | API | P0 | Sec-Fetch | TODO |
| ECC-AUTH-C708 | Corner | WS upgrade `Connection: Upgrade` with access JWT; selective revoke mid 101 handshake | upgrade fail or immediate close; no privileged WS without live session | WS | P0 | WS-upgrade∩revoke | TODO |
| ECC-AUTH-C709 | Corner | WS upgrade OK → first frame before authz room check; membership revoked | force close; NF-02; no subscribe | WS | P0 | WS∩NF-02 | TODO |
| ECC-AUTH-C710 | Corner | Maintenance mode ON mid Magic Link verify (feature flag flip) | fail-closed 503; code not silently spent without session XOR documented spend+retry | API | P0 | maintenance∩auth | TODO |
| ECC-AUTH-C711 | Corner | Feature flag `auth.magic_link` OFF while code already in Redis | verify reject; send reject; no JWT; ops can drain codes | API | P0 | flag∩code | TODO |
| ECC-AUTH-C712 | Corner | Read-only mode ON: Refresh allowed, CreateSession/login blocked | refresh OK for existing; new verify 503; no privilege expand | API | P1 | readonly∩login | TODO |
| ECC-AUTH-C713 | Corner | Avatar upload completes; GDPR hard-delete same user before Seaweed GC | avatar object deleted/orphaned-safe; no PII URL after erase; sessions dead | API | P0 | avatar∩GDPR | TODO |
| ECC-AUTH-C714 | Corner | Profile avatar PATCH + concurrent soft-delete user | 403/404; no new public avatar for trashed user | API | P0 | avatar∩trash | TODO |
| ECC-AUTH-C715 | Corner | Display name with RTL override + bidi spoof while admin ListSessions | stored safely; UI escape; audit shows canonical id not spoofed email | API/E2E | P1 | bidi∩PII | Manual |
| ECC-AUTH-C716 | Corner | Locale/TZ profile update mid open access; audit row timestamps | audit UTC; FE shows new TZ after refetch; no mixed TZ in one response | API | P1 | TZ∩audit | TODO |
| ECC-AUTH-C717 | Corner | hub-cli machine token (Future) + human Magic Link session same user parallel revoke-all | human sessions dead; machine token policy independent XOR also revoked; documented | API | P0 | cli∩session | Future |
| ECC-AUTH-C718 | Corner | PAT create (Future) while user deactivated mid-request | no PAT issued; deactivate wins | API | P0 | PAT∩deactivate | Future |
| ECC-AUTH-C719 | Corner | Org invite deep-link open → project invite accept same email before org membership commits | no project member without org member; order serialized; no orphan project ACL | API | P0 | invite chain | TODO |
| ECC-AUTH-C720 | Corner | Project invite accept while org pending not yet approved | reject or stay pending; no active project role before org approve | API | P0 | pending∩project | TODO |
| ECC-AUTH-C721 | Corner | Two org invites different emails → user merges via email change mid accept | one identity; second invite invalidated or rebound; no dual user | API | P0 | invite∩email merge | TODO |
| ECC-AUTH-C722 | Corner | Search indexer still has old email after profile email change + Magic Link to new | search user-scope no leak old; login only new; index lag OK without foreign hit | API | P0 | search∩email | TODO |
| ECC-AUTH-C723 | Corner | Soft-deleted user still in Meili/search; admin searches by email | no hit or trash-only admin path; no login via search deep-link | API | P0 | search∩trash | TODO |
| ECC-AUTH-C724 | Corner | GitHub/GitLab webhook actor mapped user revoked mid sync job | sync stops privileged writes; no issue create as revoked; audit | API | P0 | webhook∩revoke | TODO |
| ECC-AUTH-C725 | Corner | Context deadline exceeded in RequireAuth GetUserStatus after JWT parse OK | 503 fail-closed; no 200 with empty user | API | P0 | ctx∩RequireAuth | TODO |
| ECC-AUTH-C726 | Corner | Double Cookie header + Authorization Bearer mismatch identities | Bearer wins or reject ambiguous; never merge privileges | API | P0 | dual auth headers | TODO |
| ECC-AUTH-C727 | Corner | Service-worker background sync replays queued POST with pre-logout access | 401; SW clears queue; no mutation after logout | E2E | P0 | SW replay | Manual |
| ECC-AUTH-C728 | Corner | iOS/Android PWA: OS kills app mid refresh rotation; cold start with both tokens in Keychain | reuse detection or single survivor; no dual chain | Manual | P0 | mobile Keychain | Manual |
| ECC-AUTH-C729 | Corner | Audit retention job purges login rows while forensic export of same user runs | export consistent snapshot XOR empty after purge; no torn PII page | API | P1 | audit retention | TODO |
| ECC-AUTH-C730 | Corner | Strangler: auth hop to hub-auth MS (Future) times out; monolith fallback disagree on session | fail-closed; no JWT if either side says revoked | API | P0 | strangler auth | Future |

---

## 2. RBAC / Org — invite chains · seats · flag mid-authz · strangler — `ECC-RBAC-C700+`

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-RBAC-C700 | Corner | Org invite accept + project invite accept parallel same email (≠ wave3 storm) | org membership before project bits; no project-only ghost | API | P0 | org↔project chain | TODO |
| ECC-RBAC-C701 | Corner | Project invite revoked after org join but before project AddMember commit | no project membership; org member OK | API | P0 | revoke mid-chain | TODO |
| ECC-RBAC-C702 | Corner | Seat/quota max members: last seat taken by approve while invite accept races | ≤max; loser 409/429; no over-admit | API | P0 | seat race | TODO |
| ECC-RBAC-C703 | Corner | Seat freed by RemoveMember + two invites accept same ms | exactly one new member if one seat; atomic | API | P0 | seat free∩accept | TODO |
| ECC-RBAC-C704 | Corner | Feature flag `rbac.custom_roles` OFF mid AssignCustomRole | assign 503/404; existing bits policy documented; no open-perm | API | P0 | flag∩custom | TODO |
| ECC-RBAC-C705 | Corner | Maintenance mode: RequirePermission read OK, write deny mid Kanban | writes 503; reads per policy; no partial rank write | API | P0 | maint∩RBAC | TODO |
| ECC-RBAC-C706 | Corner | Client abort after RequirePermission pass, before handler commit | no DB write; cache unchanged; retry safe | API | P0 | abort∩authz | TODO |
| ECC-RBAC-C707 | Corner | HTTP/2: Stream1 demote user + Stream2 privileged write same user | write 403 if demote commits first; never escalate | API | P0 | H2∩demote | Manual |
| ECC-RBAC-C708 | Corner | CSRF cross-site POST AddMember with admin cookie session | reject; no member added | API | P0 | CSRF∩RBAC | TODO |
| ECC-RBAC-C709 | Corner | WS command `role.ping` / admin action over WS while system_role demoted | 403 close; no admin side-effect | WS | P0 | WS∩RequireAdmin | TODO |
| ECC-RBAC-C710 | Corner | Nested custom role “include role R1” fantasy (if unsupported) import | reject schema; no transitive escalate | API | P1 | nested roles | TODO |
| ECC-RBAC-C711 | Corner | Bitmask grant ALL bits via imex while founder-only bits reserved | founder bits not grantable; import reject/mask | API | P0 | imex∩founder bits | TODO |
| ECC-RBAC-C712 | Corner | Audit retention purge deletes role-change rows mid compliance export | export fail-closed or watermarked incomplete; no silent gap as full | API | P1 | audit∩RBAC | TODO |
| ECC-RBAC-C713 | Corner | Search ACL: user demoted; search still returns old issue hits from index | API filters user-scope; no title leak beyond NF-03; reindex eventual | API | P0 | search∩RBAC | TODO |
| ECC-RBAC-C714 | Corner | Webhook sync creates issue as bot mapped to removed project member | 403; no issue; mapping cleaned | API | P0 | webhook∩member | TODO |
| ECC-RBAC-C715 | Corner | Avatar of member used in UI mention; member GDPR erased mid render list | placeholder; no PII URL; membership gone | E2E | P1 | mention∩GDPR | Manual |
| ECC-RBAC-C716 | Corner | Org setting seat_limit PATCH version N while approve pending consumes seat | version/seat serialize; never exceed new limit | API | P0 | settings∩seat | TODO |
| ECC-RBAC-C717 | Corner | registration_mode `invite_only` + feature flag open registration Future conflict | documented precedence; stranger cannot verify | API | P0 | mode∩flag | Future |
| ECC-RBAC-C718 | Corner | Strangler: rbac check in monolith + hub-rbac MS disagree (Future) | deny if either denies; audit mismatch | API | P0 | strangler RBAC | Future |
| ECC-RBAC-C719 | Corner | Schema-per-service migration: membership row in old schema, cache in new (Future) | fail-closed until cutover complete; no grant-all | API | P0 | schema migrate | Future |
| ECC-RBAC-C720 | Corner | Owner transfer A→B while B’s only session is WS-upgrading | transfer OK; B authz after commit; WS re-auth if needed | API/WS | P1 | transfer∩WS | TODO |
| ECC-RBAC-C721 | Corner | Bulk CSV imex members: row seat overflow mid file | partial policy documented; no silent truncate past quota; transaction/rollback preferred | API | P0 | imex∩seats | TODO |
| ECC-RBAC-C722 | Corner | Project archived + org invite accept grants default_project_role into archived | no write membership on archived OR membership read-only; documented | API | P1 | archive∩invite | TODO |
| ECC-RBAC-C723 | Corner | Two admins: one soft-deletes custom role, other assigns it mid-flight | assign 404/409; no ghost bits | API | P0 | delete∩assign | TODO |
| ECC-RBAC-C724 | Corner | Context cancel in InvalidateProject SCAN mid demote storm | demote durable in DB; cache eventually deny; no panic | API | P1 | ctx∩Invalidate | Manual |
| ECC-RBAC-C725 | Corner | CORS-forbidden origin attempts RequireAdmin JSON POST | no admin mutation; no ACAO leak of error body with PII | API | P0 | CORS∩admin | TODO |
| ECC-RBAC-C726 | Corner | Pending user day3: project Owner invites to project; day7 purge pending | project invite cannot keep purged user; cleanup invite | API | P0 | pending∩proj invite | TODO |
| ECC-RBAC-C727 | Corner | SystemRole Admin→Member while admin’s hub-cli PAT (Future) in-flight call | call 403; PAT scopes re-checked | API | P0 | demote∩PAT | Future |
| ECC-RBAC-C728 | Corner | Locale spoof in Accept-Language changes error text mid IDOR probe | still 403/404 shape; no existence oracle via translated length alone | API | P1 | i18n∩IDOR | TODO |

---

## 3. Platform — abort · H2 · CORS · maint · audit-retention · strangler — `ECC-PLT-C700+`

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-PLT-C700 | Corner | Client abort mid BodyLimit read after 10MB of 25MB | connection drop; no entity; rate/authz not bypassed on retry | API | P0 | abort∩BodyLimit | Manual |
| ECC-PLT-C701 | Corner | Context deadline in rate middleware before handler | 429/503; counter not under-count to unlock | API | P0 | ctx∩rate | TODO |
| ECC-PLT-C702 | Corner | HTTP/2 RST_STREAM on one stream mid write; sibling streams same conn | sibling auth intact; aborted tx rolled back; no cross-stream data | API | P0 | H2 RST | Manual |
| ECC-PLT-C703 | Corner | CORS: allowed Origin list flip mid preflight→POST | POST enforces current list; no credentialed success to removed Origin | API | P0 | CORS flip | TODO |
| ECC-PLT-C704 | Corner | Maintenance mode ON: health/ready 200, mutating API 503, GETs policy | probes OK; no write; LB can drain | API | P0 | maint mode | TODO |
| ECC-PLT-C705 | Corner | Feature flag kill-switch for module mid paginated list | subsequent pages 503/empty policy; no partial ACL leak | API | P1 | flag∩pagination | TODO |
| ECC-PLT-C706 | Corner | Audit retention purge job + live request_id correlation lookup | lookup 404 after purge; live request still returns request_id; no PII resurrect | API | P1 | audit retention | TODO |
| ECC-PLT-C707 | Corner | Audit export CSV while retention deletes oldest partition | export consistent or fails; no torn rows mid-line | API | P1 | export∩purge | Manual |
| ECC-PLT-C708 | Corner | Soft-delete trash list + search reindex job for same UUID | search drops trashed; trash API shows item; no public search hit | API | P0 | search∩trash | TODO |
| ECC-PLT-C709 | Corner | Presigned URL issued; maintenance ON before client PUT | PUT may hit Seaweed; attach API 503; no ACL until maint off + recheck | API | P0 | maint∩presign | TODO |
| ECC-PLT-C710 | Corner | WS upgrade through Caddy while API rate bucket exhausted for IP | upgrade 429 or separate bucket documented; no unlimited WS bypass | API/WS | P0 | rate∩WS upgrade | TODO |
| ECC-PLT-C711 | Corner | Compression (gzip) request bomb claiming small Content-Length | reject/413; no OOM; authz not skipped | API | P0 | zip bomb | Manual |
| ECC-PLT-C712 | Corner | Content-Type `application/json` body is multipart smuggling attempt | 400; no dual parse privilege | API | P0 | CT confusion | TODO |
| ECC-PLT-C713 | Corner | Trusted proxy list reload mid XFF parse (Caddy/config SIGHUP) | no unlimited rate window; fail-safe trust previous or deny spoof | API | P0 | proxy reload | Manual |
| ECC-PLT-C714 | Corner | OpenTelemetry baggage injected with forged user_id | baggage never overrides JWT identity; no authz from baggage | API | P0 | OTEL baggage | TODO |
| ECC-PLT-C715 | Corner | Panic in rate limiter memFallback path under abort storm | Recover; fail-closed limit; no unlimited after panic | API | P0 | Recover∩rate | TODO |
| ECC-PLT-C716 | Corner | PG pool: long audit export holds conns; auth login queue | login 503/timeout fail-closed; no skip authz under pressure | API | P0 | pool∩auth | Manual |
| ECC-PLT-C717 | Corner | NATS outbox for audit full; privileged write succeeds | write OK; outbox backpressure alert; retry; no drop forever without ops | API | P1 | outbox full | TODO |
| ECC-PLT-C718 | Corner | Backup restore of platform settings mid live rate-limit config change | settings one terminal; rate keys may reset once documented; no permanent bypass | Manual | P0 | backup∩rate cfg | Manual |
| ECC-PLT-C719 | Corner | GDPR erase user + audit retention legal-hold flag on same id | legal-hold wins XOR erase redacts; documented; no raw PII if erase wins | API | P0 | GDPR∩legal hold | Future |
| ECC-PLT-C720 | Corner | Strangler: apigateway routes 10% to new MS (Future); session valid only on monolith | sticky or dual-validate; 401 on MS without session store; no privilege on empty | API | P0 | strangler % | Future |
| ECC-PLT-C721 | Corner | Seaweed avatar GC job + concurrent profile restore from trash | restore re-uploads or keeps object; no 404 avatar for restored user without policy | API | P1 | GC∩restore | TODO |
| ECC-PLT-C722 | Corner | Pagination + feature flag OFF for entity type between page1 and page2 | page2 empty/403; no foreign entity types leaked | API | P1 | flag∩pages | TODO |
| ECC-PLT-C723 | Corner | Client abort during optimistic lock retry loop (3 attempts) | no extra version bumps beyond failed attempts policy; no silent commit after abort | API | P0 | abort∩version | TODO |
| ECC-PLT-C724 | Corner | UTC leap-second / smeared second mid rate window and trash watermark | counters and purge remain correct; no mass unlock/purge | Manual | P1 | leap second | Manual |
| ECC-PLT-C725 | Corner | Dual Caddy→Fiber path: HTTP and HTTPS hosts disagree on cookie Secure | Secure cookies not sent on HTTP; no session fixation via HTTP host | API | P0 | Secure cookie | Manual |
| ECC-PLT-C726 | Corner | Request smuggling fantasy via malformed Transfer-Encoding to skip rate middleware | rejected by Caddy/Fiber; rate still applies | Manual | P0 | smuggling | Manual |
| ECC-PLT-C727 | Corner | hub-cli bulk API (Future) uses separate rate class vs browser | classes independent; browser 100/s not exhausted by cli XOR shared documented | API | P1 | cli rate class | Future |
| ECC-PLT-C728 | Corner | Observability sampling drops spans during auth failure burst | failures still rate-limited; metrics counters not sampled away; alerts fire | API | P1 | sampling∩429 | TODO |

---

## Счётчик

| Домен | ID range | Кол-во Corner | P0 | P1 | P2 |
|-------|----------|---------------|----|----|-----|
| Auth | ECC-AUTH-C700…C730 | **31** | 24 | 4 | 0 (+3 Future; Manual in set) |
| RBAC/Org | ECC-RBAC-C700…C728 | **29** | 20 | 5 | 0 (+4 Future) |
| Platform | ECC-PLT-C700…C728 | **29** | 18 | 8 | 0 (+3 Future; Manual in set) |
| **Итого** | | **89** | **62** | **17** | **0** |

| Gap | Кол-во |
|-----|-------:|
| TODO | 52 |
| Manual | 24 |
| Future | 13 |

### Фокус-покрытие (wave5)

| Фокус | Пример ID | Кол-во (прибл.) |
|-------|-----------|----------------:|
| Client abort / ctx cancel mid-pipeline | AUTH-C700–C702, C725; RBAC-C706, C724; PLT-C700–C701, C723 | ≥9 |
| HTTP/2 multiplex / RST / conn reuse | AUTH-C703–C704; RBAC-C707; PLT-C702 | ≥4 |
| CSRF / CORS / Origin / Sec-Fetch / Secure cookie | AUTH-C705–C707, C726; RBAC-C708, C725; PLT-C703, C725–C726 | ≥9 |
| WS-upgrade ∩ revoke / NF-02 / rate | AUTH-C708–C709; RBAC-C709, C720; PLT-C710 | ≥5 |
| Maintenance / feature-flag mid-flight | AUTH-C710–C712; RBAC-C704–C705, C717; PLT-C704–C705, C709, C722 | ≥10 |
| Avatar / PII / bidi / search index | AUTH-C713–C716, C722–C723; RBAC-C715; PLT-C708, C721 | ≥9 |
| Org↔project invite chains / seats | AUTH-C719–C721; RBAC-C700–C703, C716, C721, C726 | ≥10 |
| Audit retention / legal-hold / export | AUTH-C729; RBAC-C712; PLT-C706–C707, C719 | ≥5 |
| Webhook / hub-cli / PAT Future | AUTH-C717–C718, C724; RBAC-C714, C727; PLT-C727 | ≥6 |
| Strangler / schema-per-service Future | AUTH-C730; RBAC-C718–C719; PLT-C720 | ≥4 |
| Compression / CT / OTEL baggage / pool | PLT-C711–C716, C714 | ≥5 |

> Минимум 55 — **выполнено (89)**.  
> Happy / Neg / Edge — **не включены**.  
> Parent `C200+` / wave1 `C300+` / wave2 `C400+` / wave3 `C500+` / wave4 `C600+` — **не дублированы по ID и сценариям**.  
> `EDGE-CORNER-CATALOG.md` — **не изменялся**.  
> Коммит — **не создавался**.

---

*qa-manual · 2026-07-10 · wave5 · handoff → `qa-manual-ver`*
