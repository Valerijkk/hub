package mail

import (
	"context"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailfolder"
	"github.com/valerii/hub/backend/ent/mailmessagefolder"
)

type mailboxEntApplyFolderState uint8

const (
	mailboxEntApplyFolderActive mailboxEntApplyFolderState = iota + 1
	mailboxEntApplyFolderTombstone
	mailboxEntApplyFolderInFlight
)

type mailboxEntApplyFolderPlan struct {
	intent         MailboxSyncPolicyFolderIntent
	normalizedName string
	role           *mailfolder.Role
	parent         *mailboxEntApplyFolderPlan
	row            *ent.MailFolder
	state          mailboxEntApplyFolderState
}

// mailboxEntApplySyncFolders persists one complete, already canonicalized provider folder snapshot.
func (r *MailboxEntRepository) mailboxEntApplySyncFolders(ctx context.Context, tx *ent.Tx, accountID uuid.UUID, syncEpoch int64, folders []MailboxSyncPolicyFolderIntent, transactionTime time.Time) error {
	if r == nil || ctx == nil || tx == nil || accountID == uuid.Nil || !mailboxEntValidTime(transactionTime) || syncEpoch < 0 || syncEpoch > 2_147_483_647 || folders == nil || len(folders) > MaxEngineFolderItems {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if err := ctx.Err(); err != nil {
		return mailboxEntSafeError(err)
	}

	engineAccountID := ""
	if len(folders) > 0 {
		engineAccountID = folders[0].Snapshot.AccountID
		if !syncValidOpaque(engineAccountID, MaxEngineOpaqueIDBytes) {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
	}

	rows, err := tx.MailFolder.Query().
		Where(mailfolder.MailAccountIDEQ(accountID)).
		Order(ent.Asc(mailfolder.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return mailboxEntSafeError(err)
	}

	plans, absent, err := mailboxEntApplyFolderPreflight(accountID, engineAccountID, folders, rows)
	if err != nil {
		return err
	}
	for _, row := range absent {
		if err := mailboxEntApplyTombstoneFolder(ctx, tx, row, syncEpoch, transactionTime); err != nil {
			return err
		}
	}

	applied := make(map[uuid.UUID]struct{}, len(plans))
	for _, plan := range plans {
		if plan.parent != nil {
			if _, found := applied[plan.parent.intent.Folder.ID]; !found {
				return mailboxServiceError(MailboxServiceErrorInvalid)
			}
		}
		if plan.row == nil {
			if err := mailboxEntApplyCreateFolder(ctx, tx, plan, syncEpoch, transactionTime); err != nil {
				return err
			}
		} else if err := mailboxEntApplyUpdateFolder(ctx, plan, syncEpoch, transactionTime); err != nil {
			return err
		}
		applied[plan.intent.Folder.ID] = struct{}{}
	}
	return nil
}

func mailboxEntApplyFolderPreflight(accountID uuid.UUID, engineAccountID string, folders []MailboxSyncPolicyFolderIntent, rows []*ent.MailFolder) ([]*mailboxEntApplyFolderPlan, []*ent.MailFolder, error) {
	byID := make(map[uuid.UUID]*mailboxEntApplyFolderPlan, len(folders))
	byEngineID := make(map[string]*mailboxEntApplyFolderPlan, len(folders))
	plans := make([]*mailboxEntApplyFolderPlan, 0, len(folders))
	for _, intent := range folders {
		if intent.Folder.AccountID != accountID || intent.Folder.ID == uuid.Nil || intent.Snapshot.AccountID != engineAccountID || !syncValidOpaque(intent.Snapshot.FolderID, MaxEngineOpaqueIDBytes) || !syncValidOpaqueOrEmpty(intent.Snapshot.ParentID, MaxEngineOpaqueIDBytes) || !mailboxEntApplyFolderNameValid(intent.Snapshot.Name) || !syncValidFolderRole(intent.Snapshot.Role) {
			return nil, nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		normalizedName := mailboxEntApplyNormalizeFolderName(intent.Snapshot.Name)
		if !mailboxEntApplyFolderNormalizedNameValid(normalizedName) {
			return nil, nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		role, validRole := mailboxEntApplyFolderRole(intent.Snapshot.Role)
		if !validRole || (role != nil && intent.Snapshot.ParentID != "") {
			return nil, nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if _, duplicate := byID[intent.Folder.ID]; duplicate {
			return nil, nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if _, duplicate := byEngineID[intent.Snapshot.FolderID]; duplicate {
			return nil, nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		plan := &mailboxEntApplyFolderPlan{intent: intent, normalizedName: normalizedName, role: role}
		byID[intent.Folder.ID] = plan
		byEngineID[intent.Snapshot.FolderID] = plan
		plans = append(plans, plan)
	}

	rowsByID := make(map[uuid.UUID]*ent.MailFolder, len(rows))
	rowsByEngineID := make(map[string]*ent.MailFolder, len(rows))
	rowStates := make(map[uuid.UUID]mailboxEntApplyFolderState, len(rows))
	for _, row := range rows {
		state, valid := mailboxEntApplyFolderRowState(row, accountID)
		if !valid {
			return nil, nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if _, duplicate := rowsByID[row.ID]; duplicate {
			return nil, nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		rowsByID[row.ID] = row
		rowStates[row.ID] = state
		if row.EngineFolderID != nil {
			if _, duplicate := rowsByEngineID[*row.EngineFolderID]; duplicate {
				return nil, nil, mailboxServiceError(MailboxServiceErrorInvalid)
			}
			rowsByEngineID[*row.EngineFolderID] = row
		}
	}
	for _, row := range rows {
		if row.ParentID == nil {
			continue
		}
		parent, found := rowsByID[*row.ParentID]
		if !found || parent == nil || parent.MailAccountID != accountID || parent.ID == row.ID {
			return nil, nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
	}

	for _, plan := range plans {
		if plan.intent.Snapshot.ParentID != "" {
			parent, found := byEngineID[plan.intent.Snapshot.ParentID]
			if !found || parent == plan || parent.role != nil {
				return nil, nil, mailboxServiceError(MailboxServiceErrorInvalid)
			}
			plan.parent = parent
		}
		byHubID, hasHubID := rowsByID[plan.intent.Folder.ID]
		byEngine, hasEngineID := rowsByEngineID[plan.intent.Snapshot.FolderID]
		if hasHubID && hasEngineID && byHubID.ID != byEngine.ID {
			return nil, nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if hasHubID {
			if byHubID.EngineFolderID == nil || *byHubID.EngineFolderID != plan.intent.Snapshot.FolderID || !mailboxEntApplyFolderRoleEqual(byHubID.Role, plan.role) {
				return nil, nil, mailboxServiceError(MailboxServiceErrorInvalid)
			}
			plan.row = byHubID
			plan.state = rowStates[byHubID.ID]
			if plan.state != mailboxEntApplyFolderActive && plan.state != mailboxEntApplyFolderTombstone {
				return nil, nil, mailboxServiceError(MailboxServiceErrorInvalid)
			}
		} else if hasEngineID {
			return nil, nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
	}

	if err := mailboxEntApplyFolderSnapshotGraphValid(plans); err != nil {
		return nil, nil, err
	}
	if err := mailboxEntApplyFolderCollisionFree(plans, rows, rowStates); err != nil {
		return nil, nil, err
	}

	absent := make([]*ent.MailFolder, 0, len(rows))
	for _, row := range rows {
		if rowStates[row.ID] != mailboxEntApplyFolderActive {
			continue
		}
		if _, present := byEngineID[*row.EngineFolderID]; !present {
			absent = append(absent, row)
		}
	}
	sort.Slice(absent, func(left, right int) bool { return syncLessUUID(absent[left].ID, absent[right].ID) })

	ordered, err := mailboxEntApplyFolderOrder(plans)
	if err != nil {
		return nil, nil, err
	}
	return ordered, absent, nil
}

func mailboxEntApplyFolderSnapshotGraphValid(plans []*mailboxEntApplyFolderPlan) error {
	roles := make(map[mailfolder.Role]struct{}, len(plans))
	coordinates := make(map[string]uuid.UUID, len(plans))
	for _, plan := range plans {
		if plan.role != nil {
			if _, duplicate := roles[*plan.role]; duplicate {
				return mailboxServiceError(MailboxServiceErrorInvalid)
			}
			roles[*plan.role] = struct{}{}
			continue
		}
		parentKey := ""
		if plan.parent != nil {
			parentKey = plan.parent.intent.Folder.ID.String()
		}
		coordinate := parentKey + "\x00" + plan.normalizedName
		if _, duplicate := coordinates[coordinate]; duplicate {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		coordinates[coordinate] = plan.intent.Folder.ID
	}
	_, err := mailboxEntApplyFolderOrder(plans)
	return err
}

func mailboxEntApplyFolderCollisionFree(plans []*mailboxEntApplyFolderPlan, rows []*ent.MailFolder, states map[uuid.UUID]mailboxEntApplyFolderState) error {
	plansByID := make(map[uuid.UUID]*mailboxEntApplyFolderPlan, len(plans))
	for _, plan := range plans {
		plansByID[plan.intent.Folder.ID] = plan
	}
	for _, plan := range plans {
		for _, row := range rows {
			if row.ID == plan.intent.Folder.ID || row.DeletedAt != nil {
				continue
			}
			if plan.role != nil && row.Role != nil && *plan.role == *row.Role {
				return mailboxServiceError(MailboxServiceErrorInvalid)
			}
			if plan.role != nil || row.Role != nil || row.NormalizedName != plan.normalizedName || !mailboxEntApplyFolderRowHasParent(row, plan.parent) {
				continue
			}
			if _, present := plansByID[row.ID]; present || states[row.ID] != mailboxEntApplyFolderActive {
				return mailboxServiceError(MailboxServiceErrorInvalid)
			}
		}
	}
	return nil
}

func mailboxEntApplyFolderOrder(plans []*mailboxEntApplyFolderPlan) ([]*mailboxEntApplyFolderPlan, error) {
	orderedInput := append([]*mailboxEntApplyFolderPlan(nil), plans...)
	sort.Slice(orderedInput, func(left, right int) bool {
		if orderedInput[left].intent.Folder.ID != orderedInput[right].intent.Folder.ID {
			return syncLessUUID(orderedInput[left].intent.Folder.ID, orderedInput[right].intent.Folder.ID)
		}
		return orderedInput[left].intent.Snapshot.FolderID < orderedInput[right].intent.Snapshot.FolderID
	})
	marks := make(map[uuid.UUID]uint8, len(orderedInput))
	ordered := make([]*mailboxEntApplyFolderPlan, 0, len(orderedInput))
	var visit func(*mailboxEntApplyFolderPlan) error
	visit = func(plan *mailboxEntApplyFolderPlan) error {
		switch marks[plan.intent.Folder.ID] {
		case 1:
			return mailboxServiceError(MailboxServiceErrorInvalid)
		case 2:
			return nil
		}
		marks[plan.intent.Folder.ID] = 1
		if plan.parent != nil {
			if err := visit(plan.parent); err != nil {
				return err
			}
		}
		marks[plan.intent.Folder.ID] = 2
		ordered = append(ordered, plan)
		return nil
	}
	for _, plan := range orderedInput {
		if err := visit(plan); err != nil {
			return nil, err
		}
	}
	return ordered, nil
}

func mailboxEntApplyTombstoneFolder(ctx context.Context, tx *ent.Tx, row *ent.MailFolder, syncEpoch int64, transactionTime time.Time) error {
	if row == nil || row.ID == uuid.Nil || tx == nil {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if _, err := tx.MailMessageFolder.Delete().
		Where(mailmessagefolder.MailFolderIDEQ(row.ID)).
		Exec(ctx); err != nil {
		return mailboxEntSafeError(err)
	}
	if err := row.Update().
		SetControlState(mailfolder.ControlStateDeleting).
		SetDeletedAt(transactionTime).
		SetUnreadCount(0).
		SetTotalCount(0).
		SetLastSeenSyncEpoch(syncEpoch).
		SetUpdatedAt(transactionTime).
		AddVersion(1).
		Exec(ctx); err != nil {
		return mailboxEntSafeError(err)
	}
	return nil
}

func mailboxEntApplyCreateFolder(ctx context.Context, tx *ent.Tx, plan *mailboxEntApplyFolderPlan, syncEpoch int64, transactionTime time.Time) error {
	if tx == nil || plan == nil || plan.row != nil {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	create := tx.MailFolder.Create().
		SetID(plan.intent.Folder.ID).
		SetMailAccountID(plan.intent.Folder.AccountID).
		SetEngineFolderID(plan.intent.Snapshot.FolderID).
		SetName(plan.intent.Snapshot.Name).
		SetNormalizedName(plan.normalizedName).
		SetControlState(mailfolder.ControlStateActive).
		SetUnreadCount(0).
		SetTotalCount(0).
		SetLastSeenSyncEpoch(syncEpoch).
		SetCreatedAt(transactionTime).
		SetUpdatedAt(transactionTime).
		SetVersion(1)
	if plan.role != nil {
		create.SetRole(*plan.role)
	}
	if plan.parent != nil {
		create.SetParentID(plan.parent.intent.Folder.ID)
	}
	created, err := create.Save(ctx)
	if err != nil {
		return mailboxEntSafeError(err)
	}
	if created == nil || created.ID != plan.intent.Folder.ID || created.MailAccountID != plan.intent.Folder.AccountID {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	return nil
}

func mailboxEntApplyUpdateFolder(ctx context.Context, plan *mailboxEntApplyFolderPlan, syncEpoch int64, transactionTime time.Time) error {
	if plan == nil || plan.row == nil {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	row := plan.row
	if plan.state == mailboxEntApplyFolderTombstone {
		update := row.Update().
			SetControlState(mailfolder.ControlStateActive).
			ClearDeletedAt().
			SetName(plan.intent.Snapshot.Name).
			SetNormalizedName(plan.normalizedName).
			SetLastSeenSyncEpoch(syncEpoch).
			SetUpdatedAt(transactionTime).
			AddVersion(1)
		mailboxEntApplyFolderSetParent(update, plan.parent)
		if err := update.Exec(ctx); err != nil {
			return mailboxEntSafeError(err)
		}
		return nil
	}
	if plan.state != mailboxEntApplyFolderActive {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	semanticChange := row.Name != plan.intent.Snapshot.Name || row.NormalizedName != plan.normalizedName || !mailboxEntApplyFolderRowHasParent(row, plan.parent)
	if !semanticChange && row.LastSeenSyncEpoch == syncEpoch {
		return nil
	}
	update := row.Update().SetLastSeenSyncEpoch(syncEpoch)
	if semanticChange {
		update.SetName(plan.intent.Snapshot.Name).
			SetNormalizedName(plan.normalizedName).
			SetUpdatedAt(transactionTime).
			AddVersion(1)
		mailboxEntApplyFolderSetParent(update, plan.parent)
	} else {
		update.SetUpdatedAt(row.UpdatedAt)
	}
	if err := update.Exec(ctx); err != nil {
		return mailboxEntSafeError(err)
	}
	return nil
}

func mailboxEntApplyFolderSetParent(update *ent.MailFolderUpdateOne, parent *mailboxEntApplyFolderPlan) {
	if parent == nil {
		update.ClearParentID()
		return
	}
	update.SetParentID(parent.intent.Folder.ID)
}

func mailboxEntApplyFolderRowState(row *ent.MailFolder, accountID uuid.UUID) (mailboxEntApplyFolderState, bool) {
	if row == nil || row.ID == uuid.Nil || row.MailAccountID != accountID || row.Version < 1 || row.LastSeenSyncEpoch < 0 || row.LastSeenSyncEpoch > 2_147_483_647 || row.UnreadCount < 0 || row.TotalCount < 0 || row.UnreadCount > row.TotalCount || !mailboxEntApplyFolderNameValid(row.Name) || !mailboxEntApplyFolderNormalizedNameValid(row.NormalizedName) {
		return 0, false
	}
	if row.EngineFolderID != nil && !syncValidOpaque(*row.EngineFolderID, MaxEngineOpaqueIDBytes) {
		return 0, false
	}
	if row.ParentID != nil && *row.ParentID == row.ID {
		return 0, false
	}
	if row.Role != nil {
		role := EngineFolderRole(*row.Role)
		if role == EngineFolderRoleCustom || !syncValidFolderRole(role) || row.ParentID != nil {
			return 0, false
		}
	}
	if row.ControlReceiptID != nil && *row.ControlReceiptID == uuid.Nil {
		return 0, false
	}
	switch row.ControlState {
	case mailfolder.ControlStateActive:
		return mailboxEntApplyFolderActive, row.EngineFolderID != nil && row.ControlReceiptID == nil && row.DeletedAt == nil
	case mailfolder.ControlStateCreating:
		return mailboxEntApplyFolderInFlight, row.EngineFolderID == nil && row.ControlReceiptID != nil && row.DeletedAt == nil && row.UnreadCount == 0 && row.TotalCount == 0
	case mailfolder.ControlStateUnknown, mailfolder.ControlStateError:
		return mailboxEntApplyFolderInFlight, row.ControlReceiptID != nil && row.DeletedAt == nil && row.UnreadCount == 0 && row.TotalCount == 0
	case mailfolder.ControlStateDeleting:
		if row.EngineFolderID == nil || row.UnreadCount != 0 || row.TotalCount != 0 {
			return 0, false
		}
		if row.DeletedAt != nil && row.ControlReceiptID == nil {
			return mailboxEntApplyFolderTombstone, true
		}
		return mailboxEntApplyFolderInFlight, row.DeletedAt == nil && row.ControlReceiptID != nil
	default:
		return 0, false
	}
}

func mailboxEntApplyFolderRole(role EngineFolderRole) (*mailfolder.Role, bool) {
	switch role {
	case EngineFolderRoleCustom:
		return nil, true
	case EngineFolderRoleInbox, EngineFolderRoleSent, EngineFolderRoleDrafts, EngineFolderRoleArchive, EngineFolderRoleTrash, EngineFolderRoleJunk:
		value := mailfolder.Role(role)
		return &value, true
	default:
		return nil, false
	}
}

func mailboxEntApplyFolderRoleEqual(left, right *mailfolder.Role) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func mailboxEntApplyFolderRowHasParent(row *ent.MailFolder, parent *mailboxEntApplyFolderPlan) bool {
	if row == nil {
		return false
	}
	if parent == nil {
		return row.ParentID == nil
	}
	return row.ParentID != nil && *row.ParentID == parent.intent.Folder.ID
}

func mailboxEntApplyFolderNameValid(value string) bool {
	if !utf8.ValidString(value) || len(value) < 1 || len(value) > 480 || utf8.RuneCountInString(value) > 120 {
		return false
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return false
		}
	}
	return true
}

func mailboxEntApplyFolderNormalizedNameValid(value string) bool {
	return mailboxEntApplyFolderNameValid(value) && mailboxEntApplyNormalizeFolderName(value) == value
}

func mailboxEntApplyNormalizeFolderName(value string) string {
	value = strings.Join(strings.FieldsFunc(value, unicode.IsSpace), " ")
	value = norm.NFC.String(value)
	value = cases.Fold().String(value)
	return norm.NFC.String(value)
}
