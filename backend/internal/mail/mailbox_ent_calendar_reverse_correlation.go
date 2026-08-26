package mail

import (
	"bytes"
	"context"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/calendarattendee"
	"github.com/valerii/hub/backend/ent/calendarevent"
	"github.com/valerii/hub/backend/ent/calendarinboundmessagereceipt"
	"github.com/valerii/hub/backend/ent/calendarinvitedelivery"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/mailmessage"
	"github.com/valerii/hub/backend/ent/mailsubmission"
)

var _ CalendarMailboxCorrelationResolver = (*MailboxEntRepository)(nil)

// ResolveMailMessageForCalendarEvent resolves one visible calendar event
// through the only durable Calendar-to-Mail correlation paths. It deliberately
// does not use provider identifiers, folder membership, Internet Message-ID,
// or the legacy denormalized MailMessage calendar_event_id field.
func (repository *MailboxEntRepository) ResolveMailMessageForCalendarEvent(ctx context.Context, request CalendarMailboxCorrelationRequest) (CalendarMailboxCorrelationResult, error) {
	if ctx == nil || !request.valid() {
		return CalendarMailboxCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindInvalid)
	}
	if repository == nil || repository.client == nil {
		return CalendarMailboxCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}
	if err := ctx.Err(); err != nil {
		return CalendarMailboxCorrelationResult{}, err
	}

	var result CalendarMailboxCorrelationResult
	err := repository.withMailboxEntReadSnapshot(ctx, func(transaction *ent.Tx) error {
		resolved, err := mailboxEntResolveMailMessageForCalendarEvent(ctx, transaction, request)
		if err != nil {
			return err
		}
		result = resolved
		return nil
	})
	if err := ctx.Err(); err != nil {
		return CalendarMailboxCorrelationResult{}, err
	}
	if err != nil {
		if kind, ok := MailboxCalendarCorrelationErrorKindOf(err); ok {
			return CalendarMailboxCorrelationResult{}, newMailboxCalendarCorrelationError(kind)
		}
		return CalendarMailboxCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}
	if !result.valid() || !result.Matches(request) {
		return CalendarMailboxCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}
	return result, nil
}

type mailboxEntCalendarReverseCorrelationCandidates struct {
	inbound  []uuid.UUID
	outbound []uuid.UUID
	stale    bool
}

func mailboxEntResolveMailMessageForCalendarEvent(ctx context.Context, transaction *ent.Tx, request CalendarMailboxCorrelationRequest) (CalendarMailboxCorrelationResult, error) {
	if ctx == nil || transaction == nil || !request.valid() {
		return CalendarMailboxCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindInvalid)
	}

	events, err := transaction.CalendarEvent.Query().
		Where(
			calendarevent.IDEQ(request.calendarEventID),
			calendarevent.OrganizationIDEQ(request.organizationID),
			calendarevent.DeletedAtIsNil(),
			calendarevent.Or(
				calendarevent.CalendarOwnerIDEQ(request.actorID),
				calendarevent.OrganizerIDEQ(request.actorID),
				calendarevent.HasAttendeesWith(
					calendarattendee.UserIDEQ(request.actorID),
					calendarattendee.RemovedAtIsNil(),
					calendarattendee.RecurrenceIDIsNil(),
				),
			),
		).
		Order(ent.Asc(calendarevent.FieldID)).
		Limit(mailboxEntCalendarCorrelationCandidateLimit).
		All(ctx)
	if err != nil {
		return CalendarMailboxCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}
	if len(events) == 0 {
		return CalendarMailboxCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindNotFound)
	}
	if len(events) != 1 || !mailboxEntCalendarReverseEventValid(events[0], request) {
		return CalendarMailboxCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}

	accounts, err := transaction.MailAccount.Query().
		Where(
			mailaccount.OrganizationIDEQ(request.organizationID),
			mailaccount.UserIDEQ(request.actorID),
			mailaccount.DesiredEnabledEQ(true),
			mailaccount.StatusEQ(mailaccount.Status("active")),
			mailaccount.ProjectionStateEQ(mailaccount.ProjectionState("ready")),
		).
		Order(ent.Asc(mailaccount.FieldID)).
		Limit(mailboxEntCalendarCorrelationCandidateLimit).
		All(ctx)
	if err != nil {
		return CalendarMailboxCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}
	if len(accounts) == 0 {
		return CalendarMailboxCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindNotFound)
	}
	if len(accounts) != 1 || !mailboxEntCalendarReverseAccountValid(accounts[0], request) {
		return CalendarMailboxCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}
	account := accounts[0]

	candidates := mailboxEntCalendarReverseCorrelationCandidates{}
	if err := mailboxEntCalendarReverseInboundCandidates(ctx, transaction, request, account.ID, &candidates); err != nil {
		return CalendarMailboxCorrelationResult{}, err
	}
	if err := mailboxEntCalendarReverseOutboundCandidates(ctx, transaction, request, account.ID, mailboxEntCalendarReverseEventOwnerOrOrganizer(events[0], request), &candidates); err != nil {
		return CalendarMailboxCorrelationResult{}, err
	}
	return mailboxEntCalendarReverseArbitrate(request, account.ID, candidates)
}

