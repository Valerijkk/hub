# Edge expansions wave5 — Messenger / WS / Notification / Search

> **Тип:** только **Edge** (границы: mute duration enum, leave/delete/role, promote-to-issue caps, @group mentions, typing, custom-status `today`, NATS→WS bridge subjects, `call:` prefix, CORS_ORIGINS, DB CreateNotification dedup, body truncate 100, push/TG queue depths, deadline ±24h, message score×0.8, ILIKE 0.5, soft-delete Mongo $text).  
> **Волна:** wave5 · **не копии** wave1 `*100…`, wave2 `*200…`, wave3 `*300…`, wave4 `*400…`.  
> **Источник:** код (`ent_repository.MuteChannel` 1h|2h|8h|1d|forever→100y, `service_promote.go` title truncate 500 / ListThreadReplies 100, `service.go` resolveGroupMentions @all|@online|@role, `ws/bridge.go` subjects, `ws/authorizer.go` channel|call|project, `notification/repository.go` CreateNotification 2s DB-dedup, `consumer.go` body>100→`...`, `push.go` queue=200 / 404|410, `telegram.go` queue=100, `CheckDeadlines` now..now+24h, `search/service.go` Score×0.8 / ILIKE 0.5 / messages без `deleted_at`).  
> **ID:** Messenger `ECC-MSG-500+`, WS `ECC-WS-500+`, Notification `ECC-NOTIF-500+`, Search `ECC-SRCH-500+`.  
> **Gap:** только `TODO` | `Future` | `Manual`.  
> **Владелец:** `qa-manual` · 2026-07-10 · каталог не трогать · не коммитить.

Формат: `| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |`

### Отличие от wave1–4 (фокус wave5)

| Тема | Wave1–2 (не повторять) | Wave3–4 (не повторять) | Wave5 (`*500+`) — новое |
|------|------------------------|------------------------|-------------------------|
| Messenger | content/BodyLimit, thread depth, presence away/stale, rate 60/min, empty lists, reactions emoji len, pins, GIF, channel name/desc | polls 2..10, invites max_uses/TTL/token=22, voice ms, bookmarks/DM/forward, readonly, send-buffer themes | **mute enum 1h\|2h\|8h\|1d\|forever**, **leave/delete/role**, **promote-to-issue** (title/replies/issue_type), **@all/@online/@role**, **typing**, **custom-status `today`**, **unread skip-sender** |
| WS | reconnect, 4096 frame, pongWait 60s, writeWait, JWT mid-WS | send buf 256, register 64, pingPeriod 54s, multi-tab, room prefix generic | **NATS Bridge subjects**, **`call:` checker**, **nil MembershipChecker**, **CORS_ORIGINS empty prod**, **unregister=64**, **auth.session→user**, **Bridge Start/Stop** |
| Notification | digest debounce 2s in-memory, prefs matrix, quiet hours basic, bind TTL 10m, ValidateOutboundURL | MaxDeliver/AckWait/PullMax, per_page, tgLinkCodeTTL 5m, WebPush TTL 60s, spacing 50/35ms | **DB CreateNotification dedup 2s**, **body truncate 100**, **push queue 200**, **TG queue 100**, **404/410 dead sub**, **deadline window ±24h**, **DM vs mention type**, **entity_id empty skips dedup** |
| Search | empty q, q=500, Meili typo 5/9, NF-03 scope, limit 50 | shard limit 10, batch 5000, LimitReader 4096, saved-filter name 200, score ties | **message Score×0.8**, **channel ILIKE rank=0.5**, **Mongo $text без deleted_at**, **Meili fields subset**, **0 channels early exit**, **Meili HTTP 30s**, **plainto_tsquery punctuation** |

---

## 1. Messenger — `ECC-MSG-500+`

