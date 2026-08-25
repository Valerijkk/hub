package mail

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// MailboxMessageReadEngine reads one bounded message from an account-scoped engine projection.
type MailboxMessageReadEngine interface {
	GetMessage(context.Context, EngineMessageReadRequest) (EngineMessage, error)
	GetBlob(context.Context, EngineBlobReadRequest) (EngineBlob, error)
}

var _ MailboxMessageReadEngine = (*StalwartEngine)(nil)

// MailboxMessageService owner-gates one bounded engine body read against a sealed local projection.
type MailboxMessageService struct {
	owners     mailboxOwnerResolver
	repository MailboxMessageRepository
	engine     MailboxMessageReadEngine
}

// NewMailboxMessageService creates the safe public message-detail service.
func NewMailboxMessageService(owners mailboxOwnerResolver, repository MailboxMessageRepository, engine MailboxMessageReadEngine) (*MailboxMessageService, error) {
	if mailboxFolderProjectionDependencyNil(owners) || mailboxFolderProjectionDependencyNil(repository) || mailboxFolderProjectionDependencyNil(engine) {
		return nil, mailboxMessageReadError(MailboxMessageReadErrorInvalid)
	}
	return &MailboxMessageService{owners: owners, repository: repository, engine: engine}, nil
}

// GetMailboxMessage returns local safe metadata plus one normalized, owner-bound engine body.
func (service *MailboxMessageService) GetMailboxMessage(ctx context.Context, actorID, messageID uuid.UUID) (MailboxMessageDetail, error) {
	if service == nil || mailboxFolderProjectionDependencyNil(service.owners) || mailboxFolderProjectionDependencyNil(service.repository) || mailboxFolderProjectionDependencyNil(service.engine) || ctx == nil || actorID == uuid.Nil || messageID == uuid.Nil {
		return MailboxMessageDetail{}, mailboxMessageReadError(MailboxMessageReadErrorInvalid)
	}
	if ctx.Err() != nil {
		return MailboxMessageDetail{}, mailboxMessageReadError(MailboxMessageReadErrorUnavailable)
	}

	owner, err := service.owners.ResolveOwner(ctx, actorID)
	if err != nil {
		return MailboxMessageDetail{}, mailboxMessageOwnerError(err)
	}
	scope, err := mailboxMessageReadScopeFromOwner(actorID, owner)
	if err != nil {
		return MailboxMessageDetail{}, err
	}

	snapshot, err := service.repository.ResolveMailboxMessage(ctx, scope, messageID)
	if err != nil {
		return MailboxMessageDetail{}, mailboxMessageRepositoryError(err)
	}
	if !mailboxMessageSnapshotValid(snapshot) || snapshot.summary.ID != messageID || snapshot.engineAccountID != scope.EngineAccountID {
		return MailboxMessageDetail{}, mailboxMessageReadError(MailboxMessageReadErrorInvalid)
	}

	engineMessage, err := service.engine.GetMessage(ctx, EngineMessageReadRequest{
		AccountID: scope.EngineAccountID, MessageID: snapshot.engineMessageID, MaxContentBytes: mailboxMessageEngineReadMaxBytes,
	})
	if err != nil {
		return MailboxMessageDetail{}, mailboxMessageEngineError(err)
	}
	if !mailboxMessageProjectionMatches(snapshot, engineMessage.Projection) {
		return MailboxMessageDetail{}, mailboxMessageReadError(MailboxMessageReadErrorUnavailable)
	}

	content, err := mailboxMessageNormalizeEngineBody(engineMessage.Body)
	if err != nil {
		return MailboxMessageDetail{}, err
	}
	if err := service.repository.RecheckMailboxMessage(ctx, scope, snapshot); err != nil {
		return MailboxMessageDetail{}, mailboxMessageRepositoryError(err)
	}

	detail, err := mailboxMessageDetailFromSnapshot(snapshot, content)
	if err != nil {
		return MailboxMessageDetail{}, err
	}
	return mailboxMessageDetailCopy(detail), nil
}

