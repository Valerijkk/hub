package mail

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/maildraft"
	"github.com/valerii/hub/backend/ent/mailboxpolicy"
	"github.com/valerii/hub/backend/ent/mailsubmission"
)

// ReserveMailboxSubmission atomically replays or creates one queued send
// intent. It does not acquire an engine lease and makes no provider call.
func (repository *MailboxEntRepository) ReserveMailboxSubmission(ctx context.Context, request MailboxSubmissionCreateRequest) (MailboxSubmissionReservation, error) {
	if repository == nil || repository.client == nil || ctx == nil {
		return MailboxSubmissionReservation{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
	}
	if kind := mailboxEntSubmissionRequestKind(request); kind != "" {
		return MailboxSubmissionReservation{}, NewMailboxSubmissionRepositoryError(kind)
	}

	var result MailboxSubmissionReservation
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, kind := mailboxEntSubmissionLockAccount(ctx, transaction, request.AccountID)
		if kind != "" {
			return NewMailboxSubmissionRepositoryError(kind)
		}

		existing, found, kind := mailboxEntSubmissionLockIdempotencyReceipt(ctx, transaction, request)
		if kind != "" {
			return NewMailboxSubmissionRepositoryError(kind)
		}
		if found {
			result = MailboxSubmissionReservation{Kind: MailboxSubmissionReservationReplay, Submission: existing}
			return nil
		}

		draft, kind := mailboxEntSubmissionLockDraft(ctx, transaction, request.AccountID, request.DraftID)
		if kind != "" {
			return NewMailboxSubmissionRepositoryError(kind)
		}
		active, kind := mailboxEntSubmissionLockActiveByDraft(ctx, transaction, request.AccountID, request.DraftID)
		if kind != "" {
			return NewMailboxSubmissionRepositoryError(kind)
		}
		if len(active) != 0 {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorSubmissionInProgress)
		}
		policy, kind := mailboxEntSubmissionLockPolicy(ctx, transaction, account.OrganizationID)
		if kind != "" {
			return NewMailboxSubmissionRepositoryError(kind)
		}
		if kind := mailboxEntSubmissionNewAccountKind(account); kind != "" {
			return NewMailboxSubmissionRepositoryError(kind)
		}
		if kind := mailboxEntSubmissionDraftNewKind(draft, request); kind != "" {
			return NewMailboxSubmissionRepositoryError(kind)
		}

		databaseWallClock, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return mailboxEntSubmissionRepositoryError(err)
		}
		submissionKey := uuid.New().String()
		if !syncValidOpaque(submissionKey, MaxEngineOpaqueIDBytes) {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}
		binding, ok := newMailboxEngineSubmissionBinding(*account.EngineAccountID, draft.EngineDraftKey, *draft.EngineMessageID, *draft.EngineVersion, policy.MaxMessageBytes, submissionKey)
		if !ok {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}
		created, err := transaction.MailSubmission.Create().
			SetMailAccountID(request.AccountID).
			SetMailDraftID(request.DraftID).
			SetIdempotencyKeyHash(append([]byte(nil), request.IdempotencyKeyHash[:]...)).
			SetRequestHash(append([]byte(nil), request.RequestHash[:]...)).
			SetEngineSubmissionKey(submissionKey).
			SetEngineBindingVersion(binding.engineBindingVersion).
			SetEngineAccountID(binding.engineAccountID).
			SetEngineDraftKey(binding.engineDraftKey).
			SetEngineMessageID(binding.engineMessageID).
			SetEngineExpectedVersion(binding.engineExpectedVersion).
			SetEngineMaxMessageBytes(binding.engineMaxMessageBytes).
			SetEngineBindingHash(append([]byte(nil), binding.engineBindingHash[:]...)).
			SetStatus(mailsubmission.Status("queued")).
			SetNextAttemptAt(databaseWallClock).
			SetAttemptCount(0).
			SetReconcileAttempts(0).
			SetCreatedAt(databaseWallClock).
			SetUpdatedAt(databaseWallClock).
			Save(ctx)
		if err != nil {
			if ent.IsConstraintError(err) {
				return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorIdempotencyConflict)
			}
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
		}
		updated, err := draft.Update().
			SetStatus(maildraft.Status("submitting")).
			SetUpdatedAt(databaseWallClock).
			Save(ctx)
		if err != nil {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
		}
		if !mailboxEntSubmissionDraftTransitionValid(draft, updated, request, databaseWallClock) {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}
		record, ok := mailboxEntSubmissionRecord(created, request.AccountID)
		if created.AttemptCount != 0 || created.ReconcileAttempts != 0 || created.NextReconcileAt != nil || created.NextAttemptAt == nil || !created.NextAttemptAt.Equal(databaseWallClock) ||
			!created.CreatedAt.Equal(databaseWallClock) || !created.UpdatedAt.Equal(databaseWallClock) || !ok ||
			record.Status != MailboxSubmissionStatusQueued || record.DraftID == nil || *record.DraftID != request.DraftID ||
			record.ResultMailMessageID != nil || record.LastErrorCode != nil {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}
		result = MailboxSubmissionReservation{Kind: MailboxSubmissionReservationQueued, Submission: record}
		return nil
	})
	if err != nil {
		return MailboxSubmissionReservation{}, mailboxEntSubmissionRepositoryError(err)
	}
	return result, nil
}

