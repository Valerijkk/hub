package mail

import "context"

const stalwartDraftMIMEUploadContentType = "message/rfc822"

// StalwartDraftMIMEUploadResult contains the opaque account-bound blob
// reference confirmed by exactly one MIME upload. It is intentionally limited
// to the data required by the subsequent Email/import operation.
type StalwartDraftMIMEUploadResult struct {
	AccountID string
	BlobID    string
	SizeBytes int64
}

// UploadDraftMIME streams one attachment-backed draft as RFC 822 MIME to the
// exact account bound to the draft. It performs no import, persistence,
// reconciliation, retry, buffering, or logging. After UploadBlob is invoked,
// any transport ambiguity remains indeterminate and must be reconciled before
// another mutation.
func (engine *StalwartEngine) UploadDraftMIME(ctx context.Context, draft EngineDraft, capability *mailboxDraftAttachmentCapability) (StalwartDraftMIMEUploadResult, error) {
	if ctx == nil || capability == nil || !mailboxDraftAttachmentCapabilityMatches(capability, draft) {
		return StalwartDraftMIMEUploadResult{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	if err := stalwartMailboxContextError(ctx); err != nil {
		return StalwartDraftMIMEUploadResult{}, mapStalwartEngineError(err, false)
	}

	plan, err := mailboxDraftMIMEPlanForDraft(ctx, draft, capability)
	if err != nil {
		return StalwartDraftMIMEUploadResult{}, err
	}
	stream, sizeBytes, err := newMailboxDraftMIMEStream(ctx, plan)
	if err != nil {
		if contextErr := stalwartMailboxContextError(ctx); contextErr != nil {
			return StalwartDraftMIMEUploadResult{}, mapStalwartEngineError(contextErr, false)
		}
		return StalwartDraftMIMEUploadResult{}, err
	}
	client, err := engine.mailboxClient(ctx, draft.AccountID)
	if err != nil {
		_ = stream.Close()
		return StalwartDraftMIMEUploadResult{}, err
	}

	uploaded, err := client.UploadBlob(ctx, StalwartMailboxUploadRequest{
		ContentType: stalwartDraftMIMEUploadContentType,
		Content:     stream,
		SizeBytes:   sizeBytes,
	})
	if err != nil {
		// UploadBlob already marks every post-dispatch ambiguity as
		// ErrorIndeterminate. Preserve its pre-dispatch classifications while
		// keeping indeterminate results indeterminate and non-retryable here.
		return StalwartDraftMIMEUploadResult{}, mapStalwartEngineError(err, false)
	}
	if uploaded.AccountID != draft.AccountID || !validOpaqueID(uploaded.AccountID) || !validOpaqueID(uploaded.BlobID) || uploaded.SizeBytes != sizeBytes {
		return StalwartDraftMIMEUploadResult{}, stalwartEngineError(EngineErrorIndeterminate)
	}
	return StalwartDraftMIMEUploadResult{
		AccountID: uploaded.AccountID,
		BlobID:    uploaded.BlobID,
		SizeBytes: uploaded.SizeBytes,
	}, nil
}
