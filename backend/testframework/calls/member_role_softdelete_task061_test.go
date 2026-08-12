//go:build hubtest

package calls_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/channelmember"
)

// TestRepository_GetChannelMemberRole_SkipsSoftDeletedChannel_P3BE05 —
// soft-deleted channel всё ещё отдавал role (в отличие от IsChannelMember).
// Mutation: снять HasChannelWith(DeletedAtIsNil) → FAIL.
func TestRepository_GetChannelMemberRole_SkipsSoftDeletedChannel_P3BE05(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()

	role, err := f.repo.GetChannelMemberRole(ctx, f.channel.ID, f.owner.ID)
	require.NoError(t, err)
	assert.Equal(t, string(channelmember.RoleOwner), role)

	now := time.Now().UTC()
	_, err = f.client.Channel.UpdateOneID(f.channel.ID).SetDeletedAt(now).Save(ctx)
	require.NoError(t, err)

	role, err = f.repo.GetChannelMemberRole(ctx, f.channel.ID, f.owner.ID)
	require.NoError(t, err)
	assert.Equal(t, "", role)

	ok, err := f.repo.IsChannelMember(ctx, f.channel.ID, f.owner.ID)
	require.NoError(t, err)
	assert.False(t, ok)
}
