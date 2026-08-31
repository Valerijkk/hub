package messenger

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Phase151PendingRecoveryLeaderLease struct {
	HolderID   uuid.UUID
	Fence      int64
	LeaseUntil time.Time
}

type Phase151PendingRecoveryLeaderCheckpoint struct {
	Cursor *Phase151RecoveryTargetCursor
}

type Phase151PendingRecoveryLeaderLeaseAcquisition struct {
	Lease      Phase151PendingRecoveryLeaderLease
	Checkpoint Phase151PendingRecoveryLeaderCheckpoint
}

type Phase151PendingRecoveryLeaderLeaseRepository interface {
	AcquirePendingRecoveryLeaderLease(context.Context, uuid.UUID, time.Duration) (Phase151PendingRecoveryLeaderLeaseAcquisition, bool, error)
	RenewPendingRecoveryLeaderLease(context.Context, Phase151PendingRecoveryLeaderLease, time.Duration) (Phase151PendingRecoveryLeaderLease, bool, error)
	ReleasePendingRecoveryLeaderLease(context.Context, Phase151PendingRecoveryLeaderLease) (bool, error)
	AdvancePendingRecoveryLeaderCheckpoint(context.Context, Phase151PendingRecoveryLeaderLease, Phase151PendingRecoveryLeaderCheckpoint) (bool, error)
}

func ExportPhase151AcquirePendingRecoveryLeaderLease(
	ctx context.Context,
	repository Phase151PendingRecoveryLeaderLeaseRepository,
	holderID uuid.UUID,
	duration time.Duration,
) (Phase151PendingRecoveryLeaderLeaseAcquisition, bool, error) {
	if phase151PendingRecoveryLeaderLeaseDependenciesInvalid(ctx, repository) {
		return Phase151PendingRecoveryLeaderLeaseAcquisition{}, false, ErrPostCreatePipelineInvalid
	}
	if err := ctx.Err(); err != nil {
		return Phase151PendingRecoveryLeaderLeaseAcquisition{}, false, err
	}
	if holderID == uuid.Nil || duration <= 0 {
		return Phase151PendingRecoveryLeaderLeaseAcquisition{}, false, ErrPostCreatePipelineInvalid
	}

	acquisition, owned, err := repository.AcquirePendingRecoveryLeaderLease(ctx, holderID, duration)
	if ctxErr := ctx.Err(); ctxErr != nil {
		return Phase151PendingRecoveryLeaderLeaseAcquisition{}, false, ctxErr
	}
	if err != nil {
		return Phase151PendingRecoveryLeaderLeaseAcquisition{}, false, err
	}
	if !owned {
		return Phase151PendingRecoveryLeaderLeaseAcquisition{}, false, nil
	}
	if !phase151PendingRecoveryLeaderLeaseValid(acquisition.Lease) || acquisition.Lease.HolderID != holderID ||
		!phase151PendingRecoveryLeaderCheckpointValid(acquisition.Checkpoint, true) {
		return Phase151PendingRecoveryLeaderLeaseAcquisition{}, false, ErrPostCreatePipelineInvalid
	}

	return phase151CopyPendingRecoveryLeaderLeaseAcquisition(acquisition), true, nil
}

func ExportPhase151RenewPendingRecoveryLeaderLease(
	ctx context.Context,
	repository Phase151PendingRecoveryLeaderLeaseRepository,
	lease Phase151PendingRecoveryLeaderLease,
	duration time.Duration,
) (Phase151PendingRecoveryLeaderLease, bool, error) {
	if phase151PendingRecoveryLeaderLeaseDependenciesInvalid(ctx, repository) {
		return Phase151PendingRecoveryLeaderLease{}, false, ErrPostCreatePipelineInvalid
	}
	if err := ctx.Err(); err != nil {
		return Phase151PendingRecoveryLeaderLease{}, false, err
	}
	if !phase151PendingRecoveryLeaderLeaseValid(lease) || duration <= 0 {
		return Phase151PendingRecoveryLeaderLease{}, false, ErrPostCreatePipelineInvalid
	}

	renewed, owned, err := repository.RenewPendingRecoveryLeaderLease(ctx, lease, duration)
	if ctxErr := ctx.Err(); ctxErr != nil {
		return Phase151PendingRecoveryLeaderLease{}, false, ctxErr
	}
	if err != nil {
		return Phase151PendingRecoveryLeaderLease{}, false, err
	}
	if !owned {
		return Phase151PendingRecoveryLeaderLease{}, false, nil
	}
	if !phase151PendingRecoveryLeaderLeaseValid(renewed) || renewed.HolderID != lease.HolderID || renewed.Fence != lease.Fence {
		return Phase151PendingRecoveryLeaderLease{}, false, ErrPostCreatePipelineInvalid
	}

	return renewed, true, nil
}

