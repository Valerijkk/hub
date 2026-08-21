package calendar

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"reflect"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/platform/events"
)

const maxMutationReceiptResponseBytes = 1 << 20

// EventRepository задаёт чистый доменный порт канонических календарных событий.
// Авторитетное actor-to-organization разрешение допускается только через callback транзакции.
type EventRepository interface {
	WithEventTransaction(ctx context.Context, callback func(*EventTransaction) error) error
}

// EventTransactionBackend реализуется адаптером хранения и никогда не передаётся сервису напрямую.
type EventTransactionBackend interface {
	ResolveActorOrganization(ctx context.Context, actorID uuid.UUID) (uuid.UUID, error)
	ResolveCalendarEventCreateAuthorization(ctx context.Context, scope EventMutationScope, request CalendarEventCreateAuthorizationRequest) (CalendarEventCreateAuthorization, error)
	ReserveMutationReceipt(ctx context.Context, scope EventMutationScope, reservation MutationReceiptReservation) (MutationReceiptReservationResult, error)
	CompleteMutationReceipt(ctx context.Context, scope EventMutationScope, completion MutationReceiptCompletion) error
	PersistCalendarEvent(ctx context.Context, scope EventMutationScope, command CalendarEventCommand) (CalendarEventPersistenceResult, error)
	PersistCalendarEventAttendee(ctx context.Context, scope EventMutationScope, command CalendarEventAttendeeCommand) (CalendarEventAttendeePersistenceResult, error)
	PersistCalendarEventException(ctx context.Context, scope EventMutationScope, command CalendarEventExceptionCommand) (CalendarEventExceptionPersistenceResult, error)
	AppendOutbox(ctx context.Context, scope EventMutationScope, descriptor EventOutboxDescriptor) error
}

// CalendarEventUpdateAuthorizationRequest carries only the event identity and,
// when supplied, the untrusted attendee replacement candidates. The transaction
// backend is the sole authority for every current event and attendee value.
type CalendarEventUpdateAuthorizationRequest struct {
	EventID   uuid.UUID
	Attendees []CalendarEventCreateAttendeeCandidate
}

// CalendarEventUpdateTransactionBackend is an opt-in update port. Keeping it
// separate preserves the existing create/RSVP backend ABI until an adapter owns
// the update persistence implementation.
type CalendarEventUpdateTransactionBackend interface {
	ResolveCalendarEventUpdate(context.Context, EventMutationScope, CalendarEventUpdateAuthorizationRequest) (CalendarEventUpdateAuthorization, error)
	PersistCalendarEventUpdate(context.Context, EventMutationScope, CalendarEventUpdatePersistenceCommand) (CalendarEventUpdatePersistenceResult, error)
}

// CalendarEventOccurrenceCancelAuthorizationRequest carries only the master
// identity and canonical occurrence key. The optional backend owns every
// current master, attendee and existing-exception value.
type CalendarEventOccurrenceCancelAuthorizationRequest struct {
	EventID      uuid.UUID
	RecurrenceID string
}

// CalendarEventOccurrenceCancelAttendeeSnapshot is the minimal authoritative
// attendee state required to form the content-free cancellation recipient set.
type CalendarEventOccurrenceCancelAttendeeSnapshot struct {
	AttendeeID uuid.UUID
	EventID    uuid.UUID
	UserID     *uuid.UUID
	Status     RSVPStatus
	Version    uint
	RemovedAt  *time.Time
}

// CalendarEventOccurrenceCancelAuthorization is resolved in the same scoped
// transaction as the eventual compare-and-swap persistence boundary.
type CalendarEventOccurrenceCancelAuthorization struct {
	Current    CalendarEventCommandInput
	Attendees  []CalendarEventOccurrenceCancelAttendeeSnapshot
	Exception  *CalendarEventExceptionCommandInput
}

// CalendarEventOccurrenceCancelTransactionBackend is deliberately optional:
// it preserves every existing repository adapter ABI until its one atomic
// occurrence-cancellation implementation is available.
type CalendarEventOccurrenceCancelTransactionBackend interface {
	ResolveCalendarEventOccurrenceCancel(context.Context, EventMutationScope, CalendarEventOccurrenceCancelAuthorizationRequest) (CalendarEventOccurrenceCancelAuthorization, error)
	PersistCalendarEventOccurrenceCancel(context.Context, EventMutationScope, CalendarEventOccurrenceCancelPersistenceCommand) (CalendarEventOccurrenceCancelPersistenceResult, error)
}

// CalendarEventSeriesRSVPAuthorizationRequest carries only the authoritative
// event identity. The optional transaction backend resolves the active,
// actor-bound attendee inside the same transaction.
type CalendarEventSeriesRSVPAuthorizationRequest struct {
	EventID uuid.UUID
}

// CalendarEventSeriesRSVPAttendeeSnapshot is the complete durable RSVP state
// for one active attendee. Pointer fields are always copied at the port seam.
type CalendarEventSeriesRSVPAttendeeSnapshot struct {
	AttendeeID  uuid.UUID
	EventID     uuid.UUID
	UserID      *uuid.UUID
	Email       string
	DisplayName string
	Role        CalendarEventAttendeeRole
	Status      RSVPStatus
	RespondedAt *time.Time
	Version     uint
	RemovedAt   *time.Time
}

// CalendarEventSeriesRSVPAuthorization is the authoritative series state
// resolved before one RSVP compare-and-swap write.
type CalendarEventSeriesRSVPAuthorization struct {
	Current  CalendarEventCommandInput
	Attendee CalendarEventSeriesRSVPAttendeeSnapshot
}

// CalendarEventSeriesRSVPReplyDeliveryDescriptor is a durable, content-free
// reply delivery descriptor. It never carries iTIP bytes or provider state.
type CalendarEventSeriesRSVPReplyDeliveryDescriptor struct {
	AttendeeID      uuid.UUID
	EventID         uuid.UUID
	EventSequence   uint
	AttendeeVersion uint
	Status          RSVPStatus
}

// CalendarEventSeriesRSVPPersistenceCommand is the immutable all-or-nothing
// series RSVP write passed only to an adapter that opted into this port.
type CalendarEventSeriesRSVPPersistenceCommand struct {
	attendee        CalendarEventSeriesRSVPAttendeeSnapshot
	expectedVersion uint
	eventSequence   uint
	outbox          EventOutboxDescriptor
	reply           *CalendarEventSeriesRSVPReplyDeliveryDescriptor
}

// NewCalendarEventSeriesRSVPPersistenceCommand snapshots a single valid RSVP
// compare-and-swap mutation and its optional exact reply descriptor.
func NewCalendarEventSeriesRSVPPersistenceCommand(attendee CalendarEventSeriesRSVPAttendeeSnapshot, expectedVersion uint, eventSequence uint, outbox EventOutboxDescriptor, reply *CalendarEventSeriesRSVPReplyDeliveryDescriptor) (CalendarEventSeriesRSVPPersistenceCommand, error) {
	if expectedVersion == 0 || expectedVersion == ^uint(0) || !validCalendarEventSeriesRSVPAttendee(attendee) || attendee.Version != expectedVersion+1 || !outbox.valid() || !validCalendarEventSeriesRSVPReply(reply, attendee, eventSequence) {
		return CalendarEventSeriesRSVPPersistenceCommand{}, ErrEventInvalid
	}
	return CalendarEventSeriesRSVPPersistenceCommand{attendee: cloneCalendarEventSeriesRSVPAttendeeSnapshot(attendee), expectedVersion: expectedVersion, eventSequence: eventSequence, outbox: outbox, reply: cloneCalendarEventSeriesRSVPReplyDeliveryDescriptor(reply)}, nil
}

// Attendee returns an isolated authoritative attendee snapshot.
func (command CalendarEventSeriesRSVPPersistenceCommand) Attendee() CalendarEventSeriesRSVPAttendeeSnapshot {
	return cloneCalendarEventSeriesRSVPAttendeeSnapshot(command.attendee)
}

// ExpectedVersion returns the attendee version required by the compare-and-swap.
func (command CalendarEventSeriesRSVPPersistenceCommand) ExpectedVersion() uint { return command.expectedVersion }

// EventSequence returns the explicit authoritative event sequence for this RSVP.
func (command CalendarEventSeriesRSVPPersistenceCommand) EventSequence() uint { return command.eventSequence }

// Outbox returns the complete immutable outbox descriptor for this mutation.
func (command CalendarEventSeriesRSVPPersistenceCommand) Outbox() EventOutboxDescriptor { return command.outbox }

// Reply returns a value copy of the optional durable reply descriptor.
func (command CalendarEventSeriesRSVPPersistenceCommand) Reply() (CalendarEventSeriesRSVPReplyDeliveryDescriptor, bool) {
	if command.reply == nil {
		return CalendarEventSeriesRSVPReplyDeliveryDescriptor{}, false
	}
	return *command.reply, true
}

func (command CalendarEventSeriesRSVPPersistenceCommand) valid() bool {
	_, err := NewCalendarEventSeriesRSVPPersistenceCommand(command.attendee, command.expectedVersion, command.eventSequence, command.outbox, command.reply)
	return err == nil
}

// CalendarEventSeriesRSVPPersistenceResult is the immutable exact persistence
// echo returned by the optional series RSVP transaction backend.
type CalendarEventSeriesRSVPPersistenceResult struct {
	attendee      CalendarEventSeriesRSVPAttendeeSnapshot
	eventSequence uint
	outbox        EventOutboxDescriptor
	reply         *CalendarEventSeriesRSVPReplyDeliveryDescriptor
}

// NewCalendarEventSeriesRSVPPersistenceResult snapshots a valid durable RSVP
// persistence echo. The transaction rejects an echo that differs from its command.
func NewCalendarEventSeriesRSVPPersistenceResult(attendee CalendarEventSeriesRSVPAttendeeSnapshot, eventSequence uint, outbox EventOutboxDescriptor, reply *CalendarEventSeriesRSVPReplyDeliveryDescriptor) (CalendarEventSeriesRSVPPersistenceResult, error) {
	if !validCalendarEventSeriesRSVPAttendee(attendee) || !outbox.valid() || !validCalendarEventSeriesRSVPReply(reply, attendee, eventSequence) {
		return CalendarEventSeriesRSVPPersistenceResult{}, ErrEventInvalid
	}
	return CalendarEventSeriesRSVPPersistenceResult{attendee: cloneCalendarEventSeriesRSVPAttendeeSnapshot(attendee), eventSequence: eventSequence, outbox: outbox, reply: cloneCalendarEventSeriesRSVPReplyDeliveryDescriptor(reply)}, nil
}

// Attendee returns an isolated durable attendee snapshot.
func (result CalendarEventSeriesRSVPPersistenceResult) Attendee() CalendarEventSeriesRSVPAttendeeSnapshot {
	return cloneCalendarEventSeriesRSVPAttendeeSnapshot(result.attendee)
}

// EventSequence returns the explicit durable event sequence for this RSVP.
func (result CalendarEventSeriesRSVPPersistenceResult) EventSequence() uint { return result.eventSequence }

// Outbox returns the complete immutable durable outbox descriptor.
func (result CalendarEventSeriesRSVPPersistenceResult) Outbox() EventOutboxDescriptor { return result.outbox }

// Reply returns a value copy of the optional durable reply descriptor.
func (result CalendarEventSeriesRSVPPersistenceResult) Reply() (CalendarEventSeriesRSVPReplyDeliveryDescriptor, bool) {
	if result.reply == nil {
		return CalendarEventSeriesRSVPReplyDeliveryDescriptor{}, false
	}
	return *result.reply, true
}

func (result CalendarEventSeriesRSVPPersistenceResult) valid() bool {
	_, err := NewCalendarEventSeriesRSVPPersistenceResult(result.attendee, result.eventSequence, result.outbox, result.reply)
	return err == nil
}

// CalendarEventSeriesRSVPTransactionBackend is deliberately optional so every
// existing adapter keeps its create, update and occurrence-cancel ABI unchanged.
type CalendarEventSeriesRSVPTransactionBackend interface {
	ResolveCalendarEventSeriesRSVP(context.Context, EventMutationScope, CalendarEventSeriesRSVPAuthorizationRequest) (CalendarEventSeriesRSVPAuthorization, error)
	PersistCalendarEventSeriesRSVP(context.Context, EventMutationScope, CalendarEventSeriesRSVPPersistenceCommand) (CalendarEventSeriesRSVPPersistenceResult, error)
}

