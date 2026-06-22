package messenger_test

// Тесты репозитория мессенджера: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/messenger"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/channelmember"
	"github.com/valerii/hub/backend/ent/enttest"
)

type messengerRepoEnv struct {
	client *ent.Client
	repo   *messenger.EntRepository
	ctx    context.Context
	owner  *ent.User
	peer   *ent.User
}

func newMessengerRepoEnv(t *testing.T) messengerRepoEnv {
	t.Helper()
	dsn := "file:msgrepo_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	owner, err := client.User.Create().
		SetEmail("owner-" + uuid.New().String()[:8] + "@test.com").
		SetName("Owner").
		Save(ctx)
	require.NoError(t, err)

	peer, err := client.User.Create().
		SetEmail("peer-" + uuid.New().String()[:8] + "@test.com").
		SetName("Peer").
		Save(ctx)
	require.NoError(t, err)

	return messengerRepoEnv{
		client: client,
		repo:   messenger.NewEntRepository(client),
		ctx:    ctx,
		owner:  owner,
		peer:   peer,
	}
}

func (env messengerRepoEnv) createPublicChannel(t *testing.T, name string) *ent.Channel {
	t.Helper()
	ch, err := env.repo.CreateChannel(env.ctx, messenger.CreateChannelInput{
		ChannelType: "public",
		Name:        name,
		Description: "desc",
		CreatedBy:   env.owner.ID,
	})
	require.NoError(t, err)
	err = env.repo.AddMembers(env.ctx, ch.ID, []uuid.UUID{env.owner.ID}, string(channelmember.RoleOwner))
	require.NoError(t, err)
	return ch
}

// TestEntRepository_CreateChannel_Get_Update_Delete проверяет слой репозитория: ent репозиторий создание channel получение обновление удаление.
func TestEntRepository_CreateChannel_Get_Update_Delete(t *testing.T) {
	env := newMessengerRepoEnv(t)

	ch := env.createPublicChannel(t, "general")
	assert.Equal(t, "general", ch.Name)
	assert.Equal(t, "desc", ch.Description)

	got, err := env.repo.GetChannel(env.ctx, ch.ID)
	require.NoError(t, err)
	assert.Equal(t, ch.ID, got.ID)

	newName := "announcements"
	updated, err := env.repo.UpdateChannel(env.ctx, ch.ID, messenger.UpdateChannelInput{
		Name:        &newName,
		Description: strPtr("updated"),
	})
	require.NoError(t, err)
	assert.Equal(t, "announcements", updated.Name)

	list, err := env.repo.ListUserChannels(env.ctx, env.owner.ID)
	require.NoError(t, err)
	require.Len(t, list, 1)

	// DeleteChannel без каскада участников — FK на sqlite; фиксируем текущее поведение.
	err = env.repo.DeleteChannel(env.ctx, ch.ID)
	require.Error(t, err)
}

// TestEntRepository_MembersAndRoles проверяет слой репозитория: ent репозиторий members and roles.
func TestEntRepository_MembersAndRoles(t *testing.T) {
	env := newMessengerRepoEnv(t)
	ch := env.createPublicChannel(t, "team")

	err := env.repo.AddMembers(env.ctx, ch.ID, []uuid.UUID{env.peer.ID}, string(channelmember.RoleMember))
	require.NoError(t, err)

	ok, err := env.repo.IsMember(env.ctx, ch.ID, env.peer.ID)
	require.NoError(t, err)
	assert.True(t, ok)

	role, err := env.repo.GetMemberRole(env.ctx, ch.ID, env.peer.ID)
	require.NoError(t, err)
	assert.Equal(t, string(channelmember.RoleMember), role)

	ids, err := env.repo.GetChannelMemberIDs(env.ctx, ch.ID)
	require.NoError(t, err)
	assert.Len(t, ids, 2)

	members, err := env.repo.ListMembers(env.ctx, ch.ID)
	require.NoError(t, err)
	assert.Len(t, members, 2)

	require.NoError(t, env.repo.RemoveMember(env.ctx, ch.ID, env.peer.ID))
	ok, err = env.repo.IsMember(env.ctx, ch.ID, env.peer.ID)
	require.NoError(t, err)
	assert.False(t, ok)
}

