# Corner expansions C400+ (wave2) — Future / v4 / SSO / MCP / MS / UX / XDOM

> **Тип:** только **Corner** (triple-races, nested mid-flight, cascade×calendar, offline-queue×409, route-flip×outbox).  
> **Не Edge** (лимиты 50k/200/TTL — в `05-V4-AI-FUTURE.md` `*100+`).  
> **Не дублирует** parent `*C200` и wave1 `corners/C05…*C300` дословно — только углубление и новые комбинации.  
> **Gap:** только `Future` | `TODO`.  
> **Источники:** `ROADMAP-FINAL-2026-07-03.md` (W5–11, M2–M4) · `GRAY-ZONES-DECISIONS-2026-07-03.md` §2–§4 · `docs/vision_and_scope.md` US-WIKI / US-INTR / US-CAL-04 · ADR-AUTH-SSO / ADR-AI-agents (планируемые) · GZ-ARCH-09/14/16/17 · GZ-SEC-03..23 · PLAN-bases BS-5.  
> **Владелец:** `qa-manual` · 2026-07-10 · **каталог не трогать** · **не коммитить**.

Формат: `| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |`

---

## 1. v4 — Bases realtime@50k + Intranet/CRM/HR/ACC (`ECC-V4-C400+`)

> Фокус wave2: **bases realtime × 50k filter** mid-stream; leave×busy nested; CRM/ACC mid-flight. C300 = seed; здесь — filter-flip, backpressure, cascade share.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-V4-C400 | Corner | Bases @50k rows: client A applies filter F1; client B patches cell that drops out of F1 mid-WS frame | A receives remove/patch tombstone; no stale cell in filtered view; no cross-base leak | API/WS | P0 | BS-5, GZ-V4-19 | Future |
| ECC-V4-C401 | Corner | Bases @50k: filter change F1→F2 while 200 cell-patch events queued | backpressure; final view = F2 ∩ ACL; no torn rows from F1 | Manual | P0 | BS-5 | Future |
| ECC-V4-C402 | Corner | Bases @50k: two WS clients same table, opposite sort+filter; owner revokes project-share | both force-unsub; subsequent patches 403; no residual frames after revoke commit | API/WS | P0 | GZ-SEC-08/12 | Future |
| ECC-V4-C403 | Corner | Bases formula column depends on filtered-out row; concurrent edit of dependency while filter active | recalc consistent or 409; filtered client never sees foreign row body | API | P0 | PLAN-bases | Future |
| ECC-V4-C404 | Corner | Bases insert row 50_000 while realtime filter stream open; insert 50_001 | 50k OK + WS create; 50_001 → 400; stream stays healthy | API/WS | P0 | GZ-V4-19/20 | Future |
| ECC-V4-C405 | Corner | Bases view switch (grid→kanban) mid cell-edit + filter still applied | edit commits under ACL; view switch no duplicate WS subscribe leak | E2E | P1 | BS-5 views | Future |
| ECC-V4-C406 | Corner | Bases attachment field upload + row soft-delete mid-presign | PUT rejected/orphan GC; no attachment on deleted row; audit | API | P1 | storage+bases | Future |
| ECC-V4-C407 | Corner | Leave approve + calendar busy write + second leave overlapping same user same second | one busy block wins; second leave 409/422; no double-busy ghost | API | P0 | US-CAL-04, HR | Future |
| ECC-V4-C408 | Corner | Leave revoke while calendar expand RRULE job materializes leave-day instances | instances cancelled/tombstoned; team-availability free; no stale busy after job | API | P0 | HR+CAL | Future |
| ECC-V4-C409 | Corner | Accrual credit + leave submit + manager reject all same second | balance restored; no negative; audit trail ordered | API | P0 | GZ-V4-15 | Future |
| ECC-V4-C410 | Corner | CRM import batch + GDPR anonymize contact email mid-batch | row skipped/anonymized; report lists action; no PII in error dump | API | P0 | CRM, GDPR | Future |
| ECC-V4-C411 | Corner | Accounting period close + agent-drafted invoice confirm + reopen race (triple) | closed period immutable; confirm fail-closed; reopen only after reject; no ledger tear | API | P0 | ACC+MCP | Future |

**Подсчёт §1:** 12

---

## 2. Wiki — restrict × search NF-03 deep (`ECC-WIKI-C400+`)

