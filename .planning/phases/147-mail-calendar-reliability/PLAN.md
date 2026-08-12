# Phase 147 — Mail/Calendar Reliability and Integrations

> Task: TASK-240 · Status: PLANNED · Depends on: Phases 143–146

## Ordered tasks

1. Prove out-of-order/equal/lower sequence reconciliation and occurrence exceptions.
2. Complete incoming REQUEST/CANCEL and external-organizer REPLY recovery/dead-letter
   behavior with parser bounds and sender authentication results.
3. Implement privacy-preserving availability across events, calls and approved leave.
4. Route event reminders through existing notification preferences and dedupe semantics.
5. Implement L1 iCal feed subscription only if still accepted: SSRF-safe fetch, conditional
   requests, bounded recurrence, source ownership, disable/delete and stale status.
6. Add CRM authorized compose/send and manual activity link without campaign semantics.
7. Update units/contracts/docs and hand non-unit scenarios to QA.

L2 external mailbox ingest and L3 OAuth remain outside this phase.
