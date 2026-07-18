# Hub — Corner Cases Expansion (parent, 2026-07-10)

> **Тип:** только **Corner** (редкие комбинации, races, mid-flight, multi-step failures, cross-domain).  
> **Scope:** as-is + будущий (CAL-SYNC, Wiki, SSO, MCP/agents, microservices, Redesign, mail L0–L3).  
> **Merge target:** `.planning/july-2026/EDGE-CORNER-CATALOG.md` § «Corner expansion» + доменные секции.  
> **ID ranges:** `ECC-*-C200+` / `ECC-FUT-*` чтобы не пересечь базовый каталог и edge-expansions `*100+`.

Формат: `| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |`

---

## A. Auth / Sessions / Org (as-is)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-AUTH-C200 | Corner | Verify OK + сразу deactivate admin до первого API | последующие API 401/403; refresh dead | API | P0 | TC-ADMIN | TODO |
| ECC-AUTH-C201 | Corner | Two devices: logout-all vs revoke-one race | consistent session set | API | P0 | sessions | TODO |
| ECC-AUTH-C202 | Corner | Magic link send, Redis flush mid-flight, verify | 400; no orphan JWT | API | P0 | Redis | TODO |
| ECC-AUTH-C203 | Corner | Invite accept + pending purge job same second | one wins; no half-user | API | P0 | 7d purge | TODO |
| ECC-AUTH-C204 | Corner | Refresh rotation A, stolen old refresh B | B rejected (reuse detection) | API | P0 | NFR-10 | TODO |
| ECC-AUTH-C205 | Corner | Clock skew + rate-limit window boundary | no bypass of 3/10m | API | P0 | NFR-05 | TODO |
| ECC-AUTH-C206 | Corner | Onboarding incomplete + project invite deep-link | redirect onboarding then land | E2E | P1 | US-M08 | TODO |
| ECC-AUTH-C207 | Corner | Founder soft-deleted then restore | founder uniqueness intact | API | P0 | TZ founder | TODO |
| ECC-AUTH-C208 | Corner | GDPR hard-delete user mid open WS + refresh | WS kill + refresh 401 + data purge | Manual | P0 | US-S03 | TODO |
| ECC-AUTH-C209 | Corner | Email provider timeout after code stored | user can still verify; no double-send storm | API | P1 | hub-mail | TODO |

## B. RBAC / IDOR (as-is)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-RBAC-C200 | Corner | Role demote while request in-flight | fail-closed on write | API | P0 | rbac cache | TODO |
| ECC-RBAC-C201 | Corner | Custom role bitmask flip during Kanban drag | 403 mid-mutation; UI rollback | E2E | P1 | — | TODO |
| ECC-RBAC-C202 | Corner | Project archived + member still has JWT | writes 403; reads policy | API | P0 | — | TODO |
| ECC-RBAC-C203 | Corner | Transfer ownership + old owner API parallel | only new owner privileged | API | P0 | — | TODO |
| ECC-RBAC-C204 | Corner | Cross-tenant fantasy: forge other org UUID | always 404 (single-org) | API | P0 | TZ multi-org | TODO |
| ECC-RBAC-C205 | Corner | Invite to project while user deactivated | reject | API | P0 | — | TODO |
| ECC-RBAC-C206 | Corner | Remove from project while editing issue | save 403; no silent success | API/E2E | P0 | IDOR | TODO |

## C. Tracker / Sprint / Optimistic lock (as-is)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-TRCK-C200 | Corner | Tab A/B edit same issue; A saves; B stale version | B 409; retry with merge UI | E2E | P0 | NFR-24 | TODO |
| ECC-TRCK-C201 | Corner | Workflow transition + assignee change race | consistent final state or 409 | API | P0 | — | TODO |
| ECC-TRCK-C202 | Corner | Delete parent while child link create | no orphan link / 404 | API | P1 | — | TODO |
| ECC-TRCK-C203 | Corner | Soft-delete issue then key reuse attempt | keys never reuse | API | P0 | — | TODO |
| ECC-TRCK-C204 | Corner | Move to column + WIP limit hit mid-drag | reject; card returns | E2E | P1 | WIP | TODO |
| ECC-TRCK-C205 | Corner | JQL saved filter + project membership revoked | filter empty/403 no leak | API | P0 | NF-03 | TODO |
| ECC-TRCK-C206 | Corner | Bulk edit 50 issues one fails version | partial policy documented | API | P1 | — | TODO |
| ECC-SPR-C200 | Corner | Complete sprint while issue moved out concurrently | no lost issues | API | P0 | TC-SPRINT | TODO |
| ECC-SPR-C201 | Corner | Two completes same sprint | second 400; idempotent | API | P0 | — | TODO |
| ECC-SPR-C202 | Corner | Start sprint + delete sprint race | one terminal state | API | P1 | — | TODO |

