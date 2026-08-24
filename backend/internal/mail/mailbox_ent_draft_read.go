package mail

import (
	"bytes"
	"context"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/mailattachment"
	"github.com/valerii/hub/backend/ent/maildraft"
	"github.com/valerii/hub/backend/ent/mailobject"
)

type mailboxEntDraftAttachmentState struct {
	ConfirmedAttachmentIDs []uuid.UUID
	PendingAttachmentIDs   []uuid.UUID
}

// GetMailboxDraft reads one account-scoped durable draft before any engine I/O.
func (repository *MailboxEntRepository) GetMailboxDraft(ctx context.Context, request MailboxDraftGetRequest) (MailboxDraftGetResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || request.AccountID == uuid.Nil || request.DraftID == uuid.Nil {
		return MailboxDraftGetResult{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}

	var result MailboxDraftGetResult
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		if err := mailboxEntDraftReadLockAccount(ctx, transaction, request.AccountID); err != nil {
			return err
		}
		draft, err := transaction.MailDraft.Query().
			Where(maildraft.IDEQ(request.DraftID), maildraft.MailAccountIDEQ(request.AccountID)).
			ForUpdate().
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil
			}
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		if !mailboxEntDraftReadRowValid(draft, request.AccountID, request.DraftID) {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
		}
		attachments, kind := mailboxEntDraftReadAttachmentState(ctx, transaction, draft, MailboxDraftRepositoryErrorInvalid)
		if kind != "" {
			return NewMailboxDraftRepositoryError(kind)
		}
		record, ok := mailboxEntDraftReadRecord(draft, attachments)
		if !ok {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
		}
		result = MailboxDraftGetResult{Found: true, Draft: &record}
		return nil
	})
	if err != nil {
		return MailboxDraftGetResult{}, mailboxEntDraftRepositoryError(err)
	}
	return result, nil
}

// RecheckMailboxDraftRead confirms that an active draft still has the exact durable engine pair after engine I/O.
func (repository *MailboxEntRepository) RecheckMailboxDraftRead(ctx context.Context, request MailboxDraftReadRecheckRequest) (MailboxDraftReadRecheckResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || request.AccountID == uuid.Nil || request.DraftID == uuid.Nil || request.Revision < 1 || !syncValidOpaque(request.EngineMessageID, MaxEngineOpaqueIDBytes) || !syncValidOpaque(request.EngineVersion, MaxEngineOpaqueIDBytes) {
		return MailboxDraftReadRecheckResult{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}

	var result MailboxDraftReadRecheckResult
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		if err := mailboxEntDraftReadLockAccount(ctx, transaction, request.AccountID); err != nil {
			return err
		}
		draft, err := transaction.MailDraft.Query().
			Where(
				maildraft.IDEQ(request.DraftID),
				maildraft.MailAccountIDEQ(request.AccountID),
				maildraft.StatusEQ(maildraft.StatusActive),
				maildraft.RevisionEQ(request.Revision),
				maildraft.EngineMessageIDEQ(request.EngineMessageID),
				maildraft.EngineVersionEQ(request.EngineVersion),
			).
			ForUpdate().
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil
			}
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		if !mailboxEntDraftReadRowValid(draft, request.AccountID, request.DraftID) {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
		}
		attachments, kind := mailboxEntDraftReadAttachmentState(ctx, transaction, draft, MailboxDraftRepositoryErrorInvalid)
		if kind != "" {
			return NewMailboxDraftRepositoryError(kind)
		}
		if _, ok := mailboxEntDraftReadRecord(draft, attachments); !ok {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
		}
		result.Current = true
		return nil
	})
	if err != nil {
		return MailboxDraftReadRecheckResult{}, mailboxEntDraftRepositoryError(err)
	}
	return result, nil
}

