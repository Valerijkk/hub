package notification_test

// Тесты NATS consumer уведомлений: интеграционные и unit-сценарии testframework.

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/valerii/hub/backend/internal/notification"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	entnotif "github.com/valerii/hub/backend/ent/notification"
	"github.com/valerii/hub/backend/internal/auth"
	"github.com/valerii/hub/backend/internal/tms"
	"github.com/valerii/hub/backend/internal/tracker"
)

// TestConsumer_shouldDispatch_Debounce проверяет сценарий: consumer should dispatch дебаунс.
func TestConsumer_shouldDispatch_Debounce(t *testing.T) {
	c := notification.NewConsumer(nil, nil, nil)
	key := uuid.New().String()

	assert.True(t, notification.ExportShouldDispatch(c, key, notification.TypeAssigned, "issue-1"))
	assert.False(t, notification.ExportShouldDispatch(c, key, notification.TypeAssigned, "issue-1"))
	assert.True(t, notification.ExportShouldDispatch(c, key, notification.TypeAssigned, "issue-2"))
}

// TestConsumer_cleanDebounce проверяет сценарий: consumer clean дебаунс.
func TestConsumer_cleanDebounce(t *testing.T) {
	c := notification.NewConsumer(nil, nil, nil)
	key := uuid.New().String()
	notification.ExportSeedDebounceEntry(c, key+":"+notification.TypeAssigned+":entity", time.Now().Add(-10*time.Second))
	notification.ExportCleanDebounce(c)
	assert.Equal(t, 0, notification.ExportDebounceLen(c))
}

// TestConsumer_handleAssigned_CreatesNotification проверяет сценарий: consumer handle assigned creates notification.
func TestConsumer_handleAssigned_CreatesNotification(t *testing.T) {
	svc, repo, ctx, assigneeID := newNotificationTestEnv(t)
	actor, err := notification.ExportRepositoryDB(repo).User.Create().
		SetEmail("actor-" + uuid.New().String()[:8] + "@test.com").
		SetName("Actor").
		Save(ctx)
	require.NoError(t, err)

	c := notification.NewConsumer(nil, repo, svc)
	evt := tracker.IssueAssignedEvent{
		IssueID:      uuid.New().String(),
		IssueKey:     "HUB-42",
		ProjectID:    uuid.New().String(),
		AssigneeID:   assigneeID.String(),
		AssignedByID: actor.ID.String(),
	}
	data, err := json.Marshal(evt)
	require.NoError(t, err)

	notification.ExportHandleAssigned(c, ctx, data, "project.x.issue.assigned")

	notifs, err := notification.ExportRepositoryDB(repo).Notification.Query().
		Where(entnotif.UserID(assigneeID)).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, notifs, 1)
	assert.Equal(t, notification.TypeAssigned, notifs[0].Type)
	assert.Contains(t, notifs[0].Title, "HUB-42")
}

// TestConsumer_handleAssigned_SkipsSelfAssignment проверяет сценарий: consumer handle assigned skips self assignment.
func TestConsumer_handleAssigned_SkipsSelfAssignment(t *testing.T) {
	svc, repo, ctx, assigneeID := newNotificationTestEnv(t)
	c := notification.NewConsumer(nil, repo, svc)

	evt := tracker.IssueAssignedEvent{
		IssueID:      uuid.New().String(),
		IssueKey:     "HUB-1",
		ProjectID:    uuid.New().String(),
		AssigneeID:   assigneeID.String(),
		AssignedByID: assigneeID.String(),
	}
	data, err := json.Marshal(evt)
	require.NoError(t, err)

	notification.ExportHandleAssigned(c, ctx, data, "")

	count, err := notification.ExportRepositoryDB(repo).Notification.Query().
		Where(entnotif.UserID(assigneeID)).
		Count(ctx)
	require.NoError(t, err)
	assert.Zero(t, count)
}