func mailboxEntSubmissionRequestKind(request MailboxSubmissionCreateRequest) MailboxSubmissionRepositoryErrorKind {
	if request.AccountID == uuid.Nil || request.DraftID == uuid.Nil || request.ExpectedRevision < 1 ||
		!mailboxSubmissionHashPresent(request.IdempotencyKeyHash) || !mailboxSubmissionHashPresent(request.RequestHash) {
		return MailboxSubmissionRepositoryErrorInvalid
	}
	return ""
}

// mailboxEntSubmissionLockAccount is intentionally the first durable lock. It
// validates row shape but postpones active/deleting adjudication until after an
// exact idempotency replay has been found.
func mailboxEntSubmissionLockAccount(ctx context.Context, transaction *ent.Tx, accountID uuid.UUID) (*ent.MailAccount, MailboxSubmissionRepositoryErrorKind) {
	if ctx == nil || transaction == nil || accountID == uuid.Nil {
		return nil, MailboxSubmissionRepositoryErrorInvalid
	}
	account, err := transaction.MailAccount.Query().Where(mailaccount.IDEQ(accountID)).ForUpdate().Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, MailboxSubmissionRepositoryErrorNotFound
		}
		return nil, MailboxSubmissionRepositoryErrorUnavailable
	}
	if account == nil || account.ID != accountID {
		return nil, MailboxSubmissionRepositoryErrorInvalid
	}
	switch account.Status {
	case mailaccount.StatusActive, mailaccount.StatusDeleting, mailaccount.StatusProvisioning, mailaccount.StatusSuspended, mailaccount.StatusError:
	default:
		return nil, MailboxSubmissionRepositoryErrorInvalid
	}
	return account, ""
}

func mailboxEntSubmissionNewAccountKind(account *ent.MailAccount) MailboxSubmissionRepositoryErrorKind {
	if account == nil || account.ID == uuid.Nil {
		return MailboxSubmissionRepositoryErrorInvalid
	}
	switch account.Status {
	case mailaccount.StatusDeleting:
		return MailboxSubmissionRepositoryErrorAccountDeleting
	case mailaccount.StatusActive:
	case mailaccount.StatusProvisioning, mailaccount.StatusSuspended, mailaccount.StatusError:
		return MailboxSubmissionRepositoryErrorUnavailable
	default:
		return MailboxSubmissionRepositoryErrorInvalid
	}
	if !account.DesiredEnabled || account.EngineAccountID == nil || !syncValidOpaque(*account.EngineAccountID, MaxEngineOpaqueIDBytes) ||
		account.SyncEpoch < 0 || account.MutationEpoch < 1 || account.SyncLeaseEpoch < 1 || !mailboxEntValidTime(account.SyncScheduleUpdatedAt) {
		return MailboxSubmissionRepositoryErrorInvalid
	}
	if _, _, valid := mailboxEntSyncLeaseFence(account); !valid {
		return MailboxSubmissionRepositoryErrorInvalid
	}
	return ""
}