> US-WIKI-02/05 · индекс lag × cursor × issue-link × history. Не копия C300 — nested ACL + search aggregator.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-WIKI-C400 | Corner | Restrict ON + Meili still indexing old ACL + Ctrl+K cursor page 2 already open | page-2 refetch re-filters; no title/snippet leak of restricted page | API/E2E | P0 | US-WIKI-05, NF-03 | Future |
| ECC-WIKI-C401 | Corner | Restrict allow-list remove user while search aggregator query in-flight (wiki+issues) | wiki hit dropped; issue hits unaffected; no partial title in wiki facet | API | P0 | NF-03, GZ-ARCH-04 | Future |
| ECC-WIKI-C402 | Corner | Move page into restrict-parent + concurrent search by outsider with exact unique phrase | 0 hits after move commit; miss-before-reindex OK; never hit | API | P0 | US-WIKI-01/02 | Future |
| ECC-WIKI-C403 | Corner | Issue-link card open + page restrict flipped for viewer + search still shows cached chip | chip hidden/tombstone; no 403-on-click title leak; cache TTL fail-closed | E2E | P0 | US-WIKI-04 | Future |
| ECC-WIKI-C404 | Corner | History GET version N-3 + restrict allow-list shrink mid-request | 403; body not returned; audit | API | P0 | US-WIKI-03 | Future |
| ECC-WIKI-C405 | Corner | Soft-delete restricted page + restore while outsider search phrase still in Meili | trash absent from Ctrl+K; restore visible only to allow-list; outsider 0 | API | P0 | NFR-11, NF-03 | Future |
| ECC-WIKI-C406 | Corner | Two spaces: restrict page A; copy/duplicate into space B without restrict | ACL re-eval on copy; B members see only if space ACL allows; no inherit leak from A allow-list | API | P0 | US-WIKI-02 | Future |
| ECC-WIKI-C407 | Corner | Autosave draft on restricted page + space write demote + explicit publish | publish 403; draft not promoted; version unchanged | API | P0 | NFR-24 | Future |
| ECC-WIKI-C408 | Corner | System Admin recovery open + Member WS page-channel still subscribed after demote from allow-list | Member channel kill; Admin OK+audit; Member search 0 | API/WS | P0 | GZ-SEC-07 | Future |

**Подсчёт §2:** 9

---

## 3. SSO — enforce × break-glass × JIT deactivate (`ECC-SSO-C400+`)

> ADR-AUTH-SSO · GZ-SEC-03..06/10 · triple: enforce flip × founder break-glass × LDAP/JIT deactivate.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-SSO-C400 | Corner | Enforce-SSO ON + founder break-glass Magic Link verify + LDAP-sync deactivate founder email same second | deactivate wins or ordered; no active JWT if deactivated; break-glass audit if session created then killed | API | P0 | break-glass, JIT+LDAP | Future |
| ECC-SSO-C401 | Corner | Enforce-SSO flip ON while non-founder OIDC callback in-flight and Magic Link verify also queued | Magic Link reject; OIDC may complete; no dual session for Magic path | API | P0 | GZ-SEC-04/08 | Future |
| ECC-SSO-C402 | Corner | JIT create from OIDC + LDAP deactivate + refresh rotation stolen-old-token race | user inactive; refresh 401; reuse detection still fires on stolen token | API | P0 | NFR-10, LDAP | Future |
| ECC-SSO-C403 | Corner | Break-glass founder session open + admin toggles org `sso`→`both`→`sso` rapidly | founder session policy per ADR; non-founder Magic blocked under sso; audit each toggle | API | P0 | org-toggle | Future |
| ECC-SSO-C404 | Corner | JIT maps AD→Admin; LDAP-sync removes group mid first API after login | second request fail-closed (Member/deny); no sticky Admin cache | API | P0 | GZ-SEC-10 | Future |
| ECC-SSO-C405 | Corner | Pending invite Magic Link + enforce-SSO ON + OIDC same email | single UUID; invite consumed via OIDC or documented merge; no duplicate | API | P0 | JIT, invite | Future |
| ECC-SSO-C406 | Corner | Break-glass attempt by Admin under enforce + parallel founder break-glass success | Admin 403+security audit; founder OK; no privilege confusion | API | P0 | founder only | Future |
| ECC-SSO-C407 | Corner | LDAP bind timeout during deactivate sync + user already has access+refresh | fail-closed for *new* unknown; existing: refresh killed on next successful sync or documented grace; no role wipe to empty | API | P0 | LDAP | Future |
| ECC-SSO-C408 | Corner | Enforce-SSO ON + founder break-glass + WS already connected from prior OIDC | WS re-auth or kill+reconnect; no orphan privileged WS after policy flip | WS | P0 | SSO+WS | Future |
| ECC-SSO-C409 | Corner | Clock skew at OIDC `nbf` boundary + enforce flip mid-callback | accept within skew only; enforce applied to issued tokens; no bypass | Unit/API | P1 | OIDC | Future |

