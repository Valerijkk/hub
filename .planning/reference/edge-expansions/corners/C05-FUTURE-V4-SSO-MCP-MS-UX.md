# Corner expansions C300+ — Future / v4 / SSO / MCP / MS / UX / XDOM

> **Тип:** только **Corner** (races, mid-flight, dual-write, permission shrink, cascade, NF-03 leak windows).  
> **Не Edge** (лимиты 50k/200/TTL — в `05-V4-AI-FUTURE.md` `*100+`).  
> **Не дублирует** parent `00-CORNER-PARENT.md` блоки H–N `*C200` дословно — углубление и комбинации.  
> **Gap:** только `Future` | `TODO`.  
> **Источники:** `ROADMAP-FINAL-2026-07-03.md` (W5–11, M2–M4, Track P/F) · `GRAY-ZONES-DECISIONS-2026-07-03.md` §2–§4 · `docs/vision_and_scope.md` US-WIKI / US-INTR / US-CAL-04 · ADR-AUTH-SSO / ADR-AI-agents (планируемые) · GZ-ARCH-09/14/16/17.  
> **Владелец:** `qa-manual` · 2026-07-10 · **каталог не трогать** · **не коммитить**.

Формат: `| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |`

---

## 1. v4 — Intranet / CRM / HR / Accounting / Bases (`ECC-V4-C300+`)

> Фокус: poll+moderation races · leave↔calendar busy · CRM partial import · period close · bases realtime@50k · coffee/onboarding mid-flight. Parent C200 = seed; здесь — комбинации и mid-flight.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-V4-C300 | Corner | Poll vote in-flight + admin soft-deletes post (US-INTR-03) | vote 404/410; counters not incremented; trash holds poll | API | P0 | US-INTR-01/03 | Future |
| ECC-V4-C301 | Corner | Two tabs vote same poll option simultaneously | one vote per user; count +1 not +2 | API | P0 | US-INTR-01 | Future |
| ECC-V4-C302 | Corner | Poll closes (deadline) while second vote request queued | late vote reject; first counted if before close | API | P1 | US-INTR-01 | Future |
| ECC-V4-C303 | Corner | Restore post-from-trash while new poll created on same slug | no ID collision; restored poll votes intact | API | P1 | soft-delete | Future |
| ECC-V4-C304 | Corner | Onboarding checklist assign + user completes last item same second as template edit | progress consistent; no phantom incomplete | API | P1 | US-INTR-02 | Future |
| ECC-V4-C305 | Corner | Random Coffee match job + both users deactivate mid-pair | no DM; rematch or skip; no orphan channel | API | P1 | US-INTR-05 | Future |
| ECC-V4-C306 | Corner | Kudos send + receiver GDPR anonymize mid-delivery | delivery fail-closed or anonymized display; leaderboard no PII leak | API | P0 | US-INTR-04, GDPR | Future |
| ECC-V4-C307 | Corner | Leave approve (HR) + calendar team-availability read race | busy only after approve commit; free until then | API | P0 | US-CAL-04 | Future |
| ECC-V4-C308 | Corner | Leave approve then revoke while calendar event already created for leave | busy cleared; event cancelled/tombstone; no stale busy | API | P0 | HR+CAL | Future |
| ECC-V4-C309 | Corner | Accrual job credits balance while user submits leave for full balance | one wins; no negative balance; 409/422 on loser | API | P0 | GZ-V4-15 | Future |
| ECC-V4-C310 | Corner | CRM CSV import: rows 1–499 OK, row 500 invalid, rows 501+ valid | documented partial policy: all-or-nothing **or** report+rollback; DB not half-corrupt | API | P0 | GZ-V4-10 | Future |
| ECC-V4-C311 | Corner | CRM import + concurrent edit of same email contact | merge/skip policy; version conflict surfaced; no silent overwrite | API | P1 | CRM | Future |
| ECC-V4-C312 | Corner | CRM deal stage change + company soft-delete mid-mutation | 404/409; deal not orphaned without company policy | API | P1 | CRM | Future |
| ECC-V4-C313 | Corner | Accounting period close + invoice create with posting date in closed period | reject; open period only; audit | API | P0 | GZ-V4-16 | Future |
| ECC-V4-C314 | Corner | Period close + reopen race from two admins | one terminal state; journal immutable while closed | API | P0 | ACC | Future |
| ECC-V4-C315 | Corner | Journal post + period close same second | post before close OK; after close reject; no partial ledger | API | P0 | ACC | Future |
| ECC-V4-C316 | Corner | Bases: 50k rows + BS-5 realtime WS + filter change mid-stream | p95 degrade OK; no cross-base leak; backpressure | Manual | P0 | GZ-V4-19, BS-5 | Future |
| ECC-V4-C317 | Corner | Bases share project-membership revoke while client has open WS on table | force unsubscribe; subsequent cell patch 403 | API/WS | P0 | GZ-SEC-08/12 | Future |
| ECC-V4-C318 | Corner | Bases formula recalc storm + concurrent cell edit on dependency | consistent final values or 409; no torn formula | API | P1 | PLAN-bases | Future |
| ECC-V4-C319 | Corner | Bases owner→project share migrate mid-edit (BS-1) | editors re-authz; no window where outsider reads | API | P0 | BS-1 | Future |
| ECC-V4-C320 | Corner | 1C export-bridge + Hub period close mid-batch | idempotent resume; no double postings | API | P1 | onec, GZ-V4-17 | Future |

