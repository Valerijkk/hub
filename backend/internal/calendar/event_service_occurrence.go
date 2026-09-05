package calendar

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"reflect"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/platform/events"
)

const calendarEventOccurrenceRSVPFingerprintVersion = 1

// CalendarEventOccurrenceRSVPInput contains one actor-bound RSVP mutation for
// a single canonical occurrence. Transport-specific fields stay at the HTTP
// boundary and never cross into the durable occurrence port.
type CalendarEventOccurrenceRSVPInput struct {
	IdempotencyKey          string
	EventID                 uuid.UUID
	RecurrenceID            string
	ExpectedAttendeeVersion uint
	Status                  RSVPStatus
}

// CalendarEventOccurrenceRSVPCommand snapshots normalized occurrence RSVP
// ingress and the only hashes accepted by the mutation receipt boundary.
type CalendarEventOccurrenceRSVPCommand struct {
	input       CalendarEventOccurrenceRSVPInput
	keyHash     MutationReceiptHash
	requestHash MutationReceiptHash
}

// NewCalendarEventOccurrenceRSVPCommand validates a complete occurrence RSVP
// request without resolving identity, reading persistence or advancing time.
func NewCalendarEventOccurrenceRSVPCommand(input CalendarEventOccurrenceRSVPInput) (CalendarEventOccurrenceRSVPCommand, error) {
	if !validCalendarCreateIdempotencyKey(input.IdempotencyKey) || input.EventID == uuid.Nil || !validCalendarEventOccurrenceCancelCanonicalKey(input.RecurrenceID) || input.ExpectedAttendeeVersion == 0 || input.ExpectedAttendeeVersion == ^uint(0) || !validCalendarEventOccurrenceRSVPTargetStatus(input.Status) {
		return CalendarEventOccurrenceRSVPCommand{}, ErrEventInvalid
	}
	requestHash, err := calendarEventOccurrenceRSVPRequestHash(input)
	if err != nil {
		return CalendarEventOccurrenceRSVPCommand{}, ErrEventInvalid
	}
	return CalendarEventOccurrenceRSVPCommand{
		input:       input,
		keyHash:     MutationReceiptHash(sha256.Sum256([]byte(input.IdempotencyKey))),
		requestHash: requestHash,
	}, nil
}

func (command CalendarEventOccurrenceRSVPCommand) valid() bool {
	rebuilt, err := NewCalendarEventOccurrenceRSVPCommand(command.input)
	return err == nil && rebuilt.keyHash == command.keyHash && rebuilt.requestHash == command.requestHash
}

// CalendarEventOccurrenceRSVPResult contains the exact durable 200 bytes and
// whether those bytes came from an already completed mutation receipt.
type CalendarEventOccurrenceRSVPResult struct {
	response MutationReceiptResponse
	replayed bool
}

// Response returns the immutable canonical receipt response.
func (result CalendarEventOccurrenceRSVPResult) Response() MutationReceiptResponse { return result.response }

// Replayed reports whether Response came from a completed mutation receipt.
func (result CalendarEventOccurrenceRSVPResult) Replayed() bool { return result.replayed }

func (result CalendarEventOccurrenceRSVPResult) valid() bool { return result.response.valid() }

// CalendarEventOccurrenceRSVPConflictKind classifies only recoverable public
// occurrence RSVP conflicts. Its values deliberately align with series RSVP.
type CalendarEventOccurrenceRSVPConflictKind string

const (
	CalendarEventOccurrenceRSVPAttendeeVersionConflict CalendarEventOccurrenceRSVPConflictKind = "attendee_version_conflict"
	CalendarEventOccurrenceRSVPIdempotencyKeyReused    CalendarEventOccurrenceRSVPConflictKind = "idempotency_key_reused"
)

