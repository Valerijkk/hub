package mail

import (
	"bytes"
	"context"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailattachment"
	"github.com/valerii/hub/backend/ent/maildraft"
	"github.com/valerii/hub/backend/ent/mailobject"
)

// PrepareMailboxDraftReconciliation returns the exact pending draft work under the acquired fence.
func (repository *MailboxEntRepository) PrepareMailboxDraftReconciliation(ctx context.Context, request MailboxDraftReconcileRequest) (MailboxDraftReconcilePlan, error) {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxEntEngineCommandFenceValid(request.Fence) || request.DraftID == uuid.Nil {
		return MailboxDraftReconcilePlan{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}

	var result MailboxDraftReconcilePlan
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, databaseWallClock, err := mailboxEntDraftCreateLockAccount(ctx, transaction, request.Fence)
		if err != nil {
			return err
		}
		if kind := mailboxEntDraftCreateAccountKind(account); kind != "" {
			return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, kind)
		}

		draft, kind := mailboxEntDraftReservationLockDraft(ctx, transaction, request.Fence.AccountID, request.DraftID)
		if kind != "" {
			return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, kind)
		}
		attachmentIDs, kind := mailboxEntDraftReconciliationLockAttachments(ctx, transaction, draft, MailboxDraftRepositoryErrorStateConflict)
		if kind != "" {
			return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, kind)
		}
		record, ok := mailboxEntDraftReadRecord(draft, attachmentIDs)
		if !ok {
			return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
		}

		switch record.Status {
		case MailboxDraftStatusSaving:
			workKind, valid := mailboxEntDraftReconciliationWorkKind(record)
			if !valid || (record.PendingMutationKind != nil && *record.PendingMutationKind == MailboxDraftMutationCreate && record.Revision != 1) {
				return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
			}
			work := mailboxEntDraftCreateWork(record, workKind)
			if work == nil {
				return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
			}
			if err := draft.Update().
				SetStatus(maildraft.StatusUnknown).
				SetUpdatedAt(databaseWallClock).
				Exec(ctx); err != nil {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
			}
			record.Status = MailboxDraftStatusUnknown
			fence := mailboxEntDraftCreateFenceCopy(request.Fence)
			result = MailboxDraftReconcilePlan{
				Reconcile: true,
				Draft:     mailboxEntDraftReconciliationRecordCopy(record),
				Work:      work,
				Fence:     &fence,
			}
			return nil
		case MailboxDraftStatusDeleting:
			if record.PendingMutationKind == nil || *record.PendingMutationKind != MailboxDraftMutationDelete {
				return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
			}
			work := mailboxEntDraftCreateWork(record, MailboxDraftWorkReconcileDelete)
			if work == nil {
				return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
			}
			if err := draft.Update().
				SetStatus(maildraft.StatusUnknown).
				SetUpdatedAt(databaseWallClock).
				Exec(ctx); err != nil {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
			}
			record.Status = MailboxDraftStatusUnknown
			fence := mailboxEntDraftCreateFenceCopy(request.Fence)
			result = MailboxDraftReconcilePlan{
				Reconcile: true,
				Draft:     mailboxEntDraftReconciliationRecordCopy(record),
				Work:      work,
				Fence:     &fence,
			}
			return nil
		case MailboxDraftStatusUnknown:
			workKind, valid := mailboxEntDraftReconciliationWorkKind(record)
			if !valid || (record.PendingMutationKind != nil && *record.PendingMutationKind == MailboxDraftMutationCreate && record.Revision != 1) {
				return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
			}
			work := mailboxEntDraftCreateWork(record, workKind)
			if work == nil {
				return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
			}
			fence := mailboxEntDraftCreateFenceCopy(request.Fence)
			result = MailboxDraftReconcilePlan{
				Reconcile: true,
				Draft:     mailboxEntDraftReconciliationRecordCopy(record),
				Work:      work,
				Fence:     &fence,
			}
			return nil
		case MailboxDraftStatusActive, MailboxDraftStatusSubmitting, MailboxDraftStatusSubmitted, MailboxDraftStatusDeleted:
			if err := mailboxEntDraftCreateRelease(ctx, account, request.Fence, databaseWallClock); err != nil {
				return err
			}
			result = MailboxDraftReconcilePlan{Draft: mailboxEntDraftReconciliationRecordCopy(record)}
			return nil
		default:
			return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
		}
	})
	if err != nil {
		return MailboxDraftReconcilePlan{}, mailboxEntDraftRepositoryError(err)
	}
	return result, nil
}

