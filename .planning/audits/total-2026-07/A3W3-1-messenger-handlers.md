# A3 W3 агент 1 — messenger handlers (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Mirror: [A2W3-1-messenger-handlers.md](A2W3-1-messenger-handlers.md) · Assignment: [W3-ASSIGNMENTS.md](W3-ASSIGNMENTS.md) § A1  
> Периметр: 26 файлов (8 prod + 18 tests) · Прочитано: **26/26** · Пропущено: **0**

## Сверка с A2 (A2W3-1-*)

| A2 ID | Tag | Статус в HEAD | Комментарий |
|-------|-----|---------------|-------------|
| A2W3-1-001 | FIXED-OK | закрыт | `GetPresence` → `UserIDFromCtx` + `service.GetPresence(ctx, callerID, …)` + `FilterUsersSharingChannel`; cap 50 — `handler_presence.go:56-69`, `service.go:53-91`; Ent-тест `task108_invite_dm_unread_presence_test.go` (вне слота); TASK-108 |
| A2W3-1-002 | RESIDUAL | живо | `handler_wave114_test.go:41-49` всё ещё бьёт несуществующий `/channels/{bad}/polls/{id}/results`; канон `GET /polls/:pollId` — `handler.go:115` |
| A2W3-1-003 | FIXED-OK | закрыт | `ToggleReaction` парсит `channelID` → `service.ToggleReaction(ctx, channelID, messageID, …)` — `handler_reactions.go:29-48`; IDOR suite `membership_idor_task017_test.go` (вне слота); TASK-108 |

### Cascade FIXED-OK (Stage A / W3-A1-*, без регрессии)

| Stage A / prior | Evidence |
|-----------------|----------|
| W3-A1-001 bookmark owner | `RemoveBookmark(ctx, userID, id)` — `handler_reactions.go:194-209` |
| W3-A1-002/003/004 msg channel bind | Edit/Delete/ToggleReaction + channelID; service membership |
| W3-A1-005 invite `return err` | `handler_channels.go:411-414`, `:440-442`, `:463-465`, `:490-492`, `:523-524` |
| W3-A1-006 StructValidator | DTO `validate` теги — `dto.go:15-73`; wiring вне слота |
| W3-A1-007 presence scope | ≡ A2W3-1-001 |
| W3-A1-008/010 Ent-negative | task017 / authz вне mock-only (вне списка A1) |
| W3-A1-009 wrong poll URL | **не закрыт** → A3W3-1-001 RESIDUAL |
| W3-A1-011 queryInt clamp | `handler.go:31-40`: `v < 1 \|\| v > 100` → default |
| W3-A1-012 unmute POST | `handler_wave123_test.go:32-36` + route `POST …/unmute` |

**REGRESS:** 0 · **RESIDUAL (этот слот):** 1 · **NEW (этот слот):** 0

