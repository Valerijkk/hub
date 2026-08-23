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

// TestSchema_P2BE22_UniqueFKCheck — schema-only mutation markers (migrations = owner approval).
func TestSchema_P2BE22_UniqueFKCheck(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	schemaDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "ent", "schema")

	comp, err := os.ReadFile(filepath.Join(schemaDir, "component.go"))
	require.NoError(t, err)
	assert.Contains(t, string(comp), `index.Edges("project").Fields("name").Unique()`)

	jl, err := os.ReadFile(filepath.Join(schemaDir, "journalline.go"))
	require.NoError(t, err)
	assert.Contains(t, string(jl), "journalline_debit_xor_credit")
	assert.Contains(t, string(jl), "(debit_cents > 0) = (credit_cents = 0)")

	tcv, err := os.ReadFile(filepath.Join(schemaDir, "testcaseversion.go"))
	require.NoError(t, err)
	assert.Contains(t, string(tcv), `index.Edges("test_case").Fields("version_number").Unique()`)

	bm, err := os.ReadFile(filepath.Join(schemaDir, "bookmark.go"))
	require.NoError(t, err)
	assert.Contains(t, string(bm), `edge.From("channel", Channel.Type)`)
	assert.True(t, strings.Contains(string(bm), `Field("channel_id")`))
}
