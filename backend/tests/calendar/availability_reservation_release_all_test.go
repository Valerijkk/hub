package calendar_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/calendar"
)

var (
	availabilityReleaseOrganizationID = uuid.MustParse("d1111111-1111-4111-8111-111111111111")
	availabilityReleaseEventID        = uuid.MustParse("d2222222-2222-4222-8222-222222222222")
	availabilityReleaseOrganizerID    = uuid.MustParse("d3333333-3333-4333-8333-333333333333")
)

func TestCalendarAvailabilityReservationSet_ExplicitEmptyReleasesAllAndSurvivesTransactionClone(t *testing.T) {
	var zero calendar.CalendarAvailabilityReservationSet
	if occurrences := zero.Occurrences(); occurrences != nil {
		t.Fatalf("zero reservation occurrences = %#v, want nil", occurrences)
	}
	if _, err := calendar.NewCalendarAvailabilityReservationSet(availabilityReleaseOrganizationID, availabilityReleaseEventID, availabilityReleaseOrganizerID, nil, nil); !errors.Is(err, calendar.ErrAvailabilitySourceInvalid) {
		t.Fatalf("nil occurrences error = %v, want %v", err, calendar.ErrAvailabilitySourceInvalid)
	}

	empty := []calendar.CalendarAvailabilityReservationOccurrence{}
	set, err := calendar.NewCalendarAvailabilityReservationSet(availabilityReleaseOrganizationID, availabilityReleaseEventID, availabilityReleaseOrganizerID, nil, empty)
	if err != nil {
		t.Fatalf("explicit empty reservation set error = %v", err)
	}
	if participants := set.Participants(); len(participants) != 1 || participants[0] != availabilityReleaseOrganizerID {
		t.Fatalf("empty release participants = %#v, want sole organizer", participants)
	}
	if occurrences := set.Occurrences(); occurrences == nil || len(occurrences) != 0 {
		t.Fatalf("empty release occurrences = %#v, want nonnil empty", occurrences)
	}

	backend := &availabilityReleaseAllBackendFake{}
	err = calendar.RunEventTransaction(backend, func(transaction *calendar.EventTransaction) error {
		scope, err := transaction.ResolveMutationScope(context.Background(), availabilityReleaseOrganizerID, calendar.MutationReceiptEventCreate)
		if err != nil {
			return err
		}
		return transaction.ReconcileCalendarAvailabilityReservations(context.Background(), scope, set)
	})
	if err != nil {
		t.Fatalf("release-all transaction error = %v", err)
	}
	if backend.reconcileCalls != 1 {
		t.Fatalf("release-all reconcile calls = %d, want 1", backend.reconcileCalls)
	}
	if occurrences := backend.set.Occurrences(); occurrences == nil || len(occurrences) != 0 {
		t.Fatalf("cloned release-all occurrences = %#v, want nonnil empty", occurrences)
	}

	nonempty, err := calendar.NewCalendarAvailabilityReservationSet(
		availabilityReleaseOrganizationID,
		availabilityReleaseEventID,
		availabilityReleaseOrganizerID,
		nil,
		[]calendar.CalendarAvailabilityReservationOccurrence{{
			OccurrenceKey: "2031-01-02T03:04:05Z",
			StartsAt:      time.Date(2031, time.January, 2, 3, 4, 5, 0, time.UTC),
			EndsAt:        time.Date(2031, time.January, 2, 4, 4, 5, 0, time.UTC),
		}},
	)
	if err != nil {
		t.Fatalf("nonempty reservation set error = %v", err)
	}
	if occurrences := nonempty.Occurrences(); len(occurrences) != 1 || occurrences[0].OccurrenceKey != "2031-01-02T03:04:05Z" {
		t.Fatalf("nonempty reservation occurrences = %#v", occurrences)
	}
}

