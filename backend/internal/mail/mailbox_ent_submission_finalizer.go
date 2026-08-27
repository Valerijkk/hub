package mail

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/mailattachment"
	"github.com/valerii/hub/backend/ent/maildomain"
	"github.com/valerii/hub/backend/ent/maildraft"
	"github.com/valerii/hub/backend/ent/mailfolder"
	"github.com/valerii/hub/backend/ent/mailidentity"
	"github.com/valerii/hub/backend/ent/maillabel"
	"github.com/valerii/hub/backend/ent/mailmessage"
	"github.com/valerii/hub/backend/ent/mailobject"
	"github.com/valerii/hub/backend/ent/mailsubmission"
	"github.com/valerii/hub/backend/ent/mailthread"
	"github.com/valerii/hub/backend/ent/outboxevent"
	"github.com/valerii/hub/backend/ent/schema"
)

var _ MailboxSubmissionFinalizationRepository = (*MailboxEntRepository)(nil)

// ScanAcceptedMailboxSubmissions returns only stable, unleased accepted identities.
// The selection is advisory: ClaimAcceptedMailboxSubmission rechecks every row.
func (repository *MailboxEntRepository) ScanAcceptedMailboxSubmissions(ctx context.Context, request MailboxSubmissionFinalizationScanRequest) (MailboxSubmissionFinalizationScanResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxSubmissionFinalizationScanRequestValid(request) {
		return MailboxSubmissionFinalizationScanResult{}, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid)
	}

	var result MailboxSubmissionFinalizationScanResult
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable)
		}
		rows, err := transaction.MailSubmission.Query().
			Where(
				mailsubmission.StatusEQ(mailsubmission.Status("accepted")),
				mailsubmission.UpdatedAtLTE(databaseNow),
				mailsubmission.CommandLeaseTokenIsNil(),
				mailsubmission.CommandLeaseExpiresAtIsNil(),
				mailsubmission.ProviderSubmissionIDNotNil(),
				mailsubmission.ProviderMessageIDNotNil(),
				mailsubmission.DispatchAuthorizedAtNotNil(),
				mailsubmission.DispatchAuthorizationHashNotNil(),
				mailsubmission.ResultMailMessageIDIsNil(),
				mailsubmission.LastErrorCodeIsNil(),
				mailsubmission.NextAttemptAtIsNil(),
				mailsubmission.NextReconcileAtIsNil(),
				mailsubmission.MailDraftIDNotNil(),
			).
			Order(ent.Asc(mailsubmission.FieldUpdatedAt), ent.Asc(mailsubmission.FieldID)).
			Limit(request.Limit).
			All(ctx)
		if err != nil {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable)
		}
		ids := make([]uuid.UUID, 0, len(rows))
		for _, submission := range rows {
			if !mailboxEntSubmissionFinalizationScanRowValid(submission, databaseNow) {
				return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid)
			}
			ids = append(ids, submission.ID)
		}
		candidate := MailboxSubmissionFinalizationScanResult{SubmissionIDs: ids}
		if !mailboxSubmissionFinalizationScanResultValid(request, candidate) {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid)
		}
		result = mailboxSubmissionFinalizationScanResultCopy(candidate)
		return nil
	})
	if err != nil {
		return MailboxSubmissionFinalizationScanResult{}, mailboxEntSubmissionFinalizationRepositoryError(err)
	}
	if !mailboxSubmissionFinalizationScanResultValid(request, result) {
		return MailboxSubmissionFinalizationScanResult{}, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable)
	}
	return mailboxSubmissionFinalizationScanResultCopy(result), nil
}

// ClaimAcceptedMailboxSubmission freezes one accepted submission and its durable,
// clean storage-backed attachment set behind a fresh shared account/submission fence.
func (repository *MailboxEntRepository) ClaimAcceptedMailboxSubmission(ctx context.Context, request MailboxSubmissionFinalizationClaimRequest) (MailboxSubmissionFinalizationClaimResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxSubmissionFinalizationClaimRequestValid(request) {
		return MailboxSubmissionFinalizationClaimResult{}, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid)
	}

	lookup, err := repository.client.MailSubmission.Query().Where(mailsubmission.IDEQ(request.SubmissionID)).Only(ctx)
	if ent.IsNotFound(err) {
		return NewMailboxSubmissionFinalizationNoopResult(), nil
	}
	if err != nil {
		return MailboxSubmissionFinalizationClaimResult{}, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable)
	}
	if lookup == nil || lookup.ID != request.SubmissionID || lookup.MailAccountID == uuid.Nil {
		return MailboxSubmissionFinalizationClaimResult{}, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid)
	}

	result := NewMailboxSubmissionFinalizationNoopResult()
	err = repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		// Lock order: MailAccount -> MailSubmission -> MailDraft -> MailAttachment -> MailObject.
		account, err := transaction.MailAccount.Query().Where(mailaccount.IDEQ(lookup.MailAccountID)).ForUpdate().Only(ctx)
		if ent.IsNotFound(err) {
			return nil
		}
		if err != nil {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable)
		}
		submission, err := transaction.MailSubmission.Query().
			Where(mailsubmission.IDEQ(request.SubmissionID), mailsubmission.MailAccountIDEQ(account.ID)).
			ForUpdate().
			Only(ctx)
		if ent.IsNotFound(err) {
			return nil
		}
		if err != nil {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable)
		}
		if submission.MailDraftID == nil || *submission.MailDraftID == uuid.Nil {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid)
		}
		draft, err := transaction.MailDraft.Query().
			Where(maildraft.IDEQ(*submission.MailDraftID), maildraft.MailAccountIDEQ(account.ID)).
			ForUpdate().
			Only(ctx)
		if ent.IsNotFound(err) {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
		}
		if err != nil {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable)
		}
		databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable)
		}
		if mailboxEntSubmissionClaimAccountKind(account) != "" || !mailboxEntSubmissionRowValid(submission, account.ID) {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid)
		}
		if MailboxSubmissionStatus(submission.Status) != MailboxSubmissionStatusAccepted {
			return nil
		}
		if busyUntil, busy, valid := mailboxEntSubmissionClaimBusyUntil(account); !valid {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid)
		} else if busy {
			busyResult, err := NewMailboxSubmissionFinalizationBusyResult(busyUntil)
			if err != nil {
				return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid)
			}
			result = busyResult
			return nil
		}
		if !mailboxEntSubmissionFinalizationAcceptedShape(submission, draft, account.ID, databaseNow) {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid)
		}
		expected, err := mailboxEntSubmissionFinalizationLockExpectedAttachments(ctx, transaction, account, draft)
		if err != nil {
			return err
		}
		leaseUntil, ok := mailboxEntCheckedTimeAdd(databaseNow, MailboxEngineCommandLeaseDuration)
		if !ok {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid)
		}
		nextEpoch, ok := mailboxEntNextEpoch(account.MutationEpoch)
		if !ok {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid)
		}
		leaseID := uuid.New()
		fence := MailboxEngineCommandFence{AccountID: account.ID, LeaseID: leaseID, MutationEpoch: nextEpoch, LeaseUntil: leaseUntil}
		if !mailboxSubmissionClaimFenceValid(fence) {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid)
		}
		if err := account.Update().
			SetMutationLeaseToken(leaseID).
			SetMutationLeaseUntil(leaseUntil).
			SetMutationEpoch(nextEpoch).
			SetSyncScheduleUpdatedAt(databaseNow).
			Exec(ctx); err != nil {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable)
		}
		if err := submission.Update().
			SetCommandLeaseToken(leaseID).
			SetCommandLeaseExpiresAt(leaseUntil).
			SetCommandFence(nextEpoch).
			SetUpdatedAt(databaseNow).
			Exec(ctx); err != nil {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable)
		}
		claimed, err := NewMailboxSubmissionFinalizationClaimedResult(MailboxSubmissionFinalizationClaimedInput{
			SubmissionID:              submission.ID,
			DraftID:                   draft.ID,
			Fence:                     fence,
			EngineAccountID:           *account.EngineAccountID,
			ProviderSubmissionID:      *submission.ProviderSubmissionID,
			ProviderMessageID:         *submission.ProviderMessageID,
			EngineMessageID:           submission.EngineMessageID,
			EngineMaxBytes:            submission.EngineMaxMessageBytes,
			EngineBindingHash:         append([]byte(nil), submission.EngineBindingHash...),
			DispatchAuthorizationHash: append([]byte(nil), submission.DispatchAuthorizationHash...),
			DispatchAuthorizedAt:      *submission.DispatchAuthorizedAt,
			ExpectedAttachments:       expected,
		})
		if err != nil {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid)
		}
		result = claimed
		return nil
	})
	if err != nil {
		return MailboxSubmissionFinalizationClaimResult{}, mailboxEntSubmissionFinalizationRepositoryError(err)
	}
	if !mailboxSubmissionFinalizationClaimResultValid(result) {
		return MailboxSubmissionFinalizationClaimResult{}, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable)
	}
	return result, nil
}

