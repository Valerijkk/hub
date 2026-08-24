package mail

import (
	"bytes"
	"context"
	"crypto/sha256"
	"io"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/mailattachment"
	"github.com/valerii/hub/backend/ent/maildraft"
	"github.com/valerii/hub/backend/ent/mailobject"
)

// MailboxEntDraftAttachmentMaterializer proves durable storage attachments
// under the caller's current engine fence. The object store is intentionally
// used only after this transaction has committed, through sealed sources.
type MailboxEntDraftAttachmentMaterializer struct {
	repository  *MailboxEntRepository
	objectStore MailboxObjectStore
}

var _ mailboxDraftAttachmentMaterializer = (*MailboxEntDraftAttachmentMaterializer)(nil)

// NewMailboxEntDraftAttachmentMaterializer creates the concrete Ent-backed
// sealed-source materializer used by the draft service composition root.
func NewMailboxEntDraftAttachmentMaterializer(repository *MailboxEntRepository, objectStore MailboxObjectStore) (*MailboxEntDraftAttachmentMaterializer, error) {
	if repository == nil || repository.client == nil || mailboxDraftAttachmentMaterializerDependencyNil(objectStore) {
		return nil, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	return &MailboxEntDraftAttachmentMaterializer{repository: repository, objectStore: objectStore}, nil
}

// MaterializeMailboxDraftAttachments returns a sealed capability for the exact
// pending create/upload or save graph. Attachmentless graphs deliberately
// retain no capability and therefore do not open storage.
func (materializer *MailboxEntDraftAttachmentMaterializer) MaterializeMailboxDraftAttachments(ctx context.Context, request mailboxDraftAttachmentMaterializationRequest) (mailboxDraftAttachmentMaterializationResult, error) {
	if materializer == nil || materializer.repository == nil || materializer.repository.client == nil || mailboxDraftAttachmentMaterializerDependencyNil(materializer.objectStore) || ctx == nil || !mailboxDraftAttachmentMaterializationRequestValid(request) {
		return mailboxDraftAttachmentMaterializationResult{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	if len(request.EffectiveState.AttachmentIDs) == 0 {
		return mailboxDraftAttachmentMaterializationResult{}, nil
	}

	var result mailboxDraftAttachmentMaterializationResult
	err := materializer.repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		owner, err := mailboxEntResolveOwnerForUpdate(ctx, transaction, request.ActorID)
		if err != nil {
			return mailboxEntDraftRepositoryError(err)
		}

		account, err := transaction.MailAccount.Query().
			Where(mailaccount.IDEQ(request.AccountID)).
			ForUpdate().
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorNotFound)
			}
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		if !mailboxEntDraftAttachmentMaterializerOwnerMatches(owner, account, request) {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorNotFound)
		}
		if kind := mailboxEntDraftCreateAccountKind(account); kind != "" {
			return NewMailboxDraftRepositoryError(kind)
		}

		draft, err := transaction.MailDraft.Query().
			Where(
				maildraft.IDEQ(request.DraftID),
				maildraft.MailAccountIDEQ(request.AccountID),
				maildraft.DraftClientIDEQ(request.DraftClientID),
				maildraft.RevisionEQ(request.Revision),
			).
			ForUpdate().
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorStateConflict)
			}
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		if !mailboxEntDraftReadRowValid(draft, request.AccountID, request.DraftID) {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorStateConflict)
		}

		locked, kind := mailboxEntDraftAttachmentMaterializerLockAttachments(ctx, transaction, draft)
		if kind != "" {
			return NewMailboxDraftRepositoryError(kind)
		}
		record, ok := mailboxEntDraftReadRecord(draft, locked.state)
		if !ok || !mailboxEntDraftAttachmentMaterializerRecordMatches(record, request) {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorStateConflict)
		}

		databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return mailboxEntDraftRepositoryError(err)
		}
		if !mailboxEntDraftCreateFenceMatches(account, request.Fence, databaseNow) {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorFenceLost)
		}

		specs, kind := mailboxEntDraftAttachmentMaterializerSpecs(locked, request, databaseNow, materializer.objectStore)
		if kind != "" {
			return NewMailboxDraftRepositoryError(kind)
		}
		capability, err := newMailboxDraftAttachmentCapability(mailboxDraftAttachmentBinding{
			AccountID:  request.EngineAccountID,
			DraftKey:   record.EngineDraftKey,
			MutationID: request.MutationID.String(),
		}, MaxEngineAttachmentBytes, specs)
		if err != nil || capability == nil {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorStateConflict)
		}
		result = mailboxDraftAttachmentMaterializationResult{capability: capability}
		return nil
	})
	if err != nil {
		return mailboxDraftAttachmentMaterializationResult{}, mailboxEntDraftRepositoryError(err)
	}
	return result, nil
}

