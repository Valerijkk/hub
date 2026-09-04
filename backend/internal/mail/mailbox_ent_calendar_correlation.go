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

const (
	mailboxEntCalendarCorrelationCandidateLimit = 2
	mailboxEntCalendarCorrelationEventLimit     = mailboxEntCalendarCorrelationCandidateLimit * 2
)

var _ MailboxCalendarCorrelationResolver = (*MailboxEntRepository)(nil)

// ResolveCalendarEventForMessage resolves one live mailbox message through
// the only durable Mail-to-Calendar correlation paths. It deliberately does
// not use provider identifiers, folder membership, or the legacy denormalized
// MailMessage calendar_event_id field.
func (repository *MailboxEntRepository) ResolveCalendarEventForMessage(ctx context.Context, request MailboxCalendarCorrelationRequest) (MailboxCalendarCorrelationResult, error) {
	if ctx == nil || !request.valid() {
		return MailboxCalendarCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindInvalid)
	}
	if repository == nil || repository.client == nil {
		return MailboxCalendarCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}
	if err := ctx.Err(); err != nil {
		return MailboxCalendarCorrelationResult{}, err
	}

	var result MailboxCalendarCorrelationResult
	err := repository.withMailboxEntReadSnapshot(ctx, func(transaction *ent.Tx) error {
		resolved, err := mailboxEntResolveCalendarEventForMessage(ctx, transaction, request)
		if err != nil {
			return err
		}
		result = resolved
		return nil
	})
	if err := ctx.Err(); err != nil {
		return MailboxCalendarCorrelationResult{}, err
	}
	if err != nil {
		if kind, ok := MailboxCalendarCorrelationErrorKindOf(err); ok {
			return MailboxCalendarCorrelationResult{}, newMailboxCalendarCorrelationError(kind)
		}
		return MailboxCalendarCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}
	if !result.valid() || !result.Matches(request) {
		return MailboxCalendarCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}
	return result, nil
}

type mailboxEntCalendarCorrelationCandidates struct {
	inbound  []uuid.UUID
	outbound []uuid.UUID
	stale    bool
}

func mailboxEntResolveCalendarEventForMessage(ctx context.Context, transaction *ent.Tx, request MailboxCalendarCorrelationRequest) (MailboxCalendarCorrelationResult, error) {
	if ctx == nil || transaction == nil || !request.valid() {
		return MailboxCalendarCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindInvalid)
	}

	accounts, err := transaction.MailAccount.Query().
		Where(
			mailaccount.IDEQ(request.accountID),
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
		return MailboxCalendarCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}
	if len(accounts) == 0 {
		return MailboxCalendarCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindNotFound)
	}
	if len(accounts) != 1 || !mailboxEntCalendarCorrelationAccountValid(accounts[0], request) {
		return MailboxCalendarCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}

	messages, err := transaction.MailMessage.Query().
		Where(
			mailmessage.IDEQ(request.messageID),
			mailmessage.MailAccountIDEQ(request.accountID),
			mailmessage.DeletedAtIsNil(),
		).
		Order(ent.Asc(mailmessage.FieldID)).
		Limit(mailboxEntCalendarCorrelationCandidateLimit).
		All(ctx)
	if err != nil {
		return MailboxCalendarCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}
	if len(messages) == 0 {
		return MailboxCalendarCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindNotFound)
	}
	if len(messages) != 1 || !mailboxEntCalendarCorrelationMessageValid(messages[0], request) {
		return MailboxCalendarCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}
	message := messages[0]
	calendarPartLocatorHash := mailboxEntCalendarPartLocatorHashValue(message.CalendarPartLocatorHash)

	candidates := mailboxEntCalendarCorrelationCandidates{}
	if message.CalendarMethod != nil {
		if calendarPartLocatorHash == nil || len(calendarPartLocatorHash) != 32 || !mailboxEntCalendarCorrelationMethodValid(*message.CalendarMethod) {
			return MailboxCalendarCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
		}
		if err := mailboxEntCalendarCorrelationInboundCandidates(ctx, transaction, request, message, calendarPartLocatorHash, &candidates); err != nil {
			return MailboxCalendarCorrelationResult{}, err
		}
	} else if calendarPartLocatorHash != nil {
		return MailboxCalendarCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}

	if !message.IsDraft && string(message.ProjectionOrigin) == "confirmed_owner_submission" {
		if err := mailboxEntCalendarCorrelationOutboundCandidates(ctx, transaction, request, message, &candidates); err != nil {
			return MailboxCalendarCorrelationResult{}, err
		}
	}

	return mailboxEntCalendarCorrelationArbitrate(ctx, transaction, request, candidates)
}