## D. Messenger / WS / Search / Notification (as-is)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MSG-C200 | Corner | Send + remove from channel before NATS fanout | no delivery to removed; audit | API | P0 | — | TODO |
| ECC-MSG-C201 | Corner | Edit message while soft-delete | 404/gone | API | P1 | — | TODO |
| ECC-MSG-C202 | Corner | Double client retry same idempotency key | one message | API | P0 | — | TODO |
| ECC-MSG-C203 | Corner | Pin message then delete message | pin cleared | API | P1 | — | TODO |
| ECC-MSG-C204 | Corner | DM user blocks then message | 403; no leak | API | P0 | blocks | TODO |
| ECC-WS-C200 | Corner | Subscribe room without membership (NF-02) | reject | API/WS | P0 | NF-02 | TODO |
| ECC-WS-C201 | Corner | Membership revoked mid-subscribed | force unsubscribe | WS | P0 | NF-02 | TODO |
| ECC-WS-C202 | Corner | Reconnect storm 100 tabs same user | backpressure; no crash | Manual | P0 | NFR | TODO |
| ECC-WS-C203 | Corner | NATS partition: publish OK consume lag | eventual; no silent drop forever | Manual | P0 | NFR-26 | TODO |
| ECC-SRCH-C200 | Corner | Search without user-scope filter (NF-03) | only visible hits | API | P0 | NF-03 | Partial |
| ECC-SRCH-C201 | Corner | Index lag: create issue then immediate search | miss OK or wait; no foreign | API | P1 | Meili | TODO |
| ECC-NOTIF-C200 | Corner | Preference off + forced security mail | security still delivered | API | P0 | — | TODO |
| ECC-NOTIF-C201 | Corner | Telegram unbound mid-delivery | fallback in-app; no crash | API | P1 | — | TODO |
| ECC-NOTIF-C202 | Corner | WebPush endpoint SSRF (NF-07) | ValidateOutboundURL reject | API | P0 | NF-07 | TODO |

## E. Calls + LiveKit (as-is; SFU stay)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-CALL-C200 | Corner | LiveKit restart mid-call | reconnect; call not false-ended | Manual | P0 | LiveKit | Manual |
| ECC-CALL-C201 | Corner | Webhook replay old participant_joined | idempotent counts | API | P0 | SEC-02 | Partial |
| ECC-CALL-C202 | Corner | End call + recording stop race | no orphan egress | API | P0 | recording | TODO |
| ECC-CALL-C203 | Corner | Guest join after call ended | 410/404 | API | P1 | invite | TODO |
| ECC-CALL-C204 | Corner | Token expiry mid-screen-share | rejoin; share stops clean | Manual | P1 | — | Manual |
| ECC-CALL-C205 | Corner | Scheduled due worker + manual start same time | one room | API | P0 | scheduled | TODO |
| ECC-CALL-C206 | Corner | Conflict check pass then concurrent book | second warned/409 | API | P1 | FindConflicts | TODO |
| ECC-CALL-C207 | Corner | Raise-hand metadata corrupt JSON | safe fallback UI | FE | P2 | CallRoom | Partial |
| ECC-CALL-C208 | Corner | Channel deleted while scheduled active | deactivate / 404 | API | P1 | — | TODO |
| ECC-CALL-C209 | Corner | Recording webhook wrong iss after key rotate | 401; no state flip | API | P0 | FA-06 | TODO |