// CalendarEventOccurrenceRSVPAuthorizationRequest carries only the master
// event identity and canonical recurrence key. The optional backend resolves
// every durable series, occurrence and exception value in its transaction.
type CalendarEventOccurrenceRSVPAuthorizationRequest struct {
	EventID      uuid.UUID
	RecurrenceID string
}

// CalendarEventOccurrenceRSVPAttendeeSnapshot is the active RSVP state for
// one materialized occurrence. It deliberately keeps the immutable series
// identity alongside the occurrence-local recurrence key and version.
type CalendarEventOccurrenceRSVPAttendeeSnapshot struct {
	AttendeeID   uuid.UUID
	EventID      uuid.UUID
	UserID       *uuid.UUID
	Email        string
	DisplayName  string
	Role         CalendarEventAttendeeRole
	Status       RSVPStatus
	RecurrenceID string
	RespondedAt  *time.Time
	Version      uint
	RemovedAt    *time.Time
}

// CalendarEventOccurrenceRSVPAuthorization is the authoritative snapshot
// resolved before one occurrence-local RSVP compare-and-swap write.
type CalendarEventOccurrenceRSVPAuthorization struct {
	Current            CalendarEventCommandInput
	SeriesAttendee     CalendarEventSeriesRSVPAttendeeSnapshot
	OccurrenceAttendee *CalendarEventOccurrenceRSVPAttendeeSnapshot
	Exception          *CalendarEventExceptionCommandInput
}

// CalendarEventOccurrenceRSVPReplyDeliveryDescriptor is a content-free
// durable REPLY delivery descriptor for exactly one occurrence attendee.
type CalendarEventOccurrenceRSVPReplyDeliveryDescriptor struct {
	AttendeeID      uuid.UUID
	EventID         uuid.UUID
	EventSequence   uint
	AttendeeVersion uint
	RecurrenceID    string
	Status          RSVPStatus
}

// CalendarEventOccurrenceRSVPPersistenceCommand is an immutable atomic
// occurrence RSVP write. Its fields remain opaque outside this package.
type CalendarEventOccurrenceRSVPPersistenceCommand struct {
	seriesAttendee  CalendarEventSeriesRSVPAttendeeSnapshot
	prior           *CalendarEventOccurrenceRSVPAttendeeSnapshot
	attendee        CalendarEventOccurrenceRSVPAttendeeSnapshot
	expectedVersion uint
	eventSequence   uint
	outbox          EventOutboxDescriptor
	reply           *CalendarEventOccurrenceRSVPReplyDeliveryDescriptor
}

// NewCalendarEventOccurrenceRSVPPersistenceCommand snapshots either a fresh
// inherited materialization or an existing occurrence override CAS mutation.
func NewCalendarEventOccurrenceRSVPPersistenceCommand(series CalendarEventSeriesRSVPAttendeeSnapshot, prior *CalendarEventOccurrenceRSVPAttendeeSnapshot, attendee CalendarEventOccurrenceRSVPAttendeeSnapshot, expectedVersion uint, eventSequence uint, outbox EventOutboxDescriptor, reply *CalendarEventOccurrenceRSVPReplyDeliveryDescriptor) (CalendarEventOccurrenceRSVPPersistenceCommand, error) {
	if !validCalendarEventOccurrenceRSVPPersistence(series, prior, attendee, expectedVersion, eventSequence, outbox, reply) {
		return CalendarEventOccurrenceRSVPPersistenceCommand{}, ErrEventInvalid
	}
	return CalendarEventOccurrenceRSVPPersistenceCommand{
		seriesAttendee:  cloneCalendarEventSeriesRSVPAttendeeSnapshot(series),
		prior:           cloneCalendarEventOccurrenceRSVPAttendeeSnapshotPointer(prior),
		attendee:        cloneCalendarEventOccurrenceRSVPAttendeeSnapshot(attendee),
		expectedVersion: expectedVersion,
		eventSequence:   eventSequence,
		outbox:          outbox,
		reply:           cloneCalendarEventOccurrenceRSVPReplyDeliveryDescriptor(reply),
	}, nil
}

// SeriesAttendee returns an isolated immutable series attendee snapshot.
func (command CalendarEventOccurrenceRSVPPersistenceCommand) SeriesAttendee() CalendarEventSeriesRSVPAttendeeSnapshot {
	return cloneCalendarEventSeriesRSVPAttendeeSnapshot(command.seriesAttendee)
}

// Prior returns the isolated materialized occurrence snapshot when an
// occurrence override already existed before this compare-and-swap write.
func (command CalendarEventOccurrenceRSVPPersistenceCommand) Prior() (CalendarEventOccurrenceRSVPAttendeeSnapshot, bool) {
	if command.prior == nil {
		return CalendarEventOccurrenceRSVPAttendeeSnapshot{}, false
	}
	return cloneCalendarEventOccurrenceRSVPAttendeeSnapshot(*command.prior), true
}

// Attendee returns the isolated updated occurrence attendee snapshot.
func (command CalendarEventOccurrenceRSVPPersistenceCommand) Attendee() CalendarEventOccurrenceRSVPAttendeeSnapshot {
	return cloneCalendarEventOccurrenceRSVPAttendeeSnapshot(command.attendee)
}

// ExpectedVersion returns the version required by the inherited or override CAS.
func (command CalendarEventOccurrenceRSVPPersistenceCommand) ExpectedVersion() uint {
	return command.expectedVersion
}

// EventSequence returns the authoritative event sequence, including zero.
func (command CalendarEventOccurrenceRSVPPersistenceCommand) EventSequence() uint {
	return command.eventSequence
}

// Outbox returns the complete immutable outbox descriptor for this mutation.
func (command CalendarEventOccurrenceRSVPPersistenceCommand) Outbox() EventOutboxDescriptor {
	return command.outbox
}

// Reply returns a value copy of the optional durable occurrence REPLY descriptor.
func (command CalendarEventOccurrenceRSVPPersistenceCommand) Reply() (CalendarEventOccurrenceRSVPReplyDeliveryDescriptor, bool) {
	if command.reply == nil {
		return CalendarEventOccurrenceRSVPReplyDeliveryDescriptor{}, false
	}
	return *command.reply, true
}

func (command CalendarEventOccurrenceRSVPPersistenceCommand) valid() bool {
	return validCalendarEventOccurrenceRSVPPersistence(command.seriesAttendee, command.prior, command.attendee, command.expectedVersion, command.eventSequence, command.outbox, command.reply)
}

// CalendarEventOccurrenceRSVPPersistenceResult is an opaque exact durable
// echo returned by an occurrence RSVP backend.
type CalendarEventOccurrenceRSVPPersistenceResult struct {
	attendee      CalendarEventOccurrenceRSVPAttendeeSnapshot
	eventSequence uint
	outbox        EventOutboxDescriptor
	reply         *CalendarEventOccurrenceRSVPReplyDeliveryDescriptor
}

// NewCalendarEventOccurrenceRSVPPersistenceResult snapshots a valid durable
// occurrence RSVP echo. Transaction code additionally verifies exact command equality.
func NewCalendarEventOccurrenceRSVPPersistenceResult(attendee CalendarEventOccurrenceRSVPAttendeeSnapshot, eventSequence uint, outbox EventOutboxDescriptor, reply *CalendarEventOccurrenceRSVPReplyDeliveryDescriptor) (CalendarEventOccurrenceRSVPPersistenceResult, error) {
	if !validCalendarEventOccurrenceRSVPAttendee(attendee) || !outbox.valid() || !validCalendarEventOccurrenceRSVPReply(reply, attendee, eventSequence) {
		return CalendarEventOccurrenceRSVPPersistenceResult{}, ErrEventInvalid
	}
	return CalendarEventOccurrenceRSVPPersistenceResult{
		attendee:      cloneCalendarEventOccurrenceRSVPAttendeeSnapshot(attendee),
		eventSequence: eventSequence,
		outbox:        outbox,
		reply:         cloneCalendarEventOccurrenceRSVPReplyDeliveryDescriptor(reply),
	}, nil
}

// Attendee returns the isolated durable occurrence attendee snapshot.
func (result CalendarEventOccurrenceRSVPPersistenceResult) Attendee() CalendarEventOccurrenceRSVPAttendeeSnapshot {
	return cloneCalendarEventOccurrenceRSVPAttendeeSnapshot(result.attendee)
}

// EventSequence returns the durable authoritative event sequence, including zero.
func (result CalendarEventOccurrenceRSVPPersistenceResult) EventSequence() uint {
	return result.eventSequence
}

// Outbox returns the complete immutable durable outbox descriptor.
func (result CalendarEventOccurrenceRSVPPersistenceResult) Outbox() EventOutboxDescriptor {
	return result.outbox
}

// Reply returns a value copy of the optional durable occurrence REPLY descriptor.
func (result CalendarEventOccurrenceRSVPPersistenceResult) Reply() (CalendarEventOccurrenceRSVPReplyDeliveryDescriptor, bool) {
	if result.reply == nil {
		return CalendarEventOccurrenceRSVPReplyDeliveryDescriptor{}, false
	}
	return *result.reply, true
}

func (result CalendarEventOccurrenceRSVPPersistenceResult) valid() bool {
	_, err := NewCalendarEventOccurrenceRSVPPersistenceResult(result.attendee, result.eventSequence, result.outbox, result.reply)
	return err == nil
}

// CalendarEventOccurrenceRSVPTransactionBackend is deliberately optional so
// existing repository adapters and the accepted series RSVP ABI stay intact.
type CalendarEventOccurrenceRSVPTransactionBackend interface {
	ResolveCalendarEventOccurrenceRSVP(context.Context, EventMutationScope, CalendarEventOccurrenceRSVPAuthorizationRequest) (CalendarEventOccurrenceRSVPAuthorization, error)
	PersistCalendarEventOccurrenceRSVP(context.Context, EventMutationScope, CalendarEventOccurrenceRSVPPersistenceCommand) (CalendarEventOccurrenceRSVPPersistenceResult, error)
}

// CalendarEventCreateAttendeeCandidate is an untrusted create-time attendee
// reference. The transaction resolver owns every identity decision.
type CalendarEventCreateAttendeeCandidate struct {
	UserID      *uuid.UUID
	Email       string
	DisplayName *string
	Role        CalendarEventAttendeeRole
}

// CalendarEventCreateAuthorizationRequest is the immutable candidate set that
// the transaction backend resolves before a new calendar event is persisted.
type CalendarEventCreateAuthorizationRequest struct {
	Organizer       CalendarEventCreateAttendeeCandidate
	Attendees       []CalendarEventCreateAttendeeCandidate
	Title           string
	Description     string
	Location        string
	Temporal        CalendarEvent
	RRULE           string
	Exdates         []string
	ReminderMinutes *int
	ScheduledCallID *uuid.UUID
}

// CalendarEventCreateResolvedOrganizer is the authoritative organizer identity
// supplied by the in-transaction resolver.
type CalendarEventCreateResolvedOrganizer struct {
	UserID      uuid.UUID
	Email       string
	DisplayName *string
}

// CalendarEventCreateResolvedAttendee is an authoritative attendee identity
// supplied by the in-transaction resolver.
type CalendarEventCreateResolvedAttendee struct {
	UserID      *uuid.UUID
	Email       string
	DisplayName *string
	Role        CalendarEventAttendeeRole
}

// CalendarEventCreateAuthorization contains only resolved identities and the
// authoritative scheduled-call relation for a fresh create mutation.
type CalendarEventCreateAuthorization struct {
	Organizer       CalendarEventCreateResolvedOrganizer
	Attendees       []CalendarEventCreateResolvedAttendee
	ScheduledCallID *uuid.UUID
}

// EventTransaction ограничивает все мутации временем callback и не позволяет захваченному значению работать после него.
type EventTransaction struct {
	mu      sync.Mutex
	backend EventTransactionBackend
	active  bool
}

// RunEventTransaction создаёт lifetime-guarded transaction для реализации репозитория и внешних unit fake.
func RunEventTransaction(backend EventTransactionBackend, callback func(*EventTransaction) error) error {
	if backend == nil || callback == nil {
		return ErrEventInvalid
	}
	transaction := &EventTransaction{backend: backend, active: true}
	defer transaction.close()
	return callback(transaction)
}

func (transaction *EventTransaction) close() {
	transaction.mu.Lock()
	defer transaction.mu.Unlock()
	transaction.active = false
	transaction.backend = nil
}

// EventMutationScope связывает каждую мутацию с actor, revalidated organization и operation одной транзакции.
type EventMutationScope struct {
	actorID          uuid.UUID
	organizationID   uuid.UUID
	receiptOperation MutationReceiptOperation
	eventOperation   EventMutationOperation
	transaction      *EventTransaction
}

