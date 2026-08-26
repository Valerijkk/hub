package messenger_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/messenger"
)

func TestPhase151PublishTyping_RequiresFreshChannelActionTypingProofBeforePublish(t *testing.T) {
	ctx := context.Background()
	actorID, channelID := uuid.New(), uuid.New()
	resolver := &phase151ChannelAccessResolver{grants: make(map[phase151ChannelAccessKey]messenger.ChannelAccess)}
	resolver.grants[phase151ChannelAccessKey{actorID: actorID, channelID: channelID, action: messenger.ChannelActionTyping}] = messenger.ChannelAccess{
		TenantID: uuid.New(), Channel: &ent.Channel{ID: channelID}, IsMember: true, CanRead: true, CanType: true,
	}
	publisher := &capturePublisher{}

	err := messenger.ExportPhase151PublishTyping(ctx, actorID, channelID, resolver, publisher)

	require.NoError(t, err)
	assert.Equal(t, []phase151ChannelAccessKey{{actorID: actorID, channelID: channelID, action: messenger.ChannelActionTyping}}, resolver.calls)
	require.Len(t, publisher.snapshot(), 1)
	assert.Equal(t, "messenger.channel."+channelID.String()+".typing", publisher.snapshot()[0].subject)

	delete(resolver.grants, phase151ChannelAccessKey{actorID: actorID, channelID: channelID, action: messenger.ChannelActionTyping})
	err = messenger.ExportPhase151PublishTyping(ctx, actorID, channelID, resolver, publisher)
	require.Error(t, err)
	assert.Len(t, resolver.calls, 2, "typing authority must be read fresh after enqueue")
	assert.Len(t, publisher.snapshot(), 1, "revoked sender must emit zero additional NATS events")
}

func TestPhase151PublishTyping_DeniedSenderStatesEmitZeroEvents(t *testing.T) {
	for _, state := range []string{"disabled account", "private nonmember", "lost project membership"} {
		t.Run(state, func(t *testing.T) {
			actorID, channelID := uuid.New(), uuid.New()
			resolver := &phase151ChannelAccessResolver{grants: make(map[phase151ChannelAccessKey]messenger.ChannelAccess)}
			publisher := &capturePublisher{}

			err := messenger.ExportPhase151PublishTyping(
				context.Background(), actorID, channelID, resolver, publisher,
			)

			require.Error(t, err)
			assert.Equal(t, []phase151ChannelAccessKey{{actorID: actorID, channelID: channelID, action: messenger.ChannelActionTyping}}, resolver.calls)
			assert.Empty(t, publisher.snapshot())
		})
	}
}

func TestPhase151PublishTyping_RejectsNilAuthorityOrPublisher(t *testing.T) {
	actorID, channelID := uuid.New(), uuid.New()
	resolver := &phase151ChannelAccessResolver{grants: make(map[phase151ChannelAccessKey]messenger.ChannelAccess)}
	publisher := &capturePublisher{}
	assert.Panics(t, func() {
		messenger.ExportPhase151PublishTyping(context.Background(), actorID, channelID, nil, publisher)
	})
	assert.Panics(t, func() {
		messenger.ExportPhase151PublishTyping(context.Background(), actorID, channelID, resolver, nil)
	})
}