## F. Calendar as-is + CAL-SYNC future (SEED-002)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-CAL-C200 | Corner | Aggregator: soft-deleted call still listed | hidden | API | P1 | soft-delete | TODO |
| ECC-CAL-C201 | Corner | RRULE infinite + range 366d | capped instances; no OOM | Unit | P0 | expandCall | Partial |
| ECC-CAL-C202 | Corner | Accept → decline → accept race version | last writer / 409 | API | P0 | C10 | Future |
| ECC-CAL-C203 | Corner | Organizer cancel while attendee accepting | attendee 410; inbox clear | API | P0 | C1/C24 | Future |
| ECC-CAL-C204 | Corner | Reschedule + ICS L0 already sent | cancel+new ICS or UPDATE method | Manual | P0 | L0 | Future |
| ECC-CAL-C205 | Corner | Recurring exception (future) vs series accept | v1 series-only; exception backlog | API | P1 | C2 | Future |
| ECC-CAL-C206 | Corner | Channel members change after invite created | policy: snapshot vs live | API | P1 | C25 | Future |
| ECC-CAL-C207 | Corner | Tentative + leave approved same slot | availability shows leave priority | API | P1 | C12 | Future |
| ECC-CAL-C208 | Corner | External iCal event conflicts Hub RSVP | both shown; Hub wins busy? | API | P1 | L1 | Future |
| ECC-CAL-C209 | Corner | L2 IMAP ingest duplicate UID | idempotent import | API | P1 | L2 ADR | Future |
| ECC-CAL-C210 | Corner | L3 OAuth token revoke mid-sync | stop; no partial corrupt | API | P0 | L3 ADR | Future |
| ECC-CAL-C211 | Corner | Messenger create call event + calendar create same slot | conflict warn both | API | P1 | SEED-002 | Future |
| ECC-CAL-C212 | Corner | Attendee GDPR delete after accept | anonymize attendee row; event stays | API | P0 | GDPR | Future |
| ECC-CAL-C213 | Corner | DST transition across recurring standup | correct local wall times | E2E | P0 | NFR-12 | Future |
| ECC-CAL-C214 | Corner | needs_action ignored 30d | auto-expire / reminder policy | API | P2 | product | Future |
| ECC-CAL-C215 | Corner | Partial outage hub-mail: in-app OK email fail | event created; retry mail | API | P1 | L0 | Future |

## G. TMS / Storage / Dashboard (as-is)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-TMS-C200 | Corner | Allure import + case deleted mid-match | unmatched bucket; no crash | API | P1 | Allure | TODO |
| ECC-TMS-C201 | Corner | Execute result + complete run race | consistent stats | API | P0 | — | TODO |
| ECC-TMS-C202 | Corner | Rollback version while edit open | 409 or reload | API | P1 | versions | TODO |
| ECC-STOR-C200 | Corner | Presigned PUT >25MB (NF-09) | S3 policy reject | Manual | P0 | NF-09 | TODO |
| ECC-STOR-C201 | Corner | Upload then soft-delete file mid-download | 404 after TTL | API | P1 | — | TODO |
| ECC-STOR-C202 | Corner | Virus-like zip bomb (if scanned) | reject / limit | Manual | P1 | — | TODO |
| ECC-DASH-C200 | Corner | Widget query + project revoke | empty/403 no leak | API | P0 | NF-03 | TODO |
| ECC-DASH-C201 | Corner | Layout save concurrent two tabs | last-write or version | API | P1 | — | TODO |

## H. v4 as-is (intranet/crm/hr/accounting/bases)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-V4-C200 | Corner | Bases: share design change mid-edit (owner→project) | no data leak during migrate | API | P0 | BS-1 | Future |
| ECC-V4-C201 | Corner | Bases 50k rows + filter + realtime BS-5 | p95 OK or degrade gracefully | Manual | P1 | BS-5 | Future |
| ECC-V4-C202 | Corner | Intranet poll close while voting | reject late votes | API | P1 | US-INTR | Future |
| ECC-V4-C203 | Corner | Random coffee match + user deactivate | rematch / skip | API | P1 | coffee | Future |
| ECC-V4-C204 | Corner | HR leave approve + calendar busy race | busy after approve only | API | P0 | US-CAL-04 | Partial |
| ECC-V4-C205 | Corner | CRM import CSV partial fail row 500 | report; no half-corrupt | API | P1 | CRM | Future |
| ECC-V4-C206 | Corner | Accounting period close + late invoice | reject or reopen policy | API | P1 | ACC | Future |
| ECC-V4-C207 | Corner | Kudos leaderboard + soft-deleted giver | anonymize counts | API | P2 | kudos | Future |
| ECC-V4-C208 | Corner | Onboarding template assign + user already complete | no-op / skip | API | P2 | US-INTR | Future |
| ECC-V4-C209 | Corner | 1C webhook replay | idempotent | API | P0 | onec | Partial |

