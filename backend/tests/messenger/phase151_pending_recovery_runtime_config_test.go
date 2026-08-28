//go:build hubtest

package messenger_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/messenger"
)

func TestPhase151PendingRecoveryRuntimeConfigPreservesExactValuesAndCopiesExecutionConfig(t *testing.T) {
	t.Parallel()

	ownerID := uuid.New()
	execution := messenger.Phase151PendingRecoveryConfig{
		MaxPages:         3,
		MaxTargets:       7,
		PendingScanLimit: 11,
		MinimumInterval:  2 * time.Minute,
	}
	config := messenger.NewPhase151PendingRecoveryRuntimeConfig(ownerID, 45*time.Second, execution)

	require.NotNil(t, config)
	require.Equal(t, ownerID, config.OwnerID())
	require.Equal(t, 45*time.Second, config.LeaseDuration())
	require.Equal(t, execution, config.ExecutionConfig())

	execution.MaxPages = 99
	execution.MinimumInterval = time.Second
	returned := config.ExecutionConfig()
	returned.MaxTargets = 101
	returned.PendingScanLimit = 103

	require.Equal(t, messenger.Phase151PendingRecoveryConfig{
		MaxPages:         3,
		MaxTargets:       7,
		PendingScanLimit: 11,
		MinimumInterval:  2 * time.Minute,
	}, config.ExecutionConfig())
}

func TestPhase151PendingRecoveryRuntimeConfigRejectsInvalidInputs(t *testing.T) {
	t.Parallel()

	validOwnerID := uuid.New()
	validExecution := messenger.Phase151PendingRecoveryConfig{
		MaxPages:         1,
		MaxTargets:       1,
		PendingScanLimit: 1,
		MinimumInterval:  time.Second,
	}

	tests := []struct {
		name          string
		ownerID       uuid.UUID
		leaseDuration time.Duration
		execution     messenger.Phase151PendingRecoveryConfig
	}{
		{name: "zero owner ID", leaseDuration: time.Second, execution: validExecution},
		{name: "zero lease", ownerID: validOwnerID, execution: validExecution},
		{name: "negative lease", ownerID: validOwnerID, leaseDuration: -time.Nanosecond, execution: validExecution},
		{name: "zero max pages", ownerID: validOwnerID, leaseDuration: time.Second, execution: messenger.Phase151PendingRecoveryConfig{MaxTargets: 1, PendingScanLimit: 1, MinimumInterval: time.Second}},
		{name: "negative max pages", ownerID: validOwnerID, leaseDuration: time.Second, execution: messenger.Phase151PendingRecoveryConfig{MaxPages: -1, MaxTargets: 1, PendingScanLimit: 1, MinimumInterval: time.Second}},
		{name: "zero max targets", ownerID: validOwnerID, leaseDuration: time.Second, execution: messenger.Phase151PendingRecoveryConfig{MaxPages: 1, PendingScanLimit: 1, MinimumInterval: time.Second}},
		{name: "negative max targets", ownerID: validOwnerID, leaseDuration: time.Second, execution: messenger.Phase151PendingRecoveryConfig{MaxPages: 1, MaxTargets: -1, PendingScanLimit: 1, MinimumInterval: time.Second}},
		{name: "zero pending scan limit", ownerID: validOwnerID, leaseDuration: time.Second, execution: messenger.Phase151PendingRecoveryConfig{MaxPages: 1, MaxTargets: 1, MinimumInterval: time.Second}},
		{name: "negative pending scan limit", ownerID: validOwnerID, leaseDuration: time.Second, execution: messenger.Phase151PendingRecoveryConfig{MaxPages: 1, MaxTargets: 1, PendingScanLimit: -1, MinimumInterval: time.Second}},
		{name: "zero minimum interval", ownerID: validOwnerID, leaseDuration: time.Second, execution: messenger.Phase151PendingRecoveryConfig{MaxPages: 1, MaxTargets: 1, PendingScanLimit: 1}},
		{name: "negative minimum interval", ownerID: validOwnerID, leaseDuration: time.Second, execution: messenger.Phase151PendingRecoveryConfig{MaxPages: 1, MaxTargets: 1, PendingScanLimit: 1, MinimumInterval: -time.Nanosecond}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			require.Nil(t, messenger.NewPhase151PendingRecoveryRuntimeConfig(test.ownerID, test.leaseDuration, test.execution))
		})
	}
}

func TestPhase151PendingRecoveryRuntimeConfigNilReceiverAccessorsReturnZeroValues(t *testing.T) {
	t.Parallel()

	var config *messenger.Phase151PendingRecoveryRuntimeConfig

	require.Equal(t, uuid.Nil, config.OwnerID())
	require.Equal(t, time.Duration(0), config.LeaseDuration())
	require.Equal(t, messenger.Phase151PendingRecoveryConfig{}, config.ExecutionConfig())
}

func TestPhase151PendingRecoveryRuntimeConfigDistinctConfigsDoNotShareState(t *testing.T) {
	t.Parallel()

	first := messenger.NewPhase151PendingRecoveryRuntimeConfig(
		uuid.New(),
		time.Second,
		messenger.Phase151PendingRecoveryConfig{MaxPages: 1, MaxTargets: 2, PendingScanLimit: 3, MinimumInterval: 4 * time.Second},
	)
	second := messenger.NewPhase151PendingRecoveryRuntimeConfig(
		uuid.New(),
		5*time.Second,
		messenger.Phase151PendingRecoveryConfig{MaxPages: 6, MaxTargets: 7, PendingScanLimit: 8, MinimumInterval: 9 * time.Second},
	)

	require.NotNil(t, first)
	require.NotNil(t, second)
	require.NotSame(t, first, second)
	require.NotEqual(t, first.OwnerID(), second.OwnerID())
	require.NotEqual(t, first.LeaseDuration(), second.LeaseDuration())
	require.NotEqual(t, first.ExecutionConfig(), second.ExecutionConfig())
}