func TestCalendarAvailabilityReservationReleaseAllEntSourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/calendar/availability_reservation_ent.go")
	if err != nil {
		t.Fatalf("read availability reservation source: %v", err)
	}
	source := string(sourceBytes)
	start := strings.Index(source, "func (backend *eventEntTransactionBackend) ReconcileCalendarAvailabilityReservations")
	end := strings.Index(source, "func (backend *eventEntTransactionBackend) acquireCalendarAvailabilityAdvisoryLocks")
	if start < 0 || end <= start {
		t.Fatal("reconcile source bounds are missing")
	}
	reconcile := source[start:end]
	eventLock := "lockCalendarAvailabilityReservationEvent(ctx, scope, set, releaseAll)"
	constraint := "requireCalendarAvailabilityReservationConstraint(ctx)"
	releaseAllPosition := strings.Index(reconcile, "releaseAll := len(occurrences) == 0")
	eventLockPosition := strings.Index(reconcile, eventLock)
	constraintPosition := strings.Index(reconcile, constraint)
	releaseStart := strings.Index(reconcile, "if releaseAll {")
	if releaseAllPosition < 0 || eventLockPosition < 0 || constraintPosition < 0 || releaseStart < 0 {
		t.Fatal("release-all reconciliation source is missing")
	}
	if releaseAllPosition >= eventLockPosition || eventLockPosition >= constraintPosition || constraintPosition >= releaseStart {
		t.Fatal("release-all reconciliation must lock the event, require the constraint, then enter the release branch")
	}
	membersLock := "members, err := backend.lockCalendarAvailabilityReservationMembers"
	releaseEnd := strings.Index(reconcile[releaseStart:], membersLock)
	if releaseEnd < 0 {
		t.Fatal("release-all branch does not precede member resolution")
	}
	releaseBranch := reconcile[releaseStart : releaseStart+releaseEnd]
	releaseRequired := []string{
		"if releaseAll {",
		"active, err := backend.lockActiveCalendarAvailabilityReservations(ctx, set)",
		"calendarAvailabilityReservationActiveRowsMatch(active, desired)",
		"canonicalAvailabilityTime(backend.now().UTC().Truncate(time.Second))",
		"releaseObsoleteCalendarAvailabilityReservations(ctx, active, desired, now)",
	}
	previous := -1
	for _, fragment := range releaseRequired {
		position := strings.Index(releaseBranch, fragment)
		if position < 0 || position <= previous {
			t.Fatalf("release-all reconciliation source is missing or unordered %q", fragment)
		}
		previous = position
	}
	for _, forbidden := range []string{
		"lockCalendarAvailabilityReservationMembers",
		"checkCalendarAvailabilityReservationLeaves",
		"checkCalendarAvailabilityReservationEvents",
		"checkCalendarAvailabilityReservationRows",
		"checkCalendarAvailabilityReservationCalls",
		"insertMissingCalendarAvailabilityReservations",
	} {
		if strings.Contains(releaseBranch, forbidden) {
			t.Fatalf("release-all branch must skip %q", forbidden)
		}
	}
	nonempty := reconcile[releaseStart+releaseEnd:]
	nonemptyRequired := []string{
		membersLock,
		"active, err := backend.lockActiveCalendarAvailabilityReservations(ctx, set)",
		"checkCalendarAvailabilityReservationLeaves(ctx, members, occurrences)",
		"checkCalendarAvailabilityReservationEvents(ctx, set, participants, occurrences)",
		"checkCalendarAvailabilityReservationRows(ctx, set, participants, occurrences)",
		"checkCalendarAvailabilityReservationCalls(ctx, participants, occurrences)",
	}
	previous = constraintPosition
	for _, fragment := range nonemptyRequired {
		position := strings.Index(nonempty, fragment)
		if position < 0 {
			t.Fatalf("nonempty reconciliation source is missing %q", fragment)
		}
		position += releaseStart + releaseEnd
		if position <= previous {
			t.Fatalf("nonempty reconciliation source is unordered %q", fragment)
		}
		previous = position
	}
	for _, requiredScope := range []string{
		"calendaravailabilityreservation.OrganizationIDEQ(set.OrganizationID())",
		"calendaravailabilityreservation.CalendarEventIDEQ(set.EventID())",
	} {
		if !strings.Contains(source, requiredScope) {
			t.Fatalf("active reservation lock is missing exact event scope %q", requiredScope)
		}
	}
	eventStart := strings.Index(source, "func (backend *eventEntTransactionBackend) lockCalendarAvailabilityReservationEvent")
	if eventStart < 0 {
		t.Fatal("event authority lock source is missing")
	}
	eventEnd := strings.Index(source[eventStart:], "func calendarAvailabilityReservationAttendeesMatch")
	if eventEnd < 0 {
		t.Fatal("event authority lock source end is missing")
	}
	eventLockSource := source[eventStart : eventStart+eventEnd]
	releaseReturn := strings.Index(eventLockSource, "if releaseAll {")
	attendeeRead := strings.Index(eventLockSource, "backend.tx.CalendarAttendee.Query()")
	exceptionRead := strings.Index(eventLockSource, "backend.tx.CalendarEventException.Query()")
	if releaseReturn < 0 || attendeeRead < 0 || exceptionRead < 0 || releaseReturn >= attendeeRead || releaseReturn >= exceptionRead {
		t.Fatal("release-all event authority lock must return before attendee and exception source reads")
	}
}

