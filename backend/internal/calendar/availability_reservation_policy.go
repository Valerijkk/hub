package calendar

import (
	"sort"
	"time"

	"github.com/google/uuid"
)

const (
	calendarAvailabilityReservationMaxOccurrences = 1000
	calendarAvailabilityReservationMaxRows        = 51000
	calendarAvailabilityReservationMaxAttendees   = 50
)

// AvailabilityReservationErrorKind is the safe public classification for a
// final scheduling conflict. It intentionally contains no person, event or
// temporal detail.
type AvailabilityReservationErrorKind string

const (
	AvailabilityReservationErrorSlotUnavailable AvailabilityReservationErrorKind = "slot_unavailable"
)

// AvailabilityReservationError exposes only a stable scheduling result.
type AvailabilityReservationError struct {
	kind AvailabilityReservationErrorKind
}

func (err AvailabilityReservationError) Error() string {
	return "calendar availability slot unavailable"
}

func (err AvailabilityReservationError) Unwrap() error {
	return ErrEventConflict
}

func (err AvailabilityReservationError) Is(target error) bool {
	switch value := target.(type) {
	case AvailabilityReservationError:
		return err.kind == value.kind
	case *AvailabilityReservationError:
		return value != nil && err.kind == value.kind
	default:
		return false
	}
}

// Kind returns the stable machine-readable scheduling result.
func (err AvailabilityReservationError) Kind() AvailabilityReservationErrorKind {
	return err.kind
}

var ErrAvailabilitySlotUnavailable error = AvailabilityReservationError{kind: AvailabilityReservationErrorSlotUnavailable}

// CalendarAvailabilityReservationOccurrence is one materialized timed
// occurrence. The occurrence key is the canonical UTC start instant, so a
// retry has one immutable scheduling identity.
type CalendarAvailabilityReservationOccurrence struct {
	OccurrenceKey string
	StartsAt      time.Time
	EndsAt        time.Time
}

// CalendarAvailabilityReservationSet is an opaque desired busy set for one
// canonical timed event. It is intentionally independent of transport and
// persistence so a future meeting service can compose it inside one event tx.
type CalendarAvailabilityReservationSet struct {
	organizationID uuid.UUID
	eventID        uuid.UUID
	organizerID    uuid.UUID
	participants   []uuid.UUID
	occurrences    []CalendarAvailabilityReservationOccurrence
}

