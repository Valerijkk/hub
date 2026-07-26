# Phase 126 Verification — FR-INTR-08 NewsPost trash restore

**Date:** 2026-07-27  
**Auditor:** Grok 4.5 · PW2 planning-audit  
**Verdict:** FAIL

> Phase is **PLAN-only** (D3 / B1 from planning audit). No implementation. Goal not met. Not SKIPPED: gap is verifiable and still open.

## Goal

Close **FR-INTR-08**: restore soft-deleted news post from admin trash **with poll + vote counters** when present. Register `newspost` in trash list/restore/purge (NFR-11, 30 days).

## Status of PLAN tasks

All tasks remain unchecked (`- [ ]`) in `PLAN.md`. Status line: **PLANNED (docs only — no production code)**.

## Static verification (2026-07-27)

| Fact | Evidence |
|------|----------|
| Spec requires trash restore | `docs/system_specification.md` FR-INTR-08; audit `00-PLANNING-AUDIT.md` §4.3 |
| Schema has `deleted_at` | `backend/ent/schema/newspost.go` (per PLAN / audit) |
| Trash types | `backend/internal/audit/service.go` — `ListTrash` / `RestoreItem` / `SoftDelete` switch arms: `issue`, `test_case`, `file`, `project` only |
| `newspost` in trash | **absent** (`grep` / switch inspection — no `case "newspost"`) |

## Acceptance criteria

| # | Criterion | Status |
|---|-----------|--------|
| 1 | Soft-deleted NewsPost in `GET /admin/trash?type=newspost` | FAIL — type not registered |
| 2 | Restore clears `deleted_at` | FAIL — no restore path |
| 3 | Poll + vote counters survive restore | FAIL — not implemented |
| 4 | Purge after 30d | FAIL — not implemented |
| 5 | Admin-only restore; public lists hide soft-deleted | FAIL — restore N/A |
| 6 | Existing trash types unchanged | N/A (no change shipped) |

## Why FAIL (not SKIPPED)

Implementation absence and missing trash type are **positively confirmed**. SKIPPED would understate a known open FR.

## Remaining

Implement per `PLAN.md` tasks under `feat/intranet/newspost-trash-restore`; then re-verify with `go test -tags hubtest` on audit/intranet packages + docs sync.
