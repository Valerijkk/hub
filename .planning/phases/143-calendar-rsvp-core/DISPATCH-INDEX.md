# Phase 143 — exact Terra dispatch index

> **Status:** 143_01_02A_02B_03A_03B_03C_03F_STATIC_ACCEPTED · 143-04L_IMMUTABILITY_PAYLOAD_CONTRACT_AMENDMENT_STATIC_ACCEPTED_READY_FOR_CODE_REMEDIATION_REDISPATCH · RUNTIME_NOT_RUN
> **Required runtime:** `gpt-5.6-terra` / `xhigh`, no fallback
> **Execution:** strictly sequential where owned files overlap
> **Runtime evidence:** `.planning/phases/143-calendar-rsvp-core/RUNTIME-BLOCKER.md`

This index decomposes Phase 143 into one-behavior implementation packets. Every dispatch
must include the full Hub prompt contract, exact clean HEAD, current allowed files, external-
package tests first, no commit, and `BLOCKED` on scope or runtime mismatch. Root reviews
each result before the next overlapping task.

Complete standalone prompts exist for 143-01, 02A, 02B, 04G, 04H and 04L. For every other row,
assemble the Common envelope plus exactly one matching section from
`DISPATCH-REMAINING.md`; combining sections is invalid.

## Dependency graph

```text
143-01 ── 143-02A ── 143-02B
   └───── 143-03A ── 143-03B ── 143-03F
                  ├── 143-03C
                  └── 143-03D
contract 0.1.7 ── 143-04L shared versioned envelope
143-03A/B/C/D (03F satisfied) ── 143-03E + 143-04L ── 143-04A ── 143-04G

143-04A/G ──────────────── 143-04B create
143-04A/G ──────────────── 143-04C update/master cancel
143-04A/G ──────────────── 143-04D series RSVP ── 143-04E occurrence RSVP
143-02A + 143-04C/G ────── 143-04H occurrence cancel
143-02B + 143-04A/G ────── 143-04I inbound REPLY
143-01/02A + 143-04A/G ─── 143-04J call-projection policy ── 143-04K due policy

143-03E + 143-04A/G ───── 143-04F-Base
each accepted 143-04B/C/D/E/H/I ── its matching 143-04F-* transaction
143-03D + 143-04B/C/H/J ── 143-04F-J call bridge transaction
143-04F-J + 143-04K ────── 143-04F-K call due transaction

143-04B/C/D/E/H/J/K + matching F variants ── 143-05A ── 143-05B
143-05B + 143-04F-J ───── 143-05E legacy write bridge
143-05B/E ──────────────── 143-05C aggregate dedupe ── 143-05D
143-05E + 143-04F-K ───── 143-05F legacy call readers/worker ── 143-05G call calendar

143-06 production migration stays blocked until founder authority + launcher gates pass.
```

## Atomic packets