func mailboxEntCalendarCorrelationAccountValid(account *ent.MailAccount, request MailboxCalendarCorrelationRequest) bool {
	return account != nil &&
		account.ID == request.accountID &&
		account.OrganizationID == request.organizationID &&
		account.UserID == request.actorID &&
		account.DesiredEnabled &&
		string(account.Status) == "active" &&
		string(account.ProjectionState) == "ready"
}

func mailboxEntCalendarCorrelationMessageValid(message *ent.MailMessage, request MailboxCalendarCorrelationRequest) bool {
	if message == nil || message.ID != request.messageID || message.MailAccountID != request.accountID || message.DeletedAt != nil {
		return false
	}
	switch string(message.ProjectionOrigin) {
	case "inbound_or_unknown", "confirmed_owner_submission", "confirmed_owner_draft":
		return true
	default:
		return false
	}
}

func mailboxEntCalendarCorrelationMethodValid(method mailmessage.CalendarMethod) bool {
	switch string(method) {
	case "request", "cancel", "reply":
		return true
	default:
		return false
	}
}

func mailboxEntCalendarCorrelationInboundCandidates(ctx context.Context, transaction *ent.Tx, request MailboxCalendarCorrelationRequest, message *ent.MailMessage, calendarPartLocatorHash []byte, candidates *mailboxEntCalendarCorrelationCandidates) error {
	if ctx == nil || transaction == nil || message == nil || candidates == nil || message.CalendarMethod == nil || calendarPartLocatorHash == nil || len(calendarPartLocatorHash) != 32 {
		return newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}

	rows, err := transaction.CalendarInboundMessageReceipt.Query().
		Where(
			calendarinboundmessagereceipt.OrganizationIDEQ(request.organizationID),
			calendarinboundmessagereceipt.MailboxIDEQ(request.accountID),
			calendarinboundmessagereceipt.MailMessageIDEQ(request.messageID),
			calendarinboundmessagereceipt.PartLocatorHashEQ(calendarPartLocatorHash),
		).
		Order(ent.Asc(calendarinboundmessagereceipt.FieldID)).
		Limit(mailboxEntCalendarCorrelationCandidateLimit + 1).
		All(ctx)
	if err != nil || len(rows) > mailboxEntCalendarCorrelationCandidateLimit {
		return newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}

	for _, row := range rows {
		if !mailboxEntCalendarCorrelationInboundRowValid(row, request, message, calendarPartLocatorHash) {
			return newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
		}

		switch string(row.State) {
		case "pending", "processing", "deferred_missing_master":
			candidates.stale = true
		case "applied", "ignored_stale", "duplicate":
			if row.CalendarEventID == nil {
				continue
			}
			var appended bool
			candidates.inbound, appended = mailboxEntCalendarCorrelationAppendID(candidates.inbound, *row.CalendarEventID)
			if !appended {
				return newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
			}
		case "quarantined", "dead_letter":
			// Terminal quarantine and dead-letter receipts never authorize a link.
		default:
			return newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
		}
	}
	return nil
}

