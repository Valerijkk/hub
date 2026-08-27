package calendar

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/platform/events"
)

// UpdateMeeting atomically replaces one live linked meeting, including its
// canonical attendees, finite busy reservations and ScheduledCall projection.
func (service *MeetingService) UpdateMeeting(ctx context.Context, actorID uuid.UUID, command MeetingUpdateCommand) (MeetingMutationResult, error) {
	if !service.ready() || actorID == uuid.Nil || !command.valid() || !command.allowsActor(actorID) {
		return MeetingMutationResult{}, ErrEventInvalid
	}
	var result MeetingMutationResult
	err := service.repository.WithEventTransaction(ctx, func(transaction *EventTransaction) error {
		scope, err := transaction.ResolveMeetingUpdateScope(ctx, actorID)
		if err != nil {
			return err
		}
		reserved, err := meetingMutationReserveReceipt(ctx, transaction, scope, command.keyHash, command.requestHash)
		if err != nil {
			return err
		}
		if reserved.State() == MutationReceiptReplayed {
			response, found := reserved.ReplayedResponse()
			if !found || response.Status() != 200 {
				return ErrEventInternal
			}
			result = MeetingMutationResult{response: response, replayed: true}
			return nil
		}
		if reserved.State() != MutationReceiptReserved {
			return ErrEventInternal
		}
		authorization, err := transaction.ResolveMeetingUpdateAuthorization(ctx, scope, MeetingUpdateAuthorizationRequest{
			EventID: command.input.EventID,
			Desired: MeetingCreateAuthorizationRequest{
				ProjectID: command.input.Desired.ProjectID, ChannelID: command.input.Desired.ChannelID,
				AttendeeUserIDs: append([]uuid.UUID(nil), command.input.Desired.AttendeeUserIDs...),
			},
		})
		if err != nil {
			return err
		}
		if !meetingCreateAuthorizationMatches(scope, command.input.Desired, authorization.Desired) ||
			authorization.Current.EventID != command.input.EventID {
			return ErrEventInternal
		}
		if authorization.Current.EventVersion != command.input.ExpectedVersion {
			return meetingMutationVersionConflict(authorization.Current.EventID, authorization.Current.EventVersion)
		}
		if command.input.Desired.ProjectID != authorization.Current.ProjectID || command.input.Desired.ChannelID != authorization.Current.ChannelID {
			return newMeetingMutationConflict(MeetingMutationConflictScopeImmutable)
		}
		plan, err := newMeetingUpdatePlan(authorization.Current, command.input.Desired, command.input.NotifyAttendees)
		if err != nil {
			return ErrEventInternal
		}
		if authorization.Current.HasMaterializedExceptions && plan.materializedSeriesConflict {
			return newMeetingMutationConflict(MeetingMutationConflictRecurrence)
		}
		if !plan.hasChanges() {
			response, err := meetingUpdateResponse(
				authorization.Current.EventID, authorization.Current.ScheduledCallID, scope.ActorID(), command.input.Desired,
				authorization.Current.EventVersion, authorization.Current.ScheduledCallVersion,
			)
			if err != nil {
				return ErrEventInternal
			}
			completion, err := NewMutationReceiptCompletion(reserved.ReceiptID(), &authorization.Current.EventID, response)
			if err != nil {
				return ErrEventInternal
			}
			if err := transaction.CompleteMutationReceipt(ctx, scope, completion); err != nil {
				return err
			}
			result = MeetingMutationResult{response: response}
			return nil
		}
		occurrences, err := meetingCreateReservationOccurrences(command.input.Desired, authorization.Current.EventID)
		if err != nil {
			return err
		}
		reservations, err := NewCalendarAvailabilityReservationSet(
			scope.OrganizationID(), authorization.Current.EventID, scope.ActorID(),
			command.input.Desired.AttendeeUserIDs, occurrences,
		)
		if err != nil {
			return ErrEventInternal
		}
		persistence, err := NewMeetingUpdatePersistenceCommand(
			authorization.Current, command.input.Desired, authorization.Desired, reservations, command.input.ExpectedVersion,
			command.input.NotifyAttendees,
		)
		if err != nil {
			return ErrEventInternal
		}
		persisted, err := transaction.PersistMeetingUpdate(ctx, scope, persistence)
		if err != nil {
			return err
		}
		if err := service.appendMeetingUpdateOutbox(ctx, transaction, scope, command.input.Desired, authorization.Current, persisted, plan); err != nil {
			return err
		}
		response, err := meetingUpdateResponse(
			authorization.Current.EventID, authorization.Current.ScheduledCallID, scope.ActorID(), command.input.Desired,
			persisted.EventVersion(), persisted.CallVersion(),
		)
		if err != nil {
			return ErrEventInternal
		}
		completion, err := NewMutationReceiptCompletion(reserved.ReceiptID(), &authorization.Current.EventID, response)
		if err != nil {
			return ErrEventInternal
		}
		if err := transaction.CompleteMutationReceipt(ctx, scope, completion); err != nil {
			return err
		}
		result = MeetingMutationResult{response: response}
		return nil
	})
	if err != nil {
		return MeetingMutationResult{}, meetingMutationSafeError(err)
	}
	if !result.valid() {
		return MeetingMutationResult{}, ErrEventInternal
	}
	return result, nil
}

