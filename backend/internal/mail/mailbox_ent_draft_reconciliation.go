package mail

import (
	"bytes"
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/maildraft"
)

// PrepareMailboxDraftReconciliation returns the exact pending O/D work under the acquired fence.
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
		attachments, kind := mailboxEntDraftReadAttachmentState(ctx, transaction, draft, MailboxDraftRepositoryErrorStateConflict)
		if kind != "" {
			return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, kind)
		}
		record, ok := mailboxEntDraftReadRecord(draft, attachments)
		if !ok {
			return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
		}
		switch record.Status {
		case MailboxDraftStatusSaving:
			workKind, valid := mailboxEntDraftReconciliationWorkKind(record)
			if !valid || record.PendingMutationKind != nil && *record.PendingMutationKind == MailboxDraftMutationCreate && (record.Revision != 1 || record.PendingSave != nil) {
				return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
			}
			work := mailboxEntDraftCreateWork(record, workKind)
			if work == nil {
				return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
			}
			if err := draft.Update().SetStatus(maildraft.StatusUnknown).SetUpdatedAt(databaseWallClock).Exec(ctx); err != nil {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
			}
			record.Status = MailboxDraftStatusUnknown
			record.UpdatedAt = databaseWallClock.UTC()
			fence := mailboxEntDraftCreateFenceCopy(request.Fence)
			result = MailboxDraftReconcilePlan{Reconcile: true, Draft: mailboxEntDraftRecordCopy(record), Work: work, Fence: &fence}
			return nil
		case MailboxDraftStatusDeleting:
			if record.PendingMutationKind == nil || *record.PendingMutationKind != MailboxDraftMutationDelete || record.PendingSave != nil {
				return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
			}
			work := mailboxEntDraftCreateWork(record, MailboxDraftWorkReconcileDelete)
			if work == nil {
				return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
			}
			if err := draft.Update().SetStatus(maildraft.StatusUnknown).SetUpdatedAt(databaseWallClock).Exec(ctx); err != nil {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
			}
			record.Status = MailboxDraftStatusUnknown
			record.UpdatedAt = databaseWallClock.UTC()
			fence := mailboxEntDraftCreateFenceCopy(request.Fence)
			result = MailboxDraftReconcilePlan{Reconcile: true, Draft: mailboxEntDraftRecordCopy(record), Work: work, Fence: &fence}
			return nil
		case MailboxDraftStatusUnknown:
			workKind, valid := mailboxEntDraftReconciliationWorkKind(record)
			if !valid || record.PendingMutationKind != nil && *record.PendingMutationKind == MailboxDraftMutationCreate && (record.Revision != 1 || record.PendingSave != nil) {
				return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
			}
			work := mailboxEntDraftCreateWork(record, workKind)
			if work == nil {
				return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
			}
			fence := mailboxEntDraftCreateFenceCopy(request.Fence)
			result = MailboxDraftReconcilePlan{Reconcile: true, Draft: mailboxEntDraftRecordCopy(record), Work: work, Fence: &fence}
			return nil
		case MailboxDraftStatusActive, MailboxDraftStatusSubmitting, MailboxDraftStatusSubmitted, MailboxDraftStatusDeleted:
			if err := mailboxEntDraftCreateRelease(ctx, account, request.Fence, databaseWallClock); err != nil {
				return err
			}
			result = MailboxDraftReconcilePlan{Draft: mailboxEntDraftRecordCopy(record)}
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
		if err := mailboxEntDraftCreateLockConfirmedIdentity(ctx, transaction, request.Fence.AccountID, request.Confirmed.FromIdentityID); err != nil {
			return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, mailboxEntDraftCreateErrorKind(err))
		}
		if err := mailboxEntDraftCreateLockReferences(ctx, transaction, request.Fence.AccountID, request.Confirmed); err != nil {
			return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, mailboxEntDraftCreateErrorKind(err))
		}
		attachments, kind := mailboxEntDraftReadAttachmentState(ctx, transaction, draft, MailboxDraftRepositoryErrorStateConflict)
		if kind != "" {
			return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, kind)
		}
		record, ok := mailboxEntDraftReadRecord(draft, attachments)
		if !ok || !mailboxEntDraftReconciliationResumeIdentityMatches(draft, record, request, engineDraftKey) {
			return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorIdempotencyConflict)
		}
		switch record.Status {
		case MailboxDraftStatusUnknown:
			if record.PendingMutationKind == nil || *record.PendingMutationKind != MailboxDraftMutationCreate || record.LastMutationID == nil || record.Revision != 1 || record.EngineMessageID != nil || record.EngineVersion != nil || record.PendingSave != nil {
				return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
			}
			work := mailboxEntDraftCreateWork(record, MailboxDraftWorkCreate)
			if work == nil {
				return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
			}
			if err := draft.Update().SetStatus(maildraft.StatusSaving).SetUpdatedAt(databaseWallClock).Exec(ctx); err != nil {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
			}
			record.Status = MailboxDraftStatusSaving
			record.UpdatedAt = databaseWallClock.UTC()
			fence := mailboxEntDraftCreateFenceCopy(request.Fence)
			result = MailboxDraftReservation{Kind: MailboxDraftReservationDispatch, Draft: mailboxEntDraftRecordCopy(record), Work: work, Fence: &fence}
			return nil
		case MailboxDraftStatusSaving:
			if record.PendingMutationKind == nil || *record.PendingMutationKind != MailboxDraftMutationCreate || record.LastMutationID == nil || record.Revision != 1 || record.EngineMessageID != nil || record.EngineVersion != nil || record.PendingSave != nil {
				return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
			}
			work := mailboxEntDraftCreateWork(record, MailboxDraftWorkReconcileCreate)
			if work == nil {
				return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
			}
			fence := mailboxEntDraftCreateFenceCopy(request.Fence)
			result = MailboxDraftReservation{Kind: MailboxDraftReservationReconcileRequired, Draft: mailboxEntDraftRecordCopy(record), Work: work, Fence: &fence}
			return nil
		case MailboxDraftStatusActive, MailboxDraftStatusSubmitting, MailboxDraftStatusSubmitted, MailboxDraftStatusDeleted:
			if err := mailboxEntDraftCreateRelease(ctx, account, request.Fence, databaseWallClock); err != nil {
				return err
			}
			result = MailboxDraftReservation{Kind: MailboxDraftReservationReplay, Draft: mailboxEntDraftRecordCopy(record)}
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
		if kind := mailboxEntDraftReconciliationFinishFormKind(request); kind != "" {
			return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, kind)
		}
		if request.Kind == MailboxDraftMutationSave {
			if draft.PendingFromIdentityID == nil {
				return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
			}
			pending := MailboxDraftEffectiveState{FromIdentityID: *draft.PendingFromIdentityID, ReplyToMessageID: mailboxEntDraftReadUUIDPointerCopy(draft.PendingReplyToMessageID), ForwardMessageID: mailboxEntDraftReadUUIDPointerCopy(draft.PendingForwardMessageID)}
			if err := mailboxEntDraftCreateLockConfirmedIdentity(ctx, transaction, request.Fence.AccountID, pending.FromIdentityID); err != nil {
				return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, mailboxEntDraftCreateErrorKind(err))
			}
			if err := mailboxEntDraftCreateLockReferences(ctx, transaction, request.Fence.AccountID, pending); err != nil {
				return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, mailboxEntDraftCreateErrorKind(err))
			}
		}
		attachments, kind := mailboxEntDraftReadAttachmentState(ctx, transaction, draft, MailboxDraftRepositoryErrorStateConflict)
		if kind != "" {
			return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, kind)
		}
		record, ok := mailboxEntDraftReadRecord(draft, attachments)
		if !ok || record.LastMutationID == nil || *record.LastMutationID != request.MutationID || record.PendingMutationKind == nil || *record.PendingMutationKind != request.Kind {
			return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
		}
		freshDatabaseWallClock, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return mailboxEntDraftRepositoryError(err)
		}
		if !mailboxEntDraftCreateFenceMatches(account, request.Fence, freshDatabaseWallClock) {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorFenceLost)
		}
		updated, kind := mailboxEntDraftReconciliationFinishTransition(ctx, draft, record, request, freshDatabaseWallClock)
		if kind != "" {
			return mailboxEntDraftReconciliationLogicalError(ctx, account, request.Fence, freshDatabaseWallClock, kind)
		}
		action, operation, status, audit := mailboxEntDraftReconciliationAuditValues(updated, request)
		if audit {
			if _, err := transaction.AuditLog.Create().
				SetAction(action).
				SetEntityType("mail_draft").
				SetEntityID(updated.ID.String()).
				SetUserID(account.UserID).
				SetDetails(map[string]interface{}{
					"organization_id": account.OrganizationID.String(),
					"mail_account_id": account.ID.String(),
					"operation":       operation,
					"status":          status,
				}).
				SetIPAddress("").
				SetUserAgent("").
				SetCreatedAt(freshDatabaseWallClock).
				Save(ctx); err != nil {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
			}
		}
		if err := mailboxEntDraftCreateRelease(ctx, account, request.Fence, freshDatabaseWallClock); err != nil {
			return err
		}
		result = MailboxDraftFinishResult{Draft: mailboxEntDraftRecordCopy(updated)}
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
	return mailboxEntDraftEffectiveStateKind(request.Confirmed)
}

