package mail

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/mailsubmission"
	"github.com/valerii/hub/backend/ent/predicate"
)

const mailboxEntMaxSyncAttempts = int32(1<<31 - 1)

func (repository *MailboxEntRepository) ScanDueMailboxSyncAccounts(ctx context.Context, request MailboxSyncWorkerScanRequest) (MailboxSyncWorkerScanResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || request.Limit < 1 || request.Limit > 50 {
		return MailboxSyncWorkerScanResult{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	var result MailboxSyncWorkerScanResult
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		now, err := repository.mailboxEntNow()
		if err != nil {
			return err
		}

		rows, err := transaction.MailAccount.Query().
			Where(
				mailboxEntMailboxSyncSkipLocked(),
				mailaccount.StatusEQ(mailaccount.StatusActive),
				mailaccount.DesiredEnabledEQ(true),
				mailaccount.EngineAccountIDNotNil(),
				mailaccount.NextSyncAtNotNil(),
				mailaccount.NextSyncAtLTE(now),
				mailboxEntMailboxSyncNoLiveLease(now),
				mailboxEntMailboxSyncNoDueQueuedSubmission(now),
			).
			Order(ent.Asc(mailaccount.FieldNextSyncAt), ent.Asc(mailaccount.FieldID)).
			Limit(request.Limit).
			All(ctx)
		if err != nil {
			return mailboxEntSafeError(err)
		}

		accounts := make([]MailboxSyncWorkerDueAccount, 0, len(rows))
		for _, account := range rows {
			dueAccount, err := mailboxEntWorkerDueAccount(account, now)
			if err != nil {
				return err
			}
			accounts = append(accounts, dueAccount)
		}
		result = MailboxSyncWorkerScanResult{SelectedAt: now, Accounts: accounts}
		return nil
	})
	if err != nil {
		return MailboxSyncWorkerScanResult{}, mailboxEntSafeError(err)
	}
	return result, nil
}

func (repository *MailboxEntRepository) RescheduleMailboxSyncWorkerFault(ctx context.Context, request MailboxSyncWorkerFaultRequest) (MailboxSyncWorkerFaultResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxEntValidWorkerFaultRequest(request) {
		return MailboxSyncWorkerFaultResult{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	var result MailboxSyncWorkerFaultResult
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		now, err := repository.mailboxEntNow()
		if err != nil {
			return err
		}
		if now.Before(request.SelectedAt) {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		nextAttemptAt, ok := mailboxEntCheckedTimeAdd(now, request.RetryDelay)
		if !ok {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}

		account, err := transaction.MailAccount.Query().
			Where(
				mailaccount.IDEQ(request.Account.AccountID),
				mailaccount.NextSyncAtEQ(request.Account.DueAt),
				mailaccount.SyncScheduleUpdatedAtEQ(request.Account.UpdatedAt),
				mailaccount.SyncEpochEQ(request.Account.SyncEpoch),
				mailaccount.MutationEpochEQ(request.Account.MutationEpoch),
				mailaccount.SyncLeaseEpochEQ(request.Account.LeaseEpoch),
				mailaccount.SyncAttemptsEQ(request.Account.SyncAttempts),
				mailaccount.StatusEQ(mailaccount.StatusActive),
				mailaccount.DesiredEnabledEQ(true),
				mailaccount.EngineAccountIDNotNil(),
				mailaccount.NextSyncAtLTE(now),
				mailboxEntMailboxSyncNoLiveLease(now),
			).
			ForUpdate().
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil
			}
			return mailboxEntSafeError(err)
		}
		if account.ID != request.Account.AccountID {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if (account.SyncLeaseToken == nil) != (account.SyncLeaseUntil == nil) || (account.MutationLeaseToken == nil) != (account.MutationLeaseUntil == nil) || (account.SyncLeaseUntil != nil && (!mailboxEntValidTime(*account.SyncLeaseUntil) || account.SyncLeaseUntil.After(now))) || (account.MutationLeaseUntil != nil && (!mailboxEntValidTime(*account.MutationLeaseUntil) || account.MutationLeaseUntil.After(now))) {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}

		if err := account.Update().
			SetSyncAttempts(request.Account.SyncAttempts + 1).
			SetNextSyncAt(nextAttemptAt).
			SetSyncScheduleUpdatedAt(now).
			SetSyncErrorCode(mailboxEntWorkerFaultCode(request.Failure)).
			Exec(ctx); err != nil {
			return mailboxEntSafeError(err)
		}
		result = MailboxSyncWorkerFaultResult{
			Rescheduled:   true,
			RecordedAt:    now,
			NextAttemptAt: nextAttemptAt,
			SyncAttempts:  request.Account.SyncAttempts + 1,
		}
		return nil
	})
	if err != nil {
		return MailboxSyncWorkerFaultResult{}, mailboxEntSafeError(err)
	}
	return result, nil
}

