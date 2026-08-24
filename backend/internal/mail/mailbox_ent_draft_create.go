package mail

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/mailattachment"
	"github.com/valerii/hub/backend/ent/maildraft"
	"github.com/valerii/hub/backend/ent/mailidentity"
	"github.com/valerii/hub/backend/ent/mailmessage"
	"github.com/valerii/hub/backend/ent/mailobject"
)

// ReserveMailboxDraftCreate persists or recovers one draft-create operation under the exact acquired fence.
func (repository *MailboxEntRepository) ReserveMailboxDraftCreate(ctx context.Context, request MailboxDraftCreateRequest) (MailboxDraftReservation, error) {
	if repository == nil || repository.client == nil || ctx == nil {
		return MailboxDraftReservation{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	if kind := mailboxEntDraftCreateRequestKind(request); kind != "" {
		return MailboxDraftReservation{}, NewMailboxDraftRepositoryError(kind)
	}
	engineDraftKey, ok := mailboxEntDraftCreateKey(request.DraftID)
	if !ok {
		return MailboxDraftReservation{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}

	var result MailboxDraftReservation
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, databaseWallClock, err := mailboxEntDraftCreateLockAccount(ctx, transaction, request.Fence)
		if err != nil {
			return err
		}
		if kind := mailboxEntDraftCreateAccountKind(account); kind != "" {
			return mailboxEntDraftCreateReleaseError(ctx, account, request.Fence, databaseWallClock, kind)
		}
		rows, err := mailboxEntDraftCreateLockIdentityRows(ctx, transaction, request)
		if err != nil {
			return err
		}
		if err := mailboxEntDraftCreateLockConfirmedIdentity(ctx, transaction, request.Fence.AccountID, request.Confirmed.FromIdentityID); err != nil {
			return mailboxEntDraftCreateReleaseForError(ctx, account, request.Fence, databaseWallClock, err)
		}
		if err := mailboxEntDraftCreateLockReferences(ctx, transaction, request.Fence.AccountID, request.Confirmed); err != nil {
			return mailboxEntDraftCreateReleaseForError(ctx, account, request.Fence, databaseWallClock, err)
		}
		if len(rows) != 0 {
			return mailboxEntDraftCreateClassifyExisting(ctx, transaction, account, databaseWallClock, request, engineDraftKey, rows, &result)
		}
		if err := mailboxEntDraftCreateLockAttachments(ctx, transaction, request.Fence.AccountID, request.DraftClientID, request.Confirmed.AttachmentIDs); err != nil {
			return mailboxEntDraftCreateReleaseForError(ctx, account, request.Fence, databaseWallClock, err)
		}

		draftCreate := transaction.MailDraft.Create().
			SetID(request.DraftID).
			SetMailAccountID(request.Fence.AccountID).
			SetFromIdentityID(request.Confirmed.FromIdentityID).
			SetDraftClientID(request.DraftClientID).
			SetEngineDraftKey(engineDraftKey).
			SetRevision(1).
			SetStatus(maildraft.StatusSaving).
			SetCreateIdempotencyKeyHash(append([]byte(nil), request.CreateIdempotencyKeyHash[:]...)).
			SetCreateRequestHash(append([]byte(nil), request.CreateRequestHash[:]...)).
			SetLastMutationID(request.MutationID).
			SetPendingMutationKind(maildraft.PendingMutationKind("create")).
			SetCreatedAt(databaseWallClock).
			SetUpdatedAt(databaseWallClock)
		if request.Confirmed.ReplyToMessageID != nil {
			draftCreate.SetReplyToMessageID(*request.Confirmed.ReplyToMessageID)
		}
		if request.Confirmed.ForwardMessageID != nil {
			draftCreate.SetForwardMessageID(*request.Confirmed.ForwardMessageID)
		}
		if len(request.Confirmed.AttachmentIDs) != 0 {
			draftCreate.AddAttachmentIDs(request.Confirmed.AttachmentIDs...)
		}
		draft, err := draftCreate.Save(ctx)
		if err != nil {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		record, ok := mailboxEntDraftReadRecord(draft, mailboxEntDraftAttachmentState{ConfirmedAttachmentIDs: append([]uuid.UUID(nil), request.Confirmed.AttachmentIDs...)})
		if !ok {
			return mailboxEntDraftCreateReleaseError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
		}
		work := mailboxEntDraftCreateWork(record, MailboxDraftWorkCreate)
		if work == nil {
			return mailboxEntDraftCreateReleaseError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
		}
		fence := mailboxEntDraftCreateFenceCopy(request.Fence)
		result = MailboxDraftReservation{Kind: MailboxDraftReservationDispatch, Draft: record, Work: work, Fence: &fence}
		return nil
	})
	if err != nil {
		return MailboxDraftReservation{}, mailboxEntDraftRepositoryError(err)
	}
	return result, nil
}

func mailboxEntDraftCreateRequestKind(request MailboxDraftCreateRequest) MailboxDraftRepositoryErrorKind {
	if !mailboxEntEngineCommandFenceValid(request.Fence) || request.DraftID == uuid.Nil || request.DraftClientID == uuid.Nil || request.MutationID == uuid.Nil || !mailboxEntDraftCreateHashPresent(request.CreateIdempotencyKeyHash) || !mailboxEntDraftCreateHashPresent(request.CreateRequestHash) {
		return MailboxDraftRepositoryErrorInvalid
	}
	return mailboxEntDraftEffectiveStateKind(request.Confirmed)
}

func mailboxEntDraftEffectiveStateKind(state MailboxDraftEffectiveState) MailboxDraftRepositoryErrorKind {
	if state.FromIdentityID == uuid.Nil || !mailboxEntDraftCreateUUIDPointerValid(state.ReplyToMessageID) || !mailboxEntDraftCreateUUIDPointerValid(state.ForwardMessageID) {
		return MailboxDraftRepositoryErrorInvalid
	}
	if len(state.AttachmentIDs) > MaxEngineAttachmentItems {
		return MailboxDraftRepositoryErrorAttachmentLimit
	}
	if !mailboxEntDraftReadUUIDSliceValid(state.AttachmentIDs) {
		return MailboxDraftRepositoryErrorInvalid
	}
	return ""
}

func mailboxEntDraftCreateHashPresent(value [sha256.Size]byte) bool {
	var zero [sha256.Size]byte
	return value != zero
}

func mailboxEntDraftCreateUUIDPointerValid(value *uuid.UUID) bool { return value == nil || *value != uuid.Nil }

func mailboxEntDraftCreateKey(draftID uuid.UUID) (string, bool) {
	if draftID == uuid.Nil {
		return "", false
	}
	material := make([]byte, 0, len("hub-mail-draft-v1\x00")+len(draftID))
	material = append(material, "hub-mail-draft-v1\x00"...)
	material = append(material, draftID[:]...)
	sum := sha256.Sum256(material)
	key := "d1_" + base64.RawURLEncoding.EncodeToString(sum[:])
	return key, syncValidOpaque(key, MaxEngineOpaqueIDBytes)
}

func mailboxEntDraftCreateLockAccount(ctx context.Context, transaction *ent.Tx, fence MailboxEngineCommandFence) (*ent.MailAccount, time.Time, error) {
	if ctx == nil || transaction == nil || !mailboxEntEngineCommandFenceValid(fence) {
		return nil, time.Time{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	account, err := transaction.MailAccount.Query().Where(mailaccount.IDEQ(fence.AccountID)).ForUpdate().Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, time.Time{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorNotFound)
		}
		return nil, time.Time{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
	}
	if account == nil || account.ID != fence.AccountID {
		return nil, time.Time{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	databaseWallClock, err := mailboxEntDatabaseWallClock(ctx, transaction)
	if err != nil {
		return nil, time.Time{}, mailboxEntDraftRepositoryError(err)
	}
	if !mailboxEntDraftCreateFenceMatches(account, fence, databaseWallClock) {
		return nil, time.Time{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorFenceLost)
	}
	return account, databaseWallClock, nil
}

func mailboxEntDraftCreateFenceMatches(account *ent.MailAccount, fence MailboxEngineCommandFence, databaseWallClock time.Time) bool {
	return account != nil && account.ID == fence.AccountID && account.MutationLeaseToken != nil && *account.MutationLeaseToken == fence.LeaseID && account.MutationLeaseUntil != nil && mailboxEntValidTime(*account.MutationLeaseUntil) && account.MutationLeaseUntil.Equal(fence.LeaseUntil) && account.MutationEpoch == fence.MutationEpoch && mailboxEntValidTime(databaseWallClock) && account.MutationLeaseUntil.After(databaseWallClock)
}

func mailboxEntDraftCreateAccountKind(account *ent.MailAccount) MailboxDraftRepositoryErrorKind {
	if account == nil || account.ID == uuid.Nil {
		return MailboxDraftRepositoryErrorInvalid
	}
	switch account.Status {
	case mailaccount.StatusDeleting:
		return MailboxDraftRepositoryErrorAccountDeleting
	case mailaccount.StatusActive:
	default:
		return MailboxDraftRepositoryErrorStateConflict
	}
	if !account.DesiredEnabled || account.EngineAccountID == nil || !syncValidOpaque(*account.EngineAccountID, MaxEngineOpaqueIDBytes) || account.SyncEpoch < 0 || account.MutationEpoch < 1 || account.SyncLeaseEpoch < 1 || !mailboxEntValidTime(account.SyncScheduleUpdatedAt) {
		return MailboxDraftRepositoryErrorInvalid
	}
	if _, _, valid := mailboxEntSyncLeaseFence(account); !valid {
		return MailboxDraftRepositoryErrorInvalid
	}
	return ""
}

// mailboxEntDraftCreateLockIdentityRows locks the draft identity matrix before all child locks.
func mailboxEntDraftCreateLockIdentityRows(ctx context.Context, transaction *ent.Tx, request MailboxDraftCreateRequest) ([]*ent.MailDraft, error) {
	if ctx == nil || transaction == nil {
		return nil, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	rows, err := transaction.MailDraft.Query().
		Where(maildraft.MailAccountIDEQ(request.Fence.AccountID), maildraft.Or(maildraft.IDEQ(request.DraftID), maildraft.DraftClientIDEQ(request.DraftClientID), maildraft.CreateIdempotencyKeyHashEQ(request.CreateIdempotencyKeyHash[:]))).
		Order(ent.Asc(maildraft.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return nil, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
	}
	return rows, nil
}

func mailboxEntDraftCreateLockConfirmedIdentity(ctx context.Context, transaction *ent.Tx, accountID, identityID uuid.UUID) error {
	if ctx == nil || transaction == nil || accountID == uuid.Nil || identityID == uuid.Nil {
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	identity, err := transaction.MailIdentity.Query().
		Where(mailidentity.IDEQ(identityID), mailidentity.MailAccountIDEQ(accountID), mailidentity.StatusEQ(mailidentity.Status("active")), mailidentity.SendAllowedEQ(true), mailidentity.KindIn(mailidentity.Kind("primary"), mailidentity.Kind("alias"))).
		ForUpdate().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorStateConflict)
		}
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
	}
	if identity == nil || identity.ID != identityID || identity.MailAccountID != accountID || !identity.SendAllowed || identity.Status != mailidentity.Status("active") || identity.Kind != mailidentity.Kind("primary") && identity.Kind != mailidentity.Kind("alias") {
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorStateConflict)
	}
	return nil
}

func mailboxEntDraftCreateLockReferences(ctx context.Context, transaction *ent.Tx, accountID uuid.UUID, state MailboxDraftEffectiveState) error {
	if ctx == nil || transaction == nil || accountID == uuid.Nil {
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	ids := make([]uuid.UUID, 0, 2)
	if state.ReplyToMessageID != nil {
		ids = append(ids, *state.ReplyToMessageID)
	}
	if state.ForwardMessageID != nil {
		ids = append(ids, *state.ForwardMessageID)
	}
	sort.Slice(ids, func(left, right int) bool { return syncLessUUID(ids[left], ids[right]) })
	for index, id := range ids {
		if id == uuid.Nil || index > 0 && id == ids[index-1] {
			if id == uuid.Nil {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
			}
			continue
		}
		message, err := transaction.MailMessage.Query().Where(mailmessage.IDEQ(id), mailmessage.MailAccountIDEQ(accountID), mailmessage.DeletedAtIsNil()).ForUpdate().Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorStateConflict)
			}
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		if message == nil || message.ID != id || message.MailAccountID != accountID || message.DeletedAt != nil {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorStateConflict)
		}
	}
	return nil
}

// mailboxEntDraftCreateLockAttachments locks every attachment before any object, each class in raw UUID order.
func mailboxEntDraftCreateLockAttachments(ctx context.Context, transaction *ent.Tx, accountID, draftClientID uuid.UUID, attachmentIDs []uuid.UUID) error {
	if ctx == nil || transaction == nil || accountID == uuid.Nil || draftClientID == uuid.Nil {
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	attachments := make(map[uuid.UUID]*ent.MailAttachment, len(attachmentIDs))
	objectIDs := make([]uuid.UUID, 0, len(attachmentIDs))
	objectsSeen := make(map[uuid.UUID]struct{}, len(attachmentIDs))
	for _, attachmentID := range attachmentIDs {
		attachment, err := transaction.MailAttachment.Query().Where(mailattachment.IDEQ(attachmentID)).ForUpdate().Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorAttachmentConflict)
			}
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		if !mailboxEntDraftCreateAttachmentValid(attachment, draftClientID) {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorAttachmentConflict)
		}
		attachments[attachmentID] = attachment
		objectID := *attachment.StorageObjectID
		if _, exists := objectsSeen[objectID]; !exists {
			objectsSeen[objectID] = struct{}{}
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
	for _, attachmentID := range attachmentIDs {
		attachment := attachments[attachmentID]
		if !mailboxEntDraftCreateStorageObjectValid(objects[*attachment.StorageObjectID], accountID, draftClientID, attachment) {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorAttachmentConflict)
		}
	}
	return nil
}

func mailboxEntDraftCreateAttachmentValid(attachment *ent.MailAttachment, draftClientID uuid.UUID) bool {
	if attachment == nil || attachment.ID == uuid.Nil || attachment.MailMessageID != nil || attachment.MailDraftID != nil || attachment.PendingMailDraftID != nil || attachment.DraftClientID == nil || *attachment.DraftClientID != draftClientID || attachment.StorageObjectID == nil || *attachment.StorageObjectID == uuid.Nil || attachment.EngineBlobID != nil || attachment.SizeBytes < 1 || attachment.SizeBytes > MaxEngineAttachmentBytes || attachment.ScanStatus != mailattachment.ScanStatus("clean") || len(attachment.SHA256) != AttachmentSHA256Bytes || !mailboxEntApplyAttachmentOptionalTextValid(attachment.ContentID) || !mailboxEntApplyAttachmentTextValid(attachment.Filename, false) || !mailboxEntApplyAttachmentTextValid(attachment.MediaType, false) || !mailboxEntApplyAttachmentDispositionValid(attachment.Disposition) {
		return false
	}
	return attachment.Disposition != mailattachment.Disposition(EngineAttachmentDispositionInline) || attachment.ContentID != nil
}

func mailboxEntDraftCreateStorageObjectValid(object *ent.MailObject, accountID, draftClientID uuid.UUID, attachment *ent.MailAttachment) bool {
	return mailboxEntDraftReadStorageObjectValid(object, accountID, draftClientID, attachment)
}

func mailboxEntDraftCreateClassifyExisting(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, databaseWallClock time.Time, request MailboxDraftCreateRequest, engineDraftKey string, rows []*ent.MailDraft, result *MailboxDraftReservation) error {
	if result == nil || len(rows) != 1 {
		return mailboxEntDraftCreateReleaseError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorIdempotencyConflict)
	}
	draft := rows[0]
	if !mailboxEntDraftReadRowValid(draft, request.Fence.AccountID, request.DraftID) {
		return mailboxEntDraftCreateReleaseError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
	}
	attachments, kind := mailboxEntDraftReadAttachmentState(ctx, transaction, draft, MailboxDraftRepositoryErrorStateConflict)
	if kind != "" {
		return mailboxEntDraftCreateReleaseError(ctx, account, request.Fence, databaseWallClock, kind)
	}
	record, ok := mailboxEntDraftReadRecord(draft, attachments)
	if !ok || !bytes.Equal(draft.CreateIdempotencyKeyHash, request.CreateIdempotencyKeyHash[:]) || !mailboxEntDraftCreateSameRequest(record, request, engineDraftKey) {
		return mailboxEntDraftCreateReleaseError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorIdempotencyConflict)
	}
	switch record.Status {
	case MailboxDraftStatusActive, MailboxDraftStatusSubmitting, MailboxDraftStatusSubmitted, MailboxDraftStatusDeleted:
		if err := mailboxEntDraftCreateRelease(ctx, account, request.Fence, databaseWallClock); err != nil {
			return err
		}
		*result = MailboxDraftReservation{Kind: MailboxDraftReservationReplay, Draft: mailboxEntDraftRecordCopy(record)}
		return nil
	case MailboxDraftStatusSaving, MailboxDraftStatusUnknown:
		if record.PendingMutationKind == nil || *record.PendingMutationKind != MailboxDraftMutationCreate || record.LastMutationID == nil || record.EngineMessageID != nil || record.EngineVersion != nil || record.PendingSave != nil {
			return mailboxEntDraftCreateReleaseError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
		}
		work := mailboxEntDraftCreateWork(record, MailboxDraftWorkReconcileCreate)
		if work == nil {
			return mailboxEntDraftCreateReleaseError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
		}
		fence := mailboxEntDraftCreateFenceCopy(request.Fence)
		*result = MailboxDraftReservation{Kind: MailboxDraftReservationReconcileRequired, Draft: mailboxEntDraftRecordCopy(record), Work: work, Fence: &fence}
		return nil
	default:
		return mailboxEntDraftCreateReleaseError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
	}
}

