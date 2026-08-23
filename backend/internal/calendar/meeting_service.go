package calendar

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/platform/events"
)

// MeetingService composes the meeting vertical without introducing a second
// repository or any pre-commit provider/network I/O.
type MeetingService struct {
	repository EventRepository
	clock      EventClock
	ids        EventIDGenerator
}

func NewMeetingService(repository EventRepository, clock EventClock, ids EventIDGenerator) *MeetingService {
	return &MeetingService{repository: repository, clock: clock, ids: ids}
}

func (service *MeetingService) ready() bool {
	return service != nil && service.repository != nil && service.clock != nil && service.ids != nil
}

// CreateMeeting atomically creates the canonical event, attendees,
// reservations, linked ScheduledCall, outbox intents and replay receipt.
func (service *MeetingService) CreateMeeting(ctx context.Context, actorID uuid.UUID, command MeetingCreateCommand) (MeetingCreateResult, error) {
	if !service.ready() || actorID == uuid.Nil || !command.valid() || !command.allowsActor(actorID) {
		return MeetingCreateResult{}, ErrEventInvalid
	}
	var result MeetingCreateResult
	err := service.repository.WithEventTransaction(ctx, func(transaction *EventTransaction) error {
		scope, err := transaction.ResolveMeetingCreateScope(ctx, actorID)
		if err != nil {
			return err
		}
		reservation, err := NewMutationReceiptReservation(scope, command.keyHash, command.requestHash)
		if err != nil {
			return err
		}
		reserved, err := transaction.ReserveMutationReceipt(ctx, scope, reservation)
		if err != nil {
			return err
		}
		switch reserved.State() {
		case MutationReceiptReplayed:
			response, found := reserved.ReplayedResponse()
			if !found || response.Status() != 201 {
				return ErrEventInternal
			}
			result = MeetingCreateResult{response: response, replayed: true}
			return nil
		case MutationReceiptReserved:
		default:
			return ErrEventInternal
		}

		authorization, err := transaction.ResolveMeetingCreateAuthorization(ctx, scope, MeetingCreateAuthorizationRequest{
			ProjectID:       command.input.ProjectID,
			ChannelID:       command.input.ChannelID,
			AttendeeUserIDs: append([]uuid.UUID(nil), command.input.AttendeeUserIDs...),
		})
		if err != nil {
			return err
		}
		if !meetingCreateAuthorizationMatches(scope, command.input, authorization) {
			return ErrEventInternal
		}

		eventID := meetingCreateEventID(scope, command)
		callID := meetingCreateCallID(eventID)
		event, attendees, attendeeRecordIDs, err := meetingCreateEventCommands(scope, eventID, command.input, authorization)
		if err != nil {
			return err
		}
		persistedEvent, err := transaction.PersistCalendarEvent(ctx, scope, event)
		if err != nil {
			return err
		}
		if persistedEvent.EventID() != eventID || persistedEvent.Version() != 1 {
			return ErrEventInternal
		}
		for _, attendee := range attendees {
			if err := calendarCreatePersistAttendee(ctx, transaction, scope, attendee); err != nil {
				return err
			}
		}

		occurrences, err := meetingCreateReservationOccurrences(command.input, eventID)
		if err != nil {
			return err
		}
		reservationSet, err := NewCalendarAvailabilityReservationSet(
			scope.OrganizationID(),
			eventID,
			scope.ActorID(),
			command.input.AttendeeUserIDs,
			occurrences,
		)
		if err != nil {
			return ErrEventInternal
		}
		if err := transaction.ReconcileCalendarAvailabilityReservations(ctx, scope, reservationSet); err != nil {
			return err
		}

		callCommand, err := NewMeetingScheduledCallCommand(
			callID,
			eventID,
			scope.OrganizationID(),
			command.input.ProjectID,
			command.input.ChannelID,
			scope.ActorID(),
			command.input,
		)
		if err != nil {
			return ErrEventInternal
		}
		persistedCall, err := transaction.PersistMeetingScheduledCall(ctx, scope, callCommand)
		if err != nil {
			return err
		}
		if persistedCall.CallID() != callID || persistedCall.Version() == 0 {
			return ErrEventInternal
		}
		if err := service.appendMeetingCreateOutbox(ctx, transaction, scope, command.input, eventID, callID, attendeeRecordIDs); err != nil {
			return err
		}

		response, err := meetingCreateResponse(eventID, callID, scope.ActorID(), command.input, persistedEvent.Version(), persistedCall.Version())
		if err != nil {
			return ErrEventInternal
		}
		completion, err := NewMutationReceiptCompletion(reserved.ReceiptID(), &eventID, response)
		if err != nil {
			return ErrEventInternal
		}
		if err := transaction.CompleteMutationReceipt(ctx, scope, completion); err != nil {
			return err
		}
		result = MeetingCreateResult{response: response}
		return nil
	})
	if err != nil {
		return MeetingCreateResult{}, meetingCreateSafeError(err)
	}
	if !result.valid() {
		return MeetingCreateResult{}, ErrEventInternal
	}
	return result, nil
}

