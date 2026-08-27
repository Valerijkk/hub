package calendar

import (
	"bytes"
	"context"
	"errors"
	"sort"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/platform/events"
)

// CalendarEventUpdateString preserves the distinction between an omitted
// string field and an explicit empty string.
type CalendarEventUpdateString struct {
	Present bool
	Value   string
}

// CalendarEventUpdateBool preserves the distinction between an omitted bool
// field and an explicit false value.
type CalendarEventUpdateBool struct {
	Present bool
	Value   bool
}

// CalendarEventUpdateTime preserves an omitted time field versus explicit
// null, which is validated only after merging with the authoritative event.
type CalendarEventUpdateTime struct {
	Present bool
	Value   *time.Time
}

type CalendarEventUpdateDate struct {
	Present bool
	Value   *string
}

type CalendarEventUpdateRRULE struct {
	Present bool
	Value   *string
}

type CalendarEventUpdateInt struct {
	Present bool
	Value   *int
}

type CalendarEventUpdateExdates struct {
	Present bool
	Value   []string
}

type CalendarEventUpdateAttendees struct {
	Present bool
	Value   []CalendarEventCreateAttendeeCandidate
}

// CalendarEventUpdateInput is the untrusted, presence-preserving master event
// patch. Identity, organization, source and current state are never accepted
// from this input.
type CalendarEventUpdateInput struct {
	EventID          uuid.UUID
	ExpectedVersion  uint
	NotifyAttendees  bool
	Title            CalendarEventUpdateString
	Description      CalendarEventUpdateString
	Location         CalendarEventUpdateString
	StartsAt         CalendarEventUpdateTime
	EndsAt           CalendarEventUpdateTime
	StartDate        CalendarEventUpdateDate
	EndDateExclusive CalendarEventUpdateDate
	AllDay           CalendarEventUpdateBool
	Timezone         CalendarEventUpdateString
	RRULE            CalendarEventUpdateRRULE
	Exdates          CalendarEventUpdateExdates
	ReminderMinutes  CalendarEventUpdateInt
	Attendees        CalendarEventUpdateAttendees
}

// CalendarEventUpdateCommand is an immutable normalized patch.
type CalendarEventUpdateCommand struct {
	input CalendarEventUpdateInput
}

func NewCalendarEventUpdateCommand(input CalendarEventUpdateInput) (CalendarEventUpdateCommand, error) {
	if input.EventID == uuid.Nil || input.ExpectedVersion == 0 || !calendarEventUpdatePresenceValid(input) || !calendarEventUpdateHasPatch(input) {
		return CalendarEventUpdateCommand{}, ErrEventInvalid
	}
	if (input.Title.Present && !validCalendarCreateText(input.Title.Value, calendarCreateMaxTitleBytes, true)) ||
		(input.Description.Present && !validCalendarCreateText(input.Description.Value, calendarCreateMaxDescription, false)) ||
		(input.Location.Present && !validCalendarCreateText(input.Location.Value, calendarCreateMaxLocationBytes, false)) ||
		(input.Timezone.Present && (!utf8.ValidString(input.Timezone.Value) || len(input.Timezone.Value) > calendarCreateMaxTimezoneBytes)) {
		return CalendarEventUpdateCommand{}, ErrEventInvalid
	}
	if input.RRULE.Present && input.RRULE.Value != nil && (!utf8.ValidString(*input.RRULE.Value) || len(*input.RRULE.Value) > calendarCreateMaxRRuleBytes) {
		return CalendarEventUpdateCommand{}, ErrEventInvalid
	}
	if (input.Exdates.Present && input.Exdates.Value == nil) || (input.Attendees.Present && input.Attendees.Value == nil) {
		return CalendarEventUpdateCommand{}, ErrEventInvalid
	}
	if input.Exdates.Present && len(input.Exdates.Value) > calendarCreateMaxExdates {
		return CalendarEventUpdateCommand{}, ErrEventInvalid
	}
	if input.ReminderMinutes.Present && input.ReminderMinutes.Value != nil && (*input.ReminderMinutes.Value < 0 || *input.ReminderMinutes.Value > 40320) {
		return CalendarEventUpdateCommand{}, ErrEventInvalid
	}
	if input.Attendees.Present {
		attendees, err := normalizeCalendarEventUpdateCandidates(input.Attendees.Value)
		if err != nil {
			return CalendarEventUpdateCommand{}, ErrEventInvalid
		}
		input.Attendees.Value = attendees
	}
	return CalendarEventUpdateCommand{input: cloneCalendarEventUpdateInput(input)}, nil
}

func (command CalendarEventUpdateCommand) valid() bool {
	_, err := NewCalendarEventUpdateCommand(command.input)
	return err == nil
}

// CalendarEventUpdateAttendeeSnapshot is authoritative active attendee state
// supplied by the update resolver.
type CalendarEventUpdateAttendeeSnapshot struct {
	AttendeeID  uuid.UUID
	UserID      *uuid.UUID
	Email       string
	DisplayName string
	Role        CalendarEventAttendeeRole
	Status      RSVPStatus
	RespondedAt *time.Time
	Version     uint
}

// CalendarEventUpdateAuthorization is the only read model accepted by update
// and cancel. It is returned by the in-transaction resolver, not by clients.
type CalendarEventUpdateAuthorization struct {
	Current                  CalendarEventCommandInput
	Attendees                []CalendarEventUpdateAttendeeSnapshot
	ResolvedAttendees        []CalendarEventCreateResolvedAttendee
	HasMaterializedException bool
}

type CalendarEventUpdateAttendeeMutationKind string

const (
	CalendarEventUpdateAttendeeAdd    CalendarEventUpdateAttendeeMutationKind = "add"
	CalendarEventUpdateAttendeeReset  CalendarEventUpdateAttendeeMutationKind = "reset"
	CalendarEventUpdateAttendeeRemove CalendarEventUpdateAttendeeMutationKind = "remove"
)

// CalendarEventUpdateAttendeeMutation is a complete immutable attendee write
// instruction. Removal retains historical identity and RSVP state.
type CalendarEventUpdateAttendeeMutation struct {
	Kind        CalendarEventUpdateAttendeeMutationKind
	AttendeeID  uuid.UUID
	UserID      *uuid.UUID
	Email       string
	DisplayName string
	Role        CalendarEventAttendeeRole
	Status      RSVPStatus
	RespondedAt *time.Time
	Version     uint
	RemovedAt   *time.Time
}

// CalendarEventUpdatePersistenceCommand is the opaque all-or-nothing update
// boundary passed to an adapter that opted in to the update transaction port.
type CalendarEventUpdatePersistenceCommand struct {
	event           CalendarEventCommand
	expectedVersion uint
	attendees       []CalendarEventUpdateAttendeeMutation
	outbox          []EventOutboxDescriptor
}

