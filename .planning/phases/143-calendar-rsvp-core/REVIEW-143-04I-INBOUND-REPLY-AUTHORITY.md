# 143-04I — inbound iTIP REPLY authority

## Status and lifecycle

`SOURCE_RECEIPT_SYNC_CANDIDATE / ATOMIC_EXACT10_COMMIT_REQUIRED / IMMUTABLE_REVIEW_REQUIRED / RUNTIME_NOT_RUN`.

This non-authoritative dirty exact10 docs receipt records immutable source `46f4d6024e6c5475cdd943af9979c4ffadc5474a`, sole parent `04dbde1a14e9499462dbc1ea8784583f83f0b40f`, tree `e3ce6097822f360a04717d0dce61597810ca2046`, subject `feat(calendar): add inbound reply service`. Source Contract/Security/Lifecycle precommit and immutable `EXACT_SHA_PASS` are C/I/M `0/0/0`; provenance is Contract Terra/high, Security Terra/xhigh, Lifecycle Terra/high, dispatcher-enforced/no fallback/telemetry `UNKNOWN`.

The immutable source exact3, in required authored-first receipt order, is:

| Order | Path | Blob |
|---:|---|---|
| 1 | `backend/tests/calendar/event_service_inbound_reply_test.go` | `e0c4e64d9d704ebde24c87d9569590cde69d9115` |
| 2 | `backend/internal/calendar/event_repository.go` | `c5fc1df5bf0a690db199b9df922c5463d27ab2a1` |
| 3 | `backend/internal/calendar/event_service_inbound_reply.go` | `81814a334c17df9dbd24da58a6fb95a776924670` |

`IMPLEMENTATION-143-04I.md` is the canonical explicit ordered exact10 docs-receipt manifest. This dirty receipt grants no current `STATIC_ACCEPTED` status, executable or runtime authority. Only one atomic commit whose diff is set-equal to that exact10 manifest, followed by immutable receipt Contract/Security/Lifecycle triple PASS, can activate it. Only then source is `STATIC_ACCEPTED / EXACT_SHA_PASS` and lower 04I authority/source-unimplemented routes become `CONSUMED/HISTORICAL`. Runtime remains `NOT_RUN`; launcher/testframework external handoff, 04F-I, 03E, 04F-Base, Ent/generated source, generation/migration, HTTP/API/OpenAPI/UI and NATS/iTIP transport remain blocked/not-ready; no next Calendar packet is selected or authorized.

## Exact source manifest

Authored and committed in this order only:

1. NEW `backend/tests/calendar/event_service_inbound_reply_test.go`
2. MODIFY `backend/internal/calendar/event_repository.go`
3. NEW `backend/internal/calendar/event_service_inbound_reply.go`

No `itip_*`, Ent schema/generated source, Mail, HTTP/API, serializer, NATS client, testframework or existing RSVP source edits.

## Frozen ABI and inbound scope

```go
type CalendarInboundReplyInput struct { OwnerID uuid.UUID; Sender string; Message ITIPMessage }
type CalendarInboundReplyOutcome string
const (
    CalendarInboundReplyApplied CalendarInboundReplyOutcome = "applied"
    CalendarInboundReplyObservedNoChange CalendarInboundReplyOutcome = "observed_no_change"
    CalendarInboundReplyDuplicate CalendarInboundReplyOutcome = "duplicate"
    CalendarInboundReplyStale CalendarInboundReplyOutcome = "stale"
    CalendarInboundReplyQuarantined CalendarInboundReplyOutcome = "quarantined"
)
type CalendarInboundReplyResult struct { /* opaque */ }
func (result CalendarInboundReplyResult) Outcome() CalendarInboundReplyOutcome
func (service *EventService) ApplyInboundITIPReply(ctx context.Context, input CalendarInboundReplyInput) (CalendarInboundReplyResult, error)

EventMutationInboundReply EventMutationOperation = "inbound_reply"
type CalendarInboundReplyScope struct { /* ownerID, organizationID, transaction */ }
func (scope CalendarInboundReplyScope) OwnerID() uuid.UUID
func (scope CalendarInboundReplyScope) OrganizationID() uuid.UUID
func (scope CalendarInboundReplyScope) EventOperation() EventMutationOperation
type CalendarInboundReplyTransactionBackend interface {
    ResolveCalendarInboundReplyOwner(context.Context, uuid.UUID) (uuid.UUID, bool, error)
    ResolveCalendarInboundReply(context.Context, CalendarInboundReplyScope, CalendarInboundReplyAuthorizationRequest) (CalendarInboundReplyAuthorization, error)
    PersistCalendarInboundReply(context.Context, CalendarInboundReplyScope, CalendarInboundReplyPersistenceCommand) (CalendarInboundReplyPersistenceResult, error)
}
func (transaction *EventTransaction) ResolveCalendarInboundReplyScope(context.Context, uuid.UUID) (CalendarInboundReplyScope, bool, error)
func (transaction *EventTransaction) ResolveCalendarInboundReply(context.Context, CalendarInboundReplyScope, CalendarInboundReplyAuthorizationRequest) (CalendarInboundReplyAuthorization, error)
func (transaction *EventTransaction) PersistCalendarInboundReply(context.Context, CalendarInboundReplyScope, CalendarInboundReplyPersistenceCommand) (CalendarInboundReplyPersistenceResult, error)
```