// CompleteAcceptedMailboxSubmission persists the hydrated confirmed projection,
// transfers the pre-sealed durable attachments, and records mail.sent.v1 in the
// same transaction as the sent terminal state.
func (repository *MailboxEntRepository) CompleteAcceptedMailboxSubmission(ctx context.Context, completion MailboxSubmissionFinalizationCompletion) error {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxSubmissionFinalizationCompletionValid(completion) {
		return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid)
	}
	completion = mailboxSubmissionFinalizationCompletionCopy(completion)
	messageID, err := mailboxEntStableSyncID(completion.claim.fence.AccountID, "message", []byte(completion.message.Projection.MessageID))
	if err != nil {
		return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid)
	}

	err = repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		// Lock order: MailAccount -> MailSubmission -> MailDraft -> all candidate MailThread IDs
		// (sorted) -> all target/ancestor MailMessage rows (sorted) -> MailAttachment -> MailObject
		// -> folder/label links -> OutboxEvent.
		account, err := transaction.MailAccount.Query().Where(mailaccount.IDEQ(completion.claim.fence.AccountID)).ForUpdate().Only(ctx)
		if err != nil {
			return mailboxEntSubmissionFinalizationCompletionLookupError(err)
		}
		submission, err := transaction.MailSubmission.Query().
			Where(mailsubmission.IDEQ(completion.claim.submissionID), mailsubmission.MailAccountIDEQ(account.ID)).
			ForUpdate().
			Only(ctx)
		if err != nil {
			return mailboxEntSubmissionFinalizationCompletionLookupError(err)
		}
		if submission.MailDraftID == nil || *submission.MailDraftID != completion.claim.draftID {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
		}
		draft, err := transaction.MailDraft.Query().
			Where(maildraft.IDEQ(completion.claim.draftID), maildraft.MailAccountIDEQ(account.ID)).
			ForUpdate().
			Only(ctx)
		if err != nil {
			return mailboxEntSubmissionFinalizationCompletionLookupError(err)
		}
		terminal, err := mailboxEntSubmissionFinalizationTerminalReplayNoop(ctx, transaction, account, submission, draft, completion, messageID)
		if err != nil {
			return err
		}
		if terminal {
			return nil
		}
		databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable)
		}
		if !mailboxEntSubmissionFinalizationClaimFenceMatches(account, submission, draft, completion.claim, databaseNow) {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
		}

		ancestors, err := mailboxEntSubmissionFinalizationPreReadAncestors(ctx, transaction, account.ID, completion.message.Projection)
		if err != nil {
			return err
		}
		threadIdentity, err := DeriveThreadIdentity(ThreadingInput{
			References: completion.message.Projection.References, InReplyTo: completion.message.Projection.InReplyTo,
			Subject: completion.message.Projection.Subject, From: completion.message.Projection.From,
			To: completion.message.Projection.To, CC: completion.message.Projection.CC,
			Ancestors: ancestors.ancestors, AncestorLookupComplete: true,
		})
		if err != nil || threadIdentity.KeyVersion != ThreadKeyVersionV1 || threadKeyIsZero(threadIdentity.Key) {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
		}
		threadID, err := mailboxEntStableSyncID(account.ID, "thread", threadIdentity.Key[:])
		if err != nil {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid)
		}
		lockedThreads, err := mailboxEntSubmissionFinalizationLockThreads(ctx, transaction, account, ancestors, threadID)
		if err != nil {
			return err
		}
		thread, err := mailboxEntSubmissionFinalizationResolveLockedThread(ctx, transaction, account, threadID, threadIdentity, completion.message.Projection, databaseNow, lockedThreads)
		if err != nil {
			return err
		}
		messages, candidateMessages, err := mailboxEntSubmissionFinalizationLockMessages(ctx, transaction, account, messageID, completion.message.Projection)
		if err != nil {
			return err
		}
		if err := mailboxEntSubmissionFinalizationRevalidateAncestors(account.ID, completion.message.Projection, ancestors, lockedThreads, messages); err != nil {
			return err
		}
		attachments, err := mailboxEntSubmissionFinalizationLockAttachments(ctx, transaction, account, draft, completion.claim)
		if err != nil {
			return err
		}
		message, err := mailboxEntSubmissionFinalizationUpsertMessage(ctx, transaction, account, thread, messageID, candidateMessages, completion.message.Projection, databaseNow)
		if err != nil {
			return err
		}
		if err := mailboxEntSubmissionFinalizationTransferAttachments(ctx, message.ID, attachments, databaseNow); err != nil {
			return err
		}
		if err := message.Update().SetHasAttachments(len(attachments) != 0).SetUpdatedAt(databaseNow).Exec(ctx); err != nil {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable)
		}
		folders, labels, err := mailboxEntSubmissionFinalizationResolveFoldersAndLabels(ctx, transaction, account.ID, completion.message.Projection)
		if err != nil {
			return err
		}
		if _, err := mailboxEntApplyMessageFolderLinks(ctx, transaction, account.ID, message.ID, folders, account.SyncEpoch, databaseNow); err != nil {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
		}
		if _, err := mailboxEntApplyMessageLabelLinks(ctx, transaction, account.ID, message.ID, labels, databaseNow); err != nil {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
		}
		if err := mailboxEntApplySyncThreadAggregates(ctx, transaction, account.ID, account.SyncEpoch, []uuid.UUID{thread.ID}, databaseNow); err != nil {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
		}
		routes, err := mailboxEntSubmissionFinalizationResolveRoutes(ctx, transaction, account, completion.message.Projection)
		if err != nil {
			return err
		}
		event, err := NewMailboxSubmissionSentEnvelopeV1(MailboxSubmissionSentEventInput{
			AccountID: account.ID, OrganizationID: account.OrganizationID, ActorID: account.UserID,
			SubmissionID: submission.ID, MessageID: message.ID, ThreadID: thread.ID,
			DispatchAuthorizedAt: completion.claim.dispatchAuthorizedAt,
			Projection: completion.message.Projection, RecipientResolutions: routes,
		})
		if err != nil {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
		}
		if err := mailboxEntSubmissionFinalizationAppendOutbox(ctx, transaction, event, databaseNow); err != nil {
			return err
		}
		nextEpoch, ok := mailboxEntNextEpoch(account.MutationEpoch)
		if !ok {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid)
		}
		if err := submission.Update().
			SetStatus(mailsubmission.Status("sent")).
			ClearCommandLeaseToken().
			ClearCommandLeaseExpiresAt().
			SetResultMailMessageID(message.ID).
			ClearLastErrorCode().
			ClearNextAttemptAt().
			ClearNextReconcileAt().
			SetUpdatedAt(databaseNow).
			Exec(ctx); err != nil {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable)
		}
		if err := draft.Update().SetStatus(maildraft.Status("submitted")).SetUpdatedAt(databaseNow).Exec(ctx); err != nil {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable)
		}
		if err := account.Update().
			ClearMutationLeaseToken().
			ClearMutationLeaseUntil().
			SetMutationEpoch(nextEpoch).
			SetSyncScheduleUpdatedAt(databaseNow).
			Exec(ctx); err != nil {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable)
		}
		return nil
	})
	if err != nil {
		return mailboxEntSubmissionFinalizationRepositoryError(err)
	}
	return nil
}

