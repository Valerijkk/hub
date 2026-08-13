# Phase 148 — Mail Merge-readiness Gate

> Task: TASK-241 · Status: GATE_CONTRACT_PREPARED · Depends on: Phases 143–147 and founder signals
> Gate authority: `GATE-MATRIX.md` · Review prompts: `REVIEW-PACKETS.md`

## Ordered tasks

1. Review implemented diff against ADR, ERD, OpenAPI, events and every claimed FR/TC.
2. Resolve the existing `/admin/mail/broadcast` contradiction: disable/remove it under the
   standing no-marketing ADR, or implement the founder-approved replacement controls. Do
   not let mailbox code call it implicitly.
3. Run independent backend/frontend/architecture/security/product reviews within model
   routing rules; fix findings atomically.
4. Through the accepted Docker launcher run applicable feature units, generation/build,
   integration/security/recovery and bounded runtime evidence on current SHA.
5. Prove idempotent account provisioning, generic create/bulk receipts, rule-generation
   reconciliation, explicit spam/ham feedback, safe-content/image proxy and bounded search
   hydration; atomic hydrated page/event/cursor reconciliation, command-lease begin/end fencing,
   expiry/late-owner recovery, anchored no-clear rebuild/catch-up, received-event cutoff/
   marker, replica lease/shutdown, retry/
   idempotency, engine outage, visible freshness, persistence, resource limit and cleanup
   behavior; keep unavailable gates explicit.
6. Synchronize system, architecture, security, test, runbook and state documentation.
7. Finalize QA handoff without editing the protected Python test program.
8. After founder confirms Claude test-rebase is merged, rebase onto current master, resolve
   only mail-owned conflicts, repeat applicable verification and prepare merge report.

## Verdict discipline

- Each gate is `PASS`, `FAIL`, `BLOCKED`, `NOT_RUN`, `N/A`, or `STALE`; report-level
  `PARTIAL` never fills a missing gate.
- Evidence must identify exact current SHA, launcher/image/config provenance, limits,
  results and exact cleanup. Ancestor evidence becomes `STALE` after any code or rebase.
- Independent Terra reviews only report findings. Every product/test-code fix returns to
  an atomic exact-Terra packet and is re-reviewed by the affected lens.
- `PASS_FOR_LOCAL_HANDOFF` does not authorize push, PR, merge, deployment, DNS, license or
  production cutover and does not call the unfinished module/MVP complete or stable.

Push, PR, merge, deployment, DNS and license actions require explicit founder authority.
The gate may end `PARTIAL/BLOCKED`; it must never convert missing runtime evidence to PASS.
