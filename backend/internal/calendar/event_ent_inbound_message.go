package calendar

import (
	"bytes"
	"context"
	"crypto/sha256"
	"math"
	"time"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/calendarinboundmessagereceipt"
	"github.com/valerii/hub/backend/ent/predicate"
)

const (
	calendarInboundMessageLeaseTTL           = 60 * time.Second
	calendarInboundMessageMaximumAttempts    = 8
	calendarInboundMessageFinalizedRetention = 30 * 24 * time.Hour
	calendarInboundMessageDeferredLifetime   = 7 * 24 * time.Hour
)

const (
	calendarInboundMessageReceiptStatePending               = "pending"
	calendarInboundMessageReceiptStateProcessing            = "processing"
	calendarInboundMessageReceiptStateApplied               = "applied"
	calendarInboundMessageReceiptStateIgnoredStale          = "ignored_stale"
	calendarInboundMessageReceiptStateDuplicate             = "duplicate"
	calendarInboundMessageReceiptStateDeferredMissingMaster = "deferred_missing_master"
	calendarInboundMessageReceiptStateQuarantined           = "quarantined"
	calendarInboundMessageReceiptStateDeadLetter            = "dead_letter"

	calendarInboundMessageReceiptErrorMissingMasterExpired = "missing_master_expired"
	calendarInboundMessageReceiptErrorRetryExhausted       = "retry_exhausted"
)

const calendarInboundMessageInsertPendingSQL = `
INSERT INTO calendar_inbound_message_receipts (
	id,
	organization_id,
	mailbox_id,
	mail_message_id,
	part_locator_hash,
	event_envelope_id,
	state,
	attempts,
	lease_generation,
	created_at,
	updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
ON CONFLICT DO NOTHING`

var _ CalendarInboundMessageRepository = (*EventEntRepository)(nil)

// RegisterAndClaimCalendarInboundMessage registers one immutable receipt and
// returns one processing lease only after the durable row has been locked.
func (repository *EventEntRepository) RegisterAndClaimCalendarInboundMessage(
	ctx context.Context,
	registration CalendarInboundMessageRegistration,
) (CalendarInboundMessageClaimResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || !registration.valid() {
		return CalendarInboundMessageClaimResult{}, ErrEventInvalid
	}

	var result CalendarInboundMessageClaimResult
	err := repository.withCalendarInboundMessageTransaction(ctx, func(transaction *ent.Tx, clock time.Time) error {
		if err := calendarInboundMessageInsertPending(ctx, transaction, registration, clock); err != nil {
			return err
		}
		receipt, err := calendarInboundMessageLockRegistration(ctx, transaction, registration)
		if err != nil {
			return err
		}
		work, claimed, err := calendarInboundMessageClaimReceipt(ctx, transaction, receipt, clock)
		if err != nil {
			return err
		}
		if !claimed {
			result = NewCalendarInboundMessageDurableNoopResult()
			return nil
		}
		result, err = NewCalendarInboundMessageClaimedResult(work)
		if err != nil {
			return ErrEventInvalid
		}
		return nil
	})
	if err != nil {
		return CalendarInboundMessageClaimResult{}, err
	}
	if !result.valid() {
		return CalendarInboundMessageClaimResult{}, ErrEventInvalid
	}
	return result, nil
}

