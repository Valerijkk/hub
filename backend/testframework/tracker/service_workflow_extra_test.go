package tracker_test

// Тесты сервиса трекера задач: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tracker"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

// TestListWorkflowStatuses проверяет сценарий: список workflow statuses.
func TestListWorkflowStatuses(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()
	projectID := uuid.New()

	list, err := svc.ListWorkflowStatuses(ctx, projectID)
	require.NoError(t, err)
	assert.NotEmpty(t, list)
	assert.Equal(t, "To Do", list[0].Name)

	statusID := repo.addStatusWithTransition("In Progress", workflowstatus.CategoryInProgress, repo.getDefaultStatusID())
	created, err := svc.CreateWorkflowStatus(ctx, projectID, tracker.CreateWorkflowStatusRequest{
		Name:     "Review",
		Category: "done",
		Color:    "#00FF00",
	})
	require.NoError(t, err)
	assert.Equal(t, "Review", created.Name)
	_ = statusID

	require.NoError(t, svc.DeleteWorkflowStatus(ctx, uuid.MustParse(created.ID)))
}

// TestWorkflowTransitions_CRUD проверяет сценарий: workflow transitions crud.
func TestWorkflowTransitions_CRUD(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	fromID := repo.getDefaultStatusID()
	toID := repo.addStatusWithTransition("Doing", workflowstatus.CategoryInProgress, fromID)

	trans, err := svc.CreateWorkflowTransition(ctx, tracker.CreateWorkflowTransitionRequest{
		FromStatusID: fromID.String(),
		ToStatusID:   toID.String(),
	})
	require.NoError(t, err)

	list, err := svc.ListWorkflowTransitions(ctx, projectID)
	require.NoError(t, err)
	assert.NotEmpty(t, list)

	require.NoError(t, svc.DeleteWorkflowTransition(ctx, uuid.MustParse(trans.ID)))
}

// TestGetActivity_ReturnsLogs проверяет сценарий: получение activity returns logs.
func TestGetActivity_ReturnsLogs(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	issue, err := svc.CreateIssue(ctx, projectID, userID, tracker.CreateIssueRequest{
		Title: "Activity test", IssueType: "task", Priority: "medium",
	})
	require.NoError(t, err)

	entIssue := repo.issuesByKey[issue.IssueKey]
	_, err = repo.LogActivity(ctx, entIssue.ID, userID, "created", nil)
	require.NoError(t, err)

	logs, err := svc.GetActivity(ctx, projectID, issue.IssueKey)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(logs), 1)
	assert.Equal(t, "created", logs[0].Action)
}

// TestListIssueLinks_AndDelete проверяет сценарий: список issue links and удаление.
func TestListIssueLinks_AndDelete(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	source, err := svc.CreateIssue(ctx, projectID, userID, tracker.CreateIssueRequest{
		Title: "Source", IssueType: "task", Priority: "medium",
	})
	require.NoError(t, err)

	target, err := svc.CreateIssue(ctx, projectID, userID, tracker.CreateIssueRequest{
		Title: "Target", IssueType: "task", Priority: "medium",
	})
	require.NoError(t, err)

	targetIssue := repo.issuesByKey[target.IssueKey]
	link, err := svc.CreateIssueLink(ctx, projectID, source.IssueKey, userID, tracker.CreateIssueLinkRequest{
		TargetIssueID: targetIssue.ID.String(),
		LinkType:      "blocks",
	})
	require.NoError(t, err)

	links, err := svc.ListIssueLinks(ctx, projectID, source.IssueKey)
	require.NoError(t, err)
	require.Len(t, links, 1)
	assert.Equal(t, "blocks", links[0].LinkType)

	require.NoError(t, svc.DeleteIssueLink(ctx, projectID, source.IssueKey, uuid.MustParse(link.ID)))
	links, err = svc.ListIssueLinks(ctx, projectID, source.IssueKey)
	require.NoError(t, err)
	assert.Empty(t, links)
}

// TestGetIssue_Success проверяет сценарий: получение issue успех.
func TestGetIssue_Success(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	created, err := svc.CreateIssue(ctx, projectID, userID, tracker.CreateIssueRequest{
		Title: "Lookup me", IssueType: "bug", Priority: "high",
	})
	require.NoError(t, err)

	got, err := svc.GetIssue(ctx, projectID, created.IssueKey)
	require.NoError(t, err)
	assert.Equal(t, "Lookup me", got.Title)
}

// TestLabelGroups_CRUD проверяет сценарий: label groups crud.
func TestLabelGroups_CRUD(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	projectID := uuid.New()

	group, err := svc.CreateLabelGroup(ctx, projectID, tracker.CreateLabelGroupRequest{Name: "Components"})
	require.NoError(t, err)

	groups, err := svc.ListLabelGroups(ctx, projectID)
	require.NoError(t, err)
	require.Len(t, groups, 1)
	assert.Equal(t, "Components", groups[0].Name)
	_ = group
}

// TestDeleteLabel проверяет сценарий: удаление label.
func TestDeleteLabel(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()
	projectID := uuid.New()

	label, err := svc.CreateLabel(ctx, projectID, tracker.CreateLabelRequest{Name: "remove-me", Color: "#000000"})
	require.NoError(t, err)

	labelID := uuid.MustParse(label.ID)
	require.NoError(t, svc.DeleteLabel(ctx, projectID, labelID))
	assert.NotContains(t, repo.labels, labelID)
}

// TestListBoards проверяет сценарий: список boards.
func TestListBoards(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	projectID := uuid.New()

	_, err := svc.CreateBoard(ctx, projectID, tracker.CreateBoardRequest{Name: "Sprint board"})
	require.NoError(t, err)

	boards, err := svc.ListBoards(ctx, projectID)
	require.NoError(t, err)
	require.Len(t, boards, 1)
	assert.Equal(t, "Sprint board", boards[0].Name)
}

// TestCommentUpdateAndDelete проверяет сценарий: comment обновление and удаление.
func TestCommentUpdateAndDelete(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	issue, err := svc.CreateIssue(ctx, projectID, userID, tracker.CreateIssueRequest{
		Title: "Comments", IssueType: "task", Priority: "medium",
	})
	require.NoError(t, err)

	comment, err := svc.CreateComment(ctx, projectID, issue.IssueKey, userID, tracker.CreateCommentRequest{
		BodyJSON: map[string]interface{}{"text": "first"},
		BodyText: "first",
	})
	require.NoError(t, err)

	newBody := map[string]interface{}{"text": "updated"}
	updated, err := svc.UpdateComment(ctx, projectID, issue.IssueKey, userID, uuid.MustParse(comment.ID), tracker.UpdateCommentRequest{
		BodyJSON: newBody,
		BodyText: "updated",
	})
	require.NoError(t, err)
	assert.NotNil(t, updated)

	require.NoError(t, svc.DeleteComment(ctx, projectID, issue.IssueKey, userID, uuid.MustParse(comment.ID)))
	assert.Empty(t, repo.comments)
}
