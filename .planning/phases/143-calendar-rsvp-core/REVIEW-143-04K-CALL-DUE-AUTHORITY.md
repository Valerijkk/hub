# 143-04K — ScheduledCall due-policy source-authority freeze

> **Source receipt-sync gate:** `SOURCE_ACCEPTANCE_SYNC_CANDIDATE /
> ATOMIC_EXACT15_COMMIT_REQUIRED / RUNTIME_NOT_RUN`. This dirty exact15 documentation candidate
> grants no authority. Its source receipt facts activate only after one atomic commit of precisely
> the exact15 manifest and immutable acceptance. Only then is source
> `572f911227b9d50e21aa73b87c530ddcaf0f8934` (parent
> `26b377eb288ac8e363cf6d53c744b60f097e9804`, tree
> `ac96e2f5c03256b1e00a391600325a524a840673`, subject
> `feat(calendar): add call due policy`) `STATIC_ACCEPTED / EXACT_SHA_PASS`, for authored-first
> test blob `437db3af0263cedb10dcf217cd53fbbcac03cf23` then source blob
> `376f46c1cc79956fab3479efc38750fd88161b16`. Contract/Spec/Oracle, Security/Adversarial and
> Lifecycle/Dependency/Boundary precommit plus immutable reviews PASS C/I/M `0/0/0`;
> Terra/xhigh, dispatcher-enforced/no fallback, telemetry `UNKNOWN`. Only then do lower 04K
> authority-candidate/precommit/current-next clauses become `CONSUMED/HISTORICAL`; no next packet
> is selected.

## Historical exact14 authority candidate receipt

```text
5a3012b7722ad42171a79bc0fb31577a0d48d539  .planning/STATE.md
eb3e67dc73151120d72e3b4c022327f44a194b89  .planning/AUTONOMOUS-QUESTIONS-2026-08-21.md
9edd44a3bd9f89426eab4ae30f04f750df4bc35d  .planning/decisions/specs/MAIL-SPEC.md
8c343e1c37f77ec37f65ba7e3e46f88da8fe249c  .planning/phases/143-calendar-rsvp-core/.continue-here.md
b74f51b0987f64c339158f674920ee8f71f6287c  .planning/phases/143-calendar-rsvp-core/DISPATCH-INDEX.md
94008ca56d931cbae08760a0fae3e6b167bbad32  .planning/phases/143-calendar-rsvp-core/DISPATCH-REMAINING.md
a842b8ff93278bf7248d02bb02a3517a081d3bbb  .planning/phases/143-calendar-rsvp-core/PLAN.md
da1b51175f63c9410ab7b8d5bd0a93cb814cdaaa  .planning/phases/143-calendar-rsvp-core/PREIMPLEMENTATION-READINESS.md
23e2a33f6a8eb74f94d5dc11961c14bc6e088ac5  .planning/phases/143-calendar-rsvp-core/RUNTIME-BLOCKER.md
a2487f59580a4e2f2c4086b4c652f41e1ce03ecd  .planning/phases/143-calendar-rsvp-core/REVIEW-143-04K-CALL-DUE-AUTHORITY.md
70b1cf73d04a8d2956fba3dc9a5129a9e678a888  docs/mail/CONTRACT-CHANGELOG.md
f5845d4e2180fe2d4c265b1f2b7b74e3b7625f0c  docs/mail/PERSISTENCE-CONTRACT-0.1.8.md
84d58beb20479f0764844ec0d73fa25ad4b84db2  docs/mail/QA-HANDOFF.md
618ea2ed805296f5da53eafeb494177dd2ffaa30  docs/mail/events-v1.schema.json
```

## Exact source packet and boundary

The following ordered test-first exact2 is the immutable source packet recorded above; this dirty
receipt candidate does not dispatch any further packet:

1. NEW `backend/tests/calendar/event_call_due_service_test.go`;
2. NEW `backend/internal/calendar/event_call_due_service.go`.

