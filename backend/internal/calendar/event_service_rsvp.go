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

const calendarEventSeriesRSVPFingerprintVersion = 1

// CalendarEventSeriesRSVPInput contains one actor-bound RSVP mutation for a
// whole calendar series. The HTTP layer owns every request field beyond this
// frozen service input.
type CalendarEventSeriesRSVPInput struct {
	IdempotencyKey          string
	EventID                 uuid.UUID
	ExpectedAttendeeVersion uint
	Status                  RSVPStatus
}

// CalendarEventSeriesRSVPCommand snapshots normalized RSVP input and its
// receipt hashes. Its fields intentionally remain opaque outside calendar.
type CalendarEventSeriesRSVPCommand struct {
	input       CalendarEventSeriesRSVPInput
	keyHash     MutationReceiptHash
	requestHash MutationReceiptHash
}

// NewCalendarEventSeriesRSVPCommand validates one complete series RSVP
// request without resolving any identity or touching persistence.
func NewCalendarEventSeriesRSVPCommand(input CalendarEventSeriesRSVPInput) (CalendarEventSeriesRSVPCommand, error) {
	if !validCalendarCreateIdempotencyKey(input.IdempotencyKey) || input.EventID == uuid.Nil || input.ExpectedAttendeeVersion == 0 || input.ExpectedAttendeeVersion == ^uint(0) || !validCalendarEventSeriesRSVPTargetStatus(input.Status) {
		return CalendarEventSeriesRSVPCommand{}, ErrEventInvalid
	}
	requestHash, err := calendarEventSeriesRSVPRequestHash(input)
	if err != nil {
		return CalendarEventSeriesRSVPCommand{}, ErrEventInvalid
	}
	return CalendarEventSeriesRSVPCommand{
		input:       input,
		keyHash:     MutationReceiptHash(sha256.Sum256([]byte(input.IdempotencyKey))),
		requestHash: requestHash,
	}, nil
}

func (command CalendarEventSeriesRSVPCommand) valid() bool {
	rebuilt, err := NewCalendarEventSeriesRSVPCommand(command.input)
	return err == nil && rebuilt.keyHash == command.keyHash && rebuilt.requestHash == command.requestHash
}

// CalendarEventSeriesRSVPResult keeps the exact 200 receipt bytes and whether
// they came from a completed idempotency receipt.
type CalendarEventSeriesRSVPResult struct {
	response MutationReceiptResponse
	replayed bool
}

// Response returns the immutable canonical receipt response.
func (result CalendarEventSeriesRSVPResult) Response() MutationReceiptResponse { return result.response }

// Replayed reports whether Response came from a completed mutation receipt.
func (result CalendarEventSeriesRSVPResult) Replayed() bool { return result.replayed }

func (result CalendarEventSeriesRSVPResult) valid() bool { return result.response.valid() }

// CalendarEventSeriesRSVPConflictKind classifies only recoverable, public RSVP
// conflicts. All other failures remain content-free calendar errors.
type CalendarEventSeriesRSVPConflictKind string

const (
	CalendarEventSeriesRSVPAttendeeVersionConflict CalendarEventSeriesRSVPConflictKind = "attendee_version_conflict"
	CalendarEventSeriesRSVPIdempotencyKeyReused    CalendarEventSeriesRSVPConflictKind = "idempotency_key_reused"
)

// CalendarEventSeriesRSVPConflictError deliberately carries metadata only for
// a stale attendee version. Reused idempotency keys disclose no event state.
type CalendarEventSeriesRSVPConflictError struct {
	kind           CalendarEventSeriesRSVPConflictKind
	eventID        uuid.UUID
	currentVersion uint
}

func (err CalendarEventSeriesRSVPConflictError) Error() string {
	if err.kind == CalendarEventSeriesRSVPAttendeeVersionConflict || err.kind == CalendarEventSeriesRSVPIdempotencyKeyReused {
		return "calendar event series rsvp conflict"
	}
	return "calendar event series rsvp error"
}

func (err CalendarEventSeriesRSVPConflictError) Unwrap() error {
	if err.kind == CalendarEventSeriesRSVPAttendeeVersionConflict || err.kind == CalendarEventSeriesRSVPIdempotencyKeyReused {
		return ErrEventConflict
	}
	return nil
}

