//go:build hubtest

package sprint_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sprintEnt "github.com/valerii/hub/backend/ent/sprint"
	"github.com/valerii/hub/backend/internal/sprint"
)

// TestSprintMetrics_ExcludeSoftDeletedIssues_A3W1A4002 — count/burndown/velocity/detail без trash.
// Mutation: убрать DeletedAtIsNil из count/burndown/velocity/WithIssues → trash inflates → RED.
func TestSprintMetrics_ExcludeSoftDeletedIssues_A3W1A4002(t *testing.T) {
	env := newSprintRepoEnv(t)
	start := time.Now().UTC().AddDate(0, -1, 0)
	end := start.AddDate(0, 0, 14)

	sp, err := env.repo.CreateSprint(env.ctx, env.projectID, sprint.CreateSprintRequest{
		Name: "Metrics", StartDate: start, EndDate: end,
	})
	require.NoError(t, err)

	alive := env.createIssue(t, 1, env.todoID, &sp.ID)
	done := env.createIssue(t, 2, env.doneID, &sp.ID)
	trash := env.createIssue(t, 3, env.todoID, &sp.ID)
	_, err = env.client.Issue.UpdateOneID(trash.ID).
		SetDeletedAt(time.Now().UTC()).
		Save(env.ctx)
	require.NoError(t, err)

	count, err := env.repo.GetSprintIssueCount(env.ctx, sp.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, count, "A3W1-A4-002: GetSprintIssueCount без trash")

	counts, err := env.repo.GetSprintIssueCounts(env.ctx, []uuid.UUID{sp.ID})
	require.NoError(t, err)
	assert.Equal(t, 2, counts[sp.ID])

	total, doneCnt, err := env.repo.GetBurndownData(env.ctx, sp.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Equal(t, 1, doneCnt)

	_, err = env.repo.UpdateSprintStatus(env.ctx, sp.ID, sprintEnt.StatusCompleted)
	require.NoError(t, err)

	points, err := env.repo.GetVelocityData(env.ctx, env.projectID)
	require.NoError(t, err)
	require.Len(t, points, 1)
	assert.Equal(t, 2, points[0].Committed)
	assert.Equal(t, 1, points[0].Completed)

	detail, err := env.repo.GetSprint(env.ctx, sp.ID)
	require.NoError(t, err)
	require.Len(t, detail.Edges.Issues, 2)
	ids := map[uuid.UUID]bool{}
	for _, iss := range detail.Edges.Issues {
		ids[iss.ID] = true
	}
	assert.True(t, ids[alive.ID])
	assert.True(t, ids[done.ID])
	assert.False(t, ids[trash.ID], "A3W1-A4-002: WithIssues без soft-deleted")
}
