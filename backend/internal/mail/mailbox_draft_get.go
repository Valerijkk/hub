package mail

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
)

const mailboxDraftGetMaxContentBytes int64 = MaxEngineMessageBytes

type mailboxDraftGetScope struct {
	accountID       uuid.UUID
	engineAccountID string
}

// GetMailboxDraft returns the current owner-bound draft after rechecking the
// durable revision and engine message/version pair around engine I/O.
func (service *MailboxDraftService) GetMailboxDraft(ctx context.Context, actorID, draftID uuid.UUID) (MailboxDraftView, error) {
	if service == nil ||
		mailboxDraftServiceDependencyNil(service.owners) ||
		mailboxDraftServiceDependencyNil(service.repository) ||
		mailboxDraftServiceDependencyNil(service.engine) ||
		ctx == nil || actorID == uuid.Nil || draftID == uuid.Nil {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	if ctx.Err() != nil {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorUnavailable)
	}

	owner, err := service.owners.ResolveOwner(ctx, actorID)
	if err != nil {
		return MailboxDraftView{}, mailboxDraftGetOwnerError(err)
	}
	scope, err := mailboxDraftGetScopeFromOwner(actorID, owner)
	if err != nil {
		return MailboxDraftView{}, err
	}

	result, err := service.repository.GetMailboxDraft(ctx, MailboxDraftGetRequest{
		AccountID: scope.accountID,
		DraftID:   draftID,
	})
	if err != nil {
		return MailboxDraftView{}, mailboxDraftGetRepositoryError(err)
	}
	if !result.Found {
		if result.Draft != nil {
			return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
		}
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorNotFound)
	}
	if result.Draft == nil {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	if result.Draft.Status != MailboxDraftStatusActive {
		if mailboxDraftGetStatusValid(result.Draft.Status) {
			return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorConflict)
		}
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	if !mailboxDraftGetRecordValid(result.Draft, scope.accountID, draftID) {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	draft := mailboxDraftGetRecordCopy(*result.Draft)

	message, err := service.engine.GetDraft(ctx, EngineDraftReadRequest{
		AccountID:       scope.engineAccountID,
		DraftKey:        draft.EngineDraftKey,
		MessageID:       *draft.EngineMessageID,
		ExpectedVersion: *draft.EngineVersion,
		MaxContentBytes: mailboxDraftGetMaxContentBytes,
	})
	if err != nil {
		return MailboxDraftView{}, mailboxDraftGetEngineError(err)
	}
	if validateEngineMessage(message) != nil {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorUnavailable)
	}
	if int64(messageVariableBytes(message)) > mailboxDraftGetMaxContentBytes {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorPayloadTooLarge)
	}
	if !mailboxDraftGetEngineMessageMatches(scope, draft, message) {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorUnavailable)
	}

	content, err := mailboxDraftGetNormalizeContent(message.Body)
	if err != nil {
		return MailboxDraftView{}, err
	}

	rechecked, err := service.repository.RecheckMailboxDraftRead(ctx, MailboxDraftReadRecheckRequest{
		AccountID:       scope.accountID,
		DraftID:         draft.ID,
		Revision:        draft.Revision,
		EngineMessageID: *draft.EngineMessageID,
		EngineVersion:   *draft.EngineVersion,
	})
	if err != nil {
		return MailboxDraftView{}, mailboxDraftGetRepositoryError(err)
	}
	if !rechecked.Current {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorConflict)
	}

	return mailboxDraftGetViewCopy(MailboxDraftView{
		ID:            draft.ID,
		DraftClientID: draft.DraftClientID,
		Revision:      draft.Revision,
		Status:        draft.Status,
		Effective:     mailboxDraftGetEffectiveStateCopy(draft.Confirmed),
		Content: MailboxDraftContentView{
			From:       mailboxDraftGetAddressCopy(message.Projection.From),
			To:         mailboxDraftGetAddressesCopy(message.Projection.To),
			CC:         mailboxDraftGetAddressesCopy(message.Projection.CC),
			BCC:        mailboxDraftGetAddressesCopy(message.Projection.BCC),
			ReplyTo:    mailboxDraftGetAddressesCopy(message.Projection.ReplyTo),
			Subject:    strings.Clone(message.Projection.Subject),
			InReplyTo:  strings.Clone(message.Projection.InReplyTo),
			References: mailboxDraftGetStringsCopy(message.Projection.References),
			Priority:   message.Projection.Priority,
			Content:    mailboxDraftGetContentCopy(content),
		},
		UpdatedAt: draft.UpdatedAt.UTC(),
	}), nil
}