// ActorID возвращает аутентифицированного автора мутации.
func (scope EventMutationScope) ActorID() uuid.UUID {
	return scope.actorID
}

// OrganizationID возвращает организацию, разрешённую внутри той же транзакции.
func (scope EventMutationScope) OrganizationID() uuid.UUID {
	return scope.organizationID
}

// Operation возвращает разрешённую идемпотентную операцию мутации.
func (scope EventMutationScope) Operation() MutationReceiptOperation {
	return scope.receiptOperation
}

// EventOperation returns the canonical calendar mutation operation, including
// operations that intentionally do not use a mutation receipt.
func (scope EventMutationScope) EventOperation() EventMutationOperation {
	return scope.eventOperation
}

// ResolveMutationScope разрешает и повторно проверяет actor-to-organization только внутри активной транзакции.
func (transaction *EventTransaction) ResolveMutationScope(ctx context.Context, actorID uuid.UUID, operation MutationReceiptOperation) (EventMutationScope, error) {
	if transaction == nil {
		return EventMutationScope{}, ErrEventTransactionClosed
	}
	transaction.mu.Lock()
	defer transaction.mu.Unlock()
	if !transaction.active || transaction.backend == nil {
		return EventMutationScope{}, ErrEventTransactionClosed
	}
	eventOperation, found := eventMutationOperationForReceipt(operation)
	if actorID == uuid.Nil || !found {
		return EventMutationScope{}, ErrEventInvalid
	}
	organizationID, err := transaction.backend.ResolveActorOrganization(ctx, actorID)
	if err != nil || organizationID == uuid.Nil {
		return EventMutationScope{}, ErrEventForbidden
	}
	return EventMutationScope{
		actorID:          actorID,
		organizationID:   organizationID,
		receiptOperation: operation,
		eventOperation:   eventOperation,
		transaction:      transaction,
	}, nil
}

// ResolveCalendarEventUpdateScope resolves an update scope only for backends
// that explicitly support the optional update transaction port. The capability
// check precedes every backend call to avoid an observable fallback path.
func (transaction *EventTransaction) ResolveCalendarEventUpdateScope(ctx context.Context, actorID uuid.UUID) (EventMutationScope, error) {
	if transaction == nil {
		return EventMutationScope{}, ErrEventTransactionClosed
	}
	transaction.mu.Lock()
	defer transaction.mu.Unlock()
	if !transaction.active || transaction.backend == nil {
		return EventMutationScope{}, ErrEventTransactionClosed
	}
	if _, supported := transaction.backend.(CalendarEventUpdateTransactionBackend); !supported {
		return EventMutationScope{}, ErrEventInvalid
	}
	if actorID == uuid.Nil {
		return EventMutationScope{}, ErrEventInvalid
	}
	organizationID, err := transaction.backend.ResolveActorOrganization(ctx, actorID)
	if err != nil || organizationID == uuid.Nil {
		return EventMutationScope{}, ErrEventForbidden
	}
	return EventMutationScope{
		actorID:          actorID,
		organizationID:   organizationID,
		eventOperation:   EventMutationCalendarUpdate,
		transaction:      transaction,
	}, nil
}

// ResolveCalendarEventOccurrenceCancelScope first proves support for the
// optional port. In particular it makes no actor-organization backend call
// when a legacy backend lacks the occurrence-cancel capability.
func (transaction *EventTransaction) ResolveCalendarEventOccurrenceCancelScope(ctx context.Context, actorID uuid.UUID) (EventMutationScope, error) {
	if transaction == nil {
		return EventMutationScope{}, ErrEventTransactionClosed
	}
	transaction.mu.Lock()
	defer transaction.mu.Unlock()
	if !transaction.active || transaction.backend == nil {
		return EventMutationScope{}, ErrEventTransactionClosed
	}
	backend, supported := transaction.backend.(CalendarEventOccurrenceCancelTransactionBackend)
	if !supported || nilCalendarEventOccurrenceCancelTransactionBackend(backend) {
		return EventMutationScope{}, ErrEventInvalid
	}
	if actorID == uuid.Nil {
		return EventMutationScope{}, ErrEventInvalid
	}
	organizationID, err := transaction.backend.ResolveActorOrganization(ctx, actorID)
	if err != nil || organizationID == uuid.Nil {
		return EventMutationScope{}, ErrEventForbidden
	}
	return EventMutationScope{
		actorID:        actorID,
		organizationID: organizationID,
		eventOperation: EventMutationCalendarOccurrenceCancel,
		transaction:    transaction,
	}, nil
}

// ResolveCalendarEventSeriesRSVPScope proves optional RSVP-series support
// before resolving actor organization, preventing observable legacy fallback.
func (transaction *EventTransaction) ResolveCalendarEventSeriesRSVPScope(ctx context.Context, actorID uuid.UUID) (EventMutationScope, error) {
	if transaction == nil {
		return EventMutationScope{}, ErrEventTransactionClosed
	}
	transaction.mu.Lock()
	defer transaction.mu.Unlock()
	if !transaction.active || transaction.backend == nil {
		return EventMutationScope{}, ErrEventTransactionClosed
	}
	backend, supported := transaction.backend.(CalendarEventSeriesRSVPTransactionBackend)
	if !supported || nilCalendarEventSeriesRSVPTransactionBackend(backend) {
		return EventMutationScope{}, ErrEventInvalid
	}
	if actorID == uuid.Nil {
		return EventMutationScope{}, ErrEventInvalid
	}
	organizationID, err := transaction.backend.ResolveActorOrganization(ctx, actorID)
	if err != nil || organizationID == uuid.Nil {
		return EventMutationScope{}, ErrEventForbidden
	}
	return EventMutationScope{
		actorID:          actorID,
		organizationID:   organizationID,
		receiptOperation: MutationReceiptRSVPSeries,
		eventOperation:   EventMutationRSVPSeries,
		transaction:      transaction,
	}, nil
}

// ResolveCalendarEventOccurrenceRSVPScope proves support for the optional
// occurrence RSVP port before actor organization resolution. This prevents an
// observable legacy fallback and binds both receipt and event operations.
func (transaction *EventTransaction) ResolveCalendarEventOccurrenceRSVPScope(ctx context.Context, actorID uuid.UUID) (EventMutationScope, error) {
	if transaction == nil {
		return EventMutationScope{}, ErrEventTransactionClosed
	}
	transaction.mu.Lock()
	defer transaction.mu.Unlock()
	if !transaction.active || transaction.backend == nil {
		return EventMutationScope{}, ErrEventTransactionClosed
	}
	backend, supported := transaction.backend.(CalendarEventOccurrenceRSVPTransactionBackend)
	if !supported || nilCalendarEventOccurrenceRSVPTransactionBackend(backend) {
		return EventMutationScope{}, ErrEventInvalid
	}
	if actorID == uuid.Nil {
		return EventMutationScope{}, ErrEventInvalid
	}
	organizationID, err := transaction.backend.ResolveActorOrganization(ctx, actorID)
	if err != nil || organizationID == uuid.Nil {
		return EventMutationScope{}, ErrEventForbidden
	}
	return EventMutationScope{
		actorID:          actorID,
		organizationID:   organizationID,
		receiptOperation: MutationReceiptRSVPOccurrence,
		eventOperation:   EventMutationRSVPOccurrence,
		transaction:      transaction,
	}, nil
}

// ResolveCalendarEventCreateAuthorization resolves create identities only
// while the transaction and its scope are still live. Both the input and
// output are copied so neither callers nor backends can retain mutable aliases.
func (transaction *EventTransaction) ResolveCalendarEventCreateAuthorization(ctx context.Context, scope EventMutationScope, request CalendarEventCreateAuthorizationRequest) (CalendarEventCreateAuthorization, error) {
	if transaction == nil {
		return CalendarEventCreateAuthorization{}, ErrEventTransactionClosed
	}
	transaction.mu.Lock()
	defer transaction.mu.Unlock()
	if err := transaction.validateScope(scope); err != nil {
		return CalendarEventCreateAuthorization{}, err
	}
	resolved, err := transaction.backend.ResolveCalendarEventCreateAuthorization(ctx, scope, cloneCalendarEventCreateAuthorizationRequest(request))
	if err != nil {
		return CalendarEventCreateAuthorization{}, err
	}
	return cloneCalendarEventCreateAuthorization(resolved), nil
}

// ResolveCalendarEventUpdate resolves the authoritative update snapshot only
// inside an active update scope. Both sides of the optional port are copied.
func (transaction *EventTransaction) ResolveCalendarEventUpdate(ctx context.Context, scope EventMutationScope, request CalendarEventUpdateAuthorizationRequest) (CalendarEventUpdateAuthorization, error) {
	if transaction == nil {
		return CalendarEventUpdateAuthorization{}, ErrEventTransactionClosed
	}
	transaction.mu.Lock()
	defer transaction.mu.Unlock()
	if !transaction.active || transaction.backend == nil {
		return CalendarEventUpdateAuthorization{}, ErrEventTransactionClosed
	}
	backend, supported := transaction.backend.(CalendarEventUpdateTransactionBackend)
	if !supported {
		return CalendarEventUpdateAuthorization{}, ErrEventInvalid
	}
	if err := transaction.validateScope(scope); err != nil {
		return CalendarEventUpdateAuthorization{}, err
	}
	if scope.eventOperation != EventMutationCalendarUpdate || request.EventID == uuid.Nil {
		return CalendarEventUpdateAuthorization{}, ErrEventInvalid
	}
	resolved, err := backend.ResolveCalendarEventUpdate(ctx, scope, cloneCalendarEventUpdateAuthorizationRequest(request))
	if err != nil {
		return CalendarEventUpdateAuthorization{}, err
	}
	return cloneCalendarEventUpdateAuthorization(resolved), nil
}

// ResolveCalendarEventOccurrenceCancel resolves a frozen authoritative
// snapshot only during an active receipt-free occurrence-cancel scope.
func (transaction *EventTransaction) ResolveCalendarEventOccurrenceCancel(ctx context.Context, scope EventMutationScope, request CalendarEventOccurrenceCancelAuthorizationRequest) (CalendarEventOccurrenceCancelAuthorization, error) {
	if transaction == nil {
		return CalendarEventOccurrenceCancelAuthorization{}, ErrEventTransactionClosed
	}
	transaction.mu.Lock()
	defer transaction.mu.Unlock()
	if !transaction.active || transaction.backend == nil {
		return CalendarEventOccurrenceCancelAuthorization{}, ErrEventTransactionClosed
	}
	backend, supported := transaction.backend.(CalendarEventOccurrenceCancelTransactionBackend)
	if !supported || nilCalendarEventOccurrenceCancelTransactionBackend(backend) {
		return CalendarEventOccurrenceCancelAuthorization{}, ErrEventInvalid
	}
	if err := transaction.validateScope(scope); err != nil {
		return CalendarEventOccurrenceCancelAuthorization{}, err
	}
	// This seam may reject only fixed-shape identity and hostile byte ingress
	// before the owner/org-scoped resolve. Temporal canonicalization and nominal
	// membership deliberately wait for the authoritative master so a foreign
	// event cannot distinguish a malformed-but-ingress-valid recurrence key.
	if scope.eventOperation != EventMutationCalendarOccurrenceCancel || request.EventID == uuid.Nil || !validCalendarEventOccurrenceCancelIngress(request.RecurrenceID) {
		return CalendarEventOccurrenceCancelAuthorization{}, ErrEventInvalid
	}
	resolved, err := backend.ResolveCalendarEventOccurrenceCancel(ctx, scope, cloneCalendarEventOccurrenceCancelAuthorizationRequest(request))
	if err != nil {
		return CalendarEventOccurrenceCancelAuthorization{}, err
	}
	// Identity is intentionally sampled before cloning variable-size backend
	// material. This preserves the service's nondisclosing mask for forged or
	// hostile foreign snapshots without allocating their attacker-controlled
	// recurrence, attendee or exception values.
	if resolved.Current.EventID != request.EventID || resolved.Current.OrganizationID != scope.organizationID || resolved.Current.OwnerID != scope.actorID || resolved.Current.OrganizerID == nil || *resolved.Current.OrganizerID != scope.actorID {
		return CalendarEventOccurrenceCancelAuthorization{Current: CalendarEventCommandInput{
			EventID: resolved.Current.EventID, OrganizationID: resolved.Current.OrganizationID, OwnerID: resolved.Current.OwnerID, OrganizerID: copyUUID(resolved.Current.OrganizerID),
		}}, nil
	}
	// Only an identity-bound snapshot is allowed to consume variable-size
	// cloning work. The caps are adapter ingress guards; the service repeats the
	// checks before any policy traversal for defense in depth.
	if len(resolved.Current.Exdates) > calendarCreateMaxExdates || len(resolved.Attendees) > calendarCreateMaxAttendees {
		return CalendarEventOccurrenceCancelAuthorization{}, ErrEventInvalid
	}
	return cloneCalendarEventOccurrenceCancelAuthorization(resolved), nil
}

