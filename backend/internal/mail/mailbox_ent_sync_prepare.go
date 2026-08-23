package mail

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/maildraft"
	"github.com/valerii/hub/backend/ent/mailfolder"
	"github.com/valerii/hub/backend/ent/maillabel"
	"github.com/valerii/hub/backend/ent/mailmessage"
	"github.com/valerii/hub/backend/ent/mailmessagefolder"
	"github.com/valerii/hub/backend/ent/mailmessagelabel"
	"github.com/valerii/hub/backend/ent/mailsubmission"
	"github.com/valerii/hub/backend/ent/mailthread"
	"github.com/valerii/hub/backend/ent/outboxevent"
	"github.com/valerii/hub/backend/ent/predicate"
)

const mailboxEntSyncIDDomain = "hub.mail.sync-id.v1"

// PrepareMailboxSyncPage re-resolves one claimed page into deterministic, policy-ready state.
func (repository *MailboxEntRepository) PrepareMailboxSyncPage(ctx context.Context, request MailboxSyncPrepareRequest) (MailboxSyncPreparedPage, error) {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxEntPrepareRequestValid(request) {
		return MailboxSyncPreparedPage{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	var result MailboxSyncPreparedPage
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, err := transaction.MailAccount.Query().
			Where(mailaccount.IDEQ(request.Fence.AccountID)).
			ForUpdate().
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil
			}
			return mailboxEntSafeError(err)
		}
		databaseWallClock, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return err
		}
		input, prepared, err := mailboxEntPrepareInput(ctx, transaction, account, request, databaseWallClock)
		if err != nil {
			return err
		}
		if !prepared {
			return nil
		}
		if !validMailboxSyncPrepareSelection(input, request.EventCandidates) {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if _, err := ClassifyMailboxSyncPage(input); err != nil {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		result = MailboxSyncPreparedPage{Prepared: true, Input: mailboxEntPrepareCopyInput(input)}
		return nil
	})
	if err != nil {
		return MailboxSyncPreparedPage{}, mailboxEntSafeError(err)
	}
	if !result.Prepared {
		return MailboxSyncPreparedPage{}, nil
	}
	result.Input = mailboxEntPrepareCopyInput(result.Input)
	return result, nil
}

func mailboxEntPrepareRequestValid(request MailboxSyncPrepareRequest) bool {
	if !validMailboxSyncFence(request.Fence) || !syncValidOpaque(request.EngineAccountID, MaxEngineOpaqueIDBytes) || !mailboxEntValidTime(request.OccurredAt) || !syncValidCorrelationID(request.CorrelationID) || request.CorrelationID != request.Fence.LeaseID.String() || request.Folders == nil || request.EventCandidates == nil || len(request.Folders) > MaxEngineFolderItems {
		return false
	}
	page, err := syncValidatePage(request.Page)
	if err != nil || page.accountID != request.EngineAccountID || !mailboxEntPreparePageMatchesFence(page, request.Fence) {
		return false
	}
	seenFolders := make(map[string]struct{}, len(request.Folders))
	for _, folder := range request.Folders {
		if folder.AccountID != request.EngineAccountID || !syncValidOpaque(folder.FolderID, MaxEngineOpaqueIDBytes) || !syncValidOpaqueOrEmpty(folder.ParentID, MaxEngineOpaqueIDBytes) || !syncValidSafeText(folder.Name, MaxEngineHeaderValueBytes, false) || !syncValidFolderRole(folder.Role) {
			return false
		}
		if _, duplicate := seenFolders[folder.FolderID]; duplicate {
			return false
		}
		seenFolders[folder.FolderID] = struct{}{}
	}
	return mailboxEntPrepareCandidatesValid(request.Page, request.EventCandidates)
}

func mailboxEntPreparePageMatchesFence(page syncValidatedPage, fence MailboxSyncFence) bool {
	switch fence.Phase {
	case MailboxSyncPhaseIncremental, MailboxSyncPhaseCatchUp:
		return page.change != nil && page.enumeration == nil && page.change.FromCursor == fence.Cursor
	case MailboxSyncPhaseEnumeration:
		if page.change != nil || page.enumeration == nil {
			return false
		}
		if fence.BaseCursor == "" && fence.Checkpoint == "" {
			return syncValidCursor(page.enumeration.BaseCursor, false) && page.enumeration.Checkpoint == ""
		}
		return syncValidCursor(fence.BaseCursor, false) &&
			syncValidCursor(fence.Checkpoint, false) &&
			page.enumeration.BaseCursor == fence.BaseCursor &&
			page.enumeration.Checkpoint == fence.Checkpoint
	default:
		return false
	}
}

func mailboxEntPrepareCandidatesValid(page MailboxSyncPolicyPage, candidates []MailboxSyncEventCandidate) bool {
	live, destroyed, ok := mailboxSyncPageCandidateShapes(page)
	if !ok || len(candidates) != len(live)+len(destroyed) {
		return false
	}
	seenMessages := make(map[string]struct{}, len(candidates))
	seenEvents := make(map[uuid.UUID]struct{}, 2*len(candidates))
	for _, candidate := range candidates {
		if !syncValidOpaque(candidate.EngineMessageID, MaxEngineOpaqueIDBytes) || candidate.ChangedEventID == uuid.Nil || (!candidate.Destroyed && candidate.ReceivedEventID == uuid.Nil) || (candidate.Destroyed && candidate.ReceivedEventID != uuid.Nil) {
			return false
		}
		if _, duplicate := seenMessages[candidate.EngineMessageID]; duplicate {
			return false
		}
		seenMessages[candidate.EngineMessageID] = struct{}{}
		if _, duplicate := seenEvents[candidate.ChangedEventID]; duplicate {
			return false
		}
		seenEvents[candidate.ChangedEventID] = struct{}{}
		if !candidate.Destroyed {
			if _, duplicate := seenEvents[candidate.ReceivedEventID]; duplicate {
				return false
			}
			seenEvents[candidate.ReceivedEventID] = struct{}{}
		}
		if candidate.Destroyed != destroyed[candidate.EngineMessageID] || (!candidate.Destroyed && !live[candidate.EngineMessageID]) {
			return false
		}
	}
	return len(seenMessages) == len(candidates)
}

