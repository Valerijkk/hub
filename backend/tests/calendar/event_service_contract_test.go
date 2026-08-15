package calendar_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/calendar"
	"github.com/valerii/hub/backend/internal/platform/events"
)

var (
	_ calendar.EventRepository         = eventRepositoryContractFake{}
	_ calendar.EventTransactionBackend = (*eventTransactionBackendFake)(nil)
	_ calendar.EventClock              = eventClockContractFake{}
	_ calendar.EventIDGenerator        = eventIDGeneratorContractFake{}
	_ calendar.EventSerializer         = eventSerializerContractFake{}
)

var (
	contractActorID        = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	contractOrganizationID = uuid.MustParse("22222222-2222-4222-8222-222222222222")
	contractEventID        = uuid.MustParse("33333333-3333-4333-8333-333333333333")
	contractAttendeeID     = uuid.MustParse("44444444-4444-4444-8444-444444444444")
	contractExceptionID    = uuid.MustParse("55555555-5555-4555-8555-555555555555")
	contractReceiptID      = uuid.MustParse("66666666-6666-4666-8666-666666666666")
)

type eventRepositoryContractFake struct{}

func (eventRepositoryContractFake) WithEventTransaction(_ context.Context, callback func(*calendar.EventTransaction) error) error {
	return calendar.RunEventTransaction(&eventTransactionBackendFake{organizationID: contractOrganizationID}, callback)
}

type eventTransactionBackendFake struct {
	organizationID uuid.UUID
	organizations  map[uuid.UUID]uuid.UUID
	lastScope      calendar.EventMutationScope
	scopes         []calendar.EventMutationScope
	appendCalls    int
}

func (fake *eventTransactionBackendFake) recordScope(scope calendar.EventMutationScope) {
	fake.lastScope = scope
	fake.scopes = append(fake.scopes, scope)
}

func (fake *eventTransactionBackendFake) ResolveActorOrganization(_ context.Context, actorID uuid.UUID) (uuid.UUID, error) {
	if fake.organizations != nil {
		return fake.organizations[actorID], nil
	}
	return fake.organizationID, nil
}

func (fake *eventTransactionBackendFake) ReserveMutationReceipt(_ context.Context, scope calendar.EventMutationScope, _ calendar.MutationReceiptReservation) (calendar.MutationReceiptReservationResult, error) {
	fake.recordScope(scope)
	return calendar.NewReservedMutationReceiptResult(contractReceiptID)
}

func (fake *eventTransactionBackendFake) CompleteMutationReceipt(_ context.Context, scope calendar.EventMutationScope, _ calendar.MutationReceiptCompletion) error {
	fake.recordScope(scope)
	return nil
}

func (fake *eventTransactionBackendFake) PersistCalendarEvent(_ context.Context, scope calendar.EventMutationScope, command calendar.CalendarEventCommand) (calendar.CalendarEventPersistenceResult, error) {
	fake.recordScope(scope)
	return calendar.NewCalendarEventPersistenceResult(command.EventID(), command.Version())
}

func (fake *eventTransactionBackendFake) PersistCalendarEventAttendee(_ context.Context, scope calendar.EventMutationScope, command calendar.CalendarEventAttendeeCommand) (calendar.CalendarEventAttendeePersistenceResult, error) {
	fake.recordScope(scope)
	return calendar.NewCalendarEventAttendeePersistenceResult(command.AttendeeID(), command.Version())
}

func (fake *eventTransactionBackendFake) PersistCalendarEventException(_ context.Context, scope calendar.EventMutationScope, command calendar.CalendarEventExceptionCommand) (calendar.CalendarEventExceptionPersistenceResult, error) {
	fake.recordScope(scope)
	return calendar.NewCalendarEventExceptionPersistenceResult(command.ExceptionID(), command.Sequence())
}

func (fake *eventTransactionBackendFake) AppendOutbox(_ context.Context, scope calendar.EventMutationScope, _ calendar.EventOutboxDescriptor) error {
	fake.recordScope(scope)
	fake.appendCalls++
	return nil
}

type eventClockContractFake struct{}

func (eventClockContractFake) Now() time.Time {
	panic("EventService constructor must not read the clock")
}

type eventIDGeneratorContractFake struct{}

