package tracker

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

func TestListWorkflowStatuses(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()
	projectID := uuid.New()

	list, err := svc.ListWorkflowStatuses(ctx, projectID)
	require.NoError(t, err)
	assert.NotEmpty(t, list)
	assert.Equal(t, "To Do", list[0].Name)

	statusID := repo.addStatusWithTransition("In Progress", workflowstatus.CategoryInProgress, repo.getDefaultStatusID())
	created, err := svc.CreateWorkflowStatus(ctx, projectID, CreateWorkflowStatusRequest{
		Name:     "Review",
		Category: "done",
		Color:    "#00FF00",
	})
	require.NoError(t, err)
	assert.Equal(t, "Review", created.Name)
	_ = statusID

	require.NoError(t, svc.DeleteWorkflowStatus(ctx, uuid.MustParse(created.ID)))
}

func TestWorkflowTransitions_CRUD(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	fromID := repo.getDefaultStatusID()
	toID := repo.addStatusWithTransition("Doing", workflowstatus.CategoryInProgress, fromID)

	trans, err := svc.CreateWorkflowTransition(ctx, CreateWorkflowTransitionRequest{
		FromStatusID: fromID.String(),
		ToStatusID:   toID.String(),
	})
	require.NoError(t, err)

	list, err := svc.ListWorkflowTransitions(ctx, projectID)
	require.NoError(t, err)
	assert.NotEmpty(t, list)

	require.NoError(t, svc.DeleteWorkflowTransition(ctx, uuid.MustParse(trans.ID)))
}

func TestGetActivity_ReturnsLogs(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	issue, err := svc.CreateIssue(ctx, projectID, userID, CreateIssueRequest{
		Title: "Activity test", IssueType: "task", Priority: "medium",
	})
	require.NoError(t, err)

	entIssue := repo.issuesByKey[issue.IssueKey]
	_, err = repo.LogActivity(ctx, entIssue.ID, userID, "created", nil)
	require.NoError(t, err)

	logs, err := svc.GetActivity(ctx, issue.IssueKey)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(logs), 1)
	assert.Equal(t, "created", logs[0].Action)
}

func TestListIssueLinks_AndDelete(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	source, err := svc.CreateIssue(ctx, projectID, userID, CreateIssueRequest{
		Title: "Source", IssueType: "task", Priority: "medium",
	})
	require.NoError(t, err)

	target, err := svc.CreateIssue(ctx, projectID, userID, CreateIssueRequest{
		Title: "Target", IssueType: "task", Priority: "medium",
	})
	require.NoError(t, err)

	targetIssue := repo.issuesByKey[target.IssueKey]
	link, err := svc.CreateIssueLink(ctx, source.IssueKey, userID, CreateIssueLinkRequest{
		TargetIssueID: targetIssue.ID.String(),
		LinkType:      "blocks",
	})
	require.NoError(t, err)

	links, err := svc.ListIssueLinks(ctx, source.IssueKey)
	require.NoError(t, err)
	require.Len(t, links, 1)
	assert.Equal(t, "blocks", links[0].LinkType)

	require.NoError(t, svc.DeleteIssueLink(ctx, uuid.MustParse(link.ID)))
	links, err = svc.ListIssueLinks(ctx, source.IssueKey)
	require.NoError(t, err)
	assert.Empty(t, links)
}

func TestGetIssue_Success(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	created, err := svc.CreateIssue(ctx, projectID, userID, CreateIssueRequest{
		Title: "Lookup me", IssueType: "bug", Priority: "high",
	})
	require.NoError(t, err)

	got, err := svc.GetIssue(ctx, created.IssueKey)
	require.NoError(t, err)
	assert.Equal(t, "Lookup me", got.Title)
}

func TestLabelGroups_CRUD(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	projectID := uuid.New()

	group, err := svc.CreateLabelGroup(ctx, projectID, CreateLabelGroupRequest{Name: "Components"})
	require.NoError(t, err)

	groups, err := svc.ListLabelGroups(ctx, projectID)
	require.NoError(t, err)
	require.Len(t, groups, 1)
	assert.Equal(t, "Components", groups[0].Name)
	_ = group
}

func TestDeleteLabel(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()
	projectID := uuid.New()

	label, err := svc.CreateLabel(ctx, projectID, CreateLabelRequest{Name: "remove-me", Color: "#000000"})
	require.NoError(t, err)

	labelID := uuid.MustParse(label.ID)
	require.NoError(t, svc.DeleteLabel(ctx, labelID))
	assert.NotContains(t, repo.labels, labelID)
}

func TestListBoards(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	projectID := uuid.New()

	_, err := svc.CreateBoard(ctx, projectID, CreateBoardRequest{Name: "Sprint board"})
	require.NoError(t, err)

	boards, err := svc.ListBoards(ctx, projectID)
	require.NoError(t, err)
	require.Len(t, boards, 1)
	assert.Equal(t, "Sprint board", boards[0].Name)
}

func TestCommentUpdateAndDelete(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	issue, err := svc.CreateIssue(ctx, projectID, userID, CreateIssueRequest{
		Title: "Comments", IssueType: "task", Priority: "medium",
	})
	require.NoError(t, err)

	comment, err := svc.CreateComment(ctx, issue.IssueKey, userID, CreateCommentRequest{
		BodyJSON: map[string]interface{}{"text": "first"},
		BodyText: "first",
	})
	require.NoError(t, err)

	newBody := map[string]interface{}{"text": "updated"}
	updated, err := svc.UpdateComment(ctx, uuid.MustParse(comment.ID), UpdateCommentRequest{
		BodyJSON: newBody,
		BodyText: "updated",
	})
	require.NoError(t, err)
	assert.NotNil(t, updated)

	require.NoError(t, svc.DeleteComment(ctx, uuid.MustParse(comment.ID)))
	assert.Empty(t, repo.comments)
}
