package mail

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/calendarattendee"
	"github.com/valerii/hub/backend/ent/calendarevent"
	"github.com/valerii/hub/backend/ent/calendarinvitedelivery"
	"github.com/valerii/hub/backend/ent/calendarreplydelivery"
)

// AdmitCalendarInviteDeliveries persists the Calendar-owned, content-free
// invite transport intents. Every authoritative fact is re-read and locked
// from Calendar inside the write transaction; this repository never accepts
// Mail authority, recipients, or serialized calendar content from its caller.
func (repository *MailboxEntRepository) AdmitCalendarInviteDeliveries(ctx context.Context, request CalendarInviteDeliveryAdmissionRequest) error {
	if repository == nil || repository.client == nil || ctx == nil || !validCalendarInviteDeliveryAdmissionRequest(request) {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	request = calendarMailDeliveryAdmissionRequestCopy(request)
	sort.SliceStable(request.Invites, func(left, right int) bool {
		return request.Invites[left].InviteID.UUID().String() < request.Invites[right].InviteID.UUID().String()
	})

	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		for _, admission := range request.Invites {
			if err := mailboxEntAdmitCalendarInviteDelivery(ctx, transaction, admission, databaseNow); err != nil {
				return err
			}
		}
		return nil
	})
	return mailboxEntCalendarDeliveryRepositoryError(err)
}

// WakeCalendarReplyDelivery advances only a valid initial RSVP intent to the
// canonical database time. It intentionally does not create, claim, bind, or
// resolve Mail work; those lifecycles remain separate repository packets.
func (repository *MailboxEntRepository) WakeCalendarReplyDelivery(ctx context.Context, request CalendarReplyDeliveryWakeRequest) error {
	if repository == nil || repository.client == nil || ctx == nil || !validCalendarReplyDeliveryWakeRequest(request) {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}

	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		return mailboxEntWakeCalendarReplyDelivery(ctx, transaction, request, databaseNow)
	})
	return mailboxEntCalendarDeliveryRepositoryError(err)
}

func mailboxEntAdmitCalendarInviteDelivery(ctx context.Context, transaction *ent.Tx, admission CalendarInviteDeliveryAdmission, databaseNow time.Time) error {
	if ctx == nil || transaction == nil || !validCalendarInviteDeliveryAdmission(admission) || !mailboxEntValidTime(databaseNow) {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}

	event, err := transaction.CalendarEvent.Query().
		Where(calendarevent.IDEQ(admission.EventID)).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return mailboxEntCalendarDeliveryRowError(err)
	}
	attendee, err := transaction.CalendarAttendee.Query().
		Where(
			calendarattendee.IDEQ(admission.AttendeeID),
			calendarattendee.CalendarEventIDEQ(event.ID),
		).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return mailboxEntCalendarDeliveryRowError(err)
	}
	if !mailboxEntCalendarInviteDeliveryCalendarFactsValid(event, attendee, admission) {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}

	replayed, err := mailboxEntCalendarInviteDeliveryReplay(ctx, transaction, event, attendee, admission)
	if err != nil {
		return err
	}
	if replayed {
		return nil
	}

	create := transaction.CalendarInviteDelivery.Create().
		SetID(admission.InviteID.UUID()).
		SetOrganizationID(event.OrganizationID).
		SetCalendarEventID(event.ID).
		SetCalendarAttendeeID(attendee.ID).
		SetSourceEnvelopeID(admission.SourceEnvelopeID).
		SetKind(calendarinvitedelivery.Kind(admission.Kind)).
		SetEventSequence(admission.EventSequence).
		SetState(calendarinvitedelivery.State("pending")).
		SetAttempts(0).
		SetNextAttemptAt(databaseNow)
	if admission.RecurrenceID != "" {
		create.SetRecurrenceID(admission.RecurrenceID)
	}
	if _, err := create.Save(ctx); err != nil {
		if ent.IsConstraintError(err) {
			return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
		}
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return nil
}

