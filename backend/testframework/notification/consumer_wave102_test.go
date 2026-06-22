package notification_test

// Тесты NATS consumer уведомлений (волна 102): интеграционные и unit-сценарии testframework.

import (
	"encoding/json"
	"testing"

	"github.com/valerii/hub/backend/internal/notification"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	entnotif "github.com/valerii/hub/backend/ent/notification"
	"github.com/valerii/hub/backend/internal/auth"
	"github.com/valerii/hub/backend/internal/tracker"
)

// TestConsumer_handleAssigned_InvalidPayload проверяет отклонение при consumer handle assigned некорректный payload.
func TestConsumer_handleAssigned_InvalidPayload(t *testing.T) {
	c, repo, ctx, assignee, _, _, _ := newConsumerIssueEnv(t)

	notification.ExportHandleAssigned(c, ctx, []byte("{"), "")
	notification.ExportHandleAssigned(c, ctx, []byte(`{"assignee_id":"bad","assigned_by_id":"x","project_id":"y","issue_id":"z"}`), "")

	count, err := notification.ExportRepositoryDB(repo).Notification.Query().
		Where(entnotif.UserID(assignee.ID)).
		Count(ctx)
	require.NoError(t, err)
	assert.Zero(t, count)
}

// TestConsumer_handleAssigned_Debounce проверяет сценарий: consumer handle assigned дебаунс.
func TestConsumer_handleAssigned_Debounce(t *testing.T) {
	c, repo, ctx, assignee, actor, issueEnt, projectID := newConsumerIssueEnv(t)

	evt := tracker.IssueAssignedEvent{
		IssueID:      issueEnt.ID.String(),
		IssueKey:     issueEnt.IssueKey,
		ProjectID:    projectID.String(),
		AssigneeID:   assignee.ID.String(),
		AssignedByID: actor.ID.String(),
	}
	data, err := json.Marshal(evt)
	require.NoError(t, err)

	notification.ExportHandleAssigned(c, ctx, data, "")
	notification.ExportHandleAssigned(c, ctx, data, "")

	notifs, err := notification.ExportRepositoryDB(repo).Notification.Query().
		Where(entnotif.UserID(assignee.ID)).
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, notifs, 1)
}

// TestConsumer_handleStatusChanged_NoAssigneeAndSelfMove проверяет сценарий: consumer handle status changed no assignee and self move.
func TestConsumer_handleStatusChanged_NoAssigneeAndSelfMove(t *testing.T) {
	c, repo, ctx, assignee, actor, issueEnt, projectID := newConsumerIssueEnv(t)

	_, err := issueEnt.Update().ClearAssignee().Save(ctx)
	require.NoError(t, err)

	evt := tracker.IssueMovedEvent{
		IssueID:   issueEnt.ID.String(),
		IssueKey:  issueEnt.IssueKey,
		ProjectID: projectID.String(),
		MovedByID: actor.ID.String(),
	}
	data, err := json.Marshal(evt)
	require.NoError(t, err)
	notification.ExportHandleStatusChanged(c, ctx, data, "")

	_, err = issueEnt.Update().SetAssigneeID(actor.ID).Save(ctx)
	require.NoError(t, err)
	evt.MovedByID = actor.ID.String()
	data, err = json.Marshal(evt)
	require.NoError(t, err)
	notification.ExportHandleStatusChanged(c, ctx, data, "")

	count, err := notification.ExportRepositoryDB(repo).Notification.Query().
		Where(entnotif.UserID(assignee.ID)).
		Count(ctx)
	require.NoError(t, err)
	assert.Zero(t, count)
}

// TestConsumer_handleStatusChanged_InvalidJSON проверяет отклонение при consumer handle status changed некорректный json.
func TestConsumer_handleStatusChanged_InvalidJSON(t *testing.T) {
	c, repo, ctx, assignee, _, _, _ := newConsumerIssueEnv(t)
	notification.ExportHandleStatusChanged(c, ctx, []byte("not-json"), "")

	count, err := notification.ExportRepositoryDB(repo).Notification.Query().
		Where(entnotif.UserID(assignee.ID)).
		Count(ctx)
	require.NoError(t, err)
	assert.Zero(t, count)
}

// TestConsumer_handleCommentAdded_InvalidAndAuthorIsAssignee проверяет отклонение при consumer handle comment added некорректный and author is assignee.
func TestConsumer_handleCommentAdded_InvalidAndAuthorIsAssignee(t *testing.T) {
	c, repo, ctx, assignee, actor, issueEnt, projectID := newConsumerIssueEnv(t)

	notification.ExportHandleCommentAdded(c, ctx, []byte("{"), "")

	evt := tracker.CommentCreatedEvent{
		CommentID: uuid.New().String(),
		IssueID:   issueEnt.ID.String(),
		IssueKey:  issueEnt.IssueKey,
		ProjectID: projectID.String(),
		AuthorID:  assignee.ID.String(),
		Mentions:  []string{actor.ID.String()},
	}
	data, err := json.Marshal(evt)
	require.NoError(t, err)
	notification.ExportHandleCommentAdded(c, ctx, data, "")

	count, err := notification.ExportRepositoryDB(repo).Notification.Query().
		Where(entnotif.UserID(assignee.ID)).
		Count(ctx)
	require.NoError(t, err)
	assert.Zero(t, count)
}

// TestConsumer_handleNewLogin_InvalidAndDebounce проверяет отклонение при consumer handle new вход некорректный and дебаунс.
func TestConsumer_handleNewLogin_InvalidAndDebounce(t *testing.T) {
	svc, repo, ctx, userID := newNotificationTestEnv(t)
	c := notification.NewConsumer(nil, repo, svc)

	notification.ExportHandleNewLogin(c, ctx, []byte("{"))
	notification.ExportHandleNewLogin(c, ctx, []byte(`{"user_id":"not-uuid","session_id":"s1"}`))

	evt := auth.SessionCreatedEvent{
		SessionID: uuid.New().String(),
		UserID:    userID.String(),
		Device:    "",
		IPAddress: "10.0.0.1",
	}
	data, err := json.Marshal(evt)
	require.NoError(t, err)

	notification.ExportHandleNewLogin(c, ctx, data)
	notification.ExportHandleNewLogin(c, ctx, data)

	notifs, err := notification.ExportRepositoryDB(repo).Notification.Query().
		Where(entnotif.UserID(userID)).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, notifs, 1)
	assert.Equal(t, notification.TypeNewLogin, notifs[0].Type)
	assert.Contains(t, notifs[0].Body, "неизвестное устройство")
}
