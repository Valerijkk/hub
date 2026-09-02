package mail

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailattachment"
	"github.com/valerii/hub/backend/ent/maildraft"
	"github.com/valerii/hub/backend/ent/mailmessage"
	"github.com/valerii/hub/backend/ent/mailobject"
)

// ReserveMailboxDraftSave persists a complete desired D graph while retaining confirmed O.
func (repository *MailboxEntRepository) ReserveMailboxDraftSave(ctx context.Context, request MailboxDraftSaveRequest) (MailboxDraftReservation, error) {
	if repository == nil || repository.client == nil || ctx == nil {
		return MailboxDraftReservation{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	if kind := mailboxEntDraftReservationSaveRequestKind(request); kind != "" {
		return MailboxDraftReservation{}, NewMailboxDraftRepositoryError(kind)
	}

	var result MailboxDraftReservation
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, databaseWallClock, err := mailboxEntDraftCreateLockAccount(ctx, transaction, request.Fence)
		if err != nil {
			return err
		}
		if kind := mailboxEntDraftCreateAccountKind(account); kind != "" {
			return mailboxEntDraftReservationReleaseError(ctx, account, request.Fence, databaseWallClock, kind)
		}
		draft, kind := mailboxEntDraftReservationLockDraft(ctx, transaction, request.Fence.AccountID, request.DraftID)
		if kind != "" {
			return mailboxEntDraftReservationReleaseError(ctx, account, request.Fence, databaseWallClock, kind)
		}
		switch MailboxDraftStatus(draft.Status) {
		case MailboxDraftStatusActive:
			if err := mailboxEntDraftCreateLockConfirmedIdentity(ctx, transaction, request.Fence.AccountID, request.PendingSave.EffectiveState.FromIdentityID); err != nil {
				return mailboxEntDraftReservationReleaseForError(ctx, account, request.Fence, databaseWallClock, err)
			}
			if err := mailboxEntDraftCreateLockReferences(ctx, transaction, request.Fence.AccountID, request.PendingSave.EffectiveState); err != nil {
				return mailboxEntDraftReservationReleaseForError(ctx, account, request.Fence, databaseWallClock, err)
			}
			if err := mailboxEntDraftReservationLockLinkedSource(ctx, transaction, request.Fence.AccountID, request.PendingSave.EffectiveState, request.LinkedSourceProof); err != nil {
				return mailboxEntDraftReservationReleaseForError(ctx, account, request.Fence, databaseWallClock, err)
			}
		case MailboxDraftStatusSaving, MailboxDraftStatusUnknown:
			if request.LinkedSourceProof != nil {
				return mailboxEntDraftReservationReleaseError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
			}
		default:
			return mailboxEntDraftReservationReleaseError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
		}
		attachments, kind := mailboxEntDraftReservationLockSaveAttachments(ctx, transaction, draft, request)
		if kind != "" {
			return mailboxEntDraftReservationReleaseError(ctx, account, request.Fence, databaseWallClock, kind)
		}
		record, ok := mailboxEntDraftReadRecord(draft, attachments)
		if !ok {
			return mailboxEntDraftReservationReleaseError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
		}

		switch record.Status {
		case MailboxDraftStatusActive:
			if record.Revision != request.ExpectedRevision || record.DraftClientID != request.DraftClientID || record.PendingSave != nil || !mailboxEntDraftReservationLinkedStatesMatch(record.Confirmed, request.PendingSave.EffectiveState, request.LinkedSourceProof) {
				return mailboxEntDraftReservationReleaseError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
			}
			updated, err := mailboxEntDraftReservationSaveTransition(ctx, draft, request, databaseWallClock)
			if err != nil {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
			}
			updatedRecord, ok := mailboxEntDraftReadRecord(updated, mailboxEntDraftAttachmentState{ConfirmedAttachmentIDs: append([]uuid.UUID(nil), record.Confirmed.AttachmentIDs...), PendingAttachmentIDs: append([]uuid.UUID(nil), request.PendingSave.EffectiveState.AttachmentIDs...)})
			if !ok {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
			}
			work := mailboxEntDraftCreateWork(updatedRecord, MailboxDraftWorkSave)
			if work == nil {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
			}
			fence := mailboxEntDraftCreateFenceCopy(request.Fence)
			result = MailboxDraftReservation{Kind: MailboxDraftReservationDispatch, Draft: updatedRecord, Work: work, Fence: &fence}
			return nil
		case MailboxDraftStatusSaving, MailboxDraftStatusUnknown:
			if record.PendingMutationKind == nil || *record.PendingMutationKind != MailboxDraftMutationSave || record.LastMutationID == nil || *record.LastMutationID != request.MutationID || record.Revision != request.ExpectedRevision || record.DraftClientID != request.DraftClientID || !mailboxEntDraftPendingSavesEqual(record.PendingSave, &request.PendingSave) {
				return mailboxEntDraftReservationReleaseError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
			}
			work := mailboxEntDraftCreateWork(record, MailboxDraftWorkReconcileSave)
			if work == nil {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
			}
			fence := mailboxEntDraftCreateFenceCopy(request.Fence)
			result = MailboxDraftReservation{Kind: MailboxDraftReservationReconcileRequired, Draft: mailboxEntDraftRecordCopy(record), Work: work, Fence: &fence}
			return nil
		default:
			return mailboxEntDraftReservationReleaseError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
		}
	})
	if err != nil {
		return MailboxDraftReservation{}, mailboxEntDraftRepositoryError(err)
	}
	return result, nil
}

