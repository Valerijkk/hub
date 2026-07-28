# A2 W3 агент 1 — messenger handlers (re-audit)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> Зеркало: `W3-ASSIGNMENTS.md` A1 · Stage A: `A3-1-messenger-handlers.md`  
> Теги: `NEW` | `REGRESS` | `RESIDUAL` | `FIXED-OK`

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/messenger/handler.go` | 130 | 0 | FIXED-OK queryInt clamp |
| `backend/internal/messenger/handler_channels.go` | 511 | 0 | FIXED-OK invite `return err` |
| `backend/internal/messenger/handler_messages.go` | 259 | 0 | FIXED-OK Edit/Delete channel bind |
| `backend/internal/messenger/handler_presence.go` | 105 | 1 | RESIDUAL presence scope |
| `backend/internal/messenger/handler_reactions.go` | 364 | 1 | RESIDUAL path channelId на ToggleReaction |
| `backend/internal/messenger/handler_search.go` | 38 | 0 | чисто |
| `backend/internal/messenger/dto.go` | 252 | 0 | FIXED-OK validate + StructValidator |
| `backend/internal/messenger/events.go` | 99 | 0 | чисто |
| `backend/testframework/messenger/handler_channels_extra_test.go` | 112 | 0 | чисто |
| `backend/testframework/messenger/handler_flow_test.go` | 385 | 0 | чисто (+ Ent-negative вне списка) |
| `backend/testframework/messenger/handler_helpers_test.go` | 359 | 0 | чисто |
| `backend/testframework/messenger/handler_parse_test.go` | 81 | 0 | чисто |
| `backend/testframework/messenger/handler_wave100_test.go` | 68 | 0 | FIXED-OK bookmark owner (см. task017) |
| `backend/testframework/messenger/handler_wave106_test.go` | 91 | 0 | чисто |
| `backend/testframework/messenger/handler_wave109_test.go` | 93 | 0 | чисто |
| `backend/testframework/messenger/handler_wave112_test.go` | 116 | 0 | чисто |
| `backend/testframework/messenger/handler_wave113_test.go` | 75 | 0 | чисто |
| `backend/testframework/messenger/handler_wave114_test.go` | 40 | 1 | RESIDUAL wrong poll path |
| `backend/testframework/messenger/handler_wave115_test.go` | 41 | 0 | чисто |
| `backend/testframework/messenger/handler_wave116_test.go` | 40 | 0 | чисто |
| `backend/testframework/messenger/handler_wave123_test.go` | 82 | 0 | FIXED-OK unmute POST |
| `backend/testframework/messenger/handler_wave93_test.go` | 107 | 0 | чисто |
| `backend/testframework/messenger/dto_test.go` | 149 | 0 | чисто |
| `backend/testframework/messenger/validation_test.go` | 122 | 0 | чисто |
| `backend/testframework/messenger/publish_typing_a2009_test.go` | 43 | 0 | чисто |
| `backend/testframework/messenger/stubs_test.go` | 37 | 0 | чисто |

Файлов в списке: **26**. Прочитано: **26**. Пропущено: **0**.

## Находки (открытые)

| ID | Tag | Severity | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|----------|-------------|----------|---------------|------------|
| A2W3-1-002 | RESIDUAL | P2 | `handler_wave114_test.go:40-49` | Тест бьёт `GET /channels/{bad}/polls/{id}/results` — реального маршрута нет (`GET /polls/:pollId`) | W3-A1-009 **не закрыт**: любой 4xx «проходит»; регресс GetPollResults parse не ловится | Хитнуть `/api/messenger/polls/not-a-uuid` → 400 |

## FIXED-OK (Stage A → закрыто после B/C)

| Stage A ID | Evidence |
|------------|----------|
| A2W3-1-001 | TASK-108: GetPresence(callerID) + FilterUsersSharingChannel |
| A2W3-1-003 | TASK-108: ToggleReaction парсит channelID + msg.ChannelID match → 404 |
| W3-A1-001 | `RemoveBookmark(ctx, userID, id)` + repo owner-scope; `membership_idor_task017_test.go` |
| W3-A1-002 | service `IsMember(msg.ChannelID)` + denial-тесты P1-BE-16 |
| W3-A1-003 | `DeleteMessage` channel match + role по `existing.ChannelID` |
| W3-A1-004 | `EditMessage(channelID,…)` + membership + msg∈channel |
| W3-A1-005 | invite handlers `return err` (P2-BE-07); `handler_invite_error_contract_test.go` |
| W3-A1-006 | Fiber `StructValidator` в `wiring_fiber.go` + validate tags на DTO |
| W3-A1-008 | Ent-negative bookmark owner (task017) |
| W3-A1-010 | authz Ent-negative suite (task037) — mock больше не единственный слой |
| W3-A1-011 | `queryInt`: `v < 1 \|\| v > 100` → default |
| W3-A1-012 | wave123 unmute = `POST .../unmute` + assert 200 |

## Чисто

- **handler_search / events** — userID, query required, NATS constants.
- **Create/List/Join/Leave/members/star/mute/pins/polls Create** — channelID+userID; membership в service.
- **Send/List/Forward/Delete/Edit** — channel binding на service после фиксов.
- Остальные wave* handler-тесты — asserts есть; parse/invalid UUID.

## Сводка агента

| Метрика | Значение |
|---------|----------|
| Файлов | 26/26/0 |
| Открытых находок | **3** |
| NEW | 0 |
| REGRESS | 0 |
| RESIDUAL | 3 (P2×2, P3×1) |
| FIXED-OK | 10 |