func mailboxEntSubmissionLockIdempotencyReceipt(ctx context.Context, transaction *ent.Tx, request MailboxSubmissionCreateRequest) (MailboxSubmissionRecord, bool, MailboxSubmissionRepositoryErrorKind) {
	if ctx == nil || transaction == nil || request.AccountID == uuid.Nil || !mailboxSubmissionHashPresent(request.IdempotencyKeyHash) || !mailboxSubmissionHashPresent(request.RequestHash) {
		return MailboxSubmissionRecord{}, false, MailboxSubmissionRepositoryErrorInvalid
	}
	row, err := transaction.MailSubmission.Query().
		Where(
			mailsubmission.MailAccountIDEQ(request.AccountID),
			mailsubmission.IdempotencyKeyHashEQ(request.IdempotencyKeyHash[:]),
		).
		ForUpdate().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return MailboxSubmissionRecord{}, false, ""
		}
		return MailboxSubmissionRecord{}, false, MailboxSubmissionRepositoryErrorUnavailable
	}
	record, ok := mailboxEntSubmissionRecord(row, request.AccountID)
	if !ok || !bytes.Equal(row.IdempotencyKeyHash, request.IdempotencyKeyHash[:]) {
		return MailboxSubmissionRecord{}, false, MailboxSubmissionRepositoryErrorInvalid
	}
	if !bytes.Equal(row.RequestHash, request.RequestHash[:]) {
		return MailboxSubmissionRecord{}, false, MailboxSubmissionRepositoryErrorIdempotencyConflict
	}
	return record, true, ""
}

func mailboxEntSubmissionLockDraft(ctx context.Context, transaction *ent.Tx, accountID, draftID uuid.UUID) (*ent.MailDraft, MailboxSubmissionRepositoryErrorKind) {
	if ctx == nil || transaction == nil || accountID == uuid.Nil || draftID == uuid.Nil {
		return nil, MailboxSubmissionRepositoryErrorInvalid
	}
	draft, err := transaction.MailDraft.Query().
		Where(maildraft.IDEQ(draftID), maildraft.MailAccountIDEQ(accountID)).
		ForUpdate().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, MailboxSubmissionRepositoryErrorNotFound
		}
		return nil, MailboxSubmissionRepositoryErrorUnavailable
	}
	if !mailboxEntDraftReadRowValid(draft, accountID, draftID) {
		return nil, MailboxSubmissionRepositoryErrorInvalid
	}
	return draft, ""
}

