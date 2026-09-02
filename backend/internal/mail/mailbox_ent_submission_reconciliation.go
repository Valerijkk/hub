package mail

import (
	"context"
	"crypto/subtle"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/maildraft"
	"github.com/valerii/hub/backend/ent/mailsubmission"
)

var _ MailboxSubmissionReconciliationRepository = (*MailboxEntRepository)(nil)

// ScanUnknownMailboxSubmissions returns only due durable identities. It does
// not lock rows, mutate state, or perform any engine or transport I/O.
func (repository *MailboxEntRepository) ScanUnknownMailboxSubmissions(ctx context.Context, request MailboxSubmissionReconciliationScanRequest) (MailboxSubmissionReconciliationScanResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxSubmissionReconciliationScanRequestValid(request) {
		return MailboxSubmissionReconciliationScanResult{}, newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorInvalid)
	}

	var result MailboxSubmissionReconciliationScanResult
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorRepositoryUnavailable)
		}
		rows, err := transaction.MailSubmission.Query().
			Where(
				mailsubmission.StatusEQ(mailsubmission.Status("unknown")),
				mailsubmission.NextReconcileAtNotNil(),
				mailsubmission.NextReconcileAtLTE(databaseNow),
				mailsubmission.CommandLeaseTokenIsNil(),
				mailsubmission.CommandLeaseExpiresAtIsNil(),
				mailsubmission.DispatchAuthorizedAtNotNil(),
				mailsubmission.DispatchAuthorizationHashNotNil(),
				mailsubmission.ProviderSubmissionIDIsNil(),
				mailsubmission.ProviderMessageIDIsNil(),
				mailsubmission.ResultMailMessageIDIsNil(),
				mailsubmission.NextAttemptAtIsNil(),
				mailsubmission.MailDraftIDNotNil(),
				mailsubmission.AttemptCountGTE(1),
				mailsubmission.AttemptCountLTE(8),
				mailsubmission.ReconcileAttemptsGTE(0),
				mailsubmission.ReconcileAttemptsLTE(9),
				mailsubmission.LastErrorCodeIn(
					mailsubmission.LastErrorCode(MailboxSubmissionErrorCodeOutcomeUnknown),
					mailsubmission.LastErrorCode(MailboxSubmissionErrorCodeReconciliationRequired),
					mailsubmission.LastErrorCode(MailboxSubmissionErrorCodeMailboxUnavailable),
				),
			).
			Order(ent.Asc(mailsubmission.FieldNextReconcileAt), ent.Asc(mailsubmission.FieldID)).
			Limit(request.Limit).
			All(ctx)
		if err != nil {
			return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorRepositoryUnavailable)
		}

		ids := make([]uuid.UUID, 0, len(rows))
		for _, submission := range rows {
			if !mailboxEntSubmissionReconciliationScanRowValid(submission, databaseNow) {
				return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorInvalid)
			}
			ids = append(ids, submission.ID)
		}
		scanned := MailboxSubmissionReconciliationScanResult{SubmissionIDs: ids}
		if !mailboxSubmissionReconciliationScanResultValid(request, scanned) {
			return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorInvalid)
		}
		result = mailboxSubmissionReconciliationScanResultCopy(scanned)
		return nil
	})
	if err != nil {
		return MailboxSubmissionReconciliationScanResult{}, mailboxEntSubmissionReconciliationRepositoryError(err)
	}
	if !mailboxSubmissionReconciliationScanResultValid(request, result) {
		return MailboxSubmissionReconciliationScanResult{}, newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorRepositoryUnavailable)
	}
	return mailboxSubmissionReconciliationScanResultCopy(result), nil
}

