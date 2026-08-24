package mail

import (
	"context"
	"crypto/sha256"
	"time"

	"github.com/google/uuid"
)

type mailboxReplyDraftPreflightKind string

const (
	mailboxReplyDraftPreflightAbsent  mailboxReplyDraftPreflightKind = "absent"
	mailboxReplyDraftPreflightActive  mailboxReplyDraftPreflightKind = "active"
	mailboxReplyDraftPreflightBusy    mailboxReplyDraftPreflightKind = "busy"
	mailboxReplyDraftPreflightUnknown mailboxReplyDraftPreflightKind = "unknown"
)

type mailboxReplyDraftPreflightInput struct {
	DraftClientID   uuid.UUID
	FromIdentityID  uuid.UUID
	SourceMessageID uuid.UUID
	IdempotencyKey  string
}

type mailboxReplyDraftPreflightResult struct {
	Kind      mailboxReplyDraftPreflightKind
	DraftID   uuid.UUID
	BusyUntil *time.Time
}

func (service *MailboxDraftService) preflightMailboxReplyDraft(ctx context.Context, actorID uuid.UUID, input mailboxReplyDraftPreflightInput) (mailboxReplyDraftPreflightResult, error) {
	if service == nil ||
		mailboxDraftServiceDependencyNil(service.owners) ||
		mailboxDraftServiceDependencyNil(service.repository) ||
		ctx == nil || actorID == uuid.Nil || input.DraftClientID == uuid.Nil ||
		input.FromIdentityID == uuid.Nil || input.SourceMessageID == uuid.Nil {
		return mailboxReplyDraftPreflightResult{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	if ctx.Err() != nil {
		return mailboxReplyDraftPreflightResult{}, mailboxDraftServiceError(MailboxDraftServiceErrorUnavailable)
	}

	idempotencyKey, err := mailboxDraftCreateNormalizeIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return mailboxReplyDraftPreflightResult{}, err
	}
	owner, err := service.owners.ResolveOwner(ctx, actorID)
	if err != nil {
		return mailboxReplyDraftPreflightResult{}, mailboxDraftGetOwnerError(err)
	}
	scope, err := mailboxDraftCreateScopeFromOwner(actorID, owner)
	if err != nil {
		return mailboxReplyDraftPreflightResult{}, err
	}
	hash := sha256.Sum256([]byte(idempotencyKey))
	locator, err := service.repository.LookupMailboxDraftCreatePreHashLocator(ctx, MailboxDraftCreatePreHashLocatorRequest{
		AccountID: scope.accountID, DraftClientID: input.DraftClientID, CreateIdempotencyKeyHash: hash,
	})
	if err != nil {
		return mailboxReplyDraftPreflightResult{}, mailboxDraftCreateRepositoryError(err)
	}
	if !mailboxReplyDraftPreflightLocatorValid(locator) {
		return mailboxReplyDraftPreflightResult{}, mailboxDraftServiceError(MailboxDraftServiceErrorOutcomeUnknown)
	}
	if !locator.Found {
		return mailboxReplyDraftPreflightResult{Kind: mailboxReplyDraftPreflightAbsent}, nil
	}
	if locator.FromIdentityID != input.FromIdentityID || locator.ReplyToMessageID == nil || *locator.ReplyToMessageID != input.SourceMessageID || locator.ForwardMessageID != nil {
		return mailboxReplyDraftPreflightResult{}, mailboxDraftServiceError(MailboxDraftServiceErrorIdempotencyConflict)
	}

	switch locator.Status {
	case MailboxDraftStatusActive:
		return mailboxReplyDraftPreflightResult{Kind: mailboxReplyDraftPreflightActive, DraftID: locator.DraftID}, nil
	case MailboxDraftStatusSaving:
		if locator.BusyUntil == nil {
			return mailboxReplyDraftPreflightResult{Kind: mailboxReplyDraftPreflightUnknown, DraftID: locator.DraftID}, nil
		}
		busyUntil := *locator.BusyUntil
		return mailboxReplyDraftPreflightResult{Kind: mailboxReplyDraftPreflightBusy, DraftID: locator.DraftID, BusyUntil: &busyUntil}, nil
	case MailboxDraftStatusUnknown:
		return mailboxReplyDraftPreflightResult{Kind: mailboxReplyDraftPreflightUnknown, DraftID: locator.DraftID}, nil
	default:
		return mailboxReplyDraftPreflightResult{}, mailboxDraftServiceError(MailboxDraftServiceErrorOutcomeUnknown)
	}
}

func mailboxReplyDraftPreflightLocatorValid(locator MailboxDraftCreatePreHashLocatorResult) bool {
	if !locator.Found {
		return locator.DraftID == uuid.Nil && locator.Status == "" && locator.FromIdentityID == uuid.Nil &&
			locator.ReplyToMessageID == nil && locator.ForwardMessageID == nil && locator.BusyUntil == nil
	}
	if locator.DraftID == uuid.Nil || locator.FromIdentityID == uuid.Nil || !mailboxDraftGetUUIDPointerValid(locator.ReplyToMessageID) || !mailboxDraftGetUUIDPointerValid(locator.ForwardMessageID) {
		return false
	}
	switch locator.Status {
	case MailboxDraftStatusActive, MailboxDraftStatusUnknown:
		return locator.BusyUntil == nil
	case MailboxDraftStatusSaving:
		return locator.BusyUntil == nil || mailboxEntValidTime(*locator.BusyUntil)
	default:
		return false
	}
}

func mailboxReplyDraftPreflightResultValid(result mailboxReplyDraftPreflightResult) bool {
	switch result.Kind {
	case mailboxReplyDraftPreflightAbsent:
		return result.DraftID == uuid.Nil && result.BusyUntil == nil
	case mailboxReplyDraftPreflightActive, mailboxReplyDraftPreflightUnknown:
		return result.DraftID != uuid.Nil && result.BusyUntil == nil
	case mailboxReplyDraftPreflightBusy:
		return result.DraftID != uuid.Nil && result.BusyUntil != nil && mailboxEntValidTime(*result.BusyUntil)
	default:
		return false
	}
}