func mailboxEntDraftCreateSameRequest(record MailboxDraftRecord, request MailboxDraftCreateRequest, engineDraftKey string) bool {
	return record.ID == request.DraftID && record.AccountID == request.Fence.AccountID && record.DraftClientID == request.DraftClientID && record.EngineDraftKey == engineDraftKey && record.CreateRequestHash == request.CreateRequestHash && mailboxEntDraftEffectiveStatesEqual(record.Confirmed, request.Confirmed)
}

func mailboxEntDraftEffectiveStatesEqual(left, right MailboxDraftEffectiveState) bool {
	return left.FromIdentityID == right.FromIdentityID && mailboxEntDraftUUIDPointersEqual(left.ReplyToMessageID, right.ReplyToMessageID) && mailboxEntDraftUUIDPointersEqual(left.ForwardMessageID, right.ForwardMessageID) && mailboxEntDraftReservationUUIDSetsEqual(left.AttachmentIDs, right.AttachmentIDs)
}

func mailboxEntDraftUUIDPointersEqual(left, right *uuid.UUID) bool {
	return left == nil && right == nil || left != nil && right != nil && *left == *right
}

func mailboxEntDraftCreateReleaseForError(ctx context.Context, account *ent.MailAccount, fence MailboxEngineCommandFence, databaseWallClock time.Time, err error) error {
	if repositoryError, ok := err.(*MailboxDraftRepositoryError); ok && repositoryError != nil && repositoryError.Kind() == MailboxDraftRepositoryErrorUnavailable {
		return err
	}
	return mailboxEntDraftCreateReleaseError(ctx, account, fence, databaseWallClock, mailboxEntDraftCreateErrorKind(err))
}

