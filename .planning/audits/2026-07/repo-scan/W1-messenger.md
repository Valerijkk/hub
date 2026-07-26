# W1-messenger — BE Audit `backend/internal/messenger/`

> **Provenance class:** PRE_EXISTING · **Confidence:** MIXED · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> **Роль:** be-ver (Grok 4.5, readonly)
> **Дата:** 2026-07-23
> **Scope:** Mongo messages, membership, rate limits, cursor pagination, NATS
> **Код не менялся** · регрессия `go test` не запускалась (repo-scan / code freeze)

---

## Summary

Send / List / Forward / threads / GetChannel / ListMembers / ListPins держат membership (HI-01/HI-02). Mongo: soft-delete в List/Search (`deleted_at: nil`), индексы `(channel_id, created_at)` + text. Константа RL **60/мин** совпадает с TZ.

Блокеры: **`/messenger/search` без user-scope (NF-03 CRITICAL)** — в отличие от `/api/search` (W2-notification-search PASS для messages); **cross-channel DeleteMessage**; **GetPollResults без IsMember**; gaps на reaction/pin/bookmark; **MessageSendRateLimit до auth** → ключ `ip:`; NATS **fire-and-forget** без outbox.

**Вердикт:** `verified: false` — CRITICAL/HIGH на access control и RL.

---

## Findings

