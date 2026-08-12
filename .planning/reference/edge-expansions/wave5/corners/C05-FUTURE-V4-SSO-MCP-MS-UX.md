# Corner expansions C700+ (wave5) — Future / v4 / SSO / MCP / MS / UX / XDOM

> **Тип:** только **Corner** (quadruple races, partition×canary, clock-skew×multi-writer, legal-hold×purge×agent, PWA×offline×SSO, marketplace×bases ACL).  
> **Не Edge** (лимиты 50k/200/TTL — в edge `*100+`).  
> **Не дублирует** parent `*C200`, wave1 `*C300`, wave2 `*C400` (и любые `*C500`/`*C600` при появлении) дословно — только новые комбинации и углубление.  
> **Gap:** только `Future` | `TODO`.  
> **Источники:** `ROADMAP-FINAL-2026-07-03.md` (W5–11, M2–M4) · `GRAY-ZONES-DECISIONS-2026-07-03.md` §2–§4 · `docs/vision_and_scope.md` US-WIKI / US-INTR / US-CAL-04 · ADR-AUTH-SSO / ADR-AI-agents (планируемые) · GZ-ARCH-09/14/16/17 · GZ-SEC-03..23 · PLAN-bases BS-5 · marketplace/no-code.  
> **Владелец:** `qa-manual` · 2026-07-10 · **каталог не трогать** · **не коммитить**.

Формат: `| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |`

---

## 1. v4 — Bases partition / Intranet / CRM / HR / ACC (`ECC-V4-C700+`)

> Фокус wave5: **partition mid-WS**, formula×share migrate×filter, leave×RRULE×accrual quadruple, CRM merge×GDPR×import, ACC legal-hold×1C. Не копия C400 filter@50k seed.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-V4-C700 | Corner | Bases @50k: network partition A↔B 3s while A edits filtered cell and B changes filter F1→F2 | after heal: one version wins or 409; both views = ACL∩F2; no torn cell from F1 | Manual | P0 | BS-5, partition | Future |
| ECC-V4-C701 | Corner | Bases: owner→project share migrate + formula recalc storm + third client paste 200 cells | migrate ACL holds; paste backpressured; no outsider read during migrate window | API/WS | P0 | BS-1, PLAN-bases | Future |
| ECC-V4-C702 | Corner | Bases linked-record to row soft-deleted while kanban swimlane filter includes that link | tombstone chip; no body leak of deleted row; filter count consistent | API/E2E | P0 | soft-delete+bases | Future |
| ECC-V4-C703 | Corner | Bases rollup column + concurrent hard-cap insert at 50k + WS client on rollup view | insert 50_001 reject; rollup not inflated; stream healthy | API/WS | P0 | GZ-V4-19/20 | Future |
| ECC-V4-C704 | Corner | Marketplace template install creates base + project-share revoke on target project mid-install | install abort/rollback; no orphan base with open ACL; audit | API | P0 | marketplace+bases | Future |
| ECC-V4-C705 | Corner | Intranet poll + Random Coffee rematch + both voters GDPR erase same second | poll anonymized; coffee skip; no PII in DM draft or leaderboard | API | P0 | US-INTR, GDPR | Future |
| ECC-V4-C706 | Corner | Onboarding checklist complete last item + template archive + space handbook link break | progress terminal complete; handbook 404 UX; no incomplete ghost | API/E2E | P1 | US-INTR-02, WIKI | Future |
| ECC-V4-C707 | Corner | Leave approve + RRULE expand job + accrual debit + manager revoke (quadruple same second) | balance restored; instances tombstoned; team-availability free; ordered audit | API | P0 | US-CAL-04, HR | Future |
| ECC-V4-C708 | Corner | Two managers approve overlapping leave windows for same user via different devices | one 409/422; single busy block; no double debit | API | P0 | GZ-V4-15 | Future |
| ECC-V4-C709 | Corner | CRM dedupe merge A←B while import batch still writes email of B + deal stage move on A | merge policy wins; import skip/report; deal not orphaned; no dual UUID | API | P0 | CRM, GZ-V4-10 | Future |
| ECC-V4-C710 | Corner | CRM activity reminder outbox + contact legal-hold flag + GDPR anonymize request | legal-hold blocks anonymize; reminder suppressed or redacted; audit both | API | P0 | GDPR, GZ-SEC-17/18 | Future |
| ECC-V4-C711 | Corner | ACC period close + 1C webhook replay of closed-period posting + agent draft confirm | close immutable; 1C idempotent reject; agent confirm fail-closed; no ledger tear | API | P0 | ACC, onec, MCP | Future |
| ECC-V4-C712 | Corner | ACC reopen period + concurrent journal reverse + export-bridge mid-batch | one writer serialization; export resume idempotent; no double reverse | API | P0 | GZ-V4-16/17 | Future |

