# A3 W3 агент 3 — messenger repos + presence (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Baseline: [A2W3-3-messenger-repos.md](A2W3-3-messenger-repos.md) · A1: [A3-3-messenger-repos.md](A3-3-messenger-repos.md)  
> Периметр: 20 файлов (9 prod + 11 tests) · Прочитано: **20/20** · Пропущено: **0**  
> Remediation: TASK-108 (`A2W3-3-001..003`), TASK-089 (`A2W3-3-004`), W13–W17 closes Stage A W3-A3-*

## Сверка с A2 (A2W3-3-*)

| A2 ID | Tag | Статус | Комментарий |
|-------|-----|--------|-------------|
| A2W3-3-001 | FIXED-OK | закрыт | TASK-108: `IncrementInviteUses` → `UseCountLT` + `ErrInviteUsesExhausted`; Ent-тест `TestEntIncrementInviteUses_UseCountLT_A2W3_3_001` (вне слота) |
| A2W3-3-002 | RESIDUAL | частично | TASK-108: caps на ListMembers/pins/invites/bookmarks/votes/project members/`GetChannelMemberIDs`/`GetUserChannelMemberships`; **остались** `ListUserChannels` + `WithMembers()` без Limit — см. A3W3-3-001 |
| A2W3-3-003 | FIXED-OK | закрыт | TASK-108: `Expire` 30d на Increment + `ClearChannelForUsers` на leave/DeleteChannel; тест `TestUnread_TTLAndLeaveCleanup_A2W3_3_003` |
| A2W3-3-004 | FIXED-OK | закрыт | TASK-089: push с `reactions.emoji $ne` + retry `$addToSet`; concurrent тест `reaction_concurrent_task089_test.go` |

**REGRESS:** 0

### Сверка с A1 (через A2 FIXED-OK) — подтверждение на HEAD