// ReserveMailboxDraftDelete rejects every pending D graph before reserving delete work.
func (repository *MailboxEntRepository) ReserveMailboxDraftDelete(ctx context.Context, request MailboxDraftDeleteRequest) (MailboxDraftReservation, error) {
	if repository == nil || repository.client == nil || ctx == nil {
		return MailboxDraftReservation{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	if kind := mailboxEntDraftReservationDeleteRequestKind(request); kind != "" {
		return MailboxDraftReservation{}, NewMailboxDraftRepositoryError(kind)
	}

	var result MailboxDraftReservation
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, databaseWallClock, err := mailboxEntDraftCreateLockAccount(ctx, transaction, request.Fence)
		if err != nil {
			return err
		}
		if kind := mailboxEntDraftCreateAccountKind(account); kind != "" {
			return mailboxEntDraftReservationReleaseError(ctx, account, request.Fence, databaseWallClock, kind)
		}
		draft, kind := mailboxEntDraftReservationLockDraft(ctx, transaction, request.Fence.AccountID, request.DraftID)
		if kind != "" {
			return mailboxEntDraftReservationReleaseError(ctx, account, request.Fence, databaseWallClock, kind)
		}
		attachments, kind := mailboxEntDraftReadAttachmentState(ctx, transaction, draft, MailboxDraftRepositoryErrorStateConflict)
		if kind != "" {
			return mailboxEntDraftReservationReleaseError(ctx, account, request.Fence, databaseWallClock, kind)
		}
		record, ok := mailboxEntDraftReadRecord(draft, attachments)
		if !ok {
			return mailboxEntDraftReservationReleaseError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
		}
		switch record.Status {
		case MailboxDraftStatusActive:
			if record.Revision != request.ExpectedRevision || record.PendingSave != nil {
				return mailboxEntDraftReservationReleaseError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
			}
			updated, err := mailboxEntDraftReservationDeleteTransition(ctx, draft, request, databaseWallClock)
			if err != nil {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
			}
			updatedRecord, ok := mailboxEntDraftReadRecord(updated, attachments)
			if !ok {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
			}
			work := mailboxEntDraftCreateWork(updatedRecord, MailboxDraftWorkDelete)
			if work == nil {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
			}
			fence := mailboxEntDraftCreateFenceCopy(request.Fence)
			result = MailboxDraftReservation{Kind: MailboxDraftReservationDispatch, Draft: updatedRecord, Work: work, Fence: &fence}
			return nil
		case MailboxDraftStatusDeleting, MailboxDraftStatusUnknown:
			if record.PendingSave != nil || record.PendingMutationKind == nil || *record.PendingMutationKind != MailboxDraftMutationDelete || record.LastMutationID == nil || *record.LastMutationID != request.MutationID || record.Revision != request.ExpectedRevision {
				return mailboxEntDraftReservationReleaseError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
			}
			work := mailboxEntDraftCreateWork(record, MailboxDraftWorkReconcileDelete)
			if work == nil {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
			}
			fence := mailboxEntDraftCreateFenceCopy(request.Fence)
			result = MailboxDraftReservation{Kind: MailboxDraftReservationReconcileRequired, Draft: mailboxEntDraftRecordCopy(record), Work: work, Fence: &fence}
			return nil
		case MailboxDraftStatusDeleted:
			if record.Revision != request.ExpectedRevision || record.PendingSave != nil || len(record.Confirmed.AttachmentIDs) != 0 || record.EngineMessageID == nil || record.EngineVersion == nil {
				return mailboxEntDraftReservationReleaseError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
			}
			if err := mailboxEntDraftCreateRelease(ctx, account, request.Fence, databaseWallClock); err != nil {
				return err
			}
			result = MailboxDraftReservation{Kind: MailboxDraftReservationReplay, Draft: mailboxEntDraftRecordCopy(record)}
			return nil
		default:
			return mailboxEntDraftReservationReleaseError(ctx, account, request.Fence, databaseWallClock, MailboxDraftRepositoryErrorStateConflict)
		}
	})
	if err != nil {
		return MailboxDraftReservation{}, mailboxEntDraftRepositoryError(err)
	}
	return result, nil
}

