package tracker_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/tracker"
)

// TestEntRepository_LogTime_DoesNotMatchStopTimer_A3_005 —
// manual LogTime пишет StartedAt+EndedAt; StopTimer не матчит manual и не паникует.
// Mutation: убрать StartedAtNotNil / EndedAt из LogTime → StopTimer может взять manual → panic risk.
func TestEntRepository_LogTime_DoesNotMatchStopTimer_A3_005(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	issue, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "Time log", IssueType: "task", ReporterID: env.userID,
	})
	require.NoError(t, err)

	logged, err := env.repo.LogTime(env.ctx, issue.ID, env.userID, 30, "manual")
	require.NoError(t, err)
	require.NotNil(t, logged.StartedAt)
	require.NotNil(t, logged.EndedAt)
	assert.Equal(t, 30, logged.Minutes)

	_, err = env.repo.StopTimer(env.ctx, issue.ID, env.userID)
	require.Error(t, err, "manual LogTime must not be an open timer")
	assert.Contains(t, err.Error(), "активный таймер не найден")

	started, err := env.repo.StartTimer(env.ctx, issue.ID, env.userID)
	require.NoError(t, err)
	require.NotNil(t, started.StartedAt)
	require.Nil(t, started.EndedAt)

	stopped, err := env.repo.StopTimer(env.ctx, issue.ID, env.userID)
	require.NoError(t, err)
	require.NotNil(t, stopped.EndedAt)
}