type mailboxEntDraftAttachmentMaterializerLockedAttachments struct {
	state       mailboxEntDraftAttachmentState
	attachments map[uuid.UUID]*ent.MailAttachment
	objects     map[uuid.UUID]*ent.MailObject
}

func mailboxEntDraftAttachmentMaterializerOwnerMatches(owner MailboxOwnerRecord, account *ent.MailAccount, request mailboxDraftAttachmentMaterializationRequest) bool {
	return account != nil && account.ID == request.AccountID && account.UserID == request.ActorID && account.OrganizationID == request.OrganizationID && owner.ActorID == request.ActorID && owner.OrganizationID == request.OrganizationID && owner.AccountID == request.AccountID && owner.State == MailboxOwnerActive && owner.EngineAccountID != nil && *owner.EngineAccountID == request.EngineAccountID
}

// mailboxEntDraftAttachmentMaterializerLockAttachments holds O union D in raw
// UUID order before holding every referenced object in raw UUID order.
func mailboxEntDraftAttachmentMaterializerLockAttachments(ctx context.Context, transaction *ent.Tx, draft *ent.MailDraft) (mailboxEntDraftAttachmentMaterializerLockedAttachments, MailboxDraftRepositoryErrorKind) {
	if ctx == nil || transaction == nil || draft == nil || draft.ID == uuid.Nil || draft.MailAccountID == uuid.Nil || draft.DraftClientID == uuid.Nil {
		return mailboxEntDraftAttachmentMaterializerLockedAttachments{}, MailboxDraftRepositoryErrorInvalid
	}
	attachments, err := transaction.MailAttachment.Query().
		Where(mailattachment.Or(mailattachment.MailDraftIDEQ(draft.ID), mailattachment.PendingMailDraftIDEQ(draft.ID))).
		Order(ent.Asc(mailattachment.FieldID)).
		Limit(MaxEngineAttachmentItems*2 + 1).
		ForUpdate().
		All(ctx)
	if err != nil {
		return mailboxEntDraftAttachmentMaterializerLockedAttachments{}, MailboxDraftRepositoryErrorUnavailable
	}
	if len(attachments) > MaxEngineAttachmentItems*2 {
		return mailboxEntDraftAttachmentMaterializerLockedAttachments{}, MailboxDraftRepositoryErrorStateConflict
	}

	locked := mailboxEntDraftAttachmentMaterializerLockedAttachments{
		attachments: make(map[uuid.UUID]*ent.MailAttachment, len(attachments)),
		objects:     make(map[uuid.UUID]*ent.MailObject, len(attachments)),
	}
	seenAttachments := make(map[uuid.UUID]struct{}, len(attachments))
	objectIDs := make([]uuid.UUID, 0, len(attachments))
	seenObjects := make(map[uuid.UUID]struct{}, len(attachments))
	for _, attachment := range attachments {
		if attachment == nil || attachment.ID == uuid.Nil {
			return mailboxEntDraftAttachmentMaterializerLockedAttachments{}, MailboxDraftRepositoryErrorStateConflict
		}
		if _, exists := seenAttachments[attachment.ID]; exists {
			return mailboxEntDraftAttachmentMaterializerLockedAttachments{}, MailboxDraftRepositoryErrorStateConflict
		}
		seenAttachments[attachment.ID] = struct{}{}
		confirmed := attachment.MailDraftID != nil && *attachment.MailDraftID == draft.ID
		pending := attachment.PendingMailDraftID != nil && *attachment.PendingMailDraftID == draft.ID
		if (!confirmed && !pending) || !mailboxEntDraftReadAttachmentValid(attachment, draft.ID, draft.DraftClientID, confirmed, pending) {
			return mailboxEntDraftAttachmentMaterializerLockedAttachments{}, MailboxDraftRepositoryErrorStateConflict
		}
		locked.attachments[attachment.ID] = attachment
		if confirmed {
			locked.state.ConfirmedAttachmentIDs = append(locked.state.ConfirmedAttachmentIDs, attachment.ID)
		}
		if pending {
			locked.state.PendingAttachmentIDs = append(locked.state.PendingAttachmentIDs, attachment.ID)
		}
		objectID := *attachment.StorageObjectID
		if _, exists := seenObjects[objectID]; !exists {
			seenObjects[objectID] = struct{}{}
			objectIDs = append(objectIDs, objectID)
		}
	}
	if len(locked.state.ConfirmedAttachmentIDs) > MaxEngineAttachmentItems || len(locked.state.PendingAttachmentIDs) > MaxEngineAttachmentItems {
		return mailboxEntDraftAttachmentMaterializerLockedAttachments{}, MailboxDraftRepositoryErrorStateConflict
	}
	sort.Slice(locked.state.ConfirmedAttachmentIDs, func(left, right int) bool { return syncLessUUID(locked.state.ConfirmedAttachmentIDs[left], locked.state.ConfirmedAttachmentIDs[right]) })
	sort.Slice(locked.state.PendingAttachmentIDs, func(left, right int) bool { return syncLessUUID(locked.state.PendingAttachmentIDs[left], locked.state.PendingAttachmentIDs[right]) })
	sort.Slice(objectIDs, func(left, right int) bool { return syncLessUUID(objectIDs[left], objectIDs[right]) })
	for _, objectID := range objectIDs {
		object, err := transaction.MailObject.Query().Where(mailobject.IDEQ(objectID)).ForUpdate().Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return mailboxEntDraftAttachmentMaterializerLockedAttachments{}, MailboxDraftRepositoryErrorStateConflict
			}
			return mailboxEntDraftAttachmentMaterializerLockedAttachments{}, MailboxDraftRepositoryErrorUnavailable
		}
		if object == nil || object.ID != objectID {
			return mailboxEntDraftAttachmentMaterializerLockedAttachments{}, MailboxDraftRepositoryErrorStateConflict
		}
		locked.objects[objectID] = object
	}
	return locked, ""
}

