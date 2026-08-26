# Mail/Calendar completion execution plan

Authoritative specifications remain `CONTRACT.md`, `DISPATCH-PACKETS.md`,
`DISPATCH-INDEX.md`, the Phase 143-146 Mail/Calendar contracts, and Phase 148
release gates. This document only decomposes the remaining execution into SDD
tasks; it does not widen product scope or override a frozen contract.

## Global Constraints

- Work only on branch `codex/mail-module-20260813` in the existing isolated
  worktree. No push, PR, merge, deploy, secrets, or external side effects.
- Production implementation and test atoms use `gpt-5.6-terra` / `xhigh` with
  dispatcher-enforced provenance and no fallback. Independent reviews use
  `gpt-5.6-terra` / `high`.
- Host execution is limited to reading, editing, and Git. Build, lint,
  generation, tests, validators, backend/frontend processes, and runtime
  verification are Docker-only through accepted exact Compose services.
- Never read or modify `pro-qa-ai/**`. Preserve unrelated user changes. Do not
  manually edit generated Ent output.
- Keep receipts, errors, logs, events, and repository state content-free: no
  raw MIME/body, provider response, credentials, authentication evidence,
  recipient leakage, or free-form database errors.
- Every material event/exception mutation and its object state, outbox event,
  and final receipt transition are atomic. No material change means no outbox.
- Prefer one bounded Docker verification wave after source and wiring are
  complete to minimize disk I/O.

## Task 1: Complete 147-02C inbound invite persistence

Status: completed and statically accepted at Resolve `97ae294d`, Persist
`53869c07`, and owner-attendee correction `83cbe222`. Independent Terra/high
final review: `ACCEPT C/I/M 0/0/0`. Docker compile/runtime remains `NOT_RUN` and
is intentionally deferred to Task 6.

Modify only
`backend/internal/calendar/event_ent_repository_inbound_invite.go`. Implement
`eventEntTransactionBackend.PersistCalendarInboundInvite` and private helpers
needed by that method; add the compile-time interface assertion only after both
transaction methods exist. Do not add a Go DB/component test: the frozen packet
hands transaction/concurrency coverage to Python integration.

Bind the scope to the same active backend and revalidate the claim-locked
receipt, persisted owner/account authority, command/request/input/decision, and
the one DB UTC whole-second clock. Preserve the frozen lock order: receipt,
clock, Organization, active User, active-ready MailAccount, series state,
occurrence state, event, sorted attendees, exception, outbox, final receipt.
Never call the Mail service, open a nested transaction, follow a Calendar-to-Mail
Ent edge, or read raw/provider/body data.

Apply the decision table exactly:

- trusted series REQUEST creates, updates, or reactivates the imported event;
- trusted occurrence REQUEST mutates an exception only under an active matching
  master, otherwise persists `deferred_missing_master` with expiry exactly first
  deferral plus seven days;
- early CANCEL persists a safe tombstone without guessed event fields;
- series CANCEL cancels the master and every materialized exception without
  deleting their ordering high-watermarks;
- occurrence CANCEL cancels its matching exception or remains an unbound
  tombstone;
- stale, duplicate, conflict, resource-limit, and invalid revalidation never
  reach mutation in this method.

Advance `CalendarInboundObjectState` with the immutable owner, UID, optional
recurrence, organizer, sequence, method, canonical version/hash, exact state,
optional event, and expiry shape. Use Ent defaults for new UUIDs; row locks and
the frozen partial uniqueness prevent duplicates. Emit the accepted full v1
calendar invite envelope through the unchanged outbox only for a material
event/exception change, with `event_type == envelope.subject` and the command's
nonzero envelope ID. Return a valid
`CalendarInboundInvitePersistenceResult` for apply, tombstone, or defer. Receipt
finalization/worker lease policy remains Task 2 unless the existing transaction
boundary requires a narrow final-state call already frozen in the repository
contract.

Verification for this atom is static source review plus `git diff --check`.
Docker compile and Python transaction/concurrency scenarios are deferred to
Task 6 so the disk is not churned repeatedly.

## Task 2: Implement P3-D receipt repository, service, and due worker

Status: completed and independently accepted. `P3-D1` port is `8a7478dd`, the
shared transaction database clock is `4befb938`, Ent claim is `2455a123`, Ent
lock/completion is `4c030bd6`, the REPLY completion bridge is `8a238fcd`, the
orchestration service is `e1db48ce`, and the bounded due worker is `690f77e5`.
Every atom has independent Terra/high `ACCEPT C/I/M 0/0/0`; Docker compile and
runtime remain deferred to Task 6.