// ClaimUnknownMailboxSubmission takes one exact, short reconciliation lease.
// An existing lease is never stolen: stale recovery is the sole expiry owner.
func (repository *MailboxEntRepository) ClaimUnknownMailboxSubmission(ctx context.Context, request MailboxSubmissionReconciliationClaimRequest) (MailboxSubmissionReconciliationClaimResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxSubmissionReconciliationClaimRequestValid(request) {
		return MailboxSubmissionReconciliationClaimResult{}, newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorInvalid)
	}

	// This non-locking lookup obtains only the owner. Every mutable decision is
	// deferred to the ordered transaction below.
	lookup, err := repository.client.MailSubmission.Query().Where(mailsubmission.IDEQ(request.SubmissionID)).Only(ctx)
	if ent.IsNotFound(err) {
		return NewMailboxSubmissionReconciliationNoopResult(), nil
	}
	if err != nil {
		return MailboxSubmissionReconciliationClaimResult{}, newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorRepositoryUnavailable)
	}
	if lookup == nil || lookup.ID != request.SubmissionID || lookup.MailAccountID == uuid.Nil {
		return MailboxSubmissionReconciliationClaimResult{}, newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorInvalid)
	}

	result := NewMailboxSubmissionReconciliationNoopResult()
	err = repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		// Lock order: MailAccount -> MailSubmission -> MailDraft.
		account, err := transaction.MailAccount.Query().
			Where(mailaccount.IDEQ(lookup.MailAccountID)).
			ForUpdate().
			Only(ctx)
		if ent.IsNotFound(err) {
			return nil
		}
		if err != nil {
			return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorRepositoryUnavailable)
		}
		submission, err := transaction.MailSubmission.Query().
			Where(mailsubmission.IDEQ(request.SubmissionID), mailsubmission.MailAccountIDEQ(account.ID)).
			ForUpdate().
			Only(ctx)
		if ent.IsNotFound(err) {
			return nil
		}
		if err != nil {
			return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorRepositoryUnavailable)
		}
		if !mailboxEntSubmissionRowValid(submission, account.ID) || submission.ID != request.SubmissionID {
			return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorInvalid)
		}

		databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorRepositoryUnavailable)
		}
		if kind := mailboxEntSubmissionClaimAccountKind(account); kind != "" {
			return mailboxEntSubmissionReconciliationRepositoryError(NewMailboxSubmissionRepositoryError(kind))
		}
		if MailboxSubmissionStatus(submission.Status) != MailboxSubmissionStatusUnknown {
			return nil
		}
		if submission.ReconcileAttempts == 10 && submission.NextReconcileAt == nil && submission.LastErrorCode != nil &&
			MailboxSubmissionErrorCode(*submission.LastErrorCode) == MailboxSubmissionErrorCodeReconciliationExhausted {
			return nil
		}
		if busyUntil, busy, valid := mailboxEntSubmissionReconciliationOwnedLeaseBusy(account, submission); !valid {
			return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorReconciliationRequired)
		} else if busy {
			busyResult, err := NewMailboxSubmissionReconciliationBusyResult(busyUntil)
			if err != nil {
				return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorInvalid)
			}
			result = busyResult
			return nil
		}
		if busyUntil, busy, valid := mailboxEntSubmissionClaimBusyUntil(account, databaseNow); !valid {
			return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorInvalid)
		} else if busy {
			busyResult, err := NewMailboxSubmissionReconciliationBusyResult(busyUntil)
			if err != nil {
				return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorInvalid)
			}
			result = busyResult
			return nil
		}
		if submission.ReconcileAttempts < 0 || submission.ReconcileAttempts > 9 ||
			submission.NextReconcileAt == nil || submission.NextReconcileAt.After(databaseNow) ||
			submission.LastErrorCode == nil || !mailboxEntSubmissionReconciliationPendingCode(MailboxSubmissionErrorCode(*submission.LastErrorCode)) ||
			submission.MailDraftID == nil || *submission.MailDraftID == uuid.Nil {
			return nil
		}

		draft, err := transaction.MailDraft.Query().
			Where(maildraft.IDEQ(*submission.MailDraftID), maildraft.MailAccountIDEQ(account.ID)).
			ForUpdate().
			Only(ctx)
		if ent.IsNotFound(err) {
			return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorInvalid)
		}
		if err != nil {
			return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorRepositoryUnavailable)
		}
		binding, ok := mailboxEntSubmissionBindingFromRow(submission)
		if !ok || !mailboxEntSubmissionClaimDraftMatches(draft, account.ID, *submission.MailDraftID, binding) {
			return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorInvalid)
		}

		leaseUntil, ok := mailboxEntCheckedTimeAdd(databaseNow, MailboxEngineCommandLeaseDuration)
		if !ok {
			return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorInvalid)
		}
		nextEpoch, ok := mailboxEntNextEpoch(account.MutationEpoch)
		if !ok {
			return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorInvalid)
		}
		leaseID := uuid.New()
		fence := MailboxEngineCommandFence{AccountID: account.ID, LeaseID: leaseID, MutationEpoch: nextEpoch, LeaseUntil: leaseUntil}
		if !mailboxSubmissionClaimFenceValid(fence) {
			return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorInvalid)
		}

		if err := account.Update().
			SetMutationLeaseToken(leaseID).
			SetMutationLeaseUntil(leaseUntil).
			SetMutationEpoch(nextEpoch).
			SetSyncScheduleUpdatedAt(databaseNow).
			Exec(ctx); err != nil {
			return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorRepositoryUnavailable)
		}
		if err := submission.Update().
			SetStatus(mailsubmission.Status("unknown")).
			SetCommandLeaseToken(leaseID).
			SetCommandLeaseExpiresAt(leaseUntil).
			SetCommandFence(nextEpoch).
			SetReconcileAttempts(submission.ReconcileAttempts + 1).
			ClearNextReconcileAt().
			SetUpdatedAt(databaseNow).
			Exec(ctx); err != nil {
			return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorRepositoryUnavailable)
		}

		claimed, err := NewMailboxSubmissionReconciliationClaimedResult(submission.ID, fence, mailboxEntSubmissionEngineSubmission(binding))
		if err != nil {
			return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorInvalid)
		}
		result = claimed
		return nil
	})
	if err != nil {
		return MailboxSubmissionReconciliationClaimResult{}, mailboxEntSubmissionReconciliationRepositoryError(err)
	}
	if !mailboxSubmissionReconciliationClaimResultValid(result) {
		return MailboxSubmissionReconciliationClaimResult{}, newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorRepositoryUnavailable)
	}
	return result, nil
}

