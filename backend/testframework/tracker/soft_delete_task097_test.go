package tracker_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/tracker"
)

// TestEntRepository_LabelSoftDelete_A2W8_3_005 — DeleteLabel → deleted_at; list hides; row survives.
// Mutation: Label.Delete() hard → Get fails → FAIL.
// Mutation: drop DeletedAtIsNil from ListLabels → soft-deleted listed → FAIL.
func TestEntRepository_LabelSoftDelete_A2W8_3_005(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	lbl, err := env.repo.CreateLabel(env.ctx, env.projectID, "TrashMe", "#ff0000", nil)
	require.NoError(t, err)

	require.NoError(t, env.repo.DeleteLabel(env.ctx, env.projectID, lbl.ID))

	list, err := env.repo.ListLabels(env.ctx, env.projectID)
	require.NoError(t, err)
	for _, l := range list {
		assert.NotEqual(t, lbl.ID, l.ID, "A2W8-3-005: soft-deleted label must not appear in ListLabels")
	}

	raw, err := env.client.Label.Get(env.ctx, lbl.ID)
	require.NoError(t, err, "A2W8-3-005: hard Delete would remove the row")
	require.NotNil(t, raw.DeletedAt)

	assertLabelSoftDeleteSource(t)
}

// TestEntRepository_CustomFieldSoftDelete_A2W8_2_004 — DeleteCustomField soft-delete.
func TestEntRepository_CustomFieldSoftDelete_A2W8_2_004(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	cf, err := env.repo.CreateCustomField(env.ctx, env.projectID, tracker.CreateCustomFieldInput{
		Name: "CFTrash", FieldType: "text",
	})
	require.NoError(t, err)

	require.NoError(t, env.repo.DeleteCustomField(env.ctx, env.projectID, cf.ID))

	list, err := env.repo.ListCustomFields(env.ctx, env.projectID)
	require.NoError(t, err)
	for _, f := range list {
		assert.NotEqual(t, cf.ID, f.ID)
	}

	raw, err := env.client.CustomFieldDefinition.Get(env.ctx, cf.ID)
	require.NoError(t, err)
	require.NotNil(t, raw.DeletedAt)

	assertCustomFieldSoftDeleteSource(t)
}

func assertLabelSoftDeleteSource(t *testing.T) {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	backendRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
	src, err := os.ReadFile(filepath.Join(backendRoot, "internal", "tracker", "repo_labels.go"))
	require.NoError(t, err)
	chunk := string(src)
	assert.Contains(t, chunk, "SetDeletedAt", "A2W8-3-005: DeleteLabel must SetDeletedAt")
	assert.Contains(t, chunk, "DeletedAtIsNil", "A2W8-3-005: ListLabels must filter DeletedAtIsNil")
	assert.NotContains(t, chunk, "Label.Delete()", "A2W8-3-005: must not hard-delete labels")
}

func assertCustomFieldSoftDeleteSource(t *testing.T) {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	backendRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
	src, err := os.ReadFile(filepath.Join(backendRoot, "internal", "tracker", "repo_custom_fields.go"))
	require.NoError(t, err)
	chunk := string(src)
	assert.Contains(t, chunk, "SetDeletedAt")
	assert.Contains(t, chunk, "DeletedAtIsNil")
	assert.False(t, strings.Contains(chunk, "CustomFieldDefinition.Delete()"),
		"A2W8-2-004: must not hard-delete custom fields")
}
