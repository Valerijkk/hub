# Edge expansions wave4 — Messenger / WS / Notification / Search

> **Тип:** только **Edge** (границы: poll options, invite TTL/uses, voice ms, bookmark/star, WS send-buffer 256, notif page/per_page, tgLinkCodeTTL, WebPush TTL, search per-shard limit=10, saved-filter name).  
> **Источник:** NFR/TZ + код (`messenger/dto.go` poll options 2..10 / question max=500, `service_channel_invites.go` expires_in∈{1h,1d,7d,30d,""} / token len=22, `ws/hub.go` send buf=256 / register=64 / subscribe=256, `ws/handler.go` pingPeriod≈54s, `notification/handler.go` per_page clamp 1..100→20, `telegram_bind.go` tgLinkCodeTTL=5m, `push.go` TTL=60s / 50ms spacing, `telegram.go` 35ms, `search/service.go` defaultSearchLimit=10 / final limit clamp≤50, `search/dto.go` SavedFilter name max=200).  
> **Не дублирует** каталог `ECC-MSG-001…040` / `ECC-NSW-001…038`, wave1 `ECC-*-100…`, wave2 `ECC-*-200…`, wave3 `ECC-*-300…`. Sibling-ссылки допустимы.  
> **ID:** Messenger `ECC-MSG-400+`, WS `ECC-WS-400+`, Notification `ECC-NOTIF-400+`, Search `ECC-SRCH-400+`.  
> **Gap:** только `TODO` | `Future` | `Manual`.  
> **Владелец:** `qa-manual` · 2026-07-10 · каталог не трогать · не коммитить.

Формат: `| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |`

---

## 1. Messenger — `ECC-MSG-400+`

