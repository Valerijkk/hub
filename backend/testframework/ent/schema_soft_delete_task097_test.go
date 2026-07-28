//go:build hubtest

package ent_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSchema_SoftDeleteGaps_TASK097 — schema evidence for deleted_at on TASK-097 entities.
// Mutation: remove deleted_at from a listed schema → FAIL.
func TestSchema_SoftDeleteGaps_TASK097(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	schemaDir := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", "ent", "schema"))

	mustHave := map[string]string{
		"leaverequest.go":          "A2W8-3-001",
		"label.go":                 "A2W8-3-005",
		"labelgroup.go":            "A2W8-3-005",
		"customfielddefinition.go": "A2W8-2-004",
		"sprint.go":                "A2W8-4-004",
		"recording.go":             "A2W8-4-005",
		"kudo.go":                  "A2W8-3-004",
		"githubintegration.go":     "W7-A3-010",
		"gitlabintegration.go":     "W7-A3-010",
	}
	for file, finding := range mustHave {
		src, err := os.ReadFile(filepath.Join(schemaDir, file))
		require.NoError(t, err, file)
		assert.Contains(t, string(src), `field.Time("deleted_at")`, "%s (%s) must declare deleted_at", file, finding)
	}

	jl, err := os.ReadFile(filepath.Join(schemaDir, "journalline.go"))
	require.NoError(t, err)
	assert.False(t, strings.Contains(string(jl), `field.Time("deleted_at")`),
		"JournalLine must NOT have own deleted_at (ADR join-filter)")
	assert.Contains(t, string(jl), "ADR-ENT-SOFT-DELETE-GAPS-2026")
}
