package calendar

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"sort"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"

	ical "github.com/valerii/hub/backend/internal/ical"
	"github.com/valerii/hub/backend/internal/platform/events"
)

const calendarEventOccurrenceCancelMaxRecurrenceBytes = 255

// CalendarEventOccurrenceCancelInput identifies one canonical nominal
// recurrence occurrence and the authoritative master version expected by the
// caller.
type CalendarEventOccurrenceCancelInput struct {
	EventID         uuid.UUID
	RecurrenceID    string
	ExpectedVersion uint
}

// CalendarEventOccurrenceCancelCommand owns an immutable input snapshot.
type CalendarEventOccurrenceCancelCommand struct { input CalendarEventOccurrenceCancelInput }

func NewCalendarEventOccurrenceCancelCommand(input CalendarEventOccurrenceCancelInput) (CalendarEventOccurrenceCancelCommand, error) {
	if input.EventID == uuid.Nil || input.ExpectedVersion == 0 || !validCalendarEventOccurrenceCancelIngress(input.RecurrenceID) {
		return CalendarEventOccurrenceCancelCommand{}, ErrEventInvalid
	}
	return CalendarEventOccurrenceCancelCommand{input: input}, nil
}

func (command CalendarEventOccurrenceCancelCommand) EventID() uuid.UUID { return command.input.EventID }
func (command CalendarEventOccurrenceCancelCommand) RecurrenceID() string { return command.input.RecurrenceID }
func (command CalendarEventOccurrenceCancelCommand) ExpectedVersion() uint { return command.input.ExpectedVersion }
func (command CalendarEventOccurrenceCancelCommand) valid() bool {
	_, err := NewCalendarEventOccurrenceCancelCommand(command.input)
	return err == nil
}

type CalendarEventOccurrenceCancelPersistenceCommand struct {
	event           CalendarEventCommand
	expectedVersion uint
	exception       CalendarEventExceptionCommand
	outbox          EventOutboxDescriptor
}

func NewCalendarEventOccurrenceCancelPersistenceCommand(event CalendarEventCommand, expectedVersion uint, exception CalendarEventExceptionCommand, outbox EventOutboxDescriptor) (CalendarEventOccurrenceCancelPersistenceCommand, error) {
	if !validCalendarEventOccurrenceCancelPersistence(event, expectedVersion, exception, outbox, true) {
		return CalendarEventOccurrenceCancelPersistenceCommand{}, ErrEventInvalid
	}
	return CalendarEventOccurrenceCancelPersistenceCommand{event: cloneCalendarEventOccurrenceCancelEvent(event), expectedVersion: expectedVersion, exception: cloneCalendarEventOccurrenceCancelException(exception), outbox: outbox}, nil
}

func (command CalendarEventOccurrenceCancelPersistenceCommand) Event() CalendarEventCommand { return cloneCalendarEventOccurrenceCancelEvent(command.event) }
func (command CalendarEventOccurrenceCancelPersistenceCommand) ExpectedVersion() uint { return command.expectedVersion }
func (command CalendarEventOccurrenceCancelPersistenceCommand) Exception() CalendarEventExceptionCommand { return cloneCalendarEventOccurrenceCancelException(command.exception) }
func (command CalendarEventOccurrenceCancelPersistenceCommand) Outbox() EventOutboxDescriptor { return command.outbox }
func (command CalendarEventOccurrenceCancelPersistenceCommand) valid() bool {
	_, err := NewCalendarEventOccurrenceCancelPersistenceCommand(command.event, command.expectedVersion, command.exception, command.outbox)
	return err == nil
}

type CalendarEventOccurrenceCancelPersistenceResult struct {
	event     CalendarEventCommand
	exception CalendarEventExceptionCommand
	outbox    EventOutboxDescriptor
}

func NewCalendarEventOccurrenceCancelPersistenceResult(event CalendarEventCommand, exception CalendarEventExceptionCommand, outbox EventOutboxDescriptor) (CalendarEventOccurrenceCancelPersistenceResult, error) {
	if !validCalendarEventOccurrenceCancelPersistence(event, 0, exception, outbox, false) {
		return CalendarEventOccurrenceCancelPersistenceResult{}, ErrEventInvalid
	}
	return CalendarEventOccurrenceCancelPersistenceResult{event: cloneCalendarEventOccurrenceCancelEvent(event), exception: cloneCalendarEventOccurrenceCancelException(exception), outbox: outbox}, nil
}

