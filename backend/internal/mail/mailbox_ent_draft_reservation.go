package mail

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailattachment"
	"github.com/valerii/hub/backend/ent/maildraft"
	"github.com/valerii/hub/backend/ent/mailobject"
)

// ReserveMailboxDraftSave persists or recovers one draft-save operation under the exact acquired fence.
func (repository *MailboxEntRepository) ReserveMailboxDraftSave(ctx context.Context, request MailboxDraftSaveRequest) (MailboxDraftReservation, error) {
	if repository == nil || repository.client == nil || ctx == nil {
		return MailboxDraftReservation{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	if kind := mailboxEntDraftReservationSaveRequestKind(request); kind != "" {
		return MailboxDraftReservation{}, NewMailboxDraftRepositoryError(kind)
	}

	var result MailboxDraftReservation
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, databaseWallClock, err := mailboxEntDraftCreateLockAccount(ctx, transaction, request.Fence)
		if err != nil {
			return err
		}
		if kind := mailboxEntDraftCreateAccountKind(account); kind != "" {
			return mailboxEntDraftReservationReleaseError(ctx, account, request.Fence, databaseWallClock, kind)
		}

		draft, kind := mailboxEntDraftReservationLockDraft(ctx, transaction, request.Fence.AccountID, request.DraftID)
		if kind != "" {
			return mailboxEntDraftReservationReleaseError(ctx, account, request.Fence, databaseWallClock, kind)
		}
		record, kind := mailboxEntDraftReservationSaveRecord(ctx, transaction, draft, request)
		if kind != "" {
			return mailboxEntDraftReservationReleaseError(ctx, account, request.Fence, databaseWallClock, kind)
		}

		switch record.Status {
		case MailboxDraftStatusActive:
			if record.Revision != request.ExpectedRevision || record.DraftClientID != request.DraftClientID {
				return mailboxEntDraftReservationReleaseError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
			}
			updated, err := mailboxEntDraftReservationSaveTransition(ctx, draft, request, databaseWallClock)
			if err != nil {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
			}
			updatedRecord, ok := mailboxEntDraftReadRecord(updated, request.SelectedAttachmentIDs)
			if !ok {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
			}
			work := mailboxEntDraftCreateWork(updatedRecord, MailboxDraftWorkSave)
			if work == nil {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
			}
			fence := mailboxEntDraftCreateFenceCopy(request.Fence)
			result = MailboxDraftReservation{
				Kind:  MailboxDraftReservationDispatch,
				Draft: updatedRecord,
				Work:  work,
				Fence: &fence,
			}
			return nil
		case MailboxDraftStatusSaving, MailboxDraftStatusUnknown:
			if record.PendingMutationKind == nil || *record.PendingMutationKind != MailboxDraftMutationSave || record.LastMutationID == nil || *record.LastMutationID != request.MutationID || record.Revision != request.ExpectedRevision || record.DraftClientID != request.DraftClientID || !mailboxEntDraftReservationUUIDSetsEqual(record.SelectedAttachmentIDs, request.SelectedAttachmentIDs) {
				return mailboxEntDraftReservationReleaseError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
			}
			work := mailboxEntDraftCreateWork(record, MailboxDraftWorkReconcileSave)
			if work == nil {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
			}
			fence := mailboxEntDraftCreateFenceCopy(request.Fence)
			result = MailboxDraftReservation{
				Kind:  MailboxDraftReservationReconcileRequired,
				Draft: record,
				Work:  work,
				Fence: &fence,
			}
			return nil
		default:
			return mailboxEntDraftReservationReleaseError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
		}
	})
	if err != nil {
		return MailboxDraftReservation{}, mailboxEntDraftRepositoryError(err)
	}
	return result, nil
}