> Фокус: **mute duration enum**, **leave/delete/role**, **promote-to-issue**, **group mentions**, **typing**, **custom-status `today`**, **unread skip-sender**.  
> Не повторять: emoji max=32, pins, rate 60/min, name/desc caps, polls/invites/voice (это *100–*400).

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MSG-500 | Edge | Mute `duration`: `1h` / `2h` / `8h` / `1d` — boundary at `muted_until−1s` vs `+1s` | −1s still muted; +1s unmuted (`IsChannelMuted`) | API | P0 | MuteChannel enum (не past/now из *100) | TODO |
| ECC-MSG-501 | Edge | Mute `duration=forever` → `muted_until` ≈ now+100y | far-future set; unmute clears; list shows muted | API | P0 | forever→100*365d | TODO |
| ECC-MSG-502 | Edge | Mute `duration` unknown / empty / `30m` (не в enum) | 400/`некорректная длительность`; muted_until unchanged | API | P0 | switch default error | TODO |
| ECC-MSG-503 | Edge | Unmute when never muted / after forever / after 1h | 200 idempotent; muted_until null/zero | API | P1 | UnmuteChannel | TODO |
| ECC-MSG-504 | Edge | Re-mute while muted: shorter then longer duration | muted_until overwritten to new absolute; no stack | API | P1 | MuteChannel Update | TODO |
| ECC-MSG-505 | Edge | Delete channel: role=`owner` vs `admin` vs `member` | owner OK; admin/member→403 | API | P0 | DeleteChannel owner-only | TODO |
| ECC-MSG-506 | Edge | Leave channel: sole member / last owner / ordinary member | member leave OK; sole-owner policy documented (orphan or block) | API | P0 | LeaveChannel | TODO |
| ECC-MSG-507 | Edge | Leave DM channel vs leave public/group | DM leave policy (archive/hide vs 400); group removes membership | API | P1 | channel_type × leave | TODO |
| ECC-MSG-508 | Edge | Leave then send / WS subscribe same channel | 403 send; CanSubscribe reject (NF-02) | API | P0 | membership after leave | TODO |
| ECC-MSG-509 | Edge | AddMembers: caller role `member` vs `admin` vs `owner` | member→403; admin/owner OK; new role always `member` | API | P0 | AddMembers | TODO |
| ECC-MSG-510 | Edge | ListMembers: non-member caller vs member | non-member→403 (HI-01); member→full list | API | P0 | ListMembers HI-01 | TODO |
| ECC-MSG-511 | Edge | Promote-to-issue: `title` length 500 / 501 (server truncate) | 500 OK; 501→stored truncated to 500 (не 400) | API | P0 | service_promote title[:500] | TODO |
| ECC-MSG-512 | Edge | Promote-to-issue: thread replies 0 / 100 / 101 | description includes ≤100 replies; 101-й omitted | API | P0 | ListThreadReplies(..., 100) | TODO |
| ECC-MSG-513 | Edge | Promote-to-issue: `issue_type` empty / `task` / known enum / unknown | empty→`task`; known OK; unknown→400 validate | API | P1 | oneof issue_type | TODO |
| ECC-MSG-514 | Edge | Promote-to-issue: `IssuePromoter` nil (не сконфигурирован) | 500/internal «не сконфигурирован»; no partial issue | API | P1 | issuePromoter==nil | TODO |
| ECC-MSG-515 | Edge | Promote: soft-deleted parent thread + live replies | continues without parent (warn); description from available msgs | API | P1 | parent not found path | TODO |
| ECC-MSG-516 | Edge | Group mentions: `@all` in channel with N=1 (only sender) / N=100 | expands to members **excluding sender**; N=1→empty mention fanout | API | P0 | resolveGroupMentions | TODO |
| ECC-MSG-517 | Edge | Group mentions: `@online` when 0 online / all online / mixed | only currently online member IDs; sender excluded | API | P1 | @online + presence | TODO |
| ECC-MSG-518 | Edge | Group mentions: `@role` unknown project_role / empty project channel | unknown→no IDs or 400; no-project channel→skip role expand | API | P1 | ListProjectMembersByRole | TODO |
| ECC-MSG-519 | Edge | Typing: non-member POST typing / member / after leave | non-member→403; member→NATS ephemeral (no Mongo row); after leave→403 | API | P1 | PublishTyping, SubjectTyping | TODO |
| ECC-MSG-520 | Edge | Typing flood: 1/ms vs practical UI throttle | server accepts or rate-limits; no OOM; WS fans out bounded | Perf | P2 | typing ephemeral | TODO |
| ECC-MSG-521 | Edge | Custom status duration=`today`: set at 23:59:00 vs 23:59:59 local | TTL→next midnight (≥1m floor); after midnight key gone | Unit | P1 | ParseStatusDuration today | TODO |
| ECC-MSG-522 | Edge | Custom status duration=`0` → TTL ~365d; unknown → default 24h | Redis TTL matches; empty text deletes key | Unit | P2 | ParseStatusDuration 0/default | TODO |
| ECC-MSG-523 | Edge | Unread Increment: sender of message vs other members | sender unread **not** incremented; peers +1 | Unit | P0 | unread.go skip sender | TODO |
| ECC-MSG-524 | Edge | MarkRead: Redis HDEL field for channel; GetUnread omits channel | field gone; map without key (not 0-stuck) | Unit | P1 | MarkRead HDEL | TODO |
| ECC-MSG-525 | Edge | `content_json` required: `null` / `{}` / missing key | null/missing→400; `{}` policy documented with/without text | API | P1 | SendMessageRequest validate required | TODO |
| ECC-MSG-526 | Edge | Edit message: author vs other member vs admin; empty new content | author OK; other→403; empty→400 | API | P1 | EditMessage ACL | TODO |
| ECC-MSG-527 | Edge | Soft-delete message: author vs admin/owner vs plain member | author+admin/owner OK; member→403 | API | P0 | DeleteMessage roles | TODO |
| ECC-MSG-528 | Edge | Auto-channel: project key with UPPER/spaces → `#project-{key}` lowercased | exact naming; members synced; duplicate create idempotent | API | P2 | auto_channels.go | TODO |
| ECC-MSG-529 | Edge | Root message list filter: msgs with `thread_id=""` vs missing field | empty-string not treated as root **или** documented leak; `$exists:false` only | API | P1 | ListMessages thread_id filter | TODO |

