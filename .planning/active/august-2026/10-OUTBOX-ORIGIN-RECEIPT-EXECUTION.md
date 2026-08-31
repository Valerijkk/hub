# OutboxOrigin and publication receipt implementation plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use
> superpowers:subagent-driven-development or superpowers:executing-plans to
> execute one checked packet at a time. Steps use checkbox syntax for tracking.

**Goal:** Add a shared durable OutboxEvent origin pair and receipt API, then
make only the three inbound R2 writers and exact cleanup proof use it.

**Architecture:** The shared platform layer receives immutable nullable
origin metadata and returns the physical outbox row plus EnvelopeV1 identity.
Mail and Calendar write their own origin inside their already-open transaction;
the QA cleanup controller later resolves each physical row through a single
indexed origin/type predicate. No QA authority table, preallocated ID, payload
search, or cross-domain FK is introduced.

**Tech Stack:** Go 1.25, Ent, Atlas mail-calendar-v1 lineage, PostgreSQL 17,
Docker Compose, existing Python integration suite.

**Spec:** .planning/active/august-2026/09-OUTBOX-ORIGIN-RECEIPT-CONTRACT.md
and the R2 addendum in
.planning/phases/144-mailbox-domain/CONTRACT-144-06-INBOUND-FIXTURE-CLEANUP.md.

## Global constraints

- P1a from 08-PROJECT-OUTBOX-SHARED-QA-EXECUTION.md is the first serial
  migration and owns 20260831000000. Do not create or hash 20260831000001
  until P1a is accepted.
- All source and generation/build/test work is Docker-only. This plan is not
  authority to run a host command or a Docker command now.
- Every implementation packet has one Terra/xhigh owner and an independent
  Terra/high review. Root owns admission and serial migration reservation.
- Do not read or modify pro-qa-ai, change public HTTP/OAS, create a new route,
  widen Publisher.Publish, add a fake Ent transaction, or add a payload scan.
- The original 15-node inbound fixture graph is immutable. The false sent
  branch proves drafts/submissions/additional messages only; it performs no
  absent mail.sent outbox lookup.

## File map

| Path | Responsibility |
|---|---|
| backend/ent/schema/outboxevent.go | declares the nullable immutable origin pair and composite index |
| backend/migrations/atlas/mail-calendar-v1/20260831000001_outbox_event_origin.sql | additive pair/check/index migration after P1a |
| backend/migrations/atlas/mail-calendar-v1/atlas.sum | Atlas hash entry after the new SQL |
| backend/docker/entrypoints/mail-calendar-schema.sh | advances exact lineage count/version allowlist after P1a |
| backend/migrations/README.md | records the additive lineage checkpoint without claiming execution |
| backend/internal/platform/outbox/publisher.go | additive origin value, receipt, and PublishInTxWithReceipt seam |
| backend/internal/qamailcalendar/controller_publish.go | P3 mail_message origin and receipt validation |
| backend/internal/calendar/event_ent_repository_inbound_invite.go | P4 calendar_event origin and receipt validation |
| backend/internal/calendar/event_ent_rsvp.go | P5 RSVP-attendee-specific helper and both RSVP paths |
| backend/internal/qamailcalendar/controller_delete.go | exact origin/type proof and three physical outbox deletes before the unchanged graph |
| testframework/tests/integration/mail-calendar/test_mail_calendar_sync.py | final public R2 scenario and cleanup lifecycle |
| testframework/cases/mail-calendar/TC-MAIL-CALENDAR-0002.md | traceability/oracle cleanup wording only if the public lifecycle changes |
| .planning/STATE.md and the two frozen contracts | final status/evidence synchronization only |

Generated Ent outbox files are produced only by the accepted Docker schema
generation path and are never hand-edited. The implementation packet records
their expected origin fields/predicates before accepting generated output.

---

### O0 — serial-migration and contract admission

**Depends on:** Project P1a independently accepted.

