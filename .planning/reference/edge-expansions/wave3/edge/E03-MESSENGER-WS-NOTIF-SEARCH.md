# Edge expansions wave3 — Messenger / WS / Notification / Search

> **Тип:** только **Edge** (границы: poll options, invite max_uses/TTL, voice duration, bookmark/DM/members, WS send-buffer=256, JetStream MaxDeliver/AckWait/PullMax, notif pagination, tgLinkCodeTTL, push TTL, search shard limit=10 / index batch=5000).  
> **Источник:** NFR/TZ + код (`messenger/dto.go` poll options 2…10 / question max=500, `service_channel_invites.go` inviteTokenLength=22 / expires_in 1h|1d|7d|30d, `ws/hub.go` send chan=256 / register=64 / subscribe=256, `notification/consumer.go` MaxDeliver=5 / AckWait=30s / PullMaxMessages=10, `telegram_bind.go` tgLinkCodeTTL=5m, `push.go` TTL=60s, `handler.go` per_page 1…100, `search/service.go` defaultSearchLimit=10, `meilisearch_indexer.go` Limit(5000) / LimitReader 4096).  
> **Не дублирует** каталог `ECC-MSG-001…040` / `ECC-NSW-001…038`, wave1 `ECC-*-100…`, wave2 `ECC-*-200…`. Sibling-ссылки допустимы.  
> **ID:** Messenger `ECC-MSG-300+`, WS `ECC-WS-300+`, Notification `ECC-NOTIF-300+`, Search `ECC-SRCH-300+`.  
> **Gap:** только `TODO` | `Future` | `Manual`.  
> **Владелец:** `qa-manual` · 2026-07-10 · каталог не трогать · не коммитить.

Формат: `| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |`

---

## 1. Messenger — `ECC-MSG-300+`

