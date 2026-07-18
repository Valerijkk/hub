# Edge expansions wave2 — Messenger / WS / Notification / Search

> **Тип:** только **Edge** (границы: max bytes, depth, payload, TTL, digest size, typo tolerance, empty list).  
> **Источник:** NFR/TZ + код (`messenger/dto.go`, `messenger/presence.go` away=5m / stale=90s / cleanup=60s, `ws/handler.go` maxMessageSize=4096 / pongWait=60s / writeWait=10s, `notification/consumer.go` debounce=2s / cleanDebounce=5s, `search/dto.go` q max=500, Meili default typoTolerance oneTypo≥5 / twoTypos≥9).  
> **Не дублирует** каталог `ECC-MSG-001…040` / `ECC-NSW-001…038` и wave1 `ECC-*-100…`. Sibling-ссылки на `*100+` допустимы.  
> **ID:** Messenger `ECC-MSG-200+`, WS `ECC-WS-200+`, Notification `ECC-NOTIF-200+`, Search `ECC-SRCH-200+`.  
> **Gap:** только `TODO` | `Future` | `Manual`.  
> **Владелец:** `qa-manual` · 2026-07-10 · каталог не трогать · не коммитить.

Формат: `| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |`

---

## 1. Messenger — `ECC-MSG-200+`

> Фокус: **message max bytes**, **thread depth**, **presence TTL**, **empty channel list**.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MSG-200 | Edge | `content_text` byte length: 0 / 1 / practical UI max / near Fiber BodyLimit−1B / BodyLimit+1B | 0→400 (empty); 1 OK; near-limit OK или 413; overflow→413/400; `request_id` | API | P0 | NFR-04 BodyLimit 25MB, ECC-MSG-100 sibling | TODO |
| ECC-MSG-201 | Edge | `content_text` UTF-8: N runes where each rune=4 bytes (emoji) — byte-budget vs rune-budget | reject/accept по документированной политике (bytes vs runes); нет mid-grapheme truncate | API | P0 | ECC-MSG-101 sibling | TODO |
| ECC-MSG-202 | Edge | `content_text` ровно на границе max (если введён soft-cap) vs max+1 rune | max OK; max+1→400 | API | P0 | ECC-MSG-012↑ | TODO |
| ECC-MSG-203 | Edge | `content_json` + `content_text` суммарный HTTP body: BodyLimit−1 vs BodyLimit+1 | −1 parse path; +1→413 до handler | API | P0 | Fiber BodyLimit | TODO |
| ECC-MSG-204 | Edge | Message body: только whitespace / NBSP / zero-width spaces as `content_text` | 400 empty-equivalent **или** accept+trim policy documented | API | P1 | ECC-MSG-100 sibling | TODO |
| ECC-MSG-205 | Edge | Message body: binary-looking / null `\u0000` / control chars in text | 400; no Mongo/WS corruption | API | P1 | ECC-MSG-101 sibling | TODO |
| ECC-MSG-206 | Edge | Attachments array: 0 / 1 / max practical / max+1 refs in one message | ≤max OK; overflow→400; body still under BodyLimit | API | P1 | ECC-MSG-031 sibling | TODO |
| ECC-MSG-207 | Edge | Thread depth=0: send без `thread_id` (root) | 201; `thread_id` absent; `reply_count=0` | API | P0 | TC-MSG-06 | TODO |
| ECC-MSG-208 | Edge | Thread depth=1: reply к root (`thread_id`=root msg id) | 201; parent `reply_count++`; list replies содержит msg | API | P0 | TC-MSG-06, ECC-MSG-122 sibling | TODO |
| ECC-MSG-209 | Edge | Thread depth=2: reply к reply (thread_id=child, не root) | flat-to-root **или** 400 (политика 1-level); нет бесконечной вложенности | API | P0 | ECC-MSG-122 sibling | TODO |
| ECC-MSG-210 | Edge | Thread depth≥3: цепочка reply→reply→reply | reject/flatten; depth не растёт в UI/API | API | P1 | thread policy | TODO |
| ECC-MSG-211 | Edge | Thread: `thread_id` = self / non-existent UUID / soft-deleted parent | 400/404; нет orphan reply | API | P0 | ECC-MSG-021↑ | TODO |
| ECC-MSG-212 | Edge | Thread: `thread_id` из **другого** channel | 400/403; нет cross-channel thread attach | API | P0 | NF-03 edge | TODO |
| ECC-MSG-213 | Edge | Thread replies list: empty thread (0 replies) | `[]` 200; has_more=false | API | P1 | ECC-MSG-123 sibling | TODO |
| ECC-MSG-214 | Edge | Thread `reply_count`: 0→1→N; delete last reply →0 | counts exact; never negative | API | P0 | Increment/DecrementReplyCount | TODO |
| ECC-MSG-215 | Edge | Empty channel list: user с 0 memberships — `GET /channels` | `[]` 200; **не** null; UI empty-state | API | P0 | ListChannels | TODO |
| ECC-MSG-216 | Edge | Empty channel list + unread map / starred / muted filters | empty collections; no 500 on join of empty sets | API | P1 | ECC-MSG-116 sibling | TODO |
| ECC-MSG-217 | Edge | Empty channel list: new user после approve, до join/create | `[]`; create first channel → list length=1 | API | P1 | invite→approve | TODO |
| ECC-MSG-218 | Edge | Empty messages in existing channel (channel created, 0 msgs) | messages `[]` 200; cursor absent | API | P1 | ECC-MSG-106 sibling | TODO |
| ECC-MSG-219 | Edge | Presence TTL: heartbeat at `away−1s` (5m−1s) vs `away+1s` (5m+1s) | −1s → `online`; +1s → `away` (если запись ещё в ZSET) | Unit | P0 | presenceAwayThreshold=5m, TC-PRESENCE-02 | TODO |
| ECC-MSG-220 | Edge | Presence stale TTL: last heartbeat `90s−1s` vs `90s+1s` + CleanupStale | −1s остаётся в ZSET; +1s удалён → `offline` | Unit | P0 | presenceStaleThreshold=90s, TC-PRESENCE-03 | TODO |
| ECC-MSG-221 | Edge | Presence: cleanup interval boundary (60s ticker) — stale появляется между тиками | до тика может быть `away`/`online` stale; после CleanupStale → `offline` | Unit | P1 | presenceCleanupInterval=60s | TODO |
| ECC-MSG-222 | Edge | Presence interaction: stale=90s < away=5m — запись удалена до away | после 90s+cleanup GetStatus=`offline` (не `away`); documented | Unit | P0 | presence.go GetStatus vs CleanupStale | TODO |
| ECC-MSG-223 | Edge | Presence heartbeat flood: 1/ms vs 1/s | Redis ZADD ok; no CPU spin; status stable `online` | Perf | P2 | Heartbeat | TODO |
| ECC-MSG-224 | Edge | Presence custom status TTL: `30m` / `1h` / `4h` / `today` / `1d` / `1w` / `0` / unknown | parse per `ParseStatusDuration`; unknown→24h default; expiry clears | API | P1 | TC-PRESENCE-04 | TODO |
| ECC-MSG-225 | Edge | Presence custom status: empty text / max length / unicode | empty clears or 400; max OK; overflow 400 | API | P2 | SetCustomStatus | TODO |
| ECC-MSG-226 | Edge | Presence bulk: empty `user_ids` / 1 / large N | `{}` or `[]` policy; large N bounded/timeout | API | P1 | TC-PRESENCE-05 | TODO |
| ECC-MSG-227 | Edge | Message max bytes + rate limit: 60 near-max bodies / min | 60-й 201; 61-й 429; нет OOM | API | P0 | NFR-07 | TODO |
| ECC-MSG-228 | Edge | Edit message: new text max / max+1 / empty | max OK; overflow/empty→400; version/edited_at set | API | P1 | EditMessageRequest | TODO |

