package calendar

import (
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/google/uuid"
)

// EventCallRepository owns the isolated transaction boundary for the Calendar-to-call bridge.
// It deliberately does not widen EventRepository or reuse EventTransaction.
type EventCallRepository interface {
	WithEventCallTransaction(context.Context, func(*EventCallTransaction) error) error
}

// EventCallTransactionBackend is the private persistence port for one bridge transaction.
type EventCallTransactionBackend interface {
	ResolveEventCallProjection(context.Context, EventCallProjectionRequest) (EventCallProjectionInput, error)
	PersistEventCallProjection(context.Context, EventCallProjectionRequest, EventCallProjection) (EventCallProjectionSnapshot, error)
	ResolveEventCallDue(context.Context, EventCallDueRequest) (EventCallDueInput, error)
	PersistEventCallDue(context.Context, EventCallDueRequest, EventCallDueDecision) (EventCallDueSnapshot, error)
}

// EventCallProjectionRequest is the caller-owned, immutable projection fence.
type EventCallProjectionRequest struct {
	scheduledCallID      uuid.UUID
	action               EventCallProjectionAction
	boundary             EventCallProjectionBoundary
	expectedEventVersion uint
	expectedCallVersion  uint
}

// NewEventCallProjectionRequest validates the narrow projection boundary before opening a transaction.
func NewEventCallProjectionRequest(scheduledCallID uuid.UUID, action EventCallProjectionAction, boundary EventCallProjectionBoundary, expectedEventVersion, expectedCallVersion uint) (EventCallProjectionRequest, error) {
	request := EventCallProjectionRequest{
		scheduledCallID:      scheduledCallID,
		action:               action,
		boundary:             EventCallProjectionBoundary{At: boundary.At, Inclusive: boundary.Inclusive},
		expectedEventVersion: expectedEventVersion,
		expectedCallVersion:  expectedCallVersion,
	}
	if !request.valid() {
		return EventCallProjectionRequest{}, ErrEventInvalid
	}
	return request, nil
}

// ScheduledCallID returns the immutable linked call identity.
func (request EventCallProjectionRequest) ScheduledCallID() uuid.UUID { return request.scheduledCallID }

// Action returns the frozen policy action.
func (request EventCallProjectionRequest) Action() EventCallProjectionAction { return request.action }

// Boundary returns a value copy of the frozen occurrence boundary.
func (request EventCallProjectionRequest) Boundary() EventCallProjectionBoundary {
	return EventCallProjectionBoundary{At: request.boundary.At, Inclusive: request.boundary.Inclusive}
}

// ExpectedEventVersion returns the exact CalendarEvent optimistic-lock fence.
func (request EventCallProjectionRequest) ExpectedEventVersion() uint { return request.expectedEventVersion }

// ExpectedCallVersion returns the exact ScheduledCall optimistic-lock fence.
func (request EventCallProjectionRequest) ExpectedCallVersion() uint { return request.expectedCallVersion }

func (request EventCallProjectionRequest) valid() bool {
	return request.scheduledCallID != uuid.Nil && request.expectedEventVersion > 0 && request.expectedCallVersion > 0 &&
		eventCallProjectionActionKnown(request.action) && eventCallProjectionBoundaryValid(request.boundary)
}

// EventCallDueRequest is the caller-owned, immutable due-evaluation fence.
type EventCallDueRequest struct {
	scheduledCallID      uuid.UUID
	now                  time.Time
	envelope             EventCallDueEnvelopeInput
	expectedEventVersion uint
	expectedCallVersion  uint
}

// NewEventCallDueRequest validates the exact UTC evaluation time and durable envelope identity.
func NewEventCallDueRequest(scheduledCallID uuid.UUID, now time.Time, envelope EventCallDueEnvelopeInput, expectedEventVersion, expectedCallVersion uint) (EventCallDueRequest, error) {
	request := EventCallDueRequest{
		scheduledCallID:      scheduledCallID,
		now:                  now,
		envelope:             EventCallDueEnvelopeInput{EventID: envelope.EventID, OccurredAt: envelope.OccurredAt, CorrelationID: envelope.CorrelationID},
		expectedEventVersion: expectedEventVersion,
		expectedCallVersion:  expectedCallVersion,
	}
	if !request.valid() {
		return EventCallDueRequest{}, ErrEventInvalid
	}
	return request, nil
}