// AbortMailboxDraftReservation abandons an exact pending draft mutation without engine I/O.
func (repository *MailboxEntRepository) AbortMailboxDraftReservation(ctx context.Context, request MailboxDraftAbortRequest) (MailboxDraftAbortResult, error) {
	if repository == nil || repository.client == nil || ctx == nil {
		return MailboxDraftAbortResult{}, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	if kind := mailboxEntDraftAbortRequestKind(request); kind != "" {
		return MailboxDraftAbortResult{}, NewMailboxDraftRepositoryError(kind)
	}

	var result MailboxDraftAbortResult
	var postCommitErr error
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, databaseWallClock, err := mailboxEntDraftCreateLockAccount(ctx, transaction, request.Fence)
		if err != nil {
			return err
		}
		if kind := mailboxEntDraftCreateAccountKind(account); kind != "" {
			return NewMailboxDraftRepositoryError(kind)
		}
		draft, releaseFence, kind := mailboxEntDraftAbortLockDraft(ctx, transaction, request.Fence.AccountID, request.DraftID)
		if kind != "" {
			if releaseFence {
				if err := mailboxEntDraftCreateRelease(ctx, account, request.Fence, databaseWallClock); err != nil {
					return err
				}
				postCommitErr = NewMailboxDraftRepositoryError(kind)
				return nil
			}
			return NewMailboxDraftRepositoryError(kind)
		}
		attachments, kind := mailboxEntDraftReadAttachmentState(ctx, transaction, draft, MailboxDraftRepositoryErrorStateConflict)
		if kind != "" {
			return NewMailboxDraftRepositoryError(kind)
		}
		record, ok := mailboxEntDraftReadRecord(draft, attachments)
		if !ok {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorStateConflict)
		}

		switch record.Status {
		case MailboxDraftStatusSaving:
			if record.LastMutationID == nil || *record.LastMutationID != request.MutationID || record.PendingMutationKind == nil {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorStateConflict)
			}
			switch *record.PendingMutationKind {
			case MailboxDraftMutationCreate:
				if record.Revision != 1 || record.EngineMessageID != nil || record.EngineVersion != nil || record.PendingSave != nil || len(attachments.PendingAttachmentIDs) != 0 {
					return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorStateConflict)
				}
				if len(record.Confirmed.AttachmentIDs) != 0 {
					if err := draft.Update().RemoveAttachmentIDs(record.Confirmed.AttachmentIDs...).Exec(ctx); err != nil {
						return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
					}
				}
				if err := transaction.MailDraft.DeleteOne(draft).Exec(ctx); err != nil {
					return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
				}
				if err := mailboxEntDraftCreateRelease(ctx, account, request.Fence, databaseWallClock); err != nil {
					return err
				}
				result.DraftDeleted = true
				return nil
			case MailboxDraftMutationSave:
				if record.PendingSave == nil {
					return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorStateConflict)
				}
				if err := draft.Update().SetStatus(maildraft.StatusActive).ClearLastMutationID().ClearPendingMutationKind().ClearPendingFromIdentityID().ClearPendingReplyToMessageID().ClearPendingForwardMessageID().ClearPendingSaveRequestHash().ClearPendingAttachments().SetUpdatedAt(databaseWallClock).Exec(ctx); err != nil {
					return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
				}
				updated := mailboxEntDraftRecordCopy(record)
				updated.Status = MailboxDraftStatusActive
				updated.PendingSave = nil
				updated.LastMutationID = nil
				updated.PendingMutationKind = nil
				updated.UpdatedAt = databaseWallClock.UTC()
				if err := mailboxEntDraftCreateRelease(ctx, account, request.Fence, databaseWallClock); err != nil {
					return err
				}
				result.Draft = &updated
				return nil
			default:
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorStateConflict)
			}
		case MailboxDraftStatusDeleting:
			if record.LastMutationID == nil || *record.LastMutationID != request.MutationID || record.PendingMutationKind == nil || *record.PendingMutationKind != MailboxDraftMutationDelete || record.PendingSave != nil || len(attachments.PendingAttachmentIDs) != 0 {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorStateConflict)
			}
			if err := draft.Update().SetStatus(maildraft.StatusActive).ClearLastMutationID().ClearPendingMutationKind().SetUpdatedAt(databaseWallClock).Exec(ctx); err != nil {
				return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
			}
			updated := mailboxEntDraftRecordCopy(record)
			updated.Status = MailboxDraftStatusActive
			updated.LastMutationID = nil
			updated.PendingMutationKind = nil
			updated.UpdatedAt = databaseWallClock.UTC()
			if err := mailboxEntDraftCreateRelease(ctx, account, request.Fence, databaseWallClock); err != nil {
				return err
			}
			result.Draft = &updated
			return nil
		default:
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorStateConflict)
		}
	})
	if err != nil {
		return MailboxDraftAbortResult{}, mailboxEntDraftRepositoryError(err)
	}
	if postCommitErr != nil {
		return MailboxDraftAbortResult{}, postCommitErr
	}
	return result, nil
}