func ExportPhase151ReleasePendingRecoveryLeaderLease(
	ctx context.Context,
	repository Phase151PendingRecoveryLeaderLeaseRepository,
	lease Phase151PendingRecoveryLeaderLease,
) (bool, error) {
	if phase151PendingRecoveryLeaderLeaseDependenciesInvalid(ctx, repository) {
		return false, ErrPostCreatePipelineInvalid
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if !phase151PendingRecoveryLeaderLeaseValid(lease) {
		return false, ErrPostCreatePipelineInvalid
	}

	owned, err := repository.ReleasePendingRecoveryLeaderLease(ctx, lease)
	if ctxErr := ctx.Err(); ctxErr != nil {
		return false, ctxErr
	}
	if err != nil {
		return false, err
	}

	return owned, nil
}

func ExportPhase151AdvancePendingRecoveryLeaderCheckpoint(
	ctx context.Context,
	repository Phase151PendingRecoveryLeaderLeaseRepository,
	lease Phase151PendingRecoveryLeaderLease,
	checkpoint Phase151PendingRecoveryLeaderCheckpoint,
) (bool, error) {
	if phase151PendingRecoveryLeaderLeaseDependenciesInvalid(ctx, repository) {
		return false, ErrPostCreatePipelineInvalid
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if !phase151PendingRecoveryLeaderLeaseValid(lease) || !phase151PendingRecoveryLeaderCheckpointValid(checkpoint, false) {
		return false, ErrPostCreatePipelineInvalid
	}

	owned, err := repository.AdvancePendingRecoveryLeaderCheckpoint(
		ctx,
		lease,
		phase151CopyPendingRecoveryLeaderCheckpoint(checkpoint),
	)
	if ctxErr := ctx.Err(); ctxErr != nil {
		return false, ctxErr
	}
	if err != nil {
		return false, err
	}

	return owned, nil
}

func phase151PendingRecoveryLeaderLeaseDependenciesInvalid(
	ctx context.Context,
	repository Phase151PendingRecoveryLeaderLeaseRepository,
) bool {
	return phase151PostCreateDependencyNil(repository) || phase151PostCreateDependencyNil(ctx)
}

func phase151PendingRecoveryLeaderLeaseValid(lease Phase151PendingRecoveryLeaderLease) bool {
	return lease.HolderID != uuid.Nil && lease.Fence > 0 && !lease.LeaseUntil.IsZero()
}

func phase151PendingRecoveryLeaderCheckpointValid(
	checkpoint Phase151PendingRecoveryLeaderCheckpoint,
	allowNilCursor bool,
) bool {
	if checkpoint.Cursor == nil {
		return allowNilCursor
	}
	return checkpoint.Cursor.TenantID != uuid.Nil && checkpoint.Cursor.ChannelID != uuid.Nil && checkpoint.Cursor.Kind != ""
}

func phase151CopyPendingRecoveryLeaderLeaseAcquisition(
	acquisition Phase151PendingRecoveryLeaderLeaseAcquisition,
) Phase151PendingRecoveryLeaderLeaseAcquisition {
	acquisition.Checkpoint = phase151CopyPendingRecoveryLeaderCheckpoint(acquisition.Checkpoint)
	return acquisition
}

func phase151CopyPendingRecoveryLeaderCheckpoint(
	checkpoint Phase151PendingRecoveryLeaderCheckpoint,
) Phase151PendingRecoveryLeaderCheckpoint {
	if checkpoint.Cursor == nil {
		return checkpoint
	}
	cursor := *checkpoint.Cursor
	checkpoint.Cursor = &cursor
	return checkpoint
}
