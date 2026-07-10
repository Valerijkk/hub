# Corner expansions wave4 — Messenger / WS / Notification / Search (C600+)

> **Тип:** только **Corner** (poll/invite/bookmark/voice races · member-batch∩readonly · PWA offline msg replay · multi-pod rate/WS sticky · tgLinkCodeTTL∩unbind · WebPush TTL∩quiet · saved-filter∩ACL · search shard-limit∩kick · mention_groups∩mute).  
> **Не дублирует** parent (`ECC-*-C200…`), wave1 (`ECC-*-C300…`), wave2 (`ECC-*-C400…`), wave3 (`ECC-*-C500…`) и Edge `E03 *400+`.  
> **ID:** `ECC-MSG-C600+`, `ECC-WS-C600+`, `ECC-NOTIF-C600+`, `ECC-SRCH-C600+`.  
> **Gap:** `TODO` | `Manual` | `Future`.  
> **Код-якоря:** `messenger/service_polls.go`, `service_channel_invites.go`, bookmarks/star, voice fields, `ws/hub.go` send=256, `notification/telegram_bind.go` tgLinkCodeTTL=5m, `push.go` TTL=60s, `search/service.go` defaultSearchLimit=10 / clamp≤50, saved-filter name≤200.  
> **Владелец:** `qa-manual` · 2026-07-10 · wave4 · каталог `EDGE-CORNER-CATALOG.md` **не трогать**.

Формат: `| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |`

---

## 1. Messenger — polls · invites · bookmarks · voice · batch — `ECC-MSG-C600+`

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MSG-C600 | Corner | Poll open → 2 voters vote same option same ms → ClosePoll mid third vote | counts exact unique; third 400 after close; no negative/overflow totals | API | P0 | poll vote×close | TODO |
| ECC-MSG-C601 | Corner | Anonymous poll: voter A votes → soft-delete msg/poll → B fetches results | 404/gone; no voter_ids leak even if stale cache; counts not exposed post-delete | API | P0 | poll×soft-delete | TODO |
| ECC-MSG-C602 | Corner | Multi-choice poll: change vote set while admin closes + kick voter mid-change | close wins or vote atomic; kicked voter 403; final options frozen | API | P0 | poll×kick | TODO |
| ECC-MSG-C603 | Corner | Poll in thread parent; parent soft-deleted while reply+vote in-flight | vote/reply fail-closed; no orphan poll doc searchable | API | P1 | poll×thread | TODO |
| ECC-MSG-C604 | Corner | Invite max_uses=1: two acceptors race same token (two pods) | exactly one member; other «исчерпан»; use_count=1; no dual membership | API | P0 | invite race multi-pod | TODO |
| ECC-MSG-C605 | Corner | Invite expires_in=1h: accept at ExpiresAt−50ms on pod A clock vs +50ms pod B skew | UTC store authoritative; skew cannot extend; loser 410 | API | P0 | invite TTL×skew | TODO |
| ECC-MSG-C606 | Corner | Invite revoke mid AcceptInvite after token validate, before AddMember commit | fail-closed; no member; audit revoke; token dead | API | P0 | invite revoke mid-accept | TODO |
| ECC-MSG-C607 | Corner | Invite accept → channel flipped readonly + first message from joiner same second | join OK; send 403 if readonly; no ghost first msg | API | P0 | invite×readonly | TODO |
| ECC-MSG-C608 | Corner | Invite link shared; channel soft-deleted mid accept | 404/410; no resurrect channel; no membership on trash | API | P0 | invite×channel delete | TODO |
| ECC-MSG-C609 | Corner | Bookmark msg M → M soft-deleted → ListBookmarks + open deep-link | bookmark gone/tombstone; no body leak; remove idempotent | API | P0 | bookmark×soft-delete | TODO |
| ECC-MSG-C610 | Corner | Star channel + concurrent LeaveChannel / kick | star cleared; starred filter omits; no starred ghost name | API | P0 | star×kick | TODO |
| ECC-MSG-C611 | Corner | Two tabs: tab1 AddBookmark, tab2 RemoveBookmark same msg_id race | terminal empty or one bookmark; no dup rows; list consistent | API | P1 | bookmark race | TODO |
| ECC-MSG-C612 | Corner | Voice msg: presign OK → Seaweed PUT mid → kick sender before SendMessage commit | no live msg; orphan object GC/TTL; peers no voice bubble | API | P0 | voice×kick×storage | TODO |
| ECC-MSG-C613 | Corner | Voice duration valid + voice_url host flips to private IP between validate and fetch (preview) | ValidateOutboundURL fail-closed; msg text-only or reject; no SSRF | API | P0 | voice SSRF mid | TODO |
| ECC-MSG-C614 | Corner | AddMembers batch 20 UUIDs: 1 deactivated, 1 already member, 1 foreign org, rest OK | partial policy documented; no foreign org add; deactivated skipped; exact member delta | API | P0 | member batch | TODO |
| ECC-MSG-C615 | Corner | AddMembers batch mid readonly flip + concurrent RemoveMember of actor | actor 403; batch abort or none applied; no half-batch privilege | API | P0 | batch×readonly×RBAC | TODO |
| ECC-MSG-C616 | Corner | mention_groups=@eng + channel mute + security-class event same channel | mute suppresses chat mention push; security still delivered; group expand correct UUIDs | API | P0 | mention_groups×mute | TODO |
| ECC-MSG-C617 | Corner | mention_groups contains removed member UUID mid-send expand | skip removed; no notif to kicked; send OK for live members | API | P0 | mention_groups×kick | TODO |
| ECC-MSG-C618 | Corner | Forward poll message to other channel while source poll closes + option deleted | forward snapshot or 409; target no live vote on closed source; ACL on target | API | P1 | forward×poll | TODO |
| ECC-MSG-C619 | Corner | PWA offline: queue 5 sends; online after kick from channel | all 403; queue drain; no replay into channel; optimistic UI rollback | E2E | P0 | PWA offline replay | TODO |
| ECC-MSG-C620 | Corner | PWA SW caches channel history 200; after soft-delete of last page msgs, SW serves stale | client must revalidate; API truth without deleted; no stale secret body as live | E2E | P0 | PWA/SW history | TODO |
| ECC-MSG-C621 | Corner | Multi-pod: rate 60/min — 40 on pod A + 25 on pod B same user window | shared Redis; ≤60 OK then 429; no per-pod double budget | API | P0 | multi-pod msg rate | TODO |
| ECC-MSG-C622 | Corner | DM create A→B while B blocks A mid-handler on other pod | 403; no DM channel; block wins | API | P0 | DM×block multi-pod | TODO |
| ECC-MSG-C623 | Corner | Edit msg + reaction storm + bookmark add same msg_id three-way | consistent final body/counts/bookmark; WS ordered per msg_id | API/WS | P1 | edit×reaction×bookmark | TODO |