## Прочитанные файлы

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `backend/internal/messenger/handler.go` | 153 | 0 | FIXED-OK queryInt clamp; routes `GET /polls/:pollId` |
| `backend/internal/messenger/handler_channels.go` | 567 | 0 | FIXED-OK invite `return err`; Create/Update/DM + userID |
| `backend/internal/messenger/handler_messages.go` | 291 | 0 | FIXED-OK Edit/Delete channel bind; Forward/Promote OK |
| `backend/internal/messenger/handler_presence.go` | 125 | 0 | FIXED-OK A2W3-1-001 callerID → service |
| `backend/internal/messenger/handler_reactions.go` | 416 | 0 | FIXED-OK ToggleReaction channelID; bookmark owner |
| `backend/internal/messenger/handler_search.go` | 43 | 0 | чисто — userID + query required |
| `backend/internal/messenger/dto.go` | 287 | 0 | FIXED-OK validate + ChannelResponse.Version (TASK-095) |
| `backend/internal/messenger/events.go` | 122 | 0 | чисто — NATS subjects / event DTO |
| `backend/testframework/messenger/handler_channels_extra_test.go` | 133 | 0 | чисто — join/leave/star/mute/members |
| `backend/testframework/messenger/handler_flow_test.go` | 452 | 0 | mock happy-path; Ent-neg вне файла |
| `backend/testframework/messenger/handler_helpers_test.go` | 405 | 0 | чисто — queryInt negative; promote/poll denial |
| `backend/testframework/messenger/handler_parse_test.go` | 90 | 0 | чисто — parseChannelID / search bad id |
| `backend/testframework/messenger/handler_wave100_test.go` | 80 | 0 | FIXED-OK bookmark happy-path (+ owner вне файла) |
| `backend/testframework/messenger/handler_wave106_test.go` | 108 | 0 | чисто — invalid JSON/UUID |
| `backend/testframework/messenger/handler_wave109_test.go` | 111 | 0 | чисто — typing/heartbeat/before |
| `backend/testframework/messenger/handler_wave112_test.go` | 137 | 0 | чисто — unread/delete/bookmark/poll JSON |
| `backend/testframework/messenger/handler_wave113_test.go` | 88 | 0 | чисто — leave/forward/unmute |
| `backend/testframework/messenger/handler_wave114_test.go` | 49 | 1 | RESIDUAL wrong GetPollResults URL |
| `backend/testframework/messenger/handler_wave115_test.go` | 49 | 0 | чисто — pin parse |
| `backend/testframework/messenger/handler_wave116_test.go` | 48 | 0 | чисто — join/members parse |
| `backend/testframework/messenger/handler_wave123_test.go` | 96 | 0 | FIXED-OK unmute POST + assert 200 |
| `backend/testframework/messenger/handler_wave93_test.go` | 126 | 0 | чисто — replies/invite/poll vote+GET правильный path |
| `backend/testframework/messenger/dto_test.go` | 163 | 0 | чисто — toMessageResponse / poll voters |
| `backend/testframework/messenger/validation_test.go` | 138 | 0 | чисто — expires/create/poll validation |
| `backend/testframework/messenger/publish_typing_a2009_test.go` | 52 | 0 | чисто — non-member typing 403 |
| `backend/testframework/messenger/stubs_test.go` | 42 | 0 | чисто — setupChannelWithMembers helper |

Файлов в списке: **26**. Прочитано: **26**. Пропущено: **0**.

## Находки (активные A3)

| ID | Tag | Sev | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|-----|-------------|---------|---------------|------------|
| A3W3-1-001 | RESIDUAL (A2W3-1-002 / W3-A1-009) | P2 | `handler_wave114_test.go:41-49` | `GetPollResults_InvalidChannelID` бьёт `GET /channels/not-a-uuid/polls/{uuid}/results` — маршрута нет (`GET /api/messenger/polls/:pollId`, `handler.go:115`) | Любой 4xx (в т.ч. 404 Fiber «Cannot GET») «проходит»; регресс `parsePollID` / GetPollResults 400 не ловится; wave93 уже ходит по правильному path | Хитнуть `/api/messenger/polls/not-a-uuid` → 400; переименовать тест в `InvalidPollID` |

## Чисто (перепроверено)

- **Authz/IDOR (handler):** channelID+userID на Create/List/Join/Leave/members/star/mute/pins/polls Create/Send/List/Forward/Edit/Delete/ToggleReaction/TogglePin; bookmark Remove с userID; presence с caller scope.
- **ListThreadReplies:** URL `:channelId` не читается — membership по parent в service (HI-01); не IDOR при корректном service.
- **Soft-delete:** `DeleteMessage(..., false)` → service soft path.
- **OL channel:** `CreateChannelRequest.Version` + service `Version < 1` → 400; repo VersionEQ (TASK-095).
- **Layering:** нет `*ent.Client` / `http.DefaultClient` / `io.ReadAll` / `_ = err` в prod handlers слота.
- **queryInt:** clamp 1…100.
- **Тесты:** wave* asserts; unmute POST; typing non-member; Ent-negative bookmark/presence/reaction вне mock-слота.

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 1 |
| P3 | 0 |
| **Total** | **1** |

### Теги vs A2

| Tag | Count |
|-----|------:|
| FIXED-OK (A2 active) | 2 (`A2W3-1-001`, `A2W3-1-003`) |
| REGRESS | 0 |
| RESIDUAL | 1 (`A3W3-1-001` ← A2W3-1-002) |
| NEW | 0 |

A2 baseline open (file-table): 3 → FIXED-OK 2 · RESIDUAL 1 · NEW 0 · REGRESS 0.

**Топ ID:** A3W3-1-001 (wave114 GetPollResults wrong URL / W3-A1-009).
