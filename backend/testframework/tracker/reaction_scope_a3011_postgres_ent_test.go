package tracker_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/tracker"
)

// TestEntRepository_ToggleReaction_Scoped_Postgres_A3_011 — Postgres twin sqlite
// TestEntRepository_ToggleReaction_Scoped_A3_011. Реакция через чужой issueID → NotFound.
// Mutation: убрать HasIssueWith → PASS на чужом issue.
func TestEntRepository_ToggleReaction_Scoped_Postgres_A3_011(t *testing.T) {
	env := newTrackerPostgresRepoEnv(t)

	issueA, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "A", IssueType: "task", ReporterID: env.userID,
	})
	require.NoError(t, err)
	issueB, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "B", IssueType: "task", ReporterID: env.userID,
	})
	require.NoError(t, err)

	comment, err := env.repo.CreateComment(env.ctx, issueA.ID, env.userID, tracker.CreateCommentInput{
		BodyJSON: map[string]interface{}{"type": "doc"},
		BodyText: "react me",
	})
	require.NoError(t, err)

	_, err = env.repo.ToggleReaction(env.ctx, issueB.ID, comment.ID, env.userID, "👍")
	require.True(t, ent.IsNotFound(err), "wrong issue must 404, got %v", err)

	added, err := env.repo.ToggleReaction(env.ctx, issueA.ID, comment.ID, env.userID, "👍")
	require.NoError(t, err)
	require.True(t, added)
}
