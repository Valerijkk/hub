package tracker_test

// Тесты репозитория трекера задач: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tracker"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEntRepository_CommentsAndReactions проверяет слой репозитория: ent репозиторий comments and reactions.
func TestEntRepository_CommentsAndReactions(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	issue, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "Comment me", IssueType: "task", ReporterID: env.userID,
	})
	require.NoError(t, err)

	comment, err := env.repo.CreateComment(env.ctx, issue.ID, env.userID, tracker.CreateCommentInput{
		BodyJSON: map[string]interface{}{"type": "doc"},
		BodyText: "hello",
	})
	require.NoError(t, err)
	assert.Equal(t, "hello", comment.BodyText)

	comments, err := env.repo.ListComments(env.ctx, issue.ID)
	require.NoError(t, err)
	require.Len(t, comments, 1)

	updatedText := "updated"
	updated, err := env.repo.UpdateComment(env.ctx, issue.ID, env.userID, comment.ID, tracker.UpdateCommentInput{
		BodyJSON: map[string]interface{}{"type": "doc", "v": 2},
		BodyText: updatedText,
	})
	require.NoError(t, err)
	assert.Equal(t, updatedText, updated.BodyText)

	added, err := env.repo.ToggleReaction(env.ctx, issue.ID, comment.ID, env.userID, "👍")
	require.NoError(t, err)
	assert.True(t, added)

	removed, err := env.repo.ToggleReaction(env.ctx, issue.ID, comment.ID, env.userID, "👍")
	require.NoError(t, err)
	assert.False(t, removed)

	require.NoError(t, env.repo.DeleteComment(env.ctx, issue.ID, env.userID, comment.ID))
	comments, err = env.repo.ListComments(env.ctx, issue.ID)
	require.NoError(t, err)
	assert.Empty(t, comments)
}

// TestEntRepository_CustomFields проверяет слой репозитория: ent репозиторий custom fields.
func TestEntRepository_CustomFields(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	field, err := env.repo.CreateCustomField(env.ctx, env.projectID, tracker.CreateCustomFieldInput{
		Name: "Severity", FieldType: "select", Options: []string{"S1", "S2"}, Required: true,
	})
	require.NoError(t, err)

	fields, err := env.repo.ListCustomFields(env.ctx, env.projectID)
	require.NoError(t, err)
	require.Len(t, fields, 1)

	require.NoError(t, env.repo.DeleteCustomField(env.ctx, field.ID))
	fields, err = env.repo.ListCustomFields(env.ctx, env.projectID)
	require.NoError(t, err)
	assert.Empty(t, fields)
}

// TestEntRepository_IssueMessageLinks проверяет слой репозитория: ent репозиторий issue message links.
func TestEntRepository_IssueMessageLinks(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	issue, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "Linked", IssueType: "task", ReporterID: env.userID,
	})
	require.NoError(t, err)

	ch, err := env.client.Channel.Create().
		SetName("dev").
		SetChannelType("public").
		SetProjectID(env.projectID).
		SetCreatedBy(env.userID).
		Save(env.ctx)
	require.NoError(t, err)

	link, err := env.repo.CreateIssueMessageLink(env.ctx, issue.ID, ch.ID, env.userID, "msg-1", true)
	require.NoError(t, err)

	byIssue, err := env.repo.ListIssueMessageLinksByIssue(env.ctx, issue.ID)
	require.NoError(t, err)
	require.Len(t, byIssue, 1)

	byThread, err := env.repo.ListIssueMessageLinksByThread(env.ctx, ch.ID, "msg-1")
	require.NoError(t, err)
	require.Len(t, byThread, 1)

	found, err := env.repo.FindIssueMessageLink(env.ctx, link.ID)
	require.NoError(t, err)
	assert.Equal(t, link.ID, found.ID)

	require.NoError(t, env.repo.DeleteIssueMessageLink(env.ctx, link.ID, issue.ID))
	byIssue, err = env.repo.ListIssueMessageLinksByIssue(env.ctx, issue.ID)
	require.NoError(t, err)
	assert.Empty(t, byIssue)
}

