package notification_test

// Тесты модуля уведомлений: интеграционные и unit-сценарии testframework.

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/internal/notification"
)

func testNotifyProjectKey(prefix string) string {
	suffix := strings.ToUpper(strings.ReplaceAll(uuid.New().String(), "-", ""))[:4]
	return prefix + suffix
}

func newNotificationTestEnv(t *testing.T) (*notification.NotificationService, *notification.Repository, context.Context, uuid.UUID) {
	t.Helper()
	dsn := "file:tf_notif_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	user, err := client.User.Create().
		SetEmail("notif-" + uuid.New().String()[:8] + "@test.com").
		SetName("Recipient").
		Save(ctx)
	require.NoError(t, err)

	repo := notification.NewRepository(client)
	svc := notification.NewNotificationService(repo, nil, nil, nil)
	return svc, repo, ctx, user.ID
}

func newConsumerIssueEnv(t *testing.T) (*notification.Consumer, *notification.Repository, context.Context, *ent.User, *ent.User, *ent.Issue, uuid.UUID) {
	t.Helper()

	svc, repo, ctx, assigneeID := newNotificationTestEnv(t)
	assignee, err := notification.ExportRepositoryDB(repo).User.Get(ctx, assigneeID)
	require.NoError(t, err)

	actor, err := notification.ExportRepositoryDB(repo).User.Create().
		SetEmail("actor-" + uuid.New().String()[:8] + "@test.com").
		SetName("Actor").
		Save(ctx)
	require.NoError(t, err)

	proj, err := notification.ExportRepositoryDB(repo).Project.Create().
		SetName("Notify Project").
		SetKey(testNotifyProjectKey("NP")).
		Save(ctx)
	require.NoError(t, err)

	status, err := notification.ExportRepositoryDB(repo).WorkflowStatus.Create().
		SetName("Todo").
		SetCategory("todo").
		SetProjectID(proj.ID).
		SetIsDefault(true).
		Save(ctx)
	require.NoError(t, err)

	issueEnt, err := notification.ExportRepositoryDB(repo).Issue.Create().
		SetProjectID(proj.ID).
		SetStatusID(status.ID).
		SetReporterID(actor.ID).
		SetAssigneeID(assignee.ID).
		SetIssueNumber(1).
		SetIssueKey(proj.Key + "-1").
		SetTitle("Assigned issue").
		SetIssueType(issue.IssueTypeTask).
		SetPriority(issue.PriorityMedium).
		Save(ctx)
	require.NoError(t, err)

	c := notification.NewConsumer(nil, repo, svc)
	return c, repo, ctx, assignee, actor, issueEnt, proj.ID
}
