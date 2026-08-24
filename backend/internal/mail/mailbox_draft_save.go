package mail

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

// mailboxDraftSavePrepared is the complete provider-neutral save intent. The
// persisted pending state intentionally carries only its canonical hash and
// content-free graph; the canonical content stays in this request scope until
// the fenced provider call completes.
type mailboxDraftSavePrepared struct {
	scope            mailboxDraftCreateScope
	draftID          uuid.UUID
	expectedRevision int64
	canonical        MailboxDraftCanonicalRequest
	pending          MailboxDraftPendingSave
	linkedSourceProof *MailboxDraftLinkedSourceProof
}

// SaveMailboxDraft saves one active, attachment-free draft under an exact
// revision precondition. A repeated pending intent only reconciles the
// persisted mutation; it never redispatches a provider write.
func (service *MailboxDraftService) SaveMailboxDraft(ctx context.Context, actorID, draftID uuid.UUID, input MailboxDraftSaveInput) (MailboxDraftView, error) {
	if !mailboxDraftCreateServiceValid(service) || ctx == nil || actorID == uuid.Nil || draftID == uuid.Nil || input.ExpectedRevision < 1 {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	if ctx.Err() != nil {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorUnavailable)
	}
	if len(input.Attachments) != 0 {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorUnprocessable)
	}

	owner, err := service.owners.ResolveOwner(ctx, actorID)
	if err != nil {
		return MailboxDraftView{}, mailboxDraftGetOwnerError(err)
	}
	scope, err := mailboxDraftCreateScopeFromOwner(actorID, owner)
	if err != nil {
		return MailboxDraftView{}, err
	}
	current, err := service.repository.GetMailboxDraft(ctx, MailboxDraftGetRequest{
		AccountID: scope.accountID,
		DraftID:   draftID,
	})
	if err != nil {
		return MailboxDraftView{}, mailboxDraftGetRepositoryError(err)
	}
	prepared, err := service.mailboxDraftSavePrepare(ctx, scope, draftID, input, current)
	if err != nil {
		return MailboxDraftView{}, err
	}
	mutationID, pending, err := mailboxDraftSaveCorrelation(current, prepared)
	if err != nil {
		return MailboxDraftView{}, err
	}

	begin, err := service.repository.BeginEngineCommand(ctx, MailboxEngineCommandBeginRequest{AccountID: prepared.scope.accountID})
	if err != nil {
		return MailboxDraftView{}, mailboxDraftCreateRepositoryError(err)
	}
	if !mailboxDraftCreateBeginMatches(begin, prepared.scope) {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorOutcomeUnknown)
	}
	if !begin.Claimed {
		return MailboxDraftView{}, mailboxDraftServiceBusyError(*begin.BusyUntil)
	}

	reservation, err := service.repository.ReserveMailboxDraftSave(ctx, MailboxDraftSaveRequest{
		Fence:            *begin.Fence,
		DraftID:          prepared.draftID,
		DraftClientID:    prepared.canonical.DraftClientID,
		ExpectedRevision: prepared.expectedRevision,
		PendingSave:      mailboxDraftSavePendingCopy(prepared.pending),
		LinkedSourceProof: mailboxDraftSaveLinkedSourceProofCopy(prepared.linkedSourceProof),
		MutationID:       mutationID,
	})
	if err != nil {
		return MailboxDraftView{}, mailboxDraftCreateRepositoryError(err)
	}

	switch reservation.Kind {
	case MailboxDraftReservationDispatch:
		if pending || !mailboxDraftSaveReservationFenceMatches(reservation, *begin.Fence) || !mailboxDraftSaveDispatchReservationValid(reservation, prepared, mutationID) {
			return MailboxDraftView{}, service.mailboxDraftSaveClaimedUnknown(ctx, *begin.Fence, prepared.draftID, mutationID)
		}
		draft, ok := mailboxDraftSaveEngineDraft(prepared, reservation)
		if !ok {
			return MailboxDraftView{}, service.mailboxDraftSaveClaimedUnknown(ctx, *begin.Fence, prepared.draftID, mutationID)
		}
		message, saveErr := service.engine.SaveDraft(ctx, draft)
		if saveErr != nil {
			return MailboxDraftView{}, service.mailboxDraftSaveClaimedUnknown(ctx, *begin.Fence, prepared.draftID, mutationID)
		}
		applied, ok := mailboxDraftSaveApplied(message, prepared, reservation)
		if !ok {
			return MailboxDraftView{}, service.mailboxDraftSaveClaimedUnknown(ctx, *begin.Fence, prepared.draftID, mutationID)
		}
		return service.mailboxDraftSaveFinishApplied(ctx, actorID, prepared, *begin.Fence, mutationID, reservation, applied)
	case MailboxDraftReservationReconcileRequired:
		if !pending || !mailboxDraftSaveReservationFenceMatches(reservation, *begin.Fence) || !mailboxDraftSaveReconcileReservationValid(reservation, prepared, mutationID) {
			return MailboxDraftView{}, service.mailboxDraftSaveClaimedUnknown(ctx, *begin.Fence, prepared.draftID, mutationID)
		}
		reconciliation, ok := mailboxDraftSaveReconciliation(prepared, reservation)
		if !ok {
			return MailboxDraftView{}, service.mailboxDraftSaveClaimedUnknown(ctx, *begin.Fence, prepared.draftID, mutationID)
		}
		result, reconcileErr := service.engine.ReconcileDraft(ctx, reconciliation)
		if reconcileErr != nil {
			return MailboxDraftView{}, service.mailboxDraftSaveClaimedUnknown(ctx, *begin.Fence, prepared.draftID, mutationID)
		}
		appliedResult, applied := result.(EngineDraftApplied)
		if !applied {
			return MailboxDraftView{}, service.mailboxDraftSaveClaimedUnknown(ctx, *begin.Fence, prepared.draftID, mutationID)
		}
		finish, ok := mailboxDraftSaveApplied(appliedResult.Message, prepared, reservation)
		if !ok {
			return MailboxDraftView{}, service.mailboxDraftSaveClaimedUnknown(ctx, *begin.Fence, prepared.draftID, mutationID)
		}
		return service.mailboxDraftSaveFinishApplied(ctx, actorID, prepared, *begin.Fence, mutationID, reservation, finish)
	default:
		return MailboxDraftView{}, service.mailboxDraftSaveClaimedUnknown(ctx, *begin.Fence, prepared.draftID, mutationID)
	}
}