func meetingCreateAuthorizationMatches(scope EventMutationScope, input MeetingCreateInput, authorization MeetingCreateAuthorization) bool {
	if !authorization.valid() || authorization.ProjectID != input.ProjectID || authorization.ChannelID != input.ChannelID ||
		authorization.Organizer.UserID != scope.ActorID() || len(authorization.Attendees) != len(input.AttendeeUserIDs) {
		return false
	}
	for index, attendee := range authorization.Attendees {
		if attendee.UserID != input.AttendeeUserIDs[index] {
			return false
		}
	}
	return true
}

func meetingCreateEventID(scope EventMutationScope, command MeetingCreateCommand) uuid.UUID {
	actorID, organizationID := scope.ActorID(), scope.OrganizationID()
	return calendarCreateUUID("hub.calendar.meeting.event.id.v1", actorID[:], organizationID[:], command.keyHash[:], command.requestHash[:])
}

func meetingCreateCallID(eventID uuid.UUID) uuid.UUID {
	return calendarCreateUUID("hub.calendar.meeting.call.id.v1", eventID[:])
}

func meetingCreateEventCommands(scope EventMutationScope, eventID uuid.UUID, input MeetingCreateInput, authorization MeetingCreateAuthorization) (CalendarEventCommand, []CalendarEventAttendeeCommand, []uuid.UUID, error) {
	temporal := CalendarEvent{StartsAt: input.StartsAt, EndsAt: input.EndsAt, Timezone: input.Timezone}
	organizerID := scope.ActorID()
	event, err := NewCalendarEventCommand(CalendarEventCommandInput{
		EventID:       eventID,
		OrganizationID: scope.OrganizationID(),
		OwnerID:       organizerID,
		UID:           "hub-" + eventID.String(),
		Title:         input.Title,
		Description:   input.Description,
		Temporal:      temporal,
		RRULE:         input.RRULE,
		Sequence:      0,
		Version:       1,
		Source:        CalendarEventSourceHub,
		Status:        CalendarEventStatusConfirmed,
		OrganizerID:   &organizerID,
		OrganizerEmail: normalizeCalendarCreateEmail(authorization.Organizer.Email),
	})
	if err != nil {
		return CalendarEventCommand{}, nil, nil, ErrEventInternal
	}
	organizerRecordID := calendarCreateUUID("hub.calendar.attendee.id.v1", eventID[:], []byte("organizer"))
	organizer, err := newCalendarEventOrganizerAttendeeCommand(CalendarEventAttendeeCommandInput{
		AttendeeID:  organizerRecordID,
		EventID:     eventID,
		UserID:      &organizerID,
		Email:       normalizeCalendarCreateEmail(authorization.Organizer.Email),
		DisplayName: calendarCreateDisplayName(authorization.Organizer.DisplayName),
		Role:        CalendarEventAttendeeRoleOrganizer,
		Status:      RSVPAccepted,
		Version:     1,
	})
	if err != nil {
		return CalendarEventCommand{}, nil, nil, ErrEventInternal
	}
	attendees := make([]CalendarEventAttendeeCommand, 0, len(authorization.Attendees)+1)
	attendees = append(attendees, organizer)
	attendeeRecordIDs := make([]uuid.UUID, len(authorization.Attendees))
	for index, resolved := range authorization.Attendees {
		attendeeID := resolved.UserID
		recordID := calendarCreateUUID("hub.calendar.meeting.attendee.id.v1", eventID[:], attendeeID[:])
		attendee, err := NewCalendarEventAttendeeCommand(CalendarEventAttendeeCommandInput{
			AttendeeID: recordID,
			EventID:    eventID,
			UserID:     &attendeeID,
			Email:      normalizeCalendarCreateEmail(resolved.Email),
			DisplayName: calendarCreateDisplayName(resolved.DisplayName),
			Role:       CalendarEventAttendeeRoleRequired,
			Status:     RSVPNeedsAction,
			Version:    1,
		})
		if err != nil {
			return CalendarEventCommand{}, nil, nil, ErrEventInternal
		}
		attendees = append(attendees, attendee)
		attendeeRecordIDs[index] = recordID
	}
	return event, attendees, attendeeRecordIDs, nil
}