// ReleaseAcceptedMailboxSubmission releases only this exact accepted claim.
// It cannot change provider confirmation, draft state, or terminal outcome.
func (repository *MailboxEntRepository) ReleaseAcceptedMailboxSubmission(ctx context.Context, claim MailboxSubmissionFinalizationClaim) error {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxSubmissionFinalizationClaimValid(claim) {
		return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid)
	}
	claim = mailboxSubmissionFinalizationClaimCopy(claim)
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, err := transaction.MailAccount.Query().Where(mailaccount.IDEQ(claim.fence.AccountID)).ForUpdate().Only(ctx)
		if err != nil {
			return mailboxEntSubmissionFinalizationCompletionLookupError(err)
		}
		submission, err := transaction.MailSubmission.Query().
			Where(mailsubmission.IDEQ(claim.submissionID), mailsubmission.MailAccountIDEQ(account.ID)).
			ForUpdate().
			Only(ctx)
		if err != nil {
			return mailboxEntSubmissionFinalizationCompletionLookupError(err)
		}
		if submission.MailDraftID == nil || *submission.MailDraftID != claim.draftID {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
		}
		draft, err := transaction.MailDraft.Query().
			Where(maildraft.IDEQ(claim.draftID), maildraft.MailAccountIDEQ(account.ID)).
			ForUpdate().
			Only(ctx)
		if err != nil {
			return mailboxEntSubmissionFinalizationCompletionLookupError(err)
		}
		databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable)
		}
		if !mailboxEntSubmissionFinalizationClaimFenceMatches(account, submission, draft, claim, databaseNow) {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
		}
		nextEpoch, ok := mailboxEntNextEpoch(account.MutationEpoch)
		if !ok {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid)
		}
		if err := submission.Update().
			SetStatus(mailsubmission.Status("accepted")).
			ClearCommandLeaseToken().
			ClearCommandLeaseExpiresAt().
			ClearResultMailMessageID().
			ClearLastErrorCode().
			ClearNextAttemptAt().
			ClearNextReconcileAt().
			SetUpdatedAt(databaseNow).
			Exec(ctx); err != nil {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable)
		}
		if err := account.Update().
			ClearMutationLeaseToken().
			ClearMutationLeaseUntil().
			SetMutationEpoch(nextEpoch).
			SetSyncScheduleUpdatedAt(databaseNow).
			Exec(ctx); err != nil {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable)
		}
		return nil
	})
	if err != nil {
		return mailboxEntSubmissionFinalizationRepositoryError(err)
	}
	return nil
}

// mailboxEntSubmissionFinalizationTerminalReplayNoop accepts an ambiguous replay
// only when the already committed terminal snapshot is still provably the exact
// completion requested by this sealed claim.  It never repairs partial state.
func mailboxEntSubmissionFinalizationTerminalReplayNoop(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, submission *ent.MailSubmission, draft *ent.MailDraft, completion MailboxSubmissionFinalizationCompletion, messageID uuid.UUID) (bool, error) {
	if ctx == nil || transaction == nil || account == nil || submission == nil || draft == nil || messageID == uuid.Nil {
		return false, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
	}
	if MailboxSubmissionStatus(submission.Status) != MailboxSubmissionStatusSent {
		if MailboxSubmissionStatus(submission.Status) != MailboxSubmissionStatusAccepted {
			return false, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
		}
		return false, nil
	}
	if !mailboxEntSubmissionFinalizationTerminalClaimMatches(account, submission, draft, completion.claim, messageID) {
		return false, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
	}
	message, err := mailboxEntSubmissionFinalizationTerminalMessageMatches(ctx, transaction, account, messageID, completion.claim.engineMessageID)
	if err != nil {
		return false, err
	}
	if err := mailboxEntSubmissionFinalizationTerminalOutbox(ctx, transaction, account, submission, message, message.MailThreadID, completion.claim); err != nil {
		return false, err
	}
	return true, nil
}

func mailboxEntSubmissionFinalizationTerminalClaimMatches(account *ent.MailAccount, submission *ent.MailSubmission, draft *ent.MailDraft, claim MailboxSubmissionFinalizationClaim, messageID uuid.UUID) bool {
	if account == nil || submission == nil || draft == nil || !mailboxSubmissionFinalizationClaimValid(claim) ||
		mailboxEntSubmissionClaimAccountKind(account) != "" || !mailboxEntSubmissionRowValid(submission, account.ID) ||
		account.ID != claim.fence.AccountID || submission.ID != claim.submissionID || submission.MailAccountID != account.ID ||
		MailboxSubmissionStatus(submission.Status) != MailboxSubmissionStatusSent || submission.MailDraftID == nil || *submission.MailDraftID != claim.draftID ||
		submission.ResultMailMessageID == nil || *submission.ResultMailMessageID != messageID ||
		submission.CommandLeaseToken != nil || submission.CommandLeaseExpiresAt != nil ||
		submission.CommandFence != claim.fence.MutationEpoch || submission.LastErrorCode != nil || submission.NextAttemptAt != nil || submission.NextReconcileAt != nil ||
		submission.DispatchAuthorizedAt == nil || !submission.DispatchAuthorizedAt.Equal(claim.dispatchAuthorizedAt) ||
		len(submission.EngineBindingHash) != sha256.Size || subtle.ConstantTimeCompare(submission.EngineBindingHash, claim.engineBindingHash[:]) != 1 ||
		len(submission.DispatchAuthorizationHash) != sha256.Size || subtle.ConstantTimeCompare(submission.DispatchAuthorizationHash, claim.dispatchAuthorizationHash[:]) != 1 ||
		submission.ProviderSubmissionID == nil || subtle.ConstantTimeCompare([]byte(*submission.ProviderSubmissionID), []byte(claim.providerSubmissionID)) != 1 ||
		submission.ProviderMessageID == nil || subtle.ConstantTimeCompare([]byte(*submission.ProviderMessageID), []byte(claim.providerMessageID)) != 1 ||
		submission.EngineMessageID != claim.engineMessageID || submission.EngineMaxMessageBytes != claim.engineMaxBytes {
		return false
	}
	nextEpoch, ok := mailboxEntNextEpoch(claim.fence.MutationEpoch)
	if !ok || account.MutationEpoch < nextEpoch {
		return false
	}
	binding, ok := mailboxEntSubmissionBindingFromRow(submission)
	if !ok || binding.engineAccountID != claim.engineAccountID || subtle.ConstantTimeCompare(binding.engineBindingHash[:], claim.engineBindingHash[:]) != 1 ||
		!mailboxEntDraftReadRowValid(draft, account.ID, claim.draftID) || MailboxDraftStatus(draft.Status) != MailboxDraftStatusSubmitted ||
		draft.EngineDraftKey != binding.engineDraftKey || draft.EngineMessageID == nil || *draft.EngineMessageID != binding.engineMessageID ||
		draft.EngineVersion == nil || *draft.EngineVersion != binding.engineExpectedVersion {
		return false
	}
	return true
}

func mailboxEntSubmissionFinalizationTerminalMessageMatches(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, messageID uuid.UUID, engineMessageID string) (*ent.MailMessage, error) {
	if ctx == nil || transaction == nil || account == nil || account.ID == uuid.Nil || messageID == uuid.Nil || !syncValidOpaque(engineMessageID, MaxEngineOpaqueIDBytes) {
		return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
	}
	rows, err := transaction.MailMessage.Query().
		Where(mailmessage.IDEQ(messageID), mailmessage.MailAccountIDEQ(account.ID), mailmessage.EngineMessageIDEQ(engineMessageID)).
		Order(ent.Asc(mailmessage.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable)
	}
	if len(rows) != 1 || rows[0] == nil || !mailboxEntApplyMessageRowValid(rows[0], account.ID) || rows[0].DeletedAt != nil ||
		rows[0].ID != messageID || rows[0].MailAccountID != account.ID || rows[0].MailThreadID == uuid.Nil ||
		rows[0].EngineMessageID != engineMessageID || rows[0].IsDraft ||
		rows[0].ProjectionOrigin != mailmessage.ProjectionOrigin(EngineProjectionOriginConfirmedOwnerSubmission) {
		return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
	}
	return rows[0], nil
}

type mailboxEntSubmissionFinalizationTerminalOutboxEnvelope struct {
	Subject        string          `json:"subject"`
	SchemaVersion  int             `json:"schema_version"`
	EventID        uuid.UUID       `json:"event_id"`
	OccurredAt     time.Time       `json:"occurred_at"`
	OrganizationID uuid.UUID       `json:"organization_id"`
	ActorID        *uuid.UUID      `json:"actor_id"`
	CorrelationID  string          `json:"correlation_id"`
	Payload        json.RawMessage `json:"payload"`
}

func mailboxEntSubmissionFinalizationTerminalEventAddress(value mailEventAddressJSON) MailEventAddress {
	result := MailEventAddress{Email: value.Email}
	if value.Name != nil {
		name := *value.Name
		result.Name = &name
	}
	return result
}

func mailboxEntSubmissionFinalizationTerminalEventAddresses(values []mailEventAddressJSON) []MailEventAddress {
	result := make([]MailEventAddress, len(values))
	for index, value := range values {
		result[index] = mailboxEntSubmissionFinalizationTerminalEventAddress(value)
	}
	return result
}