## I. AI as-is + MCP/agents future

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-AI-C200 | Corner | BYOK key rotate mid-stream | fail soft; no leak old key | API | P0 | BYOK | TODO |
| ECC-AI-C201 | Corner | Daily token hard-stop mid-completion | stop; alert; no overage | API | P0 | GZ-SEC-22 | TODO |
| ECC-AI-C202 | Corner | Model fallback cascade all down | 503 clear error | API | P1 | LiteLLM | Future |
| ECC-MCP-C200 | Corner | Agent delete on behalf without confirm | blocked (ступенчатая автономия) | API | P0 | GZ-SEC-09 | Future |
| ECC-MCP-C201 | Corner | PAT scope read + tool tries write | deny | API | P0 | GZ-SEC-23 | Future |
| ECC-MCP-C202 | Corner | Agent-principal ∩ user rights shrink mid-task | re-check; abort | API | P0 | agent-principal | Future |
| ECC-MCP-C203 | Corner | Audit on-behalf missing agent id | reject | API | P0 | audit | Future |
| ECC-MCP-C204 | Corner | Curated toolset removed while session open | tool 404 | API | P1 | MCP | Future |

## J. Wiki future (US-WIKI)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-WIKI-C200 | Corner | Restrict page + space admin demoted | still restricted; no leak | API | P0 | GZ-SEC-07 | Future |
| ECC-WIKI-C201 | Corner | Concurrent edit same page versions | conflict / merge | API | P0 | versions | Future |
| ECC-WIKI-C202 | Corner | Delete space with pages + issue-links | cascade policy; no orphan ACL | API | P0 | US-WIKI | Future |
| ECC-WIKI-C203 | Corner | Search wiki without user-scope | NF-03 | API | P0 | NF-03 | Future |
| ECC-WIKI-C204 | Corner | Move page across spaces RBAC change | re-authz | API | P1 | — | Future |
| ECC-WIKI-C205 | Corner | Handbook HR link + wiki space deleted | broken-link UX | E2E | P1 | HR | Future |
| ECC-WIKI-C206 | Corner | Restore soft-deleted page mid-edit new | version fork policy | API | P1 | soft-delete | Future |

## K. SSO future (ADR-AUTH-SSO)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-SSO-C200 | Corner | Enforce-SSO + founder break-glass Magic Link | founder OK; others blocked | API | P0 | GZ-SEC-04 | Future |
| ECC-SSO-C201 | Corner | JIT create + LDAP later deactivates | Hub user deactivated | API | P0 | LDAP-sync | Future |
| ECC-SSO-C202 | Corner | Org toggle sso→both mid-session | existing JWT until expiry; policy clear | API | P1 | org-toggle | Future |
| ECC-SSO-C203 | Corner | OIDC nonce/state replay | reject | API | P0 | OIDC | Future |
| ECC-SSO-C204 | Corner | LDAP bind timeout + cached user | fail-closed or stale policy documented | API | P0 | LDAP | Future |
| ECC-SSO-C205 | Corner | Map only system-role; project roles unchanged | no privilege escalation | API | P0 | GZ | Future |
| ECC-SSO-C206 | Corner | Invite non-SSO user while enforce-SSO | reject or pending exception | API | P1 | invite | Future |

