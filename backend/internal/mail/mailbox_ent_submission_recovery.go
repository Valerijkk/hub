package mail

import (
	"context"
	"crypto/sha256"
	"errors"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/maildraft"
	"github.com/valerii/hub/backend/ent/mailsubmission"
	"github.com/valerii/hub/backend/ent/predicate"
)

var _ MailboxSubmissionStaleRecoveryRepository = (*MailboxEntRepository)(nil)

// ScanDueMailboxSubmissionRecovery selects a bounded, ordered advisory list
// of expired authorizing/submitting/unknown/accepted command leases. It never
// makes a provider call and its output is not a recovery capability.
func (repository *MailboxEntRepository) ScanDueMailboxSubmissionRecovery(ctx context.Context, request MailboxSubmissionRecoveryScanRequest) (MailboxSubmissionRecoveryScanResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxEntSubmissionRecoveryScanRequestValid(request) {
		return MailboxSubmissionRecoveryScanResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
	}

	var result MailboxSubmissionRecoveryScanResult
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		selectedAt, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
		}
		rows, err := transaction.MailSubmission.Query().
			Where(
				mailboxEntSubmissionRecoverySkipLocked(),
				mailsubmission.StatusIn(mailsubmission.Status("authorizing"), mailsubmission.Status("submitting"), mailsubmission.Status("unknown"), mailsubmission.Status("accepted")),
				mailsubmission.CommandLeaseExpiresAtNotNil(),
				mailsubmission.CommandLeaseExpiresAtLTE(selectedAt),
			).
			Order(ent.Asc(mailsubmission.FieldCommandLeaseExpiresAt), ent.Asc(mailsubmission.FieldID)).
			Limit(request.Limit).
			All(ctx)
		if err != nil {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
		}

		submissions := make([]MailboxSubmissionRecoveryDue, 0, len(rows))
		for _, submission := range rows {
			due, ok := mailboxEntSubmissionRecoveryDueFromRow(submission)
			if !ok {
				return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
			}
			submissions = append(submissions, due)
		}
		scanned := MailboxSubmissionRecoveryScanResult{SelectedAt: selectedAt, Submissions: submissions}
		if !mailboxEntSubmissionRecoveryScanResultValid(request, scanned) {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}
		result = mailboxEntSubmissionRecoveryScanResultCopy(scanned)
		return nil
	})
	if err != nil {
		return MailboxSubmissionRecoveryScanResult{}, mailboxEntSubmissionRecoveryRepositoryError(err)
	}
	if !mailboxEntSubmissionRecoveryScanResultValid(request, result) {
		return MailboxSubmissionRecoveryScanResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
	}
	return mailboxEntSubmissionRecoveryScanResultCopy(result), nil
}