| ID | Single owned behavior | Exact primary files | Test oracle | Entry dependency |
|---|---|---|---|---|
| 143-01 | Timed/all-day validation and sequence/RSVP reset classification — `STATIC_ACCEPTED` on `e7d5f450`; evidence: `IMPLEMENTATION-143-01.md`; runtime `NOT_RUN` | `backend/internal/calendar/event_types.go`, `event_policy.go`, `backend/tests/calendar/event_policy_test.go` | table units for every policy row, including reminder-only version change | contract 0.1.7 (policy unchanged) |
| 143-02A | Deterministic RFC 5545/5546 serialization for REQUEST/CANCEL/REPLY — `STATIC_ACCEPTED` through `d73a9380`; evidence: `IMPLEMENTATION-143-02A.md`; runtime `NOT_RUN` | `itip_types.go`, `itip_serialize.go`, `backend/tests/calendar/itip_serialize_test.go` | exact in-memory CRLF/folding/escaping output; injected DTSTAMP | 143-01 |
| 143-02B | Bounded iTIP parsing and versioned CVE-v1 canonical duplicate key — `STATIC_ACCEPTED` through `efa7814d`; evidence `IMPLEMENTATION-143-02B.md` | `itip_parse.go`, `itip_limits.go`, `backend/tests/calendar/itip_parse_test.go` | in-memory malformed/depth/size/count/TZ/UID/SEQUENCE/spoof/equivalence/version fixtures | 143-02A + reviewed Mail 0.1.7 |
| 143-03A | `CalendarEvent` Ent schema — `STATIC_ACCEPTED` on `31bfff8f58ba83d12b59600135503f98a6c92fbd`; exact CHECK and external Ent-metadata evidence recorded; runtime `NOT_RUN` | `backend/ent/schema/calendarevent.go` | static schema review; generation NOT_RUN until launcher | 143-01 |
| 143-03B | `CalendarAttendee` schema with exact partial indexes — `STATIC_ACCEPTED` on clean `dd346e31`, C/I/M `0/0/0`; source commit `61c19f9f` remains the original boundary and current attendee source later includes allowed 03F inverse edge | `backend/ent/schema/calendarattendee.go` | exact NULL/non-NULL index predicates | 143-03A |
| 143-03C | `CalendarEventException` occurrence schema — `STATIC_ACCEPTED` on `444c6f7b7b7b3d9b77bc9b24b32d1e13000ebbbc`; exact CHECK and external Ent-metadata evidence recorded; runtime `NOT_RUN` | `backend/ent/schema/calendareventexception.go` | event+recurrence unique, timed/all-day override fields | 143-03A |
| 143-03D | `CalendarMutationReceipt` plus narrow User/ScheduledCall edges — no work claimed here; technically unblocked by accepted 03A–C | `backend/ent/schema/calendarmutationreceipt.go`, `user.go`, `scheduledcall.go` | hashed key uniqueness, bounded response, optional O2O call | 143-03A..C |
| 143-03F | Durable bounded inbound-REPLY receipt schema and attendee edge — `STATIC_ACCEPTED` on `846cd00ecb243402877e18239c89a62bf07e8d14` after `SPEC PASS`/`QUALITY PASS` and final docs/whole-boundary review C/I/M `0/0/0`; the intermediate C/I/M `0/0/1` docs-sync finding is superseded/closed | `backend/ent/schema/calendarinboundreplyreceipt.go`, `calendarattendee.go`, `backend/tests/calendar/calendar_inbound_reply_receipt_schema_test.go` | external Ent metadata: exact named SHA-256 length CHECK and owned-child Cascade; runtime RED/GREEN `NOT_RUN` | 143-03B |
| 143-03E | Generate/inspect Ent output in accepted Docker job — blocked on accepted 03D plus launcher; 03F is satisfied | generated Ent only, never manually edited or committed | clean generation and schema compile evidence | launcher + 143-03A..D |
| 143-04L | Shared deterministic v1 event envelope over the unchanged outbox payload seam — immutable-envelope/payload amendment is `CONTRACT_AMENDMENT_STATIC_ACCEPTED`, ready for code remediation redispatch after final Terra/high PASS C/I/M `0/0/0`, dispatcher-enforced provenance, actual telemetry `UNKNOWN`, docs commit pending. The 255-byte subject sub-amendment remains independently accepted; the prior implementation candidate remains review-failed/unaccepted, C/I/M `0/2/2`, until remediated. Freeze: no exported mutable fields/references; JSON-object snapshot and frozen complete bytes; copy-returning `Subject()`/`EventID()` and `MarshalJSON`; semantic zero-offset UTC normalization; stable content-free `ErrInvalidPayload`. No product code accepted; launcher `NOT_READY`, runtime `NOT_RUN`, global Mail review pending | `backend/internal/platform/events/envelope.go`, `backend/tests/calendar/event_envelope_test.go` | explicit inputs; strict subject/version/id/time/org/correlation/payload validation; complete-envelope immutability oracles; no I/O/random/clock | contract 0.1.7 |
| 143-04A | Public event repository/outbox/receipt ports and service constructor | `event_repository.go`, `event_service.go`, `event_errors.go` | compile-only consumer fakes in next behavior; no persistence | 143-03E + 143-04L |
| 143-04G | Admit calendar events to shared EVENTS stream without webhook exposure | `backend/internal/platform/nats/client.go`, focused genuine unit | add `calendar.>` once; exact webhook filter unchanged | 143-04A |
| 143-04B | Create event with auth-derived owner and response replay receipt | `event_service_create.go`, `backend/tests/calendar/event_service_create_test.go` | same key/input replays; changed input conflicts; one outbox | 143-04A + 143-04G |
| 143-04C | Material/silent/reminder update and master cancellation | `event_service_update.go`, `backend/tests/calendar/event_service_update_test.go` | sequence/reset/version table; exact event version/not-mutable conflicts; one master cancel event | 143-04A + 143-04G |
| 143-04D | Series RSVP optimistic transition | `event_service_rsvp.go`, `backend/tests/calendar/event_service_rsvp_test.go` | one version increment/outbox; NFR-13 recovery metadata | 143-04A + 143-04G |
| 143-04E | HTTP occurrence RSVP with persisted mutation receipt | `event_service_occurrence.go`, `backend/tests/calendar/event_service_occurrence_test.go` | materialized override; replay/conflict; one receipt/outbox | 143-04D + 143-04G |
| 143-04H | Cancel one recurrence instance | `event_service_occurrence_cancel.go`, `backend/tests/calendar/event_service_occurrence_cancel_test.go` | exact event-version/occurrence-not-mutable conflicts; cancelled exception; sequence/event version +1; occurrence iTIP CANCEL and one outbox | 143-02A + 143-04C + 143-04G |
| 143-04I | Sender-bound inbound iTIP REPLY apply/quarantine | `event_service_inbound_reply.go`, `backend/tests/calendar/event_service_inbound_reply_test.go` | A→B→A no rollback; same-status receipt; stale/future/spoof/cap quarantine | 143-02B + 143-04A + 143-04G |
| 143-04J | Pure canonical ScheduledCall mapping, recurrence and next-effective-occurrence policy | `event_call_policy.go`, `backend/tests/calendar/event_call_policy_test.go` | exact-channel/project/org membership; linked-guest rejection; series participant projection; DTSTART/RRULE/EXDATE; pause/resume; no I/O | 143-01 + 143-02A + 143-04A/G |
| 143-04K | Pure due-call orchestration and deterministic occurrence key | `event_call_due_service.go`, `backend/tests/calendar/event_call_due_service_test.go` | due/late/retry; one outbox descriptor; advance/deactivate; no event/reminder mutation | 143-04J/G |
| 143-04F-Base | Ent adapter shell and shared transaction primitive | `event_ent_repository.go`, `backend/tests/calendar/event_ent_repository_contract_test.go` | interface assertion; injected DB/outbox; no network/post-commit publish | 143-03E + 143-04A/G |
| 143-04F-B | Create-event Ent transaction | `event_ent_repository_create.go` | receipt/event/attendees/outbox/completion in one transaction | 143-04F-Base + 143-04B |
| 143-04F-C | Update/cancel-master Ent transaction | `event_ent_repository_update.go` | event-first lock/version/update/reset/outbox atomic | 143-04F-Base + 143-04C |
| 143-04F-D | Series-RSVP Ent transaction | `event_ent_repository_rsvp.go` | event→optional call→attendee; compare/update/series participant projection/receipt/outbox atomic | 143-03D + 143-04F-Base + 143-04D |
| 143-04F-E | Occurrence-RSVP Ent transaction | `event_ent_repository_occurrence.go` | event/exception/attendee lock order; inherited override/receipt/outbox atomic | 143-04F-Base + 143-04E |
| 143-04F-H | Occurrence-cancel Ent transaction | `event_ent_repository_occurrence_cancel.go` | event→optional call→exception lock; version+sequence/projection/outbox atomic | 143-04F-Base + 143-04H |
| 143-04F-I | Inbound-REPLY Ent transaction | `event_ent_repository_inbound_reply.go` | event→optional call→optional exception→attendee; receipt ledger + series-only participant projection | 143-03D + 143-04F-Base + 143-04I |
| 143-04F-J | Paired event/ScheduledCall create-link-edit-delete transaction | `event_ent_repository_call_bridge.go` | event→call lock; org/auth/link invariant; canonical mutation + projection + invite outbox atomic | 143-03D + 143-04F-Base + 143-04B/C/H/J |
| 143-04F-K | Canonical due projection transaction | `event_ent_repository_call_due.go` | event→call lock; current occurrence recheck; one due outbox + advance/deactivate atomic | 143-04F-J + 143-04K |
| 143-05A | Canonical v2 HTTP DTO binding/validation/error metadata | `event_dto.go`, `backend/tests/calendar/event_handler_contract_test.go` | owner ignored from client; temporal pairs; exact route-specific 409 codes/recovery headers; occurrence cancel | 143-04B/C/D/E/H + accepted 143-04F-B/C/D/E/H |
| 143-05B | `/api/calendar/v2/**` handlers and composition | `event_handler.go`, `backend/internal/calendar/handler.go`, `backend/cmd/server/wiring_modules.go` | exact OpenAPI paths; legacy handler untouched | 143-05A |
| 143-05C | Add first-class event to legacy backend aggregate without envelope drift or linked-pair duplicate | `dto.go`, `repository.go`, `service.go`, focused genuine backend units | linked pair once as canonical-backed `call`; unlinked event as `event`; existing filters/export preserved | 143-05B + 143-05E |
| 143-05D | Add the new aggregate event kind to the existing frontend calendar client | `frontend/src/api/calendar.ts`, focused genuine frontend unit | existing kinds byte-shape compatible; additive event decoder/type only | 143-05C + frontend feature-test harness |
| 143-05E | Delegate legacy ScheduledCall create/update/delete to canonical bridge | `backend/internal/calls/calendar_bridge.go`, narrow `scheduled.go`/`handler_scheduled.go`/`dto.go`, `backend/cmd/server/wiring_integrations.go`, both shipped create clients, genuine units in allowed roots | compatible success shape; stable key/replay; exact-match attach; paired atomicity; last-write-wins; exact `SCHEDULED_CALL_NOT_MUTABLE`; pause/resume vs delete | 143-05B + 143-04F-J |
| 143-05F | Canonicalize ScheduledCall list/conflict/due readers | narrow `backend/internal/calls/repository.go`/`scheduled.go`, `backend/cmd/server/wiring_integrations.go`, genuine policy/service unit | linked recurrence/exceptions; exact-channel detail auth; due uses 04F-K; pre-cutover unlinked bounded; post-cutover unlinked fails | 143-05E + 143-04F-K |
| 143-05G | Remove client-side dual recurrence authority from exact-channel call calendar | `frontend/src/api/calls.ts`, `frontend/src/modules/calls/call-calendar.tsx`, narrow `calendar-types.ts`, `frontend/tests/mail/**` | paired from/to; canonical occurrences; channel-room lifecycle; due.v1 refetch without toast; no raw RRULE | 143-05F + frontend feature-test harness |
| 143-06 | Founder-reviewed production migration/backfill and all-or-nothing cutover artifact | mechanism chosen by founder; `MIGRATION-PLAN.md` is authority meanwhile | full mapping, recurrence oracle, counts/hashes, no historical outbox, bridge postconditions | founder #12/#18 + launcher + 143-05E/F |

