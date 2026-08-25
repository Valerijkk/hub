package mail

import (
	"context"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/mailsubmission"
)

var _ MailboxSubmissionReader = (*MailboxEntRepository)(nil)

// ReadMailboxSubmission reads one durable submission receipt after fencing it
// to the authenticated account. It intentionally makes no provider, worker,
// reconciliation, or engine call.
func (repository *MailboxEntRepository) ReadMailboxSubmission(ctx context.Context, request MailboxSubmissionReadRequest) (MailboxSubmissionRecord, error) {
	if repository == nil || repository.client == nil || ctx == nil ||
		request.ActorID == uuid.Nil || request.OrganizationID == uuid.Nil || request.AccountID == uuid.Nil || request.SubmissionID == uuid.Nil {
		return MailboxSubmissionRecord{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
	}
	if ctx.Err() != nil {
		return MailboxSubmissionRecord{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
	}

	var result MailboxSubmissionRecord
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, err := transaction.MailAccount.Query().
			Where(
				mailaccount.IDEQ(request.AccountID),
				mailaccount.UserIDEQ(request.ActorID),
				mailaccount.OrganizationIDEQ(request.OrganizationID),
			).
			ForUpdate().
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorNotFound)
			}
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
		}
		if account == nil || account.ID != request.AccountID || account.UserID != request.ActorID || account.OrganizationID != request.OrganizationID {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}
		if kind := mailboxEntSubmissionNewAccountKind(account); kind != "" {
			return NewMailboxSubmissionRepositoryError(kind)
		}
		projectionState := MailboxProjectionState(account.ProjectionState)
		if !validMailboxProjectionState(projectionState) {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}
		if projectionState != MailboxProjectionReady {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
		}

		row, err := transaction.MailSubmission.Query().
			Where(
				mailsubmission.IDEQ(request.SubmissionID),
				mailsubmission.MailAccountIDEQ(request.AccountID),
			).
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorNotFound)
			}
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
		}
		record, ok := mailboxEntSubmissionRecord(row, request.AccountID)
		if !ok || record.ID != request.SubmissionID {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}
		result = record
		return nil
	})
	if err != nil {
		return MailboxSubmissionRecord{}, mailboxEntSubmissionRepositoryError(err)
	}
	return mailboxSubmissionRecordCopy(result), nil
}