| Severity | file:line | Issue | Evidence | Fix |
|----------|-----------|-------|----------|-----|
| **CRITICAL** | `service_search.go:11-29` + `handler_search.go:24-37` + `mongo_repository.go:276-304` | **NF-03:** `SearchMessages` не принимает `userID`, не фильтрует по членству. Без `channel_id` — `$text` по **всем** сообщениям инстанса (DM/private); с чужим `channel_id` — чтение чужого канала | Handler: `query`/`channel_id`/`limit` только; service → `repo.SearchMessages` без scope. Контраст: `search/service.go` messages — `channel_id ∈ getUserChannelIDs` | Прокинуть `userID`; без channel — `$in` channel_ids пользователя; с channel — `IsMember` до query. Denial-тест: non-member → 403/пусто |
| **CRITICAL** | `service_messages.go:159-177` + `handler_messages.go:158` | **Cross-channel delete IDOR:** для не-автора роль берётся из **path** `channelId`, soft-delete — по произвольному `messageId` без `existing.ChannelID == path` | Admin канала A: `DELETE .../channels/{A}/messages/{msgИзB}` → `GetMemberRole(A)` ok → `SoftDeleteMessage(msgB)` | Требовать `existing.ChannelID == channelID`; роль по `existing.ChannelID`. Тест: admin A ≠ delete msg B |
| **CRITICAL** | `service_polls.go:202-214` | **GetPollResults** без `IsMember` — любой auth читает вопрос/голоса/`VoterIDs` (non-anonymous) чужого канала по UUID опроса | Create/Vote — membership; Get — сразу `GetPollWithOptions` + `ListVotes` | `IsMember(p.ChannelID, currentUserID)` до выдачи → 403 |
| **HIGH** | `service_reactions.go:13-43` + `handler_reactions.go:29-49` | **ToggleReaction** без membership; path `channelId` в handler не парсится | `FindMessageByID` → Add/Remove → NATS; нет `IsMember(msg.ChannelID)` | Membership до мутации; сверить path channel с `msg.ChannelID` |
| **HIGH** | `service_pins.go:12-27` | **TogglePin** без membership (у `ListPins` HI-01 есть) — pin/unpin в чужом канале | Сразу FindPin/AddPin/RemovePin по path `channelID` | `IsMember` (+ роль owner/admin, если продукт так требует) |
| **HIGH** | `service_bookmarks.go:19-21` + `ent_repository.go:323-325` + `handler_reactions.go:190-200` | **RemoveBookmark** по UUID без владельца | `DeleteOneID(bookmarkID)`; handler не передаёт `userID` | `Where(IDEQ, UserIDEQ)` → 404 чужому |
| **HIGH** | `service_bookmarks.go:11-16` | **AddBookmark** без `IsMember` / сверки message∈channel | Только create bookmark | `IsMember` + `FindMessageByID` + `msg.ChannelID == channelID` |
| **HIGH** | `service_channel_invites.go:207-223` + `ent_repository.go:441-445` | **RevokeInvite** IDOR: роль на path `channelID`, revoke по произвольному `inviteID` без `invite.ChannelID == channelID` | Admin A отзывает invite канала B при известном UUID | Загрузить invite; mismatch channel → 404; revoke только своего канала |
| **HIGH** | `wiring_fiber.go:58-59` + `middleware/ratelimit.go:78-106` + `auth/middleware.go:94` | **MessageSendRateLimit** на `api.Use("/messenger")` **до** `authMw` в `RegisterRoutes` → `Locals("userID")` пуст → почти всегда `ip:…`, не per-user | Order: wireFiber RL → позже Group `/messenger` + `requireAuth` sets Locals | RL после auth (на group) или только `POST …/messages` с user key |
| **HIGH** | `service_messages.go:84-93` (+ polls/channels/…) + `platform/nats/client.go:118-132` | NATS **без outbox**, `PublishJSON` fire-and-forget: Mongo/Ent OK, publish error только slog → silent loss realtime | Все `s.nats.PublishJSON` игнорируют ошибку; нет `PublishInTx` | Outbox для durable **или** retry/metric; typing — ephemeral OK |
| **MEDIUM** | `service.go:81-97` + `handler_presence.go:23-37` | **PublishTyping** без membership — спам NATS subject чужого канала | Нет `IsMember` перед publish | Membership → 403 |
| **MEDIUM** | `mongo_repository.go:93-122`, `174-202` | Cursor только `created_at $lt` — при равных timestamp skip/duplicate; нет tie-break `_id` | `before *time.Time` + sort `-created_at` + limit | Compound cursor `(created_at, _id)`; `next_before` в ответе |
| **MEDIUM** | `wiring_fiber.go:58-59` | RL 60/мин на **весь** `/messenger` (GET list/search/presence/heartbeat), не только send | Comment «отправки», middleware на prefix | Раздельные read/write лимиты; send только на POST messages |
| **MEDIUM** | `handler.go:116-117` | `POST /api/messages/:id/promote-to-issue` **вне** `/messenger` → вне MessageSendRateLimit | Отдельная группа `api.Group("/messages")` | Тот же user RL / отдельный cap |
| **MEDIUM** | `mongo_repository.go:126-153` | `UpdateMessage`/`SoftDeleteMessage`: `time.Now()` без UTC; `FindMessageByID` отдаёт soft-deleted (edit «удалённого» возможен) | `now := time.Now()`; Find без `deleted_at` | `UTC()`; reject edit/delete уже удалённых |
| **MEDIUM** | `service_messages.go:230-285` | `ForwardMessage`: path `channelID` не сверяется с `source.ChannelID` (membership source/target OK) | Первый аргумент не используется для bind | Assert path == source.ChannelID → 404 |
| **MEDIUM** | `service_channel_invites.go:178-195` + `ent_repository.go:430-438` | Accept: `IncrementInviteUses` = read-modify-write без `use_count < max` в Tx; NATS **`via_invite: token`** утекает подписчикам | NOTE в коде; `SetUseCount(inv.UseCount+1)`; publish включает token | Conditional `AddUseCount(1)` + WHERE; в событии не слать raw token |
| **MEDIUM** | `service_channels.go:36-40` | CreateChannel: fail `AddMembers(owner)` → slog, orphan channel без owner | Нет rollback | Tx create+owner; при fail — error/delete |
| **MEDIUM** | `mongo_repository.go:205-238` | Race AddReaction: concurrent first-react → дубликат emoji-entry | `$addToSet` на `$` → ModifiedCount 0 → `$push` | Upsert/`$elemMatch` atomic pattern |
| **MEDIUM** | `service_channels.go:14-30` | CreateChannel: `ProjectID` без проверки membership в проекте | Parse UUID → CreateChannel | `IsProjectMember` / RBAC до create |
| **LOW** | `events.go:14-46` | Subjects `messenger.channel.%s.*`, не `hub.messages.*` (rule 20) — drift конвенции | Константы Subject* | Документировать канон или выровнять prefix |
| **LOW** | `service_polls.go:66-88` | Poll в PG OK, Mongo wrapper fail → опрос не в ленте (soft-fail) | WARN + return poll | Retry/compensation; флаг в API |
| **LOW** | `service.go:113-117` | MarkChannelRead без IsMember — только свой Redis hash (низкий риск) | `unread.MarkRead(userID, channelID)` | Опционально IsMember |
| **PASS** | `service_messages.go:16-25,108-117,201-217,237-257` | Send/List/Thread/Forward: membership (source+target HI-02) | IsMember до R/W | Сохранить |
| **PASS** | `service_channels.go:64-73,154-172,267-276` + `service_pins.go:29-38` | GetChannel / Join private deny / ListMembers / ListPins | HI-01 | — |
| **PASS** | `mongo_repository.go:37-68,99-103` | Индексы + soft-delete filter в List | EnsureIndexes | — |
| **PASS** | `service.go:18-20` + `ent_repository.go` / `combined_repository.go` | Layering: `*ent.Client` только EntRepository; service → Repository | Нет Client в service/handler | — |
| **PASS** | `middleware/ratelimit.go:31` | Число 60/min = TZ (проблема — wiring/порядок, не константа) | MessageSendRateLimit | Починить order/scope |
| **PASS** | `handler.go:132-143` | `utils.CopyString` на path params (fasthttp) | parseChannelID | — |

