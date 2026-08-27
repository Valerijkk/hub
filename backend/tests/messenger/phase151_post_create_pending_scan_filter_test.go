package messenger_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/messenger"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestPhase151PendingPostCreateScanPlanBuildsExactRecoverableCandidateFilterAndOrdering(t *testing.T) {
	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	actorID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	channelID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	scope := messenger.AuthorizedMessageProducerScope{
		TenantID: tenantID, ActorID: actorID, ChannelID: channelID,
		Producer: messenger.TrustedMessageProducerUser,
	}
	target := messenger.Phase151PendingRecoveryTarget{
		TenantID: tenantID, ChannelID: channelID, Kind: "channel",
	}

	plan, err := messenger.ExportPhase151BuildPendingPostCreateScanPlan(scope, target, 501)

	require.NoError(t, err)
	assert.Equal(t, messenger.Phase151PendingPostCreateScanPlan{
		Filter: bson.D{
			{Key: "channel_id", Value: channelID},
			{Key: "deleted_at", Value: nil},
			{Key: "post_create_effects_state", Value: messenger.PostCreateEffectsStatePending},
			{Key: "post_create_effects_completed_at", Value: nil},
			{Key: "post_create_producer_provenance.version", Value: messenger.PostCreateProducerProvenanceVersionV1},
		},
		Sort: bson.D{
			{Key: "created_at", Value: 1},
			{Key: "_id", Value: 1},
		},
		Limit: 500,
	}, plan)

	for _, candidate := range plan.Filter {
		assert.NotContains(t, candidate.Key, "post_create_recovery_")
	}
}

func TestPhase151PendingPostCreateScanPlanFailsClosedOnInvalidAuthorityOrTarget(t *testing.T) {
	validScope := messenger.AuthorizedMessageProducerScope{
		TenantID: uuid.New(), ActorID: uuid.New(), ChannelID: uuid.New(),
		Producer: messenger.TrustedMessageProducerUser,
	}
	validTarget := messenger.Phase151PendingRecoveryTarget{
		TenantID: validScope.TenantID, ChannelID: validScope.ChannelID, Kind: "channel",
	}

	for _, scenario := range []struct {
		name         string
		mutateScope  func(*messenger.AuthorizedMessageProducerScope)
		mutateTarget func(*messenger.Phase151PendingRecoveryTarget)
		expected     error
	}{
		{
			name: "nil actor",
			mutateScope: func(scope *messenger.AuthorizedMessageProducerScope) {
				scope.ActorID = uuid.Nil
			},
			expected: messenger.ErrTrustedMessageProducerDenied,
		},
		{
			name: "unknown producer",
			mutateScope: func(scope *messenger.AuthorizedMessageProducerScope) {
				scope.Producer = messenger.TrustedMessageProducer("unknown")
			},
			expected: messenger.ErrTrustedMessageProducerDenied,
		},
		{
			name: "tenant drift",
			mutateScope: func(scope *messenger.AuthorizedMessageProducerScope) {
				scope.TenantID = uuid.New()
			},
			expected: messenger.ErrPostCreatePipelineInvalid,
		},
		{
			name: "channel drift",
			mutateScope: func(scope *messenger.AuthorizedMessageProducerScope) {
				scope.ChannelID = uuid.New()
			},
			expected: messenger.ErrPostCreatePipelineInvalid,
		},
		{
			name: "target without tenant",
			mutateTarget: func(target *messenger.Phase151PendingRecoveryTarget) {
				target.TenantID = uuid.Nil
			},
			expected: messenger.ErrPostCreatePipelineInvalid,
		},
		{
			name: "target without channel",
			mutateTarget: func(target *messenger.Phase151PendingRecoveryTarget) {
				target.ChannelID = uuid.Nil
			},
			expected: messenger.ErrPostCreatePipelineInvalid,
		},
		{
			name: "target without kind",
			mutateTarget: func(target *messenger.Phase151PendingRecoveryTarget) {
				target.Kind = ""
			},
			expected: messenger.ErrPostCreatePipelineInvalid,
		},
		{
			name:     "nonpositive limit",
			expected: messenger.ErrPostCreatePipelineInvalid,
		},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			scope := validScope
			target := validTarget
			limit := 7
			if scenario.name == "nonpositive limit" {
				limit = 0
			}
			if scenario.mutateScope != nil {
				scenario.mutateScope(&scope)
			}
			if scenario.mutateTarget != nil {
				scenario.mutateTarget(&target)
			}

			plan, err := messenger.ExportPhase151BuildPendingPostCreateScanPlan(scope, target, limit)

			require.ErrorIs(t, err, scenario.expected)
			assert.Equal(t, messenger.Phase151PendingPostCreateScanPlan{}, plan)
		})
	}
}
