//go:build hubtest

package observability_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/platform/observability"
)

// TestSanitizeOperationLabel_UnknownToOther_P3BE12 — W4-A3-008:
// неизвестный/user-controlled operation → "other" (cardinality guard).
func TestSanitizeOperationLabel_UnknownToOther_P3BE12(t *testing.T) {
	other := observability.ExportUnknownOperationLabel()
	assert.Equal(t, "other", other)

	assert.Equal(t, "tracker.ListIssues", observability.ExportSanitizeOperationLabel("tracker.ListIssues"))
	assert.Equal(t, "auth.Login", observability.ExportSanitizeOperationLabel("auth.Login"))
	assert.Equal(t, other, observability.ExportSanitizeOperationLabel("550e8400-e29b-41d4-a716-446655440000"))
	assert.Equal(t, other, observability.ExportSanitizeOperationLabel("user-controlled"))
	assert.Equal(t, other, observability.ExportSanitizeOperationLabel("tracker."))
	assert.Equal(t, other, observability.ExportSanitizeOperationLabel(".Create"))
	assert.Equal(t, "", observability.ExportSanitizeOperationLabel(""))
}

func TestRecordOperation_UnknownBucketsToOther_P3BE12(t *testing.T) {
	observability.ResetMetrics()
	t.Cleanup(observability.ResetMetrics)

	uid := "550e8400-e29b-41d4-a716-446655440000"
	observability.ExportRecordOperation(uid, 10, false)
	snap := observability.Snapshot()
	require.Len(t, snap, 1)
	assert.Equal(t, "other", snap[0].Operation)
	assert.Equal(t, int64(1), snap[0].Count)
}