func mailboxEntSubmissionLockActiveByDraft(ctx context.Context, transaction *ent.Tx, accountID, draftID uuid.UUID) ([]MailboxSubmissionRecord, MailboxSubmissionRepositoryErrorKind) {
	if ctx == nil || transaction == nil || accountID == uuid.Nil || draftID == uuid.Nil {
		return nil, MailboxSubmissionRepositoryErrorInvalid
	}
	rows, err := transaction.MailSubmission.Query().
		Where(
			mailsubmission.MailAccountIDEQ(accountID),
			mailsubmission.MailDraftIDEQ(draftID),
			mailsubmission.StatusIn(
				mailsubmission.Status("queued"),
				mailsubmission.Status("authorizing"),
				mailsubmission.Status("submitting"),
				mailsubmission.Status("unknown"),
				mailsubmission.Status("accepted"),
			),
		).
		Order(ent.Asc(mailsubmission.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return nil, MailboxSubmissionRepositoryErrorUnavailable
	}
	records := make([]MailboxSubmissionRecord, 0, len(rows))
	for _, row := range rows {
		record, ok := mailboxEntSubmissionRecord(row, accountID)
		if !ok || record.DraftID == nil || *record.DraftID != draftID {
			return nil, MailboxSubmissionRepositoryErrorInvalid
		}
		switch record.Status {
		case MailboxSubmissionStatusQueued, MailboxSubmissionStatusAuthorizing, MailboxSubmissionStatusSubmitting, MailboxSubmissionStatusUnknown, MailboxSubmissionStatusAccepted:
		default:
			return nil, MailboxSubmissionRepositoryErrorInvalid
		}
		records = append(records, record)
	}
	return records, ""
}

// mailboxEntSubmissionLockPolicy locks the unique organization policy after
// the account, draft, and active submission locks. Its current limit becomes
// part of the immutable submission binding.
func mailboxEntSubmissionLockPolicy(ctx context.Context, transaction *ent.Tx, organizationID uuid.UUID) (*ent.MailboxPolicy, MailboxSubmissionRepositoryErrorKind) {
	if ctx == nil || transaction == nil || organizationID == uuid.Nil {
		return nil, MailboxSubmissionRepositoryErrorInvalid
	}
	policy, err := transaction.MailboxPolicy.Query().
		Where(mailboxpolicy.OrganizationIDEQ(organizationID)).
		ForUpdate().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, MailboxSubmissionRepositoryErrorNotFound
		}
		return nil, MailboxSubmissionRepositoryErrorUnavailable
	}
	if policy == nil || policy.ID == uuid.Nil || policy.OrganizationID != organizationID ||
		policy.MaxMessageBytes < 1 || policy.MaxMessageBytes > MaxEngineMessageBytes {
		return nil, MailboxSubmissionRepositoryErrorInvalid
	}
	return policy, ""
}

func mailboxEntSubmissionDraftNewKind(draft *ent.MailDraft, request MailboxSubmissionCreateRequest) MailboxSubmissionRepositoryErrorKind {
	if draft == nil || !mailboxEntDraftReadRowValid(draft, request.AccountID, request.DraftID) {
		return MailboxSubmissionRepositoryErrorInvalid
	}
	if MailboxDraftStatus(draft.Status) != MailboxDraftStatusActive {
		if MailboxDraftStatus(draft.Status) == MailboxDraftStatusSubmitting {
			return MailboxSubmissionRepositoryErrorSubmissionInProgress
		}
		return MailboxSubmissionRepositoryErrorDraftNotSendable
	}
	if draft.Revision != request.ExpectedRevision {
		return MailboxSubmissionRepositoryErrorVersionConflict
	}
	if draft.EngineMessageID == nil || draft.EngineVersion == nil || !syncValidOpaque(draft.EngineDraftKey, MaxEngineOpaqueIDBytes) ||
		!syncValidOpaque(*draft.EngineMessageID, MaxEngineOpaqueIDBytes) || !syncValidOpaque(*draft.EngineVersion, MaxEngineOpaqueIDBytes) {
		return MailboxSubmissionRepositoryErrorInvalid
	}
	return ""
}