**Подсчёт §1:** 21

---

## 2. Wiki future — restrict + search NF-03 (`ECC-WIKI-C300+`)

> US-WIKI-02/05 · GZ-SEC-07/11 · индекс lag vs restrict · issue-link visibility · version/history ACL. Parent C200 = seed.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-WIKI-C300 | Corner | Restrict ON while Meili/indexer still has old ACL snapshot | search must not return hit (re-filter at query or sync ACL); no title leak | API | P0 | US-WIKI-05, NF-03 | Future |
| ECC-WIKI-C301 | Corner | Restrict OFF (widen) + search before reindex | miss OK; never show to non-allow before; after reindex visible to space members | API | P1 | NF-03 | Future |
| ECC-WIKI-C302 | Corner | User removed from restrict allow-list while reading page + open history tab | subsequent GET/history 403; WS/page channel kill if any | API/E2E | P0 | US-WIKI-02/03 | Future |
| ECC-WIKI-C303 | Corner | Space write demoted to read mid-save (version N) | save 403; version unchanged; no partial body | API | P0 | NFR-24 | Future |
| ECC-WIKI-C304 | Corner | Move page into restrict-parent subtree | child inherits effective restrict; search drops for outsiders | API | P0 | US-WIKI-01/02 | Future |
| ECC-WIKI-C305 | Corner | Issue-link from task card: page becomes restricted for viewer | link hidden in issue UI (not 403-on-click leak of title) | E2E | P0 | US-WIKI-04 | Future |
| ECC-WIKI-C306 | Corner | Link page↔issue then issue project membership revoked | wiki side keeps link tombstone; no issue title leak to wiki-only user | API | P0 | NF-03 | Future |
| ECC-WIKI-C307 | Corner | Rollback to version N-2 while concurrent save creates N+1 | 409 or serialized; history append-only; no rewrite | API | P0 | US-WIKI-03 | Future |
| ECC-WIKI-C308 | Corner | Soft-delete page subtree + search still indexed | trash pages absent from Ctrl+K for all; restore reappears per ACL | API | P0 | NFR-11, NF-03 | Future |
| ECC-WIKI-C309 | Corner | Handbook space delete while HR handbook deep-link open | 404/tombstone UX; HR no crash; audit | E2E | P1 | GZ-V4-12/13 | Future |
| ECC-WIKI-C310 | Corner | System Admin opens restricted page (recovery) + Member searches same unique phrase | Admin OK+audit; Member 0 hits | API | P0 | GZ-SEC-07 | Future |
| ECC-WIKI-C311 | Corner | Restrict allow-list add user outside space | 422; cannot widen beyond space | API | P0 | US-WIKI-02 | Future |
| ECC-WIKI-C312 | Corner | Two admins: one removes last other admin while second demotes first | ≥1 admin invariant; at least one 422 | API | P0 | US-WIKI-02 | Future |
| ECC-WIKI-C313 | Corner | Autosave draft + explicit save race + restrict flip | draft not published to unauthorized; published version ACL-checked | API | P1 | US-WIKI-01 | Future |

