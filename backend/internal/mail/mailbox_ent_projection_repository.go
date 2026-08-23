package mail

import (
	"context"
	"sort"
	"strings"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/maillabel"
	"github.com/valerii/hub/backend/ent/mailfolder"
	"github.com/valerii/hub/backend/ent/mailmessage"
	"github.com/valerii/hub/backend/ent/mailmessagefolder"
	"github.com/valerii/hub/backend/ent/mailmessagelabel"
	"github.com/valerii/hub/backend/ent/mailthread"
	"github.com/valerii/hub/backend/ent/predicate"
)

var _ MailboxFolderProjectionRepository = (*MailboxEntRepository)(nil)
var _ MailboxThreadProjectionRepository = (*MailboxEntRepository)(nil)

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
		Version:        row.Version,
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
		selector.Where(entsql.Or(
			mailboxThreadProjectionContainsFold(selector.C(mailthread.FieldNormalizedSubject), search),
			mailboxThreadProjectionContainsFold(selector.C(mailthread.FieldSnippet), search),
		))
	})
}

func mailboxThreadProjectionContainsFold(column, value string) *entsql.Predicate {
	return entsql.P(func(builder *entsql.Builder) {
		builder.Ident(column).WriteString(" ILIKE ").Arg("%" + value + "%")
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
