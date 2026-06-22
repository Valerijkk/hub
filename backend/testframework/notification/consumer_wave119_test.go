package notification_test

// Тесты NATS consumer уведомлений (волна 119): интеграционные и unit-сценарии testframework.

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

// TestConsumer_handleCommentAdded_InvalidUUIDFields проверяет отклонение при consumer handle comment added некорректный uuid fields.
func TestConsumer_handleCommentAdded_InvalidUUIDFields(t *testing.T) {
	c, repo, ctx, assignee, actor, issueEnt, projectID := newConsumerIssueEnv(t)

	evt := tracker.CommentCreatedEvent{
		CommentID: uuid.New().String(),
		IssueID:   issueEnt.ID.String(),
		IssueKey:  issueEnt.IssueKey,
		ProjectID: "bad-project",
		AuthorID:  actor.ID.String(),
	}
	data, err := json.Marshal(evt)
	require.NoError(t, err)
	notification.ExportHandleCommentAdded(c, ctx, data, "")

	evt.ProjectID = projectID.String()
	evt.AuthorID = "bad-author"
	data, err = json.Marshal(evt)
	require.NoError(t, err)
	notification.ExportHandleCommentAdded(c, ctx, data, "")

	evt.AuthorID = actor.ID.String()
	evt.IssueID = "bad-issue"
	data, err = json.Marshal(evt)
	require.NoError(t, err)
	notification.ExportHandleCommentAdded(c, ctx, data, "")

	count, err := notification.ExportRepositoryDB(repo).Notification.Query().
		Where(entnotif.UserID(assignee.ID)).
		Count(ctx)
	require.NoError(t, err)
	assert.Zero(t, count)
}