---

## 2. WS — `ECC-WS-200+`

> Фокус: **WS payload** (inbound/outbound size, framing, near-limit JSON).

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-WS-200 | Edge | Inbound frame size: 4095 / 4096 / 4097 bytes | ≤4096 accept path; 4097→close (SetReadLimit) | Unit | P0 | maxMessageSize=4096, ECC-WS-107 sibling | TODO |
| ECC-WS-201 | Edge | Inbound: empty frame / `{}` / minimal valid subscribe JSON | empty/invalid→ignore or close policy; minimal valid OK | API | P1 | ECC-WS-108 sibling | TODO |
| ECC-WS-202 | Edge | Inbound: subscribe JSON padded to ровно 4096 bytes (whitespace/junk fields) | parse OK if schema valid; else reject; no panic | Unit | P0 | protocol.go | TODO |
| ECC-WS-203 | Edge | Inbound: subscribe JSON 4096 with truncated UTF-8 tail | close/reject; no mojibake crash | Unit | P1 | maxMessageSize | TODO |
| ECC-WS-204 | Edge | Inbound: binary frame / mixed text+binary | reject/close; text-only protocol | Unit | P1 | Fiber WS | TODO |
| ECC-WS-205 | Edge | Outbound fanout: large `message_created` payload (long content_text) | deliver or truncate per policy; slow client → drop+warn, no deadlock | Unit | P0 | hub send buffer, ECC-WS-109 sibling | TODO |
| ECC-WS-206 | Edge | Outbound: burst N large events to one slow client | buffer full → drop; conn eventually close; hub alive | Perf | P0 | send buffer | TODO |
| ECC-WS-207 | Edge | Payload: room name at max length inside near-limit frame | reject invalid room; valid room under limit OK | API | P1 | CanSubscribe | TODO |
| ECC-WS-208 | Edge | Payload: nested/extra JSON fields до лимита | ignore unknown; required fields validated | API | P2 | ClientMessage | TODO |
| ECC-WS-209 | Edge | writeWait=10s: outbound write at 10s−1s vs 10s+1s | −1s OK; +1s timeout close | Unit | P1 | writeWait, ECC-WS-115 sibling | TODO |
| ECC-WS-210 | Edge | pongWait payload: pong frame empty vs oversized control | empty OK; oversized control → close | Unit | P1 | pongWait=60s | TODO |
| ECC-WS-211 | Edge | Ping/pong timing: pong at 60s−1s / 60s / 60s+1s | −1s/exact alive per impl; +1s read deadline close | Unit | P0 | ECC-WS-105 sibling | TODO |
| ECC-WS-212 | Edge | Multi-frame: client sends 2 frames totaling >4096 without being one message | each frame independently limited; no reassembly bypass | Unit | P0 | SetReadLimit | TODO |
| ECC-WS-213 | Edge | Subscribe storm with minimal payloads (100× tiny subscribe) | final room set consistent; no ghost entries | Unit | P1 | ECC-WS-103 sibling | TODO |
| ECC-WS-214 | Edge | Unsubscribe payload: missing room / empty room / valid | missing→400-equiv; empty reject; valid unsub | API | P1 | protocol | TODO |
| ECC-WS-215 | Edge | Presence/typing WS events: empty payload / max typing key length | ignore/reject; throttle flood | Manual | P2 | ECC-MSG-019↑ | TODO |