**Подсчёт §3:** 10

---

## 4. MCP agents — permission shrink × audit (`ECC-MCP-C400+`)

> GZ-SEC-09/20/21/23 · agent-principal ∩ user · confirm dialog × shrink × audit completeness.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MCP-C400 | Corner | Agent steps 1–2 OK; user loses project write; step 3 confirm dialog still open; user clicks confirm | confirm fail-closed; no write; audit user+agent+task-id+denied | API | P0 | agent-principal, autonomy | Future |
| ECC-MCP-C401 | Corner | Admin shrinks agent scopes mid multi-tool loop after tool-2 success before tool-3 | abort; tools 3+ not run; audit lists completed+aborted; no silent continue | API | P0 | GZ-SEC-09/20 | Future |
| ECC-MCP-C402 | Corner | On-behalf write succeeds but audit sink down / agent-id stripped by proxy | reject or compensating rollback; **no** side-effect without audit (GZ-SEC-21) | API | P0 | GZ-SEC-21 | Future |
| ECC-MCP-C403 | Corner | PAT write-scope revoked while agent holds in-memory tool handle for delete | next invoke 401/403; prior draft retained; confirm cannot complete | API | P0 | GZ-SEC-23 | Future |
| ECC-MCP-C404 | Corner | Two agents same user: A shrinks project role via tool; B continues write in same project | B re-checks; fail-closed after A's commit; both audited | API | P0 | MCP×RBAC | Future |
| ECC-MCP-C405 | Corner | Agent drafts messenger message + recipient blocks user + permission shrink same second | 403; no partial deliver; audit both causes | API | P0 | MSG+MCP | Future |
| ECC-MCP-C406 | Corner | Curated toolset hot-remove + autonomy confirm for removed tool | confirm 404/403; no execute; session alive | API | P1 | MCP | Future |
| ECC-MCP-C407 | Corner | Agent creates issue then project membership revoked before link-wiki tool | wiki-link aborted; issue exists or rolled back per policy; audit chain complete | API | P0 | MCP+WIKI | Future |

**Подсчёт §4:** 8

---

## 5. AI — BYOK / hard-stop mid-loop (`ECC-AI-C400+`)

> GZ-SEC-22 · virtual keys · M3 hub-ai flip × budget.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-AI-C400 | Corner | BYOK rotate + org hard-stop trip same second mid tool-loop | stop; no 4th tool; old key absent from logs; usage counted once | API | P0 | GZ-SEC-22 | Future |
| ECC-AI-C401 | Corner | Virtual key budget exhaust + parallel stream on org budget | first stop wins; no overage beyond tolerance; both streams audited | API | P0 | virtual keys | Future |
| ECC-AI-C402 | Corner | M3 hub-ai route-flip mid streaming completion + BYOK call | single writer; no dual charge; secret not in core; stream abort soft | Infra/API | P0 | M3, GZ-ARCH-16 | Future |
| ECC-AI-C403 | Corner | Model fallback mid-stream after tool already mutated tracker | no duplicate mutation on fallback restart; tool idempotency key held | API | P0 | LiteLLM+MCP | Future |
| ECC-AI-C404 | Corner | Hard-stop alert path + notification split (M2) flip mid-alert | one alert delivered after outbox dedup; no alert storm | API | P1 | hard-stop+M2 | Future |

**Подсчёт §5:** 5

---

## 6. Microservices — route-flip × outbox × dedup (`ECC-MS-C400+`)