**Files:** planning receipts only; no source change.

- [ ] Confirm P1a created and accepted exactly
  20260831000000_qa_project_outbox_authority.sql, with atlas.sum and its
  launcher/README counts updated.
- [ ] Confirm the first free next serial name is exactly
  20260831000001_outbox_event_origin.sql. If it is occupied, stop and obtain
  a new root reservation; do not renumber or overwrite a migration.
- [ ] Independently review the 09 contract against current OutboxEvent,
  Publisher, P3/P4/P5 and R2 sources. Required verdict: shared nullable pair,
  no FK/QA table, no ID injection, and no payload scan.
- [ ] Record this packet as ADMITTED or BLOCKED. Docker/build/test evidence
  remains NOT_RUN.

**Acceptance:** No code is admitted before P1a and serial ownership are both
explicitly evidenced.

### O1 — additive schema and generated Ent admission

**Depends on:** O0.

**Owner files:** backend/ent/schema/outboxevent.go;
backend/migrations/atlas/mail-calendar-v1/20260831000001_outbox_event_origin.sql;
backend/migrations/atlas/mail-calendar-v1/atlas.sum;
backend/docker/entrypoints/mail-calendar-schema.sh; backend/migrations/README.md.

- [ ] Add Optional Nillable Immutable origin_kind and origin_id to
  OutboxEvent. Add only the non-unique origin_kind/origin_id/event_type index.
- [ ] Add the one forward migration. It adds nullable columns, the all-null or
  all-set check, and the composite index; it adds no FK, edge, trigger,
  backfill, data scan, or QA table.
- [ ] Advance the schema launcher from the accepted P1a lineage state to 11
  SQL files and latest version 20260831000001. Do not hard-code the prior
  pre-P1a count.
- [ ] Generate Ent only in the prescribed Docker schema operation. Confirm
  generated OutboxEvent creates expose both setters and predicates expose the
  three indexed fields. Do not hand-edit generated files.
- [ ] Independently review the migration for nullable-pair check, immutable
  schema declaration, non-unique index, absent FK/edge and no unrelated
  schema drift.

**Future evidence:** Docker schema generation/roundtrip and its exact
schema-maintenance verification are required later; both are NOT_RUN at plan
time.

### O2 — shared Publisher receipt seam

**Depends on:** O1 generated Ent admission.

**Owner file:** backend/internal/platform/outbox/publisher.go.

- [ ] Add the closed OutboxOriginKind constants, OutboxOrigin validation,
  OutboxPublicationReceipt, and PublishInTxWithReceipt exactly as frozen.
- [ ] Marshal payload and capture trace context once. For nil origin, create
  a row with both origin columns NULL. For non-nil origin, require an admitted
  kind, non-nil ID and a matching EnvelopeV1-shaped identity before Save.
- [ ] Build the receipt only from the saved entity and validated envelope
  identity. Return no receipt on error.
- [ ] Change old PublishInTx only to delegate with nil origin and discard the
  receipt. Do not modify Publish, worker code, or other callers.
- [ ] Add or update only a genuine isolated unit if a pure origin
  validation/receipt invariant is extractable without a fake client. Database
  persistence is not represented by a fake unit; it is deferred to O7 Docker
  integration evidence.
- [ ] Independently review backward compatibility: all existing nil-origin
  callers retain their call shape and no new validation applies to them.

**Acceptance:** A non-nil origin produces a fully matching receipt in the
same caller transaction; nil origin preserves old behavior; no generated or
caller-supplied physical ID exists.

### O3 — P3 Mail receipt propagation

**Depends on:** O2.

**Owner file:** backend/internal/qamailcalendar/controller_publish.go.

- [ ] Replace only the P3 PublishInTx call in publishInboundRequestFixture
  with PublishInTxWithReceipt and the literal mail_message origin using the
  already persisted fixture MailMessage ID.
