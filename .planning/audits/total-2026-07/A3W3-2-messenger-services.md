# A3 W3 агент 2 — messenger services (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Baseline: [A2W3-2-messenger-services.md](A2W3-2-messenger-services.md) · Stage A: [A3-2-messenger-services.md](A3-2-messenger-services.md)  
> Remediation: TASK-010 (promote RBAC) · TASK-108 (invite/DM/bot) · TASK-017/042 (membership/denial)  
> Периметр: 28 файлов (14 prod + 14 tests) · Прочитано: **28/28** · Пропущено: **0**

## Сверка с A2 (A2W3-2-* / W3-A2-*)

| A2 ID | Tag | Статус в HEAD | Комментарий |
|-------|-----|---------------|-------------|
| A2W3-2-001 | FIXED-OK | закрыт | `IncrementInviteUses(…, maxUses)` + `UseCountLT` + `DecrementInviteUses` при fail AddMembers (`service_channel_invites.go:183-197`; repo `ent_repository.go:524-557`) |
| A2W3-2-002 | FIXED-OK | закрыт | CreateChannel/DM: fail AddMembers(creator/pair) → `DeleteChannel` + error (`service_channels.go:37-44`; `service_dm.go:41-48`) |
| A2W3-2-003 | FIXED-OK | закрыт | `systemBotID == Nil` guard на Find/GetChannel/MemberIDs/List*ForBot (`service_bot.go:131-178`); тест `TestServiceBot_NotConfiguredErrors` |
| A2W3-2-004 | FIXED-OK | закрыт | AcceptInvite: `IsMember` err → 500 (`service_channel_invites.go:169-173`) |
| W3-A2-001 P0 | FIXED-OK | закрыт | `promoteCheck` вызывается; nil → 403; wiring `SetPromotePermissionChecker` (`service_promote.go:129-139`; `wiring_messenger.go:76`); denial `handler_helpers_test.go` |
| W3-A2-002 | FIXED-OK | закрыт | DeleteMessage: `existing.ChannelID == path` + роль по msg channel (`service_messages.go:174-194`) |
| W3-A2-003/004/005/006/007 | FIXED-OK | закрыт | Reaction/Pin/PollResults/Bookmarks membership+owner |
| W3-A2-008 | FIXED-OK | закрыт | `RevokeChannelInvite(channelID, inviteID)` (`service_channel_invites.go:227-228`) |
| W3-A2-009 | FIXED-OK | закрыт | атомарный increment UseCountLT (см. A2W3-2-001) |
| W3-A2-010 | FIXED-OK | закрыт | NATS `via_invite:true` + `invite_id`, без raw token (`:199-205`) |
| W3-A2-011 | FIXED-OK | закрыт | см. A2W3-2-002 (orphan create) |
| W3-A2-012 | FIXED-OK | закрыт | last owner Leave → 400 (`service_channels.go:218-238`) |
| W3-A2-013 | FIXED-OK | закрыт | Mentions ∩ memberSet (`service.go:192-199`) |
| W3-A2-014 | FIXED-OK | закрыт | Forward unread++ (`service_messages.go:297-308`) |
| W3-A2-015 | FIXED-OK | закрыт | bot internal read guards (A2W3-2-003) |
| W3-A2-016 | FIXED-OK | закрыт | denial promote/pin/list; reaction membership в prod |
| W3-A2-017 | FIXED-OK | закрыт | `truncateRunes(title, 500)` (`service_promote.go:107`; `truncate.go`) |
| W3-A2-018 | FIXED-OK | закрыт | Forward `source.ChannelID == channelID` (`service_messages.go:265-268`) |