func mailboxEntSubmissionFinalizationTerminalOutbox(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, submission *ent.MailSubmission, message *ent.MailMessage, threadID uuid.UUID, claim MailboxSubmissionFinalizationClaim) error {
	if ctx == nil || transaction == nil || account == nil || submission == nil || message == nil || threadID == uuid.Nil || !mailboxSubmissionFinalizationClaimValid(claim) {
		return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
	}
	eventID := mailboxSubmissionSentEventID(submission.ID)
	rows, err := transaction.OutboxEvent.Query().Where(outboxevent.IDEQ(eventID)).ForUpdate().All(ctx)
	if err != nil {
		return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable)
	}
	if len(rows) != 1 || rows[0] == nil {
		return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
	}
	outbox := rows[0]
	if outbox.ID != eventID || outbox.EventType != "mail.sent.v1" || !json.Valid(outbox.Payload) {
		return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
	}
	var envelope mailboxEntSubmissionFinalizationTerminalOutboxEnvelope
	if err := json.Unmarshal(outbox.Payload, &envelope); err != nil {
		return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
	}
	var payload mailSentEventJSON
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
	}
	if envelope.Subject != "mail.sent.v1" || envelope.SchemaVersion != 1 || envelope.EventID != eventID ||
		!envelope.OccurredAt.Equal(claim.dispatchAuthorizedAt) || envelope.OrganizationID != account.OrganizationID || envelope.ActorID == nil ||
		*envelope.ActorID != account.UserID || envelope.CorrelationID != submission.ID.String() ||
		payload.MailboxID != account.ID || payload.SubmissionID != submission.ID || payload.MessageID != message.ID || payload.ThreadID != threadID ||
		!validMailDeliveryClass(payload.DeliveryClass) {
		return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
	}
	_, sentAtOffset := payload.SentAt.Zone()
	if !validMailEventTime(payload.SentAt) || sentAtOffset != 0 {
		return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
	}
	event, err := NewMailSentEnvelopeV1(
		envelope.EventID, envelope.OccurredAt, envelope.OrganizationID, envelope.ActorID, envelope.CorrelationID,
		MailSentEventPayload{
			Mailbox:       MailboxEventID{AccountID: account.ID, ID: payload.MailboxID},
			Submission:    MailboxEventID{AccountID: account.ID, ID: payload.SubmissionID},
			Message:       MailboxEventID{AccountID: account.ID, ID: payload.MessageID},
			Thread:        MailboxEventID{AccountID: account.ID, ID: payload.ThreadID},
			From:          mailboxEntSubmissionFinalizationTerminalEventAddress(payload.From),
			To:            mailboxEntSubmissionFinalizationTerminalEventAddresses(payload.To),
			SentAt:        payload.SentAt,
			DeliveryClass: payload.DeliveryClass,
		},
	)
	if err != nil {
		return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
	}
	canonical, err := event.MarshalJSON()
	if err != nil || subtle.ConstantTimeCompare(canonical, outbox.Payload) != 1 {
		return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil || subtle.ConstantTimeCompare(payloadJSON, envelope.Payload) != 1 {
		return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
	}
	envelopeJSON, err := json.Marshal(envelope)
	if err != nil || subtle.ConstantTimeCompare(envelopeJSON, outbox.Payload) != 1 {
		return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
	}
	return nil
}

func mailboxEntSubmissionFinalizationScanRowValid(submission *ent.MailSubmission, databaseNow time.Time) bool {
	return submission != nil && submission.ID != uuid.Nil && submission.MailAccountID != uuid.Nil &&
		mailboxEntValidTime(databaseNow) && mailboxEntSubmissionRowValid(submission, submission.MailAccountID) &&
		!submission.UpdatedAt.After(databaseNow) &&
		MailboxSubmissionStatus(submission.Status) == MailboxSubmissionStatusAccepted && submission.MailDraftID != nil && *submission.MailDraftID != uuid.Nil &&
		submission.CommandLeaseToken == nil && submission.CommandLeaseExpiresAt == nil && submission.DispatchAuthorizedAt != nil &&
		mailboxSubmissionClaimExactTimeValid(*submission.DispatchAuthorizedAt) && len(submission.DispatchAuthorizationHash) == sha256.Size &&
		submission.ProviderSubmissionID != nil && syncValidOpaque(*submission.ProviderSubmissionID, MaxEngineOpaqueIDBytes) &&
		submission.ProviderMessageID != nil && syncValidOpaque(*submission.ProviderMessageID, MaxEngineOpaqueIDBytes) &&
		submission.ResultMailMessageID == nil && submission.LastErrorCode == nil && submission.NextAttemptAt == nil &&
		submission.NextReconcileAt == nil && submission.AttemptCount >= 1 && submission.AttemptCount <= 8 &&
		submission.ReconcileAttempts >= 0 && submission.ReconcileAttempts <= 10
}

func mailboxEntSubmissionFinalizationAcceptedShape(submission *ent.MailSubmission, draft *ent.MailDraft, accountID uuid.UUID, databaseNow time.Time) bool {
	if !mailboxEntSubmissionFinalizationScanRowValid(submission, databaseNow) || submission.MailAccountID != accountID {
		return false
	}
	binding, ok := mailboxEntSubmissionBindingFromRow(submission)
	return ok && mailboxEntSubmissionClaimDraftMatches(draft, accountID, *submission.MailDraftID, binding)
}

func mailboxEntSubmissionFinalizationClaimFenceMatches(account *ent.MailAccount, submission *ent.MailSubmission, draft *ent.MailDraft, claim MailboxSubmissionFinalizationClaim, databaseNow time.Time) bool {
	if account == nil || submission == nil || draft == nil || !mailboxSubmissionFinalizationClaimValid(claim) || !mailboxEntValidTime(databaseNow) ||
		mailboxEntSubmissionClaimAccountKind(account) != "" || !mailboxEntSubmissionRowValid(submission, account.ID) ||
		account.ID != claim.fence.AccountID || submission.ID != claim.submissionID || submission.MailAccountID != account.ID ||
		MailboxSubmissionStatus(submission.Status) != MailboxSubmissionStatusAccepted || submission.MailDraftID == nil || *submission.MailDraftID != claim.draftID ||
		account.MutationLeaseToken == nil || account.MutationLeaseUntil == nil || submission.CommandLeaseToken == nil || submission.CommandLeaseExpiresAt == nil ||
		*account.MutationLeaseToken != claim.fence.LeaseID || *submission.CommandLeaseToken != claim.fence.LeaseID ||
		!account.MutationLeaseUntil.Equal(claim.fence.LeaseUntil) || !submission.CommandLeaseExpiresAt.Equal(claim.fence.LeaseUntil) ||
		account.MutationEpoch != claim.fence.MutationEpoch || submission.CommandFence != claim.fence.MutationEpoch ||
		!account.MutationLeaseUntil.After(databaseNow) || !submission.CommandLeaseExpiresAt.After(databaseNow) ||
		submission.DispatchAuthorizedAt == nil || !submission.DispatchAuthorizedAt.Equal(claim.dispatchAuthorizedAt) ||
		len(submission.EngineBindingHash) != sha256.Size || subtle.ConstantTimeCompare(submission.EngineBindingHash, claim.engineBindingHash[:]) != 1 ||
		len(submission.DispatchAuthorizationHash) != sha256.Size || subtle.ConstantTimeCompare(submission.DispatchAuthorizationHash, claim.dispatchAuthorizationHash[:]) != 1 ||
		submission.ProviderSubmissionID == nil || subtle.ConstantTimeCompare([]byte(*submission.ProviderSubmissionID), []byte(claim.providerSubmissionID)) != 1 ||
		submission.ProviderMessageID == nil || subtle.ConstantTimeCompare([]byte(*submission.ProviderMessageID), []byte(claim.providerMessageID)) != 1 ||
		submission.EngineMessageID != claim.engineMessageID || submission.EngineMaxMessageBytes != claim.engineMaxBytes ||
		submission.ResultMailMessageID != nil || submission.LastErrorCode != nil || submission.NextAttemptAt != nil || submission.NextReconcileAt != nil {
		return false
	}
	binding, ok := mailboxEntSubmissionBindingFromRow(submission)
	return ok && binding.engineAccountID == claim.engineAccountID && mailboxEntSubmissionClaimDraftMatches(draft, account.ID, claim.draftID, binding)
}

func mailboxEntSubmissionFinalizationLockExpectedAttachments(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, draft *ent.MailDraft) ([]MailboxSubmissionFinalizationExpectedAttachment, error) {
	attachments, err := mailboxEntSubmissionFinalizationLockedCleanAttachments(ctx, transaction, account, draft)
	if err != nil {
		return nil, err
	}
	result := make([]MailboxSubmissionFinalizationExpectedAttachment, len(attachments))
	for index, attachment := range attachments {
		result[index] = MailboxSubmissionFinalizationExpectedAttachment{
			AttachmentID: attachment.attachment.ID, ObjectID: attachment.object.ID, SHA256: append([]byte(nil), (*attachment.attachment.Sha256)...),
			SizeBytes: attachment.attachment.SizeBytes, ContentID: mailboxEntMessageOptionalString(attachment.attachment.ContentID),
			Filename: attachment.attachment.Filename, MediaType: attachment.attachment.MediaType, Disposition: EngineAttachmentDisposition(attachment.attachment.Disposition),
		}
	}
	return result, nil
}