// ResumeMailboxDraftCreateAfterAbsence permits the one controlled create retry after authoritative absence.
func (repository *MailboxEntRepository) ResumeMailboxDraftCreateAfterAbsence(ctx context.Context, request MailboxDraftResumeCreateRequest) (MailboxDraftReservation, error) {
	if repository == nil || repository.client == nil || ctx == nil {
		return MailboxDraftReservation{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	if kind := mailboxEntDraftReconciliationResumeRequestKind(request); kind != "" {
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
			return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, kind)
		}

		draft, kind := mailboxEntDraftReservationLockDraft(ctx, transaction, request.Fence.AccountID, request.DraftID)
		if kind != "" {
			return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, kind)
		}
		if !mailboxEntDraftReconciliationResumeIdentityMatches(draft, request, engineDraftKey) {
			return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorIdempotencyConflict)
		}
		attachmentIDs, kind := mailboxEntDraftReconciliationLockAttachments(ctx, transaction, draft, MailboxDraftRepositoryErrorAttachmentConflict)
		if kind != "" {
			return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, kind)
		}
		if !mailboxEntDraftReservationUUIDSetsEqual(attachmentIDs, request.SelectedAttachmentIDs) {
			return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorAttachmentConflict)
		}
		record, ok := mailboxEntDraftReadRecord(draft, attachmentIDs)
		if !ok {
			return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
		}

		switch record.Status {
		case MailboxDraftStatusUnknown:
			if record.PendingMutationKind == nil || *record.PendingMutationKind != MailboxDraftMutationCreate || record.LastMutationID == nil || record.Revision != 1 || record.EngineMessageID != nil || record.EngineVersion != nil {
				return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
			}
			work := mailboxEntDraftCreateWork(record, MailboxDraftWorkCreate)
			if work == nil {
				return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
			}
			if err := draft.Update().
				SetStatus(maildraft.StatusSaving).
				SetUpdatedAt(databaseWallClock).
				Exec(ctx); err != nil {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
			}
			record.Status = MailboxDraftStatusSaving
			fence := mailboxEntDraftCreateFenceCopy(request.Fence)
			result = MailboxDraftReservation{
				Kind:  MailboxDraftReservationDispatch,
				Draft: mailboxEntDraftReconciliationRecordCopy(record),
				Work:  work,
				Fence: &fence,
			}
			return nil
		case MailboxDraftStatusSaving:
			if record.PendingMutationKind == nil || *record.PendingMutationKind != MailboxDraftMutationCreate || record.LastMutationID == nil || record.Revision != 1 || record.EngineMessageID != nil || record.EngineVersion != nil {
				return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
			}
			work := mailboxEntDraftCreateWork(record, MailboxDraftWorkReconcileCreate)
			if work == nil {
				return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
			}
			fence := mailboxEntDraftCreateFenceCopy(request.Fence)
			result = MailboxDraftReservation{
				Kind:  MailboxDraftReservationReconcileRequired,
				Draft: mailboxEntDraftReconciliationRecordCopy(record),
				Work:  work,
				Fence: &fence,
			}
			return nil
		case MailboxDraftStatusActive, MailboxDraftStatusSubmitting, MailboxDraftStatusSubmitted, MailboxDraftStatusDeleted:
			if err := mailboxEntDraftCreateRelease(ctx, account, request.Fence, databaseWallClock); err != nil {
				return err
			}
			result = MailboxDraftReservation{Kind: MailboxDraftReservationReplay, Draft: mailboxEntDraftReconciliationRecordCopy(record)}
			return nil
		default:
			return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
		}
	})
	if err != nil {
		return MailboxDraftReservation{}, mailboxEntDraftRepositoryError(err)
	}
	return result, nil
}

