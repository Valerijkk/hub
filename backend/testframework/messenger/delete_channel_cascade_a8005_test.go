package messenger_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/channel"
	"github.com/valerii/hub/backend/ent/channelmember"
)

// TestDeleteChannel_SoftDelete_A8_005 — DeleteChannel ставит deleted_at;
// Get/List/IsMember скрывают soft-deleted (WS/membership не экспозят канал).
func TestDeleteChannel_SoftDelete_A8_005(t *testing.T) {
	env := newMessengerRepoEnv(t)
	ch := env.createPublicChannel(t, "soft-a8005")

	require.NoError(t, env.repo.AddMembers(env.ctx, ch.ID, []uuid.UUID{env.peer.ID}, string(channelmember.RoleMember)))
	require.NoError(t, env.repo.AddPin(env.ctx, ch.ID, "mongo-msg-1", env.owner.ID))

	require.NoError(t, env.repo.DeleteChannel(env.ctx, ch.ID))

	// Soft-delete: row остаётся с deleted_at, Get фильтрует.
	_, err := env.repo.GetChannel(env.ctx, ch.ID)
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))

	raw, err := env.client.Channel.Get(env.ctx, ch.ID)
	require.NoError(t, err)
	require.NotNil(t, raw.DeletedAt)

	ok, err := env.repo.IsMember(env.ctx, ch.ID, env.owner.ID)
	require.NoError(t, err)
	assert.False(t, ok, "WS/membership must not treat soft-deleted channel as joinable")

	list, err := env.repo.ListUserChannels(env.ctx, env.owner.ID)
	require.NoError(t, err)
	for _, c := range list {
		assert.NotEqual(t, ch.ID, c.ID)
	}

	// Children сохранены (hard-purge — отдельный admin path).
	_, err = env.repo.FindPin(env.ctx, ch.ID, "mongo-msg-1")
	require.NoError(t, err)

	memberExists, err := env.client.ChannelMember.Query().
		Where(
			channelmember.ChannelID(ch.ID),
			channelmember.UserID(env.owner.ID),
		).
		Exist(env.ctx)
	require.NoError(t, err)
	assert.True(t, memberExists)

	// Идемпотентность: повторный soft-delete → NotFound (уже deleted).
	err = env.repo.DeleteChannel(env.ctx, ch.ID)
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))

	// Фильтр DeletedAtIsNil на query-уровне (регрессия predicate).
	n, err := env.client.Channel.Query().
		Where(channel.ID(ch.ID), channel.DeletedAtIsNil()).
		Count(env.ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
}