// ReserveMailboxDraftDelete persists or recovers one draft-delete operation under the exact acquired fence.
func (repository *MailboxEntRepository) ReserveMailboxDraftDelete(ctx context.Context, request MailboxDraftDeleteRequest) (MailboxDraftReservation, error) {
	if repository == nil || repository.client == nil || ctx == nil {
		return MailboxDraftReservation{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	if kind := mailboxEntDraftReservationDeleteRequestKind(request); kind != "" {
		return MailboxDraftReservation{}, NewMailboxDraftRepositoryError(kind)
	}

	var result MailboxDraftReservation
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, databaseWallClock, err := mailboxEntDraftCreateLockAccount(ctx, transaction, request.Fence)
		if err != nil {
			return err
		}
		if kind := mailboxEntDraftCreateAccountKind(account); kind != "" {
			return mailboxEntDraftReservationReleaseError(ctx, account, request.Fence, databaseWallClock, kind)
		}

		draft, kind := mailboxEntDraftReservationLockDraft(ctx, transaction, request.Fence.AccountID, request.DraftID)
		if kind != "" {
			return mailboxEntDraftReservationReleaseError(ctx, account, request.Fence, databaseWallClock, kind)
		}
		attachmentIDs, kind := mailboxEntDraftReservationLockCurrentAttachmentIDs(ctx, draft)
		if kind != "" {
			return mailboxEntDraftReservationReleaseError(ctx, account, request.Fence, databaseWallClock, kind)
		}
		record, ok := mailboxEntDraftReadRecord(draft, attachmentIDs)
		if !ok {
			return mailboxEntDraftReservationReleaseError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
		}

		switch record.Status {
		case MailboxDraftStatusActive:
			if record.Revision != request.ExpectedRevision {
				return mailboxEntDraftReservationReleaseError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
			}
			updated, err := mailboxEntDraftReservationDeleteTransition(ctx, draft, request, databaseWallClock)
			if err != nil {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
			}
			updatedRecord, ok := mailboxEntDraftReadRecord(updated, attachmentIDs)
			if !ok {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
			}
			work := mailboxEntDraftCreateWork(updatedRecord, MailboxDraftWorkDelete)
			if work == nil {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
			}
			fence := mailboxEntDraftCreateFenceCopy(request.Fence)
			result = MailboxDraftReservation{
				Kind:  MailboxDraftReservationDispatch,
				Draft: updatedRecord,
				Work:  work,
				Fence: &fence,
			}
			return nil
		case MailboxDraftStatusDeleting, MailboxDraftStatusUnknown:
			if record.PendingMutationKind == nil || *record.PendingMutationKind != MailboxDraftMutationDelete || record.LastMutationID == nil || *record.LastMutationID != request.MutationID || record.Revision != request.ExpectedRevision {
				return mailboxEntDraftReservationReleaseError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
			}
			work := mailboxEntDraftCreateWork(record, MailboxDraftWorkReconcileDelete)
			if work == nil {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
			}
			fence := mailboxEntDraftCreateFenceCopy(request.Fence)
			result = MailboxDraftReservation{
				Kind:  MailboxDraftReservationReconcileRequired,
				Draft: record,
				Work:  work,
				Fence: &fence,
			}
			return nil
		case MailboxDraftStatusDeleted:
			if record.Revision != request.ExpectedRevision || len(record.SelectedAttachmentIDs) != 0 || record.EngineMessageID == nil || record.EngineVersion == nil {
				return mailboxEntDraftReservationReleaseError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
			}
			if err := mailboxEntDraftCreateRelease(ctx, account, request.Fence, databaseWallClock); err != nil {
				return err
			}
			result = MailboxDraftReservation{Kind: MailboxDraftReservationReplay, Draft: record}
			return nil
		default:
			return mailboxEntDraftReservationReleaseError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
		}
	})
	if err != nil {
		return MailboxDraftReservation{}, mailboxEntDraftRepositoryError(err)
	}
	return result, nil
}

func mailboxEntDraftReservationSaveRequestKind(request MailboxDraftSaveRequest) MailboxDraftRepositoryErrorKind {
	if !mailboxEntEngineCommandFenceValid(request.Fence) || request.DraftID == uuid.Nil || request.DraftClientID == uuid.Nil || request.ExpectedRevision < 1 || request.MutationID == uuid.Nil {
		return MailboxDraftRepositoryErrorInvalid
	}
	return mailboxEntDraftReservationAttachmentIDsKind(request.SelectedAttachmentIDs)
}

func mailboxEntDraftReservationDeleteRequestKind(request MailboxDraftDeleteRequest) MailboxDraftRepositoryErrorKind {
	if !mailboxEntEngineCommandFenceValid(request.Fence) || request.DraftID == uuid.Nil || request.ExpectedRevision < 1 || request.MutationID == uuid.Nil {
		return MailboxDraftRepositoryErrorInvalid
	}
	return ""
}

func mailboxEntDraftReservationAttachmentIDsKind(attachmentIDs []uuid.UUID) MailboxDraftRepositoryErrorKind {
	if len(attachmentIDs) > MaxEngineAttachmentItems {
		return MailboxDraftRepositoryErrorAttachmentLimit
	}
	for index, attachmentID := range attachmentIDs {
		if attachmentID == uuid.Nil || (index > 0 && !syncLessUUID(attachmentIDs[index-1], attachmentID)) {
			return MailboxDraftRepositoryErrorInvalid
		}
	}
	return ""
}

