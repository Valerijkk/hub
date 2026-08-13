//go:build hubtest

package tms_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/tms"
)

// errWriter fails after N successful Write calls — CSV header Write must surface err (W2-A4-013).
type errWriter struct {
	n    int
	fail int
}

func (w *errWriter) Write(p []byte) (int, error) {
	w.n++
	if w.n >= w.fail {
		return 0, errors.New("disk full")
	}
	return len(p), nil
}

// TestExportRunResultsToCSV_WriteError_TASK057 — W2-A4-013 / P3-BE-01.
// Мутация: вернуть `_ = w.Write(...)` на заголовках → err nil при broken writer.
func TestExportRunResultsToCSV_WriteError_TASK057(t *testing.T) {
	w := &errWriter{fail: 1}
	err := tms.ExportRunResultsToCSV(w, tms.TestRunExportData{
		Title: "Run", CreatedAt: "2026-07-28", Total: 1, Passed: 1, Failed: 0, PassRate: 100,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "CSV")

	var buf bytes.Buffer
	require.NoError(t, tms.ExportRunResultsToCSV(&buf, tms.TestRunExportData{
		Title: "Run", CreatedAt: "2026-07-28", Total: 0, Passed: 0, Failed: 0, PassRate: 0,
	}))
	assert.Contains(t, buf.String(), "Run Title")
}
