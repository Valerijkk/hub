package tracker_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/tracker"
)

// TestEntRepository_FindIssueInProject_RejectsForeignProject_Postgres_A3_011 —
// Postgres twin sqlite TestEntRepository_FindIssueInProject_RejectsForeignProject_A3_011.
// Issue проекта B не находится через project A.
// Mutation: снять issueProjectIDEQ → PASS на чужом проекте.
func TestEntRepository_FindIssueInProject_RejectsForeignProject_Postgres_A3_011(t *testing.T) {
	env := newTrackerPostgresRepoEnv(t)

	foreign, err := env.client.Project.Create().
		SetName("Foreign Project A3011 Target PG").
		SetKey("FRP").
		Save(env.ctx)
	require.NoError(t, err)

	status, err := env.client.WorkflowStatus.Create().
		SetName("To Do").
		SetCategory("todo").
		SetColor("#6B7280").
		SetSortOrder(0).
		SetIsDefault(true).
		SetProjectID(foreign.ID).
		Save(env.ctx)
	require.NoError(t, err)

	foreignRepo := tracker.NewEntRepository(env.client, nil)
	created, err := foreignRepo.CreateIssue(env.ctx, foreign.ID, tracker.CreateIssueInput{
		Title:      "Secret target",
		IssueType:  "task",
		Priority:   "medium",
		ReporterID: env.userID,
		StatusID:   &status.ID,
	})
	require.NoError(t, err)

	_, err = env.repo.FindIssueInProject(env.ctx, env.projectID, created.ID)
	require.True(t, ent.IsNotFound(err), "foreign issue must not be visible via project A, got %v", err)

	got, err := foreignRepo.FindIssueInProject(env.ctx, foreign.ID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
}