func NewCalendarEventUpdatePersistenceCommand(event CalendarEventCommand, expectedVersion uint, attendees []CalendarEventUpdateAttendeeMutation, outbox []EventOutboxDescriptor) (CalendarEventUpdatePersistenceCommand, error) {
	if expectedVersion == 0 || !event.valid() || event.Version() != expectedVersion+1 {
		return CalendarEventUpdatePersistenceCommand{}, ErrEventInvalid
	}
	copiedEvent, err := NewCalendarEventCommand(event.input)
	if err != nil {
		return CalendarEventUpdatePersistenceCommand{}, ErrEventInvalid
	}
	copiedAttendees := make([]CalendarEventUpdateAttendeeMutation, len(attendees))
	for index, mutation := range attendees {
		if !validCalendarEventUpdateAttendeeMutation(mutation) {
			return CalendarEventUpdatePersistenceCommand{}, ErrEventInvalid
		}
		copiedAttendees[index] = cloneCalendarEventUpdateAttendeeMutation(mutation)
	}
	sortCalendarEventUpdateMutations(copiedAttendees)
	if duplicateCalendarEventUpdateMutationIDs(copiedAttendees) {
		return CalendarEventUpdatePersistenceCommand{}, ErrEventInvalid
	}
	copiedOutbox := append([]EventOutboxDescriptor(nil), outbox...)
	for _, descriptor := range copiedOutbox {
		if !descriptor.valid() {
			return CalendarEventUpdatePersistenceCommand{}, ErrEventInvalid
		}
	}
	return CalendarEventUpdatePersistenceCommand{event: copiedEvent, expectedVersion: expectedVersion, attendees: copiedAttendees, outbox: copiedOutbox}, nil
}

func (command CalendarEventUpdatePersistenceCommand) Event() CalendarEventCommand { return command.event }
func (command CalendarEventUpdatePersistenceCommand) ExpectedVersion() uint { return command.expectedVersion }
func (command CalendarEventUpdatePersistenceCommand) Attendees() []CalendarEventUpdateAttendeeMutation {
	copyAttendees := make([]CalendarEventUpdateAttendeeMutation, len(command.attendees))
	for index, mutation := range command.attendees { copyAttendees[index] = cloneCalendarEventUpdateAttendeeMutation(mutation) }
	return copyAttendees
}
func (command CalendarEventUpdatePersistenceCommand) Outbox() []EventOutboxDescriptor { return append([]EventOutboxDescriptor(nil), command.outbox...) }
func (command CalendarEventUpdatePersistenceCommand) valid() bool {
	_, err := NewCalendarEventUpdatePersistenceCommand(command.event, command.expectedVersion, command.attendees, command.outbox)
	return err == nil
}

type CalendarEventUpdateAttendeePersistenceResult struct {
	AttendeeID uuid.UUID
	Version    uint
	Removed    bool
}

// CalendarEventUpdatePersistenceResult is the immutable adapter echo checked by
// the service before the enclosing transaction is allowed to commit.
type CalendarEventUpdatePersistenceResult struct {
	eventID    uuid.UUID
	version    uint
	sequence   uint
	attendees  []CalendarEventUpdateAttendeePersistenceResult
}

func NewCalendarEventUpdatePersistenceResult(eventID uuid.UUID, version uint, sequence uint, attendees []CalendarEventUpdateAttendeePersistenceResult) (CalendarEventUpdatePersistenceResult, error) {
	if eventID == uuid.Nil || version == 0 {
		return CalendarEventUpdatePersistenceResult{}, ErrEventInvalid
	}
	copied := append([]CalendarEventUpdateAttendeePersistenceResult(nil), attendees...)
	for _, attendee := range copied {
		if attendee.AttendeeID == uuid.Nil || attendee.Version == 0 {
			return CalendarEventUpdatePersistenceResult{}, ErrEventInvalid
		}
	}
	sort.Slice(copied, func(left, right int) bool { return bytes.Compare(copied[left].AttendeeID[:], copied[right].AttendeeID[:]) < 0 })
	for index := 1; index < len(copied); index++ {
		if copied[index-1].AttendeeID == copied[index].AttendeeID { return CalendarEventUpdatePersistenceResult{}, ErrEventInvalid }
	}
	return CalendarEventUpdatePersistenceResult{eventID: eventID, version: version, sequence: sequence, attendees: copied}, nil
}

func (result CalendarEventUpdatePersistenceResult) EventID() uuid.UUID { return result.eventID }
func (result CalendarEventUpdatePersistenceResult) Version() uint { return result.version }
func (result CalendarEventUpdatePersistenceResult) Sequence() uint { return result.sequence }
func (result CalendarEventUpdatePersistenceResult) Attendees() []CalendarEventUpdateAttendeePersistenceResult { return append([]CalendarEventUpdateAttendeePersistenceResult(nil), result.attendees...) }
func (result CalendarEventUpdatePersistenceResult) valid() bool {
	_, err := NewCalendarEventUpdatePersistenceResult(result.eventID, result.version, result.sequence, result.attendees)
	return err == nil
}

type CalendarEventChangedField string

const (
	CalendarEventChangedFieldTitle            CalendarEventChangedField = "title"
	CalendarEventChangedFieldDescription      CalendarEventChangedField = "description"
	CalendarEventChangedFieldLocation         CalendarEventChangedField = "location"
	CalendarEventChangedFieldStartsAt         CalendarEventChangedField = "starts_at"
	CalendarEventChangedFieldEndsAt           CalendarEventChangedField = "ends_at"
	CalendarEventChangedFieldStartDate        CalendarEventChangedField = "start_date"
	CalendarEventChangedFieldEndDateExclusive CalendarEventChangedField = "end_date_exclusive"
	CalendarEventChangedFieldAllDay           CalendarEventChangedField = "all_day"
	CalendarEventChangedFieldTimezone         CalendarEventChangedField = "timezone"
	CalendarEventChangedFieldRRULE            CalendarEventChangedField = "rrule"
	CalendarEventChangedFieldExdates          CalendarEventChangedField = "exdates"
	CalendarEventChangedFieldReminderMinutes  CalendarEventChangedField = "reminder_minutes"
	CalendarEventChangedFieldAttendees        CalendarEventChangedField = "attendees"
)

// CalendarEventUpdateResult is the public, copy-safe service result.
type CalendarEventUpdateResult struct {
	eventID       uuid.UUID
	version       uint
	sequence      uint
	status        CalendarEventStatus
	changedFields []CalendarEventChangedField
}

func NewCalendarEventUpdateResult(eventID uuid.UUID, version uint, sequence uint, status CalendarEventStatus, changedFields []CalendarEventChangedField) (CalendarEventUpdateResult, error) {
	if eventID == uuid.Nil || version == 0 || !validCalendarEventStatus(status) || !validCalendarEventChangedFields(changedFields) {
		return CalendarEventUpdateResult{}, ErrEventInvalid
	}
	return CalendarEventUpdateResult{eventID: eventID, version: version, sequence: sequence, status: status, changedFields: append([]CalendarEventChangedField(nil), changedFields...)}, nil
}

func (result CalendarEventUpdateResult) EventID() uuid.UUID { return result.eventID }
func (result CalendarEventUpdateResult) Version() uint { return result.version }
func (result CalendarEventUpdateResult) Sequence() uint { return result.sequence }
func (result CalendarEventUpdateResult) Status() CalendarEventStatus { return result.status }
func (result CalendarEventUpdateResult) ChangedFields() []CalendarEventChangedField { return append([]CalendarEventChangedField(nil), result.changedFields...) }

type CalendarEventUpdateConflictKind string

const (
	CalendarEventUpdateVersionConflict             CalendarEventUpdateConflictKind = "version_conflict"
	CalendarEventUpdateNotMutable                  CalendarEventUpdateConflictKind = "not_mutable"
	CalendarEventUpdateRecurrenceExceptionConflict CalendarEventUpdateConflictKind = "recurrence_exception_conflict"
)

// CalendarEventUpdateConflictError contains only machine-classifiable state;
// its public message deliberately carries no event, organization or user data.
type CalendarEventUpdateConflictError struct {
	kind           CalendarEventUpdateConflictKind
	eventID        uuid.UUID
	currentVersion uint
}

