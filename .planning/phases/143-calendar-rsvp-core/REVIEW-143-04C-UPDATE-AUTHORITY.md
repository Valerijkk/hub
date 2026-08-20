# 143-04C — Calendar update/master-cancel source authority candidate

## Combined exact17 conditional acceptance sync (2026-08-20)

Status: `COMBINED_EXACT17_ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_COMMIT_REQUIRED / IMMUTABLE_EXACT_SHA_REVIEW_REQUIRED`. This working tree is not authority. Calendar exact12 `e60ff29934b8e53c34549cebae2e10657bae1540` (parent `e21a98a7459cd68690912c33ff031e286ba92ef1`, tree `5df5e079bdfe9685957e065e534c352cede25d2a`) and Mail correction exact9 `4f92ee6b9057efccd10c0b7ac7cf7dfd0f220c71` (parent `e60ff29934b8e53c34549cebae2e10657bae1540`, tree `2c65309e71c3501cd55516bc84c88c1c605df5a3`) each have Contract/Security/Lifecycle precommit and immutable exact-SHA PASS C/I/M `0/0/0`. This receipt becomes effective only after root creates one atomic commit changing exactly the 17 named acceptance-sync documents and Contract/Security/Lifecycle independently return immutable exact-SHA PASS for that exact commit. Until then all other Calendar candidate and Mail correction/suspension clauses in this file remain current; only after those conditions they automatically become `CONSUMED/HISTORICAL`, and the accepted routes below activate without another docs sync.

On activation, Calendar `e60ff299` becomes `SOURCE_AUTHORITY_ACCEPTED / EXACT_SHA_PASS` and root may dispatch only 143-03B-T exact2, test first: `backend/tests/calendar/calendar_attendee_schema_test.go`, then `backend/ent/schema/calendarattendee.go`. 143-04C exact3 remains `BLOCKED / NOT_AUTHORIZED` until 143-03B-T source itself is `STATIC_ACCEPTED`; there is no 04C source dispatch or acceptance now. Mail correction `4f92ee6b` becomes `SOURCE_AUTHORITY_ACCEPTED / EXACT_SHA_PASS` and root may dispatch only the frozen exact4, test first and in order: `backend/tests/mail/mailbox_events_test.go`, `backend/tests/mail/mailbox_sync_policy_test.go`, `backend/internal/mail/mailbox_events.go`, `backend/internal/mail/mailbox_sync_policy.go`. The current dirty exact4 candidates remain `UNCOMMITTED / UNACCEPTED` and require their own source static review, atomic commit and immutable exact-SHA acceptance lifecycle; runtime remains `NOT_RUN`.

The 143-04A/B/G receipts remain accepted. 143-03E, 144-02E, Ent generation, migration, launcher readiness and all executable/runtime evidence remain blocked, `NOT_READY` or `NOT_RUN` as already recorded. No source outside the two activated manifests, generation, migration, service expansion, API/UI, product, phase or release authority follows; raw iTIP remains prohibited.

Status: `ROOT_SOURCE_AUTHORITY_CANDIDATE` / `PRECOMMIT_REVIEW_REQUIRED`.

Clause-scoped precedence consumes the earlier "no next Calendar code packet selected" wording
only for the following route. All prior 04A/B/G static receipts, 03E Ent generation/migration
block, launcher `NOT_READY`, and executable/runtime `NOT_RUN` remain authoritative.

After triple independent PASS, one exact-12 documentation candidate commit, immutable child-SHA
triple PASS, and separate acceptance sync, root may dispatch only these sequential manifests:

1. 143-03B-T, tests first: `backend/tests/calendar/calendar_attendee_schema_test.go`, then
   `backend/ent/schema/calendarattendee.go`.
2. Only after 143-03B-T `STATIC_ACCEPTED`, 143-04C, tests first:
   `backend/tests/calendar/event_service_update_test.go`, then
   `backend/internal/calendar/event_repository.go`, then new
   `backend/internal/calendar/event_service_update.go`.

143-03B-T owns only the `CalendarAttendee.removed_at` representation, the two live partial-index
predicates (`removed_at IS NULL`), and its schema-metadata oracle. Inactive-reader filtering,
fresh re-invite identity, and normalized-email role/resolved-identity replacement rejection are
exclusively later 143-04C service/source/test obligations, after 143-03B-T is accepted.

## Frozen Go ABI and behavior

This is the complete source contract for the later exact-three packet; it does not itself accept
source. `EventTransactionBackend` remains unchanged. The ABI additions are:

```go
type EventMutationOperation string

const (
	EventMutationEventCreate      EventMutationOperation = "event_create"
	EventMutationLegacyCallCreate EventMutationOperation = "legacy_call_create"
	EventMutationRSVPSeries       EventMutationOperation = "rsvp_series"
	EventMutationRSVPOccurrence   EventMutationOperation = "rsvp_occurrence"
	EventMutationCalendarUpdate   EventMutationOperation = "event_update"
)

type EventMutationScope struct {
	actorID, organizationID uuid.UUID
	receiptOperation        MutationReceiptOperation
	eventOperation          EventMutationOperation
	transaction             *EventTransaction
}

func (t *EventTransaction) ResolveCalendarEventUpdateScope(context.Context, uuid.UUID) (EventMutationScope, error)
func (s EventMutationScope) Operation() MutationReceiptOperation
func (s EventMutationScope) EventOperation() EventMutationOperation

type CalendarEventUpdateAuthorizationRequest struct {
	EventID   uuid.UUID
	Attendees []CalendarEventCreateAttendeeCandidate // copied only when attendee patch is present
}
type CalendarEventUpdateTransactionBackend interface {
	ResolveCalendarEventUpdate(context.Context, EventMutationScope, CalendarEventUpdateAuthorizationRequest) (CalendarEventUpdateAuthorization, error)
	PersistCalendarEventUpdate(context.Context, EventMutationScope, CalendarEventUpdatePersistenceCommand) (CalendarEventUpdatePersistenceResult, error)
}

type CalendarEventUpdateConflictKind string
const (
	CalendarEventUpdateVersionConflict CalendarEventUpdateConflictKind = "version_conflict"
	CalendarEventUpdateNotMutable CalendarEventUpdateConflictKind = "not_mutable"
	CalendarEventUpdateRecurrenceExceptionConflict CalendarEventUpdateConflictKind = "recurrence_exception_conflict"
)
type CalendarEventUpdateConflictError struct {
	kind CalendarEventUpdateConflictKind
	eventID uuid.UUID
	currentVersion uint
}
func (e CalendarEventUpdateConflictError) Error() string
func (e CalendarEventUpdateConflictError) Unwrap() error
func (e CalendarEventUpdateConflictError) Kind() CalendarEventUpdateConflictKind
func (e CalendarEventUpdateConflictError) EventID() (uuid.UUID, bool)
func (e CalendarEventUpdateConflictError) CurrentVersion() (uint, bool)
```

Backward compatibility is explicit:
`EventMutationScope` retains `receiptOperation MutationReceiptOperation` (empty only for an
update scope) and adds `eventOperation EventMutationOperation`. `EventMutationOperation` mirrors
all four existing receipt values (`event_create`, `legacy_call_create`, `rsvp_series`,
`rsvp_occurrence`) and adds `EventMutationCalendarUpdate = "event_update"`. Existing
`ResolveMutationScope(ctx, actorID, MutationReceiptOperation)` keeps its signature and maps each
valid receipt operation to both fields. New exact
`ResolveCalendarEventUpdateScope(ctx, actorID)` first checks that the active transaction backend
implements `CalendarEventUpdateTransactionBackend`, before calling `ResolveActorOrganization` or
any other backend method. An absent capability returns `ErrEventInvalid` with zero backend calls.
Only then does it resolve the actor organization, sets empty
`receiptOperation` and `eventOperation=event_update`. Existing `Operation() MutationReceiptOperation`
remains; `EventOperation() EventMutationOperation` is added. `validateScope` validates the
operation mapping. `NewMutationReceiptReservation`, receipt reserve and receipt completion require
a nonempty valid receipt operation: an update scope cannot create or complete a receipt.

`EventTransaction` exposes guarded wrappers with the two optional-interface names above. Scope
creation and each later resolve/persist wrapper repeat the capability assertion under its active
transaction guard; an absent capability returns `ErrEventInvalid` with zero backend calls,
including zero `ResolveActorOrganization`. Each wrapper deep-copies inputs/outputs. There is no
Ent adapter in this packet; 143-04F-C owns adapter persistence. The exact fake oracle proves an
absent optional backend produces `ErrEventInvalid` and leaves every backend-call counter at zero.

`CalendarEventUpdateAuthorizationRequest` contains `EventID` and, only if the attendees patch is
present, a deep-copied `[]CalendarEventCreateAttendeeCandidate`. It carries no client
organization, owner, organizer, source or version truth. `CalendarEventUpdateAuthorization`
contains authoritative full `CalendarEventCommandInput Current`, active attendee snapshots
`{AttendeeID, UserID, Email, DisplayName, Role, Status, RespondedAt, Version}`, resolved desired
attendees, and `HasMaterializedException`. It is valid only if `Current.EventID` matches, its
organization equals scope, source is Hub, organizer equals the scope actor, every identifier and
bound is valid, and normalized identities are unique. The resolver locks/filters organization,
live master and organizer and resolves requested attendees. Missing, cross-organization and
non-organizer requests all return `ErrEventNotFound` before terminal/source or `ExpectedVersion`
comparison.

