# Corner expansion C03 — Messenger / WS / Notification / Search

> **Тип:** только **Corner** (редкие комбинации, races, mid-flight, multi-step, cross-domain).  
> **ID ranges:** `ECC-MSG-C300+` · `ECC-WS-C300+` · `ECC-NOTIF-C300+` · `ECC-SRCH-C300+`  
> **Не дублирует** parent `ECC-*-C200…C204` (msg/ws/notif/search) дословно — углубляет комбинациями.  
> **Фокус:** NF-02 mid-subscribe · NF-03 leak · NATS down mid-send · idempotent retry · block+DM · pin+delete · reconnect storm · mute vs security · WebPush SSRF mid-update · index lag · unread race · reaction storm · channel delete mid-WS.  
> **Merge target:** `.planning/archive/july-2026/EDGE-CORNER-CATALOG.md` § Corner expansion (после ревью).  
> **Владелец:** `qa-manual` · 2026-07-10 · **не коммитить в каталог** без ревью.

Формат: `| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |`

---

## 1. Messenger — `ECC-MSG-C300+`

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MSG-C300 | Corner | Send OK → NATS JetStream down **после** Mongo write, **до** publish; client retry с тем же idempotency-key | одно сообщение в Mongo; outbox/retry или 503 с clear error; нет дубля при retry | API | P0 | NFR-26, ECC-MSG-C202↑ | TODO |
| ECC-MSG-C301 | Corner | Два клиента: разные idempotency-key, одинаковый content+channel в одном окне 100ms | два сообщения (ключи разные); UI не схлопывает чужой key | API | P1 | idempotent | TODO |
| ECC-MSG-C302 | Corner | Retry с тем же key **после** soft-delete исходного сообщения | 404/gone или новый msg по политике; нет «воскрешения» deleted как live | API | P0 | soft-delete, C201↑ | TODO |
| ECC-MSG-C303 | Corner | Idempotency-key reuse **в другом** channel_id | reject / isolated per-channel; нет cross-channel replay | API | P0 | idempotent | TODO |
| ECC-MSG-C304 | Corner | A блокирует B → B открывает/создаёт DM → A шлёт в существующий DM | 403 на send от A→B и B→A (или policy one-way); нет content leak в response | API | P0 | blocks, C204↑ | TODO |
| ECC-MSG-C305 | Corner | Mutual block A↔B + simultaneous unblock + send race | после unblock только один send path wins; нет half-open DM | API | P0 | blocks | TODO |
| ECC-MSG-C306 | Corner | Block mid-typing: B typing indicator в DM, A ставит block | typing stop; дальнейшие events B не доставляются A | WS/API | P1 | blocks, presence | TODO |
| ECC-MSG-C307 | Corner | Pin message M → soft-delete M **параллельно** с pin list fetch | pin list без M; нет dangling pin UUID; WS pin_cleared | API/WS | P0 | pin, C203↑ | TODO |
| ECC-MSG-C308 | Corner | Pin M → edit M content → delete M → restore M (корзина) | после restore pin **не** авто-восстанавливается (или явная политика) | API | P1 | NFR-11, pin | TODO |
| ECC-MSG-C309 | Corner | Два admin: pin разных сообщений до лимита pins + concurrent pin сверх лимита | один 200, второй 400/409; лимит не превышен | API | P1 | pin limit | TODO |
| ECC-MSG-C310 | Corner | Reaction storm: 50 клиентов toggle одного emoji на одном msg за 1s | consistent final count; нет negative/overflow; rate-limit или coalesce | API | P0 | TC-MSG-04 | TODO |
| ECC-MSG-C311 | Corner | Reaction add + soft-delete message в том же ms | reaction 404/gone; нет orphan reaction docs | API | P0 | ECC-MSG-014↑ | TODO |
| ECC-MSG-C312 | Corner | Reaction storm на thread parent + concurrent delete parent | replies/reactions consistent с cascade policy; нет panic | API | P1 | thread, C021↑ | TODO |
| ECC-MSG-C313 | Corner | Unread: A sends → B MarkRead(channel) **пока** C sends второе msg | B unread ≥1 (второе) или atomic snapshot; нет stuck unread=0 при новом msg | API | P0 | TC-MSG-09 | TODO |
| ECC-MSG-C314 | Corner | Unread race: два устройства B1 MarkRead, B2 получает WS new_message out-of-order | badge сходится к серверному GetUnread; нет вечного «1» | E2E | P0 | unread, WS | TODO |
| ECC-MSG-C315 | Corner | Increment unread + remove member mid-increment batch | removed user не получает increment; sender skip сохранён | API | P0 | unread, membership | TODO |
| ECC-MSG-C316 | Corner | Channel soft-delete mid-send (handler уже прошёл membership check) | 404/410; сообщение не публикуется в NATS / не видно в history | API | P0 | channel delete | TODO |
| ECC-MSG-C317 | Corner | Forward private msg → public channel **пока** source channel membership revoked | 403; нет partial forward body в public | API | P0 | ECC-MSG-033↑, NF-03 | TODO |
| ECC-MSG-C318 | Corner | Mute channel until T + @mention + security-class notif на тот же channel | mute глушит chat notif; security/forced всё равно доставляется (см. NOTIF) | API | P0 | mute vs security | TODO |
| ECC-MSG-C319 | Corner | Read-only flip true mid-compose: client POST после flip | 403; optimistic UI rollback; нет ghost message | API/E2E | P1 | TC-MSG-07 | TODO |
| ECC-MSG-C320 | Corner | Thread reply + parent soft-delete + reply edit race | reply edit 404/gone или orphan policy documented; reply_count consistent | API | P1 | TC-MSG-06 | TODO |
| ECC-MSG-C321 | Corner | NATS down mid-send **после** outbox enqueue; JetStream recovers | consumer догоняет; ровно один WS fanout; audit без дублей | Manual | P0 | NFR-26, C203↑WS | TODO |
| ECC-MSG-C322 | Corner | DM create concurrent A→B и B→A (оба POST /dm) | один channel (dedup); оба 200 с одним id | API | P0 | TC-MSG-02 | TODO |
| ECC-MSG-C323 | Corner | Attachment presign linked in msg + file soft-delete mid-download by peer | msg остаётся; file 404 after TTL; нет SSRF через preview URL | API | P1 | storage, NF-09 | TODO |
| ECC-MSG-C324 | Corner | Rate 60/min: 60-е OK, 61-е 429 **во время** NATS lag (slow ack) | 429 на API layer; нет обхода через retry storm | API | P0 | NFR-07 | TODO |

