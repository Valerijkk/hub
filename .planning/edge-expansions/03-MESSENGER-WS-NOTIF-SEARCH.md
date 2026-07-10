# Edge expansions — Messenger / WS / Notification / Search

> **Тип:** только **Edge** (границы: лимиты, empty/end, max size, rate, TTL, overflow, timeout, unicode).  
> **Источник:** NFR/TZ (`docs/tz_constraints.md`, `docs/system_specification.md`) + код (`messenger/dto.go`, `middleware/ratelimit.go` MessageSendRateLimit=60/min, `ws/handler.go` maxMessageSize=4096 / pongWait=60s, `notification/telegram_bind.go` tgBindTTL=10m, `search/dto.go` q max=500).  
> **Не дублирует** `ECC-MSG-001…040` / `ECC-NSW-001…038` из `.planning/EDGE-CORNER-CATALOG.md` (sibling-ссылки допустимы).  
> **ID:** Messenger `ECC-MSG-100+`, WS `ECC-WS-100+`, Notification `ECC-NOTIF-100+`, Search `ECC-SRCH-100+`.  
> **Gap:** только `TODO` | `Future` | `Manual`.  
> **Владелец:** `qa-manual` · 2026-07-10 · не коммитить в каталог без ревью.

---

## 1. Messenger — `ECC-MSG-100+`

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MSG-100 | Edge | `content_text` length: empty / 1 rune / practical max / max+1 | empty→400; 1 OK; max OK; overflow→400 | API | P0 | ECC-MSG-012 sibling, NFR | TODO |
| ECC-MSG-101 | Edge | `content_text` unicode: ZWJ emoji sequence / RTL / combining marks / null byte | stored by rune policy; null→400; no panic/truncation mid-grapheme | API | P1 | ECC-MSG-028 sibling | TODO |
| ECC-MSG-102 | Edge | `content_json` huge nested / deeply nested object near body limit | 400 or reject; no OOM; request_id в ошибке | API | P1 | Fiber BodyLimit, NFR-13 | TODO |
| ECC-MSG-103 | Edge | Message rate: 60-й send в окне 1m vs 61-й | 60-й 201; 61-й **429** | API | P0 | NFR-07, MessageSendRateLimit, ECC-MSG-015 | TODO |
| ECC-MSG-104 | Edge | Message rate window boundary: 61-й сразу после expiry окна | снова 201; счётчик сброшен | API | P0 | NFR-07 | TODO |
| ECC-MSG-105 | Edge | Message rate: burst 60 параллельных send | ≤60 success; избыток 429; нет silent drop без 429 | API | P0 | NFR-07 | TODO |
| ECC-MSG-106 | Edge | Cursor pagination: пустой канал (`before` absent, limit=50) | `[]` 200; нет next cursor / has_more=false | API | P1 | NFR-15, ECC-MSG-016 | TODO |
| ECC-MSG-107 | Edge | Cursor pagination: last page (ровно `limit` осталось / `limit−1` / 0) | last page без next; end-of-stream стабилен | API | P1 | NFR-15 | TODO |
| ECC-MSG-108 | Edge | Cursor `before`: epoch / far-future / malformed / non-UTC | valid→slice; malformed→400; far-future→empty или full first page policy | API | P1 | NFR-15 | TODO |
| ECC-MSG-109 | Edge | List messages `limit`: 1 / 50 (default) / 100 (mongo clamp) / 101 / 0 / negative | 1..100 OK (clamp); 0/neg→default 50; 101→clamp 50 | API | P1 | mongo_repository limit≤100 | TODO |
| ECC-MSG-110 | Edge | Reactions: emoji length 1 / 32 (max) / 33 | ≤32 OK; 33→400 (`validate:max=32`) | API | P1 | ToggleReactionRequest | TODO |
| ECC-MSG-111 | Edge | Reactions: max distinct emoji types на одном сообщении | accept до лимита (если есть) или graceful; overflow→400 | API | P2 | US-S12 | TODO |
| ECC-MSG-112 | Edge | Reactions: один emoji — N users (1 / 100 / 1000) | counts корректны; payload не раздувается без пагинации/агрегации | API | P2 | TC-MSG-04 | TODO |
| ECC-MSG-113 | Edge | Pins: empty pins list канала | `[]` 200 | API | P2 | ListPins | TODO |
| ECC-MSG-114 | Edge | Pins: pin max N сообщений (если лимит) / N+1 | N OK; N+1→400; toggle unpin освобождает слот | API | P1 | US-S12, TogglePin | TODO |
| ECC-MSG-115 | Edge | Pins: pin уже pinned → unpin; повтор unpin | toggle idempotent; list без дублей | API | P1 | TC-MSG pin | TODO |
| ECC-MSG-116 | Edge | Unread: 0 каналов / 1 / много каналов с count=0 | map/`[]` без null; все нули | API | P1 | TC-MSG-09, ECC-MSG-035 | TODO |
| ECC-MSG-117 | Edge | Unread overflow: count у границы int32/int64 (массовый flood) | нет wrap в отрицательное; UI/API показывают capped или exact big-int safe | API | P0 | Increment unread | TODO |
| ECC-MSG-118 | Edge | Unread: mark-read при count=0 (idempotent) | 200; остаётся 0 | API | P2 | MarkRead | TODO |
| ECC-MSG-119 | Edge | GIF: сообщение с GIF attachment (presign) в пределах 25 MB | 201; preview/link OK | API | P1 | storage NFR-04, ECC-MSG-031 | TODO |
| ECC-MSG-120 | Edge | GIF: oversized / corrupt / wrong Content-Type as GIF | 400/413; нет hang на decode | API | P1 | NFR-04 | TODO |
| ECC-MSG-121 | Edge | GIF-only / sticker-like content без `content_text` | accept если content_json валиден; иначе 400 | API | P2 | SendMessageRequest | TODO |
| ECC-MSG-122 | Edge | Thread depth: reply к root (depth=1) vs reply к reply (depth≥2) | depth=1 OK; deeper → flat-to-root или 400 (политика 1-level) | API | P1 | TC-MSG-06, ListThreadReplies | TODO |
| ECC-MSG-123 | Edge | Thread: empty replies list / limit boundaries на replies | `[]` 200; limit clamp как у messages | API | P1 | NFR-15 | TODO |
| ECC-MSG-124 | Edge | Thread: reply_count на границе 0→1 и max practical | increment точный; overflow не ломает parent | API | P1 | IncrementReplyCount | TODO |
| ECC-MSG-125 | Edge | Mentions array: empty / 1 / max members / duplicate UUIDs | empty OK; dups dedup; overflow→400 | API | P2 | SendMessageRequest.Mentions | TODO |
| ECC-MSG-126 | Edge | Channel name/description boundaries: name 1/100/101; desc 500/501 | max OK; overflow 400 | API | P2 | CreateChannelRequest | TODO |
| ECC-MSG-127 | Edge | Mute until: past (−1s) / now / future (+1s) / far future | past→unmuted; future muted_until set | API | P2 | ECC-MSG-018 sibling | TODO |
| ECC-MSG-128 | Edge | Search messages in-channel: q min=1 / long query / limit 1/100/101 | min OK; long safe; limit clamp | API | P1 | SearchMessagesRequest | TODO |

