# W3 агент A1 — messenger handlers

## Прочитанные файлы
| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| backend/internal/messenger/handler.go | 149 | 1 | замечания |
| backend/internal/messenger/handler_channels.go | 566 | 1 | замечания |
| backend/internal/messenger/handler_messages.go | 287 | 2 | замечания |
| backend/internal/messenger/handler_presence.go | 121 | 1 | замечания |
| backend/internal/messenger/handler_reactions.go | 408 | 2 | замечания |
| backend/internal/messenger/handler_search.go | 43 | 0 | чисто |
| backend/internal/messenger/dto.go | 284 | 1 | замечания |
| backend/internal/messenger/events.go | 122 | 0 | чисто |
| backend/testframework/messenger/handler_channels_extra_test.go | 133 | 0 | чисто |
| backend/testframework/messenger/handler_flow_test.go | 451 | 1 | замечания |
| backend/testframework/messenger/handler_helpers_test.go | 339 | 0 | чисто |
| backend/testframework/messenger/handler_parse_test.go | 90 | 0 | чисто |
| backend/testframework/messenger/handler_wave100_test.go | 80 | 1 | замечания |
| backend/testframework/messenger/handler_wave106_test.go | 108 | 0 | чисто |
| backend/testframework/messenger/handler_wave109_test.go | 111 | 0 | чисто |
| backend/testframework/messenger/handler_wave112_test.go | 137 | 1 | замечания |
| backend/testframework/messenger/handler_wave113_test.go | 88 | 0 | чисто |
| backend/testframework/messenger/handler_wave114_test.go | 49 | 1 | замечания |
| backend/testframework/messenger/handler_wave115_test.go | 49 | 0 | чисто |
| backend/testframework/messenger/handler_wave116_test.go | 48 | 0 | чисто |
| backend/testframework/messenger/handler_wave123_test.go | 94 | 1 | замечания |
| backend/testframework/messenger/handler_wave93_test.go | 126 | 0 | чисто |
| backend/testframework/messenger/dto_test.go | 163 | 0 | чисто |
| backend/testframework/messenger/validation_test.go | 138 | 0 | чисто |
| backend/testframework/messenger/publish_typing_a2009_test.go | 52 | 0 | чисто |
| backend/testframework/messenger/stubs_test.go | 42 | 0 | чисто |

Файлов в списке: 26. Прочитано: 26. Пропущено: 0.

## Находки
| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|---------|-------------------|------------|
| W3-A1-001 | P1 | `backend/internal/messenger/handler_reactions.go:190-204` | `RemoveBookmark` парсит `bookmarkId`, но вызывает `RemoveBookmark(ctx, id)` **без `userID`** | IDOR: любой аутентифицированный пользователь удаляет чужую закладку по UUID. Service (`service_bookmarks.go:20-21`) тоже не проверяет владельца — guard отсутствует на всём пути | `UserIDFromCtx` → `RemoveBookmark(ctx, userID, id)`; repo `Where(ID, HasUserWith(...))`; чужая → 404 |
| W3-A1-002 | P1 | `backend/internal/messenger/handler_reactions.go:29-49` | `ToggleReaction` игнорирует `:channelId` из URL; в service уходит только `messageID`+`userID`+emoji | NF-02-ish: реакция на сообщение без membership. Service (`service_reactions.go:13-43`) делает `FindMessageByID` / Add/RemoveReaction **без** `IsMember` — любой auth-user с UUID сообщения мутирует реакции и шлёт NATS в канал | Парсить `channelID`; `ToggleReaction(ctx, channelID, messageID, userID, emoji)` + `IsMember` + `msg.ChannelID == channelID`; иначе 403/404 |
| W3-A1-003 | P1 | `backend/internal/messenger/handler_messages.go:144-162` | `DeleteMessage` передаёт `channelID` из URL и `isAdmin=false`; URL-канал **не сверяется** с каналом сообщения | Admin/owner канала A + `messageId` из канала B → role check идёт по URL-каналу A, soft-delete применяется к сообщению B (`service_messages.go:167-177`). Cross-channel delete | Перед delete: `existing.ChannelID == channelID` из URL; role брать по `existing.ChannelID`; mismatch → 404 |
| W3-A1-004 | P2 | `backend/internal/messenger/handler_messages.go:109-129` | `EditMessage` не парсит `:channelId`; service — только авторство (`SenderID`) | URL-канал декоративный. Нет membership/binding message→channel на handler-слое; автор после leave всё ещё редактирует | Передать `channelID`; проверить `IsMember` + `msg.ChannelID == channelID` |
| W3-A1-005 | P2 | `backend/internal/messenger/handler_channels.go:411-414` (+439-441, 463-464, 490-491, 522-523) | Invite-handlers мапят **любую** service-ошибку в 500/400/`NotFound` и отдают `err.Error()` клиенту | Forbidden (не admin) → 500; внутренний текст утекает в API вместо `{"error","request_id"}` / AppError-кодов | `return err` как в остальных handlers; не `fiber.NewError(..., err.Error())` |
| W3-A1-006 | P2 | `backend/internal/messenger/dto.go:14-72` (+ handlers `Bind().JSON` без validator) | Теги `validate:"required,…"` на DTO не исполняются | Пустые `content_json` / `emoji` / `member_ids` / `target_user_id` принимаются Bind’ом; валидация только частично в service | Fiber StructValidator или `validate.Struct` после Bind |
| W3-A1-007 | P2 | `backend/internal/messenger/handler_presence.go:56-68` | `GetPresence` не берёт caller `userID`; любой auth-user запрашивает статусы произвольного списка `user_ids` | Перечисление online/offline по UUID без ограничения «общие каналы / org-policy» | Ограничить scope (общие каналы / свои контакты) или хотя бы rate-limit + cap длины списка |
| W3-A1-008 | P2 | `backend/testframework/messenger/handler_wave100_test.go:42-69` (+ `handler_wave112_test.go:58-87`) | `RemoveBookmark_*` — только happy-path своего bookmark на mock | Тест зелёный при живом IDOR (W3-A1-001); нет «чужой bookmarkId → 404» | Negative: bookmark user B, delete от user A → 404; guard в repo с user scope |
| W3-A1-009 | P2 | `backend/testframework/messenger/handler_wave114_test.go:40-49` | `GetPollResults_InvalidChannelID` бьёт в `/channels/{bad}/polls/{id}/results` | Реальный маршрут — `GET /api/messenger/polls/:pollId` (`handler.go:111`). Тест не вызывает `GetPollResults`; любой 4xx «проходит» | Хитнуть `/api/messenger/polls/not-a-uuid` → 400 |
| W3-A1-010 | P2 | `backend/testframework/messenger/handler_flow_test.go:21-35` | Почти все handler-тесты на `mockRepository` / `newTestService()` | Membership/IDOR (W3-A1-001…003) mock’ом не ловятся; `IsMember` в mock легко «зелёный» | Authz/IDOR — Ent/Mongo integration; mock — parse/HTTP shape |
| W3-A1-011 | P3 | `backend/internal/messenger/handler.go:27-36` | `queryInt` при `limit=-5` возвращает −5 (Atoi успешен) | Отрицательный limit уходит в `ListMessages` / search / thread | Clamp: `if v < 1 \|\| v > max { return defaultVal }` |
| W3-A1-012 | P3 | `backend/testframework/messenger/handler_wave123_test.go:32-35` | После mute шлёт `DELETE .../mute` без assert статуса | Маршрута DELETE mute нет (`unmute` = POST). Cleanup молча 404/405; тест не ловит регрессию unmute | `POST .../unmute` + `assert.Equal(200, …)` |

