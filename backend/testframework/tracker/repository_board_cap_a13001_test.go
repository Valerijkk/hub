//go:build hubtest

package tracker_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/tracker"
)

// TestEntRepository_GetBoardData_CapsIssuesPerColumn_A13_001 — без Limit колонка
// тянула все issues проекта; hard cap MaxBoardIssuesPerColumn обрезает ответ.
// Mutation: снять Limit → FAIL (len > MaxBoardIssuesPerColumn).
func TestEntRepository_GetBoardData_CapsIssuesPerColumn_A13_001(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	board, err := env.repo.CreateBoard(env.ctx, env.projectID, "Cap board", nil)
	require.NoError(t, err)

	cap := tracker.MaxBoardIssuesPerColumn
	extra := 5
	for i := 0; i < cap+extra; i++ {
		_, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
			Title:      fmt.Sprintf("Board cap %d", i),
			IssueType:  "task",
			Priority:   "medium",
			ReporterID: env.userID,
		})
		require.NoError(t, err)
	}

	data, err := env.repo.GetBoardData(env.ctx, board.ID)
	require.NoError(t, err)
	require.NotNil(t, data.Edges.Project)

	totalLoaded := 0
	maxInColumn := 0
	for _, st := range data.Edges.Project.Edges.WorkflowStatuses {
		n := len(st.Edges.Issues)
		totalLoaded += n
		if n > maxInColumn {
			maxInColumn = n
		}
		assert.LessOrEqual(t, n, cap, "column %s exceeded per-column cap", st.Name)
	}
	assert.Equal(t, cap, maxInColumn, "expected at least one column to hit the hard cap")
	assert.LessOrEqual(t, totalLoaded, cap*len(data.Edges.Project.Edges.WorkflowStatuses))
}