func mailboxEntDraftReservationSaveRequestKind(request MailboxDraftSaveRequest) MailboxDraftRepositoryErrorKind {
	if !mailboxEntEngineCommandFenceValid(request.Fence) || request.DraftID == uuid.Nil || request.DraftClientID == uuid.Nil || request.ExpectedRevision < 1 || request.MutationID == uuid.Nil || !mailboxEntDraftCreateHashPresent(request.PendingSave.RequestHash) {
		return MailboxDraftRepositoryErrorInvalid
	}
	if kind := mailboxEntDraftEffectiveStateKind(request.PendingSave.EffectiveState); kind != "" {
		return kind
	}
	if request.LinkedSourceProof != nil && !mailboxEntDraftReservationLinkedSourceProofValid(*request.LinkedSourceProof) {
		return MailboxDraftRepositoryErrorInvalid
	}
	return ""
}

func mailboxEntDraftReservationLinkedSourceProofValid(proof MailboxDraftLinkedSourceProof) bool {
	return proof.MessageID != uuid.Nil && syncValidOpaque(proof.EngineMessageID, MaxEngineOpaqueIDBytes) && syncValidOpaque(proof.EngineVersion, MaxEngineOpaqueIDBytes)
}

func mailboxEntDraftReservationLinkedSourceMatches(state MailboxDraftEffectiveState, proof *MailboxDraftLinkedSourceProof) bool {
	if proof == nil {
		return state.ReplyToMessageID == nil && state.ForwardMessageID == nil
	}
	linkedID, _, linked := mailboxDraftSaveLinkedSource(state)
	return mailboxEntDraftReservationLinkedSourceProofValid(*proof) && linked && linkedID == proof.MessageID
}