> Фокус: **polls**, **channel invites**, **voice**, **forward/bookmarks/DM/members** — границы, не покрытые *100/*200 (bytes/thread/presence/empty-list).

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MSG-300 | Edge | Poll `options` count: 1 / 2 / 10 / 11 | 1→400; 2…10 OK; 11→400 (`validate:min=2,max=10`) | API | P0 | CreatePollRequest | TODO |
| ECC-MSG-301 | Edge | Poll `question` length: 0 / 1 / 500 / 501 | 0→400; 1…500 OK; 501→400 | API | P0 | question max=500 | TODO |
| ECC-MSG-302 | Edge | Poll option text: empty string / whitespace-only / very long option | empty/whitespace→400; long→400 или truncate policy documented | API | P1 | CreatePollRequest.Options | TODO |
| ECC-MSG-303 | Edge | Poll option texts: duplicate identical strings | accept+dedup **или** 400; counts не двоятся | API | P1 | poll options | TODO |
| ECC-MSG-304 | Edge | Vote: `option_ids` empty / 1 / all options / unknown UUID | empty→400; valid OK; unknown→400/404 | API | P0 | VotePollRequest min=1 | TODO |
| ECC-MSG-305 | Edge | Vote multi-choice OFF: 2 option_ids в одном запросе | 400; single-choice enforced | API | P0 | IsMultiChoice=false | TODO |
| ECC-MSG-306 | Edge | Vote multi-choice ON: 1 / N=all / N+1 (dup ids) | 1…all OK; dups dedup; closed poll→400 | API | P1 | IsMultiChoice=true | TODO |
| ECC-MSG-307 | Edge | Anonymous poll: after votes `voter_ids` empty; `voted` flag for self | no PII leak; `voted` accurate | API | P0 | PollOptionResponse | TODO |
| ECC-MSG-308 | Edge | Closed poll: vote / re-open attempt | vote→400; close idempotent | API | P1 | IsClosed | TODO |
| ECC-MSG-309 | Edge | Invite `max_uses`: 0 (unlimited) / 1 / N / use_count=N / N+1 join | 0 unlimited; at N last OK; N+1→reject exhausted | API | P0 | MaxUses, service_channel_invites | TODO |
| ECC-MSG-310 | Edge | Invite `max_uses` negative / huge int | 400 или clamp; no wrap | API | P1 | CreateInviteLinkRequest | TODO |
| ECC-MSG-311 | Edge | Invite `expires_in`: `""` / `1h` / `1d` / `7d` / `30d` / unknown | empty→nil ExpiresAt; known set; unknown→400 | API | P0 | parseExpiresIn | TODO |
| ECC-MSG-312 | Edge | Invite expiry boundary: join at `ExpiresAt−1s` vs `+1s` | −1s OK; +1s reject expired | API | P0 | ExpiresAt | TODO |
| ECC-MSG-313 | Edge | Invite token length: generated =22 base62; redeem with 21 / 22 / 23 chars | 22 path; wrong length→invalid | API | P1 | inviteTokenLength=22 | TODO |
| ECC-MSG-314 | Edge | Invite revoked + still within max_uses/TTL | redeem→invalid; list shows IsRevoked | API | P0 | IsRevoked | TODO |
| ECC-MSG-315 | Edge | Concurrent redeem: 2 clients race on `max_uses=1` | ровно 1 success; 1 reject; UseCount≤1 | API | P0 | atomic use_count | TODO |
| ECC-MSG-316 | Edge | Voice: `voice_duration_ms` =0 / 1 / practical max / negative | 0/neg→400; 1…max OK; overflow→400 | API | P1 | VoiceDurationMs | TODO |
| ECC-MSG-317 | Edge | Voice: `voice_url` empty + duration>0 / URL without duration | 400 inconsistent; both set OK | API | P1 | SendMessageRequest voice | TODO |
| ECC-MSG-318 | Edge | Forward: target = same channel / soft-deleted channel / non-member target | same policy documented; deleted/non-member→403/404 | API | P0 | ForwardMessageRequest | TODO |
| ECC-MSG-319 | Edge | Forward chain: forward of already-forwarded (depth 1 / 2 / N) | `forwarded_from_id` points to original **или** immediate; no infinite meta | API | P1 | ForwardedFromID | TODO |
| ECC-MSG-320 | Edge | Bookmark: empty list / add 1 / duplicate same message / remove last → empty | `[]`; dup idempotent or 409; empty after remove | API | P1 | AddBookmarkRequest | TODO |
| ECC-MSG-321 | Edge | Bookmark: message soft-deleted / other-channel message_id | 404/400; no orphan bookmark | API | P1 | bookmarks | TODO |
| ECC-MSG-322 | Edge | DM create: self `target_user_id` / unknown UUID / existing DM pair | self→400; unknown→404; existing→same channel id (idempotent) | API | P0 | CreateDMRequest | TODO |
| ECC-MSG-323 | Edge | Add members: `member_ids` empty / 1 / large N / duplicate UUIDs | empty→400; dups dedup; large N bounded/timeout | API | P1 | AddMembersRequest min=1 | TODO |
| ECC-MSG-324 | Edge | Readonly channel: member send vs admin/system | member→403; elevated OK or documented | API | P0 | IsReadonly | TODO |
| ECC-MSG-325 | Edge | Soft-delete message: list hides; trash restore within 30d−1s vs 30d+1s | −1s restore OK; +1s hard-gone / GDPR path | API | P1 | soft delete 30d TZ | TODO |
| ECC-MSG-326 | Edge | `mention_groups`: empty / 1 / unknown group / all-channel `@channel` | empty OK; unknown→400; @channel fanout bounded | API | P1 | MentionGroups | TODO |
| ECC-MSG-327 | Edge | Bot outbound webhook timeout: response at `2m−1s` vs `2m+1s` | −1s OK; +1s cancel/timeout; no hung goroutine | Unit | P2 | service_bot 2m | TODO |
| ECC-MSG-328 | Edge | Promote-to-issue `title`: 0 / 1 / 500 / 501 | 0→400; ≤500 OK; 501→400 | API | P2 | service_promote max=500 | TODO |

---

## 2. WS — `ECC-WS-300+`

> Фокус: **send buffer=256**, **hub channel capacities**, **multi-tab ConnectionID**, **pingPeriod=54s** — не повтор frame-size/pongWait из *100/*200.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-WS-300 | Edge | Client `send` buffer fill: 255 / 256 / 257 queued outbound | ≤256 enqueue; 257-й drop+warn; no panic/deadlock | Unit | P0 | send chan=256, hub.go | TODO |
| ECC-WS-301 | Edge | Buffer exactly full then client drains 1 → next enqueue | after drain accept resumes; no stuck drop mode | Unit | P0 | send buffer | TODO |
| ECC-WS-302 | Edge | Multi-tab: same userID, N=1 / 2 / 10 ConnectionIDs | SendToUser fanout ×N; unregister one tab others alive | Unit | P0 | HI-07 userClients | TODO |
| ECC-WS-303 | Edge | Multi-tab: one tab buffer full, sibling tabs healthy | drop only slow tab; siblings receive | Unit | P1 | per-client send | TODO |
| ECC-WS-304 | Edge | Hub `register` chan capacity 64: burst register 64 / 65 | ≤64 queued; 65 non-block policy (drop/backpressure) documented; hub Run alive | Unit | P1 | register chan=64 | TODO |
| ECC-WS-305 | Edge | Hub `subscribe`/`unsubscribe` chan=256: burst 256 / 257 | no deadlock; final room map consistent | Unit | P1 | subscribe chan=256 | TODO |
| ECC-WS-306 | Edge | `pingPeriod`=(pongWait×9)/10 =54s: ping at 54s−1s vs schedule drift | ping fires ~54s; conn stays within pongWait | Unit | P1 | pingPeriod | TODO |
| ECC-WS-307 | Edge | Hub `stopping`: BroadcastToRoom / SendToUser during shutdown | no send-on-closed; clients drain via `done` | Unit | P0 | SEC-05 done channel | TODO |
| ECC-WS-308 | Edge | Unregister then late Subscribe race (fast reconnect) | serialized in Run; no ghost room on dead client | Unit | P0 | HI-07 subscribe after unregister | TODO |
| ECC-WS-309 | Edge | SessionID empty / mismatch on upgrade vs JWT session | reject/close; no orphan Client | API | P1 | NewClient session | TODO |
| ECC-WS-310 | Edge | Room fanout cardinality: 1 / 100 / 1000 subscribers one event | all get or bounded drop; latency OK; hub not blocked | Perf | P1 | BroadcastToRoom | TODO |
| ECC-WS-311 | Edge | Dual producer: BroadcastToRoom + SendToUser same client concurrently | no panic; ordering best-effort; buffer accounting correct | Unit | P1 | hub writers | TODO |
| ECC-WS-312 | Edge | ConnectionID uniqueness: 100 parallel NewClient same user | 100 distinct UUIDs; map size=100 | Unit | P2 | ConnectionID uuid | TODO |
| ECC-WS-313 | Edge | Outbound JSON near send-buffer message count (many tiny events) vs few huge | drop policy by count not only bytes; warn logged | Perf | P2 | buffer full | TODO |

---

## 3. Notification — `ECC-NOTIF-300+`

> Фокус: **JetStream MaxDeliver/AckWait/PullMax**, **list pagination**, **tgLinkCodeTTL=5m**, **Web Push TTL=60s** — не повтор digest-size/debounce-window из *200.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-NOTIF-300 | Edge | JetStream redelivery: fail processing 4× vs 5× (`MaxDeliver=5`) | 4→retry; 5-й last attempt then poison/skip; no infinite loop | Unit | P0 | MaxDeliver=5 | TODO |
| ECC-NOTIF-301 | Edge | AckWait=30s: handler finishes at 30s−1s vs hangs past 30s+1s | −1s ACK ok; +1s redelivery (dup risk) → idempotent create | Unit | P0 | AckWait=30s | TODO |
| ECC-NOTIF-302 | Edge | PullMaxMessages=10: backlog 9 / 10 / 11 pending | pull batches ≤10; 11th in next pull; no starve | Unit | P1 | PullMaxMessages(10) | TODO |
| ECC-NOTIF-303 | Edge | List `per_page`: 0 / 1 / 20 (default) / 100 / 101 / negative | 0/neg/101→default 20; 1…100 OK | API | P0 | handler per_page clamp | TODO |
| ECC-NOTIF-304 | Edge | List `page`: 0 / 1 / last page / page>total | 0→treat as 1 or 400; last partial; overflow→`[]` total intact | API | P1 | ListByUser offset | TODO |
| ECC-NOTIF-305 | Edge | List filter `is_read`: null / true / false at empty & full inbox | correct subset; empty→`[]`; total matches filter | API | P1 | isRead filter | TODO |
| ECC-NOTIF-306 | Edge | Pagination + mark-one-read mid-list: page boundary shift | no phantom dup/skip across pages **или** documented cursorless tradeoff | API | P1 | offset pagination | TODO |
| ECC-NOTIF-307 | Edge | `tgLinkCodeTTL=5m`: confirm at 5m−1s vs 5m+1s | −1s OK; +1s expired (отдельно от tgBindTTL=10m) | API | P0 | tgLinkCodeTTL=5m | TODO |
| ECC-NOTIF-308 | Edge | Link-code reuse after success within TTL window | second confirm→invalid; Redis key gone | API | P0 | ConfirmLinkCode | TODO |
| ECC-NOTIF-309 | Edge | Web Push payload `TTL=60`: delivery deferred 60s−1 vs 60s+ | −1 still attempt; + beyond TTL drop per browser/push svc | Manual | P2 | push.go TTL=60 | TODO |
| ECC-NOTIF-310 | Edge | Quiet hours overnight wrap: start=22:00 end=06:00 at 21:59 / 22:00 / 05:59 / 06:00 | in-window defer; out deliver; urgent bypass | Unit | P0 | isQuietHours wrap | TODO |
| ECC-NOTIF-311 | Edge | Quiet hours unset (nil start/end) vs only one set | both nil→not quiet; partial→documented fail-safe | Unit | P1 | QuietHoursStart/End | TODO |
| ECC-NOTIF-312 | Edge | NATS client nil at Consumer.Start | warn + no-op start; process stays up | Unit | P1 | NewConsumer nil nc | TODO |
| ECC-NOTIF-313 | Edge | Durable consumer recreate after crash mid-AckWait | at-least-once; duplicate notif suppressed by debounce key | Unit | P1 | notification-dispatcher | TODO |
| ECC-NOTIF-314 | Edge | Unread-count vs list total after partial page mark-read | count == unread rows; never negative | API | P0 | unread-count | TODO |
| ECC-NOTIF-315 | Edge | Preference project override: empty project_id vs UUID / deleted project | global vs project key; deleted→fallback global | API | P2 | IsEnabled project | TODO |

---

## 4. Search — `ECC-SRCH-300+`

> Фокус: **per-shard defaultSearchLimit=10**, **index batch 5000**, **Meili error body 4096**, **saved-filter name=200** — не повтор typoTolerance / q=500 из *200.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-SRCH-300 | Edge | Per-type shard hits: exactly 10 / 11 matching docs (`defaultSearchLimit=10`) | ≤10 per shard returned; 11th truncated; merged list respects API `limit` | API | P0 | defaultSearchLimit=10 | TODO |
| ECC-SRCH-301 | Edge | API `limit=50` but each shard caps at 10 — multi-type query | total ≤ min(50, 10×types); no silent inflate beyond shard cap | API | P0 | SearchRequest.Limit vs shard | TODO |
| ECC-SRCH-302 | Edge | API `limit=1` with many strong hits across types | exactly 1 in response; ranking picks best | API | P1 | limit clamp | TODO |
| ECC-SRCH-303 | Edge | IndexAllIssues batch: org with 4999 / 5000 / 5001 issues | 5000 Limit path; 5001 needs multi-page **или** documented truncation | Unit | P1 | meilisearch_indexer Limit(5000) | TODO |
| ECC-SRCH-304 | Edge | Meili error response body: 4095 / 4096 / 4097 bytes | LimitReader 4096; no OOM; error logged truncated | Unit | P1 | LimitReader 4096 | TODO |
| ECC-SRCH-305 | Edge | Saved filter `name`: 0 / 1 / 200 / 201 | 0→400; ≤200 OK; 201→400 | API | P1 | dto Name max=200 | TODO |
| ECC-SRCH-306 | Edge | Saved filter `filter_config`: empty map / huge nested / missing | required; empty policy; huge→400 no OOM | API | P1 | FilterConfig | TODO |
| ECC-SRCH-307 | Edge | Saved filter `project_id`: empty / invalid UUID / foreign project | empty/invalid→400; foreign→403/404 (scope) | API | P0 | NF-03 edge | TODO |
| ECC-SRCH-308 | Edge | Types filter: empty array / single / all known / unknown+known mix | empty→all; unknown ignore or 400; known filtered | API | P1 | SearchRequest.Types | TODO |
| ECC-SRCH-309 | Edge | Ranking tie: identical scores across 2 types | stable order documented; deterministic under same data | API | P2 | rank DESC | TODO |
| ECC-SRCH-310 | Edge | Soft-deleted entity still in Meili until reindex | excluded by scope filter **или** stale until IndexAll; no ACL leak | API | P0 | NF-03 + soft delete | TODO |
| ECC-SRCH-311 | Edge | Parallel shard: 1 shard returns err, others OK | partial+log **или** fail-closed; never mix unauthorized hits | API | P0 | SearchService parallel | TODO |
| ECC-SRCH-312 | Edge | Message search (Mongo) vs global search (Meili/PG): same `q` length edge | both respect own limits; no cross-engine assumption | API | P1 | SearchMessages vs Search | TODO |
| ECC-SRCH-313 | Edge | Reindex while query in flight | no 500; eventual consistency; user-scope always applied | API | P1 | IndexAllIssues race | TODO |
| ECC-SRCH-314 | Edge | Ctrl+K empty result after deleting last matching entity | UI empty-state; API `[]`; no stale cache hit | E2E | P2 | US-M07 | Manual |

---

## Счётчик

| Домен | ID range | Кол-во Edge | P0 | P1 | P2 |
|-------|----------|-------------|----|----|-----|
| Messenger | ECC-MSG-300…328 | **29** | 12 | 13 | 4 |
| WS | ECC-WS-300…313 | **14** | 5 | 7 | 2 |
| Notification | ECC-NOTIF-300…315 | **16** | 7 | 7 | 2 |
| Search | ECC-SRCH-300…314 | **15** | 5 | 7 | 3 |
| **Итого** | | **74** | **29** | **34** | **11** |

> Минимум 50 — **выполнено (74)**.  
> Happy / Neg / Corner — **не включены**.  
> ID `*100+` / `*200+` / каталог `*001…` — **не пересекались** (новые оси: polls/invites/voice/DM, send-buffer 256, MaxDeliver/AckWait/per_page, shard limit 10 / batch 5000).  
> `.planning/archive/july-2026/EDGE-CORNER-CATALOG.md` — **не изменялся**.  
> Коммит — **не создавался**.
