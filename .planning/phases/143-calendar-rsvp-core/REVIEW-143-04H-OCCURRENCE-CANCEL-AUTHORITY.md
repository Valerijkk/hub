# Calendar 143-04H occurrence cancellation authority candidate

> **Authority acceptance-sync status:** `AUTHORITY_ACCEPTANCE_SYNC_CANDIDATE /
> ATOMIC_EXACT15_COMMIT_REQUIRED / IMPLEMENTATION_BLOCKED / RUNTIME_NOT_RUN`.
> This dirty exact15 sync cannot self-activate and grants no authority before one atomic exact15
> commit plus immutable acceptance review. The candidate below is immutable
> `f2966bdd50631feaa210a42a232880b21ade7e3a` (parent
> `03674eace728642fc2c9b39d2e8e2c81bf016033`, tree
> `45411b0692c79677b34af485df0ae2f23f4957d9`, subject
> `docs(calendar): propose occurrence cancel authority`). Immutable Contract Terra/high, Security
> Terra/xhigh and Lifecycle Terra/high reviews are `EXACT_SHA_PASS C/I/M 0/0/0`;
> dispatcher-enforced/no fallback, telemetry `UNKNOWN`. Only after the atomic sync commit and its
> immutable acceptance-commit review PASS does the candidate become
> `SOURCE_AUTHORITY_ACCEPTED / EXACT_SHA_PASS`; only then are its lower candidate/precommit/
> exact-review-next routes `CONSUMED/HISTORICAL` and root may dispatch the ordered exact3.

Candidate exact15 blobs reviewed at `f2966bdd50631feaa210a42a232880b21ade7e3a`:

```text
3da40f87415ca4558e38a6647a302222dc57851d  .planning/STATE.md
b191260174868cd807480167c157a6304f12c80e  .planning/decisions/specs/MAIL-SPEC.md
20362241aed9337355a12cecc1c4390977f6661c  .planning/phases/143-calendar-rsvp-core/.continue-here.md
7ca4ad7511312884c9c1718585be5bcd7d9aa5b0  .planning/phases/143-calendar-rsvp-core/DISPATCH-143-04H.md
bcb5c5a24af6812cef91a93979f3efb1db397a14  .planning/phases/143-calendar-rsvp-core/DISPATCH-INDEX.md
d26d01fd34ada0aea8f1f8bb22b069d06ccaf12e  .planning/phases/143-calendar-rsvp-core/DISPATCH-REMAINING.md
e5bfe182a5838c7d4f7302a37fdf0bfa43740396  .planning/phases/143-calendar-rsvp-core/PLAN.md
00e7460b4c4aeccc590c130351d2d0cc80e35cd4  .planning/phases/143-calendar-rsvp-core/PREIMPLEMENTATION-READINESS.md
3452a1443245b9d6baf2f21e7aef1cd91a0916ad  .planning/phases/143-calendar-rsvp-core/REVIEW-143-04H-OCCURRENCE-CANCEL-AUTHORITY.md
6a65653666023b2ba3f1c59c8f8b55867e74c368  .planning/phases/143-calendar-rsvp-core/RUNTIME-BLOCKER.md
1387ac7c06ca85e54ec2104ca3337c748131d44b  docs/mail/CONTRACT-CHANGELOG.md
f0f18c8beaf0ed6e71c71b85f5fa7c3698166bd0  docs/mail/PERSISTENCE-CONTRACT-0.1.8.md
e8f60069b40d4118f3abb082c97328cce92fcf69  docs/mail/QA-HANDOFF.md
beaa45987d03bf99d3c3ca0a173b139febc9b2f5  docs/mail/events-v1.schema.json
07e5ddb25ba8021405104408909dbe70b6c6a858  docs/mail/openapi.yaml
```

Acceptance-sync author provenance is `gpt-5.6-terra` / `xhigh`, dispatcher-enforced, no fallback,
telemetry `UNKNOWN`. Source remains `UNIMPLEMENTED / UNCOMMITTED / UNACCEPTED` pending its own
test-first reviews/commit/immutable receipt. 04F-H, Ent, adapter, handler/wiring, generation,
migration, API/UI, raw iTIP/transport, AQ blockers, launcher `NOT_READY` and runtime `NOT_RUN`
remain unchanged.

