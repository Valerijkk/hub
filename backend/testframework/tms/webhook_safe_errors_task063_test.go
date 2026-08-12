//go:build hubtest

package tms_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mutation: вернуть err.Error() парсера в 400 → NotContains/Equal краснеет.
func TestWebhookAllure_InvalidJSON_StableMessage_P3BE08(t *testing.T) {
	app, projectID := setupTMSWebhookApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/webhook"

	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte("{bad")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, string(body), "невалидный JSON")
	assert.NotContains(t, string(body), "invalid character")
}

func TestWebhookJUnit_InvalidXML_StableMessage_P3BE08(t *testing.T) {
	app, projectID := setupTMSWebhookApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/webhook/junit"

	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte("<not>xml")))
	req.Header.Set("Content-Type", "application/xml")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, string(body), "невалидный JUnit XML")
}
