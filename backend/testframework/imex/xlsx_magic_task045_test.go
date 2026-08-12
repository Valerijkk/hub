//go:build hubtest

package imex_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExportIssuesExcel_XLSXMagic — W7-A3-009 / P2-BE-20: ZIP/OOXML magic PK.
func TestExportIssuesExcel_XLSXMagic(t *testing.T) {
	svc, _, ctx, projectID, _ := newImexTestEnv(t)
	data, err := svc.ExportIssuesExcel(ctx, projectID)
	require.NoError(t, err)
	require.Greater(t, len(data), 4)
	assert.True(t, bytes.HasPrefix(data, []byte{0x50, 0x4b, 0x03, 0x04}), "xlsx must start with PK\\x03\\x04")
}