---

## 2. WebSocket — `ECC-WS-C300+`

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-WS-C300 | Corner | Subscribe `channel:X` OK → kick/remove membership → **re-subscribe same room** без reconnect | reject (NF-02); stale subscription не оживает | WS | P0 | NF-02, C201↑ | TODO |
| ECC-WS-C301 | Corner | Subscribe in-flight: CanSubscribe true на check, membership revoke **до** room join commit | fail-closed; клиент не в room fanout | WS | P0 | NF-02 | TODO |
| ECC-WS-C302 | Corner | Multi-room: subscribed A+B; revoke only A; message in B | A force-unsub; B продолжает; нет cross-leak A events | WS | P0 | NF-02 | TODO |
| ECC-WS-C303 | Corner | Subscribe `project:P` + demote RBAC mid-subscribed (bit flip) | force unsub или re-authz deny; writes via WS reject | WS | P0 | NF-02, rbac cache | TODO |
| ECC-WS-C304 | Corner | Subscribe `call:C` after call ended / channel deleted | 410/404; no zombie room | WS | P1 | calls, channel delete | TODO |
| ECC-WS-C305 | Corner | Channel delete mid-WS: members subscribed → DELETE channel | all force-unsub; subsequent events dropped; REST 404 | WS/API | P0 | channel delete mid-WS | TODO |
| ECC-WS-C306 | Corner | Channel delete + concurrent subscribe from late joiner | subscribe reject; delete completes; no half-room | WS | P0 | NF-02 | TODO |
| ECC-WS-C307 | Corner | Reconnect storm: 100 tabs × rapid close/open + JWT refresh single-flight | backpressure/limit; server alive; refresh mutex не deadlocks | Manual | P0 | C202↑, FE-H1 | TODO |
| ECC-WS-C308 | Corner | Reconnect storm **во время** membership revoke на одном channel | после reconnect CanSubscribe fail; нет rejoin bypass | WS | P0 | NF-02, storm | TODO |
| ECC-WS-C309 | Corner | Reconnect with expired access + valid refresh mid-storm | one refresh; WS re-auth; no 100 parallel refresh | E2E | P0 | NFR-09/10 | TODO |
| ECC-WS-C310 | Corner | Origin mismatch mid-session (CheckOrigin only on upgrade) + forged subscribe frame | upgrade already done: subscribe всё равно authz; Origin spoof не даёт room | WS | P0 | SEC-H1 | TODO |
| ECC-WS-C311 | Corner | NATS partition: WS connected, publish path OK, consumer lag → burst catch-up | ordered/eventual; no silent forever-drop; client may dedupe by msg id | Manual | P0 | NFR-26, C203↑ | TODO |
| ECC-WS-C312 | Corner | NATS down mid-send: REST 201/503 + WS subscribers online | subscribers не видят phantom; после recover — один event | API/WS | P0 | MSG-C300 | TODO |
| ECC-WS-C313 | Corner | User soft-delete / GDPR mid open WS (multi-room) | all rooms kill; ping fails; refresh 401 | Manual | P0 | US-S03, AUTH-C208↑ | TODO |
| ECC-WS-C314 | Corner | Subscribe `user:otherUUID` (foreign personal room) | reject NF-02; только свой `user:` | WS | P0 | NF-02 | TODO |
| ECC-WS-C315 | Corner | Admin room `admin:` subscribe non-admin JWT | reject; no admin event leak | WS | P0 | NF-02, RBAC | TODO |
| ECC-WS-C316 | Corner | Ping flood during reconnect storm | throttle/close slow; no goroutine leak | Manual | P1 | NFR, TC-WS-03 | TODO |
| ECC-WS-C317 | Corner | WS binary/oversized frame mid-subscribed channel | close 1009/policy; room cleaned; no crash | Manual | P1 | Fiber WS | TODO |
| ECC-WS-C318 | Corner | Membership add mid-session: user not subscribed → auto-join? | **нет** auto-subscribe; только явный subscribe + CanSubscribe OK | WS | P1 | NF-02 | TODO |
| ECC-WS-C319 | Corner | Double subscribe same room same conn | idempotent; one membership in hub; no double fanout | WS | P1 | — | TODO |
| ECC-WS-C320 | Corner | Unsubscribe + immediate message publish race | unsubscribed client не получает; no panic on send-to-closed | WS | P0 | — | TODO |

