package mail

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/mailfolder"
	"github.com/valerii/hub/backend/ent/mailmessage"
	"github.com/valerii/hub/backend/ent/mailmutationreceipt"
	"github.com/valerii/hub/backend/ent/mailmutationreceiptmember"
	"github.com/valerii/hub/backend/ent/mailmutationreceipttarget"
	"github.com/valerii/hub/backend/ent/mailthread"
)

const (
	mailboxEntMessageMutationOperation           = "message_bulk"
	mailboxEntMessageMutationReplayLifetime      = 30 * 24 * time.Hour
	mailboxEntMessageMutationUnknownRetryDelay   = 60 * time.Second
	mailboxEntMessageMutationBoundedBusyDuration = 60 * time.Second
)

var _ MailboxMessageMutationRepository = (*MailboxEntRepository)(nil)

type mailboxEntMessageMutationGraph struct {
	targets         []*ent.MailMutationReceiptTarget
	members         []*ent.MailMutationReceiptMember
	membersByTarget map[uuid.UUID][]uuid.UUID
	result          MailboxMessageMutationResult
}

// LookupMailboxMessageMutationReplay locks the account before the receipt. An
// existing in-progress receipt is always replayed as busy, including after its
// command lease expires; this path never creates or reclaims dispatch authority.
func (repository *MailboxEntRepository) LookupMailboxMessageMutationReplay(ctx context.Context, request MailboxMessageMutationReplayRequest) (MailboxMessageMutationReplay, error) {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxEntMessageMutationReplayRequestValid(request) {
		return MailboxMessageMutationReplay{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorInvalid)
	}

	var replay MailboxMessageMutationReplay
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, databaseNow, err := mailboxEntMessageMutationLockAccount(ctx, transaction, request.Scope)
		if err != nil {
			return err
		}
		receipt, err := mailboxEntMessageMutationLockReceipt(ctx, transaction, request.Scope, request.IdempotencyKeyHash)
		if err != nil {
			return err
		}
		if receipt != nil {
			resolved, err := mailboxEntMessageMutationReplayFromReceipt(ctx, transaction, account, databaseNow, request, receipt)
			if err != nil {
				return err
			}
			replay = resolved
			return nil
		}

		busyUntil, err := mailboxEntMessageMutationFreshBusyUntil(ctx, transaction, account, databaseNow)
		if err != nil {
			return err
		}
		if busyUntil != nil {
			replay = MailboxMessageMutationReplay{Kind: MailboxMessageMutationReplayInProgress, BusyUntil: busyUntil}
			return nil
		}
		replay = MailboxMessageMutationReplay{Kind: MailboxMessageMutationReplayAbsent}
		return nil
	})
	if err != nil {
		return MailboxMessageMutationReplay{}, mailboxEntMessageMutationSafeError(err)
	}
	return replay, nil
}

// ResolveMailboxMessageMutationTargets expands one owner-scoped request under
// stable account/folder/thread/message locks. It performs no receipt write.
func (repository *MailboxEntRepository) ResolveMailboxMessageMutationTargets(ctx context.Context, request MailboxMessageMutationResolveRequest) (MailboxMessageMutationResolution, error) {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxMessageMutationScopeValid(request.Scope) || !mailboxMessageMutationInputValid(request.Input) {
		return MailboxMessageMutationResolution{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorInvalid)
	}

	var resolution MailboxMessageMutationResolution
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, _, err := mailboxEntMessageMutationLockAccount(ctx, transaction, request.Scope)
		if err != nil {
			return err
		}
		resolved, err := mailboxEntMessageMutationResolveLocked(ctx, transaction, account, request.Input)
		if err != nil {
			return err
		}
		resolution = mailboxMessageMutationResolutionCopy(resolved)
		return nil
	})
	if err != nil {
		return MailboxMessageMutationResolution{}, mailboxEntMessageMutationSafeError(err)
	}
	return mailboxMessageMutationResolutionCopy(resolution), nil
}

// ReserveMailboxMessageMutation is the only method that can create provider
// dispatch authority. Receipt, ordered snapshot rows and both copies of the
// command lease are committed in the same account-first transaction.
func (repository *MailboxEntRepository) ReserveMailboxMessageMutation(ctx context.Context, request MailboxMessageMutationReserveRequest) (MailboxMessageMutationReservation, error) {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxEntMessageMutationReserveRequestValid(request) {
		return MailboxMessageMutationReservation{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorInvalid)
	}

	var reservation MailboxMessageMutationReservation
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, databaseNow, err := mailboxEntMessageMutationLockAccount(ctx, transaction, request.Scope)
		if err != nil {
			return err
		}
		receipt, err := mailboxEntMessageMutationLockReceipt(ctx, transaction, request.Scope, request.IdempotencyKeyHash)
		if err != nil {
			return err
		}
		if receipt != nil {
			resolved, err := mailboxEntMessageMutationReservationFromReceipt(ctx, transaction, account, databaseNow, request, receipt)
			if err != nil {
				return err
			}
			reservation = resolved
			return nil
		}

		busyUntil, err := mailboxEntMessageMutationFreshBusyUntil(ctx, transaction, account, databaseNow)
		if err != nil {
			return err
		}
		if busyUntil != nil {
			reservation = MailboxMessageMutationReservation{Kind: MailboxMessageMutationReservationBusy, BusyUntil: busyUntil}
			return nil
		}

		lockedResolution, err := mailboxEntMessageMutationResolveLocked(ctx, transaction, account, request.Input)
		if err != nil {
			return err
		}
		if !mailboxEntMessageMutationResolutionEqual(lockedResolution, request.Resolution, request.Input) {
			return mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorStateConflict)
		}
		if lockedResolution.Kind == MailboxMessageMutationResolutionExpansionLimit {
			if _, err := mailboxEntMessageMutationCreateTerminalReceipt(ctx, transaction, account, databaseNow, request, mailmutationreceipt.State("failed")); err != nil {
				return err
			}
			reservation = MailboxMessageMutationReservation{Kind: MailboxMessageMutationReservationExpansionLimit}
			return nil
		}
		if lockedResolution.Snapshot == nil {
			return mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
		}
		if len(lockedResolution.Snapshot.Messages) == 0 {
			result, err := mailboxEntMessageMutationCreateTerminalReceipt(ctx, transaction, account, databaseNow, request, mailmutationreceipt.State("completed"))
			if err != nil {
				return err
			}
			reservation = MailboxMessageMutationReservation{Kind: MailboxMessageMutationReservationCompleted, Result: result}
			return nil
		}

		leaseUntil, ok := mailboxEntCheckedTimeAdd(databaseNow, MailboxEngineCommandLeaseDuration)
		if !ok {
			return mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
		}
		nextMutationEpoch, ok := mailboxEntNextEpoch(account.MutationEpoch)
		if !ok {
			return mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
		}
		leaseID := uuid.New()
		fence := MailboxEngineCommandFence{AccountID: account.ID, LeaseID: leaseID, MutationEpoch: nextMutationEpoch, LeaseUntil: leaseUntil}
		if !mailboxEntEngineCommandFenceValid(fence) {
			return mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
		}
		receipt, err = mailboxEntMessageMutationCreateReceipt(ctx, transaction, account, databaseNow, request, mailmutationreceipt.State("in_progress"), &fence)
		if err != nil {
			return err
		}
		if err := mailboxEntMessageMutationCreateSnapshotRows(ctx, transaction, receipt, *lockedResolution.Snapshot, databaseNow, true); err != nil {
			return err
		}
		if err := account.Update().
			SetMutationLeaseToken(leaseID).
			SetMutationLeaseUntil(leaseUntil).
			SetMutationEpoch(nextMutationEpoch).
			SetSyncScheduleUpdatedAt(databaseNow).
			Exec(ctx); err != nil {
			return mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
		}
		snapshot := mailboxMessageMutationSnapshotCopy(*lockedResolution.Snapshot)
		fenceCopy := fence
		reservation = MailboxMessageMutationReservation{
			Kind: MailboxMessageMutationReservationReserved, ReceiptID: receipt.ID,
			Fence: &fenceCopy, EngineAccountID: request.Scope.EngineAccountID, Snapshot: &snapshot,
		}
		return nil
	})
	if err != nil {
		return MailboxMessageMutationReservation{}, mailboxEntMessageMutationSafeError(err)
	}
	return reservation, nil
}

