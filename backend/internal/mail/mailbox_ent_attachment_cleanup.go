package mail

import (
	"bytes"
	"context"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailattachment"
	"github.com/valerii/hub/backend/ent/mailobject"
)

// mailboxEntDraftAttachmentCleanupLocked retains the exact confirmed attachment
// graph and its backing objects under the transaction's canonical lock order.
// The graph is intentionally private to the abandoned-upload transition.
type mailboxEntDraftAttachmentCleanupLocked struct {
	state       mailboxEntDraftAttachmentState
	attachments []*ent.MailAttachment
	objects     []*ent.MailObject
	objectByID  map[uuid.UUID]*ent.MailObject
}

// mailboxEntDraftAbandonLockAttachments locks only the draft's confirmed
// storage attachments, then their distinct backing objects, in raw UUID order.
// Pending or mixed ownership is corruption for abandon and fails closed.
func mailboxEntDraftAbandonLockAttachments(ctx context.Context, transaction *ent.Tx, draft *ent.MailDraft) (mailboxEntDraftAttachmentCleanupLocked, MailboxDraftRepositoryErrorKind) {
	if ctx == nil || transaction == nil || draft == nil || draft.ID == uuid.Nil || draft.MailAccountID == uuid.Nil || draft.DraftClientID == uuid.Nil {
		return mailboxEntDraftAttachmentCleanupLocked{}, MailboxDraftRepositoryErrorInvalid
	}

	attachments, err := transaction.MailAttachment.Query().
		Where(mailattachment.Or(mailattachment.MailDraftIDEQ(draft.ID), mailattachment.PendingMailDraftIDEQ(draft.ID))).
		Order(ent.Asc(mailattachment.FieldID)).
		Limit(MaxEngineAttachmentItems*2 + 1).
		ForUpdate().
		All(ctx)
	if err != nil {
		return mailboxEntDraftAttachmentCleanupLocked{}, MailboxDraftRepositoryErrorUnavailable
	}
	if len(attachments) > MaxEngineAttachmentItems {
		return mailboxEntDraftAttachmentCleanupLocked{}, MailboxDraftRepositoryErrorStateConflict
	}

	locked := mailboxEntDraftAttachmentCleanupLocked{
		state:       mailboxEntDraftAttachmentState{ConfirmedAttachmentIDs: make([]uuid.UUID, 0, len(attachments))},
		attachments: make([]*ent.MailAttachment, 0, len(attachments)),
		objects:     make([]*ent.MailObject, 0, len(attachments)),
		objectByID:  make(map[uuid.UUID]*ent.MailObject, len(attachments)),
	}
	objectIDs := make([]uuid.UUID, 0, len(attachments))
	seenAttachments := make(map[uuid.UUID]struct{}, len(attachments))
	seenObjects := make(map[uuid.UUID]struct{}, len(attachments))
	for _, attachment := range attachments {
		if attachment == nil || attachment.ID == uuid.Nil {
			return mailboxEntDraftAttachmentCleanupLocked{}, MailboxDraftRepositoryErrorStateConflict
		}
		if _, exists := seenAttachments[attachment.ID]; exists {
			return mailboxEntDraftAttachmentCleanupLocked{}, MailboxDraftRepositoryErrorStateConflict
		}
		seenAttachments[attachment.ID] = struct{}{}
		if attachment.MailDraftID == nil || *attachment.MailDraftID != draft.ID || attachment.PendingMailDraftID != nil || !mailboxEntDraftReadAttachmentValid(attachment, draft.ID, draft.DraftClientID, true, false) || attachment.ScanStatus != mailattachment.ScanStatus("clean") {
			return mailboxEntDraftAttachmentCleanupLocked{}, MailboxDraftRepositoryErrorStateConflict
		}
		objectID := *attachment.StorageObjectID
		if _, exists := seenObjects[objectID]; exists {
			return mailboxEntDraftAttachmentCleanupLocked{}, MailboxDraftRepositoryErrorStateConflict
		}
		seenObjects[objectID] = struct{}{}
		locked.state.ConfirmedAttachmentIDs = append(locked.state.ConfirmedAttachmentIDs, attachment.ID)
		locked.attachments = append(locked.attachments, attachment)
		objectIDs = append(objectIDs, objectID)
	}
	sort.Slice(objectIDs, func(left, right int) bool { return syncLessUUID(objectIDs[left], objectIDs[right]) })
	for _, objectID := range objectIDs {
		object, err := transaction.MailObject.Query().Where(mailobject.IDEQ(objectID)).ForUpdate().Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return mailboxEntDraftAttachmentCleanupLocked{}, MailboxDraftRepositoryErrorStateConflict
			}
			return mailboxEntDraftAttachmentCleanupLocked{}, MailboxDraftRepositoryErrorUnavailable
		}
		if object == nil || object.ID != objectID {
			return mailboxEntDraftAttachmentCleanupLocked{}, MailboxDraftRepositoryErrorStateConflict
		}
		locked.objects = append(locked.objects, object)
		locked.objectByID[objectID] = object
	}
	return locked, ""
}