// FinishEngineCommand applies one exact engine outcome and releases the acquired command fence.
func (repository *MailboxEntRepository) FinishEngineCommand(ctx context.Context, request MailboxDraftFinishRequest) (MailboxDraftFinishResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxEntEngineCommandFenceValid(request.Fence) || request.DraftID == uuid.Nil {
		return MailboxDraftFinishResult{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}

	var result MailboxDraftFinishResult
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, databaseWallClock, err := mailboxEntDraftCreateLockAccount(ctx, transaction, request.Fence)
		if err != nil {
			return err
		}
		if kind := mailboxEntDraftCreateAccountKind(account); kind != "" {
			return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, kind)
		}

		draft, kind := mailboxEntDraftReservationLockDraft(ctx, transaction, request.Fence.AccountID, request.DraftID)
		if kind != "" {
			return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, kind)
		}
		attachmentIDs, kind := mailboxEntDraftReconciliationLockAttachments(ctx, transaction, draft, MailboxDraftRepositoryErrorStateConflict)
		if kind != "" {
			return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, kind)
		}
		record, ok := mailboxEntDraftReadRecord(draft, attachmentIDs)
		if !ok {
			return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
		}
		if kind := mailboxEntDraftReconciliationFinishFormKind(request); kind != "" {
			return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, kind)
		}
		if record.LastMutationID == nil || *record.LastMutationID != request.MutationID || record.PendingMutationKind == nil || *record.PendingMutationKind != request.Kind {
			return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
		}

		updated, kind := mailboxEntDraftReconciliationFinishTransition(ctx, draft, record, request, databaseWallClock)
		if kind != "" {
			return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, kind)
		}
		if err := mailboxEntDraftCreateRelease(ctx, account, request.Fence, databaseWallClock); err != nil {
			return err
		}
		result = MailboxDraftFinishResult{Draft: mailboxEntDraftReconciliationRecordCopy(updated)}
		return nil
	})
	if err != nil {
		return MailboxDraftFinishResult{}, mailboxEntDraftRepositoryError(err)
	}
	return result, nil
}

func mailboxEntDraftReconciliationWorkKind(record MailboxDraftRecord) (MailboxDraftWorkKind, bool) {
	if record.LastMutationID == nil || record.PendingMutationKind == nil {
		return "", false
	}
	switch *record.PendingMutationKind {
	case MailboxDraftMutationCreate:
		return MailboxDraftWorkReconcileCreate, true
	case MailboxDraftMutationSave:
		return MailboxDraftWorkReconcileSave, true
	case MailboxDraftMutationDelete:
		return MailboxDraftWorkReconcileDelete, true
	default:
		return "", false
	}
}

func mailboxEntDraftReconciliationResumeRequestKind(request MailboxDraftResumeCreateRequest) MailboxDraftRepositoryErrorKind {
	if !mailboxEntEngineCommandFenceValid(request.Fence) || request.DraftID == uuid.Nil || request.DraftClientID == uuid.Nil || !mailboxEntDraftCreateHashPresent(request.CreateIdempotencyKeyHash) || !mailboxEntDraftCreateHashPresent(request.CreateRequestHash) {
		return MailboxDraftRepositoryErrorInvalid
	}
	return mailboxEntDraftReservationAttachmentIDsKind(request.SelectedAttachmentIDs)
}

func mailboxEntDraftReconciliationResumeIdentityMatches(draft *ent.MailDraft, request MailboxDraftResumeCreateRequest, engineDraftKey string) bool {
	return draft != nil &&
		draft.ID == request.DraftID &&
		draft.MailAccountID == request.Fence.AccountID &&
		draft.DraftClientID == request.DraftClientID &&
		draft.EngineDraftKey == engineDraftKey &&
		bytes.Equal(draft.CreateIdempotencyKeyHash, request.CreateIdempotencyKeyHash[:]) &&
		bytes.Equal(draft.CreateRequestHash, request.CreateRequestHash[:])
}