// TestConsumer_handleNewLogin_CreatesNotification проверяет сценарий: consumer handle new вход creates notification.
func TestConsumer_handleNewLogin_CreatesNotification(t *testing.T) {
	svc, repo, ctx, userID := newNotificationTestEnv(t)
	c := notification.NewConsumer(nil, repo, svc)

	evt := auth.SessionCreatedEvent{
		SessionID: uuid.New().String(),
		UserID:    userID.String(),
		Device:    "Chrome",
		IPAddress: "127.0.0.1",
	}
	data, err := json.Marshal(evt)
	require.NoError(t, err)

	notification.ExportHandleNewLogin(c, ctx, data)

	notifs, err := notification.ExportRepositoryDB(repo).Notification.Query().
		Where(entnotif.UserID(userID)).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, notifs, 1)
	assert.Equal(t, notification.TypeNewLogin, notifs[0].Type)
	assert.Contains(t, notifs[0].Body, "127.0.0.1")
}

// TestConsumer_handleStatusChanged_NotifiesAssignee проверяет сценарий: consumer handle status changed notifies assignee.
func TestConsumer_handleStatusChanged_NotifiesAssignee(t *testing.T) {
	c, repo, ctx, assignee, actor, issueEnt, projectID := newConsumerIssueEnv(t)

	evt := tracker.IssueMovedEvent{
		IssueID:   issueEnt.ID.String(),
		IssueKey:  issueEnt.IssueKey,
		ProjectID: projectID.String(),
		MovedByID: actor.ID.String(),
	}
	data, err := json.Marshal(evt)
	require.NoError(t, err)

	notification.ExportHandleStatusChanged(c, ctx, data, "")

	notifs, err := notification.ExportRepositoryDB(repo).Notification.Query().
		Where(entnotif.UserID(assignee.ID)).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, notifs, 1)
	assert.Equal(t, notification.TypeStatusChanged, notifs[0].Type)
}

// TestConsumer_handleCommentAdded_NotifiesAssigneeAndMention проверяет сценарий: consumer handle comment added notifies assignee and mention.
func TestConsumer_handleCommentAdded_NotifiesAssigneeAndMention(t *testing.T) {
	c, repo, ctx, assignee, actor, issueEnt, projectID := newConsumerIssueEnv(t)

	mention, err := notification.ExportRepositoryDB(repo).User.Create().
		SetEmail("mention-" + uuid.New().String()[:8] + "@test.com").
		SetName("Mentioned").
		Save(ctx)
	require.NoError(t, err)

	evt := tracker.CommentCreatedEvent{
		CommentID: uuid.New().String(),
		IssueID:   issueEnt.ID.String(),
		IssueKey:  issueEnt.IssueKey,
		ProjectID: projectID.String(),
		AuthorID:  actor.ID.String(),
		Mentions:  []string{mention.ID.String()},
	}
	data, err := json.Marshal(evt)
	require.NoError(t, err)

	notification.ExportHandleCommentAdded(c, ctx, data, "")

	assigneeNotifs, err := notification.ExportRepositoryDB(repo).Notification.Query().
		Where(entnotif.UserID(assignee.ID)).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, assigneeNotifs, 1)
	assert.Equal(t, notification.TypeCommentAdded, assigneeNotifs[0].Type)

	mentionNotifs, err := notification.ExportRepositoryDB(repo).Notification.Query().
		Where(entnotif.UserID(mention.ID)).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, mentionNotifs, 1)
	assert.Equal(t, notification.TypeMentioned, mentionNotifs[0].Type)
}