---

## 2. WS — `ECC-WS-100+`

> Membership deny = security (NF-02) — здесь **edge**: max rooms, stale membership, reconnect/heartbeat/payload boundaries.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-WS-100 | Edge | Reconnect storm: N=10/50/100 reconnect одного user за 10s | сервер стабилен; старые conn drain; нет FD leak | Perf | P0 | ECC-NSW-028 sibling | TODO |
| ECC-WS-101 | Edge | Reconnect storm + concurrent subscribe тех же rooms | итоговые rooms consistent; нет dup fanout × N | API | P0 | Hub subscribe | TODO |
| ECC-WS-102 | Edge | Max rooms per client: 1 / practical max / max+1 subscribe | до max OK; сверх → reject/close; membership всё ещё проверяется | API | P0 | NF-02 edge, CanSubscribe | TODO |
| ECC-WS-103 | Edge | Subscribe storm: 100 subscribe/unsubscribe циклов одной room | final state subscribed XOR not; нет ghost room entry | Unit | P1 | hub.subscribeClient | TODO |
| ECC-WS-104 | Edge | Heartbeat: pong в пределах `pongWait=60s` | connection alive | Unit | P1 | TC-WS-03, pingPeriod≈54s | TODO |
| ECC-WS-105 | Edge | Heartbeat: pong на `60s−1s` vs `60s+1s` | −1s alive; +1s close/read deadline | Unit | P0 | pongWait | TODO |
| ECC-WS-106 | Edge | Heartbeat: client ping flood (1ms interval) | throttle/ignore; сервер не CPU-spin | Manual | P2 | TypePing | TODO |
| ECC-WS-107 | Edge | Payload size: inbound frame ровно 4096 vs 4097 bytes | 4096 accept path; 4097 → close (SetReadLimit) | Unit | P0 | maxMessageSize=4096 | TODO |
| ECC-WS-108 | Edge | Payload size: empty payload / minimal subscribe JSON / near-limit JSON | empty/minimal OK if schema valid; near-limit parse OK | API | P1 | protocol.go | TODO |
| ECC-WS-109 | Edge | Outbound send buffer full (slow client) | drop + warn; connection не deadlock | Unit | P1 | hub send buffer | TODO |
| ECC-WS-110 | Edge | Stale membership: subscribed → removed from channel; next event/recheck | unsubscribe или deny further; **нет** leak events (NF-02 edge) | API | P0 | ECC-NSW-037 sibling, ECC-MSG-036 | TODO |
| ECC-WS-111 | Edge | Stale membership cache TTL boundary (−1s / +1s) | после expiry re-authz; fail-closed при error | Unit | P0 | Authorizer cache | TODO |
| ECC-WS-112 | Edge | Max concurrent rooms globally on hub (many clients × many rooms) | stable; memory bounded; no panic | Perf | P1 | NFR-02 | TODO |
| ECC-WS-113 | Edge | JWT access expiry mid-WS (15m boundary) | close/401; reconnect with refresh | API | P0 | NFR-09, ECC-NSW-023 | TODO |
| ECC-WS-114 | Edge | Subscribe room name: empty / max length / unicode / unknown prefix | reject; no panic | API | P1 | CanSubscribe | TODO |
| ECC-WS-115 | Edge | writeWait=10s boundary: slow write | timeout close; no hung goroutine | Unit | P1 | writeWait | TODO |
| ECC-WS-116 | Edge | Message rate via WS path still capped 60/min | 61-й reject/429-equivalent; no bypass HTTP limiter | API | P0 | NFR-07, ECC-NSW-029 | TODO |

