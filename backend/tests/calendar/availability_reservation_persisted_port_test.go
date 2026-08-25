package calendar_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/calendar"
)

var (
	persistedAvailabilityPortOrganizationID = uuid.MustParse("e1111111-1111-4111-8111-111111111111")
	persistedAvailabilityPortActorID        = uuid.MustParse("e2222222-2222-4222-8222-222222222222")
	persistedAvailabilityPortEventID        = uuid.MustParse("e3333333-3333-4333-8333-333333333333")
)

var _ calendar.EventTransactionBackend = (*persistedAvailabilityPortBackend)(nil)
var _ calendar.PersistedAvailabilityReservationTransactionBackend = (*persistedAvailabilityPortBackend)(nil)
var _ calendar.AvailabilityReservationTransactionBackend = (*persistedAvailabilityPortBaseBackend)(nil)

func TestEventTransactionPersistedAvailabilityReconciliation_GuardsPostPersistDelegation(t *testing.T) {
	backend := &persistedAvailabilityPortBackend{}
	ctx := context.WithValue(context.Background(), persistedAvailabilityPortContextKey{}, "preserve")
	var retained *calendar.EventTransaction
	var scope calendar.EventMutationScope
	err := calendar.RunEventTransaction(backend, func(transaction *calendar.EventTransaction) error {
		retained = transaction
		var err error
		scope, err = transaction.ResolveMutationScope(ctx, persistedAvailabilityPortActorID, calendar.MutationReceiptMeetingUpdate)
		if err != nil {
			return err
		}
		return transaction.ReconcilePersistedCalendarAvailabilityReservations(ctx, scope, persistedAvailabilityPortEventID)
	})
	if err != nil {
		t.Fatalf("RunEventTransaction() error = %v", err)
	}
	if backend.persistedCalls != 1 || backend.legacyReconcileCalls != 0 || backend.otherCalls != 0 {
		t.Fatalf("delegation calls persisted:%d legacy:%d other:%d", backend.persistedCalls, backend.legacyReconcileCalls, backend.otherCalls)
	}
	if backend.ctx != ctx || backend.scope != scope || backend.eventID != persistedAvailabilityPortEventID {
		t.Fatalf("delegation identity ctx:%v scope:%#v event:%s", backend.ctx == ctx, backend.scope, backend.eventID)
	}
	if retained == nil {
		t.Fatal("transaction was not retained for closed-transaction guard")
	}
	if err := retained.ReconcilePersistedCalendarAvailabilityReservations(context.Background(), scope, persistedAvailabilityPortEventID); !errors.Is(err, calendar.ErrEventTransactionClosed) {
		t.Fatalf("closed transaction error = %v, want %v", err, calendar.ErrEventTransactionClosed)
	}
	if backend.persistedCalls != 1 {
		t.Fatalf("closed transaction invoked backend %d times", backend.persistedCalls)
	}

	for _, operation := range []calendar.MutationReceiptOperation{
		calendar.MutationReceiptEventCreate,
		calendar.MutationReceiptRSVPSeries,
		calendar.MutationReceiptRSVPOccurrence,
	} {
		err := calendar.RunEventTransaction(backend, func(transaction *calendar.EventTransaction) error {
			invalidScope, resolveErr := transaction.ResolveMutationScope(context.Background(), persistedAvailabilityPortActorID, operation)
			if resolveErr != nil {
				return resolveErr
			}
			return transaction.ReconcilePersistedCalendarAvailabilityReservations(context.Background(), invalidScope, persistedAvailabilityPortEventID)
		})
		if !errors.Is(err, calendar.ErrEventInvalid) {
			t.Fatalf("operation %q error = %v, want %v", operation, err, calendar.ErrEventInvalid)
		}
	}
	if backend.persistedCalls != 1 || backend.legacyReconcileCalls != 0 {
		t.Fatalf("invalid operation bypassed guard persisted:%d legacy:%d", backend.persistedCalls, backend.legacyReconcileCalls)
	}

	err = calendar.RunEventTransaction(backend, func(transaction *calendar.EventTransaction) error {
		validScope, resolveErr := transaction.ResolveMutationScope(context.Background(), persistedAvailabilityPortActorID, calendar.MutationReceiptMeetingCancel)
		if resolveErr != nil {
			return resolveErr
		}
		if err := transaction.ReconcilePersistedCalendarAvailabilityReservations(nil, validScope, persistedAvailabilityPortEventID); !errors.Is(err, calendar.ErrEventInvalid) {
			t.Fatalf("nil context error = %v, want %v", err, calendar.ErrEventInvalid)
		}
		if err := transaction.ReconcilePersistedCalendarAvailabilityReservations(context.Background(), validScope, uuid.Nil); !errors.Is(err, calendar.ErrEventInvalid) {
			t.Fatalf("nil event error = %v, want %v", err, calendar.ErrEventInvalid)
		}
		if err := transaction.ReconcilePersistedCalendarAvailabilityReservations(context.Background(), calendar.EventMutationScope{}, persistedAvailabilityPortEventID); !errors.Is(err, calendar.ErrEventForbidden) {
			t.Fatalf("invalid scope error = %v, want %v", err, calendar.ErrEventForbidden)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("guard transaction error = %v", err)
	}
	if backend.persistedCalls != 1 || backend.legacyReconcileCalls != 0 {
		t.Fatalf("invalid input bypassed guard persisted:%d legacy:%d", backend.persistedCalls, backend.legacyReconcileCalls)
	}

	missing := &persistedAvailabilityPortBaseBackend{}
	err = calendar.RunEventTransaction(missing, func(transaction *calendar.EventTransaction) error {
		validScope, resolveErr := transaction.ResolveMutationScope(context.Background(), persistedAvailabilityPortActorID, calendar.MutationReceiptMeetingUpdate)
		if resolveErr != nil {
			return resolveErr
		}
		return transaction.ReconcilePersistedCalendarAvailabilityReservations(context.Background(), validScope, persistedAvailabilityPortEventID)
	})
	if !errors.Is(err, calendar.ErrEventInternal) {
		t.Fatalf("missing capability error = %v, want %v", err, calendar.ErrEventInternal)
	}
	if missing.legacyReconcileCalls != 0 || missing.otherCalls != 0 {
		t.Fatalf("missing capability fell back legacy:%d other:%d", missing.legacyReconcileCalls, missing.otherCalls)
	}

	backend.err = errors.New("storage endpoint must not leak")
	err = calendar.RunEventTransaction(backend, func(transaction *calendar.EventTransaction) error {
		validScope, resolveErr := transaction.ResolveMutationScope(context.Background(), persistedAvailabilityPortActorID, calendar.MutationReceiptMeetingUpdate)
		if resolveErr != nil {
			return resolveErr
		}
		return transaction.ReconcilePersistedCalendarAvailabilityReservations(context.Background(), validScope, persistedAvailabilityPortEventID)
	})
	if !errors.Is(err, calendar.ErrEventInternal) || errors.Is(err, backend.err) || strings.Contains(err.Error(), "storage endpoint") {
		t.Fatalf("backend error leaked as %v", err)
	}

	sourceBytes, readErr := os.ReadFile("../../internal/calendar/event_repository.go")
	if readErr != nil {
		t.Fatalf("read event repository source: %v", readErr)
	}
	source := string(sourceBytes)
	start := strings.Index(source, "func (transaction *EventTransaction) ReconcilePersistedCalendarAvailabilityReservations")
	end := strings.Index(source, "func (transaction *EventTransaction) ResolveMeetingCreateScope")
	if start < 0 || end <= start {
		t.Fatal("persisted availability port source bounds are missing")
	}
	port := source[start:end]
	for _, operation := range []string{
		"EventMutationMeetingUpdate",
		"EventMutationMeetingCancel",
		"EventMutationCalendarUpdate",
		"EventMutationCalendarOccurrenceCancel",
	} {
		if !strings.Contains(port, operation) {
			t.Fatalf("post-persist allowlist is missing %s", operation)
		}
	}
	for _, forbidden := range []string{
		"EventMutationEventCreate",
		"EventMutationMeetingCreate",
		"EventMutationRSVPSeries",
		"EventMutationRSVPOccurrence",
		"EventMutationInboundReply",
		"ReconcileCalendarAvailabilityReservations(ctx, scope",
	} {
		if strings.Contains(port, forbidden) {
			t.Fatalf("post-persist port must not accept or fall back to %q", forbidden)
		}
	}
}

func TestEventEntPersistedAvailabilityReconciliation_SourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/calendar/availability_reservation_ent.go")
	if err != nil {
		t.Fatalf("read availability reservation Ent source: %v", err)
	}
	source := string(sourceBytes)
	if !strings.Contains(source, "var _ PersistedAvailabilityReservationTransactionBackend = (*eventEntTransactionBackend)(nil)") {
		t.Fatal("Ent transaction backend must prove persisted availability capability at compile time")
	}

	reconcileStart := strings.Index(source, "func (backend *eventEntTransactionBackend) ReconcilePersistedCalendarAvailabilityReservations")
	if reconcileStart < 0 {
		t.Fatal("persisted availability reconciliation source bounds are missing")
	}
	reconcileEndRelative := strings.Index(source[reconcileStart:], "func (backend *eventEntTransactionBackend) acquireCalendarAvailabilityAdvisoryLocks")
	if reconcileEndRelative <= 0 {
		t.Fatal("persisted availability reconciliation source bounds are missing")
	}
	reconcile := source[reconcileStart : reconcileStart+reconcileEndRelative]
	for _, required := range []string{
		"backend == nil",
		"backend.tx == nil",
		"backend.now == nil",
		"ctx == nil",
		"eventID == uuid.Nil",
		"scope.ActorID() == uuid.Nil",
		"scope.OrganizationID() == uuid.Nil",
		"scope.transaction == nil",
		"validEventMutationScope(scope)",
		"EventMutationMeetingUpdate",
		"EventMutationMeetingCancel",
		"EventMutationCalendarUpdate",
		"EventMutationCalendarOccurrenceCancel",
		"backend.persistedCalendarAvailabilityReservationSet(ctx, scope, eventID)",
	} {
		if !strings.Contains(reconcile, required) {
			t.Fatalf("persisted reconciliation is missing fail-closed contract %q", required)
		}
	}
	if strings.Count(reconcile, "backend.ReconcileCalendarAvailabilityReservations(ctx, scope, set)") != 1 {
		t.Fatal("persisted reconciliation must delegate exactly once to the existing reconciliation core")
	}
	for _, forbidden := range []string{
		"AppendOutbox(",
		"ReserveMutationReceipt(",
		"CompleteMutationReceipt(",
		"CalendarEventCommand",
		"EventOutboxDescriptor",
	} {
		if strings.Contains(reconcile, forbidden) {
			t.Fatalf("persisted reconciliation must not own caller snapshot, receipt or outbox work: %q", forbidden)
		}
	}

	setStart := strings.Index(source, "func (backend *eventEntTransactionBackend) persistedCalendarAvailabilityReservationSet")
	if setStart < 0 {
		t.Fatal("persisted authoritative reservation-set source bounds are missing")
	}
	setEndRelative := strings.Index(source[setStart:], "func (backend *eventEntTransactionBackend) lockCalendarAvailabilityReservationMembers")
	if setEndRelative <= 0 {
		t.Fatal("persisted authoritative reservation-set source bounds are missing")
	}
	setSource := source[setStart : setStart+setEndRelative]
	eventRead := strings.Index(setSource, "backend.tx.CalendarEvent.Query()")
	attendeeRead := strings.Index(setSource, "backend.tx.CalendarAttendee.Query()")
	exceptionRead := strings.Index(setSource, "backend.tx.CalendarEventException.Query()")
	if eventRead < 0 || attendeeRead <= eventRead || exceptionRead <= attendeeRead {
		t.Fatal("persisted reservation derivation must read event, active master internal attendees, then exceptions")
	}
	if strings.Count(setSource, "ForUpdate()") < 3 {
		t.Fatal("persisted event, attendee and exception reads must all remain locked")
	}
	for _, required := range []string{
		"calendarevent.IDEQ(eventID)",
		"calendarevent.OrganizationIDEQ(scope.OrganizationID())",
		"calendarevent.CalendarOwnerIDEQ(scope.ActorID())",
		"calendarevent.OrganizerIDEQ(scope.ActorID())",
		"calendarattendee.UserIDNotNil()",
		"calendarattendee.RemovedAtIsNil()",
		"calendarattendee.RecurrenceIDIsNil()",
		"calendarattendee.FieldID",
		"calendareventexception.FieldRecurrenceID",
		"calendareventexception.FieldID",
		"CalendarEventStatusCancelled",
		"make([]CalendarAvailabilityReservationOccurrence, 0)",
		"availabilityEventOccurrences(event, exceptions)",
		"OccurrenceKey: startsAt.Format(time.RFC3339Nano)",
		"NewCalendarAvailabilityReservationSet(",
	} {
		if !strings.Contains(setSource, required) {
			t.Fatalf("persisted authoritative derivation is missing %q", required)
		}
	}
	if strings.Contains(setSource, "|| event.AllDay") {
		t.Fatal("all-day authoritative occurrences must be derived rather than treated as cancellation")
	}
	if strings.Contains(setSource, "OccurrenceKey: occurrence.key") {
		t.Fatal("persisted derivation must use an effective UTC start key for moved exceptions")
	}
	lockStart := strings.Index(source, "func (backend *eventEntTransactionBackend) acquireCalendarAvailabilityAdvisoryLocks")
	if lockStart < 0 {
		t.Fatal("availability advisory-lock source bounds are missing")
	}
	lockEndRelative := strings.Index(source[lockStart:], "func calendarAvailabilityReservationAdvisoryKey")
	if lockEndRelative <= 0 {
		t.Fatal("availability advisory-lock source bounds are missing")
	}
	lockSource := source[lockStart : lockStart+lockEndRelative]
	for _, required := range []string{
		"SELECT pg_try_advisory_xact_lock($1)",
		"sql.ScanOne(&rows, &acquired)",
		"rows.Close()",
		"participantID.String() <= previous",
		"if scanErr != nil || closeErr != nil",
		"if !acquired",
		"ErrAvailabilitySlotUnavailable",
	} {
		if !strings.Contains(lockSource, required) {
			t.Fatalf("availability advisory locks are missing %q", required)
		}
	}
	for _, forbidden := range []string{
		"pg_advisory_xact_lock($1)",
	} {
		if strings.Contains(lockSource, forbidden) {
			t.Fatalf("availability advisory locking must not retain %q", forbidden)
		}
	}
	eventLockStart := strings.Index(source, "func (backend *eventEntTransactionBackend) lockCalendarAvailabilityReservationEvent")
	if eventLockStart < 0 {
		t.Fatal("availability event-lock source bounds are missing")
	}
	eventLockEndRelative := strings.Index(source[eventLockStart:], "func calendarAvailabilityReservationAttendeesMatch")
	if eventLockEndRelative <= 0 {
		t.Fatal("availability event-lock source bounds are missing")
	}
	eventLockSource := source[eventLockStart : eventLockStart+eventLockEndRelative]
	for _, forbidden := range []string{"calendarevent.AllDay(false)", "event.AllDay"} {
		if strings.Contains(eventLockSource, forbidden) {
			t.Fatalf("all-day availability locking must not retain %q", forbidden)
		}
	}
	for _, forbidden := range []string{
		"AppendOutbox(",
		"ReserveMutationReceipt(",
		"CompleteMutationReceipt(",
		"CalendarEventCommand",
		"EventOutboxDescriptor",
	} {
		if strings.Contains(setSource, forbidden) {
			t.Fatalf("persisted reservation derivation must use durable rows only, not %q", forbidden)
		}
	}
}

type persistedAvailabilityPortContextKey struct{}

type persistedAvailabilityPortBackend struct {
	persistedAvailabilityPortBaseBackend
	persistedCalls int
	ctx            context.Context
	scope          calendar.EventMutationScope
	eventID        uuid.UUID
	err            error
}

func (backend *persistedAvailabilityPortBackend) ReconcilePersistedCalendarAvailabilityReservations(ctx context.Context, scope calendar.EventMutationScope, eventID uuid.UUID) error {
	backend.persistedCalls++
	backend.ctx = ctx
	backend.scope = scope
	backend.eventID = eventID
	return backend.err
}

type persistedAvailabilityPortBaseBackend struct {
	legacyReconcileCalls int
	otherCalls           int
}

func (*persistedAvailabilityPortBaseBackend) ResolveActorOrganization(_ context.Context, actorID uuid.UUID) (uuid.UUID, error) {
	if actorID != persistedAvailabilityPortActorID {
		return uuid.Nil, calendar.ErrEventForbidden
	}
	return persistedAvailabilityPortOrganizationID, nil
}

func (backend *persistedAvailabilityPortBaseBackend) ReconcileCalendarAvailabilityReservations(context.Context, calendar.EventMutationScope, calendar.CalendarAvailabilityReservationSet) error {
	backend.legacyReconcileCalls++
	return calendar.ErrEventInternal
}

func (backend *persistedAvailabilityPortBaseBackend) ResolveCalendarEventCreateAuthorization(context.Context, calendar.EventMutationScope, calendar.CalendarEventCreateAuthorizationRequest) (calendar.CalendarEventCreateAuthorization, error) {
	backend.otherCalls++
	return calendar.CalendarEventCreateAuthorization{}, calendar.ErrEventInternal
}

func (backend *persistedAvailabilityPortBaseBackend) ReserveMutationReceipt(context.Context, calendar.EventMutationScope, calendar.MutationReceiptReservation) (calendar.MutationReceiptReservationResult, error) {
	backend.otherCalls++
	return calendar.MutationReceiptReservationResult{}, calendar.ErrEventInternal
}

func (backend *persistedAvailabilityPortBaseBackend) CompleteMutationReceipt(context.Context, calendar.EventMutationScope, calendar.MutationReceiptCompletion) error {
	backend.otherCalls++
	return calendar.ErrEventInternal
}

func (backend *persistedAvailabilityPortBaseBackend) PersistCalendarEvent(context.Context, calendar.EventMutationScope, calendar.CalendarEventCommand) (calendar.CalendarEventPersistenceResult, error) {
	backend.otherCalls++
	return calendar.CalendarEventPersistenceResult{}, calendar.ErrEventInternal
}

func (backend *persistedAvailabilityPortBaseBackend) PersistCalendarEventAttendee(context.Context, calendar.EventMutationScope, calendar.CalendarEventAttendeeCommand) (calendar.CalendarEventAttendeePersistenceResult, error) {
	backend.otherCalls++
	return calendar.CalendarEventAttendeePersistenceResult{}, calendar.ErrEventInternal
}

func (backend *persistedAvailabilityPortBaseBackend) PersistCalendarEventException(context.Context, calendar.EventMutationScope, calendar.CalendarEventExceptionCommand) (calendar.CalendarEventExceptionPersistenceResult, error) {
	backend.otherCalls++
	return calendar.CalendarEventExceptionPersistenceResult{}, calendar.ErrEventInternal
}

func (backend *persistedAvailabilityPortBaseBackend) AppendOutbox(context.Context, calendar.EventMutationScope, calendar.EventOutboxDescriptor) error {
	backend.otherCalls++
	return calendar.ErrEventInternal
}
