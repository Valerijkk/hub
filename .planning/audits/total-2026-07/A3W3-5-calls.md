# A3 W3 агент A5 — calls (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Baseline: [A2W3-5-calls.md](A2W3-5-calls.md) · A1: [A3-5-calls.md](A3-5-calls.md)  
> Assignment: [W3-ASSIGNMENTS.md](W3-ASSIGNMENTS.md) § A5 · Периметр: **42** · Прочитано: **42/42** · Пропущено: **0**  
> Remediation: TASK-110 (A2W3-5-001/002/003) · TASK-018/042/061/064 (W3-A5 cascade)

## Сверка с A2 (A2W3-5-*)

| A2 ID | Tag | Статус | Комментарий |
|-------|-----|--------|-------------|
| A2W3-5-001 | FIXED-OK | закрыт | `compensateOrphanEgress` после fail `CreateRecording` (`recording.go:179–214`); IT `recording_room_guest_task110_test.go` |
| A2W3-5-002 | FIXED-OK | закрыт | `finalizeCreate` → err + `DeleteRoom` + `HardDeleteCall` (`service_room.go:29–52`); IT task110 |
| A2W3-5-003 | FIXED-OK | закрыт | crypto/rand base62×12 (~71 bit) within MaxLen(12) (`service.go:19–37,108–116`); IT task110 + guest_conflicts |

**REGRESS:** 0 · **A2 open residual:** 0 (A2 baseline был 0 open)

### Сверка A1 / FIXED-OK cascade (подтверждение на HEAD)

