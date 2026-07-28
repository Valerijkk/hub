# Phase 21 — Architecture Completion (ARCH-01)

> **Branch:** `feat/v2.5-trust` · **Depends on:** Phase 20 Milestone Gate  
> **Ref:** `.planning/audits/code-review-final/ARCHITECTURE.md` §6

## Exit criteria

Repository pattern on remaining 7 packages (incremental) · cron boilerplate fully unified

---

## Wave 1 — backup Repository

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W1-01 `backup.Repository` + `EntRepository` | be | arch-ver | ✓ |
| W1-02 `BackupService` → repo (no `*ent.Client`) | be | be-ver | ✓ |
| W1-03 Tests + main.go wiring | qa | be-ver | ✓ |

## Wave 2 — cron cleanup

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W2-01 Migrate pending-user purge to `cron.Schedule` | be | devops-engineer | ✓ |
| W2-02 `platform/cron` unit tests | qa | be-ver | ✓ |

## Wave 3 — docs

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W3-01 COVERAGE ARCH-01 progress + STATE/ROADMAP | sa | arch-ver | ✓ |

**Next waves (backlog):** dashboard, imex, storage, apigateway, rbac, auth handler DB leaks

---