func mailboxEntDraftCreateErrorKind(err error) MailboxDraftRepositoryErrorKind {
	if repositoryError, ok := err.(*MailboxDraftRepositoryError); ok && repositoryError != nil {
		return repositoryError.Kind()
	}
	return MailboxDraftRepositoryErrorUnavailable
}

func mailboxEntDraftCreateReleaseError(ctx context.Context, account *ent.MailAccount, fence MailboxEngineCommandFence, databaseWallClock time.Time, kind MailboxDraftRepositoryErrorKind) error {
	if err := mailboxEntDraftCreateRelease(ctx, account, fence, databaseWallClock); err != nil {
		return err
	}
	return NewMailboxDraftRepositoryError(kind)
}

func mailboxEntDraftCreateRelease(ctx context.Context, account *ent.MailAccount, fence MailboxEngineCommandFence, databaseWallClock time.Time) error {
	if !mailboxEntDraftCreateFenceMatches(account, fence, databaseWallClock) {
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorFenceLost)
	}
	nextMutationEpoch, ok := mailboxEntNextEpoch(account.MutationEpoch)
	if !ok {
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	if err := account.Update().ClearMutationLeaseToken().ClearMutationLeaseUntil().SetMutationEpoch(nextMutationEpoch).SetSyncScheduleUpdatedAt(databaseWallClock).Exec(ctx); err != nil {
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
	}
	return nil
}