func (result CalendarEventOccurrenceCancelPersistenceResult) Event() CalendarEventCommand { return cloneCalendarEventOccurrenceCancelEvent(result.event) }
func (result CalendarEventOccurrenceCancelPersistenceResult) Exception() CalendarEventExceptionCommand { return cloneCalendarEventOccurrenceCancelException(result.exception) }
func (result CalendarEventOccurrenceCancelPersistenceResult) Outbox() EventOutboxDescriptor { return result.outbox }
func (result CalendarEventOccurrenceCancelPersistenceResult) valid() bool {
	_, err := NewCalendarEventOccurrenceCancelPersistenceResult(result.event, result.exception, result.outbox)
	return err == nil
}

type CalendarEventOccurrenceCancelResult struct {
	eventID      uuid.UUID
	recurrenceID string
	status       CalendarEventStatus
	sequence     uint
	eventVersion uint
}

func NewCalendarEventOccurrenceCancelResult(eventID uuid.UUID, recurrenceID string, status CalendarEventStatus, sequence uint, eventVersion uint) (CalendarEventOccurrenceCancelResult, error) {
	if eventID == uuid.Nil || !validCalendarEventOccurrenceCancelCanonicalKey(recurrenceID) || status != CalendarEventStatusCancelled || sequence == 0 || eventVersion == 0 {
		return CalendarEventOccurrenceCancelResult{}, ErrEventInvalid
	}
	return CalendarEventOccurrenceCancelResult{eventID: eventID, recurrenceID: recurrenceID, status: status, sequence: sequence, eventVersion: eventVersion}, nil
}

func (result CalendarEventOccurrenceCancelResult) EventID() uuid.UUID { return result.eventID }
func (result CalendarEventOccurrenceCancelResult) RecurrenceID() string { return result.recurrenceID }
func (result CalendarEventOccurrenceCancelResult) Status() CalendarEventStatus { return result.status }
func (result CalendarEventOccurrenceCancelResult) Sequence() uint { return result.sequence }
func (result CalendarEventOccurrenceCancelResult) EventVersion() uint { return result.eventVersion }

type CalendarEventOccurrenceCancelConflictKind string

const (
	CalendarEventOccurrenceCancelVersionConflict CalendarEventOccurrenceCancelConflictKind = "version_conflict"
	CalendarEventOccurrenceCancelNotMutable      CalendarEventOccurrenceCancelConflictKind = "not_mutable"
)

type CalendarEventOccurrenceCancelConflictError struct {
	kind           CalendarEventOccurrenceCancelConflictKind
	eventID        uuid.UUID
	currentVersion uint
}

func (err CalendarEventOccurrenceCancelConflictError) Error() string { return "calendar event occurrence cancel conflict" }
func (err CalendarEventOccurrenceCancelConflictError) Unwrap() error { return ErrEventConflict }
func (err CalendarEventOccurrenceCancelConflictError) Kind() CalendarEventOccurrenceCancelConflictKind { return err.kind }
func (err CalendarEventOccurrenceCancelConflictError) EventID() (uuid.UUID, bool) {
	if err.kind != CalendarEventOccurrenceCancelVersionConflict || err.eventID == uuid.Nil { return uuid.Nil, false }
	return err.eventID, true
}
func (err CalendarEventOccurrenceCancelConflictError) CurrentVersion() (uint, bool) {
	if err.kind != CalendarEventOccurrenceCancelVersionConflict || err.currentVersion == 0 { return 0, false }
	return err.currentVersion, true
}

func calendarEventOccurrenceCancelVersionConflict(eventID uuid.UUID, currentVersion uint) CalendarEventOccurrenceCancelConflictError {
	return CalendarEventOccurrenceCancelConflictError{kind: CalendarEventOccurrenceCancelVersionConflict, eventID: eventID, currentVersion: currentVersion}
}
func calendarEventOccurrenceCancelNotMutable() CalendarEventOccurrenceCancelConflictError {
	return CalendarEventOccurrenceCancelConflictError{kind: CalendarEventOccurrenceCancelNotMutable}
}