func mailboxEntDraftReconciliationFinishFormKind(request MailboxDraftFinishRequest) MailboxDraftRepositoryErrorKind {
	if request.MutationID == uuid.Nil {
		return MailboxDraftRepositoryErrorInvalid
	}
	switch request.Kind {
	case MailboxDraftMutationCreate, MailboxDraftMutationSave, MailboxDraftMutationDelete:
	default:
		return MailboxDraftRepositoryErrorInvalid
	}
	switch outcome := request.Outcome.(type) {
	case MailboxDraftApplied:
		if !syncValidOpaque(outcome.EngineMessageID, MaxEngineOpaqueIDBytes) || !syncValidOpaque(outcome.EngineVersion, MaxEngineOpaqueIDBytes) {
			return MailboxDraftRepositoryErrorInvalid
		}
	case MailboxDraftSaveRetained:
		if kind := mailboxEntDraftReservationAttachmentIDsKind(outcome.ProviderConfirmedAttachmentIDs); kind != "" {
			return MailboxDraftRepositoryErrorInvalid
		}
	case MailboxDraftDeleteRetained, MailboxDraftDeleted, MailboxDraftUnknown:
	default:
		return MailboxDraftRepositoryErrorInvalid
	}
	return ""
}

func mailboxEntDraftReconciliationFinishTransition(ctx context.Context, draft *ent.MailDraft, record MailboxDraftRecord, request MailboxDraftFinishRequest, databaseWallClock time.Time) (MailboxDraftRecord, MailboxDraftRepositoryErrorKind) {
	if ctx == nil || draft == nil || !mailboxEntValidTime(databaseWallClock) {
		return MailboxDraftRecord{}, MailboxDraftRepositoryErrorInvalid
	}

	updated := mailboxEntDraftReconciliationRecordCopy(record)
	update := draft.Update().SetUpdatedAt(databaseWallClock)
	switch request.Kind {
	case MailboxDraftMutationCreate:
		if record.Revision != 1 || record.EngineMessageID != nil || record.EngineVersion != nil {
			return MailboxDraftRecord{}, MailboxDraftRepositoryErrorStateConflict
		}
		switch outcome := request.Outcome.(type) {
		case MailboxDraftApplied:
			if !syncValidOpaque(outcome.EngineMessageID, MaxEngineOpaqueIDBytes) || !syncValidOpaque(outcome.EngineVersion, MaxEngineOpaqueIDBytes) {
				return MailboxDraftRecord{}, MailboxDraftRepositoryErrorInvalid
			}
			if err := update.
				SetStatus(maildraft.StatusActive).
				SetEngineMessageID(outcome.EngineMessageID).
				SetEngineVersion(outcome.EngineVersion).
				SetRevision(1).
				ClearLastMutationID().
				ClearPendingMutationKind().
				Exec(ctx); err != nil {
				return MailboxDraftRecord{}, MailboxDraftRepositoryErrorUnavailable
			}
			updated.Status = MailboxDraftStatusActive
			updated.EngineMessageID = mailboxEntStringPointer(&outcome.EngineMessageID)
			updated.EngineVersion = mailboxEntStringPointer(&outcome.EngineVersion)
			updated.Revision = 1
			updated.LastMutationID = nil
			updated.PendingMutationKind = nil
			return updated, ""
		case MailboxDraftUnknown:
			if err := update.SetStatus(maildraft.StatusUnknown).Exec(ctx); err != nil {
				return MailboxDraftRecord{}, MailboxDraftRepositoryErrorUnavailable
			}
			updated.Status = MailboxDraftStatusUnknown
			return updated, ""
		case MailboxDraftSaveRetained, MailboxDraftDeleteRetained, MailboxDraftDeleted:
			return MailboxDraftRecord{}, MailboxDraftRepositoryErrorStateConflict
		default:
			return MailboxDraftRecord{}, MailboxDraftRepositoryErrorInvalid
		}
	case MailboxDraftMutationSave:
		switch outcome := request.Outcome.(type) {
		case MailboxDraftApplied:
			if !syncValidOpaque(outcome.EngineMessageID, MaxEngineOpaqueIDBytes) || !syncValidOpaque(outcome.EngineVersion, MaxEngineOpaqueIDBytes) {
				return MailboxDraftRecord{}, MailboxDraftRepositoryErrorInvalid
			}
			nextRevision, ok := mailboxEntNextEpoch(record.Revision)
			if !ok {
				return MailboxDraftRecord{}, MailboxDraftRepositoryErrorStateConflict
			}
			if err := update.
				SetStatus(maildraft.StatusActive).
				SetEngineMessageID(outcome.EngineMessageID).
				SetEngineVersion(outcome.EngineVersion).
				SetRevision(nextRevision).
				ClearLastMutationID().
				ClearPendingMutationKind().
				Exec(ctx); err != nil {
				return MailboxDraftRecord{}, MailboxDraftRepositoryErrorUnavailable
			}
			updated.Status = MailboxDraftStatusActive
			updated.EngineMessageID = mailboxEntStringPointer(&outcome.EngineMessageID)
			updated.EngineVersion = mailboxEntStringPointer(&outcome.EngineVersion)
			updated.Revision = nextRevision
			updated.LastMutationID = nil
			updated.PendingMutationKind = nil
			return updated, ""
		case MailboxDraftSaveRetained:
			if kind := mailboxEntDraftReservationAttachmentIDsKind(outcome.ProviderConfirmedAttachmentIDs); kind != "" {
				return MailboxDraftRecord{}, MailboxDraftRepositoryErrorInvalid
			}
			if !mailboxEntDraftReservationUUIDSetsEqual(record.SelectedAttachmentIDs, outcome.ProviderConfirmedAttachmentIDs) {
				return MailboxDraftRecord{}, MailboxDraftRepositoryErrorStateConflict
			}
			if err := update.
				SetStatus(maildraft.StatusActive).
				ClearLastMutationID().
				ClearPendingMutationKind().
				Exec(ctx); err != nil {
				return MailboxDraftRecord{}, MailboxDraftRepositoryErrorUnavailable
			}
			updated.Status = MailboxDraftStatusActive
			updated.LastMutationID = nil
			updated.PendingMutationKind = nil
			return updated, ""
		case MailboxDraftUnknown:
			if err := update.SetStatus(maildraft.StatusUnknown).Exec(ctx); err != nil {
				return MailboxDraftRecord{}, MailboxDraftRepositoryErrorUnavailable
			}
			updated.Status = MailboxDraftStatusUnknown
			return updated, ""
		case MailboxDraftDeleteRetained, MailboxDraftDeleted:
			return MailboxDraftRecord{}, MailboxDraftRepositoryErrorStateConflict
		default:
			return MailboxDraftRecord{}, MailboxDraftRepositoryErrorInvalid
		}
	case MailboxDraftMutationDelete:
		switch request.Outcome.(type) {
		case MailboxDraftDeleted:
			if err := update.
				SetStatus(maildraft.StatusDeleted).
				ClearLastMutationID().
				ClearPendingMutationKind().
				ClearAttachments().
				Exec(ctx); err != nil {
				return MailboxDraftRecord{}, MailboxDraftRepositoryErrorUnavailable
			}
			updated.Status = MailboxDraftStatusDeleted
			updated.LastMutationID = nil
			updated.PendingMutationKind = nil
			updated.SelectedAttachmentIDs = []uuid.UUID{}
			return updated, ""
		case MailboxDraftDeleteRetained:
			if err := update.
				SetStatus(maildraft.StatusActive).
				ClearLastMutationID().
				ClearPendingMutationKind().
				Exec(ctx); err != nil {
				return MailboxDraftRecord{}, MailboxDraftRepositoryErrorUnavailable
			}
			updated.Status = MailboxDraftStatusActive
			updated.LastMutationID = nil
			updated.PendingMutationKind = nil
			return updated, ""
		case MailboxDraftUnknown:
			if err := update.SetStatus(maildraft.StatusUnknown).Exec(ctx); err != nil {
				return MailboxDraftRecord{}, MailboxDraftRepositoryErrorUnavailable
			}
			updated.Status = MailboxDraftStatusUnknown
			return updated, ""
		case MailboxDraftApplied, MailboxDraftSaveRetained:
			return MailboxDraftRecord{}, MailboxDraftRepositoryErrorStateConflict
		default:
			return MailboxDraftRecord{}, MailboxDraftRepositoryErrorInvalid
		}
	default:
		return MailboxDraftRecord{}, MailboxDraftRepositoryErrorInvalid
	}
}