**Подсчёт §1:** 13

---

## 2. Wiki — restrict × export × search cursor deep (`ECC-WIKI-C700+`)

> US-WIKI · NF-03 · export/PDF × allow-list shrink × Meili lag × issue-link cache. Не копия C400 cursor page-2.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-WIKI-C700 | Corner | Restrict ON + PDF/export job already queued for outsider + allow-list shrink before render | job cancel/403; no PDF artifact with body; audit | API | P0 | US-WIKI-05, export | Future |
| ECC-WIKI-C701 | Corner | Nested restrict parent+child; move child out while outsider Ctrl+K phrase exact-match in-flight | after move: hit only if space ACL allows; never parent-restrict leak | API | P0 | US-WIKI-01/02, NF-03 | Future |
| ECC-WIKI-C702 | Corner | History diff N vs N-2 open + soft-delete page + restore to different space ACL | restore re-authz; outsider never gets diff body; Admin recovery audited | API | P0 | US-WIKI-03, NFR-11 | Future |
| ECC-WIKI-C703 | Corner | Issue-link preview SSR cache + restrict flip + search facet still warm | preview/chip tombstone; facet drop; no title in HTML cache after TTL fail-closed | E2E | P0 | US-WIKI-04, NF-03 | Future |
| ECC-WIKI-C704 | Corner | Two spaces copy page with attachments; source restrict allow-list not in dest | dest ACL independent; attachments re-ACL; no allow-list inherit | API | P0 | US-WIKI-02, storage | Future |
| ECC-WIKI-C705 | Corner | Autosave CRDT/OT merge + space write demote + explicit publish + Meili still indexing draft title | publish 403; draft not searchable; index purge draft title for outsiders | API | P0 | NFR-24, NF-03 | Future |
| ECC-WIKI-C706 | Corner | Handbook deep-link from HR + wiki space soft-delete + search cursor page-3 open | HR 404/tombstone; cursor refetch empty for space; no snippet leak | E2E | P0 | GZ-V4-12/13 | Future |
| ECC-WIKI-C707 | Corner | System Admin recovery edit + Member WS still subscribed after allow-list remove + aggregator query | Member WS kill; Admin OK+audit; Member search 0 including facets | API/WS | P0 | GZ-SEC-07 | Future |
| ECC-WIKI-C708 | Corner | Last space admin self-demote while second admin removes first from restrict allow-list | ≥1 admin invariant; restrict still enforced; at least one 422 | API | P0 | US-WIKI-02 | Future |

**Подсчёт §2:** 9

---

## 3. SSO — enforce × break-glass × IdP outage × JIT (`ECC-SSO-C700+`)