The frozen presence-preserving patch ABI distinguishes omitted from explicit null and empty.
The exact declarations (not merely illustrative prose) are:

```go
type CalendarEventUpdateString struct{ Present bool; Value string }
type CalendarEventUpdateBool struct{ Present bool; Value bool }
type CalendarEventUpdateTime struct{ Present bool; Value *time.Time }
type CalendarEventUpdateDate struct{ Present bool; Value *string }
type CalendarEventUpdateRRULE struct{ Present bool; Value *string }
type CalendarEventUpdateInt struct{ Present bool; Value *int }
type CalendarEventUpdateExdates struct{ Present bool; Value []string }
type CalendarEventUpdateAttendees struct{ Present bool; Value []CalendarEventCreateAttendeeCandidate }
type CalendarEventUpdateInput struct {
	EventID uuid.UUID; ExpectedVersion uint; NotifyAttendees bool
	Title CalendarEventUpdateString; Description CalendarEventUpdateString; Location CalendarEventUpdateString
	StartsAt CalendarEventUpdateTime; EndsAt CalendarEventUpdateTime; StartDate CalendarEventUpdateDate; EndDateExclusive CalendarEventUpdateDate
	AllDay CalendarEventUpdateBool; Timezone CalendarEventUpdateString; RRULE CalendarEventUpdateRRULE; Exdates CalendarEventUpdateExdates; ReminderMinutes CalendarEventUpdateInt; Attendees CalendarEventUpdateAttendees
}
type CalendarEventUpdateCommand struct{ input CalendarEventUpdateInput }
func NewCalendarEventUpdateCommand(input CalendarEventUpdateInput) (CalendarEventUpdateCommand, error)
func (s *EventService) UpdateCalendarEvent(ctx context.Context, actorID uuid.UUID, command CalendarEventUpdateCommand) (CalendarEventUpdateResult, error)
func (s *EventService) CancelCalendarEvent(ctx context.Context, actorID, eventID uuid.UUID, expectedVersion uint) (CalendarEventUpdateResult, error)
type CalendarEventUpdateAttendeeSnapshot struct { AttendeeID uuid.UUID; UserID *uuid.UUID; Email string; DisplayName string; Role CalendarEventAttendeeRole; Status RSVPStatus; RespondedAt *time.Time; Version uint }
type CalendarEventUpdateAuthorization struct { Current CalendarEventCommandInput; Attendees []CalendarEventUpdateAttendeeSnapshot; ResolvedAttendees []CalendarEventCreateResolvedAttendee; HasMaterializedException bool }
type CalendarEventUpdateAttendeeMutationKind string
const ( CalendarEventUpdateAttendeeAdd CalendarEventUpdateAttendeeMutationKind = "add"; CalendarEventUpdateAttendeeReset CalendarEventUpdateAttendeeMutationKind = "reset"; CalendarEventUpdateAttendeeRemove CalendarEventUpdateAttendeeMutationKind = "remove" )
type CalendarEventUpdateAttendeeMutation struct { Kind CalendarEventUpdateAttendeeMutationKind; AttendeeID uuid.UUID; UserID *uuid.UUID; Email string; DisplayName string; Role CalendarEventAttendeeRole; Status RSVPStatus; RespondedAt *time.Time; Version uint; RemovedAt *time.Time }
type CalendarEventUpdatePersistenceCommand struct { event CalendarEventCommand; expectedVersion uint; attendees []CalendarEventUpdateAttendeeMutation; outbox []EventOutboxDescriptor }
func NewCalendarEventUpdatePersistenceCommand(CalendarEventCommand, uint, []CalendarEventUpdateAttendeeMutation, []EventOutboxDescriptor) (CalendarEventUpdatePersistenceCommand, error)
func (c CalendarEventUpdatePersistenceCommand) Event() CalendarEventCommand
func (c CalendarEventUpdatePersistenceCommand) ExpectedVersion() uint
func (c CalendarEventUpdatePersistenceCommand) Attendees() []CalendarEventUpdateAttendeeMutation
func (c CalendarEventUpdatePersistenceCommand) Outbox() []EventOutboxDescriptor
type CalendarEventUpdateAttendeePersistenceResult struct { AttendeeID uuid.UUID; Version uint; Removed bool }
type CalendarEventUpdatePersistenceResult struct { eventID uuid.UUID; version uint; sequence uint; attendees []CalendarEventUpdateAttendeePersistenceResult }
func NewCalendarEventUpdatePersistenceResult(uuid.UUID, uint, uint, []CalendarEventUpdateAttendeePersistenceResult) (CalendarEventUpdatePersistenceResult, error)
func (r CalendarEventUpdatePersistenceResult) EventID() uuid.UUID
func (r CalendarEventUpdatePersistenceResult) Version() uint
func (r CalendarEventUpdatePersistenceResult) Sequence() uint
func (r CalendarEventUpdatePersistenceResult) Attendees() []CalendarEventUpdateAttendeePersistenceResult
type CalendarEventChangedField string
const ( CalendarEventChangedFieldTitle CalendarEventChangedField = "title"; CalendarEventChangedFieldDescription CalendarEventChangedField = "description"; CalendarEventChangedFieldLocation CalendarEventChangedField = "location"; CalendarEventChangedFieldStartsAt CalendarEventChangedField = "starts_at"; CalendarEventChangedFieldEndsAt CalendarEventChangedField = "ends_at"; CalendarEventChangedFieldStartDate CalendarEventChangedField = "start_date"; CalendarEventChangedFieldEndDateExclusive CalendarEventChangedField = "end_date_exclusive"; CalendarEventChangedFieldAllDay CalendarEventChangedField = "all_day"; CalendarEventChangedFieldTimezone CalendarEventChangedField = "timezone"; CalendarEventChangedFieldRRULE CalendarEventChangedField = "rrule"; CalendarEventChangedFieldExdates CalendarEventChangedField = "exdates"; CalendarEventChangedFieldReminderMinutes CalendarEventChangedField = "reminder_minutes"; CalendarEventChangedFieldAttendees CalendarEventChangedField = "attendees" )
type CalendarEventUpdateResult struct { eventID uuid.UUID; version uint; sequence uint; status CalendarEventStatus; changedFields []CalendarEventChangedField }
func NewCalendarEventUpdateResult(uuid.UUID, uint, uint, CalendarEventStatus, []CalendarEventChangedField) (CalendarEventUpdateResult, error)
func (r CalendarEventUpdateResult) EventID() uuid.UUID
func (r CalendarEventUpdateResult) Version() uint
func (r CalendarEventUpdateResult) Sequence() uint
func (r CalendarEventUpdateResult) Status() CalendarEventStatus
func (r CalendarEventUpdateResult) ChangedFields() []CalendarEventChangedField
```