func (service *MailboxDraftService) mailboxDraftSavePrepare(ctx context.Context, scope mailboxDraftCreateScope, draftID uuid.UUID, input MailboxDraftSaveInput, current MailboxDraftGetResult) (mailboxDraftSavePrepared, error) {
	if !mailboxDraftCreateServiceValid(service) || ctx == nil || !mailboxDraftCreateScopeValid(scope) || draftID == uuid.Nil || input.ExpectedRevision < 1 || input.DraftClientID == uuid.Nil || input.FromIdentityID == uuid.Nil || len(input.Attachments) != 0 {
		return mailboxDraftSavePrepared{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	if current.Found != (current.Draft != nil) {
		return mailboxDraftSavePrepared{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}

	inReplyTo := input.InReplyTo
	references := input.References
	replyToMessageID := mailboxDraftGetUUIDPointerCopy(input.ReplyToMessageID)
	forwardMessageID := mailboxDraftGetUUIDPointerCopy(input.ForwardMessageID)
	var linkedSourceProof *MailboxDraftLinkedSourceProof
	if current.Found && current.Draft != nil && mailboxDraftSaveLinkedEffectiveStateValid(current.Draft.Confirmed) {
		record := *current.Draft
		if record.ID != draftID || record.AccountID != scope.accountID || record.DraftClientID != input.DraftClientID {
			return mailboxDraftSavePrepared{}, mailboxDraftServiceError(MailboxDraftServiceErrorConflict)
		}
		if record.Revision != input.ExpectedRevision {
			return mailboxDraftSavePrepared{}, mailboxDraftServiceError(MailboxDraftServiceErrorPreconditionFailed)
		}
		if !mailboxDraftSaveLinkedThreadingInputEmpty(input) || !mailboxDraftSaveLinkedEffectiveStateValid(record.Confirmed) {
			return mailboxDraftSavePrepared{}, mailboxDraftServiceError(MailboxDraftServiceErrorUnprocessable)
		}
		switch record.Status {
		case MailboxDraftStatusSaving, MailboxDraftStatusUnknown:
			if record.PendingSave == nil {
				return mailboxDraftSavePrepared{}, mailboxDraftServiceError(MailboxDraftServiceErrorConflict)
			}
			prepared := mailboxDraftSavePrepared{
				scope:            scope,
				draftID:          draftID,
				expectedRevision: input.ExpectedRevision,
				canonical:        MailboxDraftCanonicalRequest{DraftClientID: input.DraftClientID},
				pending:          mailboxDraftSavePendingCopy(*record.PendingSave),
			}
			if !mailboxDraftSavePendingRecordMatches(record, prepared) || !mailboxDraftSaveLinkedEffectiveStateValid(record.PendingSave.EffectiveState) || !mailboxDraftSaveLinkedSourceIDsMatch(record.Confirmed, record.PendingSave.EffectiveState) {
				return mailboxDraftSavePrepared{}, mailboxDraftServiceError(MailboxDraftServiceErrorConflict)
			}
			return prepared, nil
		case MailboxDraftStatusActive:
			if !mailboxDraftGetRecordValid(&record, scope.accountID, draftID) || mailboxDraftServiceDependencyNil(service.messages) {
				return mailboxDraftSavePrepared{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
			}
			sourceID, replyLink, ok := mailboxDraftSaveLinkedSource(record.Confirmed)
			if !ok {
				return mailboxDraftSavePrepared{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
			}
			messageScope := MailboxMessageReadScope{ActorID: scope.actorID, OrganizationID: scope.organizationID, AccountID: scope.accountID, EngineAccountID: scope.engineAccountID}
			snapshot, err := service.messages.ResolveMailboxMessage(ctx, messageScope, sourceID)
			if err != nil {
				return mailboxDraftSavePrepared{}, mailboxReplyDraftMessageError(err)
			}
			if !mailboxReplyDraftSnapshotMatches(snapshot, messageScope, sourceID) {
				return mailboxDraftSavePrepared{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
			}
			if !mailboxReplyDraftSourceEligible(snapshot) {
				return mailboxDraftSavePrepared{}, mailboxDraftServiceError(MailboxDraftServiceErrorNotFound)
			}
			if replyLink {
				derivedReferences, err := mailboxReplyDraftReferences(snapshot)
				if err != nil {
					return mailboxDraftSavePrepared{}, err
				}
				inReplyTo = snapshot.internetMessageID
				references = derivedReferences
				replyToMessageID = &sourceID
				forwardMessageID = nil
			} else {
				inReplyTo = ""
				references = nil
				replyToMessageID = nil
				forwardMessageID = &sourceID
			}
			linkedSourceProof = &MailboxDraftLinkedSourceProof{MessageID: sourceID, EngineMessageID: snapshot.engineMessageID, EngineVersion: snapshot.engineVersion}
		default:
			return mailboxDraftSavePrepared{}, mailboxDraftServiceError(MailboxDraftServiceErrorConflict)
		}
	}

	content, err := NormalizeMailboxContent(input.Content)
	if err != nil {
		return mailboxDraftSavePrepared{}, mailboxDraftCreateContentError(err)
	}
	identity, err := service.identities.ResolveMailboxDraftIdentityProof(ctx, mailboxDraftIdentityProofRequest{
		AccountID:  scope.accountID,
		IdentityID: input.FromIdentityID,
	})
	if err != nil {
		return mailboxDraftSavePrepared{}, mailboxDraftCreateRepositoryError(err)
	}
	if identity.accountID != scope.accountID || identity.identityID != input.FromIdentityID {
		return mailboxDraftSavePrepared{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	attachmentProof, err := service.attachments.ResolveMailboxDraftAttachmentProof(ctx, mailboxDraftAttachmentProofRequest{
		AccountID:     scope.accountID,
		DraftClientID: input.DraftClientID,
		DraftID:       draftID,
		AttachmentIDs: []uuid.UUID{},
	})
	if err != nil {
		return mailboxDraftSavePrepared{}, mailboxDraftCreateRepositoryError(err)
	}
	if !mailboxDraftCreateAttachmentProofMatches(attachmentProof, scope.accountID, input.DraftClientID, draftID, []uuid.UUID{}) {
		return mailboxDraftSavePrepared{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}

	canonical, err := CanonicalizeMailboxDraftRequest(MailboxDraftRequestHashInput{
		AccountID:        scope.accountID,
		DraftClientID:    input.DraftClientID,
		FromIdentityID:   input.FromIdentityID,
		From:             identity.from,
		To:               input.To,
		CC:               input.CC,
		BCC:              input.BCC,
		ReplyTo:          input.ReplyTo,
		Subject:          input.Subject,
		InReplyTo:        inReplyTo,
		References:       references,
		Priority:         input.Priority,
		ReplyToMessageID: replyToMessageID,
		ForwardMessageID: forwardMessageID,
		Attachments:      attachmentProof.attachments,
		Body:             content,
	})
	if err != nil || len(canonical.Attachments) != 0 || !mailboxDraftCreateHashPresent(canonical.RequestHash) {
		return mailboxDraftSavePrepared{}, mailboxDraftServiceError(MailboxDraftServiceErrorUnprocessable)
	}
	effective := MailboxDraftEffectiveState{
		FromIdentityID:   canonical.FromIdentityID,
		ReplyToMessageID: mailboxDraftGetUUIDPointerCopy(canonical.ReplyToMessageID),
		ForwardMessageID: mailboxDraftGetUUIDPointerCopy(canonical.ForwardMessageID),
		AttachmentIDs:    []uuid.UUID{},
	}
	if !mailboxDraftGetEffectiveStateValid(effective) {
		return mailboxDraftSavePrepared{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	return mailboxDraftSavePrepared{
		scope:            scope,
		draftID:          draftID,
		expectedRevision: input.ExpectedRevision,
		canonical:        canonical,
		pending: MailboxDraftPendingSave{
			EffectiveState: mailboxDraftCreateEffectiveStateCopy(effective),
			RequestHash:    canonical.RequestHash,
		},
		linkedSourceProof: mailboxDraftSaveLinkedSourceProofCopy(linkedSourceProof),
	}, nil
}

func mailboxDraftSaveLinkedThreadingInputEmpty(input MailboxDraftSaveInput) bool {
	return input.InReplyTo == "" && len(input.References) == 0 && input.ReplyToMessageID == nil && input.ForwardMessageID == nil
}

func mailboxDraftSaveLinkedEffectiveStateValid(state MailboxDraftEffectiveState) bool {
	return mailboxDraftGetEffectiveStateValid(state) && len(state.AttachmentIDs) == 0 &&
		(state.ReplyToMessageID != nil) != (state.ForwardMessageID != nil)
}

func mailboxDraftSaveLinkedSource(state MailboxDraftEffectiveState) (uuid.UUID, bool, bool) {
	if !mailboxDraftSaveLinkedEffectiveStateValid(state) {
		return uuid.Nil, false, false
	}
	if state.ReplyToMessageID != nil {
		return *state.ReplyToMessageID, true, true
	}
	return *state.ForwardMessageID, false, true
}

func mailboxDraftSaveLinkedSourceIDsMatch(confirmed, pending MailboxDraftEffectiveState) bool {
	confirmedID, confirmedReply, confirmedOK := mailboxDraftSaveLinkedSource(confirmed)
	pendingID, pendingReply, pendingOK := mailboxDraftSaveLinkedSource(pending)
	return confirmedOK && pendingOK && confirmedReply == pendingReply && confirmedID == pendingID
}

func mailboxDraftSaveLinkedSourceProofCopy(value *MailboxDraftLinkedSourceProof) *MailboxDraftLinkedSourceProof {
	if value == nil {
		return nil
	}
	return &MailboxDraftLinkedSourceProof{MessageID: value.MessageID, EngineMessageID: strings.Clone(value.EngineMessageID), EngineVersion: strings.Clone(value.EngineVersion)}
}

// mailboxDraftSaveCorrelation determines whether a fresh mutation may be
// reserved or whether the same durable intent must only be reconciled.
func mailboxDraftSaveCorrelation(result MailboxDraftGetResult, prepared mailboxDraftSavePrepared) (uuid.UUID, bool, error) {
	if result.Found != (result.Draft != nil) {
		return uuid.Nil, false, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	if !result.Found {
		return uuid.Nil, false, mailboxDraftServiceError(MailboxDraftServiceErrorNotFound)
	}
	record := result.Draft
	if record == nil || record.ID != prepared.draftID || record.AccountID != prepared.scope.accountID || record.DraftClientID != prepared.canonical.DraftClientID {
		return uuid.Nil, false, mailboxDraftServiceError(MailboxDraftServiceErrorConflict)
	}
	if record.Revision != prepared.expectedRevision {
		return uuid.Nil, false, mailboxDraftServiceError(MailboxDraftServiceErrorPreconditionFailed)
	}
	if len(record.Confirmed.AttachmentIDs) != 0 {
		return uuid.Nil, false, mailboxDraftServiceError(MailboxDraftServiceErrorUnprocessable)
	}
	switch record.Status {
	case MailboxDraftStatusActive:
		if !mailboxDraftGetRecordValid(record, prepared.scope.accountID, prepared.draftID) {
			return uuid.Nil, false, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
		}
		mutationID := uuid.New()
		if mutationID == uuid.Nil {
			return uuid.Nil, false, mailboxDraftServiceError(MailboxDraftServiceErrorUnavailable)
		}
		return mutationID, false, nil
	case MailboxDraftStatusSaving, MailboxDraftStatusUnknown:
		if !mailboxDraftSavePendingRecordMatches(*record, prepared) {
			return uuid.Nil, false, mailboxDraftServiceError(MailboxDraftServiceErrorConflict)
		}
		return *record.LastMutationID, true, nil
	default:
		return uuid.Nil, false, mailboxDraftServiceError(MailboxDraftServiceErrorConflict)
	}
}

func mailboxDraftSavePendingRecordMatches(record MailboxDraftRecord, prepared mailboxDraftSavePrepared) bool {
	return record.ID == prepared.draftID &&
		record.AccountID == prepared.scope.accountID &&
		record.DraftClientID == prepared.canonical.DraftClientID &&
		record.Revision == prepared.expectedRevision &&
		(record.Status == MailboxDraftStatusSaving || record.Status == MailboxDraftStatusUnknown) &&
		syncValidOpaque(record.EngineDraftKey, MaxEngineOpaqueIDBytes) &&
		record.EngineMessageID != nil && syncValidOpaque(*record.EngineMessageID, MaxEngineOpaqueIDBytes) &&
		record.EngineVersion != nil && syncValidOpaque(*record.EngineVersion, MaxEngineOpaqueIDBytes) &&
		mailboxDraftGetEffectiveStateValid(record.Confirmed) && len(record.Confirmed.AttachmentIDs) == 0 &&
		record.PendingSave != nil &&
		mailboxDraftCreateEffectiveStatesEqual(record.PendingSave.EffectiveState, prepared.pending.EffectiveState) &&
		record.PendingSave.RequestHash == prepared.pending.RequestHash &&
		record.LastMutationID != nil && *record.LastMutationID != uuid.Nil &&
		record.PendingMutationKind != nil && *record.PendingMutationKind == MailboxDraftMutationSave &&
		record.PendingCreateAttachmentPhase == nil &&
		record.PendingCreateUploadBlobID == nil &&
		record.PendingCreateUploadSizeBytes == nil &&
		record.PendingCreateUploadAttempt == 0 &&
		record.PendingCreateUploadRetryAt == nil &&
		record.PendingCreateUploadExhaustedAt == nil &&
		mailboxDraftGetUUIDPointerValid(record.ConfirmedMutationID)
}

func mailboxDraftSaveDispatchReservationValid(reservation MailboxDraftReservation, prepared mailboxDraftSavePrepared, mutationID uuid.UUID) bool {
	return reservation.Kind == MailboxDraftReservationDispatch &&
		mailboxDraftSaveReservationValid(reservation, prepared, mutationID, MailboxDraftWorkSave, MailboxDraftStatusSaving)
}

func mailboxDraftSaveReconcileReservationValid(reservation MailboxDraftReservation, prepared mailboxDraftSavePrepared, mutationID uuid.UUID) bool {
	return reservation.Kind == MailboxDraftReservationReconcileRequired &&
		(reservation.Draft.Status == MailboxDraftStatusSaving || reservation.Draft.Status == MailboxDraftStatusUnknown) &&
		mailboxDraftSaveReservationValid(reservation, prepared, mutationID, MailboxDraftWorkReconcileSave, reservation.Draft.Status)
}

func mailboxDraftSaveReservationValid(reservation MailboxDraftReservation, prepared mailboxDraftSavePrepared, mutationID uuid.UUID, workKind MailboxDraftWorkKind, status MailboxDraftStatus) bool {
	if reservation.Work == nil || reservation.Fence == nil || mutationID == uuid.Nil || status != reservation.Draft.Status || !mailboxEntEngineCommandFenceValid(*reservation.Fence) || reservation.Fence.AccountID != prepared.scope.accountID || !mailboxDraftSavePendingRecordMatches(reservation.Draft, prepared) || reservation.Work.Kind != workKind || reservation.Work.MutationID != mutationID || reservation.Work.EngineDraftKey != reservation.Draft.EngineDraftKey || reservation.Work.Revision != prepared.expectedRevision || reservation.Work.RequestHash != prepared.pending.RequestHash || !mailboxDraftCreateEffectiveStatesEqual(reservation.Work.EffectiveState, prepared.pending.EffectiveState) || reservation.Work.EngineMessageID == nil || reservation.Work.EngineVersion == nil || reservation.Draft.EngineMessageID == nil || reservation.Draft.EngineVersion == nil || *reservation.Work.EngineMessageID != *reservation.Draft.EngineMessageID || *reservation.Work.EngineVersion != *reservation.Draft.EngineVersion || reservation.Work.PendingCreateAttachmentPhase != nil || reservation.Work.PendingCreateUploadBlobID != nil || reservation.Work.PendingCreateUploadSizeBytes != nil || reservation.Work.PendingCreateUploadAttempt != 0 || reservation.Work.PendingCreateUploadRetryAt != nil || reservation.Work.PendingCreateUploadExhaustedAt != nil {
		return false
	}
	return reservation.Draft.LastMutationID != nil && *reservation.Draft.LastMutationID == mutationID && reservation.Draft.PendingMutationKind != nil && *reservation.Draft.PendingMutationKind == MailboxDraftMutationSave && reservation.Work.ConfirmedMutationID != nil && reservation.Draft.ConfirmedMutationID != nil && *reservation.Work.ConfirmedMutationID == *reservation.Draft.ConfirmedMutationID
}

func mailboxDraftSaveReservationFenceMatches(reservation MailboxDraftReservation, expected MailboxEngineCommandFence) bool {
	return reservation.Fence != nil &&
		reservation.Fence.AccountID == expected.AccountID &&
		reservation.Fence.LeaseID == expected.LeaseID &&
		reservation.Fence.MutationEpoch == expected.MutationEpoch &&
		reservation.Fence.LeaseUntil.Equal(expected.LeaseUntil)
}

func mailboxDraftSaveEngineDraft(prepared mailboxDraftSavePrepared, reservation MailboxDraftReservation) (EngineDraft, bool) {
	if reservation.Work == nil || !mailboxDraftSaveDispatchReservationValid(reservation, prepared, reservation.Work.MutationID) {
		return EngineDraft{}, false
	}
	content, ok := mailboxDraftSaveEngineContent(prepared.canonical)
	if !ok {
		return EngineDraft{}, false
	}
	draft := EngineDraft{
		AccountID:       prepared.scope.engineAccountID,
		DraftKey:        strings.Clone(reservation.Work.EngineDraftKey),
		MessageID:       strings.Clone(*reservation.Work.EngineMessageID),
		PreviousVersion: strings.Clone(*reservation.Work.EngineVersion),
		MutationID:      reservation.Work.MutationID.String(),
		MaxContentBytes: MaxEngineMessageBytes,
		Message:         content,
	}
	if validateDraftRequest(draft) != nil || len(draft.Message.Attachments) != 0 {
		return EngineDraft{}, false
	}
	return draft, true
}

func mailboxDraftSaveEngineContent(canonical MailboxDraftCanonicalRequest) (EngineMessageContent, bool) {
	content := EngineMessageContent{
		From:       mailboxDraftGetAddressCopy(canonical.From),
		To:         mailboxDraftGetAddressesCopy(canonical.To),
		CC:         mailboxDraftGetAddressesCopy(canonical.CC),
		BCC:        mailboxDraftGetAddressesCopy(canonical.BCC),
		ReplyTo:    mailboxDraftGetAddressesCopy(canonical.ReplyTo),
		Subject:    strings.Clone(canonical.Subject),
		InReplyTo:  strings.Clone(canonical.InReplyTo),
		References: mailboxDraftGetStringsCopy(canonical.References),
		Priority:   canonical.Priority,
	}
	switch canonical.Body.SourceFormat {
	case MailboxContentFormatPlain:
		content.Format = EngineBodyPlain
		content.PlainBody = strings.Clone(canonical.Body.PlainBody)
	case MailboxContentFormatRich:
		content.Format = EngineBodyRich
		content.PlainBody = strings.Clone(canonical.Body.PlainBody)
		content.RichBody = strings.Clone(canonical.Body.SanitizedHTML)
	default:
		return EngineMessageContent{}, false
	}
	return content, true
}

func mailboxDraftSaveReconciliation(prepared mailboxDraftSavePrepared, reservation MailboxDraftReservation) (EngineDraftReconciliation, bool) {
	if reservation.Work == nil || !mailboxDraftSaveReconcileReservationValid(reservation, prepared, reservation.Work.MutationID) {
		return EngineDraftReconciliation{}, false
	}
	request := EngineDraftReconciliation{
		AccountID:           prepared.scope.engineAccountID,
		DraftKey:            strings.Clone(reservation.Work.EngineDraftKey),
		MessageID:           strings.Clone(*reservation.Work.EngineMessageID),
		PreviousVersion:     strings.Clone(*reservation.Work.EngineVersion),
		MutationID:          reservation.Work.MutationID.String(),
		ConfirmedMutationID: reservation.Work.ConfirmedMutationID.String(),
		MaxContentBytes:     MaxEngineMessageBytes,
	}
	if validateDraftReconciliation(request) != nil {
		return EngineDraftReconciliation{}, false
	}
	return request, true
}

func mailboxDraftSaveApplied(message EngineMessage, prepared mailboxDraftSavePrepared, reservation MailboxDraftReservation) (MailboxDraftApplied, bool) {
	if reservation.Work == nil || reservation.Work.EngineMessageID == nil || reservation.Work.EngineVersion == nil {
		return MailboxDraftApplied{}, false
	}
	applied, ok := mailboxDraftCreateApplied(message, mailboxDraftCreatePrepared{scope: prepared.scope}, reservation)
	if !ok || applied.EngineMessageID == *reservation.Work.EngineMessageID && applied.EngineVersion == *reservation.Work.EngineVersion {
		return MailboxDraftApplied{}, false
	}
	return applied, true
}

func (service *MailboxDraftService) mailboxDraftSaveFinishApplied(ctx context.Context, actorID uuid.UUID, prepared mailboxDraftSavePrepared, fence MailboxEngineCommandFence, mutationID uuid.UUID, reservation MailboxDraftReservation, applied MailboxDraftApplied) (MailboxDraftView, error) {
	if reservation.Work == nil || reservation.Fence == nil || !mailboxDraftSaveReservationFenceMatches(reservation, fence) || reservation.Work.MutationID != mutationID || (reservation.Work.Kind != MailboxDraftWorkSave && reservation.Work.Kind != MailboxDraftWorkReconcileSave) || !mailboxDraftSaveReservationValid(reservation, prepared, mutationID, reservation.Work.Kind, reservation.Draft.Status) {
		return MailboxDraftView{}, service.mailboxDraftSaveClaimedUnknown(ctx, fence, prepared.draftID, mutationID)
	}
	finished, err := service.repository.FinishEngineCommand(ctx, MailboxDraftFinishRequest{
		Fence:      fence,
		DraftID:    prepared.draftID,
		MutationID: mutationID,
		Kind:       MailboxDraftMutationSave,
		Outcome:    applied,
	})
	nextRevision, nextOK := mailboxEntNextEpoch(prepared.expectedRevision)
	if err != nil || !nextOK || !mailboxDraftGetRecordValid(&finished.Draft, prepared.scope.accountID, prepared.draftID) || finished.Draft.Revision != nextRevision || !mailboxDraftCreateEffectiveStatesEqual(finished.Draft.Confirmed, prepared.pending.EffectiveState) || finished.Draft.EngineMessageID == nil || finished.Draft.EngineVersion == nil || *finished.Draft.EngineMessageID != applied.EngineMessageID || *finished.Draft.EngineVersion != applied.EngineVersion {
		return MailboxDraftView{}, service.mailboxDraftSaveClaimedUnknown(ctx, fence, prepared.draftID, mutationID)
	}
	return service.GetMailboxDraft(ctx, actorID, prepared.draftID)
}

func (service *MailboxDraftService) mailboxDraftSaveClaimedUnknown(ctx context.Context, fence MailboxEngineCommandFence, draftID, mutationID uuid.UUID) error {
	if service == nil || ctx == nil || !mailboxEntEngineCommandFenceValid(fence) || draftID == uuid.Nil || mutationID == uuid.Nil {
		return mailboxDraftServiceError(MailboxDraftServiceErrorOutcomeUnknown)
	}
	_, _ = service.repository.FinishEngineCommand(ctx, MailboxDraftFinishRequest{
		Fence:      fence,
		DraftID:    draftID,
		MutationID: mutationID,
		Kind:       MailboxDraftMutationSave,
		Outcome:    MailboxDraftUnknown{},
	})
	return mailboxDraftServiceError(MailboxDraftServiceErrorOutcomeUnknown)
}

func mailboxDraftSavePendingCopy(value MailboxDraftPendingSave) MailboxDraftPendingSave {
	return MailboxDraftPendingSave{
		EffectiveState: mailboxDraftCreateEffectiveStateCopy(value.EffectiveState),
		RequestHash:    value.RequestHash,
	}
}
