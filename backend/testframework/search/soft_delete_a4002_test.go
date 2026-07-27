package search_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSearchSQL_ExcludesSoftDeleted_A4_002 — Ctrl+K shard SQL must exclude
// soft-deleted rows (and pending files). Mutation: drop deleted_at / status → FAIL.
func TestSearchSQL_ExcludesSoftDeleted_A4_002(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	servicePath := filepath.Join(filepath.Dir(thisFile), "..", "..", "internal", "search", "service.go")
	src, err := os.ReadFile(servicePath)
	require.NoError(t, err)
	body := string(src)

	for _, table := range []string{"issues", "test_cases", "projects", "file_metadata"} {
		idx := strings.Index(body, "FROM "+table)
		require.Greater(t, idx, 0, "missing FROM %s shard", table)
		chunk := body[idx:]
		end := strings.Index(chunk, "defaultSearchLimit")
		require.Greater(t, end, 0, "shard %s truncated", table)
		chunk = chunk[:end]
		assert.Contains(t, chunk, "deleted_at IS NULL",
			"A4-002: %s search must filter soft-deleted", table)
	}

	fileIdx := strings.Index(body, "FROM file_metadata")
	require.Greater(t, fileIdx, 0)
	fileChunk := body[fileIdx:]
	fileChunk = fileChunk[:strings.Index(fileChunk, "defaultSearchLimit")]
	assert.Contains(t, fileChunk, "status = 'confirmed'",
		"A4-002: file search must require confirmed status")
}
