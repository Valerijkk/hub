package mail

import (
	"context"
	"crypto/sha256"
	"errors"

	"github.com/google/uuid"
)

// MailboxMessageMutationServiceAPI is deliberately transport-neutral. The HTTP
// packet owns canonical textual UUID parsing and status mapping later.
type MailboxMessageMutationServiceAPI interface {
	MutateMailboxMessages(context.Context, uuid.UUID, MailboxMessageMutationInput) (MailboxMessageMutationResult, error)
}

// MailboxMessageMutationService coordinates the receipt/fence lifecycle around
// exactly one bounded provider batch. It never writes a local message projection.
type MailboxMessageMutationService struct {
	owners     mailboxOwnerResolver
	repository MailboxMessageMutationRepository
	engine     MailboxMessageMutationEngine
}

func NewMailboxMessageMutationService(
	owners mailboxOwnerResolver,
	repository MailboxMessageMutationRepository,
	engine MailboxMessageMutationEngine,
) (*MailboxMessageMutationService, error) {
	if mailboxMessageMutationDependencyNil(owners) ||
		mailboxMessageMutationDependencyNil(repository) ||
		mailboxMessageMutationDependencyNil(engine) {
		return nil, mailboxMessageMutationError(MailboxMessageMutationErrorInvalid)
	}
	return &MailboxMessageMutationService{owners: owners, repository: repository, engine: engine}, nil
}

// MutateMailboxMessages returns only the persisted receipt result. The repository
// owns current-projection reconciliation; a completed provider call is not treated
// as a synthetic local projection update.
func (service *MailboxMessageMutationService) MutateMailboxMessages(
	ctx context.Context,
	actorID uuid.UUID,
	input MailboxMessageMutationInput,
) (MailboxMessageMutationResult, error) {
	if service == nil || mailboxMessageMutationDependencyNil(service.owners) ||
		mailboxMessageMutationDependencyNil(service.repository) || mailboxMessageMutationDependencyNil(service.engine) ||
		ctx == nil || actorID == uuid.Nil || !mailboxMessageMutationInputValid(input) {
		return MailboxMessageMutationResult{}, mailboxMessageMutationError(MailboxMessageMutationErrorInvalid)
	}
	if ctx.Err() != nil {
		return MailboxMessageMutationResult{}, mailboxMessageMutationError(MailboxMessageMutationErrorUnavailable)
	}
	input = mailboxMessageMutationInputCopy(input)

	owner, err := service.owners.ResolveOwner(ctx, actorID)
	if err != nil {
		return MailboxMessageMutationResult{}, mailboxMessageMutationOwnerError(err)
	}
	scope, err := mailboxMessageMutationScopeFromOwner(actorID, owner)
	if err != nil {
		return MailboxMessageMutationResult{}, err
	}
	if !mailboxMessageMutationScopeValid(scope) {
		return MailboxMessageMutationResult{}, mailboxMessageMutationError(MailboxMessageMutationErrorUnavailable)
	}

	idempotencyKeyHash := sha256.Sum256(input.IdempotencyKey[:])
	requestHash := mailboxMessageMutationRequestHash(input, scope)
	replayRequest := MailboxMessageMutationReplayRequest{
		Scope:              scope,
		IdempotencyKeyHash: idempotencyKeyHash,
		RequestHash:        requestHash,
		FingerprintVersion: mailboxMessageMutationFingerprintVersion,
	}
	replay, err := service.repository.LookupMailboxMessageMutationReplay(ctx, replayRequest)
	if err != nil {
		return MailboxMessageMutationResult{}, mailboxMessageMutationRepositoryServiceError(err)
	}
	if !mailboxMessageMutationReplayValid(replay, input) {
		return MailboxMessageMutationResult{}, mailboxMessageMutationError(MailboxMessageMutationErrorUnavailable)
	}
	if result, handled, err := mailboxMessageMutationReplayResult(replay, input); handled {
		return result, err
	}

	resolution, err := service.repository.ResolveMailboxMessageMutationTargets(ctx, MailboxMessageMutationResolveRequest{
		Scope: scope,
		Input: input,
	})
	if err != nil {
		return MailboxMessageMutationResult{}, mailboxMessageMutationRepositoryServiceError(err)
	}
	if !mailboxMessageMutationResolutionValid(resolution, input) {
		return MailboxMessageMutationResult{}, mailboxMessageMutationError(MailboxMessageMutationErrorUnavailable)
	}

	reserveRequest := MailboxMessageMutationReserveRequest{
		Scope:              scope,
		Input:              input,
		IdempotencyKeyHash: idempotencyKeyHash,
		RequestHash:        requestHash,
		FingerprintVersion: mailboxMessageMutationFingerprintVersion,
		Resolution:         mailboxMessageMutationResolutionCopy(resolution),
	}
	reservation, err := service.repository.ReserveMailboxMessageMutation(ctx, reserveRequest)
	if err != nil {
		return MailboxMessageMutationResult{}, mailboxMessageMutationRepositoryServiceError(err)
	}
	if !mailboxMessageMutationReservationValid(reservation, reserveRequest) {
		return MailboxMessageMutationResult{}, mailboxMessageMutationError(MailboxMessageMutationErrorUnavailable)
	}
	if result, handled, err := mailboxMessageMutationReservationResult(reservation, input); handled {
		return result, err
	}

	fence := *reservation.Fence
	snapshot := mailboxMessageMutationSnapshotCopy(*reservation.Snapshot)
	if err := mailboxMessageMutationDispatch(ctx, service.engine, reservation.EngineAccountID, input, snapshot); err != nil {
		if !mailboxMessageMutationEngineRejected(err) {
			return service.markUnknownMailboxMessageMutation(ctx, scope, reservation.ReceiptID, fence, input)
		}
		return service.finishMailboxMessageMutation(ctx, scope, reservation.ReceiptID, fence, MailboxMessageMutationFinishRejected, input)
	}
	return service.finishMailboxMessageMutation(ctx, scope, reservation.ReceiptID, fence, MailboxMessageMutationFinishApplied, input)
}