// ClaimDueCalendarInboundMessages claims at most limit receipts whose durable
// lease or retry time is due, without ever scanning ingress-only initial rows.
func (repository *EventEntRepository) ClaimDueCalendarInboundMessages(
	ctx context.Context,
	limit int,
) ([]CalendarInboundMessageWork, error) {
	if repository == nil || repository.client == nil || ctx == nil || !validCalendarInboundMessageDueClaimLimit(limit) {
		return nil, ErrEventInvalid
	}

	claimed := make([]CalendarInboundMessageWork, 0, limit)
	err := repository.withCalendarInboundMessageTransaction(ctx, func(transaction *ent.Tx, clock time.Time) error {
		rows, err := transaction.CalendarInboundMessageReceipt.Query().
			Where(
				calendarInboundMessageDueScanSkipLocked(),
				calendarinboundmessagereceipt.Or(
					calendarinboundmessagereceipt.And(
						calendarinboundmessagereceipt.StateEQ(calendarinboundmessagereceipt.State(calendarInboundMessageReceiptStateProcessing)),
						calendarinboundmessagereceipt.LeaseUntilNotNil(),
						calendarinboundmessagereceipt.LeaseUntilLTE(clock),
					),
					calendarinboundmessagereceipt.And(
						calendarinboundmessagereceipt.StateIn(
							calendarinboundmessagereceipt.State(calendarInboundMessageReceiptStatePending),
							calendarinboundmessagereceipt.State(calendarInboundMessageReceiptStateDeferredMissingMaster),
						),
						calendarinboundmessagereceipt.NextAttemptAtNotNil(),
						calendarinboundmessagereceipt.NextAttemptAtLTE(clock),
					),
				),
			).
			Limit(limit).
			All(ctx)
		if err != nil {
			return calendarEntSafeError(err)
		}
		if len(rows) > limit {
			return ErrEventInvalid
		}
		for _, receipt := range rows {
			work, claimedReceipt, claimErr := calendarInboundMessageClaimReceipt(ctx, transaction, receipt, clock)
			if claimErr != nil {
				return claimErr
			}
			if claimedReceipt {
				claimed = append(claimed, work)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(claimed) > limit {
		return nil, ErrEventInvalid
	}
	for _, work := range claimed {
		if !work.valid() {
			return nil, ErrEventInvalid
		}
	}
	return append([]CalendarInboundMessageWork(nil), claimed...), nil
}

// CleanupFinalizedCalendarInboundMessages deletes only bounded retained final
// receipt rows; pending, processing, and deferred rows are never candidates.
func (repository *EventEntRepository) CleanupFinalizedCalendarInboundMessages(
	ctx context.Context,
	limit int,
) (int, error) {
	if repository == nil || repository.client == nil || ctx == nil || !validCalendarInboundMessageFinalizedCleanupLimit(limit) {
		return 0, ErrEventInvalid
	}

	removed := 0
	err := repository.withCalendarInboundMessageTransaction(ctx, func(transaction *ent.Tx, clock time.Time) error {
		cutoff := clock.Add(-calendarInboundMessageFinalizedRetention)
		if !calendarInboundMessageWholeSecondUTC(cutoff) {
			return ErrEventInvalid
		}
		rows, err := transaction.CalendarInboundMessageReceipt.Query().
			Where(
				calendarInboundMessageCleanupSkipLocked(),
				calendarinboundmessagereceipt.StateIn(calendarInboundMessageFinalizedStates()...),
				calendarinboundmessagereceipt.UpdatedAtLTE(cutoff),
			).
			Order(
				ent.Asc(calendarinboundmessagereceipt.FieldUpdatedAt),
				ent.Asc(calendarinboundmessagereceipt.FieldID),
			).
			Limit(limit).
			All(ctx)
		if err != nil {
			return calendarEntSafeError(err)
		}
		if len(rows) > limit {
			return ErrEventInvalid
		}
		ids := make([]uuid.UUID, 0, len(rows))
		seen := make(map[uuid.UUID]struct{}, len(rows))
		for _, receipt := range rows {
			if !calendarInboundMessageReceiptValid(receipt) ||
				!calendarInboundMessageReceiptFinalized(receipt) ||
				receipt.UpdatedAt.After(cutoff) || receipt.ID == uuid.Nil {
				return ErrEventInvalid
			}
			if _, duplicate := seen[receipt.ID]; duplicate {
				return ErrEventInvalid
			}
			seen[receipt.ID] = struct{}{}
			ids = append(ids, receipt.ID)
		}
		if len(ids) == 0 {
			return nil
		}
		affected, err := transaction.CalendarInboundMessageReceipt.Delete().
			Where(calendarinboundmessagereceipt.IDIn(ids...)).
			Exec(ctx)
		if err != nil {
			return calendarEntSafeError(err)
		}
		if affected != len(ids) || affected > limit {
			return ErrEventInvalid
		}
		removed = affected
		return nil
	})
	if err != nil {
		return 0, err
	}
	if removed < 0 || removed > limit {
		return 0, ErrEventInvalid
	}
	return removed, nil
}

// LockCalendarInboundMessageReceipt locks one live processing receipt in the
// existing Event transaction without changing its fence or lease.
func (backend *eventEntTransactionBackend) LockCalendarInboundMessageReceipt(
	ctx context.Context,
	claim CalendarInboundMessageClaim,
) (CalendarInboundMessageScope, bool, error) {
	if backend == nil || backend.tx == nil || ctx == nil || !claim.valid() {
		return CalendarInboundMessageScope{}, false, ErrEventInvalid
	}

	receipt, found, err := calendarInboundMessageLockClaimReceipt(ctx, backend.tx, claim)
	if err != nil {
		return CalendarInboundMessageScope{}, false, err
	}
	if !found {
		return CalendarInboundMessageScope{}, false, nil
	}
	if !calendarInboundMessageReceiptValid(receipt) || !calendarInboundMessageReceiptMatchesClaim(receipt, claim) {
		return CalendarInboundMessageScope{}, false, ErrEventInvalid
	}
	clock, err := backend.calendarInboundMessageDatabaseClock(ctx)
	if err != nil || !calendarInboundMessageWholeSecondUTC(clock) {
		if err != nil {
			return CalendarInboundMessageScope{}, false, err
		}
		return CalendarInboundMessageScope{}, false, ErrEventInvalid
	}
	if !receipt.LeaseUntil.After(clock) {
		return CalendarInboundMessageScope{}, false, nil
	}
	return CalendarInboundMessageScope{claim: claim}, true, nil
}

// CompleteCalendarInboundMessageReceipt is the fenced last operation of an
// Event transaction. It derives every durable due value from its cached DB
// clock and preserves the immutable receipt registration and deferral pair.
func (backend *eventEntTransactionBackend) CompleteCalendarInboundMessageReceipt(
	ctx context.Context,
	scope CalendarInboundMessageScope,
	completion CalendarInboundMessageCompletion,
) error {
	if backend == nil || backend.tx == nil || ctx == nil || scope.transaction == nil || !scope.claim.valid() || !completion.valid() {
		return ErrEventInvalid
	}
	clock, err := backend.calendarInboundMessageDatabaseClock(ctx)
	if err != nil || !calendarInboundMessageWholeSecondUTC(clock) {
		if err != nil {
			return err
		}
		return ErrEventInvalid
	}
	receipt, found, err := calendarInboundMessageLockClaimReceipt(ctx, backend.tx, scope.claim)
	if err != nil {
		return err
	}
	if !found || !calendarInboundMessageReceiptValid(receipt) ||
		!calendarInboundMessageReceiptMatchesClaim(receipt, scope.claim) || !receipt.LeaseUntil.After(clock) {
		return ErrEventConflict
	}
	return calendarInboundMessageCompleteReceipt(ctx, backend.tx, receipt, scope.claim, completion, clock)
}

func calendarInboundMessageLockClaimReceipt(
	ctx context.Context,
	transaction *ent.Tx,
	claim CalendarInboundMessageClaim,
) (*ent.CalendarInboundMessageReceipt, bool, error) {
	if ctx == nil || transaction == nil || !claim.valid() {
		return nil, false, ErrEventInvalid
	}
	receipt, err := transaction.CalendarInboundMessageReceipt.Query().
		Where(
			calendarinboundmessagereceipt.IDEQ(claim.ReceiptID),
			calendarinboundmessagereceipt.StateEQ(calendarinboundmessagereceipt.State(calendarInboundMessageReceiptStateProcessing)),
			calendarinboundmessagereceipt.LeaseTokenEQ(claim.LeaseToken),
			calendarinboundmessagereceipt.LeaseGenerationEQ(claim.LeaseGeneration),
		).
		ForUpdate().
		Only(ctx)
	if ent.IsNotFound(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, calendarEntSafeError(err)
	}
	return receipt, true, nil
}

func calendarInboundMessageReceiptMatchesClaim(
	receipt *ent.CalendarInboundMessageReceipt,
	claim CalendarInboundMessageClaim,
) bool {
	return receipt != nil && claim.valid() && receipt.ID == claim.ReceiptID &&
		calendarInboundMessageReceiptState(receipt) == calendarInboundMessageReceiptStateProcessing &&
		receipt.LeaseToken != nil && *receipt.LeaseToken == claim.LeaseToken &&
		receipt.LeaseGeneration == claim.LeaseGeneration
}

func (repository *EventEntRepository) withCalendarInboundMessageTransaction(
	ctx context.Context,
	callback func(*ent.Tx, time.Time) error,
) error {
	if repository == nil || repository.client == nil || ctx == nil || callback == nil {
		return ErrEventInvalid
	}
	transaction, err := repository.client.Tx(ctx)
	if err != nil {
		return calendarEntSafeError(err)
	}
	committed := false
	defer func() {
		if recovered := recover(); recovered != nil {
			_ = transaction.Rollback()
			panic(recovered)
		}
		if !committed {
			_ = transaction.Rollback()
		}
	}()

	clock, err := calendarInboundInviteDatabaseClock(ctx, transaction)
	if err != nil || !calendarInboundMessageWholeSecondUTC(clock) {
		if err != nil {
			return err
		}
		return ErrEventInvalid
	}
	if err := callback(transaction, clock); err != nil {
		return err
	}
	if err := transaction.Commit(); err != nil {
		return calendarEntSafeError(err)
	}
	committed = true
	return nil
}

func calendarInboundMessageInsertPending(
	ctx context.Context,
	transaction *ent.Tx,
	registration CalendarInboundMessageRegistration,
	clock time.Time,
) error {
	if ctx == nil || transaction == nil || !registration.valid() || !calendarInboundMessageWholeSecondUTC(clock) {
		return ErrEventInvalid
	}
	hash := registration.LocatorHash()
	id := uuid.New()
	if id == uuid.Nil {
		return ErrEventInvalid
	}
	if err := transaction.Driver().Exec(
		ctx,
		calendarInboundMessageInsertPendingSQL,
		[]any{
			id,
			registration.OrganizationID(),
			registration.MailboxID(),
			registration.MailMessageID(),
			hash[:],
			registration.EnvelopeID(),
			calendarInboundMessageReceiptStatePending,
			0,
			int64(0),
			clock,
			clock,
		},
		nil,
	); err != nil {
		return calendarEntSafeError(err)
	}
	return nil
}

func calendarInboundMessageLockRegistration(
	ctx context.Context,
	transaction *ent.Tx,
	registration CalendarInboundMessageRegistration,
) (*ent.CalendarInboundMessageReceipt, error) {
	if ctx == nil || transaction == nil || !registration.valid() {
		return nil, ErrEventInvalid
	}
	hash := registration.LocatorHash()
	rows, err := transaction.CalendarInboundMessageReceipt.Query().
		Where(calendarinboundmessagereceipt.Or(
			calendarinboundmessagereceipt.EventEnvelopeIDEQ(registration.EnvelopeID()),
			calendarinboundmessagereceipt.And(
				calendarinboundmessagereceipt.MailMessageIDEQ(registration.MailMessageID()),
				calendarinboundmessagereceipt.PartLocatorHashEQ(hash[:]),
			),
		)).
		Order(ent.Asc(calendarinboundmessagereceipt.FieldID)).
		ForUpdate().
		Limit(3).
		All(ctx)
	if err != nil {
		return nil, calendarEntSafeError(err)
	}
	if len(rows) != 1 || !calendarInboundMessageReceiptMatchesRegistration(rows[0], registration) || !calendarInboundMessageReceiptValid(rows[0]) {
		return nil, ErrEventInvalid
	}
	return rows[0], nil
}

func calendarInboundMessageClaimReceipt(
	ctx context.Context,
	transaction *ent.Tx,
	receipt *ent.CalendarInboundMessageReceipt,
	clock time.Time,
) (CalendarInboundMessageWork, bool, error) {
	if ctx == nil || transaction == nil || !calendarInboundMessageWholeSecondUTC(clock) || !calendarInboundMessageReceiptValid(receipt) {
		return CalendarInboundMessageWork{}, false, ErrEventInvalid
	}
	if calendarInboundMessageReceiptFinalized(receipt) {
		return CalendarInboundMessageWork{}, false, nil
	}

	switch calendarInboundMessageReceiptState(receipt) {
	case calendarInboundMessageReceiptStateProcessing:
		if receipt.LeaseUntil.After(clock) {
			return CalendarInboundMessageWork{}, false, nil
		}
		if receipt.Attempts == calendarInboundMessageMaximumAttempts {
			if err := calendarInboundMessageTerminalizeReceipt(
				ctx,
				receipt,
				calendarInboundMessageReceiptErrorRetryExhausted,
				clock,
			); err != nil {
				return CalendarInboundMessageWork{}, false, err
			}
			return CalendarInboundMessageWork{}, false, nil
		}
	case calendarInboundMessageReceiptStatePending:
		if receipt.NextAttemptAt != nil && receipt.NextAttemptAt.After(clock) {
			return CalendarInboundMessageWork{}, false, nil
		}
		if receipt.Attempts >= calendarInboundMessageMaximumAttempts {
			return CalendarInboundMessageWork{}, false, ErrEventInvalid
		}
	case calendarInboundMessageReceiptStateDeferredMissingMaster:
		if receipt.NextAttemptAt.After(clock) {
			return CalendarInboundMessageWork{}, false, nil
		}
		if !receipt.DeferredExpiresAt.After(clock) || receipt.Attempts+1 >= calendarInboundMessageMaximumAttempts {
			if err := calendarInboundMessageTerminalizeReceipt(
				ctx,
				receipt,
				calendarInboundMessageReceiptErrorMissingMasterExpired,
				clock,
			); err != nil {
				return CalendarInboundMessageWork{}, false, err
			}
			return CalendarInboundMessageWork{}, false, nil
		}
	default:
		return CalendarInboundMessageWork{}, false, ErrEventInvalid
	}

	if receipt.LeaseGeneration == math.MaxInt64 || receipt.Attempts >= calendarInboundMessageMaximumAttempts {
		return CalendarInboundMessageWork{}, false, ErrEventInvalid
	}
	leaseUntil := clock.Add(calendarInboundMessageLeaseTTL)
	if !calendarInboundMessageWholeSecondUTC(leaseUntil) || !leaseUntil.After(clock) {
		return CalendarInboundMessageWork{}, false, ErrEventInvalid
	}
	leaseToken := uuid.New()
	if leaseToken == uuid.Nil {
		return CalendarInboundMessageWork{}, false, ErrEventInvalid
	}
	registration, err := calendarInboundMessageRegistrationFromReceipt(receipt)
	if err != nil {
		return CalendarInboundMessageWork{}, false, err
	}
	claim := CalendarInboundMessageClaim{
		ReceiptID:       receipt.ID,
		LeaseToken:      leaseToken,
		LeaseGeneration: receipt.LeaseGeneration + 1,
	}
	work, err := NewCalendarInboundMessageWork(registration, claim)
	if err != nil {
		return CalendarInboundMessageWork{}, false, ErrEventInvalid
	}
	if err := transaction.CalendarInboundMessageReceipt.UpdateOne(receipt).
		SetState(calendarinboundmessagereceipt.State(calendarInboundMessageReceiptStateProcessing)).
		SetAttempts(receipt.Attempts + 1).
		SetLeaseToken(leaseToken).
		SetLeaseGeneration(claim.LeaseGeneration).
		SetLeaseUntil(leaseUntil).
		ClearNextAttemptAt().
		ClearErrorCode().
		SetUpdatedAt(clock).
		Exec(ctx); err != nil {
		return CalendarInboundMessageWork{}, false, calendarEntSafeError(err)
	}
	return work, true, nil
}

func calendarInboundMessageTerminalizeReceipt(
	ctx context.Context,
	receipt *ent.CalendarInboundMessageReceipt,
	errorCode string,
	clock time.Time,
) error {
	if ctx == nil || !calendarInboundMessageReceiptValid(receipt) || !calendarInboundMessageWholeSecondUTC(clock) ||
		!calendarInboundMessageDeadLetterErrorCodeValid(errorCode) {
		return ErrEventInvalid
	}
	if err := receipt.Update().
		SetState(calendarinboundmessagereceipt.State(calendarInboundMessageReceiptStateDeadLetter)).
		SetErrorCode(calendarinboundmessagereceipt.ErrorCode(errorCode)).
		ClearLeaseToken().
		ClearLeaseUntil().
		ClearNextAttemptAt().
		SetUpdatedAt(clock).
		Exec(ctx); err != nil {
		return calendarEntSafeError(err)
	}
	return nil
}

func calendarInboundMessageReceiptSnapshotFromPersistedFields(
	calendarComponentScope *calendarinboundmessagereceipt.CalendarComponentScope,
	recurrenceID *string,
) (calendarInboundMessageReceiptSnapshot, bool) {
	if calendarComponentScope == nil {
		if recurrenceID == nil {
			return calendarInboundMessageReceiptSnapshot{}, true
		}
		return calendarInboundMessageReceiptSnapshot{}, false
	}

	switch string(*calendarComponentScope) {
	case calendarInboundMessageReceiptSnapshotScopeSeries:
		if recurrenceID != nil {
			return calendarInboundMessageReceiptSnapshot{}, false
		}
		return calendarInboundMessageReceiptSnapshot{
			calendarComponentScope: calendarInboundMessageReceiptSnapshotScopeSeries,
		}, true
	case calendarInboundMessageReceiptSnapshotScopeOccurrence:
		if recurrenceID == nil || !validCalendarInboundRecurrenceID(*recurrenceID) {
			return calendarInboundMessageReceiptSnapshot{}, false
		}
		return calendarInboundMessageReceiptSnapshot{
			calendarComponentScope: calendarInboundMessageReceiptSnapshotScopeOccurrence,
			recurrenceID:           cloneCalendarInboundMessageStringPointer(recurrenceID),
		}, true
	default:
		return calendarInboundMessageReceiptSnapshot{}, false
	}
}

func calendarInboundMessageReceiptSnapshotEqual(
	left, right calendarInboundMessageReceiptSnapshot,
) bool {
	if !left.valid() || !right.valid() || left.calendarComponentScope != right.calendarComponentScope {
		return false
	}
	if left.recurrenceID == nil || right.recurrenceID == nil {
		return left.recurrenceID == nil && right.recurrenceID == nil
	}
	return *left.recurrenceID == *right.recurrenceID
}

func calendarInboundMessageCompleteReceipt(
	ctx context.Context,
	transaction *ent.Tx,
	receipt *ent.CalendarInboundMessageReceipt,
	claim CalendarInboundMessageClaim,
	completion CalendarInboundMessageCompletion,
	clock time.Time,
) error {
	if ctx == nil || transaction == nil || !calendarInboundMessageWholeSecondUTC(clock) ||
		!calendarInboundMessageReceiptValid(receipt) || !calendarInboundMessageReceiptMatchesClaim(receipt, claim) ||
		!receipt.LeaseUntil.After(clock) || !completion.valid() {
		return ErrEventInvalid
	}
	persistedSnapshot, persistedSnapshotValid := calendarInboundMessageReceiptSnapshotFromPersistedFields(
		receipt.CalendarComponentScope,
		receipt.RecurrenceID,
	)
	if !persistedSnapshotValid {
		return ErrEventInvalid
	}
	completionSnapshot := completion.receiptSnapshot
	setReceiptSnapshot := completionSnapshot.calendarComponentScope != ""
	var receiptSnapshotPredicate predicate.CalendarInboundMessageReceipt
	if setReceiptSnapshot {
		switch completionSnapshot.calendarComponentScope {
		case calendarInboundMessageReceiptSnapshotScopeSeries:
			receiptSnapshotPredicate = calendarinboundmessagereceipt.Or(
				calendarinboundmessagereceipt.And(
					calendarinboundmessagereceipt.CalendarComponentScopeIsNil(),
					calendarinboundmessagereceipt.RecurrenceIDIsNil(),
				),
				calendarinboundmessagereceipt.And(
					calendarinboundmessagereceipt.CalendarComponentScopeEQ(
						calendarinboundmessagereceipt.CalendarComponentScope(calendarInboundMessageReceiptSnapshotScopeSeries),
					),
					calendarinboundmessagereceipt.RecurrenceIDIsNil(),
				),
			)
		case calendarInboundMessageReceiptSnapshotScopeOccurrence:
			receiptSnapshotPredicate = calendarinboundmessagereceipt.Or(
				calendarinboundmessagereceipt.And(
					calendarinboundmessagereceipt.CalendarComponentScopeIsNil(),
					calendarinboundmessagereceipt.RecurrenceIDIsNil(),
				),
				calendarinboundmessagereceipt.And(
					calendarinboundmessagereceipt.CalendarComponentScopeEQ(
						calendarinboundmessagereceipt.CalendarComponentScope(calendarInboundMessageReceiptSnapshotScopeOccurrence),
					),
					calendarinboundmessagereceipt.RecurrenceIDEQ(*completionSnapshot.recurrenceID),
				),
			)
		default:
			return ErrEventInvalid
		}
		if persistedSnapshot.calendarComponentScope != "" &&
			!calendarInboundMessageReceiptSnapshotEqual(persistedSnapshot, completionSnapshot) {
			return ErrEventConflict
		}
	}

	state := calendarinboundmessagereceipt.State("")
	var errorCode *calendarinboundmessagereceipt.ErrorCode
	var due *time.Time
	var calendarEventID *uuid.UUID
	var setDeferredPair bool
	var deferredExpiresAt time.Time

	setDeadLetter := func(code string) {
		state = calendarinboundmessagereceipt.State(calendarInboundMessageReceiptStateDeadLetter)
		value := calendarinboundmessagereceipt.ErrorCode(code)
		errorCode = &value
	}
	setRetry := func() error {
		if receipt.DeferredExpiresAt != nil && !receipt.DeferredExpiresAt.After(clock) {
			setDeadLetter(calendarInboundMessageReceiptErrorMissingMasterExpired)
			return nil
		}
		if receipt.Attempts == calendarInboundMessageMaximumAttempts {
			setDeadLetter(calendarInboundMessageReceiptErrorRetryExhausted)
			return nil
		}
		nextAttemptAt, ok := calendarInboundMessageRetryDue(clock, receipt.Attempts)
		if !ok {
			return ErrEventInvalid
		}
		if receipt.DeferredExpiresAt != nil {
			nextAttemptAt, ok = calendarInboundMessageCapDueAtExpiry(nextAttemptAt, *receipt.DeferredExpiresAt)
			if !ok || !nextAttemptAt.After(clock) {
				return ErrEventInvalid
			}
		}
		state = calendarinboundmessagereceipt.State(calendarInboundMessageReceiptStatePending)
		value := calendarinboundmessagereceipt.ErrorCode(completion.errorCode)
		errorCode = &value
		due = &nextAttemptAt
		return nil
	}
	setDeferred := func() error {
		if receipt.DeferredObjectStateID == nil {
			if receipt.Attempts == calendarInboundMessageMaximumAttempts {
				setDeadLetter(calendarInboundMessageReceiptErrorMissingMasterExpired)
				return nil
			}
			deferredExpiresAt = clock.Add(calendarInboundMessageDeferredLifetime)
			if !calendarInboundMessageWholeSecondUTC(deferredExpiresAt) || !deferredExpiresAt.After(clock) {
				return ErrEventInvalid
			}
			nextAttemptAt, ok := calendarInboundMessageDeferredDue(clock, receipt.Attempts)
			if !ok {
				return ErrEventInvalid
			}
			nextAttemptAt, ok = calendarInboundMessageCapDueAtExpiry(nextAttemptAt, deferredExpiresAt)
			if !ok || !nextAttemptAt.After(clock) {
				return ErrEventInvalid
			}
			state = calendarinboundmessagereceipt.State(calendarInboundMessageReceiptStateDeferredMissingMaster)
			due = &nextAttemptAt
			setDeferredPair = true
			return nil
		}
		if *receipt.DeferredObjectStateID != completion.deferredObjectStateID {
			return ErrEventInvalid
		}
		if receipt.Attempts == calendarInboundMessageMaximumAttempts || !receipt.DeferredExpiresAt.After(clock) {
			setDeadLetter(calendarInboundMessageReceiptErrorMissingMasterExpired)
			return nil
		}
		nextAttemptAt, ok := calendarInboundMessageDeferredDue(clock, receipt.Attempts)
		if !ok {
			return ErrEventInvalid
		}
		nextAttemptAt, ok = calendarInboundMessageCapDueAtExpiry(nextAttemptAt, *receipt.DeferredExpiresAt)
		if !ok || !nextAttemptAt.After(clock) {
			return ErrEventInvalid
		}
		state = calendarinboundmessagereceipt.State(calendarInboundMessageReceiptStateDeferredMissingMaster)
		due = &nextAttemptAt
		return nil
	}

	switch completion.kind {
	case calendarInboundMessageCompletionApplied, calendarInboundMessageCompletionTombstoned:
		state = calendarinboundmessagereceipt.State(calendarInboundMessageReceiptStateApplied)
		calendarEventID = cloneCalendarInboundMessageUUIDPointer(completion.calendarEventID)
	case calendarInboundMessageCompletionObserved:
		switch completion.observedState {
		case CalendarInboundObservedIgnoredStale:
			state = calendarinboundmessagereceipt.State(calendarInboundMessageReceiptStateIgnoredStale)
		case CalendarInboundObservedDuplicate:
			state = calendarinboundmessagereceipt.State(calendarInboundMessageReceiptStateDuplicate)
		default:
			return ErrEventInvalid
		}
		calendarEventID = cloneCalendarInboundMessageUUIDPointer(completion.calendarEventID)
	case calendarInboundMessageCompletionQuarantined:
		state = calendarinboundmessagereceipt.State(calendarInboundMessageReceiptStateQuarantined)
		value := calendarinboundmessagereceipt.ErrorCode(completion.errorCode)
		errorCode = &value
	case calendarInboundMessageCompletionDeadLetter:
		setDeadLetter(string(completion.errorCode))
	case calendarInboundMessageCompletionRetry:
		if err := setRetry(); err != nil {
			return err
		}
	case calendarInboundMessageCompletionDeferred:
		if err := setDeferred(); err != nil {
			return err
		}
	default:
		return ErrEventInvalid
	}
	if state == "" {
		return ErrEventInvalid
	}

	receiptPredicates := []predicate.CalendarInboundMessageReceipt{
		calendarinboundmessagereceipt.IDEQ(claim.ReceiptID),
		calendarinboundmessagereceipt.StateEQ(calendarinboundmessagereceipt.State(calendarInboundMessageReceiptStateProcessing)),
		calendarinboundmessagereceipt.LeaseTokenEQ(claim.LeaseToken),
		calendarinboundmessagereceipt.LeaseGenerationEQ(claim.LeaseGeneration),
		calendarinboundmessagereceipt.LeaseUntilGT(clock),
	}
	if setReceiptSnapshot {
		receiptPredicates = append(receiptPredicates, receiptSnapshotPredicate)
	}
	update := transaction.CalendarInboundMessageReceipt.Update().
		Where(receiptPredicates...).
		SetState(state).
		ClearLeaseToken().
		ClearLeaseUntil().
		SetUpdatedAt(clock)
	if errorCode == nil {
		update.ClearErrorCode()
	} else {
		update.SetErrorCode(*errorCode)
	}
	if due == nil {
		update.ClearNextAttemptAt()
	} else {
		update.SetNextAttemptAt(*due)
	}
	if calendarEventID == nil {
		update.ClearCalendarEventID()
	} else {
		update.SetCalendarEventID(*calendarEventID)
	}
	if setDeferredPair {
		update.SetDeferredObjectStateID(completion.deferredObjectStateID).
			SetDeferredExpiresAt(deferredExpiresAt)
	}
	if setReceiptSnapshot {
		switch completionSnapshot.calendarComponentScope {
		case calendarInboundMessageReceiptSnapshotScopeSeries:
			update.SetCalendarComponentScope(
				calendarinboundmessagereceipt.CalendarComponentScope(calendarInboundMessageReceiptSnapshotScopeSeries),
			).ClearRecurrenceID()
		case calendarInboundMessageReceiptSnapshotScopeOccurrence:
			update.SetCalendarComponentScope(
				calendarinboundmessagereceipt.CalendarComponentScope(calendarInboundMessageReceiptSnapshotScopeOccurrence),
			).SetRecurrenceID(*completionSnapshot.recurrenceID)
		default:
			return ErrEventInvalid
		}
	}
	affected, err := update.Save(ctx)
	if err != nil {
		return calendarEntCASFailure(err)
	}
	if affected != 1 {
		return ErrEventConflict
	}
	return nil
}

func calendarInboundMessageRetryDue(clock time.Time, attempts int) (time.Time, bool) {
	var delay time.Duration
	switch attempts {
	case 1:
		delay = 30 * time.Second
	case 2:
		delay = 2 * time.Minute
	case 3:
		delay = 10 * time.Minute
	case 4:
		delay = 30 * time.Minute
	case 5:
		delay = 2 * time.Hour
	case 6:
		delay = 6 * time.Hour
	case 7:
		delay = 24 * time.Hour
	default:
		return time.Time{}, false
	}
	due := clock.Add(delay)
	return due, calendarInboundMessageWholeSecondUTC(due) && due.After(clock)
}

func calendarInboundMessageDeferredDue(clock time.Time, attempts int) (time.Time, bool) {
	var delay time.Duration
	switch attempts {
	case 1:
		delay = 15 * time.Minute
	case 2:
		delay = time.Hour
	case 3:
		delay = 4 * time.Hour
	case 4:
		delay = 12 * time.Hour
	case 5:
		delay = 24 * time.Hour
	case 6:
		delay = 48 * time.Hour
	case 7:
		delay = 72 * time.Hour
	default:
		return time.Time{}, false
	}
	due := clock.Add(delay)
	return due, calendarInboundMessageWholeSecondUTC(due) && due.After(clock)
}

func calendarInboundMessageCapDueAtExpiry(due, expiry time.Time) (time.Time, bool) {
	if !calendarInboundMessageWholeSecondUTC(due) || !calendarInboundMessageWholeSecondUTC(expiry) {
		return time.Time{}, false
	}
	if due.After(expiry) {
		due = expiry
	}
	return due, true
}

func calendarInboundMessageReceiptMatchesRegistration(
	receipt *ent.CalendarInboundMessageReceipt,
	registration CalendarInboundMessageRegistration,
) bool {
	if receipt == nil || !registration.valid() {
		return false
	}
	hash := registration.LocatorHash()
	return receipt.EventEnvelopeID == registration.EnvelopeID() &&
		receipt.OrganizationID == registration.OrganizationID() &&
		receipt.MailboxID == registration.MailboxID() &&
		receipt.MailMessageID == registration.MailMessageID() &&
		bytes.Equal(receipt.PartLocatorHash, hash[:])
}

func calendarInboundMessageRegistrationFromReceipt(
	receipt *ent.CalendarInboundMessageReceipt,
) (CalendarInboundMessageRegistration, error) {
	if receipt == nil || !calendarInboundMessageLocatorHashValid(receipt.PartLocatorHash) {
		return CalendarInboundMessageRegistration{}, ErrEventInvalid
	}
	var hash [sha256.Size]byte
	copy(hash[:], receipt.PartLocatorHash)
	registration, err := NewCalendarInboundMessageRegistration(
		receipt.EventEnvelopeID,
		receipt.OrganizationID,
		receipt.MailboxID,
		receipt.MailMessageID,
		hash,
	)
	if err != nil {
		return CalendarInboundMessageRegistration{}, ErrEventInvalid
	}
	return registration, nil
}

func calendarInboundMessageReceiptValid(receipt *ent.CalendarInboundMessageReceipt) bool {
	if receipt == nil || receipt.ID == uuid.Nil || receipt.LeaseGeneration < 0 ||
		receipt.Attempts < 0 || receipt.Attempts > calendarInboundMessageMaximumAttempts ||
		!calendarInboundMessagePersistedTimestampValid(receipt.CreatedAt) ||
		!calendarInboundMessagePersistedTimestampValid(receipt.UpdatedAt) || receipt.UpdatedAt.Before(receipt.CreatedAt) {
		return false
	}
	if _, err := calendarInboundMessageRegistrationFromReceipt(receipt); err != nil {
		return false
	}
	if receipt.CalendarEventID != nil && *receipt.CalendarEventID == uuid.Nil {
		return false
	}
	if _, ok := calendarInboundMessageReceiptSnapshotFromPersistedFields(
		receipt.CalendarComponentScope,
		receipt.RecurrenceID,
	); !ok {
		return false
	}
	if (receipt.LeaseToken == nil) != (receipt.LeaseUntil == nil) ||
		(receipt.DeferredObjectStateID == nil) != (receipt.DeferredExpiresAt == nil) ||
		!calendarInboundMessageReceiptErrorCodeValid(receipt.ErrorCode) {
		return false
	}
	if receipt.DeferredObjectStateID != nil {
		if *receipt.DeferredObjectStateID == uuid.Nil || !calendarInboundMessagePersistedWholeSecondValid(*receipt.DeferredExpiresAt) {
			return false
		}
		if receipt.NextAttemptAt != nil && receipt.NextAttemptAt.After(*receipt.DeferredExpiresAt) {
			return false
		}
	}

	switch calendarInboundMessageReceiptState(receipt) {
	case calendarInboundMessageReceiptStatePending:
		if receipt.LeaseToken != nil || receipt.LeaseUntil != nil || receipt.Attempts >= calendarInboundMessageMaximumAttempts {
			return false
		}
		if receipt.Attempts == 0 {
			return receipt.NextAttemptAt == nil && receipt.ErrorCode == nil
		}
		return receipt.NextAttemptAt != nil && receipt.ErrorCode != nil &&
			calendarInboundMessagePersistedWholeSecondValid(*receipt.NextAttemptAt) &&
			calendarInboundMessageRetryErrorCodeValid(string(*receipt.ErrorCode))
	case calendarInboundMessageReceiptStateProcessing:
		return receipt.LeaseToken != nil && *receipt.LeaseToken != uuid.Nil &&
			receipt.LeaseUntil != nil && calendarInboundMessagePersistedWholeSecondValid(*receipt.LeaseUntil) &&
			receipt.LeaseGeneration >= 1 && receipt.Attempts >= 1 &&
			receipt.NextAttemptAt == nil && receipt.ErrorCode == nil
	case calendarInboundMessageReceiptStateDeferredMissingMaster:
		return receipt.LeaseToken == nil && receipt.LeaseUntil == nil &&
			receipt.Attempts >= 1 && receipt.Attempts < calendarInboundMessageMaximumAttempts &&
			receipt.NextAttemptAt != nil && calendarInboundMessagePersistedWholeSecondValid(*receipt.NextAttemptAt) &&
			receipt.DeferredObjectStateID != nil && receipt.DeferredExpiresAt != nil && receipt.ErrorCode == nil
	case calendarInboundMessageReceiptStateApplied,
		calendarInboundMessageReceiptStateIgnoredStale,
		calendarInboundMessageReceiptStateDuplicate:
		return receipt.LeaseToken == nil && receipt.LeaseUntil == nil &&
			receipt.NextAttemptAt == nil && receipt.ErrorCode == nil
	case calendarInboundMessageReceiptStateQuarantined:
		return receipt.LeaseToken == nil && receipt.LeaseUntil == nil &&
			receipt.NextAttemptAt == nil && receipt.ErrorCode != nil &&
			calendarInboundMessageQuarantineErrorCodeValid(string(*receipt.ErrorCode))
	case calendarInboundMessageReceiptStateDeadLetter:
		return receipt.LeaseToken == nil && receipt.LeaseUntil == nil &&
			receipt.NextAttemptAt == nil && receipt.ErrorCode != nil &&
			calendarInboundMessageDeadLetterErrorCodeValid(string(*receipt.ErrorCode))
	default:
		return false
	}
}

func calendarInboundMessageReceiptState(receipt *ent.CalendarInboundMessageReceipt) string {
	if receipt == nil {
		return ""
	}
	return string(receipt.State)
}

func calendarInboundMessageReceiptFinalized(receipt *ent.CalendarInboundMessageReceipt) bool {
	switch calendarInboundMessageReceiptState(receipt) {
	case calendarInboundMessageReceiptStateApplied,
		calendarInboundMessageReceiptStateIgnoredStale,
		calendarInboundMessageReceiptStateDuplicate,
		calendarInboundMessageReceiptStateQuarantined,
		calendarInboundMessageReceiptStateDeadLetter:
		return true
	default:
		return false
	}
}

func calendarInboundMessageFinalizedStates() []calendarinboundmessagereceipt.State {
	return []calendarinboundmessagereceipt.State{
		calendarinboundmessagereceipt.State(calendarInboundMessageReceiptStateApplied),
		calendarinboundmessagereceipt.State(calendarInboundMessageReceiptStateIgnoredStale),
		calendarinboundmessagereceipt.State(calendarInboundMessageReceiptStateDuplicate),
		calendarinboundmessagereceipt.State(calendarInboundMessageReceiptStateQuarantined),
		calendarinboundmessagereceipt.State(calendarInboundMessageReceiptStateDeadLetter),
	}
}

func calendarInboundMessageLocatorHashValid(value []byte) bool {
	if len(value) != sha256.Size {
		return false
	}
	var zero [sha256.Size]byte
	return !bytes.Equal(value, zero[:])
}

func calendarInboundMessageTimestampValid(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC
}

func calendarInboundMessageWholeSecondUTC(value time.Time) bool {
	return calendarInboundMessageTimestampValid(value) && value.Nanosecond() == 0
}

func calendarInboundMessagePersistedTimestampValid(value time.Time) bool {
	return !value.IsZero()
}

func calendarInboundMessagePersistedWholeSecondValid(value time.Time) bool {
	return calendarInboundMessagePersistedTimestampValid(value) && value.Nanosecond() == 0
}

func calendarInboundMessageReceiptErrorCodeValid(code *calendarinboundmessagereceipt.ErrorCode) bool {
	if code == nil {
		return true
	}
	switch string(*code) {
	case "message_unavailable",
		"part_unavailable",
		"malformed_itip",
		"unsupported_method",
		"unsupported_version",
		"untrusted_sender",
		"organizer_mismatch",
		"equal_sequence_conflict",
		calendarInboundMessageReceiptErrorMissingMasterExpired,
		"resource_limit",
		calendarInboundMessageReceiptErrorRetryExhausted,
		"engine_unavailable",
		"database_unavailable",
		"internal_error":
		return true
	default:
		return false
	}
}

func calendarInboundMessageQuarantineErrorCodeValid(code string) bool {
	switch code {
	case "malformed_itip",
		"unsupported_method",
		"unsupported_version",
		"untrusted_sender",
		"organizer_mismatch",
		"equal_sequence_conflict",
		"resource_limit",
		"internal_error":
		return true
	default:
		return false
	}
}

func calendarInboundMessageDeadLetterErrorCodeValid(code string) bool {
	switch code {
	case "message_unavailable",
		"part_unavailable",
		calendarInboundMessageReceiptErrorMissingMasterExpired,
		"resource_limit",
		calendarInboundMessageReceiptErrorRetryExhausted,
		"internal_error":
		return true
	default:
		return false
	}
}

func calendarInboundMessageRetryErrorCodeValid(code string) bool {
	switch code {
	case "engine_unavailable", "database_unavailable", "internal_error":
		return true
	default:
		return false
	}
}

func calendarInboundMessageDueScanSkipLocked() predicate.CalendarInboundMessageReceipt {
	return func(selector *entsql.Selector) {
		selector.ForUpdate(entsql.WithLockAction(entsql.SkipLocked))
		dueColumn := "COALESCE(" +
			selector.C(calendarinboundmessagereceipt.FieldNextAttemptAt) + ", " +
			selector.C(calendarinboundmessagereceipt.FieldLeaseUntil) + ")"
		selector.OrderBy(dueColumn+" ASC", selector.C(calendarinboundmessagereceipt.FieldID)+" ASC")
	}
}

func calendarInboundMessageCleanupSkipLocked() predicate.CalendarInboundMessageReceipt {
	return func(selector *entsql.Selector) {
		selector.ForUpdate(entsql.WithLockAction(entsql.SkipLocked))
	}
}
