# Corner expansion C03 wave2 — Messenger / WS / Notification / Search

> **Тип:** только **Corner** (редкие комбинации, races, mid-flight, multi-step, cross-domain).  
> **ID ranges:** `ECC-MSG-C400+` · `ECC-WS-C400+` · `ECC-NOTIF-C400+` · `ECC-SRCH-C400+`  
> **Не дублирует** parent `ECC-*-C200…` и wave1 `ECC-*-C300…` — углубляет **новыми** комбинациями.  
> **Фокус:** NF-02+NF-03 combo · NATS split-brain · unread+delete race · search index+soft-delete · telegram unbind mid-push · WS subscribe after kick.  
> **Merge target:** `.planning/july-2026/EDGE-CORNER-CATALOG.md` § Corner expansion (после ревью).  
> **Владелец:** `qa-manual` · 2026-07-10 · **не коммитить в каталог** без ревью.

Формат: `| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |`

---

## 1. Messenger — `ECC-MSG-C400+` (unread×delete · NATS split-brain · membership)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MSG-C400 | Corner | A sends M1 → B soft-deletes M1 **в том же ms**, что MarkRead(channel) на B | unread не учитывает deleted; badge ≥0; GetUnread согласован с history без M1 | API | P0 | unread+delete | TODO |
| ECC-MSG-C401 | Corner | Unread=3 (M1–M3) → B deletes M2 → MarkRead до last_seen=M3 | unread=0; deleted M2 не «возвращает» badge; cursor last_read не указывает на gone | API | P0 | unread+delete | TODO |
| ECC-MSG-C402 | Corner | Flood: 20 msgs → B MarkRead(mid-cursor) **пока** soft-delete batch на первые 10 | unread = только live msgs после cursor; нет negative unread | API | P0 | unread+delete | TODO |
| ECC-MSG-C403 | Corner | Sender A soft-deletes own msg **после** IncrementUnread на peers, **до** WS fanout | peers: либо no event, либо delete tombstone; unread rollback или never-increment; нет stuck +1 | API/WS | P0 | unread+delete × NATS | TODO |
| ECC-MSG-C404 | Corner | Two devices B1/B2: B1 deletes msg, B2 MarkRead overlapping Redis unread key | final unread = server truth; no lost decrement / double-decrement below 0 | API | P0 | unread+delete | TODO |
| ECC-MSG-C405 | Corner | Channel unread map: delete last unread msg in channel X while GetUnread in-flight | response snapshot consistent (0 or 1) — не half-map; retry converges | API | P1 | unread+delete | TODO |
| ECC-MSG-C406 | Corner | NATS **split-brain**: JetStream cluster A publishes msg event, cluster B consumer lag + duplicate deliver after heal | idempotent consumer; Mongo one msg; WS fanout once per sub (dedupe by msg_id) | Manual | P0 | NATS split-brain | TODO |
| ECC-MSG-C407 | Corner | Split-brain: two publishers (API replicas) both claim outbox lease for same idempotency-key during partition | after heal — one Mongo doc; second publish no-op; audit single create | Manual | P0 | NATS split-brain, idempotent | TODO |
| ECC-MSG-C408 | Corner | Split-brain: consumer on side A acks, side B redelivers same NATS msg_id after merge | handler idempotent; unread +1 once; reaction counts not double | API | P0 | NATS split-brain | TODO |
| ECC-MSG-C409 | Corner | Split-brain mid-thread: parent indexed on A, reply published only on B; search/history after heal | parent+reply consistent; reply_count exact; no orphan reply without parent visibility | Manual | P1 | NATS split-brain, thread | TODO |
| ECC-MSG-C410 | Corner | Kick member K mid-send: membership check passed on replica-1, revoke committed on replica-2 before Mongo write | fail-closed 403/404; no msg; K not in unread fanout | API | P0 | NF-02, membership | TODO |
| ECC-MSG-C411 | Corner | Kick K + concurrent MarkRead from K on same channel | MarkRead 403; unread key for K purged/ignored; no resurrect badge after re-invite without new msgs | API | P0 | unread × kick | TODO |
| ECC-MSG-C412 | Corner | Soft-delete channel while unread counters non-zero for N members | counters cleared/tombstoned; GetUnread omits channel; no leak of channel name in unread payload | API | P0 | unread+delete, channel | TODO |
| ECC-MSG-C413 | Corner | DM: A deletes last msg while B has unread=1 and opens channel | history empty/tombstone; unread→0; no phantom preview in channel list | API/E2E | P1 | unread+delete, DM | TODO |
| ECC-MSG-C414 | Corner | Reaction on msg + soft-delete msg + unread MarkRead race (three-way) | reaction gone; unread ignores deleted; no orphan reaction incrementing search facet | API | P1 | unread+delete × reaction | TODO |
| ECC-MSG-C415 | Corner | NATS split-brain: mute preference event on A, message event on B — order after heal | mute applied before/at delivery decision; no mute-bypass due to reorder (or documented eventual + security exception) | Manual | P1 | NATS split-brain, mute | TODO |
| ECC-MSG-C416 | Corner | Outbox retry storm during split-brain heal (100 pending) | backpressure; exactly-once WS per msg_id; API rate-limit still holds | Manual | P0 | NATS split-brain, NFR-07 | TODO |
| ECC-MSG-C417 | Corner | Forward msg F → target channel while source msg soft-deleted mid-forward | 404/gone or forward of last-known per policy; target never gets body if ACL revoked; audit | API | P1 | soft-delete, NF-03 | TODO |

