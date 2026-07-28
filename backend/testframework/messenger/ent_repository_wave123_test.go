package messenger_test

// Тесты репозитория мессенджера (волна 123): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/messenger"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Фаза 123 — ent-репозиторий coverage gaps..

func TestEntRepository_Wave123_MuteAllDurations(t *testing.T) {
	env := newMessengerRepoEnv(t)
	ch := env.createPublicChannel(t, "mute-all")

	for _, dur := range []string{"2h", "8h", "1d", "forever"} {
		until, err := env.repo.MuteChannel(env.ctx, ch.ID, env.owner.ID, dur)
		require.NoError(t, err, "duration %s", dur)
		assert.True(t, until.After(time.Now()))

		muted, err := env.repo.IsChannelMuted(env.ctx, ch.ID, env.owner.ID)
		require.NoError(t, err)
		assert.True(t, muted)

		require.NoError(t, env.repo.UnmuteChannel(env.ctx, ch.ID, env.owner.ID))
	}
}

// TestEntRepository_Wave123_GetUserDisplayName проверяет слой репозитория: ent репозиторий wave 123 получение user display name.
func TestEntRepository_Wave123_GetUserDisplayName(t *testing.T) {
	env := newMessengerRepoEnv(t)

	display, err := env.client.User.UpdateOneID(env.owner.ID).
		SetDisplayName("Captain").
		Save(env.ctx)
	require.NoError(t, err)

	name, err := env.repo.GetUserDisplayName(env.ctx, display.ID)
	require.NoError(t, err)
	assert.Equal(t, "Captain", name)

	name, err = env.repo.GetUserDisplayName(env.ctx, uuid.New())
	require.NoError(t, err)
	assert.Empty(t, name)
}

// TestEntRepository_Wave123_ChannelInviteLifecycle проверяет слой репозитория: ent репозиторий wave 123 channel приглашение жизненный цикл.
func TestEntRepository_Wave123_ChannelInviteLifecycle(t *testing.T) {
	env := newMessengerRepoEnv(t)
	ch := env.createPublicChannel(t, "invites")

	exp := time.Now().Add(24 * time.Hour)
	inv, err := env.repo.CreateChannelInvite(env.ctx, messenger.CreateChannelInviteInput{
		ChannelID: ch.ID, CreatorID: env.owner.ID, Token: "tok-wave123", MaxUses: 3, ExpiresAt: &exp,
	})
	require.NoError(t, err)

	found, err := env.repo.FindChannelInviteByToken(env.ctx, "tok-wave123")
	require.NoError(t, err)
	assert.Equal(t, inv.ID, found.ID)

	require.NoError(t, env.repo.IncrementInviteUses(env.ctx, inv.ID))

	list, err := env.repo.ListChannelInvites(env.ctx, ch.ID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, 1, list[0].UseCount)

	require.NoError(t, env.repo.RevokeChannelInvite(env.ctx, ch.ID, inv.ID))
	list, err = env.repo.ListChannelInvites(env.ctx, ch.ID)
	require.NoError(t, err)
	assert.Empty(t, list)
}

// TestEntRepository_Wave123_ListProjectMembers проверяет слой репозитория: ent репозиторий wave 123 список project members.
func TestEntRepository_Wave123_ListProjectMembers(t *testing.T) {
	env := newMessengerRepoEnv(t)

	proj, err := env.client.Project.Create().
		SetName("MsgProj").
		SetKey("MSG").
		Save(env.ctx)
	require.NoError(t, err)

	_, err = env.client.ProjectMember.Create().
		SetUserID(env.owner.ID).
		SetProjectID(proj.ID).
		SetProjectRole("owner").
		Save(env.ctx)
	require.NoError(t, err)
	_, err = env.client.ProjectMember.Create().
		SetUserID(env.peer.ID).
		SetProjectID(proj.ID).
		SetProjectRole("developer").
		Save(env.ctx)
	require.NoError(t, err)

	ids, err := env.repo.ListProjectMembers(env.ctx, proj.ID)
	require.NoError(t, err)
	assert.Len(t, ids, 2)
}

// TestEntRepository_Wave123_VotePollSingleChoice проверяет слой репозитория: ent репозиторий wave 123 vote опрос single choice.
func TestEntRepository_Wave123_VotePollSingleChoice(t *testing.T) {
	env := newMessengerRepoEnv(t)
	ch := env.createPublicChannel(t, "poll-single")

	poll, options, err := env.repo.CreatePollWithOptions(env.ctx, messenger.CreatePollInput{
		ChannelID: ch.ID, CreatorID: env.owner.ID, Question: "Pick one",
		IsMultiChoice: false, Options: []string{"A", "B"},
	})
	require.NoError(t, err)
	require.Len(t, options, 2)

	require.NoError(t, env.repo.VotePoll(env.ctx, poll.ID, env.peer.ID, []uuid.UUID{options[0].ID}))
	votes, err := env.repo.ListVotes(env.ctx, poll.ID)
	require.NoError(t, err)
	require.Len(t, votes, 1)
}