> Фокус: **polls**, **channel invites**, **forward**, **bookmarks/star**, **voice duration**, **member batch**, **DM**, **mention_groups**, **soft-delete в cursor**.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MSG-400 | Edge | Poll `options` count: 1 / 2 / 10 / 11 | 1→400; 2..10 OK; 11→400 (`min=2,max=10`) | API | P0 | CreatePollRequest, service_polls.go | TODO |
| ECC-MSG-401 | Edge | Poll `question` length: 0 / 1 / 500 / 501 | 0→400; 1..500 OK; 501→400 | API | P0 | validate max=500 | TODO |
| ECC-MSG-402 | Edge | Poll option text: empty string among 2..10 options / whitespace-only option | 400; нет пустых option rows | API | P1 | CreatePollWithOptions | TODO |
| ECC-MSG-403 | Edge | Poll option text: very long single option (near practical DB/UI max) vs overflow | max OK или 400; overflow→400; no truncate mid-grapheme | API | P1 | PollOption | TODO |
| ECC-MSG-404 | Edge | Poll vote: `option_ids` length 0 / 1 / all options (multi) / duplicate IDs | 0→400; single OK; multi only if `is_multi_choice`; dups dedup or 400 | API | P0 | VotePollRequest | TODO |
| ECC-MSG-405 | Edge | Poll multi-choice: select exactly 1 vs all N options vs N+1 invalid UUID | 1..N valid OK; invalid UUID→400; closed poll→reject | API | P0 | VotePoll | TODO |
| ECC-MSG-406 | Edge | Poll anonymous: after votes, `voter_ids` empty in response; `voted` flag for self still true | no PII leak; counts exact | API | P0 | toPollResponse isAnonymous | TODO |
| ECC-MSG-407 | Edge | Poll `total_votes`: 0 voters / 1 / many unique voters (same user re-vote) | unique voters; re-vote не раздувает total без политики | API | P1 | uniqueVoters map | TODO |
| ECC-MSG-408 | Edge | Poll close then vote / vote then close | close→vote 400; after close results frozen | API | P1 | ClosePoll | TODO |
| ECC-MSG-409 | Edge | Invite `max_uses`: 0 (unlimited) / 1 / N / use_count==max / use_count==max+attempt | 0=∞; 1→второй accept invalid; at max→«исчерпан лимит» | API | P0 | ValidateInvite/AcceptInvite | TODO |
| ECC-MSG-410 | Edge | Invite `expires_in`: `""` / `1h` / `1d` / `7d` / `30d` / unknown (`2h`,`abc`) | empty→ExpiresAt=nil; known set; unknown→400 | API | P0 | parseExpiresIn | TODO |
| ECC-MSG-411 | Edge | Invite expiry boundary: accept at `ExpiresAt−1s` vs `ExpiresAt+1s` | −1s OK; +1s invalid «срок истёк» (UTC) | API | P0 | ExpiresAt.Before | TODO |
| ECC-MSG-412 | Edge | Invite token length: empty / 21 / 22 (canonical) / 23 / garbage charset | only valid 22-base62 found; else valid=false | API | P1 | inviteTokenLength=22 | TODO |
| ECC-MSG-413 | Edge | Invite list empty for channel / many invites (revoked+active) | `[]` or filtered; revoked not usable | API | P2 | ListInvites | TODO |
| ECC-MSG-414 | Edge | Forward: target = same channel / other member channel / non-member target | same policy documented; non-member→403; member→201 + `forwarded_from_id` | API | P0 | ForwardMessage | TODO |
| ECC-MSG-415 | Edge | Forward: source soft-deleted / missing UUID / empty `target_channel_id` | 404/400; no orphan forward | API | P1 | ForwardMessageRequest | TODO |
| ECC-MSG-416 | Edge | Forward chain: forward-of-forward (depth 1 vs 2+) | `forwarded_from_id` points to immediate or original per policy; no infinite metadata growth | API | P2 | ForwardedFromID | TODO |
| ECC-MSG-417 | Edge | Bookmarks list: 0 / 1 / large N | `[]` 200; order stable; no null | API | P1 | ListBookmarks | TODO |
| ECC-MSG-418 | Edge | Bookmark add: duplicate same message / deleted message / foreign channel | dup idempotent or 409; deleted→404; foreign→403 | API | P1 | AddBookmark | TODO |
| ECC-MSG-419 | Edge | Bookmark remove: last bookmark / already removed id | 200/noop or 404; list empty after last | API | P2 | RemoveBookmark | TODO |
| ECC-MSG-420 | Edge | Star toggle: unstarred→starred→unstarred; list channels with only starred filter empty | boolean flips; empty starred set → `[]` | API | P1 | ToggleStar | TODO |
| ECC-MSG-421 | Edge | Voice: `voice_duration_ms` = 0 / 1 / practical max / negative / huge int | 0/neg→400; 1..max OK; overflow→400 | API | P1 | SendMessageRequest VoiceDurationMs | TODO |
| ECC-MSG-422 | Edge | Voice: `voice_url` empty with duration>0 / SSRF-like URL / oversize URL string | empty→400; SSRF reject (ValidateOutboundURL if fetched); oversize→400 | API | P0 | NFR-21, ECC-MSG-040 sibling | TODO |
| ECC-MSG-423 | Edge | Voice-only message: no `content_text`, valid voice fields | 201 if policy allows; else 400 | API | P2 | US-S12 voice | TODO |
| ECC-MSG-424 | Edge | `mention_groups`: empty / 1 / all channel groups / unknown group id / dups | empty OK; unknown→400/ignore; dups dedup | API | P2 | MentionGroups | TODO |
| ECC-MSG-425 | Edge | Add members batch: 1 / practical max / max+1 / empty array / duplicate UUIDs | empty→400 (`min=1`); dups dedup; overflow→400 | API | P1 | AddMembersRequest | TODO |
| ECC-MSG-426 | Edge | DM create: self as target / same pair twice / deactivated target | self→400; second→same channel id; deactivated→403/404 | API | P0 | CreateDMRequest, TC-MSG-02 | TODO |
| ECC-MSG-427 | Edge | Soft-deleted message in cursor page: `before` lands on deleted / mix deleted+live | deleted hidden or marked; cursor still advances; no stuck page | API | P1 | ListMessages DeletedAt | TODO |
| ECC-MSG-428 | Edge | Channel type create: `dm`/`group`/`public`/`private` + name empty for public vs dm | public needs name; dm name optional; invalid type→400 | API | P1 | CreateChannelRequest oneof | TODO |
| ECC-MSG-429 | Edge | Readonly channel: member list empty posts / admin post / bot post boundary | non-admin→403; admin OK; bot policy documented | API | P0 | TC-MSG-07, is_readonly | TODO |

