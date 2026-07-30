# MUTATION W13 slot 4 refill — TASK-064 (qa-auto1) · 2026-07-30

> Checker: **qa-auto1** · Env: localhost `POSTGRES_TEST_URL` / `MONGODB_TEST_URL` · Go 1.25  
> Criteria: `.planning/backlog/archive/legacy/TASK-064-be-p3-ws-calls-lifecycle.md` `[MUTATION CRITERIA]`  
> Class: **P3-BE-10** · Коммит: **не создавался** (инструкция слота).

## Verdict

**5/5 PROVEN** · **1 N/A** (W3-A4-006 deferred) · **0 NOT_PROVEN** · TASK-064 → **Закрыта**

| ID | Verdict | RED test | Mutation (BREAK) | Layer |
|----|---------|----------|------------------|-------|
| W3-A4-006 | **N/A** | deferred Residual («после product») | — | handler WS upgrade |
| W3-A4-007 | **PROVEN** | `TestHub_EmptyRoomDeletedFromMap_TASK064`; `TestHub_EmptyRoomDeletedOnUnregister_TASK064` | drop `h.rooms.Delete(room)` on empty | WS Hub |
| W3-A4-008 | **PROVEN** | `TestBridge_StartFail_CallsStop_TASK064` | no `b.Stop()` on Subscribe err | WS Bridge |
| W3-A4-009 | **PROVEN** | `TestReadPump_DeniedSubscribe_SendsErrorFrame_TASK064`; `TestReadPump_AuthorizerError_SendsErrorFrame_TASK064` | no `enqueueServerMessage` on deny/authz err | WS readPump |
| W3-A4-010 | **PROVEN** | `TestExportClientRooms_CopyUnderMu_TASK064` | `return c.rooms` without copy/mu | WS Client export |
| W3-A5-013 | **PROVEN** | `TestHandler_ListProjectRecordings_NilRBAC_FailClosed_TASK064`; `…_GetRecordingDownloadURL_NilRBAC_FailClosed_TASK064` | remove both `rbacSvc == nil` checks | calls handler |

Production restored (`git checkout --` touched guards) после каждого BREAK. Leftover MUTATED: **нет**.  
`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)`.

## Per-ID evidence

### W3-A4-006 — query `?token=` (deferred)

```text
N/A — Residual TASK-064 / реестр «после product» (JWT ticket exchange).
Adjacent outside slot: TestWSHandler_QueryToken_Denied
(backend/testframework/ws/handler_auth_test.go) — not mutated this slot.
```

### W3-A4-007 — empty room `rooms.Delete`

```text
BASELINE EmptyRoomDeleted → PASS (~8.2s)

BREAK: backend/internal/ws/hub.go removeRoomMember
  h.rooms.Delete(room) → _ = room // MUTATED W3-A4-007
→ FAIL TestHub_EmptyRoomDeletedFromMap_TASK064
        empty room key must be deleted from hub.rooms
→ FAIL TestHub_EmptyRoomDeletedOnUnregister_TASK064
        unregister of last member must delete empty room
RESTORE → PASS (~6.8s)
```

### W3-A4-008 — Bridge Start → Stop rollback

```text
BASELINE Bridge_StartFail_CallsStop_TASK064 → PASS (~6.6s)

BREAK: backend/internal/ws/bridge.go Start loop
  b.Stop(); return err → return err // MUTATED W3-A4-008
→ FAIL TestBridge_StartFail_CallsStop_TASK064
        ExportBridgeSubsLen expected 0 (got prior subs)
RESTORE → PASS (~6.9s)
```

### W3-A4-009 — deny/authz → error frame

```text
BASELINE ReadPump_.*_SendsErrorFrame_TASK064 → PASS (~6.8s)

BREAK: backend/internal/ws/handler.go readPump
  enqueueServerMessage(TypeError, …) removed on authErr / !allowed
→ FAIL TestReadPump_DeniedSubscribe_SendsErrorFrame_TASK064
        expected subscribe_denied error frame (timeout)
→ FAIL TestReadPump_AuthorizerError_SendsErrorFrame_TASK064
        expected subscribe_error error frame (timeout)
RESTORE → PASS (~7.1s)
```

### W3-A4-010 — ExportClientRooms copy under mu

```text
BASELINE ExportClientRooms_CopyUnderMu_TASK064 → PASS (~6.6s)

BREAK: backend/internal/ws/export_hooks.go
  copy under mu → return c.rooms // MUTATED W3-A4-010
→ FAIL TestExportClientRooms_CopyUnderMu_TASK064
        ExportClientRooms must return a copy (snap mutates live map)
RESTORE → PASS (~6.6s)
```

### W3-A5-013 — nil RBAC fail-closed (list + download)

```text
BASELINE NilRBAC_FailClosed_TASK064 → PASS (~12.4s)

BREAK: backend/internal/calls/handler_recording.go
  remove both `if h.rbacSvc == nil { return 500 }` blocks // MUTATED W3-A5-013
→ FAIL (package) — panic: nil pointer dereference on GetEffectivePermissions
  (kill recipe said 500→200; actual RED = panic escaping Fiber Test —
   named tests still fail → PROVEN; see Gaps)
RESTORE → PASS (~8.7s)
```

```text
FINAL ws/ -run TASK064 → PASS (~7.6s)
FINAL calls/ -run NilRBAC_FailClosed_TASK064 → PASS (~7.3s)
leftover_mutated=False
```

## Commands

```text
cd backend
$env:POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
$env:MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin"

go test -tags hubtest -race -count=1 -timeout=180s ./testframework/ws/ -run EmptyRoomDeleted
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/ws/ -run Bridge_StartFail_CallsStop_TASK064
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/ws/ -run "ReadPump_.*_SendsErrorFrame_TASK064"
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/ws/ -run ExportClientRooms_CopyUnderMu_TASK064
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/calls/ -run NilRBAC_FailClosed_TASK064
```

Runner: `.planning/audits/total-2026-07/_run_mutations_w13_task064.py`  
JSON: `_mut-batch-w13-task064.json`.

## Gaps

- **W3-A4-006** — product-deferred JWT `?token=` → short-lived ticket; **N/A**, не NOT_PROVEN.
- **W3-A5-013** — criteria kill recipe «500→200»; после удаления nil-check Fiber `app.Test` падает с **nil panic**, не 200. Named tests всё равно RED → **PROVEN**; softer 200-path без panic — **GAP_NOTED** (не требует reopen).

## Handoff

- Evidence: этот файл · TASK-064 **Закрыта** · README sync  
- be-ver → code-reviewer (fix/tests already FIXED; mutation PROVEN + N/A)  
- Коммит: не создавался