> **Predecessor candidate status:** `ROOT_SOURCE_AUTHORITY_CANDIDATE / PRECOMMIT_REVIEW_REQUIRED /
> ATOMIC_EXACT15_COMMIT_REQUIRED / IMPLEMENTATION_BLOCKED / RUNTIME_NOT_RUN`.
> The frozen candidate bytes below are immutable provenance and become `CONSUMED/HISTORICAL` only
> after acceptance-sync activation. This document freezes a future source contract only;
> it does not accept source, schema implementation, persistence adapter, handler, wiring,
> generation, migration, API implementation or runtime behavior.

## 1. Lifecycle and exact manifests

The ordered future source packet is exactly:

1. NEW `backend/tests/calendar/event_service_occurrence_cancel_test.go`
2. MODIFY `backend/internal/calendar/event_repository.go`
3. NEW `backend/internal/calendar/event_service_occurrence_cancel.go`

The candidate carrier is exactly these ordered 15 paths:

1. `.planning/STATE.md`
2. `.planning/decisions/specs/MAIL-SPEC.md`
3. `.planning/phases/143-calendar-rsvp-core/.continue-here.md`
4. `.planning/phases/143-calendar-rsvp-core/DISPATCH-143-04H.md`
5. `.planning/phases/143-calendar-rsvp-core/DISPATCH-INDEX.md`
6. `.planning/phases/143-calendar-rsvp-core/DISPATCH-REMAINING.md`
7. `.planning/phases/143-calendar-rsvp-core/PLAN.md`
8. `.planning/phases/143-calendar-rsvp-core/PREIMPLEMENTATION-READINESS.md`
9. `.planning/phases/143-calendar-rsvp-core/REVIEW-143-04H-OCCURRENCE-CANCEL-AUTHORITY.md`
10. `.planning/phases/143-calendar-rsvp-core/RUNTIME-BLOCKER.md`
11. `docs/mail/CONTRACT-CHANGELOG.md`
12. `docs/mail/PERSISTENCE-CONTRACT-0.1.8.md`
13. `docs/mail/QA-HANDOFF.md`
14. `docs/mail/events-v1.schema.json`
15. `docs/mail/openapi.yaml`

Activation requires, in this exact order: Contract/Spec/Oracle, Security/Adversarial and
Lifecycle/Dependency/Boundary precommit PASS; one atomic exact15 candidate commit; three immutable
exact-SHA PASS reviews; a separate exact15 docs-only acceptance sync; and immutable review of that
acceptance commit. Only then may root dispatch the source exact3. Candidate, precommit, source-first
or root-commit wording cannot self-activate. Source is `UNIMPLEMENTED / UNCOMMITTED / UNACCEPTED`.

Author provenance: `gpt-5.6-terra` / `xhigh`, dispatcher-enforced, no fallback, telemetry
`UNKNOWN`. Required reviewer provenance: Contract/Spec/Oracle Terra/high;
Security/Adversarial Terra/xhigh; Lifecycle/Dependency/Boundary Terra/high.

## 2. Repository ABI

The following declarations are exact additions in `event_repository.go`; field order and exported
signatures are frozen:

```go
type CalendarEventOccurrenceCancelAuthorizationRequest struct {
	EventID      uuid.UUID
	RecurrenceID string
}

type CalendarEventOccurrenceCancelAttendeeSnapshot struct {
	AttendeeID uuid.UUID
	EventID    uuid.UUID
	UserID     *uuid.UUID
	Status     RSVPStatus
	Version    uint
	RemovedAt  *time.Time
}

type CalendarEventOccurrenceCancelAuthorization struct {
	Current   CalendarEventCommandInput
	Attendees []CalendarEventOccurrenceCancelAttendeeSnapshot
	Exception *CalendarEventExceptionCommandInput
}

type CalendarEventOccurrenceCancelTransactionBackend interface {
	ResolveCalendarEventOccurrenceCancel(context.Context, EventMutationScope, CalendarEventOccurrenceCancelAuthorizationRequest) (CalendarEventOccurrenceCancelAuthorization, error)
	PersistCalendarEventOccurrenceCancel(context.Context, EventMutationScope, CalendarEventOccurrenceCancelPersistenceCommand) (CalendarEventOccurrenceCancelPersistenceResult, error)
}

func (*EventTransaction) ResolveCalendarEventOccurrenceCancelScope(context.Context, uuid.UUID) (EventMutationScope, error)
func (*EventTransaction) ResolveCalendarEventOccurrenceCancel(context.Context, EventMutationScope, CalendarEventOccurrenceCancelAuthorizationRequest) (CalendarEventOccurrenceCancelAuthorization, error)
func (*EventTransaction) PersistCalendarEventOccurrenceCancel(context.Context, EventMutationScope, CalendarEventOccurrenceCancelPersistenceCommand) (CalendarEventOccurrenceCancelPersistenceResult, error)
```

The request, attendee snapshot and authorization structs are plain boundary DTOs with no added
constructors or getters. Their values are validated and deep-copied only at the transaction seam.

Add exactly:

```go
const EventMutationCalendarOccurrenceCancel EventMutationOperation = "event_occurrence_cancel"
```

It is receipt-free: `EventMutationScope.Operation()` is empty and
`EventMutationScope.EventOperation()` is `EventMutationCalendarOccurrenceCancel`. Do not add a
`MutationReceiptOperation`. `validEventMutationScope` accepts only the recognized receipt-free
update and occurrence-cancel operations in addition to existing receipt mappings.

`ResolveCalendarEventOccurrenceCancelScope` asserts the optional backend capability before
`ResolveActorOrganization` or any backend method. Unsupported capability returns
`ErrEventInvalid`; the enclosing `WithEventTransaction` callback is the only allowed earlier
repository seam. Repository input/output adapters enforce caps before allocation/traversal and
deep-copy all pointer/slice values.

The accepted immutable commands gain exact deep-copy getters:

```go
func (CalendarEventCommand) Snapshot() CalendarEventCommandInput
func (CalendarEventExceptionCommand) Snapshot() CalendarEventExceptionCommandInput
```

## 3. Service ABI

The new service file freezes these declarations and signatures:

```go
type CalendarEventOccurrenceCancelInput struct {
	EventID         uuid.UUID
	RecurrenceID    string
	ExpectedVersion uint
}

type CalendarEventOccurrenceCancelCommand struct { input CalendarEventOccurrenceCancelInput }
func NewCalendarEventOccurrenceCancelCommand(CalendarEventOccurrenceCancelInput) (CalendarEventOccurrenceCancelCommand, error)
func (CalendarEventOccurrenceCancelCommand) EventID() uuid.UUID
func (CalendarEventOccurrenceCancelCommand) RecurrenceID() string
func (CalendarEventOccurrenceCancelCommand) ExpectedVersion() uint
func (CalendarEventOccurrenceCancelCommand) valid() bool

type CalendarEventOccurrenceCancelPersistenceCommand struct {
	event           CalendarEventCommand
	expectedVersion uint
	exception       CalendarEventExceptionCommand
	outbox          EventOutboxDescriptor
}
func NewCalendarEventOccurrenceCancelPersistenceCommand(CalendarEventCommand, uint, CalendarEventExceptionCommand, EventOutboxDescriptor) (CalendarEventOccurrenceCancelPersistenceCommand, error)
func (CalendarEventOccurrenceCancelPersistenceCommand) Event() CalendarEventCommand
func (CalendarEventOccurrenceCancelPersistenceCommand) ExpectedVersion() uint
func (CalendarEventOccurrenceCancelPersistenceCommand) Exception() CalendarEventExceptionCommand
func (CalendarEventOccurrenceCancelPersistenceCommand) Outbox() EventOutboxDescriptor
func (CalendarEventOccurrenceCancelPersistenceCommand) valid() bool

type CalendarEventOccurrenceCancelPersistenceResult struct {
	event     CalendarEventCommand
	exception CalendarEventExceptionCommand
	outbox    EventOutboxDescriptor
}
func NewCalendarEventOccurrenceCancelPersistenceResult(CalendarEventCommand, CalendarEventExceptionCommand, EventOutboxDescriptor) (CalendarEventOccurrenceCancelPersistenceResult, error)
func (CalendarEventOccurrenceCancelPersistenceResult) Event() CalendarEventCommand
func (CalendarEventOccurrenceCancelPersistenceResult) Exception() CalendarEventExceptionCommand
func (CalendarEventOccurrenceCancelPersistenceResult) Outbox() EventOutboxDescriptor
func (CalendarEventOccurrenceCancelPersistenceResult) valid() bool

type CalendarEventOccurrenceCancelResult struct {
	eventID      uuid.UUID
	recurrenceID string
	status       CalendarEventStatus
	sequence     uint
	eventVersion uint
}
func NewCalendarEventOccurrenceCancelResult(uuid.UUID, string, CalendarEventStatus, uint, uint) (CalendarEventOccurrenceCancelResult, error)
func (CalendarEventOccurrenceCancelResult) EventID() uuid.UUID
func (CalendarEventOccurrenceCancelResult) RecurrenceID() string
func (CalendarEventOccurrenceCancelResult) Status() CalendarEventStatus
func (CalendarEventOccurrenceCancelResult) Sequence() uint
func (CalendarEventOccurrenceCancelResult) EventVersion() uint

func (*EventService) CancelCalendarEventOccurrence(context.Context, uuid.UUID, CalendarEventOccurrenceCancelCommand) (CalendarEventOccurrenceCancelResult, error)
```

All constructors validate, copy and reject zero/invalid state with `ErrEventInvalid`; getters never
expose aliases.

`NewCalendarEventOccurrenceCancelCommand` performs recurrence ingress preflight before any
repository, clock or ID call: byte length must be 1..255 before UTF-8, control-character or
canonical-form scanning; invalid UTF-8 and every control byte/rune are rejected; the accepted
string is preserved byte-for-byte. Temporal-kind canonicalization and nominal membership are
deliberately deferred until after the authoritative master is identity-bound.

## 4. Conflict ABI and safe error behavior

```go
type CalendarEventOccurrenceCancelConflictKind string

const (
	CalendarEventOccurrenceCancelVersionConflict CalendarEventOccurrenceCancelConflictKind = "version_conflict"
	CalendarEventOccurrenceCancelNotMutable CalendarEventOccurrenceCancelConflictKind = "not_mutable"
)

type CalendarEventOccurrenceCancelConflictError struct {
	kind           CalendarEventOccurrenceCancelConflictKind
	eventID        uuid.UUID
	currentVersion uint
}

func (CalendarEventOccurrenceCancelConflictError) Error() string
func (CalendarEventOccurrenceCancelConflictError) Unwrap() error
func (CalendarEventOccurrenceCancelConflictError) Kind() CalendarEventOccurrenceCancelConflictKind
func (CalendarEventOccurrenceCancelConflictError) EventID() (uuid.UUID, bool)
func (CalendarEventOccurrenceCancelConflictError) CurrentVersion() (uint, bool)
```

`Error()` is exactly `"calendar event occurrence cancel conflict"`; `Unwrap()` is
`ErrEventConflict`. Internal constructors are
`calendarEventOccurrenceCancelVersionConflict(eventID, currentVersion)` and
`calendarEventOccurrenceCancelNotMutable()`. Only a structurally valid version conflict exposes
event id/version; `not_mutable` exposes neither. Missing, cross-organization, wrong owner and wrong
organizer all map to the same content-free `ErrEventNotFound`. Unknown/raw/backend errors and
malformed typed values map to content-free `ErrEventInvalid`.

## 5. Safe occurrence-cancel envelope