Inbound mail is not an authenticated Hub mutation actor. Do not extend `MutationReceiptOperation`; do not call `ResolveMutationScope`, `ResolveActorOrganization`, `ReserveMutationReceipt`, `CompleteMutationReceipt` or `AppendOutbox`; do not place `OwnerID` or sender into EnvelopeV1 actor. The optional capability owns owner-to-organization resolution. Capability/typed-nil proof precedes every backend call. Owner miss is `(zero,false,nil)` and produces quarantined; success is `(nonzero organization,true,nil)`; every other shape is invalid. A valid scope always reports `EventMutationInboundReply`, and is valid only for the same live transaction; captured/foreign/closed use is rejected. The base `EventRepository` and `EventTransactionBackend` interfaces remain unchanged.

## Frozen records and invariants

```go
type CalendarInboundReplyReceiptOutcome string
const (
    CalendarInboundReplyReceiptApplied CalendarInboundReplyReceiptOutcome = "applied"
    CalendarInboundReplyReceiptObservedNoChange CalendarInboundReplyReceiptOutcome = "observed_no_change"
)
type CalendarInboundReplyReceiptKey struct { /* opaque version + sender-bound SHA-256 */ }
func NewCalendarInboundReplyReceiptKey(uint8, string, [sha256.Size]byte) (CalendarInboundReplyReceiptKey, error)
func (key CalendarInboundReplyReceiptKey) CanonicalVersion() uint8
func (key CalendarInboundReplyReceiptKey) PayloadHash() [sha256.Size]byte
type CalendarInboundReplyAuthorizationRequest struct { UID string; HasRecurrenceID bool; RecurrenceID string; Sender string; EventSequence uint; Receipt CalendarInboundReplyReceiptKey }
type CalendarInboundReplyAttendeeSnapshot struct { AttendeeID uuid.UUID; EventID uuid.UUID; Email string; Status RSVPStatus; RecurrenceID *string; RespondedAt *time.Time; Version uint; RemovedAt *time.Time }
type CalendarInboundReplyAuthorization struct { Matched bool; Current CalendarEventCommandInput; Target CalendarInboundReplyAttendeeSnapshot; Exception *CalendarEventExceptionCommandInput; ExistingReceipt bool; DistinctReceiptCount uint8 }
```

`Matched=false` requires every other authorization field zero. `Matched=true` binds server-resolved owner + organization + exact UID + exact null/non-null recurrence + normalized sender to one active target attendee. Fixed UUID/owner/org/UID/recurrence/email identity is checked before any variable-sized clone. `Current.Exdates` is capped at 1,000 before clone. `DistinctReceiptCount` is 0..8 for the exact target attendee and effective event sequence while locked. `ExistingReceipt` is the exact version+hash for that target and sequence. Occurrence REPLY requires an already materialized active occurrence attendee with the exact canonical recurrence; absent is quarantine/zero-write. Never attach an occurrence receipt to the series attendee or infer materialization from HTTP 04E.

```go
type CalendarInboundReplyPersistenceCommand struct { /* opaque */ }
func NewCalendarInboundReplyPersistenceCommand(prior CalendarInboundReplyAttendeeSnapshot, updated *CalendarInboundReplyAttendeeSnapshot, eventSequence uint, receipt CalendarInboundReplyReceiptKey, outcome CalendarInboundReplyReceiptOutcome, outbox *EventOutboxDescriptor) (CalendarInboundReplyPersistenceCommand, error)
func (command CalendarInboundReplyPersistenceCommand) Prior() CalendarInboundReplyAttendeeSnapshot
func (command CalendarInboundReplyPersistenceCommand) Updated() (CalendarInboundReplyAttendeeSnapshot, bool)
func (command CalendarInboundReplyPersistenceCommand) ExpectedVersion() uint
func (command CalendarInboundReplyPersistenceCommand) EventSequence() uint
func (command CalendarInboundReplyPersistenceCommand) Receipt() CalendarInboundReplyReceiptKey
func (command CalendarInboundReplyPersistenceCommand) ReceiptOutcome() CalendarInboundReplyReceiptOutcome
func (command CalendarInboundReplyPersistenceCommand) Outbox() (EventOutboxDescriptor, bool)
type CalendarInboundReplyPersistenceResult struct { /* opaque */ }
func NewCalendarInboundReplyPersistenceResult(attendee CalendarInboundReplyAttendeeSnapshot, eventSequence uint, receipt CalendarInboundReplyReceiptKey, outcome CalendarInboundReplyReceiptOutcome, outbox *EventOutboxDescriptor) (CalendarInboundReplyPersistenceResult, error)
func (result CalendarInboundReplyPersistenceResult) Attendee() CalendarInboundReplyAttendeeSnapshot
func (result CalendarInboundReplyPersistenceResult) EventSequence() uint
func (result CalendarInboundReplyPersistenceResult) Receipt() CalendarInboundReplyReceiptKey
func (result CalendarInboundReplyPersistenceResult) ReceiptOutcome() CalendarInboundReplyReceiptOutcome
func (result CalendarInboundReplyPersistenceResult) Outbox() (EventOutboxDescriptor, bool)
```