type availabilityReleaseAllBackendFake struct {
	reconcileCalls int
	set            calendar.CalendarAvailabilityReservationSet
}

func (backend *availabilityReleaseAllBackendFake) ResolveActorOrganization(_ context.Context, actorID uuid.UUID) (uuid.UUID, error) {
	if actorID != availabilityReleaseOrganizerID {
		return uuid.Nil, calendar.ErrEventForbidden
	}
	return availabilityReleaseOrganizationID, nil
}

func (backend *availabilityReleaseAllBackendFake) ReconcileCalendarAvailabilityReservations(_ context.Context, _ calendar.EventMutationScope, set calendar.CalendarAvailabilityReservationSet) error {
	backend.reconcileCalls++
	backend.set = set
	return nil
}

func (*availabilityReleaseAllBackendFake) ResolveCalendarEventCreateAuthorization(context.Context, calendar.EventMutationScope, calendar.CalendarEventCreateAuthorizationRequest) (calendar.CalendarEventCreateAuthorization, error) {
	return calendar.CalendarEventCreateAuthorization{}, calendar.ErrEventInvalid
}

func (*availabilityReleaseAllBackendFake) ReserveMutationReceipt(context.Context, calendar.EventMutationScope, calendar.MutationReceiptReservation) (calendar.MutationReceiptReservationResult, error) {
	return calendar.MutationReceiptReservationResult{}, calendar.ErrEventInvalid
}

func (*availabilityReleaseAllBackendFake) CompleteMutationReceipt(context.Context, calendar.EventMutationScope, calendar.MutationReceiptCompletion) error {
	return calendar.ErrEventInvalid
}

func (*availabilityReleaseAllBackendFake) PersistCalendarEvent(context.Context, calendar.EventMutationScope, calendar.CalendarEventCommand) (calendar.CalendarEventPersistenceResult, error) {
	return calendar.CalendarEventPersistenceResult{}, calendar.ErrEventInvalid
}

func (*availabilityReleaseAllBackendFake) PersistCalendarEventAttendee(context.Context, calendar.EventMutationScope, calendar.CalendarEventAttendeeCommand) (calendar.CalendarEventAttendeePersistenceResult, error) {
	return calendar.CalendarEventAttendeePersistenceResult{}, calendar.ErrEventInvalid
}

func (*availabilityReleaseAllBackendFake) PersistCalendarEventException(context.Context, calendar.EventMutationScope, calendar.CalendarEventExceptionCommand) (calendar.CalendarEventExceptionPersistenceResult, error) {
	return calendar.CalendarEventExceptionPersistenceResult{}, calendar.ErrEventInvalid
}

func (*availabilityReleaseAllBackendFake) AppendOutbox(context.Context, calendar.EventMutationScope, calendar.EventOutboxDescriptor) error {
	return calendar.ErrEventInvalid
}