func mailboxEntDraftReservationLinkedStatesMatch(confirmed, pending MailboxDraftEffectiveState, proof *MailboxDraftLinkedSourceProof) bool {
	if proof == nil {
		return confirmed.ReplyToMessageID == nil && confirmed.ForwardMessageID == nil && pending.ReplyToMessageID == nil && pending.ForwardMessageID == nil
	}
	confirmedID, confirmedReply, confirmedOK := mailboxDraftSaveLinkedSource(confirmed)
	pendingID, pendingReply, pendingOK := mailboxDraftSaveLinkedSource(pending)
	return mailboxEntDraftReservationLinkedSourceProofValid(*proof) && confirmedOK && pendingOK &&
		confirmedReply == pendingReply && confirmedID == pendingID && confirmedID == proof.MessageID
}

func mailboxEntDraftReservationLockLinkedSource(ctx context.Context, transaction *ent.Tx, accountID uuid.UUID, state MailboxDraftEffectiveState, proof *MailboxDraftLinkedSourceProof) error {
	if proof == nil {
		return nil
	}
	if ctx == nil || transaction == nil || accountID == uuid.Nil || !mailboxEntDraftReservationLinkedSourceMatches(state, proof) {
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	source, err := transaction.MailMessage.Query().Where(mailmessage.IDEQ(proof.MessageID), mailmessage.MailAccountIDEQ(accountID), mailmessage.DeletedAtIsNil()).ForUpdate().Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorStateConflict)
		}
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
	}
	if source == nil || !mailboxEntApplyMessageRowValid(source, accountID) || source.ID != proof.MessageID || source.MailAccountID != accountID || source.DeletedAt != nil || source.EngineMessageID != proof.EngineMessageID || source.EngineVersion != proof.EngineVersion || source.IsDraft || source.ProjectionOrigin != mailmessage.ProjectionOrigin(EngineProjectionOriginInboundOrUnknown) || source.CalendarMethod != nil || source.CalendarEventID != nil || mailboxEntCalendarPartLocatorHashValue(source.CalendarPartLocatorHash) != nil {
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorStateConflict)
	}
	return nil
}

func mailboxEntDraftAbortRequestKind(request MailboxDraftAbortRequest) MailboxDraftRepositoryErrorKind {
	if !mailboxEntEngineCommandFenceValid(request.Fence) || request.DraftID == uuid.Nil || request.MutationID == uuid.Nil {
		return MailboxDraftRepositoryErrorInvalid
	}
	return ""
}

func mailboxEntDraftReservationDeleteRequestKind(request MailboxDraftDeleteRequest) MailboxDraftRepositoryErrorKind {
	if !mailboxEntEngineCommandFenceValid(request.Fence) || request.DraftID == uuid.Nil || request.ExpectedRevision < 1 || request.MutationID == uuid.Nil {
		return MailboxDraftRepositoryErrorInvalid
	}
	return ""
}