func mailboxEntSubmissionDraftTransitionValid(before, after *ent.MailDraft, request MailboxSubmissionCreateRequest, updatedAt time.Time) bool {
	if before == nil || after == nil || !mailboxEntValidTime(updatedAt) || !mailboxEntDraftReadRowValid(before, request.AccountID, request.DraftID) ||
		!mailboxEntDraftReadRowValid(after, request.AccountID, request.DraftID) || before.ID != after.ID ||
		before.MailAccountID != after.MailAccountID || before.Revision != after.Revision || before.EngineDraftKey != after.EngineDraftKey ||
		before.EngineMessageID == nil || before.EngineVersion == nil || after.EngineMessageID == nil || after.EngineVersion == nil ||
		*before.EngineMessageID != *after.EngineMessageID || *before.EngineVersion != *after.EngineVersion ||
		MailboxDraftStatus(before.Status) != MailboxDraftStatusActive || MailboxDraftStatus(after.Status) != MailboxDraftStatusSubmitting ||
		!after.UpdatedAt.Equal(updatedAt) {
		return false
	}
	return before.DraftClientID == after.DraftClientID && before.FromIdentityID == after.FromIdentityID &&
		mailboxEntSubmissionUUIDPointersEqual(before.ReplyToMessageID, after.ReplyToMessageID) &&
		mailboxEntSubmissionUUIDPointersEqual(before.ForwardMessageID, after.ForwardMessageID) &&
		mailboxEntSubmissionUUIDPointersEqual(before.LastMutationID, after.LastMutationID) &&
		mailboxEntSubmissionUUIDPointersEqual(before.ConfirmedMutationID, after.ConfirmedMutationID)
}

func mailboxEntSubmissionUUIDPointersEqual(left, right *uuid.UUID) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func mailboxEntSubmissionRecord(row *ent.MailSubmission, accountID uuid.UUID) (MailboxSubmissionRecord, bool) {
	if !mailboxEntSubmissionRowValid(row, accountID) {
		return MailboxSubmissionRecord{}, false
	}
	var draftID *uuid.UUID
	if row.MailDraftID != nil {
		draftID = mailboxSubmissionUUIDPointerCopy(row.MailDraftID)
	}
	var resultMessageID *uuid.UUID
	if row.ResultMailMessageID != nil {
		resultMessageID = mailboxSubmissionUUIDPointerCopy(row.ResultMailMessageID)
	}
	var errorCode *MailboxSubmissionErrorCode
	if row.LastErrorCode != nil {
		value := MailboxSubmissionErrorCode(*row.LastErrorCode)
		errorCode = &value
	}
	return MailboxSubmissionRecord{
		ID:                  row.ID,
		AccountID:           row.MailAccountID,
		DraftID:             draftID,
		Status:              MailboxSubmissionStatus(row.Status),
		ResultMailMessageID: resultMessageID,
		LastErrorCode:       errorCode,
		CreatedAt:           row.CreatedAt.UTC(),
		UpdatedAt:           row.UpdatedAt.UTC(),
	}, true
}