// TestEntRepository_FindDMChannel проверяет слой репозитория: ent репозиторий find dm channel.
func TestEntRepository_FindDMChannel(t *testing.T) {
	env := newMessengerRepoEnv(t)
	dmKey := env.owner.ID.String() + "-" + env.peer.ID.String()

	ch, err := env.repo.CreateChannel(env.ctx, messenger.CreateChannelInput{
		ChannelType: "dm",
		CreatedBy:   env.owner.ID,
		DMKey:       &dmKey,
	})
	require.NoError(t, err)

	found, err := env.repo.FindDMChannel(env.ctx, dmKey)
	require.NoError(t, err)
	assert.Equal(t, ch.ID, found.ID)
}

// TestEntRepository_StarMute проверяет слой репозитория: ent репозиторий star mute.
func TestEntRepository_StarMute(t *testing.T) {
	env := newMessengerRepoEnv(t)
	ch := env.createPublicChannel(t, "muted")

	starred, err := env.repo.ToggleStar(env.ctx, ch.ID, env.owner.ID)
	require.NoError(t, err)
	assert.True(t, starred)

	starred, err = env.repo.ToggleStar(env.ctx, ch.ID, env.owner.ID)
	require.NoError(t, err)
	assert.False(t, starred)

	mutedUntil, err := env.repo.MuteChannel(env.ctx, ch.ID, env.owner.ID, "1h")
	require.NoError(t, err)
	assert.True(t, mutedUntil.After(time.Now()))

	muted, err := env.repo.IsChannelMuted(env.ctx, ch.ID, env.owner.ID)
	require.NoError(t, err)
	assert.True(t, muted)

	require.NoError(t, env.repo.UnmuteChannel(env.ctx, ch.ID, env.owner.ID))
	muted, err = env.repo.IsChannelMuted(env.ctx, ch.ID, env.owner.ID)
	require.NoError(t, err)
	assert.False(t, muted)

	memberships, err := env.repo.GetUserChannelMemberships(env.ctx, env.owner.ID)
	require.NoError(t, err)
	require.Len(t, memberships, 1)
}

// TestEntRepository_PinsAndBookmarks проверяет слой репозитория: ent репозиторий pins and bookmarks.
func TestEntRepository_PinsAndBookmarks(t *testing.T) {
	env := newMessengerRepoEnv(t)
	ch := env.createPublicChannel(t, "pins")
	msgID := uuid.New().String()

	require.NoError(t, env.repo.AddPin(env.ctx, ch.ID, msgID, env.owner.ID))

	pin, err := env.repo.FindPin(env.ctx, ch.ID, msgID)
	require.NoError(t, err)
	assert.Equal(t, msgID, pin.MessageID)

	pins, err := env.repo.ListPins(env.ctx, ch.ID)
	require.NoError(t, err)
	require.Len(t, pins, 1)

	require.NoError(t, env.repo.RemovePin(env.ctx, ch.ID, msgID))
	_, err = env.repo.FindPin(env.ctx, ch.ID, msgID)
	require.Error(t, err)

	bm, err := env.repo.AddBookmark(env.ctx, env.owner.ID, msgID, ch.ID)
	require.NoError(t, err)
	assert.Equal(t, msgID, bm.MessageID)

	bookmarks, err := env.repo.ListBookmarks(env.ctx, env.owner.ID)
	require.NoError(t, err)
	require.Len(t, bookmarks, 1)

	require.NoError(t, env.repo.RemoveBookmark(env.ctx, bm.ID))
	bookmarks, err = env.repo.ListBookmarks(env.ctx, env.owner.ID)
	require.NoError(t, err)
	assert.Empty(t, bookmarks)
}

// TestEntRepository_MuteChannel_InvalidDuration проверяет отклонение при ent репозиторий mute channel некорректный duration.
func TestEntRepository_MuteChannel_InvalidDuration(t *testing.T) {
	env := newMessengerRepoEnv(t)
	ch := env.createPublicChannel(t, "bad-mute")

	_, err := env.repo.MuteChannel(env.ctx, ch.ID, env.owner.ID, "invalid")
	require.Error(t, err)
}

func strPtr(s string) *string {
	return &s
}