type mailboxEntSubmissionFinalizationAttachment struct { attachment *ent.MailAttachment; object *ent.MailObject }

func mailboxEntSubmissionFinalizationLockedCleanAttachments(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, draft *ent.MailDraft) ([]mailboxEntSubmissionFinalizationAttachment, error) {
	if ctx == nil || transaction == nil || account == nil || draft == nil || account.ID == uuid.Nil || draft.ID == uuid.Nil {
		return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid)
	}
	rows, err := transaction.MailAttachment.Query().
		Where(mailattachment.Or(mailattachment.MailDraftIDEQ(draft.ID), mailattachment.PendingMailDraftIDEQ(draft.ID))).
		Order(ent.Asc(mailattachment.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable)
	}
	if len(rows) > maxMailboxSubmissionCleanAttachments {
		return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
	}
	objectIDs := make(map[uuid.UUID]struct{}, len(rows))
	for index, attachment := range rows {
		if !mailboxEntSubmissionCleanAttachmentRow(attachment, draft) || (index > 0 && !syncLessUUID(rows[index-1].ID, attachment.ID)) {
			return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
		}
		objectIDs[*attachment.StorageObjectID] = struct{}{}
	}
	orderedObjectIDs := mailboxEntApplySortedUUIDs(objectIDs)
	objects, err := transaction.MailObject.Query().Where(mailobject.IDIn(orderedObjectIDs...)).Order(ent.Asc(mailobject.FieldID)).ForUpdate().All(ctx)
	if err != nil {
		return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable)
	}
	if len(objects) != len(orderedObjectIDs) {
		return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
	}
	byID := make(map[uuid.UUID]*ent.MailObject, len(objects))
	for index, object := range objects {
		if object == nil || object.ID == uuid.Nil || (index > 0 && !syncLessUUID(objects[index-1].ID, object.ID)) {
			return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
		}
		if _, duplicate := byID[object.ID]; duplicate {
			return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
		}
		byID[object.ID] = object
	}
	result := make([]mailboxEntSubmissionFinalizationAttachment, 0, len(rows))
	for _, attachment := range rows {
		object := byID[*attachment.StorageObjectID]
		if !mailboxEntSubmissionCleanObjectRow(object, account, draft, attachment) {
			return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
		}
		result = append(result, mailboxEntSubmissionFinalizationAttachment{attachment: attachment, object: object})
	}
	return result, nil
}

func mailboxEntSubmissionFinalizationLockAttachments(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, draft *ent.MailDraft, claim MailboxSubmissionFinalizationClaim) ([]mailboxEntSubmissionFinalizationAttachment, error) {
	attachments, err := mailboxEntSubmissionFinalizationLockedCleanAttachments(ctx, transaction, account, draft)
	if err != nil {
		return nil, err
	}
	if !mailboxEntSubmissionFinalizationAttachmentsMatch(attachments, claim.expectedAttachments) {
		return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
	}
	return attachments, nil
}

func mailboxEntSubmissionFinalizationAttachmentsMatch(actual []mailboxEntSubmissionFinalizationAttachment, expected []mailboxSubmissionFinalizationExpectedAttachment) bool {
	if actual == nil || expected == nil || len(actual) != len(expected) {
		return false
	}
	for index, value := range actual {
		want := expected[index]
		if value.attachment == nil || value.object == nil || value.attachment.ID != want.attachmentID || value.object.ID != want.objectID ||
			value.attachment.Sha256 == nil || len(*value.attachment.Sha256) != sha256.Size || subtle.ConstantTimeCompare(*value.attachment.Sha256, want.sha256[:]) != 1 ||
			value.attachment.SizeBytes != want.sizeBytes || mailboxEntMessageOptionalString(value.attachment.ContentID) != want.contentID ||
			value.attachment.Filename != want.filename || value.attachment.MediaType != want.mediaType || EngineAttachmentDisposition(value.attachment.Disposition) != want.disposition ||
			value.attachment.EngineBlobID != nil || value.attachment.StorageObjectID == nil || *value.attachment.StorageObjectID != value.object.ID {
			return false
		}
	}
	return true
}

type mailboxEntSubmissionFinalizationAncestorMessage struct {
	id                uuid.UUID
	threadID          uuid.UUID
	internetMessageID string
}

type mailboxEntSubmissionFinalizationAncestorSnapshot struct {
	ancestors []ThreadAncestor
	messages  []mailboxEntSubmissionFinalizationAncestorMessage
}

func mailboxEntSubmissionFinalizationPreReadAncestors(ctx context.Context, transaction *ent.Tx, accountID uuid.UUID, projection EngineMessageProjection) (mailboxEntSubmissionFinalizationAncestorSnapshot, error) {
	candidates := mailboxEntSubmissionFinalizationAncestorCandidates(projection)
	if len(candidates) == 0 {
		return mailboxEntSubmissionFinalizationAncestorSnapshot{ancestors: []ThreadAncestor{}, messages: []mailboxEntSubmissionFinalizationAncestorMessage{}}, nil
	}
	rows, err := transaction.MailMessage.Query().
		Where(mailmessage.MailAccountIDEQ(accountID), mailmessage.InternetMessageIDIn(candidates...), mailmessage.DeletedAtIsNil()).
		Order(ent.Asc(mailmessage.FieldID)).
		All(ctx)
	if err != nil {
		return mailboxEntSubmissionFinalizationAncestorSnapshot{}, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable)
	}
	threadIDs, err := mailboxEntSubmissionFinalizationAncestorThreadIDs(accountID, rows)
	if err != nil {
		return mailboxEntSubmissionFinalizationAncestorSnapshot{}, err
	}
	threads, err := mailboxEntSubmissionFinalizationLoadThreads(ctx, transaction, accountID, threadIDs, false)
	if err != nil {
		return mailboxEntSubmissionFinalizationAncestorSnapshot{}, err
	}
	return mailboxEntSubmissionFinalizationAncestorsFromLockedRows(accountID, rows, threads)
}

func mailboxEntSubmissionFinalizationAncestorCandidates(projection EngineMessageProjection) []string {
	values := make(map[string]struct{}, len(projection.References)+1)
	for _, value := range projection.References { if validThreadMessageID(value) { values[value] = struct{}{} } }
	if validThreadMessageID(projection.InReplyTo) { values[projection.InReplyTo] = struct{}{} }
	result := make([]string, 0, len(values))
	for value := range values { result = append(result, value) }
	sort.Strings(result)
	return result
}

func mailboxEntSubmissionFinalizationAncestorThreadIDs(accountID uuid.UUID, messages []*ent.MailMessage) (map[uuid.UUID]struct{}, error) {
	threadIDs := make(map[uuid.UUID]struct{}, len(messages))
	seenMessages := make(map[string]struct{}, len(messages))
	for index, message := range messages {
		if !mailboxEntApplyMessageRowValid(message, accountID) || message.DeletedAt != nil || message.InternetMessageID == nil || message.MailThreadID == uuid.Nil || (index > 0 && !syncLessUUID(messages[index-1].ID, message.ID)) {
			return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
		}
		if _, duplicate := seenMessages[*message.InternetMessageID]; duplicate {
			return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
		}
		seenMessages[*message.InternetMessageID] = struct{}{}
		threadIDs[message.MailThreadID] = struct{}{}
	}
	return threadIDs, nil
}

func mailboxEntSubmissionFinalizationLoadThreads(ctx context.Context, transaction *ent.Tx, accountID uuid.UUID, threadIDs map[uuid.UUID]struct{}, locked bool) (map[uuid.UUID]*ent.MailThread, error) {
	if accountID == uuid.Nil || threadIDs == nil {
		return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid)
	}
	if len(threadIDs) == 0 {
		return map[uuid.UUID]*ent.MailThread{}, nil
	}
	ordered := mailboxEntApplySortedUUIDs(threadIDs)
	query := transaction.MailThread.Query().Where(mailthread.MailAccountIDEQ(accountID), mailthread.IDIn(ordered...)).Order(ent.Asc(mailthread.FieldID))
	if locked { query = query.ForUpdate() }
	rows, err := query.All(ctx)
	if err != nil {
		return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable)
	}
	if len(rows) != len(ordered) {
		return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
	}
	result := make(map[uuid.UUID]*ent.MailThread, len(rows))
	for _, row := range rows {
		if !mailboxEntApplyThreadRowValid(row, accountID) {
			return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
		}
		if _, duplicate := result[row.ID]; duplicate {
			return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
		}
		result[row.ID] = row
	}
	return result, nil
}

