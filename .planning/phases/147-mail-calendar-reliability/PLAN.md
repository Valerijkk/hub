# Phase 147 — Mail/Calendar Reliability and Integrations
> **147-02A static acceptance (2026-08-23):** source `3f4d174f888f08a119bbe381dc86e5e9b34e0301` is `STATIC_ACCEPTED / STATIC_SOURCE_ONLY` for exactly the sole production file `backend/internal/calendar/itip_ingest_policy.go`, after independent Contract/high and Security/xhigh APPROVED C/I/M `0/0/0`. Tests are intentionally absent under the current production-only constraint; all executable/generation/runtime evidence is `NOT_RUN`. This is pure REQUEST/CANCEL order policy only, without product, merge, readiness, schema, generation, migration, SQL, Docker, or runtime authority. D0-next is `CONSUMED/HISTORICAL`; P3-C, 147-02B/147-02C, D1/D2, and P3-E retain their stated blockers; no next unblocked P3 source is selected.
> **P3-D0 static acceptance (2026-08-23):** immutable source `47faf3a705e5641f4ceb7ee365c3dfeeeace725d` is `STATIC_ACCEPTED / STATIC_SOURCE_ONLY` for exactly `backend/ent/schema/calendarinboundmessagereceipt.go`. It adds required immutable `organization_id`/`mailbox_id` UUID scalars and database nonzero checks, with no Ent edges or raw data; existing uniqueness, state-shape, and due behavior are unchanged. Independent Contract/high and Security/xhigh review is APPROVED C/I/M `0/0/0`. The D0-next route is `CONSUMED/HISTORICAL`. P3-C remains blocked on Mail Ent generation; 147-02B/147-02C and D1/D2 remain blocked on active Calendar generation/dependencies; P3-E is conjunction-blocked. No next unblocked P3 source is selected. Generation, migration, SQL, build, tests, Docker, and runtime are `NOT_RUN`; no readiness, merge, or product claim.

> Task: TASK-240 · Status: CONTRACT_0_1_8_ROOT_FROZEN_GLOBAL_REVIEW_PENDING_DEPENDENCY_BLOCKED · Depends on: Phases 143–146
> Required code runtime: `gpt-5.6-terra` / `xhigh`, no fallback
> Contract: `CONTRACT.md` · Dispatch: `DISPATCH-INDEX.md` + `DISPATCH-PACKETS.md`

> **P3-D schema/migration checkpoint (2026-08-26):**
> SCHEMA_MIGRATION_VERIFIED / RUNTIME_NOT_RUN. Ent desired state and exact
> additive version 20260826000000 are authored and Docker-verified at 3
> executed/0 pending migrations; lint, snapshot and semantic parity passed.
> The broad generated candidate was rejected because it contained unrelated
> Mail churn. P3-D repository/service/worker, 147-02B/02C, D1/D2 and full
> application Docker readiness remain pending.

> **P3-D source checkpoint (2026-08-26):** `SOURCE_AUTHORED /
> INDEPENDENT_REVIEW_ACCEPT_C_I_M_0_0_0 / RUNTIME_NOT_RUN`. The accepted chain
> is receipt port `8a7478dd`, shared DB clock `4befb938`, Ent registration/due
> claim `2455a123`, fenced lock/completion `4c030bd6`, private REPLY bridge
> `8a238fcd`, orchestration service `e1db48ce`, and recovery worker `690f77e5`.
> Every production atom was Terra/xhigh with dispatcher-enforced/no-fallback
> provenance and received an independent Terra/high `ACCEPT C/I/M 0/0/0`.
> Older P3-D repository/service/worker-pending clauses are
> `CONSUMED/HISTORICAL` in this exact scope. Generated Ent ABI, compile,
> formatting, executable tests, Docker and runtime remain `NOT_RUN`; P3-E,
> authoritative HTTP/UI correlation, full-stack Docker and merge remain
> pending. This is source acceptance only, not product readiness.

## Outcome

Close deterministic incoming invitation ordering and outward REPLY recovery, private
availability, durable reminder scheduling and the CRM manual-mail seam. Keep L1 feed code
conditional on founder question 5. L2 external mailbox ingest and L3 OAuth remain outside
this phase.

## Ordered execution

1. Preserve the independently reviewed 0.1.7 reliability/bridge/canonical-hash semantics and
   independently review the full Mail 0.1.8 amendment. Prior review evidence is historical.
2. After accepted Phase 143/144 schemas, add inbound object/message receipts, reply
   delivery and reminder receipts; generate only through the accepted launcher.
3. Implement REQUEST/CANCEL order policy, atomic application, authenticated message-part
   source, consumer and bounded recovery.
4. Implement external-organizer REPLY delivery through the accepted mailbox draft/send/
   reconciliation path.
5. Implement privacy-preserving availability across canonical events, only unlinked
   transition scheduled calls and approved leave; never count a linked pair twice.
6. Add persistent in-app reminder dedupe, scheduler/outbox and notification consumer while
   keeping Telegram/push best-effort and Resend unchanged. Both due claim and consumer
   finalization recheck current offset/start/sequence/eligibility under one lock order;
   a stable user-row guard serializes preference and quiet-hours changes.
7. If founder question 5 is affirmative, separately execute the eight L1 feed packets after
   encryption, migration and SSRF gates; otherwise leave them `BLOCKED/DEFERRED`.
8. Add CRM compose context, create one ordinary idempotent Mail draft, navigate by the
   owner-scoped draft UUID, and offer explicit owned-message link only after definitive
   Mail submission or from an already owned message reader; no recipient in URL/history
   and no campaign, tracking or auto-log behavior.
9. Synchronize routed docs and hand non-unit scenarios to QA without changing protected
   testframework/test-rebase paths.

## Current gates

| Gate | State |
|---|---|
| Root reliability contract static review | 0.1.8 ROOT_FROZEN; semantic 0.1.7 PASS on exact SHA `b3fb6f48` is historical evidence |
| Independent reliability contract review | PENDING for 0.1.8; earlier Terra/high review does not cover engine/API/persistence changes |
| Exact Terra code runtime | AUTHORIZED_BY_FOUNDER / first actual packet pending |
| Phase 143/144 accepted implementation | NOT_IMPLEMENTED |
| Engine authentication mapping | BLOCKED_BY_PHASE_145 |
| Accepted Docker launcher | NOT_READY |
| Ent generation/migration | NOT_RUN / FOUNDER_DECISION_PENDING |
| L1 feed decision | UNRESOLVED |
| Product code/tests/runtime | NOT_RUN |
