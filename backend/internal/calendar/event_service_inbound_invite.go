package calendar

import (
	"context"
	"reflect"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/internal/ical"
)

// CalendarInboundInviteRequest is an opaque normalized inbound iTIP request.
type CalendarInboundInviteRequest struct {
	message        ITIPMessage
	identity       ITIPIdentity
	authentication CalendarInboundAuthenticationClass
}

func NewCalendarInboundInviteRequest(
	message ITIPMessage,
	authentication CalendarInboundAuthenticationClass,
) (CalendarInboundInviteRequest, error) {
	normalized, err := ical.NormalizeITIPMessage(message)
	if err != nil {
		return CalendarInboundInviteRequest{}, ErrEventInvalid
	}

	identity, err := ical.CanonicalITIPIdentity(normalized)
	if err != nil {
		return CalendarInboundInviteRequest{}, ErrEventInvalid
	}
	if identity.CanonicalVersion != ical.CanonicalITIPVersion ||
		(identity.Method != ITIPMethodRequest && identity.Method != ITIPMethodCancel) ||
		identity.Sequence < 0 ||
		identity.Sequence > ical.MaxITIPSequence ||
		!validCalendarInboundAuthentication(authentication) {
		return CalendarInboundInviteRequest{}, ErrEventInvalid
	}

	return CalendarInboundInviteRequest{
		message:        normalized,
		identity:       identity,
		authentication: authentication,
	}, nil
}

func (request CalendarInboundInviteRequest) valid() bool {
	normalized, err := ical.NormalizeITIPMessage(request.message)
	if err != nil {
		return false
	}

	identity, err := ical.CanonicalITIPIdentity(normalized)
	if err != nil {
		return false
	}

	return identity == request.identity &&
		identity.CanonicalVersion == ical.CanonicalITIPVersion &&
		(identity.Method == ITIPMethodRequest || identity.Method == ITIPMethodCancel) &&
		identity.Sequence >= 0 &&
		identity.Sequence <= ical.MaxITIPSequence &&
		validCalendarInboundAuthentication(request.authentication)
}

func (request CalendarInboundInviteRequest) clone() (CalendarInboundInviteRequest, error) {
	return NewCalendarInboundInviteRequest(
		request.message,
		request.authentication,
	)
}

// CalendarInboundInviteAuthorization is opaque server-resolved authorization
// and must only be created by its constructor.
type CalendarInboundInviteAuthorization struct {
	ownerID         uuid.UUID
	current         *CalendarInboundObjectState
	series          *CalendarInboundObjectState
	resourceLimited bool
}

func NewCalendarInboundInviteAuthorization(
	ownerID uuid.UUID,
	current *CalendarInboundObjectState,
	series *CalendarInboundObjectState,
	resourceLimited bool,
) (CalendarInboundInviteAuthorization, error) {
	if ownerID == uuid.Nil || (resourceLimited && (current != nil || series != nil)) {
		return CalendarInboundInviteAuthorization{}, ErrEventInvalid
	}

	if current != nil {
		value := current.clone()
		current = &value
		if !current.valid() || current.Component().OwnerID() != ownerID {
			return CalendarInboundInviteAuthorization{}, ErrEventInvalid
		}
	}

	if series != nil {
		value := series.clone()
		series = &value
		if !series.valid() ||
			series.Component().OwnerID() != ownerID ||
			!series.Component().IsSeries() {
			return CalendarInboundInviteAuthorization{}, ErrEventInvalid
		}
	}

	return CalendarInboundInviteAuthorization{
		ownerID:         ownerID,
		current:         current,
		series:          series,
		resourceLimited: resourceLimited,
	}, nil
}

func (authorization CalendarInboundInviteAuthorization) valid() bool {
	_, err := NewCalendarInboundInviteAuthorization(
		authorization.ownerID,
		authorization.current,
		authorization.series,
		authorization.resourceLimited,
	)
	return err == nil
}

func (authorization CalendarInboundInviteAuthorization) clone() (
	CalendarInboundInviteAuthorization,
	error,
) {
	return NewCalendarInboundInviteAuthorization(
		authorization.ownerID,
		authorization.current,
		authorization.series,
		authorization.resourceLimited,
	)
}