type CalendarInviteOccurrenceCancelledPayload struct {
	EventID      uuid.UUID   `json:"event_id"`
	UID          string      `json:"uid"`
	Sequence     uint        `json:"sequence"`
	RecurrenceID string      `json:"recurrence_id"`
	AttendeeIDs  []uuid.UUID `json:"attendee_ids"`
}

func NewCalendarInviteOccurrenceCancelledEnvelopeV1(eventID uuid.UUID, occurredAt time.Time, organizationID uuid.UUID, actorID *uuid.UUID, correlationID string, payload CalendarInviteOccurrenceCancelledPayload) (events.EnvelopeV1, error) {
	if eventID == uuid.Nil || organizationID == uuid.Nil || actorID == nil || *actorID == uuid.Nil || !validCalendarEventOccurrenceCancelUTC(occurredAt) || correlationID != payload.EventID.String() || payload.EventID == uuid.Nil || payload.UID != "hub-"+payload.EventID.String() || payload.Sequence == 0 || !validCalendarEventOccurrenceCancelCanonicalKey(payload.RecurrenceID) || payload.AttendeeIDs == nil || len(payload.AttendeeIDs) > calendarCreateMaxAttendees {
		return events.EnvelopeV1{}, ErrEventInvalid
	}
	attendeeIDs := make([]uuid.UUID, len(payload.AttendeeIDs))
	copy(attendeeIDs, payload.AttendeeIDs)
	payload.AttendeeIDs = attendeeIDs
	for index, attendeeID := range payload.AttendeeIDs {
		if attendeeID == uuid.Nil || (index > 0 && bytes.Compare(payload.AttendeeIDs[index-1][:], attendeeID[:]) >= 0) {
			return events.EnvelopeV1{}, ErrEventInvalid
		}
	}
	envelope, err := events.NewEnvelopeV1("calendar.invite.cancelled.v1", eventID, occurredAt, organizationID, copyUUID(actorID), correlationID, payload)
	if err != nil { return events.EnvelopeV1{}, ErrEventInvalid }
	return envelope, nil
}