---

## 2. WS — `ECC-WS-400+`

> Фокус: **send buffer=256**, **hub channel capacities**, **pingPeriod≈54s**, **multi-tab fanout**, **room prefix**, **register queue**.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-WS-400 | Edge | Client `send` buffer: fill to 255 / 256 / 257 events without consumer | ≤256 queued; 257-й → drop+warn (`send buffer full`); no deadlock | Unit | P0 | hub.go send chan 256 | TODO |
| ECC-WS-401 | Edge | Send buffer: exactly 256 then client resumes reading | drain OK; subsequent delivers; no permanent mute | Unit | P0 | Client.send | TODO |
| ECC-WS-402 | Edge | Hub `register` queue: burst 64 / 65 concurrent connects | ≤64 buffered; overflow behavior documented (block/drop); no panic | Perf | P1 | register chan 64 | TODO |
| ECC-WS-403 | Edge | Hub `subscribe` queue: burst 256 / 257 subscribe reqs | ≤256 buffered; final room set consistent after drain | Unit | P1 | subscribe chan 256 | TODO |
| ECC-WS-404 | Edge | Hub `unsubscribe` queue same as subscribe capacity boundary | no ghost rooms after burst unsub | Unit | P1 | unsubscribe chan 256 | TODO |
| ECC-WS-405 | Edge | `pingPeriod`≈54s: server ping at 54s−1s / 54s / 54s+1s cadence under load | pings within pongWait; connection stays alive if pong timely | Unit | P1 | pingPeriod=(pongWait*9)/10 | TODO |
| ECC-WS-406 | Edge | Client delays pong until just before next ping (54s window stress) | alive if within pongWait=60s; else close | Unit | P0 | pongWait vs pingPeriod | TODO |
| ECC-WS-407 | Edge | Multi-tab: same user 2 / 5 / 20 WS conns subscribed to same room | each conn gets event (fanout ×N); no cross-user leak | API | P0 | Hub Broadcast | TODO |
| ECC-WS-408 | Edge | Multi-tab: one tab buffer-full drop, others healthy | isolated drop; other tabs still receive | Unit | P1 | per-client send buf | TODO |
| ECC-WS-409 | Edge | Room prefix: `channel:{uuid}` / `project:{uuid}` / `user:{uuid}` / unknown prefix / missing colon | valid prefix+membership OK; unknown→reject; malformed→reject | API | P0 | CanSubscribe, NF-02 | TODO |
| ECC-WS-410 | Edge | Room id: valid UUID vs non-UUID suffix after prefix | non-UUID→reject; no panic | API | P1 | room parse | TODO |
| ECC-WS-411 | Edge | Origin header: missing / null origin / exact allowlist / subdomain mismatch | reject fail-closed per config; allowlist match OK | Unit | P0 | ECC-NSW-026 sibling | TODO |
| ECC-WS-412 | Edge | Concurrent register+unregister same client id storm | final hub state without leaked client maps | Unit | P1 | hub Run loop | TODO |
| ECC-WS-413 | Edge | Outbound event JSON size vs slow client: event near practical max while buffer near 256 | drops preferred over block; hub loop never stalls | Perf | P0 | NFR-02 | TODO |
| ECC-WS-414 | Edge | Subscribe then immediate hub shutdown (`stopping`) | clean close; no send-on-closed panic | Unit | P1 | ECC-NSW-025 sibling | TODO |
| ECC-WS-415 | Edge | WS message type unknown / empty type field at buffer-friendly small payload | ignore or close policy; hub stable | API | P2 | protocol.go | TODO |