---

## 2. WS — `ECC-WS-500+`

> Фокус: **NATS Bridge routing**, **`call:` prefix**, **nil checker**, **CORS_ORIGINS**, **unregister=64**, **auth.session fanout**.  
> Не повторять: frame 4096, pongWait/pingPeriod, send buf=256, multi-tab fanout (*100–*400).

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-WS-500 | Edge | Bridge subject `messenger.channel.{id}.message_created` → room `channel:{id}` | subscribers of that room get event; other rooms silent | Unit | P0 | bridge handleMessengerEvent | TODO |
| ECC-WS-501 | Edge | Bridge `project.{id}.*` → `project:{id}`; malformed `<3` parts | valid route; short subject no panic / drop | Unit | P0 | handleProjectEvent | TODO |
| ECC-WS-502 | Edge | Bridge `calls.channel.{id}.*` → messenger channel room; `calls.chat.{callID}` → `call:{callID}` | correct room mapping; no cross-leak | Unit | P0 | handleCallEvent / handleCallChatEvent | TODO |
| ECC-WS-503 | Edge | Bridge `notification.user.{userID}` → SendToUser (not room) | all tabs of user get; other users 0 | Unit | P0 | handleNotificationEvent | TODO |
| ECC-WS-504 | Edge | Bridge `auth.session.{userID}.*` (revoke/logout) | targeted user conns closed/notified; others intact | Unit | P0 | handleAuthSessionEvent | TODO |
| ECC-WS-505 | Edge | Room `call:{uuid}` with checker set vs `IsCallParticipantOrChannelMember=nil` | set→membership result; nil→deny fail-closed | Unit | P0 | authorizer call prefix | TODO |
| ECC-WS-506 | Edge | Room `channel:` / `:uuid` / `channel` (no colon) / `channel:uuid:extra` | malformed→reject (`splitRoom` ok=false); extra suffix invalid UUID→reject | API | P0 | splitRoom | TODO |
| ECC-WS-507 | Edge | Unknown prefix `user:{uuid}` / `dm:{uuid}` / `notif:{uuid}` | deny (fail-closed); **не** allow-all | API | P0 | default case false | TODO |
| ECC-WS-508 | Edge | `CORS_ORIGINS` empty + not DEV: WS upgrade | reject all origins; log hint | Unit | P0 | initOriginCheck | TODO |
| ECC-WS-509 | Edge | Origin: exact allowlist match vs trailing-slash / case / subdomain | only exact policy match OK; mismatch reject | Unit | P1 | CheckOrigin | TODO |
| ECC-WS-510 | Edge | Hub `unregister` chan capacity 64: burst unregister 64 / 65 | ≤64 buffered; overflow policy documented; no panic; maps clean | Unit | P1 | unregister chan=64 | TODO |
| ECC-WS-511 | Edge | Bridge Start then Stop then late NATS msg | after Stop no fanout / no panic on nil subs | Unit | P1 | Bridge.Stop | TODO |
| ECC-WS-512 | Edge | Bridge Start when NATS nil/disconnected | error or no-op documented; hub still serves HTTP WS | Unit | P1 | NewBridge Start | TODO |
| ECC-WS-513 | Edge | Bridge payload: empty body / invalid JSON / huge near practical max | drop+log; hub alive; no panic | Unit | P0 | Bridge unmarshal | TODO |
| ECC-WS-514 | Edge | Subscribe `project:{uuid}` with `IsProjectMember=nil` | deny; no panic | Unit | P0 | nil project checker | TODO |
| ECC-WS-515 | Edge | Non-UUID userIDStr in CanSubscribe | error; deny | Unit | P1 | uuid.Parse userID | TODO |
| ECC-WS-516 | Edge | SendToUser when user has 0 WS conns | noop; no panic; NATS path unaffected | Unit | P2 | Hub SendToUser | TODO |