Every constructor/service snapshots pointers and slices; `Present=false` requires the zero/nil
value. `Present=true` admits explicit nil only for temporal, RRULE and reminder patches; merge
then requires `CalendarEvent.Validate`. `EventID` is nonzero, `ExpectedVersion>0`, at least one
patch is present, and all create-equivalent bounds/canonical validators apply. Authorization is
strict/deep-copied and resolves all-and-only desired attendees. Mutations and persistence results
are bytewise attendee-ID sorted: add has fresh nonzero ID/version 1/no removal; reset is existing
ID/version+1/needs_action/nil responded; remove is existing ID/version+1, retains identity/role/
status/responded and has nonzero UTC `removed_at`. No-op does not construct or call persistence.
The command constructor validates then deep-copies attendee mutations, their pointers, and outbox
descriptors; its getters return copies. The persistence-result constructor validates, bytewise
sorts and copies attendee results; its getters return values/copies. The service-result constructor
uses private fields and validates/copies changed fields; its getters return values and a copied
`ChangedFields`. Persistence result validation requires event ID, version and sequence to exactly
echo the command event and expected output, plus one all-and-only attendee result per mutation,
bytewise sorted by ID, with the same attendee ID, expected mutation version, and
`Removed == (Kind == remove)`. Missing, extra, duplicate or any mismatch returns `ErrEventInvalid`
and rolls back. Oracles cover each mismatch. Wrapper and service prose consume only these getters;
there are no public mutable record fields.

`event_service_update.go` owns only these schema-shaped builders:

```go
type CalendarInviteUpdatedPayload struct { EventID uuid.UUID `json:"event_id"`; UID string `json:"uid"`; Sequence uint `json:"sequence"`; AttendeeIDs []uuid.UUID `json:"attendee_ids"`; ResetRSVP bool `json:"reset_rsvp"`; ChangedFields []CalendarEventChangedField `json:"changed_fields"` }
type CalendarInviteCancelledPayload struct { EventID uuid.UUID `json:"event_id"`; UID string `json:"uid"`; Sequence uint `json:"sequence"`; AttendeeIDs []uuid.UUID `json:"attendee_ids"` }
func NewCalendarInviteUpdatedEnvelopeV1(eventID uuid.UUID, occurredAt time.Time, organizationID uuid.UUID, actorID *uuid.UUID, correlationID string, payload CalendarInviteUpdatedPayload) (events.EnvelopeV1, error)
func NewCalendarInviteCancelledEnvelopeV1(eventID uuid.UUID, occurredAt time.Time, organizationID uuid.UUID, actorID *uuid.UUID, correlationID string, payload CalendarInviteCancelledPayload) (events.EnvelopeV1, error)
```

They validate nonzero IDs/org/time, optional nonzero actor, correlation through the events
boundary, UID 1..255 UTF-8 bytes, and at most 200 unique/nonzero bytewise-sorted attendees.
Updated additionally requires 1..13 unique canonically ordered changed fields. They call only
`events.NewEnvelopeV1`, use the existing exact subjects, and service update/cancel uses only
these builders. 04C persists only safe `calendar.invite.updated.v1` /
`calendar.invite.cancelled.v1` EnvelopeV1 outbox descriptors as notification intent: no iTIP is
queued or delivered, and no raw iTIP/body is stored. The former durable/raw/typed iTIP
request-descriptor requirement is `CONSUMED/HISTORICAL` for this exact-three packet because no
accepted transport port exists. A future separately accepted Mail/iTIP hydration-delivery packet
may consume a safe event and reconstruct transport content.
The persistence ABI is the opaque `CalendarEventUpdatePersistenceCommand` constructor/getter
record declared above; it is never a public-field struct literal.
Attendee mutation kind is `add`, `reset`, or `remove`: reset/remove require the existing ID; add
uses a fresh nonzero service `ids.NewEventID()` ID; remove includes `RemovedAt=clock.Now().UTC()`
and preserves historic identity/role/RSVP; reset preserves identity/role and sets
`needs_action`, `responded_at=nil`, `version+1`. The result must echo the same event ID/version
and exact attendee mutation IDs/versions; a mismatch is invalid and rolls back.

The exact transaction/service sequence is: resolve update scope; resolve and lock authorization;
then compare terminal/source and `ExpectedVersion`; merge and validate; return current without
persist for a no-op; otherwise build one command and persist atomically. Persistence applies the
event, attendee add/reset/tombstone and immutable outbox descriptors together. No receipt call is
allowed. Existing event builders/schema remain the sole authority for ordered `changed_fields`,
envelope version/subject/payload, update/cancel descriptors and recurrence identity; raw iTIP is
never emitted.

`CalendarEventUpdateConflictKind` has exactly `version_conflict`, `not_mutable`, and
`recurrence_exception_conflict`. Its error has private kind/event ID/current version fields,
internal constructors, stable content-free `Error`, `Unwrap() error { return ErrEventConflict }`,
`Kind()`, and `(value, bool)` getters for event ID/current version that return false for the two
non-version variants. Public mapping is exactly version to
`CALENDAR_EVENT_VERSION_CONFLICT` with ETag and Location, not-mutable to
`CALENDAR_EVENT_NOT_MUTABLE` without them, and recurrence to
`CALENDAR_EVENT_RECURRENCE_CONFLICT` without them. Not-found is 404 before If-Match/version.

143-03B-T owns only `removed_at`, the two live partial-index predicates, and schema metadata.
Its inactive-reader exclusion, fresh re-invite ID and same normalized-email role/resolved-identity
replacement rejection are exclusively later 143-04C `event_service_update_test.go` and service
flow obligations. Required tests additionally cover non-disclosure ordering, absent optional
backend with zero calls including `ResolveActorOrganization`, no receipt, transaction lifetime,
deep-copy, every persistence-result mismatch and full rollback.

If any materialized exception exists, RRULE, EXDATE, DTSTART/start-date, timezone and all-day
changes fail before any write with typed content-free `recurrence_exception_conflict`, which
unwraps/matches `ErrEventConflict`; public PATCH code is exactly
`CALENDAR_EVENT_RECURRENCE_CONFLICT`. End/duration-only change can proceed. Master DELETE is
normal cancel, not a recurrence conflict. No raw iTIP/body, outbox delivery claim, recurrence
rebase, or exception retention policy is authorized.

Provenance: author `gpt-5.6-terra` / medium, dispatcher-enforced; no fallback; telemetry
`UNKNOWN`. Static inspection only; no project executable, staging or commit.