**Подсчёт §2:** 14

---

## 3. SSO — enforce + break-glass + JIT/LDAP race (`ECC-SSO-C300+`)

> ADR-AUTH-SSO · GZ-SEC-03..06/10 · org toggle · founder break-glass · JIT∩LDAP deactivate. Parent C200 = seed.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-SSO-C300 | Corner | Enforce-SSO flip ON while non-founder Magic Link verify in-flight | verify reject after flip; no JWT issued | API | P0 | GZ-SEC-04/08 | Future |
| ECC-SSO-C301 | Corner | Founder break-glass Magic Link + parallel OIDC login same founder | one session set consistent; refresh reuse detection intact | API | P0 | break-glass | Future |
| ECC-SSO-C302 | Corner | JIT create user from OIDC + LDAP-sync deactivate same email same second | deactivated wins or ordered; no active zombie with JWT | API | P0 | JIT+LDAP | Future |
| ECC-SSO-C303 | Corner | LDAP-sync deactivate + user refresh token rotation race | refresh 401; access dead; WS kill | API | P0 | LDAP-sync | Future |
| ECC-SSO-C304 | Corner | LDAP-sync reactivate + old refresh still held by client | old refresh invalid; must re-login; no privilege revive via stale token | API | P0 | NFR-10 | Future |
| ECC-SSO-C305 | Corner | Org toggle `sso`→`both` while invite Magic Link email already sent | invite usable only if policy allows; else invalidate + resend path | API | P1 | invite | Future |
| ECC-SSO-C306 | Corner | Org toggle `both`→`sso` mid non-founder session | existing access until expiry per ADR; new Magic Link blocked; refresh policy documented | API | P1 | org-toggle | Future |
| ECC-SSO-C307 | Corner | OIDC callback + IdP returns email of pending Hub invite | merge to one UUID; no duplicate user; invite consumed | API | P0 | JIT | Future |
| ECC-SSO-C308 | Corner | AD group maps to Admin while Hub project-role was Member | system_role Admin; project roles unchanged (no auto escalate) | API | P0 | GZ-SEC-10 | Future |
| ECC-SSO-C309 | Corner | Break-glass attempt by Admin (not founder) under enforce-SSO | 403; audit security event | API | P0 | founder only | Future |
| ECC-SSO-C310 | Corner | LDAP bind timeout during sync job + concurrent OIDC login | fail-closed for new unknown; existing policy per ADR; no partial role wipe | API | P0 | LDAP | Future |
| ECC-SSO-C311 | Corner | Clock skew: OIDC `exp`/`nbf` at skew boundary during enforce | accept within skew; reject outside; no bypass of enforce | Unit | P1 | OIDC | Future |

**Подсчёт §3:** 12

---

## 4. MCP / AI agents — autonomy + mid-permission shrink (`ECC-MCP-C300+` / `ECC-AI-C300+`)