func mailboxEntDraftReconciliationLockAttachments(ctx context.Context, transaction *ent.Tx, draft *ent.MailDraft, attachmentFailureKind MailboxDraftRepositoryErrorKind) ([]uuid.UUID, MailboxDraftRepositoryErrorKind) {
	if ctx == nil || transaction == nil || draft == nil || draft.ID == uuid.Nil || draft.MailAccountID == uuid.Nil || draft.DraftClientID == uuid.Nil || attachmentFailureKind == "" {
		return nil, MailboxDraftRepositoryErrorInvalid
	}
	attachments, err := transaction.MailAttachment.Query().
		Where(mailattachment.MailDraftIDEQ(draft.ID)).
		Order(ent.Asc(mailattachment.FieldID)).
		Limit(MaxEngineAttachmentItems + 1).
		ForUpdate().
		All(ctx)
	if err != nil {
		return nil, MailboxDraftRepositoryErrorUnavailable
	}
	if len(attachments) > MaxEngineAttachmentItems {
		return nil, attachmentFailureKind
	}

	attachmentByID := make(map[uuid.UUID]*ent.MailAttachment, len(attachments))
	attachmentIDs := make([]uuid.UUID, 0, len(attachments))
	objectIDs := make([]uuid.UUID, 0, len(attachments))
	objectSeen := make(map[uuid.UUID]struct{}, len(attachments))
	for _, attachment := range attachments {
		if !mailboxEntDraftReadAttachmentValid(attachment, draft.ID, draft.DraftClientID) {
			return nil, attachmentFailureKind
		}
		if _, exists := attachmentByID[attachment.ID]; exists {
			return nil, attachmentFailureKind
		}
		attachmentByID[attachment.ID] = attachment
		attachmentIDs = append(attachmentIDs, attachment.ID)
		objectID := *attachment.StorageObjectID
		if _, exists := objectSeen[objectID]; !exists {
			objectSeen[objectID] = struct{}{}
			objectIDs = append(objectIDs, objectID)
		}
	}
	sort.Slice(attachmentIDs, func(left, right int) bool {
		return syncLessUUID(attachmentIDs[left], attachmentIDs[right])
	})
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
				return nil, attachmentFailureKind
			}
			return nil, MailboxDraftRepositoryErrorUnavailable
		}
		if object == nil || object.ID != objectID {
			return nil, attachmentFailureKind
		}
		objects[objectID] = object
	}
	if attachmentFailureKind == MailboxDraftRepositoryErrorAttachmentConflict {
		for _, attachmentID := range attachmentIDs {
			attachment := attachmentByID[attachmentID]
			if attachment.ScanStatus != mailattachment.ScanStatus("clean") || !mailboxEntDraftCreateStorageObjectValid(objects[*attachment.StorageObjectID], draft.MailAccountID, draft.DraftClientID, attachment) {
				return nil, MailboxDraftRepositoryErrorAttachmentConflict
			}
		}
	}
	return attachmentIDs, ""
}

