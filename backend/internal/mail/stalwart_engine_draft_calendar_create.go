package mail

import (
	"context"
	"errors"
)

var _ mailboxDraftCalendarCreateEngine = (*StalwartEngine)(nil)

func (engine *StalwartEngine) uploadCalendarDraftCreateMIME(
	ctx context.Context,
	draft EngineDraft,
	calendar *calendarMIMEPart,
) (mailboxDraftCalendarCreateUpload, error) {
	if engine == nil ||
		ctx == nil ||
		calendar == nil ||
		!calendar.valid() ||
		len(draft.Message.Attachments) != 0 ||
		!mailboxDraftMIMECalendarRecipientValid(draft, *calendar) {
		return mailboxDraftCalendarCreateUpload{},
			stalwartEngineError(EngineErrorRejectedPolicy)
	}

	result, err := engine.uploadCalendarDraftMIME(ctx, draft, calendar)
	if err != nil {
		return mailboxDraftCalendarCreateUpload{}, err
	}

	upload := mailboxDraftCalendarCreateUpload{
		blobID:    result.BlobID,
		sizeBytes: result.SizeBytes,
	}
	if !validAccountID(result.AccountID) ||
		result.AccountID != draft.AccountID ||
		!validMailboxDraftCalendarCreateUpload(upload) {
		return mailboxDraftCalendarCreateUpload{},
			stalwartEngineError(EngineErrorIndeterminate)
	}

	return upload, nil
}

func (engine *StalwartEngine) importCalendarDraftCreateMIME(
	ctx context.Context,
	request mailboxDraftCalendarCreateImport,
) (EngineMessage, error) {
	if engine == nil ||
		ctx == nil ||
		!validMailboxDraftCalendarCreateImport(request) {
		return EngineMessage{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}

	message, err := engine.ApplyDraftMIMEImport(ctx, StalwartDraftMIMEImportApplyRequest{
		AccountID:  request.accountID,
		DraftKey:   request.draftKey,
		MutationID: request.mutationID,
		BlobID:     request.blobID,
		SizeBytes:  request.sizeBytes,
		ReceivedAt: request.receivedAt,
	})
	if err != nil && errors.Is(err, errStalwartDraftMIMEImportExactBlobNotFound) {
		return EngineMessage{}, errMailboxDraftCalendarCreatePersistedBlobNotFound
	}
	if err != nil {
		return EngineMessage{}, err
	}

	return message, nil
}
