package tracker_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/tracker"
)

// TestEntRepository_DeleteIssueLink_Scoped_A3_011 — чужой source issue → NotFound (sqlite Ent).
// Postgres twin: TestEntRepository_DeleteIssueLink_Scoped_Postgres_A3_011.
// Mutation: снять HasSourceWith → PASS на чужом source.
func TestEntRepository_DeleteIssueLink_Scoped_A3_011(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	source, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "Link source", IssueType: "task", ReporterID: env.userID,
	})
	require.NoError(t, err)
	target, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "Link target", IssueType: "task", ReporterID: env.userID,
	})
	require.NoError(t, err)
	foreign, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "Foreign source", IssueType: "task", ReporterID: env.userID,
	})
	require.NoError(t, err)

	link, err := env.repo.CreateIssueLink(env.ctx, source.ID, target.ID, "blocks", "")
	require.NoError(t, err)

	err = env.repo.DeleteIssueLink(env.ctx, foreign.ID, link.ID)
	require.True(t, ent.IsNotFound(err), "foreign source must 404, got %v", err)

	links, err := env.repo.ListIssueLinks(env.ctx, source.ID)
	require.NoError(t, err)
	require.Len(t, links, 1, "link must survive foreign delete")

	require.NoError(t, env.repo.DeleteIssueLink(env.ctx, source.ID, link.ID))
	links, err = env.repo.ListIssueLinks(env.ctx, source.ID)
	require.NoError(t, err)
	require.Empty(t, links)
}
