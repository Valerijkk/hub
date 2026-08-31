# Shared OutboxOrigin and publication-receipt contract

> **Status:** ROOT_FROZEN / DOCS_ONLY / NOT_RUN  
> **Decision:** D — shared immutable OutboxEvent origin metadata, not QA ID injection or a QA binding table.  
> **Prerequisite:** Project Outbox P1a owns and must first accept migration 20260831000000_qa_project_outbox_authority.sql. This work owns only the next serial migration, 20260831000001_outbox_event_origin.sql, after that acceptance.

## Goal

Give an already-created domain object an exact, durable and restart-safe way to
own its outbox row without coupling the common outbox table to Mail, Calendar,
or any QA-only authority. The result is one physical-row receipt for a writer
and one indexed, constrained lookup path for cleanup/reconciliation.

This contract changes no runtime behavior by itself. Ent generation, migration
application, Docker, tests, build, deployment, merge, and product acceptance
remain NOT_RUN.

## Frozen data model

OutboxEvent gains this nullable immutable pair:

| Column | SQL / Ent shape | Meaning |
|---|---|---|
| origin_kind | nullable text / Optional Nillable Immutable string | closed shared origin discriminator |
| origin_id | nullable UUID / Optional Nillable Immutable UUID | exact domain row identity |

The migration must add exactly:

1. CHECK ((origin_kind IS NULL AND origin_id IS NULL) OR (origin_kind IS NOT NULL AND origin_id IS NOT NULL)).
2. A non-unique index over (origin_kind, origin_id, event_type), named by the
   migration in the local lineage convention.

The index is deliberately non-unique. The generic platform must not silently
collapse legitimate future repeated events for one object. The selected inbound
flow establishes one row through its own transaction/state fences; a lookup
whose cardinality is not exactly one fails closed.

There is no FK, Ent edge, cascade, trigger, polymorphic relation, QA table, or
cross-domain import. Ent immutability protects ordinary writers; every
origin-bearing writer creates the pair once and never updates it. Raw SQL is
not a second authority.

The only admitted initial kinds are platform constants with literal values:

| Constant | Stored origin_kind | origin_id |
|---|---|---|
| OutboxOriginKindMailMessage | mail_message | MailMessage.ID |
| OutboxOriginKindCalendarEvent | calendar_event | CalendarEvent.ID |
| OutboxOriginKindCalendarAttendee | calendar_attendee | CalendarAttendee.ID |

No caller may pass an empty kind, an unknown kind, uuid.Nil, or a half-present
pair. New kinds require a separately reviewed additive contract and migration
impact review; they are not accepted by an arbitrary string path.

## Publisher API and receipt

backend/internal/platform/outbox/publisher.go owns these additive platform
types and method:

    type OutboxOriginKind string

    type OutboxOrigin struct {
        Kind OutboxOriginKind
        ID   uuid.UUID
    }

    type OutboxPublicationReceipt struct {
        OutboxEventID   uuid.UUID
        EnvelopeEventID uuid.UUID
        EventType       string
        Origin          *OutboxOrigin
    }

    func (p *Publisher) PublishInTxWithReceipt(
        ctx context.Context,
        tx *ent.Tx,
        eventType string,
        payload any,
        origin *OutboxOrigin,
    ) (OutboxPublicationReceipt, error)

Existing PublishInTx remains source and behavior compatible:

    func (p *Publisher) PublishInTx(
        ctx context.Context,
        tx *ent.Tx,
        eventType string,
        payload any,
    ) error {
        _, err := p.PublishInTxWithReceipt(ctx, tx, eventType, payload, nil)
        return err
    }

PublishInTxWithReceipt performs the old one-time JSON marshal, trace-context
capture and Create().Save(ctx) in the supplied open transaction. It neither
opens, commits, rolls back, retries, nor retains that transaction. A nil
origin writes both database columns NULL and preserves old callers. A non-nil
origin requires a valid admitted kind, non-nil ID, and an EnvelopeV1-shaped
payload whose non-nil EventID and Subject equal eventType. It writes the exact
pair and returns only after Save succeeds.