The accepted prerequisite is 143-04J: source commit
`e74efa827b52f35e3d3fbe8346adf0c0121ba48d`, accepted by documentation receipt
`d7926a27c5f7e4b612cb6ac5c2fd3fc778a512cc`. 04K also consumes accepted 04A's
`EventOutboxDescriptor` boundary and the existing 04L `events.EnvelopeV1` value. It does **not**
claim Calendar-subject admission from Mail 144-04G. The plural
`calls.channel.{canonical-uuid}.scheduled_due.v1` subject and its existing `callsScheduledDue`
schema are independently established contracts; this packet creates neither a stream admission nor
a webhook route.

04K is a pure policy. It owns no repository, transaction, lock, CAS, retry loop, clock, timer,
random source, adapter, Ent schema/generated output, outbox append, NATS publish, HTTP handler,
raw iTIP, filesystem, network or logger. It never mutates an event version/sequence, RSVP,
reminder state or caller memory. 04F-K alone later owns canonical event-then-call locking, CAS,
same-transaction `EventOutboxDescriptor` append, rollback and recovery. The legacy
`calls.scheduled.due` cutover is later work. AQ-02 stays a legacy backfill-only question.

## Frozen Go ABI

`event_call_due_service.go` declares exactly these exported declarations, fields and methods in
the shown order. `EventCallDueDecision` is opaque. Its two reads return fresh value data; its
private descriptor contains an immutable frozen `events.EnvelopeV1`.

```go
type EventCallDueCallSnapshot struct {
	ScheduledCallID uuid.UUID
	CalendarEventID uuid.UUID
	OrganizationID  uuid.UUID
	ProjectID       uuid.UUID
	ChannelID       uuid.UUID
	CreatorID       uuid.UUID
	ScheduledAt     time.Time
	Active          bool
	Deleted         bool
}

type EventCallDueEnvelopeInput struct {
	EventID       uuid.UUID
	OccurredAt    time.Time
	CorrelationID string
}

type EventCallDueProjectionInput struct {
	Event      EventCallProjectionEvent
	Target     EventCallProjectionTarget
	Creator    EventCallProjectionCreator
	Attendees  []EventCallProjectionAttendee
	Exceptions []EventCallProjectionException
}

type EventCallDueInput struct {
	Call      EventCallDueCallSnapshot
	Now       time.Time
	Projection *EventCallDueProjectionInput
	Envelope  EventCallDueEnvelopeInput
}

type EventCallScheduledDuePayload struct {
	ScheduledCallID uuid.UUID  `json:"scheduled_call_id"`
	CalendarEventID uuid.UUID  `json:"calendar_event_id"`
	OccurrenceKey   string     `json:"occurrence_key"`
	RecurrenceID    *string    `json:"recurrence_id"`
	ChannelID       uuid.UUID  `json:"channel_id"`
	ProjectID       uuid.UUID  `json:"project_id"`
	Title           string     `json:"title"`
	ScheduledAt     time.Time  `json:"scheduled_at"`
	DurationMinutes int        `json:"duration_minutes"`
}

type EventCallDueDisposition string

const (
	EventCallDueNotDue     EventCallDueDisposition = "not_due"
	EventCallDueReschedule EventCallDueDisposition = "reschedule"
	EventCallDueAdvance    EventCallDueDisposition = "advance"
	EventCallDueDeactivate EventCallDueDisposition = "deactivate"
	EventCallDueNoop       EventCallDueDisposition = "noop"
)

type EventCallDueSnapshot struct {
	Disposition     EventCallDueDisposition
	ScheduledCallID uuid.UUID
	CalendarEventID uuid.UUID
	OrganizationID  uuid.UUID
	ProjectID       uuid.UUID
	ChannelID       uuid.UUID
	CreatorID       uuid.UUID
	OccurrenceKey   string
	ScheduledAt     *time.Time
	Active          bool
}

type EventCallDueDecision struct { /* unexported snapshot and envelope */ }

func (decision EventCallDueDecision) Snapshot() EventCallDueSnapshot
func (decision EventCallDueDecision) Outbox() (EventOutboxDescriptor, bool)

type EventCallDueErrorKind string

const (
	EventCallDueInvalid  EventCallDueErrorKind = "invalid"
	EventCallDueBoundary EventCallDueErrorKind = "boundary"
	EventCallDueTooLarge EventCallDueErrorKind = "too_large"
)

type EventCallDueError struct { /* unexported kind */ }

var (
	ErrEventCallDueInvalid  error
	ErrEventCallDueBoundary error
	ErrEventCallDueTooLarge error
)

func (err *EventCallDueError) Error() string
func (err *EventCallDueError) Kind() EventCallDueErrorKind
func (err *EventCallDueError) Unwrap() error
func (err *EventCallDueError) Is(target error) bool

func ProjectEventCallDue(input EventCallDueInput) (EventCallDueDecision, error)
```