// RecoverStaleMailboxSubmission re-reads one advisory submission under the
// Account -> Submission -> Draft lock order. It only resolves already-expired
// dual leases and never calls a provider, queue, or transport.
func (repository *MailboxEntRepository) RecoverStaleMailboxSubmission(ctx context.Context, request MailboxSubmissionStaleRecoveryRequest) (MailboxSubmissionStaleRecoveryResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxSubmissionStaleRecoveryRequestValid(request) {
		return MailboxSubmissionStaleRecoveryResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
	}

	lookup, err := repository.client.MailSubmission.Query().Where(mailsubmission.IDEQ(request.SubmissionID)).Only(ctx)
	if ent.IsNotFound(err) {
		return mailboxSubmissionStaleRecoveryNoop(), nil
	}
	if err != nil {
		return MailboxSubmissionStaleRecoveryResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
	}
	if lookup == nil || lookup.ID != request.SubmissionID || lookup.MailAccountID == uuid.Nil {
		return MailboxSubmissionStaleRecoveryResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
	}

	result := mailboxSubmissionStaleRecoveryNoop()
	err = repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, err := transaction.MailAccount.Query().
			Where(mailaccount.IDEQ(lookup.MailAccountID)).
			ForUpdate().
			Only(ctx)
		if ent.IsNotFound(err) {
			return nil
		}
		if err != nil {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
		}
		if kind := mailboxEntSubmissionClaimAccountKind(account); kind != "" {
			return NewMailboxSubmissionRepositoryError(kind)
		}

		submission, err := transaction.MailSubmission.Query().
			Where(mailsubmission.IDEQ(request.SubmissionID), mailsubmission.MailAccountIDEQ(account.ID)).
			ForUpdate().
			Only(ctx)
		if ent.IsNotFound(err) {
			return nil
		}
		if err != nil {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
		}
		if !mailboxEntSubmissionRecoveryStaleStatus(MailboxSubmissionStatus(submission.Status)) {
			return nil
		}
		if !mailboxEntSubmissionRowValid(submission, account.ID) {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}
		if submission.MailDraftID == nil || *submission.MailDraftID == uuid.Nil {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}

		draft, err := transaction.MailDraft.Query().
			Where(maildraft.IDEQ(*submission.MailDraftID), maildraft.MailAccountIDEQ(account.ID)).
			ForUpdate().
			Only(ctx)
		if err != nil {
			return mailboxEntSubmissionRecoveryDraftError(err)
		}
		if !mailboxEntDraftReadRowValid(draft, account.ID, *submission.MailDraftID) {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}

		databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
		}
		switch mailboxEntSubmissionRecoveryClassifyLeaseState(account, submission) {
		case mailboxEntSubmissionRecoveryLeaseStateReleased:
			if !mailboxEntSubmissionRecoveryReleasedStateValid(submission) {
				return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
			}
			return nil
		case mailboxEntSubmissionRecoveryLeaseStateHeld:
			if !mailboxEntSubmissionRecoveryCurrentStateValid(submission, account.ID) {
				return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
			}
		case mailboxEntSubmissionRecoveryLeaseStateInvalid:
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		case mailboxEntSubmissionRecoveryLeaseStateConflict:
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorVersionConflict)
		default:
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}
		if !mailboxEntSubmissionRecoveryDualFenceShapeMatches(account, submission, draft) {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorVersionConflict)
		}
		if account.MutationLeaseUntil.After(databaseNow) {
			return nil
		}
		if !mailboxEntSubmissionRecoveryDualFenceMatches(account, submission, draft, databaseNow) {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorVersionConflict)
		}

		switch MailboxSubmissionStatus(submission.Status) {
		case MailboxSubmissionStatusAuthorizing:
			if submission.AttemptCount >= 1 && submission.AttemptCount <= 7 {
				updated, err := mailboxEntSubmissionRecoveryRequeue(ctx, account, submission, draft, databaseNow)
				if err != nil {
					return err
				}
				result = updated
				return nil
			}
			if submission.AttemptCount != 8 {
				return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
			}
			updated, err := mailboxEntSubmissionRecoveryTerminal(ctx, account, submission, draft, databaseNow)
			if err != nil {
				return err
			}
			result = updated
			return nil
		case MailboxSubmissionStatusSubmitting:
			updated, err := mailboxEntSubmissionRecoveryUnknown(ctx, account, submission, draft, databaseNow)
			if err != nil {
				return err
			}
			result = updated
			return nil
		case MailboxSubmissionStatusUnknown:
			updated, err := mailboxEntSubmissionRecoveryReleaseUnknown(ctx, account, submission, draft, databaseNow)
			if err != nil {
				return err
			}
			result = updated
			return nil
		case MailboxSubmissionStatusAccepted:
			updated, err := mailboxEntSubmissionRecoveryReleaseAccepted(ctx, account, submission, draft, databaseNow)
			if err != nil {
				return err
			}
			result = updated
			return nil
		default:
			return nil
		}
	})
	if err != nil {
		return MailboxSubmissionStaleRecoveryResult{}, mailboxEntSubmissionRecoveryRepositoryError(err)
	}
	if !mailboxSubmissionStaleRecoveryResultValid(result) {
		return MailboxSubmissionStaleRecoveryResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
	}
	return mailboxSubmissionStaleRecoveryResultCopy(result), nil
}