// TestConsumer_handleRunCompleted_NotifiesTesterOnFailure проверяет сценарий: consumer handle run completed notifies tester on failure.
func TestConsumer_handleRunCompleted_NotifiesTesterOnFailure(t *testing.T) {
	svc, repo, ctx, ownerID := newNotificationTestEnv(t)
	completer, err := notification.ExportRepositoryDB(repo).User.Create().
		SetEmail("complete-" + uuid.New().String()[:8] + "@test.com").
		SetName("Completer").
		Save(ctx)
	require.NoError(t, err)

	tester, err := notification.ExportRepositoryDB(repo).User.Create().
		SetEmail("tester-" + uuid.New().String()[:8] + "@test.com").
		SetName("Tester").
		Save(ctx)
	require.NoError(t, err)

	proj, err := notification.ExportRepositoryDB(repo).Project.Create().
		SetName("TMS Proj").
		SetKey(testNotifyProjectKey("T")).
		Save(ctx)
	require.NoError(t, err)

	_, err = notification.ExportRepositoryDB(repo).ProjectMember.Create().
		SetProjectID(proj.ID).
		SetUserID(ownerID).
		SetProjectRole("developer").
		Save(ctx)
	require.NoError(t, err)

	_, err = notification.ExportRepositoryDB(repo).ProjectMember.Create().
		SetProjectID(proj.ID).
		SetUserID(tester.ID).
		SetProjectRole("tester").
		Save(ctx)
	require.NoError(t, err)

	c := notification.NewConsumer(nil, repo, svc)
	evt := tms.TestRunCompletedEvent{
		RunID:       uuid.New().String(),
		ProjectID:   proj.ID.String(),
		Title:       "Nightly",
		CompletedBy: completer.ID.String(),
		Passed:      10,
		Failed:      1,
		Skipped:     2,
	}
	data, err := json.Marshal(evt)
	require.NoError(t, err)

	notification.ExportHandleRunCompleted(c, ctx, data, "")

	ownerNotifs, err := notification.ExportRepositoryDB(repo).Notification.Query().
		Where(entnotif.UserID(ownerID)).
		All(ctx)
	require.NoError(t, err)
	assert.Empty(t, ownerNotifs, "developer should not receive run_failed")

	testerNotifs, err := notification.ExportRepositoryDB(repo).Notification.Query().
		Where(entnotif.UserID(tester.ID)).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, testerNotifs, 1)
	assert.Equal(t, notification.TypeRunFailed, testerNotifs[0].Type)
	assert.True(t, testerNotifs[0].IsUrgent)
}

// TestConsumer_handleRunCompleted_NotifiesMemberOnSuccess проверяет сценарий: consumer handle run completed notifies участник on успех.
func TestConsumer_handleRunCompleted_NotifiesMemberOnSuccess(t *testing.T) {
	svc, repo, ctx, ownerID := newNotificationTestEnv(t)
	completer, err := notification.ExportRepositoryDB(repo).User.Create().
		SetEmail("complete-" + uuid.New().String()[:8] + "@test.com").
		SetName("Completer").
		Save(ctx)
	require.NoError(t, err)

	proj, err := notification.ExportRepositoryDB(repo).Project.Create().
		SetName("TMS Proj OK").
		SetKey(testNotifyProjectKey("O")).
		Save(ctx)
	require.NoError(t, err)

	_, err = notification.ExportRepositoryDB(repo).ProjectMember.Create().
		SetProjectID(proj.ID).
		SetUserID(ownerID).
		SetProjectRole("developer").
		Save(ctx)
	require.NoError(t, err)

	c := notification.NewConsumer(nil, repo, svc)
	evt := tms.TestRunCompletedEvent{
		RunID:       uuid.New().String(),
		ProjectID:   proj.ID.String(),
		Title:       "Nightly",
		CompletedBy: completer.ID.String(),
		Passed:      10,
		Failed:      0,
		Skipped:     2,
	}
	data, err := json.Marshal(evt)
	require.NoError(t, err)

	notification.ExportHandleRunCompleted(c, ctx, data, "")

	notifs, err := notification.ExportRepositoryDB(repo).Notification.Query().
		Where(entnotif.UserID(ownerID)).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, notifs, 1)
	assert.Equal(t, notification.TypeRunCompleted, notifs[0].Type)
}
