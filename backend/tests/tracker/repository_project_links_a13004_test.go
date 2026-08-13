//go:build hubtest

package tracker_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/tracker"
)

// TestEntRepository_ListProjectIssueLinks_Batch_A13_004 — один запрос на проект
// вместо N ListIssueLinks (Gantt N+1). Cap MaxProjectIssueLinks.
func TestEntRepository_ListProjectIssueLinks_Batch_A13_004(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	a, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "A", IssueType: "task", ReporterID: env.userID,
	})
	require.NoError(t, err)
	b, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "B", IssueType: "task", ReporterID: env.userID,
	})
	require.NoError(t, err)
	c, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "C", IssueType: "task", ReporterID: env.userID,
	})
	require.NoError(t, err)

	_, err = env.repo.CreateIssueLink(env.ctx, a.ID, b.ID, "blocked_by", "")
	require.NoError(t, err)
	_, err = env.repo.CreateIssueLink(env.ctx, b.ID, c.ID, "blocks", "")
	require.NoError(t, err)

	// Before: N per-issue queries would be 3× ListIssueLinks (6 SQL with in/out).
	// After: 1 ListProjectIssueLinks.
	links, err := env.repo.ListProjectIssueLinks(env.ctx, env.projectID)
	require.NoError(t, err)
	require.Len(t, links, 2)
	assert.LessOrEqual(t, len(links), tracker.MaxProjectIssueLinks)

	svc := tracker.NewTrackerService(env.repo, nil)
	resp, err := svc.ListProjectIssueLinks(env.ctx, env.projectID)
	require.NoError(t, err)
	require.Len(t, resp, 2)

	keys := map[string]bool{}
	for _, item := range resp {
		keys[item.SourceIssueKey] = true
		keys[item.TargetIssueKey] = true
		assert.NotEmpty(t, item.SourceIssueKey)
		assert.NotEmpty(t, item.TargetIssueKey)
	}
	assert.True(t, keys[a.IssueKey])
	assert.True(t, keys[b.IssueKey])
	assert.True(t, keys[c.IssueKey])
}