func mailboxEntSubmissionRecoveryScanRequestValid(request MailboxSubmissionRecoveryScanRequest) bool {
	return mailboxSubmissionRecoveryScanRequestValid(request)
}

func mailboxEntSubmissionRecoveryScanResultValid(request MailboxSubmissionRecoveryScanRequest, result MailboxSubmissionRecoveryScanResult) bool {
	return mailboxSubmissionRecoveryScanResultValid(request, result)
}

func mailboxEntSubmissionRecoveryScanResultCopy(result MailboxSubmissionRecoveryScanResult) MailboxSubmissionRecoveryScanResult {
	return mailboxSubmissionRecoveryScanResultCopy(result)
}

func mailboxEntSubmissionRecoverySkipLocked() predicate.MailSubmission {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.SkipLocked))
	}
}

func mailboxEntSubmissionRecoveryDueFromRow(submission *ent.MailSubmission) (MailboxSubmissionRecoveryDue, bool) {
	if submission == nil || submission.ID == uuid.Nil || submission.CommandLeaseExpiresAt == nil {
		return MailboxSubmissionRecoveryDue{}, false
	}
	due := MailboxSubmissionRecoveryDue{
		SubmissionID: submission.ID,
		Status:       MailboxSubmissionStatus(submission.Status),
		LeaseUntil:   *submission.CommandLeaseExpiresAt,
	}
	return due, mailboxSubmissionRecoveryDueValid(due)
}

func mailboxEntSubmissionRecoveryStaleStatus(status MailboxSubmissionStatus) bool {
	switch status {
	case MailboxSubmissionStatusAuthorizing,
		MailboxSubmissionStatusSubmitting,
		MailboxSubmissionStatusUnknown,
		MailboxSubmissionStatusAccepted:
		return true
	default:
		return false
	}
}

type mailboxEntSubmissionRecoveryLeaseState uint8

const (
	mailboxEntSubmissionRecoveryLeaseStateInvalid mailboxEntSubmissionRecoveryLeaseState = iota
	mailboxEntSubmissionRecoveryLeaseStateReleased
	mailboxEntSubmissionRecoveryLeaseStateHeld
	mailboxEntSubmissionRecoveryLeaseStateConflict
)

func mailboxEntSubmissionRecoveryClassifyLeaseState(account *ent.MailAccount, submission *ent.MailSubmission) mailboxEntSubmissionRecoveryLeaseState {
	if account == nil || submission == nil ||
		(account.MutationLeaseToken == nil) != (account.MutationLeaseUntil == nil) ||
		(submission.CommandLeaseToken == nil) != (submission.CommandLeaseExpiresAt == nil) {
		return mailboxEntSubmissionRecoveryLeaseStateInvalid
	}
	accountHeld := account.MutationLeaseToken != nil
	submissionHeld := submission.CommandLeaseToken != nil
	switch {
	case !accountHeld && !submissionHeld:
		return mailboxEntSubmissionRecoveryLeaseStateReleased
	case accountHeld && submissionHeld:
		return mailboxEntSubmissionRecoveryLeaseStateHeld
	default:
		return mailboxEntSubmissionRecoveryLeaseStateConflict
	}
}

func mailboxEntSubmissionRecoveryReleasedStateValid(submission *ent.MailSubmission) bool {
	if submission == nil || submission.CommandLeaseToken != nil || submission.CommandLeaseExpiresAt != nil {
		return false
	}
	switch MailboxSubmissionStatus(submission.Status) {
	case MailboxSubmissionStatusUnknown:
		return (submission.ReconcileAttempts >= 0 && submission.ReconcileAttempts <= 9 && submission.NextReconcileAt != nil) ||
			(submission.ReconcileAttempts == 10 && submission.NextReconcileAt == nil)
	case MailboxSubmissionStatusAccepted:
		return submission.NextReconcileAt == nil
	default:
		return false
	}
}