func (err CalendarEventUpdateConflictError) Error() string { return "calendar event update conflict" }
func (err CalendarEventUpdateConflictError) Unwrap() error { return ErrEventConflict }
func (err CalendarEventUpdateConflictError) Kind() CalendarEventUpdateConflictKind { return err.kind }
func (err CalendarEventUpdateConflictError) EventID() (uuid.UUID, bool) {
	if err.kind != CalendarEventUpdateVersionConflict || err.eventID == uuid.Nil { return uuid.Nil, false }
	return err.eventID, true
}
func (err CalendarEventUpdateConflictError) CurrentVersion() (uint, bool) {
	if err.kind != CalendarEventUpdateVersionConflict || err.currentVersion == 0 { return 0, false }
	return err.currentVersion, true
}

func calendarEventUpdateVersionConflict(eventID uuid.UUID, currentVersion uint) CalendarEventUpdateConflictError {
	return CalendarEventUpdateConflictError{kind: CalendarEventUpdateVersionConflict, eventID: eventID, currentVersion: currentVersion}
}
func calendarEventUpdateNotMutable() CalendarEventUpdateConflictError { return CalendarEventUpdateConflictError{kind: CalendarEventUpdateNotMutable} }
func calendarEventUpdateRecurrenceExceptionConflict() CalendarEventUpdateConflictError { return CalendarEventUpdateConflictError{kind: CalendarEventUpdateRecurrenceExceptionConflict} }

// CalendarInviteUpdatedPayload is the content-free notification intent for an
// invitation update. It intentionally has no iTIP or rendered body field.
type CalendarInviteUpdatedPayload struct {
	EventID       uuid.UUID                   `json:"event_id"`
	UID           string                      `json:"uid"`
	Sequence      uint                        `json:"sequence"`
	AttendeeIDs   []uuid.UUID                 `json:"attendee_ids"`
	ResetRSVP     bool                        `json:"reset_rsvp"`
	ChangedFields []CalendarEventChangedField `json:"changed_fields"`
}

// CalendarInviteCancelledPayload is the content-free notification intent for
// a master cancellation.
type CalendarInviteCancelledPayload struct {
	EventID     uuid.UUID   `json:"event_id"`
	UID         string      `json:"uid"`
	Sequence    uint        `json:"sequence"`
	AttendeeIDs []uuid.UUID `json:"attendee_ids"`
}

func NewCalendarInviteUpdatedEnvelopeV1(eventID uuid.UUID, occurredAt time.Time, organizationID uuid.UUID, actorID *uuid.UUID, correlationID string, payload CalendarInviteUpdatedPayload) (events.EnvelopeV1, error) {
	if !validCalendarInviteEnvelopeBase(eventID, occurredAt, organizationID, actorID, payload.EventID, payload.UID, payload.AttendeeIDs) || !validCalendarEventChangedFieldsNonEmpty(payload.ChangedFields) {
		return events.EnvelopeV1{}, ErrEventInvalid
	}
	payload.AttendeeIDs = append([]uuid.UUID(nil), payload.AttendeeIDs...)
	payload.ChangedFields = append([]CalendarEventChangedField(nil), payload.ChangedFields...)
	envelope, err := events.NewEnvelopeV1("calendar.invite.updated.v1", eventID, occurredAt.UTC(), organizationID, copyUUID(actorID), correlationID, payload)
	if err != nil { return events.EnvelopeV1{}, ErrEventInvalid }
	return envelope, nil
}

func NewCalendarInviteCancelledEnvelopeV1(eventID uuid.UUID, occurredAt time.Time, organizationID uuid.UUID, actorID *uuid.UUID, correlationID string, payload CalendarInviteCancelledPayload) (events.EnvelopeV1, error) {
	if !validCalendarInviteEnvelopeBase(eventID, occurredAt, organizationID, actorID, payload.EventID, payload.UID, payload.AttendeeIDs) {
		return events.EnvelopeV1{}, ErrEventInvalid
	}
	payload.AttendeeIDs = append([]uuid.UUID(nil), payload.AttendeeIDs...)
	envelope, err := events.NewEnvelopeV1("calendar.invite.cancelled.v1", eventID, occurredAt.UTC(), organizationID, copyUUID(actorID), correlationID, payload)
	if err != nil { return events.EnvelopeV1{}, ErrEventInvalid }
	return envelope, nil
}

// UpdateCalendarEvent resolves the nondisclosing authoritative master before
// evaluating terminal, source, optimistic-version or recurrence constraints.
func (service *EventService) UpdateCalendarEvent(ctx context.Context, actorID uuid.UUID, command CalendarEventUpdateCommand) (CalendarEventUpdateResult, error) {
	if service == nil || service.repository == nil || service.clock == nil || service.ids == nil || actorID == uuid.Nil || !command.valid() {
		return CalendarEventUpdateResult{}, ErrEventInvalid
	}
	var result CalendarEventUpdateResult
	err := service.repository.WithEventTransaction(ctx, func(transaction *EventTransaction) error {
		scope, err := transaction.ResolveCalendarEventUpdateScope(ctx, actorID)
		if err != nil { return err }
		request := CalendarEventUpdateAuthorizationRequest{EventID: command.input.EventID}
		if command.input.Attendees.Present { request.Attendees = cloneCalendarEventUpdateAttendeeCandidates(command.input.Attendees.Value) }
		authorization, err := transaction.ResolveCalendarEventUpdate(ctx, scope, request)
		if err != nil { return err }
		if err := validateCalendarEventUpdateAuthorization(scope, command.input.EventID, command.input, authorization); err != nil { return err }
		if authorization.Current.Status == CalendarEventStatusCancelled || authorization.Current.Source != CalendarEventSourceHub {
			return calendarEventUpdateNotMutable()
		}
		if authorization.Current.Version != command.input.ExpectedVersion {
			return calendarEventUpdateVersionConflict(command.input.EventID, authorization.Current.Version)
		}
		updated, fields, update, err := mergeCalendarEventUpdate(authorization.Current, command.input)
		if err != nil { return err }
		if authorization.HasMaterializedException && calendarEventUpdateChangesRecurrenceIdentity(authorization.Current, updated) {
			return calendarEventUpdateRecurrenceExceptionConflict()
		}
		preAttendeePolicy := ClassifyEventUpdate(update)
		mutations, attendeeChanged, activeAttendeeIDs, err := calendarEventUpdateAttendeeMutations(authorization, command.input, preAttendeePolicy, service.clock, service.ids, scope.ActorID())
		if err != nil { return err }
		for _, mutation := range mutations {
			if mutation.Kind == CalendarEventUpdateAttendeeAdd { update.AddedAttendees = true }
			if mutation.Kind == CalendarEventUpdateAttendeeRemove { update.RemovedAttendees = true }
		}
		if attendeeChanged && !containsCalendarEventChangedField(fields, CalendarEventChangedFieldAttendees) {
			fields = append(fields, CalendarEventChangedFieldAttendees)
			fields = canonicalCalendarEventChangedFields(fields)
		}
		if len(fields) == 0 {
			result, err = NewCalendarEventUpdateResult(authorization.Current.EventID, authorization.Current.Version, authorization.Current.Sequence, authorization.Current.Status, nil)
			return err
		}
		policy := ClassifyEventUpdate(update)
		if !calendarEventUpdateCanIncrementSequence(authorization.Current.Sequence, policy.SequenceIncrement) {
			return ErrEventInvalid
		}
		updated.Version = authorization.Current.Version + policy.VersionIncrement
		updated.Sequence = authorization.Current.Sequence + policy.SequenceIncrement
		event, err := NewCalendarEventCommand(updated)
		if err != nil { return ErrEventInvalid }
		removedAttendeeIDs := calendarEventUpdateRemovedAttendeeIDs(mutations)
		var updatedNotificationID uuid.UUID
		var cancelledNotificationID uuid.UUID
		if policy.SendRequest || policy.CancelIntent == CancelIntentRemovedAttendees {
			updatedNotificationID = service.ids.NewEventID()
			if updatedNotificationID == uuid.Nil { return ErrEventInvalid }
		}
		if policy.CancelIntent == CancelIntentRemovedAttendees {
			cancelledNotificationID = service.ids.NewEventID()
			if cancelledNotificationID == uuid.Nil { return ErrEventInvalid }
		}
		if updatedNotificationID != uuid.Nil && cancelledNotificationID != uuid.Nil && updatedNotificationID == cancelledNotificationID {
			return ErrEventInvalid
		}
		outbox, err := calendarEventUpdateOutbox(updatedNotificationID, cancelledNotificationID, service.clock, scope, updated, activeAttendeeIDs, removedAttendeeIDs, policy, fields)
		if err != nil { return err }
		persistence, err := NewCalendarEventUpdatePersistenceCommand(event, command.input.ExpectedVersion, mutations, outbox)
		if err != nil { return err }
		persisted, err := transaction.PersistCalendarEventUpdate(ctx, scope, persistence)
		if err != nil { return err }
		if !calendarEventUpdatePersistenceMatches(persistence, persisted, updated.Sequence) { return ErrEventInvalid }
		result, err = NewCalendarEventUpdateResult(persisted.EventID(), persisted.Version(), persisted.Sequence(), updated.Status, fields)
		return err
	})
	if err != nil { return CalendarEventUpdateResult{}, calendarEventUpdateSafeError(err) }
	return result, nil
}