// CancelCalendarEventOccurrence cancels exactly one nominal recurrence
// occurrence. It has one atomic persistence/outbox boundary.
func (service *EventService) CancelCalendarEventOccurrence(ctx context.Context, actorID uuid.UUID, command CalendarEventOccurrenceCancelCommand) (CalendarEventOccurrenceCancelResult, error) {
	if service == nil || service.repository == nil || service.clock == nil || service.ids == nil || actorID == uuid.Nil || !command.valid() {
		return CalendarEventOccurrenceCancelResult{}, ErrEventInvalid
	}
	var result CalendarEventOccurrenceCancelResult
	err := service.repository.WithEventTransaction(ctx, func(transaction *EventTransaction) error {
		scope, err := transaction.ResolveCalendarEventOccurrenceCancelScope(ctx, actorID)
		if err != nil { return err }
		authorization, err := transaction.ResolveCalendarEventOccurrenceCancel(ctx, scope, CalendarEventOccurrenceCancelAuthorizationRequest{EventID: command.EventID(), RecurrenceID: command.RecurrenceID()})
		if err != nil { return err }
		if err := validateCalendarEventOccurrenceCancelAuthorization(scope, command, authorization); err != nil { return err }
		current := authorization.Current
		if current.Status != CalendarEventStatusConfirmed && current.Status != CalendarEventStatusTentative || current.Source != CalendarEventSourceHub || current.RRULE == "" || !calendarEventOccurrenceCancelMember(current, command.RecurrenceID()) {
			return calendarEventOccurrenceCancelNotMutable()
		}
		if containsCalendarEventOccurrenceCancelExdate(current.Exdates, command.RecurrenceID()) {
			return calendarEventOccurrenceCancelNotMutable()
		}
		if current.Version != command.ExpectedVersion() {
			return calendarEventOccurrenceCancelVersionConflict(command.EventID(), current.Version)
		}
		if authorization.Exception != nil {
			if authorization.Exception.Cancelled {
				result, err = NewCalendarEventOccurrenceCancelResult(current.EventID, command.RecurrenceID(), CalendarEventStatusCancelled, current.Sequence, current.Version)
				return err
			}
		}
		if current.Version == ^uint(0) || current.Sequence == ^uint(0) {
			return ErrEventInvalid
		}
		exceptionID := uuid.Nil
		if authorization.Exception != nil {
			exceptionID = authorization.Exception.ExceptionID
		} else {
			exceptionID = service.ids.NewEventID()
		}
		if exceptionID == uuid.Nil || exceptionID == current.EventID {
			return ErrEventInvalid
		}
		applicationEventID := service.ids.NewEventID()
		if applicationEventID == uuid.Nil || applicationEventID == current.EventID || applicationEventID == exceptionID {
			return ErrEventInvalid
		}
		updated := cloneCalendarEventCommandInput(current)
		updated.Version++
		updated.Sequence++
		event, err := NewCalendarEventCommand(updated)
		if err != nil { return ErrEventInvalid }
		exception, err := NewCalendarEventExceptionCommand(CalendarEventExceptionCommandInput{ExceptionID: exceptionID, EventID: current.EventID, RecurrenceID: command.RecurrenceID(), Sequence: updated.Sequence, Status: CalendarEventStatusCancelled, Cancelled: true})
		if err != nil { return ErrEventInvalid }
		attendeeIDs, err := calendarEventOccurrenceCancelAttendeeIDs(authorization.Attendees, current.EventID, scope.ActorID())
		if err != nil { return err }
		occurredAt := service.clock.Now()
		if !validCalendarEventOccurrenceCancelUTC(occurredAt) {
			return ErrEventInvalid
		}
		eventSnapshot := event.Snapshot()
		if eventSnapshot.EventID != current.EventID || eventSnapshot.OrganizerID == nil || *eventSnapshot.OrganizerID == uuid.Nil || eventSnapshot.Sequence != updated.Sequence {
			return ErrEventInvalid
		}
		actor := scope.ActorID()
		recurrenceID := command.RecurrenceID()
		envelope, err := NewCalendarInviteDeliveryEnvelopeV2(
			CalendarInviteCancelledSubjectV2, applicationEventID, occurredAt, scope.OrganizationID(), &actor, current.EventID.String(),
			CalendarInviteDeliveryEnvelopeV2Input{
				EventID: eventSnapshot.EventID, EventSequence: eventSnapshot.Sequence, OrganizerID: *eventSnapshot.OrganizerID,
				RecurrenceID: &recurrenceID, AttendeeIDs: attendeeIDs,
			},
		)
		if err != nil { return err }
		descriptor, err := NewEventOutboxDescriptor(envelope)
		if err != nil { return err }
		persistence, err := NewCalendarEventOccurrenceCancelPersistenceCommand(event, command.ExpectedVersion(), exception, descriptor)
		if err != nil { return err }
		persisted, err := transaction.PersistCalendarEventOccurrenceCancel(ctx, scope, persistence)
		if err != nil { return err }
		if !calendarEventOccurrenceCancelPersistenceMatches(persistence, persisted) { return ErrEventInvalid }
		result, err = NewCalendarEventOccurrenceCancelResult(updated.EventID, command.RecurrenceID(), CalendarEventStatusCancelled, updated.Sequence, updated.Version)
		return err
	})
	if err != nil { return CalendarEventOccurrenceCancelResult{}, calendarEventOccurrenceCancelSafeError(err) }
	return result, nil
}

func validCalendarEventOccurrenceCancelIngress(value string) bool {
	if len(value) == 0 || len(value) > calendarEventOccurrenceCancelMaxRecurrenceBytes || !utf8.ValidString(value) { return false }
	for _, value := range value {
		if unicode.IsControl(value) { return false }
	}
	return true
}

func calendarEventOccurrenceCancelParseTimedKey(value string) (time.Time, bool) {
	if len(value) != len("2006-01-02T15:04:05Z") { return time.Time{}, false }
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil || parsed.Location() != time.UTC || parsed.Year() < 0 || parsed.Year() > 9999 || parsed.Nanosecond() != 0 || parsed.Format("2006-01-02T15:04:05Z") != value {
		return time.Time{}, false
	}
	return parsed, true
}

func validCalendarEventOccurrenceCancelCanonicalKey(value string) bool {
	if !validCalendarEventOccurrenceCancelIngress(value) { return false }
	if len(value) == len(time.DateOnly) {
		parsed, err := time.Parse(time.DateOnly, value)
		return err == nil && parsed.Format(time.DateOnly) == value
	}
	_, valid := calendarEventOccurrenceCancelParseTimedKey(value)
	return valid
}

