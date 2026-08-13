//go:build hubtest

package tracker_test

// TASK-171 / A3W8-5-002 — WorkflowStatus soft-delete + guard при живых issues.
import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/workflowstatus"
	"github.com/valerii/hub/backend/ent/workflowtransition"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
	"github.com/valerii/hub/backend/internal/tracker"
)

// TestDeleteWorkflowStatus_SoftDelete_TASK171 — A3W8-5-002.
// Mutation: вернуть hard Delete() без soft/guard → row gone / list без soft → FAIL.
func TestDeleteWorkflowStatus_SoftDelete_TASK171(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	st, err := env.repo.CreateWorkflowStatus(env.ctx, env.projectID, "TempSoft", "todo", "#abc", 50)
	require.NoError(t, err)

	// Переход from TempSoft → default — должен исчезнуть после soft-delete.
	_, err = env.repo.CreateWorkflowTransition(env.ctx, env.projectID, st.ID, env.statusID)
	require.NoError(t, err)

	require.NoError(t, env.repo.DeleteWorkflowStatus(env.ctx, env.projectID, st.ID))

	got, err := env.client.WorkflowStatus.Get(env.ctx, st.ID)
	require.NoError(t, err, "row must remain after soft-delete")
	require.NotNil(t, got.DeletedAt)

	list, err := env.repo.ListWorkflowStatuses(env.ctx, env.projectID)
	require.NoError(t, err)
	for _, s := range list {
		assert.NotEqual(t, st.ID, s.ID, "soft-deleted status must not appear in list")
	}

	nTrans, err := env.client.WorkflowTransition.Query().
		Where(
			workflowtransition.Or(
				workflowtransition.HasFromStatusWith(workflowstatus.IDEQ(st.ID)),
				workflowtransition.HasToStatusWith(workflowstatus.IDEQ(st.ID)),
			),
		).
		Count(env.ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, nTrans, "transitions involving status must be purged")
}

// TestDeleteWorkflowStatus_ConflictWhenIssues_TASK171 — A3W8-5-002 guard.
// Mutation: убрать count issues / Conflict → delete проходит при живых issues → FAIL.
func TestDeleteWorkflowStatus_ConflictWhenIssues_TASK171(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	st, err := env.repo.CreateWorkflowStatus(env.ctx, env.projectID, "Busy", "in_progress", "#f00", 40)
	require.NoError(t, err)

	_, err = env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title:      "On busy status",
		IssueType:  "task",
		Priority:   "medium",
		ReporterID: env.userID,
		StatusID:   &st.ID,
	})
	require.NoError(t, err)

	err = env.repo.DeleteWorkflowStatus(env.ctx, env.projectID, st.ID)
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr), "expected AppError, got %v", err)
	assert.Equal(t, 409, appErr.Code)

	still, err := env.client.WorkflowStatus.Get(env.ctx, st.ID)
	require.NoError(t, err)
	assert.Nil(t, still.DeletedAt)
}

// TestEntRepository_DeleteWorkflowStatus_IdempotentSoft — повторный delete → NotFound.
func TestEntRepository_DeleteWorkflowStatus_IdempotentSoft(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	st, err := env.repo.CreateWorkflowStatus(env.ctx, env.projectID, "Once", "done", "#0f0", 60)
	require.NoError(t, err)
	require.NoError(t, env.repo.DeleteWorkflowStatus(env.ctx, env.projectID, st.ID))

	err = env.repo.DeleteWorkflowStatus(env.ctx, env.projectID, st.ID)
	require.True(t, ent.IsNotFound(err), "second soft-delete must NotFound, got %v", err)

	// sanity: deleted_at set
	got, err := env.client.WorkflowStatus.Get(env.ctx, st.ID)
	require.NoError(t, err)
	require.NotNil(t, got.DeletedAt)
	assert.WithinDuration(t, time.Now().UTC(), *got.DeletedAt, time.Minute)
}