// CalendarEventOccurrenceRSVPConflictError carries recovery metadata only for
// a stale effective attendee version. Reused keys disclose no event state.
type CalendarEventOccurrenceRSVPConflictError struct {
	kind           CalendarEventOccurrenceRSVPConflictKind
	eventID        uuid.UUID
	currentVersion uint
}

func (err CalendarEventOccurrenceRSVPConflictError) Error() string {
	if err.kind == CalendarEventOccurrenceRSVPAttendeeVersionConflict || err.kind == CalendarEventOccurrenceRSVPIdempotencyKeyReused {
		return "calendar event occurrence rsvp conflict"
	}
	return "calendar event occurrence rsvp error"
}

func (err CalendarEventOccurrenceRSVPConflictError) Unwrap() error {
	if err.kind == CalendarEventOccurrenceRSVPAttendeeVersionConflict || err.kind == CalendarEventOccurrenceRSVPIdempotencyKeyReused {
		return ErrEventConflict
	}
	return nil
}

func (err CalendarEventOccurrenceRSVPConflictError) Kind() CalendarEventOccurrenceRSVPConflictKind {
	return err.kind
}

func (err CalendarEventOccurrenceRSVPConflictError) EventID() (uuid.UUID, bool) {
	if err.kind != CalendarEventOccurrenceRSVPAttendeeVersionConflict || err.eventID == uuid.Nil {
		return uuid.Nil, false
	}
	return err.eventID, true
}

func (err CalendarEventOccurrenceRSVPConflictError) CurrentVersion() (uint, bool) {
	if err.kind != CalendarEventOccurrenceRSVPAttendeeVersionConflict || err.currentVersion == 0 {
		return 0, false
	}
	return err.currentVersion, true
}

func calendarEventOccurrenceRSVPAttendeeVersionConflict(eventID uuid.UUID, currentVersion uint) CalendarEventOccurrenceRSVPConflictError {
	return CalendarEventOccurrenceRSVPConflictError{kind: CalendarEventOccurrenceRSVPAttendeeVersionConflict, eventID: eventID, currentVersion: currentVersion}
}

func calendarEventOccurrenceRSVPIdempotencyKeyReused() CalendarEventOccurrenceRSVPConflictError {
	return CalendarEventOccurrenceRSVPConflictError{kind: CalendarEventOccurrenceRSVPIdempotencyKeyReused}
}

type calendarEventOccurrenceRSVPRequestFingerprint struct {
	Version                 int        `json:"version"`
	EventID                 uuid.UUID  `json:"event_id"`
	RecurrenceID            string     `json:"recurrence_id"`
	ExpectedAttendeeVersion uint       `json:"expected_attendee_version"`
	Status                  RSVPStatus `json:"status"`
}

func calendarEventOccurrenceRSVPRequestHash(input CalendarEventOccurrenceRSVPInput) (MutationReceiptHash, error) {
	encoded, err := json.Marshal(calendarEventOccurrenceRSVPRequestFingerprint{
		Version:                 calendarEventOccurrenceRSVPFingerprintVersion,
		EventID:                 input.EventID,
		RecurrenceID:            input.RecurrenceID,
		ExpectedAttendeeVersion: input.ExpectedAttendeeVersion,
		Status:                  input.Status,
	})
	if err != nil {
		return MutationReceiptHash{}, ErrEventInvalid
	}
	return MutationReceiptHash(sha256.Sum256(encoded)), nil
}

type calendarEventOccurrenceRSVPChangedPayload struct {
	EventID             uuid.UUID  `json:"event_id"`
	AttendeeID          uuid.UUID  `json:"attendee_id"`
	AttendeeEmail       string     `json:"attendee_email"`
	PreviousStatus      RSVPStatus `json:"previous_status"`
	NewStatus           RSVPStatus `json:"new_status"`
	Version             uint       `json:"version"`
	RecurrenceID        string     `json:"recurrence_id"`
	ReplyDeliveryStatus string     `json:"reply_delivery_status"`
}

