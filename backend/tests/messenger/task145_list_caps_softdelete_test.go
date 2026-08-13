//go:build hubtest

package messenger_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/channel"
	"github.com/valerii/hub/backend/ent/channelmember"
	"github.com/valerii/hub/backend/internal/messenger"
)

// Caps из messenger/limits.go (A2W3-3-002 / A3W3-3-001); не экспортированы — зеркало для mutation.
const (
	task145ListMembersCap           = 500
	task145UserChannelMembershipCap = 500
)

const task145BulkBatch = 50 // SQLite variable limit — не CreateBulk на 500+ строк разом.

// TestEntListUserChannels_HasHardCap_A3W3_3_001 — Limit(userChannelMembershipCap).
// Mutation: убрать Limit с ListUserChannels → FAIL (вернутся все cap+1 каналов).
func TestEntListUserChannels_HasHardCap_A3W3_3_001(t *testing.T) {
	env := newMessengerRepoEnv(t)
	total := task145UserChannelMembershipCap + 1

	chs := make([]*ent.Channel, 0, total)
	for start := 0; start < total; start += task145BulkBatch {
		end := start + task145BulkBatch
		if end > total {
			end = total
		}
		creates := make([]*ent.ChannelCreate, 0, end-start)
		for i := start; i < end; i++ {
			creates = append(creates, env.client.Channel.Create().
				SetChannelType(channel.ChannelTypePublic).
				SetName(fmt.Sprintf("cap-ch-%d", i)).
				SetCreatedBy(env.owner.ID))
		}
		batch, err := env.client.Channel.CreateBulk(creates...).Save(env.ctx)
		require.NoError(t, err)
		chs = append(chs, batch...)
	}

	for start := 0; start < len(chs); start += task145BulkBatch {
		end := start + task145BulkBatch
		if end > len(chs) {
			end = len(chs)
		}
		memberCreates := make([]*ent.ChannelMemberCreate, 0, end-start)
		for _, ch := range chs[start:end] {
			memberCreates = append(memberCreates, env.client.ChannelMember.Create().
				SetChannelID(ch.ID).
				SetUserID(env.owner.ID).
				SetRole(channelmember.RoleOwner))
		}
		_, err := env.client.ChannelMember.CreateBulk(memberCreates...).Save(env.ctx)
		require.NoError(t, err)
	}

	list, err := env.repo.ListUserChannels(env.ctx, env.owner.ID)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(list), task145UserChannelMembershipCap)
	assert.Equal(t, task145UserChannelMembershipCap, len(list),
		"без Limit вернулось бы %d; cap должен обрезать до %d", total, task145UserChannelMembershipCap)
}

// TestEntGetChannel_WithMembersCap_A3W3_3_001 — WithMembers Limit(listMembersCap).
// Mutation: убрать Limit из WithMembers на GetChannel → FAIL.
func TestEntGetChannel_WithMembersCap_A3W3_3_001(t *testing.T) {
	env := newMessengerRepoEnv(t)
	ch := env.createPublicChannel(t, "cap-eager-members")

	extra := task145ListMembersCap // owner уже в members → итого cap+1
	ids := make([]uuid.UUID, 0, extra)
	for start := 0; start < extra; start += task145BulkBatch {
		end := start + task145BulkBatch
		if end > extra {
			end = extra
		}
		userCreates := make([]*ent.UserCreate, 0, end-start)
		for i := start; i < end; i++ {
			userCreates = append(userCreates, env.client.User.Create().
				SetEmail(fmt.Sprintf("cap-m-%d-%s@test.com", i, uuid.New().String()[:8])).
				SetName(fmt.Sprintf("CapMember%d", i)))
		}
		users, err := env.client.User.CreateBulk(userCreates...).Save(env.ctx)
		require.NoError(t, err)
		for _, u := range users {
			ids = append(ids, u.ID)
		}
	}
	for start := 0; start < len(ids); start += task145BulkBatch {
		end := start + task145BulkBatch
		if end > len(ids) {
			end = len(ids)
		}
		require.NoError(t, env.repo.AddMembers(env.ctx, ch.ID, ids[start:end], string(channelmember.RoleMember)))
	}

	got, err := env.repo.GetChannel(env.ctx, ch.ID)
	require.NoError(t, err)
	require.NotNil(t, got.Edges.Members)
	assert.LessOrEqual(t, len(got.Edges.Members), task145ListMembersCap)
	assert.Equal(t, task145ListMembersCap, len(got.Edges.Members),
		"без Limit eager members было бы %d", task145ListMembersCap+1)
}

// TestEntFindDMChannel_WithMembersCap_A3W3_3_001 — тот же cap на FindDMChannel.
func TestEntFindDMChannel_WithMembersCap_A3W3_3_001(t *testing.T) {
	env := newMessengerRepoEnv(t)
	dmKey := env.owner.ID.String() + "-" + env.peer.ID.String()

	ch, err := env.repo.CreateChannel(env.ctx, messenger.CreateChannelInput{
		ChannelType: "dm",
		CreatedBy:   env.owner.ID,
		DMKey:       &dmKey,
	})
	require.NoError(t, err)
	require.NoError(t, env.repo.AddMembers(env.ctx, ch.ID,
		[]uuid.UUID{env.owner.ID, env.peer.ID}, string(channelmember.RoleMember)))

	// DM обычно 2 участника — проверяем, что eager load не падает и ≤ cap.
	found, err := env.repo.FindDMChannel(env.ctx, dmKey)
	require.NoError(t, err)
	require.NotNil(t, found.Edges.Members)
	assert.LessOrEqual(t, len(found.Edges.Members), task145ListMembersCap)
	assert.GreaterOrEqual(t, len(found.Edges.Members), 2)
}

// TestEntListMembers_SkipsSoftDeletedChannel_A3W3_3_002 — после DeleteChannel состав пуст.
// Mutation: убрать HasChannelWith(DeletedAtIsNil) с ListMembers → FAIL.
func TestEntListMembers_SkipsSoftDeletedChannel_A3W3_3_002(t *testing.T) {
	env := newMessengerRepoEnv(t)
	ch := env.createPublicChannel(t, "list-soft-members")
	require.NoError(t, env.repo.AddMembers(env.ctx, ch.ID, []uuid.UUID{env.peer.ID}, string(channelmember.RoleMember)))

	before, err := env.repo.ListMembers(env.ctx, ch.ID)
	require.NoError(t, err)
	assert.Len(t, before, 2)

	require.NoError(t, env.repo.DeleteChannel(env.ctx, ch.ID))

	after, err := env.repo.ListMembers(env.ctx, ch.ID)
	require.NoError(t, err)
	assert.Empty(t, after, "ListMembers must hide soft-deleted channel members")

	// Симметрия с GetChannelMemberIDs.
	ids, err := env.repo.GetChannelMemberIDs(env.ctx, ch.ID)
	require.NoError(t, err)
	assert.Empty(t, ids)
}