> M2–M4 · Caddy flip · outbox Msg-Id · SKIP LOCKED · OTel gate · rollback.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MS-C400 | Corner | Caddy route-flip mid POST: monolith commits outbox row; split also accepts same idempotency key | one Nats-Msg-Id; one user-visible notify; no dual-write | Manual | P0 | M2, GZ-ARCH-16/17 | Future |
| ECC-MS-C401 | Corner | Outbox publish OK → worker crash → route-flip → replay on other side | dedup by Msg-Id; at-most-one email/push | API | P0 | outbox, strangler | Future |
| ECC-MS-C402 | Corner | Two outbox replicas SKIP LOCKED + flip rollback split→monolith same row | one processor; no double publish across flip | Unit/Manual | P0 | SKIP LOCKED | Future |
| ECC-MS-C403 | Corner | Dual-run shadow consumer + real writer both see same event during canary | shadow read-only; only one side-effect; metrics diverge OK | Manual | P0 | strangler | Future |
| ECC-MS-C404 | Corner | OTel traceparent stripped on NATS during flip window after gate passed | runtime alert; no silent break; app continues; gate regression for next | Manual | P0 | M0-06, GZ-ARCH-14 | Future |
| ECC-MS-C405 | Corner | NATS schema v2 required field + canary consumer v1 + outbox retry storm | DLQ; no silent null; retries don't poison writer | Contract | P0 | GZ-QA-08/09 | Future |
| ECC-MS-C406 | Corner | hub-ws Redis membership lag + core project.deleted during M4 | force unsub all project rooms; reconcile; NF-02 hold | WS | P0 | M4, GZ-ARCH-09 | Future |
| ECC-MS-C407 | Corner | Flip rollback while in-flight calendar invite notification + mail outbox | at-most-one ICS/notify; retry safe; event row intact | Manual | P0 | M2+CAL | Future |
| ECC-MS-C408 | Corner | S2S bearer rotate mid core→notification during outbox drain | 401 then retry new bearer; Msg-Id unchanged; no loss | Infra | P1 | GZ-SEC-13/15 | Future |
| ECC-MS-C409 | Corner | hub-messenger Mongo split + lost `user.deleted` + search still indexed | reconcile tombstones; Ctrl+K no PII; outbox mention drops | API | P0 | GZ-ARCH-09 | Future |

**Подсчёт §6:** 10

---

## 7. Redesign UX — optimistic 409 × offline queue (`ECC-UX-C400+`)

> UX-polish · optimistic rollback · offline flush · a11y live-region. Не копия C300.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-UX-C400 | Corner | Optimistic Kanban move + offline mid-flight + reconnect gets 409 version | queue flush shows conflict UI; card snaps to server; no ghost column; no silent dup | E2E | P0 | NFR-24, redesign | Future |
| ECC-UX-C401 | Corner | Offline queue: create issue + create comment + server 409 on issue version after sync | comment not orphaned on wrong issue; conflict UI; queue order preserved | E2E | P0 | offline queue | Future |
| ECC-UX-C402 | Corner | Optimistic RSVP accept offline + other tab declines online + flush | rollback to decline/server; inbox consistent; toast | E2E | P0 | US-CAL | Future |
| ECC-UX-C403 | Corner | Optimistic wiki save + offline + ACL shrink before flush | flush 403; editor failed state; not claimed saved; draft local recoverable | E2E | P0 | WIKI+UX | Future |
| ECC-UX-C404 | Corner | Double-submit while offline queue draining optimistic pending | second click noop; one request after online | E2E | P1 | INP | Future |
| ECC-UX-C405 | Corner | Offline banner + optimistic success SR announce + server 409 on flush | live-region corrects to conflict; no stuck «saved» | Manual | P1 | a11y | Future |
| ECC-UX-C406 | Corner | Reduced motion ON + offline flush 409 rollback | state via toast/text not animation-only; focus preserved | E2E | P1 | a11y | Future |
| ECC-UX-C407 | Corner | Theme toggle during offline queue conflict modal open | contrast AA; focus trap holds; modal readable both themes | E2E | P1 | WCAG, NFR-16 | Future |
| ECC-UX-C408 | Corner | Optimistic bases cell edit @filtered view + 409 + offline reconnect | cell reverts; filter view consistent; no phantom row | E2E | P0 | BS-5+UX | Future |

**Подсчёт §7:** 9

---

## 8. Cross-domain mega — project delete cascade × calendar (`ECC-XDOM-C400+`)