func validCalendarEventOccurrenceCancelUTC(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC
}

func validateCalendarEventOccurrenceCancelAuthorization(scope EventMutationScope, command CalendarEventOccurrenceCancelCommand, authorization CalendarEventOccurrenceCancelAuthorization) error {
	current := authorization.Current
	if current.EventID != command.EventID() || current.OrganizationID != scope.OrganizationID() || current.OwnerID != scope.ActorID() || current.OrganizerID == nil || *current.OrganizerID != scope.ActorID() {
		return ErrEventNotFound
	}
	if len(current.Exdates) > calendarCreateMaxExdates || len(authorization.Attendees) > calendarCreateMaxAttendees {
		return ErrEventInvalid
	}
	if _, err := NewCalendarEventCommand(current); err != nil || !validCalendarEventOccurrenceCancelCurrent(current) {
		return ErrEventInvalid
	}
	for _, attendee := range authorization.Attendees {
		if attendee.AttendeeID == uuid.Nil || attendee.EventID != current.EventID || attendee.Version == 0 || attendee.Status.Validate() != nil || attendee.UserID != nil && *attendee.UserID == uuid.Nil {
			return ErrEventInvalid
		}
		if attendee.RemovedAt != nil && !validCalendarEventOccurrenceCancelUTC(*attendee.RemovedAt) { return ErrEventInvalid }
	}
	if authorization.Exception != nil && !validCalendarEventOccurrenceCancelExistingException(*authorization.Exception, current, command.RecurrenceID()) {
		return ErrEventInvalid
	}
	return nil
}

func validCalendarEventOccurrenceCancelCurrent(current CalendarEventCommandInput) bool {
	if !validCalendarEventUpdateAuthoritativeCurrent(current) {
		return false
	}
	// A non-Hub source is authoritative data but immutable for this operation;
	// reject it in the policy branch, not as malformed current state. Hub keeps
	// the stronger UID binding required for local mutations.
	if current.Source == CalendarEventSourceHub && current.UID != "hub-"+current.EventID.String() {
		return false
	}
	if !current.Temporal.AllDay && (!validCalendarEventOccurrenceCancelUTC(current.Temporal.StartsAt) || !validCalendarEventOccurrenceCancelUTC(current.Temporal.EndsAt) || current.Temporal.StartsAt.Nanosecond() != 0 || current.Temporal.EndsAt.Nanosecond() != 0) {
		return false
	}
	return true
}

func calendarEventOccurrenceCancelMember(current CalendarEventCommandInput, recurrenceID string) bool {
	_, ok := ical.ResolveNominalOccurrence(current.Temporal, current.RRULE, recurrenceID)
	return ok
}

func containsCalendarEventOccurrenceCancelExdate(exdates []string, recurrenceID string) bool {
	for _, value := range exdates { if value == recurrenceID { return true } }
	return false
}

func validCalendarEventOccurrenceCancelExistingException(exception CalendarEventExceptionCommandInput, current CalendarEventCommandInput, recurrenceID string) bool {
	if exception.ExceptionID == uuid.Nil || exception.EventID != current.EventID || exception.RecurrenceID != recurrenceID || exception.Sequence == 0 || exception.Sequence > current.Sequence || !validCalendarEventStatus(exception.Status) || exception.Cancelled != (exception.Status == CalendarEventStatusCancelled) {
		return false
	}
	if exception.Cancelled {
		return exception.Temporal == nil
	}
	return exception.Temporal != nil && exception.Temporal.Validate() == nil
}

