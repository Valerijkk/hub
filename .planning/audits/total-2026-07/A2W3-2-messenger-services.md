# A2 W3 агент 2 — messenger services (re-audit)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> Зеркало: `W3-ASSIGNMENTS.md` A2 · Stage A: `A3-2-messenger-services.md`

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/messenger/service.go` | 226 | 0 | FIXED-OK mentions∩members; presence cap |
| `backend/internal/messenger/service_bookmarks.go` | 51 | 0 | FIXED-OK IsMember + owner Remove |
| `backend/internal/messenger/service_bot.go` | 163 | 1 | RESIDUAL bot internal reads |
| `backend/internal/messenger/service_channel_invites.go` | 235 | 2 | RESIDUAL Accept race; FIXED revoke/token |
| `backend/internal/messenger/service_channels.go` | 281 | 1 | RESIDUAL Create swallow AddMembers |
| `backend/internal/messenger/service_dm.go` | 39 | 1 | RESIDUAL AddMembers swallow |
| `backend/internal/messenger/service_hook.go` | 18 | 0 | чисто |
| `backend/internal/messenger/service_messages.go` | 282 | 0 | FIXED-OK Delete/Edit/Forward |
| `backend/internal/messenger/service_pins.go` | 62 | 0 | FIXED-OK IsMember + msg∈channel |
| `backend/internal/messenger/service_polls.go` | 245 | 0 | FIXED-OK GetPollResults membership |
| `backend/internal/messenger/service_promote.go` | 281 | 0 | FIXED-OK promoteCheck fail-closed |
| `backend/internal/messenger/service_reactions.go` | 45 | 0 | FIXED-OK IsMember |
| `backend/internal/messenger/service_search.go` | 46 | 0 | чисто (A2001) |
| `backend/internal/messenger/auto_channels.go` | 153 | 0 | чисто |
| `backend/testframework/messenger/auto_channels_service_test.go` | 127 | 0 | чисто |
| `backend/testframework/messenger/auto_channels_test.go` | 153 | 0 | чисто |
| `backend/testframework/messenger/service_bot_test.go` | 125 | 0 | чисто |
| `backend/testframework/messenger/service_extended_test.go` | 437 | 0 | FIXED-OK denial gaps закрыты task017 |
| `backend/testframework/messenger/service_fanout_test.go` | 56 | 0 | чисто |
| `backend/testframework/messenger/service_mentions_test.go` | 142 | 0 | чисто |
| `backend/testframework/messenger/service_messages_test.go` | 132 | 0 | чисто |
| `backend/testframework/messenger/service_search_a2001_test.go` | 38 | 0 | чисто |
| `backend/testframework/messenger/service_test.go` | 414 | 0 | чисто |
| `backend/testframework/messenger/service_wave114_test.go` | 47 | 0 | чисто |
| `backend/testframework/messenger/service_wave115_test.go` | 124 | 0 | чисто |
| `backend/testframework/messenger/service_wave116_test.go` | 89 | 0 | чисто |
| `backend/testframework/messenger/service_wave123_test.go` | 86 | 0 | FIXED-OK promoteCheck вызывается |
| `backend/testframework/messenger/delete_channel_cascade_a8005_test.go` | 55 | 0 | чисто |

Файлов в списке: **28**. Прочитано: **28**. Пропущено: **0**.

## Находки (открытые)

| ID | Tag | Severity | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|----------|-------------|----------|---------------|------------|
| A2W3-2-001 | RESIDUAL | P2 | `service_channel_invites.go:178-188`; `ent_repository.go:466-477` | Accept: `IncrementInviteUses` (RMW) затем `AddMembers` **не** в одной tx; fail AddMembers сжигает слот; нет `UseCountLT` | W3-A2-009 / TASK-038 out-of-scope messenger — гонка >MaxUses и burned slot | Атомарный `WHERE use_count < max_uses`; rollback/декремент при fail AddMembers |
| A2W3-2-002 | RESIDUAL | P2 | `service_channels.go:37-57`; `service_dm.go:41-44` | Ошибка `AddMembers` только `slog.Error` — канал/DM возвращается клиенту | W3-A2-011: orphan channel без owner/members | Fail create / cleanup канала при ошибке members |
| A2W3-2-003 | RESIDUAL | P2 | `service_bot.go:131-148` | `FindMessageForBot` / `GetChannelEntity` / `GetChannelMemberIDsInternal` без `systemBotID != Nil` (List*ForBot — с guard) | W3-A2-015 deferred (TASK-042): holder `*MessengerService` читает без bot-gate | Единый `systemBotID` guard на все bot/internal read API |
| A2W3-2-004 | NEW | P3 | `service_channel_invites.go:168` | `isMember, _ := s.repo.IsMember` — ошибка IsMember проглочена | При сбое БД Accept идёт как «не member» → лишний Increment/AddMembers | Проверять `err`; при err → 500 |

## FIXED-OK

| Stage A ID | Evidence |
|------------|----------|
| W3-A2-001 P0 | `promoteCheck` вызывается; nil → 403; wiring `SetPromotePermissionChecker` (TASK-010) |
| W3-A2-002 | DeleteMessage: `existing.ChannelID == path` + role по msg channel |
| W3-A2-003/004/005/006/007 | Reaction/Pin/PollResults/Bookmarks membership+owner |
| W3-A2-008 | `RevokeChannelInvite(channelID, inviteID)` + ChannelIDEQ |
| W3-A2-010 | NATS `via_invite:true` + `invite_id`, без raw token |
| W3-A2-012 | last owner Leave → 400 |
| W3-A2-013 | individual Mentions ∩ memberSet |
| W3-A2-014 | Forward unread++ |
| W3-A2-016 | denial-тесты task017 / wave123 promote |
| W3-A2-017 | `truncateRunes` title |
| W3-A2-018 | Forward source.ChannelID == path channelID |

## Чисто

- Search empty-scope / A2001; cascade soft-delete; SendMessage unread fanout; DM dedup; invite create/list RBAC; bot UpdateBotMessage ownership; auto-channels lifecycleCtx.

## Сводка агента

| Метрика | Значение |
|---------|----------|
| Файлов | 28/28/0 |
| Открытых | **4** |
| NEW | 1 (P3) |
| REGRESS | 0 |
| RESIDUAL | 3 (P2×3) |
| FIXED-OK | 14 |