// TestConsumer_handleCommentAdded_NoAssignee проверяет сценарий: consumer handle comment added no assignee.
func TestConsumer_handleCommentAdded_NoAssignee(t *testing.T) {
	c, repo, ctx, assignee, actor, issueEnt, projectID := newConsumerIssueEnv(t)

	_, err := issueEnt.Update().ClearAssignee().Save(ctx)
	require.NoError(t, err)

	evt := tracker.CommentCreatedEvent{
		CommentID: uuid.New().String(),
		IssueID:   issueEnt.ID.String(),
		IssueKey:  issueEnt.IssueKey,
		ProjectID: projectID.String(),
		AuthorID:  actor.ID.String(),
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

// TestConsumer_handleCommentAdded_SkipsSelfMention проверяет сценарий: consumer handle comment added skips self mention.
func TestConsumer_handleCommentAdded_SkipsSelfMention(t *testing.T) {
	c, repo, ctx, assignee, actor, issueEnt, projectID := newConsumerIssueEnv(t)

	evt := tracker.CommentCreatedEvent{
		CommentID: uuid.New().String(),
		IssueID:   issueEnt.ID.String(),
		IssueKey:  issueEnt.IssueKey,
		ProjectID: projectID.String(),
		AuthorID:  actor.ID.String(),
		Mentions:  []string{actor.ID.String(), "not-a-uuid"},
	}
	data, err := json.Marshal(evt)
	require.NoError(t, err)
	notification.ExportHandleCommentAdded(c, ctx, data, "")

	mentionCount, err := notification.ExportRepositoryDB(repo).Notification.Query().
		Where(entnotif.Type(notification.TypeMentioned)).
		Count(ctx)
	require.NoError(t, err)
	assert.Zero(t, mentionCount)

	assigneeNotifs, err := notification.ExportRepositoryDB(repo).Notification.Query().
		Where(entnotif.UserID(assignee.ID)).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, assigneeNotifs, 1)
	assert.Equal(t, notification.TypeCommentAdded, assigneeNotifs[0].Type)
}

// TestConsumer_handleBlockerAdded_InvalidUUIDFields проверяет отклонение при consumer handle blocker added некорректный uuid fields.
func TestConsumer_handleBlockerAdded_InvalidUUIDFields(t *testing.T) {
	c, repo, ctx, assignee, _, _, projectID := newConsumerIssueEnv(t)

	payload := map[string]string{
		"source_id":  uuid.New().String(),
		"target_id":  "bad-target",
		"link_type":  "blocks",
		"project_id": projectID.String(),
		"created_by": uuid.New().String(),
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)
	notification.ExportHandleBlockerAdded(c, ctx, data, "")

	payload["target_id"] = uuid.New().String()
	payload["created_by"] = "bad-actor"
	data, err = json.Marshal(payload)
	require.NoError(t, err)
	notification.ExportHandleBlockerAdded(c, ctx, data, "")

	payload["created_by"] = uuid.New().String()
	payload["project_id"] = "bad-project"
	data, err = json.Marshal(payload)
	require.NoError(t, err)
	notification.ExportHandleBlockerAdded(c, ctx, data, "")

	count, err := notification.ExportRepositoryDB(repo).Notification.Query().
		Where(entnotif.UserID(assignee.ID)).
		Count(ctx)
	require.NoError(t, err)
	assert.Zero(t, count)
}

// TestConsumer_handleBlockerAdded_NoAssigneeOnTarget проверяет сценарий: consumer handle blocker added no assignee on target.
func TestConsumer_handleBlockerAdded_NoAssigneeOnTarget(t *testing.T) {
	svc, repo, ctx, _ := newNotificationTestEnv(t)

	actor, err := notification.ExportRepositoryDB(repo).User.Create().
		SetEmail("blocker-" + uuid.New().String()[:8] + "@test.com").
		SetName("Blocker").
		Save(ctx)
	require.NoError(t, err)

	proj, err := notification.ExportRepositoryDB(repo).Project.Create().
		SetName("Block Proj").
		SetKey(testNotifyProjectKey("B")).
		Save(ctx)
	require.NoError(t, err)

	status, err := notification.ExportRepositoryDB(repo).WorkflowStatus.Create().
		SetName("Todo").
		SetCategory("todo").
		SetProjectID(proj.ID).
		SetIsDefault(true).
		Save(ctx)
	require.NoError(t, err)

	target, err := notification.ExportRepositoryDB(repo).Issue.Create().
		SetProjectID(proj.ID).
		SetStatusID(status.ID).
		SetReporterID(actor.ID).
		SetIssueNumber(3).
		SetIssueKey(proj.Key + "-3").
		SetTitle("Blocked without assignee").
		SetIssueType(issue.IssueTypeTask).
		SetPriority(issue.PriorityMedium).
		Save(ctx)
	require.NoError(t, err)

	c := notification.NewConsumer(nil, repo, svc)
	payload := map[string]string{
		"source_id":  uuid.New().String(),
		"target_id":  target.ID.String(),
		"link_type":  "blocks",
		"project_id": proj.ID.String(),
		"created_by": actor.ID.String(),
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)
	notification.ExportHandleBlockerAdded(c, ctx, data, "")

	count, err := notification.ExportRepositoryDB(repo).Notification.Query().Count(ctx)
	require.NoError(t, err)
	assert.Zero(t, count)
}

// TestConsumer_processEvent_RunCompletedAndBlocker проверяет сценарий: consumer process event run completed and blocker.
func TestConsumer_processEvent_RunCompletedAndBlocker(t *testing.T) {
	c, repo, ctx, assignee, actor, issueEnt, projectID := newConsumerIssueEnv(t)

	runEvt := tms.TestRunCompletedEvent{
		RunID:       uuid.New().String(),
		ProjectID:   projectID.String(),
		Title:       "Smoke",
		CompletedBy: actor.ID.String(),
		Passed:      5,
		Failed:      0,
		Skipped:     0,
	}
	runData, err := json.Marshal(runEvt)
	require.NoError(t, err)
	notification.ExportProcessEvent(c, ctx, &stubJetStreamMsg{
		subject: "tms.run.completed." + runEvt.RunID,
		data:    runData,
	})

	blockerData, err := json.Marshal(map[string]string{
		"source_id":  uuid.New().String(),
		"target_id":  issueEnt.ID.String(),
		"link_type":  "blocks",
		"project_id": projectID.String(),
		"created_by": actor.ID.String(),
	})
	require.NoError(t, err)
	notification.ExportProcessEvent(c, ctx, &stubJetStreamMsg{
		subject: "project." + projectID.String() + ".issue.link_created",
		data:    blockerData,
	})

	blockerNotifs, err := notification.ExportRepositoryDB(repo).Notification.Query().
		Where(
			entnotif.UserID(assignee.ID),
			entnotif.Type(notification.TypeBlockerAdded),
		).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, blockerNotifs, 1)
}