> ADR-AUTH-SSO · IdP outage × clock skew × device revoke × LDAP group storm. Не копия C400 deactivate×break-glass.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-SSO-C700 | Corner | Enforce-SSO ON + IdP 5xx during OIDC callback + founder break-glass Magic Link same window | founder break-glass OK+audit; non-founder no Magic; OIDC fail soft no partial user | API | P0 | break-glass, IdP | Future |
| ECC-SSO-C701 | Corner | JIT create + LDAP group storm add/remove Admin 10×/s + first API after login | final group state wins; no sticky Admin; each change audited or coalesced safely | API | P0 | GZ-SEC-10 | Future |
| ECC-SSO-C702 | Corner | Break-glass founder session + selective RevokeSession on that device + OIDC login other device | break-glass device dead; OIDC device OK; refresh reuse detection intact | API | P0 | sessions, founder | Future |
| ECC-SSO-C703 | Corner | Org `sso`→`both`→`sso` while pending invite Magic Link + OIDC same email in-flight | single UUID; invite consumed once; no dual active under conflicting policy | API | P0 | invite, JIT | Future |
| ECC-SSO-C704 | Corner | LDAP bind timeout + cached Admin mapping + enforce flip ON mid refresh | refresh fail-closed or documented grace; no role wipe to empty; new Magic blocked | API | P0 | LDAP, NFR-10 | Future |
| ECC-SSO-C705 | Corner | OIDC `nbf`/`exp` at skew boundary + IdP clock jump + enforce flip mid-token mint | accept only within skew; minted tokens respect enforce; no bypass | Unit/API | P0 | OIDC | Future |
| ECC-SSO-C706 | Corner | Admin (non-founder) attempts break-glass + founder break-glass + LDAP deactivate founder email | Admin 403+security audit; founder path ordered vs deactivate; no privilege confusion | API | P0 | founder only | Future |
| ECC-SSO-C707 | Corner | Enforce-SSO ON + WS from prior OIDC + refresh stolen-old-token reuse + LDAP deactivate | reuse detection fires; WS kill; user inactive; no orphan privileged WS | WS/API | P0 | SSO+WS, NFR-10 | Future |
| ECC-SSO-C708 | Corner | Map AD→system_role only while Hub project Admin concurrently granted via UI | system_role from AD; project Admin remains Hub-local; no silent strip unless policy says | API | P0 | GZ-SEC-10 | Future |
| ECC-SSO-C709 | Corner | Homoglyph IdP email vs existing Hub ASCII email under JIT | reject or distinct; no founder/Admin merge; security audit | API | P0 | homoglyph, JIT | Future |

**Подсчёт §3:** 10

---

## 4. MCP agents — multi-agent × MS flip × audit (`ECC-MCP-C700+`)

> GZ-SEC-09/20/21/23 · agent×agent conflict · confirm×route-flip · PAT×legal-hold. Не копия C400 shrink×confirm.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MCP-C700 | Corner | Agent A deletes issue; Agent B (same user) confirms wiki-link to that issue same ms | one terminal; B fail-closed if issue gone; both audited with agent-ids | API | P0 | MCP×WIKI | Future |
| ECC-MCP-C701 | Corner | Confirm dialog open + M2 route-flip mid confirm POST + permission shrink | at-most-one side-effect; confirm fail-closed if shrink wins; Msg-Id dedup | API/Manual | P0 | MCP+MS, autonomy | Future |
| ECC-MCP-C702 | Corner | On-behalf ACC invoice draft + period close + audit sink partition 2s | no posting without audit; compensating rollback if write raced; GZ-SEC-21 hold | API | P0 | GZ-SEC-21, ACC | Future |
| ECC-MCP-C703 | Corner | PAT write revoked + second PAT same user still valid + tool handle from revoked PAT | revoked handle 401; valid PAT independent; no cross-PAT session bleed | API | P0 | GZ-SEC-23 | Future |
| ECC-MCP-C704 | Corner | Agent creates calendar event + project delete + user loses write before RSVP tool | abort chain; no half-event; audit project.deleted+agent | API | P0 | MCP+CAL+project | Future |
| ECC-MCP-C705 | Corner | Curated toolset hot-remove + autonomy confirm for tool already mid-execute | in-flight may finish per policy XOR abort; no *new* invoke; session alive | API | P1 | MCP | Future |
| ECC-MCP-C706 | Corner | Agent messenger send + recipient GDPR erase + block list update same second | 403/anonymized; no partial deliver; audit causes | API | P0 | MSG+MCP+GDPR | Future |
| ECC-MCP-C707 | Corner | Legal-hold on ACC entity + agent confirm post + admin shrink scopes | confirm reject; legal-hold intact; audit denied+hold | API | P0 | legal-hold, MCP | Future |
| ECC-MCP-C708 | Corner | Stepped autonomy: draft OK → user confirms delete → admin demotes user mid confirm | confirm fail-closed; draft retained; no delete | API | P0 | GZ-SEC-09/20 | Future |