func mailboxEntDraftReconciliationResumeIdentityMatches(draft *ent.MailDraft, record MailboxDraftRecord, request MailboxDraftResumeCreateRequest, engineDraftKey string) bool {
	return draft != nil && bytes.Equal(draft.CreateIdempotencyKeyHash, request.CreateIdempotencyKeyHash[:]) && record.ID == request.DraftID && record.AccountID == request.Fence.AccountID && record.DraftClientID == request.DraftClientID && record.EngineDraftKey == engineDraftKey && record.CreateRequestHash == request.CreateRequestHash && mailboxEntDraftEffectiveStatesEqual(record.Confirmed, request.Confirmed)
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
		return MailboxDraftRepositoryErrorInvalid
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
	updated := mailboxEntDraftRecordCopy(record)
	update := draft.Update().SetUpdatedAt(databaseWallClock)
	switch request.Kind {
	case MailboxDraftMutationCreate:
		if record.Revision != 1 || record.EngineMessageID != nil || record.EngineVersion != nil || record.PendingSave != nil {
			return MailboxDraftRecord{}, MailboxDraftRepositoryErrorStateConflict
		}
		switch outcome := request.Outcome.(type) {
		case MailboxDraftApplied:
			if err := update.SetStatus(maildraft.StatusActive).SetEngineMessageID(outcome.EngineMessageID).SetEngineVersion(outcome.EngineVersion).SetRevision(1).SetConfirmedMutationID(request.MutationID).ClearLastMutationID().ClearPendingMutationKind().Exec(ctx); err != nil {
				return MailboxDraftRecord{}, MailboxDraftRepositoryErrorUnavailable
			}
			updated.Status = MailboxDraftStatusActive
			updated.EngineMessageID = mailboxEntStringPointer(&outcome.EngineMessageID)
			updated.EngineVersion = mailboxEntStringPointer(&outcome.EngineVersion)
			updated.ConfirmedMutationID = mailboxEntDraftReadUUIDPointerCopy(&request.MutationID)
			updated.LastMutationID = nil
			updated.PendingMutationKind = nil
			updated.UpdatedAt = databaseWallClock.UTC()
			return updated, ""
		case MailboxDraftUnknown:
			if err := update.SetStatus(maildraft.StatusUnknown).Exec(ctx); err != nil {
				return MailboxDraftRecord{}, MailboxDraftRepositoryErrorUnavailable
			}
			updated.Status = MailboxDraftStatusUnknown
			updated.UpdatedAt = databaseWallClock.UTC()
			return updated, ""
		case MailboxDraftSaveRetained:
			return MailboxDraftRecord{}, MailboxDraftRepositoryErrorInvalid
		default:
			return MailboxDraftRecord{}, MailboxDraftRepositoryErrorStateConflict
		}
	case MailboxDraftMutationSave:
		if record.PendingSave == nil {
			return MailboxDraftRecord{}, MailboxDraftRepositoryErrorStateConflict
		}
		switch outcome := request.Outcome.(type) {
		case MailboxDraftApplied:
			nextRevision, ok := mailboxEntNextEpoch(record.Revision)
			if !ok {
				return MailboxDraftRecord{}, MailboxDraftRepositoryErrorStateConflict
			}
			removed := mailboxEntDraftConfirmedAttachmentDifference(record.Confirmed.AttachmentIDs, record.PendingSave.EffectiveState.AttachmentIDs)
			update.SetStatus(maildraft.StatusActive).SetEngineMessageID(outcome.EngineMessageID).SetEngineVersion(outcome.EngineVersion).SetRevision(nextRevision).SetFromIdentityID(record.PendingSave.EffectiveState.FromIdentityID).SetConfirmedMutationID(request.MutationID).ClearLastMutationID().ClearPendingMutationKind().ClearPendingFromIdentityID().ClearPendingReplyToMessageID().ClearPendingForwardMessageID().ClearPendingSaveRequestHash()
			if record.PendingSave.EffectiveState.ReplyToMessageID == nil {
				update.ClearReplyToMessageID()
			} else {
				update.SetReplyToMessageID(*record.PendingSave.EffectiveState.ReplyToMessageID)
			}
			if record.PendingSave.EffectiveState.ForwardMessageID == nil {
				update.ClearForwardMessageID()
			} else {
				update.SetForwardMessageID(*record.PendingSave.EffectiveState.ForwardMessageID)
			}
			if len(record.PendingSave.EffectiveState.AttachmentIDs) != 0 {
				update.AddAttachmentIDs(record.PendingSave.EffectiveState.AttachmentIDs...)
			}
			update.ClearPendingAttachments()
			if len(removed) != 0 {
				update.RemoveAttachmentIDs(removed...)
			}
			if err := update.Exec(ctx); err != nil {
				return MailboxDraftRecord{}, MailboxDraftRepositoryErrorUnavailable
			}
			updated.Status = MailboxDraftStatusActive
			updated.EngineMessageID = mailboxEntStringPointer(&outcome.EngineMessageID)
			updated.EngineVersion = mailboxEntStringPointer(&outcome.EngineVersion)
			updated.Revision = nextRevision
			updated.Confirmed = mailboxEntDraftEffectiveStateCopy(record.PendingSave.EffectiveState)
			updated.PendingSave = nil
			updated.ConfirmedMutationID = mailboxEntDraftReadUUIDPointerCopy(&request.MutationID)
			updated.LastMutationID = nil
			updated.PendingMutationKind = nil
			updated.UpdatedAt = databaseWallClock.UTC()
			return updated, ""
		case MailboxDraftUnknown:
			if err := update.SetStatus(maildraft.StatusUnknown).Exec(ctx); err != nil {
				return MailboxDraftRecord{}, MailboxDraftRepositoryErrorUnavailable
			}
			updated.Status = MailboxDraftStatusUnknown
			updated.UpdatedAt = databaseWallClock.UTC()
			return updated, ""
		case MailboxDraftSaveRetained:
			return MailboxDraftRecord{}, MailboxDraftRepositoryErrorInvalid
		default:
			return MailboxDraftRecord{}, MailboxDraftRepositoryErrorStateConflict
		}
	case MailboxDraftMutationDelete:
		if record.PendingSave != nil {
			return MailboxDraftRecord{}, MailboxDraftRepositoryErrorStateConflict
		}
		switch request.Outcome.(type) {
		case MailboxDraftDeleted:
			if len(record.Confirmed.AttachmentIDs) != 0 {
				update.RemoveAttachmentIDs(record.Confirmed.AttachmentIDs...)
			}
			if err := update.SetStatus(maildraft.StatusDeleted).ClearLastMutationID().ClearPendingMutationKind().Exec(ctx); err != nil {
				return MailboxDraftRecord{}, MailboxDraftRepositoryErrorUnavailable
			}
			updated.Status = MailboxDraftStatusDeleted
			updated.Confirmed.AttachmentIDs = []uuid.UUID{}
			updated.LastMutationID = nil
			updated.PendingMutationKind = nil
			updated.UpdatedAt = databaseWallClock.UTC()
			return updated, ""
		case MailboxDraftDeleteRetained:
			if err := update.SetStatus(maildraft.StatusActive).ClearLastMutationID().ClearPendingMutationKind().Exec(ctx); err != nil {
				return MailboxDraftRecord{}, MailboxDraftRepositoryErrorUnavailable
			}
			updated.Status = MailboxDraftStatusActive
			updated.LastMutationID = nil
			updated.PendingMutationKind = nil
			updated.UpdatedAt = databaseWallClock.UTC()
			return updated, ""
		case MailboxDraftUnknown:
			if err := update.SetStatus(maildraft.StatusUnknown).Exec(ctx); err != nil {
				return MailboxDraftRecord{}, MailboxDraftRepositoryErrorUnavailable
			}
			updated.Status = MailboxDraftStatusUnknown
			updated.UpdatedAt = databaseWallClock.UTC()
			return updated, ""
		case MailboxDraftSaveRetained:
			return MailboxDraftRecord{}, MailboxDraftRepositoryErrorInvalid
		default:
			return MailboxDraftRecord{}, MailboxDraftRepositoryErrorStateConflict
		}
	default:
		return MailboxDraftRecord{}, MailboxDraftRepositoryErrorInvalid
	}
}