The field order and tags above are ABI. In particular, `EventCallScheduledDuePayload` has exactly
nine fields, `RecurrenceID` is `*string`, and it has no `omitempty`; a one-time occurrence emits
the JSON field as `null`. `Snapshot()` deep-copies `ScheduledAt` on every call. `Outbox()` returns
`false` and a zero descriptor when no due envelope exists; otherwise it returns the complete frozen
`EventOutboxDescriptor` by value, whose `Envelope()` exposes the immutable `EnvelopeV1`. The
decision stores that descriptor privately and retains no pointer, slice or other mutable alias from
input.

## Frozen validation, key, envelope and state table

- Every call identity is non-zero. `Call.ScheduledAt` and `Now` are semantic UTC, whole-second,
  exact 20-byte `YYYY-MM-DDTHH:MM:SSZ` instants. `Deleted=true` requires `Active=false`; the call
  identity/time/state is validated before every fast path.
- An inactive or deleted call requires `Projection==nil` and a zero `Envelope` (zero EventID UUID, zero
  OccurredAt, empty CorrelationID). It returns the identity-only inactive `noop` immediately and
  must not construct a 04J input, call `ProjectEventCall`, construct an envelope or descriptor.
  This fast path is not a bypass for malformed call identity/time/state.
- An active non-deleted call requires a nonnil `Projection`, a non-zero valid envelope EventID,
  `Envelope.OccurredAt == Now`, and a valid canonical correlation. Before either 04J call,
  `Projection.Target` must be `Exists=true`, `Active=true`, `Deleted=false`; every non-zero raw
  Event/Target/Creator scheduled-call, event, organization, project, channel and creator identity
  must exactly equal `Call` (including `Event.OrganizerID == Creator.UserID == Call.CreatorID`). A
  zero identity is `invalid`; identity mismatch or target-state mismatch is `boundary`. 04K
  deep-clones every projection value, slice and nested pointer before use; the caller supplies no
  action or boundary.
  It constructs exactly two accepted raw 04J inputs and calls `ProjectEventCall` exactly twice:

  ```go
  EventCallProjectionInput{
      Action: EventCallProjectionReconcile,
      Event: clone.Projection.Event, Target: clone.Projection.Target,
      Creator: clone.Projection.Creator, Attendees: clone.Projection.Attendees,
      Exceptions: clone.Projection.Exceptions,
      Boundary: EventCallProjectionBoundary{At: Call.ScheduledAt, Inclusive: true},
  }
  EventCallProjectionInput{
      Action: EventCallProjectionReconcile,
      Event: clone.Projection.Event, Target: clone.Projection.Target,
      Creator: clone.Projection.Creator, Attendees: clone.Projection.Attendees,
      Exceptions: clone.Projection.Exceptions,
      Boundary: EventCallProjectionBoundary{At: Now, Inclusive: false},
  }
  ```

  Their snapshots are respectively `Current` and `AfterNow`; 04K verifies each identity against
  `Call` (including `Call.CalendarEventID == snapshot.EventID`) and enforces the state table below.
  It never accepts caller-supplied snapshots, action or boundary, so a caller cannot skip an
  occurrence. An 04J `invalid`, `boundary` or `too_large` result maps respectively to the matching
  content-free 04K typed error; no raw 04J error is exposed. 04K derives no wall time and consults
  no storage.