| A1 / A2 FIXED-OK | Tag на A3 |
|------------------|-----------|
| W3-A3-001 RemoveBookmark userID+UserIDEQ | FIXED-OK — `ent_repository.go:404-419` |
| W3-A3-002 sequential MatchedCount | FIXED-OK — + concurrent `$ne` (TASK-089) |
| W3-A3-003 away=90s / stale=10m | FIXED-OK — `presence.go:19-23` |
| W3-A3-004 pipe.Exec err | FIXED-OK — `presence.go:78-82` + `TestGetBulkStatus_RedisDown_ReturnsError` |
| W3-A3-005 unread TTL + leave HDel | FIXED-OK — `unread.go` + `limits.go:unreadKeyTTLSeconds` |
| W3-A3-006 list caps | FIXED-OK частично — residual A3W3-3-001 |
| W3-A3-007 invite UseCountLT | FIXED-OK |
| W3-A3-008 GetChannelMemberIDs soft-delete | FIXED-OK — `HasChannelWith(DeletedAtIsNil)` |
| W3-A3-009 GetStatus godoc | FIXED-OK — синхронизирован с 90s |
| W3-A3-010 ParseInt slog.Warn | FIXED-OK — `unread.go:77-82` |
| W3-A3-011 / 012 mock AddReaction + Test* | FIXED-OK — dedup + `TestMockRepository_AddReactionContract` |
| A2001 empty search scope | FIXED-OK — Mongo+mock+`mongo_search_empty_scope_a2001_test.go` |

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/messenger/repository.go` | 226 | 0 | чисто — ISP; RemoveBookmark(userID); IncrementInviteUses(maxUses); empty search scope |
| `backend/internal/messenger/ent_repository.go` | 716 | 2 | caps/UseCountLT/bookmark OK; residual ListUserChannels/WithMembers; ListMembers без soft-delete channel |
| `backend/internal/messenger/mongo_repository.go` | 330 | 0 | FIXED-OK AddReaction concurrent; List/Search limit+cursor; empty scope |
| `backend/internal/messenger/combined_repository.go` | 15 | 0 | чисто — embedding Ent+Mongo |
| `backend/internal/messenger/mock_repository.go` | 727 | 0 | FIXED-OK AddReaction dedup + empty SearchMessages |
| `backend/internal/messenger/presence.go` | 186 | 1 | FIXED-OK away/stale + pipe.Exec; godoc GetOnlineUserIDs stale «5 мин» |
| `backend/internal/messenger/unread.go` | 87 | 0 | FIXED-OK TTL Expire + ClearChannelForUsers + ParseInt slog |
| `backend/internal/messenger/export_hooks.go` | 144 | 0 | чисто — hubtest re-exports |
| `backend/internal/messenger/hookfixture.go` | 82 | 0 | чисто — in-memory @hub-ai fixture |
| `backend/testframework/messenger/ent_repository_invites_test.go` | 111 | 0 | чисто — invites CRUD + project members/role |
| `backend/testframework/messenger/ent_repository_poll_test.go` | 103 | 0 | чисто — poll lifecycle + FindGlobalChannel |
| `backend/testframework/messenger/ent_repository_test.go` | 225 | 0 | чисто — channel CRUD soft-delete, members, DM, star/mute, pins/bookmarks |
| `backend/testframework/messenger/ent_repository_wave123_test.go` | 123 | 0 | чисто — mute/displayName/invite/project/vote |
| `backend/testframework/messenger/mongo_repository_test.go` | 83 | 0 | lifecycle OK; concurrent race — вне слота (`reaction_concurrent_task089_test.go`) |
| `backend/testframework/messenger/mongo_repository_wave123_test.go` | 75 | 0 | чисто — cursor ListMessages + Find not found |
| `backend/testframework/messenger/mongo_search_empty_scope_a2001_test.go` | 45 | 0 | чисто — A2001 mutation-ready |
| `backend/testframework/messenger/mock_repository_test.go` | 767 | 0 | FIXED-OK `TestMockRepository_AddReactionContract` |
| `backend/testframework/messenger/presence_test.go` | 165 | 0 | чисто — online/away/offline/cleanup/bulk + RedisDown err |
| `backend/testframework/messenger/presence_wave100_test.go` | 99 | 0 | чисто — ParseStatusDuration, GetOnlineUserIDs, cleanup cancel |
| `backend/testframework/messenger/unread_test.go` | 105 | 0 | skip sender/MarkRead OK; TTL/ClearChannel — evidence в `task108_*` (вне слота) |

Файлов в списке: 20. Прочитано: 20. Пропущено: 0.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|----------|-------------------|------------|
| A3W3-3-001 | RESIDUAL (A2W3-3-002 / W3-A3-006) | P2 | `ent_repository.go:58-75`, `:230-238` | После TASK-108 остались unbounded: `ListUserChannels` `.All()` без Limit; `GetChannel` / `FindDMChannel` → `WithMembers()` без Limit | Sidebar/search scope и каждый GetChannel тянут всех members/каналы без cap → OOM/latency на «шумных» юзерах; соседние List* уже capped | `Limit(userChannelMembershipCap)` на ListUserChannels; `WithMembers(func(q){ q.Limit(listMembersCap) })` |
| A3W3-3-002 | NEW | P3 | `ent_repository.go:241-247` | `ListMembers` — только `ChannelID`, **без** `HasChannelWith(channel.DeletedAtIsNil())` (в отличие от `IsMember` / `GetChannelMemberIDs`) | После soft-delete канала путь ListMembers всё ещё отдаёт состав → UI/API шум vs GetChannel=404 | Добавить тот же soft-delete predicate; Ent-тест: DeleteChannel → ListMembers empty/err |
| A3W3-3-003 | NEW | P3 | `presence.go:154-162` | Godoc `GetOnlineUserIDs`: «heartbeat менее **5 минут**»; код фильтрует по `presenceAwayThreshold` (**90s**) | Рассинхрон документации с порогами после W3-A3-003/009; риск ошибочных фиксов TTL | Заменить godoc на «менее presenceAwayThreshold (90s)» |

## Чисто (подтверждено после remediation)

- IDOR: RemoveBookmark `UserIDEQ` + NotFound; RevokeChannelInvite `ChannelIDEQ`.
- Invite race: `UseCountLT` + `ErrInviteUsesExhausted` + Decrement на fail AddMembers (service/TASK-108).
- Mongo reactions: sequential MatchedCount + concurrent `$ne` push + race retry (TASK-089).
- Soft-delete: Get/ListUserChannels/IsMember/GetMemberRole/GetChannelMemberIDs/FindDM — `DeletedAtIsNil`.
- Presence: away=90s < stale=10m; `pipe.Exec` fail-closed; custom status TTL.
- Unread: Expire 30d; ClearChannelForUsers; corrupt ParseInt → slog.Warn + skip.
- Search empty scope A2001; Mongo List/Search cursor+limit.
- Polls: CreatePoll/VotePoll в Tx; ListVotes capped.
- Optimistic lock: `UpdateChannel` VersionEQ + DeletedAtIsNil + AddVersion.

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 1 |
| P3 | 2 |
| **Total** | **3** |

A2 baseline: 1 активный (A2W3-3-004) + partial A2W3-3-002 → FIXED-OK 3 (001/003/004) · RESIDUAL 1 (002→A3W3-3-001) · NEW 2 · REGRESS 0.

**Топ ID:** A3W3-3-001 (ListUserChannels/WithMembers unbounded) · A3W3-3-002 (ListMembers soft-delete) · A3W3-3-003 (GetOnlineUserIDs godoc).
