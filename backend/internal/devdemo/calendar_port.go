package devdemo

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrCalendarDemoReconcile is the stable neutral errors.Is target for adapter and orchestrator failures.
// HTTP responses remain generic; this sentinel does not mirror the Calendar error taxonomy.
var ErrCalendarDemoReconcile = errors.New("calendar demo reconciliation failed")

// ErrCalendarDemoVersionConflict is a narrow errors.Is target only for an occurrence master-version conflict.
// It is always returned underneath ErrCalendarDemoReconcile and exposes neither a Calendar error type nor a current version.
var ErrCalendarDemoVersionConflict = errors.New("calendar demo version conflict")

// CalendarDemoPort synchronously exposes the Calendar operations required by demo seeding.
// Implementations must validate all role and status strings exhaustively.
// The second UUID argument on every method is the nonzero actor Hub UserID.
type CalendarDemoPort interface {
	CreateEvent(context.Context, uuid.UUID, CalendarDemoCreateInput) (CalendarDemoCreateResult, error)
	GetEvent(context.Context, uuid.UUID, uuid.UUID) (CalendarDemoEventState, error)
	RespondSeries(context.Context, uuid.UUID, CalendarDemoRSVPInput) (CalendarDemoRSVPResult, error)
	CancelOccurrence(context.Context, uuid.UUID, CalendarDemoOccurrenceCancelInput) (CalendarDemoOccurrenceCancelResult, error)
}

// CalendarDemoAttendeeCandidate is a resolved internal attendee snapshot.
// UserID must be a nonzero internal Hub identity; external attendees and uuid.Nil are unsupported.
type CalendarDemoAttendeeCandidate struct {
	UserID uuid.UUID
	Email  string
	Role   string
}

// CalendarDemoAttendeeState is a fresh attendee projection returned by the adapter.
// UserID must be a nonzero internal Hub identity; external attendees and uuid.Nil are unsupported.
type CalendarDemoAttendeeState struct {
	UserID     uuid.UUID
	Email      string
	Role       string
	RSVPStatus string
	Version    int
}

// CalendarDemoCreateInput is a caller-owned immutable snapshot and must not be retained.
// IdempotencyKey is a stable opaque 16..128 byte caller value validated by the adapter and domain.
// Replay requires the exact same key and payload.
type CalendarDemoCreateInput struct {
	IdempotencyKey string
	Title          string
	Description    string
	Location       string
	// StartsAt, EndsAt, and Exdates must be UTC.
	StartsAt  time.Time
	EndsAt    time.Time
	Timezone  string
	RRULE     string
	Exdates   []time.Time
	Attendees []CalendarDemoAttendeeCandidate
}

// CalendarDemoCreateResult is a fresh event projection returned by the adapter.
type CalendarDemoCreateResult struct {
	EventID   uuid.UUID
	Version   int
	Sequence  int
	Attendees []CalendarDemoAttendeeState
	Replayed  bool
}

// CalendarDemoEventState is a fresh event projection returned by the adapter.
type CalendarDemoEventState struct {
	EventID  uuid.UUID
	Version  int
	Sequence int
	Status   string
	RRULE    string
	// Exdates are master exclusions, not occurrence exception state.
	Exdates   []time.Time
	Attendees []CalendarDemoAttendeeState
}

// CalendarDemoRSVPInput is a caller-owned immutable snapshot and must not be retained.
// IdempotencyKey is a stable opaque 16..128 byte caller value validated by the adapter and domain.
// Replay requires the exact same key and payload.
type CalendarDemoRSVPInput struct {
	IdempotencyKey          string
	EventID                 uuid.UUID
	ExpectedAttendeeVersion int
	Status                  string
}

// CalendarDemoRSVPResult is a fresh RSVP projection returned by the adapter.
type CalendarDemoRSVPResult struct {
	UserID   uuid.UUID
	Status   string
	Version  int
	Replayed bool
}

// CalendarDemoOccurrenceCancelInput is a caller-owned immutable snapshot and must not be retained.
// The caller must GetEvent and supply the current master Version before cancellation.
// RecurrenceID must serialize to the canonical UTC recurrence string.
type CalendarDemoOccurrenceCancelInput struct {
	EventID               uuid.UUID
	RecurrenceID          time.Time
	ExpectedMasterVersion int
}

// CalendarDemoOccurrenceCancelResult is a fresh occurrence projection returned by the adapter.
// An already cancelled exception is a semantic no-op when the current version is supplied.
// On version conflict the caller performs one bounded GetEvent and CancelOccurrence retry.
// Cancellation has no receipt or Replayed semantics.
type CalendarDemoOccurrenceCancelResult struct {
	EventID      uuid.UUID
	RecurrenceID time.Time
	Status       string
	Sequence     int
	EventVersion int
}
