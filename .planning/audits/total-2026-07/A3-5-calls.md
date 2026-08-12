# W3 агент A5 — calls

> TASK-008 · READ-ONLY · 2026-07-28  
> Фокус: LiveKit room authz, recording webhook signature (SEC-02), scheduled series trash (FR-CAL-05), chat authz A2-003, raise hand A2-006, WS call authz A2-007.

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/calls/chat.go` | 98 | 0 | чисто |
| `backend/internal/calls/chat_mongo.go` | 74 | 0 | чисто |
| `backend/internal/calls/dto.go` | 97 | 1 | замечания |
| `backend/internal/calls/export_hooks.go` | 293 | 0 | чисто |
| `backend/internal/calls/handler.go` | 282 | 0 | чисто |
| `backend/internal/calls/handler_calls.go` | 467 | 2 | замечания |
| `backend/internal/calls/handler_recording.go` | 216 | 1 | замечания |
| `backend/internal/calls/handler_scheduled.go` | 200 | 2 | замечания |
| `backend/internal/calls/models.go` | 59 | 0 | чисто |
| `backend/internal/calls/recording.go` | 344 | 2 | замечания |
| `backend/internal/calls/repository.go` | 486 | 2 | замечания |
| `backend/internal/calls/scheduled.go` | 214 | 3 | замечания |
| `backend/internal/calls/service.go` | 289 | 1 | замечания |
| `backend/internal/calls/service_chat.go` | 143 | 1 | замечания |
| `backend/internal/calls/service_events.go` | 79 | 0 | чисто |
| `backend/internal/calls/service_room.go` | 137 | 1 | замечания |
| `backend/testframework/calls/chat_authz_a2003_test.go` | 98 | 0 | чисто |
| `backend/testframework/calls/chat_mem_test.go` | 119 | 0 | чисто |
| `backend/testframework/calls/chat_mongo_wave103_test.go` | 43 | 0 | чисто |
| `backend/testframework/calls/chat_test.go` | 157 | 0 | чисто |
| `backend/testframework/calls/handler_flow_test.go` | 343 | 1 | замечания |
| `backend/testframework/calls/handler_rbac_recording_test.go` | 257 | 0 | чисто |
| `backend/testframework/calls/handler_recording_test.go` | 348 | 0 | чисто |
| `backend/testframework/calls/handler_routes_test.go` | 106 | 1 | замечания |
| `backend/testframework/calls/handler_scheduled_coverage_test.go` | 133 | 0 | чисто |
| `backend/testframework/calls/handler_validation_test.go` | 259 | 0 | чисто |
| `backend/testframework/calls/raise_hand_a2006_test.go` | 60 | 0 | чисто |
| `backend/testframework/calls/recording_egress_mock_test.go` | 113 | 0 | чисто |
| `backend/testframework/calls/recording_service_coverage_test.go` | 128 | 0 | чисто |
| `backend/testframework/calls/recording_test.go` | 113 | 0 | чисто |
| `backend/testframework/calls/recording_webhook_test.go` | 272 | 0 | чисто |
| `backend/testframework/calls/repository_series_trash_test.go` | 54 | 0 | чисто |
| `backend/testframework/calls/repository_test.go` | 217 | 0 | чисто |
| `backend/testframework/calls/service_chat_flow_test.go` | 49 | 0 | чисто |
| `backend/testframework/calls/service_extra_test.go` | 189 | 1 | замечания |
| `backend/testframework/calls/service_flow_test.go` | 416 | 0 | чисто |
| `backend/testframework/calls/service_recording_scheduled_test.go` | 287 | 0 | чисто |
| `backend/testframework/calls/service_room_test.go` | 100 | 0 | чисто |
| `backend/testframework/calls/service_test.go` | 85 | 0 | чисто |
| `backend/testframework/calls/webhook_events_test.go` | 89 | 0 | чисто |
| `backend/testframework/calls/ws_authorizer_integration_a2007_test.go` | 232 | 0 | чисто |
| `backend/testframework/calls/ws_call_authz_a2007_test.go` | 60 | 0 | чисто |

Файлов в списке: 42. Прочитано: 42. Пропущено: 0.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|---------|-------------------|------------|
| W3-A5-001 | P1 | `handler_scheduled.go:111`; `scheduled.go:81-94` | `UpdateScheduledCall` не принимает `userID` и не проверяет creator/channel/project membership | Любой authenticated user с UUID звонка меняет чужой scheduled call (IDOR). Тесты только happy-path (`handler_routes_test.go:42-68`) | Передавать `userID` из handler; service: load call → `created_by` или channel owner/admin / project RBAC → иначе `ErrCallForbidden` + 403. Denial-тест stranger |
| W3-A5-002 | P1 | `handler_scheduled.go:137`; `scheduled.go:98-102` | `DeleteScheduledCall` без authz (только soft-delete по ID) | Любой authenticated user удаляет/кладёт в корзину чужую серию. Soft-delete корректен (FR-CAL-05), но без ownership | Тот же guard, что Update; 403 для stranger. Тест: create as A, DELETE as B → 403, `deleted_at` NULL |
| W3-A5-003 | P1 | `handler_scheduled.go:50`; `scheduled.go:34-61` | `CreateScheduledCall` пишет `projectID` из path + `channel_id` из body без проверки членства и без связи channel↔project | Stranger создаёт событие в чужом проекте/канале; channel может не принадлежать project из URL | `IsChannelMember` + `GetProjectIDByChannel == projectID` (или project RBAC). 403/400. Тест denial |
| W3-A5-004 | P1 | `handler_scheduled.go:77`; `scheduled.go:66-77` | `ListScheduledCalls` без membership: любой auth user читает расписание канала | Утечка title/time/rrule чужих каналов (IDOR read) | Перед list: `IsChannelMember`; иначе 403. Denial-тест |
| W3-A5-005 | P2 | `handler_scheduled.go:193`; `scheduled.go:171-210` | `CheckConflicts` принимает произвольные `participant_ids` без authz; при overlap помечает **всех** запрошенных PID как conflict (`scheduled.go:196-203`) | Сканирование чужих календарей + ложные conflict-строки для участников не из канала конфликта | Ограничить запрос себе / членам того же project; в conflict класть только PID, реально состоящих в `sc.ChannelID` |
| W3-A5-006 | P2 | `service_chat.go:83-92`; `chat.go:55-56` | Лимит 2000 рун только в `ChatRepository.SendMessage`; ephemeral-ветка обходит | DoS/раздувание NATS/WS payload огромным `text` при `ephemeral:true` | Общий `validateChatText` до ветвления ephemeral/persist; тест ephemeral >2000 → 400 |
| W3-A5-007 | P2 | `handler_calls.go:428-431` | `GuestJoinCall` читает `h.service.repo` напрямую (layering) | Handler зависит от EntRepository через orchestrator field; ломает слой Handler→Service | `service.GuestJoin(...)` с lookup+token; handler без repo |
| W3-A5-008 | P2 | `service.go:86`; `handler.go:39-40` | Invite-код `uuid.New().String()[:8]` + публичный `POST /calls/guest-join` без отдельного rate limit | 8 hex ≈ 32 бит; брутфорс активных комнат с выдачей LiveKit JWT без аккаунта | Длиннее код (полный UUID), rate limit на guest-join, lockout/backoff |
| W3-A5-009 | P2 | `recording.go:171-181` | После успешного `StartRoomCompositeEgress` ошибка `CreateRecording` оставляет orphan egress | Частичная запись: LiveKit пишет в S3, БД без метаданных; Stop по recording_id невозможен | Компенсация `StopEgress` при ошибке DB; или create row `pending` до Start |
| W3-A5-010 | P2 | `dto.go:44-49`; `scheduled.go:56` | `DurationMinutes`, `ReminderMinutes`, `ParticipantIDs` в Create DTO не персистятся | API принимает поля → silent drop; conflict check всегда 60 мин (`scheduled.go:191`) | Persist или 400 unknown/unsupported; использовать duration в FindConflictingCalls |
| W3-A5-011 | P2 | `repository.go:394-413` | `UpdateScheduledCall` через `UpdateOneID` **без** `DeletedAtIsNil` | Можно мутировать soft-deleted серию (title/rrule/`is_active`) в обход корзины | `Where(DeletedAtIsNil())`; 404 если в корзине. Тест после Delete → Update → not found |
| W3-A5-012 | P2 | `service_room.go:34-36` | `finalizeCreate`: ошибка `AddParticipant` только `slog.Error`, ответ 201 с токеном | Создатель получает JWT, но не в `call_participant` → RaiseHand/WS chat authz могут 403 | Возвращать error из `finalizeCreate` / rollback room+call |
| W3-A5-013 | P3 | `handler_recording.go:127-139`; `handler_recording.go:190-207` | `if h.rbacSvc != nil` — при nil RBAC list/download без проверки (документировано «только тесты») | Ошибочный wiring → IDOR на recordings. Prod сейчас передаёт rbac (`wiring_integrations.go:54`) | Fail-closed: nil → 500; или обязательный RBAC в конструкторе |
| W3-A5-014 | P3 | `repository.go:248-261` | `GetChannelMemberRole` не фильтрует soft-deleted channel (в отличие от `IsChannelMember:240`) | После soft-delete канала бывший admin теоретически `EndCall` | `HasChannelWith(DeletedAtIsNil())` как в `IsChannelMember` |
| W3-A5-015 | P3 | `service_extra_test.go:24-32` | `ListProjectRecordings_Success` через `ExportNewHandlerWithRecSvc` **без** RBAC | Тест зелёный и при снятии CR-02 guard в handler | Использовать RBAC-фикстуру из `handler_rbac_recording_test.go` |
| W3-A5-016 | P3 | `handler_routes_test.go:42-94`; `handler_flow_test.go:136-180` | Scheduled HTTP-тесты только success; нет denial stranger | Регресс authz (W3-A5-001…004) не ловится | Добавить stranger 403 suite на Create/List/Update/Delete |

## Чисто

| Файл / зона | Что проверено |
|-------------|----------------|
| `handler.go:156-273` webhook | JWT HMAC + `iss` via `constantTimeStringEqual` + **обязательный** `sha256` body hash (SEC-02); empty/wrong/tampered → false. Нет skip claim |
| `recording_webhook_test.go` | Valid / wrong secret / no auth / wrong iss / no sha256 / empty sha256 / tampered body / wrong hash — asserts 200/401 |
| `webhook_events_test.go` | `room_started`/`room_finished` обновляют Ent; invalid identity → 200 без panic |
| Chat A2-003 `service_chat.go:31-56` | Channel: participant OR member; Instant: participant OR created_by; иначе `ErrCallForbidden`. Handler мапит 403 |
| `chat_authz_a2003_test.go` | Stranger send/list 403; member allowed; instant creator/participant/stranger — mutation notes в комментариях |
| Raise hand A2-006 `service.go:217-236` | `IsActiveParticipant` до LiveKit; non-participant → `ErrCallForbidden` без SDK |
| `raise_hand_a2006_test.go` | HTTP 403 + service ErrorIs + participant reaches LiveKit (не Forbidden) |
| WS A2-007 `repository.go:208-229` | Channel: member OR active participant; Instant: только active participant; missing → false |
| `ws_call_authz_a2007_test.go` + `ws_authorizer_integration_a2007_test.go` | Repo + prod `ws.NewAuthorizer` + readPump stranger deny / participant allow |
| LiveKit room authz | Token `VideoGrant{RoomJoin, Room: roomName}` (`service_room.go:125-136`); Join/Create требуют channel membership / instant creator (`service.go:54-61`, `123-143`); EndCall creator\|owner\|admin |
| Series trash `repository.go:416-424` | Soft-delete `deleted_at` + `is_active=false`; List/GetDue фильтруют `DeletedAtIsNil` |
| `repository_series_trash_test.go` | Create RRULE series → Delete → list empty, due excludes, raw row has DeletedAt |
| Recording A2-004 `recording.go:25-39` | Start/Stop: channel member / instant creator; Stop сверяет `rec.CallID == callID` |
| `handler_rbac_recording_test.go` / `recording_service_coverage_test.go` | Non-member 403; download/list RBAC на реальном Ent+RBAC |
| `chat.go` / `chat_mongo.go` | Rune limit 2000 (persist), list cap 200, cursor.Close, indexes |
| `service_events.go` | NATS subjects инкапсулированы; nil-safe publisher |
| `export_hooks.go` | hubtest re-exports / mem chat store — без prod-логики |
| `models.go` | DTO звонков/guest — UUID, без секретов |
| Остальные tests (mem/mongo/flow/validation/room/recording) | Asserts присутствуют; webhook/raise/chat/ws denial на слое guard |

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | **42 / 42 / 0** |
| Находок | **16** |
| P0 | 0 |
| P1 | 4 |
| P2 | 8 |
| P3 | 4 |

### Топ-3

1. **W3-A5-001…004 (P1)** — scheduled calls: Create/List/Update/Delete без ownership/membership → полный IDOR на расписания.
2. **W3-A5-006 (P2)** — ephemeral chat обходит лимит 2000 символов → раздувание real-time fanout.
3. **W3-A5-008 (P2)** — короткий invite + unauthenticated guest-join → брутфорс LiveKit JWT.

### Фокус-вердикт

| Тема | Статус |
|------|--------|
| LiveKit room authz | OK (membership до JWT; room-scoped grant). Guest-join — отдельный риск (W3-A5-007/008) |
| Recording webhook signature | OK (SEC-02 закрыт, тесты полные) |
| Scheduled series trash | Soft-delete OK; **authz на CRUD отсутствует** |
| Chat authz A2-003 | OK + mutation-ready tests |
| Raise hand A2-006 | OK + mutation-ready tests |
| WS call authz A2-007 | OK + Authorizer/readPump integration |
