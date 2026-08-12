package sprint_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/sprint"
)

// TestEntRepository_SprintSoftDelete_A2W8_4_004 — SoftDeleteSprint → deleted_at; list/get hide.
// Mutation: drop DeletedAtIsNil from ListSprints → soft-deleted listed → FAIL.
func TestEntRepository_SprintSoftDelete_A2W8_4_004(t *testing.T) {
	env := newSprintRepoEnv(t)

	now := time.Now().UTC()
	s, err := env.repo.CreateSprint(env.ctx, env.projectID, sprint.CreateSprintRequest{
		Name: "TrashSprint", StartDate: now, EndDate: now.AddDate(0, 0, 14),
	})
	require.NoError(t, err)

	require.NoError(t, env.repo.SoftDeleteSprint(env.ctx, env.projectID, s.ID))

	_, err = env.repo.GetSprint(env.ctx, s.ID)
	require.True(t, ent.IsNotFound(err), "A2W8-4-004: GetSprint must skip trash")

	list, err := env.repo.ListSprints(env.ctx, env.projectID)
	require.NoError(t, err)
	for _, item := range list {
		assert.NotEqual(t, s.ID, item.ID)
	}

	raw, err := env.client.Sprint.Get(env.ctx, s.ID)
	require.NoError(t, err)
	require.NotNil(t, raw.DeletedAt)

	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	backendRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
	src, err := os.ReadFile(filepath.Join(backendRoot, "internal", "sprint", "repository.go"))
	require.NoError(t, err)
	assert.Contains(t, string(src), "SetDeletedAt")
	assert.Contains(t, string(src), "DeletedAtIsNil")
}
