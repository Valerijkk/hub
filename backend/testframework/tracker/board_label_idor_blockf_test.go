//go:build hubtest

package tracker_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
)

// TestGetBoardData_RejectsForeignProject_BlockF — board UUID from project B must
// not load under project A (mutation: drop HasProjectWith → FAIL).
func TestGetBoardData_RejectsForeignProject_BlockF(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	board, err := env.repo.CreateBoard(env.ctx, env.projectID, "A board", nil)
	require.NoError(t, err)

	foreign, err := env.client.Project.Create().
		SetName("Foreign Board Proj").
		SetKey("FBD").
		Save(env.ctx)
	require.NoError(t, err)

	_, err = env.repo.GetBoardData(env.ctx, foreign.ID, board.ID)
	require.True(t, ent.IsNotFound(err), "cross-project board must 404, got %v", err)

	got, err := env.repo.GetBoardData(env.ctx, env.projectID, board.ID)
	require.NoError(t, err)
	require.Equal(t, board.ID, got.ID)
}

// TestDeleteLabel_RejectsForeignProject_BlockF — label UUID from project B must
// not delete under project A (mutation: DeleteOneID without project scope → FAIL).
func TestDeleteLabel_RejectsForeignProject_BlockF(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	label, err := env.repo.CreateLabel(env.ctx, env.projectID, "bug", "#ff0000", nil)
	require.NoError(t, err)

	foreign, err := env.client.Project.Create().
		SetName("Foreign Label Proj").
		SetKey("FLB").
		Save(env.ctx)
	require.NoError(t, err)

	err = env.repo.DeleteLabel(env.ctx, foreign.ID, label.ID)
	require.True(t, ent.IsNotFound(err), "cross-project label delete must 404, got %v", err)

	still, err := env.client.Label.Get(env.ctx, label.ID)
	require.NoError(t, err)
	require.Equal(t, label.ID, still.ID)

	require.NoError(t, env.repo.DeleteLabel(env.ctx, env.projectID, label.ID))
}
