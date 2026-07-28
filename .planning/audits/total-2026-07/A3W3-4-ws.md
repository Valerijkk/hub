# A3 W3 агент 4 — ws (FINAL re-audit)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · ветка `cursor-grok-4.5`  
> Mirror: [A2W3-4-ws.md](A2W3-4-ws.md) · Baseline Stage A: [A3-4-ws.md](A3-4-ws.md)  
> Периметр: 15 файлов (6 prod + 9 tests) · Прочитано: **15/15** · Пропущено: **0**  
> Plan: [A3-PLAN.md](A3-PLAN.md) · Assignment: [W3-ASSIGNMENTS.md](W3-ASSIGNMENTS.md) § A4

> **Вне списка assignment (не в таблице чтения):** `backend/internal/ws/ticket_protocol.go` +
> `ticket_protocol_task103_test.go` — используются handler'ом (ADR-WS-JWT-HANDSHAKE);
> беглый просмотр без отдельных находок. Сильные deny/revoke regression —
> `authz_deny_room_task037_test.go`, `lifecycle_task064_test.go`,
> `hub_register_unregister_race_test.go`, `handler_nil_authz_test.go`,
> `revoke_*_task109/042` (вне слота; evidence FIXED-OK).

## Сверка с A2 (A2W3-4-*)

| A2 ID | Tag | Статус в HEAD | Комментарий |
|-------|-----|---------------|-------------|
| A2W3-4-001 | FIXED-OK | закрыт | TASK-103 / ADR-WS-JWT-HANDSHAKE: ticket subprotocol + deny `?token=` → 401 (`handler.go:141–160`); `TestWSHandler_QueryToken_Denied` |
| A2W3-4-002 | FIXED-OK | закрыт (ещё в A2) | TASK-109: auth publish logout/revoked → bridge `UnregisterBySession` |

**REGRESS:** 0

### Cascade FIXED-OK (Stage A W3-A4-* / A2 FIXED-OK)

| Stage A / A2 | Evidence (файл:строка) |
|--------------|------------------------|
| W3-A4-001 | `registerClient` skip if `done` — `hub.go:266–274`; `hub_register_unregister_race_test.go` (вне слота) |
| W3-A4-002 | `subscribeClient` checks `done` — `hub.go:287–295` |
| W3-A4-003 | `authz == nil` → 503 — `handler.go:152–155`; `handler_nil_authz_test.go` |
| W3-A4-004 | `UnregisterBySession` + bridge revoked/logout — `hub.go:205–228`, `bridge.go:242–252` |
| W3-A4-005 | deny/error → не в room + no broadcast — `authz_deny_room_task037_test.go` (вне слота); prod `handler.go:249–269` |
| W3-A4-006 / A2W3-4-001 | ticket handshake; query JWT forbidden — см. выше |
| W3-A4-007 | empty room → `h.rooms.Delete` — `hub.go:315–332` |
| W3-A4-008 | `Start` fail → `Stop()` — `bridge.go:51–56` |
| W3-A4-009 | `EventSubscribeDenied` / `EventSubscribeError` — `protocol.go:13–21`, `handler.go:252–267` |
| W3-A4-010 | `ExportClientRooms` copy under `mu` — `export_hooks.go:88–96` |

## Прочитанные файлы

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `backend/internal/ws/authorizer.go` | 91 | 0 | чисто — NF-02 fail-closed |
| `backend/internal/ws/bridge.go` | 278 | 0 | FIXED-OK Start rollback + revoke Unregister |
| `backend/internal/ws/export_hooks.go` | 118 | 0 | FIXED-OK ExportClientRooms under mu |
| `backend/internal/ws/handler.go` | 336 | 0 | FIXED-OK ticket + deny `?token=` + nil authz + deny frames |
| `backend/internal/ws/hub.go` | 440 | 0 | FIXED-OK register race, late subscribe, empty rooms |
| `backend/internal/ws/protocol.go` | 37 | 0 | чисто (+ EventSubscribeDenied/Error) |
| `backend/testframework/ws/authorizer_test.go` | 203 | 0 | чисто |
| `backend/testframework/ws/bridge_lifecycle_test.go` | 63 | 0 | чисто |
| `backend/testframework/ws/bridge_test.go` | 205 | 0 | чисто |
| `backend/testframework/ws/handler_auth_test.go` | 167 | 1 | NEW — mutation gap query+ticket |
| `backend/testframework/ws/hub_shutdown_test.go` | 160 | 0 | чисто |
| `backend/testframework/ws/hub_test.go` | 359 | 0 | чисто |
| `backend/testframework/ws/origin_test.go` | 58 | 0 | чисто |
| `backend/testframework/ws/protocol_test.go` | 160 | 0 | чисто |
| `backend/testframework/ws/pump_test.go` | 287 | 0 | слабо deny/error (сильнее вне слота TASK-037) — FIXED-OK |

Файлов в списке: **15**. Прочитано: **15**. Пропущено: **0**.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|----------|-------------|----------|---------------|------------|
| A3W3-4-001 | NEW | P3 | `handler_auth_test.go:88–102` (vs prod `handler.go:157–160`) | `TestWSHandler_QueryToken_Denied` шлёт только `?token=` **без** ticket subprotocol | Комментарий prod: отказ даже при параллельном subprotocol. Мутация «убрать deny `?token=`, оставить ticket-required» оставляет тест зелёным (401 из-за missing ticket), а JWT снова может попадать в proxy/access URI при `?token=`+ticket | Добавить кейс: valid ticket **и** `?token=` → 401; mutation: удалить query-guard → FAIL |

## FIXED-OK (сводка)

| ID | Evidence |
|----|----------|
| A2W3-4-001 | TASK-103: `RedeemWSTicket` + `Sec-WebSocket-Protocol hub.ws.ticket.*`; query → 401 |
| A2W3-4-002 / W3-A4-004 | bridge revoke/logout → `UnregisterBySession` |
| W3-A4-001…003, 005, 007…010 | см. cascade table — без регрессии на HEAD |

## Чисто

- **NF-02:** `CanSubscribe` до `Subscribe`; nil checker / unknown prefix fail-closed; deny/error frames.
- **Origin/CSWSH:** CORS_ORIGINS allowlist; DEV_MODE warn; empty Origin — non-browser (осознанно).
- **ME-13:** deactivated → 403 до upgrade.
- **SEC-05 / HI-07 / WS-01:** send не закрывается; multi-tab ConnectionID; shutdown non-blocking.
- **Bridge routing:** project/channel/call/chat/auth/notification; malformed JSON skip.
- **Тесты слота:** authorizer table; bridge route/lifecycle; hub multi-tab/SEC-05/shutdown; origin; protocol JSON; pump ping/upgrade ticket path.

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 0 |
| P3 | 1 |
| **Total** | **1** |

### Теги vs A2

| Tag | Count |
|-----|------:|
| FIXED-OK (A2 active → closed) | 1 (`A2W3-4-001`) |
| FIXED-OK (Stage A cascade) | 10 |
| RESIDUAL | 0 |
| NEW | 1 (`A3W3-4-001`) |
| REGRESS | 0 |

**Топ IDs:** `A3W3-4-001` (query+ticket mutation gap).