// CancelCalendarEvent applies the master-cancel policy. Materialized
// recurrence exceptions intentionally do not block a master cancellation.
func (service *EventService) CancelCalendarEvent(ctx context.Context, actorID, eventID uuid.UUID, expectedVersion uint) (CalendarEventUpdateResult, error) {
	if service == nil || service.repository == nil || service.clock == nil || service.ids == nil || actorID == uuid.Nil || eventID == uuid.Nil || expectedVersion == 0 {
		return CalendarEventUpdateResult{}, ErrEventInvalid
	}
	var result CalendarEventUpdateResult
	err := service.repository.WithEventTransaction(ctx, func(transaction *EventTransaction) error {
		scope, err := transaction.ResolveCalendarEventUpdateScope(ctx, actorID)
		if err != nil { return err }
		authorization, err := transaction.ResolveCalendarEventUpdate(ctx, scope, CalendarEventUpdateAuthorizationRequest{EventID: eventID})
		if err != nil { return err }
		if err := validateCalendarEventUpdateAuthorization(scope, eventID, CalendarEventUpdateInput{EventID: eventID, ExpectedVersion: expectedVersion}, authorization); err != nil { return err }
		if authorization.Current.Status == CalendarEventStatusCancelled || authorization.Current.Source != CalendarEventSourceHub {
			return calendarEventUpdateNotMutable()
		}
		if authorization.Current.Version != expectedVersion {
			return calendarEventUpdateVersionConflict(eventID, authorization.Current.Version)
		}
		if !calendarEventUpdateCanIncrementSequence(authorization.Current.Sequence, 1) {
			return ErrEventInvalid
		}
		updated := cloneCalendarEventCommandInput(authorization.Current)
		updated.Status = CalendarEventStatusCancelled
		updated.Version++
		updated.Sequence++
		event, err := NewCalendarEventCommand(updated)
		if err != nil { return ErrEventInvalid }
		applicationEventID := service.ids.NewEventID()
		if applicationEventID == uuid.Nil { return ErrEventInvalid }
		attendeeIDs := calendarEventUpdateActiveAttendeeIDs(authorization.Attendees, scope.ActorID())
		actor := scope.ActorID()
		outbox := make([]EventOutboxDescriptor, 0, 1)
		if len(attendeeIDs) != 0 {
			if updated.OrganizerID == nil || *updated.OrganizerID == uuid.Nil { return ErrEventInvalid }
			envelope, err := NewCalendarInviteDeliveryEnvelopeV2(
				CalendarInviteCancelledSubjectV2, applicationEventID, service.clock.Now().UTC(), scope.OrganizationID(), &actor, eventID.String(),
				CalendarInviteDeliveryEnvelopeV2Input{
					EventID: eventID, EventSequence: updated.Sequence, OrganizerID: *updated.OrganizerID,
					RecurrenceID: nil, AttendeeIDs: attendeeIDs,
				},
			)
			if err != nil { return err }
			descriptor, err := NewEventOutboxDescriptor(envelope)
			if err != nil { return err }
			outbox = append(outbox, descriptor)
		}
		persistence, err := NewCalendarEventUpdatePersistenceCommand(event, expectedVersion, nil, outbox)
		if err != nil { return err }
		persisted, err := transaction.PersistCalendarEventUpdate(ctx, scope, persistence)
		if err != nil { return err }
		if !calendarEventUpdatePersistenceMatches(persistence, persisted, updated.Sequence) { return ErrEventInvalid }
		result, err = NewCalendarEventUpdateResult(persisted.EventID(), persisted.Version(), persisted.Sequence(), CalendarEventStatusCancelled, nil)
		return err
	})
	if err != nil { return CalendarEventUpdateResult{}, calendarEventUpdateSafeError(err) }
	return result, nil
}

func calendarEventUpdatePresenceValid(input CalendarEventUpdateInput) bool {
	return (input.Title.Present || input.Title.Value == "") &&
		(input.Description.Present || input.Description.Value == "") &&
		(input.Location.Present || input.Location.Value == "") &&
		(input.StartsAt.Present || input.StartsAt.Value == nil) &&
		(input.EndsAt.Present || input.EndsAt.Value == nil) &&
		(input.StartDate.Present || input.StartDate.Value == nil) &&
		(input.EndDateExclusive.Present || input.EndDateExclusive.Value == nil) &&
		(input.AllDay.Present || !input.AllDay.Value) &&
		(input.Timezone.Present || input.Timezone.Value == "") &&
		(input.RRULE.Present || input.RRULE.Value == nil) &&
		(input.Exdates.Present || input.Exdates.Value == nil) &&
		(input.ReminderMinutes.Present || input.ReminderMinutes.Value == nil) &&
		(input.Attendees.Present || input.Attendees.Value == nil)
}

func calendarEventUpdateHasPatch(input CalendarEventUpdateInput) bool {
	return input.Title.Present || input.Description.Present || input.Location.Present || input.StartsAt.Present || input.EndsAt.Present || input.StartDate.Present || input.EndDateExclusive.Present || input.AllDay.Present || input.Timezone.Present || input.RRULE.Present || input.Exdates.Present || input.ReminderMinutes.Present || input.Attendees.Present
}

func normalizeCalendarEventUpdateCandidates(candidates []CalendarEventCreateAttendeeCandidate) ([]CalendarEventCreateAttendeeCandidate, error) {
	if len(candidates) > calendarCreateMaxAttendees { return nil, ErrEventInvalid }
	copyCandidates := make([]CalendarEventCreateAttendeeCandidate, len(candidates))
	seenEmails := make(map[string]struct{}, len(candidates))
	seenUsers := make(map[uuid.UUID]struct{}, len(candidates))
	for index, candidate := range candidates {
		if !validCalendarCreateEmail(candidate.Email) || !validCalendarCreateDisplayName(candidate.DisplayName) || !validCalendarEventAttendeeRole(candidate.Role) {
			return nil, ErrEventInvalid
		}
		candidate.Email = normalizeCalendarCreateEmail(candidate.Email)
		if _, found := seenEmails[candidate.Email]; found { return nil, ErrEventInvalid }
		seenEmails[candidate.Email] = struct{}{}
		if candidate.UserID != nil {
			if *candidate.UserID == uuid.Nil { return nil, ErrEventInvalid }
			if _, found := seenUsers[*candidate.UserID]; found { return nil, ErrEventInvalid }
			seenUsers[*candidate.UserID] = struct{}{}
		}
		copyCandidates[index] = cloneCalendarEventCreateAttendeeCandidate(candidate)
	}
	return copyCandidates, nil
}

