package tms_test

// Тесты модуля TMS (волна 96): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tms"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleTestRailCSV = `ID,Title,Section,Type,Priority,Preconditions,Steps,Expected Result
1,Login flow,Authentication,Manual,High,User exists,1. Open app,App opens
2,,Regression,Manual,Medium,,,
`

const sampleTestRailXML = `<?xml version="1.0" encoding="UTF-8"?>
<cases>
  <case>
    <id>10</id>
    <title>XML import case</title>
    <section>Smoke</section>
    <type>Automated</type>
    <priority>Critical</priority>
    <steps>1. Run test</steps>
    <expected>Pass</expected>
  </case>
</cases>`

// TestParseTestRailCSV проверяет сценарий: разбор test rail csv.
func TestParseTestRailCSV(t *testing.T) {
	cases, err := tms.ExportParseTestRailCSV([]byte(sampleTestRailCSV))
	require.NoError(t, err)
	require.Len(t, cases, 2)
	assert.Equal(t, "Login flow", cases[0].Title)
	assert.Equal(t, "Authentication", cases[0].SectionName)
}

// TestParseTestRailCSV_OnlyHeaderRowFails проверяет сценарий: разбор test rail csv only header row fails.
func TestParseTestRailCSV_OnlyHeaderRowFails(t *testing.T) {
	_, err := tms.ExportParseTestRailCSV([]byte("ID,Title\n"))
	require.Error(t, err)
}

// TestParseTestRailCSV_EmptyFile проверяет сценарий: разбор test rail csv пустой file.
func TestParseTestRailCSV_EmptyFile(t *testing.T) {
	_, err := tms.ExportParseTestRailCSV([]byte("Title\n"))
	require.Error(t, err)
}

// TestTryReadCSV проверяет сценарий: try read csv.
func TestTryReadCSV(t *testing.T) {
	records, err := tms.ExportTryReadCSV([]byte("a,b\n1,2\n"), ',')
	require.NoError(t, err)
	require.Len(t, records, 2)
}

// TestParseTestRailXML проверяет сценарий: разбор test rail xml.
func TestParseTestRailXML(t *testing.T) {
	cases, err := tms.ExportParseTestRailXML([]byte(sampleTestRailXML))
	require.NoError(t, err)
	require.Len(t, cases, 1)
	assert.Equal(t, "XML import case", cases[0].Title)
}

// TestParseTestRailXML_Invalid проверяет отклонение при разбор test rail xml некорректный.
func TestParseTestRailXML_Invalid(t *testing.T) {
	_, err := tms.ExportParseTestRailXML([]byte("<broken"))
	require.Error(t, err)
}

// TestImportTestRailCSV_Service проверяет бизнес-логику: импорт test rail csv сервис.
func TestImportTestRailCSV_Service(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()
	svc := tms.NewTMSService(newMockRepo(), nil)

	result, err := svc.ImportTestRailCSV(ctx, projectID, userID, []byte(sampleTestRailCSV))
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.Created)
	assert.Equal(t, 1, result.Skipped)
	assert.GreaterOrEqual(t, result.Errors, 1)
}

// TestImportTestRailXML_Service проверяет бизнес-логику: импорт test rail xml сервис.
func TestImportTestRailXML_Service(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()
	svc := tms.NewTMSService(newMockRepo(), nil)

	result, err := svc.ImportTestRailXML(ctx, projectID, userID, []byte(sampleTestRailXML))
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.Created)
}

// TestMapTestRailType_EntEnum проверяет маппинг TestRail type → ent case_type classic/bdd.
func TestMapTestRailType_EntEnum(t *testing.T) {
	t.Parallel()
	tests := []struct {
		trType   string
		expected string
	}{
		{trType: "Manual", expected: "classic"},
		{trType: "Automated", expected: "classic"},
		{trType: "Functional", expected: "classic"},
		{trType: "Regression", expected: "classic"},
		{trType: "Other", expected: "classic"},
		{trType: "", expected: "classic"},
		{trType: "classic", expected: "classic"},
		{trType: "BDD", expected: "bdd"},
		{trType: "Scenario", expected: "bdd"},
		{trType: "Gherkin", expected: "bdd"},
	}
	for _, tt := range tests {
		t.Run(tt.trType, func(t *testing.T) {
			assert.Equal(t, tt.expected, tms.ExportMapTestRailType(tt.trType))
		})
	}
}

// TestImportTestRailCases_Empty проверяет сценарий: импорт test rail cases пустой.
func TestImportTestRailCases_Empty(t *testing.T) {
	svc := tms.NewTMSService(newMockRepo(), nil)
	result, err := tms.ExportImportTestRailCases(svc, context.Background(), uuid.New(), uuid.New(), nil)
	require.NoError(t, err)
	assert.Equal(t, 0, result.Created)
}

func setupTestRailHandlerApp(t *testing.T) (*fiber.App, uuid.UUID, uuid.UUID) {
	t.Helper()
	userID := uuid.New()
	projectID := uuid.New()
	svc := tms.NewTMSService(newMockRepo(), nil)
	h := tms.NewHandler(svc)

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	tmsGroup := api.Group("/projects/:projectId/tms", authMw)
	tmsGroup.Post("/import/testrail", h.ImportTestRailCases)
	return app, projectID, userID
}

func testRailMultipartRequest(t *testing.T, projectID uuid.UUID, filename string, content []byte, format string) *http.Request {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile("file", filename)
	require.NoError(t, err)
	_, err = part.Write(content)
	require.NoError(t, err)
	if format != "" {
		require.NoError(t, w.WriteField("format", format))
	}
	require.NoError(t, w.Close())

	url := fmt.Sprintf("/api/projects/%s/tms/import/testrail", projectID)
	req := httptest.NewRequest(http.MethodPost, url, &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

// TestHandler_ImportTestRailCases_CSV проверяет HTTP-слой: HTTP-обработчик импорт test rail cases csv.
func TestHandler_ImportTestRailCases_CSV(t *testing.T) {
	app, projectID, _ := setupTestRailHandlerApp(t)
	req := testRailMultipartRequest(t, projectID, "cases.csv", []byte(sampleTestRailCSV), "csv")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result tms.ImportResult
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, 1, result.Created)
}

// TestHandler_ImportTestRailCases_XML проверяет HTTP-слой: HTTP-обработчик импорт test rail cases xml.
func TestHandler_ImportTestRailCases_XML(t *testing.T) {
	app, projectID, _ := setupTestRailHandlerApp(t)
	req := testRailMultipartRequest(t, projectID, "cases.xml", []byte(sampleTestRailXML), "xml")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_ImportTestRailCases_MissingFile проверяет отклонение при HTTP-обработчик импорт test rail cases отсутствующий file.
func TestHandler_ImportTestRailCases_MissingFile(t *testing.T) {
	app, projectID, _ := setupTestRailHandlerApp(t)
	url := fmt.Sprintf("/api/projects/%s/tms/import/testrail", projectID)
	req := httptest.NewRequest(http.MethodPost, url, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