// ResolveCalendarEventSeriesRSVP resolves one authoritative attendee only
// inside a live, actor-bound RSVP-series scope. Foreign identity material is
// reduced to fixed-size identifiers before any variable-size cloning occurs.
func (transaction *EventTransaction) ResolveCalendarEventSeriesRSVP(ctx context.Context, scope EventMutationScope, request CalendarEventSeriesRSVPAuthorizationRequest) (CalendarEventSeriesRSVPAuthorization, error) {
	if transaction == nil {
		return CalendarEventSeriesRSVPAuthorization{}, ErrEventTransactionClosed
	}
	transaction.mu.Lock()
	defer transaction.mu.Unlock()
	if !transaction.active || transaction.backend == nil {
		return CalendarEventSeriesRSVPAuthorization{}, ErrEventTransactionClosed
	}
	backend, supported := transaction.backend.(CalendarEventSeriesRSVPTransactionBackend)
	if !supported || nilCalendarEventSeriesRSVPTransactionBackend(backend) {
		return CalendarEventSeriesRSVPAuthorization{}, ErrEventInvalid
	}
	if err := transaction.validateScope(scope); err != nil {
		return CalendarEventSeriesRSVPAuthorization{}, err
	}
	if scope.receiptOperation != MutationReceiptRSVPSeries || scope.eventOperation != EventMutationRSVPSeries || request.EventID == uuid.Nil {
		return CalendarEventSeriesRSVPAuthorization{}, ErrEventInvalid
	}
	resolved, err := backend.ResolveCalendarEventSeriesRSVP(ctx, scope, request)
	if err != nil {
		return CalendarEventSeriesRSVPAuthorization{}, err
	}
	if resolved.Current.EventID != request.EventID || resolved.Current.OrganizationID != scope.organizationID || resolved.Attendee.EventID != request.EventID || resolved.Attendee.UserID == nil || *resolved.Attendee.UserID != scope.actorID {
		return CalendarEventSeriesRSVPAuthorization{
			Current: CalendarEventCommandInput{
				EventID:        resolved.Current.EventID,
				OrganizationID: resolved.Current.OrganizationID,
			},
			Attendee: CalendarEventSeriesRSVPAttendeeSnapshot{
				AttendeeID: resolved.Attendee.AttendeeID,
				EventID:    resolved.Attendee.EventID,
				UserID:     copyUUID(resolved.Attendee.UserID),
			},
		}, nil
	}
	if !validCalendarEventSeriesRSVPAttendee(resolved.Attendee) {
		return CalendarEventSeriesRSVPAuthorization{}, ErrEventInvalid
	}
	return cloneCalendarEventSeriesRSVPAuthorization(resolved), nil
}

// ReserveMutationReceipt резервирует или воспроизводит квитанцию только в переданной transaction scope.

// ResolveCalendarEventOccurrenceRSVP resolves one authoritative series
// attendee plus an optional materialized occurrence attendee. Foreign
// identity is reduced to a deterministic zero mask before any backend data is
// cloned.
func (transaction *EventTransaction) ResolveCalendarEventOccurrenceRSVP(ctx context.Context, scope EventMutationScope, request CalendarEventOccurrenceRSVPAuthorizationRequest) (CalendarEventOccurrenceRSVPAuthorization, error) {
	if transaction == nil {
		return CalendarEventOccurrenceRSVPAuthorization{}, ErrEventTransactionClosed
	}
	transaction.mu.Lock()
	defer transaction.mu.Unlock()
	if !transaction.active || transaction.backend == nil {
		return CalendarEventOccurrenceRSVPAuthorization{}, ErrEventTransactionClosed
	}
	backend, supported := transaction.backend.(CalendarEventOccurrenceRSVPTransactionBackend)
	if !supported || nilCalendarEventOccurrenceRSVPTransactionBackend(backend) {
		return CalendarEventOccurrenceRSVPAuthorization{}, ErrEventInvalid
	}
	if err := transaction.validateScope(scope); err != nil {
		return CalendarEventOccurrenceRSVPAuthorization{}, err
	}
	if scope.receiptOperation != MutationReceiptRSVPOccurrence || scope.eventOperation != EventMutationRSVPOccurrence || request.EventID == uuid.Nil || !validCalendarEventOccurrenceCancelCanonicalKey(request.RecurrenceID) {
		return CalendarEventOccurrenceRSVPAuthorization{}, ErrEventInvalid
	}
	resolved, err := backend.ResolveCalendarEventOccurrenceRSVP(ctx, scope, cloneCalendarEventOccurrenceRSVPAuthorizationRequest(request))
	if err != nil {
		return CalendarEventOccurrenceRSVPAuthorization{}, err
	}
	if !calendarEventOccurrenceRSVPAuthorizationIdentityMatches(resolved, scope, request) {
		return calendarEventOccurrenceRSVPAuthorizationIdentityMask(), nil
	}
	if _, err := NewCalendarEventCommand(resolved.Current); err != nil || !validCalendarEventSeriesRSVPAttendee(resolved.SeriesAttendee) {
		return CalendarEventOccurrenceRSVPAuthorization{}, ErrEventInvalid
	}
	if resolved.OccurrenceAttendee != nil && (resolved.OccurrenceAttendee.RecurrenceID != request.RecurrenceID || !validCalendarEventOccurrenceRSVPAttendeeForSeries(resolved.SeriesAttendee, *resolved.OccurrenceAttendee)) {
		return CalendarEventOccurrenceRSVPAuthorization{}, ErrEventInvalid
	}
	if resolved.Exception != nil && !validCalendarEventOccurrenceRSVPException(*resolved.Exception, request) {
		return CalendarEventOccurrenceRSVPAuthorization{}, ErrEventInvalid
	}
	return cloneCalendarEventOccurrenceRSVPAuthorization(resolved), nil
}

// ReserveMutationReceipt reserves or replays a receipt only inside its scope.
func (transaction *EventTransaction) ReserveMutationReceipt(ctx context.Context, scope EventMutationScope, reservation MutationReceiptReservation) (MutationReceiptReservationResult, error) {
	if transaction == nil {
		return MutationReceiptReservationResult{}, ErrEventTransactionClosed
	}
	transaction.mu.Lock()
	defer transaction.mu.Unlock()
	if err := transaction.validateScope(scope); err != nil {
		return MutationReceiptReservationResult{}, err
	}
	if !validMutationReceiptOperation(scope.receiptOperation) {
		return MutationReceiptReservationResult{}, ErrEventInvalid
	}
	if !reservation.matches(scope) {
		return MutationReceiptReservationResult{}, ErrEventForbidden
	}
	result, err := transaction.backend.ReserveMutationReceipt(ctx, scope, reservation)
	if err != nil {
		return MutationReceiptReservationResult{}, err
	}
	if !result.valid() {
		return MutationReceiptReservationResult{}, ErrEventInvalid
	}
	return result, nil
}

// CompleteMutationReceipt завершает квитанцию только в переданной transaction scope.
func (transaction *EventTransaction) CompleteMutationReceipt(ctx context.Context, scope EventMutationScope, completion MutationReceiptCompletion) error {
	if transaction == nil {
		return ErrEventTransactionClosed
	}
	transaction.mu.Lock()
	defer transaction.mu.Unlock()
	if err := transaction.validateScope(scope); err != nil {
		return err
	}
	if !validMutationReceiptOperation(scope.receiptOperation) {
		return ErrEventInvalid
	}
	if !completion.valid() {
		return ErrEventInvalid
	}
	return transaction.backend.CompleteMutationReceipt(ctx, scope, completion)
}

// PersistCalendarEvent сохраняет immutable CalendarEvent command внутри transaction scope.
func (transaction *EventTransaction) PersistCalendarEvent(ctx context.Context, scope EventMutationScope, command CalendarEventCommand) (CalendarEventPersistenceResult, error) {
	if transaction == nil {
		return CalendarEventPersistenceResult{}, ErrEventTransactionClosed
	}
	transaction.mu.Lock()
	defer transaction.mu.Unlock()
	if err := transaction.validateScope(scope); err != nil {
		return CalendarEventPersistenceResult{}, err
	}
	if command.OrganizationID() != scope.organizationID {
		return CalendarEventPersistenceResult{}, ErrEventForbidden
	}
	if !command.valid() {
		return CalendarEventPersistenceResult{}, ErrEventInvalid
	}
	result, err := transaction.backend.PersistCalendarEvent(ctx, scope, command)
	if err != nil {
		return CalendarEventPersistenceResult{}, err
	}
	if !result.valid() {
		return CalendarEventPersistenceResult{}, ErrEventInvalid
	}
	return result, nil
}

// PersistCalendarEventAttendee сохраняет immutable attendee command внутри transaction scope.
func (transaction *EventTransaction) PersistCalendarEventAttendee(ctx context.Context, scope EventMutationScope, command CalendarEventAttendeeCommand) (CalendarEventAttendeePersistenceResult, error) {
	if transaction == nil {
		return CalendarEventAttendeePersistenceResult{}, ErrEventTransactionClosed
	}
	transaction.mu.Lock()
	defer transaction.mu.Unlock()
	if err := transaction.validateScope(scope); err != nil {
		return CalendarEventAttendeePersistenceResult{}, err
	}
	if !command.valid() {
		return CalendarEventAttendeePersistenceResult{}, ErrEventInvalid
	}
	result, err := transaction.backend.PersistCalendarEventAttendee(ctx, scope, command)
	if err != nil {
		return CalendarEventAttendeePersistenceResult{}, err
	}
	if !result.valid() {
		return CalendarEventAttendeePersistenceResult{}, ErrEventInvalid
	}
	return result, nil
}

// PersistCalendarEventException сохраняет immutable exception command внутри transaction scope.
func (transaction *EventTransaction) PersistCalendarEventException(ctx context.Context, scope EventMutationScope, command CalendarEventExceptionCommand) (CalendarEventExceptionPersistenceResult, error) {
	if transaction == nil {
		return CalendarEventExceptionPersistenceResult{}, ErrEventTransactionClosed
	}
	transaction.mu.Lock()
	defer transaction.mu.Unlock()
	if err := transaction.validateScope(scope); err != nil {
		return CalendarEventExceptionPersistenceResult{}, err
	}
	if !command.valid() {
		return CalendarEventExceptionPersistenceResult{}, ErrEventInvalid
	}
	result, err := transaction.backend.PersistCalendarEventException(ctx, scope, command)
	if err != nil {
		return CalendarEventExceptionPersistenceResult{}, err
	}
	if !result.valid() {
		return CalendarEventExceptionPersistenceResult{}, ErrEventInvalid
	}
	return result, nil
}

// PersistCalendarEventUpdate delegates one already-composed atomic update
// command to the optional update port and verifies its immutable result shape.
func (transaction *EventTransaction) PersistCalendarEventUpdate(ctx context.Context, scope EventMutationScope, command CalendarEventUpdatePersistenceCommand) (CalendarEventUpdatePersistenceResult, error) {
	if transaction == nil {
		return CalendarEventUpdatePersistenceResult{}, ErrEventTransactionClosed
	}
	transaction.mu.Lock()
	defer transaction.mu.Unlock()
	if !transaction.active || transaction.backend == nil {
		return CalendarEventUpdatePersistenceResult{}, ErrEventTransactionClosed
	}
	backend, supported := transaction.backend.(CalendarEventUpdateTransactionBackend)
	if !supported {
		return CalendarEventUpdatePersistenceResult{}, ErrEventInvalid
	}
	if err := transaction.validateScope(scope); err != nil {
		return CalendarEventUpdatePersistenceResult{}, err
	}
	if scope.eventOperation != EventMutationCalendarUpdate || !command.valid() || command.Event().OrganizationID() != scope.organizationID {
		return CalendarEventUpdatePersistenceResult{}, ErrEventInvalid
	}
	result, err := backend.PersistCalendarEventUpdate(ctx, scope, cloneCalendarEventUpdatePersistenceCommand(command))
	if err != nil {
		return CalendarEventUpdatePersistenceResult{}, err
	}
	if !result.valid() {
		return CalendarEventUpdatePersistenceResult{}, ErrEventInvalid
	}
	return cloneCalendarEventUpdatePersistenceResult(result), nil
}

