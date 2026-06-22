package tracker_test

// Тесты репозитория трекера задач (волна 120): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tracker"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEntRepository_CreateComment_PublishesOutboxWithMentions проверяет слой репозитория: ent репозиторий создание comment publishes outbox with упоминания.
func TestEntRepository_CreateComment_PublishesOutboxWithMentions(t *testing.T) {
	env := newTrackerRepoTestEnv(t)
	pub := &stubOutboxPublisher{}
	env.repo = tracker.NewEntRepository(env.client, pub)

	issue, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "Comment outbox", IssueType: "task", ReporterID: env.userID,
	})
	require.NoError(t, err)

	mention, err := env.client.User.Create().
		SetEmail("mention-" + uuid.New().String()[:8] + "@test.com").
		SetName("M").
		Save(env.ctx)
	require.NoError(t, err)

	comment, err := env.repo.CreateComment(env.ctx, issue.ID, env.userID, tracker.CreateCommentInput{
		BodyJSON: map[string]interface{}{"type": "doc"},
		BodyText: "with mention",
		Mentions: []string{mention.ID.String()},
	})
	require.NoError(t, err)
	assert.Equal(t, "with mention", comment.BodyText)
	assert.Contains(t, pub.events, "project."+env.projectID.String()+".issue.comment_created")
}

// TestAutomationService_DeleteRule_Success проверяет бизнес-логику: автоматизация сервис удаление rule успех.
func TestAutomationService_DeleteRule_Success(t *testing.T) {
	env := newTrackerRepoTestEnv(t)
	svc := tracker.NewAutomationService(env.repo)

	created, err := svc.CreateRule(env.ctx, env.projectID, env.userID, tracker.CreateAutomationRuleRequest{
		Name:        "Delete me",
		TriggerType: tracker.TriggerIssueCreated,
		ActionType:  tracker.ActionNotifyUser,
	})
	require.NoError(t, err)

	ruleID, err := uuid.Parse(created.ID)
	require.NoError(t, err)

	require.NoError(t, svc.DeleteRule(env.ctx, env.projectID, ruleID))

	_, err = svc.UpdateRule(env.ctx, env.projectID, ruleID, tracker.UpdateAutomationRuleRequest{})
	require.Error(t, err)
}