The returned receipt is a value copy. OutboxEventID is the generated physical
primary key. EnvelopeEventID is the validated EnvelopeV1 event identity.
EventType and Origin echo the accepted inputs. Any marshal, validation, save,
or receipt-consistency failure returns an error and no success receipt. The
method must not discover an event by payload query, mutate a prior row, inject
a physical ID, or make a QA-specific interface.

The existing autonomous Publish API is out of scope. No existing
PublishInTx caller changes merely to attach an origin; only the three explicit
inbound writers below do so.

## Selected inbound writer ownership

All three writes happen inside their existing writer-owned Ent transaction.
Each writer calls PublishInTxWithReceipt, validates the returned receipt before
its owner commits, and maps a failure through its current fail-closed error
path. The writer does not allocate or predict the physical OutboxEvent ID.

| R2 phase | Owner file and boundary | event_type | Exact origin |
|---|---|---|---|
| P3 | backend/internal/qamailcalendar/controller_publish.go, publishInboundRequestFixture transaction | mail.received.v1 | mail_message / persisted mailMessageID |
| P4 | backend/internal/calendar/event_ent_repository_inbound_invite.go, appendCalendarInboundInviteOutbox on backend.tx | calendar.invite.updated.v1 | calendar_event / calendarEventID |
| P5 | backend/internal/calendar/event_ent_rsvp.go, new RSVP-specific append helper on backend.tx | calendar.rsvp.changed.v1 | calendar_attendee / applied attendee ID |

P5 must not change generic EventTransaction.AppendOutbox or
eventEntTransactionBackend.AppendOutbox. Those paths serve unrelated Calendar
event types and lack the authoritative RSVP attendee identity. The two series
and occurrence branches pass their exact post-CAS attendee ID to the new
private RSVP helper. P4 similarly changes only the inbound-invite helper, not
the generic Calendar append path.

For every selected writer the receipt must prove:

1. physical OutboxEventID is non-nil;
2. receipt EventType equals the expected literal;
3. receipt EnvelopeEventID equals the exact descriptor/envelope EventID;
4. receipt Origin equals the exact kind and already-persisted origin row ID;
5. the same transaction already contains the corresponding root/relationship.

P3's origin is the known fixture MailMessage. P4's CalendarEvent is the one
already reached from the exact inbound receipt/message proof. P5's
CalendarAttendee is the attendee updated by the RSVP CAS and belongs to that
same proven CalendarEvent. A bad receipt is an error before commit, never an
alternative lookup.

## Exact proof and cleanup/restart rules

R2 uses no preassigned physical outbox ID and no payload scan. It proves each
selected row only by the indexed exact predicate:

    origin_kind = literal AND origin_id = exact proven ID AND event_type = literal

Each query must return exactly one OutboxEvent. The returned row is then
validated for non-null immutable pair equality, exact event_type, strict
EnvelopeV1 decoding, decoded EventID, and the event/message/attendee relation
already required by the R2 graph proof. This is a direct constrained lookup,
not a scan: no broad OutboxEvent query, ID NOT IN, aggregate, name/nonce
prefix, organization inference, payload predicate, or generated-ID schedule
is allowed.

Traversal order is fixed:

1. prove the deterministic original roots, including MailMessage;
2. resolve P3 by mail_message plus mail.received.v1;
3. use the existing constrained inbound receipt/message traversal to prove the
   single CalendarEvent and resolve P4 by calendar_event plus
   calendar.invite.updated.v1;
4. use the existing constrained event/attendee traversal to prove the single
   RSVP attendee and resolve P5 by calendar_attendee plus
   calendar.rsvp.changed.v1;
5. only after all proofs pass, delete those three physical OutboxEvent IDs in
   the existing cleanup transaction before the unchanged 15-node graph.

Any absent required row, two or more rows, null/half pair, wrong kind/id/type,
bad envelope, wrong envelope ID, unrelated event/attendee/message, a changed
relationship, transaction failure, or unavailable query is non-mutating 409.
The binding state remains fenced on uncertainty; it never retries a writer,
invents a row, deletes by selector expansion, or relies on process memory.