// ScheduledCallID returns the immutable linked call identity.
func (request EventCallDueRequest) ScheduledCallID() uuid.UUID { return request.scheduledCallID }

// Now returns the exact UTC due-evaluation instant.
func (request EventCallDueRequest) Now() time.Time { return request.now }

// Envelope returns a value copy of the frozen outbox envelope input.
func (request EventCallDueRequest) Envelope() EventCallDueEnvelopeInput {
	return EventCallDueEnvelopeInput{EventID: request.envelope.EventID, OccurredAt: request.envelope.OccurredAt, CorrelationID: request.envelope.CorrelationID}
}

// ExpectedEventVersion returns the exact CalendarEvent optimistic-lock fence.
func (request EventCallDueRequest) ExpectedEventVersion() uint { return request.expectedEventVersion }

// ExpectedCallVersion returns the exact ScheduledCall optimistic-lock fence.
func (request EventCallDueRequest) ExpectedCallVersion() uint { return request.expectedCallVersion }

func (request EventCallDueRequest) valid() bool {
	return request.scheduledCallID != uuid.Nil && request.expectedEventVersion > 0 && request.expectedCallVersion > 0 &&
		eventCallProjectionExactUTCSecond(request.now) && request.envelope.EventID != uuid.Nil &&
		eventCallProjectionExactUTCSecond(request.envelope.OccurredAt) && request.envelope.OccurredAt.Equal(request.now) &&
		eventCallDueCorrelationKind(request.envelope.CorrelationID) == ""
}

// EventCallTransaction is valid only for its callback lifetime.
type EventCallTransaction struct {
	mu      sync.Mutex
	backend EventCallTransactionBackend
	active  bool
}

// RunEventCallTransaction creates one lifetime-guarded bridge wrapper and always closes it on return or panic.
func RunEventCallTransaction(backend EventCallTransactionBackend, callback func(*EventCallTransaction) error) error {
	if eventCallTypedNil(backend) || callback == nil {
		return ErrEventInvalid
	}
	transaction := &EventCallTransaction{backend: backend, active: true}
	defer transaction.close()
	return callback(transaction)
}

// ResolveEventCallProjection obtains an isolated, locked durable policy input.
func (transaction *EventCallTransaction) ResolveEventCallProjection(ctx context.Context, request EventCallProjectionRequest) (EventCallProjectionInput, error) {
	if !request.valid() {
		return EventCallProjectionInput{}, ErrEventInvalid
	}
	backend, err := transaction.lockBackend()
	if err != nil {
		return EventCallProjectionInput{}, err
	}
	defer transaction.unlockBackend()
	input, err := backend.ResolveEventCallProjection(ctx, request)
	if err != nil {
		return EventCallProjectionInput{}, err
	}
	if input.Action != request.Action() || input.Boundary != request.Boundary() || input.Event.ScheduledCallID != request.ScheduledCallID() || input.Target.ScheduledCallID != request.ScheduledCallID() {
		return EventCallProjectionInput{}, ErrEventInvalid
	}
	return cloneEventCallProjectionInput(input), nil
}

// PersistEventCallProjection writes only the opaque policy snapshot behind the supplied fences.
func (transaction *EventCallTransaction) PersistEventCallProjection(ctx context.Context, request EventCallProjectionRequest, projection EventCallProjection) (EventCallProjectionSnapshot, error) {
	if !request.valid() {
		return EventCallProjectionSnapshot{}, ErrEventInvalid
	}
	backend, err := transaction.lockBackend()
	if err != nil {
		return EventCallProjectionSnapshot{}, err
	}
	defer transaction.unlockBackend()
	want := projection.Snapshot()
	snapshot, err := backend.PersistEventCallProjection(ctx, request, projection)
	if err != nil {
		return EventCallProjectionSnapshot{}, err
	}
	if !reflect.DeepEqual(snapshot, want) {
		return EventCallProjectionSnapshot{}, ErrEventInvalid
	}
	return cloneEventCallProjectionSnapshot(snapshot), nil
}

