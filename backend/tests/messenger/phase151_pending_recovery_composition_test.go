//go:build hubtest

package messenger_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/messenger"
)

var _ messenger.Phase151PostCreateRecoveryClock = (*phase151PendingRecoveryCompositionClock)(nil)
var _ messenger.Phase151PostCreateMentionResolver = (*phase151PendingRecoveryCompositionMentions)(nil)
var _ messenger.Phase151PostCreateUnreadInvalidator = (*phase151PendingRecoveryCompositionInvalidation)(nil)

type phase151PendingRecoveryCompositionClock struct {
	calls int
}

func (c *phase151PendingRecoveryCompositionClock) Now() time.Time {
	c.calls++
	return time.Time{}
}

type phase151PendingRecoveryCompositionMentions struct{}

func (*phase151PendingRecoveryCompositionMentions) ResolveEligibleMentionUserIDs(
	context.Context,
	messenger.AuthorizedMessageProducerScope,
	messenger.PostCreatePipelineRequest,
) ([]uuid.UUID, error) {
	return nil, nil
}

type phase151PendingRecoveryCompositionInvalidation struct{}

func (*phase151PendingRecoveryCompositionInvalidation) PublishUnreadInvalidation(
	context.Context,
	messenger.AuthorizedMessageProducerScope,
	*messenger.MessageDocument,
) error {
	return nil
}

func phase151PendingRecoveryCompositionExecution() messenger.Phase151PendingRecoveryConfig {
	return messenger.Phase151PendingRecoveryConfig{
		MaxPages:         3,
		MaxTargets:       5,
		PendingScanLimit: 7,
		MinimumInterval:  11 * time.Second,
	}
}

func phase151PendingRecoveryCompositionPostCreateRepository() *messenger.Phase151PostCreateRepository {
	return messenger.NewPhase151PostCreateRepository(
		messenger.NewCombinedRepository(&messenger.EntRepository{}, &messenger.MongoRepository{}),
		&phase151PendingRecoveryCompositionMentions{},
		&phase151PendingRecoveryCompositionInvalidation{},
	)
}

func TestPhase151PendingRecoveryCompositionRetainsOneOpaqueRepositoryAndExecutionValue(t *testing.T) {
	execution := phase151PendingRecoveryCompositionExecution()
	wantExecution := execution
	config := messenger.NewPhase151PendingRecoveryRuntimeConfig(uuid.New(), 13*time.Second, execution)
	require.NotNil(t, config)

	execution.MaxPages = 97
	execution.MaxTargets = 89
	execution.PendingScanLimit = 83
	execution.MinimumInterval = time.Nanosecond
	clock := &phase151PendingRecoveryCompositionClock{}
	composition := messenger.NewPhase151PendingRecoveryComposition(
		config,
		phase151PendingRecoveryCompositionPostCreateRepository(),
		clock,
	)

	require.NotNil(t, composition)
	assert.Zero(t, clock.calls)
	first := composition.PendingRecoveryRepository()
	var _ messenger.Phase151PendingRecoveryRepository = first
	require.NotNil(t, first)
	second := composition.PendingRecoveryRepository()
	require.Same(t, first, second)

	assert.Equal(t, wantExecution, composition.PendingRecoveryConfig())
	returned := composition.PendingRecoveryConfig()
	returned.MaxPages = 79
	returned.MaxTargets = 73
	returned.PendingScanLimit = 71
	returned.MinimumInterval = time.Hour
	assert.Equal(t, wantExecution, composition.PendingRecoveryConfig())
	assert.Zero(t, clock.calls)
}

func TestPhase151PendingRecoveryCompositionNilReceiverAccessorsReturnZeroValues(t *testing.T) {
	var composition *messenger.Phase151PendingRecoveryComposition

	assert.Nil(t, composition.PendingRecoveryRepository())
	assert.Equal(t, messenger.Phase151PendingRecoveryConfig{}, composition.PendingRecoveryConfig())
}