---

## 2. WebSocket — `ECC-WS-C400+` (subscribe-after-kick · NF-02 · split-brain fanout)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-WS-C400 | Corner | Kick K → K **сразу** subscribe `channel:X` на **том же** open WS (без reconnect) | reject NF-02; zero events after kick; room hub без K | WS | P0 | WS subscribe after kick | TODO |
| ECC-WS-C401 | Corner | Kick K → K reconnect WS → subscribe `channel:X` with still-valid JWT | CanSubscribe false; 403/close policy; no silent join | WS | P0 | WS subscribe after kick | TODO |
| ECC-WS-C402 | Corner | Kick K → admin re-adds K → K subscribe **before** membership cache invalidate | fail-closed until cache sees membership **or** strong read; no window of leak | WS | P0 | WS subscribe after kick, rbac cache | TODO |
| ECC-WS-C403 | Corner | Kick mid-subscribe handshake: CanSubscribe true → revoke → JoinRoom commit | fail-closed; K not in fanout list; client error frame | WS | P0 | NF-02, subscribe after kick | TODO |
| ECC-WS-C404 | Corner | Multi-tab: tab1 kicked (force-unsub), tab2 still subscribed same user — then tab2 re-subscribe after server kick broadcast | all tabs unsub; re-subscribe reject until re-added | WS | P0 | WS subscribe after kick | TODO |
| ECC-WS-C405 | Corner | Kick + subscribe race on **two** rooms: revoke only X, K re-subscribes X and Y | X reject; Y OK if still member; no cross-room event bleed | WS | P0 | NF-02 | TODO |
| ECC-WS-C406 | Corner | Subscribe after kick using forged room id + old membership proof in client cache | server re-checks DB/RBAC; reject; client cache irrelevant | WS | P0 | NF-02 | TODO |
| ECC-WS-C407 | Corner | NATS split-brain: WS hub on node-1, consumer on node-2; kick event only on node-1 | after heal K unsubscribed everywhere; no zombie sub on node-2 | Manual | P0 | NATS split-brain × NF-02 | TODO |
| ECC-WS-C408 | Corner | Split-brain: duplicate WS fanout of same msg_id to one connection | client/server dedupe; UI one bubble; unread +1 once | WS/E2E | P0 | NATS split-brain | TODO |
| ECC-WS-C409 | Corner | Kick K while K has in-flight typing + presence in channel | typing/presence cleared; further frames dropped; no presence ghost | WS | P1 | subscribe after kick, presence | TODO |
| ECC-WS-C410 | Corner | Subscribe `channel:X` after kick, then search hits for X (NF-02+NF-03) | WS reject **and** search empty for X; combo must both fail-closed | API/WS | P0 | NF-02+NF-03 combo | TODO |
| ECC-WS-C411 | Corner | Admin kicks K; K's pending WS subscribe queue drained after authz flip | queued subscribe not applied; metrics: denied_after_kick | WS | P1 | WS subscribe after kick | TODO |
| ECC-WS-C412 | Corner | Kick → soft-delete channel → K subscribe + REST history | both 404/410; no distinction that leaks existence beyond policy | WS/API | P1 | NF-02, channel delete | TODO |
| ECC-WS-C413 | Corner | Subscribe after kick on `user:K` personal room still allowed; `channel:X` denied | personal OK; channel deny; isolation intact | WS | P1 | NF-02 | TODO |
| ECC-WS-C414 | Corner | Rapid kick/re-add/kick while K spams subscribe (50/s) | always matches latest membership; no intermittent allow when kicked | WS | P0 | WS subscribe after kick | TODO |
| ECC-WS-C415 | Corner | Split-brain heal: ordered catch-up includes kick then message | kicked clients must not receive post-kick msgs even if msg event arrives first on wire | Manual | P0 | NATS split-brain, NF-02 | TODO |

