package mail_test
import (
	"github.com/valerii/hub/backend/internal/mail"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testHubMailApp(t *testing.T) *fiber.App {
	t.Helper()
	cfg := mail.Config{
		InternalToken: "hub-mail-secret",
		FromAlerts:    "alerts@hub.local",
		DailyLimit:    25,
		ResendAPIKey:  "re_test_key",
	}
	svc := mail.NewService(cfg)
	h := mail.NewHubHandler(svc)
	app := fiber.New()
	h.RegisterRoutes(app)
	return app
}

func TestHubHandler_HealthAndStatus(t *testing.T) {
	app := testHubMailApp(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var health map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&health))
	assert.Equal(t, "ok", health["status"])
	assert.Equal(t, "hub-mail", health["service"])

	req = httptest.NewRequest(http.MethodGet, "/v1/status", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var status mail.StatusResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&status))
	assert.Equal(t, "resend", status.Provider)
	assert.True(t, status.ResendReady)
	assert.Equal(t, 25, status.DailyLimit)
}

func TestHubHandler_Send_AuthAndValidation(t *testing.T) {
	app := testHubMailApp(t)

	body, _ := json.Marshal(mail.SendRequest{To: "u@test.local", Subject: "Alert"})
	req := httptest.NewRequest(http.MethodPost, "/v1/send", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	req = httptest.NewRequest(http.MethodPost, "/v1/send", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(mail.ExportInternalTokenHeader(), "hub-mail-secret")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	// A7-005: fake Resend key → real HTTP attempt fails (not silent 200 stub).
	req = httptest.NewRequest(http.MethodPost, "/v1/send", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(mail.ExportInternalTokenHeader(), "hub-mail-secret")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	// Dev no-op when key empty → 200.
	devApp := fiber.New()
	mail.NewHubHandler(mail.NewService(mail.Config{
		InternalToken: "hub-mail-secret",
		FromAlerts:    "alerts@hub.local",
	})).RegisterRoutes(devApp)
	req = httptest.NewRequest(http.MethodPost, "/v1/send", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(mail.ExportInternalTokenHeader(), "hub-mail-secret")
	resp, err = devApp.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestService_Status(t *testing.T) {
	svc := mail.NewService(mail.Config{FromAlerts: "a@b.com", DailyLimit: 10})
	st := svc.Status()
	assert.Equal(t, "a@b.com", st.FromAlerts)
	assert.False(t, st.ResendReady)
	assert.True(t, st.AlertsEnabled)
}
