package mail

import (
	"context"
	"errors"
)

var _ mailboxDraftAttachmentCreateEngine = (*StalwartEngine)(nil)

func (engine *StalwartEngine) uploadMailboxDraftCreate(ctx context.Context, draft EngineDraft, capability *mailboxDraftAttachmentCapability) (mailboxDraftAttachmentCreateUpload, error) {
	if engine == nil || ctx == nil || capability == nil || !mailboxDraftAttachmentCapabilityMatches(capability, draft) {
		return mailboxDraftAttachmentCreateUpload{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}

	uploaded, err := engine.UploadDraftMIME(ctx, draft, capability)
	if err != nil {
		return mailboxDraftAttachmentCreateUpload{}, err
	}
	result := mailboxDraftAttachmentCreateUpload{blobID: uploaded.BlobID, sizeBytes: uploaded.SizeBytes}
	if uploaded.AccountID != draft.AccountID || !mailboxDraftAttachmentCreateUploadValid(result) {
		return mailboxDraftAttachmentCreateUpload{}, stalwartEngineError(EngineErrorIndeterminate)
	}
	return result, nil
}

func (engine *StalwartEngine) importMailboxDraftCreate(ctx context.Context, request mailboxDraftAttachmentCreateImport) (EngineMessage, error) {
	if engine == nil || ctx == nil || !mailboxDraftAttachmentCreateImportValid(request) {
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
	// Only the raw Apply path emits this private, exact response proof. Generic
	// provider restart and transport ambiguity remain provider errors so the
	// caller durably records Unknown instead of re-uploading.
	if errors.Is(err, errStalwartDraftMIMEImportExactBlobNotFound) {
		return EngineMessage{}, errMailboxDraftAttachmentCreatePersistedBlobNotFound
	}
	return message, err
}
