# FIX-W10 — tx/data + shutdown + restore maintenance

**Ветка:** `cursor-grok-4.5`  
**Дата:** 2026-07-27  
**Агент:** be2  
**Правила:** 1 commit / finding · mutation · **no Ent schema without approval** · no push master

Verify vs `17-verification.md` / `08-data-layer.md` / `07-infra-domains.md` — all listed findings still OPEN in code before fixes.

| ID | Status | SHA | Notes |
|----|--------|-----|-------|
| A8-005 | FIXED | `b77ba23` | Tx cascade children before Channel hard-delete (FK). Soft-delete `deleted_at` **not** added (needs Ent approval). Mutation: skip member delete → FK FAIL |
| A8-006 | BLOCKED | — | Comment hard-delete confirmed; TZ soft-delete requires `deleted_at` on `ent/schema/comment.go` — **no schema without approval** |
| A8-007 | FIXED | `f3dbad4` | `MongoDisconnectCleanup` + `addCleanup` in `wiring_messenger`. Mutation: skip Disconnect → Ping still OK → FAIL |
| A8-008 | FIXED | `23d6abb` | `lifecycle.CancelOnListenError` instead of `os.Exit(1)` on Listen err. Mutation: skip cancel → FAIL |
| A7-003 | FIXED | `0dd0dd7` | `MaintenanceGate` + Fiber middleware 503; `setMaintenanceMode` flips gate. Mutation: log-only setMaintenance → FAIL |

## A8-005 detail
`DeleteChannel` now deletes polls/votes/options, pins, invites, issue links, bookmarks, scheduled calls; clears optional Call.channel_id; deletes members; then channel — one Ent tx.

## A8-006 detail
Still `Comment.Delete()` hard path in `tracker/repo_comments.go`. Remediations in audit require schema `deleted_at` — deferred pending approval.

## A8-007 / A8-008
Graceful shutdown path: Listen fail → `rootCancel` → `<-rootCtx.Done()` → Shutdown → `defer runCleanups()` (incl. Mongo Disconnect).

## A7-003
`wireFiber` installs `MaintenanceMiddleware` on `/api` before routes; `wireIntegrations` binds same gate into `BackupService`. Health `/api/health` bypasses; other API → 503 during restore.