func cloneCalendarEventUpdateInput(input CalendarEventUpdateInput) CalendarEventUpdateInput {
	copyInput := input
	copyInput.StartsAt.Value = cloneCalendarEventUpdateTime(input.StartsAt.Value)
	copyInput.EndsAt.Value = cloneCalendarEventUpdateTime(input.EndsAt.Value)
	copyInput.StartDate.Value = cloneCalendarEventCreateString(input.StartDate.Value)
	copyInput.EndDateExclusive.Value = cloneCalendarEventCreateString(input.EndDateExclusive.Value)
	copyInput.RRULE.Value = cloneCalendarEventCreateString(input.RRULE.Value)
	copyInput.ReminderMinutes.Value = copyInt(input.ReminderMinutes.Value)
	copyInput.Exdates.Value = append([]string(nil), input.Exdates.Value...)
	copyInput.Attendees.Value = cloneCalendarEventUpdateAttendeeCandidates(input.Attendees.Value)
	return copyInput
}

func cloneCalendarEventUpdateAttendeeCandidates(candidates []CalendarEventCreateAttendeeCandidate) []CalendarEventCreateAttendeeCandidate {
	if candidates == nil { return nil }
	copied := make([]CalendarEventCreateAttendeeCandidate, len(candidates))
	for index, candidate := range candidates { copied[index] = cloneCalendarEventCreateAttendeeCandidate(candidate) }
	return copied
}

func cloneCalendarEventUpdateTime(value *time.Time) *time.Time {
	if value == nil { return nil }
	copied := *value
	return &copied
}

func cloneCalendarEventUpdateAuthorizationRequest(request CalendarEventUpdateAuthorizationRequest) CalendarEventUpdateAuthorizationRequest {
	return CalendarEventUpdateAuthorizationRequest{EventID: request.EventID, Attendees: cloneCalendarEventUpdateAttendeeCandidates(request.Attendees)}
}

func cloneCalendarEventUpdateAuthorization(authorization CalendarEventUpdateAuthorization) CalendarEventUpdateAuthorization {
	copied := authorization
	copied.Current = cloneCalendarEventCommandInput(authorization.Current)
	copied.Attendees = make([]CalendarEventUpdateAttendeeSnapshot, len(authorization.Attendees))
	for index, attendee := range authorization.Attendees { copied.Attendees[index] = cloneCalendarEventUpdateAttendeeSnapshot(attendee) }
	copied.ResolvedAttendees = make([]CalendarEventCreateResolvedAttendee, len(authorization.ResolvedAttendees))
	for index, attendee := range authorization.ResolvedAttendees {
		copied.ResolvedAttendees[index] = CalendarEventCreateResolvedAttendee{UserID: copyUUID(attendee.UserID), Email: attendee.Email, DisplayName: cloneCalendarEventCreateString(attendee.DisplayName), Role: attendee.Role}
	}
	return copied
}

func cloneCalendarEventCommandInput(input CalendarEventCommandInput) CalendarEventCommandInput {
	copied := input
	copied.Exdates = append([]string(nil), input.Exdates...)
	copied.OrganizerID = copyUUID(input.OrganizerID)
	copied.ScheduledCallID = copyUUID(input.ScheduledCallID)
	copied.ReminderMinutes = copyInt(input.ReminderMinutes)
	return copied
}

func cloneCalendarEventUpdateAttendeeSnapshot(attendee CalendarEventUpdateAttendeeSnapshot) CalendarEventUpdateAttendeeSnapshot {
	attendee.UserID = copyUUID(attendee.UserID)
	attendee.RespondedAt = cloneCalendarEventUpdateTime(attendee.RespondedAt)
	return attendee
}

func cloneCalendarEventUpdateAttendeeMutation(mutation CalendarEventUpdateAttendeeMutation) CalendarEventUpdateAttendeeMutation {
	mutation.UserID = copyUUID(mutation.UserID)
	mutation.RespondedAt = cloneCalendarEventUpdateTime(mutation.RespondedAt)
	mutation.RemovedAt = cloneCalendarEventUpdateTime(mutation.RemovedAt)
	return mutation
}

func cloneCalendarEventUpdatePersistenceCommand(command CalendarEventUpdatePersistenceCommand) CalendarEventUpdatePersistenceCommand {
	cloned, err := NewCalendarEventUpdatePersistenceCommand(command.event, command.expectedVersion, command.attendees, command.outbox)
	if err != nil { return CalendarEventUpdatePersistenceCommand{} }
	return cloned
}

func cloneCalendarEventUpdatePersistenceResult(result CalendarEventUpdatePersistenceResult) CalendarEventUpdatePersistenceResult {
	cloned, err := NewCalendarEventUpdatePersistenceResult(result.eventID, result.version, result.sequence, result.attendees)
	if err != nil { return CalendarEventUpdatePersistenceResult{} }
	return cloned
}

