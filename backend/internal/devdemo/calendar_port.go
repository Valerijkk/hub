package devdemo

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// CalendarDemoPort synchronously exposes the Calendar operations required by demo seeding.
// Implementations must validate all role and status strings exhaustively.
type CalendarDemoPort interface {
	CreateEvent(context.Context, uuid.UUID, CalendarDemoCreateInput) (CalendarDemoCreateResult, error)
	GetEvent(context.Context, uuid.UUID, uuid.UUID) (CalendarDemoEventState, error)
	RespondSeries(context.Context, uuid.UUID, CalendarDemoRSVPInput) (CalendarDemoRSVPResult, error)
	CancelOccurrence(context.Context, uuid.UUID, CalendarDemoOccurrenceCancelInput) (CalendarDemoOccurrenceCancelResult, error)
}

// CalendarDemoAttendeeCandidate is a resolved internal attendee snapshot.
type CalendarDemoAttendeeCandidate struct {
	UserID uuid.UUID
	Email  string
	Role   string
}

// CalendarDemoAttendeeState is a fresh attendee projection returned by the adapter.
type CalendarDemoAttendeeState struct {
	UserID     uuid.UUID
	Email      string
	Role       string
	RSVPStatus string
	Version    int
}

// CalendarDemoCreateInput is a caller-owned immutable snapshot and must not be retained.
type CalendarDemoCreateInput struct {
	IdempotencyKey uuid.UUID
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
	EventID   uuid.UUID
	Version   int
	Sequence  int
	Status    string
	RRULE     string
	Exdates   []time.Time
	Attendees []CalendarDemoAttendeeState
}

// CalendarDemoRSVPInput is a caller-owned immutable snapshot and must not be retained.
type CalendarDemoRSVPInput struct {
	IdempotencyKey          uuid.UUID
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
type CalendarDemoOccurrenceCancelInput struct {
	EventID               uuid.UUID
	RecurrenceID          time.Time
	ExpectedMasterVersion int
}

// CalendarDemoOccurrenceCancelResult is a fresh occurrence projection returned by the adapter.
type CalendarDemoOccurrenceCancelResult struct {
	EventID      uuid.UUID
	RecurrenceID time.Time
	Status       string
	Sequence     int
	EventVersion int
}