---

## 2. WebSocket — buffer×kick · sticky · PWA · multi-pod hub — `ECC-WS-C600+`

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-WS-C600 | Corner | Send buffer 256 full + kick mid-drop path | force-unsub; buffer discarded; no post-kick frame delivery; hub entry gone | WS | P0 | buffer×kick | TODO |
| ECC-WS-C601 | Corner | Buffer full on channel X; catch-up NATS burst after client resumes | no deadlock; dropped oldest policy documented; membership still enforced | WS | P0 | buffer×NATS burst | TODO |
| ECC-WS-C602 | Corner | Sticky session broken: subscribe on hub-A; kick event consumed only on hub-B | after heal / pub — A unsubscribes K; no zombie fanout on A | Manual | P0 | multi-pod sticky≠truth | TODO |
| ECC-WS-C603 | Corner | Client reconnects to new pod after Caddy reload; auto re-subscribe from SW cache room list | each room CanSubscribe re-check; kicked rooms rejected; no blind restore | WS/E2E | P0 | PWA/SW re-sub | TODO |
| ECC-WS-C604 | Corner | Offline PWA: WS closed; buffered client frames include subscribe to kicked channel | on reconnect reject; frames not applied; NF-02 | E2E | P0 | PWA offline WS | TODO |
| ECC-WS-C605 | Corner | Invite accept success → client WS subscribe before membership visible on hub pod cache | fail-closed until strong read/invalidate; then OK; no pre-join history events | WS | P0 | invite×WS cache | TODO |
| ECC-WS-C606 | Corner | Poll close event fanout while subscriber send-buffer saturated | close event eventually or drop+resync; vote API still closed; no panic | WS | P1 | poll×buffer | TODO |
| ECC-WS-C607 | Corner | Voice msg WS event with attachment URL; recipient kicked between enqueue and write | no event to kicked; URL not delivered; REST also 403 | WS | P0 | voice×kick fanout | TODO |
| ECC-WS-C608 | Corner | Multi-tab: tab1 buffer-full slow consumer; tab2 healthy; kick user | both tabs unsub; slow tab closed; no leak via healthy tab after kick | WS | P0 | multi-tab×kick | TODO |
| ECC-WS-C609 | Corner | Register queue pressure (64) + subscribe storm (256) during rolling deploy | backpressure; no authz bypass; excess reject/close; hubs stable | Manual | P0 | hub capacity deploy | TODO |
| ECC-WS-C610 | Corner | PingPeriod boundary: client silent ~54s while membership revoked | close on timeout **and** revoke; no late pong resurrecting room | WS | P1 | ping×revoke | TODO |
| ECC-WS-C611 | Corner | Subscribe `channel:X` + parallel REST LeaveChannel on other pod | WS force-unsub ≤TTL; further events 0; re-sub reject | WS/API | P0 | leave multi-pod | TODO |
| ECC-WS-C612 | Corner | Bookmark/star change WS notify while user unsubscribed from channel | no require channel sub for personal events OR scoped; no channel body leak in personal frame | WS | P1 | star WS scope | TODO |
| ECC-WS-C613 | Corner | Forward msg creates target event; source member not in target; source WS must not get target room events | isolation; only target subscribers; NF-02 | WS | P0 | forward fanout ACL | TODO |
| ECC-WS-C614 | Corner | GDPR soft-delete user mid multi-room WS on two pods | all conns killed both pods; refresh 401; presence purged | Manual | P0 | GDPR×multi-pod WS | TODO |
| ECC-WS-C615 | Corner | mention_groups expand triggers per-user notif frames; one mentioned user muted channel | muted user no chat WS push (or in-app only per matrix); others OK | WS | P1 | mention_groups×mute WS | TODO |
| ECC-WS-C616 | Corner | Client sends subscribe with invite token as room id / forged prefix | reject; no invite-accept via WS; NF-02 namespace | WS | P0 | invite≠WS authz | TODO |

