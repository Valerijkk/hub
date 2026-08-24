package mail

import (
	"context"
	"crypto/sha256"
	"time"

	"github.com/google/uuid"
)

type mailboxForwardDraftPreflightKind string

const (
	mailboxForwardDraftPreflightAbsent  mailboxForwardDraftPreflightKind = "absent"
	mailboxForwardDraftPreflightActive  mailboxForwardDraftPreflightKind = "active"
	mailboxForwardDraftPreflightBusy    mailboxForwardDraftPreflightKind = "busy"
	mailboxForwardDraftPreflightUnknown mailboxForwardDraftPreflightKind = "unknown"
)

type mailboxForwardDraftPreflightInput struct {
	DraftClientID   uuid.UUID
	FromIdentityID  uuid.UUID
	SourceMessageID uuid.UUID
	IdempotencyKey  string
}

type mailboxForwardDraftPreflightResult struct {
	Kind      mailboxForwardDraftPreflightKind
	DraftID   uuid.UUID
	BusyUntil *time.Time
}

// preflightMailboxForwardDraft resolves an idempotent forward replay before
// the source message or provider are read. It deliberately proves only the
// content-free durable correlation; CreateMailboxForwardDraft derives source
// content only after an absent result.
func (service *MailboxDraftService) preflightMailboxForwardDraft(ctx context.Context, actorID uuid.UUID, input mailboxForwardDraftPreflightInput) (mailboxForwardDraftPreflightResult, error) {
	if service == nil ||
		mailboxDraftServiceDependencyNil(service.owners) ||
		mailboxDraftServiceDependencyNil(service.repository) ||
		ctx == nil || actorID == uuid.Nil || input.DraftClientID == uuid.Nil ||
		input.FromIdentityID == uuid.Nil || input.SourceMessageID == uuid.Nil {
		return mailboxForwardDraftPreflightResult{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	if ctx.Err() != nil {
		return mailboxForwardDraftPreflightResult{}, mailboxDraftServiceError(MailboxDraftServiceErrorUnavailable)
	}

	idempotencyKey, err := mailboxDraftCreateNormalizeIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return mailboxForwardDraftPreflightResult{}, err
	}
	owner, err := service.owners.ResolveOwner(ctx, actorID)
	if err != nil {
		return mailboxForwardDraftPreflightResult{}, mailboxDraftGetOwnerError(err)
	}
	scope, err := mailboxDraftCreateScopeFromOwner(actorID, owner)
	if err != nil {
		return mailboxForwardDraftPreflightResult{}, err
	}
	hash := sha256.Sum256([]byte(idempotencyKey))
	locator, err := service.repository.LookupMailboxDraftCreatePreHashLocator(ctx, MailboxDraftCreatePreHashLocatorRequest{
		AccountID: scope.accountID, DraftClientID: input.DraftClientID, CreateIdempotencyKeyHash: hash,
	})
	if err != nil {
		return mailboxForwardDraftPreflightResult{}, mailboxDraftCreateRepositoryError(err)
	}
	if !mailboxForwardDraftPreflightLocatorValid(locator) {
		return mailboxForwardDraftPreflightResult{}, mailboxDraftServiceError(MailboxDraftServiceErrorOutcomeUnknown)
	}
	if !locator.Found {
		return mailboxForwardDraftPreflightResult{Kind: mailboxForwardDraftPreflightAbsent}, nil
	}
	if locator.FromIdentityID != input.FromIdentityID || locator.ReplyToMessageID != nil || locator.ForwardMessageID == nil || *locator.ForwardMessageID != input.SourceMessageID {
		return mailboxForwardDraftPreflightResult{}, mailboxDraftServiceError(MailboxDraftServiceErrorIdempotencyConflict)
	}

	switch locator.Status {
	case MailboxDraftStatusActive:
		return mailboxForwardDraftPreflightResult{Kind: mailboxForwardDraftPreflightActive, DraftID: locator.DraftID}, nil
	case MailboxDraftStatusSaving:
		if locator.BusyUntil == nil {
			return mailboxForwardDraftPreflightResult{Kind: mailboxForwardDraftPreflightUnknown, DraftID: locator.DraftID}, nil
		}
		busyUntil := *locator.BusyUntil
		return mailboxForwardDraftPreflightResult{Kind: mailboxForwardDraftPreflightBusy, DraftID: locator.DraftID, BusyUntil: &busyUntil}, nil
	case MailboxDraftStatusUnknown:
		return mailboxForwardDraftPreflightResult{Kind: mailboxForwardDraftPreflightUnknown, DraftID: locator.DraftID}, nil
	default:
		return mailboxForwardDraftPreflightResult{}, mailboxDraftServiceError(MailboxDraftServiceErrorOutcomeUnknown)
	}
}

func mailboxForwardDraftPreflightLocatorValid(locator MailboxDraftCreatePreHashLocatorResult) bool {
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

func mailboxForwardDraftPreflightResultValid(result mailboxForwardDraftPreflightResult) bool {
	switch result.Kind {
	case mailboxForwardDraftPreflightAbsent:
		return result.DraftID == uuid.Nil && result.BusyUntil == nil
	case mailboxForwardDraftPreflightActive, mailboxForwardDraftPreflightUnknown:
		return result.DraftID != uuid.Nil && result.BusyUntil == nil
	case mailboxForwardDraftPreflightBusy:
		return result.DraftID != uuid.Nil && result.BusyUntil != nil && mailboxEntValidTime(*result.BusyUntil)
	default:
		return false
	}
}