func mailboxEntDraftAttachmentMaterializerRecordMatches(record MailboxDraftRecord, request mailboxDraftAttachmentMaterializationRequest) bool {
	if record.ID != request.DraftID || record.AccountID != request.AccountID || record.DraftClientID != request.DraftClientID || record.Revision != request.Revision || record.LastMutationID == nil || *record.LastMutationID != request.MutationID || record.Status != MailboxDraftStatusSaving {
		return false
	}
	switch request.WorkKind {
	case MailboxDraftWorkCreate:
		return record.PendingMutationKind != nil && *record.PendingMutationKind == MailboxDraftMutationCreate && record.PendingCreateAttachmentPhase != nil && *record.PendingCreateAttachmentPhase == MailboxDraftCreateAttachmentPhaseUpload && record.PendingSave == nil && mailboxEntDraftEffectiveStatesEqual(record.Confirmed, request.EffectiveState)
	case MailboxDraftWorkSave:
		return record.PendingMutationKind != nil && *record.PendingMutationKind == MailboxDraftMutationSave && record.PendingCreateAttachmentPhase == nil && record.PendingSave != nil && mailboxEntDraftEffectiveStatesEqual(record.PendingSave.EffectiveState, request.EffectiveState)
	default:
		return false
	}
}

func mailboxEntDraftAttachmentMaterializerSpecs(locked mailboxEntDraftAttachmentMaterializerLockedAttachments, request mailboxDraftAttachmentMaterializationRequest, databaseNow time.Time, objectStore MailboxObjectStore) ([]mailboxDraftAttachmentSourceSpec, MailboxDraftRepositoryErrorKind) {
	if !mailboxEntValidTime(databaseNow) || mailboxDraftAttachmentMaterializerDependencyNil(objectStore) || len(request.EffectiveState.AttachmentIDs) == 0 || len(request.EffectiveState.AttachmentIDs) > MaxEngineAttachmentItems {
		return nil, MailboxDraftRepositoryErrorInvalid
	}
	specs := make([]mailboxDraftAttachmentSourceSpec, len(request.EffectiveState.AttachmentIDs))
	for index, attachmentID := range request.EffectiveState.AttachmentIDs {
		attachment := locked.attachments[attachmentID]
		if attachment == nil || !mailboxEntDraftAttachmentMaterializerAttachmentMatches(attachment, request) {
			return nil, MailboxDraftRepositoryErrorAttachmentConflict
		}
		object := locked.objects[*attachment.StorageObjectID]
		if !mailboxEntDraftAttachmentMaterializerObjectMatches(object, attachment, request, databaseNow) {
			return nil, MailboxDraftRepositoryErrorAttachmentConflict
		}
		var digest [sha256.Size]byte
		copy(digest[:], attachment.SHA256)
		contentID := ""
		if attachment.ContentID != nil {
			contentID = *attachment.ContentID
		}
		locator := MailboxObjectLocator{Bucket: object.Bucket, Key: object.ObjectKey}
		expectedSizeBytes := object.ExpectedSizeBytes
		specs[index] = mailboxDraftAttachmentSourceSpec{
			Attachment: EngineDraftAttachment{
				AttachmentKey: attachment.ID.String(),
				ContentID:     contentID,
				Filename:      attachment.Filename,
				MediaType:     attachment.MediaType,
				Disposition:   EngineAttachmentDisposition(attachment.Disposition),
				SHA256:        digest,
			},
			SizeBytes:  expectedSizeBytes,
			ScanStatus: AttachmentScanStatusClean,
			Open:       mailboxEntDraftAttachmentMaterializerOpen(objectStore, locator, expectedSizeBytes),
		}
	}
	return specs, ""
}