func mailboxEntDraftReconciliationAuditValues(updated MailboxDraftRecord, request MailboxDraftFinishRequest) (action, operation, status string, ok bool) {
	switch request.Kind {
	case MailboxDraftMutationCreate:
		if _, applied := request.Outcome.(MailboxDraftApplied); applied && updated.Status == MailboxDraftStatusActive {
			return "create", "create", "active", true
		}
	case MailboxDraftMutationSave:
		if _, applied := request.Outcome.(MailboxDraftApplied); applied && updated.Status == MailboxDraftStatusActive {
			return "update", "save", "active", true
		}
	case MailboxDraftMutationDelete:
		switch request.Outcome.(type) {
		case MailboxDraftDeleted:
			if updated.Status == MailboxDraftStatusDeleted {
				return "delete", "delete", "deleted", true
			}
		case MailboxDraftDeleteRetained:
			if updated.Status == MailboxDraftStatusActive {
				return "delete", "delete", "active", true
			}
		}
	}
	return "", "", "", false
}

func mailboxEntDraftConfirmedAttachmentDifference(confirmed, desired []uuid.UUID) []uuid.UUID {
	desiredSet := make(map[uuid.UUID]struct{}, len(desired))
	for _, id := range desired {
		desiredSet[id] = struct{}{}
	}
	removed := make([]uuid.UUID, 0, len(confirmed))
	for _, id := range confirmed {
		if _, exists := desiredSet[id]; !exists {
			removed = append(removed, id)
		}
	}
	return removed
}

func mailboxEntDraftReconciliationLogicalError(ctx context.Context, account *ent.MailAccount, fence MailboxEngineCommandFence, databaseWallClock time.Time, kind MailboxDraftRepositoryErrorKind) error {
	if kind == MailboxDraftRepositoryErrorUnavailable {
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
	}
	return mailboxEntDraftCreateReleaseError(ctx, account, fence, databaseWallClock, kind)
}