func mailboxEntDraftReservationLockDraft(ctx context.Context, transaction *ent.Tx, accountID, draftID uuid.UUID) (*ent.MailDraft, MailboxDraftRepositoryErrorKind) {
	if ctx == nil || transaction == nil || accountID == uuid.Nil || draftID == uuid.Nil {
		return nil, MailboxDraftRepositoryErrorInvalid
	}
	draft, err := transaction.MailDraft.Query().Where(maildraft.IDEQ(draftID), maildraft.MailAccountIDEQ(accountID)).ForUpdate().Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, MailboxDraftRepositoryErrorStateConflict
		}
		return nil, MailboxDraftRepositoryErrorUnavailable
	}
	if draft == nil || draft.ID != draftID || draft.MailAccountID != accountID || !mailboxEntDraftReadRowValid(draft, accountID, draftID) {
		return nil, MailboxDraftRepositoryErrorStateConflict
	}
	return draft, ""
}

// mailboxEntDraftAbortLockDraft marks only absent or identity-mismatched rows as accepted release cases.
func mailboxEntDraftAbortLockDraft(ctx context.Context, transaction *ent.Tx, accountID, draftID uuid.UUID) (*ent.MailDraft, bool, MailboxDraftRepositoryErrorKind) {
	if ctx == nil || transaction == nil || accountID == uuid.Nil || draftID == uuid.Nil {
		return nil, false, MailboxDraftRepositoryErrorInvalid
	}
	draft, err := transaction.MailDraft.Query().Where(maildraft.IDEQ(draftID), maildraft.MailAccountIDEQ(accountID)).ForUpdate().Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, true, MailboxDraftRepositoryErrorStateConflict
		}
		return nil, false, MailboxDraftRepositoryErrorUnavailable
	}
	if draft == nil || draft.ID != draftID || draft.MailAccountID != accountID {
		return nil, true, MailboxDraftRepositoryErrorStateConflict
	}
	if !mailboxEntDraftReadRowValid(draft, accountID, draftID) {
		return nil, false, MailboxDraftRepositoryErrorStateConflict
	}
	return draft, false, ""
}

