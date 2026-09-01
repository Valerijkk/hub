package mail

import (
	"context"
	"sort"
	"strings"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/calendarevent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/maillabel"
	"github.com/valerii/hub/backend/ent/mailfolder"
	"github.com/valerii/hub/backend/ent/mailmessage"
	"github.com/valerii/hub/backend/ent/mailmessagefolder"
	"github.com/valerii/hub/backend/ent/mailmessagelabel"
	"github.com/valerii/hub/backend/ent/mailthread"
	"github.com/valerii/hub/backend/ent/predicate"
	"github.com/valerii/hub/backend/ent/schema"
)

var _ MailboxFolderProjectionRepository = (*MailboxEntRepository)(nil)
var _ MailboxThreadProjectionRepository = (*MailboxEntRepository)(nil)
var _ MailboxThreadDetailProjectionRepository = (*MailboxEntRepository)(nil)

const maxMailboxThreadDetailProjectionFolderLinks = maxMailboxThreadProjectionLimit * maxMailboxThreadProjectionLimit

// ListMailboxFolderProjection reads only the durable Ent projection in the complete authenticated scope.
func (repository *MailboxEntRepository) ListMailboxFolderProjection(ctx context.Context, scope MailboxFolderProjectionScope) (MailboxFolderProjection, error) {
	if repository == nil || repository.client == nil || ctx == nil || scope.ActorID == uuid.Nil || scope.OrganizationID == uuid.Nil || scope.AccountID == uuid.Nil {
		return MailboxFolderProjection{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	var projection MailboxFolderProjection
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, err := transaction.MailAccount.Query().
			Where(
				mailaccount.IDEQ(scope.AccountID),
				mailaccount.UserIDEQ(scope.ActorID),
				mailaccount.OrganizationIDEQ(scope.OrganizationID),
			).
			ForUpdate().
			Only(ctx)
		if err != nil {
			return mailboxEntSafeError(err)
		}
		if account.ID != scope.AccountID || account.UserID != scope.ActorID || account.OrganizationID != scope.OrganizationID {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		switch account.Status {
		case mailaccount.StatusActive:
			if !validMailboxProjectionState(MailboxProjectionState(account.ProjectionState)) || (account.LastSyncedAt != nil && !mailboxEntValidTime(*account.LastSyncedAt)) {
				return mailboxServiceError(MailboxServiceErrorInvalid)
			}
		case mailaccount.StatusDeleting:
			return mailboxServiceError(MailboxServiceErrorDeleting)
		case mailaccount.StatusProvisioning, mailaccount.StatusSuspended, mailaccount.StatusError:
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		default:
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}

		rows, err := transaction.MailFolder.Query().
			Where(
				mailfolder.MailAccountIDEQ(scope.AccountID),
				mailfolder.DeletedAtIsNil(),
				mailfolder.HasAccountWith(
					mailaccount.IDEQ(scope.AccountID),
					mailaccount.UserIDEQ(scope.ActorID),
					mailaccount.OrganizationIDEQ(scope.OrganizationID),
				),
			).
			Order(
				ent.Asc(mailfolder.FieldRole),
				ent.Asc(mailfolder.FieldSortOrder),
				ent.Asc(mailfolder.FieldNormalizedName),
				ent.Asc(mailfolder.FieldID),
			).
			Limit(maxMailboxFolderProjectionItems + 1).
			All(ctx)
		if err != nil {
			return mailboxEntSafeError(err)
		}
		if len(rows) > maxMailboxFolderProjectionItems {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}

		items := make([]MailboxFolderProjectionItem, len(rows))
		for index, row := range rows {
			item, err := mailboxEntFolderProjectionItem(scope, row)
			if err != nil {
				return err
			}
			items[index] = item
		}

		projection = MailboxFolderProjection{
			Items:           items,
			ProjectionState: MailboxProjectionState(account.ProjectionState),
			ProjectionAsOf:  mailboxFolderProjectionTimePointer(account.LastSyncedAt),
		}
		sortMailboxFolderProjectionItems(projection.Items)
		return nil
	})
	if err != nil {
		return MailboxFolderProjection{}, mailboxEntSafeError(err)
	}

	return mailboxFolderProjectionCopy(projection), nil
}

func mailboxEntFolderProjectionItem(scope MailboxFolderProjectionScope, row *ent.MailFolder) (MailboxFolderProjectionItem, error) {
	if row == nil || scope.AccountID == uuid.Nil || row.ID == uuid.Nil || row.MailAccountID != scope.AccountID || row.DeletedAt != nil {
		return MailboxFolderProjectionItem{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if _, valid := mailboxEntApplyFolderRowState(row, scope.AccountID); !valid || !mailboxEntValidTime(row.CreatedAt) || !mailboxEntValidTime(row.UpdatedAt) || !validMailboxFolderProjectionColor(row.Color) {
		return MailboxFolderProjectionItem{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	var role *string
	if row.Role != nil {
		roleValue := string(*row.Role)
		role = &roleValue
	}

	return MailboxFolderProjectionItem{
		ID:             row.ID,
		ParentID:       mailboxFolderProjectionUUIDPointer(row.ParentID),
		Role:           role,
		Name:           row.Name,
		Color:          mailboxFolderProjectionStringPointer(row.Color),
		SortOrder:      row.SortOrder,
		UnreadCount:    row.UnreadCount,
		TotalCount:     row.TotalCount,
		State:          string(row.ControlState),
		Version:        int64(row.Version),
		normalizedName: row.NormalizedName,
	}, nil
}

// ListMailboxThreadProjection reads only the durable live thread projection in the complete authenticated scope.
func (repository *MailboxEntRepository) ListMailboxThreadProjection(ctx context.Context, scope MailboxFolderProjectionScope, query MailboxThreadProjectionQuery) (MailboxThreadProjection, error) {
	if repository == nil || repository.client == nil || ctx == nil || scope.ActorID == uuid.Nil || scope.OrganizationID == uuid.Nil || scope.AccountID == uuid.Nil || !mailboxThreadProjectionQueryValid(query) {
		return MailboxThreadProjection{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	var projection MailboxThreadProjection
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, err := transaction.MailAccount.Query().
			Where(
				mailaccount.IDEQ(scope.AccountID),
				mailaccount.UserIDEQ(scope.ActorID),
				mailaccount.OrganizationIDEQ(scope.OrganizationID),
			).
			ForUpdate().
			Only(ctx)
		if err != nil {
			return mailboxEntSafeError(err)
		}
		if account.ID != scope.AccountID || account.UserID != scope.ActorID || account.OrganizationID != scope.OrganizationID {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		switch account.Status {
		case mailaccount.StatusActive:
			if !validMailboxProjectionState(MailboxProjectionState(account.ProjectionState)) || (account.LastSyncedAt != nil && !mailboxEntValidTime(*account.LastSyncedAt)) {
				return mailboxServiceError(MailboxServiceErrorInvalid)
			}
		case mailaccount.StatusDeleting:
			return mailboxServiceError(MailboxServiceErrorDeleting)
		case mailaccount.StatusProvisioning, mailaccount.StatusSuspended, mailaccount.StatusError:
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		default:
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}

		threadQuery := transaction.MailThread.Query().
			Where(
				mailthread.MailAccountIDEQ(scope.AccountID),
				mailthread.DeletedAtIsNil(),
				mailthread.HasAccountWith(
					mailaccount.IDEQ(scope.AccountID),
					mailaccount.UserIDEQ(scope.ActorID),
					mailaccount.OrganizationIDEQ(scope.OrganizationID),
				),
				mailthread.HasMessagesWith(
					mailmessage.MailAccountIDEQ(scope.AccountID),
					mailmessage.DeletedAtIsNil(),
				),
			)
		if query.Cursor != nil {
			threadQuery = threadQuery.Where(mailboxThreadProjectionAfterCursor(*query.Cursor))
		}
		if query.FolderID != nil {
			threadQuery = threadQuery.Where(mailboxThreadProjectionFolderPredicate(scope.AccountID, *query.FolderID))
		}
		if query.LabelID != nil {
			threadQuery = threadQuery.Where(mailboxThreadProjectionLabelPredicate(scope.AccountID, *query.LabelID))
		}
		if query.Unread != nil && *query.Unread {
			threadQuery = threadQuery.Where(mailthread.HasMessagesWith(
				mailmessage.MailAccountIDEQ(scope.AccountID),
				mailmessage.DeletedAtIsNil(),
				mailmessage.IsSeenEQ(false),
			))
		}
		if query.Unread != nil && !*query.Unread {
			threadQuery = threadQuery.Where(mailthread.Not(mailthread.HasMessagesWith(
				mailmessage.MailAccountIDEQ(scope.AccountID),
				mailmessage.DeletedAtIsNil(),
				mailmessage.IsSeenEQ(false),
			)))
		}
		if query.HasAttachments != nil && *query.HasAttachments {
			threadQuery = threadQuery.Where(mailthread.HasMessagesWith(
				mailmessage.MailAccountIDEQ(scope.AccountID),
				mailmessage.DeletedAtIsNil(),
				mailmessage.HasAttachmentsEQ(true),
			))
		}
		if query.HasAttachments != nil && !*query.HasAttachments {
			threadQuery = threadQuery.Where(mailthread.Not(mailthread.HasMessagesWith(
				mailmessage.MailAccountIDEQ(scope.AccountID),
				mailmessage.DeletedAtIsNil(),
				mailmessage.HasAttachmentsEQ(true),
			)))
		}
		if query.Starred != nil && *query.Starred {
			threadQuery = threadQuery.Where(mailthread.HasMessagesWith(
				mailmessage.MailAccountIDEQ(scope.AccountID),
				mailmessage.DeletedAtIsNil(),
				mailmessage.IsFlaggedEQ(true),
			))
		}
		if query.Starred != nil && !*query.Starred {
			threadQuery = threadQuery.Where(mailthread.Not(mailthread.HasMessagesWith(
				mailmessage.MailAccountIDEQ(scope.AccountID),
				mailmessage.DeletedAtIsNil(),
				mailmessage.IsFlaggedEQ(true),
			)))
		}
		if query.Search != nil {
			threadQuery = threadQuery.Where(mailboxThreadProjectionSearchPredicate(*query.Search))
		}

		rows, err := threadQuery.
			Order(
				ent.Desc(mailthread.FieldIsPinned),
				ent.Desc(mailthread.FieldLastMessageAt),
				ent.Desc(mailthread.FieldID),
			).
			Limit(query.Limit + 1).
			All(ctx)
		if err != nil {
			return mailboxEntSafeError(err)
		}
		hasNext := len(rows) > query.Limit
		if hasNext {
			rows = rows[:query.Limit]
		}
		items, err := mailboxEntThreadProjectionItems(ctx, transaction, scope.AccountID, rows)
		if err != nil {
			return err
		}

		projection = MailboxThreadProjection{
			Items:           items,
			ProjectionState: MailboxProjectionState(account.ProjectionState),
			ProjectionAsOf:  mailboxFolderProjectionTimePointer(account.LastSyncedAt),
		}
		if hasNext {
			last := projection.Items[len(projection.Items)-1]
			cursor, err := mailboxThreadProjectionEncodeCursor(MailboxThreadProjectionCursor{accountID: scope.AccountID, isPinned: last.IsPinned, lastMessageAt: last.LastMessageAt, id: last.ID, filterSignature: mailboxThreadProjectionQuerySignature(query)})
			if err != nil {
				return mailboxServiceError(MailboxServiceErrorInvalid)
			}
			projection.NextCursor = &cursor
		}
		return nil
	})
	if err != nil {
		return MailboxThreadProjection{}, mailboxEntSafeError(err)
	}
	return mailboxThreadProjectionCopy(projection), nil
}

// GetMailboxThreadDetailProjection reads one live owner-scoped thread and a bounded message-summary page.
func (repository *MailboxEntRepository) GetMailboxThreadDetailProjection(ctx context.Context, scope MailboxFolderProjectionScope, query MailboxThreadDetailProjectionQuery) (MailboxThreadDetailProjection, error) {
	if repository == nil || repository.client == nil || ctx == nil || scope.ActorID == uuid.Nil || scope.OrganizationID == uuid.Nil || scope.AccountID == uuid.Nil || !mailboxThreadDetailProjectionQueryValid(query) {
		return MailboxThreadDetailProjection{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	var projection MailboxThreadDetailProjection
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, err := transaction.MailAccount.Query().
			Where(
				mailaccount.IDEQ(scope.AccountID),
				mailaccount.UserIDEQ(scope.ActorID),
				mailaccount.OrganizationIDEQ(scope.OrganizationID),
			).
			ForUpdate().
			Only(ctx)
		if err != nil {
			return mailboxEntSafeError(err)
		}
		if account.ID != scope.AccountID || account.UserID != scope.ActorID || account.OrganizationID != scope.OrganizationID {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		switch account.Status {
		case mailaccount.StatusActive:
			if !validMailboxProjectionState(MailboxProjectionState(account.ProjectionState)) || account.LastSyncedAt != nil && !mailboxEntValidTime(*account.LastSyncedAt) {
				return mailboxServiceError(MailboxServiceErrorInvalid)
			}
		case mailaccount.StatusDeleting:
			return mailboxServiceError(MailboxServiceErrorDeleting)
		case mailaccount.StatusProvisioning, mailaccount.StatusSuspended, mailaccount.StatusError:
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		default:
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}

		thread, err := transaction.MailThread.Query().
			Where(
				mailthread.IDEQ(query.ThreadID),
				mailthread.MailAccountIDEQ(scope.AccountID),
				mailthread.DeletedAtIsNil(),
				mailthread.HasAccountWith(
					mailaccount.IDEQ(scope.AccountID),
					mailaccount.UserIDEQ(scope.ActorID),
					mailaccount.OrganizationIDEQ(scope.OrganizationID),
				),
				mailthread.HasMessagesWith(
					mailmessage.MailAccountIDEQ(scope.AccountID),
					mailmessage.DeletedAtIsNil(),
				),
			).
			ForUpdate().
			Only(ctx)
		if err != nil {
			return mailboxEntSafeError(err)
		}
		threadItems, err := mailboxEntThreadProjectionItems(ctx, transaction, scope.AccountID, []*ent.MailThread{thread})
		if err != nil || len(threadItems) != 1 {
			if err != nil {
				return err
			}
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}

		messageQuery := transaction.MailMessage.Query().
			Where(
				mailmessage.MailAccountIDEQ(scope.AccountID),
				mailmessage.MailThreadIDEQ(query.ThreadID),
				mailmessage.DeletedAtIsNil(),
				mailmessage.HasThreadWith(
					mailthread.IDEQ(query.ThreadID),
					mailthread.MailAccountIDEQ(scope.AccountID),
					mailthread.DeletedAtIsNil(),
					mailthread.HasAccountWith(
						mailaccount.IDEQ(scope.AccountID),
						mailaccount.UserIDEQ(scope.ActorID),
						mailaccount.OrganizationIDEQ(scope.OrganizationID),
					),
				),
			)
		if query.Cursor != nil {
			messageQuery = messageQuery.Where(mailboxThreadDetailProjectionAfterCursor(*query.Cursor))
		}
		messageRows, err := messageQuery.
			Order(
				ent.Asc(mailmessage.FieldCreatedAt),
				ent.Asc(mailmessage.FieldID),
			).
			Limit(query.Limit + 1).
			All(ctx)
		if err != nil {
			return mailboxEntSafeError(err)
		}
		hasNext := len(messageRows) > query.Limit
		if hasNext {
			messageRows = messageRows[:query.Limit]
		}
		messages, err := mailboxEntThreadDetailProjectionMessages(ctx, transaction, scope, query.ThreadID, messageRows)
		if err != nil {
			return err
		}

		projection = MailboxThreadDetailProjection{
			Thread:          threadItems[0],
			Messages:        messages,
			ProjectionState: MailboxProjectionState(account.ProjectionState),
			ProjectionAsOf:  mailboxFolderProjectionTimePointer(account.LastSyncedAt),
		}
		if hasNext {
			last := projection.Messages[len(projection.Messages)-1]
			cursor, err := mailboxThreadDetailProjectionEncodeCursor(MailboxThreadDetailProjectionCursor{
				accountID:       scope.AccountID,
				threadID:        query.ThreadID,
				createdAt:       last.createdAt,
				id:              last.ID,
				filterSignature: mailboxThreadDetailProjectionQuerySignature(),
			})
			if err != nil {
				return mailboxServiceError(MailboxServiceErrorInvalid)
			}
			projection.NextCursor = &cursor
		}
		return nil
	})
	if err != nil {
		return MailboxThreadDetailProjection{}, mailboxEntSafeError(err)
	}
	if !mailboxThreadDetailProjectionResponseWithinBudget(projection) {
		return MailboxThreadDetailProjection{}, errMailboxThreadDetailProjectionResponseLimit
	}
	return mailboxThreadDetailProjectionCopy(projection), nil
}

func mailboxThreadDetailProjectionAfterCursor(cursor MailboxThreadDetailProjectionCursor) predicate.MailMessage {
	return mailmessage.Or(
		mailmessage.CreatedAtGT(cursor.createdAt),
		mailmessage.And(mailmessage.CreatedAtEQ(cursor.createdAt), mailmessage.IDGT(cursor.id)),
	)
}

func mailboxEntThreadDetailProjectionMessages(ctx context.Context, transaction *ent.Tx, scope MailboxFolderProjectionScope, threadID uuid.UUID, rows []*ent.MailMessage) ([]MailboxThreadDetailProjectionMessage, error) {
	if ctx == nil || transaction == nil || scope.ActorID == uuid.Nil || scope.OrganizationID == uuid.Nil || scope.AccountID == uuid.Nil || threadID == uuid.Nil || len(rows) > maxMailboxThreadProjectionLimit {
		return nil, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if len(rows) == 0 {
		return []MailboxThreadDetailProjectionMessage{}, nil
	}
	messageIDs := make([]uuid.UUID, 0, len(rows))
	messages := make(map[uuid.UUID]struct{}, len(rows))
	for _, row := range rows {
		if !mailboxEntApplyMessageRowValid(row, scope.AccountID) || row.DeletedAt != nil || row.MailThreadID != threadID {
			return nil, mailboxServiceError(MailboxServiceErrorNotFound)
		}
		if _, duplicate := messages[row.ID]; duplicate {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		messages[row.ID] = struct{}{}
		messageIDs = append(messageIDs, row.ID)
	}
	sort.Slice(messageIDs, func(left, right int) bool { return syncLessUUID(messageIDs[left], messageIDs[right]) })
	folderIDsByMessage, err := mailboxEntThreadDetailProjectionFolderIDs(ctx, transaction, scope.AccountID, threadID, messageIDs, messages)
	if err != nil {
		return nil, err
	}
	if err := mailboxEntThreadDetailProjectionCalendarLinks(ctx, transaction, scope, rows); err != nil {
		return nil, err
	}

	result := make([]MailboxThreadDetailProjectionMessage, len(rows))
	for index, row := range rows {
		from, err := mailboxEntThreadDetailProjectionAddress(row.From)
		if err != nil {
			return nil, err
		}
		to, err := mailboxEntThreadDetailProjectionAddresses(row.To)
		if err != nil {
			return nil, err
		}
		cc, err := mailboxEntThreadDetailProjectionAddresses(row.Cc)
		if err != nil {
			return nil, err
		}
		var calendarMethod *MailCalendarMethod
		if row.CalendarMethod != nil {
			value := MailCalendarMethod(*row.CalendarMethod)
			calendarMethod = &value
		}
		result[index] = MailboxThreadDetailProjectionMessage{
			ID:              row.ID,
			ThreadID:        row.MailThreadID,
			From:            from,
			To:              to,
			CC:              cc,
			Subject:         row.Subject,
			Snippet:         row.Snippet,
			SentAt:          mailboxFolderProjectionTimePointer(row.SentAt),
			ReceivedAt:      mailboxFolderProjectionTimePointer(row.ReceivedAt),
			Priority:        string(row.Priority),
			SizeBytes:       row.SizeBytes,
			HasAttachments:  row.HasAttachments,
			IsDraft:         row.IsDraft,
			FolderIDs:       append([]uuid.UUID(nil), folderIDsByMessage[row.ID]...),
			Seen:            row.IsSeen,
			Flagged:         row.IsFlagged,
			Answered:        row.IsAnswered,
			Junk:            row.IsJunk,
			CalendarMethod:  calendarMethod,
			createdAt:       row.CreatedAt.UTC(),
		}
	}
	return result, nil
}

func mailboxEntThreadDetailProjectionAddress(value schema.MailStoredAddress) (MailboxThreadDetailProjectionAddress, error) {
	if err := value.Validate(); err != nil {
		return MailboxThreadDetailProjectionAddress{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	result := MailboxThreadDetailProjectionAddress{Email: value.Email}
	if value.Name != "" {
		name := value.Name
		result.Name = &name
	}
	if !validMailboxThreadDetailProjectionAddress(result) {
		return MailboxThreadDetailProjectionAddress{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	return result, nil
}

func mailboxEntThreadDetailProjectionAddresses(values schema.MailStoredAddresses) ([]MailboxThreadDetailProjectionAddress, error) {
	if err := values.Validate(); err != nil {
		return nil, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	result := make([]MailboxThreadDetailProjectionAddress, len(values))
	for index, value := range values {
		address, err := mailboxEntThreadDetailProjectionAddress(value)
		if err != nil {
			return nil, err
		}
		result[index] = address
	}
	return result, nil
}

const mailboxThreadDetailProjectionFolderLinksSQL = `
WITH page_messages AS (SELECT unnest($1::uuid[]) AS id)
SELECT mf.mail_message_id, mf.mail_folder_id
FROM mail_message_folders AS mf
JOIN page_messages AS pm ON pm.id = mf.mail_message_id
JOIN mail_messages AS m ON m.id = mf.mail_message_id
WHERE m.mail_account_id = $2::uuid
  AND m.mail_thread_id = $3::uuid
  AND m.deleted_at IS NULL
ORDER BY mf.mail_message_id ASC, mf.mail_folder_id ASC
LIMIT $4`

func mailboxEntThreadDetailProjectionFolderIDs(ctx context.Context, transaction *ent.Tx, accountID, threadID uuid.UUID, messageIDs []uuid.UUID, messages map[uuid.UUID]struct{}) (map[uuid.UUID][]uuid.UUID, error) {
	if ctx == nil || transaction == nil || accountID == uuid.Nil || threadID == uuid.Nil || len(messageIDs) == 0 || len(messageIDs) > maxMailboxThreadProjectionLimit || len(messageIDs) != len(messages) {
		return nil, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	messageArray, ok := mailboxThreadProjectionUUIDArray(messageIDs)
	if !ok {
		return nil, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	maximum := len(messageIDs) * maxMailboxThreadProjectionLimit
	var rows entsql.Rows
	if err := transaction.Driver().Query(ctx, mailboxThreadDetailProjectionFolderLinksSQL, []any{messageArray, accountID.String(), threadID.String(), maximum + 1}, &rows); err != nil {
		return nil, mailboxEntSafeError(err)
	}
	closed := false
	defer func() {
		if !closed {
			_ = rows.Close()
		}
	}()

	result := make(map[uuid.UUID][]uuid.UUID, len(messageIDs))
	seen := make(map[uuid.UUID]map[uuid.UUID]struct{}, len(messageIDs))
	allFolderIDs := make(map[uuid.UUID]struct{}, maximum)
	count := 0
	for rows.Next() {
		count++
		if count > maximum {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		var messageID, folderID uuid.UUID
		if err := rows.Scan(&messageID, &folderID); err != nil {
			return nil, mailboxEntSafeError(err)
		}
		if _, found := messages[messageID]; !found || folderID == uuid.Nil {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		messageFolders, found := seen[messageID]
		if !found {
			messageFolders = make(map[uuid.UUID]struct{}, maxMailboxThreadProjectionLimit)
			seen[messageID] = messageFolders
		}
		if _, duplicate := messageFolders[folderID]; duplicate || len(messageFolders) == maxMailboxThreadProjectionLimit {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		messageFolders[folderID] = struct{}{}
		allFolderIDs[folderID] = struct{}{}
		result[messageID] = append(result[messageID], folderID)
	}
	if err := rows.Err(); err != nil {
		return nil, mailboxEntSafeError(err)
	}
	if err := rows.Close(); err != nil {
		return nil, mailboxEntSafeError(err)
	}
	closed = true

	if len(allFolderIDs) > 0 {
		uniqueFolderIDs := make([]uuid.UUID, 0, len(allFolderIDs))
		for folderID := range allFolderIDs {
			uniqueFolderIDs = append(uniqueFolderIDs, folderID)
		}
		sort.Slice(uniqueFolderIDs, func(left, right int) bool { return syncLessUUID(uniqueFolderIDs[left], uniqueFolderIDs[right]) })
		folderRows, err := transaction.MailFolder.Query().
			Where(
				mailfolder.IDIn(uniqueFolderIDs...),
				mailfolder.MailAccountIDEQ(accountID),
				mailfolder.ControlStateEQ(mailfolder.ControlStateActive),
				mailfolder.DeletedAtIsNil(),
			).
			Limit(len(uniqueFolderIDs) + 1).
			All(ctx)
		if err != nil {
			return nil, mailboxEntSafeError(err)
		}
		if len(folderRows) != len(uniqueFolderIDs) {
			return nil, mailboxServiceError(MailboxServiceErrorNotFound)
		}
		for _, folderRow := range folderRows {
			if _, err := mailboxEntFolderProjectionItem(MailboxFolderProjectionScope{AccountID: accountID}, folderRow); err != nil {
				return nil, mailboxServiceError(MailboxServiceErrorNotFound)
			}
		}
	}
	for messageID := range messages {
		folderIDs := result[messageID]
		sort.Slice(folderIDs, func(left, right int) bool { return syncLessUUID(folderIDs[left], folderIDs[right]) })
		result[messageID] = folderIDs
	}
	return result, nil
}

func mailboxEntThreadDetailProjectionCalendarLinks(ctx context.Context, transaction *ent.Tx, scope MailboxFolderProjectionScope, rows []*ent.MailMessage) error {
	if ctx == nil || transaction == nil || scope.ActorID == uuid.Nil || scope.OrganizationID == uuid.Nil || scope.AccountID == uuid.Nil || len(rows) == 0 || len(rows) > maxMailboxThreadProjectionLimit {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	eventIDs := make(map[uuid.UUID]struct{}, len(rows))
	for _, row := range rows {
		if row == nil || row.MailAccountID != scope.AccountID {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if row.CalendarEventID != nil {
			if *row.CalendarEventID == uuid.Nil {
				return mailboxServiceError(MailboxServiceErrorInvalid)
			}
			eventIDs[*row.CalendarEventID] = struct{}{}
		}
	}
	if len(eventIDs) == 0 {
		return nil
	}
	ids := make([]uuid.UUID, 0, len(eventIDs))
	for eventID := range eventIDs {
		ids = append(ids, eventID)
	}
	sort.Slice(ids, func(left, right int) bool { return syncLessUUID(ids[left], ids[right]) })
	calendarRows, err := transaction.CalendarEvent.Query().
		Where(
			calendarevent.IDIn(ids...),
			calendarevent.OrganizationIDEQ(scope.OrganizationID),
			calendarevent.CalendarOwnerIDEQ(scope.ActorID),
			calendarevent.DeletedAtIsNil(),
		).
		Limit(len(ids) + 1).
		All(ctx)
	if err != nil {
		return mailboxEntSafeError(err)
	}
	if len(calendarRows) != len(ids) {
		return mailboxServiceError(MailboxServiceErrorNotFound)
	}
	return nil
}

func mailboxThreadProjectionAfterCursor(cursor MailboxThreadProjectionCursor) predicate.MailThread {
	withinPin := mailthread.Or(
		mailthread.LastMessageAtLT(cursor.lastMessageAt),
		mailthread.And(mailthread.LastMessageAtEQ(cursor.lastMessageAt), mailthread.IDLT(cursor.id)),
	)
	if cursor.isPinned {
		return mailthread.Or(
			mailthread.IsPinnedEQ(false),
			mailthread.And(mailthread.IsPinnedEQ(true), withinPin),
		)
	}
	return mailthread.And(mailthread.IsPinnedEQ(false), withinPin)
}

func mailboxThreadProjectionFolderPredicate(accountID, folderID uuid.UUID) predicate.MailThread {
	return mailthread.HasMessagesWith(
		mailmessage.MailAccountIDEQ(accountID),
		mailmessage.DeletedAtIsNil(),
		mailmessage.HasFoldersWith(
			mailmessagefolder.MailFolderIDEQ(folderID),
			mailmessagefolder.HasFolderWith(
				mailfolder.MailAccountIDEQ(accountID),
				mailfolder.ControlStateEQ(mailfolder.ControlStateActive),
				mailfolder.DeletedAtIsNil(),
			),
		),
	)
}

func mailboxThreadProjectionLabelPredicate(accountID, labelID uuid.UUID) predicate.MailThread {
	return mailthread.HasMessagesWith(
		mailmessage.MailAccountIDEQ(accountID),
		mailmessage.DeletedAtIsNil(),
		mailmessage.HasLabelsWith(
			mailmessagelabel.MailLabelIDEQ(labelID),
			mailmessagelabel.HasLabelWith(
				maillabel.MailAccountIDEQ(accountID),
				maillabel.DeletedAtIsNil(),
			),
		),
	)
}

func mailboxEntThreadProjectionItems(ctx context.Context, transaction *ent.Tx, accountID uuid.UUID, rows []*ent.MailThread) ([]MailboxThreadProjectionItem, error) {
	if ctx == nil || transaction == nil || accountID == uuid.Nil {
		return nil, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if len(rows) == 0 {
		return []MailboxThreadProjectionItem{}, nil
	}
	threadIDs := make([]uuid.UUID, 0, len(rows))
	threads := make(map[uuid.UUID]struct{}, len(rows))
	for _, row := range rows {
		if !mailboxEntApplyThreadRowValid(row, accountID) || row.DeletedAt != nil {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if _, duplicate := threads[row.ID]; duplicate {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		threads[row.ID] = struct{}{}
		threadIDs = append(threadIDs, row.ID)
	}
	sort.Slice(threadIDs, func(left, right int) bool { return syncLessUUID(threadIDs[left], threadIDs[right]) })
	labelsByThread, err := mailboxEntThreadProjectionLabels(ctx, transaction, accountID, threadIDs, threads)
	if err != nil {
		return nil, err
	}

	items := make([]MailboxThreadProjectionItem, 0, len(rows))
	for _, row := range rows {
		labels, found := labelsByThread[row.ID]
		if !found {
			labels = []MailboxThreadProjectionLabel{}
		}
		participants := make([]MailboxThreadProjectionParticipant, len(row.ParticipantAddresses))
		for index, address := range row.ParticipantAddresses {
			participants[index] = MailboxThreadProjectionParticipant{Email: address}
		}
		items = append(items, MailboxThreadProjectionItem{
			ID:             row.ID,
			Subject:        row.NormalizedSubject,
			Participants:   participants,
			MessageCount:   row.MessageCount,
			UnreadCount:    row.UnreadCount,
			LastMessageAt:  row.LastMessageAt.UTC(),
			Snippet:        row.Snippet,
			IsPinned:       row.IsPinned,
			HasAttachments: row.HasAttachments,
			Labels:         append([]MailboxThreadProjectionLabel(nil), labels...),
		})
	}
	return items, nil
}

func mailboxThreadProjectionSearchPredicate(search string) predicate.MailThread {
	return predicate.MailThread(func(selector *entsql.Selector) {
		selector.Where(mailboxThreadProjectionFTSPredicate(
			selector.C(mailthread.FieldNormalizedSubject),
			selector.C(mailthread.FieldSnippet),
			search,
		))
	})
}

func mailboxThreadProjectionFTSPredicate(normalizedSubjectColumn, snippetColumn, search string) *entsql.Predicate {
	return entsql.P(func(builder *entsql.Builder) {
		builder.WriteString("to_tsvector('simple'::regconfig, ((")
		builder.Ident(normalizedSubjectColumn).WriteString(")::text || ' '::text) || (")
		builder.Ident(snippetColumn).WriteString(")::text) @@ plainto_tsquery('simple'::regconfig, ").Arg(search).WriteString(")")
	})
}

const mailboxThreadProjectionLabelsSQL = `
WITH page_threads AS (SELECT unnest($1::uuid[]) AS id)
SELECT DISTINCT m.mail_thread_id, l.id, l.name, l.color, l.version
FROM mail_message_labels AS ml
JOIN mail_messages AS m ON m.id = ml.mail_message_id
JOIN mail_labels AS l ON l.id = ml.mail_label_id
JOIN page_threads AS pt ON pt.id = m.mail_thread_id
WHERE m.mail_account_id = $2::uuid
  AND m.deleted_at IS NULL
  AND l.mail_account_id = $2::uuid
  AND l.deleted_at IS NULL
  AND (ml.manual OR ml.engine_rule)
ORDER BY m.mail_thread_id ASC, l.name ASC, l.id ASC
LIMIT $3`

func mailboxEntThreadProjectionLabels(ctx context.Context, transaction *ent.Tx, accountID uuid.UUID, threadIDs []uuid.UUID, threads map[uuid.UUID]struct{}) (map[uuid.UUID][]MailboxThreadProjectionLabel, error) {
	if ctx == nil || transaction == nil || accountID == uuid.Nil || len(threadIDs) == 0 || len(threadIDs) > maxMailboxThreadProjectionLimit || len(threadIDs) != len(threads) {
		return nil, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	threadArray, ok := mailboxThreadProjectionUUIDArray(threadIDs)
	if !ok {
		return nil, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	maximum := len(threadIDs) * maxMailboxThreadProjectionLimit
	var rows entsql.Rows
	if err := transaction.Driver().Query(ctx, mailboxThreadProjectionLabelsSQL, []any{threadArray, accountID.String(), maximum + 1}, &rows); err != nil {
		return nil, mailboxEntSafeError(err)
	}
	closed := false
	defer func() {
		if !closed {
			_ = rows.Close()
		}
	}()

	result := make(map[uuid.UUID][]MailboxThreadProjectionLabel, len(threadIDs))
	seen := make(map[uuid.UUID]map[uuid.UUID]struct{}, len(threadIDs))
	count := 0
	for rows.Next() {
		count++
		if count > maximum {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		var threadID, labelID uuid.UUID
		var name, color string
		var version int64
		if err := rows.Scan(&threadID, &labelID, &name, &color, &version); err != nil {
			return nil, mailboxEntSafeError(err)
		}
		if _, found := threads[threadID]; !found || !mailboxThreadProjectionLabelValuesValid(labelID, name, color, version) {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		labels, found := seen[threadID]
		if !found {
			labels = make(map[uuid.UUID]struct{}, maxMailboxThreadProjectionLimit)
			seen[threadID] = labels
		}
		if _, duplicate := labels[labelID]; duplicate || len(labels) == maxMailboxThreadProjectionLimit {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		labels[labelID] = struct{}{}
		result[threadID] = append(result[threadID], MailboxThreadProjectionLabel{ID: labelID, Name: name, Color: color, Version: version})
	}
	if err := rows.Err(); err != nil {
		return nil, mailboxEntSafeError(err)
	}
	if err := rows.Close(); err != nil {
		return nil, mailboxEntSafeError(err)
	}
	closed = true

	for threadID, labels := range result {
		sort.Slice(labels, func(left, right int) bool {
			if labels[left].Name != labels[right].Name {
				return labels[left].Name < labels[right].Name
			}
			return syncLessUUID(labels[left].ID, labels[right].ID)
		})
		result[threadID] = labels
	}
	return result, nil
}

func mailboxThreadProjectionLabelValuesValid(id uuid.UUID, name, color string, version int64) bool {
	return id != uuid.Nil && version >= 1 && syncValidSafeText(name, 240, false) && syncValidSafeText(color, 128, false) && len(name) <= 240 && len(color) <= 128
}

func mailboxThreadProjectionUUIDArray(values []uuid.UUID) (string, bool) {
	if len(values) == 0 || len(values) > maxMailboxThreadProjectionLimit {
		return "", false
	}
	var builder strings.Builder
	builder.Grow(2 + len(values)*37)
	builder.WriteByte('{')
	for index, value := range values {
		if value == uuid.Nil {
			return "", false
		}
		if index > 0 {
			builder.WriteByte(',')
		}
		builder.WriteString(value.String())
	}
	builder.WriteByte('}')
	return builder.String(), true
}
