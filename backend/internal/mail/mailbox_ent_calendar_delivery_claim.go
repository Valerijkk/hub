package mail

import (
	"context"
	"time"

	"github.com/google/uuid"
	ical "github.com/valerii/hub/backend/internal/ical"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/calendarattendee"
	"github.com/valerii/hub/backend/ent/calendarevent"
	"github.com/valerii/hub/backend/ent/calendarinvitedelivery"
	"github.com/valerii/hub/backend/ent/calendarreplydelivery"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/mailidentity"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/user"
)

const mailboxEntCalendarDeliveryMaximumClaimGeneration int64 = 2147483647

// ClaimCalendarMailDelivery returns an external capability only for an
// unbound master delivery that this transaction has atomically fenced. Calendar
// facts and the canonical iTIP snapshot are validated before the state CAS so
// a malformed projection cannot strand a row in submitting.
func (repository *MailboxEntRepository) ClaimCalendarMailDelivery(ctx context.Context, deliveryID CalendarMailDeliveryID) (CalendarMailDeliveryClaimResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || ctx.Err() != nil || !validCalendarMailDeliveryID(deliveryID) {
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}

	result := CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeNoop}
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		var claimErr error
		switch deliveryID.Kind() {
		case CalendarMailDeliveryKindInvite:
			result, claimErr = mailboxEntClaimCalendarInviteDelivery(ctx, transaction, deliveryID)
		case CalendarMailDeliveryKindReply:
			result, claimErr = mailboxEntClaimCalendarReplyDelivery(ctx, transaction, deliveryID)
		default:
			claimErr = newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
		}
		return claimErr
	})
	if err != nil {
		return CalendarMailDeliveryClaimResult{}, mailboxEntCalendarDeliveryRepositoryError(err)
	}
	if !validCalendarMailDeliveryClaimResult(result) {
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return result, nil
}

func mailboxEntClaimCalendarInviteDelivery(ctx context.Context, transaction *ent.Tx, deliveryID CalendarMailDeliveryID) (CalendarMailDeliveryClaimResult, error) {
	if ctx == nil || transaction == nil || deliveryID.Kind() != CalendarMailDeliveryKindInvite || !validCalendarMailDeliveryID(deliveryID) {
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}

	delivery, err := transaction.CalendarInviteDelivery.Query().
		Where(calendarinvitedelivery.IDEQ(deliveryID.UUID())).
		ForUpdate().
		Only(ctx)
	if ent.IsNotFound(err) {
		return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeNoop}, nil
	}
	if err != nil || delivery == nil || delivery.ID != deliveryID.UUID() || delivery.RecurrenceID != nil {
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}

	databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
	if err != nil || !mailboxEntValidTime(databaseNow) {
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}

	switch CalendarMailDeliveryState(delivery.State) {
	case CalendarMailDeliveryStateUnknown:
		if !mailboxEntCalendarInviteDeliveryUnknownBoundValid(delivery) {
			return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		if delivery.NextAttemptAt.After(databaseNow) {
			return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeNoop}, nil
		}
		return mailboxEntObserveCalendarInviteDeliveryBoundSubmission(ctx, transaction, delivery, deliveryID, databaseNow)
	case CalendarMailDeliveryStateSent, CalendarMailDeliveryStateFailedPermanent:
		if !mailboxEntCalendarInviteDeliveryTerminalValid(delivery) {
			return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeNoop}, nil
	case CalendarMailDeliveryStateSuperseded:
		if !mailboxEntCalendarInviteDeliveryTerminalValid(delivery) {
			return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeSuperseded}, nil
	case CalendarMailDeliveryStatePending:
		if !mailboxEntCalendarInviteDeliveryPendingValid(delivery) {
			return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		if delivery.NextAttemptAt.After(databaseNow) {
			return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeNoop}, nil
		}
		return mailboxEntClaimDueCalendarInviteDelivery(ctx, transaction, delivery, deliveryID, databaseNow, false)
	case CalendarMailDeliveryStateSubmitting:
		if !mailboxEntCalendarInviteDeliverySubmittingValid(delivery) {
			return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		if delivery.LeaseUntil.After(databaseNow) {
			return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeBusy}, nil
		}
		return mailboxEntClaimDueCalendarInviteDelivery(ctx, transaction, delivery, deliveryID, databaseNow, true)
	default:
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
}

