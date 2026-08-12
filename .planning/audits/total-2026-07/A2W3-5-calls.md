# A2 W3 агент 5 — calls (re-audit)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> Зеркало: `W3-ASSIGNMENTS.md` A5 · Stage A: `A3-5-calls.md`

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/calls/chat.go` | 81 | 0 | чисто |
| `backend/internal/calls/chat_mongo.go` | 62 | 0 | чисто |
| `backend/internal/calls/dto.go` | 85 | 0 | FIXED-OK DurationMinutes persist |
| `backend/internal/calls/export_hooks.go` | 238 | 0 | чисто |
| `backend/internal/calls/handler.go` | 254 | 0 | FIXED-OK guest RL |
| `backend/internal/calls/handler_calls.go` | 391 | 0 | FIXED-OK GuestJoin→service |
| `backend/internal/calls/handler_recording.go` | 199 | 0 | FIXED-OK nil RBAC → 500 |
| `backend/internal/calls/handler_scheduled.go` | 214 | 0 | FIXED-OK userID+authz |
| `backend/internal/calls/models.go` | 50 | 0 | чисто |
| `backend/internal/calls/recording.go` | 303 | 0 | FIXED-OK orphan egress StopEgress (TASK-110) |
| `backend/internal/calls/repository.go` | 507 | 0 | FIXED-OK soft-delete + member role |
| `backend/internal/calls/scheduled.go` | 323 | 0 | FIXED-OK CRUD authz + conflicts |
| `backend/internal/calls/service.go` | 287 | 0 | FIXED-OK invite base62×12 (TASK-110; MaxLen UUID — ADR) |
| `backend/internal/calls/service_chat.go` | 136 | 0 | FIXED-OK ephemeral ≤2000 |
| `backend/internal/calls/service_events.go` | 69 | 0 | чисто |
| `backend/internal/calls/service_room.go` | 130 | 0 | FIXED-OK finalizeCreate rollback (TASK-110) |
| `backend/testframework/calls/chat_authz_a2003_test.go` | 82 | 0 | чисто |
| `backend/testframework/calls/chat_mem_test.go` | 101 | 0 | чисто |
| `backend/testframework/calls/chat_mongo_wave103_test.go` | 35 | 0 | чисто |
| `backend/testframework/calls/chat_test.go` | 131 | 0 | чисто |
| `backend/testframework/calls/handler_flow_test.go` | 296 | 0 | FIXED-OK (+ scheduled_authz suite) |
| `backend/testframework/calls/handler_rbac_recording_test.go` | 220 | 0 | чисто |
| `backend/testframework/calls/handler_recording_test.go` | 295 | 0 | чисто |
| `backend/testframework/calls/handler_routes_test.go` | 92 | 0 | чисто |
| `backend/testframework/calls/handler_scheduled_coverage_test.go` | 128 | 0 | чисто |
| `backend/testframework/calls/handler_validation_test.go` | 213 | 0 | чисто |
| `backend/testframework/calls/raise_hand_a2006_test.go` | 51 | 0 | чисто |
| `backend/testframework/calls/recording_egress_mock_test.go` | 97 | 0 | чисто |
| `backend/testframework/calls/recording_service_coverage_test.go` | 107 | 0 | чисто |
| `backend/testframework/calls/recording_test.go` | 97 | 0 | чисто |
| `backend/testframework/calls/recording_webhook_test.go` | 226 | 0 | чисто SEC-02 |
| `backend/testframework/calls/repository_series_trash_test.go` | 42 | 0 | чисто |
| `backend/testframework/calls/repository_test.go` | 178 | 0 | чисто |
| `backend/testframework/calls/service_chat_flow_test.go` | 39 | 0 | чисто |
| `backend/testframework/calls/service_extra_test.go` | 156 | 0 | чисто |
| `backend/testframework/calls/service_flow_test.go` | 356 | 0 | чисто |
| `backend/testframework/calls/service_recording_scheduled_test.go` | 260 | 0 | чисто |
| `backend/testframework/calls/service_room_test.go` | 88 | 0 | чисто |
| `backend/testframework/calls/service_test.go` | 69 | 0 | чисто |
| `backend/testframework/calls/webhook_events_test.go` | 70 | 0 | чисто |
| `backend/testframework/calls/ws_authorizer_integration_a2007_test.go` | 214 | 0 | чисто |
| `backend/testframework/calls/ws_call_authz_a2007_test.go` | 50 | 0 | чисто |

Файлов в списке: **42**. Прочитано: **42**. Пропущено: **0**.

## Находки (открытые)

| ID | Tag | Severity | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|----------|-------------|----------|---------------|------------|
| — | — | — | — | Нет открытых (A2W3-5-001..003 → TASK-110) | — | — |

## FIXED-OK

| Stage A ID | Evidence |
|------------|----------|
| A2W3-5-001 | TASK-110: `compensateOrphanEgress` после fail `CreateRecording`; IT `recording_room_guest_task110_test.go` |
| A2W3-5-002 | TASK-110: `finalizeCreate` → err + `DeleteRoom` + `HardDeleteCall`; IT task110 |
| A2W3-5-003 | TASK-110: crypto/rand base62×12 (~71 bit) within MaxLen(12); IT task110 + guest_conflicts |
| W3-A5-001…004 | scheduled Create/List/Update/Delete + membership/creator; `scheduled_authz_task018_test.go` |
| W3-A5-005 | CheckConflicts filter shared channels + conflict PID∈channel |
| W3-A5-006 | ephemeral rune limit до ветвления |
| W3-A5-007 | `GuestJoin` на service; handler без repo |
| W3-A5-008 | guest RL + base62×12 entropy (TASK-110; full UUID needs MaxLen migrate) |
| W3-A5-010 | DurationMinutes/Reminder/Participants persist |
| W3-A5-011 | UpdateScheduledCall `DeletedAtIsNil` |
| W3-A5-013 | nil rbac → 500 (`nil_rbac_recordings_task064`) |
| W3-A5-014 | GetChannelMemberRole + soft-delete channel |
| W3-A5-015/016 | denial suites + RBAC fixtures |
| Chat A2-003 / raise A2-006 / WS A2-007 / webhook SEC-02 / series trash | без регрессии |

## Чисто

- LiveKit room-scoped JWT; Join/Create membership; EndCall creator\|admin; recording Start/Stop callID match; webhook HMAC+sha256 body; chat authz participant/member.

## Сводка агента

| Метрика | Значение |
|---------|----------|
| Файлов | 42/42/0 |
| Открытых | **0** |
| NEW | 0 |
| REGRESS | 0 |
| RESIDUAL | 0 |
| FIXED-OK | 19 (+A2W3-5-001..003 via TASK-110) |