// PersistCalendarEventOccurrenceCancel is the sole occurrence-cancel write
// boundary. It accepts and returns immutable snapshots so aliases cannot cross
// the transaction seam.
func (transaction *EventTransaction) PersistCalendarEventOccurrenceCancel(ctx context.Context, scope EventMutationScope, command CalendarEventOccurrenceCancelPersistenceCommand) (CalendarEventOccurrenceCancelPersistenceResult, error) {
	if transaction == nil {
		return CalendarEventOccurrenceCancelPersistenceResult{}, ErrEventTransactionClosed
	}
	transaction.mu.Lock()
	defer transaction.mu.Unlock()
	if !transaction.active || transaction.backend == nil {
		return CalendarEventOccurrenceCancelPersistenceResult{}, ErrEventTransactionClosed
	}
	backend, supported := transaction.backend.(CalendarEventOccurrenceCancelTransactionBackend)
	if !supported || nilCalendarEventOccurrenceCancelTransactionBackend(backend) {
		return CalendarEventOccurrenceCancelPersistenceResult{}, ErrEventInvalid
	}
	if err := transaction.validateScope(scope); err != nil {
		return CalendarEventOccurrenceCancelPersistenceResult{}, err
	}
	event := command.Event().Snapshot()
	if scope.eventOperation != EventMutationCalendarOccurrenceCancel || !command.valid() || event.OrganizationID != scope.organizationID || event.OwnerID != scope.actorID || event.OrganizerID == nil || *event.OrganizerID != scope.actorID {
		return CalendarEventOccurrenceCancelPersistenceResult{}, ErrEventInvalid
	}
	result, err := backend.PersistCalendarEventOccurrenceCancel(ctx, scope, cloneCalendarEventOccurrenceCancelPersistenceCommand(command))
	if err != nil {
		return CalendarEventOccurrenceCancelPersistenceResult{}, err
	}
	if !result.valid() {
		return CalendarEventOccurrenceCancelPersistenceResult{}, ErrEventInvalid
	}
	return cloneCalendarEventOccurrenceCancelPersistenceResult(result), nil
}

// PersistCalendarEventSeriesRSVP delegates one immutable actor-bound RSVP
// compare-and-swap mutation and accepts only its exact durable echo.
func (transaction *EventTransaction) PersistCalendarEventSeriesRSVP(ctx context.Context, scope EventMutationScope, command CalendarEventSeriesRSVPPersistenceCommand) (CalendarEventSeriesRSVPPersistenceResult, error) {
	if transaction == nil {
		return CalendarEventSeriesRSVPPersistenceResult{}, ErrEventTransactionClosed
	}
	transaction.mu.Lock()
	defer transaction.mu.Unlock()
	if !transaction.active || transaction.backend == nil {
		return CalendarEventSeriesRSVPPersistenceResult{}, ErrEventTransactionClosed
	}
	backend, supported := transaction.backend.(CalendarEventSeriesRSVPTransactionBackend)
	if !supported || nilCalendarEventSeriesRSVPTransactionBackend(backend) {
		return CalendarEventSeriesRSVPPersistenceResult{}, ErrEventInvalid
	}
	if err := transaction.validateScope(scope); err != nil {
		return CalendarEventSeriesRSVPPersistenceResult{}, err
	}
	attendee := command.Attendee()
	if scope.receiptOperation != MutationReceiptRSVPSeries || scope.eventOperation != EventMutationRSVPSeries || !command.valid() || attendee.UserID == nil || *attendee.UserID != scope.actorID {
		return CalendarEventSeriesRSVPPersistenceResult{}, ErrEventInvalid
	}
	result, err := backend.PersistCalendarEventSeriesRSVP(ctx, scope, cloneCalendarEventSeriesRSVPPersistenceCommand(command))
	if err != nil {
		return CalendarEventSeriesRSVPPersistenceResult{}, err
	}
	if !result.valid() || !calendarEventSeriesRSVPPersistenceMatches(command, result) {
		return CalendarEventSeriesRSVPPersistenceResult{}, ErrEventInvalid
	}
	return cloneCalendarEventSeriesRSVPPersistenceResult(result), nil
}

// AppendOutbox добавляет только полный accepted EnvelopeV1 внутри transaction scope.

// PersistCalendarEventOccurrenceRSVP delegates one immutable actor-bound
// occurrence RSVP compare-and-swap mutation and accepts only its exact echo.
// The series attendee is a snapshot for validation only and is never written.
func (transaction *EventTransaction) PersistCalendarEventOccurrenceRSVP(ctx context.Context, scope EventMutationScope, command CalendarEventOccurrenceRSVPPersistenceCommand) (CalendarEventOccurrenceRSVPPersistenceResult, error) {
	if transaction == nil {
		return CalendarEventOccurrenceRSVPPersistenceResult{}, ErrEventTransactionClosed
	}
	transaction.mu.Lock()
	defer transaction.mu.Unlock()
	if !transaction.active || transaction.backend == nil {
		return CalendarEventOccurrenceRSVPPersistenceResult{}, ErrEventTransactionClosed
	}
	backend, supported := transaction.backend.(CalendarEventOccurrenceRSVPTransactionBackend)
	if !supported || nilCalendarEventOccurrenceRSVPTransactionBackend(backend) {
		return CalendarEventOccurrenceRSVPPersistenceResult{}, ErrEventInvalid
	}
	if err := transaction.validateScope(scope); err != nil {
		return CalendarEventOccurrenceRSVPPersistenceResult{}, err
	}
	series := command.SeriesAttendee()
	attendee := command.Attendee()
	if scope.receiptOperation != MutationReceiptRSVPOccurrence || scope.eventOperation != EventMutationRSVPOccurrence || !command.valid() || series.UserID == nil || *series.UserID != scope.actorID || attendee.UserID == nil || *attendee.UserID != scope.actorID {
		return CalendarEventOccurrenceRSVPPersistenceResult{}, ErrEventInvalid
	}
	result, err := backend.PersistCalendarEventOccurrenceRSVP(ctx, scope, cloneCalendarEventOccurrenceRSVPPersistenceCommand(command))
	if err != nil {
		return CalendarEventOccurrenceRSVPPersistenceResult{}, err
	}
	if !result.valid() || !calendarEventOccurrenceRSVPPersistenceMatches(command, result) {
		return CalendarEventOccurrenceRSVPPersistenceResult{}, ErrEventInvalid
	}
	return cloneCalendarEventOccurrenceRSVPPersistenceResult(result), nil
}

// AppendOutbox appends only a valid immutable envelope inside its scope.
func (transaction *EventTransaction) AppendOutbox(ctx context.Context, scope EventMutationScope, descriptor EventOutboxDescriptor) error {
	if transaction == nil {
		return ErrEventTransactionClosed
	}
	transaction.mu.Lock()
	defer transaction.mu.Unlock()
	if err := transaction.validateScope(scope); err != nil {
		return err
	}
	if !descriptor.valid() {
		return ErrEventInvalid
	}
	return transaction.backend.AppendOutbox(ctx, scope, descriptor)
}

func cloneCalendarEventCreateAuthorizationRequest(request CalendarEventCreateAuthorizationRequest) CalendarEventCreateAuthorizationRequest {
	copyRequest := request
	copyRequest.Organizer = cloneCalendarEventCreateAttendeeCandidate(request.Organizer)
	copyRequest.Attendees = make([]CalendarEventCreateAttendeeCandidate, len(request.Attendees))
	for index, attendee := range request.Attendees {
		copyRequest.Attendees[index] = cloneCalendarEventCreateAttendeeCandidate(attendee)
	}
	copyRequest.Exdates = append([]string(nil), request.Exdates...)
	copyRequest.ReminderMinutes = cloneCalendarEventCreateInt(request.ReminderMinutes)
	copyRequest.ScheduledCallID = cloneCalendarEventCreateUUID(request.ScheduledCallID)
	return copyRequest
}

func cloneCalendarEventCreateAuthorization(authorization CalendarEventCreateAuthorization) CalendarEventCreateAuthorization {
	copyAuthorization := authorization
	copyAuthorization.Organizer.DisplayName = cloneCalendarEventCreateString(authorization.Organizer.DisplayName)
	copyAuthorization.Attendees = make([]CalendarEventCreateResolvedAttendee, len(authorization.Attendees))
	for index, attendee := range authorization.Attendees {
		copyAuthorization.Attendees[index] = CalendarEventCreateResolvedAttendee{
			UserID:      cloneCalendarEventCreateUUID(attendee.UserID),
			Email:       attendee.Email,
			DisplayName: cloneCalendarEventCreateString(attendee.DisplayName),
			Role:        attendee.Role,
		}
	}
	copyAuthorization.ScheduledCallID = cloneCalendarEventCreateUUID(authorization.ScheduledCallID)
	return copyAuthorization
}

func cloneCalendarEventCreateAttendeeCandidate(candidate CalendarEventCreateAttendeeCandidate) CalendarEventCreateAttendeeCandidate {
	return CalendarEventCreateAttendeeCandidate{
		UserID:      cloneCalendarEventCreateUUID(candidate.UserID),
		Email:       candidate.Email,
		DisplayName: cloneCalendarEventCreateString(candidate.DisplayName),
		Role:        candidate.Role,
	}
}

func cloneCalendarEventCreateUUID(value *uuid.UUID) *uuid.UUID {
	if value == nil {
		return nil
	}
	copyValue := *value
	return &copyValue
}

func cloneCalendarEventCreateString(value *string) *string {
	if value == nil {
		return nil
	}
	copyValue := *value
	return &copyValue
}

func cloneCalendarEventCreateInt(value *int) *int {
	if value == nil {
		return nil
	}
	copyValue := *value
	return &copyValue
}

func (transaction *EventTransaction) validateScope(scope EventMutationScope) error {
	if !transaction.active || transaction.backend == nil {
		return ErrEventTransactionClosed
	}
	if scope.transaction != transaction || scope.actorID == uuid.Nil || scope.organizationID == uuid.Nil || !validEventMutationScope(scope) {
		return ErrEventForbidden
	}
	return nil
}

// MutationReceiptOperation определяет разрешённую HTTP-операцию квитанции календарной мутации.
type MutationReceiptOperation string

const (
	// MutationReceiptEventCreate соответствует созданию канонического события.
	MutationReceiptEventCreate MutationReceiptOperation = "event_create"
	// MutationReceiptLegacyCallCreate соответствует созданию через legacy ScheduledCall bridge.
	MutationReceiptLegacyCallCreate MutationReceiptOperation = "legacy_call_create"
	// MutationReceiptRSVPSeries соответствует RSVP всей серии.
	MutationReceiptRSVPSeries MutationReceiptOperation = "rsvp_series"
	// MutationReceiptRSVPOccurrence соответствует RSVP одного экземпляра серии.
	MutationReceiptRSVPOccurrence MutationReceiptOperation = "rsvp_occurrence"
)

// EventMutationOperation identifies all canonical calendar mutations. Receipt
// operations map one-to-one to their historic values; update is receipt-free.
type EventMutationOperation string

const (
	EventMutationEventCreate      EventMutationOperation = "event_create"
	EventMutationLegacyCallCreate EventMutationOperation = "legacy_call_create"
	EventMutationRSVPSeries       EventMutationOperation = "rsvp_series"
	EventMutationRSVPOccurrence   EventMutationOperation = "rsvp_occurrence"
	EventMutationCalendarUpdate   EventMutationOperation = "event_update"
	EventMutationCalendarOccurrenceCancel EventMutationOperation = "event_occurrence_cancel"
)

// MutationReceiptHash хранит только SHA-256 производного ключа или канонического запроса.
type MutationReceiptHash [sha256.Size]byte

// MutationReceiptReservation является неизменяемой командой резервирования квитанции.
type MutationReceiptReservation struct {
	scope       EventMutationScope
	keyHash     MutationReceiptHash
	requestHash MutationReceiptHash
}