func mailboxEntSubmissionRowValid(row *ent.MailSubmission, accountID uuid.UUID) bool {
	if row == nil || row.ID == uuid.Nil || row.MailAccountID != accountID || accountID == uuid.Nil ||
		len(row.IdempotencyKeyHash) != sha256.Size || len(row.RequestHash) != sha256.Size ||
		row.AttemptCount < 0 || row.AttemptCount > 8 || row.ReconcileAttempts < 0 || row.ReconcileAttempts > 10 || !mailboxEntValidTime(row.CreatedAt) || !mailboxEntValidTime(row.UpdatedAt) || row.UpdatedAt.Before(row.CreatedAt) ||
		!mailboxSubmissionUUIDPointerValid(row.MailDraftID) || !mailboxSubmissionUUIDPointerValid(row.ResultMailMessageID) ||
		row.CommandFence < 0 || !mailboxEntSubmissionCommandLeaseValid(row) ||
		!mailboxEntSubmissionDispatchAuthorizationValid(row) || !mailboxEntSubmissionProviderPairValid(row) ||
		(row.NextAttemptAt != nil && !mailboxEntValidTime(*row.NextAttemptAt)) ||
		(row.NextReconcileAt != nil && !mailboxEntValidTime(*row.NextReconcileAt)) {
		return false
	}
	if _, ok := mailboxEntSubmissionBindingFromRow(row); !ok {
		return false
	}
	var errorCode *MailboxSubmissionErrorCode
	if row.LastErrorCode != nil {
		value := MailboxSubmissionErrorCode(*row.LastErrorCode)
		if !validMailboxSubmissionErrorCode(value) {
			return false
		}
		errorCode = &value
	}
	switch MailboxSubmissionStatus(row.Status) {
	case MailboxSubmissionStatusQueued:
		return row.AttemptCount <= 7 && row.MailDraftID != nil && row.CommandLeaseToken == nil && row.DispatchAuthorizedAt == nil &&
			row.ProviderSubmissionID == nil && row.ResultMailMessageID == nil && errorCode == nil && row.NextAttemptAt != nil &&
			row.ReconcileAttempts == 0 && row.NextReconcileAt == nil
	case MailboxSubmissionStatusAuthorizing:
		return row.AttemptCount >= 1 && row.MailDraftID != nil && row.CommandLeaseToken != nil && row.DispatchAuthorizedAt == nil &&
			row.ProviderSubmissionID == nil && row.ResultMailMessageID == nil && errorCode == nil && row.NextAttemptAt == nil && row.CommandFence >= 1 &&
			row.ReconcileAttempts == 0 && row.NextReconcileAt == nil
	case MailboxSubmissionStatusSubmitting:
		return row.AttemptCount >= 1 && row.MailDraftID != nil && row.CommandLeaseToken != nil && row.DispatchAuthorizedAt != nil &&
			row.ProviderSubmissionID == nil && row.ResultMailMessageID == nil && errorCode == nil && row.NextAttemptAt == nil && row.CommandFence >= 1 &&
			row.ReconcileAttempts == 0 && row.NextReconcileAt == nil
	case MailboxSubmissionStatusAccepted:
		return row.AttemptCount >= 1 && row.MailDraftID != nil && row.DispatchAuthorizedAt != nil && row.ProviderSubmissionID != nil &&
			row.ResultMailMessageID == nil && errorCode == nil && row.NextAttemptAt == nil && row.CommandFence >= 1 &&
			row.ReconcileAttempts >= 0 && row.ReconcileAttempts <= 10 && row.NextReconcileAt == nil
	case MailboxSubmissionStatusSent:
		return row.AttemptCount >= 1 && row.CommandLeaseToken == nil && row.DispatchAuthorizedAt != nil && row.ProviderSubmissionID != nil &&
			row.ResultMailMessageID != nil && errorCode == nil && row.NextAttemptAt == nil && row.CommandFence >= 1 &&
			row.ReconcileAttempts >= 0 && row.ReconcileAttempts <= 10 && row.NextReconcileAt == nil
	case MailboxSubmissionStatusUnknown:
		if row.AttemptCount < 1 || row.MailDraftID == nil || row.DispatchAuthorizedAt == nil || row.ProviderSubmissionID != nil ||
			row.ResultMailMessageID != nil || errorCode == nil || row.NextAttemptAt != nil || row.CommandFence < 1 {
			return false
		}
		if row.CommandLeaseToken == nil {
			if row.ReconcileAttempts >= 0 && row.ReconcileAttempts <= 9 && row.NextReconcileAt != nil {
				return mailboxEntSubmissionReconciliationPendingCode(*errorCode)
			}
			return row.ReconcileAttempts == 10 && row.NextReconcileAt == nil && *errorCode == MailboxSubmissionErrorCodeReconciliationExhausted
		}
		return row.ReconcileAttempts >= 1 && row.ReconcileAttempts <= 10 && row.NextReconcileAt == nil &&
			mailboxEntSubmissionReconciliationPendingCode(*errorCode)
	case MailboxSubmissionStatusFailed:
		if row.AttemptCount < 1 || row.CommandLeaseToken != nil || row.DispatchAuthorizedAt != nil || row.ProviderSubmissionID != nil ||
			row.ResultMailMessageID != nil || errorCode == nil || row.NextAttemptAt != nil || row.ReconcileAttempts != 0 || row.NextReconcileAt != nil {
			return false
		}
		switch *errorCode {
		case MailboxSubmissionErrorCodeRejected,
			MailboxSubmissionErrorCodePolicy,
			MailboxSubmissionErrorCodeAttachmentNotClean,
			MailboxSubmissionErrorCodeIdentityUnavailable,
			MailboxSubmissionErrorCodeRecipientLimit,
			MailboxSubmissionErrorCodeSizeLimit:
			return true
		case MailboxSubmissionErrorCodeRetryExhausted:
			return row.AttemptCount == 8
		default:
			return false
		}
	default:
		return false
	}
}

