# Messenger / WS / Calls / Notification

**Аудит:** READ-ONLY · зона A2 · 2026-07-26  
**Ветка:** `cursor-grok-4.5`  
**Пакеты:** `backend/internal/{messenger,ws,calls,notification}` (+ wiring WS authz для call-комнат)

## Что проверено

| Область | Файлы / фокус |
|---------|----------------|
| WS gateway | `handler.go` (origin, JWT query, read/write pump), `hub.go` (register/subscribe/done), `authorizer.go` (fail-closed prefixes), `bridge.go` (NATS→room) |
| WS wiring call rooms | `backend/cmd/server/wiring_ws.go` (IsCallParticipantOrChannelMember) |
| Messenger messages | membership на send/list/edit/delete/forward/pins/polls; cursor `before` в Mongo |
| Messenger search / typing / presence | `service_search.go`, `handler_search.go`, `PublishTyping`, Redis presence thresholds |
| Calls join/end | HI-03 membership на Join/End; CreateCall; guest invite; RaiseHand; chat |
| LiveKit | webhook JWT+sha256 (`handler.go`), token generation (`service_room.go`, guest token) |
| Recording | Start/Stop/List/Download RBAC gaps |
| Notification push | `ValidateOutboundURL`, SafeHTTPClient, Save/RemoveSubscription |
| Notification consumer | JetStream ACK policy, handlers, debounce |
| Telegram | sender ctx-cancel, bind polling |

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это баг | Как чинить |
|----|----------|-------------|----------|----------------|------------|
| A2-001 | P0 | `backend/internal/messenger/service_search.go:11-29` · `handler_search.go:24-37` · `mongo_repository.go:282-288` | Поиск сообщений без user-scope / membership | Любой authenticated user читает тексты из **всех** каналов (или чужого `channel_id`). Аналог NF-03 / bases IDOR. | Передавать `userID`; без `channel_id` — искать только по каналам `IsMember`; с `channel_id` — `IsMember` → иначе 403. |
| A2-002 | P1 | `backend/internal/calls/service.go:49-65` | `CreateCall` не проверяет членство в канале | Любой user создаёт звонок/LiveKit room в чужом канале → NATS `calls.channel.*.started` у членов канала. | Перед `provisionRoom`/`CreateCall`: `IsChannelMember`; иначе `ErrCallForbidden`. |
| A2-003 | P1 | `backend/internal/calls/service_chat.go:33-79` · `handler_calls.go:311-377` | Chat звонка: send/list без participant/channel authz | Зная `callId`, любой user шлёт/читает чат и пушит в `calls.chat.{id}` → WS room `call:{id}`. | Требовать participant **или** `IsChannelMember` (для channel-call); instant — только participant/created_by. |
| A2-004 | P1 | `backend/internal/calls/recording.go:89-97` · `173-177` · `handler_recording.go:32-48` · `63-80` | StartRecording без membership; StopRecording без user/RBAC | Любой authenticated стартует egress на чужой active call; любой останавливает любую запись по `recording_id`. | Start: channel member + роль/permission; Stop: тот же authz + ownership/`userID` в сервисе. |
| A2-005 | P1 | `backend/internal/notification/push.go:230-237` · `252-261` · `handler.go:340-360` | Push: hijack endpoint (`SetUserID`) и unsubscribe без scope по user | Знание endpoint чужой подписки → перепривязка к себе (DoS уведомлений жертвы) или удаление чужой подписки (`RemoveSubscription` без `userID`; handler игнорирует свой userID: `_, err := getUserID`). | Update/Delete только `Endpoint AND UserID=caller`; при чужом endpoint — 404/403 без перепривязки. |
| A2-006 | P1 | `backend/internal/calls/handler_calls.go:268-294` | `RaiseHand` без проверки участия в звонке | Любой user с `callId` дергает LiveKit `UpdateParticipantMetadata` для своего identity / зондирует API. | Проверять active participant (или channel member) до `SetRaiseHand`. |
| A2-007 | P1 | `backend/cmd/server/wiring_ws.go:49-52` · `ws/authorizer.go:67-71` | Instant-call: WS `call:{id}` всегда deny | Checker: `ChannelID == nil → false`. Участники instant-звонка не подписываются на chat room → real-time чат мёртв (NATS `calls.chat.*` есть). | Для instant: `IsCallParticipant(callID, userID)` (или created_by); channel-call оставить member OR participant. |
| A2-008 | P2 | `backend/internal/calls/handler_calls.go:137-154` · `service.go:185-193` | `GetActiveCall` без membership | Утечка метаданных активного звонка (id, room_name, created_by, counts) по любому `channelId`. | `IsChannelMember` до lookup; иначе 403/пустой ответ без различия. |
| A2-009 | P2 | `backend/internal/messenger/handler_presence.go:23-37` · `service.go:82-97` | `PublishTyping` без `IsMember` | Не-участник инжектит typing в NATS `messenger.channel.{id}.typing` → видят подписанные члены канала. | `IsMember(channelID, userID)` перед publish. |
| A2-010 | P2 | `backend/internal/notification/consumer.go:132-170` | ACK всегда, даже после ошибок Dispatch | Комментарий «ACK только после успешной» ложен: handlers глотают err и `processEvent` всё равно `msg.Ack()` → потеря уведомлений до MaxDeliver. | Handlers возвращают error; при fail — `Nak`/`NakWithDelay`, ACK только на success. |
| A2-011 | P2 | `backend/internal/calls/service.go:74` · `handler_calls.go:391-424` | Invite-код = 8 hex (`uuid[:8]`), guest-join без auth | ~32 бита энтропии + публичный `POST /calls/guest-join` → перебор invite → LiveKit token. | Криптостойкий код ≥128 бит; rate-limit guest-join; опционально TTL/max uses. |
| A2-012 | P2 | `backend/internal/messenger/presence.go:19-22` · `61-64` · `158-161` | `away` vs cleanup: stale=90s, away threshold=5m | Cleanup удаляет heartbeat >90s → статус `away` practically unreachable; comment противоречит коду. | Выровнять пороги (например away=90s, stale=5m) или убрать cleanup раньше away. |
| A2-013 | P3 | `backend/internal/ws/handler.go:140-156` | JWT access token в query `?token=` | Токен попадает в access/proxy logs, history, Referer при ошибках — стандартный WS risk. | Short-lived WS ticket (одноразовый exchange), не сырой access JWT в query. |
| A2-014 | P3 | `backend/internal/messenger/handler.go:27-36` | `queryInt` без clamp min/max | Отрицательный/огромный `limit` доходит до service; Mongo clamps, но handler-контракт дырявый. | Clamp в `queryInt` или callers (`1..100`). |

