# Phase 143 — exact Terra dispatch index

## Combined exact17 conditional acceptance sync (2026-08-20)

Status: `COMBINED_EXACT17_ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_COMMIT_REQUIRED / IMMUTABLE_EXACT_SHA_REVIEW_REQUIRED`. This working tree is not authority. Calendar exact12 `e60ff29934b8e53c34549cebae2e10657bae1540` (parent `e21a98a7459cd68690912c33ff031e286ba92ef1`, tree `5df5e079bdfe9685957e065e534c352cede25d2a`) and Mail correction exact9 `4f92ee6b9057efccd10c0b7ac7cf7dfd0f220c71` (parent `e60ff29934b8e53c34549cebae2e10657bae1540`, tree `2c65309e71c3501cd55516bc84c88c1c605df5a3`) each have Contract/Security/Lifecycle precommit and immutable exact-SHA PASS C/I/M `0/0/0`. This receipt becomes effective only after root creates one atomic commit changing exactly the 17 named acceptance-sync documents and Contract/Security/Lifecycle independently return immutable exact-SHA PASS for that exact commit. Until then all other Calendar candidate and Mail correction/suspension clauses in this file remain current; only after those conditions they automatically become `CONSUMED/HISTORICAL`, and the accepted routes below activate without another docs sync.

On activation, Calendar `e60ff299` becomes `SOURCE_AUTHORITY_ACCEPTED / EXACT_SHA_PASS` and root may dispatch only 143-03B-T exact2, test first: `backend/tests/calendar/calendar_attendee_schema_test.go`, then `backend/ent/schema/calendarattendee.go`. 143-04C exact3 remains `BLOCKED / NOT_AUTHORIZED` until 143-03B-T source itself is `STATIC_ACCEPTED`; there is no 04C source dispatch or acceptance now. Mail correction `4f92ee6b` becomes `SOURCE_AUTHORITY_ACCEPTED / EXACT_SHA_PASS` and root may dispatch only the frozen exact4, test first and in order: `backend/tests/mail/mailbox_events_test.go`, `backend/tests/mail/mailbox_sync_policy_test.go`, `backend/internal/mail/mailbox_events.go`, `backend/internal/mail/mailbox_sync_policy.go`. The current dirty exact4 candidates remain `UNCOMMITTED / UNACCEPTED` and require their own source static review, atomic commit and immutable exact-SHA acceptance lifecycle; runtime remains `NOT_RUN`.

The 143-04A/B/G receipts remain accepted. 143-03E, 144-02E, Ent generation, migration, launcher readiness and all executable/runtime evidence remain blocked, `NOT_READY` or `NOT_RUN` as already recorded. No source outside the two activated manifests, generation, migration, service expansion, API/UI, product, phase or release authority follows; raw iTIP remains prohibited.

> **Status:** 143_01_02A_02B_03A_03B_03C_03D_03F_04L_04A_04G_STATIC_ACCEPTED · RUNTIME_NOT_RUN
> **Next packet:** none selected. 143-04B is `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; root must separately dispatch any later Calendar packet. No Ent/schema migration/runtime authority follows.
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

> **143-04B accepted source receipt:** `75dafea889055126f768227a4ae62b28c91b7393` is `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; exact blobs in `IMPLEMENTATION-143-04B.md`. Tests first; Contract/Security/Lifecycle PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Old 04B `PRECOMMIT_FAIL`/`NOT_ACCEPTED`/current-next clauses `CONSUMED/HISTORICAL`. No next Calendar code packet selected; launcher `NOT_READY`, executables/generation/migration/runtime `NOT_RUN`; 03E Ent generation/migration blocked.

## Dependency graph