func validCalendarEventOccurrenceCancelPersistence(event CalendarEventCommand, expectedVersion uint, exception CalendarEventExceptionCommand, outbox EventOutboxDescriptor, checkExpectedVersion bool) bool {
	if !event.valid() || !exception.valid() || !outbox.valid() { return false }
	current := event.Snapshot()
	if !validCalendarEventOccurrenceCancelCurrent(current) || current.Source != CalendarEventSourceHub || (current.Status != CalendarEventStatusConfirmed && current.Status != CalendarEventStatusTentative) || current.Sequence == 0 || current.OwnerID == uuid.Nil || current.OrganizerID == nil || *current.OrganizerID != current.OwnerID {
		return false
	}
	if checkExpectedVersion && (expectedVersion == 0 || expectedVersion == ^uint(0) || current.Version != expectedVersion+1) {
		return false
	}
	value := exception.Snapshot()
	if value.ExceptionID == uuid.Nil || value.EventID != current.EventID || !validCalendarEventOccurrenceCancelCanonicalKey(value.RecurrenceID) || value.Sequence != current.Sequence || value.Status != CalendarEventStatusCancelled || !value.Cancelled || value.Temporal != nil {
		return false
	}
	if current.RRULE == "" || !calendarEventOccurrenceCancelMember(current, value.RecurrenceID) || containsCalendarEventOccurrenceCancelExdate(current.Exdates, value.RecurrenceID) {
		return false
	}
	return validCalendarEventOccurrenceCancelPersistenceOutbox(outbox, current, value)
}

func validCalendarEventOccurrenceCancelPersistenceOutbox(descriptor EventOutboxDescriptor, event CalendarEventCommandInput, exception CalendarEventExceptionCommandInput) bool {
	encoded, err := descriptor.Envelope().MarshalJSON()
	if err != nil { return false }
	var value calendarInviteDeliveryV2EnvelopeJSON
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&value) != nil || decoder.Decode(&struct{}{}) != io.EOF || value.Subject != CalendarInviteCancelledSubjectV2 || value.SchemaVersion != 1 || value.EventID == uuid.Nil || value.EventID == event.EventID || value.EventID == exception.ExceptionID || !validCalendarEventOccurrenceCancelUTC(value.OccurredAt) || value.OrganizationID != event.OrganizationID || value.ActorID == nil || *value.ActorID != event.OwnerID || value.CorrelationID != event.EventID.String() {
		return false
	}
	payload, err := DecodeCalendarInviteDeliveryEnvelopeV2(descriptor.Envelope())
	if err != nil || !reflect.DeepEqual(value.Payload, payload) || payload.EventID != event.EventID || payload.EventSequence != event.Sequence || event.OrganizerID == nil || payload.OrganizerID != *event.OrganizerID || payload.RecurrenceID == nil || *payload.RecurrenceID != exception.RecurrenceID || payload.Deliveries == nil || len(payload.Deliveries) > calendarCreateMaxAttendees {
		return false
	}
	return true
}

func calendarEventOccurrenceCancelAttendeeIDs(attendees []CalendarEventOccurrenceCancelAttendeeSnapshot, eventID uuid.UUID, organizerID uuid.UUID) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0, len(attendees))
	seen := make(map[uuid.UUID]struct{}, len(attendees))
	for _, attendee := range attendees {
		if attendee.AttendeeID == uuid.Nil || attendee.EventID != eventID || attendee.Version == 0 || attendee.Status.Validate() != nil || attendee.UserID != nil && *attendee.UserID == uuid.Nil { return nil, ErrEventInvalid }
		if attendee.RemovedAt != nil { continue }
		if attendee.UserID != nil && *attendee.UserID == organizerID { continue }
		if _, found := seen[attendee.AttendeeID]; found { return nil, ErrEventInvalid }
		seen[attendee.AttendeeID] = struct{}{}
		ids = append(ids, attendee.AttendeeID)
	}
	sort.Slice(ids, func(left, right int) bool { return bytes.Compare(ids[left][:], ids[right][:]) < 0 })
	return ids, nil
}

func calendarEventOccurrenceCancelPersistenceMatches(command CalendarEventOccurrenceCancelPersistenceCommand, result CalendarEventOccurrenceCancelPersistenceResult) bool {
	if !calendarEventOccurrenceCancelSameEvent(command.Event(), result.Event()) || !calendarEventOccurrenceCancelSameException(command.Exception(), result.Exception()) { return false }
	left, right := command.Outbox().Envelope(), result.Outbox().Envelope()
	if left.Subject() != right.Subject() || left.EventID() != right.EventID() { return false }
	leftJSON, leftErr := left.MarshalJSON()
	rightJSON, rightErr := right.MarshalJSON()
	return leftErr == nil && rightErr == nil && bytes.Equal(leftJSON, rightJSON)
}