---

## 3. Notification — `ECC-NOTIF-200+`

> Фокус: **digest size** (debounce coalesce window + batch cardinality).

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-NOTIF-200 | Edge | Digest size=0: no events in window | no delivery; debounce map empty/noop | Unit | P1 | shouldDispatch | TODO |
| ECC-NOTIF-201 | Edge | Digest size=1: single event | passthrough 1 notification | Unit | P0 | ECC-NOTIF-102 sibling | TODO |
| ECC-NOTIF-202 | Edge | Digest size=N: 2 / 10 / 100 events same `user:type:entity` in <2s | coalesced → **1** delivery | Unit | P0 | debounce 2s, ECC-NSW-010↑ | TODO |
| ECC-NOTIF-203 | Edge | Digest window: 2-е событие at `2s−1ms` vs `2s+1ms` | −1ms coalesce; +1ms → 2-я notif | Unit | P0 | ECC-NOTIF-101 sibling | TODO |
| ECC-NOTIF-204 | Edge | Digest window exact `2s`: boundary inclusive/exclusive | documented; deterministic | Unit | P1 | `time.Since < 2s` | TODO |
| ECC-NOTIF-205 | Edge | Digest map growth: 10k distinct keys in burst | memory bounded / cleanDebounce; no OOM | Perf | P1 | cleanDebounce 5s | TODO |
| ECC-NOTIF-206 | Edge | cleanDebounce cutoff: entry age `5s−1ms` vs `5s+1ms` | −1ms kept; +1ms deleted from map | Unit | P1 | cleanDebounce | TODO |
| ECC-NOTIF-207 | Edge | Digest size across types: same entity, different `notifType` | separate keys → 2 deliveries | Unit | P1 | shouldDispatch key | TODO |
| ECC-NOTIF-208 | Edge | Digest size across users: same type+entity, 2 users | 2 deliveries (per-user key) | Unit | P1 | key=user:type:entity | TODO |
| ECC-NOTIF-209 | Edge | Digest + quiet hours: N coalesced non-urgent in quiet | 1 deferred digest; not N deferred | API | P1 | TC-NOTIF-05 | TODO |
| ECC-NOTIF-210 | Edge | Digest + urgent: N urgent same key in 2s | coalesce still 1 **или** urgent bypass documented | API | P0 | TC-NOTIF-06 | TODO |
| ECC-NOTIF-211 | Edge | In-app list after digest burst size=100 | inbox +1 (not +100); unread-count consistent | API | P0 | unread-count | TODO |
| ECC-NOTIF-212 | Edge | Empty notification list after mark-all on digest residue | `[]` / unread=0; 200 | API | P2 | ECC-NOTIF-109 sibling | TODO |
| ECC-NOTIF-213 | Edge | Telegram digest payload size: long title/body near TG limit | truncate/split policy; no 500 from bot API | Manual | P2 | Telegram delivery | TODO |
| ECC-NOTIF-214 | Edge | Web Push payload size near browser limit after digest | accept or reject gracefully; ValidateOutboundURL still applied | API | P1 | NF-07 | TODO |
| ECC-NOTIF-215 | Edge | Preference all-off + digest burst size=N | 0 deliveries any channel | API | P0 | ECC-NOTIF-103 sibling | TODO |