```text
143-01 ── 143-02A ── 143-02B
   └───── 143-03A ── 143-03B ── 143-03F
                  ├── 143-03C
                  └── 143-03D
contract 0.1.7 ── 143-04L shared versioned envelope
143-04L ── 143-04A ── 143-04G
143-03A/B/C/D (03F satisfied) ── 143-03E ── Ent generation/migration only

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
| 143-03B-T | attendee tombstone schema correction — `ROOT_SOURCE_AUTHORITY_CANDIDATE` / `PRECOMMIT_REVIEW_REQUIRED`; prior no-next Calendar packet clause is CONSUMED/HISTORICAL only for this sequential bridge | `backend/tests/calendar/calendar_attendee_schema_test.go`, then `backend/ent/schema/calendarattendee.go` | only `removed_at` representation, two live partial-index predicates, and schema-metadata oracle | accepted 143-03B |
| 143-03C | `CalendarEventException` occurrence schema — `STATIC_ACCEPTED` on `444c6f7b7b7b3d9b77bc9b24b32d1e13000ebbbc`; exact CHECK and external Ent-metadata evidence recorded; runtime `NOT_RUN` | `backend/ent/schema/calendareventexception.go` | event+recurrence unique, timed/all-day override fields | 143-03A |
| 143-03D | `CalendarMutationReceipt` plus narrow User/ScheduledCall edges — `STATIC_ACCEPTED` on `10b4ceead119cb74a16b325b33f9121f60d7b1b5` after final independent C/I/M `0/0/0`; evidence: `IMPLEMENTATION-143-03D.md`; runtime `NOT_RUN` | `backend/ent/schema/calendarmutationreceipt.go`, `user.go`, `scheduledcall.go`, `backend/tests/calendar/calendar_mutation_receipt_schema_test.go` | hashed key uniqueness, bounded response, optional O2O call | 143-03A..C |
| 143-03F | Durable bounded inbound-REPLY receipt schema and attendee edge — `STATIC_ACCEPTED` on `846cd00ecb243402877e18239c89a62bf07e8d14` after `SPEC PASS`/`QUALITY PASS` and final docs/whole-boundary review C/I/M `0/0/0`; the intermediate C/I/M `0/0/1` docs-sync finding is superseded/closed | `backend/ent/schema/calendarinboundreplyreceipt.go`, `calendarattendee.go`, `backend/tests/calendar/calendar_inbound_reply_receipt_schema_test.go` | external Ent metadata: exact named SHA-256 length CHECK and owned-child Cascade; runtime RED/GREEN `NOT_RUN` | 143-03B |
| 143-03E | Generate/inspect Ent output in accepted Docker job — 03A/B/C/D/F are accepted; blocked/`NOT_RUN` on the accepted launcher `NOT_READY` and the still-required founder-approved production migration mechanism/gate. The all-A product decision record does not itself establish that technical mechanism or execute/accept a migration. | generated Ent only, never manually edited or committed | clean generation and schema compile evidence | launcher + founder-approved mechanism/gate + 143-03A..D/F |
| 143-04L | Shared deterministic v1 event envelope over the unchanged outbox payload seam — `STATIC_ACCEPTED` at `4525c559da22729aa8bc51ff5336c29dad4f27e9` after the historical initial C/I/M `0/2/2`, independently accepted immutable/payload amendment, and final Terra/high PASS C/I/M `0/0/0`. Terra/xhigh implementation and Terra/high reviews are dispatcher-enforced; actual telemetry `UNKNOWN`. Evidence: `IMPLEMENTATION-143-04L.md`. Freeze: no exported mutable fields/references; JSON-object snapshot and frozen complete bytes; copy-returning `Subject()`/`EventID()` and `MarshalJSON`; semantic zero-offset UTC normalization; stable content-free `ErrInvalidPayload`. Launcher `NOT_READY`; compile/tests/runtime `NOT_RUN`; no Phase/product/global Mail acceptance. | `backend/internal/platform/events/envelope.go`, `backend/tests/calendar/event_envelope_test.go` | explicit inputs; strict subject/version/id/time/org/correlation/payload validation; complete-envelope immutability oracles; no I/O/random/clock | contract 0.1.7 |
| 143-04A | Public event repository/outbox/receipt ports and service constructor — `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `aa85f833a806a36f6d6ef2d8e7ae36dca2621574` (parent `ac9ed8e62edb9064f985a0ad7abf3a2ec42d7304`, tree `1e305267cf47f98000ce89d76f9204d9cd84df93`): `event_errors.go` `660578a29c24fe610d3c97072d922c05e876ae83`, `event_repository.go` `591608e87092ea80308a05923b22d201d4d124a9`, `event_service.go` `585fe73ecdc5b8b45d41c6f16bebc4cd7b69d5e0`, tests-first `event_service_contract_test.go` `713cb9599baeaaf16c1b95e833e93ee69a50cf08`; Terra/xhigh author, Terra/high static precommit+exact-SHA PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`; executable evidence `NOT_RUN`, launcher `NOT_READY` | `event_repository.go`, `event_service.go`, `event_errors.go`, `backend/tests/calendar/event_service_contract_test.go` | contract tests authored first; no persistence/runtime acceptance | 143-04L; former 03E/no-further-source block is `CONSUMED/HISTORICAL` |
| 143-04G | Admit calendar events to shared EVENTS stream without webhook exposure — `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `48a7273d100c07788def88e2f1b8d69c7173580f` (parent `6386f3eeb4025fc0aca55f49aaa8083ff856f8a9`, tree `65c5074a9793d24bac5f9b18aa214fbc942ac922`): `client.go` `e838704e6b4f5ebb9d1bb52d7d36b776c4453738`, tests-first `calendar_subjects_test.go` `71f5f68280eeb6b64e590b809c6df6b618695c9b`; Terra/xhigh author, Terra/high static precommit+exact-SHA PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`; executables `NOT_RUN`, launcher `NOT_READY`; former next-route is `CONSUMED/HISTORICAL` | `backend/internal/platform/nats/client.go`, `backend/tests/calendar/calendar_subjects_test.go` | add `calendar.>` once; exact webhook filter unchanged | 143-04A |
| 143-04B | `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `75dafea889055126f768227a4ae62b28c91b7393`; receipt in `IMPLEMENTATION-143-04B.md` | exact accepted source/test blobs | Contract/Security/Lifecycle PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN` | former current-next route `CONSUMED/HISTORICAL`; no next Calendar code packet selected |
| 143-04C | material/silent/reminder update and master cancellation — `ROOT_SOURCE_AUTHORITY_CANDIDATE` / `PRECOMMIT_REVIEW_REQUIRED`; eligible only after 143-03B-T `STATIC_ACCEPTED` | test first `event_service_update_test.go`, then `event_repository.go`, then new `event_service_update.go` | presence patch; authoritative locked organizer resolver; inactive-reader filtering/fresh re-invite/new ID/normalized-email role-resolved-identity rejection; opaque result echo/all-and-only rollback; typed recurrence-exception conflict | 143-03B-T + 143-04A/G |
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