After a process restart, the same origin predicates rebuild physical ownership
from proven roots. A normal in-memory binding may observe the writer receipt,
but it does not persist a second authority and it is not required for restart.
Concurrent calls are serialized by the existing per-fixture state/row fences.
If a competing committed write creates duplicate selected rows, exact
cardinality fails closed; cleanup touches neither one.

## False sent branch

For inbound-request-v1, sent_finalized remains literally false. There is no
mail.sent.v1 origin query, absent OutboxEvent ID assertion, generated sent ID,
MailDraft, MailSubmission, MailSubmissionResult, or additional result
MailMessage. The negative proof remains limited to the exact proven account:
zero drafts, zero submissions, and exactly the original inbound MailMessage.
The created 15-node graph and its deletion order do not change.

## Project Outbox boundary

Project P1a remains the first migration authority at
20260831000000_qa_project_outbox_authority.sql. Project P1b may later reuse
the additive PublishInTxWithReceipt API to obtain an actual physical
OutboxEventID, but it must retain its own QA manifest/effect authority and
tables. This contract neither writes nor reads qa_project_outbox_manifest or
qa_project_outbox_effect, and it does not generalize those QA tables.

## Admission and evidence

Implementation is blocked until P1a is accepted with its reserved migration
slot and an independent contract/security review admits the execution packet.
The next migration must be exactly 20260831000001_outbox_event_origin.sql; a
different occupied slot requires a new serial reservation before source work.
All atomic implementation work uses gpt-5.6-terra/xhigh with
dispatcher-enforced provenance and independent Terra/high review. Docker-only
evidence remains NOT_RUN.

## R4 narrow P4, decoder, receipt, and execution-order clarification

This append-only clarification supersedes only earlier clauses that could be
read to attach the P4 origin/receipt to another inbound outcome, to leave
receipt Origin pointer aliasing unspecified, or to leave stored-envelope
decoding unspecified. It does not alter the nullable immutable pair, accepted
origin kinds, P1a prerequisite, serial slot, no-FK/no-QA-table/no-payload-scan
rule, false-sent boundary, transaction ownership, P5 scope, or NOT_RUN.

P4 applies only where `CalendarInboundIngestApplyRequest` produces
`calendar.invite.updated.v1`. Its exact origin is `calendar_event` plus the
already-persisted CalendarEvent ID. Inbound tombstone series/occurrence and
cancelled-envelope branches retain generic `PublishInTx`; they receive no
origin, no receipt, no R2 proof and no cleanup ownership from this packet.

The additive public ABI is
`events.DecodeEnvelopeV1(encoded []byte) (events.EnvelopeV1, error)` in
`backend/internal/platform/events/envelope.go`. It must use a JSON decoder
with `DisallowUnknownFields`, accept exactly one top-level object and no
trailing token, validate schema version and every required field by calling
`NewEnvelopeV1`, then marshal the returned envelope and require canonical
byte-for-byte equality with `encoded`. It rejects malformed, unknown,
missing, duplicate, noncanonical, whitespace-altered, invalid or trailing
representations. The decoder has no query capability and its sole admitted R2
call is O6 after an `origin_kind/origin_id/event_type` lookup has returned
exactly one physical OutboxEvent.

`OutboxPublicationReceipt.Origin` is nil for a nil accepted origin. For a
non-nil origin it is a fresh defensive value copy made after validation and
must never alias the caller's pointer or other mutable intermediate.

The frozen packet order is: O2 owns the ABI in
`backend/internal/platform/events/envelope.go` and the receipt seam in
`backend/internal/platform/outbox/publisher.go`; O4 owns only the P4
ApplyRequest branch in
`backend/internal/calendar/event_ent_repository_inbound_invite.go`; O6 owns
the exact-row decode/proof in
`backend/internal/qamailcalendar/controller_delete.go`. R2 depends on these
packets only after P1a/O0 and O1 admission; no tombstone/cancelled branch is a
substitute dependency.
