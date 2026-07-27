//go:build hubtest

package sprint_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/internal/sprint"
)

// TestListSprints_DoesNotEagerLoadIssues_A13_006 — ListSprints без WithIssues;
// issue_count через batch/Count (см. list_batch_counts_a13006_test.go).
// Mutation: вернуть WithIssues() → Edges.Issues != nil FAIL.
func TestListSprints_DoesNotEagerLoadIssues_A13_006(t *testing.T) {
	env := newSprintRepoEnv(t)
	start := time.Now().UTC().Truncate(time.Second)
	end := start.AddDate(0, 0, 14)

	sp, err := env.repo.CreateSprint(env.ctx, env.projectID, sprint.CreateSprintRequest{
		Name: "List Cap", StartDate: start, EndDate: end,
	})
	require.NoError(t, err)

	iss, err := env.client.Issue.Create().
		SetProjectID(env.projectID).
		SetStatusID(env.todoID).
		SetReporterID(env.userID).
		SetIssueNumber(1).
		SetIssueKey("SP-1").
		SetTitle("In sprint").
		SetIssueType(issue.IssueTypeTask).
		SetPriority(issue.PriorityMedium).
		SetSprintID(sp.ID).
		Save(env.ctx)
	require.NoError(t, err)
	_ = iss

	list, err := env.repo.ListSprints(env.ctx, env.projectID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Nil(t, list[0].Edges.Issues, "ListSprints must not eager-load issues")

	cnt, err := env.repo.GetSprintIssueCount(env.ctx, sp.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, cnt)

	// GetSprint still loads issues but capped.
	detail, err := env.repo.GetSprint(env.ctx, sp.ID)
	require.NoError(t, err)
	require.NotNil(t, detail.Edges.Issues)
	assert.LessOrEqual(t, len(detail.Edges.Issues), sprint.MaxSprintIssues)
	assert.Len(t, detail.Edges.Issues, 1)

	svc := sprint.NewSprintService(env.repo, nil, nil)
	resp, err := svc.ListSprints(env.ctx, env.projectID)
	require.NoError(t, err)
	require.Len(t, resp, 1)
	assert.Equal(t, 1, resp[0].IssueCount, "service uses Count when edges empty")
}