func mailboxEntCalendarInviteDeliveryRows(ctx context.Context, transaction *ent.Tx, event *ent.CalendarEvent, attendee *ent.CalendarAttendee, admission CalendarInviteDeliveryAdmission) ([]*ent.CalendarInviteDelivery, error) {
	if ctx == nil || transaction == nil || event == nil || attendee == nil || !validCalendarInviteDeliveryAdmission(admission) {
		return nil, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}

	candidates := make(map[uuid.UUID]*ent.CalendarInviteDelivery, 3)
	collect := func(rows []*ent.CalendarInviteDelivery, err error) error {
		if err != nil {
			return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		for _, row := range rows {
			if row == nil || row.ID == uuid.Nil {
				return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
			}
			candidates[row.ID] = row
		}
		return nil
	}

	if err := collect(transaction.CalendarInviteDelivery.Query().
		Where(calendarinvitedelivery.IDEQ(admission.InviteID.UUID())).
		ForUpdate().
		All(ctx)); err != nil {
		return nil, err
	}
	if err := collect(transaction.CalendarInviteDelivery.Query().
		Where(
			calendarinvitedelivery.SourceEnvelopeIDEQ(admission.SourceEnvelopeID),
			calendarinvitedelivery.CalendarAttendeeIDEQ(attendee.ID),
		).
		ForUpdate().
		All(ctx)); err != nil {
		return nil, err
	}

	query := transaction.CalendarInviteDelivery.Query().Where(
		calendarinvitedelivery.CalendarEventIDEQ(event.ID),
		calendarinvitedelivery.CalendarAttendeeIDEQ(attendee.ID),
		calendarinvitedelivery.EventSequenceEQ(admission.EventSequence),
		calendarinvitedelivery.KindEQ(calendarinvitedelivery.Kind(admission.Kind)),
	)
	if admission.RecurrenceID == "" {
		query = query.Where(calendarinvitedelivery.RecurrenceIDIsNil())
	} else {
		query = query.Where(calendarinvitedelivery.RecurrenceIDEQ(admission.RecurrenceID))
	}
	if err := collect(query.ForUpdate().All(ctx)); err != nil {
		return nil, err
	}

	rows := make([]*ent.CalendarInviteDelivery, 0, len(candidates))
	for _, candidate := range candidates {
		rows = append(rows, candidate)
	}
	sort.Slice(rows, func(left, right int) bool {
		return rows[left].ID.String() < rows[right].ID.String()
	})
	return rows, nil
}

func mailboxEntCalendarInviteDeliveryReplay(ctx context.Context, transaction *ent.Tx, event *ent.CalendarEvent, attendee *ent.CalendarAttendee, admission CalendarInviteDeliveryAdmission) (bool, error) {
	candidates, err := mailboxEntCalendarInviteDeliveryRows(ctx, transaction, event, attendee, admission)
	if err != nil {
		return false, err
	}
	if len(candidates) == 0 {
		return false, nil
	}
	if len(candidates) != 1 || !mailboxEntCalendarInviteDeliveryImmutableMatch(candidates[0], event, attendee, admission) {
		return false, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}
	return true, nil
}

func mailboxEntCalendarInviteDeliveryImmutableMatch(row *ent.CalendarInviteDelivery, event *ent.CalendarEvent, attendee *ent.CalendarAttendee, admission CalendarInviteDeliveryAdmission) bool {
	return row != nil && event != nil && attendee != nil &&
		row.ID == admission.InviteID.UUID() &&
		row.OrganizationID == event.OrganizationID &&
		row.CalendarEventID == event.ID &&
		row.CalendarAttendeeID == attendee.ID &&
		row.SourceEnvelopeID == admission.SourceEnvelopeID &&
		string(row.Kind) == string(admission.Kind) &&
		row.EventSequence == admission.EventSequence &&
		mailboxEntCalendarDeliveryRecurrenceMatches(row.RecurrenceID, admission.RecurrenceID)
}

func mailboxEntWakeCalendarReplyDelivery(ctx context.Context, transaction *ent.Tx, request CalendarReplyDeliveryWakeRequest, databaseNow time.Time) error {
	if ctx == nil || transaction == nil || !validCalendarReplyDeliveryWakeRequest(request) || !mailboxEntValidTime(databaseNow) {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	delivery, event, attendee, err := mailboxEntCalendarReplyDeliveryRows(ctx, transaction, request)
	if err != nil {
		return err
	}
	if !mailboxEntCalendarReplyDeliveryCalendarFactsValid(delivery, event, attendee) {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}

	switch CalendarMailDeliveryState(delivery.State) {
	case CalendarMailDeliveryStateSent, CalendarMailDeliveryStateSuperseded, CalendarMailDeliveryStateFailedPermanent:
		if !mailboxEntCalendarReplyDeliveryTerminalValid(delivery) {
			return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
		}
		return nil
	case CalendarMailDeliveryStateUnknown:
		if !mailboxEntCalendarReplyDeliveryUnknownBoundValid(delivery) {
			return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
		}
		return nil
	case CalendarMailDeliveryStateSubmitting:
		if !mailboxEntCalendarReplyDeliverySubmittingValid(delivery) {
			return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
		}
		return nil
	case CalendarMailDeliveryStatePending:
		if !mailboxEntCalendarReplyDeliveryInitialPendingValid(delivery) {
			return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
		}
		if delivery.NextAttemptAt != nil && !delivery.NextAttemptAt.After(databaseNow) {
			return nil
		}
	default:
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}

	updated, err := transaction.CalendarReplyDelivery.Update().
		Where(
			calendarreplydelivery.IDEQ(delivery.ID),
			calendarreplydelivery.OrganizationIDEQ(delivery.OrganizationID),
			calendarreplydelivery.CalendarEventIDEQ(delivery.CalendarEventID),
			calendarreplydelivery.CalendarAttendeeIDEQ(delivery.CalendarAttendeeID),
			calendarreplydelivery.StateEQ(calendarreplydelivery.State("pending")),
			calendarreplydelivery.AttemptsEQ(0),
			calendarreplydelivery.ClaimGenerationEQ(0),
			calendarreplydelivery.MailSubmissionIDIsNil(),
			calendarreplydelivery.LeaseTokenIsNil(),
			calendarreplydelivery.LeaseUntilIsNil(),
			calendarreplydelivery.LastErrorCodeIsNil(),
			calendarreplydelivery.Or(
				calendarreplydelivery.NextAttemptAtIsNil(),
				calendarreplydelivery.NextAttemptAtGT(databaseNow),
			),
		).
		SetNextAttemptAt(databaseNow).
		SetUpdatedAt(databaseNow).
		Save(ctx)
	if err != nil {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	if updated != 1 {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}
	return nil
}

func mailboxEntCalendarReplyDeliveryRows(ctx context.Context, transaction *ent.Tx, request CalendarReplyDeliveryWakeRequest) (*ent.CalendarReplyDelivery, *ent.CalendarEvent, *ent.CalendarAttendee, error) {
	if ctx == nil || transaction == nil || !validCalendarReplyDeliveryWakeRequest(request) {
		return nil, nil, nil, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	delivery, err := transaction.CalendarReplyDelivery.Query().
		Where(calendarreplydelivery.IDEQ(request.ReplyID.UUID())).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return nil, nil, nil, mailboxEntCalendarDeliveryRowError(err)
	}
	event, err := transaction.CalendarEvent.Query().
		Where(calendarevent.IDEQ(delivery.CalendarEventID)).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return nil, nil, nil, mailboxEntCalendarDeliveryRowError(err)
	}
	attendee, err := transaction.CalendarAttendee.Query().
		Where(calendarattendee.IDEQ(delivery.CalendarAttendeeID)).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return nil, nil, nil, mailboxEntCalendarDeliveryRowError(err)
	}
	return delivery, event, attendee, nil
}

func mailboxEntCalendarInviteDeliveryCalendarFactsValid(event *ent.CalendarEvent, attendee *ent.CalendarAttendee, admission CalendarInviteDeliveryAdmission) bool {
	if event == nil || attendee == nil || event.ID == uuid.Nil || event.OrganizationID == uuid.Nil || attendee.ID == uuid.Nil ||
		attendee.CalendarEventID != event.ID || event.DeletedAt != nil || event.OrganizerID == nil || *event.OrganizerID != admission.OrganizerID ||
		event.Sequence < admission.EventSequence || !mailboxEntCalendarDeliveryRecurrenceMatches(attendee.RecurrenceID, admission.RecurrenceID) {
		return false
	}
	switch admission.Kind {
	case CalendarInviteDeliveryKindCreated, CalendarInviteDeliveryKindUpdated:
		return attendee.RemovedAt == nil
	case CalendarInviteDeliveryKindCancelled:
		return true
	default:
		return false
	}
}

func mailboxEntCalendarReplyDeliveryCalendarFactsValid(delivery *ent.CalendarReplyDelivery, event *ent.CalendarEvent, attendee *ent.CalendarAttendee) bool {
	return delivery != nil && event != nil && attendee != nil &&
		delivery.ID != uuid.Nil && delivery.OrganizationID != uuid.Nil && delivery.CalendarEventID != uuid.Nil && delivery.CalendarAttendeeID != uuid.Nil &&
		event.ID != uuid.Nil && event.OrganizationID != uuid.Nil && attendee.ID != uuid.Nil &&
		delivery.OrganizationID == event.OrganizationID && delivery.CalendarEventID == event.ID && delivery.CalendarAttendeeID == attendee.ID &&
		attendee.CalendarEventID == event.ID && mailboxEntCalendarDeliveryRecurrencePointersEqual(delivery.RecurrenceID, attendee.RecurrenceID)
}

func mailboxEntCalendarReplyDeliveryInitialPendingValid(delivery *ent.CalendarReplyDelivery) bool {
	if delivery == nil || CalendarMailDeliveryState(delivery.State) != CalendarMailDeliveryStatePending || delivery.Attempts != 0 || delivery.ClaimGeneration != 0 ||
		delivery.MailSubmissionID != nil || delivery.LeaseToken != nil || delivery.LeaseUntil != nil || delivery.LastErrorCode != nil {
		return false
	}
	return delivery.NextAttemptAt == nil || mailboxEntCalendarDeliveryTimestampValid(*delivery.NextAttemptAt)
}

func mailboxEntCalendarReplyDeliverySubmittingValid(delivery *ent.CalendarReplyDelivery) bool {
	return delivery != nil && CalendarMailDeliveryState(delivery.State) == CalendarMailDeliveryStateSubmitting && delivery.Attempts >= 1 && delivery.Attempts <= calendarMailDeliveryMaximumAttempts &&
		delivery.MailSubmissionID == nil && delivery.LeaseToken != nil && *delivery.LeaseToken != uuid.Nil && delivery.LeaseUntil != nil &&
		validCalendarMailDeliveryLeaseUntil(*delivery.LeaseUntil) && delivery.ClaimGeneration >= 1 && delivery.NextAttemptAt == nil && delivery.LastErrorCode == nil
}

func mailboxEntCalendarReplyDeliveryUnknownBoundValid(delivery *ent.CalendarReplyDelivery) bool {
	return delivery != nil && CalendarMailDeliveryState(delivery.State) == CalendarMailDeliveryStateUnknown && delivery.Attempts >= 1 && delivery.Attempts <= calendarMailDeliveryMaximumAttempts &&
		delivery.MailSubmissionID != nil && *delivery.MailSubmissionID != uuid.Nil && delivery.LeaseToken == nil && delivery.LeaseUntil == nil &&
		delivery.NextAttemptAt != nil && mailboxEntCalendarDeliveryTimestampValid(*delivery.NextAttemptAt) && delivery.LastErrorCode != nil
}

func mailboxEntCalendarReplyDeliveryTerminalValid(delivery *ent.CalendarReplyDelivery) bool {
	if delivery == nil || delivery.LeaseToken != nil || delivery.LeaseUntil != nil {
		return false
	}
	switch CalendarMailDeliveryState(delivery.State) {
	case CalendarMailDeliveryStateSent:
		return delivery.Attempts >= 1 && delivery.Attempts <= calendarMailDeliveryMaximumAttempts && delivery.MailSubmissionID != nil && *delivery.MailSubmissionID != uuid.Nil && delivery.NextAttemptAt == nil && delivery.LastErrorCode == nil
	case CalendarMailDeliveryStateSuperseded:
		return delivery.MailSubmissionID == nil && delivery.NextAttemptAt == nil && delivery.LastErrorCode == nil
	case CalendarMailDeliveryStateFailedPermanent:
		return delivery.Attempts >= 1 && delivery.Attempts <= calendarMailDeliveryMaximumAttempts && delivery.MailSubmissionID == nil && delivery.NextAttemptAt == nil && delivery.LastErrorCode != nil
	default:
		return false
	}
}

func mailboxEntCalendarDeliveryRecurrenceMatches(value *string, recurrenceID string) bool {
	if recurrenceID == "" {
		return value == nil
	}
	return value != nil && *value == recurrenceID
}

func mailboxEntCalendarDeliveryRecurrencePointersEqual(left, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func mailboxEntCalendarDeliveryTimestampValid(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC
}

func mailboxEntCalendarDeliveryRowError(err error) error {
	if ent.IsNotFound(err) {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}
	return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
}

func mailboxEntCalendarDeliveryRepositoryError(err error) error {
	if err == nil {
		return nil
	}
	var repositoryError *CalendarMailDeliveryRepositoryError
	if errors.As(err, &repositoryError) && repositoryError != nil {
		return newCalendarMailDeliveryRepositoryError(repositoryError.Kind())
	}
	if ent.IsConstraintError(err) {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}
	return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
}
