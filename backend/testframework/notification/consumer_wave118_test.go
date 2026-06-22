package notification_test

// Тесты NATS consumer уведомлений (волна 118): интеграционные и unit-сценарии testframework.

import (
	"encoding/json"
	"testing"

	"github.com/valerii/hub/backend/internal/notification"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/issue"
	entnotif "github.com/valerii/hub/backend/ent/notification"
	"github.com/valerii/hub/backend/internal/tms"
	"github.com/valerii/hub/backend/internal/tracker"
)

// TestConsumer_handleStatusChanged_NoAssignee проверяет сценарий: consumer handle status changed no assignee.
func TestConsumer_handleStatusChanged_NoAssignee(t *testing.T) {
	svc, repo, ctx, actorID := newNotificationTestEnv(t)

	actor, err := notification.ExportRepositoryDB(repo).User.Get(ctx, actorID)
	require.NoError(t, err)

	proj, err := notification.ExportRepositoryDB(repo).Project.Create().
		SetName("No Assignee").
		SetKey(testNotifyProjectKey("NA")).
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
		SetIssueNumber(2).
		SetIssueKey(proj.Key + "-2").
		SetTitle("Unassigned").
		SetIssueType(issue.IssueTypeTask).
		SetPriority(issue.PriorityMedium).
		Save(ctx)
	require.NoError(t, err)

	c := notification.NewConsumer(nil, repo, svc)
	evt := tracker.IssueMovedEvent{
		IssueID:   issueEnt.ID.String(),
		IssueKey:  issueEnt.IssueKey,
		ProjectID: proj.ID.String(),
		MovedByID: actor.ID.String(),
	}
	data, err := json.Marshal(evt)
	require.NoError(t, err)

	notification.ExportHandleStatusChanged(c, ctx, data, "")

	count, err := notification.ExportRepositoryDB(repo).Notification.Query().Count(ctx)
	require.NoError(t, err)
	assert.Zero(t, count)
}

// TestConsumer_handleStatusChanged_ActorIsAssignee проверяет сценарий: consumer handle status changed actor is assignee.
func TestConsumer_handleStatusChanged_ActorIsAssignee(t *testing.T) {
	c, repo, ctx, assignee, _, issueEnt, projectID := newConsumerIssueEnv(t)

	evt := tracker.IssueMovedEvent{
		IssueID:   issueEnt.ID.String(),
		IssueKey:  issueEnt.IssueKey,
		ProjectID: projectID.String(),
		MovedByID: assignee.ID.String(),
	}
	data, err := json.Marshal(evt)
	require.NoError(t, err)

	notification.ExportHandleStatusChanged(c, ctx, data, "")

	count, err := notification.ExportRepositoryDB(repo).Notification.Query().
		Where(entnotif.UserID(assignee.ID)).
		Count(ctx)
	require.NoError(t, err)
	assert.Zero(t, count)
}

// TestConsumer_handleCallStarted_ShortSubject проверяет сценарий: consumer handle call started short subject.
func TestConsumer_handleCallStarted_ShortSubject(t *testing.T) {
	svc, repo, ctx, _ := newNotificationTestEnv(t)
	c := notification.NewConsumer(nil, repo, svc)

	notification.ExportHandleCallStarted(c, ctx, []byte(`{"call_id":"x","channel_id":"y","started_by":"`+uuid.New().String()+`"}`), "calls.bad")
}

// TestConsumer_handleMessageReceived_ShortSubject проверяет сценарий: consumer handle message received short subject.
func TestConsumer_handleMessageReceived_ShortSubject(t *testing.T) {
	svc, repo, ctx, _ := newNotificationTestEnv(t)
	c := notification.NewConsumer(nil, repo, svc)

	notification.ExportHandleMessageReceived(c, ctx, []byte(`{"message_id":"m","sender_id":"`+uuid.New().String()+`"}`), "messenger.bad")
}

// TestConsumer_handleMessageReceived_SelfMentionSkipped проверяет сценарий: consumer handle message received self mention skipped.
func TestConsumer_handleMessageReceived_SelfMentionSkipped(t *testing.T) {
	svc, repo, ctx, senderID := newNotificationTestEnv(t)
	c := notification.NewConsumer(nil, repo, svc)

	chID := uuid.New()
	subject := "messenger.channel." + chID.String() + ".message_sent"
	evt := struct {
		MessageID        string   `json:"message_id"`
		ChannelID        string   `json:"channel_id"`
		SenderID         string   `json:"sender_id"`
		ContentText      string   `json:"content_text"`
		MentionedUserIDs []string `json:"mentioned_user_ids"`
	}{
		MessageID:        uuid.New().String(),
		ChannelID:        chID.String(),
		SenderID:         senderID.String(),
		ContentText:      "note",
		MentionedUserIDs: []string{senderID.String()},
	}
	data, err := json.Marshal(evt)
	require.NoError(t, err)

	notification.ExportHandleMessageReceived(c, ctx, data, subject)

	count, err := notification.ExportRepositoryDB(repo).Notification.Query().Count(ctx)
	require.NoError(t, err)
	assert.Zero(t, count)
}

// TestConsumer_handleRunCompleted_InvalidUUIDs проверяет отклонение при consumer handle run completed некорректный uui ds.
func TestConsumer_handleRunCompleted_InvalidUUIDs(t *testing.T) {
	svc, repo, ctx, _ := newNotificationTestEnv(t)
	c := notification.NewConsumer(nil, repo, svc)

	evt := tms.TestRunCompletedEvent{
		RunID:       uuid.New().String(),
		ProjectID:   "not-uuid",
		CompletedBy: uuid.New().String(),
	}
	data, err := json.Marshal(evt)
	require.NoError(t, err)

	notification.ExportHandleRunCompleted(c, ctx, data, "")
}

// TestConsumer_handleCommentAdded_InvalidJSON проверяет отклонение при consumer handle comment added некорректный json.
func TestConsumer_handleCommentAdded_InvalidJSON(t *testing.T) {
	svc, repo, ctx, _ := newNotificationTestEnv(t)
	c := notification.NewConsumer(nil, repo, svc)
	notification.ExportHandleCommentAdded(c, ctx, []byte("{"), "")
}