func mailboxEntDraftReconciliationLogicalError(ctx context.Context, account *ent.MailAccount, fence MailboxEngineCommandFence, databaseWallClock time.Time, kind MailboxDraftRepositoryErrorKind) error {
	if kind == MailboxDraftRepositoryErrorUnavailable {
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
	}
	return mailboxEntDraftCreateReleaseError(ctx, account, fence, databaseWallClock, kind)
}

func mailboxEntDraftReconciliationRecordCopy(record MailboxDraftRecord) MailboxDraftRecord {
	return MailboxDraftRecord{
		ID:                    record.ID,
		AccountID:             record.AccountID,
		DraftClientID:         record.DraftClientID,
		EngineDraftKey:        record.EngineDraftKey,
		EngineMessageID:       mailboxEntStringPointer(record.EngineMessageID),
		EngineVersion:         mailboxEntStringPointer(record.EngineVersion),
		Revision:              record.Revision,
		Status:                record.Status,
		ReplyToMessageID:      mailboxEntDraftReadUUIDPointerCopy(record.ReplyToMessageID),
		ForwardMessageID:      mailboxEntDraftReadUUIDPointerCopy(record.ForwardMessageID),
		LastMutationID:        mailboxEntDraftReadUUIDPointerCopy(record.LastMutationID),
		PendingMutationKind:   mailboxEntDraftReadMutationKindPointerCopy(record.PendingMutationKind),
		SelectedAttachmentIDs: append([]uuid.UUID(nil), record.SelectedAttachmentIDs...),
	}
}