Execute Task 2 as four separately reviewed production atoms:

1. `P3-D1` owns immutable repository/service DTOs and ports only in new
   `backend/internal/calendar/mail_inbound_repository.go`. It freezes
   content-free receipt registration, database-clock claim result, due scan,
   and bounded cleanup shapes without importing Ent or Mail internals.
2. `P3-D2` owns the Ent claim/completion capability only in new
   `backend/internal/calendar/event_ent_inbound_message.go` plus the minimum
   compile-time assertion/wiring seam in `event_ent_transaction.go`. Exact
   token+generation+live-lease CAS and database-clock schedules are mandatory.
3. `P3-D3` owns orchestration only in new
   `backend/internal/calendar/mail_inbound_service.go`: commit claim, read the
   sealed Mail source outside a transaction, then use one Event transaction for
   receipt lock, invite/reply persistence and final completion.
4. `P3-D4` owns the 30-second, batch-50, concurrency-4 due worker and bounded
   30-day finalized cleanup only in new
   `backend/internal/calendar/mail_inbound_due_worker.go`.

Across all atoms enforce a 60-second database-clock lease; attempts 1..7
transient retry at `30s, 2m, 10m, 30m, 2h, 6h, 24h`; missing-master attempts
1..7 at `15m, 1h, 4h, 12h, 24h, 48h, 72h`; and terminalize attempt 8. First
deferral writes an immutable database-clock `+7d` expiry, later due values are
capped to it, and expired/attempt-8 missing-master work dead-letters before
Mail I/O. Only the frozen content-free error allowlist may be persisted.

## Task 3: Implement P3-E NATS ingress and production composition

Status: completed and independently accepted. Strict decoder is `f55eb408`,
bounded consumer is `f186302a`, iterator backoff is `32a6e221`, and stream plus
production composition is `839b2935`. Independent Terra/high final reviews are
`ACCEPT C/I/M 0/0/0`; Docker compile/runtime remains deferred to Task 6.

Add the Mail-owned sanitized v1 decoder/consumer and server wiring in the exact
P3-E files. Validate broker subject against JSON subject, deduplicate by
application `event_id`, drop the compatibility engine id, apply bounded ACK/NAK
semantics, and wire the Calendar repository/service/worker without exposing
authentication details or raw content.

## Task 4: Complete Mail/Calendar HTTP and OpenAPI surfaces

Connect the existing Mail domain services and Calendar integration to the
owner-scoped HTTP/BFF routes required by the frozen Mail contracts: folders,
message/thread list and reader, compose/reply/forward, attachments, search,
read/unread, flags, move/delete, invitation RSVP, availability, scheduling, and
mail-to-event/event-to-mail navigation. Preserve existing API compatibility and
authorization boundaries; update OpenAPI and typed clients in the same atomized
sequence.

## Task 5: Complete the minimal Apple-Mail-style Hub UI

Implement the accepted minimalist three-pane responsive Mail experience using
the existing Hub design system. Include folders, message list/reader,
compose/reply/forward, attachments, search, read/unread, flags, move/delete,
loading/empty/error states, keyboard/accessibility behavior, RSVP controls, and
bidirectional navigation between an invitation email and its Calendar event.
Expose colleague availability when scheduling meetings and never leak private
busy details.

## Task 6: Add testframework coverage and run bounded Docker verification

Add risk-based Python API/integration/e2e coverage in the founder greenfield
`testframework/` layout plus genuine isolated Go/React units only where the new
behavior has a real unit seam. Cover authorization, ordering, idempotency,
leases/recovery, rollback, one-outbox behavior, recurrence, RSVP, availability,
attachments, search/actions, and UI navigation. Run only accepted exact Compose
QA services with resource limits, then one full local stack smoke. Do not repeat
heavy builds unless a concrete failure requires it.

## Task 7: Close findings, synchronize docs, and prepare merge handoff

Fix Docker/runtime findings through reviewed atoms, update Mail/Calendar docs,
STATE, phase receipts, QA handoff, and Phase 148 gate evidence. Run final
whole-branch Contract/Security/Lifecycle/code/UI review, verify clean Git state
and current SHA, and prepare the MR handoff. Do not push or merge.
