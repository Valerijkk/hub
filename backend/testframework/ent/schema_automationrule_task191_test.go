//go:build hubtest

package ent_test

// TASK-191 / EA-001 — schema + migrate guards for AutomationRule soft-delete.
import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchema_TASK191_SoftDeleteField(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	src, err := os.ReadFile(filepath.Join(filepath.Dir(thisFile), "..", "..", "ent", "schema", "automationrule.go"))
	require.NoError(t, err)
	body := string(src)
	assert.Contains(t, body, `field.Time("deleted_at")`)
	assert.Contains(t, body, `index.Fields("deleted_at")`)
}

func TestMigrate_TASK191_NeedsOwnerApply(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	sqlPath := filepath.Join(filepath.Dir(thisFile), "..", "..", "migrations", "20260729_task191_automationrule_soft_delete.sql")
	src, err := os.ReadFile(sqlPath)
	require.NoError(t, err)
	body := string(src)
	assert.Contains(t, body, "needs-owner-apply")
	assert.Contains(t, body, "automation_rules")
	assert.Contains(t, body, "deleted_at")
}