func calendarEventOccurrenceCancelSameEvent(left, right CalendarEventCommand) bool {
	return reflect.DeepEqual(left.Snapshot(), right.Snapshot())
}

func calendarEventOccurrenceCancelSameException(left, right CalendarEventExceptionCommand) bool {
	return reflect.DeepEqual(left.Snapshot(), right.Snapshot())
}

func cloneCalendarEventOccurrenceCancelEvent(command CalendarEventCommand) CalendarEventCommand {
	cloned, err := NewCalendarEventCommand(command.Snapshot())
	if err != nil { return CalendarEventCommand{} }
	return cloned
}

func cloneCalendarEventOccurrenceCancelException(command CalendarEventExceptionCommand) CalendarEventExceptionCommand {
	cloned, err := NewCalendarEventExceptionCommand(command.Snapshot())
	if err != nil { return CalendarEventExceptionCommand{} }
	return cloned
}

func cloneCalendarEventOccurrenceCancelPersistenceCommand(command CalendarEventOccurrenceCancelPersistenceCommand) CalendarEventOccurrenceCancelPersistenceCommand {
	cloned, err := NewCalendarEventOccurrenceCancelPersistenceCommand(command.event, command.expectedVersion, command.exception, command.outbox)
	if err != nil { return CalendarEventOccurrenceCancelPersistenceCommand{} }
	return cloned
}

func cloneCalendarEventOccurrenceCancelPersistenceResult(result CalendarEventOccurrenceCancelPersistenceResult) CalendarEventOccurrenceCancelPersistenceResult {
	cloned, err := NewCalendarEventOccurrenceCancelPersistenceResult(result.event, result.exception, result.outbox)
	if err != nil { return CalendarEventOccurrenceCancelPersistenceResult{} }
	return cloned
}

func calendarEventOccurrenceCancelSafeError(err error) (safe error) {
	defer func() {
		if recover() != nil {
			safe = ErrEventInvalid
		}
	}()
	if malformedCalendarEventOccurrenceCancelErrorChain(err) { return ErrEventInvalid }
	var conflict CalendarEventOccurrenceCancelConflictError
	if errors.As(err, &conflict) && validCalendarEventOccurrenceCancelConflict(conflict) { return conflict }
	var pointer *CalendarEventOccurrenceCancelConflictError
	if errors.As(err, &pointer) && pointer != nil && validCalendarEventOccurrenceCancelConflict(*pointer) { return *pointer }
	switch {
	case errors.Is(err, ErrEventNotFound): return ErrEventNotFound
	case errors.Is(err, ErrEventTransactionClosed): return ErrEventTransactionClosed
	default: return ErrEventInvalid
	}
}

const (
	calendarEventOccurrenceCancelMaxEnvelopeJSONBytes = 65536
	calendarEventOccurrenceCancelMaxJSONDepth         = 8
	calendarEventOccurrenceCancelMaxJSONTokens        = 1024
	calendarEventOccurrenceCancelMaxJSONKeys          = 16
	calendarEventOccurrenceCancelMaxJSONArrayValues   = 256
	calendarEventOccurrenceCancelMaxErrorNodes        = 64
)

const (
	calendarEventOccurrenceCancelJSONGeneric          = iota
	calendarEventOccurrenceCancelJSONEnvelope
	calendarEventOccurrenceCancelJSONPayload
)