- [ ] Before tx.Commit, validate physical ID, literal mail.received.v1,
  EnvelopeEventID equality, and exact origin echo. Preserve current rollback,
  publish_unknown and Controller.mu behavior; no mutex spans database work.
- [ ] Do not put the physical ID into a new durable binding record, change the
  public snapshot, schedule an ID, or add a route.
- [ ] Independently review that failed receipt validation restores the existing
  pre-commit state and an uncertain commit keeps the existing fail-closed
  publish_unknown path.

### O4 — P4 Calendar inbound-invite receipt propagation

**Depends on:** O3.

**Owner file:** backend/internal/calendar/event_ent_repository_inbound_invite.go.

- [ ] In appendCalendarInboundInviteOutbox only, replace the generic publisher
  call with PublishInTxWithReceipt using calendar_event and the already
  persisted calendarEventID.
- [ ] Validate the receipt against the immutable descriptor/envelope, literal
  calendar.invite.updated.v1 and exact event origin before backend.tx commits.
- [ ] Preserve command outcome, transaction ownership, database clock,
  receipt completion and generic AppendOutbox behavior. Do not propagate
  origin through generic EventTransaction interfaces.
- [ ] Independently review apply/tombstone descriptor handling so neither
  branch can attach a mismatched event origin.

### O5 — P5 RSVP receipt propagation

**Depends on:** O4.

**Owner file:** backend/internal/calendar/event_ent_rsvp.go.

- [ ] Add one private RSVP-only outbox helper that accepts the exact
  EventOutboxDescriptor plus the post-CAS CalendarAttendee ID and calls
  PublishInTxWithReceipt on backend.tx with calendar_attendee.
- [ ] Replace only the series and occurrence RSVP AppendOutbox calls with
  that helper. Both paths validate physical ID, literal
  calendar.rsvp.changed.v1, descriptor EnvelopeEventID and attendee origin.
- [ ] Do not alter EventTransaction.AppendOutbox,
  eventEntTransactionBackend.AppendOutbox, or non-RSVP Calendar emission
  paths; they lack a justified attendee origin.
- [ ] Independently review that the origin is the updated/applied attendee,
  belongs to the exact event, and cannot be an actor or arbitrary attendee.

### O6 — R2 exact-origin proof and cleanup

**Depends on:** O3, O4 and O5.

**Owner file:** backend/internal/qamailcalendar/controller_delete.go. A small
qa-and-not-release helper file may be admitted only if it isolates
origin-row proof without changing public routes or the 15-node deletion logic.

- [ ] Extend the pre-delete proof in fixed causal order: original roots and
  MailMessage, P3 origin lookup, constrained CalendarEvent plus P4 lookup,
  constrained RSVP attendee plus P5 lookup.
- [ ] Each lookup uses exactly origin_kind, origin_id and event_type and
  requires one row. Decode and validate only that returned row's EnvelopeV1,
  event ID and required relationship; do not issue a payload predicate or
  query arbitrary OutboxEvent rows.
- [ ] Before the existing graph deletes, delete exactly the three proven
  physical OutboxEvent IDs in the same cleanup transaction. Preserve every
  existing original-node proof, order, rollback and engine-destruction rule.
- [ ] Keep sent_finalized=false negative proof to exact account
  drafts/submissions/original-message cardinalities. Remove all scheduled or
  absent mail.sent physical-ID assertions.
- [ ] On cardinality, envelope, relation, origin, state or commit uncertainty,
  leave the binding fence unchanged and return the existing non-mutating 409.
  Never retry creation or broaden cleanup after restart.
- [ ] Independently review mapless restart and concurrency: origin lookup is
  the only physical-row recovery authority, duplicate rows remain untouched,
  and no QA binding table or payload scan appears.

### O7 — traceability, focused tests and Docker-only evidence

**Depends on:** O1 through O6 independently accepted.