---

## 3. Notification — `ECC-NOTIF-400+`

> Фокус: **page/per_page clamp**, **tgLinkCodeTTL=5m**, **WebPush TTL=60s**, **TG/Push spacing**, **title/body size**, **is_read filter pages**.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-NOTIF-400 | Edge | List `per_page`: 0 / 1 / 20 / 100 / 101 / negative | 0/neg/101→clamp 20; 1..100 OK | API | P0 | handler List per_page | TODO |
| ECC-NOTIF-401 | Edge | List `page`: 0 / 1 / last page / last+1 / huge page | 0→1; last has items; last+1→`[]` total intact; huge no 500 | API | P1 | ListByUser offset | TODO |
| ECC-NOTIF-402 | Edge | List exactly `per_page` total items (boundary page fill) | one full page; has next false or page2 empty | API | P1 | pagination | TODO |
| ECC-NOTIF-403 | Edge | Filter `is_read=true` with 0 read / all read; `is_read=false` with 0 unread | `[]` 200; unread-count consistent | API | P1 | is_read query | TODO |
| ECC-NOTIF-404 | Edge | Filter `is_read` malformed (`yes`/`2`) | ignore filter (nil) or 400 per policy; no 500 | API | P2 | ParseBool | TODO |
| ECC-NOTIF-405 | Edge | Telegram **link code** TTL: confirm at `5m−1s` vs `5m+1s` | −1s OK; +1s expired (отдельно от bind token 10m) | API | P0 | tgLinkCodeTTL=5m | TODO |
| ECC-NOTIF-406 | Edge | Link code: length/charset boundary (too short / too long / non-digit if digits-only) | reject; no Redis pollution | API | P1 | generateLinkCode | TODO |
| ECC-NOTIF-407 | Edge | Web Push payload `TTL=60`: deliver within 60s / after 60s offline device | within→delivered/queued; after→drop or provider expiry; no crash | Manual | P1 | push.go TTL:60 | Manual |
| ECC-NOTIF-408 | Edge | Web Push worker spacing: burst N endpoints at 50ms floor | ≥50ms between sends; ~20/s cap; no stampede | Unit | P1 | push rate 50ms | TODO |
| ECC-NOTIF-409 | Edge | Telegram worker spacing: burst N chats at 35ms floor | ≥35ms; ~28/s; queue drains without drop storm | Unit | P1 | telegram.go 35ms | TODO |
| ECC-NOTIF-410 | Edge | Notification title/body: empty title / 1 char / very long title+body near DB limits | empty→400 or default; max OK; overflow→400/truncate documented | API | P1 | CreateNotification | TODO |
| ECC-NOTIF-411 | Edge | Mark-read single: already read / unread→read; unread-count 1→0 / 0→0 | idempotent; count never negative | API | P1 | TC-NOTIF-03 | TODO |
| ECC-NOTIF-412 | Edge | Web Push subscribe count: 0 / 1 / many devices same user | all get push or capped; unsubscribe one leaves others | API | P1 | TC-NOTIF-07 | TODO |
| ECC-NOTIF-413 | Edge | Web Push endpoint URL length near max + query string | ValidateOutboundURL still applied; oversize→reject | API | P0 | NF-07 | TODO |
| ECC-NOTIF-414 | Edge | Quiet hours + page list: deferred notifs appear in in-app list immediately or only after deliver | documented; list/count consistent with preference | API | P1 | quiet hours in-app | TODO |
| ECC-NOTIF-415 | Edge | Unread-count API under empty inbox vs exactly 1 vs after mark-all | 0 / 1 / 0; never null | API | P2 | UnreadCount | TODO |

---

## 4. Search — `ECC-SRCH-400+`