// mailboxEntDraftReservationLockSaveAttachments snapshots O/D without locks only to assemble
// the complete target union, then locks every attachment and every object in canonical order.
func mailboxEntDraftReservationLockSaveAttachments(ctx context.Context, transaction *ent.Tx, draft *ent.MailDraft, request MailboxDraftSaveRequest) (mailboxEntDraftAttachmentState, MailboxDraftRepositoryErrorKind) {
	if ctx == nil || transaction == nil || draft == nil || draft.ID == uuid.Nil {
		return mailboxEntDraftAttachmentState{}, MailboxDraftRepositoryErrorInvalid
	}
	rows, err := transaction.MailAttachment.Query().Where(mailattachment.Or(mailattachment.MailDraftIDEQ(draft.ID), mailattachment.PendingMailDraftIDEQ(draft.ID))).Order(ent.Asc(mailattachment.FieldID)).Limit(MaxEngineAttachmentItems*2 + 1).All(ctx)
	if err != nil {
		return mailboxEntDraftAttachmentState{}, MailboxDraftRepositoryErrorUnavailable
	}
	if len(rows) > MaxEngineAttachmentItems*2 {
		return mailboxEntDraftAttachmentState{}, MailboxDraftRepositoryErrorStateConflict
	}
	union := make([]uuid.UUID, 0, len(rows)+len(request.PendingSave.EffectiveState.AttachmentIDs))
	for _, row := range rows {
		if row == nil || row.ID == uuid.Nil {
			return mailboxEntDraftAttachmentState{}, MailboxDraftRepositoryErrorStateConflict
		}
		union = append(union, row.ID)
	}
	union = append(union, request.PendingSave.EffectiveState.AttachmentIDs...)
	union = mailboxEntDraftReservationUniqueUUIDs(union)
	desired := make(map[uuid.UUID]struct{}, len(request.PendingSave.EffectiveState.AttachmentIDs))
	for _, id := range request.PendingSave.EffectiveState.AttachmentIDs {
		desired[id] = struct{}{}
	}
	locked := make(map[uuid.UUID]*ent.MailAttachment, len(union))
	objectIDs := make([]uuid.UUID, 0, len(union))
	objectsSeen := make(map[uuid.UUID]struct{}, len(union))
	state := mailboxEntDraftAttachmentState{}
	for _, attachmentID := range union {
		attachment, err := transaction.MailAttachment.Query().Where(mailattachment.IDEQ(attachmentID)).ForUpdate().Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return mailboxEntDraftAttachmentState{}, MailboxDraftRepositoryErrorAttachmentConflict
			}
			return mailboxEntDraftAttachmentState{}, MailboxDraftRepositoryErrorUnavailable
		}
		confirmed := attachment.MailDraftID != nil && *attachment.MailDraftID == draft.ID
		pending := attachment.PendingMailDraftID != nil && *attachment.PendingMailDraftID == draft.ID
		_, wantsPending := desired[attachmentID]
		if confirmed || pending {
			if !mailboxEntDraftReadAttachmentValid(attachment, draft.ID, draft.DraftClientID, confirmed, pending) {
				return mailboxEntDraftAttachmentState{}, MailboxDraftRepositoryErrorStateConflict
			}
			if confirmed {
				state.ConfirmedAttachmentIDs = append(state.ConfirmedAttachmentIDs, attachmentID)
			}
			if pending {
				state.PendingAttachmentIDs = append(state.PendingAttachmentIDs, attachmentID)
			}
			if wantsPending && attachment.ScanStatus != mailattachment.ScanStatus("clean") {
				return mailboxEntDraftAttachmentState{}, MailboxDraftRepositoryErrorAttachmentConflict
			}
		} else if !wantsPending || !mailboxEntDraftCreateAttachmentValid(attachment, request.DraftClientID) {
			return mailboxEntDraftAttachmentState{}, MailboxDraftRepositoryErrorAttachmentConflict
		}
		if attachment.StorageObjectID == nil || *attachment.StorageObjectID == uuid.Nil {
			return mailboxEntDraftAttachmentState{}, MailboxDraftRepositoryErrorAttachmentConflict
		}
		locked[attachmentID] = attachment
		objectID := *attachment.StorageObjectID
		if _, exists := objectsSeen[objectID]; !exists {
			objectsSeen[objectID] = struct{}{}
			objectIDs = append(objectIDs, objectID)
		}
	}
	if len(state.ConfirmedAttachmentIDs) > MaxEngineAttachmentItems || len(state.PendingAttachmentIDs) > MaxEngineAttachmentItems {
		return mailboxEntDraftAttachmentState{}, MailboxDraftRepositoryErrorStateConflict
	}
	sort.Slice(state.ConfirmedAttachmentIDs, func(left, right int) bool { return syncLessUUID(state.ConfirmedAttachmentIDs[left], state.ConfirmedAttachmentIDs[right]) })
	sort.Slice(state.PendingAttachmentIDs, func(left, right int) bool { return syncLessUUID(state.PendingAttachmentIDs[left], state.PendingAttachmentIDs[right]) })
	sort.Slice(objectIDs, func(left, right int) bool { return syncLessUUID(objectIDs[left], objectIDs[right]) })
	objects := make(map[uuid.UUID]*ent.MailObject, len(objectIDs))
	for _, objectID := range objectIDs {
		object, err := transaction.MailObject.Query().Where(mailobject.IDEQ(objectID)).ForUpdate().Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return mailboxEntDraftAttachmentState{}, MailboxDraftRepositoryErrorAttachmentConflict
			}
			return mailboxEntDraftAttachmentState{}, MailboxDraftRepositoryErrorUnavailable
		}
		objects[objectID] = object
	}
	for _, attachmentID := range union {
		attachment := locked[attachmentID]
		if !mailboxEntDraftReadStorageObjectValid(objects[*attachment.StorageObjectID], draft.MailAccountID, draft.DraftClientID, attachment) {
			return mailboxEntDraftAttachmentState{}, MailboxDraftRepositoryErrorAttachmentConflict
		}
	}
	return state, ""
}