func mailboxMessageReadScopeFromOwner(actorID uuid.UUID, owner MailboxOwner) (MailboxMessageReadScope, error) {
	if actorID == uuid.Nil || owner.OrganizationID() == uuid.Nil {
		return MailboxMessageReadScope{}, mailboxMessageReadError(MailboxMessageReadErrorInvalid)
	}
	accountID, hasAccount := owner.AccountID()
	projectionState, hasProjectionState := owner.ProjectionState()
	switch owner.State() {
	case MailboxOwnerNotConfigured:
		if hasAccount || accountID != uuid.Nil || owner.hasEngineAccountID || hasProjectionState {
			return MailboxMessageReadScope{}, mailboxMessageReadError(MailboxMessageReadErrorInvalid)
		}
		return MailboxMessageReadScope{}, mailboxMessageReadError(MailboxMessageReadErrorNotFound)
	case MailboxOwnerDeleting:
		if !hasAccount || accountID == uuid.Nil || owner.hasEngineAccountID || !hasProjectionState || !validMailboxProjectionState(projectionState) {
			return MailboxMessageReadScope{}, mailboxMessageReadError(MailboxMessageReadErrorInvalid)
		}
		return MailboxMessageReadScope{}, mailboxMessageReadError(MailboxMessageReadErrorDeleting)
	case MailboxOwnerProvisioning, MailboxOwnerSuspended, MailboxOwnerError:
		if !hasAccount || accountID == uuid.Nil || !hasProjectionState || !validMailboxProjectionState(projectionState) {
			return MailboxMessageReadScope{}, mailboxMessageReadError(MailboxMessageReadErrorInvalid)
		}
		return MailboxMessageReadScope{}, mailboxMessageReadError(MailboxMessageReadErrorUnavailable)
	case MailboxOwnerActive:
		if !hasAccount || accountID == uuid.Nil || !owner.hasEngineAccountID || !syncValidOpaque(owner.engineAccountID, MaxEngineOpaqueIDBytes) || !hasProjectionState || !validMailboxProjectionState(projectionState) {
			return MailboxMessageReadScope{}, mailboxMessageReadError(MailboxMessageReadErrorInvalid)
		}
		if projectionState != MailboxProjectionReady {
			return MailboxMessageReadScope{}, mailboxMessageReadError(MailboxMessageReadErrorUnavailable)
		}
		return MailboxMessageReadScope{
			ActorID: actorID, OrganizationID: owner.OrganizationID(), AccountID: accountID, EngineAccountID: owner.engineAccountID,
		}, nil
	default:
		return MailboxMessageReadScope{}, mailboxMessageReadError(MailboxMessageReadErrorInvalid)
	}
}

func mailboxMessageNormalizeEngineBody(body EngineMessageBody) (MailboxContentResult, error) {
	var input MailboxContentInput
	switch body.Format {
	case EngineBodyPlain:
		if body.RichBody != "" {
			return MailboxContentResult{}, mailboxMessageReadError(MailboxMessageReadErrorUnprocessable)
		}
		input = MailboxContentInput{SourceFormat: MailboxContentFormatPlain, Body: body.PlainBody}
	case EngineBodyRich:
		input = MailboxContentInput{SourceFormat: MailboxContentFormatRich, Body: body.RichBody}
	default:
		return MailboxContentResult{}, mailboxMessageReadError(MailboxMessageReadErrorUnprocessable)
	}
	content, err := NormalizeMailboxContent(input)
	if err != nil {
		switch {
		case errors.Is(err, ErrMailboxContentTooLarge):
			return MailboxContentResult{}, mailboxMessageReadError(MailboxMessageReadErrorPayloadTooLarge)
		case errors.Is(err, ErrMailboxContentInvalid):
			return MailboxContentResult{}, mailboxMessageReadError(MailboxMessageReadErrorUnprocessable)
		default:
			return MailboxContentResult{}, mailboxMessageReadError(MailboxMessageReadErrorUnavailable)
		}
	}
	if content.SourceFormat != input.SourceFormat || !mailboxContentResultValid(content) {
		return MailboxContentResult{}, mailboxMessageReadError(MailboxMessageReadErrorInvalid)
	}
	return content, nil
}