> GZ-SEC-09/20/21/22/23 · agent-principal ∩ user · stepped autonomy · PAT · BYOK hard-stop mid-loop.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MCP-C300 | Corner | Agent mid multi-step task: user loses project write after step 2 of 5 | steps 3–5 abort; no further writes; audit user+agent+task-id | API | P0 | agent-principal | Future |
| ECC-MCP-C301 | Corner | Autonomy: agent drafts issue OK then attempts delete without confirm | delete blocked; draft retained; confirm required | API | P0 | GZ-SEC-09/20 | Future |
| ECC-MCP-C302 | Corner | Confirm dialog open + admin shrinks agent scopes before confirm | confirm fails closed; no delete | API | P0 | autonomy | Future |
| ECC-MCP-C303 | Corner | PAT read-scope + tool chain tries messenger send | deny at tool gate; audit | API | P0 | GZ-SEC-23 | Future |
| ECC-MCP-C304 | Corner | Curated toolset hot-removed while agent session holds tool handle | next invoke 404/403; session not crash | API | P1 | MCP | Future |
| ECC-MCP-C305 | Corner | Two agents same user: one revoke PAT, other continues | revoked PAT dead; other PAT independent | API | P1 | PAT | Future |
| ECC-MCP-C306 | Corner | Agent on-behalf message + user blocks recipient mid-send | 403; no partial deliver | API | P0 | MSG+MCP | Future |
| ECC-MCP-C307 | Corner | Audit missing agent-id on on-behalf write | reject; no side-effect | API | P0 | GZ-SEC-21 | Future |
| ECC-AI-C300 | Corner | BYOK rotate mid streaming completion | stream abort soft; no old key in logs/traces; new key for retry | API | P0 | GZ-SEC-22 | Future |
| ECC-AI-C301 | Corner | Org daily hard-stop trips mid tool-loop after 3 tools | stop; usage counted; no 4th tool; alert | API | P0 | hard-stop | Future |
| ECC-AI-C302 | Corner | Virtual key budget vs org budget race two parallel streams | first to exhaust wins stop; no overage beyond tolerance | API | P0 | virtual keys | Future |
| ECC-AI-C303 | Corner | Model fallback cascade: primary fails mid-stream, secondary OK | seamless or clear restart; no duplicate side-effects from tools already run | API | P1 | LiteLLM | Future |
| ECC-AI-C304 | Corner | hub-ai split (M3): core still has stale route during flip + BYOK call | single writer; secret not in core pod; no dual charge | Infra/API | P0 | M3 | Future |

**Подсчёт §4:** 13

---

## 5. Microservices M2–M4 — route-flip / outbox / OTel (`ECC-MS-C300+`)

> M2 hub-notification · M3 hub-ai · M4 hub-ws/messenger/calls · Caddy flip · outbox dedup · OTel NATS gate · orphan UUID.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MS-C300 | Corner | Caddy route-flip `/api/notifications*` mid POST create: monolith accepts + split also accepts | idempotency key / Nats-Msg-Id → one notification; no dual-write corrupt | Manual | P0 | M2, GZ-ARCH-16/17 | Future |
| ECC-MS-C301 | Corner | Dual-run shadow: split consumer + monolith consumer both subscribed | only one side-effect writer; shadow read-only or dedup | Manual | P0 | strangler | Future |
| ECC-MS-C302 | Corner | Outbox: DB commit OK, NATS publish fails, retry with same Msg-Id | consumer sees once; at-least-once + dedup | API | P0 | outbox, BP-2026 | Future |
| ECC-MS-C303 | Corner | Outbox: publish OK, worker crashes before mark-sent; replay | dedup; no double email/push | API | P0 | outbox | Future |
| ECC-MS-C304 | Corner | Two outbox replicas SKIP LOCKED same row | one processes; no double publish | Unit | P0 | SKIP LOCKED | Future |
| ECC-MS-C305 | Corner | OTel: REST span OK but NATS header traceparent stripped | gate FAIL before M2 pilot; no silent break | Manual | P0 | M0-06, GZ-ARCH-14 | Future |
| ECC-MS-C306 | Corner | OTel collector down mid-flight during flip window | app continues; traces dropped/buffered per policy; flip not blocked by collector alone if gate pre-passed | Manual | P1 | OBS | Future |
| ECC-MS-C307 | Corner | NATS schema v2 field required, consumer still v1 during canary | DLQ/reject; no silent null defaults on required | Contract | P0 | GZ-QA-08/09 | Future |
| ECC-MS-C308 | Corner | hub-ws Redis membership lag: user removed in core, WS still subscribed | revalidate on event/heartbeat; force unsub (NF-02) | WS | P0 | M4 | Future |
| ECC-MS-C309 | Corner | hub-messenger Mongo split + `user.deleted` event lost | reconcile job tombstones; no PII in search | API | P0 | GZ-ARCH-09 | Future |
| ECC-MS-C310 | Corner | hub-calls LiveKit room exists, core call row deleted (orphan) | reconcile ends room or recreates row per runbook; no billing ghost | Manual | P0 | M4 LiveKit | Future |
| ECC-MS-C311 | Corner | Flip rollback: split→monolith while in-flight notification delivery | at-most-one user-visible notify; retry safe | Manual | P0 | M2 rollback | Future |
| ECC-MS-C312 | Corner | S2S bearer rotate mid M2 call from core→notification | brief 401 then retry with new; no data loss | Infra | P1 | GZ-SEC-13/15 | Future |
| ECC-MS-C313 | Corner | Path-filter CI: change only hub-notification, core tests skipped wrongly | path-filter correct; security packages still gated | CI | P1 | GZ-ARCH-16 | Future |