func (eventIDGeneratorContractFake) NewEventID() uuid.UUID {
	panic("EventService constructor must not generate an event id")
}

type eventSerializerContractFake struct{}

func (eventSerializerContractFake) Serialize(calendar.ITIPMessage) ([]byte, error) {
	panic("EventService constructor must not serialize an invitation")
}

func TestNewEventService_DoesNotInvokeInjectedPorts(t *testing.T) {
	service := calendar.NewEventService(
		eventRepositoryContractFake{},
		eventClockContractFake{},
		eventIDGeneratorContractFake{},
		eventSerializerContractFake{},
	)
	if service == nil {
		t.Fatal("NewEventService() returned nil")
	}
}

func TestEventTransaction_BindsEveryMutationToTransactionResolvedScope(t *testing.T) {
	backend := &eventTransactionBackendFake{organizationID: contractOrganizationID}
	ctx := context.Background()
	envelope := mustContractEnvelope(t)
	descriptor, err := calendar.NewEventOutboxDescriptor(envelope)
	if err != nil {
		t.Fatalf("NewEventOutboxDescriptor() error = %v", err)
	}

	err = calendar.RunEventTransaction(backend, func(tx *calendar.EventTransaction) error {
		scope, err := tx.ResolveMutationScope(ctx, contractActorID, calendar.MutationReceiptEventCreate)
		if err != nil {
			return err
		}
		if scope.ActorID() != contractActorID || scope.OrganizationID() != contractOrganizationID || scope.Operation() != calendar.MutationReceiptEventCreate {
			t.Fatalf("resolved mutation scope = %#v", scope)
		}
		reservation, err := calendar.NewMutationReceiptReservation(scope, calendar.MutationReceiptHash{1}, calendar.MutationReceiptHash{2})
		if err != nil {
			return err
		}
		if _, err := tx.ReserveMutationReceipt(ctx, scope, reservation); err != nil {
			return err
		}
		event := mustCalendarEventCommand(t)
		if _, err := tx.PersistCalendarEvent(ctx, scope, event); err != nil {
			return err
		}
		attendee := mustCalendarEventAttendeeCommand(t)
		if _, err := tx.PersistCalendarEventAttendee(ctx, scope, attendee); err != nil {
			return err
		}
		exception := mustCalendarEventExceptionCommand(t)
		if _, err := tx.PersistCalendarEventException(ctx, scope, exception); err != nil {
			return err
		}
		if err := tx.AppendOutbox(ctx, scope, descriptor); err != nil {
			return err
		}
		completion, err := calendar.NewMutationReceiptCompletion(contractReceiptID, &contractEventID, mustReceiptResponse(t))
		if err != nil {
			return err
		}
		return tx.CompleteMutationReceipt(ctx, scope, completion)
	})
	if err != nil {
		t.Fatalf("RunEventTransaction() error = %v", err)
	}
	if len(backend.scopes) != 6 {
		t.Fatalf("transactional mutation scope calls = %d, want 6", len(backend.scopes))
	}
	for _, receivedScope := range backend.scopes {
		if receivedScope.ActorID() != contractActorID || receivedScope.OrganizationID() != contractOrganizationID || receivedScope.Operation() != calendar.MutationReceiptEventCreate {
			t.Fatalf("backend received unbound scope = %#v", receivedScope)
		}
	}
	if backend.appendCalls != 1 {
		t.Fatalf("AppendOutbox calls = %d, want 1", backend.appendCalls)
	}
}

func TestEventTransaction_RejectsCapturedUseAfterCallback(t *testing.T) {
	backend := &eventTransactionBackendFake{organizationID: contractOrganizationID}
	ctx := context.Background()
	var captured *calendar.EventTransaction
	var scope calendar.EventMutationScope

	err := calendar.RunEventTransaction(backend, func(tx *calendar.EventTransaction) error {
		captured = tx
		var err error
		scope, err = tx.ResolveMutationScope(ctx, contractActorID, calendar.MutationReceiptEventCreate)
		return err
	})
	if err != nil {
		t.Fatalf("RunEventTransaction() error = %v", err)
	}
	descriptor, err := calendar.NewEventOutboxDescriptor(mustContractEnvelope(t))
	if err != nil {
		t.Fatalf("NewEventOutboxDescriptor() error = %v", err)
	}
	err = captured.AppendOutbox(ctx, scope, descriptor)
	if !errors.Is(err, calendar.ErrEventTransactionClosed) {
		t.Fatalf("captured AppendOutbox() error = %v, want %v", err, calendar.ErrEventTransactionClosed)
	}
	if backend.appendCalls != 0 {
		t.Fatalf("backend AppendOutbox calls after closed callback = %d, want 0", backend.appendCalls)
	}
}

