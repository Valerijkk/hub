# Corner expansion C03 wave5 — Messenger / WS / Notification / Search

> **Тип:** только **Corner** (редкие комбинации, races, mid-flight, multi-step, cross-domain).  
> **ID ranges:** `ECC-MSG-C700+` · `ECC-WS-C700+` · `ECC-NOTIF-C700+` · `ECC-SRCH-C700+`  
> **Не дублирует** parent `ECC-*-C200…`, wave1 `C300…`, wave2 `C400…`, wave3 `C500…`, wave4 `C600…` — новые комбинации.  
> **Фокус wave5:** multi-hub / pod-drain · clock-skew / DST causality · channel visibility flip · VAPID/JWT rotate mid-flight · cursor/page×ACL · link-preview SSRF · presence/typing under partition · digest×DST · NFC/NFD index · snippet XSS.  
> **Merge target:** `.planning/archive/july-2026/EDGE-CORNER-CATALOG.md` § Corner expansion (после ревью).  
> **Владелец:** `qa-manual` · 2026-07-10 · **каталог не трогать** · **не коммитить**.

Формат: `| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |`

---

## 1. Messenger — `ECC-MSG-C700+` (visibility flip · clock-skew · link-preview · drain)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MSG-C700 | Corner | Channel `public→private` flip **пока** non-member N имеет history cursor mid-page | последующие pages 403/empty; уже загруженные клиентские msgs не дают REST re-fetch; audit flip | API | P0 | visibility flip | TODO |
| ECC-MSG-C701 | Corner | `private→public` flip mid-send от member; non-member search/WS ещё на старом ACL cache | до invalidate — fail-closed (no leak); после — только новые public msgs per policy | API | P0 | visibility flip × NF-03 | TODO |
| ECC-MSG-C702 | Corner | Visibility flip + concurrent invite-link join + first message | joiner sees only post-membership msgs; no pre-private history leak | API | P0 | visibility flip, invite | TODO |
| ECC-MSG-C703 | Corner | Replica clock skew +2m: mute_until=T on primary; sender on skewed node treats as unmuted | mute decision from authoritative UTC store; no mute-bypass via skew | API | P0 | clock-skew, mute | TODO |
| ECC-MSG-C704 | Corner | Cursor `before` from client with skewed clock (future ObjectId/time) interleaved with live inserts | pagination monotonic; no skip/dup storms; no leak across pages | API | P1 | clock-skew, NFR-15 | TODO |
| ECC-MSG-C705 | Corner | Link-preview worker fetches `https://evil` → 302 to `169.254.169.254` mid-send | ValidateOutboundURL / SafeHTTP; preview fail soft; msg body OK; no SSRF | API | P0 | link-preview SSRF, NF-07 | TODO |
| ECC-MSG-C706 | Corner | Link-preview DNS rebind: resolve public at enqueue, private at fetch | re-validate at fetch; fail-closed; no metadata body in preview card | API | P0 | link-preview SSRF | TODO |
| ECC-MSG-C707 | Corner | Pod drain (SIGTERM): handler wrote Mongo msg, outbox not flushed | drain waits outbox or crash-recovery republish once; no silent loss; no dup on peer | Manual | P0 | multi-hub / pod-drain | TODO |
| ECC-MSG-C708 | Corner | Rolling deploy: API-A accepts send, API-B serves history before Mongo visible (read-your-writes miss) | causal: sender sees own msg (sticky/primary read) or documented lag bound; peers eventual | Manual | P1 | pod-drain, replica lag | TODO |
| ECC-MSG-C709 | Corner | Channel archive flip mid-compose + attachment presign already issued | send 403; presign unused/expired; no orphan Seaweed object linked as live msg | API | P1 | archive × storage | TODO |
| ECC-MSG-C710 | Corner | Moderate (admin delete others' msg) + author edit race same version/etag | one terminal: deleted XOR edited-then-deleted; no resurrect body | API | P0 | moderate × edit | TODO |
| ECC-MSG-C711 | Corner | NFC vs NFD identical-looking mention display names in same channel | mentions resolve to correct user ids; no homoglyph privilege; search/notif target right UUID | API | P1 | unicode, mentions | TODO |
| ECC-MSG-C712 | Corner | Thread reply while parent moved to another channel (if move supported) or parent ACL revoked | reply 404/403; no cross-channel orphan thread | API | P1 | thread × ACL | TODO |
| ECC-MSG-C713 | Corner | Export/imex channel dump **while** soft-delete batch + new sends | export consistent snapshot or versioned; no half-deleted bodies; import idempotent | Manual | P1 | imex × live traffic | TODO |
| ECC-MSG-C714 | Corner | Retention purge job (day 30) deletes msg **while** peer downloads attachment | msg gone; file 404 after policy; download mid-flight completes or fails clean; no trash leak to non-admin | API | P1 | retention × storage | TODO |
| ECC-MSG-C715 | Corner | DM create A→B while B's account enters soft-delete/GDPR trash same ms | DM 404/403; no channel left with deleted peer; cleanup job | API | P0 | GDPR × DM | TODO |
| ECC-MSG-C716 | Corner | Rate-limit key by user: NAT shared IP + 5 users burst send | per-user 60/min; IP limit separate; no cross-user 429 starvation OR documented fair share | API | P1 | NFR-07, NAT | TODO |
| ECC-MSG-C717 | Corner | Forward chain: private→private→public mid-chain membership revoke on hop-2 | stop at revoke; public never gets private body; audit trail | API | P0 | forward × NF-03 | TODO |
| ECC-MSG-C718 | Corner | Optimistic UI multi-tab: tab1 deletes, tab2 edits, tab3 reacts — server serializes | consistent final; WS events ordered per msg_id; clients reconcile without ghost | E2E | P1 | multi-tab causality | TODO |

---

## 2. WebSocket — `ECC-WS-C700+` (multi-hub · drain · protocol · presence partition)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-WS-C700 | Corner | Hub pod drain: client sticky to dying pod; Caddy moves to new pod mid-subscribed rooms | re-auth + re-subscribe required; **no** auto room restore without CanSubscribe; no event gap policy documented | Manual | P0 | multi-hub / pod-drain | TODO |
| ECC-WS-C701 | Corner | Two hubs: msg published via NATS; hub-A fans out, hub-B consumer lag then burst | each conn gets ≤1 event/msg_id; order per-channel best-effort; no cross-hub double without dedupe | Manual | P0 | multi-hub fanout | TODO |
| ECC-WS-C702 | Corner | JWT access expires mid-subscribe handshake (upgrade OK, authz frame late) | reject subscribe / close; no room join with expired access | WS | P0 | JWT mid-handshake, NFR-09 | TODO |
| ECC-WS-C703 | Corner | Refresh rotation on tab1 kills refresh family; tab2 WS still open with old access until expiry | after access TTL — close; no refresh via WS frame; rooms cleared | WS/E2E | P0 | session × WS | TODO |
| ECC-WS-C704 | Corner | Presence online on hub-A; network partition; hub-B shows offline; heal | presence converges; no permanent ghost; last-writer/version wins | Manual | P1 | presence under partition | TODO |
| ECC-WS-C705 | Corner | Typing start on channel X; partition; kick from X; heal delivers delayed typing | typing dropped/ignored post-kick; no typing leak to remaining members from kicked user | WS | P0 | typing × NF-02 × partition | TODO |
| ECC-WS-C706 | Corner | Client sends compressed/deflate bomb frame near read limit | close/reject; no OOM; hub stable | Manual | P0 | protocol attack | TODO |
| ECC-WS-C707 | Corner | Mixed protocol: JSON text frame then unexpected binary opcode mid-session | policy close; room cleaned; no panic in Fiber WS | Manual | P1 | protocol skew | TODO |
| ECC-WS-C708 | Corner | Old client schema (missing fields) + new server after deploy | reject/compat path; no unauthz default room; metrics schema_error | WS | P1 | protocol version skew | TODO |
| ECC-WS-C709 | Corner | Subscribe forged prefix `admin:` / `search:` / empty / `channel:`+SQL-ish | reject NF-02; no handler fallthrough to privileged bus | WS | P0 | NF-02, namespace | TODO |
| ECC-WS-C710 | Corner | Slow consumer: send buffer full + concurrent kick | force-unsub + drop; no deadlock; kicked conn closed | WS | P0 | backpressure × kick | TODO |
| ECC-WS-C711 | Corner | Caddy HTTP/2 vs HTTP/1.1 WS upgrade mismatch under reload | connection fails clean or migrates; no half-open privileged room | Manual | P1 | Caddy × WS | TODO |
| ECC-WS-C712 | Corner | Clock skew: server closes on pongWait; client clock thinks ping still valid | server authority; conn closed; client reconnects | WS | P1 | clock-skew, heartbeat | TODO |
| ECC-WS-C713 | Corner | HPA scale-in: 50 conns on pod terminated without drain hook | clients reconnect; membership re-checked; brief gap OK; no authz bypass on reconnect storm | Manual | P0 | HPA × NF-02 | TODO |
| ECC-WS-C714 | Corner | User soft-delete while presence map still lists user in channel roster WS event | roster omits deleted; GDPR; no PII in presence payload after | WS | P0 | GDPR × presence | TODO |
| ECC-WS-C715 | Corner | Concurrent unsubscribe-all + subscribe storm same conn (100 cycles) | final set exact; no ghost hub entries; fanout matches set | WS | P1 | hub consistency | TODO |
| ECC-WS-C716 | Corner | Visibility flip public→private while non-member still subscribed (stale) | force-unsub; further events dropped; NF-02 | WS | P0 | visibility flip × WS | TODO |

---

## 3. Notification — `ECC-NOTIF-C700+` (VAPID rotate · DST digest · deep-link · circuit)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-NOTIF-C700 | Corner | FA-06: VAPID key rotate mid WebPush batch (keys N→N+1) | in-flight with old key may fail once; resubscribe path; no SSRF; other users OK | Manual | P0 | VAPID rotate mid-flight | TODO |
| ECC-NOTIF-C701 | Corner | VAPID rotate + concurrent WebPush endpoint PATCH | endpoint re-validated NF-07; bound to new VAPID; old signature reject | API | P0 | VAPID × NF-07 | TODO |
| ECC-NOTIF-C702 | Corner | DST spring-forward: quiet hours window skips hour; deferred digest flush | no double-flush; no lost deferred; UTC store; local TZ correct | API | P0 | digest × DST | TODO |
| ECC-NOTIF-C703 | Corner | DST fall-back: quiet hours hour repeats; mention deferred twice candidate | exactly-once digest per event id; no dup push | API | P0 | digest × DST | TODO |
| ECC-NOTIF-C704 | Corner | Digest coalesce key collision: issue mention + channel msg same debounce key bug | separate subjects; no cross-entity swallow; security never coalesced away | API | P0 | digest isolation | TODO |
| ECC-NOTIF-C705 | Corner | Deep-link token in push expires before user taps; source msg still live | open → re-auth + ACL re-check; 403 if kicked; no body in expired token | E2E | P1 | deep-link TTL | TODO |
| ECC-NOTIF-C706 | Corner | Deep-link for soft-deleted msg + in-app row still unread | tap → gone/tombstone; mark-read OK; no content leak | API/E2E | P1 | deep-link × soft-delete | TODO |
| ECC-NOTIF-C707 | Corner | TG Bot API circuit open (5xx storm) while mail+in-app healthy | TG skipped/fallback; mail/in-app deliver; circuit per-channel not global panic | API | P0 | circuit breaker | TODO |
| ECC-NOTIF-C708 | Corner | Circuit half-open probe succeeds for user A, fails for B same second | isolation; no thundering herd retry; backoff jitter | API | P1 | circuit breaker | TODO |
| ECC-NOTIF-C709 | Corner | i18n template missing locale mid-render for mention push | fallback locale; no panic; no raw template keys in push body | API | P1 | i18n render | TODO |
| ECC-NOTIF-C710 | Corner | Preference imex import mid-delivery worker read | delivery sees atomic prefs snapshot; no half-applied matrix | API | P1 | prefs imex × delivery | TODO |
| ECC-NOTIF-C711 | Corner | GDPR erase user: pagination page1 of notifs fetched; erase; page2 | page2 empty; no PII rows; push endpoints purged; TG unbound | API | P0 | GDPR × pagination | TODO |
| ECC-NOTIF-C712 | Corner | Org announcement flood (N=1000) + per-user channel mute matrix | mutes respected for chat-class; security/announcement policy documented; no OOM | Manual | P1 | flood × mute | TODO |
| ECC-NOTIF-C713 | Corner | In-app mark-read while OS push already shown; second device still unread | server unread authoritative; devices converge; no negative badge | E2E | P1 | multi-device badge | TODO |
| ECC-NOTIF-C714 | Corner | Mail bounce (550) for security event + TG unbound + WebPush 410 | escalate/fail-visible to admin path; in-app remains; no infinite retry | API | P0 | multi-channel fail | TODO |
| ECC-NOTIF-C715 | Corner | Clock skew worker: quiet-hours evaluator +2h vs user TZ | decision from stored UTC + user TZ rules; skew doesn't bypass quiet or force early security suppress | API | P0 | clock-skew × quiet | TODO |

---

## 4. Search — `ECC-SRCH-C700+` (reindex swap · snippet XSS · NFC · page×ACL)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-SRCH-C700 | Corner | Meili zero-downtime index swap mid Ctrl+K query | 200 with scoped hits OR retry; **never** unscoped/raw; no 500 leak of foreign titles | API | P0 | reindex swap × NF-03 | TODO |
| ECC-SRCH-C701 | Corner | Reindex job dual-write old+new index; delete doc only on old | query path uses alias; deleted not visible; swap atomic from Hub view | API | P0 | reindex swap | TODO |
| ECC-SRCH-C702 | Corner | Snippet highlight contains `<script>` / markdown from msg body | escaped/sanitized in API+UI; no XSS in Ctrl+K | API/E2E | P0 | snippet XSS | TODO |
| ECC-SRCH-C703 | Corner | Snippet with RTL override + homoglyph channel name | display safe; ACL by id not by displayed name; no spoof join | API | P1 | snippet XSS, unicode | TODO |
| ECC-SRCH-C704 | Corner | Same text indexed as NFC and NFD duplicate docs after edit clients | dedupe by msg_id; one hit; delete removes both forms | API | P1 | NFC/NFD index | TODO |
| ECC-SRCH-C705 | Corner | Search page1 OK; between pages user kicked from channel of hit#1 type | page2 must re-scope; drop kicked channel hits; no continuation leak | API | P0 | page×ACL | TODO |
| ECC-SRCH-C706 | Corner | Cursor/offset search: membership added mid-pagination | new channel hits may appear eventual; must not reveal pre-join private history beyond policy | API | P1 | page×ACL | TODO |
| ECC-SRCH-C707 | Corner | Typo-tolerance / prefix match nearly hits private channel unique token | non-member: **no** near-miss leak (title/snippet/score); empty or only scoped | API | P0 | typo × NF-03 | TODO |
| ECC-SRCH-C708 | Corner | Federated multi-type search: Meili OK for issues, timeout for messages | partial degraded **scoped** or 503; no issues-only path that skips message ACL engine incorrectly leaking | API | P0 | partial federation | TODO |
| ECC-SRCH-C709 | Corner | Synonym dict hot-reload mid-query (admin updates) | query consistent snapshot; no crash; ACL unchanged | API | P2 | synonym reload | TODO |
| ECC-SRCH-C710 | Corner | Ranking boost by project activity while project membership revoked mid-score | boost ignored for out-of-scope; hit dropped | API | P0 | ranking × NF-03 | TODO |
| ECC-SRCH-C711 | Corner | Visibility flip private→public: stale Meili ACL facet still private | query-time membership/visibility wins; after flip non-members may see per policy only | API | P0 | visibility flip × search | TODO |
| ECC-SRCH-C712 | Corner | Channel rename while index has old name; search by old and new | hits by id; rename reindex eventual; no ACL bypass via old name alias | API | P1 | rename × index | TODO |
| ECC-SRCH-C713 | Corner | Saved search shared to user then their project role demoted | re-run empty/403; no cached result set replay | API | P0 | saved search × ACL | TODO |
| ECC-SRCH-C714 | Corner | Stopword-only query + private unique id token in `q` crafted | no unscoped scan; empty or safe reject; rate-limit holds | API | P1 | query injection | TODO |
| ECC-SRCH-C715 | Corner | Concurrent edit+delete+restore same msg_id hitting indexer | final Meili state = DB truth (absent or restored body); no secret stale snippet | API | P0 | indexer three-way | TODO |

---

## 5. Cross-domain combos (MSG × WS × NOTIF × SRCH)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MSG-C719 | Corner | Pod drain mid-send + WS hub migrate + notif enqueue + search index | one msg; WS re-sub sees it; notif ≤1; one Meili doc; no orphan | Manual | P0 | pod-drain cross-domain | TODO |
| ECC-WS-C717 | Corner | Visibility flip + force WS unsub + search page in-flight + TG mention queued | WS drop; search re-scope; TG/in-app ACL fail-closed | API/WS | P0 | visibility flip cross | TODO |
| ECC-NOTIF-C716 | Corner | VAPID rotate + deep-link tap + search hit for same msg | push may fail once; tap ACL OK; search scoped; no SSRF on endpoint update | E2E | P0 | VAPID × deep-link × search | TODO |
| ECC-SRCH-C716 | Corner | Snippet XSS payload msg + notif body + WS new_message fanout | all three sanitize/escape; CSP holds in UI | E2E | P0 | snippet XSS cross | TODO |
| ECC-MSG-C720 | Corner | Clock skew mute + quiet-hours notif + search index lag for muted channel | mute/quiet suppress chat notif; search still ACL-scoped; no mute-bypass via skew | API | P0 | clock-skew cross | TODO |
| ECC-WS-C718 | Corner | Presence partition heal + kick + subscribe + Ctrl+K same user | after kick: presence gone, WS deny, search empty for channel | Manual | P0 | presence × NF-02+NF-03 | TODO |
| ECC-NOTIF-C717 | Corner | DST digest flush + channel delete + unread wipe same window | no push with deleted snippets; unread 0; search empty | Manual | P0 | DST × channel delete | TODO |
| ECC-SRCH-C717 | Corner | Link-preview SSRF blocked + msg still searchable by members + WS delivered | preview empty; body indexed scoped; WS OK; outbound fail soft | API | P1 | link-preview × search × WS | TODO |
| ECC-MSG-C721 | Corner | GDPR erase peer in DM: history anonymized, WS roster, notif purge, Meili PII | all four clean; no display_name in search/notif/WS | Manual | P0 | GDPR cross-domain | TODO |
| ECC-WS-C719 | Corner | HPA storm reconnect + rate-limit send + search Ctrl+K spam | WS authz holds; 429 on send; search scoped+limited; server alive | Manual | P0 | storm cross NFR | TODO |

---

## Счётчик

| Префикс | Диапазон ID | Count |
|---------|-------------|------:|
| ECC-MSG-C* | C700–C721 | **22** |
| ECC-WS-C* | C700–C719 | **20** |
| ECC-NOTIF-C* | C700–C717 | **18** |
| ECC-SRCH-C* | C700–C717 | **18** |
| **Итого Corner** | | **78** |

Минимум 55 — **выполнен** (78).

**Фокус-покрытие (wave5, не копируя C200–C600):**

| Фокус | ID (примеры) |
|-------|----------------|
| multi-hub / pod-drain / HPA | MSG-C707–C708, C719; WS-C700–C701, C713, C719 |
| clock-skew / DST causality | MSG-C703–C704; WS-C712; NOTIF-C702–C703, C715, C717; MSG-C720 |
| channel visibility flip | MSG-C700–C702; WS-C716–C717; SRCH-C711 |
| VAPID/JWT rotate mid-flight | NOTIF-C700–C701, C716; WS-C702–C703 |
| cursor/page × ACL | MSG-C700, C704; SRCH-C705–C706, C713 |
| link-preview SSRF | MSG-C705–C706; SRCH-C717 |
| presence/typing under partition | WS-C704–C705, C718 |
| digest × DST / circuit | NOTIF-C702–C704, C707–C708 |
| NFC/NFD + snippet XSS | MSG-C711; SRCH-C702–C704, C716 |
| GDPR / imex live traffic | MSG-C713, C715, C721; NOTIF-C711; WS-C714 |

**Связь с prior waves (углубление, не дубль):**

| Prior | Wave5 углубление |
|-------|------------------|
| ECC-*-C200… | scale-out drain / DST / visibility flip (не basic deny) |
| ECC-*-C300… | не idempotency/block/pin storm — causality & protocol |
| ECC-*-C400… | не split-brain/unbind/subscribe-after-kick — multi-hub & VAPID & page×ACL |
| ECC-*-C500…/C600… | C03 ranges не заняты; IDs стартуют с C700 без пересечения |

---

*qa-manual · C03 wave5 · 2026-07-10 · каталог не тронут · коммит не создавался*
