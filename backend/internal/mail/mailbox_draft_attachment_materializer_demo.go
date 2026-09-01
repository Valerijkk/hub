//go:build !qa && !release

package mail

import "context"

// mailboxDraftAttachmentMaterializerDemo deliberately accepts only attachmentless
// draft work in the local Fake-backed runtime. It must not acquire storage or
// network dependencies.
type mailboxDraftAttachmentMaterializerDemo struct{}

var _ mailboxDraftAttachmentMaterializer = mailboxDraftAttachmentMaterializerDemo{}

// NewMailboxDraftAttachmentMaterializerDemo returns the local Fake-backed
// attachment materializer for development wiring.
func NewMailboxDraftAttachmentMaterializerDemo() mailboxDraftAttachmentMaterializer {
	return mailboxDraftAttachmentMaterializerDemo{}
}

func (mailboxDraftAttachmentMaterializerDemo) MaterializeMailboxDraftAttachments(
	ctx context.Context,
	request mailboxDraftAttachmentMaterializationRequest,
) (mailboxDraftAttachmentMaterializationResult, error) {
	if ctx == nil || !mailboxDraftAttachmentMaterializationRequestValid(request) {
		return mailboxDraftAttachmentMaterializationResult{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	if len(request.EffectiveState.AttachmentIDs) != 0 {
		return mailboxDraftAttachmentMaterializationResult{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorAttachmentConflict)
	}
	return mailboxDraftAttachmentMaterializationResult{}, nil
}