// mailboxEntDraftAbandonCleanup detaches every verified confirmed attachment and
// retains an object deletion intent. It does not perform provider or object-store I/O.
func mailboxEntDraftAbandonCleanup(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, draft *ent.MailDraft, locked mailboxEntDraftAttachmentCleanupLocked, databaseNow time.Time) MailboxDraftRepositoryErrorKind {
	if ctx == nil || transaction == nil || !mailboxEntDraftRecoveryReceiptAccountValid(account) || draft == nil || draft.ID == uuid.Nil || draft.MailAccountID != account.ID || draft.DraftClientID == uuid.Nil || !mailboxEntValidTime(databaseNow) || len(locked.attachments) == 0 || len(locked.attachments) != len(locked.objects) || len(locked.state.ConfirmedAttachmentIDs) != len(locked.attachments) || len(locked.state.PendingAttachmentIDs) != 0 {
		return MailboxDraftRepositoryErrorInvalid
	}

	for index, attachment := range locked.attachments {
		if index != 0 && !syncLessUUID(locked.attachments[index-1].ID, attachment.ID) {
			return MailboxDraftRepositoryErrorStateConflict
		}
		if !mailboxEntDraftAbandonAttachmentValid(attachment, draft) {
			return MailboxDraftRepositoryErrorStateConflict
		}
		object := locked.objectByID[*attachment.StorageObjectID]
		if !mailboxEntDraftAbandonObjectValid(object, attachment, account, draft) {
			return MailboxDraftRepositoryErrorStateConflict
		}
		references, err := transaction.MailAttachment.Query().Where(mailattachment.StorageObjectIDEQ(object.ID)).Count(ctx)
		if err != nil {
			return MailboxDraftRepositoryErrorUnavailable
		}
		if references != 1 {
			return MailboxDraftRepositoryErrorStateConflict
		}
	}
	for index, object := range locked.objects {
		if index != 0 && !syncLessUUID(locked.objects[index-1].ID, object.ID) {
			return MailboxDraftRepositoryErrorStateConflict
		}
	}

	for _, attachment := range locked.attachments {
		if err := transaction.MailAttachment.DeleteOne(attachment).Exec(ctx); err != nil {
			return MailboxDraftRepositoryErrorUnavailable
		}
	}
	for _, object := range locked.objects {
		nextDeleteAt := databaseNow
		if object.UploadDrainUntil.After(nextDeleteAt) {
			nextDeleteAt = object.UploadDrainUntil.UTC()
		}
		if err := object.Update().
			SetUploadState(mailobject.UploadState("deleted")).
			ClearDeleteLeaseToken().
			SetDeleteLeaseGeneration(0).
			ClearDeleteLeaseUntil().
			SetDeleteAttempts(0).
			SetNextDeleteAt(nextDeleteAt).
			ClearDeleteTerminalAt().
			SetUpdatedAt(databaseNow).
			Exec(ctx); err != nil {
			return MailboxDraftRepositoryErrorUnavailable
		}
	}
	return ""
}

func mailboxEntDraftAbandonAttachmentValid(attachment *ent.MailAttachment, draft *ent.MailDraft) bool {
	return attachment != nil && draft != nil && attachment.MailDraftID != nil && *attachment.MailDraftID == draft.ID && attachment.PendingMailDraftID == nil && attachment.MailMessageID == nil && attachment.DraftClientID != nil && *attachment.DraftClientID == draft.DraftClientID && attachment.StorageObjectID != nil && *attachment.StorageObjectID != uuid.Nil && attachment.EngineBlobID == nil && attachment.ScanStatus == mailattachment.ScanStatus("clean") && mailboxEntDraftReadAttachmentValid(attachment, draft.ID, draft.DraftClientID, true, false)
}

func mailboxEntDraftAbandonObjectValid(object *ent.MailObject, attachment *ent.MailAttachment, account *ent.MailAccount, draft *ent.MailDraft) bool {
	if object == nil || attachment == nil || account == nil || draft == nil || attachment.StorageObjectID == nil || object.ID == uuid.Nil || object.ID != *attachment.StorageObjectID || object.DeletionID == uuid.Nil || object.OrganizationID != account.OrganizationID || object.MailAccountID != account.ID || object.MailAccountID != draft.MailAccountID || object.UploaderUserID != account.UserID || object.DraftClientID != draft.DraftClientID || object.UploadState != mailobject.UploadState("confirmed") || object.ExpectedSizeBytes != attachment.SizeBytes || object.ActualSizeBytes == nil || *object.ActualSizeBytes != object.ExpectedSizeBytes || object.Filename != attachment.Filename || object.MediaType != attachment.MediaType || len(object.SHA256) != AttachmentSHA256Bytes || !bytes.Equal(object.SHA256, attachment.SHA256) || object.NextDeleteAt != nil || object.DeleteLeaseToken != nil || object.DeleteLeaseUntil != nil || object.DeleteLeaseGeneration != 0 || object.DeleteAttempts != 0 || object.DeleteTerminalAt != nil || !mailboxEntValidTime(object.ExpiresAt) || !mailboxEntValidTime(object.UploadDrainUntil) {
		return false
	}
	uploadDrainUntil, ok := mailboxEntCheckedTimeAdd(object.ExpiresAt, mailboxAttachmentPresignObjectDrainLifetime)
	return ok && object.UploadDrainUntil.Equal(uploadDrainUntil)
}