func mailboxMessageMutationReplayResult(replay MailboxMessageMutationReplay, input MailboxMessageMutationInput) (MailboxMessageMutationResult, bool, error) {
	switch replay.Kind {
	case MailboxMessageMutationReplayAbsent:
		return MailboxMessageMutationResult{}, false, nil
	case MailboxMessageMutationReplayCompleted, MailboxMessageMutationReplayUnknown:
		if replay.Result == nil || !mailboxMessageMutationResultValid(*replay.Result, input) {
			return MailboxMessageMutationResult{}, true, mailboxMessageMutationError(MailboxMessageMutationErrorUnavailable)
		}
		return mailboxMessageMutationResultCopy(*replay.Result), true, nil
	case MailboxMessageMutationReplayExpansionLimit:
		return MailboxMessageMutationResult{}, true, mailboxMessageMutationError(MailboxMessageMutationErrorExpansionLimit)
	case MailboxMessageMutationReplayInProgress:
		return MailboxMessageMutationResult{}, true, mailboxMessageMutationError(MailboxMessageMutationErrorBusy)
	default:
		return MailboxMessageMutationResult{}, true, mailboxMessageMutationError(MailboxMessageMutationErrorUnavailable)
	}
}

func mailboxMessageMutationReservationResult(reservation MailboxMessageMutationReservation, input MailboxMessageMutationInput) (MailboxMessageMutationResult, bool, error) {
	switch reservation.Kind {
	case MailboxMessageMutationReservationReserved:
		return MailboxMessageMutationResult{}, false, nil
	case MailboxMessageMutationReservationCompleted, MailboxMessageMutationReservationUnknown:
		if reservation.Result == nil || !mailboxMessageMutationResultValid(*reservation.Result, input) {
			return MailboxMessageMutationResult{}, true, mailboxMessageMutationError(MailboxMessageMutationErrorUnavailable)
		}
		return mailboxMessageMutationResultCopy(*reservation.Result), true, nil
	case MailboxMessageMutationReservationExpansionLimit:
		return MailboxMessageMutationResult{}, true, mailboxMessageMutationError(MailboxMessageMutationErrorExpansionLimit)
	case MailboxMessageMutationReservationBusy:
		return MailboxMessageMutationResult{}, true, mailboxMessageMutationError(MailboxMessageMutationErrorBusy)
	default:
		return MailboxMessageMutationResult{}, true, mailboxMessageMutationError(MailboxMessageMutationErrorUnavailable)
	}
}

