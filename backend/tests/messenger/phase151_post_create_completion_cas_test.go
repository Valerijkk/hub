package messenger_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/messenger"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestPhase151PostCreateCompletionPlanBuildsExactInitialWriterCASWouldCatchMissingProvenanceSenderOrFencePredicate(t *testing.T) {
	completedAt := time.Date(2026, time.August, 27, 21, 15, 30, 456000000, time.FixedZone("MSK", 3*60*60))
	scope := messenger.AuthorizedMessageProducerScope{
		TenantID: uuid.New(), ActorID: uuid.New(), ChannelID: uuid.New(),
		Producer: messenger.TrustedMessageProducerUser,
	}

	plan, err := messenger.ExportPhase151BuildPostCreateCompletionPlan(scope, "message-151-completion", completedAt)

	require.NoError(t, err)
	assert.Equal(t, messenger.Phase151PostCreateCompletionPlan{
		Filter: bson.D{
			{Key: "_id", Value: "message-151-completion"},
			{Key: "channel_id", Value: scope.ChannelID},
			{Key: "sender_id", Value: scope.ActorID},
			{Key: "post_create_producer_provenance", Value: &messenger.PostCreateProducerProvenance{
				Version: messenger.PostCreateProducerProvenanceVersionV1, Producer: messenger.TrustedMessageProducerUser,
			}},
			{Key: "deleted_at", Value: nil},
			{Key: "post_create_effects_state", Value: messenger.PostCreateEffectsStatePending},
			{Key: "post_create_effects_completed_at", Value: nil},
			{Key: "$or", Value: bson.A{
				bson.D{{Key: "post_create_recovery_fence", Value: bson.D{{Key: "$exists", Value: false}}}},
				bson.D{{Key: "post_create_recovery_fence", Value: int64(0)}},
			}},
		},
		Update: bson.D{{Key: "$set", Value: bson.D{
			{Key: "post_create_effects_state", Value: messenger.PostCreateEffectsStateComplete},
			{Key: "post_create_effects_completed_at", Value: time.Date(2026, time.August, 27, 18, 15, 30, 456000000, time.UTC)},
		}}},
	}, plan)
}

func TestPhase151PostCreateCompletionPlanEncodesSystemBotProvenanceAsExactWireBSONWouldCatchOmittedVersionBotKindOrIntegration(t *testing.T) {
	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	actorID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	channelID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	integrationID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	completedAt := time.Date(2026, time.August, 27, 18, 15, 30, 456000000, time.UTC)
	scope := messenger.AuthorizedMessageProducerScope{
		TenantID: tenantID, ActorID: actorID, ChannelID: channelID,
		Producer: messenger.TrustedMessageProducerSystemBot,
		BotKind:  messenger.BotKindSystem, IntegrationID: integrationID,
	}

	plan, err := messenger.ExportPhase151BuildPostCreateCompletionPlan(scope, "message-151-system-bot", completedAt)

	require.NoError(t, err)
	actual, err := bson.Marshal(plan.Filter)
	require.NoError(t, err)
	expected, err := bson.Marshal(bson.D{
		{Key: "_id", Value: "message-151-system-bot"},
		{Key: "channel_id", Value: channelID},
		{Key: "sender_id", Value: actorID},
		{Key: "post_create_producer_provenance", Value: bson.D{
			{Key: "version", Value: messenger.PostCreateProducerProvenanceVersionV1},
			{Key: "producer", Value: messenger.TrustedMessageProducerSystemBot},
			{Key: "bot_kind", Value: messenger.BotKindSystem},
			{Key: "integration_id", Value: integrationID},
		}},
		{Key: "deleted_at", Value: nil},
		{Key: "post_create_effects_state", Value: messenger.PostCreateEffectsStatePending},
		{Key: "post_create_effects_completed_at", Value: nil},
		{Key: "$or", Value: bson.A{
			bson.D{{Key: "post_create_recovery_fence", Value: bson.D{{Key: "$exists", Value: false}}}},
			bson.D{{Key: "post_create_recovery_fence", Value: int64(0)}},
		}},
	})
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}
func TestPhase151PostCreateCompletionPlanFailsClosedOnInvalidInputsWouldCatchDroppedTrustedScopeOrTimestampValidation(t *testing.T) {
	validScope := messenger.AuthorizedMessageProducerScope{
		TenantID: uuid.New(), ActorID: uuid.New(), ChannelID: uuid.New(),
		Producer: messenger.TrustedMessageProducerUser,
	}
	completedAt := time.Date(2026, time.August, 27, 18, 15, 30, 0, time.UTC)

	for _, scenario := range []struct {
		name        string
		messageID   string
		completedAt time.Time
		mutate      func(*messenger.AuthorizedMessageProducerScope)
		expected    error
	}{
		{
			name: "empty message ID", messageID: "", completedAt: completedAt,
			expected: messenger.ErrPostCreatePipelineInvalid,
		},
		{
			name: "zero completion time", messageID: "message-151-completion", expected: messenger.ErrPostCreatePipelineInvalid,
		},
		{
			name: "missing tenant", messageID: "message-151-completion", completedAt: completedAt,
			mutate:   func(scope *messenger.AuthorizedMessageProducerScope) { scope.TenantID = uuid.Nil },
			expected: messenger.ErrTrustedMessageProducerDenied,
		},
		{
			name: "missing sender", messageID: "message-151-completion", completedAt: completedAt,
			mutate:   func(scope *messenger.AuthorizedMessageProducerScope) { scope.ActorID = uuid.Nil },
			expected: messenger.ErrTrustedMessageProducerDenied,
		},
		{
			name: "missing channel", messageID: "message-151-completion", completedAt: completedAt,
			mutate:   func(scope *messenger.AuthorizedMessageProducerScope) { scope.ChannelID = uuid.Nil },
			expected: messenger.ErrTrustedMessageProducerDenied,
		},
		{
			name: "unknown producer", messageID: "message-151-completion", completedAt: completedAt,
			mutate: func(scope *messenger.AuthorizedMessageProducerScope) {
				scope.Producer = messenger.TrustedMessageProducer("unknown")
			},
			expected: messenger.ErrTrustedMessageProducerDenied,
		},
		{
			name: "user scope with bot kind", messageID: "message-151-completion", completedAt: completedAt,
			mutate:   func(scope *messenger.AuthorizedMessageProducerScope) { scope.BotKind = messenger.BotKindSystem },
			expected: messenger.ErrTrustedMessageProducerDenied,
		},
		{
			name: "system bot without integration", messageID: "message-151-completion", completedAt: completedAt,
			mutate: func(scope *messenger.AuthorizedMessageProducerScope) {
				scope.Producer = messenger.TrustedMessageProducerSystemBot
				scope.BotKind = messenger.BotKindSystem
			},
			expected: messenger.ErrTrustedMessageProducerDenied,
		},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			scope := validScope
			if scenario.mutate != nil {
				scenario.mutate(&scope)
			}

			plan, err := messenger.ExportPhase151BuildPostCreateCompletionPlan(scope, scenario.messageID, scenario.completedAt)

			require.ErrorIs(t, err, scenario.expected)
			assert.Equal(t, messenger.Phase151PostCreateCompletionPlan{}, plan)
		})
	}
}
