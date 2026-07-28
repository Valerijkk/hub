package messenger_test

// Тесты репозитория мессенджера: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/messenger"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/projectmember"
)

// TestEntRepository_ChannelInvites_CRUD проверяет слой репозитория: ent репозиторий channel invites crud.
func TestEntRepository_ChannelInvites_CRUD(t *testing.T) {
	env := newMessengerRepoEnv(t)
	ch := env.createPublicChannel(t, "invites")

	inv, err := env.repo.CreateChannelInvite(env.ctx, messenger.CreateChannelInviteInput{
		ChannelID: ch.ID,
		CreatorID: env.owner.ID,
		Token:     "tok-" + uuid.New().String()[:8],
		MaxUses:   5,
	})
	require.NoError(t, err)
	assert.Equal(t, 0, inv.UseCount)
	assert.False(t, inv.IsRevoked)

	found, err := env.repo.FindChannelInviteByToken(env.ctx, inv.Token)
	require.NoError(t, err)
	assert.Equal(t, inv.ID, found.ID)

	require.NoError(t, env.repo.IncrementInviteUses(env.ctx, inv.ID, inv.MaxUses))
	updated, err := env.repo.FindChannelInviteByToken(env.ctx, inv.Token)
	require.NoError(t, err)
	assert.Equal(t, 1, updated.UseCount)

	list, err := env.repo.ListChannelInvites(env.ctx, inv.ChannelID)
	require.NoError(t, err)
	require.Len(t, list, 1)

	require.NoError(t, env.repo.RevokeChannelInvite(env.ctx, inv.ChannelID, inv.ID))
	list, err = env.repo.ListChannelInvites(env.ctx, inv.ChannelID)
	require.NoError(t, err)
	assert.Empty(t, list)
}

// TestEntRepository_ProjectMembersAndDisplayName проверяет слой репозитория: ent репозиторий project members and display name.
func TestEntRepository_ProjectMembersAndDisplayName(t *testing.T) {
	env := newMessengerRepoEnv(t)

	proj, err := env.client.Project.Create().
		SetName("Messenger Proj").
		SetKey(testMessengerProjectKey()).
		Save(env.ctx)
	require.NoError(t, err)

	_, err = env.client.ProjectMember.Create().
		SetProjectID(proj.ID).
		SetUserID(env.owner.ID).
		SetProjectRole(projectmember.ProjectRoleOwner).
		Save(env.ctx)
	require.NoError(t, err)
	_, err = env.client.ProjectMember.Create().
		SetProjectID(proj.ID).
		SetUserID(env.peer.ID).
		SetProjectRole(projectmember.ProjectRoleDeveloper).
		Save(env.ctx)
	require.NoError(t, err)

	gotProj, err := env.repo.GetProjectByID(env.ctx, proj.ID)
	require.NoError(t, err)
	assert.Equal(t, proj.Key, gotProj.Key)

	ids, err := env.repo.ListProjectMembers(env.ctx, proj.ID)
	require.NoError(t, err)
	assert.Len(t, ids, 2)

	byRole, err := env.repo.ListProjectMembersByRole(env.ctx, proj.ID, string(projectmember.ProjectRoleDeveloper))
	require.NoError(t, err)
	require.Len(t, byRole, 1)
	assert.Equal(t, env.peer.ID, byRole[0])

	display, err := env.repo.GetUserDisplayName(env.ctx, env.owner.ID)
	require.NoError(t, err)
	assert.Equal(t, "Owner", display)

	missing, err := env.repo.GetUserDisplayName(env.ctx, uuid.New())
	require.NoError(t, err)
	assert.Empty(t, missing)
}

// TestEntRepository_CreatePollWithOptions проверяет слой репозитория: ent репозиторий создание опрос with options.
func TestEntRepository_CreatePollWithOptions(t *testing.T) {
	env := newMessengerRepoEnv(t)
	ch := env.createPublicChannel(t, "polls")

	poll, options, err := env.repo.CreatePollWithOptions(env.ctx, messenger.CreatePollInput{
		ChannelID:     ch.ID,
		CreatorID:     env.owner.ID,
		Question:      "Ship it?",
		IsAnonymous:   false,
		IsMultiChoice: false,
		Options:       []string{"Yes", "No"},
	})
	require.NoError(t, err)
	assert.Equal(t, "Ship it?", poll.Question)
	require.Len(t, options, 2)
}