func mailboxEntSubmissionFinalizationAncestorsFromLockedRows(accountID uuid.UUID, messages []*ent.MailMessage, threads map[uuid.UUID]*ent.MailThread) (mailboxEntSubmissionFinalizationAncestorSnapshot, error) {
	if threads == nil {
		return mailboxEntSubmissionFinalizationAncestorSnapshot{}, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid)
	}
	threadIDs, err := mailboxEntSubmissionFinalizationAncestorThreadIDs(accountID, messages)
	if err != nil {
		return mailboxEntSubmissionFinalizationAncestorSnapshot{}, err
	}
	if len(threadIDs) != 0 && len(threads) == 0 {
		return mailboxEntSubmissionFinalizationAncestorSnapshot{}, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
	}
	result := mailboxEntSubmissionFinalizationAncestorSnapshot{
		ancestors: make([]ThreadAncestor, 0, len(messages)),
		messages:  make([]mailboxEntSubmissionFinalizationAncestorMessage, 0, len(messages)),
	}
	for _, message := range messages {
		thread := threads[message.MailThreadID]
		key, ok := mailboxEntApplyThreadKeyFromRow(thread)
		if !ok {
			return mailboxEntSubmissionFinalizationAncestorSnapshot{}, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
		}
		result.ancestors = append(result.ancestors, ThreadAncestor{MessageID: *message.InternetMessageID, KeyVersion: ThreadKeyVersion(key.version), Key: key.key})
		result.messages = append(result.messages, mailboxEntSubmissionFinalizationAncestorMessage{id: message.ID, threadID: message.MailThreadID, internetMessageID: *message.InternetMessageID})
	}
	sort.Slice(result.ancestors, func(left, right int) bool { return result.ancestors[left].MessageID < result.ancestors[right].MessageID })
	return result, nil
}

func mailboxEntSubmissionFinalizationLockThreads(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, ancestors mailboxEntSubmissionFinalizationAncestorSnapshot, targetThreadID uuid.UUID) (map[uuid.UUID]*ent.MailThread, error) {
	if account == nil || account.ID == uuid.Nil || targetThreadID == uuid.Nil {
		return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid)
	}
	threadIDs := make(map[uuid.UUID]struct{}, len(ancestors.messages)+1)
	threadIDs[targetThreadID] = struct{}{}
	for _, ancestor := range ancestors.messages {
		if ancestor.id == uuid.Nil || ancestor.threadID == uuid.Nil || ancestor.internetMessageID == "" {
			return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
		}
		threadIDs[ancestor.threadID] = struct{}{}
	}
	ordered := mailboxEntApplySortedUUIDs(threadIDs)
	rows, err := transaction.MailThread.Query().
		Where(mailthread.MailAccountIDEQ(account.ID), mailthread.IDIn(ordered...)).
		Order(ent.Asc(mailthread.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable)
	}
	result := make(map[uuid.UUID]*ent.MailThread, len(rows))
	for _, row := range rows {
		if !mailboxEntApplyThreadRowValid(row, account.ID) {
			return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
		}
		if _, duplicate := result[row.ID]; duplicate {
			return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
		}
		result[row.ID] = row
	}
	for _, ancestor := range ancestors.messages {
		if result[ancestor.threadID] == nil {
			return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
		}
	}
	return result, nil
}

func mailboxEntSubmissionFinalizationResolveLockedThread(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, threadID uuid.UUID, identity ThreadIdentity, projection EngineMessageProjection, databaseNow time.Time, lockedThreads map[uuid.UUID]*ent.MailThread) (*ent.MailThread, error) {
	if account == nil || threadID == uuid.Nil || identity.KeyVersion != ThreadKeyVersionV1 || threadKeyIsZero(identity.Key) || lockedThreads == nil {
		return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid)
	}
	thread := lockedThreads[threadID]
	if thread == nil {
		if identity.Source != ThreadIdentitySourceFallback { return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired) }
		participants := schema.MailThreadParticipants(append([]string(nil), identity.CanonicalParticipants...))
		if participants.Validate() != nil { return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired) }
		created, createErr := transaction.MailThread.Create().
			SetID(threadID).SetMailAccountID(account.ID).SetThreadKeyVersion(int32(identity.KeyVersion)).SetThreadKey(identity.Key[:]).
			SetNormalizedSubject(identity.NormalizedSubject).SetParticipantAddresses(participants).
			SetLastMessageAt(mailboxEntSubmissionFinalizationMessageTime(projection, databaseNow)).SetMessageCount(1).
			SetUnreadCount(mailboxEntSubmissionFinalizationUnread(projection)).SetIsPinned(false).
			SetHasAttachments(len(projection.Attachments) != 0).SetSnippet(projection.Snippet).
			SetLastSeenSyncEpoch(account.SyncEpoch).SetCreatedAt(databaseNow).SetUpdatedAt(databaseNow).Save(ctx)
		if createErr != nil { return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable) }
		return created, nil
	}
	if !mailboxEntApplyThreadRowValid(thread, account.ID) || thread.ID != threadID || thread.ThreadKeyVersion != int32(identity.KeyVersion) || subtle.ConstantTimeCompare(thread.ThreadKey, identity.Key[:]) != 1 {
		return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
	}
	if identity.Source == ThreadIdentitySourceFallback && !mailboxEntApplyThreadFallbackMetadataMatches(thread, identity) {
		return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
	}
	return thread, nil
}

func mailboxEntSubmissionFinalizationLockMessages(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, messageID uuid.UUID, projection EngineMessageProjection) ([]*ent.MailMessage, []*ent.MailMessage, error) {
	if account == nil || account.ID == uuid.Nil || messageID == uuid.Nil {
		return nil, nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid)
	}
	ancestorCandidates := mailboxEntSubmissionFinalizationAncestorCandidates(projection)
	query := transaction.MailMessage.Query().Where(mailmessage.MailAccountIDEQ(account.ID))
	if len(ancestorCandidates) == 0 {
		query = query.Where(mailmessage.Or(mailmessage.IDEQ(messageID), mailmessage.EngineMessageIDEQ(projection.MessageID)))
	} else {
		query = query.Where(mailmessage.Or(
			mailmessage.IDEQ(messageID),
			mailmessage.EngineMessageIDEQ(projection.MessageID),
			mailmessage.And(mailmessage.InternetMessageIDIn(ancestorCandidates...), mailmessage.DeletedAtIsNil()),
		))
	}
	rows, err := query.Order(ent.Asc(mailmessage.FieldID)).ForUpdate().All(ctx)
	if err != nil {
		return nil, nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable)
	}
	targets := make([]*ent.MailMessage, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			return nil, nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
		}
		if row.ID == messageID || row.EngineMessageID == projection.MessageID {
			targets = append(targets, row)
		}
	}
	return rows, targets, nil
}

func mailboxEntSubmissionFinalizationRevalidateAncestors(accountID uuid.UUID, projection EngineMessageProjection, expected mailboxEntSubmissionFinalizationAncestorSnapshot, threads map[uuid.UUID]*ent.MailThread, lockedMessages []*ent.MailMessage) error {
	candidates := mailboxEntSubmissionFinalizationAncestorCandidates(projection)
	if len(candidates) == 0 {
		if len(expected.ancestors) == 0 && len(expected.messages) == 0 { return nil }
		return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
	}
	candidateSet := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates { candidateSet[candidate] = struct{}{} }
	ancestorRows := make([]*ent.MailMessage, 0, len(expected.messages))
	for _, message := range lockedMessages {
		if message == nil {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
		}
		if message.DeletedAt != nil || message.InternetMessageID == nil {
			continue
		}
		if _, matches := candidateSet[*message.InternetMessageID]; matches {
			ancestorRows = append(ancestorRows, message)
		}
	}
	actual, err := mailboxEntSubmissionFinalizationAncestorsFromLockedRows(accountID, ancestorRows, threads)
	if err != nil {
		return err
	}
	if len(actual.ancestors) != len(expected.ancestors) || len(actual.messages) != len(expected.messages) {
		return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
	}
	for index := range actual.ancestors {
		if actual.ancestors[index] != expected.ancestors[index] {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
		}
	}
	for index := range actual.messages {
		if actual.messages[index] != expected.messages[index] {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
		}
	}
	return nil
}

func mailboxEntSubmissionFinalizationUnread(projection EngineMessageProjection) int32 { if projection.Flags.Seen { return 0 }; return 1 }

func mailboxEntSubmissionFinalizationMessageTime(projection EngineMessageProjection, fallback time.Time) time.Time {
	if projection.SentAt != nil { return *projection.SentAt }
	if projection.ReceivedAt != nil { return *projection.ReceivedAt }
	return fallback
}

