package messenger

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Phase151PostCreateRecoveryClock interface {
	Now() time.Time
}

type Phase151ClaimedPostCreateRecoveryRepository interface {
	ClaimPendingPostCreateRecovery(
		context.Context,
		AuthorizedMessageProducerScope,
		*MessageDocument,
		uuid.UUID,
		time.Time,
		time.Duration,
	) (Phase151PostCreateRecoveryAcquisition, bool, error)
	EnsureAutoThreadFollows(context.Context, AuthorizedMessageProducerScope, *MessageDocument) error
	ReconcileReplyCount(context.Context, AuthorizedMessageProducerScope, *MessageDocument) error
	CompleteClaimedPostCreateRecovery(
		context.Context,
		AuthorizedMessageProducerScope,
		*MessageDocument,
		Phase151PostCreateRecoveryClaim,
		time.Time,
	) error
	PublishUnreadInvalidation(context.Context, AuthorizedMessageProducerScope, *MessageDocument) error
}

var _ Phase151ClaimedPostCreateRecoveryRepository = (*Phase151PostCreateRepository)(nil)

// ExportPhase151ExecuteClaimedPostCreateRecovery executes the already-bounded
// post-create effects only while holding one exact Mongo owner/fence/lease.
// A true return means a claim was acquired, including failures after claim.
func ExportPhase151ExecuteClaimedPostCreateRecovery(
	ctx context.Context,
	targetAuthority AuthorizedMessageProducerScope,
	scanned *MessageDocument,
	ownerID uuid.UUID,
	leaseDuration time.Duration,
	clock Phase151PostCreateRecoveryClock,
	repository Phase151ClaimedPostCreateRecoveryRepository,
) (bool, error) {
	if ctx == nil || phase151PostCreateDependencyNil(clock) || phase151PostCreateDependencyNil(repository) {
		return false, ErrPostCreatePipelineInvalid
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}

	scope, err := ExportPhase151ReconstructPostCreateRecoveryProducerScope(targetAuthority, scanned)
	if err != nil {
		return false, err
	}
	claimNow := clock.Now()
	expectedPlan, err := ExportPhase151BuildPostCreateRecoveryClaimPlan(
		scope,
		scanned,
		ownerID,
		claimNow,
		leaseDuration,
	)
	if err != nil {
		return false, err
	}

	acquisition, acquired, err := repository.ClaimPendingPostCreateRecovery(
		ctx,
		scope,
		scanned,
		ownerID,
		claimNow,
		leaseDuration,
	)
	if err != nil {
		return false, fmt.Errorf("claim pending post-create recovery: %w", err)
	}
	if !acquired {
		return false, nil
	}
	if err := validatePhase151ClaimedRecoveryRunnerAcquisition(
		scope,
		scanned,
		acquisition,
		expectedPlan,
	); err != nil {
		return true, err
	}
	claimedScope, err := ExportPhase151ReconstructPostCreateRecoveryProducerScope(
		targetAuthority,
		&acquisition.Message,
	)
	if err != nil {
		return true, err
	}
	if claimedScope != scope {
		return true, ErrTrustedMessageProducerDenied
	}
	if err := ctx.Err(); err != nil {
		return true, err
	}

	if err := repository.EnsureAutoThreadFollows(ctx, claimedScope, &acquisition.Message); err != nil {
		return true, fmt.Errorf("ensure automatic thread follows: %w", err)
	}
	if acquisition.Message.ThreadID != "" {
		if err := repository.ReconcileReplyCount(ctx, claimedScope, &acquisition.Message); err != nil {
			return true, fmt.Errorf("reconcile reply count: %w", err)
		}
	}

	completedAt := clock.Now()
	if _, err := ExportPhase151BuildPostCreateRecoveryCompletionPlan(
		claimedScope,
		&acquisition.Message,
		acquisition.Claim,
		completedAt,
	); err != nil {
		return true, err
	}
	if err := ctx.Err(); err != nil {
		return true, err
	}
	if err := repository.CompleteClaimedPostCreateRecovery(
		ctx,
		claimedScope,
		&acquisition.Message,
		acquisition.Claim,
		completedAt,
	); err != nil {
		return true, fmt.Errorf("complete claimed post-create recovery: %w", err)
	}
	if err := repository.PublishUnreadInvalidation(ctx, claimedScope, &acquisition.Message); err != nil {
		return true, fmt.Errorf("publish unread invalidation: %w", err)
	}
	return true, nil
}

func validatePhase151ClaimedRecoveryRunnerAcquisition(
	scope AuthorizedMessageProducerScope,
	scanned *MessageDocument,
	acquisition Phase151PostCreateRecoveryAcquisition,
	expectedPlan Phase151PostCreateRecoveryClaimPlan,
) error {
	if err := validatePhase151PostCreateRecoveryClaimResult(
		scope,
		scanned,
		&acquisition.Message,
		expectedPlan,
	); err != nil {
		return err
	}
	claim := acquisition.Claim
	if claim.TenantID != scope.TenantID || claim.MessageID != acquisition.Message.ID ||
		claim.ChannelID != scope.ChannelID || claim.OwnerID != expectedPlan.OwnerID ||
		claim.Fence != acquisition.Message.PostCreateRecoveryFence ||
		claim.LeaseUntil.IsZero() || !claim.LeaseUntil.UTC().Equal(expectedPlan.LeaseUntil) {
		return ErrPostCreatePipelineInvalid
	}
	return nil
}
