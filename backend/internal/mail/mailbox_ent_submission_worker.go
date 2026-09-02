package mail

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailsubmission"
	"github.com/valerii/hub/backend/ent/predicate"
)

var _ MailboxSubmissionWorkerDueScanner = (*MailboxEntRepository)(nil)

// ScanDueMailboxSubmissionIDs selects one bounded, ordered advisory list of
// queued submissions. It never claims, authorizes, resolves, or mutates a
// submission, and it performs no provider or transport I/O.
func (repository *MailboxEntRepository) ScanDueMailboxSubmissionIDs(ctx context.Context, request MailboxSubmissionWorkerDueScanRequest) (MailboxSubmissionWorkerDueScanResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || request.Limit < 1 || request.Limit > 50 {
		return MailboxSubmissionWorkerDueScanResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
	}

	var result MailboxSubmissionWorkerDueScanResult
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		selectedAt, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
		}
		rows, err := transaction.MailSubmission.Query().
			Where(
				mailsubmission.StatusEQ(mailsubmission.Status("queued")),
				mailsubmission.NextAttemptAtNotNil(),
				mailsubmission.NextAttemptAtLTE(selectedAt),
				mailsubmission.CommandLeaseTokenIsNil(),
				mailsubmission.CommandLeaseExpiresAtIsNil(),
				mailsubmission.DispatchAuthorizedAtIsNil(),
				mailsubmission.DispatchAuthorizationHashIsNil(),
				mailsubmission.ProviderSubmissionIDIsNil(),
				mailsubmission.ProviderMessageIDIsNil(),
				mailsubmission.ResultMailMessageIDIsNil(),
				mailsubmission.LastErrorCodeIsNil(),
				mailsubmission.AttemptCountGTE(0),
				mailsubmission.AttemptCountLTE(7),
				mailsubmission.MailDraftIDNotNil(),
				mailboxEntSubmissionWorkerNoAccountDispatchBlocker(selectedAt),
			).
			Order(ent.Asc(mailsubmission.FieldNextAttemptAt), ent.Asc(mailsubmission.FieldID)).
			Limit(request.Limit).
			All(ctx)
		if err != nil {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
		}

		submissionIDs := make([]uuid.UUID, 0, len(rows))
		for _, submission := range rows {
			if !mailboxEntSubmissionWorkerDueScanRowValid(submission, selectedAt) {
				return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
			}
			submissionIDs = append(submissionIDs, submission.ID)
		}
		scanned := MailboxSubmissionWorkerDueScanResult{SubmissionIDs: submissionIDs}
		if !mailboxEntSubmissionWorkerDueScanResultValid(request, scanned) {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}
		result = mailboxEntSubmissionWorkerDueScanResultCopy(scanned)
		return nil
	})
	if err != nil {
		return MailboxSubmissionWorkerDueScanResult{}, mailboxEntSubmissionWorkerDueScanRepositoryError(err)
	}
	if !mailboxEntSubmissionWorkerDueScanResultValid(request, result) {
		return MailboxSubmissionWorkerDueScanResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
	}
	return mailboxEntSubmissionWorkerDueScanResultCopy(result), nil
}

// mailboxEntSubmissionWorkerNoAccountDispatchBlocker keeps an advisory scan
// from filling its bounded batch with work that its account cannot dispatch.
// The claim transaction remains the authoritative concurrency boundary.
func mailboxEntSubmissionWorkerNoAccountDispatchBlocker(selectedAt time.Time) predicate.MailSubmission {
	return func(selector *sql.Selector) {
		blocker := sql.Table(mailsubmission.Table).As("mail_submission_account_dispatch_blocker")
		selector.Where(sql.NotExists(
			sql.Select().
				From(blocker).
				Where(sql.And(
					sql.ColumnsEQ(blocker.C(mailsubmission.FieldMailAccountID), selector.C(mailsubmission.FieldMailAccountID)),
					sql.ColumnsNEQ(blocker.C(mailsubmission.FieldID), selector.C(mailsubmission.FieldID)),
					sql.Or(
						sql.EQ(blocker.C(mailsubmission.FieldStatus), mailsubmission.Status("unknown")),
						sql.And(
							sql.EQ(blocker.C(mailsubmission.FieldStatus), mailsubmission.Status("submitting")),
							sql.NotNull(blocker.C(mailsubmission.FieldCommandLeaseToken)),
							sql.NotNull(blocker.C(mailsubmission.FieldCommandLeaseExpiresAt)),
							sql.LTE(blocker.C(mailsubmission.FieldCommandLeaseExpiresAt), selectedAt),
						),
					),
				)),
		))
	}
}

func mailboxEntSubmissionWorkerDueScanRowValid(submission *ent.MailSubmission, selectedAt time.Time) bool {
	return submission != nil && submission.ID != uuid.Nil && submission.Status == mailsubmission.Status("queued") &&
		submission.NextAttemptAt != nil && mailboxEntValidTime(*submission.NextAttemptAt) && !submission.NextAttemptAt.After(selectedAt) &&
		submission.CommandLeaseToken == nil && submission.CommandLeaseExpiresAt == nil &&
		submission.DispatchAuthorizedAt == nil && mailboxEntNullableBytesAbsent(submission.DispatchAuthorizationHash) &&
		submission.ProviderSubmissionID == nil && submission.ProviderMessageID == nil &&
		submission.ResultMailMessageID == nil && submission.LastErrorCode == nil &&
		submission.AttemptCount >= 0 && submission.AttemptCount <= 7 &&
		submission.MailDraftID != nil && *submission.MailDraftID != uuid.Nil
}

func mailboxEntSubmissionWorkerDueScanResultValid(request MailboxSubmissionWorkerDueScanRequest, result MailboxSubmissionWorkerDueScanResult) bool {
	if request.Limit < 1 || request.Limit > 50 || result.SubmissionIDs == nil || len(result.SubmissionIDs) > request.Limit {
		return false
	}
	seen := make(map[uuid.UUID]struct{}, len(result.SubmissionIDs))
	for _, submissionID := range result.SubmissionIDs {
		if submissionID == uuid.Nil {
			return false
		}
		if _, duplicate := seen[submissionID]; duplicate {
			return false
		}
		seen[submissionID] = struct{}{}
	}
	return true
}

func mailboxEntSubmissionWorkerDueScanResultCopy(result MailboxSubmissionWorkerDueScanResult) MailboxSubmissionWorkerDueScanResult {
	submissionIDs := make([]uuid.UUID, len(result.SubmissionIDs))
	copy(submissionIDs, result.SubmissionIDs)
	return MailboxSubmissionWorkerDueScanResult{SubmissionIDs: submissionIDs}
}

func mailboxEntSubmissionWorkerDueScanRepositoryError(err error) error {
	return mailboxEntSubmissionRepositoryError(err)
}