func TestEventTransaction_RejectsReceiptScopeFromAnotherActorOrganizationOrOperation(t *testing.T) {
	backend := &eventTransactionBackendFake{organizations: map[uuid.UUID]uuid.UUID{
		contractActorID:    contractOrganizationID,
		contractAttendeeID: contractEventID,
	}}
	ctx := context.Background()
	err := calendar.RunEventTransaction(backend, func(tx *calendar.EventTransaction) error {
		createScope, err := tx.ResolveMutationScope(ctx, contractActorID, calendar.MutationReceiptEventCreate)
		if err != nil {
			return err
		}
		otherScope, err := tx.ResolveMutationScope(ctx, contractAttendeeID, calendar.MutationReceiptRSVPSeries)
		if err != nil {
			return err
		}
		reservation, err := calendar.NewMutationReceiptReservation(createScope, calendar.MutationReceiptHash{1}, calendar.MutationReceiptHash{2})
		if err != nil {
			return err
		}
		if _, err := tx.ReserveMutationReceipt(ctx, otherScope, reservation); !errors.Is(err, calendar.ErrEventForbidden) {
			t.Fatalf("cross-scope ReserveMutationReceipt() error = %v, want %v", err, calendar.ErrEventForbidden)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("RunEventTransaction() error = %v", err)
	}
	if len(backend.scopes) != 0 {
		t.Fatalf("backend mutation calls for mismatched receipt scope = %d, want 0", len(backend.scopes))
	}
}

func TestMutationReceiptConstructors_ValidateBoundsAndFreezeReplayBytes(t *testing.T) {
	valid := mustReceiptResponse(t)
	responseBytes := valid.JSON()
	responseBytes[0] = '['
	if got := string(valid.JSON()); got != `{"ok":true}` {
		t.Fatalf("MutationReceiptResponse.JSON() = %s", got)
	}
	replayed, err := calendar.NewReplayedMutationReceiptResult(contractReceiptID, valid)
	if err != nil {
		t.Fatalf("NewReplayedMutationReceiptResult() error = %v", err)
	}
	copyOne, ok := replayed.ReplayedResponse()
	if !ok {
		t.Fatal("ReplayedResponse() is absent")
	}
	copyOneBytes := copyOne.JSON()
	copyOneBytes[0] = '['
	copyTwo, ok := replayed.ReplayedResponse()
	if !ok || string(copyTwo.JSON()) != `{"ok":true}` {
		t.Fatalf("replayed response changed after caller mutation: %#v", copyTwo)
	}

	overLimit := make([]byte, (1<<20)+1)
	for _, testCase := range []struct {
		name   string
		status int
		body   []byte
	}{
		{name: "status below HTTP range", status: 99, body: []byte(`{}`)},
		{name: "status above HTTP range", status: 600, body: []byte(`{}`)},
		{name: "empty JSON", status: 200, body: nil},
		{name: "invalid JSON", status: 200, body: []byte(`{`)},
		{name: "response over one MiB", status: 200, body: overLimit},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := calendar.NewMutationReceiptResponse(testCase.status, testCase.body)
			if !errors.Is(err, calendar.ErrEventInvalid) {
				t.Fatalf("NewMutationReceiptResponse() error = %v, want %v", err, calendar.ErrEventInvalid)
			}
		})
	}
	if _, err := calendar.NewReplayedMutationReceiptResult(uuid.Nil, valid); !errors.Is(err, calendar.ErrEventInvalid) {
		t.Fatalf("zero receipt replay error = %v", err)
	}
	if _, err := calendar.NewReplayedMutationReceiptResult(contractReceiptID, calendar.MutationReceiptResponse{}); !errors.Is(err, calendar.ErrEventInvalid) {
		t.Fatalf("zero response replay error = %v", err)
	}
	if _, err := calendar.NewMutationReceiptCompletion(contractReceiptID, nil, calendar.MutationReceiptResponse{}); !errors.Is(err, calendar.ErrEventInvalid) {
		t.Fatalf("zero response completion error = %v", err)
	}
}