**Owner files:** testframework/tests/integration/mail-calendar/test_mail_calendar_sync.py;
testframework/cases/mail-calendar/TC-MAIL-CALENDAR-0002.md if needed;
.planning/STATE.md; the two frozen contracts.

- [ ] Extend TC-MAIL-CALENDAR-0002 only through public HTTP behavior and its
  existing fixture lifecycle. It must verify correlation and accepted RSVP,
  then its finally cleanup proves the public exact cleanup result. It receives
  no database, NATS, raw outbox ID, origin value or direct cleanup query.
- [ ] Add focused source-level/genuine-unit coverage only for independently
  testable pure validations introduced in O2. Do not create a fake Ent
  transaction or a source-grep test.
- [ ] Run the exact Docker services only after root authorizes the completed
  batch: schema maintenance first as required by the accepted lineage, then
  docker compose -f docker-compose.local.yml run --rm backend-unit and
  docker compose -f docker-compose.local.yml run --rm python-integration.
- [ ] Record each service as PASS, FAIL, BLOCKED or NOT_RUN independently.
  A Python contract pass, static review, or generated source alone is never
  product/runtime evidence.
- [ ] Update STATE and this contract's evidence status only with actual
  results; retain P1-P10 and all unavailable services as NOT_RUN.

## Review matrix

| Requirement | Owning packet | Fail-closed evidence |
|---|---|---|
| pair is both-null or both-set, immutable, indexed and FK-free | O1 | schema review and Docker generation |
| old publisher callers unchanged | O2 | compile/runtime evidence after O7 |
| P3/P4/P5 each bind exact existing domain object | O3/O4/O5 | receipt validation before writer commit |
| restart never needs a generated physical ID | O6 | exact triple lookup plus cardinality one |
| no payload scan or QA authority table | O2/O6 | source/security review |
| false sent branch adds no send artifacts | O6/O7 | account-scoped public lifecycle evidence |
| Project P1b can reuse receipt but QA data stays separate | O0/O2 | contract review |

## Completion boundary

This plan is complete as documentation only. Implementation is BLOCKED on P1a
acceptance and serialized migration admission. No Docker command, code,
generated artifact, test result, commit, merge or production-readiness claim
has been made here.

## R4 authoritative O2/O4/O6 clarification

This append-only clarification supersedes only earlier O2/O4/O6 instructions
that could be read to cover a non-ApplyRequest P4 branch, omit the strict
decoder ABI/order, or permit an aliased receipt origin. It preserves O0/P1a
serial admission, O1 schema scope, P3/P5 ownership, the immutable pair,
no-FK/no-QA-table/no-payload-scan rules, false-sent boundary, transaction
ownership, review gates and NOT_RUN evidence status.

O2 owners are exactly `backend/internal/platform/events/envelope.go` and
`backend/internal/platform/outbox/publisher.go`, in that order. First define
the additive public `events.DecodeEnvelopeV1([]byte) (events.EnvelopeV1,
error)` ABI: `json.Decoder.DisallowUnknownFields`; exactly one top-level
object; no trailing token; schema-version/required-field validation through
`NewEnvelopeV1`; then re-marshal/canonical-byte equality. It is not used by
Publisher for a freshly supplied in-memory envelope and has no outbox query.
Then PublishInTxWithReceipt validates/writes through the already-open
transaction and returns a receipt whose nullable Origin is nil or a fresh
defensive value copy, never an alias of the input pointer.

O4 owner is exactly
`backend/internal/calendar/event_ent_repository_inbound_invite.go`. Replace
generic publication with origin-bearing receipt publication only where
`command.Outcome() == CalendarInboundIngestApplyRequest` and the descriptor is
`calendar.invite.updated.v1`; validate the calendar_event receipt before the
existing backend transaction commits. Inbound tombstone series/occurrence and
cancelled-envelope branches remain existing generic `PublishInTx` branches
with no origin, receipt, R2 lookup or new validation.

