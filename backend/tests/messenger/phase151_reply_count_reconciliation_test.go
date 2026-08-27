package messenger_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/messenger"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type phase151ReplyCountPort interface {
	ReconcileReplyCount(context.Context, messenger.AuthorizedMessageProducerScope, *messenger.MessageDocument) error
}

var _ phase151ReplyCountPort = (*messenger.MongoRepository)(nil)

func TestPhase151ReplyCountReconciliationPlanBindsLiveChannelAndRoot(t *testing.T) {
	tenantID := uuid.New()
	actorID := uuid.New()
	channelID := uuid.New()
	scope := messenger.AuthorizedMessageProducerScope{
		TenantID:  tenantID,
		ActorID:   actorID,
		ChannelID: channelID,
		Producer:  messenger.TrustedMessageProducerUser,
	}
	message := &messenger.MessageDocument{
		ID:                     "reply-1",
		ChannelID:              channelID,
		SenderID:               actorID,
		ThreadID:               "root-1",
		PostCreateEffectsState: messenger.PostCreateEffectsStatePending,
	}

	plan, err := messenger.ExportPhase151BuildReplyCountReconciliationPlan(scope, message)
	require.NoError(t, err)

	assert.Equal(t, bson.D{
		{Key: "channel_id", Value: channelID},
		{Key: "thread_id", Value: "root-1"},
		{Key: "deleted_at", Value: nil},
	}, plan.ReplyFilter)
	assert.Equal(t, bson.D{
		{Key: "_id", Value: "root-1"},
		{Key: "channel_id", Value: channelID},
		{Key: "deleted_at", Value: nil},
		{Key: "$or", Value: bson.A{
			bson.D{{Key: "thread_id", Value: bson.D{{Key: "$exists", Value: false}}}},
			bson.D{{Key: "thread_id", Value: ""}},
		}},
	}, plan.RootFilter)

	update, err := messenger.ExportPhase151BuildReplyCountRaiseUpdate(7)
	require.NoError(t, err)
	assert.Equal(t, bson.D{{Key: "$max", Value: bson.D{{Key: "reply_count", Value: int64(7)}}}}, update)
}

func TestPhase151ReplyCountReconciliationPlanRejectsUntrustedOrMalformedEffects(t *testing.T) {
	now := time.Now().UTC()
	validScope := messenger.AuthorizedMessageProducerScope{
		TenantID:  uuid.New(),
		ActorID:   uuid.New(),
		ChannelID: uuid.New(),
		Producer:  messenger.TrustedMessageProducerUser,
	}
	validMessage := func() *messenger.MessageDocument {
		return &messenger.MessageDocument{
			ID:                     "reply-1",
			ChannelID:              validScope.ChannelID,
			SenderID:               validScope.ActorID,
			ThreadID:               "root-1",
			PostCreateEffectsState: messenger.PostCreateEffectsStatePending,
		}
	}

	for _, scenario := range []struct {
		name     string
		scope    messenger.AuthorizedMessageProducerScope
		message  *messenger.MessageDocument
		expected error
	}{
		{name: "nil message", scope: validScope, message: nil, expected: messenger.ErrPostCreatePipelineInvalid},
		{name: "unknown producer", scope: func() messenger.AuthorizedMessageProducerScope {
			scope := validScope
			scope.Producer = messenger.TrustedMessageProducer("unknown")
			return scope
		}(), message: validMessage(), expected: messenger.ErrTrustedMessageProducerDenied},
		{name: "foreign channel", scope: validScope, message: func() *messenger.MessageDocument {
			message := validMessage()
			message.ChannelID = uuid.New()
			return message
		}(), expected: messenger.ErrPostCreatePipelineInvalid},
		{name: "foreign sender", scope: validScope, message: func() *messenger.MessageDocument {
			message := validMessage()
			message.SenderID = uuid.New()
			return message
		}(), expected: messenger.ErrPostCreatePipelineInvalid},
		{name: "not a reply", scope: validScope, message: func() *messenger.MessageDocument {
			message := validMessage()
			message.ThreadID = ""
			return message
		}(), expected: messenger.ErrPostCreatePipelineInvalid},
		{name: "self root", scope: validScope, message: func() *messenger.MessageDocument {
			message := validMessage()
			message.ThreadID = message.ID
			return message
		}(), expected: messenger.ErrPostCreatePipelineInvalid},
		{name: "deleted reply", scope: validScope, message: func() *messenger.MessageDocument {
			message := validMessage()
			message.DeletedAt = &now
			return message
		}(), expected: messenger.ErrPostCreatePipelineInvalid},
		{name: "completed effect", scope: validScope, message: func() *messenger.MessageDocument {
			message := validMessage()
			message.PostCreateEffectsState = messenger.PostCreateEffectsStateComplete
			return message
		}(), expected: messenger.ErrPostCreatePipelineInvalid},
		{name: "completed timestamp", scope: validScope, message: func() *messenger.MessageDocument {
			message := validMessage()
			message.PostCreateEffectsCompletedAt = &now
			return message
		}(), expected: messenger.ErrPostCreatePipelineInvalid},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			_, err := messenger.ExportPhase151BuildReplyCountReconciliationPlan(scenario.scope, scenario.message)
			require.ErrorIs(t, err, scenario.expected)
		})
	}
}

func TestPhase151ReplyCountRaiseUpdateRejectsImpossibleCounts(t *testing.T) {
	for _, count := range []int64{0, -1} {
		_, err := messenger.ExportPhase151BuildReplyCountRaiseUpdate(count)
		require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
	}
}