## Cross-packet hard rules

- Existing `Service`, `Repository`, `GET /api/calendar/events`, and `.ics` export remain
  operational; new canonical types use distinct names until additive integration.
- `CalendarEvent` never stores an occurrence `recurrence_id`; exceptions do.
- A linked `CalendarEvent` is the sole meeting authority. `scheduled_call_id` is immutable
  after create; no direct linked-call writer/worker or automatic mismatch winner is allowed.
- Nullable attendee uniqueness requires two partial PostgreSQL indexes.
- Event create and RSVP use `CalendarMutationReceipt`; mail submission uses
  `MailSubmission`. Outbox is not an idempotency response store.
- Every new producer builds one complete `events.EnvelopeV1` through 143-04L and passes it to
  the unchanged `PublishInTx(ctx, tx, envelope.Subject(), envelope)` seam. Stored event type,
  broker subject and JSON subject are identical. Envelope `event_id` is the application
  dedupe id; the outbox row UUID/`Hub-Event-Id` is a distinct diagnostic transport id.
  Legacy outbox payloads and the worker are not wrapped or rewritten.
- Inbound iTIP REPLY uses the bounded `CalendarInboundReplyReceipt` ledger, never a
  `CalendarMutationReceipt`; packet 143-04F-I is reviewed independently.
