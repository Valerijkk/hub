//go:build hubtest

package tms_test

// TASK-141 — A3W2-4 residual mutation/regression coverage (001..005).
import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/schema"
	"github.com/valerii/hub/backend/ent/testcase"
	"github.com/valerii/hub/backend/internal/tms"
)

// createSuiteFailRepo — CreateSuite всегда err (A3W2-4-005 mutation target).
type createSuiteFailRepo struct {
	*mockRepository
}

func (r *createSuiteFailRepo) CreateSuite(_ context.Context, _ uuid.UUID, _ tms.CreateSuiteInput) (*ent.TestSuite, error) {
	return nil, errors.New("suite create failed: synthetic")
}

// TestTask141_ExportTestCases_StepsFilled — A3W2-4-001.
// Мутация: убрать formatTestStepsForExport / Steps:"" → CSV колонка Steps пустая.
func TestTask141_ExportTestCases_StepsFilled(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	repo := newMockRepo()
	tcID := uuid.New()
	repo.testCases[tcID] = &ent.TestCase{
		ID: tcID, CaseKey: "HUB-TC-141", Title: "Export steps",
		CaseType: testcase.CaseTypeClassic, Priority: testcase.PriorityHigh,
		Status: testcase.StatusActive,
		Steps: []schema.TestStep{
			{Order: 1, Action: "Open app"},
			{Order: 2, Action: "Click login"},
		},
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
		Edges: ent.TestCaseEdges{Project: &ent.Project{ID: projectID}},
	}
	svc := tms.NewTMSService(repo, nil)

	reader, filename, err := svc.ExportTestCases(ctx, projectID, "csv")
	require.NoError(t, err)
	assert.Equal(t, "test-cases.csv", filename)
	data, err := io.ReadAll(reader)
	require.NoError(t, err)

	records, err := csv.NewReader(bytes.NewReader(data)).ReadAll()
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(records), 2)
	assert.Contains(t, records[0], "Steps")
	stepsCol := -1
	for i, h := range records[0] {
		if h == "Steps" {
			stepsCol = i
			break
		}
	}
	require.GreaterOrEqual(t, stepsCol, 0)
	require.NotEmpty(t, records[1][stepsCol], "Steps column must be populated when TC has steps")
	assert.Contains(t, records[1][stepsCol], "Open app")
	assert.Contains(t, records[1][stepsCol], "Click login")

	// Reimport round-trip: колонка Steps не теряется при parse.
	inputs, err := tms.ImportFromCSV(bytes.NewReader(data), nil)
	require.NoError(t, err)
	require.NotEmpty(t, inputs)
	require.NotEmpty(t, inputs[0].Steps)
	assert.Equal(t, "Open app", inputs[0].Steps[0].Action)
}

// TestTask141_FormatTestStepsForExport — A3W2-4-001 helper.
func TestTask141_FormatTestStepsForExport(t *testing.T) {
	out := tms.ExportFormatTestStepsForExport([]schema.TestStep{
		{Order: 1, Action: "A"},
		{Order: 2, Keyword: "When", Action: "B"},
	})
	assert.Equal(t, "1. A\n2. When B", out)
	assert.Empty(t, tms.ExportFormatTestStepsForExport(nil))
}

// TestTask141_ParseTestRailXML_DepthCap — A3W2-4-002.
// Мутация: убрать assertXMLMaxDepth из parseTestRailXML → nested XML проходит.
func TestTask141_ParseTestRailXML_DepthCap(t *testing.T) {
	nested := strings.Repeat(`<a>`, 40) + strings.Repeat(`</a>`, 40)
	_, err := tms.ExportParseTestRailXML([]byte(nested))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "глубин")
}

// TestTask141_ParseTestRailCSV_RecordsCap — A3W2-4-002.
// Мутация: убрать maxImportRecords check из parseTestRailCSV → oversized OK.
func TestTask141_ParseTestRailCSV_RecordsCap(t *testing.T) {
	var b strings.Builder
	b.WriteString("ID,Title,Section,Type,Priority\n")
	for i := 0; i < 5001; i++ {
		fmt.Fprintf(&b, "%d,Case %d,Sec,Manual,High\n", i, i)
	}
	_, err := tms.ExportParseTestRailCSV([]byte(b.String()))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "слишком много записей")
}

// TestTask141_HandleImport_FixedParseError — A3W2-4-003.
// Мутация: вернуть "ошибка парсинга файла: "+err.Error() → body leak глубин/internals.
func TestTask141_HandleImport_FixedParseError(t *testing.T) {
	app, projectID := setupTMSImportLimitsApp(t)

	nested := `{"x":0}`
	for i := 0; i < 33; i++ {
		nested = `{"a":` + nested + `}`
	}
	body := `[` + nested + `]`
	req := tmsImportMultipartRequest(t, projectID, "deep.json", []byte(body), "json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, json.Unmarshal(raw, &payload))
	errMsg, _ := payload["error"].(string)
	assert.Equal(t, "некорректный файл импорта", errMsg)
	assert.NotContains(t, string(raw), "глубин")
	assert.NotContains(t, string(raw), "ошибка парсинга файла")
}

// TestTask141_ExportTestCasesToCSV_FlushError — A3W2-4-004.
// Мутация: убрать w.Error() после Flush → errWriter fail на Flush даёт nil.
func TestTask141_ExportTestCasesToCSV_FlushError(t *testing.T) {
	w := &errWriter{fail: 1}
	err := tms.ExportTestCasesToCSV(w, []tms.TestCaseExportData{{
		Key: "K", Title: "T", Type: "classic", Priority: "high",
		Status: "active", Steps: "1. X", Precondition: "p", Suite: "s",
	}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "CSV")
}

// TestTask141_ImportTestRail_CreateSuiteFail_Partial — A3W2-4-005.
// Мутация: continue без Errors++ → Partial=false при Created>0.
func TestTask141_ImportTestRail_CreateSuiteFail_Partial(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()
	repo := &createSuiteFailRepo{mockRepository: newMockRepo()}
	svc := tms.NewTMSService(repo, nil)

	csvData := `ID,Title,Section,Type,Priority,Preconditions,Steps,Expected Result
1,Login flow,Authentication,Manual,High,User exists,1. Open app,App opens
`
	result, err := svc.ImportTestRailCSV(ctx, projectID, userID, []byte(csvData))
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.GreaterOrEqual(t, result.Errors, 1, "CreateSuite fail must increment Errors")
	assert.True(t, result.Partial, "CreateSuite fail must set Partial")
	assert.Equal(t, 1, result.Created, "case still created without suite")
}