---

## 3. Notification — `ECC-NOTIF-C300+`

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-NOTIF-C300 | Corner | Preference `channel_mute` / in-app off + **security** event (passwordless session revoke / new device) | security всё равно delivered (mail/in-app); mute не глотает | API | P0 | C200↑, mute vs security | TODO |
| ECC-NOTIF-C301 | Corner | Quiet hours + security + non-urgent mention same second | security now; mention deferred; order в inbox корректный | API | P0 | TC-NOTIF-05/06 | TODO |
| ECC-NOTIF-C302 | Corner | Global notif off + channel unmute exception + @mention | policy: global off wins **или** channel exception documented; security всё равно | API | P1 | preferences | TODO |
| ECC-NOTIF-C303 | Corner | WebPush subscribe OK → PATCH endpoint на `http://169.254.169.254/` mid-update | ValidateOutboundURL reject (NF-07); old endpoint остаётся или sub disabled; no SSRF fetch | API | P0 | NF-07, C202↑ | TODO |
| ECC-NOTIF-C304 | Corner | WebPush endpoint DNS-rebinding: FQDN resolves public → then private on delivery | reject on subscribe **и** re-validate on send; fail-closed | API | P0 | NF-07, NFR-21 | TODO |
| ECC-NOTIF-C305 | Corner | WebPush update keys+endpoint concurrent two tabs | one wins / version; no dual active SSRF-prone endpoints | API | P1 | push | TODO |
| ECC-NOTIF-C306 | Corner | Telegram unbound **после** enqueue, **до** Bot API call | fallback in-app; job not crash-loop; audit «undelivered tg» | API | P1 | C201↑ | TODO |
| ECC-NOTIF-C307 | Corner | Telegram bound + preference mute channel + urgent call invite | urgent/call policy breaks mute; tg или in-app per matrix | API | P1 | mute vs security/urgent | TODO |
| ECC-NOTIF-C308 | Corner | Mark-read notif + new notif same entity race (assign then reassign) | both rows consistent; badge = unread count server | API | P1 | TC-NOTIF-03 | TODO |
| ECC-NOTIF-C309 | Corner | Read-all + flood of NATS notif events mid-handler | after read-all, only **newer** unread; no lost-update to all-read falsely | API | P0 | TC-NOTIF-04 | TODO |
| ECC-NOTIF-C310 | Corner | Debounce window: 10 mentions same issue < debounce TTL | coalesced notif; no 10 identical pushes | API | P1 | TestShouldDispatch_Debounce | TODO |
| ECC-NOTIF-C311 | Corner | Push delivery 410 Gone from push service mid-batch | subscription pruned; other users unaffected; no retry storm | API | P1 | WebPush | TODO |
| ECC-NOTIF-C312 | Corner | Preference flip off→on while security mail already sending | in-flight security completes; subsequent respect new prefs except forced types | API | P1 | mute vs security | TODO |
| ECC-NOTIF-C313 | Corner | Soft-deleted user still in NATS consumer group for notif | skip / drop; no email to deleted; no panic | API | P0 | GDPR, soft-delete | TODO |
| ECC-NOTIF-C314 | Corner | WebPush endpoint `https://evil.test` allowed by DNS but redirects to metadata IP | SafeHTTP + ValidateOutboundURL; no follow to private | API | P0 | NF-07 | TODO |
| ECC-NOTIF-C315 | Corner | Channel mute until T −1s / +1s boundary + message event | до T muted; после T delivered; clock UTC | API | P1 | TC-MSG-08, NFR-12 | TODO |