---

## 3. Notification — tgLink TTL · WebPush TTL · quiet · digest — `ECC-NOTIF-C600+`

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-NOTIF-C600 | Corner | tgLinkCodeTTL=5m: bind confirm at T−1s vs T+1s; parallel unbind start | −1s may bind; +1s 410; unbind cancels pending code; no login via TG | API | P0 | tgLinkCodeTTL | TODO |
| ECC-NOTIF-C601 | Corner | Bind code reused after success within TTL window | second 400/410; no hijack to other user; single telegram_id | API | P0 | tg code single-use | TODO |
| ECC-NOTIF-C602 | Corner | WebPush TTL=60s: enqueue at T0; worker delayed 61s; quiet hours also active | drop/expire push; in-app remains; no late loud push after quiet if policy | API | P0 | WebPush TTL×quiet | TODO |
| ECC-NOTIF-C603 | Corner | WebPush TTL mid-batch: items 1–3 within 60s, 4–10 expired | 1–3 attempt; 4–10 skip/fallback; no crash; spacing 50ms intact | API | P1 | WebPush TTL batch | TODO |
| ECC-NOTIF-C604 | Corner | Quiet hours start mid multi-channel dispatch (TG+WebPush+in-app) | chat-class defer; security immediate; TG/WebPush respect matrix | API | P0 | quiet mid-dispatch | TODO |
| ECC-NOTIF-C605 | Corner | Preference flip WebPush-off while TTL timer still holding payload | cancel/skip push; in-app per prefs; no send after off commit | API | P0 | prefs×TTL | TODO |
| ECC-NOTIF-C606 | Corner | Poll vote notif coalesce: 20 votes < debounce + poll close | coalesced vote digest; close may separate; no 20 identical pushes | API | P1 | poll×debounce | TODO |
| ECC-NOTIF-C607 | Corner | Invite accepted → welcome notif; invite revoked race (second device) | at most one welcome; revoked acceptor no notif body with channel secrets | API | P1 | invite×notif | TODO |
| ECC-NOTIF-C608 | Corner | Bookmark reminder / star digest job + channel kick before flush | skip kicked; no deep-link content; job idempotent | API | P1 | star×kick notif | TODO |
| ECC-NOTIF-C609 | Corner | Voice mention push includes duration metadata; soft-delete voice msg before tap | deep-link gone; no audio URL in push after delete; mark-read OK | API/E2E | P0 | voice×soft-delete push | TODO |
| ECC-NOTIF-C610 | Corner | mention_groups expand 50 users; 10 muted, 2 unbound TG, 1 GDPR deleted | matrix per user; deleted skipped; muted chat suppressed; security N/A | API | P0 | mention_groups fanout | TODO |
| ECC-NOTIF-C611 | Corner | Multi-pod workers: same delivery_id leased twice during Redis blip | exactly-once side effects; idempotent key; badge +1 once | Manual | P0 | multi-pod notif lease | TODO |
| ECC-NOTIF-C612 | Corner | PWA: OS shows push; SW click opens channel after user kicked | ACL re-check; 403 UI; no cached SW HTML with msg body | E2E | P0 | PWA push×kick | TODO |
| ECC-NOTIF-C613 | Corner | Read-all in-app while WebPush TTL delivery still in-flight | badge 0 after read-all; late push must not resurrect unread wrongly (or mark sync) | API | P1 | read-all×TTL | TODO |
| ECC-NOTIF-C614 | Corner | TG 35ms spacing storm + rate-limit user + unbind mid-storm | spacing/backoff; unbind stops TG; no global worker stall | API | P1 | TG spacing×unbind | TODO |
| ECC-NOTIF-C615 | Corner | Saved-filter alert (if enabled) fires on issue hit; project membership revoked mid | no alert; NF-03; no title in push | API | P0 | saved-filter×ACL | Future |
| ECC-NOTIF-C616 | Corner | Security session-revoke mail + WebPush endpoint 410 + TG unbound same second | mail/in-app succeed; endpoints pruned; no infinite retry; visible fail path | API | P0 | multi-channel fail | TODO |