func validateCalendarEventUpdateAuthorization(scope EventMutationScope, eventID uuid.UUID, input CalendarEventUpdateInput, authorization CalendarEventUpdateAuthorization) error {
	current := authorization.Current
	if current.EventID != eventID || current.OrganizationID != scope.OrganizationID() || current.OwnerID != scope.ActorID() ||
		current.OrganizerID == nil || *current.OrganizerID != scope.ActorID() {
		return ErrEventNotFound
	}
	if _, err := NewCalendarEventCommand(current); err != nil || !validCalendarEventUpdateAuthoritativeCurrent(current) { return ErrEventInvalid }
	if len(authorization.Attendees) > calendarCreateMaxAttendees+1 { return ErrEventInvalid }
	seenIDs := make(map[uuid.UUID]struct{}, len(authorization.Attendees))
	seenEmails := make(map[string]struct{}, len(authorization.Attendees))
	seenUsers := make(map[uuid.UUID]struct{}, len(authorization.Attendees))
	organizers := 0
	for _, attendee := range authorization.Attendees {
		displayName := attendee.DisplayName
		if attendee.AttendeeID == uuid.Nil || attendee.Version == 0 || !validCalendarCreateEmail(attendee.Email) || attendee.Email != normalizeCalendarCreateEmail(attendee.Email) ||
			!validCalendarEventStoredAttendeeRoleState(attendee.Role, attendee.Status, "") || !validCalendarCreateDisplayName(&displayName) ||
			(attendee.RespondedAt != nil && attendee.RespondedAt.IsZero()) {
			return ErrEventInvalid
		}
		if attendee.UserID != nil && *attendee.UserID == uuid.Nil { return ErrEventInvalid }
		organizerIdentity := attendee.UserID != nil && *attendee.UserID == current.OwnerID && *attendee.UserID == *current.OrganizerID
		if attendee.Role == CalendarEventAttendeeRoleOrganizer || organizerIdentity {
			if attendee.Email != current.OrganizerEmail ||
				!validCalendarEventStoredOrganizerAttendeeState(attendee.Role, attendee.Status, "", attendee.UserID, current.OwnerID, *current.OrganizerID) {
				return ErrEventInvalid
			}
			organizers++
		}
		if _, found := seenIDs[attendee.AttendeeID]; found { return ErrEventInvalid }
		seenIDs[attendee.AttendeeID] = struct{}{}
		email := normalizeCalendarCreateEmail(attendee.Email)
		if _, found := seenEmails[email]; found { return ErrEventInvalid }
		seenEmails[email] = struct{}{}
		if attendee.UserID != nil {
			if _, found := seenUsers[*attendee.UserID]; found { return ErrEventInvalid }
			seenUsers[*attendee.UserID] = struct{}{}
		}
	}
	if organizers != 1 { return ErrEventInvalid }
	if !input.Attendees.Present {
		if len(authorization.ResolvedAttendees) != 0 { return ErrEventInvalid }
		return nil
	}
	if len(authorization.ResolvedAttendees) != len(input.Attendees.Value) { return ErrEventInvalid }
	requested := make(map[string]CalendarEventCreateAttendeeCandidate, len(input.Attendees.Value))
	for _, candidate := range input.Attendees.Value { requested[candidate.Email] = candidate }
	resolved := make(map[string]CalendarEventCreateResolvedAttendee, len(authorization.ResolvedAttendees))
	resolvedUsers := make(map[uuid.UUID]struct{}, len(authorization.ResolvedAttendees))
	for _, attendee := range authorization.ResolvedAttendees {
		if !validCalendarCreateEmail(attendee.Email) || !validCalendarCreateDisplayName(attendee.DisplayName) || !validCalendarEventAttendeeRole(attendee.Role) || (attendee.UserID != nil && *attendee.UserID == uuid.Nil) || (attendee.UserID != nil && *attendee.UserID == scope.ActorID()) {
			return ErrEventInvalid
		}
		email := normalizeCalendarCreateEmail(attendee.Email)
		if _, found := resolved[email]; found { return ErrEventInvalid }
		if attendee.UserID != nil {
			if _, found := resolvedUsers[*attendee.UserID]; found { return ErrEventInvalid }
			resolvedUsers[*attendee.UserID] = struct{}{}
		}
		resolved[email] = attendee
		candidate, found := requested[email]
		if !found || candidate.Role != attendee.Role || (candidate.UserID != nil && (attendee.UserID == nil || *candidate.UserID != *attendee.UserID)) { return ErrEventInvalid }
	}
	return nil
}

func validCalendarEventUpdateAuthoritativeCurrent(current CalendarEventCommandInput) bool {
	if !validCalendarCreateText(current.Title, calendarCreateMaxTitleBytes, true) ||
		!validCalendarCreateText(current.Description, calendarCreateMaxDescription, false) ||
		!validCalendarCreateText(current.Location, calendarCreateMaxLocationBytes, false) ||
		!utf8.ValidString(current.Temporal.Timezone) || len(current.Temporal.Timezone) > calendarCreateMaxTimezoneBytes ||
		!validCalendarCreateEmail(current.OrganizerEmail) || current.OrganizerEmail != normalizeCalendarCreateEmail(current.OrganizerEmail) ||
		validateITIPUID(current.UID) != nil {
		return false
	}
	if current.Source == CalendarEventSourceHub && current.UID != "hub-"+current.EventID.String() {
		return false
	}
	if current.RRULE != "" {
		normalized, err := normalizeCalendarCreateRRULE(current.RRULE, current.Temporal)
		if err != nil || normalized != current.RRULE { return false }
	}
	canonicalExdates, err := canonicalizeCalendarCreateExdates(current.Temporal, current.Exdates)
	return err == nil && sameCalendarEventUpdateStrings(canonicalExdates, current.Exdates)
}

func calendarEventUpdateCanIncrementSequence(sequence, increment uint) bool {
	return increment <= ^uint(0)-sequence
}

func mergeCalendarEventUpdate(current CalendarEventCommandInput, input CalendarEventUpdateInput) (CalendarEventCommandInput, []CalendarEventChangedField, EventUpdate, error) {
	updated := cloneCalendarEventCommandInput(current)
	fields := make([]CalendarEventChangedField, 0, 13)
	update := EventUpdate{NotifyAttendees: input.NotifyAttendees}
	if input.Title.Present && input.Title.Value != updated.Title { updated.Title = input.Title.Value; fields = append(fields, CalendarEventChangedFieldTitle); update.TitleOrAgendaChanged = true }
	if input.Description.Present && input.Description.Value != updated.Description { updated.Description = input.Description.Value; fields = append(fields, CalendarEventChangedFieldDescription); update.DescriptionChanged = true }
	if input.Location.Present && input.Location.Value != updated.Location { updated.Location = input.Location.Value; fields = append(fields, CalendarEventChangedFieldLocation); update.LocationChanged = true }
	if input.StartsAt.Present {
		value := time.Time{}
		if input.StartsAt.Value != nil { value = *input.StartsAt.Value }
		if !value.Equal(updated.Temporal.StartsAt) { updated.Temporal.StartsAt = value; fields = append(fields, CalendarEventChangedFieldStartsAt); update.TimeChanged = true }
	}
	if input.EndsAt.Present {
		value := time.Time{}
		if input.EndsAt.Value != nil { value = *input.EndsAt.Value }
		if !value.Equal(updated.Temporal.EndsAt) { updated.Temporal.EndsAt = value; fields = append(fields, CalendarEventChangedFieldEndsAt); update.TimeChanged = true }
	}
	if input.StartDate.Present {
		value := ""
		if input.StartDate.Value != nil { value = *input.StartDate.Value }
		if value != updated.Temporal.StartDate { updated.Temporal.StartDate = value; fields = append(fields, CalendarEventChangedFieldStartDate); update.RecurrenceChanged = true }
	}
	if input.EndDateExclusive.Present {
		value := ""
		if input.EndDateExclusive.Value != nil { value = *input.EndDateExclusive.Value }
		if value != updated.Temporal.EndDateExclusive { updated.Temporal.EndDateExclusive = value; fields = append(fields, CalendarEventChangedFieldEndDateExclusive); update.TimeChanged = true }
	}
	if input.AllDay.Present && input.AllDay.Value != updated.Temporal.AllDay { updated.Temporal.AllDay = input.AllDay.Value; fields = append(fields, CalendarEventChangedFieldAllDay); update.AllDayChanged = true }
	if input.Timezone.Present && input.Timezone.Value != updated.Temporal.Timezone { updated.Temporal.Timezone = input.Timezone.Value; fields = append(fields, CalendarEventChangedFieldTimezone); update.TimezoneChanged = true }
	if input.RRULE.Present {
		value := ""
		if input.RRULE.Value != nil { value = *input.RRULE.Value }
		updated.RRULE = value
	}
	if input.Exdates.Present {
		updated.Exdates = append([]string(nil), input.Exdates.Value...)
	}
	if input.ReminderMinutes.Present && !sameCalendarEventUpdateInt(input.ReminderMinutes.Value, updated.ReminderMinutes) { updated.ReminderMinutes = copyInt(input.ReminderMinutes.Value); fields = append(fields, CalendarEventChangedFieldReminderMinutes); update.ReminderChanged = true }
	if updated.Temporal.Validate() != nil { return CalendarEventCommandInput{}, nil, EventUpdate{}, ErrEventInvalid }
	if updated.RRULE != "" {
		normalized, err := normalizeCalendarCreateRRULE(updated.RRULE, updated.Temporal)
		if err != nil { return CalendarEventCommandInput{}, nil, EventUpdate{}, ErrEventInvalid }
		updated.RRULE = normalized
	}
	if input.RRULE.Present && updated.RRULE != current.RRULE {
		fields = append(fields, CalendarEventChangedFieldRRULE)
		update.RecurrenceChanged = true
	}
	exdates, err := canonicalizeCalendarCreateExdates(updated.Temporal, updated.Exdates)
	if err != nil { return CalendarEventCommandInput{}, nil, EventUpdate{}, ErrEventInvalid }
	updated.Exdates = exdates
	if input.Exdates.Present && !sameCalendarEventUpdateStrings(updated.Exdates, current.Exdates) {
		fields = append(fields, CalendarEventChangedFieldExdates)
		update.RecurrenceChanged = true
	}
	return updated, canonicalCalendarEventChangedFields(fields), update, nil
}