func meetingCreateReservationOccurrences(input MeetingCreateInput, eventID uuid.UUID) ([]CalendarAvailabilityReservationOccurrence, error) {
	temporal := CalendarEvent{StartsAt: input.StartsAt, EndsAt: input.EndsAt, Timezone: input.Timezone}
	materialized, _, err := availabilityTimedEventOccurrences(temporal, input.RRULE, nil, nil, eventID)
	if err != nil || len(materialized) == 0 || len(materialized) > calendarAvailabilityReservationMaxOccurrences {
		return nil, ErrEventInternal
	}
	occurrences := make([]CalendarAvailabilityReservationOccurrence, len(materialized))
	for index, occurrence := range materialized {
		occurrences[index] = CalendarAvailabilityReservationOccurrence{
			OccurrenceKey: occurrence.key,
			StartsAt:      occurrence.from,
			EndsAt:        occurrence.to,
		}
	}
	return occurrences, nil
}

func (service *MeetingService) appendMeetingCreateOutbox(ctx context.Context, transaction *EventTransaction, scope EventMutationScope, input MeetingCreateInput, eventID, callID uuid.UUID, attendeeRecordIDs []uuid.UUID) error {
	occurredAt, ok := canonicalAvailabilityTime(service.clock.Now().UTC().Truncate(time.Second))
	if !ok {
		return ErrEventInternal
	}
	envelopeIDs := make([]uuid.UUID, 3)
	seen := make(map[uuid.UUID]struct{}, len(envelopeIDs))
	for index := range envelopeIDs {
		envelopeIDs[index] = service.ids.NewEventID()
		if envelopeIDs[index] == uuid.Nil {
			return ErrEventInternal
		}
		if _, duplicate := seen[envelopeIDs[index]]; duplicate {
			return ErrEventInternal
		}
		seen[envelopeIDs[index]] = struct{}{}
	}
	actorID := scope.ActorID()
	correlationID := eventID.String()
	payloads := []struct {
		subject string
		id      uuid.UUID
		payload any
	}{
		{
			subject: "calendar.meeting.created.v1",
			id:      envelopeIDs[0],
			payload: meetingCalendarCreatedPayload{
				EventID: eventID, ScheduledCallID: callID, ProjectID: input.ProjectID, ChannelID: input.ChannelID,
				OrganizerID: actorID, AttendeeUserIDs: append([]uuid.UUID(nil), input.AttendeeUserIDs...),
			},
		},
		{
			subject: "calendar.invite.created.v1",
			id:      envelopeIDs[1],
			payload: calendarCreateInvitePayload{EventID: eventID, UID: "hub-" + eventID.String(), Sequence: 0, AttendeeIDs: append([]uuid.UUID(nil), attendeeRecordIDs...)},
		},
		{
			subject: "calls.channel." + input.ChannelID.String() + ".scheduled_created.v1",
			id:      envelopeIDs[2],
			payload: meetingCallScheduledPayload{
				ScheduledCallID: callID, CalendarEventID: eventID, ProjectID: input.ProjectID, ChannelID: input.ChannelID,
				CreatorID: actorID, ParticipantIDs: append([]uuid.UUID(nil), input.AttendeeUserIDs...), Title: input.Title,
				ScheduledAt: input.StartsAt, DurationMinutes: int(input.EndsAt.Sub(input.StartsAt) / time.Minute), RRULE: input.RRULE,
			},
		},
	}
	for _, item := range payloads {
		envelope, err := events.NewEnvelopeV1(item.subject, item.id, occurredAt, scope.OrganizationID(), &actorID, correlationID, item.payload)
		if err != nil {
			return ErrEventInternal
		}
		descriptor, err := NewEventOutboxDescriptor(envelope)
		if err != nil {
			return ErrEventInternal
		}
		if err := transaction.AppendOutbox(ctx, scope, descriptor); err != nil {
			return err
		}
	}
	return nil
}

func meetingCreateSafeError(err error) error {
	var meetingConflict MeetingCreateConflictError
	if errors.As(err, &meetingConflict) {
		return newMeetingCreateConflict(meetingConflict.Kind())
	}
	var createConflict *CalendarEventCreateConflictError
	if errors.As(err, &createConflict) && createConflict != nil && createConflict.Kind() == CalendarEventCreateConflictIdempotencyKeyReused {
		return newMeetingCreateConflict(MeetingCreateConflictIdempotencyKeyReused)
	}
	if errors.Is(err, ErrAvailabilitySlotUnavailable) {
		return newMeetingCreateConflict(MeetingCreateConflictSlotUnavailable)
	}
	switch {
	case errors.Is(err, ErrEventNotFound):
		return ErrEventNotFound
	case errors.Is(err, ErrEventForbidden):
		return ErrEventForbidden
	default:
		return ErrEventInternal
	}
}
