package mail

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/outboxevent"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

type mailboxEntSyncApplyProof struct {
	hashVersion int32
	hash        [sha256.Size]byte
	phase       MailboxSyncPhase
	nextAt      time.Time
	appliedAt   time.Time
}

type mailboxEntSyncApplyPosition struct {
	phase             MailboxSyncPhase
	cursor            string
	baseCursor        string
	checkpoint        string
	projectionState   mailaccount.ProjectionState
	updatesLastSynced bool
}

type mailboxEntSyncApplyNoWrite struct{}

func (mailboxEntSyncApplyNoWrite) Error() string {
	return "mailbox sync apply no write"
}

var errMailboxEntSyncApplyNoWrite = mailboxEntSyncApplyNoWrite{}

// ApplyMailboxSyncPage commits one canonical page while the originating lease
// remains fenced. The provider and policy work have already happened before
// this boundary; this method performs no external I/O.
func (repository *MailboxEntRepository) ApplyMailboxSyncPage(ctx context.Context, request MailboxSyncApplyRequest) (MailboxSyncApplyResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || !validMailboxSyncFence(request.Fence) || !mailboxEntValidTime(request.Now) || !mailboxEntValidTime(request.NextAttemptAt) {
		return MailboxSyncApplyResult{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	canonical, err := CanonicalizeMailboxSyncApplyRequest(request)
	if err != nil {
		return MailboxSyncApplyResult{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	position, err := mailboxEntSyncApplyExpectedPosition(canonical)
	if err != nil {
		return MailboxSyncApplyResult{}, err
	}

	var result MailboxSyncApplyResult
	err = repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, err := transaction.MailAccount.Query().
			Where(mailaccount.IDEQ(canonical.Fence.AccountID)).
			ForUpdate().
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return errMailboxEntSyncApplyNoWrite
			}
			return mailboxEntSafeError(err)
		}
		databaseWallClock, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return err
		}
		if account.Status != mailaccount.StatusActive || !account.DesiredEnabled {
			return errMailboxEntSyncApplyNoWrite
		}
		persistedFence, hasLease, leaseValid := mailboxEntSyncLeaseFence(account)
		if !leaseValid {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if !hasLease || persistedFence != canonical.Fence || account.SyncEpoch != canonical.Fence.SyncEpoch || account.MutationEpoch != canonical.Fence.MutationEpoch || account.SyncLeaseUntil == nil || !account.SyncLeaseUntil.After(databaseWallClock) {
			return errMailboxEntSyncApplyNoWrite
		}
		active, err := mailboxEntSyncApplyAccountValid(account, canonical.Fence.AccountID)
		if err != nil {
			return err
		}
		if !active {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		leaseUntil, ok := mailboxEntCheckedTimeAdd(canonical.NextAttemptAt, time.Second)
		if !ok || !leaseUntil.After(canonical.NextAttemptAt) {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}

		proof, hasProof, err := mailboxEntSyncApplyProofFromAccount(account, persistedFence)
		if err != nil {
			return err
		}
		if hasProof {
			if !mailboxEntSyncApplyProofMatches(canonical, proof) || !mailboxEntSyncApplyPostPositionMatches(account, canonical, position, proof, leaseUntil) {
				return mailboxServiceError(MailboxServiceErrorInvalid)
			}
			result = MailboxSyncApplyResult{Applied: true, Replayed: true, Phase: canonical.ExpectedPhase, NextAttemptAt: canonical.NextAttemptAt}
			return errMailboxEntSyncApplyNoWrite
		}

		if !canonical.NextAttemptAt.After(databaseWallClock) {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}

		prePosition, err := mailboxEntSyncClaimSnapshot(account)
		if err != nil {
			return err
		}
		if !mailboxEntSyncApplyFencePositionMatches(prePosition, canonical.Fence) {
			return errMailboxEntSyncApplyNoWrite
		}
		if err := mailboxEntSyncApplyOutboxCandidatesAvailable(ctx, transaction, canonical); err != nil {
			return err
		}
		if err := repository.mailboxEntApplySyncProjection(ctx, transaction, canonical, databaseWallClock); err != nil {
			return err
		}
		if err := mailboxEntSyncApplyAppendOutbox(ctx, transaction, canonical, databaseWallClock); err != nil {
			return err
		}
		if err := mailboxEntSyncApplyPersistAccount(ctx, account, canonical, position, databaseWallClock, leaseUntil); err != nil {
			return err
		}
		result = MailboxSyncApplyResult{Applied: true, Phase: canonical.ExpectedPhase, NextAttemptAt: canonical.NextAttemptAt}
		return nil
	})
	if err == errMailboxEntSyncApplyNoWrite {
		return result, nil
	}
	if err != nil {
		return MailboxSyncApplyResult{}, mailboxEntSafeError(err)
	}
	return result, nil
}

func mailboxEntSyncApplyAccountValid(account *ent.MailAccount, accountID uuid.UUID) (bool, error) {
	if account == nil || accountID == uuid.Nil || account.ID != accountID {
		return false, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if account.Status != mailaccount.StatusActive || !account.DesiredEnabled {
		return false, nil
	}
	if account.OrganizationID == uuid.Nil || account.UserID == uuid.Nil || account.MailDomainID == uuid.Nil || account.EngineAccountID == nil || !syncValidOpaque(*account.EngineAccountID, MaxEngineOpaqueIDBytes) || !mailboxEntValidTime(account.EventCutoffAt) || !mailboxEntValidTime(account.SyncScheduleUpdatedAt) || account.NextSyncAt == nil || !mailboxEntValidTime(*account.NextSyncAt) || account.SyncEpoch < 0 || account.MutationEpoch < 0 || account.SyncLeaseEpoch <= 0 || account.SyncAttempts < 0 || account.SyncAttempts >= mailboxEntMaxSyncAttempts {
		return false, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if account.LastSyncedAt != nil && !mailboxEntValidTime(*account.LastSyncedAt) {
		return false, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if (account.MutationLeaseToken == nil) != (account.MutationLeaseUntil == nil) || account.MutationLeaseToken != nil {
		return false, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if account.MutationLeaseUntil != nil && !mailboxEntValidTime(*account.MutationLeaseUntil) {
		return false, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if !mailboxEntSyncApplyErrorCodeValid(account.SyncErrorCode) {
		return false, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	return true, nil
}

func mailboxEntSyncApplyErrorCodeValid(value *mailaccount.SyncErrorCode) bool {
	if value == nil {
		return true
	}
	switch string(*value) {
	case "MAIL_SYNC_ENGINE_UNAVAILABLE", "MAIL_SYNC_CURSOR_INVALID", "MAIL_SYNC_RESTART_REQUIRED", "MAIL_SYNC_PAGE_REJECTED", "MAIL_SYNC_PROJECTION_LIMIT", "MAIL_SYNC_CONTROL_CONFLICT":
		return true
	default:
		return false
	}
}

func mailboxEntSyncApplyExpectedPosition(canonical MailboxSyncCanonicalApply) (mailboxEntSyncApplyPosition, error) {
	advance := canonical.Result.Advance
	if advance.Change != nil && advance.Enumeration == nil {
		if advance.Change.FromCursor != canonical.Fence.Cursor {
			return mailboxEntSyncApplyPosition{}, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		position := mailboxEntSyncApplyPosition{
			phase:            canonical.ExpectedPhase,
			cursor:           advance.Change.NextCursor,
			projectionState:  mailaccount.ProjectionState("ready"),
			updatesLastSynced: true,
		}
		if position.cursor == "" {
			position.cursor = advance.Change.FromCursor
		}
		if canonical.ExpectedPhase == MailboxSyncPhaseCatchUp {
			if advance.Change.NextCursor == "" {
				return mailboxEntSyncApplyPosition{}, mailboxServiceError(MailboxServiceErrorInvalid)
			}
			position.projectionState = mailaccount.ProjectionState("rebuilding")
			position.baseCursor = position.cursor
			position.updatesLastSynced = false
		}
		if canonical.ExpectedPhase != MailboxSyncPhaseIncremental && canonical.ExpectedPhase != MailboxSyncPhaseCatchUp {
			return mailboxEntSyncApplyPosition{}, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		return position, nil
	}
	if advance.Enumeration != nil && advance.Change == nil {
		if canonical.ExpectedPhase == MailboxSyncPhaseEnumeration && !advance.Enumeration.Complete {
			return mailboxEntSyncApplyPosition{
				phase:           MailboxSyncPhaseEnumeration,
				baseCursor:      advance.Enumeration.BaseCursor,
				checkpoint:      advance.Enumeration.NextCheckpoint,
				projectionState: mailaccount.ProjectionState("rebuilding"),
			}, nil
		}
		if canonical.ExpectedPhase == MailboxSyncPhaseCatchUp && advance.Enumeration.Complete {
			return mailboxEntSyncApplyPosition{
				phase:           MailboxSyncPhaseCatchUp,
				baseCursor:      advance.Enumeration.BaseCursor,
				projectionState: mailaccount.ProjectionState("rebuilding"),
			}, nil
		}
	}
	return mailboxEntSyncApplyPosition{}, mailboxServiceError(MailboxServiceErrorInvalid)
}

func mailboxEntSyncApplyProofFromAccount(account *ent.MailAccount, fence MailboxSyncFence) (mailboxEntSyncApplyProof, bool, error) {
	if account == nil {
		return mailboxEntSyncApplyProof{}, false, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	hasVersion := account.SyncLeaseResultHashVersion != nil
	hasHash := account.SyncLeaseResultHash != nil
	hasPhase := account.SyncLeaseAppliedPhase != nil
	hasNext := account.SyncLeaseAppliedNextAt != nil
	hasApplied := account.SyncLeaseAppliedAt != nil
	if !hasVersion && !hasHash && !hasPhase && !hasNext && !hasApplied {
		return mailboxEntSyncApplyProof{}, false, nil
	}
	if !hasVersion || !hasHash || !hasPhase || !hasNext || !hasApplied || account.SyncLeaseUntil == nil || (*account.SyncLeaseResultHashVersion != mailboxSyncCanonicalApplyVersion && *account.SyncLeaseResultHashVersion != int32(MailboxSyncCanonicalVersionV3)) || len(account.SyncLeaseResultHash) != sha256.Size || !mailboxEntValidTime(*account.SyncLeaseAppliedNextAt) || !mailboxEntValidTime(*account.SyncLeaseAppliedAt) {
		return mailboxEntSyncApplyProof{}, false, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	phase, ok := mailboxEntSyncLeaseAppliedPhaseFromStored(*account.SyncLeaseAppliedPhase)
	if !ok || !account.SyncLeaseAppliedNextAt.Before(*account.SyncLeaseUntil) || account.SyncLeaseAppliedNextAt.Before(*account.SyncLeaseAppliedAt) || !account.SyncLeaseAppliedAt.Before(*account.SyncLeaseUntil) {
		return mailboxEntSyncApplyProof{}, false, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	switch fence.Phase {
	case MailboxSyncPhaseIncremental:
		if phase != MailboxSyncPhaseIncremental {
			return mailboxEntSyncApplyProof{}, false, mailboxServiceError(MailboxServiceErrorInvalid)
		}
	case MailboxSyncPhaseCatchUp:
		if phase != MailboxSyncPhaseCatchUp && phase != MailboxSyncPhaseIncremental {
			return mailboxEntSyncApplyProof{}, false, mailboxServiceError(MailboxServiceErrorInvalid)
		}
	case MailboxSyncPhaseEnumeration:
		if phase != MailboxSyncPhaseEnumeration && phase != MailboxSyncPhaseCatchUp {
			return mailboxEntSyncApplyProof{}, false, mailboxServiceError(MailboxServiceErrorInvalid)
		}
	default:
		return mailboxEntSyncApplyProof{}, false, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	proof := mailboxEntSyncApplyProof{
		hashVersion: *account.SyncLeaseResultHashVersion,
		phase:       phase,
		nextAt:      *account.SyncLeaseAppliedNextAt,
		appliedAt:   *account.SyncLeaseAppliedAt,
	}
	copy(proof.hash[:], account.SyncLeaseResultHash)
	return proof, true, nil
}

func mailboxEntSyncApplyProofMatches(canonical MailboxSyncCanonicalApply, proof mailboxEntSyncApplyProof) bool {
	hashMatches := false
	switch proof.hashVersion {
	case mailboxSyncCanonicalApplyVersion:
		hashMatches = subtle.ConstantTimeCompare(proof.hash[:], canonical.LegacyV2ResultHash[:]) == 1
	case int32(MailboxSyncCanonicalVersionV3):
		hashMatches = subtle.ConstantTimeCompare(proof.hash[:], canonical.V3ResultHash[:]) == 1
	default:
		return false
	}
	return hashMatches && proof.phase == canonical.ExpectedPhase && proof.nextAt.Equal(canonical.NextAttemptAt)
}

func mailboxEntSyncLeaseAppliedPhaseValue(phase MailboxSyncPhase) (mailaccount.SyncLeaseAppliedPhase, bool) {
	switch phase {
	case MailboxSyncPhaseIncremental:
		return mailaccount.SyncLeaseAppliedPhase("incremental"), true
	case MailboxSyncPhaseEnumeration:
		return mailaccount.SyncLeaseAppliedPhase("enumeration"), true
	case MailboxSyncPhaseCatchUp:
		return mailaccount.SyncLeaseAppliedPhase("catch_up"), true
	default:
		return "", false
	}
}

func mailboxEntSyncLeaseAppliedPhaseFromStored(value mailaccount.SyncLeaseAppliedPhase) (MailboxSyncPhase, bool) {
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

func mailboxEntSyncApplyPostPositionMatches(account *ent.MailAccount, canonical MailboxSyncCanonicalApply, position mailboxEntSyncApplyPosition, proof mailboxEntSyncApplyProof, leaseUntil time.Time) bool {
	if account == nil || account.SyncAttempts != 0 || account.SyncErrorCode != nil || account.ProjectionState != position.projectionState || account.NextSyncAt == nil || !mailboxEntValidTime(*account.NextSyncAt) || !account.NextSyncAt.Equal(canonical.NextAttemptAt) || !mailboxEntValidTime(account.SyncScheduleUpdatedAt) || !account.SyncScheduleUpdatedAt.Equal(proof.appliedAt) || account.SyncLeaseUntil == nil || !account.SyncLeaseUntil.Equal(leaseUntil) {
		return false
	}
	postPosition, err := mailboxEntSyncClaimSnapshot(account)
	if err != nil || !mailboxEntSyncApplyReplayPositionMatches(postPosition, canonical, position) {
		return false
	}
	if position.updatesLastSynced {
		return account.LastSyncedAt != nil && mailboxEntValidTime(*account.LastSyncedAt) && account.LastSyncedAt.Equal(proof.appliedAt)
	}
	return account.LastSyncedAt == nil || mailboxEntValidTime(*account.LastSyncedAt)
}

func mailboxEntSyncApplyReplayPositionMatches(postPosition MailboxSyncFence, canonical MailboxSyncCanonicalApply, position mailboxEntSyncApplyPosition) bool {
	if position.phase != canonical.ExpectedPhase ||
		postPosition.AccountID != canonical.Fence.AccountID ||
		postPosition.SyncEpoch != canonical.Fence.SyncEpoch ||
		postPosition.MutationEpoch != canonical.Fence.MutationEpoch ||
		postPosition.Phase != position.phase {
		return false
	}

	switch position.phase {
	case MailboxSyncPhaseIncremental:
		return postPosition.Cursor == position.cursor &&
			postPosition.BaseCursor == "" &&
			postPosition.Checkpoint == "" &&
			position.baseCursor == "" &&
			position.checkpoint == ""
	case MailboxSyncPhaseEnumeration:
		return postPosition.Cursor == "" &&
			postPosition.BaseCursor == position.baseCursor &&
			postPosition.Checkpoint == position.checkpoint &&
			position.cursor == "" &&
			position.baseCursor != "" &&
			position.checkpoint != ""
	case MailboxSyncPhaseCatchUp:
		return postPosition.Cursor == position.baseCursor &&
			postPosition.BaseCursor == "" &&
			postPosition.Checkpoint == "" &&
			position.baseCursor != "" &&
			position.checkpoint == "" &&
			(position.cursor == "" || position.cursor == position.baseCursor)
	default:
		return false
	}
}

func mailboxEntSyncApplyFencePositionMatches(position MailboxSyncFence, fence MailboxSyncFence) bool {
	return position.AccountID == fence.AccountID &&
		position.SyncEpoch == fence.SyncEpoch &&
		position.MutationEpoch == fence.MutationEpoch &&
		position.Phase == fence.Phase &&
		position.Cursor == fence.Cursor &&
		position.BaseCursor == fence.BaseCursor &&
		position.Checkpoint == fence.Checkpoint
}

func mailboxEntSyncApplyOutboxCandidatesAvailable(ctx context.Context, transaction *ent.Tx, canonical MailboxSyncCanonicalApply) error {
	if ctx == nil || transaction == nil || canonical.Result.Events == nil {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if len(canonical.Result.Events) == 0 {
		return nil
	}
	ids := make([]uuid.UUID, 0, len(canonical.Result.Events))
	seen := make(map[uuid.UUID]struct{}, len(canonical.Result.Events))
	for _, envelope := range canonical.Result.Events {
		eventID := envelope.EventID()
		subject := envelope.Subject()
		if eventID == uuid.Nil || subject == "" {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if _, duplicate := seen[eventID]; duplicate {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		seen[eventID] = struct{}{}
		ids = append(ids, eventID)
	}
	rows, err := transaction.OutboxEvent.Query().Where(outboxevent.IDIn(ids...)).All(ctx)
	if err != nil {
		return mailboxEntSafeError(err)
	}
	if len(rows) != 0 {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	return nil
}

func mailboxEntSyncApplyAppendOutbox(ctx context.Context, transaction *ent.Tx, canonical MailboxSyncCanonicalApply, transactionTime time.Time) error {
	if ctx == nil || transaction == nil || !mailboxEntValidTime(transactionTime) || canonical.Result.Events == nil {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	traceContext := natsclient.TraceContextMap(ctx)
	for _, envelope := range canonical.Result.Events {
		eventID := envelope.EventID()
		subject := envelope.Subject()
		payload, err := envelope.MarshalJSON()
		if err != nil || eventID == uuid.Nil || subject == "" || len(payload) == 0 {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		create := transaction.OutboxEvent.Create().
			SetID(eventID).
			SetEventType(subject).
			SetPayload(append([]byte(nil), payload...)).
			SetCreatedAt(transactionTime)
		if traceContext != nil {
			create.SetTraceContext(traceContext)
		}
		if _, err := create.Save(ctx); err != nil {
			return mailboxEntSafeError(err)
		}
	}
	return nil
}

func mailboxEntSyncApplyPersistAccount(ctx context.Context, account *ent.MailAccount, canonical MailboxSyncCanonicalApply, position mailboxEntSyncApplyPosition, transactionTime, leaseUntil time.Time) error {
	if ctx == nil || account == nil || canonical.Version != MailboxSyncCanonicalVersionV3 || !mailboxEntValidTime(transactionTime) || !mailboxEntValidTime(leaseUntil) || !leaseUntil.After(canonical.NextAttemptAt) {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	phase, ok := mailboxEntSyncLeaseAppliedPhaseValue(canonical.ExpectedPhase)
	if !ok {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	hash := append([]byte(nil), canonical.V3ResultHash[:]...)
	update := account.Update().
		SetSyncAttempts(0).
		SetNextSyncAt(canonical.NextAttemptAt).
		SetSyncScheduleUpdatedAt(transactionTime).
		SetSyncLeaseUntil(leaseUntil).
		SetSyncLeaseResultHashVersion(int32(canonical.Version)).
		SetSyncLeaseResultHash(hash).
		SetSyncLeaseAppliedPhase(phase).
		SetSyncLeaseAppliedNextAt(canonical.NextAttemptAt).
		SetSyncLeaseAppliedAt(transactionTime).
		ClearSyncErrorCode().
		SetUpdatedAt(transactionTime)
	if position.projectionState == mailaccount.ProjectionState("ready") {
		update.SetProjectionState(mailaccount.ProjectionState("ready")).
			SetSyncCursor(position.cursor).
			ClearRebuildBaseCursor().
			ClearRebuildCheckpoint().
			SetLastSyncedAt(transactionTime)
	} else if position.projectionState == mailaccount.ProjectionState("rebuilding") {
		update.SetProjectionState(mailaccount.ProjectionState("rebuilding")).
			SetRebuildBaseCursor(position.baseCursor)
		if position.checkpoint == "" {
			update.ClearRebuildCheckpoint()
		} else {
			update.SetRebuildCheckpoint(position.checkpoint)
		}
	} else {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if err := update.Exec(ctx); err != nil {
		return mailboxEntSafeError(err)
	}
	return nil
}
