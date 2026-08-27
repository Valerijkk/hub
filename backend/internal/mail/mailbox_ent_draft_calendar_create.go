package mail

import (
	"bytes"
	"context"
	"crypto/sha256"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/maildraft"
)

// ReserveMailboxDraftCalendarCreate persists Calendar MIME upload metadata
// before the Calendar create command is dispatched.
func (repository *MailboxEntRepository) ReserveMailboxDraftCalendarCreate(ctx context.Context, request mailboxDraftCalendarCreateReserveRequest) (MailboxDraftReservation, error) {
	if repository == nil || repository.client == nil || ctx == nil || !validMailboxDraftCalendarCreateReserveRequest(request) {
		return MailboxDraftReservation{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	engineDraftKey, ok := mailboxEntDraftCreateKey(request.Create.DraftID)
	if !ok {
		return MailboxDraftReservation{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	var result MailboxDraftReservation
	var postCommitErr error
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, databaseNow, err := mailboxEntDraftCreateLockAccount(ctx, transaction, request.Create.Fence)
		if err != nil {
			return err
		}
		if kind := mailboxEntDraftCreateAccountKind(account); kind != "" {
			return mailboxEntDraftCalendarCreateLogicalError(ctx, account, request.Create.Fence, databaseNow, kind, &postCommitErr)
		}
		rows, err := mailboxEntDraftCreateLockIdentityRows(ctx, transaction, request.Create)
		if err != nil {
			return err
		}
		if err := mailboxEntDraftCreateLockConfirmedIdentity(ctx, transaction, request.Create.Fence.AccountID, request.Create.Confirmed.FromIdentityID); err != nil {
			return mailboxEntDraftCalendarCreateLogicalError(ctx, account, request.Create.Fence, databaseNow, mailboxEntDraftCreateErrorKind(err), &postCommitErr)
		}
		if err := mailboxEntDraftCreateLockReferences(ctx, transaction, request.Create.Fence.AccountID, request.Create.Confirmed); err != nil {
			return mailboxEntDraftCalendarCreateLogicalError(ctx, account, request.Create.Fence, databaseNow, mailboxEntDraftCreateErrorKind(err), &postCommitErr)
		}
		if len(rows) != 0 {
			return mailboxEntDraftCalendarCreateClassifyExisting(ctx, transaction, account, databaseNow, request, engineDraftKey, rows, &result, &postCommitErr)
		}
		draft, err := transaction.MailDraft.Create().
			SetID(request.Create.DraftID).
			SetMailAccountID(request.Create.Fence.AccountID).
			SetFromIdentityID(request.Create.Confirmed.FromIdentityID).
			SetDraftClientID(request.Create.DraftClientID).
			SetEngineDraftKey(engineDraftKey).
			SetRevision(1).
			SetStatus(maildraft.StatusSaving).
			SetCreateIdempotencyKeyHash(append([]byte(nil), request.Create.CreateIdempotencyKeyHash[:]...)).
			SetCreateRequestHash(append([]byte(nil), request.Create.CreateRequestHash[:]...)).
			SetLastMutationID(request.Create.MutationID).
			SetPendingMutationKind(maildraft.PendingMutationKind(MailboxDraftMutationCreate)).
			SetPendingCreateMIMEKind(maildraft.PendingCreateMIMEKind(MailboxDraftCreateMIMEKindCalendar)).
			SetPendingCreateMIMEPhase(maildraft.PendingCreateMIMEPhase(MailboxDraftCreateMIMEPhaseUpload)).
			SetPendingCreateMIMEBindingDigest(append([]byte(nil), request.State.BindingDigest[:]...)).
			SetPendingCreateMIMEAttempt(1).
			SetCreatedAt(databaseNow).
			SetUpdatedAt(databaseNow).
			Save(ctx)
		if err != nil {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		record, ok := mailboxEntDraftCalendarCreateRecord(draft, mailboxEntDraftAttachmentState{ConfirmedAttachmentIDs: []uuid.UUID{}, PendingAttachmentIDs: []uuid.UUID{}})
		if !ok || !mailboxEntDraftCalendarCreateMatches(record, MailboxDraftStatusSaving, MailboxDraftCreateMIMEPhaseUpload, request.Create.MutationID, 1, request.State.BindingDigest, "", 0) {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		work := mailboxEntDraftCreateWork(record, MailboxDraftWorkCreate)
		if work == nil {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		fence := mailboxEntDraftCreateFenceCopy(request.Create.Fence)
		result = MailboxDraftReservation{Kind: MailboxDraftReservationDispatch, Draft: mailboxEntDraftRecordCopy(record), Work: work, Fence: &fence}
		return nil
	})
	if err != nil {
		return MailboxDraftReservation{}, mailboxEntDraftRepositoryError(err)
	}
	if postCommitErr != nil {
		return MailboxDraftReservation{}, postCommitErr
	}
	return result, nil
}

// AdvanceMailboxDraftCalendarCreateToImport stores the exact upload handoff.
func (repository *MailboxEntRepository) AdvanceMailboxDraftCalendarCreateToImport(ctx context.Context, request mailboxDraftCalendarCreateAdvanceToImportRequest) (MailboxDraftReservation, error) {
	if repository == nil || repository.client == nil || ctx == nil || !validMailboxDraftCalendarCreateAdvanceToImportRequest(request) {
		return MailboxDraftReservation{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	var result MailboxDraftReservation
	var postCommitErr error
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		locked, kind := mailboxEntDraftCalendarCreateLock(ctx, transaction, request.Fence, request.DraftID)
		if kind != "" {
			return mailboxEntDraftCalendarCreateLockedError(ctx, locked, request.Fence, kind, &postCommitErr)
		}
		if !mailboxEntDraftCalendarCreateMatches(locked.record, MailboxDraftStatusSaving, MailboxDraftCreateMIMEPhaseUpload, request.MutationID, 1, request.ExpectedBindingDigest, "", 0) {
			return mailboxEntDraftCalendarCreateLogicalError(ctx, locked.account, request.Fence, locked.databaseNow, MailboxDraftRepositoryErrorStateConflict, &postCommitErr)
		}
		if err := locked.draft.Update().
			SetPendingCreateMIMEPhase(maildraft.PendingCreateMIMEPhase(MailboxDraftCreateMIMEPhaseImport)).
			SetPendingCreateMIMEBlobID(request.BlobID).
			SetPendingCreateMIMESizeBytes(request.SizeBytes).
			ClearPendingCreateMIMERetryAt().
			ClearPendingCreateMIMEExhaustedAt().
			SetUpdatedAt(locked.databaseNow).
			Exec(ctx); err != nil {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		updated := mailboxEntDraftCalendarCreateWithMIME(locked.record, MailboxDraftStatusSaving, MailboxDraftCreateMIMEPhaseImport, request.MutationID, 1, request.ExpectedBindingDigest, &request.BlobID, &request.SizeBytes, nil, nil, locked.databaseNow)
		work := mailboxEntDraftCreateWork(updated, MailboxDraftWorkCreate)
		if !mailboxEntDraftCalendarCreateMatches(updated, MailboxDraftStatusSaving, MailboxDraftCreateMIMEPhaseImport, request.MutationID, 1, request.ExpectedBindingDigest, request.BlobID, request.SizeBytes) || work == nil {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		fence := mailboxEntDraftCreateFenceCopy(request.Fence)
		result = MailboxDraftReservation{Kind: MailboxDraftReservationDispatch, Draft: mailboxEntDraftRecordCopy(updated), Work: work, Fence: &fence}
		return nil
	})
	if err != nil {
		return MailboxDraftReservation{}, mailboxEntDraftRepositoryError(err)
	}
	if postCommitErr != nil {
		return MailboxDraftReservation{}, postCommitErr
	}
	return result, nil
}

// MarkMailboxDraftCalendarCreateUnknown preserves only the exact uncertain
// Calendar stage and derives upload retry state from database time.
func (repository *MailboxEntRepository) MarkMailboxDraftCalendarCreateUnknown(ctx context.Context, request mailboxDraftCalendarCreateMarkUnknownRequest) (MailboxDraftRecord, error) {
	if repository == nil || repository.client == nil || ctx == nil || !validMailboxDraftCalendarCreateMarkUnknownRequest(request) {
		return MailboxDraftRecord{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	var result MailboxDraftRecord
	var postCommitErr error
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		locked, kind := mailboxEntDraftCalendarCreateLock(ctx, transaction, request.Fence, request.DraftID)
		if kind != "" {
			return mailboxEntDraftCalendarCreateLockedError(ctx, locked, request.Fence, kind, &postCommitErr)
		}
		phase, ok := mailboxEntDraftCalendarCreateUnknownStageMatches(locked.record, request)
		if !ok || locked.record.PendingCreateMIME == nil {
			return mailboxEntDraftCalendarCreateLogicalError(ctx, locked.account, request.Fence, locked.databaseNow, MailboxDraftRepositoryErrorStateConflict, &postCommitErr)
		}
		state := locked.record.PendingCreateMIME
		var retryAt, exhaustedAt *time.Time
		if phase == MailboxDraftCreateMIMEPhaseUpload {
			var scheduled bool
			retryAt, exhaustedAt, scheduled = mailboxEntDraftCalendarCreateUnknownUploadSchedule(*state, locked.databaseNow)
			if !scheduled {
				return mailboxEntDraftCalendarCreateLogicalError(ctx, locked.account, request.Fence, locked.databaseNow, MailboxDraftRepositoryErrorStateConflict, &postCommitErr)
			}
		}
		update := locked.draft.Update().SetStatus(maildraft.StatusUnknown).SetUpdatedAt(locked.databaseNow)
		mailboxEntDraftCalendarCreateApplyMIMERetry(update, retryAt, exhaustedAt)
		if err := update.Exec(ctx); err != nil {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		updated := mailboxEntDraftCalendarCreateWithMIME(locked.record, MailboxDraftStatusUnknown, phase, request.MutationID, request.ExpectedAttempt, request.ExpectedBindingDigest, state.BlobID, state.SizeBytes, retryAt, exhaustedAt, locked.databaseNow)
		if !mailboxEntDraftCalendarCreateUnknownPostStateValid(updated, request, phase) {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		if err := mailboxEntDraftCreateRelease(ctx, locked.account, request.Fence, locked.databaseNow); err != nil {
			return err
		}
		result = mailboxEntDraftRecordCopy(updated)
		return nil
	})
	if err != nil {
		return MailboxDraftRecord{}, mailboxEntDraftRepositoryError(err)
	}
	if postCommitErr != nil {
		return MailboxDraftRecord{}, postCommitErr
	}
	return result, nil
}

// FinishMailboxDraftCalendarCreateApplied activates one exact Calendar import.
func (repository *MailboxEntRepository) FinishMailboxDraftCalendarCreateApplied(ctx context.Context, request mailboxDraftCalendarCreateFinishAppliedRequest) (MailboxDraftRecord, error) {
	if repository == nil || repository.client == nil || ctx == nil || !validMailboxDraftCalendarCreateFinishAppliedRequest(request) {
		return MailboxDraftRecord{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	var result MailboxDraftRecord
	var postCommitErr error
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		locked, kind := mailboxEntDraftCalendarCreateLock(ctx, transaction, request.Fence, request.DraftID)
		if kind != "" {
			return mailboxEntDraftCalendarCreateLockedError(ctx, locked, request.Fence, kind, &postCommitErr)
		}
		state := locked.record.PendingCreateMIME
		if state == nil || !mailboxEntDraftCalendarCreateMatches(locked.record, MailboxDraftStatusSaving, MailboxDraftCreateMIMEPhaseImport, request.MutationID, state.Attempt, request.ExpectedBindingDigest, request.ExpectedBlobID, request.ExpectedSizeBytes) {
			return mailboxEntDraftCalendarCreateLogicalError(ctx, locked.account, request.Fence, locked.databaseNow, MailboxDraftRepositoryErrorStateConflict, &postCommitErr)
		}
		if err := locked.draft.Update().
			SetStatus(maildraft.StatusActive).
			SetEngineMessageID(request.Applied.EngineMessageID).
			SetEngineVersion(request.Applied.EngineVersion).
			SetRevision(1).
			SetConfirmedMutationID(request.MutationID).
			ClearLastMutationID().
			ClearPendingMutationKind().
			ClearPendingCreateMIMEKind().
			ClearPendingCreateMIMEPhase().
			ClearPendingCreateMIMEBlobID().
			ClearPendingCreateMIMESizeBytes().
			ClearPendingCreateMIMEBindingDigest().
			SetPendingCreateMIMEAttempt(0).
			ClearPendingCreateMIMERetryAt().
			ClearPendingCreateMIMEExhaustedAt().
			ClearPendingCreateAttachmentPhase().
			ClearPendingCreateUploadBlobID().
			ClearPendingCreateUploadSizeBytes().
			SetPendingCreateUploadAttempt(0).
			ClearPendingCreateUploadRetryAt().
			ClearPendingCreateUploadExhaustedAt().
			SetUpdatedAt(locked.databaseNow).
			Exec(ctx); err != nil {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		updated := mailboxEntDraftRecordCopy(locked.record)
		updated.Status = MailboxDraftStatusActive
		updated.EngineMessageID = mailboxEntStringPointer(&request.Applied.EngineMessageID)
		updated.EngineVersion = mailboxEntStringPointer(&request.Applied.EngineVersion)
		updated.Revision = 1
		updated.ConfirmedMutationID = mailboxEntDraftReadUUIDPointerCopy(&request.MutationID)
		updated.LastMutationID = nil
		updated.PendingMutationKind = nil
		updated.PendingCreateMIME = nil
		updated.PendingCreateAttachmentPhase = nil
		updated.PendingCreateUploadBlobID = nil
		updated.PendingCreateUploadSizeBytes = nil
		updated.PendingCreateUploadAttempt = 0
		updated.PendingCreateUploadRetryAt = nil
		updated.PendingCreateUploadExhaustedAt = nil
		updated.UpdatedAt = locked.databaseNow.UTC()
		if !mailboxEntDraftCalendarCreateTerminalRecordValid(updated, request.MutationID) {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		if err := mailboxEntDraftCreateRelease(ctx, locked.account, request.Fence, locked.databaseNow); err != nil {
			return err
		}
		result = mailboxEntDraftRecordCopy(updated)
		return nil
	})
	if err != nil {
		return MailboxDraftRecord{}, mailboxEntDraftRepositoryError(err)
	}
	if postCommitErr != nil {
		return MailboxDraftRecord{}, postCommitErr
	}
	return result, nil
}

// RestartMailboxDraftCalendarCreateUpload rotates an exact due unknown upload.
func (repository *MailboxEntRepository) RestartMailboxDraftCalendarCreateUpload(ctx context.Context, request mailboxDraftCalendarCreateRestartUploadRequest) (MailboxDraftReservation, error) {
	if repository == nil || repository.client == nil || ctx == nil || !validMailboxDraftCalendarCreateRestartUploadRequest(request) {
		return MailboxDraftReservation{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	var result MailboxDraftReservation
	var postCommitErr error
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		locked, kind := mailboxEntDraftCalendarCreateLock(ctx, transaction, request.Fence, request.DraftID)
		if kind != "" {
			return mailboxEntDraftCalendarCreateLockedError(ctx, locked, request.Fence, kind, &postCommitErr)
		}
		if !mailboxEntDraftCalendarCreateHashesMatch(locked.draft, locked.record, request.IdempotencyKeyHash, request.RequestHash) || !mailboxEntDraftCalendarCreateRestartableUnknown(locked.record, request.ExpectedMutationID, request.ExpectedAttempt, request.ExpectedBindingDigest, locked.databaseNow) {
			return mailboxEntDraftCalendarCreateLogicalError(ctx, locked.account, request.Fence, locked.databaseNow, MailboxDraftRepositoryErrorStateConflict, &postCommitErr)
		}
		nextAttempt := request.ExpectedAttempt + 1
		if err := locked.draft.Update().
			SetStatus(maildraft.StatusSaving).
			SetLastMutationID(request.NextMutationID).
			SetPendingMutationKind(maildraft.PendingMutationKind(MailboxDraftMutationCreate)).
			SetPendingCreateMIMEKind(maildraft.PendingCreateMIMEKind(MailboxDraftCreateMIMEKindCalendar)).
			SetPendingCreateMIMEPhase(maildraft.PendingCreateMIMEPhase(MailboxDraftCreateMIMEPhaseUpload)).
			ClearPendingCreateMIMEBlobID().
			ClearPendingCreateMIMESizeBytes().
			SetPendingCreateMIMEBindingDigest(append([]byte(nil), request.NextBindingDigest[:]...)).
			SetPendingCreateMIMEAttempt(nextAttempt).
			ClearPendingCreateMIMERetryAt().
			ClearPendingCreateMIMEExhaustedAt().
			SetUpdatedAt(locked.databaseNow).
			Exec(ctx); err != nil {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		updated := mailboxEntDraftCalendarCreateWithMIME(locked.record, MailboxDraftStatusSaving, MailboxDraftCreateMIMEPhaseUpload, request.NextMutationID, nextAttempt, request.NextBindingDigest, nil, nil, nil, nil, locked.databaseNow)
		work := mailboxEntDraftCreateWork(updated, MailboxDraftWorkCreate)
		if !mailboxEntDraftCalendarCreateMatches(updated, MailboxDraftStatusSaving, MailboxDraftCreateMIMEPhaseUpload, request.NextMutationID, nextAttempt, request.NextBindingDigest, "", 0) || work == nil {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		fence := mailboxEntDraftCreateFenceCopy(request.Fence)
		result = MailboxDraftReservation{Kind: MailboxDraftReservationDispatch, Draft: mailboxEntDraftRecordCopy(updated), Work: work, Fence: &fence}
		return nil
	})
	if err != nil {
		return MailboxDraftReservation{}, mailboxEntDraftRepositoryError(err)
	}
	if postCommitErr != nil {
		return MailboxDraftReservation{}, postCommitErr
	}
	return result, nil
}

// RestartMailboxDraftCalendarCreateUploadAfterBlobNotFound accepts only an
// exact saving/import handoff and rotates it into a fresh upload attempt.
func (repository *MailboxEntRepository) RestartMailboxDraftCalendarCreateUploadAfterBlobNotFound(ctx context.Context, request mailboxDraftCalendarCreateRestartAfterBlobNotFoundRequest) (MailboxDraftReservation, error) {
	if repository == nil || repository.client == nil || ctx == nil || !validMailboxDraftCalendarCreateRestartAfterBlobNotFoundRequest(request) {
		return MailboxDraftReservation{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	var result MailboxDraftReservation
	var postCommitErr error
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		locked, kind := mailboxEntDraftCalendarCreateLock(ctx, transaction, request.Fence, request.DraftID)
		if kind != "" {
			return mailboxEntDraftCalendarCreateLockedError(ctx, locked, request.Fence, kind, &postCommitErr)
		}
		if !mailboxEntDraftCalendarCreateHashesMatch(locked.draft, locked.record, request.IdempotencyKeyHash, request.RequestHash) || !mailboxEntDraftCalendarCreateMatches(locked.record, MailboxDraftStatusSaving, MailboxDraftCreateMIMEPhaseImport, request.ExpectedMutationID, request.ExpectedAttempt, request.ExpectedBindingDigest, request.ExpectedBlobID, request.ExpectedSizeBytes) {
			return mailboxEntDraftCalendarCreateLogicalError(ctx, locked.account, request.Fence, locked.databaseNow, MailboxDraftRepositoryErrorStateConflict, &postCommitErr)
		}
		nextAttempt := request.ExpectedAttempt + 1
		if err := locked.draft.Update().
			SetStatus(maildraft.StatusSaving).
			SetLastMutationID(request.NextMutationID).
			SetPendingMutationKind(maildraft.PendingMutationKind(MailboxDraftMutationCreate)).
			SetPendingCreateMIMEKind(maildraft.PendingCreateMIMEKind(MailboxDraftCreateMIMEKindCalendar)).
			SetPendingCreateMIMEPhase(maildraft.PendingCreateMIMEPhase(MailboxDraftCreateMIMEPhaseUpload)).
			ClearPendingCreateMIMEBlobID().
			ClearPendingCreateMIMESizeBytes().
			SetPendingCreateMIMEBindingDigest(append([]byte(nil), request.NextBindingDigest[:]...)).
			SetPendingCreateMIMEAttempt(nextAttempt).
			ClearPendingCreateMIMERetryAt().
			ClearPendingCreateMIMEExhaustedAt().
			SetUpdatedAt(locked.databaseNow).
			Exec(ctx); err != nil {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		updated := mailboxEntDraftCalendarCreateWithMIME(locked.record, MailboxDraftStatusSaving, MailboxDraftCreateMIMEPhaseUpload, request.NextMutationID, nextAttempt, request.NextBindingDigest, nil, nil, nil, nil, locked.databaseNow)
		work := mailboxEntDraftCreateWork(updated, MailboxDraftWorkCreate)
		if !mailboxEntDraftCalendarCreateMatches(updated, MailboxDraftStatusSaving, MailboxDraftCreateMIMEPhaseUpload, request.NextMutationID, nextAttempt, request.NextBindingDigest, "", 0) || work == nil {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		fence := mailboxEntDraftCreateFenceCopy(request.Fence)
		result = MailboxDraftReservation{Kind: MailboxDraftReservationDispatch, Draft: mailboxEntDraftRecordCopy(updated), Work: work, Fence: &fence}
		return nil
	})
	if err != nil {
		return MailboxDraftReservation{}, mailboxEntDraftRepositoryError(err)
	}
	if postCommitErr != nil {
		return MailboxDraftReservation{}, postCommitErr
	}
	return result, nil
}

// AbandonMailboxDraftCalendarCreateUpload writes a terminal local Calendar
// tombstone only after the exact unknown upload state was observed.
func (repository *MailboxEntRepository) AbandonMailboxDraftCalendarCreateUpload(ctx context.Context, request mailboxDraftCalendarCreateAbandonRequest) (mailboxDraftCalendarCreateAbandonResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || !validMailboxDraftCalendarCreateAbandonRequest(request) {
		return mailboxDraftCalendarCreateAbandonResult{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	var result mailboxDraftCalendarCreateAbandonResult
	var postCommitErr error
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		locked, kind := mailboxEntDraftCalendarCreateLock(ctx, transaction, request.Fence, request.DraftID)
		if kind != "" {
			return mailboxEntDraftCalendarCreateLockedError(ctx, locked, request.Fence, kind, &postCommitErr)
		}
		if !mailboxEntDraftCalendarCreateHashesMatch(locked.draft, locked.record, request.IdempotencyKeyHash, request.RequestHash) {
			return mailboxEntDraftCalendarCreateLogicalError(ctx, locked.account, request.Fence, locked.databaseNow, MailboxDraftRepositoryErrorIdempotencyConflict, &postCommitErr)
		}
		if locked.record.Status == MailboxDraftStatusAbandoned && locked.record.AbandonedMutationID != nil && *locked.record.AbandonedMutationID == request.ExpectedMutationID && locked.record.PendingCreateMIME == nil {
			if err := mailboxEntDraftCreateRelease(ctx, locked.account, request.Fence, locked.databaseNow); err != nil {
				return err
			}
			result = mailboxDraftCalendarCreateAbandonResult{Replayed: true, Draft: mailboxEntDraftRecordCopy(locked.record)}
			return nil
		}
		if !mailboxEntDraftCalendarCreateAbandonableUnknown(locked.record, request.ExpectedMutationID, request.ExpectedAttempt, request.ExpectedBindingDigest) {
			return mailboxEntDraftCalendarCreateLogicalError(ctx, locked.account, request.Fence, locked.databaseNow, MailboxDraftRepositoryErrorStateConflict, &postCommitErr)
		}
		if err := locked.draft.Update().
			SetStatus(maildraft.StatusAbandoned).
			ClearEngineMessageID().
			ClearEngineVersion().
			ClearLastMutationID().
			ClearConfirmedMutationID().
			ClearPendingMutationKind().
			ClearPendingFromIdentityID().
			ClearPendingReplyToMessageID().
			ClearPendingForwardMessageID().
			ClearPendingSaveRequestHash().
			ClearPendingCreateMIMEKind().
			ClearPendingCreateMIMEPhase().
			ClearPendingCreateMIMEBlobID().
			ClearPendingCreateMIMESizeBytes().
			ClearPendingCreateMIMEBindingDigest().
			SetPendingCreateMIMEAttempt(0).
			ClearPendingCreateMIMERetryAt().
			ClearPendingCreateMIMEExhaustedAt().
			ClearPendingCreateAttachmentPhase().
			ClearPendingCreateUploadBlobID().
			ClearPendingCreateUploadSizeBytes().
			SetPendingCreateUploadAttempt(0).
			ClearPendingCreateUploadRetryAt().
			ClearPendingCreateUploadExhaustedAt().
			SetAbandonedAt(locked.databaseNow).
			SetAbandonedMutationID(request.ExpectedMutationID).
			SetUpdatedAt(locked.databaseNow).
			Exec(ctx); err != nil {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		updated := mailboxEntDraftRecordCopy(locked.record)
		updated.Status = MailboxDraftStatusAbandoned
		updated.EngineMessageID = nil
		updated.EngineVersion = nil
		updated.Confirmed.AttachmentIDs = []uuid.UUID{}
		updated.PendingSave = nil
		updated.LastMutationID = nil
		updated.ConfirmedMutationID = nil
		updated.PendingMutationKind = nil
		updated.PendingCreateMIME = nil
		updated.PendingCreateAttachmentPhase = nil
		updated.PendingCreateUploadBlobID = nil
		updated.PendingCreateUploadSizeBytes = nil
		updated.PendingCreateUploadAttempt = 0
		updated.PendingCreateUploadRetryAt = nil
		updated.PendingCreateUploadExhaustedAt = nil
		updated.AbandonedAt = mailboxEntDraftReadTimePointerCopy(&locked.databaseNow)
		updated.AbandonedMutationID = mailboxEntDraftReadUUIDPointerCopy(&request.ExpectedMutationID)
		updated.UpdatedAt = locked.databaseNow.UTC()
		if !mailboxEntDraftCalendarCreateAbandonedRecordValid(updated, request.ExpectedMutationID) {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		if err := mailboxEntDraftCreateRelease(ctx, locked.account, request.Fence, locked.databaseNow); err != nil {
			return err
		}
		result = mailboxDraftCalendarCreateAbandonResult{Draft: mailboxEntDraftRecordCopy(updated)}
		return nil
	})
	if err != nil {
		return mailboxDraftCalendarCreateAbandonResult{}, mailboxEntDraftRepositoryError(err)
	}
	if postCommitErr != nil {
		return mailboxDraftCalendarCreateAbandonResult{}, postCommitErr
	}
	return result, nil
}

type mailboxEntDraftCalendarCreateLocked struct {
	account     *ent.MailAccount
	databaseNow time.Time
	draft       *ent.MailDraft
	record      MailboxDraftRecord
}

func mailboxEntDraftCalendarCreateLock(ctx context.Context, transaction *ent.Tx, fence MailboxEngineCommandFence, draftID uuid.UUID) (*mailboxEntDraftCalendarCreateLocked, MailboxDraftRepositoryErrorKind) {
	account, databaseNow, err := mailboxEntDraftCreateLockAccount(ctx, transaction, fence)
	if err != nil {
		return nil, mailboxEntDraftCreateErrorKind(err)
	}
	locked := &mailboxEntDraftCalendarCreateLocked{account: account, databaseNow: databaseNow}
	if kind := mailboxEntDraftCreateAccountKind(account); kind != "" {
		return locked, kind
	}
	draft, kind := mailboxEntDraftReservationLockDraft(ctx, transaction, fence.AccountID, draftID)
	if kind != "" {
		return locked, kind
	}
	attachments, kind := mailboxEntDraftReadAttachmentState(ctx, transaction, draft, MailboxDraftRepositoryErrorStateConflict)
	if kind != "" {
		return locked, kind
	}
	record, ok := mailboxEntDraftCalendarCreateRecord(draft, attachments)
	if !ok {
		return locked, MailboxDraftRepositoryErrorStateConflict
	}
	if err := mailboxEntDraftCreateLockConfirmedIdentity(ctx, transaction, fence.AccountID, record.Confirmed.FromIdentityID); err != nil {
		return locked, mailboxEntDraftCreateErrorKind(err)
	}
	locked.draft = draft
	locked.record = record
	return locked, ""
}

func mailboxEntDraftCalendarCreateLockedError(ctx context.Context, locked *mailboxEntDraftCalendarCreateLocked, fence MailboxEngineCommandFence, kind MailboxDraftRepositoryErrorKind, postCommitErr *error) error {
	if locked == nil || locked.account == nil {
		return NewMailboxDraftRepositoryError(kind)
	}
	return mailboxEntDraftCalendarCreateLogicalError(ctx, locked.account, fence, locked.databaseNow, kind, postCommitErr)
}

func mailboxEntDraftCalendarCreateLogicalError(ctx context.Context, account *ent.MailAccount, fence MailboxEngineCommandFence, databaseNow time.Time, kind MailboxDraftRepositoryErrorKind, postCommitErr *error) error {
	if postCommitErr == nil || kind == MailboxDraftRepositoryErrorUnavailable {
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
	}
	if err := mailboxEntDraftCreateRelease(ctx, account, fence, databaseNow); err != nil {
		return err
	}
	*postCommitErr = NewMailboxDraftRepositoryError(kind)
	return nil
}

func mailboxEntDraftCalendarCreateClassifyExisting(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, databaseNow time.Time, request mailboxDraftCalendarCreateReserveRequest, engineDraftKey string, rows []*ent.MailDraft, result *MailboxDraftReservation, postCommitErr *error) error {
	if result == nil || len(rows) != 1 {
		return mailboxEntDraftCalendarCreateLogicalError(ctx, account, request.Create.Fence, databaseNow, MailboxDraftRepositoryErrorIdempotencyConflict, postCommitErr)
	}
	draft := rows[0]
	attachments, kind := mailboxEntDraftReadAttachmentState(ctx, transaction, draft, MailboxDraftRepositoryErrorStateConflict)
	if kind != "" {
		return mailboxEntDraftCalendarCreateLogicalError(ctx, account, request.Create.Fence, databaseNow, kind, postCommitErr)
	}
	record, ok := mailboxEntDraftCalendarCreateRecord(draft, attachments)
	if !ok || !bytes.Equal(draft.CreateIdempotencyKeyHash, request.Create.CreateIdempotencyKeyHash[:]) || !mailboxEntDraftCreateSameRequest(record, request.Create, engineDraftKey) {
		return mailboxEntDraftCalendarCreateLogicalError(ctx, account, request.Create.Fence, databaseNow, MailboxDraftRepositoryErrorIdempotencyConflict, postCommitErr)
	}
	switch record.Status {
	case MailboxDraftStatusActive:
		if record.PendingCreateMIME != nil || record.ConfirmedMutationID == nil || *record.ConfirmedMutationID == uuid.Nil {
			return mailboxEntDraftCalendarCreateLogicalError(ctx, account, request.Create.Fence, databaseNow, MailboxDraftRepositoryErrorStateConflict, postCommitErr)
		}
		if err := mailboxEntDraftCreateRelease(ctx, account, request.Create.Fence, databaseNow); err != nil {
			return err
		}
		*result = MailboxDraftReservation{Kind: MailboxDraftReservationReplay, Draft: mailboxEntDraftRecordCopy(record)}
		return nil
	case MailboxDraftStatusSaving, MailboxDraftStatusUnknown:
		state := record.PendingCreateMIME
		if state == nil || state.Kind != MailboxDraftCreateMIMEKindCalendar || state.BindingDigest != request.State.BindingDigest || record.LastMutationID == nil || *record.LastMutationID == uuid.Nil {
			return mailboxEntDraftCalendarCreateLogicalError(ctx, account, request.Create.Fence, databaseNow, MailboxDraftRepositoryErrorStateConflict, postCommitErr)
		}
		work := mailboxEntDraftCreateWork(record, MailboxDraftWorkReconcileCreate)
		if work == nil {
			return mailboxEntDraftCalendarCreateLogicalError(ctx, account, request.Create.Fence, databaseNow, MailboxDraftRepositoryErrorStateConflict, postCommitErr)
		}
		fence := mailboxEntDraftCreateFenceCopy(request.Create.Fence)
		*result = MailboxDraftReservation{Kind: MailboxDraftReservationReconcileRequired, Draft: mailboxEntDraftRecordCopy(record), Work: work, Fence: &fence}
		return nil
	default:
		return mailboxEntDraftCalendarCreateLogicalError(ctx, account, request.Create.Fence, databaseNow, MailboxDraftRepositoryErrorStateConflict, postCommitErr)
	}
}

func mailboxEntDraftCalendarCreateRecord(draft *ent.MailDraft, attachments mailboxEntDraftAttachmentState) (MailboxDraftRecord, bool) {
	if draft == nil || len(attachments.ConfirmedAttachmentIDs) != 0 || len(attachments.PendingAttachmentIDs) != 0 || !mailboxEntDraftCalendarCreateLegacyIdle(draft) {
		return MailboxDraftRecord{}, false
	}
	record, ok := mailboxEntDraftReadRecord(draft, attachments)
	if !ok {
		return MailboxDraftRecord{}, false
	}
	state, ok := mailboxEntDraftCalendarCreateMIMEStateFromRow(draft)
	if !ok {
		return MailboxDraftRecord{}, false
	}
	record.PendingCreateMIME = state
	if !validMailboxDraftCreateMIMEStateBridge(record.PendingCreateMIME, record.Status, record.PendingMutationKind, record.PendingCreateAttachmentPhase, record.PendingCreateUploadBlobID, record.PendingCreateUploadSizeBytes, record.PendingCreateUploadAttempt, record.PendingCreateUploadRetryAt, record.PendingCreateUploadExhaustedAt) {
		return MailboxDraftRecord{}, false
	}
	return record, true
}

func mailboxEntDraftCalendarCreateLegacyIdle(draft *ent.MailDraft) bool {
	return draft != nil && draft.PendingCreateAttachmentPhase == nil && draft.PendingCreateUploadBlobID == nil && draft.PendingCreateUploadSizeBytes == nil && draft.PendingCreateUploadAttempt == 0 && draft.PendingCreateUploadRetryAt == nil && draft.PendingCreateUploadExhaustedAt == nil
}

func mailboxEntDraftCalendarCreateMIMEStateFromRow(draft *ent.MailDraft) (*MailboxDraftCreateMIMEState, bool) {
	if draft == nil {
		return nil, false
	}
	if draft.PendingCreateMIMEKind == nil && draft.PendingCreateMIMEPhase == nil && draft.PendingCreateMIMEBlobID == nil && draft.PendingCreateMIMESizeBytes == nil && draft.PendingCreateMIMEBindingDigest == nil && draft.PendingCreateMIMEAttempt == 0 && draft.PendingCreateMIMERetryAt == nil && draft.PendingCreateMIMEExhaustedAt == nil {
		return nil, true
	}
	if draft.PendingCreateMIMEKind == nil || draft.PendingCreateMIMEPhase == nil || draft.PendingCreateMIMEBindingDigest == nil || len(*draft.PendingCreateMIMEBindingDigest) != sha256.Size {
		return nil, false
	}
	state := &MailboxDraftCreateMIMEState{
		Kind:        MailboxDraftCreateMIMEKind(*draft.PendingCreateMIMEKind),
		Phase:       MailboxDraftCreateMIMEPhase(*draft.PendingCreateMIMEPhase),
		BlobID:      mailboxEntStringPointer(draft.PendingCreateMIMEBlobID),
		SizeBytes:   mailboxEntDraftReadInt64PointerCopy(draft.PendingCreateMIMESizeBytes),
		Attempt:     draft.PendingCreateMIMEAttempt,
		RetryAt:     mailboxEntDraftReadTimePointerCopy(draft.PendingCreateMIMERetryAt),
		ExhaustedAt: mailboxEntDraftReadTimePointerCopy(draft.PendingCreateMIMEExhaustedAt),
	}
	copy(state.BindingDigest[:], *draft.PendingCreateMIMEBindingDigest)
	return state, true
}

func mailboxEntDraftCalendarCreateMatches(record MailboxDraftRecord, status MailboxDraftStatus, phase MailboxDraftCreateMIMEPhase, mutationID uuid.UUID, attempt int, bindingDigest [sha256.Size]byte, blobID string, sizeBytes int64) bool {
	state := record.PendingCreateMIME
	if record.Status != status || record.Revision != 1 || record.EngineMessageID != nil || record.EngineVersion != nil || record.PendingSave != nil || record.PendingMutationKind == nil || *record.PendingMutationKind != MailboxDraftMutationCreate || record.LastMutationID == nil || *record.LastMutationID != mutationID || record.ConfirmedMutationID != nil || record.AbandonedAt != nil || record.AbandonedMutationID != nil || len(record.Confirmed.AttachmentIDs) != 0 || state == nil || state.Kind != MailboxDraftCreateMIMEKindCalendar || state.Phase != phase || state.Attempt != attempt || state.BindingDigest != bindingDigest || record.PendingCreateAttachmentPhase != nil || record.PendingCreateUploadBlobID != nil || record.PendingCreateUploadSizeBytes != nil || record.PendingCreateUploadAttempt != 0 || record.PendingCreateUploadRetryAt != nil || record.PendingCreateUploadExhaustedAt != nil {
		return false
	}
	switch phase {
	case MailboxDraftCreateMIMEPhaseUpload:
		return state.BlobID == nil && state.SizeBytes == nil
	case MailboxDraftCreateMIMEPhaseImport:
		return state.BlobID != nil && *state.BlobID == blobID && state.SizeBytes != nil && *state.SizeBytes == sizeBytes
	default:
		return false
	}
}

func mailboxEntDraftCalendarCreateUnknownStageMatches(record MailboxDraftRecord, request mailboxDraftCalendarCreateMarkUnknownRequest) (MailboxDraftCreateMIMEPhase, bool) {
	if record.Status != MailboxDraftStatusSaving {
		return "", false
	}
	switch request.Stage {
	case mailboxDraftCalendarCreateUnknownStageUpload:
		if mailboxEntDraftCalendarCreateMatches(record, MailboxDraftStatusSaving, MailboxDraftCreateMIMEPhaseUpload, request.MutationID, request.ExpectedAttempt, request.ExpectedBindingDigest, "", 0) && record.PendingCreateMIME.RetryAt == nil && record.PendingCreateMIME.ExhaustedAt == nil {
			return MailboxDraftCreateMIMEPhaseUpload, true
		}
	case mailboxDraftCalendarCreateUnknownStageAdvance:
		if mailboxEntDraftCalendarCreateMatches(record, MailboxDraftStatusSaving, MailboxDraftCreateMIMEPhaseUpload, request.MutationID, request.ExpectedAttempt, request.ExpectedBindingDigest, "", 0) && record.PendingCreateMIME.RetryAt == nil && record.PendingCreateMIME.ExhaustedAt == nil {
			return MailboxDraftCreateMIMEPhaseUpload, true
		}
		if mailboxEntDraftCalendarCreateMatches(record, MailboxDraftStatusSaving, MailboxDraftCreateMIMEPhaseImport, request.MutationID, request.ExpectedAttempt, request.ExpectedBindingDigest, request.BlobID, request.SizeBytes) {
			return MailboxDraftCreateMIMEPhaseImport, true
		}
	case mailboxDraftCalendarCreateUnknownStageImport:
		if mailboxEntDraftCalendarCreateMatches(record, MailboxDraftStatusSaving, MailboxDraftCreateMIMEPhaseImport, request.MutationID, request.ExpectedAttempt, request.ExpectedBindingDigest, request.BlobID, request.SizeBytes) {
			return MailboxDraftCreateMIMEPhaseImport, true
		}
	}
	return "", false
}

func mailboxEntDraftCalendarCreateUnknownUploadSchedule(state MailboxDraftCreateMIMEState, databaseNow time.Time) (*time.Time, *time.Time, bool) {
	if state.Kind != MailboxDraftCreateMIMEKindCalendar || state.Phase != MailboxDraftCreateMIMEPhaseUpload || state.BlobID != nil || state.SizeBytes != nil || state.RetryAt != nil || state.ExhaustedAt != nil || !mailboxEntValidTime(databaseNow) {
		return nil, nil, false
	}
	switch state.Attempt {
	case 1:
		retryAt, ok := mailboxEntCheckedTimeAdd(databaseNow, mailboxDraftCreateUploadRetryAfterFirstAttempt)
		if !ok {
			return nil, nil, false
		}
		return mailboxEntDraftReadTimePointerCopy(&retryAt), nil, true
	case 2:
		retryAt, ok := mailboxEntCheckedTimeAdd(databaseNow, mailboxDraftCreateUploadRetryAfterSecondAttempt)
		if !ok {
			return nil, nil, false
		}
		return mailboxEntDraftReadTimePointerCopy(&retryAt), nil, true
	case 3:
		return nil, mailboxEntDraftReadTimePointerCopy(&databaseNow), true
	default:
		return nil, nil, false
	}
}

func mailboxEntDraftCalendarCreateApplyMIMERetry(update *ent.MailDraftUpdateOne, retryAt, exhaustedAt *time.Time) {
	if update == nil {
		return
	}
	if retryAt == nil {
		update.ClearPendingCreateMIMERetryAt()
	} else {
		update.SetPendingCreateMIMERetryAt(*retryAt)
	}
	if exhaustedAt == nil {
		update.ClearPendingCreateMIMEExhaustedAt()
	} else {
		update.SetPendingCreateMIMEExhaustedAt(*exhaustedAt)
	}
}

func mailboxEntDraftCalendarCreateWithMIME(record MailboxDraftRecord, status MailboxDraftStatus, phase MailboxDraftCreateMIMEPhase, mutationID uuid.UUID, attempt int, bindingDigest [sha256.Size]byte, blobID *string, sizeBytes *int64, retryAt, exhaustedAt *time.Time, databaseNow time.Time) MailboxDraftRecord {
	updated := mailboxEntDraftRecordCopy(record)
	updated.Status = status
	updated.LastMutationID = mailboxEntDraftReadUUIDPointerCopy(&mutationID)
	kind := MailboxDraftMutationCreate
	updated.PendingMutationKind = &kind
	updated.PendingCreateMIME = &MailboxDraftCreateMIMEState{Kind: MailboxDraftCreateMIMEKindCalendar, Phase: phase, BlobID: mailboxEntStringPointer(blobID), SizeBytes: mailboxEntDraftReadInt64PointerCopy(sizeBytes), BindingDigest: bindingDigest, Attempt: attempt, RetryAt: mailboxEntDraftReadTimePointerCopy(retryAt), ExhaustedAt: mailboxEntDraftReadTimePointerCopy(exhaustedAt)}
	updated.UpdatedAt = databaseNow.UTC()
	return updated
}

func mailboxEntDraftCalendarCreateUnknownPostStateValid(record MailboxDraftRecord, request mailboxDraftCalendarCreateMarkUnknownRequest, phase MailboxDraftCreateMIMEPhase) bool {
	state := record.PendingCreateMIME
	if state == nil {
		return false
	}
	if phase == MailboxDraftCreateMIMEPhaseImport {
		return mailboxEntDraftCalendarCreateMatches(record, MailboxDraftStatusUnknown, phase, request.MutationID, request.ExpectedAttempt, request.ExpectedBindingDigest, request.BlobID, request.SizeBytes) && state.RetryAt == nil && state.ExhaustedAt == nil
	}
	return mailboxEntDraftCalendarCreateMatches(record, MailboxDraftStatusUnknown, phase, request.MutationID, request.ExpectedAttempt, request.ExpectedBindingDigest, "", 0) && (request.ExpectedAttempt < 3 && state.RetryAt != nil && state.ExhaustedAt == nil || request.ExpectedAttempt == 3 && state.RetryAt == nil && state.ExhaustedAt != nil)
}

func mailboxEntDraftCalendarCreateRestartableUnknown(record MailboxDraftRecord, mutationID uuid.UUID, attempt int, bindingDigest [sha256.Size]byte, databaseNow time.Time) bool {
	state := record.PendingCreateMIME
	return mailboxEntDraftCalendarCreateMatches(record, MailboxDraftStatusUnknown, MailboxDraftCreateMIMEPhaseUpload, mutationID, attempt, bindingDigest, "", 0) && state != nil && state.RetryAt != nil && !databaseNow.Before(*state.RetryAt) && state.ExhaustedAt == nil
}

func mailboxEntDraftCalendarCreateAbandonableUnknown(record MailboxDraftRecord, mutationID uuid.UUID, attempt int, bindingDigest [sha256.Size]byte) bool {
	state := record.PendingCreateMIME
	if !mailboxEntDraftCalendarCreateMatches(record, MailboxDraftStatusUnknown, MailboxDraftCreateMIMEPhaseUpload, mutationID, attempt, bindingDigest, "", 0) || state == nil {
		return false
	}
	return attempt < 3 && state.RetryAt != nil && state.ExhaustedAt == nil || attempt == 3 && state.RetryAt == nil && state.ExhaustedAt != nil
}

func mailboxEntDraftCalendarCreateHashesMatch(draft *ent.MailDraft, record MailboxDraftRecord, idempotencyKeyHash, requestHash [sha256.Size]byte) bool {
	return draft != nil && bytes.Equal(draft.CreateIdempotencyKeyHash, idempotencyKeyHash[:]) && record.CreateRequestHash == requestHash
}

func mailboxEntDraftCalendarCreateTerminalRecordValid(record MailboxDraftRecord, mutationID uuid.UUID) bool {
	return record.Status == MailboxDraftStatusActive && record.Revision == 1 && record.EngineMessageID != nil && record.EngineVersion != nil && record.ConfirmedMutationID != nil && *record.ConfirmedMutationID == mutationID && record.LastMutationID == nil && record.PendingMutationKind == nil && record.PendingCreateMIME == nil && record.PendingCreateAttachmentPhase == nil && record.PendingCreateUploadBlobID == nil && record.PendingCreateUploadSizeBytes == nil && record.PendingCreateUploadAttempt == 0 && record.PendingCreateUploadRetryAt == nil && record.PendingCreateUploadExhaustedAt == nil && len(record.Confirmed.AttachmentIDs) == 0 && record.AbandonedAt == nil && record.AbandonedMutationID == nil
}

func mailboxEntDraftCalendarCreateAbandonedRecordValid(record MailboxDraftRecord, mutationID uuid.UUID) bool {
	return record.Status == MailboxDraftStatusAbandoned && record.Revision == 1 && record.EngineMessageID == nil && record.EngineVersion == nil && record.LastMutationID == nil && record.ConfirmedMutationID == nil && record.PendingMutationKind == nil && record.PendingCreateMIME == nil && record.PendingCreateAttachmentPhase == nil && record.PendingCreateUploadBlobID == nil && record.PendingCreateUploadSizeBytes == nil && record.PendingCreateUploadAttempt == 0 && record.PendingCreateUploadRetryAt == nil && record.PendingCreateUploadExhaustedAt == nil && len(record.Confirmed.AttachmentIDs) == 0 && record.AbandonedAt != nil && mailboxEntValidTime(*record.AbandonedAt) && record.AbandonedMutationID != nil && *record.AbandonedMutationID == mutationID
}
