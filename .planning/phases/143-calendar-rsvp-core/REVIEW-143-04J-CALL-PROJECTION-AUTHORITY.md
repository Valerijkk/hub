# 143-04J — ScheduledCall projection source-authority freeze

> **Source-receipt sync:** `SOURCE_ACCEPTANCE_SYNC_CANDIDATE /
> ATOMIC_EXACT14_COMMIT_REQUIRED / RUNTIME_NOT_RUN`. These proposed receipt facts grant no
> authority before one atomic commit of the exact14 docs manifest. After it, immutable source
> `e74efa827b52f35e3d3fbe8346adf0c0121ba48d` (parent
> `06a0d6e506a31fb5546ef168a5dba72d8956a984`, tree
> `073e887a957bb65c6c06a9759e239323ab882aa7`, subject
> `feat(calendar): add call projection policy`) is `STATIC_ACCEPTED / EXACT_SHA_PASS` only for
> test blob `28273191ddd1babec2cbe61c217cd561322a7ac0` then source blob
> `cc0218bddce25936cbfb8e987a0bb6681f979d5c`. Tests were authored first. Contract/Spec/Oracle,
> Security/Adversarial and Lifecycle/Dependency/Boundary precommit plus immutable exact-SHA
> reviews PASS C/I/M `0/0/0`; author Terra/xhigh, dispatcher-enforced/no fallback, telemetry
> `UNKNOWN`. Lower 04J candidate/authority/source-unaccepted routes then become
> `CONSUMED/HISTORICAL`; the ABI below remains current. No next packet is selected.

> **Acceptance-sync gate:** `ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_EXACT13_COMMIT_REQUIRED /
> SOURCE_IMPLEMENTATION_BLOCKED / RUNTIME_NOT_RUN`. These proposed receipt facts have no authority
> before one atomic commit of the same exact13 docs manifest. After that commit, immutable
> candidate `5841621720cc03682b3d1f1e2af9cba2953797cf` (parent
> `2a2975056c3fb8001d88863a041310271282989d`, tree
> `c4cad5a344ae4972b99410c851f20ab8c73104e7`, subject
> `docs(calendar): propose call projection authority`) is `SOURCE_AUTHORITY_ACCEPTED /
> EXACT_SHA_PASS`. Contract/Spec/Oracle, Security and Lifecycle immutable reviews PASS C/I/M
> `0/0/0`; provenance is dispatcher-enforced, no fallback, telemetry `UNKNOWN`. Only then do the
> lower candidate/precommit/exact-review-next clauses become `CONSUMED/HISTORICAL`, and root may
> dispatch only the ordered test-first exact2. Source remains
> `UNIMPLEMENTED / UNCOMMITTED / UNACCEPTED` and needs its own TDD/review/commit/receipt lifecycle.

> **Status:** `ROOT_SOURCE_AUTHORITY_CANDIDATE / PRECOMMIT_REVIEW_REQUIRED /
> IMPLEMENTATION_BLOCKED / RUNTIME_NOT_RUN`. This exact13 documentation candidate is
> neither source nor acceptance authority. Only one atomic commit of these exact thirteen docs,
> independent Contract/Spec/Oracle, Security and Lifecycle immutable exact-SHA PASS reviews, and
> then a separate docs-only acceptance sync may authorize the future source packet.

## Exact source packet and boundary

Only after that full lifecycle, root may dispatch the ordered test-first exact2:

1. NEW `backend/tests/calendar/event_call_policy_test.go`;
2. NEW `backend/internal/calendar/event_call_policy.go`.

04J owns a caller-owned, deterministic projection policy only. It owns no repository, Ent schema
or generated output, adapter, handler, OpenAPI, migration, transaction, outbox, worker, NATS
client, HTTP route, raw iTIP, clock, random source, logger, filesystem or network. 04K, 04F-J and
04F-K are independent later packets. The candidate does not select them, accept source, or alter
launcher `NOT_READY` / executable-runtime `NOT_RUN`.

Prerequisites are accepted 143-01, 143-02A, 143-04A, 143-04G and 143-04B. AQ-02/founder
question 18 remains a legacy recurring-row backfill/persistence question only: 04J receives
normalized caller-owned canonical state and never chooses a legacy timezone or performs a lookup.