---

## 4. Search — `ECC-SRCH-C300+`

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-SRCH-C300 | Corner | Index lag: create private-channel msg → immediate global search as non-member | **no hit** even if Meili already indexed (NF-03 filter); miss OK for member until indexed | API | P0 | NF-03, C200/C201↑ | TODO |
| ECC-SRCH-C301 | Corner | Index lag: member creates msg → search → then kick → search again | first may miss/hit; second **must not** return msg (scope re-check) | API | P0 | NF-03 | TODO |
| ECC-SRCH-C302 | Corner | Soft-delete message still in Meili stale index | search excludes deleted; snippet no leak | API | P0 | NF-03, soft-delete | TODO |
| ECC-SRCH-C303 | Corner | Soft-delete channel mid-reindex job | no orphan channel hits; job idempotent | API | P0 | channel delete, Meili | TODO |
| ECC-SRCH-C304 | Corner | Search `types=message,issue` while project membership revoked mid-request | fail-closed scope snapshot; **no** foreign project issues | API | P0 | NF-03, NFR-19 | TODO |
| ECC-SRCH-C305 | Corner | Crafted query / filter bypass attempting empty `channel_id $in` | server always injects user channelIDs; empty membership → empty hits not unscoped | API | P0 | SEC-H2 | TODO |
| ECC-SRCH-C306 | Corner | Ctrl+K search + concurrent join public channel (index already has msgs) | after join, hits appear (eventual); before join — no leak | E2E | P1 | NF-03 | TODO |
| ECC-SRCH-C307 | Corner | Saved filter JQL from project A shared/opened by non-member | empty/403; no title/key leak in error body | API | P0 | NF-03, ECC-TRCK-031↑ | TODO |
| ECC-SRCH-C308 | Corner | Index lag issue create + search + assignee-only visibility | no hit for non-members; member miss-until-index OK | API | P1 | C201↑ | TODO |
| ECC-SRCH-C309 | Corner | Message edit (secret→public text) + stale Meili snippet | reindex or TTL; stale secret snippet **не** отдаётся после edit ACL check | API | P0 | NF-03 | TODO |
| ECC-SRCH-C310 | Corner | GDPR user hard-delete: messages anonymized but Meili still has display_name | reindex/purge; search не показывает PII deleted user | API | P0 | US-S03 | TODO |
| ECC-SRCH-C311 | Corner | Multi-type search limit race: huge Meili page + scope filter shrinks to 0 | 200 empty; no unfiltered fallback | API | P1 | TC-SEARCH-01 | TODO |
| ECC-SRCH-C312 | Corner | Search during Meili outage | 503/degraded clear; **не** raw Mongo unscoped scan | API | P0 | NF-03 | TODO |
| ECC-SRCH-C313 | Corner | Forwarded private content indexed under public channel id incorrectly | indexer uses destination ACL; source private never searchable by non-source members | API | P0 | MSG-C317, NF-03 | TODO |
| ECC-SRCH-C314 | Corner | Block+DM: A blocked B; B searches known DM keywords | no hits from DM for B (or A) per block policy; no metadata leak | API | P0 | blocks, NF-03 | TODO |
| ECC-SRCH-C315 | Corner | Unread badge vs search: unread channel msgs soft-deleted then search | unread recalc; search no deleted; badge not stuck | API | P1 | unread race × search | TODO |