```go
type CalendarInviteOccurrenceCancelledPayload struct {
	EventID      uuid.UUID   `json:"event_id"`
	UID          string      `json:"uid"`
	Sequence     uint        `json:"sequence"`
	RecurrenceID string      `json:"recurrence_id"`
	AttendeeIDs  []uuid.UUID `json:"attendee_ids"`
}

func NewCalendarInviteOccurrenceCancelledEnvelopeV1(
	eventID uuid.UUID,
	occurredAt time.Time,
	organizationID uuid.UUID,
	actorID *uuid.UUID,
	correlationID string,
	payload CalendarInviteOccurrenceCancelledPayload,
) (events.EnvelopeV1, error)
```

The builder emits subject `calendar.invite.cancelled.v1`, schema version 1, exact correlation
`payload.EventID.String()`, UID `hub-<event-id>`, nonzero actor/event/org ids and nonzero UTC time.
Sequence is positive. `recurrence_id` is already canonical, non-null and 1..255 bytes. Attendee ids
are a non-nil JSON array of at most 200 nonzero UUIDs, strictly bytewise sorted and unique. It
contains no title, description, email, MIME, iTIP, provider, delivery or transport data. All
invalid input maps to `ErrEventInvalid`.

`events-v1.schema.json` preserves the existing master-cancel shape (recurrence id absent or null)
and adds a disjoint occurrence-cancel payload variant requiring a non-null canonical all-day or
UTC-second recurrence id under the same subject. The two variants are selected with `oneOf`.

## 6. Deterministic authorization and mutation order

1. Reject nil service/repository/clock/ids, zero actor and invalid command with zero calls.
2. Enter exactly one repository transaction.
3. Assert optional occurrence-cancel backend capability before any transaction-backend call, then
   resolve actor organization.
4. Perform one owner-and-organization-scoped authorization resolve. No broad event lookup is
   allowed.
5. Before validating, capping, cloning or traversing any untrusted current string, slice, attendee
   or exception, bind only fixed-shape identity: exact EventID and organization plus
   `OwnerID == actor` and non-nil `OrganizerID == actor`. Any missing/cross-org/wrong-owner/
   wrong-organizer identity returns the same `ErrEventNotFound`, even when all remaining foreign
   snapshot content is malformed or oversized.
6. Only after identity binding, preflight `Current.Exdates <= 1000` and `Attendees <= 200` before
   cloning/traversal, then validate the complete current event, attendee and optional exception
   snapshots, including Hub source, canonical `hub-<event-id>` UID and normalized organizer email.
7. Before expected-version comparison require active confirmed/tentative status, non-empty valid
   normalized RRULE and canonical nominal membership. EXDATE makes the nominal member
   `not_mutable` but never proves membership.
8. Canonical recurrence ids are `YYYY-MM-DD` for all-day and
   `YYYY-MM-DDTHH:MM:SSZ` for timed occurrences. Reuse accepted 04B recurrence semantics and the
   1000 occurrence/work guard; skip nonexistent DST civil times and retain the accepted ambiguous
   time resolution. Malformed/unnormalized/non-member/non-recurring/external/cancelled instances
   fail closed as `not_mutable`.
9. Compare `ExpectedVersion` only after identity/mutability masking. A mismatch returns the typed
   version conflict with zero clock/ID/persist/outbox work.
10. Validate an existing exception's event id, canonical recurrence id, sequence/status/cancelled
    tuple and absence/presence of temporal override. Corrupt or future-sequence state is invalid.
11. Existing cancelled exception at the matching current expected version returns an idempotent
    result from the master with zero clock/ID/persist/outbox. Stale expected version still conflicts.
12. Check master `Version+1` and `Sequence+1` overflow before clock or ID calls.
13. Reuse a live exception id; otherwise generate one exception id exactly once. Generate one
    application envelope id exactly once. Zero ids or collisions fail.
14. Deep-clone the master and change only version and sequence by one. Preserve status, RRULE,
    EXDATE, temporal/content/identity/link/reminder fields byte-for-byte.
15. Construct one exception for the exact event/recurrence with nil temporal override, the new
    sequence, cancelled status and `Cancelled=true`.