func mailboxEntDraftReadLockAccount(ctx context.Context, transaction *ent.Tx, accountID uuid.UUID) error {
	if ctx == nil || transaction == nil || accountID == uuid.Nil {
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	account, err := transaction.MailAccount.Query().Where(mailaccount.IDEQ(accountID)).ForUpdate().Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorNotFound)
		}
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
	}
	if account == nil || account.ID != accountID {
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	switch account.Status {
	case mailaccount.StatusDeleting:
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorAccountDeleting)
	case mailaccount.StatusActive:
	case mailaccount.StatusProvisioning, mailaccount.StatusSuspended, mailaccount.StatusError:
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
	default:
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	if !account.DesiredEnabled || account.EngineAccountID == nil || !syncValidOpaque(*account.EngineAccountID, MaxEngineOpaqueIDBytes) || account.SyncEpoch < 0 || account.SyncLeaseEpoch < 1 || account.MutationEpoch < 0 {
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	if (account.MutationLeaseToken == nil) != (account.MutationLeaseUntil == nil) || (account.MutationLeaseToken != nil && *account.MutationLeaseToken == uuid.Nil) || (account.MutationLeaseUntil != nil && !mailboxEntValidTime(*account.MutationLeaseUntil)) {
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	if _, _, valid := mailboxEntSyncLeaseFence(account); !valid {
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	return nil
}

// mailboxEntDraftReadAttachmentState locks O union D in raw UUID order, then locks every
// referenced object in raw UUID order. Each graph is capped independently at 50 rows.
func mailboxEntDraftReadAttachmentState(ctx context.Context, transaction *ent.Tx, draft *ent.MailDraft, failureKind MailboxDraftRepositoryErrorKind) (mailboxEntDraftAttachmentState, MailboxDraftRepositoryErrorKind) {
	if ctx == nil || transaction == nil || draft == nil || draft.ID == uuid.Nil || draft.MailAccountID == uuid.Nil || draft.DraftClientID == uuid.Nil || failureKind == "" {
		return mailboxEntDraftAttachmentState{}, MailboxDraftRepositoryErrorInvalid
	}
	attachments, err := transaction.MailAttachment.Query().
		Where(mailattachment.Or(mailattachment.MailDraftIDEQ(draft.ID), mailattachment.PendingMailDraftIDEQ(draft.ID))).
		Order(ent.Asc(mailattachment.FieldID)).
		Limit(MaxEngineAttachmentItems*2 + 1).
		ForUpdate().
		All(ctx)
	if err != nil {
		return mailboxEntDraftAttachmentState{}, MailboxDraftRepositoryErrorUnavailable
	}
	if len(attachments) > MaxEngineAttachmentItems*2 {
		return mailboxEntDraftAttachmentState{}, failureKind
	}

	state := mailboxEntDraftAttachmentState{}
	seen := make(map[uuid.UUID]struct{}, len(attachments))
	objectIDs := make([]uuid.UUID, 0, len(attachments))
	objectsSeen := make(map[uuid.UUID]struct{}, len(attachments))
	attachmentByID := make(map[uuid.UUID]*ent.MailAttachment, len(attachments))
	for _, attachment := range attachments {
		if attachment == nil || attachment.ID == uuid.Nil {
			return mailboxEntDraftAttachmentState{}, failureKind
		}
		if _, exists := seen[attachment.ID]; exists {
			return mailboxEntDraftAttachmentState{}, failureKind
		}
		seen[attachment.ID] = struct{}{}
		confirmed := attachment.MailDraftID != nil && *attachment.MailDraftID == draft.ID
		pending := attachment.PendingMailDraftID != nil && *attachment.PendingMailDraftID == draft.ID
		if !confirmed && !pending || !mailboxEntDraftReadAttachmentValid(attachment, draft.ID, draft.DraftClientID, confirmed, pending) {
			return mailboxEntDraftAttachmentState{}, failureKind
		}
		attachmentByID[attachment.ID] = attachment
		if confirmed {
			state.ConfirmedAttachmentIDs = append(state.ConfirmedAttachmentIDs, attachment.ID)
		}
		if pending {
			state.PendingAttachmentIDs = append(state.PendingAttachmentIDs, attachment.ID)
		}
		objectID := *attachment.StorageObjectID
		if _, exists := objectsSeen[objectID]; !exists {
			objectsSeen[objectID] = struct{}{}
			objectIDs = append(objectIDs, objectID)
		}
	}
	if len(state.ConfirmedAttachmentIDs) > MaxEngineAttachmentItems || len(state.PendingAttachmentIDs) > MaxEngineAttachmentItems {
		return mailboxEntDraftAttachmentState{}, failureKind
	}
	sort.Slice(state.ConfirmedAttachmentIDs, func(left, right int) bool { return syncLessUUID(state.ConfirmedAttachmentIDs[left], state.ConfirmedAttachmentIDs[right]) })
	sort.Slice(state.PendingAttachmentIDs, func(left, right int) bool { return syncLessUUID(state.PendingAttachmentIDs[left], state.PendingAttachmentIDs[right]) })
	sort.Slice(objectIDs, func(left, right int) bool { return syncLessUUID(objectIDs[left], objectIDs[right]) })
	for _, objectID := range objectIDs {
		object, err := transaction.MailObject.Query().Where(mailobject.IDEQ(objectID)).ForUpdate().Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return mailboxEntDraftAttachmentState{}, failureKind
			}
			return mailboxEntDraftAttachmentState{}, MailboxDraftRepositoryErrorUnavailable
		}
		if object == nil || object.ID != objectID {
			return mailboxEntDraftAttachmentState{}, failureKind
		}
		for attachmentID, attachment := range attachmentByID {
			if *attachment.StorageObjectID == objectID && !mailboxEntDraftReadStorageObjectValid(object, draft.MailAccountID, draft.DraftClientID, attachment) {
				_ = attachmentID
				return mailboxEntDraftAttachmentState{}, failureKind
			}
		}
	}
	return mailboxEntDraftAttachmentState{
		ConfirmedAttachmentIDs: append([]uuid.UUID(nil), state.ConfirmedAttachmentIDs...),
		PendingAttachmentIDs:   append([]uuid.UUID(nil), state.PendingAttachmentIDs...),
	}, ""
}

func mailboxEntDraftReadAttachmentValid(attachment *ent.MailAttachment, draftID, draftClientID uuid.UUID, confirmed, pending bool) bool {
	if attachment == nil || attachment.ID == uuid.Nil || draftID == uuid.Nil || draftClientID == uuid.Nil || attachment.MailMessageID != nil || attachment.DraftClientID == nil || *attachment.DraftClientID != draftClientID || attachment.StorageObjectID == nil || *attachment.StorageObjectID == uuid.Nil || attachment.EngineBlobID != nil || attachment.CreatedAt.IsZero() || attachment.UpdatedAt.IsZero() || !mailboxEntApplyAttachmentOptionalTextValid(attachment.ContentID) || !mailboxEntApplyAttachmentTextValid(attachment.Filename, false) || !mailboxEntApplyAttachmentTextValid(attachment.MediaType, false) || attachment.SizeBytes < 1 || attachment.SizeBytes > MaxEngineAttachmentBytes || !mailboxEntApplyAttachmentDispositionValid(attachment.Disposition) || !mailboxEntApplyAttachmentScanStatusValid(attachment.ScanStatus) || (len(attachment.SHA256) != 0 && len(attachment.SHA256) != AttachmentSHA256Bytes) || ((attachment.ScanStatus == mailattachment.ScanStatus("clean") || attachment.ScanStatus == mailattachment.ScanStatus("rejected")) && len(attachment.SHA256) != AttachmentSHA256Bytes) {
		return false
	}
	if confirmed != (attachment.MailDraftID != nil && *attachment.MailDraftID == draftID) || pending != (attachment.PendingMailDraftID != nil && *attachment.PendingMailDraftID == draftID) {
		return false
	}
	if attachment.MailDraftID != nil && *attachment.MailDraftID != draftID || attachment.PendingMailDraftID != nil && *attachment.PendingMailDraftID != draftID {
		return false
	}
	return attachment.Disposition != mailattachment.Disposition(EngineAttachmentDispositionInline) || attachment.ContentID != nil
}

func mailboxEntDraftReadStorageObjectValid(object *ent.MailObject, accountID, draftClientID uuid.UUID, attachment *ent.MailAttachment) bool {
	return object != nil && attachment != nil && attachment.StorageObjectID != nil && object.ID == *attachment.StorageObjectID && object.MailAccountID == accountID && object.DraftClientID == draftClientID && object.UploadState == mailobject.UploadState("confirmed") && object.ActualSizeBytes != nil && *object.ActualSizeBytes == object.ExpectedSizeBytes && object.ExpectedSizeBytes == attachment.SizeBytes && len(object.SHA256) == AttachmentSHA256Bytes && bytes.Equal(object.SHA256, attachment.SHA256) && object.NextDeleteAt == nil && object.DeleteLeaseToken == nil && object.DeleteLeaseUntil == nil
}

func mailboxEntDraftReadRowValid(draft *ent.MailDraft, accountID, draftID uuid.UUID) bool {
	if draft == nil || draft.ID != draftID || draftID == uuid.Nil || draft.MailAccountID != accountID || accountID == uuid.Nil || draft.FromIdentityID == uuid.Nil || draft.DraftClientID == uuid.Nil || !syncValidOpaque(draft.EngineDraftKey, MaxEngineOpaqueIDBytes) || draft.Revision < 1 || !mailboxEntDraftReadHashValid(draft.CreateIdempotencyKeyHash) || !mailboxEntDraftReadHashValid(draft.CreateRequestHash) || !mailboxEntValidTime(draft.UpdatedAt) || !mailboxEntDraftReadUUIDPointerValid(draft.ReplyToMessageID) || !mailboxEntDraftReadUUIDPointerValid(draft.ForwardMessageID) || !mailboxEntDraftReadUUIDPointerValid(draft.PendingFromIdentityID) || !mailboxEntDraftReadUUIDPointerValid(draft.PendingReplyToMessageID) || !mailboxEntDraftReadUUIDPointerValid(draft.PendingForwardMessageID) || !mailboxEntDraftReadUUIDPointerValid(draft.LastMutationID) || !mailboxEntDraftReadUUIDPointerValid(draft.ConfirmedMutationID) || (len(draft.PendingSaveRequestHash) != 0 && !mailboxEntDraftReadHashValid(draft.PendingSaveRequestHash)) {
		return false
	}
	pairPresent, pairValid := mailboxEntDraftReadEnginePairValid(draft.EngineMessageID, draft.EngineVersion)
	if !pairValid {
		return false
	}
	pendingKind, pendingKindValid := mailboxEntDraftReadPendingKind(draft.PendingMutationKind)
	if !pendingKindValid || draft.ConfirmedMutationID != nil && (!pairPresent || draft.LastMutationID != nil && *draft.ConfirmedMutationID == *draft.LastMutationID) {
		return false
	}
	return mailboxEntDraftReadStateScalarValid(MailboxDraftStatus(draft.Status), pairPresent, draft.LastMutationID, pendingKind, draft.PendingFromIdentityID, draft.PendingReplyToMessageID, draft.PendingForwardMessageID, draft.PendingSaveRequestHash)
}

func mailboxEntDraftReadEnginePairValid(messageID, version *string) (bool, bool) {
	if messageID == nil || version == nil {
		return false, messageID == nil && version == nil
	}
	return true, syncValidOpaque(*messageID, MaxEngineOpaqueIDBytes) && syncValidOpaque(*version, MaxEngineOpaqueIDBytes)
}

func mailboxEntDraftReadPendingKind(value *maildraft.PendingMutationKind) (*MailboxDraftMutationKind, bool) {
	if value == nil {
		return nil, true
	}
	kind := MailboxDraftMutationKind(*value)
	switch kind {
	case MailboxDraftMutationCreate, MailboxDraftMutationSave, MailboxDraftMutationDelete:
		return &kind, true
	default:
		return nil, false
	}
}

func mailboxEntDraftReadStateScalarValid(status MailboxDraftStatus, enginePairPresent bool, mutationID *uuid.UUID, pendingKind *MailboxDraftMutationKind, pendingFrom, pendingReply, pendingForward *uuid.UUID, pendingHash []byte) bool {
	pendingScalarsPresent := pendingFrom != nil || pendingReply != nil || pendingForward != nil || pendingHash != nil
	pendingSaveValid := pendingFrom != nil && mailboxEntDraftReadHashValid(pendingHash)
	noPendingSave := !pendingScalarsPresent
	mutationPresent := mutationID != nil
	kindPresent := pendingKind != nil
	switch status {
	case MailboxDraftStatusActive, MailboxDraftStatusSubmitting, MailboxDraftStatusSubmitted:
		return enginePairPresent && !mutationPresent && !kindPresent && noPendingSave
	case MailboxDraftStatusSaving:
		return mutationPresent && kindPresent && ((*pendingKind == MailboxDraftMutationCreate && !enginePairPresent && noPendingSave) || (*pendingKind == MailboxDraftMutationSave && enginePairPresent && pendingSaveValid))
	case MailboxDraftStatusDeleting:
		return mutationPresent && kindPresent && *pendingKind == MailboxDraftMutationDelete && enginePairPresent && noPendingSave
	case MailboxDraftStatusUnknown:
		return mutationPresent && kindPresent && ((*pendingKind == MailboxDraftMutationCreate && !enginePairPresent && noPendingSave) || (*pendingKind == MailboxDraftMutationSave && enginePairPresent && pendingSaveValid) || (*pendingKind == MailboxDraftMutationDelete && enginePairPresent && noPendingSave))
	case MailboxDraftStatusDeleted:
		return !mutationPresent && !kindPresent && noPendingSave
	default:
		return false
	}
}

func mailboxEntDraftReadRecord(draft *ent.MailDraft, attachments mailboxEntDraftAttachmentState) (MailboxDraftRecord, bool) {
	if draft == nil || !mailboxEntDraftReadRowValid(draft, draft.MailAccountID, draft.ID) || len(attachments.ConfirmedAttachmentIDs) > MaxEngineAttachmentItems || len(attachments.PendingAttachmentIDs) > MaxEngineAttachmentItems || !mailboxEntDraftReadUUIDSliceValid(attachments.ConfirmedAttachmentIDs) || !mailboxEntDraftReadUUIDSliceValid(attachments.PendingAttachmentIDs) {
		return MailboxDraftRecord{}, false
	}
	pendingKind, ok := mailboxEntDraftReadPendingKind(draft.PendingMutationKind)
	if !ok {
		return MailboxDraftRecord{}, false
	}
	if pendingKind != nil && *pendingKind == MailboxDraftMutationSave {
		if draft.PendingFromIdentityID == nil || len(draft.PendingSaveRequestHash) != 32 {
			return MailboxDraftRecord{}, false
		}
	} else if len(attachments.PendingAttachmentIDs) != 0 {
		return MailboxDraftRecord{}, false
	}
	if MailboxDraftStatus(draft.Status) == MailboxDraftStatusDeleted && len(attachments.ConfirmedAttachmentIDs) != 0 {
		return MailboxDraftRecord{}, false
	}
	confirmed := MailboxDraftEffectiveState{
		FromIdentityID:   draft.FromIdentityID,
		ReplyToMessageID: mailboxEntDraftReadUUIDPointerCopy(draft.ReplyToMessageID),
		ForwardMessageID: mailboxEntDraftReadUUIDPointerCopy(draft.ForwardMessageID),
		AttachmentIDs:    append([]uuid.UUID(nil), attachments.ConfirmedAttachmentIDs...),
	}
	var pending *MailboxDraftPendingSave
	if pendingKind != nil && *pendingKind == MailboxDraftMutationSave {
		var requestHash [32]byte
		copy(requestHash[:], draft.PendingSaveRequestHash)
		pending = &MailboxDraftPendingSave{EffectiveState: MailboxDraftEffectiveState{
			FromIdentityID:   *draft.PendingFromIdentityID,
			ReplyToMessageID: mailboxEntDraftReadUUIDPointerCopy(draft.PendingReplyToMessageID),
			ForwardMessageID: mailboxEntDraftReadUUIDPointerCopy(draft.PendingForwardMessageID),
			AttachmentIDs:    append([]uuid.UUID(nil), attachments.PendingAttachmentIDs...),
		}, RequestHash: requestHash}
	}
	var createRequestHash [32]byte
	copy(createRequestHash[:], draft.CreateRequestHash)
	return MailboxDraftRecord{
		ID:                  draft.ID,
		AccountID:           draft.MailAccountID,
		DraftClientID:       draft.DraftClientID,
		EngineDraftKey:      draft.EngineDraftKey,
		EngineMessageID:     mailboxEntStringPointer(draft.EngineMessageID),
		EngineVersion:       mailboxEntStringPointer(draft.EngineVersion),
		Revision:            draft.Revision,
		Status:              MailboxDraftStatus(draft.Status),
		Confirmed:           confirmed,
		PendingSave:         mailboxEntDraftPendingSaveCopy(pending),
		CreateRequestHash:   createRequestHash,
		LastMutationID:      mailboxEntDraftReadUUIDPointerCopy(draft.LastMutationID),
		PendingMutationKind: mailboxEntDraftReadMutationKindPointerCopy(pendingKind),
		ConfirmedMutationID: mailboxEntDraftReadUUIDPointerCopy(draft.ConfirmedMutationID),
		UpdatedAt:           draft.UpdatedAt.UTC(),
	}, true
}

func mailboxEntDraftReadUUIDSliceValid(values []uuid.UUID) bool {
	for index, value := range values {
		if value == uuid.Nil || index > 0 && !syncLessUUID(values[index-1], value) {
			return false
		}
	}
	return true
}

func mailboxEntDraftReadHashValid(value []byte) bool {
	if len(value) != 32 {
		return false
	}
	for _, part := range value {
		if part != 0 {
			return true
		}
	}
	return false
}

func mailboxEntDraftEffectiveStateCopy(value MailboxDraftEffectiveState) MailboxDraftEffectiveState {
	return MailboxDraftEffectiveState{FromIdentityID: value.FromIdentityID, ReplyToMessageID: mailboxEntDraftReadUUIDPointerCopy(value.ReplyToMessageID), ForwardMessageID: mailboxEntDraftReadUUIDPointerCopy(value.ForwardMessageID), AttachmentIDs: append([]uuid.UUID(nil), value.AttachmentIDs...)}
}

func mailboxEntDraftPendingSaveCopy(value *MailboxDraftPendingSave) *MailboxDraftPendingSave {
	if value == nil {
		return nil
	}
	copied := MailboxDraftPendingSave{EffectiveState: mailboxEntDraftEffectiveStateCopy(value.EffectiveState), RequestHash: value.RequestHash}
	return &copied
}

func mailboxEntDraftReadUUIDPointerValid(value *uuid.UUID) bool { return value == nil || *value != uuid.Nil }

func mailboxEntDraftReadUUIDPointerCopy(value *uuid.UUID) *uuid.UUID {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

func mailboxEntDraftReadMutationKindPointerCopy(value *MailboxDraftMutationKind) *MailboxDraftMutationKind {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

func mailboxEntDraftReadTimeCopy(value time.Time) time.Time { return value.UTC() }
