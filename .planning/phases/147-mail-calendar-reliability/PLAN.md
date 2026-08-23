# Phase 147 — Mail/Calendar Reliability and Integrations
> **147-02A static acceptance (2026-08-23):** source `3f4d174f888f08a119bbe381dc86e5e9b34e0301` is `STATIC_ACCEPTED / STATIC_SOURCE_ONLY` for exactly the sole production file `backend/internal/calendar/itip_ingest_policy.go`, after independent Contract/high and Security/xhigh APPROVED C/I/M `0/0/0`. Tests are intentionally absent under the current production-only constraint; all executable/generation/runtime evidence is `NOT_RUN`. This is pure REQUEST/CANCEL order policy only, without product, merge, readiness, schema, generation, migration, SQL, Docker, or runtime authority. D0 remains the sole next unblocked P3 schema atom; P3-C is independently PENDING/BLOCKED on Mail Ent generation, D1 follows D0 plus active 147-02A/B/C REQUEST/CANCEL, D2 follows Calendar generation, and P3-E requires P3-C+D1+D2+engine; 147-02B/147-02C remain pending/blocked on their stated dependencies.
> **P3-D0 due-context amendment (2026-08-23):** P3-D0 is sole next `UNIMPLEMENTED/PENDING/UNBLOCKED` atom: only `backend/ent/schema/calendarinboundmessagereceipt.go` adds immutable required scalar UUID `organization_id`/`mailbox_id`, no Ent edges. Admission/dedupe/conflict copies/rechecks sanitized org/mailbox/message/locator/envelope identity; due claim returns safe identities; raw content/provider/sender/method/auth absent; adapter/DB nonzero/cross-field checks where feasible. No generated/migrated/runtime row/backfill claim; generation launcher-gated. Previous P3-C-sole-next wording is `CONSUMED/HISTORICAL` only by this amendment; P3-C remains independently blocked. D1: pure Calendar `mail_inbound_repository.go`, `mail_inbound_service.go`, `mail_inbound_due_worker.go`, no Ent import/concrete source/wiring, after active 147-02A/B/C REQUEST/CANCEL+D0. D2: `event_ent_transaction.go` plus new `event_ent_inbound_message.go` after generation. P3-E=P3-C+D1+D2+engine. Worker 30s/batch50/concurrency4, claim-before-I/O; REPLY private helper/one tx/receipt last. All executable evidence `NOT_RUN`.

> Task: TASK-240 · Status: CONTRACT_0_1_8_ROOT_FROZEN_GLOBAL_REVIEW_PENDING_DEPENDENCY_BLOCKED · Depends on: Phases 143–146
> Required code runtime: `gpt-5.6-terra` / `xhigh`, no fallback
> Contract: `CONTRACT.md` · Dispatch: `DISPATCH-INDEX.md` + `DISPATCH-PACKETS.md`

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