---

## 3. Notification — `ECC-NOTIF-C400+` (telegram unbind mid-push · races)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-NOTIF-C400 | Corner | TG bound → enqueue push → **unbind** after dequeue, **before** Bot API HTTP | no TG send (or Bot ignores); fallback in-app/mail per matrix; job complete not crash-loop | API | P0 | telegram unbind mid-push | TODO |
| ECC-NOTIF-C401 | Corner | Unbind mid-push: Bot API already in-flight (TCP sent) then unbind | at-most-one TG delivery OK; subsequent events no TG; bind audit shows unbound | API | P0 | telegram unbind mid-push | TODO |
| ECC-NOTIF-C402 | Corner | Unbind + rebind **different** chat_id while worker holds old chat_id | old chat_id must not receive; re-validate bind row at send; fail-closed to in-app | API | P0 | telegram unbind mid-push | TODO |
| ECC-NOTIF-C403 | Corner | Batch of 10 notifs: unbind between item 3 and 4 | 1–3 may TG; 4–10 fallback; no panic; no send to unbound | API | P0 | telegram unbind mid-push | TODO |
| ECC-NOTIF-C404 | Corner | Unbind mid-push + security-class event in same batch | security still delivered (mail/in-app); TG skipped if unbound; mute irrelevant | API | P0 | telegram unbind × security | TODO |
| ECC-NOTIF-C405 | Corner | Two workers: worker-A sends TG, worker-B processes unbind concurrently (split lease) | after heal — preference unbound; no duplicate storm; idempotent delivery keys | Manual | P0 | telegram unbind, NATS split-brain | TODO |
| ECC-NOTIF-C406 | Corner | Unbind during WebPush+TG multi-channel dispatch | WebPush continues if valid; TG aborted; in-app row created once | API | P1 | telegram unbind mid-push | TODO |
| ECC-NOTIF-C407 | Corner | Mention notif enqueued → kick from channel → unbind TG before delivery | no TG; in-app suppressed or deep-link 403; **no** content body in any channel | API | P0 | NF-02 × telegram unbind | TODO |
| ECC-NOTIF-C408 | Corner | Unbind mid-push while user MarkRead-all notifications | badge consistent; undelivered TG not reappearing as unread after fallback | API | P1 | telegram unbind, read-all | TODO |
| ECC-NOTIF-C409 | Corner | TG 429/5xx from Bot API **same second** as unbind | retry policy stops on unbound; no infinite retry; circuit doesn't poison other users | API | P1 | telegram unbind mid-push | TODO |
| ECC-NOTIF-C410 | Corner | Quiet hours + unbind mid-deferred-queue flush | deferred job sees unbound; delivers in-app only when quiet ends; no TG | API | P1 | telegram unbind, quiet hours | TODO |
| ECC-NOTIF-C411 | Corner | Unbind mid-push + soft-deleted source message (mention target gone) | notif omitted or tombstone; no TG with deleted snippet; search deep-link empty | API | P0 | telegram unbind × soft-delete | TODO |
| ECC-NOTIF-C412 | Corner | Preference flip TG-off concurrent with unbind mid-push | single terminal state unbound+off; one fallback path; audit coherent | API | P1 | telegram unbind | TODO |
| ECC-NOTIF-C413 | Corner | Call invite urgent push: unbind between LiveKit webhook and notif worker | invite still in-app/WS; no TG; call join ACL unchanged | API | P1 | telegram unbind, calls | TODO |