func calendarEventUpdateChangesRecurrenceIdentity(current, updated CalendarEventCommandInput) bool {
	return current.RRULE != updated.RRULE ||
		!sameCalendarEventUpdateStrings(current.Exdates, updated.Exdates) ||
		!current.Temporal.StartsAt.Equal(updated.Temporal.StartsAt) ||
		current.Temporal.StartDate != updated.Temporal.StartDate ||
		current.Temporal.Timezone != updated.Temporal.Timezone ||
		current.Temporal.AllDay != updated.Temporal.AllDay
}

func calendarEventUpdateAttendeeMutations(authorization CalendarEventUpdateAuthorization, input CalendarEventUpdateInput, policy EventUpdatePolicy, clock EventClock, ids EventIDGenerator, organizerID uuid.UUID) ([]CalendarEventUpdateAttendeeMutation, bool, []uuid.UUID, error) {
	current := make(map[string]CalendarEventUpdateAttendeeSnapshot, len(authorization.Attendees))
	for _, attendee := range authorization.Attendees {
		if attendee.UserID != nil && *attendee.UserID == organizerID { continue }
		current[normalizeCalendarCreateEmail(attendee.Email)] = attendee
	}
	mutations := make([]CalendarEventUpdateAttendeeMutation, 0, len(current)+len(authorization.ResolvedAttendees))
	changed := false
	if input.Attendees.Present {
		desired := make(map[string]CalendarEventCreateResolvedAttendee, len(authorization.ResolvedAttendees))
		for _, attendee := range authorization.ResolvedAttendees { desired[normalizeCalendarCreateEmail(attendee.Email)] = attendee }
		for email, attendee := range current {
			resolved, found := desired[email]
			if !found {
				removedAt := clock.Now().UTC()
				mutations = append(mutations, CalendarEventUpdateAttendeeMutation{Kind: CalendarEventUpdateAttendeeRemove, AttendeeID: attendee.AttendeeID, UserID: copyUUID(attendee.UserID), Email: normalizeCalendarCreateEmail(attendee.Email), DisplayName: attendee.DisplayName, Role: attendee.Role, Status: attendee.Status, RespondedAt: cloneCalendarEventUpdateTime(attendee.RespondedAt), Version: attendee.Version + 1, RemovedAt: &removedAt})
				changed = true
				continue
			}
			if resolved.Role != attendee.Role || !sameCalendarEventUpdateUUID(resolved.UserID, attendee.UserID) { return nil, false, nil, ErrEventConflict }
			if calendarEventUpdateResetsNonOrganizers(policy) {
				mutations = append(mutations, calendarEventUpdateResetMutation(attendee))
			}
		}
		for email, resolved := range desired {
			if _, found := current[email]; found { continue }
			attendeeID := ids.NewEventID()
			if attendeeID == uuid.Nil { return nil, false, nil, ErrEventInvalid }
			mutations = append(mutations, CalendarEventUpdateAttendeeMutation{Kind: CalendarEventUpdateAttendeeAdd, AttendeeID: attendeeID, UserID: copyUUID(resolved.UserID), Email: email, DisplayName: calendarCreateDisplayName(resolved.DisplayName), Role: resolved.Role, Status: RSVPNeedsAction, Version: 1})
			changed = true
		}
	} else if calendarEventUpdateResetsNonOrganizers(policy) {
		for _, attendee := range current { mutations = append(mutations, calendarEventUpdateResetMutation(attendee)) }
	}
	sortCalendarEventUpdateMutations(mutations)
	activeIDs := calendarEventUpdateActiveAttendeeIDs(authorization.Attendees, organizerID)
	for _, mutation := range mutations {
		switch mutation.Kind {
		case CalendarEventUpdateAttendeeAdd:
			activeIDs = append(activeIDs, mutation.AttendeeID)
		case CalendarEventUpdateAttendeeRemove:
			activeIDs = removeCalendarEventUpdateAttendeeID(activeIDs, mutation.AttendeeID)
		}
	}
	sortCalendarEventUpdateUUIDs(activeIDs)
	return mutations, changed, activeIDs, nil
}

func calendarEventUpdateResetMutation(attendee CalendarEventUpdateAttendeeSnapshot) CalendarEventUpdateAttendeeMutation {
	return CalendarEventUpdateAttendeeMutation{Kind: CalendarEventUpdateAttendeeReset, AttendeeID: attendee.AttendeeID, UserID: copyUUID(attendee.UserID), Email: normalizeCalendarCreateEmail(attendee.Email), DisplayName: attendee.DisplayName, Role: attendee.Role, Status: RSVPNeedsAction, Version: attendee.Version + 1}
}

func calendarEventUpdateActiveAttendeeIDs(attendees []CalendarEventUpdateAttendeeSnapshot, organizerID uuid.UUID) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(attendees))
	for _, attendee := range attendees {
		if attendee.UserID != nil && *attendee.UserID == organizerID { continue }
		ids = append(ids, attendee.AttendeeID)
	}
	sortCalendarEventUpdateUUIDs(ids)
	return ids
}

func calendarEventUpdateOutbox(updatedEventID uuid.UUID, cancelledEventID uuid.UUID, clock EventClock, scope EventMutationScope, updated CalendarEventCommandInput, activeAttendeeIDs []uuid.UUID, removedAttendeeIDs []uuid.UUID, policy EventUpdatePolicy, fields []CalendarEventChangedField) ([]EventOutboxDescriptor, error) {
	if updatedEventID == uuid.Nil && cancelledEventID == uuid.Nil { return nil, nil }
	actor := scope.ActorID()
	descriptors := make([]EventOutboxDescriptor, 0, 2)
	if updated.OrganizerID == nil || *updated.OrganizerID == uuid.Nil { return nil, ErrEventInvalid }
	if updatedEventID != uuid.Nil && len(activeAttendeeIDs) != 0 {
		envelope, err := NewCalendarInviteDeliveryEnvelopeV2(
			CalendarInviteUpdatedSubjectV2, updatedEventID, clock.Now().UTC(), scope.OrganizationID(), &actor, updated.EventID.String(),
			CalendarInviteDeliveryEnvelopeV2Input{
				EventID: updated.EventID, EventSequence: updated.Sequence, OrganizerID: *updated.OrganizerID,
				RecurrenceID: nil, AttendeeIDs: activeAttendeeIDs,
			},
		)
		if err != nil { return nil, err }
		descriptor, err := NewEventOutboxDescriptor(envelope)
		if err != nil { return nil, err }
		descriptors = append(descriptors, descriptor)
	}
	if cancelledEventID != uuid.Nil && len(removedAttendeeIDs) != 0 {
		envelope, err := NewCalendarInviteDeliveryEnvelopeV2(
			CalendarInviteCancelledSubjectV2, cancelledEventID, clock.Now().UTC(), scope.OrganizationID(), &actor, updated.EventID.String(),
			CalendarInviteDeliveryEnvelopeV2Input{
				EventID: updated.EventID, EventSequence: updated.Sequence, OrganizerID: *updated.OrganizerID,
				RecurrenceID: nil, AttendeeIDs: removedAttendeeIDs,
			},
		)
		if err != nil { return nil, err }
		descriptor, err := NewEventOutboxDescriptor(envelope)
		if err != nil { return nil, err }
		descriptors = append(descriptors, descriptor)
	}
	return descriptors, nil
}

