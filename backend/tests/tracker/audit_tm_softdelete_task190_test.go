//go:build hubtest

package tracker_test

// TASK-190 / AUDIT-TM-001…002 — soft-delete default status + board statuses.
import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/tracker"
)

// TestCreateIssue_RejectsSoftDeletedDefaultStatus_AUDIT_TM_001
// Mutation: убрать DeletedAtIsNil из findDefaultStatusTx → CreateIssue OK на trash default → FAIL.
func TestCreateIssue_RejectsSoftDeletedDefaultStatus_AUDIT_TM_001(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	_, err := env.client.WorkflowStatus.UpdateOneID(env.statusID).
		SetDeletedAt(time.Now().UTC()).
		Save(env.ctx)
	require.NoError(t, err)

	_, err = env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title:      "Should not bind soft default",
		IssueType:  "task",
		Priority:   "medium",
		ReporterID: env.userID,
	})
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err), "soft-deleted default must not be used, got %v", err)
}

// TestGetBoardData_SkipsSoftDeletedStatuses_AUDIT_TM_002
// Mutation: убрать DeletedAtIsNil в WithWorkflowStatuses → soft status на board → FAIL.
func TestGetBoardData_SkipsSoftDeletedStatuses_AUDIT_TM_002(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	extra, err := env.repo.CreateWorkflowStatus(env.ctx, env.projectID, "GhostCol", "todo", "#999", 99)
	require.NoError(t, err)
	require.NoError(t, env.repo.DeleteWorkflowStatus(env.ctx, env.projectID, extra.ID))

	board, err := env.repo.CreateBoard(env.ctx, env.projectID, "Kanban Soft", nil)
	require.NoError(t, err)

	data, err := env.repo.GetBoardData(env.ctx, env.projectID, board.ID)
	require.NoError(t, err)
	require.NotNil(t, data.Edges.Project)

	for _, st := range data.Edges.Project.Edges.WorkflowStatuses {
		assert.NotEqual(t, extra.ID, st.ID, "soft-deleted status must not appear on board")
		assert.Nil(t, st.DeletedAt)
	}
}
