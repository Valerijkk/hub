package mail

import (
	"context"
	"crypto/sha256"
	"errors"
	"log/slog"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/maildraft"
	"github.com/valerii/hub/backend/ent/mailsubmission"
	"github.com/valerii/hub/backend/ent/predicate"
)

var _ MailboxSubmissionStaleRecoveryRepository = (*MailboxEntRepository)(nil)

type mailboxEntSubmissionRecoveryStage string

const (
	mailboxEntSubmissionRecoveryStageAccountLookup           mailboxEntSubmissionRecoveryStage = "account_lookup"
	mailboxEntSubmissionRecoveryStageAccountValidation       mailboxEntSubmissionRecoveryStage = "account_validation"
	mailboxEntSubmissionRecoveryStageSubmissionLookup        mailboxEntSubmissionRecoveryStage = "submission_lookup"
	mailboxEntSubmissionRecoveryStageSubmissionState         mailboxEntSubmissionRecoveryStage = "submission_state"
	mailboxEntSubmissionRecoveryStageDraftLookup             mailboxEntSubmissionRecoveryStage = "draft_lookup"
	mailboxEntSubmissionRecoveryStageDraftValidation         mailboxEntSubmissionRecoveryStage = "draft_validation"
	mailboxEntSubmissionRecoveryStageWallClock               mailboxEntSubmissionRecoveryStage = "wall_clock"
	mailboxEntSubmissionRecoveryStageLeaseClassification     mailboxEntSubmissionRecoveryStage = "lease_classification"
	mailboxEntSubmissionRecoveryStageCurrentStateValidation  mailboxEntSubmissionRecoveryStage = "current_state_validation"
	mailboxEntSubmissionRecoveryStageReleasedStateValidation mailboxEntSubmissionRecoveryStage = "released_state_validation"
	mailboxEntSubmissionRecoveryStageOrphanFence             mailboxEntSubmissionRecoveryStage = "orphan_fence"
	mailboxEntSubmissionRecoveryStageDualFenceShape          mailboxEntSubmissionRecoveryStage = "dual_fence_shape"
	mailboxEntSubmissionRecoveryStageDualFenceExpiry         mailboxEntSubmissionRecoveryStage = "dual_fence_expiry"
	mailboxEntSubmissionRecoveryStageDualFenceMatch          mailboxEntSubmissionRecoveryStage = "dual_fence_match"
	mailboxEntSubmissionRecoveryStageStateApply              mailboxEntSubmissionRecoveryStage = "state_apply"
)

type mailboxEntSubmissionRecoveryScanStage string

const (
	mailboxEntSubmissionRecoveryScanStageWallClock        mailboxEntSubmissionRecoveryScanStage = "wall_clock"
	mailboxEntSubmissionRecoveryScanStageQuery            mailboxEntSubmissionRecoveryScanStage = "query"
	mailboxEntSubmissionRecoveryScanStageCandidate        mailboxEntSubmissionRecoveryScanStage = "candidate"
	mailboxEntSubmissionRecoveryScanStageResultValidation mailboxEntSubmissionRecoveryScanStage = "result_validation"
)

// ScanDueMailboxSubmissionRecovery selects a bounded, ordered advisory list
// of expired authorizing/submitting/unknown/accepted command leases. It never
// makes a provider call and its output is not a recovery capability.
func (repository *MailboxEntRepository) ScanDueMailboxSubmissionRecovery(ctx context.Context, request MailboxSubmissionRecoveryScanRequest) (MailboxSubmissionRecoveryScanResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxEntSubmissionRecoveryScanRequestValid(request) {
		return MailboxSubmissionRecoveryScanResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
	}

	var result MailboxSubmissionRecoveryScanResult
	stage := mailboxEntSubmissionRecoveryScanStageWallClock
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		stage = mailboxEntSubmissionRecoveryScanStageWallClock
		selectedAt, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
		}
		stage = mailboxEntSubmissionRecoveryScanStageQuery
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
			stage = mailboxEntSubmissionRecoveryScanStageCandidate
			due, ok := mailboxEntSubmissionRecoveryDueFromRow(submission)
			if !ok {
				return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
			}
			submissions = append(submissions, due)
		}
		scanned := MailboxSubmissionRecoveryScanResult{SelectedAt: selectedAt, Submissions: submissions}
		stage = mailboxEntSubmissionRecoveryScanStageResultValidation
		if !mailboxEntSubmissionRecoveryScanResultValid(request, scanned) {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}
		result = mailboxEntSubmissionRecoveryScanResultCopy(scanned)
		return nil
	})
	if err != nil {
		slog.Warn("mailbox submission stale recovery scan failed", "stage", string(stage))
		return MailboxSubmissionRecoveryScanResult{}, mailboxEntSubmissionRecoveryRepositoryError(err)
	}
	stage = mailboxEntSubmissionRecoveryScanStageResultValidation
	if !mailboxEntSubmissionRecoveryScanResultValid(request, result) {
		slog.Warn("mailbox submission stale recovery scan failed", "stage", string(stage))
		return MailboxSubmissionRecoveryScanResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
	}
	return mailboxEntSubmissionRecoveryScanResultCopy(result), nil
}