func mailboxEntDraftCreateWork(record MailboxDraftRecord, kind MailboxDraftWorkKind) *MailboxDraftWork {
	if record.LastMutationID == nil || record.ID == uuid.Nil || record.AccountID == uuid.Nil || record.Revision < 1 || !syncValidOpaque(record.EngineDraftKey, MaxEngineOpaqueIDBytes) {
		return nil
	}
	effectiveState := mailboxEntDraftEffectiveStateCopy(record.Confirmed)
	requestHash := record.CreateRequestHash
	if record.PendingMutationKind != nil && *record.PendingMutationKind == MailboxDraftMutationSave {
		if record.PendingSave == nil {
			return nil
		}
		effectiveState = mailboxEntDraftEffectiveStateCopy(record.PendingSave.EffectiveState)
		requestHash = record.PendingSave.RequestHash
	}
	if record.PendingMutationKind != nil && *record.PendingMutationKind == MailboxDraftMutationDelete {
		requestHash = [sha256.Size]byte{}
	}
	return &MailboxDraftWork{Kind: kind, EngineDraftKey: record.EngineDraftKey, EngineMessageID: mailboxEntStringPointer(record.EngineMessageID), EngineVersion: mailboxEntStringPointer(record.EngineVersion), MutationID: *record.LastMutationID, Revision: record.Revision, EffectiveState: effectiveState, RequestHash: requestHash, ConfirmedMutationID: mailboxEntDraftReadUUIDPointerCopy(record.ConfirmedMutationID)}
}

