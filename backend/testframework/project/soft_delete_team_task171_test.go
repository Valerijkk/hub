//go:build hubtest

package project_test

// TASK-171 / A3W8-5-001 — SoftDeleteTeam + list DeletedAtIsNil.
import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/team"
)

// TestSoftDeleteTeam_SetsDeletedAt_TASK171 — A3W8-5-001 soft path.
// Mutation: вернуть hard DeleteOneID в SoftDeleteTeam → row gone → FAIL.
func TestSoftDeleteTeam_SetsDeletedAt_TASK171(t *testing.T) {
	env := newProjectRepoEnv(t)

	tm, err := env.repo.CreateTeam(env.ctx, "SoftPath", "dissolve")
	require.NoError(t, err)
	_, err = env.repo.AddTeamMember(env.ctx, tm.ID, env.owner.ID)
	require.NoError(t, err)

	require.NoError(t, env.repo.SoftDeleteTeam(env.ctx, tm.ID))

	got, err := env.client.Team.Get(env.ctx, tm.ID)
	require.NoError(t, err, "soft-delete must keep row")
	require.NotNil(t, got.DeletedAt)

	_, err = env.repo.FindTeamByID(env.ctx, tm.ID)
	require.True(t, ent.IsNotFound(err))

	list, err := env.repo.ListTeamsByUserID(env.ctx, env.owner.ID)
	require.NoError(t, err)
	for _, item := range list {
		assert.NotEqual(t, tm.ID, item.ID, "soft-deleted team must not appear in list")
	}

	active, err := env.client.Team.Query().Where(team.DeletedAtIsNil(), team.IDEQ(tm.ID)).Count(env.ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, active)

	// Idempotent: second soft-delete → NotFound.
	err = env.repo.SoftDeleteTeam(env.ctx, tm.ID)
	require.True(t, ent.IsNotFound(err))
}

// TestSoftDeleteTeam_SourceUsesSetDeletedAt_TASK171 — source guard A3W8-5-001.
// Mutation: убрать SetDeletedAt из SoftDeleteTeam → FAIL.
func TestSoftDeleteTeam_SourceUsesSetDeletedAt_TASK171(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	srcPath := filepath.Join(filepath.Dir(thisFile), "..", "..", "internal", "project", "repository.go")
	src, err := os.ReadFile(srcPath)
	require.NoError(t, err)
	body := string(src)
	assert.Contains(t, body, "func (r *EntRepository) SoftDeleteTeam")
	assert.Contains(t, body, "SetDeletedAt")
	assert.Contains(t, body, "ListTeamsByUserID")
	assert.Contains(t, body, "team.DeletedAtIsNil()")
}