func mailboxEntMailboxSyncSkipLocked() predicate.MailAccount {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.SkipLocked))
	}
}

func mailboxEntMailboxSyncNoDueQueuedSubmission(now time.Time) predicate.MailAccount {
	return mailaccount.Not(mailaccount.HasSubmissionsWith(
		mailsubmission.StatusEQ(mailsubmission.Status("queued")),
		mailsubmission.NextAttemptAtNotNil(),
		mailsubmission.NextAttemptAtLTE(now),
	))
}

func mailboxEntMailboxSyncNoLiveLease(now time.Time) predicate.MailAccount {
	return mailaccount.And(
		mailaccount.Or(mailaccount.SyncLeaseUntilIsNil(), mailaccount.SyncLeaseUntilLTE(now)),
		mailaccount.Or(mailaccount.MutationLeaseUntilIsNil(), mailaccount.MutationLeaseUntilLTE(now)),
	)
}

func mailboxEntWorkerDueAccount(account *ent.MailAccount, selectedAt time.Time) (MailboxSyncWorkerDueAccount, error) {
	if account == nil || account.ID == uuid.Nil || account.Status != mailaccount.StatusActive || !account.DesiredEnabled || account.EngineAccountID == nil || !syncValidOpaque(*account.EngineAccountID, MaxEngineOpaqueIDBytes) || account.NextSyncAt == nil || !mailboxEntValidTime(*account.NextSyncAt) || !mailboxEntValidTime(account.SyncScheduleUpdatedAt) || account.NextSyncAt.After(selectedAt) || account.SyncScheduleUpdatedAt.After(selectedAt) || account.SyncEpoch < 0 || account.MutationEpoch < 0 || account.SyncLeaseEpoch <= 0 || account.SyncAttempts < 0 || (account.SyncLeaseToken == nil) != (account.SyncLeaseUntil == nil) || (account.MutationLeaseToken == nil) != (account.MutationLeaseUntil == nil) || (account.SyncLeaseUntil != nil && (!mailboxEntValidTime(*account.SyncLeaseUntil) || account.SyncLeaseUntil.After(selectedAt))) || (account.MutationLeaseUntil != nil && (!mailboxEntValidTime(*account.MutationLeaseUntil) || account.MutationLeaseUntil.After(selectedAt))) {
		return MailboxSyncWorkerDueAccount{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	return MailboxSyncWorkerDueAccount{
		AccountID:     account.ID,
		DueAt:         account.NextSyncAt.UTC(),
		UpdatedAt:     account.SyncScheduleUpdatedAt.UTC(),
		SyncEpoch:     account.SyncEpoch,
		MutationEpoch: account.MutationEpoch,
		LeaseEpoch:    account.SyncLeaseEpoch,
		SyncAttempts:  account.SyncAttempts,
	}, nil
}

func mailboxEntValidWorkerFaultRequest(request MailboxSyncWorkerFaultRequest) bool {
	account := request.Account
	if account.AccountID == uuid.Nil || !mailboxEntValidTime(request.SelectedAt) || !mailboxEntValidTime(account.DueAt) || !mailboxEntValidTime(account.UpdatedAt) || account.DueAt.After(request.SelectedAt) || account.UpdatedAt.After(request.SelectedAt) || account.SyncEpoch < 0 || account.MutationEpoch < 0 || account.LeaseEpoch <= 0 || account.SyncAttempts < 0 || account.SyncAttempts >= mailboxEntMaxSyncAttempts || request.RetryDelay <= 0 || request.RetryDelay%time.Second != 0 {
		return false
	}
	switch request.Failure {
	case MailboxSyncWorkerFaultInvalidResult, MailboxSyncWorkerFaultTimeout, MailboxSyncWorkerFaultPanic, MailboxSyncWorkerFaultUnavailable:
		return true
	default:
		return false
	}
}

func mailboxEntWorkerFaultCode(failure MailboxSyncWorkerFaultKind) mailaccount.SyncErrorCode {
	if failure == MailboxSyncWorkerFaultInvalidResult {
		return mailaccount.SyncErrorCode("MAIL_SYNC_PAGE_REJECTED")
	}
	return mailaccount.SyncErrorCode("MAIL_SYNC_ENGINE_UNAVAILABLE")
}