Observed-no-change requires `updated=nil`, `outbox=nil`, and result attendee exactly equals prior. Applied requires immutable attendee identity equality, different valid PARTSTAT, checked version+1, nonnil valid whole-second UTC `RespondedAt` and exactly one valid outbox; result exactly equals updated. ExpectedVersion is prior.Version. Every result field echoes the command or the transaction returns `ErrEventInvalid` and rolls back.

## Frozen canonical ingress and state machine

1. Reject nil context/service, typed-nil repository/clock/IDs, zero OwnerID as `ErrEventInvalid` before any call.
2. Preflight Sender at 320 bytes before UTF-8/address scans, normalize through the accepted iTIP email normalizer, then compute `CanonicalITIPIdentity(Message)`. The only accepted message is canonical version 1, method REPLY, exactly one attendee, attendee status accepted/declined/tentative, and normalized sole attendee email exactly equal to normalized authenticated-envelope sender. Any semantic/method/sender failure returns quarantined with zero transaction/clock/ID calls.
3. Derive exactly `SHA-256("hub.calendar.itip.reply" || 0x00 || U8(version) || U32BE(len(sender)) || sender || parserPayloadHash)`. No JSON/delimiter substitution, raw input or last-hash storage.
4. Resolve inbound owner scope, then resolve exact owner+UID+optional recurrence+sender+sequence+receipt under the optional port. Miss/foreign/spoof shape returns quarantined without disclosure.
5. Validate fixed identity. ExistingReceipt returns duplicate before mutable event/status/sequence comparisons, so A -> B -> redelivered A cannot roll back.
6. For new hashes, validate bounded authoritative event/target. Require active Hub event, owner is exact organizer, normalized parsed ORGANIZER equals authoritative organizer email, target active/nonremoved and exact email/key. Series targets nil recurrence. Occurrence targets exact nonnil recurrence, is a real non-EXDATE member, has no cancelled exception, and uses exception.Sequence when present; otherwise current.Sequence.
7. Incoming lower effective sequence returns stale; higher returns quarantined; both zero write. Equal proceeds.
8. Count 8 with a new hash returns quarantined. Counts above 8 or malformed backend shapes are `ErrEventInvalid`.
9. Same status persists exactly one observed_no_change receipt, with no clock/ID/attendee/outbox/event/call mutation.
10. Different status checks version overflow, samples clock once, generates one unique nonzero application event ID, increments only target attendee version, sets RespondedAt, and persists applied receipt + target CAS + one outbox. Series projection is future 04F-I adapter work; occurrence never changes linked-call participants.
11. Applied EnvelopeV1 subject is `calendar.rsvp.changed.v1`, organization comes from scope, actor is nil, correlation is event UUID, recurrence is nil for series or exact string for occurrence, reply_delivery_status is `not_required`, and no UID/hash/raw iTIP/MIME/sender-only metadata is added.
12. Persist once. Exact echo mismatch/error rolls back. No mutation changes event UID/content/sequence/organizer, creates CalendarMutationReceipt, serializes outbound REPLY, or calls transport.

## Required authored-first external-package oracles

- Reflection/signatures/opaque fields/constant and nil/typed-nil/closed/foreign scope matrix.
- Literal sender-bound SHA-256 vector; version, sender normalization, Unicode, length prefix and one-bit parser hash changes.
- Canonical REPLY only, sole attendee, no needs_action; hostile constructed ITIPMessage and 1 MiB sender are quarantine/no-panic/no backend.
- Owner/UID/org/sender/organizer/null-vs-nonnull recurrence masks before variable clone; cap-before-clone.
- Series and already-materialized occurrence; missing occurrence quarantine; cancelled/nonmember/EXDATE/cancelled-exception quarantine.
- Lower/equal/higher sequence, exception effective sequence, existing hash precedence.
- Declined A -> accepted B -> redelivered A for series and occurrence.
- New same-status observed receipt only; later distinct applied; redelivery duplicate.
- Eight distinct exact-target hashes; ninth zero-write quarantine; counts 9..255 invalid.
- Applied event exact JSON, nil actor, correlation/org, recurrence, one clock/ID/outbox; event/organizer fields unchanged.
- Version overflow; event-ID collisions; every result echo tamper; raw backend error containment; transaction rollback.
- Deep-copy input Message attendee/temporal/EXDATE pointers and every authorization/command/result pointer/slice/descriptor.
- No CalendarMutationReceipt, AppendOutbox fallback, serializer, raw iTIP, HTTP, Mail, Ent, DB, NATS, filesystem, network or goroutine in source.

## Downstream dependency

`143-04F-I` remains `BLOCKED / NOT_READY / NOT_RUN` and depends exactly on `143-03F + 143-04F-Base + accepted 143-04I`. `143-03D` is the forbidden HTTP receipt dependency and is not sufficient.