---

## 4. Search — `ECC-SRCH-C400+` (index×soft-delete · NF-03 · NF-02 combo)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-SRCH-C400 | Corner | Index msg → soft-delete → Meili still has doc → search | **no hit** (filter deleted + ACL); snippet not returned; NF-03 | API | P0 | search index+soft-delete | TODO |
| ECC-SRCH-C401 | Corner | Soft-delete → restore from trash → index lag | until reindex: miss OK; after: hit only for members; never for non-members | API | P0 | search index+soft-delete, NFR-11 | TODO |
| ECC-SRCH-C402 | Corner | Soft-delete mid-reindex: delete arrives while indexer upserts same id | final state absent/deleted in Meili; search no hit; job idempotent | API | P0 | search index+soft-delete | TODO |
| ECC-SRCH-C403 | Corner | Bulk soft-delete 100 msgs + concurrent Ctrl+K query containing keywords | zero deleted hits; live hits scoped; no timeout→unscoped fallback | API | P0 | search index+soft-delete, NF-03 | TODO |
| ECC-SRCH-C404 | Corner | Soft-delete channel + stale Meili channel docs + messages | no channel/message hits; empty not error leak of titles | API | P0 | search index+soft-delete | TODO |
| ECC-SRCH-C405 | Corner | Edit then soft-delete before reindex of edit | neither old nor new secret text in search; fail-closed | API | P0 | search index+soft-delete | TODO |
| ECC-SRCH-C406 | Corner | NF-02+NF-03: kick K → immediate search for private channel keywords + WS subscribe attempt | search empty **and** WS reject; dual invariant | API/WS | P0 | NF-02+NF-03 combo | TODO |
| ECC-SRCH-C407 | Corner | NF-02+NF-03: K still in Meili ACL cache as member after kick | query-time scope **re-check** membership; stale index ACL ignored | API | P0 | NF-02+NF-03 combo | TODO |
| ECC-SRCH-C408 | Corner | NF-02+NF-03: message event searchable by K while kick in-flight on other replica | after commit kick — no hits; during race fail-closed preferred | API | P0 | NF-02+NF-03, split-brain | TODO |
| ECC-SRCH-C409 | Corner | Soft-delete issue comment indexed as message-like doc | search excludes; issue body still scoped by project membership | API | P1 | search index+soft-delete | TODO |
| ECC-SRCH-C410 | Corner | GDPR soft-delete user display_name in msgs; Meili stale | purge/reindex; no PII hit; messages anonymized per policy | API | P0 | search index+soft-delete, US-S03 | TODO |
| ECC-SRCH-C411 | Corner | Index lag: soft-delete not yet in Meili delete queue; search with highlight | highlight path also filters deleted; no leaked fragment | API | P0 | search index+soft-delete | TODO |
| ECC-SRCH-C412 | Corner | Saved search / recent query re-run after soft-delete of all prior hits | 200 empty; no cached hit set bypassing ACL/delete | API | P1 | NF-03 | TODO |
| ECC-SRCH-C413 | Corner | Multi-type search: message soft-deleted + issue live same keyword | only issue (if in scope); message absent; no type confusion | API | P1 | search index+soft-delete | TODO |
| ECC-SRCH-C414 | Corner | Meili returns hit → Hub ACL denies (kicked) → response building | drop hit; never partial title/key in error or `data` | API | P0 | NF-02+NF-03 combo | TODO |
| ECC-SRCH-C415 | Corner | Soft-delete + unread badge channel + search same keyword | unread recalc without deleted; search no hit; badge/search agree on absence | API | P1 | unread+delete × search | TODO |

---

