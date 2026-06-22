package tms_test

// Тесты HTTP-обработчика TMS: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tms"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTMSWebhookApp(t *testing.T) (*fiber.App, uuid.UUID) {
	t.Helper()
	svc := tms.NewTMSService(newMockRepo(), nil)
	h := tms.NewHandler(svc)
	projectID := uuid.New()
	userID := uuid.New()

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	tms := api.Group("/projects/:projectId/tms", authMw)
	tms.Post("/webhook", h.WebhookAllure)
	tms.Post("/webhook/junit", h.WebhookJUnit)
	return app, projectID
}

// TestHandler_WebhookAllure_EmptyBody проверяет HTTP-слой: HTTP-обработчик webhook allure пустой body.
func TestHandler_WebhookAllure_EmptyBody(t *testing.T) {
	app, projectID := setupTMSWebhookApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/webhook"

	req := httptest.NewRequest(http.MethodPost, base, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_WebhookAllure_InvalidJSON проверяет отклонение при HTTP-обработчик webhook allure некорректный json.
func TestHandler_WebhookAllure_InvalidJSON(t *testing.T) {
	app, projectID := setupTMSWebhookApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/webhook"

	req := httptest.NewRequest(http.MethodPost, base+"?title=CI", bytes.NewReader([]byte("not-json")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_WebhookAllure_Success проверяет HTTP-слой: HTTP-обработчик webhook allure успех.
func TestHandler_WebhookAllure_Success(t *testing.T) {
	app, projectID := setupTMSWebhookApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/webhook"

	req := httptest.NewRequest(http.MethodPost, base+"?title=CI+Run", bytes.NewReader([]byte(validAllureJSON)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var wh tms.WebhookResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&wh))
	assert.Equal(t, "CI Run", wh.Title)
	assert.NotEmpty(t, wh.RunID)
	assert.GreaterOrEqual(t, wh.Total, 1)
}

// TestHandler_WebhookJUnit_Success проверяет HTTP-слой: HTTP-обработчик webhook j unit успех.
func TestHandler_WebhookJUnit_Success(t *testing.T) {
	app, projectID := setupTMSWebhookApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/webhook/junit"

	req := httptest.NewRequest(http.MethodPost, base+"?title=JUnit+CI", bytes.NewReader([]byte(junitTestSuitesXML)))
	req.Header.Set("Content-Type", "application/xml")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var wh tms.WebhookResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&wh))
	assert.Equal(t, "JUnit CI", wh.Title)
	assert.Equal(t, 3, wh.Total)
}

// TestHandler_WebhookJUnit_InvalidXML проверяет отклонение при HTTP-обработчик webhook j unit некорректный xml.
func TestHandler_WebhookJUnit_InvalidXML(t *testing.T) {
	app, projectID := setupTMSWebhookApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/webhook/junit"

	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte("<not-xml")))
	req.Header.Set("Content-Type", "application/xml")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
