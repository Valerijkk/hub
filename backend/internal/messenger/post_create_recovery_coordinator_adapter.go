package messenger

import (
	"bytes"
	"context"
	"strings"
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
	) (Phase151RecoveryTargetAuthority, error)
	ListPendingPostCreateDocuments(
		context.Context,
		Phase151RecoveryTargetAuthority,
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
	afterCopy, err := copyPhase151RecoveryTargetCursor(after)
	if err != nil {
		return Phase151PendingRecoveryTargetPage{}, err
	}
	page, err := a.scanner.ListPendingRecoveryTargets(ctx, afterCopy, limit)
	if err != nil {
		return Phase151PendingRecoveryTargetPage{}, err
	}
	return validatePhase151PendingRecoveryTargetPage(afterCopy, limit, page)
}

func validatePhase151PendingRecoveryTargetPage(
	after *Phase151RecoveryTargetCursor,
	limit int,
	page Phase151PendingRecoveryTargetPage,
) (Phase151PendingRecoveryTargetPage, error) {
	if limit <= 0 || len(page.Targets) > limit {
		return Phase151PendingRecoveryTargetPage{}, ErrPostCreatePipelineInvalid
	}
	afterCopy, err := copyPhase151RecoveryTargetCursor(after)
	if err != nil {
		return Phase151PendingRecoveryTargetPage{}, err
	}

	ownedTargets := append([]Phase151PendingRecoveryTarget(nil), page.Targets...)
	for index, target := range ownedTargets {
		if err := validatePhase151PendingRecoveryTarget(target); err != nil {
			return Phase151PendingRecoveryTargetPage{}, err
		}
		if afterCopy != nil && comparePhase151PendingRecoveryTargetToCursor(target, *afterCopy) <= 0 {
			return Phase151PendingRecoveryTargetPage{}, ErrPostCreatePipelineInvalid
		}
		if index > 0 && comparePhase151PendingRecoveryTargets(ownedTargets[index-1], target) >= 0 {
			return Phase151PendingRecoveryTargetPage{}, ErrPostCreatePipelineInvalid
		}
	}

	var nextCopy *Phase151RecoveryTargetCursor
	if page.Next != nil {
		if len(ownedTargets) == 0 {
			return Phase151PendingRecoveryTargetPage{}, ErrPostCreatePipelineInvalid
		}
		nextCopy, err = copyPhase151RecoveryTargetCursor(page.Next)
		if err != nil {
			return Phase151PendingRecoveryTargetPage{}, err
		}
		if comparePhase151PendingRecoveryTargetToCursor(ownedTargets[len(ownedTargets)-1], *nextCopy) != 0 {
			return Phase151PendingRecoveryTargetPage{}, ErrPostCreatePipelineInvalid
		}
	}

	return Phase151PendingRecoveryTargetPage{Targets: ownedTargets, Next: nextCopy}, nil
}

func copyPhase151RecoveryTargetCursor(
	cursor *Phase151RecoveryTargetCursor,
) (*Phase151RecoveryTargetCursor, error) {
	if cursor == nil {
		return nil, nil
	}
	if err := validatePhase151PendingRecoveryTarget(Phase151PendingRecoveryTarget{
		TenantID: cursor.TenantID, ChannelID: cursor.ChannelID, Kind: cursor.Kind,
	}); err != nil {
		return nil, err
	}
	copy := *cursor
	return &copy, nil
}

func comparePhase151PendingRecoveryTargetToCursor(
	target Phase151PendingRecoveryTarget,
	cursor Phase151RecoveryTargetCursor,
) int {
	return comparePhase151PendingRecoveryTargets(target, Phase151PendingRecoveryTarget{
		TenantID: cursor.TenantID, ChannelID: cursor.ChannelID, Kind: cursor.Kind,
	})
}

func comparePhase151PendingRecoveryTargets(left, right Phase151PendingRecoveryTarget) int {
	if comparison := bytes.Compare(left.TenantID[:], right.TenantID[:]); comparison != 0 {
		return comparison
	}
	if comparison := bytes.Compare(left.ChannelID[:], right.ChannelID[:]); comparison != 0 {
		return comparison
	}
	return strings.Compare(left.Kind, right.Kind)
}

func (a *Phase151ClaimedRecoveryCoordinatorAdapter) ReauthorizePendingRecoveryTarget(
	ctx context.Context,
	target Phase151PendingRecoveryTarget,
) (Phase151RecoveryTargetAuthority, error) {
	if err := a.validate(ctx); err != nil {
		return Phase151RecoveryTargetAuthority{}, err
	}
	if err := validatePhase151PendingRecoveryTarget(target); err != nil {
		return Phase151RecoveryTargetAuthority{}, err
	}
	authority, err := a.scanner.ReauthorizePendingRecoveryTarget(ctx, target)
	if err != nil {
		return Phase151RecoveryTargetAuthority{}, err
	}
	if err := ExportPhase151ValidateRecoveryTargetAuthority(Phase151PostCreateRecoveryServicePrincipal, target, authority); err != nil {
		return Phase151RecoveryTargetAuthority{}, err
	}
	return authority, nil
}

func (a *Phase151ClaimedRecoveryCoordinatorAdapter) ListPendingPostCreateDocuments(
	ctx context.Context,
	authority Phase151RecoveryTargetAuthority,
	limit int,
) ([]MessageDocument, error) {
	if err := a.validate(ctx); err != nil {
		return nil, err
	}
	target := Phase151PendingRecoveryTarget{
		TenantID: authority.TenantID, ChannelID: authority.ChannelID, Kind: authority.Kind,
	}
	if limit <= 0 {
		return nil, ErrPostCreatePipelineInvalid
	}
	if err := ExportPhase151ValidateRecoveryTargetAuthority(Phase151PostCreateRecoveryServicePrincipal, target, authority); err != nil {
		return nil, err
	}
	return a.scanner.ListPendingPostCreateDocuments(ctx, authority, limit)
}

func (a *Phase151ClaimedRecoveryCoordinatorAdapter) ReconcilePendingPostCreateEffects(
	ctx context.Context,
	authority Phase151RecoveryTargetAuthority,
	document MessageDocument,
) error {
	if err := a.validate(ctx); err != nil {
		return err
	}
	target := Phase151PendingRecoveryTarget{
		TenantID: authority.TenantID, ChannelID: authority.ChannelID, Kind: authority.Kind,
	}
	if err := ExportPhase151ValidateRecoveryTargetAuthority(Phase151PostCreateRecoveryServicePrincipal, target, authority); err != nil {
		return err
	}
	_, err := ExportPhase151ExecuteClaimedPostCreateRecovery(
		ctx,
		authority,
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