// NewMutationReceiptReservation создаёт команду без исходного Idempotency-Key и с обязательным transaction scope.
func NewMutationReceiptReservation(scope EventMutationScope, keyHash MutationReceiptHash, requestHash MutationReceiptHash) (MutationReceiptReservation, error) {
	if scope.transaction == nil || scope.actorID == uuid.Nil || scope.organizationID == uuid.Nil || !validMutationReceiptOperation(scope.receiptOperation) {
		return MutationReceiptReservation{}, ErrEventInvalid
	}
	return MutationReceiptReservation{scope: scope, keyHash: keyHash, requestHash: requestHash}, nil
}

func (reservation MutationReceiptReservation) matches(scope EventMutationScope) bool {
	return reservation.scope.transaction == scope.transaction && reservation.scope.actorID == scope.actorID && reservation.scope.organizationID == scope.organizationID && reservation.scope.receiptOperation == scope.receiptOperation
}

// KeyHash возвращает копию SHA-256 производного idempotency key.
func (reservation MutationReceiptReservation) KeyHash() MutationReceiptHash {
	return reservation.keyHash
}

// RequestHash возвращает копию SHA-256 канонического запроса.
func (reservation MutationReceiptReservation) RequestHash() MutationReceiptHash {
	return reservation.requestHash
}

// MutationReceiptResponse хранит неизменяемый HTTP-ответ, пригодный для replay.
type MutationReceiptResponse struct {
	status       int
	responseJSON string
}

// NewMutationReceiptResponse принимает только ограниченный корректный UTF-8 JSON-ответ.
func NewMutationReceiptResponse(status int, responseJSON []byte) (MutationReceiptResponse, error) {
	if status < 100 || status > 599 || len(responseJSON) == 0 || len(responseJSON) > maxMutationReceiptResponseBytes || !utf8.Valid(responseJSON) || !json.Valid(responseJSON) {
		return MutationReceiptResponse{}, ErrEventInvalid
	}
	return MutationReceiptResponse{status: status, responseJSON: string(responseJSON)}, nil
}

func (response MutationReceiptResponse) valid() bool {
	return response.status >= 100 && response.status <= 599 && len(response.responseJSON) > 0 && len(response.responseJSON) <= maxMutationReceiptResponseBytes && utf8.ValidString(response.responseJSON) && json.Valid([]byte(response.responseJSON))
}

// Status возвращает сохранённый HTTP-статус replay-ответа.
func (response MutationReceiptResponse) Status() int {
	return response.status
}

// JSON возвращает свежую копию сохранённого JSON-ответа.
func (response MutationReceiptResponse) JSON() []byte {
	return []byte(response.responseJSON)
}

// MutationReceiptReservationState описывает исход резервирования квитанции.
type MutationReceiptReservationState string

const (
	// MutationReceiptReserved означает новую квитанцию текущей транзакции.
	MutationReceiptReserved MutationReceiptReservationState = "reserved"
	// MutationReceiptReplayed означает завершённую квитанцию с сохранённым ответом.
	MutationReceiptReplayed MutationReceiptReservationState = "replayed"
)

// MutationReceiptReservationResult является неизменяемым результатом резервирования.
type MutationReceiptReservationResult struct {
	state     MutationReceiptReservationState
	receiptID uuid.UUID
	replay    *MutationReceiptResponse
}

// NewReservedMutationReceiptResult создаёт результат новой резервируемой квитанции.
func NewReservedMutationReceiptResult(receiptID uuid.UUID) (MutationReceiptReservationResult, error) {
	if receiptID == uuid.Nil {
		return MutationReceiptReservationResult{}, ErrEventInvalid
	}
	return MutationReceiptReservationResult{state: MutationReceiptReserved, receiptID: receiptID}, nil
}

// NewReplayedMutationReceiptResult создаёт результат completed receipt для byte-equivalent replay.
func NewReplayedMutationReceiptResult(receiptID uuid.UUID, response MutationReceiptResponse) (MutationReceiptReservationResult, error) {
	if receiptID == uuid.Nil || !response.valid() {
		return MutationReceiptReservationResult{}, ErrEventInvalid
	}
	replay := response
	return MutationReceiptReservationResult{state: MutationReceiptReplayed, receiptID: receiptID, replay: &replay}, nil
}

// State возвращает исход резервирования квитанции.
func (result MutationReceiptReservationResult) State() MutationReceiptReservationState {
	return result.state
}

// ReceiptID возвращает идентификатор квитанции.
func (result MutationReceiptReservationResult) ReceiptID() uuid.UUID {
	return result.receiptID
}

// ReplayedResponse возвращает сохранённый ответ только для replay.
func (result MutationReceiptReservationResult) ReplayedResponse() (MutationReceiptResponse, bool) {
	if result.replay == nil {
		return MutationReceiptResponse{}, false
	}
	return *result.replay, true
}

func (result MutationReceiptReservationResult) valid() bool {
	if result.receiptID == uuid.Nil {
		return false
	}
	switch result.state {
	case MutationReceiptReserved:
		return result.replay == nil
	case MutationReceiptReplayed:
		return result.replay != nil && result.replay.valid()
	default:
		return false
	}
}

// MutationReceiptCompletion является неизменяемой командой завершения квитанции.
type MutationReceiptCompletion struct {
	receiptID uuid.UUID
	eventID   *uuid.UUID
	response  MutationReceiptResponse
}

// NewMutationReceiptCompletion создаёт команду завершения с необязательным событием.
func NewMutationReceiptCompletion(receiptID uuid.UUID, eventID *uuid.UUID, response MutationReceiptResponse) (MutationReceiptCompletion, error) {
	if receiptID == uuid.Nil || !response.valid() {
		return MutationReceiptCompletion{}, ErrEventInvalid
	}
	var eventIDCopy *uuid.UUID
	if eventID != nil {
		if *eventID == uuid.Nil {
			return MutationReceiptCompletion{}, ErrEventInvalid
		}
		copied := *eventID
		eventIDCopy = &copied
	}
	return MutationReceiptCompletion{receiptID: receiptID, eventID: eventIDCopy, response: response}, nil
}

// ReceiptID возвращает идентификатор завершаемой квитанции.
func (completion MutationReceiptCompletion) ReceiptID() uuid.UUID {
	return completion.receiptID
}

// EventID возвращает идентификатор события только если он связан с квитанцией.
func (completion MutationReceiptCompletion) EventID() (uuid.UUID, bool) {
	if completion.eventID == nil {
		return uuid.Nil, false
	}
	return *completion.eventID, true
}

// Response возвращает неизменяемый replay-ответ квитанции.
func (completion MutationReceiptCompletion) Response() MutationReceiptResponse {
	return completion.response
}

func (completion MutationReceiptCompletion) valid() bool {
	return completion.receiptID != uuid.Nil && completion.response.valid() && (completion.eventID == nil || *completion.eventID != uuid.Nil)
}

// CalendarEventSource определяет происхождение канонического события.
type CalendarEventSource string

const (
	// CalendarEventSourceHub обозначает событие, созданное Hub.
	CalendarEventSourceHub CalendarEventSource = "hub"
	// CalendarEventSourceMailImport обозначает событие, импортированное из почты.
	CalendarEventSourceMailImport CalendarEventSource = "mail_import"
	// CalendarEventSourceICalFeed обозначает событие из iCalendar feed.
	CalendarEventSourceICalFeed CalendarEventSource = "ical_feed"
)

// CalendarEventStatus определяет состояние канонического события.
type CalendarEventStatus string

const (
	// CalendarEventStatusConfirmed обозначает подтверждённое событие.
	CalendarEventStatusConfirmed CalendarEventStatus = "confirmed"
	// CalendarEventStatusTentative обозначает предварительное событие.
	CalendarEventStatusTentative CalendarEventStatus = "tentative"
	// CalendarEventStatusCancelled обозначает отменённое событие.
	CalendarEventStatusCancelled CalendarEventStatus = "cancelled"
)

// CalendarEventCommandInput содержит изменяемый вход, который конструктор немедленно snapshot-ит.
type CalendarEventCommandInput struct {
	EventID         uuid.UUID
	OrganizationID  uuid.UUID
	OwnerID         uuid.UUID
	UID             string
	Title           string
	Description     string
	Location        string
	Temporal        CalendarEvent
	RRULE           string
	Exdates         []string
	Sequence        uint
	Version         uint
	Source          CalendarEventSource
	Status          CalendarEventStatus
	OrganizerID     *uuid.UUID
	OrganizerEmail  string
	ScheduledCallID *uuid.UUID
	ReminderMinutes *int
}

// CalendarEventCommand является неизменяемой persistence-командой CalendarEvent.
type CalendarEventCommand struct {
	input CalendarEventCommandInput
}

// NewCalendarEventCommand проверяет и snapshot-ит полный доменный снимок CalendarEvent.
func NewCalendarEventCommand(input CalendarEventCommandInput) (CalendarEventCommand, error) {
	if input.EventID == uuid.Nil || input.OrganizationID == uuid.Nil || input.OwnerID == uuid.Nil || strings.TrimSpace(input.UID) == "" || strings.TrimSpace(input.OrganizerEmail) == "" || input.Version == 0 || !validCalendarEventSource(input.Source) || !validCalendarEventStatus(input.Status) || input.Temporal.Validate() != nil {
		return CalendarEventCommand{}, ErrEventInvalid
	}
	if input.OrganizerID != nil && *input.OrganizerID == uuid.Nil {
		return CalendarEventCommand{}, ErrEventInvalid
	}
	if input.ScheduledCallID != nil && *input.ScheduledCallID == uuid.Nil {
		return CalendarEventCommand{}, ErrEventInvalid
	}
	if input.ReminderMinutes != nil && (*input.ReminderMinutes < 0 || *input.ReminderMinutes > 40320) {
		return CalendarEventCommand{}, ErrEventInvalid
	}
	input.Exdates = append([]string(nil), input.Exdates...)
	input.OrganizerID = copyUUID(input.OrganizerID)
	input.ScheduledCallID = copyUUID(input.ScheduledCallID)
	input.ReminderMinutes = copyInt(input.ReminderMinutes)
	return CalendarEventCommand{input: input}, nil
}

// EventID возвращает идентификатор события.
func (command CalendarEventCommand) EventID() uuid.UUID { return command.input.EventID }

// OrganizationID возвращает организацию persistence-команды.
func (command CalendarEventCommand) OrganizationID() uuid.UUID { return command.input.OrganizationID }

// Version возвращает optimistic version события.
func (command CalendarEventCommand) Version() uint { return command.input.Version }

// Exdates возвращает свежую копию исключённых recurrence значений.
func (command CalendarEventCommand) Exdates() []string { return append([]string(nil), command.input.Exdates...) }

// Snapshot returns an isolated full event input for a follow-on atomic command.
func (command CalendarEventCommand) Snapshot() CalendarEventCommandInput {
	return cloneCalendarEventCommandInput(command.input)
}

func (command CalendarEventCommand) valid() bool {
	_, err := NewCalendarEventCommand(command.input)
	return err == nil
}

// CalendarEventAttendeeRole определяет роль участника.
type CalendarEventAttendeeRole string

const (
	// CalendarEventAttendeeRoleRequired обозначает обязательного участника.
	CalendarEventAttendeeRoleRequired CalendarEventAttendeeRole = "required"
	// CalendarEventAttendeeRoleOptional обозначает необязательного участника.
	CalendarEventAttendeeRoleOptional CalendarEventAttendeeRole = "optional"
)

// CalendarEventAttendeeCommandInput содержит изменяемый вход attendee persistence-команды.
type CalendarEventAttendeeCommandInput struct {
	AttendeeID   uuid.UUID
	EventID      uuid.UUID
	UserID       *uuid.UUID
	Email        string
	DisplayName  string
	Role         CalendarEventAttendeeRole
	Status       RSVPStatus
	RecurrenceID string
	Version      uint
}

// CalendarEventAttendeeCommand является неизменяемой persistence-командой attendee.
type CalendarEventAttendeeCommand struct {
	input CalendarEventAttendeeCommandInput
}

// NewCalendarEventAttendeeCommand проверяет и snapshot-ит attendee-команду.
func NewCalendarEventAttendeeCommand(input CalendarEventAttendeeCommandInput) (CalendarEventAttendeeCommand, error) {
	if input.AttendeeID == uuid.Nil || input.EventID == uuid.Nil || strings.TrimSpace(input.Email) == "" || input.Version == 0 || !validCalendarEventAttendeeRole(input.Role) || input.Status.Validate() != nil {
		return CalendarEventAttendeeCommand{}, ErrEventInvalid
	}
	if input.UserID != nil && *input.UserID == uuid.Nil {
		return CalendarEventAttendeeCommand{}, ErrEventInvalid
	}
	input.UserID = copyUUID(input.UserID)
	return CalendarEventAttendeeCommand{input: input}, nil
}

