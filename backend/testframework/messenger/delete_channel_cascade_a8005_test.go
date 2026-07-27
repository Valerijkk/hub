package messenger_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/channelmember"
)

// TestDeleteChannel_CascadesMembersPins_A8_005 — hard-delete канала с members/pins
// больше не падает на FK (tx чистит children до DeleteOneID).
func TestDeleteChannel_CascadesMembersPins_A8_005(t *testing.T) {
	env := newMessengerRepoEnv(t)
	ch := env.createPublicChannel(t, "cascade-a8005")

	require.NoError(t, env.repo.AddMembers(env.ctx, ch.ID, []uuid.UUID{env.peer.ID}, string(channelmember.RoleMember)))
	require.NoError(t, env.repo.AddPin(env.ctx, ch.ID, "mongo-msg-1", env.owner.ID))

	require.NoError(t, env.repo.DeleteChannel(env.ctx, ch.ID))

	_, err := env.repo.GetChannel(env.ctx, ch.ID)
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))

	ok, err := env.repo.IsMember(env.ctx, ch.ID, env.owner.ID)
	require.NoError(t, err)
	assert.False(t, ok)

	_, err = env.repo.FindPin(env.ctx, ch.ID, "mongo-msg-1")
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))
}