## Чисто (проверено, проблем нет)

- **WS subscribe authz (channel/project):** `readPump` → `Authorizer.CanSubscribe` с timeout; unknown prefix fail-closed (`authorizer.go`, `handler.go:228-246`).
- **Hub concurrency:** `done` вместо close(`send`); multi-tab по `ConnectionID`; subscribe через канал Run (`hub.go`).
- **CheckOrigin:** CORS_ORIGINS allowlist; empty Origin allowed (non-browser); DEV_MODE warn (`handler.go:82-106`).
- **LiveKit webhook:** HMAC JWT + mandatory `sha256` body claim; API key via constant-time hash compare; no APIKey in logs (`calls/handler.go:217-282`).
- **Push SSRF:** `ValidateOutboundURL` на save+send; `pushHTTPClient` = SafeHTTPClient (`notification/push.go:19-25`, `151-165`, `218-220`).
- **Messenger message path authz:** Send/List/Edit/Delete/Forward/Thread/Pins/Polls — `IsMember` (`service_messages.go` и др.).
- **Cursor chat pagination:** `before` RFC3339 + Mongo `$lt` + limit clamp 1..100 (`mongo_repository.go:94-110`, `handler_messages.go:74-86`) — off-by-one по курсору не найден.
- **JoinCall/EndCall HI-03:** channel membership / creator+admin (`service.go:111-181`).
- **List/Download recordings RBAC:** `PermProjectView` когда `rbacSvc != nil` (`handler_recording.go:108-121`, `171-188`).
- **Telegram sender:** ctx-aware interrupt на shutdown (`telegram.go:107-127`).
- **Notification list/mark/delete:** scoped by `userID` в repository.

## Оценка зоны: 5/10

Realtime-ядро (WS membership для channel/project, Hub lifecycle, LiveKit webhook, push SSRF) выглядит зрелым, но **P0 search IDOR** и серия **P1 authz-дыр в calls/recording/push/typing** делают зону непригодной для «security-ready» без точечных фиксов.