**Подсчёт §4:** 9

---

## 5. AI — BYOK / budget / M3 canary (`ECC-AI-C700+`)

> GZ-SEC-22 · virtual keys × hard-stop × canary dual-charge × tool idempotency. Не копия C400 rotate×hard-stop.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-AI-C700 | Corner | BYOK rotate + virtual-key budget exhaust + org hard-stop (triple) mid tool-loop | stop; no further tools; old key absent from logs; usage counted once | API | P0 | GZ-SEC-22 | Future |
| ECC-AI-C701 | Corner | M3 hub-ai canary 10% + BYOK call routed to both shadow and live writers | single charge; shadow read-only; secret not in core; stream abort soft on flip | Infra/API | P0 | M3, GZ-ARCH-16 | Future |
| ECC-AI-C702 | Corner | Model fallback after tool mutated CRM contact + GDPR anonymize same contact | no duplicate mutation on fallback; anonymize wins; tool idempotency key held | API | P0 | LiteLLM+CRM | Future |
| ECC-AI-C703 | Corner | Parallel streams: org budget 1 token left; both claim; hard-stop alert via M2 outbox | one stream stops; at-most-one alert after dedup; no overage beyond tolerance | API | P0 | hard-stop+M2 | Future |
| ECC-AI-C704 | Corner | Provider returns PII in completion while org DLP/redact policy ON mid-stream | stream redact/abort; no PII persisted in transcript store; audit | API | P0 | DLP, AI | Future |
| ECC-AI-C705 | Corner | Ollama local + BYOK cloud failover mid-completion after tracker comment tool | no duplicate comment; failover audited; keys not mixed in traces | API | P1 | Ollama+BYOK | Future |
| ECC-AI-C706 | Corner | Hard-stop trip + notification prefs mute AI alerts + security channel forced | user mute ignored for hard-stop/security; one alert; no storm | API | P1 | hard-stop+NOTIF | Future |

**Подсчёт §5:** 7

---

## 6. Microservices — canary × outbox × schema × rollback (`ECC-MS-C700+`)

> M2–M4 · dual-canary × SKIP LOCKED × schema v3 · Redis split-brain · S2S. Не копия C400 route-flip×Msg-Id.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MS-C700 | Corner | Dual canary: 5% split + 5% shadow + monolith writer; same idempotency key POST | one side-effect; shadows read-only; metrics may diverge | Manual | P0 | strangler, M2 | Future |
| ECC-MS-C701 | Corner | Outbox row claimed SKIP LOCKED on split; flip rollback to monolith before mark-sent | monolith resume or dedup; at-most-one email/push; no lost row | Manual | P0 | outbox, rollback | Future |
| ECC-MS-C702 | Corner | NATS schema v3 required enum + canary v2 consumer + retry storm during flip | DLQ; no silent default; writer not poisoned; gate alert | Contract | P0 | GZ-QA-08/09 | Future |
| ECC-MS-C703 | Corner | Redis split-brain: hub-ws membership on A, core revoke on B, heal after 2s | force unsub on heal; NF-02; no privileged residual frames | WS | P0 | M4, GZ-ARCH-09 | Future |
| ECC-MS-C704 | Corner | OTel baggage/tracestate truncated across NATS during canary window after gate pass | runtime alert; app continues; next gate fails until fixed | Manual | P0 | M0-06, GZ-ARCH-14 | Future |
| ECC-MS-C705 | Corner | hub-messenger Mongo primary stepdown mid `user.deleted` consumer + search still hot | reconcile tombstones; Ctrl+K no PII; mention outbox drops | API | P0 | GZ-ARCH-09 | Future |
| ECC-MS-C706 | Corner | hub-calls LiveKit room OK; core call soft-deleted; flip M4 mid reconcile | room ended XOR row restored per runbook; no billing ghost; single cancel notify | Manual | P0 | M4 LiveKit | Future |
| ECC-MS-C707 | Corner | S2S bearer rotate + clock skew between core and notification + outbox drain | 401 then retry; Msg-Id stable; no loss; skew within tolerance only | Infra | P0 | GZ-SEC-13/15 | Future |
| ECC-MS-C708 | Corner | Path-filter CI skips core while notification contract breaking change merges | contract gate/CI fails; no silent pilot; security packages still gated | CI | P1 | GZ-ARCH-16 | Future |
| ECC-MS-C709 | Corner | Calendar invite + mail outbox + notification flip + hub-mail 429 storm | event OK; retries backoff; at-most-one ICS after dedup; no invite storm | Manual | P0 | M2+CAL+mail | Future |