type CalendarInboundInvitePersistenceCommand struct {
	request          CalendarInboundInviteRequest
	input            CalendarInboundIngestInput
	decision         CalendarInboundIngestResult
	outboxEnvelopeID uuid.UUID
}

func NewCalendarInboundInvitePersistenceCommand(
	request CalendarInboundInviteRequest,
	input CalendarInboundIngestInput,
	decision CalendarInboundIngestResult,
	outboxEnvelopeID uuid.UUID,
) (CalendarInboundInvitePersistenceCommand, error) {
	if !request.valid() || !input.valid() || !decision.valid() {
		return CalendarInboundInvitePersistenceCommand{}, ErrEventInvalid
	}

	switch decision.Outcome() {
	case CalendarInboundIngestApplyRequest,
		CalendarInboundIngestTombstoneOccurrence,
		CalendarInboundIngestTombstoneSeries,
		CalendarInboundIngestDeferred:
	default:
		return CalendarInboundInvitePersistenceCommand{}, ErrEventInvalid
	}

	if !calendarInboundInviteRequestMatchesInput(request, input) {
		return CalendarInboundInvitePersistenceCommand{}, ErrEventInvalid
	}

	switch decision.Outcome() {
	case CalendarInboundIngestApplyRequest:
		if outboxEnvelopeID == uuid.Nil {
			return CalendarInboundInvitePersistenceCommand{}, ErrEventInvalid
		}
	case CalendarInboundIngestDeferred:
		if outboxEnvelopeID != uuid.Nil {
			return CalendarInboundInvitePersistenceCommand{}, ErrEventInvalid
		}
	}

	clonedRequest, err := request.clone()
	if err != nil {
		return CalendarInboundInvitePersistenceCommand{}, ErrEventInvalid
	}

	clonedInput, err := NewCalendarInboundIngestInput(
		input.incoming,
		input.current,
		input.series,
	)
	if err != nil {
		return CalendarInboundInvitePersistenceCommand{}, ErrEventInvalid
	}

	return CalendarInboundInvitePersistenceCommand{
		request:          clonedRequest,
		input:            clonedInput,
		decision:         decision,
		outboxEnvelopeID: outboxEnvelopeID,
	}, nil
}

func (command CalendarInboundInvitePersistenceCommand) Outcome() CalendarInboundIngestOutcome {
	return command.decision.Outcome()
}

func (command CalendarInboundInvitePersistenceCommand) OutboxEnvelopeID() uuid.UUID {
	return command.outboxEnvelopeID
}

func (command CalendarInboundInvitePersistenceCommand) valid() bool {
	_, err := NewCalendarInboundInvitePersistenceCommand(
		command.request,
		command.input,
		command.decision,
		command.outboxEnvelopeID,
	)
	return err == nil
}

func (command CalendarInboundInvitePersistenceCommand) clone() (
	CalendarInboundInvitePersistenceCommand,
	error,
) {
	return NewCalendarInboundInvitePersistenceCommand(
		command.request,
		command.input,
		command.decision,
		command.outboxEnvelopeID,
	)
}

func calendarInboundInviteRequestMatchesInput(
	request CalendarInboundInviteRequest,
	input CalendarInboundIngestInput,
) bool {
	component := input.incoming.Component()
	payload := input.incoming.Payload()

	return component.UID() == request.identity.UID &&
		calendarInboundInviteSameRecurrence(
			request.identity.HasRecurrenceID,
			request.identity.RecurrenceID,
			component.RecurrenceID(),
		) &&
		component.IsSeries() == !request.identity.HasRecurrenceID &&
		request.message.Organizer.Email == input.incoming.OrganizerEmail() &&
		request.authentication == input.incoming.Authentication() &&
		request.identity.Method == input.incoming.Method() &&
		uint32(request.identity.Sequence) == input.incoming.Sequence() &&
		request.identity.CanonicalVersion == payload.CanonicalVersion() &&
		request.identity.PayloadHash == payload.Hash()
}

func calendarInboundInviteSameRecurrence(
	hasRecurrenceID bool,
	recurrenceID string,
	other *string,
) bool {
	if !hasRecurrenceID {
		return other == nil
	}
	return other != nil && *other == recurrenceID
}

type CalendarInboundInvitePersistenceResult struct {
	outcome          CalendarInboundIngestOutcome
	objectStateID    uuid.UUID
	calendarEventID  *uuid.UUID
	materialChanged  bool
	outboxPersisted  bool
}