// ResolveUnknownMailboxSubmissionReconciliation applies only a sealed claim.
// It never reclaims a lease, calls a provider, or requeues a submission.
func (repository *MailboxEntRepository) ResolveUnknownMailboxSubmissionReconciliation(ctx context.Context, request MailboxSubmissionReconciliationResolveRequest) error {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxSubmissionReconciliationResolveRequestValid(request) {
		return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorInvalid)
	}
	request.Claim = mailboxSubmissionReconciliationClaimCopy(request.Claim)

	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		// Lock order: MailAccount -> MailSubmission -> MailDraft.
		account, err := transaction.MailAccount.Query().
			Where(mailaccount.IDEQ(request.Claim.fence.AccountID)).
			ForUpdate().
			Only(ctx)
		if ent.IsNotFound(err) {
			return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorReconciliationRequired)
		}
		if err != nil {
			return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorRepositoryUnavailable)
		}
		submission, err := transaction.MailSubmission.Query().
			Where(mailsubmission.IDEQ(request.Claim.submissionID), mailsubmission.MailAccountIDEQ(account.ID)).
			ForUpdate().
			Only(ctx)
		if ent.IsNotFound(err) {
			return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorReconciliationRequired)
		}
		if err != nil {
			return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorRepositoryUnavailable)
		}
		if submission.MailDraftID == nil || *submission.MailDraftID == uuid.Nil {
			return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorReconciliationRequired)
		}
		draft, err := transaction.MailDraft.Query().
			Where(maildraft.IDEQ(*submission.MailDraftID), maildraft.MailAccountIDEQ(account.ID)).
			ForUpdate().
			Only(ctx)
		if ent.IsNotFound(err) {
			return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorReconciliationRequired)
		}
		if err != nil {
			return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorRepositoryUnavailable)
		}
		databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorRepositoryUnavailable)
		}
		if !mailboxEntSubmissionReconciliationFenceMatches(account, submission, draft, request, databaseNow) {
			return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorReconciliationRequired)
		}
		nextEpoch, ok := mailboxEntNextEpoch(account.MutationEpoch)
		if !ok {
			return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorInvalid)
		}

		switch request.Observation {
		case MailboxSubmissionReconciliationObservationAccepted:
			if !syncValidOpaque(request.EngineSubmissionID, MaxEngineOpaqueIDBytes) ||
				!syncValidOpaque(request.EngineMessageID, MaxEngineOpaqueIDBytes) ||
				subtle.ConstantTimeCompare([]byte(request.EngineMessageID), []byte(request.Claim.submission.MessageID)) != 1 ||
				subtle.ConstantTimeCompare([]byte(request.EngineMessageID), []byte(submission.EngineMessageID)) != 1 {
				return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorReconciliationRequired)
			}
			if err := submission.Update().
				SetStatus(mailsubmission.Status("accepted")).
				ClearCommandLeaseToken().
				ClearCommandLeaseExpiresAt().
				SetProviderSubmissionID(request.EngineSubmissionID).
				SetProviderMessageID(request.EngineMessageID).
				ClearResultMailMessageID().
				ClearLastErrorCode().
				ClearNextAttemptAt().
				ClearNextReconcileAt().
				SetUpdatedAt(databaseNow).
				Exec(ctx); err != nil {
				return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorRepositoryUnavailable)
			}
		case MailboxSubmissionReconciliationObservationRetry:
			if submission.ReconcileAttempts < 1 || submission.ReconcileAttempts > 10 ||
				!mailboxEntSubmissionReconciliationPendingCode(request.RetryCode) {
				return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorInvalid)
			}
			update := submission.Update().
				SetStatus(mailsubmission.Status("unknown")).
				ClearCommandLeaseToken().
				ClearCommandLeaseExpiresAt().
				ClearNextAttemptAt().
				SetReconcileAttempts(submission.ReconcileAttempts).
				SetUpdatedAt(databaseNow)
			if submission.ReconcileAttempts == 10 {
				update.SetLastErrorCode(mailsubmission.LastErrorCode(MailboxSubmissionErrorCodeReconciliationExhausted)).
					ClearNextReconcileAt()
			} else {
				backoff, ok := mailboxEntSubmissionReconciliationBackoff(submission.ReconcileAttempts)
				if !ok {
					return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorInvalid)
				}
				retryAt, ok := mailboxEntCheckedTimeAdd(databaseNow, backoff)
				if !ok {
					return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorInvalid)
				}
				update.SetLastErrorCode(mailsubmission.LastErrorCode(request.RetryCode)).
					SetNextReconcileAt(retryAt)
			}
			if err := update.Exec(ctx); err != nil {
				return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorRepositoryUnavailable)
			}
		default:
			return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorInvalid)
		}

		if err := account.Update().
			ClearMutationLeaseToken().
			ClearMutationLeaseUntil().
			SetMutationEpoch(nextEpoch).
			SetSyncScheduleUpdatedAt(databaseNow).
			Exec(ctx); err != nil {
			return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorRepositoryUnavailable)
		}
		return nil
	})
	if err != nil {
		return mailboxEntSubmissionReconciliationRepositoryError(err)
	}
	return nil
}