func (service *MailboxMessageMutationService) finishMailboxMessageMutation(
	ctx context.Context,
	scope MailboxMessageMutationScope,
	receiptID uuid.UUID,
	fence MailboxEngineCommandFence,
	outcome MailboxMessageMutationFinishOutcome,
	input MailboxMessageMutationInput,
) (MailboxMessageMutationResult, error) {
	result, err := service.repository.FinishMailboxMessageMutation(ctx, MailboxMessageMutationFinishRequest{
		Scope: scope, ReceiptID: receiptID, Fence: fence, Outcome: outcome,
	})
	if err == nil && result.State == MailboxMessageMutationResultCompleted && mailboxMessageMutationResultValid(result, input) {
		return mailboxMessageMutationResultCopy(result), nil
	}
	// Provider I/O has already occurred. Seal the durable receipt as unknown
	// before returning; repository reservations are additionally fail-closed for
	// any in-progress receipt if this exact transition cannot be persisted.
	return service.markUnknownMailboxMessageMutation(ctx, scope, receiptID, fence, input)
}

func (service *MailboxMessageMutationService) markUnknownMailboxMessageMutation(
	ctx context.Context,
	scope MailboxMessageMutationScope,
	receiptID uuid.UUID,
	fence MailboxEngineCommandFence,
	input MailboxMessageMutationInput,
) (MailboxMessageMutationResult, error) {
	result, err := service.repository.MarkMailboxMessageMutationUnknown(ctx, MailboxMessageMutationUnknownRequest{
		Scope: scope, ReceiptID: receiptID, Fence: fence,
	})
	if err != nil || result.State != MailboxMessageMutationResultUnknown || !mailboxMessageMutationResultValid(result, input) {
		return MailboxMessageMutationResult{}, mailboxMessageMutationError(MailboxMessageMutationErrorOutcomeUnknown)
	}
	return mailboxMessageMutationResultCopy(result), nil
}

func mailboxMessageMutationDispatch(
	ctx context.Context,
	engine MailboxMessageMutationEngine,
	engineAccountID string,
	input MailboxMessageMutationInput,
	snapshot MailboxMessageMutationSnapshot,
) error {
	if mailboxMessageMutationDependencyNil(engine) || ctx == nil || ctx.Err() != nil || !syncValidOpaque(engineAccountID, MaxEngineOpaqueIDBytes) ||
		!mailboxMessageMutationInputValid(input) || !mailboxMessageMutationSnapshotValid(snapshot, input) {
		return &EngineError{Kind: EngineErrorIndeterminate}
	}
	messageIDs := mailboxMessageMutationEngineMessageIDs(snapshot, input)
	if len(messageIDs) < 1 || len(messageIDs) > MailboxMessageMutationMaximumResolvedMessages || !validOpaqueIDs(messageIDs) {
		return &EngineError{Kind: EngineErrorIndeterminate}
	}

	switch input.Action {
	case MailboxMessageMutationActionMove:
		return engine.MoveMessages(ctx, engineAccountID, messageIDs, snapshot.EngineFolderID)
	case MailboxMessageMutationActionMarkRead:
		return engine.SetFlags(ctx, engineAccountID, messageIDs, FlagPatch{Seen: FlagValue{Present: true, Value: true}})
	case MailboxMessageMutationActionMarkUnread:
		return engine.SetFlags(ctx, engineAccountID, messageIDs, FlagPatch{Seen: FlagValue{Present: true, Value: false}})
	case MailboxMessageMutationActionFlag:
		return engine.SetFlags(ctx, engineAccountID, messageIDs, FlagPatch{Flagged: FlagValue{Present: true, Value: true}})
	case MailboxMessageMutationActionUnflag:
		return engine.SetFlags(ctx, engineAccountID, messageIDs, FlagPatch{Flagged: FlagValue{Present: true, Value: false}})
	case MailboxMessageMutationActionTrash:
		return engine.DeleteMessages(ctx, engineAccountID, messageIDs, false)
	default:
		return &EngineError{Kind: EngineErrorRejectedPolicy}
	}
}

