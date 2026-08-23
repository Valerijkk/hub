package mail

import (
	"context"
	"time"

	"github.com/valerii/hub/backend/ent"
)

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
		return err
	}
	if err := r.mailboxEntApplySyncMessages(ctx, tx, canonical, transactionTime); err != nil {
		return err
	}
	return r.mailboxEntApplySyncAggregates(ctx, tx, accountID, result.Advance.SyncEpoch, result.Counts, transactionTime)
}