---

## 3. Notification — `ECC-NOTIF-500+`

> Фокус: **DB-layer dedup 2s**, **body truncate 100**, **push/TG queue depths**, **404/410**, **deadline ±24h**, **DM vs mention type**.  
> Не повторять: in-memory digest size, per_page clamp, tgLinkCodeTTL=5m, push TTL/spacing (*100–*400).

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-NOTIF-500 | Edge | DB `CreateNotification` same `(user,type,entity)` at `2s−1ms` vs `2s+1ms` | −1ms → nil skip (no 2nd row); +1ms → 2nd row | Unit | P0 | repository CreateNotification cutoff | TODO |
| ECC-NOTIF-501 | Edge | DB dedup with `entity_id=""` (empty) | dedup **skipped**; both inserts allowed | Unit | P0 | `if entityID != ""` | TODO |
| ECC-NOTIF-502 | Edge | DB dedup vs in-memory `shouldDispatch` debounce: both fire same key | at most 1 inbox row; layers complementary not double-drop bug | Unit | P1 | consumer + repository | TODO |
| ECC-NOTIF-503 | Edge | Message notif body: 100 / 101 / 1000 chars | 100 unchanged; 101→100+`...` (≤103); no panic | Unit | P0 | consumer truncate >100 | TODO |
| ECC-NOTIF-504 | Edge | Body truncate mid-UTF-8 / emoji at byte 100 | no invalid UTF-8 in stored body **или** documented byte-cut | Unit | P1 | truncate bytes vs runes | TODO |
| ECC-NOTIF-505 | Edge | Push queue depth: enqueue 200 / 201 | ≤200 queued; 201st drop+warn; workers drain | Unit | P0 | push.go chan 200 | TODO |
| ECC-NOTIF-506 | Edge | Telegram queue depth: 100 / 101 | ≤100 queued; 101st drop+warn | Unit | P0 | telegram.go chan 100 | TODO |
| ECC-NOTIF-507 | Edge | Push endpoint HTTP 404 / 410 | subscription deleted from DB; no retry storm | Unit | P0 | sendToSubscription 404/410 | TODO |
| ECC-NOTIF-508 | Edge | Push endpoint HTTP 500 / timeout at 10s | no delete sub; retry/backoff; client timeout bound | Unit | P1 | safehttp 10s | TODO |
| ECC-NOTIF-509 | Edge | CheckDeadlines: due at `now−1s` / `now` / `now+24h−1s` / `now+24h+1s` | only `(now, now+24h]` (или documented inclusive); outside→no notif | Unit | P0 | ListIssuesWithUpcomingDeadline | TODO |
| ECC-NOTIF-510 | Edge | CheckDeadlines: issue without assignee / soft-deleted / due null | skipped; no panic; count only eligible | Unit | P1 | CheckDeadlines loop | TODO |
| ECC-NOTIF-511 | Edge | DM message event → `TypeMessageReceived`; channel @mention → `TypeMentioned` | type split by channel_type / mentions; prefs key differs | Unit | P0 | consumer DM branch | TODO |
| ECC-NOTIF-512 | Edge | Unknown NATS subject → `SubjectToType=""` | no CreateNotification; ack/skip; no 500 | Unit | P1 | events.SubjectToType | TODO |
| ECC-NOTIF-513 | Edge | Preference types matrix cardinality: all known types × channels on/off | each cell independent; unknown type→400 | API | P1 | PutPreferences | TODO |
| ECC-NOTIF-514 | Edge | Delete single notification: own / foreign / already deleted | own 200; foreign 404; second→404 | API | P2 | DeleteNotification | TODO |
| ECC-NOTIF-515 | Edge | Link code length exactly 6 vs 5 / 7 (digits) | 6 path; ≠6→`ErrTelegramInvalidCode` | API | P1 | ConfirmLinkCode len!=6 (≠ bind TTL *100) | TODO |
| ECC-NOTIF-516 | Edge | Push worker: 0 subscriptions for user | no-op; queue not blocked | Unit | P2 | len(subs)==0 | TODO |

---

## 4. Search — `ECC-SRCH-500+`

