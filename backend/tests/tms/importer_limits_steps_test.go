package tms_test

// P1-BE-13..14 / TASK-016: LimitReader + depth/record caps + Steps mapping.
//
// Mutation (C):
//   - убрать io.LimitReader / size check в handleImport → oversized body больше не 413
//   - убрать case "steps" / parseTestRailSteps в ImportFromJSON → Steps остаётся nil
import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/platform/middleware"
	"github.com/valerii/hub/backend/internal/tms"
)

// maxImportFileBytesMirror совпадает с maxImportFileBytes в importer.go (не экспортирован).
const maxImportFileBytesMirror = 25 * 1024 * 1024

func tmsImportMultipartRequest(t *testing.T, projectID uuid.UUID, filename string, content []byte, format string) *http.Request {
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

	url := fmt.Sprintf("/api/projects/%s/tms/import", projectID)
	req := httptest.NewRequest(http.MethodPost, url, &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

func setupTMSImportLimitsApp(t *testing.T) (*fiber.App, uuid.UUID) {
	t.Helper()
	userID := uuid.New()
	projectID := uuid.New()
	svc := tms.NewTMSService(newMockRepo(), nil)
	h := tms.NewHandler(svc)

	// BodyLimit выше maxImportFileBytes, чтобы 413 срабатывал в handleImport, а не в Fiber.
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
		BodyLimit:    40 * 1024 * 1024,
	})
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	group := api.Group("/projects/:projectId/tms", authMw)
	group.Post("/import", h.ImportTestCases)
	return app, projectID
}

// TestImportFromCSV_StepsColumn_Mapped проверяет P1-BE-14: колонка Steps → непустой Steps.
func TestImportFromCSV_StepsColumn_Mapped(t *testing.T) {
	csvData := "Title,Steps,Priority\nLogin flow,\"1. Open page\n2. Click login\",high\n"
	inputs, err := tms.ImportFromCSV(strings.NewReader(csvData), nil)
	require.NoError(t, err)
	require.Len(t, inputs, 1)
	require.NotEmpty(t, inputs[0].Steps, "Steps must be mapped from CSV column")
	assert.Equal(t, "Open page", inputs[0].Steps[0].Action)
	assert.Equal(t, "Click login", inputs[0].Steps[1].Action)
}

// TestImportFromCSV_StepsColumn_RussianHeader проверяет маппинг колонки «Шаги».
func TestImportFromCSV_StepsColumn_RussianHeader(t *testing.T) {
	csvData := "Название,Шаги\nКейс,\"1. Шаг А\n2. Шаг Б\"\n"
	inputs, err := tms.ImportFromCSV(strings.NewReader(csvData), nil)
	require.NoError(t, err)
	require.Len(t, inputs, 1)
	require.Len(t, inputs[0].Steps, 2)
	assert.Equal(t, "Шаг А", inputs[0].Steps[0].Action)
}

// TestImportFromJSON_StepsMapped проверяет P1-BE-14: JSON field steps → CreateTestCaseInput.Steps.
func TestImportFromJSON_StepsMapped(t *testing.T) {
	body := `[{"title":"Case A","priority":"high","steps":"1. Do X\n2. Do Y"}]`
	inputs, err := tms.ImportFromJSON(strings.NewReader(body))
	require.NoError(t, err)
	require.Len(t, inputs, 1)
	require.Len(t, inputs[0].Steps, 2)
	assert.Equal(t, "Do X", inputs[0].Steps[0].Action)
	assert.Equal(t, "Do Y", inputs[0].Steps[1].Action)
}

// TestImportFromJSON_DeepNestingRejected проверяет P1-BE-13: глубокий JSON → ошибка (400 на HTTP).
func TestImportFromJSON_DeepNestingRejected(t *testing.T) {
	// Глубина > maxImportNestDepth (32): массив + объект + 32 вложенных ключа.
	nested := `{"x":0}`
	for i := 0; i < 33; i++ {
		nested = `{"a":` + nested + `}`
	}
	body := `[` + nested + `]`
	_, err := tms.ImportFromJSON(strings.NewReader(body))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "глубин")
}

// TestImportFromXML_DeepNestingRejected проверяет P1-BE-13: глубокий XML → ошибка.
func TestImportFromXML_DeepNestingRejected(t *testing.T) {
	var b strings.Builder
	b.WriteString(`<testcases>`)
	const depth = 40
	for i := 0; i < depth; i++ {
		fmt.Fprintf(&b, `<n%d>`, i)
	}
	b.WriteString(`x`)
	for i := depth - 1; i >= 0; i-- {
		fmt.Fprintf(&b, `</n%d>`, i)
	}
	b.WriteString(`</testcases>`)
	_, err := tms.ImportFromXML(strings.NewReader(b.String()))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "глубин")
}

// TestImportFromCSV_TooManyRecordsRejected проверяет hard-cap числа записей.
func TestImportFromCSV_TooManyRecordsRejected(t *testing.T) {
	var b strings.Builder
	b.WriteString("Title,Priority\n")
	// maxImportRecords = 5000; 5001 строк данных → отказ.
	for i := 0; i < 5001; i++ {
		fmt.Fprintf(&b, "Case %d,high\n", i)
	}
	_, err := tms.ImportFromCSV(strings.NewReader(b.String()), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "слишком много записей")
}

// TestHandler_ImportTestCases_OversizedBody_413 проверяет P1-BE-13: oversized → 413.
func TestHandler_ImportTestCases_OversizedBody_413(t *testing.T) {
	app, projectID := setupTMSImportLimitsApp(t)

	// payload чуть больше лимита; content — валидный CSV-префикс + padding.
	payload := make([]byte, maxImportFileBytesMirror+1)
	copy(payload, []byte("Title,Priority\nCase,high\n"))
	for i := 20; i < len(payload); i++ {
		payload[i] = 'x'
	}

	req := tmsImportMultipartRequest(t, projectID, "huge.csv", payload, "csv")
	resp, err := app.Test(req, -1) // без таймаута: тело ~25 MB
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusRequestEntityTooLarge, resp.StatusCode)
}

// TestHandler_ImportTestCases_DeepJSON_400 проверяет nested bomb → 400 на HTTP-слое.
func TestHandler_ImportTestCases_DeepJSON_400(t *testing.T) {
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
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_ImportTestCases_StepsMappedEndToEnd проверяет, что CSV со Steps создаёт кейс с шагами в репо.
func TestHandler_ImportTestCases_StepsMappedEndToEnd(t *testing.T) {
	userID := uuid.New()
	projectID := uuid.New()
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	h := tms.NewHandler(svc)

	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler})
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	group := api.Group("/projects/:projectId/tms", authMw)
	group.Post("/import", h.ImportTestCases)

	csvData := "Title,Steps,Priority\nE2E case,\"1. First\n2. Second\",medium\n"
	req := tmsImportMultipartRequest(t, projectID, "cases.csv", []byte(csvData), "csv")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result tms.ImportResult
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, 1, result.Created)

	require.Len(t, repo.testCases, 1)
	for _, tc := range repo.testCases {
		require.NotEmpty(t, tc.Steps, "CreateTestCase must receive mapped Steps")
		require.GreaterOrEqual(t, len(tc.Steps), 2)
		return
	}
}
