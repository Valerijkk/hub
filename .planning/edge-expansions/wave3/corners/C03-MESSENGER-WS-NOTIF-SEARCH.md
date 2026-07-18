# Corner expansion C03 wave3 — Messenger / WS / Notification / Search

> **Тип:** только **Corner** (редкие комбинации, races, mid-flight, multi-step, cross-domain).  
> **ID ranges:** `ECC-MSG-C500+` · `ECC-WS-C500+` · `ECC-NOTIF-C500+` · `ECC-SRCH-C500+`  
> **Не дублирует** parent `ECC-*-C200…` · wave1 `ECC-*-C300…` · wave2 `ECC-*-C400…` — углубляет **новыми** комбинациями.  
> **Фокус:** Redis failover/stampede · JWT/session revoke mid-WS+push · cursor×concurrent mutation · sticky-loss multi-node WS · slow-consumer backpressure · VAPID/WebPush key rotate · quiet-hours TZ flip · deep-link×soft-delete×kick · badge Redis≠Mongo · Meili rolling reindex×homoglyph/XSS highlight · channel archive/readonly×thread · attachment ACL mid-download · rolling deploy mid-fanout.  
> **Merge target:** `.planning/july-2026/EDGE-CORNER-CATALOG.md` § Corner expansion (после ревью).  
> **Владелец:** `qa-manual` · 2026-07-10 · **не коммитить в каталог** без ревью.

Формат: `| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |`

---

## 1. Messenger — `ECC-MSG-C500+` (Redis failover · cursor×mutation · archive/readonly · attachment ACL)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MSG-C500 | Corner | Redis primary failover mid-IncrementUnread (ZSET/hash half-written) | after failover unread ≥0 and converges to Mongo/history truth; no permanent +N stampede | API | P0 | Redis failover, unread | TODO |
| ECC-MSG-C501 | Corner | Redis cache stampede: 200 GetUnread after cold miss during failover | single-flight/lock or bounded thundering herd; API p99 bounded; no negative counters | API | P0 | Redis stampede | TODO |
| ECC-MSG-C502 | Corner | Cursor history: client holds `before=M50` while 30 new msgs + soft-delete M40–M45 | page consistent snapshot; no deleted bodies; has_more correct; no skip/dup live msgs | API | P0 | cursor×mutation | TODO |
| ECC-MSG-C503 | Corner | Infinite-scroll two tabs: tab1 cursor older, tab2 inserts mid-page boundary | no duplicate msg_id in either tab after merge; server cursors opaque/stable | E2E | P1 | cursor×mutation | TODO |
| ECC-MSG-C504 | Corner | Channel → archive (or readonly) **while** thread reply in-flight past membership check | 403/410; reply not persisted; reply_count unchanged; WS no phantom reply | API | P0 | archive×thread | TODO |
| ECC-MSG-C505 | Corner | Archive channel + concurrent MarkRead + pin list fetch | MarkRead/pin fail-closed or no-op; no resurrect activity in archived channel | API | P0 | archive, pin | TODO |
| ECC-MSG-C506 | Corner | Un-archive mid-send from client that cached readonly=true | server authz wins; if un-archived allow else 403; no double msg on retry with new key | API | P1 | archive, idempotent | TODO |
| ECC-MSG-C507 | Corner | Attachment: presigned GET issued → member kicked → download within TTL | fail-closed: 403/404 on GET **or** short-lived URL + ACL re-check; no body for kicked | API | P0 | attachment ACL mid-download | TODO |
| ECC-MSG-C508 | Corner | Presign PUT started → channel soft-deleted before complete → msg create with object key | reject msg or orphan GC; object not reachable via public URL; audit | API | P0 | attachment × channel delete | TODO |
| ECC-MSG-C509 | Corner | Link-preview fetch mid-send: URL DNS rebinds to metadata IP after ValidateOutboundURL at enqueue | SafeHTTP re-validate; no SSRF; msg without preview OK | API | P0 | NF-07, link-preview | TODO |
| ECC-MSG-C510 | Corner | Edit msg (secret→benign) + concurrent forward of **pre-edit** snapshot | forward uses ACL+version policy; secret not published to wider channel; audit versions | API | P0 | edit×forward, NF-03 | TODO |
| ECC-MSG-C511 | Corner | @mention soft-deleted user + @channel in same msg while mute until T | no crash; deleted user mention ignored/tombstone; mute rules for @channel per policy; security unaffected | API | P1 | mention × mute | TODO |
| ECC-MSG-C512 | Corner | Star/bookmark msg → soft-delete msg → list starred + restore from trash | starred list drops deleted; restore does **not** auto-restar unless policy; no dangling UUID | API | P1 | star × soft-delete | TODO |
| ECC-MSG-C513 | Corner | Mongo secondary lag: write primary, history read secondary mid-failover | read-your-writes or documented stale window; UI refresh converges; no cross-user leak | Manual | P0 | Mongo lag | TODO |
| ECC-MSG-C514 | Corner | Presence away flip (5m) **same second** as DM send + typing stop | presence eventual; msg delivered; typing cleared; no stuck «typing» for away peer | API/WS | P1 | presence × DM | TODO |
| ECC-MSG-C515 | Corner | Channel settings version optimistic lock: two admins rename+topic concurrent | one 200 version++; other 409; WS one canonical name; no split-brain title in list | API | P1 | optimistic lock | TODO |
| ECC-MSG-C516 | Corner | Export/IMEX channel dump started → bulk soft-delete 50% msgs mid-export | export consistent snapshot or fail; no partial secret leak in incomplete file without ACL | Manual | P1 | imex × soft-delete | TODO |
| ECC-MSG-C517 | Corner | Redis unread key TTL expire mid-MarkRead + new msg | MarkRead + increment ordered; badge matches GetUnread; no stuck ghost unread | API | P0 | Redis TTL × unread | TODO |

