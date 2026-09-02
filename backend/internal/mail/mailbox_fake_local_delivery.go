package mail

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"log/slog"
	"reflect"
	"sort"
	"time"
)

const (
	fakeLocalDeliveryCalendarLocatorDomain  = "hub.mail.fake.local-delivery.calendar-locator.v1"
	fakeLocalDeliveryCalendarLocatorVersion = byte(1)
)

type fakeLocalDeliveryFailureStage string

const (
	fakeLocalDeliveryFailureStageDraftBinding        fakeLocalDeliveryFailureStage = "draft_binding"
	fakeLocalDeliveryFailureStageSenderSentFolder    fakeLocalDeliveryFailureStage = "sender_sent_folder"
	fakeLocalDeliveryFailureStageRecipientRouting    fakeLocalDeliveryFailureStage = "recipient_routing"
	fakeLocalDeliveryFailureStageCalendarStage       fakeLocalDeliveryFailureStage = "calendar_stage"
	fakeLocalDeliveryFailureStageFault               fakeLocalDeliveryFailureStage = "fault"
	fakeLocalDeliveryFailureStageCandidateValidation fakeLocalDeliveryFailureStage = "candidate_validation"
	fakeLocalDeliveryFailureStageCommitRecheck       fakeLocalDeliveryFailureStage = "commit_recheck"
	fakeLocalDeliveryFailureStageCandidateCollision  fakeLocalDeliveryFailureStage = "candidate_collision"
	fakeLocalDeliveryFailureStageCommitted           fakeLocalDeliveryFailureStage = "committed"
)

type fakeLocalDeliveryCalendarMismatchReason string

const (
	fakeLocalDeliveryCalendarMismatchReasonBCC                   fakeLocalDeliveryCalendarMismatchReason = "bcc"
	fakeLocalDeliveryCalendarMismatchReasonDraftsFolder          fakeLocalDeliveryCalendarMismatchReason = "drafts_folder"
	fakeLocalDeliveryCalendarMismatchReasonStageDraftIdentity    fakeLocalDeliveryCalendarMismatchReason = "stage_draft_identity"
	fakeLocalDeliveryCalendarMismatchReasonStageDraftImportState fakeLocalDeliveryCalendarMismatchReason = "stage_draft_import_state"
	fakeLocalDeliveryCalendarMismatchReasonMethodInvalid         fakeLocalDeliveryCalendarMismatchReason = "method_invalid"
)

type fakeLocalDeliveryTarget struct {
	accountID     string
	inboxFolderID string
}

type fakeLocalDeliveryCalendar struct {
	method        MailCalendarMethod
	itip          []byte
	bindingDigest [sha256.Size]byte
	contentDigest [sha256.Size]byte
}

type fakeLocalDeliveryPlan struct {
	senderSentFolderID string
	targets            []fakeLocalDeliveryTarget
	calendar           *fakeLocalDeliveryCalendar
}

type fakeLocalDeliveryCandidateIDs struct {
	submissionID        string
	senderMessageID     string
	recipientMessageIDs []string
	recipientVersions   []string
	reservedIDs         []string
}

type fakeLocalDeliveryRecipientCandidate struct {
	message EngineMessage
	part    *FakeCalendarPart
}

