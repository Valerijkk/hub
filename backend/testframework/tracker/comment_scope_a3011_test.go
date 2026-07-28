//go:build hubtest

package tracker_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/tracker"
)

// TestEntRepository_UpdateDeleteComment_Scoped_A3_011 — чужой issue/author → NotFound.
// Mutation: UpdateOneID/DeleteOneID без Where → PASS на чужом issue/author.
func TestEntRepository_UpdateDeleteComment_Scoped_A3_011(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	issueA, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "A", IssueType: "task", ReporterID: env.userID,
	})
	require.NoError(t, err)
	issueB, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "B", IssueType: "task", ReporterID: env.userID,
	})
	require.NoError(t, err)

	otherUser, err := env.client.User.Create().
		SetEmail("a3011-" + uuid.New().String()[:8] + "@test.com").
		SetName("Other").
		Save(env.ctx)
	require.NoError(t, err)

	comment, err := env.repo.CreateComment(env.ctx, issueA.ID, env.userID, tracker.CreateCommentInput{
		BodyJSON: map[string]interface{}{"type": "doc"},
		BodyText: "mine",
	})
	require.NoError(t, err)

	_, err = env.repo.UpdateComment(env.ctx, issueB.ID, env.userID, comment.ID, tracker.UpdateCommentInput{
		BodyJSON: map[string]interface{}{"type": "doc"},
		BodyText: "hijack",
		Version:  1,
	})
	require.True(t, ent.IsNotFound(err), "wrong issue must 404, got %v", err)

	_, err = env.repo.UpdateComment(env.ctx, issueA.ID, otherUser.ID, comment.ID, tracker.UpdateCommentInput{
		BodyJSON: map[string]interface{}{"type": "doc"},
		BodyText: "hijack",
		Version:  1,
	})
	require.True(t, ent.IsNotFound(err), "wrong author must 404, got %v", err)

	err = env.repo.DeleteComment(env.ctx, issueB.ID, env.userID, comment.ID)
	require.True(t, ent.IsNotFound(err), "wrong issue delete must 404, got %v", err)

	err = env.repo.DeleteComment(env.ctx, issueA.ID, otherUser.ID, comment.ID)
	require.True(t, ent.IsNotFound(err), "wrong author delete must 404, got %v", err)

	require.NoError(t, env.repo.DeleteComment(env.ctx, issueA.ID, env.userID, comment.ID))
}