func mailboxEntDraftAttachmentMaterializerAttachmentMatches(attachment *ent.MailAttachment, request mailboxDraftAttachmentMaterializationRequest) bool {
	if attachment == nil || attachment.ID == uuid.Nil || attachment.DraftClientID == nil || *attachment.DraftClientID != request.DraftClientID || attachment.StorageObjectID == nil || *attachment.StorageObjectID == uuid.Nil || attachment.MailMessageID != nil || attachment.EngineBlobID != nil || attachment.ScanStatus != mailattachment.ScanStatus("clean") || len(attachment.SHA256) != sha256.Size || attachment.SizeBytes < 1 || attachment.SizeBytes > MaxEngineAttachmentBytes || !mailboxEntApplyAttachmentOptionalTextValid(attachment.ContentID) || !mailboxEntApplyAttachmentTextValid(attachment.Filename, false) || !mailboxEntApplyAttachmentTextValid(attachment.MediaType, false) || !mailboxEntApplyAttachmentDispositionValid(attachment.Disposition) {
		return false
	}
	if attachment.Disposition == mailattachment.Disposition(EngineAttachmentDispositionInline) && attachment.ContentID == nil {
		return false
	}
	switch request.WorkKind {
	case MailboxDraftWorkCreate:
		return attachment.MailDraftID != nil && *attachment.MailDraftID == request.DraftID && attachment.PendingMailDraftID == nil
	case MailboxDraftWorkSave:
		return attachment.PendingMailDraftID != nil && *attachment.PendingMailDraftID == request.DraftID && (attachment.MailDraftID == nil || *attachment.MailDraftID == request.DraftID)
	default:
		return false
	}
}

func mailboxEntDraftAttachmentMaterializerObjectMatches(object *ent.MailObject, attachment *ent.MailAttachment, request mailboxDraftAttachmentMaterializationRequest, databaseNow time.Time) bool {
	return object != nil && attachment != nil && attachment.StorageObjectID != nil && object.ID == *attachment.StorageObjectID && object.OrganizationID == request.OrganizationID && object.MailAccountID == request.AccountID && object.UploaderUserID == request.ActorID && object.DraftClientID == request.DraftClientID && object.UploadState == mailobject.UploadState("confirmed") && object.ExpectedSizeBytes == attachment.SizeBytes && object.ActualSizeBytes != nil && *object.ActualSizeBytes == object.ExpectedSizeBytes && object.Filename == attachment.Filename && object.MediaType == attachment.MediaType && len(object.SHA256) == sha256.Size && bytes.Equal(object.SHA256, attachment.SHA256) && object.NextDeleteAt == nil && object.DeleteLeaseToken == nil && object.DeleteLeaseUntil == nil && mailboxEntValidTime(object.ExpiresAt) && mailboxEntValidTime(object.UploadDrainUntil) && object.UploadDrainUntil.Equal(object.ExpiresAt.Add(mailboxAttachmentPresignObjectDrainLifetime)) && !databaseNow.Before(object.UploadDrainUntil) && object.ObjectKey == mailboxAttachmentPresignObjectKey(request.OrganizationID, request.AccountID, object.ID) && validMailboxObjectLocator(MailboxObjectLocator{Bucket: object.Bucket, Key: object.ObjectKey})
}

func mailboxEntDraftAttachmentMaterializerOpen(objectStore MailboxObjectStore, locator MailboxObjectLocator, expectedSizeBytes int64) mailboxDraftAttachmentOpen {
	return func(ctx context.Context) (io.ReadCloser, error) {
		if ctx == nil || mailboxDraftAttachmentMaterializerDependencyNil(objectStore) || !validMailboxObjectLocator(locator) || expectedSizeBytes < 1 || expectedSizeBytes > MaxEngineAttachmentBytes {
			return nil, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
		}
		stream, err := objectStore.Open(ctx, MailboxObjectReadRequest{Object: locator})
		if err != nil {
			return nil, mailboxEntDraftAttachmentMaterializerObjectStoreError(err)
		}
		if !validMailboxObjectReadStream(stream) || stream.ContentLength != expectedSizeBytes {
			if stream.Body != nil {
				_ = stream.Body.Close()
			}
			return nil, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		return stream.Body, nil
	}
}

func mailboxEntDraftAttachmentMaterializerObjectStoreError(err error) error {
	if kind, ok := MailboxObjectStoreErrorKindOf(err); ok && kind == MailboxObjectStoreErrorNotFound {
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorAttachmentConflict)
	}
	return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
}
