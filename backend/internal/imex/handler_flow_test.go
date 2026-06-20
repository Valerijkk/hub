package imex

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupImexHandlerApp(t *testing.T) (*fiber.App, uuid.UUID, uuid.UUID) {
	t.Helper()
	svc, _, projectID, userID := newImexTestEnv(t)
	h := NewHandler(svc)

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)
	return app, projectID, userID
}

func TestHandler_ExportIssuesCSV(t *testing.T) {
	app, projectID, _ := setupImexHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/projects/"+projectID.String()+"/export/issues?format=csv", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "text/csv")

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "Key")
}

func TestHandler_ExportIssuesExcel(t *testing.T) {
	app, projectID, _ := setupImexHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/projects/"+projectID.String()+"/export/issues?format=excel", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "spreadsheetml")
}

func TestHandler_ExportProjectZIP(t *testing.T) {
	app, projectID, _ := setupImexHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/projects/"+projectID.String()+"/export/full", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/zip", resp.Header.Get("Content-Type"))
}

func TestHandler_ImportIssues_MissingFile(t *testing.T) {
	app, projectID, _ := setupImexHandlerApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID.String()+"/import/issues", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_ImportIssues_Preview(t *testing.T) {
	app, projectID, _ := setupImexHandlerApp(t)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "issues.csv")
	require.NoError(t, err)
	_, err = part.Write([]byte("title,type,priority\nHandler issue,task,medium\n"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID.String()+"/import/issues?mode=preview", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_InvalidProjectID(t *testing.T) {
	app, _, _ := setupImexHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/projects/not-a-uuid/export/issues", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_ExportRunPDF_InvalidRunID(t *testing.T) {
	app, projectID, _ := setupImexHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/projects/"+projectID.String()+"/test-runs/not-a-uuid/export/pdf", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_ImportJira_MissingFile(t *testing.T) {
	app, projectID, _ := setupImexHandlerApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID.String()+"/import/jira", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
