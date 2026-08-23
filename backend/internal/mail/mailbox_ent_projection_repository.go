package mail

import (
	"context"
	"sort"
	"time"

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
			cursor, err := mailboxThreadProjectionEncodeCursor(MailboxThreadProjectionCursor{accountID: scope.AccountID, isPinned: last.IsPinned, lastMessageAt: last.LastMessageAt, id: last.ID})
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

type mailboxThreadProjectionAggregate struct {
	messageCount   int64
	unreadCount    int64
	hasAttachments bool
	lastMessageAt  time.Time
	lastMessageID  uuid.UUID
	snippet        string
}

func mailboxEntThreadProjectionItems(ctx context.Context, transaction *ent.Tx, accountID uuid.UUID, rows []*ent.MailThread) ([]MailboxThreadProjectionItem, error) {
	if ctx == nil || transaction == nil || accountID == uuid.Nil {
		return nil, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if len(rows) == 0 {
		return []MailboxThreadProjectionItem{}, nil
	}
	threadIDs := make([]uuid.UUID, 0, len(rows))
	threads := make(map[uuid.UUID]*ent.MailThread, len(rows))
	for _, row := range rows {
		if !mailboxEntApplyThreadRowValid(row, accountID) || row.DeletedAt != nil {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if _, duplicate := threads[row.ID]; duplicate {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		threads[row.ID] = row
		threadIDs = append(threadIDs, row.ID)
	}
	sort.Slice(threadIDs, func(left, right int) bool { return syncLessUUID(threadIDs[left], threadIDs[right]) })

	messages, err := transaction.MailMessage.Query().
		Where(
			mailmessage.MailAccountIDEQ(accountID),
			mailmessage.MailThreadIDIn(threadIDs...),
			mailmessage.DeletedAtIsNil(),
		).
		Order(ent.Asc(mailmessage.FieldID)).
		All(ctx)
	if err != nil {
		return nil, mailboxEntSafeError(err)
	}
	messageIDs := make([]uuid.UUID, 0, len(messages))
	messagesByID := make(map[uuid.UUID]*ent.MailMessage, len(messages))
	aggregates := make(map[uuid.UUID]mailboxThreadProjectionAggregate, len(rows))
	for _, message := range messages {
		if !mailboxEntApplyMessageRowValid(message, accountID) || message.DeletedAt != nil {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if _, found := threads[message.MailThreadID]; !found {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if _, duplicate := messagesByID[message.ID]; duplicate {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		messagesByID[message.ID] = message
		messageIDs = append(messageIDs, message.ID)

		aggregate := aggregates[message.MailThreadID]
		aggregate.messageCount++
		if !message.IsSeen {
			aggregate.unreadCount++
		}
		aggregate.hasAttachments = aggregate.hasAttachments || message.HasAttachments
		effectiveTime := mailboxThreadProjectionMessageEffectiveTime(message)
		if aggregate.lastMessageID == uuid.Nil || effectiveTime.After(aggregate.lastMessageAt) || effectiveTime.Equal(aggregate.lastMessageAt) && syncLessUUID(aggregate.lastMessageID, message.ID) {
			aggregate.lastMessageAt = effectiveTime
			aggregate.lastMessageID = message.ID
			aggregate.snippet = message.Snippet
		}
		aggregates[message.MailThreadID] = aggregate
	}
	if len(messagesByID) == 0 {
		return nil, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	for _, threadID := range threadIDs {
		row := threads[threadID]
		aggregate, found := aggregates[threadID]
		if !found || aggregate.messageCount != int64(row.MessageCount) || aggregate.unreadCount != int64(row.UnreadCount) || aggregate.hasAttachments != row.HasAttachments || !aggregate.lastMessageAt.Equal(row.LastMessageAt) || aggregate.snippet != row.Snippet {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
	}

	if err := mailboxEntThreadProjectionValidateFolderLinks(ctx, transaction, accountID, messagesByID, messageIDs); err != nil {
		return nil, err
	}
	labelsByThread, err := mailboxEntThreadProjectionLabels(ctx, transaction, accountID, messagesByID, messageIDs)
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

func mailboxThreadProjectionMessageEffectiveTime(message *ent.MailMessage) time.Time {
	if message.SentAt == nil {
		if message.ReceivedAt != nil {
			return message.ReceivedAt.UTC()
		}
		return message.UpdatedAt.UTC()
	}
	if message.ReceivedAt == nil || !message.ReceivedAt.After(*message.SentAt) {
		return message.SentAt.UTC()
	}
	return message.ReceivedAt.UTC()
}

func mailboxEntThreadProjectionValidateFolderLinks(ctx context.Context, transaction *ent.Tx, accountID uuid.UUID, messagesByID map[uuid.UUID]*ent.MailMessage, messageIDs []uuid.UUID) error {
	links, err := transaction.MailMessageFolder.Query().
		Where(mailmessagefolder.MailMessageIDIn(messageIDs...)).
		Order(ent.Asc(mailmessagefolder.FieldID)).
		All(ctx)
	if err != nil {
		return mailboxEntSafeError(err)
	}
	folderIDs := make(map[uuid.UUID]struct{}, len(links))
	seenByMessage := make(map[uuid.UUID]map[uuid.UUID]struct{}, len(messagesByID))
	for _, link := range links {
		if link == nil || link.ID == uuid.Nil || link.MailMessageID == uuid.Nil || link.MailFolderID == uuid.Nil || link.LastSeenSyncEpoch < 0 || link.LastSeenSyncEpoch > 2_147_483_647 || !syncValidTime(link.ProjectedAt) || !syncValidTime(link.CreatedAt) || !syncValidTime(link.UpdatedAt) || link.EngineState != nil && !syncValidSafeText(*link.EngineState, 4_096, false) {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if _, found := messagesByID[link.MailMessageID]; !found {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		seen, found := seenByMessage[link.MailMessageID]
		if !found {
			seen = make(map[uuid.UUID]struct{})
			seenByMessage[link.MailMessageID] = seen
		}
		if _, duplicate := seen[link.MailFolderID]; duplicate {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		seen[link.MailFolderID] = struct{}{}
		folderIDs[link.MailFolderID] = struct{}{}
	}
	if len(folderIDs) == 0 {
		return nil
	}
	ids := mailboxThreadProjectionSortedUUIDSet(folderIDs)
	folders, err := transaction.MailFolder.Query().
		Where(mailfolder.MailAccountIDEQ(accountID), mailfolder.IDIn(ids...), mailfolder.DeletedAtIsNil()).
		Order(ent.Asc(mailfolder.FieldID)).
		All(ctx)
	if err != nil {
		return mailboxEntSafeError(err)
	}
	if len(folders) != len(ids) {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	seenFolders := make(map[uuid.UUID]struct{}, len(folders))
	for _, folder := range folders {
		state, valid := mailboxEntApplyFolderRowState(folder, accountID)
		if !valid || state != mailboxEntApplyFolderActive || !syncValidTime(folder.CreatedAt) || !syncValidTime(folder.UpdatedAt) {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if _, duplicate := seenFolders[folder.ID]; duplicate {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		seenFolders[folder.ID] = struct{}{}
	}
	return nil
}

func mailboxEntThreadProjectionLabels(ctx context.Context, transaction *ent.Tx, accountID uuid.UUID, messagesByID map[uuid.UUID]*ent.MailMessage, messageIDs []uuid.UUID) (map[uuid.UUID][]MailboxThreadProjectionLabel, error) {
	links, err := transaction.MailMessageLabel.Query().
		Where(mailmessagelabel.MailMessageIDIn(messageIDs...)).
		Order(ent.Asc(mailmessagelabel.FieldID)).
		All(ctx)
	if err != nil {
		return nil, mailboxEntSafeError(err)
	}
	labelIDs := make(map[uuid.UUID]struct{}, len(links))
	linksByThread := make(map[uuid.UUID][]*ent.MailMessageLabel)
	seenByMessage := make(map[uuid.UUID]map[uuid.UUID]struct{}, len(messagesByID))
	for _, link := range links {
		if link == nil || link.ID == uuid.Nil || link.MailMessageID == uuid.Nil || link.MailLabelID == uuid.Nil || (!link.Manual && !link.EngineRule) || !syncValidTime(link.CreatedAt) || !syncValidTime(link.UpdatedAt) {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		message, found := messagesByID[link.MailMessageID]
		if !found {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		seen, found := seenByMessage[link.MailMessageID]
		if !found {
			seen = make(map[uuid.UUID]struct{})
			seenByMessage[link.MailMessageID] = seen
		}
		if _, duplicate := seen[link.MailLabelID]; duplicate {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		seen[link.MailLabelID] = struct{}{}
		labelIDs[link.MailLabelID] = struct{}{}
		linksByThread[message.MailThreadID] = append(linksByThread[message.MailThreadID], link)
	}
	if len(labelIDs) == 0 {
		return map[uuid.UUID][]MailboxThreadProjectionLabel{}, nil
	}
	ids := mailboxThreadProjectionSortedUUIDSet(labelIDs)
	rows, err := transaction.MailLabel.Query().
		Where(maillabel.MailAccountIDEQ(accountID), maillabel.IDIn(ids...), maillabel.DeletedAtIsNil()).
		Order(ent.Asc(maillabel.FieldID)).
		All(ctx)
	if err != nil {
		return nil, mailboxEntSafeError(err)
	}
	if len(rows) != len(ids) {
		return nil, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	labels := make(map[uuid.UUID]MailboxThreadProjectionLabel, len(rows))
	for _, row := range rows {
		if !mailboxThreadProjectionLabelRowValid(row, accountID) {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if _, duplicate := labels[row.ID]; duplicate {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		labels[row.ID] = MailboxThreadProjectionLabel{ID: row.ID, Name: row.Name, Color: row.Color, Version: row.Version}
	}

	result := make(map[uuid.UUID][]MailboxThreadProjectionLabel, len(linksByThread))
	for threadID, threadLinks := range linksByThread {
		seen := make(map[uuid.UUID]struct{}, len(threadLinks))
		threadLabels := make([]MailboxThreadProjectionLabel, 0, len(threadLinks))
		for _, link := range threadLinks {
			if _, duplicate := seen[link.MailLabelID]; duplicate {
				continue
			}
			label, found := labels[link.MailLabelID]
			if !found {
				return nil, mailboxServiceError(MailboxServiceErrorInvalid)
			}
			seen[link.MailLabelID] = struct{}{}
			threadLabels = append(threadLabels, label)
		}
		if len(threadLabels) > maxMailboxThreadProjectionLimit {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		sort.Slice(threadLabels, func(left, right int) bool {
			if threadLabels[left].Name != threadLabels[right].Name {
				return threadLabels[left].Name < threadLabels[right].Name
			}
			return syncLessUUID(threadLabels[left].ID, threadLabels[right].ID)
		})
		result[threadID] = threadLabels
	}
	return result, nil
}

func mailboxThreadProjectionLabelRowValid(row *ent.MailLabel, accountID uuid.UUID) bool {
	return row != nil && row.ID != uuid.Nil && row.MailAccountID == accountID && row.DeletedAt == nil && row.Version >= 1 && syncValidSafeText(row.Name, 240, false) && syncValidSafeText(row.NormalizedName, 240, false) && syncValidSafeText(row.Color, 128, false) && syncValidOpaque(row.EngineKeyword, 57) && syncValidTime(row.CreatedAt) && syncValidTime(row.UpdatedAt)
}

func mailboxThreadProjectionSortedUUIDSet(values map[uuid.UUID]struct{}) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(values))
	for value := range values {
		if value == uuid.Nil {
			return nil
		}
		result = append(result, value)
	}
	sort.Slice(result, func(left, right int) bool { return syncLessUUID(result[left], result[right]) })
	return result
}