func mailboxDraftGetScopeFromOwner(actorID uuid.UUID, owner MailboxOwner) (mailboxDraftGetScope, error) {
	if actorID == uuid.Nil || owner.OrganizationID() == uuid.Nil {
		return mailboxDraftGetScope{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	accountID, hasAccount := owner.AccountID()
	projectionState, hasProjectionState := owner.ProjectionState()
	switch owner.State() {
	case MailboxOwnerNotConfigured:
		if hasAccount || accountID != uuid.Nil || owner.hasEngineAccountID || hasProjectionState {
			return mailboxDraftGetScope{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
		}
		return mailboxDraftGetScope{}, mailboxDraftServiceError(MailboxDraftServiceErrorNotFound)
	case MailboxOwnerDeleting:
		if !hasAccount || accountID == uuid.Nil || owner.hasEngineAccountID || !hasProjectionState || !validMailboxProjectionState(projectionState) {
			return mailboxDraftGetScope{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
		}
		return mailboxDraftGetScope{}, mailboxDraftServiceError(MailboxDraftServiceErrorDeleting)
	case MailboxOwnerProvisioning, MailboxOwnerSuspended, MailboxOwnerError:
		if !hasAccount || accountID == uuid.Nil || !hasProjectionState || !validMailboxProjectionState(projectionState) {
			return mailboxDraftGetScope{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
		}
		return mailboxDraftGetScope{}, mailboxDraftServiceError(MailboxDraftServiceErrorUnavailable)
	case MailboxOwnerActive:
		if !hasAccount || accountID == uuid.Nil || !owner.hasEngineAccountID || !syncValidOpaque(owner.engineAccountID, MaxEngineOpaqueIDBytes) || !hasProjectionState || !validMailboxProjectionState(projectionState) {
			return mailboxDraftGetScope{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
		}
		if projectionState != MailboxProjectionReady {
			return mailboxDraftGetScope{}, mailboxDraftServiceError(MailboxDraftServiceErrorUnavailable)
		}
		return mailboxDraftGetScope{accountID: accountID, engineAccountID: owner.engineAccountID}, nil
	default:
		return mailboxDraftGetScope{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
}

func mailboxDraftGetRecordValid(draft *MailboxDraftRecord, accountID, draftID uuid.UUID) bool {
	if draft == nil || accountID == uuid.Nil || draftID == uuid.Nil ||
		draft.ID != draftID || draft.AccountID != accountID || draft.DraftClientID == uuid.Nil ||
		!syncValidOpaque(draft.EngineDraftKey, MaxEngineOpaqueIDBytes) || draft.Revision < 1 ||
		draft.Status != MailboxDraftStatusActive || draft.PendingSave != nil || draft.LastMutationID != nil ||
		draft.PendingMutationKind != nil || !mailboxEntValidTime(draft.UpdatedAt) ||
		!mailboxDraftGetEffectiveStateValid(draft.Confirmed) || !mailboxDraftGetUUIDPointerValid(draft.ConfirmedMutationID) {
		return false
	}
	return draft.EngineMessageID != nil && draft.EngineVersion != nil &&
		syncValidOpaque(*draft.EngineMessageID, MaxEngineOpaqueIDBytes) &&
		syncValidOpaque(*draft.EngineVersion, MaxEngineOpaqueIDBytes)
}

func mailboxDraftGetStatusValid(status MailboxDraftStatus) bool {
	switch status {
	case MailboxDraftStatusActive,
		MailboxDraftStatusSaving,
		MailboxDraftStatusDeleting,
		MailboxDraftStatusUnknown,
		MailboxDraftStatusSubmitting,
		MailboxDraftStatusSubmitted,
		MailboxDraftStatusDeleted:
		return true
	default:
		return false
	}
}

func mailboxDraftGetEffectiveStateValid(state MailboxDraftEffectiveState) bool {
	if state.FromIdentityID == uuid.Nil || !mailboxDraftGetUUIDPointerValid(state.ReplyToMessageID) || !mailboxDraftGetUUIDPointerValid(state.ForwardMessageID) || len(state.AttachmentIDs) > MaxEngineAttachmentItems {
		return false
	}
	for index, attachmentID := range state.AttachmentIDs {
		if attachmentID == uuid.Nil || index > 0 && !syncLessUUID(state.AttachmentIDs[index-1], attachmentID) {
			return false
		}
	}
	return true
}

func mailboxDraftGetUUIDPointerValid(value *uuid.UUID) bool {
	return value == nil || *value != uuid.Nil
}

func mailboxDraftGetEngineMessageMatches(scope mailboxDraftGetScope, draft MailboxDraftRecord, message EngineMessage) bool {
	return scope.accountID != uuid.Nil &&
		syncValidOpaque(scope.engineAccountID, MaxEngineOpaqueIDBytes) &&
		draft.EngineMessageID != nil && draft.EngineVersion != nil &&
		message.Projection.AccountID == scope.engineAccountID &&
		message.Projection.MessageID == *draft.EngineMessageID &&
		message.Projection.Version == *draft.EngineVersion &&
		message.Projection.Flags.Draft &&
		message.Projection.Origin == EngineProjectionOriginConfirmedOwnerDraft
}

func mailboxDraftGetNormalizeContent(body EngineMessageBody) (MailboxContentResult, error) {
	var input MailboxContentInput
	switch body.Format {
	case EngineBodyPlain:
		if body.RichBody != "" {
			return MailboxContentResult{}, mailboxDraftServiceError(MailboxDraftServiceErrorUnprocessable)
		}
		input = MailboxContentInput{SourceFormat: MailboxContentFormatPlain, Body: body.PlainBody}
	case EngineBodyRich:
		input = MailboxContentInput{SourceFormat: MailboxContentFormatRich, Body: body.RichBody}
	default:
		return MailboxContentResult{}, mailboxDraftServiceError(MailboxDraftServiceErrorUnprocessable)
	}

	content, err := NormalizeMailboxContent(input)
	if err != nil {
		switch {
		case errors.Is(err, ErrMailboxContentTooLarge):
			return MailboxContentResult{}, mailboxDraftServiceError(MailboxDraftServiceErrorPayloadTooLarge)
		case errors.Is(err, ErrMailboxContentInvalid):
			return MailboxContentResult{}, mailboxDraftServiceError(MailboxDraftServiceErrorUnprocessable)
		default:
			return MailboxContentResult{}, mailboxDraftServiceError(MailboxDraftServiceErrorUnavailable)
		}
	}
	if content.SourceFormat != input.SourceFormat || !mailboxContentResultValid(content) {
		return MailboxContentResult{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	return mailboxDraftGetContentCopy(content), nil
}

func mailboxDraftGetOwnerError(err error) error {
	var ownerError *MailboxServiceError
	if !errors.As(err, &ownerError) || ownerError == nil {
		return mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	switch ownerError.Kind {
	case MailboxServiceErrorNotFound:
		return mailboxDraftServiceError(MailboxDraftServiceErrorNotFound)
	case MailboxServiceErrorDeleting:
		return mailboxDraftServiceError(MailboxDraftServiceErrorDeleting)
	case MailboxServiceErrorUnavailable:
		return mailboxDraftServiceError(MailboxDraftServiceErrorUnavailable)
	case MailboxServiceErrorInvalid:
		return mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	default:
		return mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
}

func mailboxDraftGetRepositoryError(err error) error {
	var repositoryError *MailboxDraftRepositoryError
	if !errors.As(err, &repositoryError) || repositoryError == nil {
		return mailboxDraftServiceError(MailboxDraftServiceErrorUnavailable)
	}
	switch repositoryError.Kind() {
	case MailboxDraftRepositoryErrorInvalid:
		return mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	case MailboxDraftRepositoryErrorNotFound:
		return mailboxDraftServiceError(MailboxDraftServiceErrorNotFound)
	case MailboxDraftRepositoryErrorAccountDeleting:
		return mailboxDraftServiceError(MailboxDraftServiceErrorDeleting)
	case MailboxDraftRepositoryErrorStateConflict, MailboxDraftRepositoryErrorFenceLost,
		MailboxDraftRepositoryErrorIdempotencyConflict, MailboxDraftRepositoryErrorAttachmentConflict:
		return mailboxDraftServiceError(MailboxDraftServiceErrorConflict)
	case MailboxDraftRepositoryErrorUnavailable, MailboxDraftRepositoryErrorAttachmentLimit:
		return mailboxDraftServiceError(MailboxDraftServiceErrorUnavailable)
	default:
		return mailboxDraftServiceError(MailboxDraftServiceErrorUnavailable)
	}
}

func mailboxDraftGetEngineError(err error) error {
	if errors.Is(err, ErrEnginePayloadTooLarge) {
		return mailboxDraftServiceError(MailboxDraftServiceErrorPayloadTooLarge)
	}
	return mailboxDraftServiceError(MailboxDraftServiceErrorUnavailable)
}

func mailboxDraftGetViewCopy(view MailboxDraftView) MailboxDraftView {
	view.Effective = mailboxDraftGetEffectiveStateCopy(view.Effective)
	view.Content = MailboxDraftContentView{
		From:       mailboxDraftGetAddressCopy(view.Content.From),
		To:         mailboxDraftGetAddressesCopy(view.Content.To),
		CC:         mailboxDraftGetAddressesCopy(view.Content.CC),
		BCC:        mailboxDraftGetAddressesCopy(view.Content.BCC),
		ReplyTo:    mailboxDraftGetAddressesCopy(view.Content.ReplyTo),
		Subject:    strings.Clone(view.Content.Subject),
		InReplyTo:  strings.Clone(view.Content.InReplyTo),
		References: mailboxDraftGetStringsCopy(view.Content.References),
		Priority:   view.Content.Priority,
		Content:    mailboxDraftGetContentCopy(view.Content.Content),
	}
	view.UpdatedAt = view.UpdatedAt.UTC()
	return view
}

func mailboxDraftGetRecordCopy(draft MailboxDraftRecord) MailboxDraftRecord {
	draft.Confirmed = mailboxDraftGetEffectiveStateCopy(draft.Confirmed)
	draft.EngineMessageID = mailboxDraftGetStringPointerCopy(draft.EngineMessageID)
	draft.EngineVersion = mailboxDraftGetStringPointerCopy(draft.EngineVersion)
	draft.ConfirmedMutationID = mailboxDraftGetUUIDPointerCopy(draft.ConfirmedMutationID)
	draft.UpdatedAt = draft.UpdatedAt.UTC()
	return draft
}

func mailboxDraftGetEffectiveStateCopy(state MailboxDraftEffectiveState) MailboxDraftEffectiveState {
	return MailboxDraftEffectiveState{
		FromIdentityID:   state.FromIdentityID,
		ReplyToMessageID: mailboxDraftGetUUIDPointerCopy(state.ReplyToMessageID),
		ForwardMessageID: mailboxDraftGetUUIDPointerCopy(state.ForwardMessageID),
		AttachmentIDs:    append([]uuid.UUID(nil), state.AttachmentIDs...),
	}
}

func mailboxDraftGetContentCopy(content MailboxContentResult) MailboxContentResult {
	return MailboxContentResult{
		SourceFormat:        content.SourceFormat,
		SanitizerVersion:    content.SanitizerVersion,
		PlainBody:           strings.Clone(content.PlainBody),
		SanitizedHTML:       strings.Clone(content.SanitizedHTML),
		RemoteImagesBlocked: content.RemoteImagesBlocked,
	}
}

func mailboxDraftGetAddressCopy(address EngineAddress) EngineAddress {
	return EngineAddress{Email: strings.Clone(address.Email), Name: strings.Clone(address.Name)}
}

func mailboxDraftGetAddressesCopy(addresses []EngineAddress) []EngineAddress {
	result := make([]EngineAddress, len(addresses))
	for index, address := range addresses {
		result[index] = mailboxDraftGetAddressCopy(address)
	}
	return result
}

func mailboxDraftGetStringsCopy(values []string) []string {
	result := make([]string, len(values))
	for index, value := range values {
		result[index] = strings.Clone(value)
	}
	return result
}

func mailboxDraftGetUUIDPointerCopy(value *uuid.UUID) *uuid.UUID {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

func mailboxDraftGetStringPointerCopy(value *string) *string {
	if value == nil {
		return nil
	}
	copied := strings.Clone(*value)
	return &copied
}