## L. Microservices / M2–M4 future

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MS-C200 | Corner | Caddy route-flip mid-request notification | retry/idempotent; no dual-write corrupt | Manual | P0 | M2 | Future |
| ECC-MS-C201 | Corner | NATS schema v1 consumer + v2 producer | reject/compat window | API | P0 | contracts | Future |
| ECC-MS-C202 | Corner | Outbox publish success DB commit fail | at-least-once; dedup consumer | API | P0 | outbox | Future |
| ECC-MS-C203 | Corner | hub-ws membership Redis stale vs core | revalidate; NF-02 | WS | P0 | M4 | Future |
| ECC-MS-C204 | Corner | hub-calls LiveKit OK but core call row missing | reconcile job | Manual | P0 | M4 LiveKit | Future |
| ECC-MS-C205 | Corner | OTel trace broken across NATS | gate fail before pilot | Manual | P0 | M0-06 | Future |
| ECC-MS-C206 | Corner | Split compose profile vs monolith dual-run | feature flag one writer | Manual | P1 | GZ-ARCH-12 | Future |
| ECC-MS-C207 | Corner | user.deleted orphan UUID in messenger Mongo | idempotent consumer + reconcile | API | P0 | GZ-ARCH-09 | Future |

## M. Redesign / UX future (Cloud Code)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-UX-C200 | Corner | Keyboard-only RSVP accept in new inbox | focus order WCAG | E2E | P1 | RD | Future |
| ECC-UX-C201 | Corner | Dark/light token mismatch mid-theme toggle | no unreadable contrast | E2E | P1 | a11y | Future |
| ECC-UX-C202 | Corner | Reduced motion + call reconnect UI | no essential info only in animation | Manual | P1 | a11y | Future |
| ECC-UX-C203 | Corner | Optimistic UI accept + server 409 | rollback + toast | E2E | P0 | CAL | Future |
| ECC-UX-C204 | Corner | Offline create event then reconnect | queue or clear error | E2E | P1 | — | Future |

## N. Cross-domain mega-corners (P0)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-XDOM-C200 | Corner | Deactivate user: WS+refresh+search+calendar+calls | all fail-closed | Manual | P0 | security | TODO |
| ECC-XDOM-C201 | Corner | Project delete: issues+TMS+channels+calls+files | cascade/reconcile; no IDOR leftovers | Manual | P0 | — | TODO |
| ECC-XDOM-C202 | Corner | Soft-delete 30d purge job during restore click | one wins; no resurrect after purge | API | P0 | NFR-11 | TODO |
| ECC-XDOM-C203 | Corner | FA-06 rotate LiveKit+JWT+Resend same window | old tokens die; new OK; docs runbook | Manual | P0 | FA-06 | TODO |
| ECC-XDOM-C204 | Corner | Calendar accept creates call; LiveKit down | event OK; call provision retry | API | P0 | CAL+CALL | Future |
| ECC-XDOM-C205 | Corner | Agent creates issue + user loses permission | agent stop; audit | API | P0 | MCP | Future |
| ECC-XDOM-C206 | Corner | SSO enforce + Magic Link invite email in flight | invite invalid under enforce | API | P1 | SSO | Future |
| ECC-XDOM-C207 | Corner | Search indexes wiki page then restrict flag | disappear from others' search | API | P0 | WIKI+NF-03 | Future |
| ECC-XDOM-C208 | Corner | Notification from calendar invite + prefs mute calendar | still show in-app inbox | API | P1 | CAL+NOTIF | Future |
| ECC-XDOM-C209 | Corner | Backup restore mid-write wave | consistency checklist | Manual | P0 | backup | TODO |

---

## Счётчики (этот файл)

| Блок | Corner |
|------|-------:|
| A Auth | 10 |
| B RBAC | 7 |
| C Tracker/Sprint | 10 |
| D Msg/WS/Search/Notif | 14 |
| E Calls | 10 |
| F Calendar/CAL-SYNC | 16 |
| G TMS/Stor/Dash | 8 |
| H v4 | 10 |
| I AI/MCP | 8 |
| J Wiki | 7 |
| K SSO | 7 |
| L Microservices | 8 |
| M Redesign UX | 5 |
| N Cross-domain | 10 |
| **ИТОГО Corner** | **130** |

Все Gap = TODO / Manual / Future / Partial — для QH-P2 TC authoring.

---

*Parent agent 2026-07-10 · merge after edge-expansions 01–05 land*