func mailboxEntSubmissionRecoveryCurrentStateValid(submission *ent.MailSubmission, accountID uuid.UUID) bool {
	if submission == nil || accountID == uuid.Nil {
		return false
	}
	switch MailboxSubmissionStatus(submission.Status) {
	case MailboxSubmissionStatusAuthorizing, MailboxSubmissionStatusSubmitting:
		return mailboxEntSubmissionRowValid(submission, accountID) && submission.ReconcileAttempts == 0 && submission.NextReconcileAt == nil
	case MailboxSubmissionStatusUnknown:
		return mailboxEntSubmissionRecoveryUnknownLeaseStateValid(submission, accountID)
	case MailboxSubmissionStatusAccepted:
		return mailboxEntSubmissionRecoveryAcceptedLeaseStateValid(submission, accountID)
	default:
		return false
	}
}

func mailboxEntSubmissionRecoveryLeaseStateBaseValid(submission *ent.MailSubmission, accountID uuid.UUID) bool {
	if submission == nil || submission.ID == uuid.Nil || submission.MailAccountID != accountID || accountID == uuid.Nil ||
		len(submission.IdempotencyKeyHash) != sha256.Size || len(submission.RequestHash) != sha256.Size ||
		submission.AttemptCount < 1 || submission.AttemptCount > 8 || !mailboxEntValidTime(submission.CreatedAt) || !mailboxEntValidTime(submission.UpdatedAt) || submission.UpdatedAt.Before(submission.CreatedAt) ||
		!mailboxSubmissionUUIDPointerValid(submission.MailDraftID) || !mailboxSubmissionUUIDPointerValid(submission.ResultMailMessageID) ||
		submission.CommandFence < 1 || !mailboxEntSubmissionCommandLeaseValid(submission) ||
		!mailboxEntSubmissionDispatchAuthorizationValid(submission) || !mailboxEntSubmissionProviderPairValid(submission) ||
		(submission.NextAttemptAt != nil && !mailboxEntValidTime(*submission.NextAttemptAt)) ||
		(submission.NextReconcileAt != nil && !mailboxEntValidTime(*submission.NextReconcileAt)) ||
		submission.ReconcileAttempts < 0 || submission.ReconcileAttempts > 10 {
		return false
	}
	_, ok := mailboxEntSubmissionBindingFromRow(submission)
	return ok
}

func mailboxEntSubmissionRecoveryUnknownLeaseStateValid(submission *ent.MailSubmission, accountID uuid.UUID) bool {
	return mailboxEntSubmissionRecoveryLeaseStateBaseValid(submission, accountID) &&
		MailboxSubmissionStatus(submission.Status) == MailboxSubmissionStatusUnknown && submission.MailDraftID != nil &&
		submission.CommandLeaseToken != nil && submission.DispatchAuthorizedAt != nil && submission.DispatchAuthorizationHash != nil &&
		submission.ProviderSubmissionID == nil && submission.ProviderMessageID == nil && submission.ResultMailMessageID == nil &&
		submission.NextAttemptAt == nil && submission.ReconcileAttempts >= 1 && submission.ReconcileAttempts <= 10 && submission.NextReconcileAt == nil
}

func mailboxEntSubmissionRecoveryAcceptedLeaseStateValid(submission *ent.MailSubmission, accountID uuid.UUID) bool {
	return mailboxEntSubmissionRecoveryLeaseStateBaseValid(submission, accountID) &&
		MailboxSubmissionStatus(submission.Status) == MailboxSubmissionStatusAccepted && submission.MailDraftID != nil &&
		submission.CommandLeaseToken != nil && submission.DispatchAuthorizedAt != nil && submission.DispatchAuthorizationHash != nil &&
		submission.ProviderSubmissionID != nil && submission.ProviderMessageID != nil && submission.ResultMailMessageID == nil &&
		submission.LastErrorCode == nil && submission.NextAttemptAt == nil && submission.NextReconcileAt == nil
}