func mailboxEntDraftReservationLockDraft(ctx context.Context, transaction *ent.Tx, accountID, draftID uuid.UUID) (*ent.MailDraft, MailboxDraftRepositoryErrorKind) {
	if ctx == nil || transaction == nil || accountID == uuid.Nil || draftID == uuid.Nil {
		return nil, MailboxDraftRepositoryErrorInvalid
	}
	draft, err := transaction.MailDraft.Query().
		Where(
			maildraft.IDEQ(draftID),
			maildraft.MailAccountIDEQ(accountID),
		).
		ForUpdate().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, MailboxDraftRepositoryErrorStateConflict
		}
		return nil, MailboxDraftRepositoryErrorUnavailable
	}
	if draft == nil || draft.ID != draftID || draft.MailAccountID != accountID || !mailboxEntDraftReadRowValid(draft, accountID, draftID) {
		return nil, MailboxDraftRepositoryErrorStateConflict
	}
	return draft, ""
}

func mailboxEntDraftReservationSaveRecord(ctx context.Context, transaction *ent.Tx, draft *ent.MailDraft, request MailboxDraftSaveRequest) (MailboxDraftRecord, MailboxDraftRepositoryErrorKind) {
	if ctx == nil || transaction == nil || draft == nil || !mailboxEntDraftReadRowValid(draft, request.Fence.AccountID, request.DraftID) {
		return MailboxDraftRecord{}, MailboxDraftRepositoryErrorStateConflict
	}
	currentAttachmentIDs, kind := mailboxEntDraftReservationCurrentAttachmentIDs(ctx, transaction, draft)
	if kind != "" {
		return MailboxDraftRecord{}, kind
	}
	attachmentIDs := mailboxEntDraftReservationUnionAttachmentIDs(currentAttachmentIDs, request.SelectedAttachmentIDs)
	attachments, kind := mailboxEntDraftReservationLockAttachments(ctx, transaction, attachmentIDs, currentAttachmentIDs)
	if kind != "" {
		return MailboxDraftRecord{}, kind
	}
	if kind := mailboxEntDraftReservationValidateSaveAttachments(ctx, transaction, draft, request, currentAttachmentIDs, attachmentIDs, attachments); kind != "" {
		return MailboxDraftRecord{}, kind
	}
	record, ok := mailboxEntDraftReadRecord(draft, currentAttachmentIDs)
	if !ok {
		return MailboxDraftRecord{}, MailboxDraftRepositoryErrorStateConflict
	}
	return record, ""
}

func mailboxEntDraftReservationCurrentAttachmentIDs(ctx context.Context, transaction *ent.Tx, draft *ent.MailDraft) ([]uuid.UUID, MailboxDraftRepositoryErrorKind) {
	if ctx == nil || transaction == nil || draft == nil || draft.ID == uuid.Nil {
		return nil, MailboxDraftRepositoryErrorInvalid
	}
	attachments, err := transaction.MailAttachment.Query().
		Where(mailattachment.MailDraftIDEQ(draft.ID)).
		Order(ent.Asc(mailattachment.FieldID)).
		Limit(MaxEngineAttachmentItems + 1).
		All(ctx)
	if err != nil {
		return nil, MailboxDraftRepositoryErrorUnavailable
	}
	if len(attachments) > MaxEngineAttachmentItems {
		return nil, MailboxDraftRepositoryErrorStateConflict
	}
	attachmentIDs := make([]uuid.UUID, 0, len(attachments))
	seen := make(map[uuid.UUID]struct{}, len(attachments))
	for _, attachment := range attachments {
		if attachment == nil || attachment.ID == uuid.Nil {
			return nil, MailboxDraftRepositoryErrorStateConflict
		}
		if _, duplicate := seen[attachment.ID]; duplicate {
			return nil, MailboxDraftRepositoryErrorStateConflict
		}
		seen[attachment.ID] = struct{}{}
		attachmentIDs = append(attachmentIDs, attachment.ID)
	}
	sort.Slice(attachmentIDs, func(left, right int) bool {
		return syncLessUUID(attachmentIDs[left], attachmentIDs[right])
	})
	return attachmentIDs, ""
}

func mailboxEntDraftReservationUnionAttachmentIDs(current, desired []uuid.UUID) []uuid.UUID {
	attachmentIDs := make([]uuid.UUID, 0, len(current)+len(desired))
	attachmentIDs = append(attachmentIDs, current...)
	attachmentIDs = append(attachmentIDs, desired...)
	sort.Slice(attachmentIDs, func(left, right int) bool {
		return syncLessUUID(attachmentIDs[left], attachmentIDs[right])
	})
	if len(attachmentIDs) < 2 {
		return attachmentIDs
	}
	unique := attachmentIDs[:1]
	for _, attachmentID := range attachmentIDs[1:] {
		if attachmentID != unique[len(unique)-1] {
			unique = append(unique, attachmentID)
		}
	}
	return unique
}

