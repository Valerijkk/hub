# Phase 147 — Mail/Calendar Reliability and Integrations

> Task: TASK-240 · Status: CONTRACT_0_1_6_ROOT_REVIEW_PENDING_INDEPENDENT_PENDING_DEPENDENCY_BLOCKED · Depends on: Phases 143–146
> Required code runtime: `gpt-5.6-terra` / `xhigh`, no fallback
> Contract: `CONTRACT.md` · Dispatch: `DISPATCH-INDEX.md` + `DISPATCH-PACKETS.md`

## Outcome

Close deterministic incoming invitation ordering and outward REPLY recovery, private
availability, durable reminder scheduling and the CRM manual-mail seam. Keep L1 feed code
conditional on founder question 5. L2 external mailbox ingest and L3 OAuth remain outside
this phase.

## Ordered execution

1. Independently review Mail 0.1.6 reliability/bridge semantics and close contract findings.
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
| Root reliability contract static review | 0.1.6 PASS on exact SHA `5a351d5b`; 0.1.5 `eae6f5e5` historical STALE |
| Independent reliability contract review | Terra/high PASS on exact SHA `5a351d5b`; no P0/P1/P2 |
| Exact Terra code runtime | AUTHORIZED_BY_FOUNDER / first actual packet pending |
| Phase 143/144 accepted implementation | NOT_IMPLEMENTED |
| Engine authentication mapping | BLOCKED_BY_PHASE_145 |
| Accepted Docker launcher | NOT_READY |
| Ent generation/migration | NOT_RUN / FOUNDER_DECISION_PENDING |
| L1 feed decision | UNRESOLVED |
| Product code/tests/runtime | NOT_RUN |