- `Envelope.EventID`, `OccurredAt` and `CorrelationID` are a fresh candidate only for active 04K
  evaluation. A non-due, reschedule or no-outbox deactivation discards it without persistence. Only
  after a due outcome can future 04F-K atomically reserve the exact
  `(CalendarEventID, ScheduledCallID, OccurrenceKey)` tuple plus candidate values; when that
  reservation wins, all three become immutable and are reloaded for CAS retry/redelivery before
  rerunning 04K. 04K allocates and stores no ID, reservation or durable state. Correlation stays an
  injected safe job value, never a required UUID-derived value. Repeating the same complete 04K
  input returns byte-identical `EventOutboxDescriptor` data.
- An upsert title is valid UTF-8, control-free and `1..255` bytes. Its duration is exact positive
  whole minutes `1..1440`; arithmetic overflow is invalid. Envelope correlation is valid UTF-8,
  control-free, non-empty, canonical (no trim/normalization change) and at most 128 bytes.
  `Envelope.EventID` is non-zero and no envelope field is generated by the policy. Structural/
  canonical violations are `invalid`; caps are `too_large`.
- For a due upsert, `OccurrenceKey` is exactly lower-case hexadecimal SHA-256 of this byte
  preimage, with no delimiter, JSON, UUID text or timezone data:

  ```text
  "hub.calendar.scheduled_due.v1\x00" || CalendarEventID[16 raw bytes] ||
  ScheduledCallID[16 raw bytes] || nominal-occurrence-key[20 ASCII bytes]
  ```

  `CalendarEventID` is `Call.CalendarEventID` and equals both 04J snapshot `EventID` values. The
  nominal key is `Current.Occurrence.Key`, even when the
  effective occurrence was moved; different event or call IDs therefore produce different keys.
  The policy rejects a non-20-byte/non-UTC-second nominal key before hashing. It inherits 04J's
  DTSTART-first, UNTIL-inclusive, DST-gap-skip/not-counted and earliest-fold effective occurrence
  semantics; it never expands a RRULE itself.

  The non-tautological fixed vector is CalendarEventID
  `00112233-4455-6677-8899-aabbccddeeff`, ScheduledCallID
  `10213243-5465-7687-8899-aabbccddeeff`, nominal key `2026-08-21T09:00:00Z`: the exact 82-byte
  preimage above hashes to lowercase SHA-256
  `02ccb08c267ddd7eb7049ea6ab52d72bbe3cdecfac0b61758119aaa81316d0ab`. The distinct envelope
  EventID literal is `20314253-6475-8697-a8b9-cadbecfd0e1f`; it is explicitly not in the key
  preimage.
- A due envelope is built only as
  `events.NewEnvelopeV1(subject, EventID, OccurredAt, OrganizationID, nil, CorrelationID, payload)`
  followed by `NewEventOutboxDescriptor(envelope)`. The descriptor stays private inside the opaque
  decision and is not appended here. The subject is canonical plural
  `calls.channel.{lowercase Call.ChannelID}.scheduled_due.v1`; the payload is exactly the
  nine-field value above. `OccurrenceKey` is the hash; recurring `RecurrenceID` points to the
  nominal key, one-time `RecurrenceID` is nil; `ScheduledAt`, title and duration come only from
  the validated `Current` effective occurrence. Because generic `NewEnvelopeV1` and descriptor
  construction do not know this family, 04K must then validate its own frozen JSON: exactly those
  nine payload fields/tags and values, no unknown field, lower-case plural subject, and equality
  of subject channel, payload `scheduled_call_id`, `calendar_event_id`, `project_id` and
  `channel_id` with `Call`; and equality of envelope organization, event ID, occurred-at and
  correlation with `Call.OrganizationID` and `Envelope` input. A binding mismatch is `invalid`.
  It returns no raw iTIP/body or direct NATS intent.