## Frozen Go ABI

`event_call_policy.go` declares exactly these exported declarations, fields and methods in the
shown order. `EventCallProjection` is opaque and its sole read is a deep-copying `Snapshot()`.

```go
type EventCallProjectionAction string

const (
	EventCallProjectionReconcile EventCallProjectionAction = "reconcile"
	EventCallProjectionPause     EventCallProjectionAction = "pause"
	EventCallProjectionResume    EventCallProjectionAction = "resume"
	EventCallProjectionDelete    EventCallProjectionAction = "delete"
)

type EventCallProjectionCreator struct {
	UserID              uuid.UUID
	Email               string
	OrganizationID      uuid.UUID
	ProjectID           uuid.UUID
	ChannelID           uuid.UUID
	Active              bool
	ActiveProjectMember bool
	ActiveChannelMember bool
}

type EventCallProjectionAttendee struct {
	UserID              *uuid.UUID
	Email               string
	OrganizationID      uuid.UUID
	ProjectID           uuid.UUID
	Role                CalendarEventAttendeeRole
	Status              RSVPStatus
	RecurrenceID        string
	Active              bool
	ActiveProjectMember bool
}

type EventCallProjectionException struct {
	RecurrenceID string
	Temporal     *CalendarEvent
	Status       CalendarEventStatus
	Cancelled    bool
}

type EventCallProjectionEvent struct {
	EventID         uuid.UUID
	ScheduledCallID uuid.UUID
	OrganizationID  uuid.UUID
	ProjectID       uuid.UUID
	ChannelID       uuid.UUID
	OrganizerID     uuid.UUID
	OrganizerEmail  string
	Source          CalendarEventSource
	Status          CalendarEventStatus
	Deleted         bool
	Title           string
	Description     string
	Temporal        CalendarEvent
	RRULE           string
	Exdates         []string
	ReminderMinutes *int
}

type EventCallProjectionTarget struct {
	ScheduledCallID uuid.UUID
	EventID         uuid.UUID
	CreatorID       uuid.UUID
	OrganizationID  uuid.UUID
	ProjectID       uuid.UUID
	ChannelID       uuid.UUID
	Exists          bool
	Active          bool
	Deleted         bool
}

type EventCallProjectionBoundary struct {
	At        time.Time
	Inclusive bool
}

type EventCallProjectionInput struct {
	Action     EventCallProjectionAction
	Event      EventCallProjectionEvent
	Target     EventCallProjectionTarget
	Creator    EventCallProjectionCreator
	Attendees  []EventCallProjectionAttendee
	Exceptions []EventCallProjectionException
	Boundary   EventCallProjectionBoundary
}

type EventCallProjectionDisposition string

const (
	EventCallProjectionUpsert    EventCallProjectionDisposition = "upsert"
	EventCallProjectionPaused    EventCallProjectionDisposition = "paused"
	EventCallProjectionCompleted EventCallProjectionDisposition = "completed"
	EventCallProjectionDeleted   EventCallProjectionDisposition = "deleted"
)

type EventCallProjectionOccurrence struct {
	Key      string
	StartsAt time.Time
	EndsAt   time.Time
}

type EventCallProjectionSnapshot struct {
	Disposition    EventCallProjectionDisposition
	ScheduledCallID uuid.UUID
	EventID        uuid.UUID
	OrganizationID uuid.UUID
	ProjectID      uuid.UUID
	ChannelID      uuid.UUID
	CreatorID      uuid.UUID
	Title          string
	Description    string
	DurationMinutes int
	ReminderMinutes *int
	RRULE          string
	Occurrence     *EventCallProjectionOccurrence
	ParticipantIDs []uuid.UUID
}

type EventCallProjection struct { /* unexported snapshot */ }

func (projection EventCallProjection) Snapshot() EventCallProjectionSnapshot

type EventCallProjectionErrorKind string

const (
	EventCallProjectionInvalid  EventCallProjectionErrorKind = "invalid"
	EventCallProjectionBoundary EventCallProjectionErrorKind = "boundary"
	EventCallProjectionTooLarge EventCallProjectionErrorKind = "too_large"
)

type EventCallProjectionError struct { /* unexported kind */ }

var (
	ErrEventCallProjectionInvalid  error
	ErrEventCallProjectionBoundary error
	ErrEventCallProjectionTooLarge error
)

func (err *EventCallProjectionError) Error() string
func (err *EventCallProjectionError) Kind() EventCallProjectionErrorKind
func (err *EventCallProjectionError) Unwrap() error
func (err *EventCallProjectionError) Is(target error) bool

func ProjectEventCall(input EventCallProjectionInput) (EventCallProjection, error)
```

