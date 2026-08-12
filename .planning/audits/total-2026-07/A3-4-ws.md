# W3 агент A4 — ws

> TASK-008 · READ-ONLY · 2026-07-28  
> Фокус: NF-02 (subscribe только после membership/RBAC), origin/CSWSH, hub shutdown / goroutine leaks, pump race (SEC-05).  
> Сверка с `grok-2026-07/02-realtime.md`: A2-013 (JWT в query) — residual; NF-02 channel/project path в prod wiring — закрыт; A2-007 call-checker — вне этого списка (wiring/calls).

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/ws/authorizer.go` | 91 | 0 | чисто |
| `backend/internal/ws/bridge.go` | 283 | 1 | замечания |
| `backend/internal/ws/export_hooks.go` | 97 | 1 | замечания |
| `backend/internal/ws/handler.go` | 299 | 3 | замечания |
| `backend/internal/ws/hub.go` | 362 | 3 | замечания |
| `backend/internal/ws/protocol.go` | 28 | 0 | чисто |
| `backend/testframework/ws/authorizer_test.go` | 203 | 0 | чисто |
| `backend/testframework/ws/bridge_lifecycle_test.go` | 63 | 0 | чисто |
| `backend/testframework/ws/bridge_test.go` | 205 | 0 | чисто |
| `backend/testframework/ws/handler_auth_test.go` | 104 | 0 | чисто |
| `backend/testframework/ws/hub_shutdown_test.go` | 160 | 0 | чисто |
| `backend/testframework/ws/hub_test.go` | 359 | 0 | чисто |
| `backend/testframework/ws/origin_test.go` | 58 | 0 | чисто |
| `backend/testframework/ws/protocol_test.go` | 160 | 0 | чисто |
| `backend/testframework/ws/pump_test.go` | 283 | 2 | замечания |

Файлов в списке: 15. Прочитано: 15. Пропущено: 0.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|---------|-------------------|------------|
| W3-A4-001 | P1 | `hub.go:114-139`; `hub.go:173-184`; `hub.go:221-226`; `handler.go:181-194` | **Register/Unregister reorder:** `Run` `select` между `register`/`unregister` недетерминирован. Handler шлёт `Register`, затем при быстром disconnect — `Unregister`. Если `removeClient` отработает **до** `registerClient`, `done` уже закрыт, а затем клиент всё равно попадает в `clients`/`userClients`. Повторного Unregister нет | Authenticated DoS / leak: зомби-`*Client` (буфер `send` 256) копятся при churn upgrade→close. `writePump` уже вышел по `done`, cleanup maps не повторяется до `shutdownAllClients` | `registerClient`: если `<-done` уже closed — не Store / сразу skip. Или sync Register (ack) до старта pumps. Или generation/`registered` flag: Unregister no-op только если register ещё не применялся, а поздний register отклонять |
| W3-A4-002 | P2 | `hub.go:204-209`; `hub.go:232-239`; `hub.go:323-361` | `subscribeClient` не проверяет наличие в `h.clients` и не смотрит `client.done`. Подписка send-and-forget: `Unregister` может обработаться **раньше** уже поставленного в канал `Subscribe` → зомби снова в `h.rooms` | Утечка членства комнаты (pointer на мёртвый Client). Broadcast безопасен (select на `done`), но `sync.Map` комнаты растёт; HI-07 follow-through закрыл «затирание», не late-subscribe | В `subscribeClient`: `if _, ok := h.clients.Load(connID); !ok { return }` + `select { case <-client.done: return; default: }`. Аналогично для `unsubscribeClient` по желанию |
| W3-A4-003 | P2 | `handler.go:148-151`; `handler.go:130-137` | `Handler(..., authz Authorizer)` при `authz == nil` подставляет `allowAllAuthorizer` — **fail-open** | Любой будущий/тестовый вызов без authz = broadcast-IDOR (регресс NF-02/SEC-H1). Prod `wiring_ws.go:37-54` передаёт authz, но API пакета провоцирует ошибку | `panic`/`slog.Error`+отказ upgrade при nil в non-test build; либо убрать nil-ветку и требовать явный `ExportAllowAllAuthorizer()` только из hubtest |
| W3-A4-004 | P2 | `handler.go:164-166`; `hub.go:24-27`; `bridge.go:230-257` | `SessionID` копируется в `Client`, но нигде не используется. `auth.session.{userID}.*` только `SendToUser` — Hub **не** делает `Unregister` по revoke | Instant revocation refresh в Redis не рвёт уже открытый WS: комнаты продолжают получать события до ручного close клиента (окно до pong timeout) | На `revoked`/logout: `Unregister` всех conn пользователя (или filter по `SessionID`); опционально закрыть conn. Тест: revoke → client.done закрыт, broadcast не доходит |
| W3-A4-005 | P2 | `pump_test.go:130-152`; `pump_test.go:230-251` | `TestReadPump_DeniedSubscribe` / `TestReadPump_AuthorizerError` ждут только выход `readPump`, **без** `ExportClientInRoom` / отсутствия delivery | Мутация «убрать `if !allowed` / игнорировать authErr» в `handler.go:241-246` **не краснеет** в пакете `ws`. Сильный тест есть в calls (`ws_authorizer_integration_a2007`), но NF-02 guard в `readPump` внутри списка A4 не mutation-proof | После deny/error: `assert.False(ExportClientInRoom(...))`; затем `BroadcastToRoom` + assert нет сообщения на `send` (как A2-007) |
| W3-A4-006 | P3 | `handler.go:153-157`; `handler.go:140-141` | Residual **A2-013:** access JWT в `?token=` | Токен в access/proxy logs, history, Referer | Short-lived WS ticket (одноразовый exchange), не сырой access JWT |
| W3-A4-007 | P3 | `hub.go:242-251`; `hub.go:349-353` | Пустые комнаты: `members.Delete(connID)`, но ключ `room` из `h.rooms` не удаляется при пустой inner map (в отличие от `userClients` empty-cleanup `:329-338`) | Долгий uptime + много уникальных room id → рост `h.rooms` пустыми `*sync.Map` | После Delete: Range empty → `h.rooms.Delete(room)` (с осторожностью к concurrent Store) |
| W3-A4-008 | P3 | `bridge.go:32-79`; `bridge.go:96-100` | `Start`: при ошибке N-й подписки предыдущие уже в `b.subs`, rollback/`Stop` не вызывается | Утечка NATS subscriptions, если вызывающий не сделает `Stop` (wiring делает `os.Exit(1)` — процесс умирает; переиспользование Bridge в тестах/других entrypoints рискованнее) | При err: `b.Stop(); return err` |
| W3-A4-009 | P3 | `handler.go:236-245` | Deny / authz error: только `slog.Warn`, клиенту нет `error`/`subscribe_denied` frame | Клиент не отличит сеть/баг от 403 membership; усложняет UX и отладку NF-02 | После deny отправить `ServerMessage` с типом ошибки (без утечки существования комнаты сверх необходимого) |
| W3-A4-010 | P3 | `export_hooks.go:82` | `ExportClientRooms` возвращает `c.rooms` **без** `mu` (в отличие от `ExportClientInRoom` `:84-88`) | Под `-race` параллельный Subscribe/Unregister + тест, читающий map → data race | Либо удалить экспорт, либо копировать map под `c.mu` |

## Чисто

| Файл / зона | Что проверено |
|-------------|----------------|
| `authorizer.go` NF-02 | Fail-closed: nil checker → deny; unknown prefix → deny; UUID parse user/entity; `splitRoom` отвергает пустой prefix/id. Делегирование channel/call/project без allow-all |
| `handler.go` readPump SEC-H1 | `TypeSubscribe` → `CanSubscribe` (timeout 3s) **до** `hub.Subscribe` (`:228-246`). Empty room skip. Unsubscribe без authz (только снятие своей подписки) — ок |
| `handler.go` origin | `CheckOrigin` + `CORS_ORIGINS` allowlist; `*` в env игнорируется; prod без origins → reject; DEV_MODE warn+allow; пустой Origin — осознанно для non-browser (`:82-106`) |
| `handler.go` auth upgrade | JWT query обязателен; ParseAccessToken; ME-13 deactivated → 403; CopyString до hijack |
| `handler.go` pumps | Один reader / один writer; `maxMessageSize=4096`; pong deadline; SEC-05: ping→send с `select` на `done` |
| `hub.go` SEC-05 / HI-07 | `send` не закрывается; только `done`+`closeOnce`; multi-tab `ConnectionID`; Subscribe/Unsubscribe через каналы Run; Broadcast/SendToUser select done/default drop |
| `hub.go` shutdown WS-01 | `beginShutdown` closes `stopping`; публичные методы не блокируются; `shutdownAllClients` → `removeClient` |
| `bridge.go` routing | Subjects → rooms: project/channel/call/notification/auth; malformed JSON skip; short subject no panic; `traced` consumer span |
| `protocol.go` | Константы типов; DTO JSON tags — без логики |
| `export_hooks.go` (кроме Rooms) | hubtest re-export Register/pumps/origin/bridge handlers — без prod-логики |
| `authorizer_test.go` | Table-driven: invalid user/room/entity, nil checkers, allow/deny/error для channel/call/project — asserts есть |
| `bridge_test.go` / `bridge_lifecycle_test.go` | Маршрутизация всех handlers + Start/Stop (6 subs) + live NATS; malformed/short subject |
| `handler_auth_test.go` | Missing/invalid token → 401; deactivated → 403 |
| `hub_shutdown_test.go` | Drain всех клиентов; post-stop Register/Subscribe не блокируют; concurrent cancel+ops под нагрузкой |
| `hub_test.go` | Multi-tab fan-out; SEC-05 concurrent senders+remove; remove cleans membership; broadcast/unsub; buffer full drop |
| `origin_test.go` | Allowlist / empty Origin / DEV_MODE / unset prod reject |
| `protocol_test.go` | JSON round-trip; NewClient ConnectionID unique; ExportRegister multi-tab (прямой registerClient — дополняет channel-backed в hub_test) |

**NF-02 вердикт (prod path):** при wiring `NewAuthorizer` + checkers membership (`wiring_ws.go`, вне списка, сверка) подписка на `channel:`/`project:`/`call:` без членства невозможна через `readPump`. Дыра — не пропуск CanSubscribe, а fail-open nil API, late Subscribe race и слабые deny-тесты в `ws` package.

**Pump race (SEC-05):** закрыт — тесты `hub_test.go` SEC-05 + shutdown suite; паники `send on closed channel` в текущем коде не видно.

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | **15 / 15 / 0** |
| Находок | **10** |
| P0 | 0 |
| P1 | 1 |
| P2 | 4 |
| P3 | 5 |

### Топ-3

1. **W3-A4-001 (P1)** — гонка Register/Unregister в `Hub.Run`: зомби-клиенты после быстрого disconnect (resource leak / DoS).
2. **W3-A4-002 (P2)** — late `subscribeClient` после Unregister реанимирует членство в `rooms`.
3. **W3-A4-005 (P2)** — deny/authz-error тесты `pump_test` без assert на отсутствие подписки → NF-02 регресс в `readPump` не ловится в домене ws.