func mailboxEntSubmissionRecoveryDualFenceShapeMatches(account *ent.MailAccount, submission *ent.MailSubmission, draft *ent.MailDraft) bool {
	if account == nil || submission == nil || draft == nil || account.MutationLeaseToken == nil || account.MutationLeaseUntil == nil ||
		submission.CommandLeaseToken == nil || submission.CommandLeaseExpiresAt == nil || *account.MutationLeaseToken == uuid.Nil ||
		*submission.CommandLeaseToken == uuid.Nil || *account.MutationLeaseToken != *submission.CommandLeaseToken ||
		!account.MutationLeaseUntil.Equal(*submission.CommandLeaseExpiresAt) || account.MutationEpoch < 1 ||
		submission.CommandFence != account.MutationEpoch || submission.MailDraftID == nil || *submission.MailDraftID == uuid.Nil ||
		!mailboxEntValidTime(*account.MutationLeaseUntil) ||
		!mailboxEntValidTime(*submission.CommandLeaseExpiresAt) {
		return false
	}
	binding, ok := mailboxEntSubmissionBindingFromRow(submission)
	return ok && mailboxEntSubmissionClaimDraftMatches(draft, account.ID, *submission.MailDraftID, binding)
}

func mailboxEntSubmissionRecoveryDualFenceMatches(account *ent.MailAccount, submission *ent.MailSubmission, draft *ent.MailDraft, databaseNow time.Time) bool {
	return mailboxEntSubmissionRecoveryDualFenceShapeMatches(account, submission, draft) && !account.MutationLeaseUntil.After(databaseNow) &&
		!submission.CommandLeaseExpiresAt.After(databaseNow)
}

func mailboxEntSubmissionRecoveryRequeue(ctx context.Context, account *ent.MailAccount, submission *ent.MailSubmission, draft *ent.MailDraft, databaseNow time.Time) (MailboxSubmissionStaleRecoveryResult, error) {
	if submission.AttemptCount < 1 || submission.AttemptCount > 7 || !mailboxEntSubmissionRecoveryDualFenceMatches(account, submission, draft, databaseNow) {
		return MailboxSubmissionStaleRecoveryResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorVersionConflict)
	}
	nextEpoch, ok := mailboxEntNextEpoch(account.MutationEpoch)
	if !ok {
		return MailboxSubmissionStaleRecoveryResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
	}
	if err := submission.Update().
		SetStatus(mailsubmission.Status("queued")).
		ClearCommandLeaseToken().
		ClearCommandLeaseExpiresAt().
		ClearDispatchAuthorizedAt().
		ClearDispatchAuthorizationHash().
		ClearProviderSubmissionID().
		ClearProviderMessageID().
		ClearResultMailMessageID().
		ClearLastErrorCode().
		SetNextAttemptAt(databaseNow).
		SetAttemptCount(submission.AttemptCount).
		SetUpdatedAt(databaseNow).
		Exec(ctx); err != nil {
		return MailboxSubmissionStaleRecoveryResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
	}
	if err := account.Update().
		ClearMutationLeaseToken().
		ClearMutationLeaseUntil().
		SetMutationEpoch(nextEpoch).
		SetSyncScheduleUpdatedAt(databaseNow).
		Exec(ctx); err != nil {
		return MailboxSubmissionStaleRecoveryResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
	}
	return mailboxSubmissionStaleRecoveryRequeued(databaseNow), nil
}

