# A2 W3 агент 4 — ws (re-audit)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> Зеркало: `W3-ASSIGNMENTS.md` A4 · Stage A: `A3-4-ws.md`

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/ws/authorizer.go` | 81 | 0 | чисто NF-02 fail-closed |
| `backend/internal/ws/bridge.go` | 240 | 1 | RESIDUAL auth publish; FIXED Start rollback + revoke Unregister |
| `backend/internal/ws/export_hooks.go` | 90 | 0 | FIXED-OK ExportClientRooms under mu |
| `backend/internal/ws/handler.go` | 293 | 1 | RESIDUAL JWT `?token=`; FIXED nil authz + deny frames |
| `backend/internal/ws/hub.go` | 398 | 0 | FIXED-OK register race, late subscribe, empty rooms |
| `backend/internal/ws/protocol.go` | 32 | 0 | чисто (+ EventSubscribeDenied) |
| `backend/testframework/ws/authorizer_test.go` | 194 | 0 | чисто |
| `backend/testframework/ws/bridge_lifecycle_test.go` | 53 | 0 | чисто |
| `backend/testframework/ws/bridge_test.go` | 179 | 0 | чисто |
| `backend/testframework/ws/handler_auth_test.go` | 88 | 0 | чисто |
| `backend/testframework/ws/hub_shutdown_test.go` | 137 | 0 | чисто |
| `backend/testframework/ws/hub_test.go` | 304 | 0 | чисто |
| `backend/testframework/ws/origin_test.go` | 45 | 0 | чисто |
| `backend/testframework/ws/protocol_test.go` | 136 | 0 | чисто |
| `backend/testframework/ws/pump_test.go` | 235 | 0 | FIXED-OK deny → ExportClientInRoom |

Файлов в списке: **15**. Прочитано: **15**. Пропущено: **0**.

## Находки (открытые)

| ID | Tag | Severity | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|----------|-------------|----------|---------------|------------|
| A2W3-4-001 | RESIDUAL | P3 | `handler.go:140-157` | Access JWT всё ещё в `?token=` | W3-A4-006 deferred (TASK-064 / product ADR): токен в access/proxy logs, Referer | Short-lived WS ticket exchange |
| A2W3-4-002 | RESIDUAL | P2 | `bridge.go:249-252` + auth `service_session.go:265-276` (вне списка, сверка) | WS `UnregisterBySession` готов на `revoked`/`logout`, но **auth Logout/Revoke не публикует** `auth.session.{uid}.*` | W3-A4-004 half-fixed (TASK-042): instant revoke refresh не рвёт открытый WS end-to-end | Publish NATS на Logout/Revoke с `session_id`; тест revoke→client.done |

## FIXED-OK

| Stage A ID | Evidence |
|------------|----------|
| W3-A4-001 | `registerClient`: `select <-done` → skip Store; `hub_register_unregister_race_test.go` |
| W3-A4-002 | `subscribeClient` checks `done`; late Subscribe не реанимирует |
| W3-A4-003 | `authz == nil` → 503 refuse upgrade; `handler_nil_authz_test.go` |
| W3-A4-004 (WS) | `UnregisterBySession` + bridge on revoked/logout |
| W3-A4-005 | pump deny/error: `ExportClientInRoom` / no delivery (`authz_deny_room_task037`) |
| W3-A4-007 | empty room → `h.rooms.Delete` |
| W3-A4-008 | `Start` fail → `Stop()` rollback |
| W3-A4-009 | `EventSubscribeDenied` / `EventSubscribeError` frames |
| W3-A4-010 | `ExportClientRooms` copy under `mu` |

## Чисто

- NF-02 CanSubscribe до Subscribe; origin allowlist; ME-13 deactivated; SEC-05 send/done; shutdown WS-01; bridge subject→room routing.

## Сводка агента

| Метрика | Значение |
|---------|----------|
| Файлов | 15/15/0 |
| Открытых | **2** |
| NEW | 0 |
| REGRESS | 0 |
| RESIDUAL | 2 (P2×1, P3×1) |
| FIXED-OK | 9 |