func mailboxEntCalendarCorrelationInboundRowValid(row *ent.CalendarInboundMessageReceipt, request MailboxCalendarCorrelationRequest, message *ent.MailMessage, calendarPartLocatorHash []byte) bool {
	if row == nil || message == nil || calendarPartLocatorHash == nil || len(calendarPartLocatorHash) != 32 || row.ID == uuid.Nil || row.OrganizationID != request.organizationID || row.MailboxID != request.accountID || row.MailMessageID != request.messageID || !bytes.Equal(row.PartLocatorHash, calendarPartLocatorHash) {
		return false
	}
	return row.CalendarEventID == nil || *row.CalendarEventID != uuid.Nil
}

func mailboxEntCalendarCorrelationOutboundCandidates(ctx context.Context, transaction *ent.Tx, request MailboxCalendarCorrelationRequest, message *ent.MailMessage, candidates *mailboxEntCalendarCorrelationCandidates) error {
	if ctx == nil || transaction == nil || message == nil || candidates == nil || message.IsDraft || string(message.ProjectionOrigin) != "confirmed_owner_submission" {
		return newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}

	submissions, err := transaction.MailSubmission.Query().
		Where(
			mailsubmission.MailAccountIDEQ(request.accountID),
			mailsubmission.ResultMailMessageIDEQ(request.messageID),
		).
		Order(ent.Asc(mailsubmission.FieldID)).
		Limit(mailboxEntCalendarCorrelationCandidateLimit + 1).
		All(ctx)
	if err != nil || len(submissions) > mailboxEntCalendarCorrelationCandidateLimit {
		return newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}

	for _, submission := range submissions {
		if !mailboxEntCalendarCorrelationSubmissionValid(submission, request) {
			return newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
		}

		if mailboxEntCalendarCorrelationSubmissionNonterminal(submission.Status) {
			candidates.stale = true
		} else if string(submission.Status) != "sent" && string(submission.Status) != "failed" {
			return newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
		}

		deliveries, err := transaction.CalendarInviteDelivery.Query().
			Where(
				calendarinvitedelivery.OrganizationIDEQ(request.organizationID),
				calendarinvitedelivery.MailSubmissionIDEQ(submission.ID),
			).
			Order(ent.Asc(calendarinvitedelivery.FieldID)).
			Limit(mailboxEntCalendarCorrelationCandidateLimit + 1).
			All(ctx)
		if err != nil || len(deliveries) > mailboxEntCalendarCorrelationCandidateLimit {
			return newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
		}

		for _, delivery := range deliveries {
			if !mailboxEntCalendarCorrelationDeliveryValid(delivery, request, submission) {
				return newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
			}

			switch string(delivery.State) {
			case "pending", "submitting", "unknown":
				candidates.stale = true
			case "sent":
				if string(submission.Status) != "sent" {
					continue
				}
				if err := mailboxEntCalendarCorrelationAppendOutboundEvent(ctx, transaction, delivery, candidates); err != nil {
					return err
				}
			case "superseded", "failed_permanent":
				// Terminal delivery states do not authorize a link.
			default:
				return newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
			}
		}
	}
	return nil
}

func mailboxEntCalendarCorrelationSubmissionValid(submission *ent.MailSubmission, request MailboxCalendarCorrelationRequest) bool {
	if submission == nil || submission.ID == uuid.Nil || submission.MailAccountID != request.accountID || submission.ResultMailMessageID == nil || *submission.ResultMailMessageID != request.messageID {
		return false
	}
	switch string(submission.Status) {
	case "queued", "authorizing", "submitting", "unknown", "accepted", "sent", "failed":
		return true
	default:
		return false
	}
}

func mailboxEntCalendarCorrelationSubmissionNonterminal(status mailsubmission.Status) bool {
	switch string(status) {
	case "queued", "authorizing", "submitting", "unknown", "accepted":
		return true
	default:
		return false
	}
}

func mailboxEntCalendarCorrelationDeliveryValid(delivery *ent.CalendarInviteDelivery, request MailboxCalendarCorrelationRequest, submission *ent.MailSubmission) bool {
	if delivery == nil || submission == nil || delivery.ID == uuid.Nil || delivery.OrganizationID != request.organizationID || delivery.MailSubmissionID == nil || *delivery.MailSubmissionID != submission.ID || delivery.CalendarEventID == uuid.Nil || delivery.CalendarAttendeeID == uuid.Nil {
		return false
	}
	switch string(delivery.State) {
	case "pending", "submitting", "unknown", "sent", "superseded", "failed_permanent":
		return true
	default:
		return false
	}
}