// ResolveEventCallDue obtains an isolated, locked durable due-policy input.
func (transaction *EventCallTransaction) ResolveEventCallDue(ctx context.Context, request EventCallDueRequest) (EventCallDueInput, error) {
	if !request.valid() {
		return EventCallDueInput{}, ErrEventInvalid
	}
	backend, err := transaction.lockBackend()
	if err != nil {
		return EventCallDueInput{}, err
	}
	defer transaction.unlockBackend()
	input, err := backend.ResolveEventCallDue(ctx, request)
	if err != nil {
		return EventCallDueInput{}, err
	}
	if input.Call.ScheduledCallID != request.ScheduledCallID() || !input.Now.Equal(request.Now()) {
		return EventCallDueInput{}, ErrEventInvalid
	}
	if !input.Call.Active || input.Call.Deleted {
		if input.Projection != nil || !eventCallDueZeroEnvelope(input.Envelope) {
			return EventCallDueInput{}, ErrEventInvalid
		}
	} else if input.Envelope != request.Envelope() {
		return EventCallDueInput{}, ErrEventInvalid
	}
	return cloneEventCallDueInput(input), nil
}

// PersistEventCallDue writes only the opaque due decision and its optional outbox descriptor.
func (transaction *EventCallTransaction) PersistEventCallDue(ctx context.Context, request EventCallDueRequest, decision EventCallDueDecision) (EventCallDueSnapshot, error) {
	if !request.valid() {
		return EventCallDueSnapshot{}, ErrEventInvalid
	}
	backend, err := transaction.lockBackend()
	if err != nil {
		return EventCallDueSnapshot{}, err
	}
	defer transaction.unlockBackend()
	want := decision.Snapshot()
	snapshot, err := backend.PersistEventCallDue(ctx, request, decision)
	if err != nil {
		return EventCallDueSnapshot{}, err
	}
	if !reflect.DeepEqual(snapshot, want) {
		return EventCallDueSnapshot{}, ErrEventInvalid
	}
	return cloneEventCallDueSnapshot(snapshot), nil
}

func (transaction *EventCallTransaction) lockBackend() (EventCallTransactionBackend, error) {
	if transaction == nil {
		return nil, ErrEventTransactionClosed
	}
	transaction.mu.Lock()
	if !transaction.active || eventCallTypedNil(transaction.backend) {
		transaction.mu.Unlock()
		return nil, ErrEventTransactionClosed
	}
	return transaction.backend, nil
}

func (transaction *EventCallTransaction) unlockBackend() { transaction.mu.Unlock() }

func (transaction *EventCallTransaction) close() {
	transaction.mu.Lock()
	defer transaction.mu.Unlock()
	transaction.active = false
	transaction.backend = nil
}

func eventCallTypedNil(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

func cloneEventCallProjectionInput(input EventCallProjectionInput) EventCallProjectionInput {
	clone := input
	clone.Event.Exdates = append([]string(nil), input.Event.Exdates...)
	clone.Event.ReminderMinutes = cloneEventCallProjectionInt(input.Event.ReminderMinutes)
	clone.Attendees = make([]EventCallProjectionAttendee, len(input.Attendees))
	for index, attendee := range input.Attendees {
		clone.Attendees[index] = attendee
		clone.Attendees[index].UserID = eventEntUUIDPointer(attendee.UserID)
	}
	clone.Exceptions = make([]EventCallProjectionException, len(input.Exceptions))
	for index, exception := range input.Exceptions {
		clone.Exceptions[index] = exception
		if exception.Temporal != nil {
			temporal := *exception.Temporal
			clone.Exceptions[index].Temporal = &temporal
		}
	}
	return clone
}

func cloneEventCallDueInput(input EventCallDueInput) EventCallDueInput {
	clone := input
	clone.Envelope = EventCallDueEnvelopeInput{EventID: input.Envelope.EventID, OccurredAt: input.Envelope.OccurredAt, CorrelationID: input.Envelope.CorrelationID}
	if input.Projection != nil {
		projection := cloneEventCallDueProjectionInput(*input.Projection)
		clone.Projection = &projection
	}
	return clone
}