func mailboxEntClaimCalendarReplyDelivery(ctx context.Context, transaction *ent.Tx, deliveryID CalendarMailDeliveryID) (CalendarMailDeliveryClaimResult, error) {
	if ctx == nil || transaction == nil || deliveryID.Kind() != CalendarMailDeliveryKindReply || !validCalendarMailDeliveryID(deliveryID) {
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}

	delivery, err := transaction.CalendarReplyDelivery.Query().
		Where(calendarreplydelivery.IDEQ(deliveryID.UUID())).
		ForUpdate().
		Only(ctx)
	if ent.IsNotFound(err) {
		return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeNoop}, nil
	}
	if err != nil || delivery == nil || delivery.ID != deliveryID.UUID() || delivery.RecurrenceID != nil {
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}

	databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
	if err != nil || !mailboxEntValidTime(databaseNow) {
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}

	switch CalendarMailDeliveryState(delivery.State) {
	case CalendarMailDeliveryStateUnknown:
		if !mailboxEntCalendarReplyDeliveryUnknownBoundValid(delivery) {
			return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		if delivery.NextAttemptAt.After(databaseNow) {
			return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeNoop}, nil
		}
		return mailboxEntObserveCalendarReplyDeliveryBoundSubmission(ctx, transaction, delivery, deliveryID, databaseNow)
	case CalendarMailDeliveryStateSent, CalendarMailDeliveryStateFailedPermanent:
		if !mailboxEntCalendarReplyDeliveryTerminalValid(delivery) {
			return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeNoop}, nil
	case CalendarMailDeliveryStateSuperseded:
		if !mailboxEntCalendarReplyDeliveryTerminalValid(delivery) {
			return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeSuperseded}, nil
	case CalendarMailDeliveryStatePending:
		if mailboxEntCalendarReplyDeliveryInitialPendingUnscheduledValid(delivery) {
			return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeNoop}, nil
		}
		if !mailboxEntCalendarReplyDeliveryScheduledPendingValid(delivery) {
			return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		if delivery.NextAttemptAt.After(databaseNow) {
			return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeNoop}, nil
		}
		return mailboxEntClaimDueCalendarReplyDelivery(ctx, transaction, delivery, deliveryID, databaseNow, false)
	case CalendarMailDeliveryStateSubmitting:
		if !mailboxEntCalendarReplyDeliverySubmittingValid(delivery) {
			return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		if delivery.LeaseUntil.After(databaseNow) {
			return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeBusy}, nil
		}
		return mailboxEntClaimDueCalendarReplyDelivery(ctx, transaction, delivery, deliveryID, databaseNow, true)
	default:
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
}

type mailboxEntCalendarDeliveryClaimSnapshot struct {
	organizationID  uuid.UUID
	eventID         uuid.UUID
	attendeeID      uuid.UUID
	organizerID     uuid.UUID
	accountID       uuid.UUID
	senderActorID   uuid.UUID
	fromIdentityID  uuid.UUID
	recipient       EngineAddress
	method          CalendarMailDeliveryMethod
	eventSequence   int
	attendeeVersion int
	message         ical.ITIPMessage
}