// NewCalendarAvailabilityReservationSet validates a bounded, duplicate-free
// organizer and internal-attendee set before any transaction or I/O begins.
func NewCalendarAvailabilityReservationSet(organizationID, eventID, organizerID uuid.UUID, attendeeIDs []uuid.UUID, occurrences []CalendarAvailabilityReservationOccurrence) (CalendarAvailabilityReservationSet, error) {
	if organizationID == uuid.Nil || eventID == uuid.Nil || organizerID == uuid.Nil || len(attendeeIDs) > calendarAvailabilityReservationMaxAttendees || occurrences == nil || len(occurrences) > calendarAvailabilityReservationMaxOccurrences {
		return CalendarAvailabilityReservationSet{}, ErrAvailabilitySourceInvalid
	}

	participants := make([]uuid.UUID, 0, len(attendeeIDs)+1)
	participants = append(participants, organizerID)
	seenParticipants := map[uuid.UUID]struct{}{organizerID: {}}
	for _, attendeeID := range attendeeIDs {
		if attendeeID == uuid.Nil {
			return CalendarAvailabilityReservationSet{}, ErrAvailabilitySourceInvalid
		}
		if _, duplicate := seenParticipants[attendeeID]; duplicate {
			return CalendarAvailabilityReservationSet{}, ErrAvailabilitySourceInvalid
		}
		seenParticipants[attendeeID] = struct{}{}
		participants = append(participants, attendeeID)
	}
	sort.Slice(participants, func(left, right int) bool {
		return participants[left].String() < participants[right].String()
	})

	if len(participants)*len(occurrences) > calendarAvailabilityReservationMaxRows {
		return CalendarAvailabilityReservationSet{}, ErrAvailabilityLimitExceeded
	}
	clonedOccurrences := make([]CalendarAvailabilityReservationOccurrence, len(occurrences))
	seenOccurrences := make(map[string]struct{}, len(occurrences))
	for index, occurrence := range occurrences {
		startsAt, endsAt, ok := canonicalCalendarAvailabilityReservationInterval(occurrence.StartsAt, occurrence.EndsAt)
		if !ok || occurrence.OccurrenceKey != startsAt.Format(time.RFC3339Nano) {
			return CalendarAvailabilityReservationSet{}, ErrAvailabilitySourceInvalid
		}
		if _, duplicate := seenOccurrences[occurrence.OccurrenceKey]; duplicate {
			return CalendarAvailabilityReservationSet{}, ErrAvailabilitySourceInvalid
		}
		seenOccurrences[occurrence.OccurrenceKey] = struct{}{}
		clonedOccurrences[index] = CalendarAvailabilityReservationOccurrence{OccurrenceKey: occurrence.OccurrenceKey, StartsAt: startsAt, EndsAt: endsAt}
	}
	sort.Slice(clonedOccurrences, func(left, right int) bool {
		if !clonedOccurrences[left].StartsAt.Equal(clonedOccurrences[right].StartsAt) {
			return clonedOccurrences[left].StartsAt.Before(clonedOccurrences[right].StartsAt)
		}
		return clonedOccurrences[left].OccurrenceKey < clonedOccurrences[right].OccurrenceKey
	})

	return CalendarAvailabilityReservationSet{
		organizationID: organizationID,
		eventID:        eventID,
		organizerID:    organizerID,
		participants:   participants,
		occurrences:    clonedOccurrences,
	}, nil
}

func (set CalendarAvailabilityReservationSet) OrganizationID() uuid.UUID { return set.organizationID }
func (set CalendarAvailabilityReservationSet) EventID() uuid.UUID        { return set.eventID }
func (set CalendarAvailabilityReservationSet) OrganizerID() uuid.UUID    { return set.organizerID }

// Participants returns a sorted, detached organizer-plus-internal-attendee set.
func (set CalendarAvailabilityReservationSet) Participants() []uuid.UUID {
	if set.participants == nil {
		return nil
	}
	participants := make([]uuid.UUID, len(set.participants))
	copy(participants, set.participants)
	return participants
}

// Occurrences returns detached, canonical materialized occurrences.
func (set CalendarAvailabilityReservationSet) Occurrences() []CalendarAvailabilityReservationOccurrence {
	if set.occurrences == nil {
		return nil
	}
	occurrences := make([]CalendarAvailabilityReservationOccurrence, len(set.occurrences))
	copy(occurrences, set.occurrences)
	return occurrences
}

func (set CalendarAvailabilityReservationSet) valid() bool {
	_, err := NewCalendarAvailabilityReservationSet(set.organizationID, set.eventID, set.organizerID, set.participantsWithoutOrganizer(), set.occurrences)
	return err == nil
}

func (set CalendarAvailabilityReservationSet) participantsWithoutOrganizer() []uuid.UUID {
	attendees := make([]uuid.UUID, 0, len(set.participants))
	for _, participantID := range set.participants {
		if participantID != set.organizerID {
			attendees = append(attendees, participantID)
		}
	}
	return attendees
}

func canonicalCalendarAvailabilityReservationInterval(startsAt, endsAt time.Time) (time.Time, time.Time, bool) {
	if startsAt.IsZero() || endsAt.IsZero() || startsAt.Location() != time.UTC || endsAt.Location() != time.UTC || !endsAt.After(startsAt) {
		return time.Time{}, time.Time{}, false
	}
	return startsAt.UTC(), endsAt.UTC(), true
}
