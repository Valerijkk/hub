package mail

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
)

const mailboxEntMaxInt64 = int64(1<<63 - 1)

func (repository *MailboxEntRepository) ClaimMailboxSyncStep(ctx context.Context, request MailboxSyncClaimRequest) (MailboxSyncClaimResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || request.AccountID == uuid.Nil || request.LeaseID == uuid.Nil || !mailboxEntValidTime(request.Now) || !mailboxEntValidTime(request.LeaseUntil) || !request.LeaseUntil.After(request.Now) {
		return MailboxSyncClaimResult{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	var result MailboxSyncClaimResult
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		now, err := repository.mailboxEntNow()
		if err != nil {
			return err
		}
		if !request.LeaseUntil.After(now) {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}

		account, err := transaction.MailAccount.Query().
			Where(mailaccount.IDEQ(request.AccountID)).
			ForUpdate().
			Only(ctx)
		if err != nil {
			return mailboxEntSafeError(err)
		}
		if account.ID != request.AccountID {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if account.Status != mailaccount.StatusActive || !account.DesiredEnabled {
			result = mailboxEntSyncNonClaim(now)
			return nil
		}
		if account.EngineAccountID == nil || !syncValidOpaque(*account.EngineAccountID, MaxEngineOpaqueIDBytes) || account.NextSyncAt == nil || !mailboxEntValidTime(*account.NextSyncAt) || !mailboxEntValidTime(account.EventCutoffAt) || account.SyncEpoch < 0 || account.MutationEpoch < 0 || account.SyncLeaseEpoch <= 0 {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if (account.MutationLeaseUntil != nil && !mailboxEntValidTime(*account.MutationLeaseUntil)) || (account.SyncLeaseUntil != nil && !mailboxEntValidTime(*account.SyncLeaseUntil)) {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if (account.MutationLeaseToken == nil) != (account.MutationLeaseUntil == nil) {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if account.MutationLeaseUntil != nil {
			if account.MutationLeaseUntil.After(now) {
				result = mailboxEntSyncNonClaim(*account.MutationLeaseUntil)
				return nil
			}
			nextMutationEpoch, ok := mailboxEntNextEpoch(account.MutationEpoch)
			if !ok {
				return mailboxServiceError(MailboxServiceErrorInvalid)
			}
			if err := account.Update().
				ClearMutationLeaseToken().
				ClearMutationLeaseUntil().
				SetMutationEpoch(nextMutationEpoch).
				SetSyncScheduleUpdatedAt(now).
				Exec(ctx); err != nil {
				return mailboxEntSafeError(err)
			}
			account.MutationLeaseToken = nil
			account.MutationLeaseUntil = nil
			account.MutationEpoch = nextMutationEpoch
		}

		persistedFence, hasLease, leaseValid := mailboxEntSyncLeaseFence(account)
		if !leaseValid {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if hasLease && account.SyncLeaseUntil.After(now) {
			if persistedFence.LeaseID != request.LeaseID {
				result = mailboxEntSyncNonClaim(*account.SyncLeaseUntil)
				return nil
			}
			if persistedFence.SyncEpoch != account.SyncEpoch || persistedFence.MutationEpoch != account.MutationEpoch {
				return mailboxServiceError(MailboxServiceErrorInvalid)
			}
			leaseUntil := request.LeaseUntil
			if account.SyncLeaseUntil.After(leaseUntil) {
				leaseUntil = *account.SyncLeaseUntil
			}
			if err := account.Update().
				SetSyncLeaseUntil(leaseUntil).
				SetSyncScheduleUpdatedAt(now).
				Exec(ctx); err != nil {
				return mailboxEntSafeError(err)
			}
			result = MailboxSyncClaimResult{
				Claimed:         true,
				Fence:           persistedFence,
				EngineAccountID: *account.EngineAccountID,
				EventCutoff:     account.EventCutoffAt,
			}
			return nil
		}
		if account.NextSyncAt.After(now) {
			result = mailboxEntSyncNonClaim(*account.NextSyncAt)
			return nil
		}

		leaseEpoch := account.SyncLeaseEpoch
		if hasLease {
			var ok bool
			leaseEpoch, ok = mailboxEntNextEpoch(leaseEpoch)
			if !ok {
				return mailboxServiceError(MailboxServiceErrorInvalid)
			}
		}
		fence, err := mailboxEntSyncClaimSnapshot(account)
		if err != nil {
			return err
		}
		fence.LeaseID = request.LeaseID
		fence.LeaseEpoch = leaseEpoch
		if !validMailboxSyncFence(fence) {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}

		update := account.Update().
			SetSyncLeaseToken(request.LeaseID).
			SetSyncLeaseUntil(request.LeaseUntil).
			SetSyncLeaseEpoch(leaseEpoch).
			SetSyncLeaseSyncEpoch(fence.SyncEpoch).
			SetSyncLeaseMutationEpoch(fence.MutationEpoch).
			ClearSyncLeaseCursor().
			ClearSyncLeaseBaseCursor().
			ClearSyncLeaseCheckpoint().
			ClearSyncLeaseResultHashVersion().
			ClearSyncLeaseResultHash().
			ClearSyncLeaseAppliedPhase().
			ClearSyncLeaseAppliedNextAt().
			ClearSyncLeaseAppliedAt().
			SetSyncScheduleUpdatedAt(now)
		phase, ok := mailboxEntSyncLeasePhaseValue(fence.Phase)
		if !ok {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		update.SetSyncLeasePhase(phase)
		switch fence.Phase {
		case MailboxSyncPhaseIncremental, MailboxSyncPhaseCatchUp:
			update.SetSyncLeaseCursor(fence.Cursor)
		case MailboxSyncPhaseEnumeration:
			if fence.BaseCursor != "" {
				update.SetSyncLeaseBaseCursor(fence.BaseCursor)
				update.SetSyncLeaseCheckpoint(fence.Checkpoint)
			}
		default:
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if err := update.Exec(ctx); err != nil {
			return mailboxEntSafeError(err)
		}
		result = MailboxSyncClaimResult{
			Claimed:         true,
			Fence:           fence,
			EngineAccountID: *account.EngineAccountID,
			EventCutoff:     account.EventCutoffAt,
		}
		return nil
	})
	if err != nil {
		return MailboxSyncClaimResult{}, mailboxEntSafeError(err)
	}
	return result, nil
}

func (repository *MailboxEntRepository) ResolveMailboxSyncStep(ctx context.Context, request MailboxSyncResolveRequest) (MailboxSyncResolveResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || !validMailboxSyncFence(request.Fence) || !mailboxEntValidTime(request.Now) || !mailboxEntValidTime(request.NextAttemptAt) || request.NextAttemptAt.Before(request.Now) || !mailboxEntValidResolveFailure(request.Failure, request.Fence.Phase) {
		return MailboxSyncResolveResult{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	var result MailboxSyncResolveResult
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
		transactionTime, err := mailboxEntTransactionTime(ctx, transaction)
		if err != nil {
			return err
		}
		if account.Status != mailaccount.StatusActive || !account.DesiredEnabled || account.EngineAccountID == nil || !syncValidOpaque(*account.EngineAccountID, MaxEngineOpaqueIDBytes) {
			return nil
		}
		persistedFence, hasLease, leaseValid := mailboxEntSyncLeaseFence(account)
		if !leaseValid || !hasLease || persistedFence != request.Fence || !account.SyncLeaseUntil.After(transactionTime) {
			return nil
		}
		if account.SyncAttempts < 0 || account.SyncAttempts >= mailboxEntMaxSyncAttempts {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}

		phase := request.Fence.Phase
		update := account.Update().
			SetProjectionState(mailaccount.ProjectionState("degraded")).
			SetSyncAttempts(account.SyncAttempts + 1).
			SetNextSyncAt(request.NextAttemptAt).
			SetSyncScheduleUpdatedAt(transactionTime).
			ClearSyncLeaseToken().
			ClearSyncLeaseUntil().
			ClearSyncLeaseSyncEpoch().
			ClearSyncLeaseMutationEpoch().
			ClearSyncLeasePhase().
			ClearSyncLeaseCursor().
			ClearSyncLeaseBaseCursor().
			ClearSyncLeaseCheckpoint().
			ClearSyncLeaseResultHashVersion().
			ClearSyncLeaseResultHash().
			ClearSyncLeaseAppliedPhase().
			ClearSyncLeaseAppliedNextAt().
			ClearSyncLeaseAppliedAt()
		switch request.Failure {
		case MailboxSyncFailureInvalidCursor:
			nextSyncEpoch, ok := mailboxEntNextEpoch(account.SyncEpoch)
			if !ok {
				return mailboxServiceError(MailboxServiceErrorInvalid)
			}
			update.SetSyncEpoch(nextSyncEpoch).
				SetProjectionState(mailaccount.ProjectionState("degraded")).
				ClearRebuildBaseCursor().
				ClearRebuildCheckpoint().
				SetSyncErrorCode(mailaccount.SyncErrorCode("MAIL_SYNC_CURSOR_INVALID"))
			phase = MailboxSyncPhaseEnumeration
		case MailboxSyncFailureRestartRequired:
			nextSyncEpoch, ok := mailboxEntNextEpoch(account.SyncEpoch)
			if !ok {
				return mailboxServiceError(MailboxServiceErrorInvalid)
			}
			update.SetSyncEpoch(nextSyncEpoch).
				SetProjectionState(mailaccount.ProjectionState("degraded")).
				ClearRebuildBaseCursor().
				ClearRebuildCheckpoint().
				SetSyncErrorCode(mailaccount.SyncErrorCode("MAIL_SYNC_RESTART_REQUIRED"))
			phase = MailboxSyncPhaseEnumeration
		case MailboxSyncFailureTransient:
			update.SetSyncErrorCode(mailaccount.SyncErrorCode("MAIL_SYNC_ENGINE_UNAVAILABLE"))
		case MailboxSyncFailureRejected:
			update.SetSyncErrorCode(mailaccount.SyncErrorCode("MAIL_SYNC_PAGE_REJECTED"))
		default:
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if err := update.Exec(ctx); err != nil {
			return mailboxEntSafeError(err)
		}
		result = MailboxSyncResolveResult{Resolved: true, Phase: phase, NextAttemptAt: request.NextAttemptAt}
		return nil
	})
	if err != nil {
		return MailboxSyncResolveResult{}, mailboxEntSafeError(err)
	}
	return result, nil
}

func mailboxEntSyncClaimSnapshot(account *ent.MailAccount) (MailboxSyncFence, error) {
	if account == nil || account.ID == uuid.Nil || account.SyncEpoch < 0 || account.MutationEpoch < 0 || account.SyncLeaseEpoch <= 0 || account.EngineAccountID == nil || !syncValidOpaque(*account.EngineAccountID, MaxEngineOpaqueIDBytes) || !mailboxEntValidTime(account.EventCutoffAt) {
		return MailboxSyncFence{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	fence := MailboxSyncFence{
		AccountID:     account.ID,
		SyncEpoch:     account.SyncEpoch,
		MutationEpoch: account.MutationEpoch,
	}
	switch string(account.ProjectionState) {
	case "bootstrapping":
		if account.SyncCursor != nil || account.RebuildBaseCursor != nil || account.RebuildCheckpoint != nil {
			return MailboxSyncFence{}, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		fence.Phase = MailboxSyncPhaseEnumeration
	case "ready":
		if account.SyncCursor == nil || !syncValidCursor(*account.SyncCursor, false) || account.RebuildBaseCursor != nil || account.RebuildCheckpoint != nil {
			return MailboxSyncFence{}, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		fence.Phase = MailboxSyncPhaseIncremental
		fence.Cursor = *account.SyncCursor
	case "degraded":
		if account.RebuildBaseCursor != nil || account.RebuildCheckpoint != nil {
			return MailboxSyncFence{}, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if account.SyncErrorCode != nil && (string(*account.SyncErrorCode) == "MAIL_SYNC_CURSOR_INVALID" || string(*account.SyncErrorCode) == "MAIL_SYNC_RESTART_REQUIRED") {
			fence.Phase = MailboxSyncPhaseEnumeration
			return fence, nil
		}
		if account.SyncCursor == nil || !syncValidCursor(*account.SyncCursor, false) {
			return MailboxSyncFence{}, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		fence.Phase = MailboxSyncPhaseIncremental
		fence.Cursor = *account.SyncCursor
	case "rebuilding":
		if account.RebuildBaseCursor == nil || !syncValidCursor(*account.RebuildBaseCursor, false) {
			return MailboxSyncFence{}, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if account.RebuildCheckpoint == nil {
			fence.Phase = MailboxSyncPhaseCatchUp
			fence.Cursor = *account.RebuildBaseCursor
			return fence, nil
		}
		if !syncValidCursor(*account.RebuildCheckpoint, false) {
			return MailboxSyncFence{}, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		fence.Phase = MailboxSyncPhaseEnumeration
		fence.BaseCursor = *account.RebuildBaseCursor
		fence.Checkpoint = *account.RebuildCheckpoint
	default:
		return MailboxSyncFence{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	return fence, nil
}

func mailboxEntSyncLeaseFence(account *ent.MailAccount) (MailboxSyncFence, bool, bool) {
	if account == nil {
		return MailboxSyncFence{}, false, false
	}
	if (account.SyncLeaseToken == nil) != (account.SyncLeaseUntil == nil) {
		return MailboxSyncFence{}, false, false
	}
	if account.SyncLeaseUntil != nil && !mailboxEntValidTime(*account.SyncLeaseUntil) {
		return MailboxSyncFence{}, false, false
	}
	if account.SyncLeaseToken == nil {
		if account.SyncLeaseSyncEpoch != nil || account.SyncLeaseMutationEpoch != nil || account.SyncLeasePhase != nil || account.SyncLeaseCursor != nil || account.SyncLeaseBaseCursor != nil || account.SyncLeaseCheckpoint != nil || account.SyncLeaseResultHashVersion != nil || account.SyncLeaseResultHash != nil || account.SyncLeaseAppliedPhase != nil || account.SyncLeaseAppliedNextAt != nil || account.SyncLeaseAppliedAt != nil {
			return MailboxSyncFence{}, false, false
		}
		return MailboxSyncFence{}, false, true
	}
	if account.SyncLeaseSyncEpoch == nil || account.SyncLeaseMutationEpoch == nil || account.SyncLeasePhase == nil {
		return MailboxSyncFence{}, true, false
	}
	phase, ok := mailboxEntSyncLeasePhaseFromStored(*account.SyncLeasePhase)
	if !ok {
		return MailboxSyncFence{}, true, false
	}
	fence := MailboxSyncFence{
		AccountID:     account.ID,
		LeaseID:       *account.SyncLeaseToken,
		LeaseEpoch:    account.SyncLeaseEpoch,
		SyncEpoch:     *account.SyncLeaseSyncEpoch,
		MutationEpoch: *account.SyncLeaseMutationEpoch,
		Phase:         phase,
	}
	if account.SyncLeaseCursor != nil {
		fence.Cursor = *account.SyncLeaseCursor
	}
	if account.SyncLeaseBaseCursor != nil {
		fence.BaseCursor = *account.SyncLeaseBaseCursor
	}
	if account.SyncLeaseCheckpoint != nil {
		fence.Checkpoint = *account.SyncLeaseCheckpoint
	}
	return fence, true, validMailboxSyncFence(fence)
}

func mailboxEntSyncLeasePhaseValue(phase MailboxSyncPhase) (mailaccount.SyncLeasePhase, bool) {
	switch phase {
	case MailboxSyncPhaseIncremental:
		return mailaccount.SyncLeasePhase("incremental"), true
	case MailboxSyncPhaseEnumeration:
		return mailaccount.SyncLeasePhase("enumeration"), true
	case MailboxSyncPhaseCatchUp:
		return mailaccount.SyncLeasePhase("catch_up"), true
	default:
		return "", false
	}
}

func mailboxEntSyncLeasePhaseFromStored(value mailaccount.SyncLeasePhase) (MailboxSyncPhase, bool) {
	switch string(value) {
	case "incremental":
		return MailboxSyncPhaseIncremental, true
	case "enumeration":
		return MailboxSyncPhaseEnumeration, true
	case "catch_up":
		return MailboxSyncPhaseCatchUp, true
	default:
		return 0, false
	}
}

func mailboxEntSyncNonClaim(nextAttemptAt time.Time) MailboxSyncClaimResult {
	return MailboxSyncClaimResult{NextAttemptAt: nextAttemptAt}
}

func mailboxEntNextEpoch(value int64) (int64, bool) {
	if value < 0 || value >= mailboxEntMaxInt64 {
		return 0, false
	}
	return value + 1, true
}

func mailboxEntValidResolveFailure(failure MailboxSyncFailureKind, phase MailboxSyncPhase) bool {
	switch failure {
	case MailboxSyncFailureInvalidCursor:
		return phase == MailboxSyncPhaseIncremental || phase == MailboxSyncPhaseCatchUp
	case MailboxSyncFailureRestartRequired:
		return phase == MailboxSyncPhaseEnumeration
	case MailboxSyncFailureTransient, MailboxSyncFailureRejected:
		return validMailboxSyncPhase(phase)
	default:
		return false
	}
}