// FinishMailboxMessageMutation records only the provider outcome in the durable
// receipt graph. It deliberately does not mutate MailMessage projection rows.
func (repository *MailboxEntRepository) FinishMailboxMessageMutation(ctx context.Context, request MailboxMessageMutationFinishRequest) (MailboxMessageMutationResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxEntMessageMutationFinishRequestValid(request) {
		return MailboxMessageMutationResult{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorInvalid)
	}

	var result MailboxMessageMutationResult
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, databaseNow, err := mailboxEntMessageMutationLockAccount(ctx, transaction, request.Scope)
		if err != nil {
			return err
		}
		receipt, graph, err := mailboxEntMessageMutationLockFinishGraph(ctx, transaction, account, databaseNow, request.ReceiptID, request.Fence)
		if err != nil {
			return err
		}
		if err := mailboxEntMessageMutationWriteTargetOutcome(ctx, graph, databaseNow, request.Outcome); err != nil {
			return err
		}
		replayUntil, ok := mailboxEntCheckedTimeAdd(databaseNow, mailboxEntMessageMutationReplayLifetime)
		if !ok {
			return mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
		}
		if err := receipt.Update().
			SetState(mailmutationreceipt.State("completed")).
			ClearSafeCode().
			SetOriginalHTTPStatus(200).
			SetReconcileAttempts(0).
			ClearNextReconcileAt().
			ClearReconcileLeaseToken().
			ClearReconcileLeaseUntil().
			SetTerminalAt(databaseNow).
			SetReplayUntil(replayUntil).
			ClearCompactedAt().
			SetUpdatedAt(databaseNow).
			Exec(ctx); err != nil {
			return mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
		}
		if err := mailboxEntMessageMutationReleaseAccount(ctx, account, request.Fence, databaseNow); err != nil {
			return err
		}
		result = mailboxEntMessageMutationResultFromGraph(graph, MailboxMessageMutationResultCompleted)
		return nil
	})
	if err != nil {
		return MailboxMessageMutationResult{}, mailboxEntMessageMutationSafeError(err)
	}
	return mailboxMessageMutationResultCopy(result), nil
}

// MarkMailboxMessageMutationUnknown seals the dispatched receipt as durable
// unknown and releases its account command fence without claiming projection success.
func (repository *MailboxEntRepository) MarkMailboxMessageMutationUnknown(ctx context.Context, request MailboxMessageMutationUnknownRequest) (MailboxMessageMutationResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxEntMessageMutationUnknownRequestValid(request) {
		return MailboxMessageMutationResult{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorInvalid)
	}

	var result MailboxMessageMutationResult
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, databaseNow, err := mailboxEntMessageMutationLockAccount(ctx, transaction, request.Scope)
		if err != nil {
			return err
		}
		receipt, graph, err := mailboxEntMessageMutationLockFinishGraph(ctx, transaction, account, databaseNow, request.ReceiptID, request.Fence)
		if err != nil {
			return err
		}
		if err := mailboxEntMessageMutationWriteUnknownTargets(ctx, graph, databaseNow); err != nil {
			return err
		}
		nextReconcileAt, ok := mailboxEntCheckedTimeAdd(databaseNow, mailboxEntMessageMutationUnknownRetryDelay)
		if !ok {
			return mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
		}
		if err := receipt.Update().
			SetState(mailmutationreceipt.State("unknown")).
			SetSafeCode(mailmutationreceipt.SafeCode(MailboxMessageMutationCodeOutcomeUnknown)).
			ClearOriginalHTTPStatus().
			SetReconcileAttempts(1).
			SetNextReconcileAt(nextReconcileAt).
			ClearReconcileLeaseToken().
			ClearReconcileLeaseUntil().
			ClearTerminalAt().
			ClearReplayUntil().
			ClearCompactedAt().
			SetUpdatedAt(databaseNow).
			Exec(ctx); err != nil {
			return mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
		}
		if err := mailboxEntMessageMutationReleaseAccount(ctx, account, request.Fence, databaseNow); err != nil {
			return err
		}
		result = mailboxEntMessageMutationResultFromGraph(graph, MailboxMessageMutationResultUnknown)
		return nil
	})
	if err != nil {
		return MailboxMessageMutationResult{}, mailboxEntMessageMutationSafeError(err)
	}
	return mailboxMessageMutationResultCopy(result), nil
}

func mailboxEntMessageMutationReplayRequestValid(request MailboxMessageMutationReplayRequest) bool {
	return mailboxMessageMutationScopeValid(request.Scope) && request.FingerprintVersion == mailboxMessageMutationFingerprintVersion && mailboxEntMessageMutationDigestValid(request.IdempotencyKeyHash) && mailboxEntMessageMutationDigestValid(request.RequestHash)
}

func mailboxEntMessageMutationReserveRequestValid(request MailboxMessageMutationReserveRequest) bool {
	if !mailboxMessageMutationScopeValid(request.Scope) || !mailboxMessageMutationInputValid(request.Input) || request.FingerprintVersion != mailboxMessageMutationFingerprintVersion || !mailboxEntMessageMutationDigestValid(request.IdempotencyKeyHash) || !mailboxEntMessageMutationDigestValid(request.RequestHash) || !mailboxMessageMutationResolutionValid(request.Resolution, request.Input) {
		return false
	}
	return sha256.Sum256(request.Input.IdempotencyKey[:]) == request.IdempotencyKeyHash && mailboxMessageMutationRequestHash(request.Input, request.Scope) == request.RequestHash
}

func mailboxEntMessageMutationFinishRequestValid(request MailboxMessageMutationFinishRequest) bool {
	return mailboxMessageMutationScopeValid(request.Scope) && request.ReceiptID != uuid.Nil && mailboxEntEngineCommandFenceValid(request.Fence) && request.Fence.AccountID == request.Scope.AccountID && (request.Outcome == MailboxMessageMutationFinishApplied || request.Outcome == MailboxMessageMutationFinishRejected)
}

func mailboxEntMessageMutationUnknownRequestValid(request MailboxMessageMutationUnknownRequest) bool {
	return mailboxMessageMutationScopeValid(request.Scope) && request.ReceiptID != uuid.Nil && mailboxEntEngineCommandFenceValid(request.Fence) && request.Fence.AccountID == request.Scope.AccountID
}

func mailboxEntMessageMutationDigestValid(value [sha256.Size]byte) bool {
	for _, item := range value {
		if item != 0 {
			return true
		}
	}
	return false
}