---

## 2. WebSocket — `ECC-WS-C500+` (sticky-loss · JWT revoke · slow-consumer · rolling restart)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-WS-C500 | Corner | Multi-node: client sticky lost → reconnect lands on node-B while node-A still has room membership | no double fanout; old node cleans; CanSubscribe re-checked on B; NF-02 intact | Manual | P0 | sticky-loss | TODO |
| ECC-WS-C501 | Corner | Sticky-loss mid-kick: kick applied on node-A, client reconnects to node-B before kick event | B deny subscribe; A force-unsub; zero post-kick events on either node | Manual | P0 | sticky-loss × NF-02 | TODO |
| ECC-WS-C502 | Corner | JWT access expired + refresh **revoked** (logout all) mid open WS | connection killed; further frames dropped; refresh 401; rooms cleared | WS | P0 | JWT revoke mid-WS | TODO |
| ECC-WS-C503 | Corner | Session revoke on device-1 while device-2 same user still WS-subscribed | only revoked session dies **or** all sessions per policy; no cross-device authz bypass | WS | P0 | JWT revoke, multi-device | TODO |
| ECC-WS-C504 | Corner | Slow consumer: client reads 1 frame/5s while channel floods 60 msg/min | backpressure/drop/close policy; hub not OOM; other subscribers unaffected | Manual | P0 | slow-consumer | TODO |
| ECC-WS-C505 | Corner | Slow consumer + kick mid-buffer | buffered frames for kicked user discarded; no flush-after-kick leak | WS | P0 | slow-consumer × NF-02 | TODO |
| ECC-WS-C506 | Corner | Rolling deploy: WS hub drain — unsubscribe all → new pods; in-flight subscribe | clients reconnect; re-authz; no silent room without CanSubscribe | Manual | P0 | rolling deploy | TODO |
| ECC-WS-C507 | Corner | WriteDeadline exceeded mid-large fanout frame to one conn | that conn closed/cleaned; room continues for others; no panic send-on-closed | WS | P0 | write deadline | TODO |
| ECC-WS-C508 | Corner | Redis pub membership cache fail → WS CanSubscribe during Redis outage | fail-closed deny **or** strong PG read; **never** allow on cache error | WS | P0 | Redis failover × NF-02 | TODO |
| ECC-WS-C509 | Corner | Subscribe `issue:I` / `project:P` while project membership revoked mid-handshake | reject; no tracker event leak on WS | WS | P0 | NF-02, tracker rooms | TODO |
| ECC-WS-C510 | Corner | Presence broadcast storm (200 heartbeats) + concurrent GDPR user delete | deleted user purged from presence; no ghost online; WS personal rooms dead | Manual | P0 | presence × GDPR | TODO |
| ECC-WS-C511 | Corner | Client sends subscribe with room id = another user's UUID forged as `channel:` | CanSubscribe DB check fail; reject; no enumeration via distinct error timing beyond policy | WS | P0 | NF-02 | TODO |
| ECC-WS-C512 | Corner | Unsubscribe-all + immediate multi-room resubscribe during JWT refresh single-flight | after refresh: only rooms still allowed; no subscribe with stale token claims alone | WS | P1 | JWT refresh × NF-02 | TODO |
| ECC-WS-C513 | Corner | Hub metrics: 10k rooms × one user kicked from 100 rooms in 1s | force-unsub completes; latency bounded; no missed unsub leaving zombie | Manual | P1 | kick scale | TODO |
| ECC-WS-C514 | Corner | Binary frame + text subscribe interleaved on same conn during reconnect storm | protocol error close; no partial room join; no crash | Manual | P1 | Fiber WS | TODO |
| ECC-WS-C515 | Corner | Clock skew: server NTP jump −5m mid-pongWait | connections reaped or survive per policy; no mass false-dead; clients reconnect cleanly | Manual | P1 | clock skew | TODO |

