package mail

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/lib/pq"
	"github.com/valerii/hub/backend/ent"
)

type mailboxEntSyncProjectionStage string

const (
	mailboxEntSyncProjectionStageFolders    mailboxEntSyncProjectionStage = "folders"
	mailboxEntSyncProjectionStageMessages   mailboxEntSyncProjectionStage = "messages"
	mailboxEntSyncProjectionStageAggregates mailboxEntSyncProjectionStage = "aggregates"
)

func mailboxEntSyncProjectionErrorCategory(err error) string {
	var postgresError *pq.Error
	switch {
	case err == nil:
		return "none"
	case errors.As(err, &postgresError):
		return "postgres"
	case errors.Is(err, context.Canceled):
		return "context_canceled"
	case errors.Is(err, context.DeadlineExceeded):
		return "context_deadline_exceeded"
	default:
		return "other"
	}
}

func mailboxEntLogSyncProjectionError(stage mailboxEntSyncProjectionStage, err error) {
	attributes := []any{
		"apply_substage", "projection",
		"projection_stage", string(stage),
		"error_category", mailboxEntSyncProjectionErrorCategory(err),
	}
	var postgresError *pq.Error
	if errors.As(err, &postgresError) && postgresError != nil {
		attributes = append(attributes, "postgres_sqlstate", string(postgresError.Code))
	}
	slog.Warn("mailbox sync projection atom failed", attributes...)
}

// mailboxEntApplySyncProjection composes the bounded persistence atoms for one
// canonical Mail sync page. Its caller owns the fenced transaction, database
// clock, account progress and outbox effects.
func (r *MailboxEntRepository) mailboxEntApplySyncProjection(ctx context.Context, tx *ent.Tx, canonical MailboxSyncCanonicalApply, transactionTime time.Time) error {
	if r == nil || ctx == nil || tx == nil || !mailboxEntValidTime(transactionTime) {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if err := ctx.Err(); err != nil {
		return mailboxEntSafeError(err)
	}

	result, err := mailboxEntApplyCanonicalSyncResult(canonical)
	if err != nil {
		return err
	}
	accountID := canonical.Fence.AccountID
	if err := r.mailboxEntApplySyncFolders(ctx, tx, accountID, result.Advance.SyncEpoch, result.Folders, transactionTime); err != nil {
		mailboxEntLogSyncProjectionError(mailboxEntSyncProjectionStageFolders, err)
		return err
	}
	if err := r.mailboxEntApplySyncMessages(ctx, tx, canonical, transactionTime); err != nil {
		mailboxEntLogSyncProjectionError(mailboxEntSyncProjectionStageMessages, err)
		return err
	}
	if err := r.mailboxEntApplySyncAggregates(ctx, tx, accountID, result.Advance.SyncEpoch, result.Counts, transactionTime); err != nil {
		mailboxEntLogSyncProjectionError(mailboxEntSyncProjectionStageAggregates, err)
		return err
	}
	return nil
}
