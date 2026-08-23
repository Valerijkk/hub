//go:build hubtest

package task167_test

// TASK-167 / A3W8-1 — schema/ADR guards (отдельный пакет: sibling hubtest в ent/
// сейчас не компилируется без codegen TASK-170/171).
import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func schemaDir(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "ent", "schema")
}

// TestSchema_TASK167_AIUserPreferenceDefaultProviderSetNull — A3W8-1-001.
// Mutation: убрать Annotations(entsql.OnDelete(entsql.SetNull)) с default_provider → FAIL.
func TestSchema_TASK167_AIUserPreferenceDefaultProviderSetNull(t *testing.T) {
	src, err := os.ReadFile(filepath.Join(schemaDir(t), "aiuserpreference.go"))
	require.NoError(t, err)
	body := string(src)
	assert.Contains(t, body, `edge.To("default_provider"`)
	assert.Contains(t, body, `entsql.OnDelete(entsql.SetNull)`)
	assert.Contains(t, body, `"entgo.io/ent/dialect/entsql"`)
}

// TestSchema_TASK167_ChannelMemberLWW_NoVersionMixin — A3W8-1-002 path B.
func TestSchema_TASK167_ChannelMemberLWW_NoVersionMixin(t *testing.T) {
	// Decision traceability:
	// .planning/decisions/ADR-ENT-CHANNELMEMBER-LWW-2026.md
	// Its required approval markers are enforced by
	// .planning/audits/test-rebase/validate_planning_contracts.py in PR Guard,
	// because the canonical backend container intentionally mounts only backend/.
	src, err := os.ReadFile(filepath.Join(schemaDir(t), "channelmember.go"))
	require.NoError(t, err)
	body := string(src)
	assert.NotContains(t, body, "VersionMixin{}", "ChannelMember LWW path: no VersionMixin until migrate wave")
	assert.NotContains(t, body, `field.Int("version")`)
}
