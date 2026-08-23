package tms_test

// Тесты модуля TMS: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tms"
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleExportCases() []tms.TestCaseExportData {
	return []tms.TestCaseExportData{{
		Key: "HUB-1", Title: "Login flow", Type: "classic", Priority: "high",
		Status: "active", Steps: "1. Open app", Precondition: "User exists", Suite: "Auth",
	}}
}

// TestExportTestCasesToJSON проверяет сценарий: экспорт test cases to json.
func TestExportTestCasesToJSON(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, tms.ExportTestCasesToJSON(&buf, sampleExportCases()))

	var decoded []tms.TestCaseExportData
	require.NoError(t, json.Unmarshal(buf.Bytes(), &decoded))
	require.Len(t, decoded, 1)
	assert.Equal(t, "HUB-1", decoded[0].Key)
}

// TestExportTestCasesToXML проверяет сценарий: экспорт test cases to xml.
func TestExportTestCasesToXML(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, tms.ExportTestCasesToXML(&buf, sampleExportCases()))
	out := buf.String()
	assert.Contains(t, out, "<testcases>")
	assert.Contains(t, out, "<key>HUB-1</key>")
}

// TestExportTestCasesToCSV проверяет сценарий: экспорт test cases to csv.
func TestExportTestCasesToCSV(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, tms.ExportTestCasesToCSV(&buf, sampleExportCases()))
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	require.GreaterOrEqual(t, len(lines), 2)
	assert.Contains(t, lines[0], "Key")
	assert.Contains(t, lines[1], "HUB-1")
}