func mailboxEntDraftReservationLockAttachments(ctx context.Context, transaction *ent.Tx, attachmentIDs, currentAttachmentIDs []uuid.UUID) (map[uuid.UUID]*ent.MailAttachment, MailboxDraftRepositoryErrorKind) {
	if ctx == nil || transaction == nil {
		return nil, MailboxDraftRepositoryErrorInvalid
	}
	current := make(map[uuid.UUID]struct{}, len(currentAttachmentIDs))
	for _, attachmentID := range currentAttachmentIDs {
		current[attachmentID] = struct{}{}
	}
	attachments := make(map[uuid.UUID]*ent.MailAttachment, len(attachmentIDs))
	for _, attachmentID := range attachmentIDs {
		attachment, err := transaction.MailAttachment.Query().
			Where(mailattachment.IDEQ(attachmentID)).
			ForUpdate().
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				if _, found := current[attachmentID]; found {
					return nil, MailboxDraftRepositoryErrorStateConflict
				}
				return nil, MailboxDraftRepositoryErrorAttachmentConflict
			}
			return nil, MailboxDraftRepositoryErrorUnavailable
		}
		if attachment == nil || attachment.ID != attachmentID {
			return nil, MailboxDraftRepositoryErrorStateConflict
		}
		attachments[attachmentID] = attachment
	}
	return attachments, ""
}

func mailboxEntDraftReservationValidateSaveAttachments(ctx context.Context, transaction *ent.Tx, draft *ent.MailDraft, request MailboxDraftSaveRequest, currentAttachmentIDs, attachmentIDs []uuid.UUID, attachments map[uuid.UUID]*ent.MailAttachment) MailboxDraftRepositoryErrorKind {
	if ctx == nil || transaction == nil || draft == nil || len(attachments) != len(attachmentIDs) {
		return MailboxDraftRepositoryErrorInvalid
	}
	current := make(map[uuid.UUID]struct{}, len(currentAttachmentIDs))
	for _, attachmentID := range currentAttachmentIDs {
		current[attachmentID] = struct{}{}
	}
	desired := make(map[uuid.UUID]struct{}, len(request.SelectedAttachmentIDs))
	for _, attachmentID := range request.SelectedAttachmentIDs {
		desired[attachmentID] = struct{}{}
	}

	objectIDs := make([]uuid.UUID, 0, len(attachmentIDs))
	objectCurrent := make(map[uuid.UUID]bool, len(attachmentIDs))
	seenObjects := make(map[uuid.UUID]struct{}, len(attachmentIDs))
	for _, attachmentID := range attachmentIDs {
		attachment := attachments[attachmentID]
		_, isCurrent := current[attachmentID]
		_, isDesired := desired[attachmentID]
		if isCurrent && !mailboxEntDraftReadAttachmentValid(attachment, draft.ID, draft.DraftClientID) {
			return MailboxDraftRepositoryErrorStateConflict
		}
		if isDesired {
			if isCurrent {
				if !mailboxEntDraftReservationDesiredCurrentAttachmentValid(attachment, draft.ID, draft.DraftClientID) {
					return MailboxDraftRepositoryErrorAttachmentConflict
				}
			} else if !mailboxEntDraftCreateAttachmentValid(attachment, request.DraftClientID) {
				return MailboxDraftRepositoryErrorAttachmentConflict
			}
		}
		if attachment == nil || attachment.StorageObjectID == nil || *attachment.StorageObjectID == uuid.Nil {
			if isCurrent {
				return MailboxDraftRepositoryErrorStateConflict
			}
			return MailboxDraftRepositoryErrorAttachmentConflict
		}
		objectID := *attachment.StorageObjectID
		if _, found := seenObjects[objectID]; !found {
			seenObjects[objectID] = struct{}{}
			objectIDs = append(objectIDs, objectID)
		}
		if isCurrent {
			objectCurrent[objectID] = true
		}
	}
	sort.Slice(objectIDs, func(left, right int) bool {
		return syncLessUUID(objectIDs[left], objectIDs[right])
	})

	objects := make(map[uuid.UUID]*ent.MailObject, len(objectIDs))
	for _, objectID := range objectIDs {
		object, err := transaction.MailObject.Query().
			Where(mailobject.IDEQ(objectID)).
			ForUpdate().
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				if objectCurrent[objectID] {
					return MailboxDraftRepositoryErrorStateConflict
				}
				return MailboxDraftRepositoryErrorAttachmentConflict
			}
			return MailboxDraftRepositoryErrorUnavailable
		}
		if object == nil || object.ID != objectID {
			return MailboxDraftRepositoryErrorStateConflict
		}
		objects[objectID] = object
	}

	for _, attachmentID := range attachmentIDs {
		attachment := attachments[attachmentID]
		object := objects[*attachment.StorageObjectID]
		if attachment.DraftClientID == nil || !mailboxEntDraftCreateStorageObjectValid(object, draft.MailAccountID, *attachment.DraftClientID, attachment) {
			if _, isCurrent := current[attachmentID]; isCurrent {
				return MailboxDraftRepositoryErrorStateConflict
			}
			return MailboxDraftRepositoryErrorAttachmentConflict
		}
	}
	return ""
}