type calendarEventOccurrenceRSVPResponseBody struct {
	ID                  uuid.UUID                `json:"id"`
	UserID              *uuid.UUID               `json:"user_id"`
	Email               string                   `json:"email"`
	DisplayName         *string                  `json:"display_name"`
	Role                CalendarEventAttendeeRole `json:"role"`
	RSVPStatus          RSVPStatus               `json:"rsvp_status"`
	RecurrenceID        *string                  `json:"recurrence_id"`
	RespondedAt         *time.Time               `json:"responded_at"`
	Version             uint                     `json:"version"`
	ReplyDeliveryStatus string                   `json:"reply_delivery_status"`
}

// RespondToCalendarOccurrence atomically reserves or replays one receipt,
// resolves an actor-bound occurrence view and persists one occurrence-local
// RSVP CAS. It never mutates the inherited series attendee.
func (service *EventService) RespondToCalendarOccurrence(ctx context.Context, actorID uuid.UUID, command CalendarEventOccurrenceRSVPCommand) (CalendarEventOccurrenceRSVPResult, error) {
	if service == nil || nilCalendarEventOccurrenceRSVPDependency(service.repository) || nilCalendarEventOccurrenceRSVPDependency(service.clock) || nilCalendarEventOccurrenceRSVPDependency(service.ids) || actorID == uuid.Nil || !command.valid() {
		return CalendarEventOccurrenceRSVPResult{}, ErrEventInvalid
	}
	var result CalendarEventOccurrenceRSVPResult
	err := service.repository.WithEventTransaction(ctx, func(transaction *EventTransaction) error {
		if transaction == nil {
			return ErrEventInvalid
		}
		scope, err := transaction.ResolveCalendarEventOccurrenceRSVPScope(ctx, actorID)
		if err != nil {
			return err
		}
		reservation, err := NewMutationReceiptReservation(scope, command.keyHash, command.requestHash)
		if err != nil {
			return err
		}
		reserved, err := transaction.ReserveMutationReceipt(ctx, scope, reservation)
		if err != nil {
			if calendarEventOccurrenceRSVPErrorContainsKind(err, EventErrorConflict) {
				return calendarEventOccurrenceRSVPIdempotencyKeyReused()
			}
			return err
		}
		switch reserved.State() {
		case MutationReceiptReplayed:
			response, found := reserved.ReplayedResponse()
			if !found {
				return ErrEventInvalid
			}
			result = CalendarEventOccurrenceRSVPResult{response: response, replayed: true}
			return nil
		case MutationReceiptReserved:
		default:
			return ErrEventInvalid
		}

		authorization, err := transaction.ResolveCalendarEventOccurrenceRSVP(ctx, scope, CalendarEventOccurrenceRSVPAuthorizationRequest{EventID: command.input.EventID, RecurrenceID: command.input.RecurrenceID})
		if err != nil {
			return err
		}
		if err := validateCalendarEventOccurrenceRSVPAuthorization(scope, command, authorization); err != nil {
			return err
		}
		current := authorization.Current
		effectiveSequence := current.Sequence
		if authorization.Exception != nil {
			effectiveSequence = authorization.Exception.Sequence
		}

		prior := cloneCalendarEventOccurrenceRSVPAttendeeSnapshotPointer(authorization.OccurrenceAttendee)
		expectedVersion := authorization.SeriesAttendee.Version
		if prior != nil {
			expectedVersion = prior.Version
		}
		if expectedVersion == ^uint(0) {
			return ErrEventInvalid
		}
		if expectedVersion != command.input.ExpectedAttendeeVersion {
			return calendarEventOccurrenceRSVPAttendeeVersionConflict(command.input.EventID, expectedVersion)
		}

		occurredAt := service.clock.Now().UTC()
		if !validCalendarEventSeriesRSVPUTC(occurredAt) {
			return ErrEventInvalid
		}

		updated := CalendarEventOccurrenceRSVPAttendeeSnapshot{}
		if prior == nil {
			occurrenceAttendeeID := service.ids.NewEventID()
			if occurrenceAttendeeID == uuid.Nil || occurrenceAttendeeID == current.EventID || occurrenceAttendeeID == authorization.SeriesAttendee.AttendeeID {
				return ErrEventInvalid
			}
			updated = CalendarEventOccurrenceRSVPAttendeeSnapshot{
				AttendeeID: occurrenceAttendeeID, EventID: authorization.SeriesAttendee.EventID, UserID: copyUUID(authorization.SeriesAttendee.UserID),
				Email: authorization.SeriesAttendee.Email, DisplayName: authorization.SeriesAttendee.DisplayName, Role: authorization.SeriesAttendee.Role,
				Status: authorization.SeriesAttendee.Status, RecurrenceID: command.input.RecurrenceID, RespondedAt: cloneCalendarEventSeriesRSVPTime(authorization.SeriesAttendee.RespondedAt), Version: authorization.SeriesAttendee.Version,
			}
		} else {
			updated = cloneCalendarEventOccurrenceRSVPAttendeeSnapshot(*prior)
		}
		previousStatus := updated.Status
		updated.Status = command.input.Status
		updated.Version++
		respondedAt := occurredAt
		updated.RespondedAt = &respondedAt

		applicationEventID := service.ids.NewEventID()
		if applicationEventID == uuid.Nil || applicationEventID == current.EventID || applicationEventID == authorization.SeriesAttendee.AttendeeID || applicationEventID == updated.AttendeeID {
			return ErrEventInvalid
		}

		deliveryStatus := "not_required"
		var reply *CalendarEventOccurrenceRSVPReplyDeliveryDescriptor
		if current.OrganizerID == nil {
			deliveryStatus = "pending"
			reply = &CalendarEventOccurrenceRSVPReplyDeliveryDescriptor{
				AttendeeID: updated.AttendeeID, EventID: updated.EventID, EventSequence: effectiveSequence,
				AttendeeVersion: updated.Version, RecurrenceID: updated.RecurrenceID, Status: updated.Status,
			}
		}
		actor := scope.ActorID()
		envelope, err := events.NewEnvelopeV1(
			"calendar.rsvp.changed.v1", applicationEventID, occurredAt, scope.OrganizationID(), &actor, current.EventID.String(),
			calendarEventOccurrenceRSVPChangedPayload{
				EventID: current.EventID, AttendeeID: updated.AttendeeID, AttendeeEmail: updated.Email,
				PreviousStatus: previousStatus, NewStatus: updated.Status, Version: updated.Version,
				RecurrenceID: updated.RecurrenceID, ReplyDeliveryStatus: deliveryStatus,
			},
		)
		if err != nil {
			return ErrEventInvalid
		}
		descriptor, err := NewEventOutboxDescriptor(envelope)
		if err != nil {
			return ErrEventInvalid
		}
		persistence, err := NewCalendarEventOccurrenceRSVPPersistenceCommand(authorization.SeriesAttendee, prior, updated, expectedVersion, effectiveSequence, descriptor, reply)
		if err != nil {
			return ErrEventInvalid
		}
		persisted, err := transaction.PersistCalendarEventOccurrenceRSVP(ctx, scope, persistence)
		if err != nil {
			return err
		}
		if !calendarEventOccurrenceRSVPPersistenceMatches(persistence, persisted) {
			return ErrEventInvalid
		}
		response, err := calendarEventOccurrenceRSVPResponse(updated, deliveryStatus)
		if err != nil {
			return ErrEventInvalid
		}
		completion, err := NewMutationReceiptCompletion(reserved.ReceiptID(), &current.EventID, response)
		if err != nil {
			return ErrEventInvalid
		}
		if err := transaction.CompleteMutationReceipt(ctx, scope, completion); err != nil {
			return err
		}
		result = CalendarEventOccurrenceRSVPResult{response: response}
		return nil
	})
	if err != nil {
		return CalendarEventOccurrenceRSVPResult{}, calendarEventOccurrenceRSVPSafeError(err)
	}
	if !result.valid() {
		return CalendarEventOccurrenceRSVPResult{}, ErrEventInvalid
	}
	return result, nil
}