// submitMailboxSubmissionWithLocalDelivery creates the existing sender
// projection together with all resolved local inbox projections as one Fake
// state transition. All routing and candidate validation precede reservation.
func (f *FakeMailEngine) submitMailboxSubmissionWithLocalDelivery(submission EngineSubmission) (result EngineSubmissionResult, err error) {
	stage := fakeLocalDeliveryFailureStageDraftBinding
	calendarMismatchReason := fakeLocalDeliveryCalendarMismatchReason("")
	submissionWritten := false
	secondPassLocked := false
	defer func() {
		if secondPassLocked {
			f.mu.Unlock()
		}
		if err != nil {
			attributes := []any{
				"fake_local_delivery_stage", string(stage),
				"submission_written", submissionWritten,
			}
			if calendarMismatchReason != "" {
				attributes = append(attributes, "fake_local_delivery_calendar_mismatch_reason", string(calendarMismatchReason))
			}
			slog.Warn(
				"mail fake local delivery submit failed",
				attributes...,
			)
		}
	}()

	f.mu.Lock()
	if stored, ok := f.submissions[submission.AccountID][submission.SubmissionKey]; ok {
		if !fakeSubmissionBindingMatches(stored, submission) {
			f.mu.Unlock()
			return nil, fakeEngineError(EngineErrorIndeterminate)
		}
		f.mu.Unlock()
		return stored.accepted, nil
	}
	draft, ok := f.drafts[submission.AccountID][submission.DraftKey]
	if !ok || draft.message.Projection.MessageID != submission.MessageID {
		f.mu.Unlock()
		return nil, fakeEngineError(EngineErrorNotFound)
	}
	if draft.message.Projection.Version != submission.ExpectedVersion {
		f.mu.Unlock()
		return nil, fakeEngineError(EngineErrorConflict)
	}
	if len(draft.message.Projection.Attachments) != 0 {
		f.mu.Unlock()
		return nil, fakeEngineError(EngineErrorRejectedPolicy)
	}
	plan, planStage, planCalendarMismatchReason, err := f.fakeLocalDeliveryPlanLocked(submission, draft)
	if err != nil {
		stage = planStage
		calendarMismatchReason = planCalendarMismatchReason
		f.mu.Unlock()
		return nil, err
	}
	stage = fakeLocalDeliveryFailureStageFault
	fault, faulted := f.takeFaultLocked(FakeMailOperationSubmit)
	if faulted && !fault.AfterDispatch {
		f.mu.Unlock()
		return nil, fakeEngineError(fault.Kind)
	}
	message := cloneMessage(draft.message)
	f.mu.Unlock()

	stage = fakeLocalDeliveryFailureStageCandidateValidation
	message.Projection.MessageID = ""
	if int64(messageVariableBytes(message)) > submission.MaxMessageBytes {
		return nil, fakeEngineError(EngineErrorPayloadTooLarge)
	}

	sentAt := f.currentTime()
	if !validMailEventTime(sentAt) {
		return nil, fakeEngineError(EngineErrorRejectedPolicy)
	}
	ids, err := f.fakeLocalDeliveryCandidateIDs(submission.MessageID, len(plan.targets))
	if err != nil {
		return nil, err
	}
	message.Projection.MessageID = ids.senderMessageID
	message.Projection.Flags.Draft = false
	message.Projection.Origin = EngineProjectionOriginConfirmedOwnerSubmission
	message.Projection.FolderIDs = []string{plan.senderSentFolderID}
	message.Projection.SentAt = &sentAt
	if int64(messageVariableBytes(message)) > submission.MaxMessageBytes {
		return nil, fakeEngineError(EngineErrorPayloadTooLarge)
	}
	if err := validateEngineMessage(message); err != nil {
		return nil, err
	}
	recipients, err := fakeLocalDeliveryRecipientCandidates(message, plan, ids, sentAt)
	if err != nil {
		return nil, err
	}

	stage = fakeLocalDeliveryFailureStageCommitRecheck
	f.mu.Lock()
	secondPassLocked = true
	if stored, ok := f.submissions[submission.AccountID][submission.SubmissionKey]; ok {
		if !fakeSubmissionBindingMatches(stored, submission) {
			return nil, fakeEngineError(EngineErrorIndeterminate)
		}
		return stored.accepted, nil
	}
	draft, ok = f.drafts[submission.AccountID][submission.DraftKey]
	if !ok || draft.message.Projection.MessageID != submission.MessageID {
		return nil, fakeEngineError(EngineErrorNotFound)
	}
	if draft.message.Projection.Version != submission.ExpectedVersion {
		return nil, fakeEngineError(EngineErrorConflict)
	}
	if len(draft.message.Projection.Attachments) != 0 {
		return nil, fakeEngineError(EngineErrorRejectedPolicy)
	}
	currentPlan, planStage, planCalendarMismatchReason, err := f.fakeLocalDeliveryPlanLocked(submission, draft)
	if err != nil {
		stage = planStage
		calendarMismatchReason = planCalendarMismatchReason
		return nil, err
	}
	stage = fakeLocalDeliveryFailureStageCandidateCollision
	if !fakeLocalDeliveryPlansEqual(plan, currentPlan) {
		return nil, fakeEngineError(EngineErrorConflict)
	}
	storedDraftMessage, exists := f.messages[message.Projection.AccountID][message.Projection.MessageID]
	if !exists || !reflect.DeepEqual(cloneMessage(draft.message), cloneMessage(storedDraftMessage)) {
		return nil, fakeEngineError(EngineErrorConflict)
	}
	for _, recipient := range recipients {
		if _, exists := f.messages[recipient.message.Projection.AccountID][recipient.message.Projection.MessageID]; exists {
			return nil, fakeEngineError(EngineErrorConflict)
		}
		if recipient.part == nil {
			continue
		}
		key := fakeCalendarPartKey{
			accountID:   recipient.part.AccountID,
			messageID:   recipient.part.MessageID,
			locatorHash: recipient.part.LocatorHash,
		}
		if _, exists := f.calendarParts[key]; exists {
			return nil, fakeEngineError(EngineErrorConflict)
		}
	}
	if err := f.reserveOpaqueIDsLocked(ids.reservedIDs...); err != nil {
		return nil, err
	}
	f.putMessageLocked(message, true)
	for _, recipient := range recipients {
		f.putMessageLocked(recipient.message, true)
		if recipient.part == nil {
			continue
		}
		if f.calendarParts == nil {
			f.calendarParts = make(map[fakeCalendarPartKey]fakeCalendarPart)
		}
		f.calendarParts[fakeCalendarPartKey{
			accountID:   recipient.part.AccountID,
			messageID:   recipient.part.MessageID,
			locatorHash: recipient.part.LocatorHash,
		}] = fakeCalendarPart{
			version: recipient.part.Version,
			method:  recipient.part.Method,
			itip:    append([]byte(nil), recipient.part.ITIP...),
		}
	}
	if f.submissions[submission.AccountID] == nil {
		f.submissions[submission.AccountID] = make(map[string]fakeSubmission)
	}
	accepted := EngineSubmissionAccepted{SubmissionID: ids.submissionID, MessageID: ids.senderMessageID}
	f.submissions[submission.AccountID][submission.SubmissionKey] = fakeSubmission{
		accepted:        accepted,
		draftKey:        submission.DraftKey,
		messageID:       submission.MessageID,
		expectedVersion: submission.ExpectedVersion,
	}
	submissionWritten = true
	stage = fakeLocalDeliveryFailureStageCommitted
	if faulted && fault.AfterDispatch {
		return nil, fakeEngineError(fault.Kind)
	}
	return accepted, nil
}

