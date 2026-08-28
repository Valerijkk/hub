package messenger

import (
	"time"

	"github.com/google/uuid"
)

// Phase151PendingRecoveryRuntimeConfig holds immutable owner, lease, and execution
// parameters for a later pending-recovery factory. Its owner ID represents an
// intended future process-lifetime claim owner; it does not prove process
// singleton or distributed uniqueness.
type Phase151PendingRecoveryRuntimeConfig struct {
	ownerID       uuid.UUID
	leaseDuration time.Duration
	execution     Phase151PendingRecoveryConfig
}

// NewPhase151PendingRecoveryRuntimeConfig creates a runtime configuration when
// every owner, lease, and execution parameter is positive and valid. It returns
// nil when any required value is invalid.
func NewPhase151PendingRecoveryRuntimeConfig(
	ownerID uuid.UUID,
	leaseDuration time.Duration,
	execution Phase151PendingRecoveryConfig,
) *Phase151PendingRecoveryRuntimeConfig {
	if ownerID == uuid.Nil || leaseDuration <= 0 || execution.MaxPages <= 0 ||
		execution.MaxTargets <= 0 || execution.PendingScanLimit <= 0 || execution.MinimumInterval <= 0 {
		return nil
	}

	return &Phase151PendingRecoveryRuntimeConfig{
		ownerID:       ownerID,
		leaseDuration: leaseDuration,
		execution:     execution,
	}
}

// OwnerID returns the future process-lifetime claim owner identifier.
func (c *Phase151PendingRecoveryRuntimeConfig) OwnerID() uuid.UUID {
	if c == nil {
		return uuid.Nil
	}

	return c.ownerID
}

// LeaseDuration returns the configured future claim lease duration.
func (c *Phase151PendingRecoveryRuntimeConfig) LeaseDuration() time.Duration {
	if c == nil {
		return 0
	}

	return c.leaseDuration
}

// ExecutionConfig returns a value copy of the pending-recovery execution configuration.
func (c *Phase151PendingRecoveryRuntimeConfig) ExecutionConfig() Phase151PendingRecoveryConfig {
	if c == nil {
		return Phase151PendingRecoveryConfig{}
	}

	return c.execution
}