---

## 3. Notification — `ECC-NOTIF-C500+` (VAPID rotate · TZ quiet-hours · deep-link · badge divergence)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-NOTIF-C500 | Corner | VAPID/WebPush key rotation: old endpoint still subscribed, worker uses new keys mid-batch | re-subscribe or fail-closed prune; no send with mismatched keys leaking payload to wrong party; other users OK | API | P0 | VAPID rotate | TODO |
| ECC-NOTIF-C501 | Corner | VAPID rotate **same second** as WebPush 410 Gone from vendor | prune old; no infinite retry; in-app fallback once | API | P0 | VAPID × 410 | TODO |
| ECC-NOTIF-C502 | Corner | User TZ preference flip UTC→UTC+12 while quiet-hours deferred queue flushes | quiet evaluation uses **new** TZ at flush **or** documented snapshot; no double delivery; security still immediate | API | P0 | quiet-hours TZ flip | TODO |
| ECC-NOTIF-C503 | Corner | Quiet-hours boundary ±1s with NTP step on worker | at-most-one delivery path; no skip of security; audit timestamps UTC | API | P1 | quiet-hours, NFR-12 | TODO |
| ECC-NOTIF-C504 | Corner | Deep-link token in push → source msg soft-deleted → user opens link | 404/gone UI; **no** msg body in notif cache; search empty | E2E | P0 | deep-link × soft-delete | TODO |
| ECC-NOTIF-C505 | Corner | Deep-link + kick from channel between push display and click | open → 403/404; no WS auto-subscribe; no title leak beyond policy | E2E | P0 | deep-link × kick | TODO |
| ECC-NOTIF-C506 | Corner | Badge: Redis counter=5, Mongo inbox unread=3 after partial worker crash | reconcile job/API GetUnreadNotif returns one truth; UI not stuck | API | P0 | badge Redis≠Mongo | TODO |
| ECC-NOTIF-C507 | Corner | Mark-read-all in-app while Redis INCR from NATS flood | final badge=only newer; no lost-update all-zero falsely then stuck | API | P0 | badge race | TODO |
| ECC-NOTIF-C508 | Corner | Digest/coalesce window: preference flip off mid-digest build | digest omitted or stripped; no send after off except security | API | P1 | digest × prefs | TODO |
| ECC-NOTIF-C509 | Corner | Dedup key collision: mention + assign same entity_id different types same TTL | both or coalesced per matrix; **no** drop of security twin | API | P1 | dedup | TODO |
| ECC-NOTIF-C510 | Corner | Poison NATS notif payload (invalid UUID) in consumer | skip/DLQ; no crash loop; subsequent jobs process | API | P0 | poison message | TODO |
| ECC-NOTIF-C511 | Corner | Telegram Bot API global 429 shared across users mid-unbind for one user | only affected backoff; unbound user stops; others drain with jitter; no cross-user drop forever | API | P1 | TG 429 × unbind | TODO |
| ECC-NOTIF-C512 | Corner | Call invite + calendar reminder + chat mention same second (3 channels) | all eligible per prefs; no single mutex dropping two; ordering stable in inbox | API | P1 | multi-type fanout | TODO |
| ECC-NOTIF-C513 | Corner | Session-revoke security mail + JWT revoke mid-WS + push enqueue | security delivered; WS dead; chat pushes stop for revoked session | API | P0 | JWT revoke × security | TODO |
| ECC-NOTIF-C514 | Corner | Notif retention purge job deletes old rows while user has inbox open + MarkRead | no 500; badge ≥0; purged ids 404 on get-by-id | API | P1 | retention | TODO |