---

## 3. Notification — `ECC-NOTIF-100+`

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-NOTIF-100 | Edge | Digest/debounce batch: 2 события одного subject в окне 2s | coalesced → 1 delivery | Unit | P1 | consumer debounce, ECC-NSW-010 | TODO |
| ECC-NOTIF-101 | Edge | Digest window boundary: 2-е событие на `2s−1ms` vs `2s+1ms` | −1ms coalesce; +1ms отдельная notif | Unit | P1 | debounce map | TODO |
| ECC-NOTIF-102 | Edge | Digest batch size: 0 / 1 / N burst same key | 0 no-op; 1 passthrough; N→1 | Unit | P1 | cleanDebounce | TODO |
| ECC-NOTIF-103 | Edge | Preference matrix: type × channel (in_app/telegram/push) all-off | no delivery any channel | API | P0 | IsEnabled prefs | TODO |
| ECC-NOTIF-104 | Edge | Preference matrix: all-on vs mixed (in_app on, telegram off) | только enabled channels | API | P1 | TC-NOTIF prefs | TODO |
| ECC-NOTIF-105 | Edge | Preference: project-scoped override vs global nil project_id | project override wins; fallback global | API | P1 | repository IsEnabled | TODO |
| ECC-NOTIF-106 | Edge | Preference PUT: empty body / unknown type / unknown channel | 400; known matrix OK | API | P1 | PutPreferences | TODO |
| ECC-NOTIF-107 | Edge | Quiet hours: start==end / full-day / 1-minute window | documented behavior; no panic | API | P1 | quiet hours, NFR-12 | TODO |
| ECC-NOTIF-108 | Edge | Quiet hours TZ boundary: local midnight ±1s | defer/deliver per local TZ; UTC storage intact | API | P1 | ECC-NSW-032 sibling | TODO |
| ECC-NOTIF-109 | Edge | Empty notification list / unread-count=0 | `[]` / `0`; 200 not null | API | P2 | ECC-NSW-011 | TODO |
| ECC-NOTIF-110 | Edge | Unread-count overflow после массовых events | non-negative; capped or big-safe | API | P1 | unread-count | TODO |
| ECC-NOTIF-111 | Edge | Telegram bind: token TTL `10m−1s` vs `10m+1s` | −1s confirm OK; +1s reject expired | API | P0 | tgBindTTL=10m, ADR-AUTH-01 | TODO |
| ECC-NOTIF-112 | Edge | Telegram bind: confirm с empty / wrong / reused code | 400/401; reused invalid | API | P0 | ConfirmLinkCode | TODO |
| ECC-NOTIF-113 | Edge | Telegram bind: already linked to another account | 409 conflict | API | P0 | handler_telegram_bind | TODO |
| ECC-NOTIF-114 | Edge | Telegram bind: unlink when not linked (idempotent) | 200/noop; status linked=false | API | P2 | Unlink | TODO |
| ECC-NOTIF-115 | Edge | Telegram bind: bot not configured → CreateBindToken | 503 | API | P1 | 503 path | TODO |
| ECC-NOTIF-116 | Edge | Telegram test send: linked vs not linked | linked OK; not→400 | API | P2 | SendTest | TODO |
| ECC-NOTIF-117 | Edge | Mark-all-read при 0 / 1 / 1000 unread | все is_read; count→0; идемпотентно | API | P1 | TC-NOTIF-04 | TODO |
| ECC-NOTIF-118 | Edge | Web Push subscribe: endpoint length min/max / trailing junk | valid OK; overflow/SSRF→reject (ValidateOutboundURL) | API | P0 | NF-07, NFR-21 | TODO |