// CancelMeeting atomically cancels the canonical event and linked call and
// releases every active reservation before durable outbox intents are stored.
func (service *MeetingService) CancelMeeting(ctx context.Context, actorID uuid.UUID, command MeetingCancelCommand) (MeetingMutationResult, error) {
	if !service.ready() || actorID == uuid.Nil || !command.valid() {
		return MeetingMutationResult{}, ErrEventInvalid
	}
	var result MeetingMutationResult
	err := service.repository.WithEventTransaction(ctx, func(transaction *EventTransaction) error {
		scope, err := transaction.ResolveMeetingCancelScope(ctx, actorID)
		if err != nil {
			return err
		}
		reserved, err := meetingMutationReserveReceipt(ctx, transaction, scope, command.keyHash, command.requestHash)
		if err != nil {
			return err
		}
		if reserved.State() == MutationReceiptReplayed {
			response, found := reserved.ReplayedResponse()
			if !found || response.Status() != 200 {
				return ErrEventInternal
			}
			result = MeetingMutationResult{response: response, replayed: true}
			return nil
		}
		if reserved.State() != MutationReceiptReserved {
			return ErrEventInternal
		}
		authorization, err := transaction.ResolveMeetingCancelAuthorization(ctx, scope, command.input.EventID)
		if err != nil {
			return err
		}
		if authorization.Current.EventVersion != command.input.ExpectedVersion {
			return meetingMutationVersionConflict(authorization.Current.EventID, authorization.Current.EventVersion)
		}
		persistence, err := NewMeetingCancelPersistenceCommand(authorization.Current, command.input.ExpectedVersion)
		if err != nil {
			return ErrEventInternal
		}
		persisted, err := transaction.PersistMeetingCancel(ctx, scope, persistence)
		if err != nil {
			return err
		}
		if err := service.appendMeetingCancelOutbox(ctx, transaction, scope, authorization.Current, persisted); err != nil {
			return err
		}
		response, err := meetingCancelResponse(
			authorization.Current.EventID, authorization.Current.ScheduledCallID,
			persisted.EventVersion(), persisted.CallVersion(),
		)
		if err != nil {
			return ErrEventInternal
		}
		completion, err := NewMutationReceiptCompletion(reserved.ReceiptID(), &authorization.Current.EventID, response)
		if err != nil {
			return ErrEventInternal
		}
		if err := transaction.CompleteMutationReceipt(ctx, scope, completion); err != nil {
			return err
		}
		result = MeetingMutationResult{response: response}
		return nil
	})
	if err != nil {
		return MeetingMutationResult{}, meetingMutationSafeError(err)
	}
	if !result.valid() {
		return MeetingMutationResult{}, ErrEventInternal
	}
	return result, nil
}

func meetingMutationReserveReceipt(ctx context.Context, transaction *EventTransaction, scope EventMutationScope, keyHash, requestHash MutationReceiptHash) (MutationReceiptReservationResult, error) {
	reservation, err := NewMutationReceiptReservation(scope, keyHash, requestHash)
	if err != nil {
		return MutationReceiptReservationResult{}, err
	}
	return transaction.ReserveMutationReceipt(ctx, scope, reservation)
}