`Snapshot()` returns fresh copies of its pointer, occurrence and participant data every time.
`paused`, `completed` and `deleted` snapshots carry only their exact identity
(`ScheduledCallID`, event/organization/project/channel/creator IDs) and disposition; title, description,
duration, reminder, RRULE, occurrence and participants are zero/nil. No result retains input
aliases or mutates its input.

## Frozen state, validation and mapping

- Every identity is non-zero. `Target.ScheduledCallID == Event.ScheduledCallID`,
  `Target.EventID == Event.EventID`, and
  `Target.CreatorID == Event.OrganizerID == Creator.UserID`. Event, target and creator have
  exactly equal organization, project and channel IDs. `Exists=false` still carries that complete
  identity and requires `Active=false`, `Deleted=false`; `Deleted=true` requires `Exists=true`,
  `Active=false`. A target-only or creator-only authorization context is invalid.
- Event source is exactly `CalendarEventSourceHub`. `Deleted=true` requires cancelled status.
  `confirmed` and `tentative` are active canonical states only when `Deleted=false`; cancelled
  (whether deleted or not) is terminal. The
  active actions are `reconcile`, `pause` and `resume`; `delete` requires the canonical event to
  be cancelled or deleted and produces only the deleted projection disposition. No action can
  cancel, revive or otherwise mutate the canonical event.
- Creator is an active, active-project-member, active-channel-member resolved internal user and
  exactly matches the event organizer ID and normalized organizer email. Every attendee is an
  active, active-project-member resolved member of the same organization/project; email-only/
  external, inactive, foreign, cross-project, duplicate normalized email or non-nil duplicate
  user identity is invalid. Role is exactly `required|optional`; a non-empty attendee
  `RecurrenceID` uses the same canonical timed-key format and must identify a nominal occurrence
  of this event; an empty-RRULE one-time event forbids every occurrence attendee row. There are at
  most 200 attendees.
  Exactly one series attendee
  (`RecurrenceID==""`) matches the creator user ID and normalized email as organizer. Only distinct
  series-level non-organizers with `needs_action|accepted|tentative` form sorted UUID-byte
  `ParticipantIDs`; every occurrence attendee (`RecurrenceID!=""`) is excluded regardless of
  status. `declined` is valid but does not project.
- The event is timed. Master `StartsAt` and `EndsAt`, plus every non-nil exception temporal
  instant, must have `Location()==time.UTC`, `Nanosecond()==0`, and exact whole-second RFC3339
  rendering; offsets, fractional seconds and implicit truncation/normalization reject. All-day,
  invalid timezone/interval, invalid source/status relationship or malformed input is invalid.
  Duration is checked exact whole minutes `1..1440`: 1 and 1440 pass;
  0, 1441 and arithmetic overflow reject. Title is valid UTF-8, control-free, non-empty and at
  most 255 bytes; description is valid UTF-8, control-free and at most 2000 bytes. Reminder is
  `nil|0..40320`. RRULE uses only the accepted 04B canonical timed subset and is at most 1000
  bytes; Exdates and Exceptions each contain at most 1000 entries. These limits do not broaden
  04B's RRULE grammar or 1000 nominal-occurrence cap.