**Подсчёт §5:** 14

---

## 6. Redesign UX — optimistic UI / a11y races (`ECC-UX-C300+`)

> UX-polish · optimistic rollback 409 · theme/motion · focus · INP. Parent C200 = seed.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-UX-C300 | Corner | Optimistic Kanban move UI + server 409 version | card snaps back; toast; no ghost column | E2E | P0 | NFR-24, redesign | Future |
| ECC-UX-C301 | Corner | Optimistic RSVP accept + server 409 (concurrent decline other tab) | UI rollback to prior; inbox consistent | E2E | P0 | US-CAL, ECC-UX-C203↑ | Future |
| ECC-UX-C302 | Corner | Optimistic wiki save indicator + 403 mid-flight (ACL shrink) | editor marks failed; content not claimed saved | E2E | P0 | WIKI+UX | Future |
| ECC-UX-C303 | Corner | Theme toggle dark↔light during modal open + focus trap | contrast AA; focus remains trapped; no invisible focus | E2E | P1 | WCAG, NFR-16 | Future |
| ECC-UX-C304 | Corner | Reduced motion ON + optimistic rollback animation | state change still communicated (toast/text); not animation-only | E2E | P1 | a11y | Future |
| ECC-UX-C305 | Corner | Sticky redesign chrome + keyboard focus to control under header | Focus Not Obscured (2.4.11); scroll-into-view | E2E | P1 | WCAG 2.2 | Future |
| ECC-UX-C306 | Corner | Double-submit CTA while optimistic pending | second click noop/disabled; one request | E2E | P1 | INP | Future |
| ECC-UX-C307 | Corner | Offline banner + optimistic create then reconnect 409 | queue flush shows conflict UI; no silent duplicate | E2E | P1 | redesign | Future |
| ECC-UX-C308 | Corner | Screen reader announces optimistic success then server error | polite live-region corrects; no stuck «saved» | Manual | P1 | a11y | Future |

**Подсчёт §6:** 9

---

## 7. Cross-domain mega (`ECC-XDOM-C300+`)