func mailboxEntSubmissionRecoveryTerminal(ctx context.Context, account *ent.MailAccount, submission *ent.MailSubmission, draft *ent.MailDraft, databaseNow time.Time) (MailboxSubmissionStaleRecoveryResult, error) {
	if submission.AttemptCount != 8 || !mailboxEntSubmissionRecoveryDualFenceMatches(account, submission, draft, databaseNow) {
		return MailboxSubmissionStaleRecoveryResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorVersionConflict)
	}
	nextEpoch, ok := mailboxEntNextEpoch(account.MutationEpoch)
	if !ok {
		return MailboxSubmissionStaleRecoveryResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
	}
	if err := submission.Update().
		SetStatus(mailsubmission.Status("failed")).
		ClearCommandLeaseToken().
		ClearCommandLeaseExpiresAt().
		ClearDispatchAuthorizedAt().
		ClearDispatchAuthorizationHash().
		ClearProviderSubmissionID().
		ClearProviderMessageID().
		ClearResultMailMessageID().
		SetLastErrorCode(mailsubmission.LastErrorCode(MailboxSubmissionErrorCodeRetryExhausted)).
		ClearNextAttemptAt().
		SetUpdatedAt(databaseNow).
		Exec(ctx); err != nil {
		return MailboxSubmissionStaleRecoveryResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
	}
	if err := draft.Update().
		SetStatus(maildraft.Status("active")).
		SetUpdatedAt(databaseNow).
		Exec(ctx); err != nil {
		return MailboxSubmissionStaleRecoveryResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
	}
	if err := account.Update().
		ClearMutationLeaseToken().
		ClearMutationLeaseUntil().
		SetMutationEpoch(nextEpoch).
		SetSyncScheduleUpdatedAt(databaseNow).
		Exec(ctx); err != nil {
		return MailboxSubmissionStaleRecoveryResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
	}
	return mailboxSubmissionStaleRecoveryFailed(), nil
}

func mailboxEntSubmissionRecoveryUnknown(ctx context.Context, account *ent.MailAccount, submission *ent.MailSubmission, draft *ent.MailDraft, databaseNow time.Time) (MailboxSubmissionStaleRecoveryResult, error) {
	if submission.AttemptCount < 1 || submission.AttemptCount > 8 || !mailboxEntSubmissionRecoveryDualFenceMatches(account, submission, draft, databaseNow) {
		return MailboxSubmissionStaleRecoveryResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorVersionConflict)
	}
	nextEpoch, ok := mailboxEntNextEpoch(account.MutationEpoch)
	if !ok {
		return MailboxSubmissionStaleRecoveryResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
	}
	if err := submission.Update().
		SetStatus(mailsubmission.Status("unknown")).
		ClearCommandLeaseToken().
		ClearCommandLeaseExpiresAt().
		ClearProviderSubmissionID().
		ClearProviderMessageID().
		ClearResultMailMessageID().
		SetLastErrorCode(mailsubmission.LastErrorCode(MailboxSubmissionErrorCodeReconciliationRequired)).
		ClearNextAttemptAt().
		SetReconcileAttempts(0).
		SetNextReconcileAt(databaseNow).
		SetUpdatedAt(databaseNow).
		Exec(ctx); err != nil {
		return MailboxSubmissionStaleRecoveryResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
	}
	if err := account.Update().
		ClearMutationLeaseToken().
		ClearMutationLeaseUntil().
		SetMutationEpoch(nextEpoch).
		SetSyncScheduleUpdatedAt(databaseNow).
		Exec(ctx); err != nil {
		return MailboxSubmissionStaleRecoveryResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
	}
	return mailboxSubmissionStaleRecoveryReconciliationRequired(), nil
}

func mailboxEntSubmissionRecoveryReleaseUnknown(ctx context.Context, account *ent.MailAccount, submission *ent.MailSubmission, draft *ent.MailDraft, databaseNow time.Time) (MailboxSubmissionStaleRecoveryResult, error) {
	if MailboxSubmissionStatus(submission.Status) != MailboxSubmissionStatusUnknown || submission.ReconcileAttempts < 1 || submission.ReconcileAttempts > 10 ||
		!mailboxEntSubmissionRecoveryDualFenceMatches(account, submission, draft, databaseNow) {
		return MailboxSubmissionStaleRecoveryResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorVersionConflict)
	}
	update := submission.Update().
		SetStatus(mailsubmission.Status("unknown")).
		ClearCommandLeaseToken().
		ClearCommandLeaseExpiresAt().
		ClearProviderSubmissionID().
		ClearProviderMessageID().
		ClearResultMailMessageID().
		ClearNextAttemptAt().
		SetReconcileAttempts(submission.ReconcileAttempts).
		SetUpdatedAt(databaseNow)
	if submission.ReconcileAttempts == 10 {
		update.SetLastErrorCode(mailsubmission.LastErrorCode(MailboxSubmissionErrorCodeReconciliationExhausted)).
			ClearNextReconcileAt()
	} else {
		update.SetLastErrorCode(mailboxEntSubmissionRecoveryPendingCode(submission.LastErrorCode)).
			SetNextReconcileAt(databaseNow)
	}
	if err := update.Exec(ctx); err != nil {
		return MailboxSubmissionStaleRecoveryResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
	}
	if err := mailboxEntSubmissionRecoveryReleaseExactAccountLease(ctx, account, submission, databaseNow); err != nil {
		return MailboxSubmissionStaleRecoveryResult{}, err
	}
	return mailboxSubmissionStaleRecoveryLeaseReleased(), nil
}

