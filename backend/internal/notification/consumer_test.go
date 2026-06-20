package notification

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/notification"
	"github.com/valerii/hub/backend/internal/auth"
	"github.com/valerii/hub/backend/internal/tracker"
	"github.com/valerii/hub/backend/internal/tms"
)

func testNotifyProjectKey(prefix string) string {
	suffix := strings.ToUpper(strings.ReplaceAll(uuid.New().String(), "-", ""))[:4]
	return prefix + suffix
}

func TestConsumer_shouldDispatch_Debounce(t *testing.T) {
	c := NewConsumer(nil, nil, nil)
	key := uuid.New().String()

	assert.True(t, c.shouldDispatch(key, TypeAssigned, "issue-1"))
	assert.False(t, c.shouldDispatch(key, TypeAssigned, "issue-1"))
	assert.True(t, c.shouldDispatch(key, TypeAssigned, "issue-2"))
}

func TestConsumer_cleanDebounce(t *testing.T) {
	c := NewConsumer(nil, nil, nil)
	key := uuid.New().String()
	c.debounce[key+":"+TypeAssigned+":entity"] = time.Now().Add(-10 * time.Second)
	c.cleanDebounce()
	assert.Empty(t, c.debounce)
}

func TestConsumer_handleAssigned_CreatesNotification(t *testing.T) {
	svc, repo, ctx, assigneeID := newNotificationTestEnv(t)
	actor, err := repo.db.User.Create().
		SetEmail("actor-" + uuid.New().String()[:8] + "@test.com").
		SetName("Actor").
		Save(ctx)
	require.NoError(t, err)

	c := NewConsumer(nil, repo, svc)
	evt := tracker.IssueAssignedEvent{
		IssueID:      uuid.New().String(),
		IssueKey:     "HUB-42",
		ProjectID:    uuid.New().String(),
		AssigneeID:   assigneeID.String(),
		AssignedByID: actor.ID.String(),
	}
	data, err := json.Marshal(evt)
	require.NoError(t, err)

	c.handleAssigned(ctx, data, "project.x.issue.assigned")

	notifs, err := repo.db.Notification.Query().
		Where(notification.UserID(assigneeID)).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, notifs, 1)
	assert.Equal(t, TypeAssigned, notifs[0].Type)
	assert.Contains(t, notifs[0].Title, "HUB-42")
}

func TestConsumer_handleAssigned_SkipsSelfAssignment(t *testing.T) {
	svc, repo, ctx, assigneeID := newNotificationTestEnv(t)
	c := NewConsumer(nil, repo, svc)

	evt := tracker.IssueAssignedEvent{
		IssueID:      uuid.New().String(),
		IssueKey:     "HUB-1",
		ProjectID:    uuid.New().String(),
		AssigneeID:   assigneeID.String(),
		AssignedByID: assigneeID.String(),
	}
	data, err := json.Marshal(evt)
	require.NoError(t, err)

	c.handleAssigned(ctx, data, "")

	count, err := repo.db.Notification.Query().
		Where(notification.UserID(assigneeID)).
		Count(ctx)
	require.NoError(t, err)
	assert.Zero(t, count)
}

func TestConsumer_handleNewLogin_CreatesNotification(t *testing.T) {
	svc, repo, ctx, userID := newNotificationTestEnv(t)
	c := NewConsumer(nil, repo, svc)

	evt := auth.SessionCreatedEvent{
		SessionID: uuid.New().String(),
		UserID:    userID.String(),
		Device:    "Chrome",
		IPAddress: "127.0.0.1",
	}
	data, err := json.Marshal(evt)
	require.NoError(t, err)

	c.handleNewLogin(ctx, data)

	notifs, err := repo.db.Notification.Query().
		Where(notification.UserID(userID)).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, notifs, 1)
	assert.Equal(t, TypeNewLogin, notifs[0].Type)
	assert.Contains(t, notifs[0].Body, "127.0.0.1")
}

func newConsumerIssueEnv(t *testing.T) (*Consumer, *Repository, context.Context, *ent.User, *ent.User, *ent.Issue, uuid.UUID) {
	t.Helper()

	svc, repo, ctx, assigneeID := newNotificationTestEnv(t)
	assignee, err := repo.db.User.Get(ctx, assigneeID)
	require.NoError(t, err)

	actor, err := repo.db.User.Create().
		SetEmail("actor-" + uuid.New().String()[:8] + "@test.com").
		SetName("Actor").
		Save(ctx)
	require.NoError(t, err)

	proj, err := repo.db.Project.Create().
		SetName("Notify Proj").
		SetKey(testNotifyProjectKey("N")).
		Save(ctx)
	require.NoError(t, err)

	status, err := repo.db.WorkflowStatus.Create().
		SetName("Todo").
		SetCategory("todo").
		SetProjectID(proj.ID).
		SetIsDefault(true).
		Save(ctx)
	require.NoError(t, err)

	issueEnt, err := repo.db.Issue.Create().
		SetProjectID(proj.ID).
		SetStatusID(status.ID).
		SetReporterID(actor.ID).
		SetAssigneeID(assignee.ID).
		SetIssueNumber(1).
		SetIssueKey(proj.Key+"-1").
		SetTitle("Assigned task").
		SetIssueType(issue.IssueTypeTask).
		SetPriority(issue.PriorityMedium).
		Save(ctx)
	require.NoError(t, err)

	c := NewConsumer(nil, repo, svc)
	return c, repo, ctx, assignee, actor, issueEnt, proj.ID
}

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

	c.handleStatusChanged(ctx, data, "")

	notifs, err := repo.db.Notification.Query().
		Where(notification.UserID(assignee.ID)).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, notifs, 1)
	assert.Equal(t, TypeStatusChanged, notifs[0].Type)
}

func TestConsumer_handleCommentAdded_NotifiesAssigneeAndMention(t *testing.T) {
	c, repo, ctx, assignee, actor, issueEnt, projectID := newConsumerIssueEnv(t)

	mention, err := repo.db.User.Create().
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

	c.handleCommentAdded(ctx, data, "")

	assigneeNotifs, err := repo.db.Notification.Query().
		Where(notification.UserID(assignee.ID)).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, assigneeNotifs, 1)
	assert.Equal(t, TypeCommentAdded, assigneeNotifs[0].Type)

	mentionNotifs, err := repo.db.Notification.Query().
		Where(notification.UserID(mention.ID)).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, mentionNotifs, 1)
	assert.Equal(t, TypeMentioned, mentionNotifs[0].Type)
}

func TestConsumer_handleRunCompleted_NotifiesProjectMember(t *testing.T) {
	svc, repo, ctx, ownerID := newNotificationTestEnv(t)
	completer, err := repo.db.User.Create().
		SetEmail("complete-" + uuid.New().String()[:8] + "@test.com").
		SetName("Completer").
		Save(ctx)
	require.NoError(t, err)

	proj, err := repo.db.Project.Create().
		SetName("TMS Proj").
		SetKey(testNotifyProjectKey("T")).
		Save(ctx)
	require.NoError(t, err)

	_, err = repo.db.ProjectMember.Create().
		SetProjectID(proj.ID).
		SetUserID(ownerID).
		SetProjectRole("developer").
		Save(ctx)
	require.NoError(t, err)

	c := NewConsumer(nil, repo, svc)
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

	c.handleRunCompleted(ctx, data, "")

	notifs, err := repo.db.Notification.Query().
		Where(notification.UserID(ownerID)).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, notifs, 1)
	assert.Equal(t, TypeRunCompleted, notifs[0].Type)
}