func mailboxEntSubmissionReconciliationScanRowValid(submission *ent.MailSubmission, databaseNow time.Time) bool {
	if submission == nil || submission.ID == uuid.Nil || submission.MailAccountID == uuid.Nil || !mailboxEntValidTime(databaseNow) ||
		!mailboxEntSubmissionRowValid(submission, submission.MailAccountID) || MailboxSubmissionStatus(submission.Status) != MailboxSubmissionStatusUnknown ||
		submission.NextReconcileAt == nil || submission.NextReconcileAt.After(databaseNow) || submission.CommandLeaseToken != nil ||
		submission.CommandLeaseExpiresAt != nil || submission.DispatchAuthorizedAt == nil || mailboxEntNullableBytesAbsent(submission.DispatchAuthorizationHash) ||
		submission.ProviderSubmissionID != nil || submission.ProviderMessageID != nil || submission.ResultMailMessageID != nil ||
		submission.NextAttemptAt != nil || submission.MailDraftID == nil || *submission.MailDraftID == uuid.Nil ||
		submission.AttemptCount < 1 || submission.AttemptCount > 8 || submission.ReconcileAttempts < 0 || submission.ReconcileAttempts > 9 ||
		submission.LastErrorCode == nil {
		return false
	}
	return mailboxEntSubmissionReconciliationPendingCode(MailboxSubmissionErrorCode(*submission.LastErrorCode))
}

func mailboxEntSubmissionReconciliationOwnedLeaseBusy(account *ent.MailAccount, submission *ent.MailSubmission) (time.Time, bool, bool) {
	if account == nil || submission == nil || (submission.CommandLeaseToken == nil) != (submission.CommandLeaseExpiresAt == nil) {
		return time.Time{}, false, false
	}
	if submission.CommandLeaseToken == nil {
		return time.Time{}, false, true
	}
	if account.MutationLeaseToken == nil || account.MutationLeaseUntil == nil || *submission.CommandLeaseToken == uuid.Nil ||
		!mailboxEntValidTime(*submission.CommandLeaseExpiresAt) || *account.MutationLeaseToken != *submission.CommandLeaseToken ||
		!account.MutationLeaseUntil.Equal(*submission.CommandLeaseExpiresAt) || account.MutationEpoch < 1 ||
		submission.CommandFence != account.MutationEpoch {
		return time.Time{}, false, false
	}
	return submission.CommandLeaseExpiresAt.Round(0).UTC(), true, true
}