func (f *FakeMailEngine) saveDraftWithLocalFolder(draft EngineDraft) (EngineMessage, error) {
	f.mu.Lock()
	draftFolderID, err := f.fakeLocalDeliverySystemFolderIDLocked(draft.AccountID, EngineFolderRoleDrafts)
	if err != nil {
		f.mu.Unlock()
		return EngineMessage{}, err
	}
	if draftMinimumBytesInFolder(draft, draftFolderID) > draft.MaxContentBytes {
		f.mu.Unlock()
		return EngineMessage{}, fakeEngineError(EngineErrorPayloadTooLarge)
	}
	fault, faulted := f.takeFaultLocked(FakeMailOperationSaveDraft)
	if faulted && !fault.AfterDispatch {
		f.mu.Unlock()
		return EngineMessage{}, fakeEngineError(fault.Kind)
	}
	existing, exists := f.drafts[draft.AccountID][draft.DraftKey]
	if draft.MessageID == "" {
		if exists {
			if !fakeLocalDeliveryDraftMessageInFolder(existing.message, draft.AccountID, draftFolderID) {
				f.mu.Unlock()
				return EngineMessage{}, fakeEngineError(EngineErrorConflict)
			}
			message := cloneMessage(existing.message)
			f.mu.Unlock()
			if faulted && fault.AfterDispatch {
				return EngineMessage{}, fakeEngineError(fault.Kind)
			}
			return message, nil
		}
	} else {
		if !exists || existing.message.Projection.MessageID != draft.MessageID {
			f.mu.Unlock()
			return EngineMessage{}, fakeEngineError(EngineErrorNotFound)
		}
		if existing.message.Projection.Version != draft.PreviousVersion {
			f.mu.Unlock()
			return EngineMessage{}, fakeEngineError(EngineErrorConflict)
		}
		if !fakeLocalDeliveryDraftMessageInFolder(existing.message, draft.AccountID, draftFolderID) {
			f.mu.Unlock()
			return EngineMessage{}, fakeEngineError(EngineErrorConflict)
		}
	}
	f.mu.Unlock()

	var candidateMessageID string
	if draft.MessageID == "" {
		candidateMessageID, err = f.nextOpaqueID()
		if err != nil {
			return EngineMessage{}, err
		}
	}
	candidateVersion, err := f.nextOpaqueID()
	if err != nil {
		return EngineMessage{}, err
	}
	messageID := draft.MessageID
	if messageID == "" {
		messageID = candidateMessageID
	}
	message := messageFromDraftInFolder(draft, messageID, candidateVersion, f.currentTime(), draftFolderID)
	if int64(messageVariableBytes(message)) > draft.MaxContentBytes || validateEngineMessage(message) != nil {
		return EngineMessage{}, fakeEngineError(EngineErrorPayloadTooLarge)
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	currentDraftFolderID, err := f.fakeLocalDeliverySystemFolderIDLocked(draft.AccountID, EngineFolderRoleDrafts)
	if err != nil {
		return EngineMessage{}, err
	}
	if currentDraftFolderID != draftFolderID {
		return EngineMessage{}, fakeEngineError(EngineErrorConflict)
	}
	existing, exists = f.drafts[draft.AccountID][draft.DraftKey]
	if draft.MessageID == "" {
		if exists {
			if !fakeLocalDeliveryDraftMessageInFolder(existing.message, draft.AccountID, draftFolderID) {
				return EngineMessage{}, fakeEngineError(EngineErrorConflict)
			}
			if faulted && fault.AfterDispatch {
				return EngineMessage{}, fakeEngineError(fault.Kind)
			}
			return cloneMessage(existing.message), nil
		}
		if err := f.reserveOpaqueIDsLocked(candidateMessageID, candidateVersion); err != nil {
			return EngineMessage{}, err
		}
	} else {
		if !exists || existing.message.Projection.MessageID != draft.MessageID {
			return EngineMessage{}, fakeEngineError(EngineErrorNotFound)
		}
		if existing.message.Projection.Version != draft.PreviousVersion {
			return EngineMessage{}, fakeEngineError(EngineErrorConflict)
		}
		if !fakeLocalDeliveryDraftMessageInFolder(existing.message, draft.AccountID, draftFolderID) {
			return EngineMessage{}, fakeEngineError(EngineErrorConflict)
		}
		if err := f.reserveOpaqueIDsLocked(candidateVersion); err != nil {
			return EngineMessage{}, err
		}
	}
	if f.drafts[draft.AccountID] == nil {
		f.drafts[draft.AccountID] = make(map[string]fakeDraft)
	}
	f.retainDraftBlobsLocked(draft)
	f.drafts[draft.AccountID][draft.DraftKey] = fakeDraft{
		message: message,
		identity: fakeDraftIdentity{
			messageID:       draft.MessageID,
			previousVersion: draft.PreviousVersion,
			mutationID:      draft.MutationID,
		},
	}
	f.putMessageLocked(message, true)
	if faulted && fault.AfterDispatch {
		return EngineMessage{}, fakeEngineError(fault.Kind)
	}
	return cloneMessage(message), nil
}

func fakeLocalDeliveryDraftMessageInFolder(message EngineMessage, accountID, folderID string) bool {
	projection := message.Projection
	return validAccountID(accountID) &&
		validOpaqueID(folderID) &&
		validateEngineMessage(message) == nil &&
		projection.AccountID == accountID &&
		projection.Origin == EngineProjectionOriginConfirmedOwnerDraft &&
		projection.Flags.Draft &&
		len(projection.FolderIDs) == 1 &&
		projection.FolderIDs[0] == folderID
}

func (f *FakeMailEngine) fakeLocalDeliveryPlanLocked(submission EngineSubmission, draft fakeDraft) (fakeLocalDeliveryPlan, fakeLocalDeliveryFailureStage, fakeLocalDeliveryCalendarMismatchReason, error) {
	senderSentFolderID, err := f.fakeLocalDeliverySystemFolderIDLocked(submission.AccountID, EngineFolderRoleSent)
	if err != nil {
		return fakeLocalDeliveryPlan{}, fakeLocalDeliveryFailureStageSenderSentFolder, "", err
	}
	targets, err := f.fakeLocalDeliveryTargetsLocked(draft.message.Projection)
	if err != nil {
		return fakeLocalDeliveryPlan{}, fakeLocalDeliveryFailureStageRecipientRouting, "", err
	}
	plan := fakeLocalDeliveryPlan{senderSentFolderID: senderSentFolderID, targets: targets}
	if len(targets) == 0 {
		return plan, "", "", nil
	}
	calendar, calendarMismatchReason, err := f.fakeLocalDeliveryCalendarLocked(submission, draft)
	if err != nil {
		return fakeLocalDeliveryPlan{}, fakeLocalDeliveryFailureStageCalendarStage, calendarMismatchReason, err
	}
	plan.calendar = calendar
	return plan, "", "", nil
}

func (f *FakeMailEngine) fakeLocalDeliveryTargetsLocked(projection EngineMessageProjection) ([]fakeLocalDeliveryTarget, error) {
	if !validAddresses(projection.To, MaxEngineAddressItems) ||
		!validAddresses(projection.CC, MaxEngineAddressItems) ||
		!validAddresses(projection.BCC, MaxEngineAddressItems) ||
		!uniqueRecipients(projection.To, projection.CC, projection.BCC) {
		return nil, fakeEngineError(EngineErrorRejectedPolicy)
	}
	recipientEmails := make(map[string]struct{}, len(projection.To)+len(projection.CC)+len(projection.BCC))
	for _, addresses := range [][]EngineAddress{projection.To, projection.CC, projection.BCC} {
		for _, address := range addresses {
			recipientEmails[address.Email] = struct{}{}
		}
	}
	if len(recipientEmails) == 0 {
		return nil, nil
	}
	accountIDs := make([]string, 0, len(f.accounts))
	for accountID, account := range f.accounts {
		if accountID != account.AccountID || !validAccountID(accountID) || account.State != EngineAccountActive {
			continue
		}
		accountIDs = append(accountIDs, accountID)
	}
	sort.Strings(accountIDs)
	targetIDs := make(map[string]struct{}, len(recipientEmails))
	for email := range recipientEmails {
		matchedAccountID := ""
		for _, accountID := range accountIDs {
			if _, ok := f.identityEmails[accountID][email]; !ok {
				continue
			}
			if matchedAccountID != "" {
				return nil, fakeEngineError(EngineErrorConflict)
			}
			matchedAccountID = accountID
		}
		if matchedAccountID != "" {
			targetIDs[matchedAccountID] = struct{}{}
		}
	}
	targetAccountIDs := make([]string, 0, len(targetIDs))
	for accountID := range targetIDs {
		targetAccountIDs = append(targetAccountIDs, accountID)
	}
	sort.Strings(targetAccountIDs)
	targets := make([]fakeLocalDeliveryTarget, 0, len(targetAccountIDs))
	for _, accountID := range targetAccountIDs {
		inboxFolderID, err := f.fakeLocalDeliverySystemFolderIDLocked(accountID, EngineFolderRoleInbox)
		if err != nil {
			return nil, err
		}
		targets = append(targets, fakeLocalDeliveryTarget{accountID: accountID, inboxFolderID: inboxFolderID})
	}
	return targets, nil
}

func (f *FakeMailEngine) fakeLocalDeliverySystemFolderIDLocked(accountID string, role EngineFolderRole) (string, error) {
	if !validAccountID(accountID) {
		return "", fakeEngineError(EngineErrorRejectedPolicy)
	}
	switch role {
	case EngineFolderRoleInbox, EngineFolderRoleSent, EngineFolderRoleDrafts, EngineFolderRoleArchive, EngineFolderRoleTrash, EngineFolderRoleJunk:
	default:
		return "", fakeEngineError(EngineErrorRejectedPolicy)
	}
	folderID := ""
	for candidateID, folder := range f.folders[accountID] {
		if folder.Role != role {
			continue
		}
		if folder.AccountID != accountID || folder.FolderID != candidateID || folder.ParentID != "" ||
			!validOpaqueID(candidateID) || !validSafeText(folder.Name, MaxEngineMIMEFieldBytes, true) {
			return "", fakeEngineError(EngineErrorUnavailable)
		}
		if folderID != "" {
			return "", fakeEngineError(EngineErrorUnavailable)
		}
		folderID = candidateID
	}
	if folderID == "" {
		return "", fakeEngineError(EngineErrorUnavailable)
	}
	return folderID, nil
}

func (f *FakeMailEngine) fakeLocalDeliveryCalendarLocked(submission EngineSubmission, draft fakeDraft) (*fakeLocalDeliveryCalendar, fakeLocalDeliveryCalendarMismatchReason, error) {
	key := fakeCalendarDraftCreateKey{
		accountID:  submission.AccountID,
		draftKey:   submission.DraftKey,
		mutationID: draft.identity.mutationID,
	}
	stage, found := f.calendarDraftCreates[key]
	if !found {
		return nil, "", nil
	}
	if len(draft.message.Projection.BCC) != 0 {
		return nil, fakeLocalDeliveryCalendarMismatchReasonBCC, fakeEngineError(EngineErrorRejectedPolicy)
	}
	draftFolderID, err := f.fakeLocalDeliverySystemFolderIDLocked(submission.AccountID, EngineFolderRoleDrafts)
	if err != nil {
		return nil, fakeLocalDeliveryCalendarMismatchReasonDraftsFolder, err
	}
	if reason := fakeLocalDeliveryCalendarStageMatchesDraft(key, stage, draft, draftFolderID); reason != "" {
		return nil, reason, fakeEngineError(EngineErrorConflict)
	}
	method, ok := fakeLocalDeliveryCalendarMethod(stage.calendar.method)
	if !ok {
		return nil, fakeLocalDeliveryCalendarMismatchReasonMethodInvalid, fakeEngineError(EngineErrorUnavailable)
	}
	return &fakeLocalDeliveryCalendar{
		method:        method,
		itip:          stage.calendar.calendarBytes(),
		bindingDigest: stage.bindingDigest,
		contentDigest: stage.calendarContentDigest,
	}, "", nil
}

func fakeLocalDeliveryCalendarStageMatchesDraft(key fakeCalendarDraftCreateKey, stage fakeCalendarDraftCreateStage, draft fakeDraft, draftFolderID string) fakeLocalDeliveryCalendarMismatchReason {
	receivedAt := draft.message.Projection.ReceivedAt
	if receivedAt == nil || key.accountID != stage.draft.AccountID || key.draftKey != stage.draft.DraftKey || key.mutationID != stage.draft.MutationID {
		return fakeLocalDeliveryCalendarMismatchReasonStageDraftIdentity
	}
	if !fakeCalendarDraftCreateImportedMessageMatches(draft, stage, mailboxDraftCalendarCreateImport{
		accountID:  key.accountID,
		draftKey:   key.draftKey,
		mutationID: key.mutationID,
		blobID:     stage.blobID,
		sizeBytes:  stage.sizeBytes,
		receivedAt: *receivedAt,
	}, draftFolderID) {
		return fakeLocalDeliveryCalendarMismatchReasonStageDraftImportState
	}
	return ""
}

func fakeLocalDeliveryCalendarMethod(method CalendarMailDeliveryMethod) (MailCalendarMethod, bool) {
	switch method {
	case CalendarMailDeliveryMethodRequest:
		return MailCalendarMethodRequest, true
	case CalendarMailDeliveryMethodCancel:
		return MailCalendarMethodCancel, true
	case CalendarMailDeliveryMethodReply:
		return MailCalendarMethodReply, true
	default:
		var empty MailCalendarMethod
		return empty, false
	}
}

func fakeLocalDeliveryCalendarDeliveryMethod(method MailCalendarMethod) (CalendarMailDeliveryMethod, bool) {
	switch method {
	case MailCalendarMethodRequest:
		return CalendarMailDeliveryMethodRequest, true
	case MailCalendarMethodCancel:
		return CalendarMailDeliveryMethodCancel, true
	case MailCalendarMethodReply:
		return CalendarMailDeliveryMethodReply, true
	default:
		var empty CalendarMailDeliveryMethod
		return empty, false
	}
}

func (f *FakeMailEngine) fakeLocalDeliveryCandidateIDs(senderMessageID string, targetCount int) (fakeLocalDeliveryCandidateIDs, error) {
	if !validOpaqueID(senderMessageID) || targetCount < 0 {
		return fakeLocalDeliveryCandidateIDs{}, fakeEngineError(EngineErrorRejectedPolicy)
	}
	submissionID, err := f.nextOpaqueID()
	if err != nil {
		return fakeLocalDeliveryCandidateIDs{}, err
	}
	result := fakeLocalDeliveryCandidateIDs{
		submissionID:        submissionID,
		senderMessageID:     senderMessageID,
		recipientMessageIDs: make([]string, targetCount),
		recipientVersions:   make([]string, targetCount),
		reservedIDs:         make([]string, 0, 1+targetCount*2),
	}
	result.reservedIDs = append(result.reservedIDs, submissionID)
	for index := 0; index < targetCount; index++ {
		messageID, err := f.nextOpaqueID()
		if err != nil {
			return fakeLocalDeliveryCandidateIDs{}, err
		}
		version, err := f.nextOpaqueID()
		if err != nil {
			return fakeLocalDeliveryCandidateIDs{}, err
		}
		result.recipientMessageIDs[index] = messageID
		result.recipientVersions[index] = version
		result.reservedIDs = append(result.reservedIDs, messageID, version)
	}
	return result, nil
}

func fakeLocalDeliveryRecipientCandidates(
	sender EngineMessage,
	plan fakeLocalDeliveryPlan,
	ids fakeLocalDeliveryCandidateIDs,
	receivedAt time.Time,
) ([]fakeLocalDeliveryRecipientCandidate, error) {
	if len(plan.targets) != len(ids.recipientMessageIDs) || len(plan.targets) != len(ids.recipientVersions) {
		return nil, fakeEngineError(EngineErrorRejectedPolicy)
	}
	if len(plan.targets) == 0 {
		return nil, nil
	}
	if !validMailEventTime(receivedAt) {
		return nil, fakeEngineError(EngineErrorRejectedPolicy)
	}
	result := make([]fakeLocalDeliveryRecipientCandidate, 0, len(plan.targets))
	for index, target := range plan.targets {
		candidate, err := fakeLocalDeliveryRecipientCandidateForTarget(
			sender,
			target,
			ids.recipientMessageIDs[index],
			ids.recipientVersions[index],
			receivedAt,
			plan.calendar,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, candidate)
	}
	return result, nil
}

func fakeLocalDeliveryRecipientCandidateForTarget(
	sender EngineMessage,
	target fakeLocalDeliveryTarget,
	messageID, version string,
	receivedAt time.Time,
	calendar *fakeLocalDeliveryCalendar,
) (fakeLocalDeliveryRecipientCandidate, error) {
	if !validAccountID(target.accountID) || !validOpaqueID(target.inboxFolderID) || !validOpaqueID(messageID) || !validOpaqueID(version) || !validMailEventTime(receivedAt) {
		return fakeLocalDeliveryRecipientCandidate{}, fakeEngineError(EngineErrorRejectedPolicy)
	}
	message := cloneMessage(sender)
	message.Projection.AccountID = target.accountID
	message.Projection.MessageID = messageID
	message.Projection.Version = version
	message.Projection.BCC = nil
	message.Projection.FolderIDs = []string{target.inboxFolderID}
	message.Projection.Flags = EngineMessageFlags{}
	message.Projection.Origin = EngineProjectionOriginInboundOrUnknown
	message.Projection.AuthenticationClass = MailAuthenticationClassTrustedInternal
	instant := receivedAt.UTC()
	message.Projection.SentAt = nil
	message.Projection.ReceivedAt = &instant
	message.Projection.CalendarMethod = nil
	message.Projection.CalendarPartLocatorHash = nil

	var part *FakeCalendarPart
	if calendar != nil {
		locatorHash, err := fakeLocalDeliveryCalendarLocatorHash(target.accountID, messageID, version, *calendar)
		if err != nil {
			return fakeLocalDeliveryRecipientCandidate{}, err
		}
		method := calendar.method
		message.Projection.CalendarMethod = &method
		message.Projection.CalendarPartLocatorHash = append([]byte(nil), locatorHash[:]...)
		candidatePart := FakeCalendarPart{
			AccountID:   target.accountID,
			MessageID:   messageID,
			Version:     version,
			LocatorHash: locatorHash,
			Method:      method,
			ITIP:        append([]byte(nil), calendar.itip...),
		}
		if !validFakeCalendarPart(candidatePart) {
			return fakeLocalDeliveryRecipientCandidate{}, fakeEngineError(EngineErrorUnavailable)
		}
		part = &candidatePart
	}
	if err := validateEngineMessage(message); err != nil {
		return fakeLocalDeliveryRecipientCandidate{}, err
	}
	return fakeLocalDeliveryRecipientCandidate{message: message, part: part}, nil
}

func fakeLocalDeliveryCalendarLocatorHash(accountID, messageID, version string, calendar fakeLocalDeliveryCalendar) ([sha256.Size]byte, error) {
	deliveryMethod, deliveryMethodOK := fakeLocalDeliveryCalendarDeliveryMethod(calendar.method)
	if !deliveryMethodOK || !validAccountID(accountID) || !validOpaqueID(messageID) || !validOpaqueID(version) || !validCalendarPartMethod(calendar.method) || len(calendar.itip) == 0 || len(calendar.itip) > MaxCalendarPartBytes || calendar.bindingDigest == [sha256.Size]byte{} || calendar.contentDigest == [sha256.Size]byte{} || calendar.contentDigest != calendarMIMEPartContentHash(deliveryMethod, calendar.itip) {
		return [sha256.Size]byte{}, fakeEngineError(EngineErrorUnavailable)
	}
	hasher := sha256.New()
	_, _ = hasher.Write([]byte(fakeLocalDeliveryCalendarLocatorDomain))
	_, _ = hasher.Write([]byte{0, fakeLocalDeliveryCalendarLocatorVersion})
	fakeLocalDeliveryHashFrame(hasher, []byte(accountID))
	fakeLocalDeliveryHashFrame(hasher, []byte(messageID))
	fakeLocalDeliveryHashFrame(hasher, []byte(version))
	fakeLocalDeliveryHashFrame(hasher, []byte(calendar.method))
	fakeLocalDeliveryHashFrame(hasher, calendar.itip)
	fakeLocalDeliveryHashFrame(hasher, calendar.bindingDigest[:])
	fakeLocalDeliveryHashFrame(hasher, calendar.contentDigest[:])
	var result [sha256.Size]byte
	copy(result[:], hasher.Sum(nil))
	if calendarPartLocatorHashZero(result) {
		return [sha256.Size]byte{}, fakeEngineError(EngineErrorUnavailable)
	}
	return result, nil
}

func fakeLocalDeliveryHashFrame(hasher interface{ Write([]byte) (int, error) }, value []byte) {
	var length [4]byte
	binary.BigEndian.PutUint32(length[:], uint32(len(value)))
	_, _ = hasher.Write(length[:])
	_, _ = hasher.Write(value)
}

func fakeLocalDeliveryPlansEqual(left, right fakeLocalDeliveryPlan) bool {
	if left.senderSentFolderID != right.senderSentFolderID {
		return false
	}
	if len(left.targets) != len(right.targets) {
		return false
	}
	for index := range left.targets {
		if left.targets[index] != right.targets[index] {
			return false
		}
	}
	if left.calendar == nil || right.calendar == nil {
		return left.calendar == nil && right.calendar == nil
	}
	return left.calendar.method == right.calendar.method &&
		left.calendar.bindingDigest == right.calendar.bindingDigest &&
		left.calendar.contentDigest == right.calendar.contentDigest &&
		bytes.Equal(left.calendar.itip, right.calendar.itip)
}

func fakeCalendarMIMEPartCopy(part calendarMIMEPart) calendarMIMEPart {
	part.serialized = part.calendarBytes()
	return part
}