func TestPhase151PendingRecoveryCompositionRejectsInvalidConstructionBeforeClockAccess(t *testing.T) {
	validConfig := messenger.NewPhase151PendingRecoveryRuntimeConfig(
		uuid.New(), 13*time.Second, phase151PendingRecoveryCompositionExecution(),
	)
	require.NotNil(t, validConfig)

	tests := []struct {
		name       string
		config     *messenger.Phase151PendingRecoveryRuntimeConfig
		repository *messenger.Phase151PostCreateRepository
		clock      messenger.Phase151PostCreateRecoveryClock
	}{
		{name: "nil runtime config", repository: phase151PendingRecoveryCompositionPostCreateRepository(), clock: &phase151PendingRecoveryCompositionClock{}},
		{name: "nil post-create repository", config: validConfig, clock: &phase151PendingRecoveryCompositionClock{}},
		{name: "nil combined repository", config: validConfig, repository: &messenger.Phase151PostCreateRepository{}, clock: &phase151PendingRecoveryCompositionClock{}},
		{name: "nil ent repository", config: validConfig, repository: &messenger.Phase151PostCreateRepository{CombinedRepository: messenger.NewCombinedRepository(nil, &messenger.MongoRepository{})}, clock: &phase151PendingRecoveryCompositionClock{}},
		{name: "nil mongo repository", config: validConfig, repository: &messenger.Phase151PostCreateRepository{CombinedRepository: messenger.NewCombinedRepository(&messenger.EntRepository{}, nil)}, clock: &phase151PendingRecoveryCompositionClock{}},
		{name: "typed nil clock", config: validConfig, repository: phase151PendingRecoveryCompositionPostCreateRepository(), clock: phase151PendingRecoveryCompositionTypedNilClock()},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			clock, ok := test.clock.(*phase151PendingRecoveryCompositionClock)
			composition := messenger.NewPhase151PendingRecoveryComposition(test.config, test.repository, test.clock)

			require.Nil(t, composition)
			if ok && clock != nil {
				assert.Zero(t, clock.calls)
			}
		})
	}
}

func phase151PendingRecoveryCompositionTypedNilClock() messenger.Phase151PostCreateRecoveryClock {
	var clock *phase151PendingRecoveryCompositionClock
	return clock
}

func TestPhase151PendingRecoveryCompositionUsesOnlyValidatedRuntimeConfig(t *testing.T) {
	validOwnerID := uuid.New()
	validExecution := phase151PendingRecoveryCompositionExecution()
	tests := []struct {
		name          string
		ownerID       uuid.UUID
		leaseDuration time.Duration
		execution     messenger.Phase151PendingRecoveryConfig
	}{
		{name: "zero owner", leaseDuration: time.Second, execution: validExecution},
		{name: "zero lease", ownerID: validOwnerID, execution: validExecution},
		{name: "zero max pages", ownerID: validOwnerID, leaseDuration: time.Second, execution: messenger.Phase151PendingRecoveryConfig{MaxTargets: 1, PendingScanLimit: 1, MinimumInterval: time.Second}},
		{name: "zero max targets", ownerID: validOwnerID, leaseDuration: time.Second, execution: messenger.Phase151PendingRecoveryConfig{MaxPages: 1, PendingScanLimit: 1, MinimumInterval: time.Second}},
		{name: "zero pending scan limit", ownerID: validOwnerID, leaseDuration: time.Second, execution: messenger.Phase151PendingRecoveryConfig{MaxPages: 1, MaxTargets: 1, MinimumInterval: time.Second}},
		{name: "zero minimum interval", ownerID: validOwnerID, leaseDuration: time.Second, execution: messenger.Phase151PendingRecoveryConfig{MaxPages: 1, MaxTargets: 1, PendingScanLimit: 1}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := messenger.NewPhase151PendingRecoveryRuntimeConfig(test.ownerID, test.leaseDuration, test.execution)
			require.Nil(t, config)
			clock := &phase151PendingRecoveryCompositionClock{}

			composition := messenger.NewPhase151PendingRecoveryComposition(
				config,
				phase151PendingRecoveryCompositionPostCreateRepository(),
				clock,
			)

			require.Nil(t, composition)
			assert.Zero(t, clock.calls)
		})
	}
}