func mailboxEntCalendarReverseEventValid(event *ent.CalendarEvent, request CalendarMailboxCorrelationRequest) bool {
	return event != nil &&
		event.ID == request.calendarEventID &&
		event.OrganizationID == request.organizationID &&
		event.DeletedAt == nil
}

func mailboxEntCalendarReverseEventOwnerOrOrganizer(event *ent.CalendarEvent, request CalendarMailboxCorrelationRequest) bool {
	return event != nil && (event.CalendarOwnerID == request.actorID || event.OrganizerID != nil && *event.OrganizerID == request.actorID)
}

func mailboxEntCalendarReverseAccountValid(account *ent.MailAccount, request CalendarMailboxCorrelationRequest) bool {
	return account != nil &&
		account.ID != uuid.Nil &&
		account.OrganizationID == request.organizationID &&
		account.UserID == request.actorID &&
		account.DesiredEnabled &&
		string(account.Status) == "active" &&
		string(account.ProjectionState) == "ready"
}

func mailboxEntCalendarReverseInboundCandidates(ctx context.Context, transaction *ent.Tx, request CalendarMailboxCorrelationRequest, accountID uuid.UUID, candidates *mailboxEntCalendarReverseCorrelationCandidates) error {
	if ctx == nil || transaction == nil || accountID == uuid.Nil || candidates == nil {
		return newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}

	receipts, err := transaction.CalendarInboundMessageReceipt.Query().
		Where(
			calendarinboundmessagereceipt.OrganizationIDEQ(request.organizationID),
			calendarinboundmessagereceipt.MailboxIDEQ(accountID),
			calendarinboundmessagereceipt.CalendarEventIDEQ(request.calendarEventID),
		).
		Order(ent.Asc(calendarinboundmessagereceipt.FieldID)).
		Limit(mailboxEntCalendarCorrelationCandidateLimit + 1).
		All(ctx)
	if err != nil || len(receipts) > mailboxEntCalendarCorrelationCandidateLimit {
		return newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}

	for _, receipt := range receipts {
		if !mailboxEntCalendarReverseReceiptValid(receipt, request, accountID) {
			return newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
		}
		switch string(receipt.State) {
		case "pending", "processing", "deferred_missing_master":
			candidates.stale = true
		case "applied", "ignored_stale", "duplicate":
			message, err := mailboxEntCalendarReverseLiveMessage(ctx, transaction, accountID, receipt.MailMessageID)
			if err != nil {
				return err
			}
			if message == nil {
				continue
			}
			if !mailboxEntCalendarReverseInboundMessageValid(message, accountID, receipt) {
				return newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
			}
			if string(*message.CalendarMethod) == "cancel" {
				continue
			}
			var appended bool
			candidates.inbound, appended = mailboxEntCalendarCorrelationAppendID(candidates.inbound, message.ID)
			if !appended {
				return newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
			}
		case "quarantined", "dead_letter", "tombstoned":
			// Terminal non-link states never authorize a reverse association.
		default:
			return newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
		}
	}
	return nil
}

func mailboxEntCalendarReverseReceiptValid(receipt *ent.CalendarInboundMessageReceipt, request CalendarMailboxCorrelationRequest, accountID uuid.UUID) bool {
	return receipt != nil &&
		receipt.ID != uuid.Nil &&
		receipt.OrganizationID == request.organizationID &&
		receipt.MailboxID == accountID &&
		receipt.MailMessageID != uuid.Nil &&
		receipt.CalendarEventID != nil &&
		*receipt.CalendarEventID == request.calendarEventID &&
		len(receipt.PartLocatorHash) == 32
}

func mailboxEntCalendarReverseInboundMessageValid(message *ent.MailMessage, accountID uuid.UUID, receipt *ent.CalendarInboundMessageReceipt) bool {
	return message != nil &&
		receipt != nil &&
		message.ID == receipt.MailMessageID &&
		message.MailAccountID == accountID &&
		message.DeletedAt == nil &&
		message.CalendarMethod != nil &&
		mailboxEntCalendarCorrelationMethodValid(*message.CalendarMethod) &&
		len(message.CalendarPartLocatorHash) == 32 &&
		bytes.Equal(message.CalendarPartLocatorHash, receipt.PartLocatorHash)
}

