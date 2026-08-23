package mail

import (
	"context"
	"reflect"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailfolder"
	"github.com/valerii/hub/backend/ent/mailmessage"
	"github.com/valerii/hub/backend/ent/mailmessagefolder"
	"github.com/valerii/hub/backend/ent/mailthread"
)

const mailboxEntApplyAggregateInt32Maximum int64 = 2_147_483_647

type mailboxEntApplyFolderAggregateRow struct {
	MailFolderID uuid.UUID `json:"mail_folder_id"`
	Count        int64     `json:"count"`
}

type mailboxEntApplyThreadAggregateRow struct {
	MailThreadID     uuid.UUID `json:"mail_thread_id"`
	MessageCount     int64     `json:"message_count"`
	UnreadCount      int64     `json:"unread_count"`
	HasAttachments   bool      `json:"has_attachments"`
	LastMessageAt    time.Time `json:"last_message_at"`
	Snippet          string    `json:"snippet"`
}

// mailboxEntApplySyncAggregates recomputes only the bounded folder and thread
// projections affected by an already canonicalized Mail sync page. The caller
// owns account fencing, transaction lifetime and its single database clock.
func (r *MailboxEntRepository) mailboxEntApplySyncAggregates(ctx context.Context, tx *ent.Tx, accountID uuid.UUID, syncEpoch int64, counts MailboxSyncPolicyCountIntent, transactionTime time.Time) error {
	if r == nil || ctx == nil || tx == nil || accountID == uuid.Nil || syncEpoch < 0 || syncEpoch > mailboxEntApplySyncEpochMaximum || !mailboxEntValidTime(transactionTime) || counts.Folders == nil || counts.Threads == nil {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if err := ctx.Err(); err != nil {
		return mailboxEntSafeError(err)
	}
	canonical, err := canonicalizeCountIntent(accountID, counts)
	if err != nil || !reflect.DeepEqual(canonical, counts) {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}

	folderIDs := make([]uuid.UUID, len(counts.Folders))
	for index, folder := range counts.Folders {
		folderIDs[index] = folder.ID
	}
	if err := mailboxEntApplySyncFolderAggregates(ctx, tx, accountID, folderIDs, transactionTime); err != nil {
		return err
	}

	threadIDs := make([]uuid.UUID, len(counts.Threads))
	for index, thread := range counts.Threads {
		threadIDs[index] = thread.ID
	}
	return mailboxEntApplySyncThreadAggregates(ctx, tx, accountID, syncEpoch, threadIDs, transactionTime)
}

func mailboxEntApplySyncFolderAggregates(ctx context.Context, tx *ent.Tx, accountID uuid.UUID, folderIDs []uuid.UUID, transactionTime time.Time) error {
	if ctx == nil || tx == nil || accountID == uuid.Nil || !mailboxEntValidTime(transactionTime) {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if len(folderIDs) == 0 {
		return nil
	}
	rows, err := tx.MailFolder.Query().
		Where(mailfolder.IDIn(folderIDs...)).
		Order(ent.Asc(mailfolder.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return mailboxEntSafeError(err)
	}
	if len(rows) != len(folderIDs) {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}

	byID := make(map[uuid.UUID]*ent.MailFolder, len(rows))
	activeIDs := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		state, valid := mailboxEntApplyFolderRowState(row, accountID)
		if !valid || !mailboxEntValidTime(row.CreatedAt) || !mailboxEntValidTime(row.UpdatedAt) || (row.DeletedAt != nil && !mailboxEntValidTime(*row.DeletedAt)) {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if _, duplicate := byID[row.ID]; duplicate {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		byID[row.ID] = row
		if state == mailboxEntApplyFolderActive {
			activeIDs = append(activeIDs, row.ID)
		}
	}
	for _, folderID := range folderIDs {
		if byID[folderID] == nil {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
	}

	totals, unread, err := mailboxEntApplyLiveFolderCounts(ctx, tx, accountID, activeIDs)
	if err != nil {
		return err
	}
	for _, folderID := range folderIDs {
		row := byID[folderID]
		state, valid := mailboxEntApplyFolderRowState(row, accountID)
		if !valid {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		total, unreadCount := int64(0), int64(0)
		if state == mailboxEntApplyFolderActive {
			total = totals[folderID]
			unreadCount = unread[folderID]
		}
		if total < 0 || unreadCount < 0 || unreadCount > total || total > mailboxEntApplyAggregateInt32Maximum || unreadCount > mailboxEntApplyAggregateInt32Maximum {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if row.TotalCount == int32(total) && row.UnreadCount == int32(unreadCount) {
			continue
		}
		if err := row.Update().
			SetTotalCount(int32(total)).
			SetUnreadCount(int32(unreadCount)).
			SetUpdatedAt(transactionTime).
			Exec(ctx); err != nil {
			return mailboxEntSafeError(err)
		}
	}
	return nil
}

func mailboxEntApplyLiveFolderCounts(ctx context.Context, tx *ent.Tx, accountID uuid.UUID, folderIDs []uuid.UUID) (map[uuid.UUID]int64, map[uuid.UUID]int64, error) {
	if ctx == nil || tx == nil || accountID == uuid.Nil {
		return nil, nil, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	totals := make(map[uuid.UUID]int64, len(folderIDs))
	unread := make(map[uuid.UUID]int64, len(folderIDs))
	if len(folderIDs) == 0 {
		return totals, unread, nil
	}
	targets := make(map[uuid.UUID]struct{}, len(folderIDs))
	for _, folderID := range folderIDs {
		if folderID == uuid.Nil {
			return nil, nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if _, duplicate := targets[folderID]; duplicate {
			return nil, nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		targets[folderID] = struct{}{}
	}
	liveMemberships := []mailboxEntApplyFolderAggregateRow{}
	err := tx.MailMessageFolder.Query().
		Where(
			mailmessagefolder.MailFolderIDIn(folderIDs...),
			mailmessagefolder.HasFolderWith(
				mailfolder.MailAccountIDEQ(accountID),
				mailfolder.ControlStateEQ(mailfolder.ControlStateActive),
				mailfolder.DeletedAtIsNil(),
			),
			mailmessagefolder.HasMessageWith(
				mailmessage.MailAccountIDEQ(accountID),
				mailmessage.DeletedAtIsNil(),
			),
		).
		GroupBy(mailmessagefolder.FieldMailFolderID).
		Aggregate(ent.As(ent.Count(), "count")).
		Scan(ctx, &liveMemberships)
	if err != nil {
		return nil, nil, mailboxEntSafeError(err)
	}
	for _, aggregate := range liveMemberships {
		if aggregate.MailFolderID == uuid.Nil || aggregate.Count < 1 || aggregate.Count > mailboxEntApplyAggregateInt32Maximum {
			return nil, nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if _, found := targets[aggregate.MailFolderID]; !found {
			return nil, nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if _, duplicate := totals[aggregate.MailFolderID]; duplicate {
			return nil, nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		totals[aggregate.MailFolderID] = aggregate.Count
	}

	liveUnreadMemberships := []mailboxEntApplyFolderAggregateRow{}
	err = tx.MailMessageFolder.Query().
		Where(
			mailmessagefolder.MailFolderIDIn(folderIDs...),
			mailmessagefolder.HasFolderWith(
				mailfolder.MailAccountIDEQ(accountID),
				mailfolder.ControlStateEQ(mailfolder.ControlStateActive),
				mailfolder.DeletedAtIsNil(),
			),
			mailmessagefolder.HasMessageWith(
				mailmessage.MailAccountIDEQ(accountID),
				mailmessage.DeletedAtIsNil(),
				mailmessage.IsSeenEQ(false),
			),
		).
		GroupBy(mailmessagefolder.FieldMailFolderID).
		Aggregate(ent.As(ent.Count(), "count")).
		Scan(ctx, &liveUnreadMemberships)
	if err != nil {
		return nil, nil, mailboxEntSafeError(err)
	}
	for _, aggregate := range liveUnreadMemberships {
		if aggregate.MailFolderID == uuid.Nil || aggregate.Count < 1 || aggregate.Count > mailboxEntApplyAggregateInt32Maximum {
			return nil, nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if _, found := targets[aggregate.MailFolderID]; !found {
			return nil, nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if _, duplicate := unread[aggregate.MailFolderID]; duplicate {
			return nil, nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		unread[aggregate.MailFolderID] = aggregate.Count
		if aggregate.Count > totals[aggregate.MailFolderID] {
			return nil, nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
	}
	return totals, unread, nil
}

func mailboxEntApplySyncThreadAggregates(ctx context.Context, tx *ent.Tx, accountID uuid.UUID, syncEpoch int64, threadIDs []uuid.UUID, transactionTime time.Time) error {
	if ctx == nil || tx == nil || accountID == uuid.Nil || syncEpoch < 0 || syncEpoch > mailboxEntApplySyncEpochMaximum || !mailboxEntValidTime(transactionTime) {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if len(threadIDs) == 0 {
		return nil
	}
	rows, err := tx.MailThread.Query().
		Where(mailthread.IDIn(threadIDs...)).
		Order(ent.Asc(mailthread.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return mailboxEntSafeError(err)
	}
	if len(rows) != len(threadIDs) {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	byID := make(map[uuid.UUID]*ent.MailThread, len(rows))
	for _, row := range rows {
		if !mailboxEntApplyThreadRowValid(row, accountID) || !syncValidSafeText(row.Snippet, MaxEngineSnippetBytes, true) {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if _, duplicate := byID[row.ID]; duplicate {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		byID[row.ID] = row
	}
	for _, threadID := range threadIDs {
		if byID[threadID] == nil {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
	}

	aggregates, err := mailboxEntApplyLiveThreadAggregates(ctx, tx, accountID, threadIDs)
	if err != nil {
		return err
	}
	for _, threadID := range threadIDs {
		row := byID[threadID]
		aggregate, present := aggregates[threadID]
		if !present {
			if err := mailboxEntApplyZeroLiveThreadAggregate(ctx, row, syncEpoch, transactionTime); err != nil {
				return err
			}
			continue
		}
		if err := mailboxEntApplyNonEmptyThreadAggregate(ctx, row, syncEpoch, transactionTime, aggregate); err != nil {
			return err
		}
	}
	return nil
}

func mailboxEntApplyLiveThreadAggregates(ctx context.Context, tx *ent.Tx, accountID uuid.UUID, threadIDs []uuid.UUID) (map[uuid.UUID]mailboxEntApplyThreadAggregateRow, error) {
	if ctx == nil || tx == nil || accountID == uuid.Nil {
		return nil, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	targets := make(map[uuid.UUID]struct{}, len(threadIDs))
	for _, threadID := range threadIDs {
		if threadID == uuid.Nil {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if _, duplicate := targets[threadID]; duplicate {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		targets[threadID] = struct{}{}
	}
	records := []mailboxEntApplyThreadAggregateRow{}
	err := tx.MailMessage.Query().
		Where(
			mailmessage.MailAccountIDEQ(accountID),
			mailmessage.MailThreadIDIn(threadIDs...),
			mailmessage.DeletedAtIsNil(),
		).
		GroupBy(mailmessage.FieldMailThreadID).
		Aggregate(
			ent.As(ent.Count(), "message_count"),
			ent.As(mailboxEntApplyThreadUnreadAggregate, "unread_count"),
			ent.As(mailboxEntApplyThreadAttachmentsAggregate, "has_attachments"),
			ent.As(mailboxEntApplyThreadLastMessageAggregate, "last_message_at"),
			ent.As(mailboxEntApplyThreadSnippetAggregate, "snippet"),
		).
		Scan(ctx, &records)
	if err != nil {
		return nil, mailboxEntSafeError(err)
	}
	result := make(map[uuid.UUID]mailboxEntApplyThreadAggregateRow, len(records))
	for _, aggregate := range records {
		if aggregate.MailThreadID == uuid.Nil || aggregate.MessageCount < 1 || aggregate.MessageCount > mailboxEntApplyAggregateInt32Maximum || aggregate.UnreadCount < 0 || aggregate.UnreadCount > aggregate.MessageCount || aggregate.UnreadCount > mailboxEntApplyAggregateInt32Maximum || !mailboxEntValidTime(aggregate.LastMessageAt) || !syncValidSafeText(aggregate.Snippet, MaxEngineSnippetBytes, true) {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if _, found := targets[aggregate.MailThreadID]; !found {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if _, duplicate := result[aggregate.MailThreadID]; duplicate {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		result[aggregate.MailThreadID] = aggregate
	}
	return result, nil
}

func mailboxEntApplyThreadUnreadAggregate(selector *sql.Selector) string {
	return "COALESCE(SUM(CASE WHEN " + selector.C(mailmessage.FieldIsSeen) + " = FALSE THEN 1 ELSE 0 END), 0)"
}

func mailboxEntApplyThreadAttachmentsAggregate(selector *sql.Selector) string {
	return "BOOL_OR(" + selector.C(mailmessage.FieldHasAttachments) + ")"
}

func mailboxEntApplyThreadLastMessageAggregate(selector *sql.Selector) string {
	return "MAX(" + mailboxEntApplyThreadEffectiveTimeSQL(selector) + ")"
}

func mailboxEntApplyThreadSnippetAggregate(selector *sql.Selector) string {
	effectiveTime := mailboxEntApplyThreadEffectiveTimeSQL(selector)
	return "(array_agg(" + selector.C(mailmessage.FieldSnippet) + " ORDER BY " + effectiveTime + " DESC, " + selector.C(mailmessage.FieldID) + " DESC))[1]"
}

func mailboxEntApplyThreadEffectiveTimeSQL(selector *sql.Selector) string {
	sentAt := selector.C(mailmessage.FieldSentAt)
	receivedAt := selector.C(mailmessage.FieldReceivedAt)
	updatedAt := selector.C(mailmessage.FieldUpdatedAt)
	return "CASE WHEN " + sentAt + " IS NULL THEN COALESCE(" + receivedAt + ", " + updatedAt + ") WHEN " + receivedAt + " IS NULL OR " + sentAt + " >= " + receivedAt + " THEN " + sentAt + " ELSE " + receivedAt + " END"
}

func mailboxEntApplyNonEmptyThreadAggregate(ctx context.Context, row *ent.MailThread, syncEpoch int64, transactionTime time.Time, aggregate mailboxEntApplyThreadAggregateRow) error {
	if ctx == nil || row == nil || aggregate.MessageCount < 1 || aggregate.MessageCount > mailboxEntApplyAggregateInt32Maximum || aggregate.UnreadCount < 0 || aggregate.UnreadCount > aggregate.MessageCount || aggregate.UnreadCount > mailboxEntApplyAggregateInt32Maximum || !mailboxEntValidTime(transactionTime) || !mailboxEntValidTime(aggregate.LastMessageAt) || !syncValidSafeText(aggregate.Snippet, MaxEngineSnippetBytes, true) {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	changed := row.DeletedAt != nil || row.MessageCount != int32(aggregate.MessageCount) || row.UnreadCount != int32(aggregate.UnreadCount) || row.HasAttachments != aggregate.HasAttachments || !row.LastMessageAt.Equal(aggregate.LastMessageAt) || row.Snippet != aggregate.Snippet
	if !changed && row.LastSeenSyncEpoch == syncEpoch {
		return nil
	}
	update := row.Update().
		SetMessageCount(int32(aggregate.MessageCount)).
		SetUnreadCount(int32(aggregate.UnreadCount)).
		SetHasAttachments(aggregate.HasAttachments).
		SetLastMessageAt(aggregate.LastMessageAt).
		SetSnippet(aggregate.Snippet).
		SetLastSeenSyncEpoch(syncEpoch)
	if row.DeletedAt != nil {
		update.ClearDeletedAt()
	}
	if changed {
		update.SetUpdatedAt(transactionTime)
	} else {
		update.SetUpdatedAt(row.UpdatedAt)
	}
	if err := update.Exec(ctx); err != nil {
		return mailboxEntSafeError(err)
	}
	return nil
}

func mailboxEntApplyZeroLiveThreadAggregate(ctx context.Context, row *ent.MailThread, syncEpoch int64, transactionTime time.Time) error {
	if ctx == nil || row == nil || !mailboxEntValidTime(transactionTime) {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if row.DeletedAt == nil {
		if err := row.Update().
			SetDeletedAt(transactionTime).
			SetLastSeenSyncEpoch(syncEpoch).
			SetUpdatedAt(transactionTime).
			Exec(ctx); err != nil {
			return mailboxEntSafeError(err)
		}
		return nil
	}
	if row.LastSeenSyncEpoch == syncEpoch {
		return nil
	}
	if err := row.Update().
		SetLastSeenSyncEpoch(syncEpoch).
		SetUpdatedAt(row.UpdatedAt).
		Exec(ctx); err != nil {
		return mailboxEntSafeError(err)
	}
	return nil
}