- Linked and reminder-relevant transactions share one global order: event → optional
  ScheduledCall → optional exception → optional attendee → stable user-policy guard →
  receipt/outbox/notification. Skipped row types do not change the order. This lets Phase
  147 serialize last-chance reminder revalidation without deadlock or a stale call/reminder
  projection.
- Phase 143 admits `calendar.>` to the shared EVENTS stream before any calendar producer;
  linked-call due uses versioned `calls.channel.{channelId}.scheduled_due.v1` inside the
  existing admitted `calls.channel.>`; Phase 144 later adds `mail.>`.
  None enters `WebhookFilterSubjects`.
- NFR-13 body remains exactly `{error,request_id}`. Machine conflict codes use
  `X-Hub-Error-Code`; stale RSVP additionally returns current-version `ETag` and a same-
  origin canonical-event `Location`.
- No generated Ent file is manually edited or committed. No schema migration is invented.
- No host test/build/lint/generate/validator command and no direct Docker/Compose fallback.
- No change enters protected QA/test-rebase paths.

## Gate status

| Gate | Status |
|---|---|
| Contract root/independent review | Calendar CVE-v1 identity semantics PASS on `b3fb6f48`; global Mail 0.1.7 later reopened for engine/schema bounds, and calendar list/legacy aggregate HTTP bounds are also P1-open |
| Requested-route provenance | 03B requested Terra/high independent re-review, dispatcher-enforced, actual telemetry `UNKNOWN`; 03F author Terra/xhigh and independent reviewer Terra/high, dispatcher-enforced, actual telemetry `UNKNOWN` |
| Accepted Docker launcher | NOT_READY |
| Ent generation | NOT_RUN |
| Production migration authority | UNRESOLVED |
| Runtime/tests/build | NOT_RUN |