| Ordered validated state | Decision | Outbox | Future 04F-K command intent |
| --- | --- | --- | --- |
| inactive or deleted call | `noop`, inactive, nil scheduled time, empty key | none | none |
| active, `Now < Call.ScheduledAt`, `AfterNow=upsert` at the stored start | `not_due`, active with copied call time, empty key | none | none |
| active, `Now < Call.ScheduledAt`, `AfterNow=upsert` at a different start (moved earlier/later or EXDATE skip) | `reschedule`, active with `AfterNow` start, empty key | none | set after-now start only |
| active, `Now < Call.ScheduledAt`, `AfterNow=completed` | `deactivate`, inactive, nil scheduled time, empty key | none | deactivate only |
| active, `Now >= Call.ScheduledAt`, `Current=completed` and `AfterNow=completed` | `deactivate`, inactive, nil scheduled time, empty key | none | deactivate only |
| active, `Now >= Call.ScheduledAt`, `Current=upsert` at a start different from stored | `reschedule`, active with current start, empty key | none | set current start only |
| active, `Now >= Call.ScheduledAt`, `Current=upsert` at stored start and `AfterNow=upsert` successor | `advance`, active with after-now start and due occurrence key | one | append descriptor and advance |
| active, `Now >= Call.ScheduledAt`, `Current=upsert` at stored start and `AfterNow=completed` | `deactivate`, inactive with nil scheduled time and due occurrence key | one | append descriptor and deactivate |

The rows are evaluated in the displayed order. In the early branch (`Now < Call.ScheduledAt`),
`AfterNow` is authoritative: `Current` cannot override it or create an earlier-than-current
boundary. At exact due time, strict-after `AfterNow` is a successor, so exactly one stored due
occurrence is emitted; an overdue stored occurrence has the same single outcome and no catch-up
burst exists. Any relation not in the table — notably an early `AfterNow` before `Now`, a due
`AfterNow` that is not a successor, equal effective instant with a different nominal key,
inconsistent completion, or an attempt to serialize a paused/deleted 04J result — returns zero
decision plus `ErrEventCallDueBoundary`.

Known error texts are exactly `calendar: event call due invalid`,
`calendar: event call due boundary mismatch`, and `calendar: event call due too large`. Nil
receiver or unknown kind returns generic content-free `calendar: event call due failed`, zero
kind, nil unwrap and no sentinel match. Errors reveal no title, UUID, correlation, recurrence,
payload or body.

## Mandatory external-package unit oracles

The authored-first genuine isolated unit file must reflection-freeze every exported type, field
order/type/tag, constant, method/function signature, sentinel and known/nil/unknown error behavior.
It must prove the full table, including every output `Active`/`ScheduledAt`/`OccurrenceKey` value
and outbox absence/presence for early, exact due, overdue, reschedule, finite exhaustion, inactive,
deleted and invalid relations; target active/existence/deleted and identity preflight rows; early
moved-earlier/moved-later/EXDATE-skip rows; raw-projection deep-clone/nonmutation; no
caller-controlled action/boundary; 04J kind-to-04K kind mapping; UTC-second, title, duration,
`Now == Envelope.OccurredAt` and correlation edges; `Current` effective duration for moved occurrences;
one-time `recurrence_id:null`; recurring nominal recurrence ID; the exact nine JSON payload fields
and plural lower-case subject/channel invariant; the fixed 82-byte SHA-256 vector above, differing
CalendarEventIDs, moved nominal key and DST fixtures; a changed only Envelope.EventID leaves the
occurrence key unchanged but changes frozen envelope bytes/event_id, while a changed CalendarEventID
changes key and payload calendar_event_id; deep copies/input non-mutation/repeated-call determinism;
observable
absence of dependency interfaces and callbacks; reservation-owned replay with byte-identical
descriptor plus every payload/envelope/subject binding; and no event/version/sequence, RSVP or
reminder mutation. Tests remain units: no adapter, generated Ent, transaction or integration
substitute is permitted. Security/Lifecycle static review, not a unit oracle, owns the exact two
internal calls/boundaries and import/global/I-O absence.

04F-K remains the only future CAS/lock/rollback/same-transaction outbox packet. Ent generation,
migration, concrete adapters, API/UI, raw iTIP, legacy cutover, launcher and executable/runtime
evidence remain `BLOCKED`, `NOT_READY` or `NOT_RUN`.