func (err CalendarEventSeriesRSVPConflictError) Kind() CalendarEventSeriesRSVPConflictKind {
	return err.kind
}

func (err CalendarEventSeriesRSVPConflictError) EventID() (uuid.UUID, bool) {
	if err.kind != CalendarEventSeriesRSVPAttendeeVersionConflict || err.eventID == uuid.Nil {
		return uuid.Nil, false
	}
	return err.eventID, true
}

func (err CalendarEventSeriesRSVPConflictError) CurrentVersion() (uint, bool) {
	if err.kind != CalendarEventSeriesRSVPAttendeeVersionConflict || err.currentVersion == 0 {
		return 0, false
	}
	return err.currentVersion, true
}

func calendarEventSeriesRSVPAttendeeVersionConflict(eventID uuid.UUID, currentVersion uint) CalendarEventSeriesRSVPConflictError {
	return CalendarEventSeriesRSVPConflictError{kind: CalendarEventSeriesRSVPAttendeeVersionConflict, eventID: eventID, currentVersion: currentVersion}
}

func calendarEventSeriesRSVPIdempotencyKeyReused() CalendarEventSeriesRSVPConflictError {
	return CalendarEventSeriesRSVPConflictError{kind: CalendarEventSeriesRSVPIdempotencyKeyReused}
}

type calendarEventSeriesRSVPRequestFingerprint struct {
	Version                 int        `json:"version"`
	EventID                 uuid.UUID  `json:"event_id"`
	ExpectedAttendeeVersion uint       `json:"expected_attendee_version"`
	Status                  RSVPStatus `json:"status"`
}

func calendarEventSeriesRSVPRequestHash(input CalendarEventSeriesRSVPInput) (MutationReceiptHash, error) {
	encoded, err := json.Marshal(calendarEventSeriesRSVPRequestFingerprint{
		Version:                 calendarEventSeriesRSVPFingerprintVersion,
		EventID:                 input.EventID,
		ExpectedAttendeeVersion: input.ExpectedAttendeeVersion,
		Status:                  input.Status,
	})
	if err != nil {
		return MutationReceiptHash{}, ErrEventInvalid
	}
	return MutationReceiptHash(sha256.Sum256(encoded)), nil
}

type calendarEventSeriesRSVPChangedPayload struct {
	EventID             uuid.UUID  `json:"event_id"`
	AttendeeID          uuid.UUID  `json:"attendee_id"`
	AttendeeEmail       string     `json:"attendee_email"`
	PreviousStatus      RSVPStatus `json:"previous_status"`
	NewStatus           RSVPStatus `json:"new_status"`
	Version             uint       `json:"version"`
	RecurrenceID        *string    `json:"recurrence_id"`
	ReplyDeliveryStatus string     `json:"reply_delivery_status"`
}