func mailboxEntSubmissionRecoveryPendingCode(code *mailsubmission.LastErrorCode) mailsubmission.LastErrorCode {
	if code != nil {
		switch MailboxSubmissionErrorCode(*code) {
		case MailboxSubmissionErrorCodeOutcomeUnknown, MailboxSubmissionErrorCodeReconciliationRequired, MailboxSubmissionErrorCodeMailboxUnavailable:
			return *code
		}
	}
	return mailsubmission.LastErrorCode(MailboxSubmissionErrorCodeReconciliationRequired)
}

func mailboxEntSubmissionRecoveryReleaseAccepted(ctx context.Context, account *ent.MailAccount, submission *ent.MailSubmission, draft *ent.MailDraft, databaseNow time.Time) (MailboxSubmissionStaleRecoveryResult, error) {
	if MailboxSubmissionStatus(submission.Status) != MailboxSubmissionStatusAccepted ||
		!mailboxEntSubmissionRecoveryDualFenceMatches(account, submission, draft, databaseNow) {
		return MailboxSubmissionStaleRecoveryResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorVersionConflict)
	}
	if err := submission.Update().
		ClearCommandLeaseToken().
		ClearCommandLeaseExpiresAt().
		SetUpdatedAt(databaseNow).
		Exec(ctx); err != nil {
		return MailboxSubmissionStaleRecoveryResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
	}
	if err := mailboxEntSubmissionRecoveryReleaseExactAccountLease(ctx, account, submission, databaseNow); err != nil {
		return MailboxSubmissionStaleRecoveryResult{}, err
	}
	return mailboxSubmissionStaleRecoveryLeaseReleased(), nil
}

func mailboxEntSubmissionRecoveryReleaseExactAccountLease(ctx context.Context, account *ent.MailAccount, submission *ent.MailSubmission, databaseNow time.Time) error {
	if account == nil || submission == nil || account.MutationLeaseToken == nil || account.MutationLeaseUntil == nil ||
		submission.CommandLeaseToken == nil || submission.CommandLeaseExpiresAt == nil ||
		*account.MutationLeaseToken != *submission.CommandLeaseToken || !account.MutationLeaseUntil.Equal(*submission.CommandLeaseExpiresAt) ||
		account.MutationEpoch != submission.CommandFence {
		return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorVersionConflict)
	}
	nextEpoch, ok := mailboxEntNextEpoch(account.MutationEpoch)
	if !ok {
		return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
	}
	if err := account.Update().
		ClearMutationLeaseToken().
		ClearMutationLeaseUntil().
		SetMutationEpoch(nextEpoch).
		SetSyncScheduleUpdatedAt(databaseNow).
		Exec(ctx); err != nil {
		return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
	}
	return nil
}

func mailboxEntSubmissionRecoveryDraftError(err error) error {
	if ent.IsNotFound(err) {
		return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
	}
	return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
}

func mailboxEntSubmissionRecoveryRepositoryError(err error) error {
	var repositoryErr *MailboxSubmissionRepositoryError
	if errors.As(err, &repositoryErr) {
		return repositoryErr
	}
	return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
}
