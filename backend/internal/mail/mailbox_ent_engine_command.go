package mail

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
)

// BeginEngineCommand claims the account-wide durable engine-command lease.
func (repository *MailboxEntRepository) BeginEngineCommand(ctx context.Context, request MailboxEngineCommandBeginRequest) (MailboxEngineCommandBeginResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || request.AccountID == uuid.Nil {
		return MailboxEngineCommandBeginResult{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}

	var result MailboxEngineCommandBeginResult
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, err := transaction.MailAccount.Query().
			Where(mailaccount.IDEQ(request.AccountID)).
			ForUpdate().
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorNotFound)
			}
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
		}
		if account.ID != request.AccountID {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
		}

		databaseWallClock, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return mailboxEntEngineCommandRepositoryError(err)
		}
		leaseUntil, ok := mailboxEntCheckedTimeAdd(databaseWallClock, MailboxEngineCommandLeaseDuration)
		if !ok {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
		}

		switch account.Status {
		case mailaccount.StatusDeleting:
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorAccountDeleting)
		case mailaccount.StatusActive:
		default:
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorStateConflict)
		}
		if !account.DesiredEnabled {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorStateConflict)
		}
		if account.EngineAccountID == nil || !syncValidOpaque(*account.EngineAccountID, MaxEngineOpaqueIDBytes) || account.MutationEpoch < 0 {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
		}
		if (account.MutationLeaseToken == nil) != (account.MutationLeaseUntil == nil) ||
			(account.MutationLeaseToken != nil && *account.MutationLeaseToken == uuid.Nil) ||
			(account.MutationLeaseUntil != nil && !mailboxEntValidTime(*account.MutationLeaseUntil)) {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
		}
		_, hasSyncLease, syncLeaseValid := mailboxEntSyncLeaseFence(account)
		if !syncLeaseValid {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
		}

		var busyUntil *time.Time
		if account.MutationLeaseUntil != nil && account.MutationLeaseUntil.After(databaseWallClock) {
			busyUntil = mailboxEntEngineCommandTimeCopy(*account.MutationLeaseUntil)
		}
		if hasSyncLease && account.SyncLeaseUntil.After(databaseWallClock) && (busyUntil == nil || account.SyncLeaseUntil.Before(*busyUntil)) {
			busyUntil = mailboxEntEngineCommandTimeCopy(*account.SyncLeaseUntil)
		}
		if busyUntil != nil {
			result = MailboxEngineCommandBeginResult{BusyUntil: busyUntil}
			return nil
		}

		leaseID := uuid.New()
		nextMutationEpoch, ok := mailboxEntNextEpoch(account.MutationEpoch)
		if !ok {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
		}
		fence := MailboxEngineCommandFence{
			AccountID:     request.AccountID,
			LeaseID:       leaseID,
			MutationEpoch: nextMutationEpoch,
			LeaseUntil:    leaseUntil,
		}
		if !mailboxEntEngineCommandFenceValid(fence) {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
		}
		if err := account.Update().
			SetMutationLeaseToken(leaseID).
			SetMutationLeaseUntil(leaseUntil).
			SetMutationEpoch(nextMutationEpoch).
			SetSyncScheduleUpdatedAt(databaseWallClock).
			Exec(ctx); err != nil {
			return mailboxEntEngineCommandRepositoryError(mailboxEntSafeError(err))
		}

		result = MailboxEngineCommandBeginResult{
			Claimed:         true,
			EngineAccountID: *account.EngineAccountID,
			Fence:           &fence,
		}
		return nil
	})
	if err != nil {
		return MailboxEngineCommandBeginResult{}, mailboxEntEngineCommandRepositoryError(err)
	}
	return result, nil
}

func mailboxEntEngineCommandFenceValid(fence MailboxEngineCommandFence) bool {
	return fence.AccountID != uuid.Nil && fence.LeaseID != uuid.Nil && fence.MutationEpoch > 0 && mailboxEntValidTime(fence.LeaseUntil)
}

func mailboxEntEngineCommandRepositoryError(err error) error {
	var repositoryError *MailboxDraftRepositoryError
	if errors.As(err, &repositoryError) && repositoryError != nil {
		return NewMailboxDraftRepositoryError(repositoryError.Kind())
	}
	switch {
	case errors.Is(err, ErrMailboxServiceInvalid):
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	case errors.Is(err, ErrMailboxServiceNotFound):
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorNotFound)
	case errors.Is(err, ErrMailboxServiceDeleting):
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorAccountDeleting)
	default:
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
	}
}

func mailboxEntEngineCommandTimeCopy(value time.Time) *time.Time {
	copy := value
	return &copy
}
