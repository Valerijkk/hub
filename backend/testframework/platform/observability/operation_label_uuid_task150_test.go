//go:build hubtest

package observability_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/platform/observability"
)

// TestSanitizeOperationLabel_DomainUUIDToOther_A3W4_3_002 — tracker.<uuid> → other (cardinality).
func TestSanitizeOperationLabel_DomainUUIDToOther_A3W4_3_002(t *testing.T) {
	other := observability.ExportUnknownOperationLabel()

	assert.Equal(t, other,
		observability.ExportSanitizeOperationLabel("tracker.550e8400-e29b-41d4-a716-446655440000"),
		"A3W4-3-002: domain+UUID must bucket to other")
	assert.Equal(t, other,
		observability.ExportSanitizeOperationLabel("auth.123e4567-e89b-12d3-a456-426614174000"))

	// Valid allowlisted labels stay intact.
	assert.Equal(t, "tracker.CreateIssue",
		observability.ExportSanitizeOperationLabel("tracker.CreateIssue"))
	assert.Equal(t, "outbox.publish",
		observability.ExportSanitizeOperationLabel("outbox.publish"))
	assert.Equal(t, "auth.Login",
		observability.ExportSanitizeOperationLabel("auth.Login"))

	// Prior W4-A3-008 guards still hold.
	assert.Equal(t, other, observability.ExportSanitizeOperationLabel("550e8400-e29b-41d4-a716-446655440000"))
	assert.Equal(t, other, observability.ExportSanitizeOperationLabel("user-controlled"))
	assert.Equal(t, other, observability.ExportSanitizeOperationLabel("unknown.Create"))
}

func TestRecordOperation_DomainUUIDBucketsToOther_A3W4_3_002(t *testing.T) {
	observability.ResetMetrics()
	t.Cleanup(observability.ResetMetrics)

	uid := "tracker.550e8400-e29b-41d4-a716-446655440000"
	observability.ExportRecordOperation(uid, 12, false)
	snap := observability.Snapshot()
	require.Len(t, snap, 1)
	assert.Equal(t, "other", snap[0].Operation)
	assert.Equal(t, int64(1), snap[0].Count)
}