func mailboxEntCalendarReverseOutboundCandidates(ctx context.Context, transaction *ent.Tx, request CalendarMailboxCorrelationRequest, accountID uuid.UUID, outboundEligible bool, candidates *mailboxEntCalendarReverseCorrelationCandidates) error {
	if ctx == nil || transaction == nil || accountID == uuid.Nil || candidates == nil {
		return newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}
	if !outboundEligible {
		return nil
	}

	deliveries, err := transaction.CalendarInviteDelivery.Query().
		Where(
			calendarinvitedelivery.OrganizationIDEQ(request.organizationID),
			calendarinvitedelivery.CalendarEventIDEQ(request.calendarEventID),
		).
		Order(ent.Asc(calendarinvitedelivery.FieldID)).
		Limit(mailboxEntCalendarCorrelationCandidateLimit + 1).
		All(ctx)
	if err != nil || len(deliveries) > mailboxEntCalendarCorrelationCandidateLimit {
		return newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}

	for _, delivery := range deliveries {
		if !mailboxEntCalendarReverseDeliveryValid(delivery, request) {
			return newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
		}
		if err := mailboxEntCalendarReverseDeliveryAttendeeValid(ctx, transaction, delivery); err != nil {
			return err
		}
		switch string(delivery.State) {
		case "pending":
			candidates.stale = true
			continue
		case "superseded", "failed_permanent":
			continue
		case "submitting":
			candidates.stale = true
			if delivery.MailSubmissionID == nil {
				continue
			}
		}

		submission, err := mailboxEntCalendarReverseSubmission(ctx, transaction, accountID, *delivery.MailSubmissionID)
		if err != nil {
			return err
		}
		if !mailboxEntCalendarReverseSubmissionValid(submission, accountID) {
			return newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
		}
		if string(delivery.State) == "submitting" || string(delivery.State) == "unknown" || mailboxEntCalendarCorrelationSubmissionNonterminal(submission.Status) {
			candidates.stale = true
			continue
		}
		if string(submission.Status) != "sent" {
			continue
		}
		if submission.ResultMailMessageID == nil || *submission.ResultMailMessageID == uuid.Nil {
			return newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
		}
		message, err := mailboxEntCalendarReverseLiveMessage(ctx, transaction, accountID, *submission.ResultMailMessageID)
		if err != nil {
			return err
		}
		if message == nil {
			continue
		}
		if !mailboxEntCalendarReverseOutboundMessageValid(message, accountID, *submission.ResultMailMessageID) {
			return newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
		}
		var appended bool
		candidates.outbound, appended = mailboxEntCalendarCorrelationAppendID(candidates.outbound, message.ID)
		if !appended {
			return newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
		}
	}
	return nil
}

func mailboxEntCalendarReverseDeliveryValid(delivery *ent.CalendarInviteDelivery, request CalendarMailboxCorrelationRequest) bool {
	if delivery == nil || delivery.ID == uuid.Nil || delivery.OrganizationID != request.organizationID || delivery.CalendarEventID != request.calendarEventID || delivery.CalendarAttendeeID == uuid.Nil {
		return false
	}
	switch string(delivery.State) {
	case "pending", "superseded":
		return delivery.MailSubmissionID == nil
	case "unknown", "sent":
		return delivery.MailSubmissionID != nil && *delivery.MailSubmissionID != uuid.Nil
	case "submitting", "failed_permanent":
		return delivery.MailSubmissionID == nil || *delivery.MailSubmissionID != uuid.Nil
	default:
		return false
	}
}

func mailboxEntCalendarReverseDeliveryAttendeeValid(ctx context.Context, transaction *ent.Tx, delivery *ent.CalendarInviteDelivery) error {
	if ctx == nil || transaction == nil || delivery == nil || delivery.CalendarEventID == uuid.Nil || delivery.CalendarAttendeeID == uuid.Nil {
		return newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}

	attendee, err := transaction.CalendarAttendee.Query().
		Where(
			calendarattendee.IDEQ(delivery.CalendarAttendeeID),
			calendarattendee.CalendarEventIDEQ(delivery.CalendarEventID),
			calendarattendee.RemovedAtIsNil(),
		).
		Only(ctx)
	if err != nil || attendee == nil || attendee.ID != delivery.CalendarAttendeeID || attendee.CalendarEventID != delivery.CalendarEventID || attendee.RemovedAt != nil || !mailboxEntCalendarCorrelationStringPointerEqual(attendee.RecurrenceID, delivery.RecurrenceID) {
		return newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}
	return nil
}