16. Project payload attendee ids from active attendees (`RemovedAt == nil`), excluding the
    organizer's matching UserID while retaining external attendees; sort and de-duplicate.
17. Read clock once, build exactly one safe EnvelopeV1 and issue one atomic persistence CAS using
    the original expected version.
18. Require the persistence result to echo every master field, every exception field including
    nil distinctions, and descriptor subject/event id/exact frozen EnvelopeV1 JSON. Any mismatch is
    `ErrEventInvalid` and rolls back.
19. Return only event id, recurrence id, cancelled status, sequence and event version.

## 7. Mandatory non-tautological external-package oracles

- Compile/reflection-freeze all exact exported fields, order, types, methods, constructors and
  interfaces; reject extra receipt operation and prove the new event operation string.
- Unsupported/typed-nil occurrence port: exactly one outer transaction seam and zero backend
  calls. Prove capability precedes actor organization resolution.
- Scope has empty receipt operation and exact occurrence event operation; forged, update-operation,
  escaped and closed transaction scopes fail without backend access.
- Missing/cross-org/wrong owner/wrong organizer produce identical not-found bytes and never expose
  version/location. Hostile foreign snapshots with malformed/oversized strings, EXDATE/attendee
  slices, pointer aliases and corrupt exceptions must still mask as identical not-found before
  validation, caps, traversal or allocation; source/status/RRULE/member/EXDATE failures on an
  identity-bound snapshot precede version comparison.
- Command ingress covers exact 1/255-byte acceptance, zero/256-byte rejection, invalid UTF-8 and
  every control category. A 1 MiB recurrence string returns `ErrEventInvalid` with zero repository,
  backend, clock and ID calls and no panic; accepted recurrence bytes are unchanged until
  authoritative temporal canonicalization.
- Cover all-day/timed recurrence ids, DAILY/WEEKLY/MONTHLY, INTERVAL, BYDAY/BYMONTHDAY,
  COUNT/UNTIL boundaries, DTSTART-first, UTC-second rejection of fractional input, DST gap/fold,
  non-member and 1000/1001 bounds. Prove EXDATE is not membership proof.
- Existing-cancelled/current-version is idempotent with zero clock/IDs/persist/outbox; stale version
  conflicts first. Corrupt exception state fails closed.
- Happy path checks call order, exact one or zero exception-ID calls, exact one envelope-ID call,
  checked counter increments, one persistence call and no mutation of the authorization fixture.
- Payload tests cover organizer exclusion, active internal/external attendees, removed exclusion,
  sorted uniqueness, cap rejection and empty `[]` rather than `null`.
- Constructors, snapshot getters, repository clones and returned results resist hostile mutation of
  every slice/pointer input and output.
- Table-tamper every master field, exception field/nil distinction and envelope subject/event id/
  correlation/payload JSON; each mismatch must be content-free invalid and roll back.
- Prove counter overflow, zero/colliding IDs, zero/non-UTC clock, raw errors and malformed typed
  errors cause zero committed writes. Prove rollback for resolve, build, persist and echo failures.
- The unit asserts exact safe envelope JSON/schema binding and behavioral zero
  `EventSerializer` calls. Static Security/Lifecycle review, not the external unit, owns the
  source-token and no-I/O scan for raw `METHOD:CANCEL`, MIME, iTIP, provider, delivery and
  transport dependencies.

## 8. Dependency and non-authority boundary

Hard static dependencies are accepted 143-04A descriptor/event ABI, 143-04B recurrence policy,
143-04C update/cancel transaction pattern, 143-04G immutable envelope and 143-04L subject/schema.
Accepted 143-04J/04K remain independent and are not invoked. The old 143-02A serializer,
`METHOD:CANCEL`, `EventSerializer` and source-first clauses are `CONSUMED/HISTORICAL` for 04H.

143-04F-H alone later owns Ent locks/CAS/exception/outbox persistence. Ent, adapter, handler,
wiring, generation, migration, transport, API implementation, UI and runtime are not authorized.
Launcher stays `NOT_READY`; all executable evidence is `NOT_RUN`; AQ production migration and
recurring-call backfill blockers remain open.
