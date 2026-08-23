package mail

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/maillabel"
	"github.com/valerii/hub/backend/ent/mailfolder"
	"github.com/valerii/hub/backend/ent/mailmessagefolder"
	"github.com/valerii/hub/backend/ent/mailmessagelabel"
)

func mailboxEntApplyMessageFolderLinks(ctx context.Context, tx *ent.Tx, accountID, messageID uuid.UUID, desired []MailboxEventID, syncEpoch int64, transactionTime time.Time) (bool, error) {
	if ctx == nil || tx == nil || accountID == uuid.Nil || messageID == uuid.Nil || desired == nil || syncEpoch < 0 || syncEpoch > 2_147_483_647 || !mailboxEntValidTime(transactionTime) || len(desired) > MaxEngineFolderItems {
		return false, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if err := ctx.Err(); err != nil {
		return false, mailboxEntSafeError(err)
	}

	desiredByFolderID := make(map[uuid.UUID]struct{}, len(desired))
	desiredFolderIDs := make([]uuid.UUID, len(desired))
	for index, value := range desired {
		if err := validateMailboxEventIDForAccount(accountID, value); err != nil {
			return false, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if index > 0 && !syncLessUUID(desired[index-1].ID, value.ID) {
			return false, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		desiredByFolderID[value.ID] = struct{}{}
		desiredFolderIDs[index] = value.ID
	}

	existingLinks, err := tx.MailMessageFolder.Query().
		Where(mailmessagefolder.MailMessageIDEQ(messageID)).
		Order(ent.Asc(mailmessagefolder.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return false, mailboxEntSafeError(err)
	}
	if len(existingLinks) > MaxEngineFolderItems {
		return false, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	existingByFolderID := make(map[uuid.UUID]*ent.MailMessageFolder, len(existingLinks))
	targetFolderIDs := make(map[uuid.UUID]struct{}, len(desired)+len(existingLinks))
	for _, folderID := range desiredFolderIDs {
		targetFolderIDs[folderID] = struct{}{}
	}
	for _, link := range existingLinks {
		if link == nil || link.ID == uuid.Nil || link.MailMessageID != messageID || link.MailFolderID == uuid.Nil || link.LastSeenSyncEpoch < 0 || link.LastSeenSyncEpoch > 2_147_483_647 || link.ProjectedAt.IsZero() || link.CreatedAt.IsZero() || link.UpdatedAt.IsZero() || (link.EngineState != nil && !syncValidSafeText(*link.EngineState, 4_096, false)) {
			return false, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if _, duplicate := existingByFolderID[link.MailFolderID]; duplicate {
			return false, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		existingByFolderID[link.MailFolderID] = link
		targetFolderIDs[link.MailFolderID] = struct{}{}
	}

	orderedTargetFolderIDs := make([]uuid.UUID, 0, len(targetFolderIDs))
	for folderID := range targetFolderIDs {
		orderedTargetFolderIDs = append(orderedTargetFolderIDs, folderID)
	}
	sort.Slice(orderedTargetFolderIDs, func(left, right int) bool { return syncLessUUID(orderedTargetFolderIDs[left], orderedTargetFolderIDs[right]) })
	if len(orderedTargetFolderIDs) != 0 {
		folders, err := tx.MailFolder.Query().
			Where(
				mailfolder.MailAccountIDEQ(accountID),
				mailfolder.IDIn(orderedTargetFolderIDs...),
				mailfolder.ControlStateEQ(mailfolder.ControlStateActive),
				mailfolder.DeletedAtIsNil(),
			).
			Order(ent.Asc(mailfolder.FieldID)).
			ForUpdate().
			All(ctx)
		if err != nil {
			return false, mailboxEntSafeError(err)
		}
		if len(folders) != len(orderedTargetFolderIDs) {
			return false, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		seenFolders := make(map[uuid.UUID]struct{}, len(folders))
		for _, folder := range folders {
			state, valid := mailboxEntApplyFolderRowState(folder, accountID)
			if !valid || state != mailboxEntApplyFolderActive {
				return false, mailboxServiceError(MailboxServiceErrorInvalid)
			}
			if _, duplicate := seenFolders[folder.ID]; duplicate {
				return false, mailboxServiceError(MailboxServiceErrorInvalid)
			}
			seenFolders[folder.ID] = struct{}{}
		}
	}

	obsolete := make([]uuid.UUID, 0, len(existingByFolderID))
	for folderID := range existingByFolderID {
		if _, desired := desiredByFolderID[folderID]; !desired {
			obsolete = append(obsolete, folderID)
		}
	}
	sort.Slice(obsolete, func(left, right int) bool { return syncLessUUID(obsolete[left], obsolete[right]) })
	if len(obsolete) != 0 {
		obsoleteLinkIDs := make([]uuid.UUID, 0, len(obsolete))
		for _, folderID := range obsolete {
			obsoleteLinkIDs = append(obsoleteLinkIDs, existingByFolderID[folderID].ID)
		}
		if _, err := tx.MailMessageFolder.Delete().Where(mailmessagefolder.IDIn(obsoleteLinkIDs...)).Exec(ctx); err != nil {
			return false, mailboxEntSafeError(err)
		}
	}

	semanticChanged := len(obsolete) != 0
	for _, folderID := range desiredFolderIDs {
		link, exists := existingByFolderID[folderID]
		if !exists {
			if _, err := tx.MailMessageFolder.Create().
				SetMailMessageID(messageID).
				SetMailFolderID(folderID).
				SetLastSeenSyncEpoch(syncEpoch).
				SetProjectedAt(transactionTime).
				SetCreatedAt(transactionTime).
				SetUpdatedAt(transactionTime).
				Save(ctx); err != nil {
				return false, mailboxEntSafeError(err)
			}
			semanticChanged = true
			continue
		}
		if err := link.Update().
			SetLastSeenSyncEpoch(syncEpoch).
			SetProjectedAt(transactionTime).
			SetUpdatedAt(transactionTime).
			Exec(ctx); err != nil {
			return false, mailboxEntSafeError(err)
		}
	}
	return semanticChanged, nil
}

func mailboxEntApplyMessageLabelLinks(ctx context.Context, tx *ent.Tx, accountID, messageID uuid.UUID, desired []MailboxSyncPolicyLabelState, transactionTime time.Time) (bool, error) {
	if ctx == nil || tx == nil || accountID == uuid.Nil || messageID == uuid.Nil || desired == nil || !mailboxEntValidTime(transactionTime) || len(desired) > MaxEnginePageItems {
		return false, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if err := ctx.Err(); err != nil {
		return false, mailboxEntSafeError(err)
	}

	desiredByLabelID := make(map[uuid.UUID]MailboxSyncPolicyLabelState, len(desired))
	desiredLabelIDs := make([]uuid.UUID, len(desired))
	for index, value := range desired {
		if err := validateMailboxEventIDForAccount(accountID, value.Label); err != nil || (!value.Manual && !value.EngineRule) {
			return false, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if index > 0 && !syncLessUUID(desired[index-1].Label.ID, value.Label.ID) {
			return false, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		desiredByLabelID[value.Label.ID] = value
		desiredLabelIDs[index] = value.Label.ID
	}

	existingLinks, err := tx.MailMessageLabel.Query().
		Where(mailmessagelabel.MailMessageIDEQ(messageID)).
		Order(ent.Asc(mailmessagelabel.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return false, mailboxEntSafeError(err)
	}
	if len(existingLinks) > MaxEnginePageItems {
		return false, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	existingByLabelID := make(map[uuid.UUID]*ent.MailMessageLabel, len(existingLinks))
	targetLabelIDs := make(map[uuid.UUID]struct{}, len(desired)+len(existingLinks))
	for _, labelID := range desiredLabelIDs {
		targetLabelIDs[labelID] = struct{}{}
	}
	for _, link := range existingLinks {
		if link == nil || link.ID == uuid.Nil || link.MailMessageID != messageID || link.MailLabelID == uuid.Nil || (!link.Manual && !link.EngineRule) || link.CreatedAt.IsZero() || link.UpdatedAt.IsZero() {
			return false, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if _, duplicate := existingByLabelID[link.MailLabelID]; duplicate {
			return false, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		existingByLabelID[link.MailLabelID] = link
		targetLabelIDs[link.MailLabelID] = struct{}{}
	}

	orderedTargetLabelIDs := make([]uuid.UUID, 0, len(targetLabelIDs))
	for labelID := range targetLabelIDs {
		orderedTargetLabelIDs = append(orderedTargetLabelIDs, labelID)
	}
	sort.Slice(orderedTargetLabelIDs, func(left, right int) bool { return syncLessUUID(orderedTargetLabelIDs[left], orderedTargetLabelIDs[right]) })
	if len(orderedTargetLabelIDs) != 0 {
		labels, err := tx.MailLabel.Query().
			Where(maillabel.MailAccountIDEQ(accountID), maillabel.IDIn(orderedTargetLabelIDs...), maillabel.DeletedAtIsNil()).
			Order(ent.Asc(maillabel.FieldID)).
			ForUpdate().
			All(ctx)
		if err != nil {
			return false, mailboxEntSafeError(err)
		}
		if len(labels) != len(orderedTargetLabelIDs) {
			return false, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		seenLabels := make(map[uuid.UUID]struct{}, len(labels))
		for _, label := range labels {
			if label == nil || label.ID == uuid.Nil || label.MailAccountID != accountID || label.DeletedAt != nil {
				return false, mailboxServiceError(MailboxServiceErrorInvalid)
			}
			if _, duplicate := seenLabels[label.ID]; duplicate {
				return false, mailboxServiceError(MailboxServiceErrorInvalid)
			}
			seenLabels[label.ID] = struct{}{}
		}
	}

	obsolete := make([]uuid.UUID, 0, len(existingByLabelID))
	for labelID := range existingByLabelID {
		if _, desired := desiredByLabelID[labelID]; !desired {
			obsolete = append(obsolete, labelID)
		}
	}
	sort.Slice(obsolete, func(left, right int) bool { return syncLessUUID(obsolete[left], obsolete[right]) })
	if len(obsolete) != 0 {
		obsoleteLinkIDs := make([]uuid.UUID, 0, len(obsolete))
		for _, labelID := range obsolete {
			obsoleteLinkIDs = append(obsoleteLinkIDs, existingByLabelID[labelID].ID)
		}
		if _, err := tx.MailMessageLabel.Delete().Where(mailmessagelabel.IDIn(obsoleteLinkIDs...)).Exec(ctx); err != nil {
			return false, mailboxEntSafeError(err)
		}
	}

	semanticChanged := len(obsolete) != 0
	for _, labelID := range desiredLabelIDs {
		desiredState := desiredByLabelID[labelID]
		link, exists := existingByLabelID[labelID]
		if !exists {
			if _, err := tx.MailMessageLabel.Create().
				SetMailMessageID(messageID).
				SetMailLabelID(labelID).
				SetManual(desiredState.Manual).
				SetEngineRule(desiredState.EngineRule).
				SetCreatedAt(transactionTime).
				SetUpdatedAt(transactionTime).
				Save(ctx); err != nil {
				return false, mailboxEntSafeError(err)
			}
			semanticChanged = true
			continue
		}
		if link.Manual == desiredState.Manual && link.EngineRule == desiredState.EngineRule {
			continue
		}
		if err := link.Update().
			SetManual(desiredState.Manual).
			SetEngineRule(desiredState.EngineRule).
			SetUpdatedAt(transactionTime).
			Exec(ctx); err != nil {
			return false, mailboxEntSafeError(err)
		}
		semanticChanged = true
	}
	return semanticChanged, nil
}

func mailboxEntApplyDeleteMessageFolderLinks(ctx context.Context, tx *ent.Tx, messageID uuid.UUID) (bool, error) {
	if ctx == nil || tx == nil || messageID == uuid.Nil {
		return false, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if err := ctx.Err(); err != nil {
		return false, mailboxEntSafeError(err)
	}
	deleted, err := tx.MailMessageFolder.Delete().Where(mailmessagefolder.MailMessageIDEQ(messageID)).Exec(ctx)
	if err != nil {
		return false, mailboxEntSafeError(err)
	}
	return deleted > 0, nil
}