O6 owner is exactly `backend/internal/qamailcalendar/controller_delete.go`
(or its separately admitted qa-and-not-release proof helper). After each
indexed `origin_kind/origin_id/event_type` lookup proves cardinality one, and
only then, call `events.DecodeEnvelopeV1` on that returned row's payload;
prove decoded EventID and the existing relation before any physical delete.
O6 may neither select by payload nor invoke the decoder on an unselected row.
The dependency order remains P1a/O0 -> O1 -> O2 -> O3 -> O4 -> O5 -> O6.

## R5 authoritative inbound cleanup-lifecycle replacement

Status: `DOCS_FROZEN` / `NOT_RUN`. This R5 supersedes only R2-R4 binding-state
granularity, P4/P5 Controller callbacks, mapless-absent `204`, and cleanup
proof-retention directions. It preserves O1-O6 ordering/origin semantics,
false `sent_finalized`, the 15-node closure, and public `204`-or-`409`.

No production P4/P5 Controller observer hook is admitted. Existing
process-local state becomes `QAInboundFixtureBindingV2`. `Published` records
only a committed P3 and cannot authorize cleanup. In one cleanup transaction,
first prove immutable 15-node roots plus Calendar closure, zero MailDraft,
zero MailSubmission, exactly one inbound MailMessage, and exactly three
cardinality-one origin/type rows. The exact tuples are P3
`mail_message`/MailMessage/`mail.received.v1`, P4
`calendar_event`/CalendarEvent/`calendar.invite.updated.v1`, and P5
`calendar_attendee`/RSVP attendee/`calendar.rsvp.changed.v1`; decode each
selected row strictly and canonically and prove EventID, origin, type and
relation.

After full proof, CAS that same immutable V2 record `Published` ->
`CleanupClaimed`, retaining only value copies of Calendar receipt, objectState,
event, RSVP attendee, replyDelivery and mutationReceipt IDs. Never retain,
schedule, or derive a physical OutboxEvent ID. Re-prove in that transaction,
delete returned physical IDs P3 -> P4 -> P5, then preserve the existing
15-node deletion order. Known pre-commit failure rolls back and CASes to
`Published` with proof cleared. Commit error is `CleanupCommitUnknown` with
domain proof retained: exact present restores `Published`; exact all-known-ID
absence may proceed to engine destroy; partial returns `409`. Confirmed commit
then destroys the engine outside the transaction: success -> `Cleaned`, failure
-> terminal `CleanupEngineFailed`/`409`; same-process `Cleaned` is `204` only
after exact known-ID absence. Mapless restart derives only a full present graph
from deterministic roots and cleans DB-only; absent, partial, ambiguous, and
mapless commit-error states return `409`, never guessed `204`.

Execution packets are: roots proof, external-closure guard, Calendar proof,
O6 exact outbox proof, V2 lifecycle coordinator, Python `finally`/tests, then
authorized Docker gates. No persistent QA table/schema, payload/broad/name/
nonce scan, `mail.sent` lookup, or physical-ID scheduling/retention is allowed.

### R5 strict V2 lifecycle and reconciliation rules

`QAInboundFixtureBindingV2` is one exact comparable private value: existing
nonce, immutable plan identity, engineFixture, persisted roots, and state,
plus `cleanupProof`. `cleanupProof` has an explicit presence bit and value
copies of exactly six UUIDs: CalendarInboundMessageReceipt,
CalendarInboundObjectState, CalendarEvent, RSVP CalendarAttendee,
CalendarReplyDelivery, and CalendarMutationReceipt. It contains no pointer,
envelope, row, or physical OutboxEvent ID. Controller map and Fake engine are
created or replaced together at boot; all record mutation is private; a
`Cleaned` record is retained and is never deleted.
The cleanupProof is absent in Creating, Created, Deleting, Publishing,
PublishUnknown, Published, and CleanupFailed; it is present and exact in
CleanupClaimed, CleanupCommitUnknown, CleanupEngineFailed, and Cleaned.
CleanupCommitUnknown exact-present -> Published clears cleanupProof in that CAS.

