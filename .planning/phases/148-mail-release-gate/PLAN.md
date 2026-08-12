# Phase 148 — Mail Merge-readiness Gate

> Task: TASK-241 · Status: PLANNED · Depends on: Phases 143–147 and founder signals

## Ordered tasks

1. Review implemented diff against ADR, ERD, OpenAPI, events and every claimed FR/TC.
2. Resolve the existing `/admin/mail/broadcast` contradiction: disable/remove it under the
   standing no-marketing ADR, or implement the founder-approved replacement controls. Do
   not let mailbox code call it implicitly.
3. Run independent backend/frontend/architecture/security/product reviews within model
   routing rules; fix findings atomically.
4. Through the accepted Docker launcher run applicable feature units, generation/build,
   integration/security/recovery and bounded runtime evidence on current SHA.
5. Prove projection reconciliation, retry/idempotency, engine outage, persistence, resource
   limit and cleanup behavior; keep unavailable gates explicit.
6. Synchronize system, architecture, security, test, runbook and state documentation.
7. Finalize QA handoff without editing the protected Python test program.
8. After founder confirms Claude test-rebase is merged, rebase onto current master, resolve
   only mail-owned conflicts, repeat applicable verification and prepare merge report.

Push, PR, merge, deployment, DNS and license actions require explicit founder authority.
The gate may end `PARTIAL/BLOCKED`; it must never convert missing runtime evidence to PASS.