func mailboxMessageMutationEngineRejected(err error) bool {
	return errors.Is(err, ErrEngineNotFound) || errors.Is(err, ErrEngineConflict) ||
		errors.Is(err, ErrEngineInvalidCursor) || errors.Is(err, ErrEngineRejectedPolicy) ||
		errors.Is(err, ErrEngineUnauthorizedAccount) || errors.Is(err, ErrEnginePayloadTooLarge)
}

func mailboxMessageMutationScopeFromOwner(actorID uuid.UUID, owner MailboxOwner) (MailboxMessageMutationScope, error) {
	readScope, err := mailboxMessageReadScopeFromOwner(actorID, owner)
	if err != nil {
		switch kind, ok := mailboxMessageReadErrorKindOf(err); {
		case !ok:
			return MailboxMessageMutationScope{}, mailboxMessageMutationError(MailboxMessageMutationErrorUnavailable)
		case kind == MailboxMessageReadErrorNotFound:
			return MailboxMessageMutationScope{}, mailboxMessageMutationError(MailboxMessageMutationErrorNotFound)
		case kind == MailboxMessageReadErrorDeleting:
			return MailboxMessageMutationScope{}, mailboxMessageMutationError(MailboxMessageMutationErrorDeleting)
		case kind == MailboxMessageReadErrorUnavailable:
			return MailboxMessageMutationScope{}, mailboxMessageMutationError(MailboxMessageMutationErrorUnavailable)
		default:
			return MailboxMessageMutationScope{}, mailboxMessageMutationError(MailboxMessageMutationErrorInvalid)
		}
	}
	scope := MailboxMessageMutationScope{
		ActorID: readScope.ActorID, OrganizationID: readScope.OrganizationID,
		AccountID: readScope.AccountID, EngineAccountID: readScope.EngineAccountID,
	}
	if !mailboxMessageMutationScopeValid(scope) {
		return MailboxMessageMutationScope{}, mailboxMessageMutationError(MailboxMessageMutationErrorUnavailable)
	}
	return scope, nil
}

func mailboxMessageMutationOwnerError(err error) error {
	var ownerError *MailboxServiceError
	if !errors.As(err, &ownerError) || ownerError == nil {
		return mailboxMessageMutationError(MailboxMessageMutationErrorUnavailable)
	}
	switch ownerError.Kind {
	case MailboxServiceErrorNotFound:
		return mailboxMessageMutationError(MailboxMessageMutationErrorNotFound)
	case MailboxServiceErrorDeleting:
		return mailboxMessageMutationError(MailboxMessageMutationErrorDeleting)
	case MailboxServiceErrorUnavailable:
		return mailboxMessageMutationError(MailboxMessageMutationErrorUnavailable)
	default:
		return mailboxMessageMutationError(MailboxMessageMutationErrorInvalid)
	}
}

func mailboxMessageMutationRepositoryServiceError(err error) error {
	kind, ok := MailboxMessageMutationRepositoryErrorKindOf(err)
	if !ok {
		return mailboxMessageMutationError(MailboxMessageMutationErrorUnavailable)
	}
	switch kind {
	case MailboxMessageMutationRepositoryErrorNotFound:
		return mailboxMessageMutationError(MailboxMessageMutationErrorNotFound)
	case MailboxMessageMutationRepositoryErrorAccountDeleting:
		return mailboxMessageMutationError(MailboxMessageMutationErrorDeleting)
	case MailboxMessageMutationRepositoryErrorStateConflict:
		return mailboxMessageMutationError(MailboxMessageMutationErrorBusy)
	case MailboxMessageMutationRepositoryErrorIdempotencyConflict:
		return mailboxMessageMutationError(MailboxMessageMutationErrorIdempotencyConflict)
	case MailboxMessageMutationRepositoryErrorFenceLost:
		return mailboxMessageMutationError(MailboxMessageMutationErrorOutcomeUnknown)
	case MailboxMessageMutationRepositoryErrorUnavailable:
		return mailboxMessageMutationError(MailboxMessageMutationErrorUnavailable)
	default:
		return mailboxMessageMutationError(MailboxMessageMutationErrorInvalid)
	}
}