func mailboxEntMessageMutationLockAccount(ctx context.Context, transaction *ent.Tx, scope MailboxMessageMutationScope) (*ent.MailAccount, time.Time, error) {
	if ctx == nil || transaction == nil || !mailboxMessageMutationScopeValid(scope) {
		return nil, time.Time{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorInvalid)
	}
	account, err := transaction.MailAccount.Query().
		Where(
			mailaccount.IDEQ(scope.AccountID),
			mailaccount.UserIDEQ(scope.ActorID),
			mailaccount.OrganizationIDEQ(scope.OrganizationID),
		).
		ForUpdate().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, time.Time{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorNotFound)
		}
		return nil, time.Time{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
	}
	if account == nil || account.ID != scope.AccountID || account.UserID != scope.ActorID || account.OrganizationID != scope.OrganizationID {
		return nil, time.Time{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorNotFound)
	}
	if account.Status == mailaccount.StatusDeleting {
		return nil, time.Time{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorAccountDeleting)
	}
	if account.Status != mailaccount.StatusActive || !account.DesiredEnabled {
		return nil, time.Time{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorStateConflict)
	}
	if account.EngineAccountID == nil || *account.EngineAccountID != scope.EngineAccountID || !syncValidOpaque(*account.EngineAccountID, MaxEngineOpaqueIDBytes) || MailboxProjectionState(account.ProjectionState) != MailboxProjectionReady || account.LastSyncedAt == nil || !mailboxEntValidTime(*account.LastSyncedAt) || account.SyncEpoch < 0 || account.SyncLeaseEpoch < 1 || account.MutationEpoch < 0 || !mailboxEntValidTime(account.SyncScheduleUpdatedAt) {
		return nil, time.Time{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
	}
	if (account.MutationLeaseToken == nil) != (account.MutationLeaseUntil == nil) || account.MutationLeaseToken != nil && *account.MutationLeaseToken == uuid.Nil || account.MutationLeaseUntil != nil && !mailboxEntValidTime(*account.MutationLeaseUntil) {
		return nil, time.Time{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
	}
	if _, _, valid := mailboxEntSyncLeaseFence(account); !valid {
		return nil, time.Time{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
	}
	databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
	if err != nil {
		return nil, time.Time{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
	}
	return account, databaseNow, nil
}

func mailboxEntMessageMutationResolveLocked(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, input MailboxMessageMutationInput) (MailboxMessageMutationResolution, error) {
	if ctx == nil || transaction == nil || account == nil || account.ID == uuid.Nil || !mailboxMessageMutationInputValid(input) {
		return MailboxMessageMutationResolution{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorInvalid)
	}
	snapshot := MailboxMessageMutationSnapshot{Targets: make([]MailboxMessageMutationResolvedTarget, len(input.TargetIDs))}
	if input.Action == MailboxMessageMutationActionMove {
		folder, err := transaction.MailFolder.Query().
			Where(
				mailfolder.IDEQ(*input.FolderID),
				mailfolder.MailAccountIDEQ(account.ID),
				mailfolder.ControlStateEQ(mailfolder.ControlState("active")),
				mailfolder.DeletedAtIsNil(),
			).
			ForUpdate().
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return MailboxMessageMutationResolution{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorNotFound)
			}
			return MailboxMessageMutationResolution{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
		}
		if folder == nil || folder.ID != *input.FolderID || folder.MailAccountID != account.ID || folder.EngineFolderID == nil || !syncValidOpaque(*folder.EngineFolderID, MaxEngineOpaqueIDBytes) || folder.ControlReceiptID != nil || folder.DeletedAt != nil {
			return MailboxMessageMutationResolution{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorNotFound)
		}
		snapshot.EngineFolderID = *folder.EngineFolderID
	}

	var rows []*ent.MailMessage
	switch input.TargetType {
	case MailboxMessageMutationTargetMessage:
		preliminary, err := transaction.MailMessage.Query().
			Where(mailmessage.IDIn(input.TargetIDs...), mailmessage.MailAccountIDEQ(account.ID), mailmessage.DeletedAtIsNil()).
			Order(ent.Asc(mailmessage.FieldID)).
			Limit(MailboxMessageMutationMaximumResolvedMessages + 1).
			All(ctx)
		if err != nil || len(preliminary) > MailboxMessageMutationMaximumResolvedMessages {
			return MailboxMessageMutationResolution{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
		}
		threadIDs := mailboxEntMessageMutationThreadIDs(preliminary)
		if len(threadIDs) > 0 {
			threads, err := transaction.MailThread.Query().
				Where(mailthread.IDIn(threadIDs...), mailthread.MailAccountIDEQ(account.ID), mailthread.DeletedAtIsNil()).
				Order(ent.Asc(mailthread.FieldID)).
				ForUpdate().
				All(ctx)
			if err != nil || !mailboxEntMessageMutationThreadsExact(threads, threadIDs, account.ID) {
				return MailboxMessageMutationResolution{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
			}
		}
		rows, err = transaction.MailMessage.Query().
			Where(mailmessage.IDIn(input.TargetIDs...), mailmessage.MailAccountIDEQ(account.ID), mailmessage.DeletedAtIsNil()).
			Order(ent.Asc(mailmessage.FieldID)).
			Limit(MailboxMessageMutationMaximumResolvedMessages + 1).
			ForUpdate().
			All(ctx)
		if err != nil || len(rows) > MailboxMessageMutationMaximumResolvedMessages {
			return MailboxMessageMutationResolution{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
		}
		if !mailboxEntMessageMutationRowsUseThreads(rows, threadIDs) {
			return MailboxMessageMutationResolution{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
		}
	case MailboxMessageMutationTargetThread:
		threads, err := transaction.MailThread.Query().
			Where(mailthread.IDIn(input.TargetIDs...), mailthread.MailAccountIDEQ(account.ID), mailthread.DeletedAtIsNil()).
			Order(ent.Asc(mailthread.FieldID)).
			ForUpdate().
			All(ctx)
		if err != nil {
			return MailboxMessageMutationResolution{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
		}
		threadIDs := mailboxEntMessageMutationThreadRowIDs(threads, account.ID)
		if threadIDs == nil && len(threads) > 0 {
			return MailboxMessageMutationResolution{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
		}
		if len(threadIDs) > 0 {
			rows, err = transaction.MailMessage.Query().
				Where(mailmessage.MailThreadIDIn(threadIDs...), mailmessage.MailAccountIDEQ(account.ID), mailmessage.DeletedAtIsNil()).
				Order(ent.Asc(mailmessage.FieldID)).
				Limit(MailboxMessageMutationMaximumResolvedMessages + 1).
				ForUpdate().
				All(ctx)
			if err != nil {
				return MailboxMessageMutationResolution{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
			}
		}
		if len(rows) > MailboxMessageMutationMaximumResolvedMessages {
			return MailboxMessageMutationResolution{Kind: MailboxMessageMutationResolutionExpansionLimit}, nil
		}
	default:
		return MailboxMessageMutationResolution{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorInvalid)
	}

	sort.Slice(rows, func(left, right int) bool { return mailboxMessageMutationUUIDLess(rows[left].ID, rows[right].ID) })
	messageByID := make(map[uuid.UUID]MailboxMessageMutationResolvedMessage, len(rows))
	engineIDs := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		if row == nil || row.ID == uuid.Nil || row.MailAccountID != account.ID || row.MailThreadID == uuid.Nil || row.DeletedAt != nil || !syncValidOpaque(row.EngineMessageID, MaxEngineOpaqueIDBytes) {
			return MailboxMessageMutationResolution{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
		}
		if _, exists := messageByID[row.ID]; exists {
			return MailboxMessageMutationResolution{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
		}
		if _, exists := engineIDs[row.EngineMessageID]; exists {
			return MailboxMessageMutationResolution{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
		}
		resolved := MailboxMessageMutationResolvedMessage{MessageID: row.ID, ThreadID: row.MailThreadID, EngineMessageID: row.EngineMessageID}
		messageByID[row.ID] = resolved
		engineIDs[row.EngineMessageID] = struct{}{}
		snapshot.Messages = append(snapshot.Messages, resolved)
	}
	for index, targetID := range input.TargetIDs {
		target := MailboxMessageMutationResolvedTarget{RequestIndex: index, TargetType: input.TargetType, TargetID: targetID}
		if input.TargetType == MailboxMessageMutationTargetMessage {
			if _, exists := messageByID[targetID]; exists {
				target.MessageIDs = []uuid.UUID{targetID}
			}
		} else {
			for _, message := range snapshot.Messages {
				if message.ThreadID == targetID {
					target.MessageIDs = append(target.MessageIDs, message.MessageID)
				}
			}
		}
		snapshot.Targets[index] = target
	}
	if !mailboxMessageMutationSnapshotValid(snapshot, input) {
		return MailboxMessageMutationResolution{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
	}
	copy := mailboxMessageMutationSnapshotCopy(snapshot)
	return MailboxMessageMutationResolution{Kind: MailboxMessageMutationResolutionResolved, Snapshot: &copy}, nil
}

func mailboxEntMessageMutationThreadIDs(rows []*ent.MailMessage) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(rows))
	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		if row == nil || row.MailThreadID == uuid.Nil {
			continue
		}
		if _, exists := seen[row.MailThreadID]; exists {
			continue
		}
		seen[row.MailThreadID] = struct{}{}
		ids = append(ids, row.MailThreadID)
	}
	sort.Slice(ids, func(left, right int) bool { return mailboxMessageMutationUUIDLess(ids[left], ids[right]) })
	return ids
}

func mailboxEntMessageMutationThreadRowIDs(rows []*ent.MailThread, accountID uuid.UUID) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		if row == nil || row.ID == uuid.Nil || row.MailAccountID != accountID || row.DeletedAt != nil {
			return nil
		}
		ids = append(ids, row.ID)
	}
	sort.Slice(ids, func(left, right int) bool { return mailboxMessageMutationUUIDLess(ids[left], ids[right]) })
	if !mailboxMessageMutationUUIDsStrictlySorted(ids) {
		return nil
	}
	return ids
}

func mailboxEntMessageMutationThreadsExact(rows []*ent.MailThread, expected []uuid.UUID, accountID uuid.UUID) bool {
	ids := mailboxEntMessageMutationThreadRowIDs(rows, accountID)
	return len(ids) == len(expected) && mailboxEntMessageMutationUUIDSlicesEqual(ids, expected)
}

func mailboxEntMessageMutationRowsUseThreads(rows []*ent.MailMessage, threadIDs []uuid.UUID) bool {
	allowed := make(map[uuid.UUID]struct{}, len(threadIDs))
	for _, id := range threadIDs {
		allowed[id] = struct{}{}
	}
	for _, row := range rows {
		if row == nil {
			return false
		}
		if _, exists := allowed[row.MailThreadID]; !exists {
			return false
		}
	}
	return true
}

func mailboxEntMessageMutationUUIDSlicesEqual(left, right []uuid.UUID) bool {
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

func mailboxEntMessageMutationResolutionEqual(left, right MailboxMessageMutationResolution, input MailboxMessageMutationInput) bool {
	if !mailboxMessageMutationResolutionValid(left, input) || !mailboxMessageMutationResolutionValid(right, input) || left.Kind != right.Kind {
		return false
	}
	if left.Kind == MailboxMessageMutationResolutionExpansionLimit {
		return true
	}
	return left.Snapshot != nil && right.Snapshot != nil && mailboxMessageMutationSnapshotEqual(*left.Snapshot, *right.Snapshot, input)
}

func mailboxEntMessageMutationLockReceipt(ctx context.Context, transaction *ent.Tx, scope MailboxMessageMutationScope, idempotencyKeyHash [sha256.Size]byte) (*ent.MailMutationReceipt, error) {
	if ctx == nil || transaction == nil || !mailboxMessageMutationScopeValid(scope) || !mailboxEntMessageMutationDigestValid(idempotencyKeyHash) {
		return nil, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorInvalid)
	}
	rows, err := transaction.MailMutationReceipt.Query().
		Where(
			mailmutationreceipt.OrganizationIDEQ(scope.OrganizationID),
			mailmutationreceipt.ActorUserIDEQ(scope.ActorID),
			mailmutationreceipt.OperationEQ(mailmutationreceipt.Operation(mailboxEntMessageMutationOperation)),
			mailmutationreceipt.IdempotencyKeyHashEQ(idempotencyKeyHash[:]),
		).
		ForUpdate().
		All(ctx)
	if err != nil || len(rows) > 1 {
		return nil, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	if rows[0] == nil {
		return nil, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
	}
	return rows[0], nil
}

func mailboxEntMessageMutationFreshBusyUntil(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, databaseNow time.Time) (*time.Time, error) {
	if ctx == nil || transaction == nil || account == nil || account.ID == uuid.Nil || !mailboxEntValidTime(databaseNow) {
		return nil, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorInvalid)
	}
	var busyUntil *time.Time
	if account.MutationLeaseUntil != nil && account.MutationLeaseUntil.After(databaseNow) {
		busyUntil = mailboxEntEngineCommandTimeCopy(*account.MutationLeaseUntil)
	}
	_, hasSyncLease, syncLeaseValid := mailboxEntSyncLeaseFence(account)
	if !syncLeaseValid {
		return nil, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
	}
	if hasSyncLease && account.SyncLeaseUntil.After(databaseNow) && (busyUntil == nil || account.SyncLeaseUntil.Before(*busyUntil)) {
		busyUntil = mailboxEntEngineCommandTimeCopy(*account.SyncLeaseUntil)
	}
	if busyUntil != nil {
		return busyUntil, nil
	}
	manualOnly, err := transaction.MailMutationReceipt.Query().
		Where(
			mailmutationreceipt.MailAccountIDEQ(account.ID),
			mailmutationreceipt.OperationEQ(mailmutationreceipt.Operation(mailboxEntMessageMutationOperation)),
			mailmutationreceipt.StateEQ(mailmutationreceipt.State("unknown")),
			mailmutationreceipt.SafeCodeEQ(mailmutationreceipt.SafeCode(MailboxMessageMutationCodeOutcomeUnknown)),
			mailmutationreceipt.ReconcileAttemptsEQ(10),
			mailmutationreceipt.NextReconcileAtIsNil(),
			mailmutationreceipt.ReconcileLeaseTokenIsNil(),
			mailmutationreceipt.ReconcileLeaseUntilIsNil(),
			mailmutationreceipt.TerminalAtIsNil(),
			mailmutationreceipt.ReplayUntilIsNil(),
			mailmutationreceipt.CompactedAtIsNil(),
		).
		Order(ent.Asc(mailmutationreceipt.FieldID)).
		Limit(1).
		ForUpdate().
		All(ctx)
	if err != nil {
		return nil, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
	}
	if len(manualOnly) == 0 {
		return nil, nil
	}
	bounded, ok := mailboxEntCheckedTimeAdd(databaseNow, mailboxEntMessageMutationBoundedBusyDuration)
	if !ok {
		return nil, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
	}
	return mailboxEntEngineCommandTimeCopy(bounded), nil
}

func mailboxEntMessageMutationReplayFromReceipt(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, databaseNow time.Time, request MailboxMessageMutationReplayRequest, receipt *ent.MailMutationReceipt) (MailboxMessageMutationReplay, error) {
	if !mailboxEntMessageMutationReceiptIdentityValid(receipt, request.Scope, request.IdempotencyKeyHash) {
		return MailboxMessageMutationReplay{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorIdempotencyConflict)
	}
	if !bytes.Equal(receipt.RequestHash, request.RequestHash[:]) {
		return MailboxMessageMutationReplay{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorIdempotencyConflict)
	}
	if receipt.FingerprintVersion != request.FingerprintVersion || !mailboxEntMessageMutationReceiptActionValid(receipt) || !mailboxEntMessageMutationReceiptLifecycleValid(receipt, databaseNow) {
		return MailboxMessageMutationReplay{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
	}
	graph, err := mailboxEntMessageMutationLoadGraph(ctx, transaction, receipt)
	if err != nil || !mailboxEntMessageMutationGraphMatchesReceiptState(graph, receipt.State) {
		return MailboxMessageMutationReplay{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
	}
	switch receipt.State {
	case mailmutationreceipt.State("in_progress"):
		busyUntil := mailboxEntMessageMutationInProgressBusyUntil(account, receipt, databaseNow)
		return MailboxMessageMutationReplay{Kind: MailboxMessageMutationReplayInProgress, BusyUntil: busyUntil}, nil
	case mailmutationreceipt.State("unknown"):
		result := mailboxEntMessageMutationResultFromGraph(graph, MailboxMessageMutationResultUnknown)
		return MailboxMessageMutationReplay{Kind: MailboxMessageMutationReplayUnknown, Result: &result}, nil
	case mailmutationreceipt.State("completed"):
		result := mailboxEntMessageMutationResultFromGraph(graph, MailboxMessageMutationResultCompleted)
		return MailboxMessageMutationReplay{Kind: MailboxMessageMutationReplayCompleted, Result: &result}, nil
	case mailmutationreceipt.State("failed"):
		return MailboxMessageMutationReplay{Kind: MailboxMessageMutationReplayExpansionLimit}, nil
	default:
		return MailboxMessageMutationReplay{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
	}
}

func mailboxEntMessageMutationReservationFromReceipt(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, databaseNow time.Time, request MailboxMessageMutationReserveRequest, receipt *ent.MailMutationReceipt) (MailboxMessageMutationReservation, error) {
	if !mailboxEntMessageMutationReceiptIdentityValid(receipt, request.Scope, request.IdempotencyKeyHash) || !bytes.Equal(receipt.RequestHash, request.RequestHash[:]) || receipt.FingerprintVersion != request.FingerprintVersion {
		return MailboxMessageMutationReservation{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorIdempotencyConflict)
	}
	if !mailboxEntMessageMutationReceiptMatchesInput(receipt, request.Input) || !mailboxEntMessageMutationReceiptLifecycleValid(receipt, databaseNow) {
		return MailboxMessageMutationReservation{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
	}
	graph, err := mailboxEntMessageMutationLoadGraph(ctx, transaction, receipt)
	if err != nil || !mailboxEntMessageMutationGraphMatchesInput(graph, request.Input) || !mailboxEntMessageMutationGraphMatchesReceiptState(graph, receipt.State) {
		return MailboxMessageMutationReservation{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
	}
	switch receipt.State {
	case mailmutationreceipt.State("in_progress"):
		return MailboxMessageMutationReservation{Kind: MailboxMessageMutationReservationBusy, BusyUntil: mailboxEntMessageMutationInProgressBusyUntil(account, receipt, databaseNow)}, nil
	case mailmutationreceipt.State("unknown"):
		result := mailboxEntMessageMutationResultFromGraph(graph, MailboxMessageMutationResultUnknown)
		return MailboxMessageMutationReservation{Kind: MailboxMessageMutationReservationUnknown, Result: &result}, nil
	case mailmutationreceipt.State("completed"):
		result := mailboxEntMessageMutationResultFromGraph(graph, MailboxMessageMutationResultCompleted)
		return MailboxMessageMutationReservation{Kind: MailboxMessageMutationReservationCompleted, Result: &result}, nil
	case mailmutationreceipt.State("failed"):
		return MailboxMessageMutationReservation{Kind: MailboxMessageMutationReservationExpansionLimit}, nil
	default:
		return MailboxMessageMutationReservation{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
	}
}

func mailboxEntMessageMutationInProgressBusyUntil(account *ent.MailAccount, receipt *ent.MailMutationReceipt, databaseNow time.Time) *time.Time {
	if account != nil && account.MutationLeaseUntil != nil && account.MutationLeaseUntil.After(databaseNow) {
		return mailboxEntEngineCommandTimeCopy(*account.MutationLeaseUntil)
	}
	if receipt != nil && receipt.ReconcileLeaseUntil != nil && receipt.ReconcileLeaseUntil.After(databaseNow) {
		return mailboxEntEngineCommandTimeCopy(*receipt.ReconcileLeaseUntil)
	}
	bounded, ok := mailboxEntCheckedTimeAdd(databaseNow, mailboxEntMessageMutationBoundedBusyDuration)
	if !ok {
		bounded = databaseNow
	}
	return mailboxEntEngineCommandTimeCopy(bounded)
}

func mailboxEntMessageMutationReceiptIdentityValid(receipt *ent.MailMutationReceipt, scope MailboxMessageMutationScope, idempotencyKeyHash [sha256.Size]byte) bool {
	return receipt != nil && receipt.ID != uuid.Nil && receipt.OrganizationID == scope.OrganizationID && receipt.ActorUserID == scope.ActorID && receipt.MailAccountID != nil && *receipt.MailAccountID == scope.AccountID && receipt.Operation == mailmutationreceipt.Operation(mailboxEntMessageMutationOperation) && receipt.KeySource == mailmutationreceipt.KeySource("client") && bytes.Equal(receipt.IdempotencyKeyHash, idempotencyKeyHash[:]) && mailboxEntMessageMutationDigestBytesValid(receipt.RequestHash) && receipt.TargetKind == nil && receipt.TargetID == nil
}

func mailboxEntMessageMutationReceiptActionValid(receipt *ent.MailMutationReceipt) bool {
	if receipt == nil || receipt.BulkAction == nil {
		return false
	}
	action := MailboxMessageMutationAction(*receipt.BulkAction)
	if !validMailboxMessageMutationAction(action) {
		return false
	}
	if action == MailboxMessageMutationActionMove {
		return receipt.ArgumentKind != nil && *receipt.ArgumentKind == mailmutationreceipt.ArgumentKind("folder") && receipt.ArgumentID != nil && *receipt.ArgumentID != uuid.Nil
	}
	return receipt.ArgumentKind == nil && receipt.ArgumentID == nil
}

func mailboxEntMessageMutationReceiptMatchesInput(receipt *ent.MailMutationReceipt, input MailboxMessageMutationInput) bool {
	if !mailboxEntMessageMutationReceiptActionValid(receipt) || !mailboxMessageMutationInputValid(input) || MailboxMessageMutationAction(*receipt.BulkAction) != input.Action {
		return false
	}
	if input.Action == MailboxMessageMutationActionMove {
		return input.FolderID != nil && receipt.ArgumentID != nil && *receipt.ArgumentID == *input.FolderID
	}
	return true
}

func mailboxEntMessageMutationReceiptLifecycleValid(receipt *ent.MailMutationReceipt, databaseNow time.Time) bool {
	if receipt == nil || !mailboxEntValidTime(databaseNow) || receipt.ID == uuid.Nil || !mailboxEntValidTime(receipt.CreatedAt) || !mailboxEntValidTime(receipt.UpdatedAt) || receipt.UpdatedAt.Before(receipt.CreatedAt) || receipt.CompactedAt != nil {
		return false
	}
	leasePaired := (receipt.ReconcileLeaseToken == nil) == (receipt.ReconcileLeaseUntil == nil)
	if !leasePaired {
		return false
	}
	leaseValid := receipt.ReconcileLeaseToken == nil || *receipt.ReconcileLeaseToken != uuid.Nil && mailboxEntValidTime(*receipt.ReconcileLeaseUntil)
	if !leaseValid {
		return false
	}
	switch receipt.State {
	case mailmutationreceipt.State("in_progress"):
		return receipt.SafeCode == nil && receipt.OriginalHTTPStatus == nil && receipt.ReconcileAttempts == 0 && receipt.NextReconcileAt == nil && receipt.ReconcileLeaseToken != nil && receipt.TerminalAt == nil && receipt.ReplayUntil == nil
	case mailmutationreceipt.State("unknown"):
		if receipt.SafeCode == nil || *receipt.SafeCode != mailmutationreceipt.SafeCode(MailboxMessageMutationCodeOutcomeUnknown) || receipt.OriginalHTTPStatus != nil || receipt.ReconcileAttempts < 1 || receipt.ReconcileAttempts > 10 || receipt.TerminalAt != nil || receipt.ReplayUntil != nil {
			return false
		}
		if receipt.ReconcileAttempts == 10 {
			return receipt.NextReconcileAt == nil && receipt.ReconcileLeaseToken == nil
		}
		return receipt.NextReconcileAt != nil && mailboxEntValidTime(*receipt.NextReconcileAt)
	case mailmutationreceipt.State("completed"):
		return receipt.SafeCode == nil && receipt.OriginalHTTPStatus != nil && *receipt.OriginalHTTPStatus == 200 && receipt.ReconcileAttempts == 0 && receipt.NextReconcileAt == nil && receipt.ReconcileLeaseToken == nil && mailboxEntMessageMutationTerminalReplayValid(receipt, databaseNow)
	case mailmutationreceipt.State("failed"):
		return receipt.SafeCode != nil && *receipt.SafeCode == mailmutationreceipt.SafeCode("MAIL_BULK_EXPANSION_LIMIT") && receipt.OriginalHTTPStatus != nil && *receipt.OriginalHTTPStatus == 422 && receipt.ReconcileAttempts == 0 && receipt.NextReconcileAt == nil && receipt.ReconcileLeaseToken == nil && mailboxEntMessageMutationTerminalReplayValid(receipt, databaseNow)
	default:
		return false
	}
}

func mailboxEntMessageMutationTerminalReplayValid(receipt *ent.MailMutationReceipt, databaseNow time.Time) bool {
	return receipt != nil && receipt.TerminalAt != nil && mailboxEntValidTime(*receipt.TerminalAt) && receipt.ReplayUntil != nil && mailboxEntValidTime(*receipt.ReplayUntil) && receipt.ReplayUntil.Equal(receipt.TerminalAt.Add(mailboxEntMessageMutationReplayLifetime)) && databaseNow.Before(*receipt.ReplayUntil)
}

func mailboxEntMessageMutationDigestBytesValid(value []byte) bool {
	if len(value) != sha256.Size {
		return false
	}
	for _, item := range value {
		if item != 0 {
			return true
		}
	}
	return false
}

func mailboxEntMessageMutationCreateTerminalReceipt(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, databaseNow time.Time, request MailboxMessageMutationReserveRequest, state mailmutationreceipt.State) (*MailboxMessageMutationResult, error) {
	receipt, err := mailboxEntMessageMutationCreateReceipt(ctx, transaction, account, databaseNow, request, state, nil)
	if err != nil {
		return nil, err
	}
	if state == mailmutationreceipt.State("failed") {
		if err := mailboxEntMessageMutationCreateUnavailableTargets(ctx, transaction, receipt, request.Input, databaseNow); err != nil {
			return nil, err
		}
		graph, err := mailboxEntMessageMutationLoadGraph(ctx, transaction, receipt)
		if err != nil || !mailboxEntMessageMutationGraphMatchesInput(graph, request.Input) || !mailboxEntMessageMutationGraphMatchesReceiptState(graph, state) {
			return nil, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
		}
		return nil, nil
	}
	if state != mailmutationreceipt.State("completed") || request.Resolution.Snapshot == nil || len(request.Resolution.Snapshot.Messages) != 0 {
		return nil, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorInvalid)
	}
	if err := mailboxEntMessageMutationCreateSnapshotRows(ctx, transaction, receipt, *request.Resolution.Snapshot, databaseNow, false); err != nil {
		return nil, err
	}
	graph, err := mailboxEntMessageMutationLoadGraph(ctx, transaction, receipt)
	if err != nil || !mailboxEntMessageMutationGraphMatchesInput(graph, request.Input) || !mailboxEntMessageMutationGraphMatchesReceiptState(graph, state) {
		return nil, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
	}
	result := mailboxEntMessageMutationResultFromGraph(graph, MailboxMessageMutationResultCompleted)
	return &result, nil
}

func mailboxEntMessageMutationCreateReceipt(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, databaseNow time.Time, request MailboxMessageMutationReserveRequest, state mailmutationreceipt.State, fence *MailboxEngineCommandFence) (*ent.MailMutationReceipt, error) {
	if ctx == nil || transaction == nil || account == nil || account.ID != request.Scope.AccountID || !mailboxEntValidTime(databaseNow) || !mailboxEntMessageMutationReserveRequestValid(request) {
		return nil, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorInvalid)
	}
	builder := transaction.MailMutationReceipt.Create().
		SetOrganizationID(request.Scope.OrganizationID).
		SetActorUserID(request.Scope.ActorID).
		SetMailAccountID(request.Scope.AccountID).
		SetOperation(mailmutationreceipt.Operation(mailboxEntMessageMutationOperation)).
		SetKeySource(mailmutationreceipt.KeySource("client")).
		SetState(state).
		SetIdempotencyKeyHash(append([]byte(nil), request.IdempotencyKeyHash[:]...)).
		SetRequestHash(append([]byte(nil), request.RequestHash[:]...)).
		SetFingerprintVersion(request.FingerprintVersion).
		SetBulkAction(mailmutationreceipt.BulkAction(request.Input.Action)).
		SetCreatedAt(databaseNow).
		SetUpdatedAt(databaseNow)
	if request.Input.Action == MailboxMessageMutationActionMove {
		builder.SetArgumentKind(mailmutationreceipt.ArgumentKind("folder")).SetArgumentID(*request.Input.FolderID)
	}
	switch state {
	case mailmutationreceipt.State("in_progress"):
		if fence == nil || !mailboxEntEngineCommandFenceValid(*fence) || fence.AccountID != account.ID {
			return nil, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorInvalid)
		}
		builder.SetReconcileLeaseToken(fence.LeaseID).SetReconcileLeaseUntil(fence.LeaseUntil)
	case mailmutationreceipt.State("completed"), mailmutationreceipt.State("failed"):
		if fence != nil {
			return nil, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorInvalid)
		}
		replayUntil, ok := mailboxEntCheckedTimeAdd(databaseNow, mailboxEntMessageMutationReplayLifetime)
		if !ok {
			return nil, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
		}
		builder.SetTerminalAt(databaseNow).SetReplayUntil(replayUntil)
		if state == mailmutationreceipt.State("completed") {
			builder.SetOriginalHTTPStatus(200)
		} else {
			builder.SetOriginalHTTPStatus(422).SetSafeCode(mailmutationreceipt.SafeCode("MAIL_BULK_EXPANSION_LIMIT"))
		}
	default:
		return nil, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorInvalid)
	}
	receipt, err := builder.Save(ctx)
	if err != nil || receipt == nil || receipt.ID == uuid.Nil {
		return nil, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
	}
	return receipt, nil
}

func mailboxEntMessageMutationCreateUnavailableTargets(ctx context.Context, transaction *ent.Tx, receipt *ent.MailMutationReceipt, input MailboxMessageMutationInput, databaseNow time.Time) error {
	if ctx == nil || transaction == nil || receipt == nil || receipt.ID == uuid.Nil || !mailboxMessageMutationInputValid(input) || !mailboxEntValidTime(databaseNow) {
		return mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorInvalid)
	}
	for index, targetID := range input.TargetIDs {
		if _, err := transaction.MailMutationReceiptTarget.Create().
			SetReceiptID(receipt.ID).
			SetRequestIndex(int32(index)).
			SetTargetType(mailmutationreceipttarget.TargetType(input.TargetType)).
			SetTargetID(targetID).
			SetStatus(mailmutationreceipttarget.Status("failed")).
			SetAffectedCount(0).
			SetSucceededCount(0).
			SetFailedCount(0).
			SetTrainingStatus(mailmutationreceipttarget.TrainingStatus("not_requested")).
			SetSafeCode(mailmutationreceipttarget.SafeCode(MailboxMessageMutationCodeTargetUnavailable)).
			SetCreatedAt(databaseNow).
			SetUpdatedAt(databaseNow).
			Save(ctx); err != nil {
			return mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
		}
	}
	return nil
}

func mailboxEntMessageMutationCreateSnapshotRows(ctx context.Context, transaction *ent.Tx, receipt *ent.MailMutationReceipt, snapshot MailboxMessageMutationSnapshot, databaseNow time.Time, provisionalUnknown bool) error {
	if ctx == nil || transaction == nil || receipt == nil || receipt.ID == uuid.Nil || !mailboxEntValidTime(databaseNow) || len(snapshot.Targets) < 1 || len(snapshot.Targets) > MailboxMessageMutationMaximumResolvedMessages || len(snapshot.Messages) > MailboxMessageMutationMaximumResolvedMessages {
		return mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorInvalid)
	}
	for _, target := range snapshot.Targets {
		memberCount := int32(len(target.MessageIDs))
		builder := transaction.MailMutationReceiptTarget.Create().
			SetReceiptID(receipt.ID).
			SetRequestIndex(int32(target.RequestIndex)).
			SetTargetType(mailmutationreceipttarget.TargetType(target.TargetType)).
			SetTargetID(target.TargetID).
			SetAffectedCount(memberCount).
			SetSucceededCount(0).
			SetTrainingStatus(mailmutationreceipttarget.TrainingStatus("not_requested")).
			SetCreatedAt(databaseNow).
			SetUpdatedAt(databaseNow)
		if memberCount == 0 {
			builder.SetStatus(mailmutationreceipttarget.Status("failed")).SetFailedCount(0).SetSafeCode(mailmutationreceipttarget.SafeCode(MailboxMessageMutationCodeTargetUnavailable))
		} else if provisionalUnknown {
			builder.SetStatus(mailmutationreceipttarget.Status("unknown")).SetFailedCount(0).SetSafeCode(mailmutationreceipttarget.SafeCode(MailboxMessageMutationCodeOutcomeUnknown))
		} else {
			return mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorInvalid)
		}
		targetRow, err := builder.Save(ctx)
		if err != nil || targetRow == nil || targetRow.ID == uuid.Nil {
			return mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
		}
		for _, messageID := range target.MessageIDs {
			if _, err := transaction.MailMutationReceiptMember.Create().
				SetReceiptID(receipt.ID).
				SetTargetID(targetRow.ID).
				SetMessageID(messageID).
				SetCreatedAt(databaseNow).
				SetUpdatedAt(databaseNow).
				Save(ctx); err != nil {
				return mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
			}
		}
	}
	return nil
}

func mailboxEntMessageMutationLoadGraph(ctx context.Context, transaction *ent.Tx, receipt *ent.MailMutationReceipt) (mailboxEntMessageMutationGraph, error) {
	if ctx == nil || transaction == nil || receipt == nil || receipt.ID == uuid.Nil {
		return mailboxEntMessageMutationGraph{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorInvalid)
	}
	targets, err := transaction.MailMutationReceiptTarget.Query().
		Where(mailmutationreceipttarget.ReceiptIDEQ(receipt.ID)).
		Order(ent.Asc(mailmutationreceipttarget.FieldRequestIndex)).
		ForUpdate().
		All(ctx)
	if err != nil || len(targets) < 1 || len(targets) > MailboxMessageMutationMaximumResolvedMessages {
		return mailboxEntMessageMutationGraph{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
	}
	members, err := transaction.MailMutationReceiptMember.Query().
		Where(mailmutationreceiptmember.ReceiptIDEQ(receipt.ID)).
		Order(ent.Asc(mailmutationreceiptmember.FieldMessageID)).
		ForUpdate().
		All(ctx)
	if err != nil || len(members) > MailboxMessageMutationMaximumResolvedMessages {
		return mailboxEntMessageMutationGraph{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
	}
	graph := mailboxEntMessageMutationGraph{
		targets: targets, members: members,
		membersByTarget: make(map[uuid.UUID][]uuid.UUID, len(targets)),
	}
	targetByID := make(map[uuid.UUID]*ent.MailMutationReceiptTarget, len(targets))
	targetIdentity := make(map[string]struct{}, len(targets))
	var commonType MailboxMessageMutationTargetType
	for index, target := range targets {
		if target == nil || target.ID == uuid.Nil || target.ReceiptID != receipt.ID || target.RequestIndex != int32(index) || target.TargetID == uuid.Nil || !mailboxEntValidTime(target.CreatedAt) || !mailboxEntValidTime(target.UpdatedAt) || target.UpdatedAt.Before(target.CreatedAt) {
			return mailboxEntMessageMutationGraph{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
		}
		targetType := MailboxMessageMutationTargetType(target.TargetType)
		if !validMailboxMessageMutationTargetType(targetType) || index > 0 && targetType != commonType {
			return mailboxEntMessageMutationGraph{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
		}
		commonType = targetType
		identity := string(targetType) + ":" + target.TargetID.String()
		if _, exists := targetIdentity[identity]; exists {
			return mailboxEntMessageMutationGraph{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
		}
		targetIdentity[identity] = struct{}{}
		targetByID[target.ID] = target
	}
	seenMessages := make(map[uuid.UUID]struct{}, len(members))
	for index, member := range members {
		if member == nil || member.ID == uuid.Nil || member.ReceiptID != receipt.ID || member.TargetID == uuid.Nil || member.MessageID == uuid.Nil || !mailboxEntValidTime(member.CreatedAt) || !mailboxEntValidTime(member.UpdatedAt) || member.UpdatedAt.Before(member.CreatedAt) || index > 0 && !mailboxMessageMutationUUIDLess(members[index-1].MessageID, member.MessageID) {
			return mailboxEntMessageMutationGraph{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
		}
		if _, exists := targetByID[member.TargetID]; !exists {
			return mailboxEntMessageMutationGraph{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
		}
		if _, exists := seenMessages[member.MessageID]; exists {
			return mailboxEntMessageMutationGraph{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
		}
		seenMessages[member.MessageID] = struct{}{}
		graph.membersByTarget[member.TargetID] = append(graph.membersByTarget[member.TargetID], member.MessageID)
	}
	results := make([]MailboxMessageMutationTargetResult, len(targets))
	unknown := false
	for index, target := range targets {
		memberIDs := graph.membersByTarget[target.ID]
		if !mailboxMessageMutationUUIDsStrictlySorted(memberIDs) || target.TargetType == mailmutationreceipttarget.TargetType("message") && len(memberIDs) > 0 && (len(memberIDs) != 1 || memberIDs[0] != target.TargetID) {
			return mailboxEntMessageMutationGraph{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
		}
		code := (*MailboxMessageMutationSafeCode)(nil)
		if target.SafeCode != nil {
			value := MailboxMessageMutationSafeCode(*target.SafeCode)
			code = &value
		}
		result := MailboxMessageMutationTargetResult{
			TargetType: MailboxMessageMutationTargetType(target.TargetType),
			TargetID: target.TargetID,
			Status: MailboxMessageMutationTargetStatus(target.Status),
			AffectedMessageCount: int(target.AffectedCount),
			SucceededMessageCount: int(target.SucceededCount),
			FailedMessageCount: int(target.FailedCount),
			TrainingStatus: MailboxMessageMutationTrainingStatus(target.TrainingStatus),
			Code: code,
		}
		if !mailboxMessageMutationTargetResultValid(result) || result.TrainingStatus != MailboxMessageMutationTrainingNotRequested || result.AffectedMessageCount != len(memberIDs) {
			return mailboxEntMessageMutationGraph{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
		}
		unknown = unknown || result.Status == MailboxMessageMutationTargetUnknown
		results[index] = result
	}
	state := MailboxMessageMutationResultCompleted
	if unknown {
		state = MailboxMessageMutationResultUnknown
	}
	graph.result = MailboxMessageMutationResult{State: state, Results: results}
	return graph, nil
}

func mailboxEntMessageMutationGraphMatchesInput(graph mailboxEntMessageMutationGraph, input MailboxMessageMutationInput) bool {
	if !mailboxMessageMutationInputValid(input) || len(graph.targets) != len(input.TargetIDs) || len(graph.result.Results) != len(input.TargetIDs) {
		return false
	}
	for index, target := range graph.targets {
		if target == nil || target.RequestIndex != int32(index) || MailboxMessageMutationTargetType(target.TargetType) != input.TargetType || target.TargetID != input.TargetIDs[index] {
			return false
		}
	}
	return true
}

func mailboxEntMessageMutationGraphMatchesReceiptState(graph mailboxEntMessageMutationGraph, state mailmutationreceipt.State) bool {
	if len(graph.targets) < 1 || len(graph.targets) != len(graph.result.Results) || len(graph.members) > MailboxMessageMutationMaximumResolvedMessages {
		return false
	}
	switch state {
	case mailmutationreceipt.State("in_progress"), mailmutationreceipt.State("unknown"):
		if len(graph.members) < 1 || graph.result.State != MailboxMessageMutationResultUnknown {
			return false
		}
		for index, target := range graph.targets {
			if len(graph.membersByTarget[target.ID]) == 0 {
				if graph.result.Results[index].Status != MailboxMessageMutationTargetFailed || graph.result.Results[index].Code == nil || *graph.result.Results[index].Code != MailboxMessageMutationCodeTargetUnavailable {
					return false
				}
				continue
			}
			if graph.result.Results[index].Status != MailboxMessageMutationTargetUnknown || graph.result.Results[index].Code == nil || *graph.result.Results[index].Code != MailboxMessageMutationCodeOutcomeUnknown {
				return false
			}
		}
		return true
	case mailmutationreceipt.State("completed"):
		return graph.result.State == MailboxMessageMutationResultCompleted
	case mailmutationreceipt.State("failed"):
		if len(graph.members) != 0 || graph.result.State != MailboxMessageMutationResultCompleted {
			return false
		}
		for _, target := range graph.result.Results {
			if target.Status != MailboxMessageMutationTargetFailed || target.AffectedMessageCount != 0 || target.Code == nil || *target.Code != MailboxMessageMutationCodeTargetUnavailable {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func mailboxEntMessageMutationLockFinishGraph(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, databaseNow time.Time, receiptID uuid.UUID, fence MailboxEngineCommandFence) (*ent.MailMutationReceipt, mailboxEntMessageMutationGraph, error) {
	if ctx == nil || transaction == nil || account == nil || receiptID == uuid.Nil || !mailboxEntEngineCommandFenceValid(fence) || !mailboxEntValidTime(databaseNow) {
		return nil, mailboxEntMessageMutationGraph{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorInvalid)
	}
	if !mailboxEntDraftCreateFenceMatches(account, fence, databaseNow) {
		return nil, mailboxEntMessageMutationGraph{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorFenceLost)
	}
	receipt, err := transaction.MailMutationReceipt.Query().
		Where(
			mailmutationreceipt.IDEQ(receiptID),
			mailmutationreceipt.MailAccountIDEQ(account.ID),
			mailmutationreceipt.OrganizationIDEQ(account.OrganizationID),
			mailmutationreceipt.ActorUserIDEQ(account.UserID),
			mailmutationreceipt.OperationEQ(mailmutationreceipt.Operation(mailboxEntMessageMutationOperation)),
		).
		ForUpdate().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, mailboxEntMessageMutationGraph{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorNotFound)
		}
		return nil, mailboxEntMessageMutationGraph{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
	}
	if receipt == nil || receipt.ID != receiptID || receipt.MailAccountID == nil || *receipt.MailAccountID != account.ID || receipt.State != mailmutationreceipt.State("in_progress") || !mailboxEntMessageMutationReceiptActionValid(receipt) || !mailboxEntMessageMutationReceiptLifecycleValid(receipt, databaseNow) {
		return nil, mailboxEntMessageMutationGraph{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorStateConflict)
	}
	if receipt.ReconcileLeaseToken == nil || *receipt.ReconcileLeaseToken != fence.LeaseID || receipt.ReconcileLeaseUntil == nil || !receipt.ReconcileLeaseUntil.Equal(fence.LeaseUntil) {
		return nil, mailboxEntMessageMutationGraph{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorFenceLost)
	}
	graph, err := mailboxEntMessageMutationLoadGraph(ctx, transaction, receipt)
	if err != nil || !mailboxEntMessageMutationGraphMatchesReceiptState(graph, receipt.State) {
		return nil, mailboxEntMessageMutationGraph{}, mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
	}
	return receipt, graph, nil
}

func mailboxEntMessageMutationWriteTargetOutcome(ctx context.Context, graph mailboxEntMessageMutationGraph, databaseNow time.Time, outcome MailboxMessageMutationFinishOutcome) error {
	if ctx == nil || !mailboxEntValidTime(databaseNow) || (outcome != MailboxMessageMutationFinishApplied && outcome != MailboxMessageMutationFinishRejected) {
		return mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorInvalid)
	}
	for index, target := range graph.targets {
		memberCount := int32(len(graph.membersByTarget[target.ID]))
		if memberCount == 0 {
			continue
		}
		update := target.Update().
			SetAffectedCount(memberCount).
			SetTrainingStatus(mailmutationreceipttarget.TrainingStatus("not_requested")).
			SetUpdatedAt(databaseNow)
		if outcome == MailboxMessageMutationFinishApplied {
			if err := update.
				SetStatus(mailmutationreceipttarget.Status("succeeded")).
				SetSucceededCount(memberCount).
				SetFailedCount(0).
				ClearSafeCode().
				Exec(ctx); err != nil {
				return mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
			}
			target.Status = mailmutationreceipttarget.Status("succeeded")
			target.SucceededCount = memberCount
			target.FailedCount = 0
			target.SafeCode = nil
			graph.result.Results[index] = MailboxMessageMutationTargetResult{
				TargetType: MailboxMessageMutationTargetType(target.TargetType), TargetID: target.TargetID,
				Status: MailboxMessageMutationTargetSucceeded, AffectedMessageCount: int(memberCount),
				SucceededMessageCount: int(memberCount), TrainingStatus: MailboxMessageMutationTrainingNotRequested,
			}
			continue
		}
		if err := update.
			SetStatus(mailmutationreceipttarget.Status("failed")).
			SetSucceededCount(0).
			SetFailedCount(memberCount).
			SetSafeCode(mailmutationreceipttarget.SafeCode(MailboxMessageMutationCodeActionRejected)).
			Exec(ctx); err != nil {
			return mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
		}
		code := MailboxMessageMutationCodeActionRejected
		target.Status = mailmutationreceipttarget.Status("failed")
		target.SucceededCount = 0
		target.FailedCount = memberCount
		storedCode := mailmutationreceipttarget.SafeCode(code)
		target.SafeCode = &storedCode
		graph.result.Results[index] = MailboxMessageMutationTargetResult{
			TargetType: MailboxMessageMutationTargetType(target.TargetType), TargetID: target.TargetID,
			Status: MailboxMessageMutationTargetFailed, AffectedMessageCount: int(memberCount),
			FailedMessageCount: int(memberCount), TrainingStatus: MailboxMessageMutationTrainingNotRequested, Code: &code,
		}
	}
	graph.result.State = MailboxMessageMutationResultCompleted
	return nil
}

func mailboxEntMessageMutationWriteUnknownTargets(ctx context.Context, graph mailboxEntMessageMutationGraph, databaseNow time.Time) error {
	if ctx == nil || !mailboxEntValidTime(databaseNow) {
		return mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorInvalid)
	}
	for index, target := range graph.targets {
		memberCount := int32(len(graph.membersByTarget[target.ID]))
		if memberCount == 0 {
			continue
		}
		if err := target.Update().
			SetStatus(mailmutationreceipttarget.Status("unknown")).
			SetAffectedCount(memberCount).
			SetSucceededCount(0).
			SetFailedCount(0).
			SetTrainingStatus(mailmutationreceipttarget.TrainingStatus("not_requested")).
			SetSafeCode(mailmutationreceipttarget.SafeCode(MailboxMessageMutationCodeOutcomeUnknown)).
			SetUpdatedAt(databaseNow).
			Exec(ctx); err != nil {
			return mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
		}
		code := MailboxMessageMutationCodeOutcomeUnknown
		target.Status = mailmutationreceipttarget.Status("unknown")
		target.SucceededCount = 0
		target.FailedCount = 0
		storedCode := mailmutationreceipttarget.SafeCode(code)
		target.SafeCode = &storedCode
		graph.result.Results[index] = MailboxMessageMutationTargetResult{
			TargetType: MailboxMessageMutationTargetType(target.TargetType), TargetID: target.TargetID,
			Status: MailboxMessageMutationTargetUnknown, AffectedMessageCount: int(memberCount),
			TrainingStatus: MailboxMessageMutationTrainingNotRequested, Code: &code,
		}
	}
	graph.result.State = MailboxMessageMutationResultUnknown
	return nil
}

func mailboxEntMessageMutationResultFromGraph(graph mailboxEntMessageMutationGraph, state MailboxMessageMutationResultState) MailboxMessageMutationResult {
	result := mailboxMessageMutationResultCopy(graph.result)
	result.State = state
	return result
}

func mailboxEntMessageMutationReleaseAccount(ctx context.Context, account *ent.MailAccount, fence MailboxEngineCommandFence, databaseNow time.Time) error {
	if ctx == nil || account == nil || !mailboxEntDraftCreateFenceMatches(account, fence, databaseNow) {
		return mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorFenceLost)
	}
	nextMutationEpoch, ok := mailboxEntNextEpoch(account.MutationEpoch)
	if !ok {
		return mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
	}
	if err := account.Update().
		ClearMutationLeaseToken().
		ClearMutationLeaseUntil().
		SetMutationEpoch(nextMutationEpoch).
		SetSyncScheduleUpdatedAt(databaseNow).
		Exec(ctx); err != nil {
		return mailboxEntMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
	}
	return nil
}

func mailboxEntMessageMutationRepositoryError(kind MailboxMessageMutationRepositoryErrorKind) error {
	return NewMailboxMessageMutationRepositoryError(kind)
}

func mailboxEntMessageMutationSafeError(err error) error {
	var repositoryError *MailboxMessageMutationRepositoryError
	if errors.As(err, &repositoryError) && repositoryError != nil {
		return NewMailboxMessageMutationRepositoryError(repositoryError.Kind())
	}
	switch {
	case errors.Is(err, ErrMailboxServiceInvalid):
		return NewMailboxMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorInvalid)
	case errors.Is(err, ErrMailboxServiceNotFound):
		return NewMailboxMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorNotFound)
	case errors.Is(err, ErrMailboxServiceDeleting):
		return NewMailboxMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorAccountDeleting)
	default:
		return NewMailboxMessageMutationRepositoryError(MailboxMessageMutationRepositoryErrorUnavailable)
	}
}
