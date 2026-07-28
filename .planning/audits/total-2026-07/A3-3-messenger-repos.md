# W3 агент A3 — messenger repos + presence

> TASK-008 · стадия A · READ-ONLY · 2026-07-28  
> Фокус: Mongo cursor/limit, Ent membership queries, presence Redis TTL, unread, empty search scope **A2001**.  
> Перепроверка grok-2026-07: **A2-001 / A2-012**.

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/messenger/repository.go` | 219 | 0 | чисто — ISP-контракты; SearchMessages: пустой `channelIDs` → пустой результат |
| `backend/internal/messenger/ent_repository.go` | 606 | 4 | RemoveBookmark IDOR; lists без Limit; invite RMW; memberIDs без soft-delete |
| `backend/internal/messenger/mongo_repository.go` | 306 | 1 | AddReaction: `ModifiedCount==0` → дубликат emoji; List/Search limit+cursor OK |
| `backend/internal/messenger/combined_repository.go` | 15 | 0 | чисто — embedding Ent+Mongo |
| `backend/internal/messenger/mock_repository.go` | 660 | 1 | AddReaction без дедупа userID; Search empty scope OK |
| `backend/internal/messenger/presence.go` | 179 | 3 | away vs stale TTL; pipe.Exec `_`; комментарий GetStatus |
| `backend/internal/messenger/unread.go` | 61 | 2 | нет TTL на Redis hash; ParseInt `_` |
| `backend/internal/messenger/export_hooks.go` | 133 | 0 | чисто — hubtest re-exports |
| `backend/internal/messenger/hookfixture.go` | 82 | 0 | чисто — in-memory fixture для @hub-ai hook |
| `backend/testframework/messenger/ent_repository_invites_test.go` | 111 | 0 | чисто — invites CRUD + project members/role + CreatePoll |
| `backend/testframework/messenger/ent_repository_poll_test.go` | 103 | 0 | чисто — poll lifecycle + FindGlobalChannel; handler smoke |
| `backend/testframework/messenger/ent_repository_test.go` | 225 | 0 | чисто — channel CRUD soft-delete, members, DM, star/mute, pins/bookmarks |
| `backend/testframework/messenger/ent_repository_wave123_test.go` | 123 | 0 | чисто — mute durations, displayName, invite, project members, vote |
| `backend/testframework/messenger/mongo_repository_test.go` | 83 | 0 | чисто — lifecycle + reactions + thread + search (без double-react assert) |
| `backend/testframework/messenger/mongo_repository_wave123_test.go` | 75 | 0 | чисто — cursor ListMessages + Find not found + search scoped |
| `backend/testframework/messenger/mongo_search_empty_scope_a2001_test.go` | 45 | 0 | чисто — A2-001/A2001 mutation-ready на Mongo |
| `backend/testframework/messenger/mock_repository_test.go` | 662 | 1 | нет `Test*` — дубликат mock-фикстуры, assert-free |
| `backend/testframework/messenger/presence_test.go` | 150 | 0 | чисто — heartbeat/online/away/offline/cleanup/custom/bulk (away без Cleanup) |
| `backend/testframework/messenger/presence_wave100_test.go` | 99 | 0 | чисто — ParseStatusDuration, GetOnlineUserIDs, cleanup cancel, HookTestRepo |
| `backend/testframework/messenger/unread_test.go` | 105 | 0 | чисто — skip sender, multi-incr, MarkRead, GetAll |

Файлов в списке: 20. Прочитано: 20. Пропущено: 0.

## Перепроверка grok-2026-07 (A2-001 / A2-012)

| ID | Было | Текущий код | Статус |
|----|------|-------------|--------|
| A2-001 / A2001 | Search без membership; пустой scope = все каналы | Контракт `repository.go:89-91`; Mongo early-return `mongo_repository.go:281-283`; mock `mock_repository.go:372-378`; тест `mongo_search_empty_scope_a2001_test.go` | **FIXED** |
| A2-012 | `away` vs cleanup: stale=90s, away=5m | `presence.go:19-22`, `61-64`, `158-161` — без изменений | **OPEN** (= W3-A3-003) |

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|----------|-------------------|------------|
| W3-A3-001 | P1 | `ent_repository.go:346-347`; `service_bookmarks.go:19-21` (контекст) | `RemoveBookmark` — `DeleteOneID(bookmarkID)` без `userID` / ownership | Зная UUID чужой закладки, любой auth-user удаляет её (IDOR); handler передаёт только path-id | `Delete().Where(IDEQ, UserIDEQ(userID))`; signature + service + handler с `userID` из JWT; affected==0 → 404 |
| W3-A3-002 | P1 | `mongo_repository.go:218-235` | `AddReaction`: при существующем emoji `$addToSet` no-op даёт `ModifiedCount==0` → ветка `$push` создаёт **второй** `ReactionEntry` с тем же emoji | Повторный клик / retry / гонка → дубликаты реакций, кривые счётчики UI | Если `MatchedCount>0` — return nil; `$push` только при `MatchedCount==0`. Тест: double AddReaction same emoji |
| W3-A3-003 | P2 | `presence.go:19-22`, `61-64`, `158-161` | `presenceStaleThreshold=90s` < `presenceAwayThreshold=5m`; Cleanup удаляет heartbeat раньше, чем GetStatus может вернуть `away` | В проде с `StartCleanupLoop` статус `away` практически недостижим; тесты сажают score 6m вручную без cleanup | Выровнять пороги (напр. away=90s, stale=5–10m) или не чистить записи, пока GetStatus ещё показывает away |
| W3-A3-004 | P2 | `presence.go:76` | `GetBulkStatus`: `_, _ = pipe.Exec(ctx)` — ошибка pipeline проглочена | Redis/сеть fail → все статусы тихо `"offline"`, caller думает, что успех | Проверять `err` от `Exec`; return err (или partial+err) |
| W3-A3-005 | P2 | `unread.go:11-40`, `49-50` | Ключи `unread:{userID}` без TTL; MarkRead только `HDel` поля; нет очистки при leave/delete канала | Hash растёт orphan-полями; GetAll тянет мусор навсегда | TTL/refresh на hash; HDel всех членов при leave/soft-delete channel; опц. фильтр по ListUserChannels на service |
| W3-A3-006 | P2 | `ent_repository.go:69-75`, `135-146`, `186-190`, `292-295`, `330-333`, `351-354`, `472-478`, `590-593`, `359-373`, `386-403` | List*/Get*Member* / pins / bookmarks / invites / votes / project members — `.All()` без Limit/cap | Большой канал/орг → unbounded memory/latency на горячих путях | Limit/cursor где API; cap на member/project lists |
| W3-A3-007 | P2 | `ent_repository.go:449-460` | `IncrementInviteUses` — read-modify-write без атомарного `WHERE use_count < max_uses` | Гонка Accept → превышение MaxUses (комментарий в коде это признаёт) | `Update().Where(IDEQ, UseCountLT(max)).AddUseCount(1)`; affected==0 → deny |
| W3-A3-008 | P3 | `ent_repository.go:135-146` | `GetChannelMemberIDs` без `HasChannelWith(DeletedAtIsNil)` (в отличие от `IsMember`) | После soft-delete канала путь без IsMember (bot/internal) всё ещё видит members → unread/notify шум | Добавить тот же soft-delete predicate, что у IsMember |
| W3-A3-009 | P3 | `presence.go:47-48` | Комментарий GetStatus: «&lt;5 мин online, &lt;90 сек away» — противоречит коду (`elapsed < 5m → online`, иначе `away`) | Вводит в заблуждение при сопровождении / фиксах TTL | Синхронизировать godoc с фактическими порогами |
| W3-A3-010 | P3 | `unread.go:57` | `n, _ := strconv.ParseInt(v, 10, 64)` — битые значения Redis → 0 без ошибки | Тихая порча счётчиков при повреждённых данных | Логировать/пропускать поле с err; не маскировать в 0 без сигнала |
| W3-A3-011 | P3 | `mock_repository_test.go:1-662` | Файл `*_test.go` без единой `Test*` — полная копия mock-фикстуры testframework | Чек-лист TASK-008: assert-free «тест»; дубль с `internal/.../mock_repository.go` drift-риск | Переименовать в `mock_repository.go` (тот же package) или добавить контрактные тесты empty-scope/AddReaction |
| W3-A3-012 | P3 | `mock_repository.go:308-320` | Mock `AddReaction` всегда `append` userID без проверки дубликата | Тесты service на моке не поймают Mongo-баг W3-A3-002 | Дедуп userID в mock; regression на реальном Mongo |

## Чисто

- **A2001 / A2-001 SearchMessages** — пустой `channelIDs` → `[]` (не org-wide `$text`); Mongo + mock + mutation-ready Ent-слой тест.
- **Mongo ListMessages / ListThreadReplies** — cursor `before` + `SetLimit` cap 1..100 (default 50); `cursor.Close`; soft-delete `deleted_at: nil`.
- **SearchMessages limit** — cap 1..100, default 20; filter `$in` channelIDs.
- **IsMember / GetMemberRole / ListUserChannels / GetChannel / FindDM / DeleteChannel** — soft-delete (`DeletedAtIsNil` / `HasChannelWith`); A8-005.
- **CreatePollWithOptions / VotePoll** — Tx Rollback на ошибках, Commit в конце.
- **Custom status Redis TTL** — `SetCustomStatus` через `ParseStatusDuration` + `SET` с TTL.
- **Unread Increment** — skip sender; pipeline; MarkRead = HDel.
- **combined / export_hooks / hookfixture / repository interfaces** — без дефектов по чек-листу фокуса.
- **Ent-тесты списка** — assert’ы есть; soft-delete Get после Delete; membership CRUD.
- **presence/unread тесты** — assert’ы есть; gap: away-сценарий не гоняет CleanupStale (ожидаемо при OPEN A2-012).

## Сводка

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 2 |
| P2 | 5 |
| P3 | 5 |
| **Σ** | **12** |

Файлов: **20/20/0** (список / прочитано / пропущено).

**Топ-3:** W3-A3-001 (RemoveBookmark IDOR) · W3-A3-002 (AddReaction duplicate emoji) · W3-A3-003 (presence away vs stale TTL, A2-012 OPEN).