---

## 4. Search — `ECC-SRCH-100+`

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-SRCH-100 | Edge | Empty query `q=""` / whitespace-only | 400 (`validate:required,min=1`) | API | P1 | search/dto.go, ECC-NSW-017 | TODO |
| ECC-SRCH-101 | Edge | Query length: 1 / 500 / 501 runes | 1..500 OK; 501→400 | API | P1 | validate max=500 | TODO |
| ECC-SRCH-102 | Edge | Query: unicode / emoji / CJK / special chars (`*`, `"`, `\`) | safe; no injection; stable 200/[] | API | P1 | ECC-NSW-033 sibling | TODO |
| ECC-SRCH-103 | Edge | User-scope: member с 0 проектов / 0 memberships | `[]` 200; **нет** чужих hits (NF-03) | API | P0 | NF-03, NFR-19 | TODO |
| ECC-SRCH-104 | Edge | User-scope: query matches only foreign-project docs | `[]`; zero leak | API | P0 | ECC-NSW-016 sibling | TODO |
| ECC-SRCH-105 | Edge | User-scope: soft-deleted entities still indexed? | excluded from results | API | P1 | ECC-NSW-019 sibling | TODO |
| ECC-SRCH-106 | Edge | `limit`: 1 / 20 default / 50 max / 51 / 0 / negative | 1..50 OK; 51→clamp 20; 0/neg→20 | API | P1 | SearchService limit clamp | TODO |
| ECC-SRCH-107 | Edge | Meilisearch timeout / slow host | graceful fallback PG/Mongo или 503; no hang >deadline | API | P0 | MeiliIndexer optional | TODO |
| ECC-SRCH-108 | Edge | Meilisearch down / MEILI_HOST unset | search всё ещё работает без Meili (sidecar optional) | API | P1 | NewMeiliIndexerFromEnv | TODO |
| ECC-SRCH-109 | Edge | Meilisearch: empty index / zero docs | `[]` 200; no 500 | API | P2 | IndexAllIssues | TODO |
| ECC-SRCH-110 | Edge | Multi-type filter: empty types / one type / all types / unknown type | empty→all; unknown ignore или 400 | API | P2 | SearchRequest.Types | TODO |
| ECC-SRCH-111 | Edge | Ctrl+K: очень частые запросы (debounce UI + API) | no 429 storm beyond API 100/s; UI debounce | E2E | P2 | US-M07 | Manual |
| ECC-SRCH-112 | Edge | Saved filter: name empty / max length / duplicate name | empty 400; max OK; dup policy | API | P2 | TC-SEARCH-03 | TODO |
| ECC-SRCH-113 | Edge | JQL edge: empty JQL / max-length JQL / unbalanced quotes | 400 parse; no panic | API | P1 | ParseJQL | TODO |
| ECC-SRCH-114 | Edge | Parallel shard search: one shard timeout, others OK | partial results + error log **или** fail-closed documented | API | P1 | SearchService parallel | TODO |

---

## Счётчик

| Домен | ID range | Кол-во Edge | P0 | P1 | P2 |
|-------|----------|-------------|----|----|-----|
| Messenger | ECC-MSG-100…128 | **29** | 5 | 16 | 8 |
| WS | ECC-WS-100…116 | **17** | 8 | 7 | 2 |
| Notification | ECC-NOTIF-100…118 | **19** | 6 | 9 | 4 |
| Search | ECC-SRCH-100…114 | **15** | 3 | 8 | 4 |
| **Итого** | | **80** | **22** | **40** | **18** |

> Минимум 45 — **выполнено (80)**.  
> Happy / Neg / Corner — **не включены**.  
> Основной `EDGE-CORNER-CATALOG.md` — **не изменялся**.  
> Коммит — **не создавался**.
