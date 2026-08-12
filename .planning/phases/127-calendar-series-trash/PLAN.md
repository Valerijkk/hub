# Phase 127 — FR-CAL-05 ScheduledCall series soft-delete + trash

**Status:** IMPLEMENTED 2026-07-28 (TASK-002) — see `VERIFICATION.md`  
**Source:** `docs/system_specification.md` FR-CAL-05 (US-CAL-02), NFR-11 (and resource-conflict clause of FR-CAL-05 / NFR-33 where already in scope)  
**Audit:** `.planning/audits/planning-audit-2026-07/00-PLANNING-AUDIT.md` §4.3 · `01-PLANNING-FIXES.md` B2  
**Index note:** `.planning/audits/planning-audit-2026-07/PW4-D3-FR-PLANS.md`  
**Suggested branch (when implementing):** `feat/calendar/scheduledcall-series-trash`

## Goal

Close confirmed gap **FR-CAL-05** for the **existing** recurring series entity `ScheduledCall`: deletion must soft-delete the series (all RRULE occurrences disappear from active views) into the **30-day trash**, with admin restore via the unified Trash API. Today delete only flips `is_active=false` and the schema has **no** `deleted_at` — neither soft-delete nor trash.

## Gap (confirmed)

| Fact | Path |
|------|------|
| Spec: series delete → trash 30d | `docs/system_specification.md` FR-CAL-05 · `DELETE /api/calendar/events/:eventId` (Wave 2 calendar_event wording) |
| Series real via `rrule` | `backend/ent/schema/scheduledcall.go` |
| **No** `deleted_at` on schema | `backend/ent/schema/scheduledcall.go` (`grep deleted_at` = 0) |
| Delete = deactivate only | `backend/internal/calls/repository.go` (`DeleteScheduledCall` → `SetIsActive(false)`) · `backend/internal/calls/scheduled.go` · `backend/internal/calls/handler_scheduled.go` |
| Trash types lack scheduled call | `backend/internal/audit/service.go` (`ListTrash` / `RestoreItem` / `SoftDelete` / `PurgeExpired`) |
| Calendar aggregation of calls | `backend/internal/calendar/` (RRULE expansion on read) |

> **Scope clarification:** Audit maps the unmet FR to **`ScheduledCall`** (shipped series). Wave 2 `calendar_event` + `event_recurrence` trash is a separate future slice if those schemas are not yet in `backend/ent/schema/` — do not block this phase on full Wave 2 calendar CRUD.

## Tasks

- [x] Add `deleted_at` (optional/nillable timestamptz UTC) + index to `backend/ent/schema/scheduledcall.go`; generate Ent + Atlas migration.
- [x] Change delete path: soft-set `deleted_at` (and keep inactive / exclude from due-job queries). Prefer clearing or ignoring `is_active` consistently so soft-deleted series never fire reminders or LiveKit scheduling.
- [x] Register trash type `scheduled_call` (or `scheduledcall` — pick one string and document in API) in `backend/internal/audit/service.go` + `repository.go`: list / restore (clear `deleted_at`, re-activate if needed) / soft-delete / purge after `TrashRetentionDays` (30).
- [x] Ensure “all occurrences” semantics: occurrences are **not** materialized; soft-deleting the parent series removes them from `GET` calendar/calls listings and RRULE expansion filters (`deleted_at IS NULL`).
- [x] Calendar/calls list and `GetDueScheduledCalls` must exclude soft-deleted rows.
- [x] Tests in `backend/testframework/calls/` and `backend/testframework/audit/` (+ calendar aggregation if touched): create series with `rrule` → delete → absent from active lists → present in trash → restore → series and expanded occurrences visible again; purge past cutoff; restore after purge → 404.
- [x] Docs sync on implement: `docs/wiki/API.md` (trash type + scheduled-call delete semantics); `.planning/STATE.md` session note.

## Acceptance criteria

1. `ScheduledCall` schema has `deleted_at`; migration applies cleanly.
2. `DELETE` on a scheduled series soft-deletes the row; no active list/calendar expansion/due-job includes it (FR-CAL-05 “all occurrences”).
3. Soft-deleted series appears in `GET /api/admin/trash` for the new type with ≤30 days remaining (NFR-11).
4. `POST /api/admin/trash/<type>/:id/restore` clears `deleted_at`; series and RRULE expansion return to active views.
5. After retention, purge hard-deletes expired soft-deleted series.
6. Existing one-off scheduled calls (empty `rrule`) follow the same soft-delete/trash path (consistent delete semantics).
7. Regression: issue/test_case/file/project trash unchanged; LiveKit call session APIs unrelated to scheduled metadata unchanged.

## Out of scope

- Full Wave 2 `calendar_event` / `resource` / `resource_booking` / iCal subscription modules (FR-CAL-01..04, 06..08) unless already present — this phase closes the **scheduledcall** gap called out in the audit.
- Per-occurrence delete/edit (spec: single-occurrence edit is out of Wave 2 scope).
- Resource booking conflict `409` + conflict dates list (FR-CAL-05 resource clause / NFR-33) — only if bookings already exist on `ScheduledCall`; otherwise defer with calendar_event booking phase.
- Frontend redesign of trash or calendar UI beyond wiring restore if an admin trash type picker needs a new label.
- Production code in this planning commit — **PLAN only**.

## Depends on

- **126** intranet-newspost-trash — shared admin trash registration pattern (`ListTrash` / `RestoreItem` / `PurgeExpired`).
- Soft: `ScheduledCall.rrule` already in schema (no Wave 2 calendar_event required).

## Dependencies / notes

- Pattern: `backend/internal/audit/service.go` issue/project soft-delete + restore + purge.
- Keep UTC timestamps (NFR-12 / TZ constraints).
- Coordinate type string with FE/docs (`scheduled_call` preferred for readability in `/admin/trash/:type/...`).
- **Ent schema change** requires owner review before merge (`50-git.mdc`).

## Verify (when implementing)

```bash
cd backend && go test -tags hubtest -race -count=1 ./testframework/calls/... ./testframework/audit/... ./testframework/calendar/...
```
