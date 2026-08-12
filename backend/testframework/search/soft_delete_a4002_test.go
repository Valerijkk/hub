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

// TestSearchSQL_ExcludesSoftDeleted_A4_002 — secondary static guard (A2W5-A5-003).
// Primary evidence = soft_delete_a4002_runtime_test.go (Postgres behavior).
func TestSearchSQL_ExcludesSoftDeleted_A4_002(t *testing.T) {
	t.Log("A2W5-A5-003: static Contains is secondary; primary = runtime soft-delete tests")

	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	servicePath := filepath.Join(filepath.Dir(thisFile), "..", "..", "internal", "search", "service.go")
	src, err := os.ReadFile(servicePath)
	require.NoError(t, err)
	body := string(src)

	for _, table := range []string{"issues", "test_cases", "projects", "file_metadata", "channels"} {
		idx := strings.Index(body, "FROM "+table)
		require.Greater(t, idx, 0, "missing FROM %s shard", table)
		chunk := body[idx:]
		end := strings.Index(chunk, "defaultSearchLimit")
		require.Greater(t, end, 0, "shard %s truncated", table)
		chunk = chunk[:end]
		assert.Contains(t, chunk, "deleted_at IS NULL",
			"A4-002/P1-BE-32: %s search must filter soft-deleted", table)
	}

	fileIdx := strings.Index(body, "FROM file_metadata")
	require.Greater(t, fileIdx, 0)
	fileChunk := body[fileIdx:]
	fileChunk = fileChunk[:strings.Index(fileChunk, "defaultSearchLimit")]
	assert.Contains(t, fileChunk, "status = 'confirmed'",
		"A4-002: file search must require confirmed status")

	chIdx := strings.Index(body, "FROM channels")
	require.Greater(t, chIdx, 0)
	chChunk := body[chIdx:]
	chChunk = chChunk[:strings.Index(chChunk, "defaultSearchLimit")]
	assert.Contains(t, chChunk, "channel_type = 'public'",
		"P1-BE-32: public channels stay project-scoped")
	assert.Contains(t, chChunk, "channel_members",
		"P1-BE-32: private/group require channel membership")

	msgIdx := strings.Index(body, "func (s *SearchService) searchMessages")
	require.Greater(t, msgIdx, 0)
	msgChunk := body[msgIdx:]
	endMsg := strings.Index(msgChunk, "\nfunc ")
	if endMsg > 0 {
		msgChunk = msgChunk[:endMsg]
	}
	assert.Contains(t, msgChunk, `"deleted_at": nil`,
		"P1-BE-32: message search must filter soft-deleted Mongo docs")
}