---

## 4. Search — shard limit · saved-filter · poll/invite index · PWA — `ECC-SRCH-C600+`

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-SRCH-C600 | Corner | defaultSearchLimit=10: Meili returns 10; 3 fail ACL mid-build (kick) | response ≤10 scoped; may be <10; **never** backfill unscoped to fill 10 | API | P0 | shard-limit×kick | TODO |
| ECC-SRCH-C601 | Corner | Client asks limit=50; membership shrinks mid federated multi-type merge | clamp≤50; all types re-scoped; no foreign fill | API | P0 | limit clamp×ACL | TODO |
| ECC-SRCH-C602 | Corner | Per-shard overflow: shard A private hits filtered to 0; shard B public OK | only B; empty A not error; no cross-shard ACL skip | API | P0 | per-shard NF-03 | TODO |
| ECC-SRCH-C603 | Corner | SavedFilter name max=200: rename to 200 OK while sharing filter to non-member | rename OK; share 403; name not leak via error to stranger | API | P0 | saved-filter ACL | TODO |
| ECC-SRCH-C604 | Corner | SavedFilter execute: filter references channel user left mid-query | channel clause empty/drop; no hits from left channel; 200 | API | P0 | saved-filter×leave | TODO |
| ECC-SRCH-C605 | Corner | Index poll question + options; soft-delete poll; search option text | no hit; snippet absent; Meili delete eventual | API | P0 | poll index×delete | TODO |
| ECC-SRCH-C606 | Corner | Invite token string accidentally in msg body; non-member searches token | no hit for non-members; members scoped; token not global | API | P0 | invite token≠public index | TODO |
| ECC-SRCH-C607 | Corner | Bookmark-only msgs: user bookmarks then loses channel ACL; Ctrl+K keyword | no hit; bookmarks list tombstone; search NF-03 | API | P0 | bookmark×search ACL | TODO |
| ECC-SRCH-C608 | Corner | Voice transcript/filename indexed; kick before search | no hit; no voice_url in any search payload | API | P0 | voice×NF-03 | TODO |
| ECC-SRCH-C609 | Corner | mention_groups name collision across channels; search group label | hits only member channels; no foreign group metadata | API | P1 | mention_groups search | TODO |
| ECC-SRCH-C610 | Corner | Multi-pod: indexer on A, query on B; membership revoke on A not yet in B cache | query-time membership re-check; stale allowlist ignored | API | P0 | multi-pod search ACL | TODO |
| ECC-SRCH-C611 | Corner | PWA Ctrl+K: SW caches previous hit set; after kick replays cache | UI must not show stale foreign hits; network revalidate; API empty | E2E | P0 | PWA search cache | TODO |
| ECC-SRCH-C612 | Corner | Forwarded poll body indexed under target channel; source private unique token | non-source members: no source token hit; target ACL only | API | P0 | forward×index ACL | TODO |
| ECC-SRCH-C613 | Corner | Reindex job + saved-filter alert cron same second + project archive | archive → no hits/alerts; jobs idempotent; no title leak | API | P1 | reindex×archive | TODO |
| ECC-SRCH-C614 | Corner | Empty membership user: limit=10 path short-circuit | 200 `[]`; no Meili unscoped call; metrics scoped_empty | API | P0 | empty scope | TODO |
| ECC-SRCH-C615 | Corner | Unicode poll option (emoji ZWJ) edit mid-reindex + search | one doc by id; highlight safe; no dup NFC/NFD privilege | API | P1 | unicode poll index | TODO |
| ECC-SRCH-C616 | Corner | Search hit deep-link to msg; channel invite-only and user not member | 403/404 open; no snippet beyond policy; WS subscribe also deny | API/E2E | P0 | search×invite gate | TODO |