| Stage A / A2 FIXED class | Tag | Evidence |
|--------------------------|-----|----------|
| W3-A5-001…004 scheduled authz | FIXED-OK | `scheduled.go` requireChannelMember + authorizeScheduledCallMutate; handler передаёт userID; IT `scheduled_authz_task018_test.go` |
| W3-A5-005 CheckConflicts | FIXED-OK | `filterConflictParticipants` + PID∈channel (`scheduled.go:257–347`) |
| W3-A5-006 ephemeral ≤2000 | FIXED-OK | `service_chat.go:81–84` до ветвления |
| W3-A5-007 GuestJoin→service | FIXED-OK | `service.go:153–172`; handler без repo |
| W3-A5-008 guest RL + entropy | FIXED-OK | `handler.go:42–47` RL 10/min; base62×12 (полный UUID — ADR MaxLen) |
| W3-A5-010 Duration/Reminder/PIDs | FIXED-OK | `repository.go:449–473` Create persist |
| W3-A5-011 soft-delete Update | FIXED-OK | `UpdateScheduledCall` `DeletedAtIsNil` (`repository.go:498–521`) |
| W3-A5-013 nil RBAC → 500 | FIXED-OK | `handler_recording.go:126–130,191–195`; IT `nil_rbac_recordings_task064_test.go` |
| W3-A5-014 member role soft-delete | FIXED-OK | `GetChannelMemberRole` HasChannelWith DeletedAtIsNil |
| W3-A5-015/016 denial suites | FIXED-OK | RBAC fixture в `service_extra_test.go`; scheduled_authz suite |
| Chat A2-003 / raise A2-006 / WS A2-007 / webhook SEC-02 / series trash | FIXED-OK | без регрессии |

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/calls/chat.go` | 98 | 0 | чисто · rune 2000 · list cap 200 |
| `backend/internal/calls/chat_mongo.go` | 74 | 0 | чисто · cursor.Close · indexes |
| `backend/internal/calls/dto.go` | 97 | 0 | FIXED-OK DurationMinutes persist |
| `backend/internal/calls/export_hooks.go` | 313 | 0 | hubtest re-exports · чисто |
| `backend/internal/calls/handler.go` | 289 | 0 | FIXED-OK guest RL · SEC-02 webhook |
| `backend/internal/calls/handler_calls.go` | 436 | 1 | GuestJoin OK; guest_name MaxLen gap |
| `backend/internal/calls/handler_recording.go` | 220 | 0 | FIXED-OK nil RBAC → 500 |
| `backend/internal/calls/handler_scheduled.go` | 239 | 0 | FIXED-OK userID+authz+conflicts |
| `backend/internal/calls/models.go` | 59 | 0 | чисто |
| `backend/internal/calls/recording.go` | 367 | 1 | orphan egress OK; Nil project_id gap |
| `backend/internal/calls/repository.go` | 593 | 1 | soft-delete/role OK; FindConflicts channel filter |
| `backend/internal/calls/scheduled.go` | 363 | 0 | FIXED-OK CRUD authz + conflicts |
| `backend/internal/calls/service.go` | 340 | 2 | invite+GuestJoin OK; orphan room on DB fail; title MaxLen |
| `backend/internal/calls/service_chat.go` | 148 | 0 | FIXED-OK ephemeral ≤2000 · A2-003 |
| `backend/internal/calls/service_events.go` | 79 | 0 | чисто · nil-safe NATS |
| `backend/internal/calls/service_room.go` | 158 | 0 | FIXED-OK finalizeCreate rollback |
| `backend/testframework/calls/chat_authz_a2003_test.go` | 98 | 0 | чисто |
| `backend/testframework/calls/chat_mem_test.go` | 119 | 0 | чисто |
| `backend/testframework/calls/chat_mongo_wave103_test.go` | 43 | 0 | чисто |
| `backend/testframework/calls/chat_test.go` | 157 | 0 | чисто |
| `backend/testframework/calls/handler_flow_test.go` | 345 | 0 | FIXED-OK (+ scheduled_authz suite) |
| `backend/testframework/calls/handler_rbac_recording_test.go` | 257 | 0 | чисто |
| `backend/testframework/calls/handler_recording_test.go` | 348 | 0 | чисто |
| `backend/testframework/calls/handler_routes_test.go` | 108 | 0 | чисто |
| `backend/testframework/calls/handler_scheduled_coverage_test.go` | 150 | 0 | чисто |
| `backend/testframework/calls/handler_validation_test.go` | 259 | 0 | чисто |
| `backend/testframework/calls/raise_hand_a2006_test.go` | 60 | 0 | чисто |
| `backend/testframework/calls/recording_egress_mock_test.go` | 113 | 0 | чисто |
| `backend/testframework/calls/recording_service_coverage_test.go` | 128 | 0 | чисто |
| `backend/testframework/calls/recording_test.go` | 113 | 0 | чисто |
| `backend/testframework/calls/recording_webhook_test.go` | 272 | 0 | чисто SEC-02 |
| `backend/testframework/calls/repository_series_trash_test.go` | 54 | 0 | чисто |
| `backend/testframework/calls/repository_test.go` | 218 | 0 | чисто |
| `backend/testframework/calls/service_chat_flow_test.go` | 49 | 0 | чисто |
| `backend/testframework/calls/service_extra_test.go` | 189 | 0 | FIXED-OK RBAC fixture (W3-A5-015) |
| `backend/testframework/calls/service_flow_test.go` | 424 | 0 | чисто |
| `backend/testframework/calls/service_recording_scheduled_test.go` | 303 | 0 | чисто |
| `backend/testframework/calls/service_room_test.go` | 100 | 0 | чисто |
| `backend/testframework/calls/service_test.go` | 85 | 0 | чисто |
| `backend/testframework/calls/webhook_events_test.go` | 89 | 0 | чисто |
| `backend/testframework/calls/ws_authorizer_integration_a2007_test.go` | 242 | 0 | чисто |
| `backend/testframework/calls/ws_call_authz_a2007_test.go` | 60 | 0 | чисто |

Файлов в списке: **42**. Прочитано: **42**. Пропущено: **0**.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|---------|-------------------|------------|
| A3W3-A5-001 | NEW | P2 | `service.go:84–92`, `120–123` | После `provisionRoom` ошибка `CreateCall` / `CreateInstantCall` (DB) возвращается **без** `DeleteRoom` | Тот же класс, что закрытый A2W3-5-001/002: LiveKit-комната остаётся orphan (нет call row → никто не EndCall/room_finished cleanup). Invite-fail путь уже чистит room (`108–116`) — асимметрия | На DB-err: `DeleteRoom(roomName)` (как invite-fail / `rollbackCreate`); regression IT с mock RoomService |
| A3W3-A5-002 | NEW | P3 | `repository.go:554–560` | `FindConflictingCalls` берёт `ChannelMember` **без** `HasChannelWith(DeletedAtIsNil)` | `ShareActiveChannel` / `ListChannelMemberIDs` / `IsChannelMember` фильтруют soft-deleted channel; conflict-query — нет → title/schedule soft-deleted каналов в conflict response | Добавить `HasChannelWith(channel.DeletedAtIsNil())` в members-query; тест soft-deleted channel не попадает в conflicts |
| A3W3-A5-003 | NEW | P3 | `recording.go:128–131`; `repository.go:335–337` | `GetProjectIDByChannel` при `ch.ProjectID == nil` → `(uuid.Nil, nil)`; `StartRecording` не проверяет Nil | Ent Recording.project_id Required edge → FK fail → **500**, не 400; канал без project теоретически существует | Repo: not-found/ErrInvalidInput если ProjectID nil; service: `projectID == uuid.Nil` → явная ошибка; IT |
| A3W3-A5-004 | NEW | P3 | `service.go:117–120`; `handler_calls.go:422–424`; schema `call.Title` MaxLen(255) | `CreateInstantCall` title и `GuestJoin` guest_name без лимита длины → Ent/JWT раздувание → **500**/тяжёлый token | Класс MaxLen→500 (как A3W2-A5-003 bases): клиентский ввод должен давать 400 | `utf8.RuneCountInString` ≤255 (title) / разумный cap guest_name (напр. 100) → `ErrInvalidInput`/400; regression |

## Чисто (подтверждено после remediation)

- **LiveKit room JWT** room-scoped (`VideoGrant{RoomJoin, Room}`); Join/Create membership; EndCall creator\|owner\|admin.
- **Recording** Start/Stop callID match + channel authz; orphan egress `StopEgress` (TASK-110); list/download fail-closed без RBAC.
- **Webhook SEC-02:** HMAC JWT + обязательный sha256 body; empty/wrong/tampered → 401.
- **Scheduled CRUD authz** Create/List membership + project↔channel; Update/Delete creator\|owner\|admin; soft-delete trash FR-CAL-05.
- **CheckConflicts** shared-channel filter + conflict PID∈channel; stranger → 403.
- **Chat A2-003** participant/member/created_by; ephemeral rune ≤2000 до ветвления.
- **Raise hand A2-006** IsActiveParticipant до LiveKit.
- **WS A2-007** channel member OR active participant; instant — только participant.
- **Invite** base62×12 crypto/rand; guest-join RL 10/min; GuestJoin на service-слое.
- **finalizeCreate** rollback room+HardDeleteCall при AddParticipant fail.

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 1 |
| P3 | 3 |
| **Total** | **4** |

### Теги vs A2

| Tag | Count |
|-----|------:|
| FIXED-OK (A2 active) | 3 (`A2W3-5-001`…`003`) |
| FIXED-OK (A1 cascade classes) | 16 |
| RESIDUAL | 0 |
| NEW | 4 (`A3W3-A5-001`…`004`) |
| REGRESS | 0 |

A2 baseline: **0** open → FIXED-OK 3 · RESIDUAL 0 · NEW 4 · REGRESS 0 · open **4** (P2=1 · P3=3).

**Топ ID:** A3W3-A5-001 (orphan LiveKit room on DB create fail) · A3W3-A5-002 (FindConflicts soft-delete channel) · A3W3-A5-003 (Nil project_id → 500) · A3W3-A5-004 (title/guest_name MaxLen→500).
