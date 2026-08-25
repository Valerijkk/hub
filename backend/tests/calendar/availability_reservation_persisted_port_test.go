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