> Фокус: **per-shard `defaultSearchLimit=10`**, **score merge/ties**, **saved filter name/config**, **share**, **JQL boundaries**, **types×limit**.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-SRCH-400 | Edge | Per-shard hit cap: each source returns exactly 10; client `limit=50` | merge ≤50; no shard returns >10 raw; ranking global | API | P0 | defaultSearchLimit=10 | TODO |
| ECC-SRCH-401 | Edge | Per-shard: one type has ≫10 matches, others 0; `limit=5` | top-5 by score from that type; no underflow panic | API | P0 | Search parallel | TODO |
| ECC-SRCH-402 | Edge | Score ties: equal scores across issue/message/user | stable secondary order (id/type); deterministic pages | API | P1 | sort by score | TODO |
| ECC-SRCH-403 | Edge | Score boundary: 0.0 hits / negative rank anomaly / NaN from backend | filtered or ranked last; no JSON NaN break | API | P1 | SearchResult.Score | TODO |
| ECC-SRCH-404 | Edge | Types filter: single type with `limit=1` / all types with `limit=1` | exactly 1 result; best global score wins for all-types | API | P1 | SearchRequest.Types | TODO |
| ECC-SRCH-405 | Edge | Types: duplicate type strings / empty string in array | dedup/ignore; no double-count shard | API | P2 | types parse | TODO |
| ECC-SRCH-406 | Edge | Saved filter `name`: 0 / 1 / 200 / 201 runes | 0→400; 1..200 OK; 201→400 | API | P0 | SavedFilterCreateRequest max=200 | TODO |
| ECC-SRCH-407 | Edge | Saved filter `filter_config`: `{}` / missing / huge nested map near body limit | `{}` policy; missing→400; huge→400/413; no OOM | API | P1 | filter_config required | TODO |
| ECC-SRCH-408 | Edge | Saved filter list: 0 filters / only owned / owned+shared from project | `[]` or union; shared visible to project members only | API | P1 | ListSavedFilters | TODO |
| ECC-SRCH-409 | Edge | Share toggle: shared=false→true→false; non-owner share attempt | owner OK; non-owner→403; list updates | API | P1 | ShareFilter | TODO |
| ECC-SRCH-410 | Edge | Delete saved filter: last filter / already deleted id | 200 then empty list; second→404 | API | P2 | DeleteSavedFilter | TODO |
| ECC-SRCH-411 | Edge | JQL: single condition / max practical AND chain / empty after trim | empty→400; long chain OK or 400; no panic | API | P1 | ParseJQL | TODO |
| ECC-SRCH-412 | Edge | JQL ORDER BY: allowed field / unknown field / missing direction | unknown→400; default ASC/DESC documented | API | P1 | JQLOrderBy | TODO |
| ECC-SRCH-413 | Edge | JQL value boundaries: empty quotes `""` / very long string value / `currentUser()` | empty per field policy; long safe; currentUser resolves to caller | API | P1 | ToEntPredicates | TODO |
| ECC-SRCH-414 | Edge | Global search + in-channel message search same `q`: membership scope both paths | no foreign channel hits either path (NF-03) | API | P0 | NF-03, SearchMessages | TODO |
| ECC-SRCH-415 | Edge | Ctrl+K: `limit` omitted vs `limit=50` vs results exactly 50 | default 20; 50 max fill; UI shows capped set | E2E | P2 | US-M07 | Manual |

---

## Счётчик

| Домен | ID range | Кол-во Edge | P0 | P1 | P2 |
|-------|----------|-------------|----|----|-----|
| Messenger | ECC-MSG-400…429 | **30** | 12 | 13 | 5 |
| WS | ECC-WS-400…415 | **16** | 7 | 7 | 2 |
| Notification | ECC-NOTIF-400…415 | **16** | 4 | 10 | 2 |
| Search | ECC-SRCH-400…415 | **16** | 5 | 8 | 3 |
| **Итого** | | **78** | **28** | **38** | **12** |

> Минимум 50 — **выполнено (78)**.  
> Happy / Neg / Corner — **не включены**.  
> ID `*100+` / `*200+` / `*300+` / каталог `*001…` — **не пересекались** (не копии).  
> `.planning/july-2026/EDGE-CORNER-CATALOG.md` — **не изменялся**.  
> Коммит — **не создавался**.
