# W3 агент A2 — messenger services

**Дата:** 2026-07-28  
**Режим:** READ-ONLY (кроме этого отчёта)  
**Задача:** TASK-008 / волна W3 / агент A2  
**Фокус:** membership checks, DM/invite, cascade delete, fanout, search empty-scope, bot permissions

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/messenger/service.go` | 245 | 1 | mentions: individ. без member-filter; typing OK |
| `backend/internal/messenger/service_bookmarks.go` | 40 | 2 | Add без membership; Remove без ownership |
| `backend/internal/messenger/service_bot.go` | 180 | 1 | internal helpers без systemBotID-guard |
| `backend/internal/messenger/service_channel_invites.go` | 269 | 3 | Revoke IDOR; Accept race + token в NATS |
| `backend/internal/messenger/service_channels.go` | 296 | 2 | Create swallow AddMembers; Leave без last-owner |
| `backend/internal/messenger/service_dm.go` | 47 | 1 | AddMembers ошибка глотается |
| `backend/internal/messenger/service_hook.go` | 22 | 0 | чисто (интерфейс hook) |
| `backend/internal/messenger/service_messages.go` | 286 | 2 | DeleteMessage channelID-mismatch; Forward без unread |
| `backend/internal/messenger/service_pins.go` | 55 | 1 | TogglePin без IsMember |
| `backend/internal/messenger/service_polls.go` | 269 | 1 | GetPollResults без membership |
| `backend/internal/messenger/service_promote.go` | 311 | 2 | promoteCheck мёртв; byte-truncate title |
| `backend/internal/messenger/service_reactions.go` | 44 | 1 | ToggleReaction без IsMember |
| `backend/internal/messenger/service_search.go` | 50 | 0 | чисто (A2-001 + empty-scope) |
| `backend/internal/messenger/auto_channels.go` | 174 | 0 | чисто (lifecycleCtx, AddMembersInternal) |
| `backend/testframework/messenger/auto_channels_service_test.go` | 150 | 0 | чисто |
| `backend/testframework/messenger/auto_channels_test.go` | 180 | 0 | чисто (simulate path) |
| `backend/testframework/messenger/service_bot_test.go` | 151 | 0 | чисто (bot sender guard на Update) |
| `backend/testframework/messenger/service_extended_test.go` | 501 | 1 | pin/bookmark happy-path без denial |
| `backend/testframework/messenger/service_fanout_test.go` | 67 | 0 | чисто (unread excludes sender) |
| `backend/testframework/messenger/service_mentions_test.go` | 170 | 0 | чисто (@all/@role; delete auth) |
| `backend/testframework/messenger/service_messages_test.go` | 156 | 0 | чисто (edit/list/forward/search/DM) |
| `backend/testframework/messenger/service_search_a2001_test.go` | 45 | 0 | чисто (non-member + empty scope) |
| `backend/testframework/messenger/service_test.go` | 504 | 1 | ToggleReaction без membership denial |
| `backend/testframework/messenger/service_wave114_test.go` | 56 | 0 | чисто (revoke/list invite RBAC) |
| `backend/testframework/messenger/service_wave115_test.go` | 145 | 0 | чисто (accept invite edges) |
| `backend/testframework/messenger/service_wave116_test.go` | 105 | 0 | чисто (update/delete channel roles) |
| `backend/testframework/messenger/service_wave123_test.go` | 106 | 1 | promoteCheck только setter, не вызов |
| `backend/testframework/messenger/delete_channel_cascade_a8005_test.go` | 69 | 0 | чисто (soft-delete A8-005) |

Файлов в списке: 28. Прочитано: 28. Пропущено: 0.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|----------|-------------------|------------|
| W3-A2-001 | P0 | `service_promote.go:56-62`, `90-174` (+ wiring: `SetIssuePromoter` без `SetPromotePermissionChecker`) | `promoteCheck` задаётся сеттером, но **нигде не вызывается** в `PromoteToIssue`; `CreateIssueFromPromotion` (tracker) явно без RBAC; маршрут только `authMw` | Любой member канала может создать задачу в **любом** `project_id` (PrivEsc / cross-project write). Комментарий контракта врёт про проверку `PermTaskCreate` | Перед `CreateIssueFromPromotion`: если `promoteCheck != nil` → `CanCreateTaskInProject` иначе 403; в `wiring_messenger.go` обязательно `SetPromotePermissionChecker`; fail-closed при nil checker |
| W3-A2-002 | P1 | `service_messages.go:161-177` | `DeleteMessage`: роль owner/admin берётся по `channelID` из аргумента (path), а не по `existing.ChannelID` | Owner канала A удаляет сообщение канала B: `DELETE .../channels/{A}/messages/{msgB}` → GetMemberRole(A)=owner → soft-delete чужого сообщения + NATS delete в room B | Сверять `uuid.Parse(channelID)==existing.ChannelID` (иначе 404); роль только по `existing.ChannelID` |
| W3-A2-003 | P1 | `service_reactions.go:13-41` | `ToggleReaction` без `IsMember(msg.ChannelID, userID)` | Non-member ставит реакцию и публикует NATS в room канала (IDOR + fanout spam) | После `FindMessageByID`: `IsMember` → 403; затем mutate |
| W3-A2-004 | P1 | `service_pins.go:12-26` | `TogglePin` без membership (ListPins — с HI-01) | Non-member закрепляет/открепляет; pin чужого `messageID` в свой `channelID` | `IsMember` + проверка, что сообщение ∈ канала |
| W3-A2-005 | P1 | `service_polls.go:202-214` | `GetPollResults` без membership; для неанонимных опросов отдаёт `VoterIDs` | Утечка вопроса/голосов/UUID голосующих по `pollId` любому auth-пользователю | `IsMember(p.ChannelID, currentUserID)` → 403 |
| W3-A2-006 | P1 | `service_bookmarks.go:19-21` | `RemoveBookmark(bookmarkID)` без `userID` | Зная UUID закладки, чужой пользователь удаляет чужую закладку (IDOR) | `RemoveBookmark(ctx, userID, bookmarkID)` + repo `Where(ID, UserID)` |
| W3-A2-007 | P1 | `service_bookmarks.go:11-16` | `AddBookmark` без `IsMember(channelID, userID)` | Закладка на сообщение/канал, к которому нет доступа (метаданные + мусор в списке) | `IsMember` (+ опц. сверка message∈channel) |
| W3-A2-008 | P1 | `service_channel_invites.go:207-223` | `RevokeInvite`: admin канала A, `inviteID` чужого канала — `RevokeChannelInvite(inviteID)` без `ChannelIDEQ` | Cross-channel revoke invite (IDOR) | Repo: `Where(ID, ChannelID)`; mismatch → 404 |
| W3-A2-009 | P2 | `service_channel_invites.go:178-188` | Accept: `IncrementInviteUses` затем `AddMembers` не в одной tx; repo increment — read-modify-write | Гонка >MaxUses; fail AddMembers после increment «сжигает» слот | Атомарный `WHERE use_count < max_uses`; rollback/декремент при fail AddMembers |
| W3-A2-010 | P2 | `service_channel_invites.go:191-195` | NATS `member.joined` содержит `"via_invite": token` | Токен invite в шине событий — лишний носитель секрета для подписчиков subject | Публиковать `invite_id` / флаг `via_invite:true` без raw token |
| W3-A2-011 | P2 | `service_channels.go:37-40`; `service_dm.go:41-44` | Ошибка `AddMembers` только `slog.Error` — канал/DM возвращается клиенту | Orphan channel/DM без members (неуправляемый канал; DM без второго участника) | Fail create при ошибке members; или tx + cleanup канала |
| W3-A2-012 | P2 | `service_channels.go:193-204` | `LeaveChannel` без проверки роли / последнего owner | Owner уходит → канал без owner (зависает управление) | Запрет leave единственному owner; transfer или delete |
| W3-A2-013 | P2 | `service.go:147-152` | Индивидуальные `Mentions` попадают в `resolvedMentions` без пересечения с `memberSet` | Упоминания/нотификации UUID вне канала (spam / leak intent) | Фильтровать individual mentions через `memberSet` |
| W3-A2-014 | P2 | `service_messages.go:230-285` | `ForwardMessage` пишет сообщение + NATS, **не** инкрементирует unread (в отличие от `SendMessage:67-78`) | Получатели не видят unread badge на forward | Тот же unread fanout, что в SendMessage |
| W3-A2-015 | P2 | `service_bot.go:131-148` | `FindMessageForBot` / `GetChannelEntity` / `GetChannelMemberIDsInternal` без `systemBotID != Nil` (List*ForBot — с проверкой) | Defense-in-depth: любой holder `*MessengerService` читает канал/сообщения без bot-gate | Единый guard `systemBotID` на все bot/internal read API |
| W3-A2-016 | P2 | `service_test.go:184-208`; `service_extended_test.go:180-195`, `156-177`; `service_wave123_test.go:54-61` | Тесты зелёные на текущих дырах: reaction/pin без membership; bookmark remove без owner; promoteCheck только setter | Регрессия membership/RBAC не ловится (assert на сломанном guard) | Denial-тесты outsider→403; Promote с `allowed:false` → 403 + мутация вызова checker |
| W3-A2-017 | P3 | `service_promote.go:109-111` | `title = title[:500]` по байтам | Mid-rune truncate UTF-8 заголовка | `[]rune` / `utf8.RuneCount` truncate |
| W3-A2-018 | P3 | `service_messages.go:231` | Параметр `channelID` в `ForwardMessage` не используется для source-scope | Dead API surface; path channel игнорируется | Assert source.ChannelID == channelID или убрать параметр |

## Чисто

- **Search empty-scope / A2-001:** `service_search.go` — explicit channel → `IsMember`; global → `ListUserChannels`, при `len==0` сразу `[]` (repo с пустым scope не зовётся); тест `service_search_a2001_test.go`.
- **Cascade A8-005:** service `DeleteChannel` → repo soft-delete; тест `delete_channel_cascade_a8005_test.go` (deleted_at, IsMember false, children сохранены).
- **Fanout unread (SendMessage):** sender исключается; `service_fanout_test.go` green.
- **Membership на чтении/отправке:** GetChannel, ListMembers, ListMessages, ListThreadReplies, SendMessage, Forward (source+target), CreatePoll, VotePoll, PublishTyping, ListPins — `IsMember` / role guards на месте.
- **DM дедуп:** `dm_key` из отсортированных UUID; self-DM → 400.
- **Invite create/list:** owner/admin; Validate/Accept edges (revoked/expired/max) покрыты wave115.
- **Bot UpdateBotMessage:** чужое сообщение → Forbidden; SendBot не вызывает hook (anti-recursion).
- **Auto-channels:** lifecycleCtx вместо Background; creator filter перед AddMembersInternal.
- **Promote source membership:** `IsMember` на канале сообщения — есть; дыра — целевой проект (W3-A2-001).

### Сводка

| Severity | Count |
|----------|------:|
| P0 | 1 |
| P1 | 7 |
| P2 | 8 |
| P3 | 2 |
| **Всего** | **18** |

**Файлов:** 28/28/0  

**Топ-3:**
1. **W3-A2-001** — PromoteToIssue без RBAC целевого проекта (`promoteCheck` мёртв, wiring не ставит checker)
2. **W3-A2-002** — DeleteMessage: роль по path-channel, не по каналу сообщения (cross-channel delete)
3. **W3-A2-003/004/005** — ToggleReaction / TogglePin / GetPollResults без membership (IDOR + NATS fanout / утечка голосов)