## 5. Cross-domain combos (MSG × WS × NOTIF × SRCH)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MSG-C418 | Corner | NATS split-brain heal: duplicate msg → unread+2 risk + WS double + search two docs | dedupe key wins; unread+1; one WS; one Meili doc | Manual | P0 | NATS split-brain × unread | TODO |
| ECC-WS-C416 | Corner | Kick → subscribe reject → TG mention push still queued for K | notif suppressed/403 deep-link; no WS; search empty (NF-02+NF-03+unbind path) | API/WS | P0 | NF-02+NF-03 × telegram | TODO |
| ECC-NOTIF-C414 | Corner | Unbind mid-push + soft-delete mentioned msg + kick in one second | no TG content; in-app gone/403; search miss; WS unsub | Manual | P0 | telegram unbind × soft-delete × kick | TODO |
| ECC-SRCH-C416 | Corner | Split-brain: indexer on A deletes doc, B re-upserts from old outbox after heal | compaction/version wins deleted; search no hit | Manual | P0 | NATS split-brain × index+soft-delete | TODO |
| ECC-MSG-C419 | Corner | Unread+delete race on M + K kicked + K searches M text + K resubscribes WS | unread peers OK; K: search empty, WS deny, no badge for channel | API/WS | P0 | NF-02+NF-03 × unread+delete | TODO |
| ECC-WS-C417 | Corner | Subscribe after kick during NATS partition (kick event delayed) | until kick applied: least privilege — prefer deny if membership read fails; after — deny | Manual | P0 | WS subscribe after kick × split-brain | TODO |
| ECC-NOTIF-C415 | Corner | Security mail + TG unbind mid-push + search for session-id string in msgs | security delivered; TG stop; search only scoped hits (no session secrets indexed) | API | P1 | telegram unbind × NF-03 | TODO |
| ECC-SRCH-C417 | Corner | Channel delete mid-WS + mid-reindex + unread flush + notif fanout | all four stop cleanly; no zombie push/search/WS/unread | Manual | P0 | channel delete cross-domain | TODO |
| ECC-MSG-C420 | Corner | Idempotent client retry during split-brain + kick of recipient mid-retry | one msg max; kicked recipient never gets WS/notif/search; sender may see own msg | API | P0 | NATS split-brain × NF-02 | TODO |
| ECC-WS-C418 | Corner | After kick: client uses search hit deep-link to open channel + WS subscribe | REST 403/404; WS reject; no message body in any error | E2E | P0 | NF-02+NF-03 combo | TODO |

---

## Счётчик

| Префикс | Диапазон ID | Count |
|---------|-------------|------:|
| ECC-MSG-C* | C400–C420 | **21** |
| ECC-WS-C* | C400–C418 | **19** |
| ECC-NOTIF-C* | C400–C415 | **16** |
| ECC-SRCH-C* | C400–C417 | **18** |
| **Итого Corner** | | **74** |

Минимум 55 — **выполнен** (74).

**Фокус-покрытие (wave2, не копируя C200/C300):**

| Фокус | ID (примеры) |
|-------|----------------|
| NF-02+NF-03 combo | WS-C410, SRCH-C406–C408, C414, MSG-C419, WS-C416, WS-C418 |
| NATS split-brain | MSG-C406–C409, C415–C416, WS-C407–C408, C415, NOTIF-C405, SRCH-C416, MSG-C418/C420 |
| unread+delete race | MSG-C400–C405, C412–C414, SRCH-C415, MSG-C419 |
| search index+soft-delete | SRCH-C400–C405, C409–C413, C416–C417 |
| telegram unbind mid-push | NOTIF-C400–C413, C414–C415 |
| WS subscribe after kick | WS-C400–C406, C409, C411–C414, C417 |

**Связь с prior waves (углубление, не дубль):**

| Prior | Wave2 углубление |
|-------|------------------|
| ECC-*-C200… | split-brain / dual-invariant NF-02+NF-03 / mid-unbind batch |
| ECC-*-C300… | subscribe-after-kick multi-tab+cache, unread×delete multi-device, Meili delete-vs-upsert race |

---

*qa-manual · C03 wave2 · 2026-07-10 · каталог не тронут · коммит не создавался*
