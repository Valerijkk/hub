# Phase 21 — Architecture Completion (ARCH-01 W1–W3) — VERIFICATION

**Date:** 2026-07-27  
**Method:** Repository interface inspection + cron wiring + ROADMAP checkbox  
**Sources:** `PLAN.md` (W1–W3 ✓; further packages deferred to later phases), live `backup/` + `platform/cron`

## Goal (exit criteria for this phase)

Repository pattern on **backup** · unify pending-user purge onto `cron.Schedule` · docs progress note. (Remaining packages explicitly backlog → phases 22+.)

## Measured evidence

| Criterion | Evidence (2026-07-27) | Result |
|-----------|----------------------|--------|
| backup.Repository + EntRepository | `backend/internal/backup/repository.go`: `Repository` interface + `EntRepository` with `*ent.Client` confined to repo | PASS |
| BackupService without `*ent.Client` | `BackupService` struct field is `repo Repository` only (`service.go`) | PASS |
| cron.Schedule pending purge | `wiring_auth.go`: `cron.Schedule(…, "pending-user-purge", …)` (ORG-06) | PASS |
| platform/cron + tests | `backend/internal/platform/cron/cron.go` (`Schedule`); `backend/testframework/platform/cron/cron_test.go` | PASS |
| ROADMAP | Phase 21 [x] 2026-06-22; later note ARCH-01 17/17 at Phase 27 | PASS |

## Notes

- Plan listed separate `ent_repository.go`; implementation colocates `EntRepository` in `repository.go` — same architectural outcome.
- W3 “COVERAGE ARCH-01 progress” is historical; subsequent phases completed remaining packages — out of scope for FAIL.

## Verdict: **PASS**