---

## 4. Search — `ECC-SRCH-C500+` (rolling reindex · homoglyph · highlight XSS · ranking leak)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-SRCH-C500 | Corner | Meili full rolling reindex (swap index) mid Ctrl+K query | 200 empty/degraded **or** consistent hit set; **never** unscoped; no 500 leak | API | P0 | rolling reindex | TODO |
| ECC-SRCH-C501 | Corner | Reindex swap: old index has private hit, new index empty for query — dual-read window | fail-closed to scoped empty/new only; no private from old after cutover | API | P0 | rolling reindex, NF-03 | TODO |
| ECC-SRCH-C502 | Corner | Homoglyph query (Cyrillic «а» vs Latin «a») matching private channel keyword | ACL filter still applies; homoglyph does not bypass NF-03 | API | P0 | homoglyph × NF-03 | TODO |
| ECC-SRCH-C503 | Corner | Zero-width / bidi override chars in q + indexed body | sanitized; no hit that bypasses scope; no UI spoof in snippet alone granting access | API | P0 | ZWJ/bidi | TODO |
| ECC-SRCH-C504 | Corner | Highlight snippet contains `<script>` / markdown from msg body | escaped/sanitized in API+UI; no XSS; ACL still drops unauthorized | API/E2E | P0 | highlight XSS | TODO |
| ECC-SRCH-C505 | Corner | Ranking boost / typoTolerance returns near-match from **out-of-scope** project | Hub post-filter drops; response never includes foreign key/title | API | P0 | ranking leak, NF-03 | TODO |
| ECC-SRCH-C506 | Corner | Typo tolerance expands token → matches deleted+live; deleted still in Meili | deleted excluded; live scoped; highlight not from deleted doc | API | P0 | typo × soft-delete | TODO |
| ECC-SRCH-C507 | Corner | Pagination: offset page2 built from Meili, meanwhile kick removes half hits | page may shrink; **no** filler from outside ACL to «pad» limit | API | P0 | pagination × kick | TODO |
| ECC-SRCH-C508 | Corner | Ctrl+K debounce: stale response for query-A arrives after query-B UI | UI ignores stale; network layer cancel/seq; no flash of wrong ACL set | E2E | P1 | debounce stale | TODO |
| ECC-SRCH-C509 | Corner | Project key rename mid-search (HUB→ACME) while Meili has old key | hits use current key or reindex; no dual keys leaking extra projects | API | P1 | rename × index | TODO |
| ECC-SRCH-C510 | Corner | Message moved/forwarded channel A→B mid-index; search as member-only-A | no hit via A ACL after move; B members see per B ACL; no dual | API | P0 | move × NF-03 | TODO |
| ECC-SRCH-C511 | Corner | Meili outage + Redis search-cache hit from **previous** user session reused | cache key must be user-scoped; **no** cross-user cache hit | API | P0 | cache × NF-03 | TODO |
| ECC-SRCH-C512 | Corner | Analytics/audit log of search q contains secrets from private msgs? | log q only, not hit bodies; PII policy; no full snippet in info logs | Manual | P1 | privacy | TODO |
| ECC-SRCH-C513 | Corner | Multi-type search: issue+message+user; user soft-deleted mid-request | user hits omitted; others scoped; no 500 | API | P1 | GDPR mid-search | TODO |
| ECC-SRCH-C514 | Corner | Synonym/stop-word dict hot-reload mid-query storm | results eventual; always scoped; no empty-filter bypass when tokens stripped to empty | API | P1 | tokenizer | TODO |
| ECC-SRCH-C515 | Corner | Export search results CSV while membership revoked mid-stream | stream aborts/filters; file no foreign rows; partial file not downloadable unscoped | Manual | P0 | export × NF-03 | TODO |

---