func mailboxEntClaimDueCalendarInviteDelivery(ctx context.Context, transaction *ent.Tx, delivery *ent.CalendarInviteDelivery, deliveryID CalendarMailDeliveryID, databaseNow time.Time, reclaim bool) (CalendarMailDeliveryClaimResult, error) {
	if ctx == nil || transaction == nil || delivery == nil || deliveryID.Kind() != CalendarMailDeliveryKindInvite || !mailboxEntValidTime(databaseNow) || reclaim != (CalendarMailDeliveryState(delivery.State) == CalendarMailDeliveryStateSubmitting) {
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	nextGeneration, leaseUntil, token, ok := mailboxEntCalendarDeliveryClaimLease(delivery.ClaimGeneration, databaseNow)
	if !ok || (!reclaim && delivery.Attempts >= calendarMailDeliveryMaximumAttempts) {
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	snapshot, err := mailboxEntCalendarInviteDeliveryClaimSnapshot(ctx, transaction, delivery, databaseNow)
	if err != nil {
		return CalendarMailDeliveryClaimResult{}, err
	}
	claim, err := mailboxEntCalendarInviteDeliveryClaim(delivery, deliveryID, snapshot, token, nextGeneration, leaseUntil)
	if err != nil {
		return CalendarMailDeliveryClaimResult{}, err
	}

	updated, err := mailboxEntCASCalendarInviteDeliveryClaim(ctx, transaction, delivery, databaseNow, token, nextGeneration, leaseUntil, reclaim)
	if err != nil {
		return CalendarMailDeliveryClaimResult{}, err
	}
	if updated != 1 {
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}

	return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeClaimed, claim: claim}, nil
}

func mailboxEntClaimDueCalendarReplyDelivery(ctx context.Context, transaction *ent.Tx, delivery *ent.CalendarReplyDelivery, deliveryID CalendarMailDeliveryID, databaseNow time.Time, reclaim bool) (CalendarMailDeliveryClaimResult, error) {
	if ctx == nil || transaction == nil || delivery == nil || deliveryID.Kind() != CalendarMailDeliveryKindReply || !mailboxEntValidTime(databaseNow) || reclaim != (CalendarMailDeliveryState(delivery.State) == CalendarMailDeliveryStateSubmitting) {
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	nextGeneration, leaseUntil, token, ok := mailboxEntCalendarDeliveryClaimLease(delivery.ClaimGeneration, databaseNow)
	if !ok || (!reclaim && delivery.Attempts >= calendarMailDeliveryMaximumAttempts) {
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	snapshot, err := mailboxEntCalendarReplyDeliveryClaimSnapshot(ctx, transaction, delivery, databaseNow)
	if err != nil {
		return CalendarMailDeliveryClaimResult{}, err
	}
	claim, err := mailboxEntCalendarReplyDeliveryClaim(delivery, deliveryID, snapshot, token, nextGeneration, leaseUntil)
	if err != nil {
		return CalendarMailDeliveryClaimResult{}, err
	}

	updated, err := mailboxEntCASCalendarReplyDeliveryClaim(ctx, transaction, delivery, databaseNow, token, nextGeneration, leaseUntil, reclaim)
	if err != nil {
		return CalendarMailDeliveryClaimResult{}, err
	}
	if updated != 1 {
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorConflict)
	}

	return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeClaimed, claim: claim}, nil
}

func mailboxEntCalendarDeliveryClaimLease(generation int64, databaseNow time.Time) (CalendarMailDeliveryClaimGeneration, time.Time, CalendarMailDeliveryClaimToken, bool) {
	if generation < 0 || generation >= mailboxEntCalendarDeliveryMaximumClaimGeneration || !mailboxEntValidTime(databaseNow) {
		return 0, time.Time{}, CalendarMailDeliveryClaimToken{}, false
	}
	leaseUntil, ok := mailboxEntCheckedTimeAdd(databaseNow, calendarMailDeliveryLease)
	if !ok || !leaseUntil.After(databaseNow) {
		return 0, time.Time{}, CalendarMailDeliveryClaimToken{}, false
	}
	token := CalendarMailDeliveryClaimToken(uuid.New())
	if uuid.UUID(token) == uuid.Nil {
		return 0, time.Time{}, CalendarMailDeliveryClaimToken{}, false
	}
	return CalendarMailDeliveryClaimGeneration(generation + 1), leaseUntil, token, true
}

func mailboxEntCalendarInviteDeliveryClaimSnapshot(ctx context.Context, transaction *ent.Tx, delivery *ent.CalendarInviteDelivery, databaseNow time.Time) (mailboxEntCalendarDeliveryClaimSnapshot, error) {
	if ctx == nil || transaction == nil || delivery == nil || delivery.RecurrenceID != nil || !mailboxEntValidTime(databaseNow) {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}

	event, err := transaction.CalendarEvent.Query().
		Where(
			calendarevent.IDEQ(delivery.CalendarEventID),
			calendarevent.OrganizationIDEQ(delivery.OrganizationID),
		).
		ForUpdate().
		Only(ctx)
	if err != nil || event == nil || event.DeletedAt != nil || event.OrganizerID == nil || *event.OrganizerID == uuid.Nil ||
		event.Sequence != delivery.EventSequence || event.ID != delivery.CalendarEventID || event.OrganizationID != delivery.OrganizationID {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}

	attendees, err := transaction.CalendarAttendee.Query().
		Where(
			calendarattendee.CalendarEventIDEQ(event.ID),
			calendarattendee.RecurrenceIDIsNil(),
		).
		Order(ent.Asc(calendarattendee.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil || len(attendees) == 0 {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}

	var target *ent.CalendarAttendee
	participants := make([]ical.ITIPParticipant, 0, len(attendees))
	for _, attendee := range attendees {
		if attendee == nil || attendee.ID == uuid.Nil || attendee.CalendarEventID != event.ID || attendee.RecurrenceID != nil {
			return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		if attendee.ID == delivery.CalendarAttendeeID {
			target = attendee
		}
		if attendee.RemovedAt != nil && attendee.ID != delivery.CalendarAttendeeID {
			continue
		}
		participant, ok := mailboxEntCalendarDeliveryITIPParticipant(attendee, attendee.ID == delivery.CalendarAttendeeID && delivery.Kind != calendarinvitedelivery.KindCancelled)
		if !ok {
			return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		participants = append(participants, participant)
	}
	if target == nil || target.CalendarEventID != event.ID || target.RecurrenceID != nil ||
		!mailboxEntCalendarInviteDeliveryFactsMatch(delivery, event, target) {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}

	sourceUser, account, identity, err := mailboxEntCalendarDeliverySourceIdentity(ctx, transaction, event.OrganizationID, *event.OrganizerID)
	if err != nil {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, err
	}
	recipient, err := NormalizeAddress(EngineAddress{Email: target.NormalizedEmail, Name: mailboxEntCalendarAttendeeDisplayName(target)})
	if err != nil {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	method, ok := mailboxEntCalendarInviteDeliveryMethod(delivery.Kind)
	if !ok {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	message, ok := mailboxEntCalendarDeliveryMasterITIPMessage(event, method, mailboxEntCalendarUserDisplayName(sourceUser), event.OrganizerEmail, participants, databaseNow)
	if !ok || !calendarMailDeliveryMessageHasAttendee(message, recipient.Email) {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return mailboxEntCalendarDeliveryClaimSnapshot{
		organizationID: event.OrganizationID,
		eventID:        event.ID,
		attendeeID:     target.ID,
		organizerID:    *event.OrganizerID,
		accountID:      account.ID,
		senderActorID:  sourceUser.ID,
		fromIdentityID: identity.ID,
		recipient:      recipient,
		method:         method,
		eventSequence:  delivery.EventSequence,
		message:        message,
	}, nil
}

func mailboxEntCalendarReplyDeliveryClaimSnapshot(ctx context.Context, transaction *ent.Tx, delivery *ent.CalendarReplyDelivery, databaseNow time.Time) (mailboxEntCalendarDeliveryClaimSnapshot, error) {
	if ctx == nil || transaction == nil || delivery == nil || delivery.RecurrenceID != nil || !mailboxEntValidTime(databaseNow) {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}

	event, err := transaction.CalendarEvent.Query().
		Where(
			calendarevent.IDEQ(delivery.CalendarEventID),
			calendarevent.OrganizationIDEQ(delivery.OrganizationID),
		).
		ForUpdate().
		Only(ctx)
	if err != nil || event == nil || event.DeletedAt != nil || event.OrganizerID != nil || event.Sequence != delivery.EventSequence ||
		event.ID != delivery.CalendarEventID || event.OrganizationID != delivery.OrganizationID {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}

	attendee, err := transaction.CalendarAttendee.Query().
		Where(
			calendarattendee.IDEQ(delivery.CalendarAttendeeID),
			calendarattendee.CalendarEventIDEQ(event.ID),
			calendarattendee.RecurrenceIDIsNil(),
		).
		ForUpdate().
		Only(ctx)
	if err != nil || attendee == nil || attendee.RemovedAt != nil || attendee.UserID == nil || *attendee.UserID == uuid.Nil ||
		attendee.CalendarEventID != event.ID || attendee.RecurrenceID != nil || attendee.RSVPStatus != calendarattendee.RSVPStatus(delivery.RSVPStatus) ||
		delivery.AttendeeVersion < 1 {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}

	sourceUser, account, identity, err := mailboxEntCalendarDeliverySourceIdentity(ctx, transaction, event.OrganizationID, *attendee.UserID)
	if err != nil {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, err
	}
	recipient, err := NormalizeAddress(EngineAddress{Email: event.OrganizerEmail})
	if err != nil {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	participant, ok := mailboxEntCalendarDeliveryITIPParticipant(attendee, false)
	if !ok || participant.Status == ical.RSVPNeedsAction {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	message, ok := mailboxEntCalendarDeliveryMasterITIPMessage(event, CalendarMailDeliveryMethodReply, "", recipient.Email, []ical.ITIPParticipant{participant}, databaseNow)
	if !ok || len(message.Attendees) != 1 || message.Attendees[0].Email != attendee.NormalizedEmail || message.Organizer.Email != recipient.Email {
		return mailboxEntCalendarDeliveryClaimSnapshot{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return mailboxEntCalendarDeliveryClaimSnapshot{
		organizationID:  event.OrganizationID,
		eventID:         event.ID,
		attendeeID:      attendee.ID,
		accountID:       account.ID,
		senderActorID:   sourceUser.ID,
		fromIdentityID:  identity.ID,
		recipient:       recipient,
		method:          CalendarMailDeliveryMethodReply,
		eventSequence:   delivery.EventSequence,
		attendeeVersion: delivery.AttendeeVersion,
		message:         message,
	}, nil
}

// mailboxEntCalendarDeliverySourceIdentity chooses the only primary identity
// that can be used for calendar delivery. Multiple or incomplete records are
// intentionally unavailable rather than silently selecting a sender.
func mailboxEntCalendarDeliverySourceIdentity(ctx context.Context, transaction *ent.Tx, organizationID, sourceUserID uuid.UUID) (*ent.User, *ent.MailAccount, *ent.MailIdentity, error) {
	if ctx == nil || transaction == nil || organizationID == uuid.Nil || sourceUserID == uuid.Nil {
		return nil, nil, nil, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}

	users, err := transaction.User.Query().
		Where(
			user.IDEQ(sourceUserID),
			user.StatusEQ(user.StatusActive),
			user.HasOrganizationWith(organization.IDEQ(organizationID)),
		).
		Order(ent.Asc(user.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil || len(users) != 1 || users[0] == nil || users[0].ID != sourceUserID {
		return nil, nil, nil, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}

	accounts, err := transaction.MailAccount.Query().
		Where(
			mailaccount.OrganizationIDEQ(organizationID),
			mailaccount.UserIDEQ(sourceUserID),
			mailaccount.StatusEQ(mailaccount.StatusActive),
			mailaccount.DesiredEnabledEQ(true),
			mailaccount.EngineAccountIDNotNil(),
		).
		Order(ent.Asc(mailaccount.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil || len(accounts) != 1 || accounts[0] == nil || accounts[0].ID == uuid.Nil || accounts[0].OrganizationID != organizationID ||
		accounts[0].UserID != sourceUserID || accounts[0].Status != mailaccount.StatusActive || !accounts[0].DesiredEnabled ||
		accounts[0].EngineAccountID == nil || *accounts[0].EngineAccountID == "" {
		return nil, nil, nil, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}

	identities, err := transaction.MailIdentity.Query().
		Where(
			mailidentity.MailAccountIDEQ(accounts[0].ID),
			mailidentity.KindEQ(mailidentity.KindPrimary),
			mailidentity.StatusEQ(mailidentity.StatusActive),
			mailidentity.SendAllowedEQ(true),
			mailidentity.EngineIdentityIDNotNil(),
		).
		Order(ent.Asc(mailidentity.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil || len(identities) != 1 || identities[0] == nil || identities[0].ID == uuid.Nil || identities[0].MailAccountID != accounts[0].ID ||
		identities[0].Kind != mailidentity.KindPrimary || identities[0].Status != mailidentity.StatusActive || !identities[0].SendAllowed ||
		identities[0].EngineIdentityID == nil || *identities[0].EngineIdentityID == "" {
		return nil, nil, nil, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return users[0], accounts[0], identities[0], nil
}

func mailboxEntCalendarInviteDeliveryFactsMatch(delivery *ent.CalendarInviteDelivery, event *ent.CalendarEvent, attendee *ent.CalendarAttendee) bool {
	if delivery == nil || event == nil || attendee == nil || delivery.ID == uuid.Nil || delivery.OrganizationID == uuid.Nil ||
		delivery.CalendarEventID != event.ID || delivery.CalendarAttendeeID != attendee.ID || event.OrganizationID != delivery.OrganizationID ||
		attendee.CalendarEventID != event.ID || event.DeletedAt != nil || event.OrganizerID == nil || *event.OrganizerID == uuid.Nil ||
		event.Sequence != delivery.EventSequence || delivery.RecurrenceID != nil || attendee.RecurrenceID != nil {
		return false
	}
	switch delivery.Kind {
	case calendarinvitedelivery.KindCreated, calendarinvitedelivery.KindUpdated:
		return attendee.RemovedAt == nil && (event.Status == calendarevent.StatusConfirmed || event.Status == calendarevent.StatusTentative)
	case calendarinvitedelivery.KindCancelled:
		return event.Status == calendarevent.StatusCancelled
	default:
		return false
	}
}

func mailboxEntCalendarInviteDeliveryMethod(kind calendarinvitedelivery.Kind) (CalendarMailDeliveryMethod, bool) {
	switch kind {
	case calendarinvitedelivery.KindCreated, calendarinvitedelivery.KindUpdated:
		return CalendarMailDeliveryMethodRequest, true
	case calendarinvitedelivery.KindCancelled:
		return CalendarMailDeliveryMethodCancel, true
	default:
		return "", false
	}
}

func mailboxEntCalendarDeliveryITIPParticipant(attendee *ent.CalendarAttendee, rsvp bool) (ical.ITIPParticipant, bool) {
	if attendee == nil || attendee.ID == uuid.Nil || attendee.CalendarEventID == uuid.Nil || attendee.RecurrenceID != nil {
		return ical.ITIPParticipant{}, false
	}
	role := ical.ITIPParticipantRequired
	switch attendee.Role {
	case calendarattendee.RoleOptional:
		role = ical.ITIPParticipantOptional
	case calendarattendee.RoleOrganizer, calendarattendee.RoleRequired:
	default:
		return ical.ITIPParticipant{}, false
	}
	status, ok := mailboxEntCalendarDeliveryRSVPStatus(attendee.RSVPStatus)
	if !ok {
		return ical.ITIPParticipant{}, false
	}
	return ical.ITIPParticipant{
		Email:       attendee.NormalizedEmail,
		DisplayName: mailboxEntCalendarAttendeeDisplayName(attendee),
		Role:        role,
		Status:      status,
		RSVP:        rsvp,
	}, true
}

func mailboxEntCalendarDeliveryRSVPStatus(status calendarattendee.RSVPStatus) (ical.RSVPStatus, bool) {
	switch status {
	case calendarattendee.RSVPStatusNeedsAction:
		return ical.RSVPNeedsAction, true
	case calendarattendee.RSVPStatusAccepted:
		return ical.RSVPAccepted, true
	case calendarattendee.RSVPStatusDeclined:
		return ical.RSVPDeclined, true
	case calendarattendee.RSVPStatusTentative:
		return ical.RSVPTentative, true
	default:
		return "", false
	}
}

func mailboxEntCalendarDeliveryMasterITIPMessage(event *ent.CalendarEvent, method CalendarMailDeliveryMethod, organizerDisplayName, organizerEmail string, attendees []ical.ITIPParticipant, databaseNow time.Time) (ical.ITIPMessage, bool) {
	if event == nil || event.ID == uuid.Nil || event.OrganizationID == uuid.Nil || event.DeletedAt != nil || !mailboxEntValidTime(databaseNow) {
		return ical.ITIPMessage{}, false
	}
	itipMethod, ok := mailboxEntCalendarDeliveryITIPMethod(method)
	if !ok {
		return ical.ITIPMessage{}, false
	}
	status, ok := mailboxEntCalendarDeliveryITIPStatus(event.Status)
	if !ok || (method == CalendarMailDeliveryMethodCancel && status != ical.ITIPStatusCancelled) ||
		((method == CalendarMailDeliveryMethodRequest || method == CalendarMailDeliveryMethodReply) && status == ical.ITIPStatusCancelled) {
		return ical.ITIPMessage{}, false
	}
	temporal, exdates, ok := mailboxEntCalendarDeliveryMasterTemporal(event)
	if !ok {
		return ical.ITIPMessage{}, false
	}
	message := ical.ITIPMessage{
		Method:      itipMethod,
		UID:         event.UID,
		Stamp:       databaseNow,
		Sequence:    event.Sequence,
		Status:      status,
		Summary:     event.Title,
		Description: event.Description,
		Location:    event.Location,
		Temporal:    temporal,
		Organizer:  ical.ITIPOrganizer{Email: organizerEmail, DisplayName: organizerDisplayName},
		Attendees:  append([]ical.ITIPParticipant(nil), attendees...),
		RRULE:      mailboxEntCalendarEventRRULE(event),
		EXDates:    exdates,
	}
	normalized, err := ical.NormalizeITIPMessage(message)
	if err != nil {
		return ical.ITIPMessage{}, false
	}
	identity, err := ical.CanonicalITIPIdentity(normalized)
	if err != nil || identity.Method != itipMethod || identity.UID != normalized.UID || identity.Sequence != event.Sequence ||
		identity.HasRecurrenceID || identity.RecurrenceID != "" {
		return ical.ITIPMessage{}, false
	}
	return calendarMailDeliveryITIPMessageCopy(normalized), true
}

func mailboxEntCalendarDeliveryITIPMethod(method CalendarMailDeliveryMethod) (ical.ITIPMethod, bool) {
	switch method {
	case CalendarMailDeliveryMethodRequest:
		return ical.ITIPMethodRequest, true
	case CalendarMailDeliveryMethodCancel:
		return ical.ITIPMethodCancel, true
	case CalendarMailDeliveryMethodReply:
		return ical.ITIPMethodReply, true
	default:
		return "", false
	}
}

func mailboxEntCalendarDeliveryITIPStatus(status calendarevent.Status) (ical.ITIPStatus, bool) {
	switch status {
	case calendarevent.StatusConfirmed:
		return ical.ITIPStatusConfirmed, true
	case calendarevent.StatusTentative:
		return ical.ITIPStatusTentative, true
	case calendarevent.StatusCancelled:
		return ical.ITIPStatusCancelled, true
	default:
		return "", false
	}
}

func mailboxEntCalendarDeliveryMasterTemporal(event *ent.CalendarEvent) (ical.ITIPTemporal, []ical.ITIPRecurrenceID, bool) {
	if event == nil {
		return ical.ITIPTemporal{}, nil, false
	}
	var temporal ical.ITIPTemporal
	if event.AllDay {
		if event.StartDate == nil || event.EndDateExclusive == nil || event.StartsAt != nil || event.EndsAt != nil {
			return ical.ITIPTemporal{}, nil, false
		}
		temporal.AllDay = &ical.ITIPAllDay{
			StartDate:        event.StartDate.Format("2006-01-02"),
			EndDateExclusive: event.EndDateExclusive.Format("2006-01-02"),
			Timezone:         event.Timezone,
		}
	} else {
		if event.StartsAt == nil || event.EndsAt == nil || event.StartDate != nil || event.EndDateExclusive != nil {
			return ical.ITIPTemporal{}, nil, false
		}
		temporal.Timed = &ical.ITIPTimed{
			StartsAt: *event.StartsAt,
			EndsAt:   *event.EndsAt,
			Timezone: event.Timezone,
		}
	}

	exdates := make([]ical.ITIPRecurrenceID, 0, len(event.Exdates))
	for _, value := range event.Exdates {
		if temporal.AllDay != nil {
			parsed, err := ical.ParseCalendarDate(value)
			if err != nil || parsed.Format("2006-01-02") != value {
				return ical.ITIPTemporal{}, nil, false
			}
			exdates = append(exdates, ical.ITIPRecurrenceID{Date: value})
			continue
		}
		parsed, err := time.Parse("2006-01-02T15:04:05Z", value)
		if err != nil || parsed.Location() != time.UTC || parsed.Format("2006-01-02T15:04:05Z") != value {
			return ical.ITIPTemporal{}, nil, false
		}
		exdates = append(exdates, ical.ITIPRecurrenceID{Time: parsed})
	}
	return temporal, exdates, true
}

func mailboxEntCalendarEventRRULE(event *ent.CalendarEvent) string {
	if event == nil || event.Rrule == nil {
		return ""
	}
	return *event.Rrule
}

func mailboxEntCalendarAttendeeDisplayName(attendee *ent.CalendarAttendee) string {
	if attendee == nil || attendee.DisplayName == nil {
		return ""
	}
	return *attendee.DisplayName
}

func mailboxEntCalendarUserDisplayName(sourceUser *ent.User) string {
	if sourceUser == nil {
		return ""
	}
	if sourceUser.DisplayName != "" {
		return sourceUser.DisplayName
	}
	return sourceUser.Name
}

func mailboxEntCalendarInviteDeliveryClaim(delivery *ent.CalendarInviteDelivery, deliveryID CalendarMailDeliveryID, snapshot mailboxEntCalendarDeliveryClaimSnapshot, token CalendarMailDeliveryClaimToken, generation CalendarMailDeliveryClaimGeneration, leaseUntil time.Time) (CalendarMailDeliveryClaim, error) {
	inviteID, err := NewCalendarInviteDeliveryID(delivery.ID)
	if err != nil || deliveryID.Kind() != CalendarMailDeliveryKindInvite || snapshot.organizerID == uuid.Nil || snapshot.attendeeVersion != 0 {
		return CalendarMailDeliveryClaim{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	claim := CalendarMailDeliveryClaim{
		deliveryID:      deliveryID,
		inviteID:        inviteID,
		kind:            CalendarMailDeliveryKindInvite,
		method:          snapshot.method,
		state:           CalendarMailDeliveryStateSubmitting,
		claimToken:      token,
		claimGeneration: generation,
		organizationID:  snapshot.organizationID,
		leaseUntil:      leaseUntil,
		eventID:         snapshot.eventID,
		attendeeID:      snapshot.attendeeID,
		organizerID:     snapshot.organizerID,
		accountID:       snapshot.accountID,
		senderActorID:   snapshot.senderActorID,
		fromIdentityID:  snapshot.fromIdentityID,
		recipient:       snapshot.recipient,
		materialization: CalendarMailDeliveryMaterializationInput{
			deliveryID:    deliveryID,
			kind:          CalendarMailDeliveryKindInvite,
			method:        snapshot.method,
			eventID:       snapshot.eventID,
			attendeeID:    snapshot.attendeeID,
			organizerID:   snapshot.organizerID,
			eventSequence: snapshot.eventSequence,
			message:       snapshot.message,
		},
	}
	if !validCalendarMailDeliveryClaim(claim) {
		return CalendarMailDeliveryClaim{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return calendarMailDeliveryClaimCopy(claim), nil
}

func mailboxEntCalendarReplyDeliveryClaim(delivery *ent.CalendarReplyDelivery, deliveryID CalendarMailDeliveryID, snapshot mailboxEntCalendarDeliveryClaimSnapshot, token CalendarMailDeliveryClaimToken, generation CalendarMailDeliveryClaimGeneration, leaseUntil time.Time) (CalendarMailDeliveryClaim, error) {
	replyID, err := NewCalendarReplyDeliveryID(delivery.ID)
	if err != nil || deliveryID.Kind() != CalendarMailDeliveryKindReply || snapshot.organizerID != uuid.Nil || snapshot.attendeeVersion != delivery.AttendeeVersion {
		return CalendarMailDeliveryClaim{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	claim := CalendarMailDeliveryClaim{
		deliveryID:      deliveryID,
		replyID:         replyID,
		kind:            CalendarMailDeliveryKindReply,
		method:          CalendarMailDeliveryMethodReply,
		state:           CalendarMailDeliveryStateSubmitting,
		claimToken:      token,
		claimGeneration: generation,
		organizationID:  snapshot.organizationID,
		leaseUntil:      leaseUntil,
		eventID:         snapshot.eventID,
		attendeeID:      snapshot.attendeeID,
		accountID:       snapshot.accountID,
		senderActorID:   snapshot.senderActorID,
		fromIdentityID:  snapshot.fromIdentityID,
		recipient:       snapshot.recipient,
		materialization: CalendarMailDeliveryMaterializationInput{
			deliveryID:      deliveryID,
			kind:            CalendarMailDeliveryKindReply,
			method:          CalendarMailDeliveryMethodReply,
			eventID:         snapshot.eventID,
			attendeeID:      snapshot.attendeeID,
			eventSequence:   snapshot.eventSequence,
			attendeeVersion: snapshot.attendeeVersion,
			message:         snapshot.message,
		},
	}
	if !validCalendarMailDeliveryClaim(claim) {
		return CalendarMailDeliveryClaim{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return calendarMailDeliveryClaimCopy(claim), nil
}

func mailboxEntCASCalendarInviteDeliveryClaim(ctx context.Context, transaction *ent.Tx, delivery *ent.CalendarInviteDelivery, databaseNow time.Time, token CalendarMailDeliveryClaimToken, generation CalendarMailDeliveryClaimGeneration, leaseUntil time.Time, reclaim bool) (int, error) {
	if ctx == nil || transaction == nil || delivery == nil || uuid.UUID(token) == uuid.Nil || generation < 1 || !mailboxEntValidTime(databaseNow) ||
		!validCalendarMailDeliveryLeaseUntil(leaseUntil) || !leaseUntil.After(databaseNow) {
		return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	if reclaim && (delivery.LeaseToken == nil || delivery.LeaseUntil == nil) {
		return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	if !reclaim && delivery.NextAttemptAt == nil {
		return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	update := transaction.CalendarInviteDelivery.Update().Where(
		calendarinvitedelivery.IDEQ(delivery.ID),
		calendarinvitedelivery.OrganizationIDEQ(delivery.OrganizationID),
		calendarinvitedelivery.CalendarEventIDEQ(delivery.CalendarEventID),
		calendarinvitedelivery.CalendarAttendeeIDEQ(delivery.CalendarAttendeeID),
		calendarinvitedelivery.MailSubmissionIDIsNil(),
		calendarinvitedelivery.AttemptsEQ(delivery.Attempts),
		calendarinvitedelivery.ClaimGenerationEQ(delivery.ClaimGeneration),
		calendarinvitedelivery.TerminalAtIsNil(),
	)
	if reclaim {
		update = update.Where(
			calendarinvitedelivery.StateEQ(calendarinvitedelivery.StateSubmitting),
			calendarinvitedelivery.LeaseTokenEQ(*delivery.LeaseToken),
			calendarinvitedelivery.LeaseUntilEQ(*delivery.LeaseUntil),
			calendarinvitedelivery.NextAttemptAtIsNil(),
			calendarinvitedelivery.LastErrorCodeIsNil(),
		)
	} else {
		update = update.Where(
			calendarinvitedelivery.StateEQ(calendarinvitedelivery.StatePending),
			calendarinvitedelivery.LeaseTokenIsNil(),
			calendarinvitedelivery.LeaseUntilIsNil(),
			calendarinvitedelivery.NextAttemptAtEQ(*delivery.NextAttemptAt),
		)
		if delivery.LastErrorCode == nil {
			update = update.Where(calendarinvitedelivery.LastErrorCodeIsNil())
		} else {
			update = update.Where(calendarinvitedelivery.LastErrorCodeEQ(*delivery.LastErrorCode))
		}
	}
	if !reclaim {
		update.SetAttempts(delivery.Attempts + 1)
	}
	updated, err := update.
		SetState(calendarinvitedelivery.StateSubmitting).
		SetClaimGeneration(int64(generation)).
		SetLeaseToken(uuid.UUID(token)).
		SetLeaseUntil(leaseUntil).
		ClearNextAttemptAt().
		ClearLastErrorCode().
		SetUpdatedAt(databaseNow).
		Save(ctx)
	if err != nil {
		return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return updated, nil
}

func mailboxEntCASCalendarReplyDeliveryClaim(ctx context.Context, transaction *ent.Tx, delivery *ent.CalendarReplyDelivery, databaseNow time.Time, token CalendarMailDeliveryClaimToken, generation CalendarMailDeliveryClaimGeneration, leaseUntil time.Time, reclaim bool) (int, error) {
	if ctx == nil || transaction == nil || delivery == nil || uuid.UUID(token) == uuid.Nil || generation < 1 || !mailboxEntValidTime(databaseNow) ||
		!validCalendarMailDeliveryLeaseUntil(leaseUntil) || !leaseUntil.After(databaseNow) {
		return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	if reclaim && (delivery.LeaseToken == nil || delivery.LeaseUntil == nil) {
		return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	if !reclaim && delivery.NextAttemptAt == nil {
		return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	update := transaction.CalendarReplyDelivery.Update().Where(
		calendarreplydelivery.IDEQ(delivery.ID),
		calendarreplydelivery.OrganizationIDEQ(delivery.OrganizationID),
		calendarreplydelivery.CalendarEventIDEQ(delivery.CalendarEventID),
		calendarreplydelivery.CalendarAttendeeIDEQ(delivery.CalendarAttendeeID),
		calendarreplydelivery.MailSubmissionIDIsNil(),
		calendarreplydelivery.AttemptsEQ(delivery.Attempts),
		calendarreplydelivery.ClaimGenerationEQ(delivery.ClaimGeneration),
	)
	if reclaim {
		update = update.Where(
			calendarreplydelivery.StateEQ(calendarreplydelivery.StateSubmitting),
			calendarreplydelivery.LeaseTokenEQ(*delivery.LeaseToken),
			calendarreplydelivery.LeaseUntilEQ(*delivery.LeaseUntil),
			calendarreplydelivery.NextAttemptAtIsNil(),
			calendarreplydelivery.LastErrorCodeIsNil(),
		)
	} else {
		update = update.Where(
			calendarreplydelivery.StateEQ(calendarreplydelivery.StatePending),
			calendarreplydelivery.LeaseTokenIsNil(),
			calendarreplydelivery.LeaseUntilIsNil(),
			calendarreplydelivery.NextAttemptAtEQ(*delivery.NextAttemptAt),
		)
		if delivery.LastErrorCode == nil {
			update = update.Where(calendarreplydelivery.LastErrorCodeIsNil())
		} else {
			update = update.Where(calendarreplydelivery.LastErrorCodeEQ(*delivery.LastErrorCode))
		}
	}
	if !reclaim {
		update.SetAttempts(delivery.Attempts + 1)
	}
	updated, err := update.
		SetState(calendarreplydelivery.StateSubmitting).
		SetClaimGeneration(int64(generation)).
		SetLeaseToken(uuid.UUID(token)).
		SetLeaseUntil(leaseUntil).
		ClearNextAttemptAt().
		ClearLastErrorCode().
		SetUpdatedAt(databaseNow).
		Save(ctx)
	if err != nil {
		return 0, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return updated, nil
}

func mailboxEntCalendarInviteDeliveryPendingValid(delivery *ent.CalendarInviteDelivery) bool {
	if delivery == nil || CalendarMailDeliveryState(delivery.State) != CalendarMailDeliveryStatePending || delivery.Attempts < 0 ||
		delivery.Attempts > calendarMailDeliveryMaximumAttempts || delivery.ClaimGeneration < 0 || delivery.MailSubmissionID != nil ||
		delivery.LeaseToken != nil || delivery.LeaseUntil != nil || delivery.NextAttemptAt == nil || !mailboxEntCalendarDeliveryTimestampValid(*delivery.NextAttemptAt) ||
		delivery.TerminalAt != nil {
		return false
	}
	if delivery.Attempts == 0 {
		return delivery.LastErrorCode == nil
	}
	return delivery.LastErrorCode != nil && mailboxEntCalendarInviteDeliveryRetryCodeValid(string(*delivery.LastErrorCode))
}

func mailboxEntCalendarInviteDeliveryRetryCodeValid(code string) bool {
	switch code {
	case "mailbox_unavailable", "engine_unavailable", "internal_error":
		return true
	default:
		return false
	}
}

func mailboxEntCalendarInviteDeliverySubmittingValid(delivery *ent.CalendarInviteDelivery) bool {
	return delivery != nil && CalendarMailDeliveryState(delivery.State) == CalendarMailDeliveryStateSubmitting &&
		delivery.Attempts >= 1 && delivery.Attempts <= calendarMailDeliveryMaximumAttempts && delivery.ClaimGeneration >= 1 &&
		delivery.MailSubmissionID == nil && delivery.LeaseToken != nil && *delivery.LeaseToken != uuid.Nil &&
		delivery.LeaseUntil != nil && validCalendarMailDeliveryLeaseUntil(*delivery.LeaseUntil) &&
		delivery.NextAttemptAt == nil && delivery.LastErrorCode == nil && delivery.TerminalAt == nil
}

func mailboxEntCalendarInviteDeliveryUnknownBoundValid(delivery *ent.CalendarInviteDelivery) bool {
	return delivery != nil && CalendarMailDeliveryState(delivery.State) == CalendarMailDeliveryStateUnknown &&
		delivery.Attempts >= 1 && delivery.Attempts <= calendarMailDeliveryMaximumAttempts &&
		delivery.MailSubmissionID != nil && *delivery.MailSubmissionID != uuid.Nil && delivery.LeaseToken == nil && delivery.LeaseUntil == nil &&
		delivery.NextAttemptAt != nil && mailboxEntCalendarDeliveryTimestampValid(*delivery.NextAttemptAt) && delivery.LastErrorCode != nil &&
		delivery.TerminalAt == nil
}

func mailboxEntCalendarInviteDeliveryTerminalValid(delivery *ent.CalendarInviteDelivery) bool {
	if delivery == nil || delivery.LeaseToken != nil || delivery.LeaseUntil != nil || delivery.TerminalAt == nil ||
		!mailboxEntCalendarDeliveryTimestampValid(*delivery.TerminalAt) {
		return false
	}
	switch CalendarMailDeliveryState(delivery.State) {
	case CalendarMailDeliveryStateSent:
		return delivery.Attempts >= 1 && delivery.Attempts <= calendarMailDeliveryMaximumAttempts &&
			delivery.MailSubmissionID != nil && *delivery.MailSubmissionID != uuid.Nil && delivery.NextAttemptAt == nil && delivery.LastErrorCode == nil
	case CalendarMailDeliveryStateSuperseded:
		return delivery.MailSubmissionID == nil && delivery.NextAttemptAt == nil && delivery.LastErrorCode == nil
	case CalendarMailDeliveryStateFailedPermanent:
		return delivery.Attempts >= 1 && delivery.Attempts <= calendarMailDeliveryMaximumAttempts &&
			delivery.MailSubmissionID == nil && delivery.NextAttemptAt == nil && delivery.LastErrorCode != nil
	default:
		return false
	}
}