func mailboxEntSubmissionFinalizationUpsertMessage(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, thread *ent.MailThread, messageID uuid.UUID, rows []*ent.MailMessage, projection EngineMessageProjection, databaseNow time.Time) (*ent.MailMessage, error) {
	if account == nil || thread == nil || messageID == uuid.Nil || syncValidateProjection(projection) != nil || projection.Origin != EngineProjectionOriginConfirmedOwnerSubmission || projection.Flags.Draft {
		return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
	}
	references := schema.MailMessageReferences(append([]string(nil), projection.References...))
	from := schema.MailStoredAddress{Email: projection.From.Email, Name: projection.From.Name}
	to, err := mailboxEntApplyStoredAddresses(projection.To); if err != nil { return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired) }
	cc, err := mailboxEntApplyStoredAddresses(projection.CC); if err != nil { return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired) }
	bcc, err := mailboxEntApplyStoredAddresses(projection.BCC); if err != nil { return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired) }
	replyTo, err := mailboxEntApplyStoredAddresses(projection.ReplyTo); if err != nil { return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired) }
	if references.Validate() != nil || from.Validate() != nil { return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired) }
	var existing *ent.MailMessage
	for _, row := range rows {
		if !mailboxEntApplyMessageRowValid(row, account.ID) || row.DeletedAt != nil || (row.ID != messageID && row.EngineMessageID != projection.MessageID) || existing != nil {
			return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
		}
		existing = row
	}
	if existing == nil {
		create := transaction.MailMessage.Create().
			SetID(messageID).SetMailAccountID(account.ID).SetMailThreadID(thread.ID).SetEngineMessageID(projection.MessageID).SetEngineVersion(projection.Version).
			SetReferences(references).SetFrom(from).SetTo(to).SetCc(cc).SetBcc(bcc).SetReplyTo(replyTo).SetSubject(projection.Subject).SetSnippet(projection.Snippet).
			SetSizeBytes(projection.SizeBytes).SetPriority(mailmessage.Priority(projection.Priority)).SetHasAttachments(false).
			SetIsDraft(false).SetIsSeen(projection.Flags.Seen).SetIsAnswered(projection.Flags.Answered).SetIsFlagged(projection.Flags.Flagged).SetIsJunk(projection.Flags.Junk).
			SetProjectionOrigin(mailmessage.ProjectionOrigin(EngineProjectionOriginConfirmedOwnerSubmission)).SetAuthenticationClass(mailmessage.AuthenticationClass(projection.AuthenticationClass)).
			SetFlagsSyncedAt(databaseNow).SetLastSeenSyncEpoch(account.SyncEpoch).SetCreatedAt(databaseNow).SetUpdatedAt(databaseNow)
		mailboxEntSubmissionFinalizationSetMessageOptionalsCreate(create, projection)
		created, err := create.Save(ctx)
		if err != nil { return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable) }
		return created, nil
	}
	if existing.ID != messageID || existing.EngineMessageID != projection.MessageID || existing.ReceivedEventID != nil {
		return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
	}
	update := existing.Update().SetMailThreadID(thread.ID).SetEngineVersion(projection.Version).SetReferences(references).SetFrom(from).SetTo(to).SetCc(cc).SetBcc(bcc).SetReplyTo(replyTo).SetSubject(projection.Subject).SetSnippet(projection.Snippet).
		SetSizeBytes(projection.SizeBytes).SetPriority(mailmessage.Priority(projection.Priority)).SetIsDraft(false).SetIsSeen(projection.Flags.Seen).SetIsAnswered(projection.Flags.Answered).SetIsFlagged(projection.Flags.Flagged).SetIsJunk(projection.Flags.Junk).
		SetProjectionOrigin(mailmessage.ProjectionOrigin(EngineProjectionOriginConfirmedOwnerSubmission)).SetAuthenticationClass(mailmessage.AuthenticationClass(projection.AuthenticationClass)).SetFlagsSyncedAt(databaseNow).SetLastSeenSyncEpoch(account.SyncEpoch).SetUpdatedAt(databaseNow)
	mailboxEntSubmissionFinalizationSetMessageOptionalsUpdate(update, projection)
	if err := update.Exec(ctx); err != nil { return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable) }
	return existing, nil
}

func mailboxEntSubmissionFinalizationSetMessageOptionalsCreate(create *ent.MailMessageCreate, projection EngineMessageProjection) {
	if projection.InternetMessageID != "" { create.SetInternetMessageID(projection.InternetMessageID) }
	if projection.InReplyTo != "" { create.SetInReplyTo(projection.InReplyTo) }
	if projection.SentAt != nil { create.SetSentAt(*projection.SentAt) }
	if projection.ReceivedAt != nil { create.SetReceivedAt(*projection.ReceivedAt) }
	if projection.CalendarMethod != nil { create.SetCalendarMethod(mailmessage.CalendarMethod(*projection.CalendarMethod)); create.SetCalendarPartLocatorHash(syncCloneBytes(projection.CalendarPartLocatorHash)) }
}

func mailboxEntSubmissionFinalizationSetMessageOptionalsUpdate(update *ent.MailMessageUpdateOne, projection EngineMessageProjection) {
	if projection.InternetMessageID == "" { update.ClearInternetMessageID() } else { update.SetInternetMessageID(projection.InternetMessageID) }
	if projection.InReplyTo == "" { update.ClearInReplyTo() } else { update.SetInReplyTo(projection.InReplyTo) }
	if projection.SentAt == nil { update.ClearSentAt() } else { update.SetSentAt(*projection.SentAt) }
	if projection.ReceivedAt == nil { update.ClearReceivedAt() } else { update.SetReceivedAt(*projection.ReceivedAt) }
	if projection.CalendarMethod == nil { update.ClearCalendarMethod(); update.ClearCalendarPartLocatorHash() } else { update.SetCalendarMethod(mailmessage.CalendarMethod(*projection.CalendarMethod)); update.SetCalendarPartLocatorHash(syncCloneBytes(projection.CalendarPartLocatorHash)) }
}

func mailboxEntSubmissionFinalizationTransferAttachments(ctx context.Context, messageID uuid.UUID, attachments []mailboxEntSubmissionFinalizationAttachment, databaseNow time.Time) error {
	if messageID == uuid.Nil || attachments == nil { return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid) }
	for _, value := range attachments {
		if value.attachment == nil || value.object == nil || value.attachment.StorageObjectID == nil || *value.attachment.StorageObjectID != value.object.ID || value.attachment.EngineBlobID != nil {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
		}
		if err := value.attachment.Update().ClearMailDraftID().SetMailMessageID(messageID).SetUpdatedAt(databaseNow).Exec(ctx); err != nil {
			return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable)
		}
	}
	return nil
}

func mailboxEntSubmissionFinalizationResolveFoldersAndLabels(ctx context.Context, transaction *ent.Tx, accountID uuid.UUID, projection EngineMessageProjection) ([]MailboxEventID, []MailboxSyncPolicyLabelState, error) {
	if projection.FolderIDs == nil || projection.ReservedKeywords == nil { return nil, nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired) }
	folderKeys := append([]string(nil), projection.FolderIDs...); sort.Strings(folderKeys)
	for index, value := range folderKeys { if !syncValidOpaque(value, MaxEngineOpaqueIDBytes) || (index > 0 && folderKeys[index-1] == value) { return nil, nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired) } }
	folders := make([]MailboxEventID, 0, len(folderKeys))
	if len(folderKeys) != 0 {
		rows, err := transaction.MailFolder.Query().Where(mailfolder.MailAccountIDEQ(accountID), mailfolder.EngineFolderIDIn(folderKeys...), mailfolder.ControlStateEQ(mailfolder.ControlStateActive), mailfolder.DeletedAtIsNil()).Order(ent.Asc(mailfolder.FieldID)).ForUpdate().All(ctx)
		if err != nil { return nil, nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable) }
		if len(rows) != len(folderKeys) { return nil, nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired) }
		for _, row := range rows { if row == nil || row.ID == uuid.Nil || row.MailAccountID != accountID || row.EngineFolderID == nil || row.DeletedAt != nil || row.ControlState != mailfolder.ControlStateActive { return nil, nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired) }; folders = append(folders, MailboxEventID{AccountID: accountID, ID: row.ID}) }
		sort.Slice(folders, func(left, right int) bool { return syncLessUUID(folders[left].ID, folders[right].ID) })
	}
	keywords := make([]string, len(projection.ReservedKeywords)); for index, value := range projection.ReservedKeywords { keywords[index] = string(value) }; sort.Strings(keywords)
	keywordSet := make(map[string]struct{}, len(keywords))
	for index, value := range keywords {
		if !syncValidOpaque(value, MaxEngineOpaqueIDBytes) || (index > 0 && keywords[index-1] == value) { return nil, nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired) }
		keywordSet[value] = struct{}{}
	}
	labels := make([]MailboxSyncPolicyLabelState, 0, len(keywords))
	if len(keywords) != 0 {
		rows, err := transaction.MailLabel.Query().Where(maillabel.MailAccountIDEQ(accountID), maillabel.EngineKeywordIn(keywords...), maillabel.DeletedAtIsNil()).Order(ent.Asc(maillabel.FieldID)).ForUpdate().All(ctx)
		if err != nil { return nil, nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable) }
		seenLabels := make(map[uuid.UUID]struct{}, len(rows))
		seenKeywords := make(map[string]struct{}, len(rows))
		for _, row := range rows {
			if row == nil || row.ID == uuid.Nil || row.MailAccountID != accountID || row.DeletedAt != nil || !syncValidOpaque(row.EngineKeyword, MaxEngineOpaqueIDBytes) { return nil, nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired) }
			if _, known := keywordSet[row.EngineKeyword]; !known { return nil, nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired) }
			if _, duplicate := seenLabels[row.ID]; duplicate { return nil, nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired) }
			if _, duplicate := seenKeywords[row.EngineKeyword]; duplicate { return nil, nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired) }
			seenLabels[row.ID] = struct{}{}
			seenKeywords[row.EngineKeyword] = struct{}{}
			labels = append(labels, MailboxSyncPolicyLabelState{Label: MailboxEventID{AccountID: accountID, ID: row.ID}, EngineRule: true})
		}
		sort.Slice(labels, func(left, right int) bool { return syncLessUUID(labels[left].Label.ID, labels[right].Label.ID) })
	}
	return folders, labels, nil
}