func mailboxEntSubmissionReconciliationPendingCode(code MailboxSubmissionErrorCode) bool {
	switch code {
	case MailboxSubmissionErrorCodeOutcomeUnknown,
		MailboxSubmissionErrorCodeReconciliationRequired,
		MailboxSubmissionErrorCodeMailboxUnavailable:
		return true
	default:
		return false
	}
}

func mailboxEntSubmissionBindingFromRow(row *ent.MailSubmission) (mailboxEngineSubmissionBinding, bool) {
	if row == nil || len(row.EngineBindingHash) != sha256.Size {
		return mailboxEngineSubmissionBinding{}, false
	}
	binding := mailboxEngineSubmissionBinding{
		engineBindingVersion:  row.EngineBindingVersion,
		engineAccountID:       row.EngineAccountID,
		engineDraftKey:        row.EngineDraftKey,
		engineMessageID:       row.EngineMessageID,
		engineExpectedVersion: row.EngineExpectedVersion,
		engineMaxMessageBytes: row.EngineMaxMessageBytes,
		engineSubmissionKey:   row.EngineSubmissionKey,
	}
	copy(binding.engineBindingHash[:], row.EngineBindingHash)
	return binding, mailboxEngineSubmissionBindingValid(binding)
}

func mailboxEntSubmissionCommandLeaseValid(row *ent.MailSubmission) bool {
	if row == nil || (row.CommandLeaseToken == nil) != (row.CommandLeaseExpiresAt == nil) {
		return false
	}
	return row.CommandLeaseToken == nil || (*row.CommandLeaseToken != uuid.Nil && mailboxEntValidTime(*row.CommandLeaseExpiresAt))
}

func mailboxEntSubmissionDispatchAuthorizationValid(row *ent.MailSubmission) bool {
	if row == nil || (row.DispatchAuthorizedAt == nil) != (row.DispatchAuthorizationHash == nil) {
		return false
	}
	return row.DispatchAuthorizedAt == nil || (mailboxEntValidTime(*row.DispatchAuthorizedAt) && len(*row.DispatchAuthorizationHash) == sha256.Size)
}

func mailboxEntSubmissionProviderPairValid(row *ent.MailSubmission) bool {
	if row == nil || (row.ProviderSubmissionID == nil) != (row.ProviderMessageID == nil) {
		return false
	}
	return row.ProviderSubmissionID == nil ||
		(syncValidOpaque(*row.ProviderSubmissionID, MaxEngineOpaqueIDBytes) && syncValidOpaque(*row.ProviderMessageID, MaxEngineOpaqueIDBytes))
}

func mailboxEntSubmissionRepositoryError(err error) error {
	var repositoryError *MailboxSubmissionRepositoryError
	if errors.As(err, &repositoryError) && repositoryError != nil {
		return NewMailboxSubmissionRepositoryError(repositoryError.Kind())
	}
	switch {
	case errors.Is(err, ErrMailboxServiceInvalid):
		return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
	case errors.Is(err, ErrMailboxServiceNotFound):
		return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorNotFound)
	case errors.Is(err, ErrMailboxServiceDeleting):
		return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorAccountDeleting)
	default:
		return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
	}
}