// AttendeeID возвращает идентификатор attendee.
func (command CalendarEventAttendeeCommand) AttendeeID() uuid.UUID { return command.input.AttendeeID }

// EventID возвращает идентификатор родительского события attendee.
func (command CalendarEventAttendeeCommand) EventID() uuid.UUID { return command.input.EventID }

// Version возвращает optimistic version attendee.
func (command CalendarEventAttendeeCommand) Version() uint { return command.input.Version }

func (command CalendarEventAttendeeCommand) valid() bool {
	_, err := NewCalendarEventAttendeeCommand(command.input)
	return err == nil
}

// CalendarEventExceptionCommandInput содержит изменяемый вход exception persistence-команды.
type CalendarEventExceptionCommandInput struct {
	ExceptionID  uuid.UUID
	EventID      uuid.UUID
	RecurrenceID string
	Temporal     *CalendarEvent
	Sequence     uint
	Status       CalendarEventStatus
	Cancelled    bool
}

// CalendarEventExceptionCommand является неизменяемой persistence-командой exception.
type CalendarEventExceptionCommand struct {
	input CalendarEventExceptionCommandInput
}

// NewCalendarEventExceptionCommand проверяет и snapshot-ит exception-команду.
func NewCalendarEventExceptionCommand(input CalendarEventExceptionCommandInput) (CalendarEventExceptionCommand, error) {
	if input.ExceptionID == uuid.Nil || input.EventID == uuid.Nil || strings.TrimSpace(input.RecurrenceID) == "" || !validCalendarEventStatus(input.Status) || (input.Cancelled && input.Status != CalendarEventStatusCancelled) {
		return CalendarEventExceptionCommand{}, ErrEventInvalid
	}
	if input.Temporal != nil {
		if input.Temporal.Validate() != nil {
			return CalendarEventExceptionCommand{}, ErrEventInvalid
		}
		copied := *input.Temporal
		input.Temporal = &copied
	}
	return CalendarEventExceptionCommand{input: input}, nil
}

// ExceptionID возвращает идентификатор exception.
func (command CalendarEventExceptionCommand) ExceptionID() uuid.UUID { return command.input.ExceptionID }

// EventID возвращает идентификатор родительского события exception.
func (command CalendarEventExceptionCommand) EventID() uuid.UUID { return command.input.EventID }

// Sequence возвращает sequence exception.
func (command CalendarEventExceptionCommand) Sequence() uint { return command.input.Sequence }

// Snapshot returns an isolated exception input, including any temporal value.
func (command CalendarEventExceptionCommand) Snapshot() CalendarEventExceptionCommandInput {
	return cloneCalendarEventExceptionCommandInput(command.input)
}

func (command CalendarEventExceptionCommand) valid() bool {
	_, err := NewCalendarEventExceptionCommand(command.input)
	return err == nil
}

// CalendarEventPersistenceResult является неизменяемым результатом сохранения event.
type CalendarEventPersistenceResult struct { eventID uuid.UUID; version uint }

// NewCalendarEventPersistenceResult создаёт result event persistence.
func NewCalendarEventPersistenceResult(eventID uuid.UUID, version uint) (CalendarEventPersistenceResult, error) {
	if eventID == uuid.Nil || version == 0 { return CalendarEventPersistenceResult{}, ErrEventInvalid }
	return CalendarEventPersistenceResult{eventID: eventID, version: version}, nil
}

// EventID returns the immutable event identifier persisted by the adapter.
func (result CalendarEventPersistenceResult) EventID() uuid.UUID { return result.eventID }

// Version returns the optimistic version persisted by the adapter.
func (result CalendarEventPersistenceResult) Version() uint { return result.version }

func (result CalendarEventPersistenceResult) valid() bool { return result.eventID != uuid.Nil && result.version > 0 }

// CalendarEventAttendeePersistenceResult является неизменяемым результатом attendee persistence.
type CalendarEventAttendeePersistenceResult struct { attendeeID uuid.UUID; version uint }

// NewCalendarEventAttendeePersistenceResult создаёт result attendee persistence.
func NewCalendarEventAttendeePersistenceResult(attendeeID uuid.UUID, version uint) (CalendarEventAttendeePersistenceResult, error) {
	if attendeeID == uuid.Nil || version == 0 { return CalendarEventAttendeePersistenceResult{}, ErrEventInvalid }
	return CalendarEventAttendeePersistenceResult{attendeeID: attendeeID, version: version}, nil
}

// AttendeeID returns the immutable attendee identifier persisted by the adapter.
func (result CalendarEventAttendeePersistenceResult) AttendeeID() uuid.UUID { return result.attendeeID }

// Version returns the optimistic version persisted by the adapter.
func (result CalendarEventAttendeePersistenceResult) Version() uint { return result.version }

func (result CalendarEventAttendeePersistenceResult) valid() bool { return result.attendeeID != uuid.Nil && result.version > 0 }

// CalendarEventExceptionPersistenceResult является неизменяемым результатом exception persistence.
type CalendarEventExceptionPersistenceResult struct { exceptionID uuid.UUID; sequence uint }

// NewCalendarEventExceptionPersistenceResult создаёт result exception persistence.
func NewCalendarEventExceptionPersistenceResult(exceptionID uuid.UUID, sequence uint) (CalendarEventExceptionPersistenceResult, error) {
	if exceptionID == uuid.Nil { return CalendarEventExceptionPersistenceResult{}, ErrEventInvalid }
	return CalendarEventExceptionPersistenceResult{exceptionID: exceptionID, sequence: sequence}, nil
}

// ExceptionID returns the immutable exception identifier persisted by the adapter.
func (result CalendarEventExceptionPersistenceResult) ExceptionID() uuid.UUID { return result.exceptionID }

// Sequence returns the exception sequence persisted by the adapter.
func (result CalendarEventExceptionPersistenceResult) Sequence() uint { return result.sequence }

func (result CalendarEventExceptionPersistenceResult) valid() bool { return result.exceptionID != uuid.Nil }

// EventOutboxDescriptor переносит полный принятый EnvelopeV1 в транзакционный outbox.
type EventOutboxDescriptor struct { envelope events.EnvelopeV1 }

// NewEventOutboxDescriptor запрещает inner-payload-only описание outbox-события.
func NewEventOutboxDescriptor(envelope events.EnvelopeV1) (EventOutboxDescriptor, error) {
	if envelope.Subject() == "" || envelope.EventID() == uuid.Nil { return EventOutboxDescriptor{}, ErrEventInvalid }
	return EventOutboxDescriptor{envelope: envelope}, nil
}

// Envelope возвращает полный неизменяемый EnvelopeV1 для PublishInTx seam.
func (descriptor EventOutboxDescriptor) Envelope() events.EnvelopeV1 { return descriptor.envelope }

func (descriptor EventOutboxDescriptor) valid() bool {
	return descriptor.envelope.Subject() != "" && descriptor.envelope.EventID() != uuid.Nil
}

func validMutationReceiptOperation(operation MutationReceiptOperation) bool {
	switch operation {
	case MutationReceiptEventCreate, MutationReceiptLegacyCallCreate, MutationReceiptRSVPSeries, MutationReceiptRSVPOccurrence:
		return true
	default:
		return false
	}
}

func eventMutationOperationForReceipt(operation MutationReceiptOperation) (EventMutationOperation, bool) {
	switch operation {
	case MutationReceiptEventCreate:
		return EventMutationEventCreate, true
	case MutationReceiptLegacyCallCreate:
		return EventMutationLegacyCallCreate, true
	case MutationReceiptRSVPSeries:
		return EventMutationRSVPSeries, true
	case MutationReceiptRSVPOccurrence:
		return EventMutationRSVPOccurrence, true
	default:
		return "", false
	}
}

func validEventMutationScope(scope EventMutationScope) bool {
	if operation, found := eventMutationOperationForReceipt(scope.receiptOperation); found {
		return scope.eventOperation == operation
	}
	if scope.receiptOperation != "" {
		return false
	}
	return scope.eventOperation == EventMutationCalendarUpdate || scope.eventOperation == EventMutationCalendarOccurrenceCancel
}

func validCalendarEventSource(source CalendarEventSource) bool {
	return source == CalendarEventSourceHub || source == CalendarEventSourceMailImport || source == CalendarEventSourceICalFeed
}

func validCalendarEventStatus(status CalendarEventStatus) bool {
	return status == CalendarEventStatusConfirmed || status == CalendarEventStatusTentative || status == CalendarEventStatusCancelled
}

func validCalendarEventAttendeeRole(role CalendarEventAttendeeRole) bool {
	return role == CalendarEventAttendeeRoleRequired || role == CalendarEventAttendeeRoleOptional
}

func copyUUID(value *uuid.UUID) *uuid.UUID {
	if value == nil { return nil }
	copied := *value
	return &copied
}

func copyInt(value *int) *int {
	if value == nil { return nil }
	copied := *value
	return &copied
}

func cloneCalendarEventOccurrenceCancelAuthorizationRequest(request CalendarEventOccurrenceCancelAuthorizationRequest) CalendarEventOccurrenceCancelAuthorizationRequest {
	return CalendarEventOccurrenceCancelAuthorizationRequest{EventID: request.EventID, RecurrenceID: request.RecurrenceID}
}

func cloneCalendarEventOccurrenceCancelAuthorization(authorization CalendarEventOccurrenceCancelAuthorization) CalendarEventOccurrenceCancelAuthorization {
	cloned := authorization
	cloned.Current = cloneCalendarEventCommandInput(authorization.Current)
	if authorization.Attendees == nil {
		cloned.Attendees = nil
	} else {
		cloned.Attendees = make([]CalendarEventOccurrenceCancelAttendeeSnapshot, len(authorization.Attendees))
		for index, attendee := range authorization.Attendees {
			cloned.Attendees[index] = cloneCalendarEventOccurrenceCancelAttendeeSnapshot(attendee)
		}
	}
	if authorization.Exception == nil {
		cloned.Exception = nil
	} else {
		value := cloneCalendarEventExceptionCommandInput(*authorization.Exception)
		cloned.Exception = &value
	}
	return cloned
}

func cloneCalendarEventOccurrenceCancelAttendeeSnapshot(attendee CalendarEventOccurrenceCancelAttendeeSnapshot) CalendarEventOccurrenceCancelAttendeeSnapshot {
	attendee.UserID = copyUUID(attendee.UserID)
	if attendee.RemovedAt != nil {
		value := *attendee.RemovedAt
		attendee.RemovedAt = &value
	}
	return attendee
}

func validCalendarEventSeriesRSVPAttendee(attendee CalendarEventSeriesRSVPAttendeeSnapshot) bool {
	if attendee.AttendeeID == uuid.Nil || attendee.EventID == uuid.Nil || attendee.UserID == nil || *attendee.UserID == uuid.Nil || attendee.Version == 0 || attendee.RemovedAt != nil || attendee.Email != normalizeCalendarCreateEmail(attendee.Email) || !validCalendarCreateEmail(attendee.Email) || !validCalendarCreateDisplayName(&attendee.DisplayName) || !validCalendarEventAttendeeRole(attendee.Role) || attendee.Status.Validate() != nil {
		return false
	}
	if attendee.Status == RSVPNeedsAction {
		return attendee.RespondedAt == nil
	}
	return attendee.RespondedAt != nil && validCalendarEventSeriesRSVPUTC(*attendee.RespondedAt)
}

func validCalendarEventSeriesRSVPUTC(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC
}

func validCalendarEventSeriesRSVPReply(reply *CalendarEventSeriesRSVPReplyDeliveryDescriptor, attendee CalendarEventSeriesRSVPAttendeeSnapshot, eventSequence uint) bool {
	if reply == nil {
		return true
	}
	return reply.AttendeeID == attendee.AttendeeID && reply.EventID == attendee.EventID && reply.EventSequence == eventSequence && reply.AttendeeVersion == attendee.Version && reply.Status == attendee.Status && reply.Status.Validate() == nil
}

func cloneCalendarEventSeriesRSVPAuthorization(authorization CalendarEventSeriesRSVPAuthorization) CalendarEventSeriesRSVPAuthorization {
	authorization.Current = cloneCalendarEventCommandInput(authorization.Current)
	authorization.Attendee = cloneCalendarEventSeriesRSVPAttendeeSnapshot(authorization.Attendee)
	return authorization
}

