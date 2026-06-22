package notification_test

// Тесты репозитория уведомлений: интеграционные и unit-сценарии testframework.

import (
	"context"
	"testing"

	"github.com/valerii/hub/backend/internal/notification"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/channel"
	"github.com/valerii/hub/backend/ent/channelmember"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/issue"
)

// TestRepository_GetChannelAndMembers проверяет слой репозитория: репозиторий получение channel and members.
func TestRepository_GetChannelAndMembers(t *testing.T) {
	dsn := "file:notifrepo_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
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

	ch, err := client.Channel.Create().
		SetChannelType(channel.ChannelTypePublic).
		SetName("general").
		SetCreatedBy(owner.ID).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.ChannelMember.Create().
		SetChannelID(ch.ID).
		SetUserID(owner.ID).
		SetRole(channelmember.RoleOwner).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.ChannelMember.Create().
		SetChannelID(ch.ID).
		SetUserID(peer.ID).
		SetRole(channelmember.RoleMember).
		Save(ctx)
	require.NoError(t, err)

	repo := notification.NewRepository(client)

	got, err := repo.GetChannel(ctx, ch.ID)
	require.NoError(t, err)
	assert.Equal(t, "general", got.Name)

	members, err := repo.GetChannelMembers(ctx, ch.ID)
	require.NoError(t, err)
	assert.Len(t, members, 2)
}

// TestRepository_TelegramFields проверяет слой репозитория: репозиторий Telegram fields.
func TestRepository_TelegramFields(t *testing.T) {
	dsn := "file:notifrepo2_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	u, err := client.User.Create().
		SetEmail("tg-" + uuid.New().String()[:8] + "@test.com").
		SetName("TG").
		Save(ctx)
	require.NoError(t, err)

	repo := notification.NewRepository(client)
	require.NoError(t, repo.SetUserTelegram(ctx, u.ID, "123456789"))

	got, err := repo.GetUserTelegram(ctx, u.ID)
	require.NoError(t, err)
	require.NotNil(t, got.TelegramID)
	assert.Equal(t, "123456789", *got.TelegramID)

	found, err := repo.FindUserByTelegramID(ctx, "123456789")
	require.NoError(t, err)
	assert.Equal(t, u.ID, found.ID)

	require.NoError(t, repo.ClearUserTelegram(ctx, u.ID))
	cleared, err := repo.GetUser(ctx, u.ID)
	require.NoError(t, err)
	assert.Nil(t, cleared.TelegramID)
}

// TestRepository_GetIssueAssignee_NoAssignee проверяет слой репозитория: репозиторий получение issue assignee no assignee.
func TestRepository_GetIssueAssignee_NoAssignee(t *testing.T) {
	dsn := "file:notifrepo3_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	reporter, err := client.User.Create().
		SetEmail("rep-" + uuid.New().String()[:8] + "@test.com").
		SetName("Reporter").
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().
		SetName("Notify Project").
		SetKey(testNotifyProjectKey("N")).
		Save(ctx)
	require.NoError(t, err)

	status, err := client.WorkflowStatus.Create().
		SetName("Todo").
		SetCategory("todo").
		SetProjectID(proj.ID).
		SetIsDefault(true).
		Save(ctx)
	require.NoError(t, err)

	issueEnt, err := client.Issue.Create().
		SetProjectID(proj.ID).
		SetStatusID(status.ID).
		SetReporterID(reporter.ID).
		SetIssueNumber(1).
		SetIssueKey(proj.Key + "-1").
		SetTitle("Unassigned").
		SetIssueType(issue.IssueTypeTask).
		SetPriority(issue.PriorityMedium).
		Save(ctx)
	require.NoError(t, err)

	repo := notification.NewRepository(client)
	assignee, err := repo.GetIssueAssignee(ctx, issueEnt.ID)
	require.NoError(t, err)
	assert.Nil(t, assignee)
}
