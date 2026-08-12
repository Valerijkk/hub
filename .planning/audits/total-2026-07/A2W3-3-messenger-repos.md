# A2 W3 агент 3 — messenger repos + presence (re-audit)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> Зеркало: `W3-ASSIGNMENTS.md` A3 · Stage A: `A3-3-messenger-repos.md`

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/messenger/repository.go` | 198 | 0 | чисто (контракты + empty search scope) |
| `backend/internal/messenger/ent_repository.go` | 575 | 2 | RESIDUAL lists/invite RMW; FIXED bookmark/soft-delete |
| `backend/internal/messenger/mongo_repository.go` | 268 | 1 | NEW concurrent reaction race; FIXED MatchedCount |
| `backend/internal/messenger/combined_repository.go` | 13 | 0 | чисто |
| `backend/internal/messenger/mock_repository.go` | 600 | 0 | FIXED-OK AddReaction dedup |
| `backend/internal/messenger/presence.go` | 168 | 0 | FIXED-OK away/stale + pipe.Exec err |
| `backend/internal/messenger/unread.go` | 58 | 1 | RESIDUAL no TTL; FIXED ParseInt slog |
| `backend/internal/messenger/export_hooks.go` | 115 | 0 | чисто |
| `backend/internal/messenger/hookfixture.go` | 72 | 0 | чисто |
| `backend/testframework/messenger/ent_repository_invites_test.go` | 92 | 0 | чисто |
| `backend/testframework/messenger/ent_repository_poll_test.go` | 85 | 0 | чисто |
| `backend/testframework/messenger/ent_repository_test.go` | 182 | 0 | чисто |
| `backend/testframework/messenger/ent_repository_wave123_test.go` | 99 | 0 | чисто |
| `backend/testframework/messenger/mongo_repository_test.go` | 68 | 0 | чисто |
| `backend/testframework/messenger/mongo_repository_wave123_test.go` | 64 | 0 | чисто |
| `backend/testframework/messenger/mongo_search_empty_scope_a2001_test.go` | 37 | 0 | чисто |
| `backend/testframework/messenger/mock_repository_test.go` | 633 | 0 | FIXED-OK `TestMockRepository_AddReactionContract` |
| `backend/testframework/messenger/presence_test.go` | 133 | 0 | чисто |
| `backend/testframework/messenger/presence_wave100_test.go` | 85 | 0 | чисто |
| `backend/testframework/messenger/unread_test.go` | 86 | 0 | чисто |

Файлов в списке: **20**. Прочитано: **20**. Пропущено: **0**.

## Находки (открытые)

| ID | Tag | Severity | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|----------|-------------|----------|---------------|------------|
| A2W3-3-004 | NEW | P2 | `mongo_repository.go:205-240` | После фикса MatchedCount: два concurrent first-react на один emoji оба видят `MatchedCount==0` → двойной `$push` | Гонка TOCTOU между check и push; sequential retry закрыт, concurrent — нет | Unique compound / pipeline `$addToSet` на emoji array; или findAndModify (TASK-089) |

## FIXED-OK

| Stage A ID | Evidence |
|------------|----------|
| A2W3-3-001 | TASK-108: `Update().Where(IDEQ, UseCountLT(max)).AddUseCount(1)` → `ErrInviteUsesExhausted` |
| A2W3-3-002 | TASK-108: hard caps Limit на ListMembers/pins/invites/bookmarks/votes/project members |
| A2W3-3-003 | TASK-108: unread TTL (Expire 30d) + HDel на leave/DeleteChannel |
| W3-A3-001 | `RemoveBookmark(userID, bookmarkID)` + UserIDEQ |
| W3-A3-002 | sequential: `MatchedCount>0` → return (не `$push` duplicate) |
| W3-A3-003 | away=90s, stale=10m |
| W3-A3-004 | `pipe.Exec` err → return |
| W3-A3-008 | `GetChannelMemberIDs` + `DeletedAtIsNil` (TASK-061) |
| W3-A3-009 | godoc GetStatus синхронизирован |
| W3-A3-010 | corrupt ParseInt → slog.Warn + skip |
| W3-A3-011 | `TestMockRepository_AddReactionContract` добавлен |
| W3-A3-012 | mock AddReaction dedup userID |
| A2001 | empty channelIDs → `[]` (Mongo+mock+тест) |

## Чисто

- Mongo ListMessages/Search limit+cursor; IsMember soft-delete; CreatePoll Tx; presence custom TTL; unread Increment skip sender.

## Сводка агента

| Метрика | Значение |
|---------|----------|
| Файлов | 20/20/0 |
| Открытых | **4** |
| NEW | 1 (P2) |
| REGRESS | 0 |
| RESIDUAL | 3 (P2×3) |
| FIXED-OK | 10 |