func mailboxMessageDetailFromSnapshot(snapshot MailboxMessageSnapshot, content MailboxContentResult) (MailboxMessageDetail, error) {
	if !mailboxMessageSnapshotValid(snapshot) || !mailboxContentResultValid(content) {
		return MailboxMessageDetail{}, mailboxMessageReadError(MailboxMessageReadErrorInvalid)
	}
	summary := mailboxThreadDetailProjectionMessageCopy(snapshot.summary)
	var sanitizedHTML *string
	if content.SourceFormat == MailboxContentFormatRich {
		value := content.SanitizedHTML
		sanitizedHTML = &value
	}
	detail := MailboxMessageDetail{
		ID: summary.ID, ThreadID: summary.ThreadID, From: summary.From, To: summary.To, CC: summary.CC,
		Subject: summary.Subject, Snippet: summary.Snippet, SentAt: summary.SentAt, ReceivedAt: summary.ReceivedAt,
		Priority: summary.Priority, SizeBytes: summary.SizeBytes, HasAttachments: summary.HasAttachments,
		IsDraft: summary.IsDraft, FolderIDs: summary.FolderIDs, Seen: summary.Seen, Flagged: summary.Flagged,
		Answered: summary.Answered, Junk: summary.Junk, CalendarMethod: summary.CalendarMethod,
		CalendarEventID: summary.CalendarEventID, PlainBody: content.PlainBody, SanitizedHTML: sanitizedHTML,
		RemoteImagesBlocked: true, Attachments: make([]MailboxMessageAttachment, len(snapshot.attachments)), Calendar: nil,
	}
	for index, attachment := range snapshot.attachments {
		detail.Attachments[index] = mailboxMessageAttachmentCopy(attachment)
	}
	if !mailboxMessageDetailValid(detail) {
		return MailboxMessageDetail{}, mailboxMessageReadError(MailboxMessageReadErrorInvalid)
	}
	return detail, nil
}

func mailboxMessageOwnerError(err error) error {
	var serviceError *MailboxServiceError
	if !errors.As(err, &serviceError) || serviceError == nil {
		return mailboxMessageReadError(MailboxMessageReadErrorInvalid)
	}
	switch serviceError.Kind {
	case MailboxServiceErrorNotFound:
		return mailboxMessageReadError(MailboxMessageReadErrorNotFound)
	case MailboxServiceErrorDeleting:
		return mailboxMessageReadError(MailboxMessageReadErrorDeleting)
	case MailboxServiceErrorUnavailable:
		return mailboxMessageReadError(MailboxMessageReadErrorUnavailable)
	case MailboxServiceErrorInvalid:
		return mailboxMessageReadError(MailboxMessageReadErrorInvalid)
	default:
		return mailboxMessageReadError(MailboxMessageReadErrorInvalid)
	}
}

func mailboxMessageRepositoryError(err error) error {
	if kind, ok := mailboxMessageReadErrorKindOf(err); ok {
		return mailboxMessageReadError(kind)
	}
	return mailboxMessageReadError(MailboxMessageReadErrorUnavailable)
}

func mailboxMessageEngineError(err error) error {
	switch {
	case errors.Is(err, ErrEnginePayloadTooLarge):
		return mailboxMessageReadError(MailboxMessageReadErrorPayloadTooLarge)
	case errors.Is(err, ErrEngineNotFound), errors.Is(err, ErrEngineUnauthorizedAccount), errors.Is(err, ErrEngineConflict),
		errors.Is(err, ErrEngineInvalidCursor), errors.Is(err, ErrEngineUnavailable), errors.Is(err, ErrEngineIndeterminate),
		errors.Is(err, ErrEngineRejectedPolicy), errors.Is(err, ErrEngineRestartRequired):
		return mailboxMessageReadError(MailboxMessageReadErrorUnavailable)
	default:
		return mailboxMessageReadError(MailboxMessageReadErrorUnavailable)
	}
}
