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

func repoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..")
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
	src, err := os.ReadFile(filepath.Join(schemaDir(t), "channelmember.go"))
	require.NoError(t, err)
	body := string(src)
	assert.NotContains(t, body, "VersionMixin{}", "ChannelMember LWW path: no VersionMixin until migrate wave")
	assert.NotContains(t, body, `field.Int("version")`)

	adrPath := filepath.Join(repoRoot(t), ".planning", "decisions", "ADR-ENT-CHANNELMEMBER-LWW-2026.md")
	adr, err := os.ReadFile(adrPath)
	require.NoError(t, err, adrPath)
	adrBody := string(adr)
	assert.Contains(t, adrBody, "LWW accepted")
	assert.Contains(t, adrBody, "A3W8-1-002")
	assert.Contains(t, adrBody, "Owner lock / approve LWW path")
	assert.Contains(t, adrBody, "Accepted 2026-07-29")
}
