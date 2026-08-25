package mail

import (
	"context"
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
// of expired authorizing/submitting command leases. It never makes a provider
// call and its output is not a recovery capability.
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
				mailsubmission.StatusIn(mailsubmission.Status("authorizing"), mailsubmission.Status("submitting")),
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
		if !mailboxEntSubmissionRowValid(submission, account.ID) {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}
		if submission.Status != mailsubmission.Status("authorizing") && submission.Status != mailsubmission.Status("submitting") {
			return nil
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