func mailboxEntPrepareInput(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, request MailboxSyncPrepareRequest, now time.Time) (MailboxSyncPolicyInput, bool, error) {
	if transaction == nil || account == nil || !mailboxEntValidTime(now) {
		return MailboxSyncPolicyInput{}, false, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if account.Status != mailaccount.StatusActive || !account.DesiredEnabled {
		return MailboxSyncPolicyInput{}, false, nil
	}
	if account.ID != request.Fence.AccountID || account.OrganizationID == uuid.Nil || account.UserID == uuid.Nil || account.MailDomainID == uuid.Nil || account.EngineAccountID == nil || *account.EngineAccountID != request.EngineAccountID || !syncValidOpaque(*account.EngineAccountID, MaxEngineOpaqueIDBytes) || !mailboxEntValidTime(account.EventCutoffAt) {
		return MailboxSyncPolicyInput{}, false, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	persistedFence, hasLease, leaseValid := mailboxEntSyncLeaseFence(account)
	if !leaseValid {
		return MailboxSyncPolicyInput{}, false, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if !hasLease || persistedFence != request.Fence || account.SyncEpoch != request.Fence.SyncEpoch || account.MutationEpoch != request.Fence.MutationEpoch || account.SyncLeaseUntil == nil || !account.SyncLeaseUntil.After(now) {
		return MailboxSyncPolicyInput{}, false, nil
	}

	page := cloneMailboxSyncPolicyPage(request.Page)
	folders, err := mailboxEntPrepareFolders(ctx, transaction, account, request.Folders)
	if err != nil {
		return MailboxSyncPolicyInput{}, false, err
	}
	projections := mailboxEntPrepareProjectionPointers(&page)
	engineMessageIDs := mailboxEntPrepareProjectionIDs(projections, page)
	messageRows, err := mailboxEntPrepareMessageRows(ctx, transaction, account, engineMessageIDs)
	if err != nil {
		return MailboxSyncPolicyInput{}, false, err
	}
	liveEngineMessageIDs := make([]string, 0, len(projections))
	for _, projection := range projections {
		liveEngineMessageIDs = append(liveEngineMessageIDs, projection.MessageID)
	}
	sort.Strings(liveEngineMessageIDs)

	provenance, err := mailboxEntPrepareProvenance(ctx, transaction, account, liveEngineMessageIDs)
	if err != nil {
		return MailboxSyncPolicyInput{}, false, err
	}
	for _, projection := range projections {
		proof := provenance[projection.MessageID]
		if proof.ConfirmedSubmission != nil && proof.ConfirmedDraft != nil {
			return MailboxSyncPolicyInput{}, false, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		switch {
		case proof.ConfirmedSubmission != nil:
			projection.Origin = EngineProjectionOriginConfirmedOwnerSubmission
		case proof.ConfirmedDraft != nil:
			projection.Origin = EngineProjectionOriginConfirmedOwnerDraft
		default:
			projection.Origin = EngineProjectionOriginInboundOrUnknown
		}
	}
	labels, err := mailboxEntPrepareLabels(ctx, transaction, account, projections)
	if err != nil {
		return MailboxSyncPolicyInput{}, false, err
	}
	stateRows := make(map[string]*ent.MailMessage, len(messageRows))
	for _, projection := range projections {
		if row, ok := messageRows[projection.MessageID]; ok {
			stateRows[projection.MessageID] = row
		}
	}
	if page.Change != nil {
		for _, engineMessageID := range page.Change.DestroyedIDs {
			if row, ok := messageRows[engineMessageID]; ok && row.DeletedAt == nil {
				stateRows[engineMessageID] = row
			}
		}
	}
	states, err := mailboxEntPreparePersistedStates(ctx, transaction, account, stateRows)
	if err != nil {
		return MailboxSyncPolicyInput{}, false, err
	}
	threads, err := mailboxEntPrepareThreads(ctx, transaction, account, projections)
	if err != nil {
		return MailboxSyncPolicyInput{}, false, err
	}
	messages, destroyed, err := mailboxEntPrepareMessages(ctx, transaction, account, page, messageRows, states, threads, provenance)
	if err != nil {
		return MailboxSyncPolicyInput{}, false, err
	}
	input := MailboxSyncPolicyInput{
		Page: page, Folders: folders, Messages: messages, Destroyed: destroyed, Labels: labels,
		SyncEpoch: account.SyncEpoch, EventCutoff: account.EventCutoffAt,
		Event: MailboxSyncPolicyEventContext{Mailbox: MailboxEventID{AccountID: account.ID, ID: account.ID}, OrganizationID: account.OrganizationID, OccurredAt: request.OccurredAt, CorrelationID: request.CorrelationID},
	}
	if err := mailboxEntPrepareCandidateAvailability(ctx, transaction, request.EventCandidates); err != nil {
		return MailboxSyncPolicyInput{}, false, err
	}
	if err := mailboxEntPrepareSelectEvents(&input, request.EventCandidates); err != nil {
		return MailboxSyncPolicyInput{}, false, err
	}
	return mailboxEntPrepareCopyInput(input), true, nil
}

func mailboxEntStableSyncID(accountID uuid.UUID, kind string, identity []byte) (uuid.UUID, error) {
	if accountID == uuid.Nil || (kind != "folder" && kind != "message" && kind != "thread") || len(identity) == 0 {
		return uuid.Nil, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	hasher := sha256.New()
	writeComponent := func(component []byte) {
		var length [4]byte
		binary.BigEndian.PutUint32(length[:], uint32(len(component)))
		_, _ = hasher.Write(length[:])
		_, _ = hasher.Write(component)
	}
	writeComponent([]byte(mailboxEntSyncIDDomain))
	writeComponent([]byte(kind))
	writeComponent(accountID[:])
	writeComponent(identity)
	digest := hasher.Sum(nil)
	var value [16]byte
	copy(value[:], digest[:16])
	value[6] = (value[6] & 0x0f) | 0x80
	value[8] = (value[8] & 0x3f) | 0x80
	result, err := uuid.FromBytes(value[:])
	if err != nil || result == uuid.Nil {
		return uuid.Nil, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	return result, nil
}

func mailboxEntPrepareFolders(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, snapshots []EngineFolder) ([]MailboxSyncPolicyFolder, error) {
	if len(snapshots) == 0 {
		return make([]MailboxSyncPolicyFolder, 0), nil
	}
	engineFolderIDs := make([]string, len(snapshots))
	for index, snapshot := range snapshots { engineFolderIDs[index] = snapshot.FolderID }
	sort.Strings(engineFolderIDs)
	rows, err := transaction.MailFolder.Query().Where(mailfolder.MailAccountIDEQ(account.ID), mailfolder.EngineFolderIDIn(engineFolderIDs...)).All(ctx)
	if err != nil { return nil, mailboxEntSafeError(err) }
	existing := make(map[string]*ent.MailFolder, len(rows))
	for _, row := range rows {
		if row == nil || row.ID == uuid.Nil || row.MailAccountID != account.ID || row.EngineFolderID == nil || !syncValidOpaque(*row.EngineFolderID, MaxEngineOpaqueIDBytes) { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
		if _, duplicate := existing[*row.EngineFolderID]; duplicate { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
		existing[*row.EngineFolderID] = row
	}
	proposed := make(map[uuid.UUID]string, len(snapshots))
	result := make([]MailboxSyncPolicyFolder, 0, len(snapshots))
	for _, snapshot := range snapshots {
		id := uuid.Nil
		if row, found := existing[snapshot.FolderID]; found { id = row.ID } else {
			id, err = mailboxEntStableSyncID(account.ID, "folder", []byte(snapshot.FolderID))
			if err != nil { return nil, err }
			if other, duplicate := proposed[id]; duplicate && other != snapshot.FolderID { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
			proposed[id] = snapshot.FolderID
		}
		result = append(result, MailboxSyncPolicyFolder{Snapshot: snapshot, Folder: MailboxEventID{AccountID: account.ID, ID: id}})
	}
	if err := mailboxEntPrepareFolderCollisions(ctx, transaction, account, proposed); err != nil { return nil, err }
	if err := mailboxEntPrepareStableIDsUnclaimed(ctx, transaction, "folder", mailboxEntPrepareSortedUUIDStrings(proposed)); err != nil { return nil, err }
	sort.Slice(result, func(left, right int) bool { return result[left].Snapshot.FolderID < result[right].Snapshot.FolderID })
	return result, nil
}

func mailboxEntPrepareFolderCollisions(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, proposed map[uuid.UUID]string) error {
	if len(proposed) == 0 { return nil }
	ids := mailboxEntPrepareSortedUUIDStrings(proposed)
	rows, err := transaction.MailFolder.Query().Where(mailfolder.IDIn(ids...)).All(ctx)
	if err != nil { return mailboxEntSafeError(err) }
	for _, row := range rows {
		want, found := proposed[row.ID]
		if !found || row.MailAccountID != account.ID || row.EngineFolderID == nil || *row.EngineFolderID != want { return mailboxServiceError(MailboxServiceErrorInvalid) }
	}
	return nil
}

func mailboxEntPrepareProjectionPointers(page *MailboxSyncPolicyPage) []*EngineMessageProjection {
	if page == nil { return nil }
	if page.Change != nil && page.Enumeration == nil {
		result := make([]*EngineMessageProjection, len(page.Change.CreatedOrUpdated))
		for index := range page.Change.CreatedOrUpdated { result[index] = &page.Change.CreatedOrUpdated[index] }
		return result
	}
	if page.Enumeration != nil && page.Change == nil {
		result := make([]*EngineMessageProjection, len(page.Enumeration.Messages))
		for index := range page.Enumeration.Messages { result[index] = &page.Enumeration.Messages[index] }
		return result
	}
	return nil
}

func mailboxEntPrepareProjectionIDs(projections []*EngineMessageProjection, page MailboxSyncPolicyPage) []string {
	result := make([]string, 0, len(projections))
	for _, projection := range projections { result = append(result, projection.MessageID) }
	if page.Change != nil { result = append(result, page.Change.DestroyedIDs...) }
	sort.Strings(result)
	return result
}

func mailboxEntPrepareMessageRows(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, engineMessageIDs []string) (map[string]*ent.MailMessage, error) {
	if len(engineMessageIDs) == 0 { return map[string]*ent.MailMessage{}, nil }
	rows, err := transaction.MailMessage.Query().Where(mailmessage.MailAccountIDEQ(account.ID), mailmessage.EngineMessageIDIn(engineMessageIDs...)).All(ctx)
	if err != nil { return nil, mailboxEntSafeError(err) }
	result := make(map[string]*ent.MailMessage, len(rows))
	for _, row := range rows {
		if row == nil || row.ID == uuid.Nil || row.MailAccountID != account.ID || !syncValidOpaque(row.EngineMessageID, MaxEngineOpaqueIDBytes) { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
		if _, duplicate := result[row.EngineMessageID]; duplicate { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
		result[row.EngineMessageID] = row
	}
	return result, nil
}

func mailboxEntPrepareProvenance(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, engineMessageIDs []string) (map[string]MailboxSyncPolicyProvenance, error) {
	result := make(map[string]MailboxSyncPolicyProvenance, len(engineMessageIDs))
	if len(engineMessageIDs) == 0 { return result, nil }
	submissions, err := transaction.MailSubmission.Query().Where(mailsubmission.MailAccountIDEQ(account.ID), mailsubmission.EngineResultMessageIDIn(engineMessageIDs...)).All(ctx)
	if err != nil { return nil, mailboxEntSafeError(err) }
	for _, row := range submissions {
		if row == nil || row.ID == uuid.Nil || row.MailAccountID != account.ID || string(row.Status) != "sent" || row.EngineSubmissionID == nil || row.EngineResultMessageID == nil || row.ResultMailMessageID == nil || !syncValidOpaque(*row.EngineSubmissionID, MaxEngineOpaqueIDBytes) || !syncValidOpaque(*row.EngineResultMessageID, MaxEngineOpaqueIDBytes) {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		current := result[*row.EngineResultMessageID]
		if current.ConfirmedSubmission != nil { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
		current.ConfirmedSubmission = &MailboxSyncPolicyProvenanceProof{Relation: MailboxEventID{AccountID: account.ID, ID: row.ID}, EngineMessageID: *row.EngineResultMessageID}
		if current.ConfirmedDraft != nil { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
		result[*row.EngineResultMessageID] = current
	}
	drafts, err := transaction.MailDraft.Query().Where(maildraft.MailAccountIDEQ(account.ID), maildraft.EngineMessageIDIn(engineMessageIDs...)).All(ctx)
	if err != nil { return nil, mailboxEntSafeError(err) }
	for _, row := range drafts {
		if row == nil || row.ID == uuid.Nil || row.MailAccountID != account.ID || (string(row.Status) != "active" && string(row.Status) != "submitting" && string(row.Status) != "submitted") || row.EngineMessageID == nil || row.EngineVersion == nil || !syncValidOpaque(*row.EngineMessageID, MaxEngineOpaqueIDBytes) || !syncValidOpaque(*row.EngineVersion, MaxEngineOpaqueIDBytes) {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		current := result[*row.EngineMessageID]
		if current.ConfirmedDraft != nil || current.ConfirmedSubmission != nil { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
		current.ConfirmedDraft = &MailboxSyncPolicyProvenanceProof{Relation: MailboxEventID{AccountID: account.ID, ID: row.ID}, EngineMessageID: *row.EngineMessageID}
		result[*row.EngineMessageID] = current
	}
	return result, nil
}

func mailboxEntPrepareLabels(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, projections []*EngineMessageProjection) ([]MailboxSyncPolicyLabelDefinition, error) {
	keywordSet := make(map[EngineKeywordMarker]struct{})
	for _, projection := range projections {
		projectionKeywords := make(map[EngineKeywordMarker]struct{}, len(projection.ReservedKeywords))
		for _, keyword := range projection.ReservedKeywords {
			if !syncValidOpaque(string(keyword), MaxEngineOpaqueIDBytes) { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
			if _, duplicate := projectionKeywords[keyword]; duplicate { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
			projectionKeywords[keyword] = struct{}{}
			keywordSet[keyword] = struct{}{}
		}
	}
	keywords := make([]string, 0, len(keywordSet))
	for keyword := range keywordSet { keywords = append(keywords, string(keyword)) }
	sort.Strings(keywords)
	if len(keywords) == 0 { return make([]MailboxSyncPolicyLabelDefinition, 0), nil }
	rows, err := transaction.MailLabel.Query().Where(maillabel.MailAccountIDEQ(account.ID), maillabel.EngineKeywordIn(keywords...), maillabel.DeletedAtIsNil()).All(ctx)
	if err != nil { return nil, mailboxEntSafeError(err) }
	result := make([]MailboxSyncPolicyLabelDefinition, 0, len(rows))
	seen := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		if row == nil || row.ID == uuid.Nil || row.MailAccountID != account.ID || row.DeletedAt != nil || !syncValidOpaque(row.EngineKeyword, MaxEngineOpaqueIDBytes) { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
		if _, duplicate := seen[row.EngineKeyword]; duplicate { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
		seen[row.EngineKeyword] = struct{}{}
		result = append(result, MailboxSyncPolicyLabelDefinition{Label: MailboxEventID{AccountID: account.ID, ID: row.ID}, EngineKeyword: EngineKeywordMarker(row.EngineKeyword)})
	}
	sort.Slice(result, func(left, right int) bool { return string(result[left].EngineKeyword) < string(result[right].EngineKeyword) })
	return result, nil
}

type mailboxEntPreparePersistedState struct {
	previous       *MailboxSyncPolicyPreviousMessage
	calendarMethod *MailCalendarMethod
}

func mailboxEntPreparePersistedStates(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, rows map[string]*ent.MailMessage) (map[uuid.UUID]mailboxEntPreparePersistedState, error) {
	existing := make(map[uuid.UUID]*ent.MailMessage, len(rows))
	threadSet := make(map[uuid.UUID]struct{}, len(rows))
	for _, row := range rows {
		if row == nil || row.ID == uuid.Nil || row.MailAccountID != account.ID || !syncValidOpaque(row.EngineMessageID, MaxEngineOpaqueIDBytes) || row.MailThreadID == uuid.Nil { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
		if _, duplicate := existing[row.ID]; duplicate { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
		existing[row.ID] = row
		threadSet[row.MailThreadID] = struct{}{}
	}
	result := make(map[uuid.UUID]mailboxEntPreparePersistedState, len(existing))
	if len(existing) == 0 { return result, nil }
	threadIDs := mailboxEntPrepareSortedUUIDSet(threadSet)
	threadRows, err := transaction.MailThread.Query().Where(mailthread.MailAccountIDEQ(account.ID), mailthread.IDIn(threadIDs...)).All(ctx)
	if err != nil { return nil, mailboxEntSafeError(err) }
	threads := make(map[uuid.UUID]*ent.MailThread, len(threadRows))
	for _, row := range threadRows {
		if _, ok := mailboxEntPrepareThreadKeyFromRow(row); !ok || row.ID == uuid.Nil || row.MailAccountID != account.ID { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
		if _, duplicate := threads[row.ID]; duplicate { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
		threads[row.ID] = row
	}
	if len(threads) != len(threadSet) { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }

	messageIDs := make([]uuid.UUID, 0, len(existing))
	for messageID := range existing { messageIDs = append(messageIDs, messageID) }
	sort.Slice(messageIDs, func(left, right int) bool { return syncLessUUID(messageIDs[left], messageIDs[right]) })
	folderLinks, err := transaction.MailMessageFolder.Query().Where(mailmessagefolder.MailMessageIDIn(messageIDs...)).All(ctx)
	if err != nil { return nil, mailboxEntSafeError(err) }
	folderIDs := make(map[uuid.UUID]struct{}, len(folderLinks))
	foldersByMessage := make(map[uuid.UUID][]*ent.MailMessageFolder, len(existing))
	for _, link := range folderLinks {
		if link == nil || link.ID == uuid.Nil || link.MailMessageID == uuid.Nil || link.MailFolderID == uuid.Nil { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
		message, found := existing[link.MailMessageID]
		if !found { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
		if message.DeletedAt != nil { continue }
		foldersByMessage[link.MailMessageID] = append(foldersByMessage[link.MailMessageID], link)
		folderIDs[link.MailFolderID] = struct{}{}
	}
	folders := make(map[uuid.UUID]*ent.MailFolder, len(folderIDs))
	if len(folderIDs) != 0 {
		rows, err := transaction.MailFolder.Query().Where(mailfolder.MailAccountIDEQ(account.ID), mailfolder.IDIn(mailboxEntPrepareSortedUUIDSet(folderIDs)...), mailfolder.DeletedAtIsNil()).All(ctx)
		if err != nil { return nil, mailboxEntSafeError(err) }
		for _, row := range rows {
			if row == nil || row.ID == uuid.Nil || row.MailAccountID != account.ID || row.DeletedAt != nil { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
			if _, duplicate := folders[row.ID]; duplicate { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
			folders[row.ID] = row
		}
		if len(folders) != len(folderIDs) { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
	}

	labelLinks, err := transaction.MailMessageLabel.Query().Where(mailmessagelabel.MailMessageIDIn(messageIDs...)).All(ctx)
	if err != nil { return nil, mailboxEntSafeError(err) }
	labelIDs := make(map[uuid.UUID]struct{}, len(labelLinks))
	labelsByMessage := make(map[uuid.UUID][]*ent.MailMessageLabel, len(existing))
	for _, link := range labelLinks {
		if link == nil || link.ID == uuid.Nil || link.MailMessageID == uuid.Nil || link.MailLabelID == uuid.Nil || (!link.Manual && !link.EngineRule) { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
		if _, found := existing[link.MailMessageID]; !found { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
		labelsByMessage[link.MailMessageID] = append(labelsByMessage[link.MailMessageID], link)
		labelIDs[link.MailLabelID] = struct{}{}
	}
	labels := make(map[uuid.UUID]*ent.MailLabel, len(labelIDs))
	if len(labelIDs) != 0 {
		rows, err := transaction.MailLabel.Query().Where(maillabel.MailAccountIDEQ(account.ID), maillabel.IDIn(mailboxEntPrepareSortedUUIDSet(labelIDs)...)).All(ctx)
		if err != nil { return nil, mailboxEntSafeError(err) }
		for _, row := range rows {
			if row == nil || row.ID == uuid.Nil || row.MailAccountID != account.ID { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
			if _, duplicate := labels[row.ID]; duplicate { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
			labels[row.ID] = row
		}
		if len(labels) != len(labelIDs) { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
	}

	for messageID, row := range existing {
		thread, found := threads[row.MailThreadID]
		if !found || thread.ID != row.MailThreadID || (row.DeletedAt == nil && thread.DeletedAt != nil) { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
		folderState := make([]MailboxEventID, 0)
		flags := EngineMessageFlags{}
		if row.DeletedAt == nil {
			folderState = make([]MailboxEventID, 0, len(foldersByMessage[messageID]))
			seenFolders := make(map[uuid.UUID]struct{}, len(foldersByMessage[messageID]))
			for _, link := range foldersByMessage[messageID] {
				if _, found := folders[link.MailFolderID]; !found { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
				if _, duplicate := seenFolders[link.MailFolderID]; duplicate { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
				seenFolders[link.MailFolderID] = struct{}{}
				folderState = append(folderState, MailboxEventID{AccountID: account.ID, ID: link.MailFolderID})
			}
			sort.Slice(folderState, func(left, right int) bool { return syncLessUUID(folderState[left].ID, folderState[right].ID) })
			flags = EngineMessageFlags{Seen: row.IsSeen, Answered: row.IsAnswered, Draft: row.IsDraft, Flagged: row.IsFlagged, Junk: row.IsJunk}
		}
		labelState := make([]MailboxSyncPolicyLabelState, 0, len(labelsByMessage[messageID]))
		seenLabels := make(map[uuid.UUID]struct{}, len(labelsByMessage[messageID]))
		for _, link := range labelsByMessage[messageID] {
			if _, duplicate := seenLabels[link.MailLabelID]; duplicate { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
			seenLabels[link.MailLabelID] = struct{}{}
			label, found := labels[link.MailLabelID]
			if !found { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
			if label.DeletedAt != nil {
				if row.DeletedAt == nil { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
				continue
			}
			labelState = append(labelState, MailboxSyncPolicyLabelState{Label: MailboxEventID{AccountID: account.ID, ID: link.MailLabelID}, Manual: link.Manual, EngineRule: link.EngineRule})
		}
		sort.Slice(labelState, func(left, right int) bool { return syncLessUUID(labelState[left].Label.ID, labelState[right].Label.ID) })
		var calendarMethod *MailCalendarMethod
		if row.CalendarMethod != nil {
			value := MailCalendarMethod(*row.CalendarMethod)
			if !validMailCalendarMethod(&value) { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
			calendarMethod = &value
		}
		result[messageID] = mailboxEntPreparePersistedState{
			previous: &MailboxSyncPolicyPreviousMessage{
				Thread: MailboxEventID{AccountID: account.ID, ID: thread.ID}, FolderIDs: folderState,
				Flags: flags,
				Labels: labelState, ReceivedEventID: syncCopyUUIDPointer(row.ReceivedEventID),
			},
			calendarMethod: calendarMethod,
		}
	}
	return result, nil
}

func mailboxEntPrepareSortedUUIDStrings(values map[uuid.UUID]string) []uuid.UUID {
	set := make(map[uuid.UUID]struct{}, len(values))
	for value := range values { set[value] = struct{}{} }
	return mailboxEntPrepareSortedUUIDSet(set)
}

func mailboxEntPrepareSortedUUIDSet(values map[uuid.UUID]struct{}) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(values))
	for value := range values { result = append(result, value) }
	sort.Slice(result, func(left, right int) bool { return syncLessUUID(result[left], result[right]) })
	return result
}

type mailboxEntPrepareThreadKey struct { version ThreadKeyVersion; key ThreadKey }

func mailboxEntPrepareThreads(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, projections []*EngineMessageProjection) (map[string]MailboxSyncPolicyThreadResolution, error) {
	ancestors, err := mailboxEntPrepareAncestors(ctx, transaction, account, projections)
	if err != nil { return nil, err }
	identities := make(map[string]ThreadIdentity, len(projections))
	keys := make(map[mailboxEntPrepareThreadKey]struct{}, len(projections))
	for _, projection := range projections {
		identity, err := DeriveThreadIdentity(ThreadingInput{References: projection.References, InReplyTo: projection.InReplyTo, Subject: projection.Subject, From: projection.From, To: projection.To, CC: projection.CC, Ancestors: ancestors[projection.MessageID], AncestorLookupComplete: true})
		if err != nil || identity.KeyVersion != ThreadKeyVersionV1 || threadKeyIsZero(identity.Key) { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
		identities[projection.MessageID] = identity
		keys[mailboxEntPrepareThreadKey{version: identity.KeyVersion, key: identity.Key}] = struct{}{}
	}
	if len(keys) == 0 { return map[string]MailboxSyncPolicyThreadResolution{}, nil }
	orderedKeys := mailboxEntPrepareOrderedThreadKeys(keys)
	threadPredicates := make([]predicate.MailThread, 0, len(orderedKeys))
	for _, threadKey := range orderedKeys {
		identity := ThreadIdentity{KeyVersion: threadKey.version, Key: threadKey.key}
		threadPredicates = append(threadPredicates, mailthread.And(mailthread.ThreadKeyVersionEQ(int32(identity.KeyVersion)), mailthread.ThreadKeyEQ(identity.Key[:])))
	}
	rows, err := transaction.MailThread.Query().Where(mailthread.MailAccountIDEQ(account.ID), mailthread.Or(threadPredicates...), mailthread.DeletedAtIsNil()).All(ctx)
	if err != nil { return nil, mailboxEntSafeError(err) }
	existing := make(map[mailboxEntPrepareThreadKey]*ent.MailThread, len(rows))
	for _, row := range rows {
		key, ok := mailboxEntPrepareThreadKeyFromRow(row)
		if !ok || row.MailAccountID != account.ID || row.ID == uuid.Nil || row.DeletedAt != nil { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
		if _, duplicate := existing[key]; duplicate { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
		existing[key] = row
	}
	proposed := make(map[uuid.UUID]mailboxEntPrepareThreadKey, len(orderedKeys))
	resolved := make(map[mailboxEntPrepareThreadKey]MailboxEventID, len(orderedKeys))
	for _, key := range orderedKeys {
		if row, found := existing[key]; found {
			resolved[key] = MailboxEventID{AccountID: account.ID, ID: row.ID}
			continue
		}
		identityBytes := make([]byte, 1+len(key.key))
		identityBytes[0] = byte(key.version)
		copy(identityBytes[1:], key.key[:])
		id, err := mailboxEntStableSyncID(account.ID, "thread", identityBytes)
		if err != nil { return nil, err }
		if other, duplicate := proposed[id]; duplicate && other != key { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
		proposed[id] = key
		resolved[key] = MailboxEventID{AccountID: account.ID, ID: id}
	}
	if err := mailboxEntPrepareThreadCollisions(ctx, transaction, account, proposed); err != nil { return nil, err }
	result := make(map[string]MailboxSyncPolicyThreadResolution, len(projections))
	for _, projection := range projections {
		identity := identities[projection.MessageID]
		key := mailboxEntPrepareThreadKey{version: identity.KeyVersion, key: identity.Key}
		thread, found := resolved[key]
		if !found { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
		result[projection.MessageID] = MailboxSyncPolicyThreadResolution{Thread: thread, KeyVersion: identity.KeyVersion, Key: identity.Key, Ancestors: append([]ThreadAncestor(nil), ancestors[projection.MessageID]...), AncestorLookupComplete: true}
	}
	return result, nil
}

func mailboxEntPrepareAncestors(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, projections []*EngineMessageProjection) (map[string][]ThreadAncestor, error) {
	candidateSet := make(map[string]struct{})
	for _, projection := range projections {
		for _, reference := range projection.References { if validThreadMessageID(reference) { candidateSet[reference] = struct{}{} } }
		if validThreadMessageID(projection.InReplyTo) { candidateSet[projection.InReplyTo] = struct{}{} }
	}
	result := make(map[string][]ThreadAncestor, len(projections))
	if len(candidateSet) == 0 { return result, nil }
	candidates := make([]string, 0, len(candidateSet))
	for candidate := range candidateSet { candidates = append(candidates, candidate) }
	sort.Strings(candidates)
	messages, err := transaction.MailMessage.Query().Where(mailmessage.MailAccountIDEQ(account.ID), mailmessage.InternetMessageIDIn(candidates...), mailmessage.DeletedAtIsNil()).All(ctx)
	if err != nil { return nil, mailboxEntSafeError(err) }
	threadSet := make(map[uuid.UUID]struct{}, len(messages))
	for _, row := range messages {
		if row == nil || row.ID == uuid.Nil || row.MailAccountID != account.ID || row.DeletedAt != nil || row.InternetMessageID == nil || !validThreadMessageID(*row.InternetMessageID) || row.MailThreadID == uuid.Nil { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
		threadSet[row.MailThreadID] = struct{}{}
	}
	threads := make(map[uuid.UUID]*ent.MailThread, len(threadSet))
	if len(threadSet) != 0 {
		rows, err := transaction.MailThread.Query().Where(mailthread.MailAccountIDEQ(account.ID), mailthread.IDIn(mailboxEntPrepareSortedUUIDSet(threadSet)...), mailthread.DeletedAtIsNil()).All(ctx)
		if err != nil { return nil, mailboxEntSafeError(err) }
		for _, row := range rows {
			key, ok := mailboxEntPrepareThreadKeyFromRow(row)
			if !ok || row.MailAccountID != account.ID || row.DeletedAt != nil { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
			_ = key
			threads[row.ID] = row
		}
	}
	for _, message := range messages {
		thread, found := threads[message.MailThreadID]
		if !found { continue }
		key, ok := mailboxEntPrepareThreadKeyFromRow(thread)
		if !ok { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
		for _, projection := range projections {
			for _, candidate := range append(append([]string(nil), projection.References...), projection.InReplyTo) {
				if candidate == *message.InternetMessageID { result[projection.MessageID] = append(result[projection.MessageID], ThreadAncestor{MessageID: candidate, KeyVersion: key.version, Key: key.key}) }
			}
		}
	}
	for messageID := range result {
		sort.Slice(result[messageID], func(left, right int) bool {
			if result[messageID][left].MessageID != result[messageID][right].MessageID { return result[messageID][left].MessageID < result[messageID][right].MessageID }
			if result[messageID][left].KeyVersion != result[messageID][right].KeyVersion { return result[messageID][left].KeyVersion < result[messageID][right].KeyVersion }
			return string(result[messageID][left].Key[:]) < string(result[messageID][right].Key[:])
		})
	}
	return result, nil
}

func mailboxEntPrepareOrderedThreadKeys(values map[mailboxEntPrepareThreadKey]struct{}) []mailboxEntPrepareThreadKey {
	result := make([]mailboxEntPrepareThreadKey, 0, len(values))
	for value := range values { result = append(result, value) }
	sort.Slice(result, func(left, right int) bool {
		if result[left].version != result[right].version { return result[left].version < result[right].version }
		return string(result[left].key[:]) < string(result[right].key[:])
	})
	return result
}

func mailboxEntPrepareThreadKeyFromRow(row *ent.MailThread) (mailboxEntPrepareThreadKey, bool) {
	if row == nil || row.ThreadKeyVersion != int32(ThreadKeyVersionV1) || len(row.ThreadKey) != len(ThreadKey{}) { return mailboxEntPrepareThreadKey{}, false }
	var key ThreadKey
	copy(key[:], row.ThreadKey)
	if threadKeyIsZero(key) { return mailboxEntPrepareThreadKey{}, false }
	return mailboxEntPrepareThreadKey{version: ThreadKeyVersion(row.ThreadKeyVersion), key: key}, true
}

func mailboxEntPrepareThreadCollisions(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, proposed map[uuid.UUID]mailboxEntPrepareThreadKey) error {
	if len(proposed) == 0 { return nil }
	proposedIDs := make(map[uuid.UUID]struct{}, len(proposed))
	for id := range proposed { proposedIDs[id] = struct{}{} }
	rows, err := transaction.MailThread.Query().Where(mailthread.IDIn(mailboxEntPrepareSortedUUIDSet(proposedIDs)...)).All(ctx)
	if err != nil { return mailboxEntSafeError(err) }
	for _, row := range rows {
		want, found := proposed[row.ID]
		actual, ok := mailboxEntPrepareThreadKeyFromRow(row)
		if !found || !ok || row.MailAccountID != account.ID || actual != want { return mailboxServiceError(MailboxServiceErrorInvalid) }
	}
	return mailboxEntPrepareStableIDsUnclaimed(ctx, transaction, "thread", mailboxEntPrepareSortedUUIDSet(proposedIDs))
}

func mailboxEntPrepareStableIDsUnclaimed(ctx context.Context, transaction *ent.Tx, kind string, ids []uuid.UUID) error {
	if len(ids) == 0 { return nil }
	if kind != "folder" {
		rows, err := transaction.MailFolder.Query().Where(mailfolder.IDIn(ids...)).All(ctx)
		if err != nil { return mailboxEntSafeError(err) }
		if len(rows) != 0 { return mailboxServiceError(MailboxServiceErrorInvalid) }
	}
	if kind != "message" {
		rows, err := transaction.MailMessage.Query().Where(mailmessage.IDIn(ids...)).All(ctx)
		if err != nil { return mailboxEntSafeError(err) }
		if len(rows) != 0 { return mailboxServiceError(MailboxServiceErrorInvalid) }
	}
	if kind != "thread" {
		rows, err := transaction.MailThread.Query().Where(mailthread.IDIn(ids...)).All(ctx)
		if err != nil { return mailboxEntSafeError(err) }
		if len(rows) != 0 { return mailboxServiceError(MailboxServiceErrorInvalid) }
	}
	rows, err := transaction.MailLabel.Query().Where(maillabel.IDIn(ids...)).All(ctx)
	if err != nil { return mailboxEntSafeError(err) }
	if len(rows) != 0 { return mailboxServiceError(MailboxServiceErrorInvalid) }
	return nil
}

func mailboxEntPrepareMessages(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, page MailboxSyncPolicyPage, messageRows map[string]*ent.MailMessage, states map[uuid.UUID]mailboxEntPreparePersistedState, threads map[string]MailboxSyncPolicyThreadResolution, provenance map[string]MailboxSyncPolicyProvenance) ([]MailboxSyncPolicyMessage, []MailboxSyncPolicyDestroyed, error) {
	projections := mailboxEntPrepareProjectionPointers(&page)
	messageIDs, err := mailboxEntPrepareMessageIDs(ctx, transaction, account, projections, messageRows)
	if err != nil { return nil, nil, err }
	messages := make([]MailboxSyncPolicyMessage, 0, len(projections))
	for _, projection := range projections {
		thread, found := threads[projection.MessageID]
		if !found { return nil, nil, mailboxServiceError(MailboxServiceErrorInvalid) }
		var previous *MailboxSyncPolicyPreviousMessage
		var calendarMethod *MailCalendarMethod
		if row, found := messageRows[projection.MessageID]; found {
			state, found := states[row.ID]
			if !found { return nil, nil, mailboxServiceError(MailboxServiceErrorInvalid) }
			previous = cloneMailboxSyncPrevious(state.previous)
			if state.calendarMethod != nil { value := *state.calendarMethod; calendarMethod = &value }
		}
		messages = append(messages, MailboxSyncPolicyMessage{EngineMessageID: projection.MessageID, Message: MailboxEventID{AccountID: account.ID, ID: messageIDs[projection.MessageID]}, Thread: thread, Previous: previous, Provenance: MailboxSyncPolicyProvenance{ConfirmedSubmission: cloneMailboxSyncProof(provenance[projection.MessageID].ConfirmedSubmission), ConfirmedDraft: cloneMailboxSyncProof(provenance[projection.MessageID].ConfirmedDraft)}, CalendarMethod: calendarMethod})
	}
	destroyed := make([]MailboxSyncPolicyDestroyed, 0)
	if page.Change != nil {
		destroyed = make([]MailboxSyncPolicyDestroyed, 0, len(page.Change.DestroyedIDs))
		for _, engineMessageID := range page.Change.DestroyedIDs {
			row, found := messageRows[engineMessageID]
			if !found || row.DeletedAt != nil {
				destroyed = append(destroyed, MailboxSyncPolicyDestroyed{EngineMessageID: engineMessageID, Projected: false})
				continue
			}
			state, found := states[row.ID]
			if !found || state.previous == nil { return nil, nil, mailboxServiceError(MailboxServiceErrorInvalid) }
			destroyed = append(destroyed, MailboxSyncPolicyDestroyed{EngineMessageID: engineMessageID, Projected: true, Message: MailboxEventID{AccountID: account.ID, ID: row.ID}, Thread: state.previous.Thread, Labels: syncCopyLabelStates(state.previous.Labels)})
		}
	}
	sort.Slice(messages, func(left, right int) bool { return messages[left].EngineMessageID < messages[right].EngineMessageID })
	sort.Slice(destroyed, func(left, right int) bool { return destroyed[left].EngineMessageID < destroyed[right].EngineMessageID })
	return messages, destroyed, nil
}

func mailboxEntPrepareMessageIDs(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, projections []*EngineMessageProjection, existing map[string]*ent.MailMessage) (map[string]uuid.UUID, error) {
	proposed := make(map[uuid.UUID]string, len(projections))
	result := make(map[string]uuid.UUID, len(projections))
	for _, projection := range projections {
		if row, found := existing[projection.MessageID]; found {
			if row.ID == uuid.Nil || row.MailAccountID != account.ID { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
			result[projection.MessageID] = row.ID
			continue
		}
		id, err := mailboxEntStableSyncID(account.ID, "message", []byte(projection.MessageID))
		if err != nil { return nil, err }
		if other, duplicate := proposed[id]; duplicate && other != projection.MessageID { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
		proposed[id] = projection.MessageID
		result[projection.MessageID] = id
	}
	if len(proposed) == 0 { return result, nil }
	ids := mailboxEntPrepareSortedUUIDStrings(proposed)
	rows, err := transaction.MailMessage.Query().Where(mailmessage.IDIn(ids...)).All(ctx)
	if err != nil { return nil, mailboxEntSafeError(err) }
	for _, row := range rows {
		want, found := proposed[row.ID]
		if !found || row.MailAccountID != account.ID || row.EngineMessageID != want { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }
	}
	if err := mailboxEntPrepareStableIDsUnclaimed(ctx, transaction, "message", ids); err != nil { return nil, err }
	return result, nil
}

func mailboxEntPrepareCandidateAvailability(ctx context.Context, transaction *ent.Tx, candidates []MailboxSyncEventCandidate) error {
	candidateSet := make(map[uuid.UUID]struct{}, 2*len(candidates))
	for _, candidate := range candidates {
		candidateSet[candidate.ChangedEventID] = struct{}{}
		if !candidate.Destroyed { candidateSet[candidate.ReceivedEventID] = struct{}{} }
	}
	candidateIDs := mailboxEntPrepareSortedUUIDSet(candidateSet)
	if len(candidateIDs) == 0 { return nil }
	rows, err := transaction.OutboxEvent.Query().Where(outboxevent.IDIn(candidateIDs...)).All(ctx)
	if err != nil { return mailboxEntSafeError(err) }
	if len(rows) != 0 { return mailboxServiceError(MailboxServiceErrorInvalid) }
	return nil
}

func mailboxEntPrepareSelectEvents(input *MailboxSyncPolicyInput, candidates []MailboxSyncEventCandidate) error {
	if input == nil { return mailboxServiceError(MailboxServiceErrorInvalid) }
	accountID, err := syncValidateEventContext(input.Event)
	if err != nil { return mailboxServiceError(MailboxServiceErrorInvalid) }
	page, err := syncValidatePage(input.Page)
	if err != nil { return mailboxServiceError(MailboxServiceErrorInvalid) }
	folders, _, err := syncIndexFolders(accountID, page.accountID, input.Folders)
	if err != nil { return mailboxServiceError(MailboxServiceErrorInvalid) }
	labels, err := syncIndexLabels(accountID, input.Labels)
	if err != nil { return mailboxServiceError(MailboxServiceErrorInvalid) }
	candidateByEngineID := make(map[string]MailboxSyncEventCandidate, len(candidates))
	for _, candidate := range candidates { candidateByEngineID[candidate.EngineMessageID] = candidate }
	projections := make(map[string]EngineMessageProjection, len(page.projections))
	for _, projection := range page.projections { projections[projection.MessageID] = projection }
	for index := range input.Messages {
		message := &input.Messages[index]
		projection, found := projections[message.EngineMessageID]
		candidate, candidateFound := candidateByEngineID[message.EngineMessageID]
		if !found || !candidateFound || candidate.Destroyed { return mailboxServiceError(MailboxServiceErrorInvalid) }
		if message.Previous == nil {
			if projection.Origin == EngineProjectionOriginInboundOrUnknown && projection.ReceivedAt != nil && projection.ReceivedAt.After(input.EventCutoff) { message.ReceivedEventID = candidate.ReceivedEventID }
			continue
		}
		currentFolders, err := syncResolveFolders(accountID, projection.FolderIDs, folders)
		if err != nil { return mailboxServiceError(MailboxServiceErrorInvalid) }
		previousFolders, previousLabels, _, err := syncValidatePrevious(accountID, message.Previous)
		if err != nil { return mailboxServiceError(MailboxServiceErrorInvalid) }
		desiredLabels, err := syncDesiredLabels(accountID, projection.ReservedKeywords, previousLabels, labels)
		if err != nil { return mailboxServiceError(MailboxServiceErrorInvalid) }
		if len(syncChangedFields(currentFolders, previousFolders, projection.Flags, message.Previous, desiredLabels, previousLabels)) != 0 { message.ChangedEvent = MailboxSyncPolicyEventReservation{ID: candidate.ChangedEventID} }
	}
	for index := range input.Destroyed {
		destroyed := &input.Destroyed[index]
		candidate, found := candidateByEngineID[destroyed.EngineMessageID]
		if !found || !candidate.Destroyed { return mailboxServiceError(MailboxServiceErrorInvalid) }
		if destroyed.Projected { destroyed.ChangedEvent = MailboxSyncPolicyEventReservation{ID: candidate.ChangedEventID} }
	}
	return nil
}

func mailboxEntPrepareCopyInput(input MailboxSyncPolicyInput) MailboxSyncPolicyInput { return cloneMailboxSyncPolicyInput(input) }