func NewCalendarInboundInvitePersistenceResult(
	outcome CalendarInboundIngestOutcome,
	objectStateID uuid.UUID,
	calendarEventID *uuid.UUID,
	materialChanged bool,
	outboxPersisted bool,
) (CalendarInboundInvitePersistenceResult, error) {
	if objectStateID == uuid.Nil ||
		!validCalendarInboundMessageUUIDPointer(calendarEventID) ||
		materialChanged != outboxPersisted {
		return CalendarInboundInvitePersistenceResult{}, ErrEventInvalid
	}

	switch outcome {
	case CalendarInboundIngestApplyRequest:
		if calendarEventID == nil || !materialChanged || !outboxPersisted {
			return CalendarInboundInvitePersistenceResult{}, ErrEventInvalid
		}
	case CalendarInboundIngestTombstoneOccurrence,
		CalendarInboundIngestTombstoneSeries:
		if calendarEventID == nil && (materialChanged || outboxPersisted) {
			return CalendarInboundInvitePersistenceResult{}, ErrEventInvalid
		}
	case CalendarInboundIngestDeferred:
		if calendarEventID != nil || materialChanged || outboxPersisted {
			return CalendarInboundInvitePersistenceResult{}, ErrEventInvalid
		}
	default:
		return CalendarInboundInvitePersistenceResult{}, ErrEventInvalid
	}

	return CalendarInboundInvitePersistenceResult{
		outcome:          outcome,
		objectStateID:    objectStateID,
		calendarEventID:  cloneCalendarInboundMessageUUIDPointer(calendarEventID),
		materialChanged:  materialChanged,
		outboxPersisted:  outboxPersisted,
	}, nil
}

func (result CalendarInboundInvitePersistenceResult) Outcome() CalendarInboundIngestOutcome {
	return result.outcome
}

func (result CalendarInboundInvitePersistenceResult) ObjectStateID() uuid.UUID {
	return result.objectStateID
}

func (result CalendarInboundInvitePersistenceResult) CalendarEventID() *uuid.UUID {
	return cloneCalendarInboundMessageUUIDPointer(result.calendarEventID)
}

func (result CalendarInboundInvitePersistenceResult) MaterialChanged() bool {
	return result.materialChanged
}

func (result CalendarInboundInvitePersistenceResult) OutboxPersisted() bool {
	return result.outboxPersisted
}

func (result CalendarInboundInvitePersistenceResult) valid() bool {
	_, err := NewCalendarInboundInvitePersistenceResult(
		result.outcome,
		result.objectStateID,
		result.calendarEventID,
		result.materialChanged,
		result.outboxPersisted,
	)
	return err == nil
}

func (result CalendarInboundInvitePersistenceResult) clone() (
	CalendarInboundInvitePersistenceResult,
	error,
) {
	return NewCalendarInboundInvitePersistenceResult(
		result.outcome,
		result.objectStateID,
		result.calendarEventID,
		result.materialChanged,
		result.outboxPersisted,
	)
}

type CalendarInboundInviteTransactionBackend interface {
	ResolveCalendarInboundInvite(
		context.Context,
		CalendarInboundMessageScope,
		ITIPIdentity,
	) (CalendarInboundInviteAuthorization, error)

	PersistCalendarInboundInvite(
		context.Context,
		CalendarInboundInviteScope,
		CalendarInboundInvitePersistenceCommand,
	) (CalendarInboundInvitePersistenceResult, error)
}

type CalendarInboundInviteScope struct {
	messageScope  CalendarInboundMessageScope
	authorization CalendarInboundInviteAuthorization
	transaction   *EventTransaction
}

func (scope CalendarInboundInviteScope) isBoundTo(
	transaction *EventTransaction,
) bool {
	return transaction != nil &&
		scope.transaction == transaction &&
		scope.messageScope.isBoundTo(transaction) &&
		scope.authorization.valid()
}