> Фокус: **Score×0.8**, **ILIKE 0.5**, **Mongo soft-delete gap**, **Meili field subset**, **0-channels**, **tsquery punctuation**.  
> Не повторять: empty q / q=500 / typoTolerance / shard=10 merge / LimitReader 4096 / batch 5000 as primary (*100–*400).

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-SRCH-500 | Edge | Message hit raw textScore=1.0 → returned Score=0.8 | demotion applied; ranks below equal PG hit | API | P0 | `doc.Score * 0.8` | TODO |
| ECC-SRCH-501 | Edge | Equal raw scores: issue ts_rank=0.8 vs message textScore=1.0 (→0.8) | tie policy documented; both may appear; deterministic order | API | P1 | merge sort Score | TODO |
| ECC-SRCH-502 | Edge | Channel search: tsvector miss → ILIKE fallback rank **0.5** | constant 0.5; ordered by name ASC within fallback | API | P0 | `0.5::float8` fallback | TODO |
| ECC-SRCH-503 | Edge | Channel: tsvector hit vs ILIKE-only hit same name | ts_rank preferred path; no dup rows | API | P1 | searchChannels dual path | TODO |
| ECC-SRCH-504 | Edge | Soft-deleted message still in Mongo `$text` (filter **без** `deleted_at`) | may return deleted — document as gap **или** exclude; NF-03 still scoped to channels | API | P0 | searchMessages filter | TODO |
| ECC-SRCH-505 | Edge | Soft-deleted **issue** still in Meili until reindex | scope/DeletedAt filter excludes **или** stale until IndexAll; no ACL leak | API | P0 | meili + soft delete | TODO |
| ECC-SRCH-506 | Edge | Meili indexed fields only `id,issue_key,title,project_id` — query matches description-only | no Meili hit on description; PG path may still hit | API | P1 | meiliIssueDoc fields | TODO |
| ECC-SRCH-507 | Edge | User with **0 channels**: `type=message` | `[]` early; no Mongo $text | API | P0 | len(channelIDs)==0 | TODO |
| ECC-SRCH-508 | Edge | User with **0 projects**: `type=issue` / `project` / `file` / `test_case` | those shards empty; other types (user/channel) may still return | API | P0 | len(projectIDs)==0 | TODO |
| ECC-SRCH-509 | Edge | Meili HTTP client Timeout=30s: response at 30s−1 vs hang | −1 OK; hang→timeout error; search fallback/503 | Unit | P1 | NewMeiliIndexerFromEnv | TODO |
| ECC-SRCH-510 | Edge | `plainto_tsquery` / `$text` with punctuation-only `q` (`!!!`, `***`) | `[]` or 400; no 500/panic | API | P1 | tsquery edge | TODO |
| ECC-SRCH-511 | Edge | Message hit title/snippet truncate at 100 + `...` | title ≤103; full body not dumped | API | P1 | searchMessages title truncate | TODO |
| ECC-SRCH-512 | Edge | Types CSV parse: `issue,message` / spaces / trailing comma | known types searched; empties ignored; no double shard | API | P2 | handler strings.Split | TODO |
| ECC-SRCH-513 | Edge | Parallel shards: message Mongo down, PG shards OK | partial results + joined error/log; never unauthorized hits | API | P0 | Search ME-2 | TODO |
| ECC-SRCH-514 | Edge | File search: filename exact / extension-only `.pdf` / path-like | scoped by project membership; no cross-project leak (NF-03) | API | P0 | search files | TODO |
| ECC-SRCH-515 | Edge | User search: display_name unicode / deactivated user | deactivated hidden or marked; no email enumerator beyond scope | API | P1 | search users | TODO |
| ECC-SRCH-516 | Edge | Ctrl+K: switch type filter empty→single type mid-debounce | last request wins; no stale multi-type results stuck | E2E | P2 | US-M07 | Manual |

---

## Счётчик

| Домен | ID range | Кол-во Edge | P0 | P1 | P2 |
|-------|----------|-------------|----|----|-----|
| Messenger | ECC-MSG-500…529 | **30** | 13 | 14 | 3 |
| WS | ECC-WS-500…516 | **17** | 11 | 5 | 1 |
| Notification | ECC-NOTIF-500…516 | **17** | 8 | 7 | 2 |
| Search | ECC-SRCH-500…516 | **17** | 8 | 7 | 2 |
| **Итого** | | **81** | **40** | **33** | **8** |

> Минимум 50 — **выполнено (81)**.  
> Happy / Neg / Corner — **не включены**.  
> ID `*100+` / `*200+` / `*300+` / `*400+` / каталог `*001…` — **не копии** (новые оси: mute enum/promote/bridge/DB-dedup/score×0.8).  
> Явно избегали копий wave1: emoji≤32, pins, rate 60, name/desc, q=500, typoTolerance, LimitReader/batch как primary.  
> `.planning/archive/july-2026/EDGE-CORNER-CATALOG.md` — **не изменялся**.  
> Коммит — **не создавался**.
