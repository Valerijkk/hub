package messenger

import "github.com/google/uuid"

// Phase151PendingRecoveryComposition holds one inactive, fully composed pending
// recovery repository with the immutable execution value for a separately
// authorized future installer.
type Phase151PendingRecoveryComposition struct {
	repository *Phase151ClaimedRecoveryCoordinatorAdapter
	execution  Phase151PendingRecoveryConfig
}

// NewPhase151PendingRecoveryComposition composes the one post-create storage
// graph, process-local interval gate, and claimed recovery coordinator without
// invoking recovery execution or retaining the mutable runtime configuration.
func NewPhase151PendingRecoveryComposition(
	config *Phase151PendingRecoveryRuntimeConfig,
	postCreateRepository *Phase151PostCreateRepository,
	clock Phase151PostCreateRecoveryClock,
) *Phase151PendingRecoveryComposition {
	if config == nil {
		return nil
	}
	ownerID := config.OwnerID()
	leaseDuration := config.LeaseDuration()
	execution := config.ExecutionConfig()
	if ownerID == uuid.Nil || leaseDuration <= 0 || execution.MaxPages <= 0 ||
		execution.MaxTargets <= 0 || execution.PendingScanLimit <= 0 || execution.MinimumInterval <= 0 {
		return nil
	}
	if postCreateRepository == nil || postCreateRepository.CombinedRepository == nil ||
		postCreateRepository.CombinedRepository.EntRepository == nil ||
		postCreateRepository.CombinedRepository.MongoRepository == nil {
		return nil
	}
	if phase151PostCreateDependencyNil(clock) {
		return nil
	}

	entRepository := postCreateRepository.CombinedRepository.EntRepository
	mongoRepository := postCreateRepository.CombinedRepository.MongoRepository
	reauthorizer := NewPGPhase151PendingRecoveryTargetReauthorizer(entRepository)
	if reauthorizer == nil {
		return nil
	}
	scanner := NewPhase151PostCreateRecoveryScannerAdapter(entRepository, reauthorizer, mongoRepository)
	if scanner == nil {
		return nil
	}
	intervalGate := NewPhase151ProcessLocalPendingRecoveryIntervalGate(clock)
	if intervalGate == nil {
		return nil
	}
	repository := NewPhase151ClaimedRecoveryCoordinatorAdapter(
		scanner,
		intervalGate,
		ownerID,
		leaseDuration,
		clock,
		postCreateRepository,
	)
	if repository == nil {
		return nil
	}

	return &Phase151PendingRecoveryComposition{repository: repository, execution: execution}
}

// PendingRecoveryRepository returns the held, inactive pending-recovery
// coordinator without executing it.
func (c *Phase151PendingRecoveryComposition) PendingRecoveryRepository() Phase151PendingRecoveryRepository {
	if c == nil {
		return nil
	}
	return c.repository
}

// PendingRecoveryConfig returns a value copy of the execution configuration
// retained during composition.
func (c *Phase151PendingRecoveryComposition) PendingRecoveryConfig() Phase151PendingRecoveryConfig {
	if c == nil {
		return Phase151PendingRecoveryConfig{}
	}
	return c.execution
}
