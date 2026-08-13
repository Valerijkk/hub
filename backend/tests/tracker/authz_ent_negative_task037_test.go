//go:build hubtest

package tracker_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/tracker"
)

// TASK-037 / P2-BE-02 — Ent-negative authz для tracker (не mock-only handler_flow).
// Mutation: снять HasAuthorWith / project-scope FindIssueByKey → FAIL.

func TestDeleteComment_RejectsForeignAuthor_Ent_TASK037(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	issue, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "Commented", IssueType: "task", Priority: "medium", ReporterID: env.userID,
	})
	require.NoError(t, err)

	comment, err := env.repo.CreateComment(env.ctx, issue.ID, env.userID, tracker.CreateCommentInput{
		BodyJSON: map[string]interface{}{"type": "doc"},
		BodyText: "owned by reporter",
	})
	require.NoError(t, err)

	outsider, err := env.client.User.Create().
		SetEmail("cmt-out-" + uuid.New().String()[:8] + "@test.local").
		SetName("Outsider").
		Save(env.ctx)
	require.NoError(t, err)

	err = env.repo.DeleteComment(env.ctx, issue.ID, outsider.ID, comment.ID)
	require.True(t, ent.IsNotFound(err), "foreign author must not soft-delete comment, got %v", err)

	still, err := env.client.Comment.Get(env.ctx, comment.ID)
	require.NoError(t, err)
	require.Nil(t, still.DeletedAt)

	require.NoError(t, env.repo.DeleteComment(env.ctx, issue.ID, env.userID, comment.ID))
}

func TestDeleteComment_ServiceRejectsForeignProject_Ent_TASK037(t *testing.T) {
	env := newTrackerRepoTestEnv(t)
	svc := tracker.NewTrackerService(env.repo, nil)

	issue, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "Host", IssueType: "task", Priority: "medium", ReporterID: env.userID,
	})
	require.NoError(t, err)

	comment, err := env.repo.CreateComment(env.ctx, issue.ID, env.userID, tracker.CreateCommentInput{
		BodyJSON: map[string]interface{}{"type": "doc"},
		BodyText: "scoped",
	})
	require.NoError(t, err)

	foreign, err := env.client.Project.Create().
		SetName("Foreign DelCmt").
		SetKey("FDC").
		Save(env.ctx)
	require.NoError(t, err)

	err = svc.DeleteComment(env.ctx, foreign.ID, issue.IssueKey, env.userID, comment.ID)
	require.Error(t, err)
	requireAppErrorCode(t, err, 404)

	still, err := env.client.Comment.Get(env.ctx, comment.ID)
	require.NoError(t, err)
	require.Nil(t, still.DeletedAt)
}

func TestUpdateIssue_RejectsForeignProjectKey_Ent_TASK037(t *testing.T) {
	env := newTrackerRepoTestEnv(t)
	svc := tracker.NewTrackerService(env.repo, nil)

	issue, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "Local", IssueType: "task", Priority: "medium", ReporterID: env.userID,
	})
	require.NoError(t, err)

	foreign, err := env.client.Project.Create().
		SetName("Foreign UpdIss").
		SetKey("FUI").
		Save(env.ctx)
	require.NoError(t, err)

	title := "Hacked"
	_, err = svc.UpdateIssue(env.ctx, foreign.ID, issue.IssueKey, env.userID, tracker.UpdateIssueRequest{
		Title:   &title,
		Version: issue.Version,
	})
	require.Error(t, err)
	requireAppErrorCode(t, err, 404)

	got, err := env.client.Issue.Get(env.ctx, issue.ID)
	require.NoError(t, err)
	require.Equal(t, "Local", got.Title)
}