type calendarEventSeriesRSVPResponseBody struct {
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

// RespondToCalendarSeries atomically reserves or replays one receipt, resolves
// one actor-bound active attendee and persists exactly one RSVP CAS mutation.
func (service *EventService) RespondToCalendarSeries(ctx context.Context, actorID uuid.UUID, command CalendarEventSeriesRSVPCommand) (CalendarEventSeriesRSVPResult, error) {
	if service == nil || nilCalendarEventSeriesRSVPDependency(service.repository) || nilCalendarEventSeriesRSVPDependency(service.clock) || nilCalendarEventSeriesRSVPDependency(service.ids) || actorID == uuid.Nil || !command.valid() {
		return CalendarEventSeriesRSVPResult{}, ErrEventInvalid
	}
	var result CalendarEventSeriesRSVPResult
	err := service.repository.WithEventTransaction(ctx, func(transaction *EventTransaction) error {
		if transaction == nil {
			return ErrEventInvalid
		}
		scope, err := transaction.ResolveCalendarEventSeriesRSVPScope(ctx, actorID)
		if err != nil {
			return err
		}
		reservation, err := NewMutationReceiptReservation(scope, command.keyHash, command.requestHash)
		if err != nil {
			return err
		}
		reserved, err := transaction.ReserveMutationReceipt(ctx, scope, reservation)
		if err != nil {
			if calendarEventSeriesRSVPErrorContainsKind(err, EventErrorConflict) {
				return calendarEventSeriesRSVPIdempotencyKeyReused()
			}
			return err
		}
		switch reserved.State() {
		case MutationReceiptReplayed:
			response, found := reserved.ReplayedResponse()
			if !found {
				return ErrEventInvalid
			}
			result = CalendarEventSeriesRSVPResult{response: response, replayed: true}
			return nil
		case MutationReceiptReserved:
		default:
			return ErrEventInvalid
		}

		authorization, err := transaction.ResolveCalendarEventSeriesRSVP(ctx, scope, CalendarEventSeriesRSVPAuthorizationRequest{EventID: command.input.EventID})
		if err != nil {
			return err
		}
		if err := validateCalendarEventSeriesRSVPAuthorization(scope, command, authorization); err != nil {
			return err
		}
		current := authorization.Current
		if authorization.Attendee.Version != command.input.ExpectedAttendeeVersion {
			return calendarEventSeriesRSVPAttendeeVersionConflict(command.input.EventID, authorization.Attendee.Version)
		}

		occurredAt := service.clock.Now().UTC()
		if !validCalendarEventSeriesRSVPUTC(occurredAt) {
			return ErrEventInvalid
		}
		applicationEventID := service.ids.NewEventID()
		if applicationEventID == uuid.Nil {
			return ErrEventInvalid
		}
		updated := cloneCalendarEventSeriesRSVPAttendeeSnapshot(authorization.Attendee)
		previousStatus := updated.Status
		updated.Status = command.input.Status
		updated.Version++
		respondedAt := occurredAt
		updated.RespondedAt = &respondedAt

		deliveryStatus := "not_required"
		var reply *CalendarEventSeriesRSVPReplyDeliveryDescriptor
		if current.OrganizerID == nil {
			deliveryStatus = "pending"
			reply = &CalendarEventSeriesRSVPReplyDeliveryDescriptor{
				AttendeeID:      updated.AttendeeID,
				EventID:         updated.EventID,
				EventSequence:   current.Sequence,
				AttendeeVersion: updated.Version,
				Status:          updated.Status,
			}
		}
		actor := scope.ActorID()
		envelope, err := events.NewEnvelopeV1(
			"calendar.rsvp.changed.v1",
			applicationEventID,
			occurredAt,
			scope.OrganizationID(),
			&actor,
			current.EventID.String(),
			calendarEventSeriesRSVPChangedPayload{
				EventID: current.EventID, AttendeeID: updated.AttendeeID, AttendeeEmail: updated.Email,
				PreviousStatus: previousStatus, NewStatus: updated.Status, Version: updated.Version,
				RecurrenceID: nil, ReplyDeliveryStatus: deliveryStatus,
			},
		)
		if err != nil {
			return ErrEventInvalid
		}
		descriptor, err := NewEventOutboxDescriptor(envelope)
		if err != nil {
			return ErrEventInvalid
		}
		persistence, err := NewCalendarEventSeriesRSVPPersistenceCommand(updated, command.input.ExpectedAttendeeVersion, current.Sequence, descriptor, reply)
		if err != nil {
			return ErrEventInvalid
		}
		persisted, err := transaction.PersistCalendarEventSeriesRSVP(ctx, scope, persistence)
		if err != nil {
			return err
		}
		if !calendarEventSeriesRSVPPersistenceMatches(persistence, persisted) {
			return ErrEventInvalid
		}
		response, err := calendarEventSeriesRSVPResponse(updated, deliveryStatus)
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
		result = CalendarEventSeriesRSVPResult{response: response}
		return nil
	})
	if err != nil {
		return CalendarEventSeriesRSVPResult{}, calendarEventSeriesRSVPSafeError(err)
	}
	if !result.valid() {
		return CalendarEventSeriesRSVPResult{}, ErrEventInvalid
	}
	return result, nil
}

func validCalendarEventSeriesRSVPTargetStatus(status RSVPStatus) bool {
	return status == RSVPAccepted || status == RSVPDeclined || status == RSVPTentative
}

func validateCalendarEventSeriesRSVPAuthorization(scope EventMutationScope, command CalendarEventSeriesRSVPCommand, authorization CalendarEventSeriesRSVPAuthorization) error {
	current := authorization.Current
	attendee := authorization.Attendee
	if current.EventID != command.input.EventID || current.OrganizationID != scope.OrganizationID() || attendee.EventID != command.input.EventID || attendee.UserID == nil || *attendee.UserID != scope.ActorID() {
		return ErrEventNotFound
	}
	if current.Status == CalendarEventStatusCancelled {
		return ErrEventNotFound
	}
	if _, err := NewCalendarEventCommand(current); err != nil || !validCalendarEventUpdateAuthoritativeCurrent(current) || !validCalendarEventSeriesRSVPAttendee(attendee) {
		return ErrEventInvalid
	}
	return nil
}

func calendarEventSeriesRSVPResponse(attendee CalendarEventSeriesRSVPAttendeeSnapshot, deliveryStatus string) (MutationReceiptResponse, error) {
	if !validCalendarEventSeriesRSVPAttendee(attendee) || (deliveryStatus != "not_required" && deliveryStatus != "pending") {
		return MutationReceiptResponse{}, ErrEventInvalid
	}
	var displayName *string
	if attendee.DisplayName != "" {
		value := attendee.DisplayName
		displayName = &value
	}
	encoded, err := json.Marshal(calendarEventSeriesRSVPResponseBody{
		ID: attendee.AttendeeID, UserID: copyUUID(attendee.UserID), Email: attendee.Email, DisplayName: displayName,
		Role: attendee.Role, RSVPStatus: attendee.Status, RecurrenceID: nil, RespondedAt: cloneCalendarEventSeriesRSVPTime(attendee.RespondedAt),
		Version: attendee.Version, ReplyDeliveryStatus: deliveryStatus,
	})
	if err != nil || !utf8.Valid(encoded) {
		return MutationReceiptResponse{}, ErrEventInvalid
	}
	return NewMutationReceiptResponse(200, encoded)
}

func calendarEventSeriesRSVPSafeError(err error) (safe error) {
	defer func() {
		if recover() != nil {
			safe = ErrEventInvalid
		}
	}()
	chain, valid := calendarEventSeriesRSVPSafeErrorChain(err)
	if !valid {
		return ErrEventInvalid
	}
	for _, current := range chain {
		switch conflict := current.(type) {
		case CalendarEventSeriesRSVPConflictError:
			if validCalendarEventSeriesRSVPConflict(conflict) {
				return conflict
			}
		case *CalendarEventSeriesRSVPConflictError:
			if conflict != nil && validCalendarEventSeriesRSVPConflict(*conflict) {
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

func calendarEventSeriesRSVPErrorContainsKind(err error, wanted EventErrorKind) bool {
	chain, valid := calendarEventSeriesRSVPSafeErrorChain(err)
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

func validCalendarEventSeriesRSVPConflict(conflict CalendarEventSeriesRSVPConflictError) bool {
	switch conflict.kind {
	case CalendarEventSeriesRSVPAttendeeVersionConflict:
		return conflict.eventID != uuid.Nil && conflict.currentVersion != 0
	case CalendarEventSeriesRSVPIdempotencyKeyReused:
		return conflict.eventID == uuid.Nil && conflict.currentVersion == 0
	default:
		return false
	}
}

const calendarEventSeriesRSVPMaxErrorNodes = 64

func nilCalendarEventSeriesRSVPDependency(value any) bool {
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

func calendarEventSeriesRSVPSafeErrorChain(err error) (chain []error, valid bool) {
	defer func() {
		if recover() != nil {
			chain, valid = nil, false
		}
	}()
	pending := []error{err}
	visited := 0
	for len(pending) > 0 {
		if visited >= calendarEventSeriesRSVPMaxErrorNodes {
			return nil, false
		}
		index := len(pending) - 1
		current := pending[index]
		pending = pending[:index]
		visited++
		if current == nil {
			continue
		}
		if nilCalendarEventSeriesRSVPDependency(current) {
			return nil, false
		}
		chain = append(chain, current)
		if joined, multiple := current.(interface{ Unwrap() []error }); multiple {
			children := joined.Unwrap()
			if len(children) > calendarEventSeriesRSVPMaxErrorNodes-visited-len(pending) {
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
