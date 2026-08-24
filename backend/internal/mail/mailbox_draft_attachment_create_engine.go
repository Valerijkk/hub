package mail

import (
	"context"
	"errors"
	"time"
)

// errMailboxDraftAttachmentCreatePersistedBlobNotFound is private evidence
// that the provider rejected the exact durable import blob before mutation.
// It is intentionally distinct from the provider-wide restart sentinel.
var errMailboxDraftAttachmentCreatePersistedBlobNotFound = errors.New("mailbox draft create persisted blob not found")

// mailboxDraftAttachmentCreateEngine is the provider-neutral private bridge
// for the two-step attachment-backed initial draft create. It deliberately is
// not part of MailboxDraftEngine: ordinary draft saves do not receive blob
// ownership or MIME-import capabilities.
type mailboxDraftAttachmentCreateEngine interface {
	uploadMailboxDraftCreate(context.Context, EngineDraft, *mailboxDraftAttachmentCapability) (mailboxDraftAttachmentCreateUpload, error)
	importMailboxDraftCreate(context.Context, mailboxDraftAttachmentCreateImport) (EngineMessage, error)
}

type mailboxDraftAttachmentCreateUpload struct {
	blobID    string
	sizeBytes int64
}

type mailboxDraftAttachmentCreateImport struct {
	accountID  string
	draftKey   string
	mutationID string
	blobID     string
	sizeBytes  int64
	receivedAt time.Time
}

func mailboxDraftAttachmentCreatePersistedBlobNotFound(err error) bool {
	return errors.Is(err, errMailboxDraftAttachmentCreatePersistedBlobNotFound)
}

func mailboxDraftAttachmentCreateUploadValid(value mailboxDraftAttachmentCreateUpload) bool {
	return validOpaqueID(value.blobID) && value.sizeBytes >= 1 && value.sizeBytes <= MaxEngineMessageBytes
}

func mailboxDraftAttachmentCreateImportValid(value mailboxDraftAttachmentCreateImport) bool {
	return validAccountID(value.accountID) &&
		validSafeText(value.draftKey, MaxEngineOpaqueIDBytes, true) &&
		validSafeText(value.mutationID, MaxEngineOpaqueIDBytes, true) &&
		validOpaqueID(value.blobID) &&
		value.sizeBytes >= 1 && value.sizeBytes <= MaxEngineMessageBytes &&
		mailboxDraftAttachmentCreateReceivedAtValid(value.receivedAt)
}

func mailboxDraftAttachmentCreateReceivedAtValid(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond() == 0
}

// mailboxDraftAttachmentCreateDraft copies only approved attachment metadata
// from the sealed capability. Object locators and attachment bytes remain
// unreachable from the service and its response types.
func mailboxDraftAttachmentCreateDraft(base EngineDraft, capability *mailboxDraftAttachmentCapability) (EngineDraft, bool) {
	if capability == nil || base.Message.Attachments != nil || !mailboxDraftAttachmentCreateBindingMatches(capability, base) {
		return EngineDraft{}, false
	}

	draft := base
	draft.Message.Attachments = make([]EngineDraftAttachment, len(capability.sources))
	for index, source := range capability.sources {
		if source == nil || source.open == nil || source.sizeBytes < 1 || source.sizeBytes > MaxEngineAttachmentBytes || source.scanStatus != AttachmentScanStatusClean {
			return EngineDraft{}, false
		}
		attachment := cloneMailboxDraftAttachment(source.attachment)
		if len(attachment.Bytes) != 0 {
			return EngineDraft{}, false
		}
		draft.Message.Attachments[index] = attachment
	}
	// validateDraftRequest intentionally rejects byte-less attachments. The
	// provider uploader receives these attachments through the sealed source
	// capability instead, so capability matching is the relevant validation.
	if !mailboxDraftAttachmentCapabilityMatches(capability, draft) {
		return EngineDraft{}, false
	}
	return draft, true
}

func mailboxDraftAttachmentCreateBindingMatches(capability *mailboxDraftAttachmentCapability, draft EngineDraft) bool {
	return capability != nil &&
		mailboxDraftAttachmentBindingValid(capability.binding) &&
		capability.binding.AccountID == draft.AccountID &&
		capability.binding.DraftKey == draft.DraftKey &&
		capability.binding.MutationID == draft.MutationID &&
		capability.maxAttachmentBytes >= 1 &&
		capability.maxAttachmentBytes <= MaxEngineAttachmentBytes &&
		len(capability.sources) != 0 &&
		len(capability.sources) <= MaxEngineAttachmentItems
}