func (service *MeetingService) appendMeetingUpdateOutbox(ctx context.Context, transaction *EventTransaction, scope EventMutationScope, desired MeetingCreateInput, current MeetingMutationSnapshot, persisted MeetingMutationPersistenceResult, plan meetingUpdatePlan) error {
	removedIDs := persisted.RemovedAttendeeIDs()
	activeAttendeeIDs := persisted.ActiveAttendeeIDs()
	sendCancel := plan.policy.CancelIntent == CancelIntentRemovedAttendees
	if !plan.hasChanges() || (len(removedIDs) != 0) != sendCancel {
		return ErrEventInternal
	}
	count := 1
	if plan.policy.SendRequest {
		count++
	}
	if plan.callChanged {
		count++
	}
	if sendCancel {
		count++
	}
	ids, occurredAt, err := service.meetingMutationEnvelopeIDs(count)
	if err != nil {
		return err
	}
	actorID := scope.ActorID()
	correlationID := current.EventID.String()
	meetingEnvelope, err := events.NewEnvelopeV1(
		"calendar.meeting.updated.v1", ids[0], occurredAt, scope.OrganizationID(), &actorID, correlationID,
		meetingMutationPayload{
			EventID: current.EventID, ScheduledCallID: current.ScheduledCallID, ProjectID: desired.ProjectID,
			ChannelID: desired.ChannelID, OrganizerID: actorID, AttendeeUserIDs: append([]uuid.UUID(nil), desired.AttendeeUserIDs...),
			Status: CalendarEventStatusConfirmed, Sequence: persisted.Sequence(),
		},
	)
	if err != nil {
		return ErrEventInternal
	}
	envelopes := make([]events.EnvelopeV1, 0, count)
	envelopes = append(envelopes, meetingEnvelope)
	nextID := 1
	if plan.policy.SendRequest {
		sourceEnvelopeID := ids[nextID]
		nextID++
		if len(activeAttendeeIDs) != 0 {
			inviteEnvelope, err := NewCalendarInviteDeliveryEnvelopeV2(
				CalendarInviteUpdatedSubjectV2, sourceEnvelopeID, occurredAt, scope.OrganizationID(), &actorID, correlationID,
				CalendarInviteDeliveryEnvelopeV2Input{
					EventID: current.EventID, EventSequence: persisted.Sequence(), OrganizerID: current.OwnerID,
					RecurrenceID: nil, AttendeeIDs: activeAttendeeIDs,
				},
			)
			if err != nil {
				return ErrEventInternal
			}
			envelopes = append(envelopes, inviteEnvelope)
		}
	}
	if plan.callChanged {
		callEnvelope, err := events.NewEnvelopeV1(
			"calls.channel."+desired.ChannelID.String()+".scheduled_updated.v1", ids[nextID], occurredAt,
			scope.OrganizationID(), &actorID, correlationID,
			meetingCallMutationPayload{
				ScheduledCallID: current.ScheduledCallID, CalendarEventID: current.EventID, ProjectID: desired.ProjectID,
				ChannelID: desired.ChannelID, CreatorID: actorID, IsActive: true, Version: persisted.CallVersion(),
			},
		)
		if err != nil {
			return ErrEventInternal
		}
		envelopes = append(envelopes, callEnvelope)
		nextID++
	}
	if sendCancel {
		sourceEnvelopeID := ids[nextID]
		if len(removedIDs) != 0 {
			cancelled, err := NewCalendarInviteDeliveryEnvelopeV2(
				CalendarInviteCancelledSubjectV2, sourceEnvelopeID, occurredAt, scope.OrganizationID(), &actorID, correlationID,
				CalendarInviteDeliveryEnvelopeV2Input{
					EventID: current.EventID, EventSequence: persisted.Sequence(), OrganizerID: current.OwnerID,
					RecurrenceID: nil, AttendeeIDs: removedIDs,
				},
			)
			if err != nil {
				return ErrEventInternal
			}
			envelopes = append(envelopes, cancelled)
		}
	}
	return appendMeetingMutationOutbox(ctx, transaction, scope, envelopes)
}