func mailboxEntSubmissionFinalizationResolveRoutes(ctx context.Context, transaction *ent.Tx, sender *ent.MailAccount, projection EngineMessageProjection) ([]MailboxSubmissionSentRecipientResolution, error) {
	if sender == nil { return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid) }
	emails := make([]string, 0, len(projection.To)+len(projection.CC)+len(projection.BCC))
	seen := make(map[string]struct{}, cap(emails))
	domainNameSet := make(map[string]struct{}, cap(emails))
	for _, group := range [][]EngineAddress{projection.To, projection.CC, projection.BCC} { for _, recipient := range group { normalized, err := NormalizeAddress(recipient); if err != nil || normalized.Email != recipient.Email { return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired) }; if _, duplicate := seen[recipient.Email]; duplicate { return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired) }; at := strings.LastIndexByte(recipient.Email, '@'); if at <= 0 || at == len(recipient.Email)-1 { return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired) }; seen[recipient.Email] = struct{}{}; domainNameSet[recipient.Email[at+1:]] = struct{}{}; emails = append(emails, recipient.Email) } }
	if len(emails) == 0 { return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired) }
	sort.Strings(emails)
	domainNames := make([]string, 0, len(domainNameSet)); for domainName := range domainNameSet { domainNames = append(domainNames, domainName) }; sort.Strings(domainNames)
	identities, err := transaction.MailIdentity.Query().Where(mailidentity.NormalizedAddressIn(emails...), mailidentity.StatusEQ(mailidentity.Status("active"))).Order(ent.Asc(mailidentity.FieldID)).All(ctx)
	if err != nil { return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable) }
	identityByEmail := make(map[string]*ent.MailIdentity, len(identities)); accountIDs := make(map[uuid.UUID]struct{}, len(identities))
	for _, identity := range identities { if identity == nil || identity.ID == uuid.Nil || identity.MailAccountID == uuid.Nil || identity.Status != mailidentity.Status("active") { return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired) }; if _, requested := seen[identity.NormalizedAddress]; !requested { return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired) }; if _, duplicate := identityByEmail[identity.NormalizedAddress]; duplicate { return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired) }; identityByEmail[identity.NormalizedAddress] = identity; accountIDs[identity.MailAccountID] = struct{}{} }
	recipientAccounts := map[uuid.UUID]*ent.MailAccount{}
	if len(accountIDs) != 0 { ids := mailboxEntApplySortedUUIDs(accountIDs); rows, err := transaction.MailAccount.Query().Where(mailaccount.IDIn(ids...), mailaccount.StatusEQ(mailaccount.StatusActive), mailaccount.DesiredEnabledEQ(true)).Order(ent.Asc(mailaccount.FieldID)).All(ctx); if err != nil { return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable) }; if len(rows) != len(ids) { return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired) }; for _, row := range rows { if row == nil || row.ID == uuid.Nil || row.OrganizationID == uuid.Nil || row.MailDomainID == uuid.Nil || row.Status != mailaccount.StatusActive || !row.DesiredEnabled { return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired) }; if _, duplicate := recipientAccounts[row.ID]; duplicate { return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired) }; recipientAccounts[row.ID] = row } }
	domains, err := transaction.MailDomain.Query().Where(maildomain.DomainIn(domainNames...), maildomain.DesiredPresentEQ(true), maildomain.EnabledEQ(true), maildomain.StateEQ(maildomain.State("verified")), maildomain.VerificationStatusEQ(maildomain.VerificationStatus("verified")), maildomain.DeletedAtIsNil()).Order(ent.Asc(maildomain.FieldID)).All(ctx)
	if err != nil { return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable) }
	domainByName := make(map[string]*ent.MailDomain, len(domains)); for _, domain := range domains { if domain == nil || domain.ID == uuid.Nil || domain.OrganizationID == uuid.Nil || domain.DeletedAt != nil || !domain.DesiredPresent || !domain.Enabled || domain.State != maildomain.State("verified") || domain.VerificationStatus != maildomain.VerificationStatus("verified") { return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired) }; if _, requested := domainNameSet[domain.Domain]; !requested { return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired) }; if _, duplicate := domainByName[domain.Domain]; duplicate { return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired) }; domainByName[domain.Domain] = domain }
	result := make([]MailboxSubmissionSentRecipientResolution, 0, len(emails))
	for _, email := range emails {
		domain := domainByName[email[strings.LastIndexByte(email, '@')+1:]]
		identity := identityByEmail[email]
		class := MailDeliveryClassExternal
		if identity != nil {
			recipient := recipientAccounts[identity.MailAccountID]
			if recipient == nil { return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired) }
			if recipient.OrganizationID == sender.OrganizationID {
				if domain == nil || domain.OrganizationID != sender.OrganizationID || recipient.MailDomainID != domain.ID { return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired) }
				class = MailDeliveryClassInternal
			} else {
				class = MailDeliveryClassInstanceDomain
			}
		} else if domain != nil { class = MailDeliveryClassInstanceDomain }
		result = append(result, MailboxSubmissionSentRecipientResolution{Email: email, DeliveryClass: class})
	}
	return result, nil
}

func mailboxEntSubmissionFinalizationAppendOutbox(ctx context.Context, transaction *ent.Tx, event any, databaseNow time.Time) error {
	envelope, ok := event.(interface { EventID() uuid.UUID; Subject() string; MarshalJSON() ([]byte, error) })
	if !ok || envelope.EventID() == uuid.Nil || envelope.Subject() == "" { return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid) }
	rows, err := transaction.OutboxEvent.Query().Where(outboxevent.IDEQ(envelope.EventID())).All(ctx)
	if err != nil { return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable) }
	if len(rows) != 0 { return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired) }
	payload, err := envelope.MarshalJSON(); if err != nil || len(payload) == 0 || !json.Valid(payload) { return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired) }
	if _, err := transaction.OutboxEvent.Create().SetID(envelope.EventID()).SetEventType(envelope.Subject()).SetPayload(append([]byte(nil), payload...)).SetCreatedAt(databaseNow).Save(ctx); err != nil {
		return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable)
	}
	return nil
}

func mailboxEntSubmissionFinalizationCompletionLookupError(err error) error {
	if ent.IsNotFound(err) { return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired) }
	return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable)
}

func mailboxEntSubmissionFinalizationRepositoryError(err error) error {
	var finalization *MailboxSubmissionFinalizationError
	if errors.As(err, &finalization) && finalization != nil { return newMailboxSubmissionFinalizationError(finalization.Kind()) }
	var repositoryError *MailboxSubmissionRepositoryError
	if errors.As(err, &repositoryError) && repositoryError != nil {
		switch repositoryError.Kind() {
		case MailboxSubmissionRepositoryErrorInvalid: return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid)
		case MailboxSubmissionRepositoryErrorVersionConflict: return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
		}
	}
	return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable)
}