---

## 5. Cross-domain combos (MSG × WS × NOTIF × SRCH)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MSG-C624 | Corner | Poll close + WS buffer-full + search still shows open question + vote notif in-flight | API closed; WS/resync closed; search eventual closed/absent; notif no further votes | Manual | P0 | poll cross-domain | TODO |
| ECC-WS-C617 | Corner | Invite accept → WS sub → mention_groups notif → search first msg | all four require membership; kick between any step fail-closes remainder | API/WS | P0 | invite pipeline | TODO |
| ECC-NOTIF-C617 | Corner | PWA offline queue send + WebPush TTL expire + SW cached search hits after kick | no send; no late push body; search UI cleared; re-auth | E2E | P0 | PWA cross-domain | TODO |
| ECC-SRCH-C617 | Corner | Multi-pod rate 429 on send + indexer lag + notif debounce + WS drop | 429 no msg; no phantom search/notif/WS; retry idempotent | Manual | P0 | rate×fanout | TODO |
| ECC-MSG-C625 | Corner | Voice upload + bookmark + star channel + GDPR erase peer mid | peer gone; voice/ACL clean; star/bookmark no PII; search anonymized | API | P0 | GDPR×voice×star | TODO |
| ECC-WS-C618 | Corner | Sticky break: notif deep-link opens UI; WS on other pod; search on third | each path independent ACL; all deny if kicked; no pod-local allow | Manual | P0 | multi-pod triple path | TODO |

---

## Счётчик

| Префикс | Диапазон ID | Count |
|---------|-------------|------:|
| ECC-MSG-C* | C600–C625 | **26** |
| ECC-WS-C* | C600–C618 | **19** |
| ECC-NOTIF-C* | C600–C617 | **18** |
| ECC-SRCH-C* | C600–C617 | **18** |
| **Итого Corner** | | **81** |

### Фокус-покрытие (wave4)

| Фокус | Пример ID | Кол-во (прибл.) |
|-------|-----------|----------------:|
| Poll vote/close/delete/kick | MSG-C600–C603, C618, C624; NOTIF-C606; SRCH-C605, C615 | ≥9 |
| Invite accept race / TTL / revoke | MSG-C604–C608; WS-C605, C616; NOTIF-C607; SRCH-C606, C616; WS-C617 | ≥10 |
| Bookmark / star × ACL | MSG-C609–C611, C623; WS-C612; NOTIF-C608; SRCH-C607 | ≥7 |
| Voice × storage / SSRF / kick | MSG-C612–C613; WS-C607; NOTIF-C609; SRCH-C608; MSG-C625 | ≥6 |
| Member batch / mention_groups | MSG-C614–C617; WS-C615; NOTIF-C610; SRCH-C609 | ≥7 |
| PWA / SW offline replay | MSG-C619–C620; WS-C603–C604; NOTIF-C612; SRCH-C611; NOTIF-C617 | ≥8 |
| Multi-pod rate / sticky / lease | MSG-C621–C622; WS-C602, C611, C614; NOTIF-C611; SRCH-C610; WS-C618 | ≥9 |
| WebPush TTL × quiet / tgLink TTL | NOTIF-C600–C605, C613–C614, C616 | ≥9 |
| Search shard-limit / saved-filter | SRCH-C600–C604, C613–C614 | ≥7 |
| Cross-domain pipelines | MSG-C624–C625; WS-C617–C618; NOTIF-C617; SRCH-C617 | ≥6 |

> Минимум 55 — **выполнено (81)**.  
> Happy / Neg / Edge — **не включены**.  
> Parent `C200+` / wave1 `C300+` / wave2 `C400+` / wave3 `C500+` — **не дублированы по ID и сценариям**.  
> `EDGE-CORNER-CATALOG.md` — **не изменялся**.  
> Коммит — **не создавался**.

---

*qa-manual · 2026-07-10 · wave4 · handoff → `qa-manual-ver`*
