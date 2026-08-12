# FIX-W10 — tx/data + shutdown + restore maintenance

**Ветка:** `cursor-grok-4.5`  
**Дата:** 2026-07-27  
**Агент:** be2  
**Правила:** 1 commit / finding · mutation · **no Ent schema without approval** · no push master

Verify vs `17-verification.md` / `08-data-layer.md` / `07-infra-domains.md` — all listed findings still OPEN in code before fixes.

| ID | Status | SHA | Notes |
|----|--------|-----|-------|
| A8-005 | FIXED | `25028b5`+`293dee7`+`14c7929` | Soft-delete `deleted_at` on Channel + filtered List/Get/IsMember; cascade hard path retired (parent soft-delete enough for TZ) |
| A8-006 | FIXED | `5c74170` / `9f4ce5b` / `c40c621` | Comment `deleted_at` + soft-delete via SetDeletedAt; List/Update/Toggle filter DeletedAtIsNil; reactions kept, hidden with comment |
| A8-007 | FIXED | `f3dbad4` | `MongoDisconnectCleanup` + `addCleanup` in `wiring_messenger`. Mutation: skip Disconnect → Ping still OK → FAIL |
| A8-008 | FIXED | `23d6abb` | `lifecycle.CancelOnListenError` instead of `os.Exit(1)` on Listen err. Mutation: skip cancel → FAIL |
| A7-003 | FIXED | `0dd0dd7` | `MaintenanceGate` + Fiber middleware 503; `setMaintenanceMode` flips gate. Mutation: log-only setMaintenance → FAIL |

## A8-005 detail
`DeleteChannel` soft-deletes via `SetDeletedAt` (A8-005). List/Get/FindDM/FindGlobal filter `DeletedAtIsNil`. `IsMember` / calls `IsChannelMember` require active channel so WS does not expose soft-deleted rooms. Child cascade hard-delete retired — orphans kept for future admin hard-purge/restore.

**Status FIXED:** TZ soft-delete closed after Ent schema approval.

## A8-006 detail
`DeleteComment` soft-deletes via `SetDeletedAt` (A8-006). `ListComments` / `UpdateComment` / `ToggleReaction` filter `DeletedAtIsNil`. Reactions are not hard-deleted — they remain on the row but disappear from lists with the soft-deleted comment. Schema: `ent/schema/comment.go` `deleted_at` + index.

**Status FIXED:** TZ soft-delete / 30-day trash closed after Ent schema approval.

## A8-007 / A8-008
Graceful shutdown path: Listen fail → `rootCancel` → `<-rootCtx.Done()` → Shutdown → `defer runCleanups()` (incl. Mongo Disconnect).

## A7-003
`wireFiber` installs `MaintenanceMiddleware` on `/api` before routes; `wireIntegrations` binds same gate into `BackupService`. Health `/api/health` bypasses; other API → 503 during restore.
