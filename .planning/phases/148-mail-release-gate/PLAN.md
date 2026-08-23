# Phase 148 — Mail Merge-readiness Gate
> **Current narrow deletion-source authority (2026-08-23; highest precedence):** accepted plan `3f7aa2c58407ac3a10ad427cfa86256e0a9d340a` is `DOCS_PLAN_STATIC_ACCEPTED / SOURCE_AUTHORITY_ACCEPTED` only for one future atomic no-test exact2 production packet: `backend/internal/mail/mailbox_deletion.go` and `backend/internal/mail/mailbox_deletion_fake.go`. Plan Contract/high, Security/xhigh and Lifecycle/high reviews are C/I/M `0/0/0`; `provenance_source: dispatcher-enforced`, no fallback, telemetry `UNKNOWN`. Source remains `UNIMPLEMENTED / UNCOMMITTED / UNACCEPTED` until its own commit and immutable C/S/L triple `0/0/0`. Adapter capability proof is not required only for this provider-neutral ABI plus two standalone fakes exception, but remains mandatory for every concrete provider adapter. Persistence/Ent/schema/generation/migration/SQL/worker/storage drain/account finalization/User-GDPR/HTTP/NATS/Docker/runtime stay `BLOCKED` / `NOT_RUN`. Old blanket deletion-source-blocked or plan-pending claims are `CONSUMED/HISTORICAL` only within this exact2 scope; unrelated gates are unchanged.
> **Engine deletion-port adjudication (2026-08-23):** `STATIC_ACCEPTED / DOCS_DESIGN_ONLY`; former deletion-port-next selection is `CONSUMED/HISTORICAL`, never source authority. Freeze separate opaque future ports: `EngineAccountDeletionPort` (`DeleteAccount`, `ReconcileAccountAbsence`) and `MailObjectDeletionPort` (`DeleteObject`, `ReconcileObjectAbsence`); do not widen normal `MailEngine` or combine account delete with drain. Requests durably retain nonzero `DeletionID` plus account provisioning key+engine account ID or object bucket+key; opaque IDs use `MaxEngineOpaqueIDBytes` and bucket/key existing storage bounds. Results are `deleted|authoritatively_absent`, observations `present|authoritatively_absent`; bare `NotFound` preserves intent unless scoped authoritative absence is proven. No provider/raw IDs or causes are public/logged. Private intent/control is not owner projection; short lease transactions keep I/O outside locks, use exact token+generation+scope CAS, reclaim expired lease only, and stale token/zero rows cause no I/O/compensation. Object drain starts after `upload_drain_until` and attachment/reference terminality; account removal is last after exact child/object/receipt/control/activity manifest; hard-delete only after closed evidence. No Stalwart route/OAuth/DNS/OpenAPI assumption. Adapter capability and independent contract/security/lifecycle reviews are mandatory; future source is bounded ABI/fake then separately planned persistence/worker. Ent/source/generated/migration/SQL/tests/Docker/runtime and physical User/GDPR remain `BLOCKED` / `NOT_RUN`.

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
   hydration; atomic projection-only page/event/cursor reconciliation, command-lease begin/end fencing,
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