func mailboxEntDraftCreateFenceCopy(fence MailboxEngineCommandFence) MailboxEngineCommandFence {
	return MailboxEngineCommandFence{AccountID: fence.AccountID, LeaseID: fence.LeaseID, MutationEpoch: fence.MutationEpoch, LeaseUntil: fence.LeaseUntil}
}

func mailboxEntDraftRecordCopy(record MailboxDraftRecord) MailboxDraftRecord {
	return MailboxDraftRecord{ID: record.ID, AccountID: record.AccountID, DraftClientID: record.DraftClientID, EngineDraftKey: record.EngineDraftKey, EngineMessageID: mailboxEntStringPointer(record.EngineMessageID), EngineVersion: mailboxEntStringPointer(record.EngineVersion), Revision: record.Revision, Status: record.Status, Confirmed: mailboxEntDraftEffectiveStateCopy(record.Confirmed), PendingSave: mailboxEntDraftPendingSaveCopy(record.PendingSave), CreateRequestHash: record.CreateRequestHash, LastMutationID: mailboxEntDraftReadUUIDPointerCopy(record.LastMutationID), PendingMutationKind: mailboxEntDraftReadMutationKindPointerCopy(record.PendingMutationKind), ConfirmedMutationID: mailboxEntDraftReadUUIDPointerCopy(record.ConfirmedMutationID), UpdatedAt: mailboxEntDraftReadTimeCopy(record.UpdatedAt)}
}