## 5. Cross-domain combos (MSG × WS × NOTIF × SRCH)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MSG-C518 | Corner | Redis flushall mid: unread + WS presence + notif badge + search cache | all rebuild fail-closed; NF-02/03 hold; badges converge; no allow-on-empty-cache | Manual | P0 | Redis flush cross-domain | TODO |
| ECC-WS-C516 | Corner | JWT revoke mid-WS + deep-link push click + search for channel keywords | WS dead; deep-link 401/403; search empty if kicked/session dead | E2E | P0 | JWT revoke × deep-link × NF-03 | TODO |
| ECC-NOTIF-C515 | Corner | VAPID rotate + quiet-hours TZ flip + security event same minute | security delivered (mail/in-app); WebPush rekeyed or pruned; no mute of security | API | P0 | VAPID × TZ × security | TODO |
| ECC-SRCH-C516 | Corner | Rolling Meili reindex + channel archive + WS subscribers still in room | search no hits for archived; WS force-unsub or readonly events only; REST deny writes | Manual | P0 | reindex × archive × WS | TODO |
| ECC-MSG-C519 | Corner | Attachment ACL revoke mid-download + notif preview URL + search filename hit | download deny; notif no reachable secret URL; search hit without exposing file bytes | API | P0 | attachment × notif × search | TODO |
| ECC-WS-C517 | Corner | Sticky-loss + slow-consumer buffer + kick on node-A | after reconnect B: deny; no flush of A-buffer to client; search empty | Manual | P0 | sticky × slow-consumer × kick | TODO |
| ECC-NOTIF-C516 | Corner | Badge Redis≠Mongo + MarkRead-all + new msg WS + Ctrl+K | badge/inbox/WS/search agree after reconcile; no ghost notif body in search | API/E2E | P0 | badge × WS × search | TODO |
| ECC-SRCH-C517 | Corner | Homoglyph query + highlight XSS payload + kicked user | empty hits; if any escape path — still no XSS and no ACL bypass | API | P0 | homoglyph × XSS × NF-03 | TODO |
| ECC-MSG-C520 | Corner | Rolling BE deploy mid-send: Mongo OK, NATS publish to draining pod, WS on new pod | exactly-one delivery after drain; idempotent retry safe; search index once | Manual | P0 | rolling deploy cross-domain | TODO |
| ECC-WS-C518 | Corner | Clock skew NTP −10m: presence stale cleanup + quiet-hours + pongWait mass reconnect | systems recover; no permanent mute of security; NF-02 on resubscribe | Manual | P1 | clock skew cross-domain | TODO |

---

## Счётчик

| Префикс | Диапазон ID | Count |
|---------|-------------|------:|
| ECC-MSG-C* | C500–C520 | **21** |
| ECC-WS-C* | C500–C518 | **19** |
| ECC-NOTIF-C* | C500–C516 | **17** |
| ECC-SRCH-C* | C500–C517 | **18** |
| **Итого Corner** | | **75** |

Минимум 55 — **выполнен** (75).

**Фокус-покрытие (wave3, не копируя C200/C300/C400):**

| Фокус | ID (примеры) |
|-------|----------------|
| Redis failover/stampede/flush | MSG-C500–C501, C517, WS-C508, MSG-C518 |
| JWT/session revoke mid-WS+push | WS-C502–C503, C512, NOTIF-C513, WS-C516 |
| cursor×concurrent mutation | MSG-C502–C503 |
| sticky-loss multi-node WS | WS-C500–C501, C517 |
| slow-consumer backpressure | WS-C504–C505, C517 |
| VAPID/WebPush key rotate | NOTIF-C500–C501, C515 |
| quiet-hours TZ flip | NOTIF-C502–C503, C515 |
| deep-link×soft-delete×kick | NOTIF-C504–C505, WS-C516 |
| badge Redis≠Mongo | NOTIF-C506–C507, C516 |
| Meili rolling reindex | SRCH-C500–C501, C516 |
| homoglyph / ZWJ / highlight XSS | SRCH-C502–C504, C517 |
| ranking/pagination ACL leak | SRCH-C505–C507, C511, C515 |
| channel archive/readonly×thread | MSG-C504–C506, SRCH-C516 |
| attachment ACL mid-download | MSG-C507–C508, C519 |
| rolling deploy mid-fanout | WS-C506, MSG-C520 |

**Связь с prior waves (углубление, не дубль):**

| Prior | Wave3 углубление |
|-------|------------------|
| ECC-*-C200… / C300… | Redis/Mongo lag, cursor races, archive, attachment ACL, VAPID, TZ quiet-hours |
| ECC-*-C400… | sticky-loss / slow-consumer / rolling reindex / homoglyph+XSS — не split-brain/unbind/subscribe-after-kick дословно |

---

*qa-manual · C03 wave3 · 2026-07-10 · каталог не тронут · коммит не создавался*
