package tracker_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/timelog"
	"github.com/valerii/hub/backend/ent/user"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
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

// TestEntRepository_StartTimer_RejectsOpenTimer_A3_005 —
// повторный StartTimer при уже открытом таймере → 409; первый остаётся открытым.
func TestEntRepository_StartTimer_RejectsOpenTimer_A3_005(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	created, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "Double start", IssueType: "task", ReporterID: env.userID,
	})
	require.NoError(t, err)

	first, err := env.repo.StartTimer(env.ctx, created.ID, env.userID)
	require.NoError(t, err)
	require.NotNil(t, first.StartedAt)
	require.Nil(t, first.EndedAt)

	_, err = env.repo.StartTimer(env.ctx, created.ID, env.userID)
	require.Error(t, err)

	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr), "expected AppError, got %T: %v", err, err)
	assert.Equal(t, 409, appErr.Code)
	assert.Contains(t, appErr.Message, "таймер уже запущен")

	open, err := env.client.TimeLog.Query().
		Where(
			timelog.IDEQ(first.ID),
			timelog.EndedAtIsNil(),
		).
		Only(env.ctx)
	require.NoError(t, err)
	require.NotNil(t, open.StartedAt)
	require.Nil(t, open.EndedAt)

	openCount, err := env.client.TimeLog.Query().
		Where(
			timelog.HasIssueWith(issue.IDEQ(created.ID)),
			timelog.HasUserWith(user.IDEQ(env.userID)),
			timelog.EndedAtIsNil(),
			timelog.StartedAtNotNil(),
		).
		Count(env.ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, openCount, "exactly one open timer must remain")
}