func (service *MeetingService) appendMeetingCancelOutbox(ctx context.Context, transaction *EventTransaction, scope EventMutationScope, current MeetingMutationSnapshot, persisted MeetingMutationPersistenceResult) error {
	ids, occurredAt, err := service.meetingMutationEnvelopeIDs(3)
	if err != nil {
		return err
	}
	actorID := scope.ActorID()
	correlationID := current.EventID.String()
	meetingEnvelope, err := events.NewEnvelopeV1(
		"calendar.meeting.cancelled.v1", ids[0], occurredAt, scope.OrganizationID(), &actorID, correlationID,
		meetingMutationPayload{
			EventID: current.EventID, ScheduledCallID: current.ScheduledCallID, ProjectID: current.ProjectID,
			ChannelID: current.ChannelID, OrganizerID: actorID, AttendeeUserIDs: current.activeInviteeUserIDs(),
			Status: CalendarEventStatusCancelled, Sequence: persisted.Sequence(),
		},
	)
	if err != nil {
		return ErrEventInternal
	}
	envelopes := make([]events.EnvelopeV1, 0, 3)
	envelopes = append(envelopes, meetingEnvelope)
	activeAttendeeIDs := persisted.ActiveAttendeeIDs()
	sourceEnvelopeID := ids[1]
	if len(activeAttendeeIDs) != 0 {
		inviteEnvelope, err := NewCalendarInviteDeliveryEnvelopeV2(
			CalendarInviteCancelledSubjectV2, sourceEnvelopeID, occurredAt, scope.OrganizationID(), &actorID, correlationID,
			CalendarInviteDeliveryEnvelopeV2Input{
				EventID: current.EventID, EventSequence: persisted.Sequence(), OrganizerID: current.OwnerID,
				RecurrenceID: nil, AttendeeIDs: activeAttendeeIDs,
			},
		)
		if err != nil {
			return ErrEventInternal
		}
		envelopes = append(envelopes, inviteEnvelope)
	}
	callEnvelope, err := events.NewEnvelopeV1(
		"calls.channel."+current.ChannelID.String()+".scheduled_cancelled.v1", ids[2], occurredAt,
		scope.OrganizationID(), &actorID, correlationID,
		meetingCallMutationPayload{
			ScheduledCallID: current.ScheduledCallID, CalendarEventID: current.EventID, ProjectID: current.ProjectID,
			ChannelID: current.ChannelID, CreatorID: actorID, IsActive: false, Version: persisted.CallVersion(),
		},
	)
	if err != nil {
		return ErrEventInternal
	}
	envelopes = append(envelopes, callEnvelope)
	return appendMeetingMutationOutbox(ctx, transaction, scope, envelopes)
}

func (service *MeetingService) meetingMutationEnvelopeIDs(count int) ([]uuid.UUID, time.Time, error) {
	if !service.ready() || count < 1 || count > 4 {
		return nil, time.Time{}, ErrEventInternal
	}
	occurredAt, ok := canonicalAvailabilityTime(service.clock.Now().UTC().Truncate(time.Second))
	if !ok {
		return nil, time.Time{}, ErrEventInternal
	}
	ids := make([]uuid.UUID, count)
	seen := make(map[uuid.UUID]struct{}, count)
	for index := range ids {
		ids[index] = service.ids.NewEventID()
		if ids[index] == uuid.Nil {
			return nil, time.Time{}, ErrEventInternal
		}
		if _, duplicate := seen[ids[index]]; duplicate {
			return nil, time.Time{}, ErrEventInternal
		}
		seen[ids[index]] = struct{}{}
	}
	return ids, occurredAt, nil
}

func appendMeetingMutationOutbox(ctx context.Context, transaction *EventTransaction, scope EventMutationScope, envelopes []events.EnvelopeV1) error {
	if transaction == nil || len(envelopes) == 0 {
		return ErrEventInternal
	}
	for _, envelope := range envelopes {
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

func meetingMutationSafeError(err error) error {
	var meetingConflict MeetingMutationConflictError
	if errors.As(err, &meetingConflict) {
		if meetingConflict.Kind() == MeetingMutationConflictVersion {
			eventID, eventOK := meetingConflict.EventID()
			version, versionOK := meetingConflict.CurrentVersion()
			if !eventOK || !versionOK {
				return ErrEventInternal
			}
			return meetingMutationVersionConflict(eventID, version)
		}
		return newMeetingMutationConflict(meetingConflict.Kind())
	}
	var createConflict *CalendarEventCreateConflictError
	if errors.As(err, &createConflict) && createConflict != nil &&
		createConflict.Kind() == CalendarEventCreateConflictIdempotencyKeyReused {
		return newMeetingMutationConflict(MeetingMutationConflictIdempotencyKeyReused)
	}
	if errors.Is(err, ErrAvailabilitySlotUnavailable) {
		return newMeetingMutationConflict(MeetingMutationConflictSlotUnavailable)
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