func mailboxEntCalendarReverseSubmission(ctx context.Context, transaction *ent.Tx, accountID uuid.UUID, submissionID uuid.UUID) (*ent.MailSubmission, error) {
	if ctx == nil || transaction == nil || accountID == uuid.Nil || submissionID == uuid.Nil {
		return nil, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}

	submissions, err := transaction.MailSubmission.Query().
		Where(
			mailsubmission.IDEQ(submissionID),
			mailsubmission.MailAccountIDEQ(accountID),
		).
		Order(ent.Asc(mailsubmission.FieldID)).
		Limit(mailboxEntCalendarCorrelationCandidateLimit).
		All(ctx)
	if err != nil || len(submissions) != 1 || submissions[0] == nil {
		return nil, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}
	return submissions[0], nil
}

func mailboxEntCalendarReverseSubmissionValid(submission *ent.MailSubmission, accountID uuid.UUID) bool {
	if submission == nil || submission.ID == uuid.Nil || submission.MailAccountID != accountID {
		return false
	}
	switch string(submission.Status) {
	case "queued", "authorizing", "submitting", "unknown", "accepted", "sent", "failed":
		return true
	default:
		return false
	}
}

func mailboxEntCalendarReverseDeliveryNonterminal(state calendarinvitedelivery.State) bool {
	switch string(state) {
	case "pending", "submitting", "unknown":
		return true
	default:
		return false
	}
}

func mailboxEntCalendarReverseLiveMessage(ctx context.Context, transaction *ent.Tx, accountID uuid.UUID, messageID uuid.UUID) (*ent.MailMessage, error) {
	if ctx == nil || transaction == nil || accountID == uuid.Nil || messageID == uuid.Nil {
		return nil, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}

	messages, err := transaction.MailMessage.Query().
		Where(
			mailmessage.IDEQ(messageID),
			mailmessage.MailAccountIDEQ(accountID),
			mailmessage.DeletedAtIsNil(),
		).
		Order(ent.Asc(mailmessage.FieldID)).
		Limit(mailboxEntCalendarCorrelationCandidateLimit).
		All(ctx)
	if err != nil {
		return nil, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}
	if len(messages) == 0 {
		return nil, nil
	}
	if len(messages) != 1 || messages[0] == nil || messages[0].ID != messageID || messages[0].MailAccountID != accountID || messages[0].DeletedAt != nil {
		return nil, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}
	return messages[0], nil
}

func mailboxEntCalendarReverseOutboundMessageValid(message *ent.MailMessage, accountID uuid.UUID, messageID uuid.UUID) bool {
	return message != nil &&
		message.ID == messageID &&
		message.MailAccountID == accountID &&
		message.DeletedAt == nil &&
		!message.IsDraft &&
		string(message.ProjectionOrigin) == "confirmed_owner_submission"
}

func mailboxEntCalendarReverseArbitrate(request CalendarMailboxCorrelationRequest, accountID uuid.UUID, candidates mailboxEntCalendarReverseCorrelationCandidates) (CalendarMailboxCorrelationResult, error) {
	if candidates.stale {
		return CalendarMailboxCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindStale)
	}

	messageIDs := mailboxEntCalendarReverseMessageIDs(candidates)
	if len(messageIDs) == 0 {
		return CalendarMailboxCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindNotFound)
	}
	if len(messageIDs) > 1 {
		return CalendarMailboxCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindAmbiguous)
	}

	path := MailboxCalendarCorrelationPathOutboundInvite
	if mailboxEntCalendarCorrelationContains(candidates.inbound, messageIDs[0]) {
		path = MailboxCalendarCorrelationPathInboundReceipt
	}
	result, err := NewCalendarMailboxCorrelationResult(request, accountID, messageIDs[0], path)
	if err != nil {
		return CalendarMailboxCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}
	return result, nil
}

func mailboxEntCalendarReverseMessageIDs(candidates mailboxEntCalendarReverseCorrelationCandidates) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(candidates.inbound)+len(candidates.outbound))
	for _, candidate := range candidates.inbound {
		if !mailboxEntCalendarCorrelationContains(result, candidate) {
			result = append(result, candidate)
		}
	}
	for _, candidate := range candidates.outbound {
		if !mailboxEntCalendarCorrelationContains(result, candidate) {
			result = append(result, candidate)
		}
	}
	return result
}