Legal V2 states retain `Creating`, `Created`, `Deleting` (created-only legacy),
`Publishing`, `PublishUnknown`, `Published`, and `CleanupFailed`
(created-only legacy), and add `CleanupClaimed`, `CleanupCommitUnknown`,
`CleanupEngineFailed`, and `Cleaned`. Existing create and P3 transitions are
unchanged. `Published -> CleanupClaimed` requires the complete proof below.
`CleanupClaimed` permits only confirmed rollback -> `Published`, commit error
-> `CleanupCommitUnknown`, or confirmed commit plus engine success -> `Cleaned`
/ engine failure -> terminal `CleanupEngineFailed`. `CleanupCommitUnknown`
permits exact-present -> `Published`, exact-absent -> the single engine step
then `Cleaned` or `CleanupEngineFailed`. `Cleaned` is same-process idempotent;
`CleanupEngineFailed` is terminal. Every CAS under `Controller.mu` compares
nonce, plan identity, persisted roots, engineFixture, prior state, and expected
cleanupProof presence/value; database, engine, waits, and HTTP remain outside
the mutex.

Published cleanup and reconciliation use Ent `BeginTx` at `SERIALIZABLE`; all
selected exact rows/sets are locked, and the full predicate plus external
closure reads occur in that transaction through commit. Normal
`READ COMMITTED` cleanup is forbidden. A serialization/deadlock failure is
pre-commit only after rollback is confirmed; only then may a renewed full proof
precede restoring `Published`. Otherwise return `409` or retain
`CleanupCommitUnknown`.

`exact-present` means the full closed 15-node graph, external-closure guard,
all three cardinality-one origin proofs, and false-sent guard. `exact-absent`
means every deterministic root/relation ID is absent; all six retained Calendar
IDs are absent; each exact origin triple has zero rows; MailDraft and
MailSubmission sets are zero; and the original inbound MailMessage is absent.
Physical OutboxEvent IDs are never retained or required.

Before the synchronous public RSVP has succeeded and the full P5 closure,
DELETE is non-mutating `409`. A successful public RSVP response after the
completed mutation receipt is the readiness signal. Python `finally` invokes
published cleanup only after that signal; an earlier body failure may make
cleanup fail while preserving the primary test exception.

After full proof, CAS `Published -> CleanupClaimed`, re-prove in the same
transaction, then delete transient selected physical OutboxEvent IDs in this
exact order: P3 -> P4 -> P5; then delete the unchanged closed graph in exact
order: ReplyDelivery -> MutationReceipt -> InboundReceipt -> ObjectState ->
Attendee -> Event -> messageFolder -> message -> thread -> folder -> identity
-> account -> domain -> user -> organization.

Restore `Published` and clear cleanupProof only after rollback success is
confirmed. Rollback error or unavailable rollback produces
`CleanupCommitUnknown` and retains the proof. After DB commit and engine
destroy success, a failed or lost CAS to `Cleaned` returns `409`; do not retry
the DB or engine operation and retain current claim/unknown evidence.

A mapless complete-present graph is admissible only while a monotonic
fresh-boot-unmutated Controller flag proves the Controller and its task-owned
Fake were constructed together and that Fake has accepted no seed or mutation
in this boot; it then performs DB-only cleanup. The flag starts true only in
that constructor path and becomes permanently false before the first QA seed
or engine mutation. A missing record with a false flag, absent/partial/
ambiguous graph, or mapless commit error
returns fail-closed `409`, never guessed `204`. No persistent QA table/schema,
payload/broad/name/nonce scan, `mail.sent` lookup, or physical-ID scheduling or
retention is allowed. Status remains `DOCS_FROZEN` / `NOT_RUN`.