- Preflight enforces all count and byte caps before cloning slices, allocating maps, loading a
  timezone or expanding recurrence. Before normalization, every organizer/creator/attendee email
  is valid UTF-8, ASCII/control/space-safe, a strict non-empty address of at most 320 bytes and
  byte-equal to the accepted canonical lowercase/trimmed form. Before `time.LoadLocation`, timezone
  is valid UTF-8, control-free, non-empty and at most 64 bytes. `Boundary.At` is exact UTC
  whole-second `YYYY-MM-DDTHH:MM:SSZ`
  (`2006-01-02T15:04:05Z`), never fractional/date-only/offset. `Boundary.Inclusive=true` selects
  the first effective occurrence at-or-after `At`; false selects strictly after it. This one
  explicit bit is the only strict-after seam later 04K may use.
- Empty RRULE means exactly the DTSTART master and forbids every Exdate or Exception. Otherwise
  04J inherits 04B's DTSTART-first nominal expansion, inclusive UNTIL, DST-gap skip without count
  consumption and earliest-instant fold resolution. Every timed Exdate, exception `RecurrenceID`
  and occurrence `Key` uses the same exact UTC whole-second format. Exdates are sorted and unique.
  An exception key identifies an existing nominal occurrence; a foreign nominal key rejects.
  `Cancelled=true` iff status is `CalendarEventStatusCancelled`; `Cancelled=false` requires
  `CalendarEventStatusConfirmed` or `CalendarEventStatusTentative`. A cancelled exception removes
  its nominal occurrence. A non-cancelled exception with nil `Temporal` retains its nominal interval;
  a non-nil timed temporal override replaces the effective
  start/end while the key remains nominal. EXDATE/exception overlap, duplicate exception key,
  duplicate effective UTC instant, impossible status/cancelled pairs, all-day/malformed moved
  temporal values or a moved duration outside `1..1440` reject. Effective occurrences sort by
  actual UTC start, then nominal key.
- `reconcile` needs an active non-deleted target state or a new (`Exists=false`) target and emits
  the first effective occurrence under the boundary as `upsert`; no effective occurrence emits
  identity-only `completed`. `pause` requires an existing active non-deleted target and emits
  identity-only `paused`. `resume` requires an existing inactive non-deleted target and emits the
  first effective occurrence or identity-only `completed` when exhausted. `delete` requires an
  existing non-deleted target and a cancelled/deleted canonical event, then emits identity-only
  `deleted`. State/action failure is `boundary`, never a mutation.
- Structural, identity, canonicalization and source errors return zero projection plus
  `ErrEventCallProjectionInvalid`; cap/byte/count overflow returns zero plus
  `ErrEventCallProjectionTooLarge`; boundary/action/terminal mismatch returns zero plus
  `ErrEventCallProjectionBoundary`. Known error texts are exactly `calendar: event call projection
  invalid`, `calendar: event call projection boundary mismatch` and `calendar: event call projection too
  large`. Nil receiver or unknown kind maps generically to `calendar: event call projection
  failed`, zero `Kind()`, nil `Unwrap()` and no known sentinel match. Errors expose no caller
  title, email, UUID, recurrence, timezone or body.

## Mandatory external-package unit oracles

The authored-first test reflection-freezes every public type/field order/type, string constant,
method/function signature, sentinel and known/nil/unknown error behavior. It proves full
event-target-creator identity binding, source/status/delete rows, every target state/action,
external/foreign/inactive/duplicate/role member rejection, exactly-one-series-organizer and
occurrence-attendee exclusion, creator exclusion and participant sorting,
all text/count/recurrence bounds, 1/1440 duration pass and 0/1441/overflow reject, daily/weekly/
monthly COUNT/UNTIL/DST fixtures, inclusive and strict-after boundary rows, exact recurrence key,
master/exception UTC-second rejection, attendee nominal-membership and one-time-row rejection,
email 320/timezone 64 preflight, empty-RRULE master-only,
EXDATE/cancelled/nil-temporal/moved exception behavior and both literal
cancelled/status mismatch rows, nominal cap before removal,
identity-only terminal
snapshots, deep copies/no aliases/no mutation and absence of persistence/outbox/raw iTIP/NATS/
event-version/sequence/RSVP/reminder effects.

04K, 04F-J, 04F-K, Ent generation, migration, concrete adapters, API/UI, raw iTIP, launcher and
runtime stay outside this candidate authority and `BLOCKED` or `NOT_RUN`.