> Project.deleted × calendar events × wiki/search/SSO/MCP/MS — сильнее C300.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-XDOM-C400 | Corner | Project delete cascade: issues + TMS + channels + **calendar events/scheduled calls** + files + wiki issue-links + bases share | all tombstoned/reconciled; calendar aggregator hides events; search empty; no IDOR leftovers | Manual | P0 | GZ-ARCH-09, project.deleted | Future |
| ECC-XDOM-C401 | Corner | Project delete while calendar invite accept in-flight + LiveKit room provisioning | accept 404/410; room ended/reconciled; no orphan egress; single cancel notify after dedup | API | P0 | CAL+CALL+MS | Future |
| ECC-XDOM-C402 | Corner | Project delete mid agent tool creating calendar event in project | agent abort; no half-event; audit user+agent+project.deleted | API | P0 | MCP+CAL | Future |
| ECC-XDOM-C403 | Corner | Project delete + leave-busy calendar blocks for project team channel events | channel events gone; personal leave busy untouched; no cross-wipe | API | P0 | HR+CAL+project | Future |
| ECC-XDOM-C404 | Corner | Project delete + wiki restrict pages with issue-links + search cursor open | links tombstone; search drops project keys+wiki hits; no title leak | API/E2E | P0 | WIKI+NF-03 | Future |
| ECC-XDOM-C405 | Corner | Project delete during M2 notification outbox drain for project invites | outbox drops/anonymizes; at-most-one cancel; no invite storm after delete | API | P0 | M2+outbox | Future |
| ECC-XDOM-C406 | Corner | SSO LDAP deactivate + project delete + open bases WS + calendar tab | refresh 401; WS kill; calendar reads 401/empty; fail-closed within TTL | Manual | P0 | SSO+V4+CAL | Future |
| ECC-XDOM-C407 | Corner | Soft-delete 30d purge + project.deleted reconcile + calendar restore click same UUID | one terminal state; no resurrect after purge; no orphan Mongo/LiveKit | API | P0 | NFR-11 | Future |
| ECC-XDOM-C408 | Corner | Optimistic offline queue has project calendar RSVP + project deleted before flush | flush 404/410; queue item failed UI; no resurrect event | E2E | P0 | UX+CAL | Future |
| ECC-XDOM-C409 | Corner | GDPR user hard-delete + project delete race + outbox mentions + calendar attendee rows | authorship anonymized; attendees anonymized; legal-hold ACC/HR intact; search clean | API | P0 | GDPR, GZ-SEC-17/18 | Future |

**Подсчёт §8:** 10

---

## Счётчик

| Домен | Prefix | Corner rows | Gap |
|-------|--------|------------:|-----|
| v4 / Bases realtime@50k | `ECC-V4-C400+` | **12** | Future |
| Wiki restrict × search | `ECC-WIKI-C400+` | **9** | Future |
| SSO enforce×break-glass×JIT | `ECC-SSO-C400+` | **10** | Future |
| MCP permission shrink×audit | `ECC-MCP-C400+` | **8** | Future |
| AI BYOK/hard-stop | `ECC-AI-C400+` | **5** | Future |
| MS route-flip×outbox×dedup | `ECC-MS-C400+` | **10** | Future |
| Redesign optimistic×offline | `ECC-UX-C400+` | **9** | Future |
| Cross-domain cascade×calendar | `ECC-XDOM-C400+` | **10** | Future |
| **Итого Corner** | | **73** | только Corner; Edge/Happy/Neg не включались |

**Минимум 65:** выполнен (73 ≥ 65).  
**C200/C300:** не копировались — wave2 углубление (triple-race, offline×409, cascade×calendar, filter@50k).  
**Основной каталог** `.planning/archive/july-2026/EDGE-CORNER-CATALOG.md`: не изменялся.  
**Коммит:** не создавался.

### Фокус-покрытие (чеклист запроса)

| Тема | Примеры ID |
|------|------------|
| SSO enforce×break-glass×JIT deactivate | SSO-C400–C404, C406–C408 |
| Agent×permission shrink×audit | MCP-C400–C404, C407; AI-C400/C403 |
| Route-flip×outbox×dedup | MS-C400–C403, C407–C408 |
| Wiki restrict×search | WIKI-C400–C405, C408; XDOM-C404 |
| Bases realtime×50k filter | V4-C400–C405, C408 UX |
| Redesign optimistic 409×offline queue | UX-C400–C405, C408; XDOM-C408 |
| Project delete cascade×calendar events | XDOM-C400–C403, C405, C408 |

---

*qa-manual 2026-07-10 · wave2 · merge candidate after `qa-manual-ver` · QH-P2 TC authoring*
