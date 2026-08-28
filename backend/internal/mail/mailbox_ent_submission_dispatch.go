package mail

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"time"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/maildraft"
	"github.com/valerii/hub/backend/ent/mailsubmission"
)

var _ MailboxSubmissionDispatchRepository = (*MailboxEntRepository)(nil)

// ResolveMailboxSubmissionDispatch durably records the one result observed
// after provider I/O. It never reclaims a stale lease or makes provider calls.
func (repository *MailboxEntRepository) ResolveMailboxSubmissionDispatch(ctx context.Context, request MailboxSubmissionDispatchResolutionRequest) error {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxEntSubmissionDispatchRequestValid(request) {
		return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
	}
	authorization := mailboxSubmissionDispatchAuthorizationCopy(request.Authorization)

	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, err := transaction.MailAccount.Query().
			Where(mailaccount.IDEQ(authorization.fence.AccountID)).
			ForUpdate().
			Only(ctx)
		if err != nil {
			return mailboxEntSubmissionClaimLookupError(err)
		}
		submission, err := transaction.MailSubmission.Query().
			Where(mailsubmission.IDEQ(authorization.submissionID), mailsubmission.MailAccountIDEQ(account.ID)).
			ForUpdate().
			Only(ctx)
		if err != nil {
			return mailboxEntSubmissionClaimLookupError(err)
		}
		if submission.MailDraftID == nil {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}
		draft, err := transaction.MailDraft.Query().
			Where(maildraft.IDEQ(*submission.MailDraftID), maildraft.MailAccountIDEQ(account.ID)).
			ForUpdate().
			Only(ctx)
		if err != nil {
			return mailboxEntSubmissionClaimLookupError(err)
		}
		databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return mailboxEntSubmissionRepositoryError(err)
		}
		if !mailboxEntSubmissionDispatchFenceMatches(account, submission, draft, authorization, databaseNow) {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorVersionConflict)
		}
		nextEpoch, ok := mailboxEntNextEpoch(account.MutationEpoch)
		if !ok {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}

		accepted := request.Resolution == MailboxSubmissionDispatchResolutionAccepted &&
			syncValidOpaque(request.ProviderSubmissionID, MaxEngineOpaqueIDBytes) &&
			syncValidOpaque(request.ProviderMessageID, MaxEngineOpaqueIDBytes) &&
			subtle.ConstantTimeCompare([]byte(request.ProviderMessageID), []byte(submission.EngineMessageID)) == 1
		if accepted {
			err = submission.Update().
				SetStatus(mailsubmission.Status("accepted")).
				ClearCommandLeaseToken().
				ClearCommandLeaseExpiresAt().
				SetProviderSubmissionID(request.ProviderSubmissionID).
				SetProviderMessageID(request.ProviderMessageID).
				ClearResultMailMessageID().
				ClearLastErrorCode().
				ClearNextAttemptAt().
				SetUpdatedAt(databaseNow).
				Exec(ctx)
		} else {
			err = submission.Update().
				SetStatus(mailsubmission.Status("unknown")).
				ClearCommandLeaseToken().
				ClearCommandLeaseExpiresAt().
				ClearProviderSubmissionID().
				ClearProviderMessageID().
				ClearResultMailMessageID().
				SetLastErrorCode(mailsubmission.LastErrorCode(MailboxSubmissionErrorCodeOutcomeUnknown)).
				ClearNextAttemptAt().
				SetReconcileAttempts(0).
				SetNextReconcileAt(databaseNow).
				SetUpdatedAt(databaseNow).
				Exec(ctx)
		}
		if err != nil {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
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
	})
	if err != nil {
		return mailboxEntSubmissionRepositoryError(err)
	}
	return nil
}

func mailboxEntSubmissionDispatchRequestValid(request MailboxSubmissionDispatchResolutionRequest) bool {
	if !mailboxSubmissionDispatchAuthorizationValid(request.Authorization) || !validMailboxSubmissionDispatchResolution(request.Resolution) {
		return false
	}
	switch request.Resolution {
	case MailboxSubmissionDispatchResolutionAccepted:
		return true
	case MailboxSubmissionDispatchResolutionUnknown:
		return request.ProviderSubmissionID == "" && request.ProviderMessageID == ""
	default:
		return false
	}
}

func mailboxEntSubmissionDispatchFenceMatches(account *ent.MailAccount, submission *ent.MailSubmission, draft *ent.MailDraft, authorization MailboxSubmissionDispatchAuthorization, databaseNow time.Time) bool {
	if account == nil || submission == nil || draft == nil || !mailboxSubmissionDispatchAuthorizationValid(authorization) ||
		!mailboxEntValidTime(databaseNow) || mailboxEntSubmissionClaimAccountKind(account) != "" ||
		!mailboxEntSubmissionRowValid(submission, account.ID) || account.ID != authorization.fence.AccountID ||
		submission.ID != authorization.submissionID || submission.MailAccountID != account.ID ||
		MailboxSubmissionStatus(submission.Status) != MailboxSubmissionStatusSubmitting || submission.MailDraftID == nil ||
		account.MutationLeaseToken == nil || account.MutationLeaseUntil == nil || submission.CommandLeaseToken == nil || submission.CommandLeaseExpiresAt == nil ||
		*account.MutationLeaseToken != authorization.fence.LeaseID || *submission.CommandLeaseToken != authorization.fence.LeaseID ||
		!account.MutationLeaseUntil.Equal(authorization.fence.LeaseUntil) || !submission.CommandLeaseExpiresAt.Equal(authorization.fence.LeaseUntil) ||
		account.MutationEpoch != authorization.fence.MutationEpoch || submission.CommandFence != authorization.fence.MutationEpoch ||
		!account.MutationLeaseUntil.After(databaseNow) || !submission.CommandLeaseExpiresAt.After(databaseNow) ||
		submission.DispatchAuthorizedAt == nil || !submission.DispatchAuthorizedAt.Equal(authorization.authorizedAt) ||
		len(*submission.DispatchAuthorizationHash) != sha256.Size ||
		subtle.ConstantTimeCompare(*submission.DispatchAuthorizationHash, authorization.authorizationHash[:]) != 1 ||
		submission.ProviderSubmissionID != nil || submission.ProviderMessageID != nil || submission.ResultMailMessageID != nil ||
		submission.LastErrorCode != nil || submission.NextAttemptAt != nil || submission.ReconcileAttempts != 0 || submission.NextReconcileAt != nil {
		return false
	}
	binding, ok := mailboxEntSubmissionBindingFromRow(submission)
	return ok && mailboxEntSubmissionClaimDraftMatches(draft, account.ID, *submission.MailDraftID, binding)
}