func TestCalendarPersistenceCommands_AreValidatedAndImmutable(t *testing.T) {
	exdates := []string{"2026-08-22"}
	reminder := 15
	organizer := contractActorID
	event, err := calendar.NewCalendarEventCommand(calendar.CalendarEventCommandInput{
		EventID:         contractEventID,
		OrganizationID:  contractOrganizationID,
		OwnerID:         contractActorID,
		UID:             "calendar-contract-event",
		Temporal:        calendar.CalendarEvent{StartsAt: time.Date(2026, time.August, 20, 9, 0, 0, 0, time.UTC), EndsAt: time.Date(2026, time.August, 20, 10, 0, 0, 0, time.UTC), Timezone: "UTC"},
		Exdates:         exdates,
		Sequence:        0,
		Version:         1,
		Source:          calendar.CalendarEventSourceHub,
		Status:          calendar.CalendarEventStatusConfirmed,
		OrganizerID:     &organizer,
		OrganizerEmail:  "owner@example.com",
		ReminderMinutes: &reminder,
	})
	if err != nil {
		t.Fatalf("NewCalendarEventCommand() error = %v", err)
	}
	exdates[0] = "changed"
	if got := event.Exdates(); len(got) != 1 || got[0] != "2026-08-22" {
		t.Fatalf("Event command did not freeze exdates: %#v", got)
	}
	if _, err := calendar.NewCalendarEventCommand(calendar.CalendarEventCommandInput{}); !errors.Is(err, calendar.ErrEventInvalid) {
		t.Fatalf("invalid event command error = %v", err)
	}
	if _, err := calendar.NewCalendarEventAttendeeCommand(calendar.CalendarEventAttendeeCommandInput{}); !errors.Is(err, calendar.ErrEventInvalid) {
		t.Fatalf("invalid attendee command error = %v", err)
	}
	if _, err := calendar.NewCalendarEventExceptionCommand(calendar.CalendarEventExceptionCommandInput{}); !errors.Is(err, calendar.ErrEventInvalid) {
		t.Fatalf("invalid exception command error = %v", err)
	}
	if _, err := calendar.NewCalendarEventPersistenceResult(uuid.Nil, 1); !errors.Is(err, calendar.ErrEventInvalid) {
		t.Fatalf("invalid event persistence result error = %v", err)
	}
	if _, err := calendar.NewCalendarEventAttendeePersistenceResult(contractAttendeeID, 0); !errors.Is(err, calendar.ErrEventInvalid) {
		t.Fatalf("invalid attendee persistence result error = %v", err)
	}
	if _, err := calendar.NewCalendarEventExceptionPersistenceResult(uuid.Nil, 0); !errors.Is(err, calendar.ErrEventInvalid) {
		t.Fatalf("invalid exception persistence result error = %v", err)
	}
}

func TestEventOutboxDescriptor_ContainsCompleteEnvelope(t *testing.T) {
	descriptor, err := calendar.NewEventOutboxDescriptor(mustContractEnvelope(t))
	if err != nil {
		t.Fatalf("NewEventOutboxDescriptor() error = %v", err)
	}
	encoded, err := json.Marshal(descriptor.Envelope())
	if err != nil {
		t.Fatalf("json.Marshal(descriptor.Envelope()) error = %v", err)
	}

	var complete map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &complete); err != nil {
		t.Fatalf("json.Unmarshal(complete envelope) error = %v", err)
	}
	for _, field := range []string{
		"subject",
		"schema_version",
		"event_id",
		"occurred_at",
		"organization_id",
		"actor_id",
		"correlation_id",
		"payload",
	} {
		if _, ok := complete[field]; !ok {
			t.Fatalf("complete envelope does not contain %q", field)
		}
	}
}