**REGRESS:** 0

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/messenger/service.go` | 282 | 0 | FIXED-OK mentions∩members; presence cap+FilterUsersSharingChannel |
| `backend/internal/messenger/service_bookmarks.go` | 57 | 1 | FIXED-OK IsMember/owner Remove; NEW msg∉channel |
| `backend/internal/messenger/service_bot.go` | 189 | 0 | FIXED-OK systemBotID на всех bot/internal reads |
| `backend/internal/messenger/service_channel_invites.go` | 279 | 0 | FIXED-OK accept/uses/revoke/token |
| `backend/internal/messenger/service_channels.go` | 349 | 2 | FIXED-OK create-owner/leave-owner; NEW Join IsMember err; Create extra members |
| `backend/internal/messenger/service_dm.go` | 51 | 0 | FIXED-OK AddMembers fail → DeleteChannel |
| `backend/internal/messenger/service_hook.go` | 22 | 0 | чисто |
| `backend/internal/messenger/service_messages.go` | 321 | 0 | FIXED-OK Delete/Edit/Forward/unread |
| `backend/internal/messenger/service_pins.go` | 70 | 0 | FIXED-OK IsMember + msg∈channel |
| `backend/internal/messenger/service_polls.go` | 278 | 1 | FIXED-OK GetPollResults membership; NEW ClosePoll order |
| `backend/internal/messenger/service_promote.go` | 319 | 0 | FIXED-OK promoteCheck fail-closed + truncateRunes |
| `backend/internal/messenger/service_reactions.go` | 57 | 0 | FIXED-OK IsMember + path==msg.ChannelID |
| `backend/internal/messenger/service_search.go` | 50 | 0 | чисто (A2001 empty-scope) |
| `backend/internal/messenger/auto_channels.go` | 174 | 0 | чисто (lifecycleCtx) |
| `backend/testframework/messenger/auto_channels_service_test.go` | 150 | 0 | чисто |
| `backend/testframework/messenger/auto_channels_test.go` | 180 | 0 | чисто |
| `backend/testframework/messenger/service_bot_test.go` | 159 | 0 | FIXED-OK NotConfiguredErrors |
| `backend/testframework/messenger/service_extended_test.go` | 513 | 0 | чисто (happy bookmark; pin denial ListPins) |
| `backend/testframework/messenger/service_fanout_test.go` | 67 | 0 | чисто |
| `backend/testframework/messenger/service_mentions_test.go` | 170 | 0 | чисто |
| `backend/testframework/messenger/service_messages_test.go` | 156 | 0 | чисто |
| `backend/testframework/messenger/service_search_a2001_test.go` | 45 | 0 | чисто |
| `backend/testframework/messenger/service_test.go` | 499 | 0 | чисто |
| `backend/testframework/messenger/service_wave114_test.go` | 56 | 0 | чисто |
| `backend/testframework/messenger/service_wave115_test.go` | 145 | 0 | чисто |
| `backend/testframework/messenger/service_wave116_test.go` | 105 | 0 | чисто |
| `backend/testframework/messenger/service_wave123_test.go` | 106 | 0 | FIXED-OK promoteCheck wiring |
| `backend/testframework/messenger/delete_channel_cascade_a8005_test.go` | 69 | 0 | чисто |

Файлов в списке: **28**. Прочитано: **28**. Пропущено: **0**.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|----------|-------------------|------------|
| A3W3-2-001 | NEW | P2 | `service_channels.go:200` | `JoinChannel`: `isMember, _ := s.repo.IsMember` — err глотается | Зеркало A2W3-2-004 (Accept): при DB err `isMember=false` → попытка AddMembers / ложный успех/шум вместо 500 | Как в AcceptInvite: `if err != nil { return NewInternal(...) }` |
| A3W3-2-002 | NEW | P2 | `service_channels.go:58-62` | После успешного AddMembers(owner) fail доп. `MemberIDs` → 500, канал **остаётся** | Клиент считает create проваленным; `ListMyChannels` уже видит канал → retry → дубликаты. Orphan-owner (A2W3-2-002) закрыт; partial members — нет | При fail extra: DeleteChannel+error **или** вернуть канал + partial warning; лучше tx/rollback |
| A3W3-2-003 | NEW | P2 | `service_bookmarks.go:13-26` | `AddBookmark` проверяет IsMember(channelID), **не** `msg.ChannelID == channelID` | В отличие от TogglePin (`service_pins.go:22-28`): member канала B может закрепить bookmark на messageID из канала A с `channelID=B` → мусор метаданных / cross-channel bookmark | После IsMember: `FindMessageByID` + `msg.ChannelID == channelID` иначе 404; denial-тест |
| A3W3-2-004 | NEW | P3 | `service_channels.go:288-312` | ToggleStar/Mute/Unmute: repo NotFound (не member) → `NewInternal` 500 | Ent/mock `ChannelMember.Only` для non-member → NotFound; сервис мапит в Internal. Контракт соседних методов — 403 Forbidden | `ent.IsNotFound` → Forbidden/NotFound; опц. явный IsMember до repo |
| A3W3-2-005 | NEW | P3 | `service_polls.go:167-190` | `ClosePoll`: `ClosePoll` в repo, затем `GetPollResults` (membership) | Creator, покинувший канал: опрос уже `IsClosed`, клиент получает 403 от GetPollResults — inconsistent success | Проверить IsMember/creator до Close; или вернуть DTO без GetPollResults при Forbidden после close |
| A3W3-2-006 | NEW | P3 | `service_channels.go:279`, `:176` | `AddMembers` / `DeleteChannel` возвращают сырой `repo` err | Non-AppError (Ent) может утечь без `{"error","request_id"}` wrap | Wrap → AppError; AppError пробрасывать |

## Чисто

- Promote fail-closed + wiring RBAC (TASK-010) — без регрессии; denial `allowed:false` / nil checker покрыты.
- Invite accept uses/revoke/token-NATS, Create/DM orphan cleanup, bot guards — FIXED-OK (TASK-108).
- Delete/Edit/Forward channel-scope, Reaction/Pin/PollResults/Bookmarks owner+member — FIXED-OK.
- Mentions ∩ members, Forward unread, last-owner Leave, Search A2001 empty-scope, cascade soft-delete — без регрессии.
- Auto-channels lifecycleCtx; presence FilterUsersSharingChannel + cap 50.

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 3 |
| P3 | 3 |
| **Total** | **6** |

A2 baseline: open заявлены закрытыми TASK-108/010 → FIXED-OK **18** (A2W3-2-001…004 + W3-A2-001…018) · REGRESS **0** · RESIDUAL **0** · NEW **6**.
