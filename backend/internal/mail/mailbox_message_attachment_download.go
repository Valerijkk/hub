package mail

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"unicode/utf8"

	"github.com/google/uuid"
)

// MailboxMessageAttachmentDownload is one safe, owner-bound attachment payload.
type MailboxMessageAttachmentDownload struct {
	Filename  string
	SizeBytes int64
	Bytes     []byte
}

// DownloadMailboxMessageAttachment returns one bounded engine blob only after local ownership and projection recheck.
func (service *MailboxMessageService) DownloadMailboxMessageAttachment(ctx context.Context, actorID, messageID, attachmentID uuid.UUID) (MailboxMessageAttachmentDownload, error) {
	if service == nil || mailboxFolderProjectionDependencyNil(service.owners) || mailboxFolderProjectionDependencyNil(service.repository) || mailboxFolderProjectionDependencyNil(service.engine) ||
		ctx == nil || actorID == uuid.Nil || messageID == uuid.Nil || attachmentID == uuid.Nil {
		return MailboxMessageAttachmentDownload{}, mailboxMessageReadError(MailboxMessageReadErrorInvalid)
	}
	if ctx.Err() != nil {
		return MailboxMessageAttachmentDownload{}, mailboxMessageReadError(MailboxMessageReadErrorUnavailable)
	}
	owner, err := service.owners.ResolveOwner(ctx, actorID)
	if err != nil {
		return MailboxMessageAttachmentDownload{}, mailboxMessageOwnerError(err)
	}
	scope, err := mailboxMessageReadScopeFromOwner(actorID, owner)
	if err != nil {
		return MailboxMessageAttachmentDownload{}, err
	}
	snapshot, err := service.repository.ResolveMailboxMessage(ctx, scope, messageID)
	if err != nil {
		return MailboxMessageAttachmentDownload{}, mailboxMessageRepositoryError(err)
	}
	if !mailboxMessageSnapshotValid(snapshot) || snapshot.summary.ID != messageID || snapshot.engineAccountID != scope.EngineAccountID {
		return MailboxMessageAttachmentDownload{}, mailboxMessageReadError(MailboxMessageReadErrorInvalid)
	}
	attachment, err := mailboxMessageAttachmentFromSnapshot(snapshot, attachmentID)
	if err != nil {
		return MailboxMessageAttachmentDownload{}, err
	}
	maxBytes := attachment.SizeBytes
	if maxBytes == 0 {
		maxBytes = 1
	}
	blob, err := service.engine.GetBlob(ctx, EngineBlobReadRequest{
		AccountID: scope.EngineAccountID,
		BlobID:    attachment.engineBlobID,
		MaxBytes:  maxBytes,
	})
	if err != nil {
		return MailboxMessageAttachmentDownload{}, mailboxMessageEngineError(err)
	}
	result, ok := mailboxMessageAttachmentDownloadFromBlob(scope, attachment, maxBytes, blob)
	if !ok {
		return MailboxMessageAttachmentDownload{}, mailboxMessageReadError(MailboxMessageReadErrorUnavailable)
	}
	if err := service.repository.RecheckMailboxMessage(ctx, scope, snapshot); err != nil {
		return MailboxMessageAttachmentDownload{}, mailboxMessageRepositoryError(err)
	}
	return result, nil
}

func mailboxMessageAttachmentFromSnapshot(snapshot MailboxMessageSnapshot, attachmentID uuid.UUID) (MailboxMessageAttachment, error) {
	if !mailboxMessageSnapshotValid(snapshot) || attachmentID == uuid.Nil {
		return MailboxMessageAttachment{}, mailboxMessageReadError(MailboxMessageReadErrorInvalid)
	}
	var attachment MailboxMessageAttachment
	found := 0
	for _, candidate := range snapshot.attachments {
		if candidate.ID == attachmentID {
			attachment = mailboxMessageAttachmentCopy(candidate)
			found++
		}
	}
	switch found {
	case 0:
		return MailboxMessageAttachment{}, mailboxMessageReadError(MailboxMessageReadErrorNotFound)
	case 1:
		return attachment, nil
	default:
		return MailboxMessageAttachment{}, mailboxMessageReadError(MailboxMessageReadErrorInvalid)
	}
}

func mailboxMessageAttachmentDownloadFromBlob(scope MailboxMessageReadScope, attachment MailboxMessageAttachment, maxBytes int64, blob EngineBlob) (MailboxMessageAttachmentDownload, bool) {
	exactBound := attachment.SizeBytes > 0 && maxBytes == attachment.SizeBytes || attachment.SizeBytes == 0 && maxBytes == 1
	if !mailboxMessageReadScopeValid(scope) || !mailboxMessageAttachmentValid(attachment) || !exactBound ||
		blob.AccountID != scope.EngineAccountID || blob.BlobID != attachment.engineBlobID || blob.SizeBytes != attachment.SizeBytes ||
		blob.SizeBytes < 0 || blob.SizeBytes > maxBytes || blob.SizeBytes > MaxEngineAttachmentBytes || int64(len(blob.Bytes)) != blob.SizeBytes {
		return MailboxMessageAttachmentDownload{}, false
	}
	content := append([]byte(nil), blob.Bytes...)
	digest := sha256.Sum256(content)
	if hex.EncodeToString(digest[:]) != attachment.SHA256 {
		return MailboxMessageAttachmentDownload{}, false
	}
	result := MailboxMessageAttachmentDownload{Filename: attachment.Filename, SizeBytes: attachment.SizeBytes, Bytes: content}
	return result, mailboxMessageAttachmentDownloadValid(result)
}

func mailboxMessageAttachmentDownloadValid(attachment MailboxMessageAttachmentDownload) bool {
	return syncValidSafeText(attachment.Filename, MaxEngineMIMEFieldBytes, false) && utf8.RuneCountInString(attachment.Filename) <= 255 &&
		attachment.SizeBytes >= 0 && attachment.SizeBytes <= MaxEngineAttachmentBytes && int64(len(attachment.Bytes)) == attachment.SizeBytes
}