func TestEventErrors_AreExactTypedAndDistinct(t *testing.T) {
	cases := []struct {
		error   error
		literal string
	}{
		{error: calendar.ErrEventNotFound, literal: "calendar event not found"},
		{error: calendar.ErrEventForbidden, literal: "calendar event access denied"},
		{error: calendar.ErrEventConflict, literal: "calendar event conflict"},
		{error: calendar.ErrEventInvalid, literal: "calendar event invalid"},
		{error: calendar.ErrEventTransactionClosed, literal: "calendar event transaction closed"},
	}
	for index, testCase := range cases {
		if got := testCase.error.Error(); got != testCase.literal {
			t.Fatalf("error literal = %q, want %q", got, testCase.literal)
		}
		wrapped := fmt.Errorf("operation failed: %w", testCase.error)
		if !errors.Is(wrapped, testCase.error) {
			t.Fatalf("errors.Is(%v, %v) = false", wrapped, testCase.error)
		}
		if strings.Contains(testCase.error.Error(), "SENSITIVE-CALENDAR-04A") {
			t.Fatalf("safe error leaked fixture value: %q", testCase.error)
		}
		for otherIndex, other := range cases {
			if index != otherIndex && errors.Is(testCase.error, other.error) {
				t.Fatalf("errors.Is(%v, %v) = true for distinct sentinels", testCase.error, other.error)
			}
		}
	}
}

func mustReceiptResponse(t *testing.T) calendar.MutationReceiptResponse {
	t.Helper()
	response, err := calendar.NewMutationReceiptResponse(201, []byte(`{"ok":true}`))
	if err != nil {
		t.Fatalf("NewMutationReceiptResponse() error = %v", err)
	}
	return response
}

func mustCalendarEventCommand(t *testing.T) calendar.CalendarEventCommand {
	t.Helper()
	command, err := calendar.NewCalendarEventCommand(calendar.CalendarEventCommandInput{
		EventID:        contractEventID,
		OrganizationID: contractOrganizationID,
		OwnerID:        contractActorID,
		UID:            "calendar-contract-event",
		Temporal:       calendar.CalendarEvent{StartsAt: time.Date(2026, time.August, 20, 9, 0, 0, 0, time.UTC), EndsAt: time.Date(2026, time.August, 20, 10, 0, 0, 0, time.UTC), Timezone: "UTC"},
		Version:        1,
		Source:         calendar.CalendarEventSourceHub,
		Status:         calendar.CalendarEventStatusConfirmed,
		OrganizerEmail: "owner@example.com",
	})
	if err != nil {
		t.Fatalf("NewCalendarEventCommand() error = %v", err)
	}
	return command
}

func mustCalendarEventAttendeeCommand(t *testing.T) calendar.CalendarEventAttendeeCommand {
	t.Helper()
	command, err := calendar.NewCalendarEventAttendeeCommand(calendar.CalendarEventAttendeeCommandInput{
		AttendeeID: contractAttendeeID,
		EventID:    contractEventID,
		Email:      "attendee@example.com",
		Role:       calendar.CalendarEventAttendeeRoleRequired,
		Status:     calendar.RSVPNeedsAction,
		Version:    1,
	})
	if err != nil {
		t.Fatalf("NewCalendarEventAttendeeCommand() error = %v", err)
	}
	return command
}

func mustCalendarEventExceptionCommand(t *testing.T) calendar.CalendarEventExceptionCommand {
	t.Helper()
	command, err := calendar.NewCalendarEventExceptionCommand(calendar.CalendarEventExceptionCommandInput{
		ExceptionID:  contractExceptionID,
		EventID:      contractEventID,
		RecurrenceID: "2026-08-21T09:00:00Z",
		Sequence:     1,
		Status:       calendar.CalendarEventStatusCancelled,
		Cancelled:    true,
	})
	if err != nil {
		t.Fatalf("NewCalendarEventExceptionCommand() error = %v", err)
	}
	return command
}

func mustContractEnvelope(t *testing.T) events.EnvelopeV1 {
	t.Helper()
	envelope, err := events.NewEnvelopeV1(
		"calendar.event.created.v1",
		contractEventID,
		time.Date(2030, time.May, 6, 7, 8, 9, 0, time.UTC),
		contractOrganizationID,
		nil,
		"calendar-event-contract",
		map[string]any{"event_id": contractEventID.String()},
	)
	if err != nil {
		t.Fatalf("NewEnvelopeV1() error = %v", err)
	}
	return envelope
}