func mailboxEntDraftReservationDesiredCurrentAttachmentValid(attachment *ent.MailAttachment, draftID, draftClientID uuid.UUID) bool {
	return mailboxEntDraftReadAttachmentValid(attachment, draftID, draftClientID) &&
		attachment.ScanStatus == mailattachment.ScanStatus("clean")
}

func mailboxEntDraftReservationSaveTransition(ctx context.Context, draft *ent.MailDraft, request MailboxDraftSaveRequest, databaseWallClock time.Time) (*ent.MailDraft, error) {
	if ctx == nil || draft == nil || !mailboxEntValidTime(databaseWallClock) {
		return nil, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	update := draft.Update().
		SetStatus(maildraft.StatusSaving).
		SetLastMutationID(request.MutationID).
		SetPendingMutationKind(maildraft.PendingMutationKind("save")).
		SetUpdatedAt(databaseWallClock).
		ClearAttachments()
	if len(request.SelectedAttachmentIDs) != 0 {
		update.AddAttachmentIDs(request.SelectedAttachmentIDs...)
	}
	return update.Save(ctx)
}

func mailboxEntDraftReservationLockCurrentAttachmentIDs(ctx context.Context, draft *ent.MailDraft) ([]uuid.UUID, MailboxDraftRepositoryErrorKind) {
	if ctx == nil || draft == nil || draft.ID == uuid.Nil || draft.DraftClientID == uuid.Nil {
		return nil, MailboxDraftRepositoryErrorInvalid
	}
	attachments, err := draft.QueryAttachments().
		Order(ent.Asc(mailattachment.FieldID)).
		Limit(MaxEngineAttachmentItems + 1).
		ForUpdate().
		All(ctx)
	if err != nil {
		return nil, MailboxDraftRepositoryErrorUnavailable
	}
	if len(attachments) > MaxEngineAttachmentItems {
		return nil, MailboxDraftRepositoryErrorStateConflict
	}
	attachmentIDs := make([]uuid.UUID, 0, len(attachments))
	seen := make(map[uuid.UUID]struct{}, len(attachments))
	for _, attachment := range attachments {
		if !mailboxEntDraftReadAttachmentValid(attachment, draft.ID, draft.DraftClientID) {
			return nil, MailboxDraftRepositoryErrorStateConflict
		}
		if _, duplicate := seen[attachment.ID]; duplicate {
			return nil, MailboxDraftRepositoryErrorStateConflict
		}
		seen[attachment.ID] = struct{}{}
		attachmentIDs = append(attachmentIDs, attachment.ID)
	}
	sort.Slice(attachmentIDs, func(left, right int) bool {
		return syncLessUUID(attachmentIDs[left], attachmentIDs[right])
	})
	return attachmentIDs, ""
}

func mailboxEntDraftReservationDeleteTransition(ctx context.Context, draft *ent.MailDraft, request MailboxDraftDeleteRequest, databaseWallClock time.Time) (*ent.MailDraft, error) {
	if ctx == nil || draft == nil || !mailboxEntValidTime(databaseWallClock) {
		return nil, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	return draft.Update().
		SetStatus(maildraft.StatusDeleting).
		SetLastMutationID(request.MutationID).
		SetPendingMutationKind(maildraft.PendingMutationKind("delete")).
		SetUpdatedAt(databaseWallClock).
		Save(ctx)
}

func mailboxEntDraftReservationUUIDSetsEqual(left, right []uuid.UUID) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func mailboxEntDraftReservationReleaseError(ctx context.Context, account *ent.MailAccount, fence MailboxEngineCommandFence, databaseWallClock time.Time, kind MailboxDraftRepositoryErrorKind) error {
	if kind == MailboxDraftRepositoryErrorUnavailable {
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
	}
	return mailboxEntDraftCreateReleaseError(ctx, account, fence, databaseWallClock, kind)
}
