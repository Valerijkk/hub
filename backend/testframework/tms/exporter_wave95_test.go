package tms_test

// Тесты модуля TMS (волна 95): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tms"
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/testcase"
	"github.com/valerii/hub/backend/ent/testresult"
)

func sampleRunExportData() tms.TestRunExportData {
	return tms.TestRunExportData{
		Title:     "Regression Sprint 95",
		CreatedAt: "2026-06-21",
		Total:     3,
		Passed:    2,
		Failed:    1,
		Skipped:   0,
		Broken:    0,
		PassRate:  66.7,
		Results: []tms.ResultExportData{
			{CaseKey: "TC-1", Title: "Login", Status: "passed", Duration: "1.2s", Comment: "ok"},
			{CaseKey: "TC-2", Title: strings.Repeat("Очень длинное название кейса ", 5), Status: "failed", Duration: "0.5s", Comment: strings.Repeat("comment ", 20)},
			{CaseKey: "TC-3", Title: "Logout", Status: "broken", Duration: "2s", Comment: ""},
		},
	}
}

// TestExportTestCasesToExcel проверяет сценарий: экспорт test cases to excel.
func TestExportTestCasesToExcel(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, tms.ExportTestCasesToExcel(&buf, sampleExportCases()))
	assert.Greater(t, buf.Len(), 100)
}

// TestExportRunResultsToCSV проверяет сценарий: экспорт run results to csv.
func TestExportRunResultsToCSV(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, tms.ExportRunResultsToCSV(&buf, sampleRunExportData()))
	out := buf.String()
	assert.Contains(t, out, "Regression Sprint 95")
	assert.Contains(t, out, "TC-1")
	assert.Contains(t, out, "Pass Rate")
}

// TestExportTestRunToPDF проверяет сценарий: экспорт test run to pdf.
func TestExportTestRunToPDF(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, tms.ExportTestRunToPDF(&buf, sampleRunExportData()))
	assert.Greater(t, buf.Len(), 500)
}

// TestTruncate проверяет сценарий: truncate.
func TestTruncate(t *testing.T) {
	assert.Equal(t, "short", tms.ExportTruncate("short", 10))
	long := strings.Repeat("а", 50)
	out := tms.ExportTruncate(long, 10)
	assert.True(t, strings.HasSuffix(out, "..."))
	assert.LessOrEqual(t, len([]rune(out)), 13)
}

// TestTMSService_ExportTestCases_Excel проверяет бизнес-логику: tms сервис экспорт test cases excel.
func TestTMSService_ExportTestCases_Excel(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	repo := newMockRepo()
	tcID := uuid.New()
	repo.testCases[tcID] = &ent.TestCase{
		ID: tcID, CaseKey: "TC-95", Title: "Excel export",
		CaseType: testcase.CaseTypeClassic, Priority: testcase.PriorityHigh,
		Status: testcase.StatusActive, Precondition: "ready",
		Edges: ent.TestCaseEdges{Suite: &ent.TestSuite{Name: "Regression"}},
	}
	svc := tms.NewTMSService(repo, nil)

	reader, filename, err := svc.ExportTestCases(ctx, projectID, "excel")
	require.NoError(t, err)
	assert.Equal(t, "test-cases.xlsx", filename)
	data, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Greater(t, len(data), 200)
}

// TestTMSService_ExportTestRunReport_PDF проверяет бизнес-логику: tms сервис экспорт test run report pdf.
func TestTMSService_ExportTestRunReport_PDF(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	runID := uuid.New()
	tcID := uuid.New()
	dur := int64(1500)
	repo := newMockRepo()
	repo.runs[runID] = &ent.TestRun{
		ID: runID, Title: "Wave 95 run", TotalCount: 1, PassedCount: 1,
		CreatedAt: time.Now(),
		Edges:     ent.TestRunEdges{Project: &ent.Project{ID: projectID}},
	}
	repo.results[uuid.New()] = &ent.TestResult{
		ID: uuid.New(), Status: testresult.StatusPassed, Comment: "ok", DurationMs: &dur,
		Edges: ent.TestResultEdges{
			TestCase: &ent.TestCase{ID: tcID, CaseKey: "TC-PDF", Title: "PDF case"},
		},
	}
	svc := tms.NewTMSService(repo, nil)

	reader, filename, err := svc.ExportTestRunReport(ctx, projectID, runID, "pdf")
	require.NoError(t, err)
	assert.Equal(t, "run-report.pdf", filename)
	data, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Greater(t, len(data), 500)
}

// TestImportFromJSON проверяет сценарий: импорт from json.
func TestImportFromJSON(t *testing.T) {
	body := `[{"title":"Case A","priority":"high","status":"active","case_type":"classic"}]`
	inputs, err := tms.ImportFromJSON(strings.NewReader(body))
	require.NoError(t, err)
	require.Len(t, inputs, 1)
	assert.Equal(t, "Case A", inputs[0].Title)
}

// TestImportFromXML проверяет сценарий: импорт from xml.
func TestImportFromXML(t *testing.T) {
	body := `<testcases><testcase><title>Case B</title><priority>medium</priority></testcase></testcases>`
	inputs, err := tms.ImportFromXML(strings.NewReader(body))
	require.NoError(t, err)
	require.Len(t, inputs, 1)
	assert.Equal(t, "Case B", inputs[0].Title)
}

// TestUpdateTestResult_Success проверяет сценарий: обновление test result успех.
func TestUpdateTestResult_Success(t *testing.T) {
	ctx := context.Background()
	resultID := uuid.New()
	runID := uuid.New()
	projectID := uuid.New()
	repo := newMockRepo()
	status := "failed"
	repo.runs[runID] = &ent.TestRun{
		ID: runID,
		Edges: ent.TestRunEdges{Project: &ent.Project{ID: projectID}},
	}
	repo.results[resultID] = &ent.TestResult{
		ID:     resultID,
		Status: testresult.StatusUnknown,
		Edges:  ent.TestResultEdges{Run: &ent.TestRun{ID: runID}},
	}
	svc := tms.NewTMSService(repo, nil)
	comment := "failed on step 2"

	resp, err := svc.UpdateTestResult(ctx, projectID, runID, resultID, tms.UpdateTestResultRequest{
		Status:  &status,
		Comment: &comment,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "failed", resp.Status)
}