**Подсчёт §6:** 10

---

## 7. Redesign UX — PWA / offline / a11y / multi-tab (`ECC-UX-C700+`)

> UX-polish · PWA SW × offline queue × multi-tab 409 × focus. Не копия C400 offline×409 Kanban.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-UX-C700 | Corner | PWA SW caches optimistic «saved» wiki HTML; ACL shrink; offline flush 403 | SW must not serve privileged body; editor failed; draft local recoverable | E2E | P0 | PWA+WIKI | Future |
| ECC-UX-C701 | Corner | Multi-tab: Tab A optimistic Kanban move; Tab B moves same card; both offline then flush | one 409; both UIs converge to server; no ghost duplex columns | E2E | P0 | NFR-24, offline | Future |
| ECC-UX-C702 | Corner | Offline queue: create base row + formula-dependent edit + reconnect at 50k cap | create may 400; dependent edit not applied on missing row; conflict UI ordered | E2E | P0 | BS-5+UX | Future |
| ECC-UX-C703 | Corner | Optimistic RSVP + bfcache restore of inbox after revoke-all sessions | revalidate; no privileged mutation from restored form; toast login | E2E | P0 | bfcache, CAL | Future |
| ECC-UX-C704 | Corner | Offline banner + SR live-region success + server 409 + reduced motion ON | live-region corrects; state via text not animation-only; focus preserved | Manual | P1 | a11y | Future |
| ECC-UX-C705 | Corner | Theme toggle + high-contrast OS + conflict modal during offline drain | contrast AA; focus trap; modal readable; no invisible controls | E2E | P1 | WCAG, NFR-16 | Future |
| ECC-UX-C706 | Corner | Double-submit + keyboard Enter spam while queue draining optimistic pending | one request; second noop; INP OK | E2E | P1 | INP | Future |
| ECC-UX-C707 | Corner | Sticky chrome + Focus Not Obscured during conflict modal open on mobile viewport | scroll-into-view; focus visible; modal not under header | E2E | P1 | WCAG 2.2 | Future |
| ECC-UX-C708 | Corner | Optimistic CRM stage move + GDPR anonymize contact mid-flight + reconnect | UI rollback/tombstone; no PII flash after anonymize; toast | E2E | P0 | CRM+UX+GDPR | Future |

**Подсчёт §7:** 9

---

## 8. Cross-domain mega — cascade × SSO × MCP × MS × UX (`ECC-XDOM-C700+`)