---

## 4. Search — `ECC-SRCH-200+`

> Фокус: **Meili typo tolerance boundary** (+ empty/limit edges complementary).

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-SRCH-200 | Edge | Meili typo: query word length 4 (below oneTypo min=5) with 1-char typo | **no** typo match (exact only); `[]` or exact hits only | API | P0 | Meili typoTolerance oneTypo≥5 | TODO |
| ECC-SRCH-201 | Edge | Meili typo: word length 5 with 1-char typo vs exact | 1-typo match enabled; ranked below exact | API | P0 | oneTypo boundary | TODO |
| ECC-SRCH-202 | Edge | Meili typo: word length 8 (still < twoTypos=9) with 2-char typo | **no** 2-typo match; 1-typo OK | API | P0 | twoTypos≥9 | TODO |
| ECC-SRCH-203 | Edge | Meili typo: word length 9 with 2-char typo | 2-typo match enabled | API | P0 | twoTypos boundary | TODO |
| ECC-SRCH-204 | Edge | Meili typo: word length 9 with 3-char typo | no match (beyond 2 typos) | API | P1 | typoTolerance max | TODO |
| ECC-SRCH-205 | Edge | Meili typo: disableOnWords / disableOnAttributes if configured | exact-only for those terms/attrs | API | P2 | Meili settings | TODO |
| ECC-SRCH-206 | Edge | Meili typo + user-scope: typo would match foreign-project only | `[]`; **no** leak (NF-03) | API | P0 | NF-03, ECC-SRCH-104 sibling | TODO |
| ECC-SRCH-207 | Edge | Meili typo on CJK / emoji / single-codepoint tokens | stable; no panic; typo rules may N/A | API | P1 | ECC-SRCH-102 sibling | TODO |
| ECC-SRCH-208 | Edge | Meili typo: multi-word query — one word below min, one above | partial: short exact, long allows typo | API | P1 | matchingStrategy | TODO |
| ECC-SRCH-209 | Edge | Meili down during typo query | fallback PG/Mongo or 503; no hang | API | P0 | ECC-SRCH-107 sibling | TODO |
| ECC-SRCH-210 | Edge | Empty search index + typo query | `[]` 200 | API | P2 | ECC-SRCH-109 sibling | TODO |
| ECC-SRCH-211 | Edge | Query length boundary with typo intent: 1 / 500 / 501 | 1..500 OK; 501→400 | API | P1 | validate max=500 | TODO |
| ECC-SRCH-212 | Edge | `limit` 1 / 50 / 51 with typo-expanded hit set | clamp ≤50; scores ordered | API | P1 | ECC-SRCH-106 sibling | TODO |
| ECC-SRCH-213 | Edge | Empty `q` / whitespace vs typo-looking `q="a"` | empty→400; `"a"` exact-only (len<5) | API | P1 | ECC-SRCH-100 sibling | TODO |
| ECC-SRCH-214 | Edge | Ctrl+K: typo query debounce — rapid 4→5 char upgrade | last query wins; no stale wider typo results stuck | E2E | P2 | US-M07 | Manual |
| ECC-SRCH-215 | Edge | In-channel message search: typo vs Mongo text index (no Meili) | Mongo behavior documented; no false Meili assumptions | API | P1 | TC-MSG-10, SearchMessages | TODO |

---

## Счётчик

| Домен | ID range | Кол-во Edge | P0 | P1 | P2 |
|-------|----------|-------------|----|----|-----|
| Messenger | ECC-MSG-200…228 | **29** | 14 | 12 | 3 |
| WS | ECC-WS-200…215 | **16** | 7 | 7 | 2 |
| Notification | ECC-NOTIF-200…215 | **16** | 5 | 8 | 3 |
| Search | ECC-SRCH-200…215 | **16** | 6 | 7 | 3 |
| **Итого** | | **77** | **32** | **34** | **11** |

> Минимум 50 — **выполнено (77)**.  
> Happy / Neg / Corner — **не включены**.  
> ID `*100+` / каталог `*001…` — **не пересекались**.  
> `.planning/july-2026/EDGE-CORNER-CATALOG.md` — **не изменялся**.  
> Коммит — **не создавался**.