// TestEntRepository_WorkflowTransitionsAndDefault проверяет слой репозитория: ent репозиторий workflow transitions and default.
func TestEntRepository_WorkflowTransitionsAndDefault(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	defaultStatus, err := env.repo.FindDefaultStatus(env.ctx, env.projectID)
	require.NoError(t, err)
	assert.True(t, defaultStatus.IsDefault)

	done, err := env.repo.CreateWorkflowStatus(env.ctx, env.projectID, "Review", "in_progress", "#AAA", 5)
	require.NoError(t, err)

	newName := "Code Review"
	updated, err := env.repo.UpdateWorkflowStatus(env.ctx, done.ID, &newName, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, newName, updated.Name)

	tr, err := env.repo.CreateWorkflowTransition(env.ctx, env.statusID, done.ID)
	require.NoError(t, err)

	transitions, err := env.repo.ListWorkflowTransitions(env.ctx, env.projectID)
	require.NoError(t, err)
	assert.NotEmpty(t, transitions)

	_, err = env.repo.FindTransition(env.ctx, env.statusID, done.ID)
	require.NoError(t, err)

	require.NoError(t, env.repo.DeleteWorkflowTransition(env.ctx, tr.ID))
}

// TestEntRepository_DeleteAndMoveIssue проверяет слой репозитория: ent репозиторий удаление and move issue.
func TestEntRepository_DeleteAndMoveIssue(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	issue, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "Move me", IssueType: "task", ReporterID: env.userID,
	})
	require.NoError(t, err)

	done, err := env.repo.CreateWorkflowStatus(env.ctx, env.projectID, "Done2", "done", "#0F0", 50)
	require.NoError(t, err)

	moved, err := env.repo.MoveIssue(env.ctx, issue.ID, done.ID, "rank-1")
	require.NoError(t, err)
	assert.Equal(t, done.ID, moved.Edges.Status.ID)

	require.NoError(t, env.repo.DeleteIssue(env.ctx, issue.ID))
	_, err = env.repo.FindIssueByID(env.ctx, issue.ID)
	require.Error(t, err)
}

// TestEntRepository_GetBoardDataWithFilters проверяет слой репозитория: ent репозиторий получение board data with filters.
func TestEntRepository_GetBoardDataWithFilters(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	issue, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "Board task", IssueType: "task", Priority: "high", ReporterID: env.userID,
	})
	require.NoError(t, err)
	_ = issue

	board, err := env.repo.CreateBoard(env.ctx, env.projectID, "Filtered", map[string]interface{}{
		"issue_type": "task",
		"priority":   "high",
	})
	require.NoError(t, err)

	data, err := env.repo.GetBoardData(env.ctx, board.ID)
	require.NoError(t, err)
	require.NotNil(t, data.Edges.Project)
}

// TestEntRepository_StartStopTimer проверяет слой репозитория: ent репозиторий start stop timer.
func TestEntRepository_StartStopTimer(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	issue, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "Timer", IssueType: "task", ReporterID: env.userID,
	})
	require.NoError(t, err)

	started, err := env.repo.StartTimer(env.ctx, issue.ID, env.userID)
	require.NoError(t, err)
	require.NotNil(t, started.StartedAt)

	stopped, err := env.repo.StopTimer(env.ctx, issue.ID, env.userID)
	require.NoError(t, err)
	require.NotNil(t, stopped.EndedAt)
	assert.GreaterOrEqual(t, stopped.Minutes, 1)
}

// TestEntRepository_GetDB проверяет слой репозитория: ent репозиторий получение db.
func TestEntRepository_GetDB(t *testing.T) {
	env := newTrackerRepoTestEnv(t)
	assert.NotNil(t, env.repo.GetDB())
}