func validCalendarEventOccurrenceRSVPTargetStatus(status RSVPStatus) bool {
	return status == RSVPAccepted || status == RSVPDeclined || status == RSVPTentative
}

func validateCalendarEventOccurrenceRSVPAuthorization(scope EventMutationScope, command CalendarEventOccurrenceRSVPCommand, authorization CalendarEventOccurrenceRSVPAuthorization) error {
	current := authorization.Current
	series := authorization.SeriesAttendee
	if current.EventID != command.input.EventID || current.OrganizationID != scope.OrganizationID() || series.EventID != command.input.EventID || series.UserID == nil || *series.UserID != scope.ActorID() {
		return ErrEventNotFound
	}
	if current.Status == CalendarEventStatusCancelled {
		return ErrEventNotFound
	}
	if _, err := NewCalendarEventCommand(current); err != nil || !validCalendarEventUpdateAuthoritativeCurrent(current) || !validCalendarEventSeriesRSVPAttendee(series) {
		return ErrEventInvalid
	}
	if current.RRULE == "" || (current.Status != CalendarEventStatusConfirmed && current.Status != CalendarEventStatusTentative) || !calendarEventOccurrenceCancelMember(current, command.input.RecurrenceID) || containsCalendarEventOccurrenceCancelExdate(current.Exdates, command.input.RecurrenceID) {
		return ErrEventNotFound
	}
	if authorization.OccurrenceAttendee != nil && !validCalendarEventOccurrenceRSVPAttendeeForSeries(series, *authorization.OccurrenceAttendee) {
		return ErrEventInvalid
	}
	if authorization.Exception == nil {
		return nil
	}
	exception := *authorization.Exception
	if exception.EventID != current.EventID || exception.RecurrenceID != command.input.RecurrenceID || exception.ExceptionID == uuid.Nil {
		return ErrEventInvalid
	}
	if exception.Cancelled {
		if exception.Status != CalendarEventStatusCancelled {
			return ErrEventInvalid
		}
		return ErrEventNotFound
	}
	if exception.Status != CalendarEventStatusConfirmed && exception.Status != CalendarEventStatusTentative {
		return ErrEventInvalid
	}
	if exception.Temporal != nil && exception.Temporal.Validate() != nil {
		return ErrEventInvalid
	}
	return nil
}