func cloneCalendarEventSeriesRSVPAttendeeSnapshot(attendee CalendarEventSeriesRSVPAttendeeSnapshot) CalendarEventSeriesRSVPAttendeeSnapshot {
	attendee.UserID = copyUUID(attendee.UserID)
	attendee.RespondedAt = cloneCalendarEventSeriesRSVPTime(attendee.RespondedAt)
	attendee.RemovedAt = cloneCalendarEventSeriesRSVPTime(attendee.RemovedAt)
	return attendee
}

func cloneCalendarEventSeriesRSVPTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneCalendarEventSeriesRSVPReplyDeliveryDescriptor(reply *CalendarEventSeriesRSVPReplyDeliveryDescriptor) *CalendarEventSeriesRSVPReplyDeliveryDescriptor {
	if reply == nil {
		return nil
	}
	cloned := *reply
	return &cloned
}

func cloneCalendarEventSeriesRSVPPersistenceCommand(command CalendarEventSeriesRSVPPersistenceCommand) CalendarEventSeriesRSVPPersistenceCommand {
	cloned, err := NewCalendarEventSeriesRSVPPersistenceCommand(command.attendee, command.expectedVersion, command.eventSequence, command.outbox, command.reply)
	if err != nil {
		return CalendarEventSeriesRSVPPersistenceCommand{}
	}
	return cloned
}

func cloneCalendarEventSeriesRSVPPersistenceResult(result CalendarEventSeriesRSVPPersistenceResult) CalendarEventSeriesRSVPPersistenceResult {
	cloned, err := NewCalendarEventSeriesRSVPPersistenceResult(result.attendee, result.eventSequence, result.outbox, result.reply)
	if err != nil {
		return CalendarEventSeriesRSVPPersistenceResult{}
	}
	return cloned
}

func calendarEventSeriesRSVPPersistenceMatches(command CalendarEventSeriesRSVPPersistenceCommand, result CalendarEventSeriesRSVPPersistenceResult) bool {
	if !reflect.DeepEqual(command.Attendee(), result.Attendee()) || command.EventSequence() != result.EventSequence() || !reflect.DeepEqual(command.Outbox(), result.Outbox()) {
		return false
	}
	commandReply, commandHasReply := command.Reply()
	resultReply, resultHasReply := result.Reply()
	return commandHasReply == resultHasReply && (!commandHasReply || commandReply == resultReply)
}

func cloneCalendarEventOccurrenceRSVPAuthorizationRequest(request CalendarEventOccurrenceRSVPAuthorizationRequest) CalendarEventOccurrenceRSVPAuthorizationRequest {
	return CalendarEventOccurrenceRSVPAuthorizationRequest{EventID: request.EventID, RecurrenceID: request.RecurrenceID}
}

func cloneCalendarEventOccurrenceRSVPAuthorization(authorization CalendarEventOccurrenceRSVPAuthorization) CalendarEventOccurrenceRSVPAuthorization {
	cloned := authorization
	cloned.Current = cloneCalendarEventCommandInput(authorization.Current)
	cloned.SeriesAttendee = cloneCalendarEventSeriesRSVPAttendeeSnapshot(authorization.SeriesAttendee)
	cloned.OccurrenceAttendee = cloneCalendarEventOccurrenceRSVPAttendeeSnapshotPointer(authorization.OccurrenceAttendee)
	if authorization.Exception == nil {
		cloned.Exception = nil
	} else {
		value := cloneCalendarEventExceptionCommandInput(*authorization.Exception)
		cloned.Exception = &value
	}
	return cloned
}

func cloneCalendarEventOccurrenceRSVPAttendeeSnapshot(attendee CalendarEventOccurrenceRSVPAttendeeSnapshot) CalendarEventOccurrenceRSVPAttendeeSnapshot {
	attendee.UserID = copyUUID(attendee.UserID)
	attendee.RespondedAt = cloneCalendarEventSeriesRSVPTime(attendee.RespondedAt)
	attendee.RemovedAt = cloneCalendarEventSeriesRSVPTime(attendee.RemovedAt)
	return attendee
}

func cloneCalendarEventOccurrenceRSVPAttendeeSnapshotPointer(attendee *CalendarEventOccurrenceRSVPAttendeeSnapshot) *CalendarEventOccurrenceRSVPAttendeeSnapshot {
	if attendee == nil {
		return nil
	}
	cloned := cloneCalendarEventOccurrenceRSVPAttendeeSnapshot(*attendee)
	return &cloned
}

func cloneCalendarEventOccurrenceRSVPReplyDeliveryDescriptor(reply *CalendarEventOccurrenceRSVPReplyDeliveryDescriptor) *CalendarEventOccurrenceRSVPReplyDeliveryDescriptor {
	if reply == nil {
		return nil
	}
	cloned := *reply
	return &cloned
}

func validCalendarEventOccurrenceRSVPPersistence(series CalendarEventSeriesRSVPAttendeeSnapshot, prior *CalendarEventOccurrenceRSVPAttendeeSnapshot, attendee CalendarEventOccurrenceRSVPAttendeeSnapshot, expectedVersion uint, eventSequence uint, outbox EventOutboxDescriptor, reply *CalendarEventOccurrenceRSVPReplyDeliveryDescriptor) bool {
	if expectedVersion == 0 || expectedVersion == ^uint(0) || !validCalendarEventSeriesRSVPAttendee(series) || !validCalendarEventOccurrenceRSVPAttendeeForSeries(series, attendee) || !outbox.valid() || !validCalendarEventOccurrenceRSVPReply(reply, attendee, eventSequence) {
		return false
	}
	if prior == nil {
		return expectedVersion == series.Version && attendee.AttendeeID != series.AttendeeID && attendee.Version == expectedVersion+1
	}
	if !validCalendarEventOccurrenceRSVPAttendeeForSeries(series, *prior) || !calendarEventOccurrenceRSVPImmutableIdentityMatches(*prior, attendee) {
		return false
	}
	return expectedVersion == prior.Version && attendee.Version == expectedVersion+1
}

func validCalendarEventOccurrenceRSVPAttendeeForSeries(series CalendarEventSeriesRSVPAttendeeSnapshot, attendee CalendarEventOccurrenceRSVPAttendeeSnapshot) bool {
	return validCalendarEventOccurrenceRSVPAttendee(attendee) && attendee.AttendeeID != series.AttendeeID && attendee.EventID == series.EventID && attendee.UserID != nil && series.UserID != nil && *attendee.UserID == *series.UserID && attendee.Email == series.Email && attendee.DisplayName == series.DisplayName && attendee.Role == series.Role
}

func validCalendarEventOccurrenceRSVPAttendee(attendee CalendarEventOccurrenceRSVPAttendeeSnapshot) bool {
	if attendee.AttendeeID == uuid.Nil || attendee.EventID == uuid.Nil || attendee.UserID == nil || *attendee.UserID == uuid.Nil || attendee.Version == 0 || attendee.RemovedAt != nil || !validCalendarEventOccurrenceCancelCanonicalKey(attendee.RecurrenceID) || attendee.Email != normalizeCalendarCreateEmail(attendee.Email) || !validCalendarCreateEmail(attendee.Email) || !validCalendarCreateDisplayName(&attendee.DisplayName) || !validCalendarEventAttendeeRole(attendee.Role) || attendee.Status.Validate() != nil {
		return false
	}
	if attendee.Status == RSVPNeedsAction {
		return attendee.RespondedAt == nil
	}
	return attendee.RespondedAt != nil && validCalendarEventSeriesRSVPUTC(*attendee.RespondedAt)
}

func calendarEventOccurrenceRSVPImmutableIdentityMatches(prior, attendee CalendarEventOccurrenceRSVPAttendeeSnapshot) bool {
	return prior.AttendeeID == attendee.AttendeeID && prior.EventID == attendee.EventID && prior.UserID != nil && attendee.UserID != nil && *prior.UserID == *attendee.UserID && prior.Email == attendee.Email && prior.DisplayName == attendee.DisplayName && prior.Role == attendee.Role && prior.RecurrenceID == attendee.RecurrenceID
}

func validCalendarEventOccurrenceRSVPReply(reply *CalendarEventOccurrenceRSVPReplyDeliveryDescriptor, attendee CalendarEventOccurrenceRSVPAttendeeSnapshot, eventSequence uint) bool {
	if reply == nil {
		return true
	}
	return reply.AttendeeID == attendee.AttendeeID && reply.EventID == attendee.EventID && reply.EventSequence == eventSequence && reply.AttendeeVersion == attendee.Version && reply.RecurrenceID == attendee.RecurrenceID && reply.Status == attendee.Status && reply.Status.Validate() == nil
}

func validCalendarEventOccurrenceRSVPException(exception CalendarEventExceptionCommandInput, request CalendarEventOccurrenceRSVPAuthorizationRequest) bool {
	if exception.EventID != request.EventID || exception.RecurrenceID != request.RecurrenceID {
		return false
	}
	_, err := NewCalendarEventExceptionCommand(exception)
	return err == nil
}

func calendarEventOccurrenceRSVPAuthorizationIdentityMatches(authorization CalendarEventOccurrenceRSVPAuthorization, scope EventMutationScope, request CalendarEventOccurrenceRSVPAuthorizationRequest) bool {
	if authorization.Current.EventID != request.EventID || authorization.Current.OrganizationID != scope.organizationID || authorization.SeriesAttendee.EventID != request.EventID || authorization.SeriesAttendee.UserID == nil || *authorization.SeriesAttendee.UserID != scope.actorID {
		return false
	}
	if authorization.OccurrenceAttendee != nil && (authorization.OccurrenceAttendee.EventID != request.EventID || authorization.OccurrenceAttendee.UserID == nil || *authorization.OccurrenceAttendee.UserID != scope.actorID || authorization.OccurrenceAttendee.RecurrenceID != request.RecurrenceID) {
		return false
	}
	return authorization.Exception == nil || (authorization.Exception.EventID == request.EventID && authorization.Exception.RecurrenceID == request.RecurrenceID)
}

func calendarEventOccurrenceRSVPAuthorizationIdentityMask() CalendarEventOccurrenceRSVPAuthorization {
	return CalendarEventOccurrenceRSVPAuthorization{}
}

func cloneCalendarEventOccurrenceRSVPPersistenceCommand(command CalendarEventOccurrenceRSVPPersistenceCommand) CalendarEventOccurrenceRSVPPersistenceCommand {
	cloned, err := NewCalendarEventOccurrenceRSVPPersistenceCommand(command.seriesAttendee, command.prior, command.attendee, command.expectedVersion, command.eventSequence, command.outbox, command.reply)
	if err != nil {
		return CalendarEventOccurrenceRSVPPersistenceCommand{}
	}
	return cloned
}

func cloneCalendarEventOccurrenceRSVPPersistenceResult(result CalendarEventOccurrenceRSVPPersistenceResult) CalendarEventOccurrenceRSVPPersistenceResult {
	cloned, err := NewCalendarEventOccurrenceRSVPPersistenceResult(result.attendee, result.eventSequence, result.outbox, result.reply)
	if err != nil {
		return CalendarEventOccurrenceRSVPPersistenceResult{}
	}
	return cloned
}

func calendarEventOccurrenceRSVPPersistenceMatches(command CalendarEventOccurrenceRSVPPersistenceCommand, result CalendarEventOccurrenceRSVPPersistenceResult) bool {
	if !reflect.DeepEqual(command.Attendee(), result.Attendee()) || command.EventSequence() != result.EventSequence() || !reflect.DeepEqual(command.Outbox(), result.Outbox()) {
		return false
	}
	commandReply, commandHasReply := command.Reply()
	resultReply, resultHasReply := result.Reply()
	return commandHasReply == resultHasReply && (!commandHasReply || commandReply == resultReply)
}

func cloneCalendarEventExceptionCommandInput(input CalendarEventExceptionCommandInput) CalendarEventExceptionCommandInput {
	cloned := input
	if input.Temporal != nil {
		value := *input.Temporal
		cloned.Temporal = &value
	}
	return cloned
}

func nilCalendarEventOccurrenceCancelTransactionBackend(backend CalendarEventOccurrenceCancelTransactionBackend) bool {
	if backend == nil { return true }
	value := reflect.ValueOf(backend)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func nilCalendarEventOccurrenceRSVPTransactionBackend(backend CalendarEventOccurrenceRSVPTransactionBackend) bool {
	if backend == nil { return true }
	value := reflect.ValueOf(backend)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func nilCalendarEventSeriesRSVPTransactionBackend(backend CalendarEventSeriesRSVPTransactionBackend) bool {
	if backend == nil { return true }
	value := reflect.ValueOf(backend)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