// validCalendarEventOccurrenceCancelStrictEnvelopeJSON rejects duplicate JSON
// keys before typed decoding. Decoder.DisallowUnknownFields alone deliberately
// accepts last-key-wins duplicates. The frozen occurrence envelope needs only
// a shallow outer object, payload object and attendee array, so bounded token
// and nesting limits are fail-closed ingress guards.
func validCalendarEventOccurrenceCancelStrictEnvelopeJSON(encoded []byte) bool {
	if len(encoded) == 0 || len(encoded) > calendarEventOccurrenceCancelMaxEnvelopeJSONBytes || !utf8.Valid(encoded) { return false }
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	tokens := 0
	read := func() (json.Token, bool) {
		if tokens >= calendarEventOccurrenceCancelMaxJSONTokens { return nil, false }
		tokens++
		value, err := decoder.Token()
		return value, err == nil
	}
	var value func(int, int) bool
	value = func(depth int, objectKind int) bool {
		if depth > calendarEventOccurrenceCancelMaxJSONDepth { return false }
		token, valid := read()
		if !valid { return false }
		delimiter, structured := token.(json.Delim)
		if !structured { return objectKind == calendarEventOccurrenceCancelJSONGeneric }
		switch delimiter {
		case '{':
			keys := make(map[string]struct{}, calendarEventOccurrenceCancelMaxJSONKeys)
			for decoder.More() {
				if len(keys) >= calendarEventOccurrenceCancelMaxJSONKeys { return false }
				rawKey, valid := read()
				key, stringKey := rawKey.(string)
				if !valid || !stringKey { return false }
				if _, duplicate := keys[key]; duplicate { return false }
				if !validCalendarEventOccurrenceCancelStrictEnvelopeKey(objectKind, key) { return false }
				keys[key] = struct{}{}
				childKind := calendarEventOccurrenceCancelJSONGeneric
				if objectKind == calendarEventOccurrenceCancelJSONEnvelope && key == "payload" {
					childKind = calendarEventOccurrenceCancelJSONPayload
				}
				if !value(depth + 1, childKind) { return false }
			}
			closing, valid := read()
			return valid && closing == json.Delim('}')
		case '[':
			count := 0
			for decoder.More() {
				count++
				if count > calendarEventOccurrenceCancelMaxJSONArrayValues || !value(depth + 1, calendarEventOccurrenceCancelJSONGeneric) { return false }
			}
			closing, valid := read()
			return valid && closing == json.Delim(']')
		default:
			return false
		}
	}
	if !value(0, calendarEventOccurrenceCancelJSONEnvelope) { return false }
	_, err := decoder.Token()
	return err == io.EOF
}

// Encoding/json folds field names while decoding a struct. The envelope is an
// explicit schema boundary instead: only exact JSON tags may appear, so a
// case-variant cannot survive frozen bytes as an ignored private field.
func validCalendarEventOccurrenceCancelStrictEnvelopeKey(objectKind int, key string) bool {
	switch objectKind {
	case calendarEventOccurrenceCancelJSONEnvelope:
		switch key {
		case "subject", "schema_version", "event_id", "occurred_at", "organization_id", "actor_id", "correlation_id", "payload":
			return true
		default:
			return false
		}
	case calendarEventOccurrenceCancelJSONPayload:
		switch key {
		case "event_id", "uid", "sequence", "recurrence_id", "attendee_ids":
			return true
		default:
			return false
		}
	default:
		return true
	}
}

// errors.As may invoke methods on a wrapped typed-nil error. Traverse both
// single and joined error chains before asking errors.As to inspect them. The
// bounded worklist turns attacker-controlled cycles, enormous joins and panic
// wrappers into a content-free invalid error.
func malformedCalendarEventOccurrenceCancelErrorChain(err error) (malformed bool) {
	defer func() {
		if recover() != nil { malformed = true }
	}()
	pending := []error{err}
	visited := 0
	for len(pending) > 0 {
		if visited >= calendarEventOccurrenceCancelMaxErrorNodes { return true }
		index := len(pending) - 1
		current := pending[index]
		pending = pending[:index]
		visited++
		if current == nil { continue }
		if nilCalendarEventOccurrenceCancelError(current) { return true }
		if joined, multiple := current.(interface{ Unwrap() []error }); multiple {
			children := joined.Unwrap()
			if len(children) > calendarEventOccurrenceCancelMaxErrorNodes-visited-len(pending) { return true }
			pending = append(pending, children...)
			continue
		}
		if wrapped, single := current.(interface{ Unwrap() error }); single {
			child := wrapped.Unwrap()
			if child != nil { pending = append(pending, child) }
		}
	}
	return false
}

func nilCalendarEventOccurrenceCancelError(err error) bool {
	if err == nil { return false }
	value := reflect.ValueOf(err)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func validCalendarEventOccurrenceCancelConflict(conflict CalendarEventOccurrenceCancelConflictError) bool {
	switch conflict.kind {
	case CalendarEventOccurrenceCancelVersionConflict:
		return conflict.eventID != uuid.Nil && conflict.currentVersion != 0
	case CalendarEventOccurrenceCancelNotMutable:
		return conflict.eventID == uuid.Nil && conflict.currentVersion == 0
	default:
		return false
	}
}
