package tms

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleExportCases() []TestCaseExportData {
	return []TestCaseExportData{{
		Key: "HUB-1", Title: "Login flow", Type: "classic", Priority: "high",
		Status: "active", Steps: "1. Open app", Precondition: "User exists", Suite: "Auth",
	}}
}

func TestExportTestCasesToJSON(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, ExportTestCasesToJSON(&buf, sampleExportCases()))

	var decoded []TestCaseExportData
	require.NoError(t, json.Unmarshal(buf.Bytes(), &decoded))
	require.Len(t, decoded, 1)
	assert.Equal(t, "HUB-1", decoded[0].Key)
}

func TestExportTestCasesToXML(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, ExportTestCasesToXML(&buf, sampleExportCases()))
	out := buf.String()
	assert.Contains(t, out, "<testcases>")
	assert.Contains(t, out, "<key>HUB-1</key>")
}

func TestExportTestCasesToCSV(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, ExportTestCasesToCSV(&buf, sampleExportCases()))
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	require.GreaterOrEqual(t, len(lines), 2)
	assert.Contains(t, lines[0], "Key")
	assert.Contains(t, lines[1], "HUB-1")
}
