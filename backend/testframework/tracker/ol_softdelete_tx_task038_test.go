//go:build hubtest

package tracker_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
	"github.com/valerii/hub/backend/internal/tracker"
)

// TestEntRepository_MoveIssue_OptimisticLock_W1A3004 — concurrent move с одним version:
// второй → 409. Mutation: убрать VersionEQ из MoveIssue → оба Move 200 → FAIL.
func TestEntRepository_MoveIssue_OptimisticLock_W1A3004(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	done, err := env.repo.CreateWorkflowStatus(env.ctx, env.projectID, "Done OL", "done", "#22C55E", 2)
	require.NoError(t, err)

	issue, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "Move lock", IssueType: "task", ReporterID: env.userID, StatusID: &env.statusID,
	})
	require.NoError(t, err)
	v := issue.Version

	moved, err := env.repo.MoveIssue(env.ctx, issue.ID, done.ID, "rank-a", v)
	require.NoError(t, err)
	assert.Equal(t, v+1, moved.Version)

	_, err = env.repo.MoveIssue(env.ctx, issue.ID, env.statusID, "rank-b", v) // stale
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 409, appErr.Code)
}

// TestEntRepository_UpdateIssue_SoftDeleted_W1A3005 — Update soft-deleted → 404.
// Mutation: снять DeletedAtIsNil из UpdateIssue Where → 409/200 → FAIL.
func TestEntRepository_UpdateIssue_SoftDeleted_W1A3005(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	issue, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "Trash me", IssueType: "task", ReporterID: env.userID,
	})
	require.NoError(t, err)
	require.NoError(t, env.repo.DeleteIssue(env.ctx, issue.ID))

	title := "should-not-apply"
	_, err = env.repo.UpdateIssue(env.ctx, issue.ID, tracker.UpdateIssueInput{
		Title: &title, Version: issue.Version,
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 404, appErr.Code)
}

// TestEntRepository_LogTime_AtomicTx_W1A3006 — LogTime пишет TimeLog и сумму в одной tx.
// Mutation: разнести create+AddTimeSpent без tx → при fail 2-го шага orphan log.
func TestEntRepository_LogTime_AtomicTx_W1A3006(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	issue, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "Time", IssueType: "task", ReporterID: env.userID,
	})
	require.NoError(t, err)

	tl, err := env.repo.LogTime(env.ctx, issue.ID, env.userID, 15, "work")
	require.NoError(t, err)
	assert.Equal(t, 15, tl.Minutes)

	got, err := env.repo.FindIssueByID(env.ctx, issue.ID)
	require.NoError(t, err)
	assert.Equal(t, 15, got.TimeSpentMinutes)

	logs, err := env.repo.ListTimeLogs(env.ctx, issue.ID)
	require.NoError(t, err)
	require.Len(t, logs, 1)
	assert.WithinDuration(t, time.Now(), logs[0].CreatedAt, time.Minute)
}