func calendarEventUpdateRemovedAttendeeIDs(mutations []CalendarEventUpdateAttendeeMutation) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(mutations))
	for _, mutation := range mutations {
		if mutation.Kind == CalendarEventUpdateAttendeeRemove { ids = append(ids, mutation.AttendeeID) }
	}
	sortCalendarEventUpdateUUIDs(ids)
	return ids
}

func calendarEventUpdateResetsNonOrganizers(policy EventUpdatePolicy) bool {
	return policy.RSVPReset == RSVPResetAllNonOrganizers || policy.RSVPReset == RSVPResetAffectedNonOrganizers
}

func calendarEventUpdatePersistenceMatches(command CalendarEventUpdatePersistenceCommand, result CalendarEventUpdatePersistenceResult, expectedSequence uint) bool {
	if result.EventID() != command.Event().EventID() || result.Version() != command.Event().Version() || result.Sequence() != expectedSequence { return false }
	mutations, results := command.Attendees(), result.Attendees()
	if len(mutations) != len(results) { return false }
	for index, mutation := range mutations {
		result := results[index]
		if result.AttendeeID != mutation.AttendeeID || result.Version != mutation.Version || result.Removed != (mutation.Kind == CalendarEventUpdateAttendeeRemove) { return false }
	}
	return true
}

func calendarEventUpdateSafeError(err error) error {
	var conflict CalendarEventUpdateConflictError
	if errors.As(err, &conflict) { return conflict }
	switch {
	case errors.Is(err, ErrEventNotFound): return ErrEventNotFound
	case errors.Is(err, ErrEventForbidden): return ErrEventForbidden
	case errors.Is(err, ErrEventTransactionClosed): return ErrEventTransactionClosed
	case errors.Is(err, ErrEventConflict): return ErrEventConflict
	default: return ErrEventInvalid
	}
}

func validCalendarEventUpdateAttendeeMutation(mutation CalendarEventUpdateAttendeeMutation) bool {
	displayName := mutation.DisplayName
	if mutation.AttendeeID == uuid.Nil || mutation.Version == 0 || !validCalendarCreateEmail(mutation.Email) || !validCalendarEventAttendeeRole(mutation.Role) || mutation.Status.Validate() != nil || !validCalendarCreateDisplayName(&displayName) || (mutation.UserID != nil && *mutation.UserID == uuid.Nil) {
		return false
	}
	switch mutation.Kind {
	case CalendarEventUpdateAttendeeAdd:
		return mutation.Version == 1 && mutation.Status == RSVPNeedsAction && mutation.RespondedAt == nil && mutation.RemovedAt == nil
	case CalendarEventUpdateAttendeeReset:
		return mutation.Status == RSVPNeedsAction && mutation.RespondedAt == nil && mutation.RemovedAt == nil
	case CalendarEventUpdateAttendeeRemove:
		return mutation.RemovedAt != nil && !mutation.RemovedAt.IsZero() && mutation.RemovedAt.Location() == time.UTC
	default:
		return false
	}
}

func sortCalendarEventUpdateMutations(mutations []CalendarEventUpdateAttendeeMutation) {
	sort.Slice(mutations, func(left, right int) bool { return bytes.Compare(mutations[left].AttendeeID[:], mutations[right].AttendeeID[:]) < 0 })
}
func duplicateCalendarEventUpdateMutationIDs(mutations []CalendarEventUpdateAttendeeMutation) bool {
	for index := 1; index < len(mutations); index++ { if mutations[index-1].AttendeeID == mutations[index].AttendeeID { return true } }
	return false
}
func sortCalendarEventUpdateUUIDs(ids []uuid.UUID) {
	sort.Slice(ids, func(left, right int) bool { return bytes.Compare(ids[left][:], ids[right][:]) < 0 })
}
func removeCalendarEventUpdateAttendeeID(ids []uuid.UUID, target uuid.UUID) []uuid.UUID {
	result := ids[:0]
	for _, id := range ids { if id != target { result = append(result, id) } }
	return result
}
func sameCalendarEventUpdateStrings(left, right []string) bool {
	if len(left) != len(right) { return false }
	for index := range left { if left[index] != right[index] { return false } }
	return true
}
func sameCalendarEventUpdateInt(left, right *int) bool { return (left == nil && right == nil) || (left != nil && right != nil && *left == *right) }
func sameCalendarEventUpdateUUID(left, right *uuid.UUID) bool { return (left == nil && right == nil) || (left != nil && right != nil && *left == *right) }
func containsCalendarEventChangedField(fields []CalendarEventChangedField, wanted CalendarEventChangedField) bool { for _, field := range fields { if field == wanted { return true } }; return false }

func canonicalCalendarEventChangedFields(fields []CalendarEventChangedField) []CalendarEventChangedField {
	ordered := make([]CalendarEventChangedField, 0, len(fields))
	for _, candidate := range calendarEventChangedFieldOrder() { if containsCalendarEventChangedField(fields, candidate) { ordered = append(ordered, candidate) } }
	return ordered
}
func calendarEventChangedFieldOrder() []CalendarEventChangedField {
	return []CalendarEventChangedField{CalendarEventChangedFieldTitle, CalendarEventChangedFieldDescription, CalendarEventChangedFieldLocation, CalendarEventChangedFieldStartsAt, CalendarEventChangedFieldEndsAt, CalendarEventChangedFieldStartDate, CalendarEventChangedFieldEndDateExclusive, CalendarEventChangedFieldAllDay, CalendarEventChangedFieldTimezone, CalendarEventChangedFieldRRULE, CalendarEventChangedFieldExdates, CalendarEventChangedFieldReminderMinutes, CalendarEventChangedFieldAttendees}
}
func validCalendarEventChangedFields(fields []CalendarEventChangedField) bool {
	seen := make(map[CalendarEventChangedField]struct{}, len(fields))
	order := calendarEventChangedFieldOrder()
	last := -1
	for _, field := range fields {
		index := -1
		for position, allowed := range order { if field == allowed { index = position; break } }
		if index < 0 || index <= last { return false }
		if _, found := seen[field]; found { return false }
		seen[field] = struct{}{}
		last = index
	}
	return true
}
func validCalendarEventChangedFieldsNonEmpty(fields []CalendarEventChangedField) bool { return len(fields) > 0 && len(fields) <= 13 && validCalendarEventChangedFields(fields) }

func validCalendarInviteEnvelopeBase(eventID uuid.UUID, occurredAt time.Time, organizationID uuid.UUID, actorID *uuid.UUID, payloadEventID uuid.UUID, uid string, attendees []uuid.UUID) bool {
	if eventID == uuid.Nil || payloadEventID == uuid.Nil || organizationID == uuid.Nil || occurredAt.IsZero() || !utf8.ValidString(uid) || len(uid) == 0 || len(uid) > 255 { return false }
	if actorID != nil && *actorID == uuid.Nil { return false }
	if len(attendees) > calendarCreateMaxAttendees { return false }
	for index, attendeeID := range attendees {
		if attendeeID == uuid.Nil || (index > 0 && bytes.Compare(attendees[index-1][:], attendeeID[:]) >= 0) { return false }
	}
	return true
}