func mailboxEntDraftReservationUniqueUUIDs(values []uuid.UUID) []uuid.UUID {
	sort.Slice(values, func(left, right int) bool { return syncLessUUID(values[left], values[right]) })
	if len(values) == 0 {
		return []uuid.UUID{}
	}
	unique := values[:1]
	for _, value := range values[1:] {
		if value != unique[len(unique)-1] {
			unique = append(unique, value)
		}
	}
	return unique
}

func mailboxEntDraftReservationSaveTransition(ctx context.Context, draft *ent.MailDraft, request MailboxDraftSaveRequest, databaseWallClock time.Time) (*ent.MailDraft, error) {
	if ctx == nil || draft == nil || !mailboxEntValidTime(databaseWallClock) {
		return nil, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	update := draft.Update().SetStatus(maildraft.StatusSaving).SetLastMutationID(request.MutationID).SetPendingMutationKind(maildraft.PendingMutationKind("save")).SetPendingFromIdentityID(request.PendingSave.EffectiveState.FromIdentityID).SetPendingSaveRequestHash(append([]byte(nil), request.PendingSave.RequestHash[:]...)).SetUpdatedAt(databaseWallClock)
	if request.PendingSave.EffectiveState.ReplyToMessageID != nil {
		update.SetPendingReplyToMessageID(*request.PendingSave.EffectiveState.ReplyToMessageID)
	}
	if request.PendingSave.EffectiveState.ForwardMessageID != nil {
		update.SetPendingForwardMessageID(*request.PendingSave.EffectiveState.ForwardMessageID)
	}
	if len(request.PendingSave.EffectiveState.AttachmentIDs) != 0 {
		update.AddPendingAttachmentIDs(request.PendingSave.EffectiveState.AttachmentIDs...)
	}
	return update.Save(ctx)
}

func mailboxEntDraftReservationDeleteTransition(ctx context.Context, draft *ent.MailDraft, request MailboxDraftDeleteRequest, databaseWallClock time.Time) (*ent.MailDraft, error) {
	if ctx == nil || draft == nil || !mailboxEntValidTime(databaseWallClock) {
		return nil, NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorInvalid)
	}
	return draft.Update().SetStatus(maildraft.StatusDeleting).SetLastMutationID(request.MutationID).SetPendingMutationKind(maildraft.PendingMutationKind("delete")).SetUpdatedAt(databaseWallClock).Save(ctx)
}

func mailboxEntDraftPendingSavesEqual(left, right *MailboxDraftPendingSave) bool {
	return left != nil && right != nil && left.RequestHash == right.RequestHash && mailboxEntDraftEffectiveStatesEqual(left.EffectiveState, right.EffectiveState)
}

func mailboxEntDraftReservationUUIDSetsEqual(left, right []uuid.UUID) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func mailboxEntDraftReservationReleaseForError(ctx context.Context, account *ent.MailAccount, fence MailboxEngineCommandFence, databaseWallClock time.Time, err error) error {
	if repositoryError, ok := err.(*MailboxDraftRepositoryError); ok && repositoryError != nil && repositoryError.Kind() == MailboxDraftRepositoryErrorUnavailable {
		return err
	}
	return mailboxEntDraftReservationReleaseError(ctx, account, fence, databaseWallClock, mailboxEntDraftCreateErrorKind(err))
}

func mailboxEntDraftReservationReleaseError(ctx context.Context, account *ent.MailAccount, fence MailboxEngineCommandFence, databaseWallClock time.Time, kind MailboxDraftRepositoryErrorKind) error {
	if kind == MailboxDraftRepositoryErrorUnavailable {
		return NewMailboxDraftRepositoryError(MailboxDraftRepositoryErrorUnavailable)
	}
	return mailboxEntDraftCreateReleaseError(ctx, account, fence, databaseWallClock, kind)
}
