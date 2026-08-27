package mail

import (
	"context"
	"crypto/sha256"
	"sort"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailattachment"
	"github.com/valerii/hub/backend/ent/maildraft"
	"github.com/valerii/hub/backend/ent/mailobject"
)

var _ MailboxDraftAttachmentProofResolver = (*MailboxEntRepository)(nil)

// ResolveMailboxDraftAttachmentProof locks and proves the selected attachments for one draft request.
func (repository *MailboxEntRepository) ResolveMailboxDraftAttachmentProof(ctx context.Context, request mailboxDraftAttachmentProofRequest) (mailboxDraftAttachmentProof, error) {
	if repository == nil || repository.client == nil || ctx == nil || request.AccountID == uuid.Nil || request.DraftClientID == uuid.Nil || request.DraftID == uuid.Nil {
		return mailboxDraftAttachmentProof{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	if kind := mailboxEntDraftAttachmentProofRequestKind(request.AttachmentIDs); kind != "" {
		return mailboxDraftAttachmentProof{}, NewMailboxDraftRepositoryError(kind)
	}

	attachmentIDs := append([]uuid.UUID(nil), request.AttachmentIDs...)
	sort.Slice(attachmentIDs, func(left, right int) bool { return syncLessUUID(attachmentIDs[left], attachmentIDs[right]) })

	var result mailboxDraftAttachmentProof
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		if err := mailboxEntDraftReadLockAccount(ctx, transaction, request.AccountID); err != nil {
			return err
		}

		draftFound := false
		draft, err := transaction.MailDraft.Query().
			Where(
				maildraft.IDEQ(request.DraftID),
				maildraft.MailAccountIDEQ(request.AccountID),
				maildraft.DraftClientIDEQ(request.DraftClientID),
			).
			ForUpdate().
			Only(ctx)
		if err != nil {
			if !ent.IsNotFound(err) {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
			}
		} else {
			if draft == nil || draft.ID != request.DraftID || draft.MailAccountID != request.AccountID || draft.DraftClientID != request.DraftClientID {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
			}
			draftFound = true
		}

		attachments := make(map[uuid.UUID]*ent.MailAttachment, len(attachmentIDs))
		objectIDs := make([]uuid.UUID, 0, len(attachmentIDs))
		objectIDsSeen := make(map[uuid.UUID]struct{}, len(attachmentIDs))
		for _, attachmentID := range attachmentIDs {
			attachment, err := transaction.MailAttachment.Query().Where(mailattachment.IDEQ(attachmentID)).ForUpdate().Only(ctx)
			if err != nil {
				if ent.IsNotFound(err) {
					return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorAttachmentConflict)
				}
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
			}
			if !mailboxEntDraftAttachmentProofAttachmentValid(attachment, request, draftFound) {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorAttachmentConflict)
			}

			attachments[attachmentID] = attachment
			objectID := *attachment.StorageObjectID
			if _, found := objectIDsSeen[objectID]; !found {
				objectIDsSeen[objectID] = struct{}{}
				objectIDs = append(objectIDs, objectID)
			}
		}

		sort.Slice(objectIDs, func(left, right int) bool { return syncLessUUID(objectIDs[left], objectIDs[right]) })
		objects := make(map[uuid.UUID]*ent.MailObject, len(objectIDs))
		for _, objectID := range objectIDs {
			object, err := transaction.MailObject.Query().Where(mailobject.IDEQ(objectID)).ForUpdate().Only(ctx)
			if err != nil {
				if ent.IsNotFound(err) {
					return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorAttachmentConflict)
				}
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
			}
			if object == nil || object.ID != objectID {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorAttachmentConflict)
			}
			objects[objectID] = object
		}

		proof := make([]MailboxDraftRequestHashAttachment, len(request.AttachmentIDs))
		for index, attachmentID := range request.AttachmentIDs {
			attachment := attachments[attachmentID]
			if attachment == nil || !mailboxEntDraftReadStorageObjectValid(objects[*attachment.StorageObjectID], request.AccountID, request.DraftClientID, attachment) {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorAttachmentConflict)
			}
			var digest [sha256.Size]byte
			copy(digest[:], attachment.Sha256)
			proof[index] = MailboxDraftRequestHashAttachment{AttachmentID: attachmentID, SHA256: digest}
		}

		result = mailboxDraftAttachmentProof{
			accountID:     request.AccountID,
			draftClientID: request.DraftClientID,
			draftID:       request.DraftID,
			attachments:   proof,
		}
		return nil
	})
	if err != nil {
		return mailboxDraftAttachmentProof{}, mailboxEntDraftRepositoryError(err)
	}
	return result, nil
}

func mailboxEntDraftAttachmentProofRequestKind(attachmentIDs []uuid.UUID) MailboxDraftRepositoryErrorKind {
	if len(attachmentIDs) > MaxEngineAttachmentItems {
		return MailboxDraftRepositoryErrorAttachmentLimit
	}

	seen := make(map[uuid.UUID]struct{}, len(attachmentIDs))
	for _, attachmentID := range attachmentIDs {
		if attachmentID == uuid.Nil {
			return MailboxDraftRepositoryErrorInvalid
		}
		if _, found := seen[attachmentID]; found {
			return MailboxDraftRepositoryErrorInvalid
		}
		seen[attachmentID] = struct{}{}
	}
	return ""
}

func mailboxEntDraftAttachmentProofAttachmentValid(attachment *ent.MailAttachment, request mailboxDraftAttachmentProofRequest, draftFound bool) bool {
	if attachment == nil || attachment.ID == uuid.Nil || attachment.MailMessageID != nil || attachment.EngineBlobID != nil || attachment.DraftClientID == nil || *attachment.DraftClientID != request.DraftClientID || attachment.StorageObjectID == nil || *attachment.StorageObjectID == uuid.Nil || attachment.SizeBytes < 1 || attachment.SizeBytes > MaxEngineAttachmentBytes || attachment.ScanStatus != mailattachment.ScanStatus("clean") || !mailboxEntDraftAttachmentProofSHA256Valid(attachment.Sha256) || !mailboxEntApplyAttachmentOptionalTextValid(attachment.ContentID) || !mailboxEntApplyAttachmentTextValid(attachment.Filename, false) || !mailboxEntApplyAttachmentTextValid(attachment.MediaType, false) || !mailboxEntApplyAttachmentDispositionValid(attachment.Disposition) {
		return false
	}
	if attachment.Disposition == mailattachment.Disposition(EngineAttachmentDispositionInline) && attachment.ContentID == nil {
		return false
	}
	if attachment.MailDraftID == nil {
		return attachment.PendingMailDraftID == nil
	}
	if *attachment.MailDraftID != request.DraftID {
		return false
	}
	if !draftFound {
		return false
	}
	return attachment.PendingMailDraftID == nil || *attachment.PendingMailDraftID == request.DraftID
}

func mailboxEntDraftAttachmentProofSHA256Valid(value []byte) bool {
	if len(value) != sha256.Size {
		return false
	}
	for _, byteValue := range value {
		if byteValue != 0 {
			return true
		}
	}
	return false
}