// RecoverStaleMailboxSubmission re-reads one advisory submission under the
// Account -> Submission -> Draft lock order. It only resolves already-expired
// dual leases or safely fenced orphaned submission leases and never calls a
// provider, queue, or transport.
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
	stage := mailboxEntSubmissionRecoveryStageAccountLookup
	err = repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		stage = mailboxEntSubmissionRecoveryStageAccountLookup
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
		stage = mailboxEntSubmissionRecoveryStageAccountValidation
		if kind := mailboxEntSubmissionClaimAccountKind(account); kind != "" {
			return NewMailboxSubmissionRepositoryError(kind)
		}

		stage = mailboxEntSubmissionRecoveryStageSubmissionLookup
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
		stage = mailboxEntSubmissionRecoveryStageSubmissionState
		if !mailboxEntSubmissionRecoveryStaleStatus(MailboxSubmissionStatus(submission.Status)) {
			return nil
		}
		if !mailboxEntSubmissionRowValid(submission, account.ID) {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}
		if submission.MailDraftID == nil || *submission.MailDraftID == uuid.Nil {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}

		stage = mailboxEntSubmissionRecoveryStageDraftLookup
		draft, err := transaction.MailDraft.Query().
			Where(maildraft.IDEQ(*submission.MailDraftID), maildraft.MailAccountIDEQ(account.ID)).
			ForUpdate().
			Only(ctx)
		if err != nil {
			return mailboxEntSubmissionRecoveryDraftError(err)
		}
		stage = mailboxEntSubmissionRecoveryStageDraftValidation
		if !mailboxEntDraftReadRowValid(draft, account.ID, *submission.MailDraftID) {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}

		stage = mailboxEntSubmissionRecoveryStageWallClock
		databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
		}
		orphaned := false
		stage = mailboxEntSubmissionRecoveryStageLeaseClassification
		switch mailboxEntSubmissionRecoveryClassifyLeaseState(account, submission) {
		case mailboxEntSubmissionRecoveryLeaseStateReleased:
			stage = mailboxEntSubmissionRecoveryStageReleasedStateValidation
			if !mailboxEntSubmissionRecoveryReleasedStateValid(submission) {
				return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
			}
			return nil
		case mailboxEntSubmissionRecoveryLeaseStateHeld:
			stage = mailboxEntSubmissionRecoveryStageCurrentStateValidation
			if !mailboxEntSubmissionRecoveryCurrentStateValid(submission, account.ID) {
				return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
			}
		case mailboxEntSubmissionRecoveryLeaseStateOrphaned:
			stage = mailboxEntSubmissionRecoveryStageCurrentStateValidation
			if !mailboxEntSubmissionRecoveryCurrentStateValid(submission, account.ID) {
				return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
			}
			orphaned = true
		case mailboxEntSubmissionRecoveryLeaseStateInvalid:
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		case mailboxEntSubmissionRecoveryLeaseStateConflict:
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorVersionConflict)
		default:
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}
		if orphaned {
			stage = mailboxEntSubmissionRecoveryStageOrphanFence
			if !mailboxEntSubmissionRecoveryOrphanedFenceMatches(account, submission, draft, databaseNow) {
				return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorVersionConflict)
			}
		} else {
			stage = mailboxEntSubmissionRecoveryStageDualFenceShape
			if !mailboxEntSubmissionRecoveryDualFenceShapeMatches(account, submission, draft) {
				return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorVersionConflict)
			}
			stage = mailboxEntSubmissionRecoveryStageDualFenceExpiry
			if account.MutationLeaseUntil.After(databaseNow) {
				return nil
			}
			stage = mailboxEntSubmissionRecoveryStageDualFenceMatch
			if !mailboxEntSubmissionRecoveryDualFenceMatches(account, submission, draft, databaseNow) {
				return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorVersionConflict)
			}
		}

		stage = mailboxEntSubmissionRecoveryStageStateApply
		switch MailboxSubmissionStatus(submission.Status) {
		case MailboxSubmissionStatusAuthorizing:
			if submission.AttemptCount >= 1 && submission.AttemptCount <= 7 {
				var updated MailboxSubmissionStaleRecoveryResult
				if orphaned {
					updated, err = mailboxEntSubmissionRecoveryRequeueOrphaned(ctx, account, submission, draft, databaseNow)
				} else {
					updated, err = mailboxEntSubmissionRecoveryRequeue(ctx, account, submission, draft, databaseNow)
				}
				if err != nil {
					return err
				}
				result = updated
				return nil
			}
			if submission.AttemptCount != 8 {
				return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
			}
			var updated MailboxSubmissionStaleRecoveryResult
			if orphaned {
				updated, err = mailboxEntSubmissionRecoveryTerminalOrphaned(ctx, account, submission, draft, databaseNow)
			} else {
				updated, err = mailboxEntSubmissionRecoveryTerminal(ctx, account, submission, draft, databaseNow)
			}
			if err != nil {
				return err
			}
			result = updated
			return nil
		case MailboxSubmissionStatusSubmitting:
			var updated MailboxSubmissionStaleRecoveryResult
			if orphaned {
				updated, err = mailboxEntSubmissionRecoveryUnknownOrphaned(ctx, account, submission, draft, databaseNow)
			} else {
				updated, err = mailboxEntSubmissionRecoveryUnknown(ctx, account, submission, draft, databaseNow)
			}
			if err != nil {
				return err
			}
			result = updated
			return nil
		case MailboxSubmissionStatusUnknown:
			var updated MailboxSubmissionStaleRecoveryResult
			if orphaned {
				updated, err = mailboxEntSubmissionRecoveryReleaseUnknownOrphaned(ctx, account, submission, draft, databaseNow)
			} else {
				updated, err = mailboxEntSubmissionRecoveryReleaseUnknown(ctx, account, submission, draft, databaseNow)
			}
			if err != nil {
				return err
			}
			result = updated
			return nil
		case MailboxSubmissionStatusAccepted:
			var updated MailboxSubmissionStaleRecoveryResult
			if orphaned {
				updated, err = mailboxEntSubmissionRecoveryReleaseAcceptedOrphaned(ctx, account, submission, draft, databaseNow)
			} else {
				updated, err = mailboxEntSubmissionRecoveryReleaseAccepted(ctx, account, submission, draft, databaseNow)
			}
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
		slog.Warn("mailbox submission stale recovery failed", "stage", string(stage))
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
	leaseUntil := (*submission.CommandLeaseExpiresAt).Round(0).UTC()
	due := MailboxSubmissionRecoveryDue{
		SubmissionID: submission.ID,
		Status:       MailboxSubmissionStatus(submission.Status),
		LeaseUntil:   leaseUntil,
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
	mailboxEntSubmissionRecoveryLeaseStateOrphaned
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
	case !accountHeld && submissionHeld:
		return mailboxEntSubmissionRecoveryLeaseStateOrphaned
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
		submission.CommandLeaseToken != nil && submission.DispatchAuthorizedAt != nil && !mailboxEntNullableBytesAbsent(submission.DispatchAuthorizationHash) &&
		submission.ProviderSubmissionID == nil && submission.ProviderMessageID == nil && submission.ResultMailMessageID == nil &&
		submission.NextAttemptAt == nil && submission.ReconcileAttempts >= 1 && submission.ReconcileAttempts <= 10 && submission.NextReconcileAt == nil
}

func mailboxEntSubmissionRecoveryAcceptedLeaseStateValid(submission *ent.MailSubmission, accountID uuid.UUID) bool {
	return mailboxEntSubmissionRecoveryLeaseStateBaseValid(submission, accountID) &&
		MailboxSubmissionStatus(submission.Status) == MailboxSubmissionStatusAccepted && submission.MailDraftID != nil &&
		submission.CommandLeaseToken != nil && submission.DispatchAuthorizedAt != nil && !mailboxEntNullableBytesAbsent(submission.DispatchAuthorizationHash) &&
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

func mailboxEntSubmissionRecoveryOrphanedFenceMatches(account *ent.MailAccount, submission *ent.MailSubmission, draft *ent.MailDraft, databaseNow time.Time) bool {
	if account == nil || submission == nil || draft == nil ||
		account.MutationLeaseToken != nil || account.MutationLeaseUntil != nil ||
		submission.CommandLeaseToken == nil || submission.CommandLeaseExpiresAt == nil ||
		*submission.CommandLeaseToken == uuid.Nil || !mailboxEntValidTime(*submission.CommandLeaseExpiresAt) ||
		submission.CommandLeaseExpiresAt.After(databaseNow) ||
		account.MutationEpoch < 1 || submission.CommandFence < 1 ||
		account.MutationEpoch <= submission.CommandFence ||
		submission.MailDraftID == nil || *submission.MailDraftID == uuid.Nil {
		return false
	}
	binding, ok := mailboxEntSubmissionBindingFromRow(submission)
	return ok && mailboxEntSubmissionClaimDraftMatches(draft, account.ID, *submission.MailDraftID, binding)
}

func mailboxEntSubmissionRecoveryRequeueOrphaned(ctx context.Context, account *ent.MailAccount, submission *ent.MailSubmission, draft *ent.MailDraft, databaseNow time.Time) (MailboxSubmissionStaleRecoveryResult, error) {
	if submission == nil || MailboxSubmissionStatus(submission.Status) != MailboxSubmissionStatusAuthorizing ||
		submission.AttemptCount < 1 || submission.AttemptCount > 7 ||
		!mailboxEntSubmissionRecoveryOrphanedFenceMatches(account, submission, draft, databaseNow) {
		return MailboxSubmissionStaleRecoveryResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorVersionConflict)
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
	return mailboxSubmissionStaleRecoveryRequeued(databaseNow), nil
}

func mailboxEntSubmissionRecoveryTerminalOrphaned(ctx context.Context, account *ent.MailAccount, submission *ent.MailSubmission, draft *ent.MailDraft, databaseNow time.Time) (MailboxSubmissionStaleRecoveryResult, error) {
	if submission == nil || MailboxSubmissionStatus(submission.Status) != MailboxSubmissionStatusAuthorizing ||
		submission.AttemptCount != 8 ||
		!mailboxEntSubmissionRecoveryOrphanedFenceMatches(account, submission, draft, databaseNow) {
		return MailboxSubmissionStaleRecoveryResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorVersionConflict)
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
	return mailboxSubmissionStaleRecoveryFailed(), nil
}

func mailboxEntSubmissionRecoveryUnknownOrphaned(ctx context.Context, account *ent.MailAccount, submission *ent.MailSubmission, draft *ent.MailDraft, databaseNow time.Time) (MailboxSubmissionStaleRecoveryResult, error) {
	if submission == nil || MailboxSubmissionStatus(submission.Status) != MailboxSubmissionStatusSubmitting ||
		submission.AttemptCount < 1 || submission.AttemptCount > 8 ||
		!mailboxEntSubmissionRecoveryOrphanedFenceMatches(account, submission, draft, databaseNow) {
		return MailboxSubmissionStaleRecoveryResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorVersionConflict)
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
	return mailboxSubmissionStaleRecoveryReconciliationRequired(), nil
}

func mailboxEntSubmissionRecoveryReleaseUnknownOrphaned(ctx context.Context, account *ent.MailAccount, submission *ent.MailSubmission, draft *ent.MailDraft, databaseNow time.Time) (MailboxSubmissionStaleRecoveryResult, error) {
	if submission == nil || MailboxSubmissionStatus(submission.Status) != MailboxSubmissionStatusUnknown ||
		submission.ReconcileAttempts < 1 || submission.ReconcileAttempts > 10 ||
		!mailboxEntSubmissionRecoveryOrphanedFenceMatches(account, submission, draft, databaseNow) {
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
	return mailboxSubmissionStaleRecoveryLeaseReleased(), nil
}

func mailboxEntSubmissionRecoveryReleaseAcceptedOrphaned(ctx context.Context, account *ent.MailAccount, submission *ent.MailSubmission, draft *ent.MailDraft, databaseNow time.Time) (MailboxSubmissionStaleRecoveryResult, error) {
	if submission == nil || MailboxSubmissionStatus(submission.Status) != MailboxSubmissionStatusAccepted ||
		!mailboxEntSubmissionRecoveryOrphanedFenceMatches(account, submission, draft, databaseNow) {
		return MailboxSubmissionStaleRecoveryResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorVersionConflict)
	}
	if err := submission.Update().
		ClearCommandLeaseToken().
		ClearCommandLeaseExpiresAt().
		SetUpdatedAt(databaseNow).
		Exec(ctx); err != nil {
		return MailboxSubmissionStaleRecoveryResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
	}
	return mailboxSubmissionStaleRecoveryLeaseReleased(), nil
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