---

## Checklist (фокус задачи)

| Контроль | Статус | Комментарий |
|----------|--------|-------------|
| Mongo soft-delete + indexes | **PASS\*** | List/Search OK; FindByID без фильтра deleted |
| Cursor pagination | **PARTIAL** | `before`+limit; нет tie-break `_id` |
| Rate limit 60/min (user) | **FAIL** | Число OK; ключ IP (middleware до auth) + весь prefix |
| NATS JetStream/events | **PARTIAL** | Subjects/payloads есть; core Publish F&F, нет outbox; token leak invite |
| Membership read/write | **FAIL** | Core OK; Search/Poll GET/Reaction/Pin toggle/Bookmark/Revoke invite — дыры |
| Layering H→S→R | **PASS** | — |

\*FindByID / edit path — soft-deleted ещё доступны.

---

## 5 top risks

1. **Search без scope (CRITICAL)** — любой сотрудник читает DM/private через `GET /api/messenger/search`.
2. **Delete cross-channel (CRITICAL)** — admin любого канала удаляет сообщения чужих каналов по UUID.
3. **Poll results IDOR (CRITICAL)** — утечка содержимого/голосов private каналов.
4. **Reaction/Pin/Bookmark/Revoke gaps (HIGH)** — мутации + existence/revoke oracle.
5. **RL не per-user + весь prefix (HIGH)** — TZ 60/min фактически per-IP; GET сжигает send-бюджет.

---

## Verification

| Этап | Результат |
|------|-----------|
| Scope files | `handler*`, `service*`, `mongo/ent/combined` repo, `events`, `unread`, `auto_channels`, `wiring_fiber`/`wiring_messenger`, `ratelimit`, `nats/client` |
| Grep membership gaps | Search / Reaction / Pin toggle / Bookmark / Typing / GetPollResults / RevokeInvite — confirmed |
| DeleteMessage channel bind | path role ≠ `message.channel` — confirmed |
| RL vs auth order | `api.Use("/messenger", RL)` в wireFiber **до** `Group(..., requireAuth)` — confirmed |
| Регрессия `-race` | **не запускалась** (readonly repo-scan) |
| Docs-sync | N/A (код не менялся) |

---

## Верификация (BE): messenger

- Воспроизведено: статически по цепочкам Handler→Service→Repo (search / delete / poll GET / RL order)
- Layering: OK — `*ent.Client` только `ent_repository.go`
- Security: FAIL — NF-03 search, delete IDOR, poll GET, reaction/pin/bookmark/revoke
- Concurrency: PARTIAL — invite use_count RMW; reaction race; NATS F&F
- Регрессия: не гонялась (scan-only)
- Вердикт: **verified: false** — 2026-07-23 01:10 UTC+3

---

## Handoff

| Owner | Действие |
|-------|----------|
| be1 (поток A) | CRITICAL/HIGH: Search scope, DeleteMessage bind, GetPollResults, Reaction/Pin/Bookmark, RevokeInvite |
| devops/be | MessageSendRateLimit после auth; сузить на POST messages (+ promote path) |
| qa-auto | Denial-suite: search foreign, delete cross-channel, poll GET, reaction/pin/bookmark/revoke |
| team-lead | Не считать messenger membership-hardened до закрытия CRITICAL |

**Stop.** Код не менялся.
