# PW4-D3 — PLAN.md for unmet FR (FR-INTR-08, FR-CAL-05)

> Planning-audit follow-up · **docs only** · 2026-07-27  
> Source: `.planning/planning-audit-2026-07/00-PLANNING-AUDIT.md` §4.3 · fixes B1/B2 in `01-PLANNING-FIXES.md`.

Two confirmed unmet FR now have executable phase plans (no production code in this note).

| FR | Spec | Gap (audit) | Phase PLAN |
|----|------|-------------|------------|
| **FR-INTR-08** | `docs/system_specification.md` — restore news post from trash with poll + vote counters | `newspost` has `deleted_at` but missing from trash types in `backend/internal/audit/service.go` | [`.planning/phases/126-intranet-newspost-trash/PLAN.md`](../phases/126-intranet-newspost-trash/PLAN.md) |
| **FR-CAL-05** | `docs/system_specification.md` — series delete → trash 30d | `backend/ent/schema/scheduledcall.go` has `rrule`, **no** `deleted_at`; delete only `is_active=false` | [`.planning/phases/127-calendar-series-trash/PLAN.md`](../phases/127-calendar-series-trash/PLAN.md) |

## Next (implementation, separate work)

1. Implement Phase 126 → register `newspost` in ListTrash / RestoreItem / SoftDelete / PurgeExpired; poll-aware restore AC.
2. Implement Phase 127 → add `deleted_at` to `ScheduledCall`, soft-delete series, register trash type, exclude from calendar/due queries.
3. After code: docs sync (`docs/wiki/API.md`, `.planning/STATE.md`) + domain `testframework` tests; then mark PLAN checkboxes.

**Not in this PW:** Wiki FR-WIKI-01..04, NFR-24 VersionMixin rollout, broad NFR-11 entity expansion (audit B3/B4).