func mailboxEntSubmissionReconciliationFenceMatches(account *ent.MailAccount, submission *ent.MailSubmission, draft *ent.MailDraft, request MailboxSubmissionReconciliationResolveRequest, databaseNow time.Time) bool {
	if account == nil || submission == nil || draft == nil || !mailboxSubmissionReconciliationResolveRequestValid(request) ||
		!mailboxEntValidTime(databaseNow) || mailboxEntSubmissionClaimAccountKind(account) != "" ||
		!mailboxEntSubmissionRowValid(submission, account.ID) || account.ID != request.Claim.fence.AccountID ||
		submission.ID != request.Claim.submissionID || submission.MailAccountID != account.ID ||
		MailboxSubmissionStatus(submission.Status) != MailboxSubmissionStatusUnknown || submission.MailDraftID == nil ||
		submission.ReconcileAttempts < 1 || submission.ReconcileAttempts > 10 || submission.NextReconcileAt != nil ||
		submission.LastErrorCode == nil || !mailboxEntSubmissionReconciliationPendingCode(MailboxSubmissionErrorCode(*submission.LastErrorCode)) ||
		submission.DispatchAuthorizedAt == nil || mailboxEntNullableBytesAbsent(submission.DispatchAuthorizationHash) ||
		submission.ProviderSubmissionID != nil || submission.ProviderMessageID != nil || submission.ResultMailMessageID != nil ||
		submission.NextAttemptAt != nil || account.MutationLeaseToken == nil || account.MutationLeaseUntil == nil ||
		submission.CommandLeaseToken == nil || submission.CommandLeaseExpiresAt == nil ||
		*account.MutationLeaseToken != request.Claim.fence.LeaseID || *submission.CommandLeaseToken != request.Claim.fence.LeaseID ||
		!account.MutationLeaseUntil.Equal(request.Claim.fence.LeaseUntil) || !submission.CommandLeaseExpiresAt.Equal(request.Claim.fence.LeaseUntil) ||
		account.MutationEpoch != request.Claim.fence.MutationEpoch || submission.CommandFence != request.Claim.fence.MutationEpoch ||
		!account.MutationLeaseUntil.After(databaseNow) || !submission.CommandLeaseExpiresAt.After(databaseNow) ||
		submission.EngineAccountID != request.Claim.submission.AccountID {
		return false
	}
	binding, ok := mailboxEntSubmissionBindingFromRow(submission)
	return ok && mailboxEntSubmissionEngineSubmission(binding) == request.Claim.submission &&
		mailboxEntSubmissionClaimDraftMatches(draft, account.ID, *submission.MailDraftID, binding)
}

func mailboxEntSubmissionReconciliationBackoff(attempt int32) (time.Duration, bool) {
	switch attempt {
	case 1:
		return time.Minute, true
	case 2:
		return 5 * time.Minute, true
	case 3:
		return 15 * time.Minute, true
	case 4:
		return 30 * time.Minute, true
	case 5:
		return time.Hour, true
	case 6:
		return 2 * time.Hour, true
	case 7:
		return 4 * time.Hour, true
	case 8, 9:
		return 6 * time.Hour, true
	default:
		return 0, false
	}
}

func mailboxEntSubmissionReconciliationRepositoryError(err error) error {
	var reconciliationError *MailboxSubmissionReconciliationError
	if errors.As(err, &reconciliationError) && reconciliationError != nil {
		return newMailboxSubmissionReconciliationError(reconciliationError.Kind())
	}
	var repositoryError *MailboxSubmissionRepositoryError
	if errors.As(err, &repositoryError) && repositoryError != nil {
		switch repositoryError.Kind() {
		case MailboxSubmissionRepositoryErrorInvalid:
			return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorInvalid)
		case MailboxSubmissionRepositoryErrorVersionConflict:
			return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorReconciliationRequired)
		default:
			return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorRepositoryUnavailable)
		}
	}
	return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorRepositoryUnavailable)
}
