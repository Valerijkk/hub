package messenger

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Phase151PendingRecoveryScanRepository interface {
	TryAcquirePendingRecoveryInterval(context.Context, time.Duration) (bool, error)
	ListPendingRecoveryTargets(
		context.Context,
		*Phase151RecoveryTargetCursor,
		int,
	) (Phase151PendingRecoveryTargetPage, error)
	ReauthorizePendingRecoveryTarget(
		context.Context,
		Phase151PendingRecoveryTarget,
	) (AuthorizedMessageProducerScope, error)
	ListPendingPostCreateDocuments(
		context.Context,
		AuthorizedMessageProducerScope,
		Phase151PendingRecoveryTarget,
		int,
	) ([]MessageDocument, error)
}

type Phase151ClaimedRecoveryCoordinatorAdapter struct {
	scanner       Phase151PendingRecoveryScanRepository
	ownerID       uuid.UUID
	leaseDuration time.Duration
	clock         Phase151PostCreateRecoveryClock
	repository    Phase151ClaimedPostCreateRecoveryRepository
}

var _ Phase151PendingRecoveryRepository = (*Phase151ClaimedRecoveryCoordinatorAdapter)(nil)

func NewPhase151ClaimedRecoveryCoordinatorAdapter(
	scanner Phase151PendingRecoveryScanRepository,
	ownerID uuid.UUID,
	leaseDuration time.Duration,
	clock Phase151PostCreateRecoveryClock,
	repository Phase151ClaimedPostCreateRecoveryRepository,
) *Phase151ClaimedRecoveryCoordinatorAdapter {
	if phase151PostCreateDependencyNil(scanner) || ownerID == uuid.Nil || leaseDuration <= 0 ||
		phase151PostCreateDependencyNil(clock) || phase151PostCreateDependencyNil(repository) {
		return nil
	}
	return &Phase151ClaimedRecoveryCoordinatorAdapter{
		scanner: scanner, ownerID: ownerID, leaseDuration: leaseDuration,
		clock: clock, repository: repository,
	}
}

func (a *Phase151ClaimedRecoveryCoordinatorAdapter) TryAcquirePendingRecoveryInterval(
	ctx context.Context,
	minimum time.Duration,
) (bool, error) {
	if err := a.validate(ctx); err != nil || minimum <= 0 {
		if err != nil {
			return false, err
		}
		return false, ErrPostCreatePipelineInvalid
	}
	return a.scanner.TryAcquirePendingRecoveryInterval(ctx, minimum)
}

func (a *Phase151ClaimedRecoveryCoordinatorAdapter) ListPendingRecoveryTargets(
	ctx context.Context,
	after *Phase151RecoveryTargetCursor,
	limit int,
) (Phase151PendingRecoveryTargetPage, error) {
	if err := a.validate(ctx); err != nil || limit <= 0 {
		if err != nil {
			return Phase151PendingRecoveryTargetPage{}, err
		}
		return Phase151PendingRecoveryTargetPage{}, ErrPostCreatePipelineInvalid
	}
	return a.scanner.ListPendingRecoveryTargets(ctx, after, limit)
}

func (a *Phase151ClaimedRecoveryCoordinatorAdapter) ReauthorizePendingRecoveryTarget(
	ctx context.Context,
	target Phase151PendingRecoveryTarget,
) (AuthorizedMessageProducerScope, error) {
	if err := a.validate(ctx); err != nil {
		return AuthorizedMessageProducerScope{}, err
	}
	if err := validatePhase151PendingRecoveryTarget(target); err != nil {
		return AuthorizedMessageProducerScope{}, err
	}
	scope, err := a.scanner.ReauthorizePendingRecoveryTarget(ctx, target)
	if err != nil {
		return AuthorizedMessageProducerScope{}, err
	}
	if scope.TenantID != target.TenantID || scope.ChannelID != target.ChannelID ||
		scope.ActorID == uuid.Nil {
		return AuthorizedMessageProducerScope{}, ErrPostCreatePipelineInvalid
	}
	if err := ExportPhase151ValidateTrustedMessageProducer(scope, scope.Producer); err != nil {
		return AuthorizedMessageProducerScope{}, err
	}
	return scope, nil
}

func (a *Phase151ClaimedRecoveryCoordinatorAdapter) ListPendingPostCreateDocuments(
	ctx context.Context,
	scope AuthorizedMessageProducerScope,
	target Phase151PendingRecoveryTarget,
	limit int,
) ([]MessageDocument, error) {
	if err := a.validate(ctx); err != nil {
		return nil, err
	}
	if limit <= 0 || validatePhase151PendingRecoveryTarget(target) != nil ||
		scope.TenantID != target.TenantID || scope.ChannelID != target.ChannelID ||
		scope.ActorID == uuid.Nil {
		return nil, ErrPostCreatePipelineInvalid
	}
	if err := ExportPhase151ValidateTrustedMessageProducer(scope, scope.Producer); err != nil {
		return nil, err
	}
	return a.scanner.ListPendingPostCreateDocuments(ctx, scope, target, limit)
}

func (a *Phase151ClaimedRecoveryCoordinatorAdapter) ReconcilePendingPostCreateEffects(
	ctx context.Context,
	targetAuthority AuthorizedMessageProducerScope,
	document MessageDocument,
) error {
	if err := a.validate(ctx); err != nil {
		return err
	}
	_, err := ExportPhase151ExecuteClaimedPostCreateRecovery(
		ctx,
		targetAuthority,
		&document,
		a.ownerID,
		a.leaseDuration,
		a.clock,
		a.repository,
	)
	return err
}

func (a *Phase151ClaimedRecoveryCoordinatorAdapter) validate(ctx context.Context) error {
	if a == nil || ctx == nil || phase151PostCreateDependencyNil(a.scanner) ||
		a.ownerID == uuid.Nil || a.leaseDuration <= 0 ||
		phase151PostCreateDependencyNil(a.clock) ||
		phase151PostCreateDependencyNil(a.repository) {
		return ErrPostCreatePipelineInvalid
	}
	return ctx.Err()
}

func validatePhase151PendingRecoveryTarget(target Phase151PendingRecoveryTarget) error {
	if target.TenantID == uuid.Nil || target.ChannelID == uuid.Nil || target.Kind == "" {
		return ErrPostCreatePipelineInvalid
	}
	return nil
}