---

## 5. Cross-domain combos (MSG × WS × NOTIF × SRCH)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MSG-C325 | Corner | Channel delete mid-WS + in-flight notif fanout + search reindex | WS unsub; notif stop; search empty for channel; no zombie push | Manual | P0 | channel delete mid-WS | TODO |
| ECC-WS-C321 | Corner | NF-02 revoke mid-subscribe + client still receives NATS via bug path | **must not**; integration assert zero events after revoke | API/WS | P0 | NF-02 | TODO |
| ECC-NOTIF-C316 | Corner | Preference mute + WebPush SSRF update attempt + security event | SSRF rejected; security mail/in-app still sent | API | P0 | NF-07 × security | TODO |
| ECC-SRCH-C316 | Corner | Reaction storm updates indexed reaction counts + search snippet | no crash; counts eventual; scope intact | API | P2 | reaction storm | TODO |
| ECC-MSG-C326 | Corner | Idempotent retry after NATS recover + WS reconnect storm overlapping | one msg; one WS event per subscriber; unread +1 once | Manual | P0 | idempotent × storm | TODO |
| ECC-WS-C322 | Corner | Pin cleared (delete msg) event to WS while subscriber membership revoked same ms | revoked client gets **neither** pin nor msg events | WS | P0 | pin+delete × NF-02 | TODO |
| ECC-NOTIF-C317 | Corner | Index lag: mention notif delivered before search finds msg | notif OK; search miss OK; notif deep-link 404 until indexed/visible | E2E | P1 | index lag × notif | TODO |
| ECC-SRCH-C317 | Corner | Admin searches across org while user deactivated mid-query | deactivated excluded; no privilege escalation via stale token in search worker | API | P0 | NF-03, authz | TODO |

---

## Счётчик

| Префикс | Диапазон ID | Count |
|---------|-------------|------:|
| ECC-MSG-C* | C300–C326 | **27** |
| ECC-WS-C* | C300–C322 | **23** |
| ECC-NOTIF-C* | C300–C317 | **18** |
| ECC-SRCH-C* | C300–C317 | **18** |
| **Итого Corner** | | **86** |

Минимум 60 — **выполнен** (86).

**Связь с parent (не дубли, а углубление):**

| Parent | Углублено в C03 |
|--------|-----------------|
| ECC-MSG-C200…204 | C300–C309, C316–C317, C321–C326 (NATS×idempotency, block races, pin×restore, forward×revoke) |
| ECC-WS-C200…203 | C300–C322 (mid-subscribe races, multi-room, storm×auth, channel delete, foreign rooms) |
| ECC-NOTIF-C200…202 | C300–C317 (mute×security matrix, SSRF mid-update/DNS-rebind, debounce, GDPR) |
| ECC-SRCH-C200…201 | C300–C317 (lag×kick, stale Meili, filter bypass, Meili down, block+DM search) |

---

*qa-manual · C03 · 2026-07-10 · каталог не тронут · коммит не создавался*