func (transaction *EventTransaction) ResolveCalendarInboundInvite(
	ctx context.Context,
	messageScope CalendarInboundMessageScope,
	identity ITIPIdentity,
) (CalendarInboundInviteScope, error) {
	if transaction == nil {
		return CalendarInboundInviteScope{}, ErrEventTransactionClosed
	}
	if ctx == nil ||
		!messageScope.isBoundTo(transaction) ||
		!validCalendarInboundInviteIdentity(identity) {
		return CalendarInboundInviteScope{}, ErrEventInvalid
	}

	transaction.mu.Lock()
	defer transaction.mu.Unlock()

	if !transaction.active {
		return CalendarInboundInviteScope{}, ErrEventTransactionClosed
	}

	backend, ok := transaction.backend.(CalendarInboundInviteTransactionBackend)
	if !ok || calendarInboundInviteTransactionBackendTypedNil(backend) {
		return CalendarInboundInviteScope{}, ErrEventInvalid
	}

	authorization, err := backend.ResolveCalendarInboundInvite(
		ctx,
		messageScope,
		identity,
	)
	if err != nil {
		return CalendarInboundInviteScope{}, err
	}
	if !authorization.valid() {
		return CalendarInboundInviteScope{}, ErrEventInvalid
	}

	authorization, err = authorization.clone()
	if err != nil {
		return CalendarInboundInviteScope{}, ErrEventInvalid
	}

	return CalendarInboundInviteScope{
		messageScope:  messageScope,
		authorization: authorization,
		transaction:   transaction,
	}, nil
}

func (transaction *EventTransaction) PersistCalendarInboundInvite(
	ctx context.Context,
	scope CalendarInboundInviteScope,
	command CalendarInboundInvitePersistenceCommand,
) (CalendarInboundInvitePersistenceResult, error) {
	if transaction == nil {
		return CalendarInboundInvitePersistenceResult{}, ErrEventTransactionClosed
	}
	if ctx == nil || !scope.isBoundTo(transaction) || !command.valid() {
		return CalendarInboundInvitePersistenceResult{}, ErrEventInvalid
	}

	transaction.mu.Lock()
	defer transaction.mu.Unlock()

	if !transaction.active {
		return CalendarInboundInvitePersistenceResult{}, ErrEventTransactionClosed
	}

	backend, ok := transaction.backend.(CalendarInboundInviteTransactionBackend)
	if !ok || calendarInboundInviteTransactionBackendTypedNil(backend) {
		return CalendarInboundInvitePersistenceResult{}, ErrEventInvalid
	}

	clonedCommand, err := command.clone()
	if err != nil {
		return CalendarInboundInvitePersistenceResult{}, ErrEventInvalid
	}

	clonedAuthorization, err := scope.authorization.clone()
	if err != nil {
		return CalendarInboundInvitePersistenceResult{}, ErrEventInvalid
	}

	clonedScope := CalendarInboundInviteScope{
		messageScope:  scope.messageScope,
		authorization: clonedAuthorization,
		transaction:   transaction,
	}

	result, err := backend.PersistCalendarInboundInvite(
		ctx,
		clonedScope,
		clonedCommand,
	)
	if err != nil {
		return CalendarInboundInvitePersistenceResult{}, err
	}
	if !result.valid() ||
		result.Outcome() != command.Outcome() ||
		(result.MaterialChanged() && command.OutboxEnvelopeID() == uuid.Nil) {
		return CalendarInboundInvitePersistenceResult{}, ErrEventInvalid
	}

	clonedResult, err := result.clone()
	if err != nil {
		return CalendarInboundInvitePersistenceResult{}, ErrEventInvalid
	}
	return clonedResult, nil
}

func validCalendarInboundInviteIdentity(identity ITIPIdentity) bool {
	if identity.CanonicalVersion != ical.CanonicalITIPVersion ||
		(identity.Method != ITIPMethodRequest &&
			identity.Method != ITIPMethodCancel) ||
		identity.Sequence < 0 ||
		identity.Sequence > ical.MaxITIPSequence {
		return false
	}
	if err := ical.ValidateITIPUID(identity.UID); err != nil {
		return false
	}
	if !identity.HasRecurrenceID {
		return identity.RecurrenceID == ""
	}
	return identity.RecurrenceID != "" &&
		validCalendarInboundRecurrenceID(identity.RecurrenceID)
}

func calendarInboundInviteTransactionBackendTypedNil(
	backend CalendarInboundInviteTransactionBackend,
) bool {
	if backend == nil {
		return true
	}

	value := reflect.ValueOf(backend)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