> Комбинации Wiki×Search×SSO×MCP×MS×Project cascade — сильнее parent N.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-XDOM-C300 | Corner | Project delete cascade: issues + TMS runs + channels + scheduled calls + files + wiki issue-links + bases project-share | reconcile/tombstone all; no IDOR leftovers; search empty for project keys | Manual | P0 | GZ-ARCH-09, project.deleted | Future |
| ECC-XDOM-C301 | Corner | Project delete mid agent tool creating issue in that project | agent abort; issue not half-created or rolled back; audit | API | P0 | MCP+TRCK | Future |
| ECC-XDOM-C302 | Corner | SSO LDAP deactivate + open bases WS + calendar busy leave + wiki restrict page tab | all fail-closed within TTL; refresh 401; WS kill | Manual | P0 | SSO+V4+WIKI | Future |
| ECC-XDOM-C303 | Corner | Restrict wiki page indexed → search hit → flip enforce-SSO → user only had Magic Link | user loses session path; search/API 401; no content leak via cached Ctrl+K | E2E | P0 | WIKI+SSO+NF-03 | Future |
| ECC-XDOM-C304 | Corner | Calendar leave busy + CRM activity reminder + notification prefs mute | leave busy correct; reminder policy documented; security/HR still deliver if forced | API | P1 | CAL+CRM+NOTIF | Future |
| ECC-XDOM-C305 | Corner | M2 notification flip + calendar invite create + hub-mail timeout | event persisted; notification/mail retry via outbox; no duplicate ICS storm | API | P0 | M2+CAL+mail | Future |
| ECC-XDOM-C306 | Corner | Agent drafts accounting invoice in open period; admin closes period before confirm | confirm reject; no posting in closed period | API | P0 | MCP+ACC | Future |
| ECC-XDOM-C307 | Corner | GDPR user hard-delete while outbox has pending «mention» notifications + wiki authorship | anonymize authorship; outbox drops/anonymizes; legal-hold ACC/HR intact | API | P0 | GDPR, GZ-SEC-17/18 | Future |
| ECC-XDOM-C308 | Corner | FA-06 rotate JWT+LiveKit during M4 call + SSO session | reconnect with new creds; old tokens dead; call not false-ended if policy allows rejoin | Manual | P0 | FA-06, M4 | TODO |
| ECC-XDOM-C309 | Corner | Search aggregator (core) + wiki restrict + bases membership revoke same query window | results only permitted; no title leak from either domain | API | P0 | NF-03, GZ-ARCH-04 | Future |
| ECC-XDOM-C310 | Corner | Soft-delete 30d purge job + restore click + project.deleted reconcile same UUID second | one terminal state; no resurrect after purge; no orphan Mongo | API | P0 | NFR-11 | Future |
| ECC-XDOM-C311 | Corner | Optimistic UI accept calendar + LiveKit down + notification split flip | event OK; call provision retry; single invite notify after dedup | E2E/Manual | P0 | UX+CALL+MS | Future |

**Подсчёт §7:** 12

---

## Счётчик

| Домен | Prefix | Corner rows | Gap |
|-------|--------|------------:|-----|
| v4 Intranet/CRM/HR/Accounting/Bases | `ECC-V4-C300+` | **21** | Future |
| Wiki (restrict + NF-03) | `ECC-WIKI-C300+` | **14** | Future |
| SSO enforce/break-glass/JIT | `ECC-SSO-C300+` | **12** | Future |
| MCP + AI agents | `ECC-MCP-C300+` / `ECC-AI-C300+` | **13** | Future |
| Microservices M2–M4 | `ECC-MS-C300+` | **14** | Future |
| Redesign UX | `ECC-UX-C300+` | **9** | Future |
| Cross-domain mega | `ECC-XDOM-C300+` | **12** | Future/TODO |
| **Итого Corner** | | **95** | только Corner; Edge/Happy/Neg не включались |

**Минимум 70:** выполнен (95 ≥ 70).  
**Parent C200:** не копировался дословно — углубление mid-flight/комбинаций.  
**Основной каталог** `.planning/archive/july-2026/EDGE-CORNER-CATALOG.md`: не изменялся.  
**Коммит:** не создавался.

### Фокус-покрытие (чеклист запроса)

| Тема | Примеры ID |
|------|------------|
| Wiki restrict + search NF-03 | C300–C302, C308, C310; XDOM-C303/C309 |
| SSO enforce + break-glass + JIT deactivate race | C300–C304, C307, C309 |
| Agent autonomy mid-permission shrink | MCP-C300–C302; XDOM-C301/C306 |
| Caddy route-flip dual-write | MS-C300–C301, C311 |
| Outbox dedup | MS-C302–C304 |
| OTel break | MS-C305–C306 |
| Bases realtime + 50k | V4-C316–C319 |
| Leave + calendar busy | V4-C307–C309 |
| CRM import partial | V4-C310–C311 |
| Accounting period close | V4-C313–C315; XDOM-C306 |
| Optimistic UI 409 rollback | UX-C300–C302, C307 |
| Project delete cascade | XDOM-C300–C301, C310 |

---

*qa-manual 2026-07-10 · merge candidate after `qa-manual-ver` · QH-P2 TC authoring*