func calendarEventOccurrenceRSVPResponse(attendee CalendarEventOccurrenceRSVPAttendeeSnapshot, deliveryStatus string) (MutationReceiptResponse, error) {
	if !validCalendarEventOccurrenceRSVPAttendee(attendee) || (deliveryStatus != "not_required" && deliveryStatus != "pending") {
		return MutationReceiptResponse{}, ErrEventInvalid
	}
	var displayName *string
	if attendee.DisplayName != "" {
		value := attendee.DisplayName
		displayName = &value
	}
	recurrenceID := attendee.RecurrenceID
	encoded, err := json.Marshal(calendarEventOccurrenceRSVPResponseBody{
		ID: attendee.AttendeeID, UserID: copyUUID(attendee.UserID), Email: attendee.Email, DisplayName: displayName,
		Role: attendee.Role, RSVPStatus: attendee.Status, RecurrenceID: &recurrenceID, RespondedAt: cloneCalendarEventSeriesRSVPTime(attendee.RespondedAt),
		Version: attendee.Version, ReplyDeliveryStatus: deliveryStatus,
	})
	if err != nil || !utf8.Valid(encoded) {
		return MutationReceiptResponse{}, ErrEventInvalid
	}
	return NewMutationReceiptResponse(200, encoded)
}

func calendarEventOccurrenceRSVPSafeError(err error) (safe error) {
	defer func() {
		if recover() != nil {
			safe = ErrEventInvalid
		}
	}()
	chain, valid := calendarEventOccurrenceRSVPSafeErrorChain(err)
	if !valid {
		return ErrEventInvalid
	}
	for _, current := range chain {
		switch conflict := current.(type) {
		case CalendarEventOccurrenceRSVPConflictError:
			if validCalendarEventOccurrenceRSVPConflict(conflict) {
				return conflict
			}
		case *CalendarEventOccurrenceRSVPConflictError:
			if conflict != nil && validCalendarEventOccurrenceRSVPConflict(*conflict) {
				return *conflict
			}
		}
	}
	for _, current := range chain {
		switch eventError := current.(type) {
		case EventError:
			switch eventError.Kind() {
			case EventErrorNotFound:
				return ErrEventNotFound
			case EventErrorForbidden:
				return ErrEventForbidden
			case EventErrorConflict:
				return ErrEventConflict
			case EventErrorTransactionClosed:
				return ErrEventTransactionClosed
			}
		case *EventError:
			if eventError == nil {
				return ErrEventInvalid
			}
			switch eventError.Kind() {
			case EventErrorNotFound:
				return ErrEventNotFound
			case EventErrorForbidden:
				return ErrEventForbidden
			case EventErrorConflict:
				return ErrEventConflict
			case EventErrorTransactionClosed:
				return ErrEventTransactionClosed
			}
		}
	}
	return ErrEventInvalid
}