> Project/org storms × calendar × wiki × search × agent × flip — сильнее C400 cascade×calendar.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-XDOM-C700 | Corner | Project delete + LDAP deactivate assignee + agent tool mid-create + M2 outbox drain | agent abort; outbox cancel/anonymize; cascade tombstones; search empty; audit chain | Manual | P0 | GZ-ARCH-09, MCP+SSO+MS | Future |
| ECC-XDOM-C701 | Corner | Project delete while calendar RRULE expand + LiveKit provision + offline RSVP queued | accept 404/410; room reconciled; queue item failed UI; at-most-one cancel notify | API/E2E | P0 | CAL+CALL+UX | Future |
| ECC-XDOM-C702 | Corner | Wiki restrict + bases share revoke + Ctrl+K aggregator + enforce-SSO flip same query | results only permitted; Magic path blocked; no title leak; sessions policy per ADR | API/E2E | P0 | NF-03, WIKI+SSO+V4 | Future |
| ECC-XDOM-C703 | Corner | Soft-delete 30d purge + legal-hold ACC invoice + agent confirm + restore click | purge skips hold; confirm reject; restore policy documented; no resurrect purged non-hold | API | P0 | NFR-11, legal-hold | Future |
| ECC-XDOM-C704 | Corner | GDPR user hard-delete + CRM merge mid-batch + wiki authorship + calendar attendee + outbox mention | anonymize all; merge skip; legal-hold intact; search clean | API | P0 | GDPR, GZ-SEC-17/18 | Future |
| ECC-XDOM-C705 | Corner | FA-06 rotate JWT+LiveKit+Resend during M4 call + founder break-glass + IdP outage | old tokens dead; break-glass may mint; call rejoin policy; mail uses new creds; runbook | Manual | P0 | FA-06, SSO+M4 | Future |
| ECC-XDOM-C706 | Corner | Marketplace install (base+wiki+CRM fields) + project membership revoke mid-install + search index | install rollback; no partial ACL; search no orphan titles | API | P0 | marketplace+NF-03 | Future |
| ECC-XDOM-C707 | Corner | Leave busy + CRM reminder + AI hard-stop alert + notification mute prefs | leave busy correct; hard-stop forced; CRM reminder per prefs; no alert storm | API | P1 | HR+CRM+AI+NOTIF | Future |
| ECC-XDOM-C708 | Corner | Optimistic offline wiki publish + restrict allow-list remove + Meili lag + flush | flush 403; index eventually 0 for user; no claimed saved; draft recoverable | E2E | P0 | UX+WIKI+NF-03 | Future |
| ECC-XDOM-C709 | Corner | ACC period close + 1C replay + MCP invoice + M3 AI route-flip mid confirm | closed immutable; confirm fail-closed; single writer charge; audit complete | API | P0 | ACC+MCP+AI+onec | Future |

**Подсчёт §8:** 10

---

## Счётчик

| Домен | Prefix | Corner rows | Gap |
|-------|--------|------------:|-----|
| v4 / Bases partition×HR/CRM/ACC | `ECC-V4-C700+` | **13** | Future |
| Wiki restrict×export×search | `ECC-WIKI-C700+` | **9** | Future |
| SSO IdP×break-glass×JIT | `ECC-SSO-C700+` | **10** | Future |
| MCP multi-agent×MS×audit | `ECC-MCP-C700+` | **9** | Future |
| AI BYOK/budget/canary | `ECC-AI-C700+` | **7** | Future |
| MS canary×outbox×schema | `ECC-MS-C700+` | **10** | Future |
| Redesign PWA×offline×a11y | `ECC-UX-C700+` | **9** | Future |
| Cross-domain cascade×SSO×MCP | `ECC-XDOM-C700+` | **10** | Future |
| **Итого Corner** | | **77** | только Corner; Edge/Happy/Neg не включались |

**Минимум 65:** выполнен (77 ≥ 65).  
**C200–C600:** не копировались — wave5 углубление (partition, IdP outage, dual-canary, PWA×ACL, legal-hold×purge×agent, marketplace×cascade).  
**Основной каталог** `.planning/archive/july-2026/EDGE-CORNER-CATALOG.md`: не изменялся.  
**Коммит:** не создавался.

### Фокус-покрытие (чеклист)

| Тема | Примеры ID |
|------|------------|
| Bases partition / migrate / marketplace | V4-C700–C704, C702–C703 |
| Leave×RRULE×accrual / CRM merge×GDPR / ACC×1C | V4-C707–C712 |
| Wiki export×restrict / nested move / Meili | WIKI-C700–C705, C707 |
| SSO IdP outage × break-glass × group storm | SSO-C700–C702, C706–C707 |
| MCP multi-agent × route-flip × legal-hold | MCP-C700–C702, C707–C708 |
| AI triple-stop / canary dual-charge / DLP | AI-C700–C704 |
| MS dual-canary / Redis split-brain / schema v3 | MS-C700–C704, C709 |
| PWA×offline×multi-tab 409 / a11y | UX-C700–C705, C708 |
| XDOM cascade×SSO×MCP×MS×UX | XDOM-C700–C709 |

---

*qa-manual 2026-07-10 · wave5 · merge candidate after `qa-manual-ver` · QH-P2 TC authoring*