func mailboxEntCalendarCorrelationAppendOutboundEvent(ctx context.Context, transaction *ent.Tx, delivery *ent.CalendarInviteDelivery, candidates *mailboxEntCalendarCorrelationCandidates) error {
	if ctx == nil || transaction == nil || delivery == nil || candidates == nil || delivery.CalendarEventID == uuid.Nil || delivery.CalendarAttendeeID == uuid.Nil {
		return newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}
	if err := mailboxEntCalendarCorrelationDeliveryAttendeeValid(ctx, transaction, delivery); err != nil {
		return err
	}

	var appended bool
	candidates.outbound, appended = mailboxEntCalendarCorrelationAppendID(candidates.outbound, delivery.CalendarEventID)
	if !appended {
		return newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}
	return nil
}

func mailboxEntCalendarCorrelationStringPointerEqual(left, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func mailboxEntCalendarCorrelationAppendID(ids []uuid.UUID, candidate uuid.UUID) ([]uuid.UUID, bool) {
	if candidate == uuid.Nil {
		return nil, false
	}
	for _, value := range ids {
		if value == candidate {
			return ids, true
		}
	}
	if len(ids) >= mailboxEntCalendarCorrelationCandidateLimit {
		return nil, false
	}
	return append(ids, candidate), true
}

func mailboxEntCalendarCorrelationArbitrate(ctx context.Context, transaction *ent.Tx, request MailboxCalendarCorrelationRequest, candidates mailboxEntCalendarCorrelationCandidates) (MailboxCalendarCorrelationResult, error) {
	if candidates.stale {
		return MailboxCalendarCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindStale)
	}

	eventIDs := mailboxEntCalendarCorrelationEventIDs(candidates)
	if len(eventIDs) == 0 {
		return MailboxCalendarCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindNotFound)
	}
	if len(eventIDs) > 1 {
		return MailboxCalendarCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindAmbiguous)
	}
	if len(eventIDs) > mailboxEntCalendarCorrelationEventLimit {
		return MailboxCalendarCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}

	events, err := transaction.CalendarEvent.Query().
		Where(
			calendarevent.IDIn(eventIDs...),
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
		Limit(len(eventIDs) + 1).
		All(ctx)
	if err != nil || len(events) > len(eventIDs) {
		return MailboxCalendarCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}

	visible := make([]uuid.UUID, 0, len(events))
	for _, event := range events {
		if event == nil || event.ID == uuid.Nil || event.OrganizationID != request.organizationID || event.DeletedAt != nil || !mailboxEntCalendarCorrelationContains(eventIDs, event.ID) || mailboxEntCalendarCorrelationContains(visible, event.ID) {
			return MailboxCalendarCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
		}
		visible = append(visible, event.ID)
	}
	if len(visible) == 0 {
		// A missing, deleted, out-of-organization, or invisible event is never
		// distinguished to the caller.
		return MailboxCalendarCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindNotFound)
	}
	if len(visible) != 1 {
		return MailboxCalendarCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindAmbiguous)
	}

	path := MailboxCalendarCorrelationPathOutboundInvite
	if mailboxEntCalendarCorrelationContains(candidates.inbound, visible[0]) {
		path = MailboxCalendarCorrelationPathInboundReceipt
	}
	result, err := NewMailboxCalendarCorrelationResult(request, visible[0], path)
	if err != nil {
		return MailboxCalendarCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}
	return result, nil
}

func mailboxEntCalendarCorrelationEventIDs(candidates mailboxEntCalendarCorrelationCandidates) []uuid.UUID {
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

func mailboxEntCalendarCorrelationContains(values []uuid.UUID, target uuid.UUID) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