func calendarEventOccurrenceRSVPErrorContainsKind(err error, wanted EventErrorKind) bool {
	chain, valid := calendarEventOccurrenceRSVPSafeErrorChain(err)
	if !valid {
		return false
	}
	for _, current := range chain {
		switch eventError := current.(type) {
		case EventError:
			if eventError.Kind() == wanted {
				return true
			}
		case *EventError:
			if eventError != nil && eventError.Kind() == wanted {
				return true
			}
		}
	}
	return false
}

func validCalendarEventOccurrenceRSVPConflict(conflict CalendarEventOccurrenceRSVPConflictError) bool {
	switch conflict.kind {
	case CalendarEventOccurrenceRSVPAttendeeVersionConflict:
		return conflict.eventID != uuid.Nil && conflict.currentVersion != 0
	case CalendarEventOccurrenceRSVPIdempotencyKeyReused:
		return conflict.eventID == uuid.Nil && conflict.currentVersion == 0
	default:
		return false
	}
}

const calendarEventOccurrenceRSVPMaxErrorNodes = 64

func nilCalendarEventOccurrenceRSVPDependency(value any) bool {
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

func calendarEventOccurrenceRSVPSafeErrorChain(err error) (chain []error, valid bool) {
	defer func() {
		if recover() != nil {
			chain, valid = nil, false
		}
	}()
	pending := []error{err}
	visited := 0
	for len(pending) > 0 {
		if visited >= calendarEventOccurrenceRSVPMaxErrorNodes {
			return nil, false
		}
		index := len(pending) - 1
		current := pending[index]
		pending = pending[:index]
		visited++
		if current == nil {
			continue
		}
		if nilCalendarEventOccurrenceRSVPDependency(current) {
			return nil, false
		}
		chain = append(chain, current)
		if joined, multiple := current.(interface{ Unwrap() []error }); multiple {
			children := joined.Unwrap()
			if len(children) > calendarEventOccurrenceRSVPMaxErrorNodes-visited-len(pending) {
				return nil, false
			}
			pending = append(pending, children...)
			continue
		}
		if wrapped, single := current.(interface{ Unwrap() error }); single {
			child := wrapped.Unwrap()
			if child != nil {
				pending = append(pending, child)
			}
		}
	}
	return chain, true
}
