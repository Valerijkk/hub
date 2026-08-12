# Phase 127 — VERIFICATION (FR-CAL-05 ScheduledCall series trash)

**Measured:** 2026-07-28  
**Verdict:** **PASS**

## Goal check

| Criterion | Result | Evidence |
|-----------|--------|----------|
| `deleted_at` on ScheduledCall + index | PASS | `backend/ent/schema/scheduledcall.go`; `go generate ./ent` OK |
| Delete soft-sets `deleted_at` (+ `is_active=false`) | PASS | `calls/repository.go` `DeleteScheduledCall` |
| All occurrences via RRULE parent soft-delete | PASS | no materialized children; calendar/calls filters |
| Type `scheduled_call` in List/Restore/SoftDelete/Purge | PASS | `audit/service.go` + `repository.go` |
| Retention 30d | PASS | `TestRepository_ScheduledCall_RetentionHidesOlderThan30Days` |
| Restore returns series; RRULE count unchanged | PASS | `TestEntRepository_ListScheduledCalls_ExcludesSoftDeleted` (5 events) |
| Lists/due exclude soft-deleted | PASS | calls + calendar `DeletedAtIsNil` |

## Mutation proof (TASK-002)

1. Remove `scheduledcall.DeletedAtIsNil()` from `calendar.EntRepository.ListScheduledCalls`.
2. Run: `go test -tags hubtest -count=1 -run TestEntRepository_ListScheduledCalls_ExcludesSoftDeleted ./testframework/calendar/`
3. Expect FAIL: soft-deleted series (is_active=true, deleted_at set) appears in list (len=2).
4. Restore filter → PASS (len=1).

## Commands

```bash
cd backend && go generate ./ent
go test -tags hubtest -race -count=1 ./testframework/calls/... ./testframework/audit/... ./testframework/calendar/...
```

## Migration note

Hub uses `client.Schema.Create` (auto-migrate). Existing rows get `deleted_at=NULL`. Owner review required before merge (ent schema change per `50-git.mdc`).
