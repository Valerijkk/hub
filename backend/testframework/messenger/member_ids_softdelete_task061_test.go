//go:build hubtest

package messenger_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/channelmember"
)

// TestEntRepository_GetChannelMemberIDs_SkipsSoftDeleted_P3BE05 — после soft-delete
// канала IsMember=false, но GetChannelMemberIDs без predicate всё ещё отдавал members.
// Mutation: снять HasChannelWith(DeletedAtIsNil) → FAIL.
func TestEntRepository_GetChannelMemberIDs_SkipsSoftDeleted_P3BE05(t *testing.T) {
	env := newMessengerRepoEnv(t)
	ch := env.createPublicChannel(t, "soft-members")

	require.NoError(t, env.repo.AddMembers(env.ctx, ch.ID, []uuid.UUID{env.peer.ID}, string(channelmember.RoleMember)))

	ids, err := env.repo.GetChannelMemberIDs(env.ctx, ch.ID)
	require.NoError(t, err)
	assert.Len(t, ids, 2)

	require.NoError(t, env.repo.DeleteChannel(env.ctx, ch.ID))

	ids, err = env.repo.GetChannelMemberIDs(env.ctx, ch.ID)
	require.NoError(t, err)
	assert.Empty(t, ids)

	ok, err := env.repo.IsMember(env.ctx, ch.ID, env.peer.ID)
	require.NoError(t, err)
	assert.False(t, ok)
}