## Чисто

Файлы без находок — что проверялось:

- **handler_search.go** — `UserIDFromCtx`; обязательный `query`; `channel_id`+`limit` в service (membership при filter — зона A2); нет `_ = err`; nil-deref нет.
- **events.go** — только NATS subject constants и event DTO; секретов/логики нет.
- **handler_channels_extra_test.go** — join/leave/star/mute/members/invite/edit-delete/typing; asserts есть; happy-path member.
- **handler_helpers_test.go** — splitComma/queryInt; Promote/Vote/Close error paths; `ListMessages_NonMemberForbidden`; `CreateInviteLink_RequiresAdmin`; asserts; membership denial на service (не assert-free).
- **handler_parse_test.go** — parseChannelID invalid/valid; parseMessageID; search bad channel_id; presence/unread nil-safe.
- **handler_wave106/109/113/115/116** — invalid JSON/UUID, private join denial, poll/message parse; asserts присутствуют.
- **handler_wave93_test.go** — thread replies + revoke invite + poll vote/close; asserts; URL channel для replies декоративен, но service HI-01 membership по parent — ок на service-слое.
- **dto_test.go** — toMessageResponse/channel/invite/poll anonymous voters; asserts.
- **validation_test.go** — expires_in, invite token, CreateChannel/CreatePoll/CreateInvite validation; non-member CreatePoll.
- **publish_typing_a2009_test.go** — non-member typing → 403, member OK; membership guard на service (A2-009).
- **stubs_test.go** — `setupChannelWithMembers` helper; логики/assert-free тестов нет.

Дополнительно по «чистым» handler-путям (кроме находок):

- **Create/List/Get/Update/Delete channel, Join/Leave, Add/List members, star/mute** — `parseChannelID` + `UserIDFromCtx`; membership делегируется service.
- **Send/List messages, Forward, pins, polls Create** — channelID+userID прокидываются.
- **ListThreadReplies** — URL `:channelId` не читается, но service проверяет `IsMember(parent.ChannelID)` (HI-01); не IDOR при корректном service.
- **SearchMessages** — userID обязателен; channel filter валидируется в service.
- **PromoteToIssue** — messageID+userID; membership в service (тест non-member в helpers).
- Soft-delete сообщений: handler → `DeleteMessage(..., false)` → service `SoftDeleteMessage` (корзина), не hard delete.
- `_ =` / проглоченные error в prod handlers списка — не найдены (кроме test `_ = userID` / `_ = json.Marshal` в тестах).
