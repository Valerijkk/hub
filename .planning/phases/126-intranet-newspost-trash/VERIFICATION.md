# Phase 126 — VERIFICATION (FR-INTR-08 NewsPost trash)

**Measured:** 2026-07-28  
**Verdict:** **PASS** (partial on poll fixtures — see notes)

## Goal check

| Criterion | Result | Evidence |
|-----------|--------|----------|
| `newspost` in ListTrash / RestoreItem / SoftDelete / PurgeExpired | PASS | `backend/internal/audit/service.go` switch arms |
| Soft-deleted post in `GET /admin/trash?type=newspost` within 30d | PASS | `TestAuditService_RestoreNewsPost_ViaTrashAPI`, `TestRepository_SoftDeleteAndRestoreNewsPost` |
| Restore clears `deleted_at`; post readable again | PASS | same + `RestoreNewsPost` → `ClearDeletedAt` |
| Retention: older than 30d not listed | PASS | `TestRepository_NewsPost_RetentionHidesOlderThan30Days` |
| Purge hard-deletes past cutoff | PASS | `TestRepository_PurgeExpiredNewsPosts` |
| Poll + vote counters restored with post (FR-INTR-08 literal) | **READY / deferred** | `news_poll*` schemas absent in `backend/ent/schema/`; soft-delete of post alone hides future FK-linked poll; restore is `ClearDeletedAt` only — no separate poll wipe. Phase PLAN out-of-scope until Wave 2 poll CRUD. |
| Regression: issue / test_case / file / project | PASS | `go test -tags hubtest ./testframework/audit/...` |

## Mutation proof (TASK-001)

1. Remove `case "newspost"` from `AuditService.RestoreItem` in `backend/internal/audit/service.go`.
2. Run: `go test -tags hubtest -count=1 -run TestAuditService_RestoreNewsPost_ViaTrashAPI ./testframework/audit/`
3. Expect FAIL: `неизвестный тип сущности: newspost`.
4. Restore the case → same test PASS.

## Commands

```bash
cd backend
go test -tags hubtest -race -count=1 ./testframework/audit/... ./testframework/intranet/...
```

## Notes

- Intranet delete path already soft-sets `deleted_at` (`SoftDeleteNewsPost`); trash registration closes the restore gap.
- Admin UI tab «Новости» added in `frontend/src/components/admin/trash-manager.tsx`.
